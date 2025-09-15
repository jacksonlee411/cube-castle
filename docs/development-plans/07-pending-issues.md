# 07 — 组织时态与审计分工报告（替换原 Pending Issues）

最后更新：2025-09-15  
维护团队：后端组（主责）+ 架构组 + QA组  
状态：报告定稿（分工确定、迁移与门禁已接入；按本报告执行与验证）

—

## 执行摘要
- 背景：历史上由数据库触发器承担“邻接修补、endDate 计算、审计写入”等业务，INSERT 会连锁触发多次 UPDATE，并为每次 UPDATE 产生日志，导致同一 recordId 的审计条目偏多、显示“混杂”。
- 分工决策：查询统一 GraphQL；命令统一 REST。时间轴修补、当前/未来判定、审计写入统一由“应用层（命令服务）”接管；数据库仅负责结构性约束、索引与最小技术性触发（非业务）。
- 当前成果：
  - 空 UPDATE（严格口径）= 0；recordId 归属一致；历史条目偏多属历史累积。
  - 迁移 021/022 与 CI 审计门禁已接入，确保“仅值变更才更新、移除目标触发器、审计一致性”。
- 验收：新写入应收敛至“1 CREATE + ≤1 有效 UPDATE（有未来版本时）”；PR 必须通过审计一致性门禁。

—

## 1. 分工原则（CQRS 与单一数据源）
- 查询（GraphQL, 8090）：只读，承载时态查询、层级与统计。  
- 命令（REST, 9090）：写入，承载“邻接修补 + 标志重算 + 审计写入”完整业务事务。  
- 数据源：PostgreSQL 单一事实来源；禁止引入 CDC/额外数据源。
- 命名与契约：API 对外 camelCase；先契约（OpenAPI/GraphQL），后实现；权限以 OpenAPI 为准。

—

## 2. 时态判定与业务规则
- 当前记录（current）：`effective_date <= asOf < COALESCE(end_date, +∞)`（asOf 缺省为“今天”）。
- 未来记录（future）：`effective_date > asOf`。
- 历史记录（historical）：`COALESCE(end_date, -∞) < asOf`。
- 相邻与区间：
  - 后一版本生效日为 `E`，则前一版本 `end_date = E - 1 day`。
  - 单一当前：任一日每个 `{tenantId, code}` 至多 1 条当前。
  - 禁止重叠：同一 `{tenantId, code}` 的 `[effective_date, end_date]` 闭区间不重叠（端点可相接）。

—

## 3. 应用层职责（写路径接管）
- 事务与并发：
  - 单事务完成邻接修补、endDate 计算、`is_current` 重算与审计写入。
  - 对同一 `{tenantId, code}` 使用 `pg_advisory_xact_lock(hashtext(tenant||code))` 并发互斥。
- 典型流程：
  - 插入中间版本（生效 E）：修补 `prev.end_date = E-1d`；计算 `curr.end_date = min(next.effectiveDate-1d)` 或 NULL。
  - 生效日变更：等价“旧位置删除 + 新位置插入”或 UPDATE，同时修补前/后邻与重算 `is_current`。
  - 软删/恢复：维护业务状态与时间轴，并保持“当前唯一、无重叠”。
  - 仅值变更才 UPDATE：所有 UPDATE 使用 `IS DISTINCT FROM`；无实质变化直接跳过（避免空 UPDATE）。
- 审计写入：
  - 审计表 `audit_logs`：`event_type`、`before_data/after_data`、`changes`、`modified_fields`、`record_id`、`business_context`。
  - UPDATE 无字段变化不记审计；邻接修补的 UPDATE 归属“邻接记录的 record_id”。
  - 可设置 `SET LOCAL app.request_id`、`SET LOCAL app.context`（如 `system-auto-enddate`）。

—

## 4. 数据库职责（去触发器化）
- 结构与约束：
  - 主键：`record_id UUID`；分区键/隔离：`tenant_id`。
  - 时态列：`effective_date DATE`、`end_date DATE`、（可选）`is_current BOOLEAN`（由应用维护/导出用）。
  - 约束与索引：唯一/部分唯一（当前唯一）、父子层级有效性、时态检索索引、路径/层级索引等。
- 触发器策略：
  - 移除业务触发器（审计/自动 endDate/生命周期/软删标志等）；避免连锁副作用。
  - 可保留非业务的技术性触发（如维护更新时间戳），不参与时间轴与审计业务。
- 审计表结构治理：统一字段，历史数据的回填/修复仅以一次性脚本完成，不常驻触发。

—

## 5. 历史问题与根因（结论）
- 现象：插入中间版本时，由 BEFORE/AFTER 多级触发器造成 1 次 INSERT → 多次 UPDATE；每次 UPDATE 触发审计 → 条目倍增；聚合口径正确但被放大为“条数偏多/混杂”。
- 根因：数据库触发链承担业务流程，产生连锁更新与多次审计写入。

—

