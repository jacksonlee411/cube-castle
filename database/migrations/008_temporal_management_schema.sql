-- =============================================
-- 时态管理数据库结构扩展
-- 支持组织架构的时态查询和版本管理
-- =============================================

-- 1. 扩展现有组织单元表，添加时态字段
ALTER TABLE organization_units 
ADD COLUMN IF NOT EXISTS effective_from TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS effective_to TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS is_temporal BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS change_reason TEXT,
ADD COLUMN IF NOT EXISTS changed_by UUID,
ADD COLUMN IF NOT EXISTS approved_by UUID,
ADD COLUMN IF NOT EXISTS metadata JSONB;

-- 创建时态查询索引
CREATE INDEX IF NOT EXISTS idx_organization_units_effective_from ON organization_units(effective_from);
CREATE INDEX IF NOT EXISTS idx_organization_units_effective_to ON organization_units(effective_to);
CREATE INDEX IF NOT EXISTS idx_organization_units_temporal ON organization_units(is_temporal, effective_from, effective_to);
CREATE INDEX IF NOT EXISTS idx_organization_units_version ON organization_units(code, version);

-- 2. 创建组织单元历史版本表
CREATE TABLE IF NOT EXISTS organization_unit_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(7) NOT NULL,
    version INTEGER NOT NULL,
    
    -- 业务字段快照
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE', 'PLANNED', 'DELETED')),
    level INTEGER NOT NULL,
    path TEXT NOT NULL,
    sort_order INTEGER DEFAULT 0,
    description TEXT,
    parent_code VARCHAR(7),
    
    -- 时态管理字段
    effective_from TIMESTAMPTZ NOT NULL,
    effective_to TIMESTAMPTZ,
    is_temporal BOOLEAN DEFAULT TRUE,
    change_reason TEXT,
    changed_by UUID,
    approved_by UUID,
    metadata JSONB,
    
    -- 审计字段
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 外键约束
    FOREIGN KEY (organization_code) REFERENCES organization_units(code) ON DELETE CASCADE,
    FOREIGN KEY (parent_code) REFERENCES organization_units(code) ON DELETE SET NULL,
    
    -- 唯一约束
    UNIQUE(organization_code, version)
);

-- 版本表索引
CREATE INDEX IF NOT EXISTS idx_org_versions_code ON organization_unit_versions(organization_code);
CREATE INDEX IF NOT EXISTS idx_org_versions_effective ON organization_unit_versions(effective_from, effective_to);
CREATE INDEX IF NOT EXISTS idx_org_versions_temporal ON organization_unit_versions(organization_code, effective_from, effective_to);
CREATE INDEX IF NOT EXISTS idx_org_versions_tenant ON organization_unit_versions(tenant_id);

-- 3. 创建时间线事件表
CREATE TYPE IF NOT EXISTS timeline_event_type AS ENUM (
    'create', 'update', 'delete', 'activate', 'deactivate',
    'restructure', 'merge', 'split', 'transfer', 'rename'
);

CREATE TYPE IF NOT EXISTS timeline_event_status AS ENUM (
    'pending', 'approved', 'rejected', 'completed', 'cancelled'
);

CREATE TABLE IF NOT EXISTS organization_timeline_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(7) NOT NULL,
    
    -- 事件基本信息
    event_type timeline_event_type NOT NULL,
    event_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_date TIMESTAMPTZ,
    status timeline_event_status NOT NULL DEFAULT 'pending',
    
    -- 事件内容
    title VARCHAR(255) NOT NULL,
    description TEXT,
    metadata JSONB,
    
    -- 变更信息
    previous_value JSONB,
    new_value JSONB,
    affected_fields TEXT[],
    
    -- 责任人信息
    triggered_by UUID,
    approved_by UUID,
    
    -- 审计字段
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 外键约束
    FOREIGN KEY (organization_code) REFERENCES organization_units(code) ON DELETE CASCADE
);

