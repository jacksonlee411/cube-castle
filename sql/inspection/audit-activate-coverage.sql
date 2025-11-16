-- 审计表 ACTIVATE 覆盖范围巡检脚本（只读）
-- 目标：评估 POSITION / JOB_CATALOG 是否存在历史 ACTIVATE 事件，以决定迁移范围
-- 运行方式（容器内 psql）：
--   docker compose -f docker-compose.dev.yml exec -T postgres \
--     psql -U user -d cubecastle -v ON_ERROR_STOP=1 -f - < sql/inspection/audit-activate-coverage.sql
\pset pager off

\echo '\n[1] ACTIVATE vs REACTIVATE - 按资源类型汇总'
SELECT resource_type, event_type, count(*) AS cnt
FROM audit_logs
WHERE event_type IN ('ACTIVATE','REACTIVATE')
GROUP BY 1,2
ORDER BY 1,2;

\echo '\n[2] POSITION/JOB_CATALOG ACTIVATE 总量'
SELECT resource_type, count(*) AS cnt
FROM audit_logs
WHERE event_type = 'ACTIVATE'
  AND resource_type IN ('POSITION','JOB_CATALOG')
GROUP BY 1
ORDER BY 1;

\echo '\n[3] POSITION/JOB_CATALOG ACTIVATE - 近90天按月'
SELECT resource_type, date_trunc('month', timestamp) AS month, count(*) AS cnt
FROM audit_logs
WHERE event_type = 'ACTIVATE'
  AND resource_type IN ('POSITION','JOB_CATALOG')
  AND timestamp >= now() - interval '90 days'
GROUP BY 1,2
ORDER BY 1,2;

\echo '\n[4] POSITION/JOB_CATALOG ACTIVATE 数据质量：changes / before/after / record_id'
SELECT resource_type,
       sum(CASE WHEN changes IS NULL OR jsonb_typeof(changes) <> 'array' OR jsonb_array_length(changes)=0 THEN 1 ELSE 0 END) AS empty_changes,
       sum(CASE WHEN request_data IS NULL OR request_data='{}'::jsonb THEN 1 ELSE 0 END) AS empty_before,
       sum(CASE WHEN response_data IS NULL OR response_data='{}'::jsonb THEN 1 ELSE 0 END) AS empty_after,
       sum(CASE WHEN record_id IS NULL THEN 1 ELSE 0 END) AS null_record_id,
       count(*) AS total
FROM audit_logs
WHERE event_type='ACTIVATE'
  AND resource_type IN ('POSITION','JOB_CATALOG')
GROUP BY resource_type
ORDER BY resource_type;

\echo '\n[5] 示例：POSITION 最近 20 条 ACTIVATE（便于人工抽样核验）'
SELECT id,
       timestamp,
       resource_id,
       record_id::text,
       COALESCE(changes->0->>'field','')      AS sample_field,
       COALESCE(changes->0->>'newValue','')   AS sample_new
FROM audit_logs
WHERE resource_type='POSITION'
  AND event_type='ACTIVATE'
ORDER BY timestamp DESC
LIMIT 20;

\echo '\n[6] 示例：JOB_CATALOG 最近 20 条 ACTIVATE（便于人工抽样核验）'
SELECT id,
       timestamp,
       resource_id,
       record_id::text,
       COALESCE(changes->0->>'field','')      AS sample_field,
       COALESCE(changes->0->>'newValue','')   AS sample_new
FROM audit_logs
WHERE resource_type='JOB_CATALOG'
  AND event_type='ACTIVATE'
ORDER BY timestamp DESC
LIMIT 20;

