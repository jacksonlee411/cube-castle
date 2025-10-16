# 86号文档：职位任职管理 Stage 4 实施计划

**版本**: v0.1 草案  
**创建日期**: 2025-10-17  
**维护团队**: 命令服务团队 · 查询服务团队 · 前端团队 · QA 团队 · 架构组  
**状态**: 待评审  
**关联计划**: 80号职位管理方案 · 84号 Stage 2 实施计划（归档） · 85号 Stage 3 执行计划（归档） · 06号集成团队协作日志  
**遵循原则**: `CLAUDE.md` 资源唯一性与跨层一致性 · CQRS 分工（REST 命令 / GraphQL 查询） · Docker 容器化强制

---

## 1. 背景与目标

- 80号方案第7.5节将 Stage 4 定义为“职位任职管理”扩展，需补齐 `position_assignments` 的多重任职、代理任职与历史追踪能力，并提供唯一事实来源的租户隔离校验。  
- 84号计划已在 Stage 2 交付中完成 `position_assignments` 表及基础 Fill/Vacate/Transfer 命令，实现单一任职的端到端流程。  
- 85号计划在 Stage 3 完成编制统计、空缺看板与转移界面，但仍依赖 Stage 4 交付来支持多席位任职、代理任职与完整历史视图。  
- 本计划旨在于两个迭代（共4周）内完成 Stage 4 的全部必需交付，关闭 80号方案中尚未勾选的四项任务，并补齐 1710 行待办的跨租户回归测试。

---

## 2. 唯一事实来源确认

| 主题 | 唯一事实来源 |
|------|--------------|
| Stage 4 范围与勾选项 | `docs/development-plans/80-position-management-with-temporal-tracking.md` 第7.5节、第7.6节 |
| 现有命令 / 查询能力 | `cmd/organization-command-service`、`cmd/organization-query-service` 源码（Stage 2/3 合并版本） |
| 迁移与数据结构 | `database/migrations/044_create_position_assignments.sql`、`045_drop_position_legacy_columns.sql` |
| 当前测试与风险 | `docs/development-plans/06-integrated-teams-progress-log.md` Stage 3 收尾记录 |

> 所有设计、契约、迁移与实现变更必须首先更新上述事实来源；若发现冲突，优先回滚变更并修正唯一事实来源。

---

## 3. 范围界定

### 3.1 目标范围（Included）
1. **Position Assignment 实体完善**  
   - 补充 `assignment_id` 主键、`assignment_type`（PRIMARY/SECONDARY/ACTING）枚举扩展。  
   - 引入 `fte_ratio`、`effective_date`、`end_date`、`acting_until` 等字段，支持部分工时与代理期限。  
   - 增强租户隔离与乐观锁策略（`SELECT ... FOR UPDATE` + `version` 字段）。
2. **多重任职支持（Primary/Secondary）**  
   - REST 命令：`/positions/{code}/assignments` POST/PUT/PATCH，实现主/副任职创建、终止、切换。  
   - GraphQL 查询：`positionAssignments`, `positionTimeline` 扩展返回多重任职节点，支持 `roleFilter`、`assignmentType` 过滤。  
   - 编制统计：更新 `positionHeadcountStats` 计算逻辑，按 `fte_ratio` 统计编制占用。
3. **代理任职（Acting）**  
   - 命令服务新增 `actingAssignments` 操作，允许指定代理期限、原任职保留策略。  
   - GraphQL 返回代理链条，前端详情页显示当前代理、到期提醒。  
   - 审计记录：新增 `assignment_operation` 类型，确保代理启停可追踪。
4. **任职历史追踪**  
   - 时间轴与审计整合：在 `positionTimeline`、`auditHistory` 中串联 Fill/Vacate/Transfer 与 Assignment 记录，支持 `asOfDate`。  
   - 前端时间轴视图：新增“任职历史”页签，展示多重任职/代理切换、导出 CSV。  
   - 跨租户引用回归测试：REST + GraphQL 覆盖 403 `JOB_CATALOG_TENANT_MISMATCH` 与 assignment 交互。

### 3.2 非目标范围（Excluded）
- 员工主数据服务改造（例如 HRIS 员工档案同步）；Stage 4 仅消费现有 `employeeId`。  
- 薪酬、绩效、排班等后续模块。  
- 事件流/消息队列集成：保持同步接口，后续视 Stage 5 规划决议。  
- BI 仪表盘（Grafana、Looker）的深度集成，当前仅提供 CSV 导出与 GraphQL 统计。

