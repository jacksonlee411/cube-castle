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
- 选择器集中：`frontend/src/shared/testing/temporalSelectors.ts`  
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
| `frontend/src/shared/testing/temporalSelectors.ts` | 统一 E2E 选择器 |
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
- 向上兼容的别名（与 Plan 240 文案对齐；负载相同）  
  `PositionPageHydration` ≈ `position.hydrate.done`  
  `PositionTabSwitch` ≈ `position.tab.change`

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
- 别名输出开关：`VITE_OBS_ALIAS_ENABLED`（默认 `'true'`）；为 `false` 时仅输出正式事件

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
