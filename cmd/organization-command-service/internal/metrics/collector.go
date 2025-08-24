package metrics

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector Prometheus指标收集器
type MetricsCollector struct {
	// 请求指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// 业务指标
	organizationsTotal       prometheus.Gauge
	organizationOperations   *prometheus.CounterVec
	auditEventsTotal         *prometheus.CounterVec
	cascadeTasksTotal        *prometheus.CounterVec
	validationErrorsTotal    *prometheus.CounterVec

	// 系统指标
	dbConnectionsActive      prometheus.Gauge
	dbConnectionsIdle        prometheus.Gauge
	dbQueriesTotal          *prometheus.CounterVec
	dbQueryDuration         *prometheus.HistogramVec

	logger *log.Logger
}

func NewMetricsCollector(logger *log.Logger) *MetricsCollector {
	collector := &MetricsCollector{
		// HTTP请求指标
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cube_castle",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		httpRequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cube_castle",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),

		// 业务指标
		organizationsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cube_castle",
				Subsystem: "business",
				Name:      "organizations_total",
				Help:      "Total number of active organizations",
			},
		),
		organizationOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "business",
				Name:      "organization_operations_total",
				Help:      "Total number of organization operations",
			},
			[]string{"operation", "status", "tenant_id"},
		),
		auditEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "audit",
				Name:      "events_total",
				Help:      "Total number of audit events",
			},
			[]string{"event_type", "resource_type", "success"},
		),
		cascadeTasksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "cascade",
				Name:      "tasks_total",
				Help:      "Total number of cascade tasks",
			},
			[]string{"task_type", "status"},
		),
		validationErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "validation",
				Name:      "errors_total",
				Help:      "Total number of validation errors",
			},
			[]string{"error_type", "field"},
		),

		// 系统指标
		dbConnectionsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cube_castle",
				Subsystem: "db",
				Name:      "connections_active",
				Help:      "Number of active database connections",
			},
		),
		dbConnectionsIdle: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "cube_castle",
				Subsystem: "db",
				Name:      "connections_idle",
				Help:      "Number of idle database connections",
			},
		),
		dbQueriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "cube_castle",
				Subsystem: "db",
				Name:      "queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"query_type", "success"},
		),
		dbQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "cube_castle",
				Subsystem: "db",
				Name:      "query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"query_type"},
		),

		logger: logger,
	}

	// 注册所有指标
	prometheus.MustRegister(
		collector.httpRequestsTotal,
		collector.httpRequestDuration,
		collector.httpRequestsInFlight,
		collector.organizationsTotal,
		collector.organizationOperations,
		collector.auditEventsTotal,
		collector.cascadeTasksTotal,
		collector.validationErrorsTotal,
		collector.dbConnectionsActive,
		collector.dbConnectionsIdle,
		collector.dbQueriesTotal,
		collector.dbQueryDuration,
	)

	logger.Println("✅ Prometheus指标收集器已初始化")
	return collector
}

// HTTP中间件 - 记录HTTP请求指标
func (m *MetricsCollector) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 增加并发请求计数
		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		// 包装ResponseWriter来捕获状态码
		wrw := &responseWriter{ResponseWriter: w, statusCode: 200}

		// 执行下一个处理器
		next.ServeHTTP(wrw, r)

		// 记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrw.statusCode)
		
		m.httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		m.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)

		m.logger.Printf("📊 HTTP指标: %s %s -> %s (%.3fs)", 
			r.Method, r.URL.Path, status, duration)
	})
}

// responseWriter 包装器用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// 业务指标记录方法

// RecordOrganizationOperation 记录组织操作指标
func (m *MetricsCollector) RecordOrganizationOperation(operation, status, tenantID string) {
	m.organizationOperations.WithLabelValues(operation, status, tenantID).Inc()
	m.logger.Printf("📊 组织操作指标: %s -> %s (租户: %s)", operation, status, tenantID)
}

// RecordAuditEvent 记录审计事件指标
func (m *MetricsCollector) RecordAuditEvent(eventType, resourceType string, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	m.auditEventsTotal.WithLabelValues(eventType, resourceType, successStr).Inc()
	m.logger.Printf("📊 审计事件指标: %s/%s -> %s", eventType, resourceType, successStr)
}

// RecordCascadeTask 记录级联任务指标
func (m *MetricsCollector) RecordCascadeTask(taskType, status string) {
	m.cascadeTasksTotal.WithLabelValues(taskType, status).Inc()
	m.logger.Printf("📊 级联任务指标: %s -> %s", taskType, status)
}

// RecordValidationError 记录验证错误指标
func (m *MetricsCollector) RecordValidationError(errorType, field string) {
	m.validationErrorsTotal.WithLabelValues(errorType, field).Inc()
	m.logger.Printf("📊 验证错误指标: %s (字段: %s)", errorType, field)
}

// RecordDBQuery 记录数据库查询指标
func (m *MetricsCollector) RecordDBQuery(queryType string, duration time.Duration, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	m.dbQueriesTotal.WithLabelValues(queryType, successStr).Inc()
	m.dbQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
	
	m.logger.Printf("📊 数据库查询指标: %s -> %s (%.3fs)", 
		queryType, successStr, duration.Seconds())
}

// UpdateOrganizationsCount 更新组织总数
func (m *MetricsCollector) UpdateOrganizationsCount(count float64) {
	m.organizationsTotal.Set(count)
	m.logger.Printf("📊 组织总数更新: %.0f", count)
}

// UpdateDBConnections 更新数据库连接数
func (m *MetricsCollector) UpdateDBConnections(active, idle int) {
	m.dbConnectionsActive.Set(float64(active))
	m.dbConnectionsIdle.Set(float64(idle))
	m.logger.Printf("📊 数据库连接: 活跃=%d, 空闲=%d", active, idle)
}

// GetHandler 返回Prometheus HTTP处理器
func (m *MetricsCollector) GetHandler() http.Handler {
	return promhttp.Handler()
}

// GetMetricsMiddleware 返回指标中间件
func (m *MetricsCollector) GetMetricsMiddleware() func(http.Handler) http.Handler {
	return m.HTTPMiddleware
}