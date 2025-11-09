# Plan 234 — organization_units 触发器清理评估

**关联需求**：Plan 233 在执行 `scripts/validate-audit-recordid-consistency.sql` 时仍检测到 6 个触发器（`audit_changes_trigger` 等）存在，与脚本“TRIGGERS ON organization_units (SHOULD BE 0 AFTER 022)” 的期望不符。为避免在未来的迁移/SQL 验证中持续提示，需要单独评估这些触发器的保留价值与回收步骤。

---

## 1. 背景
- `scripts/validate-audit-recordid-consistency.sql` 明确打印 `TRIGGERS ON organization_units (SHOULD BE 0 AFTER 022)`，并在 `64-70` 行直接查询 `pg_trigger`，因此当前验证脚本仍把任何残留触发器视为违规（`scripts/validate-audit-recordid-consistency.sql:64-70`）。
- 最新基线迁移仍创建 6 个触发器，说明数据库层面的“022 之后清空”目标尚未兑现（`database/migrations/20251106000000_base_schema.sql:1528-1566`）。
- API 变更日志已经宣布移除 `log_audit_changes/audit_changes_trigger` 并改由应用层结构化审计作为唯一事实来源，且要求按 `027_*`、`028_*` 迁移完成退场，这与基线 Schema 形成冲突（`docs/api/CHANGELOG.md:16-42`）。
- 现行应用层已经提供了与这些触发器等价的约束：审计写入由命令服务调用 `AuditLogger` 插入 `audit_logs`（`internal/organization/audit/logger.go:131-189`）；父子层级、状态与时态校验由 `BusinessRuleValidator` 的 `ORG-DEPTH`/`ORG-CIRC`/`ORG-STATUS`/`ORG-TEMPORAL` 规则执行（`internal/organization/validator/organization_rules.go:30-360`）；`TemporalTimelineManager` 负责维护 `is_current`、`end_date` 等派生字段，替代 `update_hierarchy_paths` 类触发器（`internal/organization/repository/temporal_timeline_manager.go:1-120`）。继续保留触发器会造成双写、冲突与性能风险。

---

## 2. 工作目标
1. **盘点触发器职责**：逐一梳理 `audit_changes_trigger`、`enforce_temporal_flags_trigger`、`trg_prevent_update_deleted`、`update_hierarchy_paths_trigger`、`validate_parent_available_trigger`、`validate_parent_available_update_trigger` 的 SQL 定义、调用链与替代实现。
2. **评估是否可安全移除**：结合 `internal/organization/repository` 与 `temporal_timeline` 的实现，确认应用层是否已提供等效约束；若无，需要制定替代方案或更新脚本期望。
3. **形成清理方案**：若确认可删除，给出迁移脚本（Goose）和回滚策略；若需要保留，则更新 `scripts/validate-audit-recordid-consistency.sql` 说明，并在参考文档中同步期望值。

---

## 3. 行动项
| 编号 | 任务 | 负责人 | 说明 |
| --- | --- | --- | --- |
| A | 导出触发器定义 | DB 工程 | 通过 `pg_get_triggerdef` 获取 6 个触发器的 SQL，整理进入本计划附件。 |
| B | 代码路径分析 | Backend | 逐个对照 `internal/organization/repository/*` 与 `temporal timeline` 逻辑，判断是否存在重复校验。 |
| C | 风险评估会议 | 架构/DB | 输出触发器保留/删除的决策表，包含数据一致性风险、回滚策略。 |
| D | 产出迁移或脚本更新 | DB 工程 | 若决定删除：先更新唯一事实来源 `database/schema.sql`，再根据 `database/migrations/README.md:3-7` 生成 Goose 迁移（drop trigger + 回滚脚本），并同步调整 `scripts/validate-audit-recordid-consistency.sql` 的期望；若决定保留：同样更新脚本注释，解释例外理由。 |
| E | 文档同步 | 所有相关 | 将结论同步至 `docs/reference/05-AUDIT-AUTH-GUIDE.md` / `database/migrations/README.md`，确保唯一事实来源一致。 |
| F | 契约校验回归 | Backend/QA | 以 `docs/api/CHANGELOG.md:16-42` 为基线复核命令服务审计逻辑（`internal/organization/audit/logger.go`）与 `validate-audit-recordid-consistency.sql` 输出，记录验证截图/日志，防止再次产生“脚本期望 vs. Schema”漂移。 |

---

## 4. 验收标准
- 6 个触发器的定义与现状被完整记录并有明确去留方案，且引用的事实来源（`scripts/validate-audit-recordid-consistency.sql`、`docs/api/CHANGELOG.md`、`database/migrations/20251106000000_base_schema.sql`）在计划中可追溯。
- 若删除：`database/schema.sql`、对应 Goose 迁移与回滚脚本、`scripts/validate-audit-recordid-consistency.sql` 均已更新，`docker compose ... validate-audit-recordid-consistency.sql` 执行结果为零触发器。
- 若保留：脚本与文档解释了例外，CI 校验与 `docs/reference/05-AUDIT-AUTH-GUIDE.md` 保持一致，并通过 `AuditLogger`/业务规则验证证明不可替代原因。

---

## 5. 附录
- 校验输出：`docker compose -f docker-compose.dev.yml exec -T postgres psql -U user -d cubecastle -f scripts/validate-audit-recordid-consistency.sql`
- 触发器名单：`audit_changes_trigger`、`enforce_temporal_flags_trigger`、`trg_prevent_update_deleted`、`update_hierarchy_paths_trigger`、`validate_parent_available_trigger`、`validate_parent_available_update_trigger`
