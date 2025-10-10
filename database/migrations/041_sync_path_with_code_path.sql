-- 041_sync_path_with_code_path.sql
-- 目标: 在迁移阶段将 organization_units.path 与 code_path 对齐，清理遗留的 path_codepath_mismatch 异常
-- 说明:
--   1. 本迁移仅更新层级路径字段，不涉及业务状态或时态字段
--   2. 保留 path 列供旧链路读取，但内容统一与 code_path 一致，后续可安全移除

BEGIN;

UPDATE organization_units
SET path = code_path,
    updated_at = GREATEST(updated_at, NOW())
WHERE (path IS NULL OR TRIM(BOTH '/' FROM path) IS DISTINCT FROM TRIM(BOTH '/' FROM code_path))
  AND code_path IS NOT NULL;

COMMIT;
