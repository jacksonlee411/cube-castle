-- 职位管理7位编码彻底重构迁移脚本 v1.1 (修正版)
-- 创建日期: 2025-08-05
-- 策略: 彻底重构，清空现有数据，全新7位编码架构
-- 修正: 解决外键约束和数据插入问题

-- ===========================================
-- 第1步: 安全备份现有数据
-- ===========================================

DROP TABLE IF EXISTS positions_backup_20250805;
CREATE TABLE positions_backup_20250805 AS SELECT * FROM positions;

DROP TABLE IF EXISTS position_assignments_backup_20250805;
CREATE TABLE position_assignments_backup_20250805 AS SELECT * FROM position_assignments;

DROP TABLE IF EXISTS employee_positions_backup_20250805;
CREATE TABLE employee_positions_backup_20250805 AS SELECT * FROM employee_positions;

-- ===========================================
-- 第2步: 彻底清理现有结构
-- ===========================================

-- 删除所有相关表和约束
DROP TABLE IF EXISTS position_assignments CASCADE;
DROP TABLE IF EXISTS employee_positions CASCADE;
DROP TABLE IF EXISTS positions CASCADE;

-- ===========================================
-- 第3步: 创建全新7位编码职位表
-- ===========================================

CREATE TABLE positions (
    code VARCHAR(7) PRIMARY KEY CHECK (code ~ '^[0-9]{7}$' AND code::INTEGER BETWEEN 1000000 AND 9999999),
    organization_code VARCHAR(7) NOT NULL,
    manager_position_code VARCHAR(7),
    position_type VARCHAR(50) NOT NULL CHECK (position_type IN 
        ('FULL_TIME', 'PART_TIME', 'CONTINGENT_WORKER', 'INTERN')),
    job_profile_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN 
        ('OPEN', 'FILLED', 'FROZEN', 'PENDING_ELIMINATION')),
    budgeted_fte NUMERIC(3,2) NOT NULL DEFAULT 1.00 CHECK (budgeted_fte > 0 AND budgeted_fte <= 5.00),
    details JSONB,
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 外键约束 (延迟添加自引用约束)
ALTER TABLE positions 
ADD CONSTRAINT fk_positions_organization 
FOREIGN KEY (organization_code) REFERENCES organization_units(code);

-- ===========================================
-- 第4步: 创建高性能索引
-- ===========================================

CREATE INDEX idx_positions_organization ON positions(organization_code);
CREATE INDEX idx_positions_manager ON positions(manager_position_code);
CREATE INDEX idx_positions_status ON positions(status);
CREATE INDEX idx_positions_type ON positions(position_type);
CREATE INDEX idx_positions_tenant ON positions(tenant_id);
CREATE INDEX idx_positions_updated ON positions(updated_at);
CREATE INDEX idx_positions_job_profile ON positions(job_profile_id);

-- 复合索引
CREATE INDEX idx_positions_org_status ON positions(organization_code, status);
CREATE INDEX idx_positions_type_status ON positions(position_type, status);
CREATE INDEX idx_positions_tenant_org ON positions(tenant_id, organization_code);

-- ===========================================
-- 第5步: 编码生成系统
-- ===========================================

