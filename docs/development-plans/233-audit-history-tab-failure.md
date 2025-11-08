# Plan 233 — 审计历史页签加载失败调查

**编号**: 233  
**触发来源**: 组织架构与职位管理详情页的“审计历史”标签持续提示“加载审计历史失败”  
**关联模块**: Frontend AuditHistorySection、GraphQL `auditHistory(recordId)`、审计写入器（command service）  
**创建时间**: 2025-11-09  

---

## 1. 背景
- GraphQL 合约 `auditHistory(recordId)` 是唯一的审计读取入口（`docs/api/schema.graphql:253`），期望按 **recordId** 精确查询组织/职位/职类的时态版本审计轨迹。
- 前端 `AuditHistorySection` 组件（`frontend/src/features/audit/components/AuditHistorySection.tsx`）在组织与职位详情页复用：当 GraphQL 请求报错时，会渲染“加载审计历史失败”提示（同文件第 129-189 行）。
- 目前组织与职位页面都无法加载审计信息，说明服务端在处理 `auditHistory` 时返回了 GraphQL error（React Query `error` 分支被触发）。

---

## 2. 复现与影响范围
1. 在组织详情页 (`TemporalMasterDetailView`) 选择任意版本并切换到“审计历史”标签；组件直接显示“加载审计历史失败”，并提示重试仍旧失败（`frontend/src/features/temporal/components/TemporalMasterDetailView.tsx:118-153, 280-320`）。
2. 职位详情页 (`PositionTemporalPage`) 只要当前版本包含 `recordId`，同样通过 `AuditHistorySection` 调用 GraphQL，结果一致失败（`frontend/src/features/positions/PositionTemporalPage.tsx:592-631`）。
3. 问题影响所有依赖审计追溯的操作（版本核查、合规稽核、时态异常排查），属于高优先级可观测性缺陷。

---

## 3. 调查事实

### 3.1 前端调用链（事实来源：`frontend/src/features/audit/components/AuditHistorySection.tsx`）
- 使用 `unifiedGraphQLClient.request` 调用 `auditHistory`，必传 `recordId`，支持 limit/时间/操作人等过滤（第 69-117 行）。
- React Query 在 `error` 分支渲染 "加载审计历史失败"（第 129-189 行），说明 GraphQL 响应 `errors` 非空或 HTTP 非 2xx。

### 3.2 GraphQL 查询实现（事实来源：`internal/organization/repository/postgres_audit.go:74-125`）
- SQL 约束 `resource_id::uuid = $2::uuid` 且 `resource_type IN ('ORGANIZATION','POSITION','JOB_CATALOG')`。
- `resource_id` 字段在表结构中定义为 `character varying(100)`（`database/migrations/20251106000000_base_schema.sql:641-664`），并未强制为 UUID。
- Postgres 在评估 `resource_id::uuid` 时会对每一行进行强制类型转换；当列值不是合法 UUID 时立即抛出 `invalid input syntax for type uuid`，导致整个查询终止，GraphQL 返回 500。
- 同一文件中已经存在 `record_id uuid` 列（同一迁移文件第 641-664 行）和以 `record_id` 为索引的查询需求，却未在 SQL 中使用。

### 3.3 审计写入行为（事实来源：`internal/organization/audit/logger.go:514-556`）
- `LogOrganizationDelete` 在缺乏 `org` 详情时退化为 `resourceID = code`，并以 `TODO-TEMPORARY` 形式注明“会导致 UUID 类型错误，需在 2025-09-20 前回收”。
- 由于删除/异常场景都会触发该逻辑，`audit_logs.resource_id` 同时混入了 `1000002` 等组织/职位代码以及真正的 `recordId`。
- 这些非 UUID 值在 GraphQL 查询阶段被 `resource_id::uuid` 直接击穿，造成组织与职位所有审计查询失败。

