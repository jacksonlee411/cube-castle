package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	auth "cube-castle/internal/auth"
	"cube-castle/internal/middleware"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
)

// DevToolsHandler 开发工具处理器
type DevToolsHandler struct {
	jwtMiddleware *auth.JWTMiddleware
	logger        pkglogger.Logger
	devMode       bool
	db            *sql.DB
}

// NewDevToolsHandler 创建开发工具处理器
func NewDevToolsHandler(jwtMiddleware *auth.JWTMiddleware, baseLogger pkglogger.Logger, devMode bool, db *sql.DB) *DevToolsHandler {
	return &DevToolsHandler{
		jwtMiddleware: jwtMiddleware,
		logger:        scopedLogger(baseLogger, "devTools", pkglogger.Fields{"routeGroup": "devtools"}),
		devMode:       devMode,
		db:            db,
	}
}

func (h *DevToolsHandler) requestLogger(r *http.Request, action string) pkglogger.Logger {
	fields := pkglogger.Fields{
		"action": action,
	}
	if r != nil {
		fields["method"] = r.Method
		fields["path"] = r.URL.Path
		fields["requestId"] = middleware.GetRequestID(r.Context())
	}
	return h.logger.WithFields(fields)
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
			r.Get("/database-status", h.DatabaseStatus)
			r.Get("/performance-metrics", h.PerformanceMetrics)
			r.Post("/test-api", h.TestAPI)
		})
	}
}

// GenerateDevToken 生成开发测试令牌
func (h *DevToolsHandler) GenerateDevToken(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}
	logger := h.requestLogger(r, "generateDevToken")

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
		req.TenantID = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
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

	scope := strings.TrimSpace(req.Scope)
	var permissions []string
	for _, p := range req.Permissions {
		p = strings.TrimSpace(p)
		if p != "" {
			permissions = append(permissions, p)
		}
	}

	// 生成令牌
	token, err := h.jwtMiddleware.GenerateTestTokenWithClaims(req.UserID, req.TenantID, req.Roles, scope, permissions, duration)
	if err != nil {
		logger.WithFields(pkglogger.Fields{"error": err}).Error("generate dev token failed")
		h.writeErrorResponse(w, "TOKEN_GENERATION_FAILED", "Failed to generate token", http.StatusInternalServerError, r)
		return
	}

	response := auth.TestTokenResponse{
		Token:       token,
		ExpiresAt:   time.Now().Add(duration),
		UserID:      req.UserID,
		TenantID:    req.TenantID,
		Roles:       req.Roles,
		Scope:       scope,
		Permissions: permissions,
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
		"userId":      claims.UserID,
		"tenantId":    claims.TenantID,
		"roles":       claims.Roles,
		"scope":       claims.Scope,
		"permissions": claims.Permissions,
		"expiresAt":   time.Unix(claims.ExpiresAt, 0),
		"valid":       time.Now().Unix() < claims.ExpiresAt,
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
			"jwtDevTools":    true,
			"testEndpoints":  true,
			"debugEndpoints": true,
			"mockData":       true,
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
			{"method": "GET", "path": "/dev/database-status", "description": "Check database connection and table stats"},
			{"method": "GET", "path": "/dev/performance-metrics", "description": "Get runtime performance metrics"},
			{"method": "POST", "path": "/dev/test-api", "description": "Test API endpoints with custom requests"},
		},
		"api": []map[string]string{
			{"method": "POST", "path": "/api/v1/organization-units", "description": "Create organization unit"},
			{"method": "PUT", "path": "/api/v1/organization-units/{code}", "description": "Update organization unit"},
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.requestLogger(r, "writeSuccessResponse").
			WithFields(pkglogger.Fields{"error": err}).
			Error("encode devtools success response failed")
	}
}

// DatabaseStatus 数据库状态检查
func (h *DevToolsHandler) DatabaseStatus(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	status := map[string]interface{}{
		"connected": false,
		"tables":    make(map[string]interface{}),
		"stats":     make(map[string]interface{}),
	}

	// 检查数据库连接
	if err := h.db.Ping(); err != nil {
		h.writeErrorResponse(w, "DATABASE_DISCONNECTED", fmt.Sprintf("Database connection failed: %v", err), http.StatusServiceUnavailable, r)
		return
	}
	status["connected"] = true

	// 检查主要表的记录数（固定白名单，避免动态拼接）
	tableQueries := map[string]string{
		"organization_units":         "SELECT COUNT(*) FROM organization_units",
		"organization_units_history": "SELECT COUNT(*) FROM organization_units_history",
		"audit_logs":                 "SELECT COUNT(*) FROM audit_logs",
	}
	tableOrder := []string{"organization_units", "organization_units_history", "audit_logs"}
	for _, table := range tableOrder {
		query := tableQueries[table]
		var count int
		if err := h.db.QueryRow(query).Scan(&count); err != nil {
			status["tables"].(map[string]interface{})[table] = map[string]interface{}{
				"error": err.Error(),
				"count": -1,
			}
			continue
		}
		status["tables"].(map[string]interface{})[table] = map[string]interface{}{
			"count":  count,
			"status": "healthy",
		}
	}

	// 数据库统计信息
	var dbSize string
	if err := h.db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize); err == nil {
		status["stats"].(map[string]interface{})["database_size"] = dbSize
	}

	h.writeSuccessResponse(w, status, "Database status retrieved", r)
}

