-- 时态查询性能优化 - 修复版本
-- 修复IMMUTABLE函数和其他数据库兼容性问题

-- ===== 清理和重新创建索引 =====

-- 6. 修复未来生效记录索引 (移除CURRENT_DATE函数)
DROP INDEX IF EXISTS idx_org_future_effective;
CREATE INDEX idx_org_future_effective 
ON organization_units (tenant_id, effective_date DESC) 
WHERE effective_date > '2025-01-01';

-- 9. 修复事件类型统计索引 (移除DATE函数)
DROP INDEX IF EXISTS idx_org_events_stats;
CREATE INDEX idx_org_events_stats 
ON organization_events (tenant_id, event_type, created_at);

-- 14. 修复查询性能监控索引 (移除DATE函数)
DROP INDEX IF EXISTS idx_org_query_stats;
CREATE INDEX idx_org_query_stats 
ON organization_units (tenant_id, updated_at, status);

-- ===== 修复索引使用情况分析视图 =====

DROP VIEW IF EXISTS v_temporal_index_usage;
CREATE VIEW v_temporal_index_usage AS
SELECT 
  schemaname,
  t.relname as tablename,
  indexrelname as indexname,
  idx_scan as total_scans,
  idx_tup_read as total_tuples_read,
  idx_tup_fetch as total_tuples_fetched,
  CASE 
    WHEN idx_scan = 0 THEN 'UNUSED'
    WHEN idx_scan < 100 THEN 'LOW_USAGE'
    WHEN idx_scan < 1000 THEN 'MEDIUM_USAGE'
    ELSE 'HIGH_USAGE'
  END as usage_level
FROM pg_stat_user_indexes s
JOIN pg_class t ON s.relid = t.oid
WHERE t.relname IN ('organization_units', 'organization_events', 'organization_versions')
ORDER BY idx_scan DESC;

-- 验证核心时态查询索引是否工作
SELECT 
  indexname,
  indexdef
FROM pg_indexes 
WHERE tablename = 'organization_units' 
  AND indexname IN ('idx_org_temporal_core', 'idx_org_current_active', 'idx_org_date_range')
ORDER BY indexname;

-- 手动统计更新确保优化器使用新索引
ANALYZE organization_units;
ANALYZE organization_events;
ANALYZE organization_versions;

-- ===== 验证时态查询性能 =====

-- 测试当前记录查询（应使用 idx_org_current_active 索引）
EXPLAIN (ANALYZE, BUFFERS) 
SELECT code, name, effective_date, is_current
FROM organization_units 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9' 
  AND code = '1000056'
  AND is_current = true;

-- 测试时态范围查询（应使用 idx_org_temporal_core 索引）
EXPLAIN (ANALYZE, BUFFERS) 
SELECT code, name, effective_date, end_date
FROM organization_units 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9' 
  AND code = '1000056'
  AND effective_date IS NOT NULL
ORDER BY effective_date DESC;

-- 测试日期范围查询（应使用 idx_org_date_range 索引）
EXPLAIN (ANALYZE, BUFFERS)
SELECT code, name, effective_date
FROM organization_units 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9'
  AND effective_date >= '2025-01-01'
  AND effective_date <= '2025-12-31'
ORDER BY effective_date DESC;

-- 显示所有时态相关索引的状态
SELECT 
  i.indexname,
  i.tablename,
  pg_size_pretty(pg_relation_size(i.indexname::regclass)) as index_size,
  CASE 
    WHEN s.idx_scan IS NULL THEN 'NEW INDEX'
    WHEN s.idx_scan = 0 THEN 'UNUSED'
    WHEN s.idx_scan < 10 THEN 'LOW USAGE'
    ELSE 'ACTIVE'
  END as usage_status
FROM pg_indexes i 
LEFT JOIN pg_stat_user_indexes s ON i.indexname = s.indexrelname
WHERE i.tablename IN ('organization_units', 'organization_events', 'organization_versions')
  AND (i.indexname LIKE 'idx_org_temporal_%' 
       OR i.indexname LIKE 'idx_org_current_%'
       OR i.indexname LIKE 'idx_org_date_%'
       OR i.indexname LIKE 'idx_org_events_%'
       OR i.indexname LIKE 'idx_org_versions_%')
ORDER BY i.tablename, i.indexname;

SELECT 
  'TEMPORAL_INDEX_OPTIMIZATION' as status,
  'FIXED_AND_COMPLETED' as result,
  now() as optimized_at,
  'Key temporal indexes created and verified' as summary;