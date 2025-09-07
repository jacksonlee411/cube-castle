-- 027_validate_parent_on_update.sql
-- 目的：在 UPDATE 时也校验 parent_code 的有效性（必须指向“当前且未删除”的父节点）。
-- 策略：仅当 parent_code 变更且为非 NULL 时触发校验；设置为 NULL 不拦截（允许降级为根）。

-- 复用现有校验函数：validate_parent_available()
-- 该函数仅依赖 NEW.* 字段，可同时用于 INSERT/UPDATE 触发。

-- 清理已有同名触发器（幂等）
DROP TRIGGER IF EXISTS validate_parent_available_update_trigger ON organization_units;

-- 在 UPDATE 时新增校验触发器
CREATE TRIGGER validate_parent_available_update_trigger
    BEFORE UPDATE ON organization_units
    FOR EACH ROW
    WHEN (
        NEW.parent_code IS NOT NULL
        AND (NEW.parent_code IS DISTINCT FROM OLD.parent_code)
    )
    EXECUTE FUNCTION validate_parent_available();

-- 可选：一次性体检报告（不修改数据，仅供上线后人工核查）
-- 说明：如需发现历史上已存在的“无效父引用”，可执行下列查询：
-- SELECT c.code AS child_code, c.parent_code, p.is_current, p.is_deleted
--   FROM organization_units c
--   LEFT JOIN organization_units p ON p.code = c.parent_code
--  WHERE c.parent_code IS NOT NULL
--    AND (p.code IS NULL OR p.is_current = FALSE OR p.is_deleted = TRUE)
--  LIMIT 100;

-- 结束

