-- 027_fix_audit_trigger_op_type_from_tg_op.sql
-- 目的: 修复审计触发器对已移除列 NEW.operation_type 的引用，
--       遵循 API 优先与动态标记不入库原则，
--       由 TG_OP 推导 operation_type，避免对表结构的倒退。
-- 影响范围: audit_logs 写入逻辑；organization_units INSERT/UPDATE/DELETE 审计
-- 风险评估: 低（幂等重建触发器函数）；无数据破坏

BEGIN;

-- 清理旧版本函数（若存在）
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

-- 以 TG_OP 推导操作类型；过滤动态标记字段
CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    op_type VARCHAR(20);
    change_summary TEXT;
    rec_id UUID;
    before_json JSONB;
    after_json JSONB;
BEGIN
    -- 1) 推导操作类型（不再读取 NEW.operation_type）
    IF TG_OP = 'INSERT' THEN
        op_type := 'CREATE';
        rec_id := NEW.record_id;
        change_summary := 'Created organization unit: ' || NEW.name;
    ELSIF TG_OP = 'UPDATE' THEN
        op_type := 'UPDATE';
        rec_id := NEW.record_id;
        change_summary := 'Updated organization unit: ' || NEW.name;
    ELSIF TG_OP = 'DELETE' THEN
        op_type := 'DELETE';
        rec_id := OLD.record_id;
        change_summary := 'Deleted organization unit: ' || OLD.name;
    END IF;

    -- 2) 构建过滤后的镜像（排除动态标记: is_current/is_temporal/is_future）
    before_json := CASE WHEN TG_OP = 'INSERT'
                        THEN NULL
                        ELSE to_jsonb(OLD) - 'is_current' - 'is_temporal' - 'is_future'
                   END;
    after_json  := CASE WHEN TG_OP = 'DELETE'
                        THEN NULL
                        ELSE to_jsonb(NEW) - 'is_current' - 'is_temporal' - 'is_future'
                   END;

    -- 3) 写入审计日志（兼容旧列 + 对齐统一契约列）
    INSERT INTO audit_logs (
        -- 旧列（向后兼容）
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
        -- 新列（统一契约）
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
        -- 新列
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

-- 重新创建触发器（幂等）
DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();

COMMIT;

