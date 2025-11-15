# Plan 240 – 职位管理页面重构与稳定性计划

**编号**: 240  
**标题**: Position Management UI Refactor & E2E Stabilization  
**创建日期**: 2025-11-09  
**关联计划**: Plan 232（Playwright P0 稳定）、Plan 06（集成验证）、80/86/107 号职位阶段性文档  
**状态**: 进行中（A–D/BT 已完成，E 回归与 Runbook 执行中）  

---

## 0. 完成登记与现状评估（2025-11-15）

为避免第二事实来源，子计划状态以各自计划文档与落盘证据为准，本节仅做索引与归档登记：

- 240A – 职位详情 Layout 对齐与骨架替换：已完成（验收通过）  
  - 事实来源：`docs/development-plans/240A-position-layout-alignment.md`  
  - 证据：`logs/plan240/A/architecture-validator.log`、`logs/plan240/A/document-sync.log`、`logs/plan240/A/selector-guard.log`、`logs/plan240/A/e2e-position-tabs-{chromium,firefox}.log`、`logs/plan240/A/playwright-trace/trace.zip`
- 240B – 数据装载链路与等待治理：已完成  
  - 事实来源：`docs/development-plans/240B-position-loading-governance.md`  
  - 证据：`logs/plan240/B/guard-plan245.log`、`logs/plan240/B/guard-selectors-246.log`、`logs/plan240/B/e2e-*.log`
- 240C – DOM/TestId 治理与选择器统一：已完成（验收通过）  
  - 事实来源：`docs/development-plans/240C-position-selectors-unification.md`  
  - 证据：`logs/plan240/C/selector-guard.log`、`logs/plan240/C/unit-and-e2e.log`（如有）
- 240D – 职位详情可观测性与指标注入：已完成（验收通过）  
  - 事实来源：`docs/development-plans/240D-position-observability.md`、`CHANGELOG.md:3`（v1.5.7）  
  - 证据：`logs/plan240/D/obs-position-observability-chromium.log`、`frontend/playwright-report/index.html`
- 240E – 回归与 Runbook：进行中（T0/T4 门禁已跑通，回归套件待 CI 完成）  
  - 事实来源：`docs/development-plans/240E-position-regression-and-runbook.md`  
 - 现有证据：`logs/plan240/E/architecture-validator.log`、`logs/plan240/E/selector-guard.log`、`logs/plan240/E/temporary-tags.log`、`logs/plan240/E/guard-plan245.log`
 - 240BT – 组织详情“白屏”治理（路由解耦）：已完成（验收通过）  
   - 事实来源：`docs/archive/development-plans/240bt-org-detail-blank-page-mitigation.md`  
   - 证据：`logs/plan240/BT/smoke-org-detail.log`、`logs/plan240/BT/health-checks.log`  

依赖状态同步：
- Plan 242 – 命名抽象：已完成，作为 240 的关键解锁条件（见 `docs/development-plans/244-temporal-timeline-status-plan.md` “完成登记”与 `CHANGELOG.md:26`）。  
- Plan 244 – Timeline/Status 抽象：已完成（验收通过）；240B/240D 已按新命名空间落地。  
- 契约与实现一致性：`docs/api/openapi.yaml`、`docs/api/schema.graphql` 已按 242/244 更新；240 不新增契约，遵循“先契约后实现”。  

> 本节为状态索引，具体标准与字段定义仅引用各子计划与 `docs/reference/temporal-entity-experience-guide.md`，不在此重复，避免第二事实来源。

## 0.1 影响评估：240 先于 241 完成的权衡与回补计划

背景：Plan 241（前端页面与框架一体化重构）当前为“部分完成 · 验收未通过”（见 `docs/development-plans/241-frontend-framework-refactor.md`）。本计划的 A–D 子任务已交付在既有骨架与 Hook 上，存在“在 241 交付前的临时对齐与薄适配”的客观状态。

