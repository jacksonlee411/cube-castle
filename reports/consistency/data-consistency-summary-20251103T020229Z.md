# 数据一致性巡检报告 (20251103T020229Z)

| 检查项 | 异常数量 |
|--------|----------|
| 多个 is_current 版本冲突 | 0 |
| 时态区间重叠 | 0 |
| 无效父节点 | 0 |
| 软删除仍标记为当前 | 0 |
| 最近 7 天审计日志数 | 0 |
| 结论 | ✅ PASS |

- SQL 来源：`scripts/data-consistency-check.sql`
- 原始输出：`data-consistency-20251103T020229Z.csv`
- 生成时间（UTC）：20251103T020229Z

如发现异常，请参考 `docs/architecture/temporal-consistency-implementation-report.md` 制定修复计划，并在修复后重新执行本脚本。
