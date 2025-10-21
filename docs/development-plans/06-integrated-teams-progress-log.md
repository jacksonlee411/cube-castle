# 06号文档：集成团队协作进展日志

> **更新时间**：2025-10-20 23:30
> **负责人**：架构组 + 前端团队 · 职位域
> **关联计划**：92号《职位管理二级导航实施方案》 v2.2 · 87号《时态字段命名一致性决策文档》

---

## 1. 当前活跃工作

### 1.1 最近完成的里程碑（2025-10-15 ~ 2025-10-20）

- ✅ **92号职位管理二级导航**：Phase 0-4 全部完成，312px 灰底侧栏、权限过滤、Job Catalog 四层页面、GraphQL/REST 集成、Playwright E2E、文档对齐全部交付
- ✅ **93号职位详情多页签**：六个页签布局、时间轴侧栏、审计页签接入完成并归档
- ✅ **97号 TypeScript 错误修复**：Phase 0-4 完成，Canvas Kit 迁移、枚举修复、Storybook 隔离，`npm run build` 通过
- ✅ **101~104号计划收口**：Position Playwright hardening、PositionForm 抽象、组件目录整理、设计规范 v0.1 全部归档
- ✅ **101~104号计划归档确认**（2025-10-21）：文档迁移至 `docs/archive`，88号差距分析与 99号收口指引同步更新

### 1.2 当前进行中

- 🚧 **文案国际化**：导航配置及表单提示接入 i18n 方案（P1，截止 2025-10-24）
- 🚧 **性能监控**：Job Catalog 缓存刷新策略优化、懒加载优化监控（按需恢复）
- 🚧 **删除能力补齐**：`useJobCatalogMutations` 仍缺删除 Hook（Phase 3 命令测试跟踪中）

---

## 2. 页面验证步骤（参考）

详细验证步骤参考 92号文档附录 A。快速验证流程：

```bash
# 1. 启动环境
make docker-up && make run-dev && make frontend-dev

# 2. 生成开发令牌
make jwt-dev-mint

# 3. 访问前端
open http://localhost:5173

# 4. 运行回归测试
npm --prefix frontend run test -- --run src/features/job-catalog
npm --prefix frontend run test:e2e -- tests/e2e/job-catalog-secondary-navigation.spec.ts
```

### 2.2 组织软删除后的恢复步骤

当组织单元被软删除（`status='DELETED'`）后，所有历史版本均标记为删除状态，时间轴对外为空。如需恢复，可通过以下方式重建：

**方案A：通过 REST API 重新创建组织**
```bash
# 1. 重新创建组织（会生成新的首条版本）
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "DEPT-001",
    "name": "已恢复部门",
    "type": "DEPARTMENT",
    "effectiveDate": "2025-10-20",
    "parentCode": "DIV-001"
  }'

# 2. 如需恢复历史版本，逐条通过版本接口插入
curl -X POST http://localhost:9090/api/v1/organization-units/DEPT-001/versions \
  -H "Authorization: Bearer $TOKEN" \
  -H "If-Match: <latest-etag>" \
  -H "Content-Type: application/json" \
  -d '{
    "effectiveDate": "2025-01-01",
    "name": "历史部门名称",
    "parentCode": "DIV-001"
  }'
```

**方案B：数据库层恢复（需谨慎使用）**
```sql
-- 仅在确认无数据冲突时使用，需在维护窗口执行
BEGIN;
-- 将软删除的版本恢复为 ACTIVE 状态
UPDATE organization_units
SET status = 'ACTIVE'
WHERE tenant_id = '<tenant-uuid>'
  AND code = 'DEPT-001'
  AND status = 'DELETED';

-- 触发时间轴重算，重新设置 is_current 和 end_date
-- 通过运维接口或手动调用 RecalculateTimeline
COMMIT;
```

**方案C：使用时间轴重算服务**
```bash
# 恢复状态后，调用运维接口触发时间轴重算
curl -X POST http://localhost:9090/api/v1/operational/tasks/recalculate-timeline \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenantId": "<tenant-uuid>",
    "code": "DEPT-001"
  }'
```

**注意事项**：
- 方案A最安全，但会重新生成 `record_id`，审计链路中断
- 方案B和C保留原始 `record_id`，但需确保无并发写入冲突
- 恢复后建议立即检查时间轴一致性：`/api/v1/operational/health`
- 所有恢复操作均应记录在审计日志中，包含操作原因

**相关文档**：
- 时态时间轴一致性指南：`docs/architecture/temporal-timeline-consistency-guide.md`
- 停用与删除治理计划：`docs/archive/development-plans/13-organization-suspend-delete-governance.md`
- 70号调查报告：`docs/archive/development-plans/70-temporal-timeline-lifecycle-investigation.md`

---

## 3. 待办任务清单

