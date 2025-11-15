# Plan 241 – 前端页面与框架一体化重构

**编号**: 241  
**标题**: Unified Frontend Layout & Data Framework Refactor  
**创建日期**: 2025-11-10  
**关联计划**: Plan 240（职位页面重构）、Plan 232/232T（Playwright 稳定）、Plan 06（集成验证）  
**状态**: 部分完成 · 验收未通过（暂不关闭）· 2025-11-15 复核  
（原“暂缓”解锁条件已满足：Plan 242/244 均已验收，通过后恢复执行；本次为阶段性验收结论与收尾计划）

---

## 0. 对齐与合规（AGENTS.md）

- 单一事实来源与跨层一致性：不新增第二事实来源；骨架/Hook/选择器均引用既有权威入口  
  - 选择器：`frontend/src/shared/testids/temporalEntity.ts`（`temporalEntitySelectors`）  
  - 命名与指南：`docs/reference/temporal-entity-experience-guide.md`  
  - 类型与 Hook：`frontend/src/shared/types/temporal-entity.ts`、`frontend/src/shared/hooks/useTemporalEntityDetail.ts`  
- 先契约后实现：不引入 REST/GraphQL 契约变更；若确需字段，先更新 `docs/api/*` 并跑实现清单脚本。  
- CQRS 原则：仅涉及前端查询与 UI 架构，不改变命令/查询边界。  
- Docker 强制：不涉及宿主服务安装，开发/验收沿用 `make docker-up` 等既有流程。  
- 临时方案治理：临时兼容以 `// TODO-TEMPORARY:` 标注并登记，回收期不超过一个迭代。  

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
| A. 统一时态详情骨架 | 将 `TemporalMasterDetailView` 抽象为中性的 `TemporalEntityLayout`（含 Shell/Sidebar/Tabs），组织/职位等实体共享骨架与 Skeleton | E2E 录屏/截图 + 组件 API 评审（Storybook 可选） |
| B. 标准化详情数据 Hook | 采纳并增强统一 Hook `useTemporalEntityDetail`（必要时补充轻量 `createTemporalDetailLoader` 作为内部工厂），由 `usePositionDetail` 与 `useEnterpriseOrganizations` 的 detail 适配器薄封装调用，实现统一的 Suspense/caching/失效规则 | Hook 单测 + React Query 日志 |
| C. DOM/TestId 合规 | 将选择器集中到 `frontend/src/shared/testids/temporalEntity.ts`（`temporalEntitySelectors`），组织/职位 Playwright 复用 | Playwright trace + selector diff |
| D. 可观测性基线 | 在共享 layout 中注入 `performance.mark` + logger 事件，输出到统一前端管线（`logger`）而非 ad-hoc 文件 | 浏览器日志 + `frontend/tests/e2e/temporal-management-integration.spec.ts` 指标断言 |
| E. 文档/Runbook 对齐 | 更新 Plan 06、Plan 232 引用；将框架指南内容并入 `docs/reference/temporal-entity-experience-guide.md`（不新增第二事实来源） | 文档 PR + 评审纪要 |

---

## 3. 解锁条件与依赖

1. **硬依赖**  
   - Plan 244（Timeline/Status 抽象）验收通过：Adapter/StatusMeta 已合并、契约同步、基础 E2E 绿灯；  
   - Plan 232/232T 最新 E2E 日志必须可用，以验证统一 selector 后的稳定性。  
   - OpenAPI/GraphQL 契约不可漂移；若需字段扩展，先更新 `docs/api/` 并跑 `node scripts/generate-implementation-inventory.js`（AGENTS.md:13-17）。  
2. **软依赖**  
   - 与设计团队就 Canvas Kit Token/交互保持同步，引用 `docs/reference/temporal-entity-experience-guide.md`。  
   - 与 Plan 240 协调，确保其页面改动直接落在新框架上，而非再写临时组件。

---

## 4. 范围与任务拆解

