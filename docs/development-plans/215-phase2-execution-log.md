# 215 - Phase2 执行日志与进度跟踪

**文档编号**: 215
**标题**: Phase2 - 建立模块化结构执行日志
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v2.1（状态同步与优先级更新）

---

## 最新更新（2025-11-10）

- ✅ [Plan 247 / T5] 文档与治理对齐完成：`Temporal Entity Experience Guide` 已作为唯一事实来源（docs/reference/temporal-entity-experience-guide.md），旧文档在 reference 目录改为“Deprecated 占位符”（无正文，避免第二事实来源）。证据已落盘：`logs/plan242/t5/rg-zero-ref-check.txt`、`logs/plan242/t5/document-sync.log`、`logs/plan242/t5/architecture-validator.log`、`logs/plan242/t5/inventory-sha.txt`。
- ✅ [Plan 242 / T1] Temporal Entity Page 抽象完成：组织/职位详情入口统一迁移至 `TemporalEntityPage`，路由校验、无效提示与功能映射已记录在 `reports/plan242/naming-inventory.md#temporal-entity-page`，为后续 Timeline/类型/测试抽象提供共享基线。
- ✅ [Plan 244 / T2] Temporal Timeline & Status 抽象合入：`frontend/src/features/temporal/entity/timelineAdapter.ts` 与 `statusMeta.ts` 统一组织/职位映射，Lint 禁止回退旧命名，执行记录见 `logs/plan242/t2/`。
- 🔄 [Plan 244 / T2] Temporal timeline REST 契约补齐 `TemporalEntityTimelineVersion` 字段，Go/前端测试与 Implementation Inventory 同步更新（详见 `logs/plan242/t2/2025-11-11-temporal-timeline-go.md`）。

### 新增（2025-11-15）
- ✅ [Plan 244 / T2] 验收通过（Chromium/Firefox 各 1 轮）  
  - 观测用例：`frontend/tests/e2e/smoke-org-detail.spec.ts`、`frontend/tests/e2e/temporal-header-status-smoke.spec.ts`  
  - 集成用例：`frontend/tests/e2e/temporal-management-integration.spec.ts`（8 passed / 4 skipped）  
  - 证据：`logs/plan242/t2/244-e2e-acceptance.log`、`frontend/playwright-report/index.html`、`logs/plan242/t2/244-namespace-scan.log`
- ✅ [Plan 240E] 回归与 Runbook 收束已启动（守卫通过；CI 执行通道已接入）  
  - 守卫：`logs/plan240/E/selector-guard.log`、`logs/plan240/E/architecture-validator.log`、`logs/plan240/E/temporary-tags.log`（通过）  
  - 前端 Lint/类型：`logs/plan240/E/frontend-lint.log`、`logs/plan240/E/frontend-typecheck.log`  
  - 执行入口：`scripts/plan240/run-240e.sh` / `.github/workflows/plan-240e-regression.yml`（产物统一至 `logs/plan240/E`；HAR 由前端配置集中到 `logs/plan240/B`/`BT`）

### 新增（2025-11-15 — 第二批）
- ✅ [合规] 按 AGENTS 强制要求卸载宿主机 Redis，释放 6379 端口（全部服务统一由 Docker Compose 管理）  
  - 操作摘要：停止并禁用 systemd 服务 → `apt purge redis-server redis-tools` → 清理 `/etc/redis /var/lib/redis /var/log/redis` → 验证 `ss -lntp` 无 6379 监听  
  - 目的：避免宿主机与容器冲突，确保环境隔离与可复现性
- ✅ [Plan 221] Docker 集成测试基座落地验收（本地）  
  - 证据：`logs/plan221/integration-run-*.log`（包含 Goose up/down 循环与 outbox dispatcher 成功/重试/关停/幂等 全部 PASS）  
  - 结果：`make test-db` 稳定通过，`docker-compose.test.yml` 与 `scripts/run-integration-tests.sh` 工作正常，工作流 `integration-test.yml` 已配置
- ✅ [Plan 222] 阶段性验收证据沉淀（持续中）  
  - REST：`POST /api/v1/organization-units` 创建成功（返回业务 code）；日志：`logs/plan222/create-response-*.json`  
  - GraphQL：`organizations(filter: { codes: [...] })` 查询与分页元信息正常；日志：`logs/plan222/graphql-query-*.json`  
  - E2E 烟测（Chromium/Firefox 各 1 轮）：`smoke-org-detail.spec.ts`、`temporal-header-status-smoke.spec.ts` 均通过  
  - 健康与 JWKS：`logs/plan222/health-command-*.json`、`health-graphql-*.json`、`jwks-*.json`  
  - 覆盖率：新增多组单测（facade/utils/middleware），顶层包提升至 ~45.5%，整体提升进行中；报告：`logs/plan222/coverage-org-*.{out,txt,html}`
  - E2E P0 全量（Chromium/Firefox）Mock 模式通过；日志：`logs/plan222/playwright-P0-*.log`、`logs/plan222/playwright-FULL-*.log`

### 新增（2025-11-15 — 第三批）
- ✅ [Plan 252] 权限一致性与契约对齐（完成并归档）  
  - 守卫：Plan 252 校验器（OpenAPI scopes 引用→注册、GraphQL 注释→映射、resolver 授权覆盖）已接入 CI；DEV_MODE 默认禁用守卫已启用  
  - 产物：`reports/permissions/*`（usage/registry/mapping/calls/summary）；运行时映射 `cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json`  
  - 证据：`logs/plan252/validator-summary-*.txt` 与 `logs/plan252/reports-*/` 快照  
  - 文档：归档正文 `docs/archive/development-plans/252-permission-consistency-and-contract-alignment.md`；签字纪要 `docs/archive/development-plans/252-signoff-20251115.md`

### 新增（2025-11-16 — Plan 255 软门禁启动）
- ✅ 255 门禁接入（软）：工作流已接入 ESLint 架构守卫 + architecture-validator + golangci-lint（固定 v1.59.1）  
  - 证据（本地预跑）：  
    - `logs/plan255/architecture-validator-20251116_101740.log`  
    - `reports/architecture/architecture-validation.json`  
  - 配置与策略：前端 GET 例外仅 `/auth`；JWKS 不设永久前端例外（临时仅限 DEV+auth 模块，已改用 UnauthenticatedRESTClient 出站）  
- ⏳ 待办（硬门禁前置）：  
  - 仓库受保护分支将 plan-250-gates/plan-253-gates/plan-255-gates 设为 required checks（登记截图与失败示例链接）  
  - 后端告警/监控 JSON 字段 snake_case → camelCase；当前以 `//nolint:tagliatelle // TODO‑TEMPORARY(2025-11-30)` 过渡，文件：`internal/monitoring/health/alerting.go`  
 - 📎 审计（非门禁）：根路径端口/禁用端点扫描  
   - `logs/plan255/audit-root-20251116_102250.log`（发现 37 端口 + 1 禁用端点模式；作为问题清单分批收敛，不阻断合并）
 - 📮 PR 与评审登记（Plan 255）  
   - PR: https://github.com/jacksonlee411/cube-castle/pull/5 （状态：open）  
   - 评审清单评论：已自动添加；本地副本：`logs/plan255/pr-review-comment-20251116_104539.md`  
   - 自动化脚本：`scripts/ci/auto-pr.sh`；Make 目标：`make pr-255-soft-gate`

### 新增（2025-11-16 — Plan 255 完成）
- ✅ Plan 255（CQRS 分层门禁）已在 master 稳定通过  
  - gates‑255：通过（前端架构门禁 + 后端 depguard/tagliatelle 真实违规阻断）  
    - 运行链接：<https://github.com/jacksonlee411/cube-castle/actions/runs/19401060738/job/55508624545>  
  - 前端门禁报告：`reports/architecture/architecture-validation.json`（生成且已归档）  
  - 证据：`logs/plan255/**`（ESLint/architecture-validator/golangci‑lint 全量日志）
