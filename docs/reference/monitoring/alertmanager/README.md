# Alertmanager 规则指南（219D3 预设）

- 镜像：`quay.io/prometheus/alertmanager:v0.27.0`
- 端口：`9093`
- 默认路由配置文件：`scheduler.yml`（计划 219D3 输出）

## 告警命名规范
`<domain>_<metric>_<condition>`，例如：
- `SCHEDULER_TASK_FAILURE_RATE_HIGH`
- `SCHEDULER_QUEUE_BACKLOG_EXCEEDS_THRESHOLD`

## 配置要素
- `receivers`: `pagerduty`, `slack`, `email` 可按环境裁剪，默认提供 `slack-monitoring`（webhook 从 secrets 读取）
- `route.group_by`: `['alertname','task']`
- `inhibit_rules`: 防止重复告警（例如失败率触发后禁止 backlog 告警重复推送）

## 测试要求
1. Prometheus 指向 Alertmanager：`alerting.alertmanagers` -> `alertmanager:9093`
2. 使用 `amtool` 或 `curl` 触发示例告警
3. 在 `sandbox-validation.md` 记录通知截图/日志
