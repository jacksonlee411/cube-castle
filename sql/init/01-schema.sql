-- Cube Castle组织架构管理系统数据库Schema
-- 版本: v4.2.1
-- 架构: PostgreSQL单一数据源 + 时态数据支持

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- 组织单元表 (单一数据源)
CREATE TABLE organization_units (
    -- 主键和业务标识
    code VARCHAR(7) PRIMARY KEY CHECK (code ~ '^[1-9][0-9]{6}$'),
    parent_code VARCHAR(7) REFERENCES organization_units(code),
    tenant_id UUID NOT NULL DEFAULT '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    
    -- 基础信息
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(20) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'ORGANIZATION_UNIT', 'COMPANY', 'PROJECT_TEAM')),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    
    -- 层级信息 (PostgreSQL原生优化)
    level INTEGER NOT NULL DEFAULT 1 CHECK (level BETWEEN 1 AND 17),
    hierarchy_depth INTEGER NOT NULL DEFAULT 1 CHECK (hierarchy_depth BETWEEN 1 AND 17),
    code_path TEXT NOT NULL DEFAULT '/',
    name_path TEXT NOT NULL DEFAULT '/',
    sort_order INTEGER NOT NULL DEFAULT 0,
    
    -- 扩展配置
    description TEXT,
    profile JSONB NOT NULL DEFAULT '{}',
    
    -- 时态信息
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT true,
    is_future BOOLEAN NOT NULL DEFAULT false,
    
    -- 审计信息
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    operation_type VARCHAR(20) NOT NULL DEFAULT 'CREATE' CHECK (operation_type IN ('CREATE', 'UPDATE', 'SUSPEND', 'REACTIVATE', 'DELETE')),
    operated_by_id UUID NOT NULL,
    operated_by_name VARCHAR(255) NOT NULL,
    operation_reason TEXT,
    record_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    
    -- 约束
    UNIQUE (code, effective_date, record_id),
    CHECK (end_date IS NULL OR end_date > effective_date),
    CHECK (NOT (is_current AND is_future)),
    CHECK (NOT (is_deleted AND is_current))
);


-- 审计日志表
CREATE TABLE audit_logs (
    audit_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL DEFAULT 'organization_unit',
    entity_code VARCHAR(7) NOT NULL,
    operation_type VARCHAR(20) NOT NULL,
    operation_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    operated_by_id UUID NOT NULL,
    operated_by_name VARCHAR(255) NOT NULL,
    operation_reason TEXT,
    before_data JSONB,
    after_data JSONB NOT NULL,
    changes_summary TEXT,
    business_entity_id VARCHAR(50),
    tenant_id UUID NOT NULL,
    
    INDEX (entity_code, operation_timestamp),
    INDEX (operated_by_id, operation_timestamp),
    INDEX (tenant_id, operation_timestamp)
);

-- 创建26个专用索引用于性能优化
CREATE INDEX CONCURRENTLY idx_org_units_tenant_code ON organization_units (tenant_id, code);
CREATE INDEX CONCURRENTLY idx_org_units_parent_code ON organization_units (parent_code) WHERE parent_code IS NOT NULL;
CREATE INDEX CONCURRENTLY idx_org_units_status_current ON organization_units (status, is_current) WHERE NOT is_deleted;
CREATE INDEX CONCURRENTLY idx_org_units_level_path ON organization_units (level, code_path);
CREATE INDEX CONCURRENTLY idx_org_units_effective_date ON organization_units (effective_date, end_date);
CREATE INDEX CONCURRENTLY idx_org_units_is_current ON organization_units (is_current) WHERE is_current = true;
CREATE INDEX CONCURRENTLY idx_org_units_unit_type ON organization_units (unit_type, status);
CREATE INDEX CONCURRENTLY idx_org_units_created_at ON organization_units (created_at);
CREATE INDEX CONCURRENTLY idx_org_units_updated_at ON organization_units (updated_at);
CREATE INDEX CONCURRENTLY idx_org_units_operation_type ON organization_units (operation_type, updated_at);

-- GIN索引用于JSONB和文本搜索
CREATE INDEX CONCURRENTLY idx_org_units_profile_gin ON organization_units USING gin (profile);
CREATE INDEX CONCURRENTLY idx_org_units_name_text ON organization_units USING gin (to_tsvector('english', name));
CREATE INDEX CONCURRENTLY idx_org_units_description_text ON organization_units USING gin (to_tsvector('english', description));

-- 层级查询专用索引
CREATE INDEX CONCURRENTLY idx_org_units_hierarchy ON organization_units (tenant_id, level, parent_code, code_path);
CREATE INDEX CONCURRENTLY idx_org_units_subtree ON organization_units (tenant_id, code_path) WHERE NOT is_deleted;

-- 时态查询专用索引
CREATE INDEX CONCURRENTLY idx_org_units_temporal ON organization_units (code, effective_date, end_date, is_current);
-- 历史表已移除：所有时态版本统一存储在 organization_units 单表中，
-- 通过 effective_date/end_date + asOf 查询获取“当前/历史/计划”。

-- 审计和统计专用索引
CREATE INDEX CONCURRENTLY idx_audit_logs_entity_time ON audit_logs (entity_code, operation_timestamp);
CREATE INDEX CONCURRENTLY idx_audit_logs_operator_time ON audit_logs (operated_by_id, operation_timestamp);
CREATE INDEX CONCURRENTLY idx_audit_logs_tenant_time ON audit_logs (tenant_id, operation_timestamp);