// PerformanceMetrics 性能指标
func (h *DevToolsHandler) PerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"memory": map[string]interface{}{
			"allocated_mb":    fmt.Sprintf("%.2f", float64(memStats.Alloc)/1024/1024),
			"total_allocated": fmt.Sprintf("%.2f", float64(memStats.TotalAlloc)/1024/1024),
			"sys_mb":          fmt.Sprintf("%.2f", float64(memStats.Sys)/1024/1024),
			"gc_cycles":       memStats.NumGC,
		},
		"goroutines": runtime.NumGoroutine(),
		"timestamp":  time.Now().UTC(),
	}

	// 数据库连接池统计
	if h.db != nil {
		dbStats := h.db.Stats()
		metrics["database"] = map[string]interface{}{
			"open_connections":    dbStats.OpenConnections,
			"in_use":              dbStats.InUse,
			"idle":                dbStats.Idle,
			"wait_count":          dbStats.WaitCount,
			"wait_duration_ms":    dbStats.WaitDuration.Milliseconds(),
			"max_idle_closed":     dbStats.MaxIdleClosed,
			"max_lifetime_closed": dbStats.MaxLifetimeClosed,
		}
	}

	h.writeSuccessResponse(w, metrics, "Performance metrics retrieved", r)
}

// TestAPI API测试工具
func (h *DevToolsHandler) TestAPI(w http.ResponseWriter, r *http.Request) {
	if !h.devMode {
		h.writeErrorResponse(w, "DEV_MODE_DISABLED", "Development tools are disabled", http.StatusForbidden, r)
		return
	}

	var req struct {
		Method  string            `json:"method"`
		Path    string            `json:"path"`
		Headers map[string]string `json:"headers"`
		Body    interface{}       `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "INVALID_REQUEST", "Invalid JSON request body", http.StatusBadRequest, r)
		return
	}

	// 构建测试请求
	var bodyReader io.Reader
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			h.writeErrorResponse(w, "BODY_MARSHAL_ERROR", "Failed to marshal request body", http.StatusBadRequest, r)
			return
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	// 构建目标URL
	targetURL := fmt.Sprintf("http://localhost:9090%s", req.Path)

	// 创建请求
	testReq, err := http.NewRequest(req.Method, targetURL, bodyReader)
	if err != nil {
		h.writeErrorResponse(w, "REQUEST_CREATE_ERROR", err.Error(), http.StatusInternalServerError, r)
		return
	}

	// 设置头部
	for key, value := range req.Headers {
		testReq.Header.Set(key, value)
	}

	// 默认设置Content-Type
	if req.Body != nil && testReq.Header.Get("Content-Type") == "" {
		testReq.Header.Set("Content-Type", "application/json")
	}

	// 执行请求
	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(testReq)
	duration := time.Since(start)

	if err != nil {
		h.writeErrorResponse(w, "REQUEST_EXECUTION_ERROR", err.Error(), http.StatusInternalServerError, r)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.writeErrorResponse(w, "RESPONSE_READ_ERROR", err.Error(), http.StatusInternalServerError, r)
		return
	}

	// 构建测试结果
	result := map[string]interface{}{
		"request": map[string]interface{}{
			"method":  req.Method,
			"url":     targetURL,
			"headers": req.Headers,
			"body":    req.Body,
		},
		"response": map[string]interface{}{
			"status_code":  resp.StatusCode,
			"status":       resp.Status,
			"headers":      resp.Header,
			"body":         string(respBody),
			"content_type": resp.Header.Get("Content-Type"),
		},
		"timing": map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
			"duration":    duration.String(),
		},
		"test_metadata": map[string]interface{}{
			"timestamp": time.Now().UTC(),
			"success":   resp.StatusCode >= 200 && resp.StatusCode < 400,
		},
	}

	h.writeSuccessResponse(w, result, "API test completed", r)
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
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.requestLogger(r, "writeErrorResponse").
			WithFields(pkglogger.Fields{"error": err}).
			Error("encode devtools error response failed")
	}
}
