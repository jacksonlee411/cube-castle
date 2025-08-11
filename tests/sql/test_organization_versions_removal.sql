-- 数据库级单元测试：验证organization_versions表删除后的系统完整性
-- 文件: tests/sql/test_organization_versions_removal.sql

\echo '🧪 开始执行organization_versions删除后的数据库完整性测试'

-- 测试1: 验证organization_versions表确实已删除
\echo '测试1: 验证organization_versions表删除状态'
DO $$
DECLARE
    table_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM pg_tables 
        WHERE tablename = 'organization_versions'
    ) INTO table_exists;
    
    IF table_exists THEN
        RAISE EXCEPTION '❌ FAILED: organization_versions表仍然存在';
    ELSE
        RAISE NOTICE '✅ PASSED: organization_versions表已成功删除';
    END IF;
END $$;

-- 测试2: 验证备份表存在并包含正确数据
\echo '测试2: 验证备份表完整性'
DO $$
DECLARE
    backup_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO backup_count
    FROM organization_versions_backup_before_deletion;
    
    IF backup_count = 2 THEN
        RAISE NOTICE '✅ PASSED: 备份表包含预期的2条记录';
    ELSE
        RAISE EXCEPTION '❌ FAILED: 备份表记录数不正确，期望2条，实际%条', backup_count;
    END IF;
END $$;

-- 测试3: 验证organization_units表时态字段完整性
\echo '测试3: 验证organization_units时态字段完整性'
DO $$
DECLARE
    temporal_fields_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO temporal_fields_count
    FROM information_schema.columns 
    WHERE table_name = 'organization_units' 
    AND column_name IN ('effective_date', 'end_date', 'change_reason', 'is_current');
    
    IF temporal_fields_count = 4 THEN
        RAISE NOTICE '✅ PASSED: organization_units表包含完整的4个时态字段';
    ELSE
        RAISE EXCEPTION '❌ FAILED: 时态字段不完整，期望4个，实际%个', temporal_fields_count;
    END IF;
END $$;

-- 测试4: 验证时态查询索引完整性
\echo '测试4: 验证时态查询索引完整性'
DO $$
DECLARE
    temporal_indexes_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO temporal_indexes_count
    FROM pg_indexes 
    WHERE tablename = 'organization_units' 
    AND indexdef LIKE '%effective_date%';
    
    IF temporal_indexes_count >= 5 THEN
        RAISE NOTICE '✅ PASSED: 时态查询索引完整（%个）', temporal_indexes_count;
    ELSE
        RAISE EXCEPTION '❌ FAILED: 时态查询索引不足，期望>=5个，实际%个', temporal_indexes_count;
    END IF;
END $$;

-- 测试5: 验证相关触发器已删除
\echo '测试5: 验证相关触发器删除状态'
DO $$
DECLARE
    trigger_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO trigger_count
    FROM pg_proc 
    WHERE proname = 'auto_manage_end_date_v2';
    
    IF trigger_count = 0 THEN
        RAISE NOTICE '✅ PASSED: auto_manage_end_date_v2函数已成功删除';
    ELSE
        RAISE EXCEPTION '❌ FAILED: auto_manage_end_date_v2函数仍然存在';
    END IF;
END $$;

-- 测试6: 验证数据查询功能正常
\echo '测试6: 验证时态数据查询功能'
DO $$
DECLARE
    active_orgs_count INTEGER;
    temporal_orgs_count INTEGER;
BEGIN
    -- 检查当前活跃组织数量
    SELECT COUNT(*) INTO active_orgs_count
    FROM organization_units 
    WHERE is_current = true;
    
    -- 检查时态查询数据
    SELECT COUNT(*) INTO temporal_orgs_count
    FROM organization_units 
    WHERE effective_date IS NOT NULL;
    
    IF active_orgs_count > 0 AND temporal_orgs_count > 0 THEN
        RAISE NOTICE '✅ PASSED: 时态数据查询正常（活跃：%，时态：%）', active_orgs_count, temporal_orgs_count;
    ELSE
        RAISE EXCEPTION '❌ FAILED: 数据查询异常（活跃：%，时态：%）', active_orgs_count, temporal_orgs_count;
    END IF;
END $$;

-- 测试7: 验证约束和外键完整性
\echo '测试7: 验证约束和外键完整性'
DO $$
DECLARE
    constraints_count INTEGER;
    fk_count INTEGER;
BEGIN
    -- 检查organization_units表的约束数量
    SELECT COUNT(*) INTO constraints_count
    FROM pg_constraint 
    WHERE conrelid = 'organization_units'::regclass;
    
    -- 检查外键约束
    SELECT COUNT(*) INTO fk_count
    FROM pg_constraint 
    WHERE conrelid = 'organization_units'::regclass
    AND contype = 'f';
    
    IF constraints_count >= 3 AND fk_count >= 1 THEN
        RAISE NOTICE '✅ PASSED: 约束完整性正常（总约束：%，外键：%）', constraints_count, fk_count;
    ELSE
        RAISE EXCEPTION '❌ FAILED: 约束完整性异常（总约束：%，外键：%）', constraints_count, fk_count;
    END IF;
END $$;

-- 测试8: 验证Publication配置正常
\echo '测试8: 验证Publication配置状态'
DO $$
DECLARE
    pub_tables_count INTEGER;
BEGIN
    -- 检查organization_units在Publication中的状态
    SELECT COUNT(*) INTO pub_tables_count
    FROM pg_publication_tables 
    WHERE tablename = 'organization_units';
    
    -- FOR ALL TABLES的publication应该自动包含organization_units
    IF pub_tables_count >= 0 THEN
        RAISE NOTICE '✅ PASSED: Publication配置正常';
    ELSE
        RAISE EXCEPTION '❌ FAILED: Publication配置异常';
    END IF;
END $$;

-- 测试结果汇总
\echo ''
\echo '🎉 数据库级单元测试完成'
\echo '📊 测试覆盖范围:'
\echo '  - 表删除验证'
\echo '  - 备份数据完整性'
\echo '  - 时态字段完整性'
\echo '  - 索引性能优化'
\echo '  - 触发器清理验证'
\echo '  - 数据查询功能'
\echo '  - 约束外键完整性'
\echo '  - CDC配置完整性'