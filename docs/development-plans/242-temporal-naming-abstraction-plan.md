# Plan 242 – 时态实体命名抽象与统一治理

**编号**: 242  
**标题**: Temporal Entity Naming Convergence & Governance  
**创建日期**: 2025-11-10  
**关联计划**: Plan 241（框架重构）、Plan 240（职位页面重构）、Plan 06（集成验证）  
**状态**: 草案 · 待评审  

---

## 1. 背景与动因

1. 组织/职位详情入口仍以实体名硬编码：`OrganizationTemporalPage` 直接绑定 `TemporalMasterDetailView`（frontend/src/features/organizations/OrganizationTemporalPage.tsx:11-78），`PositionTemporalPage` 则自建一套 Tab/状态/表单（frontend/src/features/positions/PositionTemporalPage.tsx:1-146）。命名缺乏统一抽象，阻碍 Plan 241 推动的共享框架落地。
2. Timeline/版本工具与状态元数据沿用职位专属命名：`timelineAdapter.ts` 只处理 `PositionRecord`（frontend/src/features/positions/timelineAdapter.ts:1-105），状态配置散落在 `statusMeta.ts` 与 `shared/utils/statusUtils.ts`，前者面向职位、后者面向组织（frontend/src/features/positions/statusMeta.ts:1-44、frontend/src/shared/utils/statusUtils.ts:1-58）。缺乏 `TemporalEntity` 级别的命名会导致未来新增实体时再次复制粘贴。
3. E2E 测试与 fixtures 采用实体专属命名：`position-tabs.spec.ts` 依赖 `position-temporal-page-*` testid（frontend/tests/e2e/position-tabs.spec.ts:91-145），而组织用例 `organization-create.spec.ts` 使用 `organization-*` 前缀（frontend/tests/e2e/organization-create.spec.ts:4-41）。`utils/positionFixtures.ts` 亦与职位强绑定（frontend/tests/e2e/utils/positionFixtures.ts:1-160），阻碍 selector/fixture 共用。

> **结论**：需要在命名层面建立“Temporal Entity” 中性抽象，覆盖页面、组件、Hook、Timeline、状态配置与测试资产，为 Plan 241/240 提供一致的命名基线。

---

## 2. 目标与验收

| 目标 | 说明 | 验收方式 |
| --- | --- | --- |
| A. 页面命名抽象 | 以 `TemporalEntityPage` 为统一入口，组织/职位仅通过配置扩展，不再保留硬编码页面命名 | 新路由/组件命名 PR + Storybook 录屏 |
| B. Timeline & Status 抽象 | 输出 `TemporalEntityTimelineAdapter` 与 `TemporalEntityStatusMeta`，由各实体适配器注入字段映射 | Adapter 单测 + API 对照表 |
| C. Selector & Fixture 统一 | 建立 `temporalEntitySelectors` 与中性 fixtures（如 `temporalEntityFixtures.ts`），E2E 用例仅使用中性命名 | Playwright diff + utils 重用证明 |
| D. 类型与契约命名 | 在 `shared/types` / GraphQL Operation / React Query Key 层统一导出 `TemporalEntityRecord`、`TemporalEntityStatus` 等别名，组织/职位仅作类型别名 | TypeScript diff + GraphQL schema 变更说明 |
| E. 文档与指南 | 更新 `docs/reference/positions-tabbed-experience-guide.md` 为 `temporal-entity-experience-guide`，同步命名映射 | 文档 PR + 评审纪要 |

---

## 3. 范围与任务

### 3.1 T0 – 现状盘点（1 天）
- 运行 `node scripts/generate-implementation-inventory.js` 并对前后端（Go/TS）执行 `rg "Organization.*Temporal|Position.*Temporal"`，覆盖 `cmd/`、`tests/`、`frontend/`、`scripts/`。  
- 输出 `reports/plan242/naming-inventory.md`，列出所有命名分布、引用文件与行号（含 Go/E2E/文档），作为迁移基线。

### 3.2 T1 – 页面与路由命名抽象（1.5 天）
- 将 `OrganizationTemporalPage`、`PositionTemporalPage` 抽象为 `TemporalEntityPage` + `TemporalEntityRouteConfig`，保留实体特有校验（组织 7 位数字、职位 `P\d{7}`）。  
- 更新 `react-router` 配置与懒加载入口，确保 `/organizations/:code/temporal`、`/positions/:code` 通过统一入口创建页面。  
- 引入 `TemporalEntityPage.Organization` / `.Position` 适配器，内部仅传入文案、操作按钮策略，命名全部使用 `temporalEntity-*` 前缀。

### 3.3 T2 – Timeline/Status 命名抽象（4 天）
- 迁移 `frontend/src/features/positions/timelineAdapter.ts` 为 `TemporalEntityTimelineAdapter`，并提供 `createTimelineAdapter({ entity, labelBuilder })` 工厂；组织/职位复用一套类型定义，避免 `unitType = 'POSITION'` 等硬编码。  
- 整合 `frontend/src/features/positions/statusMeta.ts` 与 `frontend/src/shared/utils/statusUtils.ts`，建立 `TemporalEntityStatusMeta`（含 `statusConfig.position`, `statusConfig.organization`），命名、色板、标签统一在一个映射表中。  
- 更新所有引用（Position version list、组织 StatusBadge 等）为新命名，同时通过 lint 规则阻止直接引用旧文件；预留时间处理 Storybook/Vitest 回归。

