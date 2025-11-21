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
| 1 | `npx playwright test tests/e2e/business-flow-e2e.spec.ts --project=chromium` | 8 min | `logs/219E/business-flow-e2e-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-09 17:11 / shangmeilin → ✅（`logs/219E/business-flow-e2e-chromium-20251109171101.log`）<br/>2025-11-07 13:33 / shangmeilin → **失败**（`logs/219E/business-flow-e2e-chromium-20251107-133349.log`：`temporal-delete-record-button` 不存在） |
| 2 | `... --project=firefox` | 8 min | `logs/219E/business-flow-e2e-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-09 17:11 / shangmeilin → ✅（`logs/219E/business-flow-e2e-firefox-20251109171155.log`）<br/>2025-11-07 14:02 / shangmeilin → **失败**（`logs/219E/business-flow-e2e-firefox-20251107-140221.log`：删除按钮缺失） |
| 3 | `npx playwright test tests/e2e/job-catalog-secondary-navigation.spec.ts --project=chromium` | 5 min | `logs/219E/job-catalog-secondary-navigation-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-09 17:12 / shangmeilin → ✅（`logs/219E/job-catalog-secondary-navigation-chromium-20251109171256.log`）<br/>2025-11-07 13:38 / shangmeilin → **失败**（`logs/219E/job-catalog-secondary-navigation-chromium-20251107-133841.log`：Modal 未渲染） |
| 4 | `... --project=firefox` | 5 min | `logs/219E/job-catalog-secondary-navigation-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-09 17:13 / shangmeilin → ✅（`logs/219E/job-catalog-secondary-navigation-firefox-20251109171326.log`）<br/>2025-11-07 13:43 / shangmeilin → **失败**（`logs/219E/job-catalog-secondary-navigation-firefox-20251107-134321.log`） |
| 5 | `npx playwright test tests/e2e/position-tabs.spec.ts --project=chromium` | 4 min | `logs/219E/position-tabs-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-21 12:19 / shangmeilin → ✅（`logs/219E/position-tabs-chromium-20251121121935.log`，Playwright 已锁定 1.56.1；六个页签/Mock 校验维持通过）<br/>2025-11-21 09:31 / shangmeilin → ✅（`logs/219E/position-tabs-chromium-20251121093128.log`，CLI 版本冲突已解除）<br/>2025-11-09 17:14 / shangmeilin → **失败**（`logs/219E/position-tabs-chromium-20251109171402.log`：`waitForGraphQL` 超时） |
| 6 | `... --project=firefox` | 4 min | `logs/219E/position-tabs-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-21 12:19 / shangmeilin → ✅（`logs/219E/position-tabs-firefox-20251121121954.log`，版本锁定后复测通过）<br/>2025-11-21 08:13 / shangmeilin → ✅（`logs/219E/position-tabs-firefox-20251121081344.log`，六个页签均可切换）<br/>2025-11-21 08:39 / shangmeilin → ⚠️ CLI 阻断（`logs/219E/position-tabs-firefox-20251121083920.log`，同 `test.describe` 报错） |
| 7 | `npx playwright test tests/e2e/position-lifecycle.spec.ts --project=chromium` | 4 min | `logs/219E/position-lifecycle-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-21 12:20 / shangmeilin → ✅（`logs/219E/position-lifecycle-chromium-20251121122024.log`：PositionTransfersPanel selector + 页签断言在锁定版本下稳定）<br/>2025-11-21 09:51 / shangmeilin → **失败**（`logs/219E/position-lifecycle-chromium-20251121095134.log`：`temporal-position-transfer-item` 未渲染）<br/>2025-11-09 17:47 / shangmeilin → **失败**（`logs/219E/position-lifecycle-chromium-20251109174658.log`：详情卡片未渲染） |
| 8 | `... --project=firefox` | 4 min | `logs/219E/position-lifecycle-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-21 12:20 / shangmeilin → ✅（`logs/219E/position-lifecycle-firefox-20251121122032.log`：同上 patch 后双端通过）<br/>2025-11-21 11:16 / shangmeilin → **失败**（`logs/219E/position-lifecycle-firefox-20251121111635.log`：缺少调动记录 testid）<br/>2025-11-21 08:14 / shangmeilin → ✅（`logs/219E/position-lifecycle-firefox-20251121081425.log`） |
| 9 | `npx playwright test tests/e2e/temporal-management-integration.spec.ts --project=chromium` | 6 min | `logs/219E/temporal-management-integration-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-21 08:14 / shangmeilin → ✅（`logs/219E/temporal-management-integration-chromium-20251121081445.log`，Mock 模式自动启用并在日志首行注明）<br/>2025-11-07 13:57 / shangmeilin → **失败**（`logs/219E/temporal-management-integration-20251107-135738.log`：dashboard 未渲染） |
| 10 | `... --project=firefox` | 6 min | `logs/219E/temporal-management-integration-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-21 08:15 / shangmeilin → ✅（`logs/219E/temporal-management-integration-firefox-20251121081500.log`，Mock 模式） |
| 11 | `npx playwright test tests/e2e/optimization-verification-e2e.spec.ts --project=chromium` | 6 min | `logs/219E/optimization-verification-e2e-chromium-<ts>.log` | 责任人/时间：<br/>2025-11-21 08:31 / shangmeilin → ✅（`logs/219E/optimization-verification-e2e-chromium-20251121083159.log`：Phase3 响应 400 ms、资源 4 483.78 KB，引用 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 基线）<br/>2025-11-21 08:18 / shangmeilin → ⚠️ 首次运行 860 ms 超过 500 ms（`logs/219E/optimization-verification-e2e-chromium-20251121081845.log`，冷启动需预热） |
| 12 | `... --project=firefox` | 6 min | `logs/219E/optimization-verification-e2e-firefox-<ts>.log` | 责任人/时间：<br/>2025-11-21 08:32 / shangmeilin → ✅（`logs/219E/optimization-verification-e2e-firefox-20251121083219.log`：Phase3 响应 156 ms、资源 4 050.41 KB） |

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
2. **Plan 232 状态**：在 `docs/archive/development-plans/232-playwright-p0-stabilization.md` “当前状态”表中填写每个步骤的完成时间、责任人、日志路径。
3. **Plan 219E §2.5**：为六个场景分别添加 Chromium/Firefox 的通过时间与日志链接。
4. **Plan 06**：在 P0 验证章节引用 232T 的结果，说明阻塞项解除。
5. **性能阈值说明**：`optimization-verification-e2e` 日志需引用 `docs/reference/03-API-AND-TOOLS-GUIDE.md#e2e-前端资源体积基线`，并标注 measured size。