- 影响评估（风险与现状）
  - 骨架重复与收敛负担：职位端依旧基于既有 `TemporalMasterDetailView` 外层与 `PositionDetailView` 自身布局进行对齐（240A），未采用 241 计划中的中性骨架；后续需二次替换为 241 的共享 Layout（不引入第二事实来源，引用 241 的组件/指南）。  
  - 数据 Hook 临时薄适配：240B 按“统一 Loader/取消/重试/失效 SSoT”落地，但 241 的统一 Hook/工厂作为最终承载框架仍在收尾；后续需将职位域薄适配层切换到 241 的唯一入口，回收临时适配。  
  - 可观测点分散：240D 在职位详情端注入 `[OBS] position.*` 事件；241 完成后需上移到共享骨架统一注入，去除页面内重复注入。  
  - 选择器稳定性：240C 已集中到 `temporalEntitySelectors`。241 完成后骨架变更不应改变选择器标识来源，但需复跑 E2E 以确认不引入结构性变化。  
  - 路由白屏：240BT 已完成路由解耦与错误边界，满足 240B“避免白屏”的前置治理；241 收敛后仍应保持该解耦与边界，不得回退。

- 回补计划（241 完成后，240 的待回归事项）
  1) 骨架切换（不改契约）  
     - 将职位详情外层容器替换为 241 交付的中性骨架组件（参考：`docs/development-plans/241-frontend-framework-refactor.md` 的收尾目标“TemporalEntityLayout”）。  
     - A11y 与视觉间距维持与组织端一致；任何差异在 MR 中登记原因与后续修正（不在本文件复制规范正文，唯一事实来源：`docs/reference/temporal-entity-experience-guide.md`）。
  2) Hook/Loader 统一（薄适配回收）  
     - 将职位域数据读取入口切换为 241 的统一 Hook/工厂（如 `useTemporalEntityDetail`/`createTemporalDetailLoader`），保留 240B 的取消/重试/失效策略，通过 241 的统一配置导出（SSoT）。  
     - 移除职位专有 Loader/Hook 的临时分支；若存在 `// TODO-TEMPORARY:` 标记，需在合并时清零并通过 `scripts/check-temporary-tags.sh` 校验。
  3) 可观测性归一  
     - 将 hydrate/tab/version/export/graphql.error 的注入位置上移至共享骨架与统一错误处理（241 指南）；职位页面保留调用点必要事件（如导出回调）。  
     - 保持事件词汇与字段以 `docs/reference/temporal-entity-experience-guide.md` 为唯一事实来源；不在 240 内定义新事件。
  4) 选择器/E2E 回归  
     - 骨架切换后复跑职位域 P0 集合与集成用例（参考 Plan 06 与 232/232T），落盘日志至 `logs/plan240/E/` 并更新 Plan 06 执行登记。  
     - `npm run guard:selectors-246` 通过（旧前缀计数不升高）；`node scripts/quality/architecture-validator.js` 通过。  
  5) Feature Flag 收敛  
     - 评估并回收 240A 的 `VITE_POSITION_LAYOUT_V2` 与 240B 的 `VITE_TEMPORAL_DETAIL_LOADER`；如 241 提供统一门控，以 241 的门控为准并移除同功能重复开关。  
  6) 清理与归档  
     - 移除与旧骨架/旧 Loader 强绑定的临时代码与注释；更新本计划“完成登记”索引与 215/06 的状态表；不新增第二事实来源文档。

- 回归验收（241 完成后触发）  
  - UI/行为：组织与职位详情在共享骨架下的录屏/截图比对一致（可引用 241 的对比资产，240 不复制）。  
  - Hook/Loader：`useTemporalEntityDetail` 薄封装接入，相关单测由 241 提供；职位端不再直连自有 Loader。  
  - 可观测性：事件集中由骨架发射；职位端不重复注入 hydrate/tab 事件。  
  - E2E：职位域 P0 + 集成用例（Chromium/Firefox 各至少 1 轮）通过；`logs/plan240/E/*.log` 记录结果。  
  - 门禁：`guard:selectors-246`、`architecture-validator`、`check-temporary-tags.sh` 全部通过。  

## 1. 背景与动因

