package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type GraphQLPermissionMiddleware struct {
	jwtMiddleware     *JWTMiddleware
	permissionChecker *PBACPermissionChecker
	logger            *log.Logger
	devMode           bool // 开发模式标志
}

func NewGraphQLPermissionMiddleware(
	jwtMiddleware *JWTMiddleware,
	permissionChecker *PBACPermissionChecker,
	logger *log.Logger,
	devMode bool,
) *GraphQLPermissionMiddleware {
	return &GraphQLPermissionMiddleware{
		jwtMiddleware:     jwtMiddleware,
		permissionChecker: permissionChecker,
		logger:            logger,
		devMode:           devMode,
	}
}

// Middleware HTTP中间件
func (g *GraphQLPermissionMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 跳过OPTIONS请求
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			// 开发模式下的宽松认证
			if g.devMode {
				g.handleDevMode(w, r, next)
				return
			}

			// 生产模式的严格JWT认证
			g.handleProductionMode(w, r, next)
		})
	}
}

// handleDevMode 开发模式处理
func (g *GraphQLPermissionMiddleware) handleDevMode(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// 检查Authorization头
	authHeader := r.Header.Get("Authorization")
	
	var claims *Claims
	if authHeader != "" {
		// 如果有JWT令牌，尝试验证
		var err error
		claims, err = g.jwtMiddleware.ValidateToken(authHeader)
		if err != nil {
			g.logger.Printf("JWT validation failed in dev mode (continuing with mock): %v", err)
			claims = g.createMockClaims(r)
		}
	} else {
		// 没有JWT令牌，使用模拟用户
		g.logger.Printf("No JWT token provided in dev mode, using mock claims")
		claims = g.createMockClaims(r)
	}

	// 设置用户上下文
	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// handleProductionMode 生产模式处理
func (g *GraphQLPermissionMiddleware) handleProductionMode(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// 提取Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.writeErrorResponse(w, "UNAUTHORIZED", "Authorization header required", 401)
		return
	}

	// 验证JWT令牌
	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		g.logger.Printf("JWT validation failed: %v", err)
		g.writeErrorResponse(w, "INVALID_TOKEN", err.Error(), 401)
		return
	}

	// 设置用户上下文
	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// createMockClaims 创建模拟用户声明（开发模式）
func (g *GraphQLPermissionMiddleware) createMockClaims(r *http.Request) *Claims {
	// 检查是否有特殊的开发头部
	mockUser := r.Header.Get("X-Mock-User")
	mockRoles := r.Header.Get("X-Mock-Roles")
	
	claims := &Claims{
		UserID:   "dev-user",
		TenantID: "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9", // 默认租户
		Roles:    []string{"ADMIN"},                         // 默认管理员权限
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

// CheckQueryPermission GraphQL查询级权限检查
func (g *GraphQLPermissionMiddleware) CheckQueryPermission(ctx context.Context, queryName string) error {
	if g.devMode {
		return g.permissionChecker.MockPermissionCheck(ctx, queryName)
	}
	return g.permissionChecker.CheckGraphQLQuery(ctx, queryName)
}

// writeErrorResponse 写入错误响应
func (g *GraphQLPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
		"timestamp": "2025-08-24T02:00:00Z", // 使用固定时间戳或实际时间
	}

	json.NewEncoder(w).Encode(response)
}

// 企业级响应格式
type EnterpriseErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     map[string]interface{} `json:"error"`
	Timestamp string                 `json:"timestamp"`
	RequestID string                 `json:"requestId,omitempty"`
}

// WriteEnterpriseErrorResponse 写入企业级错误响应
func (g *GraphQLPermissionMiddleware) WriteEnterpriseErrorResponse(w http.ResponseWriter, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := EnterpriseErrorResponse{
		Success: false,
		Error: map[string]interface{}{
			"code":    code,
			"message": message,
		},
		Timestamp: "2025-08-24T02:00:00Z",
	}

	json.NewEncoder(w).Encode(response)
}