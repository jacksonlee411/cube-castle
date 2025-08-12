-- =============================================
-- CQRS时态管理PostgreSQL优化 - 安全版
-- 先清理数据问题，再进行优化
-- =============================================
BEGIN;

-- 1. 备份当前数据
DROP TABLE IF EXISTS organization_units_backup_temporal;
CREATE TABLE organization_units_backup_temporal AS
SELECT * FROM organization_units;

-- 2. 数据清理：修复日期冲突问题
-- 修复end_date <= effective_date的问题
UPDATE organization_units 
SET end_date = effective_date + INTERVAL '1 day'
WHERE end_date IS NOT NULL AND end_date <= effective_date;

-- 3. 清理旧索引
DROP INDEX IF EXISTS idx_org_hierarchy_temporal;
DROP INDEX IF EXISTS idx_org_units_type_level; 
DROP INDEX IF EXISTS idx_org_units_path_gin;

-- 4. 创建时态优化索引
CREATE INDEX IF NOT EXISTS idx_organization_temporal_main
ON organization_units(tenant_id, code, effective_date DESC NULLS LAST, is_current);

CREATE INDEX IF NOT EXISTS idx_organization_current_only
ON organization_units(tenant_id, code)
WHERE is_current = true;

CREATE INDEX IF NOT EXISTS idx_organization_date_range
ON organization_units(tenant_id, effective_date, end_date);

-- 5. 确保数据完整性
UPDATE organization_units 
SET is_current = true 
WHERE is_current IS NULL;

UPDATE organization_units 
SET effective_date = created_at::date 
WHERE effective_date IS NULL;

-- 6. 创建时态查询函数
CREATE OR REPLACE FUNCTION get_organization_temporal(
    p_tenant_id UUID,
    p_code VARCHAR(20),
    p_as_of_date DATE DEFAULT CURRENT_DATE
)
RETURNS TABLE (
    code VARCHAR(20),
    name VARCHAR(200),
    unit_type VARCHAR(50),
    status VARCHAR(20),
    parent_code VARCHAR(20),
    effective_date DATE,
    end_date DATE,
    is_current BOOLEAN,
    change_reason TEXT
) LANGUAGE SQL STABLE AS $$
    SELECT 
        ou.code,
        ou.name,
        ou.unit_type,
        ou.status,
        ou.parent_code,
        ou.effective_date,
        ou.end_date,
        ou.is_current,
        ou.change_reason
    FROM organization_units ou
    WHERE ou.tenant_id = p_tenant_id
      AND ou.code = p_code
      AND COALESCE(ou.effective_date, CURRENT_DATE) <= p_as_of_date
      AND (ou.end_date IS NULL OR ou.end_date > p_as_of_date)
    ORDER BY ou.effective_date DESC
    LIMIT 1;
$$;

-- 7. 创建当前组织视图
CREATE OR REPLACE VIEW organization_current AS
SELECT 
    ou.tenant_id,
    ou.code,
    ou.parent_code,
    ou.name,
    ou.unit_type,
    ou.status,
    ou.level,
    ou.path,
    ou.sort_order,
    ou.description,
    ou.profile,
    ou.effective_date,
    ou.end_date,
    ou.is_current,
    ou.change_reason,
    ou.created_at,
    ou.updated_at
FROM organization_units ou
WHERE ou.is_current = true
  AND (ou.end_date IS NULL OR ou.end_date > CURRENT_DATE);

-- 8. 更新统计信息
ANALYZE organization_units;

-- 9. 验证优化结果
SELECT 
    'PostgreSQL时态优化完成 - 安全版' as status,
    count(*) as total_records,
    count(*) FILTER (WHERE is_current = true) as current_records,
    count(*) FILTER (WHERE effective_date IS NOT NULL AND is_current = false) as historical_records,
    count(*) FILTER (WHERE end_date IS NOT NULL AND end_date <= effective_date) as date_conflicts
FROM organization_units;

COMMIT;

-- 最终验证
SELECT 
    indexname,
    indexdef
FROM pg_indexes 
WHERE tablename = 'organization_units'
  AND indexname LIKE 'idx_organization_%'
ORDER BY indexname;