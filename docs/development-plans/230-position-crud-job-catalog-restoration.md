# 230 – Position CRUD 参考数据修复计划（母计划）

## 1. 背景与目标

- 219T/219E 回归时，`tests/e2e/position-crud-full-lifecycle.spec.ts` 在 Step 1 调用 `POST /api/v1/positions` 返回 422，错误体 `JOB_CATALOG_NOT_FOUND: JobFamilyGroup OPER is inactive or missing`（日志：`frontend/test-results/position-crud-full-lifecyc-b9f01-RUD生命周期-Step-1-创建职位-Create--chromium/`、`logs/219E/position-lifecycle-20251107-135246.log`）。  
- CLI 复现同样失败，确认命令服务缺失 `OPER` Job Catalog 参考数据，导致 Position CRUD、Assignment、Job Catalog UI 全链路阻断。  
- Plan 230 作为母计划，负责协调恢复 Job Catalog 数据、自检机制、E2E 复验与文档同步，确保 219E 能进入最终验收。

---

## 2. 子计划划分

| 子计划 | 职责范围 | 关键输出 | 状态 |
| --- | --- | --- | --- |
| [230A – Job Catalog 现状取证与根因分析](./230A-job-catalog-audit.md) | 运行 SQL 取证、审计迁移历史、记录 `logs/230/job-catalog-audit-*.log` 与 `logs/230/root-cause.md` | 取证日志、根因说明（最新：`logs/230/job-catalog-audit-20251108T092533.log`） | ✅ |
| [230B – OPER Job Catalog 数据修复迁移](./230B-job-catalog-restoration.md) | 编写幂等迁移 `database/migrations/<timestamp>_230_job_catalog_oper_fix.sql`，恢复 `OPER` 层级 | `database/migrations/20251107123000_230_job_catalog_oper_fix.sql`、`logs/230/job-catalog-audit-20251108T092533.log` | ✅ |
| [230C – Job Catalog 自检脚本与 Make 集成](./230C-job-catalog-diagnostics.md) | 新增 `scripts/diagnostics/check-job-catalog.sh` 并接入 `make status` | 脚本、Makefile 更新、`logs/230/job-catalog-check-20251108T093645.log` | ✅ |
| [230D – Position CRUD 数据链路恢复与 E2E 验证](./230D-position-crud-e2e.md) | 播种 Position/Assignment 数据、复跑 Playwright、归档测试产物 | `logs/230/position-env-check-20251108T095108.log`、`logs/230/position-seed-20251108T094735.log`、`logs/230/position-crud-playwright-20251108T102815.log`、`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` | ✅ |
| [230E – 219T/219E 文档与报告同步](./230E-documentation-sync.md) | 将修复结果同步到 219T/219E/06-log 文档，解除阻塞标记 | 文档 diff、引用路径 | ⏳ |
| [230F – 职位管理功能对齐与测试映射](./230F-position-readiness.md) | 编制功能→测试映射、更新 `logs/230/position-module-readiness.md`、登记 TODO | Readiness 表、测试注记 | ⏳ |

> 若后续需要扩展 Job Catalog 其他 Code，可在 230C/230D 基础上追加 230G 等子计划，此处暂保留 6 个执行单元。

---

## 3. 执行顺序与依赖

1. **230A → 230B**：先完成取证与根因分析，再编写迁移脚本，避免误修。  
2. **230B → 230C**：数据修复完成后才能做健康检查脚本，否则无法验证通过路径。  
3. **230C → 230D**：`make status` 中的诊断需通过后才能启动 Position CRUD E2E。  
4. **230D → 230E/F**：E2E 输出是文档同步与功能对齐的事实来源。  
5. **219E 依赖**：Plan 219E 2.4 “重启前置条件”中的 “Position + Assignment 数据链路恢复” 与 “Playwright P0 场景” 均以 230D/E/F 的结果作为解锁条件。

---

## 4. 管理方式

- 每个子计划在对应文档中维护目标、范围、任务、依赖、验收标准；执行日志统一放在 `logs/230/` 子目录。  
- 当子计划完成后，请在本母计划的表格中更新“状态”列（✅/⏳/阻塞），并在 `docs/development-plans/06-integrated-teams-progress-log.md` 记录一次同步。  
- 若执行中发现与 `CLAUDE.md`、`AGENTS.md` 原则冲突，须立即暂停相关子计划并提交修订。

---

## 5. 验收标准（母计划层级）

1. **数据完整**：230B 迁移执行后，230A SQL 取证显示 `OPER` 及子项 `status='ACTIVE'`。  
2. **健康检查**：`make status` 包含 Job Catalog 自检，缺口会立即失败。  
3. **E2E 恢复**: `tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium）通过，`frontend/test-results/...` 产物归档，RequestId 记录完整。  
4. **文档同步**：219T/219E/06-log 已将 Position CRUD 阻塞标记为解除，并引用新的日志路径。  
5. **功能对齐**：`logs/230/position-module-readiness.md` 记录功能→测试映射，无未说明的断言。  
6. **唯一事实来源**：所有结论引用 `logs/230/*`、`database/migrations/<timestamp>_230_job_catalog_oper_fix.sql`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 等权威文件，未产生二次来源。

---

## 6. 参考资料

- `frontend/test-results/position-crud-full-lifecyc-b9f01-RUD生命周期-Step-1-创建职位-Create--chromium/`  
- `logs/219E/position-lifecycle-20251107-135246.log`  
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md`  
- `docs/api/openapi.yaml`

---

> 更新时间：2025-11-07。若后续阶段需要扩展，请在各子计划文档中更新，并同步本母计划的表格与状态。
