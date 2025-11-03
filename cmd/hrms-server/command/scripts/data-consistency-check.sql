-- 数据一致性检查脚本
-- 用途：检查时态数据的一致性和完整性
-- 执行频率：每日运行，或在关键操作后运行
-- 作者：时态数据一致性系统
-- 版本：v1.0

-- 检查结果汇总表
CREATE TEMP TABLE consistency_check_results (
    check_name VARCHAR(100),
    status VARCHAR(20),
    issue_count INTEGER DEFAULT 0,
    details TEXT
);

-- 1. 检查重复的当前记录
DO $$
DECLARE
    duplicate_count integer := 0;
    duplicate_details text := '';
BEGIN
    -- 查找重复的当前记录
    SELECT COUNT(*), STRING_AGG(tenant_id || ':' || code, ', ') 
    INTO duplicate_count, duplicate_details
    FROM (
        SELECT tenant_id, code, COUNT(*) as current_count
        FROM organization_units 
        WHERE is_current = true
        GROUP BY tenant_id, code
        HAVING COUNT(*) > 1
    ) duplicates;
    
    INSERT INTO consistency_check_results VALUES (
        'DUPLICATE_CURRENT_RECORDS',
        CASE WHEN duplicate_count = 0 THEN 'PASS' ELSE 'FAIL' END,
        duplicate_count,
        CASE WHEN duplicate_count = 0 
             THEN '无重复当前记录' 
             ELSE '发现重复当前记录: ' || duplicate_details 
        END
    );
END $$;

-- 2. 检查缺失的当前记录
DO $$
DECLARE
    missing_count integer := 0;
    missing_details text := '';
BEGIN
    -- 查找没有当前记录的组织
    SELECT COUNT(*), STRING_AGG(tenant_id || ':' || code, ', ')
    INTO missing_count, missing_details
    FROM (
        SELECT DISTINCT tenant_id, code
        FROM organization_units
        WHERE (tenant_id, code) NOT IN (
            SELECT tenant_id, code 
            FROM organization_units 
            WHERE is_current = true
        )
        -- 排除所有版本都是未来生效的组织
        AND (tenant_id, code) NOT IN (
            SELECT tenant_id, code
            FROM organization_units
            GROUP BY tenant_id, code
            HAVING MIN(effective_date) > CURRENT_DATE
        )
    ) missing;
    
    INSERT INTO consistency_check_results VALUES (
        'MISSING_CURRENT_RECORDS',
        CASE WHEN missing_count = 0 THEN 'PASS' ELSE 'FAIL' END,
        missing_count,
        CASE WHEN missing_count = 0 
             THEN '无缺失当前记录' 
             ELSE '缺失当前记录: ' || missing_details 
        END
    );
END $$;

-- 3. 检查时间线重叠
DO $$
DECLARE
    overlap_count integer := 0;
    overlap_details text := '';
BEGIN
    -- 查找时间线重叠的记录
    WITH overlaps AS (
        SELECT 
            o1.tenant_id, o1.code,
            o1.record_id as record1, o2.record_id as record2,
            o1.effective_date as start1, COALESCE(o1.end_date, '9999-12-31'::date) as end1,
            o2.effective_date as start2, COALESCE(o2.end_date, '9999-12-31'::date) as end2
        FROM organization_units o1
        JOIN organization_units o2 ON (
            o1.tenant_id = o2.tenant_id 
            AND o1.code = o2.code 
            AND o1.record_id != o2.record_id
        )
        WHERE 
            o1.effective_date < COALESCE(o2.end_date, '9999-12-31'::date)
            AND o2.effective_date < COALESCE(o1.end_date, '9999-12-31'::date)
    )
    SELECT COUNT(*), STRING_AGG(DISTINCT tenant_id || ':' || code, ', ')
    INTO overlap_count, overlap_details
    FROM overlaps;
    
    INSERT INTO consistency_check_results VALUES (
        'TIMELINE_OVERLAPS',
        CASE WHEN overlap_count = 0 THEN 'PASS' ELSE 'FAIL' END,
        overlap_count,
        CASE WHEN overlap_count = 0 
             THEN '无时间线重叠' 
             ELSE '发现时间线重叠: ' || overlap_details 
        END
    );
END $$;

-- 4. 检查is_current标志一致性
DO $$
DECLARE
    inconsistent_count integer := 0;
    inconsistent_details text := '';
