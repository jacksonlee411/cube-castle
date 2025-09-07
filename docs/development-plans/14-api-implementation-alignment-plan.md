# API 与实现一致性对齐计划 v1.0

目的
- 归档本次“API 文档与代码实现”一致性排查结果，形成可执行的对齐计划与验收标准。

范围
- 仅覆盖组织域相关端点：创建、更新、版本管理、事件（DEACTIVATE）、停用/启用、整单位删除（已移除）。

一、排查结论（不一致与异味）

1) 更新动词不一致（高优先级）
- 文档：`PATCH /api/v1/organization-units/{code}`（部分更新）
- 实现：`PUT /api/v1/organization-units/{code}`（organization.go:720）
- 影响：契约与实现不一致，可能导致客户端 405/404。

2) 版本维护端点形态不一致（高优先级）
- 文档：聚焦 `POST /{code}/versions`（新增版本），删除推荐走 `/{code}/events`（DEACTIVATE），历史修正走 `/{code}/history/{record_id}`。
- 实现：另外暴露了 `DELETE /organization-units/versions/{recordId}` 与 `PUT /organization-units/versions/{recordId}/effective-date`（无 `{code}` 段）。
- 影响：路径风格不统一（部分有 code、部分仅 recordId），客户端难以把握最佳实践。

3) 租户隔离头未在文档统一声明（高优先级）
- 实现：所有 handler 通过 `X-Tenant-ID` 获取租户（organization.go:666）。
- 文档：多数端点未声明该 Header。
- 影响：多租户调用不规范，可能命中默认租户导致数据串扰。

4) 停用/启用返回体未完全在文档固定（中优先级）
- 实现：返回 `timeline[]`（组织时态最新非删除时间线）。
- 文档：已补充“Temporal Behavior + timeline 返回”；需确保字段名一致（recordId/code/effectiveDate/endDate/isCurrent/status）。

5) 事件端点（DEACTIVATE）返回体（中优先级）
- 实现：返回 `code/record_id/timeline`，符合“事务内重算 + 返回最新时间线”。
- 文档：已补充相同结构及示例，基本一致。

6) 整单位删除端点（低优先级）
- 文档：已移除 `DELETE /api/v1/organization-units/{code}`。
- 实现：当前未暴露该路由，保持一致。

7) 运维端点未纳入公开 OpenAPI（低优先级）
- 实现：`/api/v1/operational/*`（健康、指标、任务）。
- 文档：无定义（可保留为内部端点）。

8) 次要命名异味（建议）
- 版本子资源路径：建议统一为 `/{code}/versions/{recordId}`（当前有扁平 `/versions/{recordId}`）。
- 错误码：建议在 OpenAPI 固化 `TEMPORAL_POINT_CONFLICT`、`TEMPORAL_OVERLAP_CONFLICT`、`ORGANIZATION_NOT_FOUND` 等。

二、对齐策略（两条路线，择一或混合）

方案 A（文档收敛到实现，低风险）
- 将文档的 `PATCH` 改为 `PUT`，声明“整体更新”（保留将来增加 PATCH 的余地）。
- 在 OpenAPI 补充：
  - `DELETE /organization-units/versions/{recordId}`（删除版本）
  - `PUT /organization-units/versions/{recordId}/effective-date`（修改生效日）
- 在 OpenAPI 的 `components.parameters` 增加 `X-Tenant-ID` Header，并在相关端点引用。

方案 B（实现收敛到文档，中风险）
- 新增 `PATCH /{code}` 路由（同时保留 `PUT` 兼容期或降级为 PATCH-only）。
- 逐步下线扁平版本端点（`/versions/{recordId}`），统一通过：
  - 事件（DEACTIVATE）删除单条版本
  - 历史（`/{code}/history/{record_id}`）修改内容/日期（日期按“删旧+插新”语义）
- 推出迁移指南与客户端适配期限。

推荐选择
- 当前已在 OpenAPI 增补“事件/版本/停用/启用”的时间线规则与返回契约。为快速稳定交付，建议先采用方案 A（文档收敛到实现），随后在 vNext 评估 B 的收益。

三、实施计划（任务清单）

P0（立即）
- [ ] 统一文档中的 `X-Tenant-ID` 头（components.parameters.TenantIdHeader，并在各端点引用）。
- [ ] 确认文档中的 `PUT /{code}` 用语（去除 PATCH）并标注“禁止修改时态与状态字段”。
- [ ] 在版本章节补充基于 recordId 的删除/改生效日端点说明与示例。

P1（短期）
- [ ] 统一返回体字段名，写清最小稳定子集（recordId/code/effectiveDate/endDate/isCurrent/status）。
- [ ] 在错误码章节补充 `TEMPORAL_*` 与常见 404/409 的示例与触发条件。

P2（中期）
- [ ] 评估将 `/versions/{recordId}` 迁移为 `/{code}/versions/{recordId}` 的价值与影响（与客户端/日志/鉴权对齐）。
- [ ] 评估引入 `PATCH /{code}` 的必要性与落地计划（同时保持 `PUT` 兼容期）。

四、验收标准（可观察结果）
- 所有对外端点在 OpenAPI 与实现路由一致（方法、路径、参数）。
- 多租户调用均带 `X-Tenant-ID` 且通过契约校验。
- 版本管理（新增/删除/改生效日）在契约中有明确路径与示例，并与事件/历史端点口径不冲突。
- 停用/启用/DEACTIVATE 的“结束日期自动更新 + 返回 timeline”在文档与实现一致。
- 合同测试（contract tests）通过：
  - `Create/Update/Versions/Events/Suspend/Activate` 的 happy path
  - `TEMPORAL_POINT_CONFLICT`、`TEMPORAL_OVERLAP_CONFLICT`、`ORGANIZATION_NOT_FOUND` 等错误路径

五、参考定位
- 路由实现：`cmd/organization-command-service/internal/handlers/organization.go:717-729`
- 事件返回 timeline：`internal/handlers/organization.go:520-566`
- 停用/启用返回 timeline：`internal/handlers/organization.go:420-475`
- 租户头读取：`internal/handlers/organization.go:666`
- OpenAPI 关键处：
  - 版本端点：`docs/api/openapi.yaml:329+`
  - 事件端点：`docs/api/openapi.yaml:443+`
  - 停用/启用：`docs/api/openapi.yaml:611+ / 692+`
  - 参数与权限：`docs/api/openapi.yaml:1170+`

—— 以上

