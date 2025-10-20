# 99号文档：职位管理相关计划收口顺序建议

**版本**: v0.2
**创建日期**: 2025-10-21
**最新更新**: 2025-10-20
**状态**: 待执行
**编制人**: Claude Code
**关联计划**: 80、86、87、88、93、95、96 号文档
**遵循原则**: `CLAUDE.md` 资源唯一性与跨层一致性 · `AGENTS.md` 开发前必检规范

---

## 1. 背景

- 近期在评估 93 号与 96 号计划是否可归档时，发现多份职位管理相关文档间存在状态不一致与依赖尚未闭环的问题。
- 为避免重复沟通与跨层不一致，特建立 99 号文档，对 80–96 号区间内仍在推进的计划进行统一盘点，并给出建议的关闭顺序与前置条件。
- 本文档仅汇总现状与顺序建议，具体整改行动仍需在原计划内执行并回写权威事实来源。

## 2. 唯一事实来源

- `docs/development-plans/80-position-management-with-temporal-tracking.md`
- `docs/development-plans/86-position-assignment-stage4-plan.md`
- `docs/development-plans/87-temporal-field-naming-consistency-decision.md`
- `docs/development-plans/88-position-frontend-gap-analysis.md` 及 `docs/archive/development-plans/88-position-frontend-gap-analysis-review.md`
- `docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md`
- `docs/archive/development-plans/93-position-detail-tabbed-experience-acceptance.md`
- `docs/archive/development-plans/95-status-fields-review.md`
- `docs/development-plans/96-position-job-catalog-layout-alignment.md`
- `docs/archive/development-plans/06-design-review-task-assessment.md`
- `docs/archive/development-plans/92-position-secondary-navigation-implementation.md`

## 3. 计划现状与阻塞因素

| 文档 | 当前状态摘要 | 主要阻塞 / 待回写 | 与其他计划的依赖 |
| --- | --- | --- | --- |
| **80号** 职位管理方案 | 标记「Stage 2 已完成，Stage 3 已批准」；未同步 Stage 4 后续 | 需等待 Stage 4（86号）执行成果与字段命名决策（87号）后更新全局方案 | 依赖 86、87 |
| **86号** 任职 Stage4 计划 | 状态「待复审」，GraphQL `currentAssignment/assignmentHistory` 缺失 | 先补齐查询服务 resolver 与仓储实现；待 87 号对字段命名给出结论 | 依赖 87、与 80 同步 |
| **87号** 时态字段命名决策 | ✅ 已完成（2025-10-21） | CLI 迁移 + 代码/契约/前端同步统一为 `effective_date` | 已解除 80、86 的前置约束 |
| **88号** 前端差距分析 | P0/P1 完成，P2 进入多页签规划 | 需等待 93 号方案定稿与 95 号结论回写 | 依赖 93、95 |
| **93号** 方案文档 | 已归档为 v0.2（状态「已完成」） | -- | 与 88、95、06 号引用保持同步 |
| **93号** 验收报告 | 标记 ✅ 通过，但部分回写（实现清单、06号日志）未完成 | 确认所有回写闭环后方可归档 | 需要与 93 号方案、88 号同步 |
| **95号** Status Fields Review | 已归档（2025-10-20），架构决策确认不扩展五态 | 已定稿并更新 93 号文档，技术债务转P3清单 | 与 93 方案一致，不阻塞 88 号 |
| **96号** Job Catalog 布局校准 | 文档中仍写「必须改造」，与现有实现不符 | **验证步骤见 3.2 节**，需确认改造已完成 | 依赖 92号归档、06号日志 |

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

### 5.1 第一批次：架构决策链

1. **87号**（时态字段命名决策）— ✅ 已完成（开发环境）
   - 结果：047 迁移 + 全栈改造合并，OpenAPI/GraphQL/前端已统一为 `effective_date`。
   - 后续：待生产迁移执行后，将文档归档至 `docs/archive/` 并在 06 号日志记录上线验证。

