-- 组织单元API彻底激进优化 - 数据库迁移脚本
-- 从UUID+6位business_id 迁移到 7位code主键系统
-- 版本: v1.0
-- 创建日期: 2025-08-05

BEGIN;

-- 1. 备份当前数据到临时表
CREATE TABLE organization_units_backup AS 
SELECT * FROM organization_units;

-- 2. 创建新的7位编码序列
CREATE SEQUENCE IF NOT EXISTS org_unit_code_seq 
    START WITH 1000000 
    INCREMENT BY 1 
    MAXVALUE 9999999;

-- 3. 创建新表结构
CREATE TABLE organization_units_new (
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. 创建临时映射表 (UUID -> 7位编码)
CREATE TEMP TABLE id_mapping AS
SELECT 
    id as old_id,
    LPAD((ROW_NUMBER() OVER (ORDER BY created_at) + 999999)::text, 7, '0') as new_code
FROM organization_units;

-- 5. 迁移数据到新表结构
INSERT INTO organization_units_new (
    code, parent_code, tenant_id, name, unit_type, status, 
    level, path, sort_order, description, profile, created_at, updated_at
)
SELECT 
    m.new_code,
    pm.new_code as parent_code,
    o.tenant_id,
    o.name,
    o.unit_type,
    o.status,
    COALESCE(o.level, 1) as level,
    ('/' || m.new_code) as path,  -- 临时路径，触发器会重新计算
    0 as sort_order,
    o.description,
    COALESCE(o.profile, '{}') as profile,
    o.created_at,
    o.updated_at
FROM organization_units o
JOIN id_mapping m ON o.id = m.old_id
LEFT JOIN id_mapping pm ON o.parent_unit_id = pm.old_id;

-- 6. 创建自动生成7位编码的触发器函数
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
        FROM organization_units_new 
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

-- 7. 创建触发器
CREATE TRIGGER set_org_unit_code 
    BEFORE INSERT OR UPDATE ON organization_units_new 
    FOR EACH ROW EXECUTE FUNCTION generate_org_unit_code();

-- 8. 更新路径信息（使用递归CTE）
WITH RECURSIVE org_tree AS (
    -- 根节点
    SELECT code, parent_code, 1 as level, ('/' || code) as path
    FROM organization_units_new 
    WHERE parent_code IS NULL
    
    UNION ALL
    
    -- 子节点
    SELECT o.code, o.parent_code, t.level + 1, t.path || '/' || o.code
    FROM organization_units_new o
    INNER JOIN org_tree t ON o.parent_code = t.code
)
UPDATE organization_units_new 
SET level = org_tree.level, path = org_tree.path
FROM org_tree 
WHERE organization_units_new.code = org_tree.code;

-- 9. 添加约束和外键
ALTER TABLE organization_units_new 
    ADD CONSTRAINT fk_parent_code 
    FOREIGN KEY (parent_code) 
    REFERENCES organization_units_new(code);

ALTER TABLE organization_units_new 
    ADD CONSTRAINT uk_tenant_code 
    UNIQUE (tenant_id, code);

-- 10. 创建高性能索引
CREATE INDEX idx_org_units_parent_code_new ON organization_units_new(parent_code);
CREATE INDEX idx_org_units_tenant_status_new ON organization_units_new(tenant_id, status);
CREATE INDEX idx_org_units_type_level_new ON organization_units_new(unit_type, level);
CREATE INDEX idx_org_units_path_gin_new ON organization_units_new USING gin(path gin_trgm_ops);
CREATE INDEX idx_org_units_name_gin_new ON organization_units_new USING gin(name gin_trgm_ops);

-- 11. 创建更新时间戳触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organization_units_updated_at
    BEFORE UPDATE ON organization_units_new
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 12. 更新employees表的外键引用
-- 首先添加新的外键列
ALTER TABLE employees ADD COLUMN department_code VARCHAR(10);

-- 使用映射表更新employees的department_code
UPDATE employees 
SET department_code = m.new_code
FROM id_mapping m
WHERE employees.department_id = m.old_id;

-- 13. 原子替换表
-- 删除旧表的外键约束
ALTER TABLE employees DROP CONSTRAINT IF EXISTS employees_department_id_fkey;

-- 删除旧表
DROP TABLE organization_units;

-- 重命名新表
ALTER TABLE organization_units_new RENAME TO organization_units;

-- 重命名索引
ALTER INDEX idx_org_units_parent_code_new RENAME TO idx_org_units_parent_code;
ALTER INDEX idx_org_units_tenant_status_new RENAME TO idx_org_units_tenant_status;
ALTER INDEX idx_org_units_type_level_new RENAME TO idx_org_units_type_level;
ALTER INDEX idx_org_units_path_gin_new RENAME TO idx_org_units_path_gin;
ALTER INDEX idx_org_units_name_gin_new RENAME TO idx_org_units_name_gin;

-- 14. 更新employees表外键约束
ALTER TABLE employees 
    ADD CONSTRAINT employees_department_code_fkey 
    FOREIGN KEY (department_code) 
    REFERENCES organization_units(code);

-- 删除旧的department_id列
ALTER TABLE employees DROP COLUMN department_id;

-- 15. 创建变更通知触发器（如果需要）
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

-- 16. 更新序列值到当前最大值+1
SELECT setval('org_unit_code_seq', 
    (SELECT COALESCE(MAX(code::bigint), 999999) + 1 FROM organization_units));

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