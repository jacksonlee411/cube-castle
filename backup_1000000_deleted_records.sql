-- 1000000组织记录清理备份 (2025-09-07)
-- 保留记录: 2010-01-01 de88e079-cfcd-4425-a6b0-3d4251f1216d 高谷集团总部
-- 删除的记录备份:

-- 记录1: fd056c2d-4067-40d4-a4b3-8ee3731d8e4c (2025-09-01)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('fd056c2d-4067-40d4-a4b3-8ee3731d8e4c', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '高谷集团总部 - 测试版本', 'ORGANIZATION_UNIT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-06 14:23:21.29777+00', NULL, '2025-09-01', '2025-09-02', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录2: 7c167763-99d6-40d2-b774-86f82a2743fb (2025-09-01)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('7c167763-99d6-40d2-b774-86f82a2743fb', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '高谷集团总部', 'ORGANIZATION_UNIT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-06 23:35:39.265348+00', NULL, '2025-09-01', '2025-09-02', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录3: 7a7ec722-d2af-4651-9c05-736efa66ba8e (2025-09-01)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('7a7ec722-d2af-4651-9c05-736efa66ba8e', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '时间轴测试组织', 'DEPARTMENT', 'ACTIVE', 1, NULL, NULL, NULL, NULL, '2025-09-07 00:21:25.483519+00', NULL, '2025-09-01', '2025-09-04', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录4: 271318df-0546-4f54-b678-f4b2beb0efb8 (2025-09-01)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('271318df-0546-4f54-b678-f4b2beb0efb8', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '高谷集团总部', 'ORGANIZATION_UNIT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-07 01:06:56.729554+00', NULL, '2025-09-01', '2025-09-04', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录5: 8c43a220-6c9f-4422-b239-3cce23c8c19a (2025-09-03)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('8c43a220-6c9f-4422-b239-3cce23c8c19a', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '测试API响应格式', 'ORGANIZATION_UNIT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-06 14:24:12.904355+00', NULL, '2025-09-03', '2025-09-04', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录6: 944750e9-e844-4dd1-abb0-71e187c99e01 (2025-09-05)
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('944750e9-e844-4dd1-abb0-71e187c99e01', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '前端错误处理修复验证', 'ORGANIZATION_UNIT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-06 14:26:19.661607+00', NULL, '2025-09-05', '2025-09-11', false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 记录7: 4dff97ca-8ed2-42af-a779-b92a83b14f50 (2025-09-12) - 这是之前设置为is_current=true的记录
INSERT INTO organization_units (record_id, tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, profile, created_at, updated_at, effective_date, end_date, is_current, is_temporal, change_reason, deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason) VALUES 
('4dff97ca-8ed2-42af-a779-b92a83b14f50', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '1000000', NULL, '诊断测试版本', 'DEPARTMENT', 'DELETED', 1, NULL, NULL, NULL, NULL, '2025-09-07 01:04:10.694872+00', NULL, '2025-09-12', NULL, false, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

-- 保留的记录信息:
-- record_id: de88e079-cfcd-4425-a6b0-3d4251f1216d
-- name: 高谷集团总部
-- unit_type: ORGANIZATION_UNIT
-- status: ACTIVE
-- effective_date: 2010-01-01
-- end_date: 2025-08-31
-- is_current: false (需要更新为true)