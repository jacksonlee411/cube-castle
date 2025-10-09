# Cube Castle 实现清单（Implementation Inventory）

版本: v1.9.1 软删除状态迁移完成版
维护人: 架构组（与IIG护卫系统协同维护）
范围: 基于最新IIG扫描的完整实现清单（API优先+CQRS架构+状态字段统一）
最后更新: 2025-10-09（脚本刷新：26个命令端点 + 9个GraphQL查询 + 26个Go处理器 + 19个Go服务类型 + 172个前端导出项；来源 `reports/implementation-inventory.json` 快照 2025-10-09T03:39:34.022Z）

## 🔄 **重要架构变更记录**
**2025-09-27**: ✅ **软删除判定统一为仅依赖status字段**（14号计划完成）
- 数据层面：审计确认13条记录状态完全一致，零差异迁移
- 代码层面：业务逻辑已统一使用status字段，deleted_at仅保留为审计信息
- 测试验证：Go单元/集成测试、前端测试全部通过

> 目的（Purpose）
> - 中文: 统一登记当前已实现的 API、导出函数与接口，以及所属文件与简要说明，避免重复造轮子，便于新成员快速定位能力与复用。
> - EN: Centralized, bilingual catalog of implemented APIs, exported functions and interfaces with file locations and short descriptions to reduce duplication and speed onboarding.

## 🛡️ **实现清单自动更新工具** ⭐ **最重要**

### **📋 标准扫描工具：`scripts/generate-implementation-inventory.js`**

**用途**: 自动扫描项目实现状态，生成最新清单，防止重复造轮子

#### 🚀 **使用方法（本页需以脚本输出为准进行校对）**
```bash
# 查看当前实现清单（强制开发前检查）
node scripts/generate-implementation-inventory.js

# 生成最新清单（校对用）
node scripts/generate-implementation-inventory.js > temp-inventory.md
# 对比后再更新本文档；禁止凭记忆填写规模/路径/行号
```

#### 📊 **扫描能力覆盖** ⭐ **基于最新IIG扫描结果**
- ✅ **REST API端点**: 从 `docs/api/openapi.yaml` 提取 (26个端点：运维/命令/认证完整登记)
- ✅ **GraphQL查询**: 从 `docs/api/schema.graphql` 提取 (9个公开查询)
- ✅ **Go后端组件**: 扫描 handlers / services (26个导出处理器 + 19个服务类型)
- ✅ **前端TypeScript导出**: 扫描 class/function/const (172 个导出符号)

#### 🔍 **IIG护卫集成** (Implementation Inventory Guardian)
- **预开发强制检查**: 每次新功能开发前必须运行此脚本
- **重复检测防护**: 与P3.1重复代码检测系统深度集成
- **架构一致性**: 与P3.2架构验证器联动
- **功能登记验证**: 新增功能后重新扫描，确保正确登记

#### 🚨 **强制使用场景**
1. **新功能开发前**: 检查是否已有可复用的API/Hook/组件/服务
2. **API设计时**: 验证OpenAPI/GraphQL中已定义的端点
3. **代码审查前**: 确保没有重复实现相同功能
4. **文档更新时**: 保持清单与实际代码同步

#### ⚠️ **护卫原则**
- **最高优先级**: 资源唯一性与端到端一致性任何时候都高于功能交付，发现冲突立即停工处理。
- **现有资源优先**: 发现可用实现必须优先使用，禁止重复创建
- **实现唯一性**: 每个功能只能有一种实现方式
- **强制登记**: 新增功能后必须运行脚本验证登记成功
- **质量门禁**: 与企业级P3防控系统100%集成

---

## 🎯 **API优先原则与维护规则** ⭐ **项目核心原则**

### **📋 API优先开发原则 (API-First Development)**
- **💡 API优先哲学**: 先设计API契约，后实现代码逻辑 - "Contract First, Code Second"
- **📰 权威来源**: API端点与权限以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一权威
- **🔄 变更顺序**: 任何API变更必须先更新规范文档，后修改代码实现
- **🎯 设计驱动**: 基于业务需求设计API接口，避免技术实现驱动的API设计
- **📝 文档即规范**: API文档不是后补，而是开发的起点和契约

### **🏗️ 维护与收录原则（Maintaining Rules）**
- **单一来源（最高优先级）**: API 端点与权限以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一权威；此清单仅做导航索引（No divergence from spec）。
- **CQRS架构**: 查询统一 GraphQL；命令统一 REST。清单按"Query/Command"分区（Follow CQRS split）
- **命名一致**: API 层字段一律 camelCase；路径参数 `{code}`（Naming consistency per CLAUDE.md）
- **API优先验证**: 新增端点前必须先更新API契约，通过契约测试后再实现代码
- **粒度控制**: 收录"对外可复用/可调用"的导出符号（exported/public）；内部私有函数不在本表（Public symbols only）
- **更新时机**: 每次合并涉及新端点/导出函数，需同步更新本清单（Update on merge）
- **诚实与唯一性**: 规模/性能数据与路径引用必须可复现（脚本/报告为证）；同一能力只在一个权威位置登记，避免重复与冲突。

---

## 目录（Index）
- REST 命令 API（Command, OpenAPI）
- GraphQL 查询 API（Query, Schema）
- 后端（Go）关键处理器/服务/中间件（Handlers/Services/Middleware）
- 前端（TypeScript/React）API 客户端、Hooks、主要组件
- 运维与脚本（DevOps/Scripts）

---

## REST 命令 API（Command Service, Port 9090）
权威规范: `docs/api/openapi.yaml`

> 说明: 基于实际代码扫描的端点清单，与 OpenAPI 规范保持一致

### 🎯 **API优先设计端点** (26个端点，按类别汇总)

> **数据来源**: `node scripts/generate-implementation-inventory.js` 自动扫描的 OpenAPI v2025-10-09（快照 2025-10-09T03:39:34.022Z），详见 `reports/implementation-inventory.json.openapiPaths`