-- 时间线事件索引
CREATE INDEX IF NOT EXISTS idx_timeline_events_org_code ON organization_timeline_events(organization_code);
CREATE INDEX IF NOT EXISTS idx_timeline_events_date ON organization_timeline_events(event_date);
CREATE INDEX IF NOT EXISTS idx_timeline_events_type ON organization_timeline_events(event_type);
CREATE INDEX IF NOT EXISTS idx_timeline_events_status ON organization_timeline_events(status);
CREATE INDEX IF NOT EXISTS idx_timeline_events_effective ON organization_timeline_events(effective_date);
CREATE INDEX IF NOT EXISTS idx_timeline_events_tenant ON organization_timeline_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_timeline_events_composite ON organization_timeline_events(organization_code, event_date, status);

-- 4. 创建时态查询视图
CREATE OR REPLACE VIEW organization_temporal_current AS
SELECT 
    ou.*,
    CASE 
        WHEN ou.effective_from IS NULL AND ou.effective_to IS NULL THEN 'always_active'
        WHEN ou.effective_from <= NOW() AND (ou.effective_to IS NULL OR ou.effective_to > NOW()) THEN 'currently_active'
        WHEN ou.effective_from > NOW() THEN 'future_active' 
        WHEN ou.effective_to <= NOW() THEN 'expired'
        ELSE 'unknown'
    END as temporal_status
FROM organization_units ou
WHERE (ou.effective_from IS NULL OR ou.effective_from <= NOW())
  AND (ou.effective_to IS NULL OR ou.effective_to > NOW())
  AND ou.status IN ('ACTIVE', 'PLANNED');

