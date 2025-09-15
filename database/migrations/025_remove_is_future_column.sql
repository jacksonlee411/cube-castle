-- 025_remove_is_future_column.sql
-- Purpose: Remove is_future column and related triggers; derive isFuture at read time.

BEGIN;

-- Drop dependent views and recreate without is_future
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_views WHERE viewname = 'organization_temporal_current') THEN
    EXECUTE 'DROP VIEW organization_temporal_current';
  END IF;
END$$;

-- Drop triggers that enforce is_future
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'enforce_temporal_flags_trigger') THEN
    EXECUTE 'DROP TRIGGER enforce_temporal_flags_trigger ON organization_units';
  END IF;
END$$;

-- Drop function if exists
DROP FUNCTION IF EXISTS enforce_temporal_flags() CASCADE;

-- Drop any indexes referencing is_future
DO $$
DECLARE r RECORD;
BEGIN
  FOR r IN
    SELECT indexname
    FROM pg_indexes
    WHERE tablename = 'organization_units'
      AND indexdef ILIKE '%is_future%'
  LOOP
    EXECUTE format('DROP INDEX IF EXISTS %I', r.indexname);
  END LOOP;
END$$;

-- Drop column if exists
ALTER TABLE organization_units
  DROP COLUMN IF EXISTS is_future;

-- Recreate simplified view without is_future
CREATE OR REPLACE VIEW organization_temporal_current AS
  SELECT * FROM organization_units WHERE is_current = true;

COMMIT;
