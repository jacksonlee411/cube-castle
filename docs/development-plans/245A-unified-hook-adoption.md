# Plan 245A – Unified Hook Adoption（组织详情子组件渐进采纳）

关联计划：Plan 245（类型 & 契约统一，已完成）；Plan 242（T3 命名抽象）

状态：进行中（非破坏性、小步推进）

—

1. 目标与范围
- 目标：在不改变后端契约与现有行为的前提下，逐步将组织详情页面的子组件切换为读取统一 Hook（useTemporalEntityDetail）提供的统一 record 字段（displayName/status/effectiveDate/endDate 等），减少零散数据源依赖，巩固“单一事实来源”。
- 范围（组织端）：
  - TemporalMasterDetailHeader：标题/状态等优先读取 record 字段（已初步接入，状态兜底已实施）
  - TemporalMasterDetailAlerts：提示文案（如当前状态）与可重试逻辑中涉及展示的数据读取统一化
  - TemporalEditForm（InlineNewVersionForm 同步）：表单初始化与提交时的默认值/展示字段统一化
  - ParentOrganizationSelector：asOfDate 的统一接入与上下游一致性（仅展示/校验，保持调用方契约不变）
  - 其他与组织详情关联的展示组件（若存在），按需增量纳入
- 非目标（本计划不做）：
  - 不重命名 GraphQL/REST 字段；不引入任何破坏性接口调整
  - 不重写版本加载/状态切换/删除等成熟链路（沿用现有 API，统一 Hook 仅兜底与统一展示字段）

—

2. 原则与约束
- 单一事实来源：组织详情展示优先读取 `useTemporalEntityDetail('organization')` 返回的 `record` 字段；与时间线/版本加载数据不一致时，以现有版本/时间线数据为准，统一 Hook 作兜底
- 非破坏性：保留原有字段读取作为回退逻辑，确保 UI 行为稳定；任何变更配套最小化 diff 与严格回归
- 日志留痕：每个子任务提交前执行 codegen/Typecheck/Vitest/Plan245 守卫，输出保存到 `logs/plan245A/`（同 245 的留痕规范）

—

3. 交付物
- 代码变更：按子组件拆分的最小提交（每次只触达一个组件/一个行为点）
- 日志与证据：
  - GraphQL codegen：`logs/plan245A/xx-frontend-codegen.log`
  - TypeScript 类型检查：`logs/plan245A/xx-frontend-typecheck.log`
  - Vitest：`logs/plan245A/xx-frontend-vitest.log`
  - Plan 245 守卫：`logs/plan245A/xx-plan245-guard.log`
- 文档：在 `reports/plan242/naming-inventory.md` 中追加“采用进度”小节；在本计划文件维护任务清单与状态

—

4. 子任务清单（滚动维护）
- T1 Header：标题/状态从 `record.displayName/record.status` 读取（已完成状态兜底；本任务完成后可移除冗余回退）
- T2 Alerts：涉及状态/名称的提示文本统一读取 record 字段（保留原回退）
- T3 EditForm：初始化默认值与生效日期展示读取 record 字段（不更改提交契约；仅改善展示/默认）
- T4 ParentSelector：统一 asOfDate 的来源与传递（读取 record.effectiveDate 作为默认，保持回退逻辑）
- T5 其他展示组件（如有）：逐步切换

每个子任务完成标准：
1) UI 行为与现有一致（或更优）；
2) codegen/Typecheck/Vitest/守卫全部通过；
3) `reports/plan242/naming-inventory.md` 记录“采纳项 + 依据”；
4) 若出现不一致，以时间线/版本加载数据为准，统一 Hook 仅作展示兜底；
5) 变更范围仅限一个子组件/行为点。

—

5. 与 Plan 245 的关系
- Plan 245 已完成“类型 & 契约统一、统一 Hook 引入、operation 命名统一、守卫接入及契约注释”；本计划作为“统一方案的深入采纳”继续推进 UI 层统一化，不阻塞 Plan 245 已关闭状态。

—

6. 风险与回滚
- 风险：组件切换期间出现展示不一致或 UI 回归
- 缓解：保留旧路径回退、最小粒度提交、严格回归测试
- 回滚：若出现不可接受的行为差异，立即回滚到上个 tag，并保留日志以复现与定位

—

7. 附录：已完成与参考
- 已完成（Plan 245）：
  - 统一类型与 Hook：`frontend/src/shared/types/temporal-entity.ts`、`frontend/src/shared/hooks/useTemporalEntityDetail.ts`
  - 职位详情页使用统一 Hook；组织主从视图以统一 Hook 兜底名称/状态
  - operation 命名统一为 `TemporalEntity*`（详情/版本/路径/审计；树查询部分保留测试敏感项）
  - Plan 245 守卫接入 CI（agents-compliance / frontend-quality-gate）
- 参考：
  - `docs/development-plans/245-temporal-entity-type-contract-plan.md`（完成说明与证据）
  - `docs/api/schema.graphql` / `docs/api/openapi.yaml`（Plan 245 注释）

