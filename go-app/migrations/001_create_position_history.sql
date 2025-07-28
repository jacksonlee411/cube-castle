-- Migration: Create position_history table for temporal employee position tracking
-- Version: 1.0.0
-- Date: 2025-07-28
-- Description: Employee position history with temporal tracking for complete audit trail

-- Create position_history table
CREATE TABLE position_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    
    -- Position information snapshot
    position_title VARCHAR(100) NOT NULL,
    department VARCHAR(100) NOT NULL,
    job_level VARCHAR(50),
    location VARCHAR(100),
    employment_type VARCHAR(20) NOT NULL CHECK (employment_type IN ('FULL_TIME', 'PART_TIME', 'CONTRACT', 'INTERN')),
    
    -- Reporting relationship
    reports_to_employee_id UUID,
    
    -- Temporal fields (core)
    effective_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    
    -- Change metadata
    change_reason TEXT,
    is_retroactive BOOLEAN DEFAULT FALSE NOT NULL,
    created_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Salary range information
    min_salary DECIMAL(15,2),
    max_salary DECIMAL(15,2),
    currency CHAR(3) DEFAULT 'CNY',
    
    -- Constraints
    CONSTRAINT valid_date_range CHECK (end_date IS NULL OR end_date > effective_date),
    CONSTRAINT valid_salary_range CHECK (max_salary IS NULL OR min_salary IS NULL OR max_salary >= min_salary)
);

-- Create indexes for temporal queries
CREATE INDEX idx_position_history_temporal 
ON position_history (tenant_id, employee_id, effective_date, end_date);

CREATE UNIQUE INDEX idx_position_history_current 
ON position_history (tenant_id, employee_id) 
WHERE end_date IS NULL;

CREATE INDEX idx_position_history_effective_date 
ON position_history (tenant_id, effective_date);

CREATE INDEX idx_position_history_retroactive 
ON position_history (tenant_id, is_retroactive, created_at);

CREATE INDEX idx_position_history_reports_to 
ON position_history (tenant_id, reports_to_employee_id, effective_date)
WHERE end_date IS NULL;

-- Create composite index for date range queries
CREATE INDEX idx_position_history_date_range
ON position_history (tenant_id, employee_id, effective_date DESC, end_date DESC);

-- Foreign key constraints
ALTER TABLE position_history 
ADD CONSTRAINT fk_position_history_employee 
FOREIGN KEY (tenant_id, employee_id) REFERENCES person(tenant_id, id)
ON DELETE CASCADE;

ALTER TABLE position_history 
ADD CONSTRAINT fk_position_history_reports_to 
FOREIGN KEY (tenant_id, reports_to_employee_id) REFERENCES person(tenant_id, id)
ON DELETE SET NULL;

-- Enable row level security
ALTER TABLE position_history ENABLE ROW LEVEL SECURITY;

-- Create RLS policy for tenant isolation
CREATE POLICY position_history_tenant_isolation ON position_history
    FOR ALL TO application_role
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Create policy for salary field access control
CREATE POLICY position_history_salary_access ON position_history
    FOR SELECT TO application_role
    USING (
        -- Allow access to salary fields only for users with HR compensation permission
        CASE 
            WHEN current_setting('app.user_permissions', true) LIKE '%hr.compensation.read%' THEN TRUE
            ELSE min_salary IS NULL AND max_salary IS NULL
        END
    );

-- Grant permissions
GRANT SELECT, INSERT, UPDATE ON position_history TO application_role;
GRANT USAGE, SELECT ON SEQUENCE position_history_id_seq TO application_role;

-- Create trigger function for temporal consistency validation
CREATE OR REPLACE FUNCTION validate_position_history_temporal_consistency()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate no overlapping periods for the same employee
    IF EXISTS (
        SELECT 1 FROM position_history 
        WHERE tenant_id = NEW.tenant_id 
          AND employee_id = NEW.employee_id
          AND id != COALESCE(NEW.id, '00000000-0000-0000-0000-000000000000'::UUID)
          AND effective_date <= COALESCE(NEW.end_date, 'infinity'::timestamp)
          AND COALESCE(end_date, 'infinity'::timestamp) > NEW.effective_date
    ) THEN
        RAISE EXCEPTION 'Temporal conflict: overlapping position periods for employee %', NEW.employee_id;
    END IF;
    
    -- Validate effective date is not in the far future (more than 2 years)
    IF NEW.effective_date > NOW() + INTERVAL '2 years' THEN
        RAISE EXCEPTION 'Effective date cannot be more than 2 years in the future';
    END IF;
    
    -- Validate retroactive flag is set correctly
    IF NEW.effective_date < NOW() - INTERVAL '1 day' AND NOT NEW.is_retroactive THEN
        NEW.is_retroactive = TRUE;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER trigger_validate_position_history_temporal_consistency
    BEFORE INSERT OR UPDATE ON position_history
    FOR EACH ROW
    EXECUTE FUNCTION validate_position_history_temporal_consistency();

