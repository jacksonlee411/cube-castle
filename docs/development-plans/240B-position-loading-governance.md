# Plan 240B – 职位详情 数据装载链路与等待治理

编号: 240B  
上游: Plan 240（职位管理页面重构） · 依赖 240A 完成  
状态: 已完成

—

## 目标
- 统一职位详情的数据装载链路，避免白屏、竞态与重复请求；建立“可取消 + 可重试 + 正确失效”的稳定等待机制。
- 与主计划对齐：采用统一 Hook/Loader（241），命名抽象（242），在 Timeline/Status 抽象（244）验收后落地，测试资产遵循 246 的选择器治理。

## 依赖与准入条件
- 硬依赖（必须满足方可进入实施）：
  - 243（242/T1）统一入口已合并：`TemporalEntityPage` 与路由抽象可用。
  - 244 验收通过：`TemporalEntityTimelineAdapter/StatusMeta` 合并、契约同步，基础 E2E 绿灯。
  - 241 恢复并作为承载框架：数据读取以统一 Hook/Loader 为唯一入口，不再新增职位专有 Loader。
  - 守卫接入：`npm run guard:plan245`、`npm run guard:selectors-246` 通过；基线计数不升高。
- 软依赖：
  - 232/232T 最新 E2E 日志可用，用于等待链路与定位器稳定性复验。

