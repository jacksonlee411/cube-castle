-- =============================================
-- CQRS时态管理PostgreSQL优化
-- 专注CUD操作，移除层级计算职责
-- =============================================

BEGIN;

-- 1. 备份当前数据
CREATE TABLE IF NOT EXISTS organization_units_backup_cqrs AS 
SELECT * FROM organization_units WHERE 1=1;

-- 2. 移除层级计算相关字段和索引
-- PostgreSQL专注存储，不负责层级路径计算
ALTER TABLE organization_units 
DROP COLUMN IF EXISTS level CASCADE,
DROP COLUMN IF EXISTS path CASCADE;

-- 删除层级相关索引
DROP INDEX IF EXISTS idx_org_hierarchy_temporal;
DROP INDEX IF EXISTS idx_org_units_type_level;
DROP INDEX IF EXISTS idx_org_units_path_gin;

-- 3. 优化时态索引，专注CUD操作
-- 主查询：按日期查找特定组织的有效记录
CREATE INDEX idx_temporal_cud_primary 
ON organization_units(code, effective_date DESC, is_current) 
WHERE effective_date IS NOT NULL;

-- 命令操作：快速查找当前有效记录进行更新
CREATE INDEX idx_temporal_current_update 
ON organization_units(tenant_id, code, is_current) 
WHERE is_current = true;

-- 时态范围查询：支持特定日期查询
CREATE INDEX idx_temporal_date_lookup 
ON organization_units(tenant_id, effective_date, end_date) 
WHERE effective_date IS NOT NULL;

-- CDC发布优化：快速识别变更记录
CREATE INDEX idx_temporal_cdc_changes 
ON organization_units(updated_at DESC, code, effective_date) 
WHERE updated_at >= CURRENT_DATE - INTERVAL '1 day';

-- 4. 添加时态数据完整性约束
-- 确保同一组织在同一时间点只有一条有效记录
CREATE UNIQUE INDEX uk_temporal_single_current 
ON organization_units(tenant_id, code) 
WHERE is_current = true;

-- 确保时态日期逻辑正确
ALTER TABLE organization_units 
ADD CONSTRAINT check_temporal_date_logic 
CHECK (
    (end_date IS NULL AND is_current = true) OR 
    (end_date IS NOT NULL AND effective_date <= end_date)
);

-- 5. 优化时态触发器，专注数据完整性
CREATE OR REPLACE FUNCTION temporal_cud_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- INSERT操作：自动设置时态字段
    IF TG_OP = 'INSERT' THEN
        -- 如果没有指定生效日期，默认为今天
        IF NEW.effective_date IS NULL THEN
            NEW.effective_date := CURRENT_DATE;
        END IF;
        
        -- 如果没有指定当前标志，默认为true
        IF NEW.is_current IS NULL THEN
            NEW.is_current := true;
        END IF;
        
        -- 更新时间戳
        NEW.updated_at := NOW();
        
        RETURN NEW;
    END IF;
    
    -- UPDATE操作：维护时态完整性
    IF TG_OP = 'UPDATE' THEN
        -- 如果更新当前记录为非当前，必须设置结束日期
        IF OLD.is_current = true AND NEW.is_current = false AND NEW.end_date IS NULL THEN
            NEW.end_date := CURRENT_DATE;
        END IF;
        
        -- 更新时间戳
        NEW.updated_at := NOW();
        
        RETURN NEW;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 重建时态触发器
DROP TRIGGER IF EXISTS temporal_cud_trigger ON organization_units;
CREATE TRIGGER temporal_cud_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION temporal_cud_trigger();

