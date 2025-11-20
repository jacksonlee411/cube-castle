# PR 描述（Plan 222 关单草案）

标题：docs/acceptance: align Plan 222 with AGENTS; add Phase2 execution report (draft); mark Plan 221 done; update acceptance guidance

---

## 目的
- 推进并收敛 Plan 222 验收关单材料，确保与 AGENTS.md 强制约束一致（Docker/Make 唯一路径、单一事实来源）。
- 标注 Plan 221（Docker 集成测试基座）为已完成，并引用可审计证据。
- 补充 Phase2 执行验收报告（草案），作为最终 PASS 前置材料。

## 变更范围
- 文档
  - 更新 Plan 222 文档（修正本地开发/测试命令为 Make 目标、补充产物登记路径、调整阶段状态与证据引用）
    - docs/development-plans/222-organization-verification.md
  - 新增 Phase2 执行验收报告（草案）
    - reports/phase2-execution-report.md
- 脚本
  - 修正验收脚本 GraphQL 默认端口为 8090，统一与 CQRS（查询=8090）事实来源；保留 9090 单体挂载仅作历史兼容、非默认
    - scripts/plan222/run-acceptance.sh
- 不涉及代码与契约更改；仅为验收与合规文档补全。

## 一致性与约束对齐（AGENTS.md）
- Docker 强制：仅使用 `make docker-up`/`make run-dev`/`make test-db` 等入口；文档已移除 `docker-compose` 直调与 `go run` 启动示例。
- 单一事实来源：REST/GraphQL 回归基于 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`；证据产物落盘至 `logs/plan221/*`、`logs/plan222/*`。
- CQRS 边界：无变更；仅更新文档示例与指引。

## 验收证据
- 集成测试：logs/plan221/integration-run-*.log
- 健康/JWKS：logs/plan222/health-*.json、jwks-*.json
- REST：logs/plan222/create-response-*.json、put-response-*.json、acceptance-rest.txt
- GraphQL：logs/plan222/graphql-query-*.json
- 覆盖率：logs/plan222/coverage-org-*.{out,txt,html}
- E2E：logs/plan222/playwright-P0-*.log、playwright-FULL-*.log、playwright-LIVE-*.log

## 待办与门槛（不阻塞本 PR）
- 232（P0 稳定）Live 模式双浏览器全量 → 通过后将 222 状态切换为 ✅ 最终 PASS，并更新 reports/phase2-execution-report.md。
- 222B 性能基准复跑（完整参数）→ 更新指标章节与日志索引（logs/219E/perf-rest-*.log）。
- 覆盖率阶段目标（≥30%→55%→80%，按 255/256 路线补齐 repository/service/handler 高频与错误分支）。

## 回滚方案
- 文档型改动，若需回退，可整体回滚此提交；不影响运行时代码与契约。

## 与其他计划关系
- 依赖：Plan 232（E2E Live 与稳定性）；权限契约 Plan 252 已完成（门禁生效），不再作为阻塞项。
- 已完成：Plan 221 标记完成并引用日志证据。
