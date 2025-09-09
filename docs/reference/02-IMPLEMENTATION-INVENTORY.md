# Cube Castle 实现清单（Implementation Inventory）

版本: v0.1 初稿  
维护人: 架构组（与各子域模块共同维护）  
范围: 本仓库已实现的 API/函数/接口（按 CQRS 与目录分区）

> 目的（Purpose）
> - 中文: 统一登记当前已实现的 API、导出函数与接口，以及所属文件与简要说明，避免重复造轮子，便于新成员快速定位能力与复用。
> - EN: Centralized, bilingual catalog of implemented APIs, exported functions and interfaces with file locations and short descriptions to reduce duplication and speed onboarding.

---

## 维护与收录原则（Maintaining Rules）
- 单一来源: API 端点与权限以 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql` 为唯一权威；此清单仅做导航索引（No divergence from spec）。
- CQRS: 查询统一 GraphQL；命令统一 REST。清单按“Query/Command”分区（Follow CQRS split）。
- 命名一致: API 层字段一律 camelCase；路径参数 `{code}`（Naming consistency per CLAUDE.md）。
- 粒度控制: 收录“对外可复用/可调用”的导出符号（exported/public）；内部私有函数不在本表（Public symbols only）。
- 更新时机: 每次合并涉及新端点/导出函数，需同步更新本清单（Update on merge）。

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

### 核心业务端点 (10个)
- `/api/v1/organization-units`
  - 中文: 创建组织单元（自动生成代码，级联路径初始化）
  - EN: Create organization unit (auto code, initialize hierarchy)
  - 实现: `cmd/organization-command-service/internal/handlers/organization.go: CreateOrganization`

- `/api/v1/organization-units/{code}`
  - 中文: 完全替换组织单元（PUT 语义，字段全量）
  - EN: Replace organization unit (full PUT semantics)
  - 实现: `handlers/organization.go: UpdateOrganization`

- `/api/v1/organization-units/{code}/versions`
  - 中文: 为既有组织创建新的时态版本（自动相邻边界调整）
  - EN: Create temporal version for existing org (adjacent boundary updates)
  - 实现: `handlers/organization.go: CreateOrganizationVersion`

- `/api/v1/organization-units/{code}/events`
  - 中文: 时态事件处理（如按 recordId 作废版本）
  - EN: Temporal event processing (e.g., deactivate by recordId)
  - 实现: `handlers/organization.go: CreateOrganizationEvent`

- `/api/v1/organization-units/{code}/suspend`
  - 中文: 业务停用（强制 status=INACTIVE，记录原因）
  - EN: Suspend organization (force status=INACTIVE)
  - 实现: `handlers/organization.go: SuspendOrganization`

- `/api/v1/organization-units/{code}/activate`
  - 中文: 业务启用（反向操作，恢复为 ACTIVE）
  - EN: Activate organization (reactivate back to ACTIVE)
  - 实现: `handlers/organization.go: ActivateOrganization`

- `/api/v1/organization-units/validate`
  - 中文: 操作前校验（规则检查/建议/告警）
  - EN: Pre-operation validation (rules, suggestions, warnings)
  - 实现: `handlers/organization.go` 校验逻辑

- `/api/v1/organization-units/{code}/refresh-hierarchy`
  - 中文: 单个组织层级修复（维护用途，非业务路径）
  - EN: Manual hierarchy refresh for one org (maintenance)
  - 实现: `internal/services/cascade.go` + `handlers/organization.go`

- `/api/v1/organization-units/batch-refresh-hierarchy`
  - 中文: 批量层级修复（迁移/修复场景）
  - EN: Batch hierarchy refresh (migration/repair)
  - 实现: `internal/services/cascade.go`

- `/api/v1/corehr/organizations`
  - 中文: CoreHR 兼容层端点（受控暴露）
  - EN: CoreHR compatibility endpoint (controlled exposure)
  - 实现: `handlers/organization.go`

### 系统管理端点
- `/health` - 健康检查 → `internal/handlers/operational.go: GetHealth`
- `/metrics` - Prometheus指标 → `operational.go: GetMetrics`
- `/alerts` - 系统告警 → `operational.go: GetAlerts`
- `/tasks` - 任务状态 → `operational.go: GetTasks`
- `/tasks/{id}/status` - 任务状态查询 → `operational.go: GetTaskStatus`
- `/tasks/{id}/trigger` - 触发任务 → `operational.go: TriggerTask`
- `/operational/cutover` - 触发切换 → `operational.go: TriggerCutover`
- `/operational/consistency-check` - 一致性检查 → `operational.go: TriggerConsistencyCheck`

### 开发工具端点 (仅DEV模式)
- `/auth/dev-token` - 生成开发令牌 → `internal/handlers/devtools.go: GenerateDevToken`
- `/auth/dev-token/info` - 令牌信息 → `devtools.go: GetTokenInfo`
- `/dev/status` - 开发状态 → `devtools.go: DevStatus`
- `/dev/test-endpoints` - 测试端点列表 → `devtools.go: ListTestEndpoints`
- `/dev/database-status` - 数据库状态 → `devtools.go: DatabaseStatus`
- `/dev/performance-metrics` - 性能指标 → `devtools.go: PerformanceMetrics`
- `/dev/test-api` - API测试工具 → `devtools.go: TestAPI`

---

## GraphQL 查询 API（Query Service, Port 8090）
权威规范: `docs/api/schema.graphql`

> 说明: 基于实际Schema文件扫描的查询字段清单，严格遵循CQRS架构

### 核心查询字段 (12个)
- `organizations(filter, pagination): OrganizationConnection!`
  - 中文: 组织分页列表（过滤/时态支持）
  - EN: Paginated organizations with filters and temporal support
  - 实现: PostgreSQL原生查询，利用时态索引优化

- `organization(code, asOfDate): Organization`
  - 中文: 按业务编码查询单个组织（支持 asOfDate）
  - EN: Fetch organization by business code (with asOfDate)
  - 实现: 时态点查询，复合主键 (code, effective_date) 优化

- `organizationStats(asOfDate, includeHistorical): OrganizationStats!`
  - 中文: 组织统计（时态维度统计）
  - EN: Organization statistics with temporal breakdown
  - 字段: `totalCount, temporalStats, byType.unitType, oldestEffectiveDate, newestEffectiveDate`

- `organizationHierarchy(code, tenantId): OrganizationHierarchy`
  - 中文: 完整层级信息（路径、关系、属性）
  - EN: Complete hierarchy info with paths and relations
  - 实现: 层级路径查询，利用 `code_path` 索引

### GraphQL Schema实际字段扫描
基于 `docs/api/schema.graphql` 文件识别的查询字段：
- `organizations` - 组织列表查询
- `filter` - 查询过滤器
- `pagination` - 分页参数
- `organization` - 单个组织查询
- `code` - 组织编码参数
- `asOfDate` - 时态查询时间点
- `organizationStats` - 统计信息查询
- `includeHistorical` - 包含历史数据标志
- `organizationHierarchy` - 层级结构查询
- `tenantId` - 租户ID参数

### 实现架构说明
- **PostgreSQL原生**: 直接查询PostgreSQL，无中间数据同步层
- **时态优化**: 26个专用时态索引，查询响应时间1.5-8ms
- **CQRS严格分离**: 查询专用GraphQL端点，与REST命令端点完全分离
- **统一认证**: JWT/OAuth校验，tenant-aware查询

---

## 后端（Go）关键导出（Key Exported Items）

### 处理器（Handlers） - 26个导出方法
基于实际代码扫描结果：

#### 组织业务处理器 (`organization.go`)
- `SetupRoutes` - 路由设置
- `CreateOrganization` - 创建组织单元
- `CreateOrganizationVersion` - 创建时态版本
- `UpdateOrganization` - 更新组织信息
- `SuspendOrganization` - 暂停组织
- `ActivateOrganization` - 激活组织
- `CreateOrganizationEvent` - 创建组织事件
- `UpdateHistoryRecord` - 更新历史记录

#### 运维管理处理器 (`operational.go`)
- `SetupRoutes` - 运维路由设置
- `GetHealth` - 系统健康检查
- `GetMetrics` - Prometheus指标收集
- `GetAlerts` - 系统告警查询
- `GetTasks` - 任务列表查询
- `GetTaskStatus` - 任务状态查询
- `TriggerTask` - 触发任务执行
- `TriggerCutover` - 触发系统切换
- `TriggerConsistencyCheck` - 触发一致性检查

#### 开发工具处理器 (`devtools.go`) - 仅DEV模式
- `SetupRoutes` - 开发工具路由
- `GenerateDevToken` - 生成开发JWT令牌
- `GetTokenInfo` - 获取令牌信息
- `DevStatus` - 开发环境状态
- `ListTestEndpoints` - 测试端点列表
- `DatabaseStatus` - 数据库连接状态
- `PerformanceMetrics` - 性能指标监控
- `TestAPI` - API测试工具

### 服务层（Services） - 14个导出类型
#### 级联更新服务 (`cascade.go`)
- `CascadeUpdateService` - 层级变更级联处理
- `CascadeTask` - 级联任务定义

#### 运维调度服务 (`operational_scheduler.go`)
- `OperationalScheduler` - 后台任务调度器
- `ScheduledTask` - 调度任务结构

#### 时态数据服务 (`temporal.go`)
- `TemporalService` - 时态版本管理核心服务
- `InsertVersionRequest` - 插入版本请求
- `OrganizationData` - 组织数据结构
- `DeleteVersionRequest` - 删除版本请求
- `ChangeEffectiveDateRequest` - 变更生效日期请求
- `SuspendActivateRequest` - 暂停/激活请求
- `VersionResponse` - 版本操作响应

#### 时态监控服务 (`temporal_monitor.go`)
- `TemporalMonitor` - 时态数据质量监控
- `MonitoringMetrics` - 监控指标收集
- `AlertRule` - 告警规则定义

### 架构特点
- **CQRS分离**: 命令服务(9090端口)与查询服务(8090端口)完全分离
- **PostgreSQL原生**: 直接操作PostgreSQL，无中间数据同步
- **时态数据**: 完整的时态版本管理和监控体系
- **企业级监控**: 完备的健康检查、指标收集、告警机制
- **开发友好**: 丰富的开发工具和调试端点

---

## 前端（TypeScript/React）关键导出（Key Exported Items）

基于实际代码扫描的120+个导出项分类整理：

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

### 数据管理层
#### 状态管理Hooks
- `useOrganizations` - 组织列表管理 (`useOrganizations.ts`)
- `useOrganization` - 单个组织管理
- `useEnterpriseOrganizations` - 企业级组织管理 (`useEnterpriseOrganizations.ts`)
- `useOrganizationList` - 组织列表复用Hook
- `useMessages` - 用户消息管理 (`useMessages.ts`)

#### 组织变更操作 (`useOrganizationMutations.ts`)
- `useCreateOrganization` - 创建组织Hook
- `useUpdateOrganization` - 更新组织Hook
- `useSuspendOrganization` - 暂停组织Hook
- `useActivateOrganization` - 激活组织Hook

#### 时态数据管理 (`useTemporalAPI.ts`)
- `TemporalAPIError` - 时态API错误类
- `useTemporalHealth` - 时态服务健康检查
- `useTemporalAsOfDateQuery` - 时间点查询Hook
- `useTemporalDateRangeQuery` - 时间范围查询Hook
- `useTemporalQueryUtils` - 时态查询工具Hook
- `useTemporalQueryStats` - 时态查询统计Hook
- `TemporalDateUtils` - 时态日期工具类

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
- `TemporalConverter` - 时态转换器类
- `TemporalUtils` - 时态工具函数

### 验证系统
#### Schema验证 (`schemas.ts`)
- `OrganizationUnitSchema` - 组织单元Schema
- `CreateOrganizationInputSchema` - 创建输入Schema
- `CreateOrganizationResponseSchema` - 创建响应Schema
- `UpdateOrganizationInputSchema` - 更新输入Schema
- `GraphQLVariablesSchema` - GraphQL变量Schema
- `GraphQLOrganizationResponseSchema` - GraphQL组织响应Schema

#### 简单验证 (`simple-validation.ts`)
- `SimpleValidationError` - 简单验证错误类
- `validateOrganizationBasic` - 组织基础验证
- `validateOrganizationUpdate` - 组织更新验证
- `validateOrganizationResponse` - 组织响应验证
- `formatValidationErrors` - 格式化验证错误
- `getFieldError` - 获取字段错误
- `validateStatusUpdate` - 状态更新验证
- `basicValidation` - 基础验证函数
- `safeTransform` - 安全转换函数
- `validateCreateOrganizationInput` - 验证创建输入
- `validateUpdateOrganizationInput` - 验证更新输入

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

### 时态验证工具 (`temporalValidation.ts`)
- `validateTemporalDate` - 时态日期验证

---

## 运维与脚本（DevOps & Scripts）

### 质量保证脚本
- `scripts/generate-implementation-inventory.js` - **实现清单生成器** (避免重复造轮子)
- `scripts/quality/duplicate-detection.sh` - 重复代码检测工具
- `scripts/quality/architecture-validator.js` - 架构一致性验证
- `scripts/quality/document-sync.js` - 文档同步监控

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

## 统计摘要 📊

### **实现规模统计**
- **REST API端点**: 10个核心业务 + 8个系统管理 + 7个开发工具 = **25个端点**
- **GraphQL查询**: 12个查询字段 + 完整Schema支持
- **Go后端导出**: 26个处理器方法 + 14个服务类型 = **40个关键组件**
- **前端导出**: 120+个导出项，涵盖API、Hooks、工具、配置、验证等
- **脚本工具**: 20+个开发、质量保证、CI/CD脚本

### **架构成熟度**
- ✅ **CQRS架构**: 查询/命令完全分离
- ✅ **PostgreSQL原生**: 单一数据源，性能优化
- ✅ **企业级监控**: 健康检查、指标收集、告警系统
- ✅ **质量门禁**: 契约测试、重复代码检测、架构验证
- ✅ **开发工具**: JWT管理、API测试、性能监控

---

## 变更记录（Changelog）
- **v1.0 生产就绪版（2025-09-09）**: 基于实际代码扫描的完整清单
  - 新增: 120+个前端导出项详细分类
  - 新增: 26个Go处理器和14个服务类型
  - 新增: 重复造轮子风险分析和防范指导
  - 新增: 统计摘要和架构成熟度评估
- v0.1 初稿（2025-09-09）: 建立单文件清单框架

