# 95 号文档：Status Fields Review 调查报告

**创建日期**：2025-10-21  
**状态**：已完成（待归档）  
**维护人**：架构组 · 前端治理小组

---

## 1. 背景与目标

- **问题起因**：在 "Status Fields Review" 讨论中，团队认为 `TimelineComponent` 依赖组织模块的 `lifecycleStatus`、`businessStatus`、`dataStatus` 三个字段（分别对应生命周期五态、业务启停与数据状态），但实际运行表现与描述不符，需要确认结论并定位误解来源。
- **调查目标**：核对前端实现与 GraphQL 契约，澄清三个字段的真实来源与用途，说明误解如何产生，并提出修正建议。
- **范围界定**：仅聚焦前端时间轴组件及其数据转换；后端命令/查询实现未纳入本次调查。

---

## 2. 权威事实来源

- `frontend/src/features/temporal/components/TimelineComponent.tsx`：时间轴渲染与状态映射逻辑。
- `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`：组织版本 GraphQL 查询与 `TimelineVersion` 构造函数。
- `frontend/src/features/temporal/components/inlineNewVersionForm/formActions.ts`：客户端插入/编辑历史记录时构造的 `TimelineVersion`。
- `docs/api/schema.graphql`：`organizationVersions` 查询返回字段定义（`status`、`isCurrent` 等）。
- `docs/development-plans/93-position-detail-tabbed-experience-plan.md` 第 94 行：对 `lifecycleStatus`/`businessStatus`/`dataStatus` 的复用要求（误解来源关键记录）。

上述文件互为佐证，未引入第二事实来源；若未来实现有变动，应首先更新这些权威文件并同步调整本报告。

---

## 3. 调查结论

| 字段 | 实际实现 | 与假设不符点 |
| --- | --- | --- |
| `lifecycleStatus` | 在 `temporalMasterDetailApi.ts` / `formActions.ts` 中仅依据 `isCurrent` 或 `status === 'ACTIVE'` 映射为 `CURRENT` 或 `HISTORICAL`。未出现 `PLANNED`、`INACTIVE`、`DELETED`。 | 不存在所谓的“五态生命周期”驱动色彩；时间轴标题注释虽写有五态，但代码从未提供额外取值。 |
| `businessStatus` | 由 `status === 'ACTIVE'` 映射为 `'ACTIVE'`，否则 `'INACTIVE'`。时间轴组件仅在 badge 前置逻辑中用作“已停用”提示。 | `TimelineComponent` 中 `StatusBadge` 仍使用 `version.status`，并未与 `businessStatus` 联动显示“生效中/已停用”标签。 |
| `dataStatus` | 所有构造函数均写死 `'NORMAL'`，未有任何路径赋值 `'DELETED'`。 | 时间轴“软删除特殊样式”逻辑从未触发，组件注释与真实数据不一致。 |

---

## 4. 误解来源分析

1. **生命周期五态假设**  
   - `TimelineComponent.tsx` 顶部注释写明“基于五状态生命周期管理字段”，且 `TimelineVersion` 接口允许 `'PLANNED' | 'INACTIVE' | 'DELETED'`。  
   - `docs/development-plans/93-position-detail-tabbed-experience-plan.md` 明确要求“复用组织模块的 `lifecycleStatus` 字段”，团队在移植职位时间轴时沿用了该设想，却未重新校验 `temporalMasterDetailApi.ts` 的实际映射逻辑（仅返回 `CURRENT/HISTORICAL`）。

2. **业务状态复用误差**  
   - 组织模块早期在设计评审中拟定“两层状态”区分：业务启停（Active/Inactive）+ 生命周期。相关描述保留在 Plan 93 与组件注释中。  
   - 当前实现中 `businessStatus` 只是 `status` 的派生值，`StatusBadge` 继续读取 `timeLineVersion.status`；缺乏额外展示导致团队误以为两套状态均生效。

3. **数据状态（软删除）预期落空**  
   - Plan 93 中为了兼容未来的“软删除版本”场景，要求补齐 `dataStatus`。然而 `organizationVersions` 查询并无该字段，API 调用也未传递 `includeDeleted`，客户端便在构造对象时写死 `'NORMAL'`。  
   - 由于缺少对应 GraphQL 字段且未执行端到端验证，“软删除触发特殊样式”的设想一直处于未实现状态。

总结：误解源于复用早期设计和计划文档的设想而未再验证唯一事实来源（GraphQL 契约与实现代码），导致口径长期停留在规划阶段的假定值。

---

## 5. 风险与影响

- **视觉与状态提示偏差**：团队以为时间轴颜色和标签已经覆盖五态/软删除，实则仅区分当前/历史与启停，可能误导后续 UX 评审或准生产验收。
- **跨模块复用风险**：职位模块若继续复用 `TimelineComponent`，会沿用这些默认值，造成“文档描述 ≠ 实际效果”的跨层不一致。
- **测试盲区**：现有 Vitest/E2E 未覆盖 `PLANNED` 或软删除场景，容易放大认知偏差。

---

## 6. 建议与下一步

1. **修正文档口径**：在 Plan 93 等设计文档中明确注明当前实现仅提供 `CURRENT/HISTORICAL` 与 `ACTIVE/INACTIVE`，`dataStatus` 为占位字段，杜绝对“五态”或软删除展示的描述。  
2. **组件内兜底逻辑**：保持 `TimelineComponent` 仅依赖既有字段，对未提供的状态分支给予显式降级或移除，避免传播不存在的语义。  
3. **测试覆盖**：新增用例校验现有字段取值，确保未来若继续限制为 `CURRENT/HISTORICAL` 仍能通过回归；软删除/计划态无需构造模拟数据。

---

## 7. 归档与同步

- 待 Plan 93 更新完毕并决定是否扩展 GraphQL 字段后，再将本报告移动至 `docs/archive/development-plans/95-status-fields-review.md`。  
- 若后续确有变更需求，应以新的契约文件为唯一事实来源，并同步更新本报告；在正式决定前不得提前传播假设字段。
