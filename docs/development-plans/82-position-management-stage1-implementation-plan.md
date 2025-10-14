# 82号文档：职位管理 Stage 1 实施计划（数据库 + 后端契约落地）

**版本**: v0.1 草案  
**创建日期**: 2025-10-15  
**维护团队**: 命令服务团队 + 查询服务团队 + 前端团队（协同）  
**关联计划**: 80号职位管理方案、81号契约更新方案  
**遵循原则**: CLAUDE.md / AGENTS.md / 81号计划验收标准

---

## 1. 背景与目标

- Stage 0 已完成前端原型与交互验收；Stage 1 需要将职位管理契约落实到数据库与服务实现。  
- 81 号计划已将 OpenAPI (v4.7.0) 与 GraphQL (v4.7.0) 契约合并主干，本阶段需支撑该契约的实际 CRUD / 查询能力。  
- 目标：在保持租户隔离与时态一致性的前提下，实现职位与 Job Catalog 的数据库结构、命令服务、查询服务，并提供最小可用的后端接口供前端接入。

---

## 2. 范围与交付物

### 2.1 必做交付物（Blocking）

1. **数据库层**  
   - 编写并执行 `043_create_positions_and_job_catalog.sql`（positions + job_catalog 表、索引、约束）。  
   - 迁移脚本通过 `make db-migrate-all`、并在 Stage 1 完成后重新运行租户隔离 SQL 巡检。  

2. **命令服务（REST）**  
   - 仓储层：新增 PositionRepository、JobCatalogRepository（插入、版本管理、租户校验）。  
   - 服务层：实现 PositionService（创建、替换、版本、事件、填充/清空/转移—临时）、JobCatalogService（家族/职务/职级版本管理）。  
   - 处理器：实现 `/api/v1/positions*` 与 `/api/v1/job-*` 路由、请求校验、审计、错误处理；返回 `SuccessResponse` / `PositionSuccessResponse`。  

3. **查询服务（GraphQL）**  
   - Schema / resolver：实现 `positions`, `position`, `positionTimeline`, `vacantPositions`, `positionHeadcountStats`、以及 Job Catalog 查询。  
   - 数据源：编写 SQL / Query 层，确保分页、排序、过滤、租户隔离与时态派生逻辑。  

4. **验证与巡检**  
   - 质量门禁（lint:api、contract:generate、architecture-validator、field-naming）。  
   - 重新执行 `docs/development-plans/81-tenant-isolation-checks.sql` 并归档真实结果。  
   - 更新 `reports/contracts/position-api-diff.md`、`reports/implementation-inventory.json`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`。  

### 2.2 可选扩展（P2，Stage 1.1 之后）

- 前端 Hook / UI 接入（Stage 1.1 与 Stage 2 并行）。  
- Position 事件审计报表。  
- Job Catalog 同步任务（外部系统接入）。  

---

## 3. 工作拆解（阶段 / 任务子项）

| 阶段 | 任务 | 主要责任人 | 依赖 | 输出 |
|------|------|------------|------|------|
| Phase 1 | 数据库迁移与回滚策略 | DB 工程师 + 架构组 | 81 号计划 | `043_create_positions_and_job_catalog.sql`、迁移执行记录、回滚脚本草案 |
| Phase 2 | 数据模型与仓储层 | 命令服务后端 | Phase 1 | PositionRepository、JobCatalogRepository、接口单测 |
| Phase 3 | 命令服务业务层 & Handler | 命令服务后端 | Phase 2 | PositionService、JobCatalogService、handler、审计日志 |
| Phase 4 | 查询服务 Schema/Resolver | 查询服务后端 | Phase 1/2 | gqlgen 更新、resolver、数据访问层、单测 |
| Phase 5 | Smoke 测试与质量门禁 | 命令+查询+QA | Phase 2/3/4 | 集成测试（REST + GraphQL）、lint、contract:generate、architecture-validator |
| Phase 6 | 租户隔离巡检与资料更新 | 命令服务后端 + 架构组 | Phase 5 | SQL 巡检结果、实现清单与差异报告 |

---

## 4. 详细任务说明

### Phase 1：数据库迁移
- 参考 80 号文档 §3.3/3.4，创建 `job_family_groups/job_families/job_roles/job_levels/positions` 表与索引。  
- 引入复合外键（record_id + tenant_id）与 `UNIQUE ... WHERE is_current` 约束，保障租户隔离/单当前版本。  
- 执行 `make db-migrate-all`，并记录迁移日志；准备 `rollback` 草案（如需）。  

#### ✅ 进度更新（2025-10-16）
- [x] 调整 `043_create_positions_and_job_catalog.sql`，补充 `record_id + tenant_id` 唯一约束与 `positions` 复合唯一键，满足 Stage 1 复合外键约束需求。
- [ ] 执行 `make db-migrate-all` 并归档迁移日志（待仓储与服务层接入完成后统一验证）。

### Phase 2：仓储层
- 封装数据库读写：含插入、更新、版本管理、并发控制（SELECT ... FOR UPDATE）、校验 `tenant_id` 对齐。  
- 定义结构化错误并统一返回（示例：`ErrTenantMismatch`、`ErrVersionConflict`、`ErrPositionStatusTransition`、`ErrJobCatalogMissing`），服务层据此映射 HTTP 状态码。  
- 单元测试：使用事务回滚或 mock 驱动。  

#### ✅ 进度更新（2025-10-16）
- [x] 新增 `JobCatalogRepository` 与 `PositionRepository`，实现时态重算、版本插入、状态更新等基础方法。
- [x] 仓储层接口通过 `go test ./...` 编译验证。

### Phase 3：命令服务
- PositionService：创建/替换、版本管理、事件处理、临时 Fill/Vacate/Transfer（审计与租户校验）。  
- JobCatalogService：四级分类 CRUD + 版本（create + versions）。  
- Handler：路由注册、请求体验证（使用 `go-playground/validator` 或自建校验）、统一响应 `utils.ResponseBuilder`。  
- 临时端点 Fill/Vacate/Transfer 需添加 `// TODO-TEMPORARY` 注释（引用 17 号治理计划），并在 Stage 1 收尾清单中记录 deadline 与回收方案。  
- 审计：复用 auditLogger，记录 actor、tenant、operationType、reason。  

