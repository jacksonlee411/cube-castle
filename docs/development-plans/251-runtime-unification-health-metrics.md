# Plan 251 - 运行时统一与健康指标对齐

文档编号: 251  
标题: 运行时统一与健康指标对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.1  
关联计划: 202、203、215、218（logger/metrics）

---

## 1. 目标
- 统一运行时配置加载与健康检查接口（/health、/metrics）；
- 建立基础可观测性指标（HTTP、审计、时态操作、outbox 派发等）；
- 对齐 Docker 化本地/CI 行为（健康探针、退出策略），严格遵循 AGENTS 约束（不得调整端口映射）。

## 2. 交付物
- 运行时统一与健康/指标规范文档（本文件，引用唯一事实来源）；
- 健康检查基线（command/query 一致）：返回 camelCase JSON，包含整体 status 与各子服务状态；
- 指标基线：唯一事实来源与标签维度约定（见 4.2），并与代码实现一致；
- 验收脚本与证据：`scripts/quality/validate-metrics.sh`、`scripts/quality/validate-health.sh`；日志落盘 `logs/plan251/*`。

## 3. 验收标准（可执行）
- 单体主路径（对齐 202：单进程/单端口 9090）
  - 单一二进制、单端口（9090）可同时提供 `/health`、`/metrics`、`/api/v1`、`/graphql`
  - 仅本地允许“过渡模式”（旧的 8090 查询服务）；CI 禁用（参见运行时自检与环境开关）
- /health 统一：返回结构与状态码规范一致；字段 camelCase；command 与 query 均接入统一健康管理器；
  - 状态码：healthy=200、degraded=206、unhealthy=503（来源：`internal/monitoring/health/health.go` Handler 实现）
  - 必含字段：`service, version, status, timestamp, uptime, checks[], summary{total,healthy,degraded,failed}`
- /metrics 可达且指标命名/标签一致：公共指标与业务指标分层
- 公共指标（两端统一）：`http_requests_total{method,route,status}`、DB 连接池与耗时（见 4.2）
  - 业务指标（按模块）：`temporal_operations_total{operation,status}`、`audit_writes_total{status}`、`outbox_dispatch_total{result,event_type}`、（可选）查询专属 `organization_operations_total{operation}`
- 运行时配置来源单一：JWT/调度器等使用 `internal/config/*`；DB/Redis 配置覆盖规则一致（默认→文件→环境），记录覆盖元数据（sources/overrides）
- Docker 健康探针：dev 与 e2e Compose 对 rest/query 服务的探针语义一致（探测 `/health` 返回可接受状态），不可更改端口映射
- 证据：`logs/plan251/health-*.json|txt`、`logs/plan251/metrics-*.txt`，由脚本生成；CI 与本地命令一致
- 避免重复注册：单体路径下 GraphQL 请求复用统一 HTTP 指标（`internal/organization/utils/metrics.go`），独立查询进程使用相同指标定义与标签，禁止在同一进程重复注册同名不同描述的指标

---

维护者: 平台与后端联合（与 218 保持一致）

---

## 4. 决策与规范（SSoT）

4.1 健康检查
- 唯一实现：`internal/monitoring/health/health.go`（HealthManager + Handler）、`internal/monitoring/health/reporter.go`（Dashboard 可选）
- 路由规范：
  - 必选：`GET /health` → `HealthManager.Handler()`（JSON，按状态映射 200/206/503）
  - 可选：`GET /healthz` → `StatusReporter.DashboardHandler()`（HTML/JSON，便于人工观察）
- checks 约定：PostgreSQL（`PostgreSQLChecker`）、Redis（`RedisChecker`）、启动项/外部依赖按需增加；超时参数由实现内置
- 环境信息展示（Dashboard）：从运行时获取（如 `runtime.Version()`、主机名、环境变量），禁止硬编码静态值
- 鉴权白名单：`/health`、`/metrics` 必须公开（无鉴权）；其余遵循权限契约（OpenAPI scopes）
  - 网络层限制：生产环境应通过网络层（服务网格/Ingress/Compose 网络）限制 `/metrics` 暴露范围，避免对公网开放