1. `position-tabs`、`position-lifecycle` 场景在 232T 复测中持续失败，暴露 GraphQL 等待链路、Tab 渲染与 DOM 标识缺口，Firefox 甚至尚未补跑（docs/development-plans/232t-test-checklist.md:25-69）。  
2. Plan 06 最新纪要同样记录职位模块 P0 用例在 `任职历史`、`position-detail-card` 等关键元素上无法渲染，说明问题已影响集成回归（docs/development-plans/06-integrated-teams-progress-log.md:11-19）。  
3. 80/86/107 号计划虽宣告职位管理功能上线，但当前 UI 结构沿袭 Stage 2 的临时实现，缺乏统一的状态机、GraphQL 数据缓存与指标监控，导致一旦数据延迟或租户切换即出现“白屏/元素缺失”。  

> **结论**：231/232 线索已经证明职位管理页面存在结构性问题，必须通过专项重构降低 DOM/数据耦合、补齐可观测性，并重新建立 P0 级别的 E2E 证据。因此提出 Plan 240 以统一推进。

---

## 2. 目标与验收概览

| 目标 | 说明 | 验收方式 |
| --- | --- | --- |
| A. 对齐组织详情页面范式 | 职位详情必须复用 `TemporalMasterDetailView` 的“左侧版本列表 + 右侧页签”骨架（参考 docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md, 32/11/15 号文档）并保证交互一致 | Storybook 对比截图 + 交互录屏 + 评审纪要 |
| B. 消除页面结构性缺陷 | 在对齐布局的基础上重构 PositionDashboard/PositionTemporalPage，将 Tab、时间线、任职子页面抽象为统一容器，避免状态错配 | 前端 Vitest + Storybook 截图基线，代码评审 |
| C. 建立可靠的数据装载链路 | GraphQL hook 统一使用 Suspense-aware loading manager，支持 Promise.all + AbortController，解决等待超时 | `frontend/src/features/positions` 新 hook 单测 + E2E 通过 |
| D. 稳定 DOM 语义与可定位性 | 新增 data-testid/token，集中维护在 `frontend/src/shared/testing/positionsSelectors.ts`，并与组织页面 selector 规则保持一致 | 观测 diff + Playwright 录屏 |
| E. 可观测性与监控 | 增加 UI latency meter、GraphQL error logging、FP/FCP 指标上报 | 前端日志、`frontend/tests/e2e/position-lifecycle.spec.ts` 指标断言 |
| F. 文档 & Runbook | 更新 80/86/107 号引用、Plan 06 状态表，新增故障排查 Runbook | 文档 PR + 运行手册链接 |

---

## 3. 解锁条件与依赖

1. **硬依赖**  
   - Plan 242 命名抽象：已完成（解锁）。  
   - Plan 244 Timeline/Status 抽象：已完成（解锁）。  
   - Plan 232 的 P0 场景日志需与 240 同步推进，重构阶段必须提供最新 Chromium/Firefox 报告才能关闭（docs/development-plans/232-playwright-p0-stabilization.md:300-310）。  
   - 职位后端契约禁止修改（OpenAPI/GraphQL 以 80/86/107 为唯一事实来源）。若需新增字段，必须先更新契约并跑 `node scripts/generate-implementation-inventory.js`。  
2. **软依赖**  
   - Plan 06 可复用本计划输出的 E2E 结果，故需在每轮迭代后同步日志路径。  
   - 若引入性能优化（例如 bundle 拆分），需参考 Plan 232T 的资源体积阈值（<=5 MB 基线）。  

---

## 4. 范围与任务拆解

### 4.1 T0 – 前置工作（0.5 天）
- 校准本地环境：`make docker-up`, `make run-dev`, `make frontend-dev`，确认 Position API 与 GraphQL 可用。  
- 收集基线数据：导出当前 `frontend/test-results/position-*.json`、`logs/219E/position-*.log`；建立 `reports/plan240/baseline/`。

### 4.2 T1 – 组件结构重构与组织详情对齐（1.5 天）
- 以 `TemporalMasterDetailView` 为模板提炼 `PositionViewLayout`，左侧沿用组织详情的版本列表/时间轴组件（如需差异须在文档列出原因）；右侧页签顺序与组织详情保持一致（概览、任职、调动、时间线、版本、审计）。  
- 将 `PositionTemporalPage` 的本地状态改为基于 `usePositionRouteState` Hook，并确保版本切换只更新内容不触发编辑模态（对齐 docs/archive/development-plans/15-organization-timeline-navigation-investigation.md）。  
- 引入 `SuspenseBoundary` + skeleton，解决白屏，并复用组织页骨架样式。

