-- 011_audit_record_id_fix.sql
-- 修复审计历史查询问题：添加record_id字段到audit_logs表
-- 解决前端按recordId查询审计历史时找不到记录的问题

-- Migration: 添加record_id字段到audit_logs表
-- Author: System
-- Date: 2025-09-06
-- Related Issue: 审计历史页签查询不到修改记录

-- 第一步：添加record_id列（允许NULL，稍后回填）
ALTER TABLE audit_logs ADD COLUMN record_id UUID;

-- 第二步：从before_data或after_data中提取record_id进行回填
-- 此操作会将JSON中的record_id字段提取到专门的列中
UPDATE audit_logs 
SET record_id = COALESCE(
    -- 下划线风格
    NULLIF((before_data->>'record_id'), '')::uuid,
    NULLIF((after_data->>'record_id'), '')::uuid,
    -- 驼峰风格（兼容历史数据）
    NULLIF((before_data->>'recordId'), '')::uuid,
    NULLIF((after_data->>'recordId'), '')::uuid
);

-- 第三步：为record_id创建索引（使用实际的字段名timestamp）
-- 修正：时间列为 operation_timestamp 而非 timestamp
DROP INDEX IF EXISTS idx_audit_logs_record_id_time;
CREATE INDEX idx_audit_logs_record_id_time 
ON audit_logs (record_id, operation_timestamp DESC);

-- 第四步：设置record_id为NOT NULL（仅在回填成功后）
-- 注意：如果有历史数据没有record_id，这步会失败，需要手动处理
-- ALTER TABLE audit_logs ALTER COLUMN record_id SET NOT NULL;

-- 第五步：添加注释
COMMENT ON COLUMN audit_logs.record_id IS '组织单元时态版本的唯一标识，用于精确审计查询';

-- 第六步：更新审计触发器，确保新记录包含record_id
DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    operation_type_val VARCHAR(20);
    changes_summary_val TEXT;
BEGIN
    -- 确定操作类型
    IF TG_OP = 'INSERT' THEN
        operation_type_val := NEW.operation_type;
    ELSIF TG_OP = 'UPDATE' THEN
        operation_type_val := NEW.operation_type;
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
    
    -- 记录审计日志（新增record_id字段）
    INSERT INTO audit_logs (
        entity_code,
        operation_type,
        operated_by_id,
        operated_by_name,
        operation_reason,
        before_data,
        after_data,
        changes_summary,
        tenant_id,
        record_id  -- 新增字段
    ) VALUES (
        COALESCE(NEW.code, OLD.code),
        operation_type_val,
        COALESCE(NEW.operated_by_id, OLD.operated_by_id),
        COALESCE(NEW.operated_by_name, OLD.operated_by_name),
        COALESCE(NEW.operation_reason, OLD.operation_reason),
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE row_to_json(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE row_to_json(NEW) END,
        changes_summary_val,
        COALESCE(NEW.tenant_id, OLD.tenant_id),
        COALESCE(NEW.record_id, OLD.record_id)  -- 提取record_id
    );
    
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 重新创建触发器
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();
