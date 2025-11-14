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

## Types & Contracts (T3 / Plan 245 起步)

- 新增统一类型（Shared Types）：
  - `frontend/src/shared/types/temporal-entity.ts`：`TemporalEntityRecord`、`TemporalEntityTimelineEntry`、`TemporalEntityStatus`（组织/职位保留为类型别名以便渐进迁移）
  - `frontend/src/shared/hooks/useTemporalEntityDetail.ts`：统一 Hook 骨架，职位复用 `usePositionDetail`，组织按 GraphQL 详情归一化为 `TemporalEntityRecord`
- 现状：`PositionDetailQuery`、`OrganizationUnit` 仍在仓库存在，作为清理对象（见 logs/plan242/t3/36-old-naming-baseline.log）；契约与实现的重命名将在后续提交中分阶段收敛并补充证据
- GraphQL Operation 统一（第一步）：
  - 将职位详情文档操作名由 `PositionDetail` 重命名为 `TemporalEntityDetail`（文件：`frontend/src/shared/hooks/useEnterprisePositions.ts`），并重新生成 `frontend/src/generated/graphql-types.ts`
  - 防回归：新增 Plan 245 Guard（`scripts/quality/plan245-guard.js`）冻结 `query PositionDetail` / `PositionDetailQuery` 的新增使用；首次基线见 `reports/plan245/baseline.json`
  - 组织树查询（不影响测试 Mock 的项）：
    - `GetChildren` → `TemporalEntityTreeChildren`、`GetOrganizationSubtree` → `TemporalEntitySubtree`（文件：`frontend/src/features/organizations/components/OrganizationTree.tsx`）
    - 说明：`GetRootOrganizations` / `GetRootChildrenCount` 被测试用例通过字符串匹配模拟，暂不改名，避免破坏用例
