# 241A – TemporalEntityLayout 合流与最小接入

编号: 241A  
标题: TemporalEntityLayout 合流与最小接入（Shell/Sidebar/Tabs + 可观测性注入）  
创建日期: 2025-11-15  
状态: 待实施  
上游关联: 241（框架重构 · 收尾）、242（命名抽象 · 已验收）、244（Timeline/Status 抽象 · 已验收）、240（职位页面重构 · 已完成 A–D）

---

## 1. 背景与目标

- 背景：当前组织与职位详情页承载在两套骨架（组织端 `TemporalMasterDetailView`、职位端自有布局），导致后续 UI/指标/可访问性需双处维护，违反“跨层一致性”的执行预期。  
- 目标：提供中性骨架 `TemporalEntityLayout`（Shell/Sidebar/Tabs），以“最小接入”将组织/职位的外层容器合流到统一骨架，保持 DOM/testid 与交互不变，集中可观测性与键盘可达性注入点。

---

## 2. 范围与产物

- 新增骨架组件（中性）：
  - `frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`（导出命名空间：`TemporalEntityLayout.Shell/Sidebar/Tabs`；仅承载容器与行为注入，不含领域组件）
- 最小接入改造：
  - 组织：将 `TemporalMasterDetailView` 的外层容器替换为 `TemporalEntityLayout.Shell`；内部业务组件保持不变
  - 职位：将 `PositionDetailView` 的外层 `Box/Flex` 替换为 `TemporalEntityLayout.Shell`；内部业务组件保持不变
- 选择器与可访问性：
  - 保持既有 `data-testid`，统一从 `frontend/src/shared/testids/temporalEntity.ts` 导出（`temporalEntitySelectors`）
  - 骨架层提供 Tab 导航键盘可达性（左右键切换、Enter 激活）
- 可观测性注入：
  - 骨架层注入 `performance.mark('obs:temporal:hydrate')`、Tab 切换标记（仅 performance 标记）；事件按实体前缀转发到既有命名（职位沿用 `position.*`）；组织端不新增未定义事件名，避免第二事实来源

---

## 3. 验收标准

1) DOM 与交互保持稳定：关键 testid 不变，E2E 断言无需更新（排除对齐性变更）  
2) 可观测性：浏览器控制台可见职位详情页的 `[OBS] position.*` 事件；组织页仅有 `performance.mark('obs:temporal:*')`（不新增 `organization.*` 事件名）  
3) 可访问性：Tab 导航在键盘左右键与 Enter 触发下可达，`role=tablist`/`role=tab`/`aria-selected` 等属性正确  
4) 选择器唯一来源：改造后组件与测试均只从 `temporalEntitySelectors` 获取 testid  
5) E2E 冒烟通过：`temporal-management-integration`（组织、职位各 1 次/浏览器），Chromium/Firefox 各 1 轮

---

## 4. 依赖与边界

- 不改变 REST/GraphQL 契约；不调整 OpenAPI/GraphQL Schema；不新增事件名至权威指南
- 命名与指南唯一来源：`docs/reference/temporal-entity-experience-guide.md`
- 选择器唯一来源：`frontend/src/shared/testids/temporalEntity.ts`
- 已完成依赖：242/244；240 的页面改动以本骨架为外层容器承载

---

## 5. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
|---|---|---|---|
| 合流导致 E2E 回归 | 中 | 中 | 严格保持 testid 与 DOM 层级，骨架仅包裹外层容器；先跑冒烟再提交 MR |
| 组织端事件命名未定义 | 低 | 低 | 仅注入 performance 标记，不新增 `organization.*` 事件；待权威指南扩展后再开启 |
| Tab 键盘行为差异 | 低 | 中 | 严格对标职位页当前行为；增加 vitest/RTL 快速用例验证 |

---

## 6. 执行步骤

1) 新建骨架组件并落地 perf 标记与 tab 键盘支持  
2) 组织/职位页面外层容器替换为骨架组件（内层卡片/表单/列表不动）  
3) 选择器引用统一回到 `temporalEntitySelectors`（若仍有遗留，先以兼容导出保留 1 个迭代，登记 `// TODO-TEMPORARY:`）  
4) 跑 E2E 冒烟（Chromium/Firefox）并记录日志

---

## 7. 产出与登记

- 代码：`frontend/src/features/temporal/layout/TemporalEntityLayout.tsx` + 最小接入变更  
- 日志：`logs/plan241/A/e2e-smoke-{chromium,firefox}.log`、`frontend/playwright-report/**`  
- 备注：不产生任何契约变更；事件命名不更新权威文档

---

## 8. 退出准则

- 两端页面均运行在统一骨架上，E2E 冒烟通过；指标与键盘可达性符合预期；无 testid 抖动  
- 所有改动登记于本文件；如发生兼容性保留，须以 `// TODO-TEMPORARY:` 标注并设定回收日期

