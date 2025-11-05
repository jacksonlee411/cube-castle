# Plan 219C1 – 审计事件底座与事务化改造

**文档编号**: 219C1  
**上级计划**: [219C – Audit & Validator 规则收敛](./219C-audit-validator.md)  
**目标周期**: Week 4 Day 20-21  
**负责人**: 后端团队（审计小组协助）  

---

## 1. 目标

1. 建立统一的 `AuditEvent` 数据模型，补齐 `recordId`、`correlationId`、`businessContext`、`entityCode` 等字段，确保与 `database/migrations/20251106000000_base_schema.sql` 的 `audit_logs` 表结构完全对齐。
2. 实现审计写入的事务化封装，支持 `LogEventInTransaction` 并在组织、职位、Job Catalog 服务内与业务更新同事务提交，写入失败时整体回滚。
3. 补充请求 ID/链路 ID 映射：命令服务处理链路需传递 `requestId` → `correlationId`，并统一写入 `business_context` 以便后续追踪。
4. 输出执行规范：更新 `internal/organization/README.md#audit`，明确必审计操作、字段含义及调用模式。

---

## 2. 范围

| 模块 | 内容 |
|---|---|
| `internal/organization/audit` | 扩展 `AuditEvent` 结构、实现 `LogEventInTransaction`、封装 JSON 序列化与 `businessContext` 聚合。 |
| `internal/organization/service` | 组织、职位、Job Catalog 服务替换直接 `LogEvent` 调用，改为事务内写入并传递 `entityCode`/`actor` 信息。 |
| `internal/organization/handler` | 确保请求入口补齐 `requestId`，传入审计层；处理失败时调用 `LogError`，保留同一事务回滚语义。 |
| 文档 | 更新 `internal/organization/README.md#audit`、在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 引用新的审计指引（新增条目“Transaction-aware Audit Logger”）。 |

不包含：审计查询接口优化（由查询团队负责）、审计保留期自动化配置（归属平台组）。

---

## 3. 详细任务

### 3.1 审计事件模型重构
- [x] 为 `AuditEvent` 增补 `CorrelationID`、`EntityCode`、`BusinessContext`、`RecordID`、`ActorName` 字段，结构化落盘。
- [x] 补充 `marshalOrDefault` 等工具函数，避免 JSON `null` 破坏审计契约。
- [x] `LogError` 应写入 `business_context.payload`，提供失败上下文。
- [x] 按下表核对字段与 `audit_logs` 表结构，逐项验证代码输出：

| audit_logs 字段 | 计划要求 | 代码落点 |
| --- | --- | --- |
| `tenant_id` | 必填，来自处理链路租户上下文 | `AuditEvent.TenantID` |
| `resource_type` / `resource_id` / `record_id` | `resource_id` 保存 record UUID 或 fallback 标识，`record_id` 必须写入可解析 UUID | 服务层调用 `LogEventInTransaction` 时传入 |
| `request_id` / `business_context->correlationId` | `requestId` 直接落盘，`correlationId` 与 header 对应 | handler → service |
| `business_context.payload` | 保存 AfterData 快照或错误请求体 | `AuditLogger` 中注入 |
| `business_context.actorName` / `operation_reason` | 由服务层传入，缺省值需兜底 | 组织/职位/Job Catalog 服务 |
| `modified_fields` / `changes` | 序列化数组，保证非空时为 `[]` | `marshalOrDefault` |
| `timestamp` / `success` | 审计时间 `UTC`、结果枚举 | `logEvent` |
| `error_code` / `error_message` | 错误场景由 `LogError` 写入 | handler/service 错误路径 |

### 3.2 事务内写入
- [x] 在各服务中实现事务写入：`OrganizationTemporalService`、`JobCatalogService`、`PositionService`、`repository.AuditWriter` 之前的直接调用均需改造为 `LogEventInTransaction`（详见下列表）。
- [x] `audit.LogEvent` 保持向后兼容，但内部根据 `queryExecutor` 接口自动选择事务或裸连接。
- [x] 由服务层提供 `entityCode` / `actorName` / `operationReason`，以保证审计上下文完整。
- [x] 清理遗留的 `logCatalogEvent`、`logPositionEvent` 等函数，统一封装在同事务上下文中，调用点须列出清单并逐项勾选（参照下列表执行并在代码评审时验证）。

