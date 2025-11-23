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
| `effective_date` | `standard_object_versions.effective_from` | DATE 语义映射到版本的生效开始时间。 |
| `end_date` | `standard_object_versions.effective_to` | DATE 语义映射到版本的生效结束时间。 |
| `effective_from`, `effective_to`（timestamp） | `standard_object_versions.effective_from/to` | 若现有记录提供更细粒度时间，直接覆盖 DATE 版本。 |
| `change_reason`, `operation_type` | `standard_object_versions.audit` | 作为 `audit.changeReason`、`audit.operation` 保存，供版本追溯。 |
| `is_current` | `standard_object_versions.is_current` | 迁移到版本表，配合 Plan 400 状态规则确定活跃版本。 |
| `deleted_at`, `deleted_by`, `deletion_reason` | `standard_object_versions.audit` | 视为终止原因，仍属于版本层审计信息。 |
| `suspended_at`, `suspended_by`, `suspension_reason` | `standard_object_versions.audit` | 仍由版本记录某次挂起的上下文。 |
| `operated_by_id`, `operated_by_name` | `standard_object_versions.audit` | 统一进入 `auditTrail`，与 `changed_by/approved_by` 一起追溯操作者。 |
| `changed_by`, `approved_by` | `standard_object_versions.audit` | 审批链条属于版本审计。 |

> 备注：`code_path/name_path` 等派生列迁移后由 `standard_object_links` + `standard_object_hierarchy_snapshots` 重建；`organization_units` 表本身不再持久化这些冗余字段，确保唯一事实来源集中在 SOM 三表。

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
|  | `schema_version text` (可选) | payload schema 版本标记，辅助校验器 |
|  | `created_by uuid` / `created_at timestamptz` / `updated_at timestamptz` | 对象级审计字段 |
| `standard_object_versions` | `id uuid PK` | 单个版本记录 ID |
|  | `object_id uuid FK` | 指向 `standard_objects.id` |
|  | `version_code text` | 版本编号/外部 ID（可映射旧 `record_id+effective_date`） |
|  | `effective_from timestamptz` / `effective_to timestamptz` | 版本生效区间；日期粒度可在视图层转化 |
|  | `is_current boolean` | 标识活跃版本，配合生命周期校验 |
|  | `payload jsonb` | 版本化字段载体（名称、描述、profile/metadata、扩展属性等） |
|  | `audit jsonb` | 审计轨迹：`changeReason/operation/changedBy/approvedBy/operatedBy/deleted*/suspended*` 等 |
|  | `checksum text` (可选) | payload 校验码，支持迁移脚本比对 |
|  | `created_at timestamptz` / `updated_at timestamptz` | 版本行级审计（可并入 `audit`） |
| `standard_object_links` | `id uuid PK` | 链接记录 ID |
|  | `source_object_id uuid FK` / `target_object_id uuid FK` | 父子/起止对象 ID |
|  | `link_type text` | 关系类型（`ORG_HIERARCHY`、`POSITION_BELONGS_TO_ORG` 等），发行方需维护枚举表 |
|  | `attributes jsonb` | 链接附加信息（`sortOrder`、`level`、`hierarchyDepth`、`codePath/namePath` 缓存、`isPrimary`、可选 `effectiveFrom/To` 等） |
|  | `tenant_code text` | 保持多租隔离（可从 source 继承） |
|  | `created_by uuid` / `created_at timestamptz` / `updated_at timestamptz` / `updated_by uuid` | 链路级审计 |

> 索引/约束：`standard_object_versions` 至少需要 `(object_id, version_code)` 与 `(object_id, effective_from)` 唯一索引；`is_current` 过滤索引用于快查；`standard_object_links` 需 `UNIQUE (source_object_id, target_object_id, link_type)` 并启用外键级联。若业务需要携带 schema hash/outbox 关联，可在 Phase B 追加列，但此表为 Plan 400/402 的最小字段集合。

---

## 5. 执行计划（建议 3 阶段）

