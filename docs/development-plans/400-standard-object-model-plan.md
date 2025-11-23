# 400 · 标准对象模型统一方案（Standard Object Model）

**版本**: v0.1  
**创建日期**: 2025-11-25  
**状态**: 草案（待立项）  
**责任人**: 架构组（对接 organization/workforce/contract 模块负责人）  
**关联文档**: `docs/development-plans/200-Go语言ERP系统最佳实践.md`、`docs/development-plans/201-Go实践对齐分析.md`、`docs/development-plans/203-hrms-module-division-plan.md`、`docs/reference/temporal-entity-experience-guide.md`

---

## 1. 背景与动机

- Core HR 域目前已上线组织（Organization）与职位（Position）模块，但二者在命名、生命周期、API、UI 骨架上存在差异；这导致版本管理、审核、Playwright 证据等重复维护，难以扩展到 workforce/contract。
- 200 号文档强调“模块化单体 + DDD + 统一端口/适配器”，201 号分析也指出“端口/适配器模式尚未固化”“Temporal 实体体验需复用”。我们需要以标准对象模型（Standard Object Model，后简称 SOM）为 SSoT，把组织、职位视为同类对象，统一生命周期、属性、行为与 UI 通路，以便 Phase3/4 直接扩展。
- 本方案编号 400，用于交付可复用的 SOM 能力，并将其纳入 203 计划的公共能力层。

---

## 2. 目标与验收标准

| 目标 | 验收 | 备注 |
|------|------|------|
| 建立可复用的 SOM 元模型 | `internal/standardobject/api.go` 定义 `ObjectService`、`LifecyclePolicy`、`MetadataRepository` 等接口，命令/查询服务在启动时注入 | 满足 200 文档“端口/适配器 + 编译期边界”要求 |
| 统一生命周期与状态机 | REST/OpenAPI 新增 `/standard-objects/{type}` 契约，GraphQL 暴露 `StandardObject` + `StandardObjectVersion` 类型，字段含 `status`、`effectiveInterval`、`parentLink` | 契约先于实现，更新 `docs/api/*` 并通过 Plan 258/259 守卫 |
| 数据与事件一致 | PostgreSQL 引入 `standard_objects`、`standard_object_versions`、`standard_object_links` 三张表及 sqlc 生成物；命令服务写入 outbox 事件 `standard_object.*` | 满足 200 文档“迁移即真源 + sqlc 类型安全” |
| UI/UX 复用 | 前端 `TemporalEntityLayout` 接入 `StandardObjectAdapter`，组织/职位页面通过配置项注入标签/权限，Playwright 选择器统一 `temporalEntity-*` | 与 `docs/reference/temporal-entity-experience-guide.md` 一致 |
| 迁移现有模块 | 组织/职位模块调用 SOM Port，不再直接持有仓储；旧仓储逐步迁往 `internal/standardobject/repository` 并由 sqlc 生成，并同步构建闭包表/快照表，确保读写一致 | 确保无重复逻辑；保留回滚策略 |

验收需满足：① `make test`/`make test-db`/`npm run test`/`npm run test:e2e` 全部通过；② `scripts/quality/*`（Plan 252/255/259）零回归；③ 提供 `logs/plan400/*` 运行记录（迁移、集成测试、Playwright OBS）。

---

## 3. 范围与非目标

### 3.1 范围
1. 组织（organization.unit）与职位（position.role）对象，后续 workforce（employee、assignment）可复用同一模型。
2. 生命周期：草稿 → 就绪 → 生效 → 休眠/冻结 → 归档，覆盖版本创建、调度、版本比较。
3. 数据契约：OpenAPI/GraphQL/数据库/事件命名统一，包含层级关系（parent-child、position-to-organization）与属性包。
4. UI 套件：Temporal Entity 页面框架、观察事件、选择器、命令/查询 API 对接。