### 4.2.1 UI 指南（详细要求）
| 区域 | 职位页面要求 | 组织参考 |
| --- | --- | --- |
| 顶部操作条 | 左对齐面包屑 + 标题 + 状态 pill；右侧依次放置“停用/启用”、“新增版本”、“更多操作”按钮，按钮顺序与 `TemporalMasterDetailView` 完全一致，禁用时使用同一 token | `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:600-680`、docs/archive/development-plans/11-organization-suspend-reactivate-ui-enhancement.md |
| 左侧版本/时间轴 | 复用 `TimelineComponent` 与 `VersionList` 样式，展现“当前版本高亮、历史版本展开、includeDeleted 过滤”三元素；点击节点只切换右侧内容，不弹 modal | docs/archive/development-plans/32-organization-delete-button-plan.md, 15-organization-timeline-navigation-investigation.md |
| 右侧页签导航 | 页签顺序固定（概览→任职→调动→时间线→版本→审计），标签文案与 Canvas Kit Tabs 示例保持一致；小屏折叠成 dropdown。需引用 `docs/reference/temporal-entity-experience-guide.md` 的 Canvas `Flex`/Tabs 指南（键盘导航、底边高亮）。 | docs/archive/development-plans/93-position-detail-tabbed-experience-plan.md:61-130、docs/reference/temporal-entity-experience-guide.md:23-30 |
| 概览卡片 | 使用与组织详情一致的卡片阴影、标题行、字段布局（两列栅格）；状态 pill、版本信息位置保持一致 | docs/reference/01-DEVELOPER-QUICK-REFERENCE.md:437 |
| 任职/调动列表 | 列表空态、加载骨架、分页条样式与组织“成员列表”一致；Action 区域放置在列表标题右侧 | docs/archive/development-plans/11-organization-suspend-reactivate-ui-enhancement.md 附图 |
| 时间线页签 | 纵向时间轴点使用相同颜色 token（current=primary,future=neutral,deleted=danger）；说明文字同组织版式 | docs/archive/development-plans/15-organization-timeline-navigation-investigation.md |
| 版本/审计页签 | 全量复用组织组件 `VersionHistoryTable`、`AuditHistorySection`，仅在 props 中传职位字段 | docs/archive/development-plans/25-version-field-contract-alignment-plan.md, 14/233 号审计方案 |
| Skeleton/空态 | 直接引用组织详情骨架（`TemporalDetailSkeleton`）；空态文案需保持“暂未加载数据/暂无数据”格式 | `frontend/src/features/temporal/components/TemporalDetailSkeleton.tsx` |

> 若因职位特性需要新增 UI 元素（如编制信息卡），必须：1) 在本计划附录记录；2) 与设计/架构评审确认不破坏组织对齐；3) 为组织页面准备兼容策略。

### 4.3 T2 – 数据加载与缓存治理（2 天）
- 实现 `usePositionDetailQuery`（统一 GraphQL fetch + abort + stale-while-revalidate）。  
- 对接 GraphQL `positionTimeline`, `positionAssignments`，提前发起 Promise.all，解决 `waitForGraphQL` 超时。  
- 在命令返回后触发 React Query cache invalidation，确保 Fill/Vacate 结果立即可见。

### 4.4 T3 – DOM 语义 & Design Token（1 天）
- 新建 `frontend/src/shared/testing/positionsSelectors.ts` 定义 testid 常量。  
- 所有按钮、Tab、时间线节点统一使用 token + `data-testid`。  
- 引入 design token 校验（ESLint 规则或 lint-staged 钩子），避免样式漂移。

### 4.5 T4 – 可观测性与性能（1 天）
 - 注入 `performance.mark` / `measure`，上报 `position.hydrate.done`、`position.tab.change`（详见 `docs/reference/temporal-entity-experience-guide.md` 第 7 节）。  
 - 将关键 GraphQL error 通过前端 logger pipeline 以 `[OBS] position.graphql.error` 形式输出，由 E2E 采集并落盘至 `logs/plan240/D/*`。  
- 分析 bundle：必要时开启路由级 code-splitting，保持 `position` chunk < 1.2 MB gzip。

