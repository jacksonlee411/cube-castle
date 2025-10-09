-- 040_normalize_audit_logs_json.sql
-- 目的：确保 audit_logs 中的数组/对象字段不会保存 JSON null，避免查询时触发 jsonb_array_length 错误。

BEGIN;

UPDATE audit_logs
SET changes = '[]'::jsonb
WHERE changes IS NULL OR jsonb_typeof(changes) = 'null';

UPDATE audit_logs
SET modified_fields = '[]'::jsonb
WHERE modified_fields IS NULL OR jsonb_typeof(modified_fields) = 'null';

UPDATE audit_logs
SET request_data = '{}'::jsonb
WHERE request_data IS NULL OR jsonb_typeof(request_data) = 'null';

UPDATE audit_logs
SET response_data = '{}'::jsonb
WHERE response_data IS NULL OR jsonb_typeof(response_data) = 'null';

COMMIT;

SELECT 'normalized audit_logs JSON columns' AS info, NOW() AS applied_at;
