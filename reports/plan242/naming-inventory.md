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
