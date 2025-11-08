# Plan 230D – Position CRUD 数据链路恢复与 E2E 验证

**文档编号**: 230D  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**前置计划**: 230B（数据修复）+ 230C（自检脚本）  
**关联计划**: 219E 端到端测试、219T 验收报告  
**负责人**: QA 团队 + 前端 + Backend

---

## 1. 背景与目标

- 在 219T 回归中，`tests/e2e/position-crud-full-lifecycle.spec.ts` Step 1 因 `OPER` 缺失而失败。230B 恢复数据后，需要重新播种职位/任职链路，并在真实环境中完成 Playwright 验证。  
- 230D 负责确保 Position CRUD + Assignment 流程可在 Chromium 至少一次跑通，并生成可追溯的日志与测试产物，为 219E 解除 P0 阻塞。

---

## 2. 范围

1. 运行 `scripts/dev/seed-position-crud.sh`（若尚未存在则本计划负责补齐）或使用 REST API 手动创建测试职位，对应的 Request/Response 记录到 `logs/230/position-seed-*.log`。  
2. 复跑 Playwright 脚本 `tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium，必要时 Firefox），并归档 `frontend/test-results/position-crud-full-lifecyc-<commit>-chromium/`。  
3. 若 Playwright 失败，需根据日志明确判定是数据问题还是 UI/test locator 缺陷；数据问题需反馈 230B/230C，UI 问题则交由前端修复。  
4. 将成功运行的 RequestId、截图、trace、Junit 报告记录在 `logs/230/position-crud-playwright-*.log`。

---

## 3. 任务与步骤

| 步骤 | 描述 | 输出 |
| --- | --- | --- |
| D1 | 在 `make docker-up && make run-dev` 环境下，确认 `curl http://localhost:9090/health`、`curl http://localhost:8090/health` 均返回 200 | `logs/230/position-env-check-*.log` |
| D2 | 运行数据播种脚本或 REST/GraphQL 手动创建一套 Position + Assignment，记录 RequestId/响应 | `logs/230/position-seed-YYYYMMDD.log` |
| D3 | 准备 Playwright：`cd frontend && npm install`（如需）、`npx playwright install --with-deps`（仅容器中执行），确保 `.cache/dev.jwt`、`PW_TENANT_ID`、`PW_JWT` 可用 | 命令输出附在 `logs/230/position-crud-playwright-*.log` |
| D4 | 执行 `npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts --project=chromium --workers=1 --reporter=line,junit`，必要时串行重试一次 | `frontend/test-results/position-crud-full-lifecyc-<commit>-chromium/`、`logs/230/position-crud-playwright-*.log` |
| D5 | 将成功 run 的 RequestId、关键步骤截图、trace 链接写入日志，并在 PR 中引用 | 同上 |

---

## 4. 依赖

- 230B 的迁移已合入主干，`OPER` 数据可用。  
- 230C 的脚本已在 `make status` 中启用，执行 Playwright 前需确保诊断通过。  
- `.cache/dev.jwt`、`PW_TENANT_ID`、`PW_JWT` 已由 `make jwt-dev-setup && make jwt-dev-mint` 生成。  
- Docker 环境按照 `AGENTS.md` 运行。

---

## 5. 验收标准

1. `tests/e2e/position-crud-full-lifecycle.spec.ts` 在 Chromium 环境一次通过；如 Firefox 也执行，需分别记录日志。  
2. `frontend/test-results/position-crud-full-lifecyc-<commit>-chromium/` 中无新的 422/`JOB_CATALOG_NOT_FOUND`，Junit 报告全绿。  
3. `logs/230/position-crud-playwright-*.log` 列出了 Step1~Step6 的 RequestId，便于追踪。  
4. 若测试失败，日志需指出是否因数据、UI 或脚本问题，并将责任转交相关计划；未关闭的问题需登记在 `docs/development-plans/06-integrated-teams-progress-log.md`。  
5. 230D 的结果（日志、测试产物路径）被引用到 219E Plan 2.4/2.5 表中。

---

## 6. 交付记录（2025-11-08）

- **环境与种子**：`logs/230/position-env-check-20251108T095108.log` 记录 REST/GraphQL 健康检查均为 200；`logs/230/position-seed-20251108T094735.log` 保存 `scripts/dev/seed-position-crud.sh` 播种的 REST RequestId（新职位 `P1000027`）。  
- **Playwright 结果**：`logs/230/position-crud-playwright-20251108T102815.log` 显示 `tests/e2e/position-crud-full-lifecycle.spec.ts` 9 个步骤（Create→Delete）全部通过，RequestId（Create `741b50a6-...` / Update `b185d534-...` / Fill `ef51d7d4-...` / Vacate `5f420a06-...` / Delete `d0fc3afb-...`）已归档；对应产物集中于 `frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/`（含 trace、video、Junit）。  
- **问题回归**：前期的 `position-detail-card`/`position-temporal-page` 加载失败已由 Playwright 脚本验证恢复；Step 5 断言现改为确认页面剔除任职人记录。

---

> 唯一事实来源：`logs/230/position-env-check-20251108T095108.log`、`logs/230/position-seed-20251108T094735.log`、`logs/230/position-crud-playwright-20251108T102815.log`、`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/`。  
> 更新时间：2025-11-08。
