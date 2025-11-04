# Plan 219D – Scheduler / Temporal 迁移与监控完善

**文档编号**: 219D  
**关联路线图**: Plan 219  
**依赖子计划**: 219A 目录、219B 查询、219C 审计/验证  
**目标周期**: Week 4 Day 1-2  
**负责人**: 后端团队 + 平台团队

---

## 1. 目标

1. 将 `organization_temporal_service.go`、`operational_scheduler.go` 等调度逻辑迁移到 `internal/organization/scheduler/`，统一管理。
2. 确保 Temporal 工作流、定时任务（position version 激活、timeline 修复等）在新结构下行为等同并通过测试。
3. 完成监控与告警配置：Prometheus 指标、Grafana 面板、Alertmanager 规则。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| Scheduler 目录 | Temporal workflow / activity / cron jobs；操作性任务（operational endpoints） |
| 配置 | YAML/ENV 中的调度配置、Cron 表达式、重试策略 |
| 监控 | Prometheus 指标注册、Grafana 面板、告警阈值 |
| 文档 | README 更新、运行手册 |

不包含：Assignment 查询与缓存逻辑（219B）、新增业务功能。

---

## 3. 详细任务

1. **代码迁移**
   - 将 Temporal workflow、activity、scheduler 逻辑从 `cmd/hrms-server/command/internal/...` 迁至 `internal/organization/scheduler/`。
   - 更新依赖注入：`main.go` 在启动时注册 scheduler，使用 Facade/Service 代替旧引用。

2. **配置梳理**
   - 统一调度相关配置（cron 间隔、队列名称、重试次数）到 config 包或 `.env`；记录默认值。
   - 更新 Makefile / README，说明启动 scheduler 的方式。

3. **监控与告警**
   - 注册指标：`temporal_workflow_duration_ms`、`temporal_workflow_failure_total`、`organization_event_dispatch_total` 等。
   - 更新 `prometheus.yml` / Grafana 面板，新增监控图表。
   - 编写 alert 规则（如失败率 > 1% 告警）。

4. **测试验证**
   - 单元测试：使用 Temporal SDK 测试工具对 workflow/activity 进行模拟，断言 retry/backoff 逻辑。
   - 集成测试：在 sandbox 环境跑一次完整的 workflow（position version 激活），记录任务队列、执行日志、指标。
   - 故障注入：模拟 activity 超时/失败，验证 retry/backoff 与告警触发。
   - 输出验证 checklist（workflow 名称、预期事件、关键指标阈值）。

5. **文档**
   - 更新 `internal/organization/README.md`：调度逻辑、配置、监控说明。
   - 输出运行手册（启动、停止、排错指南）。

---

## 4. 验收标准

- [ ] Scheduler/Temporal 代码迁移完成，旧目录无残留。
- [ ] 工作流/定时任务在新目录下可构建并运行（本地 + sandbox）。
- [ ] 指标在 Prometheus 中可查询，Grafana 告警生效。
- [ ] README/运行手册更新完毕。
- [ ] 测试脚本（单元/集成）通过；提供示例命令。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Workflow 行为改变 | 高 | 保留对照测试；执行真实数据的 dry-run |
| 监控配置遗漏 | 中 | 与平台团队审查；在测试环境验证告警 |
| 配置变更影响生产 | 中 | 变更前记录旧值，必要时灰度发布 |

---

## 6. 交付物

- Scheduler 目录及代码
- 配置/README 更新
- Prometheus/Grafana/Alertmanager 配置
- 测试报告（workflow + 监控）
