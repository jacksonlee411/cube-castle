-- Fix temporal timeline continuity by recomputing end_date
-- for each (tenant_id, code) ordered by effective_date.
-- Non-DELETED rows only participate in the timeline.
-- Also realign is_temporal to match end_date presence.

BEGIN;

WITH ordered AS (
  SELECT
    record_id,
    tenant_id,
    code,
    effective_date,
    LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
  FROM organization_units
  WHERE status != 'DELETED' AND deleted_at IS NULL
),
updates AS (
  SELECT
    record_id,
    CASE
      WHEN next_effective IS NULL THEN NULL
      ELSE (next_effective - INTERVAL '1 day')::date
    END AS new_end
  FROM ordered
)
UPDATE organization_units u
SET end_date = up.new_end,
    updated_at = NOW()
FROM updates up
WHERE u.record_id = up.record_id
  AND u.status != 'DELETED' AND u.deleted_at IS NULL
  AND (u.end_date IS DISTINCT FROM up.new_end);

COMMIT;
