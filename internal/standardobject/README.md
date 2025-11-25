# internal/standardobject

本目录封装 Plan 400/402 约定的 Standard Object Port，供命令/查询服务与工具链统一读写 SOM：

- `api.go`：定义 `ObjectService`、`ObjectAggregate`、`TemporalVersion` 等内核结构，是命令/查询服务与工具链的唯一接口。
- `adapter/sqlc/`：默认适配器，组合 sqlc 生成物与 `pkg/temporal/clock`，被 `cmd/hrms-server/command/main.go`、`internal/organization/api.go`、`cmd/tools/standardobject-*` 注入使用。
- `adapter/noop/`：仅在测试或 `standardobject.NewNoopService()` 显式注入时启用，生产路径不再依赖 Feature Flag；`featureflag/` 仅保留给 CLI/回归场景做显式门控。
- `repository/`：维护 `Upsert/Get` 等数据库操作，并在写入时自动填充 `validity_range/transaction_range` 以及 TC1/TC2 约束；所有迁移/校验工具（`cmd/tools/standardobject-migrator|validator`）共用此实现。

运行 `go test ./internal/standardobject/...` 可快速校验 Port/Adapter/Repository 是否保持一致性。Schema/DEC/OCL 改动必须先更新 `docs/development-plans/402A-standard-object-mapping-spec.md`、`docs/development-plans/402B-som-schema-and-tooling.md` 并通过实现清单脚本。

## 402C：命令 + Outbox 接入进展
- **组织 handler**：`internal/organization/handler/standard_object_adapter.go` 暴露 `upsertStandardObject`/`syncOrganizationTimeline`，被 `Create/Update/Version/Suspend/Activate/Delete` 以及事件入口（`organization_events.go`）复用；一旦 `standardObjects.Upsert` 失败即回滚当前事务。
- **职位服务**：`internal/organization/service/position_standard_object_adapter.go` + `PositionService.syncPositionStandardObject`/`emitStandardObjectEvent` 负责 `Create/Replace/CreateVersion` 的聚合映射，并生成 `POSITION_BELONGS_TO_ORG` Link。
- **Outbox 事件**：组织与职位在 upsert 成功后分别调用 `OrganizationHandler.emitStandardObjectEvent` 与 `PositionService.emitStandardObjectEvent`，写入 `standard_object.created/updated/versioned/status_changed/retired` 事件；事件写入失败视同主事务失败，日志与 dispatcher 巡检记录在 `logs/plan402/eventbus/*.log`。
- **版本号与事务时间**：所有入口统一通过 `standardobject.MakeVersionCode(code, effectiveDate, updatedAt, recordId)` 生成幂等版本号，`pkg/temporal/clock` 提供 `transaction_range`；相关 Runbook 记录在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md#standard-object-双写守则`。
- **证据登记**：命令/查询接入、Manifest/Slot 适配与 OBS/Playwright 证据在 `logs/plan402/capability/*.log`、`logs/plan402/ui/*.log`、`logs/plan402/migration|validator|schema|metrics` 中落盘，归档方法见 `docs/reference/standard-object-evidence-guide.md`。

如需新增实体或扩展 payload，请同步更新上述文档与日志目录，确保资源唯一性。
