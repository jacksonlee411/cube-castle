-- 047_rename_position_assignments_start_date.sql
-- 目的：统一 position_assignments 表的时态字段命名，从 start_date 调整为 effective_date，保持与其它时态表一致。

BEGIN;

ALTER TABLE position_assignments
    RENAME COLUMN start_date TO effective_date;

ALTER TABLE position_assignments
    DROP CONSTRAINT IF EXISTS chk_position_assignments_dates;

ALTER TABLE position_assignments
    ADD CONSTRAINT chk_position_assignments_dates
        CHECK (end_date IS NULL OR end_date > effective_date);

DROP INDEX IF EXISTS uk_position_assignments_start;
CREATE UNIQUE INDEX IF NOT EXISTS uk_position_assignments_effective
    ON position_assignments(tenant_id, position_code, employee_id, effective_date);

DROP INDEX IF EXISTS idx_position_assignments_position;
CREATE INDEX IF NOT EXISTS idx_position_assignments_position
    ON position_assignments(tenant_id, position_code, effective_date DESC);

DROP INDEX IF EXISTS idx_position_assignments_employee;
CREATE INDEX IF NOT EXISTS idx_position_assignments_employee
    ON position_assignments(tenant_id, employee_id, effective_date DESC);

COMMIT;