#### 运维与可观测性（9）
- `/api/v1/operational/health` — 健康检查 (GetHealth)
- `/api/v1/operational/metrics` — Prometheus 指标 (GetMetrics)
- `/api/v1/operational/alerts` — 系统告警列表 (GetAlerts)
- `/api/v1/operational/rate-limit/stats` — 速率限制统计 (GetRateLimitStats)
- `/api/v1/operational/tasks` — 运维任务列表 (GetTasks)
- `/api/v1/operational/tasks/status` — 任务批量状态概览 (GetTaskStatus)
- `/api/v1/operational/tasks/{taskName}/trigger` — 触发指定任务 (TriggerTask)
- `/api/v1/operational/cutover` — 运维切换控制 (TriggerCutover)
- `/api/v1/operational/consistency-check` — 数据一致性巡检 (TriggerConsistencyCheck)

#### 认证与 OIDC（7）
- `/auth/login` — OAuth2 登录授权
- `/auth/callback` — OAuth2 回调处理
- `/auth/session` — 会话状态查询
- `/auth/refresh` — 刷新访问令牌
- `/auth/logout` — 注销并清理会话
- `/.well-known/oidc` — OIDC Discovery 文档
- `/.well-known/jwks.json` — 公钥 JWKS 配置

#### 组织命令与互操作（10）
- `/api/v1/organization-units` — 创建组织单元（命令入口）
- `/api/v1/organization-units/{code}` — 更新/替换组织单元
- `/api/v1/organization-units/{code}/versions` — 创建时态版本（CQRS 命令 → 版本管理）
- `/api/v1/organization-units/{code}/events` — 时态事件（如 DEACTIVATE 对应版本作废）
- `/api/v1/organization-units/{code}/suspend` — 暂停组织（业务停用）
- `/api/v1/organization-units/{code}/activate` — 恢复组织（业务启用）
- `/api/v1/organization-units/validate` — 提交前验证（规则/告警）
- `/api/v1/organization-units/{code}/refresh-hierarchy` — 单组织层级重算（维护场景）
- `/api/v1/organization-units/batch-refresh-hierarchy` — 批量层级重算（迁移/修复）
- `/api/v1/corehr/organizations` — CoreHR 兼容输出（受控暴露）

> 🛈 DEV 专用工具端点（如 `/auth/dev-token`、`/dev/status` 等）保留在命令服务 `devtools` 路由，仅在开发模式启用，不计入 OpenAPI 对外契约。

---

## GraphQL 查询 API（Query Service, Port 8090）
权威规范: `docs/api/schema.graphql`

> 说明: 以 Schema v4.6.0 为唯一事实来源；若脚本与 Schema 结果不一致，以 Schema 为准并立即提报修复。

### 查询列表（Schema v4.6.0，共 9 个公开查询）
- `organizations(filter, pagination): OrganizationConnection!` — 组织分页查询，支持过滤、时态视图与统一分页结构。
- `organization(code, asOfDate): Organization` — 按业务编码获取单个组织，支持 asOfDate 指定时间点快照。
- `organizationStats(asOfDate, includeHistorical): OrganizationStats!` — 多维统计（总量、类型分布、最早/最新生效日），可包含历史态。
- `organizationHierarchy(code, tenantId): OrganizationHierarchy` — 输出完整层级信息（路径、父链、叶子判断）。
- `organizationSubtree(code, tenantId, maxDepth, includeInactive): [OrganizationHierarchy!]!` — 多层级子树查询，适配 17 层深度展示与可选停用节点。
- `hierarchyStatistics(tenantId, includeIntegrityCheck): HierarchyStatistics!` — 层级结构完整性统计与一致性检查入口。
- `auditHistory(recordId, ...): [AuditLogDetail!]!` — 按 temporal `recordId` 返回完整审计轨迹，含操作人、前后快照。
- `auditLog(auditId: String!): AuditLogDetail` — 获取单条审计记录（before/after/changedFields）。
- `organizationVersions(code: String!, includeDeleted: Boolean = false): [Organization!]!` — 返回指定组织的所有时态版本，按 `effectiveDate` 升序，默认过滤软删记录。

### 关键实现要点
- **PostgreSQL 原生**：所有查询直连 PostgreSQL，利用递归 CTE、分区索引与物化视图缓存（详见 `docs/architecture/query-layer.md`）。
- **时态支持**：统一通过 `effectiveDate/endDate` 字段派生当前、未来、历史状态；`asOfDate` 汇聚服务端判断，前端无需重复逻辑。
- **层级性能**：`organizationSubtree`/`hierarchyStatistics` 共用层级缓存与 `codePath` 前缀索引，保障 17 层深度 <200ms。
- **审计链路**：`auditHistory`/`auditLog` 依赖最新审计模型（recordId 粒度），输出字段与 `docs/api/schema.graphql` 对齐。
- **认证约束**：所有查询均需 `Authorization: Bearer` 与 `X-Tenant-ID`；细粒度权限分别为 `org:read`/`org:read:history`/`org:read:hierarchy`/`org:read:stats`/`org:read:audit`。

### 自动化校验与脚本状态
- `node scripts/generate-implementation-inventory.js` 已与 Schema 同步输出 9 个公开查询，确保清单与契约一致。
- 如 Schema 更新需同步运行脚本并刷新本文档，保持唯一事实来源一致性。

---

## 🔎 验证命令与报告路径

- 生成实现清单（校对用）
  - `node scripts/generate-implementation-inventory.js > temp-inventory.md`
- 架构一致性校验
  - `node scripts/quality/architecture-validator.js`（报告：`reports/architecture/architecture-validation.json`）
- 契约文件权威位置
  - REST OpenAPI: `docs/api/openapi.yaml`
  - GraphQL Schema: `docs/api/schema.graphql`

---

## 后端（Go）关键导出（Key Exported Items）

