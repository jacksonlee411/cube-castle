-- 015_unify_operation_type_reactivate.sql
-- Purpose: Enforce unified operation_type semantics (REACTIVATE instead of ACTIVATE)
-- Scope: audit_logs table only (API contract defines operationType ∈ {CREATE, UPDATE, SUSPEND, REACTIVATE, DELETE})

BEGIN;

-- 1) Backfill any legacy ACTIVATE values to REACTIVATE for consistency
UPDATE audit_logs
SET operation_type = 'REACTIVATE'
WHERE operation_type = 'ACTIVATE';

-- 2) Add CHECK constraint to restrict allowed values
DO $$
DECLARE
    constraint_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM information_schema.table_constraints tc
        WHERE tc.table_name = 'audit_logs'
          AND tc.constraint_type = 'CHECK'
          AND tc.constraint_name = 'audit_logs_operation_type_check'
    ) INTO constraint_exists;

    IF NOT constraint_exists THEN
        EXECUTE $$ALTER TABLE audit_logs
                 ADD CONSTRAINT audit_logs_operation_type_check
                 CHECK (operation_type IN ('CREATE','UPDATE','SUSPEND','REACTIVATE','DELETE'))$$;
    END IF;
END$$;

COMMIT;

