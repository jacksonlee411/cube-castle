# Plan 240A – 职位详情 Layout 对齐与骨架替换

编号: 240A  
上游: Plan 240（职位管理页面重构） · 依赖 Plan 242/247（命名与文档治理闭环）  
状态: 进行中（T1）

—

## 目标
- 使职位详情页落在与组织一致的“左侧版本导航 + 右侧六页签”骨架上，消除自建布局差异。
- 不改变契约与数据获取路径，优先实现布局与选择器对齐，降低回归面。

## 范围
- 组件：`frontend/src/features/positions/PositionDetailView.tsx` 与相关子组件（版本列表/工具条/详情卡）。
- 布局：对齐 `TemporalMasterDetailView` 的区域划分、间距、标题/工具条顺序；仅做视觉/语义对齐，不得复制组织骨架的状态/数据流程代码。
- 选择器：统一使用 `temporalEntity-*` 前缀（集中在 `shared/testids/temporalEntity`）。

## 临时方案约束
- 禁止复制组织端骨架逻辑（状态管理、数据流程、内部 Hook）。如确需过渡性适配层（shim），必须以最薄方式实现并标注：`// TODO-TEMPORARY(240A): 原因 | 负责人与移除时间（≤1个迭代）`，并在 MR 说明列出移除计划。
- 唯一事实来源：布局/交互/可访问性以 `docs/reference/temporal-entity-experience-guide.md` 为准，不在本计划复制规范正文，仅引用与量化验收项。

## 不做（本子计划）
- 不引入框架级抽象（该项进入 Plan 241）。
- 不变更 REST/GraphQL 契约与字段。

## 任务清单
1) 结构对齐（视觉/语义）  
   - 设计 Token 与断点（量化）：
     - 左侧版本导航：桌面端固定 320px；宽度 < 960px 时收敛为抽屉（点击打开/选择后自动关闭）
     - 页内垂直间距：24px（SimpleStack）；左卡片与右主体留 `space.l`
     - 六页签顺序固定：概览 → 任职记录 → 调动记录 → 时间线 → 版本历史 → 审计历史
   - A11y（量化断言）：
     - Tab 支持键盘左右切换、焦点可见；版本行具备 `aria-selected`，与时间轴高亮同步
   - 引用规范：其余视觉/交互细节依从 `docs/reference/temporal-entity-experience-guide.md`（不复制正文）
2) 选择器统一  
   - 原则：组件与用例仅通过 `frontend/src/shared/testids/temporalEntity.ts` 暴露/引用选择器，不得硬编码字符串
   - 迁移策略：保留旧 test id 一次迭代的兼容映射；新增旧前缀计数必须不升高（`scripts/quality/selector-guard-246.js`）
   - 必迁移示例（职位端现存硬编码 → 统一选择器）：
     - `position-create-page` → `temporalEntitySelectors.position.temporalPage`
     - `position-mock-banner` → `temporalEntitySelectors.position.mockBanner`
     - `position-detail-layout` → `temporalEntitySelectors.position.detailCard`（或在 selectors 中新增 `detailLayout` 后切换）
     - `position-detail-error` → 在 selectors 增补 `detailError`（`temporal-position-detail-error`）并改造引用
     - `position-tab-{key}` → 在 selectors 增补 `tabId: (key) => string`，统一为 `temporal-position-tab-{key}`
     - `position-edit-button` → 在 selectors 增补 `editButton`
     - `position-version-button` → 在 selectors 增补 `createVersionButton`
   - 用例替换：批量将 Playwright 用例引用切换到集中选择器（保留旧选择器 fallback 1个迭代）
3) 基线与验证  
   - Storybook 对比（组织 vs 职位）并生成截图：`reports/plan240/baseline/storybook/*.png`
   - E2E 冒烟：恢复 `tests/e2e/position-tabs.spec.ts` 的关键元素渲染（页眉、版本导航/抽屉、六页签、选中态）
   - 浏览器覆盖与重复跑：Chromium/Firefox 各 2 次通过；保存 trace：`logs/plan240/A/playwright-trace/*`
   - 门禁：`node scripts/quality/architecture-validator.js`、`node scripts/quality/document-sync.js` 退出码 0；`node scripts/quality/selector-guard-246.js` 无新增旧前缀

## 验收标准
- 视觉/语义一致：与组织详情在骨架/间距/断点/Tab 行为上一致；任何差异在 MR 记录原因与后续修正计划
- A11y 通过：Tab 键可达、左右切换、焦点可见；版本行 `aria-selected` 与时间轴高亮同步
- E2E 恢复：`position-tabs.spec.ts` 在 Chromium/Firefox 各 2 次通过；关键元素（页眉、左侧导航/抽屉、六页签、选中态）断言成功
- 选择器守卫：`selector-guard-246` 基线计数不升高；组件与用例均引用集中选择器
- 门禁通过：`architecture-validator`、`document-sync` 退出码 0

## 证据与落盘
- Storybook 截图：`reports/plan240/baseline/storybook/*.png`（组织 vs 职位）
- E2E 输出与 trace：`logs/plan240/A/playwright-*.log`、`logs/plan240/A/playwright-trace/*`
- 守卫与门禁：`logs/plan240/A/architecture-validator.log`、`logs/plan240/A/document-sync.log`、`logs/plan240/A/selector-guard.log`
- 基线与对比说明：`reports/plan240/baseline/README.md`

## 回滚策略
- Feature Flag：`VITE_POSITION_LAYOUT_V2`（默认 `true`）。当设为 `false` 时，路由/职位详情回退到旧布局（不改契约）。
- 触发位置：路由入口（`frontend/src/App.tsx` 的 Position 路由）或 `PositionDetailView` 顶层根据 flag 决定使用新/旧布局。
- 回滚操作流程：切换 env → 重启前端 → 复跑冒烟；回滚日志记录至 `logs/plan240/A/rollback-*.log`
