-- scripts/data-consistency-check.sql
-- 目的：输出关键一致性问题清单，用于运维自检与监控采集

-- 1) 同一 code 存在多个 is_current=true 的冲突
SELECT 'MULTIPLE_CURRENT' AS issue,
       code,
       COUNT(*) AS current_count
  FROM organization_units
 WHERE is_current = TRUE
 GROUP BY code
HAVING COUNT(*) > 1;

-- 2) 时态区间重叠（同一 code 的版本时间区间相交）
SELECT 'TEMPORAL_OVERLAP' AS issue,
       u1.code,
       u1.effective_date AS eff1,
       u1.end_date       AS end1,
       u2.effective_date AS eff2,
       u2.end_date       AS end2
  FROM organization_units u1
  JOIN organization_units u2
    ON u1.code = u2.code
   AND u1.record_id <> u2.record_id
 WHERE daterange(u1.effective_date, COALESCE(u1.end_date, 'infinity'::date), '[]') &&
       daterange(u2.effective_date, COALESCE(u2.end_date, 'infinity'::date), '[]')
   AND NOT (
     u1.end_date = u2.effective_date OR u2.end_date = u1.effective_date
   )
 LIMIT 100;

-- 3) 子节点指向无效父节点（父节点不存在/非当前/已删除）
SELECT 'INVALID_PARENT' AS issue,
       c.code           AS child_code,
       c.parent_code    AS parent_code
  FROM organization_units c
  LEFT JOIN organization_units p
    ON p.code = c.parent_code
   AND p.is_current = TRUE
 WHERE c.parent_code IS NOT NULL
   AND (p.code IS NULL)
  LIMIT 100;

-- 4) 软删除记录不应为当前
SELECT 'DELETED_BUT_CURRENT' AS issue,
       code
  FROM organization_units
 WHERE status = 'DELETED'
   AND is_current = TRUE
 LIMIT 100;

-- 5) 审计日志最小列校验（存在性与近期记录）
SELECT 'AUDIT_RECENT' AS info,
       COUNT(*)       AS records_last_7d
  FROM audit_logs
 WHERE timestamp >= NOW() - INTERVAL '7 days';
