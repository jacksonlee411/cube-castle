# Plan 244 – Temporal Timeline & Status 抽象执行记录

date: 2025-11-10
window: Day 6-8
status: completed

## delta
- 新增 `frontend/src/features/temporal/entity/timelineAdapter.ts`，提供 position/organization 双适配器与事件映射。
- 新增 `frontend/src/features/temporal/entity/statusMeta.ts`，输出 `TEMPORAL_ENTITY_STATUS_META` 及实体专用 getter。
- `PositionDetailView`、职位列表/版本组件改用共享适配器与元数据。
- `temporalMasterDetailApi` 引入组织适配器；GraphQL/REST loader 删除重复 mapper。
- ESLint 禁止引用 Legacy 职位 timeline/status 元数据路径，统一指向 Temporal Entity 命名空间。
- Docs/Inventory/Plan 242 命名清单同步更新。

## verification
- `npm run lint`（前端 API 合规规则）——验证新 `no-restricted-imports` 配置与共享适配器引用。

## notes
- `logs/plan242/t2/` 作为 T2 阶段日志根目录；若再增量更改，按日期追加。
