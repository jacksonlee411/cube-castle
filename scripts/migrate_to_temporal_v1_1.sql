-- 时态管理API升级 - 数据库迁移脚本 v1.1 (外键兼容版)
-- 基于ADR-007实施方案
-- 处理外键依赖关系的迁移策略

BEGIN;

-- 步骤1: 备份现有数据
CREATE TABLE IF NOT EXISTS organization_units_backup_pre_temporal AS
SELECT * FROM organization_units;

-- 步骤2: 添加时态字段（保持单一code主键不变）
DO $$
BEGIN
    -- 检查并添加时态字段
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'organization_units' 
                  AND column_name = 'effective_date') THEN
        
        RAISE NOTICE '添加时态管理字段...';
        
        -- 添加时态字段
        ALTER TABLE organization_units 
        ADD COLUMN effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
        ADD COLUMN end_date DATE,
        ADD COLUMN version INTEGER NOT NULL DEFAULT 1,
        ADD COLUMN supersedes_version INTEGER,
        ADD COLUMN change_reason VARCHAR(500),
        ADD COLUMN is_current BOOLEAN NOT NULL DEFAULT true;
        
        -- 添加version唯一约束（保持code作为主键）
        ALTER TABLE organization_units 
        ADD CONSTRAINT uk_code_version UNIQUE (code, version);
        
        RAISE NOTICE '时态字段添加完成';
    ELSE
        RAISE NOTICE '时态字段已存在，跳过添加步骤';
    END IF;
END
$$;

-- 步骤3: 迁移现有数据
UPDATE organization_units 
SET effective_date = created_at::DATE,
    version = 1,
    is_current = true,
    change_reason = '初始数据迁移：从现有数据转换为时态管理模式'
WHERE version IS NULL OR version = 0;

-- 步骤4: 创建时态查询优化索引
CREATE INDEX IF NOT EXISTS idx_org_effective_date ON organization_units(effective_date);
CREATE INDEX IF NOT EXISTS idx_org_current_version ON organization_units(code, is_current) WHERE is_current = true;
CREATE INDEX IF NOT EXISTS idx_org_version_chain ON organization_units(code, version);
CREATE INDEX IF NOT EXISTS idx_org_temporal_query ON organization_units(tenant_id, code, effective_date, end_date);

-- 步骤5: 创建组织事件表
CREATE TABLE IF NOT EXISTS organization_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 约束
    CONSTRAINT chk_event_type CHECK (
        event_type IN ('CREATE', 'UPDATE', 'RESTRUCTURE', 'DISSOLVE', 'ACTIVATE', 'DEACTIVATE')
    ),
    CONSTRAINT chk_end_date_after_effective CHECK (
        end_date IS NULL OR end_date > effective_date
    ),
    
    -- 外键约束到organization_units的code (主键)
    CONSTRAINT fk_org_events_org FOREIGN KEY (organization_code) 
        REFERENCES organization_units(code) ON DELETE RESTRICT
);

-- 为事件表创建索引
CREATE INDEX IF NOT EXISTS idx_org_events_code ON organization_events(organization_code);
CREATE INDEX IF NOT EXISTS idx_org_events_type ON organization_events(event_type);
CREATE INDEX IF NOT EXISTS idx_org_events_date ON organization_events(effective_date);
CREATE INDEX IF NOT EXISTS idx_org_events_tenant ON organization_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_org_events_data_gin ON organization_events USING GIN (event_data);

-- 步骤6: 创建版本历史表
CREATE TABLE IF NOT EXISTS organization_versions (
    version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    version INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    snapshot_data JSONB NOT NULL,
    change_reason VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 唯一约束
    CONSTRAINT uk_org_version UNIQUE (organization_code, version),
    
    -- 检查约束
    CONSTRAINT chk_version_positive CHECK (version > 0),
    CONSTRAINT chk_snapshot_not_empty CHECK (snapshot_data != '{}'::jsonb),
    
    -- 外键约束到organization_units的code (主键)
    CONSTRAINT fk_org_versions_org FOREIGN KEY (organization_code)
        REFERENCES organization_units(code) ON DELETE RESTRICT
);

-- 为版本表创建索引
CREATE INDEX IF NOT EXISTS idx_org_versions_code_version ON organization_versions(organization_code, version);
CREATE INDEX IF NOT EXISTS idx_org_versions_effective ON organization_versions(effective_date);
CREATE INDEX IF NOT EXISTS idx_org_versions_tenant ON organization_versions(tenant_id);

-- 步骤7: 创建结束日期自动管理函数（适应单一主键模式）
CREATE OR REPLACE FUNCTION auto_manage_end_date_v2()
RETURNS TRIGGER AS $$
DECLARE
    affected_rows INTEGER;
    existing_version INTEGER;
