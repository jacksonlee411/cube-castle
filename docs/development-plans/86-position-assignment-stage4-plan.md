# 86号文档：职位任职管理 Stage 4 增量计划

**版本**: v0.3 修订版  
**创建日期**: 2025-10-17  
**最新更新**: 2025-10-19  
**维护团队**: 命令服务团队 · 查询服务团队 · 前端团队 · QA 团队 · 架构组  
**状态**: 待复审（修订稿）  
**关联计划**: 80号职位管理方案 · 84号 Stage 2 实施计划（归档） · 85号 Stage 3 执行计划（归档） · 06号集成团队协作日志  
**遵循原则**: `CLAUDE.md` 资源唯一性与跨层一致性 · `AGENTS.md` 开发前必检规范 · CQRS 分工（命令 REST / 查询 GraphQL） · Docker 容器化强制

---

## 1. 背景与目标

- Stage 2（见 84 号计划）已经落地 `position_assignments` 表、Fill/Vacate/Transfer 全链路、命令/查询层仓储，以及前端任职展示，确保职位任职的唯一事实来源。  
- Stage 3（见 85 号计划）已完成编制统计、空缺看板、转移界面及 Playwright 验收。  
- 80 号方案第 7.5 节提出的 Stage 4 剩余四项勾选（Position Assignment 实体、Multiple Assignments、Acting、History）仍需针对“高级任职管理”做增量完善，特别是代理自动恢复、专用任职 API、历史视图与跨租户验证。  
- 本修订稿在复盘现有实现后，仅聚焦缺失能力与质量补强，避免重复造轮子，并将周期压缩为 2 周（加 1 周缓冲）。

---

## 2. 现有能力复盘（Stage 2/3）

| 能力 | 当前状态 | 事实来源 |
|------|----------|----------|
| Assignment 表结构 | ✅ `assignment_id` 主键、`assignment_type` (PRIMARY/SECONDARY/ACTING)、`fte`、`effective_date/end_date`、租户外键、唯一约束 | `database/migrations/047_rename_position_assignments_start_date.sql` |
| 命令服务仓储与服务 | ✅ Create/List/Close/FTE 聚合、Fill/Vacate/Transfer 写入任职历史 | `cmd/organization-command-service/internal/repository/position_assignment_repository.go`、`position_service.go` |
| GraphQL 查询 | ⚠️ `currentAssignment` / `assignmentHistory` 已上线；缺少高级过滤、时间轴聚合与性能基线 | `cmd/organization-query-service/internal/model/models.go`、`cmd/organization-query-service/internal/repository/postgres_positions.go`、`docs/archive/development-plans/89-position-crud-verification-report.md` |
| 前端展示 | ✅ `PositionDetails` 任职列表/历史，`PositionDashboard` 读取 GraphQL 数据 | `frontend/src/features/positions` |
| 编制统计 | ✅ `positionHeadcountStats` 复用 FTE 计算并驱动 `PositionHeadcountDashboard` | `cmd/organization-query-service`、`frontend` |

> ✅ GraphQL 查询服务可正常运行，Stage 4 聚焦在现有能力上扩展过滤、历史聚合与性能观测，而不是阻塞性修复。

---

## 2.1 前置增强事项（GraphQL & 数据访问）

| 项目 | 目标 | 责任团队 | 验收标准 |
|------|------|----------|----------|
| GraphQL 查询增强 | 扩展 `cmd/organization-query-service/internal/model` 与 `repository/postgres_positions.go`，支持任职过滤（类型/状态/日期范围）、分页与租户隔离 | 查询服务团队 | `go test ./cmd/organization-query-service/...` 通过；`make run-dev` 下 GraphQL 查询 P95 < 250ms |
| 时间轴整合 | 在 `GetPositionTimeline` 聚合 Acting/Primary 任职节点，输出时间顺序与标识 | 查询服务团队 · 架构组 | GraphQL `positionTimeline` 返回任职节点；前端时间轴验收通过 |
| 事实来源同步 | 更新 `docs/api/schema.graphql` 注释，执行 `node scripts/generate-implementation-inventory.js` 并记录差异 | 架构组 | 契约、实现、实现清单一致；06 号日志留存 |

