-- ===============================================
-- Phase 4 Week 1 Day 1: 清理Outbox系统数据库对象
-- ===============================================

-- 删除outbox.events表相关的RLS策略
DROP POLICY IF EXISTS outbox_events_tenant_isolation ON outbox.events;
DROP POLICY IF EXISTS outbox_events_insert_policy ON outbox.events;
DROP POLICY IF EXISTS tenant_isolation_outbox_events ON outbox.events;

-- 删除outbox相关的索引
DROP INDEX IF EXISTS idx_outbox_events_processed;
DROP INDEX IF EXISTS idx_outbox_events_aggregate;
DROP INDEX IF EXISTS idx_outbox_events_tenant_id;
DROP INDEX IF EXISTS idx_outbox_events_tenant_id_performance;

-- 删除outbox.events表
DROP TABLE IF EXISTS outbox.events;

-- 撤销outbox schema的权限
REVOKE ALL ON SCHEMA outbox FROM cube_castle_app;
REVOKE ALL ON ALL TABLES IN SCHEMA outbox FROM cube_castle_app;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA outbox FROM cube_castle_app;

-- 删除outbox schema
DROP SCHEMA IF EXISTS outbox CASCADE;

-- 验证清理结果
DO $$ 
BEGIN
    -- 检查outbox schema是否已删除
    IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'outbox') THEN
        RAISE NOTICE '✅ Outbox schema 已成功删除';
    ELSE
        RAISE NOTICE '❌ Outbox schema 删除失败';
    END IF;
    
    -- 检查outbox.events表是否已删除
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'outbox' AND table_name = 'events') THEN
        RAISE NOTICE '✅ Outbox.events 表已成功删除';
    ELSE
        RAISE NOTICE '❌ Outbox.events 表删除失败';
    END IF;
    
    RAISE NOTICE '==============================================';
    RAISE NOTICE '✅ Outbox系统数据库清理完成！';
    RAISE NOTICE '==============================================';
END $$;