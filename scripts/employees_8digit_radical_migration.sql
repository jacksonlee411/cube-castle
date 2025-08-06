-- 员工管理系统8位编码彻底迁移脚本
-- 版本: v1.0 Radical Migration
-- 创建日期: 2025-08-05
-- 策略: 彻底重构，不考虑向后兼容

-- ============================================
-- 清理现有结构
-- ============================================

BEGIN;

-- 删除现有的员工职位关联表
DROP TABLE IF EXISTS employee_positions CASCADE;

-- 删除现有的员工表  
DROP TABLE IF EXISTS employees CASCADE;

-- 删除相关序列
DROP SEQUENCE IF EXISTS employee_code_seq CASCADE;
DROP SEQUENCE IF EXISTS employee_positions_id_seq CASCADE;

-- ============================================
-- 创建8位编码员工表
-- ============================================

-- 员工编码序列 (10000000-99999999)
CREATE SEQUENCE employee_code_seq 
    START WITH 10000000 
    INCREMENT BY 1 
    MAXVALUE 99999999 
    NO CYCLE;

-- 核心员工表 - 8位编码主键
CREATE TABLE employees (
    -- 8位编码主键
    code VARCHAR(8) PRIMARY KEY CHECK (
        code ~ '^[0-9]{8}$' AND 
        code::INTEGER BETWEEN 10000000 AND 99999999
    ),
    
    -- 直接关联关系 (零转换)
    organization_code VARCHAR(7) NOT NULL,     -- 直接关联组织
    primary_position_code VARCHAR(7),          -- 主要职位
    
    -- 员工类型和状态
    employee_type VARCHAR(20) NOT NULL CHECK (
        employee_type IN ('FULL_TIME', 'PART_TIME', 'CONTRACTOR', 'INTERN')
    ),
    employment_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (
        employment_status IN ('ACTIVE', 'TERMINATED', 'ON_LEAVE', 'PENDING_START')
    ),
    
    -- 基本个人信息
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    personal_email VARCHAR(255),
    phone_number VARCHAR(20),
    
    -- 入职和离职信息
    hire_date DATE NOT NULL,
    termination_date DATE,
    
    -- 扩展信息 (JSON格式)
    personal_info JSONB,           -- 个人详细信息
    employee_details JSONB,        -- 员工工作详情
    
    -- 系统字段
    tenant_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束条件
    UNIQUE(email, tenant_id),
    FOREIGN KEY (organization_code) REFERENCES organization_units(code) ON DELETE RESTRICT,
    FOREIGN KEY (primary_position_code) REFERENCES positions(code) ON DELETE SET NULL
);

-- 8位编码自动生成触发器
CREATE OR REPLACE FUNCTION generate_employee_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('employee_code_seq')::TEXT, 8, '0');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER employee_code_trigger
    BEFORE INSERT ON employees
    FOR EACH ROW
    EXECUTE FUNCTION generate_employee_code();

-- 更新时间戳触发器
CREATE OR REPLACE FUNCTION update_employee_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER employee_updated_at_trigger
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION update_employee_updated_at();

-- ============================================
-- 员工职位关联表 (支持多职位)
-- ============================================

CREATE TABLE employee_positions (
    id SERIAL PRIMARY KEY,
    
    -- 8位员工编码 + 7位职位编码
    employee_code VARCHAR(8) NOT NULL,
    position_code VARCHAR(7) NOT NULL,
    
    -- 分配类型和状态
    assignment_type VARCHAR(20) NOT NULL DEFAULT 'PRIMARY' CHECK (
        assignment_type IN ('PRIMARY', 'SECONDARY', 'TEMPORARY', 'ACTING')
    ),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (
        status IN ('ACTIVE', 'INACTIVE', 'PENDING', 'ENDED')
    ),
    
    -- 任职时间
    start_date DATE NOT NULL,
    end_date DATE,
    
    -- 系统字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束条件
    UNIQUE(employee_code, position_code, assignment_type, start_date),
    CHECK (end_date IS NULL OR end_date >= start_date),
    
    -- 外键约束
    FOREIGN KEY (employee_code) REFERENCES employees(code) ON DELETE CASCADE,
    FOREIGN KEY (position_code) REFERENCES positions(code) ON DELETE CASCADE
);

