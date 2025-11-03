package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"cube-castle/internal/middleware"
	"cube-castle/internal/types"
)

// RESTPermissionMiddleware 针对 REST API 的权限中间件
type RESTPermissionMiddleware struct {
	jwtMiddleware     *JWTMiddleware
	permissionChecker *PBACPermissionChecker
	logger            *log.Logger
	devMode           bool
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

			if r.devMode && (strings.HasPrefix(req.URL.Path, "/auth/dev") || strings.HasPrefix(req.URL.Path, "/dev/")) {
				r.logger.Printf("Dev mode: Skipping authentication for %s %s", req.Method, req.URL.Path)
				next.ServeHTTP(w, req)
				return
			}

			if r.devMode {
				r.handleDevMode(w, req, next)
				return
			}
			r.handleProductionMode(w, req, next)
		})
	}
}

func (r *RESTPermissionMiddleware) handleDevMode(w http.ResponseWriter, req *http.Request, next http.Handler) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		r.logger.Printf("Dev mode: Authorization header required for %s %s", req.Method, req.URL.Path)
		r.writeErrorResponse(w, req, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", http.StatusUnauthorized)
		return
	}

	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		r.logger.Printf("Dev mode: JWT validation failed: %v", err)
		r.writeErrorResponse(w, req, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(req.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		r.writeErrorResponse(w, req, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		r.writeErrorResponse(w, req, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader
	r.logger.Printf("Dev mode: Valid JWT token provided for user: %s", claims.UserID)

	ctx := SetUserContext(req.Context(), claims)
	if err := r.permissionChecker.MockRESTPermissionCheck(ctx, req.Method, req.URL.Path); err != nil {
		r.logger.Printf("Permission denied in dev mode: %v", err)
		r.writeErrorResponse(w, req, "INSUFFICIENT_PERMISSIONS", err.Error(), http.StatusForbidden)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

func (r *RESTPermissionMiddleware) handleProductionMode(w http.ResponseWriter, req *http.Request, next http.Handler) {
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		r.writeErrorResponse(w, req, "UNAUTHORIZED", "Authorization header required", http.StatusUnauthorized)
		return
	}

	claims, err := r.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		r.logger.Printf("JWT validation failed: %v", err)
		r.writeErrorResponse(w, req, "INVALID_TOKEN", err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(req.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		r.writeErrorResponse(w, req, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		r.writeErrorResponse(w, req, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader

	ctx := SetUserContext(req.Context(), claims)
	if err := r.permissionChecker.CheckRESTPermission(ctx, req.Method, req.URL.Path); err != nil {
		r.logger.Printf("Permission denied: %v", err)
		r.writeErrorResponse(w, req, "INSUFFICIENT_PERMISSIONS", err.Error(), http.StatusForbidden)
		return
	}

	next.ServeHTTP(w, req.WithContext(ctx))
}

// writeErrorResponse 写入统一错误响应
func (r *RESTPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, req *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	requestID := middleware.GetRequestID(req.Context())
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		r.logger.Printf("failed to encode REST error response: %v", err)
	}
}

// CheckAPIPermission 用于处理器内手动检查权限
func (r *RESTPermissionMiddleware) CheckAPIPermission(ctx context.Context, method, path string) error {
	if r.devMode {
		return r.permissionChecker.MockRESTPermissionCheck(ctx, method, path)
	}
	return r.permissionChecker.CheckRESTPermission(ctx, method, path)
}
