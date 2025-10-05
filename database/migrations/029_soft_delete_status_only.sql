-- 029_soft_delete_status_only.sql
-- 目标: 将软删除判定切换为仅依赖 status 字段，deleted_at 仅保留为审计信息。
--      本迁移同步更新触发器、检查约束、部分唯一索引及辅助函数，确保
--      CQRS 全链路使用统一的 status-only 语义。

SET search_path TO public;

-- =====================================================================
-- 1. 更新层级校验与时态标志触发器，仅依赖 status
-- =====================================================================

CREATE OR REPLACE FUNCTION validate_hierarchy_changes()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        IF check_circular_reference(NEW.code, NEW.parent_code, NEW.tenant_id) THEN
            RAISE EXCEPTION '不能设置父组织，会导致循环引用！组织 % 尝试设置父组织 %', NEW.code, NEW.parent_code;
        END IF;
    END IF;

    IF NEW.parent_code IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM organization_units
             WHERE code = NEW.parent_code
               AND tenant_id = NEW.tenant_id
               AND is_current = true
               AND status <> 'DELETED'
        ) THEN
            RAISE EXCEPTION '父组织不可用（不存在/已删除/非当前）！父组织编码: %', NEW.parent_code;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE FUNCTION calculate_org_hierarchy(p_code character varying, p_tenant_id uuid)
RETURNS TABLE(
    calculated_level integer,
    calculated_code_path character varying,
    calculated_name_path character varying,
    calculated_hierarchy_depth integer
)
LANGUAGE plpgsql
AS $$
DECLARE
    parent_info RECORD;
    current_name VARCHAR(255);
BEGIN
    SELECT name INTO current_name
      FROM organization_units
     WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
     LIMIT 1;

    SELECT
        ou.code,
        ou.level,
        ou.code_path,
        ou.name_path,
        ou.hierarchy_depth
      INTO parent_info
      FROM organization_units ou
     WHERE ou.code = (
            SELECT parent_code
              FROM organization_units
             WHERE code = p_code AND tenant_id = p_tenant_id AND is_current = true
             LIMIT 1
          )
       AND ou.tenant_id = p_tenant_id
       AND ou.is_current = true
       AND ou.status <> 'DELETED'
     LIMIT 1;

    IF parent_info.code IS NULL THEN
        calculated_level := 1;
        calculated_hierarchy_depth := 1;
        calculated_code_path := '/' || p_code;
        calculated_name_path := '/' || COALESCE(current_name, p_code);
    ELSE
        calculated_level := parent_info.level + 1;
        calculated_hierarchy_depth := parent_info.hierarchy_depth + 1;
        calculated_code_path := COALESCE(parent_info.code_path, '/' || parent_info.code) || '/' || p_code;
        calculated_name_path := COALESCE(parent_info.name_path, '/' || current_name) || '/' || COALESCE(current_name, p_code);
        IF calculated_hierarchy_depth > 17 THEN
            RAISE EXCEPTION '组织层级超过最大限制17级！当前尝试创建第%级组织。', calculated_hierarchy_depth;
        END IF;
    END IF;
    RETURN NEXT;
END;
$$;

CREATE OR REPLACE FUNCTION enforce_soft_delete_temporal_flags()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.status = 'DELETED' THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.effective_date > CURRENT_DATE THEN
        NEW.is_current := FALSE;
    ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= CURRENT_DATE THEN
        NEW.is_current := FALSE;
    ELSE
        NEW.is_current := TRUE;
    END IF;
    RETURN NEW;
END;
$$;

-- =====================================================================
-- 2. 刷新检查约束与禁止更新触发器
-- =====================================================================

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'organization_units'::regclass
          AND conname = 'chk_deleted_not_current'
    ) THEN
        ALTER TABLE organization_units
          DROP CONSTRAINT chk_deleted_not_current;
    END IF;

    ALTER TABLE organization_units
      ADD CONSTRAINT chk_deleted_not_current
      CHECK (CASE WHEN status = 'DELETED' THEN is_current = FALSE ELSE TRUE END);
END $$;

CREATE OR REPLACE FUNCTION prevent_update_deleted()
RETURNS trigger AS $$
BEGIN
    IF (OLD.status = 'DELETED') THEN
        RAISE EXCEPTION 'READ_ONLY_DELETED: cannot modify deleted record %', OLD.record_id
            USING ERRCODE = '55000';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_update_deleted ON organization_units;
