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

## 事件词汇表与 Schema
- 统一前缀与命名
  - 事件名前缀：`position.`；日志前缀：`[OBS]`（便于 Playwright 过滤）
  - 正式事件：
    - `position.hydrate.start` / `position.hydrate.done`
    - `position.tab.change`
    - `position.version.select`
    - `position.version.export.start` / `position.version.export.done` / `position.version.export.error`
    - `position.graphql.error`
  - 向上兼容别名（对齐 Plan 240 原始措辞；默认开启）
    - `PositionPageHydration` ≈ `position.hydrate.done`
    - `PositionTabSwitch` ≈ `position.tab.change`
- 负载字段（JSON；不得包含 PII/令牌/响应体）
  - `entity`: 固定 `'position'`
  - `code`: 职位编码（可选）
  - `recordId`: 当前版本记录 ID（可选）
  - `tabFrom`/`tabTo`: 页签切换来源/目标（tab.change）
  - `versionKey`: 版本行 key（version.select）
  - `durationMs`: 时长（hydrate.done/export.done，来自 performance.measure）
  - `sizeBytes`: 导出字节数（export.done，来自 Blob.size）
  - `queryName`: GraphQL 操作名（graphql.error）
  - `status`: HTTP 状态（graphql.error；数值）
  - `ts`: ISO 时间戳
  - `source`: 固定 `'ui'`
- 日志输出（统一通过 logger）
  - DEV：`logger.info('[OBS] position.hydrate.done', payload)` 等
  - CI：`logger.mutation('[OBS] position.hydrate.done', payload)`（详见“开关与环境策略”）
  - 别名（默认开启，payload 同步）：  
    - `logger.info('[OBS] PositionPageHydration', payload)` / `logger.mutation('[OBS] PositionPageHydration', payload)`  
    - `logger.info('[OBS] PositionTabSwitch', payload)` / `logger.mutation('[OBS] PositionTabSwitch', payload)`
- Performance 标记
  - `obs:position:hydrate:start` / `obs:position:hydrate:end` → `obs:position:hydrate:duration`
  - `obs:position:export:start` / `obs:position:export:end` → `obs:position:export:duration`

示例：
```
[OBS] position.hydrate.done {"entity":"position","code":"P9000001","durationMs":842,"ts":"2025-11-15T10:20:30.123Z","source":"ui"}
```

## 开关与环境策略
- 开启控制
  - `VITE_OBS_ENABLED`：DEV 默认 `true`；CI 建议设置 `true`；生产默认 `false`
  - `VITE_OBS_ALIAS_ENABLED`：默认 `true`；设置为 `false` 时仅输出正式事件
- 与现有 logger 对齐（确保 CI 可见）
  - DEV 环境：使用 `logger.info('[OBS] ...')`
  - CI 环境：使用 `logger.mutation('[OBS] ...')` 并设置 `VITE_ENABLE_MUTATION_LOGS='true'` 直通门控
  - 生产：关闭 OBS 信息级日志，仅保留错误上报（error）

## 注入点与代码映射（不改契约）
- 首屏 hydrate
  - 起点：组件首渲染 `useEffect([])` → `performance.mark('obs:position:hydrate:start')`
  - 终点：详情数据就绪且关键 DOM 可见（timeline/overview） → `performance.mark('obs:position:hydrate:end')` → `performance.measure` → 输出 `position.hydrate.done`（并清理 marks/measures）
  - 参考：`frontend/src/features/positions/PositionDetailView.tsx:103-149`
- 页签切换
  - `TabsNavigation` 的 `onTabChange` 输出 `position.tab.change`（from/to、ts）
  - 参考：`frontend/src/features/positions/PositionDetailView.tsx:493`（定义），`:511`、`:524`（调用）
- 版本选择
  - `handleVersionRowSelect` 输出 `position.version.select`（versionKey、ts）
  - 参考：`frontend/src/features/positions/PositionDetailView.tsx:224`
- 导出
  - 开始：`position.version.export.start`；完成：`export.done`（`durationMs`、`sizeBytes`）；失败：`export.error`
  - 参考：成功路径 `:201`，失败日志 `:215`
- GraphQL 错误
  - 职位详情 GraphQL 查询失败或恢复流程进入错误分支时输出 `position.graphql.error`（`queryName`、`status`、`ts`）
  - 位置：职位详情查询失败分支或统一 error handling（`frontend/src/shared/api/error-handling.ts`）
