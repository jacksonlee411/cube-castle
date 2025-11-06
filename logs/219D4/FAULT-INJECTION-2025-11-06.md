# 故障注入演练 – Plan 219D4

**目标：** 模拟 Scheduler 时态监控在出现“重复当前记录”时触发 CRITICAL 告警，并验证 Alertmanager 指标可被 Prometheus 抓取（依赖 219D3 配置）。

**执行命令：**
```bash
./scripts/dev/scheduler-alert-smoke.sh
```

**输出节选：**
```
[scheduler-alert-smoke] running Temporal monitor alert injection test
=== RUN   TestTemporalMonitorCheckAlertsCritical
--- PASS: TestTemporalMonitorCheckAlertsCritical (0.00s)
PASS
ok   cube-castle/internal/organization/scheduler  0.006s
```

**说明：**
- 脚本内部使用 `GOCACHE=$(mktemp -d)` 隔离 Go build cache，避免宿主环境 `~/.cache/go-build` 权限不足导致测试失败。
- 该测试协调 SQLMock 返回重复当前记录计数，引导 `TemporalMonitor.CheckAlerts` 生成 `[CRITICAL]` 告警信息；Prometheus/Alertmanager 规则在 219D3 中已定义，本演练确认业务逻辑能够触发告警。
- 如需扩展 sandbox 验证，可在执行脚本前启动 `make docker-up` 并查看 Alertmanager UI (`http://localhost:9093`) 的 active alerts。