2. **86号**（职位任职 Stage 4 计划）
   - 前提：87 号已完成；查询服务 resolver 落地。
   - 执行动作：解除 GraphQL 阻塞、更新计划状态、补充测试与日志回写。
   - 完成后更新 80 号方案 Stage 4 里程碑。

3. **80号**（职位管理总体方案）
   - 前提：86 号 Stage 4 已启动或完成，87 号决策已落实。
   - 动作：回填 Stage 3/4 实际进展，说明后续迭代计划或归档。

### 5.2 第二批次：前端体验链

4. **95号**（Status Fields Review）
   - 目标：定稿时间轴字段结论，明确当前仅有 `ACTIVE/INACTIVE` + `CURRENT/HISTORICAL`。
   - 完成标志：推送修订至 93 号方案/验收及相关文档，必要时创建后续计划处理软删除或五态扩展。
   - 动作：✅ **已完成**（2025-10-20）——架构决策确认不扩展五态，避免过度设计；技术债务转P3清单；文档已归档。

5. **93号方案**（职位详情多页签体验）
   - 前提：95 号结论已经合入。
   - 动作：✅ 已完成——方案文档已更新为 v0.2 并归档，引用路径同步完成。
   - 备注：归档版本保留与 94 号评审的差异说明，供后续迭代参考。

6. **93号验收报告**
   - 前提：方案稿已更新且所有回写完成。
   - 动作：✅ 已完成——文档已归档，相关引用与回写均确认闭环。

7. **88号**（前端差距分析）
   - 前提：93 号计划与验收均已闭环，95 号决策已明确。
   - 动作：将 P2 状态更新为「完成」，或者将剩余事项拆分到新的计划；同步引用 93、95 的最终结论。

### 5.3 第三批次：布局一致性链

8. **96号**（Job Catalog 布局校准）
   - 前提：确认 92 号归档与 06 号日志已表明改造完成。
   - 动作：将文档中的「必须改造」陈述更新为完成情况，补充截图/测试证据，并归档。

## 6. 下一步与责任分工

| 顺序 | 行动 | 责任角色 | 需要的证据/输出 |
| --- | --- | --- | --- |
| 1 | 87 号决策讨论并定稿 | 架构组 & 数据库团队 | 决策纪要、迁移计划、代码修订任务链接 |
| 2 | 86 号解除 GraphQL 阻塞 | 查询服务团队 | Resolver/仓储变更 PR、`go test ./cmd/organization-query-service/...` 记录 |
| 3 | 更新 80 号方案 | 业务架构组 | Stage 3/4 进展回写、与 86 号状态一致的时间线 |
| 4 | 95 号定稿并同步到 93/88 | 前端治理小组 | ✅ 已完成（架构决策确认不扩展五态，文档已归档） |
| 5 | 93 号方案与验收闭环 | 职位体验小组 | ✅ 已完成（实现清单与文档引用已同步） |
| 6 | 88 号调整状态 | 前端团队 | 文档状态更新或拆分后续计划 |
| 7 | 96 号状态更新与归档 | 前端团队 · 设计团队 | 截图佐证、92 号归档/06 号日志引用、文档状态切换 |

## 7. 审阅与更新机制

- 本文档定位为协调性指导，建议在每次计划状态发生变化时同步更新。
- 当上述 8 个节点全部完成后，可将本文件归档至 `docs/archive/development-plans/99-plan-closure-alignment.md`，并在 06 号日志中记录收口结果。
- 若过程中出现新的阻塞或计划新增，请先更新原计划，再回写至本指导文档，以确保唯一事实来源不漂移。

---

> **备注**：本文内容全部来源于列出的唯一事实来源文档及最新代码库状态，未引入第二事实来源；如未来出现差异，请以原计划文档和代码实现为准，并同步修订此文档。
