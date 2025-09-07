-- 026_unify_temporal_timezone_utc.sql
-- 目的：统一时态计算使用 UTC 自然日，避免应用与数据库时区差异导致的 is_current/is_future 短暂不一致。

-- 重建 enforce_temporal_flags()，使用 UTC 日期
CREATE OR REPLACE FUNCTION enforce_temporal_flags()
RETURNS TRIGGER AS $$
DECLARE
    utc_date DATE := (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date;
BEGIN
    -- 软删除必须非当前且非未来
    IF NEW.is_deleted IS TRUE THEN
        NEW.is_current := FALSE;
        NEW.is_future := FALSE;
        RETURN NEW;
    END IF;

    -- 根据 effective/end 日期自动推导 is_current/is_future（以 UTC 为准）
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
$$ LANGUAGE plpgsql;

-- 可选：一次性规范化现有记录（基于 UTC 日期）
UPDATE organization_units
   SET is_current = CASE 
                        WHEN effective_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date THEN FALSE
                        WHEN end_date IS NOT NULL AND end_date <= (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date THEN FALSE
                        ELSE TRUE
                    END,
       is_future = CASE WHEN effective_date > (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')::date THEN TRUE ELSE FALSE END
 WHERE is_deleted = FALSE;

-- 结束

