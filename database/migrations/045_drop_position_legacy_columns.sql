-- 045_drop_position_legacy_columns.sql
-- 移除职位表中的临时任职冗余字段，统一任职事实来源到 position_assignments

BEGIN;

DROP INDEX IF EXISTS idx_positions_holder;

ALTER TABLE positions
    DROP COLUMN IF EXISTS current_holder_id,
    DROP COLUMN IF EXISTS current_holder_name,
    DROP COLUMN IF EXISTS filled_date,
    DROP COLUMN IF EXISTS current_assignment_type;

COMMIT;