#### ✅ 进度更新（2025-10-16）
- [x] 实现 `PositionService` / `JobCatalogService`，覆盖创建、版本管理、Fill/Vacate/Transfer 等业务逻辑。
- [x] 完成职位与 Job Catalog REST Handler 接入、统一企业级响应与审计事件记录。

### Phase 4：查询服务
- 更新 gqlgen schema（已由 81 号计划完成契约）；生成 `graphql-types.ts`。  
- Resolver：实现 positions & Job Catalog 查询、支持分页/排序/过滤、头部注入 `X-Tenant-ID`。  
- 性能：依赖索引 + 只读隔离；positions 查询默认 pageSize=20、最大 100（与组织查询保持一致），确保 GraphQL 层返回 camelCase 字段。  

#### ✅ 进度更新（2025-10-16）
- [x] 新增 GraphQL Resolver，支持职位列表/详情/时间线/空缺及编制统计查询。
- [x] 构建职位、职位分类模型与查询仓储，覆盖日期过滤、排序与租户隔离逻辑。

### Phase 5：质量门禁
- 自动化：`npm run lint:api`、`npm --prefix frontend run contract:generate`、`validate:schema`、`node scripts/quality/architecture-validator.js`、字段命名脚本。  
- 集成测试：最小验证流程（创建 -> 版本 -> 查询）。  
- 记录命令输出，归档至 `reports/contracts/position-api-diff.md`、`reports/implementation-inventory.json`。  

#### ✅ 进度更新（2025-10-16）
- [x] 执行 `gofmt` 与 `go test ./...`，当前命令与查询实现编译通过。
- [ ] 生成差异报告与前端契约校验（待 Stage 1 收尾统一执行）。

### Phase 6：租户隔离巡检与资料更新
- Stage 1 迁移后执行 `81-tenant-isolation-checks.sql`，输出 `tenant-isolation-check-stage1-YYYYMMDD.sql`。  
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md` 第10节执行记录。  
- 完成 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`、`docs/development-plans/81-position-api-contract-update-plan.md` 第 10 节最后两项勾选。  

---

## 5. 进度与里程碑（建议日期）

| 里程碑 | 内容 | 预计完成日 | 状态 |
|--------|------|------------|------|
| M1 | 数据库迁移脚本合并 | 2025-10-16 | ☐ |
| M2 | 命令服务 Position/JobCatalog 实现 | 2025-10-21 | ✅ |
| M3 | 查询服务 Position/JobCatalog resolver 完成 | 2025-10-23 | ✅ |
| M4 | 集成测试 & 质量门禁通过 | 2025-10-24 | ☐ |
| M5 | Stage 1 租户隔离巡检归档 | 2025-10-25 | ☐ |

---

## 6. 验收标准

- [ ] 数据库迁移执行完成，`positions` 与 job catalog 表结构与 80 号方案一致。  
- [ ] 命令服务所有 REST 端点按照 OpenAPI v4.7.0 返回成功/错误响应，并通过集成测试。  
- [ ] 查询服务 (GraphQL) 能返回职位与 Job Catalog 数据，复杂过滤/排序/分页正常。  
- [ ] 租户隔离巡检 SQL 全部返回空集。  
- [ ] `reports/contracts/position-api-diff.md`、`reports/implementation-inventory.json`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 同步最新端点/查询。  
- [ ] `docs/development-plans/81-position-api-contract-update-plan.md` 第 10 节余下项完成勾选。  

---

## 7. 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| 租户隔离约束遗漏 | 数据泄露 | 中 | 使用复合外键 + 巡检 SQL + 集成测试双重保障 |
| 时态逻辑复杂导致并发冲突 | 数据不一致 | 中 | 引入 SELECT FOR UPDATE + 乐观锁 (If-Match) |
| 临时 Fill/Vacate 流程超期 | 临时方案遗留 | 中 | 在代码中加 `// TODO-TEMPORARY`，记录 deadline 并纳入 17 号治理计划 |
| GraphQL 查询性能不足 | 体验下降 | 低 | 使用现有组织层索引模式 + LIMIT/OFFSET 分页 |

---

## 8. 参考资料

- `docs/development-plans/80-position-management-with-temporal-tracking.md` — 数据模型、状态机、权限定义  
- `docs/development-plans/81-position-api-contract-update-plan.md` — Stage 0/4 契约要求与质量门禁  
- `docs/api/openapi.yaml`、`docs/api/schema.graphql` v4.7.0 — REST/GraphQL 契约  
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md` — 现有端点与查询清单  
- `docs/development-plans/06-integrated-teams-progress-log.md` — 评审结论与 Stage 验收记录  

---

> 本计划将 Stage 1 拆分为数据库、命令服务、查询服务、质量保障、租户巡检五个子任务。每项完成后请同步更新里程碑与验收项，确保 Stage 1 在租户隔离与时态一致性前提下高质量落地。
