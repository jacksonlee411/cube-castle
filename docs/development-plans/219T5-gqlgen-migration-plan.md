# 219T5 – 查询服务 GraphQL 运行时迁移（graph-gophers → gqlgen）

> 唯一事实来源：`docs/api/schema.graphql`、`docs/development-plans/06-integrated-teams-progress-log.md`、`logs/219E/graphql-inspector-diff-20251107-181341.log`、`cmd/hrms-server/query/`。

## 1. 背景与问题陈述
- 当前查询服务使用 `github.com/graph-gophers/graphql-go v1.5.0`，在 GraphQL 标准指令（`@deprecated` 对 `INPUT_FIELD_DEFINITION`、`@oneOf`、`@specifiedBy`）上落后官方规范。
- Plan 06 要求 `npx graphql-inspector diff docs/api/schema.graphql http://localhost:8090/graphql` 无差；即使通过 `scripts/dev/sync-graphql-schema.cjs` 同步 SDL，仍有 6 项差异全部来自运行时不支持最新指令（日志：`logs/219E/graphql-inspector-diff-20251107-181341.log`）。
- graph-gophers 项目维护缓慢（v1.5.0 → v1.6 也未覆盖 2022 规范），继续停留在该运行时无法满足 Plan 06 的“契约唯一事实来源”门槛，并导致前后端 diff 工具反复报错。
- gqlgen 提供 schema-first codegen、支持最新 GraphQL 规范、自动生成 resolver scaffold，可降低契约漂移风险，并与我们的工具链（GraphQL Inspector、Playwright E2E）保持一致。

## 2. 目标
1. 将查询服务从 graph-gophers 迁移至 gqlgen（>= v0.17），确保内置指令/SDL 与 `docs/api/schema.graphql` 完全一致。
2. 通过 `graphql-inspector diff` 验证“无差”要求，解除 Plan 06 硬门槛。
3. 建立 schema-first 流程：SDL → gqlgen.yml → 自动生成 resolver／model，避免手工 struct 漂移。
4. 回归关键 E2E/GraphQL 契约测试，保证查询服务功能等价或行为明确记录。

## 3. 范围与假设
- **范围**：`cmd/hrms-server/query/`（GraphQL handler/resolver）、相关模型/loader、Docker 镜像构建流程、CI 脚本（如 `make run-dev`、`make run-dev-debug`）。
- **不在范围**：命令服务、前端 GraphQL 客户端、数据库 Schema（使用现有 migrations）。
- **假设**：
  - `docs/api/schema.graphql` 将继续作为唯一事实来源；迁移期间若发现 SDL 与需求不符，需先更新 SDL 再生成代码。
  - gqlgen 生成的 resolver 可复用现有 domain/service 层，主要工作集中在 GraphQL 绑定层。