### 3.2 非目标
- 不处理 Payroll/Compensation/Performance 等领域对象，它们需要独立计划。
- 不引入新的对象存储（保持单一 PostgreSQL）；也不在本阶段升级 GraphQL 引擎。
- 不替换既有权限守卫（Plan 252/259），仅提供对象级 scope，使其可被守卫消费。

---

## 4. 标准对象模型（SOM）设计

### 4.1 元模型组成
| 模块 | 说明 | 对应产物 |
|------|------|----------|
| `ObjectKernel` | 核心对象结构，含 `objectType`, `code`, `displayName`, `status`, `tenant`, `labels`, `createdBy` | `internal/standardobject/domain/object.go` |
| `TemporalVersion` | 版本信息：`versionCode`, `validity_range`, `transaction_range`, `effective_date/end_date` 派生列、`payload`（JSONB）、`auditTrail` | `standard_object_versions` 表 + sqlc |
| `LifecyclePolicy` | 不同对象类型的状态转换/校验策略（组织可滞后，职位需校验梯队/编制） | `internal/standardobject/policy/*.go` |
| `Link` | 层级与挂载关系，支持 parent-child、position->organization | `standard_object_links` 表、GraphQL `StandardObjectLink` |
| `EventEnvelope` | 向 outbox 写 `standard_object.created/updated/versioned/status_changed` | `pkg/eventbus` + dispatcher |
| `SchemaRegistry` | 管理 payload schema 版本、ISO 11179 Data Element Concept（DEC）语义 ID、OCL 守卫与发布策略，供后端/前端/生成脚本读取 | `standard_object_schemas` 表 + 文档化流程 |
| `Translation`/`Attachment` | 多语言 label、外部文档/批复等扩展维度 | `standard_object_translations`、`standard_object_attachments` |
| `Observability & Metrics` | 记录快照刷新、校验差异、事件消费等指标/OBS 事件 | `standard_object_metrics`、日志规范 |

#### 4.1.1 语义锚点与 OCL 验证
- Schema Registry 需与 Plan 403 提出的 ISO 11179 语义锚点保持一致：每个 payload 字段都绑定 Data Element Concept（DEC）ID、语义版本、可选同义词列表。`docs/reference/schema-registry.json` 在生成时必须输出 `{ fieldPath, decId, glossaryUrl }` 结构，QA 在 `logs/plan400/schema/*.log` 验证缺失项。
- 同一 Schema 条目携带 `oclGuard` 数组，落地 403 文档的“组合层 OCL”要求。命令服务在写入前、migration/validator/Playwright 在验收前均调用共享 `pkg/ocl` 引擎执行 `preState`/`postState` 校验，违反则返回 `422 STANDARD_OBJECT_OCL_VIOLATION`。
- Manifest/前端生成脚本使用 DEC ID 决定显示名称、默认提示与多语言描述，确保 UI/文档的唯一语义来源；任何新增字段若未在 Schema Registry 注册 DEC，将被 `scripts/quality/architecture-validator.js` 阻断。

#### 4.1.2 双时态（Bi-Temporal）基线
为满足审计与补录场景，SOM 必须像 SAP/Workday 一样同时捕获“业务生效时间（Valid Time）”与“系统事务时间（Transaction Time）”：

| 维度 | 说明 | 字段/结构 | 证据 |
|------|------|-----------|------|
| Valid Time | 现实世界事件生效区间，用户按 `effectiveDate/endDate` 查询 | 数据层使用 `validity_range tstzrange`；对外通过视图列 `effective_date`/`end_date` 暴露 | `EXCLUDE USING gist (object_id WITH =, validity_range WITH &&)`，`logs/plan400/migration/time-constraint-report.log` |
| Transaction Time | 系统得知/采用该记录的区间，回答“某日系统视图如何” | `transaction_range tstzrange`（或派生列 `transaction_from/transaction_to`）；命令服务只追加行 | `logs/plan400/audit/transaction-range-report.log`，`standard_object_audit` 视图 |

