# Plan 240D – 职位详情可观测性与指标注入

编号: 240D  
上游: Plan 240（职位管理页面重构） · 依赖 240A/240B 完成  
状态: 计划就绪（待实施）

—

## 目标
- 在职位详情骨架注入关键 `performance.mark/measure` 与结构化日志事件，支撑 E2E/CI 的时序与行为断言；产出可落盘的观测证据，且不引入第二事实来源。

## 范围
- 事件：首屏渲染（hydrate）、Tab 切换、版本切换、导出（开始/完成/失败）、职位详情 GraphQL 错误；统一通过 logger 管线输出。
- 开关：默认 DEV 开启；CI 可通过 env 开关强制开启；生产默认关闭（仅保留 error）。

## 事件词汇表与 Schema（单一事实来源）
- 本计划不再重复定义事件列表与字段，统一引用：`docs/reference/temporal-entity-experience-guide.md:104` 起的“7. 可观测性与指标”。  
- 说明：若事件命名/字段/门控有变更，必须先更新上述参考规范，再在本计划中“引用”该变更，避免出现第二事实来源。

示例：
```
[OBS] position.hydrate.done {"entity":"position","code":"P9000001","durationMs":842,"ts":"2025-11-15T10:20:30.123Z","source":"ui"}
```

## 开关与环境策略
- 开启控制（简化为单一功能门控）
  - `VITE_OBS_ENABLED`：DEV 默认开启；CI 建议设置为 `true`；生产默认 `false`（仅保留 error）。
- 输出通道（复用既有 logger 规则）
  - DEV：`logger.info('[OBS] ...')`
  - CI：`logger.mutation('[OBS] ...')`（需设置 `VITE_ENABLE_MUTATION_LOGS='true'`）
  - 生产：不输出信息级 `[OBS]` 日志，仅保留 `logger.error`。

## 注入点与代码映射（不改契约）
- 首屏 hydrate（PositionDetailView）
  - 起点：组件首渲染的 `useEffect([])` → `performance.mark('obs:position:hydrate:start')`
  - 终点：详情数据就绪且关键 DOM（概览或时间线）可见 → `performance.mark('obs:position:hydrate:end')` → `performance.measure` → 输出 `position.hydrate.done` → 清理 marks/measures
- 页签切换（TabsNavigation）
  - 在 `onTabChange` 中输出 `position.tab.change`（from/to、ts）
- 版本选择（版本列表/时间线选中回调）
  - 在版本行/时间线点击回调中输出 `position.version.select`（`versionKey`、ts）
- 导出（版本导出回调）
  - `position.version.export.start` → 成功 `position.version.export.done`（含 `durationMs`、`sizeBytes`）或失败 `position.version.export.error`
- GraphQL 错误（统一拦截）
  - 在 GraphQL 统一错误处理处（如 `frontend/src/shared/api/unified-client.ts` 或 `frontend/src/shared/api/error-handling.ts`）输出 `position.graphql.error`。  
  - `queryName` 建议通过 `operationName` 传入；若短期无法提供，则 v1 允许为 `'unknown'`，并以 `// TODO-TEMPORARY:` 标注补齐计划。
- 去重策略（StrictMode/双渲染）
  - 对 `hydrate.*` 与 `tab.change` 使用一次性标记（`ref`）确保单次生命周期仅上报一次。
- 实现去重复的轻量封装（建议）
  - 在前端新增极薄工具 `frontend/src/shared/observability/obs.ts`，提供 `emitObs(event, payload)`，内部统一 `[OBS]` 前缀、门控与去重，避免在多个组件手写重复逻辑（非新框架，仅最小工具）。

## Playwright 采集与落盘（可执行）
- 采集方式
  - 在用例中监听 `page.on('console', ...)`，筛选 `msg.text().startsWith('[OBS] ')` 行，解析 JSON 负载并累计。
- 落盘路径
  - 统一：`logs/plan240/D/obs-{spec}-{browser}.log`（唯一事实来源）。  
  - 不再写入 `logs/ui/position-page.log`，避免证据双写。