### 阶段 A · 契约与映射（1 Sprint）
- 任务
  1. 编写《Standard Object 映射规格》：列出 `organization_units` → `standard_objects`/`standard_object_versions`/`standard_object_links` 字段映射、触发器替代方案、payload schema 设计。
  2. 补充 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 中与 SOM 相关的实体/枚举，明确状态机、版本/链接字段（与 Plan 400 同步）。
  3. 设计临时兼容视图（如 `vw_organization_units`) 及 Port 适配，确保迁移期命令/查询代码可并行验证。
- 交付：映射规格、API 契约 diff、视图/Port 设计草案。

### 阶段 B · 数据层改造（1-2 Sprint）
- 任务
  1. 实施 `standard_objects*` 迁移（`atlas`/`goose`），生成 sqlc 仓储，放置于 `internal/standardobject/repository/sqlc`。
  2. 在命令/查询服务中注入 `ObjectService`，组织/职位模块通过 Port 操作 SOM，保留 feature flag `STANDARD_OBJECTS_ENABLED` 以支撑回滚。
  3. 构建双写/校验脚本：`cmd/tools/standardobject-migrator`（一次性导入）与 `cmd/tools/standardobject-validator`（比对记录与层级）。
- 交付：迁移 SQL、sqlc 代码、Go/TS 适配、双写脚本与运行日志（`logs/plan402/migration/*.log`）。

### 阶段 C · 切换与收敛（1 Sprint）
- 任务
  1. 关闭旧仓储写路径，只读 `organization_units` 视图；验证 outbox 事件、快照/闭包刷新均引用 SOM。
  2. 更新前端 Temporal 页面至 `standardObjectAdapter`（复用 Plan 400 设计），清理组织/职位特有的冗余表单逻辑。
  3. 执行 `make test`、`make test-db`、`npm run test`、`npm run test:e2e` 及 `scripts/quality/*`，在 `logs/plan402/` 留存证据；若出现重大缺陷，按照回滚流程重新启用旧表并 Goose Down。
- 交付：切换报告、测试证据、回滚记录（若触发）。

---

## 6. 子计划（402A 起）

为便于立项与并行推进，将 Plan 402 拆分为以下子计划；每个子计划完成后方可进入下一阶段：

