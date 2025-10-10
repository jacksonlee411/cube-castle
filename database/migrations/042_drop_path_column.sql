-- 042_drop_path_column.sql
-- 移除 organization_units.path 列，统一以 code_path/name_path 作为唯一层级路径字段

BEGIN;

DROP VIEW IF EXISTS organization_temporal_current;
DROP VIEW IF EXISTS organization_current;

ALTER TABLE organization_units
    DROP COLUMN IF EXISTS path;

ALTER TABLE organization_units_backup_temporal
    DROP COLUMN IF EXISTS path;

ALTER TABLE organization_units_unittype_backup
    DROP COLUMN IF EXISTS path;

CREATE VIEW organization_temporal_current AS
SELECT
  record_id,
  tenant_id,
  code,
  parent_code,
  name,
  unit_type,
  status,
  level,
  hierarchy_depth,
  code_path,
  name_path,
  sort_order,
  description,
  profile,
  created_at,
  updated_at,
  effective_date,
  end_date,
  change_reason,
  is_current,
  deleted_at,
  deleted_by,
  deletion_reason,
  suspended_at,
  suspended_by,
  suspension_reason,
  operated_by_id,
  operated_by_name,
  metadata,
  effective_from,
  effective_to,
  changed_by,
  approved_by
FROM organization_units
WHERE is_current = true;

CREATE VIEW organization_current AS
SELECT
    ou.tenant_id,
    ou.code,
    ou.parent_code,
    ou.name,
    ou.unit_type,
    ou.status,
    ou.level,
    ou.code_path,
    ou.name_path,
    ou.sort_order,
    ou.description,
    ou.profile,
    ou.effective_date,
    ou.end_date,
    ou.is_current,
    ou.change_reason,
    ou.created_at,
    ou.updated_at
FROM organization_units ou
WHERE ou.is_current = true
  AND (ou.end_date IS NULL OR ou.end_date > CURRENT_DATE);

COMMIT;
