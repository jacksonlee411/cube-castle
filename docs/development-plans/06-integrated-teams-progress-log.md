> 本文件已归档，不再作为唯一事实来源（SSoT）。请勿在此处追加或修改内容。  
> 当前权威来源：
> - P0 用例清单与门禁：`docs/archive/development-plans/232-playwright-p0-stabilization.md`、`docs/development-plans/232t-test-checklist.md`
> - 执行/进度汇总：`docs/development-plans/215-phase2-execution-log.md`
> - 职位域回归与运行手册：`docs/development-plans/240E-position-regression-and-runbook.md`
> - 权限契约与 PBAC 一致性（Plan 252）：`docs/archive/development-plans/252-signoff-20251115.md`

# Plan 06 – 集成测试验证纪要（已归档）

本文件已迁移至：`docs/archive/development-plans/06-integrated-teams-progress-log.md`（只读）。

如需执行或更新回归，请参考以上“当前权威来源”。

## 4. 文档治理与命名抽象（Plan 247）
- 完成文档与治理对齐（T5）：`Temporal Entity Experience Guide` 成为唯一事实来源；旧 Positions 指南路径在 reference 目录仅保留“Deprecated 占位符”（无正文）。  
- 证据已落盘：  
  - `logs/plan242/t5/rg-zero-ref-check.txt`（旧文档名零引用检查，排除 `docs/archive/**`）  
  - `logs/plan242/t5/document-sync.log`、`logs/plan242/t5/architecture-validator.log`（文档同步与架构守护运行日志）  
  - `logs/plan242/t5/inventory-sha.txt`（实现清单快照哈希）  
- 参考入口已在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 增补“Temporal Entity 命名与文档入口”，仅链接权威文档，未复制可变细节。
   - `OrganizationTemporalPage.tsx`、`PositionTemporalPage.tsx` 已替换为共享路由 + `PositionDetailView`。后续 Timeline/类型/selector 抽象将依赖此基线。

## 4. Plan 245 – Temporal Entity 类型 & 契约统一（结项纪要 · 2025-11-14）
- 结果：按“统一 Hook + 守卫冻结 + 渐进替换”策略交付，不引入破坏性契约变更。职位详情已切换统一 Hook；组织详情主从视图以统一 Hook 兜底名称/状态；operation 名在不改字段前提下统一为 `TemporalEntity*`（树查询保留测试敏感名不变）。
- 关键产物：
  - 统一类型/Hook：`frontend/src/shared/types/temporal-entity.ts`、`frontend/src/shared/hooks/useTemporalEntityDetail.ts`
  - operation 统一：Positions/Organizations/Audit/Tree 若干处改名；详情/版本/路径命名以 `TemporalEntity*` 为基线
  - 守卫：`scripts/quality/plan245-guard.js` + `reports/plan245/baseline.json`，冻结 `query PositionDetail/PositionDetailQuery` 新增使用
  - 契约注释：`docs/api/schema.graphql` 与 `docs/api/openapi.yaml` 增补 Plan 245 注释（索引统一命名，保持字段不变）
- 验证证据（均通过）：`logs/plan242/t3/31-frontend-codegen.log`、`32-implementation-inventory.log`、`33-architecture-validator.log`、`43/44/45/46/47/48/49/50`（Typecheck/Vitest）、`38-go-unit-tests.log`、`10-health-*.json`、`20-db-migrate-all.log`
- 后续跟踪（不阻塞关闭）：CI 接入 `npm run guard:plan245`；组织详情子组件逐步读取统一 record；统一更多 `TemporalEntity*` operation；OpenAPI 存量 `no-$ref-siblings` 错误独立修复

> 说明：GraphQL diff 阻塞已在 2025-11-08 通过 gqlgen runtime SDL 快照 + GraphQL Inspector 验证解除，详见上表与日志 `logs/219T5/graphql-inspector-diff-20251108-015138.txt`。

## 4. 待办清单
| 优先级 | 待办 | 说明 |
| --- | --- | --- |
| P0 | ✅ GraphQL 运行时已切换至 gqlgen，`graphql-inspector diff` 与 runtime SDL 快照无差（`logs/219T5/graphql-inspector-diff-20251108-015138.txt`） | Plan 06 第 3 节硬门槛已解除 |
| P0 | 恢复 business-flow/job-catalog/position-tabs/position-lifecycle/temporal-management 场景所需数据，Chromium 与 Firefox 全绿 | 满足退出准则第 1 条 |
| P1 | ✅ `logs/219E/rest-benchmark-20251107-140709.log` JSON 摘录已写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md:302-336` | 补全性能证据 |
| P1 | ✅ `docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 已于 2025-11-08 回填 Position CRUD 恢复详情（命令/时间戳/RequestId），后续如有新增执行需继续更新 | 保持唯一事实来源 |
| P2 | 更新 `frontend/test-results/app-loaded.png` 与最新 screenshots/trace/video 路径 | 对齐 Plan 06 §4 要求 |

