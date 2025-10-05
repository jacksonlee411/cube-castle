-- 015_unify_operation_type_reactivate.sql
-- Purpose: Enforce unified operation_type semantics (REACTIVATE instead of ACTIVATE)
-- Scope: audit_logs table only (API contract defines operationType ∈ {CREATE, UPDATE, SUSPEND, REACTIVATE, DELETE})

BEGIN;

-- 1) Backfill any legacy ACTIVATE values to REACTIVATE for consistency
UPDATE audit_logs
SET event_type = 'REACTIVATE'
WHERE event_type = 'ACTIVATE';

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
          AND tc.constraint_name = 'audit_logs_event_type_check_v2'
    ) INTO constraint_exists;

    IF NOT constraint_exists THEN
        -- 移除旧约束（若存在）
        SELECT EXISTS (
            SELECT 1
            FROM information_schema.table_constraints tc
            WHERE tc.table_name = 'audit_logs'
              AND tc.constraint_type = 'CHECK'
              AND tc.constraint_name = 'audit_logs_event_type_check'
        ) INTO constraint_exists;

        IF constraint_exists THEN
            EXECUTE 'ALTER TABLE audit_logs DROP CONSTRAINT audit_logs_event_type_check';
        END IF;

        EXECUTE 'ALTER TABLE audit_logs ADD CONSTRAINT audit_logs_event_type_check_v2
                 CHECK (event_type IN (
                     ''CREATE'',''UPDATE'',''DELETE'',''SUSPEND'',''REACTIVATE'',
                     ''QUERY'',''VALIDATION'',''AUTHENTICATION'',''ERROR''
                 ))';
    END IF;
END$$;

COMMIT;
