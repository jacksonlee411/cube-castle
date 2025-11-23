# internal/standardobject

本目录提供 Plan 400/402 约定的 SOM 端口，目前已进入 402B 的“schema + 仓储 + 工具链”阶段：

- `api.go` 定义对象内核、版本与 Link 聚合体，命令/查询服务通过该接口解耦时态模型。
- `featureflag/` 暴露 `STANDARD_OBJECTS_ENABLED` Toggle，所有入口都需先判断该开关。
- `adapter/noop` 仍作为 Feature Flag 关闭时的回退桩；`adapter/sqlc` 则在数据库连通且 Flag 开启时创建真实仓储（使用 `internal/standardobject/repository` 与 sqlc 生成物）。
- `repository/` 封装 sqlc 查询与 `pkg/temporal/constraints`，在 `Upsert/Get` 中填充 `validity_range/transaction_range`。

命令服务通过 `cmd/hrms-server/command/main.go` 中的 `sqlcadapter.Provide` 注入真实 Port；工具链（`cmd/tools/standardobject-*`）同样依赖该仓储以满足 402B 的迁移/校验需求。运行 `go test ./internal/standardobject/...` 可快速验证 Feature Flag 与 skeleton 未被破坏。更多规范见 `docs/development-plans/402A-standard-object-mapping-spec.md` 与 `docs/development-plans/402B-som-schema-and-tooling.md`。
