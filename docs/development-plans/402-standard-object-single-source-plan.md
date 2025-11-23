# 402 · Standard Object 单一事实来源整改计划

状态：调研确认（需立项）  
最后更新：2025-11-26  
适用范围：Core HR（organization/position）、shared CQRS 基座、数据库迁移  
事实来源：`docs/development-plans/400-standard-object-model-plan.md`、`database/migrations/20251106000000_base_schema.sql`、`internal/organization/**`

---

## 1. 背景与动机

- Plan 400 要求以 `standard_objects`/`standard_object_versions`/`standard_object_links` 为 SSoT，统一组织、职位等对象的生命周期与版本（见 `docs/development-plans/400-standard-object-model-plan.md:60-107`）。  
- 现状：组织模块仍依赖单表 `organization_units` 存储主记录与时间版本（`database/migrations/20251106000000_base_schema.sql:762-808`），仓储/调度器/位置模块直接引用该表字段（例如 `internal/organization/repository/organization_create.go:20-112`、`internal/organization/repository/postgres_positions.go:430-487`）。  
- 风险：若在此基础上“复用单表”去满足 Plan 400，将出现两个事实来源（旧表 + SOM 契约）以及跨模块耦合，违反“资源唯一性与跨层一致性”。同时，单表无法提供 Plan 400 指定的版本实体、链接表与 sqlc 生成接口，CQRS/事件链路无法按目标落地。

---

## 2. 目标与验收

| 目标 | 验收方式 |
|------|----------|
| 明确单表模式的不可行性并锁定多表 SOM 作为唯一方向 | 发布《单表退场评估》附录，描述字段映射与缺失点；经架构组评审备案 |
| 制定 `organization_units` → SOM 三表的迁移蓝图（含过渡期视图/Port 适配） | 交付迁移设计文档 + sqlc schema 变更路线；在 `docs/development-plans/00-README.md` 登记 |
| 定义执行阶段（接口层改造、迁移脚本、服务接入、回滚） | 输出阶段性清单 + 验收脚本，覆盖命令、查询、前端及 outbox 事件 |

验收标准：① 所有文档引用保持单一事实来源；② 迁移策略提供回滚与证据目录；③ 不新增宿主服务依赖，全部通过 Docker Compose 运行。

---

## 3. 范围与非目标

**范围**
1. 组织与职位命令/查询服务所依赖的数据层、Domain Port、事件与前端接口。
2. PostgreSQL schema、sqlc 生成文件、Go/TS 代码中关于组织/职位对象状态与版本的实现。
3. 迁移工具、快照/闭包构建脚本，以及 `logs/plan400/`/`logs/plan402/` 证据规范。

**非目标**
- 不扩展至 Payroll/Compensation 等未纳入 Plan 400 的领域。
- 不引入第二数据库或服务；保持单一 PostgreSQL。
- 不重写 Playwright/OBS 规范，仅确保新的 SSoT 能被现有门禁消费。

---

## 4. 调研结论（单表模式的结构性缺陷）

| 缺陷 | 说明 & 证据 | 影响 |
|------|-------------|------|
| 对象/版本耦合 | `organization_units` 同时保存对象主信息与 `effective_date/end_date/is_current`（`database/migrations/20251106000000_base_schema.sql:762-808`），无法像 Plan 400 要求那样以版本实体驱动状态机和 outbox payload | `StandardObjectVersion` 无法落地；REST/GraphQL 难以返回版本列表 |
| 无统一链接表 | 组织层级依赖 `parent_code` + 触发器，职位通过 FK 引用组织（`internal/organization/repository/postgres_positions.go:430-487`），缺少 `standard_object_links` 承载跨类型关系 | 无法复用 link 机制支持 position→organization、未来 modules；违背 Plan 400 对 `Link` 模块的定义 |
| 无法提供统一 Port | 当前仓储直接操作 `organization_units`，并在应用层拼接字段；Plan 400 期望 `internal/standardobject/api.go` 作为注入入口 | 模块无法通过标准接口调用；CQRS 边界继续被表结构绑定 |
| 迁移不可控 | 如保留单表，将同时存在 `organization_units` 与未来的 `standard_objects*`，出现双事实来源；若试图只保留单表又难以支持 `payload JSONB + schema_version` | 破坏资源唯一性；影响 sqlc/Atlas 生成链路，无法通过 Plan 201 差异项 |