## 4. 实施步骤
| 阶段 | 操作 | 说明 | 产物 |
| --- | --- | --- | --- |
| P0 – 准备 | 1. 评估现有 GraphQL handler（`cmd/hrms-server/query/main.go`、`internal/graphql/...`）与 graph-gophers-specific 代码。<br>2. 设计 gqlgen 目录结构并将生成物全部放在查询服务目录下（例如 `cmd/hrms-server/query/internal/graphql/generated.go`、`.../model`、`.../resolver`），避免在仓库根目录散落文件。<br>3. 编写 `gqlgen.yml` 并直接引用唯一事实来源 `docs/api/schema.graphql`（禁止复制到 `graph/schema.graphqls` 等副本）。 | 需要明确 resolver 的依赖注入方式，确保与现有 service/DAO 可复用。 | `gqlgen.yml` 初稿、迁移评估记录。 |
| P1 – 代码生成与编译 | 1. 引入 gqlgen 依赖（`go.mod`/`go.sum`）。<br>2. 运行 `go run github.com/99designs/gqlgen generate` 生成 resolver scaffold。<br>3. 将现有 resolver 逻辑迁移/植入到 gqlgen 生成的 `resolver.go` 与 `graph/resolver/*.go` 中。<br>4. 更新 `cmd/hrms-server/query/main.go` 的 http handler（改用 gqlgen `handler.GraphQL` + `playground`）。 | 需要小步提交，保持功能等价。 | 通过 `go build ./cmd/hrms-server/query`。 |
| P2 – 契约与测试 | 1. 执行 `npm run test:e2e -- tests/e2e/basic-functionality-test.spec.ts` 验证读模型未回退。<br>2. 执行 `npx graphql-inspector diff docs/api/schema.graphql http://localhost:8090/graphql --header ...`，预期无 diff。<br>3. 运行 `scripts/e2e/org-lifecycle-smoke.sh` 确认命令→查询链路可用。 | 如出现行为差异，记录在 219T 报告“变更说明”章节。 | 无 diff 的日志、更新后的测试记录。 |
| P3 – 清理与文档 | 1. 移除 graph-gophers 依赖及相关工具函数。<br>2. 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/06-integrated-teams-progress-log.md`、`docs/development-plans/219T-e2e-validation-report.md`，记录迁移行为、测试证据。<br>3. 在 `docs/development-plans/219T5-gqlgen-migration-plan.md` 添加验收记录。 | 调整 Makefile/CI 中的 GraphQL build 步骤。 | 文档更新、Plan 06 退出准则满足截图/log。 |

## 5. 风险与缓解
| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| gqlgen 生成的 resolver 与现有 service 接口不匹配 | 编译失败或行为 regress | 在 P0 评估阶段梳理依赖层次，将 domain/service 通过接口注入 resolver，减少耦合。 |
| SDL 与实现存在历史差异 | 生成代码后编译不过或 runtime panic | 以 `docs/api/schema.graphql` 为契约唯一来源，必要时先更新 SDL 并走评审，再运行 `gqlgen generate`；严禁通过 `scripts/dev/sync-graphql-schema.cjs` 将运行时 introspection 回写契约，如需排查只可在临时位置生成 diff。 |
| 迁移期间 E2E 阻塞 Plan 06 | 验证无法推进 | 采用 feature 分支 + `make run-dev` 自测，通过 `npm run test:e2e --grep basic` 快速回归，待 gqlgen 版本稳定后再合并。 |

## 6. 验收标准
1. 查询服务编译/运行基于 gqlgen，graph-gophers 依赖已移除。
2. `npx graphql-inspector diff ...` 在注入 `Authorization`/`X-Tenant-ID` 后返回“Schemas are the same”。
3. `scripts/e2e/org-lifecycle-smoke.sh`、`npm run test:e2e --project=chromium tests/e2e/basic-functionality-test.spec.ts` 均通过。
4. Plan 06 文档记录“GraphQL 契约 diff 阻塞已解除”。
5. 迁移过程中的行为变化（如 resolver 返回字段顺序、错误 message）有配套说明或回归测试覆盖。

---

## 7. P0 评估记录（2025-11-08）
- **GraphQL 入口**：`cmd/hrms-server/query/internal/app/app.go` 通过 `github.com/graph-gophers/graphql-go` 解析 `docs/api/schema.graphql` 并借助 `relay.Handler` 暴露 `/graphql`（参见 `app.go:24-120`、`app.go:200-260`）。当前 HTTP 栈（Chi + 自定义中间件）与 graph-gophers 强绑定，迁移后需改为 `handler.GraphQL(generated.NewExecutableSchema(...))` 并保留 Envelope/JWT/Prometheus 中间件。
- **Schema 加载**：运行时使用 `internal/graphql/schema_loader.go` 直接读取 `docs/api/schema.graphql`。gqlgen 流程需保持此单一来源，但改为“编译期生成 + 运行时校验”模式，避免现有 `MustParseSchema` 依赖。
- **Resolver 依赖**：领域解析器位于 `internal/organization/resolver/resolver.go`，直接引用 `graphqlgo.NullBool`、`graphqlgo.ID` 等类型（例如 `resolver.go:170-210`、`resolver.go:520-580`），并依赖 `auth.GetTenantID`、`shared/config.DefaultTenantID`、`dto` 模型。迁移时必须将 `NullBool` 等替换为标准 `*bool` 或 gqlgen 输入模型，同时保留现有日志与权限检查。
- **DTO/Scalar 绑定**：`internal/organization/dto/*.go`、`dto/scalars.go` 通过 `ImplementsGraphQLType` 与 graph-gophers 适配。gqlgen 需要新的 `Marshal/Unmarshal` 函数（例如 `Date`、`UUID`、`JSON`、`PositionCode`），建议在 `cmd/hrms-server/query/internal/graphql/scalars` 下实现，并在 `gqlgen.yml` 中显式映射，保持 `docs/api/schema.graphql` 的自定义标量语义不变。
- **目录规划**：现有 `cmd/hrms-server/query/internal/graphql/` 为空目录，可作为 gqlgen 生成物的宿主，具体布局参考下文 `gqlgen.yml`（`generated.go`、`model/models_gen.go`、`resolver/*.resolver.go`、`scalars/*.go`），确保所有代码位于查询服务模块内部且不污染仓库根目录。
- **契约一致性**：确认 `docs/api/schema.graphql` 为唯一事实来源，禁止再创建 `graph/schema.graphqls` 等副本；`scripts/dev/sync-graphql-schema.cjs` 仅可用于分析 diff，不能覆写契约文件。

### gqlgen 目录规划（P0 输出）
| 路径 | 角色 | 说明 |
| --- | --- | --- |
| `cmd/hrms-server/query/gqlgen.yml` | 配置 | 指向 `docs/api/schema.graphql`，声明生成目标目录、autobind 包、标量映射。 |
| `cmd/hrms-server/query/internal/graphql/generated.go` | 生成 | gqlgen 自动生成的 `ExecutableSchema`，package 暂定为 `graphqlruntime`。 |
| `cmd/hrms-server/query/internal/graphql/model/` | 生成 | 存放 `models_gen.go`，用于声明 gqlgen 需要的辅助模型（若 DTO 未覆盖）。 |
| `cmd/hrms-server/query/internal/graphql/resolver/` | 手写 | 连接 gqlgen Resolver 接口与 `internal/organization/resolver` 领域实现，细分查询/变更文件。 |
| `cmd/hrms-server/query/internal/graphql/scalars/` | 手写 | 提供 `Marshal/Unmarshal` 实现，将 GraphQL 自定义标量映射到 `internal/organization/dto` 类型或原生类型。 |
| `cmd/hrms-server/query/internal/graphql/doc.go` | 手写 | 简要说明目录与生成流程，提醒引用 `docs/api/schema.graphql`。 |

## 8. 进展与下一步（2025-11-08）
- **依赖与代码生成**：已在 `go.mod` 中加入 `github.com/99designs/gqlgen v0.17.45` 并通过 `replace github.com/99designs/gqlgen => ./third_party/github.com/99designs/gqlgen` 指向离线源码，执行 `GOFLAGS=-mod=mod go run github.com/99designs/gqlgen generate --config cmd/hrms-server/query/gqlgen.yml` 生成 `internal/graphql/generated.go`、`model/models_gen.go` 与 `resolver/schema.resolvers.go`。
- **领域层准备**：`internal/organization/resolver/resolver.go` 及测试移除 `graphqlgo.NullBool`，统一使用 `*bool`，保证 gqlgen 生成的输入能够直接复用现有 DTO；`go test ./internal/organization/resolver` 通过。
- **待办**：
  1. 在 `cmd/hrms-server/query/internal/graphql/resolver/` 中实现 gqlgen Resolver，将 Query/字段 proxy 到 `internal/organization/resolver`，补齐枚举与嵌套对象映射。
  2. 更新 `cmd/hrms-server/query/internal/app/app.go` 切换至 gqlgen `handler.GraphQL`，并移除 `github.com/graph-gophers/graphql-go` 依赖。
  3. 完成迁移后执行 `go build ./cmd/hrms-server/query`、`npx graphql-inspector diff docs/api/schema.graphql http://localhost:8090/graphql`、`scripts/e2e/org-lifecycle-smoke.sh` 等验证，满足 P2/P3 验收。

> 更新时间：2025-11-08 10:20 CST；负责人：查询服务小队。

## 9. 进展更新（2025-11-09）
- **契约同步工具**：`scripts/dev/sync-graphql-schema.cjs` 改为仅在 `logs/graphql-snapshots/` 下生成最新/历史 SDL 快照，不再覆盖 `docs/api/schema.graphql`，保证 SDL 仍是唯一事实来源。
- **gqlgen 运行时代码**：
  - 新增 `cmd/hrms-server/query/internal/graphql/custom_scalars.go`、`json_scalar_helpers.go`，结合 DTO 自带的 Marshal/Unmarshal 实现 gqlgen 所需的所有自定义标量，并通过 `return_pointers_in_unmarshalinput: true` 避免 `map[string]interface{}` 指针不兼容问题。
  - 在 `cmd/hrms-server/query/internal/graphql/resolver/resolver.go` + `resolver/converter.go` 中引入通用转换器，`schema.resolvers.go` 的 Query Resolver 现已全部调用 `internal/organization/resolver` 的现有逻辑，返回值再映射到 gqlgen `model.*`。
  - `go test ./cmd/hrms-server/query/internal/graphql/...` 与 `go test ./internal/organization/resolver` 均通过，证明桥接层未破坏领域实现。
- **结构调整**：目录 README 更新为 `custom_scalars.go` 说明，废弃 `scalars/` 空目录，新增自动生成的 `generated.go` 与 `model/models_gen.go` 作为查询服务私有实现。

### 下一步（优先级按 P1→P3）
1. **HTTP Handler 切换（P1）**：在 `cmd/hrms-server/query/internal/app/app.go` 中引入 `graphqlruntime.NewExecutableSchema`，替换 graph-gophers/relay，删除旧依赖；更新 `make run-dev`、`docker` 镜像以使用 gqlgen handler。
2. **契约无差校验（P2）**：在 gqlgen handler 生效后运行 `npx graphql-inspector diff ...`，并将结果回填到 Plan 06，确认 `@deprecated/@oneOf/@specifiedBy` 差异已消除。
3. **端到端验证与文档（P3）**：用 Playwright/GraphQL 测试覆盖新 runtime，记录在 `docs/development-plans/06-integrated-teams-progress-log.md` 与本计划验收章节，最后清理 graph-gophers 相关引用。