### 处理器（Handlers） - 26个导出方法
> **数据来源**: `reports/implementation-inventory.json.goHandlers`（自动扫描 `cmd/organization-command-service/internal/handlers`）

#### 命令服务 · 开发工具 (`devtools.go`)
- `SetupRoutes` — 开发工具路由注册
- `GenerateDevToken` — 生成 Dev JWT 令牌
- `GetTokenInfo` — 查询 Dev 令牌信息
- `DevStatus` — DEV 环境运行状况
- `ListTestEndpoints` — 列出测试端点清单
- `DatabaseStatus` — 数据库状态页
- `PerformanceMetrics` — 性能指标快照
- `TestAPI` — 集成测试辅助端点

#### 命令服务 · 运维可观测 (`operational.go`)
- `SetupRoutes` — 运维路由注册
- `GetRateLimitStats` — 速率限制统计
- `GetHealth` — 服务健康检查
- `GetMetrics` — Prometheus 指标输出
- `GetAlerts` — 告警汇总
- `GetTasks` — 运维任务列表
- `GetTaskStatus` — 任务运行状态
- `TriggerTask` — 触发指定运维任务
- `TriggerCutover` — 运维切换控制
- `TriggerConsistencyCheck` — 数据一致性巡检

#### 命令服务 · 组织业务 (`organization.go`)
- `SetupRoutes` — 组织业务路由注册
- `CreateOrganization` — 创建组织单元
- `CreateOrganizationVersion` — 创建时态版本
- `UpdateOrganization` — 更新组织信息
- `SuspendOrganization` — 暂停组织
- `ActivateOrganization` — 激活组织
- `CreateOrganizationEvent` — 处理组织事件（含 DEACTIVATE 作废流程）
- `UpdateHistoryRecord` — 更新历史记录

### 服务层（Services） - 19个导出类型
> **数据来源**: `reports/implementation-inventory.json.goServices`

#### 层级级联 (`internal/services/cascade.go`)
- `CascadeUpdateService` — 层级变更级联处理
- `CascadeTask` — 级联任务定义

#### 运维调度 (`internal/services/operational_scheduler.go`)
- `OperationalScheduler` — 后台任务调度器
- `ScheduledTask` — 调度任务结构

#### 组织时态服务 (`internal/services/organization_temporal_service.go`)
- `OrganizationTemporalService` — 时态领域协调服务
- `TemporalCreateVersionRequest` — 创建版本请求模型
- `TemporalUpdateVersionRequest` — 更新版本请求模型
- `TemporalDeleteVersionRequest` — 删除版本请求模型
- `TemporalStatusChangeRequest` — 状态变更请求模型

#### 时态核心服务 (`internal/services/temporal.go`)
- `TemporalService` — 时态版本管理核心
- `InsertVersionRequest` — 插入版本请求载体
- `OrganizationData` — 组织数据快照
- `DeleteVersionRequest` — 删除版本请求载体
- `ChangeEffectiveDateRequest` — 生效日期调整请求
- `SuspendActivateRequest` — 暂停/激活请求载体
- `VersionResponse` — 版本操作响应模型

#### 时态监控 (`internal/services/temporal_monitor.go`)
- `TemporalMonitor` — 时态数据健康监控
- `MonitoringMetrics` — 监控指标汇总
- `AlertRule` — 告警规则定义

### 中间件层（Middleware） - 8个导出类型
#### REST中间件 (`internal/middleware/`)
- `PerformanceMiddleware` - 性能监控中间件 (`performance.go`)
- `RequestMiddleware` - 请求处理中间件 (`request.go`)
- `RateLimitMiddleware` - 速率限制中间件 (`ratelimit.go`)

#### GraphQL中间件 (`internal/middleware/`)
- `GraphQLEnvelopeMiddleware` - GraphQL响应封装中间件 (`graphql_envelope.go`)
- `RequestIDMiddleware` - 请求ID生成中间件 (`request_id.go`)

#### 认证中间件
- `JWTMiddleware` - JWT认证中间件 (REST: `auth/rest_middleware.go`)
- `GraphQLAuthMiddleware` - GraphQL认证中间件 (`auth/graphql_middleware.go`)

### 工具层（Utils） - 6个导出类型
#### 验证工具 (`internal/utils/validation.go`)
- `ValidationUtils` - 通用验证工具

#### 响应工具 (`internal/utils/response.go`)
- `ResponseBuilder` - 统一响应构建器

#### 业务验证器 (`internal/validators/business.go`)
- `BusinessValidator` - 业务规则验证器

### 审计与监控 - 4个导出类型
#### 审计日志 (`internal/audit/logger.go`)
- `AuditLogger` - 结构化审计日志记录器
  - 审计生产对齐 API 优先：before_data/after_data 中排除动态时态标记字段（is_current、is_temporal、is_future），以数据库触发器 `log_audit_changes()` 统一实现（见 `database/migrations/023_audit_exclude_dynamic_temporal_flags.sql`）。

#### 指标收集 (`internal/metrics/collector.go`)
- `MetricsCollector` - Prometheus指标收集器

### 架构特点
- **CQRS分离**: 命令服务(9090端口)与查询服务(8090端口)完全分离
- **PostgreSQL原生**: 直接操作PostgreSQL，无中间数据同步
- **时态数据**: 完整的时态版本管理和监控体系
- **企业级监控**: 完备的健康检查、指标收集、告警机制
- **开发友好**: 丰富的开发工具和调试端点

---

## 前端（TypeScript/React）关键导出（Key Exported Items）

基于最新IIG扫描的172个导出项（详见 `reports/implementation-inventory.json.tsExports`），下列按领域归纳关键模块：

### API客户端架构
#### 统一客户端 (`unified-client.ts`)
- `UnifiedGraphQLClient` - GraphQL查询专用客户端 (CQRS-Query)
- `UnifiedRESTClient` - REST命令专用客户端 (CQRS-Command)
- `unifiedGraphQLClient` - GraphQL客户端实例
- `unifiedRESTClient` - REST客户端实例
- `createGraphQLClient` - GraphQL客户端工厂
- `createRESTClient` - REST客户端工厂
- `validateCQRSUsage` - CQRS使用规范验证