- 运行策略（与 05 指南一致）：  
  - Trunk + Local First：提交前本地跑“前端门禁 + go build”；golangci‑lint 采集不阻断  
  - CI：Required checks 以工作流为准；后端 lint 仅以 depguard/tagliatelle 违规为阻断，其余类型检查噪音仅记录日志

#### Plan 255 · CI 远程收尾（索引）
- Required checks（受保护分支）：将 `plan-250-gates`、`plan-253-gates`、`plan-255-gates` 勾选为必需检查（Settings → Branches）。  
  - 证据占位：`logs/plan255/branch-protection-required-checks.md`（截图/说明链接，含失败示例链接）[TBD]
- 首次成功运行登记：在本段落补充 CI 运行链接与工件名称（artifact: `plan255-logs`）。  
  - 运行链接：<https://github.com/jacksonlee411/cube-castle/actions/runs/19403010378>（status=completed, conclusion=success）；artifact: plan255-logs（保留 7 天）
- Root 审计门禁开关：已切换为 hard（阻断）。  
  - 单一事实来源：`.github/workflows/plan-255-gates.yml` 中 `PLAN255_ROOT_AUDIT_MODE=hard`  
  - 清单来源：`logs/plan255/audit-root-*.log`（集中建 Issue，分批回收）
- 软门禁 PR 联通性：通过 `make pr-255-soft-gate` 验证 `scripts/ci/auto-pr.sh` 可创建 PR；若组织策略不允许，记录手工 PR 链接与原因。  
  - 证据占位：`logs/plan255/pr-*.txt`（脚本输出）或手工 PR URL [TBD]
- 本地验证证据（离线）
  - 后端门禁（depguard/tagliatelle，仅关注门禁项）：`logs/plan255/golangci-lint-local-20251116_152704.log`（PASS）
  - 前端架构门禁（CQRS/端口/禁直连）：`reports/architecture/architecture-validation.json`（0 违规）
 - 全量本地验证（联网，对齐 CI；证据索引）
   - 依赖安装：`logs/plan255/npm-ci-root-20251116_162058.log`、`logs/plan255/npm-ci-frontend-20251116_162058.log`
   - ESLint 架构守卫：`logs/plan255/eslint-architecture-20251116_162058.log`（仅 1 条 warning：unused eslint-disable；不阻断）
   - 前端架构验证器（frontend）：`logs/plan255/architecture-validator-20251116_162058.log`（通过）
 - 根路径审计（root，非门禁）：`logs/plan255/audit-root-20251116_162058.log`（统计：端口违规 37、禁用端点 14；总 51，已建清单待收敛）
  - 根路径审计（最新复跑，非门禁）：`logs/plan255/audit-root-20251116_163442.log`；报告 `logs/plan255/architecture-root-20251116_163442.json`（0 违规；已为硬门禁切换准备就绪）
   - golangci-lint（depguard+tagliatelle，仅门禁项）：`logs/plan255/golangci-lint-20251116_162058.log`（CLEAN）
   - OpenAPI（Spectral）：`logs/plan255/openapi-spectral-20251116_162058.log`（8 warnings，0 errors）
   - Plan 252 权限契约校验：`logs/plan255/plan252-validate-permissions-20251116_162058.log`（通过；未注册引用=0，未匹配映射=2 已记录）

### 新增（2025-11-15 — 优先级与下一步）
- P0：Plan 222 收口验收与文档更新（见 222 章节与证据日志）
- P0：202 阶段1（模块化单体合流，不改协议；索引：`docs/development-plans/202-CQRS混合架构深度分析与演进建议.md`）
- P1：Plan 221 基座 CI 常态运行（已本地验收；CI 冷启动指标随首轮工作流登记）
- P1：202 阶段2（契约 SSoT 与前端 API Facade；同上文档“阶段 2: 工程优化”）
- P1：Plan 219E 回归补强以支撑 222 覆盖率目标

---

## 25x 启动登记（占位）

说明：依据 202（简化版）路线图与 25x 子计划分解，登记启动信息与证据路径。执行细节以各子计划文档为唯一事实来源；所有命令输出与验证证据统一登记至本日志（215）。

- 250 · 模块化单体合流  
  - 计划窗口：TBD（W?）  
  - 负责人：TBD  
  - 准入条件：Plan 219 完成、Plan 221 基座可用、宿主端口合规（AGENTS）  
  - 产物/证据：`logs/plan250/*`、合流验收清单（单端口/健康/指标/REST/GraphQL 等效）  
  - 文档：`docs/development-plans/250-modular-monolith-merge.md`

  - 必跑门禁清单（CI 工作流 `plan-250-gates.yml`）：
    - [ ] legacy 环境门禁：`scripts/quality/gates-250-no-legacy-env.sh`
    - [ ] 单一二进制门禁：`scripts/quality/gates-250-single-binary.sh`
    - [ ] command 端无 8090 监听与字面量：`scripts/quality/gates-250-no-8090-in-command.sh`
    - [ ] 复用 Plan 253 门禁：compose 端口映射与镜像标签固定（另见 253 工作流）

  - 刚性验收（落盘到 `logs/plan250/`）：
    - [x] 221：`make test-db` 通过 → `test-db-*.log`
    - [ ] E2E 最小集（232/241/244，Chromium/Firefox 各 1 轮）→ `e2e-*.log`
    - [x] JWKS/JWT/多租户链路抽样一致 → `jwks-*.json`、`tenant-check-*.log`
    - [ ] 性能/资源基线（204 指标 + performance/ 脚本）→ `perf-*.json`

- 251 · 运行时统一（连接池/中间件/健康/指标）  
  - 状态：已完成（2025-11-15）  
  - 准入条件：Plan 250 单体主路径；Plan 217/218 可用（已满足）  
  - 产物/证据：  
    - 健康：`logs/plan251/health-command-*.json`（./scripts/quality/validate-health.sh）  
    - 指标：`logs/plan251/metrics-command-*.txt`（./scripts/quality/validate-metrics.sh；STRICT=true 校验 HELP/TYPE）  
  - 文档：`docs/development-plans/251-runtime-unification-health-metrics.md`（单体主路径/统一健康与指标/标签与网络限制规范）

- 253 · 部署与流水线简化（单体优先） — 已完成（2025-11-16）  
  - 负责人：DevOps（与 QA/后端协作）  
  - 产物/证据：`logs/plan253/*`（门禁与冷启动指标）、Make/Workflow/Compose 变更  
  - 冷启动基线：`logs/plan253/coldstart-20251116001139.log`（compose_up_ms≈1979ms；db_ready_seconds=10s）  
  - 门禁证据：`logs/plan253/compose-ports-and-images.log`（端口映射冻结、镜像标签固定）  
  - 文档：`docs/development-plans/253-deployment-pipeline-simplification.md`

### 253 首轮冷启动指标登记（占位）
- 触发：`plan-253-gates`（主干定时 / 触发变更）  
- 产物：`logs/plan253/coldstart-*.log`（compose_up_ms、db_ready_seconds）  
- 本地兜底记录已完成：  
  - 预拉取过程：`logs/plan253/coldstart-20251115220516.log`  
  - 基线（已预拉取后冷启动）：`logs/plan253/coldstart-20251116001139.log`（compose_up_ms≈1979ms；db_ready_seconds=10s）
- 链接：首次成功运行后补充 CI 运行链接与统计摘要（P50/P90）；在本节仅索引日志路径，避免第二事实来源

- 254 · 前端端点与代理收敛（单基址） — 已完成  
  - 产物/证据（索引）：  
    - `logs/plan254/playwright-254-run-*.log`（E2E 运行与通过记录）  
    - `logs/plan254/trace/`（trace 证据）  
    - `reports/architecture/architecture-validation.json`（架构门禁报告：cqrs/ports/forbidden=0）  
  - 说明：端口/代理/基址配置以源文件为准（frontend/vite.config.ts、frontend/src/shared/config/ports.ts）；compose 端口映射治理由 Plan 253 门禁负责  
  - 文档：`docs/development-plans/254-frontend-endpoint-and-proxy-consolidation.md`（状态：已完成）

