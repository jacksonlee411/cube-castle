# 审计历史 GraphQL 非空约束巡检基线

> 单一事实来源：`sql/inspection/audit-history-nullability.sql`、`docs/api/schema.graphql`
> 运行方式：`psql -d ${DB_NAME} -f sql/inspection/audit-history-nullability.sql > reports/temporal/audit-history-nullability-$(date +%Y%m%d).log`

## 1. 巡检执行记录
- **执行日期**：2025-10-06（基线） / 2025-10-07（迁移 034 + 数据回填后复检）
- **执行人**：shangmeilin / 代理
- **数据快照时间**：2025-10-06 08:55 CST / 2025-10-07 16:25 CST
- **环境**：本地开发容器（Postgres + 最新迁移）
- **巡检日志**：`reports/temporal/audit-history-nullability-20251006.log`、`reports/temporal/audit-history-nullability-20251007.log`

## 2. 统计摘要
| 指标 | 数值 | 说明 |
| --- | --- | --- |
| 总审计记录数 | 3 | 租户 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 触发 3 条 UPDATE 审计（含 2 条复测新记录） |
| modified_fields NULL/非数组 条数 | 0 | `查询2` 无异常 ✅ |
| changes NULL/非数组 条数 | 0 | `查询2` 无异常 ✅ |
| 缺失 dataType 的条目 | 0 | `查询3` 已清零，历史记录通过补数据修复 ✅ |
| 受影响租户数 | 1 | 租户 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 |

## 3. 受影响租户与事件分布
> 数据来源：`sql/inspection/audit-history-nullability.sql` 输出

```
🧪 1. 数据库表基本统计
              tenant_id               | event_type | total_records
--------------------------------------+------------+---------------
 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 | UPDATE     |             3

🧪 2. changes NULL 或 非数组 的记录统计
 tenant_id | event_type | suspect_count
-----------+------------+---------------
(0 rows)

🧪 3. changes 数组内缺失 dataType 的条目明细
 tenant_id | event_type | missing_data_type
-----------+------------+-------------------
(0 rows)

🧪 4. 示例抽样（2025-10-07 复检）
-- 无异常记录输出
```

## 4. 样本分析
> 从 `查询4` 获取的抽样数据中挑选代表性样本，并追溯上游来源（命令事件、迁移脚本或手动修复）。

- **样本 ID**: 46d5a344-5b96-4a02-93c2-bb125642bacf（2025-10-07 新审计记录）
  - **业务背景**: 手动更新组织名称以验证新触发器（recordId: 8fee4ec4-865c-494b-8d5c-2bc72c312733）
  - **结果**: `changes` 数组包含 `dataType="string"`，并补录 `version` 字段差异，符合 GraphQL 契约。

- **样本 ID**: 57f77292-b2d1-44c1-be82-f835d3235edb（2025-10-07 恢复原名）
  - **结果**: 再次确认新触发器输出 `dataType` 与字段差异完整记录，`modified_fields` 自动去重。

- **遗留样本 ID**: 5a380d66-e581-4700-b7f3-803042babd7c
  - **处理**: 通过 `jsonb_set` 手工补齐 `dataType="string"`（2025-10-07），异常已消除。

## 5. 根因汇总
| 根因类型 | 描述 | 对应源文件/脚本 | 责任团队 | 状态 |
| --- | --- | --- | --- | --- |
| 审计触发器未填充 dataType | 旧版本触发器未输出 `dataType` 字段 | `database/migrations/031_cleanup_temporal_triggers.sql` 旧实现 | 数据库/命令服务团队 | ✅ 通过迁移 034 重建触发器并补数据 | 
| 空变更的 UPDATE 记录 | 历史日志存在空 `changes`/`modified_fields` | `database/migrations/033_cleanup_audit_empty_changes.sql` | 数据库团队 | ✅ 迁移 033 已清理，无新增样本 |

## 6. GraphQL 查询验证结果

### 6.1 测试请求
```graphql
query AuditHistoryCheck($recordId: String!, $limit: Int) {
  auditHistory(recordId: $recordId, limit: $limit) {
    auditId
    recordId
    operation
    timestamp
    modifiedFields
    changes {
      field
      oldValue
      newValue
      dataType
    }
  }
}
```

**变量**:
```json
{
  "recordId": "8fee4ec4-865c-494b-8d5c-2bc72c312733",
  "limit": 10
}
```

