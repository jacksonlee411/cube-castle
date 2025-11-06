# PromQL 速查（Scheduler / Temporal）

| 场景 | 查询示例 | 说明 |
|------|----------|------|
| 任务成功率 | `sum(rate(scheduler_task_executions_total{status="success"}[5m])) / sum(rate(scheduler_task_executions_total[5m]))` | 计算过去 5 分钟任务成功率 |
| 失败速率 | `rate(scheduler_task_executions_total{status="failed"}[5m])` | 监控失败趋势，可结合 Alertmanager 阈值 |
| 任务耗时 | `histogram_quantile(0.95, rate(scheduler_task_execution_duration_seconds_bucket[5m]))` | 观察 P95 耗时 |
| 告警触发量 | `rate(scheduler_monitor_alerts_total[10m])` | 与 Alertmanager 规则对应，可判断噪声 |
| 下一次运行时间 | `scheduler_task_next_run_timestamp_seconds{task="daily_cutover"}` | 配合 Grafana 显示 Cron 计划 |

> 验证参考：`logs/219D3/VALIDATION-2025-11-06.md`
