-- 012_fix_audit_trigger_compatibility.sql
-- 修复审计触发器与实际表结构的兼容性问题
-- 解决触发器中引用不存在字段的问题

-- 删除旧的触发器函数并创建兼容的新版本
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    operation_type_val VARCHAR(20);
    changes_summary_val TEXT;
    operated_by_id_val UUID;
    operated_by_name_val VARCHAR(255);
    operation_reason_val TEXT;
BEGIN
    -- 确定操作类型（基于TG_OP，不依赖表字段）
    IF TG_OP = 'INSERT' THEN
        operation_type_val := 'CREATE';
    ELSIF TG_OP = 'UPDATE' THEN
        operation_type_val := 'UPDATE';
    ELSIF TG_OP = 'DELETE' THEN
        operation_type_val := 'DELETE';
    END IF;
    
    -- 生成变更摘要
    IF TG_OP = 'INSERT' THEN
        changes_summary_val := 'Created organization unit: ' || NEW.name;
    ELSIF TG_OP = 'UPDATE' THEN
        changes_summary_val := 'Updated organization unit: ' || NEW.name;
    ELSIF TG_OP = 'DELETE' THEN
        changes_summary_val := 'Deleted organization unit: ' || OLD.name;
    END IF;
    
    -- 设置操作者信息（使用默认值，因为表中没有这些字段）
    operated_by_id_val := '550e8400-e29b-41d4-a716-446655440000'::UUID; -- System user
    operated_by_name_val := 'System';
    
    -- 获取操作原因（从实际存在的字段）
    operation_reason_val := COALESCE(NEW.change_reason, OLD.change_reason, 'System operation');
    
    -- 插入到新的audit_logs表结构中
    INSERT INTO audit_logs (
        tenant_id,
        event_type,
        resource_type,
        actor_id,
        actor_type,
        action_name,
        request_id,
        resource_id,
        operation_reason,
        before_data,
        after_data,
        record_id
    ) VALUES (
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        operation_type_val,
        'ORGANIZATION',
        operated_by_id_val::VARCHAR,
        'SYSTEM',
        operation_type_val || '_ORGANIZATION',
        gen_random_uuid()::VARCHAR,
        COALESCE(NEW.record_id, OLD.record_id),
        operation_reason_val,
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE row_to_json(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE row_to_json(NEW) END,
        COALESCE(NEW.record_id, OLD.record_id)
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 重新创建触发器
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();