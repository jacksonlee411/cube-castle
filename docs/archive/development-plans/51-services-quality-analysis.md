# 51号文档：服务层（Services）质量分析

## 背景与唯一事实来源
- 本文覆盖 `cmd/organization-command-service/internal/services/` 目录下的 `cascade.go`、`operational_scheduler.go`、`organization_temporal_service.go`、`temporal.go`、`temporal_monitor.go` 五个服务实现，所有结论均直接来源于当前源码。
- 已对照 37~41 号质量分析文档，确认此前结论仍成立；本文聚焦新增的交叉层面问题与合规性校验，避免引入第二事实来源。

## 代码质量评估
- `cascade.go`：并发模型清晰但缺乏回压控制，`Stop` 后无法安全重启；日志中夹杂 emoji，难以与集中式日志平台集成。
- `operational_scheduler.go`：基础调度循环可读性良好，但状态持久化缺失，`monitor` 依赖未与生命周期解耦。
- `organization_temporal_service.go`：事务包装完整，能统一审计与时间线，但存在混合直接 SQL 查询与仓储调用的模式，增加维护成本。
- `temporal.go`：代码量大且自带事务封装，与仓储层实现重复，仍保留大量“简化实现”占位逻辑。
- `temporal_monitor.go`：指标项覆盖较全，但 SQL 分支高度重复，统计与告警逻辑耦合度高，缺乏可配置能力。

## 主要问题
### 并发与生命周期缺陷
- `CascadeUpdateService.Stop` 在关闭 `taskQueue` 前仅通过 `running` 标志防护，存在 `send on closed channel` 的竞态；关闭后也不会重新构建通道/`shutdown`，导致服务不可重启。
- `processHierarchyUpdate/processPathUpdate` 为每个子任务额外启动 goroutine 调用 `ScheduleTask`，一旦队列满片段会产生短时间大量协程，且失败结果被忽略。
- `OperationalScheduler` 关闭后未重置 `stopCh`，同时未向 `TemporalMonitor` 下发独立的 cancel，上层调用 `Stop` 时监控循环继续运行；调度器与监控形成两个周期任务入口，易出现重复告警。

### 数据一致性与事务边界
- `CascadeUpdateService` 的业务分支多数仅打印日志，缺乏对层级/状态变更的落库保障；`Priority` 剪枝策略导致 3 层以上组织树无法完成级联更新。
- `OrganizationTemporalService` 在事务内混用 `QueryRowContext` 与仓储方法，`timelineManager` 的写入与 `orgRepo.ComputeHierarchyForNew` 读取使用不同连接，仍存在读到旧层级数据的窗口。
- `TemporalService` 自行维护版本插入、重算、状态变更，与 `repository.TemporalTimelineManager` 提供的能力高度重合；两套实现缺乏统一事务策略，容易出现行为不一致。

### 可观察性与运维保障
- `OperationalScheduler` 的任务执行记录仅打印日志，`GetTaskStatus` 返回重新计算的默认值，运维无法获知真实执行结果或失败原因。
- `TemporalMonitor` 将“全局巡检”与“租户级巡检”写成重复 SQL，且依赖 `auth.GetTenantID(ctx)` 返回字符串后直接拼装查询，缺乏 UUID 校验与 PBAC 限制；健康分扣分模型与阈值均为硬编码，难以按租户规模调节。

## 过度设计分析
- `CascadeTask.Priority` 和多任务类型枚举尚未落地优先级调度；当前使用无序缓冲通道，无论优先级与否处理顺序一致，形成“虚假复杂度”。
- `OperationalScheduler` 内建脚本目录发现、任务循环、监控执行多套机制，实际可交由现成 cron/任务框架或统一的 `TemporalMonitor` 触发器处理。
- `TemporalService` 与 `OrganizationTemporalService` 并行维护版本逻辑，形成两层服务抽象，增加审计与事务一致性验证成本。

## 重复造轮子情况
- `TemporalService` 重新实现了时间线重算、版本插入、状态切换等功能，这些能力已由 `repository.TemporalTimelineManager` 与 `OrganizationTemporalService` 提供，导致维护同一业务域的两份代码。
- `OperationalScheduler` 用手写循环驱动“每日/每小时”任务，缺少对成熟调度库（如 cron 表达式解析）或统一任务注册中心的复用；监控定时器与调度循环均在重复调度 `TemporalMonitor`。
- `CascadeUpdateService` 自建任务队列但也存在 `repository` 层可用的层级刷新方法（如基于 SQL 的批量更新），未评估是否可以复用数据库侧的递归能力。

## 综合改进建议
1. **统一服务分层**：评估废弃或封装 `TemporalService`，统一由 `OrganizationTemporalService` + 仓储层提供时态操作；同时补齐事务内读取接口，确保所有层级/审计写入处于同一连接。
2. **重构任务调度模型**：为级联与运维任务引入可重启的 worker 池或通用调度组件，去除 `go ScheduleTask` 等无回压调用；在 `Stop` 中等待现有任务完成并重建通道/信号以支持二次启动。
3. **落地可观察性**：为调度任务和监控巡检建立状态存储（表或 KV），`GetTaskStatus`、`CheckAlerts` 返回真实记录；同时支持阈值、扣分模型通过配置加载。
4. **权限与上下文治理**：监控、调度使用系统级 `context.Background()` 派生上下文，并显式校验租户/权限；为 SQL 查询增加 UUID 参数校验与查询级超时。
5. **消除无效复杂度**：在 `CascadeUpdateService` 中移除未使用的优先级字段或改为真正的优先队列；为任务类型保留的占位分支补全具体业务动作或收敛为更简单的策略。

## 验收标准
- [ ] 服务停止后可安全重启，worker 通道与停止信号均在 `Start` 过程中重新初始化，`ScheduleTask` 在队列关闭时不会 panic，并提供错误返回。
- [ ] 级联更新、时态操作均依赖统一的仓储接口或事务辅助函数，避免重复实现；删除 `TemporalService` 中的占位逻辑或确保其与仓储实现共享同一代码路径。
- [ ] 运维调度与监控任务执行结果持久化（含成功/失败状态、耗时），`GetTaskStatus`、`CheckAlerts` 返回真实数据并支持阈值配置。
- [ ] 监控与调度查询使用校验过的租户 ID，关键 SQL 具备上下文超时，并对全局/租户模式复用同一查询模板以减少重复维护。
- [ ] 对 `CascadeTask` 优先级与任务类型提供单元测试覆盖，验证在深层组织树、大量子任务场景下无任务丢失并能按期完成。

## 一致性校验说明
- 本文所有问题与建议均在源码中复核，不与既有 37~41 号文档冲突；在实施整改时需同步检查 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，确保命令/查询契约保持 camelCase 与 `{code}` 路径约定。
- 文档位于 `docs/development-plans/`，整改完成后按流程归档至 `docs/archive/development-plans/`，维持唯一事实来源链路。
