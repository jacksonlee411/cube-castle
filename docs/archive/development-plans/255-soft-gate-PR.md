# Plan 255B — Soft Gate PR 与门禁落地（已归档）

标题: `refactor(health-alerting): migrate JSON tags to camelCase and harden Plan 255 gates`  
归档日期: 2025-11-20  
状态: ✅ 合并完结（功能与门禁均在主干启用）

---

## 1. 完成内容回顾
- **JSON 字段迁移**：监控/告警导出结构 `resolved_at/max_retries/enabled_by/status_equals/response_time_gt/consecutive_fails` 全部改为 camelCase；`internal/monitoring/health/alerting.go` 不再依赖 `//nolint:tagliatelle` 例外。
- **Plan 255 门禁扩展**：  
  - `.github/workflows/plan-255-gates.yml` 引入 ESLint Flat Config（`eslint.config.architecture.mjs`）、前端 architecture-validator、根路径审计、以及 `golangci-lint` depguard/tagliatelle 组合。  
  - Added `PLAN255_ROOT_AUDIT_MODE` 并最终切换至 `hard`，确保根路径违规会阻断。  
  - 工作流日志与工件：`logs/plan255/eslint-architecture-*.log`、`logs/plan255/architecture-validator-*.log`、`logs/plan255/golangci-lint-*.log`、artifact `plan255-logs`.
- **词表/语义统一**：状态字段在 ESLint 与 architecture-validator 中统一为 `status / isCurrent / isFuture / isTemporal`；`GET` 例外仅限 `/auth`，JWKS 访问已改用 `UnauthenticatedRESTClient`。
- **Branch Protection**：`plan-250-gates`、`plan-253-gates`、`plan-255-gates` 均设为 Required；相关 run 链接与说明记录于 `logs/plan255/branch-protection-required-checks.md`。

## 2. 证据
- `reports/architecture/architecture-validation.json` — 最新摘要 `totalViolations=0`（2025-11-20 12:34:58Z）。  
- `logs/plan255/architecture-validator-20251116_101740.log` — 首次软门禁运行日志。  
- `logs/plan255/audit-root-20251116_102250.log` → `logs/plan255/architecture-root-20251117_133935.json` — 根路径整改前/后的完整记录。  
- GitHub Actions run（plan‑255-gates）：<https://github.com/jacksonlee411/cube-castle/actions/runs/19403570892>（成功，artifact: `plan255-logs`）。  
- 受保护分支 Required checks 说明：`logs/plan255/branch-protection-required-checks.md`。

## 3. 影响与兼容性
- 外部消费 snake_case JSON 的系统需在一个迭代内完成兼容；窗口结束后只保留 camelCase。  
- 通过门禁统一，计划 202/250/253/254/258 的约束可共享：REST 仅用于命令、GraphQL 用于查询，端口与代理策略以单基址约束。

## 4. 结论
Plan 255 的“软门禁→硬门禁”过渡已在主干生效；相关 PR、脚本与 CI 工件已归档。后续如需新增例外或修改门禁，请以本归档为参考在新的计划文档中记录。
