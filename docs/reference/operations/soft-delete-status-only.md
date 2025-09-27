# 软删除 status-only 运维手册（初稿）

## 1. 背景
- 本手册记录软删除改为仅依赖 `status` 字段后的运维指引，包括备份、回滚、监控与告警要求。

## 2. 备份与回滚
- **备份命令示例**
  ```bash
  pg_dump "$DATABASE_URL" --table=organization_units > backup/org_units_pre_status_only.sql
  ```
- **回滚命令示例**
  ```bash
  psql "$DATABASE_URL" -f backup/org_units_pre_status_only.sql
  ```
- 执行前需在 `docs/development-plans/06-integrated-teams-progress-log.md` 记录审批。

## 3. 迁移执行注意事项
- 在维护窗口执行 `database/migrations/029_soft_delete_status_only.sql`；
- 迁移过程中关注锁等待与 statement timeout 日志；
- 如有失败，使用回滚脚本恢复并阻断后续阶段。

## 4. 监控与告警
- Prometheus 指标：
  - `org_units_active_total` —— 统计 `status <> 'DELETED'` 的记录数；
  - `org_units_deleted_total` —— 统计 `status = 'DELETED'` 的记录数；
- Grafana 面板修改：
  - 更新软删趋势图数据源为 `status` 判定；
  - 确认阈值已同步给运维团队。

## 5. 操作日志与审计
- 所有数据修复需写入 `logs/temporal-migration-status-only.log`，格式参考 `.log.example`；
- `audit_logs` 内的 `change_reason` 统一为 `STATUS_ONLY_MIGRATION`。

## 6. 应急联系人
- 命令服务 Owner: 
- 数据平台代表: 
- 运维值班: 

---

> 最终版本需在迁移完成前补充所有占位信息，并在 Phase 5 归档。 
