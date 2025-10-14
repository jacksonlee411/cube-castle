-- =============================================
-- 时态管理数据库结构扩展
-- 支持组织架构的时态查询和版本管理
-- =============================================

-- 0. 如果缺失基础表结构，则创建最小初始化版本（兼容旧环境）
CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name = 'organization_units'
    ) THEN
        CREATE TABLE public.organization_units (
            record_id UUID NOT NULL DEFAULT gen_random_uuid(),
            tenant_id UUID NOT NULL,
            code VARCHAR(12) NOT NULL,
            parent_code VARCHAR(12),
            name VARCHAR(255) NOT NULL,
            unit_type VARCHAR(64) NOT NULL,
            status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
            level INTEGER NOT NULL DEFAULT 1,
            hierarchy_depth INTEGER NOT NULL DEFAULT 0,
            path TEXT NOT NULL DEFAULT '',
            code_path TEXT NOT NULL DEFAULT '',
            name_path TEXT NOT NULL DEFAULT '',
            sort_order INTEGER NOT NULL DEFAULT 0,
            description TEXT,
            profile JSONB NOT NULL DEFAULT '{}'::jsonb,
            metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
            end_date DATE,
            change_reason TEXT,
            is_current BOOLEAN NOT NULL DEFAULT false,
            deleted_at TIMESTAMPTZ,
            deleted_by UUID,
            deletion_reason TEXT,
            suspended_at TIMESTAMPTZ,
            suspended_by UUID,
            suspension_reason TEXT,
            operated_by_id UUID,
            operated_by_name TEXT,
            operation_type VARCHAR(20) DEFAULT 'CREATE',
            operation_reason TEXT,
            effective_from TIMESTAMPTZ,
            effective_to TIMESTAMPTZ,
            changed_by UUID,
            approved_by UUID,
            CONSTRAINT organization_units_pkey PRIMARY KEY (code, effective_date)
        );

        CREATE UNIQUE INDEX IF NOT EXISTS uk_org_ver 
            ON organization_units (tenant_id, code, effective_date);
        CREATE UNIQUE INDEX IF NOT EXISTS uk_org_current 
            ON organization_units (tenant_id, code) WHERE is_current = true;
        CREATE INDEX IF NOT EXISTS idx_org_units_tenant 
            ON organization_units (tenant_id);
        CREATE INDEX IF NOT EXISTS idx_org_units_parent 
            ON organization_units (parent_code);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'audit_logs'
    ) THEN
        CREATE TABLE public.audit_logs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            tenant_id UUID NOT NULL,
            event_type VARCHAR(50) NOT NULL,
            resource_type VARCHAR(50) NOT NULL,
            resource_id VARCHAR(100),
            actor_id VARCHAR(100),
            actor_type VARCHAR(50),
            action_name VARCHAR(100),
            request_id VARCHAR(100),
            operation_reason TEXT,
            "timestamp" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            success BOOLEAN NOT NULL DEFAULT TRUE,
            error_code VARCHAR(100),
            error_message TEXT,
            request_data JSONB NOT NULL DEFAULT '{}'::jsonb,
            response_data JSONB NOT NULL DEFAULT '{}'::jsonb,
            modified_fields JSONB NOT NULL DEFAULT '[]'::jsonb,
            changes JSONB NOT NULL DEFAULT '[]'::jsonb,
            record_id UUID,
            business_context JSONB NOT NULL DEFAULT '{}'::jsonb
        );

        CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp 
            ON audit_logs ("timestamp");
        CREATE INDEX IF NOT EXISTS idx_audit_logs_resource 
            ON audit_logs (resource_type, resource_id);
    END IF;
END
$$;

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

-- 创建完成提示
SELECT 
    'Temporal database schema setup completed successfully!' as status,
    COUNT(*) as total_tables
FROM information_schema.tables 
WHERE table_schema = 'public' 
  AND table_name LIKE '%organization%';