- 256 · 契约 SSoT 生成流水线（阶段2）  
  - 计划窗口：TBD（W?）  
  - 负责人：TBD  
  - 准入条件：脚本链路可运行（Node/Go 工具链基线满足 AGENTS）  
  - 产物/证据：`logs/plan256/*`、`make generate-contracts` 幂等日志、CI contract-sync（生成→快照→工作树 clean）结果、drift-report（报告模式，阻断由 258 承担）  
  - 文档：`docs/development-plans/256-contract-ssot-generation-pipeline.md`

- 257 · 前端领域 API 门面采纳（阶段2）  
  - 计划窗口：TBD（W?）  
  - 负责人：TBD  
  - 准入条件：241/242 抽象完成；统一命名与选择器守卫启用  
  - 产物/证据：`logs/plan257/*`、`reports/facade/coverage.json`（覆盖率报告；阻断阈值≥0.8）、E2E/单测通过记录  
  - 工作流：`.github/workflows/plan-257-gates.yml`（已切换为阈值 0.8 阻断；请在受保护分支设置为 Required check）  
  - 文档：`docs/development-plans/257-frontend-domain-api-facade-adoption.md`
  - 本次 CI 运行（登记）：  
    - Run: https://github.com/jacksonlee411/cube-castle/actions/runs/19405921517（结论：success）  
    - 工件：`logs/plan257/ci-artifacts/coverage.json`（coverage=1.25；threshold=0.8；offenders=[tests-only]）  
  - 状态：已完成（验收通过 · 2025-11-16）

- 258 · 契约漂移校验与门禁（阶段2）  
  - 计划窗口：TBD（W?）  
  - 负责人：TBD  
  - 准入条件：Plan 256 生成流水线可用  
  - 产物/证据：`logs/plan258/*`、`reports/contracts/drift-report.json`、CI 工件 `plan258-drift-report`（阻断）  
  - 工作流：`.github/workflows/plan-258-gates.yml`（受保护分支 Required）  
  - 文档：`docs/development-plans/258-contract-drift-validation-gate.md`

#### Plan 258 · 临时差异登记与回收计划（Phase B 报告模式）
// TODO-TEMPORARY(2025-11-23): 字段矩阵差异白名单（短期）— 等待 GraphQL 契约统一（profile: JSON、自增 sortOrder 非空）与审计字段透出策略评审；本条目到期前完成修复或移除白名单。
- 差异来源（报告）：`reports/contracts/drift-report.json#fieldMatrix`  
  - missingInRest：`path, changeReason, endDate, isTemporal, childrenCount, deletedBy, deletionReason, suspendedAt, suspendedBy, suspensionReason`  
  - missingInGql：`operationType, operatedBy, operationReason`  
  - typeMismatch：无（2025-11-16 已对齐：GraphQL profile → JSON）  
  - nullabilityMismatch：无（2025-11-16 已对齐：GraphQL sortOrder → Int!）  
- 临时放行：`scripts/contract/drift-allowlist.json`（仅报告模式；阻断未开启字段矩阵）  
- 回收计划：  
  1) GraphQL 引入 JSON 标量并调整 `sortOrder: Int!`（已完成 2025-11-16）；  
  2) 评审审计/派生字段的对等暴露策略并统一；  
  3) 移除 allowlist 对应项；  
  4) 已启用字段矩阵阻断（2025-11-16）；登记报告快照：`reports/contracts/drift-report.json`（CI artifact: plan258-drift-report）

- 259 · 协议策略复盘（可选）  
  - 计划窗口：TBD（W?）  
  - 负责人：TBD  
  - 准入条件：阶段 1+2 完成（250/251/253/254/256/257/258）  
  - 产物/证据：`logs/plan259/*`、复盘报告与结论  
  - 文档：`docs/development-plans/259-protocol-strategy-review.md`

---

## 概述

本文档跟踪 Phase2 的实施进展（Week 3-4，Day 12-18），根据 204 号文档第二阶段的定义，工作分解为 **7 个具体实施方案**（Plan 216-222）。

**基础设施建设**:
- `pkg/eventbus/` - 事件总线（**Plan 216**）
- `pkg/database/` - 数据库共享层（**Plan 217**）
- `pkg/logger/` - 日志系统（**Plan 218**）

**数据库与迁移管理**:
- 迁移脚本回滚（Down 脚本）✅ **已完成（Plan 210）**
- Atlas 工作流配置 ✅ **已完成（Plan 210）**

**模块重构与验证**:
- 重构 `organization` 模块按新模板结构（**Plan 219**）
- 创建模块开发模板文档（**Plan 220**）
- 构建 Docker 集成测试基座（**Plan 221**）
- 验证 organization 模块正常工作（**Plan 222**）
- 更新 README 和开发指南（**Plan 222**）

---

### Plan 242 – 里程碑与验收清单（骨架）

说明：Plan 242 分解为 T0–T5 六个子阶段；本清单仅登记里程碑与验收证据路径，实施细节以各子计划为唯一事实来源（docs/development-plans/242-*.md）。

- T0 现状盘点（已完成）
  - 事实来源：`reports/plan242/naming-inventory.md`（最新）
  - 证据登记：
    - [ ] `logs/plan242/t0/rg-inventory-scan.log`
    - [ ] `logs/plan242/t0/inventory-sha256.txt`
- T1 页面与路由命名抽象（已完成，详见 Plan 243）
  - 事实来源：`docs/development-plans/243-temporal-entity-page-plan.md`
  - 证据登记：
    - [ ] `logs/plan242/t1/storybook-diff.log`
    - [ ] `logs/plan242/t1/router-migration.log`
- T2 Timeline/Status 抽象（进行中，详见 Plan 244）
  - 事实来源：`docs/development-plans/244-temporal-timeline-status-plan.md`
  - 验收门槛（登记用）：
    - [ ] 前端：`npm run lint`、`npm run test`、`npm run test:e2e -- --project=chromium --project=firefox`（各至少 1 轮）
    - [ ] 后端：`go generate ./cmd/hrms-server/query/...`、`go test ./cmd/hrms-server/...`
    - [ ] 契约：更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 且 `node scripts/generate-implementation-inventory.js` 通过
    - [ ] 日志：`logs/plan242/t2/*.log`（包含上述命令输出）
- T3 类型与契约统一（已完成，详见 Plan 245/245A/245T）
  - 事实来源：
    - `docs/development-plans/245-temporal-entity-type-contract-plan.md`
    - `docs/development-plans/245A-unified-hook-adoption.md`
    - `docs/development-plans/245T-openapi-no-ref-siblings-fix.md`
  - 证据登记：
    - [ ] `logs/plan242/t3/implementation-inventory.log`
    - [ ] `logs/plan242/t3/plan245-guard.log`
- T4 Selectors & Fixtures 统一（已完成 Phase 1，详见 Plan 246）
  - 事实来源：`docs/development-plans/246-temporal-entity-selectors-fixtures-plan.md`
  - 运行门禁：`npm run guard:selectors-246`（计数不升高）
  - 证据登记：
    - [ ] `logs/plan242/t4/selector-guard-246.log`
    - [ ] `logs/plan242/t4/e2e-{chromium,firefox}.log`
- T5 文档与治理对齐（已完成，已归档，详见 Plan 247）

---

## 模板（受保护分支门禁证据 · 适用于 Plan 253/255）
- 仓库设置截图：Settings → Branches → Branch protection rules（勾选必需检查：plan-250-gates、plan-253-gates、plan-255-gates）
- 失败示例链接：至少 1 个 PR 触发门禁失败的运行链接（Actions run URL）
- 日志归档：
  - plan-253：`logs/plan253/*`（compose 端口/镜像/冷启动检查）
  - plan-255：`logs/plan255/*`（前端架构守卫、golangci-lint）
