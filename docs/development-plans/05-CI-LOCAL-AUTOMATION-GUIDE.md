# 05 — CI/本地一键自动化指引（Plan 254 经验沉淀）

版本: v1.2  
最后更新: 2025-11-16  
适用范围: 前端 E2E 与统一门禁（CQRS/端口/禁用端点）  
唯一事实来源: 工作流与脚本文件（见“权威索引”），本指南仅作执行指引与经验沉淀

---

## 目标
- 用“唯一门禁 + 统一证据”替代分散检查，避免重复造轮子与口径分叉
- 在 CI 与本地 VS Code 一键化复刻同样的门禁/E2E流程
- 以统一落盘的报告/trace 作为自动化与人工验收的共同依据（JSON SUMMARY 后续计划）

## 权威索引（仅索引，不复制）
- 工作流（CI）
  - `.github/workflows/plan-255-gates.yml`（前端统一门禁：CQRS/端口/禁用端点 + 架构报告归档）
  - `.github/workflows/frontend-e2e-devserver.yml`（仅后端 compose，前端由 Playwright dev server 启动，统一本地/CI）
  - `.github/workflows/frontend-e2e.yml`、`.github/workflows/e2e-tests.yml`（历史 E2E：包含前端容器）
- 前端与门禁配置
  - `frontend/playwright.config.ts`（本地/CI 一致的 E2E 行为；reporter 当前为 html，trace/har 按需）
  - `scripts/quality/architecture-validator.js`（CQRS/端口/禁用端点；禁止硬编码端口，需统一走配置）
- VS Code 任务（本地仅供操作，文件未入库）
  - `.vscode/tasks.json`（一键“Local Gate”、“仅门禁”、“仅 E2E”）

## 统一门禁与证据
- 单一工具链：仅使用 `scripts/quality/architecture-validator.js` 作为门禁（规则：cqrs, ports, forbidden）
- 证据目录统一：按“计划号 plan<ID>”归档到 `logs/plan<ID>/*`
  - 运行日志：`logs/plan<ID>/playwright-run-*.log`
  - 失败场景 trace：`logs/plan<ID>/trace/*.zip`
  - HTML 报告：`logs/plan<ID>/report-<ts>/`
  - JSON 结果：`logs/plan<ID>/results-*.json`（Playwright JSON reporter，已启用）
  - 可选 HAR：设置 `E2E_SAVE_HAR=1`

## CI 工作流要点（通用）
- 门禁：`plan-255-gates.yml` 在 CI 中运行 `architecture-validator`（规则：cqrs,ports,forbidden），并归档报告
- E2E（统一推荐）：`frontend-e2e-devserver.yml` 仅 compose 后端（postgres/redis/rest/graphql），前端由 Playwright dev server 启动（`PW_SKIP_SERVER=0`）
- 历史 E2E：`frontend-e2e.yml` / `e2e-tests.yml` 使用包含前端容器的完整栈（逐步迁移中）

## 本地一键（VS Code/命令行）
- VS Code 任务（Terminal → Run Task…）：
  - “Plan 254: Local Gate (CI-like)”：compose/dev 栈→迁移→门禁→E2E→证据归档
  - “Architecture Gate (frontend)”：仅门禁
  - “E2E: Plan 254 (evidence)”：仅 E2E 与证据
- 命令行：
  - 启动最小依赖并运行服务（迁移内置在 run-dev 中）：
    - `make docker-up && make run-dev`
  - 统一门禁（前端架构一致性）：
    - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
  - 通用 E2E（按计划号归档）：
    - `cd frontend && E2E_PLAN_ID=254 PW_SKIP_SERVER=${PW_SKIP_SERVER:-0} npm run -s test:e2e:plan`
  - 常用环境变量：
    - `E2E_SAVE_HAR=1` 归档 HAR
    - `PW_BASE_URL=http://localhost:3000` 覆盖前端基址
    - `PW_SKIP_SERVER=1` 跳过 webServer（本地已手动 `npm run dev` 时）

## SUMMARY 打印与远程抓取
- 本地/CI 打印：`node scripts/ci/print-e2e-summary.js <planId>` 会扫描 `logs/plan<ID>/results-*.json` 并输出机器可读汇总
- 远程抓取：`scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id>` 从 Actions run 的日志压缩包中提取包含 `SUMMARY` 的行
  - Token 加载顺序：`secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量（`GITHUB_TOKEN`/`GH_TOKEN`）
  - 工件与日志仍是验收依据，SUMMARY 便于快速登记/比对

## 最佳实践
- 唯一门禁：不要叠加多个扫描器；统一依赖 architecture-validator，减少维护与分叉
- 证据规范：按计划号将报告/trace 统一落在 `logs/plan<ID>/*`，避免“复制脚本”膨胀
- 一致性：本地与 CI 尽量使用相同的端口/基址/鉴权路径；当 CI 策略有差异时，以工作流注释说明为准
- 产物可见性：始终归档 `logs/plan<ID>/*` 与 `reports/architecture/architecture-validation.json` 作为工件，便于 215 登记与索引状态更新

## 验收与登记
- 验收门槛（示例）：
  - 统一门禁关键违规=0；E2E 退出码=0；`SUMMARY_ALL failed=0`
  - 证据：`logs/plan<ID>/*`（报告、trace、JSON、可选 HAR）；`reports/architecture/architecture-validation.json`
- 登记路径：
  - 在 `docs/development-plans/215-phase2-execution-log.md` 登记 run 链接/工件与 SUMMARY 摘要
  - 在 `docs/development-plans/HRMS-DOCUMENTATION-INDEX.md` 更新状态为“已交付”

---

维护：架构与前端协作组（CI/门禁/前端）  
冲突与疑问：以工作流与脚本为唯一事实来源，如本指南有偏差，请以内置脚本/工作流为准