结论：单表模式无法满足 Plan 400 的统一模型、接口与治理要求，必须以三表 SOM 为唯一实现路径。

### 4.1 `organization_units` → SOM 字段映射

结合 Plan 400 对三张表职责的定义（`docs/development-plans/400-standard-object-model-plan.md:68-104`），全部字段的迁移归属如下，供 Phase A 的《Standard Object 映射规格》落档：

| `organization_units` 字段 | 目标表 | 说明 |
|---------------------------|--------|------|
| `record_id` | `standard_object_versions.id` | 现有主键等同于单个版本记录，迁移后直接映射为版本 ID。 |
| `tenant_id` | `standard_objects.tenant_code` | 租户维度在对象层保持恒定；版本通过对象 FK 获取。 |
| `code` | `standard_objects.code` | 业务唯一键；对象层索引供 REST/GraphQL 查询（Plan 400 4.4）。 |
| `parent_code` | `standard_object_links` | 组织层级由父子 link 承载，`link_type=ORG_HIERARCHY`，并通过对象 ID 关联。 |
| `name`, `description` | `standard_object_versions.payload` | 名称/描述属于版本化字段；如需快速检索，可在对象表维护冗余 `display_name`（不作为事实来源）。 |
| `unit_type` | `standard_objects.labels.unitType` | 对象级分类标签，生命周期内稳定；移动后作为标签参与查询。 |
| `status` | `standard_objects.status` | 生命周期状态由对象层管理，版本表只记录某版本是否当前。 |
| `level`, `hierarchy_depth` | `standard_object_links.attributes` | 这些值描述层级关系，应随着 link（或其导出的快照）保存。 |
| `code_path`, `name_path` | `standard_object_links` 衍生 | 通过 link 快照/闭包重建路径，取消在行内维护（Plan 400 4.7）。 |
| `sort_order` | `standard_object_links.attributes.sortOrder` | 父子顺序与 link 绑定。 |
| `profile`, `metadata` | `standard_object_versions.payload` | JSON 配置与扩展属性随版本变化，集中进 payload JSONB。 |
| `created_at`, `updated_at` | `standard_objects.created_at/updated_at` | 对象创建与最近对象级更新时间，Plan 400 要求存放在对象表。 |
| `effective_date` | `standard_object_versions.effective_date` | 版本的生效日期；在 SOM 中统一命名为 `effective_date`（DATE）。 |
| `end_date` | `standard_object_versions.end_date` | 版本的结束日期；与 `effective_date` 一致保持统一命名。 |
| `effective_from`, `effective_to`（timestamp） | – | 不再引入 `_from/_to` 命名，如需更细粒度的时间点，可追加独立字段而不是混用命名。 |
| `change_reason`, `operation_type` | `standard_object_versions.audit` | 作为 `audit.changeReason`、`audit.operation` 保存，供版本追溯。 |
| `is_current` | `standard_object_versions.is_current` | 迁移到版本表，配合 Plan 400 状态规则确定活跃版本。 |
| `deleted_at`, `deleted_by`, `deletion_reason` | `standard_object_versions.audit` | 视为终止原因，仍属于版本层审计信息。 |
| `suspended_at`, `suspended_by`, `suspension_reason` | `standard_object_versions.audit` | 仍由版本记录某次挂起的上下文。 |
| `operated_by_id`, `operated_by_name` | `standard_object_versions.audit` | 统一进入 `auditTrail`，与 `changed_by/approved_by` 一起追溯操作者。 |
| `changed_by`, `approved_by` | `standard_object_versions.audit` | 审批链条属于版本审计。 |

