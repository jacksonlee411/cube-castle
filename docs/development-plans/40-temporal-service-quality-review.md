# 40号文档：TemporalService 时态核心质量分析

## 背景与单一事实来源
- 本文聚焦 `cmd/organization-command-service/internal/services/temporal.go`，该文件实现 `TemporalService` 的版本插入、删除、日期变更、状态切换与时间线重算逻辑，是命令服务时态域的关键实现。本分析仅基于该源码及其直接调用的仓储接口，未引入其他事实来源，确保资源唯一性与跨层一致性。

## 现状问题
1. **时间线维护不完整**：`InsertIntermediateVersion` 仅回填前一版本的 `end_date`，未同步更新后一版本（新版本的 `end_date` 始终为 `NULL`），也未触发重算，导致时间轴出现重叠区间。同样，`SuspendActivate` 仅插入状态版本，没有清除上一条记录的 `is_current` 标记。
2. **事务与依赖脱节**：`insertVersion` 在事务中调用 `OrganizationRepository.ComputeHierarchyForNew`（使用独立连接），该查询不受当前事务保护，容易在并发情境下读取到旧层级路径，破坏原子性。
3. **幂等路径返回占位数据**：`SuspendActivate` 中 `currentStatus == TargetStatus` 时直接返回 `getCurrentVersion`，但该函数仅返回硬编码的 `Status: "ACTIVE"` 与提示信息，缺少真实的记录 ID/时间线信息，调用方无法据此更新缓存。
4. **审计/时间线事件缺失**：`writeTimelineEvent` 为空实现，意味着版本删除、日期变更等操作不会留下任何事件记录或可观测数据，违背“单事务维护时间轴与审计”的目标。
5. **错误处理和防御不足**：大量 SQL 更新（如 `updateEndDate`, `deleteVersionByDate`）未检查影响行数，若目标不存在会被静默忽略；`getCurrentStatus` 直接返回底层错误，无法区分组织缺失与数据库异常；`getAdjacentVersionsForUpdate` 等查询未使用 `Context`，无法响应取消。

## 改进建议
1. **补齐时间线更新**：在插入或状态切换后，显式更新下一版本的 `end_date` 与 `is_current`，或统一调用 `recomputeTimelineInTx` 保证边界/当前态正确；为新增版本设置准确的 `end_date`。
2. **事务内查询统一传递 Tx**：为 `OrganizationRepository` 增加接受 `*sql.Tx` 的接口，或在服务层重写必要查询，确保层级计算与写入处于同一事务中，避免读取过期数据。
3. **返回真实版本数据**：`getCurrentVersion` 应查询当前版本并返回完整字段（`recordId`、`status`、`effectiveDate` 等）；幂等路径需向调用方明确资源状态。
4. **实现时间线事件/审计写入**：补全 `writeTimelineEvent`，将关键操作记录到审计日志或时间线表，并在失败时回滚事务，保持“事实唯一性”链路。
5. **强化错误与上下文处理**：对关键更新检查受影响行数并在异常时返回结构化错误；所有数据库交互改用 `QueryContext`/`ExecContext` 并传入上游 `ctx`；对 `sql.ErrNoRows` 进行显式分支，提高调用者可读性。

## 验收标准
- [ ] 新增/状态变更后时间线边界与 `is_current` 始终一致（含单元测试覆盖插入、删除、状态切换组合）。
- [ ] 层级计算等读取在同一事务内执行，确保并发一致性，新增测试验证并发插入不破坏路径。
- [ ] 幂等路径返回真实版本数据，并被现有 handler 使用以正确更新缓存。
- [ ] `writeTimelineEvent` 写入审计/事件记录，操作失败会导致事务回滚，相关日志可追踪。
- [ ] 所有 SQL 操作采用带上下文的方法并检查影响行数，对 `ErrNoRows` 等情形返回清晰的业务错误。

## 一致性校验说明
- 所有结论基于 `temporal.go` 当前实现；优化期间需同步核对 `repository` 层接口与 OpenAPI 契约，确保返回结构与字段命名保持一致。
- 本文存放于 `docs/development-plans/`，待改进完成后可归档至 `docs/archive/development-plans/`，持续维护单一事实来源。