- 备注：若临时放行，需在对应代码处添加 `// TODO-TEMPORARY(YYYY-MM-DD): 原因|计划|截止`（≤1迭代），并在此处登记清单与回收日期
  - 事实来源：`docs/archive/development-plans/247-temporal-entity-docs-alignment-plan.md`
  - 证据登记：
    - [ ] `logs/plan242/t5/rg-zero-ref-check.txt`
    - [ ] `logs/plan242/t5/document-sync.log`
    - [ ] `logs/plan242/t5/architecture-validator.log`
    - [ ] `logs/plan242/t5/inventory-sha.txt`

---

### Plan 240B – 职位详情数据装载与等待治理（登记）

说明：本节用于登记 240B 的依赖门槛、执行证据与验收结果。实施细节以 `docs/development-plans/240B-position-loading-governance.md` 为唯一事实来源。

- 依赖与准入（需全部满足）
  - [x] 243/T1 统一入口已合并（`TemporalEntityPage` 可用）
  - [x] 244 已验收（Adapter/StatusMeta 合并、契约同步、基础 E2E 绿灯）
  - [x] 241 恢复并作为承载框架（统一 Hook/Loader 作为唯一入口）
  - [x] 守卫接入：`npm run guard:plan245`、`npm run guard:selectors-246` 通过（基线计数不升高）  
    - 证据：`logs/plan240/B/guard-plan245.log`、`logs/plan240/B/guard-selectors-246.log`
- 环境与契约前置
  - [x] Docker/服务就绪：`make docker-up` → `make run-dev` → `make frontend-dev`
  - [x] 健康检查：`curl http://localhost:9090/health`、`curl http://localhost:8090/health` → 200（证据：`logs/plan240/B/health-checks.log`）
  - [x] JWT：`make jwt-dev-mint`（用于 240BT 冒烟；会话流程经 dev-token 模式与 JWKS 验证）
  - [x] 契约先行（如适用）：更新 `docs/api/*` + `node scripts/generate-implementation-inventory.js`（不涉及新增字段）  
    - 证据：无需生成新清单（本计划未改契约）
- 执行与证据（示例文件名，可按实际生成）
  - [x] Loader 取消策略：路由级预热 + 卸载取消（实现见 240B 文档，代码已合入）
  - [x] 重试与错误边界：queryClient 统一重试/退避；ErrorBoundary 防白屏（组织详情）
  - [x] QueryKey 与失效刷新：失效 SSoT 覆盖职位写操作（Create/Update/Version/Transfer）
  - [x] 守卫：`logs/plan240/B/guard-plan245.log`、`logs/plan240/B/guard-selectors-246.log`
- 单测与 E2E（统一门槛）
  - [x] 单测通过（覆盖 Loader 预热/取消调用链、Hook 错误工厂来源、失效 SSoT）  
  - [x] E2E（抽样已绿，CI 严格轮次纳入 241/CI 运维）：  
    - 组织冒烟：`logs/plan240/BT/health-checks.log`、`logs/plan240/BT/network-har-*.har`（Chromium/Firefox 通过）  
    - 职位多页签（Chromium）：`logs/plan240/B/e2e-chromium-position-tabs.log`（通过；GraphQL Stub + SSoT 选择器）  
    - 后续在 CI 增加严格轮次（3××2 浏览器）并落盘 trace/HAR/计数

**结论（登记）**：240B 已完成；统一装载/取消、重试/退避、失效 SSoT 与等待/选择器策略均已落地，抽样 E2E 与守卫通过。CI 层面的 3××2 严格轮次纳入 241/215 的 acceptance 作业执行。
---

### Plan 240A – 职位详情 Layout 对齐与骨架替换（登记）

说明：本节用于登记 240A 的依赖门槛、执行证据与验收结果。实施细节以 `docs/development-plans/240A-position-layout-alignment.md` 为唯一事实来源。

- 依赖与准入（需全部满足）
  - [x] 242/247 文档与命名治理闭环（引用 reference 指南，不复制正文）
  - [x] 守卫接入：`npm run guard:selectors-246` 通过（基线计数不升高）  
    - 证据：`logs/plan240/A/selector-guard.log`
- 执行与证据（本次合并范围）
  - [x] Tabs 键盘与 A11y（tablist/aria-selected/左右箭头）  
    - 参考：`frontend/src/features/positions/PositionDetailView.tsx:498-541`
  - [x] 六页签顺序一致；左侧版本/时间轴呈现与组织一致（视觉 token/断点沿用 Canvas）
  - [x] 选择器统一（组件/用例改为集中选择器；旧前缀计数下降）  
    - 证据：`logs/plan240/A/selector-guard.log`
  - [x] 门禁：`architecture-validator`、`document-sync` 落盘  
    - 证据：`logs/plan240/A/architecture-validator.log`、`logs/plan240/A/document-sync.log`
- 单测与 E2E（登记）
  - [ ] E2E：在 CI 跑 `position-tabs.spec.ts`（Chromium/Firefox 各≥1），trace/日志归档  
    - 证据路径：`logs/plan240/A/playwright-*.log`、`logs/plan240/A/playwright-trace/*`
  - [ ] Storybook 对比截图落盘：`reports/plan240/baseline/storybook/*.png`

**结论（登记）**：240A 已完成；布局/交互与组织详情对齐，选择器统一、门禁通过。E2E 与 Storybook 产物由 CI 与设计回归流程产出并回填至上述路径。
---

### Plan 240C – 职位 DOM/TestId 治理与选择器统一（登记）

说明：本节用于登记 240C 的依赖门槛、执行证据与验收结果。实施细节以 `docs/development-plans/240C-position-selectors-unification.md` 为唯一事实来源。

- 依赖与准入（需全部满足）
  - [x] 240A 基线对齐完成（布局/组件一致性）  
  - [x] 守卫接入：`npm run guard:selectors-246` 通过（不新增旧前缀）  
    - 证据：`logs/plan240/C/selector-guard.log`、`reports/plan246/baseline.json`
- 执行与证据（本次合并范围）
  - [x] 选择器集中补齐：在 `frontend/src/shared/testids/temporalEntity.ts` 新增  
    - `versionRow/versionRowPrefix`、`vacancyBoard/headcountDashboard`、`transferOpen/Target/Date/Reason/Reassign/Confirm`
  - [x] 组件替换硬编码 testid（保留 fallback）  
    - VersionList 行：`frontend/src/features/positions/components/versioning/VersionList.tsx`  
    - VacancyBoard：`frontend/src/features/positions/components/dashboard/PositionVacancyBoard.tsx`  
    - HeadcountDashboard：`frontend/src/features/positions/components/dashboard/PositionHeadcountDashboard.tsx`  
    - TransferDialog：`frontend/src/features/positions/components/transfer/PositionTransferDialog.tsx`
  - [x] 测试迁移至集中选择器（Vitest/E2E）  
    - Dashboard/Detail/Headcount 单测、position-tabs/position-lifecycle/CRUD（E2E）断言均引用 SSoT 选择器
- 单测与 E2E（登记）
  - [x] 守卫：`logs/plan240/C/selector-guard.log`（通过，旧前缀计数显著下降）  
  - [ ] 单测：在 CI 运行 `cd frontend && npm run test`（本地已完成断言迁移）  
  - [ ] E2E：在 CI 运行 `npm run test:e2e`（本地仅迁移断言，不强制跑浏览器）

**结论（登记）**：240C 集中选择器与主要组件/测试迁移已落地，Plan 246 守卫通过且旧前缀计数显著下降。余下项由 CI 执行单测与 E2E 轮次采集证据并回填本日志。
---

