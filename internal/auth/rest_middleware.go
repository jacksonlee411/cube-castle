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

// RESTPermissionMiddleware 针对 REST API 的权限中间件
type RESTPermissionMiddleware struct {
	jwtMiddleware     *JWTMiddleware
	permissionChecker *PBACPermissionChecker
	logger            pkglogger.Logger
	devMode           bool
}

func NewRESTPermissionMiddleware(
	jwtMiddleware *JWTMiddleware,
	permissionChecker *PBACPermissionChecker,
	logger pkglogger.Logger,
	devMode bool,
) *RESTPermissionMiddleware {
	componentLogger := scopedLogger(logger, "restPermissionMiddleware", pkglogger.Fields{
		"module": "auth",
	})
	return &RESTPermissionMiddleware{
		jwtMiddleware:     jwtMiddleware,
		permissionChecker: permissionChecker,
		logger:            componentLogger,
		devMode:           devMode,
	}
}

// Middleware 返回 HTTP 中间件
func (r *RESTPermissionMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodOptions {
				next.ServeHTTP(w, req)
				return
			}

			if req.URL.Path == "/health" || req.URL.Path == "/metrics" {
				next.ServeHTTP(w, req)
				return
			}

			rqLogger := requestLogger(r.logger, req, "RESTMiddleware", pkglogger.Fields{
				"middleware": "restPermission",
			})
			if r.devMode && (strings.HasPrefix(req.URL.Path, "/auth/dev") || strings.HasPrefix(req.URL.Path, "/dev/")) {
				rqLogger.WithFields(pkglogger.Fields{"mode": "dev"}).Info("skipping authentication for dev endpoint")
				next.ServeHTTP(w, req)
				return
			}

			if r.devMode {
				r.handleDevMode(w, req, next, rqLogger)
				return
			}
			r.handleProductionMode(w, req, next, rqLogger)
		})
	}
}

func (r *RESTPermissionMiddleware) handleDevMode(w http.ResponseWriter, req *http.Request, next http.Handler, logger pkglogger.Logger) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		logger.WithFields(pkglogger.Fields{"mode": "dev"}).Warn("authorization header required in dev mode")
		r.writeErrorResponse(w, req, logger, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", http.StatusUnauthorized)
		return
	}

	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"mode": "dev", "error": err}).Warn("JWT validation failed in dev mode")
		r.writeErrorResponse(w, req, logger, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(req.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		r.writeErrorResponse(w, req, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		r.writeErrorResponse(w, req, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader
	logger.WithFields(pkglogger.Fields{"mode": "dev", "userId": claims.UserID}).Info("validated JWT token in dev mode")

	ctx := SetUserContext(req.Context(), claims)
	if err := r.permissionChecker.MockRESTPermissionCheck(ctx, req.Method, req.URL.Path); err != nil {
		logger.WithFields(pkglogger.Fields{"mode": "dev", "error": err}).Warn("permission denied in dev mode")
		r.writeErrorResponse(w, req, logger, "INSUFFICIENT_PERMISSIONS", err.Error(), http.StatusForbidden)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

func (r *RESTPermissionMiddleware) handleProductionMode(w http.ResponseWriter, req *http.Request, next http.Handler, logger pkglogger.Logger) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		r.writeErrorResponse(w, req, logger, "UNAUTHORIZED", "Authorization header required", http.StatusUnauthorized)
		return
	}

	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("JWT validation failed")
		r.writeErrorResponse(w, req, logger, "INVALID_TOKEN", err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(req.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		r.writeErrorResponse(w, req, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		r.writeErrorResponse(w, req, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader

	ctx := SetUserContext(req.Context(), claims)
	if err := r.permissionChecker.CheckRESTPermission(ctx, req.Method, req.URL.Path); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("REST permission denied")
		r.writeErrorResponse(w, req, logger, "INSUFFICIENT_PERMISSIONS", err.Error(), http.StatusForbidden)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

// writeErrorResponse 写入统一错误响应
func (r *RESTPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, req *http.Request, logger pkglogger.Logger, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	requestID := middleware.GetRequestID(req.Context())
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to encode REST error response")
	}
}

// CheckAPIPermission 用于处理器内手动检查权限
func (r *RESTPermissionMiddleware) CheckAPIPermission(ctx context.Context, method, path string) error {
	if r.devMode {
		return r.permissionChecker.MockRESTPermissionCheck(ctx, method, path)
	}
	return r.permissionChecker.CheckRESTPermission(ctx, method, path)
}