## 5. 退出准则复核
- **Chromium/Firefox Playwright 全绿**：未满足（多场景失败）。  
- **GraphQL 契约 diff**：已通过 `npx graphql-inspector diff docs/api/schema.graphql logs/graphql-snapshots/runtime-schema.graphql`（`logs/219T5/graphql-inspector-diff-20251108-015138.txt`）。  
- **REST/性能脚本证据**：REST Node 驱动基线已写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md:302-336`，GraphQL/回退仍待补充。  
- **文档回填**：`docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md` 均需更新当前日志/结论。

> 结论：Plan 06 暂无法关闭，需完成上述 P0/P1 待办后重新评估。

## 6. 219E 重启前置条件推进（2025-11-08 10:30 CST）

| 项目 | Owner | 状态 | 说明 | 证据 |
| --- | --- | --- | --- | --- |
| 219E 文档更新（阻塞列表 + 前置条件表） | Codex + QA | ✅ 完成 | `docs/development-plans/219E-e2e-validation.md` 已记录 Docker 权限解除、Playwright/性能/回退等前置事项及日志来源 | `docs/development-plans/219E-e2e-validation.md` |
| Playwright P0 场景修复（business-flow、job-catalog、position-tabs、temporal-management） | 前端团队 | ⏳ 进行中 | 需恢复缺失的 data-testid、UI 文案与数据，完成后回填 `logs/219E/*.log` 与 `frontend/test-results/*` | `logs/219E/business-flow-e2e-*.log`、`logs/219E/job-catalog-secondary-navigation-*.log` |
| Position/Assignment 数据链路恢复 | 命令 + 查询团队 | ✅ 完成 | 230B/C/D 已交付 Job Catalog 迁移、自检脚本与播种 + Playwright 复验：`scripts/diagnostics/check-job-catalog.sh`、`scripts/dev/seed-position-crud.sh`、`npx playwright test tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium）。现可据此解锁 `position-lifecycle`/`organization-validator` 套件。 | `logs/230/job-catalog-check-20251108T093645.log`、`logs/230/position-seed-20251108T094735.log`、`logs/230/position-crud-playwright-20251108T102815.log` |
| Outbox/Dispatcher 指标验证 | 命令 + 平台团队 | ✅ 完成 | 2025-11-08 重新执行 Runbook O1-O6（`BASE_URL_COMMAND=http://localhost:9090 ./scripts/219C3-rest-self-test.sh`），`outbox_events` 成功写入 position/assignment/jobLevel 事件并被 dispatcher 发布，Prometheus 指标 + GraphQL 读模型同步刷新；Plan 231 已回填闭环记录 | `../archive/development-plans/231-outbox-dispatcher-gap.md`、`logs/219E/outbox-dispatcher-events-20251108T050948Z.log`、`logs/219E/outbox-dispatcher-sql-20251108T050948Z.log`、`logs/219E/outbox-dispatcher-metrics-20251108T051005Z.log`、`logs/219E/outbox-dispatcher-run-20251108T051024Z.log`、`logs/219E/position-gql-outbox-20251108T051126Z.log` |
| 性能基准回填（REST/GraphQL） | QA + SRE | ⏳ 待记录 | 借助 Node 驱动日志撰写对比并更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md`、219T 报告 | `logs/219E/rest-benchmark-20251107-140709.log`、`docs/development-plans/219T-e2e-validation-report.md:21-33` |
| 回退演练脚本与记录 | SRE + 后端 | ⏳ 待安排 | 依照 219D1/219D5 指南执行一次全量回退并归档日志，作为 219E 验收资料 | `logs/219D4/FAULT-INJECTION-2025-11-06.md`、`docs/development-plans/219D5-scheduler-docs.md` |

## 7. Plan 230 同步（2025-11-08 11:35 CST）

- ✅ **230E 文档更新完成**：`docs/development-plans/219T-e2e-validation-report.md` 与 `docs/development-plans/219E-e2e-validation.md` 增加 Position CRUD 恢复章节，记录命令、时间戳、RequestId 及 `frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` 产物，解除“Job Catalog 缺失”阻塞。  
- ✅ **230F readiness 输出完成**：`logs/230/position-module-readiness.md` 建立功能 × 测试映射，`frontend/tests/e2e/position-crud-full-lifecycle.spec.ts:362-384` 加入 `// TODO-TEMPORARY(230F)` 注记提示 `/positions/{code}/versions` 覆盖缺口；相关链接已写入 219E §2.4/§2.6。  
- 📌 **唯一事实来源**：Plan 230 母计划状态更新与本节互为引用，若后续扩充 Job Catalog 代码需在 Plan 06 中追加时间戳说明。
