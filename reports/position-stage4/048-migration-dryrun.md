# 048 Extend Position Assignments — Migration Dry Run Log

> **目的**：演练 `048_extend_position_assignments.sql` 迁移与回滚，验证 `acting_until`/`auto_revert`/`reminder_sent_at` 字段在 PostgreSQL 上的兼容性与性能影响，为 Stage 4 上线做准备。

## 1. 环境准备

- **数据库版本**：PostgreSQL 13+（与生产一致）
- **基线数据**：使用 `047` 之后的最新 schema 与职位任职样本数据
- **工具**：`psql`、`make db-migrate-all`（如需通过 Makefile 驱动）
- **注意事项**：
  - 迁移前备份 `position_assignments` 表（例如 `pg_dump -t position_assignments`）。
  - 演练期间禁用业务写入，确保数据集保持静态。

## 2. 执行步骤

```bash
# 1. 验证当前 schema 版本
SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;

# 2. 执行迁移
psql "$DATABASE_URL" -f database/migrations/048_extend_position_assignments.sql

# 3. 校验新增列与索引
\d+ position_assignments
SELECT column_name, data_type FROM information_schema.columns
  WHERE table_name = 'position_assignments'
    AND column_name IN ('acting_until','auto_revert','reminder_sent_at');

# 4. 数据检查（示例）
SELECT assignment_id, acting_until, auto_revert, reminder_sent_at
  FROM position_assignments
  ORDER BY updated_at DESC
  LIMIT 10;

# 5. 性能观测（可选）
EXPLAIN ANALYZE
SELECT * FROM position_assignments
 WHERE tenant_id = :tenant
   AND assignment_type = 'ACTING'
   AND auto_revert = true
   AND acting_until < CURRENT_DATE + INTERVAL '7 days';
```

## 3. 回滚步骤（如需）

```bash
psql "$DATABASE_URL" -f database/migrations/rollback/048_drop_position_assignments_extensions.sql
```

回滚后请重新执行 `\d+ position_assignments` 确认列与索引已还原，并确保应用日志无异常。

## 4. 验收检查清单

- [ ] 迁移脚本执行成功并记录时间戳
- [ ] 新增列可见且默认值符合预期（`auto_revert=false`）
- [ ] 受影响的业务查询（SumActiveFTE、任职列表）正确返回
- [ ] 回滚脚本验证通过（如执行）
- [ ] 日志及监控（PG locks / long running queries）无异常

## 5. 备注

- 演练完成后请将上述检查项结果补充至此文档，并在 86 号计划中勾选“048 迁移 & 回滚脚本 + 演练日志”。
- 若演练过程中发现性能或锁冲突问题，请同步更新 `docs/development-plans/06-integrated-teams-progress-log.md` 的风险记录。

## 6. 演练记录（2025-10-20）

- [x] 前置：执行 `047_rename_position_assignments_start_date.sql`，确保字段已统一为 `effective_date`。
- [x] 迁移：`psql 'postgresql://user:password@localhost:5432/cubecastle?sslmode=disable' -f database/migrations/048_extend_position_assignments.sql`
  - 新增列校验：`acting_until` / `auto_revert` / `reminder_sent_at` 均存在，默认值 `auto_revert=false`。
  - 示例查询与 `EXPLAIN` 正常，未观察到锁等待。
- [x] 回滚：已于 2025-10-21 在开发环境演练（见 `reports/position-stage4/048-migration-dryrun-20251021.log`）。
- [x] 受影响功能回归：REST/GraphQL 脚本已运行（参考 `reports/position-stage4/position-assignments-cross-tenant.log` 与 `reports/position-stage4/position-assignments-graphql-cross-tenant.log`）。
