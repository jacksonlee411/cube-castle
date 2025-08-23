-- ===============================================
-- Operation Type字段迁移脚本
-- 目标：简化时间戳结构，使用操作类型语义化时间戳含义
-- 作者：Claude Code Assistant  
-- 创建时间：2025-08-23
-- ===============================================

BEGIN;

-- 1. 添加 operation_type 字段
-- ===============================================
ALTER TABLE organization_units 
ADD COLUMN IF NOT EXISTS operation_type VARCHAR(20) DEFAULT 'CREATE';

-- 2. 创建操作类型枚举约束
-- ===============================================
ALTER TABLE organization_units 
ADD CONSTRAINT operation_type_check 
CHECK (operation_type IN ('CREATE', 'UPDATE', 'SUSPEND', 'REACTIVATE', 'DELETE'));

-- 3. 根据现有数据设置 operation_type 值
-- ===============================================

-- 3.1 标记删除操作
UPDATE organization_units 
SET operation_type = 'DELETE' 
WHERE deleted_at IS NOT NULL OR status = 'DELETED';

-- 3.2 标记暂停操作  
UPDATE organization_units 
SET operation_type = 'SUSPEND' 
WHERE suspended_at IS NOT NULL AND operation_type != 'DELETE';

-- 3.3 标记重新激活操作 (暂停后又变为ACTIVE的记录)
UPDATE organization_units 
SET operation_type = 'REACTIVATE' 
WHERE status = 'ACTIVE' 
  AND suspended_at IS NOT NULL 
  AND operation_type != 'DELETE'
  AND updated_at > suspended_at;

-- 3.4 标记创建操作 (created_at = updated_at 的记录)
UPDATE organization_units 
SET operation_type = 'CREATE' 
WHERE ABS(EXTRACT(EPOCH FROM (created_at - updated_at))) < 1 
  AND operation_type NOT IN ('DELETE', 'SUSPEND', 'REACTIVATE');

-- 3.5 其他记录标记为更新操作
UPDATE organization_units 
SET operation_type = 'UPDATE' 
WHERE operation_type NOT IN ('CREATE', 'DELETE', 'SUSPEND', 'REACTIVATE');

-- 4. 创建索引优化查询性能
-- ===============================================
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_operation_type 
    ON organization_units(tenant_id, operation_type, updated_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_operation_status 
    ON organization_units(operation_type, status, is_current) 
    WHERE is_current = true;

-- 5. 更新触发器以自动设置operation_type
-- ===============================================
CREATE OR REPLACE FUNCTION set_operation_type()
RETURNS TRIGGER AS $$
BEGIN
    -- 插入操作默认为CREATE
    IF TG_OP = 'INSERT' THEN
        NEW.operation_type = COALESCE(NEW.operation_type, 'CREATE');
        RETURN NEW;
    END IF;
    
    -- 更新操作根据状态变化判断
    IF TG_OP = 'UPDATE' THEN
        -- 如果明确设置了operation_type，保持不变
        IF NEW.operation_type IS NOT NULL AND NEW.operation_type != OLD.operation_type THEN
            RETURN NEW;
        END IF;
        
        -- 根据状态变化自动判断操作类型
        IF OLD.status != 'DELETED' AND NEW.status = 'DELETED' THEN
            NEW.operation_type = 'DELETE';
        ELSIF OLD.status != 'SUSPENDED' AND NEW.status = 'SUSPENDED' THEN
            NEW.operation_type = 'SUSPEND';
        ELSIF OLD.status = 'SUSPENDED' AND NEW.status = 'ACTIVE' THEN
            NEW.operation_type = 'REACTIVATE';
        ELSE
            NEW.operation_type = 'UPDATE';
        END IF;
        
        RETURN NEW;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS set_operation_type_trigger ON organization_units;
CREATE TRIGGER set_operation_type_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION set_operation_type();

-- 6. 数据验证查询
-- ===============================================
DO $$
DECLARE
    total_records INTEGER;
    create_count INTEGER;
    update_count INTEGER;
    suspend_count INTEGER;
    reactivate_count INTEGER;
    delete_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_records FROM organization_units;
    SELECT COUNT(*) INTO create_count FROM organization_units WHERE operation_type = 'CREATE';
    SELECT COUNT(*) INTO update_count FROM organization_units WHERE operation_type = 'UPDATE';
    SELECT COUNT(*) INTO suspend_count FROM organization_units WHERE operation_type = 'SUSPEND';
    SELECT COUNT(*) INTO reactivate_count FROM organization_units WHERE operation_type = 'REACTIVATE';
    SELECT COUNT(*) INTO delete_count FROM organization_units WHERE operation_type = 'DELETE';
    
    RAISE NOTICE '=== Operation Type 迁移统计 ===';
    RAISE NOTICE '总记录数: %', total_records;
    RAISE NOTICE 'CREATE: %', create_count;
    RAISE NOTICE 'UPDATE: %', update_count;
    RAISE NOTICE 'SUSPEND: %', suspend_count;
    RAISE NOTICE 'REACTIVATE: %', reactivate_count;
    RAISE NOTICE 'DELETE: %', delete_count;
    RAISE NOTICE '验证: % + % + % + % + % = %', 
        create_count, update_count, suspend_count, reactivate_count, delete_count,
        create_count + update_count + suspend_count + reactivate_count + delete_count;
    
    -- 验证总数一致性
    IF total_records != create_count + update_count + suspend_count + reactivate_count + delete_count THEN
        RAISE EXCEPTION '操作类型统计不匹配！请检查迁移逻辑。';
    END IF;
    
    RAISE NOTICE '✅ Operation Type 迁移验证通过';
END $$;

COMMIT;

-- 7. 查看迁移结果示例
-- ===============================================
SELECT 
    operation_type,
    status,
    COUNT(*) as count,
    MIN(created_at) as earliest,
    MAX(updated_at) as latest
FROM organization_units 
GROUP BY operation_type, status 
ORDER BY operation_type, status;

-- 查看最近的操作记录
SELECT 
    code, 
    name,
    operation_type,
    status,
    created_at,
    updated_at,
    EXTRACT(EPOCH FROM (updated_at - created_at)) as seconds_diff
FROM organization_units 
WHERE is_current = true
ORDER BY updated_at DESC 
LIMIT 10;