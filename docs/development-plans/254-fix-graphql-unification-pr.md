# fix(plan254): unify GraphQL to monolith (:9090) and remove 8090 assumptions

Scope
- Plan 254 – 前端端点与代理整合（对齐 250A 运行时合流，GraphQL 由单体进程提供）

What
- 将前端与 E2E 的 GraphQL 端点统一至单体进程（:9090），移除对 legacy :8090 的健康检测与直连假设。
- 保持“单基址”访问：开发与 E2E 仅通过 `/api/v1` 与 `/graphql`；禁止直连端口（符合 AGENTS.md 端口与代理强约束）。

Key Changes
- frontend/src/shared/config/ports.ts：将 `QUERY_BASE/GRAPHQL_ENDPOINT/GRAPHQL_PLAYGROUND/METRICS_QUERY` 指向 `REST_COMMAND_SERVICE`（:9090）。
- frontend/tests/config/ports.ts：`TEST_SERVICE_PORTS` 不再包含 8090，仅健康检查 9090。
- frontend/tests/e2e/config/test-environment.ts：E2E 默认 `GRAPHQL_*` 改为 :9090。

Why
- 250A 已将 GraphQL 查询路由（/graphql）挂载到单体进程；前端与 E2E 不应再探测 8090。
- 254 CI 失败的根因是测试/配置仍假设存在独立 GraphQL 容器（:8090）。本 PR 收敛到 9090，消除历史遗留。

Acceptance
- Plan‑254‑gates：在 ubuntu/self-hosted 矩阵通过；工件包含 playwright 报告与 logs/plan254/*。
- 架构门禁：`architecture-validator` 在 `cqrs,ports,forbidden` 规则下 0 违规；无 `:9090/:8090` 直连（E2E 走单基址）。
- 不修改 docker-compose 端口映射（AGENTS 强制）；遇端口冲突按“卸载宿主服务”原则处理。

Notes
- 如仍存在用例层 TS/正则解析问题（evaluate + 模板字符串转义），后续 PR 将按文件/行号最小改动修复，不改变断言语义。