### 3.4 T3 – Types/GraphQL/Hook 命名抽象（4 天）
- 在 `frontend/src/shared/types` 内新增 `TemporalEntityRecord`, `TemporalEntityTimelineEntry`, `TemporalEntityStatus` 等统一接口，由 `PositionRecord`, `OrganizationUnit` 转为类型别名；所有消费端通过实体适配器映射字段。  
- 统一 GraphQL operation 与 React Query key 命名：例如 `POSITION_DETAIL_QUERY_NAME` → `TEMPORAL_ENTITY_DETAIL_QUERY` + entity 参数；`positionDetailQueryKey` → `temporalEntityDetailQueryKey`。  
- 在本计划内直接交付 `useTemporalEntityDetail` Hook（含 QueryKey、React Query integration、单测），并让 `usePositionDetail`、`useOrganizationDetail` 成为该 Hook 的薄封装，彻底摆脱 Plan 241 依赖。  
- 更新 `docs/api/openapi.yaml`、`docs/api/schema.graphql` 与后端调用说明，使 GraphQL operation 名字同步改为 `TemporalEntity*`，并运行 `node scripts/generate-implementation-inventory.js` 校验唯一事实来源；不保留向后兼容别名。

### 3.5 T4 – Selectors & Fixtures（3 天）
- 新增 `frontend/src/shared/testing/temporalEntitySelectors.ts`，集中维护 `temporal-entity-*` testid；将 `position-tabs.spec.ts`、`organization-create.spec.ts`、`temporal-management-integration.spec.ts` 等 E2E 用例替换为中性 selector（frontend/tests/e2e/position-tabs.spec.ts:91-145；frontend/tests/e2e/organization-create.spec.ts:4-41）。  
- 合并 `frontend/tests/e2e/utils/positionFixtures.ts` 与组织 fixtures，输出 `temporalEntityFixtures.ts`，通过 `entityType` 参数生成 GraphQL 响应，命名遵循 `{entity}Fixture` 而非 `POSITION_*` 常量。  
- 更新 `waitPatterns`/`auth-setup` 等工具函数中的常量名，保证 e2e utils 无实体专属前缀；编写 codemod 和临时 alias，安排双写验证窗口。

### 3.6 T5 – 文档与工具（1 天）
- 将 `docs/reference/positions-tabbed-experience-guide.md` 重命名/改写为 `docs/reference/temporal-entity-experience-guide.md`，同步 Plan 06、Plan 240/241 的引用。  
- 更新 `README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`，对齐 Phase2 文档更新要求（参考 `docs/development-plans/215-phase2-summary-overview.md:250-269`）。  
- 在 `docs/development-plans/06-integrated-teams-progress-log.md` 以及 `215-phase2-execution-log.md` 追加命名迁移记录，确保 Phase2 追踪与唯一事实来源同步。

---

## 4. 里程碑

| 里程碑 | 时间 | 交付物 |
| --- | --- | --- |
| M1 | Day 3 | T0/T1 完成，提交 `TemporalEntityPage` MR |
| M2 | Day 8 | 完成 Timeline/Status 抽象 MR + 单测 |
| M3 | Day 12 | Types/GraphQL/Hook 命名抽象 MR |
| M4 | Day 15 | Selector/Fixture 统一 MR + Playwright 绿灯 |
| M5 | Day 16 | 文档更新 + 命名库存档案（含 README/Quick Reference/Inventory） |

---

## 5. 验收标准

1. 任意页面/组件/Hook/test 中不再直接引用 `OrganizationTemporalPage`、`PositionTemporalPage`、`timelineAdapter`、`statusMeta` 等旧命名；lint/TS 审核确保删除。  
2. TypeScript/GraphQL/React Query 层不再出现 `PositionDetailQuery`, `OrganizationVersionsQuery` 等命名，全部通过 `TemporalEntity*` 命名导出；ESLint/TS 检查确保类型别名唯一。  
3. Playwright 用例仅使用 `temporalEntitySelectors` 暴露的 testid，`position-tabs.spec.ts`、`temporal-management-integration.spec.ts`、`organization-create.spec.ts` 通过 Chromium/Firefox 连续 3 次运行。  
4. `temporalEntityFixtures.ts` 成为唯一事实来源；旧 `positionFixtures.ts` 标记删除或代理导出且附弃用说明。  
5. 文档/计划引用统一使用 `Temporal Entity` 命名，且 Phase2 要求的 README/Quick Reference/Implementation Inventory/215 执行日志均已更新，无平行事实来源。

---

## 6. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
| --- | --- | --- | --- |
| 路由抽象影响历史书签或深链 | 中 | 中 | 提供 Redirect 方案（旧路径 -> 新入口），在 QA 前验证 |
| Timeline Adapter 抽象导致类型发散 | 中 | 低 | 通过泛型约束 + 单测覆盖组织/职位双场景 |
| E2E Selector 替换成本高 | 中 | 中 | 编写 codemod/脚本批量替换 testid，短期保留别名 |
| 文档重命名导致外部链接失效 | 低 | 中 | 在 docs/ 下保留 stub，或更新 README/Plan 06 中的导航 |

---

## 7. 汇报

- 每日于 Plan 06 “阻塞”栏目同步命名迁移进度；Plan 240/241 更新 should mention Plan 242 输出。  
- 产出 `logs/plan242/*.log` 与 `reports/plan242/naming-inventory.md`，供评审/归档。  
- 完成后将本计划归档至 `docs/archive/development-plans/242-temporal-naming-abstraction-plan.md`。
