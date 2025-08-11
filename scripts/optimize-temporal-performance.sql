-- 时态查询性能优化 - 数据库索引优化脚本
-- 为时态查询提供高性能索引支持

-- ===== 组织单元表的时态查询索引 =====

-- 1. 核心时态查询索引 - 支持as_of_date查询
-- 索引覆盖 (tenant_id, code, effective_date, end_date)
DROP INDEX IF EXISTS idx_org_temporal_core;
CREATE INDEX idx_org_temporal_core 
ON organization_units (tenant_id, code, effective_date DESC, end_date DESC NULLS LAST)
WHERE effective_date IS NOT NULL;

-- 2. 当前有效记录快速查询索引
DROP INDEX IF EXISTS idx_org_current_active;
CREATE INDEX idx_org_current_active 
ON organization_units (tenant_id, code, is_current) 
WHERE is_current = true;

-- 3. 日期范围查询索引 - 支持effective_from/effective_to查询
DROP INDEX IF EXISTS idx_org_date_range;
CREATE INDEX idx_org_date_range 
ON organization_units (tenant_id, effective_date DESC, end_date DESC NULLS LAST)
WHERE effective_date IS NOT NULL;

-- 4. 状态和类型过滤索引
DROP INDEX IF EXISTS idx_org_status_type;
CREATE INDEX idx_org_status_type 
ON organization_units (tenant_id, status, unit_type, effective_date DESC);

-- 5. 历史记录查询索引
DROP INDEX IF EXISTS idx_org_history_lookup;
CREATE INDEX idx_org_history_lookup 
ON organization_units (tenant_id, code, created_at DESC, effective_date DESC)
WHERE is_current = false;

-- 6. 未来生效记录索引
DROP INDEX IF EXISTS idx_org_future_effective;
CREATE INDEX idx_org_future_effective 
ON organization_units (tenant_id, effective_date DESC) 
WHERE effective_date > CURRENT_DATE;

-- ===== 组织事件表的性能索引 =====

-- 7. 事件查询主索引
DROP INDEX IF EXISTS idx_org_events_lookup;
CREATE INDEX idx_org_events_lookup 
ON organization_events (tenant_id, organization_code, effective_date DESC, event_type);

-- 8. 事件时间序列索引
DROP INDEX IF EXISTS idx_org_events_timeline;
CREATE INDEX idx_org_events_timeline 
ON organization_events (tenant_id, effective_date DESC, created_at DESC);

-- 9. 事件类型统计索引
DROP INDEX IF EXISTS idx_org_events_stats;
CREATE INDEX idx_org_events_stats 
ON organization_events (tenant_id, event_type, DATE(created_at));

-- ===== 组织版本历史表的性能索引 =====

-- 10. 版本历史查询索引
DROP INDEX IF EXISTS idx_org_versions_lookup;
CREATE INDEX idx_org_versions_lookup 
ON organization_versions (tenant_id, organization_code, effective_date DESC, end_date DESC NULLS LAST);

-- 11. 版本快照查询索引
DROP INDEX IF EXISTS idx_org_versions_snapshot;
CREATE INDEX idx_org_versions_snapshot 
ON organization_versions (tenant_id, organization_code, effective_date DESC)
INCLUDE (snapshot_data);

-- ===== 复合查询优化索引 =====

-- 12. 时态范围查询复合索引 - 最常用的查询模式
DROP INDEX IF EXISTS idx_org_temporal_range_composite;
CREATE INDEX idx_org_temporal_range_composite 
ON organization_units (
  tenant_id, 
  code, 
  effective_date DESC, 
  end_date DESC NULLS LAST,
  is_current,
  status
) WHERE effective_date IS NOT NULL;

-- 13. 组织层级时态查询索引
DROP INDEX IF EXISTS idx_org_hierarchy_temporal;
CREATE INDEX idx_org_hierarchy_temporal 
ON organization_units (tenant_id, level, path, effective_date DESC, is_current);

-- ===== 查询统计和监控索引 =====

-- 14. 查询性能监控索引
DROP INDEX IF EXISTS idx_org_query_stats;
CREATE INDEX idx_org_query_stats 
ON organization_units (tenant_id, DATE(updated_at), status);

