# Plan 222 – 最终验收摘要（阶段性刷新）

环境：Docker（postgres/redis）+ monolith（9090，挂载 /graphql）+ RS256/JWKS；Node/Playwright 基线已安装。

本次结论（2025-11-15 18:26）：
- 集成测试：make test-db 通过（Goose up/down + outbox dispatcher 场景 PASS）。
- REST 回归：创建 + PUT 完成并登记新证据（见下方列表）。
- E2E（Chromium/Mock）：Smoke 通过（6 passed / 1 skipped）。
- GraphQL：单体 /graphql 路由本地请求返回 404，但容器日志显示已注册。记录为后续修复项（不阻塞阶段性通过；历史基础用例已在上轮通过）。

证据（新增/本轮）：
- 集成：logs/plan221/integration-run-20251115_182410.log
- 健康/JWKS：logs/plan222/health-command-20251115-181959.json、jwks-20251115-181959.json
- REST：logs/plan222/create-headers-20251115-181959.txt、create-response-20251115-181959.json、put-response-1000505.json
- E2E Smoke：logs/plan222/playwright-P0-20251115-182431.log

剩余事项与计划：
- GraphQL 单体路由 404：在命令服务中已注册 POST /graphql；需排查路由优先级/分组与中间件链，补充结构化日志与探活用例。修复后补登记 `graphql-query-*.json`。
- 覆盖率阶段目标：按 255/256 持续推进（≥30% → 55% → 80%）。

状态：⏳ PARTIAL PASS（不阻塞主路径交付；按 232/252 与 GraphQL 路由修复后切换为 ✅）。

