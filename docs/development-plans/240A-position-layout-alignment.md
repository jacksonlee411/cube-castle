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
- 布局：对齐 `TemporalMasterDetailView` 的区域划分、间距、标题/工具条顺序；必要时仅做视觉/语义对齐，不强行复用内部 Hook。
- 选择器：统一使用 `temporalEntity-*` 前缀（集中在 `shared/testids/temporalEntity`）。

## 不做（本子计划）
- 不引入框架级抽象（该项进入 Plan 241）。
- 不变更 REST/GraphQL 契约与字段。

## 任务清单
1) 结构对齐（视觉/语义）  
   - 顶部工具条顺序、版本导航宽度与抽屉行为、六页签顺序与命名对齐组织端规范。  
   - 对齐 `temporal-entity-experience-guide.md` 中的 A11y 与可用性要求（键盘导航等）。
2) 选择器统一  
   - 仅暴露 `temporalEntity-*` 前缀的 test id；移除或 Deprecated 旧 test id 的直接引用。  
   - 更新 Playwright 用例对应的选择器集中入口（如需）。
3) 基线与验证  
   - Storybook 对比截图（组织 vs 职位）；E2E嵌套冒烟恢复 `position-tabs.spec.ts` 的关键元素渲染。  
   - `node scripts/quality/architecture-validator.js`、`node scripts/quality/document-sync.js` 必须 0 退出。

## 验收标准
- 视觉/语义一致：对比组织详情，布局区域与交互行为一致；差异需在 MR 说明。
- E2E 恢复：`position-tabs.spec.ts` 渲染成功；关键元素 test id 对齐统一规则。
- 门禁通过：架构验证/文档同步均通过；无新增旧前缀选择器引用。

## 证据与落盘
- 日志：`logs/plan240/A/*.log`（Storybook 截图索引、E2E 输出、架构与文档同步日志）。
- 基线：`reports/plan240/baseline/`（变更前后对比说明与截图）。

## 回滚策略
- 若对齐后 E2E 回归明显，保留 UI 层开关以回退到旧布局（不改契约）。

