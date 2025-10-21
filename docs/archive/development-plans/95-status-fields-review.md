# 95 号文档：Status Fields Review 调查报告

**创建日期**：2025-10-21  
**最新更新**：2025-10-21 18:45  
**状态**：✅ 已完成（复检结论同步归档）  
**维护人**：架构组 · 前端治理小组

---

## 1. 背景与目标

- **问题起因**：在“Status Fields Review”讨论中，团队误以为时间轴组件已经复用 `lifecycleStatus`、`businessStatus`、`dataStatus` 三套状态；实际运行时仅呈现当前/历史与启停，需要澄清事实来源。
- **调查目标**：核对前端实现与 GraphQL 契约，明确三个字段的真实来源、用途及误解成因，并提出修正建议。
- **范围界定**：聚焦查询服务返回的数据与前端映射逻辑；命令服务改造不在本次调查范围内。

## 2. 权威事实来源

- `frontend/src/features/temporal/components/TimelineComponent.tsx`
- `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`
- `frontend/src/features/temporal/components/inlineNewVersionForm/formActions.ts`
- `docs/api/schema.graphql`（`organizationVersions` 查询定义）
- `docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md`

上述文件为唯一事实来源；若未来实现变动，应先更新这些文件并回写本报告。

## 3. 主要结论

| 字段 | 实际实现 | 结论 |
| --- | --- | --- |
| `lifecycleStatus` | 仅根据 `isCurrent`/`status === 'ACTIVE'` 映射为 `CURRENT/HISTORICAL` | 未落地“五态”设想 |
| `businessStatus` | `status === 'ACTIVE'` → `'ACTIVE'`，否则 `'INACTIVE'` | 仅起到展示启停的作用 |
| `dataStatus` | 客户端写死 `'NORMAL'` | 软删除展示尚未实现 |

## 4. 误解来源

1. 早期设计文档保留“五态生命周期”设想，未再次核对实现。
2. `TimelineComponent` 注释强调两层状态，但实际仍仅使用 `version.status`。
3. `dataStatus` 为未来软删除预留字段，GraphQL 当前未返回。

## 5. 风险评估

- 文档口径与实现不符，易在评审或上线前产生误解。
- 职位模块复用时间轴组件时会沿用当前简化逻辑，需明确限制。
- 测试缺乏对 `PLANNED/DELETED` 的覆盖，未来扩展需同步补充。

## 6. 处理动作（已执行）

1. **修正文档口径**：归档版 Plan 93、107 号报告 v2.0 已更新描述，只保留 `CURRENT/HISTORICAL` 与 `ACTIVE/INACTIVE`。
2. **组件降级策略**：保留 `TimelineComponent` 既有降级逻辑，待契约更新后再开放新状态样式。
3. **测试计划**：维持现有回归用例，如需覆盖软删除或计划态，将另立新计划。

## 7. 复检记录（2025-10-19）

- 复核确认 GraphQL 仍仅返回 `ACTIVE/INACTIVE`，软删除版本默认过滤；前端组件新增的 `resolveLifecycleStatus`、`isSoftDeleted` 逻辑目前仍只能由调用方派生触发。
- 该结论与初版一致，误解根源在于文档与代码未同步更新。

## 8. 归档说明

- 2025-10-21：确认所有修正动作完成，95 号调查报告与相关计划状态一致。
- 本文档将与 88、99、107 等关联文档一并迁移至 `docs/archive/development-plans/`。
- 后续若扩展 GraphQL 字段或软删除展示能力，请以新的契约和计划另行更新。
