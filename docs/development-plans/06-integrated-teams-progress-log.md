# Plan 06 – 集成测试验证纪要（2025-11-08 10:30 CST）

## 1. 环境与前置校验
- `make docker-up && make run-dev`：PostgreSQL/Redis/REST/GraphQL 容器均处于 healthy，宿主机未占用 5432/6379/7233。
- `go version` 输出 `go1.24.9`、`node --version` 输出 `v22.17.1`；`make db-migrate-all` 显示最新版本 `20251107123000` 已应用。
- `make jwt-dev-mint` 更新 `.cache/dev.jwt`，所有 Playwright/脚本通过 `PW_JWT`、`PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9` 注入。

## 2. 已执行验证
| 步骤 | 结果 | 证据 |
| --- | --- | --- |
| `npm run test:e2e -- --project=chromium tests/e2e/business-flow-e2e.spec.ts` | ❌ 删除阶段 `temporal-delete-record-button` 未出现 | `logs/219E/business-flow-e2e-chromium-20251107-133349.log` |
| `npm run test:e2e -- --project=firefox tests/e2e/business-flow-e2e.spec.ts` | ❌ 同上 | `logs/219E/business-flow-e2e-firefox-20251107-140221.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/job-catalog-secondary-navigation.spec.ts` | ❌ 未渲染“编辑职类信息”标题 | `logs/219E/job-catalog-secondary-navigation-chromium-20251107-133841.log` |
| `npm run test:e2e -- --project=firefox tests/e2e/job-catalog-secondary-navigation.spec.ts` | ❌ 同上 | `logs/219E/job-catalog-secondary-navigation-firefox-20251107-134321.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/name-validation-parentheses.spec.ts` | ✅ REST/GraphQL 均成功，带 `requestId` | `logs/219E/name-validation-parentheses-20251107-134801.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-tabs.spec.ts` | ❌ `任职历史` 文案缺失 | `logs/219E/position-tabs-20251107-134806.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-lifecycle.spec.ts` | ❌ `position-detail-card` 未出现 | `logs/219E/position-lifecycle-20251107-135246.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-crud-full-lifecycle.spec.ts` | ✅ 完整 CRUD（Create→Delete），最新职位 `P1000031`，记录 RequestId | `logs/230/position-crud-playwright-20251108T102815.log`、`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` |
| `npm run test:e2e -- --project=chromium tests/e2e/temporal-management-integration.spec.ts` | ❌ 无法定位 `organization-dashboard` | `logs/219E/temporal-management-integration-20251107-135738.log` |
| `scripts/e2e/org-lifecycle-smoke.sh` | ✅ 完成创建/停用/启用/GraphQL 校验 | `logs/219E/org-lifecycle-smoke-20251107-140705.log` |
| `LOAD_DRIVER=node REQUEST_COUNT=40 CONCURRENCY=4 THROTTLE_DELAY_MS=30 scripts/perf/rest-benchmark.sh` | ✅ 获得 201/429 统计与延迟分布 | `logs/219E/rest-benchmark-20251107-140709.log` |
| `npx graphql-inspector diff docs/api/schema.graphql logs/graphql-snapshots/runtime-schema.graphql` | ✅ `No changes detected`，runtime SDL 经 `go run ./cmd/hrms-server/query/tools/dump-schema --out logs/graphql-snapshots/runtime-schema.graphql` 导出 | `logs/219T5/graphql-inspector-diff-20251108-015138.txt` |
| `scripts/diagnostics/check-job-catalog.sh` | ✅ `OPER` Job Catalog 通过（roles=1、levels=S1/S2/S3） | `logs/230/job-catalog-check-20251108T093645.log` |
| `scripts/dev/seed-position-crud.sh` | ✅ 创建/填充/空缺职位 `P1000027`，播种日志可复用 | `logs/230/position-seed-20251108T094735.log` |

## 3. 当前阻塞
1. **Playwright P0 场景仍需修复**  
   - `business-flow-e2e`：Temporal 删除按钮缺失。  
   - `job-catalog-secondary-navigation`：Chromium/Firefox 均缺少“编辑职类信息”。  
   - `position-tabs`、`position-lifecycle`：需在最新 Job Catalog 数据下重新执行，验证 UI/data-testid 是否仍异常。  
   - `temporal-management-integration`：`organization-dashboard` 仍无法加载。  
   → Position CRUD 已由 `logs/230/position-crud-playwright-20251108T102815.log` 验证通过，但其余 P0 仍需 UI/数据联调。
