# Plan 241 – 前端页面与框架一体化重构

**编号**: 241  
**标题**: Unified Frontend Layout & Data Framework Refactor  
**创建日期**: 2025-11-10  
**关联计划**: Plan 240（职位页面重构）、Plan 232/232T（Playwright 稳定）、Plan 06（集成验证）  
**状态**: 暂缓 · 依赖 Plan 242 命名抽象完成后再恢复  

---

## 1. 背景与动因

1. Plan 240 在 T1/T2 中暴露“布局与数据 Hook 重复造轮子”风险：职位页面试图重新实现 `TemporalMasterDetailView` 骨架与 `usePositionDetail` 行为（docs/development-plans/240-position-management-page-refactor.md:52-78），违背唯一事实来源原则。
2. 现有组织/职位详情在组件、GraphQL hook、测试标识等方面缺乏共享抽象，导致 DOM/test id 漂移与 Playwright 选择器不稳定（docs/development-plans/240-position-management-page-refactor.md:70-88）。
3. 232/232T 的 P0 用例持续失败与等待链路超时（docs/development-plans/232t-test-checklist.md:25-69），说明前端数据加载与可观测性必须统一治理，而非在单个页面内“补丁式”处理。

> **结论**：需以 Plan 241 统筹 UI 骨架、数据 Hook、Design Token 与可观测性框架，提供一次性重构，避免 Plan 240 等专项重复造轮子。

---

## 2. 目标与验收概览

| 目标 | 说明 | 验收方式 |
| --- | --- | --- |
| A. 统一时态详情骨架 | 将 `TemporalMasterDetailView` 抽象为中性的 `TemporalEntityLayout`（含 Shell/Sidebar/Tabs），组织/职位等实体共享骨架与 Skeleton | Storybook 对比 + 组件 API 评审 |
| B. 标准化详情数据 Hook | 构建通用的 `useTemporalEntityDetail`/`createTemporalDetailLoader`，由 `usePositionDetail` 与 `useEnterpriseOrganizations` 的 detail 适配器薄封装调用，实现同一套 Suspense/caching 规则 | Hook 单测 + React Query 日志 |
| C. DOM/TestId 合规 | 将选择器集中到 `frontend/src/shared/testing/temporalSelectors.ts`，组织/职位 Playwright 复用 | Playwright trace + selector diff |
| D. 可观测性基线 | 在共享 layout 中注入 `performance.mark` + logger 事件，输出到统一前端管线（`logger`）而非 ad-hoc 文件 | 浏览器日志 + `frontend/tests/e2e/temporal-management-integration.spec.ts` 指标断言 |
| E. 文档/Runbook 对齐 | 更新 Plan 06、Plan 232 引用，提交新的框架指南 | 文档 PR + 评审纪要 |

---

## 3. 解锁条件与依赖

1. **硬依赖**  
   - Plan 232/232T 最新 E2E 日志必须可用，以验证统一 selector 后的稳定性。  
   - OpenAPI/GraphQL 契约不可漂移；若需字段扩展，先更新 `docs/api/` 并跑 `node scripts/generate-implementation-inventory.js`（AGENTS.md:13-17）。  
2. **软依赖**  
   - 与设计团队就 Canvas Kit Token/交互保持同步，引用 `docs/reference/positions-tabbed-experience-guide.md`。  
   - 与 Plan 240 协调，确保其页面改动直接落在新框架上，而非再写临时组件。

---

## 4. 范围与任务拆解

### 4.1 T0 – 现状盘点（0.5 天）
- 运行 `node scripts/generate-implementation-inventory.js`，梳理组织/职位相关组件与 hook。  
- 导出当前 Playwright selector 使用清单（`tests/e2e/*position*.spec.ts`, `tests/e2e/*organization*.spec.ts`）。  
- 建立 `reports/plan241/baseline/`：包含 DOM diff、bundle size、React Query 缓存日志。

### 4.2 T1 – Layout 抽象（1.5 天）
- 将 `TemporalMasterDetailView` 重构为 `TemporalEntityLayout` 命名空间（`TemporalEntityLayout.Shell`、`TemporalEntityLayout.Sidebar`、`TemporalEntityLayout.Tabs`），保留组织现有功能（frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:150-260）。  
- `PositionTemporalPage`/`OrganizationTemporalPage` 仅作为适配层，分别通过 slot/props 注入特定卡片或按钮，禁止复制布局。  
- 输出 Storybook 场景（组织 vs 职位）确保视觉一致，并为未来实体预留扩展说明。

