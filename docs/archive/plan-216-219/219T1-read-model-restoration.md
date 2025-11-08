# 219T1 – GraphQL 读模型修复子方案

## 1. 背景与目标
- 219T 报告指出：`logs/219E/org-lifecycle-20251107-073626.log` 中 REST 创建的组织无法在 GraphQL 查询中读到，`organization` 返回 `null`。
- 219E 验收依赖“命令→查询”一致性；读模型异常会导致业务流、Playwright 与回退演练全部失效。
- 本子方案聚焦恢复查询服务对 PostgreSQL 最新数据的消费与投影。

## 2. 范围
1. 诊断 `cmd/hrms-server/query` 消费链路（Temporal worker、CDC、物化视图脚本），并记录复制延迟、slot 位点、任务重试等指标。
2. 排查 Docker 环境（`docker-compose.dev.yml`）中 temporal/worker/graphQL 容器日志，确保诊断日志归档到 `logs/219E/BLOCKERS-*`。
3. 针对发现的问题实现修复（订阅滞后、投影表 schema 漂移、重放失败等），并补充单元/集成测试。
4. 对现有读模型执行一次性回放/重建（如刷新物化视图、重放 CDC backlog），确保历史组织与增量数据一致。
5. 在 `logs/219E/` 中生成新的冒烟日志，证明 GraphQL 可读取刚创建的组织，并复测至少一个依赖真实读模型的 Playwright 场景。
6. 将修复细节回填至 219T 报告与相关计划文档。

## 3. 行动项
| 步骤 | 负责人 | 输出 |
| --- | --- | --- |
| 3.1 收集容器日志并抓取 CDC/Temporal 指标：`docker compose -f docker-compose.dev.yml logs -f graphql-service rest-service`（若 Temporal/worker 运行在其它 compose，追加其日志）+ `psql -c "select slot_name, confirmed_flush_lsn, active from pg_replication_slots where plugin = 'wal2json';"` | Backend | 故障堆栈与复制位点记录（存入 `logs/219E/BLOCKERS-*`） |
| 3.2 对比 `database/migrations` 与查询层 schema，确认物化视图/索引未漂移 | DB | 校验笔记（含差异截图或 `psql \d` 输出） |
| 3.3 修复消费链路（SQL/Temporal/配置），并补充针对 `cmd/hrms-server/query` 的单元或集成测试 | Backend | PR + 变更说明 + 测试结果 |
| 3.4 运行投影回放：重建受影响物化视图或重放 CDC backlog（记录执行命令与成功标记） | Backend | `logs/219E/read-model-replay-YYYYMMDD.log` |
| 3.5 运行 `scripts/e2e/org-lifecycle-smoke.sh` + 手工 GraphQL 查询，记录 `logs/219E/org-lifecycle-YYYYMMDD-hhmmss.log` 新版本 | QA | 新日志（含 `success: true`） |
| 3.6 复测任一依赖实时读模型的 Playwright 用例（如 `tests/e2e/position-crud-full-lifecycle.spec.ts`），确保读取到新组织 | QA | `frontend/test-results/*` 新产物 + 结论 |
| 3.7 更新 `docs/development-plans/219T-e2e-validation-report.md` “下一步” 章节状态并回填修复摘要 | QA | 文档 |

## 4. 依赖与前置
- Docker 环境必须保持 `make run-dev` 状态，Temporal/Redis/Postgres 容器运行正常，并允许访问 `psql`。
- 需读取 `internal/query` 与 Temporal worker 源码，以及 CDC/物化视图脚本；必要时与 219D 团队同步。
- 可访问 Playwright 依赖的 GraphQL stub/真实模式配置，以便复测。

## 5. 验收标准
1. 诊断阶段提交的日志/指标展示出问题根因（例如 CDC slot 堵塞、Temporal worker 异常），并上传至 `logs/219E/BLOCKERS-*`。
2. REST 创建的新组织，经 GraphQL `organization` 查询可立即读取到，`logs/219E/org-lifecycle-*.log` 中包含 `success: true` 与正确字段。
3. `scripts/e2e/org-lifecycle-smoke.sh` 与至少一个 Playwright 场景在真实读模型下通过，测试产物归档。
4. 读模型投影表已执行回放或刷新操作，提供执行记录与确认无 backlog。
5. 修复细节、测试结果与后续步骤已同步至 219T 报告及相关计划文档。

## 6. 进展纪要（2025-11-07）
- `scripts/e2e/org-lifecycle-smoke.sh` 于 08:31 CST / 00:37 UTC 各执行一轮，对应日志 `logs/219E/org-lifecycle-20251107-083118.log`、`logs/219E/org-lifecycle-20251107-003713.log`，GraphQL 查询均返回 `success: true`。
- Playwright 复测命令：
  - `cd frontend && npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1`，产物位于 `frontend/test-results/position-crud-full-lifecyc-b9f01-*/`。该用例的 Job Catalog 参考数据缺口已转交 230 号计划处理。
  - `cd frontend && PW_JWT=$(cat ../.cache/dev.jwt) PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 npx playwright test tests/e2e/basic-functionality-test.spec.ts --project=chromium --workers=1 --reporter=line`，生成 `frontend/test-results/{app-loaded.png,organizations-page.png,interaction-test.png,error-handling.png}`，验证真实 GraphQL 读模型下的组织仪表板可正常加载。测试报告：`frontend/playwright-report/index.html`。

---

> 单一事实来源：`logs/219E/`、`docker-compose.dev.yml`、`cmd/hrms-server/query/`。  
> 更新时间：2025-11-07。
