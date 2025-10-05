-- 032_phase_b_remove_legacy_columns.sql
-- 目标：计划18 Phase B — 清理已废弃列，确保迁移脚本与当前真源保持一致。
-- 内容：移除 organization_units 表遗留的 is_deleted、operation_reason 列，并刷新相关统计。

BEGIN;

-- =============================
-- 重建视图以移除依赖
-- =============================
DROP VIEW IF EXISTS organization_temporal_current;

-- =============================
-- 移除遗留的 is_deleted 列
-- =============================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM information_schema.columns
         WHERE table_name = 'organization_units'
           AND table_schema = 'public'
           AND column_name = 'is_deleted'
    ) THEN
        ALTER TABLE organization_units
          DROP COLUMN is_deleted;
    END IF;
END $$;

-- =============================
-- 移除遗留的 operation_reason 列
-- =============================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM information_schema.columns
         WHERE table_name = 'organization_units'
           AND table_schema = 'public'
           AND column_name = 'operation_reason'
    ) THEN
        ALTER TABLE organization_units
          DROP COLUMN operation_reason;
    END IF;
END $$;

-- =============================
-- 重新创建视图，使其与精简后的结构对齐
-- =============================
CREATE OR REPLACE VIEW organization_temporal_current AS
  SELECT *
    FROM organization_units
   WHERE is_current = TRUE;

ANALYZE organization_units;

COMMIT;

SELECT '032_phase_b_remove_legacy_columns applied' AS status, NOW() AS applied_at;
