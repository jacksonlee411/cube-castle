-- cleanup_legacy_triggers.sql
-- 目的：清理不属于当前迁移链且可能误执行的遗留触发器/函数。
-- 注意：仅删除 IF EXISTS 的对象，且仅针对下列白名单名称；请在执行前确认目标库不依赖这些对象。

BEGIN;

-- 组织表（organization_units）上不再使用的触发器
DROP TRIGGER IF EXISTS auto_end_date_trigger ON organization_units;
DROP TRIGGER IF EXISTS organization_units_change_trigger ON organization_units;
DROP TRIGGER IF EXISTS set_org_unit_code ON organization_units;
DROP TRIGGER IF EXISTS simple_temporal_gap_fill_trigger ON organization_units;
DROP TRIGGER IF EXISTS smart_hierarchy_management ON organization_units;
DROP TRIGGER IF EXISTS validate_hierarchy ON organization_units;
DROP TRIGGER IF EXISTS auto_lifecycle_status_trigger ON organization_units;
DROP TRIGGER IF EXISTS set_operation_type_trigger ON organization_units;

-- 历史/过渡期表上的旧触发器（如存在）
DROP TRIGGER IF EXISTS trigger_auto_end_date_v2 ON organization_versions;

-- 可选：删除相关函数（若确认不再引用）
-- DROP FUNCTION IF EXISTS notify_organization_change() CASCADE;
-- DROP FUNCTION IF EXISTS smart_hierarchy_trigger() CASCADE;
-- DROP FUNCTION IF EXISTS auto_manage_end_dates() CASCADE;
-- DROP FUNCTION IF EXISTS simple_temporal_gap_fill_trigger() CASCADE;
-- DROP FUNCTION IF EXISTS set_org_unit_code_func() CASCADE;
-- DROP FUNCTION IF EXISTS validate_hierarchy_changes() CASCADE;

COMMIT;