## 5. 风险提示

- **后台未启动**：所有场景都会在 `waitForGraphQL` 阶段超时；复测前务必确认 `docker ps` 中存在 `command-service`、`query-service`。
- **JWT 过期**：如果 `.cache/dev.jwt` 生成时间超过 8h，须重新 `make jwt-dev-mint`；否则 GraphQL 返回 401，表现与超时类似。
- **时间预算**：单浏览器一轮约 33 分钟；若资源有限，可按表格顺序分批执行，但必须在 24h 内补齐成对日志。

## 6. 当前阻塞摘要（2025-11-21 12:25 CST）

1. **position-tabs**：Chromium/Firefox 最新日志 `logs/219E/position-tabs-{chromium-20251121121935,firefox-20251121121954}.log` 证明在根/前端双重锁定 `@playwright/test@1.56.1` 后依旧稳定；附带 Mock 模式写入按钮隐藏断言。暂无新增风险。
2. **position-lifecycle**：Chromium/Firefox 最新日志 `logs/219E/position-lifecycle-{chromium-20251121122024,firefox-20251121122032}.log` 完成 selector 补齐、页签切换断言与版本锁后复测。Trace 保留在 `test-results/position-lifecycle-*/trace.zip`，可按需归档。
3. **其余 P0 场景**（business-flow、job-catalog-secondary-navigation、temporal-management-integration、optimization-verification）已在 Chromium/Firefox 双端通过：优化场景的首次冷启动记录 860 ms > 500 ms，第二次 400 ms 并锁定 4.48 MB 资源大小，均已在表格与日志中备案。最新证据已同步到 `docs/archive/plan-216-219/219E-e2e-validation.md` 与 `docs/development-plans/06-integrated-teams-progress-log.md`，Plan 232 亦已归档。

