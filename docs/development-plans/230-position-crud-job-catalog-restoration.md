# 230 – Position CRUD 参考数据修复计划

## 1. 背景与目标
- 219T 回归中，`tests/e2e/position-crud-full-lifecycle.spec.ts` 在 Step 1 调用 `POST /api/v1/positions` 返回 422，错误体为 `JOB_CATALOG_NOT_FOUND: JobFamilyGroup OPER is inactive or missing`，请求日志编号 `requestId=8a8739b3-8580-4da4-b503-9367560055b1`（产物：`frontend/test-results/position-crud-full-lifecyc-b9f01-RUD生命周期-Step-1-创建职位-Create--chromium/`）。
- 同步 CLI 复现显示相同响应，确认命令服务在 219E 环境缺失或禁用了 `OPER` 系列 Job Family 数据（`curl -X POST http://localhost:9090/api/v1/positions ...` 返回 422）。
- 该缺口阻断职位 CRUD 验收流程，但不属于 219T1 读模型范畴，特立 230 号计划专注恢复 Job Catalog 参考数据，使 Playwright/REST 测试能在真实数据上通过。

## 2. 范围
1. 对照 `database/migrations/` 与现有 PostgreSQL 中的 `job_family_groups / job_families / job_roles / job_levels` 表，确认 `OPER` 及其关联实体缺失或 status=INACTIVE 的原因。
2. 制定一次性数据修复方案（SQL migration 或数据回灌脚本），确保 219E/开发环境具备完整的 `OPER` 层级。
3. 为命令服务新增自检脚本，例如 `scripts/diagnostics/check-job-catalog.sh`，在 `make status` 中暴露缺失项。
4. 复跑 `tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium 即可），并更新 219T 报告的 Position CRUD 条目。

## 3. 行动项
| 步骤 | 负责人 | 输出 |
| --- | --- | --- |
| 3.1 导出/查询 Job Catalog 现状：`psql -c "SELECT code, status FROM job_family_groups WHERE code IN ('OPER');"`，并截取结果 | DB | `logs/230/job-catalog-audit-YYYYMMDD.log` |
| 3.2 审核迁移历史，定位 `OPER` 被删除/禁用的 commit，与 219E 期望对比 | Backend | 结论记录到 `logs/230/root-cause.md` |
| 3.3 编写修复迁移或数据脚本，重建 `OPER` JobFamilyGroup、`OPER-OPS` JobFamily、`OPER-OPS-MGR` JobRole 及必要的等级（S1-S3）；脚本位于 `database/migrations/230_job_catalog_oper_fix.sql`（示例） | DB | PR + SQL 文件 |
| 3.4 为 `make status` 增加 Job Catalog 健康检查（未通过时提示运行 230 脚本） | DevEx | `scripts/diagnostics/check-job-catalog.sh` + Makefile 更新 |
| 3.5 运行 Playwright：`cd frontend && npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1`，归档产物 | QA | `frontend/test-results/position-crud-full-lifecyc-*_chromium/` |
| 3.6 回填 `docs/development-plans/219T-e2e-validation-report.md` 与 230 计划状态，标记 Position CRUD 已解除阻塞 | QA | 文档更新 |
| 3.7 职位管理功能对齐验证：在测试前核对实现范围与测试用例预期（含 UI、REST） | QA + Frontend | `logs/230/position-module-readiness.md` |

## 3A. 任务拆解与检查点
### 3A.1 Job Catalog 现状取证（对应 3.1）
- **准备**：先运行 `make docker-up` 启动 Docker Compose（遵循 `AGENTS.md` 的容器优先约束），确认 `docker compose -f docker-compose.dev.yml ps postgres` 状态为 `healthy`；如需联动其他服务，可在此基础上执行 `make run-dev`。
- **取证命令**：运行  
  ```bash
  docker compose -f docker-compose.dev.yml exec postgres psql -U user -d cubecastle \
    -c "SELECT 'job_family_groups' AS table, code, status FROM job_family_groups WHERE code='OPER' \
        UNION ALL SELECT 'job_families', code, status FROM job_families WHERE code LIKE 'OPER%';"
  docker compose -f docker-compose.dev.yml exec postgres psql -U user -d cubecastle \
    -c "SELECT code, status FROM job_roles WHERE code LIKE 'OPER%' ORDER BY code;"
  docker compose -f docker-compose.dev.yml exec postgres psql -U user -d cubecastle \
    -c "SELECT role_code, level_code, status FROM job_levels WHERE role_code LIKE 'OPER%' ORDER BY role_code, level_code;"
  ```
- **产物**：将标准输出/错误重定向到 `logs/230/job-catalog-audit-YYYYMMDD.log`，并记录执行的 Git commit 与 docker 镜像标签，方便回溯。
- **检查点**：日志中若任何行缺失或 `status != 'ACTIVE'`，需立即阻断后续步骤并进入 3A.2。

### 3A.2 迁移历史审计（对应 3.2）
- **定位变更**：使用 `rg -n "OPER" database/migrations` 查找涉及 OPER 的迁移脚本，再结合 `git log -p -- database/migrations/**/*oper*.sql` 确认最近对 Job Catalog 的修改。
- **与 219E 期望对比**：对照 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 中记录的参考组织结构，列出缺失字段（如 `name`, `description`, `level`）。必要时参考 `database/migrations/*seed*.sql` 的插入语句。
- **文档输出**：在 `logs/230/root-cause.md` 中记录：时间、操作者、涉及的 commit/PR、影响表、预期与实际差异、是否需要回滚或新增修复。

### 3A.3 修复迁移与数据脚本（对应 3.3）
- **脚本结构**：在 `database/migrations/` 下新增 `230_job_catalog_oper_fix.sql`（以 goose/atlas 要求命名），包含：
  1. 幂等检查（`DO $$ BEGIN IF NOT EXISTS(...) THEN ... END IF; END $$;`）确保重复执行不会插入重复记录；
  2. 对 `job_family_groups`、`job_families`、`job_roles`、`job_levels` 的插入或 `UPDATE ... SET status='ACTIVE'`；
  3. 回滚段（若工具支持 `-- +goose Down`），清理 230 引入的数据。
- **本地验证**：运行 `make db-migrate-all`，随后重复 3A.1 的 SQL，确认日志中所有 `OPER` 记录为 `ACTIVE`。
- **审阅要点**：PR 需附带 `logs/230/job-catalog-audit-YYYYMMDD.log`、`logs/230/root-cause.md` 片段与 `git diff database/migrations/`，便于 Reviewer 核对。

### 3A.4 Job Catalog 自检脚本（对应 3.4）
- **脚本内容**：`scripts/diagnostics/check-job-catalog.sh` 以 `bash` 编写，执行 `psql -Atqc "SELECT COUNT(*) FROM job_roles WHERE code LIKE 'OPER%'"` 等检查，若计数为 0 或存在非 ACTIVE 状态则 `exit 1` 并输出修复指引（示例：“检测到 OPER Job Role 缺失，请运行 database/migrations/230_job_catalog_oper_fix.sql”）。
- **Make 集成**：在 `Makefile` 的 `status` 目标中追加 `bash scripts/diagnostics/check-job-catalog.sh`，与现有健康检查保持同级；脚本失败时 `make status` 必须整体失败。
- **复用**：脚本应允许 `JOB_CATALOG_CODES=OPER,FINANCE` 这样的环境变量以便未来扩展。

### 3A.5 Playwright 回归与产物归档（对应 3.5）
- **准备**：确保 `.cache/dev.jwt` 存在（若缺失，运行 `make jwt-dev-setup && make jwt-dev-mint`），并导出 `PW_TENANT_ID`、`PW_JWT` 到 shell。
- **执行**：`cd frontend && npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1 --reporter=line,junit`。将 `playwright-report` 与 `test-results` 目录复制/重命名为 `frontend/test-results/position-crud-full-lifecyc-<commit>-chromium/`。
- **检查点**：测试日志中 Step 1 不再出现 422/`JOB_CATALOG_NOT_FOUND`，Junit 报告全部通过。必要时附加 `curl -X POST http://localhost:9090/api/v1/positions ...` 的成功响应截图。
- **执行记录（2025-11-07 14:54 CST）**：`npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1 --reporter=line` 9 步骤全部通过，日志 `logs/230/position-crud-playwright-20251107T065443Z.log`，产物归档 `frontend/test-results/position-crud-full-lifecyc-99d5ffcf-chromium/`（内含 `playwright-report` 与原始 `test-results`）。RequestId 记录：Step1 Create `17671b78-1372-4abf-9fdd-e00f37b5decc`、Step3 Update `a173b7c3-7330-4bea-a2fa-0dac1d18e859`、Step4 Fill `4272d7fd-949d-4da5-8989-067291ccb0da`、Step5 Vacate `f6a80f66-5e44-43d1-b2f3-2308a5776684`、Step6 Delete `7393b8ff-3f0c-4c5b-a9c3-9310199a00d3`。

### 3A.6 文档同步（对应 3.6）
- **操作**：更新 `docs/development-plans/219T-e2e-validation-report.md` 的 Position CRUD 条目（记录修复脚本版本、Playwright 跑批时间、产物路径）；同步更新本计划的“验收标准”小节中的证据链接。
- **交付格式**：在 PR 描述中引用新的日志、Playwright 产物与 `make status` 输出；若 219T 报告使用表格，需附加“阻塞解除时间”列。
- **完成信号**：文档 MR / PR 获批且主干存在上述证据，方可将 230 计划标记为完成。

### 3A.7 职位管理功能对齐验证（对应 3.7）
- **功能基线**：对照 `docs/api/openapi.yaml` 与 `frontend/src/features/positions` 中的实现，列出当前迭代已交付/未交付的 Position Management 功能（创建、编辑、权限校验、版本控制等），在 `logs/230/position-module-readiness.md` 中形成表格。
- **测试映射**：为 `tests/e2e/position-crud-full-lifecycle.spec.ts` 及关联 Playwright/REST 脚本建立“功能 → testcase”映射，若发现测试覆盖到尚未交付的功能，需在测试中显式标注 `// TODO-TEMPORARY` 并调整断言（或在计划中登记上线时间）。
- **对齐动作**：若功能差异源于后端缺陷，先确认是否纳入 230 范围；否则在 219T 追踪表中登记，避免过度断言导致“功能不完整”误报。完成后将对齐记录附在 PR/评审说明中。

## 4. 依赖与前置
- 需有权访问 Docker PostgreSQL 容器（`make run-dev`），并允许执行迁移。
- Job Catalog 数据可能被其他团队使用，修复前需与 HR 领域负责人确认目标状态。
- Playwright 运行前需确保 `npm install`、`.cache/dev.jwt`、`PW_TENANT_ID` 等已配置。

## 5. 验收标准
1. `psql` 查询结果显示 `OPER` 及其子项 `status='ACTIVE'`，且具备预期字段（code、name、level）。
2. 修复脚本/迁移可重复执行（幂等），并在 CI 中通过 `make db-migrate-all`。
3. `npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1` 全部步骤通过；`frontend/test-results/` 中无新的 422 错误。
4. `make status` 若检测到 Job Catalog 缺口会报错；修复后命令返回 OK。
5. 219T 报告中 Position CRUD 条目更新为 “通过”，引用新的日志与测试产物。

---

> 单一事实来源：`frontend/test-results/position-crud-full-lifecyc-99d5ffcf-chromium/`、`logs/230/position-crud-playwright-20251107T065443Z.log`、`logs/219E/org-lifecycle-20251107-083118.log`、`curl` CLI 复现日志。  
> 更新时间：2025-11-07。  
