-- 020_align_audit_logs_schema.sql (compatibility rewrite)
-- 说明：现网 audit_logs 已对齐统一契约（无 operation_reason / operation_type 等旧列），
--       本迁移仅确保缺失列与索引补齐，并在 legacy 架构缺席时静默跳过。

BEGIN;

-- 1. 按需补齐通用列（若历史环境仍缺失）
ALTER TABLE audit_logs
  ADD COLUMN IF NOT EXISTS event_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS resource_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS actor_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS action_name VARCHAR(100),
  ADD COLUMN IF NOT EXISTS request_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS resource_id VARCHAR(100),
  ADD COLUMN IF NOT EXISTS business_context JSONB DEFAULT '{}'::jsonb;

-- 2. 若 legacy 架构仍保留旧列，则尝试一次性回填（避免影响已兼容环境）
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name = 'audit_logs' AND column_name = 'entity_code'
  ) THEN
    EXECUTE '
      UPDATE audit_logs
         SET event_type = COALESCE(event_type, operation_type),
             resource_type = COALESCE(resource_type, entity_type, ''ORGANIZATION''),
             resource_id = COALESCE(resource_id, entity_code),
             actor_id = COALESCE(actor_id, operated_by_id::text)
       WHERE TRUE;
    ';
  END IF;
END$$;

-- 3. 建立常用索引（现网已存在则跳过）
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
     WHERE c.relname = 'idx_audit_logs_record_id_time' AND n.nspname = 'public'
  ) THEN
    EXECUTE 'CREATE INDEX idx_audit_logs_record_id_time ON audit_logs (record_id, "timestamp" DESC)';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
     WHERE c.relname = 'idx_audit_logs_resource_timestamp' AND n.nspname = 'public'
  ) THEN
    EXECUTE 'CREATE INDEX idx_audit_logs_resource_timestamp ON audit_logs (resource_type, resource_id, "timestamp" DESC)';
  END IF;
END$$;

-- 4. 若需要，可对 legacy log_audit_changes 进行补救；无旧列时保持现有实现
DO $$
BEGIN
  IF EXISTS (
      SELECT 1 FROM pg_proc WHERE proname = 'log_audit_changes'
  ) AND EXISTS (
      SELECT 1 FROM information_schema.columns
       WHERE table_name = 'organization_units' AND column_name = 'operation_reason'
  ) THEN
    -- 若 legacy 仍有 operation_reason，可保持原逻辑（此处留作兼容占位）
    RAISE NOTICE 'log_audit_changes legacy variant detected; please review manually.';
  END IF;
END$$;

COMMIT;

-- 迁移说明：对于已采用统一审计结构的环境，上述语句均为幂等操作，不会改写现有触发器。
