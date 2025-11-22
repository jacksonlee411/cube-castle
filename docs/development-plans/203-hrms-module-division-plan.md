# 203号方案：HRMS系统模块化演进与领域划分

**版本**: v3.0  
**创建日期**: 2025-11-03  
**作者**: 架构组  
**状态**: Phase2 已完成并归档（Plan 215 2025-11-23 验收），Phase3（workforce 模块）准备启动  
**最后更新**: 2025-11-22  
**关联文档**:
- `docs/archive/development-plans/204-HRMS-Implementation-Roadmap.md`（已归档路线图）
- `docs/development-plans/215-phase2-summary-overview.md`（Phase2 交付回顾）
- `docs/archive/development-plans/210-database-baseline-reset-plan.md` 等 Phase1 归档
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`

---

## 1. 概述

> 2025-11-22：原 204 路线图内容已并入本文件，204 文档迁移至 `docs/archive/development-plans/204-HRMS-Implementation-Roadmap.md` 仅作历史查询。

本文件定义 HRMS 模块化单体的领域划分、阶段性目标与里程碑。随着 Plan 210-215 已完成，本版聚焦：
- 巩固 Phase2 交付成果（事件总线、数据库共享层、标准化模块骨架、Docker 集成测试基座）。
- 提供下一阶段（Phase3 workforce / Phase4 contract）所需的清晰职责与依赖。
- 用 <300 行的精简形式作为架构对齐的唯一事实来源，避免与 204/215 重复。

---

## 2. 战略目标

1. **单一事实来源**：REST 由 `docs/api/openapi.yaml` 定义、GraphQL 由 `docs/api/schema.graphql` 定义，数据库变更全部由 `database/migrations/` 驱动。
2. **模块化单体 + PostgreSQL 原生 CQRS**：命令=REST、查询=GraphQL，命令服务/查询服务共享统一的基础设施（事件总线、事务性发件箱、结构化日志）。
3. **渐进式交付**：Phase1（基线）、Phase2（基础设施）、Phase3（workforce 模块）、Phase4（contract 模块），“每阶段均可独立验收并回滚”。
4. **可观测与治理并行**：Plan 252/254/255/259 的门禁与权限治理与功能交付同步推进，防止技术债反复。

---

## 3. 核心原则（摘自 AGENTS.md）

- Docker Compose 强制：PostgreSQL、Redis 等全部容器化，不允许本机服务占用 5432/6379。
- 分支策略：仅使用 `master` 与共享分支 `feat/shared-dev`，其他分支一律禁止。
- 迁移即真源：不得再依赖 `sql/init/*.sql`；所有 Schema 变更通过 Goose + Atlas。
- 权限与契约外部化：角色→scope 映射只能来自 OpenAPI/GraphQL 注释与生成物，禁止在代码硬编码。
- 可观测性基线：结构化日志 + Prometheus 指标 + Playwright 证据落盘到 `logs/plan*/`。

---

## 4. 领域划分与模块矩阵

| 模块域 | 主要职责 | 当前状态 | 关键依赖 |
|--------|----------|----------|----------|
| **平台 & 基础设施** (`pkg/eventbus`, `pkg/database`, `pkg/logger`, Docker 基座) | 提供事件驱动、事务性发件箱、统一日志、集成测试环境 | ✅ Phase2（Plan 215）完成，证据 `logs/plan221/**`、`logs/plan222/**` | Plans 210-215、Plan 252/255/259 门禁 |
| **Core HR / Organization** (`internal/organization`) | 组织、职位、调度、审计，提供 GraphQL/REST Facade | ✅ 重构完成，模板化结构落地 | 事件总线、数据库层、Plan 219 系列 |
| **Core HR / Workforce** (`internal/workforce` 预留) | 员工主数据、Assignments、生命周期事件 | ⏳ Phase3 目标（Week5-8），Plan 220 模块模板 & Plan 221 Docker 基座为前置 | Phase2 交付、GraphQL 契约扩展 |
| **Core HR / Contract** (`internal/contract` 预留) | 合同模板、签署流程、合同事件 | 🔜 Phase4 目标（Week9-12） | Workforce 模块、权限治理 |
| **Shared Capabilities** (`internal/auth`, `internal/graphql`, `internal/middleware`, `cmd/hrms-server/...`) | 鉴权、CQRS 中间件、缓存、命令/查询服务 | ✅ Phase1 完成并持续演进 | Plan 210-214、Plan 252/254/255 |
| **未来扩展**（Performance、Compensation、Payroll 等） | 按 `docs/reference/temporal-entity-experience-guide.md` 与 300/301 计划逐步实现 | 📝 规划中 | 203 当前文档 + 204 路线图 |

---

## 5. 分阶段路线图（与 215 对齐）

### Phase1（Week1-2）— 基线统一 ✅
- Plans 210-214 已归档：go.mod 统一、共享目录抽取、Go 1.24.9 基线与迁移脚本补齐。
- 输出：`module cube-castle`、命令/查询服务可同时编译、`internal/*` 结构成型。

### Phase2（Week3-4）— 基础设施构建 ✅ （详见 Plan 215）
- 交付：
  1. `pkg/eventbus` 并发安全实现 + 指标接口。
  2. `pkg/database`（连接池、事务、outbox 仓储）。
  3. `pkg/logger` 结构化日志 + WithFields + noop/标准库桥。
  4. `internal/organization` 重构 + 统一模块骨架。
  5. 模块模板文档 & Docker 测试基座 + Plan 222 验收报告。
- 门禁：Plan 252（权限契约）、Plan 254（端点统一）、Plan 255（CQRS 架构）均已接入 Required Checks。

### Phase3（Week5-8，进行中）— workforce 模块
- 目标：在 `internal/workforce/` 实现 Employee 主数据、sqlc 仓储、REST/GraphQL Facade、事件/outbox 集成。
- 关键动作：
  1. 扩展 OpenAPI/GraphQL 契约（由 204 计划驱动）。
  2. 依赖 Plan 220 模板与 Plan 221 Docker 基座输出，实现模块骨架复用。
  3. `make sqlc-generate` 纳入 CI（Plan 257 领域 Facade 覆盖率守卫）。
  4. Playwright 场景新增“员工入职”流程，证据落盘 `logs/plan24x/`。

### Phase4（Week9-12，规划中）— contract 模块
- 目标：合同模板 + 生命周期管理 + workforce 集成。
- 依赖：Phase3 全量验收、权限矩阵（Plan 259）硬门禁、性能基准。

---

## 6. 里程碑与成功标准

| 里程碑 | 目标日期 | 验收标准 |
|--------|----------|----------|
| Phase1 基线收官 | 2025-11-06 | Plans 210-214 全部归档并通过 Goose/Atlas 验证 |
| Phase2 基础设施收官 | 2025-11-23 | Plan 215 验收报告 + Plan 221/222 日志 + 门禁上线 |
| Phase3 workforce MVP | 2025-12-XX | `internal/workforce` 功能完备、单测覆盖 >80%、`make test-db` + Playwright 场景通过 |
| Phase4 contract MVP | 2026-Q1 | 合同全链路 E2E 通过、Core HR 域全部 GA |
| Core HR 生产就绪 | 2026-Q1 | 组织/员工/合同闭环、门禁全覆盖、性能基线达标 |

---

## 7. 资源与依赖

- **团队配置**：架构师 1、后端 4、QA 2、DevOps 1、文档 0.5、合规 0.5。Phase3 需至少 2 名熟悉 sqlc 与 gqlgen 的后端。
- **工具链**：Go ≥1.24.9、Node ≥18、Docker Compose、Goose + Atlas、sqlc、Playwright 1.56。
- **必跑命令**：`make run-dev`、`make test-db`、`npm run test:e2e`、`node scripts/quality/architecture-validator.js`、`scripts/check-temporary-tags.sh`。

---

## 8. 风险与对策（精简版）

| 风险 | 影响 | 对策 |
|------|------|------|
| 模块模板被绕过导致命名/结构漂移 | 高 | Plan 220 文档 + Plan 255 架构门禁，新增模块必须通过模板清单审查 |
| sqlc/Outbox 融合复杂度超预期 | 高 | 在 workforce 模块先行试点，保留手工 SQL fallback，但禁止新建 `sql/init` |
| 权限/契约漂移 | 中 | Plan 252/259 门禁常态化，仓库变量 `PLAN259_BUSINESS_GET_THRESHOLD=0` 已硬门禁 |
| Docker 集成环境不稳定 | 中 | Plan 221 脚本 + 253 镜像门禁，“docker compose test” 走专用网络及卷 |
| 多阶段并行导致沟通成本高 | 中 | 保持 `feat/shared-dev` 单分支开发 + `docs/development-plans/215-phase2-execution-log.md` 每日更新 |

---

## 9. 下一步行动（2025-11-22 周期）

1. 在 `docs/archive/development-plans/204-HRMS-Implementation-Roadmap.md` 中补录 Phase2 完成情况，并将 Phase3 里程碑同步到 203 文档（本次已完成）。
2. 建立 `internal/workforce/README.md` 与基本骨架，引用 Plan 220 模板，确保结构与 organization 模块一致。
3. 将 workforce 契约草稿提交至 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，并先跑 Plan 258/259 门禁确认无漂移。
4. 将 Playwright 场景 `tests/e2e/workforce-onboarding.spec.ts` 立项，脚本输出到 `logs/plan24x/`。

---

## 10. 文档治理

- 本文件为 203 计划唯一事实来源。若需更详细的执行日志，参考：
  - `docs/development-plans/215-phase2-execution-log.md`（日更日志、证据索引）
  - `docs/archive/development-plans/204-HRMS-Implementation-Roadmap.md`（历史跨阶段甘特，仅做参考）
- 编辑要求：保持 <300 行，所有新增行动项需附责任角色、依赖与目标日期；若信息已迁移到其他文档，应在此处留下指针而非重复内容。
