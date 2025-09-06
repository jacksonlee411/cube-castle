-- 019_prevent_update_deleted.sql
-- 确保已软删除（status=DELETED 或 deleted_at 非空）的记录为只读：禁止UPDATE

CREATE OR REPLACE FUNCTION prevent_update_deleted()
RETURNS TRIGGER AS $$
BEGIN
    -- 当旧记录已是删除状态时，阻止任何更新
    IF (OLD.status = 'DELETED' OR OLD.deleted_at IS NOT NULL) THEN
        RAISE EXCEPTION 'READ_ONLY_DELETED: cannot modify deleted record %', OLD.record_id
            USING ERRCODE = '55000'; -- object_not_in_prerequisite_state
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_update_deleted ON organization_units;
CREATE TRIGGER trg_prevent_update_deleted
    BEFORE UPDATE ON organization_units
    FOR EACH ROW
    WHEN (OLD.status = 'DELETED' OR OLD.deleted_at IS NOT NULL)
    EXECUTE FUNCTION prevent_update_deleted();

