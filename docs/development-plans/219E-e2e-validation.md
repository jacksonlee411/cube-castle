# Plan 219E – 端到端测试与性能验收

**文档编号**: 219E  
**关联路线图**: Plan 219  
**依赖子计划**: 219A~219D 全部完成  
**目标周期**: Week 5 Day 23-25（Day 26 作为缓冲、对齐 Plan 204 行动 2.9-2.10 后续验收）  
**负责人**: QA 团队 + 后端团队

---

## 1. 目标

1. 执行组织聚合的端到端回归（REST/GraphQL/Temporal/Audit/Validator 等全链路）。
2. 完成性能基准对比（重构前 vs. 重构后），确保 P99 延迟不退化。
3. 验证回退策略（rollback），确保在异常情况下可恢复。

---

## 2. 测试范围

### 2.1 当前执行矩阵（2025-11-06）

| 场景 | 脚本/入口 | 状态 | 说明 |
|------|-----------|------|------|
| Organization + Department 生命周期 | `scripts/e2e/org-lifecycle-smoke.sh` | ✅ 已执行 | `logs/219E/org-lifecycle-20251107-073626.log`、`...083118.log`、`...140705.log` 记录 REST/GraphQL 全链路成功 |
| Position + Assignment 流程 | `tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium） | ✅ 已恢复 | `logs/230/position-crud-playwright-20251108T102815.log` 记录 Step1-6 全部通过，产物归档 `frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` |
| Job Catalog 导入/导出 | `frontend/tests/e2e/job-catalog-*.spec.ts` | ❌ UI 阻塞 | `logs/219E/job-catalog-secondary-navigation-chromium-20251107-133841.log` / firefox 记录“编辑职类信息”标题缺失 |
| Outbox → Dispatcher → Query 缓存 | `scripts/dev/219C3-rest-self-test.sh` + 新增 PromQL | ⏳ 计划中 | 需在容器运行且 Prometheus/Grafana 指标可用时执行，等待与 219C3 owner 对齐 |
| 故障恢复（Temporal/Scheduler） | `scripts/dev/scheduler-alert-smoke.sh` + `logs/219D4/FAULT-INJECTION-2025-11-06.md` | ✅ 输入可复用 | 可直接复用 219D4 故障注入脚本验证 retry 与报警 |

### 2.1 端到端场景（至少覆盖以下 5 组）
1. **Organization + Department 生命周期**：创建 → 子部门 → 更新 → 层级移动 → 删除；验证 REST、GraphQL、Audit。
2. **Position + Assignment 流程**：创建职位 → Fill → Transfer → Vacate → 删除；验证 Temporal timeline 与 Assignment 查询。
3. **Job Catalog 导入/导出**：导入 → 校验版本 → 导出 → 对比 checksum；验证 validator 拦截冲突。
4. **Outbox → Dispatcher → Query 缓存**：更新组织数据，检查 outbox、dispatcher 指标、Query 缓存刷新日志。
5. **故障恢复**：模拟事务失败/dispatcher 失败/Temporal 中断，验证 retry 与报警。

### 2.2 性能测试
- REST P99：重点接口 `/api/v1/organizations`, `/api/v1/positions`, `/api/v1/job-family-groups`。目标：P99 不超过基线 +10%。
- GraphQL P95/P99：`organizations`, `positions`, `assignmentHistory`。目标：不退化。
- 资源消耗：CPU、内存、DB 连接数。

### 2.3 回退验证
- 定义回退步骤清单（恢复旧目录/适配层、切回旧配置）。
- 在测试环境演练一次回退 → 再恢复新结构。

### 2.4 重启前置条件（2025-11-07 更新）

| 任务 | 优先级 | 状态 | 说明与负责人 | 证据 |
|------|--------|------|--------------|------|
| Playwright P0 场景修复（business-flow、job-catalog、position-tabs、temporal-management） | P0 | 进行中 | 前端团队补齐缺失 data-testid、UI 文案与 GraphQL 数据，Chromium/Firefox 均需通过 | `logs/219E/business-flow-e2e-*.log`、`logs/219E/job-catalog-secondary-navigation-*.log`、`docs/development-plans/06-integrated-teams-progress-log.md:8-34` |
| Position + Assignment 数据链路恢复 | P0 | ✅ 完成 | 230B/C/D + 230F 产物（`scripts/dev/seed-position-crud.sh`、`logs/230/position-env-check-20251108T095108.log`、`logs/230/position-module-readiness.md`、`logs/230/position-crud-playwright-20251108T102815.log`）确认 Job Catalog、Position CRUD、功能→测试映射均可用，已作为 219E 重新开启 Position 场景的事实来源 | `logs/230/job-catalog-check-20251108T093645.log`、`logs/230/position-module-readiness.md`、`logs/230/position-crud-playwright-20251108T102815.log`、`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` |
| 性能与 REST 基准回填 | P1 | 待记录 | `scripts/perf/rest-benchmark.sh` Node 驱动日志已生成，需将 P50/P95/P99 摘要写入 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 并与旧基线对比 | `logs/219E/rest-benchmark-20251107-140709.log`、`docs/development-plans/219T-e2e-validation-report.md:21-33` |
| 回退演练计划 | P1 | 待安排 | 参考 219D1/219D5 回退指引，补充演练脚本与记录，确保 219E 验收可复用 | `logs/219D4/FAULT-INJECTION-2025-11-06.md`、`docs/development-plans/219D5-scheduler-docs.md` |
| Outbox/Dispatcher 指标验证 | P2 | ✅ 完成 | 2025-11-08 13:11 CST 复测 Runbook O1-O6：`./scripts/219C3-rest-self-test.sh` 触发 Position/Assignment/JobLevel 命令，`outbox_events` 成功写入并由 dispatcher 发布，Prometheus `outbox_dispatch_total{result=\"success\"}` 出现 `assignment.closed/assignment.filled/position.created/jobLevel.versionCreated`，GraphQL 读模型展现最新 `assignmentHistory`。Plan 231 已据此关闭。 | `docs/development-plans/231-outbox-dispatcher-gap.md`、`logs/219E/outbox-dispatcher-events-20251108T050948Z.log`、`logs/219E/outbox-dispatcher-sql-20251108T050948Z.log`、`logs/219E/outbox-dispatcher-metrics-20251108T051005Z.log`、`logs/219E/outbox-dispatcher-run-20251108T051024Z.log`、`logs/219E/position-gql-outbox-20251108T051126Z.log` |

### 2.5 Playwright 阻塞明细（2025-11-07）

| 场景 | 失败表现 | 推测原因/需修复内容 | Owner | 证据 |
| --- | --- | --- | --- | --- |
| business-flow-e2e | `locator.getByTestId('temporal-delete-record-button')` 超时，删除阶段无法完成 | Temporal UI 已移除按钮或 data-testid；需补充 timeline 等待并对齐新按钮命名 | 前端 + Temporal/QA | `logs/219E/business-flow-e2e-chromium-20251107-133349.log`、`...-firefox-20251107-140221.log` |
| job-catalog-secondary-navigation | 点击“编辑当前版本”后标题 `编辑职类信息` 未渲染 | Modal 标题/组件重命名或 GraphQL 数据缺失；需同步 UI 结构并确保 job catalog fixture 就绪 | 前端（Job Catalog） | `logs/219E/job-catalog-secondary-navigation-chromium-20251107-133841.log`、`...firefox-20251107-134321.log` |
| position-tabs | `getByTestId('position-temporal-page')` 不可见，多个 tab 缺失 | 230D 复跑（`logs/230/position-crud-playwright-20251108T102815.log`）已证明 `position-temporal-page` 恢复，需要重新执行 `tests/e2e/position-tabs.spec.ts` 更新日志 | 前端（Position） + QA | `logs/219E/position-tabs-20251107-134806.log`、`logs/230/position-crud-playwright-20251108T102815.log` |
| position-lifecycle | 标题 `职位管理（Stage 1 数据接入）` 未找到 | 缺失已由 Plan 230 修复，`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/` 可见完整 CRUD；需重新跑 `tests/e2e/position-lifecycle.spec.ts` 获取最新证据 | 同上 | `logs/230/position-crud-playwright-20251108T102815.log` |
| temporal-management-integration | 搜索输入 placeholder 不匹配/页面未渲染 | UI 文案变更 + 组织列表加载慢；需统一 locator 并等待 GraphQL 返回 | 前端（Temporal Dashboard） | `logs/219E/temporal-management-integration-20251107-135738.log` |
| name-validation-parentheses | ✅ 2025-11-08 复测通过（REST/GraphQL 均返回 200） | 已在 `tests/e2e/name-validation-parentheses.spec.ts` 中补齐 JWT + `X-Tenant-ID` 请求头，并使用 `BASE_URL=http://localhost:9090` 实测成功 | 前端 + API 契约 | `logs/219E/name-validation-parentheses-20251108T052717Z.log` |
| optimization-verification-e2e | bundle size 4.59 MB 超出 4MB 阈值 | 构建体积目标需重新评估或优化 JS chunk | 前端 Perf | `frontend/test-results/optimization-verification-e2e-*/trace.zip`（日志同目录） |

> 以上条目需在前端仓库与测试固化前同步更新 data-testid registry、契约与 Job Catalog 基础数据；完成后将新日志、trace/video 回填至 `logs/219E/` 与 `frontend/test-results/`。

### 2.6 Position + Assignment 数据恢复计划

1. **唯一事实来源**：`docs/development-plans/230-position-crud-job-catalog-restoration.md`、`database/migrations/20251107123000_230_job_catalog_oper_fix.sql`、`logs/230/position-module-readiness.md`、`logs/230/position-crud-playwright-20251108T102815.log`。
2. **诊断脚本**：运行 `scripts/diagnostics/check-job-catalog.sh`，确认 `OPER` Job Family/Job Function/Job Catalog 三层均为 `ACTIVE`；输出归档 `logs/230/job-catalog-check-20251108T093645.log`，为 `make status` 的 Job Catalog 子检查提供证据。
3. **数据补种**：
   - 若缺失，执行 `make db-migrate-all` 以应用最新迁移；必要时按 Plan 230 说明回灌 `backup_other_organizations_full.sql`。
   - 使用 `scripts/dev/seed-position-crud.sh`（Plan 230 产出）或 REST API 手动创建职位 + Assignment，确保 Temporal timeline 与 Query 读模型感知变更；请求/响应写入 `logs/230/position-seed-20251108T094735.log` 与 `logs/219E/position-seed-*.log`。
4. **查询链路验证**：通过 `graphql-client --endpoint=http://localhost:8090/graphql --query-file tests/e2e/fixtures/positions.graphql` 校验 `positions`、`assignmentHistory` 返回新数据，并记录 `logs/219E/position-gql-*.log`；必要时参考 readiness 表中的功能→测试映射，确认断言范围。
5. **验收标准**：`tests/e2e/position-crud-full-lifecycle.spec.ts`（Chromium：`frontend/test-results/position-crud-full-lifecyc-5b6e484b-chromium/`）提供最新 RequestId；当 `tests/e2e/position-lifecycle.spec.ts`、`tests/e2e/organization-validator/*.spec.ts` 在 Chromium/Firefox 真实环境下全绿时，将日志/trace 回填至 `logs/219E/position-lifecycle-*.log` 并在 Plan 06 Section 6 及本文件 2.4 表格勾选完成。
6. **功能对齐**：参考 `logs/230/position-module-readiness.md` 的功能 × 测试映射，若 Playwright 覆盖未交付功能（如 Transfer、Assignments 子路由），需在测试中添加 `// TODO-TEMPORARY(230F)` 注记并登记在 Plan 06 待办，确保断言与实现保持一致。

#### 2.6.1 Outbox / Dispatcher 取证记录（2025-11-08 11:21 CST，详见 Plan 231）

- **触发脚本**：`./scripts/219C3-rest-self-test.sh BASE_URL_COMMAND=http://localhost:9090 REQUEST_PREFIX=outbox-e2e-20251108T112139`。日志摘录存于 `logs/219E/outbox-dispatcher-events-20251108T112139.log`，对应 requestId：`d27a0d72-9506-450d-8b4d-b3d621bc610d`（headcount exceeded）、`7b52d81d-d97a-4f25-91d3-0123f94303fe`（assignment close）、`fbcf86a1-6178-4b04-8be9-cbb25b13b393`、`15cc7502-9aa4-468e-b11c-7fcdd23f66dc`、`e12520fd-4e63-4ffe-be6c-507022ffacda` 等。
- **数据库取证**：`psql` 输出 `logs/219E/outbox-dispatcher-sql-20251108T112236.log`，当前 `outbox_events` 查询为空（即便命令执行成功，未观察到 pending 或已发布记录），需与 217/230 负责人确认写入链路是否尚未启用。
- **Prometheus 指标**：`logs/219E/outbox-dispatcher-metrics-20251108T112459.log` 显示 `outbox_dispatch_{success,failure,retry}_total=0`、`outbox_dispatch_active=0`；意味着 dispatcher 已启动但未记录派发事件。
- **服务日志**：`logs/219E/outbox-dispatcher-run-20251108T112541.log` 捕获 `outbox dispatcher started`、`✅ Outbox dispatcher 已启动`，确认组件已随命令服务启动。
- **GraphQL 验证**：`logs/219E/position-gql-outbox-20251108T112820.log` 查询 `position(code: \"P1000032\")`，显示 `assignmentHistory` 已进入 `ENDED` 状态、`currentAssignment=null`，验证 Query 读模型能看到自测脚本造成的变更。
- **结论**：Runbook O1-O6 已完成，但由于 Outbox 表与 dispatcher 指标均为 0，怀疑命令链路尚未真正写入 outbox；需在下一轮执行前明确是否存在漏接或配置问题（或由其他计划负责的后续建设），否则 219E 的 Outbox 验收无法关闭。

---

## 3. 工具与脚本

- API 测试：Postman collection / curl / scripts in `tests/organization/`。
- GraphQL：GraphQL Playground 脚本或 CLI（`graphql-client`）。
- 性能：`hey` 或 `ab`；必要时使用 k6。
  - 例：`hey -n 1000 -c 20 -H "Authorization: Bearer $JWT" http://localhost:9090/api/v1/organizations`
  - 例：`graphql-client --endpoint=http://localhost:8090/graphql --query-file tests/organization/perf/positions.graphql`
  - 对比脚本：`scripts/perf/compare-benchmark.sh baseline.json current.json`（需输出差异报告）
- Temporal：`tctl workflow describe` 验证 workflow 结果。
- 监控：Prometheus/Grafana 面板（记录 P99、CPU、DB 连接）。

---

## 4. 验收标准

- [ ] 端到端场景执行完毕，结果记录于测试报告（预期路径：`logs/219E/org-lifecycle-*.log` 等）。
- [ ] 性能基准与重构前对比，P99 差异在可接受范围内；若超出，提供优化方案或评估。`scripts/perf/rest-benchmark.sh` 负责采集数据。
- [ ] 回退演练完成并记录步骤（参考 219D1 回退指引 + `make run-dev SCHEDULER_ENABLED=false` 演练）。
- [ ] 测试脚本/工具更新入库（本次新增 `scripts/e2e/org-lifecycle-smoke.sh`、`scripts/perf/rest-benchmark.sh`），并在 `internal/organization/README.md` 的“测试与验收”小节记录引用。
- [ ] 发布测试结论（通过/阻塞、剩余风险），同步到 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 性能章节。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 端到端场景覆盖不足 | 高 | 与业务/架构确认清单；回顾历史缺陷 |
| 性能指标退化 | 高 | 预留性能优化时间；收集指标分析瓶颈 |
| 回退流程复杂 | 中 | 提前拟定脚本；演练前确认恢复点 |

---

## 6. 交付物

- 端到端测试报告（包含场景、结果、日志链接，统一归档至 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 性能与验收章节）
- 性能基准报告（对比表 + 图表，同步更新至上述章节）
- 回退演练记录（补充到 `internal/organization/README.md` 的回退小节）
- 阶段结论（供上线/合并决策）

---

## 当前阻塞（2025-11-07 更新）
- Docker 权限问题已在 `docs/development-plans/06-integrated-teams-progress-log.md:3-7` 中验证解除，`make docker-up && make run-dev` 可正常启动 PostgreSQL/Redis/REST/GraphQL。
- 仍需恢复 Playwright P0 场景与 Position/Assignment 数据链路，详见 `docs/development-plans/06-integrated-teams-progress-log.md:8-34`；未完成前 219E 无法进入最终验收。
- 性能/REST 基准 JSON 摘要尚未公开到 `docs/reference/03-API-AND-TOOLS-GUIDE.md`，需以 `logs/219E/rest-benchmark-20251107-140709.log` 为事实来源回填（`docs/development-plans/219T-e2e-validation-report.md:21-33`）。
- 回退演练尚未执行，需按照「重启前置条件」表安排脚本、责任人与日志归档。
