# 99号文档：职位管理相关计划收口顺序建议

**版本**: v0.3
**创建日期**: 2025-10-21
**最新更新**: 2025-10-21
**状态**: 执行中（收尾协调）
**编制人**: Claude Code
**关联计划**: 80、86、87、88、93、95、96、101、102、103、104 号文档
**遵循原则**: `CLAUDE.md` 资源唯一性与跨层一致性 · `AGENTS.md` 开发前必检规范

---

## 1. 背景

- 近期在评估 93 号与 96 号计划是否可归档时，发现多份职位管理相关文档间存在状态不一致与依赖尚未闭环的问题。
- 为避免重复沟通与跨层不一致，特建立 99 号文档，对 80–96 号区间内仍在推进的计划进行统一盘点，并给出建议的关闭顺序与前置条件。
- 88 号计划的 P2 后续事项已拆分为 101–104 号计划并分别完成，需将其纳入整体收口顺序与归档动作中。
- 本文档仅汇总现状与顺序建议，具体整改行动仍需在原计划内执行并回写权威事实来源。

## 2. 唯一事实来源

- `docs/development-plans/80-position-management-with-temporal-tracking.md`
- `docs/archive/development-plans/86-position-assignment-stage4-plan.md`
- `docs/development-plans/87-temporal-field-naming-consistency-decision.md`
- `docs/archive/development-plans/88-position-frontend-gap-analysis.md` 及 `docs/archive/development-plans/88-position-frontend-gap-analysis-review.md`
- `docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md`
- `docs/archive/development-plans/93-position-detail-tabbed-experience-acceptance.md`
- `docs/archive/development-plans/95-status-fields-review.md`
- `docs/development-plans/96-position-job-catalog-layout-alignment.md`
- `docs/archive/development-plans/101-position-playwright-hardening.md`
- `docs/archive/development-plans/102-positionform-data-layer-consolidation.md`
- `docs/archive/development-plans/103-position-components-tidy-up.md`
- `docs/archive/development-plans/104-ds147-positions-tabbed-experience.md`
- `docs/archive/development-plans/06-design-review-task-assessment.md`
- `docs/archive/development-plans/92-position-secondary-navigation-implementation.md`
- `docs/development-plans/107-position-closeout-gap-report.md`

## 3. 计划现状与阻塞因素

| 文档 | 当前状态摘要 | 主要阻塞 / 待回写 | 与其他计划的依赖 |
| --- | --- | --- | --- |
| **80号** 职位管理方案 | Stage 2/3 结论已记录，Stage 4 成果已回填 | 采纳 107 号差距报告中的性能/测试/设计补救措施，并回写时间线 | 依赖 86 归档、87 生产迁移确认 |
| **86号** 任职 Stage 4 计划（已归档） | ✅ 跨租户脚本、047/048 迁移、CI 验证全部完成 | 无（如有新增需求另立新计划） | 为 80 号方案更新提供最终结论 |
| **87号** 时态字段命名决策 | 开发侧迁移完成并归档；待与 86 号上线窗口联动执行生产迁移 | 按档案 §11（迁移流程）+ §12（与 86 号联动）完成上线验证并回写 06 号日志 | 为 80、86、88 提供命名一致性约束 |
| **88号** 前端差距分析（已归档） | ✅ 全量差距闭环，2025-10-21 移至 `docs/archive` | 无 | 提供历史差距与交付记录参照 |
| **93号** 方案（已归档） | ✅ 归档至 `docs/archive`，引用已更新 | 无 | 为 88、101–104 提供设计基线 |
| **93号** 验收报告（已归档） | ✅ 归档并补齐回写 | 无 | 与 88、101–104 的验收引用保持同步 |
| **95号** 状态字段评审（已归档） | ✅ 结论“暂不扩展五态” 已生效 | 无 | 供 88、101–104 引用，无额外依赖 |
| **96号** Job Catalog 布局校准 | 文档仍写“必须改造”，需验证现状 | 按第 3.2 节执行代码/截图验证，决定归档或新建实施计划 | 依赖 92 号归档证据、06 号日志记录 |
| **101号** Playwright Hardening（已归档） | ✅ 2025-10-20 完成，2025-10-21 移至 `docs/archive` | 无 | 为 88 第 12.3 节与 QA checklist 提供证据 |
| **102号** PositionForm 数据层整合（已归档） | ✅ 2025-10-20 完成，2025-10-21 移至 `docs/archive` | 无 | 为 88 建议 2 的交付证明 |
| **103号** 组件结构整理（已归档） | ✅ 2025-10-20 完成，2025-10-21 移至 `docs/archive` | 无 | 为 88 建议 3 的交付证明 |
| **104号** DS-147 设计规范（已归档） | ✅ 2025-10-20 发布，2025-10-21 移至 `docs/archive` | 无 | 为 88 设计一致性与后续评审提供依据 |

