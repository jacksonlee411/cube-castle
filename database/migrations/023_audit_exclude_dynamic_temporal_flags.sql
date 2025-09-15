-- 023_audit_exclude_dynamic_temporal_flags.sql
-- Purpose: Align audit production with API-first principle by excluding
--          dynamic temporal flags (is_current, is_temporal, is_future)
--          from before_data/after_data. These flags are derived and should
--          not be persisted as part of immutable audit payloads.

BEGIN;

-- Replace log_audit_changes() to filter dynamic flags from row JSON
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    op_type VARCHAR(20);
    change_summary TEXT;
    rec_id UUID;
    before_json JSONB;
    after_json JSONB;
BEGIN
    IF TG_OP = 'INSERT' THEN
        op_type := NEW.operation_type;
        rec_id := NEW.record_id;
        change_summary := 'Created organization unit: ' || NEW.name;
    ELSIF TG_OP = 'UPDATE' THEN
        op_type := NEW.operation_type;
        rec_id := NEW.record_id;
        change_summary := 'Updated organization unit: ' || NEW.name;
    ELSIF TG_OP = 'DELETE' THEN
        op_type := 'DELETE';
        rec_id := OLD.record_id;
        change_summary := 'Deleted organization unit: ' || OLD.name;
    END IF;

    -- Build filtered snapshots (exclude dynamic flags)
    before_json := CASE WHEN TG_OP = 'INSERT'
                        THEN NULL
                        ELSE to_jsonb(OLD) - 'is_current' - 'is_temporal' - 'is_future'
                   END;
    after_json  := CASE WHEN TG_OP = 'DELETE'
                        THEN NULL
                        ELSE to_jsonb(NEW) - 'is_current' - 'is_temporal' - 'is_future'
                   END;

    INSERT INTO audit_logs (
        -- Legacy columns (backward compatible)
        entity_code,
        operation_type,
        operated_by_id,
        operated_by_name,
        operation_reason,
        before_data,
        after_data,
        changes_summary,
        tenant_id,
        record_id,
        -- Unified contract columns
        event_type,
        resource_type,
        resource_id,
        actor_id,
        business_context
    ) VALUES (
        COALESCE(NEW.code, OLD.code),
        op_type,
        COALESCE(NEW.operated_by_id, OLD.operated_by_id),
        COALESCE(NEW.operated_by_name, OLD.operated_by_name),
        COALESCE(NEW.operation_reason, OLD.operation_reason),
        before_json,
        after_json,
        change_summary,
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        rec_id,
        -- New contract
        op_type,
        'ORGANIZATION',
        rec_id,
        COALESCE(NEW.operated_by_id, OLD.operated_by_id)::text,
        jsonb_build_object(
          'actor_name', COALESCE(NEW.operated_by_name, OLD.operated_by_name),
          'operation_reason', COALESCE(NEW.operation_reason, OLD.operation_reason)
        )
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();

COMMIT;

