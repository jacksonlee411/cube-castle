# Plan 219 – Organization 模块重构路线图 (索引版)

**文档编号**: 219  
**最后更新**: 2025-11-04  
**维护者**: Codex（AI 助手）

本文件仅作为 219 系列子计划（219A~219E）的索引、时间表与风险提示。具体实施步骤、验收标准请参见对应子计划文档。

---

## 1. 总览

- 目标：将现有 `organization` 模块按 Phase2 模板重构，保持契约不变并补齐查询、审计、调度、测试体系。
- 子计划：
  - **219A** – 目录重构与 Facade 基线（3 天）
  - **219B** – Assignment 查询链路与缓存刷新（1 天）
  - **219C** – Audit & Validator 规则收敛（2 天）
  - **219D** – Scheduler/Temporal 迁移与监控完善（2 天）
  - **219E** – 端到端测试与性能验收（3 天）

---

## 2. 子计划索引

| 子计划 | 主题 | 主要交付 | 依赖 |
|--------|------|----------|------|
| [219A-directory-refactor](219A-directory-refactor.md) | 目录/Facade/迁移清单 | 新结构、api.go、适配层 | 基础设施 (216-218) |
| [219B-assignment-query](219B-assignment-query.md) | Assignment 查询 + 缓存刷新 | 查询仓储、GraphQL resolver、缓存策略 | 219A |
| [219C-audit-validator](219C-audit-validator.md) | 审计 & 业务验证规则 | Audit/Validator 实现、规则清单 | 219A |
| [219D-scheduler-monitoring](219D-scheduler-monitoring.md) | Temporal/Scheduler 迁移 + 监控 | Scheduler 目录、指标、告警 | 219A,219B,219C |
| [219E-e2e-validation](219E-e2e-validation.md) | E2E + 性能验收 + 回退 | 测试/性能报告、回退演练 | 219A~219D |

---

## 3. 时间轴（建议）

| 周次 | 日程 | 里程碑 | 说明 |
|------|------|--------|------|
| **Week 3** | Day 1-3 | 完成 219A | 目录迁移、api.go、迁移清单 v1 | 
| | Day 4 | 219B | Assignment 查询链路、缓存刷新 |
| | Day 5 – Week 4 Day 1 | 219C | Audit & Validator 规则落地 |
| **Week 4** | Day 2-3 | 219D | Scheduler/Temporal 迁移 + 监控 |
| | Day 4-5 | 219E | 端到端验收、性能基准、回退演练 |

> 时间表可根据实际执行情况调整，确保前置依赖达成后再启动下一子计划。

---

## 4. 依赖与前置条件

- Plan 216（eventbus）、217（database/outbox）、217B（dispatcher）、218（logger）已交付并可用。
- 契约遵循 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，Department 仍为 Organization 聚合内节点（`unitType=DEPARTMENT`）。
- 启动各子计划前请确认对应 P0 补充项已完成：
  - 审计/验证规则清单（219C 输出）
  - Department 聚合说明（219A 输出）
  - Temporal 测试方案、性能脚本模板（219D/219E 输出）

---

## 5. 风险概要

| 风险 | 影响 | 对策 |
|------|------|------|
| 迁移量大导致 219A 超期 | 高 | 预留 3 天 + 每日同步；遇循环依赖及时拆分 |
| Assignment 查询或缓存遗漏 | 中 | 219B 完成后进行冒烟/端到端验证 |
| Audit/Validator 漏项 | 高 | 规则清单由架构/安全共同评审 | 
| Temporal 行为变更 | 高 | 219D 在 sandbox 对照运行，保留回退脚本 |
| 性能退化 | 中 | 219E 对比基线，预留优化时间 |

---

## 6. 文档与交付物（全局）

- 子计划文档（219A~219E）持续更新。
- 迁移清单、审计/验证规则、监控/告警配置、测试/性能报告均需落库。
- `internal/organization/README.md` 作为唯一事实来源，统一记录聚合边界、依赖、调试方式。

---

**备注**: 任何变更若产出附加资料（例如 `audit-rules.md`、`validator-rules.md`、性能对比报告），请在相应子计划和 README 中引用。
