-- 035_drop_temporal_history_tables.sql
-- 任务：移除已废弃的组织时态历史/事件表及相关对象

BEGIN;

-- 删除依赖已废弃表/视图的对象
DROP VIEW IF EXISTS temporal_performance_stats;
DROP FUNCTION IF EXISTS get_organization_as_of_date(varchar, timestamptz);
DROP FUNCTION IF EXISTS get_organizations_as_of_date(timestamptz, boolean, integer, integer);
DROP FUNCTION IF EXISTS cleanup_expired_temporal_data(integer, boolean);

-- 删除表（自动级联索引/约束）
DROP TABLE IF EXISTS organization_timeline_events;
DROP TABLE IF EXISTS organization_unit_versions;

-- 删除枚举类型（确认无残留依赖）
DROP TYPE IF EXISTS timeline_event_type;
DROP TYPE IF EXISTS timeline_event_status;

COMMIT;

SELECT '035_drop_temporal_history_tables applied' AS status, NOW() AS applied_at;
