-- 018_switch_pk_to_record_id_and_partial_uniques.sql
-- 目标:
--  1) 将主键从 (code, effective_date) 切换为技术主键 record_id
--  2) 为非删除版本建立“部分唯一”约束，允许忽略已删除版本的唯一性
--  3) 保持“单一当前版本”约束，但仅针对未删除版本生效
--
-- 适配当前线上表结构:
--  - 删除语义: status='DELETED' 或 deleted_at 非空
--  - 现有主键: organization_units_pkey (code, effective_date)
--  - 现有唯一: uk_org_ver (tenant_id, code, effective_date)
--  - 现有唯一: uk_org_current (tenant_id, code) WHERE is_current=true
--
-- 注意:
--  - 带 CONCURRENTLY 的索引需在事务外执行
--  - DROP/ADD CONSTRAINT 需短暂表锁, 建议在维护窗口执行

-- ===== 0) 运行前自检(只读) =====
-- 是否存在非删除版本在 (tenant_id, code, effective_date) 上冲突(不应有):
-- SELECT tenant_id, code, effective_date, COUNT(*)
-- FROM organization_units
-- WHERE status <> 'DELETED' AND deleted_at IS NULL
-- GROUP BY tenant_id, code, effective_date HAVING COUNT(*) > 1;

-- ===== 1) 预创建新约束所需索引 (在线, 并发) =====
-- 非删除版本的“时间点唯一” (tenant_id, code, effective_date)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uk_org_ver_active_only
ON organization_units (tenant_id, code, effective_date)
WHERE status <> 'DELETED' AND deleted_at IS NULL;

-- 非删除版本的“单一当前” (tenant_id, code) WHERE is_current=true
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uk_org_current_active_only
ON organization_units (tenant_id, code)
WHERE is_current = TRUE AND status <> 'DELETED' AND deleted_at IS NULL;

-- record_id 唯一索引 (若无)
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uidx_org_record_id
ON organization_units (record_id);

-- ===== 2) 切换主键到 record_id (维护窗口内执行) =====
DO $$
BEGIN
  -- 删除旧主键 (code, effective_date)
  IF EXISTS (
    SELECT 1 FROM pg_constraint 
    WHERE conrelid = 'organization_units'::regclass
      AND contype = 'p'
      AND conname = 'organization_units_pkey'
  ) THEN
    ALTER TABLE organization_units DROP CONSTRAINT organization_units_pkey;
  END IF;

  -- 使用已建唯一索引设为新主键
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint 
    WHERE conrelid = 'organization_units'::regclass
      AND contype = 'p'
  ) THEN
    ALTER TABLE organization_units
      ADD CONSTRAINT pk_org_record_id PRIMARY KEY USING INDEX uidx_org_record_id;
  END IF;
END $$;

-- ===== 3) 清理/替换旧唯一索引 =====
DO $$
BEGIN
  -- 删除旧版本唯一 (全量), 保留新“部分唯一”替代
  IF EXISTS (
    SELECT 1 FROM pg_indexes WHERE tablename='organization_units' AND indexname='uk_org_ver'
  ) THEN
    DROP INDEX uk_org_ver;
  END IF;

  -- 删除旧“单一当前”唯一 (未区分删除), 保留新“部分唯一”替代
  IF EXISTS (
    SELECT 1 FROM pg_indexes WHERE tablename='organization_units' AND indexname='uk_org_current'
  ) THEN
    DROP INDEX uk_org_current;
  END IF;
END $$;

-- ===== 4) 收尾: 统计信息 =====
ANALYZE organization_units;

-- ===== 5) 运行后验证 =====
-- 1) 主键检查: 
--   SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint 
--    WHERE conrelid='organization_units'::regclass AND contype='p';
-- 2) 约束验证:
--   -- 非删除版本，同一时间点不得重复 (应=0)
--   SELECT COUNT(*) FROM (
--     SELECT tenant_id, code, effective_date FROM organization_units
--      WHERE status <> 'DELETED' AND deleted_at IS NULL
--      GROUP BY 1,2,3 HAVING COUNT(*)>1
--   ) t;
--   -- 非删除版本，每个 code 仅一个当前 (应=0)
--   SELECT COUNT(*) FROM (
--     SELECT tenant_id, code FROM organization_units
--      WHERE is_current=true AND status <> 'DELETED' AND deleted_at IS NULL
--      GROUP BY 1,2 HAVING COUNT(*)>1
--   ) t;