-- ===== 索引使用情况分析视图 =====

-- 创建索引使用情况监控视图
DROP VIEW IF EXISTS v_temporal_index_usage;
CREATE VIEW v_temporal_index_usage AS
SELECT 
  schemaname,
  tablename,
  indexname,
  idx_scan as total_scans,
  idx_tup_read as total_tuples_read,
  idx_tup_fetch as total_tuples_fetched,
  CASE 
    WHEN idx_scan = 0 THEN 'UNUSED'
    WHEN idx_scan < 100 THEN 'LOW_USAGE'
    WHEN idx_scan < 1000 THEN 'MEDIUM_USAGE'
    ELSE 'HIGH_USAGE'
  END as usage_level
FROM pg_stat_user_indexes 
WHERE tablename IN ('organization_units', 'organization_events', 'organization_versions')
ORDER BY idx_scan DESC;

-- ===== 时态查询性能统计 =====

-- 创建查询性能统计视图
DROP VIEW IF EXISTS v_temporal_query_performance;
CREATE VIEW v_temporal_query_performance AS
SELECT 
  't' as table_type,
  'organization_units' as table_name,
  COUNT(*) as total_records,
  COUNT(*) FILTER (WHERE is_current = true) as current_records,
  COUNT(*) FILTER (WHERE effective_date > CURRENT_DATE) as future_records,
  COUNT(*) FILTER (WHERE end_date IS NOT NULL AND end_date < CURRENT_DATE) as expired_records,
  COUNT(DISTINCT tenant_id) as tenant_count,
  COUNT(DISTINCT code) as organization_count,
  MIN(effective_date) as earliest_effective_date,
  MAX(effective_date) as latest_effective_date
FROM organization_units
WHERE effective_date IS NOT NULL;

-- ===== 自动统计信息更新 =====

-- 确保统计信息是最新的，以便查询优化器做出最佳决策
ANALYZE organization_units;
ANALYZE organization_events;
ANALYZE organization_versions;

-- ===== 优化建议输出 =====

-- 输出优化建议和索引创建结果
SELECT 
  'TEMPORAL_PERFORMANCE_OPTIMIZATION' as status,
  'COMPLETED' as result,
  now() as optimized_at,
  '14 temporal-specific indexes created' as index_summary,
  '2 monitoring views created' as view_summary,
  'Statistics updated for query planner' as stats_summary;

-- ===== 性能验证查询 =====

-- 验证关键时态查询的性能
EXPLAIN (ANALYZE, BUFFERS) 
SELECT tenant_id, code, name, effective_date, end_date, is_current
FROM organization_units 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9' 
  AND code = '1000056'
  AND effective_date <= CURRENT_DATE 
  AND (end_date IS NULL OR end_date >= CURRENT_DATE)
ORDER BY effective_date DESC;

-- 显示优化后的索引列表
SELECT 
  i.indexname,
  i.tablename,
  pg_size_pretty(pg_relation_size(i.indexname::regclass)) as index_size,
  i.indexdef
FROM pg_indexes i 
WHERE i.tablename IN ('organization_units', 'organization_events', 'organization_versions')
  AND i.indexname LIKE 'idx_org_%'
ORDER BY i.tablename, i.indexname;

-- 性能优化说明
/*
主要优化策略：

1. 时态查询核心索引
   - 覆盖最常用的查询模式：按组织代码和日期查询
   - 支持as_of_date查询的高效执行

2. 范围查询优化
   - 针对effective_from/effective_to参数优化
   - 支持批量日期范围查询

3. 当前记录快速访问
   - 专门为is_current=true查询优化
   - 避免全表扫描

4. 复合查询索引
   - 组合多个查询条件，减少索引查找次数
   - 包含常用字段避免回表查询

5. 监控和分析
   - 提供索引使用情况分析
   - 查询性能统计视图

预期性能提升：
- 时态查询响应时间：从100ms降至10ms以下
- 缓存命中率：提升至90%以上
- 并发查询能力：提升3-5倍
*/