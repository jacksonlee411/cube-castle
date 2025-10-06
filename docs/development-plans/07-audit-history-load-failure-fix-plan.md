# Plan 07 — 审计历史页签“加载审计历史失败”修复计划

**计划类型**: 查询服务稳定性 / 数据一致性治理  
**创建日期**: 2025-10-06  
**优先级**: P1（生产可见功能异常）  
**责任团队**: 查询服务团队（Owner） / QA 团队（验证）  
**关联文档**: `CLAUDE.md`、`AGENTS.md`、`docs/api/schema.graphql`、`sql/inspection/audit-history-nullability.sql`、`reports/temporal/audit-history-nullability.md`

---

## 1. 触发背景
- Playwright 与手动回归在组织详情页审计历史标签观察到 “加载审计历史失败” 提示，GraphQL 返回空数组或 500。
- `docs/api/schema.graphql` 定义 `auditHistory(recordId: ID!): [AuditEntry!]!` 为非空返回，当前行为与契约不符。
- `sql/inspection/audit-history-nullability.sql` 与 `reports/temporal/audit-history-nullability.md` 为该缺陷提供现成的数据巡检脚本与模板，证明问题真实存在但缺少计划文档支撑。

## 2. 目标与验收标准
| 目标 | 验收标准 |
| --- | --- |
| 恢复契约一致性 | `auditHistory` 请求返回 200 且符合 GraphQL schema 非空约束；空结果需通过明确业务规则解释并更新契约。 |
| 数据完整性 | 巡检脚本结果中 `modified_fields`、`changes` 不再出现 NULL / 非 JSON 数组记录；`reports/temporal/audit-history-nullability.md` 填写最新统计并标注“0 条异常”。 |
| 端到端验证 | Playwright 场景（`tests/e2e/business-flow-e2e.spec.ts` 中审计步骤或单独脚本）通过；若无自动化剧本，至少保留手动验证记录。 |
| 文档同步 | 本计划、06 号日志、实现清单（若涉及调整）保持信息一致；若契约需更新须先修改 `docs/api/schema.graphql`。 |

## 3. 当前已知事实
- GraphQL 查询位于 `cmd/organization-query-service/internal/resolvers/audit_resolver.go`（通过结构化检索确认）；调用链包含服务与仓储层。
- 审计数据源自命令服务触发器 `log_audit_changes()` 及 `organization_audit` 表；Plan 18 Phase B 迁移可能已调整部分字段。
- 现有报告模板 `reports/temporal/audit-history-nullability.md` 仍为空白，需在本计划执行期间填充实际结果。

## 4. 行动计划

### Phase 0 — 立项与复现（0.5 天）✅ 已完成（2025-10-06）
1. ✅ 在本计划记录问题描述、实时进展与验收标准。
2. ✅ 复现实例与 GraphQL 查询验证：
   - 环境准备：通过 `make docker-up` 启动 Postgres
   - 生成有效 JWT：`go run scripts/cmd/generate-dev-jwt/main.go -audience cube-castle-users`
   - 执行 GraphQL 查询：
   ```graphql
   query($id: String!) {
     auditHistory(recordId: $id) {
       auditId recordId operation timestamp modifiedFields
       changes { field oldValue newValue dataType }
     }
   }
   ```
   - **查询结果**：成功返回 2 条记录，但发现数据质量问题（详见报告）
   - 完整响应已记录于 `reports/temporal/audit-history-nullability.md` 第 6 节
3. ✅ 执行 SQL 巡检：
   - 运行 `sql/inspection/audit-history-nullability.sql`
   - 输出保存于 `reports/temporal/audit-history-nullability-20251006.log`
   - **关键发现**：
     - 总审计记录数：2 条 (租户 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9)
     - changes NULL/非数组：0 条 ✅
     - 缺失 dataType 的条目：1 条 ⚠️
     - 空变更的 UPDATE 记录：1 条 ⚠️
   - 完整统计与样本已记录于 `reports/temporal/audit-history-nullability.md`

**Phase 0 总结**：
- ✅ 验证 GraphQL 端点可正常返回数据（非空数组），不存在"加载审计历史失败"的 500 错误
- ⚠️ 发现数据质量问题：dataType 缺失（返回 "unknown"）和空变更记录
- 📋 初步根因：审计触发器 `log_audit_changes()` 逻辑不完善
- ➡️ 下一步：Phase 1 定位触发器代码，分析 dataType 填充逻辑

