# Plan 255A — 根路径端口/禁用端点审计整改（已归档）

文档编号: 255-AUDIT-ROOT  
归档日期: 2025-11-20  
状态: ✅ 完成（PLAN255_ROOT_AUDIT_MODE=hard 已在 `plan-255-gates` 中常态启用）

---

## 1. 交付摘要
- `plan-255-gates` 工作流（required check）现以 `PLAN255_ROOT_AUDIT_MODE=hard` 运行，前端 ESLint 架构守卫 + architecture-validator + golangci-lint + 根路径端口/禁用端点审计全部阻断违规（.github/workflows/plan-255-gates.yml）。
- 最新一次根路径审计 `logs/plan255/architecture-root-20251117_133935.json` 显示 `totalViolations=0`，端口/禁用端点问题已回收；报告中 `passedFiles=325 / failedFiles=0`。
- 前端架构门禁报告 `reports/architecture/architecture-validation.json`（2025-11-20 12:34:58Z）同样显示 CQRS/Ports/Contracts 全部 0 违规；作为 Plan 255 的长期基线。
- Branch Protection 已将 `plan-250-gates`、`plan-253-gates`、`plan-255-gates` 设为 Required；gates‑255 成功 run 示例：<https://github.com/jacksonlee411/cube-castle/actions/runs/19403570892>（artifact: `plan255-logs`）。失败示例与设置描述记录于 `logs/plan255/branch-protection-required-checks.md`。

## 2. 主要变更
1. **根路径审计脚本**  
   - `node scripts/quality/architecture-validator.js --scope root --rule ports,forbidden` 新增硬门禁模式，统一日志落盘 `logs/plan255/audit-root-<ts>.log`。  
   - 评估期内（2025-11-16~17）多次运行收敛 37 个端口硬编码与 1 个禁用端点，最终快照零违规。
2. **端口/端点整改策略**  
   - Playwright/E2E 改用 `PW_BASE_URL` 与 `SERVICE_PORTS` 注入，不再硬编码 `:9090/:8090`。  
   - `/api/v1` 与 `/graphql` 仅通过代理入口访问；禁用端点（`/graphql/playground` 等）保留在工具层白名单，业务代码禁用。
3. **日志与证据**  
   - 根路径快照：`logs/plan255/architecture-root-20251116_163330.json`（整改前，36+ 违规）→ `logs/plan255/architecture-root-20251117_133935.json`（整改后 0）。  
   - 审计日志：`logs/plan255/audit-root-*`；Plan 215 中记录的 PR/issue 均已关闭。

## 3. 验收结论
- CI Required checks（plan-250/253/255）全部开启并通过多轮 run；Plan 255 现为阻断门禁。  
- 根路径端口/禁用端点清单清零，无需继续维护“整改任务表”；如未来出现新违规，将由 gates‑255 阻断并在 `logs/plan255/` 再次登记。
- 本计划的唯一事实来源迁移至本归档文件；如需追溯整改详情，可参考 `logs/plan255/**` 与 `docs/development-plans/215-phase2-execution-log.md` 对应章节。

## 4. 残余风险 / 监控
- 若新增第三方示例或工具脚本，请在脚本忽略列表中显式登记，防止误报。
- Architecture validator 默认忽略 `third_party/**` 与 `frontend/playwright-report/**`；如后续目录变更需同步更新。
