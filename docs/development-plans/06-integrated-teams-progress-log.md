# Plan 06 – 集成测试验证要求

> 唯一事实来源：`docker-compose.dev.yml`、`docs/development-plans/219T-e2e-validation-report.md`、`frontend/tests/e2e/*`。  
> 本文用于指导 219 系列在恢复 Docker 环境后如何执行端到端验证、收集证据并回填报告。

## 1. 环境前提（所有验证必须满足）
- **Docker 优先**：执行 `make docker-up && make run-dev`，确保 PostgreSQL/Redis/Temporal 仅在容器内运行，宿主机不得占用 5432/6379/7233。
- **Go/Node 版本**：`go version` 输出 `go1.24.x`，`node --version` 与仓库 `.nvmrc` 一致；若版本不符，禁止继续验证。
- **JWT 与租户**：执行 `make jwt-dev-mint`，在 `.cache/dev.jwt` 中获取令牌；`DEFAULT_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9` 必须写入环境变量或 Playwright 配置。
- **数据基线**：运行 `make db-migrate-all` 后，确认 Job Catalog 参考数据完整（`OPER` 系列 Active），否则职位 CRUD 场景需 `test.skip` 并在 230 计划中补数。

## 2. Playwright 验证要求（219T3）
| 场景 | 验证步骤 | 产物要求 |
| --- | --- | --- |
| business-flow-e2e | `npm run test:e2e -- --project=chromium tests/e2e/business-flow-e2e.spec.ts`；运行日志需展示删除阶段等待 `temporal-timeline` 完成 | `frontend/test-results/business-flow-e2e-*/trace.zip`、`run-playwright.log` |
| job-catalog-secondary-navigation | 指定 Chromium/Firefox 双浏览器运行，截图必须包含“编辑职类信息”标题 | `frontend/test-results/job-catalog-secondary-navigation-*/` |
| name-validation-parentheses | REST PUT 返回 200 且 GraphQL 二次读取验证新名称；日志需附 `requestId` | `logs/219E/name-validation-*.log` |
| position-tabs / position-lifecycle | 通过 `POSITION_FIXTURE_CODE` + GraphQL stub 渲染；验证 `position-temporal-page`、六个页签截图 | `frontend/test-results/position-tabs-*`, `frontend/test-results/position-lifecycle-*` |
| position-crud-full-lifecycle | 若 Job Catalog 参考数据存在，则完整跑 Create→Delete；若返回 422，必须在报告中引用 `test.skip` 原因并链接 Plan 230 | `frontend/test-results/position-crud-full-lifecycle-*`, `logs/219E/position-crud-*.log` |
| temporal-management-integration | 使用 `waitForOrganizationSearchInput` 辅助器定位搜索框，需录制导航至 `/organizations/{code}/temporal` 的视频 | `frontend/test-results/temporal-management-integration-*/video.webm` |

执行顺序建议：先跑 `npm run test:e2e -- --project=chromium`, 通过后再追加 `--project=firefox`；任何失败必须在 `docs/development-plans/219T-e2e-validation-report.md` “Playwright 用例整改”章节附上原因与日志。

## 3. REST/GraphQL 验证要求
1. **组织冒烟**：运行 `scripts/e2e/org-lifecycle-smoke.sh`，日志写入 `logs/219E/org-lifecycle-YYYYMMDD-HHMMSS.log`，截图 `frontend/test-results/app-loaded.png` 需重新生成。
2. **性能基准**：执行 `LOAD_DRIVER=node REQUEST_COUNT=40 CONCURRENCY=4 THROTTLE_DELAY_MS=30 scripts/perf/rest-benchmark.sh`，将 JSON Summary 摘要粘贴至 `docs/reference/03-API-AND-TOOLS-GUIDE.md`。
3. **GraphQL 合约**：`curl http://localhost:8090/health` 返回 200 后，使用 `npx graphql-inspector diff docs/api/schema.graphql http://localhost:8090/graphql`，无 diff 方可继续。

## 4. 证据与回填
- **日志目录**：所有脚本输出集中到 `logs/219E/`，命名规则 `scenario-YYYYMMDD-HHMMSS.log`。
- **测试产物**：`frontend/test-results/` 与 `frontend/playwright-report/` 需保存最新一次运行的截图/trace/video，不得覆盖历史记录。
- **报告同步**：
  1. `docs/development-plans/219T-e2e-validation-report.md`：更新失败表、添加新的“脚本整改”条目与链接。
  2. `docs/development-plans/219E-e2e-validation.md`：在“测试范围”和“证据”表格中引用上述日志与产物。
  3. 若某场景因参考数据缺失被跳过，必须注明关联计划（如 230）与预计恢复时间。

## 5. 退出准则
仅当以下条件全部满足时，Plan 06 的验证阶段可视为完成：
1. Chromium 与 Firefox 上 `npm run test:e2e` 成功（允许被 Plan 230 接管的职位 CRUD 场景跳过，并在报告中注明）。
2. `scripts/e2e/org-lifecycle-smoke.sh` 与 `scripts/perf/rest-benchmark.sh` 最新日志附在 219T/219E 文档中，且 `make status` 显示健康。
3. `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 均已回填本次验证的产物链接与结论。
