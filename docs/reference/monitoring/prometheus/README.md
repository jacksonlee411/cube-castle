# Prometheus 抓取配置（示例，按计划启用）

- 目标：收集 Scheduler 指标（基于 Cron/SQL 的时态数据一致性维护：任务耗时、失败率、自定义验证链等）。
- 默认端口：`9091`
- 数据卷：`prometheus_data`
- Docker Compose 服务名建议：`monitoring-prometheus`

## 配置说明（按需启用）
- `scheduler-prometheus.yml`（按计划提交）可定义：
  - `scrape_interval: 15s`
  - `static_configs` 指向 `rest-service:9090/metrics`（命令服务）与 `graphql-service:8090/metrics`
  - 可选 `alerting` 块指向 `alertmanager:9093`
- 环境变量通过 `.env.example` 暴露 `PROMETHEUS_RETENTION=15d`、`PROMETHEUS_STORAGE_PATH=/prometheus`，由 Compose 注入。

## 验证步骤（启用后）
1. `make docker-up`（或在 Compose 中显式增加 monitoring 服务后启动）
2. 访问 `http://localhost:9091` 验证 targets、alerts 均为 `UP`
3. 在 `promQL-snippets.md` 记录查询命令与预期输出
