\set ON_ERROR_STOP on

BEGIN;
SET LOCAL search_path TO public;

-- 目标组织编码
CREATE TEMP TABLE plan26_target_codes(code text primary key);
INSERT INTO plan26_target_codes(code) VALUES
  ('1000000'), ('1000001'), ('1000002'), ('1000003'), ('1000004');

-- 待保留的 record_id
CREATE TEMP TABLE plan26_target_record_ids(record_id uuid primary key);
INSERT INTO plan26_target_record_ids(record_id)
SELECT record_id
FROM organization_units
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND code IN (SELECT code FROM plan26_target_codes);

-- 待删除的 record_id
CREATE TEMP TABLE plan26_remove_record_ids(record_id uuid primary key);
INSERT INTO plan26_remove_record_ids(record_id)
SELECT record_id
FROM organization_units
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND code NOT IN (SELECT code FROM plan26_target_codes);

\echo 'Skipping organization_timeline_events cleanup (table removed in plan27)'
\echo 'Skipping organization_unit_versions cleanup (table removed in plan27)'

\echo 'Deleting audit_logs...'
WITH removed AS (
  DELETE FROM audit_logs
  WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
    AND resource_id IN (
      SELECT record_id::text FROM plan26_remove_record_ids
    )
  RETURNING id
)
SELECT COUNT(*) AS audit_logs_deleted FROM removed;

\echo 'Deleting organization_units_backup_temporal...'
WITH removed AS (
  DELETE FROM organization_units_backup_temporal
  WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
    AND code NOT IN (SELECT code FROM plan26_target_codes)
  RETURNING record_id
)
SELECT COUNT(*) AS backup_temporal_deleted FROM removed;

\echo 'Deleting organization_units_unittype_backup...'
WITH removed AS (
  DELETE FROM organization_units_unittype_backup
  WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
    AND code NOT IN (SELECT code FROM plan26_target_codes)
  RETURNING record_id
)
SELECT COUNT(*) AS backup_unittype_deleted FROM removed;

\echo 'Deleting organization_units...'
WITH removed AS (
  DELETE FROM organization_units ou
  WHERE ou.tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
    AND ou.code NOT IN (SELECT code FROM plan26_target_codes)
  RETURNING code, record_id
)
SELECT COUNT(*) AS organization_units_deleted FROM removed;

COMMIT;
