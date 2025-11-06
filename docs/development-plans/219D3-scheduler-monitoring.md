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
   - 选择合适类型（Counter/Histogram/Gauge）并定义 metric name、labels，遵循项目命名规范。

2. **指标实现**
   - 在新目录的 Facade/活动中注册指标；确保 `init()` 或 dependency wiring 中完成指标注册。
   - 更新 `internal/observability/`（若有）或新增 Prometheus collector。

3. **监控配置更新**
   - 修改 `deploy/monitoring/prometheus.yml`（或对应路径）增加抓取目标。
   - 新建/更新 Grafana Dashboard JSON，覆盖关键图表：成功率、失败率、耗时分布、队列积压。
   - 撰写 Alertmanager 规则：如失败率 >1% 持续 5 分钟触发；队列积压超阈值告警。

4. **联调验证**
   - 在 sandbox 启动 Docker Compose，运行 workflow 触发指标。
   - 确认 Prometheus 可查询、Grafana 图表更新；手动制造失败触发告警并记录时间线。

---

## 4. 验收标准

- [ ] Prometheus 指标在本地/sandbox 可观测，命名、标签符合规范。
- [ ] Grafana Dashboard 与 Alertmanager 规则更新完成，截图或链接存档。
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
- Prometheus/Grafana/Alertmanager 配置文件或变更摘要。
- sandbox 验证记录（截图、日志、告警邮件/消息）。