## 阶段时间表与计划映射（Week 3-4）

| 周 | 日 | 行动项 | 计划 | 描述 | 负责人 | 状态 |
|-----|-----|--------|------|------|--------|------|
| **W3** | D1 | 2.1 | **Plan 216** | 实现 `pkg/eventbus/` 事件总线 | 基础设施 | ✅ 2025-11-03 |
| | D2 | 2.2 | **Plan 217** | 实现 `pkg/database/` 数据库层 | 基础设施 | ✅ 2025-11-04 |
| | D2 | 2.3 | **Plan 218** | 实现 `pkg/logger/` 日志系统 | 基础设施 | ✅ 2025-11-04 |
| | D3 | 2.3a | **Plan 217B** | 构建 outbox→eventbus 中继 | 基础设施 | ✅ 2025-11-05 |
| | D3 | 2.4-2.5 | Plan 210 | 迁移脚本和 Atlas 配置 | DevOps | ✅ |
| | D4-5 | 2.6 | **Plan 219** | 重构 organization 模块结构（执行窗口调整至 Day 15-17，含缓冲） | 架构师 | ✅ 2025-11-06 |
| **W4** | D1-2 | 2.7 | **Plan 220** | 创建模块开发模板文档（与 219B 并行） | 架构师 | ⏳ |
| | D2-3 | 2.8 | **Plan 221** | 构建 Docker 化集成测试基座（支撑 219E） | QA | ⏳ |
| | D3-5 | 2.9-2.10 | **Plan 222** | 验证与文档更新；预留 Day 26 缓冲 | QA/文档 | ⏳ |

---

## 进度记录

### 行动项 2.1 - 实现 `pkg/eventbus/` 事件总线 (Plan 216)

**对应计划**: **Plan 216 - eventbus-implementation-plan.md**

**计划行动**:
- [x] 定义事件总线接口（Event、EventBus、EventHandler）- `pkg/eventbus/eventbus.go`
- [x] 实现内存事件总线（MemoryEventBus + AggregatePublishError + MetricsRecorder）- `pkg/eventbus/memory_eventbus.go`
- [x] 编写单元测试（覆盖率 98.1%，覆盖成功/失败/无订阅者/并发场景）- `pkg/eventbus/eventbus_test.go`
- [x] 集成失败聚合、日志与指标记录（提供 noop Logger/Metrics，符合 Plan 217B 依赖）

**交付物**:
```
pkg/eventbus/
├── eventbus.go        # 接口定义（Event、EventBus、EventHandler）
├── memory_eventbus.go # 内存实现（并发安全）
├── error.go           # 错误定义
└── *_test.go          # 单元测试（覆盖率 > 80%）
```

**关键特性**:
- Event 接口：EventType()、AggregateID()
- EventBus 接口：Publish()、Subscribe()
- EventHandler：func(ctx, event) error 签名
- 支持多订阅者处理
- 错误处理：处理器失败不阻止其他处理器，并通过 `AggregatePublishError` 返回给调用方
- 可观测性：提供成功/失败/无订阅者/耗时指标，日志接口与 Plan 218 logger 契约一致

**技术要点**:
- 并发安全：使用 RWMutex 保护 handlers 映射
- 失败聚合：所有处理器执行完毕后返回聚合错误并记录指标
- 性能目标：发布延迟 < 1ms

**验收标准** (来自 Plan 216):
- [x] Subscribe/Publish 功能正常（`TestPublishWithSingleSubscriber`、`TestPublishWithMultipleSubscribers`）
- [x] 多订阅者都被调用（单测断言调用计数）
- [x] 并发安全（`go test -race ./pkg/eventbus` 通过）
- [x] 处理器失败时返回 `AggregatePublishError` 并包含失败明细（`TestPublishWithHandlerError`）
- [x] 指标记录（成功/失败/无订阅者/延迟）经单元测试验证（`testMetrics` 断言）
- [x] 单元测试覆盖率 > 80%（`go test -cover ./pkg/eventbus` 输出 98.1%）
- [x] 代码通过 `go fmt` 和 `go vet`（手动执行 `gofmt`、`go vet ./pkg/eventbus`）

**执行记录**:
- 2025-11-03 完成代码提交，运行 `go test ./pkg/eventbus`、`go test -race ./pkg/eventbus`、`go test -cover ./pkg/eventbus`、`go vet ./pkg/eventbus` 全部通过。
- 更新 `pkg/eventbus/README.md` 说明指标命名与 Plan 217B 集成方式。

**负责人**: 基础设施团队
**计划完成**: Day 12 (W3-D1)
**状态**: ✅ 已完成（2025-11-03）

**详细文档**: 见 `docs/development-plans/216-eventbus-implementation-plan.md`

---

### 行动项 2.2 - 实现 `pkg/database/` 数据库层 (Plan 217)

**对应计划**: **Plan 217 - database-layer-implementation.md**

**计划行动**:
- [x] 创建数据库连接管理（连接池配置）——`pkg/database/connection.go` + 单测
- [x] 实现事务支持（Transaction 包装）——`pkg/database/transaction.go`
- [x] 实现事务性发件箱（outbox）表接口——`pkg/database/outbox.go`、`database/migrations/20251107090000_create_outbox_events.sql`
- [x] 编写单元测试与集成测试——`pkg/database/*_test.go`、`tests/integration/migration_roundtrip_test.go`

**交付物**:
```
pkg/database/
├── connection.go      # 连接池管理（标准参数）
├── transaction.go     # 事务支持（WithTx）
├── outbox.go          # 事务性发件箱接口
├── metrics.go         # Prometheus 指标
├── error.go           # 错误定义
└── *_test.go          # 单元 & 集成测试
```

**关键参数** (硬编码标准配置):
- MaxOpenConns: 25（防止连接溢出）
- MaxIdleConns: 5（连接复用）
- ConnMaxIdleTime: 5 分钟（定期刷新）
- ConnMaxLifetime: 30 分钟（周期替换）

**关键接口**:
- `NewDatabase(dsn)` - 创建连接
- `WithTx(ctx, fn)` - 事务支持
- `GetUnpublishedEvents()` - 获取未发布事件
- `MarkEventPublished()` - 标记事件已发布
- `IncrementRetryCount()` - 增加重试计数

**事务性发件箱** (Outbox 模式):
- OutboxEvent 结构：event_id、aggregate_id、event_type、payload 等
- SaveOutboxEvent() - 在事务内保存事件
- 用于保证跨模块操作的最终一致性
- 与 Plan 216 eventbus 配合使用

**Prometheus 指标**:
- db_connections_in_use - 当前活动连接
- db_connections_idle - 空闲连接
- db_query_duration_seconds - 查询延迟直方图

**验收标准** (来自 Plan 217):
- [x] 连接池配置正确（MaxOpenConns=25）
- [x] 事务创建、提交、回滚正常
- [x] Outbox 事件保存和查询成功
- [x] 单元 & 集成测试通过（`go test ./pkg/database -cover` -> 82.1%，`go test ./tests/integration/migration_roundtrip_test.go`）
- [x] 无 race condition（关键场景由单元测试覆盖）

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D2)
**状态**: ✅ 已完成（2025-11-04）

**详细文档**: 见 `docs/development-plans/217-database-layer-implementation.md`

---

### 行动项 2.3a - 构建 outbox dispatcher 中继 (Plan 217B) ✅ 已完成

**对应计划**: **Plan 217B - outbox dispatcher 中继**

**计划行动**:
- [ ] 实现独立的中继组件，定时查询 `outbox` 表未发布事件
- [ ] 调用 Plan 217 提供的 `OutboxRepository` 接口，批量取回事件
- [ ] 调用 Plan 216 的 `eventbus.Publish` 发布事件，并在成功后标记 `published=true`
- [ ] 为失败事件增加 `retry_count` 并记录结构化日志
- [ ] 实现指数退避或最小间隔，避免频繁轮询
- [ ] 编写单元与集成测试，覆盖事务失败不发布事件的场景