| 优先级 | 项目 | 负责人 | 截止 | 状态 |
|--------|------|--------|------|------|
| P1 | 文案国际化：导航配置及表单提示接入 i18n | 前端国际化负责人 | 2025-10-24 | 🚧 进行中 |
| P2 | Job Catalog 删除能力补齐 | 命令服务团队 | 2025-10-30 | 📋 待启动 |
| P2 | 缓存刷新策略深度优化（按需） | 前端性能组 | 待定 | 📊 监控中 |
| P2 | 职级详情页上下文依赖优化 | 前端架构组 | 待定 | 🔍 评估中 |
| ✅ | **87号时态字段命名一致性实施** | 数据库 + 全栈团队 | 2025-10-21 | ✅ 047 迁移+全栈改造已合并，命名统一为 effective_date |

> **说明**：87号方案已完成架构评审，建议纳入下一个迭代执行（详见第6节）

---

## 4. 风险与依赖（活跃项）

### 4.1 当前风险

- 🔶 **国际化延期风险**：P1 任务截止 2025-10-24，需确认资源到位
- 🔸 **职级详情页直接访问体验**：缺少 `roleCode` 上下文时用户体验不佳，需评估 URL 参数方案
- 🔷 **删除能力缺失**：Job Catalog 仍无法删除记录，影响完整 CRUD 闭环

### 4.2 依赖项

- ✅ **后端写接口联调**：已完成，REST `/api/v1/job-*` PUT 接口与 PBAC 权限映射就绪
- ✅ **Playwright E2E 覆盖**：已完成，真实后端联调与 Mock 守护双模式验证通过
- 🔶 **87号决策依赖**：需架构委员会召集决策会议，确认统一时态字段命名方案

---

## 5. 已关闭事项（归档记录）

以下事项已完成并归档至对应文档：

- ✅ 92号 Phase 0-4 全部验收通过（2025-10-20）
- ✅ 93号职位详情多页签验收归档（2025-10-19）
- ✅ 97号 TypeScript 错误修复 Phase 0-4 完成（2025-10-20）
- ✅ 101~104号计划全部归档（2025-10-20）
- ✅ 前后端联调：REST 更新接口 + 权限校验（2025-10-22）
- ✅ Playwright 脚本覆盖二级导航与职类 CRUD（2025-10-24）
- ✅ 设计评审：Job Catalog 列表/详情视觉稿确认（2025-10-20）
- ✅ **70号组织时间轴全生命周期调查报告归档**（2025-10-20）：完成13个场景覆盖性分析，确认时间轴重算机制符合实施指南约束，组织恢复步骤已补充至本文档 Section 2.2
- ✅ **95号 Status Fields Review 归档**（2025-10-20）：完成状态字段实现范围调查，架构决策确认不扩展五态避免过度设计，技术债务转P3清单

---

## 6. 87号时态字段命名一致性方案评审结果

### 6.1 评审概要

**评审日期**：2025-10-20
**评审方**：架构组 + Claude Code 助手
**文档版本**：87号文档 v1.0
**评审状态**：✅ **已完成（2025-10-21 迁移与代码同步交付）**

### 6.2 方案背景

**问题**：此前任职记录（`position_assignments`）使用 `start_date`，而组织架构、职位主数据、Job Catalog 均为 `effective_date`，导致跨层命名不一致。现已通过 047 迁移与代码同步改造统一为 `effective_date`。

**影响**：
- 违反 CLAUDE.md 最高优先级原则（资源唯一性与跨层一致性）
- 违背 80号方案承诺（"完全复用组织架构模式"）
- 增加查询复杂度、API 响应不一致、代码维护成本

**建议方案**：统一为 `effective_date`（唯一合理选择）

### 6.3 评审结论

#### 综合评分：⭐⭐⭐⭐（4/5）

| 评审项 | 评分 | 说明 |
|--------|------|------|
| 问题诊断 | ⭐⭐⭐⭐⭐ | 证据完整，根因清晰 |
| 架构一致性 | ⭐⭐⭐⭐⭐ | 完全符合 CLAUDE.md 最高优先级原则 |
| 方案设计 | ⭐⭐⭐⭐ | 结论正确，但缺少替代方案对比 |
| 实施步骤 | ⭐⭐⭐⭐ | 清晰可行，但缺少测试与验证细节 |
| 风险评估 | ⭐⭐⭐ | 基本充分，但需补充生产环境影响分析 |
| 工作量评估 | ⭐⭐⭐ | 基本合理但偏乐观，建议增加20% |

#### 核心优点

1. ✅ **符合最高优先级原则**：完全对齐 CLAUDE.md 资源唯一性与跨层一致性要求
2. ✅ **兑现架构承诺**：实现 80号方案"完全复用组织架构模式"承诺
3. ✅ **长期收益显著**：统一命名标准，可维护性、可扩展性大幅提升
4. ✅ **一次性成本可控**：2-3个工作日可完成
5. ✅ **时机良好**：Stage 3 刚完成，Stage 4 未启动，改动窗口最佳