实施要求：
- `standard_object_versions` 以 append-only 模式写入：新版本插入时设置 `transaction_range = tstzrange(now(), 'infinity')`；撤销/更正通过更新上一行的 `transaction_range` 上界完成，禁止覆写历史行。
- Outbox 事件 payload 新增 `transactionTimestamp`，供快照/下游系统按事务时间回放；`cmd/tools/standardobject-snapshot-refresh` 需支持 `as_of_valid` 与 `as_of_transaction` 两种模式。
- `standardobject-validator` 必须检查双时态约束：TC1/TC2 既不允许 `validity_range` 重叠，也不允许 `transaction_range` 重叠；审计日志需保存补录/撤销事件。
- PR/验收模板需显式填入“本次对象类型的 `validity_range`/`transaction_range` 处理策略”，并附 `logs/plan400/audit/*.log` 作为证据。

#### 4.1.2 时间约束类型（SAP 对标）
对齐 SAP HR 的 Time Constraint 体系，SOM 必须在 Schema Registry 中声明 `timeConstraint ∈ {TC1, TC2, TC3}` 并在命令/迁移/校验阶段强制执行：

| 类型 | 说明 | 典型对象 | 存储与验证 |
|------|------|----------|------------|
| `TC1` | 任意时刻恰好一条记录，不允许空窗或重叠；系统需自动裁剪区间。 | 组织/职位主记录、法律实体等“历史唯一”对象 | `standard_object_versions` 触发器在写入前执行 time slicing，保证 `[effectiveDate, endDate]` 全覆盖；`standardobject-validator` 输出 coverage/overlap 报告 |
| `TC2` | 任意时刻最多一条记录，可存在空窗；禁止重叠。 | profile 扩展字段、附属属性 | 触发器/validator 阻止交叉区间，其他逻辑与 TC1 相同但不自动补齐 |
| `TC3` | 允许同一时间存在多条记录。 | 备注、附件、观察指标 | 不做区间冲突校验；查询/快照层通过排序或 Link 属性决策 |

实施要求：
- `standard_object_schemas` 增加 `time_constraint` 列，`docs/reference/schema-registry.json` 输出 `{ objectType, timeConstraint }`，并在 `docs/reference/standard-object-evidence-guide.md` 记录每种类型的验证脚本。
- 命令服务写入前调用 `pkg/temporal/constraints`：TC1 自动裁剪并合并相邻版本，TC2 仅检查重叠，TC3 直接放行；若违反规则则返回 `409 STANDARD_OBJECT_TEMPORAL_CONSTRAINT_VIOLATION`。
- Migrator/Validator 需在 `logs/plan400/migration/time-constraint-report.log` 中记录每个对象类型的空窗、重叠、被裁剪区间数量；TC1 出现任何空窗或重叠即视为 blocker。
- 快照/读模型刷新逻辑须根据 `timeConstraint` 决定策略：TC1 事件需触发即时刷新（保证 asOf 精准），TC2/TC3 可批量刷新但必须在指标中标记延迟。
- PR/验收模板需列出本次涉及的对象类型及其 `timeConstraint`，审阅者据此核对 Schema Registry、触发器与 validator 证据，避免“无约束”时间字段再次混入。

### 4.2 生命周期与状态机
状态：`DRAFT` → `READY` → `ACTIVE` → `SUSPENDED` → `RETIRED`。  
规则：
1. `READY` 仅可由 `DRAFT` 进入，需通过类型特定 `LifecyclePolicy.ValidateReady`.
2. `ACTIVE` 必须有至少一个有效版本，版本的 `effectiveDate`（或 `lower(validity_range)`）≤ 当前时间。
3. `SUSPENDED` 可返回 `ACTIVE`，需记录原因（存储在版本 payload 内 `suspensionNote`）。
4. `RETIRED` 为终态，不可再创建新版本；组织/职位 retire 时必须广播事件 `standard_object.retired`.
5. 所有状态变更经由 `ObjectCommandService`（REST）执行，保证 CQRS 原则。

