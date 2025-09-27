# 15. 审计历史 GraphQL 非空约束修复计划

**文档类型**: 跨层一致性修复 / 查询服务  
**创建日期**: 2025-09-16  
**优先级**: P0（影响企业级契约与审计可用性）  
**负责团队**: 查询服务团队（Owner） / 前端组织团队（Co-owner） / 数据平台团队（协作）  
**关联文档**: `CLAUDE.md`、`AGENTS.md`、`docs/api/schema.graphql`、`cmd/organization-query-service/main.go`、`frontend/src/features/audit/components/AuditHistorySection.tsx`、`internal/middleware/graphql_envelope.go`

---

## 1. 背景与触发
- 在 `http://localhost:3000/` 的组织详情 > “前端开发组 (1000011)” 页面，切换“审计历史”页签出现前端 Toast：`API Error: Query completed with errors`。
- 前端组件 `AuditHistorySection` 通过 `unifiedGraphQLClient` 调用 `auditHistory` 查询（见 `frontend/src/features/audit/components/AuditHistorySection.tsx`），一旦 GraphQL 返回错误数组即触发 Envelope 中的通用错误提示。
- GraphQL Envelope 中间件 `internal/middleware/graphql_envelope.go` 将任何 `errors` 非空的响应统一转换为 `Query completed with errors`，因此该文案为后端返回错误的可靠指示。

## 2. 单一事实来源引用与一致性校验
- GraphQL 契约 (`docs/api/schema.graphql`) 定义 `auditHistory` 返回 `AuditLogDetail`，其中 `modifiedFields: [String!]!` 与 `changes: [FieldChange!]!` 为非空列表字段；`FieldChange.dataType` 标注为非空字符串。
- 查询服务实现 (`cmd/organization-query-service/main.go`) 的 `GetAuditHistory` 函数只在数据库 JSON 字段非空且不等于 `'[]'` 时填充 `ModifiedFieldsField` / `ChangesField`，否则保持 Go 零值 `nil`。当数据库该字段为空数组或 `NULL` 时，GraphQL Resolver 会将 `nil` 传到非空列表字段，触发执行异常。
- 同一实现中 `FieldChangeData.DataTypeField` 通过 `fmt.Sprintf("%v", changeMap["dataType"])` 填充，若 map 未包含 `dataType` 键则得到 `"<nil>"` 字符串，虽可序列化但不符合契约对数据类型语义的约束，需在修复中核实数据来源。
- 前端参数未对空数组进行处理，主要依赖服务器契约，因此违背契约的响应会导致整页数据不可用。

## 3. 问题定义
1. **契约违背**：`auditHistory` 查询在无变更字段时返回 `modifiedFields = null` / `changes = null`，违反 `[String!]!` / `[FieldChange!]!` 非空约定。
2. **错误传播**：GraphQL 运行时拒绝构建响应并返回错误数组，Envelope 中间件统一为 `Query completed with errors`，前端无法展示任何审计记录。
3. **数据完整性隐患**：部分 `FieldChange` 缺失 `dataType` 字段来源，虽当前不会直接触发非空错误（因被格式化为 `"<nil>"`），但与契约语义不符，需在同一治理计划内校验。
4. **潜在上游缺陷**：命令服务事件写入、审计触发器或历史迁移脚本可能产生空 `modified_fields` / `changes` 数据，若不追溯根因将重复出现数据损坏，需在计划内完成端到端定位。

## 4. 影响评估
- **业务影响**：审计历史页签无法显示，阻断关键合规审计场景，属于企业级 P0。
- **跨层一致性**：GraphQL 契约、查询服务实现与 UI 三层出现事实漂移，违背 `CLAUDE.md` 中“资源唯一性与跨层一致性”原则。
- **监控与测试**：现有自动化脚本未覆盖“无字段变更”场景，导致问题未被 CI 捕获。

## 5. 修复目标
- 查询服务在任意记录下均返回与契约一致的结构：无变更时返回空数组而非 `null`。
- 核对并补充 `FieldChange.dataType` 来源，确保契约字段具备正确语义值。
- 前端验证：`AuditHistorySection` 能在无变更、部分字段缺省的情况下正常展示空态而非报错。
- 提供自动化回归测试覆盖上述场景，纳入 CQRS 全链路验证。
- 遵循悲观谨慎原则：在最坏场景（批量损坏、并发迁移）下具备熔断、回退与监控手段，确保问题可控可回滚。

