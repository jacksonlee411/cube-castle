# 05 — CI/本地一键自动化指引（Plan 254 经验沉淀）

版本: v1.0  
最后更新: 2025-11-16  
适用范围: 前端 E2E 与统一门禁（CQRS/端口/禁用端点）  
唯一事实来源: 工作流与脚本文件（见“权威索引”），本指南仅作执行指引与经验沉淀

---

## 目标
- 用“唯一门禁 + 统一证据”替代分散检查，避免重复造轮子与口径分叉
- 在 CI 与本地 VS Code 一键化复刻同样的门禁/E2E流程
- 以机器可读 JSON 结果作为自动化判断基准，不依赖本地 44800 报告 UI

## 权威索引（仅索引，不复制）
- 工作流（CI）
  - `.github/workflows/plan-254-gates.yml`（统一门禁 + E2E + 证据归档 + JSON SUMMARY 打印）
- 前端与门禁配置
  - `frontend/playwright.config.ts`（E2E_PLAN_ID 证据目录，JSON reporter 输出 `logs/plan<id>/results-*.json`）
  - `scripts/quality/architecture-validator.js`（CQRS/端口/禁用端点；含“禁止直连 :9090/:8090”）
- 本地自动化脚本
  - `scripts/ci/plan-254-local.sh`（本地 CI-like：compose→迁移→统一门禁→E2E→JSON SUMMARY）
  - `scripts/ci/fetch-gh-summary.sh`（从 GitHub Actions 拉取 SUMMARY，支持从 `.env`/`secrets/.env.local` 加载 token）
- VS Code 任务（本地仅供操作，文件未入库）
  - `.vscode/tasks.json`（一键“Local Gate”、“仅门禁”、“仅 E2E”）

## 统一门禁与证据
- 单一工具链：仅使用 `scripts/quality/architecture-validator.js` 作为门禁（规则：cqrs, ports, forbidden）
- 证据目录统一：通过 `E2E_PLAN_ID` 参数，E2E 证据与 JSON 统一落在 `logs/plan<ID>/*`
  - 254 计划：`logs/plan254/playwright-254-run-*.log`、`trace/*.zip`、`report-<ts>/`、`results-*.json`（可选 `har/*.har`）
- 机器可读结果：Playwright 配置 JSON reporter，脚本读取 `results-*.json` 并打印 `SUMMARY total=… passed=… failed=…`

## CI 工作流要点（Plan 254）
- Compose 后端服务（postgres/redis/rest/graphql），避免前端容器构建触发 TS 生产构建错误
- 由 Playwright 启动 dev webServer（`PW_SKIP_SERVER=0`），保证前端运行与本地一致
- 执行顺序：
  1) 统一门禁（`node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`）
  2) 前端 E2E（254）：`E2E_PLAN_ID=254`，证据归档至 `logs/plan254/*`
  3) 打印机器可读 SUMMARY（从 `logs/plan254/results-*.json` 抽取）

## 本地一键（VS Code/命令行）
- VS Code 任务（Terminal → Run Task…）：
  - “Plan 254: Local Gate (CI-like)”：compose/dev 栈→迁移→门禁→E2E→打印 SUMMARY（证据归档）
  - “Architecture Gate (frontend)”：仅门禁
  - “E2E: Plan 254 (evidence)”：仅 E2E 与证据
- 命令行：
  - `bash scripts/ci/plan-254-local.sh`
  - 环境变量：
    - `SKIP_INSTALL=1` 跳过 npm ci
    - `E2E_SAVE_HAR=1` 归档 HAR
    - `FRONTEND_BASE=http://localhost:3000` 覆盖前端基址

## 远程 SUMMARY 抓取（可选）
- 使用 `scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id>` 拉取 Actions run 的 SUMMARY 行
- Token 加载顺序：`secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量
- 安全与合规：`secrets/` 下的 `.env.local` 已被 `.gitignore` 忽略，密钥不得提交版本库

## 最佳实践
- 唯一门禁：不要叠加多个扫描器；统一依赖 architecture-validator，减少维护与分叉
- 证据参数化：通过 `E2E_PLAN_ID` 统一证据目录，避免“按计划复制脚本”的横向膨胀
- CI 与本地一致：CI 只 compose 后端；前端由 Playwright dev server 启动，保持一致性
- 产物可见性：始终归档 `logs/plan<ID>/*` 作为工件，便于 215 登记与索引状态更新

## 验收与登记
- 验收门槛（示例）：
  - 统一门禁关键违规=0；E2E 退出码=0；JSON SUMMARY 中 `failed=0`
  - 证据：`logs/plan<ID>/*`（报告、trace、JSON、可选 HAR）
- 登记路径：
  - 在 `docs/development-plans/215-phase2-execution-log.md` 登记 run 链接/工件与 SUMMARY 摘要
  - 在 `docs/development-plans/HRMS-DOCUMENTATION-INDEX.md` 更新状态为“已交付”

---

维护：架构与前端协作组（CI/门禁/前端）  
冲突与疑问：以工作流与脚本为唯一事实来源，如本指南有偏差，请以内置脚本/工作流为准
