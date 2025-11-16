# Plan 222 – 最终验收摘要（阶段性刷新 · GraphQL 路由通过）

环境：Docker（postgres/redis）+ monolith（9090，挂载 /graphql）+ RS256/JWKS；Node/Playwright 基线已安装。

本次结论（2025-11-15 20:59）：
- 集成测试：make run-dev 内置 Goose up 完成。
- 健康/JWKS：新增健康与 JWKS 证据已登记。
- GraphQL：在 9090 单体路由 /graphql 使用 Authorization + X-Tenant-ID 成功返回数据（见新日志）。
- REST 回归：创建 + PUT 成功（7 位数字 code，含 X-Tenant-ID 与 Authorization）。

新增证据：
- 健康/JWKS：logs/plan222/health-command-20251115-125838.json、jwks-20251115-125838.json
- GraphQL：logs/plan222/graphql-query-20251115-125943.json
- REST：logs/plan222/create-headers-20251115-130022.txt、create-response-20251115-130022.json、put-response-1031964.json

剩余事项与计划（仍为 PARTIAL PASS）：
- E2E Live（activate/suspend 等 API 级用例）仍按 232/252 守护开关默认跳过，待对齐后全量放开并登记 logs/plan222/playwright-LIVE-*.log。
- 覆盖率整体目标（≥80%）按 255/256 推进，阶段目标先达 ≥30%、≥55%。
- 222B 性能完整基准复跑并登记 logs/219E/perf-rest-*.log。

状态：⏳ PARTIAL PASS（核心路径通过；待 232/252 与覆盖率/222B 完成后切换为 ✅）。

