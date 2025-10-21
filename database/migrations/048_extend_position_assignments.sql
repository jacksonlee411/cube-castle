-- 048_extend_position_assignments.sql
-- 目的：扩展 position_assignments 表以支持 Stage4 代理任职能力，新增 acting_until / auto_revert / reminder_sent_at 字段，
--      并补充约束与索引，保证自动恢复扫描与数据一致性。

BEGIN;

ALTER TABLE position_assignments
    ADD COLUMN IF NOT EXISTS acting_until DATE,
    ADD COLUMN IF NOT EXISTS auto_revert BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS reminder_sent_at TIMESTAMPTZ;

ALTER TABLE position_assignments
    DROP CONSTRAINT IF EXISTS chk_position_assignments_dates;

ALTER TABLE position_assignments
    ADD CONSTRAINT chk_position_assignments_dates
        CHECK (
            (end_date IS NULL OR end_date > effective_date)
            AND (acting_until IS NULL OR acting_until > effective_date)
        );

ALTER TABLE position_assignments
    DROP CONSTRAINT IF EXISTS chk_position_assignments_auto_revert;

ALTER TABLE position_assignments
    ADD CONSTRAINT chk_position_assignments_auto_revert
        CHECK (
            auto_revert = false
            OR (assignment_type = 'ACTING' AND acting_until IS NOT NULL)
        );

DROP INDEX IF EXISTS idx_position_assignments_auto_revert_due;
CREATE INDEX IF NOT EXISTS idx_position_assignments_auto_revert_due
    ON position_assignments(tenant_id, auto_revert, acting_until)
    WHERE assignment_type = 'ACTING';

COMMIT;
