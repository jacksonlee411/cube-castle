# 56号文档：前端配置与常量质量分析

## 背景与唯一事实来源
- 本文覆盖 `frontend/src/shared/config/constants.ts`、`environment.ts`、`ports.ts`、`tenant.ts`，`frontend/tests/config/ports.ts`、`frontend/tests/e2e/config/test-environment.ts`，`frontend/src/design-system/tokens/brand.ts`，`frontend/src/features/organizations/constants/formConfig.ts`、`tableConfig.ts`，`frontend/src/features/temporal/constants/temporalStatus.ts`、`index.ts`，以及文档提及的 `frontend/src/shared/utils/constants.ts`（当前仓库中不存在）。所有结论直接源于现有源码或该缺失文件的实际状态。
- 代码审查遵循“资源唯一性与跨层一致性”为最高约束，必要时对照后端服务层实现（如组织层级限制）验证配置一致性。

## 代码质量评估
- `constants.ts`：分类明确但常量粒度极细、夹杂大量日志与 emoji，难以追踪真实用例；部分值（如组织层级上限）与后端不一致。`generateConstantsReport` 只统计一层键数量，报告信息有限。
- `environment.ts`：封装即用配置对象，缺少对敏感字段的防泄漏约束，`clientSecret` 等静态默认值直接暴露在前端 bundle 中；验证函数仅覆盖两个字段，缺少对 GraphQL 端点等关键配置的校验。
- `ports.ts`：端口集中管理但 `buildServiceURL` 依赖 `process.env.*`，在 Vite/浏览器环境中为空。端口校验逻辑未纳入 `FRONTEND_DEV` (<1024) 场景；`CQRS_ENDPOINTS` 与测试端代码产生重复拼接。
- `tenant.ts`：单例初始化时直接访问 `localStorage`，在 SSR、Vitest 环境会抛出 `ReferenceError`；`getTenantConfig` 返回 `"Unknown Tenant"`，缺少对后端查询能力的集成；UUID 校验仅覆盖 1-5 版本。
- 测试配置：`frontend/tests/config/ports.ts` 与 `frontend/tests/e2e/config/test-environment.ts` 均实现 `checkPortAvailability`，逻辑重复；E2E 配置文件大量使用 `fetch` + `AbortSignal.timeout`，在较旧 Node 环境不兼容。
- 设计令牌与业务常量：`brand.ts` 仅导出裸色值，与组件使用的语义令牌未对齐；组织表单/表格常量与 `shared/config/constants` 中同类型值重复维护，并存在取值矛盾（表单层级上限为 10，而后端和服务层逻辑允许 17）。
- 时态模块：`temporalStatus.ts` 同时维护 `TEMPORAL_STATUS_COLORS` 与 `temporalStatusUtils.getStatusColor`（后者使用硬编码 hex）；`calculateStatus` 永远不会返回 `EXPIRED`，与常量枚举不符；`temporal/index.ts` 内部 `temporalUtils.getTemporalStatus` 返回 `future/expired/active` 等非枚举值，导致调用者需要额外映射。
- 缺失文件：文档引用的 `frontend/src/shared/utils/constants.ts` 实际不存在，会导致开发者查阅失败或误以为文件被误删。

## 主要问题
### 跨层数据不一致
- `BUSINESS_CONSTANTS.ORG_LEVEL_MAX`、`formConfig.ORGANIZATION_LEVELS.MAX` 均为 10，而后端 `TemporalService`、`CascadeUpdateService` 和 GraphQL 查询允许 17 级；表单发送的数据将被后端拒绝或导致 UI 缺少层级选项。
- `FEATURE_FLAGS` 与 `TemporalStatus` 相关颜色定义与设计系统令牌脱节，使 UI 难以保持统一；`temporalUtils` 返回的状态值与 `TEMPORAL_STATUS_OPTIONS` 不同，前端内部状态映射产生分叉。

### 可用性与容错不足
- `TenantManager` 构造函数内直接访问 `localStorage` 与 `import.meta.env`，在 Vitest、SSR 或 Storybook 环境均无保护；`getTenantIdFromToken` 始终返回 `null`，却缺乏钩子或注入点。
- `buildServiceURL` 运行在浏览器时 `process.env.SERVICE_HOST` 为 `undefined`，最终拼接 `http://undefined:9090`，导致生产构建后的测试用例失效。
- `checkPortAvailability` 的 `fetch` 调用缺少兜底超时（除 E2E 版本外），若目标端口关闭会等待默认超时；重复实现难以统一增强。

