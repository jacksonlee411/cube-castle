# 技术架构设计方案

版本: v4.2 | 最后更新: 2025-11-17 | 状态: 生产就绪架构

---

## 事实来源与目标
- 契约优先：REST 以 `docs/api/openapi.yaml`、GraphQL 以 `docs/api/schema.graphql` 为唯一事实来源，权限 scope 由契约派生。
- 数据真源：所有结构演进均经 `database/migrations/`，禁止直接改表或依赖历史脚本；`sql/seed` 仅供演示。
- 模块化单体：命令=REST、查询=GraphQL，同一进程遵循 PostgreSQL 单一数据源；跨模块只能经依赖注入或事务性 outbox。
- Docker 强制：`make run-dev`/`docker-compose.dev.yml` 拉起 PostgreSQL 16 与 Redis 7；如端口冲突需卸载宿主服务而非改映射。

## 运行时拓扑
- 单体服务 `cmd/hrms-server/command` 在 9090 端口托管 REST、GraphQL、BFF、监控端点；`ENABLE_LEGACY_DUAL_SERVICE=true` 时可附带 8090 旧查询进程，仅供本地排障。
- 内部依赖：`internal/organization` 聚合组织、职位、职位目录、运维模块；`pkg/database` 暴露连接池/指标；`pkg/eventbus` 提供内存事件总线。
- 前端 `frontend/` 通过 `frontend/src/shared/api` 的客户端访问 REST/GraphQL，状态由 React Query+Zustand 组合驱动。
- 异步链路：`cmd/hrms-server/command/internal/outbox` + `internal/organization/events` 负责 outbox 持久化、重放与回退，禁止直接用内存队列跨边界。

## 前端体验层
- 栈：React 19.1 + Vite 7 + TypeScript 5.8 + Canvas Kit 13；GraphQL 客户端为 Apollo，REST 统一在 `frontend/src/shared/api/restClient.ts`。
- 时态实体框架：`frontend/src/features/temporal-entities` 与 `docs/reference/temporal-entity-experience-guide.md` 定义 UX 契约，QueryKey/重试策略集中在 `frontend/src/shared/api/queryKeys.ts`。
- 质量守卫：Vitest + Playwright 1.56（`frontend/tests/e2e`），提交前需跑 `npm run lint`、`npm run validate:field-naming` 与 `npm run validate:ports`。
- 构建输出通过 `vite build` 生成静态资源，由反向代理或静态托管服务提供；开发态 `make frontend-dev` 绑定 3000 端口。

## 命令域 (REST + BFF)
- 核心路由：`internal/organization/handler/*` 暴露 `/api/v1/organization-units`、`/positions`、`/job-catalog`、`/operational` 族群；端点示例 `POST /organization-units/{code}/versions`、`POST /positions/{code}/timeline`。
- 认证链：`internal/auth` 的 JWT 中间件 + `auth.NewPBACPermissionChecker` 检查 scope；`authbff` 模块负责 `/auth/*`、`/.well-known/jwks.json`、OIDC 模拟登录并可选 Redis 会话存储。
- 业务协作：`internal/organization/service` 和 `.../scheduler` 提供时态写入、级联、审计、职位编制等能力；`organization.CascadeService` 确保多层更新与回滚。
- 数据访问：命令侧通过 `internal/organization/repository` 使用 `database/sql` + 预编译语句写入，所有写操作进入 outbox（事件由 `pkg/eventbus` 异步广播）。
- 防护：`internal/organization/middleware` 实现请求 ID、速率限制、性能标记；`cmd/hrms-server/command` 将 `/health`、`/metrics`、`/debug/rate-limit/*` 对外暴露。

## GraphQL 查询域
- 运行模式：GraphQL 由 `cmd/hrms-server/query/publicgraphql` 在单体进程挂载 `/graphql`（GET/POST）与 `/graphiql`；legacy 独立进程保留 `//go:build legacy`，仅在需要独立伸缩时编译。
- 技术栈：gqlgen (`cmd/hrms-server/query/internal/graphql`) + chi + `internal/middleware` 的 GraphQL Envelope；Resolver 依赖 `organization.NewQueryResolver`。
- 支持的核心查询：`organizations`、`organization`、`organizationStats`、`organizationHierarchy/Subtree`、`positions`、`positionTimeline/Versions`，全部支持时态参数与租户隔离。
- 缓存与派生：`organization.NewAssignmentFacade` + Redis 缓存常用分配数据，Audit 配置由 `internal/config` 注入，可在 `AllowFallback` 与 `circuitThreshold` 间调优。

