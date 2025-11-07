# GraphQL Runtime (gqlgen)

该目录存放查询服务 gqlgen 运行时代码：

- `generated.go`：gqlgen 自动生成的 `ExecutableSchema`，引用 `docs/api/schema.graphql`。
- `model/`：必要时由 gqlgen 生成的辅助模型；默认依赖 `cube-castle/internal/organization/dto`。
- `resolver/`：手写桥接层，调用 `internal/organization/resolver` 中的领域逻辑。
- `custom_scalars.go`：实现 gqlgen 所需的 `Marshal/Unmarshal` 函数，并复用 `internal/organization/dto` 类型。

生成命令：`go run github.com/99designs/gqlgen generate -c cmd/hrms-server/query/gqlgen.yml`。