### 维护成本高
- 常量定义散落在 `shared/config`、`features/*/constants`、`temporal/index` 等多个位置互相引用，缺少清晰的事实来源；生成报告等工具函数未纳入测试。
- Emoji 日志与报告字符串在编译输出中增加非必要体积，也会干扰集中式日志搜集。

## 过度设计
- `constants.ts`、`ports.ts` 引入“报告生成器”“硬编码消除率”一类展示性逻辑，但实际未在 UI 中调用，属于额外负担。
- `TenantManager` 以单例 + `localStorage` 方式实现租户切换，却没有实际的多租户接入或订阅机制；与普通的 `useState + context` 相比复杂度更高。
- E2E `discoverActivePort`/`validateTestEnvironment` 同时处理动态发现与日志输出，多处 `console.log`/`console.warn`，可由 Playwright 全局前置脚本或测试框架配置更简单地完成。

## 重复造轮子
- 时间/重试常量分别定义在 `constants.ts` 和 `temporal/index.ts`（`TEMPORAL_CONSTANTS`），还与 React Query 配置、服务端响应超时等重复；未复用既有 util。
- 端口检测函数在单元测试与 E2E 配置中重复，实现和常量来源略有差异，增加同步成本。
- 颜色与状态映射在 `brand.ts`、`tableConfig.ts`、`temporalStatus.ts` 多次维护，缺乏统一的设计令牌或主题系统对接。

## 改进建议
1. **统一事实来源**：将组织层级、状态枚举、分页等通用常量集中到 `shared/config/constants.ts`，其他文件通过导出别名使用；同步调整上限为后端约定的 17，并添加跨层快照测试。
2. **强化环境适配**：在 `TenantManager`、`buildServiceURL` 等模块中检测运行环境（浏览器 / SSR / Node），对 `localStorage`、`process.env` 提供安全访问包装，并支持配置注入以便测试。
3. **精简过度设计**：移除或下沉 `generateConstantsReport`、`generatePortConfigReport` 等报告型工具到开发脚本；在运行时代码中仅保留必要常量与校验，削减 bundle 体积与噪音日志。
4. **复用设计令牌**：统一使用 `cubecastleBrandTokens` 的语义颜色，并为状态颜色提供从令牌到组件的映射表；移除硬编码 hex。
5. **整合测试配置**：在 `frontend/tests/` 内建立共享的 `ports`/`environment` helper，将 `checkPortAvailability`、`discoverActivePort` 等逻辑合并，提供明确的超时与错误处理策略，同时利用 `node:net` 端口探测替代 `fetch` + 跨环境兼容问题。
6. **补充缺失文件说明**：更新相关文档或脚本，说明 `shared/utils/constants.ts` 已被 `shared/config/constants.ts` 取代，防止开发者误解；若确有需求，可新增文件 re-export 唯一事实来源。

## 验收标准
- [ ] 组织层级、状态、分页等配置在前后端保持一致（含自动化校验），`formConfig`/`tableConfig` 不再硬编码冲突值。
- [ ] `TenantManager`、`buildServiceURL` 在无 `localStorage` 或 `process.env` 的环境下可安全运行，并附带单元测试覆盖。
- [ ] 端口与环境检测辅助方法统一到单一模块，测试/E2E 共用，并具备超时、错误日志与类型定义。
- [ ] 颜色与状态常量引用设计系统令牌，`temporalStatusUtils` 返回值与 `TemporalStatus` 枚举一致，新增测试确保映射完整。
- [ ] 运行时代码移除开发报告型函数或隔离到仅在开发脚本中使用的模块，确保 bundle 体积与日志整洁。

## 一致性校验说明
- 本文结论已与后端服务层（组织层级深度、状态枚举）、设计系统令牌对照，确认存在跨层分歧；整改需同步更新相关契约文档或测试。
- 文档存放于 `docs/development-plans/`，完成整改后请归档至 `docs/archive/development-plans/` 并在验收报告中记录跨层一致性检查结果。***
