-- 030_fix_is_current_with_utc_alignment.sql
-- 目的：修复 is_current 标记在 UTC 对齐场景下的初始化错误
-- 背景：旧版本在本地时区比较 effective_date，导致根组织等记录的 is_current 被错误地置为 false
-- 行动：
--   1. 重新计算每个组织单元的 end_date（基于下一版本的 effective_date - 1 日）
--   2. 以 UTC 自然日重新推导 is_current 标记，确保当前生效版本唯一为 true

BEGIN;

SET LOCAL TIME ZONE 'UTC';

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'audit_changes_trigger'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units DISABLE TRIGGER audit_changes_trigger';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'trg_prevent_update_deleted'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units DISABLE TRIGGER trg_prevent_update_deleted';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'organization_version_trigger'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units DISABLE TRIGGER organization_version_trigger';
    END IF;
END$$;

WITH ordered AS (
    SELECT
        record_id,
        tenant_id,
        code,
        effective_date,
        LEAD(effective_date) OVER (
            PARTITION BY tenant_id, code
            ORDER BY effective_date
        ) AS next_effective
    FROM organization_units
    WHERE status <> 'DELETED'
)
UPDATE organization_units ou
SET end_date = CASE
        WHEN ordered.next_effective IS NULL THEN NULL
        ELSE ordered.next_effective - INTERVAL '1 day'
    END
FROM ordered
WHERE ou.record_id = ordered.record_id;

WITH current_candidates AS (
    SELECT
        record_id,
        tenant_id,
        code,
        ROW_NUMBER() OVER (
            PARTITION BY tenant_id, code
            ORDER BY effective_date DESC NULLS LAST
        ) AS rn,
        effective_date,
        end_date
    FROM organization_units
    WHERE status <> 'DELETED'
      AND effective_date <= CURRENT_DATE
      AND (end_date IS NULL OR end_date >= CURRENT_DATE)
)
UPDATE organization_units ou
SET is_current = cc.rn = 1
FROM current_candidates cc
WHERE ou.record_id = cc.record_id;

WITH current_candidates AS (
    SELECT
        record_id,
        tenant_id,
        code,
        ROW_NUMBER() OVER (
            PARTITION BY tenant_id, code
            ORDER BY effective_date DESC NULLS LAST
        ) AS rn,
        effective_date,
        end_date
    FROM organization_units
    WHERE status <> 'DELETED'
      AND effective_date <= CURRENT_DATE
      AND (end_date IS NULL OR end_date >= CURRENT_DATE)
)
UPDATE organization_units ou
SET is_current = FALSE
WHERE status <> 'DELETED'
  AND NOT EXISTS (
      SELECT 1
      FROM current_candidates cc
      WHERE cc.record_id = ou.record_id
  );

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'organization_version_trigger'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units ENABLE TRIGGER organization_version_trigger';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'trg_prevent_update_deleted'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units ENABLE TRIGGER trg_prevent_update_deleted';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'audit_changes_trigger'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units ENABLE TRIGGER audit_changes_trigger';
    END IF;
END$$;

COMMIT;
