# Status-Only 数据修复流程清单

## 1. 前置条件
- 已完成 Phase 0 契约签核并在进展日志登记；
- 数据库备份 `backup/org_units_pre_status_only.sql` 可用；
- 获取 `audit_logs` 查询权限。

## 2. 执行步骤
1. 运行 `psql -f sql/inspection/status_deleted_audit.sql > reports/temporal/status-only-audit.json`；
2. 根据 JSON 审计结果生成修复任务列表，分配负责人与审批人；
3. 按以下优先级修复：
   - `status='DELETED'` & `deletedAt` 为空 → 回填时间；
   - `status<>'DELETED'` & `deletedAt` 非空 → 业务确认后修复；
   - 冲突记录（`conflictingDeletedAtValues`）→ 分析历史版本，保留最新值；
4. 每条修复执行后插入 JSON 日志至 `logs/temporal-migration-status-only.log` 并写入 `audit_logs`；
5. 修复完成后再次运行审计脚本，将输出保存为 `reports/temporal/status-only-audit-after.json`，使用 `jq -s '{baseline: .[0].summary, current: .[1].summary}' reports/temporal/status-only-audit.json reports/temporal/status-only-audit-after.json` 或等效工具比对差异，并同步更新 `reports/temporal/status-only-migration_diff.md`；
6. 在 `docs/development-plans/06-integrated-teams-progress-log.md` 更新状态与剩余风险。

## 3. 审批节点
- 业务确认：组织域负责人；
- 法务审批：法务代表；
- 技术审批：数据平台代表；

## 4. 注意事项
- 禁止直接修改历史备份脚本或绕过审计流程；
- 所有 SQL 操作需包含 `WHERE tenant_id = ?` 条件，避免跨租户影响；
- 如遇异常数据无法判定，需开立缺陷单并暂停迁移。

---

> 完成清单后请附于 Phase 1 报告并归档。 
