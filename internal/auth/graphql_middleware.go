package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"cube-castle/internal/middleware"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
)

type GraphQLPermissionMiddleware struct {
	jwtMiddleware     *JWTMiddleware
	permissionChecker *PBACPermissionChecker
	logger            pkglogger.Logger
	devMode           bool // 开发模式标志
}

func NewGraphQLPermissionMiddleware(
	jwtMiddleware *JWTMiddleware,
	permissionChecker *PBACPermissionChecker,
	logger pkglogger.Logger,
	devMode bool,
) *GraphQLPermissionMiddleware {
	componentLogger := scopedLogger(logger, "graphqlPermissionMiddleware", pkglogger.Fields{
		"module": "auth",
	})
	return &GraphQLPermissionMiddleware{
		jwtMiddleware:     jwtMiddleware,
		permissionChecker: permissionChecker,
		logger:            componentLogger,
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
				reqLogger := requestLogger(g.logger, r, "GraphQLMiddleware", pkglogger.Fields{"mode": "dev"})
				g.handleDevMode(w, r, next, reqLogger)
				return
			}

			// 生产模式的严格JWT认证
			reqLogger := requestLogger(g.logger, r, "GraphQLMiddleware", pkglogger.Fields{})
			g.handleProductionMode(w, r, next, reqLogger)
		})
	}
}

// handleDevMode 开发模式处理 - 生产就绪版本：严格JWT认证
func (g *GraphQLPermissionMiddleware) handleDevMode(w http.ResponseWriter, r *http.Request, next http.Handler, logger pkglogger.Logger) {
	// 检查Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// 开发模式也必须提供JWT令牌
		logger.Warn("authorization header required in dev mode")
		g.writeErrorResponse(w, r, logger, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", 401)
		return
	}

	// 验证JWT令牌
	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("JWT validation failed in dev mode")
		g.writeErrorResponse(w, r, logger, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), 401)
		return
	}

	// 校验租户头并与JWT一致
	tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		g.writeErrorResponse(w, r, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", 401)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		g.writeErrorResponse(w, r, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", 403)
		return
	}
	claims.TenantID = tenantHeader

	logger.WithFields(pkglogger.Fields{"userId": claims.UserID}).Info("validated JWT token in dev mode")

	// 设置用户上下文
	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// handleProductionMode 生产模式处理
func (g *GraphQLPermissionMiddleware) handleProductionMode(w http.ResponseWriter, r *http.Request, next http.Handler, logger pkglogger.Logger) {
	// 提取Authorization头
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.writeErrorResponse(w, r, logger, "UNAUTHORIZED", "Authorization header required", 401)
		return
	}

	// 验证JWT令牌
	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("JWT validation failed")
		g.writeErrorResponse(w, r, logger, "INVALID_TOKEN", err.Error(), 401)
		return
	}

	// 校验租户头并与JWT一致（强制）
	tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		g.writeErrorResponse(w, r, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", 401)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		g.writeErrorResponse(w, r, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", 403)
		return
	}
	claims.TenantID = tenantHeader

	// 设置用户上下文
	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

// CheckQueryPermission GraphQL查询级权限检查
func (g *GraphQLPermissionMiddleware) CheckQueryPermission(ctx context.Context, queryName string) error {
	var err error
	if g.devMode {
		err = g.permissionChecker.MockPermissionCheck(ctx, queryName)
	} else {
		err = g.permissionChecker.CheckGraphQLQuery(ctx, queryName)
	}
	return err
}

// writeErrorResponse 写入错误响应
func (g *GraphQLPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, r *http.Request, logger pkglogger.Logger, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一的企业级错误响应格式
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to encode error response")
	}
}

// WriteEnterpriseErrorResponse 写入企业级错误响应
func (g *GraphQLPermissionMiddleware) WriteEnterpriseErrorResponse(w http.ResponseWriter, r *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// 获取请求ID
	requestID := middleware.GetRequestID(r.Context())

	// 使用统一的企业级错误响应格式
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		requestLogger(g.logger, r, "WriteEnterpriseErrorResponse", pkglogger.Fields{
			"code": code,
		}).WithFields(pkglogger.Fields{"error": err}).Error("failed to encode error response")
	}
}