### 4.1 T0 – 现状盘点（0.5 天）
- 环境前置校验（Docker 强制，遵循 AGENTS.md）：
  - `make docker-up` → `make run-dev` → `make frontend-dev`
  - 健康检查：`curl http://localhost:9090/health`、`curl http://localhost:8090/health`（预期 200）
  - 鉴权准备：`make jwt-dev-mint`（`.cache/dev.jwt` 存在）
- 运行 `node scripts/generate-implementation-inventory.js`，梳理组织/职位相关组件与 hook。  
- 导出当前 Playwright selector 使用清单（`tests/e2e/*position*.spec.ts`, `tests/e2e/*organization*.spec.ts`）。  
- 建立 `reports/plan241/baseline/`：包含 DOM diff、bundle size、React Query 缓存日志。

### 4.2 T1 – Layout 抽象（1.5 天）
- 交付中性骨架：新建 `frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`，拆分 `Shell/Sidebar/Tabs`；仅承载容器/Slot 与共享行为，不包含领域组件。  
- 职责边界：`TemporalEntityPage` 负责路由校验与实体适配；`TemporalEntityLayout` 仅提供页面骨架与统一注入点（可观测性、键盘可达性）。  
- 最小接入：组织端将 `TemporalMasterDetailView` 外层容器替换为新骨架；职位端将现有外层 `Box/Flex` 替换为新骨架。内部卡片/列表/表单不改动。  
- 可观测性注入：在骨架层注入 `obs` 性能标记与 tab 切换事件；事件前缀采用“实体前缀转发”策略（职位沿用 `position.*`；组织如未在权威指南定义，仅注入 performance.mark，不新增事件名，避免第二事实来源）。  
- 选择器稳定：骨架保留既有 `data-testid` 并统一由 `temporalEntitySelectors` 输出；避免 E2E 断裂。  
- 输出 Playwright/E2E 场景（组织 vs 职位）确保视觉一致；如已配置 Storybook，可补充对比截图作为辅证（非阻塞）。

### 4.3 T2 – Data Hook 统一（2 天）
- 采纳并增强统一 Hook：`frontend/src/shared/hooks/useTemporalEntityDetail.ts`；仅在必要时补充 `createTemporalDetailLoader` 作为内部工厂，避免重复实现。  
- 新增组织薄封装：导出 `useOrganizationDetail`（内部仅透传 `useTemporalEntityDetail('organization', ...)`）。  
- 外部约束：通过 ESLint 自定义规则限制外部仅引用权威入口，禁止重复实现 Hook 的入口路径。  
- 新增 Hook 单测覆盖：错误状态、租户切换、includeDeleted 切换、cache invalidation 正确性与无重复 fetch（结合 React Query Devtools 日志）。

### 4.4 T3 – Selector & Token 治理（1 天）
- 统一到 `frontend/src/shared/testids/temporalEntity.ts` 并导出 `temporalEntitySelectors`；作为测试与运行时代码的唯一选择器事实来源。  
- 在 ESLint 增设“禁止硬编码 data-testid”规则（白名单仅允许 `frontend/src/shared/testids/temporalEntity.ts`），避免字符串散落。  
- 守卫脚本：沿用 `scripts/quality/selector-guard-246.js`，冻结旧前缀增量并建立基线 `reports/plan246/baseline.json`；基线不升。

### 4.5 T4 – 可观测性统一（1 天）
- 在 `TemporalEntityLayout.Shell` 注入 `performance.mark('obs:temporal:hydrate')` 与 tab 切换标记；事件经骨架按实体前缀转发到现有命名（职位：`position.*`）。  
- 输出通道与门控：复用 `obs` 与 `logger`，遵循 `VITE_OBS_ENABLED` 与 `VITE_ENABLE_MUTATION_LOGS` 门控；生产不输出信息级 `[OBS]`。  
- Playwright 指标断言：复用 `frontend/tests/e2e/temporal-management-integration.spec.ts` 并在职位相关用例中监听 console；组织端仅断言性能标记存在（不新增事件名）。

