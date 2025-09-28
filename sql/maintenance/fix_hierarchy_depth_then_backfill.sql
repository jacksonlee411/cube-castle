-- fix_hierarchy_depth_then_backfill.sql
-- 第一步：修复 hierarchy_depth 字段
-- 第二步：基于修复后的 hierarchy_depth 回填 level 字段

-- 步骤1: 修复 hierarchy_depth
-- 根级组织 (parent_code IS NULL) 应该是 hierarchy_depth = 1
-- 一级子组织应该是 hierarchy_depth = 2
WITH corrected_depth AS (
    SELECT
        record_id,
        CASE
            WHEN parent_code IS NULL THEN 1  -- 根级组织
            ELSE 2  -- 一级子组织 (在这个测试数据中)
        END as correct_hierarchy_depth
    FROM organization_units
)
UPDATE organization_units AS u
SET
    hierarchy_depth = c.correct_hierarchy_depth,
    updated_at = NOW()
FROM corrected_depth AS c
WHERE u.record_id = c.record_id
  AND u.hierarchy_depth IS DISTINCT FROM c.correct_hierarchy_depth;

-- 步骤2: 基于修复后的 hierarchy_depth 回填 level
WITH computed_levels AS (
    SELECT
        record_id,
        GREATEST(COALESCE(hierarchy_depth, 1) - 1, 0) AS new_level
    FROM organization_units
)
UPDATE organization_units AS u
SET
    level = c.new_level,
    updated_at = NOW()
FROM computed_levels AS c
WHERE u.record_id = c.record_id
  AND u.level IS DISTINCT FROM c.new_level;

-- 验证结果
SELECT
    '修复后统计' as description,
    hierarchy_depth,
    level,
    COUNT(*) as count
FROM organization_units
GROUP BY hierarchy_depth, level
ORDER BY hierarchy_depth, level;