-- =============================================
-- CQRS时态管理PostgreSQL优化 - 修复版
-- 专注CUD操作，移除层级计算职责
-- =============================================
BEGIN;

-- 1. 备份当前数据
CREATE TABLE IF NOT EXISTS organization_units_backup_cqrs AS
SELECT * FROM organization_units WHERE 1=1;

-- 2. 清理旧的非必需索引
DROP INDEX IF EXISTS idx_org_hierarchy_temporal;
DROP INDEX IF EXISTS idx_org_units_type_level;
DROP INDEX IF EXISTS idx_org_units_path_gin;

-- 3. 创建时态优化索引（修复版）
-- 时态查询主索引 - 移除IMMUTABLE函数
CREATE INDEX IF NOT EXISTS idx_organization_temporal_main
ON organization_units(tenant_id, code, effective_date DESC, is_current)
WHERE effective_date IS NOT NULL;

-- 当前版本快速查询索引
CREATE INDEX IF NOT EXISTS idx_organization_current_only
ON organization_units(tenant_id, code, is_current)
WHERE is_current = true;

-- 时间范围查询索引
CREATE INDEX IF NOT EXISTS idx_organization_date_range
ON organization_units(tenant_id, effective_date, end_date);

-- 4. 验证现有数据结构
DO $$
DECLARE
    missing_columns TEXT[] := '{}';
    col_name TEXT;
BEGIN
    -- 检查必需的时态字段
    FOR col_name IN SELECT unnest(ARRAY['effective_date', 'end_date', 'is_current', 'change_reason'])
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'organization_units' 
            AND column_name = col_name
        ) THEN
            missing_columns := array_append(missing_columns, col_name);
        END IF;
    END LOOP;
    
    IF array_length(missing_columns, 1) > 0 THEN
        RAISE NOTICE '缺少时态字段: %', array_to_string(missing_columns, ', ');
        
        -- 添加缺失字段
        IF 'effective_date' = ANY(missing_columns) THEN
            ALTER TABLE organization_units ADD COLUMN effective_date DATE;
        END IF;
        
        IF 'end_date' = ANY(missing_columns) THEN
            ALTER TABLE organization_units ADD COLUMN end_date DATE;
        END IF;
        
        IF 'is_current' = ANY(missing_columns) THEN
            ALTER TABLE organization_units ADD COLUMN is_current BOOLEAN DEFAULT true;
        END IF;
        
        IF 'change_reason' = ANY(missing_columns) THEN
            ALTER TABLE organization_units ADD COLUMN change_reason TEXT;
        END IF;
    END IF;
END $$;

-- 5. 数据清理和标准化
-- 确保所有记录都有有效的时态字段
UPDATE organization_units 
SET is_current = true 
WHERE is_current IS NULL;

UPDATE organization_units 
SET effective_date = created_at::date 
WHERE effective_date IS NULL;

-- 6. 添加时态约束（可选，生产环境建议）
-- 确保时态数据的完整性
DO $$
BEGIN
    -- 添加检查约束：结束日期必须晚于生效日期
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'organization_units' 
        AND constraint_name = 'chk_org_temporal_dates'
    ) THEN
        ALTER TABLE organization_units 
        ADD CONSTRAINT chk_org_temporal_dates 
        CHECK (end_date IS NULL OR end_date > effective_date);
    END IF;
    
    -- 添加唯一约束：同一组织在同一时间只能有一个当前版本
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'organization_units' 
        AND constraint_name = 'uq_org_current_version'
    ) THEN
        CREATE UNIQUE INDEX uq_org_current_version
        ON organization_units(tenant_id, code)
        WHERE is_current = true;
    END IF;
END $$;

-- 7. 创建时态查询辅助函数
CREATE OR REPLACE FUNCTION get_organization_at_date(
    p_tenant_id UUID,
    p_code VARCHAR(20),
    p_as_of_date DATE DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    code VARCHAR(20),
    name VARCHAR(200),
    unit_type VARCHAR(50),
    status VARCHAR(20),
    effective_date DATE,
    end_date DATE,
    is_current BOOLEAN
) LANGUAGE SQL STABLE AS $$
    SELECT 
        ou.code,
        ou.name,
        ou.unit_type,
        ou.status,
        ou.effective_date,
        ou.end_date,
        ou.is_current
    FROM organization_units ou
    WHERE ou.tenant_id = p_tenant_id
      AND ou.code = p_code
      AND COALESCE(ou.effective_date, CURRENT_DATE) <= p_as_of_date
      AND (ou.end_date IS NULL OR ou.end_date > p_as_of_date)
    ORDER BY ou.effective_date DESC
    LIMIT 1;
$$;

-- 8. 创建层级计算辅助视图（简化版）
CREATE OR REPLACE VIEW organization_hierarchy_current AS
SELECT 
    ou.tenant_id,
    ou.code,
    ou.parent_code,
    ou.name,
    ou.unit_type,
    ou.status,
    ou.level,
    ou.path,
    ou.effective_date,
    ou.end_date,
    ou.is_current
FROM organization_units ou
WHERE ou.is_current = true
  AND (ou.end_date IS NULL OR ou.end_date > CURRENT_DATE);

-- 9. 优化查询统计信息
ANALYZE organization_units;

-- 10. 验证优化结果
SELECT 
    '时态优化完成' as status,
    count(*) as total_records,
    count(*) FILTER (WHERE is_current = true) as current_records,
    count(*) FILTER (WHERE effective_date IS NOT NULL AND is_current = false) as historical_records
FROM organization_units;

COMMIT;

-- 11. 验证索引创建
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'organization_units'
  AND indexname LIKE 'idx_organization_%'
ORDER BY indexname;