> 备注：`code_path/name_path` 等派生列迁移后由 `standard_object_links` + `standard_object_hierarchy_snapshots` 重建；`organization_units` 表本身不再持久化这些冗余字段，确保唯一事实来源集中在 SOM 三表。

> 语义锚点要求：迁移规格需为上述每个字段提供 ISO 11179 Data Element Concept（DEC）映射、来源链接与语义版本，确保 Schema Registry/生成脚本/前端提示引用同一 ID；缺失的 DEC 必须在 402A 立项阶段补齐或记录风险并给出回收期限。

#### 4.1.1 标准对象模型应具备的通用维度（与 Plan 400 对齐）
为避免“映射只关注列迁移而忽视元模型完整性”，本计划沿用 Plan 400 的标准对象元模型，将以下维度视为强制要求：

- **ObjectKernel（对象内核）**：由 `standard_objects` 承载，包含 `objectType`、`code`、`displayName`、`status`、`tenantCode`、`labels`、`schemaVersion`、`dataClassification`、`retentionPolicy`、`createdBy/createdAt/updatedAt` 等字段；所有对象级元数据、标签、权限/保留策略都应在此维度维护。
- **TemporalVersion（时态版本）**：由 `standard_object_versions` 表表示，字段至少包括 `versionCode`、`effectiveDate`、`endDate`、`payload`（JSONB）、`isCurrent`、`audit`、`checksum`、`createdAt/updatedAt`。任何业务属性或审计信息都必须放入 payload/audit，确保“变更 = 新版本”。
- **LifecyclePolicy（生命周期策略）**：对象类型需绑定状态机（`DRAFT→READY→ACTIVE→SUSPENDED→RETIRED`）和校验策略，既有字段（如 `status`、`isCurrent`、`effectiveDate`）必须能驱动状态变更；策略定义在 `internal/standardobject/policy/*`，并由命令服务注入。
- **Link（关系/层级）**：`standard_object_links` 负责父子、归属、挂载等关系，字段包含 `linkType`、`source_object_id`、`target_object_id`、`attributes`（储存 sortOrder、hierarchyDepth、path 缓存等）。任何层级/跨模块引用都需通过 Link 实现，禁止在版本表冗余 parent_code。
- **EventEnvelope（事件信封）**：命令服务写 `standard_object.*` outbox 事件，payload 至少包含 `{objectType, code, versionCode?, status?, occurredAt}`；Plan 400/402 的迁移、双写、快照刷新都以该事件为触发源。
- **读模型/快照**：由 `standard_object_hierarchy_snapshots`、`*_mv` 物化视图、`cmd/tools/standardobject-snapshot-refresh` 提供，确保查询层可以按 `asOfDate`/`tenant` 快速读取层级数据，并与事件日志、Plan 401 方案一致。
- **Schema Registry & 扩展维度**：统一维护 `standard_object_schemas`（payload schema/哈希）、`standard_object_translations`（多语言）、`standard_object_attachments`（外部文档）、`standard_object_metrics`（可观测性/刷新指标），与 Plan 300/401/272 的治理要求对齐。

以上维度是标准对象模型的“最小集合”。402 的映射、迁移、双写、前端接入都必须证明这些维度被完整覆盖，并在各阶段交付物/日志中给出证据，否则视为与 Plan 400 不一致。

### 4.2 SOM 三表字段目标清单

作为 Phase A 的合同基线，结合 Plan 400 §4.1-4.4 对 ObjectKernel/TemporalVersion/Link 的定义，与本节映射表相呼应，三张核心表应包含以下字段（如无特殊说明均为 `NOT NULL`）：

