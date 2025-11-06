# 219D3 监控验证记录 – 2025-11-06

## 实验环境
- Docker Compose：`make docker-up`（新增 `monitoring-prometheus`、`monitoring-grafana`、`monitoring-alertmanager` 已自动启动）
- Prometheus: http://localhost:9091
- Grafana: http://localhost:3001 （Dashboard UID: `scheduler-observability`）
- Alertmanager: http://localhost:9093

## 验证步骤与结论
1. **指标曝光**  
   ```bash
   curl -s http://localhost:9090/metrics | rg \"scheduler_task_execution_duration_seconds\"
   curl -s http://localhost:9090/metrics | rg \"scheduler_monitor_alerts_total\"
   ```
   - Scheduler 任务执行耗时/状态指标均可查询；Prometheus target `rest-service:9187/metrics` 状态 `UP`。

2. **手动触发任务**  
   ```bash
   curl -s -X POST http://localhost:9090/api/v1/operational/tasks/daily_cutover/trigger
   ```
   - `scheduler_task_executions_total{task=\"daily_cutover\",status=\"success\"}` 计数 +1
   - Grafana 面板【任务成功率】实时更新。

3. **告警触发演练**  
   ```bash
   # 通过环境变量模拟队列积压阈值降低
   export SCHEDULER_MONITOR_ALERT_BACKLOG_THRESHOLD=0
   make run-dev SCHEDULER_MONITOR_ENABLED=true
   ```
   - 触发 `SCHEDULER_QUEUE_BACKLOG_EXCEEDS_THRESHOLD` 告警，Alertmanager 产生 active alert；5 分钟后恢复阈值并复位成功。

4. **PromQL 验证**  
   ```promql
   rate(scheduler_task_executions_total{status=\"failed\"}[5m])
   rate(scheduler_monitor_alerts_total[10m])
   scheduler_task_next_run_timestamp_seconds{task=\"daily_cutover\"}
   ```
   - 查询结果与预期一致，Dashboard 曲线与 Prometheus 数据对齐。

## 后续建议
- 将 sandbox 环境的 webhook URL 写入 `docs/reference/monitoring/alertmanager/README.md` 中的占位符。
- 与平台团队对齐 Alertmanager 路由策略，接入企业级告警通道。
