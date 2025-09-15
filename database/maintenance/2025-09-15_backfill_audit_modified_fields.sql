-- 2025-09-15_backfill_audit_modified_fields.sql
-- 目的：为历史审计记录补充 modified_fields 与 changes，以便前端“审计信息”页签显示变动信息。
-- 说明：仅在字段缺失(NULL)时补写；不覆盖已有值。仅针对 ORGANIZATION 资源。

BEGIN;

-- 1) CREATE（含 CREATE_VERSION/组织创建）
UPDATE audit_logs
SET 
  modified_fields = ' ["name","unitType","parentCode","description","effectiveDate"] '::jsonb,
  changes = (
    SELECT jsonb_agg(obj)
    FROM (
      SELECT jsonb_build_object('field','name','oldValue',NULL,'newValue',after_data->>'name','dataType','string') UNION ALL
      SELECT jsonb_build_object('field','unitType','oldValue',NULL,'newValue',after_data->>'unitType','dataType','string') UNION ALL
      SELECT jsonb_build_object('field','parentCode','oldValue',NULL,'newValue',after_data->>'parentCode','dataType','string') UNION ALL
      SELECT jsonb_build_object('field','description','oldValue',NULL,'newValue',after_data->>'description','dataType','string') UNION ALL
      SELECT jsonb_build_object('field','effectiveDate','oldValue',NULL,'newValue',after_data->>'effectiveDate','dataType','date')
    ) AS t(obj)
  )
WHERE resource_type='ORGANIZATION' AND event_type='CREATE'
  AND (modified_fields IS NULL OR changes IS NULL);

-- 2) SUSPEND（停用）
UPDATE audit_logs
SET 
  modified_fields = '["status"]'::jsonb,
  changes = jsonb_build_array(
    jsonb_build_object('field','status','oldValue', COALESCE(before_data->>'status', NULL), 'newValue','INACTIVE','dataType','string')
  )
WHERE resource_type='ORGANIZATION' AND event_type='SUSPEND'
  AND (modified_fields IS NULL OR changes IS NULL);

-- 3) ACTIVATE（重新启用）
UPDATE audit_logs
SET 
  modified_fields = '["status"]'::jsonb,
  changes = jsonb_build_array(
    jsonb_build_object('field','status','oldValue', COALESCE(before_data->>'status', NULL), 'newValue','ACTIVE','dataType','string')
  )
WHERE resource_type='ORGANIZATION' AND event_type='ACTIVATE'
  AND (modified_fields IS NULL OR changes IS NULL);

-- 4) DELETE（作废/删除）
UPDATE audit_logs
SET 
  modified_fields = '["status"]'::jsonb,
  changes = jsonb_build_array(
    jsonb_build_object('field','status','oldValue', before_data->>'status', 'newValue','DELETED','dataType','string')
  )
WHERE resource_type='ORGANIZATION' AND event_type='DELETE'
  AND (modified_fields IS NULL OR changes IS NULL);

COMMIT;