| 表 | 字段 | 说明 |
|----|------|------|
| `standard_objects` | `id uuid PK` | SOM 主键，供版本/链接引用 |
|  | `code text UNIQUE` | 业务主键，REST/GraphQL `{objectType}/{code}` 查询入口 |
|  | `object_type text` | 对象类别（如 `ORGANIZATION_UNIT`、`POSITION`） |
|  | `tenant_code text` | 多租隔离键 |
|  | `display_name text` | 当前展示名称，镜像自最新版本 payload，便于列表查询 |
|  | `status text` | 生命周期状态（Plan 400 §4.2：`DRAFT/READY/ACTIVE/SUSPENDED/RETIRED`） |
|  | `labels jsonb` | 对象级标签/分类（如 `unitType`、`region`） |
|  | `schema_version text` | payload schema 版本标记，对应 schema registry 条目 |
|  | `data_classification text` / `retention_policy text` | 数据主权/保留策略，用于审计与 PBAC 扩展 |
|  | `created_by uuid` / `created_at timestamptz` / `updated_at timestamptz` | 对象级审计字段 |
| `standard_object_versions` | `id uuid PK` | 单个版本记录 ID |
|  | `object_id uuid FK` | 指向 `standard_objects.id` |
|  | `version_code text` | 版本编号/外部 ID（可映射旧 `record_id+effective_date`） |
|  | `effective_date date` / `end_date date` | 版本生效区间（DATE）；如未来需要秒级精度，可新增扩展列但命名仍沿用 `*_date` |
|  | `is_current boolean` | 标识活跃版本，配合生命周期校验 |
|  | `payload jsonb` | 版本化字段载体（名称、描述、profile/metadata、扩展属性等） |
|  | `audit jsonb` | 审计轨迹：`changeReason/operation/changedBy/approvedBy/operatedBy/deleted*/suspended*` 等 |
|  | `checksum text` (可选) | payload 校验码，支持迁移脚本比对 |
|  | `created_at timestamptz` / `updated_at timestamptz` | 版本行级审计（可并入 `audit`） |
| `standard_object_links` | `id uuid PK` | 链接记录 ID |
|  | `source_object_id uuid FK` / `target_object_id uuid FK` | 父子/起止对象 ID |
|  | `link_type text` | 关系类型（`ORG_HIERARCHY`、`POSITION_BELONGS_TO_ORG` 等），发行方需维护枚举表 |
|  | `attributes jsonb` | 链接附加信息（`sortOrder`、`level`、`hierarchyDepth`、`codePath/namePath` 缓存、`isPrimary`、可选 `effectiveDate/endDate` 等） |
|  | `tenant_code text` | 保持多租隔离（可从 source 继承） |
|  | `created_by uuid` / `created_at timestamptz` / `updated_at timestamptz` / `updated_by uuid` | 链路级审计 |

> 索引/约束：`standard_object_versions` 至少需要 `(object_id, version_code)` 与 `(object_id, effective_date)` 唯一索引；`is_current` 过滤索引用于快查；`standard_object_links` 需 `UNIQUE (source_object_id, target_object_id, link_type)` 并启用外键级联。若业务需要携带 schema hash/outbox 关联，可在 Phase B 追加列，但此表为 Plan 400/402 的最小字段集合。

> **命名唯一性原则**：SOM 及所有过渡视图/仓储仅允许 `effective_date/end_date` 字段，严禁再次引入 `effective_from/effective_to` 或类似变体；若未来需要更精细的时间点，需通过新增扩展列并保留 `*_date` 命名。

> 补充通用维度：SOM 需同时覆盖 schema registry、翻译、附件与可观测性元数据。计划在阶段 B 建立 `standard_object_schemas`（记录 JSON Schema/哈希/发布策略）、`standard_object_translations`（`locale`, `display_name`, `description` 等多语言字段）、`standard_object_attachments`（外部文档/批复，含存储 URI、类别、合规标签）以及 `standard_object_metrics`（累计观察指标，用于快照/刷新监控）等扩展表，并通过 Feature Flag 渐进启用。

