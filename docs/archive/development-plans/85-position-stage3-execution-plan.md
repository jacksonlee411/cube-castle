# 85号文档：职位生命周期 Stage 3 执行计划（含 Stage 2 尾项收束）

**版本**: v0.2 修订版  
**创建日期**: 2025-10-17  
**最新修订**: 2025-10-17  
**维护团队**: 命令服务团队 · 查询服务团队 · 前端团队 · QA 团队 · 架构组  
**状态**: 待复审（v0.2）  
**关联计划**: 80号职位管理方案 · 06号集成团队进展日志 · 84号 Stage 2 实施计划（归档）  
**唯一事实来源**:  
- `docs/development-plans/80-position-management-with-temporal-tracking.md`（Stage 0/1/2 里程碑与 Stage 3 目标）  
- `docs/development-plans/06-integrated-teams-progress-log.md`（Stage 2 完成情况与后续建议、评审意见）  
- `docs/archive/development-plans/84-position-lifecycle-stage2-implementation-plan.md`（Stage 2 验收细节）

---

## 1. 背景与目标

- Stage 2 已于 2025-10-16 完成，职位 Fill/Vacate/Transfer 流程稳定运行。  
- 80 号方案明确 Stage 3 为 **两周周期**，目标是补齐编制统计能力，同时完成空缺看板与转移界面。  
- 评审指出 85 号 v0.1 存在时间估算偏差、重复劳动及概念漂移。本次 v0.2 修订：  
  - 将 Stage 2 未完成的前端交互纳入 Stage 3 Week 1，不再引入 “Stage 2.5”。  
  - 以现存 GraphQL Schema/Repository 实现为基础，仅补齐缺失的 UI、聚合逻辑与测试。  
  - 保持两周交付节奏，与 80 号方案保持完全一致。

---

## 2. 前置核查记录（2025-10-17）

| 命令 | 目的 | 结果 |
|------|------|------|
| `grep -A20 "positionHeadcountStats" docs/api/schema.graphql` | 确认 GraphQL 契约 | 已定义 `positionHeadcountStats` 查询（含参数/类型说明） |
| `rg "GetPositionHeadcountStats" cmd/organization-query-service` | 确认查询服务实现 | PostgreSQL 仓储 `GetPositionHeadcountStats` 实现完备，Resolver 已接线 |
| `rg "vacantPositions" cmd/organization-query-service/internal/model` | 确认空缺查询模型 | 已存在连接模型、过滤与排序输入类型 |
| `node scripts/generate-implementation-inventory.js | grep -i headcount` | 检查实现清单 | 返回现有 headcount 相关条目，证实无需重复定义 |

> 核查结果表明：统计查询基础已存在，后续工作聚焦前端展示、聚合补强与测试覆盖。

---

## 3. 范围与交付物

### 3.1 必须交付（Blocking）
1. **空缺看板与转移界面（Stage 2 尾项）**  
   - 完成 `VacantPositions` 数据接入与 UI 实现，支持筛选/分页/空态。  
   - 实现 Position Transfer 操作界面（表单、权限校验、结果提示）。  
   - Playwright 补充对应 E2E 场景。
2. **编制与统计（Stage 3 核心）**  
   - 巩固 `positionHeadcountStats` 聚合逻辑（含组织范围、职位类型、职级维度）；必要时扩展缓存/失效策略。  
   - GraphQL 查询返回数据对齐前端需求（含 KPI、分布、可用编制）。  
   - 前端编制看板（图表、指标卡、导出能力），复用现有 Hook 模式。
3. **质量门禁与文档**  
   - 单元测试：仓储、Resolver、前端组件与 Hook。  
   - Playwright + Simplified E2E 补充统计/看板场景。  
   - 文档同步：实现清单、契约差异报告、06 号日志 Stage 3 条目、80 号计划 Stage 3 勾选。

### 3.2 可选交付（P2）
- 统计结果 CSV/Excel 导出。  
- 租户统计异常报警脚本。  
- Grafana 面板模板。

---

## 4. 执行节奏（两周计划）