#### 认证管理 (`auth.ts`)
- `AuthManager` - OAuth认证管理器
- `authManager` - 认证管理器实例

#### 错误处理系统 (`error-handling.ts`)
- `OAuthError` - OAuth专用错误类
- `ErrorHandler` - 统一错误处理器
- `UserFriendlyError` - 用户友好错误类
- `isUserFriendlyError` - 用户友好错误判断
- `isOAuthError` - OAuth错误判断
- `withErrorHandling` - 错误处理装饰器
- `useErrorHandler` - 错误处理Hook
- `withRetry` - 重试装饰器
- `withOAuthRetry` - OAuth重试装饰器
- `withOAuthAwareErrorHandling` - OAuth感知错误处理
- ⚠️ 导入指引：错误处理类与运行时守卫统一从 `frontend/src/shared/api/error-handling.ts`、`frontend/src/shared/api/type-guards.ts` 获取；`frontend/src/shared/types/api.ts` 仅保留纯类型定义，禁止再度扩散临时导出。

### 数据管理层
#### 状态管理Hooks ⭐ **已修复稳定版**
- `useEnterpriseOrganizations` - 企业级组织管理 (`useEnterpriseOrganizations.ts`) ✅ **主要Hook - 已修复初始化逻辑**
- `useOrganizations` - 组织列表管理 (`useOrganizations.ts`) ⚠️ **已废弃** - 兼容封装，调用useEnterpriseOrganizations
- `useOrganization` - 单个组织管理 ⚠️ **已废弃** - 兼容封装，调用useEnterpriseOrganizations
- `useMessages` - 用户消息管理 (`useMessages.ts`) ✅ **稳定**

#### 组织变更操作 (`useOrganizationMutations.ts`)
- `useCreateOrganization` - 创建组织Hook
- `useUpdateOrganization` - 更新组织Hook
- `useSuspendOrganization` - 暂停组织Hook
- `useActivateOrganization` - 激活组织Hook
- `useCreateOrganizationVersion` - 创建时态版本Hook（契约端点 `/api/v1/organization-units/{code}/versions`）

#### 时态数据管理（GraphQL `organizationVersions`）
- `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:164` 使用 `unifiedGraphQLClient` 调用 GraphQL 查询 `organizationVersions(code: String!)` 获取时间线。
- 查询失败时回退至 `organization(code: String!)` 单体快照逻辑，保证历史数据不可用时仍有兜底视图。
- 版本操作（创建/作废）完成后通过 `useCreateOrganizationVersion`、`POST /{code}/events` 触发缓存刷新，保持前端视图与契约一致。

### 类型系统与验证
#### 类型守卫 (`type-guards.ts`)
- `ValidationError` - 验证错误类
- `validateOrganizationUnit` - 组织单元验证
- `validateCreateOrganizationInput` - 创建输入验证
- `validateUpdateOrganizationInput` - 更新输入验证
- `validateCreateOrganizationResponse` - 创建响应验证
- `validateGraphQLVariables` - GraphQL变量验证
- `validateGraphQLOrganizationResponse` - GraphQL响应验证
- `validateGraphQLOrganizationList` - GraphQL列表验证
- `isGraphQLError` - GraphQL错误判断
- `isGraphQLSuccessResponse` - GraphQL成功响应判断
- `isAPIError` - API错误判断
- `isValidationError` - 验证错误判断
- `isNetworkError` - 网络错误判断
- `safeTransformGraphQLToOrganizationUnit` - 安全类型转换
- `safeTransformCreateInputToAPI` - 安全输入转换

#### 类型转换器 (`converters.ts`)
- `convertGraphQLToOrganizationUnit` - GraphQL到组织单元转换
- `convertGraphQLToTemporalOrganizationUnit` - 时态组织单元转换
- `convertCreateInputToREST` - 创建输入到REST转换
- `convertUpdateInputToREST` - 更新输入到REST转换
- `validateOrganizationUnit` - 组织单元验证
- `validateOrganizationUnitList` - 组织列表验证
- `checkTypeConsistency` - 类型一致性检查
- `generateTypeDefinition` - 类型定义生成
- `logTypeSyncReport` - 类型同步报告

### 配置管理系统
#### 端口配置 (`ports.ts`)
- `SERVICE_PORTS` - 服务端口配置
- `getServicePort` - 端口获取函数
- `buildServiceURL` - 服务URL构建
- `CQRS_ENDPOINTS` - CQRS端点配置
- `FRONTEND_ENDPOINTS` - 前端端点配置
- `INFRASTRUCTURE_ENDPOINTS` - 基础设施端点
- `MONITORING_ENDPOINTS` - 监控端点配置
- `validatePortConfiguration` - 端口配置验证
- `generatePortConfigReport` - 端口配置报告

#### 统一常量管理 (`constants.ts`) ⭐ **P2级配置常量集中管理完成**
- `TIMEOUTS` - 时间和超时常量 (15+个)
  - `API_REQUEST` - API请求超时 (30秒)
  - `DEBOUNCE_SEARCH` - 搜索防抖 (1.5秒)
  - `E2E_TEST_SUITE` - E2E测试套件超时 (30秒)
  - `QUERY_STALE_TIME` - React Query数据新鲜度 (30秒)
- `LIMITS` - 性能和限制常量 (10+个)
  - `PAGE_SIZE_DEFAULT` - 默认分页大小 (20)
  - `PAGE_SIZE_MAX` - 最大分页大小 (100)
  - `SEARCH_MIN_LENGTH` - 搜索最小长度 (2)
  - `FILE_UPLOAD_MAX_SIZE` - 文件上传最大大小 (10MB)
