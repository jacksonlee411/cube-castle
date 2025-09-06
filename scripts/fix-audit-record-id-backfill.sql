-- fix-audit-record-id-backfill.sql
-- 一次性数据修复脚本：为历史审计日志按组织版本时间范围补齐 record_id
-- 适用场景：audit_logs.record_id 存在 NULL，导致 auditHistory(recordId) 错配
-- 原理：使用 entity_code 与 operation_timestamp 在 organization_units 的生效区间内匹配到唯一版本的 record_id

BEGIN;

-- 可选：先看看还有多少空记录
-- SELECT count(*) AS null_record_id_count FROM audit_logs WHERE record_id IS NULL;

WITH ranked_matches AS (
  SELECT
    a.audit_id,
    u.record_id AS matched_record_id,
    u.is_current,
    u.effective_date,
    ROW_NUMBER() OVER (
      PARTITION BY a.audit_id
      ORDER BY 
        u.is_current DESC, -- 优先当前版本
        ABS(EXTRACT(EPOCH FROM (a.operation_timestamp - u.effective_date))) ASC -- 次优先开始时间更接近
    ) AS rn
  FROM audit_logs a
  JOIN organization_units u
    ON u.code = a.entity_code
   AND a.operation_timestamp >= u.effective_date
   AND (u.end_date IS NULL OR a.operation_timestamp < u.end_date)
  WHERE a.record_id IS NULL
)
UPDATE audit_logs a
SET record_id = r.matched_record_id
FROM ranked_matches r
WHERE a.audit_id = r.audit_id
  AND r.rn = 1;

-- 可选：核对剩余未匹配情况（例如越界数据或历史脏数据）
-- SELECT count(*) AS remain_null_after_fix FROM audit_logs WHERE record_id IS NULL;

COMMIT;

-- 说明：
-- 1) 该脚本仅填充 record_id 为 NULL 的行，对已存在值不做修改。
-- 2) 采用严格的时间窗口匹配：[effective_date, end_date)；end_date 为空视为无上界。
-- 3) 多候选时采用当前版本优先、其次按生效时间距操作时间最近的版本。
-- 4) 如仍有未匹配的 NULL，请逐条核查 entity_code 是否存在、操作时间是否落于任何版本区间。

