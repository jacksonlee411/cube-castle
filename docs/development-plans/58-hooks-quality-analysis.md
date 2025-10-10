# 58号文档：Hooks质量分析

## 背景与唯一事实来源
- 本文聚焦以下前端 Hook 实现：
  - `frontend/src/shared/hooks/useEnterpriseOrganizations.ts`
  - `frontend/src/shared/hooks/useMessages.ts`
  - `frontend/src/shared/hooks/useOrganizationMutations.ts`
- 结论仅来源于上述源码及同一目录已有的类型/客户端定义（例如 `frontend/src/shared/types/` 与 `frontend/src/shared/api/unified-client.ts`），已核对未引用其它事实来源，保持前后端契约一致性。

## 代码质量评估
- **useEnterpriseOrganizations**：集中式状态对象可读性尚可，但内部手写 GraphQL 查询、变量映射与 `setState` 更新逻辑超过 200 行，缺少请求取消与竞态保护；统计结构 `OrganizationStats` 仅保留 `total/active/inactive`，与查询结果中的 `plannedCount/deletedCount/byType` 等字段脱节。
- **useMessages**：接口简洁，但 `showSuccess/showError` 使用的 `setTimeout` 没有保存与清理句柄，重新触发时旧定时器仍会清空最新消息，卸载组件后也存在悬挂定时器。
- **useOrganizationMutations**：Mutation 定义覆盖常见操作，但 `onSettled` 同时调用 `invalidateQueries` 与 `refetchQueries`，并辅以大量 `removeQueries`，造成冗余网络请求；`ensureSuccess`、ETag 处理、变量类型等重复逻辑散落在单个文件内，维护成本高。

## 主要问题
### 数据获取与状态管理缺陷
- `useEnterpriseOrganizations` 每次 `fetchOrganizations`/`fetchStats` 都通过动态导入获取 `unifiedGraphQLClient` 并直接 `setState`，未使用 `react-query` 等既有缓存机制，也没有处理请求竞态或组件卸载，存在状态闪烁与内存泄漏风险。
- `initialParams` 被直接放入 `useEffect` 依赖，一旦调用方以字面量对象传参会触发无限重拉；`refreshData` 简单串行调用两次网络请求，没有复用已有 Promise。
- 统计信息将 `inactive` 计算为 `historicalCount + futureCount`，忽略返回中的 `inactiveCount` 字段，导致数据失真。

### 错误处理与契约一致性薄弱
- `useEnterpriseOrganizations` 自行拼装 `APIResponse`，错误分支仅返回 `GRAPHQL_ERROR` 或 `HOOK_ERROR` 字符串，未复用 `shared/api/error-handling.ts` 中的统一错误语义，也未记录 `requestId`。
- `useOrganizationMutations` 将 `OrganizationRequest`（内部 `code` 可选）传给 `PUT /organization-units/${data.code}`，随后在 `onSettled` 中使用 `variables.code!`，若调用方漏传将导致运行时异常；`EnsureSuccess` 抛出的错误未与全局 `ErrorHandler` 对齐。

### Hook 内部状态处理与副作用缺口
- `useMessages` 每次调用 `showSuccess` 或 `showError` 都创建新的 `setTimeout`，没有存储句柄或在组件卸载时清理，容易出现过期计时器影响后续消息；同时缺乏“立即清除仍保留定时器”逻辑，导致短时间多次调用时消息闪烁。
- `useOrganizationMutations` 在 `onSettled` 中既 `invalidate` 又手动 `refetch` 同一个 `queryKey`，并重复对 `organizations`、`organization-stats` 执行“失效+立即 refetch”，破坏 React Query 默认的去抖策略，加重后端负载。
- 多处手动调用 `removeQueries` 清理 inactive 缓存，但没有配合乐观更新或缓存时间设置，极易和其他页面共享的查询互相打架。

## 过度设计分析
- `useEnterpriseOrganizations` 将列表查询、单个查询、统计、错误清除等功能全部塞入一个 Hook，且内置企业级响应封装，导致调用方难以挑选所需能力；与 `frontend/src/shared/api/graphql-enterprise-adapter.ts` 已提供的企业级适配重复。
- `useOrganizationMutations` 在 `onSettled` 中大篇幅日志、缓存失效、即时 `setQueryData`、删除 inactive 查询等“全功能”流程，对于简单的创建或状态切换是过度设计，亦让测试复杂度飙升。

## 重复造轮子情况
- `useEnterpriseOrganizations` 手写 GraphQL 查询与企业级响应包装，未复用现有的 `graphqlEnterpriseAdapter` 或集中定义的查询片段；同样的过滤/分页参数映射在多个模块已实现。
- `ensureSuccess`、`formatIfMatchHeader`、`OrganizationStateMutationVariables` 与 `DeleteOrganizationVariables` 在 `useOrganizationMutations` 内重复实现，功能上与统一 API 错误处理、ETag 归一化和通用操作请求类型重叠。
- `useMessages` 自行管理消息定时清除，与项目内的 `MessageDisplay` 组件及 Canvas Kit 提供的通知模式重叠，缺乏与全局通知中心的对接。

## 综合改进建议
1. **拆分与标准化查询 Hook**：将组织列表、单项查询与统计拆分为基于 React Query 的独立 Hook，复用 `graphqlEnterpriseAdapter` 和统一的查询 key 常量，消除手写状态与重复动态导入。
2. **对齐错误处理与契约**：在查询/Mutation 中统一使用 `ErrorHandler`，补全 `requestId`、`details`、租户错误等信息，同时确保 `OrganizationRequest` 在更新/删除路径上的 `code` 为必填。
3. **收敛缓存与副作用逻辑**：审视 `onSettled` 中的 `invalidate`、`refetch`、`removeQueries` 组合，保留单一策略（推荐仅 `invalidateQueries` 并让 React Query 管理重新获取），必要时通过 `queryClient.setQueryData` 做受控更新。
4. **重构消息 Hook**：引入 `useRef` 储存定时器句柄并在 `useEffect` 清理，提供可配置的自动消失时间或无自动清除选项，避免旧计时器覆盖新消息。
5. **抽取通用工具**：将 ETag 格式化、Idempotency-Key 注入、默认删除原因等工具函数迁移到 `shared/api` 或 `shared/utils`，供多个 Mutation 共享，减少重复代码。

## 验收标准
- [ ] 组织相关查询 Hook 改用统一的 GraphQL 适配器与 React Query 缓存，并具备请求取消/竞态保护。
- [ ] Mutation `onSettled` 逻辑精简为一致策略，实际网络调用次数与缓存行为符合 React Query 推荐模式。
- [ ] 消息 Hook 清理定时器且支持配置化清除，避免重复触发导致的闪烁或内存泄漏。
- [ ] 统一错误处理输出包含契约要求的 `code/message/requestId` 字段，用例覆盖空数据、权限错误等关键路径。
- [ ] 重复工具与类型抽出为共用模块，Mutation 与查询 Hook 均引用同一实现。

## 一致性校验说明
- 已确认本文涉及的 Hook 与 `frontend/src/shared/api`、`frontend/src/shared/types` 中的契约定义一致，未引入额外事实来源。
- 后续整改时需要再次对照 `docs/api/schema.graphql` 与 `docs/api/openapi.yaml`，确保 GraphQL 查询字段、REST Mutation 路径与契约保持一致，整改完成后请按流程归档本计划。