### 3.1 93号方案状态更新

2025-10-21 已完成以下动作：

- 93号方案文档迁移至 `docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md`，状态更新为「已完成」，并保留事后补充说明；
- 93号验收报告已归档至 `docs/archive/development-plans/93-position-detail-tabbed-experience-acceptance.md`，三项回写任务（实现清单、88号引用、06号日志）已完成；
- 88号、95号、06号文档以及实现清单的引用路径均已同步为归档位置。

**结论**：93号相关冲突已消除，无需进一步协调。

### 3.2 96号改造状态验证步骤

**问题**：96号文档描述"必须改造"，但92号已归档，需确认改造是否真的完成。

**验证清单**：
1. **代码验证**：
   ```bash
   # 检查 CardContainer 使用情况
   grep -r "CardContainer" frontend/src/features/job-catalog/

   # 预期：应在 family-groups/families/roles/levels 的 List 和 Detail 组件中使用
   ```

2. **文档验证**：
   - 检查92号归档时的"Phase 4 F4 - Job Catalog 页面"验收项是否勾选
   - 查看 `docs/archive/development-plans/92-position-secondary-navigation-implementation.md` 第816-822行
   - 确认06号日志中P1任务"设计评审"是否标记完成

3. **视觉验证**：
   - 查看布局截图是否存在：`frontend/artifacts/layout/job-family-groups-list.png`
   - 对比职位列表与 Job Catalog 列表的布局一致性

**处理决策**：
- **如果验证全部通过**：更新96号文档，将"必须改造"改为"已完成改造"，补充证据链接，申请归档
- **如果验证未通过**：保持96号"必须改造"描述，创建独立执行计划，96号暂不归档

### 3.3 88号拆分计划执行情况

- **101号计划**：`frontend/tests/e2e/README.md` 已记录真实/Mock 双模式执行步骤，`position-crud-live.spec.ts` 与 `position-tabs.spec.ts` 增补守护断言，状态标记为“已完成（2025-10-20）”。
- **102号计划**：`usePositionCatalogOptions` 抽离至共享 Hook，payload/validation/Storybook 全量校准，状态为“已完成（2025-10-20）”。
- **103号计划**：`frontend/src/features/positions/components` 完成目录分层与聚合导出，相关测试通过，状态为“已完成（2025-10-20）”。
- **104号计划**：`docs/reference/positions-tabbed-experience-guide.md` 发布 v0.1 指南，`frontend/artifacts/layout/README.md` 对截图命名与存放进行约定，状态为“已完成（2025-10-20）”。

> **结论**：101–104 号计划已于 2025-10-21 归档，88 号文档第 12 节随即更新完成。

## 4. 归档标准统一定义

为避免歧义，明确各类文档的归档条件：

### 4.1 计划文档归档条件（必须全部满足）
1. ✅ 主要功能/目标已完成（或明确放弃并说明原因）
2. ✅ 验收报告已发布（如适用），或文档自身包含验收结论
3. ✅ 所有待办事项已完成或转移至其他计划
4. ✅ 关联文档的引用已更新为归档路径
5. ✅ 在06号日志中记录归档事件
6. ✅ 文档移动至 `docs/archive/development-plans/`