BEGIN
    -- 查找is_current标志不一致的记录
    WITH current_check AS (
        SELECT 
            tenant_id, code, record_id,
            is_current,
            CASE 
                WHEN effective_date <= CURRENT_DATE 
                     AND (end_date IS NULL OR end_date > CURRENT_DATE) 
                THEN true 
                ELSE false 
            END as should_be_current
        FROM organization_units
    )
    SELECT COUNT(*), STRING_AGG(tenant_id || ':' || code || ':' || record_id, ', ')
    INTO inconsistent_count, inconsistent_details
    FROM current_check
    WHERE is_current != should_be_current;
    
    INSERT INTO consistency_check_results VALUES (
        'CURRENT_FLAG_CONSISTENCY',
        CASE WHEN inconsistent_count = 0 THEN 'PASS' ELSE 'FAIL' END,
        inconsistent_count,
        CASE WHEN inconsistent_count = 0 
             THEN 'is_current标志一致' 
             ELSE 'is_current标志不一致: ' || inconsistent_details 
        END
    );
END $$;

-- 5. 检查is_future标志一致性
DO $$
DECLARE
    inconsistent_count integer := 0;
    inconsistent_details text := '';
BEGIN
    -- 查找is_future标志不一致的记录
    SELECT COUNT(*), STRING_AGG(tenant_id || ':' || code || ':' || record_id, ', ')
    INTO inconsistent_count, inconsistent_details
    FROM organization_units
    WHERE is_future != (effective_date > CURRENT_DATE);
    
    INSERT INTO consistency_check_results VALUES (
        'FUTURE_FLAG_CONSISTENCY',
        CASE WHEN inconsistent_count = 0 THEN 'PASS' ELSE 'FAIL' END,
        inconsistent_count,
        CASE WHEN inconsistent_count = 0 
             THEN 'is_future标志一致' 
             ELSE 'is_future标志不一致: ' || inconsistent_details 
        END
    );
END $$;

-- 6. 检查外键引用完整性
DO $$
DECLARE
    orphan_count integer := 0;
    orphan_details text := '';
BEGIN
    -- 查找父级组织不存在的记录
    SELECT COUNT(*), STRING_AGG(tenant_id || ':' || code, ', ')
    INTO orphan_count, orphan_details
    FROM organization_units o1
    WHERE 
        parent_code IS NOT NULL
        AND NOT EXISTS (
            SELECT 1 FROM organization_units o2 
            WHERE o2.tenant_id = o1.tenant_id 
                AND o2.code = o1.parent_code 
                AND o2.is_current = true
        );
    
    INSERT INTO consistency_check_results VALUES (
        'PARENT_REFERENCE_INTEGRITY',
        CASE WHEN orphan_count = 0 THEN 'PASS' ELSE 'WARN' END,
        orphan_count,
        CASE WHEN orphan_count = 0 
             THEN '父级引用完整' 
             ELSE '发现孤立记录: ' || orphan_details 
        END
    );
END $$;

-- 7. 生成检查报告
DO $$
DECLARE
    total_checks integer;
    failed_checks integer;
    warned_checks integer;
    passed_checks integer;
    overall_status text;
BEGIN
    -- 统计各种状态的检查数量
    SELECT COUNT(*), 
           COUNT(*) FILTER (WHERE status = 'FAIL'),
           COUNT(*) FILTER (WHERE status = 'WARN'),
           COUNT(*) FILTER (WHERE status = 'PASS')
    INTO total_checks, failed_checks, warned_checks, passed_checks
    FROM consistency_check_results;
    
    -- 确定总体状态
    overall_status := CASE 
        WHEN failed_checks > 0 THEN 'CRITICAL'
        WHEN warned_checks > 0 THEN 'WARNING'
        ELSE 'HEALTHY'
    END;
    
    -- 输出总体报告
    RAISE NOTICE '========================================';
    RAISE NOTICE '数据一致性检查报告';
    RAISE NOTICE '执行时间: %', NOW();
    RAISE NOTICE '总体状态: %', overall_status;
    RAISE NOTICE '总检查项: %, 通过: %, 警告: %, 失败: %', 
                 total_checks, passed_checks, warned_checks, failed_checks;
    RAISE NOTICE '========================================';
END $$;

-- 8. 输出详细检查结果
DO $$
DECLARE
    check_record RECORD;
BEGIN
    FOR check_record IN 
        SELECT * FROM consistency_check_results ORDER BY 
            CASE status 
                WHEN 'FAIL' THEN 1 
                WHEN 'WARN' THEN 2 
                WHEN 'PASS' THEN 3 
            END, check_name
    LOOP
        RAISE NOTICE '[%] %: %', 
                     check_record.status,
                     check_record.check_name, 
                     check_record.details;
    END LOOP;
END $$;

-- 9. 记录检查结果到审计日志
INSERT INTO audit_logs (
    operation_type,
    business_entity_type,
    business_entity_id,
    operated_by,
    operation_reason,
    changes_summary,
    created_at
) 
SELECT 
    'SYSTEM_CHECK',
    'TEMPORAL_DATA',
    'CONSISTENCY_CHECK',
    'SYSTEM',
    '数据一致性检查',
    STRING_AGG(check_name || ':' || status || '(' || issue_count || ')', '; '),
    NOW()
FROM consistency_check_results;

-- 清理临时表
DROP TABLE consistency_check_results;