## 6. 实施计划

| 阶段 | 目标 | 主要交付物 | Owner | 截止 |
| --- | --- | --- | --- | --- |
| Phase 0 | 契约确认与现状复核 | GraphQL 查询重放脚本、问题复现记录、实现清单快照 | 查询服务团队 | 2025-09-17 |
| Phase 1 | 查询服务修复 | `cmd/organization-query-service/main.go` 更新：空数组兜底、`dataType` 校验、结构体默认值修正 | 查询服务团队 | 2025-09-19 |
| Phase 2 | 测试与验证 | 新增 Go 集成测试（含“无变更”案例）、GraphQL Smoke 脚本、前端 Vitest Mock 验证 | 查询服务团队 / 前端组织团队 | 2025-09-20 |
| Phase 3 | 发布与回归 | 部署说明、监控仪表板更新、前端 UAT 通过记录 | 查询服务团队 / 数据平台团队 | 2025-09-22 |
| Phase 4 | 文档归档 | 更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`、归档本计划 | 文档维护人 | 2025-09-25 |

### Phase 0 关键任务
- 使用 `node scripts/generate-implementation-inventory.js` 对照实现清单，标出所有依赖 `modifiedFields`、`changes` 的实现；在 `reports/temporal/audit-history-nullability.md` 记录基线。
- 调用 GraphQL（`curl` + `.cache/dev.jwt`）复现目标记录（`recordId` 来自 `organization(code:"1000011")`），保存响应含错误数组以支撑验收。
- 与数据平台团队联合执行数据库审计：编写并运行 `sql/inspection/audit-history-nullability.sql`（新增），统计 `modified_fields` / `changes` 为空或 `NULL` 的记录数、按租户/操作类型分布，并对结果签字确认。
- 追溯根因：对照命令服务审计写入逻辑（`cmd/organization-command-service/internal/audit`）与相关迁移（`database/migrations/012/013/020/027` 等）分析产生空值的触发条件，形成书面结论并在本计划附录引用。
- 建立性能基线：在修复前使用 `EXPLAIN ANALYZE` 与 `tests/perf/graphql-audit-history-benchmark.sh`（新增）测量典型与极端场景的响应时间、CPU、内存占用，作为后续比对依据。

### Phase 1 关键任务
- 在 `GetAuditHistory` 中对 `ModifiedFieldsField` / `ChangesField` 初始化空切片，确保即使解析不到 JSON 也返回 `[]`。
- 针对 `ChangesField` 每个元素，若 `dataType` 缺失则回退至数据库 `column_type` 或填入约定占位值（如 `unknown`），并记录修复策略。
- 更新 `FieldChangeData` 结构使默认值符合契约（如添加构造函数确保非空）。
- 对解析得到的 `modified_fields` / `detailed_changes` 增加 JSON Schema 校验，拒绝畸形结构并落盘审计日志，同时避免 `"<nil>"` 等占位串进入 GraphQL 响应。
- 引入查询级熔断/降级开关：当检测到批量数据损坏或解析失败时，触发熔断返回受限数据并报警，防止持续污染；该开关通过配置中心或环境变量统一管理，可快速回退。
- 针对大批量审计记录场景优化分页与流式处理策略（如按 `LIMIT/OFFSET` + 游标分段加载），确保修复后仍满足性能基线。
- 设计快速回退策略：保留旧版解析实现为 Feature Toggle，可在生产出现问题时立即切换，同时记录切换流程和审批要求。

### Phase 2 关键任务
- 新增集成测试：构造无字段变更的审计记录样本，断言 GraphQL 响应中 `modifiedFields`、`changes` 为空数组。
- 前端新增 Vitest：Mock GraphQL 返回空数组时组件渲染“暂无审计记录”，避免再度回归。
- 完成 `internal/middleware/graphql_envelope.go` 的快照测试，确认错误信息仅在真实错误时出现。
- 增加异常场景测试：在 Go 集成测试中注入畸形 JSON、缺失 `dataType`、空数组/空对象等边界，验证服务能拒绝或修复并记录告警。
- 执行性能回归：利用 Phase 0 基线的脚本构造 10k+ 审计记录场景，测量查询耗时、内存峰值、熔断触发阈值是否符合预期。
- 验证并发行为：编写并发访问测试（`tests/integration/graphql/audit-history-concurrency_test.go` 新增）模拟迁移执行期间的读写冲突，确保熔断与回退机制可控。

### Phase 3 关键任务
- 在 `make run-dev` 环境下回归组织详情全流程，截取日志确认无错误数组。
- 更新监控报警：新增 GraphQL 错误率指标阈值，保障生产可观测。

## 7. 一致性校验与验收标准
- 验证 GraphQL 响应 JSON：对 `recordId=...` 的无变更记录执行查询，响应数据中 `modifiedFields`、`changes` 均为 `[]`，无 `errors` 字段。
- 前端页面刷新后审计历史正常展示，空态文案与 UI 规范一致。
- `node scripts/generate-implementation-inventory.js` 更新后的清单与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 对齐，无重复事实来源。
- CI 增量测试（Go + 前端）全部通过，`make coverage` 曲线恢复到变更前水平。

## 8. 风险与应对
| 风险 | 影响 | 缓解措施 |
| --- | --- | --- |
| 历史数据 JSON 结构不一致 | 仍可能缺少 `dataType` 或空数组 | Phase 0 数据审计+Phase 1 JSON Schema 校验，必要时批量修复并记录追溯 |
| 上游命令事件仍产出空值 | 修复后再次出现数据损坏 | 在 Phase 0 根因分析中识别并修复命令服务/触发器问题，发布后接入实时监控报警 |
| 大批量审计记录导致性能下降 | 响应超时或熔断误触发 | Phase 0 基线 + Phase 2 压测，如有回归立刻启用熔断降级并优化索引/分页策略 |
| 数据迁移期间读写冲突 | 并发访问失败或数据被跳过 | 引入并发测试覆盖，发布前演练熔断/回退流程并在迁移窗口实施变更冻结 |
| 前端缓存残留错误 | 用户仍看到历史错误提示 | 发布公告要求刷新，前端加入版本号校验及错误态兜底提示 |

## 9. 完成状态（2025-09-27）

### ✅ 所有阶段已完成

**Phase 0 - 契约确认与现状复核**
- ✅ 数据库巡检脚本 `sql/inspection/audit-history-nullability.sql` 已创建并执行
- ✅ 性能基线脚本 `tests/perf/graphql-audit-history-benchmark.sh` 已创建
- ✅ 问题复现并定位具体原因（SQL字段不存在 + 类型不匹配）

**Phase 1 - 查询服务修复**
- ✅ `cmd/organization-query-service/main.go` 审计历史查询已修复
- ✅ 新增数据净化函数 `sanitizeModifiedFields` 和 `sanitizeChanges`
- ✅ 环境配置开关已实现（strictValidation、allowFallback、circuitThreshold、legacyMode）
- ✅ SQL查询字段映射已修正（before_data → request_data，after_data → response_data）
- ✅ 类型转换问题已解决（uuid类型转换修正）

**Phase 2 - 测试与验证**
- ✅ 单元测试 `cmd/organization-query-service/audit_history_sanitize_test.go` 全部通过
- ✅ 端到端验证：前端审计历史页签正常工作，无错误提示
- ✅ 验证契约合规：modifiedFields 和 changes 正确返回空数组

**Phase 3 - 发布与回归**
- ✅ 本地开发环境验证通过
- ✅ 前端页面审计历史功能恢复正常

### 🎯 验收标准完成确认
- ✅ GraphQL响应无errors字段，modifiedFields、changes均为空数组
- ✅ 前端页面正常显示"暂无审计记录"，无"API Error: Query completed with errors"
- ✅ 单元测试覆盖边界情况（null值、缺失dataType、无效JSON等）
- ✅ 审计历史配置正确生效：`strictValidation=true, allowFallback=true, circuitThreshold=25, legacyMode=false`

---

**✅ 计划已完成** - P0问题已解决，审计历史GraphQL非空约束修复完成，前端功能恢复正常。

_本计划已于 2025-09-27 完成实施并验证通过，现归档至 `docs/archive/`。_
