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
	devMode           bool
}

func scopedLogger(base pkglogger.Logger, component string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{
		"component": component,
	}
	for k, v := range extra {
		fields[k] = v
	}
	return base.WithFields(fields)
}

func requestLogger(base pkglogger.Logger, r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	if base == nil {
		base = pkglogger.NewNoopLogger()
	}
	fields := pkglogger.Fields{}
	for k, v := range extra {
		fields[k] = v
	}
	if action != "" {
		fields["action"] = action
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
		fields["requestId"] = middleware.GetRequestID(r.Context())
		if tenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID")); tenant != "" {
			fields["tenantId"] = tenant
		}
	}
	return base.WithFields(fields)
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

func (g *GraphQLPermissionMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			if g.devMode {
				reqLogger := requestLogger(g.logger, r, "GraphQLMiddleware", pkglogger.Fields{"mode": "dev"})
				g.handleDevMode(w, r, next, reqLogger)
				return
			}

			reqLogger := requestLogger(g.logger, r, "GraphQLMiddleware", pkglogger.Fields{})
			g.handleProductionMode(w, r, next, reqLogger)
		})
	}
}

func (g *GraphQLPermissionMiddleware) handleDevMode(w http.ResponseWriter, r *http.Request, next http.Handler, logger pkglogger.Logger) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logger.Warn("authorization header required in dev mode")
		g.writeErrorResponse(w, r, logger, "DEV_UNAUTHORIZED", "Authorization header required even in development mode", http.StatusUnauthorized)
		return
	}

	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("JWT validation failed in dev mode")
		g.writeErrorResponse(w, r, logger, "DEV_INVALID_TOKEN", "Invalid JWT token in development mode: "+err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		g.writeErrorResponse(w, r, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		g.writeErrorResponse(w, r, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader

	logger.WithFields(pkglogger.Fields{"userId": claims.UserID}).Info("validated JWT token in dev mode")

	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (g *GraphQLPermissionMiddleware) handleProductionMode(w http.ResponseWriter, r *http.Request, next http.Handler, logger pkglogger.Logger) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		g.writeErrorResponse(w, r, logger, "UNAUTHORIZED", "Authorization header required", http.StatusUnauthorized)
		return
	}

	claims, err := g.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Warn("JWT validation failed")
		g.writeErrorResponse(w, r, logger, "INVALID_TOKEN", err.Error(), http.StatusUnauthorized)
		return
	}

	tenantHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantHeader == "" {
		g.writeErrorResponse(w, r, logger, "TENANT_HEADER_REQUIRED", "X-Tenant-ID header required", http.StatusUnauthorized)
		return
	}
	if claims.TenantID != "" && tenantHeader != claims.TenantID {
		g.writeErrorResponse(w, r, logger, "TENANT_MISMATCH", "X-Tenant-ID does not match tenant in token", http.StatusForbidden)
		return
	}
	claims.TenantID = tenantHeader

	ctx := SetUserContext(r.Context(), claims)
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (g *GraphQLPermissionMiddleware) CheckQueryPermission(ctx context.Context, queryName string) error {
	if g.devMode {
		return g.permissionChecker.MockPermissionCheck(ctx, queryName)
	}
	return g.permissionChecker.CheckGraphQLQuery(ctx, queryName)
}

func (g *GraphQLPermissionMiddleware) writeErrorResponse(w http.ResponseWriter, r *http.Request, logger pkglogger.Logger, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	requestID := middleware.GetRequestID(r.Context())
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("failed to encode GraphQL error response")
	}
}

func (g *GraphQLPermissionMiddleware) WriteEnterpriseErrorResponse(w http.ResponseWriter, r *http.Request, code, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	requestID := middleware.GetRequestID(r.Context())
	errorResponse := types.WriteErrorResponse(code, message, requestID, nil)
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		requestLogger(g.logger, r, "WriteEnterpriseErrorResponse", pkglogger.Fields{"code": code}).WithFields(pkglogger.Fields{"error": err}).Error("failed to encode enterprise error response")
	}
}
