# Plan 222C – REST/GraphQL 契约对齐

编号: 222C  
上游: Plan 222（验证与文档更新）  
依赖: OpenAPI (`docs/api/openapi.yaml`)、GraphQL (`docs/api/schema.graphql`) 为唯一事实来源  
状态: 草案（待启动）

---

## 目标
- 全量核对 organization REST/GraphQL 响应与唯一契约（OpenAPI/GraphQL schema），确保字段/错误/分页/权限信息一致。
- 将所有差异分类（实现偏差、文档待补、前后端不一致），并在本子计划内完成修复或明确临时过渡（带 TODO-TEMPORARY 标签与回收日期）。

## 范围
- REST：`/api/v1/organization-units/*`、`/api/v1/positions/*` 等组织模块命令端点。
- GraphQL：`organizations`、`organizationSubtree`、`auditHistory` 等组织查询。
- 工具：OpenAPI / GraphQL schema diff、`scripts/plan222/run-acceptance.sh`、`node scripts/quality/contract-drift-report.js`（若需）。
- 证据：`logs/plan222/contract-alignment-*.md`（对齐记录）、API/GraphQL 响应样本。

## 不做
- 不擅自动更新契约文件；若契约需修订，需基于唯一事实来源讨论并通过规范流程。
- 不引入新的端点。

## 任务清单
1) 基线采集  
   - 执行 `scripts/plan222/run-acceptance.sh`、GraphQL 探针，收集最新 REST/GraphQL 响应样本。  
   - 将样本与 OpenAPI/GraphQL schema 对照，记录差异（字段命名、nullable、错误码、状态码等）。
2) 差异分类与整改  
   - **实现偏差**：更改后端实现或返回结构；增加测试保证（REST/GraphQL）。  
   - **契约陈旧**：若 schema 需更新，提出修改并在 MR 中说明证据。  
   - **临时方案**：如需暂时放宽（例如 232 未解锁的路径），使用 `// TODO-TEMPORARY(YYYY-MM-DD)` 标注，截止 ≤1 个迭代。
3) 自动校验  
   - 若可，用脚本（如 `scripts/quality/contract-drift-report.js`）生成 diff 并落盘 `logs/plan222/contract-drift-*.log`。  
   - 将关键断言纳入测试，防止回归。
4) 文档与索引更新  
   - `222-organization-verification.md` REST/GraphQL 验收项勾选。  
   - 在 `reports/phase2-execution-report.md` 中更新状态。

## 验收标准
- REST 与 OpenAPI 的字段、HTTP 状态、错误体一致，全量核对完成。
- GraphQL 响应/错误与 `schema.graphql` 一致，含分页/嵌套字段。
- 所有差异都有处理结论（修复或标注临时方案并登记回收日期）。
- 文档状态更新，相关证据落盘。

## 产物与落盘
- 对照记录：`logs/plan222/contract-alignment-*.md`
- 自动校验日志（若执行）：`logs/plan222/contract-drift-*.log`
- REST/GraphQL 响应样本：沿用 `logs/plan222/*response*.json`、`graphql-query-*.json`

## 回滚策略
- 修复过程中如需调整实现，确保有回滚路径（按 commit revert）；契约修改需双向评审，避免引入第二事实来源。

---

维护者: Codex（AI 助手）  
目标完成: Day 3（相对 222 收口节奏）  
最后更新: 2025-11-16 (草案)
