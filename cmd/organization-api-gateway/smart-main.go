package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
)

// 项目默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// ===== 服务端点配置 =====

type ServiceEndpoints struct {
	GraphQLService string // GraphQL查询服务
	RestService    string // REST API服务
	CommandService string // 命令服务
}

var endpoints = ServiceEndpoints{
	GraphQLService: "http://localhost:8090",
	RestService:    "http://localhost:8080",
	CommandService: "http://localhost:9090",
}

// ===== 服务健康状态管理 =====

type ServiceHealth struct {
	Available     bool      `json:"available"`
	LastCheck     time.Time `json:"last_check"`
	ResponseTime  int64     `json:"response_time_ms"`
	ErrorCount    int       `json:"error_count"`
	ConsecutiveErrors int   `json:"consecutive_errors"`
}

type HealthMonitor struct {
	services map[string]*ServiceHealth
	mutex    sync.RWMutex
	logger   *log.Logger
}

func NewHealthMonitor(logger *log.Logger) *HealthMonitor {
	return &HealthMonitor{
		services: map[string]*ServiceHealth{
			"graphql": {Available: true, LastCheck: time.Now()},
			"rest":    {Available: true, LastCheck: time.Now()},
			"command": {Available: true, LastCheck: time.Now()},
		},
		mutex:  sync.RWMutex{},
		logger: logger,
	}
}

func (hm *HealthMonitor) CheckService(serviceName, url string) {
	start := time.Now()
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url + "/health")
	
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	health, exists := hm.services[serviceName]
	if !exists {
		health = &ServiceHealth{}
		hm.services[serviceName] = health
	}
	
	health.LastCheck = time.Now()
	health.ResponseTime = time.Since(start).Milliseconds()
	
	if err != nil || resp.StatusCode != http.StatusOK {
		health.Available = false
		health.ErrorCount++
		health.ConsecutiveErrors++
		hm.logger.Printf("⚠️  服务健康检查失败 [%s]: %v", serviceName, err)
	} else {
		if !health.Available {
			hm.logger.Printf("✅ 服务恢复正常 [%s]", serviceName)
		}
		health.Available = true
		health.ConsecutiveErrors = 0
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func (hm *HealthMonitor) IsServiceAvailable(serviceName string) bool {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	health, exists := hm.services[serviceName]
	if !exists {
		return false
	}
	
	// 如果超过30秒没有检查，认为可能不可用
	if time.Since(health.LastCheck) > 30*time.Second {
		return false
	}
	
	// 如果连续失败超过3次，认为不可用
	return health.Available && health.ConsecutiveErrors < 3
}

func (hm *HealthMonitor) GetServiceHealth(serviceName string) *ServiceHealth {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	if health, exists := hm.services[serviceName]; exists {
		// 返回副本避免并发问题
		return &ServiceHealth{
			Available:         health.Available,
			LastCheck:         health.LastCheck,
			ResponseTime:      health.ResponseTime,
			ErrorCount:        health.ErrorCount,
			ConsecutiveErrors: health.ConsecutiveErrors,
		}
	}
	return nil
}

func (hm *HealthMonitor) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // 每10秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			hm.logger.Println("健康监控服务停止")
			return
		case <-ticker.C:
			go hm.CheckService("graphql", endpoints.GraphQLService)
			go hm.CheckService("rest", endpoints.RestService)
			go hm.CheckService("command", endpoints.CommandService)
		}
	}
}

// ===== GraphQL请求和响应类型 =====

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message string `json:"message"`
	Path    []interface{} `json:"path,omitempty"`
}

// ===== REST响应类型 =====

type StandardOrganization struct {
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	UnitType    string                 `json:"unit_type"`
	Status      string                 `json:"status"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
	SortOrder   int                    `json:"sort_order"`
	Description string                 `json:"description"`
	Profile     map[string]interface{} `json:"profile,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type StandardOrganizationsResponse struct {
	Organizations []StandardOrganization `json:"organizations"`
	TotalCount    int                    `json:"total_count"`
	Page          int                    `json:"page"`
	PageSize      int                    `json:"page_size"`
	HasNext       bool                   `json:"has_next"`
}

type StandardStatsResponse struct {
	TotalCount int            `json:"total_count"`
	ByType     map[string]int `json:"by_type"`
	ByStatus   map[string]int `json:"by_status"`
	ByLevel    map[string]int `json:"by_level"`
}

