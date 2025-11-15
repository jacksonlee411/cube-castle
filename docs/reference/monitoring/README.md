# 监控与告警体系（如已启用）

> 说明：项目容器化与强制约束以 `AGENTS.md` 为唯一事实来源。监控栈（Prometheus/Grafana/Alertmanager）为可选能力，按计划落地（如 219D3）。当前仓库默认的 docker-compose 未必包含监控服务；若需启用，请依据对应计划文件在本目录维护权威配置与验证材料。

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

## Docker Compose 与端口（范例）
- Prometheus：`prom/prometheus:v2.54.1`，端口建议 `9091`，数据卷建议 `prometheus_data`
- Grafana：`grafana/grafana:11.0.0`，端口建议 `3001`，数据卷建议 `grafana_data`
- Alertmanager：`quay.io/prometheus/alertmanager:v0.27.0`，端口建议 `9093`

如需集成，请在计划文档与 Compose 文件中显式新增服务，并在本目录登记配置与验证证据（如 `sandbox-validation.md`）；严禁为迁就宿主机端口而修改容器端口映射，端口冲突须通过卸载宿主服务解决（参见 AGENTS.md）。