### 前置增强步骤（预计 4-6 小时）
1. **查询扩展**（1h）：在 `repository/postgres_positions.go` 添加任职过滤、分页参数与租户校验。  
2. **模型调整**（1h）：复用 `AssignmentHistoryField`，保证空值语义，同时承载过滤结果缓存。  
3. **时间轴组装**（1h）：在 `GetPositionTimeline` 输出 Acting/Primary 节点与事件排序。  
4. **性能校准**（1h）：运行 `make run-dev`，采集 GraphQL 查询延迟，生成 `reports/position-stage4/latency-baseline.md`。  
5. **契约同步**（≤2h）：更新 GraphQL Schema 注释、实现清单并记录 06 号日志时间戳。

---

## 3. 差距与目标范围

### 3.1 差距分析

| 分类 | 现状 | 差距 | Stage 4 增量目标 |
|------|------|------|------------------|
| 代理任职生命周期 | Fill 可创建 `assignment_type=ACTING`，但未自动恢复 | 代理到期需人工 Vacate，缺少自动化 & 提醒 | 实现代理到期自动恢复、提醒通知、冲突校验 |
| 任职操作接口 | Fill/Vacate/Transfer 混合处理 Assignment & Position | 缺少任务专用端点、难以复用 | 提供 `/assignments` 专用 REST API，同时复用现有 Fill/Vacate 流程 |
| 任职历史视图 | 前端展示列表，但无筛选/导出/时间轴增强 | 缺少可视化时间线与更细粒度过滤 | 扩充“任职历史”页签，加时间轴、筛选、CSV 导出 |
| 租户一致性测试 | 80 号第 7.6 节待办、1710 行回归测试缺失 | 现有 Playwright 未覆盖跨租户/非法引用 | 完成交叉租户集成测试脚本与自动化执行 |
| 运营监控 | 缺少代理队列/统计监控 | 风险不可见 | 增补 Prometheus 指标和日志辅助 |

### 3.2 本次范围（Included）
1. **代理任职自动化**：到期恢复主任职、提前提醒、冲突检测、审计日志补强。  
2. **任职专用 API**：在保持 Fill/Vacate 兼容的前提下，新增 `/api/v1/positions/{code}/assignments/*` 端点，对外暴露 CRUD 与分页查询。  
3. **历史视图增强**：GraphQL & 前端支持按 AssignmentType/Status/日期筛选，提供时间轴可视化、CSV 导出、代理标识。  
4. **租户隔离回归**：补齐 REST/GraphQL 跨租户测试脚本、CI 集成，覆盖 403 `JOB_CATALOG_TENANT_MISMATCH` 与 `POSITION_ASSIGNMENT_TENANT_MISMATCH`。  
5. **监控告警**：增加代理任职到期计数、滞留检测指标，接入日志与 dashboard。

### 3.3 非目标范围（Excluded）
- 员工主数据服务改造、外部 HRIS 集成。  
- 薪酬、绩效等后续模块。  
- 组织事件异步总线（保留同步方式）。  
- Grafana/Looker 深度可视化（仅提供指标与 CSV）。

---

## 4. 开发前必检（强制）

在开展任何 Stage 4 工作前，必须一次性执行并归档以下命令：

```bash
cd /home/shangmeilin/cube-castle

# 1. 实现清单核对
node scripts/generate-implementation-inventory.js | grep -i "position assignment"

# 2. IIG 护卫检查
node scripts/quality/iig-guardian.js "Position Assignment Stage4" --guard

# 3. Stage 2/3 实现审计
grep -A40 "position_assignments" database/migrations/044_create_position_assignments.sql
rg "assignment" cmd/organization-command-service/internal -n
rg "positionAssignments" cmd/organization-query-service/internal -n

# 4. 差距报告初始化
mkdir -p reports/position-stage4
echo "Stage4 差距分析（现状 vs 目标）" > reports/position-stage4/gap-analysis.md
```

执行结果需附在 06 号进展日志 Stage 4 小节。

---

## 5. 时间线与里程碑（2 周 + 1 周缓冲）

| 周次 | 核心目标 | 责任团队 | 产出物 & 验收 |
|------|----------|----------|----------------|
| **Week 1** | 代理任职自动化 & 契约对齐 | 命令服务 · 数据库 · 架构 | 048 迁移、REST `/assignments`、OpenAPI/Schema 更新、单元测试、审计日志 |
| **Week 2** | 历史视图增强 & 跨租户测试 | 查询服务 · 前端 · QA · 运维 | GraphQL 扩展、前端时间轴、调度集成、Playwright/集成测试 |
| **Week 3 (缓冲)** | 监控指标 & 文档归档 | 全员 | Prometheus 指标、调度运行日志、文档同步、计划归档 |

