package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 全局Prometheus指标

var (
	// === HTTP请求指标 ===
	
	// HTTPRequestsTotal HTTP请求总数计数器
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	// HTTPRequestDuration HTTP请求持续时间直方图
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cube_castle_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)
	
	// HTTPRequestsInFlight 正在处理的HTTP请求数量
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)

	// === 业务指标 ===
	
	// EmployeesCreatedTotal 员工创建总数
	EmployeesCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_employees_created_total",
			Help: "Total number of employees created",
		},
		[]string{"tenant_id"},
	)
	
	// EmployeesUpdatedTotal 员工更新总数
	EmployeesUpdatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_employees_updated_total",
			Help: "Total number of employees updated",
		},
		[]string{"tenant_id"},
	)
	
	// EmployeesDeletedTotal 员工删除总数
	EmployeesDeletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_employees_deleted_total",
			Help: "Total number of employees deleted",
		},
		[]string{"tenant_id"},
	)
	
	// ActiveEmployeesGauge 当前活跃员工数量
	ActiveEmployeesGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cube_castle_active_employees",
			Help: "Current number of active employees",
		},
		[]string{"tenant_id"},
	)
	
	// OrganizationsCreatedTotal 组织创建总数
	OrganizationsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_organizations_created_total",
			Help: "Total number of organizations created",
		},
		[]string{"tenant_id"},
	)

	// === AI服务指标 ===
	
	// AIRequestsTotal AI请求总数
	AIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_ai_requests_total",
			Help: "Total number of AI requests",
		},
		[]string{"intent", "status"},
	)
	
	// AIRequestDuration AI请求处理时间
	AIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cube_castle_ai_request_duration_seconds",
			Help:    "AI request processing duration",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
		},
		[]string{"intent"},
	)
	
	// AISessionsActive 活跃AI会话数量
	AISessionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_ai_sessions_active",
			Help: "Number of active AI sessions",
		},
	)
	
	// AIIntentAccuracy AI意图识别准确率
	AIIntentAccuracy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cube_castle_ai_intent_accuracy",
			Help: "AI intent recognition accuracy rate",
		},
		[]string{"intent"},
	)

	// === 数据库指标 ===
	
	// DatabaseOperationsTotal 数据库操作总数
	DatabaseOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_database_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)
	
	// DatabaseOperationDuration 数据库操作持续时间
	DatabaseOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cube_castle_database_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
		[]string{"operation", "table"},
	)
	
	// DatabaseConnectionsActive 活跃数据库连接数
	DatabaseConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_database_connections_active",
			Help: "Number of active database connections",
		},
	)
	
	// DatabaseConnectionsIdle 空闲数据库连接数
	DatabaseConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_database_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	// === 发件箱指标 ===
	
	// OutboxEventsTotal 发件箱事件总数
	OutboxEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_outbox_events_total",
			Help: "Total number of outbox events",
		},
		[]string{"event_type", "status"},
	)
	
	// OutboxEventsProcessingDuration 发件箱事件处理时间
	OutboxEventsProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cube_castle_outbox_events_processing_duration_seconds",
			Help:    "Outbox events processing duration",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"event_type"},
	)
	
	// OutboxEventsPending 待处理发件箱事件数量
	OutboxEventsPending = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_outbox_events_pending",
			Help: "Number of pending outbox events",
		},
	)

	// === 系统指标 ===
	
	// ServiceUptime 服务运行时间
	ServiceUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_service_uptime_seconds",
			Help: "Service uptime in seconds",
		},
	)
	
	// MemoryUsage 内存使用量
	MemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cube_castle_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"type"}, // heap, stack, sys
	)
	
	// GoroutinesActive 活跃协程数量
	GoroutinesActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cube_castle_goroutines_active",
			Help: "Number of active goroutines",
		},
	)

	// === 错误指标 ===
	
	// ErrorsTotal 错误总数
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_errors_total",
			Help: "Total number of errors",
		},
		[]string{"component", "error_type"},
	)
	
	// PanicRecoveries 恐慌恢复总数
	PanicRecoveries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cube_castle_panic_recoveries_total",
			Help: "Total number of panic recoveries",
		},
		[]string{"component"},
	)
)

// === 业务指标记录方法 ===

// RecordEmployeeCreated 记录员工创建事件
func RecordEmployeeCreated(tenantID string) {
	EmployeesCreatedTotal.WithLabelValues(tenantID).Inc()
}

// RecordEmployeeUpdated 记录员工更新事件
func RecordEmployeeUpdated(tenantID string) {
	EmployeesUpdatedTotal.WithLabelValues(tenantID).Inc()
}

