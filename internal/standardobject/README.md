# internal/standardobject

本目录提供 Plan 400/402 约定的 SOM 端口骨架，当前内容聚焦在 402A 阶段所需的“前置桩”：

- `api.go` 定义对象内核、版本与 Link 聚合体，供命令/查询服务依赖注入。
- `featureflag/` 暴露 `STANDARD_OBJECTS_ENABLED` 等特性开关的最小实现，所有服务应通过该接口判断是否可写入 SOM。
- `adapter/noop` 交付 `Provide` 工厂，默认返回 `NoopService`。Feature Flag 打开后仍会返回 `ErrAdapterNotConfigured`，以提醒 Phase B 之前尚未绑定仓储。

运行 `go test ./internal/standardobject/...` 可确保骨架在 CI 中保持可用；Feature Flag 与日志的使用规范记载在 `docs/development-plans/402A-standard-object-mapping-spec.md` §4。
