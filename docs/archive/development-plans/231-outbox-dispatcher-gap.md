# Plan 231 – Outbox Dispatcher 接入差距分析

**编号**: 231  
**上级计划**: Plan 219E / Plan 06  
**创建时间**: 2025-11-08 11:35 CST  
**负责人**: 命令服务团队 + 平台团队  

---

## 1. 背景

- Plan 219E §2.4 要求验证 “Outbox → Dispatcher → Query 缓存” 链路。2025-11-08 依据 `logs/219E/outbox-dispatcher-plan.md` 执行 Runbook（O1-O6），触发 Position/Assignment/Job Level 场景以期生成 Outbox 事件。
- 结果显示命令请求与 GraphQL 读模型均成功，但 `outbox_events` 表为空、所有 outbox 指标为 0，意味着命令层尚未将事件写入 Outbox，Dispatcher 也没有可派发的记录。
- 为避免重复排查，单独立项记录事实、影响与后续整改路径。

---

## 2. 关键发现（2025-11-08）

| 现象 | 详情 | 证据 |
| --- | --- | --- |
| Outbox 表无记录 | `psql` 查询 `outbox_events` 返回空结果，说明命令服务未写入任何事件 | `logs/219E/outbox-dispatcher-sql-20251108T112236.log` |
| Dispatcher 指标全为 0 | `curl http://localhost:9090/metrics | rg 'outbox_dispatch'` 显示 success/failure/retry 均为 0，`outbox_dispatch_active=0` | `logs/219E/outbox-dispatcher-metrics-20251108T112459.log` |
| 服务日志仅有启动信息 | `docker logs cubecastle-rest` 只有 “outbox dispatcher started”，无 “dispatch event” 类日志 | `logs/219E/outbox-dispatcher-run-20251108T112541.log` |
| 命令请求成功但无 Outbox 记录 | 自测脚本成功创建/填充/关闭任职，并在 GraphQL 读模型中可见，但对应 requestId 没有 outbox 事件 | `logs/219E/outbox-dispatcher-events-20251108T112139.log`、`logs/219C3/validation.log`、`logs/219E/position-gql-outbox-20251108T112820.log` |
| 代码无 Outbox 写入调用 | 仓库中未找到对 `database.NewOutboxEvent` 或 `OutboxRepository.Save` 的业务引用，暗示 Outbox 集成尚未接入 Position/Assignment/Job Catalog 事务 | `rg` 搜索结果（命令服务仅在 main.go 初始化 OutboxRepository/Dispatcher） |

---

## 3. 风险与影响

1. **计划阻塞**：Plan 219E/Plan 06 的 “Outbox → Dispatcher → Query 缓存” 退出条件无法完成，219E 重启 gating 仍然为 ❌。
2. **事件驱动缺失**：没有 Outbox 事件意味着 Query 缓存刷新、下游订阅或审计中继无法依赖 Dispatcher，业务回退/监控也缺乏依据。
3. **指标失真**：Prometheus 中 `outbox_dispatch_*` 恒为 0，不具备报警/可视化价值。

---

## 4. 待办与责任

