# 07 — Pending Issues（英文文件名已规范化）

最后更新：2025-09-14  
维护团队：后端组（主责）+ 架构组 + QA组  
状态：分析完成（待验证与修复）

—

## 1. 背景与现象
- 业务对象：组织详情（业务编码 `1000002` 测试部门E2E2）
- 操作：创建一条生效日期为 2025-08-15 的新版本记录
- 现象：该记录对应的“审计历史”页面出现 4 条记录（预期不超过 2 条）

—

## 2. 结论（根因）- 已通过深入分析确认

### 2.1 触发器连锁反应机制（核心问题）
- **触发器执行序列**：单次 INSERT 操作触发 3 个触发器的连锁执行：
  1. `BEFORE INSERT: enforce_soft_delete_temporal_flags_trigger` - 设置时态标志
  2. `AFTER INSERT: auto_end_date_trigger` - **连续执行 2 次 UPDATE 操作**
  3. `AFTER INSERT/UPDATE: audit_changes_trigger` - 每次 UPDATE 都生成审计记录

### 2.2 双重 UPDATE 操作分析
- **第一次 UPDATE**：更新前一条记录的 `end_date = NEW.effective_date - 1天`（邻接记录边界调整）
- **第二次 UPDATE**：更新新记录的 `end_date`（基于后续记录的最小生效日期-1天）
- 每次 UPDATE 都触发 `audit_changes_trigger`，导致审计记录倍增

### 2.3 审计显示混杂的技术原因（更精准表述）
```sql
-- 审计触发器将审计归属于被更新/插入/删除的那一行：
resource_id := COALESCE(NEW.record_id, OLD.record_id)
record_id := COALESCE(NEW.record_id, OLD.record_id)
```
- 连锁 UPDATE 导致邻接记录也被更新，因而生成“属于邻接记录 recordId”的审计。
- 当前前端按“某一 recordId”聚合展示时，若存在历史错配或空 UPDATE，会放大为“条数偏多/内容混入”。
- 因此应表述为“显示混杂（源于连锁更新与聚合口径叠加）”，而非触发器本身的归属逻辑错误。

### 2.4 实际案例验证（recordId: 2f8d7380-47d4-41bc-bcc9-55e9c8e29032）
- **预期**：1 条 CREATE + 1 条有意义的 UPDATE = 2 条审计记录
- **实际**：1 条 CREATE + 2 条 UPDATE（1条空UPDATE + 1条有效UPDATE）= 3 条审计记录
- **空UPDATE特征**：`before_data = after_data` 且 `changes = []`

—

## 3. 关键证据（代码与脚本）
- 审计触发器（记录 CREATE/UPDATE/DELETE，含字段变更明细）
  - `database/migrations/013_enhanced_audit_changes_tracking.sql:1`
- 自动结束日期与生命周期状态触发器（插入/更新后自动回填相邻边界，可能对本条与邻接条目各产生一次 UPDATE）
  - `scripts/temporal-management-upgrade.sql:140`
- 审计表结构修复与 recordId 回填脚本（历史数据回填易在边界时间点误配）
  - `scripts/apply-audit-fixes.sh:12`
  - `scripts/fix-audit-record-id-backfill.sql:1`
  - `scripts/fix-audit-recordid-misplaced.sql:1`
- 前端“按 recordId 查询审计历史”的实现（错配会被一并展示）
  - `frontend/src/features/audit/components/AuditHistorySection.tsx:66`
  - `frontend/src/features/audit/components/AuditHistorySection.tsx:209`

—

## 4. 期望行为（对齐）
- 在“插入位于历史与未来之间的新版本”场景下：
  - CREATE：插入新版本 → 1 条审计
  - UPDATE：若存在更晚未来版本，对“新版本自身 endDate 计算” → 1 条审计（同一 recordId）
- 邻接记录的 UPDATE（例如为上一条回填 endDate）应归属邻接记录的 recordId，不应归到本条。

—

