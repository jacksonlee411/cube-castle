package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"organization-command-service/internal/audit"
	"organization-command-service/internal/auth"
	"organization-command-service/internal/authbff"
	"organization-command-service/internal/config"
	"organization-command-service/internal/handlers"
	"organization-command-service/internal/middleware"
	"organization-command-service/internal/repository"
	"organization-command-service/internal/services"
	"organization-command-service/internal/validators"
)

func main() {
	logger := log.New(os.Stdout, "[COMMAND-SERVICE] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 启动组织命令服务...")
	authOnlyMode := os.Getenv("AUTH_ONLY_MODE") == "true"

	var db *sql.DB
	if !authOnlyMode {
		// 数据库连接
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
		}

		var err error
		db, err = sql.Open("postgres", dbURL)
		if err != nil {
			logger.Fatalf("数据库连接失败: %v", err)
		}
		defer db.Close()

		// 验证数据库连接
		if err := db.Ping(); err != nil {
			logger.Fatalf("数据库连接验证失败: %v", err)
		}

		logger.Println("✅ 数据库连接成功")
	} else {
		logger.Println("🟡 AUTH_ONLY_MODE=true：跳过数据库连接，仅启用 BFF /auth 与 /.well-known 端点")
	}

	var (
		orgRepo           *repository.OrganizationRepository
		hierarchyRepo     *repository.HierarchyRepository
		cascadeService    *services.CascadeUpdateService
		auditLogger       *audit.AuditLogger
		businessValidator *validators.BusinessRuleValidator
	)
	if !authOnlyMode {
		// 初始化仓储层
		orgRepo = repository.NewOrganizationRepository(db, logger)
		hierarchyRepo = repository.NewHierarchyRepository(db, logger)

		// 初始化业务服务层
		cascadeService = services.NewCascadeUpdateService(hierarchyRepo, 4, logger)
		businessValidator = validators.NewBusinessRuleValidator(hierarchyRepo, orgRepo, logger)
		auditLogger = audit.NewAuditLogger(db, logger)

		// 启动级联更新服务
		cascadeService.Start()
		logger.Println("✅ 级联更新服务已启动")
		logger.Println("✅ 结构化审计日志系统已初始化")
		logger.Println("✅ Prometheus指标收集系统已初始化")
	}

	// 初始化JWT中间件 - 使用统一配置
	jwtConfig := config.GetJWTConfig()
	devMode := os.Getenv("DEV_MODE") == "true"
	if os.Getenv("DEV_MODE") == "" {
		devMode = true // 默认开发模式
	}

	var (
		pubPEM  []byte
		privPEM []byte
	)
	if jwtConfig.HasPublicKey() {
		if b, err := os.ReadFile(jwtConfig.PublicKeyPath); err == nil {
			pubPEM = b
		} else {
			logger.Fatalf("[FATAL] 无法读取JWT公钥 (%s): %v", jwtConfig.PublicKeyPath, err)
		}
	}
	if !jwtConfig.HasPrivateKey() {
		logger.Fatalf("[FATAL] 启用了RS256但未配置JWT_PRIVATE_KEY_PATH。请运行 make jwt-dev-setup 或提供正式私钥文件。")
	}
	if b, err := os.ReadFile(jwtConfig.PrivateKeyPath); err == nil {
		privPEM = b
	} else {
		logger.Fatalf("[FATAL] 无法读取JWT私钥 (%s): %v", jwtConfig.PrivateKeyPath, err)
	}

	jwtMiddleware := auth.NewJWTMiddlewareWithOptions(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, auth.Options{
		Alg:           jwtConfig.Algorithm,
		JWKSURL:       jwtConfig.JWKSUrl,
		PublicKeyPEM:  pubPEM,
		PrivateKeyPEM: privPEM,
		KeyID:         jwtConfig.KeyID,
		ClockSkew:     jwtConfig.AllowedClockSkew,
	})
	var restAuthMiddleware *auth.RESTPermissionMiddleware
	if !authOnlyMode {
		permissionChecker := auth.NewPBACPermissionChecker(db, logger)
		restAuthMiddleware = auth.NewRESTPermissionMiddleware(
			jwtMiddleware,
			permissionChecker,
			logger,
			devMode,
		)
	}

	logger.Printf("🔐 JWT认证初始化完成 (开发模式: %v, Alg=%s, Issuer=%s, Audience=%s)", devMode, jwtConfig.Algorithm, jwtConfig.Issuer, jwtConfig.Audience)

	// 初始化中间件
	performanceMiddleware := middleware.NewPerformanceMiddleware(logger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(middleware.DefaultRateLimitConfig, logger)

	// 初始化时态服务
	var temporalService *services.TemporalService
	if !authOnlyMode {
		temporalService = services.NewTemporalService(db)
	}

	// 初始化监控服务
	var temporalMonitor *services.TemporalMonitor
	if !authOnlyMode {
		temporalMonitor = services.NewTemporalMonitor(db, logger)
	}

	// 初始化运维调度器
	var operationalScheduler *services.OperationalScheduler
	if !authOnlyMode {
		operationalScheduler = services.NewOperationalScheduler(db, logger, temporalMonitor)
	}

	// 初始化时态时间轴管理器
	var timelineManager *repository.TemporalTimelineManager
	if !authOnlyMode {
		timelineManager = repository.NewTemporalTimelineManager(db, logger)
	}

	// 初始化处理器
	var (
		orgHandler         *handlers.OrganizationHandler
		devToolsHandler    *handlers.DevToolsHandler
		operationalHandler *handlers.OperationalHandler
	)
	if !authOnlyMode {
		orgHandler = handlers.NewOrganizationHandler(orgRepo, temporalService, auditLogger, logger, timelineManager, hierarchyRepo, businessValidator)
		operationalHandler = handlers.NewOperationalHandler(temporalMonitor, operationalScheduler, rateLimitMiddleware, logger)
	}
	// 开发工具路由即使在 authOnly 模式下也允许初始化（内部会根据 devMode 控制）
	devToolsHandler = handlers.NewDevToolsHandler(jwtMiddleware, logger, devMode, db)

	// 设置路由
	r := chi.NewRouter()

	// 基础中间件链 (无认证要求的中间件)
	r.Use(middleware.RequestIDMiddleware)     // 请求追踪中间件
	r.Use(rateLimitMiddleware.Middleware())   // 限流中间件 - 最先执行
	r.Use(performanceMiddleware.Middleware()) // 性能监控中间件
	r.Use(chi_middleware.Logger)
	r.Use(chi_middleware.Recoverer)
	r.Use(chi_middleware.Timeout(30 * time.Second))

	// CORS设置
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001", "http://localhost:3002", "http://localhost:3003", "http://localhost:3004"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "healthy", "service": "organization-command-service", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	// 限流状态监控端点（Dev-only）
	if devMode {
		r.Get("/debug/rate-limit/stats", func(w http.ResponseWriter, r *http.Request) {
			stats := rateLimitMiddleware.GetStats()
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
                "totalRequests": %d,
                "blockedRequests": %d,
                "activeClients": %d,
                "lastReset": "%s",
                "blockRate": "%.2f%%"
            }`, stats.TotalRequests, stats.BlockedRequests, stats.ActiveClients,
				stats.LastReset.Format(time.RFC3339),
				float64(stats.BlockedRequests)/float64(stats.TotalRequests)*100)
		})

		r.Get("/debug/rate-limit/clients", func(w http.ResponseWriter, r *http.Request) {
			clients := rateLimitMiddleware.GetActiveClients()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"activeClients": %d, "timestamp": "%s"}`, len(clients), time.Now().Format(time.RFC3339))
		})

		logger.Println("🚦 限流监控端点(Dev): http://localhost:9090/debug/rate-limit/stats")
	}

	// 设置开发工具路由 (仅开发模式，无认证要求)
	if !authOnlyMode {
		devToolsHandler.SetupRoutes(r)
	}

	// 📎 BFF 认证路由（生产态登录/会话管理） - 不要求已有Authorization
	bffHandler := authbff.NewBFFHandler(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, logger, devMode, auditLogger)
	bffHandler.SetupRoutes(r)

	if !authOnlyMode {
		// 为需要认证的API路由创建子路由器
		r.Group(func(r chi.Router) {
			r.Use(restAuthMiddleware.Middleware()) // JWT认证和权限验证中间件
			// 设置组织相关路由 (需要认证)
			orgHandler.SetupRoutes(r)
			// 设置运维管理路由 (需要认证)
			operationalHandler.SetupRoutes(r)
		})
	}

	// 服务启动
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 启动运维调度器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if !authOnlyMode {
		operationalScheduler.Start(ctx)
		logger.Println("✅ 运维任务调度器已启动")
	}

	// 优雅关闭
	go func() {
		logger.Printf("🎯 组织命令服务启动在端口 %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("服务启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("🛑 正在关闭服务...")

	if !authOnlyMode {
		// 停止级联更新服务
		cascadeService.Stop()
		logger.Println("✅ 级联更新服务已停止")

		// 停止运维调度器
		operationalScheduler.Stop()
		logger.Println("✅ 运维任务调度器已停止")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Printf("服务关闭错误: %v", err)
	} else {
		logger.Println("✅ 服务已安全关闭")
	}
}
