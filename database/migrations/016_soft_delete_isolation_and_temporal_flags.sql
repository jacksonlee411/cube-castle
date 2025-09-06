-- 016_soft_delete_isolation_and_temporal_flags.sql
-- 目的: 
--  1) 软删除不参与层级/当前计算
--  2) 自动规范化 is_current/is_future 与 soft-delete 联动
--  3) 插入时禁止引用已删除或非当前的父节点
--  4) 数据修复: 纠正历史数据，重建层级路径

-- 注意: 包含 CREATE INDEX CONCURRENTLY 语句，需在事务外执行。

-- 1. 约束: 软删不可为当前
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conrelid = 'organization_units'::regclass
          AND conname = 'chk_org_units_not_deleted_current'
    ) THEN
        ALTER TABLE organization_units
        ADD CONSTRAINT chk_org_units_not_deleted_current
        CHECK (NOT (is_deleted AND is_current));
    END IF;
END $$;

-- 2. 触发器: 父节点有效性 (仅插入)
CREATE OR REPLACE FUNCTION validate_parent_available()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_code IS NOT NULL THEN
        PERFORM 1 FROM organization_units p
         WHERE p.code = NEW.parent_code
           AND p.is_current = true
           AND p.is_deleted = false
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

-- 3. 触发器: 时态标志自动规范化
CREATE OR REPLACE FUNCTION enforce_temporal_flags()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_deleted IS TRUE THEN
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
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS enforce_temporal_flags_trigger ON organization_units;
CREATE TRIGGER enforce_temporal_flags_trigger
    BEFORE INSERT OR UPDATE ON organization_units
    FOR EACH ROW
    EXECUTE FUNCTION enforce_temporal_flags();

-- 4. 触发器: 层级路径函数仅使用未删除且当前的父节点，并在缺失时降级为根
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
          AND parent.is_current = true
          AND parent.is_deleted = false
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

-- 5. 索引: 加速父节点有效性与当前未删除节点查找
-- 注意: CONCURRENTLY 需在事务外执行
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_units_code_current_active 
    ON organization_units (code) WHERE is_current = true AND is_deleted = false;

-- 6. 数据修复
-- 6.1 已软删的记录不应为当前或未来
UPDATE organization_units
   SET is_current = FALSE,
       is_future = FALSE
 WHERE is_deleted = TRUE
   AND (is_current = TRUE OR is_future = TRUE);

-- 6.2 规范化未删除记录的时态标志（按当前日期推导）
UPDATE organization_units
   SET is_current = CASE 
                        WHEN effective_date > CURRENT_DATE THEN FALSE
                        WHEN end_date IS NOT NULL AND end_date <= CURRENT_DATE THEN FALSE
                        ELSE TRUE
                    END,
       is_future = CASE WHEN effective_date > CURRENT_DATE THEN TRUE ELSE FALSE END
 WHERE is_deleted = FALSE;

-- 6.3 断开已删除父节点的子节点引用，避免路径计算被干扰
UPDATE organization_units c
   SET parent_code = NULL
 WHERE parent_code IS NOT NULL
   AND EXISTS (
        SELECT 1 FROM organization_units p
         WHERE p.code = c.parent_code AND p.is_deleted = TRUE
   );

-- 6.4 触发层级路径与级别重建
UPDATE organization_units SET name = name;

-- 6.5 统计与分析
ANALYZE organization_units;
ANALYZE audit_logs;

-- 迁移结束