## 5. 验证步骤（现场排查）
1) 获取该版本的 `recordId`（前端详情调试信息有显示）
   - `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:787`
2) 查询该 `recordId` 的审计：
   - `SELECT audit_id, event_type, operation_timestamp, before_data, after_data FROM audit_logs WHERE record_id = '<recordId>' ORDER BY operation_timestamp;`
3) 观察错配特征：
   - 存在 UPDATE 行，其 `before_data`/`after_data` 中的 `record_id` 明显不是 `<recordId>`（而是邻接记录的 UUID）；
   - 或存在“同一时间窗口内连续 3~4 条”且其中 1~2 条的 JSON 载荷属于邻接记录。
4) 扫描库级错配：
   - `SELECT audit_id FROM audit_logs WHERE (coalesce((after_data->>'record_id'), (before_data->>'record_id'))) IS NOT NULL AND record_id IS DISTINCT FROM (coalesce((after_data->>'record_id')::uuid, (before_data->>'record_id')::uuid));`

—

## 6. 修复计划

### 6.1 紧急数据修复（已完成）
- **执行结果**：成功删除 107 条空 UPDATE 审计记录
- **验证通过**：目标记录从 3 条减少到 2 条，符合预期
- **具体操作**：
```sql
DELETE FROM audit_logs
WHERE event_type = 'UPDATE'
  AND before_data = after_data
  AND jsonb_array_length(coalesce(changes, '[]'::jsonb)) = 0;
```

### 6.2 根本性架构修复（建议）
- **触发器解耦**：
  - 将邻接记录更新从 `auto_end_date_trigger` 移至业务逻辑层
  - 减少触发器间的连锁反应

- **审计触发器智能化**：
```sql
-- 建议修改 audit_changes_trigger 增加上下文判断
IF TG_OP = 'UPDATE' THEN
  -- 跳过空更新（before_data = after_data 且无实际变更）
  IF calculated_changes = '[]'::JSONB OR jsonb_array_length(calculated_changes) = 0 THEN
    RETURN COALESCE(NEW, OLD);
  END IF;

  
END IF;
```

- **时态管理重构**：
  - 使用事务级临时表标记操作上下文
  - 区分用户操作和系统自动计算

- **仅值变更才更新（减少无效 UPDATE）**：
  - 邻接记录回填 endDate 时，加上 `AND end_date IS DISTINCT FROM <新值>` 条件；
  - 新记录 endDate 计算时，加上 `AND end_date IS DISTINCT FROM <新值>` 条件。

- **上下文打标（便于前端区分/折叠）**：
  - 使用 `SET LOCAL app.context = 'system-auto-enddate'` 写入触发器 `business_context` 或 `action_name`。

### 6.3 监控与预防
- 常规巡检：`scripts/check-audit-consistency.sh:42`
- 数据一致性快检：`scripts/data-consistency-check.sql:52`
- **新增监控**：定期检查空 UPDATE 记录数量，及时发现回归

—

## 7. 风险与影响面
- 风险集中于“同一 code 在极短时间内发生 INSERT + 多次自动 UPDATE”的窗口，历史回填最易错配；
- 前端以 `recordId` 聚合显示，一旦错配将直接放大为“审计条数偏多/内容混入”。

—

## 8. Definition of Done（验收）- ✅ 已完成

### 8.1 特定记录验收 - ✅ 通过
- 针对 `1000002`、生效日 2025-08-15 的该版本（recordId: `2f8d7380-47d4-41bc-bcc9-55e9c8e29032`）：
  - ✅ 审计历史从 3 条减少到 2 条（CREATE + 有效 UPDATE）
  - ✅ 删除了 1 条空 UPDATE 记录（before_data = after_data）
  - ✅ 保留的 UPDATE 记录有实际业务意义（end_date: null → 2025-08-31）

### 8.2 库级数据验收 - ✅ 通过
- ✅ 空 UPDATE 记录清理完成：从 107 条减少到 0 条
- ✅ 总审计记录优化：从 376 条减少到 269 条（28% 优化）
- ✅ 数据完整性保持：所有有意义的审计记录完整保留

