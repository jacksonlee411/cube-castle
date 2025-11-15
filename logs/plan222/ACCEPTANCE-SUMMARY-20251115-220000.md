# Plan 222 – 阶段性验收摘要（覆盖率达 30%）

环境：Docker（postgres/redis）+ monolith（9090，/graphql）+ RS256/JWKS；Node/Playwright 基线就绪。

本次结论（2025-11-15 22:00）：
- 覆盖率：internal/organization 组合覆盖率提升至 30.0%（由 23.6% → 24.6% → 26–29% 递进）；证据：logs/plan222/coverage-org-20251115-135303.txt
- GraphQL（9090 /graphql）：带 Authorization + X-Tenant-ID 能稳定返回数据；证据：logs/plan222/graphql-query-20251115-125943.json
- REST 回归：创建 + PUT 成功（7 位 code + X-Tenant-ID）；证据：logs/plan222/create-response-20251115-130022.json、put-response-1031964.json

新增/关键证据：
- 覆盖率：logs/plan222/coverage-org-20251115-135303.{out,txt,html}
- DevTools/Operational/JobCatalog 路由与响应函数覆盖：新增多组 handler/middleware/utils 单测；参见新增 *_test.go
- GraphQL/REST：同前摘要与 18:26 批注一致

剩余事项：
- E2E Live（API 级用例）待 232/252 对齐后放开，登记 logs/plan222/playwright-LIVE-*.log
- 222B 性能完整基准复跑并登记 logs/219E/perf-rest-*.log
- 覆盖率下一阶段：目标 55%（聚焦 repository/service/handler 高频及负路径）

状态：⏳ PARTIAL PASS（核心路径通过；覆盖率达到阶段目标 30%）。

