-- Temporal consistency checks for organization_units
-- Run as a read-only inspection. No changes are made.

-- 1) Current must not have an end_date
SELECT record_id, tenant_id, code, effective_date, end_date
FROM organization_units
WHERE is_current = true AND end_date IS NOT NULL
ORDER BY tenant_id, code, effective_date;

-- 2) removed: is_temporal column has been dropped (isTemporal is derived from end_date)

-- 3) Overlap check within each (tenant_id, code)
-- Expect: prev.end_date < curr.effective_date (no overlap; allow direct adjacency)
WITH ordered AS (
  SELECT record_id, tenant_id, code, effective_date, end_date,
         LAG(end_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS prev_end
  FROM organization_units
  WHERE status != 'DELETED'
)
SELECT *
FROM ordered
WHERE prev_end IS NOT NULL AND prev_end >= effective_date
ORDER BY tenant_id, code, effective_date;

-- 4) Gap check within each (tenant_id, code)
-- Expect: prev.end_date = curr.effective_date - INTERVAL '1 day' (no gaps)
WITH ordered AS (
  SELECT record_id, tenant_id, code, effective_date,
         LAG(end_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS prev_end
  FROM organization_units
  WHERE status != 'DELETED'
)
SELECT *,
       (effective_date - INTERVAL '1 day') AS expected_prev_end
FROM ordered
WHERE prev_end IS NOT NULL AND prev_end < (effective_date - INTERVAL '1 day')
ORDER BY tenant_id, code, effective_date;
