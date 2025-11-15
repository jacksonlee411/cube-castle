# 时态实体多页签详情设计规范（Temporal Entity Experience Guide）

版本: v1.0  
更新时间: 2025-11-14  
适用范围: 适用于“时态实体（Temporal Entity）”详情体验，包括组织、职位及后续扩展实体的“列表 → 多页签详情”流程

---

## 1. 信息架构

1) 入口与路由  
- 组织：`/organizations/:code/temporal`（7 位数字编码或 `new`）  
- 职位：`/positions/:code`（`P\d{7}` 或 `new`）  
- 统一 shell：由 `TemporalEntityPage`（前端）承载路由校验与无效编码提示

2) 页面结构（统一骨架）  
- 左侧：版本导航（含时间轴与版本列表），桌面端默认 320px  
- 右侧：六个页签（顺序固定）  
  - 概览 → 任职记录 → 调动记录 → 时间线 → 版本历史 → 审计历史  
- 顶部工具栏（右侧）：返回、编辑、创建版本、更多操作（按权限/Mock 模式控制可见）

3) 窄屏表现  
- 宽度 < 960px：左侧版本导航折叠为抽屉；页签区支持横向滚动

---

## 2. 交互准则

| 区域 | 交互说明 |
|------|----------|
| 版本导航 | 点击节点切换 `selectedVersion`，时间轴与版本列表同步高亮；抽屉模式下选择后自动关闭 |
| 页签栏 | 使用 Canvas `Flex` + 底边高亮；支持键盘导航（左右切换、Enter 激活） |
| Mock 模式 | 顶部显示橙色 Banner，禁用“创建/编辑/新增版本”按钮；所有关键元素需有稳定 `data-testid` |
| 审计页签 | 如缺少 `recordId`，显示提示卡片并避免无效请求 |
| 空态 | 任职/调动/版本列表统一浅灰文案：`暂无 XXX 记录` |

---

## 3. 视觉与可访问性

1) 配色与 Token  
- 统一沿用 Canvas token；Banner 使用 `cinnamon100`/`cinnamon600`；选中行 `soap200`

2) 状态标签  
- 调用统一元数据：`TEMPORAL_ENTITY_STATUS_META`  
- “计划版本”标注为“计划”，“当前版本”标注为“当前”

3) 间距与响应式  
- 页内垂直间距建议 `24px`（`SimpleStack`）  
- 左侧卡片与右侧主体留 `space.l`  
- 1280px/960–1279px/<960px 三档布局

4) 可访问性（A11y）  
- 交互元素使用原生 `button`/`a` 或 `role=button` + 键盘可达  
- 版本行应有 `aria-selected`，与时间轴同步  
- Tab 导航支持左右键切换并具备可见焦点态  
- Mock 提示提供文字说明与解法，不仅依赖颜色

---

## 4. 技术映射与命名治理

1) 路由与页面  
- `TemporalEntityPage` + `TemporalEntityRouteConfig`（统一路由校验与错误提示）  
- 实体适配：`TemporalEntityPage.Organization` / `.Position` 注入文案与操作策略

2) 时间线与状态  
- 时间线适配器：`frontend/src/features/temporal/entity/timelineAdapter.ts`  
- 状态元数据：`frontend/src/features/temporal/entity/statusMeta.ts`（`TEMPORAL_ENTITY_STATUS_META`）

3) 统一类型与 Hook  
- 类型：`frontend/src/shared/types/temporal-entity.ts`（`TemporalEntityRecord` 等）  
- Hook：`useTemporalEntityDetail` + `createTemporalDetailLoader`（由实体薄封装复用）

4) 选择器与测试
- 选择器集中：`frontend/src/shared/testids/temporalEntity.ts`（导出 `temporalEntitySelectors`；禁止在组件/测试中硬编码 `data-testid`，统一从此处导入）  
- E2E 用例仅使用中性 `temporalEntity-*` 前缀的 `data-testid`

---

## 5. 契约与一致性

