# 57号文档：API 与类型质量分析

## 背景与唯一事实来源
- 本文覆盖以下前端共享模块源码：
  - `frontend/src/shared/api/auth.ts`
  - `frontend/src/shared/api/contract-testing.ts`
  - `frontend/src/shared/api/error-handling.ts`
  - `frontend/src/shared/api/error-messages.ts`
  - `frontend/src/shared/api/graphql-enterprise-adapter.ts`
  - `frontend/src/shared/api/type-guards.ts`
  - `frontend/src/shared/api/unified-client.ts`
  - `frontend/src/shared/types/api.ts`
  - `frontend/src/shared/types/converters.ts`
- 复核过程中确认 `frontend/src/shared/types/type-guards.ts` 在当前仓库不存在，已记录为一致性风险；未引用外部资料，保持唯一事实来源。
- 其余结论均直接来源于上述源码，未引入第二事实来源。

## 代码质量评估
- `auth.ts`：`AuthManager` 同时承担令牌刷新、环境模式切换、JWKS 校验与本地存储迁移，逻辑庞杂；每次取 Token 前强制访问 `/.well-known/jwks.json`，在未部署 JWKS 的开发环境会直接抛错；`mapKeysToSnakeCase` 仅处理浅层键，未覆盖嵌套字段。
- `contract-testing.ts`：导出的 API 仅通过 `setTimeout` 返回硬编码成功结果，与文档要求的契约校验无关，易导致测试误判。
- `error-handling.ts`：`withErrorHandling`、`withRetry`、`withOAuthRetry` 等高阶函数把入参类型限制为 `RawError[]`，与实际 API 函数签名不符，TypeScript 无法在真实调用处通过编译；`ErrorHandler.handleAPIError` 始终抛出包装后的错误，后续 `throw error` 成为死代码。
- `error-messages.ts`：错误映射结构完整，但 `OPERATION_REASON_REQUIRED` 的用户文案与 `technicalMessage` 互相矛盾；`isRecord` 与 `types/json.ts` 的 `isJsonObject` 重复实现。
- `graphql-enterprise-adapter.ts`：纯粹包裹 `UnifiedGraphQLClient` 并随机生成 `requestId`，返回的“企业级信封”并未与后端协商字段，且占位注释大量未实现逻辑。
- `type-guards.ts`：将 Zod 校验、错误类型守卫、GraphQL 转换混在单文件，`safeTransformGraphQLToOrganizationUnit` 与 `types/converters.ts` 中的转换函数重复；`isNetworkError` 只匹配 `TypeError + includes('fetch')`，对 Axios/Node Fetch 场景无效。
- `unified-client.ts`：GraphQL 与 REST 客户端各自实现 401/403/500 处理、JSON 解析与重试机制，未复用 `error-handling.ts` 的统一逻辑；`validateCQRSUsage` 仅靠字符串 `includes('GraphQL')`/`'REST'` 判定，难以覆盖真实调用。
- `types/api.ts`：在纯类型文件中混入 `isGraphQLResponse`、`hasGraphQLErrors` 等运行时函数，与 `api/type-guards.ts` 的守卫定义重复，削弱“类型真源”定位。
- `types/converters.ts`：提供 GraphQL/REST 转换、类型一致性检查、动态类型生成等多功能集合，但核心转换与 `api/type-guards.ts` 重叠；日志中大量 emoji，且 `generateTypeDefinition`/`logTypeSyncReport` 无直接调用方。
- `frontend/src/shared/types/type-guards.ts`：文件缺失但仍被计划文档引用，说明类型目录的真源定义与文档脱节。

## 主要问题
### 认证与客户端耦合不当
- `AuthManager` 在获取令牌前无条件调用 `ensureRS256`，若命令服务未启动 JWKS（常见于本地与 CI），所有前端请求都会失败；统一客户端本应允许在开发模式降级使用 HS256/模拟令牌。
- GraphQL/REST 客户端重复实现 401 重试与租户头注入，却未复用统一错误处理模块，导致认证失败无法统一上报，也增加后续维护复杂度。

### 契约校验与类型真源失真
- 契约测试接口被硬编码为“全部通过”，无法反映 `docs/api/openapi.yaml` 与 `schema.graphql` 的真实偏差，违背“资源唯一性”原则。
- 类型体系存在缺口：`types/api.ts`、`api/type-guards.ts`、`types/converters.ts` 同时维护 GraphQL/REST 转换，且计划文档仍引用不存在的 `types/type-guards.ts`，真源类型与实现脱离。