## 6. 修复方案与迁移
- 数据清理（一次性）：
  - 删除“空 UPDATE 审计”：`event_type='UPDATE' AND before_data=after_data AND jsonb_array_length(coalesce(changes,'[]'))=0`。
- 迁移步骤：
  - `021_audit_and_temporal_sane_updates.sql`：仅值变更才 UPDATE；UPDATE 无变化跳过审计；上下文字段规范化。
  - `022_remove_db_triggers_and_functions.sql`：移除审计/时态/标志触发器与相关函数；应用层完全接管写路径。
- 目标：新写入在 022 生效 + 应用接管后，应收敛至“1 CREATE + ≤1 有效 UPDATE（存在未来版本时）”。

—

## 7. 工具与门禁（CI 强制）
- 脚本：
  - 报告版 SQL：`scripts/validate-audit-recordid-consistency.sql`
  - 断言版 SQL：`scripts/validate-audit-recordid-consistency-assert.sql`
  - 一键脚本：`scripts/apply-audit-fixes.sh`（`ENFORCE=1` 启用断言；`APPLY_FIXES=0` 仅校验，`APPLY_FIXES=1` 修复+校验）
- 工作流：
  - `.github/workflows/audit-consistency.yml`：Postgres16 → 应用 021/022 → `ENFORCE=1 APPLY_FIXES=0` 强制校验。
  - `.github/workflows/consistency-guard.yml`（job: audit）：同样流程并列执行。
- 断言口径：
  - 空 UPDATE=0；recordId 与载荷一致；目标触发器不存在：`audit_changes_trigger`、`auto_end_date_trigger`、`auto_lifecycle_status_trigger`、`enforce_soft_delete_temporal_flags_trigger`。

—

## 8. 测试与验收（DoD）
- GraphQL 契约/集成：
  - `organizationVersions(code)` 获取 recordId 列表；对每个 recordId 执行 `auditHistory(recordId)`。
  - 断言：recordId mismatch=0；空 UPDATE（严格口径）=0；新写入条数符合“1 CREATE + ≤1 UPDATE（有未来版本时）”。
- SQL 巡检：
  - 无重叠区间；每 code 当前唯一；删除态不为当前。
  - 审计一致性：空 UPDATE=0；recordId 与载荷一致；目标触发器=0。
- CI 门禁：PR 必须通过“Audit Consistency Gate / Consistency Guard（audit）”。

—

## 9. 运维与监控
- 例行巡检：发布前后与每日定时执行报告版 SQL；ENFORCE=1 模式用于 PR/主干。
- 告警：若发现空 UPDATE>0、recordId 错配>0、目标触发器>0 任一，则标红并阻断合并。
- 观测：新写入的审计条数分布、时态区间重叠率、当前唯一性违规。

—

## 10. 风险与回滚
- 风险：在应用接管完成前移除触发器，可能导致时间轴与审计缺失或不一致。
- 回滚：蓝绿或开关控制；在 `022` 执行前确认应用接管与测试通过；迁移/清理前执行全量备份。

—

## 11. 关键文件与命令
- 迁移：
  - `database/migrations/021_audit_and_temporal_sane_updates.sql`
  - `database/migrations/022_remove_db_triggers_and_functions.sql`
- 脚本：
  - `scripts/validate-audit-recordid-consistency.sql`
  - `scripts/validate-audit-recordid-consistency-assert.sql`
  - `scripts/apply-audit-fixes.sh`
- 工作流：
  - `.github/workflows/audit-consistency.yml`
  - `.github/workflows/consistency-guard.yml`
- 本地仅校验（不改动数据）：
  - `ENFORCE=1 APPLY_FIXES=0 bash scripts/apply-audit-fixes.sh`
- 本地修复+校验（建议先执行 021→022）：
  - `ENFORCE=1 APPLY_FIXES=1 bash scripts/apply-audit-fixes.sh`

—

## 附：现场巡检要点（2025-09-15）
- 环境健康：8090/9090 healthy；Postgres/Redis healthy。
- 样本（tenantId: 3b99930c-...；code=1000002）：
  - 版本：`2025-04-01 (a42811c1-...)`、`2025-08-01 (1a0a5ad9-...)`、`2025-09-06 (2f8d7380-...)`。
  - 审计（recordId=2f8d...）：总 11 条（CREATE=2，UPDATE=9）；mismatch=0；空 UPDATE（严格口径）=0。
  - 审计（recordId=a428...）：总 10 条；mismatch=0；空 UPDATE（AND 条件）=0；但存在 changes 为空的 UPDATE 若干（多为被过滤的元字段差异）。
- 一致性（库级）：空 UPDATE=0；recordId 归属一致。
- 说明：历史条目偏多源于过往触发链；在 `022` 生效与应用接管后，新写入应按本报告的 DoD 收敛。

—

## 变更记录
- 2025-09-15：报告定稿。明确“应用层接管写路径、数据库去触发器化”的分工；沉淀 021/022 迁移、SQL 校验与 CI 门禁；给出 DoD 与运维/回滚方案。替换原 Pending Issues 文档。