- `BUSINESS_CONSTANTS` - 业务相关常量 (10+个)
  - `ROOT_ORG_CODE` - 根组织编码 ("1000000")
  - `ORG_LEVEL_MIN/MAX` - 组织层级限制 (1-10)
  - `ORG_NAME_MAX_LENGTH` - 组织名称最大长度 (100)
- `UI_CONSTANTS` - UI相关常量 (响应式断点、Z-index层级、动画时长)
- `API_CONSTANTS` - API相关常量 (版本、路径、状态码、重试策略)
- `TEST_CONSTANTS` - 测试相关常量 (超时、性能基准、测试数据)
- `RETRY_CONSTANTS` - 重试和回退常量 (网络、UI交互、数据同步)
- `VALIDATION_CONSTANTS` - 验证相关常量 (组织、时态、通用验证规则)
- `FEATURE_FLAGS` - 功能开关常量 (实验性功能、性能优化、调试功能)
- `generateConstantsReport` - 常量使用统计报告生成器

#### 租户管理 (`tenant.ts`)
- `TenantManager` - 租户管理器
- `DEFAULT_TENANT_CONFIG` - 默认租户配置
- `tenantManager` - 租户管理器实例
- `getCurrentTenantId` - 获取当前租户ID
- `isDefaultTenant` - 默认租户判断
- `getTenantConfig` - 获取租户配置

#### 环境配置 (`environment.ts`)
- `env` - 环境变量配置
- `validateEnvironmentConfig` - 环境配置验证

### 业务配置常量
#### 表单配置 (`formConfig.ts`)
- `ORGANIZATION_UNIT_TYPES` - 组织单元类型
- `ORGANIZATION_STATUSES` - 组织状态配置
- `BUSINESS_STATUSES` - 业务状态配置
- `ORGANIZATION_LEVELS` - 组织层级配置
- `FORM_DEFAULTS` - 表单默认值
- `PAGINATION_DEFAULTS` - 分页默认配置

#### 表格配置 (`tableConfig.ts`)
- `TABLE_COLUMNS` - 表格列定义
- `STATUS_COLORS` - 状态颜色映射
- `LOADING_STATES` - 加载状态配置

#### 时态配置 (`temporalStatus.ts` & `temporal/index.ts`)
- `TEMPORAL_STATUS_COLORS` - 时态状态颜色
- `temporalStatusUtils` - 时态状态工具
- `TEMPORAL_CONSTANTS` - 时态常量
- `temporalUtils` - 时态工具函数

### 工具函数库
#### 统一工具导出 (`utils/index.ts`) ⭐ **P2级工具统一架构完成**
- **统一时态工具导出**:
  - `TemporalConverter` - 完整时态转换器类 (来自 temporal-converter.ts)
  - `TemporalUtils` - 时态工具函数集合
  - `DateUtils` - 便捷日期工具简化访问对象
- **性能优化设计**:
  - 懒加载导入策略，减少初始包大小
  - 统一导出接口，避免重复导入
  - Tree-shaking优化支持
- **向后兼容支持**:
  - 保持现有导入路径可用
  - 渐进式迁移机制
  - 完整的功能覆盖保证

#### 业务工具 (`organization-helpers.ts`)
- `normalizeParentCode` - 标准化父级编码
- `isRootOrganization` - 根组织判断
- `getOrganizationLevelText` - 组织层级文本

#### 权限工具 (`organizationPermissions.ts`)
- `getOperationPermissionsByScopes` - 按作用域获取操作权限
- `getOperationPermissions` - 获取操作权限

#### 状态工具 (`statusUtils.ts`)
- `STATUS_CONFIG` - 状态配置
- `statusUtils` - 状态工具函数

#### 时态工具 (`temporal-converter.ts`) 
- `TemporalConverter` - 时态转换器类 (权威实现)
- `TemporalUtils` - 时态工具函数 (唯一来源)

### 验证系统
#### 统一验证导出 (`validation/index.ts`) ⭐ **P1级验证系统统一完成**
- **统一验证架构**:
  - 完整的错误处理体系 (来自 type-guards.ts)
  - Zod Schema验证支持 (来自 schemas.ts)
  - 业务逻辑验证规则集成
- **错误处理统一**:
  - `ValidationError` - 统一验证错误类
  - `validateOrganizationUnit` - 组织单元验证
  - `validateCreateOrganizationInput` - 创建输入验证
  - `validateUpdateOrganizationInput` - 更新输入验证
  - `validateGraphQLVariables` - GraphQL变量验证
- **类型守卫集成**:
  - `isValidationError` - 验证错误判断
  - `isAPIError` - API错误判断
  - `isNetworkError` - 网络错误判断
  - `safeTransformGraphQLToOrganizationUnit` - 安全类型转换
- **向后兼容机制**:
  - 保持现有验证函数可用
  - 渐进式迁移支持
  - 完整的功能覆盖保证

#### Schema验证 (`schemas.ts`)
- `OrganizationUnitSchema` - 组织单元Schema
- `CreateOrganizationInputSchema` - 创建输入Schema
- `CreateOrganizationResponseSchema` - 创建响应Schema
- `UpdateOrganizationInputSchema` - 更新输入Schema
- `GraphQLVariablesSchema` - GraphQL变量Schema
- `GraphQLOrganizationResponseSchema` - GraphQL组织响应Schema

#### 简单验证 (`simple-validation.ts`) ⚠️ **已弃用 - 迁移至统一验证系统**
- `SimpleValidationError` - 简单验证错误类 (已弃用)
- `validateOrganizationBasic` - 组织基础验证 (已弃用)
- `validateOrganizationUpdate` - 组织更新验证 (已弃用)
- `validateOrganizationResponse` - 组织响应验证 (已弃用)
- `formatValidationErrors` - 格式化验证错误 (已弃用)
- `getFieldError` - 获取字段错误 (已弃用)
- `validateStatusUpdate` - 状态更新验证 (已弃用)
- `basicValidation` - 基础验证函数 (已弃用)
- `safeTransform` - 安全转换函数 (已弃用)
- `validateCreateOrganizationInput` - 验证创建输入 (已弃用)
- `validateUpdateOrganizationInput` - 验证更新输入 (已弃用)

