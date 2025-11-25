# 411 · Temporal UI SOM 体验重构

**关联计划**：Plan 241、Plan 300、Plan 400、Plan 402D、Plan 410  
**状态**：规划 - 待启动  
**范围**：前端组织/职位 Temporal 体验、Manifest/Slot、GraphQL Facade、Playwright/OBS 证据  
**日志要求**：`logs/plan411/ui/*.log`、`logs/plan411/manifest/*.log`、`logs/plan411/obs/*.log`

> 目标：丢弃 legacy 字段与 timeline adapter，全面使用 `standardObjectAdapter` 与 SOM Facade 作为唯一数据源，同时重新建立 Manifest、OBS、Playwright 证据体系。

---

## 1. 目标

1. Temporal UI（组织/职位列表、详情、Tab、表单）仅消费 SOM 字段与 Facade API，不再解析 legacy 字段。
2. `TemporalMasterDetail`、`TimelineComponent`、`TabNavigation` 等组件全部迁移到 `standardObjectAdapter` 输出，删除 `organizationTimelineAdapter`。
3. Manifest/Slot 重新生成，与 SOM 字段对齐，`npm run manifest:*`、`npm run quality:preflight -- guard=manifest`、`npm run test:e2e -- --grep temporal-manifest` 全部通过。
4. Playwright + OBS 用例覆盖 SOM 场景（读、写、回滚、asOf），日志/截图落盘 `logs/plan411/ui/` 与 `logs/plan411/obs/`。
5. Facade 层提供 `queryStandardObjectTimeline`、`queryStandardObjectKernel` 等 API，组件任何 GraphQL 请求都通过 Facade。

---

## 2. 工作项

### U1 · 数据接入
- 扩展 `frontend/src/shared/api/facade/organization.ts` / `.../position.ts`，统一封装 SOM GraphQL 查询。
- `useTemporalMasterDetail`、`useTemporalEntityDetail`、`TimelineComponent` 仅接受 SOM 聚合。
- 删除 `organizationTimelineAdapter`、`positionTimelineAdapter` 中 legacy 特判；`standardObjectAdapter` 扩展必要字段（audit、links）。

### U2 · UI/Tab/表单
- `TemporalMasterDetailView`、`InlineNewVersionForm`、`TabNavigation` 全部以 SOM 字段渲染：
  - Tab 中展示 `versionCode`、`transaction_range`、`links`、`DEC` 元数据。
  - 表单默认值由 SOM 版本提供；历史版本编辑与回滚操作对齐 `standardobjectapi`。
- 更新 data-testid，使用 `temporalEntitySelectors` 新常量；`selector-guard-246` 必须 0 警告。

### U3 · Manifest/Slot
- 重新运行 `npm run manifest:forms`、`npm run manifest:columns`，输出 `logs/plan411/manifest/*.log`。
- Manifest 记录 SOM 字段（`code`, `versionCode`, `asOfValid`, `asOfTransaction`，links 等）。
- `node scripts/generate-implementation-inventory.js` 更新引用，防止重复实现。

### U4 · Playwright / OBS
- 更新 `frontend/tests/e2e/temporal-graphql-comprehensive.spec.ts`、`standard-object-lifecycle.spec.ts`：
  - 增加 “禁写期间只读提示”、“版本回滚”、“asOf 参数” 三类脚本。
  - OBS 事件 `[OBS] standardObject.*` 收敛为 SSoT，在 `logs/plan411/obs/*.log` 留证。
- `PW_OBS=1 VITE_OBS_ENABLED=true npm run test:e2e -- --grep standard-object` 必须通过。

### U5 · 文档与 Runbook
- 更新 `docs/reference/temporal-entity-experience-guide.md`、`docs/reference/standard-object-evidence-guide.md`，说明 SOM 数据链路。
- 在 Plan 402D/411 中记录 UI 变更、日志路径、Manifest 版本。

---

## 3. 交付物

- Facade API、Hook、组件改造 diff。
- Manifest/Slot 生成日志，Playwright/OBS 运行日志。
- 文档更新（参考手册、Runbook）。
- QA 验收报告，列出所有日志及截图。

---

## 4. 验收标准

1. Temporal UI 中所有数据均来源于 SOM（可通过断网 legacy API 验证）。
2. `npm run manifest:*`, `npm run quality:preflight -- guard=manifest`, `npm run test:e2e -- --grep temporal-manifest`、`PW_OBS=1 VITE_OBS_ENABLED=true npm run test:e2e -- --grep standard-object` 全部通过。
3. OBS/Playwright 日志落盘并在文档中引用，`selector-guard-246` 0 告警。
4. 前端无 `organizationTimelineAdapter` / legacy 字段解析；`standardObjectAdapter` 覆盖所有 UI 入口。
5. 文档、Runbook、实施清单已更新事实来源。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Hook/组件切换导致 UI 回归 | 组织/职位详情不可用 | 在每个步骤后立即运行 `npm run test` + Playwright，日志留存；蓝绿发布策略。 |
| Manifest/OBS 不一致 | CI 阻断、体验不一致 | Manifest 生成与 OBS 测试强制配对执行，日志不可缺。 |
| Facade 接口不完整 | 组件重复造轮子 | 统一在 Facade 暴露 GraphQL 查询，禁止直接 import GraphQL 文档。 |

---

## 6. 依赖与事实来源
- `frontend/src/shared/api/facade/*`
- `frontend/src/features/temporal/**`
- `docs/reference/temporal-entity-experience-guide.md`
- Plan 241（前端框架重构）、Plan 300（Manifest 规范）、Plan 400/402
