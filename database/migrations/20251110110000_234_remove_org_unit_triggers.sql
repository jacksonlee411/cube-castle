-- +goose Up
DROP TRIGGER IF EXISTS validate_parent_available_update_trigger ON public.organization_units;
DROP TRIGGER IF EXISTS validate_parent_available_trigger ON public.organization_units;
DROP TRIGGER IF EXISTS update_hierarchy_paths_trigger ON public.organization_units;
DROP TRIGGER IF EXISTS trg_prevent_update_deleted ON public.organization_units;
DROP TRIGGER IF EXISTS enforce_temporal_flags_trigger ON public.organization_units;
DROP TRIGGER IF EXISTS audit_changes_trigger ON public.organization_units;

DROP FUNCTION IF EXISTS public.validate_parent_available();
DROP FUNCTION IF EXISTS public.update_hierarchy_paths();
DROP FUNCTION IF EXISTS public.prevent_update_deleted();
DROP FUNCTION IF EXISTS public.log_audit_changes();
DROP FUNCTION IF EXISTS public.enforce_temporal_flags();

-- +goose Down
CREATE FUNCTION public.enforce_temporal_flags() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
BEGIN
    IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.effective_date > utc_date THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.end_date IS NOT NULL AND NEW.end_date <= utc_date THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    RETURN NEW;
END;
$$;

CREATE FUNCTION public.log_audit_changes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
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
$$;

CREATE FUNCTION public.prevent_update_deleted() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF (OLD.status = 'DELETED') THEN
        RAISE EXCEPTION 'READ_ONLY_DELETED: cannot modify deleted record %', OLD.record_id
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.update_hierarchy_paths() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.parent_code IS NULL THEN
        NEW.code_path := '/' || NEW.code;
        NEW.name_path := '/' || NEW.name;
        NEW.level := 1;
    ELSE
        SELECT 
            parent.code_path || '/' || NEW.code,
            parent.name_path || '/' || NEW.name,
            parent.level + 1
          INTO NEW.code_path, NEW.name_path, NEW.level
          FROM organization_units parent
         WHERE parent.code = NEW.parent_code
           AND parent.tenant_id = NEW.tenant_id
           AND parent.is_current = TRUE
           AND parent.status <> 'DELETED'
           AND parent.deleted_at IS NULL
         LIMIT 1;

        IF NOT FOUND THEN
            NEW.parent_code := NULL;
            NEW.code_path := '/' || NEW.code;
            NEW.name_path := '/' || NEW.name;
            NEW.level := 1;
        END IF;
    END IF;

    NEW.hierarchy_depth := NEW.level;
    RETURN NEW;
END;
$$;

CREATE FUNCTION public.validate_parent_available() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        PERFORM 1
          FROM organization_units p
         WHERE p.code = NEW.parent_code
           AND p.tenant_id = NEW.tenant_id
           AND p.is_current = TRUE
           AND p.status <> 'DELETED'
           AND p.deleted_at IS NULL
         LIMIT 1;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'PARENT_NOT_AVAILABLE: parent % is not current or has been deleted', NEW.parent_code
                USING ERRCODE = 'foreign_key_violation';
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER audit_changes_trigger AFTER INSERT OR DELETE OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.log_audit_changes();
CREATE TRIGGER enforce_temporal_flags_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.enforce_temporal_flags();
CREATE TRIGGER trg_prevent_update_deleted BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((old.status)::text = 'DELETED'::text)) EXECUTE FUNCTION public.prevent_update_deleted();
CREATE TRIGGER update_hierarchy_paths_trigger BEFORE INSERT OR UPDATE ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.update_hierarchy_paths();
CREATE TRIGGER validate_parent_available_trigger BEFORE INSERT ON public.organization_units FOR EACH ROW EXECUTE FUNCTION public.validate_parent_available();
CREATE TRIGGER validate_parent_available_update_trigger BEFORE UPDATE ON public.organization_units FOR EACH ROW WHEN (((new.parent_code IS NOT NULL) AND ((new.parent_code)::text IS DISTINCT FROM (old.parent_code)::text))) EXECUTE FUNCTION public.validate_parent_available();
