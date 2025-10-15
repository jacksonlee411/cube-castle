-- 044_create_position_assignments.sql
-- 创建职位任职记录表，支持事件周期模式与多租户隔离

BEGIN;

CREATE TABLE IF NOT EXISTS position_assignments (
    assignment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    position_code VARCHAR(8) NOT NULL,
    position_record_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    employee_name VARCHAR(255) NOT NULL,
    employee_number VARCHAR(64),
    assignment_type VARCHAR(20) NOT NULL,
    assignment_status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    fte NUMERIC(5,2) NOT NULL DEFAULT 1.0,
    start_date DATE NOT NULL,
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT false,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_position_assignments_type
        CHECK (assignment_type IN ('PRIMARY', 'SECONDARY', 'ACTING')),
    CONSTRAINT chk_position_assignments_status
        CHECK (assignment_status IN ('PENDING', 'ACTIVE', 'ENDED')),
    CONSTRAINT chk_position_assignments_dates
        CHECK (end_date IS NULL OR end_date > start_date),
    CONSTRAINT chk_position_assignments_fte
        CHECK (fte >= 0 AND fte <= 1),
    CONSTRAINT fk_position_assignments_position
        FOREIGN KEY (tenant_id, position_code, position_record_id)
        REFERENCES positions(tenant_id, code, record_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_position_assignments_active
    ON position_assignments(tenant_id, position_code, employee_id)
    WHERE is_current = true AND assignment_status = 'ACTIVE';

CREATE UNIQUE INDEX IF NOT EXISTS uk_position_assignments_start
    ON position_assignments(tenant_id, position_code, employee_id, start_date);

CREATE INDEX IF NOT EXISTS idx_position_assignments_position
    ON position_assignments(tenant_id, position_code, start_date DESC);

CREATE INDEX IF NOT EXISTS idx_position_assignments_employee
    ON position_assignments(tenant_id, employee_id, start_date DESC);

CREATE INDEX IF NOT EXISTS idx_position_assignments_status
    ON position_assignments(tenant_id, assignment_status, is_current);

COMMIT;