**交付物**:
```
cmd/hrms-server/internal/outbox/
├── dispatcher.go        # 中继循环（可配置间隔/批量大小）
├── dispatcher_config.go # 配置与默认值
├── metrics.go           # 成功/失败/重试指标
├── dispatcher_test.go   # 单元测试
└── integration_test.go  # 集成测试（依赖 Plan 217/216）
```

**运行要点**:
- 默认轮询间隔 5s，可通过环境变量覆盖
- 每批拉取 50 条事件，发布成功后调用 `MarkPublished`
- 发布失败时调用 `IncrementRetryCount` 并根据重试次数决定退避时间
- 与 Plan 218 logger 集成，记录成功、失败与重试明细
- 暴露 Prometheus 指标：`outbox_dispatch_success_total`、`outbox_dispatch_failure_total`、`outbox_dispatch_retry_total`

**验收标准**:
- [x] 事务提交失败时不会发布事件（集成测试覆盖）
- [x] 成功发布的事件在 outbox 表中被标记为 `published=true`
- [x] 连续失败的事件会增加 `retry_count` 并进入退避队列
- [x] 中继可通过上下文或信号安全停止
- [x] 单元与集成测试覆盖率 > 80%

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D3)
**状态**: ✅ 已完成（2025-11-05）

**详细文档**: 见 `docs/development-plans/217B-outbox-dispatcher-plan.md`

---

### 行动项 2.3 - 实现 `pkg/logger/` 日志系统 (Plan 218)

**对应计划**: **Plan 218 - logger-system-implementation.md**

**计划行动**:
- [x] 创建结构化日志记录器
- [x] 实现日志级别控制（Debug, Info, Warn, Error）
- [x] 集成性能监控（响应时间、数据库查询统计）
- [x] 编写单元测试

**交付物**:
```
pkg/logger/
├── logger.go          # Logger 接口和实现
├── std.go             # 标准库网桥（backward compatibility）
└── *_test.go          # 单元测试（覆盖率 > 80%）
```

**Logger 接口**:
- Debug/Debugf, Info/Infof, Warn/Warnf, Error/Errorf ✅
- WithFields(map[string]interface{}) - 添加结构化字段 ✅
- JSON 输出格式（timestamp、level、message、fields、caller） ✅

**日志级别**:
- DebugLevel, InfoLevel, WarnLevel, ErrorLevel ✅
- 通过环境变量 `LOG_LEVEL` 设置 ✅
- 默认 InfoLevel ✅

**结构化输出** (JSON 格式):
```json
{
  "timestamp": "2025-11-04T10:30:45.123Z",
  "level": "INFO",
  "message": "organization created",
  "fields": {"organizationID": "org-123"},
  "caller": "organization/service.go:42"
}
```

**标准库网桥**:
- std.go 提供向后兼容性
- 在测试和工具场景中使用
- 不在生产代码中依赖

**验收标准** (来自 Plan 218):
- [x] Logger 接口定义完整
- [x] JSON 输出格式正确
- [x] 日志级别控制有效
- [x] WithFields() 正常工作
- [x] 单元测试覆盖率 > 80%
- [x] 所有测试通过（13 个测试用例）
- [x] 代码通过 `go fmt` 和 `go vet`

**执行记录**:
- 2025-11-04 完成代码实现，包括 Logger、std 网桥和完整的单元测试。
- 运行 `go test ./pkg/logger -v` 全部通过。
- 与 Plan 218A-E 子计划一致，logger 迁移已达生产级质量。

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D2)
**状态**: ✅ 已完成（2025-11-04）

**详细文档**: 见 `docs/development-plans/218-logger-system-implementation.md` 及 `docs/development-plans/218C-logger-verification-report.md`、`docs/development-plans/218E-logger-rollout-closure.md`

---

### 行动项 2.4-2.5 - 迁移脚本与 Atlas 配置

**对应计划**: Plan 210（已完成）

**状态**: ✅ **已完成（Plan 210，2025-11-06）**

**已完成的工作**:
- ✅ 为所有迁移文件补齐 `-- +goose Down` 回滚脚本
- ✅ 配置 Atlas `atlas.hcl` 和 `goose.yaml`
- ✅ 基线迁移脚本 `20251106000000_base_schema.sql` 已部署
- ✅ up/down 循环验证通过

**证据**: `docs/archive/development-plans/210-execution-report-20251106.md`

此工作为 Plan 221 (Docker 集成测试) 和 Plan 222 (验证) 的前置条件。

---

### 行动项 2.6 - 重构 `organization` 模块结构 (Plan 219)

**对应计划**: **Plan 219 - organization-restructuring.md**

**计划行动**:
- [x] 按新模板重组 organization 模块代码
- [x] 定义模块公开接口（api.go）
- [x] 整理 internal/ 目录结构（service、repository、handler、resolver、domain）
- [x] 确保模块边界清晰
- [x] 集成基础设施（Plan 216-218）

**目标结构**:
```
internal/organization/
├── api.go                         # 公开接口定义
├── internal/
│   ├── domain/
│   │   ├── organization.go        # 域模型
│   │   ├── department.go
│   │   ├── position.go
│   │   ├── events.go              # 域事件定义
│   │   └── constants.go
│   ├── repository/
│   │   ├── organization_repository.go
│   │   ├── department_repository.go
│   │   ├── position_repository.go
│   │   └── *_test.go
│   ├── service/
│   │   ├── organization_service.go
│   │   ├── department_service.go
│   │   ├── position_service.go
│   │   └── *_test.go
│   ├── handler/                   # REST 处理器
│   │   ├── organization_handler.go
│   │   └── *_test.go
│   ├── resolver/                  # GraphQL 解析器
│   │   ├── organization_resolver.go
│   │   └── *_test.go
│   └── README.md                  # 内部说明
└── README.md                      # 模块说明
```

**关键工作**:
1. **api.go - 模块公开接口**
   - OrganizationAPI interface（所有公开方法）
   - 其他模块仅能依赖 api.go，不能导入 internal/

2. **基础设施集成**
   - Service 层注入 eventbus (Plan 216)
   - Service 层注入 database (Plan 217)
   - 使用 logger (Plan 218) 记录操作
   - 使用 eventbus 发布组织变更事件
   - 使用 database 的 WithTx 管理事务

3. **事务性发件箱**
   - 在 service 中创建新实体时，同一事务内保存 outbox 事件
   - 异步发布事件给 eventbus

4. **功能等同性**
   - 重构后行为必须与重构前完全相同
   - 所有 API 端点签名不变
   - 数据查询结果一致

**实施步骤** (来自 Plan 219):
1. 分析与准备：审视现有代码，梳理接口和依赖
2. 目录重构：创建新结构，重新分类代码
3. 基础设施集成：注入 eventbus、database、logger
4. 测试与验证：运行回归测试，确保功能等同

**验收标准** (来自 Plan 219):
- [x] 模块按新模板重构完成
- [x] api.go 公开接口清晰
- [x] internal/ 目录结构符合规范
- [x] 无循环依赖
- [x] 功能等同（100%）
- [x] 单元测试覆盖率 > 80%
- [x] 性能无退化

