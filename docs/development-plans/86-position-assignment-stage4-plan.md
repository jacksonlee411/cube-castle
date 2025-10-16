# 86号文档：职位任职管理 Stage 4 增量计划

**版本**: v0.2 修订版  
**创建日期**: 2025-10-17  
**最新更新**: 2025-10-17  
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
| Assignment 表结构 | ✅ `assignment_id` 主键、`assignment_type` (PRIMARY/SECONDARY/ACTING)、`fte`、`start_date/end_date`、租户外键、唯一约束 | `database/migrations/044_create_position_assignments.sql` |
| 命令服务仓储与服务 | ✅ Create/List/Close/FTE 聚合、Fill/Vacate/Transfer 写入任职历史 | `cmd/organization-command-service/internal/repository/position_assignment_repository.go`、`position_service.go` |
| GraphQL 查询 | ✅ `positionAssignments`、`assignmentHistory`、`PositionAssignmentType` 枚举、租户上下文解析 | `docs/api/schema.graphql`、`cmd/organization-query-service/internal/repository/postgres_positions.go` |
| 前端展示 | ✅ `PositionDetails` 任职列表/历史，`PositionDashboard` 读取 GraphQL 数据 | `frontend/src/features/positions` |
| 编制统计 | ✅ `positionHeadcountStats` 复用 FTE 计算并驱动 `PositionHeadcountDashboard` | `cmd/organization-query-service`、`frontend` |

> 结论：Stage 4 仅需补齐先进管理场景与质量治理，不再重新定义实体或基础 CRUD。

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
| **Week 1** | 代理任职自动化 & 任职 API | 命令服务 · 数据库 · 架构 | 046 迁移、REST 端点、单元测试、审计日志 |
| **Week 2** | 历史视图增强 & 跨租户测试 | 查询服务 · 前端 · QA | GraphQL 扩展、前端时间轴、Playwright/集成测试 |
| **Week 3 (缓冲)** | 监控指标 & 文档归档 | 全员 | Prometheus 指标、脚本、文档同步、计划归档 |

每周周三风控例会、周五演示与风险复盘；重大事项写入 06 号日志。

---

## 6. 任务拆解

### Week 1 — 代理任职自动化 & 任职 API
- **数据库**：`046_extend_position_assignments.sql` 增加 `acting_until`, `auto_revert`, `reminder_sent_at` 等字段（含回滚脚本、演练日志）。  
- **命令服务**：  
  - 新增 `/positions/{code}/assignments` 系列端点（create/update/end/list），并与现有 Fill/Vacate 保持幂等。  
  - 实现代理到期扫描器（定时任务/函数），自动将 Acting 记录转换为 Ended 并恢复主任职，记录审计。  
  - 扩展 `SumActiveFTE` 逻辑以支持 Acting 结束后的即刻更新。  
  - 单元测试覆盖冲突检测（主任职同时存在 Acting）、FTE 校验、自动恢复路径。  
- **监控初步**：埋点 Prometheus Counter/Gauge，记录代理即将到期数量。

### Week 2 — 历史视图增强 & 跨租户测试
- **GraphQL**：  
  - 扩展 `positionAssignments` Filter（支持 `assignmentTypes`, `status`, `dateRange`）。  
  - 在 `positionTimeline` 中插入 Assignment 节点（含 Acting 标识）。  
  - 输出 `PositionAssignmentAudit` 以支持 CSV 导出。  
- **前端**：  
  - 新增“任职历史”页签，使用时间轴组件展示主任职/代理切换；提供筛选、导出按钮。  
  - 在 `PositionTransferDialog` 与 `PositionSummaryCards` 显示代理提醒与 FTE 总览。  
- **QA**：  
  - Playwright 增补 Acting 场景：创建代理 → 自动到期 → 恢复。  
  - REST/GraphQL 跨租户测试脚本：验证 403 `JOB_CATALOG_TENANT_MISMATCH`、`POSITION_ASSIGNMENT_TENANT_MISMATCH`。  
  - 将脚本对接 CI（`make test-integration`）。

### Week 3 — 缓冲 & 收尾
- 验证自动恢复任务运行（附运行日志）。  
- 完成监控指标与告警文档（写入 `docs/development-tools/`）。  
- 更新 80 号方案 Stage 4 勾选、06 号日志 Stage 4 小节、实现清单、API 差异报告。  
- 将本计划归档至 `docs/archive/development-plans/86-position-assignment-stage4-plan.md`。

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
| 自动恢复误触发 | 高 | 中 | 双重条件校验（到期 + 当前状态），预演演练脚本；提供回滚操作。 |
| 任职 API 与 Fill/Vacate 冲突 | 中 | 中 | 先由 Rest API 包装现有服务，再决定是否逐步迁移；保留特性开关。 |
| 跨租户脚本复杂 | 中 | 中 | 先在 sandbox 演练，输出脚本与结果；纳入 CI 逐步运行。 |
| 代理扫描任务性能 | 中 | 低 | 分批处理 + 指标监控；必要时引入任务队列。 |
| 前端时间轴复杂度 | 中 | 低 | 复用现有组件，分阶段上线（beta feature flag）。 |

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

- [ ] 046 迁移 & 回滚脚本 + 演练日志  
- [ ] 任职专用 REST API 代码与测试  
- [ ] 代理自动恢复任务代码、日志、监控指标  
- [ ] GraphQL & 前端任职历史增强 + Playwright 场景  
- [ ] 跨租户回归测试脚本（REST/GraphQL）及 CI 集成  
- [ ] 文档同步（80 号 Stage 4 勾选、06 号日志、实现清单、API 差异报告、监控指南）  
- [ ] 计划归档（完成后移动至 `docs/archive/development-plans/`）

---

## 11. 变更记录

| 版本 | 日期 | 说明 | 作者 |
|------|------|------|------|
| v0.2 | 2025-10-17 | 根据 06 号评审意见修订，聚焦增量能力与差距分析 | 项目智能助手 |
| v0.1 | 2025-10-17 | 初始草案（已废弃） | 项目智能助手 |