### 4.3 数据模型与迁移
新增 Goose/Atlas 迁移（示例字段）：
- `standard_objects`: `id (uuid)`, `code (text unique)`, `object_type`, `object_type_potency smallint`, `object_type_inheritance text`, `tenant_code`, `display_name`, `status`, `labels jsonb`, `schema_version text`, `data_classification text`, `retention_policy text`, `created_by`, `created_at`, `updated_at`.
- `standard_object_versions`: `id`, `object_id`, `version_code`, `validity_range tstzrange`, `transaction_range tstzrange`, `effective_date generated always as (lower(validity_range)) stored`, `end_date generated always as (coalesce(upper(validity_range), 'infinity'::timestamptz)) stored`, `payload jsonb`, `is_current`, `audit jsonb`, `checksum`, `created_at`, `updated_at`.
- `standard_object_links`: `id`, `source_object_id`, `target_object_id`, `link_type`, `attributes jsonb`, `validity_range tstzrange`, `transaction_range tstzrange`, `tenant_code`, `created_by`, `created_at`, `updated_at`, `updated_by`.
- `standard_object_schemas`: `id`, `object_type`, `schema_version`, `schema_hash`, `definition jsonb`, `dec_bindings jsonb`, `ocl_guards jsonb`, `glossary_url text`, `published_at`, `rollback_version`, `maintainer`.
- `standard_object_translations`: `id`, `object_id`, `locale`, `display_name`, `description`, `labels jsonb`, `updated_at`, `updated_by`.
- `standard_object_attachments`: `id`, `object_id`, `attachment_type`, `storage_uri`, `metadata jsonb`, `uploaded_by`, `uploaded_at`.
- `standard_object_metrics`: `id`, `object_id`, `metric_type`, `metric_value numeric`, `recorded_at`, `labels jsonb`。

实现要求：
1. 采用 sqlc（Plan 201 差异项）生成仓储接口，生成命令置于 `internal/standardobject/repository/sqlc`；接口需提供 `LookupObjectTypeMetadata` 以读取 potency/继承策略、`timeConstraint` 与双时态策略。
2. `internal/organization`、`internal/position` 仅通过 `ObjectService` 访问公共仓储；保留特定字段（如组织扩展属性）通过 `payload` + schema 校验。`ObjectService` 在创建/迁移时必须校验 `object_type_potency`、`timeConstraint`、`validity_range`/`transaction_range`，并调用共享的 `pkg/temporal/constraints`。
3. 迁移脚本命名 `20251201090000_create_standard_objects.sql`，必须包含 Up/Down，并在同一批次创建 Schema Registry、Translation、Attachment、Metrics 等扩展表；`standard_object_versions`/`links` 需建立 `EXCLUDE USING gist` 约束（`(object_id, validity_range)`、`(object_id, transaction_range)` 等）和 GiST 索引，Schema Registry 表需包含 `dec_bindings`、`ocl_guards`、`time_constraint`、`glossary_url` 等列；操作规范需更新至 `docs/reference`。

### 4.4 契约与 API
1. **REST**：`POST /standard-objects/{objectType}` 创建对象，`POST /standard-objects/{objectType}/{code}/versions` 创建版本，`PATCH /standard-objects/{objectType}/{code}/status`，`GET /standard-objects/{objectType}` 列表。路径参数统一 `{objectType}/{code}`。
2. **GraphQL**：新增 `StandardObject`, `StandardObjectVersion`, `StandardObjectLink`，查询通过 `standardObject(code: ID!, objectType: StandardObjectType!): StandardObject`.
3. **事件**：outbox payload 统一结构 `{objectType, code, eventType, versionCode?, status?, occurredAt, transactionTimestamp}`，订阅方（如 workforce）通过 `pkg/eventbus` adapter 接入，可按 `transactionTimestamp` 回放系统视图。
4. 所有字段命名 camelCase，并写入 `docs/api/openapi.yaml` / `docs/api/schema.graphql` → 跑 Plan 258/259 守卫。

