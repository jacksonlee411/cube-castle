-- 014_normalize_audit_logs.sql
-- 目标：规范化 audit_logs 表结构，补齐 012/013 期望的列，确保后续修复脚本可执行

BEGIN;

-- 基础列补齐（若不存在则添加）
ALTER TABLE audit_logs
  ADD COLUMN IF NOT EXISTS event_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS resource_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS actor_id VARCHAR(64),
  ADD COLUMN IF NOT EXISTS actor_type VARCHAR(50),
  ADD COLUMN IF NOT EXISTS action_name VARCHAR(100),
  ADD COLUMN IF NOT EXISTS request_id VARCHAR(64),
  ADD COLUMN IF NOT EXISTS resource_id VARCHAR(64),
  ADD COLUMN IF NOT EXISTS changes JSONB,
  ADD COLUMN IF NOT EXISTS modified_fields JSONB;

-- 兼容：若不存在 record_id 列则创建（011 理论已创建）
DO $$ BEGIN
  PERFORM 1 FROM information_schema.columns 
   WHERE table_name = 'audit_logs' AND column_name = 'record_id';
  IF NOT FOUND THEN
    EXECUTE 'ALTER TABLE audit_logs ADD COLUMN record_id UUID';
  END IF;
END $$;

-- 修复：若出现 record_id 被错误写为 actor_id 的情况，优先以 resource_id 校正
UPDATE audit_logs
   SET record_id = resource_id::uuid
 WHERE record_id IS NOT NULL
   AND actor_id IS NOT NULL
   AND resource_id IS NOT NULL
   AND record_id = actor_id::uuid
   AND resource_id ~* '^[0-9a-f-]{36}$'
   AND record_id <> resource_id::uuid;

-- 索引：保障按 (record_id, operation_timestamp) 查询性能
DO $$ BEGIN
  PERFORM 1 FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
   WHERE c.relname = 'idx_audit_logs_record_id_time'
     AND n.nspname = 'public';
  IF NOT FOUND THEN
    EXECUTE 'CREATE INDEX idx_audit_logs_record_id_time ON audit_logs (record_id, "timestamp" DESC)';
  END IF;
END $$;

COMMIT;