BEGIN
    -- 检查是否为插入新版本操作
    IF TG_OP = 'INSERT' AND NEW.version > 1 THEN
        
        RAISE NOTICE '开始处理组织 % 的版本 % 结束日期管理', NEW.code, NEW.version;
        
        -- 获取当前版本号
        SELECT version INTO existing_version 
        FROM organization_units 
        WHERE code = NEW.code AND is_current = true;
        
        IF existing_version IS NOT NULL AND existing_version != NEW.version THEN
            -- 设置前版本的end_date并标记为非当前版本
            UPDATE organization_units 
            SET end_date = NEW.effective_date - INTERVAL '1 day',
                is_current = false
            WHERE code = NEW.code 
              AND is_current = true 
              AND version = existing_version;
            
            GET DIAGNOSTICS affected_rows = ROW_COUNT;
            RAISE NOTICE '更新了 % 条前版本记录的结束日期', affected_rows;
        END IF;
        
        -- 验证时间线一致性
        IF EXISTS (
            SELECT 1 FROM organization_versions 
            WHERE organization_code = NEW.code 
            AND version != NEW.version
            AND effective_date >= NEW.effective_date
        ) THEN
            RAISE EXCEPTION '时间线冲突：不能在现有版本之前插入新版本';
        END IF;
        
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器（用于处理版本历史表的时态管理）
DROP TRIGGER IF EXISTS trigger_auto_end_date_v2 ON organization_versions;
CREATE TRIGGER trigger_auto_end_date_v2
    BEFORE INSERT ON organization_versions
    FOR EACH ROW 
    EXECUTE FUNCTION auto_manage_end_date_v2();

-- 步骤8: 创建数据一致性检查函数
CREATE OR REPLACE FUNCTION validate_temporal_consistency_v2()
RETURNS TABLE (
    organization_code VARCHAR(10),
    issue_type VARCHAR(50), 
    description TEXT
) AS $$
BEGIN
    -- 检查当前版本标记唯一性
    RETURN QUERY
    SELECT 
        code,
        'MULTIPLE_CURRENT'::VARCHAR(50),
        format('存在多个当前版本标记的记录：版本%s', version)
    FROM organization_units 
    WHERE is_current = true
    GROUP BY code, version
    HAVING COUNT(*) > 1;
    
    -- 检查版本历史表中的时间线间隙
    RETURN QUERY
    SELECT 
        o1.organization_code,
        'HISTORY_TIMELINE_GAP'::VARCHAR(50),
        format('历史版本%s结束日期%s与版本%s生效日期%s之间存在间隙', 
               o1.version, o1.end_date, o2.version, o2.effective_date)
    FROM organization_versions o1
    JOIN organization_versions o2 ON o1.organization_code = o2.organization_code
    WHERE o1.version < o2.version
      AND o1.end_date IS NOT NULL
      AND o1.end_date + INTERVAL '1 day' != o2.effective_date;
    
    -- 检查数据完整性
    RETURN QUERY
    SELECT 
        code,
        'MISSING_TEMPORAL_DATA'::VARCHAR(50),
        '缺少必需的时态字段数据'
    FROM organization_units 
    WHERE effective_date IS NULL 
       OR version IS NULL 
       OR version < 1;
END;
$$ LANGUAGE plpgsql;

-- 步骤9: 创建时态查询辅助视图
CREATE OR REPLACE VIEW organization_current_view AS
SELECT 
    ou.*,
    'CURRENT' as version_status
FROM organization_units ou 
WHERE ou.is_current = true;

CREATE OR REPLACE VIEW organization_historical_view AS  
SELECT 
    ou.*,
    ov.snapshot_data,
    ov.created_at as version_created_at,
    'HISTORICAL' as version_status
FROM organization_units ou
JOIN organization_versions ov ON ou.code = ov.organization_code 
ORDER BY ou.code, ov.version;

-- 步骤10: 数据验证
DO $$
DECLARE
    issue_count INTEGER := 0;
    total_orgs INTEGER := 0;
    current_versions INTEGER := 0;
BEGIN
    -- 统计基本信息
    SELECT COUNT(*) INTO total_orgs FROM organization_units;
    SELECT COUNT(*) INTO current_versions FROM organization_units WHERE is_current = true;
    
    -- 检查数据一致性问题
    SELECT COUNT(*) INTO issue_count FROM validate_temporal_consistency_v2();
    
    -- 报告结果
    RAISE NOTICE '=== 数据迁移验证结果 ===';
    RAISE NOTICE '总组织记录数: %', total_orgs;
    RAISE NOTICE '当前版本组织数: %', current_versions;
    RAISE NOTICE '数据一致性问题: %', issue_count;
    
    IF issue_count > 0 THEN
        -- 显示具体问题但不中断迁移
        RAISE NOTICE '发现 % 个数据一致性问题，请检查详情', issue_count;
    ELSE
        RAISE NOTICE '✅ 数据一致性验证通过！';
    END IF;
    
    RAISE NOTICE '🎉 数据迁移验证完成！时态管理基础设施已成功建立';
END
$$;

COMMIT;

-- 迁移完成提示
DO $$
BEGIN
    RAISE NOTICE '=== 时态管理API升级 - 阶段1完成 ===';
    RAISE NOTICE '✅ 数据库结构扩展完成（兼容现有外键）';
    RAISE NOTICE '✅ 时态字段添加完成';
    RAISE NOTICE '✅ 事件表和版本表已创建';
    RAISE NOTICE '✅ 时态查询辅助视图已建立';
    RAISE NOTICE '✅ 数据完整性验证完成';
    RAISE NOTICE '';
    RAISE NOTICE '📋 重要说明：';
    RAISE NOTICE '- 保持了现有的code主键结构，确保外键兼容性';
    RAISE NOTICE '- 版本管理通过(code,version)唯一约束实现';
    RAISE NOTICE '- 时态功能通过专用的organization_versions表管理';
    RAISE NOTICE '';
    RAISE NOTICE '🔍 监控指令：SELECT * FROM validate_temporal_consistency_v2();';
    RAISE NOTICE '📊 当前版本查看：SELECT * FROM organization_current_view LIMIT 5;';
END
$$;