-- 043_create_positions_and_job_catalog.sql
-- 创建职位管理与职位分类(Job Catalog)核心数据表

BEGIN;

CREATE TABLE IF NOT EXISTS job_family_groups (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    family_group_code VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, family_group_code, effective_date),
    UNIQUE (record_id, tenant_id)
);

CREATE TABLE IF NOT EXISTS job_families (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    family_code VARCHAR(20) NOT NULL,
    family_group_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, family_code, effective_date),
    UNIQUE (record_id, tenant_id),
    CONSTRAINT fk_job_families_group FOREIGN KEY (parent_record_id, tenant_id)
        REFERENCES job_family_groups(record_id, tenant_id)
);

CREATE TABLE IF NOT EXISTS job_roles (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    role_code VARCHAR(20) NOT NULL,
    family_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    competency_model JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, role_code, effective_date),
    UNIQUE (record_id, tenant_id),
    CONSTRAINT fk_job_roles_family FOREIGN KEY (parent_record_id, tenant_id)
        REFERENCES job_families(record_id, tenant_id)
);

CREATE TABLE IF NOT EXISTS job_levels (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    level_code VARCHAR(20) NOT NULL,
    role_code VARCHAR(20) NOT NULL,
    parent_record_id UUID NOT NULL,
    level_rank VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    salary_band JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant_id, level_code, effective_date),
    UNIQUE (record_id, tenant_id),
    CONSTRAINT fk_job_levels_role FOREIGN KEY (parent_record_id, tenant_id)
        REFERENCES job_roles(record_id, tenant_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_job_family_groups_record
    ON job_family_groups(record_id, tenant_id, family_group_code);
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_families_record
    ON job_families(record_id, tenant_id, family_code);
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_roles_record
    ON job_roles(record_id, tenant_id, role_code);
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_levels_record
    ON job_levels(record_id, tenant_id, level_code);

CREATE UNIQUE INDEX IF NOT EXISTS uk_job_family_groups_current
    ON job_family_groups(tenant_id, family_group_code)
    WHERE is_current = true;
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_families_current
    ON job_families(tenant_id, family_code)
    WHERE is_current = true;
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_roles_current
    ON job_roles(tenant_id, role_code)
    WHERE is_current = true;
CREATE UNIQUE INDEX IF NOT EXISTS uk_job_levels_current
    ON job_levels(tenant_id, level_code)
    WHERE is_current = true;

CREATE TABLE IF NOT EXISTS positions (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    code VARCHAR(8) NOT NULL,
    title VARCHAR(120) NOT NULL,
    job_profile_code VARCHAR(64),
    job_profile_name VARCHAR(255),
    job_family_group_code VARCHAR(20) NOT NULL,
    job_family_group_name VARCHAR(255) NOT NULL,
    job_family_group_record_id UUID NOT NULL,
    job_family_code VARCHAR(20) NOT NULL,
    job_family_name VARCHAR(255) NOT NULL,
    job_family_record_id UUID NOT NULL,
    job_role_code VARCHAR(20) NOT NULL,
    job_role_name VARCHAR(255) NOT NULL,
    job_role_record_id UUID NOT NULL,
    job_level_code VARCHAR(20) NOT NULL,
    job_level_name VARCHAR(255) NOT NULL,
    job_level_record_id UUID NOT NULL,
    organization_code VARCHAR(7) NOT NULL,
    organization_name VARCHAR(255),
    position_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PLANNED',
    employment_type VARCHAR(50) NOT NULL,
    headcount_capacity NUMERIC(5,2) NOT NULL DEFAULT 1.0,
    headcount_in_use NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    grade_level VARCHAR(20),
    cost_center_code VARCHAR(50),
    current_holder_id UUID,
    current_holder_name VARCHAR(255),
    filled_date DATE,
    current_assignment_type VARCHAR(20),
    reports_to_position_code VARCHAR(8),
    profile JSONB NOT NULL DEFAULT '{}'::jsonb,
    effective_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    operation_type VARCHAR(20) NOT NULL DEFAULT 'CREATE',
    operated_by_id UUID NOT NULL,
    operated_by_name VARCHAR(255) NOT NULL,
    operation_reason TEXT,
    UNIQUE (tenant_id, code, effective_date),
    UNIQUE (tenant_id, code, record_id),
    CONSTRAINT fk_positions_family_group FOREIGN KEY (job_family_group_record_id, tenant_id)
        REFERENCES job_family_groups(record_id, tenant_id),
    CONSTRAINT fk_positions_family FOREIGN KEY (job_family_record_id, tenant_id)
        REFERENCES job_families(record_id, tenant_id),
    CONSTRAINT fk_positions_role FOREIGN KEY (job_role_record_id, tenant_id)
        REFERENCES job_roles(record_id, tenant_id),
    CONSTRAINT fk_positions_level FOREIGN KEY (job_level_record_id, tenant_id)
        REFERENCES job_levels(record_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_positions_org_code
    ON positions(tenant_id, organization_code, is_current);
CREATE INDEX IF NOT EXISTS idx_positions_current
    ON positions(tenant_id)
    WHERE is_current = true;
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'positions'
          AND column_name = 'current_holder_id'
    ) THEN
        EXECUTE 'CREATE INDEX IF NOT EXISTS idx_positions_holder ON positions(tenant_id, current_holder_id) WHERE current_holder_id IS NOT NULL';
    ELSE
        RAISE NOTICE 'idx_positions_holder skipped: current_holder_id column missing';
    END IF;
END
$$;
CREATE INDEX IF NOT EXISTS idx_positions_effective_date
    ON positions(tenant_id, effective_date);
CREATE INDEX IF NOT EXISTS idx_positions_status
    ON positions(tenant_id, status, is_current);
CREATE INDEX IF NOT EXISTS idx_positions_job_family_group
    ON positions(tenant_id, job_family_group_code, is_current);
CREATE INDEX IF NOT EXISTS idx_positions_job_family
    ON positions(tenant_id, job_family_code, is_current);
CREATE INDEX IF NOT EXISTS idx_positions_job_role
    ON positions(tenant_id, job_role_code, is_current);
CREATE UNIQUE INDEX IF NOT EXISTS uk_positions_current_active
    ON positions(tenant_id, code)
    WHERE is_current = true AND status <> 'DELETED';

COMMIT;