每周周三风控例会、周五演示与风险复盘；重大事项写入 06 号日志。

---

## 6. 任务拆解

### API 契约定义（Stage 4 增量）
- **REST — `/api/v1/positions/{code}/assignments` 套件**（OpenAPI 将新增/更新以下条目，均要求 `position:assignments:write` 或 `position:assignments:read` 权限）：  
  - `GET /api/v1/positions/{code}/assignments`: 查询当前与历史任职，支持 `assignmentTypes[]`、`assignmentStatus`、`asOfDate`、`includeHistorical`、分页参数。响应主体为 `PositionAssignmentListResponse`，字段沿用 Stage 2/3 输出，新增 `actingUntil`、`autoRevert`、`reminderSentAt`。  
  - `POST /api/v1/positions/{code}/assignments`: 创建任职。请求体需提供 `employeeId`、`employeeName`、`assignmentType`、`effectiveDate`、`fte`，可选 `actingUntil`、`autoRevert`、`notes`。成功返回 201 + 新建记录。  
  - `PATCH /api/v1/positions/{code}/assignments/{assignmentId}`: 更新任职（调整 `fte`、`actingUntil`、`autoRevert`、`notes`），返回 200。  
  - `POST /api/v1/positions/{code}/assignments/{assignmentId}/close`: 结束任职。请求体包含 `endDate`、可选 `notes`，返回 200 并写入审计。  
  - 所有端点必须校验租户一致性（`tenantId` header/claims），返回标准错误码：`403 JOB_CATALOG_TENANT_MISMATCH`、`409 POSITION_ASSIGNMENT_CONFLICT`、`422 POSITION_ASSIGNMENT_VALIDATION_FAILED`。
- **GraphQL — `docs/api/schema.graphql` 增量**：  
  - `positionAssignments(positionCode: PositionCode!, filter: PositionAssignmentFilterInput, pagination: PaginationInput, sorting: [PositionAssignmentSortInput!]): PositionAssignmentConnection!` 新增 filter 字段：`assignmentTypes: [PositionAssignmentType!]`, `status: PositionAssignmentStatus`, `dateRange: DateRangeInput`, `includeActingOnly: Boolean`.  
  - `type PositionAssignment` 新增只读字段：`actingUntil: Date`, `autoRevert: Boolean!`, `reminderSentAt: DateTime`.  
  - `type PositionTimelineEntry` 增补 `assignmentType: PositionAssignmentType`、`assignmentStatus: PositionAssignmentStatus`，并允许 `timelineCategory: POSITION_ASSIGNMENT`.  
  - `type PositionAssignmentAudit`（新）用于 CSV 导出：包含 `assignmentId`, `eventType`, `effectiveDate`, `endDate`, `actor`, `changes`.  
  - 权限要求：查询需 `position:assignments:read` scope，导出需额外 `position:assignments:audit`.

### Week 1 — 代理任职自动化 & 契约对齐
- **数据库**：交付 `048_extend_position_assignments.sql`（及回滚脚本），新增 `acting_until DATE`, `auto_revert BOOLEAN DEFAULT false`, `reminder_sent_at TIMESTAMPTZ`，并更新索引/校验。输出演练日志与延迟评估。  
- **命令服务**：实现上述 REST 契约，对接 `PositionAssignmentRepository`，保持 Fill/Vacate 兼容，新增幂等锁与审计事件。  
- **自动化任务**：实现代理到期扫描器（使用 `OperationalScheduler` 任务定义），支持重试、失败告警、审计写入。  
- **单元与契约测试**：扩展 `position_handler_test.go`、`assignment_repository_test.go`，覆盖冲突/FTE 验证；新增 OpenAPI contract tests。  
- **契约同步**：更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，运行 `node scripts/generate-implementation-inventory.js` 并在 06 号日志记录差异。

### Week 2 — 历史视图增强 & 跨租户测试
- **GraphQL 查询服务**：落地前置增强事项，提供过滤、分页、时间轴整合及性能基线；新增 `positionAssignmentAudit` 查询导出。  
- **前端**：在 `frontend/src/features/positions` 新增任职历史页签、时间轴视图、过滤器与 CSV 导出；更新 `PositionTransferDialog`、`PositionSummaryCards` 展示代理提醒。  
- **命令服务调度**：将代理恢复任务接入 `OperationalScheduler` 配置（默认每日 02:00），提供手动触发脚本与日志归档。  
- **质量验证**：编写 Playwright 场景（代理创建→到期→恢复→时间轴验证）与 REST/GraphQL 跨租户脚本，纳入 `make test-integration`。

