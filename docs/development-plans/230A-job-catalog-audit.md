# Plan 230A – Job Catalog 现状取证与根因分析

**文档编号**: 230A  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**关联路线图**: Plan 219（219E 端到端验收前置）  
**负责人**: 数据库团队 + Backend

---

## 1. 背景与目标

- 219T/219E 回归中，`tests/e2e/position-crud-full-lifecycle.spec.ts` 在 Step 1 调用 `POST /api/v1/positions` 返回 `JOB_CATALOG_NOT_FOUND`，日志来源：`frontend/test-results/position-crud-full-lifecyc-b9f01-RUD生命周期-Step-1-创建职位-Create--chromium/`。  
- CLI 复现（`curl -X POST http://localhost:9090/api/v1/positions ...`）也获得 422，表明 219E 环境 Job Catalog `OPER` 系列数据缺失或被置为 `INACTIVE`。  
- 本子计划聚焦在**取证**与**根因分析**：在 Docker PostgreSQL 中输出完整证据，并定位迁移/脚本导致的差异，为后续修复提供唯一事实来源。

---

## 2. 范围

1. 运行数据库取证 SQL，覆盖 `job_family_groups`、`job_families`、`job_roles`、`job_levels` 表中 `OPER%` 相关记录，输出状态与必要字段。  
2. 审计 `database/migrations/` 历史，确认导致 `OPER` 缺失或 status=INACTIVE 的 commit / 脚本。  
3. 生成结构化日志与 root-cause 说明，归档到 `logs/230/`，供 230B/230D/219E 复用。

不包含：任何数据修复、迁移脚本编写、自检脚本或 Playwright 执行；这些内容分别由 230B~230F 负责。

---

## 3. 任务与产物

| 步骤 | 描述 | 产物 |
| --- | --- | --- |
| A1 | 启动 Docker Compose（`make docker-up`）并确认 `postgres` 容器 healthy | `logs/230/job-catalog-audit-YYYYMMDD.log` 中记录 `docker compose ps` 与 git commit |
| A2 | 运行以下 SQL，输出 code/status/关键字段，必要时附加 `\gx`：<br>```bash
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U user -d cubecastle \
  -c "SELECT 'job_family_groups' AS table, code, status FROM job_family_groups WHERE code='OPER'
      UNION ALL SELECT 'job_families', code, status FROM job_families WHERE code LIKE 'OPER%';"
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U user -d cubecastle \
  -c "SELECT code, status FROM job_roles WHERE code LIKE 'OPER%' ORDER BY code;"
docker compose -f docker-compose.dev.yml exec postgres \
  psql -U user -d cubecastle \
  -c "SELECT role_code, level_code, status FROM job_levels WHERE role_code LIKE 'OPER%' ORDER BY role_code, level_code;"
``` | 同 `logs/230/job-catalog-audit-YYYYMMDD.log`，要求完整 stdout/stderr |
| A3 | 使用 `rg -n "OPER" database/migrations`、`git log -p -- database/migrations/*oper*.sql` 审计迁移历史，记录修改者、commit、影响字段、是否已上线 | `logs/230/root-cause.md`（含时间、commit、影响分析、后续建议） |
| A4 | 对照 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中 Job Catalog 章节，列出与期望字段的差异（例如缺少 `OPER-OPS-MGR`、等级 S1-S3） | 在 `logs/230/root-cause.md` 中新增“对照表”小节，附上事实来源引用 |

---

## 4. 依赖与前置

- 容器化数据库：严格使用 `make docker-up` + `docker compose -f docker-compose.dev.yml exec postgres`（参考 `AGENTS.md` Docker 约束）。  
- Go 工具链 / 后续脚本暂不需要，但需保留当前 git commit 与 docker 镜像标签。  
- 访问 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 以确认 Job Catalog 期望结构。

---

## 5. 验收标准

1. `logs/230/job-catalog-audit-YYYYMMDD.log` 中包含三段 SQL 输出，且每行记录 `status` 字段；如存在 `NULL` 或 `INACTIVE` 必须高亮说明。  
2. `logs/230/root-cause.md` 说明至少包含：涉及迁移文件路径、commit hash、推断根因、影响范围、是否需要回滚/新增迁移。  
3. 本子计划的结论被引用到 Plan 230 主文档与 230B、230D 的“依赖”章节，保持唯一事实来源。  
4. 无数据写操作；若取证发现严重不一致，应立即通知 230B owner 并暂停后续步骤。

---

> 唯一事实来源：`frontend/test-results/position-crud-full-lifecyc-b9f01-RUD生命周期-Step-1-创建职位-Create--chromium/`、`logs/219E/position-lifecycle-20251107-135246.log`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`。  
> 更新时间：2025-11-07。
