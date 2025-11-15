# Plan 240C – 职位 DOM/TestId 治理与选择器统一

编号: 240C  
上游: Plan 240（职位管理页面重构） · 依赖 240A 完成  
状态: 进行中（集中选择器已落地，单测与部分组件待收敛）

—

## 目标
- 统一通过 `temporalEntitySelectors` 暴露职位相关选择器（SSoT），组件与测试不得直接硬编码 testid 字符串；前缀采用 `temporal-*` 族。

## 范围
- 覆盖：职位详情、列表、版本区块、空缺看板、编制仪表板；Vitest 与 Playwright 用例同步迁移。  
- 守卫：与 Plan 246 Guard 协同，冻结 `position-*/organization-*` 旧前缀新增。

## 任务清单
1) 选择器集中完善  
   - 在 `frontend/src/shared/testids/temporalEntity.ts` 增补职位域缺失条目：  
     - `vacancyBoard`、`headcountDashboard`、`versionRow(key)`、`versionRowPrefix` 等。  
   - 将 `VersionList` 行、`VacancyBoard`、`HeadcountDashboard` 等组件的 `data-testid` 切换为集中选择器。  
2) 单测迁移（Vitest）  
   - 将 `frontend/src/features/positions/__tests__` 下的断言由硬编码 `position-*` 替换为 `temporalEntitySelectors.position.*`。  
3) E2E 迁移（Playwright）  
   - 统一改为使用集中选择器；避免直接硬编码旧前缀。  
4) 守卫  
   - 使用 `scripts/quality/selector-guard-246.js` 冻结旧前缀；`npm run guard:selectors-246` 通过（基线不升高）。  
   - 后续增强（建议）：补充 ESLint 规则/脚本，限制直接硬编码 `temporal-*`，必须经由 `temporalEntitySelectors` 引用。

## 验收标准
- 代码侧：职位域组件不再出现 `data-testid="position-*"`；关键组件改为 `temporalEntitySelectors`。  
- 测试侧：职位域 Vitest/Playwright 用例均通过，且用例中 `position-*` 旧前缀计数较基线不升高（优先下降）。  
- 守卫：`npm run guard:selectors-246` 通过。

## 证据与落盘
- 日志：`logs/plan240/C/selector-guard.log`、`logs/plan240/C/unit-and-e2e.log`  
- 守卫基线：`reports/plan246/baseline.json`（Plan 246 产物，作为冻结基线使用）