### 错误处理系统可用性不足
- `withErrorHandling`/`withRetry` 等包装器的泛型约束错误，真实业务函数无法通过类型检查，导致调用方绕过统一错误治理。
- `ErrorHandler`、`UnifiedErrorHandler`、客户端内联错误处理三套体系并存，既无统一格式也未输出结构化日志，告警与可观察性难以落地。

### 企业级 GraphQL 适配流于形式
- `GraphQLEnterpriseAdapter` 在前端伪造 `requestId` 并构造成功/失败信封，与命令服务的真实结构无关；批量接口直接复用单个请求结果，未对失败分支进行容错，易掩盖后端错误。

## 过度设计分析
- `AuthManager` 同时处理 RS256 校验、legacy localStorage 清理、HS256 迁移提示等历史分支，缺乏与运行模式解耦的配置；在现代 RS256-only 策略下可下沉到专用服务或 CLI。
- 错误处理模块引入多套“统一”/“感知 OAuth”的包装器，但由于泛型设计错误几乎无人使用，形成纸面上的复杂架构。
- `types/converters.ts` 的类型同步报告、动态接口生成、路径字段兼容等功能在当前界面中无调用者，导致编辑负担远高于实际价值。
- `GraphQLEnterpriseAdapter` 预留大量注释型 TODO，却没有明确的后端契约支持；相较直接消费 GraphQL 响应，额外封装只带来更多维护负担。

## 重复造轮子情况
- `mapKeysToSnakeCase`、`camelToSnakeCase` 再次实现大小写转换，项目已有 `shared/utils/string` 中的格式化工具可供复用。
- 认证客户端自建 `ensureRS256`+`decodeBase64Url`，而项目已引入 `jose` 库，可直接使用成熟的 JWT 解析与算法校验。
- GraphQL/REST 客户端手写 401/403 重试与 JSON 解析，与 `error-handling.ts`、`withOAuthRetry`、`withErrorHandling` 的目标一致，却因接口不兼容而各自为政。
- `safeTransformGraphQLToOrganizationUnit` 与 `convertGraphQLToOrganizationUnit`、`convertCreateInputToREST` 等函数实现重复字段映射，维护成本翻倍。

## 综合改进建议
1. **简化认证流程**：为 `AuthManager` 注入配置驱动的 JWKS 校验开关，或在开发模式降级为本地令牌；将请求签名校验下沉至网关或后端统一服务，前端仅关注获取/刷新流程。
2. **收敛错误处理**：修复 `withErrorHandling`/`withRetry` 的泛型定义（参数应为 `unknown[]`），并在统一客户端调用链复用这些包装器，输出结构化日志而非 emoji。
3. **恢复契约测试真源**：替换 `contract-testing.ts` 的模拟实现，改为调用实际的契约校验端点或运行 npm 脚本；同时在文档中声明唯一事实来源。
4. **统一类型与转换**：在 `shared/types` 目录重新定义真源模块，合并 `type-guards` 与转换工具的职责，移除缺失文件引用，并补齐导出图谱。
5. **评估 GraphQL 适配价值**：若后端尚未提供企业级信封，暂时直接返回 `UnifiedGraphQLClient` 的结果，待契约明确后再引入适配层，避免前端伪造请求 ID。

## 验收标准
- [ ] `AuthManager` 支持在本地/CI 环境下跳过 JWKS 远程校验，GraphQL/REST 客户端统一复用错误处理包装器并输出结构化日志。
- [ ] 契约测试接口实际执行后端校验或本地脚本，拨除硬编码结果，并在失败时返回精准信息。
- [ ] 错误处理高阶函数泛型修复，已有 API 调用链通过编译并能正确捕获/重抛 `UserFriendlyError`。
- [ ] `shared/types` 目录重新对齐真源结构，删除缺失文件引用，合并重复转换逻辑并补充单元测试覆盖关键字段映射。
- [ ] GraphQL 适配层根据实际契约决定是否保留；若继续使用，`requestId`、`timestamp` 等字段需来源于服务端。

## 一致性校验说明
- 本文所有结论均来自上述源码，未引用外部口头或二手资料；确认与现有 53~56 号文档无重复分析。
- 整改落地后需按流程将本计划归档至 `docs/archive/development-plans/`，提交信息引用“57号文档”，以保持计划-实现一致性链路。
