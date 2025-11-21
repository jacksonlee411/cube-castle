# Plan 222D – E2E Live 完整性与日志链路

编号: 222D  
上游: Plan 222（最终验收）  
依赖: Plan 232（Playwright P0 双浏览器全绿解锁）、Plan 252（权限门禁）  
状态: 草案（待 232 解锁后启动）

---

## 目标
- 在 Playwright Live 模式中恢复并锁定组织模块 P0 全量（含 activate/suspend API 级用例），验证数据一致性、事件投递与日志链路。
- 确认 eventbus 事件、审计日志、API 侧日志三者一致并落盘证据，构成 222 最终验收的 Live 侧凭证。

## 范围
- 用例：`frontend/tests/e2e/*` 中与组织相关的 P0 规格（basic-functionality、simple-connection、organization-create、activate-suspend-workflow 等）。
- 配置：`PW_ENABLE_ORG_ACTIVATE_API=1`、`PW_TENANT_ID`、`PW_JWT`；Live 环境以 Docker 命令服务（9090）+ GraphQL `/graphql`。
- 证据：`logs/plan222/playwright-LIVE-*.log`、`logs/plan222/eventbus-*.json`、`logs/plan222/audit-*.json`。

## 不做
- 不在本子计划新增 UI 功能；仅聚焦 Live 流程验证与证据。
- 不跳过 CI 门禁；如需暂时豁免，严格按 `AGENTS.md` 的 TODO-TEMPORARY 规范记录。

## 任务清单
1) 环境就绪  
   - 等待 Plan 232 标注“Chromium/Firefox 双浏览器全绿”；确认 `.github/workflows/agents-compliance.yml` 无硬阻塞。  
   - 准备 JWT（`make jwt-dev-setup && make jwt-dev-mint`）并设置 `PW_ENABLE_ORG_ACTIVATE_API=1`。
2) 用例运行  
   - Live：Chromium/Firefox 各 ≥1 轮，记录 trace/screenshot。  
   - Mock：保留基线，确保对比 Live 的差异仅来自后端行为。  
   - 对 activate/suspend 用例收紧断言（HTTP 409 幂等、权限错误码 `INSUFFICIENT_PERMISSIONS`）。  
   - 产生日志：`logs/plan222/playwright-LIVE-*.log`、`trace.zip` 等。
3) 数据一致性校验  
   - REST/GraphQL 跟进：验证创建/更新后的树结构一致。  
   - 事件：`pkg/eventbus` 投递记录（可通过数据库查事务性发件箱或日志）。  
   - 审计：GraphQL `auditHistory` 对应组织 ID。  
   - 汇总到 `logs/plan222/live-audit-*.md`。
4) 风险与临时项  
   - 若发现 Live 与 Mock 差异，登记问题编号与预计修复时间。  
   - 临时放宽必须 `// TODO-TEMPORARY(YYYY-MM-DD)` 标注并在 `222` 文档列出。

## 验收标准
- Playwright Live P0 全量在 Chromium/Firefox 通过（含 activate/suspend）。  
- 数据一致性、事件、日志均与期望一致或有处置计划。  
- 相关证据完整落盘并在 `222-organization-verification.md` 勾选“数据一致性/事件/日志”项。

## 产物与落盘
- `logs/plan222/playwright-LIVE-*.log`、`*.trace.zip`  
- `logs/plan222/live-audit-*.md`（一致性记录）  
- 如需 eventbus/审计快照，分别落盘为 `eventbus-*.json`、`audit-*.json`

## 回滚策略
- 如 Live 用例阻塞，可临时关闭 `PW_ENABLE_ORG_ACTIVATE_API`（需记录 TODO-TEMPORARY，截止 ≤1 个迭代），确保主干可用；修复完成后重新启用并复验。

---

维护者: Codex（AI 助手）  
目标完成: Plan 232 解锁后 Day 2 内完成首轮，Day 3 完成复验  
最后更新: 2025-11-16 (草案)