4.2 指标（Prometheus）
- 唯一事实来源（命名与标签）：`internal/organization/utils/metrics.go`
  - `http_requests_total{method,route,status}`
  - `temporal_operations_total{operation,status}`
  - `audit_writes_total{status}`
  - `outbox_dispatch_total{result,event_type}`
- 数据库指标（217 标准，来源：`pkg/database/metrics.go`）
  - `db_connections_in_use{service}`、`db_connections_idle{service}`
  - `db_query_duration_seconds{service,query_type}`（直方图）
- 端点暴露：`/metrics` 使用 `promhttp.Handler()`（command/query 一致）
- 记录位置（示例，不扩展第二事实来源，仅作索引）：
  - HTTP 计数：由服务中间件或包裹器统一打点（query 端将 `path` 统一更名为 `route`）
  - 业务计数：时态操作（scheduler/service）、审计写入、outbox 派发（dispatcher）
  - 数据库直方图：通过 `pkg/database` 包装或在查询路径显式记录；连接池指标通过周期性调用 `RecordConnectionStats`
- 说明：Counter 在未被 Inc() 前不会出现在 `/metrics` 输出，这是 Prometheus 标准行为
- 路由标签规范（避免基数爆炸）：`route` 必须使用路由模板或受控常量（如 `/api/v1/organization-units/{code}`），不得使用原始 URL Path
- 单体冲突规避：禁止在同一进程重复注册同名不同标签的指标；GraphQL 若合流到单体，必须复用统一 HTTP 指标
- 高基数标签禁用：禁止引入 `request_id`、`user_id`、`tenant_id`、`error_message` 等高基数标签；错误原因以枚举/分级标识（如 `outcome`、`result`）表达

4.3 运行时配置（来源单一）
- JWT：`internal/config/jwt.go` 为唯一入口（RS256 强制），rest/query 共用；元数据记录：`Algorithm/Issuer/Audience/KeyID/AllowedClockSkew`
- 调度器：`internal/config/scheduler.go`（含 `Metadata.Sources/Overrides/ValidationError`）
- DB/Redis：本计划统一覆盖规则（默认→文件→环境变量），并在 query 侧迁移到共享方式；记录来源（sources/overrides）

4.4 Docker 健康探针
- dev/e2e 一致：`wget --spider -q http://localhost:{PORT}/health || exit 1`
- 禁止变更容器端口映射以规避宿主冲突（按 AGENTS 强制卸载宿主占用服务）

---

## 5. 变更清单（实施项）
- 接入统一健康：
  - command：用 `HealthManager` 替换手写 `/health` 响应（文件：`cmd/hrms-server/command/main.go`）
  - query：同上（文件：`cmd/hrms-server/query/internal/app/app.go`）
- 指标一致化：
  - 将 query 侧 `http_requests_total` 的标签从 `method,path,status` 改为 `method,route,status`
  - 保留/对齐 Prometheus 端点 `/metrics`（两端已使用 `promhttp.Handler()`）
- 数据库指标接入（217 对齐）：
  - 注册指标：在服务启动时调用 `database.RegisterMetrics(prometheus.DefaultRegisterer)`
  - 直方图：读写路径使用 `pkg/database` 的查询封装或显式调用 `database.ObserveQueryDuration`
  - 连接池：周期性调用 `dbClient.RecordConnectionStats(serviceName)`（Ticker）记录 `db_connections_*`
- HTTP 路由模板化（避免基数爆炸）：
  - 中间件层获取 chi 路由模板或受控常量作为 `route` 标签；禁止使用原始 URL Path
  - 单体路径下 GraphQL `/graphql` 请求复用统一 HTTP 指标，不额外注册重复 Collector
- 运行时配置一致化：
  - JWT：rest/query 均使用 `internal/config/jwt.go`（已达成）
  - DB/Redis：在 query 侧引入与 rest 等价的加载/覆盖模式，并记录覆盖元信息
- Docker 健康探针：
  - dev Compose 增加与 e2e 同步的健康检查（不更改端口映射）