### 3.3 链路 ID 贯通
- [x] 中间件（`internal/organization/middleware/request.go` 及命令服务共享的 `internal/middleware/request_id.go`）输出 `requestId`，服务层调用审计时同步写入 `correlationId`。
- [x] 若上游提供 `X-Correlation-ID`，优先使用其值并在 `business_context.sourceCorrelation` 标注来源；若缺失则以 `requestId` 作为 `correlationId`，避免生成新的随机 ID，确保兼容现有 `cmd/hrms-server/command` 中间件。
- [x] 需要时扩展现有中间件以捕获 header（遵守仓库请求头命名规范）。

### 3.4 文档更新
- [x] `internal/organization/README.md#audit`：增补字段规范、事务调用流程、必审计操作表。
- [x] `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`：加入审计检查清单引用，提示执行 `go test ./internal/organization/audit`。
- [x] `docs/reference/02-IMPLEMENTATION-INVENTORY.md`：更新审计模块章节，标注 `AuditLogger` 与服务侧事务化使用方式。


---

## 4. 交付物

- 更新后的 `internal/organization/audit/logger.go` 与相关服务调用代码。
- README & 速查文档审计章节。
- 审计单元测试覆盖 `LogEventInTransaction`、`LogError`。
- `go test ./internal/organization/audit` 通过。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 审计写入失败导致主事务回滚，引发业务不可用 | 高 | 开发阶段开启详细日志并在 staging 环境压测；如需 feature flag，则仅在 staging 环境允许关闭事务化，正式环境保持开启，并在 README 标注配置项位置。 |
| 旧代码缺少 `requestId`，导致 `correlationId` 为空 | 中 | 在 handler 层兜底生成请求 ID，并在文档中强调调用规范。 |
| JSON 序列化失败导致审计记录写入异常 | 中 | 使用统一序列化函数并在失败时回退 `{}` / `[]`。 |
- ### 3.5 事务化改造清单（执行时逐项勾选）

| 组件 | 影响函数/方法 | 备注 |
| --- | --- | --- |
| Job Catalog 服务 | `JobCatalogService.logCatalogEvent` 及所有调用点（create/update 系列方法） | ✅ 已替换为 `LogEventInTransaction`，传入 `entityCode` 与 `actorName`。 |
| Position 服务 | `PositionService.logPositionEvent`、`createAssignment`、`CreateAssignmentRecord`、`UpdateAssignmentRecord` 等 | ✅ 事务闭包内写审计，并传递 `entityCode`、`operationReason`。 |
| 组织模块 | `OrganizationTemporalService` 中的 `CreateVersion`/`UpdateVersionEffectiveDate`/`DeleteVersion` 等，以及 handler 层直接调用 audit 的路径 | ✅ 统一使用 `AuditLogger` 事务接口。 |
| 审计辅助 | `repository.AuditWriter` 与其调用链 | ✅ 所有调用已迁移至 `AuditLogger`；旧实现已归档至 `docs/archive/internal/organization/audit_writer.go`。 |

---

## 6. 验收记录（完成于 2025-11-05T01:54:19Z）

| 勾选项 | 状态 | 证据 |
|---|---|---|
| `go test ./internal/organization/audit` 通过 | ✅ | `logs/219C1/test.log`（缓存命中亦记录，命令执行于 2025-11-05T01:54:19Z UTC） |
| 审计规范文档更新 | ✅ | `internal/organization/README.md:11` 起的 “## 审计规范（219C1）” 段落，与 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`, `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 互引 |
| 事务化调用点清单闭合 | ✅ | 本文件 §3.5 勾选完成，代码对应 `internal/organization/service` 与 `internal/organization/audit/logger.go` |

---

## 7. 归档说明

- 本计划已按目标完成，所有代码与文档同步至唯一事实来源。
- 相关测试与命令输出保存于 `logs/219C1/`，供后续审计追溯。
- 归档操作：将本文件复制至 `docs/archive/development-plans/219C1-audit-foundation.md`，并在 219C 总计划中标记已归档。
