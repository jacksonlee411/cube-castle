-- validate-audit-recordid-consistency.sql
-- 审计日志与时态版本 record_id 一致性校验报告

-- 1) 统计 audit_logs.record_id 为空的记录
SELECT count(*) AS audit_logs_record_id_null
FROM audit_logs
WHERE record_id IS NULL;

-- 2) 统计审计行的 record_id 在 organization_units 中不存在的情况
SELECT count(*) AS audit_logs_record_id_orphan
FROM audit_logs a
LEFT JOIN organization_units u ON u.record_id = a.record_id
WHERE a.record_id IS NOT NULL AND u.record_id IS NULL;

-- 3) 校验 entity_code 与 record_id 指向的版本 code 不一致的异常
SELECT count(*) AS audit_entity_code_mismatch
FROM audit_logs a
JOIN organization_units u ON u.record_id = a.record_id
WHERE a.record_id IS NOT NULL AND a.entity_code <> u.code;

-- 4) 抽样列出异常详情（限制前100条）
SELECT a.audit_id, a.entity_code AS audit_code, u.code AS version_code,
       a.record_id, a.operation_timestamp
FROM audit_logs a
JOIN organization_units u ON u.record_id = a.record_id
WHERE a.record_id IS NOT NULL AND a.entity_code <> u.code
ORDER BY a.operation_timestamp DESC
LIMIT 100;

