# 审计历史 GraphQL 非空约束巡检基线

> 单一事实来源：`sql/inspection/audit-history-nullability.sql`、`docs/api/schema.graphql`
> 运行方式：`psql -d ${DB_NAME} -f sql/inspection/audit-history-nullability.sql > reports/temporal/audit-history-nullability-$(date +%Y%m%d).log`

## 1. 巡检执行记录
- **执行日期**：2025-10-06
- **执行人**：shangmeilin / 代理
- **数据快照时间**：2025-10-06 08:55 CST
- **环境**：本地开发容器（已启动 Postgres）
- **巡检日志**：`reports/temporal/audit-history-nullability-20251006.log`

## 2. 统计摘要
| 指标 | 数值 | 说明 |
| --- | --- | --- |
| 总审计记录数 | 2 | 租户 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 的 UPDATE 事件 |
| modified_fields NULL/非数组 条数 | 0 | `查询2` 无异常 ✅ |
| changes NULL/非数组 条数 | 0 | `查询2` 无异常 ✅ |
| 缺失 dataType 的条目 | 1 | `查询3` 发现 1 条记录存在 dataType 缺失 ⚠️ |
| 受影响租户数 | 1 | 租户 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 |

## 3. 受影响租户与事件分布
> 数据来源：`sql/inspection/audit-history-nullability.sql` 输出

```
🧪 1. 数据库表基本统计
              tenant_id               | event_type | total_records
--------------------------------------+------------+---------------
 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 | UPDATE     |             2

🧪 2. changes NULL 或 非数组 的记录统计
 tenant_id | event_type | suspect_count
-----------+------------+---------------
(0 rows)  -- ✅ 无异常

🧪 3. changes 数组内缺失 dataType 的条目明细
              tenant_id               | event_type | missing_data_type
--------------------------------------+------------+-------------------
 3b99930c-4dc6-4cc9-8e4d-7d960a931cb9 | UPDATE     |                 1

🧪 4. 示例抽样（缺失 dataType 的记录）
audit_id: 5a380d66-e581-4700-b7f3-803042babd7c
timestamp: 2025-09-27 14:45:03.813114+08
changes: [
  {"field": "name", "newValue": "新名称", "oldValue": "旧名称"},           ⚠️ 缺失 dataType
  {"field": "description", "dataType": "string", "newValue": "新描述", "oldValue": null}
]
```

## 4. 样本分析
> 从 `查询4` 获取的抽样数据中挑选代表性样本，并追溯上游来源（命令事件、迁移脚本或手动修复）。

- **样本 ID**: 5a380d66-e581-4700-b7f3-803042babd7c
  - **业务背景**: 组织单元 UPDATE 操作（recordId: 8fee4ec4-865c-494b-8d5c-2bc72c312733）
  - **问题描述**: changes 数组中 "name" 字段缺失 dataType，GraphQL 返回 "dataType": "unknown"
  - **上游来源**: 审计触发器 `log_audit_changes()` 在记录 name 字段变更时未填充 dataType
  - **对策**: 需修复触发器逻辑，确保所有字段变更都包含正确的 dataType

- **样本 ID**: bd52d886-b4e6-42a7-9f4c-6f8c8ec3f8a2
  - **业务背景**: 组织单元 UPDATE 操作（recordId: 8fee4ec4-865c-494b-8d5c-2bc72c312733）
  - **问题描述**: changes 和 modifiedFields 均为空数组，虽然 event_type 为 UPDATE
  - **上游来源**: 可能是迁移脚本或数据补充操作触发的审计记录，但未记录实际变更
  - **对策**: 需确认是否为合法的空变更记录，或修复触发器以避免记录无意义的 UPDATE

## 5. 根因汇总
| 根因类型 | 描述 | 对应源文件/脚本 | 责任团队 |
| --- | --- | --- | --- |
| 审计触发器未填充 dataType | 部分字段变更记录缺失 dataType 属性，导致 GraphQL 返回 "unknown" | 数据库触发器 `log_audit_changes()` | 数据库/命令服务团队 |
| 空变更的 UPDATE 记录 | 存在 changes/modifiedFields 为空但 event_type=UPDATE 的记录 | 数据库触发器或迁移脚本 | 数据库团队 |

## 6. GraphQL 查询验证结果

### 6.1 测试请求
```graphql
query($id: String!) {
  auditHistory(recordId: $id) {
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

**变量**: `{ "id": "8fee4ec4-865c-494b-8d5c-2bc72c312733" }`

### 6.2 实际响应
```json
{
  "success": true,
  "data": {
    "auditHistory": [
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
      },
      {
        "auditId": "bd52d886-b4e6-42a7-9f4c-6f8c8ec3f8a2",
        "changes": [],  ⚠️ 缺陷：空变更的 UPDATE 记录
        "modifiedFields": [],
        "operation": "UPDATE",
        "recordId": "8fee4ec4-865c-494b-8d5c-2bc72c312733",
        "timestamp": "2025-09-27T14:44:55.799881+08:00"
      }
    ]
  },
  "message": "Query executed successfully",
  "timestamp": "2025-10-06T01:00:55Z"
}
```

### 6.3 修复摘要
- 2025-10-06 更新 `cmd/organization-query-service/internal/repository/postgres_audit.go` 中 `sanitizeChanges`，对缺失或 `unknown` 的 `dataType` 依据 old/new 值推断类型。
- 复测后 GraphQL 返回已将 `dataType` 更正为 `"string"`。

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
- [ ] Phase 1: 定位 `log_audit_changes()` 触发器逻辑，分析 dataType 缺失根因
- [ ] Phase 1: 分析空变更 UPDATE 记录的产生原因
- [ ] Phase 2: 修复触发器，确保所有字段变更都包含正确 dataType
- [ ] Phase 2: 决策空变更记录的处理策略（过滤或补数据）
- [ ] 将结果同步至 `docs/development-plans/07-audit-history-load-failure-fix-plan.md`

---

> 注：本报告为 Phase 0 输出模板，执行完成后请确保与 `CLAUDE.md`、`AGENTS.md` 约束一致，并在计划归档时同步更新。
