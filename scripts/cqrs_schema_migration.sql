-- CQRS Schema Migration for Organization Module
-- Operation Phoenix Database Schema Update

-- 备份当前数据
CREATE TABLE IF NOT EXISTS employees_backup AS SELECT * FROM employees;
CREATE TABLE IF NOT EXISTS organization_units_backup AS SELECT * FROM organization_units;

-- 删除旧表的约束和索引
DROP TABLE IF EXISTS employees CASCADE;
DROP TABLE IF EXISTS organization_units CASCADE;
DROP TABLE IF EXISTS positions CASCADE;

-- 创建新的CQRS优化的员工表
CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_type VARCHAR(50) NOT NULL CHECK (employee_type IN ('FULL_TIME', 'PART_TIME', 'CONTRACTOR', 'INTERN')),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    position_id UUID,
    hire_date DATE NOT NULL,
    termination_date DATE,
    employment_status VARCHAR(50) NOT NULL DEFAULT 'PENDING_START' CHECK (employment_status IN ('PENDING_START', 'ACTIVE', 'TERMINATED', 'ON_LEAVE')),
    personal_info JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- 索引优化
    CONSTRAINT employees_email_tenant_unique UNIQUE (email, tenant_id)
);

-- 创建组织单元表
CREATE TABLE organization_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM')),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_unit_id UUID REFERENCES organization_units(id),
    profile JSONB,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- 防止循环引用
    CONSTRAINT organization_units_name_tenant_unique UNIQUE (name, tenant_id)
);

-- 创建职位表
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    title VARCHAR(100) NOT NULL,
    department VARCHAR(100) NOT NULL,
    level VARCHAR(50) NOT NULL,
    description TEXT,
    requirements TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT positions_title_dept_tenant_unique UNIQUE (title, department, tenant_id)
);

-- 创建员工职位关联表
CREATE TABLE employee_positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    position_id UUID NOT NULL REFERENCES positions(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT employee_positions_unique UNIQUE (employee_id, position_id, effective_date)
);

-- 创建索引提升查询性能
CREATE INDEX idx_employees_tenant_id ON employees(tenant_id);
CREATE INDEX idx_employees_email ON employees(email);
CREATE INDEX idx_employees_status ON employees(employment_status);
CREATE INDEX idx_employees_hire_date ON employees(hire_date);
CREATE INDEX idx_employees_position ON employees(position_id);

CREATE INDEX idx_org_units_tenant_id ON organization_units(tenant_id);
CREATE INDEX idx_org_units_parent ON organization_units(parent_unit_id);
CREATE INDEX idx_org_units_type ON organization_units(unit_type);
CREATE INDEX idx_org_units_active ON organization_units(is_active);

CREATE INDEX idx_positions_tenant_id ON positions(tenant_id);
CREATE INDEX idx_positions_dept ON positions(department);
CREATE INDEX idx_positions_active ON positions(is_active);

CREATE INDEX idx_emp_positions_employee ON employee_positions(employee_id);
CREATE INDEX idx_emp_positions_position ON employee_positions(position_id);
CREATE INDEX idx_emp_positions_tenant ON employee_positions(tenant_id);
CREATE INDEX idx_emp_positions_primary ON employee_positions(is_primary) WHERE is_primary = true;

-- 更新时间戳触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_organization_units_updated_at BEFORE UPDATE ON organization_units FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_positions_updated_at BEFORE UPDATE ON positions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 重新创建发布以支持新的表结构
DROP PUBLICATION IF EXISTS organization_publication;
CREATE PUBLICATION organization_publication FOR TABLE employees, organization_units, positions, employee_positions;

-- 插入测试数据
INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name, email, hire_date, employment_status)
VALUES (
    gen_random_uuid(),
    gen_random_uuid(),
    'FULL_TIME',
    'Phoenix',
    'TestEmployee',
    'phoenix.test@cubecastle.com',
    CURRENT_DATE,
    'ACTIVE'
);

INSERT INTO organization_units (id, tenant_id, unit_type, name, description, is_active)
VALUES (
    gen_random_uuid(),
    (SELECT tenant_id FROM employees WHERE first_name = 'Phoenix' LIMIT 1),
    'DEPARTMENT',
    'Operation Phoenix部门',
    'CQRS+CDC架构实施团队',
    true
);

-- 验证数据
SELECT 'CQRS Schema Migration完成' as status;
SELECT COUNT(*) as employee_count FROM employees;
SELECT COUNT(*) as org_unit_count FROM organization_units;