### 3.4 现有一致性脚本佐证
- 仓库已提供 `scripts/fix-audit-recordid-misplaced.sql`、`scripts/validate-audit-recordid-consistency.sql` 等脚本，目标都是“确保 `record_id` 与 payload 内的 record_id 保持一致”。这进一步印证 GraphQL 查询应基于 `record_id` 而非 `resource_id`。

---

## 4. 根因总结
1. **查询层硬性将 `resource_id` 转 UUID** — `internal/organization/repository/postgres_audit.go:86-88` 对所有行执行 `resource_id::uuid`，但该列本质是 `VARCHAR`，存在大量非 UUID 数据。
2. **命令服务在删除/异常路径写入非 UUID** — `internal/organization/audit/logger.go:514-533` 在缺少组织快照时把 `resourceID` 填成 `code`，已知会产生“UUID 类型错误”但未按 TODO 回收。
3. **未使用 `record_id` 列** — 数据库已具备 `record_id uuid` 字段与索引，但查询依旧绑定在不可靠的 `resource_id` 上，没有 fallback 或数据修补逻辑。

结果：只要租户存在一条 `resource_id='1000002'`（或任何不可转 UUID 的值）的审计记录，所有 `auditHistory(recordId)` 请求都会因为 Postgres 的强制转换失败而返回 500，前端统一显示“加载审计历史失败”。

---

## 5. 修复计划

| 编号 | 动作 | owner | 说明 |
| --- | --- | --- | --- |
| A | **修正 GraphQL 查询条件** | Query Service | 改为 `WHERE record_id = $2::uuid`；若需兼容遗留数据，可追加 `OR (resource_id = $2 AND resource_id ~ '^[0-9a-f-]{8}-...$')`，禁止全列表强制转换。 |
| B | **数据清理 / 回填** | DB Ops | 运行 `scripts/fix-audit-recordid-misplaced.sql`，并补充 SQL 将 `resource_id` 为代码但 `record_id` 为空的行回填正确 UUID。 |
| C | **移除命令服务的代码 fallback** | Command Service | `LogOrganizationDelete` 必须从调用者获取 `recordId`；若无法获取，应阻止写入或先查询数据库，而不是写入 `code`。保留 `TODO-TEMPORARY` 的实现需立即替换。 |
| D | **端到端回归** | QA/Frontend | - `curl` / GraphQL Playground 验证 `auditHistory` 能返回数据<br/>- 打开组织/职位详情页验证 UI 进入“审计历史”标签不再报错。 |
| E | **自动化保护** | Backend | 为 `GetAuditHistory` 添加单元 / 集成测试，覆盖 `resource_id` 存在非法值时仍可查询 `record_id` 的场景；再追加 SQL 断言禁止将 `resource_id` 写成非 UUID。 |

---

## 6. 验收标准
- GraphQL `auditHistory(recordId)` 对至少一个组织和一个职位返回非空数组，无错误日志；Postgres 查询 plan 中不再包含 `::uuid` 对整列转换。
- `docker compose logs graphql-service` 中无 `invalid input syntax for type uuid`、`AUDIT_HISTORY_*` 错误。
- 前端组织与职位详情页的“审计历史”标签均可展示记录（可复用 Playwright `position-tabs` / Temporal 页面用例进行烟测）。
- 命令服务的审计写入在无 `org` 详情时会显式失败或补齐 `recordId`，并新增日志/指标监控。

---

## 7. 风险与后续
- 需要在数据修复前备份 `audit_logs`，避免批量更新造成历史追溯丢失。
- 若存在第三方依赖 `resource_id` 搜索，需要同步通知改用 `record_id`。
- 请在 `docs/reference/05-AUDIT-AUTH-GUIDE.md` 或相关参考文档中记录此次变更，确保唯一事实来源一致。

> 备注：执行完成后请将本计划状态同步到 `docs/development-plans/` 与相关归档（若全部交付完成，可移入 `docs/archive/development-plans/`）。
