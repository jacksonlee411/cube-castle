-- Rollback for 048_extend_position_assignments.sql
-- 目的：移除 Stage4 任职扩展字段与索引，恢复到 047 版本的约束定义。

BEGIN;

DROP INDEX IF EXISTS idx_position_assignments_auto_revert_due;

ALTER TABLE position_assignments
    DROP CONSTRAINT IF EXISTS chk_position_assignments_auto_revert;

ALTER TABLE position_assignments
    DROP CONSTRAINT IF EXISTS chk_position_assignments_dates;

ALTER TABLE position_assignments
    DROP COLUMN IF EXISTS reminder_sent_at,
    DROP COLUMN IF EXISTS auto_revert,
    DROP COLUMN IF EXISTS acting_until;

ALTER TABLE position_assignments
    ADD CONSTRAINT chk_position_assignments_dates
        CHECK (end_date IS NULL OR end_date > effective_date);

COMMIT;