-- 员工职位关联更新触发器
CREATE TRIGGER employee_positions_updated_at_trigger
    BEFORE UPDATE ON employee_positions
    FOR EACH ROW
    EXECUTE FUNCTION update_employee_updated_at();

-- ============================================
-- 高性能索引策略
-- ============================================

-- 员工表核心索引
CREATE INDEX idx_employees_organization ON employees(organization_code);
CREATE INDEX idx_employees_primary_position ON employees(primary_position_code);
CREATE INDEX idx_employees_status ON employees(employment_status);
CREATE INDEX idx_employees_type ON employees(employee_type);
CREATE INDEX idx_employees_hire_date ON employees(hire_date);
CREATE INDEX idx_employees_email ON employees(email);
CREATE INDEX idx_employees_tenant ON employees(tenant_id);
CREATE INDEX idx_employees_name ON employees(first_name, last_name);

-- 复合索引优化关联查询
CREATE INDEX idx_employees_org_status ON employees(organization_code, employment_status);
CREATE INDEX idx_employees_type_status ON employees(employee_type, employment_status);
CREATE INDEX idx_employees_active ON employees(employment_status) WHERE employment_status = 'ACTIVE';

-- 员工职位关联表索引
CREATE INDEX idx_emp_pos_employee ON employee_positions(employee_code);
CREATE INDEX idx_emp_pos_position ON employee_positions(position_code);
CREATE INDEX idx_emp_pos_status ON employee_positions(status);
CREATE INDEX idx_emp_pos_assignment ON employee_positions(assignment_type);
CREATE INDEX idx_emp_pos_dates ON employee_positions(start_date, end_date);

-- 复合索引优化特定查询
CREATE INDEX idx_emp_pos_active ON employee_positions(employee_code, status) WHERE status = 'ACTIVE';
CREATE INDEX idx_emp_pos_primary ON employee_positions(employee_code, assignment_type) WHERE assignment_type = 'PRIMARY';

-- ============================================
-- 测试数据插入
-- ============================================

-- 插入测试员工数据
INSERT INTO employees (
    organization_code, primary_position_code, employee_type, employment_status,
    first_name, last_name, email, hire_date, tenant_id,
    personal_info, employee_details
) VALUES 
(
    '1000000', '1000001', 'FULL_TIME', 'ACTIVE',
    '张', '伟', 'zhang.wei@company.com', '2024-01-15', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '{"age": 30, "gender": "M", "address": "北京市朝阳区"}',
    '{"title": "高级软件工程师", "level": "P6", "salary": 28000}'
),
(
    '1000000', '1000002', 'FULL_TIME', 'ACTIVE', 
    '李', '娜', 'li.na@company.com', '2024-02-01', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '{"age": 28, "gender": "F", "address": "上海市浦东新区"}',
    '{"title": "软件架构师", "level": "P7", "salary": 35000}'
),
(
    '1000001', '1000003', 'FULL_TIME', 'ACTIVE',
    '王', '强', 'wang.qiang@company.com', '2024-03-10', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '{"age": 32, "gender": "M", "address": "深圳市南山区"}', 
    '{"title": "产品经理", "level": "P6", "salary": 30000}'
),
(
    '1000001', '1000004', 'PART_TIME', 'ACTIVE',
    '刘', '敏', 'liu.min@company.com', '2024-04-05', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '{"age": 26, "gender": "F", "address": "广州市天河区"}',
    '{"title": "UI设计师", "level": "P4", "hourly_rate": 200}'
),
(
    '1000002', '1000005', 'INTERN', 'ACTIVE',
    '陈', '阳', 'chen.yang@company.com', '2024-05-20', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '{"age": 22, "gender": "M", "address": "杭州市西湖区"}',
    '{"title": "前端开发实习生", "university": "浙江大学", "stipend": 3000}'
);

-- 插入员工职位关联关系
INSERT INTO employee_positions (employee_code, position_code, assignment_type, status, start_date) 
SELECT code, primary_position_code, 'PRIMARY', 'ACTIVE', hire_date 
FROM employees 
WHERE primary_position_code IS NOT NULL;