- 示例命令
  - 本地（DEV）：`PW_OBS=1 VITE_OBS_ENABLED=true npx playwright test tests/e2e/position-tabs.spec.ts`
  - CI：`PW_OBS=1 VITE_OBS_ENABLED=true VITE_ENABLE_MUTATION_LOGS=true npx playwright test`
- 断言建议
  - hydrate.done 至少 1 条，且含 `durationMs`、`code`
  - 导出：`export.start` 与 `export.done` 成对；或 `export.error` 出现（互斥）
  - tab.change：from/to 不相等
  - graphql.error：注入故障时至少 1 条，含 `queryName` 与数值 `status`

## 任务清单
1) 注入实现
   - 在 PositionDetailView/TabsNavigation/导出回调注入 `performance.mark/measure` 与 OBS 事件输出。
   - 开关读取：`import.meta.env.VITE_OBS_ENABLED === 'true' || import.meta.env.DEV === true`。
   - 统一发射：新增 `emitObs` 极薄封装以去除重复代码（可选，但推荐）。
   - GraphQL 错误：在统一错误处理处输出 `position.graphql.error`；`queryName` v1 允许 `'unknown'`，并以 `// TODO-TEMPORARY:` 标注补齐计划。
   - CI 通道：在 CI 使用 `logger.mutation('[OBS] ...')`，并设置 `VITE_ENABLE_MUTATION_LOGS=true`。
   - 去重实现：`hydrate`/`tab.change` 一次性标记；导出场景清理 marks/measures。
2) Playwright 用例
   - 新增/扩充职位详情用例，监听 console，断言 hydrate/tab/version/export/graphql.error 事件；将事件写入 `logs/plan240/D/*.log`。
3) 文档与 runbook
   - “事件词汇表与 Schema”仅引用 `docs/reference/temporal-entity-experience-guide.md`；本计划不再复制。
4) Make（可选）
   - 保持通用 `e2e` 命令 + 环境变量；如需便捷脚本，新增通用 `make e2e-obs`（参数化计划编号），避免为单计划新增专用目标。

## 验收标准（量化）
- 事件可见性
  - hydrate：`position.hydrate.done` ≥ 1，含 `code`、`durationMs`、`ts`、`source='ui'`
  - tab/version/export：按用例步骤产生对应事件且字段齐全
  - 错误路径：注入 GraphQL 故障时出现 `position.graphql.error`，含 `queryName` 与 `status`
- 噪声约束
  - 非 `[OBS]` 控制台日志在 CI 轮次中不超过 20 条（信息级）且无 error；生产构建下不输出 `[OBS]` 信息级日志
- 指标阈值（CI）
  - 首轮仅采集；聚合与阈值将在后续 MR 中引入（建立 `reports/plan240/baseline/obs-summary.json` 后再启用，例如 `hydrate.done` median ≤ 3000ms 或 p95 ≤ 5000ms）
- 产物
  - `logs/plan240/D/obs-*.log`（按用例/浏览器落盘，唯一）  
  - （后续）聚合报告 `reports/plan240/baseline/obs-summary.json`（待脚本落地）

## 风险与回滚
| 风险 | 影响 | 缓解/回滚 |
| --- | --- | --- |
| 事件过量导致噪声 | E2E 日志污染、难排查 | 严格门控（开关+生产关闭），事件采样与字段最小化 |
| 注入时机错误导致测量偏差 | 指标失真 | 以“数据就绪 + 关键 DOM 可见”为准；代码评审列出注入位置 |
| JSON 解析失败 | 用例断言不稳 | 强制统一输出格式：前缀 + 单行 JSON |

## 证据与落盘
- 日志：`logs/plan240/D/*.log`（Playwright 监听 console 写入，唯一）  
- 报告：`reports/plan240/baseline/obs-summary.json`（后续 MR 引入，记录分位数与失败比）  
- 参考：`frontend/tests/e2e/optimization-verification-e2e.spec.ts:66`（console 监听模式），`frontend/src/shared/utils/logger.ts`（门控/通道策略），`docs/reference/temporal-entity-experience-guide.md:104`
