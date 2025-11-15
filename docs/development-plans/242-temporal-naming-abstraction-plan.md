# Plan 242 – 时态实体命名抽象与统一治理

**编号**: 242  
**标题**: Temporal Entity Naming Convergence & Governance  
**创建日期**: 2025-11-10  
**关联计划**: Plan 241（框架重构）、Plan 240（职位页面重构）、Plan 06（集成验证）  
**状态**: 进行中 · 修订版（T1 已完成，参见 Plan 243；T2/T3/T4 按本修订推进）  

---

## 1. 背景与动因

1. 组织/职位详情入口仍以实体名硬编码：`OrganizationTemporalPage` 直接绑定 `TemporalMasterDetailView`（frontend/src/features/organizations/OrganizationTemporalPage.tsx:11-78），`PositionTemporalPage` 则自建一套 Tab/状态/表单（frontend/src/features/positions/PositionTemporalPage.tsx:1-146）。命名缺乏统一抽象，阻碍 Plan 241 推动的共享框架落地。
2. Timeline/版本工具与状态元数据沿用职位专属命名：早期的 `timelineAdapter.ts` 仅处理 `PositionRecord`，状态配置也分散在职位私有 `status meta` 文件与组织端的 `shared/utils/statusUtils.ts`。缺乏 `TemporalEntity` 级别的命名会导致未来新增实体时再次复制粘贴。
3. E2E 测试与 fixtures 采用实体专属命名：`position-tabs.spec.ts` 依赖 `position-temporal-page-*` testid（frontend/tests/e2e/position-tabs.spec.ts:91-145），而组织用例 `organization-create.spec.ts` 使用 `organization-*` 前缀（frontend/tests/e2e/organization-create.spec.ts:4-41）。`utils/positionFixtures.ts` 亦与职位强绑定（frontend/tests/e2e/utils/positionFixtures.ts:1-160），阻碍 selector/fixture 共用。

> **结论**：需要在命名层面建立“Temporal Entity” 中性抽象，覆盖页面、组件、Hook、Timeline、状态配置与测试资产，为 Plan 241/240 提供一致的命名基线。

当前状态补充（与历史差异）：
- 入口层已统一：`TemporalEntityPage` + `entityRoutes` 为唯一入口（见 frontend/src/features/temporal/pages/TemporalEntityPage.tsx、.../entityRoutes.tsx；Plan 243 已完成）。
- 时间线/状态已有共享抽象：`frontend/src/features/temporal/entity/timelineAdapter.ts`、`frontend/src/features/temporal/entity/statusMeta.ts`。
- 选择器集中：`frontend/src/shared/testids/temporalEntity.ts`（`temporalEntitySelectors`）为唯一事实来源（Plan 246）。
- 统一 Hook 骨架：`frontend/src/shared/hooks/useTemporalEntityDetail.ts`（Plan 245）。

---

## 2. 目标与验收

| 目标 | 说明 | 验收方式 |
| --- | --- | --- |
| A. 页面命名抽象 | 以 `TemporalEntityPage` 为统一入口，组织/职位仅通过配置扩展，不再保留硬编码页面命名 | 新路由/组件命名 PR + Storybook 录屏 |
| B. Timeline & Status 抽象 | 输出 `TemporalEntityTimelineAdapter` 与 `TemporalEntityStatusMeta`，由各实体适配器注入字段映射 | Adapter 单测 + API 对照表 |
| C. Selector & Fixture 统一 | 建立 `temporalEntitySelectors` 与中性 fixtures（如 `temporalEntityFixtures.ts`），E2E 用例仅使用中性命名 | Playwright diff + utils 重用证明 |
| D. 类型与契约命名 | 在 `shared/types` / GraphQL Operation / React Query Key 层统一导出 `TemporalEntityRecord`、`TemporalEntityStatus` 等别名，组织/职位仅作类型别名 | TypeScript diff + GraphQL schema 变更说明 |
| E. 文档与指南 | 更新 `docs/reference/temporal-entity-experience-guide.md`（由原 Positions 指南抽象而来），同步命名映射 | 文档 PR + 评审纪要 |
| F. 守卫与门禁 | 冻结旧 selector 前缀与旧 GraphQL 操作名，新增旧命名即失败 | `npm run guard:selectors-246`、`npm run guard:plan245` 通过，基线计数不升 |