### 4.6 T5 – 文档与迁移（1 天）
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md`、Plan 232/240 文件，说明框架重构影响。  
- 将布局/Hook/selector 规格并入 `docs/reference/temporal-entity-experience-guide.md`（不新增平行指南，避免第二事实来源）。  
- 编制 Runbook：如何在新框架上新增时态实体页面。

---

## 5. 里程碑

| 里程碑 | 时间 | 目标 | 产物 |
| --- | --- | --- | --- |
| M1 | Day 1 | 完成 T0/T1，提交 Layout 抽象 MR | MR#1 + E2E 录屏/截图（Storybook diff 可选） |
| M2 | Day 3 | Hook 采纳/增强完成（T2） | MR#2 + Hook 单测报告 |
| M3 | Day 4 | Selector/Token + 可观测性落地（T3/T4） | MR#3 + lint 规则 + 指标日志 |
| M4 | Day 5 | 文档/Runbook 更新（T5） | 文档 PR + Runbook |

---

## 6. 验收标准

1. **布局一致性**：组织与职位详情在 Playwright 录屏/截图对比下无视觉差异；如已配置 Storybook，可作为辅证；任何差异需在 Plan 241 附录记录并获批准。  
2. **Hook 行为**：`useTemporalEntityDetail`（经 `usePositionDetail` / `useOrganizationDetail` 适配层调用）无重复 fetch，命令链路后触发正确的 queryKey 失效并完成刷新；对应 `performance.mark('TemporalEntity:*')` 事件与 logger 日志可观测。  
3. **Selector 稳定性**：`tests/e2e/position-tabs.spec.ts`、`frontend/tests/e2e/temporal-management-integration.spec.ts` 在 Chromium/Firefox 连续 3 次通过，且只引用 `temporalEntitySelectors`；ESLint 规则禁止硬编码 `data-testid`；`npm run guard:selectors-246` 与 `npm run guard:plan245` 均通过，且基线计数不升高。  
4. **可观测性**：浏览器控制台出现 `TemporalLayout` 指标日志，Playwright 指标断言通过（基于事件命中与行为验证，而非硬时长阈值）。  
5. **文档治理**：框架指南内容并入 `temporal-entity-experience-guide.md`，Plan 06/232/240 引用更新；所有日志、录屏归档至 `logs/plan241/`、`reports/plan241/`。

---

## 7. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
| --- | --- | --- | --- |
| Layout 重构引发组织页面回归 | 高 | 中 | 先为现有组件补齐 Storybook/Vitest，Diff 驱动 MR 审查 |
| Hook 改动触发契约变更 | 中 | 低 | 若需新字段，提前更新 `docs/api/` 并跑实现清单脚本 |
| Selector 改动导致 E2E 大面积更新 | 中 | 中 | 编写 codemod 或脚本批量替换，保留旧 selector fallback 一周；以 `selector-guard-246` 作为唯一强约束入口 |
| 可观测性事件噪声增加 | 低 | 中 | 默认仅在 `import.meta.env.DEV` 输出，线上通过 feature flag 控制 |
| Page/Layout 双入口职责重叠 | 中 | 中 | 明确“Page=路由/适配、Layout=骨架组件”；提供 rename/alias 过渡策略，避免双轨 |
| 组织端事件命名未定义 | 低 | 低 | 骨架仅注入 performance.mark；不新增未定义事件名；待权威指南扩展后再开启 org.* 事件 |

---

## 8. 汇报与资料

- 每日向 Plan 06 提交阻塞/进度；Plan 232/240 需引用本计划交付物。  
- 所有日志、Storybook 截图、指标报告保存至 `logs/plan241/`、`reports/plan241/`。  
- 本文件为 Plan 241 唯一事实来源；签核后移至 `docs/archive/development-plans/241-frontend-framework-refactor.md`。

---

## 9. 阶段性验收结论（2025-11-15）

结论：不满足关闭条件，保留为“部分完成”，进入收尾阶段（Scope 缩减不跨层，不新增第二事实来源）。

- 布局抽象（T1）未达标  
  - 仍使用 `TemporalMasterDetailView`（frontend/src/features/temporal/components/TemporalMasterDetailView.tsx）作为组织端骨架；未交付中性 `TemporalEntityLayout.*`。  
  - 职位端采用 `PositionDetailView` 自有布局（frontend/src/features/positions/PositionDetailView.tsx），未与组织端共享骨架。
- 数据 Hook 统一（T2）部分达成  
  - 已提供 `useTemporalEntityDetail`（frontend/src/shared/hooks/useTemporalEntityDetail.ts），并在职位与组织主从视图内部消费（示例：frontend/src/features/positions/PositionDetailView.tsx:96、frontend/src/features/temporal/components/hooks/useTemporalMasterDetail.ts:74）。  
  - 未完成“薄封装统一入口”与 Hook 单测；`useOrganizationDetail` 薄封装缺失；`__tests__` 内无该 Hook 的覆盖。
- Selector/Token（T3）部分达成  
  - 已集中选择器 `frontend/src/shared/testids/temporalEntity.ts` 并广泛应用于 E2E。  
  - ESLint 禁止硬编码 data-testid 的规则与 `guard:selectors-246` 脚本未落地，仍可能出现散落选择器字符串。
- 可观测性（T4）部分达成  
  - 职位详情已按权威文档注入 `[OBS] position.*` 事件与 `performance.mark`（frontend/src/features/positions/PositionDetailView.tsx）。  
  - 共享 Layout 未交付，故未集中注入 `TemporalEntity:*` 级别指标。
- 文档与 Runbook（T5）基本对齐  
  - 权威文档已建立并指向 `docs/reference/temporal-entity-experience-guide.md`（见 Plan 247 结论）；Plan 06/232 已更新引用。  
  - 本 Plan 未产出 `logs/plan241/**`、`reports/plan241/**` 验收资产，E2E 连跑 3 轮的证据缺失。

基于上述核查，本计划“不可关闭”；进入收尾阶段以完成最小必要交付，确保不再形成第二事实来源。

---

## 10. 收尾计划（不改变对外契约）

目标：在不改动 OpenAPI/GraphQL 契约的前提下，于 1–2 个工作日完成 241 的最小闭环。

1) 交付 `TemporalEntityLayout`（中性骨架）  
   - 输出 `frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`，拆分 `Shell/Sidebar/Tabs`；组织端替换 `TemporalMasterDetailView` 外层容器为新骨架，职位端替换现有 `Box/Flex` 外层容器。  
   - 在 Layout 中注入 `obs.markStart('obs:temporal:hydrate')/markEndAndMeasure` 与统一 tab 事件桥，复用现有 `obs`。
2) Hook 统一与单测  
   - 新增 `useOrganizationDetail`（薄封装至 `useTemporalEntityDetail('organization', ...)`）；  
   - 为 `useTemporalEntityDetail` 补充 Vitest 覆盖：错误状态、租户切换、`includeDeleted`、失效/刷新。产物落盘 `logs/plan241/hook-tests.log`。  
3) Selector 守卫  
   - ESLint 规则：禁止在 `src/**/*.{ts,tsx}` 直接硬编码 `data-testid`，仅允许从 `shared/testids/temporalEntity.ts` 导入；  
   - 脚本 `npm run guard:selectors-246`：扫描 diff 中新增的硬编码选择器并失败，基线在 `reports/plan246/baseline.json`。  
4) E2E 验收资产  
   - 复跑 `frontend/tests/e2e/temporal-management-integration.spec.ts` 与 `position-tabs.spec.ts`（Chromium/Firefox 各 3 轮）；  
   - 控制台收集 `[OBS]` 日志，产物归档至 `logs/plan241/e2e/*.log` 与报告快照 `frontend/playwright-report/**`。

退出准则（收尾版）  
- 组织/职位页面均运行在 `TemporalEntityLayout` 上；  
- `useTemporalEntityDetail` 单测通过，`useOrganizationDetail` 对外可用；  
- ESLint 规则与 selectors 守卫生效（本地/CI）；  
- E2E 指标与 UI 断言通过（2 浏览器 × 3 轮），证据落盘。

说明  
- 上述工作量不引入新契约与新事实来源；命名统一仍以 `docs/reference/temporal-entity-experience-guide.md` 为权威。  
- 若遇阻塞（如 Playwright 浏览器工件缺失），允许在 Plan 06 中记录并延期 E2E 连跑次数，但不得影响布局/Hook/守卫的落地。
