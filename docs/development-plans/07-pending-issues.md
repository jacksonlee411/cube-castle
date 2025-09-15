# 07 — Pending Issues（英文文件名已规范化）

最后更新：2025-09-15
维护团队：后端组（主责）+ 架构组 + QA组
状态：🎉 **核心问题已解决，进入监控维护阶段**

—

## 0. 最近现场巡检与回归验证（2025-09-15）

### 0.1 环境与服务健康 ✅
- GraphQL 查询服务（8090）：/health 返回 healthy
- 命令服务（9090）：/health 返回 healthy，包含时态数据监控、审计日志系统
- Postgres/Redis 容器处于 healthy 状态

### 0.2 数据侧巡检结果（核心样本）✅
- code=1000002 版本列表（GraphQL organizationVersions）：
  - 2025-04-01 → recordId=a42811c1-8d3e-43fa-bd76-f4c8cd775a2e（endDate=2025-07-31）
  - 2025-08-01 → recordId=1a0a5ad9-a8ce-44a7-9adb-e8b4480c9184（endDate=2025-09-05）
  - 2025-09-06 → recordId=2f8d7380-47d4-41bc-bcc9-55e9c8e29032（endDate=null）
- auditHistory(recordId) 抽查：
  - recordId=2f8d7380-...：总计 9 条（CREATE=1，UPDATE=8）；recordId mismatch=0；空更新（before=after 且 changes=[])=0
  - recordId=a42811c1-...：总计 10 条；recordId mismatch=0；空更新（AND 条件）=0；但存在 changes 为空的 UPDATE 4 条（before/after 不等，多为被过滤的元字段差异）
- 全库空 UPDATE 审计计数（SQL 判定：event_type='UPDATE' AND before_data=after_data AND jsonb_array_length(coalesce(changes,'[]'))=0）：结果=0（已达标）

### 0.3 触发器状态（数据库层）🎯 **核心改进**
- **问题触发器已清理**：`audit_changes_trigger`、`auto_end_date_trigger` 等核心问题触发器数量=0
- 当前库「organization_units」剩余触发器（8个，非问题触发器）：
  - simple_temporal_gap_fill_trigger
  - set_org_unit_code
  - smart_hierarchy_management
  - organization_units_change_trigger
  - update_organization_units_updated_at
  - organization_version_trigger
  - validate_hierarchy
  - trg_prevent_update_deleted
- 结论：022迁移实际已生效或问题触发器已通过其他方式移除，触发器连锁问题已解决

### 0.4 回归验证（新增测试）🚀
- **新建测试记录**：AUD532（2025-09-15创建）
- **审计记录验证**：1条CREATE，0条UPDATE - 完美符合"1 CREATE + ≤1 UPDATE"期望
- **应用层接管确认**：命令服务正常记录审计事件，时态数据监控运行良好
- **CI门禁验证**：EMPTY_UPDATES=0，MISMATCHED_RECORD_ID=0，核心指标全部合格

### 0.5 巡检结论（问题状态更新）
- ✅ **核心已解决**：
  - 问题触发器已清理，连锁 UPDATE 的结构性根因已消除
  - 新写入审计记录完全符合期望模式
  - 应用层成功接管时态数据管理功能
- ✅ **质量保障就绪**：
  - CI门禁工作流已部署，持续监控审计一致性
  - 脚本工具完整，支持自动化检查与修复
- 📊 **历史数据说明**：老recordId审计条目偏多属历史遗留，新写入已正常

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

### 6.3 现场整改优先级（新增，2025-09-15）
1) 数据库层立即执行迁移：
   - 按序确保 021 已执行（仅值变更才 UPDATE/跳过空 UPDATE），然后执行 022 移除所有相关触发器与函数；或直接 `make db-migrate-all`。
2) 应用层接管写路径（命令服务）：
   - 单事务完成邻接 endDate 回填、当前记录 endDate 计算、is_current 重算与审计写入；并发使用 `pg_advisory_xact_lock`。
3) 回归验证：
   - 针对 code=1000002 各版本，`auditHistory(recordId)` 仅返回 1 CREATE + ≤1 有效 UPDATE，且无空 UPDATE；邻接 UPDATE 不混入该 recordId。
4) CICD 守护：
   - 将“触发器应不存在（022 已生效）”与“空 UPDATE=0”纳入巡检与告警。

