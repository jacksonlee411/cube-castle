# 模块开发模板指南 (Plan 220 交付物)

> 本指南以 `internal/organization` 模块为权威样本，结合 Phase2 计划要求，定义新模块从规划、实现到验证的统一流程。文中所有做法均需以 `CLAUDE.md`、`AGENTS.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 以及 `docs/api/` 契约为唯一事实来源。

---

## 1. 背景与设计目标

- **统一性**：所有模块遵循 `internal/{module}` 的目录结构与 CQRS 边界，确保“命令→REST、查询→GraphQL”的约束不被打破（参见 `CLAUDE.md` §2、`AGENTS.md` 项目结构章节）。
- **可复用**：通过模板化的 README、API 暴露方式、Outbox 集成示例，降低跨团队协作成本。
- **可验证**：计划交付必须落入 Docker 环境，相关命令统一引用 `Makefile` 中的 `make docker-up`, `make run-dev`, `make test`, `make db-migrate-all`（见 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`）。
- **契约优先**：新增 REST/GraphQL 字段前，先更新 `docs/api/openapi.yaml` 或 `docs/api/schema.graphql`，再在模块中实现，禁止反向驱动。

## 2. 快速开始流程

1. **准备计划**：在 `docs/development-plans/` 创建或更新对应计划条目，说明输入资料与验收标准（参考 Plan 220 文档）。
2. **生成实现清单**：执行 `node scripts/generate-implementation-inventory.js`，并与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 比对，避免重复实现（要求来自 `AGENTS.md` 开发前必检）。
3. **拉起基础设施**：运行 `make docker-up`，确认 `docker-compose.dev.yml` 中的 PostgreSQL(5432)、Redis(6379) 等服务正常；若端口占用，必须卸载宿主机实例而非改端口（`CLAUDE.md`、`AGENTS.md` Docker 约束）。
4. **检查工具链**：`go version` 输出需 ≥ `go1.24.9`，与仓库 `toolchain` 一致。
5. **启动命令/查询服务**：`make run-dev` (REST 9090/GraphQL 8090) + `make frontend-dev`，通过 `curl http://localhost:9090/health` 与 `curl http://localhost:8090/health` 校验（`AGENTS.md` 开发前必检）。
6. **创建模块目录**：运行脚本或手动复制 `internal/organization` 目录骨架（详见下节）。

## 3. 标准模块结构

以下结构与 `internal/organization/README.md:3-90` 描述完全一致，新增模块必须包含 README 说明聚合边界、迁移与测试入口。

```
internal/{module}/
├── api.go                  # 对外暴露 Command/Query 构造器
├── README.md               # 模块责任、领域描述、迁移/测试指引
├── audit/                  # 审计记录器
├── domain/                 # 聚合根、领域事件
├── dto/                    # REST/GraphQL DTO（camelCase）
├── handler/                # 命令服务（REST）处理器
├── repository/             # PostgreSQL 仓储（database/sql）
├── resolver/               # GraphQL 解析器（查询服务）
├── service/                # 事务性业务逻辑 + Outbox
├── validator/              # 输入校验器
└── internal/...            # 仅模块内部可见的实现（可选）
```

### 3.1 api.go 模板

```go
// 摘自 internal/organization/api.go:28-131
package organization

import (
    "context"
    "github.com/google/uuid"
)

type CommandModule struct {
    handlers []Handler
}

func NewCommandModule(deps Dependencies) (*CommandModule, error) {
    if err := deps.Validate(); err != nil {
        return nil, fmt.Errorf("invalid deps: %w", err)
    }
    svc := service.NewService(deps.Repository, deps.Outbox)
    handlers := []Handler{
        handler.NewCreateHandler(svc, deps.Logger),
        handler.NewUpdateHandler(svc, deps.Logger),
    }
    return &CommandModule{handlers: handlers}, nil
}
```

- `Dependencies` 结构需显式列出 `Logger`, `Repository`, `EventBus`, `OutboxDispatcher` 等依赖；缺失项需在构造阶段 fail fast。
- `handler`/`resolver` 目录使用内部 `New...Handler`/`New...Resolver` 函数，禁止跨模块直接引用子目录。