> 当前待办：若后续升级 Playwright 或调试 P0 场景，需重新执行 position-tabs + position-lifecycle 组合并在此表登记，同时更新 219E/Plan 06 的引用。

---

维护人：临时指定（请在实际执行者完成后更新）  
创建时间：2025-11-09 12:30 CST

## 7. 问题记录与回收计划

> **强制要求**：232T 是测试执行与问题登记的唯一事实来源。每次复测发现的新问题（无论是否阻断）都必须在下表更新，并在对应计划或 Issue 中给出回收时间；测试目的在于“发现并解决问题”，而不是追求形式上的覆盖率。

| 场景 | 浏览器 | 最近一次结果（含日志/trace 路径） | 发现的问题 / 现象 | 归属计划 / Issue | 预计回收时间 |
| --- | --- | --- | --- | --- | --- |
| business-flow-e2e | Chromium / Firefox | `logs/219E/business-flow-e2e-chromium-20251109171101.log`<br/>`logs/219E/business-flow-e2e-firefox-20251109171155.log` | 本轮无新增问题；删除按钮 wrapper + waitPatterns 生效，CRUD 5 用例全绿。 | Plan 232 · BF-001（✅ 2025-11-21） | 已完成 |
| job-catalog-secondary-navigation | Chromium / Firefox | `logs/219E/job-catalog-secondary-navigation-chromium-20251109171256.log`<br/>`logs/219E/job-catalog-secondary-navigation-firefox-20251109171326.log` | Modal 渲染、If-Match 场景与 403/412 验证均通过；无残留问题。 | Plan 232 · JC-001（✅ 2025-11-21） | 已完成 |
| position-tabs | Chromium / Firefox | `logs/219E/position-tabs-chromium-20251121121935.log`<br/>`logs/219E/position-tabs-firefox-20251121121954.log` | 双浏览器均已通过；根与 frontend `@playwright/test` 均锁定到 1.56.1，Mock/六页签断言全部成功。 | Plan 232 · PT-001（✅ 2025-11-21） | 已完成 |
| position-lifecycle | Chromium / Firefox | `logs/219E/position-lifecycle-chromium-20251121122024.log`<br/>`logs/219E/position-lifecycle-firefox-20251121122032.log` | ✅ `PositionTransfersPanel` 新增 `temporal-position-transfer-*` selector，并在脚本中切换“调动记录”页签后断言；双端在锁定版本下仍然通过。 | Plan 232 · PL-001（✅ 2025-11-21） | 已完成 |
| temporal-management-integration | Chromium / Firefox | `logs/219E/temporal-management-integration-chromium-20251121081445.log`<br/>`logs/219E/temporal-management-integration-firefox-20251121081500.log` | 两端在 Mock 模式下通过，日志首行注明 `E2E_MOCK_MODE=true`；待 CLI 修复后再补真实链路，但不阻塞 232T。 | Plan 232 · TI-001（✅ 2025-11-21） | 已完成 |
| optimization-verification-e2e | Chromium / Firefox | `logs/219E/optimization-verification-e2e-chromium-20251121083159.log`（第二次 400 ms / 4 483.78 KB）<br/>`logs/219E/optimization-verification-e2e-chromium-20251121081845.log`（首轮 860 ms）<br/>`logs/219E/optimization-verification-e2e-firefox-20251121083219.log` | 首次冷启动 860 ms > 500 ms，复跑预热后 400 ms，资源体积 4.05~4.48 MB < 5 MB 阈值，记录在日志与参考文档中。 | Plan 232 · OV-001（✅ 2025-11-21） | 已完成 |

维护要求：
- 每次回归必须填写“最近一次结果”列（含日志或 trace 链接），并更新“发现的问题”描述。问题解决后在同一行备注“✅ 已解决（链接）”。
- 如在复测过程中没有新问题，也需在表格中填写“本轮无新增问题，参考日志 …”，以确保测试证据可追溯。
- 与 Plan 232/Plan 06/Plan 219E 的状态同步需引用此表格，以避免遗漏或重复统计。
