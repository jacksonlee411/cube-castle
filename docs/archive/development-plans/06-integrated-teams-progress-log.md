# [Archived] Plan 06 – 集成测试验证纪要（只读）

归档时间: 2025-11-15  
归档说明: 本文件已归档，仅保留历史记录，不再作为唯一事实来源（SSoT）。  
现行来源：
- P0 用例清单与门禁：`docs/archive/development-plans/232-playwright-p0-stabilization.md`、`docs/archive/development-plans/232t-test-checklist.md`
- 执行/进度汇总：`docs/development-plans/215-phase2-execution-log.md`
- 职位域回归与运行手册：`docs/development-plans/240E-position-regression-and-runbook.md`
- 2025-11-21 更新：Plan 232 已在 Chromium/Firefox 双端复测并锁定 `@playwright/test@1.56.1`，最新日志保存在 `logs/219E/position-tabs-{chromium-20251121121935,firefox-20251121121954}.log`、`logs/219E/position-lifecycle-{chromium-20251121122024,firefox-20251121122032}.log` 中；请以上述现行来源为准。

---

（以下为归档原文，仅供查阅）

# Plan 06 – 集成测试验证纪要（2025-11-08 10:30 CST）

## 1. 环境与前置校验
- `make docker-up && make run-dev`：PostgreSQL/Redis/REST/GraphQL 容器均处于 healthy，宿主机未占用 5432/6379。
- `go version` 输出 `go1.24.9`、`node --version` 输出 `v22.17.1`；`make db-migrate-all` 显示最新版本 `20251107123000` 已应用。
- `make jwt-dev-mint` 更新 `.cache/dev.jwt`，所有 Playwright/脚本通过 `PW_JWT`、`PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9` 注入。

## 2. 已执行验证
> 2025-11-21 提示：下表保留 11 月 7 日的历史结论，最新回归结果与日志路径已汇总到 `docs/archive/development-plans/232t-test-checklist.md`。
| 步骤 | 结果 | 证据 |
| --- | --- | --- |
| `npm run test:e2e -- --project=chromium tests/e2e/business-flow-e2e.spec.ts` | ❌ 删除阶段 `temporal-delete-record-button` 未出现 | `logs/219E/business-flow-e2e-chromium-20251107-133349.log` |
| `npm run test:e2e -- --project=firefox tests/e2e/business-flow-e2e.spec.ts` | ❌ 同上 | `logs/219E/business-flow-e2e-firefox-20251107-140221.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/job-catalog-secondary-navigation.spec.ts` | ❌ 未渲染“编辑职类信息”标题 | `logs/219E/job-catalog-secondary-navigation-chromium-20251107-133841.log` |
| `npm run test:e2e -- --project=firefox tests/e2e/job-catalog-secondary-navigation.spec.ts` | ❌ 同上 | `logs/219E/job-catalog-secondary-navigation-firefox-20251107-134321.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/name-validation-parentheses.spec.ts` | ✅ 2025-11-08 复测通过（补齐 JWT/租户请求头后 REST/GraphQL 均 200） | `logs/219E/name-validation-parentheses-20251108T052717Z.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-tabs.spec.ts` | ❌ `任职历史` 文案缺失 | `logs/219E/position-tabs-20251107-134806.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-lifecycle.spec.ts` | ❌ `position-detail-card` 未出现 | `logs/219E/position-lifecycle-20251107-135246.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/position-crud-full-lifecycle.spec.ts` | ✅ 完整 CRUD（Create→Delete），最新职位 `P1000031`，记录 RequestId | `logs/230/position-crud-playwright-20251108T102815.log`、`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` |
| `npm run test:e2e -- --project=chromium tests/e2e/temporal-management-integration.spec.ts` | ❌ 无法定位 `organization-dashboard` | `logs/219E/temporal-management-integration-20251107-135738.log` |
| `scripts/e2e/org-lifecycle-smoke.sh` | ✅ 完成创建/停用/启用/GraphQL 校验 | `logs/219E/org-lifecycle-smoke-20251107-140705.log` |
| `LOAD_DRIVER=node REQUEST_COUNT=40 CONCURRENCY=4 THROTTLE_DELAY_MS=30 scripts/perf/rest-benchmark.sh` | ✅ 获得 201/429 统计与延迟分布 | `logs/219E/rest-benchmark-20251107-140709.log` |
| `npx graphql-inspector diff docs/api/schema.graphql logs/graphql-snapshots/runtime-schema.graphql` | ✅ `No changes detected`，runtime SDL 经 `go run ./cmd/hrms-server/query/tools/dump-schema --out logs/graphql-snapshots/runtime-schema.graphql` 导出 | `logs/219T5/graphql-inspector-diff-20251108-015138.txt` |
| `scripts/diagnostics/check-job-catalog.sh` | ✅ `OPER` Job Catalog 通过（roles=1、levels=S1/S2/S3） | `logs/230/job-catalog-check-20251108T093645.log` |
| `scripts/dev/seed-position-crud.sh` | ✅ 创建/填充/空缺职位 `P1000027`，播种日志可复用 | `logs/230/position-seed-20251108T094735.log` |
| 240D 职位详情观测用例（Chromium） | ✅ 观测事件输出与落盘通过（hydrate/tab）；使用现有职位 `PW_POSITION_CODE` 模式 | `logs/plan240/D/obs-position-observability-chromium.log`、`frontend/playwright-report/index.html` |

## 3. 当前阻塞
（略；归档原文保持不变）
