-- 每日cutover任务：维护时态数据一致性
-- 用途：每天运行一次，更新is_current和is_future状态标志
-- 执行时间：建议在业务低峰期（如凌晨2点）运行
-- 作者：时态数据一致性系统
-- 版本：v1.0

BEGIN;

-- 1. 更新is_current标志：将过期的当前记录标记为非当前
UPDATE organization_units 
SET 
    is_current = false,
    updated_at = NOW()
WHERE 
    is_current = true 
    AND end_date IS NOT NULL 
    AND end_date <= CURRENT_DATE;

-- 2. 更新is_current标志：将生效的记录标记为当前
UPDATE organization_units 
SET 
    is_current = true,
    updated_at = NOW()
WHERE 
    is_current = false 
    AND effective_date <= CURRENT_DATE 
    AND (end_date IS NULL OR end_date > CURRENT_DATE)
    -- 确保每个组织只有一个当前版本
    AND NOT EXISTS (
        SELECT 1 FROM organization_units ou2 
        WHERE ou2.tenant_id = organization_units.tenant_id 
            AND ou2.code = organization_units.code 
            AND ou2.is_current = true
            AND ou2.record_id != organization_units.record_id
    );

-- 3. 更新is_future标志：标记未来生效的记录
UPDATE organization_units 
SET 
    is_future = CASE 
        WHEN effective_date > CURRENT_DATE THEN true 
        ELSE false 
    END,
    updated_at = NOW()
WHERE 
    is_future != (effective_date > CURRENT_DATE);

-- 4. 数据一致性验证：检查是否有重复的当前记录
DO $$
DECLARE
    duplicate_count integer;
BEGIN
    SELECT COUNT(*) INTO duplicate_count
    FROM (
        SELECT tenant_id, code, COUNT(*) as current_count
        FROM organization_units 
        WHERE is_current = true
        GROUP BY tenant_id, code
        HAVING COUNT(*) > 1
    ) duplicates;
    
    IF duplicate_count > 0 THEN
        RAISE EXCEPTION '发现 % 个组织存在多个当前版本记录，数据一致性检查失败', duplicate_count;
    END IF;
    
    RAISE NOTICE '数据一致性检查通过：无重复当前记录';
END $$;

-- 5. 统计更新结果
DO $$
DECLARE
    current_records_count integer;
    future_records_count integer;
    historical_records_count integer;
BEGIN
    -- 统计当前记录数
    SELECT COUNT(*) INTO current_records_count
    FROM organization_units WHERE is_current = true;
    
    -- 统计未来记录数
    SELECT COUNT(*) INTO future_records_count
    FROM organization_units WHERE is_future = true;
    
    -- 统计历史记录数
    SELECT COUNT(*) INTO historical_records_count
    FROM organization_units WHERE is_current = false AND is_future = false;
    
    RAISE NOTICE 'Cutover任务完成统计:';
    RAISE NOTICE '- 当前记录数: %', current_records_count;
    RAISE NOTICE '- 未来记录数: %', future_records_count;
    RAISE NOTICE '- 历史记录数: %', historical_records_count;
    RAISE NOTICE '- 总记录数: %', current_records_count + future_records_count + historical_records_count;
END $$;

COMMIT;

-- 记录执行日志
INSERT INTO audit_logs (
    operation_type,
    business_entity_type,
    business_entity_id,
    operated_by,
    operation_reason,
    changes_summary,
    created_at
) VALUES (
    'SYSTEM_MAINTENANCE',
    'TEMPORAL_DATA',
    'DAILY_CUTOVER',
    'SYSTEM',
    '每日cutover任务：维护时态数据一致性',
    'is_current和is_future状态标志更新完成',
    NOW()
);