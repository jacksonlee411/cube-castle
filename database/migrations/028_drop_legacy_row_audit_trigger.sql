-- 028_drop_legacy_row_audit_trigger.sql
-- 目的: 移除与现行 audit_logs 架构不兼容的行级审计触发器，
--       避免在 organization_units INSERT/UPDATE/DELETE 时因缺失列/NOT NULL 约束导致失败。
-- 说明: 审计改由应用层结构化审计 (cmd/organization-command-service/internal/audit) 负责。

BEGIN;

-- 幂等删除触发器与函数
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'audit_changes_trigger'
  ) THEN
    EXECUTE 'DROP TRIGGER IF EXISTS audit_changes_trigger ON organization_units';
  END IF;
END$$;

DROP FUNCTION IF EXISTS log_audit_changes() CASCADE;

COMMIT;