- 查询统一 GraphQL，命令统一 REST，单一数据源 PostgreSQL（CQRS）  
- 对外字段命名 camelCase，路径参数统一 `{code}`  
- 增量扩展字段须先更新 `docs/api/openapi.yaml` / `docs/api/schema.graphql` 并通过实现清单生成器校验  
- 禁止在本文档复制“实现细节”；仅提供权威入口与不变约束，易变实现以生成器快照与计划日志为准

---

## 6. 资产与参考

| 文件/路径 | 用途 |
|-----------|------|
| `frontend/artifacts/layout/*.png` | 视觉参考、布局截图 |
| `frontend/src/features/temporal/*` | 组件骨架、适配器与元数据 |
| `frontend/src/shared/types/temporal-entity.ts` | 统一类型导出 |
| `frontend/src/shared/testids/temporalEntity.ts` | 统一 E2E 选择器 |
| `docs/api/*` | OpenAPI/GraphQL 契约 |
| `reports/plan242/naming-inventory.md` | 命名与入口盘点 |
| `logs/plan242/t2|t3|t5/*` | 执行记录与校验日志 |

---

## 7. 可观测性与指标（Observability & Metrics）

本节定义时态实体详情页面的观测事件与输出约束，作为唯一事实来源。职位（position）详情作为首个落地实体；其它实体复用同一模式（事件前缀替换为实体名）。

7.1 事件词汇表（职位详情）
- 正式事件（前缀 `[OBS]` + 事件名 + JSON 负载）  
  `position.hydrate.start` / `position.hydrate.done`  
  `position.tab.change`  
  `position.version.select`  
  `position.version.export.start` / `.done` / `.error`  
  `position.graphql.error`

7.2 负载 Schema（JSON；不得包含 PII/令牌/响应体）
- 通用字段：`entity`（固定 `'position'`）、`code`（职位编码，可选）、`ts`（ISO 时间）、`source='ui'`
- 事件特定字段：  
  hydrate.done：`durationMs`（由 performance.measure 计算）  
  tab.change：`tabFrom`、`tabTo`  
  version.select：`versionKey`  
  export.done：`durationMs`、`sizeBytes`（由 Blob.size）  
  graphql.error：`queryName`、`status`（数值；`statusText` 可选且需脱敏）

7.3 输出通道与环境门控
- DEV：使用 `logger.info('[OBS] <event>', payload)` 输出  
- CI：使用 `logger.mutation('[OBS] <event>', payload)` 输出；设置 `VITE_OBS_ENABLED=true VITE_ENABLE_MUTATION_LOGS=true`  
- 生产：默认关闭信息级 OBS 日志，仅保留错误（error）  
  （不提供别名事件，统一使用正式事件名）

7.4 Performance 标记（建议）
- `obs:position:hydrate:start` / `obs:position:hydrate:end` → `obs:position:hydrate:duration`  
- `obs:position:export:start` / `obs:position:export:end` → `obs:position:export:duration`  
- 严格模式/双渲染：对 `hydrate.*`、`tab.change` 采用一次性标记避免重复

7.5 采集与落盘（E2E/CI）
- Playwright 监听 console：筛选 `msg.text().startsWith('[OBS] ')`，解析 JSON  
- 产物路径（职位）：主 `logs/plan240/D/obs-{spec}-{browser}.log`；兼容 `logs/ui/position-page.log`（可选汇总）  
- 运行示例：  
  本地：`PW_OBS=1 VITE_OBS_ENABLED=true npx playwright test`  
  CI：`PW_OBS=1 VITE_OBS_ENABLED=true VITE_ENABLE_MUTATION_LOGS=true npx playwright test`

示例日志行：
```
[OBS] position.hydrate.done {"entity":"position","code":"P9000001","durationMs":842,"ts":"2025-11-15T10:20:30.123Z","source":"ui"}
```

约束说明：本规范仅定义稳定的事件命名、字段与门控策略；实现细节（注入位置、测量方式）由相应计划与 MR 描述；严禁在本指南复制实现代码片段。

---

维护者：前端/设计/QA 联合小组  
反馈渠道：在 Plan 06 的“设计与命名规范”条目下留言，或在相关 MR 发起评审

---

## 附录 A – 框架与工程实践清单（索引）

