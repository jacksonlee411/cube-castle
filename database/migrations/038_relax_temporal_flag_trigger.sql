-- 038_relax_temporal_flag_trigger.sql
-- 目的：调整 enforce_temporal_flags 触发器，只在记录不应为当前版本时强制 is_current=false，
--       避免在状态变更（激活/停用）插入新版本时触发唯一索引冲突。

BEGIN;

CREATE OR REPLACE FUNCTION enforce_temporal_flags()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
DECLARE
    utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
BEGIN
    -- 删除状态或已标记删除的记录始终不可作为当前版本
    IF NEW.status = 'DELETED' OR NEW.deleted_at IS NOT NULL THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    -- 未来生效或已过期的记录不应标记为当前版本
    IF NEW.effective_date > utc_date THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    IF NEW.end_date IS NOT NULL AND NEW.end_date <= utc_date THEN
        NEW.is_current := FALSE;
        RETURN NEW;
    END IF;

    -- 其它情况保留调用方指定值（交由时间轴重算流程统一处理）
    RETURN NEW;
END;
$function$;

-- 保持列默认值与触发器语义一致（显式为 false，时间轴重算会更新实际当前版本）
ALTER TABLE organization_units ALTER COLUMN is_current SET DEFAULT false;

COMMIT;

SELECT 'relaxed temporal flags trigger' AS info, NOW() AS applied_at;