| 周次 | 时间窗口 | 核心目标 | 责任团队 | 验收要点 |
|------|----------|----------|----------|----------|
| **Week 1（Stage 3.1）** | Day 1-5 | 完成空缺看板+转移界面，并补齐 headcount 聚合 | 前端 · 查询服务 · 命令服务 · QA | 空缺看板上线、转移界面可用、Headcount API 可调用、单元测试通过 |
| **Week 2（Stage 3.2）** | Day 6-10 | 完成编制看板成品、E2E/文档收束 | 前端 · QA · 架构组 | 编制看板完成、Playwright 场景通过、文档同步、计划归档准备 |

---

## 5. 任务拆解

### Week 1 — Stage 3.1：基础能力整合
- **查询服务**  
  - [x] 审视 `GetVacantPositionConnection` 与 `GetPositionHeadcountStats`；新增职种聚合 `byFamily` 并保留职位类型/职级统计（Go 实现见 commit 3a7e16b1）。  
  - [x] 增加租户隔离与 asOfDate 单元测试（`TestResolver_VacantPositions_ForwardsAsOfDate` 等，commit 3a7e16b1）。  
- **命令服务**  
  - [x] 核查 Fill/Vacate/Transfer 后 headcount 缓存逻辑：当前无本地缓存，统计实时读取 PostgreSQL，确认无需额外刷新机制（2025-10-17 评审记录）。  
- **前端**  
  - [x] 实现 `PositionVacancyBoard` 组件与 `useVacantPositions` Hook。  
  - [x] 实现 `PositionTransferDialog` 表单与流程（包含权限错误处理）。  
  - [x] 将 `PositionDashboard` 接入真实数据，保留 mock fallback。  
- **QA**  
  - [x] 更新 `frontend/tests/e2e/position-lifecycle.spec.ts`，新增空缺看板、转移入口与编制统计验证（2025-10-17，提交 `feat: add position headcount dashboard` 搭配路由更新）。  
  - [x] 补充 Vitest 覆盖新组件（`PositionHeadcountDashboard.test.tsx`）。

### Week 2 — Stage 3.2：编制看板与验收
- **前端**  
  - [x] 构建 `PositionHeadcountDashboard`（KPI 卡片、趋势图、分组表、导出按钮）。  
  - [x] 实现 `usePositionHeadcountStats` Hook，处理加载/错误/空态。  
  - [x] Vitest 覆盖导出、筛选、渲染逻辑（`PositionHeadcountDashboard.test.tsx`）。  
    - 2025-10-17 提交 `feat: add position headcount dashboard`（c2481957）交付上述 3 项。  
- **QA**  
  - [x] Playwright 新增编制看板验证（复用 `position-lifecycle.spec.ts`，校验空缺/编制区域）。  
  - [x] `simplified-e2e-test.sh` 补充统计验证步骤（2025-10-17 新增职位空缺/编制校验）。  
- **文档与同步**  
  - [x] 更新实现清单、契约差异报告（`node scripts/generate-implementation-inventory.js` 2025-10-17 再执行；GraphQL schema 同步）。  
  - [x] 06 号日志：新增 Stage 3 进展段落与验收总结。  
  - [x] 80 号计划：Stage 3 复选框更新并附脚注。  
  - 完成后准备将 85 号计划归档。

---

## 6. 里程碑与验收标准

| 里程碑 | 验收标准 | 核验方式 |
|--------|----------|----------|
| **M1 — Week 1 结束** | 空缺看板上线、转移界面可用、Headcount API 返回数据，单元测试通过 | `npm --prefix frontend run test:e2e -- tests/e2e/position-lifecycle.spec.ts --project=chromium`（含新增步骤）· `go test ./cmd/organization-query-service/internal/...` |
| **M2 — Week 2 中段** | 编制看板接入真实数据，导出功能可用 | `npm --prefix frontend run test -- PositionHeadcount*.test.tsx`（Vitest） |
| **M3 — Week 2 结束** | Playwright & Simplified E2E 全绿，文档同步完成，80/06 更新 | `npm --prefix frontend run test:e2e` · `./simplified-e2e-test.sh` · 文档审查 |
| **M4 — 计划归档** | Stage 3 验收总结存档，85 号计划移至 `docs/archive/...` | 归档记录 + 06 号日志验收条目 |

