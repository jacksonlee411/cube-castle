-- scripts/daily-cutover.sql
-- 目的：执行每日时态数据维护，并记录审计日志
-- 注意：本脚本为幂等与非破坏性设计，不会删除业务数据

BEGIN;

-- 1) 规范删除记录的标志（软删除不可为当前/未来）
-- 软删除不可为当前
UPDATE organization_units
   SET is_current = FALSE
 WHERE deleted_at IS NOT NULL OR status = 'DELETED';

-- 2) 将未来生效的版本标记为 is_future
-- 未来版本不设为当前（isFuture 读时派生）
UPDATE organization_units
   SET is_current = FALSE
 WHERE effective_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date
   AND is_current = TRUE;

-- 3) 为每个 code 选择当前应生效的版本（最新且有效期覆盖今天）并规范 is_current
WITH latest AS (
  SELECT code, MAX(effective_date) AS eff
    FROM organization_units
   WHERE is_deleted = FALSE
     AND effective_date <= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date
     AND (end_date IS NULL OR end_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date)
   GROUP BY code
)
UPDATE organization_units u
   SET is_current = (u.effective_date = l.eff AND (u.end_date IS NULL OR u.end_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date))
  FROM latest l
 WHERE u.code = l.code
   AND u.is_current IS DISTINCT FROM (u.effective_date = l.eff AND (u.end_date IS NULL OR u.end_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date));

-- 4) 记录一次系统审计日志（使用统一契约列）
INSERT INTO audit_logs (
  tenant_id,
  event_type,
  resource_type,
  actor_id,
  actor_type,
  action_name,
  request_id,
  business_context,
  operation_timestamp
) VALUES (
  '00000000-0000-0000-0000-000000000000'::uuid,
  'UPDATE',
  'SYSTEM',
  'DAILY_CUTOVER_SYSTEM',
  'SYSTEM',
  'TEMPORAL_MAINTENANCE',
  'daily-cutover-' || to_char(NOW(), 'YYYY-MM-DD-HH24-MI-SS'),
  jsonb_build_object(
    'task', 'daily_cutover',
    'status', 'completed',
    'runAt', NOW()
  ),
  NOW()
);

COMMIT;
