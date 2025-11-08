# Plan 06 – 集成测试验证纪要（2025-11-07 14:35 CST）

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
| `npm run test:e2e -- --project=chromium tests/e2e/position-crud-full-lifecycle.spec.ts` | ✅ 完整 CRUD，生成职位 `P1000017` | `logs/219E/position-crud-full-lifecycle-20251107-135724.log` |
| `npm run test:e2e -- --project=chromium tests/e2e/temporal-management-integration.spec.ts` | ❌ 无法定位 `organization-dashboard` | `logs/219E/temporal-management-integration-20251107-135738.log` |
| `scripts/e2e/org-lifecycle-smoke.sh` | ✅ 完成创建/停用/启用/GraphQL 校验 | `logs/219E/org-lifecycle-smoke-20251107-140705.log` |
| `LOAD_DRIVER=node REQUEST_COUNT=40 CONCURRENCY=4 THROTTLE_DELAY_MS=30 scripts/perf/rest-benchmark.sh` | ✅ 获得 201/429 统计与延迟分布 | `logs/219E/rest-benchmark-20251107-140709.log` |
| `npx graphql-inspector diff docs/api/schema.graphql logs/graphql-snapshots/runtime-schema.graphql` | ✅ `No changes detected`，runtime SDL 经 `go run ./cmd/hrms-server/query/tools/dump-schema --out logs/graphql-snapshots/runtime-schema.graphql` 导出 | `logs/219T5/graphql-inspector-diff-20251108-015138.txt` |

## 3. 当前阻塞
1. **Playwright 219T3 多场景失败**  
   - `business-flow-e2e`：Temporal 删除按钮缺失。  
   - `job-catalog-secondary-navigation`：Chromium/Firefox 均缺少“编辑职类信息”。  
   - `position-tabs`、`position-lifecycle`：fixture 数据未渲染六个页签及生命周期文案。  
   - `temporal-management-integration`：`waitForOrganizationSearchInput` 无法找到 `organization-dashboard`。  
   → 需恢复前端依赖数据或 UI 逻辑，重新收集 trace/video。
2. **证据回填未完成**  
   - REST Benchmark 摘要尚未写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md`。  
   - `docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 未记录本轮执行结果及日志链接。

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