### Week 3 — 缓冲 & 收尾
- 观察自动恢复任务运行（收集调度日志、Prometheus 指标），若异常则回滚或调整。  
- 完成监控与告警文档（`docs/development-tools/position-assignment-monitoring.md`），更新 80 号方案与 06 号日志。  
- 整理 API 契约差异报告、实现清单、回归测试记录，并准备归档。  
- 将计划移动至 `docs/archive/development-plans/86-position-assignment-stage4-plan.md`。

---

## 7. 质量与验收标准

| 类别 | 验收标准 |
|------|----------|
| 功能 | 代理任职自动恢复 + 提醒日志；专用任职 API 通过 REST 集成测试；时间轴展示主任职/副任职/代理状态。 |
| 数据一致性 | Acting 到期后 FTE 回落，`HeadcountInUse` 与 `positionHeadcountStats` 同步；跨租户操作返回 403。 |
| 性能 | 任职 API P95 < 200ms；代理自动化任务执行 < 2s/1000条；GraphQL 过滤查询 P95 < 250ms。 |
| 测试 | `go test ./cmd/organization-command-service/...` 覆盖率 ≥ 80%；`npm --prefix frontend run test -- PositionDetails`；Playwright Acting 场景通过；跨租户脚本纳入 CI。 |
| 文档 | 80号 Stage 4 勾选完成；06 号日志更新；实现清单/契约差异/监控文档同步。 |
| 监控 | 新增 `position_assignment_acting_total` 等指标并接入报警；运行记录归档到 `reports/position-stage4/`. |

---

## 8. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 自动恢复误触发 | 高 | 中 | 双重条件校验（到期 + 当前状态），留存手动回滚脚本与审计确认。 |
| 任职 API 与 Fill/Vacate 冲突 | 中 | 中 | 通过 `/assignments` 封装现有逻辑，设置特性开关逐步启用并监控审计事件。 |
| 调度器集成不稳定 | 中 | 中 | 在 `OperationalScheduler` 引入幂等锁、失败重试、Prometheus 告警，首周每日人工巡检。 |
| 契约漂移 | 中 | 中 | 每次变更前更新 OpenAPI/GraphQL 并执行实现清单脚本，CI 增加契约 diff 校验。 |
| 跨租户脚本复杂 | 中 | 中 | 先在 sandbox 演练并记录结果，再纳入 CI 分阶段执行。 |

---

## 9. 协作机制

- **例会**：周一计划同步、周三风控、周五演示。  
- **责任人**：  
  - 命令服务负责人：迁移、REST、定时任务。  
  - 查询服务负责人：GraphQL 扩展、性能。  
  - 前端负责人：时间轴与导出。  
  - QA 负责人：集成/Playwright/跨租户脚本。  
  - 架构组：事实来源守护、评审、质量门禁。  
- **文档记录**：所有关键决策、测试结果、监控数据写入 06 号日志 & `reports/position-stage4/`。

---

## 10. 交付与归档清单

- [ ] 048 迁移 & 回滚脚本 + 演练日志  
- [ ] 任职专用 REST API 代码与测试  
- [ ] 代理自动恢复任务（含 `OperationalScheduler` 集成）、日志、监控指标  
- [ ] GraphQL & 前端任职历史增强 + Playwright 场景  
- [ ] 跨租户回归测试脚本（REST/GraphQL）及 CI 集成  
- [ ] 契约与文档同步（OpenAPI/GraphQL、实现清单、80 号 Stage 4 勾选、06 号日志、监控指南）  
- [ ] 计划归档（完成后移动至 `docs/archive/development-plans/`）

---

## 11. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v0.3 | 2025-10-19 | 新增 API 契约定义、更新迁移编号与调度集成、补充风险缓解 | 项目智能助手 |
| v0.2 | 2025-10-17 | 根据 06 号评审意见修订，聚焦增量能力与差距分析 | 项目智能助手 |
| v0.1 | 2025-10-17 | 初始草案（已废弃） | 项目智能助手 |
