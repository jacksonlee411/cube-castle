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

5. 验收标准与关闭条件
- 组件迁移完成度：
  - 列表中的核心子组件（Header/Alerts/EditForm/ParentSelector）均已完成采纳，并在代码中优先读取统一 Hook 的 `record` 字段；
  - 禁止在上述子组件内新增对旧路径（零散 API 字段）的直接读取（例外需在 allowlist 记录并注明过期时间），最终收敛为“统一 Hook + 时间线/版本数据”双来源。
- 一致性优先级断言（单测/E2E 覆盖）：
  - 展示字段读取优先级固定为：时间线/版本数据 > 统一 Hook `record` > 旧路径回退（仅过渡期）；过渡结束后不再读取旧路径。
- 测试与门禁：
  - GraphQL codegen、Typecheck、Vitest、Plan 245 守卫每次提交前均通过，并在 `logs/plan245A/` 留痕；
  - 新增 1–2 条 E2E 场景（见第 8 节），在 CI 上稳定通过。
- 日志与文档：
  - `reports/plan242/naming-inventory.md` 中增加“采纳进度”条目，记录组件迁移与证据链接；
  - `logs/plan245A/**` 收集并在 CI 中作为 artifact 归档（待工作流接入）。
- 关闭条件：
  - 上述核心子组件全部采纳完成；E2E 场景稳定通过；Plan 245 守卫 0 新增旧命名；无例外 allowlist 或 allowlist 清零；评审完成并在本计划文件标记“已完成”。

—

6. asOfDate 默认优先级（组织详情）
- 统一默认值与传递优先级：
  1) 所选版本（selectedVersion.effectiveDate）
  2) 当前版本有效日期（isCurrent 版本的 effectiveDate）
  3) 统一 Hook `record.effectiveDate`（兜底）
- 说明：asOfDate 代表“查看视角”的时间点，不应直接使用任意版本的生效日期；需与时间线/版本选择保持一致。

—

7. 缓存一致性与失效策略
- 在以下时机失效并刷新缓存：
  - 版本/状态操作成功后：invalidate 统一 Hook 的详情 queryKey + 版本/timeline 的 queryKey，随后主动 `refetch`；
  - 切换时间线版本后：刷新读取路径，使 Header/Alerts/EditForm 等展示字段与选中版本一致；
  - 组织/上级变更后：使 ParentSelector 的候选/校验所依赖的查询与详情保持一致。
- 要求：失效与刷新封装在对应的 handler 中（例如状态切换按钮、表单提交、版本选择），并有 Vitest 断言。

—

8. E2E 场景要求
- 场景 A：切换不同版本后，Header/Alerts/表单展示字段（名称/状态/日期）与时间线选中版本一致。
- 场景 B：状态切换后，刷新后展示字段一致；统一 Hook 与版本数据无漂移。
- 输出：将 Playwright 结果上传到 CI artifact，并在 `logs/plan245A/` 留痕。

—

9. CI/守卫与 Artifact（实施指引）
- Plan 245 守卫已在 CI 接入，用于冻结 `query PositionDetail/PositionDetailQuery` 新增；
- 建议新增“旧字段直读”软守卫（仅告警，产出列表与 allowlist），引导逐步收敛；
- 建议在前端质量工作流中上传 `logs/plan245A/**` 作为 artifact（实施在 CI 层进行，本计划仅给出要求）。

—

10. 与 Plan 245 的关系
- Plan 245 已完成“类型 & 契约统一、统一 Hook 引入、operation 命名统一、守卫接入及契约注释”；本计划作为“统一方案的深入采纳”继续推进 UI 层统一化，不阻塞 Plan 245 已关闭状态。

—

11. 风险与回滚
- 风险：组件切换期间出现展示不一致或 UI 回归
- 缓解：保留旧路径回退、最小粒度提交、严格回归测试
- 回滚：若出现不可接受的行为差异，立即回滚到上个 tag，并保留日志以复现与定位

—

12. Owner / 时间线（占位）
- Owner：TBD（前端负责人）
- 目标节奏：每周完成 ≥1 个子组件的迁移，首周完成 T2（Alerts），次周完成 T3（EditForm）
- 每次提交附：codegen / Typecheck / Vitest / Plan 245 Guard 日志，路径 `logs/plan245A/`

—

13. 附录：已完成与参考
- 已完成（Plan 245）：
  - 统一类型与 Hook：`frontend/src/shared/types/temporal-entity.ts`、`frontend/src/shared/hooks/useTemporalEntityDetail.ts`
  - 职位详情页使用统一 Hook；组织主从视图以统一 Hook 兜底名称/状态
  - operation 命名统一为 `TemporalEntity*`（详情/版本/路径/审计；树查询部分保留测试敏感项）
  - Plan 245 守卫接入 CI（agents-compliance / frontend-quality-gate）
- 参考：
  - `docs/development-plans/245-temporal-entity-type-contract-plan.md`（完成说明与证据）
  - `docs/api/schema.graphql` / `docs/api/openapi.yaml`（Plan 245 注释）