### 4.3 T2 – Data Hook 统一（2 天）
- 实现 `createTemporalDetailLoader` + `useTemporalEntityDetail` 泛型 Hook，封装 Suspense、`AbortController`、Promise.all、cache invalidation（参考 frontend/src/shared/hooks/useEnterprisePositions.ts:1280-1521）。  
- `usePositionDetail` 与 `useEnterpriseOrganizations`（导出 `useOrganizationDetail` 薄封装）都调用该共享实现，继续通过 ESLint 限制外部只能引用权威入口。  
- 新增 Hook 单测覆盖：错误状态、租户切换、includeDeleted 切换，并验证多实体共享缓存策略。

### 4.4 T3 – Selector & Token 治理（1 天）
- 新建 `frontend/src/shared/testing/temporalSelectors.ts`，集中维护组织/职位组件 `data-testid`。  
- 更新所有职位/组织页面与 Playwright 用例引用该选择器。  
- 在 ESLint 规则或 lint-staged 脚本中校验 token/selector 引用，防止手写字符串。

### 4.5 T4 – 可观测性统一（1 天）
- 在 `TemporalEntityLayout.Shell` 注入 `performance.mark('TemporalEntity:hydrate')`、`performance.mark('TemporalEntity:tab-switch')`，并利用 `logger` 输出结构化事件（frontend/src/shared/utils/logger.ts:1-75），同时扩展 logger 以 `force=true` 或 `VITE_ENABLE_TEMPORAL_METRICS` 方式在 CI/Playwright 中可观测。  
- 将事件上传路径与 Plan 232 的可观测性指标保持一致，避免写入本地 `logs/`（Plan 240 的 T4 要求不再单独实现）。  
- 为 Playwright 场景增加指标断言（复用 `frontend/tests/e2e/temporal-management-integration.spec.ts` 并在 `tests/e2e/position-tabs.spec.ts` 中监听 console）。

### 4.6 T5 – 文档与迁移（1 天）
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md`、Plan 232/240 文件，说明框架重构影响。  
- 新增 `docs/reference/temporal-layout-guide.md` 概述组件 API、Hook 签名、selector 规范。  
- 编制 Runbook：如何在新框架上新增时态实体页面。

---

## 5. 里程碑

| 里程碑 | 时间 | 目标 | 产物 |
| --- | --- | --- | --- |
| M1 | Day 1 | 完成 T0/T1，提交 Layout 抽象 MR | MR#1 + Storybook diff |
| M2 | Day 3 | Hook 重构完成（T2） | MR#2 + Hook 单测报告 |
| M3 | Day 4 | Selector/Token + 可观测性落地（T3/T4） | MR#3 + lint 规则 + 指标日志 |
| M4 | Day 5 | 文档/Runbook 更新（T5） | 文档 PR + Runbook |

---

## 6. 验收标准

1. **布局一致性**：组织与职位详情在 Storybook/Playwright 录屏对比下无视觉差异；任何差异需在 Plan 241 附录记录并获批准。  
2. **Hook 行为**：`useTemporalEntityDetail`（经 `usePositionDetail` / `useOrganizationDetail` 适配层调用）在 Skeleton 完成后 400 ms 内返回数据，命令链路后 1 s 内完成 cache invalidation；React Query Devtools 无重复 fetch。  
3. **Selector 稳定性**：`tests/e2e/position-tabs.spec.ts`、`frontend/tests/e2e/temporal-management-integration.spec.ts` 在 Chromium/Firefox 连续 3 次通过，且只引用 `temporalSelectors`.  
4. **可观测性**：浏览器控制台出现 `TemporalLayout` 指标日志，Playwright 指标断言通过。  
5. **文档治理**：新框架指南已发布，Plan 06/232/240 引用更新；所有日志、录屏归档至 `logs/plan241/`、`reports/plan241/`。

---

## 7. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
| --- | --- | --- | --- |
| Layout 重构引发组织页面回归 | 高 | 中 | 先为现有组件补齐 Storybook/Vitest，Diff 驱动 MR 审查 |
| Hook 改动触发契约变更 | 中 | 低 | 若需新字段，提前更新 `docs/api/` 并跑实现清单脚本 |
| Selector 改动导致 E2E 大面积更新 | 中 | 中 | 编写 codemod 或脚本批量替换，保留旧 selector fallback 一周 |
| 可观测性事件噪声增加 | 低 | 中 | 默认仅在 `import.meta.env.DEV` 输出，线上通过 feature flag 控制 |

---

## 8. 汇报与资料

- 每日向 Plan 06 提交阻塞/进度；Plan 232/240 需引用本计划交付物。  
- 所有日志、Storybook 截图、指标报告保存至 `logs/plan241/`、`reports/plan241/`。  
- 本文件为 Plan 241 唯一事实来源；签核后移至 `docs/archive/development-plans/241-frontend-framework-refactor.md`。
