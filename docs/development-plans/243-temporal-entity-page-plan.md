# Plan 243 – Temporal Entity Page 抽象实施方案

**关联主计划**: Plan 242（T1）  
**目标窗口**: Day 1-3  
**范围**: 将 `OrganizationTemporalPage`、`PositionTemporalPage` 抽象为 `TemporalEntityPage`  
**状态**: ✅ 已完成（2025-11-10），详见 `frontend/src/features/temporal/pages/TemporalEntityPage.tsx`、`frontend/src/features/positions/PositionDetailView.tsx`

## 背景
- `OrganizationTemporalPage` 直接绑定 `TemporalMasterDetailView`（frontend/src/features/organizations/OrganizationTemporalPage.tsx）  
- `PositionTemporalPage` 自建一套 Tab/状态链路（frontend/src/features/positions/PositionTemporalPage.tsx）  
- 缺乏统一入口导致后续命名抽象难以推进

## 工作内容
1. **TemporalEntityRouteConfig**：封装实体校验（组织 7 位数字、职位 `P\d{7}`/`NEW`）、创建模式判定、mock banner 控制、版本导出/表单可用性等策略。  
2. **功能映射表**：列出现有组织/职位页面的专属能力（如返回列表、CSV 导出、版本抽屉、mock banner、错误提示），明确在新 `TemporalEntityPage` 中对应的 Slot/适配器注入方式；映射表需提交至 `reports/plan242/naming-inventory.md#temporal-entity-page` 并作为 Plan 242 验收输入。  
3. **TemporalEntityPage 组件**：提供 Shell + Slot（Header、Sidebar、Tabs、ActionBar、Drawer）以承载组织/职位逻辑，保证创建模式、编码错误、返回导航等行为等同。  
4. **路由更新**：统一 `react-router` 配置，`/organizations/:code/temporal` 与 `/positions/:code` 通过配置化入口挂载；确保 `Internal/organization` 目录结构符合 Plan 219 要求。  
5. **测试与 Storybook**：为组织/职位各自输出基线（创建/查看模式）录屏与截图，并更新 E2E 样例。  
6. **文档/协调**：在 Plan 242 主文档、`215-phase2-execution-log.md`、Plan 219 README 记录决策与影响，维护唯一事实来源。

## 里程碑 & 验收
- Day 2：提交 MR（组件 + 路由 + 单测 + 功能映射表）  
- Day 3：Storybook、E2E 样例通过，组织/职位创建/查看/错误路径验证完毕  
- 验收标准：
  - 仓库不再存在 `OrganizationTemporalPage`/`PositionTemporalPage` 命名  
  - 创建模式（组织 `new`、职位 `NEW`）校验、返回导航、mock banner、版本导出等行为与现有实现一致  
  - `temporalEntity` 前缀覆盖全部 testid，并通过 Plan 219 目录/接口审查

## 汇报
- 交付后在 `215-phase2-execution-log.md`、Plan 219 README、Plan 242 主文档及 `reports/plan242/naming-inventory.md#temporal-entity-page` 更新（已完成）。若 T2/T3 进度阻塞（Timeline/类型尚未切换），需在汇报中注明临时兼容策略（例如 Slot 降级）并保持唯一事实来源。

## 完成情况摘要
- `TemporalEntityPage` + `entityRoutes` 统一组织/职位路由、无效提示与导航；`PositionDetailView` 承载原 `PositionTemporalPage` 功能（创建/查看/版本抽屉/Mock Banner）。
- 相关文档（README、Implementation Inventory、Plan 06、Plan 215、Plan 242）均已记录改动，功能映射表存放于 `reports/plan242/naming-inventory.md#temporal-entity-page`。
- 单测：`npm run test -- PositionDetailView`、`npm run test -- OrganizationTemporalPage`；E2E 基线待 Plan 242 T4/T5 继续推进。
