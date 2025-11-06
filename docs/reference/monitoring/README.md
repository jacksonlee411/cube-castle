# 监控与告警体系（219D3 预备）

本目录承载 Scheduler/Temporal 监控栈（Prometheus/Grafana/Alertmanager）的唯一事实来源，所有指标、面板、告警规则与验证记录须在此登记，并在 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`internal/organization/README.md#scheduler` 引用。

## 结构

```
monitoring/
├── promQL-snippets.md              # 常用查询语句
├── sandbox-validation.md           # sandbox 演练记录（引用 `logs/219D3/VALIDATION-2025-11-06.md`）
├── prometheus/
│   ├── README.md                   # 抓取配置与 job 列表
│   └── scheduler-prometheus.yml    # Prometheus 主配置（219D3 输出）
├── grafana/
│   ├── README.md                   # Dashboard 规范
│   └── scheduler-dashboard.json    # 计划 219D3 输出
└── alertmanager/
    ├── README.md                   # 告警路由/规约
    └── scheduler.yml               # Alertmanager 规则（219D3 输出）
```

> ⚠️ **单一事实来源要求**：Prometheus/Grafana/Alertmanager 相关文件只允许出现在本目录；若需在其他文档引用，必须指向此处路径。

## Docker Compose 与端口
- Prometheus：`prom/prometheus:v2.54.1`，端口 `9091`，数据卷 `prometheus_data`
- Grafana：`grafana/grafana:11.0.0`，端口 `3001`，数据卷 `grafana_data`
- Alertmanager：`quay.io/prometheus/alertmanager:v0.27.0`，端口 `9093`

219D3 将在 `docker-compose.dev.yml`、`docker-compose.e2e.yml` 中新增上述服务并绑定 `.env`/`config/` 配置；验证脚本与截图需同步到 `sandbox-validation.md`。