### 6.2 实际响应（2025-10-07 查询服务重启后复测）
```json
{
  "success": true,
  "data": {
    "auditHistory": [
      {
        "auditId": "57f77292-b2d1-44c1-be82-f835d3235edb",
        "changes": [
          {
            "dataType": "string",
            "field": "name",
            "newValue": "高谷集团",
            "oldValue": "高谷集团（审计修复验证）"
          },
          {
            "dataType": "number",
            "field": "version",
            "newValue": "3",
            "oldValue": "2"
          }
        ],
        "modifiedFields": ["name", "version"],
        "operation": "UPDATE",
        "recordId": "8fee4ec4-865c-494b-8d5c-2bc72c312733",
        "timestamp": "2025-10-07T16:21:48.744897+08:00"
      },
      {
        "auditId": "46d5a344-5b96-4a02-93c2-bb125642bacf",
        "changes": [
          {
            "dataType": "string",
            "field": "name",
            "newValue": "高谷集团（审计修复验证）",
            "oldValue": "高谷集团"
          },
          {
            "dataType": "number",
            "field": "version",
            "newValue": "2",
            "oldValue": "1"
          }
        ],
        "modifiedFields": ["name", "version"],
        "operation": "UPDATE",
        "recordId": "8fee4ec4-865c-494b-8d5c-2bc72c312733",
        "timestamp": "2025-10-07T16:21:42.315882+08:00"
      },
      {
        "auditId": "5a380d66-e581-4700-b7f3-803042babd7c",
        "changes": [
          {
            "dataType": "string",
            "field": "name",
            "newValue": "新名称",
            "oldValue": "旧名称"
          },
          {
            "dataType": "string",
            "field": "description",
            "newValue": "新描述",
            "oldValue": null
          }
        ],
        "modifiedFields": ["description", "name"],
        "operation": "UPDATE",
        "recordId": "8fee4ec4-865c-494b-8d5c-2bc72c312733",
        "timestamp": "2025-09-27T14:45:03.813114+08:00"
      }
    ]
  },
  "message": "Query executed successfully",
  "timestamp": "2025-10-07T08:30:28Z",
  "requestId": "unknown"
}
```

### 6.3 修复摘要
- 2025-10-06 更新 `sanitizeChanges` 推断缺失/`unknown` 的 `dataType`。
- 同步忽略空 `changes`/`modifiedFields` 的遗留 UPDATE 记录（历史 `log_audit_changes` 触发器输出）。
- 2025-10-07 重新执行 `database/migrations/034_rebuild_audit_trigger_with_diff.sql`，触发器现已输出完整字段差异与 `dataType`。
- 2025-10-07 对历史记录 `5a380d66-e581-4700-b7f3-803042babd7c` 补齐 `dataType`，并通过两次实际 UPDATE 操作验证新触发器输出。
- 2025-10-07 查询服务重启后完成 GraphQL 复测（见上方响应），`dataType` 字段符合契约。

### 6.3 发现的问题
1. **dataType 为 "unknown"**: name 字段的 dataType 应为 "string" 而非 "unknown"，与契约不符
2. **空变更记录**: 第二条记录的 changes 和 modifiedFields 都为空，不应作为有效的 UPDATE 事件记录
3. **GraphQL 契约符合性**: 查询能够成功返回数据（非空数组），但数据质量存在问题

## 7. 性能基线
> 使用 `tests/perf/graphql-audit-history-benchmark.sh` 获取基线数据（Phase 1 待执行）。

| 场景 | 记录数 | P50 (ms) | P95 (ms) | 总耗时 (s) | 备注 |
| --- | --- | --- | --- | --- | --- |
| 单租户-默认参数 (baseline) | | | | | |
| 单租户-记录拉满 (stress) | | | | | |

## 8. 待办与跟踪
- [x] Phase 0: 执行 SQL 巡检并记录结果（2025-10-06 完成）
- [x] Phase 0: 执行 GraphQL 查询验证并记录响应（2025-10-06 完成）
- [x] Phase 1: 定位 `log_audit_changes()` 触发器逻辑，分析 dataType 缺失根因（2025-10-07 完成）
- [x] Phase 1: 分析空变更 UPDATE 记录的产生原因（2025-10-07 完成）
- [x] Phase 2: 修复触发器，确保所有字段变更都包含正确 dataType（迁移 034 已执行并多次验证）
- [x] Phase 2: 决策空变更记录的处理策略（迁移 033 + 触发器更新后已无新样本）
- [x] Phase 2: 复跑 SQL 巡检，记录迁移 034 后的结果（2025-10-07 全绿）
- [x] Phase 2: GraphQL 查询复测（2025-10-07 查询服务重启后验证通过）
- [x] 将结果同步至 `docs/development-plans/07-audit-history-load-failure-fix-plan.md`（2025-10-07 更新）

---

> 注：本报告为 Phase 0 输出模板，执行完成后请确保与 `CLAUDE.md`、`AGENTS.md` 约束一致，并在计划归档时同步更新。
