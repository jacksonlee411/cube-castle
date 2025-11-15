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
- /health 统一：返回结构与状态码规范一致；字段 camelCase；command 与 query 均接入统一健康管理器；
  - 状态码：healthy=200、degraded=206、unhealthy=503（来源：`internal/monitoring/health/health.go` Handler 实现）
  - 必含字段：`service, version, status, timestamp, uptime, checks[], summary{total,healthy,degraded,failed}`
- /metrics 可达且指标命名/标签一致：`http_requests_total{method,route,status}`、`temporal_operations_total{operation,status}`、`audit_writes_total{status}`、`outbox_dispatch_total{result,event_type}`
- 运行时配置来源单一：JWT/调度器等使用 `internal/config/*`；DB/Redis 配置覆盖规则一致（默认→文件→环境），记录覆盖元数据（sources/overrides）
- Docker 健康探针：dev 与 e2e Compose 对 rest/query 服务的探针语义一致（探测 `/health` 返回可接受状态），不可更改端口映射
- 证据：`logs/plan251/health-*.json|txt`、`logs/plan251/metrics-*.txt`，由脚本生成；CI 与本地命令一致

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

4.2 指标（Prometheus）
- 唯一事实来源（命名与标签）：`internal/organization/utils/metrics.go`
  - `http_requests_total{method,route,status}`
  - `temporal_operations_total{operation,status}`
  - `audit_writes_total{status}`
  - `outbox_dispatch_total{result,event_type}`
- 端点暴露：`/metrics` 使用 `promhttp.Handler()`（command/query 一致）
- 记录位置（示例，不扩展第二事实来源，仅作索引）：
  - HTTP 计数：由服务中间件或包裹器统一打点（query 端将 `path` 统一更名为 `route`）
  - 业务计数：时态操作（scheduler/service）、审计写入、outbox 派发（dispatcher）
- 说明：Counter 在未被 Inc() 前不会出现在 `/metrics` 输出，这是 Prometheus 标准行为

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
- 运行时配置一致化：
  - JWT：rest/query 均使用 `internal/config/jwt.go`（已达成）
  - DB/Redis：在 query 侧引入与 rest 等价的加载/覆盖模式，并记录覆盖元信息
- Docker 健康探针：
  - dev Compose 增加与 e2e 同步的健康检查（不更改端口映射）
- 文档修正：
  - `docs/reference/03-API-AND-TOOLS-GUIDE.md` 指标与路径索引更新为现有代码路径（`internal/organization/utils/metrics.go`）
  - 如涉及 Implementation Inventory，保持引用一致

---

## 6. 验收步骤与脚本

6.1 快速校验（本地/CI 一致）
```bash
# 启动基础设施与服务（Docker 强制）
make docker-up
make run-dev

# 健康检查（应返回 200 或 206）
curl -s http://localhost:9090/health | jq '.status,.summary'
curl -s http://localhost:8090/health | jq '.status,.summary'

# 指标检查（存在 http_requests_total；业务指标可按需触发）
./scripts/quality/validate-metrics.sh
```

6.2 统一健康验证
```bash
# 记录健康响应到证据目录
ts=$(date +%Y%m%d-%H%M%S)
mkdir -p logs/plan251
curl -s http://localhost:9090/health | tee logs/plan251/health-command-$ts.json | jq -r '.status'
curl -s http://localhost:8090/health | tee logs/plan251/health-query-$ts.json   | jq -r '.status'
```

6.3 业务指标触发（示例）
```bash
# 触发后重查指标
./scripts/quality/validate-metrics.sh
```

---

## 7. 风险与回滚
- 风险：改造 /health 可能影响外部探针；回滚：临时保留旧路由为别名（仅开发环境），生产只保留规范端点
- 风险：指标标签变更影响已有仪表盘；回滚：保留兼容标签的迁移期窗口（以文档约定为准）
- 风险：配置统一导致环境变量不兼容；回滚：保留旧变量读取但发出警告（限一个迭代，需 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注）

---

## 8. 里程碑（建议）
- M1：/health 统一与状态码规范合入（两端）→ 验收日志落盘
- M2：指标标签统一、文档修正 → 指标校验通过
- M3：运行时配置一致化（query 侧迁移）→ sources/overrides 记录与验证
- M4：dev Compose 健康探针对齐 → `make status` 输出新增检查项

---

## 9. 单一事实来源索引（与实现对齐）
- 健康实现：`internal/monitoring/health/health.go`、`internal/monitoring/health/reporter.go`
- 指标实现：`internal/organization/utils/metrics.go`；outbox 指标补充：`cmd/hrms-server/command/internal/outbox/metrics.go`
- 认证配置：`internal/config/jwt.go`；调度器配置：`internal/config/scheduler.go`
- 运行与探针：`docker-compose.dev.yml`、`docker-compose.e2e.yml`
- 验收脚本：`scripts/quality/validate-metrics.sh`、`scripts/quality/validate-health.sh`

---

维护说明：本文件为原则与索引唯一事实来源；仅在原则或索引变更时更新。