### 4.5 UI/UX 统一策略
1. 在 `frontend/src/features/temporal/entity` 下新增 `standardObjectAdapter.ts`，负责把 GraphQL 响应映射到 `TemporalEntityRecord`.
2. `OrganizationTemporalPage` 与 `PositionTemporalPage` 只注入对象类型、字段映射、表单 schema；页面骨架、tab、版本操作复用 `TemporalEntityLayout`。
3. 表单配置：采用 JSON Schema + 动态组件，放置 `frontend/src/shared/forms/standard-object`，便于 workforce/contract 共享。
4. Playwright：新增 `frontend/tests/e2e/standard-object-lifecycle.spec.ts`，收敛 selectors（`temporalEntitySelectors.*`），`logs/plan400/ui/*.log` 落盘 `[OBS]` 事件。
5. Manifest/Schema 生成：结合 Plan 300，在 `scripts/generate-forms-from-openapi.ts`、`scripts/generate-columns-from-graphql.ts`、`docs/reference/schema-registry.json` 输出中引入 SOM 实体，记录证据到 `logs/plan400/manifest/*.log`。

### 4.6 开发阶段（建议 4 Sprint）
| 阶段 | 时间 | 交付 | 依赖 |
|------|------|------|------|
| Sprint 1 – 元模型 & 契约 | W1-W2 | 设计 SOM schema、补充 OpenAPI/GraphQL、完成 Goose/Atlas 迁移及 sqlc 生成，`ObjectService` 接口齐备 | `docs/api/*`、Atlas、sqlc |
| Sprint 2 – 后端集成 | W3-W4 | 命令/查询服务接入 SOM，组织/职位命令迁移，outbox 事件、权限 scope(`scope:standard-object.write`) 落地 | `pkg/eventbus`、Plan 252/259 |
| Sprint 3 – 读模型（快照/闭包） | W5 | 落地 `standard_object_hierarchy_snapshots`、`*_mv` 物化视图，建立 `cmd/tools/standardobject-snapshot-refresh` Job；Outbox dispatcher 触发刷新；提供 `logs/plan400/snapshots/*.log` | Plan 401 方案、Outbox dispatcher |
| Sprint 4 – UI & 验收 | W6 | 前端 Temporal 页面接入、Playwright 场景、`make test-db`、`npm run test:e2e`、snapshot 刷新证据、迁移报告 | `frontend/src/features/temporal/*`, Plan 222 证据 |

### 4.7 读模型与快照设计

- **快照表**：新增 `standard_object_hierarchy_snapshots`（`snapshot_id`, `as_of_date`, `tenant_code`, `ancestor_code`, `descendant_code`, `path_text`, `depth`, `metadata jsonb`, `refreshed_at`）。按 `as_of_date` + `tenant_code` 分区或创建 BRIN 索引。
- **物化视图**：为常用日期（如 `CURRENT_DATE`、每月 1 日）创建 `CREATE MATERIALIZED VIEW standard_object_snapshot_mv AS ...`，并在 `cmd/tools/standardobject-snapshot-refresh` 中调用 `REFRESH MATERIALIZED VIEW CONCURRENTLY`。
- **刷新流程**：Outbox dispatcher 捕获 `standard_object.*` 事件后，将任务写入刷新队列，Go Job 或 SQL 调度器（例如 `pg_cron`）执行 `UPSERT`/`DELETE + INSERT` 更新快照；运行日志写入 `logs/plan400/snapshots/*.log`。
- **查询策略**：REST/GraphQL 读路径优先命中快照/物化视图；若缺少特定日期快照，则即时计算并触发异步刷新。所有查询需支持 `asOfValid` 与 `asOfTransaction` 两个参数，默认 `asOfTransaction = now()`，以便审计回放系统视图。

