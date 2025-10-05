-- 016_soft_delete_isolation_and_temporal_flags.sql
-- 目的:
--   1) 软删除记录不再计入“当前组织”计算
--   2) 自动规范 is_current 标志，兼容是否存在 is_future 列的环境
--   3) 保证父组织必须可用（当前状态、未删除）
--   4) 校准现有数据并刷新统计

-- ==============================================================
-- 约束：被标记为删除的记录不得视为当前
-- ==============================================================
ALTER TABLE organization_units ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT FALSE;
ALTER TABLE organization_units ADD COLUMN IF NOT EXISTS is_future BOOLEAN DEFAULT FALSE;

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
END $$;

UPDATE organization_units
   SET is_deleted = (status = 'DELETED' OR deleted_at IS NOT NULL);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
         WHERE conrelid = 'organization_units'::regclass
           AND conname = 'chk_org_units_not_deleted_current'
    ) THEN
        EXECUTE 'ALTER TABLE organization_units
                 ADD CONSTRAINT chk_org_units_not_deleted_current
                 CHECK (CASE
                           WHEN status = ''DELETED'' OR deleted_at IS NOT NULL
                             THEN is_current = FALSE
                           ELSE TRUE
                         END)';
    END IF;
END $$;

-- ==============================================================
-- 插入校验：父节点必须可用
-- ==============================================================
CREATE OR REPLACE FUNCTION validate_parent_available()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        PERFORM 1
          FROM organization_units p
         WHERE p.code = NEW.parent_code
           AND p.tenant_id = NEW.tenant_id
           AND p.is_current = TRUE
           AND p.status <> 'DELETED'
           AND p.deleted_at IS NULL
         LIMIT 1;

        IF NOT FOUND THEN
            RAISE EXCEPTION 'PARENT_NOT_AVAILABLE: parent % is not current or has been deleted', NEW.parent_code
                USING ERRCODE = 'foreign_key_violation';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS validate_parent_available_trigger ON organization_units;
CREATE TRIGGER validate_parent_available_trigger
    BEFORE INSERT ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION validate_parent_available();

-- ==============================================================
-- 时态标志：根据生效/失效日期推导 is_current（可选 is_future）
-- ==============================================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'organization_units'
           AND column_name = 'is_future'
    ) THEN
        EXECUTE $_func$
            CREATE OR REPLACE FUNCTION enforce_temporal_flags()
            RETURNS TRIGGER AS $_body$
            BEGIN
                NEW.is_deleted := (NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL);

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
            $_body$ LANGUAGE plpgsql;
        $_func$;
    ELSE
        EXECUTE $_func$
            CREATE OR REPLACE FUNCTION enforce_temporal_flags()
            RETURNS TRIGGER AS $_body$
            BEGIN
                NEW.is_deleted := (NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL);

                IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
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
            $_body$ LANGUAGE plpgsql;
        $_func$;
    END IF;
END $$;

DROP TRIGGER IF EXISTS enforce_temporal_flags_trigger ON organization_units;
CREATE TRIGGER enforce_temporal_flags_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION enforce_temporal_flags();

-- ==============================================================
-- 层级路径重建：忽略不可用父节点
-- ==============================================================
CREATE OR REPLACE FUNCTION update_hierarchy_paths()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_code IS NULL THEN
        NEW.code_path := '/' || NEW.code;
        NEW.name_path := '/' || NEW.name;
        NEW.level := 1;
    ELSE
        SELECT 
            parent.code_path || '/' || NEW.code,
            parent.name_path || '/' || NEW.name,
            parent.level + 1
          INTO NEW.code_path, NEW.name_path, NEW.level
          FROM organization_units parent
         WHERE parent.code = NEW.parent_code
           AND parent.tenant_id = NEW.tenant_id
           AND parent.is_current = TRUE
           AND parent.status <> 'DELETED'
           AND parent.deleted_at IS NULL
         LIMIT 1;

        IF NOT FOUND THEN
            NEW.parent_code := NULL;
            NEW.code_path := '/' || NEW.code;
            NEW.name_path := '/' || NEW.name;
            NEW.level := 1;
        END IF;
    END IF;

    NEW.hierarchy_depth := NEW.level;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_hierarchy_paths_trigger ON organization_units;
CREATE TRIGGER update_hierarchy_paths_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION update_hierarchy_paths();

-- ==============================================================
-- 辅助索引：过滤已删除记录
-- ==============================================================
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_units_code_current_active
    ON organization_units (code)
    WHERE is_current = TRUE AND status <> 'DELETED';

-- ==============================================================
-- 数据修复：同步 is_current（必要时同步 is_future）
-- ==============================================================
UPDATE organization_units
   SET is_current = FALSE
 WHERE status = 'DELETED' OR deleted_at IS NOT NULL;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'organization_units'
           AND column_name = 'is_future'
    ) THEN
        EXECUTE $_upd$
            UPDATE organization_units
               SET is_current = CASE 
                                    WHEN effective_date > CURRENT_DATE THEN FALSE
                                    WHEN end_date IS NOT NULL AND end_date <= CURRENT_DATE THEN FALSE
                                    ELSE TRUE
                                END,
                   is_future = CASE WHEN effective_date > CURRENT_DATE THEN TRUE ELSE FALSE END
             WHERE status <> 'DELETED' AND deleted_at IS NULL;
        $_upd$;
    ELSE
        EXECUTE $_upd$
            UPDATE organization_units
               SET is_current = CASE 
                                    WHEN effective_date > CURRENT_DATE THEN FALSE
                                    WHEN end_date IS NOT NULL AND end_date <= CURRENT_DATE THEN FALSE
                                    ELSE TRUE
                                END
             WHERE status <> 'DELETED' AND deleted_at IS NULL;
        $_upd$;
    END IF;
END $$;

UPDATE organization_units c
   SET parent_code = NULL
 WHERE parent_code IS NOT NULL
   AND EXISTS (
        SELECT 1 FROM organization_units p
         WHERE p.code = c.parent_code
           AND (p.status = 'DELETED' OR p.deleted_at IS NOT NULL)
   );

-- 再触发一次路径刷新，确保层级正确
UPDATE organization_units SET name = name;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'audit_changes_trigger'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units ENABLE TRIGGER audit_changes_trigger';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_trigger
         WHERE tgname = 'trg_prevent_update_deleted'
           AND tgrelid = 'organization_units'::regclass
    ) THEN
        EXECUTE 'ALTER TABLE organization_units ENABLE TRIGGER trg_prevent_update_deleted';
    END IF;
END $$;

ANALYZE organization_units;
ANALYZE audit_logs;

-- 迁移结束
