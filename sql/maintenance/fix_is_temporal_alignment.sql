-- Fix is_temporal to align with end_date semantics
-- - is_temporal = true  when end_date IS NOT NULL (historical)
-- - is_temporal = false when end_date IS NULL     (open tail / current or future)
-- Safe to run multiple times; only updates rows that are out of sync.

BEGIN;

-- Optional: report how many rows are out-of-sync before fixing
-- SELECT COUNT(*) AS out_of_sync
-- FROM organization_units
-- WHERE is_temporal IS DISTINCT FROM (end_date IS NOT NULL);

UPDATE organization_units
SET is_temporal = CASE WHEN end_date IS NULL THEN false ELSE true END,
    updated_at  = NOW()
WHERE is_temporal IS DISTINCT FROM (end_date IS NOT NULL);

COMMIT;

-- Post-check (optional):
-- SELECT COUNT(*) AS remaining_out_of_sync
-- FROM organization_units
-- WHERE is_temporal IS DISTINCT FROM (end_date IS NOT NULL);

