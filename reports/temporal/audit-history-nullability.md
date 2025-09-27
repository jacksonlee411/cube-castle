# 审计历史 GraphQL 非空约束巡检基线

> 单一事实来源：`sql/inspection/audit-history-nullability.sql`、`docs/api/schema.graphql`
> 运行方式：`psql -d ${DB_NAME} -f sql/inspection/audit-history-nullability.sql > reports/temporal/audit-history-nullability-$(date +%Y%m%d).log`

## 1. 巡检执行记录
- **执行日期**：
- **执行人**：
- **数据快照时间**：
- **环境**：

## 2. 统计摘要
| 指标 | 数值 | 说明 |
| --- | --- | --- |
| modified_fields NULL/非数组 条数 | | `查询1` 结果总计 |
| changes NULL/非数组 条数 | | `查询2` 结果总计 |
| 缺失 dataType 的条目 | | `查询3` 结果总计 |
| 受影响租户数 | | 

## 3. 受影响租户与事件分布
> 将查询结果按租户、事件类型粘贴或引用，确保引用单一事实来源。

```
<粘贴/引用查询结果>
```

## 4. 样本分析
> 从 `查询4` 获取的抽样数据中挑选代表性样本，并追溯上游来源（命令事件、迁移脚本或手动修复）。

- 样本 ID：
  - 业务背景：
  - 上游来源：
  - 对策：

## 5. 根因汇总
| 根因类型 | 描述 | 对应源文件/脚本 | 责任团队 |
| --- | --- | --- | --- |
| 例：命令服务未填充 operation_reason | | `cmd/organization-command-service/internal/audit/logger.go` | 命令服务团队 |

## 6. 性能基线
> 使用 `tests/perf/graphql-audit-history-benchmark.sh` 获取基线数据。

| 场景 | 记录数 | P50 (ms) | P95 (ms) | 总耗时 (s) | 备注 |
| --- | --- | --- | --- | --- | --- |
| 单租户-默认参数 (baseline) | | | | | |
| 单租户-记录拉满 (stress) | | | | | |

## 7. 待办与跟踪
- [ ] 更新实施计划状态
- [ ] 通知命令服务/数据平台团队跟进根因
- [ ] 将结果同步至 `docs/development-plans/15-audit-history-graphql-nullability-plan.md`

---

> 注：本报告为 Phase 0 输出模板，执行完成后请确保与 `CLAUDE.md`、`AGENTS.md` 约束一致，并在计划归档时同步更新。
