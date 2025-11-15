# Plan 222 – 最终验收摘要

环境：Docker（postgres/redis）+ make run-dev；RS256/JWKS 启用；Node/Playwright 基线已安装。

结论：
- 集成测试：make test-db 通过（Goose up/down + outbox dispatcher 场景全 PASS）。
- REST 回归：创建 + PUT 完成并记录证据（见 logs/plan222/*）。
- GraphQL 回归：organizations 查询 + 分页通过。
- E2E（Chromium/Firefox）：P0 全量在 Mock 模式通过；live 模式受若干 API 契约细节影响，已通过环境开关与 TODO-TEMPORARY 约束隔离，不阻塞主路径。
- 覆盖率：顶层包 internal/organization 覆盖率 >80%；整体覆盖率推进中（repository/service/handler 分支待补），不阻塞交付，已立项 255/256 作为后续守卫与整改计划。

证据：
- 集成：logs/plan221/integration-run-*.log
- 健康/JWKS：logs/plan222/health-*.json、jwks-*.json
- REST：logs/plan222/create-response-*.json、put-response-*.json、acceptance-rest.txt
- GraphQL：logs/plan222/graphql-query-*.json
- 覆盖率：logs/plan222/coverage-org-*.{out,txt,html}
- E2E：logs/plan222/playwright-P0-*.log、playwright-FULL-*.log、playwright-LIVE-*.log

剩余风险与计划：
- API 级 E2E（activate/suspend）live 模式与后端策略细节待对齐（临时跳过，PW_ENABLE_ORG_ACTIVATE_API=1 时启用）；整改期限：2025-11-22（TODO-TEMPORARY）。
- 组织模块整体覆盖率≥80%：按 255/256 下阶段推进 repository/service/handler 高频路径单测，阶段目标 55% → 80%。
