package middleware

import (
	"context"
	"net/http"
	"time"

	pkglogger "cube-castle/pkg/logger"
)

// PerformanceMiddleware 性能监控中间件
type PerformanceMiddleware struct {
	logger pkglogger.Logger
}

// NewPerformanceMiddleware 创建性能监控中间件
func NewPerformanceMiddleware(logger pkglogger.Logger) *PerformanceMiddleware {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &PerformanceMiddleware{
		logger: logger.WithFields(pkglogger.Fields{
			"component":  "middleware",
			"middleware": "performance",
		}),
	}
}

// ResponseWriterWrapper 响应包装器，用于记录响应状态和大小
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *ResponseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriterWrapper) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Middleware 性能监控中间件
func (p *PerformanceMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// 包装响应写入器
			wrapper := &ResponseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// 添加性能监控上下文
			ctx := context.WithValue(r.Context(), "start_time", startTime)
			r = r.WithContext(ctx)

			// 设置性能相关头部
			wrapper.Header().Set("X-Response-Time", "")
			wrapper.Header().Set("X-Service", "organization-command-service")

			// 执行请求处理
			next.ServeHTTP(wrapper, r)

			// 计算执行时间
			duration := time.Since(startTime)

			// 设置响应时间头部
			wrapper.Header().Set("X-Response-Time", duration.String())

			// 记录性能日志
			p.logPerformance(r, wrapper.statusCode, wrapper.size, duration)
		})
	}
}

// logPerformance 记录性能日志
func (p *PerformanceMiddleware) logPerformance(r *http.Request, statusCode, responseSize int, duration time.Duration) {
	// 获取请求ID
	requestID := GetRequestID(r.Context())

	// 分析请求类型
	requestType := "READ"
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
		requestType = "WRITE"
	}

	// 性能等级分析
	level := "NORMAL"
	icon := "✅"

	if duration > 500*time.Millisecond {
		level = "SLOW"
		icon = "⚠️"
	}
	if duration > 1*time.Second {
		level = "VERY_SLOW"
		icon = "🐌"
	}
	if duration > 3*time.Second {
		level = "CRITICAL"
		icon = "🚨"
	}

	fields := pkglogger.Fields{
		"requestId":      requestID,
		"method":         r.Method,
		"path":           r.URL.Path,
		"statusCode":     statusCode,
		"responseSize":   responseSize,
		"requestType":    requestType,
		"duration":       duration.String(),
		"durationMillis": duration.Milliseconds(),
		"performance":    level,
		"icon":           icon,
	}
	p.logger.WithFields(fields).Info("http request completed")

	// 记录详细的慢请求信息
	if duration > 1*time.Second {
		p.logSlowRequestDetails(r, statusCode, responseSize, duration, requestID)
	}
}

// logSlowRequestDetails 记录慢请求详细信息
func (p *PerformanceMiddleware) logSlowRequestDetails(r *http.Request, statusCode, responseSize int, duration time.Duration, requestID string) {
	suggestions := p.analyzePerformanceIssues(r, duration)
	fields := pkglogger.Fields{
		"requestId":    requestID,
		"method":       r.Method,
		"url":          r.URL.String(),
		"statusCode":   statusCode,
		"responseSize": responseSize,
		"duration":     duration.String(),
		"userAgent":    r.UserAgent(),
		"clientIP":     getClientIP(r),
	}
	if len(suggestions) > 0 {
		fields["suggestions"] = suggestions
	}
	p.logger.WithFields(fields).Warn("slow request detected")
}

// analyzePerformanceIssues 分析性能问题
func (p *PerformanceMiddleware) analyzePerformanceIssues(r *http.Request, duration time.Duration) []string {
	suggestions := []string{}

	// 根据请求路径分析
	if r.URL.Path == "/api/v1/organization-units" && r.Method == "POST" {
		suggestions = append(suggestions, "创建组织可能涉及复杂的层级计算")
		suggestions = append(suggestions, "检查数据库索引是否优化")
	}

	if r.URL.Path == "/graphql" {
		suggestions = append(suggestions, "GraphQL查询可能包含复杂的关联查询")
		suggestions = append(suggestions, "考虑使用数据加载器(DataLoader)优化N+1问题")
	}

	// 根据执行时间分析
	if duration > 3*time.Second {
		suggestions = append(suggestions, "考虑添加缓存机制")
		suggestions = append(suggestions, "检查数据库连接池配置")
		suggestions = append(suggestions, "考虑异步处理非关键操作")
	}

	return suggestions
}

// GetPerformanceMetrics 获取性能指标
func GetPerformanceMetrics(ctx context.Context) map[string]interface{} {
	startTime, ok := ctx.Value("start_time").(time.Time)
	if !ok {
		return nil
	}

	return map[string]interface{}{
		"executionTime": time.Since(startTime).String(),
		"startTime":     startTime.Format(time.RFC3339),
		"endTime":       time.Now().Format(time.RFC3339),
	}
}

// WithPerformanceData 添加性能数据到响应
func WithPerformanceData(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	metrics := GetPerformanceMetrics(ctx)
	if metrics != nil && data != nil {
		if meta, exists := data["meta"]; exists {
			if metaMap, ok := meta.(map[string]interface{}); ok {
				for k, v := range metrics {
					metaMap[k] = v
				}
			}
		} else {
			data["meta"] = map[string]interface{}{
				"performance": metrics,
			}
		}
	}
	return data
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头部
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	// 检查X-Real-IP头部
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// 使用RemoteAddr
	return r.RemoteAddr
}

// LogAPICall 记录API调用日志
func (p *PerformanceMiddleware) LogAPICall(method, path string, statusCode int, duration time.Duration, requestID string) {
	p.logger.WithFields(pkglogger.Fields{
		"requestId":  requestID,
		"method":     method,
		"path":       path,
		"statusCode": statusCode,
		"duration":   duration.String(),
		"durationMs": duration.Milliseconds(),
	}).Info("api call completed")
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	Threshold time.Duration
	Handler   func(r *http.Request, duration time.Duration)
}

// NewPerformanceAlert 创建性能告警
func NewPerformanceAlert(threshold time.Duration, handler func(r *http.Request, duration time.Duration)) *PerformanceAlert {
	return &PerformanceAlert{
		Threshold: threshold,
		Handler:   handler,
	}
}

// Check 检查性能阈值
func (pa *PerformanceAlert) Check(r *http.Request, duration time.Duration) {
	if duration > pa.Threshold {
		pa.Handler(r, duration)
	}
}

// DefaultPerformanceAlertHandler 默认性能告警处理器
func DefaultPerformanceAlertHandler(logger pkglogger.Logger) func(r *http.Request, duration time.Duration) {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	alertLogger := logger.WithFields(pkglogger.Fields{
		"component":  "middleware",
		"middleware": "performanceAlert",
	})
	return func(r *http.Request, duration time.Duration) {
		requestID := GetRequestID(r.Context())
		alertLogger.WithFields(pkglogger.Fields{
			"requestId": requestID,
			"method":    r.Method,
			"path":      r.URL.Path,
			"duration":  duration.String(),
		}).Warn("performance threshold exceeded")

		// 这里可以添加更多告警逻辑，如发送邮件、短信等
		// 例如: sendAlert(r, duration)
	}
}
