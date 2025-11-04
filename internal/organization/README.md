# internal/organization

组织聚合模块的唯一事实来源。目录结构：

- `audit/`: 审计记录写入器及依赖。
- `handler/`: REST/BFF 处理器（命令侧）。
- `middleware/`: 组织模块专用中间件（性能、限流、请求 ID）。
- `repository/`: 命令/查询共享的 PostgreSQL 仓储实现与时间轴管理器。
- `resolver/`: GraphQL Resolver（查询侧入口）。
- `service/`: 领域服务、时态服务、Job Catalog/Position/Cascade 等。
- `validator/`: BusinessRuleValidator 及规则定义。
- `dto/`: GraphQL 查询/响应 DTO 与共享类型。
- `utils/`: 处理器/仓储共用的工具函数（响应、验证、metrics、parent code 等）。

## 聚合边界

- Department 是 Organization 聚合内节点，通过 `unitType=DEPARTMENT` 表示。
- Position/JobCatalog/Assignment 共用 PostgreSQL 数据源，命令侧通过 `service/*` 管理，查询侧通过 `resolver` + `repository`。

## 迁移清单（219A）

| 旧路径 | 新路径 | 说明 |
| --- | --- | --- |
| `cmd/hrms-server/command/internal/handlers/*` | `internal/organization/handler/*` | REST/BFF 入口集中于 handler 包。 |
| `cmd/hrms-server/command/internal/services/*` | `internal/organization/service/*` | 领域服务共享给命令适配层。 |
| `cmd/hrms-server/command/internal/repository/*` | `internal/organization/repository/*` | 命令/查询仓储统一。 |
| `cmd/hrms-server/command/internal/audit/*` | `internal/organization/audit/*` | 审计日志实现。 |
| `cmd/hrms-server/command/internal/validators/*` | `internal/organization/validator/*` | 业务校验统一入口。 |
| `cmd/hrms-server/command/internal/utils/*` | `internal/organization/utils/*` | 公共工具函数。 |
| `cmd/hrms-server/query/internal/graphql/*` | `internal/organization/resolver/*` | GraphQL Resolver 共享。 |
| `cmd/hrms-server/query/internal/repository/*` | `internal/organization/repository/*` | 查询仓储共用组织模块。 |
| `cmd/hrms-server/query/internal/model/*` | `internal/organization/dto/*` | GraphQL DTO 单一来源。 |

## API 适配

- `internal/organization/api.go` 暴露 `CommandModule` 及 `CommandHandlers` 构建函数，命令服务只需依赖该 API。
- 查询服务通过 `internal/organization/resolver` & `repository` 注入 GraphQL 应用。

## 查询与缓存（219B）

- `AssignmentQueryFacade` 提供统一的任职查询、历史与统计接口，并负责 Redis 缓存键管理（前缀 `org:assignment:stats`）。
- 缓存策略：职位维度统计命中 Redis，TTL 默认 2 分钟，命令侧 Outbox Dispatcher 发布 `assignment.*` 事件后调用 `RefreshPositionCache` 触发失效。
- GraphQL 新增查询：`assignments`、`assignmentHistory`、`assignmentStats` 均通过 Facade 获取数据，保持与 `docs/api/schema.graphql` 契约一致。

后续 219B~219E 将在本 README 中继续补充审计/验证规则、调度说明、测试脚本等章节。
