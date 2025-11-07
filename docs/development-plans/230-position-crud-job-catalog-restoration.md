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

> 单一事实来源：`frontend/test-results/position-crud-full-lifecyc-b9f01-*/error-context.md`、`logs/219E/org-lifecycle-20251107-083118.log`、`curl` CLI 复现日志。  
> 更新时间：2025-11-07。  
