# Sandbox 验证记录（219D3）

- **日志来源**：`logs/219D3/VALIDATION-2025-11-06.md`
- **环境**：Docker Compose（含 monitoring-* 服务），Go 1.24.9
- **已验证项**：
  1. Prometheus 指标抓取（`scheduler_task_executions_total`、`scheduler_monitor_alerts_total`）
  2. Grafana Dashboard 导入与实时更新
  3. Alertmanager 告警触发与恢复（重复当前记录故障注入）
- **复现脚本**：
  ```bash
  make docker-up
  ./scripts/dev/scheduler-alert-smoke.sh         # 触发 CRITICAL 告警
  curl -s http://localhost:9091/api/v1/query --data-urlencode 'query=rate(scheduler_monitor_alerts_total[5m])'
  ```