// RecordEmployeeDeleted 记录员工删除事件
func RecordEmployeeDeleted(tenantID string) {
	EmployeesDeletedTotal.WithLabelValues(tenantID).Inc()
}

// UpdateActiveEmployees 更新活跃员工数量
func UpdateActiveEmployees(tenantID string, count float64) {
	ActiveEmployeesGauge.WithLabelValues(tenantID).Set(count)
}

// RecordOrganizationCreated 记录组织创建事件
func RecordOrganizationCreated(tenantID string) {
	OrganizationsCreatedTotal.WithLabelValues(tenantID).Inc()
}

// RecordAIRequest 记录AI请求事件
func RecordAIRequest(intent, status string, duration time.Duration) {
	AIRequestsTotal.WithLabelValues(intent, status).Inc()
	AIRequestDuration.WithLabelValues(intent).Observe(duration.Seconds())
}

// UpdateAISessionsActive 更新活跃AI会话数量
func UpdateAISessionsActive(count float64) {
	AISessionsActive.Set(count)
}

// UpdateAIIntentAccuracy 更新AI意图识别准确率
func UpdateAIIntentAccuracy(intent string, accuracy float64) {
	AIIntentAccuracy.WithLabelValues(intent).Set(accuracy)
}

// RecordDatabaseOperation 记录数据库操作事件
func RecordDatabaseOperation(operation, table, status string, duration time.Duration) {
	DatabaseOperationsTotal.WithLabelValues(operation, table, status).Inc()
	DatabaseOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// UpdateDatabaseConnections 更新数据库连接数
func UpdateDatabaseConnections(active, idle int) {
	DatabaseConnectionsActive.Set(float64(active))
	DatabaseConnectionsIdle.Set(float64(idle))
}

// RecordOutboxEvent 记录发件箱事件
func RecordOutboxEvent(eventType, status string, duration time.Duration) {
	OutboxEventsTotal.WithLabelValues(eventType, status).Inc()
	OutboxEventsProcessingDuration.WithLabelValues(eventType).Observe(duration.Seconds())
}

// UpdateOutboxEventsPending 更新待处理发件箱事件数量
func UpdateOutboxEventsPending(count float64) {
	OutboxEventsPending.Set(count)
}

// RecordError 记录错误事件
func RecordError(component, errorType string) {
	ErrorsTotal.WithLabelValues(component, errorType).Inc()
}

// RecordPanicRecovery 记录恐慌恢复事件
func RecordPanicRecovery(component string) {
	PanicRecoveries.WithLabelValues(component).Inc()
}

// === HTTP中间件 ===

// PrometheusMiddleware Prometheus监控中间件
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 增加正在处理的请求数
		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()
		
		// 包装ResponseWriter以捕获状态码
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
		
		// 处理请求
		next.ServeHTTP(wrappedWriter, r)
		
		// 记录指标
		duration := time.Since(start)
		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrappedWriter.statusCode)).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// === 指标端点 ===

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// === 系统指标更新器 ===

// UpdateSystemMetrics 更新系统指标（应该定期调用）
func UpdateSystemMetrics(uptime time.Duration, heapMemory, stackMemory, sysMemory uint64, goroutines int) {
	ServiceUptime.Set(uptime.Seconds())
	MemoryUsage.WithLabelValues("heap").Set(float64(heapMemory))
	MemoryUsage.WithLabelValues("stack").Set(float64(stackMemory))
	MemoryUsage.WithLabelValues("sys").Set(float64(sysMemory))
	GoroutinesActive.Set(float64(goroutines))
}

// === 自定义指标收集器 ===

// BusinessMetricsCollector 业务指标收集器
type BusinessMetricsCollector struct {
	// 可以添加自定义的指标收集逻辑
}

// NewBusinessMetricsCollector 创建业务指标收集器
func NewBusinessMetricsCollector() *BusinessMetricsCollector {
	return &BusinessMetricsCollector{}
}

// CollectCustomMetrics 收集自定义业务指标
func (c *BusinessMetricsCollector) CollectCustomMetrics() {
	// 这里可以实现自定义的指标收集逻辑
	// 例如：查询数据库获取业务统计信息
}

// === 指标重置器（用于测试） ===

// ResetMetricsForTesting 重置所有指标（仅用于测试）
func ResetMetricsForTesting() {
	// 注意：这个方法应该只在测试时使用
	HTTPRequestsTotal.Reset()
	HTTPRequestDuration.Reset()
	EmployeesCreatedTotal.Reset()
	AIRequestsTotal.Reset()
	DatabaseOperationsTotal.Reset()
	OutboxEventsTotal.Reset()
	ErrorsTotal.Reset()
}