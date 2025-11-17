> 重要：远程分支策略为“主干开发 + 远程 PR 守卫”。远程 `master` 禁止直接推送；请通过短生命周期分支发起 PR（默认使用 squash-merge）。PR 必须关联 Issue 且通过所有 Required checks。

## 变更说明

- 概述本次变更的目的与范围：
- 是否涉及后向不兼容变更：
- 验证方式与可复现实验步骤：
- 关联 Issue（必填）：Closes #____

## 对齐 CLAUDE.md 与 AGENTS.md 的合规检查

- [ ] 诚实/谨慎：描述基于可验证事实，无夸大用语
- [ ] 临时方案：如有 `// TODO-TEMPORARY:`，已包含原因、计划、截止日期(YYYY-MM-DD)
- [ ] 新功能审批：新增 API/组件/服务/表 已获用户批准（附链接/证据）
- [ ] CQRS 分工：查询走 GraphQL，命令走 REST；未引入额外数据源
- [ ] API 命名：字段均为 camelCase；组织路由路径参数使用 `{code}`（未使用 `/{id}`）
- [ ] 契约命名自查：前端新增/修改字段均遵循 camelCase；无 `*_status`/`*_date` 等 snake_case 遗留
- [ ] 权限契约：相关端点权限已在 `docs/api/openapi.yaml` 定义并与实现一致
- [ ] 测试与验证：包含必要单测/集成/契约测试或最小验证
- [ ] 文档更新：READMEs/规范/变更日志已同步
- [ ] 实现清单：已运行 `node scripts/generate-implementation-inventory.js`（报告：`reports/implementation-inventory.json`），并用其校对本次导出/端点变更

## 文档治理与目录边界（Reference vs Plans）

- [ ] 已在 `docs/development-plans/` 建立/更新本次变更的计划/进展（完成后将归档至 `archived/`）
- [ ] 未将计划/进展类文档置于 `docs/reference/`（reference 仅保留长期稳定参考）
- [ ] 如新增参考类文档，已确认其长期稳定性并与 plans 分离
- [ ] `docs/README.md` 与各目录 `00-README.md` 的导航与边界说明保持一致

## CLAUDE 索引一致性检查（精简版原则落实）

- [ ] 未在 `CLAUDE.md` 增添易变内容（变更通告/流程清单/命令示例等）
- [ ] 易变内容按类型落位：
  - 变更通告/进展 → `CHANGELOG.md` 或 `docs/development-plans/`
  - 开发前必检/禁止事项/操作清单 → `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - API 一致性与工具细则 → `docs/reference/03-API-AND-TOOLS-GUIDE.md`
  - 文档治理与目录边界 → `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`、`docs/README.md`
- [ ] 如新增/调整 Reference 或 Plans 文档，已补充“权威链接”引用（`CLAUDE.md` / `AGENTS.md` / `docs/api/*` / 文档治理）

## 风险与回滚

- 风险点：
- 监控/告警：
- 回滚方案：

## 证据与工件（必填）

- Actions 运行/工件链接（例如 plan254-logs、playwright-report 等）：
- 本地/CI 关键日志片段（如需）：
