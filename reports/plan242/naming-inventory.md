# Plan 242 – Naming Inventory

## Temporal Entity Page (T1 / Plan 243)

| 实体 | 原入口 | 新入口 | 备注 |
| --- | --- | --- | --- |
| 组织 | `frontend/src/features/organizations/OrganizationTemporalPage.tsx` | `frontend/src/features/temporal/pages/entityRoutes.tsx` (`OrganizationTemporalEntityRoute`) | 由 `TemporalEntityPage` 统一处理路由验证与错误提示，内容层复用 `TemporalMasterDetailView` |
| 职位 | `frontend/src/features/positions/PositionTemporalPage.tsx` | `frontend/src/features/positions/PositionDetailView.tsx` + `TemporalEntityPage` | 路由 shell 统一化，内容由 `PositionDetailView` 承载 |

- 头部导航（返回列表、数据来源提示、动作按钮）由 `PositionDetailView` 内部渲染；组织端由 `TemporalMasterDetailView` 负责。
- 路由校验统一由 `TemporalEntityRouteConfig` 处理，组织支持 7 位数字 & `new`，职位支持 `P\d{7}` 与 `new`。
- 新 `TemporalEntityPage` 在无效编码时输出标准提示，并提供返回列表按钮。
- `reports/plan242/logs/t1/` 将保存阶段性日志（需执行时创建）。

## Temporal Entity Timeline Adapter (T2)

| 实体 | 旧命名 | 新命名 | 备注 |
| --- | --- | --- | --- |
| 组织 | GraphQL/REST 各自手写 mapper | `frontend/src/features/temporal/entity/timelineAdapter.ts` (`organizationTimelineAdapter`) | 统一 `TimelineVersion` 字段与排序；`temporalMasterDetailApi` 直接复用，删除重复逻辑 |
| 职位 | Legacy 职位 timeline adapter（已删除） | `frontend/src/features/temporal/entity/timelineAdapter.ts` (`positionTimelineAdapter`, `positionTimelineEventAdapter`) | 统一 recordId 规则、生命周期/业务状态映射，供 `PositionDetailView` 与 CSV 导出共享 |
- Go 层：`internal/organization/repository/temporal_timeline_manager.go`、`internal/organization/handler/organization_update.go` 输出 `unitType/level/codePath/namePath/sortOrder` 等 `TemporalEntityTimelineVersion` 字段，保证 REST timeline 与前端适配器结构一致；相关响应契约记录在 `docs/api/openapi.yaml#TemporalEntityTimelineVersion`。

## Temporal Entity Status Meta (T2)

| 实体 | 旧命名 | 新命名 | 备注 |
| --- | --- | --- | --- |
| 组织 | `shared/utils/statusUtils.ts` (`STATUS_CONFIG`) | `frontend/src/features/temporal/entity/statusMeta.ts` (`TEMPORAL_ENTITY_STATUS_META.organization`) | `StatusBadge` 与公共组件读取统一元数据，`statusUtils` 仅保留动作判断 |
| 职位 | Legacy 职位 status meta（已删除） | `frontend/src/features/temporal/entity/statusMeta.ts` (`TEMPORAL_ENTITY_STATUS_META.position`) | Position 列表/时间线/版本列表共享同一标签与色板，避免重复维护 |
