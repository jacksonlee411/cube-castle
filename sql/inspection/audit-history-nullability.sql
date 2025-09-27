-- 审计历史 GraphQL 非空约束巡检脚本
-- 唯一事实来源：audit_logs 表
-- 使用方法：psql -d ${DB_NAME} -f sql/inspection/audit-history-nullability.sql > reports/temporal/audit-history-nullability-$(date +%Y%m%d).log

\echo '🧪 1. 数据库表基本统计'
SELECT tenant_id,
       event_type,
       COUNT(*) AS total_records
FROM audit_logs
WHERE resource_type = 'ORGANIZATION'
GROUP BY tenant_id, event_type
ORDER BY total_records DESC;

\echo '\n🧪 2. changes NULL 或 非数组 的记录统计'
SELECT tenant_id,
       event_type,
       COUNT(*) AS suspect_count
FROM audit_logs
WHERE resource_type = 'ORGANIZATION'
  AND (changes IS NULL OR jsonb_typeof(changes) <> 'array')
GROUP BY tenant_id, event_type
HAVING COUNT(*) > 0
ORDER BY suspect_count DESC;

\echo '\n🧪 3. changes 数组内缺失 dataType 的条目明细（按租户/事件聚合）'
SELECT tenant_id,
       event_type,
       COUNT(*) AS missing_data_type
FROM audit_logs
WHERE resource_type = 'ORGANIZATION'
  AND changes IS NOT NULL
  AND jsonb_typeof(changes) = 'array'
  AND EXISTS (
        SELECT 1
        FROM jsonb_array_elements(changes) elem
        WHERE NOT (elem ? 'dataType')
           OR elem->>'dataType' IS NULL
           OR elem->>'dataType' = ''
      )
GROUP BY tenant_id, event_type
HAVING COUNT(*) > 0
ORDER BY missing_data_type DESC;

\echo '\n🧪 4. 示例抽样（每租户3条）'
WITH suspect AS (
    SELECT tenant_id,
           event_type,
           id,
           changes,
           timestamp,
           ROW_NUMBER() OVER (PARTITION BY tenant_id ORDER BY timestamp DESC) AS rn
    FROM audit_logs
    WHERE resource_type = 'ORGANIZATION'
      AND (
        changes IS NULL OR jsonb_typeof(changes) <> 'array'
        OR EXISTS (
              SELECT 1 FROM jsonb_array_elements(changes) elem
              WHERE NOT (elem ? 'dataType')
                 OR elem->>'dataType' IS NULL
                 OR elem->>'dataType' = ''
            )
      )
)
SELECT tenant_id,
       event_type,
       id,
       timestamp,
       changes
FROM suspect
WHERE rn <= 3
ORDER BY tenant_id, timestamp DESC;

\echo '\n✅ 巡检完成：请将结果汇总至 reports/temporal/audit-history-nullability.md'