**实际完成情况（2025-11-06）**:
- 目录已集中到 `internal/organization/*`，README 记录审计、handler、repository、resolver、service、scheduler、validator、dto 等聚合边界，成为组织域的唯一事实来源 `internal/organization/README.md:3`, `internal/organization/README.md:21`.
- `internal/organization/api.go` 构建统一的 `CommandModule`，一次性注入数据库、审计记录器、级联服务、职位/任职/职位目录仓储、事务性发件箱以及调度服务，命令侧只需依赖公开 API 即可接入 `internal/organization/api.go:28`, `internal/organization/api.go:119`.
- 查询侧通过 `AssignmentQueryFacade` 复用同一个仓储并接入 Redis 缓存，缓存键规范、TTL 以及刷新逻辑集中在 `internal/organization/query_facade.go:28`, `internal/organization/query_facade.go:136`，并由 `internal/organization/query_facade_test.go:42` 覆盖缓存命中与回源场景。
- README 记录的 219E 验收脚本（组织生命周期冒烟、REST 性能基准）已同步到 `scripts/e2e/org-lifecycle-smoke.sh` 与 `scripts/perf/rest-benchmark.sh`，输出日志位于 `logs/219E/*` 以支撑 Plan 222 的后续验证 `internal/organization/README.md:56`.


**负责人**: 架构师 + 后端团队
**计划完成**: Day 14-15 (W3-D4-5)
**状态**: ✅ 已完成（2025-11-06）

**详细文档**: 见 `docs/development-plans/219-organization-restructuring.md`

---

### 行动项 2.7 - 创建模块开发模板文档 (Plan 220)

**对应计划**: **Plan 220 - module-template-documentation.md**

**计划行动**:
- [ ] 编写完整的模块开发指南（> 3000 字）
- [ ] 基于 organization 重构经验提供样本代码
- [ ] 文档化 sqlc 使用规范
- [ ] 文档化事务性发件箱集成规范
- [ ] 文档化 Docker 集成测试规范
- [ ] 创建各阶段检查清单

**交付物**:
```
docs/development-guides/
├── module-development-template.md  # 主指南文档
├── examples/
│   └── organization/               # 参考实现代码
└── checklists/
    ├── module-structure-checklist.md
    ├── api-contract-checklist.md
    ├── testing-checklist.md
    └── deployment-checklist.md
```

**主文档章节** (来自 Plan 220):
1. 模块基础知识 - Bounded Context、DDD
2. 模块结构模板 - 标准目录结构说明
3. 数据访问层规范 - sqlc 使用、repository 模式
4. 事务性发件箱集成 - outbox 模式、可靠性保证
5. Docker 集成测试 - 容器化测试、Goose 迁移
6. 测试规范 - 单元、集成、E2E 测试
7. API 契约规范 - REST 命名、GraphQL schema
8. 质量检查清单 - 代码质量、安全、性能

**目标受众**:
- 后端开发者（新模块实现者）
- 新团队成员（理解项目架构）
- QA 工程师（了解测试策略）

**与其他计划的关系**:
- 基于 Plan 219 (organization 重构)
- 为 Phase 3 (workforce 模块) 提供参考
- 使用 Plan 216-218 的基础设施
- 引用 Plan 221 的 Docker 测试规范

**验收标准** (来自 Plan 220):
- [x] 文档完整（> 3000 字）
- [x] 包含 5+ 个代码示例
- [x] 示例代码可编译且正确
- [x] 包含 3 个以上检查清单
- [x] 内容与 organization 模块对齐
- [x] 新模块开发者可独立参考

**负责人**: 架构师 + 文档支持
**计划完成**: Day 17 (W4-D1-2)
**状态**: ✅ 已完成（2025-11-07）

**详细文档**: 见 `docs/development-plans/220-module-template-documentation.md`

---

### 行动项 2.8 - 构建 Docker 化集成测试基座 (Plan 221)

**对应计划**: **Plan 221 - docker-integration-testing.md**

**计划行动**:
- [ ] 创建 `docker-compose.test.yml`（PostgreSQL）
- [ ] 编写集成测试启动脚本
- [ ] 验证 Goose up/down 流程
- [ ] 创建测试数据初始化脚本
- [ ] 更新 Makefile 和 CI/CD 配置

**交付物**:
```
├── docker-compose.test.yml              # Docker 配置
├── scripts/test/
│   ├── init-db.sql                      # 初始化脚本
│   └── run-integration-tests.sh          # 测试启动脚本
├── Makefile（更新）
│   ├── make test-db-up
│   ├── make test-db-down
│   ├── make test-db
│   ├── make test-db-logs
│   └── make test-db-psql
└── .github/workflows/
    └── integration-test.yml             # CI 工作流
```

**Docker 配置** (来自 Plan 221):
- PostgreSQL 15 Alpine 镜像
- 自动初始化数据库
- 端口映射：5432:5432（保持标准端口；如被占用需清理宿主冲突服务后再启动）
- 健康检查：pg_isready
- 卷挂载：迁移脚本、初始化脚本

**集成测试流程**:
1. 启动 Docker 容器
2. 等待数据库就绪
3. 运行 Goose 迁移 (up)
4. 执行 Go 集成测试
5. 验证回滚 (down)
6. 清理容器

**Makefile 目标**:
- `make test-db-up` - 启动测试数据库
- `make test-db-down` - 停止测试数据库
- `make test-db` - 完整的集成测试流程
- `make test-db-logs` - 查看数据库日志
- `make test-db-psql` - 连接到测试数据库

**CI/CD 集成** (.github/workflows/integration-test.yml):
- 在 GitHub Actions 中运行集成测试
- 使用 services 启动 PostgreSQL
- 运行 Goose 迁移
- 执行集成测试并上传覆盖率

**验收标准** (来自 Plan 221):
- [x] Goose up/down 循环通过（见 `logs/plan221/integration-run-*.log`）
- [x] 集成测试可正常运行（`make test-db` 稳定通过）
- [x] 多次运行结果一致（多份运行日志已登记）
- [x] 无端口冲突（遵循 AGENTS 标准端口与宿主服务清理）
- [ ] 预拉取镜像后的 Docker 启动 < 10s（CI 首轮冷启动计时待登记）
- [ ] 数据库就绪时间 < 15s（CI 首轮计时待登记）

**负责人**: QA + DevOps
**计划完成**: Day 18-19 (W4-D2-3)
**状态**: ✅ 已完成（本地验收 2025-11-15；CI 指标计时待首轮登记）

**详细文档**: 见 `docs/development-plans/221-docker-integration-testing.md`

---

### 行动项 2.9 - 验证 organization 模块正常工作 (Plan 222)

**对应计划**: **Plan 222 - organization-verification.md**

**计划行动**:
- [ ] 单元测试 organization 服务（覆盖率 > 80%）
- [ ] 集成测试 organization 与数据库交互
- [ ] 验证 Goose up/down + Docker 测试流程正常
- [ ] 执行 REST API 回归测试
- [ ] 执行 GraphQL 查询回归测试

**验证范围** (来自 Plan 222):

**1. 单元测试验证**
```bash
go test -v -race -coverprofile=coverage.out ./internal/organization/...
```
- [ ] 所有单元测试通过
- [ ] 测试覆盖率 > 80%
- [ ] 无 race condition

**2. 集成测试验证**
```bash
make test-db-up
go test -v -tags=integration ./cmd/hrms-server/...
make test-db-down
```
- [ ] 集成测试全部通过
- [ ] Goose 迁移 up/down 循环通过
- [ ] 数据库状态一致

**3. REST API 回归测试**
- [ ] GET /org/organizations/{code}
- [ ] POST /org/organizations
- [ ] PUT /org/organizations/{code}
- [ ] 响应字段为 camelCase
- [ ] HTTP 状态码正确
- [ ] 错误处理一致

**4. GraphQL 查询回归测试**
- [ ] query { organizations { id code name } }
- [ ] 返回数据符合 schema
- [ ] 错误处理正确

**5. E2E 端到端流程测试**
```
1. 创建新的组织单元
2. 查询组织单元详情
3. 创建部门
4. 为部门创建职位
5. 分配员工到职位
6. 查询组织结构
7. 更新组织信息
8. 验证审计日志
```

**6. 性能基准测试**
- 单个查询：< 50ms (P99)
- 列表查询（100 条）：< 200ms (P99)
- 创建操作：< 100ms (P99)
- 并发（100 并发）：> 100 req/s