### 4.6 T5 – 测试与文档收束（1 天）
- 更新 Vitest：`PositionDashboard.test.tsx`, `PositionTabs.test.tsx`，并新增用例验证组织/职位页面在相同版本切换脚本下行为一致。  
- Playwright：`position-tabs`, `position-lifecycle`, `position-crud-full-lifecycle` 重新录制 trace，确保 Chromium/Firefox 全绿并附日志；同时复用组织详情脚本（如 `temporal-management-integration`）验证布局一致性。  
- 文档：更新 `docs/development-plans/06-integrated-teams-progress-log.md` 的表格、`docs/archive/development-plans/80*.md` 的 Stage 4 附注，补充与组织详情对齐的截图/说明，并新增 `reports/plan240/execution-log.md`。

---

## 5. 里程碑与交付节奏

| 里程碑 | 时间（工作日） | 目标 | 产物 |
| --- | --- | --- | --- |
| M1 | Day 1 | 完成 T0/T1，提交组件重构 MR | MR#1 + Storybook 截图 |
| M2 | Day 3 | 完成 GraphQL Hook 与缓存治理（T2） | MR#2 + Hook 单测报告 |
| M3 | Day 4 | DOM 语义与监控落地（T3/T4） | MR#3 + 性能日志 + lint 规则 |
| M4 | Day 5 | Playwright 绿灯 + 文档收束（T5） | `logs/plan240/*.log` + 文档 PR |
| Sign-off | Day 6 | 更新 Plan 06、232T、80/86/107 引用，提交归档申请 | `reports/plan240/execution-log.md` v1.0 |

---

## 6. 验收标准

1. **体验一致性**：聘用“组织详情 vs 职位详情”对比脚本（相同操作序列）录屏，无论是版本切换、时间轴点击还是停用/启用入口呈现，交互完全一致；差异点需记录并获架构组批准。  
2. **功能**：职位详情页在 Skeleton 结束后 400 ms 内渲染 Tabs，切换 Tab 不触发整页刷新；Fill/Vacate 后 UI 与 GraphQL 数据立即同步。  
3. **稳定性**：`tests/e2e/position-tabs.spec.ts`、`tests/e2e/position-lifecycle.spec.ts` 在 Chromium + Firefox 连续 3 次通过；`waitForGraphQL` 不再超时。  
4. **可观测性**：新增性能指标在浏览器控制台可见，并可通过 `logs/ui/position-page.log` 追踪错误。  
5. **文档**：Plan 06/232T/80/86/107 与本计划互相引用，所有日志、trace、bundle 报告归档至 `logs/plan240/`。  
6. **治理**：无未解释的 `TODO-TEMPORARY`；若存在临时方案，必须登记后续计划并设定回收日期。

---

## 7. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
| --- | --- | --- | --- |
| GraphQL 契约需新增字段 | 中 | 低 | 先在 docs/api/ 更新 schema，跑实现清单脚本，确保唯一事实来源 |
| Firefox 行为差异 | 中 | 中 | 在 Playwright 配置里开启 `strictSelectors=true`，并收集 trace |
| Bundle 拆分影响其它路由 | 中 | 低 | 与 Plan 232T 协调，控制 chunk 大小并更新性能文档 |
| 监控图表尚未落地 | 低 | 中 | 先输出 `logs/ui/*` 文本基线，再申请接入 Observability 平台 |

---

## 8. 汇报与资料

- 每日同步：在 Plan 06 “当前阻塞”章节追加 240 计划进展、日志链接。  
- 计划文档维护：本文件即 Plan 240 唯一事实来源，所有决策、日志、脚本更新必须先记录于此，再同步相关计划。  
- 归档输出：`reports/plan240/execution-log.md`、`logs/plan240/*.log`、更新后的 Playwright trace、性能截图。
- Canvas Kit 合规：与设计/前端负责人一起走 Canvas Kit 组件评审（参考 docs/reference/temporal-entity-experience-guide.md、docs/archive/development-plans/105-navigation-ui-alignment-fix.md），确保新增/修改组件仍沿用官方 token 与交互；评审结论附于本计划附录并在 MR 中引用。

> 若后续出现新的职位相关缺陷或性能指标变化，需在本计划附录中追加条目，并在完成后归档至 `docs/archive/development-plans/240-position-management-page-refactor.md`。