---

## 4. 时间线与里程碑（4 周，两个迭代）

| 周次 | 时间窗口 | 核心交付 | 责任团队 | 验收方式 |
|------|----------|----------|----------|----------|
| **Week 0：准备期** | Day 1-2 | 契约评审、迁移差异草案、事实来源复核 | 架构组 + 命令服务 | 完成契约差异草案 PR（OpenAPI/GraphQL） |
| **Week 1：基础能力** | Day 3-7 | `position_assignments` 字段扩展、REST CRUD、仓储单元测试 | 命令服务 + 数据库 | `go test ./cmd/organization-command-service/...` |
| **Week 2：多重任职 & 统计** | Day 8-14 | GraphQL 多重任职查询、`positionHeadcountStats` 调整、前端 Hook 接入 | 查询服务 + 前端 | `npm --prefix frontend run test -- PositionHeadcountDashboard` |
| **Week 3：代理任职 & 时间轴** | Day 15-21 | Acting 命令/查询、时间轴整合、前端任职历史页签 | 命令服务 + 查询服务 + 前端 | Playwright `position-lifecycle.spec.ts` 新增代理场景 |
| **Week 4：验收与回归** | Day 22-28 | 跨租户测试、性能评估、文档归档、实施清单更新 | QA + 架构组 | `make test-integration`、文档签收（80号勾选 Stage 4） |

> 每周周三召开风控例会，周五进行阶段性演示及风险复盘；重大风险需记录在 06 号日志中。

---

## 5. 任务拆解与执行步骤

