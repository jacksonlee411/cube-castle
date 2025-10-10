# 35号文档：OperationalHandler 运维接口质量分析

## 背景与单一事实来源
- 本文以 `cmd/organization-command-service/internal/handlers/operational.go` 为唯一事实来源，聚焦其中运维相关 REST 端点 (`/api/v1/operational/**`) 的实现；未引入其他源码或运行态数据，确保资源唯一性。
- 该文件直接依赖 `services.TemporalMonitor`、`services.OperationalScheduler` 与 `middleware.RateLimitMiddleware`，分析中假设其接口契约与当前实现保持一致，未对其内部行为做额外推断。

## 现状问题
1. **响应格式不统一**：大部分端点使用 `map[string]interface{}` 自行拼装 JSON，而出错时调用 `http.Error` 返回纯文本，导致上游调用方需要同时处理 JSON 与纯文本响应。
2. **类型缺失与数据准确性风险**：`GetTaskStatus` 统计 `runningCount` 时未根据任务状态递增，实际响应固定为 0；同时 `map[string]interface{}` 使字段缺乏编译期约束，易引发运行时错误。
3. **上下文与限流信息缺少保护**：`GetRateLimitStats` 直接暴露内部统计，未结合 PBAC/租户上下文过滤；`scheduler.GetTaskStatus()` 等调用也未使用 `context`，当请求被取消时无法及时回收。
4. **运维操作仅占位实现**：`executeCutover` / `executeConsistencyCheck` 仅打印日志并返回成功，`TriggerTask` 也只是切换占位函数，与 `OperationalScheduler` 真正调度逻辑脱节，容易让使用者误判执行结果。
5. **重复模板代码**：每个 handler 都手动设置响应头并编码 JSON，缺少统一的响应封装，与 `DevToolsHandler` 类似问题并行存在，增加维护成本。

## 改进建议
1. **统一响应结构**：引入 `SuccessResponse[T]`、`ErrorResponse` 等结构体并提供公共写入函数，保证成功/失败都以 JSON 返回；顺便持有 `requestId`、`timestamp` 等元数据。
2. **修正任务统计逻辑**：在 `GetTaskStatus` 遍历时根据任务的运行状态（例如 `task.Active` 或 `task.Running` 字段）统计 `runningCount`，并为该函数添加单元测试验证统计正确性。
3. **上下文与安全控制**：对 `scheduler`、`monitor`、`rateLimit` 的调用传递 `r.Context()` 并设置合适超时；`GetRateLimitStats` 应结合 PBAC 校验调用者角色，必要时仅返回聚合数据或脱敏信息。
4. **打通调度执行链路**：在 `executeCutover`、`executeConsistencyCheck` 中调用 `OperationalScheduler` 的实际执行接口（例如 `RunTask(ctx, name)`），并根据返回状态决定 HTTP 响应；保留日志但不要默认成功。
5. **抽取公共工具**：与 `DevToolsHandler` 共用的响应逻辑可下沉到内部包，避免重复代码；同时复用统一的时间格式与日志前缀以提升可读性。

## 验收标准
- [ ] 所有 `/api/v1/operational` 端点在成功或失败时均返回一致的 JSON 结构，并包含 `requestId` 与 ISO8601 时间戳。
- [ ] 任务统计接口准确反映启用与运行任务数量，附带单测覆盖常见任务状态组合。
- [ ] `TemporalMonitor`、`OperationalScheduler`、`RateLimitMiddleware` 的调用均接受请求上下文，取消请求可立即中止后台操作。
- [ ] `TriggerTask`、`TriggerCutover`、`TriggerConsistencyCheck` 实际调用调度器执行逻辑，并根据结果返回成功/失败。
- [ ] 代码通过现有 `go test`、`make lint` 校验，且新增封装不破坏 PBAC/RBAC 约束。

## 一致性校验说明
- 以上结论与建议严格依据 `operational.go` 当前实现，涉及的改动需在合入前对照 `docs/api/openapi.yaml` 及内部 PBAC 规则，避免接口契约与权限模型出现偏差。
- 文档保存于 `docs/archive/development-plans/`，便于后续变更完成后归档至 `docs/archive/development-plans/`，持续满足单一事实来源与跨层一致性要求。
