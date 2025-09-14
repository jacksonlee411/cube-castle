-- fix-temporal-overlap-and-current.sql
-- 目的：一次性修复 organization_units 的时间轴重叠与“多个当前”问题。
-- 策略：
--  1) 仅处理 data_status='NORMAL' 的版本（若无该列则忽略条件）
--  2) 按 (tenant_id, code, effective_date ASC) 排序，计算新 end_date = lead(effective_date) - 1 天，最后一条为 NULL。
--  3) 仅当值变化时更新 end_date，避免空 UPDATE。
--  4) 依据今天日期重算 is_current：对每个 (tenant_id, code) 选择 effective_date<=today 且 end_date为空或>today 的“最新一条”为 TRUE，其余为 FALSE。
-- 安全：
--  - 使用临时表与窗口函数；
--  - 输出修复前后差异计数。

BEGIN;

-- 规范化：构建工作集（仅正常数据；若不存在 data_status 列则降级为全量）
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
     WHERE table_name='organization_units' AND column_name='data_status'
  ) THEN
    CREATE TEMP TABLE _org_work AS
    SELECT record_id, tenant_id, code, effective_date, end_date
      FROM organization_units;
  ELSE
    CREATE TEMP TABLE _org_work AS
    SELECT record_id, tenant_id, code, effective_date, end_date
      FROM organization_units
     WHERE data_status = 'NORMAL';
  END IF;
END $$;

-- 计算新 end_date
CREATE TEMP TABLE _recalc AS
SELECT w.record_id,
       w.tenant_id,
       w.code,
       w.effective_date,
       w.end_date                       AS old_end_date,
       (LEAD(w.effective_date) OVER (PARTITION BY w.tenant_id, w.code ORDER BY w.effective_date)
          - INTERVAL '1 day')::date     AS new_end_date
  FROM _org_work w
 ORDER BY w.tenant_id, w.code, w.effective_date;

-- 将分组中的最后一条的 new_end_date 置为 NULL（LEAD 为 NULL 的情况本已为 NULL）
-- 实际无需额外处理，因为 (NULL - interval) 为 NULL；此处仅确保为 date NULL。

-- 统计即将发生的变更
WITH diff AS (
  SELECT r.*
    FROM _recalc r
   WHERE r.old_end_date IS DISTINCT FROM r.new_end_date
)
SELECT 'TO_UPDATE_ENDDATE' AS tag, COUNT(*) AS cnt FROM diff;

UPDATE organization_units u
   SET end_date = r.new_end_date
  FROM _recalc r
 WHERE u.record_id = r.record_id
   AND COALESCE(u.status,'ACTIVE') <> 'DELETED'
   AND u.end_date IS DISTINCT FROM r.new_end_date;

-- 重算 is_current：先全部置 FALSE，再按规则置 TRUE
UPDATE organization_units SET is_current = FALSE WHERE COALESCE(status,'ACTIVE') <> 'DELETED';

WITH eligible AS (
  SELECT u.*
    FROM organization_units u
   WHERE (u.end_date IS NULL OR u.end_date > CURRENT_DATE)
     AND u.effective_date <= CURRENT_DATE
), picked AS (
  SELECT e.record_id
    FROM (
      SELECT e.*, ROW_NUMBER() OVER (PARTITION BY e.tenant_id, e.code ORDER BY e.effective_date DESC) AS rn
        FROM eligible e
    ) e
   WHERE e.rn = 1
)
UPDATE organization_units u
   SET is_current = TRUE
  FROM picked p
 WHERE u.record_id = p.record_id
   AND COALESCE(u.status,'ACTIVE') <> 'DELETED';

-- 输出结果统计
SELECT 'CURRENT_COUNT_PER_CODE' AS tag, code, COUNT(*) AS current_count
  FROM organization_units
 WHERE is_current = TRUE
 GROUP BY code
HAVING COUNT(*) > 1
 ORDER BY code;

COMMIT;

-- 使用方法：
-- psql "$DATABASE_URL" -f scripts/fix-temporal-overlap-and-current.sql
