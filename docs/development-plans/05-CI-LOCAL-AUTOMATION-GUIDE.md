# 05 — 本地一键自动化指引（Local Only）

版本: v1.3  
最后更新: 2025-11-16  
适用范围: 本地 VS Code 任务、Git hooks、一键门禁（CQRS/端口/禁用端点）  
唯一事实来源: 本地脚本与任务文件（见“权威索引”）；本指南仅为“本地运行”的执行指引

---

## 目标
- 在“完全本地”的前提下，一键跑通门禁与自检，不依赖远端
- 用 VS Code 任务 + Git hooks 保证本地与团队口径一致
- 统一证据输出路径，便于登记（215）与复核

## 权威索引（仅索引，不复制）
- 本地门禁脚本
  - `scripts/quality/architecture-validator.js`（CQRS/端口/禁用端点自检）
  - `eslint.config.architecture.mjs`（ESLint 架构守卫）
  - `.golangci.yml`（depguard/tagliatelle 配置）
- 快速检查（Plan 250/253，本地可选）
  - `scripts/quality/gates-250-*.sh`（无 Docker）
  - `scripts/quality/gates-253-*.sh`（需要 Docker）
- VS Code（本地）
  - `.vscode/tasks.json`（本地任务，仓库忽略，不提交）
- 证据输出
  - `reports/architecture/architecture-validation.json`
  - `logs/plan<ID>/*`（建议：plan255）

## 本地一键门禁（255）
前置：确保根目录已安装依赖（仅一次）  
`npm ci`（根目录，用于 ESLint）；`golangci-lint` 可本地安装，或用容器/VS Code 扩展。

- 直接跑命令（前端 + 后端）
  - `node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden`
  - `npx eslint -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}"`
  - `golangci-lint run`
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
      "args": ["-lc", "node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden && npx eslint -c eslint.config.architecture.mjs \"frontend/src/**/*.{ts,tsx}\" && golangci-lint run"],
      "options": { "cwd": "${workspaceFolder}" },
      "problemMatcher": [],
      "presentation": { "reveal": "always", "panel": "dedicated", "clear": true }
    },
    {
      "label": "255: ESLint Arch Only",
      "type": "shell",
      "command": "bash",
      "args": ["-lc", "npx eslint -c eslint.config.architecture.mjs \"frontend/src/**/*.{ts,tsx}\""],
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
  - `npx eslint -c eslint.config.architecture.mjs "frontend/src/**/*.{ts,tsx}" 2>&1 | tee logs/plan255/eslint-arch-$ts.log || true`
  - `golangci-lint run 2>&1 | tee logs/plan255/golangci-$ts.log || true`

## 本地 E2E（可选）
- 启动服务：`make docker-up && make run-dev`（迁移内置）
- 前端 Dev：`cd frontend && npm ci && npm run dev`
- 运行 E2E（示例，按需改 plan id 与项目）：
  - `cd frontend && E2E_PLAN_ID=254 PW_SKIP_SERVER=1 PW_BASE_URL=http://localhost:3000 npm run -s test:e2e:254`
- 证据路径（建议）：`logs/plan254/*`（trace、html 报告、JSON 结果）

## 可选（联网）— 自动推送与创建 PR
本节可选；完全本地时忽略。  
`bash scripts/ci/auto-pr.sh --title "<标题>" --body-file <正文md> --base master --head <分支>`  
Token 从 `secrets/.env.local` → `secrets/.env` → `.env.local` → `.env` → 环境变量 加载。  
产出日志：`logs/plan255/pr-<ts>.txt`、`logs/plan255/pr-response-<ts>.json`

## 最佳实践
- 唯一门禁：不要叠加多个扫描器；统一依赖 architecture-validator，减少维护与分叉
- 证据规范：按计划号将报告/trace 统一落在 `logs/plan<ID>/*`，避免“复制脚本”膨胀
- 一致性：本地门禁与团队口径一致；如需 CI，请参考工作流注释，但本文件不依赖 CI
- 可见性：建议始终生成 `logs/plan<ID>/*` 与 `reports/architecture/architecture-validation.json` 以便登记与复核

---

维护：架构与前端协作组（本地门禁/前端）  
冲突与疑问：以本地脚本与任务文件为唯一事实来源，如本指南有偏差，请以脚本为准
