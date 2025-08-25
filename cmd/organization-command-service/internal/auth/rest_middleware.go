package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"organization-command-service/internal/middleware"
	"organization-command-service/internal/types"
)

type RESTPermissionMiddleware struct {
	jwtMiddleware     *JWTMiddleware
	permissionChecker *PBACPermissionChecker
	logger            *log.Logger
	devMode           bool // 开发模式标志
}

func NewRESTPermissionMiddleware(
	jwtMiddleware *JWTMiddleware,
	permissionChecker *PBACPermissionChecker,
	logger *log.Logger,
	devMode bool,
) *RESTPermissionMiddleware {
	return &RESTPermissionMiddleware{
		jwtMiddleware:     jwtMiddleware,
		permissionChecker: permissionChecker,
		logger:            logger,
		devMode:           devMode,
	}
}

// Middleware HTTP中间件
func (r *RESTPermissionMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// 跳过OPTIONS请求
			if req.Method == "OPTIONS" {
				next.ServeHTTP(w, req)
				return
			}

			// 跳过健康检查和指标端点
			if req.URL.Path == "/health" || req.URL.Path == "/metrics" {
				next.ServeHTTP(w, req)
				return
			}

			// 开发模式下的宽松认证
			if r.devMode {
				r.handleDevMode(w, req, next)
				return
			}

			// 生产模式的严格JWT认证
			r.handleProductionMode(w, req, next)
		})
	}
}

// handleDevMode 开发模式处理 - 生产就绪版本：严格JWT认证
func (r *RESTPermissionMiddleware) handleDevMode(w http.ResponseWriter, req *http.Request, next http.Handler) {
	// 检查Authorization头
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		// 开发模式也必须提供JWT令牌
		r.logger.Printf("Dev mode: Authorization header required for %s %s", req.Method, req.URL.Path)
		r.writeErrorResponse(w, req, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", 401)
		return
	}

	// 验证JWT令牌
	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		r.logger.Printf("Dev mode: JWT validation failed: %v", err)
		r.writeErrorResponse(w, req, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), 401)
		return
	}

	r.logger.Printf("Dev mode: Valid JWT token provided for user: %s", claims.UserID)

	// 设置用户上下文
	ctx := SetUserContext(req.Context(), claims)

	// 开发模式下的权限检查（相对宽松）
	if err := r.permissionChecker.MockPermissionCheck(ctx, req.Method, req.URL.Path); err != nil {
		r.logger.Printf("Permission denied in dev mode: %v", err)
		r.writeErrorResponse(w, req, "PERMISSION_DENIED", err.Error(), 403)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

// handleProductionMode 生产模式处理
func (r *RESTPermissionMiddleware) handleProductionMode(w http.ResponseWriter, req *http.Request, next http.Handler) {
	// 提取Authorization头
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		r.writeErrorResponse(w, req, "UNAUTHORIZED", "Authorization header required", 401)
		return
	}

	// 验证JWT令牌
	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		r.logger.Printf("JWT validation failed: %v", err)
		r.writeErrorResponse(w, req, "INVALID_TOKEN", err.Error(), 401)
		return
	}

	// 设置用户上下文
	ctx := SetUserContext(req.Context(), claims)

	// 严格的权限检查
	if err := r.permissionChecker.CheckRESTAPI(req.WithContext(ctx)); err != nil {
		r.logger.Printf("Permission denied: %v", err)
		r.writeErrorResponse(w, req, "PERMISSION_DENIED", err.Error(), 403)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

// createMockClaims 创建模拟用户声明（开发模式）
func (r *RESTPermissionMiddleware) createMockClaims(req *http.Request) *Claims {
	// 检查是否有特殊的开发头部
	mockUser := req.Header.Get("X-Mock-User")
	mockRoles := req.Header.Get("X-Mock-Roles")

	claims := &Claims{
		UserID:   "dev-user",
		TenantID: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", // 默认租户
		Roles:    []string{"ADMIN"},                      // 默认管理员权限
	}

	if mockUser != "" {
		claims.UserID = mockUser
	}

	if mockRoles != "" {
		claims.Roles = strings.Split(mockRoles, ",")
		for i, role := range claims.Roles {
			claims.Roles[i] = strings.TrimSpace(role)
		}
	}

	return claims
}

// writeErrorResponse 写入错误响应
func (r *RESTPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, req *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 获取请求ID
	requestID := middleware.GetRequestID(req.Context())

	// 使用统一的企业级错误响应格式
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	json.NewEncoder(w).Encode(errorResponse)
}

// CheckAPIPermission REST API级权限检查
func (r *RESTPermissionMiddleware) CheckAPIPermission(ctx context.Context, method, path string) error {
	if r.devMode {
		return r.permissionChecker.MockPermissionCheck(ctx, method, path)
	}
	return r.permissionChecker.CheckPermission(ctx, method, path)
}