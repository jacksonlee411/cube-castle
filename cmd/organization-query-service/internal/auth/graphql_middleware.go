package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
    "time"
	
	"postgresql-graphql-service/internal/middleware"
	"postgresql-graphql-service/internal/types"
    gqlmetrics "postgresql-graphql-service/internal/metrics"
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

// handleDevMode 开发模式处理 - 生产就绪版本：严格JWT认证
func (g *GraphQLPermissionMiddleware) handleDevMode(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// 检查Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// 开发模式也必须提供JWT令牌
		g.logger.Printf("Dev mode: Authorization header required")
		g.writeErrorResponse(w, r, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", 401)
		return
	}

	// 验证JWT令牌
	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		g.logger.Printf("Dev mode: JWT validation failed: %v", err)
		g.writeErrorResponse(w, r, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), 401)
		return
	}

    // 校验租户头并与JWT一致
    tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
    if tenantHeader == "" {
        g.writeErrorResponse(w, r, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", 401)
        return
    }
    if claims.TenantID != "" && tenantHeader != claims.TenantID {
        g.writeErrorResponse(w, r, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", 403)
        return
    }
    claims.TenantID = tenantHeader

    g.logger.Printf("Dev mode: Valid JWT token provided for user: %s", claims.UserID)

    // 设置用户上下文
    ctx := SetUserContext(r.Context(), claims)
    next.ServeHTTP(w, r.WithContext(ctx))
}

// handleProductionMode 生产模式处理
func (g *GraphQLPermissionMiddleware) handleProductionMode(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// 提取Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.writeErrorResponse(w, r, "UNAUTHORIZED", "Authorization header required", 401)
		return
	}

	// 验证JWT令牌
	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		g.logger.Printf("JWT validation failed: %v", err)
		g.writeErrorResponse(w, r, "INVALID_TOKEN", err.Error(), 401)
		return
	}

    // 校验租户头并与JWT一致（强制）
    tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
    if tenantHeader == "" {
        g.writeErrorResponse(w, r, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", 401)
        return
    }
    if claims.TenantID != "" && tenantHeader != claims.TenantID {
        g.writeErrorResponse(w, r, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", 403)
        return
    }
    claims.TenantID = tenantHeader

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

// CheckQueryPermission GraphQL查询级权限检查
func (g *GraphQLPermissionMiddleware) CheckQueryPermission(ctx context.Context, queryName string) error {
    start := time.Now()
    var err error
    if g.devMode {
        err = g.permissionChecker.MockPermissionCheck(ctx, queryName)
    } else {
        err = g.permissionChecker.CheckGraphQLQuery(ctx, queryName)
    }
    gqlmetrics.RecordPermissionCheck(queryName, err == nil, time.Since(start))
    return err
}

// writeErrorResponse 写入错误响应
func (g *GraphQLPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, r *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一的企业级错误响应格式
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	json.NewEncoder(w).Encode(errorResponse)
}

// WriteEnterpriseErrorResponse 写入企业级错误响应
func (g *GraphQLPermissionMiddleware) WriteEnterpriseErrorResponse(w http.ResponseWriter, r *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一的企业级错误响应格式
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	json.NewEncoder(w).Encode(errorResponse)
}