-- ============================================
-- 数据完整性验证
-- ============================================

-- 验证员工编码格式
DO $$
DECLARE
    invalid_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO invalid_count 
    FROM employees 
    WHERE code !~ '^[0-9]{8}$' OR code::INTEGER < 10000000 OR code::INTEGER > 99999999;
    
    IF invalid_count > 0 THEN
        RAISE EXCEPTION '发现 % 个无效的员工编码', invalid_count;
    END IF;
    
    RAISE NOTICE '✅ 员工编码格式验证通过';
END $$;

-- 验证关联关系完整性
DO $$
DECLARE
    orphan_employees INTEGER;
    orphan_positions INTEGER;
BEGIN
    -- 检查孤立的员工记录
    SELECT COUNT(*) INTO orphan_employees
    FROM employees e
    LEFT JOIN organization_units o ON e.organization_code = o.code
    WHERE o.code IS NULL;
    
    IF orphan_employees > 0 THEN
        RAISE EXCEPTION '发现 % 个员工没有有效的组织关联', orphan_employees;
    END IF;
    
    -- 检查孤立的职位关联
    SELECT COUNT(*) INTO orphan_positions
    FROM employee_positions ep
    LEFT JOIN employees e ON ep.employee_code = e.code
    LEFT JOIN positions p ON ep.position_code = p.code
    WHERE e.code IS NULL OR p.code IS NULL;
    
    IF orphan_positions > 0 THEN
        RAISE EXCEPTION '发现 % 个无效的员工职位关联', orphan_positions;
    END IF;
    
    RAISE NOTICE '✅ 关联关系完整性验证通过';
END $$;

-- 验证索引创建
DO $$
DECLARE
    index_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes 
    WHERE tablename IN ('employees', 'employee_positions');
    
    IF index_count < 15 THEN
        RAISE EXCEPTION '索引创建不完整，期望至少15个索引，实际 %', index_count;
    END IF;
    
    RAISE NOTICE '✅ 索引创建验证通过，共创建 % 个索引', index_count;
END $$;

-- 性能测试查询
DO $$
DECLARE
    start_time TIMESTAMP;
    end_time TIMESTAMP;
    duration INTERVAL;
BEGIN
    -- 测试8位编码直接查询性能
    start_time := clock_timestamp();
    PERFORM * FROM employees WHERE code = '10000001';
    end_time := clock_timestamp();
    duration := end_time - start_time;
    
    RAISE NOTICE '🚀 8位编码直接查询耗时: %', duration;
    
    -- 测试关联查询性能
    start_time := clock_timestamp();
    PERFORM e.code, e.first_name, e.last_name, o.name, p.details->>'title'
    FROM employees e
    LEFT JOIN organization_units o ON e.organization_code = o.code
    LEFT JOIN positions p ON e.primary_position_code = p.code
    WHERE e.code = '10000001';
    end_time := clock_timestamp();
    duration := end_time - start_time;
    
    RAISE NOTICE '🔗 关联查询耗时: %', duration;
END $$;

-- 最终统计信息
SELECT 
    '员工总数' as metric, COUNT(*) as value FROM employees
UNION ALL
SELECT 
    '活跃员工', COUNT(*) FROM employees WHERE employment_status = 'ACTIVE'
UNION ALL  
SELECT 
    '全职员工', COUNT(*) FROM employees WHERE employee_type = 'FULL_TIME'
UNION ALL
SELECT 
    '职位关联', COUNT(*) FROM employee_positions WHERE status = 'ACTIVE';

COMMIT;

-- ============================================
-- 迁移完成提示
-- ============================================
\echo '🎉 员工管理系统8位编码迁移完成！'
\echo '📊 核心特性:'
\echo '   • 8位编码主键 (10000000-99999999)'
\echo '   • 零转换直接查询架构'  
\echo '   • 高性能B-tree索引'
\echo '   • 员工-职位-组织直接关联'
\echo '   • 自动编码生成机制'
\echo '   • 完整的约束和验证'
\echo ''
\echo '🚀 下一步: 开发Go API服务器'