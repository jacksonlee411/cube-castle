-- 触发器备份和回滚脚本
-- 创建日期: 2025-09-07
-- 用途: 触发器优化前的完整备份和回滚能力

-- ====================================================================
-- 第一部分：创建备份表
-- ====================================================================

-- 备份触发器定义
DROP TABLE IF EXISTS trigger_backup_20250907;
CREATE TABLE trigger_backup_20250907 AS
SELECT 
    tgname as trigger_name,
    tgenabled as enabled_status,
    pg_get_triggerdef(oid) as full_definition,
    current_timestamp as backup_time
FROM pg_trigger 
WHERE tgrelid = 'organization_units'::regclass
AND tgname NOT LIKE 'RI_ConstraintTrigger_%'; -- 排除外键约束触发器

-- 备份触发器函数
DROP TABLE IF EXISTS trigger_functions_backup_20250907;
CREATE TABLE trigger_functions_backup_20250907 AS
SELECT 
    proname as function_name,
    prosrc as function_source,
    pg_get_functiondef(oid) as full_definition,
    current_timestamp as backup_time
FROM pg_proc 
WHERE proname IN (
    'auto_manage_end_dates',
    'notify_organization_change', 
    'generate_org_unit_code',
    'smart_hierarchy_trigger',
    'enforce_soft_delete_temporal_flags',
    'simple_temporal_gap_fill_trigger',
    'calculate_org_hierarchy'
);

-- 验证备份
SELECT 
    'Triggers backed up: ' || count(*) as backup_status
FROM trigger_backup_20250907;

SELECT 
    'Functions backed up: ' || count(*) as backup_status  
FROM trigger_functions_backup_20250907;

-- ====================================================================
-- 第二部分：阶段1 - 禁用高冲突触发器
-- ====================================================================

-- 禁用时态处理冲突触发器（应用层RecalculateTimeline已覆盖）
ALTER TABLE organization_units DISABLE TRIGGER auto_end_date_trigger;
ALTER TABLE organization_units DISABLE TRIGGER simple_temporal_gap_fill_trigger;
ALTER TABLE organization_units DISABLE TRIGGER enforce_soft_delete_temporal_flags_trigger;

-- 禁用层级计算重叠触发器（历史脚本产生，职责重复）
ALTER TABLE organization_units DISABLE TRIGGER set_org_unit_code;
ALTER TABLE organization_units DISABLE TRIGGER smart_hierarchy_management;

-- 禁用可能无用的通知触发器
ALTER TABLE organization_units DISABLE TRIGGER organization_units_change_trigger;

-- 禁用冗余的时间戳更新触发器
ALTER TABLE organization_units DISABLE TRIGGER update_organization_units_updated_at;

-- 验证禁用状态
SELECT 
    tgname,
    CASE 
        WHEN tgenabled = 'O' THEN 'ENABLED'
        WHEN tgenabled = 'D' THEN 'DISABLED'
        ELSE 'UNKNOWN'
    END as status
FROM pg_trigger 
WHERE tgrelid = 'organization_units'::regclass
ORDER BY tgname;

-- ====================================================================
-- 第三部分：完整回滚脚本（紧急使用）
-- ====================================================================

/*
-- 紧急回滚：重新启用所有触发器
ALTER TABLE organization_units ENABLE TRIGGER auto_end_date_trigger;
ALTER TABLE organization_units ENABLE TRIGGER simple_temporal_gap_fill_trigger;
ALTER TABLE organization_units ENABLE TRIGGER enforce_soft_delete_temporal_flags_trigger;
ALTER TABLE organization_units ENABLE TRIGGER set_org_unit_code;
ALTER TABLE organization_units ENABLE TRIGGER smart_hierarchy_management;
ALTER TABLE organization_units ENABLE TRIGGER organization_units_change_trigger;
ALTER TABLE organization_units ENABLE TRIGGER update_organization_units_updated_at;

-- 验证回滚成功
SELECT 
    tgname,
    CASE 
        WHEN tgenabled = 'O' THEN 'ENABLED'
        WHEN tgenabled = 'D' THEN 'DISABLED'
        ELSE 'UNKNOWN'
    END as status
FROM pg_trigger 
WHERE tgrelid = 'organization_units'::regclass
ORDER BY tgname;
*/

-- ====================================================================
-- 第四部分：系统验证脚本
-- ====================================================================

-- 检查当前触发器状态
CREATE OR REPLACE FUNCTION check_trigger_health()
RETURNS TABLE(
    trigger_name text,
    status text,
    function_exists boolean,
    last_error text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        t.tgname::text,
        CASE 
            WHEN t.tgenabled = 'O' THEN 'ENABLED'
            WHEN t.tgenabled = 'D' THEN 'DISABLED'
            ELSE 'UNKNOWN'
        END::text as status,
        EXISTS(
            SELECT 1 FROM pg_proc p 
            WHERE p.oid = t.tgfoid
        ) as function_exists,
        ''::text as last_error
    FROM pg_trigger t
    WHERE t.tgrelid = 'organization_units'::regclass
    ORDER BY t.tgname;
END;
$$ LANGUAGE plpgsql;

-- 执行健康检查
SELECT * FROM check_trigger_health();

-- ====================================================================
-- 第五部分：测试数据验证
-- ====================================================================

-- 创建测试记录验证功能
DO $$
DECLARE
    test_code VARCHAR(7) := '9999999';
    test_record_count INTEGER;
BEGIN
    -- 清理可能存在的测试记录
    DELETE FROM organization_units WHERE code = test_code;
    
    -- 插入测试记录
    INSERT INTO organization_units (
        code, name, unit_type, tenant_id, 
        effective_date, operated_by_id, operated_by_name
    ) VALUES (
        test_code, '触发器测试组织', 'DEPARTMENT', 
        '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
        CURRENT_DATE, 
        '00000000-0000-0000-0000-000000000000',
        'System Test'
    );
    
    -- 验证记录创建
    SELECT COUNT(*) INTO test_record_count 
    FROM organization_units 
    WHERE code = test_code;
    
    IF test_record_count = 1 THEN
        RAISE NOTICE '✅ 测试记录创建成功 - 触发器禁用后系统正常';
    ELSE
        RAISE NOTICE '❌ 测试记录创建失败 - 需要检查系统状态';
    END IF;
    
    -- 清理测试记录
    DELETE FROM organization_units WHERE code = test_code;
    
    RAISE NOTICE '🧪 触发器健康检查完成';
END $$;

-- ====================================================================
-- 使用说明
-- ====================================================================

/*
📋 使用指南：

1. 执行备份（必须）：
   运行第一部分和第二部分脚本

2. 验证系统稳定性：
   - 运行完整测试套件
   - 监控应用日志24小时
   - 执行第五部分的测试脚本

3. 如需回滚：
   取消注释第三部分的回滚脚本并执行

4. 健康检查：
   定期运行 SELECT * FROM check_trigger_health();

⚠️  警告：
- 在生产环境执行前请在测试环境验证
- 确保有完整的数据库备份
- 建议在低峰期执行
*/