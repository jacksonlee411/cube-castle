# 219E – Outbox/Dispatcher 指标验证 Runbook（2025-11-08）

## 1. 目标
- 通过真实命令流量触发 `outbox_events`，观察 dispatcher 队列、Prometheus 指标与 Query 读模型刷新，解除 Plan 219E §2.4 “Outbox → Dispatcher → Query 缓存” 阻塞。
- 产出可复现的日志与 PromQL 记录，供 Plan 06/219E 阶段验收引用。

## 2. 唯一事实来源
- Dispatcher 配置与指标：`cmd/hrms-server/command/internal/outbox/config.go`、`cmd/hrms-server/command/internal/outbox/dispatcher.go`、`cmd/hrms-server/command/internal/outbox/metrics.go`。
- 指标注册：`internal/organization/utils/metrics.go`（`outbox_dispatch_total` 及 HTTP/Audit 计数器）。
- 触发脚本：`scripts/219C3-rest-self-test.sh`（命令场景注入 + requestId 记录）。

## 3. 执行步骤
| 步骤 | 操作 | 产出 |
| --- | --- | --- |
| O1 | **环境预检**：`make docker-up && make run-dev`（确认命令服务日志 `outbox dispatcher started`），`make jwt-dev-mint` 更新 `.cache/dev.jwt`，`export PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`。 | `run-dev-command.log` 片段 |
| O2 | **触发 Outbox 事件**：运行 `./scripts/219C3-rest-self-test.sh BASE_URL_COMMAND=http://localhost:9090 REQUEST_PREFIX=outbox-e2e`；脚本会执行 Position Fill/Vacate/Assignment Close 等命令。记录输出中的 `requestId` 并附加至 `logs/219C3/validation.log`。 | 现有 `logs/219C3/validation.log`、`logs/219C3/report.json`；新建 `logs/219E/outbox-dispatcher-events-<timestamp>.log`（截取 requestId + eventType） |
| O3 | **数据库取证**：连接 Postgres（`psql "$DATABASE_URL"`）执行：<br>`SELECT event_id, event_type, aggregate_id, retry_count, available_at, published_at FROM outbox_events ORDER BY created_at DESC LIMIT 20;`<br>将结果保存为 `logs/219E/outbox-dispatcher-sql-<timestamp>.log`。 | 验证事件生成/已发布状态 |
| O4 | **Prometheus 指标抓取**：命令服务启动时默认注册 `outbox_dispatch_*` counter/gauge。执行：<br>`curl -s http://localhost:9090/metrics | rg 'outbox_dispatch'`<br>并保存为 `logs/219E/outbox-dispatcher-metrics-<timestamp>.log`。<br>PromQL 建议：<br>`outbox_dispatch_success_total`、`outbox_dispatch_failure_total`、`rate(outbox_dispatch_retry_total[5m])`、`outbox_dispatch_active`. | 指标日志 + PromQL 记录 |
| O5 | **日志核查**：`docker logs cubecastle-command | rg 'outbox dispatcher' -A2`，确保每批发布成功/失败均有记录；保存至 `logs/219E/outbox-dispatcher-run-<timestamp>.log`。 | Dispatcher 运行证据 |
| O6 | **Query 缓存刷新验证**：对步骤 O2 创建/关闭的职位运行 GraphQL：`graphql-client --endpoint=http://localhost:8090/graphql --query-file tests/e2e/fixtures/positions.graphql > logs/219E/position-gql-outbox-<timestamp>.log`，核对 Assignment 时间线是否反映最新事件。 | 读模型同步日志 |
| O7 | **Prometheus 图表**（可选）：若 `monitoring-prometheus`/`monitoring-grafana` 已启动，使用 `promtool query instant` 或 Grafana 面板记录 `PNG`/`CSV`，路径建议 `logs/219E/outbox-dispatcher-promql-<timestamp>.txt`。 | 图表/Query 导出 |

## 4. 指标说明
- `outbox_dispatch_success_total` / `_failure_total` / `_retry_total`：来源 `cmd/hrms-server/command/internal/outbox/metrics.go`，前缀受 `OUTBOX_DISPATCH_METRIC_PREFIX` 控制（默认 `outbox_dispatch`）。
- `outbox_dispatch_active`：dispatcher 轮询状态 gauge，可用于判断是否卡死。
- `outbox_dispatch_total{result,event_type}`：`internal/organization/utils/metrics.go` 统一注册的 counter，若调用 `utils.RecordOutboxDispatch`（后续若自动派发 assignment cache 刷新），需同步观察此指标。
- 组合 PromQL 示例：<br>`sum(increase(outbox_dispatch_success_total[5m])) by ()`、`sum by (event_type)(increase(outbox_dispatch_total{result="success"}[10m]))`。

## 5. 产出与归档
- `logs/219E/outbox-dispatcher-events-<timestamp>.log`
- `logs/219E/outbox-dispatcher-sql-<timestamp>.log`
- `logs/219E/outbox-dispatcher-metrics-<timestamp>.log`
- `logs/219E/outbox-dispatcher-run-<timestamp>.log`
- `logs/219E/outbox-dispatcher-promql-<timestamp>.txt`（可选）
- `logs/219E/position-gql-outbox-<timestamp>.log`

> 完成上述取证后，将日志路径回填至 `docs/development-plans/219E-e2e-validation.md` §2.4 与 `docs/development-plans/06-integrated-teams-progress-log.md` Section 6，作为 Outbox/Dispatcher 验收证据。
