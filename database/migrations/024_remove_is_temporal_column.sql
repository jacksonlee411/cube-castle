-- 024_remove_is_temporal_column.sql
-- Purpose: Remove redundant is_temporal column; isTemporal is derived from end_date

BEGIN;

-- Drop dependent views and recreate them without is_temporal
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'organization_temporal_current') THEN
    EXECUTE 'DROP VIEW organization_temporal_current';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'temporal_performance_stats') THEN
    EXECUTE 'DROP VIEW temporal_performance_stats';
  END IF;
END$$;

-- Drop indexes on organization_units that reference is_temporal
DO $$
DECLARE r RECORD;
BEGIN
  FOR r IN
    SELECT indexname
    FROM pg_indexes
    WHERE tablename = 'organization_units'
      AND indexdef ILIKE '%is_temporal%'
  LOOP
    EXECUTE format('DROP INDEX IF EXISTS %I', r.indexname);
  END LOOP;
END$$;

-- Drop column if exists
ALTER TABLE organization_units
  DROP COLUMN IF EXISTS is_temporal;

-- Recreate simplified view without is_temporal (select current records)
CREATE OR REPLACE VIEW organization_temporal_current AS
  SELECT * FROM organization_units WHERE is_current = true;

CREATE OR REPLACE VIEW temporal_performance_stats AS
SELECT 
    'current_temporal_orgs' AS metric,
    COUNT(*) AS value,
    NOW() AS collected_at
FROM organization_units 
WHERE is_current = TRUE
  AND (end_date IS NULL OR end_date > NOW())

UNION ALL

SELECT 
    'inactive_orgs' AS metric,
    COUNT(*) AS value,
    NOW() AS collected_at
FROM organization_units 
WHERE is_current = TRUE
  AND status <> 'ACTIVE';

COMMIT;
