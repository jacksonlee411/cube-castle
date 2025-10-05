-- 026_unify_temporal_timezone_utc.sql (compatibility rewrite)
-- 目的：统一 is_current 推导使用 UTC 日期，并在存在 is_future 时同步维护。

DO $$
DECLARE
    has_is_future BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'organization_units'
           AND column_name = 'is_future'
    ) INTO has_is_future;

    IF has_is_future THEN
        EXECUTE $_func$
            CREATE OR REPLACE FUNCTION enforce_temporal_flags()
            RETURNS TRIGGER AS $_body$
            DECLARE
                utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
            BEGIN
                IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
                    NEW.is_current := FALSE;
                    NEW.is_future := FALSE;
                    RETURN NEW;
                END IF;

                IF NEW.effective_date > utc_date THEN
                    NEW.is_current := FALSE;
                    NEW.is_future := TRUE;
                ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= utc_date THEN
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
            DECLARE
                utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
            BEGIN
                IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
                    NEW.is_current := FALSE;
                    RETURN NEW;
                END IF;

                IF NEW.effective_date > utc_date THEN
                    NEW.is_current := FALSE;
                ELSIF NEW.end_date IS NOT NULL AND NEW.end_date <= utc_date THEN
                    NEW.is_current := FALSE;
                ELSE
                    NEW.is_current := TRUE;
                END IF;
                RETURN NEW;
            END;
            $_body$ LANGUAGE plpgsql;
        $_func$;
    END IF;

    -- 重新计算现有数据
    IF has_is_future THEN
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
        EXECUTE '
            UPDATE organization_units
               SET is_current = CASE
                                    WHEN effective_date > (CURRENT_TIMESTAMP AT TIME ZONE ''UTC'')::date THEN FALSE
                                    WHEN end_date IS NOT NULL AND end_date <= (CURRENT_TIMESTAMP AT TIME ZONE ''UTC'')::date THEN FALSE
                                    ELSE TRUE
                                END,
                   is_future = CASE WHEN effective_date > (CURRENT_TIMESTAMP AT TIME ZONE ''UTC'')::date THEN TRUE ELSE FALSE END
             WHERE deleted_at IS NULL OR status <> ''DELETED'';
        ';
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
    ELSE
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
        EXECUTE '
            UPDATE organization_units
               SET is_current = CASE
                                    WHEN effective_date > (CURRENT_TIMESTAMP AT TIME ZONE ''UTC'')::date THEN FALSE
                                    WHEN end_date IS NOT NULL AND end_date <= (CURRENT_TIMESTAMP AT TIME ZONE ''UTC'')::date THEN FALSE
                                    ELSE TRUE
                                END
             WHERE deleted_at IS NULL OR status <> ''DELETED'';
        ';
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
    END IF;
END$$;

-- 函数更新后需重新绑定触发器
DROP TRIGGER IF EXISTS enforce_temporal_flags_trigger ON organization_units;
CREATE TRIGGER enforce_temporal_flags_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION enforce_temporal_flags();
