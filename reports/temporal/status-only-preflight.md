# Status-Only Migration Preflight Checklist

## 1. 执行信息
- **执行日期**: 2025-09-17
- **操作人**: 自动化试运行（Agent）
- **审批人**: 待会签
- **脚本版本**: `node scripts/generate-implementation-inventory.js` @ HEAD

## 2. 实现清单对比
- **差异概述**: 当前实现清单与既有契约一致，未发现新增或缺失的软删除相关条目。
- **受影响模块**: 
  - 命令服务: 软删除逻辑集中于 `TemporalService` 与 `OrganizationTemporalService`，后续需统一改造。
  - 查询服务: GraphQL Schema 未暴露 `deletedAt` 判定细节，需配合契约更新。
  - 前端: `statusUtils`、`organizationPermissions` 等仍依赖 status 状态，需要确认删除逻辑调整影响。
  - 脚本/工具: 多个 temporal 维护脚本包含 `deleted_at` 判定，迁移时需同步。
- **未覆盖风险**: 需验证是否存在未列入实现清单的 ad-hoc SQL 或外部集成依赖 `deleted_at`。

## 3. 契约更新摘要
- OpenAPI 变更: 待更新（需描述软删除仅依赖 status）。
- GraphQL 变更: 待更新（需在 schema 中标注 deletedAt 为审计字段）。
- 额外文档: `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 待同步说明。

## 4. 评审会纪要
- 会议日期: 待安排（P0）
- 参会人员: 命令服务 Owner、查询服务 Co-owner、前端代表、数据平台代表、QA、运维
- 核心结论: 待填
- 待办事项: 待填

## 5. 附件列表
- `implementation-inventory-before.json`: `reports/temporal/implementation-inventory-before.json`
- `implementation-inventory-after.json`: `reports/temporal/implementation-inventory-after.json`
- `status-only-audit.json`: `reports/temporal/status-only-audit.json`
- 合同差异截图/链接: 待补充（契约更新后）

---

> 填写完成后，请将本清单存档于 `reports/temporal/`，并在 `docs/development-plans/06-integrated-teams-progress-log.md` 登记。 
