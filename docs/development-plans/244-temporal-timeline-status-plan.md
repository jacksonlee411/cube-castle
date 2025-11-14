# Plan 244 – Temporal Entity Timeline & Status 抽象

**关联主计划**: Plan 242（T2）  
**目标窗口**: Day 3-8  
**范围**: 统一 Timeline Adapter 与状态元数据命名

## 背景
- 职位模块沿用旧版 timeline/status 适配器，仅覆盖 `PositionRecord`  
- 组织模块使用 `shared/utils/statusUtils.ts` 与独立 Timeline 结构  
- Phase2 强调组织模块标准化（Plan 219），需统一命名与类型

## 前提依赖与对齐
- **Plan 242 T1 完成后方可开启本计划**：需先合并 `TemporalEntityPage`/路由抽象与统一命名入口，避免旧页面仍消费新 Adapter 产生交叉状态。
- **命名库存对齐**：继承 Plan 242 T0 的 `reports/plan242/naming-inventory.md` 结果，用于核实所有 Timeline/Status 触点是否已纳入迁移范围。
- **执行日志同步**：与 Plan 242 共用 `logs/plan242/t2/`，每日状态写入 `docs/development-plans/215-phase2-execution-log.md`，保持唯一事实来源。

## 工作内容
1. **TemporalEntityTimelineAdapter**：抽象 `createTemporalTimelineAdapter`，覆盖组织/职位，提供 default mapper + entity override。  
2. **TemporalEntityStatusMeta**：集中 `statusConfig`，提供 `position`、`organization` 配置及扩展点。  
3. **引用更新**：Position/Organization 组件、Storybook、Vitest、GraphQL loader 全面替换。  
4. **Lint 规则**：新增 lint 以阻止引用旧 `timelineAdapter`/`statusMeta`。  
5. **回归**：Storybook 截图、Vitest/Playwright 基线更新。
6. **后端同步**：  
   - 命令/查询服务中的 Timeline/Status DTO、GraphQL resolver 与 REST handler 必须同步重命名，更新 `cmd/hrms-server/command/internal/services/temporal*.go`、`cmd/hrms-server/query/internal/resolvers/*` 以及相应 schema 生成文件（`go generate ./cmd/hrms-server/query/...`）。  
   - 更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 后运行 `make test`、`go test ./cmd/hrms-server/...`、`go vet ./cmd/hrms-server/...` 验证契约一致。  
   - 若 Timeline/Status 字段触及数据库/Temporal Monitor，需在 `database/migrations/` 与 `cmd/hrms-server/command/internal/services/temporal_monitor.go` 同步字段并执行 `make db-migrate-all`/`make docker-up`.

## 契约与文档同步
- 更新 `docs/api/schema.graphql` 与 `docs/api/openapi.yaml`，将 Timeline/Status 相关 operation/字段命名统一为 `TemporalEntity*`，并附带 diff 说明。
- 跟进 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/reference/temporal-entity-experience-guide.md` 等引用，删除旧 `timelineAdapter/statusMeta` 的命名残留，并运行 `node scripts/generate-implementation-inventory.js` 确认快照包含最新 Go handler/service/TS 导出。
- 在 `reports/plan242/naming-inventory.md` 中追加 Timeline/Status 改动记录，作为单一事实来源；完成后将计划文档归档并在 `CHANGELOG.md`/Plan 06 进展日志提及。

## 里程碑 & 验收
- Day 6：完成 Adapter/Status 重构 MR + 单测  
- Day 8：完成引用替换与 lint 规则  
- 验收标准：  
  - `rg "timelineAdapter"|statusMeta` 仅在新命名空间出现；组织/职位 UI 行为一致。  
  - `go test ./cmd/hrms-server/...`、`make test`、`npm run lint`、`npm run test`、`npm run test:e2e -- --project=chromium --project=firefox`（各至少一次）全部通过。  
  - `docs/api/*.yaml`、`docs/api/schema.graphql` 的 `TemporalEntity*` 改名与 Go 生成代码保持一致，并在 `reports/plan242/naming-inventory.md`、`logs/plan242/t2/` 记录验证命令。

## 测试与验证
- **单元/组件测试**：新增 Adapter 与 StatusMeta 的 TypeScript/Vitest 覆盖，确保 position/organization 双实体映射一致。
- **前端集成**：运行 `npm run test`、`npm run lint`、`npm run test:e2e -- --project=chromium --project=firefox` 连续三次，验证新 selector/lint 规则。
- **Storybook/截图基线**：重新生成组织与职位 Timeline/Status 组件截图，并将比较结果附加到 MR。
- **命令链路**：如影响 GraphQL Loader，需执行 `make test`/`make test-integration` 与 `node scripts/quality/architecture-validator.js`，保证后端契约与前端使用保持一致。
- **Go 层验证**：在命名替换完成后，执行 `go test ./cmd/hrms-server/...`、`go vet ./cmd/hrms-server/...`，并对 Query 服务运行 `go generate ./cmd/hrms-server/query/...` + `go test ./cmd/hrms-server/query/...`，确保 schema 与 resolver 同步。

## 汇报
- 在 `215-phase2-execution-log.md` 记录阶段性节点；日志输出 `logs/plan242/t2/`.

## 风险与回滚
| 风险 | 描述 | 缓解/回滚 |
| --- | --- | --- |
| 路由抽象尚未合并导致引用错位 | 旧 `OrganizationTemporalPage` 仍被消费 | 阶段检查 Plan 242 T1 合并状态，必要时在 MR 中加守卫并阻塞合并 |
| Timeline Adapter 泛化破坏实体特有逻辑 | 组织/职位状态颜色或排序异常 | 双实体截图/单测对比，提供 feature flag 开关，问题时回滚至上一个 tag |
| ESLint 规则误伤遗留代码 | 构建失败阻塞 | linter 先以 warning 模式运行并列出允许列表，确认引用清零后再切换为 error |
| Playwright testid 全量替换失败 | E2E flakiness 或漏测 | 引入临时 alias (`temporalEntitySelectors.position.*`) 并设置弃用期限，确保多浏览器跑通后移除 |
| Go 契约未同步导致 REST/GraphQL 断裂 | 命令/查询服务未更新字段或 schema | 在 T2 MR 中设置 gating：合并前必须附 `go test ./cmd/hrms-server/...`、`go generate` 输出及 schema diff；若出现回归，立即回滚至上一 tag，并通过 feature flag 限流 |
| Implementation Inventory 漏更 | `reports/implementation-inventory.json` 缺失 Go handler/service | 每日运行 `node scripts/generate-implementation-inventory.js` 并将结果附在 `logs/plan242/t2/`；如脚本出错，先修复脚本再继续开发 |

## 完成情况摘要
- 2025-11-11：TemporalTimelineManager/REST handler 已输出 `unitType/level/codePath/namePath/sortOrder` 等 `TemporalEntityTimelineVersion` 字段，OpenAPI/GraphQL/Implementation Inventory 与命名清单同步更新，详见 `logs/plan242/t2/2025-11-11-temporal-timeline-go.md`。
- 验证命令：`go test ./cmd/hrms-server/...`、`make test`、`npm run lint`、`cd frontend && npm run test`、`node scripts/quality/architecture-validator.js`、`node scripts/generate-implementation-inventory.js`。
- 受限于容器环境未安装 Playwright 浏览器，`npm run test:e2e -- --project=chromium --project=firefox` 需在本地工作站/CI 上补跑三轮以满足 Plan 244 验收条款。
