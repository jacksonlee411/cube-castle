# 204-HRMS 系统实施路线图

**文档编号**: 204  
**创建日期**: 2025-11-03  
**版本**: v2.0  
**状态**: Phase2 已完成并归档（Plan 215 2025-11-23 验收）；Phase3 workforce 模块准备启动  
**最后更新**: 2025-11-22  
**关联计划**: 203 总体划分、205 过渡方案、210-215 归档文档、Plan 252/254/255/259 门禁

---

## 1. 路线图概览

| 阶段 | 周期 | 目标 | 当前状态 |
|------|------|------|----------|
| Phase1：基线统一 (Plans 210-214) | Week1-2 | go.mod 合流、迁移基线、工具链统一 | ✅ 2025-11-06 归档 |
| Phase2：基础设施 (Plan 215) | Week3-4 | 事件总线、数据库层、日志系统、organization 重构、Docker 基座 | ✅ 2025-11-23 验收 |
| Phase3：workforce 模块 | Week5-8 | 员工主数据、sqlc 仓储、REST/GraphQL Facade、outbox 事件 | ⏳ 准备中 |
| Phase4：contract 模块 | Week9-12 | 合同模板、签署工作流、与 workforce 集成 | 🔜 规划中 |
| 后续：performance/compensation/payroll | 2026+ | HR 扩展域 | 📝 规划中 |

> Phase2 证据：`docs/development-plans/215-phase2-summary-overview.md`、`logs/plan221/**`、`logs/plan222/**`、`reports/architecture/architecture-validation.json`

---

## 2. 阶段目标与交付

### Phase1 — Baseline ✅
- Goose + Atlas 迁移、`module cube-castle`、命令/查询服务可并行编译。
- 输出：统一目录结构、共享 `internal/*` 组件、Go 1.24.9 基线。

### Phase2 — Infrastructure ✅
- `pkg/eventbus`: 并发安全 + Metrics 接口。
- `pkg/database`: 连接池、事务、事务性发件箱 + Prometheus 指标。
- `pkg/logger`: JSON 结构化日志、WithFields、noop/标准库桥。
- `internal/organization`: 模块骨架（api/service/repository/resolver 等）+ README。
- `docs/archive/development-plans/220-module-template-documentation.md`: 模块开发指南。
- `docker-compose.test.yml` + `scripts/run-integration-tests.sh` + `make test-db`：Docker 集成基座。
- 门禁：Plan 252/254/255/259 已纳入受保护分支 Required checks。

### Phase3 — Workforce（进行中）
- 扩充 OpenAPI / GraphQL 契约，定义 `internal/workforce/` 目录骨架。
- 首批交付：员工 CRUD、Assign/Unassign、事件（EmployeeCreated/Transferred/Terminated）。
- 技术动作：
  1. 引入 sqlc 生成器（`make sqlc-generate`）并纳入 CI。
  2. 复用 Plan 220 模板、Plan 221 Docker 基座，确保标准化骨架。
  3. 新增 Playwright 场景（员工入职）与 `make test-db` 覆盖。

### Phase4 — Contract（规划）
- 目标：合同模板管理、签署/续签/终止工作流、合同事件。
- 依赖：Phase3 验收、权限治理门禁、性能基线。

---

## 3. 关键里程碑

| 里程碑 | 时间 | 说明 |
|--------|------|------|
| Phase1 收官 | 2025-11-06 | Plans 210-214 归档，Goose/Atlas 测试通过 |
| Phase2 收官 | 2025-11-23 | Plan 215 验收、Plan 221/222 证据入库、门禁上线 |
| Phase3 MVP | 2025-12 （预计） | workforce 模块上线，单测覆盖 >80%，Playwright 场景收录 |
| Phase4 MVP | 2026-Q1 | contract 模块与 workforce 集成，Core HR 闭环 |
| Core HR 生产就绪 | 2026-Q1 | 组织/员工/合同完整流程可回归，门禁全绿 |

---

## 4. 资源与依赖

- **团队**：架构 1、后端 4、QA 2、DevOps 1、文档 0.5、合规 0.5。
- **工具链**：Go ≥1.24.9、Node ≥18、Docker Compose、Goose/Atlas、sqlc、Playwright 1.56。
- **必跑命令**：`make run-dev`、`make test-db`、`npm run test:e2e`、`node scripts/quality/architecture-validator.js`、`scripts/check-temporary-tags.sh`。
- **门禁**：Plan 252（权限契约）、Plan 254（端点/代理统一）、Plan 255（CQRS 架构）、Plan 259（REST 业务 GET=0）。

---

## 5. 风险与对策（精简版）

| 风险 | 影响 | 对策 |
|------|------|------|
| 模块模板被绕过导致结构漂移 | 高 | Plan 220 模板 + Plan 255 架构门禁，新增模块需提交模板对照表 |
| sqlc/Outbox 集成复杂度 | 高 | 先在 workforce 试点，保留回滚策略；禁止新增手写 init SQL |
| 契约/权限漂移 | 中 | Plan 252/259 门禁常态化，阈值 `PLAN259_BUSINESS_GET_THRESHOLD=0` |
| Docker 集成环境不稳 | 中 | 依赖 Plan 221 脚本，必要时 `plan253-coldstart` 预热镜像 |
| 多阶段并行带来的沟通成本 | 中 | 保持 `feat/shared-dev` 单分支开发，`docs/development-plans/215-phase2-execution-log.md` 持续记录 |

---

## 6. 下一步行动（2025-11-22）

1. 在 `internal/workforce/` 创建 README 与骨架，引用 Plan 220 模板条款。
2. 将 workforce 契约草稿提交至 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，并先执行 Plan 258/259 门禁。
3. 更新 Playwright 场景与 `make test-db`，确保 workforce 相关测试入口可在 CI 上运行。
4. 维护 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，登记 Phase3 的基础设施复用策略。

---

## 7. 文档治理

- 本文件聚焦“阶段路线图 + 状态”；详细执行日志见 `docs/development-plans/215-phase2-execution-log.md`；模块划分细节见 `docs/development-plans/203-hrms-module-division-plan.md`。
- 任何状态更新必须引用真实证据（日志、PR、Plan 文档）。行数保持 <200，避免重复引用。

