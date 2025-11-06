# Grafana Dashboard 规范（219D3 预设）

- 容器：`grafana/grafana:11.0.0`
- 端口：`3001`
- 数据卷：`grafana_data`
- 默认登录：`admin / admin`（首次启动后请修改）

## Dashboard 命名规范

| 文件名 | 用途 | 引用位置 |
|--------|------|----------|
| `scheduler-dashboard.json` | Scheduler/Temporal 指标 | `docs/development-plans/219D3-scheduler-monitoring.md`、`internal/organization/README.md` |

- 每个 Dashboard JSON 必须写入标题、描述、面板标签，便于导入。
- 面板至少涵盖：任务成功率、失败率、活动耗时直方图、队列长度、重试次数。

## 导入导出流程
1. `make docker-up monitoring` 启动 Grafana
2. 登录后导入 JSON（保持 UID 与 `scheduler` 前缀）
3. 导出更新后 JSON 并覆盖本目录文件
4. 在 `sandBox-validation.md` 记录截图与验证步骤
