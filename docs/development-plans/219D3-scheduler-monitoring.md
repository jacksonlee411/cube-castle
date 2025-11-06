# Plan 219D3 – 调度监控与告警完善

**文档编号**: 219D3  
**关联路线图**: Plan 219 → 219D  
**依赖子计划**: 219D1 代码迁移、219D2 配置集中化  
**目标周期**: Week 4 Day 22（与平台团队协同）  
**负责人**: 平台团队（主导）+ 后端团队（支持）

---

## 1. 目标

1. 为 Scheduler/Temporal 工作流注册核心指标（耗时、状态、失败率、队列积压）。
2. 更新 Prometheus 抓取配置、Grafana 面板、Alertmanager 告警，形成可视化与告警闭环。
3. 验证监控数据在 sandbox 环境可查询、告警可触发。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| 指标埋点 | Go 代码中注册/更新 Prometheus 指标 |
| 监控配置 | Prometheus 抓取配置、Grafana Dashboard、Alertmanager 规则 |
| 验证 | sandbox 环境联调，记录数据样例与告警验证步骤 |

不包含：深度故障测试（219D4）、文档汇总（219D5）。

---

## 3. 详细任务

1. **指标设计**
   - 与后端梳理关键指标列表：流程耗时、失败次数、队列长度、重试次数等。
   - 明确复用 `internal/organization/utils/metrics.go` 既有注册器，必要时新增 label 或独立指标，避免重复注册导致 `prometheus: duplicate metrics collector registration attempted`。
   - 选择合适类型（Counter/Histogram/Gauge）并定义 metric name、labels，遵循项目命名规范。

2. **指标实现**
   - 在新目录的 Facade/活动中注册指标；确保 `init()` 或 dependency wiring 中完成指标注册。
   - 如需新增 collector，统一放在现有 `internal/monitoring/`（若需子目录则本计划内创建 `internal/monitoring/organization/metrics.go` 并在 README 关联），并与 `internal/organization/utils/metrics.go` 保持同步记录，确保唯一事实来源。

3. **监控配置更新**
   - 在现有配置体系内追加抓取目标：优先更新 `internal/config/`、`.env.example`、`docker-compose.*` 中的 Prometheus 相关项，并在 219D2 输出的 Scheduler 配置清单内登记链接，不另起 `deploy/monitoring/` 平行目录；若需新增 Prometheus/Grafana/Alertmanager 容器，统一在 `docker-compose.dev.yml`（及必要的 `docker-compose.e2e.yml`）中新增服务并纳入 `make docker-up`，镜像/端口规范如下：`prom/prometheus:v2.54.1`（端口 9091）、`grafana/grafana:11.0.0`（端口 3001）、`quay.io/prometheus/alertmanager:v0.27.0`（端口 9093），持久化卷统一命名为 `prometheus_data`、`grafana_data`。更新 Makefile 目标（`docker-up`、`run-dev`) 以确保新服务自动启动，并在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 记录端口说明。
   - 新建/更新 Grafana Dashboard JSON，文件纳入 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 引用并在 `internal/organization/README.md#scheduler` 标注路径，覆盖关键图表：成功率、失败率、耗时分布、队列积压；同时将 Dashboard 源文件放入 `docs/reference/monitoring/grafana/`（本计划建立该目录并在 `docs/reference/README.md` 添加索引，目录下新增 `README.md` 说明命名规范与导入方式）。
   - 撰写 Alertmanager 规则并落在 `docs/reference/monitoring/alertmanager/`（本计划负责创建目录及 README 链接，明确文件命名 `scheduler.yml`、`README.md` 记录触发条件与回滚流程），例如失败率 >1% 持续 5 分钟、队列积压超阈值；规则更新需同步 `internal/organization/README.md#scheduler` 与 `docs/reference/03-API-AND-TOOLS-GUIDE.md`。

4. **联调验证**
   - 在 sandbox 启动 Docker Compose（包含新增的 Prometheus/Grafana/Alertmanager 服务），并与平台团队确认所使用的 Compose 文件与本地一致，必要时提交变更窗口申请；运行 workflow 触发指标。
   - 确认 Prometheus 可查询、Grafana 图表更新；手动制造失败触发告警并记录时间线，并将验证脚本/命令整理入 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 监控章节，同时记录 sandbox 与本地差异。

---

## 4. 验收标准

- [ ] Prometheus 指标在本地/sandbox 可观测，命名、标签符合规范。
- [ ] Grafana Dashboard 与 Alertmanager 规则更新完成，截图或链接存档，并在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 与 `internal/organization/README.md#scheduler` 引用。
- [ ] 告警触发与恢复流程经验证并记录操作步骤。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 指标设计不合理导致噪声或缺口 | 中 | 邀请后端、平台共同评审指标列表 |
| 监控配置改动影响现有面板 | 中 | 在新 Dashboard 上迭代，保留旧面板副本；版本化配置 |
| sandbox 环境权限不足 | 中 | 提前预约平台环境窗口；必要时使用本地 Docker stack 验证 |

---

## 6. 交付物

- 新增/更新的监控指标代码片段。
- Prometheus/Grafana/Alertmanager 配置文件或变更摘要，均指向既有配置目录并在 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`internal/organization/README.md` 登记。
- sandbox 验证记录（截图、日志、告警邮件/消息）。
