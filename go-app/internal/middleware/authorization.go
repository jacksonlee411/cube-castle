package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gaogu/cube-castle/go-app/internal/authorization"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
)

// AuthorizationMiddleware 授权中间件
func AuthorizationMiddleware(authorizer *authorization.OPAAuthorizer, logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 跳过不需要授权的端点
			if shouldSkipAuthorization(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// 从上下文获取用户信息
			userID := getUserIDFromContext(r.Context())
			tenantID := getTenantIDFromContext(r.Context())
			
			if userID == "" {
				logger.LogSecurityEvent("missing_user_id", "", getClientIP(r), "Missing user ID in request", "high")
				http.Error(w, "Unauthorized: Missing user identification", http.StatusUnauthorized)
				return
			}

			// 验证用户信息
			user, err := authorizer.ValidateUser(r.Context(), userID, tenantID)
			if err != nil {
				logger.LogError("user_validation", "Failed to validate user", err, map[string]interface{}{
					"user_id":   userID,
					"tenant_id": tenantID,
				})
				metrics.RecordError("authorization", "user_validation_error")
				http.Error(w, "Unauthorized: User validation failed", http.StatusUnauthorized)
				return
			}

			// 执行授权检查
			result, err := authorizer.AuthorizeHTTPRequest(r.Context(), userID, tenantID, r.Method, r.URL.Path, *user)
			if err != nil {
				logger.LogError("authorization", "Authorization check failed", err, map[string]interface{}{
					"user_id":   userID,
					"tenant_id": tenantID,
					"method":    r.Method,
					"path":      r.URL.Path,
				})
				metrics.RecordError("authorization", "authorization_check_error")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// 检查授权结果
			if !result.Allowed {
				logger.LogSecurityEvent("access_denied", userID, getClientIP(r), 
					"Access denied: "+result.Reason, "medium")
				
				// 返回详细的错误信息（在生产环境中可能需要简化）
				errorResponse := map[string]interface{}{
					"error":   "Forbidden",
					"message": "Access denied",
					"reason":  result.Reason,
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			// 将用户信息添加到上下文中，供后续处理器使用
			ctx := context.WithValue(r.Context(), "user_info", user)
			r = r.WithContext(ctx)

			// 记录成功的访问
			logger.LogAccessAttempt(userID, r.URL.Path, r.Method, true, "Authorized access")

			next.ServeHTTP(w, r)
		})
	}
}

// shouldSkipAuthorization 检查是否应该跳过授权
func shouldSkipAuthorization(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// getUserIDFromContext 从上下文获取用户ID
func getUserIDFromContext(ctx context.Context) string {
	if userID := ctx.Value(UserIDKey); userID != nil {
		return userID.(string)
	}
	return ""
}

// getTenantIDFromContext 从上下文获取租户ID
func getTenantIDFromContext(ctx context.Context) string {
	if tenantID := ctx.Value(TenantIDKey); tenantID != nil {
		return tenantID.(string)
	}
	return ""
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP地址
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// RoleBasedMiddleware 基于角色的中间件（简化版）
func RoleBasedMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从上下文获取用户信息
			userInfo := r.Context().Value("user_info")
			if userInfo == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, ok := userInfo.(*authorization.UserInfo)
			if !ok {
				http.Error(w, "Invalid user information", http.StatusInternalServerError)
				return
			}

			// 检查用户角色
			hasRole := false
			for _, requiredRole := range requiredRoles {
				if user.Role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				errorResponse := map[string]interface{}{
					"error":          "Forbidden",
					"message":        "Insufficient permissions",
					"required_roles": requiredRoles,
					"user_role":      user.Role,
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnlyMiddleware 仅限管理员访问的中间件
func AdminOnlyMiddleware() func(http.Handler) http.Handler {
	return RoleBasedMiddleware("admin")
}

// HROnlyMiddleware 仅限HR访问的中间件
func HROnlyMiddleware() func(http.Handler) http.Handler {
	return RoleBasedMiddleware("admin", "hr")
}

// ManagerOnlyMiddleware 仅限经理及以上访问的中间件
func ManagerOnlyMiddleware() func(http.Handler) http.Handler {
	return RoleBasedMiddleware("admin", "hr", "manager")
}

// ResourceOwnerMiddleware 资源所有者中间件
func ResourceOwnerMiddleware(getResourceOwnerID func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := r.Context().Value("user_info")
			if userInfo == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, ok := userInfo.(*authorization.UserInfo)
			if !ok {
				http.Error(w, "Invalid user information", http.StatusInternalServerError)
				return
			}

			// 管理员可以访问任何资源
			if user.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// 检查资源所有权
			resourceOwnerID := getResourceOwnerID(r)
			if resourceOwnerID != user.ID {
				errorResponse := map[string]interface{}{
					"error":   "Forbidden", 
					"message": "You can only access your own resources",
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantIsolationMiddleware 租户隔离中间件
func TenantIsolationMiddleware(logger *logging.StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := r.Context().Value("user_info")
			if userInfo == nil {
				next.ServeHTTP(w, r)
				return
			}

			user, ok := userInfo.(*authorization.UserInfo)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			// 获取请求中的租户ID
			requestTenantID := getTenantIDFromContext(r.Context())
			
			// 检查租户隔离
			if user.TenantID != requestTenantID && user.Role != "admin" {
				logger.LogSecurityEvent("tenant_isolation_violation", user.ID, getClientIP(r),
					"Attempted cross-tenant access", "high")
				
				errorResponse := map[string]interface{}{
					"error":   "Forbidden",
					"message": "Cross-tenant access not allowed",
				}
				
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(errorResponse)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}