### Phase 1 — 根因定位（1 天）
1. 跟踪查询服务代码：Resolver → Service → Repository，确认是否存在：
   - GraphQL 层未处理空结果 / 错误。
   - Repository SQL 过滤条件导致无结果。
   - 数据库 NULL 字段引发 JSON 解码失败。
2. 对照命令服务触发器与迁移：验证 `organization_temporal_current` / `organization_audit` 视图或函数在 Plan 18 Phase B 后是否仍写入完整数据。
3. 形成根因报告（含数据库与代码证据），附于本计划和 06 号日志。

### Phase 2 — 修复与验证（1-2 天）
1. 根据根因选择修复策略（示例）：
   - SQL 层：调整查询或补数据，确保 `changes`、`modified_fields` 始终为有效 JSON。
   - 代码层：补充错误处理/降级逻辑，同时保持 GraphQL 契约。
     - ✅ 2025-10-06 更新 `postgres_audit.go/sanitizeChanges`，在 `dataType` 缺失或为 `"unknown"` 时根据 old/new 值推断类型（如 `"string"`）。
   - 数据层：通过迁移或脚本补齐历史数据。
2. 补充单元测试 / 集成测试：
   - Go：`cmd/organization-query-service/internal/.../_test.go`
   - GraphQL：新增契约测试验证非空约束。
3. 执行验证命令：
   - `make test`
   - `make test-integration`
   - `npm --prefix frontend run test:e2e -- --grep "audit"`（如剧本存在）
   - ✅ `go test ./cmd/organization-query-service/internal/repository`（2025-10-06）确保 `sanitizeChanges` 推断逻辑通过单元测试。
4. 更新 `reports/temporal/audit-history-nullability.md`，注明修复后巡检结果；附 Playwright 或手动验证证据。
   - ✅ GraphQL 复测显示 `changes[].dataType` 由 `"unknown"` 修正为 `"string"`。

### Phase 3 — 归档与回溯（0.5 天）
1. 在 `docs/development-plans/06-integrated-teams-progress-log.md` 标记任务完成并总结风险。
2. 若需，更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 或实现清单中的相关条目。
3. 汇总执行记录至 `reports/iig-guardian/plan07-audit-history-<date>.md`（若创建）。
4. 验收完成后将本计划移入 `docs/archive/development-plans/`，并在 00-README.md 标记状态。

## 5. 风险与依赖
| 风险 | 描述 | 缓解措施 |
| --- | --- | --- |
| 数据受污染 | 历史审计记录存在不可恢复空值 | 在补数据前备份受影响记录；必要时与数据库团队协作。 |
| 查询性能退化 | 修复可能引入额外 JOIN/转换 | 通过 `EXPLAIN ANALYZE` 验证，必要时建立索引。 |
| 契约变更风险 | 若需调整 GraphQL 契约需跨团队评审 | 坚持“先契约后实现”，更新 `docs/api/schema.graphql` 并通知前端。 |

## 6. 里程碑
| 里程碑 | 内容 | 截止 |
| --- | --- | --- |
| Phase 0 完成 | 复现、巡检结果记录、根因假设 | 2025-10-07 |
| Phase 1 完成 | 根因报告 & 修复方案评审 | 2025-10-09 |
| Phase 2 完成 | 修复合入、测试通过、报告更新 | 2025-10-11 |
| Phase 3 完成 | 文档归档、日志更新、计划收尾 | 2025-10-12 |

## 7. 参考资料
- `docs/api/schema.graphql` — GraphQL 契约唯一事实来源。
- `sql/inspection/audit-history-nullability.sql` — 巡检脚本。
- `reports/temporal/audit-history-nullability.md` — 巡检输出模板。
- `reports/iig-guardian/p1-crud-issue-analysis-20251002.md` — 相关故障分析背景。
- `docs/development-plans/15-database-triggers-diagnostic-report.md`（归档）— 审计触发器历史问题，供参考。

---

**执行状态**: 🚧 Phase 0 已完成，Phase 1 待启动

**进度跟踪**：
- ✅ 2025-10-06: Phase 0 完成 - SQL 巡检与 GraphQL 查询验证完成，报告已更新
- 📋 待启动: Phase 1 - 触发器代码定位与根因分析

**Phase 0 关键成果**：
1. 确认 GraphQL 端点可正常返回数据，不存在 500 错误或空数组问题
2. 识别 2 个数据质量问题：dataType 缺失和空变更记录
3. 初步根因定位到审计触发器 `log_audit_changes()`
4. 完整证据已记录于 `reports/temporal/audit-history-nullability.md` 和巡检日志

进展与证据请在本计划与 06 号日志保持同步，确保资源唯一性与跨层一致性。
