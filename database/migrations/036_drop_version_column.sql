-- 036_drop_version_column.sql
-- 目标：彻底移除 organization_units 及备份表中的 version 字段，并清理解耦触发器

BEGIN;

-- 先移除依赖 version 的视图
DROP VIEW IF EXISTS organization_temporal_current;

-- 移除触发器与函数（version 专用）
DROP TRIGGER IF EXISTS organization_version_trigger ON organization_units;
DROP FUNCTION IF EXISTS organization_version_trigger();

-- 移除依赖 version 的索引
DROP INDEX IF EXISTS idx_organization_units_version;

-- 删除主表 version 字段
ALTER TABLE organization_units
  DROP COLUMN IF EXISTS version;

-- 备份表同样移除 version 字段，保持结构一致
ALTER TABLE organization_units_backup_temporal
  DROP COLUMN IF EXISTS version;

ALTER TABLE organization_units_unittype_backup
  DROP COLUMN IF EXISTS version;

-- 重建视图（不包含 version 字段）
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
  code_path AS path,
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

COMMIT;