### 设计系统
#### 品牌令牌 (`brand.ts`)
- `cubecastleBrandTokens` - Cube Castle品牌令牌

#### 颜色系统 (`colorTokens.ts`)
- `baseColors` - 基础颜色
- `statusColors` - 状态颜色
- `legacyColors` - 遗留颜色

### 专用API客户端
#### 企业级GraphQL (`graphql-enterprise-adapter.ts`)
- `GraphQLEnterpriseAdapter` - 企业级GraphQL适配器
- `graphqlEnterpriseAdapter` - 适配器实例
- `useEnterpriseGraphQL` - 企业级GraphQL Hook

#### 契约测试 (`contract-testing.ts`)
- `contractTestingAPI` - 契约测试API客户端

### API类型系统 (`api.ts`)
- `APIError` - API错误基类
- `ValidationError` - 验证错误类
- `isGraphQLResponse` - GraphQL响应判断
- `hasGraphQLErrors` - GraphQL错误检查
- `isAPIError` - API错误判断
- `isValidationError` - 验证错误判断

### 错误消息系统 (`error-messages.ts`)
- `getErrorMessage` - 获取错误消息
- `formatErrorForUser` - 格式化用户错误
- `SUCCESS_MESSAGES` - 成功消息常量

### 表单验证规则 (`ValidationRules.ts`)
- `validateForm` - 表单验证函数

### 时态验证适配层 (`temporal-validation-adapter.ts`)
- `validateTemporalDate` - 与遗留接口保持一致的时态日期验证包装

---

## 运维与脚本（DevOps & Scripts）

### 质量保证脚本
- `scripts/generate-implementation-inventory.js` - **实现清单生成器** (避免重复造轮子)
- `scripts/quality/duplicate-detection.sh` - 重复代码检测工具
- `scripts/quality/architecture-validator.js` - 架构一致性验证
- `scripts/quality/document-sync.js` - 文档同步监控
- `scripts/quality/hierarchy-consistency-guard.sh` - 组织层级一致性守卫 (需要数据库连接)

### 数据库修复脚本 ⭐ **P1级问题修复完成**
- `scripts/fix-graphql-scan-issue.sql` - **GraphQL扫描问题修复脚本**
  - 功能: 修复 "sql: expected 25 destination arguments in Scan, not 24" 错误
  - 解决: audit_logs 表缺失 business_entity_type 字段
  - 操作: 添加字段并更新 141 条现有记录的默认值
  - 状态: ✅ 已执行，GraphQL查询服务正常运行
- `scripts/maintenance/run-hierarchy-consistency-check.sh` - 组织层级巡检脚本，输出 CSV 汇总异常数据（需 `psql`）
- `sql/hierarchy-consistency-check.sql` - 只读巡检 SQL，供脚本/CI 调用

### 开发环境脚本
- **根目录 Makefile** - 统一开发命令入口
  - `make docker-up` - 启动PostgreSQL + Redis
  - `make run-dev` - 启动后端服务 (9090 + 8090)
  - `make frontend-dev` - 启动前端开发服务器
  - `make jwt-dev-mint` - 生成开发JWT令牌
  - `make status` - 查看所有服务状态

### CI/CD工作流
- `.github/workflows/contract-testing.yml` - 契约测试自动化
- `.github/workflows/duplicate-code-detection.yml` - 重复代码检测
- `.git/hooks/pre-commit` - 提交前质量检查

### 监控与部署
- `docker-compose.yml` - 本地开发环境编排
- `docker-compose.monitoring.yml` - 监控服务编排 (Prometheus/Grafana)
- 各种启动脚本: `start.sh`, `start_smart.sh` 等

---

## 使用与更新指引（How to Use & Update）

### 🚨 **强制流程** (基于CLAUDE.md第9条原则)
1) **开发前必检**: 运行 `node scripts/generate-implementation-inventory.js` 查看现有实现
2) **避免重复造轮子**: 优先使用现有的API/函数/组件，禁止重复创建相同功能
3) **契约优先**: 新增端点前先更新契约文件 (OpenAPI/GraphQL)，通过评审后再实现
4) **强制登记**: 新增功能后必须重新运行清单生成器，验证功能已正确登记

### 🔧 **最近修复状态更新 (2025-09-13)**

#### ✅ **组织列表数据显示问题修复**
- **问题**: 前端页面显示"暂无组织数据"，但后端实际有3个组织
- **根本原因**: `useEnterpriseOrganizations` Hook的初始化逻辑有缺陷
- **修复文件**: `frontend/src/shared/hooks/useEnterpriseOrganizations.ts:426-430`
- **修复内容**: 移除`if (initialParams)`条件，确保无参数时也调用`fetchOrganizations()`

#### ✅ **组织详情页面GraphQL查询错误修复**
- **问题**: 点击"详情管理"报错`Cannot query field "organizationVersions" on type "Query"`
- **根本原因**: `TemporalMasterDetailView`使用了不存在的GraphQL查询字段
- **修复文件**: `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:140-220`
- **修复内容**: 
  - 将`organizationVersions`查询替换为`organization`查询
  - 修复数据映射逻辑，从数组处理改为单对象处理
  - 修正语法错误：`}));` → `}];`

#### 📊 **修复效果验证**
- ✅ 组织列表正常显示3个组织单元
- ✅ 组织详情页面成功加载
- ✅ 时间轴导航功能正常
- ✅ 后端GraphQL服务性能良好（4-8ms响应时间）
- ✅ JWT认证正常工作

### 📋 **更新维护**
1) **自动更新**: 使用 `scripts/generate-implementation-inventory.js` 自动生成最新清单
2) **手动补充**: 对脚本无法识别的重要组件，手动补充到相应分类
3) **保持同步**: 代码变更后及时更新清单，确保文档与代码一致
4) **版本管理**: 重大变更时更新版本号和变更记录

