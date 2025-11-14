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
  - `go test ./cmd/hrms-server/...`、`make test`、`npm run lint`、`cd frontend && npm run test`、`npm run test:e2e -- --project=chromium --project=firefox`、`node scripts/quality/architecture-validator.js`、`node scripts/generate-implementation-inventory.js` 均通过。  
  - `reports/plan242/naming-inventory.md` 与日志中记录 codemod/验证命令，Plan 242 文档同步更新。

## 汇报
- 每日更新 `logs/plan242/t3/`；阶段完成后更新 `reports/plan242/naming-inventory.md`、Plan 242 文档、`215-phase2-execution-log.md`。

## 风险与回滚
| 风险 | 描述 | 缓解 |
| --- | --- | --- |
| Codemod 漏网 | 某些文件仍引用旧类型/Query | 建立 `rg` 守卫脚本 + ESLint 规则；提交前运行 `rg "PositionDetailQuery"` |
| GraphQL/REST 契约漂移 | schema 改名未同步 Go 生成 | MR gating：附 schema diff、`go generate`/`go test` 输出；问题即回滚到上个 tag |
| Implementation Inventory 缺项 | 新类型/Hook 未记录 | `node scripts/generate-implementation-inventory.js` 失败时优先修复脚本 |
| Playwright 无法运行 | 容器缺浏览器/后端 | 在本地/CI 环境执行三轮 `npm run test:e2e -- --project=chromium --project=firefox` 并在日志注明 |
