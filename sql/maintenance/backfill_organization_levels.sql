-- backfill_organization_levels.sql
-- 目标: 基于现有 hierarchy_depth 或 code_path 重算 organization_units.level
-- 说明:
--   - 根节点应为 level = 0，对应 hierarchy_depth = 1
--   - 子节点 level = hierarchy_depth - 1
--   - 更新同时刷新 updated_at，保留 existing audit 逻辑

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

-- 可选: 使用 code_path 计算的备用逻辑（仅当 hierarchy_depth 缺失时启用）
-- WITH computed_levels AS (
--     SELECT
--         record_id,
--         GREATEST(array_length(string_to_array(COALESCE(NULLIF(code_path, ''), '/' || code), '/'), 1) - 1, 0) AS new_level
--     FROM organization_units
-- )
-- UPDATE organization_units AS u
-- SET
--     level = c.new_level,
--     updated_at = NOW()
-- FROM computed_levels AS c
-- WHERE u.record_id = c.record_id
--   AND u.level IS DISTINCT FROM c.new_level;

-- 运行后建议执行:
-- 1) SELECT level, COUNT(*) FROM organization_units GROUP BY level ORDER BY level;
-- 2) 针对重点组织抽样确认 level 是否符合预期。