### 4.8 多视点矩阵
结合 Plan 403 的 Multi-view Projection Pattern，SOM 的契约/实现/证据必须以视点矩阵交付：

| 视点 | 关注点 | 产物/证据 |
|------|--------|-----------|
| **结构视点** | ObjectKernel/Link、Schema Registry、DEC 列表 | `docs/api/*`, `docs/reference/schema-registry.json`, `logs/plan400/schema/*.log` |
| **运行视点** | 状态机、版本事件、快照/闭包刷新、Outbox 指标 | `pkg/eventbus` 事件、`standard_object_hierarchy_snapshots`, `logs/plan400/snapshots/*.log` |
| **观察视点** | OBS 事件、Playwright 选择器守卫、性能指标 | `logs/plan400/ui/*.log`, `[OBS] standardObject.*` 事件、`standard_object_metrics` |
| **协作视点** | Manifest/Slot 注册、PBAC scope、生成器输出 | `frontend/src/features/temporal/*`, `scripts/generate-*` 证据、`scripts/quality/auth-permission-contract-validator.js` |

任何新增模块（如 contract/workforce）需声明其视点覆盖关系，并在 PR/验收模板中粘贴对应矩阵条目，避免“契约→实现→证据”跨层不一致。

### 4.9 多级建模与对象类型效力
为满足 403 文档提出的 Deep Instantiation 要求：
- `objectType` 元数据引入 `potency`（默认 1）。`potency>1` 的类型（例如“POSITION_FAMILY”）可以在 M1 层继续派生子类型，再在 M0 赋值字段。Schema Registry 需记录 `potency`、允许派生的属性以及 `inheritanceStrategy`（`FIXED`/`EXTENDABLE`）。
- `internal/standardobject/policy` 在加载类型配置时校验 `potency` 与状态机是否兼容（例如 `potency=2` 的类型在派生层同样遵循 `DRAFT→ACTIVE`）。
- Migrator/Validator 需对旧表数据生成默认的 `potency=1` 记录，并在 `logs/plan400/migration/*.log` 给出“未声明 potency 的对象类型数量”，以防扩展时遗漏模型层语义。

---

## 5. 迁移策略与回滚
1. **双写期**（可选，最长 1 Sprint）：组织/职位接口写 SOM + 旧表，读取以 SOM 为主，出现不一致立即回滚至旧表并修复迁移脚本。
2. **数据迁移脚本**：提供 `cmd/tools/standardobject-migrator`（Go），读取现有 `organization_units` / `positions` 表，写入新表并生成校验报告（计数、校验码），记录在 `logs/plan400/migration/*.log`；迁移完成后，立即重建闭包表与快照表，输出 `logs/plan400/snapshots/rebuild.log`。
3. **回滚**：保留旧仓储 1 个版本（feature toggle `STANDARD_OBJECTS_ENABLED`）。若出现严重缺陷，切回旧仓储并执行 Goose Down，同时清空新建闭包表/快照表（或保留只读副本），确保读模型不会引用失效数据；所有回滚操作记录至 `logs/plan400/rollback/*.log`。

---

## 6. 风险与缓解
| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| sqlc 引入节奏慢 | 延误 Phase3 | 在 Sprint1 先完成 sqlc pipeline（`make sqlc-generate`），由 Plan 201 差异项负责人跟进 |
| 多对象共享 payload Schema 复杂 | 影响扩展字段 | 采用 JSON Schema + `schema_version` 字段；在生成器中校验 schema hash |
| 权限 scope 不一致 | 破坏 PBAC | 在 Plan 252/259 的数据源中新增 `standardObject.*` 条目，命令/查询服务在注入策略前做校验 |
| UI 适配成本高 | 前端进度受阻 | 复用 `TemporalEntity` 规范，并提供 `standardObjectAdapter`，只允许在 adapter 中做对象特定逻辑 |

---

