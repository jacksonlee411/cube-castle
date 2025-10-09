# Plan 28 — Version 字段退役验证报告（2025-10-09）

**执行时间**: 2025-10-09  
**执行者**: automation  
**关联计划**: `docs/development-plans/28-version-field-decommission-plan.md`

---

## 1. 数据库验证

> 所有命令均在 `postgresql://user:password@localhost:5432/cubecastle?sslmode=disable` 下执行。

| 检查项 | SQL / 操作 | 结果 |
| --- | --- | --- |
| 主表移除 `version` 列 | `\d organization_units` | `version` 列不存在；触发器 `organization_version_trigger` 已不在列表内 |
| 备份表同步 | `\d organization_units_backup_temporal` / `\d organization_units_unittype_backup` | `version` 列不存在 |
| 相关视图重建 | `SELECT definition FROM pg_views WHERE viewname = 'organization_temporal_current';` | 视图已去除 `version` 字段，并重新创建 |

## 2. 应用构建验证

| 组件 | 命令 | 结果 |
| --- | --- | --- |
| 命令服务 | `go build ./cmd/organization-command-service` | ✅ 成功 |
| 查询服务 | `go build ./cmd/organization-query-service` | ✅ 成功 |
| 前端 | `npm --prefix frontend run lint` | ✅ 成功 |

## 3. API 契约同步

- `docs/api/schema.graphql`：`type Organization` 移除 `version` 字段。
- `docs/api/openapi.yaml`：`OrganizationUnit` 响应模型已移除 `version` 属性。
- 前端 GraphQL 查询 (`OrganizationVersions`, `GetOrganization`) 不再请求 `version` 字段；对应类型与转换器已更新。

## 4. 代码清理摘要

- 命令服务仓储不再查询 / 扫描 `version` 列；触发器相关逻辑删除。
- 查询服务 `model.Organization`、PostgreSQL 仓储、GraphQL 模型同步去除 `VersionField`。
- 前端类型（`OrganizationUnit`、`TimelineVersion`、`TemporalInfo` 等）及展示组件不再使用 `version` 属性。

## 5. 待办 / 风险

- **GraphQL 运行验证**：后续重启查询服务后需通过 `curl http://localhost:8090/graphql` 验证 `organization` / `organizationVersions` 查询响应无 `version` 字段并与契约一致。
- **API 文档更新传播**：需要通知使用 REST / GraphQL 客户端的团队，确认对 `version` 字段移除的兼容性影响。

---

**状态**: ✅ 数据层与代码层变更已完成，等待服务重启后的运行时验证。提交记录及迁移脚本参见 `database/migrations/036_drop_version_column.sql` 与本报告。 
