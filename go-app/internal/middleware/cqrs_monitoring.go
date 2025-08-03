package middleware

import (
	"context"
	"net/http"
	"time"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// CQRSMonitoringMiddleware CQRS监控中间件
type CQRSMonitoringMiddleware struct {
	logger *logging.StructuredLogger
}

// NewCQRSMonitoringMiddleware 创建CQRS监控中间件
func NewCQRSMonitoringMiddleware(logger *logging.StructuredLogger) *CQRSMonitoringMiddleware {
	return &CQRSMonitoringMiddleware{
		logger: logger,
	}
}

// MonitorCommands 监控命令执行
func (m *CQRSMonitoringMiddleware) MonitorCommands(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// 创建响应包装器以捕获状态码
		ww := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		
		// 执行命令
		next.ServeHTTP(ww, r)
		
		duration := time.Since(startTime)
		
		// 记录命令执行指标
		m.logger.Info("CQRS Command Executed",
			"command_type", "employee_command",
			"endpoint", r.URL.Path,
			"method", r.Method,
			"status_code", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"success", ww.statusCode >= 200 && ww.statusCode < 300,
		)
		
		// 记录性能指标
		if duration > 1*time.Second {
			m.logger.Warn("Slow Command Execution",
				"endpoint", r.URL.Path,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", 1000,
			)
		}
		
		// 记录错误指标
		if ww.statusCode >= 400 {
			m.logger.Error("Command Execution Error",
				"endpoint", r.URL.Path,
				"status_code", ww.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	})
}

// MonitorQueries 监控查询执行
func (m *CQRSMonitoringMiddleware) MonitorQueries(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		
		// 创建响应包装器以捕获状态码
		ww := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		
		// 执行查询
		next.ServeHTTP(ww, r)
		
		duration := time.Since(startTime)
		
		// 记录查询执行指标
		m.logger.Info("CQRS Query Executed",
			"query_type", "employee_query",
			"endpoint", r.URL.Path,
			"method", r.Method,
			"status_code", ww.statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", r.RemoteAddr,
			"cache_source", r.Header.Get("X-Cache-Source"), // Neo4j或缓存
			"success", ww.statusCode >= 200 && ww.statusCode < 300,
		)
		
		// 记录性能指标
		if duration > 500*time.Millisecond {
			m.logger.Warn("Slow Query Execution",
				"endpoint", r.URL.Path,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", 500,
			)
		}
		
		// 记录错误指标
		if ww.statusCode >= 400 {
			m.logger.Error("Query Execution Error",
				"endpoint", r.URL.Path,
				"status_code", ww.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		}
	})
}

// MonitorEvents 监控事件发布
func (m *CQRSMonitoringMiddleware) MonitorEvents(ctx context.Context, eventType, eventID string, duration time.Duration, err error) {
	if err != nil {
		m.logger.Error("Event Publishing Failed",
			"event_type", eventType,
			"event_id", eventID,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
		)
	} else {
		m.logger.Info("Event Published Successfully",
			"event_type", eventType,
			"event_id", eventID,
			"duration_ms", duration.Milliseconds(),
		)
		
		// 记录慢事件发布
		if duration > 100*time.Millisecond {
			m.logger.Warn("Slow Event Publishing",
				"event_type", eventType,
				"event_id", eventID,
				"duration_ms", duration.Milliseconds(),
				"threshold_ms", 100,
			)
		}
	}
}

// responseWrapper 响应包装器
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}