## 7. 依赖与资源
- **人力**：后端 3（shared + organization + position）、前端 2、QA 1、架构/DBA 0.5、文档 0.5。
- **工具**：Go ≥1.24.9、Node ≥18、sqlc、Atlas、Docker Compose、Playwright 1.56。
- **依赖计划**：Plan 203（模块划分）、Plan 215（基础设施）、Plan 220（模块模板）、Plan 252/255/259（门禁）。

---

## 8. 输出与证据
1. `docs/api/openapi.yaml` / `docs/api/schema.graphql` 中新增的 Standard Object 契约变更。
2. `database/migrations/20251201090000_create_standard_objects.sql` + `configs/sqlc/sqlc.yaml` 更新 + 生成代码 diff。
3. `internal/standardobject/**` 模块与组织/职位调用示例（402A 阶段提供 `adapter/noop` + Feature Flag skeleton，后续阶段再替换实现）。
4. `docs/reference/schema-registry.json` 中 objectType 映射的 DEC/OCL 绑定与 `logs/plan400/schema/*`、`logs/plan402/mapping/*` 的证据。
5. `frontend` 的 adapter、表单配置、Playwright 日志。
6. `logs/plan400/`：迁移脚本、`make test-db`, `npm run test:e2e`, `scripts/quality/*` 运行截图或日志；`logs/plan400/snapshots/*.log` 记录闭包/快照刷新与验证。

---

## 9. 后续工作
- Phase4 contract 模块接入 SOM（单独开 Plan 4xx 子文档）。
- 将 SOM 纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，提供统一入口。
- 评估将 SOM 事件写入数据仓库或审计总线的需求（与 `pkg/eventbus` Redis Adapter 计划同步）。

---

## 10. 与 Plan 200/201 的对齐要求

为确保本方案完全继承 200/201 的工程基线，以下守卫为强制项：

1. **端口/适配器模式**：`internal/standardobject/api.go` 暴露的接口需要在 `cmd/hrms-server/*` 入口统一注入，禁止服务层直接引用仓储实现。所有依赖关系必须通过 Go 原生的端口/适配器结构与 `internal/.../adapter` 目录实现。
2. **sqlc + Atlas Pipeline**：仓库根需保留 `make sqlc-generate`、`atlas migrate diff` 守卫；在 PR/验收阶段必须提供 `logs/plan400/sqlc-generate.log`、`logs/plan400/atlas-diff.log`，证明生成产物与迁移脚本为最新版本。
3. **Docker-first 测试**：所有数据库/集成测试须运行 `make test-db`（调用 `scripts/run-integration-tests.sh`，在 Docker Compose 中执行 Goose Up → Go Integration Tests → Goose Down），并在 `logs/plan400/test-db/*.log` 落证；禁止依赖宿主 PostgreSQL。
4. **事务性发件箱与持久化 EventBus**：Outbox dispatcher 必须注入 Plan 201 规划的持久化传输（Redis/Faktory/Asynq adapter），并在 `pkg/eventbus` 层提供指标；`STANDARD_OBJECTS` 相关事件不得使用内存总线。
5. **PBAC 外部化**：Plan 252/259 的策略源需纳入 Schema Registry/对象模型，命令/查询服务不得硬编码 scope；验收时需提供 `scripts/quality/auth-permission-contract-validator.js` 报告。
6. **质量守卫**：在 `scripts/quality/*`（Plan 255/258/259）中新增/更新规则以检查 Schema Registry、Manifest/Slot、禁止 `_from/_to` 命名等；CI 必须运行 `make lint`, `make fmt`, `npm run lint`, `npm run quality:preflight` 并附日志。

只有在满足以上守卫并提供相应日志的情况下，SOM 交付才视为与 Plan 200/201 对齐。

---

**维护人**：Plan 400 Owner（由架构组指派）。如有变更请同步 `docs/development-plans/00-README.md` 索引，并确保本文件保持唯一事实来源。
