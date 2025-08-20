-- ============================================================================
-- 数据清理和测试数据创建脚本
-- 功能：清除不符合规则的数据，创建标准测试数据
-- 版本：v2.0
-- 创建时间：2025-08-18
-- ============================================================================

BEGIN;

-- ============================================================================
-- 第一部分：数据清理 - 清除所有现有组织数据
-- ============================================================================

-- 1. 临时禁用触发器进行批量操作
ALTER TABLE organization_units DISABLE TRIGGER auto_end_date_trigger;
ALTER TABLE organization_units DISABLE TRIGGER auto_lifecycle_status_trigger;

-- 2. 备份当前数据（可选，保留重要组织代码）
CREATE TABLE IF NOT EXISTS organization_backup_20250818 AS 
SELECT DISTINCT code, name, unit_type FROM organization_units WHERE code IN ('1000000', '1000001', '1000002', '1000003', '1000004');

-- 3. 清理所有组织单元数据
TRUNCATE TABLE organization_units RESTART IDENTITY CASCADE;

-- ============================================================================
-- 第二部分：创建标准测试数据 - 规范的时态记录
-- ============================================================================

-- 4. 重新启用触发器
ALTER TABLE organization_units ENABLE TRIGGER auto_end_date_trigger;
ALTER TABLE organization_units ENABLE TRIGGER auto_lifecycle_status_trigger;

-- 插入测试数据 - 高谷集团总公司
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date, 
    lifecycle_status, business_status, data_status, 
    change_reason, level, path, sort_order, description
) VALUES (
    '1000000', 
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '高谷集团', 
    'COMPANY', 
    'ACTIVE',
    '2010-01-01',
    'CURRENT',
    'ACTIVE', 
    'NORMAL',
    '公司成立',
    1, 
    '/1000000', 
    0, 
    '高谷集团总部，负责整体战略规划和管理'
);

-- 插入测试数据 - 组织1000004的规范时态历史
-- 历史记录1：人事科 (2010-2013)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '人事科',
    'DEPARTMENT',
    'INACTIVE', 
    '2010-01-01',
    'HISTORICAL',
    'ACTIVE',
    'NORMAL',
    '公司成立初期的人事管理部门',
    2,
    '/1000000/1000004',
    0,
    '负责基础人事管理工作',
    '1000000'
);

-- 历史记录2：人力资源科 (2013-2015)  
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '人力资源科',
    'DEPARTMENT',
    'INACTIVE',
    '2013-01-01',
    'HISTORICAL', 
    'ACTIVE',
    'NORMAL',
    '业务扩展，人事管理职能升级',
    2,
    '/1000000/1000004',
    0,
    '人事管理职能扩展，增加招聘培训',
    '1000000'
);

-- 历史记录3：人力资源部 (2015-2020)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '人力资源部',
    'DEPARTMENT',
    'INACTIVE',
    '2015-01-01',
    'HISTORICAL',
    'ACTIVE', 
    'NORMAL',
    '组织架构调整，部门级别提升',
    2,
    '/1000000/1000004',
    0,
    '承担全面人力资源管理职责',
    '1000000'
);

-- 历史记录4：战略人力资源部 (2020-2025)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '战略人力资源部',
    'DEPARTMENT',
    'INACTIVE',
    '2020-01-01',
    'HISTORICAL',
    'ACTIVE',
    'NORMAL',
    '战略转型期，对接业务发展',
    2,
    '/1000000/1000004',
    0,
    '实施战略性人力资源管理转型',
    '1000000'
);

-- 当前记录：数字化人力资源部 (2025-现在)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '数字化人力资源部',
    'DEPARTMENT',
    'ACTIVE',
    '2025-01-01',
    'CURRENT',
    'ACTIVE',
    'NORMAL',
    '数字化转型升级',
    2,
    '/1000000/1000004',
    0,
    '负责人力资源数字化转型和智能化管理',
    '1000000'
);

-- 计划记录：AI智能人力资源中心 (2026年计划)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000004',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    'AI智能人力资源中心',
    'DEPARTMENT',
    'PLANNED',
    '2026-01-01',
    'PLANNED',
    'ACTIVE',
    'NORMAL',
    '人工智能技术升级计划',
    2,
    '/1000000/1000004',
    0,
    '基于AI技术的下一代人力资源管理中心',
    '1000000'
);

-- ============================================================================
-- 第三部分：创建其他测试组织数据
-- ============================================================================

-- 财务部 - 演示不同的时态历史
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000001',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '财务科',
    'DEPARTMENT',
    'INACTIVE',
    '2010-01-01',
    'HISTORICAL',
    'ACTIVE',
    'NORMAL',
    '公司成立时的财务管理部门',
    2,
    '/1000000/1000001',
    1,
    '负责基础财务核算工作',
    '1000000'
);

INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000001',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '财务部',
    'DEPARTMENT',
    'ACTIVE',
    '2018-01-01',
    'CURRENT',
    'ACTIVE',
    'NORMAL',
    '财务管理职能扩展升级',
    2,
    '/1000000/1000001',
    1,
    '负责全面财务管理和投资分析',
    '1000000'
);

-- 技术部 - 演示停用状态
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    suspended_at, suspension_reason,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000002',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '技术研发部',
    'DEPARTMENT',
    'INACTIVE',
    '2020-01-01',
    'CURRENT',
    'SUSPENDED',
    'NORMAL',
    '2024-12-01',
    '业务重组，暂时停用',
    '技术研发部门成立',
    2,
    '/1000000/1000002',
    2,
    '负责产品技术研发工作（当前已停用）',
    '1000000'
);

-- 市场部 - 演示软删除（已删除的记录）
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, effective_date,
    lifecycle_status, business_status, data_status,
    deleted_at, deletion_reason,
    change_reason, level, path, sort_order, description, parent_code
) VALUES (
    '1000003',
    '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    '市场推广部',
    'DEPARTMENT',
    'INACTIVE',
    '2022-01-01',
    'HISTORICAL',
    'ACTIVE',
    'DELETED',
    '2024-06-01',
    '组织重构，部门撤销',
    '市场推广部门成立',
    2,
    '/1000000/1000003',
    3,
    '负责市场推广和品牌宣传（已撤销）',
    '1000000'
);

COMMIT;

-- ============================================================================
-- 验证数据完整性
-- ============================================================================

-- 查看创建的测试数据概览
DO $$ 
BEGIN 
    RAISE NOTICE '============================================';
    RAISE NOTICE '测试数据创建完成！';
    RAISE NOTICE '============================================';
END $$;

-- 显示组织1000004的完整时态历史
SELECT 
    code,
    name,
    effective_date,
    end_date,
    lifecycle_status,
    business_status,
    data_status,
    change_reason
FROM organization_units 
WHERE code = '1000004' AND data_status = 'NORMAL'
ORDER BY effective_date;