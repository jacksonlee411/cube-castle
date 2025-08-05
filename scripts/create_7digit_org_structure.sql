-- 组织单元API彻底激进优化 - 空数据库重构脚本
-- 直接创建7位code主键系统
-- 版本: v1.0
-- 创建日期: 2025-08-05

BEGIN;

-- 1. 删除现有的organization_units表
DROP TABLE IF EXISTS organization_units CASCADE;

-- 2. 创建7位编码序列
DROP SEQUENCE IF EXISTS org_unit_code_seq CASCADE;
CREATE SEQUENCE org_unit_code_seq 
    START WITH 1000000 
    INCREMENT BY 1 
    MAXVALUE 9999999;

-- 3. 创建新的组织单元表结构
CREATE TABLE organization_units (
    code VARCHAR(10) PRIMARY KEY,              -- 7位编码直接作为主键
    parent_code VARCHAR(10),                   -- 父级7位编码
    tenant_id UUID NOT NULL,                  -- 租户隔离
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM')),
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'PLANNED')),
    level INTEGER NOT NULL DEFAULT 1,
    path VARCHAR(1000),                       -- 层级路径: /1000000/1000001/1000002
    sort_order INTEGER DEFAULT 0,            -- 同级排序
    description TEXT,
    profile JSONB DEFAULT '{}',               -- 多态配置
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. 添加外键约束
ALTER TABLE organization_units 
    ADD CONSTRAINT fk_parent_code 
    FOREIGN KEY (parent_code) 
    REFERENCES organization_units(code)
    ON DELETE CASCADE;

-- 5. 添加唯一约束
ALTER TABLE organization_units 
    ADD CONSTRAINT uk_tenant_code 
    UNIQUE (tenant_id, code);

ALTER TABLE organization_units 
    ADD CONSTRAINT uk_tenant_name 
    UNIQUE (tenant_id, name);

-- 6. 创建高性能索引
CREATE INDEX idx_org_units_parent_code ON organization_units(parent_code);
CREATE INDEX idx_org_units_tenant_status ON organization_units(tenant_id, status);
CREATE INDEX idx_org_units_type_level ON organization_units(unit_type, level);

-- 安装pg_trgm扩展（如果还没有）
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 创建全文搜索索引
CREATE INDEX idx_org_units_path_gin ON organization_units USING gin(path gin_trgm_ops);
CREATE INDEX idx_org_units_name_gin ON organization_units USING gin(name gin_trgm_ops);

-- 7. 创建自动生成7位编码的触发器函数
CREATE OR REPLACE FUNCTION generate_org_unit_code()
RETURNS TRIGGER AS $$
BEGIN
    -- 自动生成7位编码（如果为空）
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_unit_code_seq')::text, 7, '0');
    END IF;
    
    -- 自动计算层级和路径
    IF NEW.parent_code IS NOT NULL THEN
        SELECT level + 1, path || '/' || NEW.code 
        INTO NEW.level, NEW.path
        FROM organization_units 
        WHERE code = NEW.parent_code;
    ELSE
        NEW.level := 1;
        NEW.path := '/' || NEW.code;
    END IF;
    
    -- 更新时间戳
    NEW.updated_at := NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 8. 创建触发器
CREATE TRIGGER set_org_unit_code 
    BEFORE INSERT OR UPDATE ON organization_units 
    FOR EACH ROW EXECUTE FUNCTION generate_org_unit_code();

-- 9. 创建更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 10. 创建更新时间戳触发器
CREATE TRIGGER update_organization_units_updated_at
    BEFORE UPDATE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 11. 创建变更通知触发器
CREATE OR REPLACE FUNCTION notify_organization_change() 
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'DELETE' then
        PERFORM pg_notify('organization_change', json_build_object('operation', TG_OP, 'code', OLD.code)::text);
        RETURN OLD;
    ELSE
        PERFORM pg_notify('organization_change', json_build_object('operation', TG_OP, 'code', NEW.code)::text);
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organization_units_change_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW EXECUTE FUNCTION notify_organization_change();

-- 12. 更新employees表结构以使用code
-- 首先删除旧的外键约束
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_id_fkey;
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_code_fkey;

-- 删除旧列，添加新列
ALTER TABLE employees DROP COLUMN IF EXISTS department_id;
ALTER TABLE employees ADD COLUMN IF NOT EXISTS department_code VARCHAR(10);

-- 添加新的外键约束
ALTER TABLE employees 
    ADD CONSTRAINT employees_department_code_fkey 
    FOREIGN KEY (department_code) 
    REFERENCES organization_units(code)
    ON DELETE SET NULL;

-- 创建employees表的索引
CREATE INDEX IF NOT EXISTS idx_employees_department_code ON employees(department_code);

-- 13. 插入测试数据
INSERT INTO organization_units (tenant_id, name, unit_type, description, profile) VALUES
(gen_random_uuid(), '高谷集团', 'COMPANY', '集团总公司', '{"type": "headquarters"}'),
(gen_random_uuid(), '技术部', 'DEPARTMENT', '技术研发部门', '{"type": "rd"}'),
(gen_random_uuid(), '产品部', 'DEPARTMENT', '产品管理部门', '{"type": "product"}'),
(gen_random_uuid(), '销售部', 'DEPARTMENT', '销售业务部门', '{"type": "sales"}'),
(gen_random_uuid(), '人事部', 'DEPARTMENT', '人力资源部门', '{"type": "hr"}');

-- 更新子部门的parent_code
UPDATE organization_units 
SET parent_code = (SELECT code FROM organization_units WHERE name = '高谷集团' LIMIT 1)
WHERE name IN ('技术部', '产品部', '销售部', '人事部');

COMMIT;

-- 验证数据完整性
SELECT 
    COUNT(*) as total_units,
    COUNT(DISTINCT code) as unique_codes,
    MIN(LENGTH(code)) as min_code_len,
    MAX(LENGTH(code)) as max_code_len,
    COUNT(CASE WHEN parent_code IS NULL THEN 1 END) as root_units,
    MAX(level) as max_level
FROM organization_units;

-- 显示迁移结果样本
SELECT code, parent_code, name, unit_type, level, path 
FROM organization_units 
ORDER BY path 
LIMIT 10;