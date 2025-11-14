# Plan 240C – 职位 DOM/TestId 治理与选择器统一

编号: 240C  
上游: Plan 240（职位管理页面重构） · 依赖 240A 完成  
状态: 待启动

—

## 目标
- 将职位相关 `data-testid` 与选择器统一到 `temporalEntity-*` 前缀；集中导出选择器，稳定 Playwright。

## 范围
- 影响面：职位详情、列表与相关组件；Playwright 用例选择器替换。  
- 守卫：与 Plan 245 Guard 协同，禁止新增旧前缀。

## 任务清单
1) 选择器集中：`frontend/src/shared/testids/temporalEntity.ts` 导出 position 相关选择器。  
2) 用例替换：批量将用例引用切换到集中选择器；旧选择器在一个迭代窗口内保留 fallback。  
3) 守卫：配置 `scripts/quality/selector-guard-246.js`（如已有则扩充），计数不升高即通过。

## 验收标准
- Playwright 用例（职位相关）全部改用统一选择器并通过；旧前缀计数不升高。

## 证据与落盘
- 日志：`logs/plan240/C/*.log`；报告：`reports/plan240/baseline/` 更新对比。

