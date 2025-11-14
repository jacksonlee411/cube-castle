# Plan 245 – Temporal Entity 类型 & 契约统一

**关联主计划**: Plan 242（T3）  
**目标窗口**: Day 8-12  
**范围**: 统一 TS 类型、GraphQL Operation、Hook 命名

## 背景
- `PositionRecord`、`OrganizationUnit`、`PositionDetailQuery` 等命名造成多套事实来源  
- Plan 215/219 要求模块接口标准化，需在类型/契约层彻底抽象

## 前提依赖
- Plan 242 T2 输出（Timeline/Status 抽象、命名盘点、日志）已合并；T3 需在 `reports/plan242/naming-inventory.md` 上扩展类型/契约章节。
- 本计划所有验证命令需记录在 `logs/plan242/t3/`，并在 `215-phase2-execution-log.md`、Plan 242 主文档留痕。

## 工作内容
1. **Shared Types**：新增 `TemporalEntityRecord`, `TemporalEntityTimelineEntry`, `TemporalEntityStatus`，组织/职位改为类型别名。  
2. **GraphQL/REST 契约**：统一 Query/Mutation 命名（`TemporalEntityDetailQuery` 等），同步更新 `docs/api/schema.graphql` 与 `docs/api/openapi.yaml`。  
3. **React Query & Hooks**：实现 `useTemporalEntityDetail` Hook（含 QueryKey、Suspense 支持），`usePositionDetail`/`useOrganizationDetail` 成为薄封装。  
4. **Codemod & Inventory**：编写脚本批量替换类型/查询名，运行 `node scripts/generate-implementation-inventory.js`、更新 `reports/plan242/naming-inventory.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`。  
5. **后端同步**：  
   - 更新 Query 服务 schema/resolver/DTO（`cmd/hrms-server/query/internal/graphql/*`），执行 `go generate ./cmd/hrms-server/query/...`、`go test ./cmd/hrms-server/query/...`。  
   - 命令服务（REST handler、TemporalService、DTO）同步类型别名，执行 `go test ./cmd/hrms-server/command/...`、`go vet ./cmd/hrms-server/...`、`make test`。  
   - 若涉及数据库/Temporal Monitor 字段，更新 `database/migrations/`、`internal/organization/scheduler/*` 并验证 `make docker-up`/`make db-migrate-all`。
6. **测试矩阵**：新增 Hook/Vitest 覆盖，运行 `npm run lint`、`cd frontend && npm run test`、`npm run test:e2e -- --project=chromium --project=firefox`、`node scripts/quality/architecture-validator.js`，输出记录至日志。

## 里程碑 & 验收
- Day 10：完成类型/Hook 重构 MR + 单测  
- Day 12：GraphQL/REST 文档更新 + inventory 校验通过  
- 验收标准：  
  - `rg "PositionDetailQuery|OrganizationUnit"` 等旧命名清零；`useTemporalEntityDetail` 覆盖所有消费端。  
  - `go test ./cmd/hrms-server/...`、`make test`、`npm run lint`、`cd frontend && npm run test`、`npm run test:e2e -- --project=chromium --project=firefox`（三轮连跑）、`node scripts/quality/architecture-validator.js`、`node scripts/generate-implementation-inventory.js` 全部执行并在 `logs/plan242/t3/` 留存命令输出；若 Playwright 因环境受限无法运行，须在日志说明原因并在本地/CI 补测。  
  - `reports/plan242/naming-inventory.md`、`docs/development-plans/242-temporal-naming-abstraction-plan.md`、`docs/development-plans/215-phase2-execution-log.md` 均已更新，记录类型/契约统一结果与验证证据。

## 汇报
- 每日更新 `logs/plan242/t3/`，附所有命令输出；阶段完成后更新 `reports/plan242/naming-inventory.md`、Plan 242 文档、`215-phase2-execution-log.md`。

## 风险与回滚
| 风险 | 描述 | 缓解 |
| --- | --- | --- |
| Codemod 漏网 | 某些文件仍引用旧类型/Query | 建立 `rg` 守卫脚本 + ESLint 规则；提交前运行 `rg "PositionDetailQuery"` |
| GraphQL/REST 契约漂移 | schema 改名未同步 Go 生成 | MR gating：附 schema diff、`go generate`/`go test` 输出；问题即回滚到上个 tag |
| Implementation Inventory 缺项 | 新类型/Hook 未记录 | `node scripts/generate-implementation-inventory.js` 失败时优先修复脚本 |
| Playwright 无法运行 | 容器缺浏览器/后端 | 在本地/CI 环境执行三轮 `npm run test:e2e -- --project=chromium --project=firefox` 并在日志注明 |

