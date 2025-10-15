-- Rollback script for 045_drop_position_legacy_columns.sql

BEGIN;

ALTER TABLE positions
    ADD COLUMN IF NOT EXISTS current_holder_id UUID,
    ADD COLUMN IF NOT EXISTS current_holder_name VARCHAR(255),
    ADD COLUMN IF NOT EXISTS filled_date DATE,
    ADD COLUMN IF NOT EXISTS current_assignment_type VARCHAR(20);

CREATE INDEX IF NOT EXISTS idx_positions_holder
    ON positions(tenant_id, current_holder_id)
    WHERE current_holder_id IS NOT NULL;

COMMIT;
