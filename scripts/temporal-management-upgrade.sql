-- ============================================================================
-- 时态管理系统升级SQL脚本
-- 功能：自动结束日期管理 + 五状态生命周期管理
-- 版本：v2.0
-- 创建时间：2025-08-18
-- ============================================================================

BEGIN;

-- ============================================================================
-- 第一部分：Schema升级 - 添加新的状态管理字段
-- ============================================================================

-- 1. 添加新的状态管理字段
ALTER TABLE organization_units 
ADD COLUMN IF NOT EXISTS lifecycle_status VARCHAR(20) DEFAULT 'CURRENT',
ADD COLUMN IF NOT EXISTS business_status VARCHAR(20) DEFAULT 'ACTIVE', 
ADD COLUMN IF NOT EXISTS data_status VARCHAR(20) DEFAULT 'NORMAL',
ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS suspended_by UUID,
ADD COLUMN IF NOT EXISTS suspension_reason TEXT,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS deleted_by UUID,
ADD COLUMN IF NOT EXISTS deletion_reason TEXT;

-- 2. 更新状态约束
ALTER TABLE organization_units DROP CONSTRAINT IF EXISTS organization_units_status_check;
ALTER TABLE organization_units ADD CONSTRAINT lifecycle_status_check 
    CHECK (lifecycle_status IN ('CURRENT', 'HISTORICAL', 'PLANNED'));
ALTER TABLE organization_units ADD CONSTRAINT business_status_check 
    CHECK (business_status IN ('ACTIVE', 'SUSPENDED'));
ALTER TABLE organization_units ADD CONSTRAINT data_status_check 
    CHECK (data_status IN ('NORMAL', 'DELETED'));

-- ============================================================================
-- 第二部分：自动结束日期管理函数和触发器
-- ============================================================================

-- 创建自动结束日期管理函数
CREATE OR REPLACE FUNCTION auto_manage_end_dates()
RETURNS TRIGGER AS $$
BEGIN
    -- 场景1: 插入新记录
    IF TG_OP = 'INSERT' THEN
        -- 1.1 查找同一组织的前一条有效记录，设置其结束日期
        UPDATE organization_units 
        SET end_date = (NEW.effective_date - INTERVAL '1 day')::date,
            updated_at = NOW()
        WHERE code = NEW.code 
          AND tenant_id = NEW.tenant_id
          AND data_status = 'NORMAL'
          AND effective_date < NEW.effective_date
          AND end_date IS NULL
          AND record_id != NEW.record_id;  -- 避免自我引用
        
        -- 1.2 查找同一组织的后续记录，更新当前记录的结束日期
        UPDATE organization_units 
        SET end_date = (
            SELECT MIN(effective_date - INTERVAL '1 day')::date 
            FROM organization_units future 
            WHERE future.code = NEW.code 
              AND future.tenant_id = NEW.tenant_id
              AND future.data_status = 'NORMAL'
              AND future.effective_date > NEW.effective_date
              AND future.record_id != NEW.record_id
        )
        WHERE record_id = NEW.record_id;
        
        RETURN NEW;
    END IF;
    
    -- 场景2: 更新记录 - 生效日期变更
    IF TG_OP = 'UPDATE' THEN
        -- 如果修改了生效日期，需要重新计算时间轴
        IF OLD.effective_date != NEW.effective_date AND NEW.data_status = 'NORMAL' THEN
            -- 2.1 重置原来影响的前序记录
            UPDATE organization_units 
            SET end_date = (
                SELECT MIN(effective_date - INTERVAL '1 day')::date 
                FROM organization_units next_records 
                WHERE next_records.code = NEW.code 
                  AND next_records.tenant_id = NEW.tenant_id
                  AND next_records.data_status = 'NORMAL'
                  AND next_records.effective_date > organization_units.effective_date
                  AND next_records.record_id != NEW.record_id
            ),
            updated_at = NOW()
            WHERE code = NEW.code 
              AND tenant_id = NEW.tenant_id
              AND data_status = 'NORMAL'
              AND effective_date < NEW.effective_date
              AND record_id != NEW.record_id;
            
            -- 2.2 重新计算当前记录的结束日期
            UPDATE organization_units 
            SET end_date = (
                SELECT MIN(effective_date - INTERVAL '1 day')::date 
                FROM organization_units future 
                WHERE future.code = NEW.code 
                  AND future.tenant_id = NEW.tenant_id
                  AND future.data_status = 'NORMAL'
                  AND future.effective_date > NEW.effective_date
                  AND future.record_id != NEW.record_id
            )
            WHERE record_id = NEW.record_id;
        END IF;
        
        RETURN NEW;
    END IF;
    
    -- 场景3: 软删除记录 - 重新连接时间轴
    IF TG_OP = 'UPDATE' AND OLD.data_status = 'NORMAL' AND NEW.data_status = 'DELETED' THEN
        -- 查找前一条记录，扩展其结束日期到下一条记录
        UPDATE organization_units prev
        SET end_date = (
            SELECT MIN(effective_date - INTERVAL '1 day')::date 
            FROM organization_units future 
            WHERE future.code = NEW.code 
              AND future.tenant_id = NEW.tenant_id
              AND future.data_status = 'NORMAL'
              AND future.effective_date > prev.effective_date
              AND future.record_id != NEW.record_id
        ),
        updated_at = NOW()
        WHERE prev.code = NEW.code 
          AND prev.tenant_id = NEW.tenant_id
          AND prev.data_status = 'NORMAL'
          AND prev.effective_date < NEW.effective_date
          AND prev.record_id != NEW.record_id
          AND prev.end_date = (NEW.effective_date - INTERVAL '1 day')::date;
        
        RETURN NEW;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 删除现有触发器（如果存在）
