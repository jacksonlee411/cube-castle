# 05 — CI/本地一键自动化指引（本地兜底 = PR 等效）

版本: v1.4  
最后更新: 2025-11-20  
适用范围: 前端 E2E 与统一门禁（CQRS/端口/禁用端点）+ 后端 golangci-lint  
唯一事实来源: 工作流与脚本文件（见“权威索引”），本指南仅作执行指引与经验沉淀（若有偏差，以工作流与脚本为准）

---

## 目标
- 用“唯一门禁 + 统一证据”替代分散检查，避免重复造轮子与口径分叉
- 在 CI 与本地 VS Code 一键化复刻“与 PR 远程相同”的门禁/E2E 流程
- 以统一落盘的报告/trace 作为自动化与人工验收的共同依据（JSON SUMMARY 后续计划）

## 权威索引（仅索引，不复制）
- 工作流（CI）
  - `.github/workflows/plan-255-gates.yml`（统一门禁：前端 CQRS/端口/禁用端点 + 后端 golangci-lint + 报告归档）
  - `.github/workflows/frontend-e2e-devserver.yml`（仅后端 compose，前端由 Playwright dev server 启动，统一本地/CI）
  - `.github/workflows/frontend-e2e.yml`、`.github/workflows/e2e-tests.yml`（历史 E2E：包含前端容器）
- 前端与门禁配置
  - `frontend/playwright.config.ts`（本地/CI 一致的 E2E 行为；reporter 当前为 html，trace/har 按需）
  - `scripts/quality/architecture-validator.js`（CQRS/端口/禁用端点；禁止硬编码端口，需统一走配置）
  - `eslint.config.architecture.mjs`（AST 级别“架构守卫”规则集；CI 中为非阻断，仅记录日志）
- 后端门禁
  - `golangci-lint` 固定版本：`v1.59.1`（与 CI 对齐，保证结果可复现）
  - 固定调用路径（避免误用 PATH 里旧版本）：`$(go env GOPATH)/bin/golangci-lint`
- VS Code 任务（本地仅供操作，文件未入库）
  - `.vscode/tasks.json`（一键“Local Gate”、“仅门禁”、“仅 E2E”）

## 统一门禁与证据
- 门禁工具链（阻断）：
  - 前端：`node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
  - 后端：`golangci-lint run`（版本固定 v1.59.1）
- 辅助守卫（非阻断，记录日志）
  - `npx eslint -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}"`
- 证据目录统一：按“计划号 plan<ID>”归档到 `logs/plan<ID>/*`
  - 运行日志：`logs/plan<ID>/playwright-run-*.log`
  - 失败场景 trace：`logs/plan<ID>/trace/*.zip`
  - HTML 报告：`logs/plan<ID>/report-<ts>/`
  - JSON 结果：`logs/plan<ID>/results-*.json`（Playwright JSON reporter，已启用）
  - 可选 HAR：设置 `E2E_SAVE_HAR=1`

## CI 工作流要点（通用）
- 门禁：`plan-255-gates.yml` 在 CI 中运行 `architecture-validator`（规则：cqrs,ports,forbidden），并归档报告
- 后端门禁：`golangci-lint run`（版本固定，阻断）
- ESLint 架构守卫：记录日志，不阻断
- E2E（统一推荐）：`frontend-e2e-devserver.yml` 仅 compose 后端（postgres/redis/rest/graphql），前端由 Playwright dev server 启动（`PW_SKIP_SERVER=0`）
- 历史 E2E：`frontend-e2e.yml` / `e2e-tests.yml` 使用包含前端容器的完整栈（逐步迁移中）

## 自托管 Runner 选型（Plan 265 + Plan 269）
- **默认：WSL Runner**
  - 2025-11-20（Plan 269）起，唯一受支持的自托管形态；Runner 裸部署在 WSL（Ubuntu 20.04+/22.04），标签 `[self-hosted,cubecastle,wsl]`。
  - 通过 `scripts/ci/runner/wsl-install.sh`/`wsl-uninstall.sh`/`wsl-verify.sh` 管理；步骤详见 `docs/reference/wsl-runner-setup.md`。
  - 所有 workflow matrix 中的 self-hosted job 仅保留 `wsl` 标签，`wsl_only` 条件用于限制触发场景。
- **历史方案：Docker Runner（已退役）**
  - 旧版通过 `docker-compose.runner*.yml` 维护容器 Runner；现已关闭，不再作为回退路径。
  - 如确需恢复 Docker Runner，必须创建新的计划条目并重新审批。
- **CI 工作流调整**
  - 所有引用自托管 Runner 的 workflow（`document-sync`, `api-compliance`, `consistency-guard`, `ci-selfhosted-*`, `iig-guardian`, `contract-testing`, `plan-254-gates`, `e2e-smoke` 等）均已增加 `wsl` 标签。
  - Workflow matrix 需包含 `[self-hosted, cubecastle, wsl]` 并在 job 注释中声明：Runner 必须具备 Docker CLI/Compose，业务服务依旧在容器中运行。
