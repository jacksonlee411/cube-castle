package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
	"github.com/gaogu/cube-castle/go-app/internal/middleware"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/common"
)

const (
	ServiceName = "cube-castle-api"
	Version     = "v1.4.0"
)

func main() {
	// 初始化结构化日志器
	logger := logging.NewStructuredLogger()
	
	// 记录服务启动
	startTime := time.Now()
	logger.LogServiceStartup(ServiceName, Version, map[string]interface{}{
		"go_version": runtime.Version(),
		"arch":       runtime.GOARCH,
		"os":         runtime.GOOS,
		"port":       "8080",
	})

	// 初始化数据库连接
	db := common.InitDatabaseConnection()
	if db == nil {
		logger.LogError("database_init", "Failed to initialize database", nil, map[string]interface{}{
			"service": ServiceName,
		})
		// 在开发模式下继续运行（使用Mock）
		logger.Info("Running in mock mode - using in-memory data")
	} else {
		logger.Info("Database connected successfully")
	}

	// 初始化服务
	coreHRService := initializeCoreHRService(db, logger)

	// 创建路由器
	router := setupRoutes(logger, coreHRService)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 启动系统指标更新协程
	go startSystemMetricsUpdater(logger, startTime)

	// 启动服务器
	go func() {
		logger.Info("🚀 Cube Castle API Server starting",
			"service", ServiceName,
			"version", Version,
			"port", "8080",
			"health_check", "http://localhost:8080/health",
			"metrics", "http://localhost:8080/metrics",
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError("server_start", "Failed to start server", err, map[string]interface{}{
				"port": "8080",
			})
			log.Fatal(err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.LogError("server_shutdown", "Server forced to shutdown", err, nil)
		log.Fatal(err)
	}

	// 记录服务关闭
	uptime := time.Since(startTime)
	logger.LogServiceShutdown(ServiceName, "graceful_shutdown", uptime)
	logger.Info("✅ Server exited successfully")
}

// setupRoutes 设置路由
func setupRoutes(logger *logging.StructuredLogger, coreHRService *corehr.Service) *chi.Mux {
	r := chi.NewRouter()

	// 添加中间件
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(metrics.PrometheusMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.TenantMiddleware)
	r.Use(middleware.AuthMiddleware(logger))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// 健康检查端点（不需要认证）
	r.Get("/health", middleware.HealthCheckMiddleware(logger))

	// Prometheus指标端点（不需要认证）
	r.Handle("/metrics", metrics.MetricsHandler())

	// API v1 路由组
	r.Route("/api/v1", func(r chi.Router) {
		// CoreHR 模块路由
		r.Route("/corehr", func(r chi.Router) {
			r.Get("/employees", handleListEmployees(coreHRService, logger))
			r.Post("/employees", handleCreateEmployee(coreHRService, logger))
			r.Route("/employees/{employeeID}", func(r chi.Router) {
				r.Get("/", handleGetEmployee(coreHRService, logger))
				r.Put("/", handleUpdateEmployee(coreHRService, logger))
				r.Delete("/", handleDeleteEmployee(coreHRService, logger))
				r.Get("/manager", handleGetEmployeeManager(coreHRService, logger))
			})
			
			// 组织架构路由
			r.Get("/organizations", handleListOrganizations(coreHRService, logger))
			r.Get("/organizations/tree", handleGetOrganizationTree(coreHRService, logger))
			r.Post("/organizations", handleCreateOrganization(coreHRService, logger))
		})

		// Intelligence Gateway 路由
		r.Route("/intelligence", func(r chi.Router) {
			r.Post("/interpret", handleInterpretText(logger))
			r.Get("/health", handleIntelligenceHealth(logger))
		})

		// 监控和管理路由
		r.Route("/admin", func(r chi.Router) {
			r.Get("/metrics/business", handleBusinessMetrics(logger))
			r.Get("/health/detailed", handleDetailedHealth(logger))
			r.Post("/cache/clear", handleClearCache(logger))
		})
	})

	return r
}

// initializeCoreHRService 初始化CoreHR服务
func initializeCoreHRService(db interface{}, logger *logging.StructuredLogger) *corehr.Service {
	if db == nil {
		// Mock模式
		logger.Info("Initializing CoreHR service in mock mode")
		return corehr.NewMockService()
	}

	// 实际模式 - 这里需要根据实际的数据库连接类型进行调整
	logger.Info("Initializing CoreHR service with database connection")
	return corehr.NewMockService() // 暂时使用Mock，等数据库集成完成后更新
}

// startSystemMetricsUpdater 启动系统指标更新器
func startSystemMetricsUpdater(logger *logging.StructuredLogger, startTime time.Time) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 更新系统指标
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			uptime := time.Since(startTime)
			goroutines := runtime.NumGoroutine()
			
			metrics.UpdateSystemMetrics(
				uptime,
				m.HeapAlloc,
				m.StackInuse,
				m.Sys,
				goroutines,
			)

			// 记录性能指标
			logger.LogPerformanceMetric("memory_heap", float64(m.HeapAlloc), "bytes", map[string]string{
				"service": ServiceName,
			})
			logger.LogPerformanceMetric("goroutines", float64(goroutines), "count", map[string]string{
				"service": ServiceName,
			})
		}
	}
}