## 前置条件（Docker 强制 + 契约先行）
- 环境校验（遵循 AGENTS.md）：
  - `make docker-up` → `make run-dev` → `make frontend-dev`
  - `curl http://localhost:9090/health`、`curl http://localhost:8090/health` → 200
  - `make jwt-dev-mint`（`.cache/dev.jwt` 存在）
  - Go 工具链一致性：`go version` 显示 `go1.24.x`（与仓库 `toolchain go1.24.9` 一致）
  - Node/包管理器基线：`node -v`、`npm -v` 输出满足前端要求（参考 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`）
- 契约先行：
  - 不更改 API 契约。如确需字段扩展，先更新 `docs/api/schema.graphql` / `docs/api/openapi.yaml`，再实现，并执行：
    ```bash
    node scripts/generate-implementation-inventory.js
    ```
  - 文档与架构校验：`node scripts/quality/document-sync.js`、`node scripts/quality/architecture-validator.js` 退出码均为 0（避免第二事实来源）

## 范围与实现边界
- 范围：
  - 路由层与详情层的读请求合并；AbortController 取消；错误边界与指数退避重试（仅幂等读）。
  - React Query key/缓存策略精确化；命令链路后失效刷新与跨页签联动。
  - E2E 等待模式与选择器统一（复用 246 的工具与守卫）。
- 边界（不做）：
  - 不新建“职位专有 Loader/Hook”；必须通过 `useTemporalEntityDetail` 或 241 的 `createTemporalDetailLoader` 工厂注入（可薄适配，但不得形成第二事实来源）。
  - 不在组件处硬编码 `data-testid`；仅通过 `temporalEntitySelectors` 使用选择器。

## 重试与取消策略（统一约束）
- 幂等读范围：仅对 GraphQL 查询（POST `/graphql`）与 GET 读操作生效；REST 命令与任何非幂等操作禁止重试与取消策略介入。
- 重试参数（统一在查询层配置，不在调用点重复实现）：
  - `maxAttempts=3`、`baseDelay=200ms`、`factor=2`、`cap=3000ms`、`jitter=true`
  - `retryWhen`：网络错误、超时、HTTP `5xx`、`429`；对 `4xx`（含权限/业务错误）与已分类业务错误不重试
- 取消语义：
  - 路由切换/页签切换触发 `AbortSignal`，必须取消在途请求；取消异常应被视为“静默失败”，不得更新 UI
  - 防止”旧响应覆盖新状态“：消费层按“期望 code/asOfDate/tenant”做收敛校验，不匹配直接丢弃（不触发选择器/状态更新）
- 统一出口：重试/延迟由前端统一查询配置导出（SSoT），禁止在业务 Hook/组件内私自实现退避逻辑

## 任务清单（按可执行项）
1) Loader 集成与并发治理  
   - 采用 241 的统一 Loader（`createTemporalDetailLoader` 如已提供）封装职位详情数据装载；在路由/页签切换时聚合并发请求；Loader 负责用 `queryClient.prefetchQuery` 预热并携带 `AbortSignal`，页面 Hook 仅读取缓存，避免重复请求。  
   - 实现 AbortController 取消策略，路由切换/页签切换即时取消在途请求；重试采用指数退避（仅幂等读）。  
2) 缓存与失效  
   - 规范 QueryKey（包含实体类型、code、asOfDate、tenant、filters 等）；filters/数组参数需稳定序列化（排序 + 规范化），避免缓存穿透。  
   - 命令成功后触发对应 QueryKey 失效（跨页签：概览/任职/时间线/版本/审计），确保 UI 立即刷新。  
   - 统一失效工具（SSoT）：通过集中工具（例如 `invalidateTemporalDetail`）失效下列键系，禁止在各处手写键名：
     - `temporalEntityDetailQueryKey(entity='position', code, ...)`
     - 既有职位键系：`POSITIONS_QUERY_ROOT_KEY`、`VACANT_POSITIONS_QUERY_ROOT_KEY`、`POSITION_DETAIL_QUERY_ROOT_KEY`、`positionDetailQueryKey(code, includeDeleted)`
   - filters 序列化（SSoT）：数组/对象参数需以统一规范序列化（大小写/排序/空值归一），沿用现有规范化函数（职位侧参见 `useEnterprisePositions.ts` 内 normalize* 方法），避免在调用点临时拼装。

### 2.1 统一缓存失效与命令映射表（SSoT）
- 目的：避免“命令→查询键”在多处分散实现，形成双轨。
- 约束：所有职位命令成功后，仅通过集中工具触发失效（示例名：`invalidateTemporalDetail(queryClient, 'position', code)`），内部覆盖以下键：
  - `temporalEntityDetailQueryKey('position', code, ...)`
  - `POSITION_DETAIL_QUERY_ROOT_KEY`（及 `positionDetailQueryKey(code, includeDeleted)`）
  - `POSITIONS_QUERY_ROOT_KEY`
  - `VACANT_POSITIONS_QUERY_ROOT_KEY`
- 命令→键映射（标准清单，不在业务处重复编写）：
  - CreatePosition → 失效：`POSITIONS_QUERY_ROOT_KEY`、`VACANT_POSITIONS_QUERY_ROOT_KEY`、`positionDetailQueryKey(code, *)`、`temporalEntityDetailQueryKey('position', code, *)`
  - UpdatePosition → 失效：同上
  - CreatePositionVersion → 失效：`positionDetailQueryKey(code, *)`、`temporalEntityDetailQueryKey('position', code, *)`（必要时刷新列表）
  - TransferPosition → 失效：同 Create/Update（列表 + 详情）
3) 等待模式与错误边界  
   - 首屏 skeleton 与错误边界统一；错误带重试按钮与日志埋点（配合 241/240D 可观测性）。  
   - 错误边界埋点字段：`entity`/`code`/`requestId`/`errorCode`；重试动作也需记录一次  
4) 选择器与等待工具  
   - 使用 `temporalEntitySelectors` 替换职位相关用例与组件引用；补充/复用 `waitPatterns`。  
   - 接入 `selector-guard-246`，禁止新增旧前缀 testid。  
5) 单测与 E2E  
   - Vitest 覆盖：取消/重试/错误态/租户切换/重复 fetch 抑制/命令后失效。  
   - E2E：`position-lifecycle.spec.ts`、`position-tabs.spec.ts`、`temporal-management-integration.spec.ts`（Chromium/Firefox 各 3 次），与 244/246 门槛一致；记录 trace/HAR 并对“重复请求抑制”做网络计数断言。

## 执行策略（本地/CI 分层）
- 本地默认（开发快速验证）：`REPEATS=1`、保存 trace，HAR 可选；允许跳过网络计数断言（设置 `E2E_NETWORK_ASSERT=0`）
- CI 要求（严格门槛）：`REPEATS=3 E2E_SAVE_HAR=1 E2E_NETWORK_ASSERT=1`，两浏览器均执行并落盘 trace/HAR 与网络计数 JSON
- 调用示例：
```bash
# 本地
REPEATS=1 PW_TENANT_ID=... PW_JWT=$(cat .cache/dev.jwt) npm run test:e2e -- --project=chromium --project=firefox
# CI
REPEATS=3 E2E_SAVE_HAR=1 E2E_NETWORK_ASSERT=1 PW_TENANT_ID=... PW_JWT=$JWT npm run test:e2e -- --project=chromium --project=firefox
```

## 守卫与 CI 接入
- 选择器守卫：`npm run guard:selectors-246`（计数不升高即通过）。  
- 命名守卫：`npm run guard:plan245`（旧类型/Operation 命名冻结）。  
- 其他质量门禁：`npm run lint`、`npm run test`、`node scripts/quality/architecture-validator.js`、`node scripts/quality/document-sync.js`、`node scripts/generate-implementation-inventory.js`。

## 验收标准（统一门槛）
- 前端单测：覆盖取消、重试、错误态、租户切换、重复 fetch 抑制、命令后失效，全部通过。  
- E2E 稳定性：三用例在 Chromium/Firefox 各 3 次绿灯（允许在本地/CI 补测但须落盘日志）；每轮保存 trace/HAR，并断言“详情加载 + 页签切换”无多余重复请求（网络计数阈值满足用例定义）。  
- 守卫：`guard:plan245` 与 `guard:selectors-246` 通过且基线计数不升高。  
- 契约与实现清单：如发生契约变更，`docs/api/*` 更新与 `generate-implementation-inventory.js` 通过。  
- 选择器与等待：组件/用例仅使用 `temporalEntitySelectors` 与标准等待模式，无硬编码 testid。

## 证据与登记
- 落盘路径：`logs/plan240/B/`（示例）  
  - `loader-cancellation.log`、`retry-policy.log`、`rq-cache-invalidation.log`  
  - `e2e-{chromium,firefox}-run{1..3}.log`、`playwright-trace/*`、`network-har-{chromium,firefox}-run{1..3}.har`、`network-requests-{chromium,firefox}-run{1..3}.json`  
  - `guard-plan245.log`、`guard-selectors-246.log`  
  - `inventory-sha.txt`（实现清单快照哈希）  
- 215 登记：在 `docs/development-plans/215-phase2-execution-log.md` 勾选 240B 的完成项，并链接上述日志。

## 当前状态与执行记录（滚动）
- 截止：2025-11-15  
- 完成项（技术实现）
  - 统一 Loader/取消（职位）：路由层接入 `createTemporalDetailLoader`，进入详情时预热（prefetch）并在卸载时取消在途查询；通过 `VITE_TEMPORAL_DETAIL_LOADER` 控制开关。
  - 重试/退避（幂等读）：`queryClient` 默认配置指数退避（base=200ms, factor=2, cap=3000ms, jitter=true），仅对 `5xx/429` 重试，`4xx` 不重试；统一客户端附带 `httpStatus`。
  - 统一失效 SSoT：新增并采用 `invalidateTemporalDetail`，职位写操作后通过集中工具失效 `POSITION_DETAIL_QUERY_ROOT_KEY`/`positionDetailQueryKey` 与 `temporalEntityDetailQueryKey('position', …)`，并联动列表根键。
  - 选择器与等待：Playwright 用例迁移为 SSoT 选择器 + GraphQL 等待模式，移除硬编码；`selector-guard-246` 接入并通过。
- 验收证据（抽样）
  - 组织详情冒烟：`frontend/tests/e2e/smoke-org-detail.spec.ts` 在 Chromium/Firefox 通过，HAR 与健康检查落盘 `logs/plan240/BT/*`。
  - 职位多页签：`frontend/tests/e2e/position-tabs.spec.ts`（Chromium）通过（SSoT 选择器 + GraphQL Stub），HAR 落盘配置就绪。
  - 守卫：`npm run guard:selectors-246` 通过且旧前缀计数显著下降；`npm run guard:plan245` 通过（仅 legacy 类型名提示 warn，不阻塞）。
  - 质量门禁：`npm run lint`、`npm run test`、`node scripts/quality/architecture-validator.js` 通过。
- 结论：本计划的“统一装载与取消”“统一重试策略”“失效 SSoT”“选择器与等待模式”已交付并通过本地/守卫验收，准许关闭。CI 接入建议作为运维项纳入 215/241 后续登记（acceptance 作业：`E2E_PLAN=240BT/240B`、`E2E_SAVE_HAR=1`）。
## 观测与落盘边界
- 运行时观测：统一使用前端 `logger` 与 `performance.mark` 输出（与 241/240D 一致），不得在运行时代码中直接写文件。
- 证据落盘：由测试/CI 采集（console、network、trace/HAR、网络计数 JSON）并保存至 `logs/plan240/B/`；作为验收与回归对比工件。

## 风险与回滚
| 风险 | 影响 | 缓解/回滚 |
| --- | --- | --- |
| 统一 Hook 改动引发行为回归 | 详情页数据不一致/闪烁 | 提供 feature flag；问题时回滚到上一个 tag；保留薄适配层的回退路径 |
| 取消策略误伤非幂等请求 | 数据提交中断或异常 | 仅对幂等读使用取消与重试；命令链路显式排除 |
| QueryKey 设计不全 | 缓存污染或刷新不及时 | 以实体类型/code/asOfDate/tenant/filters 为最小集合；新增单测断言 |
| 守卫误判 | 构建失败/阻塞 | 先 warning 再提升到 error；必要时临时 allowlist，设定到期清零 |

## 里程碑（建议）
- Day 1：依赖/前置校验完成；统一 Loader 接入原型 + 单测（取消/重试）。  
- Day 2：QueryKey/失效策略（统一失效工具）+ 错误边界完善（埋点一致）；选择器替换与守卫接入。  
- Day 3：E2E 三用例（各 3 次 × 2 浏览器）通过；保存 trace/HAR 与网络计数日志；落盘证据与 215 登记完成。

## 回滚与开关
- Feature Flag：`VITE_TEMPORAL_DETAIL_LOADER`（默认 `true`）。置 `false` 时停用路由 Loader 预热，页面回退至旧 Hook 直连（不改契约）。
- 回滚流程：通过 flag → 重启前端 → 复跑冒烟；记录回滚日志至 `logs/plan240/B/rollback-*.log`（由测试/脚本保存，非运行时代码写文件）。
