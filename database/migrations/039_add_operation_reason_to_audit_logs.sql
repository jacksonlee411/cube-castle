-- 039_add_operation_reason_to_audit_logs.sql
-- 目标：恢复 audit_logs.operation_reason 字段，并让触发器/服务写入一致的操作原因。

BEGIN;

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS operation_reason TEXT;

UPDATE audit_logs
SET operation_reason = COALESCE(operation_reason, business_context->>'operation_reason', business_context->>'change_reason')
WHERE operation_reason IS NULL
  AND (
        business_context ? 'operation_reason'
        OR business_context ? 'change_reason'
      );

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    op_type TEXT;
    actor_id_text TEXT;
    actor_name TEXT;
    actor_type TEXT;
    request_token TEXT;
    before_snapshot JSONB := NULL;
    after_snapshot JSONB := NULL;
    change_items JSONB := '[]'::JSONB;
    modified_fields JSONB := '[]'::JSONB;
    key TEXT;
    old_value JSONB;
    new_value JSONB;
    data_type TEXT;
    change_reason TEXT;
    changed_by_val UUID;
    target_record UUID;
    target_tenant UUID;
    excluded_keys TEXT[] := ARRAY[
        'created_at','updated_at','tenant_id','record_id','path','code_path','name_path',
        'hierarchy_depth','metadata','changed_by','approved_by','request_id','is_current',
        'is_temporal','is_future'
    ];
BEGIN
    IF TG_OP = 'INSERT' THEN
        op_type := 'CREATE';
        target_record := NEW.record_id;
        target_tenant := NEW.tenant_id;
        changed_by_val := NEW.changed_by;
        change_reason := NEW.change_reason;
    ELSIF TG_OP = 'UPDATE' THEN
        op_type := 'UPDATE';
        target_record := COALESCE(NEW.record_id, OLD.record_id);
        target_tenant := COALESCE(NEW.tenant_id, OLD.tenant_id);
        changed_by_val := COALESCE(NEW.changed_by, OLD.changed_by);
        change_reason := COALESCE(NEW.change_reason, OLD.change_reason);
    ELSIF TG_OP = 'DELETE' THEN
        op_type := 'DELETE';
        target_record := OLD.record_id;
        target_tenant := OLD.tenant_id;
        changed_by_val := OLD.changed_by;
        change_reason := OLD.change_reason;
    ELSE
        op_type := TG_OP;
        target_record := COALESCE(NEW.record_id, OLD.record_id);
        target_tenant := COALESCE(NEW.tenant_id, OLD.tenant_id);
        changed_by_val := COALESCE(NEW.changed_by, OLD.changed_by);
        change_reason := COALESCE(NEW.change_reason, OLD.change_reason);
    END IF;

    request_token := COALESCE(
        current_setting('cube.request_id', true),
        current_setting('app.request_id', true),
        gen_random_uuid()::TEXT
    );

    actor_id_text := COALESCE(
        current_setting('cube.actor_id', true),
        COALESCE(changed_by_val::TEXT, 'system')
    );

    actor_name := COALESCE(
        current_setting('cube.actor_name', true),
        CASE WHEN actor_id_text = 'system' THEN 'System' ELSE actor_id_text END
    );

    actor_type := CASE WHEN actor_id_text = 'system' THEN 'SYSTEM' ELSE 'USER' END;

    change_reason := COALESCE(
        current_setting('cube.change_reason', true),
        change_reason
    );
    change_reason := NULLIF(change_reason, '');

    IF TG_OP <> 'INSERT' THEN
        before_snapshot := to_jsonb(OLD);
    END IF;

    IF TG_OP <> 'DELETE' THEN
        after_snapshot := to_jsonb(NEW);
    END IF;

    IF before_snapshot IS NOT NULL THEN
        FOREACH key IN ARRAY excluded_keys LOOP
            before_snapshot := before_snapshot - key;
        END LOOP;
    END IF;

    IF after_snapshot IS NOT NULL THEN
        FOREACH key IN ARRAY excluded_keys LOOP
            after_snapshot := after_snapshot - key;
        END LOOP;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        FOR key IN SELECT jsonb_object_keys(after_snapshot)
        LOOP
            old_value := COALESCE(before_snapshot -> key, 'null'::JSONB);
            new_value := COALESCE(after_snapshot -> key, 'null'::JSONB);

            IF old_value IS DISTINCT FROM new_value THEN
                data_type := infer_audit_change_datatype(
                    CASE
                        WHEN new_value IS NULL OR new_value = 'null'::JSONB THEN old_value
                        ELSE new_value
                    END
                );

                change_items := change_items || jsonb_build_array(
                    jsonb_build_object(
                        'field', key,
                        'oldValue', old_value,
                        'newValue', new_value,
                        'dataType', data_type
                    )
                );
                modified_fields := modified_fields || jsonb_build_array(key);
            END IF;
        END LOOP;

        IF jsonb_array_length(change_items) = 0 THEN
            RETURN NEW;
        END IF;

        IF jsonb_array_length(modified_fields) > 0 THEN
            modified_fields := (
                SELECT jsonb_agg(DISTINCT elem)
                  FROM jsonb_array_elements_text(modified_fields) AS t(elem)
            );
        END IF;
    END IF;

    INSERT INTO audit_logs (
        tenant_id,
        event_type,
        resource_type,
        actor_id,
        actor_type,
        action_name,
        request_id,
        success,
        resource_id,
        record_id,
        operation_reason,
        request_data,
        response_data,
        changes,
        modified_fields,
        business_context
    ) VALUES (
        target_tenant,
        op_type,
        'ORGANIZATION',
        actor_id_text,
        actor_type,
        op_type || '_ORGANIZATION',
        request_token,
        TRUE,
        target_record::TEXT,
        target_record,
        change_reason,
        COALESCE(before_snapshot, '{}'::JSONB),
        COALESCE(after_snapshot, '{}'::JSONB),
        change_items,
        COALESCE(modified_fields, '[]'::JSONB),
        jsonb_strip_nulls(jsonb_build_object(
            'actor_name', actor_name,
            'change_reason', change_reason,
            'operation_reason', change_reason,
            'trigger', 'log_audit_changes'
        ))
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

COMMIT;

SELECT 'audit_logs operation_reason column added' AS info, NOW() AS applied_at;