- **记录与监控**
  - 每次切换或新增 Runner 节点，必须在 Plan 265/266 中记录 Run ID、脚本日志与回滚方案。
  - `scripts/network/verify-github-connectivity.sh` 与 `docs/reference/docker-network-playbook.md` 提供网络守护；`scripts/ci/runner/watchdog.sh` 可扩展自动校验。

## 本地一键（VS Code/命令行）
- VS Code 任务（Terminal → Run Task…）：
  - “Plan 254: Local Gate (CI-like)”：compose/dev 栈→迁移→门禁→E2E→证据归档
  - “Architecture Gate (frontend)”：仅门禁
  - “E2E: Plan 254 (evidence)”：仅 E2E 与证据
- 命令行：
  - 本地一键（通用）：
    - `E2E_PLAN_ID=254 bash scripts/ci/plan-local.sh`
  - 或逐步执行：
    - 启动最小依赖并运行服务（迁移内置在 run-dev 中）：`make docker-up && make run-dev`
  - 统一门禁（前端架构一致性）：
    - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
  - 前端 E2E（按计划号归档）：
    - `cd frontend && E2E_PLAN_ID=254 PW_SKIP_SERVER=${PW_SKIP_SERVER:-0} npm run -s test:e2e:plan`
  - 常用环境变量：
    - `E2E_SAVE_HAR=1` 归档 HAR
    - `PW_BASE_URL=http://localhost:3000` 覆盖前端基址
    - `PW_SKIP_SERVER=1` 跳过 webServer（本地已手动 `npm run dev` 时）

## PR 等效的“本地兜底流程”（强烈建议）
以下步骤严格复刻 PR 的远程门禁（plan-255），用于本地自查。若某一步失败，请先以“权威索引”中的脚本/工作流为准排障。

1) 版本与端口预检（符合 AGENTS 强约束）
   - Go ≥1.24（与仓库 toolchain 对齐），Node ≥18（E2E 建议 Node 20）
   - 严禁宿主机安装 Postgres/Redis；确保 5432/6379 未被宿主占用 → `make docker-up`
2) 启动与鉴权
   - `make run-dev`（内置迁移与健康检查，后端端口 9090/8090）
   - `make jwt-dev-setup && make jwt-dev-mint`（生成 `.cache/dev.jwt`）
3) 前端统一门禁（阻断，唯一口径）
   - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
   - 产物：`reports/architecture/architecture-validation.json`
4) ESLint 架构守卫（非阻断，记录日志）
   - `npx eslint -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}" 2>&1 | tee "logs/plan255/eslint-architecture-$(date +%Y%m%d_%H%M%S).log" || true`
5) 后端门禁（阻断，版本固定）
   - `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1`
   - `golangci-lint run 2>&1 | tee "logs/plan255/golangci-lint-$(date +%Y%m%d_%H%M%S).log"; test ${PIPESTATUS[0]} -eq 0`
6) 前端 E2E（DevServer，与 CI 一致）
   - `cd frontend && npm ci && npx playwright install --with-deps`
   - `cd frontend && E2E_PLAN_ID=255 PW_SKIP_SERVER=0 PW_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 PW_JWT=$(cat ../.cache/dev.jwt) npm run -s test:e2e:plan`
   - 证据：`logs/plan255/results-*.json`、`logs/plan255/trace/*.zip`、`logs/plan255/report-*/`
7) 打印 JSON SUMMARY（验收总览）
   - `node scripts/ci/print-e2e-summary.js 255`

门禁判定（与 CI 同步）
- 关键违规=0：前端 architecture-validator 退出码=0；后端 golangci-lint 退出码=0
- E2E 退出码=0；`SUMMARY_ALL failed=0`
- 证据完整：`logs/plan255/**` 与 `reports/architecture/architecture-validation.json`

提示
- 若你已经本地 `npm run dev`，可将 `PW_SKIP_SERVER=1` 以复用本地前端服务
- 不得修改 `docker-compose.dev.yml` 端口映射来规避冲突；若冲突，请卸载/停止宿主机服务（遵循 AGENTS.md 强制约束）

## 无网络/容器受限环境的等效方案（推荐）
当本地无法拉取容器/浏览器依赖时，执行“CI 远程跑 + 本地取证”流程，仍然实现 PR 等效的门禁：

- 触发 CI（任选其一）：
  - 推送分支或创建 PR（自动触发 plan-255-gates 与 Frontend E2E DevServer）
  - 手动触发工作流（GitHub Actions → Frontend E2E (DevServer) / plan-255-gates）
- 本地抓取 SUMMARY（机器可读汇总）：
  - `scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id>`
- 本地拉取 CI 工件（证据与报告）：
  - `scripts/ci/fetch-gh-artifact.sh <owner/repo> <run_id> 'frontend-e2e-devserver|plan255-logs' logs/plan255/ci-artifacts`
  - 说明：工具会将 CI 中的 `logs/plan<ID>/*` 与报告解压至本地，便于登记/对比
