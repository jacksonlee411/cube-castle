// internal/service/temporal_metrics_collector.go
package service

import (
	"fmt"
	"sync"
	"time"
)

// TemporalMetricsCollector collects and tracks performance metrics for temporal queries
type TemporalMetricsCollector struct {
	mutex             sync.RWMutex
	queryCount        int64
	totalExecutionTime time.Duration
	totalRecordsReturned int64
	totalRecordsScanned  int64
	queryTypeStats    map[string]*QueryTypeStats
	performanceHistory []QueryPerformanceData
	maxHistorySize    int
}

// QueryTypeStats tracks statistics for specific query types
type QueryTypeStats struct {
	Count              int64         `json:"count"`
	TotalExecutionTime time.Duration `json:"total_execution_time"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	MinExecutionTime   time.Duration `json:"min_execution_time"`
	MaxExecutionTime   time.Duration `json:"max_execution_time"`
	TotalRecordsReturned int64       `json:"total_records_returned"`
	TotalRecordsScanned  int64       `json:"total_records_scanned"`
	ErrorCount         int64         `json:"error_count"`
	CacheHitRate       float64       `json:"cache_hit_rate"`
}

// QueryPerformanceData represents performance data for a single query
type QueryPerformanceData struct {
	Timestamp       time.Time     `json:"timestamp"`
	QueryType       string        `json:"query_type"`
	ExecutionTime   time.Duration `json:"execution_time"`
	RecordsReturned int           `json:"records_returned"`
	RecordsScanned  int           `json:"records_scanned"`
	CacheHit        bool          `json:"cache_hit"`
	IndexesUsed     []string      `json:"indexes_used"`
}

// TemporalMetricsSummary provides a summary of temporal query metrics
type TemporalMetricsSummary struct {
	TotalQueries         int64                    `json:"total_queries"`
	AverageExecutionTime time.Duration            `json:"average_execution_time"`
	TotalRecordsReturned int64                    `json:"total_records_returned"`
	TotalRecordsScanned  int64                    `json:"total_records_scanned"`
	QueryTypeBreakdown   map[string]*QueryTypeStats `json:"query_type_breakdown"`
	PerformanceTrends    []QueryPerformanceData   `json:"performance_trends"`
	GeneratedAt          time.Time                `json:"generated_at"`
}

// NewTemporalMetricsCollector creates a new metrics collector
func NewTemporalMetricsCollector() *TemporalMetricsCollector {
	return &TemporalMetricsCollector{
		queryTypeStats:     make(map[string]*QueryTypeStats),
		performanceHistory: make([]QueryPerformanceData, 0),
		maxHistorySize:     1000, // Keep last 1000 queries
	}
}

// RecordQuery records metrics for a query execution
func (c *TemporalMetricsCollector) RecordQuery(executionTime time.Duration, recordsReturned, recordsScanned int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.queryCount++
	c.totalExecutionTime += executionTime
	c.totalRecordsReturned += int64(recordsReturned)
	c.totalRecordsScanned += int64(recordsScanned)

	// Add to performance history
	perfData := QueryPerformanceData{
		Timestamp:       time.Now(),
		QueryType:       "timeline", // Default for now
		ExecutionTime:   executionTime,
		RecordsReturned: recordsReturned,
		RecordsScanned:  recordsScanned,
		CacheHit:        false, // TODO: Add cache hit detection
	}

	c.performanceHistory = append(c.performanceHistory, perfData)

	// Trim history if too large
	if len(c.performanceHistory) > c.maxHistorySize {
		c.performanceHistory = c.performanceHistory[1:]
	}
}

// RecordQueryByType records metrics for a specific query type
func (c *TemporalMetricsCollector) RecordQueryByType(queryType string, executionTime time.Duration, recordsReturned, recordsScanned int, cacheHit bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Update overall metrics
	c.queryCount++
	c.totalExecutionTime += executionTime
	c.totalRecordsReturned += int64(recordsReturned)
	c.totalRecordsScanned += int64(recordsScanned)

	// Update query type specific metrics
	stats, exists := c.queryTypeStats[queryType]
	if !exists {
		stats = &QueryTypeStats{
			MinExecutionTime: executionTime,
			MaxExecutionTime: executionTime,
		}
		c.queryTypeStats[queryType] = stats
	}

	stats.Count++
	stats.TotalExecutionTime += executionTime
	stats.AverageExecutionTime = time.Duration(int64(stats.TotalExecutionTime) / stats.Count)
	stats.TotalRecordsReturned += int64(recordsReturned)
	stats.TotalRecordsScanned += int64(recordsScanned)

	if executionTime < stats.MinExecutionTime {
		stats.MinExecutionTime = executionTime
	}
	if executionTime > stats.MaxExecutionTime {
		stats.MaxExecutionTime = executionTime
	}

	// Update cache hit rate
	if cacheHit {
		stats.CacheHitRate = (stats.CacheHitRate*float64(stats.Count-1) + 1.0) / float64(stats.Count)
	} else {
		stats.CacheHitRate = (stats.CacheHitRate * float64(stats.Count-1)) / float64(stats.Count)
	}

	// Add to performance history
	perfData := QueryPerformanceData{
		Timestamp:       time.Now(),
		QueryType:       queryType,
		ExecutionTime:   executionTime,
		RecordsReturned: recordsReturned,
		RecordsScanned:  recordsScanned,
		CacheHit:        cacheHit,
	}

	c.performanceHistory = append(c.performanceHistory, perfData)

	// Trim history if too large
	if len(c.performanceHistory) > c.maxHistorySize {
		c.performanceHistory = c.performanceHistory[1:]
	}
}

// RecordError records an error for a specific query type
func (c *TemporalMetricsCollector) RecordError(queryType string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	stats, exists := c.queryTypeStats[queryType]
	if !exists {
		stats = &QueryTypeStats{}
		c.queryTypeStats[queryType] = stats
	}

	stats.ErrorCount++
}

// GetSummary returns a summary of all collected metrics
func (c *TemporalMetricsCollector) GetSummary() *TemporalMetricsSummary {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var averageExecutionTime time.Duration
	if c.queryCount > 0 {
		averageExecutionTime = time.Duration(int64(c.totalExecutionTime) / c.queryCount)
	}

	// Create a copy of query type stats
	queryTypeBreakdown := make(map[string]*QueryTypeStats)
	for queryType, stats := range c.queryTypeStats {
		queryTypeBreakdown[queryType] = &QueryTypeStats{
			Count:                stats.Count,
			TotalExecutionTime:   stats.TotalExecutionTime,
			AverageExecutionTime: stats.AverageExecutionTime,
			MinExecutionTime:     stats.MinExecutionTime,
			MaxExecutionTime:     stats.MaxExecutionTime,
			TotalRecordsReturned: stats.TotalRecordsReturned,
			TotalRecordsScanned:  stats.TotalRecordsScanned,
			ErrorCount:           stats.ErrorCount,
			CacheHitRate:         stats.CacheHitRate,
		}
	}

	// Create a copy of performance history
	performanceTrends := make([]QueryPerformanceData, len(c.performanceHistory))
	copy(performanceTrends, c.performanceHistory)

	return &TemporalMetricsSummary{
		TotalQueries:         c.queryCount,
		AverageExecutionTime: averageExecutionTime,
		TotalRecordsReturned: c.totalRecordsReturned,
		TotalRecordsScanned:  c.totalRecordsScanned,
		QueryTypeBreakdown:   queryTypeBreakdown,
		PerformanceTrends:    performanceTrends,
		GeneratedAt:          time.Now(),
	}
}

// GetQueryTypeStats returns statistics for a specific query type
func (c *TemporalMetricsCollector) GetQueryTypeStats(queryType string) *QueryTypeStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats, exists := c.queryTypeStats[queryType]
	if !exists {
		return &QueryTypeStats{}
	}

	// Return a copy to prevent external modification
	return &QueryTypeStats{
		Count:                stats.Count,
		TotalExecutionTime:   stats.TotalExecutionTime,
		AverageExecutionTime: stats.AverageExecutionTime,
		MinExecutionTime:     stats.MinExecutionTime,
		MaxExecutionTime:     stats.MaxExecutionTime,
		TotalRecordsReturned: stats.TotalRecordsReturned,
		TotalRecordsScanned:  stats.TotalRecordsScanned,
		ErrorCount:           stats.ErrorCount,
		CacheHitRate:         stats.CacheHitRate,
	}
}

// GetRecentPerformance returns performance data for recent queries
func (c *TemporalMetricsCollector) GetRecentPerformance(limit int) []QueryPerformanceData {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if limit <= 0 || limit > len(c.performanceHistory) {
		limit = len(c.performanceHistory)
	}

	// Return the last 'limit' entries
	start := len(c.performanceHistory) - limit
	result := make([]QueryPerformanceData, limit)
	copy(result, c.performanceHistory[start:])

	return result
}

// Reset clears all collected metrics
func (c *TemporalMetricsCollector) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.queryCount = 0
	c.totalExecutionTime = 0
	c.totalRecordsReturned = 0
	c.totalRecordsScanned = 0
	c.queryTypeStats = make(map[string]*QueryTypeStats)
	c.performanceHistory = make([]QueryPerformanceData, 0)
}

// GetHealthMetrics returns health-related metrics
func (c *TemporalMetricsCollector) GetHealthMetrics() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	healthMetrics := make(map[string]interface{})

	// Calculate average execution time for recent queries
	recentQueries := c.GetRecentPerformance(100)
	var recentAvgTime time.Duration
	if len(recentQueries) > 0 {
		var totalTime time.Duration
		for _, query := range recentQueries {
			totalTime += query.ExecutionTime
		}
		recentAvgTime = time.Duration(int64(totalTime) / int64(len(recentQueries)))
	}

	healthMetrics["total_queries"] = c.queryCount
	healthMetrics["recent_average_execution_time_ms"] = recentAvgTime.Milliseconds()
	healthMetrics["total_records_returned"] = c.totalRecordsReturned
	healthMetrics["total_records_scanned"] = c.totalRecordsScanned

	// Calculate efficiency ratio
	if c.totalRecordsScanned > 0 {
		healthMetrics["efficiency_ratio"] = float64(c.totalRecordsReturned) / float64(c.totalRecordsScanned)
	} else {
		healthMetrics["efficiency_ratio"] = 1.0
	}

	// Calculate error rates by query type
	errorRates := make(map[string]float64)
	for queryType, stats := range c.queryTypeStats {
		if stats.Count > 0 {
			errorRates[queryType] = float64(stats.ErrorCount) / float64(stats.Count)
		}
	}
	healthMetrics["error_rates"] = errorRates

	return healthMetrics
}

// GetPerformanceAlerts returns alerts based on performance thresholds
func (c *TemporalMetricsCollector) GetPerformanceAlerts() []PerformanceAlert {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var alerts []PerformanceAlert

	// Check for slow queries (>1 second average)
	for queryType, stats := range c.queryTypeStats {
		if stats.AverageExecutionTime > time.Second {
			alerts = append(alerts, PerformanceAlert{
				AlertType: "SLOW_QUERY",
				Severity:  "WARNING",
				Message:   fmt.Sprintf("Query type '%s' has slow average execution time: %v", queryType, stats.AverageExecutionTime),
				QueryType: queryType,
				Threshold: time.Second,
				ActualValue: stats.AverageExecutionTime,
			})
		}
	}

	// Check for high error rates (>5%)
	for queryType, stats := range c.queryTypeStats {
		if stats.Count > 0 {
			errorRate := float64(stats.ErrorCount) / float64(stats.Count)
			if errorRate > 0.05 {
				alerts = append(alerts, PerformanceAlert{
					AlertType: "HIGH_ERROR_RATE",
					Severity:  "ERROR",
					Message:   fmt.Sprintf("Query type '%s' has high error rate: %.2f%%", queryType, errorRate*100),
					QueryType: queryType,
					Threshold: 0.05,
					ActualValue: errorRate,
				})
			}
		}
	}

	// Check for low efficiency (scan-to-return ratio > 10:1)
	if c.totalRecordsScanned > 0 {
		scanRatio := float64(c.totalRecordsScanned) / float64(c.totalRecordsReturned)
		if scanRatio > 10.0 {
			alerts = append(alerts, PerformanceAlert{
				AlertType: "LOW_EFFICIENCY",
				Severity:  "WARNING",
				Message:   fmt.Sprintf("Overall query efficiency is low. Scan-to-return ratio: %.2f:1", scanRatio),
				Threshold: 10.0,
				ActualValue: scanRatio,
			})
		}
	}

	return alerts
}

// PerformanceAlert represents a performance-related alert
type PerformanceAlert struct {
	AlertType   string      `json:"alert_type"`
	Severity    string      `json:"severity"` // INFO, WARNING, ERROR, CRITICAL
	Message     string      `json:"message"`
	QueryType   string      `json:"query_type,omitempty"`
	Threshold   interface{} `json:"threshold"`
	ActualValue interface{} `json:"actual_value"`
	GeneratedAt time.Time   `json:"generated_at"`
}