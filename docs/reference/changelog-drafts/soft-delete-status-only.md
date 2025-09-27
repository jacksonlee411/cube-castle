# 软删除判定改为仅依赖 status - 契约变更草案

## 1. 变更摘要
- **目标**: 软删除判定仅依赖 `status` 字段，`deletedAt` 保留为可选审计字段。
- **涉及接口**: 
  - REST: `/api/v1/organization-units/**`
  - GraphQL: `Organization`, `OrganizationTimeline`, 其他返回组织状态的查询/片段。
- **预期上线窗口**: 目标 2025 Q4，具体日期待发布委员会确认

## 2. 契约差异（待确认）
- OpenAPI 变更点: 
  - 在 `OrganizationUnit` 响应模型 description 中明确 `deletedAt` 为审计字段；
  - 删除关于 `deleted_at` 判定的字面描述，将软删除判定集中在 `status`；
  - 更新示例请求/响应，确保描述与字段语义一致；
- GraphQL Schema 变更点: 
  - 更新 `type Organization` 中 `deletedAt` 字段说明；
  - 在 Schema 顶部注释补充“软删除仅依赖 status”；
- 其他文档同步（参考手册/FAQ 等）: 
  - 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`；
  - 前端 FAQ、运营手册同步软删除语义变化。

## 3. 会签清单
| 角色 | 负责人 | 状态 | 备注 |
| --- | --- | --- | --- |
| 命令服务 Owner | 待指派 | 未开始 |  |
| 查询服务 Co-owner | 待指派 | 未开始 |  |
| 前端代表 | 待指派 | 未开始 |  |
| 数据平台代表 | 待指派 | 未开始 |  |
| QA 代表 | 待指派 | 未开始 |  |
| 运维/发布管理 | 待指派 | 未开始 |  |

## 4. 风险与兼容性评估
- 回滚路径: 回退至旧版 migration 并恢复 `backup/org_units_pre_status_only.sql`
- 客户端版本要求: 前端需同步更新；旧版客户端需兼容 `deletedAt` 为空场景
- 外部集成影响: 通知所有直接查询数据库的集成方更新 SQL 判定条件

## 5. 待办与开放问题
- [ ] 更新 OpenAPI 描述与示例
- [ ] 更新 GraphQL schema & description
- [ ] 通知外部集成/前端消费方
- [ ] 复核实现清单差异与 TODO
- [ ] 准备发布公告/FAQ

## 6. 会议纪要（填写示例）
- 会议日期: 待定（Phase 0 内需完成）
- 结论概述: 待填
- 行动项: 待填
- 下次检查点: 待填

---

> 契约签核完成后，请将本草案更新为“已发布”状态，并在 `docs/development-plans/06-integrated-teams-progress-log.md` 登记。 