### 4.2 验收报告归档条件
1. ✅ 验收结论已明确（通过/不通过/有条件通过）
2. ✅ 对应的方案文档已归档或状态已同步
3. ✅ 所有验收项的证据已保存（测试记录、截图等）
4. ✅ 未完成项已明确转移至其他计划或标记为技术债务

### 4.3 决策文档归档条件
1. ✅ 决策已定稿并获得相关方确认
2. ✅ 决策结果已同步至受影响的计划文档
3. ✅ 实施计划已创建或关联至现有计划
4. ✅ 如有代码变更，已创建对应的 PR 或任务链接

---

## 5. 推荐关闭顺序

> 93 号方案/验收与 95 号决策已按第 3.1 节说明完成归档，此处仅列仍需动作的计划。

### 5.1 阶段一：命名与任职收口

1. **87号**（时态字段命名决策）
   - 行动：依据档案 §11（迁移流程）完成预生产演练、上线窗口执行与回滚验证，并按 §12 与 86 号上线计划同步发布/记录。
   - 产出：预生产演练日志、上线执行记录、06 号日志回写、联合发布说明。

2. **86号**（职位任职 Stage 4 计划）
   - 行动：补齐跨租户脚本与监控文档、收集 `reports/position-stage4/` 基线，更新计划清单后归档。
   - 产出：脚本执行截图、性能基线、归档提交及 06 号日志时间戳。

3. **80号**（职位管理总体方案）
   - 行动：在 86 号归档后回填 Stage 4 实际进展与 acting 自动化成果，更新时间线并确认是否进入下一阶段或归档。
   - 产出：Stage 4 进展段落、引用 87/101–104 的链接、若完成则归档提交。

### 5.2 阶段二：前端差距与交付归档

4. **101–104号**（拆分计划归档）
   - 行动：✅ 2025-10-21 完成归档并回写 88 号与 06 号日志，无需额外动作。
   - 产出：归档提交与日志条目已就位。

5. **88号**（前端差距分析） — ✅ 2025-10-21 归档完成，无需额外动作。

### 5.3 阶段三：布局一致性收尾

6. **96号**（Job Catalog 布局校准）
   - 行动：依据第 3.2 节清单完成代码/截图验证，根据结果更新文档为“已完成”并归档，或创建新的实施计划编号。
   - 产出：验证记录、截图链接或新计划 ID、文档状态更新。

## 6. 下一步与责任分工

| 顺序 | 行动 | 责任角色 | 需要的证据/输出 |
| --- | --- | --- | --- |
| 1 | 87 号执行生产迁移并归档 | 架构组 · 数据库团队 | 迁移执行/回滚记录、06 号日志条目、文档归档 PR |
| 2 | 86 号归档 | 命令服务团队 · 查询服务团队 · QA 团队 | ✅ 跨租户脚本、reports/position-stage4/ 基线、归档提交 |
| 3 | 80 号方案更新 | 业务架构组 | Stage 4 进展段落、引用 87/101–104 链接、归档或后续计划说明 |
| 4 | 96 号验证并决策 | 前端团队 · 设计团队 | 验证清单结果、截图或新计划编号、文档状态更新 |

## 7. 审阅与更新机制

- 本文档定位为协调性指导，建议在每次计划状态发生变化时同步更新。
- 当上述 6 个节点全部完成后，可将本文件归档至 `docs/archive/development-plans/99-plan-closure-alignment.md`，并在 06 号日志中记录收口结果。
- 若过程中出现新的阻塞或计划新增，请先更新原计划，再回写至本指导文档，以确保唯一事实来源不漂移。

---

> **备注**：本文内容全部来源于列出的唯一事实来源文档及最新代码库状态，未引入第二事实来源；如未来出现差异，请以原计划文档和代码实现为准，并同步修订此文档。
