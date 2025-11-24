# internal/standardobject

本目录提供 Plan 400/402 约定的 SOM 端口，目前已进入 402B 的“schema + 仓储 + 工具链”阶段：

- `api.go` 定义对象内核、版本与 Link 聚合体，命令/查询服务通过该接口解耦时态模型。
- `featureflag/` 暴露 `STANDARD_OBJECTS_ENABLED` Toggle，所有入口都需先判断该开关。
- `adapter/noop` 仍作为 Feature Flag 关闭时的回退桩；`adapter/sqlc` 则在数据库连通且 Flag 开启时创建真实仓储（使用 `internal/standardobject/repository` 与 sqlc 生成物）。
- `repository/` 封装 sqlc 查询与 `pkg/temporal/constraints`，在 `Upsert/Get` 中填充 `validity_range/transaction_range`。

命令服务通过 `cmd/hrms-server/command/main.go` 中的 `sqlcadapter.Provide` 注入真实 Port；工具链（`cmd/tools/standardobject-*`）同样依赖该仓储以满足 402B 的迁移/校验需求。运行 `go test ./internal/standardobject/...` 可快速验证 Feature Flag 与 skeleton 未被破坏。更多规范见 `docs/development-plans/402A-standard-object-mapping-spec.md` 与 `docs/development-plans/402B-som-schema-and-tooling.md`。

## 402C 进展
- 组织模块：`Create/Update/Version/Suspend/Activate/Delete` 等 handler 已在单事务内调用 `standardobject.ObjectService`，失败会回滚主事务；timeline 作废/删除会根据最新版本回写 SOM。
- 职位模块：`PositionService.Create/Replace/CreateVersion` 注入 `ObjectService` 并生成 `POSITION_BELONGS_TO_ORG` Link，payload 含岗位/编制/目录元数据。
- 事件：所有命令入口在 upsert 成功后通过 `emitStandardObjectEvent`/`PositionService.emitStandardObjectEvent` 写入 `standard_object.created/updated/versioned/status_changed/retired` outbox 事件，事件失败同样阻断主事务；运行中的 dispatcher 巡检与样本日志存放在 `logs/plan402/eventbus/*.log`。
- 版本号：统一通过 `standardobject.MakeVersionCode(code, effectiveDate, updatedAt, recordId)` 生成，确保同日多次纠偏仍满足 `(object_id, version_code)` 唯一约束。
- 日志：双写/迁移证据统一落在 `logs/plan402/migration|validator|snapshots|schema|metrics` 与 `logs/plan402/capability/*.log`，PR 需引用对应文件。
