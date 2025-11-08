# 230 – Job Catalog 缺口 Root Cause（2025-11-07 12:37 CST）

- 操作者：Codex CLI（commit 39d9a9e8ca60c9a44e8dd59dde9b6decdc54c5d6）
- 发现：`docker compose -f docker-compose.dev.yml exec postgres psql ...` 查询结果为 0 行（详见 `logs/230/job-catalog-audit-20251107T1236.log`），表明 `job_family_groups`/`job_families`/`job_roles`/`job_levels` 无 `OPER%` 记录。
- 迁移审计：`database/migrations/20251106000000_base_schema.sql` 仅定义表结构，未插入任何 Job Catalog 参考数据；`rg -n "OPER" database/migrations` 无结果。历史 `sql/init/02-sample-data.sql` 仅包含组织示例，同样缺失 Job Catalog 种子。说明 219E/本地环境在 20251106 之后未再回灌 OPER 层级，导致参考数据完全缺席。
- 结论：`OPER` 层级并非被禁用，而是根本未在当前迁移链/示例数据中创建。需要新增迁移脚本回灌 `OPER` JobFamilyGroup → `OPER-OPS` JobFamily → `OPER-OPS-MGR` 等角色与等级，确保 Playwright/REST 可用。
- 后续：进入 3A.3，编写 `database/migrations/230_job_catalog_oper_fix.sql` 并在 `make db-migrate-all` 中验证幂等。

## 更新（2025-11-08 09:25 CST）

- 操作者：Codex CLI（commit 5b6e484be4002edaa0d329d195ea9d5fe434cfbd）
- 取证：`logs/230/job-catalog-audit-20251108T092533.log` 显示 `OPER` → `OPER-OPS` → `OPER-OPS-MGR` → `S1~S3` 均为 `ACTIVE`。说明最新数据库已经存在参考数据。
- 迁移定位：`rg -n "OPER" database/migrations` 发现 `database/migrations/20251107123000_230_job_catalog_oper_fix.sql`，脚本内容与 230B 目标一致（幂等 Upsert + Down），推测已有成员提交草稿。
- 判断：当前本地数据库可能已执行上述迁移，而 219E 环境尚未统一执行，导致 Playwright 仍报缺失。需要：
  1. 在 230B 中确认 `20251107123000_230_job_catalog_oper_fix.sql` 是否已合入主干并通过 `make db-migrate-all`；
  2. 在 219E/219T 环境确保相同迁移执行，避免“部分环境有数据、部分没有”的跨层不一致；
  3. 后续诊断脚本（230C）必须检测到“完全缺失 / status!=ACTIVE”两类情况，以提示运行该迁移。
