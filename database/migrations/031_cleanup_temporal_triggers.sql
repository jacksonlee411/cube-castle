-- 031_cleanup_temporal_triggers.sql
-- 目的：移除依赖已弃用列的触发器/函数，防止创建组织时返回 500 (CREATE_ERROR)
-- 核心修复：
--   1. 重建 log_audit_changes() 函数，避免引用 operation_reason 等已迁移字段
--   2. 重建 organization_version_trigger()，移除对 is_temporal 及历史版本表的依赖

BEGIN;

-- 1. 重建审计触发器函数，兼容精简后的 audit_logs 架构
DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
DROP FUNCTION IF EXISTS log_audit_changes();

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    op_type TEXT;
    actor_uuid UUID;
    actor_type TEXT;
    request_token TEXT := COALESCE(current_setting('cube.request_id', true), gen_random_uuid()::text);
    before_snapshot JSONB := NULL;
    after_snapshot JSONB := NULL;
BEGIN
    IF TG_OP = 'INSERT' THEN
        op_type := 'CREATE';
        after_snapshot := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN
        op_type := 'UPDATE';
        before_snapshot := to_jsonb(OLD);
        after_snapshot := to_jsonb(NEW);
    ELSE
        op_type := 'DELETE';
        before_snapshot := to_jsonb(OLD);
    END IF;

    actor_uuid := COALESCE(NEW.changed_by, OLD.changed_by);
    actor_type := CASE WHEN actor_uuid IS NULL THEN 'SYSTEM' ELSE 'USER' END;

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
        operation_type,
        record_id,
        before_data,
        after_data,
        business_context
    ) VALUES (
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        op_type,
        'ORGANIZATION',
        COALESCE(actor_uuid::text, 'system'),
        actor_type,
        op_type,
        request_token,
        TRUE,
        COALESCE(NEW.record_id, OLD.record_id),
        op_type,
        COALESCE(NEW.record_id, OLD.record_id),
        before_snapshot,
        after_snapshot,
        jsonb_strip_nulls(jsonb_build_object(
            'change_reason', COALESCE(NEW.change_reason, OLD.change_reason),
            'trigger', 'log_audit_changes'
        ))
    );

    RETURN COALESCE(NEW, OLD);
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'log_audit_changes fallback: %', SQLERRM;
        RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();

-- 2. 重建组织版本触发器，移除对 is_temporal/历史表的依赖
DROP TRIGGER IF EXISTS organization_version_trigger ON organization_units;
DROP FUNCTION IF EXISTS organization_version_trigger();

CREATE OR REPLACE FUNCTION organization_version_trigger()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' THEN
        NEW.version := COALESCE(OLD.version, 0) + 1;
        NEW.updated_at := NOW();
    ELSIF TG_OP = 'INSERT' THEN
        NEW.version := COALESCE(NEW.version, 1);
        NEW.created_at := COALESCE(NEW.created_at, NOW());
        NEW.updated_at := COALESCE(NEW.updated_at, NOW());
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER organization_version_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION organization_version_trigger();

COMMIT;