DROP TRIGGER IF EXISTS auto_end_date_trigger ON organization_units;

-- 创建新的自动结束日期管理触发器
CREATE TRIGGER auto_end_date_trigger
    AFTER INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION auto_manage_end_dates();

-- ============================================================================
-- 第三部分：生命周期状态自动更新函数
-- ============================================================================

-- 创建生命周期状态自动更新函数
CREATE OR REPLACE FUNCTION auto_update_lifecycle_status()
RETURNS TRIGGER AS $$
BEGIN
    -- 自动计算 lifecycle_status
    IF TG_OP = 'INSERT' OR (TG_OP = 'UPDATE' AND OLD.effective_date != NEW.effective_date) THEN
        -- 设置生命周期状态
        NEW.lifecycle_status = CASE
            WHEN NEW.effective_date <= CURRENT_DATE 
                 AND (NEW.end_date IS NULL OR NEW.end_date > CURRENT_DATE) THEN 'CURRENT'
            WHEN NEW.effective_date > CURRENT_DATE THEN 'PLANNED'
            ELSE 'HISTORICAL'
        END;
        
        -- 设置 is_current 标志
        NEW.is_current = (NEW.lifecycle_status = 'CURRENT');
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 删除现有触发器（如果存在）
DROP TRIGGER IF EXISTS auto_lifecycle_status_trigger ON organization_units;

-- 创建生命周期状态触发器
CREATE TRIGGER auto_lifecycle_status_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION auto_update_lifecycle_status();

-- ============================================================================
-- 第四部分：创建优化索引
-- ============================================================================

-- 时态查询优化索引
CREATE INDEX IF NOT EXISTS idx_org_temporal_query_optimized 
    ON organization_units(tenant_id, code, data_status, effective_date DESC, end_date DESC NULLS LAST)
    WHERE data_status = 'NORMAL';

-- 生命周期状态查询索引
CREATE INDEX IF NOT EXISTS idx_org_lifecycle_query 
    ON organization_units(tenant_id, lifecycle_status, business_status, data_status)
    WHERE data_status = 'NORMAL';

-- 当前记录快速查询索引
CREATE INDEX IF NOT EXISTS idx_org_current_records 
    ON organization_units(tenant_id, code, lifecycle_status)
    WHERE lifecycle_status = 'CURRENT' AND data_status = 'NORMAL';

-- 软删除记录查询索引
CREATE INDEX IF NOT EXISTS idx_org_soft_deleted 
    ON organization_units(tenant_id, data_status, deleted_at DESC)
    WHERE data_status = 'DELETED';

-- ============================================================================
-- 第五部分：数据完整性约束
-- ============================================================================

-- 时间序列约束
ALTER TABLE organization_units 
DROP CONSTRAINT IF EXISTS check_temporal_sequence;
ALTER TABLE organization_units 
ADD CONSTRAINT check_temporal_sequence 
CHECK (effective_date <= COALESCE(end_date + INTERVAL '1 day', '9999-12-31'::date));

-- 唯一性约束，防止同一时间点重复记录
DROP INDEX IF EXISTS idx_unique_temporal_point;
CREATE UNIQUE INDEX idx_unique_temporal_point 
ON organization_units(code, effective_date, tenant_id)
WHERE data_status = 'NORMAL';

-- 软删除约束
ALTER TABLE organization_units 
ADD CONSTRAINT check_deleted_metadata 
CHECK (
    (data_status = 'DELETED' AND deleted_at IS NOT NULL) OR 
    (data_status != 'DELETED' AND deleted_at IS NULL)
);

-- 停用约束
ALTER TABLE organization_units 
ADD CONSTRAINT check_suspended_metadata 
CHECK (
    (business_status = 'SUSPENDED' AND suspended_at IS NOT NULL) OR 
    (business_status != 'SUSPENDED' AND suspended_at IS NULL)
);

COMMIT;

-- ============================================================================
-- 升级完成信息
-- ============================================================================
DO $$ 
BEGIN 
    RAISE NOTICE '============================================';
    RAISE NOTICE '时态管理系统升级完成！';
    RAISE NOTICE '新功能：';
    RAISE NOTICE '1. 自动结束日期管理';
    RAISE NOTICE '2. 五状态生命周期管理';
    RAISE NOTICE '3. 优化的查询索引';
    RAISE NOTICE '4. 数据完整性约束';
    RAISE NOTICE '============================================';
END $$;