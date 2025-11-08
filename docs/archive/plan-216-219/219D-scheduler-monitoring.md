# Plan 219D – Scheduler / Temporal 迁移与监控完善

**文档编号**: 219D  
**关联路线图**: Plan 219  
**依赖子计划**: 219A 目录、219B 查询、219C 审计/验证  
**目标周期**: Week 4 Day 21-22（紧随 219C，衔接 204 行动 2.9）  
**负责人**: 后端团队 + 平台团队

---

## 总体目标

- 在统一目录下完成 Scheduler/Temporal 迁移（落地于 `internal/organization/scheduler/*`），保持行为等价与可回退。
- 形成覆盖配置、监控、测试、文档的闭环输入，为 219E E2E 验收提供基础。
- 保持唯一事实来源：代码归属 `internal/organization/scheduler/`，配置归属 config 体系，监控/文档进入既有权威文件。

---

## 子任务索引

| 子任务 | 负责人 | 核心交付 | 计划链接 |
|--------|--------|----------|-----------|
| 219D1 – 代码迁移 | 后端 | Scheduler 目录与 Facade、回退说明 | [docs/development-plans/219D1-scheduler-migration.md](219D1-scheduler-migration.md) |
| 219D2 – 配置集中化 | 后端 | 调度配置清单、启动流程更新 | [docs/development-plans/219D2-scheduler-config.md](219D2-scheduler-config.md) |
| 219D3 – 监控与告警 | 平台+后端 | 指标埋点、Grafana 面板、告警规则 | [docs/development-plans/219D3-scheduler-monitoring.md](219D3-scheduler-monitoring.md) |
| 219D4 – 测试与故障注入 | 后端 | 单元/集成测试、故障脚本、验证报告 | [docs/development-plans/219D4-scheduler-testing.md](219D4-scheduler-testing.md) |
| 219D5 – 文档收敛 | 后端+文档 | README、参考文档、复盘清单 | [docs/development-plans/219D5-scheduler-docs.md](219D5-scheduler-docs.md) |

---

## 依赖与协作

- **前置依赖**：219A 目录调整提供结构约束；219B 查询、219C 审计输出校验策略与数据点。
- **外部协作**：平台团队支持 219D3 监控落地与 sandbox 告警校验；测试团队在 219D4 阶段参与故障注入演练。
- **回退策略**：219D1 记录的旧目录引用与配置默认值需在 219D5 汇总，确保任意阶段可回滚。

---

## 里程碑与交付节奏

1. Day 21 上午完成 219D1、219D2 主要改动并进行初步冒烟。（✅ 219D1、219D2 均已在 2025-11-06 完成并通过验收，日志与测试记录存档于 `logs/219D1/`、`logs/219D2/`。）
2. Day 21 下午至 Day 22 上午完成 219D3 指标落地与 sandbox 联调。
3. Day 22 同步推进 219D4 测试矩阵与故障注入，输出验证报告。
4. Day 22 末由 219D5 汇总文档、复盘成果，并将资料纳入参考文档。

---

## 最终收敛与验收

- [ ] 219D1~219D5 均完成并通过各自验收标准，交付物归档。
- [ ] `internal/organization/scheduler/` 目录、配置、监控、测试、文档形成闭环并对齐唯一事实来源。
- [ ] Sandbox 验证记录、告警演练、回退指南在 219D5 文档中可追溯，供 219E 继续复用。
