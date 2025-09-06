-- 020_align_audit_logs_schema.sql
-- 目的：对齐审计日志(audit_logs)表结构到统一契约：
--   - 列：event_type, resource_type, resource_id(uuid), actor_id, actor_type, action_name, request_id, business_context(jsonb)
--   - 向后兼容：保留旧列(entity_type, entity_code, operation_type, operated_by_*, changes_summary, business_entity_id)
--   - 数据回填：用旧列回填新列，确保现有查询(基于 resource_id/event_type 等)可用

BEGIN;

-- 1) 新增缺失列（若不存在）
ALTER TABLE audit_logs
  ADD COLUMN IF NOT EXISTS event_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS resource_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS actor_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS action_name VARCHAR(100),
  ADD COLUMN IF NOT EXISTS request_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS resource_id TEXT,
  ADD COLUMN IF NOT EXISTS business_context JSONB;

-- 2) 统一 resource_id 类型为 UUID（若可转换）
DO $$
BEGIN
  -- 仅在列存在且非 uuid 类型的情况下尝试转换
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
     WHERE table_name='audit_logs' AND column_name='resource_id' AND data_type <> 'uuid'
  ) THEN
    -- 尝试将可转换的值转换为uuid，其余保留为NULL
    ALTER TABLE audit_logs
      ALTER COLUMN resource_id TYPE uuid USING (
        CASE 
          WHEN resource_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' 
          THEN resource_id::uuid ELSE NULL END
      );
  END IF;
END
$$;

-- 3) 用旧列回填新列
UPDATE audit_logs SET
  event_type = COALESCE(event_type, operation_type),
  resource_type = COALESCE(resource_type, entity_type, 'ORGANIZATION'),
  actor_id = COALESCE(actor_id, operated_by_id::text),
  business_context = COALESCE(
    business_context,
    jsonb_build_object(
      'actor_name', operated_by_name,
      'operation_reason', operation_reason,
      'changes_summary', changes_summary
    )
  )
WHERE TRUE;

-- resource_id 优先使用现有列；否则回填为 record_id（若存在）
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
     WHERE table_name='audit_logs' AND column_name='record_id'
  ) THEN
    UPDATE audit_logs
       SET resource_id = COALESCE(resource_id, record_id)
     WHERE resource_id IS NULL AND record_id IS NOT NULL;
  END IF;
END
$$;

-- 4) 为查询建立或修复索引（按代码使用：resource_id + 时间）
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
     WHERE c.relname='idx_audit_logs_resource_time' AND n.nspname='public'
  ) THEN
    CREATE INDEX idx_audit_logs_resource_time ON audit_logs (resource_id, operation_timestamp DESC);
  END IF;
END
$$;

-- 5) 更新触发器：log_audit_changes 若存在，则改为同时写入新列（保持幂等）
DO $$
BEGIN
  IF EXISTS (
      SELECT 1 FROM pg_proc WHERE proname='log_audit_changes'
  ) THEN
    DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;
  END IF;
END
$$;

CREATE OR REPLACE FUNCTION log_audit_changes()
RETURNS TRIGGER AS $$
DECLARE
    op_type VARCHAR(20);
    change_summary TEXT;
    rec_id UUID;
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
        CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE row_to_json(OLD) END,
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE row_to_json(NEW) END,
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

DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units;
CREATE TRIGGER audit_changes_trigger
    AFTER INSERT OR UPDATE OR DELETE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION log_audit_changes();

COMMIT;

