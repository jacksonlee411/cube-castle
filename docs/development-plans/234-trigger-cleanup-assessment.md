# Plan 234 — organization_units 触发器清理评估

**关联需求**：Plan 233 在执行 `scripts/validate-audit-recordid-consistency.sql` 时仍检测到 6 个触发器（`audit_changes_trigger` 等）存在，与脚本“TRIGGERS ON organization_units (SHOULD BE 0 AFTER 022)” 的期望不符。为避免在未来的迁移/SQL 验证中持续提示，需要单独评估这些触发器的保留价值与回收步骤。

---

## 1. 背景
- `scripts/validate-audit-recordid-consistency.sql` 第 60-72 行将组织表触发器视为“022 之后应该清空”的治理项，本地校验仍输出 6 条触发器记录。
- 历史文档（参考 `docs/archive/development-plans/210-database-baseline-reset-plan.md` 与 `database/migrations/20251106000000_base_schema.sql`）显示，这批触发器与早期计算层/层级校验逻辑绑定；但现行架构已将相应逻辑迁移到应用层或存储过程，继续保留可能造成双写与性能开销。

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
| D | 产出迁移或脚本更新 | DB 工程 | 若决定删除：编写 Goose 迁移（drop trigger + 文档更新）；若决定保留：更新 `scripts/validate-audit-recordid-consistency.sql` 的期望及注释。 |
| E | 文档同步 | 所有相关 | 将结论同步至 `docs/reference/05-AUDIT-AUTH-GUIDE.md` / `database/migrations/README.md`，确保唯一事实来源一致。 |

---

## 4. 验收标准
- 6 个触发器的定义与现状被完整记录并有明确去留方案。
- 若删除：提供迁移、回滚与验证步骤；`validate-audit-recordid-consistency.sql` 不再报错。
- 若保留：脚本及文档说明更新，CI 提示与实际期望一致。

---

## 5. 附录
- 校验输出：`docker compose -f docker-compose.dev.yml exec -T postgres psql -U user -d cubecastle -f scripts/validate-audit-recordid-consistency.sql`
- 触发器名单：`audit_changes_trigger`、`enforce_temporal_flags_trigger`、`trg_prevent_update_deleted`、`update_hierarchy_paths_trigger`、`validate_parent_available_trigger`、`validate_parent_available_update_trigger`
