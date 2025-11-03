# 212-Shared Architecture Alignment Plan

**编号**: 212  
**标题**: Day6-7 shared/internal 复用决策与 `pkg/health` 归属审查  
**创建日期**: 2025-11-03  
**最近更新**: 2025-11-03  
**状态**: 🟡 执行中（纳入 Plan 211 Phase1 日程）  
**关联文档**:
- `211-phase1-module-unification-plan.md`（母计划）
- `06-integrated-teams-progress-log.md`（进展记录）
- `reports/phase1-module-unification.md`（执行日志）
- `reports/phase1-architecture-review.md`（Day6-7 审查资料）

---

## 1. 范围与目标

- **范围**：在 Plan 211 Phase1 框架内，聚焦 Day6-7 架构审查环节，明确以下决策并形成正式纪要：
  - command 与 shared/internal 代码复用界限、共享封装方式
  - `pkg/health` 模块归属与未来演进路径
  - 审查发现的目录或依赖问题的整改方案与责任人
- **目标**：形成唯一事实来源的决议文档，并将结论同步至 Plan 211、06 号文档及相关执行日志。

---

## 2. 关键交付物

| 序号 | 交付物 | 描述 | 归档位置 |
|------|--------|------|----------|
| D1 | 架构审查会议纪要 | Day6-7 审查会结论、责任人、截止时间 | `reports/phase1-architecture-review.md` 追加章节 |
| D2 | 目录/依赖调整清单 | 对应待办与整改计划（含 PR/脚本指引） | `reports/phase1-module-unification.md` Day7 节 |
| D3 | 文档更新记录 | Plan 211、06 号文档同步更新（引用本计划） | `docs/development-plans/06-integrated-teams-progress-log.md` |

---

## 3. 工作分解与时间线

| 日期 | 任务 | 负责人 | 说明 |
|------|------|--------|------|
| Day6 上午 | 现状复盘 | 架构师 + 后端 TL | 对照 `reports/phase1-architecture-review.md`，确认共享模块使用情况 |
| Day6 下午 | 风险梳理 | 架构师 + QA | 识别潜在循环依赖、命名漂移、测试覆盖缺口 |
| Day7 上午 | 架构审查会议 | 架构师（主持）+ Codex + 后端 TL + QA + DevOps | 形成复用策略、`pkg/health` 处置方案、后续整改清单 |
| Day7 下午 | 结论落地 | Codex + 文档支持 | 更新 `reports/phase1-module-unification.md`、Plan 211、06 文档，发布纪要 |

---

## 4. 审查要点

1. **共享模块复用**  
   - `internal/auth/config/middleware/types` 是否同时服务 command/query；如需复用，明确 API/初始化方式。  
   - 若维持双份实现，需记录差异原因与回收计划。
2. **`pkg/health` 归属**  
   - 判断是否合入 `internal/monitoring` 或保留独立包；如保留，说明未来演进路线及调用方。  
   - 确认 CI/监控脚本引用的唯一路径。
3. **依赖矩阵**  
   - 使用 `go list`/`scripts/tools/phase1-import-audit.py` 等工具验证无 legacy import。  
   - 对新共享模块输出单元测试或契约保障。
4. **整改追踪**  
   - 为必要改动定义责任人、截止时间、验证方式，记录在 `reports/phase1-module-unification.md` Day7 节。  
   - 若决议影响 API/文档，同步更新相应事实来源。

---

## 5. 沟通与审批

- 日常同步：沿用 16:00 日常站会（Codex 主持）。  
- 架构审查会：Day7 上午 10:00-11:00，会议纪要 4 小时内发布。  
- Steering 报告：如结论涉及跨计划风险或回滚方案，纳入 Day7 状态更新，由 Codex 汇报。  
- 本计划不新增决策权限矩阵；使用既有 Steering 决议流程（记录于 06 号文档）。

---

## 6. 风险与应对

| 风险 | 等级 | 说明 | 对策 |
|------|------|------|------|
| 共享模块耦合 | 高 | command/query 相互引用导致循环依赖 | 审查时强制列出依赖图，必要时拆分或加 Facade |
| 文档漂移 | 中 | 目录/策略调整未更新 Plan 211/06 | Day7 下午对齐文档支持，更新引用 |
| 决议落地延迟 | 中 | 责任人不明确或 PR 未提交 | 纪要中指明 Owner + 截止时间 + 验证方式 |

---

## 7. 验收标准

- Day7 审查纪要发布，列明所有结论与整改项。  
- `reports/phase1-module-unification.md` 更新 Day7 行动项及状态。  
- 无新的 legacy import；必要的复用/拆分计划落地或进入后续迭代（记录在 Plan 211/203）。  
- 06 号文档“后续事项 / 当前待办”已引用此计划，并更新进展。

---

**版本历史**  
- v1.0 (2025-11-03)：初始版本，纳入 Plan 211 Phase1 Day6-7 行动清单。
