-- 033_cleanup_audit_empty_changes.sql
-- 目的：清理 Audit Logs 中由旧触发器生成、changes/modified_fields 为空的冗余 UPDATE 记录
-- 影响：仅删除 resource_type='ORGANIZATION' 且无字段差异的历史日志，不影响真实审计信息

BEGIN;

DELETE FROM audit_logs
WHERE resource_type = 'ORGANIZATION'
  AND event_type = 'UPDATE'
  AND COALESCE(jsonb_array_length(changes), 0) = 0
  AND COALESCE(jsonb_array_length(modified_fields), 0) = 0
  AND COALESCE(request_data, '{}'::jsonb) = '{}'::jsonb
  AND COALESCE(response_data, '{}'::jsonb) = '{}'::jsonb;

COMMIT;