// ===== 智能路由网关 =====

type SmartAPIGateway struct {
	healthMonitor *HealthMonitor
	httpClient    *http.Client
	logger        *log.Logger
	
	// 路由统计
	graphqlAttempts int64
	graphqlFailures int64
	restFallbacks   int64
	mutex           sync.RWMutex
}

func NewSmartAPIGateway(logger *log.Logger) *SmartAPIGateway {
	return &SmartAPIGateway{
		healthMonitor: NewHealthMonitor(logger),
		httpClient:    &http.Client{Timeout: 10 * time.Second},
		logger:        logger,
	}
}

// GraphQL-first智能查询路由
func (gw *SmartAPIGateway) SmartQuery(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	gw.logger.Printf("[%s] 📡 智能查询路由开始", requestID)
	
	// 1. 首先尝试GraphQL
	if gw.healthMonitor.IsServiceAvailable("graphql") {
		gw.incrementAttempts("graphql")
		success, graphqlResp := gw.tryGraphQLQuery(r, requestID)
		if success {
			gw.writeGraphQLResponse(w, graphqlResp, requestID)
			gw.logger.Printf("[%s] ✅ GraphQL查询成功", requestID)
			return
		}
		gw.incrementFailures("graphql")
	}
	
	// 2. GraphQL失败，降级到REST API
	gw.logger.Printf("[%s] ⚠️  GraphQL不可用，降级到REST API", requestID)
	gw.incrementFallbacks("rest")
	
	if !gw.healthMonitor.IsServiceAvailable("rest") {
		gw.logger.Printf("[%s] ❌ REST API也不可用", requestID)
		http.Error(w, "All query services unavailable", http.StatusServiceUnavailable)
		return
	}
	
	// 获取查询类型并转换为REST调用
	queryType := gw.determineQueryType(r)
	switch queryType {
	case "organizations":
		gw.forwardToREST(w, r, "/api/v1/organization-units", requestID)
	case "organizationStats":
		gw.forwardToREST(w, r, "/api/v1/organization-units/stats", requestID)
	default:
		gw.forwardToREST(w, r, "/api/v1/organization-units", requestID)
	}
	
	gw.logger.Printf("[%s] ✅ REST降级查询完成", requestID)
}

func (gw *SmartAPIGateway) tryGraphQLQuery(r *http.Request, requestID string) (bool, *GraphQLResponse) {
	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gw.logger.Printf("[%s] 读取请求体失败: %v", requestID, err)
		return false, nil
	}
	
	// 转发到GraphQL服务
	url := endpoints.GraphQLService + "/graphql"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		gw.logger.Printf("[%s] 创建GraphQL请求失败: %v", requestID, err)
		return false, nil
	}
	
	// 复制头部
	gw.copyHeaders(r, req)
	
	start := time.Now()
	resp, err := gw.httpClient.Do(req)
	if err != nil {
		gw.logger.Printf("[%s] GraphQL请求失败: %v (耗时: %v)", requestID, err, time.Since(start))
		return false, nil
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		gw.logger.Printf("[%s] 读取GraphQL响应失败: %v", requestID, err)
		return false, nil
	}
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		gw.logger.Printf("[%s] GraphQL响应错误: %d", requestID, resp.StatusCode)
		return false, nil
	}
	
	// 解析GraphQL响应
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &graphqlResp); err != nil {
		gw.logger.Printf("[%s] 解析GraphQL响应失败: %v", requestID, err)
		return false, nil
	}
	
	// 检查GraphQL错误
	if len(graphqlResp.Errors) > 0 {
		gw.logger.Printf("[%s] GraphQL查询错误: %v", requestID, graphqlResp.Errors)
		return false, &graphqlResp
	}
	
	gw.logger.Printf("[%s] GraphQL请求成功 (耗时: %v)", requestID, time.Since(start))
	return true, &graphqlResp
}

