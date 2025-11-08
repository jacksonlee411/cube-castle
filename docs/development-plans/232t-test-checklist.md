# 232T – Playwright 复测清单（M2/M3）

> 目的：在 Plan 232 基础设施（T1/T2/T7）完成后，对六个 P0 场景进行 Chromium + Firefox 双浏览器验证，并为 Plan 219E §2.5 与 Plan 06 提供可追溯日志。

## 1. 前置条件与责任人登记

| 项目 | 验证命令 / 说明 | 责任人 | 完成时间 |
| --- | --- | --- | --- |
| 后端服务可用 | `make docker-up && make run-dev` → `curl localhost:9090/health` / `curl localhost:8090/health` |  |  |
| 前端 Dev Server 就绪 | `make frontend-dev` 或 `npm --prefix frontend run dev`，确认 `http://localhost:3000/organizations` 可访问 |  |  |
| JWT/租户信息 | `make jwt-dev-mint && eval $(make jwt-dev-export)` 或校验 `PW_JWT`, `PW_TENANT_ID` 环境变量 |  |  |
| Playwright 版本 | `npx playwright test --version` ≥ 1.56；`npm --prefix frontend ci` 已执行 |  |  |
| 等待/selector 基线 | `frontend/tests/e2e/utils/waitPatterns.ts` 与 Plan 232 T1 范围内的 testid 已合并到当前分支 |  |  |

> 若以上任一项未完成，请不要启动复测；先在表格中补齐责任人和完成时间，再进行第 2 章流程。

## 2. 执行顺序（记录实际完成情况）

| 步骤 | 命令 | 预计耗时 | 产物 | 备注 |
| --- | --- | --- | --- | --- |
| 1 | `npx playwright test tests/e2e/business-flow-e2e.spec.ts --project=chromium` | 8 min | `logs/219E/business-flow-e2e-chromium-<ts>.log` | 责任人/时间： | 
| 2 | `... --project=firefox` | 8 min | `logs/219E/business-flow-e2e-firefox-<ts>.log` | 责任人/时间： |
| 3 | `npx playwright test tests/e2e/job-catalog-secondary-navigation.spec.ts --project=chromium` | 5 min | `logs/219E/job-catalog-secondary-navigation-chromium-<ts>.log` | 若需播种数据，先执行 `scripts/dev/seed-job-catalog.sh` |
| 4 | `... --project=firefox` | 5 min | `logs/219E/job-catalog-secondary-navigation-firefox-<ts>.log` |  |
| 5 | `npx playwright test tests/e2e/position-tabs.spec.ts --project=chromium` | 4 min | `logs/219E/position-tabs-chromium-<ts>.log` | 可与步骤 7 并行，但日志需独立记录 |
| 6 | `... --project=firefox` | 4 min | `logs/219E/position-tabs-firefox-<ts>.log` |  |
| 7 | `npx playwright test tests/e2e/position-lifecycle.spec.ts --project=chromium` | 4 min | `logs/219E/position-lifecycle-chromium-<ts>.log` |  |
| 8 | `... --project=firefox` | 4 min | `logs/219E/position-lifecycle-firefox-<ts>.log` |  |
| 9 | `npx playwright test tests/e2e/temporal-management-integration.spec.ts --project=chromium` | 6 min | `logs/219E/temporal-management-integration-chromium-<ts>.log` | 后端不可用时需在日志首行注明 `E2E_MOCK_MODE=true` |
| 10 | `... --project=firefox` | 6 min | `logs/219E/temporal-management-integration-firefox-<ts>.log` |  |
| 11 | `npx playwright test tests/e2e/optimization-verification-e2e.spec.ts --project=chromium` | 6 min | `logs/219E/optimization-verification-e2e-chromium-<ts>.log` | 日志需记录 `totalSize` 与基线引用 |
| 12 | `... --project=firefox` | 6 min | `logs/219E/optimization-verification-e2e-firefox-<ts>.log` | 若 Firefox 收集指标受限，请在日志说明 |

> **命名规范**：`<ts>` 使用 `date +%Y%m%d%H%M%S`；同一场景的 Chromium/Firefox 日志必须在 PR 中引用。

## 3. 成功判定

1. Playwright CLI 返回 0，`frontend/test-results/.last-run.json` 中 `status === "passed"`。
2. `logs/219E/` 目录具备上述 12 个日志，文件首行说明：
   - 执行命令
   - 运行环境（Chromium / Firefox、PW_JWT 来源）
   - 是否命中 Mock 模式
3. 对于 optimization-verification：
   - 控制台打印 `前端资源总大小: X.XXKB`；
   - 小于 `5 * 1024 * 1024`，并在日志中写明对照 `docs/reference/03-API-AND-TOOLS-GUIDE.md#e2e-前端资源体积基线`。
4. Chrome/Firefox 若任一失败，需即刻收集 `frontend/playwright-report/*` 目录，记录 trace 链接并创建 Issue。

## 4. 复盘、证据与文档同步

1. **日志归档**：每条命令运行结束立即将 `playwright-report/index.html` 与 `data/` 留存到 `logs/219E/artifacts/{scenario}-{browser}-{ts}/`（便于审阅者复现）。
2. **Plan 232 状态**：在 `docs/development-plans/232-playwright-p0-stabilization.md` “当前状态”表中填写每个步骤的完成时间、责任人、日志路径。
3. **Plan 219E §2.5**：为六个场景分别添加 Chromium/Firefox 的通过时间与日志链接。
4. **Plan 06**：在 P0 验证章节引用 232T 的结果，说明阻塞项解除。
5. **性能阈值说明**：`optimization-verification-e2e` 日志需引用 `docs/reference/03-API-AND-TOOLS-GUIDE.md#e2e-前端资源体积基线`，并标注 measured size。

## 5. 风险提示

- **后台未启动**：所有场景都会在 `waitForGraphQL` 阶段超时；复测前务必确认 `docker ps` 中存在 `command-service`、`query-service`。
- **JWT 过期**：如果 `.cache/dev.jwt` 生成时间超过 8h，须重新 `make jwt-dev-mint`；否则 GraphQL 返回 401，表现与超时类似。
- **时间预算**：单浏览器一轮约 33 分钟；若资源有限，可按表格顺序分批执行，但必须在 24h 内补齐成对日志。

---

维护人：临时指定（请在实际执行者完成后更新）  
创建时间：2025-11-09 12:30 CST
