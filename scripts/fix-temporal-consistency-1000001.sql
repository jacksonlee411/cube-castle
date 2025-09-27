-- 时态数据一致性修复脚本 - 组织1000001
-- 目标: 清理重复记录、修复时间重叠、重新计算边界日期  
-- 重要: 只处理正常记录(status != 'DELETED')
-- 执行前请备份数据！

-- =============================================
-- 第一步: 数据备份
-- =============================================
CREATE TEMP TABLE backup_1000001 AS 
SELECT * FROM organization_units WHERE code = '1000001';

-- 显示备份记录数（区分正常和已删除）
SELECT 
    '备份完成' as message,
    COUNT(*) as total_records,
    COUNT(CASE WHEN status != 'DELETED' THEN 1 END) as active_records,
    COUNT(CASE WHEN status = 'DELETED' THEN 1 END) as deleted_records
FROM backup_1000001;

-- =============================================
-- 第二步: 问题记录分析
-- =============================================
WITH problem_analysis AS (
  SELECT 
    record_id,
    name,
    effective_date,
    end_date,
    is_current,
    created_at,
    ROW_NUMBER() OVER (PARTITION BY effective_date ORDER BY created_at DESC) as rn_by_date,
    COUNT(*) OVER (PARTITION BY effective_date) as duplicate_count
  FROM organization_units 
  WHERE code = '1000001'
)
SELECT 
  'Problem Analysis' as step,
  COUNT(CASE WHEN duplicate_count > 1 THEN 1 END) as duplicate_records,
  COUNT(CASE WHEN is_current = false AND end_date IS NULL THEN 1 END) as invalid_current,
  COUNT(*) as total_records
FROM problem_analysis;

-- =============================================
-- 第三步: 清理重复记录 (保留每个日期最新创建的记录)
-- =============================================
WITH duplicates_to_delete AS (
  SELECT record_id
  FROM (
    SELECT 
      record_id,
      ROW_NUMBER() OVER (PARTITION BY effective_date ORDER BY created_at DESC) as rn
    FROM organization_units 
    WHERE code = '1000001'
  ) ranked
  WHERE rn > 1
)
DELETE FROM organization_units 
WHERE record_id IN (SELECT record_id FROM duplicates_to_delete);

-- 显示清理结果
SELECT '重复记录清理完成' as message, 
       (SELECT COUNT(*) FROM backup_1000001) - COUNT(*) as deleted_count,
       COUNT(*) as remaining_count
FROM organization_units WHERE code = '1000001';

-- =============================================
-- 第四步: 重新计算时态边界
-- =============================================
WITH ordered_records AS (
  SELECT 
    record_id,
    effective_date,
    LEAD(effective_date) OVER (ORDER BY effective_date) as next_effective_date,
    ROW_NUMBER() OVER (ORDER BY effective_date DESC) as rn_desc
  FROM organization_units 
  WHERE code = '1000001'
  ORDER BY effective_date
)
UPDATE organization_units 
SET 
  end_date = CASE 
    -- 最新记录(当前记录)：无结束日期
    WHEN r.rn_desc = 1 THEN NULL
    -- 其他记录：结束日期为下一条记录生效日期的前一天
    ELSE (r.next_effective_date - INTERVAL '1 day')::date
  END,
  is_current = CASE 
    -- 只有最新记录是当前记录
    WHEN r.rn_desc = 1 THEN true
    ELSE false
  END,
  updated_at = NOW()
FROM ordered_records r
WHERE organization_units.record_id = r.record_id;

-- =============================================
-- 第五步: 验证修复结果
-- =============================================
SELECT 
  '修复结果验证' as step,
  code,
  name,
  TO_CHAR(effective_date, 'YYYY-MM-DD') as effective_date,
  TO_CHAR(end_date, 'YYYY-MM-DD') as end_date,
  is_current,
  CASE 
    WHEN is_current = true AND end_date IS NULL THEN '✅ 正确的当前记录'
    WHEN is_current = false AND end_date IS NOT NULL THEN '✅ 正确的历史记录'
    ELSE '❌ 仍有问题'
  END as status
FROM organization_units 
WHERE code = '1000001'
ORDER BY effective_date;

-- =============================================
-- 第六步: 时态连续性检查
-- =============================================
WITH continuity_check AS (
  SELECT 
    effective_date,
    end_date,
    LEAD(effective_date) OVER (ORDER BY effective_date) as next_effective_date,
    CASE 
      WHEN LEAD(effective_date) OVER (ORDER BY effective_date) IS NULL THEN '✅ 当前记录'
      WHEN end_date + INTERVAL '1 day' = LEAD(effective_date) OVER (ORDER BY effective_date) THEN '✅ 连续'
      ELSE '❌ 不连续'
    END as continuity_status
  FROM organization_units 
  WHERE code = '1000001'
  ORDER BY effective_date
)
SELECT 
  '时态连续性检查' as step,
  TO_CHAR(effective_date, 'YYYY-MM-DD') as from_date,
  TO_CHAR(end_date, 'YYYY-MM-DD') as to_date,
  TO_CHAR(next_effective_date, 'YYYY-MM-DD') as next_date,
  continuity_status
FROM continuity_check;

-- =============================================
-- 完成报告
-- =============================================
SELECT 
  '=== 修复完成报告 ===' as message,
  'Organization 1000001 temporal consistency fixed' as details,
  NOW() as completed_at;