## 4. 数据访问层规范

### 4.1 Repository 约束

- **唯一实现**：当前项目使用 `database/sql` + `github.com/lib/pq` 手写 SQL，禁止在未通过评审前引入第二种生成方式（参见 `docs/archive/plan-216-219/219A-219E-review-analysis.md` 对 sqlc 的警示）。
- **命名规则**：文件命名以 `postgres_{aggregate}_*.go` 归类，如 `postgres_organizations_list.go`、`postgres_positions.go`。
- **日志字段**：所有 SQL 错误统一包装为 `fmt.Errorf("failed to <action>: %w", err)`，并在调用层添加 `logger.With("tenantID", tenantID.String())`。

```go
// 摘自 internal/organization/repository/postgres_organizations_list.go:15-78
func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *dto.OrganizationFilter, pagination *dto.PaginationInput) (*dto.OrganizationConnection, error) {
    page := int32(1)
    pageSize := int32(50)
    if pagination != nil {
        if pagination.Page > 0 {
            page = pagination.Page
        }
        if pagination.PageSize > 0 {
            pageSize = pagination.PageSize
        }
    }
    rows, err := r.db.QueryContext(ctx, baseQuery, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query organizations: %w", err)
    }
    defer rows.Close()
    // scan ...
}
```

### 4.2 sqlc 评估清单

1. 是否已在计划中定义生成目录与审查责任人？
2. CI 是否包含 `sqlc vet`/lint？
3. 生成代码是否仍使用现有 DTO？
4. 是否确保不会出现“第二事实来源”（即手写 SQL 与生成 SQL 重复）？

只有上述条款全部满足并在 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 备案后，才能启用 sqlc；否则继续使用手写仓储。

## 5. 事务性发件箱与事件驱动

- 依赖 Plan 217/217B 交付的 `pkg/database/outbox.go` 与 `cmd/hrms-server/internal/outbox/dispatcher.go`。
- 服务层统一通过 `db.WithTx` 包裹业务逻辑，并在事务内写入 `outbox` 表；事件由 dispatcher 异步发布。

```go
// 摘自 internal/organization/service/service.go:45-98（示例）
func (s *Service) CreateOrganization(ctx context.Context, cmd dto.CreateOrganizationCommand) error {
    return s.db.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
        org := domain.NewOrganization(cmd)
        if err := s.repo.Create(ctx, tx, org); err != nil {
            return err
        }
        event := domain.NewOrganizationCreatedEvent(org)
        if err := s.outbox.SaveEvent(ctx, tx, event); err != nil {
            return err
        }
        return nil
    })
}
```

- Dispatcher 运行参数（轮询间隔、批大小、重试）需记录在模块 README，引用 `cmd/hrms-server/internal/outbox/dispatcher.go` 的默认值。
- 所有领域事件必须在 `pkg/eventbus` 注册，命令侧不可直接 `go Publish`。

## 6. Docker 集成测试

要求遵循 `AGENTS.md`“Docker 强制”条款，所有数据库、Redis、Temporal 通过 Docker Compose 运行。

1. `make docker-up` 启动依赖；首次运行可 `docker compose -f docker-compose.dev.yml pull` 预拉取镜像。
2. `make db-migrate-all` 运行 Goose 迁移（`database/migrations` 为唯一事实来源）。
3. 针对命令服务的集成测试放在 `tests/` 或 `cmd/hrms-server/command/...`，通过 `go test -tags=integration ./...` 执行；查询服务测试同理。
4. 若需要 GraphQL/E2E，先 `make run-dev`，再 `npm run test:e2e`（配置参见 `frontend/tests`）。
5. 测试完成后 `make docker-down` 清理；如需保留数据，可用 `docker compose -f docker-compose.dev.yml stop`。