- 门禁判定口径：
  - 与本地一致（前端/后端门禁=0 关键违规；E2E 失败用例=0；证据完整）

注意
- 远程抓取需要 `GITHUB_TOKEN`（加载顺序：`secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量）
- 若仅本地门禁可执行（architecture-validator + golangci-lint）但无法 E2E，请以 CI 产出的 E2E 证据作为验收对等物（登记时标注“E2E=CI 取证”）

### Plan 267-D：WSL 静态 hosts 覆盖（GitHub 应急）

> 适用范围：受限网络下 `github.com` 被劫持到 11.x.x.x，导致 `git ls-remote` / `curl` 在 WSL/Docker 中失败。Plan 267-D 已验证“WSL 停止自动生成 hosts + 静态写入官方 IP”能恢复访问。

1. **预检**：确保 `/etc/wsl.conf` 已包含：
   ```ini
   [network]
   generateHosts=false
   ```
   修改后执行 `wsl.exe --shutdown`（Windows PowerShell）重启 WSL，防止 `/etc/hosts` 被自动覆盖。
2. **写入官方 IP**：在 WSL 内运行：
   ```bash
   sudo bash scripts/network/configure-github-hosts.sh
   ```
   脚本会先备份 `HOSTS_FILE`（示例 `/etc/hosts.plan267.<timestamp>.bak`），然后写入 `# Plan 267-D GitHub override` 块，包含 `github.com`、`codeload.github.com`、`api.github.com`、`raw.githubusercontent.com`、`github-releases.githubusercontent.com`、`objects.githubusercontent.com`、`objects-origin.githubusercontent.com`、`release-assets.githubusercontent.com` 等关键域名的最新官方 IP（2025-11-19 取自 `getent ahostsv4`）。
3. **同步 Runner**：执行
   ```bash
   bash scripts/network/apply-github-hosts-to-runner.sh
   ```
   该脚本会先更新宿主 `/etc/hosts`，再将 `configure-github-hosts.sh` 拷贝到 `cubecastle-gh-runner` 容器并用 `sudo bash` 应用，确保自托管 Runner 与宿主保持一致。
4. **验证命令**（所有命令需返回 200 并解析到脚本写入的 IP）：
   ```bash
   getent hosts github.com
   curl -I https://github.com
   git ls-remote https://github.com/jacksonlee411/cube-castle
   docker compose -f docker-compose.runner.persist.yml exec gh-runner curl -I https://github.com
   ```
   建议将 `git ls-remote` 结果接入 Watchdog/诊断脚本，发现异常时自动重新执行本脚本。
5. **回退/更新 IP**：若网络团队放通或 GitHub IP 变更，记录最近一次备份文件并执行：
   ```bash
   sudo cp /etc/hosts.plan267.<timestamp>.bak /etc/hosts
   wsl.exe --shutdown
   ```
   然后依据新的官方 IP（必要时重新采集）再运行 `bash scripts/network/apply-github-hosts-to-runner.sh`。

## SUMMARY 打印与远程抓取
- 本地/CI 打印：`node scripts/ci/print-e2e-summary.js <planId>` 会扫描 `logs/plan<ID>/results-*.json` 并输出机器可读汇总
- 远程抓取：`scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id>` 从 Actions run 的日志压缩包中提取包含 `SUMMARY` 的行
  - Token 加载顺序：`secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量（`GITHUB_TOKEN`/`GH_TOKEN`）
  - 工件与日志仍是验收依据，SUMMARY 便于快速登记/比对

## 最佳实践
- 唯一门禁：不要叠加多个扫描器；统一依赖 architecture-validator，减少维护与分叉
- 后端门禁：golangci-lint 版本固定（v1.59.1）以确保结果可复现
- 证据规范：按计划号将报告/trace 统一落在 `logs/plan<ID>/*`，避免“复制脚本”膨胀
- 一致性：本地与 CI 尽量使用相同的端口/基址/鉴权路径；当 CI 策略有差异时，以工作流注释说明为准
- 产物可见性：始终归档 `logs/plan<ID>/*` 与 `reports/architecture/architecture-validation.json` 作为工件，便于 215 登记与索引状态更新

## 验收与登记
- 验收门槛（示例）：
  - 统一门禁关键违规=0（前端 architecture-validator + 后端 golangci-lint）；E2E 退出码=0；`SUMMARY_ALL failed=0`
  - 证据：`logs/plan<ID>/*`（报告、trace、JSON、可选 HAR）；`reports/architecture/architecture-validation.json`
- 登记路径：
  - 在 `docs/development-plans/215-phase2-execution-log.md` 登记 run 链接/工件与 SUMMARY 摘要
  - 在 `docs/development-plans/HRMS-DOCUMENTATION-INDEX.md` 更新状态为“已交付”

---

维护：架构与前端协作组（CI/门禁/前端）  
冲突与疑问：以工作流与脚本为唯一事实来源，如本指南有偏差，请以内置脚本/工作流为准
