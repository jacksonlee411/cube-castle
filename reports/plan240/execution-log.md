# Plan 240 – 执行登记与观测证据（240D）

更新日期: 2025-11-15  
关联计划: 240D – 职位详情可观测性与指标注入  
状态: 已完成（验收通过）

---

## 1. 执行摘要
- 观测发射器与注入点按“彻底方案”落地（无运行时别名、无重复定义）  
- 事件统一前缀与 Schema 引用 `docs/reference/temporal-entity-experience-guide.md`（唯一事实来源）  
- 运行门控：`VITE_OBS_ENABLED`（功能） + `VITE_ENABLE_MUTATION_LOGS`（CI 通道）；生产不输出信息级 `[OBS]`

## 2. 代码映射
- 发射器：`frontend/src/shared/observability/obs.ts`  
- 职位详情注入：`frontend/src/features/positions/PositionDetailView.tsx`  
  - Hydration：`position.hydrate.start/.done`（含 `durationMs`）  
  - Tabs：`position.tab.change`  
  - 版本选择：`position.version.select`  
  - 导出：`position.version.export.start/.done/.error`（含 `durationMs/sizeBytes`）  
- GraphQL 错误：`frontend/src/shared/hooks/useEnterprisePositions.ts` 失败路径发 `position.graphql.error`

## 3. 用例与产物
- 用例：`frontend/tests/e2e/position-observability.spec.ts`  
- 模式：
  - `PW_POSITION_CODE=<现有职位>`：跳过创建，仅验证 hydrate/tab（尽力尝试版本/导出）  
  - 不设置 `PW_POSITION_CODE`：自动创建职位与版本，强制断言 version.select 与 export.*  
- 报告：`frontend/playwright-report/index.html`  
- 证据目录（唯一）：`logs/plan240/D/`  
  - 样例：`logs/plan240/D/obs-position-observability-chromium.log`

## 4. 样例事件
```
[OBS] position.hydrate.start {"entity":"position","code":"P9000001","ts":"...","source":"ui"}
[OBS] position.hydrate.done {"entity":"position","code":"P9000001","durationMs":370,"ts":"...","source":"ui"}
[OBS] position.tab.change {"tabFrom":"overview","tabTo":"timeline","entity":"position","code":"P9000001","ts":"..."}
```

## 5. 结论
- 240D 完成并登记：事件可见性满足、目录统一、门控一致、无第二事实来源  
- 后续（非阻塞）：基线聚合 `reports/plan240/baseline/obs-summary.json` 可在下一迭代引入