| 子计划 | 范围与任务 | 关键交付物 / 依赖 |
|--------|------------|-------------------|
| **402A – SOM 映射与契约准备** | 1) 完成 `organization_units` → SOM 三表字段映射（含本文 4.1/4.2 内容落地为《Standard Object 映射规格》）；2) 设计临时兼容视图/Port（`vw_organization_units` + `internal/standardobject/adapter` 草案）；3) 在 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 中补齐 Standard Object 实体、枚举与权限 scopes。 | 映射规格文档、API diff、Port/视图设计。依赖：Plan 400 契约、Plan 201 sqlc 规范。 |
| **402B – SOM Schema & 工具链** | 1) 通过 Atlas/Goose 创建 `standard_objects*` 迁移（含 Up/Down）；2) 更新 `sqlc.yaml` 并生成 `internal/standardobject/repository/sqlc`；3) 实作 `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator` 验证脚本；4) 建立 `logs/plan402/migration/*.log` 证据规范。 | 数据库迁移脚本、sqlc 生成物、迁移/校验工具。依赖：402A 文档输出、Atlas/Goose 工具、Plan 215/255 Guards。 |
| **402C – 服务接入与双写** | 1) 命令/查询服务注入 `ObjectService` Port，组织/职位仓储切换至 SOM，并保留 `STANDARD_OBJECTS_ENABLED` feature flag；2) 构建双写链路（旧表 + SOM）与自动校验；3) 补齐 outbox 事件、快照/闭包刷新逻辑对接新表。 | Go/TS 服务改造 diff、feature flag 配置、双写/校验运行日志。依赖：402B schema、Plan 401 快照方案、Plan 252/259 权限守卫。 |
| **402D – 切换与回收** | 1) 执行一次性迁移（关闭旧表写入，视图只读）；2) 更新前端 Temporal 页面接入 `standardObjectAdapter` 并完成 Playwright 证据；3) 完整跑门禁（`make test*`、`npm run test*`、`scripts/quality/*`）并在 `logs/plan402/`、`reports/` 存证；4) 定义 Goose Down / 数据回滚脚本并实测。 | 切换报告、回滚脚本、前端适配 diff、测试日志。依赖：402C 双写稳定、Plan 222/Plan 255 门禁。 |
| **402E – 收敛与治理**（可选） | 1) 清理 `organization_units` 相关代码/触发器，仅保留兼容视图或最终删除；2) 将 SOM 纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/development-plans/400/401` 的唯一事实来源；3) 监控/指标/OBS 事件对接；4) 规划未来模块（positions 之后的 payroll/workforce）如何直接消费 SOM。 | 清理脚本、文档更新、OBS/监控布置。依赖：402D 验收通过、Plan 400/401 更新、Plan 215 文档治理。 |

> 行动顺序：402A（立项确认）→ 402B（schema/tooling）→ 402C（接入双写）→ 402D（切换验收）→ 402E（善后治理）。在发布日期前，任一子计划未完成不得进入下一子计划，以确保“资源唯一性”。

---

### 子计划 402A · SOM 映射与契约准备

状态：拟立项（需调研佐证）  
依赖：Plan 400 契约、Plan 201 sqlc 规范、`docs/api/*` 现有定义  
覆盖范围：组织/职位命令与查询模块、Docs/API 契约、临时 Port/视图设计

**目标**
1. 形成《Standard Object 映射规格》，把本节 4.1/4.2 的字段映射扩展为可执行规范，明确字段来源、转换规则、回滚策略。
2. 在 API 契约层引入 Standard Object 实体、枚举与权限 scope（REST + GraphQL），确保“先契约后实现”。
3. 设计临时兼容层（视图 + Port Adapter），使 `internal/organization`/`internal/position` 能在不立即改写仓储的情况下读取 SOM 结构，为双写做准备。

**工作项**
- **A1 · 文档固化**：  
  - 将本计划 4.1/4.2 内容转化为《Standard Object 映射规格》，含字段映射、默认值、触发器替代方案。  
  - 输出迁移矩阵（`organization_units` → SOM 三表），列出每列处理方式、验证 SQL、潜在风险。  
- **A2 · 契约补全**：  
  - 更新 `docs/api/openapi.yaml`，新增 `/standard-objects/{objectType}` REST 端点、状态机、scope（参考 Plan 400 §4.4）。  
  - 更新 `docs/api/schema.graphql`，加入 `StandardObject`, `StandardObjectVersion`, `StandardObjectLink` 类型、输入输出结构。  
  - 同步 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`/Plan 252/259 索引，声明新的权限守卫来源。  
- **A3 · 兼容设计**：  
  - 设计 `vw_organization_units`/`vw_position_units` 视图，用 SOM 三表模拟旧表结构，说明列映射、索引策略。  
  - 起草 `internal/standardobject/adapter` 接口定义（命令/查询所需 Port、DTO、错误模型），并描述与现有仓储的注入方式。  
  - 定义 Feature Flag 与回滚策略文档，为 402C 双写做铺垫。

**交付物**
- 《Standard Object 映射规格》（含 SQL/脚本示例、校验 checklist）。
- OpenAPI & GraphQL diff（带守卫脚本输出）以及权限 scope 更新说明。
- 视图/Port 设计草案（UML 或代码 skeleton）、Feature Flag/回滚策略文档。

**验收标准 & 证据**
- 映射规格通过架构组评审，登记于 `docs/development-plans/00-README.md`，并在 `logs/plan402/mapping/*.log` 留存评审记录。
- OpenAPI/GraphQL 变更通过 `scripts/quality/architecture-validator.js`、`node scripts/quality/contract-checker.js`，无命名/字段偏差。
- 视图/Port 设计在 `internal/standardobject` skeleton 中可编译（编译守卫通过），并附 `make sqlc-generate` 预检日志。

**风险**
- 契约变更影响现有客户端 → 需与 Plan 222 前端负责人同步时间表，提前暴露 schema diff。
- 视图方案可能导致性能回退 → 需在 A3 输出基准 SQL 及 explain plan，对大 tenant（≥10k 组织）给出估算。

完成 402A 后方可启动 402B（迁移与工具链），否则禁止对数据库 schema/仓储进行 SOM 相关改动。

---

### 子计划 402B · SOM Schema 与工具链

状态：待启动（依赖 402A 输出）  
依赖：402A 映射规格 & 契约、Atlas/Goose 工具链、Plan 215/255 门禁  
覆盖范围：数据库迁移、sqlc 配置、迁移/校验工具、日志规范

**目标**
1. 在数据库层创建 `standard_objects`/`standard_object_versions`/`standard_object_links` 三表（含索引、约束、触发器），并提供 Up/Down 迁移脚本，符合 Plan 400 4.3 要求。
2. 更新 `sqlc.yaml` 与相关包结构，生成 `internal/standardobject/repository/sqlc`，为后续服务接入提供类型安全的仓储接口。
3. 交付一次性迁移工具（migrator）与对账工具（validator），用于从 `organization_units` 等旧表导入 SOM，并验证数据一致性，同时制定日志/证据规范。

**工作项**
- **B1 · 迁移脚本**：  
  - 使用 Atlas/Goose 创建 `20251201090000_create_standard_objects.sql`（Up/Down），包含三表、索引、约束、触发器（如 `is_current` 自动控制、链接级联）。  
  - 在同批次脚本中新增 `standard_object_hierarchy_snapshots` 等后续依赖的骨架（可留空，但需注释）。  
  - 更新 `database/migrations/README.md` 记录，并在 `logs/plan402/migration/*.log` 保存 `atlas diff`/`goose up` 运行日志。  
- **B2 · sqlc & 包结构**：  
  - 调整 `sqlc.yaml`，新增 SOM schema 相关查询（CRUD、列表、链接维护），生成文件放置 `internal/standardobject/repository/sqlc`。  
  - 在 `internal/standardobject/domain` 目录创建实体/DTO 定义，并提供 `ObjectRepository` 接口供后续服务依赖。  
  - 更新 `make sqlc-generate` pipeline，确保 CI/本地均可生成；记录运行日志。  
- **B3 · 迁移/校验工具**：  
  - `cmd/tools/standardobject-migrator`：读取旧 `organization_units`（以及 positions 相关表），写入新表；支持 dry-run、并行批处理、失败回滚；日志落盘 `logs/plan402/migration/migrator-*.log`。  
  - `cmd/tools/standardobject-validator`：对比旧表与 SOM 数据（计数、哈希、层级一致性），输出 JSON 报告至 `logs/plan402/validator/*.json`。  
  - 定义工具使用手册和回滚流程（若 migrator 失败如何清理数据等）。  

**交付物**
- `database/migrations/20251201090000_create_standard_objects.sql`（含 Down）、Atlas diff 记录。
- 更新后的 `sqlc.yaml`、生成的 Go 代码、`internal/standardobject` 包 Skeleton。
- migrator/validator 源码、编译产物、使用说明；`logs/plan402/migration/*.log` / `logs/plan402/validator/*.json` 证据样例。

**验收标准**
- `make db-migrate-all`、`make db-rollback-last` 针对新脚本执行通过（含 Docker Compose 环境），并在日志目录留存输出。
- `make sqlc-generate` 在本地和 CI 均可无误生成，`git status` 清洁；Plan 201 差异项脚本通过。
- migrator/validator 至少在沙盒数据集上跑通一遍，落盘报告，与映射规格一致；若发现数据差异，能产出可执行的对账结论。

**风险**
- **Schema 约束冲突**：旧数据可能违反 SOM 约束（如缺少 `version_code`）。缓解：在 migrator 中提供修复策略（自动生成 versionCode、补齐缺失字段），并记录 hazard list。  
- **生成物不一致**：sqlc 版本/配置不一致导致不同开发者生成不同代码。缓解：锁定 sqlc 版本（Go toolchain/gobin）并在 `Makefile` 强制校验。  
- **工具运行性能**：迁移/校验涉及大数据量，可能耗时或锁表。缓解：使用分批/分页策略，默认 read-only 模式，提供限速参数和监控（写入 `logs/plan402/migration/*.log`）。  

仅在 402B 验收通过后，方可进入 402C（接入与双写）；否则禁止服务层引用 SOM 仓储，以免出现“双事实来源”。

---

### 子计划 402C · 服务接入与双写

状态：待启动（依赖 402A/402B 输出）  
依赖：402A 映射 + 契约、402B Schema & 工具链、Plan 401 快照/闭包方案、Plan 252/259 权限守卫  
覆盖范围：组织/职位命令服务、查询服务（GraphQL）、Feature Flag、Outbox/快照刷新

**目标**
1. 让组织/职位命令与查询服务通过 `ObjectService` Port 操作 SOM，保留受控 Feature Flag (`STANDARD_OBJECTS_ENABLED`) 和回滚策略。
2. 构建双写机制：写入操作同时更新 SOM 与旧表（或旧表只读、根据策略决定），并提供自动校验与告警，确保数据一致。
3. 适配 outbox 事件、闭包/快照刷新、查询缓存，使读取路径能在双写期间优先消费 SOM（或通过视图兼容）。

**工作项**
- **C1 · Port 接入**：  
  - 在命令服务（`cmd/hrms-server/command/internal/organization` 等）注入 `ObjectCommandService`，使用 402B 的 sqlc 仓储；保留旧仓储以便回滚。  
  - 查询服务（GraphQL）通过 `internal/standardobject/query` 读取 SOM（或视图），补齐 DTO/Resolver 层。  
  - `STANDARD_OBJECTS_ENABLED` Feature Flag：通过环境变量/配置文件控制，明确默认值与回滚步骤。  
- **C2 · 双写与校验**：  
  - 写路径：当 Flag 开启时，命令服务写 SOM + 旧表（或通过视图写 SOM），并记录双写日志（`logs/plan402/doublewrite/*.log`）。  
  - 读路径：优先读 SOM；若 Flag 关闭则回退旧表。  
  - 校验：集成 `standardobject-validator` 或新增定时任务，比对双写结果，异常时触发告警并自动切回旧表。  
- **C3 · Outbox/快照接入**：  
  - outbox dispatcher 写入 `standard_object.*` 事件（Plan 400 4.4），并更新消费方（快照刷新、下游模块）。  
  - 快照/闭包工具（Plan 401）在 SOM 写操作后触发刷新；保留日志 `logs/plan402/snapshots/*.log`。  
  - 更新缓存/GraphQL 数据加载器，确保 `standard_object_links` 能驱动层级查询。  

**交付物**
- 命令/查询服务改造 diff（Go/TS），含 Feature Flag 配置、依赖注入代码、Port/Adapter 实现。
- 双写/校验运行日志、回滚脚本说明（如何关闭 Flag、清理未完成数据）。
- Outbox 事件/快照刷新改造文档与证据（日志、Playbook）。

**验收标准**
- 在 `STANDARD_OBJECTS_ENABLED=true/false` 两种模式下，`make test`、`make test-db`、`npm run test` 全量通过；`logs/plan402/doublewrite/*.log` 记录双写成功/失败情况。
- 双写期间 `standardobject-validator` 无差异（或差异清单在允许范围并有回滚策略）；异常触发时可在 5 分钟内自动切回旧表。
- Outbox/快照链路跑通，Plan 401/Plan 250 相关脚本通过（GraphQL/REST 读 SOM 符合契约，快照刷新日志齐备）。

**风险**
- **双写一致性**：复杂事务可能导致 SOM 与旧表不一致。缓解：采用事务内双写 + 失败回滚、校验脚本实时监控；必要时短期只读旧表。  
- **性能退化**：双写/校验增加负载。缓解：引入批量写/异步校验选项，监控关键指标（Postgres TPS、延迟）。  
- **Flag 管理**：配置不当导致环境混乱。缓解：记录 Flag 状态于 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，并在 `logs/plan402/flag-history.log` 记录切换。

402C 验收通过后，才允许进入 402D 的切换与回收；未通过前禁止关闭旧表写路径。

---

### 子计划 402D · 切换与回收

状态：待启动（依赖 402A~402C 全部验收通过）  
依赖：402C 双写稳定、Plan 222/Plan 255 门禁、Plan 401 快照刷新、Plan 215 文档治理  
覆盖范围：一次性数据迁移、切换流程、前端适配、测试与回滚

**目标**
1. 完成 SOM 正式切换：停止旧表写入、以视图或镜像保障只读，所有服务改为使用 SOM 数据。
2. 更新前端 Temporal/管理界面接入 `standardObjectAdapter`，并在 Playwright/Obs 证据中证明功能完好。
3. 跑通全套门禁与回滚演练，确保出现问题时可快速恢复旧实现。

**工作项**
- **D1 · 切换执行**：  
  - 执行 `standardobject-migrator` 在生产级数据上完成最终导入，并运行 validator 确认无差异；日志存入 `logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`。  
  - 关闭旧仓储写路径（Feature Flag 默认 `true`，旧表改为视图只读或直接冻结），同时更新监控/告警规则。  
  - 输出切换 Runbook（步骤、负责人、回滚条件），并签字确认。  
- **D2 · 前端与查询适配**：  
  - 前端（组织/职位相关页面）接入 `standardObjectAdapter`，删除组织特有的冗余字段；完成 `npm run test`、`npm run test:e2e`，Playwright 证据落入 `logs/plan402/ui/*.log`。  
  - GraphQL/REST 查询全面指向 SOM（视图或直接查询），并更新缓存/selector/OBS 事件。  
- **D3 · 门禁 & 回滚**：  
  - 跑通 `make test`、`make test-db`、`scripts/quality/*`、Plan 255/Plan 252/259 守卫，并记录在 `logs/plan402/verification/*.log`。  
  - 编写 Goose Down + 数据回滚脚本（恢复旧表写入、清理 SOM 记录），并在 staging 环境实测；在 `logs/plan402/rollback/*.log` 记录演练结果。  
  - 交付切换报告（包含时间线、负责人、指标、风险、回滚状态），存入 `docs/development-plans/402-switch-report.md` 或归档目录。  

**交付物**
- 切换 Runbook + 执行报告、迁移/验证日志、监控指标截图。
- 前端适配 diff（含 adapter/selector 改动）、Playwright/性能日志。
- 回滚脚本、演练记录、门禁运行日志。

**验收标准**
- 切换完成后所有写操作仅落在 SOM，`organization_units` 表保持只读或归档，监控显示 0 双写差异。
- 前端 UI 与 GraphQL/REST 功能在 SOM 模式下通过全量测试（含 Playwright 关键场景）。
- 回滚演练通过：能够在 30 分钟内恢复旧表写入并保证数据一致；演练日志和脚本可复用。

**风险**
- **切换失败**：大规模数据迁移期间出现错误。缓解：严格遵循 Runbook，设置保守回滚点，提前在 staging 压测。  
- **前端回归**：适配过程中漏改字段。缓解：严格跑 `npm run test:e2e`、`make frontend-dev` 守卫，并由 Plan 222 前端负责人复核。  
- **监控缺失**：切换后没有及时发现异常。缓解：在 D1 前就配置 SOM 专用指标/告警（写入失败、差异、快照延迟）。

402D 完成后，进入 402E 做善后治理（清理旧表/文档/OBS 等）。

---

### 子计划 402E · 收敛与治理（可选/善后）

状态：待启动（依赖 402D 完成并稳定运行至少 1 个迭代）  
依赖：402D 切换报告、Plan 215 文档治理、Plan 400/401 更新、Plan 272 运行产物治理  
覆盖范围：旧表/触发器清理、文档/监控同步、OBS/指标完善、未来模块接入指南

**目标**
1. 清理 `organization_units` 等旧实现遗留（表、触发器、仓储、文档引用），只保留必要视图或归档。
2. 将 SOM 成果纳入全局文档（Plan 400/401、开发者速查、API 契约索引），确保唯一事实来源。
3. 完善监控/OBS 体系，沉淀未来模块接入指南，使 SOM 成为默认模式。

**工作项**
- **E1 · 数据与代码清理**：  
  - 删除/归档旧表（如 `organization_units_backup_temporal`）、触发器、迁移脚本；若需保留视图，明确只读属性。  
  - 清理 `internal/organization` 中不再使用的仓储/DTO/触发器逻辑；更新 `standardobject` Port 作为唯一入口。  
  - 在 `database/migrations` 中提供清理脚本（例如 `20251215090000_drop_legacy_org_tables.sql`），并记录日志。  
- **E2 · 文档与治理**：  
  - 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、Plan 400/401 章节，注明 SOM 为唯一事实来源，旧表已退场。  
  - 在 `docs/development-plans/00-README.md`、`docs/reference/temporal-entity-experience-guide.md` 登记新的参考链接。  
  - 将 402 计划归档到 `docs/archive/development-plans/`，保留验收标准与证据索引。  
- **E3 · 监控/OBS/未来指南**：  
  - 为 SOM 相关操作（版本创建、链接更新）配置指标/告警、OBS 事件规范；将日志路径纳入 Plan 272 的运行产物治理。  
  - 编写《SOM 接入指南》，指导 payroll/workforce 等未来模块如何复用 `standardobject` Port、迁移策略、测试要求。  
  - 在 `scripts/quality/` 追加守卫（如验证旧表未再被引用、`standardobject` schema hash 检查）。  

**交付物**
- 数据/代码清理脚本及执行日志、仓储清理 MR。
- 更新后的参考文档、归档版本（含验收标准、证据索引）。
- 监控/OBS 配置、SOM 接入指南、质量守卫脚本。

**验收标准**
- 仓库中不再存在直接引用 `organization_units` 的生产代码（除兼容视图），`rg organization_units` 仅出现在归档或视图文件中。
- 文档索引与开发者速查全部指向 SOM；Plan 400/401/402 归档更新完毕。
- 监控/OBS 指标上线，Plan 272 运行产物治理脚本记录最新产物；质量守卫可以阻止回退到旧表实现。

**风险**
- **遗留引用遗漏**：某些脚本/工具仍依赖旧表。缓解：在 `scripts/quality/architecture-validator.js` 中新增检查；运行 `rg`/CI 守卫确认。  
- **文档不同步**：多个计划/参考文件引用旧实现。缓解：在 E2 列出文档清单，使用 `scripts/document-sync` 验证。  
- **监控空白**：若不及时建立 SOM 指标，后续问题难以排查。缓解：在 Plan 272 指标治理中新增 SOM 专项任务。  

402E 完成后，Plan 402 进入归档阶段，SOM 成为组织/职位以及后续模块的唯一事实来源。

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
4. `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator`、闭包/快照 refresh 日志。  
5. `frontend` adapter、表单配置与 Playwright 证据。  
6. `logs/plan402/*`：迁移/校验/测试/回滚完整链路。

---

## 8. 后续工作

- 在 Plan 400/401 中同步引用 402 的迁移结论，保持索引唯一性。  
- 计划 Phase4（如 contract/workforce）直接消费 SOM Port，禁止新增单表实现。  
- 将本计划纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 的数据库章节，提醒开发者优先查阅 Plan 402 以避免重复造轮子。

---

**维护人**：Plan 402 Owner（同 Plan 400 Owner，若调整需更新索引）。如有变更，务必同步 `docs/development-plans/00-README.md` 并在归档时迁移至 `docs/archive/development-plans/`。