CREATE TABLE position_code_sequence (
    tenant_id UUID PRIMARY KEY,
    last_code INTEGER NOT NULL DEFAULT 1000000,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE OR REPLACE FUNCTION generate_position_code(p_tenant_id UUID) 
RETURNS VARCHAR(7) AS $$
DECLARE
    current_code INTEGER;
    new_code VARCHAR(7);
BEGIN
    INSERT INTO position_code_sequence (tenant_id, last_code)
    VALUES (p_tenant_id, 1000000)
    ON CONFLICT (tenant_id) DO NOTHING;
    
    UPDATE position_code_sequence 
    SET last_code = last_code + 1,
        updated_at = CURRENT_TIMESTAMP
    WHERE tenant_id = p_tenant_id
    RETURNING last_code INTO current_code;
    
    IF current_code > 9999999 THEN
        RAISE EXCEPTION 'Position code overflow for tenant %', p_tenant_id;
    END IF;
    
    new_code := LPAD(current_code::TEXT, 7, '0');
    RETURN new_code;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION auto_generate_position_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := generate_position_code(NEW.tenant_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_auto_position_code
    BEFORE INSERT ON positions
    FOR EACH ROW
    EXECUTE FUNCTION auto_generate_position_code();

CREATE OR REPLACE FUNCTION update_position_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_position_timestamp
    BEFORE UPDATE ON positions
    FOR EACH ROW
    EXECUTE FUNCTION update_position_updated_at();

-- ===========================================
-- 第6步: 创建关联表
-- ===========================================

CREATE TABLE employee_positions (
    id SERIAL PRIMARY KEY,
    employee_code VARCHAR(8) NOT NULL,
    position_code VARCHAR(7) NOT NULL,
    assignment_type VARCHAR(20) NOT NULL DEFAULT 'PRIMARY' CHECK (assignment_type IN 
        ('PRIMARY', 'SECONDARY', 'TEMPORARY', 'ACTING')),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN 
        ('ACTIVE', 'INACTIVE', 'PENDING', 'ENDED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_code, position_code, assignment_type, start_date)
);

ALTER TABLE employee_positions 
ADD CONSTRAINT fk_employee_positions_position 
FOREIGN KEY (position_code) REFERENCES positions(code);

CREATE TABLE position_assignments (
    id SERIAL PRIMARY KEY,
    position_code VARCHAR(7) NOT NULL,
    employee_code VARCHAR(8),
    assignment_date DATE NOT NULL,
    end_date DATE,
    assignment_reason VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_position_assignments_position 
        FOREIGN KEY (position_code) REFERENCES positions(code) ON DELETE CASCADE
);

-- 关联表索引
CREATE INDEX idx_employee_positions_employee ON employee_positions(employee_code);
CREATE INDEX idx_employee_positions_position ON employee_positions(position_code);
CREATE INDEX idx_employee_positions_status ON employee_positions(status);
CREATE INDEX idx_position_assignments_position ON position_assignments(position_code);
CREATE INDEX idx_position_assignments_employee ON position_assignments(employee_code);

-- ===========================================
-- 第7步: 插入测试数据 (分步插入)
-- ===========================================

-- 插入基础职位数据 (不包含管理关系)
INSERT INTO positions (organization_code, position_type, job_profile_id, status, budgeted_fte, details, tenant_id) VALUES
('1000000', 'FULL_TIME', '550e8400-e29b-41d4-a716-446655440000', 'OPEN', 1.0, 
 '{"title": "高级软件工程师", "salary_range": {"min": 20000, "max": 35000, "currency": "CNY"}}', 
 '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'),
('1000000', 'FULL_TIME', '550e8400-e29b-41d4-a716-446655440000', 'OPEN', 1.0, 
 '{"title": "软件架构师", "salary_range": {"min": 30000, "max": 50000, "currency": "CNY"}}', 
 '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'),
('1000001', 'FULL_TIME', '550e8400-e29b-41d4-a716-446655440000', 'FILLED', 1.0, 
 '{"title": "产品经理", "salary_range": {"min": 25000, "max": 40000, "currency": "CNY"}}', 
 '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'),
('1000001', 'PART_TIME', '550e8400-e29b-41d4-a716-446655440000', 'OPEN', 0.5, 
 '{"title": "UI设计师", "hourly_rate": 200, "max_hours_per_week": 20}', 
 '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'),
('1000002', 'INTERN', '550e8400-e29b-41d4-a716-446655440000', 'FILLED', 0.8, 
 '{"title": "前端开发实习生", "stipend": 3000, "internship_duration": "3m"}', 
 '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9');

-- 现在添加自引用外键约束
ALTER TABLE positions 
ADD CONSTRAINT fk_positions_manager 
FOREIGN KEY (manager_position_code) REFERENCES positions(code);

-- 设置管理关系 (使用实际生成的编码)
UPDATE positions SET manager_position_code = '1000000' WHERE code = '1000002';
UPDATE positions SET manager_position_code = '1000001' WHERE code = '1000003';
UPDATE positions SET manager_position_code = '1000000' WHERE code = '1000004';

-- ===========================================
-- 第8步: 验证和性能测试
-- ===========================================

-- 验证数据
SELECT 
    '🎉 职位管理7位编码重构完成' as status,
    COUNT(*) as total_positions,
    MIN(code) as min_code,
    MAX(code) as max_code,
    COUNT(DISTINCT organization_code) as organizations,
    COUNT(CASE WHEN manager_position_code IS NOT NULL THEN 1 END) as with_managers
FROM positions;

-- 性能测试
EXPLAIN ANALYZE SELECT * FROM positions WHERE code = '1000000';
EXPLAIN ANALYZE SELECT * FROM positions WHERE organization_code = '1000000';
EXPLAIN ANALYZE SELECT * FROM positions WHERE status = 'OPEN';

-- 显示新索引
SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'positions' ORDER BY indexname;