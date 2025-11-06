# Plan 219 – Organization 模块重构路线图 (索引版)

**文档编号**: 219  
**最后更新**: 2025-11-06  
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
| **Week 3** | Day 15-17 | 完成 219A | 目录迁移、api.go、迁移清单 v1（对应 204 计划行动 2.6） |
| **Week 4** | Day 18 | 219B | Assignment 查询链路、缓存刷新（与行动 2.7/2.8 并行可行） |
|  | Day 19-20 | 219C | Audit & Validator 规则落地，含规则清单入库 |
|  | Day 21-22 | 219D | Scheduler/Temporal 迁移 + 监控配置 |
| **Week 5** | Day 23-25 | 219E | 端到端验收、性能基准、回退演练；Day 26 预留缓冲 |

> 时间表可根据实际执行情况调整，确保前置依赖达成后再启动下一子计划。

> 进度更新：2025-11-05 — ✅ 219B 已完成并合入主干，Assignment 查询能力与缓存刷新策略生效。  
> 进度更新：2025-11-06 — ✅ 219D1/219D2 完成验收（详见 `docs/development-plans/06-integrated-teams-progress-log.md` 与 `logs/219D2/ACCEPTANCE-RECORD-2025-11-06.md`），Scheduler 目录迁移与配置集中化达成，为 219D3 监控落地准备就绪。  
> 进度更新：2025-11-06 — ✅ 219D3（监控与告警）完成，指标、Dashboard、Alertmanager 规则与验证记录已入库（参见 `docs/reference/monitoring/`、`logs/219D3/VALIDATION-2025-11-06.md`）。

---

## 4. 依赖与前置条件

- Plan 216（eventbus）、217（database/outbox）、217B（dispatcher）、218（logger）已交付并可用。
- 契约遵循 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，Department 仍为 Organization 聚合内节点（`unitType=DEPARTMENT`）。
- 启动各子计划前请确认对应 P0 补充项已完成，并保证唯一事实来源：
  - 审计/验证规则清单：由 219C 更新 `internal/organization/README.md` 的“审计与业务规则”章节并同步引用 `docs/reference/`（避免新增散落文件）。
  - Department 聚合说明：在 219A 内更新 `internal/organization/README.md` 的聚合边界。
  - Temporal 测试方案、性能脚本模板：219D/219E 输出统一存放于 `tests/organization/` 与 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 对应章节。

---

## 5. 风险概要

| 风险 | 影响 | 对策 |
|------|------|------|
| 迁移量大导致 219A 超期 | 高 | 预留 Day 26 缓冲 + 每日同步；遇循环依赖及时拆分 |
| Assignment 查询或缓存遗漏 | 中 | 219B 完成后进行冒烟/端到端验证 |
| Audit/Validator 漏项 | 高 | 规则清单由架构/安全共同评审并更新唯一事实来源 | 
| Temporal 行为变更 | 高 | 219D 在 sandbox 对照运行，保留回退脚本 |
| 性能退化 | 中 | 219E 对比基线，预留优化时间 |

---

## 6. 文档与交付物（全局）

- 子计划文档（219A~219E）持续更新。
- 迁移清单、审计/验证规则、监控/告警配置、测试/性能报告均需落库。
- `internal/organization/README.md` 作为唯一事实来源，统一记录聚合边界、依赖、调试方式。

---

**备注**: 任何变更若产出附加资料（例如 `audit-rules.md`、`validator-rules.md`、性能对比报告），请在相应子计划和 README 中引用。
