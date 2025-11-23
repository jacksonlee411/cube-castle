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
| 迁移现有模块 | 组织/职位模块调用 SOM Port，不再直接持有仓储；旧仓储逐步迁往 `internal/standardobject/repository` 并由 sqlc 生成 | 确保无重复逻辑；保留回滚策略 |

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
| `TemporalVersion` | 版本信息：`versionCode`, `effectiveFrom`, `effectiveTo`, `payload`（JSONB）、`auditTrail` | `standard_object_versions` 表 + sqlc |
| `LifecyclePolicy` | 不同对象类型的状态转换/校验策略（组织可滞后，职位需校验梯队/编制） | `internal/standardobject/policy/*.go` |
| `Link` | 层级与挂载关系，支持 parent-child、position->organization | `standard_object_links` 表、GraphQL `StandardObjectLink` |
| `EventEnvelope` | 向 outbox 写 `standard_object.created/updated/versioned/status_changed` | `pkg/eventbus` + dispatcher |

### 4.2 生命周期与状态机
状态：`DRAFT` → `READY` → `ACTIVE` → `SUSPENDED` → `RETIRED`。  
规则：
1. `READY` 仅可由 `DRAFT` 进入，需通过类型特定 `LifecyclePolicy.ValidateReady`.
2. `ACTIVE` 必须有至少一个有效版本，版本的 `effectiveFrom` ≤ 当前时间。
3. `SUSPENDED` 可返回 `ACTIVE`，需记录原因（存储在版本 payload 内 `suspensionNote`）。
4. `RETIRED` 为终态，不可再创建新版本；组织/职位 retire 时必须广播事件 `standard_object.retired`.
5. 所有状态变更经由 `ObjectCommandService`（REST）执行，保证 CQRS 原则。

### 4.3 数据模型与迁移
新增 Goose/Atlas 迁移（示例字段）：
- `standard_objects`: `id (uuid)`, `code (text unique)`, `object_type`, `tenant_code`, `display_name`, `status`, `labels jsonb`, `created_by`, `created_at`, `updated_at`.
- `standard_object_versions`: `id`, `object_id`, `version_code`, `effective_from`, `effective_to`, `payload jsonb`, `is_current`, `audit jsonb`.
- `standard_object_links`: `id`, `source_object_id`, `target_object_id`, `link_type`, `attributes jsonb`.

实现要求：
1. 采用 sqlc（Plan 201 差异项）生成仓储接口，生成命令置于 `internal/standardobject/repository/sqlc`.
2. `internal/organization`、`internal/position` 仅通过 `ObjectService` 访问公共仓储；保留特定字段（如组织扩展属性）通过 `payload` + schema 校验。
3. 迁移脚本命名 `20251201090000_create_standard_objects.sql`，必须包含 Up/Down。

### 4.4 契约与 API
1. **REST**：`POST /standard-objects/{objectType}` 创建对象，`POST /standard-objects/{objectType}/{code}/versions` 创建版本，`PATCH /standard-objects/{objectType}/{code}/status`，`GET /standard-objects/{objectType}` 列表。路径参数统一 `{objectType}/{code}`。
2. **GraphQL**：新增 `StandardObject`, `StandardObjectVersion`, `StandardObjectLink`，查询通过 `standardObject(code: ID!, objectType: StandardObjectType!): StandardObject`.
3. **事件**：outbox payload 统一结构 `{objectType, code, eventType, versionCode?, status?, occurredAt}`，订阅方（如 workforce）通过 `pkg/eventbus` adapter 接入。
4. 所有字段命名 camelCase，并写入 `docs/api/openapi.yaml` / `docs/api/schema.graphql` → 跑 Plan 258/259 守卫。

### 4.5 UI/UX 统一策略
1. 在 `frontend/src/features/temporal/entity` 下新增 `standardObjectAdapter.ts`，负责把 GraphQL 响应映射到 `TemporalEntityRecord`.
2. `OrganizationTemporalPage` 与 `PositionTemporalPage` 只注入对象类型、字段映射、表单 schema；页面骨架、tab、版本操作复用 `TemporalEntityLayout`。
3. 表单配置：采用 JSON Schema + 动态组件，放置 `frontend/src/shared/forms/standard-object`，便于 workforce/contract 共享。
4. Playwright：新增 `frontend/tests/e2e/standard-object-lifecycle.spec.ts`，收敛 selectors（`temporalEntitySelectors.*`），`logs/plan400/ui/*.log` 落盘 `[OBS]` 事件。

### 4.6 开发阶段（建议 3 Sprint）
| 阶段 | 时间 | 交付 | 依赖 |
|------|------|------|------|
| Sprint 1 – 元模型 & 契约 | W1-W2 | 设计 SOM schema、补充 OpenAPI/GraphQL、完成 Goose/Atlas 迁移及 sqlc 生成，`ObjectService` 接口齐备 | `docs/api/*`、Atlas、sqlc |
| Sprint 2 – 后端集成 | W3-W4 | 命令/查询服务接入 SOM，组织/职位命令迁移，outbox 事件、权限 scope(`scope:standard-object.write`) 落地 | `pkg/eventbus`、Plan 252/259 |
| Sprint 3 – UI & 验收 | W5-W6 | 前端 Temporal 页面接入、Playwright 场景、`make test-db`、`npm run test:e2e` 证据、迁移报告 | `frontend/src/features/temporal/*`, Plan 222 证据 |

---

## 5. 迁移策略与回滚
1. **双写期**（可选，最长 1 Sprint）：组织/职位接口写 SOM + 旧表，读取以 SOM 为主，出现不一致立即回滚至旧表并修复迁移脚本。
2. **数据迁移脚本**：提供 `cmd/tools/standardobject-migrator`（Go），读取现有 `organization_units` / `positions` 表，写入新表并生成校验报告（计数、校验码），记录在 `logs/plan400/migration/*.log`。
3. **回滚**：保留旧仓储 1 个版本（feature toggle `STANDARD_OBJECTS_ENABLED`）。若出现严重缺陷，切回旧仓储并执行 Goose Down，保留审计日志。

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
2. `database/migrations/20251201090000_create_standard_objects.sql` + `sqlc.yaml` 更新 + 生成代码 diff。
3. `internal/standardobject/**` 模块与组织/职位调用示例。
4. `frontend` 的 adapter、表单配置、Playwright 日志。
5. `logs/plan400/`：迁移脚本、`make test-db`, `npm run test:e2e`, `scripts/quality/*` 运行截图或日志。

---

## 9. 后续工作
- Phase4 contract 模块接入 SOM（单独开 Plan 4xx 子文档）。
- 将 SOM 纳入 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，提供统一入口。
- 评估将 SOM 事件写入数据仓库或审计总线的需求（与 `pkg/eventbus` Redis Adapter 计划同步）。

---

**维护人**：Plan 400 Owner（由架构组指派）。如有变更请同步 `docs/development-plans/00-README.md` 索引，并确保本文件保持唯一事实来源。