## 数据与迁移
- PostgreSQL 16 单库：主 schema 由 Atlas+Goose 管理；`PRIMARY KEY (code, effective_date)` + `record_id UUID` 保证版本唯一，`is_future/is_temporal` 均为派生字段。
- 时态编排：`internal/organization/repository/temporal_*.go` 负责插入、更新、截断和删除；`scheduler/temporal_monitor.go` 做巡检并写审计。
- Redis 7：可选缓存（assignment、职位列表）与 BFF 会话；离线可降级为内存实现，但需记录 WARN。
- 观测表：审计事件、outbox、操作日志在 `database/migrations/*audit*` 中维护，任何新增表都需 Down 语句与相应索引。
- 备份流程：`make db-migrate-all` 前默认执行快照（CI 以容器卷代替），生产需结合云平台快照；文档记录在 `CHANGELOG.md` 与 `docs/development-plans/`。

## 安全与权限
- 令牌：全链路 RS256 JWT（`make jwt-dev-setup` 生成密钥，`.cache/dev.jwt` 供开发）；BFF 会生成受控短期 token，为前端注入 `PW_JWT`。
- 权限：PBAC + scope 外部化，任何变更先改 OpenAPI/GraphQL，再同步 `internal/auth/scopes.go` 的映射；角色→scope 禁止硬编码在 UI。
- 多租户：`tenant_id` 是所有主数据表、GraphQL 查询、REST 输入的必填字段；`X-Tenant-ID` 请求头由中间件校验并注入上下文。
- 传输：默认启用 CORS 限制与 `SECURE_COOKIES`；OIDC 真正落地前的模拟登录以 `// TODO-TEMPORARY` 注记，截止不超过一个迭代。

## 可观测性与运维
- 健康检查：`internal/monitoring/health` + `v9RedisChecker` 在 `/health` 输出 PostgreSQL/Redis 状态；CI 使用 `curl http://localhost:9090/health`/`8090` 验证。
- 指标：`pkg/database`、`cmd/hrms-server/query/internal/app`、`command/main.go` 注册 Prometheus 指标，默认收集 HTTP、DB 池、outbox；Grafana/Alertmanager 由平台对接。
- 日志：`pkg/logger` 输出 JSON，写入 stdout 并可由 Docker logging driver 收集；`audit.Logger` 记录安全关键操作到 PostgreSQL。
- 性能守卫：`performanceMiddleware` 标记关键路径，`/debug/rate-limit/stats` 在 DEV 暴露限流数据；`run-dev*.log` 存放于仓库根 `logs/`。

## 质量门禁与交付
- Go 侧：`make test`、`make coverage`、`make lint`、`make fmt`、`node scripts/quality/architecture-validator.js`、`scripts/check-temporary-tags.sh` 必须绿灯。
- 前端：`npm run lint`、`npm run test`、`npm run test:e2e`、`npm run validate:no-direct-backend`、`npm run typecheck`，必要时附加 `npm run dashboard:generate` 更新可视文档。
- 库存校验：提交前运行 `node scripts/generate-implementation-inventory.js` 并核对 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，防止重复造轮子。
- CI 策略：全部 workflow 运行在 GitHub `ubuntu-latest`，禁用 `self-hosted,cubecastle,wsl` 标签；仅允许 `feat/shared-dev` 分支推送并经 PR squash。
- 发布：`make build` 生成命令服务二进制；镜像构建由 Dockerfile + `docker-compose`.yml 驱动，升级需同步 `CHANGELOG.md`。

## 演进与风险
- 当前重点：完成 PBAC 策略外部化（配置→数据库）与 OIDC 对接，预计分两迭代；未完成前保留模拟登录，并在每次发布前验证 `AUTH_ONLY_MODE=true` 流程。
- GraphQL 独立化：保留 legacy 8090 进程作为回滚通道，计划在完成负载验证后于 2026Q1 删除 build tag；如需单独扩缩容，可通过 `cmd/hrms-server/query` 独立部署。
- Outbox 与 Redis 依赖：Redis 不可用时会回退至内存缓存但丢失共享状态，需要在运营手册中记录恢复流程。
- 临时方案清单：所有 `// TODO-TEMPORARY` 带截止日期，集中在 `scripts/todo-temporary-allowlist.txt`，每次迭代由架构守卫复盘。
