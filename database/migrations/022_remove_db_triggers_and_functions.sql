-- 022_remove_db_triggers_and_functions.sql
-- 目的：按决策立即移除数据库触发器逻辑，由应用层全面接管。
-- 移除范围：审计触发、时态自动 endDate、生命周期/软删除标志触发及相关函数。

BEGIN;

-- 1) 删除触发器（若存在）
DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
DROP TRIGGER IF EXISTS auto_end_date_trigger ON organization_units;
DROP TRIGGER IF EXISTS auto_lifecycle_status_trigger ON organization_units;
DROP TRIGGER IF EXISTS enforce_soft_delete_temporal_flags_trigger ON organization_units;

-- 2) 删除函数（若存在）
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;
DROP FUNCTION IF EXISTS auto_manage_end_dates() CASCADE;
DROP FUNCTION IF EXISTS auto_update_lifecycle_status() CASCADE;
DROP FUNCTION IF EXISTS enforce_soft_delete_temporal_flags() CASCADE;
DROP FUNCTION IF EXISTS calculate_field_changes_min(JSONB, JSONB) CASCADE;

COMMIT;

