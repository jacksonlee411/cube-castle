package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// RequestIDKey 请求ID键
	RequestIDKey ContextKey = "request_id"
	// UserIDKey 用户ID键
	UserIDKey ContextKey = "user_id"
	// TenantIDKey 租户ID键
	TenantIDKey ContextKey = "tenant_id"
)

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// 生成请求ID
			requestID := uuid.New().String()
			
			// 添加请求ID到上下文
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			r = r.WithContext(ctx)
			
			// 创建请求上下文日志器
			reqLogger := logger.WithRequestContext(requestID, "", "")
			
			// 记录请求开始
			reqLogger.Info("HTTP request started",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"content_length", r.ContentLength,
			)
			
			// 包装ResponseWriter以捕获状态码
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
			
			// 处理请求
			next.ServeHTTP(wrappedWriter, r)
			
			// 记录请求完成
			duration := time.Since(start)
			reqLogger.LogAPIRequest(
				r.Method,
				r.URL.Path,
				wrappedWriter.statusCode,
				duration,
				r.UserAgent(),
			)
			
			reqLogger.Info("HTTP request completed",
				"status_code", wrappedWriter.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrappedWriter.bytesWritten,
			)
		})
	}
}

// RecoveryMiddleware 恐慌恢复中间件
func RecoveryMiddleware(logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 记录恐慌信息
					stack := debug.Stack()
					
					reqLogger := logger.WithContext(r.Context())
					reqLogger.LogError("panic_recovered", "Panic recovered in HTTP handler", 
						&PanicError{Message: err, Stack: stack}, 
						map[string]interface{}{
							"method": r.Method,
							"path":   r.URL.Path,
							"stack":  string(stack),
						})
					
					// 记录指标
					metrics.RecordPanicRecovery("http_handler")
					metrics.RecordError("http_handler", "panic")
					
					// 返回500错误
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID, X-Tenant-ID")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// TenantMiddleware 租户上下文中间件
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取租户ID
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			// 如果没有提供租户ID，可以从JWT token中提取或使用默认值
			tenantID = "default"
		}
		
		// 添加租户ID到上下文
		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		r = r.WithContext(ctx)
		
		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware 认证中间件（简化版）
func AuthMiddleware(logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 跳过健康检查和指标端点的认证
			if r.URL.Path == "/health" || r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			
			// 从请求头获取Authorization token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// 在开发模式下，允许无认证访问
				// 生产环境中应该返回401
				userID := "anonymous"
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				r = r.WithContext(ctx)
				
				next.ServeHTTP(w, r)
				return
			}
			
			// 这里应该验证JWT token
			// 简化实现：直接提取用户ID
			userID := "user123" // 应该从JWT中提取
			
			// 添加用户ID到上下文
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			r = r.WithContext(ctx)
			
			// 记录认证事件
			tenantID := r.Context().Value(TenantIDKey).(string)
			reqLogger := logger.WithContext(r.Context())
			reqLogger.LogAuthEvent("jwt_validation", userID, tenantID, true, "token_valid")
			
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter 包装http.ResponseWriter以捕获响应信息
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += int64(n)
	return n, err
}

// PanicError 恐慌错误类型
type PanicError struct {
	Message interface{}
	Stack   []byte
}

func (e *PanicError) Error() string {
	return "panic occurred"
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 执行健康检查
		health := performHealthCheck()
		duration := time.Since(start)
		
		// 记录健康检查日志
		logger.LogHealthCheck("http_service", health.Status, duration, map[string]interface{}{
			"checks": health.Checks,
		})
		
		// 设置响应
		if health.Status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(health.ToJSON()))
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status string                 `json:"status"`
	Checks map[string]interface{} `json:"checks"`
}

// ToJSON 转换为JSON字符串
func (h *HealthStatus) ToJSON() string {
	// 简化实现，生产环境应该使用json.Marshal
	if h.Status == "healthy" {
		return `{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
	}
	return `{"status":"unhealthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
}

// performHealthCheck 执行健康检查
func performHealthCheck() *HealthStatus {
	// 这里应该检查各种依赖服务的健康状态
	// 例如：数据库连接、Redis连接、外部API等
	
	// 简化实现，总是返回健康状态
	return &HealthStatus{
		Status: "healthy",
		Checks: map[string]interface{}{
			"database": "connected",
			"redis":    "connected",
			"memory":   "normal",
		},
	}
}