2. **文档与性能摘要待回填**  
   - REST Benchmark JSON 摘要尚未写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md`。  
   - `docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 需追加 2025-11-08 的 Job Catalog/Position CRUD 证据。  
   - 219E 回退/Outbox 指标仍缺可引用的日志条目。

> 说明：GraphQL diff 阻塞已在 2025-11-08 通过 gqlgen runtime SDL 快照 + GraphQL Inspector 验证解除，详见上表与日志 `logs/219T5/graphql-inspector-diff-20251108-015138.txt`。

## 4. 待办清单
| 优先级 | 待办 | 说明 |
| --- | --- | --- |
| P0 | ✅ GraphQL 运行时已切换至 gqlgen，`graphql-inspector diff` 与 runtime SDL 快照无差（`logs/219T5/graphql-inspector-diff-20251108-015138.txt`） | Plan 06 第 3 节硬门槛已解除 |
| P0 | 恢复 business-flow/job-catalog/position-tabs/position-lifecycle/temporal-management 场景所需数据，Chromium 与 Firefox 全绿 | 满足退出准则第 1 条 |
| P1 | 将 `logs/219E/rest-benchmark-20251107-140709.log` JSON 摘录写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md` | 补全性能证据 |
| P1 | 在 `docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 回填本轮执行与失败原因 | 保持唯一事实来源 |
| P2 | 更新 `frontend/test-results/app-loaded.png` 与最新 screenshots/trace/video 路径 | 对齐 Plan 06 §4 要求 |

## 5. 退出准则复核
- **Chromium/Firefox Playwright 全绿**：未满足（多场景失败）。  
- **GraphQL 契约 diff**：已通过 `npx graphql-inspector diff docs/api/schema.graphql logs/graphql-snapshots/runtime-schema.graphql`（`logs/219T5/graphql-inspector-diff-20251108-015138.txt`）。  
- **REST/性能脚本证据**：脚本已执行，但尚未写入参考文档。  
- **文档回填**：`docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 均需更新当前日志/结论。

> 结论：Plan 06 暂无法关闭，需完成上述 P0/P1 待办后重新评估。

## 6. 219E 重启前置条件推进（2025-11-08 10:30 CST）

| 项目 | Owner | 状态 | 说明 | 证据 |
| --- | --- | --- | --- | --- |
| 219E 文档更新（阻塞列表 + 前置条件表） | Codex + QA | ✅ 完成 | `docs/development-plans/219E-e2e-validation.md` 已记录 Docker 权限解除、Playwright/性能/回退等前置事项及日志来源 | `docs/development-plans/219E-e2e-validation.md` |
| Playwright P0 场景修复（business-flow、job-catalog、position-tabs、temporal-management） | 前端团队 | ⏳ 进行中 | 需恢复缺失的 data-testid、UI 文案与数据，完成后回填 `logs/219E/*.log` 与 `frontend/test-results/*` | `logs/219E/business-flow-e2e-*.log`、`logs/219E/job-catalog-secondary-navigation-*.log` |
| Position/Assignment 数据链路恢复 | 命令 + 查询团队 | ✅ 完成 | 230B/C/D 已交付 Job Catalog 迁移、自检脚本与播种 + Playwright 复验：`scripts/diagnostics/check-job-catalog.sh`、`scripts/dev/seed-position-crud.sh`、`npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium）。现可据此解锁 `position-lifecycle`/`organization-validator` 套件。 | `logs/230/job-catalog-check-20251108T093645.log`、`logs/230/position-seed-20251108T094735.log`、`logs/230/position-crud-playwright-20251108T102815.log` |
| 性能基准回填（REST/GraphQL） | QA + SRE | ⏳ 待记录 | 借助 Node 驱动日志撰写对比并更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、219T 报告 | `logs/219E/rest-benchmark-20251107-140709.log`、`docs/development-plans/219T-e2e-validation-report.md:21-33` |
| 回退演练脚本与记录 | SRE + 后端 | ⏳ 待安排 | 依照 219D1/219D5 指南执行一次全量回退并归档日志，作为 219E 验收资料 | `logs/219D4/FAULT-INJECTION-2025-11-06.md`、`docs/development-plans/219D5-scheduler-docs.md` |