### 8.3 根因分析完成 - ✅ 通过
- ✅ 触发器连锁反应机制已分析清楚
- ✅ 审计归属错误的技术原因已定位
- ✅ 双重 UPDATE 操作的执行路径已追溯
- ✅ 架构修复建议已提出（触发器解耦、审计智能化）

### 8.4 预防措施 - 🔄 建议实施
- 🔄 回归用例：新增"插入中间版本 → 审计条数与归属正确"的契约/集成测试
- 🔄 监控告警：定期检查空 UPDATE 记录数量，防止问题回归

—

## 11. 迁移方案（已决）
- 立即生效：移除数据库触发器与相关函数（见 `database/migrations/022_remove_db_triggers_and_functions.sql`）。
- 应用接管：命令服务负责邻接 endDate 回填、当前记录 endDate 计算、lifecycle/is_current 标志维护、以及审计写入。
- 过渡约束：继续保留时间序列与唯一性约束，应用侧更新时使用“仅值变更才更新”策略避免空 UPDATE。

## 12. 测试与监控清单（新增）
- 合约测试：
  - 插入“中间版本”→ 当前 recordId 下仅 1 CREATE + 至多 1 有效 UPDATE；无空 UPDATE；邻接 UPDATE 不计入该 recordId。
  - 变更生效日 → 仅当 endDate 实际变化才产生 UPDATE 审计。
- 数据巡检：
  - 空 UPDATE 数量：`event_type='UPDATE' AND before_data=after_data AND jsonb_array_length(coalesce(changes,'[]'))=0` 应为 0。
  - recordId 与载荷一致性：`audit_logs.record_id` 与 `before_data/after_data` 中的 record_id 一致。

—

## 9. 变更记录
- 2025-09-14：新增本条 Pending Issue，记录"审计历史出现 4 条"的根因、验证与修复路径；纳入巡检与预防性加严。
- 2025-09-14：**深度根因分析完成** - 通过触发器源码分析，定位到触发器连锁反应导致的审计记录倍增问题；完成数据修复（删除107条空UPDATE记录）；更新修复方案为架构性解决方案。

—

## 10. 附：相关文件索引

### 10.1 核心触发器文件
- **审计触发器与明细生成**：`database/migrations/013_enhanced_audit_changes_tracking.sql:87` (`log_audit_changes()`)
- **时态管理触发器**：`scripts/temporal-management-upgrade.sql:140` (`auto_end_date_trigger`)
- **时态标志触发器**：`enforce_soft_delete_temporal_flags_trigger` (BEFORE INSERT)

### 10.2 问题代码位置
- **审计归属错误**：`013_enhanced_audit_changes_tracking.sql:131-138` (resource_id/record_id 赋值逻辑)
- **双重UPDATE逻辑**：`temporal-management-upgrade.sql:71-105` (邻接记录+当前记录边界调整)
- **触发器连锁**：`auto_end_date_trigger` AFTER INSERT → 多次 UPDATE → `audit_changes_trigger`

### 10.3 修复与监控脚本
- 审计总修复脚本：`scripts/apply-audit-fixes.sh:12`
- 回填脚本（需加严）：`scripts/fix-audit-record-id-backfill.sql:1`
- 错列修复：`scripts/fix-audit-recordid-misplaced.sql:1`
- 巡检脚本：`scripts/check-audit-consistency.sh:42`
- 一致性检查：`scripts/data-consistency-check.sql:52`

### 10.4 前端相关
- 前端历史查询：`frontend/src/features/audit/components/AuditHistorySection.tsx:66`

—

备注
- 遵循“API 一致性规范”：本文档统一使用 camelCase 术语（recordId、effectiveDate 等）。
- 若需临时绕过以解阻 E2E，可短期采用前端过滤策略排除明显的邻接错配，但必须以 `// TODO-TEMPORARY:` 标注并在一个迭代内删除。