#### 需补充内容

1. **补充生产环境迁移计划**（新增第11节）：
   - 评估现有数据量
   - 明确迁移窗口（停机或热迁移）
   - 提供分批迁移方案（若数据量大）

2. **补充测试清单**（扩展4.5节）：
   - 列出需要更新的测试文件清单
   - 明确回归测试验证场景（fill/vacate/transfer/timeline）
   - 前端缓存清理验证

3. **补充下游影响分析**（扩展3.4节）：
   - 检查是否有外部系统依赖此字段
   - 评估 API 契约变更通知策略

4. **调整工作量评估**（更新4.5节）：
   - 从 16小时 调整为 20-24小时
   - 增加测试与验证时间

### 6.4 实施结果（2025-10-21）

- ✅ 数据库：新增 `047_rename_position_assignments_start_date.sql` 迁移，重命名 `start_date` → `effective_date` 并重建唯一索引、检查约束。
- ✅ 命令服务：更新实体、仓储与服务层，审计日志字段同步调整为 `assignmentEffective`。
- ✅ 查询服务：模型/仓储/Resolver 同步改造，排序字段改为 `EFFECTIVE_DATE` 并兼容历史枚举。
- ✅ 契约同步：`docs/api/schema.graphql`、`docs/api/openapi.yaml` 均已对齐字段命名与排序枚举。
- ✅ 前端：类型、GraphQL 查询、生成代码及 E2E 基线数据全部改为 `effectiveDate`。

### 6.5 测试与验证

- `go test ./cmd/organization-command-service/...`
- `go test ./cmd/organization-query-service/...`
- `npm --prefix frontend run typecheck`

上述校验于 2025-10-21 完成，结果全部通过。

### 6.6 决策与归档状态

**决议状态**：✅ 架构组异步确认完成，后续生产执行按 87 号文档第 11 节迁移流程操作。

**下一步**：监控生产部署窗口及外部集成方升级情况，如需发布额外通知请在 87 号文档补充记录。

### 6.7 相关文档

- 📄 87号文档：`docs/development-plans/87-temporal-field-naming-consistency-decision.md`
- 📄 80号文档：`docs/development-plans/80-position-management-with-temporal-tracking.md`
- 📄 86号文档：`docs/development-plans/86-position-assignment-stage4-plan.md`
- 📄 项目原则：`CLAUDE.md` - 资源唯一性与跨层一致性原则

### 6.8 86号计划后续任务（2025-10-20）

- **完成 048 迁移演练**：记录执行与回滚日志至 `reports/position-stage4/`，确认操作手册可复用。
- **跨租户验证脚本**：补全 REST/GraphQL 跨租户校验脚本并在 CI 侧留存执行截图/结果。
- **计划文档归档**：更新 86 号计划完成项后移动至 `docs/archive/development-plans/`，确保引用唯一性。
- **QA 场景补强**：协调 QA 在 `tests/e2e/` 添加代理任职自动恢复端到端脚本。
- **监控指标暂缓**：根据决策，`position_assignment_acting_total` 等指标暂不实施，待后续需求触发。

---

## 7. 更新日志

| 日期 | 更新内容 | 负责人 |
|------|----------|--------|
| 2025-10-21 14:30 | 记录 Stage 4 047 迁移演练 T-0/T+1 凭证（见 reports/position-stage4/*20251021*），供上线前复核 | 架构组 + Claude Code |
| 2025-10-21 13:00 | 87 号档案补充 §12（与 86 号收尾联动迁移策略），同步更新 99 号指引 | 架构组 + Claude Code |
| 2025-10-21 12:30 | 87 号决策文档归档至 `docs/archive/development-plans/87-temporal-field-naming-consistency-decision.md` | 架构组 + Claude Code |
| 2025-10-21 12:00 | 完成 87 号命名统一实施并同步相关文档/契约 | 数据库 + 全栈团队 |
| 2025-10-20 23:30 | 记录70号、95号文档归档事件至已关闭事项清单 | 架构组 + Claude Code |
| 2025-10-20 23:00 | 补充组织软删除恢复步骤（响应70号调查报告行动项2） | 架构组 + Claude Code |
| 2025-10-20 21:30 | 清理已完成任务，添加87号方案评审结果 | 架构组 + Claude Code |
| 2025-10-20 19:45 | 更新92号计划最新进展，记录101~104归档 | 前端团队 |
| 2025-10-20 15:20 | 复核 87 号迁移后回归：执行 `go test ./cmd/organization-command-service/...`、`go test ./cmd/organization-query-service/...`、`npm --prefix frontend run typecheck` | 架构组 + Claude Code |
| 2025-10-19 16:00 | 记录93号职位详情多页签验收完成 | 前端团队 |

---

**下一次更新触发条件**：
- 87号方案决策会议完成并形成决议；或
- 文案国际化任务完成；或
- Job Catalog 删除能力补齐启动
