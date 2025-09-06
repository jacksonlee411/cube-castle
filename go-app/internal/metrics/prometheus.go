// Package metrics provides Prometheus metrics for the organization API
// 为组织API提供Prometheus监控指标采集
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP请求总数
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of API requests",
		},
		[]string{"method", "endpoint", "status", "operation_type"},
	)

	// HTTP请求延迟
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "api_request_duration_seconds",
			Help: "API request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "operation_type"},
	)

	// 组织操作成功计数 (ADR-008 SLO指标)
	activateSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "activate_success_total",
			Help: "Total number of successful activate operations",
		},
	)

	activateRequests = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "activate_requests_total", 
			Help: "Total number of activate operation requests",
		},
	)

	suspendSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "suspend_success_total",
			Help: "Total number of successful suspend operations",
		},
	)

	suspendRequests = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "suspend_requests_total",
			Help: "Total number of suspend operation requests",
		},
	)

	// 组织操作延迟 (SLO指标)
	activateDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "activate_duration_seconds",
			Help: "Duration of activate operations in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.15, 0.2, 0.5, 1.0, 2.0, 5.0},
		},
	)

	suspendDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "suspend_duration_seconds", 
			Help: "Duration of suspend operations in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.15, 0.2, 0.5, 1.0, 2.0, 5.0},
		},
	)

	// 弃用端点访问计数 (ADR-008合规指标)
	deprecatedEndpointUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "deprecated_endpoint_used_total",
			Help: "Total number of deprecated endpoint accesses",
		},
		[]string{"endpoint", "client_id", "user_agent"},
	)

	// 审计日志写入指标
	auditWriteSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "audit_write_success_total",
			Help: "Total number of successful audit log writes",
		},
	)

	auditWriteFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "audit_write_failures_total",
			Help: "Total number of failed audit log writes",
		},
	)

	auditWriteAttempts = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "audit_write_attempts_total",
			Help: "Total number of audit log write attempts",
		},
	)

	// 权限检查指标
	permissionCheckSuccess = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "permission_check_success_total",
			Help: "Total number of successful permission checks",
		},
	)

	permissionCheckTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "permission_check_total",
			Help: "Total number of permission checks",
		},
	)

	permissionCheckDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "permission_check_duration_seconds",
			Help: "Duration of permission checks in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.2, 0.5},
		},
	)
)

// PrometheusMiddleware 为Gin路由提供Prometheus指标采集中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 执行请求
		c.Next()
		
		// 采集指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		method := c.Request.Method
		
		// 判断操作类型
		operationType := getOperationType(endpoint, method)
		
		// 记录通用指标
		requestsTotal.WithLabelValues(method, endpoint, status, operationType).Inc()
		requestDuration.WithLabelValues(method, endpoint, operationType).Observe(duration)
		
		// 记录特定操作指标 (SLO相关)
		recordOperationMetrics(endpoint, method, status, duration)
	}
}

// recordOperationMetrics 记录特定操作的SLO指标
func recordOperationMetrics(endpoint, method, status string, duration float64) {
	switch endpoint {
	case "/api/v1/organization-units/:code/activate":
		if method == "POST" {
			activateRequests.Inc()
			activateDuration.Observe(duration)
			
			if status == "200" {
				activateSuccess.Inc()
			}
		}
		
	case "/api/v1/organization-units/:code/suspend":
		if method == "POST" {
			suspendRequests.Inc()
			suspendDuration.Observe(duration)
			
			if status == "200" {
				suspendSuccess.Inc()
			}
		}
	}
}

// getOperationType 根据端点和方法判断操作类型
func getOperationType(endpoint, method string) string {
	switch {
	case endpoint == "/api/v1/organization-units" && method == "POST":
		return "CREATE"
	case endpoint == "/api/v1/organization-units/:code" && (method == "PUT" || method == "PATCH"):
		return "UPDATE"  
	case endpoint == "/api/v1/organization-units/:code/activate" && method == "POST":
		return "REACTIVATE"
	case endpoint == "/api/v1/organization-units/:code/suspend" && method == "POST":
		return "SUSPEND"
	case endpoint == "/api/v1/organization-units/:code" && method == "DELETE":
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

// RecordDeprecatedEndpointUsage 记录弃用端点访问 (ADR-008合规监控)
func RecordDeprecatedEndpointUsage(endpoint, clientID, userAgent string) {
	deprecatedEndpointUsed.WithLabelValues(endpoint, clientID, userAgent).Inc()
}

// RecordAuditWrite 记录审计日志写入结果
func RecordAuditWrite(success bool) {
	auditWriteAttempts.Inc()
	
	if success {
		auditWriteSuccess.Inc()
	} else {
		auditWriteFailures.Inc()
	}
}

// RecordPermissionCheck 记录权限检查结果
func RecordPermissionCheck(success bool, duration time.Duration) {
	permissionCheckTotal.Inc()
	permissionCheckDuration.Observe(duration.Seconds())
	
	if success {
		permissionCheckSuccess.Inc()
	}
}

// Handler 返回Prometheus HTTP处理器用于/metrics端点
func Handler() gin.HandlerFunc {
	h := promhttp.Handler()
	return gin.WrapH(h)
}

// RegisterCustomMetrics 注册自定义业务指标
func RegisterCustomMetrics() {
	// 可以在这里注册其他自定义指标
	// 例如：活跃组织数量、层级统计等业务指标
	
	prometheus.MustRegister(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "organization_active_count",
				Help: "Number of currently active organizations",
			},
			func() float64 {
				// 这里应该调用实际的业务逻辑获取活跃组织数
				// return getActiveOrganizationCount()
				return 0 // 占位符
			},
		),
	)
}