---

## 7. 风险与缓解措施

| 风险 | 级别 | 缓解策略 |
|------|------|----------|
| 空缺看板/转移界面继续延迟 | 高 | 将 Week 1 完成作为进入 Week 2 的前置门禁；未完成不得进入统计开发 |
| 统计查询在大租户场景性能不足 | 中 | 先基于现有 SQL 调优；必要时引入物化视图或缓存（需架构组评估） |
| Playwright 夹具与真实数据漂移 | 中 | Week 2 Day 1 开始使用真实 GraphQL 数据进行一次完整跑批，保留 fallback fixture |
| 导出功能引入新依赖 | 低 | 优先使用现有工具链（如 Papaparse）；若需新依赖须走依赖评审流程 |

---

## 8. 前置检查清单（执行前完成）

- [x] `node scripts/generate-implementation-inventory.js`（确认统计条目存在，避免重复造轮子） — 2025-10-17 执行，输出已记录。  
- [x] `grep -A20 "positionHeadcountStats" docs/api/schema.graphql`（核对 Schema 无漂移） — 2025-10-17 调整查询返回结构并复查。  
- [x] `make docker-up && make run-dev`（参考 `run-dev*.log` 记录；本次补充以 GraphQL/REST 自动化脚本验证核心接口）  
- [x] 复盘 Stage 2 Playwright Trace，标记可复用步骤（结果已同步至 `frontend/tests/e2e/position-lifecycle.spec.ts`）  
- [x] 在 `docs/development-plans/06-integrated-teams-progress-log.md` 添加 Stage 3 进度分节（通过后更新内容）

---

## 9. 退出标准

- 空缺看板与转移界面在 80 号计划中标记完成，并在 06 号日志记录上线结果。  
- 编制看板、统计 API、测试与文档全部完成并通过 QA 验收。  
- 本计划归档至 `docs/archive/development-plans/85-position-stage3-execution-plan.md`，同时在 80 号方案勾选 Stage 3。  
- 无新增 `TODO-TEMPORARY` 条目，若有临时方案须按 17 号流程登记并设定回收时间。

---

## 10. 附录

- Playwright 场景命令模板：  
  ```bash
  npm --prefix frontend run test:e2e -- tests/e2e/position-lifecycle.spec.ts --project=chromium
  npm --prefix frontend run test:e2e -- tests/e2e/position-lifecycle.spec.ts --project=firefox
  ```
- GraphQL 调试示例（空缺看板）：  
  ```bash
  curl -sS -X POST http://localhost:8090/graphql \
    -H "Authorization: Bearer $PW_JWT" \
    -H "X-Tenant-ID: $PW_TENANT_ID" \
    -H "Content-Type: application/json" \
    -d '{"query":"query Vacant($page:Int,$pageSize:Int){ vacantPositions(pagination:{page:$page,pageSize:$pageSize}) { data { position { code title organization { code name } } vacancySince headcountCapacity headcountInUse } pagination { total page pageSize hasNext } } }","variables":{"page":1,"pageSize":10}}'
  ```
- 统计查询草案：参见 `sql/positions/headcount-draft.sql`（若缺失，将在 Week 1 内创建并纳入迁移流程）。

## 11. Stage 3 验收补充

- GraphQL 解析与聚合：`go test ./cmd/organization-query-service/internal/graphql/...` 已验证 `positionHeadcountStats` 解析租户并输出家族维度。
- 前端统计报表：`npm --prefix frontend run test -- PositionHeadcountDashboard` 覆盖 CSV 导出与 byFamily 渲染；Playwright `tests/e2e/position-lifecycle.spec.ts` 断言包含 byFamily 表格，待集成流程跑批。
- 文档同步：06 号进展日志与 80 号方案即刻更新 Stage 3 收尾，计划归档步骤已准备。

---

*本修订版根据 2025-10-17 评审意见调整，待团队复审通过后生效。*
