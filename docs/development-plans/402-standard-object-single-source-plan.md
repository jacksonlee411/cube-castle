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