| # | 任务 | Owner | 说明 | 依赖 |
| --- | --- | --- | --- | --- |
| A1 | 梳理命令事务生成点 | 命令服务团队 | 明确哪些业务操作需写入 Outbox（Position、Assignment、Job Catalog、Organization 等），并在《事件契约章节》（新增）记录事件类型与字段。 | `pkg/database/outbox.go` 接口 |
| A2 | 接入 Outbox 写入 | 同上 | 在相关 usecase 中构造 `database.OutboxEvent` 并调用 `OutboxRepository.Save`（复用事务），同时将事件契约同步到 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`。 | A1 |
| A3 | 补充 Dispatcher 指标记录 | 平台团队 | 在 dispatcher success/failure 时调用 `internal/organization/utils.RecordOutboxDispatch`，并在 `docs/reference/monitoring/` 补充指标说明。 | A2 |
| A4 | 重新执行 Runbook | QA + 平台团队 | 修复完成后重跑 `logs/219E/outbox-dispatcher-plan.md` O1-O7，产出新的 `outbox-dispatcher-*.log`。 | A1-A3 |
| A5 | 更新 Plan 219E / Plan 06 验收条目 | QA | 按 Plan 06 §6 / 219E §2.4 的 gating 要求回填 Runbook 结果；若事件写入未达标，则保持阻塞并在本计划更新原因。 | A4 |

---

## 6. 解决方案与实施计划

### 6.1 设计原则
1. **事务一致性**：Outbox 事件必须与命令事务同库、同事务提交（复用现有 `repository.WithTx` / `dbClient.WithTx`）。
2. **统一结构**：以 `database.OutboxEvent` 为唯一实体，`event_type` 使用 `<聚合>.<动作>`（例如 `position.filled`、`assignment.closed`、`jobLevel.versionCreated`），`payload` 保存扁平 JSON（含 tenantId、positionCode/assignmentId 等）。
3. **领域扩展性**：首批覆盖 Position/Assignment/Job Catalog，保留钩子以便后续 Organization/Temporal 复用。
4. **可观测性**：Dispatcher 成功/失败需记录到 `internal/organization/utils.RecordOutboxDispatch`，Prometheus 指标与 Runbook 取证保持一致。

### 6.2 实施步骤
| 阶段 | 描述 | 输出 | 负责人 | 预计用时 |
| --- | --- | --- | --- | --- |
| P1 | **命令层设计梳理**：列出所有需要对外广播的事件（Position Create/Fill/Vacate/Delete、Assignment Close、Job Level Version Create/Conflict 等），定义 `event_type` 与 `payload` 字段；在 `internal/organization/events`（新建目录）内统一封装构建函数，并在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 新增“Outbox 事件契约”章节。 | 设计说明 + 事件枚举表 + 文档 diff | 命令服务架构师 | 0.5 天 |
| P2 | **接入 Position / Assignment**：在 `position_service.go` 的 `CreatePosition`、`FillPosition`、`VacatePosition`、`CloseAssignment` 等事务内调用 `withTx` 时追加 `outboxRepo.Save(...)`。可在 `positionRepository` 返回结构中附带 `tenantID`、`positionCode`，构造 payload。 | 代码改动 + 单元测试（mock OutboxRepository） | 命令服务团队 | 1.5 天 |
| P3 | **接入 Job Catalog**：在 `job_catalog_service.go` 的版本创建/冲突逻辑中写入 outbox（`event_type=jobLevel.versionCreated`/`jobLevel.versionConflict`），payload 包含 `jobLevelCode`、`recordId`、`effectiveDate`。 | 代码改动 + SQLMock 测试 | 命令服务团队 | 0.5 天 |
| P4 | **Dispatcher 指标强化**：在 `cmd/hrms-server/command/internal/outbox/dispatcher.go` 的 `publishOne` 成功/失败、`IncrementRetryCount` 路径调用 `utils.RecordOutboxDispatch(result,eventType)`；将 eventType 透传，以便 Prometheus 可按聚合维度观测。 | 代码改动 + go test | 平台团队 | 0.5 天 |
| P5 | **端到端验证**：运行 `scripts/219C3-rest-self-test.sh` + Runbook O1-O7，再执行额外的 Organization 更新以确保多领域事件写入；生成新的 `outbox-dispatcher-*.log`，确认 SQL 有记录、指标计数 >0。 | 日志、截图、PromQL 记录 | QA + 平台 | 0.5 天 |
| P6 | **文档/计划更新**：将成功日志回填至 219E/Plan 06（注明 Outbox gating 完成），在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` / `docs/reference/monitoring/` 记录事件契约与指标；若有额外领域需要接入，列入后续里程碑。 | 文档 diff | QA | 0.5 天 |

