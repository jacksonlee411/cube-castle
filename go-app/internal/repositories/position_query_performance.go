package repositories

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
)

// QueryPerformanceMonitor 查询性能监控器
type QueryPerformanceMonitor struct {
	slowQueryThreshold time.Duration
	logger            Logger
	metricsCollector  MetricsCollector
}

type MetricsCollector interface {
	RecordQueryDuration(queryType string, duration time.Duration)
	RecordSlowQuery(queryType string, duration time.Duration, query string)
	IncrementQueryCount(queryType string)
}

// MonitoredPositionQueryRepository 带性能监控的职位查询仓储
type MonitoredPositionQueryRepository struct {
	base    PositionQueryRepository
	monitor *QueryPerformanceMonitor
	cache   QueryCache
}

type QueryCache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
}

func NewMonitoredPositionQueryRepository(
	base PositionQueryRepository,
	monitor *QueryPerformanceMonitor,
	cache QueryCache,
) *MonitoredPositionQueryRepository {
	return &MonitoredPositionQueryRepository{
		base:    base,
		monitor: monitor,
		cache:   cache,
	}
}

// GetPositionWithRelations 带缓存和监控的职位查询
func (r *MonitoredPositionQueryRepository) GetPositionWithRelations(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*PositionWithRelations, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.monitor.metricsCollector.RecordQueryDuration("GetPositionWithRelations", duration)
		
		if duration > r.monitor.slowQueryThreshold {
			r.monitor.logger.Warn("Slow query detected", 
				"operation", "GetPositionWithRelations",
				"duration", duration,
				"position_id", id,
				"tenant_id", tenantID)
		}
	}()

	// 尝试从缓存获取
	cacheKey := fmt.Sprintf("position_relations:%s:%s", tenantID.String(), id.String())
	if cached, found := r.cache.Get(ctx, cacheKey); found {
		r.monitor.metricsCollector.IncrementQueryCount("GetPositionWithRelations_cache_hit")
		return cached.(*PositionWithRelations), nil
	}

	// 缓存未命中，执行查询
	r.monitor.metricsCollector.IncrementQueryCount("GetPositionWithRelations_cache_miss")
	result, err := r.base.GetPositionWithRelations(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}

	// 缓存结果（5分钟TTL）
	r.cache.Set(ctx, cacheKey, result, 5*time.Minute)
	
	return result, nil
}

// SearchPositions 带分页优化的搜索
func (r *MonitoredPositionQueryRepository) SearchPositions(ctx context.Context, params SearchPositionsParams) ([]Position, int, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.monitor.metricsCollector.RecordQueryDuration("SearchPositions", duration)
	}()

	// 对大偏移量进行警告
	if params.Offset > 10000 {
		r.monitor.logger.Warn("Large offset detected in search",
			"offset", params.Offset,
			"limit", params.Limit)
	}

	return r.base.SearchPositions(ctx, params)
}

// PositionQueryOptimizer 查询优化器
type PositionQueryOptimizer struct {
	monitor *QueryPerformanceMonitor
}

// OptimizeSearchParams 优化搜索参数
func (o *PositionQueryOptimizer) OptimizeSearchParams(params *SearchPositionsParams) {
	// 限制最大返回数量
	if params.Limit > 1000 {
		params.Limit = 1000
		o.monitor.logger.Warn("Search limit capped at 1000")
	}
	
	// 对深分页进行优化建议
	if params.Offset > 5000 {
		o.monitor.logger.Warn("Deep pagination detected, consider using cursor-based pagination",
			"offset", params.Offset)
	}
}

// IndexRecommendations 索引建议
type IndexRecommendations struct {
	Neo4jIndexes     []string `json:"neo4j_indexes"`
	PostgresIndexes  []string `json:"postgres_indexes"`
	QueryOptimizations []string `json:"query_optimizations"`
}

// GetIndexRecommendations 获取索引建议
func (o *PositionQueryOptimizer) GetIndexRecommendations() IndexRecommendations {
	return IndexRecommendations{
		Neo4jIndexes: []string{
			"CREATE INDEX position_tenant_status IF NOT EXISTS FOR (p:Position) ON (p.tenant_id, p.status)",
			"CREATE INDEX position_department IF NOT EXISTS FOR (p:Position) ON (p.department_id)",
			"CREATE INDEX assignment_employee_current IF NOT EXISTS FOR ()-[a:ASSIGNED]-() ON (a.employee_id, a.is_current)",
			"CREATE INDEX assignment_position_current IF NOT EXISTS FOR ()-[a:ASSIGNED]-() ON (a.position_id, a.is_current)",
		},
		PostgresIndexes: []string{
			"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_assignments_tenant_employee_current ON position_assignments(tenant_id, employee_id, is_current) WHERE is_current = true",
			"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_position_assignments_tenant_position_current ON position_assignments(tenant_id, position_id, is_current) WHERE is_current = true",
			"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_positions_tenant_dept_status ON positions(tenant_id, department_id, status)",
		},
		QueryOptimizations: []string{
			"使用游标分页替代OFFSET/LIMIT进行深分页",
			"预计算和缓存复杂的组织架构查询结果",
			"使用读副本进行复杂的分析查询",
			"考虑使用EXPLAIN分析Neo4j查询执行计划",
		},
	}
}

// CacheInvalidationStrategy 缓存失效策略
type CacheInvalidationStrategy struct {
	cache QueryCache
}

// InvalidatePositionCache 使职位相关缓存失效
func (s *CacheInvalidationStrategy) InvalidatePositionCache(ctx context.Context, tenantID, positionID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf("position_relations:%s:%s", tenantID.String(), positionID.String()),
		fmt.Sprintf("position_hierarchy:%s:*", tenantID.String()),
		fmt.Sprintf("position_search:%s:*", tenantID.String()),
	}
	
	for _, pattern := range patterns {
		if err := s.cache.Clear(ctx, pattern); err != nil {
			return err
		}
	}
	
	return nil
}

// InvalidateEmployeePositionCache 使员工职位相关缓存失效
func (s *CacheInvalidationStrategy) InvalidateEmployeePositionCache(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	patterns := []string{
		fmt.Sprintf("employee_positions:%s:%s", tenantID.String(), employeeID.String()),
		fmt.Sprintf("position_search:%s:*", tenantID.String()),
	}
	
	for _, pattern := range patterns {
		if err := s.cache.Clear(ctx, pattern); err != nil {
			return err
		}
	}
	
	return nil
}