-- 6. 创建时态CUD操作函数
-- 创建新的时态组织记录
CREATE OR REPLACE FUNCTION create_temporal_organization(
    p_tenant_id UUID,
    p_code VARCHAR(10),
    p_parent_code VARCHAR(10),
    p_name VARCHAR(255),
    p_unit_type VARCHAR(50),
    p_status VARCHAR(20) DEFAULT 'ACTIVE',
    p_effective_date DATE DEFAULT CURRENT_DATE,
    p_change_reason VARCHAR(500) DEFAULT NULL
) RETURNS BOOLEAN AS $$
BEGIN
    INSERT INTO organization_units (
        tenant_id, code, parent_code, name, unit_type, status,
        effective_date, is_current, change_reason, is_temporal
    ) VALUES (
        p_tenant_id, p_code, p_parent_code, p_name, p_unit_type, p_status,
        p_effective_date, true, p_change_reason, true
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql;

-- 更新时态组织记录（创建新版本）
CREATE OR REPLACE FUNCTION update_temporal_organization(
    p_tenant_id UUID,
    p_code VARCHAR(10),
    p_name VARCHAR(255) DEFAULT NULL,
    p_unit_type VARCHAR(50) DEFAULT NULL,
    p_status VARCHAR(20) DEFAULT NULL,
    p_parent_code VARCHAR(10) DEFAULT NULL,
    p_effective_date DATE DEFAULT CURRENT_DATE,
    p_change_reason VARCHAR(500) DEFAULT NULL
) RETURNS BOOLEAN AS $$
DECLARE
    current_record organization_units%ROWTYPE;
BEGIN
    -- 获取当前有效记录
    SELECT * INTO current_record 
    FROM organization_units 
    WHERE tenant_id = p_tenant_id 
      AND code = p_code 
      AND is_current = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION '未找到组织代码: %', p_code;
    END IF;
    
    -- 结束当前记录
    UPDATE organization_units 
    SET is_current = false, 
        end_date = p_effective_date - INTERVAL '1 day',
        updated_at = NOW()
    WHERE tenant_id = p_tenant_id 
      AND code = p_code 
      AND is_current = true;
    
    -- 创建新版本记录
    INSERT INTO organization_units (
        tenant_id, code, parent_code, name, unit_type, status,
        effective_date, is_current, change_reason, is_temporal
    ) VALUES (
        p_tenant_id, 
        p_code, 
        COALESCE(p_parent_code, current_record.parent_code),
        COALESCE(p_name, current_record.name),
        COALESCE(p_unit_type, current_record.unit_type),
        COALESCE(p_status, current_record.status),
        p_effective_date,
        true,
        p_change_reason,
        true
    );
    
    RETURN true;
EXCEPTION
    WHEN OTHERS THEN
        RETURN false;
END;
$$ LANGUAGE plpgsql;

-- 按特定日期查询组织（PostgreSQL端基础查询）
CREATE OR REPLACE FUNCTION get_organization_by_date(
    p_tenant_id UUID,
    p_code VARCHAR(10),
    p_as_of_date DATE DEFAULT CURRENT_DATE
) RETURNS TABLE (
    code VARCHAR(10),
    parent_code VARCHAR(10),
    name VARCHAR(255),
    unit_type VARCHAR(50),
    status VARCHAR(20),
    effective_date DATE,
    end_date DATE,
    is_current BOOLEAN,
    change_reason VARCHAR(500)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ou.code, ou.parent_code, ou.name, ou.unit_type, ou.status,
        ou.effective_date, ou.end_date, ou.is_current, ou.change_reason
    FROM organization_units ou
    WHERE ou.tenant_id = p_tenant_id
      AND ou.code = p_code
      AND ou.effective_date <= p_as_of_date
      AND (ou.end_date IS NULL OR ou.end_date >= p_as_of_date)
    ORDER BY ou.effective_date DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- 7. 更新CDC发布配置
-- 确保时态变更能够正确发布到Kafka
DROP PUBLICATION IF EXISTS temporal_organization_publication;
CREATE PUBLICATION temporal_organization_publication 
FOR TABLE organization_units;

-- 8. 创建时态性能监控视图
CREATE OR REPLACE VIEW temporal_cud_performance AS
SELECT 
    'total_temporal_records' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE is_temporal = true

UNION ALL

SELECT 
    'current_active_records' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE is_current = true

UNION ALL

SELECT 
    'historical_records' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE is_current = false

UNION ALL

SELECT 
    'records_last_30d' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE updated_at >= CURRENT_DATE - INTERVAL '30 days';

COMMENT ON VIEW temporal_cud_performance IS 'PostgreSQL时态CUD操作性能监控';

COMMIT;

-- 验证优化结果
SELECT 
    'PostgreSQL时态优化完成' as status,
    COUNT(*) as total_records,
    COUNT(CASE WHEN is_current = true THEN 1 END) as current_records,
    COUNT(CASE WHEN is_temporal = true THEN 1 END) as temporal_records
FROM organization_units;