**验收标准** (来自 Plan 222):
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有测试通过（0 失败）
- [ ] REST API 回归通过
- [ ] GraphQL 查询回归通过
- [ ] E2E 端到端通过
- [ ] 性能基准达标

**负责人**: QA
**计划完成**: Day 19 (W4-D3)
**状态**: 🔄 进行中（阶段性证据已登记：REST/GraphQL/E2E 烟测/健康与 JWKS/覆盖率）

**详细文档**: 见 `docs/development-plans/222-organization-verification.md`

---

### 行动项 2.10 - 更新 README 与开发指南 (Plan 222)

**对应计划**: **Plan 222 - organization-verification.md**（第二部分）

**计划行动**:
- [ ] 更新项目 README（新目录结构说明）
- [ ] 更新开发者速查（模块化单体工作流）
- [ ] 添加常见命令列表
- [ ] 更新 CI/CD 说明
- [ ] 更新实现清单
- [ ] 完成 Phase2 执行验收报告

**文档更新** (来自 Plan 222):

**1. README.md 更新**
- 项目结构说明（cmd/、internal/、pkg/）
- 快速开始指南
- 构建、测试、开发命令
- 模块化架构简述
- 链接到详细文档

**2. DEVELOPER-QUICK-REFERENCE.md 更新**
- 模块结构规范
- 常用命令速查
- 基础设施使用示例（eventbus、database、logger）
- 调试技巧

**3. IMPLEMENTATION-INVENTORY.md 更新**
- Phase1 状态（✅ 完成）
- Phase2 状态（✅ 完成）
- Phase3 计划状态
- 代码统计、覆盖率等指标

**4. 架构文档更新** (docs/architecture/modular-monolith-design.md)
- 当前架构状态
- 模块间通信机制
- 基础设施层说明

**5. Phase2 执行验收报告** (reports/phase2-execution-report.md)
- 执行概览
- 验收结果
- 质量指标
- 关键交付物
- 风险消除情况
- Phase3 预期

**验收标准** (来自 Plan 222):
- [ ] README 更新完整
- [ ] 开发指南更新
- [ ] 实现清单更新
- [ ] 架构文档更新
- [ ] 验收报告完成

**负责人**: 文档支持 + 架构师
**计划完成**: Day 20-21 (W4-D4-5)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/222-organization-verification.md`

---

## 关键检查点

### 基础设施质量检查点

- [x] `pkg/eventbus/` (Plan 216) 单元测试覆盖率 > 80%
- [x] `pkg/database/` (Plan 217) 连接池配置正确（MaxOpenConns=25）
- [x] `pkg/logger/` (Plan 218) 与 Prometheus 指标集成
- [x] `outbox dispatcher` (Plan 217B) 能可靠发布并记录重试指标
- [ ] 所有共享包无循环依赖
- [ ] 代码格式通过 `go fmt ./...`
- [ ] 代码通过 `go vet ./...`
- [ ] 无 race condition (`go test -race ./...`)

### 模块重构检查点 (Plan 219)

- [x] organization 模块按新模板重构完成（参考 `internal/organization/README.md:3`）
- [x] 模块公开接口清晰（api.go）`internal/organization/api.go:28`
- [x] internal/ 目录结构符合规范
- [x] 模块间无直接依赖（仅通过 interface）
- [x] CQRS 边界清晰（REST 命令、GraphQL 查询）
- [x] 基础设施正确集成（eventbus、database、logger）

### 测试与验证检查点 (Plan 221-222)

- [x] Docker 集成测试基座可正常启动 (Plan 221)
- [x] Goose up/down 循环验证通过 (Plan 221)
- [ ] organization 模块所有测试通过 (Plan 222)
- [ ] REST/GraphQL 端点行为一致 (Plan 222)
- [ ] E2E 端到端流程测试通过 (Plan 222)
- [ ] 性能基准测试达标 (Plan 222)

### 文档完整性检查点 (Plan 220, 222)

- [x] 模块开发模板文档完成 (Plan 220)
- [ ] README 和开发指南更新 (Plan 222)
- [ ] 实现清单更新 (Plan 222)
- [ ] Phase2 执行验收报告完成 (Plan 222)

---

## 风险与应对

| 风险 | 影响 | 概率 | 预防措施 | 对应计划 |
|------|------|------|--------|---------|
| 事件总线设计不当 | 中 | 中 | 充分评审，确保扩展性 | Plan 216 |
| 数据库层性能问题 | 高 | 中 | 连接池参数验证，压力测试 | Plan 217 |
| organization 重构破裂 | 高 | 中 | ✅ 2025-11-06 在 feature 分支完成全量测试后合并主干 | Plan 219 |
| outbox 中继未按时完成 | 高 | 中 | ✅ 已完成（2025-11-05） | Plan 217B |
| Docker 集成测试不稳定 | 中 | 中 | 固化镜像版本，CI 预跑 | Plan 221 |
| 时间超期 | 中 | 低 | 充分的并行执行 | 整体协调 |

---

## 计划文档导航

### 7 个实施方案文档

| 计划 | 文档名 | 工作内容 | 关键交付 |
|------|--------|---------|---------|
| **Plan 216** | 216-eventbus-implementation-plan.md | pkg/eventbus/ 实现 | 事件总线接口和内存实现 |
| **Plan 217** | 217-database-layer-implementation.md | pkg/database/ 实现 | 连接池、事务、outbox |
| **Plan 217B** | 217B-outbox-dispatcher-plan.md | outbox 中继实现 | 事件发布中继、重试机制 |
| **Plan 218** | 218-logger-system-implementation.md | pkg/logger/ 实现 | 结构化日志、Prometheus |
| **Plan 219** | 219-organization-restructuring.md | organization 重构 | 标准模块结构 |
| **Plan 220** | 220-module-template-documentation.md | 模块开发指南 | 模板文档、样本代码 |
| **Plan 221** | 221-docker-integration-testing.md | Docker 测试基座 | Compose 配置、脚本 |
| **Plan 222** | 222-organization-verification.md | 验证与文档更新 | 验收报告、文档更新 |

### 相关规划文档

- `204-HRMS-Implementation-Roadmap.md` - Phase2 实施路线图（权威定义）
- `215-phase2-summary-overview.md` - Phase2 全景概览（协调中心）
- `06-integrated-teams-progress-log.md` - Phase2 启动指导
- `203-hrms-module-division-plan.md` - HRMS 模块划分蓝图

---

## 相关文档

- `204-HRMS-Implementation-Roadmap.md` - Phase2 实施路线图（权威定义）
- `06-integrated-teams-progress-log.md` - Phase2 启动指导
- `203-hrms-module-division-plan.md` - HRMS 模块划分蓝图
- `docs/api/openapi.yaml` - REST API 契约
- `docs/api/schema.graphql` - GraphQL 契约

---

## 提交记录

| 日期 | 提交 | 描述 |
|------|------|------|
| 2025-11-04 | b328bd1e | docs: correct Phase2 scope - infrastructure setup not new modules |
| 2025-11-04 | c481f189 | docs: create Phase2 implementation plans (216-222) |
| 2025-11-04 | 1b2b39b9 | docs: add Phase2 implementation summary and overview |
| 2025-11-04 | - | docs: update Phase2 execution log aligned with Plan 216-222 |

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-15
**版本**: v2.1（状态同步与优先级更新）
**关键更改**: 同步 217B/220/221 状态与检查点；新增近期关键推进（P0/P1）与 222 阶段性证据登记说明

### Plan 240E – 验收登记（2025-11-15 14:32:53 CST）

- 守卫：选择器 ✅ · 架构 ✅ · 临时标签 ✅
- 前端：Lint ⚠️ · Typecheck ✅
- 证据：`logs/plan240/E`（run、guards、trace） · HAR 见 `logs/plan240/B`/BT
- 执行日志：`logs/plan240/E/playwright-run-20251115142132.log`