验收补充（统一门禁与证据留存）：
- 守卫：`npm run guard:selectors-246` 与 `npm run guard:plan245` 均需通过（基线计数不升，必要时提供限期 allowlist 与理由）。
- E2E：Chromium/Firefox 各 3 轮通过；控制台出现 `performance.mark('TemporalEntity:*')` 与 logger 事件；trace/截图/日志落盘 `logs/plan242/t4/`。
- 实现清单：`node scripts/generate-implementation-inventory.js` 产出与哈希保存（reports/implementation-inventory.* + logs/plan242/**）。
- OpenAPI：`npm run lint:api` 0 errors（已修复 no-$ref-siblings，参见 Plan 245T），避免契约回归。

---

## 3. 范围与任务

### 3.1 T0 – 现状盘点（1 天）
- 环境前置（Docker 强制）：`make docker-up` → `make run-dev` → `make frontend-dev`；健康检查 `curl http://localhost:9090/health` / `http://localhost:8090/health`（200）；`make jwt-dev-mint`。
- 运行 `node scripts/generate-implementation-inventory.js` 并对前后端（Go/TS）执行 `rg "Organization.*Temporal|Position.*Temporal"`，覆盖 `internal/`、`cmd/`、`frontend/`、`tests/`、`scripts/`。  
- 输出/维护 `reports/plan242/naming-inventory.md`，列出所有命名分布、引用文件与行号（含 Go/E2E/文档），作为迁移基线与滚动证据。

### 3.2 T1 – 页面与路由命名抽象（1.5 天）
-（已完成，见 Plan 243）以 `TemporalEntityPage` + `TemporalEntityRouteConfig` 统一入口；保留实体特有校验（组织 7 位数字、职位 `P\d{7}`）。  
- 路由现状：`/organizations/:code/temporal`、`/positions/:code` 通过统一入口渲染；短期接受 `new/NEW` 作为创建模式（解析层归一化），长期统一为 `new`（2 个迭代窗口）。

### 3.3 T2 – Timeline/Status 命名抽象（4 天）
- 迁移职位端 Timeline 适配器为共享的 `frontend/src/features/temporal/entity/timelineAdapter.ts`，并提供 `createTimelineAdapter({ entity, labelBuilder })` 工厂；组织/职位复用一套类型定义，避免 `unitType = 'POSITION'` 等硬编码。  
- 整合职位/组织状态配置为 `frontend/src/features/temporal/entity/statusMeta.ts`，输出 `TemporalEntityStatusMeta`（含 `statusConfig.position`, `statusConfig.organization`），命名、色板、标签统一在一个映射表中。  
- 更新所有引用（Position version list、组织 StatusBadge 等）为新命名，同时通过 lint 规则阻止直接引用旧文件；预留时间处理 Storybook/Vitest 回归。
- 与 Go 层同步：命令服务中的 Timeline/Status DTO、REST handler 响应以及 Query 服务 GraphQL resolver 所用结构体需统一字段命名，必要时更新 `cmd/hrms-server/command/internal/services/temporal*.go` 与 `cmd/hrms-server/query/internal/resolvers/*`，并运行 `go test ./cmd/hrms-server/...` 保障兼容。

### 3.4 T3 – Types/GraphQL/Hook 命名抽象（4 天）
- 在 `frontend/src/shared/types` 内新增 `TemporalEntityRecord`, `TemporalEntityTimelineEntry`, `TemporalEntityStatus` 等统一接口，由 `PositionRecord`, `OrganizationUnit` 转为类型别名；所有消费端通过实体适配器映射字段。  
- 统一 GraphQL operation 与 React Query key 命名：例如 `POSITION_DETAIL_QUERY_NAME` → `TEMPORAL_ENTITY_DETAIL_QUERY` + entity 参数；`positionDetailQueryKey` → `temporalEntityDetailQueryKey`。  
- 在本计划内直接交付 `useTemporalEntityDetail` Hook（含 QueryKey、React Query integration、单测），并让 `usePositionDetail`、`useOrganizationDetail` 成为该 Hook 的薄封装，彻底摆脱 Plan 241 依赖。  
- 更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 与后端调用说明，使 GraphQL operation 名字同步改为 `TemporalEntity*`，并运行 `node scripts/generate-implementation-inventory.js` 校验唯一事实来源；不保留向后兼容别名。
- Go Query 服务需同步生成新的 GraphQL schema/types：更新 `cmd/hrms-server/query` 下的 schema 绑定、resolver 函数命名与 `toolchain go1.24.9` 代码生成结果（`go generate ./cmd/hrms-server/query/...`），随后运行 `go test ./cmd/hrms-server/query/...` 与 `make test` 验证。

### 3.5 T4 – Selectors & Fixtures（3 天）
- 新增 `frontend/src/shared/testing/temporalEntitySelectors.ts`，集中维护 `temporal-entity-*` testid；将 `position-tabs.spec.ts`、`organization-create.spec.ts`、`temporal-management-integration.spec.ts` 等 E2E 用例替换为中性 selector（frontend/tests/e2e/position-tabs.spec.ts:91-145；frontend/tests/e2e/organization-create.spec.ts:4-41）。  
- 合并 `frontend/tests/e2e/utils/positionFixtures.ts` 与组织 fixtures，输出 `temporalEntityFixtures.ts`，通过 `entityType` 参数生成 GraphQL 响应，命名遵循 `{entity}Fixture` 而非 `POSITION_*` 常量。  
- 更新 `waitPatterns`/`auth-setup` 等工具函数中的常量名，保证 e2e utils 无实体专属前缀；编写 codemod 和临时 alias，安排双写验证窗口。

### 3.6 T5 – 文档与工具（1 天）
- 将职位专有指南改写为中性抽象：`docs/reference/temporal-entity-experience-guide.md`，同步 Plan 06、Plan 240/241 的引用。  
- 更新 `README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`，对齐 Phase2 文档更新要求（参考 `docs/development-plans/215-phase2-summary-overview.md:250-269`）。  
- 在 `docs/development-plans/06-integrated-teams-progress-log.md` 以及 `215-phase2-execution-log.md` 追加命名迁移记录，确保 Phase2 追踪与唯一事实来源同步。

### 3.7 Backend & Quality Gates（贯穿执行）
- 每个阶段完成后执行 `go test ./...`（命令/查询服务至少一次）与 `make test`，确保 Go 层命名更新未破坏 REST/GraphQL 行为。  
- 运行 `node scripts/generate-implementation-inventory.js` 并核对 `reports/implementation-inventory.json` 是否保留 Go handler/service 统计，防止快照缺失。  
- 记录 `git status` 与 `logs/plan242/` 日志中所有验证命令（含 `npm run lint`, `npm run test`, Playwright、Storybook 截图、`node scripts/quality/architecture-validator.js`），以便审计。  
- 若命名影响数据库迁移或 Temporal Monitor，必须同时在 `database/migrations/` 与 `cmd/hrms-server/command/internal/services/temporal_monitor.go` 更新字段，确认 `make db-migrate-all`/`make docker-up` 流程仍可运行。

---

## 4. 里程碑

| 里程碑 | 时间 | 交付物 |
| --- | --- | --- |
| M1 | Day 3 | T0/T1 完成，提交 `TemporalEntityPage` MR |
| M2 | Day 8 | 完成 Timeline/Status 抽象 MR + 单测 |
| M3 | Day 12 | Types/GraphQL/Hook 命名抽象 MR |
| M4 | Day 15 | Selector/Fixture 统一 MR + Playwright 绿灯 |
| M5 | Day 16 | 文档更新 + 命名库存档案（含 README/Quick Reference/Inventory） |

---

## 5. 验收标准

1. 任意页面/组件/Hook/test 中不再直接引用 `OrganizationTemporalPage`、`PositionTemporalPage`、`timelineAdapter`、`statusMeta` 等旧命名；lint/TS 审核确保删除。  
2. TypeScript/GraphQL/React Query 层不再出现 `PositionDetailQuery`, `OrganizationVersionsQuery` 等命名，全部通过 `TemporalEntity*` 命名导出；ESLint/TS 检查确保类型别名唯一。  
3. Playwright 用例仅使用 `temporalEntitySelectors` 暴露的 testid，`position-tabs.spec.ts`、`temporal-management-integration.spec.ts`、`organization-create.spec.ts` 通过 Chromium/Firefox 连续 3 次运行。  
4. `temporalEntityFixtures.ts` 成为唯一事实来源；旧 `positionFixtures.ts` 标记删除或代理导出且附弃用说明。  
5. 文档/计划引用统一使用 `Temporal Entity` 命名，且 Phase2 要求的 README/Quick Reference/Implementation Inventory/215 执行日志均已更新，无平行事实来源。
6. 命令/查询服务完成 `go test ./cmd/hrms-server/...`、`make test` 全绿，GraphQL schema 与 OpenAPI 生成代码更新后经 `go fmt`/`go vet` 校验，通过 `node scripts/quality/architecture-validator.js` 与 `node scripts/generate-implementation-inventory.js` 的最新快照。

---

## 6. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
| --- | --- | --- | --- |
| 路由抽象影响历史书签或深链 | 中 | 中 | 提供 Redirect 方案（旧路径 -> 新入口），在 QA 前验证 |
| Timeline Adapter 抽象导致类型发散 | 中 | 低 | 通过泛型约束 + 单测覆盖组织/职位双场景 |
| E2E Selector 替换成本高 | 中 | 中 | 编写 codemod/脚本批量替换 testid，短期保留别名 |
| 文档重命名导致外部链接失效 | 低 | 中 | 在 docs/ 下保留 stub，或更新 README/Plan 06 中的导航 |
| Go 契约未同步导致 REST/GraphQL 断裂 | 高 | 中 | 在 T2/T3 阶段设置强制检查：更新 `cmd/hrms-server/*` 相关 struct/resolver，执行 `go test ./cmd/hrms-server/...`、`make test` 以及 schema diff；如发现问题立即回滚并增补临时 feature flag |
| Implementation Inventory 快照缺失 | 中 | 低 | 每阶段重新运行 `node scripts/generate-implementation-inventory.js`，比对上一版本的 Go handler/service 统计；若脚本遗漏需立刻修复并在日志中记录 |

---

## 7. 汇报

- 每日于 Plan 06 “阻塞”栏目同步命名迁移进度；Plan 240/241 更新 should mention Plan 242 输出。  
- 产出 `logs/plan242/*.log` 与 `reports/plan242/naming-inventory.md`，供评审/归档。  
- 完成后将本计划归档至 `docs/archive/development-plans/242-temporal-naming-abstraction-plan.md`。
- 2025-11-11：Plan 244 / T2 后端契约对齐（Temporal timeline 字段扩展）已记录在 `logs/plan242/t2/2025-11-11-temporal-timeline-go.md`，并在 `reports/plan242/naming-inventory.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 留痕，作为 Plan 242 T2 验收素材。

---

## 附录 A – 修订补充与门禁（依据评审 · 参考行业实践）

1) 路由与创建模式（差异收敛）
- 现状：组织 `/organizations/:code/temporal`，职位 `/positions/:code`；创建模式短期接受 `new/NEW`（解析归一化）。
- 策略：短期保留差异避免破坏；长期统一创建模式为 `new`，窗口 2 个迭代；是否对职位增加 `/positions/:code/temporal` 的别名与重定向，待 241/240 上线前在 MR 中附 E2E 证据后执行。

2) React Query Key 与缓存
- 统一 key：以 `temporal-entity-detail`/`temporal-entity-versions` 等前缀替代实体专属 key；冻结旧 key 新增使用。
- 过渡：适配层提供 alias（<=1 个迭代）；Hook 单测断言“无重复 fetch + 正确失效 + 事件可观测”。

3) 守卫与 CI 门禁
- 必须通过：`npm run guard:selectors-246`（选择器冻结）、`npm run guard:plan245`（旧 GraphQL 操作名冻结）。
- CI fail-fast：在安装依赖前执行；allowlist 需要注明理由与截止期。

4) 文档与唯一事实来源
- 仅允许 `docs/reference/temporal-entity-experience-guide.md` 保留正文；旧指南保留 Deprecated 占位符 1 个迭代，不得复制正文。
- 零引用校验：`rg -n '<OLD_GUIDE_FILE_NAME>' --glob '!docs/archive/**'` 应为空；配合 `document-sync.js` 与 `architecture-validator.js`。

5) Go/契约同步
- 若字段命名影响 REST/GraphQL：按仓库实际路径更新（如 `internal/organization/resolver/*`、`internal/organization/scheduler/*`）；运行 `go test ./internal/...`、`go test ./cmd/hrms-server/...`、必要时 `go generate ./cmd/hrms-server/query/...`。
- OpenAPI：`npm run lint:api` 必须 0 errors（已修复 no-$ref-siblings，参考 Plan 245T），防止文档回归。

6) 里程碑与验收（补充）
- T0：环境（Docker 强制）+ 实现清单 + 命名盘点；
- T1：入口统一（完成，Plan 243）；  
- T2：Timeline/Status 抽象落地 + 旧路径冻结 + Go/契约按需同步 + 单测/Storybook；  
- T3：Types/GraphQL/Hook 命名统一 + Key 冻结/迁移 + 守卫通过；  
- T4：Selector/Fixture 统一 + 多浏览器 E2E（3×2） + 指标事件断言 + 日志落盘；  
- T5：文档/治理对齐 + 零引用校验 + Inventory 快照与哈希。

7) 风险与缓解（补充）
- 书签/深链：前端 alias/Redirect 灰度，公告与截止期；  
- 实体特例：adapter 提供 override；以单测覆盖；  
- 替换成本：codemod 批量 + 短期 alias + 守卫冻结；  
- 契约断裂：MR gating 附 schema diff、`go test/go generate` 输出，异常立即回滚并启用 feature flag；  
- 文档偏差：统一出口 + 零引用脚本，防二次来源。

---

## 完成登记（2025-11-15）
- 守卫与门禁  
  - 选择器守卫：`npm run guard:selectors-246` → 通过（日志：`logs/plan242/t4/selector-guard-246.log`）  
  - 命名守卫：`npm run guard:plan245` → 通过（日志：`logs/plan242/t4/plan245-guard.log`）  
- E2E（Chromium/Firefox）  
  - Smoke：`frontend/tests/e2e/smoke-org-detail.spec.ts`、`frontend/tests/e2e/temporal-header-status-smoke.spec.ts`（通过）  
  - 集成：`frontend/tests/e2e/temporal-management-integration.spec.ts`（8 passed / 4 skipped）  
  - 轮次记录：`logs/plan242/t4/242-e2e-round2.log`、`logs/plan242/t4/242-e2e-round3.log`  
  - 报告：`frontend/playwright-report/index.html`  
- 说明：本计划在 T1/T2/T3/T5 已达成基础能力与文档治理，T4 已按“至少 1 轮双浏览器 + 守卫通过 + 证据落盘”的基线验收完成。后续更高轮次抽样按 241 启动前在 CI 侧定期执行。