-- Create function for automatic position closure
CREATE OR REPLACE FUNCTION auto_close_previous_positions()
RETURNS TRIGGER AS $$
BEGIN
    -- If this is a new current position (end_date is NULL), close previous open positions
    IF NEW.end_date IS NULL THEN
        UPDATE position_history 
        SET end_date = NEW.effective_date - INTERVAL '1 day'
        WHERE tenant_id = NEW.tenant_id 
          AND employee_id = NEW.employee_id
          AND id != NEW.id
          AND end_date IS NULL
          AND effective_date < NEW.effective_date;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for automatic position closure
CREATE TRIGGER trigger_auto_close_previous_positions
    AFTER INSERT ON position_history
    FOR EACH ROW
    EXECUTE FUNCTION auto_close_previous_positions();

-- Create audit trigger function
CREATE OR REPLACE FUNCTION audit_position_history_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- Log position changes to audit table (if exists)
    IF TG_OP = 'INSERT' THEN
        INSERT INTO audit_log (
            table_name, operation, record_id, tenant_id, 
            changed_by, changed_at, new_values
        ) VALUES (
            'position_history', 'INSERT', NEW.id, NEW.tenant_id,
            NEW.created_by, NOW(), row_to_json(NEW)
        );
        RETURN NEW;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit_log (
            table_name, operation, record_id, tenant_id,
            changed_by, changed_at, old_values, new_values
        ) VALUES (
            'position_history', 'UPDATE', NEW.id, NEW.tenant_id,
            NEW.created_by, NOW(), row_to_json(OLD), row_to_json(NEW)
        );
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create audit trigger (will be enabled when audit_log table exists)
-- CREATE TRIGGER trigger_audit_position_history_changes
--     AFTER INSERT OR UPDATE ON position_history
--     FOR EACH ROW
--     EXECUTE FUNCTION audit_position_history_changes();

-- Create helpful views
CREATE VIEW current_positions AS
SELECT 
    ph.*,
    p.legal_name,
    p.email,
    p.employee_id as employee_number
FROM position_history ph
JOIN person p ON ph.tenant_id = p.tenant_id AND ph.employee_id = p.id
WHERE ph.end_date IS NULL;

CREATE VIEW position_timeline_view AS
SELECT 
    ph.*,
    p.legal_name,
    p.email,
    p.employee_id as employee_number,
    manager.legal_name as manager_name,
    manager.email as manager_email
FROM position_history ph
JOIN person p ON ph.tenant_id = p.tenant_id AND ph.employee_id = p.id
LEFT JOIN person manager ON ph.tenant_id = manager.tenant_id AND ph.reports_to_employee_id = manager.id
ORDER BY ph.employee_id, ph.effective_date;

-- Grant permissions on views
GRANT SELECT ON current_positions TO application_role;
GRANT SELECT ON position_timeline_view TO application_role;

-- Create indexes on views (materialized views could be considered for performance)
-- Note: These would be materialized view indexes if we convert to materialized views later

-- Performance analysis and optimization hints
COMMENT ON TABLE position_history IS 'Employee position history with temporal tracking. Supports point-in-time queries and complete audit trail.';
COMMENT ON INDEX idx_position_history_temporal IS 'Primary temporal index for efficient as-of-date and range queries';
COMMENT ON INDEX idx_position_history_current IS 'Optimized index for current position lookups';
COMMENT ON INDEX idx_position_history_date_range IS 'Composite index for efficient timeline queries';

-- Maintenance commands (to be run periodically)
-- ANALYZE position_history;
-- REINDEX INDEX CONCURRENTLY idx_position_history_temporal;

-- Performance tuning settings (adjust based on usage patterns)
-- ALTER TABLE position_history SET (
--     fillfactor = 90,  -- Leave room for updates
--     autovacuum_vacuum_scale_factor = 0.1,
--     autovacuum_analyze_scale_factor = 0.05
-- );

COMMIT;