### ⚠️ **重要提醒**
- **权威性**: API规范 (`docs/api/*`) 为唯一权威来源，本清单仅作导航
- **CQRS分离**: 严格区分查询(GraphQL)和命令(REST)，不得混用
- **命名一致性**: 遵循camelCase字段命名，路径参数使用{code}
- **类型安全**: 前端组件必须使用类型守卫和验证系统

---

## 重复造轮子风险提醒 🚨

### **高风险重复区域** (已有完整实现)
- ❌ **API客户端**: 统一的GraphQL/REST客户端已存在
- ❌ **错误处理**: 完整的错误处理和用户友好消息系统
- ❌ **类型转换**: GraphQL/REST类型转换器已完备
- ❌ **状态管理**: 组织CRUD操作的所有Hook都已实现
- ❌ **配置管理**: 端口、租户、环境配置系统已完善
- ❌ **验证系统**: Schema验证和类型守卫已全面覆盖

### **安全扩展区域** (可以新增)
- ✅ **新业务领域**: 员工管理、权限系统等全新模块
- ✅ **专用工具**: 特定业务场景的专用组件
- ✅ **集成适配**: 外部系统集成适配器
- ✅ **监控增强**: 新的监控指标和告警规则

---

## 📊 **统计摘要** ⭐ **2025-09-14 IIG护卫系统集成完成版**

### **🎯 API优先架构成果统计** ⭐ **基于最新扫描结果**

- **REST API端点**: 26个端点（运维9 + 认证7 + 组织命令10）⭐ **API优先设计完成**
  - 权威来源: `docs/api/openapi.yaml` 完整定义
  - 分类说明: 运维可观测性端点9个、认证/OIDC端点7个、组织命令端点10个
  - 认证架构: OAuth2 + OIDC标准实现，JWT + JWKS公钥验证
  - 业务端点: 组织CRUD + 时态版本管理 + 层级维护
  - 实现状态: 所有端点严格按照OpenAPI规范实现
- **GraphQL Schema**: 9个公开查询 ⭐ **Schema优先设计**
  - 权威来源: `docs/api/schema.graphql` 类型安全定义
  - 查询支持: 组织CRUD + 时态查询 + 统计聚合
  - 类型验证: 100%强类型检查和契约一致性

#### **🏗️ 实现层统计** (Implementation Layer) ⭐ **IIG护卫验证完成**
- **Go后端组件**: 26个处理器方法 + 19个服务类型 = **45个关键组件**
  - 开发工具处理器: 8个方法（JWT令牌、状态检查、性能监控）
  - 运维管理处理器: 10个方法（健康检查、指标收集、任务管理、切换控制）
  - 组织业务处理器: 8个方法（组织CRUD、时态版本、事件处理）
  - 核心服务层: 19个服务类型（级联更新、时态管理、监控告警）
- **前端统一架构**: 172个导出组件 ⭐ **IIG护卫扫描覆盖**
  - API客户端架构: 统一GraphQL/REST客户端，OAuth认证管理
  - 数据管理层: 企业级组织管理Hook，时态数据API
  - 类型系统: 完整类型守卫和转换器，Zod Schema验证
  - 配置管理: 端口配置、租户管理、环境变量、常量系统
  - 工具函数库: 业务工具、权限工具、状态工具、时态转换器

#### **🛡️ 质量保证统计** (Quality Assurance) ⭐ **IIG护卫系统完成**
- **IIG护卫保护**: 223个实现组件分析，0个重复检测，100%唯一性保证（来源：`reports/iig-guardian/iig-guardian-report.json`）
- **P3系统集成**: 重复代码率2.11%，架构违规0个，质量门禁通过
- **API一致性**: camelCase命名100%合规，Schema验证通过
- **架构合规**: CQRS协议分离100%执行，PostgreSQL原生架构
- **实时监控**: IIG护卫系统与P3防控深度集成，自动化检测运行良好

#### **🔧 工具链统计** (Toolchain)
- **质量保证脚本**: 12个企业级扫描和验证工具
- **API管理工具**: OpenAPI/GraphQL契约管理和验证
- **CI/CD集成**: API契约变更自动验证和部署门禁

### **🎯 API优先架构成熟度评估**
- ✅ **API优先设计**: 100%端点先定义契约后实现代码
- ✅ **契约驱动开发**: OpenAPI + GraphQL Schema作为开发起点
- ✅ **API契约测试**: 32个测试验证API规范与实现一致性
- ✅ **CQRS架构**: 查询/命令API完全分离，协议专用化
- ✅ **PostgreSQL原生**: 单一数据源，API性能优化
- ✅ **API版本管理**: 规范化的API版本控制和向后兼容
- ✅ **质量门禁**: API契约变更自动验证，阻止不合规实现
- ✅ **开发工具**: API测试、契约验证、Schema管理工具
- ✅ **生产就绪**: API服务100%可用，契约一致性保证

---

## 🛡️ IIG护卫系统集成状态 ⭐ **新增 (2025-09-10)**

### **实现清单护卫系统 (Implementation Inventory Guardian)**
**核心职责**: 防止重复开发，维护实现唯一性，管理功能清单

#### 🔧 **护卫机制**
- **预开发检查**: 运行 `node scripts/generate-implementation-inventory.js` 强制检查现有实现
- **重复检测防护**: 与P3.1重复代码检测系统深度集成 (当前重复率: 2.11%)
- **架构一致性**: 与P3.2架构验证器联动，确保CQRS+端口+契约标准
- **文档同步**: 与P3.3文档同步引擎协作，确保清单与代码一致

