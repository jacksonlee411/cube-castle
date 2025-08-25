package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"organization-command-service/internal/auth"
	"organization-command-service/internal/middleware"
)

// DevToolsHandler 开发工具处理器
type DevToolsHandler struct {
	jwtMiddleware *auth.JWTMiddleware
	logger        *log.Logger
	devMode       bool
}

// NewDevToolsHandler 创建开发工具处理器
func NewDevToolsHandler(jwtMiddleware *auth.JWTMiddleware, logger *log.Logger, devMode bool) *DevToolsHandler {
	return &DevToolsHandler{
		jwtMiddleware: jwtMiddleware,
		logger:        logger,
		devMode:       devMode,
	}
}

// SetupRoutes 设置开发工具路由
func (h *DevToolsHandler) SetupRoutes(r chi.Router) {
	// 只在开发模式下启用开发工具端点
	if h.devMode {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/dev-token", h.GenerateDevToken)
			r.Get("/dev-token/info", h.GetTokenInfo)
		})

		r.Route("/dev", func(r chi.Router) {
			r.Get("/status", h.DevStatus)
			r.Get("/test-endpoints", h.ListTestEndpoints)
		})
	}
}

// GenerateDevToken 生成开发测试令牌
func (h *DevToolsHandler) GenerateDevToken(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	var req auth.TestTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "INVALID_REQUEST", "Invalid JSON request body", http.StatusBadRequest, r)
		return
	}

	// 设置默认值
	if req.UserID == "" {
		req.UserID = "dev-user"
	}
	if req.TenantID == "" {
		req.TenantID = "dev-tenant"
	}
	if len(req.Roles) == 0 {
		req.Roles = []string{"ADMIN", "USER"}
	}
	if req.Duration == "" {
		req.Duration = "24h"
	}

	// 解析持续时间
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		h.writeErrorResponse(w, "INVALID_DURATION", "Invalid duration format", http.StatusBadRequest, r)
		return
	}

	// 生成令牌
	token, err := h.jwtMiddleware.GenerateTestToken(req.UserID, req.TenantID, req.Roles, duration)
	if err != nil {
		h.logger.Printf("Failed to generate dev token: %v", err)
		h.writeErrorResponse(w, "TOKEN_GENERATION_FAILED", "Failed to generate token", http.StatusInternalServerError, r)
		return
	}

	response := auth.TestTokenResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
		UserID:    req.UserID,
		TenantID:  req.TenantID,
		Roles:     req.Roles,
	}

	h.writeSuccessResponse(w, response, "Dev token generated successfully", r)
}

// GetTokenInfo 获取令牌信息
func (h *DevToolsHandler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeErrorResponse(w, "NO_TOKEN", "No authorization token provided", http.StatusBadRequest, r)
		return
	}

	claims, err := h.jwtMiddleware.ValidateToken(authHeader)
	if err != nil {
		h.writeErrorResponse(w, "INVALID_TOKEN", err.Error(), http.StatusUnauthorized, r)
		return
	}

	info := map[string]interface{}{
		"userId":    claims.UserID,
		"tenantId":  claims.TenantID,
		"roles":     claims.Roles,
		"expiresAt": time.Unix(claims.ExpiresAt, 0),
		"valid":     time.Now().Unix() < claims.ExpiresAt,
	}

	h.writeSuccessResponse(w, info, "Token information retrieved", r)
}

// DevStatus 开发环境状态
func (h *DevToolsHandler) DevStatus(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	status := map[string]interface{}{
		"devMode":     h.devMode,
		"timestamp":   time.Now().UTC(),
		"service":     "organization-command-service",
		"environment": "development",
		"features": map[string]bool{
			"jwtDevTools":     true,
			"testEndpoints":   true,
			"debugEndpoints":  true,
			"mockData":        true,
		},
	}

	h.writeSuccessResponse(w, status, "Development status retrieved", r)
}

// ListTestEndpoints 列出测试端点
func (h *DevToolsHandler) ListTestEndpoints(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	endpoints := map[string]interface{}{
		"devTools": []map[string]string{
			{"method": "POST", "path": "/auth/dev-token", "description": "Generate development JWT token"},
			{"method": "GET", "path": "/auth/dev-token/info", "description": "Get token information"},
			{"method": "GET", "path": "/dev/status", "description": "Get development status"},
			{"method": "GET", "path": "/dev/test-endpoints", "description": "List all test endpoints"},
		},
		"api": []map[string]string{
			{"method": "POST", "path": "/api/v1/organization-units", "description": "Create organization unit"},
			{"method": "PUT", "path": "/api/v1/organization-units/{code}", "description": "Update organization unit"},
			{"method": "DELETE", "path": "/api/v1/organization-units/{code}", "description": "Delete organization unit"},
			{"method": "POST", "path": "/api/v1/organization-units/{code}/suspend", "description": "Suspend organization unit"},
			{"method": "POST", "path": "/api/v1/organization-units/{code}/activate", "description": "Activate organization unit"},
		},
		"system": []map[string]string{
			{"method": "GET", "path": "/health", "description": "Health check endpoint"},
			{"method": "GET", "path": "/metrics", "description": "Prometheus metrics endpoint"},
		},
	}

	h.writeSuccessResponse(w, endpoints, "Test endpoints listed", r)
}

// writeSuccessResponse 写入成功响应
func (h *DevToolsHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}, message string, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	
	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"requestId": requestID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse 写入错误响应
func (h *DevToolsHandler) writeErrorResponse(w http.ResponseWriter, code, message string, statusCode int, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())
	
	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"requestId": requestID,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}