> OCL 守卫：按照 Plan 403 建议，`standard_object_schemas` 每条记录需携带 `oclGuards`（不变量/前置/后置条件）。402 计划要在 migrator、validator、命令/查询服务中复用统一 OCL 引擎，确保字段迁移与版本状态机符合组合约束，否则直接阻断写入。

### 4.3 模块化联邦与能力契约
参考 403 文档的 Modular Federation Pattern，Plan 402 在迁移期必须维持“能力契约”目录，描述每个模块提供/消费的 Standard Object 特性，并在子计划验收时复核：

| 模块 | 提供能力 | 依赖能力 | 约束与验证 |
|------|----------|----------|------------|
| Organization Federate | `ORG_HIERARCHY` link、`organization.unit` payload schema、`standard_object.*` 事件 | SOM Kernel、Link、Schema Registry | OCL：组织节点 `labels.unitType` 必须存在；DEC：`tenantCode` 与 `links.tenantCode` 相同 |
| Position Federate | `position.role` payload、`POSITION_BELONGS_TO_ORG` link | SOM Kernel、Organization Federate 事件 | OCL：职位版本必须引用 ACTIVE 的组织；DEC：`positionProfile` 引用共享 DEC |
| Shared Federate（未来 workforce/contract） | `standardObjectAdapter`、Schema Registry 读写、Manifest slots | Kernel/Version/Link/Events | OCL：`effectiveDate ≤ endDate`；视点：结构+运行+观察 |

能力契约文档存放 `docs/development-plans/402-capability-contracts.md`（由本计划输出），并在 402B 之后纳入 `scripts/quality/architecture-validator.js` 检查，严禁新增“单表例外”。

---

## 5. 执行计划（建议 3 阶段）

### 阶段 A · 契约与映射（1 Sprint）
- 目标与交付详见 [402A · SOM 映射与契约准备](./402A-standard-object-mapping-and-contracts.md)。该阶段聚焦字段映射、OpenAPI/GraphQL 契约与兼容视图/Port 设计，并输出 DEC/OCL hazard list 及 `logs/plan402/mapping/*` 证据。

### 阶段 B · 数据层改造（1-2 Sprint）
- 详见 [402B · SOM Schema 与工具链](./402B-som-schema-and-tooling.md)。负责创建 `standard_objects*` 迁移、sqlc 仓储、migrator/validator/snapshot-refresh 工具，以及 Schema Registry/扩展表与日志样例。

### 阶段 C · 服务接入与双写（1-2 Sprint）
- 详见 [402C · 服务接入与双写](./402C-service-integration-and-double-write.md)。主要任务包括命令/查询服务接入、Feature Flag、双写/校验、outbox/快照、Manifest/Slot 与能力契约巡检。

### 阶段 D · 切换与回收（1 Sprint）
- 详见 [402D · 切换与回收](./402D-cutover-and-recovery.md)。该阶段执行最终迁移、关闭旧表写入、完成前端/查询适配、跑门禁与回滚演练，并生成切换 Runbook。

### 阶段 E · 收敛与治理（可选）
- 详见 [402E · 收敛与治理](./402E-convergence-and-governance.md)。集中处理旧表/代码清理、文档/索引更新、监控/OBS 完善与未来模块接入指南。

---

## 6. 子计划（402A 起）

为便于立项与并行推进，将 Plan 402 拆分为以下子计划；每个子计划完成后方可进入下一阶段：