- 去重策略（StrictMode/双渲染）
  - 为 `hydrate.*` 与 `tab.change` 引入一次性标记（ref），单次生命周期仅上报一次

## Playwright 采集与落盘（可执行）
- 采集方式
  - 在用例中监听 `page.on('console', ...)`，筛选 `msg.text().startsWith('[OBS] ')` 行，解析 JSON 负载并累计
- 落盘路径
  - 主：`logs/plan240/D/obs-{spec}-{browser}.log`
  - 兼容汇总（可选）：`logs/ui/position-page.log`（流水式追加，便于历史脚本沿用）
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
   - 在 PositionDetailView/TabsNavigation/导出回调注入 `performance.mark/measure` 与 OBS 事件输出
   - 开关读取：`import.meta.env.VITE_OBS_ENABLED === 'true' || import.meta.env.DEV === true`
   - 别名输出：在 `hydrate.done` 与 `tab.change` 同步输出 `PositionPageHydration`、`PositionTabSwitch`（受 `VITE_OBS_ALIAS_ENABLED` 控制）
   - GraphQL 错误：在职位详情查询错误路径或统一 error handling 中增加 `position.graphql.error`
   - CI 通道：在 CI 使用 `logger.mutation('[OBS] ...')`，并要求设置 `VITE_ENABLE_MUTATION_LOGS=true`
   - 去重实现：`hydrate`/`tab.change` 加一次性标记；导出场景清理 marks/measures
2) Playwright 用例
   - 新增/扩充职位详情用例，监听 console，断言 hydrate/tab/version/export/graphql.error 事件；将事件写入 `logs/plan240/D/*.log`
   - 可选：输出一份 `logs/ui/position-page.log` 汇总（别名兼容）
3) 文档与 runbook
   - 将“事件词汇表与 Schema”并入 `docs/reference/temporal-entity-experience-guide.md`（仅引用）
4) Makefile（建议）
   - 增加 `make e2e-240d`：设置 `PW_OBS=1 VITE_OBS_ENABLED=true [VITE_ENABLE_MUTATION_LOGS=true]` 执行关键 E2E 并落盘到 `logs/plan240/D/`

## 验收标准（量化）
- 事件可见性
  - hydrate：`position.hydrate.done` ≥ 1，含 `code`、`durationMs`、`ts`、`source='ui'`
  - tab/version/export：按用例步骤产生对应事件且字段齐全
  - 别名：当别名开关开启时，`PositionPageHydration`、`PositionTabSwitch` 必须出现
  - 错误路径：注入 GraphQL 故障时出现 `position.graphql.error`，含 `queryName` 与 `status`
- 噪声约束
  - 非 `[OBS]` 控制台日志在 CI 轮次中不超过 20 条（信息级）且无 error；生产构建下不输出 `[OBS]` 信息级日志
- 指标阈值（CI）
  - 首轮仅采集；待 `reports/plan240/baseline/obs-summary.json` 建立后启用阈值：`hydrate.done` median ≤ 3000ms（或 p95 ≤ 5000ms）
- 产物
  - 主：`logs/plan240/D/obs-*.log`（按用例/浏览器落盘）  
  - 兼容：`logs/ui/position-page.log`（可选）  
  - （可选）聚合报告 `reports/plan240/baseline/obs-summary.json`

## 风险与回滚
| 风险 | 影响 | 缓解/回滚 |
| --- | --- | --- |
| 事件过量导致噪声 | E2E 日志污染、难排查 | 严格门控（开关+生产关闭），事件采样与字段最小化 |
| 注入时机错误导致测量偏差 | 指标失真 | 以“数据就绪 + 关键 DOM 可见”为准；代码评审列出注入位置 |
| JSON 解析失败 | 用例断言不稳 | 强制统一输出格式：前缀 + 单行 JSON |

## 证据与落盘
- 日志：`logs/plan240/D/*.log`（Playwright 监听 console 写入）  
- 报告：`reports/plan240/baseline/obs-summary.json`（可选，记录中位数/分位数与失败比）  
- 参考：`frontend/tests/e2e/optimization-verification-e2e.spec.ts:66`（console 监听模式），`frontend/src/shared/utils/logger.ts`（门控/通道策略）