> 总计 3-4 个工作日（含缓冲），与 Plan 219E/Plan 06 时间线保持一致。

### 6.3 技术要点
- **事件构建工具**：建议新增 `internal/organization/events/outbox_builder.go`，提供 `NewPositionEvent(ctx, tenantID, positionCode, action, payload interface{})` 等 helper，避免各 service 重复拼装 JSON。
- **多事件写入**：若单一事务触发多个事件（例如 Fill + AssignmentCreated），可在同一事务中多次调用 `outboxRepo.Save`；`event_id` 使用 `uuid.New()`。
- **错误处理**：Outbox 写入失败时应使事务整体回滚（返回错误），防止业务提交但事件丢失。
- **配置验证**：在 `cmd/hrms-server/command/main.go` 启动阶段增加日志，打印 `OUTBOX_DISPATCH_*` 配置与是否成功注册 `prometheus.DefaultRegisterer`。
- **Prometheus 指标**：除 `utils.RecordOutboxDispatch` 外，保留 dispatcher 内部 counter（`outbox_dispatch_success_total` 等），Runbook 需同时检查两个来源的指标。

### 6.4 风险与缓解
| 风险 | 影响 | 缓解 |
| --- | --- | --- |
| 事务膨胀/性能影响 | 多次写入 Outbox 可能拉长事务时间 | 批量插入（同事务多条 INSERT）量不大；若后续热点明显，可考虑批量写入或异步缓冲，但当前以正确性优先 |
| 事件 schema 漂移 | 不同领域 payload 不一致，难以消费 | 事件 builder 统一字段（tenantId、aggregateId、positionCode 等）并在文档中记录 schema；必要时引入版号 |
| Dispatcher 重复派发 | 重试逻辑需保证幂等 | Outbox 模式已有 `event_id` 唯一约束；消费方需保证幂等，dispatcher 仅在 `MarkPublished` 成功后删除 |
| 多领域接入顺序 | Position/Assignment 优先，其他领域尚未安排 | 先满足 Plan 219E 的最小闭环（Position/Assignment/Job Level），其他领域列入后续迭代 |

---

## 5. 参考资料

- Runbook：`logs/219E/outbox-dispatcher-plan.md`
- 执行日志：  
  - `logs/219E/outbox-dispatcher-events-20251108T112139.log`  
  - `logs/219E/outbox-dispatcher-sql-20251108T112236.log`  
  - `logs/219E/outbox-dispatcher-metrics-20251108T112459.log`  
  - `logs/219E/outbox-dispatcher-run-20251108T112541.log`  
  - `logs/219E/position-gql-outbox-20251108T112820.log`
- 代码入口：`cmd/hrms-server/command/main.go`、`cmd/hrms-server/command/internal/outbox/*`、`pkg/database/outbox.go`

> 本文档为 Outbox 接入差距的唯一事实来源，后续进展请在此更新并同步到 Plan 219E / Plan 06。

## 7. 2025-11-08 实施进展

- ✅ 事件构建：`internal/organization/events/outbox.go` 落地统一 `Context`/helper，所有 payload 均包含 `tenantId/requestId/correlationId/source/occurredAt/aggregateId`，assignment/position/jobLevel 事件由此生成。
- ✅ Position/Assignment 写入：`internal/organization/service/position_service.go` 在 Create/Replace/Version/Transfer/ApplyEvent、Fill/Update/Vacate/Close assignment 路径调用 outbox（含 requestId、operationReason、headcount 数据），依赖注入于 `internal/organization/api.go`。
- ✅ Job Level 版本：`internal/organization/service/job_catalog_service.go` 在 create/version 成功时写 `jobLevel.versionCreated`，发生 “already exists for effective date” 冲突时追加 `jobLevel.versionConflict` 事件（独立事务），同样由 `api.go` 注入 OutboxRepo。
- ✅ Dispatcher 观测：`cmd/hrms-server/command/internal/outbox/dispatcher.go` 在 publish success/failure/retry 时调用 `internal/organization/utils.RecordOutboxDispatch`，Prometheus `outbox_dispatch_total` 可按事件/结果区分。
- ✅ 回归：`go test ./...`（覆盖 organization 服务、dispatcher、pkg/database 等）全部通过，保证 CQRS 层新增依赖可编译。

