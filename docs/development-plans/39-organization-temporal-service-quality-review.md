# 39号文档：组织时态服务质量分析

## 背景与单一事实来源
- 本文聚焦 `cmd/organization-command-service/internal/services/organization_temporal_service.go`，该文件实现组织版本创建、日期更新、删除以及状态变更的服务层逻辑。分析仅依据该源码及其直接依赖（例如 `repository.TemporalTimelineManager`、`repository.AuditWriter`），未引入其他事实来源，确保资源唯一性与跨层一致性。

## 现状问题
1. **跨事务一致性缺失**：服务层为每次操作开启 SQL 事务（`BeginTx`），但随后调用的 `timelineManager` 方法内部再次开启并提交独立事务。若时态写入成功而审计写入失败，外层事务回滚也无法撤销已提交的数据，导致业务与审计状态不一致。
2. **咨询锁未覆盖时态写入**：`pg_advisory_xact_lock` 仅作用于服务层事务；由于时间轴操作在新的连接上执行，锁无法保护内部读写，仍可能出现并发版本冲突或链重算竞态。
3. **时间轴结果未校验**：`changeOrganizationStatus` 假设 `timeline` 非空并直接迭代；若底层返回 `nil` 或空切片（例如目标版本不存在），会触发 panic 或导致响应不完整。
4. **审计信息缺乏上下文**：所有审计记录固定 `ActorType: "SYSTEM"` 与 `ActionName` 常量，未写入真实操作者角色；同时部分字段（如 `BeforeData`、`AfterData`）未对齐契约命名，增加后续追踪难度。
5. **输入标准化不统一**：`ParentCode`、`OperationReason` 在服务层与 handler 中存在重复清洗逻辑，不同入口可能产生空字符串与 `nil` 混用，影响数据库约束与审计记录一致性。

## 改进建议
1. **统一事务边界**：将时间轴与审计写入合并在同一显式事务中运行，可通过为 `TemporalTimelineManager` 提供 “InTx” 版本或允许外部传入 `Tx`，确保任一环节失败时整体回滚。
2. **扩展锁粒度**：在时间轴操作内部复用同一事务或显式传递 advisory lock，保证并发写入受控；必要时在表级增加 `(tenant_id, code, effective_date)` 约束并对冲突返回结构化错误。
3. **强化结果校验**：在处理 `timeline` 前做空值检查，缺失时记录告警并返回明确错误；同时补充单元测试覆盖无版本或链为空的场景。
4. **丰富审计上下文**：将调用方传入的 `actorID`、角色信息与请求来源写入审计，确保 `ActorType` 区分 USER/SYSTEM；字段命名与契约对齐（例如统一使用 camelCase），避免解析差异。
5. **集中输入清洗**：封装公共的 `normalizeOperationReason`、`normalizeParentCode` 等方法，在服务层统一处理，减少 handler 与服务层重复逻辑并保证 `nil`/空字符串语义一致。

## 验收标准
- [ ] 时间轴写入与审计记录在同一事务中完成，任一失败均回滚，已有测试覆盖成功与失败路径。
- [ ] 并发写入场景通过锁或唯一约束得到保护，`TemporalTimelineManager` 可安全在并发下运行，并对冲突返回结构化错误。
- [ ] 所有时态操作在 `timeline` 缺失时返回可观测错误而非 panic，相关单元测试覆盖异常路径。
- [ ] 审计记录包含真实操作者类型与上下文，字段命名与契约保持一致，新增测试验证生成的数据。
- [ ] 输入标准化逻辑集中管理，`OperationReason` 等字段在数据库与审计中表现一致（空值即 `nil`），通过测试验证。

## 一致性校验说明
- 文档结论全部基于 `organization_temporal_service.go` 当前实现及其直接依赖。后续改动需同步检查 `docs/api/openapi.yaml`、审计契约等权威文档，确保跨层名称与语义一致。
- 本文存放于 `docs/development-plans/`，后续落实后可迁移至 `docs/archive/development-plans/`，持续维护唯一事实来源。