- 文档修正：
  - `docs/reference/03-API-AND-TOOLS-GUIDE.md` 指标与路径索引更新为现有代码路径（`internal/organization/utils/metrics.go`）
  - 如涉及 Implementation Inventory，保持引用一致
- 单体门禁（与 202 对齐）：
  - 运行时自检：默认禁用 8090（单体模式仅使用 9090）；如需过渡测试，仅限本地环境，CI 禁止（以守卫脚本与运行时校验落实）。注意：`ENABLE_LEGACY_DUAL_SERVICE` 仅用于 CI Gate 检测，不作为运行时开关；本地过渡请通过独立容器启动查询服务（:8090），单体仍使用 :9090

---

## 6. 验收步骤与脚本

6.1 快速校验（本地/CI 一致）
```bash
# 启动基础设施与服务（Docker 强制）
make docker-up
make run-dev

# 健康检查（应返回 200 或 206），单体主路径（:9090）
curl -s http://localhost:9090/health | jq '.status,.summary'

# 指标检查（存在 http_requests_total；业务指标可按需触发）
./scripts/quality/validate-metrics.sh
# 严格模式（校验 HELP/TYPE 行）
STRICT=true ./scripts/quality/validate-metrics.sh
```

6.2 统一健康验证
```bash
# 记录健康响应到证据目录
ts=$(date +%Y%m%d-%H%M%S)
mkdir -p logs/plan251
curl -s http://localhost:9090/health | tee logs/plan251/health-command-$ts.json | jq -r '.status'
```

6.3 业务指标触发（示例）
```bash
# 触发后重查指标
./scripts/quality/validate-metrics.sh
# 可选：校验路由模板而非原始 Path（应出现模板或受控常量）
curl -s http://localhost:9090/metrics | grep -E '^http_requests_total\\{.*route=.*(\\/graphql|\\/api\\/v1\\/[^=]*\\{[^}]+\\}).*\\}'
```

---

## 7. 风险与回滚
- 风险：改造 /health 可能影响外部探针；回滚：临时保留旧路由为别名（仅开发环境），生产只保留规范端点
- 风险：指标标签变更影响已有仪表盘；回滚：保留兼容标签的迁移期窗口（以文档约定为准）
- 风险：配置统一导致环境变量不兼容；回滚：保留旧变量读取但发出警告（限一个迭代，需 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注）
- 风险：DB 指标在未接入 `pkg/database` 包装或未定期上报连接池时数据为空；回滚：先验收 HELP/TYPE 存在，再逐步补齐数据点

---

## 8. 里程碑（建议）
- M1：/health 统一与状态码规范合入（两端）→ 验收日志落盘
- M2：指标标签统一、文档修正 + DB 指标注册 → 指标校验通过
- M3：运行时配置一致化（query 侧迁移）→ sources/overrides 记录与验证
- M4：dev Compose 健康探针对齐 → `make status` 输出新增检查项

---

## 9. 单一事实来源索引（与实现对齐）
- 健康实现：`internal/monitoring/health/health.go`、`internal/monitoring/health/reporter.go`
- 指标实现：`internal/organization/utils/metrics.go`；outbox 指标补充：`cmd/hrms-server/command/internal/outbox/metrics.go`
- 认证配置：`internal/config/jwt.go`；调度器配置：`internal/config/scheduler.go`
- 运行与探针：`docker-compose.dev.yml`、`docker-compose.e2e.yml`
- 验收脚本：`scripts/quality/validate-metrics.sh`、`scripts/quality/validate-health.sh`
- 专属指标索引（示例，分类归档，避免第二事实来源）：
  - 查询服务：`cmd/hrms-server/query/internal/app/app.go` → `organization_operations_total{operation}`
  - Outbox：`cmd/hrms-server/command/internal/outbox/metrics.go`（内部计数）与 `internal/organization/utils/metrics.go`（统一 `outbox_dispatch_total`）

---

维护说明：本文件为原则与索引唯一事实来源；仅在原则或索引变更时更新。
