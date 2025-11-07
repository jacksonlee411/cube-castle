# 230 – Job Catalog 缺口 Root Cause（2025-11-07 12:37 CST）

- 操作者：Codex CLI（commit 39d9a9e8ca60c9a44e8dd59dde9b6decdc54c5d6）
- 发现：`docker compose -f docker-compose.dev.yml exec postgres psql ...` 查询结果为 0 行（详见 `logs/230/job-catalog-audit-20251107T1236.log`），表明 `job_family_groups`/`job_families`/`job_roles`/`job_levels` 无 `OPER%` 记录。
- 迁移审计：`database/migrations/20251106000000_base_schema.sql` 仅定义表结构，未插入任何 Job Catalog 参考数据；`rg -n "OPER" database/migrations` 无结果。历史 `sql/init/02-sample-data.sql` 仅包含组织示例，同样缺失 Job Catalog 种子。说明 219E/本地环境在 20251106 之后未再回灌 OPER 层级，导致参考数据完全缺席。
- 结论：`OPER` 层级并非被禁用，而是根本未在当前迁移链/示例数据中创建。需要新增迁移脚本回灌 `OPER` JobFamilyGroup → `OPER-OPS` JobFamily → `OPER-OPS-MGR` 等角色与等级，确保 Playwright/REST 可用。
- 后续：进入 3A.3，编写 `database/migrations/230_job_catalog_oper_fix.sql` 并在 `make db-migrate-all` 中验证幂等。