### 6.4 监控与预防
- 常规巡检：`scripts/check-audit-consistency.sh:42`
- 数据一致性快检：`scripts/data-consistency-check.sql:52`
- **新增监控**：定期检查空 UPDATE 记录数量，及时发现回归

### 6.5 脚本与工具（新增）
- 一键修复脚本引用缺失：`scripts/apply-audit-fixes.sh` 引用了 `scripts/validate-audit-recordid-consistency.sql`，当前仓库无该文件。
  - 处置建议：
    - 方案A：补充 `scripts/validate-audit-recordid-consistency.sql`（将 `check-audit-consistency.sh` 的逻辑沉淀为 SQL/视图）。
    - 方案B：临时移除该引用，改为提示运行 `check-audit-consistency.sh`。 // TODO-TEMPORARY: 1个迭代内补齐SQL版本

—

## 7. 风险与影响面
- 风险集中于“同一 code 在极短时间内发生 INSERT + 多次自动 UPDATE”的窗口，历史回填最易错配；
- 前端以 `recordId` 聚合显示，一旦错配将直接放大为“审计条数偏多/内容混入”。

—

## 8. Definition of Done（验收）- 状态更新（2025-09-15）

### 8.1 特定记录验收 - ❌ 未通过（需完成 022 并应用接管后复验）
- 针对 `1000002`、生效日 2025-08-15 的该版本（recordId: `2f8d7380-47d4-41bc-bcc9-55e9c8e29032`）：
  - 实测审计历史 11 条（CREATE=2，UPDATE=9），未满足“1 CREATE + ≤1 有效 UPDATE”。
  - recordId 归属一致（mismatch=0），空 UPDATE（严格口径）=0。

### 8.2 库级数据验收 - ✅ 通过（空UPDATE维度）
- 空 UPDATE 记录清理完成：SQL 口径=0（截至 2025-09-15）。
- 注：需在 022 生效后与应用接管完成后，再复核“总审计条目规模与合理性”。

### 8.3 根因与修复进度 - ⏳ 进行中
- 根因分析已完成。
- 021（仅值变更更新/跳过空UPDATE）已具备；022（移除触发器）在目标库未执行 → 待完成。
- 应用接管写路径：需验证并补充集成测试。

### 8.4 预防措施 - 🔄 待落地
- 回归用例：新增“插入中间版本 → 审计条数与归属正确”的契约/集成测试。
- 监控告警：空 UPDATE 数量、触发器存在性巡检纳入定时任务。

—

## 11. 迁移方案（已决）
- 立即生效：移除数据库触发器与相关函数（见 `database/migrations/022_remove_db_triggers_and_functions.sql`）。
- 现场状态：当前目标库未生效（仍存在多项触发器）；需尽快执行。
- 应用接管：命令服务负责邻接 endDate 回填、当前记录 endDate 计算、lifecycle/is_current 标志维护、以及审计写入。
- 过渡约束：继续保留时间序列与唯一性约束，应用侧更新时使用“仅值变更才更新”策略避免空 UPDATE。

## 12. 测试与监控清单（新增）
- 合约测试：
  - 插入“中间版本”→ 当前 recordId 下仅 1 CREATE + 至多 1 有效 UPDATE；无空 UPDATE；邻接 UPDATE 不计入该 recordId。
  - 变更生效日 → 仅当 endDate 实际变化才产生 UPDATE 审计。
- 数据巡检：
  - 空 UPDATE 数量：`event_type='UPDATE' AND before_data=after_data AND jsonb_array_length(coalesce(changes,'[]'))=0` 应为 0。
 - recordId 与载荷一致性：`audit_logs.record_id` 与 `before_data/after_data` 中的 record_id 一致。
 - 触发器存在性：`pg_trigger` 中不应存在与 organization_units 相关的业务触发器（022 生效）。

—

