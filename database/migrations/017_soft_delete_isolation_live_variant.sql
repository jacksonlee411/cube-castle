-- 017_soft_delete_isolation_live_variant.sql
-- 软删除隔离 & 时态标志规范（适配包含 status/deleted_at 的现网表结构）

SET search_path TO public;

-- 校验父节点：必须当前且未删除
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
               AND deleted_at IS NULL
        ) THEN
            RAISE EXCEPTION '父组织不可用（不存在/已删除/非当前）！父组织编码: %', NEW.parent_code;
        END IF;
    END IF;
    RETURN NEW;
END;
$$;

-- 计算层级：忽略已删除父节点
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
       AND ou.deleted_at IS NULL
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

-- 时态标志规范
CREATE OR REPLACE FUNCTION enforce_soft_delete_temporal_flags()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
        NEW.is_current := FALSE;
        NEW.is_future := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.effective_date > CURRENT_DATE THEN
        NEW.is_current := FALSE;
        NEW.is_future := TRUE;
    ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= CURRENT_DATE THEN
        NEW.is_current := FALSE;
        NEW.is_future := FALSE;
    ELSE
        NEW.is_current := TRUE;
        NEW.is_future := FALSE;
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS enforce_soft_delete_temporal_flags_trigger ON organization_units;
CREATE TRIGGER enforce_soft_delete_temporal_flags_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION enforce_soft_delete_temporal_flags();

-- 约束：DELETED 不得为当前
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
         WHERE conrelid = 'organization_units'::regclass
           AND conname = 'chk_deleted_not_current'
    ) THEN
        ALTER TABLE organization_units
          ADD CONSTRAINT chk_deleted_not_current
          CHECK (CASE WHEN status = 'DELETED' OR deleted_at IS NOT NULL THEN is_current = FALSE ELSE TRUE END);
    END IF;
END $$;

-- 数据修复（与在线执行版本一致）
UPDATE organization_units
   SET is_current = FALSE,
       is_future = FALSE
 WHERE (status = 'DELETED' OR deleted_at IS NOT NULL)
   AND (is_current = TRUE OR is_future = TRUE);

UPDATE organization_units
   SET is_current = CASE 
                        WHEN effective_date > CURRENT_DATE THEN FALSE
                        WHEN end_date IS NOT NULL AND end_date <= CURRENT_DATE THEN FALSE
                        ELSE TRUE
                    END,
       is_future = CASE WHEN effective_date > CURRENT_DATE THEN TRUE ELSE FALSE END
 WHERE status <> 'DELETED' AND deleted_at IS NULL;

UPDATE organization_units c
   SET parent_code = NULL
 WHERE parent_code IS NOT NULL
   AND EXISTS (
        SELECT 1 FROM organization_units p
         WHERE p.code = c.parent_code 
           AND (p.status = 'DELETED' OR p.deleted_at IS NOT NULL)
   );

UPDATE organization_units SET name = name;

ANALYZE organization_units;