func (gw *SmartAPIGateway) forwardToREST(w http.ResponseWriter, r *http.Request, path string, requestID string) {
	url := endpoints.RestService + path
	if r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		gw.logger.Printf("[%s] 创建REST请求失败: %v", requestID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	gw.copyHeaders(r, req)
	
	start := time.Now()
	resp, err := gw.httpClient.Do(req)
	if err != nil {
		gw.logger.Printf("[%s] REST请求失败: %v", requestID, err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	
	gw.logger.Printf("[%s] REST请求成功 (耗时: %v)", requestID, time.Since(start))
	
	// 直接转发响应
	gw.copyResponse(w, resp)
}

func (gw *SmartAPIGateway) determineQueryType(r *http.Request) string {
	// 从请求体中提取GraphQL查询类型
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "organizations"
	}
	
	// 重新设置请求体以便后续读取
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	
	var graphqlReq GraphQLRequest
	if err := json.Unmarshal(body, &graphqlReq); err != nil {
		return "organizations"
	}
	
	query := strings.ToLower(graphqlReq.Query)
	if strings.Contains(query, "organizationstats") {
		return "organizationStats"
	}
	if strings.Contains(query, "organization(") {
		return "organization"
	}
	return "organizations"
}

func (gw *SmartAPIGateway) writeGraphQLResponse(w http.ResponseWriter, resp *GraphQLResponse, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		gw.logger.Printf("[%s] 编码GraphQL响应失败: %v", requestID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// 命令操作直接转发到命令服务（不需要智能路由）
func (gw *SmartAPIGateway) ForwardCommand(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetReqID(r.Context())
	gw.logger.Printf("[%s] 📝 命令操作转发", requestID)
	
	if !gw.healthMonitor.IsServiceAvailable("command") {
		gw.logger.Printf("[%s] ❌ 命令服务不可用", requestID)
		http.Error(w, "Command service unavailable", http.StatusServiceUnavailable)
		return
	}
	
	body, err := io.ReadAll(r.Body)
	if err != nil {
		gw.logger.Printf("[%s] 读取请求体失败: %v", requestID, err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	
	url := endpoints.CommandService + r.URL.Path
	req, err := http.NewRequest(r.Method, url, bytes.NewBuffer(body))
	if err != nil {
		gw.logger.Printf("[%s] 创建命令请求失败: %v", requestID, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	
	gw.copyHeaders(r, req)
	
	start := time.Now()
	resp, err := gw.httpClient.Do(req)
	if err != nil {
		gw.logger.Printf("[%s] 命令请求失败: %v", requestID, err)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	
	gw.logger.Printf("[%s] ✅ 命令操作完成 (耗时: %v)", requestID, time.Since(start))
	gw.copyResponse(w, resp)
}

// 辅助方法
func (gw *SmartAPIGateway) copyHeaders(src, dst *http.Request) {
	importantHeaders := []string{
		"X-Tenant-ID", "Authorization", "Content-Type", 
		"Accept", "User-Agent", "X-Request-ID",
	}
	
	for _, header := range importantHeaders {
		if value := src.Header.Get(header); value != "" {
			dst.Header.Set(header, value)
		}
	}
	
	// 确保有默认租户ID
	if dst.Header.Get("X-Tenant-ID") == "" {
		dst.Header.Set("X-Tenant-ID", DefaultTenantIDString)
	}
}

func (gw *SmartAPIGateway) copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// 统计方法
func (gw *SmartAPIGateway) incrementAttempts(service string) {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	if service == "graphql" {
		gw.graphqlAttempts++
	}
}

func (gw *SmartAPIGateway) incrementFailures(service string) {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	if service == "graphql" {
		gw.graphqlFailures++
	}
}

func (gw *SmartAPIGateway) incrementFallbacks(service string) {
	gw.mutex.Lock()
	defer gw.mutex.Unlock()
	if service == "rest" {
		gw.restFallbacks++
	}
}

func (gw *SmartAPIGateway) GetStats() map[string]interface{} {
	gw.mutex.RLock()
	defer gw.mutex.RUnlock()
	
	var graphqlSuccessRate float64
	if gw.graphqlAttempts > 0 {
		graphqlSuccessRate = float64(gw.graphqlAttempts-gw.graphqlFailures) / float64(gw.graphqlAttempts) * 100
	}
	
	return map[string]interface{}{
		"graphql_attempts":      gw.graphqlAttempts,
		"graphql_failures":      gw.graphqlFailures,
		"graphql_success_rate":  fmt.Sprintf("%.1f%%", graphqlSuccessRate),
		"rest_fallbacks":        gw.restFallbacks,
		"services": map[string]interface{}{
			"graphql": gw.healthMonitor.GetServiceHealth("graphql"),
			"rest":    gw.healthMonitor.GetServiceHealth("rest"),
			"command": gw.healthMonitor.GetServiceHealth("command"),
		},
	}
}

// ===== 主程序 =====

func main() {
	logger := log.New(os.Stdout, "[SMART-GATEWAY] ", log.LstdFlags)

	// 创建智能API网关
	gateway := NewSmartAPIGateway(logger)

	// 启动健康监控
	ctx, cancel := context.WithCancel(context.Background())
	go gateway.healthMonitor.StartMonitoring(ctx)

	// 创建HTTP路由器
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// GraphQL端点 - 智能路由
	r.Post("/graphql", gateway.SmartQuery)
	
	// GraphiQL开发界面代理
	r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		url := endpoints.GraphQLService + "/graphiql"
		resp, err := gateway.httpClient.Get(url)
		if err != nil {
			http.Error(w, "GraphQL service unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		gateway.copyResponse(w, resp)
	})

	// 组织API路径
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		// 查询端点 - 使用智能路由
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// 将REST请求转换为GraphQL格式
			graphqlQuery := `{"query": "{ organizations { code name unitType status level path sortOrder description createdAt updatedAt } }"}`
			r.Body = io.NopCloser(strings.NewReader(graphqlQuery))
			r.Header.Set("Content-Type", "application/json")
			gateway.SmartQuery(w, r)
		})
		r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
			// 将REST统计请求转换为GraphQL格式
			graphqlQuery := `{"query": "{ organizationStats { totalCount } }"}`
			r.Body = io.NopCloser(strings.NewReader(graphqlQuery))
			r.Header.Set("Content-Type", "application/json")
			gateway.SmartQuery(w, r)
		})
		
		// 命令端点 - 直接转发
		r.Post("/", gateway.ForwardCommand)
		r.Put("/{code}", gateway.ForwardCommand)
		r.Delete("/{code}", gateway.ForwardCommand)
	})

	// 网关状态和统计
	r.Get("/gateway/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gateway.GetStats())
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		stats := gateway.GetStats()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "healthy",
			"service":  "smart-api-gateway",
			"stats":    stats,
			"features": []string{
				"GraphQL-First Routing",
				"Intelligent Fallback",
				"Health Monitoring",
				"Auto-Recovery",
			},
		})
	})

	// 根路径信息
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "Smart Organization API Gateway",
			"version": "2.0.0",
			"strategy": "GraphQL-First with Intelligent Fallback",
			"endpoints": map[string]interface{}{
				"graphql": map[string]interface{}{
					"url": "/graphql",
					"description": "GraphQL查询端点（智能路由）",
				},
				"rest": map[string]interface{}{
					"url": "/api/v1/organization-units",
					"description": "REST API端点（智能路由支持）",
				},
				"stats": map[string]interface{}{
					"url": "/gateway/stats",
					"description": "网关路由统计",
				},
			},
			"services": map[string]string{
				"graphql": endpoints.GraphQLService,
				"rest":    endpoints.RestService,
				"command": endpoints.CommandService,
			},
		})
	})

	// 创建HTTP服务器
	server := &http.Server{
		Addr:    ":8000", // 智能网关使用8000端口
		Handler: r,
	}

	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Println("正在关闭智能API网关...")
		cancel() // 停止健康监控
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Printf("智能API网关关闭失败: %v", err)
		}
	}()

	logger.Printf("🚀 智能组织API网关启动成功 - 端口 :8000")
	logger.Printf("🧠 策略: GraphQL-First with Intelligent Fallback")
	logger.Printf("📍 GraphQL端点: http://localhost:8000/graphql")
	logger.Printf("📍 REST API端点: http://localhost:8000/api/v1/organization-units")
	logger.Printf("📊 网关统计: http://localhost:8000/gateway/stats")
	logger.Printf("🔍 GraphQL开发界面: http://localhost:8000/graphiql")
	logger.Printf("📡 后端服务:")
	logger.Printf("   - GraphQL: %s", endpoints.GraphQLService)
	logger.Printf("   - REST: %s", endpoints.RestService)
	logger.Printf("   - Command: %s", endpoints.CommandService)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("智能API网关启动失败: %v", err)
	}

	logger.Println("智能API网关已关闭")
}