-- 5. 创建时态查询函数
CREATE OR REPLACE FUNCTION get_organization_as_of_date(
    p_org_code VARCHAR(7),
    p_as_of_date TIMESTAMPTZ DEFAULT NOW()
) RETURNS TABLE (
    code VARCHAR(7),
    name VARCHAR(255),
    unit_type VARCHAR(50),
    status VARCHAR(20),
    level INTEGER,
    path TEXT,
    sort_order INTEGER,
    description TEXT,
    parent_code VARCHAR(7),
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    version INTEGER,
    change_reason TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ou.code, ou.name, ou.unit_type, ou.status, ou.level, ou.path,
        ou.sort_order, ou.description, ou.parent_code,
        ou.effective_from, ou.effective_to, ou.version, ou.change_reason,
        ou.created_at, ou.updated_at
    FROM organization_units ou
    WHERE ou.code = p_org_code
      AND (ou.effective_from IS NULL OR ou.effective_from <= p_as_of_date)
      AND (ou.effective_to IS NULL OR ou.effective_to > p_as_of_date)
    
    UNION ALL
    
    SELECT 
        ouv.organization_code, ouv.name, ouv.unit_type, ouv.status, ouv.level, ouv.path,
        ouv.sort_order, ouv.description, ouv.parent_code,
        ouv.effective_from, ouv.effective_to, ouv.version, ouv.change_reason,
        ouv.created_at, ouv.updated_at
    FROM organization_unit_versions ouv
    WHERE ouv.organization_code = p_org_code
      AND ouv.effective_from <= p_as_of_date
      AND (ouv.effective_to IS NULL OR ouv.effective_to > p_as_of_date)
    
    ORDER BY version DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- 6. 创建批量时态查询函数
CREATE OR REPLACE FUNCTION get_organizations_as_of_date(
    p_as_of_date TIMESTAMPTZ DEFAULT NOW(),
    p_include_inactive BOOLEAN DEFAULT FALSE,
    p_limit INTEGER DEFAULT 50,
    p_offset INTEGER DEFAULT 0
) RETURNS TABLE (
    code VARCHAR(7),
    name VARCHAR(255),
    unit_type VARCHAR(50),
    status VARCHAR(20),
    level INTEGER,
    path TEXT,
    sort_order INTEGER,
    description TEXT,
    parent_code VARCHAR(7),
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    version INTEGER,
    temporal_status TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    WITH temporal_orgs AS (
        -- 当前表中的时态数据
        SELECT 
            ou.code, ou.name, ou.unit_type, ou.status, ou.level, ou.path,
            ou.sort_order, ou.description, ou.parent_code,
            ou.effective_from, ou.effective_to, ou.version, ou.change_reason,
            ou.created_at, ou.updated_at,
            CASE 
                WHEN ou.effective_from <= p_as_of_date AND (ou.effective_to IS NULL OR ou.effective_to > p_as_of_date) THEN 'active'
                WHEN ou.effective_from > p_as_of_date THEN 'future'
                WHEN ou.effective_to <= p_as_of_date THEN 'expired'
                ELSE 'unknown'
            END as temporal_status,
            1 as priority
        FROM organization_units ou
        WHERE (ou.effective_from IS NULL OR ou.effective_from <= p_as_of_date)
          AND (ou.effective_to IS NULL OR ou.effective_to > p_as_of_date)
          AND (p_include_inactive OR ou.status != 'INACTIVE')
        
        UNION ALL
        
        -- 历史版本表中的时态数据
        SELECT 
            ouv.organization_code, ouv.name, ouv.unit_type, ouv.status, ouv.level, ouv.path,
            ouv.sort_order, ouv.description, ouv.parent_code,
            ouv.effective_from, ouv.effective_to, ouv.version, ouv.change_reason,
            ouv.created_at, ouv.updated_at,
            CASE 
                WHEN ouv.effective_from <= p_as_of_date AND (ouv.effective_to IS NULL OR ouv.effective_to > p_as_of_date) THEN 'active'
                WHEN ouv.effective_from > p_as_of_date THEN 'future'
                WHEN ouv.effective_to <= p_as_of_date THEN 'expired'
                ELSE 'unknown'
            END as temporal_status,
            2 as priority
        FROM organization_unit_versions ouv
        WHERE ouv.effective_from <= p_as_of_date
          AND (ouv.effective_to IS NULL OR ouv.effective_to > p_as_of_date)
          AND (p_include_inactive OR ouv.status != 'INACTIVE')
    ),
    deduped_orgs AS (
        SELECT DISTINCT ON (code) *
        FROM temporal_orgs
        WHERE temporal_status = 'active'
        ORDER BY code, priority, version DESC
    )
    SELECT 
        do.code, do.name, do.unit_type, do.status, do.level, do.path,
        do.sort_order, do.description, do.parent_code,
        do.effective_from, do.effective_to, do.version, do.temporal_status,
        do.created_at, do.updated_at
    FROM deduped_orgs do
    ORDER BY do.level, do.sort_order, do.name
    LIMIT p_limit OFFSET p_offset;
END;
$$ LANGUAGE plpgsql;

-- 7. 创建版本管理触发器
CREATE OR REPLACE FUNCTION organization_version_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- 当组织单元更新时，如果启用时态管理，则创建历史版本
    IF TG_OP = 'UPDATE' AND OLD.is_temporal = TRUE THEN
        -- 将旧版本插入历史版本表
        INSERT INTO organization_unit_versions (
            organization_code, version, name, unit_type, status, level, path,
            sort_order, description, parent_code, effective_from, effective_to,
            is_temporal, change_reason, changed_by, approved_by, metadata,
            tenant_id
        ) VALUES (
            OLD.code, OLD.version, OLD.name, OLD.unit_type, OLD.status, OLD.level, OLD.path,
            OLD.sort_order, OLD.description, OLD.parent_code, OLD.effective_from, OLD.effective_to,
            OLD.is_temporal, OLD.change_reason, OLD.changed_by, OLD.approved_by, OLD.metadata,
            OLD.tenant_id
        );
        
        -- 增加新版本号
        NEW.version = COALESCE(OLD.version, 0) + 1;
        NEW.updated_at = NOW();
        
        -- 记录时间线事件
        INSERT INTO organization_timeline_events (
            organization_code, event_type, event_date, effective_date, status,
            title, description, metadata, previous_value, new_value,
            triggered_by, tenant_id
        ) VALUES (
            NEW.code, 'update', NOW(), NEW.effective_from, 'completed',
            '组织信息更新', NEW.change_reason,
            jsonb_build_object('version', NEW.version),
            to_jsonb(OLD.*), to_jsonb(NEW.*),
            NEW.changed_by, NEW.tenant_id
        );
    END IF;
    
    -- 创建操作记录事件
    IF TG_OP = 'INSERT' THEN
        INSERT INTO organization_timeline_events (
            organization_code, event_type, event_date, effective_date, status,
            title, description, triggered_by, tenant_id
        ) VALUES (
            NEW.code, 'create', NOW(), NEW.effective_from, 'completed',
            '创建组织单元', '新建组织: ' || NEW.name,
            NEW.changed_by, NEW.tenant_id
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS organization_version_trigger ON organization_units;
CREATE TRIGGER organization_version_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION organization_version_trigger();

-- 8. 创建时态数据清理函数
CREATE OR REPLACE FUNCTION cleanup_expired_temporal_data(
    p_retention_days INTEGER DEFAULT 730, -- 默认保留2年
    p_dry_run BOOLEAN DEFAULT TRUE
) RETURNS TABLE (
    action TEXT,
    table_name TEXT,
    affected_rows INTEGER
) AS $$
DECLARE
    expired_versions_count INTEGER;
    expired_events_count INTEGER;
    cutoff_date TIMESTAMPTZ;
BEGIN
    cutoff_date := NOW() - (p_retention_days || ' days')::INTERVAL;
    
    -- 统计将要清理的过期版本数据
    SELECT COUNT(*) INTO expired_versions_count
    FROM organization_unit_versions
    WHERE effective_to IS NOT NULL 
      AND effective_to < cutoff_date;
    
    -- 统计将要清理的过期事件数据  
    SELECT COUNT(*) INTO expired_events_count
    FROM organization_timeline_events
    WHERE status = 'completed'
      AND event_date < cutoff_date
      AND event_type NOT IN ('create', 'delete'); -- 保留关键事件
    
    -- 返回预览信息
    RETURN QUERY SELECT 'PREVIEW'::TEXT, 'organization_unit_versions'::TEXT, expired_versions_count;
    RETURN QUERY SELECT 'PREVIEW'::TEXT, 'organization_timeline_events'::TEXT, expired_events_count;
    
    -- 如果不是dry run，执行实际清理
    IF NOT p_dry_run THEN
        -- 清理过期版本数据
        DELETE FROM organization_unit_versions
        WHERE effective_to IS NOT NULL 
          AND effective_to < cutoff_date;
        
        GET DIAGNOSTICS expired_versions_count = ROW_COUNT;
        
        -- 清理过期事件数据
        DELETE FROM organization_timeline_events
        WHERE status = 'completed'
          AND event_date < cutoff_date
          AND event_type NOT IN ('create', 'delete');
        
        GET DIAGNOSTICS expired_events_count = ROW_COUNT;
        
        RETURN QUERY SELECT 'DELETED'::TEXT, 'organization_unit_versions'::TEXT, expired_versions_count;
        RETURN QUERY SELECT 'DELETED'::TEXT, 'organization_timeline_events'::TEXT, expired_events_count;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 9. 插入示例时态数据 (用于测试)
INSERT INTO organization_units (
    code, name, unit_type, status, level, path, sort_order,
    effective_from, effective_to, is_temporal, version, change_reason,
    tenant_id
) VALUES 
(
    '1000001', '研发部门', 'DEPARTMENT', 'ACTIVE', 1, '/1000001', 1,
    '2024-01-01 00:00:00+00'::TIMESTAMPTZ, 
    '2024-06-30 23:59:59+00'::TIMESTAMPTZ,
    TRUE, 1, '部门初始创建',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'::UUID
),
(
    '1000002', '产品部门', 'DEPARTMENT', 'PLANNED', 2, '/1000002', 2,
    '2024-07-01 00:00:00+00'::TIMESTAMPTZ, 
    NULL,
    TRUE, 1, '新部门规划',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'::UUID
)
ON CONFLICT (code) DO NOTHING;

-- 10. 创建时态查询性能监控视图
CREATE OR REPLACE VIEW temporal_performance_stats AS
SELECT 
    'current_temporal_orgs' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE is_temporal = TRUE
  AND (effective_to IS NULL OR effective_to > NOW())

UNION ALL

SELECT 
    'total_versions' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_unit_versions

UNION ALL

SELECT 
    'timeline_events_last_30d' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_timeline_events 
WHERE event_date >= NOW() - INTERVAL '30 days'

UNION ALL

SELECT 
    'expired_temporal_orgs' as metric,
    COUNT(*) as value,
    NOW() as collected_at
FROM organization_units 
WHERE is_temporal = TRUE
  AND effective_to IS NOT NULL 
  AND effective_to <= NOW();

COMMENT ON VIEW temporal_performance_stats IS '时态管理性能统计视图';

-- 创建完成提示
SELECT 
    'Temporal database schema setup completed successfully!' as status,
    COUNT(*) as total_tables
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name LIKE '%organization%';