// === HTTP处理器函数 ===

// handleListEmployees 处理员工列表请求
func handleListEmployees(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		
		// 获取查询参数
		page := getIntParam(r, "page", 1)
		pageSize := getIntParam(r, "page_size", 20)
		search := r.URL.Query().Get("search")
		tenantID := getTenantID(r.Context())
		
		// 调用服务
		response, err := service.ListEmployees(r.Context(), tenantID, page, pageSize, search)
		if err != nil {
			reqLogger.LogError("list_employees", "Failed to list employees", err, map[string]interface{}{
				"page": page,
				"page_size": pageSize,
				"search": search,
			})
			metrics.RecordError("corehr", "list_employees_error")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		metrics.RecordDatabaseOperation("SELECT", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("SELECT", "employees", len(*response.Employees), duration, true)

		// 返回响应
		respondJSON(w, http.StatusOK, response)
	}
}

// handleCreateEmployee 处理创建员工请求
func handleCreateEmployee(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 解析请求体
		var req CreateEmployeeRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request", "Failed to parse create employee request", err, nil)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 调用服务
		employee, err := service.CreateEmployee(r.Context(), tenantID, &req)
		if err != nil {
			reqLogger.LogError("create_employee", "Failed to create employee", err, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"first_name": req.FirstName,
				"last_name": req.LastName,
			})
			metrics.RecordError("corehr", "create_employee_error")
			http.Error(w, "Failed to create employee", http.StatusInternalServerError)
			return
		}

		// 记录指标和日志
		duration := time.Since(start)
		metrics.RecordEmployeeCreated(tenantID.String())
		metrics.RecordDatabaseOperation("INSERT", "employees", "success", duration)
		reqLogger.LogEmployeeCreated(*employee.Id, tenantID, req.EmployeeNumber)

		// 返回响应
		respondJSON(w, http.StatusCreated, employee)
	}
}

// 其他处理器函数的实现可以类似地添加...

// === 辅助函数 ===

// getIntParam 获取整数参数
func getIntParam(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getTenantID 从上下文获取租户ID
func getTenantID(ctx context.Context) uuid.UUID {
	if tenantID := ctx.Value(middleware.TenantIDKey); tenantID != nil {
		if id, err := uuid.Parse(tenantID.(string)); err == nil {
			return id
		}
	}
	// 返回默认租户ID
	return uuid.MustParse("00000000-0000-0000-0000-000000000000")
}

// parseJSON 解析JSON请求体
func parseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

// respondJSON 发送JSON响应
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// CreateEmployeeRequest 临时类型定义（应该从openapi生成的代码中导入）
type CreateEmployeeRequest = openapi.CreateEmployeeRequest

// === 占位符处理器（待实现） ===

func handleGetEmployee(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleUpdateEmployee(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleDeleteEmployee(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleGetEmployeeManager(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleListOrganizations(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleGetOrganizationTree(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleCreateOrganization(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleInterpretText(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleIntelligenceHealth(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}
}

func handleBusinessMetrics(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}

func handleDetailedHealth(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
	}
}

func handleClearCache(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "not_implemented"})
	}
}