#### 📋 **护卫工作流**
```yaml
开发前强制检查 (IIG护卫启动):
  第一步: 执行实现清单生成 → 分析现有70+个后端组件 + 4个前端统一系统
  第二步: 搜索相关功能关键词 → 验证API/Hook/组件/服务是否已实现
  第三步: 评估重复风险 → 基于P3.1检测结果和清单对比分析
  第四步: 提供复用建议 → 推荐现有API端点、Hook、组件的复用方案
  
功能登记强制流程 (避免重复造轮子):
  新增后: 重新生成实现清单 → 验证新功能正确登记
  验证期: 运行P3系统全套检查 → 确保架构一致性和质量标准
  文档更新: 同步更新此清单文档 → 为团队提供最新功能索引
```

#### 🚨 **IIG护卫强制禁止事项**
- **跳过清单检查**: 不运行 `generate-implementation-inventory.js` 就开始新功能开发
- **忽视现有实现**: 在清单中发现可用资源仍重复创建相同功能  
- **功能未登记**: 新增API/Hook/组件后不更新实现清单
- **违反护卫原则**: 忽视"现有资源优先"和"实现唯一性"原则

#### 📊 **护卫效果统计**
- **重复防护率**: 93%+ (120+个分散导出 → 4个统一系统)
- **清单覆盖度**: 100% (26个REST端点 + 9个GraphQL查询 + 45个后端组件 + 172个前端导出)
- **质量门禁**: 与P3系统100%集成，自动化检测和报告
- **团队效率**: 显著减少"重复造轮子"问题，提升代码复用率

---

## 变更记录（Changelog）
- **v1.9.0 实现统计刷新版（2025-09-24）**: ⭐ **命令端点/实现清单全面同步**
- **同步**: OpenAPI 26、GraphQL 9、Go处理器 26、Go服务类型 19、前端导出 146（与最新 IIG 报告一致）
  - **调整**: REST 命令端点重分组（运维 9 + 认证 7 + 组织命令 10），去除历史路径引用
  - **校验**: Go 处理器/服务列表按脚本输出重新排列，确保与 IIG 报告一致
  - **记录**: 文档顶部版本/统计更新至 v1.9.0，强化 JSON 作为单一事实来源
  - **监控**: IIG 护卫报告 `analysedImplementations=223`、重复检测 0、duplicateCodeRate 2.11%
- **v1.8.0 IIG护卫系统实时监控版（2025-09-15）**: ⭐ **IIG护卫系统实时监控完成**
  - **扫描**: 前端组件从146个增至162个（新增16个导出项）
  - **更新**: Go后端处理器从26个增至28个（新增2个处理器方法）
  - **监控**: IIG护卫系统实时监控运行良好，检测210个组件，0重复
  - **集成**: P3系统完全集成，代码重复率2.11%，架构违规0个
  - **验证**: 97个前端文件架构验证100%通过，质量门禁达标
  - **成果**: IIG护卫系统进入实时监控模式，为持续集成提供保障
- **v1.7.0 IIG护卫系统集成完成版（2025-09-14）**: ⭐ **IIG护卫系统正式上线**
  - **新增**: 基于最新IIG扫描的17个REST API端点（新增7个OAuth2/OIDC认证端点）
  - **更新**: Go后端组件从39个增至45个（26个处理器 + 19个服务类型）
  - **扫描**: 前端146个导出组件完整扫描和分类整理
  - **护卫**: IIG护卫系统与P3系统100%集成，实现210个组件分析，0重复检测
  - **验证**: 代码重复率降至2.11%，架构违规0个，质量门禁100%通过
  - **成果**: 确立IIG护卫作为项目核心防护系统，实现"重复造轮子"零容忍
- **v1.6.1 文档一致性修订版（2025-09-14）**: 事务化版本删除对齐完成
- **v1.5 API优先原则强化版（2025-09-10）**: ⭐ **API优先开发原则全面实施**
  - 新增: API优先开发原则和维护规则章节，强调"Contract First, Code Second"
  - 更新: 基于最新扫描结果的完整实现清单 (10个REST端点 + 9个GraphQL查询 + 39个Go组件)
  - 强化: API契约层统计和架构成熟度评估，突出契约驱动开发
  - 优化: 统计摘要重构为API优先架构视角，展示契约测试100%覆盖率
  - 成果: 确立API优先为项目核心开发原则，实现契约与代码100%一致性
- **v1.4 IIG护卫系统集成版（2025-09-10）**: ⭐ **实现清单护卫系统启动**
  - 新增: IIG护卫系统完整设计和工作流程
  - 集成: 与P3三层防控系统深度融合
  - 强化: 预开发检查和功能登记强制流程
  - 成果: 实现清单护卫系统正式上线，为项目提供重复开发防护
- **v1.3 重复代码消除完成版（2025-09-09）**: ⭐ **S级重复代码消除工程完成**
  - 新增: 统一配置管理系统 (`config/constants.ts`) - 85+个常量集中管理
  - 新增: 统一工具函数导出 (`utils/index.ts`) - 时态工具统一架构
  - 新增: 统一验证系统导出 (`validation/index.ts`) - P1级验证系统统一
  - 新增: 数据库修复脚本 (`scripts/fix-graphql-scan-issue.sql`) - P1级问题修复
  - 优化: 前端架构从120+个分散导出精简为4个核心统一系统 (93%重复消除)
  - 成果: P1级验证系统 + P2级工具/配置统一 + GraphQL扫描问题彻底修复
- **v1.2 完整登记版（2025-09-09）**: 基于实际代码扫描的完整实现登记
  - 完善: GraphQL查询处理器完整功能描述和实现路径
  - 更新: 架构成熟度和系统可用性状态
  - 优化: 移除冗余的修复过程记录，专注于当前实现状态
- **v1.0 生产就绪版（2025-09-09）**: 基于实际代码扫描的完整清单
  - 新增: 120+个前端导出项详细分类
  - 新增: 26个Go处理器和14个服务类型
  - 新增: 重复造轮子风险分析和防范指导
  - 新增: 统计摘要和架构成熟度评估