说明：本附录仅充当“索引与核对清单”，不复制实现细节；若需变更，请先更新对应计划/契约文档，再回填引用，确保单一事实来源。

- 跨层硬约束（契约/CQRS/容器化）
  - 契约先行：增量字段/操作须先更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql`，并运行实现清单脚本校验（参考 240B 前置，docs/development-plans/240B-position-loading-governance.md:30）
  - CQRS：命令=REST、查询=GraphQL、单一数据源 PostgreSQL（参考 241 对齐，docs/development-plans/241-frontend-framework-refactor.md:19）
  - Docker 强制：本地/CI 以 `make docker-up` → `make run-dev` → `make frontend-dev` 为前置（参考 240B 前置，docs/development-plans/240B-position-loading-governance.md:23）

- 后端演进与可靠通信（供前端协作约束）
  - 模块化单体 + DDD：模块按 Bounded Context 划分，避免技术/表结构驱动（docs/development-plans/203-hrms-module-division-plan.md:18）
  - 同步依赖注入、异步“事务性发件箱”强制（docs/development-plans/203-hrms-module-division-plan.md:333）
  - 权限策略外部化并与 OpenAPI scope 对齐（docs/development-plans/203-hrms-module-division-plan.md:592）

- 前端框架与数据读取（241/240B）
  - 统一骨架：以 `TemporalEntityLayout`（Shell/Sidebar/Tabs）承载组织/职位等（docs/development-plans/241A-temporal-entity-layout-integration.md:20）
  - 统一 Hook/Loader：`useTemporalEntityDetail` + 内部 Loader 工厂，页面仅读缓存；预热/并发/取消/重试/失效在 Loader 管理（docs/development-plans/241-frontend-framework-refactor.md:40，docs/development-plans/240B-position-loading-governance.md:37）
  - 幂等读韧性：仅对幂等读启用指数退避重试；路由/页签切换触发 Abort 取消；旧响应丢弃；重试/延迟配置集中导出（docs/development-plans/240B-position-loading-governance.md:46）
  - QueryKey/失效 SSoT：标准化键维度与统一失效工具，禁止散落键名（docs/development-plans/240B-position-loading-governance.md:60）

- 命名与测试资产（242/246/240C）
  - Selector SSoT：仅从 `frontend/src/shared/testids/temporalEntity.ts` 导入；禁硬编码 `data-testid`；ESLint + Guard 双门禁（docs/development-plans/241B-unified-hook-and-selector-guard.md:24，docs/development-plans/246-temporal-entity-selectors-fixtures-plan.md:59）
  - 旧前缀冻结与渐进收敛：`organization-*`/`position-*` 新增计数禁止上升，迁移到 `temporal-*`（docs/development-plans/240C-position-selectors-unification.md:9）
  - 时间线/状态命名抽象：统一 `TemporalEntity*` 适配器与元数据（docs/development-plans/244-temporal-timeline-status-plan.md:18）

- 可观测性与证据（240D/241C）
  - OBS 事件与 performance.mark 在骨架与关键交互注入，按环境门控输出；E2E 采集并落盘（docs/development-plans/240D-position-observability.md:25，docs/development-plans/241C-e2e-acceptance-and-observability-evidence.md:27）
  - 证据路径唯一：E2E 控制台/trace/HAR 由测试侧采集，落盘到计划指定目录，禁止运行时代码多处写文件（docs/development-plans/240D-position-observability.md:55）

- 回滚与临时方案（240A/240B/240E）
  - Feature Flag 回退：布局/Loader 等关键变更提供特性开关与冒烟回滚流程（docs/development-plans/240A-position-layout-alignment.md:66，docs/development-plans/240B-position-loading-governance.md:152）
  - 临时方案治理：仅允许最薄适配层，并以 `// TODO-TEMPORARY(YYYY-MM-DD)` 标注与限期清零（docs/development-plans/240A-position-layout-alignment.md:18）

核对用法：在评审或合并前，按上述条目比对对应计划与代码引用是否到位；若发现不一致或第二事实来源，先回滚并对齐计划/契约，再推进实现。