## 8. Runbook 复测与验收（2025-11-08 13:11 CST）

| 验收项 | 结论 | 证据 |
| --- | --- | --- |
| O2-O3 命令触发 + Outbox 取证：`BASE_URL_COMMAND=http://localhost:9090 DATABASE_URL=postgres://user:password@localhost:5432/cubecastle?sslmode=disable ./scripts/219C3-rest-self-test.sh`，随后 `SELECT event_id,event_type,... FROM outbox_events` | ✅ Position/Assignment/JobLevel 事件均进入 outbox，并带上 `requestId/positionCode/jobLevelCode`，dispatcher 已将 `published` 标记为 `true` | `logs/219E/outbox-dispatcher-events-20251108T050948Z.log`、`logs/219E/outbox-dispatcher-sql-20251108T050948Z.log` |
| O4 Prometheus 指标：`curl -s http://localhost:9090/metrics | rg 'outbox_dispatch'` | ✅ `outbox_dispatch_success_total=5`，`outbox_dispatch_total{result="success"}` 已按 `assignment.closed/assignment.filled/position.created/jobLevel.versionCreated` 细分 | `logs/219E/outbox-dispatcher-metrics-20251108T051005Z.log` |
| O5 Dispatcher 日志：`docker logs cubecastle-rest --since 5m | rg 'outbox dispatcher' -A2` | ✅ Dispatcher 启动记录与指标钩子可用，无错误/重试输出 | `logs/219E/outbox-dispatcher-run-20251108T051024Z.log` |
| O6 Query 缓存校验：GraphQL 查询 `P1000033` | ✅ `position.assignmentHistory` 已同步 `assignment.closed` 事件，headcount 回落为 0 | `logs/219E/position-gql-outbox-20251108T051126Z.log` |

关键信息：

- Outbox 记录与请求 ID 一一对应：`position.created (requestId=187fb99c-f430-427b-8cca-8a4d8bf2e9b1)`、`assignment.filled (e947eb8e-4f94-418d-8b9a-6331f4614fe0)`、`assignment.closed (e4d585f5-3dd7-4789-8c6e-7456fc537ea3)`、`jobLevel.versionCreated (c385af48-006e-4a90-b766-db3c94939ce5 / 694440a8-c858-4d63-9f17-109c24270ca7)`。
- Dispatcher 成功计数与 `utils.RecordOutboxDispatch` 暴露的 `outbox_dispatch_total` 一致（全部为 `result="success"`，未触发 retry），满足 Plan 06/219E 对“指标可观测”的要求。
- GraphQL 读模型已在 `assignmentHistory` 中呈现 `assignmentId=5e3aa262-5676-42a2-8c9b-54c692f5b7f4` 的结束记录，并返回 `headcountInUse=0`，证实 O6 读链路刷新。

## 9. 关闭结论

1. Task A1-A4（事件梳理、命令写入、dispatcher 指标、Runbook 复跑）全部完成并具备唯一事实来源；A5（Plan 06/219E 验收同步）已在对应母计划中更新引用（参见 `docs/development-plans/06-integrated-teams-progress-log.md` §6 与 `docs/development-plans/219E-e2e-validation.md` §2.4）。
2. Outbox → Dispatcher → Query 缓存链路解除阻塞，Plan 219E/Plan 06 可以引用上述日志作为退出准则证据。
3. 本计划进入归档阶段：请将本文复制至 `docs/archive/development-plans/231-outbox-dispatcher-gap.md`（已完成），后续如需拓展其他聚合事件，请以本档案为单一事实来源。
