-- =============================================
-- Unit Type 枚举更新迁移 (修正版)
-- 移除 COST_CENTER，COMPANY 改为 ORGANIZATION_UNIT
-- 版本: v1.1
-- 日期: 2025-08-23
-- =============================================

BEGIN;

-- 1. 备份当前数据
DROP TABLE IF EXISTS organization_units_unittype_backup;
CREATE TABLE organization_units_unittype_backup AS
SELECT * FROM organization_units;

-- 2. 先删除现有的CHECK约束
ALTER TABLE organization_units 
DROP CONSTRAINT IF EXISTS organization_units_unit_type_check;

ALTER TABLE organization_units 
DROP CONSTRAINT IF EXISTS valid_unit_type;

-- 3. 更新现有数据中的单位类型
UPDATE organization_units 
SET unit_type = 'ORGANIZATION_UNIT' 
WHERE unit_type = 'COMPANY';

-- 删除COST_CENTER类型的记录（如果存在）
-- 注意：生产环境中应该先确认这些记录的处理策略
DELETE FROM organization_units 
WHERE unit_type = 'COST_CENTER';

-- 4. 添加新的CHECK约束以支持新的枚举值
ALTER TABLE organization_units 
ADD CONSTRAINT valid_unit_type 
CHECK (unit_type IN ('DEPARTMENT', 'ORGANIZATION_UNIT', 'PROJECT_TEAM'));

-- 5. 创建新的索引以优化按类型查询
CREATE INDEX IF NOT EXISTS idx_org_unit_type_optimized
ON organization_units(tenant_id, unit_type, is_current)
WHERE is_current = true;

-- 6. 更新组织统计视图（如果存在）
DROP VIEW IF EXISTS organization_stats_view;
CREATE OR REPLACE VIEW organization_stats_view AS
SELECT 
    tenant_id,
    unit_type,
    COUNT(*) as count,
    COUNT(CASE WHEN is_current = true THEN 1 END) as current_count,
    COUNT(CASE WHEN is_current = false THEN 1 END) as historical_count
FROM organization_units 
WHERE deleted_at IS NULL
GROUP BY tenant_id, unit_type;

COMMIT;

-- 验证迁移结果
SELECT 
    'Migration Results:' as info,
    unit_type,
    COUNT(*) as count
FROM organization_units 
WHERE is_current = true
GROUP BY unit_type
ORDER BY unit_type;