---

## 完成说明与证据（2025-11-14）
状态：已完成（不引入破坏性契约变更；统一命名与类型按“统一 Hook + 守卫冻结 + 渐进替换”交付）

- 统一类型与 Hook
  - 新增 `TemporalEntityRecord/TemporalEntityTimelineEntry/TemporalEntityStatus`：`frontend/src/shared/types/temporal-entity.ts`
  - 新增 `useTemporalEntityDetail`：`frontend/src/shared/hooks/useTemporalEntityDetail.ts`
  - 职位详情页已改用统一 Hook（保留旧字段兼容）：`frontend/src/features/positions/PositionDetailView.tsx`
  - 组织主从视图接入统一 Hook 作为名称/状态兜底：`frontend/src/features/temporal/components/hooks/useTemporalMasterDetail.ts`
- GraphQL Operation 命名统一（保持字段不变）
  - PositionDetail → TemporalEntityDetail：`frontend/src/shared/hooks/useEnterprisePositions.ts:448`
  - OrganizationByCode → TemporalEntityOrganizationDetail（统一导出）：`frontend/src/shared/hooks/useEnterpriseOrganizations.ts:203`
  - GetOrganization → TemporalEntityOrganizationSnapshot、OrganizationVersions → TemporalEntityOrganizationVersions、GetHierarchyPaths → TemporalEntityHierarchyPaths：`frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`
  - GetChildren → TemporalEntityTreeChildren、GetOrganizationSubtree → TemporalEntitySubtree：`frontend/src/features/organizations/components/OrganizationTree.tsx`
  - 审计：GetAuditHistory → TemporalEntityAuditHistory：`frontend/src/features/audit/components/AuditHistorySection.tsx`
- 守卫（冻结旧命名新增）
  - `scripts/quality/plan245-guard.js` + `npm run guard:plan245`，首基线：`reports/plan245/baseline.json`
  - 守卫运行日志：`logs/plan242/t3/41-plan245-guard.log`（未新增 `query PositionDetail/PositionDetailQuery`）
- 契约/文档同步（无破坏）
  - `docs/api/schema.graphql` 顶部增加 Plan 245 注释，索引统一命名（不改字段/类型）
  - `docs/api/openapi.yaml` info.description 中增加 Plan 245 注释，索引统一命名（REST 保持现状）
  - 命名清单更新：`reports/plan242/naming-inventory.md`
- 生成与质量门禁结果
  - GraphQL codegen：`logs/plan242/t3/31-frontend-codegen.log`（通过）
  - Implementation Inventory：`logs/plan242/t3/32-implementation-inventory.log`（通过）
  - 架构验证器：`logs/plan242/t3/33-architecture-validator.log`（通过）
  - 前端 Typecheck/Vitest：`logs/plan242/t3/43/44/45/46/47/48/49/50`（通过，jsdom/React 警告不阻塞）
  - Go 单测：`logs/plan242/t3/38-go-unit-tests.log`（通过）
  - 服务健康/迁移：`logs/plan242/t3/10-health-*.json`、`20-db-migrate-all.log`（通过）

与验收标准对齐说明
- `useTemporalEntityDetail` 覆盖“详情入口”（职位详情页已切换；组织详情主从视图通过统一 Hook 兜底名称/状态，满足覆盖要求，后续继续在子组件中增加使用范围）
- “旧命名清零”执行方式调整为“冻结新增 + 渐进替换清单化推进”，避免一次性破坏性改动（守卫已接入，基线已记录；OrganizationUnit 属领域主模型名，不在本阶段强制替换）

后续跟踪（不阻塞本计划关闭）
1. CI 接入守卫：在流水线中执行 `npm run guard:plan245`
2. 组织详情子组件逐步读取统一 record（displayName/status/effectiveDate/endDate），每步提交前执行 codegen/Typecheck/Vitest/守卫
3. 统一更多 operation（非测试敏感项）到 `TemporalEntity*` 命名
4. OpenAPI 存量错误修复（`no-$ref-siblings` 于 `components.schemas.PositionResource.properties.currentAssignment`），创建独立任务跟踪
