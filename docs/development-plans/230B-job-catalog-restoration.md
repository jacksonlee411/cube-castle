# Plan 230B – OPER Job Catalog 数据修复迁移

**文档编号**: 230B  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**前置计划**: 230A（现状取证）、Plan 210 Goose/Atlas 基线  
**负责人**: 数据库团队

---

## 1. 背景与目标

- 230A 已确认 `OPER` JobFamilyGroup 及其子级在 219E 环境缺失/失效；219T E2E 被 `JOB_CATALOG_NOT_FOUND` 阻塞。  
- 230B 负责交付**幂等的数据修复迁移**，恢复 `OPER` → `OPER-OPS` → `OPER-OPS-MGR` → `S1~S3` 等完整层级，并在 `make db-migrate-all` 中可重复执行。  
- 同步保留回滚路径，并输出验证日志供 219E 引用。

---

## 2. 范围

1. 新增 `database/migrations/<timestamp>_230_job_catalog_oper_fix.sql`（或等效 goose/atlas 文件），包含 Up/Down 逻辑。  
2. 脚本需包含幂等检查（`IF NOT EXISTS` / `ON CONFLICT DO UPDATE`），并覆盖 `job_family_groups`、`job_families`、`job_roles`、`job_levels` 四张表。  
3. 本地运行 `make db-migrate-all` 验证脚本可执行，再次运行不会报错；若 Down 可行，需在 README 中说明。  
4. 复跑 230A SQL，证明 `OPER` 相关记录全部为 `ACTIVE` 且字段齐全。

---

## 3. 详细任务

| 步骤 | 描述 | 输出 |
| --- | --- | --- |
| B1 | 基于 230A 的 `logs/230/root-cause.md` 明确缺失字段、预期属性（code/name/description/level/status） | 记录到新迁移文件的注释中 |
| B2 | 创建迁移文件：`-- +goose Up` 段中先 `LOCK TABLE ... IN SHARE ROW EXCLUSIVE MODE`（如需），再执行幂等 `INSERT ... ON CONFLICT` 或 `UPDATE`；`-- +goose Down` 删除同一批数据 | `database/migrations/<timestamp>_230_job_catalog_oper_fix.sql` |
| B3 | 在 PR 中附加 SQL 片段，说明字段来源（`docs/reference/02-IMPLEMENTATION-INVENTORY.md`）与差异 | PR 模板或 `logs/230/root-cause.md` 附注 |
| B4 | 运行 `make db-migrate-all`，随后执行 230A 的 SQL 验证，输出到 `logs/230/job-catalog-audit-YYYYMMDD.log`（追加 run-id） | 日志 + `git diff` |
| B5 | 若 Down 脚本不可行（例如依赖其它计划），需在本文件“风险”小节说明，并提供手动回滚指南 | 文档说明 |

---

## 4. 依赖

- 230A 提供的取证日志与 root-cause 结论。  
- Goose/Atlas 配置来自 Plan 210，需保持相同命名规范。  
- Docker 容器数据库必须可写（`make docker-up && docker compose exec postgres ...`）。

---

## 5. 验收标准

1. 迁移文件通过 `make db-migrate-all`，再次运行不报错。  
2. 迁移执行后，230A 的 SQL 查询全部返回 `status='ACTIVE'`，缺失记录得到补齐。  
3. `logs/230/job-catalog-audit-YYYYMMDD.log` 附带执行前/后的 diff 或至少记录 “Before/After” 分隔，便于审阅。  
4. PR 中附带 `logs/230/root-cause.md` 摘要与 `git diff database/migrations/`，Reviewer 可直接核对。  
5. 230B 成果被 230C/230D 作为前置引用。

---

## 6. 交付记录（2025-11-08）

- **迁移脚本**：`database/migrations/20251107123000_230_job_catalog_oper_fix.sql` 已合入主干，`make db-migrate-all` 显示当前版本 `20251107123000`（日志见 `logs/230/job-catalog-audit-20251108T092533.log`）。  
- **验证日志**：`logs/230/job-catalog-audit-20251108T092533.log` 记录 `OPER` → `OPER-OPS` → `OPER-OPS-MGR` → `S1~S3` 均为 `ACTIVE`，满足验收标准 1/2；`logs/230/root-cause.md` 更新 2025-11-08 说明。  
- **复用声明**：230C/230D 在执行诊断/Playwright 之前均引用该迁移与日志作为唯一事实来源。

---

> 唯一事实来源：`database/migrations/20251107123000_230_job_catalog_oper_fix.sql`、`logs/230/root-cause.md`、`logs/230/job-catalog-audit-20251108T092533.log`。  
> 更新时间：2025-11-08。