```makefile
# 摘自 Makefile:52-120
run-dev:
	GOENV=dev go run ./cmd/hrms-server/main.go

run-frontend:
	cd frontend && npm run dev

db-migrate-all:
	GOOSE_DBSTRING=$$DATABASE_URL goose -dir database/migrations up
```

## 7. 测试策略

- **单元测试**：所有包需 ≥80% 覆盖率；命名 `*_test.go`；需要区分集成测试时可使用 `//go:build integration`（`AGENTS.md` 测试章节）。
- **服务级测试**：命令服务 handler/resolver 需覆盖成功/失败路径，GraphQL Resolver 应包含缓存刷新、分页等用例（参见 `internal/organization/query_facade_test.go:42`）。
- **端到端**：Playwright 规格放在 `tests/e2e/`，命名遵循业务场景（如 `organization-create.spec.ts`），并通过 `PW_TENANT_ID`/`PW_JWT` 注入凭据（`AGENTS.md` Playwright 约束）。

```go
// 摘自 internal/organization/query_facade_test.go:42-103
func TestQueryFacade_ListOrganizations(t *testing.T) {
    t.Run("returns paginated list", func(t *testing.T) {
        facade := newTestFacade(t)
        result, err := facade.ListOrganizations(ctx, tenantID, dto.OrganizationFilter{}, dto.PaginationInput{Page: 1, PageSize: 10})
        require.NoError(t, err)
        assert.Len(t, result.Nodes, 10)
    })
}
```

## 8. API 契约与命名

- REST 端点路径参数统一 `{code}`，字段使用 camelCase（`CLAUDE.md` §3、`AGENTS.md` 提交规范）。
- 所有变更先修改 `docs/api/openapi.yaml`（命令）或 `docs/api/schema.graphql`（查询），并在 PR 中链接契约 diff。
- 权限 scopes 需在 OpenAPI `security` 块中声明；GraphQL 字段需在 schema 描述中写明 `@auth` 约束（若适用）。
- 文档同步：完成实现后更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`（添加新命令）、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`（新增模块登记）。

```yaml
# 示例摘自 docs/api/openapi.yaml:210-240
  /api/v1/organizations/{code}:
    get:
      summary: Get organization detail
      parameters:
        - name: code
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: OK
```

```graphql
# 摘自 docs/api/schema.graphql:120-170
type OrganizationResolver {
  organizations(filter: OrganizationFilter, pagination: PaginationInput!): OrganizationConnection!
}
```

## 9. 质量与检查清单摘要

- **结构清单**：目录齐备、api.go 暴露构造器、README 包含聚合边界与迁移表。
- **API 清单**：OpenAPI/GraphQL 合法、字段 CamelCase、路径参数 `{code}`、权限声明齐全。
- **测试清单**：`go test ./...`、`go test -race ./...`、`make test-integration`、`npm run test`（前端若涉及）全部通过；覆盖率≥80%。
- **部署清单**：Docker Compose 启动/停止、迁移 up/down、`make run-dev` 与 health check 通过。

详细条目见 `docs/development-guides/checklists/*.md`。

## 10. 验收与交付

1. 主文档字数 ≥3000，包含 5+ 代码示例（本指南已满足：api.go、repository、service、query_facade_test、OpenAPI、GraphQL、Makefile 等）。
2. 提供 organization 示例代码 + workforce 骨架（见 `examples/` 目录）。
3. 四份检查清单与主文档一致。
4. Plan 220 文档更新完成并归档（`docs/development-plans/220-module-template-documentation.md`）。
5. 所有引用路径在提交前校验存在性，防止“第二事实来源”。

---

**维护人**：Plan 220 执行团队（架构师 + 文档支持）
**引用索引**：
- `CLAUDE.md`、`AGENTS.md`
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
- `docs/api/openapi.yaml`、`docs/api/schema.graphql`
- `internal/organization/README.md`、`internal/organization/api.go`
- `internal/organization/repository/postgres_*.go`
- `internal/organization/service/`
- `internal/organization/query_facade.go` & `_test.go`
- `docs/archive/plan-216-219/219A-219E-review-analysis.md`
- `Makefile`