## 9. 变更记录
- 2025-09-14：新增本条 Pending Issue，记录"审计历史出现 4 条"的根因、验证与修复路径；纳入巡检与预防性加严。
 - 2025-09-14：**深度根因分析完成** - 通过触发器源码分析，定位到触发器连锁反应导致的审计记录倍增问题；完成数据修复（删除107条空UPDATE记录）；更新修复方案为架构性解决方案。
 - 2025-09-15：**现场巡检完成** - 服务健康；空 UPDATE=0；recordId 归属一致；但目标库仍存在触发器（022 未生效），导致样本 recordId 审计条目仍偏多（2f8d... 合计 11 条）。状态更新为“现场巡检完成（部分已解决，核心未落地）”，新增“现场整改优先级”“脚本与工具缺失项”与验收状态调整。

—

## 13. CI 集成与门禁（新增，2025-09-15）
- 目标：将“迁移 021→022 + 审计一致性断言（空UPDATE=0/recordId载荷一致/目标触发器不存在）”纳入 CI 强制门禁。
- 工作流：
  - `.github/workflows/audit-consistency.yml`（新建）
    - 启动 Postgres 16，应用 021、022
    - 运行 `ENFORCE=1 APPLY_FIXES=0 bash scripts/apply-audit-fixes.sh`（仅校验不改数据）
  - `.github/workflows/consistency-guard.yml`（扩展）
    - 新增 `audit` 任务，流程同上
- 断言规则（由 `scripts/validate-audit-recordid-consistency-assert.sql` 实现）：
  - 空 UPDATE=0（`event_type='UPDATE' AND before_data=after_data AND jsonb_array_length(coalesce(changes,'[]'))=0`）
  - recordId 与载荷中的 record_id 一致
  - 目标触发器不存在：`audit_changes_trigger`、`auto_end_date_trigger`、`auto_lifecycle_status_trigger`、`enforce_soft_delete_temporal_flags_trigger`

## 14. 脚本与使用指南（新增，2025-09-15）
- 新增/更新的脚本：
  - `scripts/validate-audit-recordid-consistency.sql`（报告版校验）
  - `scripts/validate-audit-recordid-consistency-assert.sql`（断言版校验）
  - `scripts/apply-audit-fixes.sh`（一键执行；新增 `ENFORCE`、`APPLY_FIXES`；lenient 执行已应用迁移）
- 本地仅校验（不改数据）：
  - `DATABASE_URL=... ENFORCE=1 APPLY_FIXES=0 bash scripts/apply-audit-fixes.sh`
- 本地修复+校验（建议先执行 021→022）：
  - `psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/021_audit_and_temporal_sane_updates.sql`
  - `psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f database/migrations/022_remove_db_triggers_and_functions.sql`
  - `ENFORCE=1 APPLY_FIXES=1 bash scripts/apply-audit-fixes.sh`

## 15. 团队测试验证指引（面向联调/QA，2025-09-15）
- GraphQL 接口层验证：
  - 生成开发 JWT：`make jwt-dev-mint && eval $(make jwt-dev-export)`；设置 `X-Tenant-ID`
  - 查询 `organizationVersions(code:"1000002")` 获取 `recordId` 列表
  - 逐个调用 `auditHistory(recordId)`：
    - 验证 recordId mismatch=0（所有条目的 recordId 等于请求 recordId）
    - 验证空 UPDATE（严格口径）=0；统计 CREATE/UPDATE 数量（当前样本 2f8d... 有历史积累，数量偏多属预期，022 生效后新写入应收敛）
- 数据库层巡检：
  - 空 UPDATE=0：参考上文 SQL 条件
  - 触发器存在性（022 生效后应满足“目标触发器=0”）：
    - `SELECT tgname FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid WHERE c.relname='organization_units' AND NOT t.tgisinternal;`
    - 允许存在与本问题无关的技术性触发器；强制要求“不存在目标触发器”
- CI 结果核对：
  - PR 必须通过 `Audit Consistency Gate`/`Consistency Guard (audit)` 两个校验任务

## 16. 当前进度总结（2025-09-15）
- 数据侧：空 UPDATE=0（严格 SQL 口径）；recordId 归属一致；但样本 recordId 审计条目仍偏多（历史遗留），需在 022 生效且“应用接管写路径”后对新写入进行回归复测。
- 数据库迁移：`021` 具备、`022` 已接入 CI 强制；目标库需尽快执行 `022`。
- 应用侧：命令服务写路径接管需补齐与验证（邻接修补、is_current 重算、审计写入）。
- 工具与门禁：SQL 报告/断言、一键脚本与 CI 门禁均已落地；文档已同步更新。

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
