# 05 — 本地一键自动化指引（Trunk + Local First）

版本: v1.4  
最后更新: 2025-11-16  
适用范围: 本地 VS Code 任务、Git hooks、一键门禁（CQRS/端口/禁用端点）  
唯一事实来源: 本地脚本与任务文件（见“权威索引”）；本指南仅为“本地运行”的执行指引

---

## 目标
- 在“仅保留 master（主干开发）”的前提下，以本地兜底为主，一键跑通门禁与自检
- 用 VS Code 任务 + Git hooks 保证本地与团队口径一致
- 统一证据输出路径，便于登记（215）与复核

## 分支与门禁原则
- 单人开发：强制主干开发（Trunk-Based），仅保留 master；提交前在本地通过快速门禁与编译（Local First）。
- CI 如开启受保护分支 Required checks，以工作流为唯一事实来源；本指南不依赖 CI，可离线执行。
- 证据必须本地落盘（reports/ 与 logs/），登记到 215。

## 权威索引（仅索引，不复制）
- 本地门禁脚本
  - `scripts/quality/architecture-validator.js`（CQRS/端口/禁用端点自检）
  - `eslint.config.architecture.mjs`（ESLint 架构守卫）
  - `.golangci.yml`（depguard/tagliatelle 配置）
  - `scripts/quality/golangci-fast.yml`（本地软门禁：仅 depguard + tagliatelle，避免 typecheck 噪音）
- 快速检查（Plan 250/253，本地可选）
  - `scripts/quality/gates-250-*.sh`（无 Docker）
  - `scripts/quality/gates-253-*.sh`（需要 Docker）
- VS Code（本地）
  - `.vscode/tasks.json`（本地任务，仓库忽略，不提交）
- 证据输出
  - `reports/architecture/architecture-validation.json`
  - `logs/plan<ID>/*`（建议：plan255）
  - 守卫脚本：`scripts/quality/root-whitelist-guard.sh`（根目录白名单）

## 本地一键门禁（255）
前置：确保根目录已安装依赖（仅一次）  
`npm ci`（根目录，用于 ESLint）；`golangci-lint` 可本地安装，或用容器/VS Code 扩展。

- 直接跑命令（本地兜底：前端 + 后端）
  - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
  - `npx eslint --no-warn-ignored -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}"`
  - `go build ./...`
  - （可选采集）`golangci-lint run -c scripts/quality/golangci-fast.yml || true`
- VS Code 一键（推荐）
  在本机 `.vscode/tasks.json` 添加如下任务（不提交到仓库）：
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "255: Local Gate (frontend+backend)",
      "type": "shell",
      "command": "bash",
      "args": ["-lc", "set -e; node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden && npx eslint --no-warn-ignored -c eslint.config.architecture.mjs 'frontend/src/**/*.{ts,tsx}' && go build ./... && (command -v golangci-lint >/dev/null 2>&1 && golangci-lint run -c scripts/quality/golangci-fast.yml || true)"],
      "options": { "cwd": "${workspaceFolder}" },
      "problemMatcher": [],
      "presentation": { "reveal": "always", "panel": "dedicated", "clear": true }
    },
    {
      "label": "255: ESLint Arch Only",
      "type": "shell",
      "command": "bash",
      "args": ["-lc", "npx eslint --no-warn-ignored -c eslint.config.architecture.mjs 'frontend/src/**/*.{ts,tsx}'"],
      "options": { "cwd": "${workspaceFolder}" }
    },
    {
      "label": "255: Static Arch Only",
      "type": "shell",
      "command": "bash",
      "args": ["-lc", "node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden"],
      "options": { "cwd": "${workspaceFolder}" }
    }
  ]
}
```
- 打开 VS Code → Terminal → Run Task… → 选择 “255: Local Gate (frontend+backend)”

## 本地快速检查（可选：250/253）
- Plan 250（无 Docker）
  - `bash scripts/quality/gates-250-no-legacy-env.sh`
  - `bash scripts/quality/gates-250-single-binary.sh`
  - `bash scripts/quality/gates-250-no-8090-in-command.sh`
- Plan 253（需要 Docker）
  - `bash scripts/quality/gates-253-compose-ports-and-images.sh`

## 证据与报告（本地）
- 架构报告（自动生成）：`reports/architecture/architecture-validation.json`
- 建议将终端输出 tee 到本地日志（示例）：
  - `ts=$(date +%Y%m%d_%H%M%S); mkdir -p logs/plan255; node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden 2>&1 | tee logs/plan255/architecture-validator-$ts.log`
  - `npx eslint --no-warn-ignored -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}" 2>&1 | tee logs/plan255/eslint-arch-$ts.log || true`
  - `go build ./... 2>&1 | tee logs/plan255/gobuild-$ts.log || true`
  - `golangci-lint run -c scripts/quality/golangci-fast.yml 2>&1 | tee logs/plan255/golangci-$ts.log || true`