### 5.1 Week 0 准备
- 运行 `node scripts/generate-implementation-inventory.js`，确认现有 Assignment 相关条目。  
- 审阅 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`：新增 `PositionAssignmentInput`、`AssignmentTypeEnum`、`ActingAssignmentInput` 等契约草案。  
- 草拟 `database/migrations/046_extend_position_assignments.sql`：新增字段、索引、租户约束；同时准备回滚脚本。  
- 更新 80号方案 Stage 4 章节草稿，确保与本计划一致。

### 5.2 Week 1 基础能力
- 实施迁移 046（含 sandbox 演练、回滚验证，输出日志至 `reports/database/assignment-stage4-dryrun-YYYYMMDD.txt`）。  
- 扩展命令服务仓储：`InsertAssignmentTx`、`UpdateAssignmentTx`、`CloseAssignmentTx`，添加幂等校验与 FTE 验证。  
- 新增 REST 端点：  
  - `POST /api/v1/positions/{code}/assignments`（创建主/副/代理任职）  
  - `PATCH /api/v1/positions/{code}/assignments/{assignmentId}`（修改 FTE、代理期限、状态）  
  - `POST /api/v1/positions/{code}/assignments/{assignmentId}/end`（终止任职）  
- 补充单元测试与集成测试，模拟并发/跨租户场景。

### 5.3 Week 2 多重任职与统计
- GraphQL：扩展 `Position`, `PositionTimeline`, `positionAssignments`, `positionHeadcountStats` schema 与 resolver。  
- 新增仓储查询：支持按 `assignmentType`、`effectiveDate`、`employeeId` 过滤，并计算 `fte_ratio`。  
- 前端：  
  - 更新 `useEnterprisePositions`、`usePositionAssignments` Hook。  
  - 在 `PositionDetails`、`PositionSummaryCards` 中展示多席位占用。  
  - 调整 CSV 导出逻辑，保留 FTE 信息。  
- QA：运行 Vitest 组件测试、前端契约测试，确保类型与 GraphQL 响应对齐。

### 5.4 Week 3 代理任职与历史追踪
- 命令服务：实现代理任职逻辑（支持指定代理范围、自动回滚至原主任职），并记录审计日志。  
- 查询服务：  
  - `positionTimeline` 补充代理节点与多重任职并发时间线。  
  - `auditHistory` 输出 assignment 操作详情。  
- 前端：  
  - 新增“任职历史”页签，复用时间轴组件；支持筛选主任职/代理/副任职。  
  - 在 `PositionTransferDialog` 显示当前代理与任职冲突提醒。  
- Playwright：扩充 `tests/e2e/position-lifecycle.spec.ts`，模拟主任职→代理→恢复流程。

### 5.5 Week 4 验收与收束
- 执行跨租户回归测试：REST + GraphQL 触发 `JOB_CATALOG_TENANT_MISMATCH`、`POSITION_ASSIGNMENT_TENANT_MISMATCH`。  
- 性能评估：  
  - GraphQL 统计查询（多任职 + 代理） P95 < 200ms。  
  - REST 任职操作 P95 < 250ms。  
- 文档更新：  
  - 80号方案 Stage 4 勾选并新增验收小节。  
  - 06号日志补充 Stage 4 进度条目。  
  - 更新实现清单、API 差异报告（`reports/contracts/position-assignment-diff.md`）。  
- 准备归档：本计划达成后移入 `docs/archive/development-plans/`。

---

## 6. 质量与验收标准

| 类别 | 验收项 | 说明 |
|------|--------|------|
| 功能 | ✅ 多重任职：同一职位允许 1 个 PRIMARY + N 个 SECONDARY，FTE 累加 ≤ `headcountCapacity` | REST/GraphQL 均返回正确占用 |
| 功能 | ✅ 代理任职：支持设定有效期，超期自动恢复主任职；支持手动终止 | Playwright 代理场景通过 |
| 功能 | ✅ 任职历史：时间轴展示 Fill/Vacate/Transfer/Assignment 事件，支持 CSV 导出 | 前端页面 + GraphQL 查询一致 |
| 安全 | ✅ 租户隔离：所有命令/查询携带 `X-Tenant-ID`，跨租户操作返回 403 | 集成测试覆盖 |
| 性能 | ✅ `positionHeadcountStats` / `positionAssignments` 查询 P95 < 200ms | 运行 `go test -run TestPerformance...` 或专用脚本 |
| 质量 | ✅ 单元测试覆盖率：命令/查询服务 ≥ 80%；前端 Vitest 关键组件全通过 | CI 记录 |
| 文档 | ✅ 契约、实现清单、计划文档同步；80号 Stage 4 勾选完成 | 文档审查记录 |

---

## 7. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 任职与编制计算不一致 | 高 | 中 | 引入 FTE 校验、统计单元测试，复核 SQL 聚合并与业务方确认 | 
| 代理任职与主任职冲突 | 中 | 中 | 命令服务串行化同一职位操作，加入冲突检测与幂等键 | 
| 多租户引用错误 | 高 | 低 | 针对分类、职位、员工外键增加租户校验与回归测试 | 
| 回滚复杂度高 | 中 | 低 | 046 迁移提供回滚脚本，执行前导出快照并记录日志 | 
| 前端交互复杂导致延期 | 中 | 中 | 提前完成 UI 原型评审，复用现有时间轴组件，页面渐进式上线 | 

---

## 8. 协作与沟通机制

- 周会：周一计划同步、周三风控、周五演示 / 回顾。  
- 责任人：  
  - 命令服务负责人：负责迁移、REST、事务与审计。  
  - 查询服务负责人：负责 GraphQL Schema、Resolver、性能。  
  - 前端负责人：负责 Hook、组件、E2E。  
  - QA 负责人：负责测试计划、自动化执行。  
  - 架构组：守护唯一事实来源、评审契约与迁移。  
- 所有节点、风险、决策需记录在 06号进展日志；关键输出（迁移日志、测试报告）归档至 `reports/`。

---

## 9. 交付与归档清单

- [ ] 046 迁移脚本与回滚脚本（含演练日志）  
- [ ] OpenAPI / GraphQL 契约更新及差异报告  
- [ ] 命令/查询服务代码与测试（Go）  
- [ ] 前端 Hook、组件、Playwright 场景与文档  
- [ ] 编制统计与任职历史文档更新（80号方案、06号日志）  
- [ ] 跨租户集成测试脚本与结果归档  
- [ ] 本计划完成后迁移至 `docs/archive/development-plans/`

若以上条目全部完成且验收标准达成，则宣布 Stage 4 交付完成，并在 80号方案中勾选“Position Assignment 实体 / 多重任职 / 代理任职 / 任职历史追踪”四项。

---

## 10. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v0.1 | 2025-10-17 | 初始草案，提交评审 | 项目智能助手 |