| 子计划 | 范围与任务 | 关键交付物 / 依赖 |
|--------|------------|-------------------|
| [**402A – SOM 映射与契约准备**](./402A-standard-object-mapping-and-contracts.md) | 1) 完成 `organization_units` → SOM 三表字段映射（含 DEC/OCL 记录）；2) 设计兼容视图/Port；3) 补齐 OpenAPI/GraphQL 契约与权限 scope。 | 映射规格、API diff、Port/视图设计。依赖：Plan 400 契约、Plan 201 sqlc 规范。 |
| [**402B – SOM Schema & 工具链**](./402B-som-schema-and-tooling.md) | 1) 创建 `standard_objects*` 迁移；2) 更新 `sqlc.yaml` 与仓储；3) 实作 migrator/validator/snapshot-refresh 工具；4) 建立日志规范。 | 迁移脚本、sqlc 生成物、迁移/校验工具与日志。依赖：402A、Atlas/Goose、Plan 215/255。 |
| [**402C – 服务接入与双写**](./402C-service-integration-and-double-write.md) | 1) 命令/查询服务注入 Port 并保留 Feature Flag；2) 构建双写 + 自动校验；3) 对接 outbox、快照、Manifest/Slot。 | Go/TS 改造 diff、Flag/双写/校验日志。依赖：402B、Plan 401、Plan 252/259。 |
| [**402D – 切换与回收**](./402D-cutover-and-recovery.md) | 1) 执行一次性迁移并关闭旧写入；2) 更新前端/查询；3) 跑门禁与回滚演练；4) 输出 Runbook。 | 切换报告、回滚脚本、测试日志。依赖：402C、Plan 222/255。 |
| [**402E – 收敛与治理**](./402E-convergence-and-governance.md)（可选） | 1) 清理旧表/代码；2) 更新文档索引；3) 完善监控/OBS；4) 输出 SOM 接入指南。 | 清理脚本、文档更新、OBS/监控布置。依赖：402D、Plan 400/401、Plan 215。 |

> 行动顺序：402A（立项确认）→ 402B（schema/tooling）→ 402C（接入双写）→ 402D（切换验收）→ 402E（善后治理）。在发布日期前，任一子计划未完成不得进入下一子计划，以确保“资源唯一性”。

---

详细执行、交付物与验收标准请参阅对应子计划文档（402A~402E）。本文件仅保留总体背景与输出清单，避免跨层信息重复。

---

## 6. 风险与缓解


| 风险 | 影响 | 缓解 |
|------|------|------|
| sqlc/Atlas 引入节奏过慢 | 阻塞迁移脚本生成 | 阶段 A 前置完成 `make sqlc-generate` pipeline 校验，与 Plan 201 负责人协同 |
| 双写期间数据不一致 | 影响生产准确性 | 建立校验脚本对比 `organization_units` 与 SOM 版本，异常触发 feature flag 回滚 |
| 前端适配延误 | UI/测试无法按期收敛 | 阶段 B 同步提供 adapter API，使前端只做映射，不重新实现字段逻辑 |
| 回滚路径不清晰 | 影响上线决策 | 在阶段 B 即编写 Goose Down 与数据清理脚本，实测后才允许合入 |

---

## 7. 输出与证据

1. 《Standard Object 映射规格》 + 单表退场评估（附映射表/风险）。  
2. `database/migrations/20251201090000_create_standard_objects.sql`（或后续补丁）与 sqlc 生成物。  
3. `internal/standardobject/**` Port/Repository、`internal/organization/**` / `internal/position/**` 适配 diff。  
4. `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator`、`cmd/tools/standardobject-snapshot-refresh` 及快照/闭包 refresh 日志。  
5. `frontend` adapter、Manifest/Slot 接入、契约生成的 forms/columns 与 Playwright/Preflight 证据。  
6. `logs/plan402/*`：迁移/校验/快照/事件总线/测试/回滚完整链路（含 `logs/plan402/eventbus/*.log`）。  

---

## 8. 后续工作

- 在 Plan 400/401 中同步引用 402 的迁移结论，保持索引唯一性。  
- 计划 Phase4（如 contract/workforce）直接消费 SOM Port，禁止新增单表实现。  
- 将本计划纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 的数据库章节，提醒开发者优先查阅 Plan 402 以避免重复造轮子。

---

**维护人**：Plan 402 Owner（同 Plan 400 Owner，若调整需更新索引）。如有变更，务必同步 `docs/development-plans/00-README.md` 并在归档时迁移至 `docs/archive/development-plans/`。