## 本地 E2E（可选）
- 启动服务：`make docker-up && make run-dev`（迁移内置）
- 前端 Dev：`cd frontend && npm ci && npm run dev`
- 运行 E2E（示例，按需改 plan id 与项目）：
  - `cd frontend && E2E_PLAN_ID=254 PW_SKIP_SERVER=1 PW_BASE_URL=http://localhost:3000 npm run -s test:e2e:254`
- 证据路径（建议）：`logs/plan254/*`（trace、html 报告、JSON 结果）

## Git Hooks（推荐）
- 预推送（.git/hooks/pre-push）：建议运行“255 本地门禁 + go build”，golangci-lint 使用 `scripts/quality/golangci-fast.yml` 采集不阻塞；保持主干稳定、离线可用。

## 故障排除 / Playbook（本地优先）
- npm ci 失败（lock 与依赖不一致）
  - 现象：Install Node deps 步骤失败，或本地 `npm ci` 报 lock 不一致
  - 处理（Node 18）：`nvm use 18 && npm install --package-lock-only && git commit -m "chore(ci): refresh lock"`，随后 `npm ci`
- ESLint（文件被忽略/版本不兼容）
  - 使用 flat 配置：`eslint -c eslint.config.architecture.mjs 'frontend/src/**/*.{ts,tsx}' --no-warn-ignored`
  - 避免混用 frontend/.eslintrc.*；架构守卫统一走根的 flat 配置
- golangci-lint 类型噪音
  - 本地：`golangci-lint run -c scripts/quality/golangci-fast.yml` 或 `--disable-all -E depguard -E tagliatelle`
  - 预推送钩子仅采集不阻塞；阻断由 `go build ./...` 执行
- Plan 250 快修
  - 单二进制门禁：若 `cmd` 下存在第 2 个 main，加 `//go:build legacy`（示例：`cmd/hrms-server/query/tools/dump-schema/main.go`）
  - 8090 硬编码：去除 `":8090"`/`"8090"` 字面量，改读取 `PORT`；保留数值禁用逻辑（解析端口后比较 `== 8090`）
- 文档/合规脚本
  - 根目录白名单失败：本地运行 `bash scripts/quality/root-whitelist-guard.sh` 查看不被允许的根文件；将文件移至合适目录或按需维护白名单
  - 契约校验脚本缺失文件：`scripts/quality/lint-validation.js` 会跳过缺失文件；在存在的路径上保持严格

## 从 CI 收集“Required checks”反馈（本地）
- 令牌加载顺序：`secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量（`GITHUB_TOKEN`/`GH_TOKEN`）
- 一键拉取 SUMMARY（需 unzip 或 bsdtar）
  - `bash scripts/ci/fetch-gh-summary.sh <owner/repo> <run_id> > logs/plan255/ci-summary-<run_id>.txt`
- 直接读取运行与检查（无需解压）
  - 运行概览：`curl -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/repos/<owner>/<repo>/actions/runs/<run_id> | jq '{workflow:.name,event,status,conclusion,branch:.head_branch,sha:.head_sha,html_url}'`
  - 运行 Jobs：`curl -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/repos/<owner>/<repo>/actions/runs/<run_id>/jobs?per_page=100 | jq -r '.jobs[] | [.name,.status,.conclusion] | @tsv'`
  - 提交检查：`curl -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/repos/<owner>/<repo>/commits/$SHA/check-runs?per_page=100 | jq -r '.check_runs[] | [.name,.status,.conclusion] | @tsv'`

## 最佳实践
- 唯一门禁：不要叠加多个扫描器；统一依赖 architecture-validator，减少维护与分叉
- 证据规范：按计划号将报告/trace 统一落在 `logs/plan<ID>/*`，避免“复制脚本”膨胀
- 一致性：本地门禁与团队口径一致；如需 CI，请参考工作流注释，但本文件以“本地兜底”为准
- 可见性：建议始终生成 `logs/plan<ID>/*` 与 `reports/architecture/architecture-validation.json` 以便登记与复核

---

维护：架构与前端协作组（本地门禁/前端）  
冲突与疑问：以本地脚本与任务文件为唯一事实来源，如本指南有偏差，请以脚本为准