CREATE TRIGGER trg_prevent_update_deleted
    BEFORE UPDATE ON organization_units
    FOR EACH ROW
    WHEN (OLD.status = 'DELETED')
    EXECUTE FUNCTION prevent_update_deleted();

-- =====================================================================
-- 3. 重建仅依赖 status 的部分唯一索引及相关索引
-- =====================================================================

DROP INDEX IF EXISTS uk_org_ver_active_only;
CREATE UNIQUE INDEX uk_org_ver_active_only
ON organization_units (tenant_id, code, effective_date)
WHERE status <> 'DELETED';

DROP INDEX IF EXISTS uk_org_current_active_only;
CREATE UNIQUE INDEX uk_org_current_active_only
ON organization_units (tenant_id, code)
WHERE is_current = TRUE AND status <> 'DELETED';

DROP INDEX IF EXISTS uk_org_temporal_point;
CREATE UNIQUE INDEX uk_org_temporal_point
ON organization_units (tenant_id, code, effective_date)
WHERE status <> 'DELETED';

DROP INDEX IF EXISTS uk_org_current;
CREATE UNIQUE INDEX uk_org_current
ON organization_units (tenant_id, code)
WHERE is_current = TRUE AND status <> 'DELETED';

DROP INDEX IF EXISTS ix_org_temporal_query;
CREATE INDEX ix_org_temporal_query
ON organization_units (tenant_id, code, effective_date DESC)
WHERE status <> 'DELETED';

DROP INDEX IF EXISTS ix_org_adjacent_versions;
CREATE INDEX ix_org_adjacent_versions
ON organization_units (tenant_id, code, effective_date, record_id)
WHERE status <> 'DELETED';

DROP INDEX IF EXISTS ix_org_current_lookup;
CREATE INDEX ix_org_current_lookup
ON organization_units (tenant_id, code, is_current)
WHERE is_current = TRUE AND status <> 'DELETED';

DROP INDEX IF EXISTS ix_org_temporal_boundaries;
CREATE INDEX ix_org_temporal_boundaries
ON organization_units (code, effective_date, end_date, is_current)
WHERE status <> 'DELETED';

DROP INDEX IF EXISTS ix_org_daily_transition;
CREATE INDEX ix_org_daily_transition
ON organization_units (effective_date, end_date, is_current)
WHERE status <> 'DELETED';

-- =====================================================================
-- 4. 更新辅助校验函数，移除对 deleted_at 的依赖
-- =====================================================================

CREATE OR REPLACE FUNCTION check_temporal_continuity(
    p_tenant_id UUID,
    p_code VARCHAR(7)
) RETURNS TABLE(
    issue_type TEXT,
    effective_date DATE,
    end_date DATE,
    message TEXT
) AS $$
BEGIN
    RETURN QUERY
    WITH ordered_versions AS (
        SELECT
            effective_date,
            end_date,
            ROW_NUMBER() OVER (ORDER BY effective_date) AS rn
        FROM organization_units
        WHERE tenant_id = p_tenant_id
          AND code = p_code
          AND status <> 'DELETED'
        ORDER BY effective_date
    ), version_overlaps AS (
        SELECT
            curr.effective_date,
            curr.end_date,
            'OVERLAP'::TEXT AS issue_type,
            'Version overlaps with next version'::TEXT AS message
        FROM ordered_versions curr
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL
          AND curr.end_date >= nxt.effective_date
    ), gaps AS (
        SELECT
            curr.effective_date,
            curr.end_date,
            'GAP'::TEXT AS issue_type,
            'Gap between versions'::TEXT AS message
        FROM ordered_versions curr
        JOIN ordered_versions nxt ON nxt.rn = curr.rn + 1
        WHERE curr.end_date IS NOT NULL
          AND curr.end_date + INTERVAL '1 day' < nxt.effective_date
    )
    SELECT * FROM version_overlaps
    UNION ALL
    SELECT * FROM gaps;
END;
$$ LANGUAGE plpgsql;

SELECT '029_soft_delete_status_only applied' AS status, NOW() AS applied_at;