-- 复合索引用于复杂查询
CREATE INDEX CONCURRENTLY idx_org_units_complex_1 ON organization_units (tenant_id, status, is_current, level) WHERE NOT is_deleted;
CREATE INDEX CONCURRENTLY idx_org_units_complex_2 ON organization_units (tenant_id, unit_type, parent_code, sort_order) WHERE NOT is_deleted;
CREATE INDEX CONCURRENTLY idx_org_units_complex_3 ON organization_units (tenant_id, effective_date, end_date, is_current) WHERE NOT is_deleted;

-- 性能监控索引
CREATE INDEX CONCURRENTLY idx_org_units_monitoring ON organization_units (tenant_id, created_at, operation_type);

-- 加速父节点有效性与当前未删除节点查找
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_units_code_current_active 
    ON organization_units (code) WHERE is_current = true AND is_deleted = false;

-- 更新时间戳触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organization_units_updated_at
    BEFORE UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 层级路径自动更新触发器
CREATE OR REPLACE FUNCTION update_hierarchy_paths()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新code_path和name_path
    IF NEW.parent_code IS NULL THEN
        NEW.code_path := '/' || NEW.code;
        NEW.name_path := '/' || NEW.name;
        NEW.level := 1;
    ELSE
        -- 仅允许未删除且当前的父节点参与层级计算
        SELECT 
            parent.code_path || '/' || NEW.code,
            parent.name_path || '/' || NEW.name,
            parent.level + 1
        INTO NEW.code_path, NEW.name_path, NEW.level
        FROM organization_units parent
        WHERE parent.code = NEW.parent_code 
          AND parent.is_current = true
          AND parent.is_deleted = false
        LIMIT 1;

        -- 未找到有效父节点时，降级为根节点（防御性处理）
        IF NOT FOUND THEN
            NEW.parent_code := NULL;
            NEW.code_path := '/' || NEW.code;
            NEW.name_path := '/' || NEW.name;
            NEW.level := 1;
        END IF;
    END IF;
    
    NEW.hierarchy_depth := NEW.level;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_hierarchy_paths_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION update_hierarchy_paths();

-- 历史归档触发器已移除：改用单表时态与审计日志/时间线事件追踪变更。

-- 父节点有效性校验触发器（仅在插入时阻止引用已删除或非当前的父节点）
CREATE OR REPLACE FUNCTION validate_parent_available()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        PERFORM 1 FROM organization_units p
         WHERE p.code = NEW.parent_code
           AND p.is_current = true
           AND p.is_deleted = false
         LIMIT 1;
        IF NOT FOUND THEN
            RAISE EXCEPTION 'PARENT_NOT_AVAILABLE: parent % is not current or has been deleted', NEW.parent_code
                USING ERRCODE = 'foreign_key_violation';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS validate_parent_available_trigger ON organization_units;
CREATE TRIGGER validate_parent_available_trigger
    BEFORE INSERT ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION validate_parent_available();

-- 时态标志自动规范化（与软删除联动）
CREATE OR REPLACE FUNCTION enforce_temporal_flags()
RETURNS TRIGGER AS $$
BEGIN
    -- 软删除必须非当前且非未来
    IF NEW.is_deleted IS TRUE THEN
        NEW.is_current := FALSE;
        NEW.is_future := FALSE;
        RETURN NEW;
    END IF;

    -- 根据effective/end日期自动推导is_current/is_future
    IF NEW.effective_date > CURRENT_DATE THEN
        NEW.is_current := FALSE;
        NEW.is_future := TRUE;
    ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= CURRENT_DATE THEN
        NEW.is_current := FALSE;
        NEW.is_future := FALSE;
    ELSE
        NEW.is_current := TRUE;
        NEW.is_future := FALSE;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS enforce_temporal_flags_trigger ON organization_units;
CREATE TRIGGER enforce_temporal_flags_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION enforce_temporal_flags();

-- 审计日志自动记录触发器
CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    operation_type_val VARCHAR(20);
    changes_summary_val TEXT;
BEGIN
    -- 确定操作类型
    IF TG_OP = 'INSERT' THEN
        operation_type_val := NEW.operation_type;
    ELSIF TG_OP = 'UPDATE' THEN
        operation_type_val := NEW.operation_type;
    ELSIF TG_OP = 'DELETE' THEN
        operation_type_val := 'DELETE';
    END IF;
    
    -- 生成变更摘要
    IF TG_OP = 'INSERT' THEN
        changes_summary_val := 'Created organization unit: ' || NEW.name;
    ELSIF TG_OP = 'UPDATE' THEN
        changes_summary_val := 'Updated organization unit: ' || NEW.name;
    ELSIF TG_OP = 'DELETE' THEN
        changes_summary_val := 'Deleted organization unit: ' || OLD.name;
    END IF;
    
    -- 记录审计日志
    INSERT INTO audit_logs (
        entity_code,
        operation_type,
        operated_by_id,
        operated_by_name,
        operation_reason,
        before_data,
        after_data,
        changes_summary,
        tenant_id
    ) VALUES (
        COALESCE(NEW.code, OLD.code),
        operation_type_val,
        COALESCE(NEW.operated_by_id, OLD.operated_by_id),
        COALESCE(NEW.operated_by_name, OLD.operated_by_name),
        COALESCE(NEW.operation_reason, OLD.operation_reason),
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE row_to_json(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE row_to_json(NEW) END,
        changes_summary_val,
        COALESCE(NEW.tenant_id, OLD.tenant_id)
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();
