-- scripts/validate-audit-recordid-consistency.sql
-- 目的：库内校验审计记录的一致性与触发器状态；供 apply-audit-fixes.sh 末尾调用。
-- 特点：
--  - 纯 SQL，无需应用层依赖；适用于迁移后离线校验
--  - 不主动抛错，仅输出汇总与疑点样本；配合 psql 的 ON_ERROR_STOP=1 保证语法层失败即中止
--  - 可用于 CI/运维巡检（建议按需增加“阈值断言”版本）

\echo '== AUDIT CONSISTENCY VALIDATION =='

-- 1) 汇总：空UPDATE计数、record_id错配计数、OU表触发器数量
WITH empty_updates AS (
  SELECT COUNT(*)::bigint AS cnt
  FROM audit_logs
  WHERE event_type = 'UPDATE'
    AND before_data = after_data
    AND jsonb_array_length(coalesce(changes, '[]'::jsonb)) = 0
),
mismatched_recordid AS (
  SELECT COUNT(*)::bigint AS cnt
  FROM audit_logs
  WHERE coalesce((after_data->>'record_id'), (before_data->>'record_id')) IS NOT NULL
    AND record_id IS DISTINCT FROM coalesce((after_data->>'record_id')::uuid, (before_data->>'record_id')::uuid)
),
triggers_present AS (
  SELECT COUNT(*)::bigint AS cnt
  FROM pg_trigger t
  JOIN pg_class c ON c.oid = t.tgrelid
  WHERE c.relname = 'organization_units'
    AND NOT t.tgisinternal
)
SELECT * FROM (
  SELECT 'EMPTY_UPDATES' AS check_item, cnt FROM empty_updates
  UNION ALL
  SELECT 'MISMATCHED_RECORD_ID', cnt FROM mismatched_recordid
  UNION ALL
  SELECT 'OU_TRIGGERS_PRESENT', cnt FROM triggers_present
) s
ORDER BY check_item;

\echo '== DETAILS: TOP 50 RECORD_ID PAYLOAD MISMATCHES =='
SELECT audit_id,
       record_id                               AS audit_record_id,
       (before_data->>'record_id')             AS before_record_id,
       (after_data->>'record_id')              AS after_record_id,
       event_type,
       timestamp
FROM audit_logs
WHERE coalesce((after_data->>'record_id'), (before_data->>'record_id')) IS NOT NULL
  AND record_id IS DISTINCT FROM coalesce((after_data->>'record_id')::uuid, (before_data->>'record_id')::uuid)
ORDER BY timestamp DESC
LIMIT 50;

\echo '== DETAILS: UPDATE WITH EMPTY CHANGES BUT BEFORE!=AFTER (TOP 50) =='
SELECT audit_id,
       event_type,
       timestamp
FROM audit_logs
WHERE event_type = 'UPDATE'
  AND jsonb_array_length(coalesce(changes, '[]'::jsonb)) = 0
  AND NOT (before_data = after_data)
ORDER BY timestamp DESC
LIMIT 50;

\echo '== TRIGGERS ON organization_units (SHOULD BE 0 AFTER 022) =='
SELECT tgname
FROM pg_trigger t
JOIN pg_class c ON c.oid = t.tgrelid
WHERE c.relname = 'organization_units'
  AND NOT t.tgisinternal
ORDER BY tgname;

\echo '== END OF AUDIT CONSISTENCY VALIDATION =='

