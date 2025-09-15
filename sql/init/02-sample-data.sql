-- Cube Castle 示例数据
-- 创建层级组织结构用于演示

-- 设置默认租户ID
SET session.default_tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

-- 插入根级组织 (Level 1)
INSERT INTO organization_units (
    code, tenant_id, name, unit_type, status, description, profile,
    effective_date, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '高谷集团', 'COMPANY', 'ACTIVE', 
 '高谷科技集团总公司', 
 '{"legalName": "高谷科技集团有限公司", "registrationNumber": "91110000123456789X", "taxId": "110000123456789", "industry": "软件开发", "incorporationDate": "2020-03-15"}',
 '2024-01-01', '789e0123-e89b-12d3-a456-426614174001', 'System Admin', '公司成立');

-- 插入二级部门 (Level 2)
INSERT INTO organization_units (
    code, parent_code, tenant_id, name, unit_type, status, description, profile,
    effective_date, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000001', '1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '技术部', 'DEPARTMENT', 'ACTIVE',
 '负责产品研发和技术创新',
 '{"budget": 5000000, "managerPositionCode": "POS-MGR-001", "costCenterCode": "CC001", "headCountLimit": 50, "establishedDate": "2024-01-01"}',
 '2024-01-01', '789e0123-e89b-12d3-a456-426614174002', 'Zhang San', '业务发展需要'),

('1000002', '1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '产品部', 'DEPARTMENT', 'ACTIVE',
 '负责产品设计和用户体验',
 '{"budget": 3000000, "managerPositionCode": "POS-MGR-002", "costCenterCode": "CC002", "headCountLimit": 30, "establishedDate": "2024-01-01"}',
 '2024-01-01', '789e0123-e89b-12d3-a456-426614174002', 'Li Si', '产品线扩展'),

('1000003', '1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '市场部', 'DEPARTMENT', 'ACTIVE',
 '负责市场推广和品牌建设',
 '{"budget": 2000000, "managerPositionCode": "POS-MGR-003", "costCenterCode": "CC003", "headCountLimit": 25, "establishedDate": "2024-01-01"}',
 '2024-01-01', '789e0123-e89b-12d3-a456-426614174003', 'Wang Wu', '市场战略需要');

-- 插入三级子部门 (Level 3)
INSERT INTO organization_units (
    code, parent_code, tenant_id, name, unit_type, status, description, profile,
    effective_date, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000011', '1000001', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '前端开发组', 'DEPARTMENT', 'ACTIVE',
 '专注于前端技术开发',
 '{"budget": 1500000, "managerPositionCode": "POS-MGR-011", "costCenterCode": "CC011", "headCountLimit": 15, "establishedDate": "2024-02-01"}',
 '2024-02-01', '789e0123-e89b-12d3-a456-426614174004', 'Chen Liu', '技术专业化分工'),

('1000012', '1000001', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '后端开发组', 'DEPARTMENT', 'ACTIVE',
 '专注于后端技术开发',
 '{"budget": 2000000, "managerPositionCode": "POS-MGR-012", "costCenterCode": "CC012", "headCountLimit": 20, "establishedDate": "2024-02-01"}',
 '2024-02-01', '789e0123-e89b-12d3-a456-426614174005', 'Zhao Qi', '技术架构优化'),

('1000021', '1000002', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', 'UI设计组', 'DEPARTMENT', 'ACTIVE',
 '用户界面设计和视觉设计',
 '{"budget": 800000, "managerPositionCode": "POS-MGR-021", "costCenterCode": "CC021", "headCountLimit": 8, "establishedDate": "2024-02-01"}',
 '2024-02-01', '789e0123-e89b-12d3-a456-426614174006', 'Zhou Ba', '设计专业化');

-- 插入项目团队 (特殊组织类型)
INSERT INTO organization_units (
    code, parent_code, tenant_id, name, unit_type, status, description, profile,
    effective_date, end_date, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000101', '1000001', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', 'AI项目组', 'PROJECT_TEAM', 'ACTIVE',
 '人工智能产品开发项目',
 '{"projectCode": "PROJ-2025-001", "projectManager": "EMP-001", "startDate": "2025-01-01", "endDate": "2025-12-31", "budget": 3000000}',
 '2025-01-01', '2025-12-31', '789e0123-e89b-12d3-a456-426614174007', 'Wu Jiu', '战略项目启动');

-- 插入未来生效的组织 (演示时态功能)
INSERT INTO organization_units (
    code, parent_code, tenant_id, name, unit_type, status, description, profile,
    effective_date, is_current, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000004', '1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '国际业务部', 'DEPARTMENT', 'ACTIVE',
'负责国际市场拓展',
'{"budget": 4000000, "managerPositionCode": "POS-MGR-004", "costCenterCode": "CC004", "headCountLimit": 40, "establishedDate": "2025-06-01"}',
'2025-06-01', false, '789e0123-e89b-12d3-a456-426614174008', 'Zheng Shi', '国际化战略布局');

-- 插入已暂停的组织 (演示状态管理)
INSERT INTO organization_units (
    code, parent_code, tenant_id, name, unit_type, status, description, profile,
    effective_date, operation_type, operated_by_id, operated_by_name, operation_reason
) VALUES 
('1000099', '1000000', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '临时项目部', 'DEPARTMENT', 'INACTIVE',
 '已完成的临时项目部门',
 '{"budget": 0, "managerPositionCode": "", "costCenterCode": "CC099", "headCountLimit": 0, "establishedDate": "2024-01-01"}',
 '2024-01-01', 'SUSPEND', '789e0123-e89b-12d3-a456-426614174009', 'Admin User', '项目结束');

-- 更新统计信息
ANALYZE organization_units;
ANALYZE audit_logs;

-- 验证数据完整性
DO $$
DECLARE
    total_count INTEGER;
    active_count INTEGER;
    level_distribution TEXT;
BEGIN
    SELECT COUNT(*) INTO total_count FROM organization_units WHERE is_current = true;
    SELECT COUNT(*) INTO active_count FROM organization_units WHERE is_current = true AND status = 'ACTIVE';
    
    RAISE NOTICE '✅ 组织单元数据初始化完成:';
    RAISE NOTICE '   - 总计组织: % 个', total_count;
    RAISE NOTICE '   - 活跃组织: % 个', active_count;
    RAISE NOTICE '   - 层级分布: 1级(1个), 2级(3个), 3级(3个), 项目组(1个)';
    RAISE NOTICE '   - 包含未来生效组织和状态演示数据';
    RAISE NOTICE '🎯 演示数据可用于GraphQL查询和REST API测试';
END $$;
