package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cube-castle/cmd/hrms-server/command/internal/audit"
	"cube-castle/cmd/hrms-server/command/internal/authbff"
	"cube-castle/cmd/hrms-server/command/internal/handlers"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/outbox"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/cmd/hrms-server/command/internal/utils"
	"cube-castle/cmd/hrms-server/command/internal/validators"
	auth "cube-castle/internal/auth"
	config "cube-castle/internal/config"
	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	baseLogger := pkglogger.NewLogger(
		pkglogger.WithWriter(os.Stdout),
		pkglogger.WithLevelString(os.Getenv("COMMAND_LOG_LEVEL")),
		pkglogger.WithCallerSkip(1),
	)
	commandLogger := baseLogger.WithFields(pkglogger.Fields{
		"service":   "command",
		"component": "bootstrap",
	})
	commandLogger.Info("🚀 启动组织命令服务...")
	authOnlyMode := os.Getenv("AUTH_ONLY_MODE") == "true"

	var (
		dbClient   *database.Database
		sqlDB      *sql.DB
		outboxRepo database.OutboxRepository
	)
	if !authOnlyMode {
		// 数据库连接
		dbURL := os.Getenv("DATABASE_URL")
		if dbURL == "" {
			dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
		}

		var err error
		dbClient, err = database.NewDatabaseWithConfig(database.ConnectionConfig{
			DSN:         dbURL,
			ServiceName: "command-service",
		})
		if err != nil {
			commandLogger.Errorf("数据库连接失败: %v", err)
			os.Exit(1)
		}
		defer dbClient.Close()
		sqlDB = dbClient.GetDB()
		database.RegisterMetrics(prometheus.DefaultRegisterer)

		// 验证数据库连接
		if err := sqlDB.Ping(); err != nil {
			commandLogger.Errorf("数据库连接验证失败: %v", err)
			os.Exit(1)
		}

		commandLogger.Info("✅ 数据库连接成功")
		outboxRepo = database.NewOutboxRepository(dbClient)
		commandLogger.Infof("✅ Outbox 仓储初始化完成（impl=%T）", outboxRepo)
	} else {
		commandLogger.Info("🟡 AUTH_ONLY_MODE=true：跳过数据库连接，仅启用 BFF /auth 与 /.well-known 端点")
	}

	eventBus := eventbus.NewMemoryEventBus(commandLogger, nil)
	commandLogger.Info("✅ 事件总线初始化完成（内存实现）")

	var dispatcher *outbox.Dispatcher
	if !authOnlyMode {
		outboxCfg, err := outbox.LoadConfig()
		if err != nil {
			commandLogger.Errorf("[FATAL] Outbox dispatcher 配置无效: %v", err)
			os.Exit(1)
		}
		dispatcher = outbox.NewDispatcher(outboxCfg, outboxRepo, eventBus, commandLogger, prometheus.DefaultRegisterer, dbClient.WithTx)
		commandLogger.Infof("✅ Outbox dispatcher 预备就绪 (interval=%s batch=%d maxRetry=%d)", outboxCfg.PollInterval, outboxCfg.BatchSize, outboxCfg.MaxRetry)
	}

	var (
		orgRepo                *repository.OrganizationRepository
		jobCatalogRepo         *repository.JobCatalogRepository
		positionRepo           *repository.PositionRepository
		positionAssignmentRepo *repository.PositionAssignmentRepository
		hierarchyRepo          *repository.HierarchyRepository
		cascadeService         *services.CascadeUpdateService
		auditLogger            *audit.AuditLogger
		businessValidator      *validators.BusinessRuleValidator
	)
	if !authOnlyMode {
		// 初始化仓储层
		orgRepo = repository.NewOrganizationRepository(sqlDB, commandLogger)
		jobCatalogRepo = repository.NewJobCatalogRepository(sqlDB, commandLogger)
		positionRepo = repository.NewPositionRepository(sqlDB, commandLogger)
		positionAssignmentRepo = repository.NewPositionAssignmentRepository(sqlDB, commandLogger)
		hierarchyRepo = repository.NewHierarchyRepository(sqlDB, commandLogger)

		// 初始化业务服务层
		cascadeService = services.NewCascadeUpdateService(hierarchyRepo, 4, commandLogger)
		businessValidator = validators.NewBusinessRuleValidator(hierarchyRepo, orgRepo, commandLogger)
		auditLogger = audit.NewAuditLogger(sqlDB, commandLogger)

		// 启动级联更新服务
		cascadeService.Start()
		commandLogger.Info("✅ 级联更新服务已启动")
		commandLogger.Info("✅ 结构化审计日志系统已初始化")
		commandLogger.Info("✅ Prometheus指标收集系统已初始化")
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
			commandLogger.Errorf("[FATAL] 无法读取JWT公钥 (%s): %v", jwtConfig.PublicKeyPath, err)
			os.Exit(1)
		}
	}
	if !jwtConfig.HasPrivateKey() {
		commandLogger.Error("[FATAL] 启用了RS256但未配置JWT_PRIVATE_KEY_PATH。请运行 make jwt-dev-setup 或提供正式私钥文件。")
		os.Exit(1)
	}
	if b, err := os.ReadFile(jwtConfig.PrivateKeyPath); err == nil {
		privPEM = b
	} else {
		commandLogger.Errorf("[FATAL] 无法读取JWT私钥 (%s): %v", jwtConfig.PrivateKeyPath, err)
		os.Exit(1)
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
		permissionChecker := auth.NewPBACPermissionChecker(sqlDB, commandLogger)
		restAuthMiddleware = auth.NewRESTPermissionMiddleware(
			jwtMiddleware,
			permissionChecker,
			commandLogger,
			devMode,
		)
	}

	commandLogger.Infof("🔐 JWT认证初始化完成 (开发模式: %v, Alg=%s, Issuer=%s, Audience=%s)", devMode, jwtConfig.Algorithm, jwtConfig.Issuer, jwtConfig.Audience)

	// 初始化中间件
	performanceMiddleware := middleware.NewPerformanceMiddleware(commandLogger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(middleware.DefaultRateLimitConfig, commandLogger)

	// 初始化时态服务
	var temporalService *services.TemporalService
	if !authOnlyMode {
		temporalService = services.NewTemporalService(sqlDB, commandLogger, orgRepo)
	}

	// 初始化监控服务
	var temporalMonitor *services.TemporalMonitor
	if !authOnlyMode {
		temporalMonitor = services.NewTemporalMonitor(sqlDB, commandLogger)
	}

	// 初始化运维调度器占位
	var operationalScheduler *services.OperationalScheduler

	// 初始化时态时间轴管理器
	var timelineManager *repository.TemporalTimelineManager
	if !authOnlyMode {
		timelineManager = repository.NewTemporalTimelineManager(sqlDB, commandLogger)
	}

	// 初始化处理器
	var (
		orgHandler         *handlers.OrganizationHandler
		positionHandler    *handlers.PositionHandler
		jobCatalogHandler  *handlers.JobCatalogHandler
		devToolsHandler    *handlers.DevToolsHandler
		operationalHandler *handlers.OperationalHandler
	)
	if !authOnlyMode {
		positionService := services.NewPositionService(positionRepo, positionAssignmentRepo, jobCatalogRepo, orgRepo, auditLogger, commandLogger)
		jobCatalogService := services.NewJobCatalogService(jobCatalogRepo, auditLogger, commandLogger)
		operationalScheduler = services.NewOperationalScheduler(sqlDB, commandLogger, temporalMonitor, positionService)

		orgHandler = handlers.NewOrganizationHandler(orgRepo, temporalService, auditLogger, commandLogger, timelineManager, hierarchyRepo, businessValidator)
		positionHandler = handlers.NewPositionHandler(positionService, commandLogger)
		jobCatalogHandler = handlers.NewJobCatalogHandler(jobCatalogService, commandLogger)
		operationalHandler = handlers.NewOperationalHandler(temporalMonitor, operationalScheduler, rateLimitMiddleware, commandLogger)
	}
	// 开发工具路由即使在 authOnly 模式下也允许初始化（内部会根据 devMode 控制）
	devToolsHandler = handlers.NewDevToolsHandler(jwtMiddleware, commandLogger, devMode, sqlDB)

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

	// Prometheus metrics 端点（无需认证，供监控系统采集）
	if !authOnlyMode {
		// 确保 metrics 已注册
		utils.RecordHTTPRequest("GET", "/metrics", 200) // 触发初始化
		r.Handle("/metrics", promhttp.Handler())
		commandLogger.Info("📊 Prometheus metrics 端点: http://localhost:9090/metrics")
	}

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

		commandLogger.Info("🚦 限流监控端点(Dev): http://localhost:9090/debug/rate-limit/stats")
	}

	// 设置开发工具路由 (仅开发模式，无认证要求)
	if !authOnlyMode {
		devToolsHandler.SetupRoutes(r)
	}

	// 📎 BFF 认证路由（生产态登录/会话管理） - 不要求已有Authorization
	bffHandler := authbff.NewBFFHandler(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, commandLogger, devMode, auditLogger)
	bffHandler.SetupRoutes(r)

	if !authOnlyMode {
		// 为需要认证的API路由创建子路由器
		r.Group(func(r chi.Router) {
			r.Use(restAuthMiddleware.Middleware()) // JWT认证和权限验证中间件
			// 设置组织相关路由 (需要认证)
			if positionHandler != nil {
				positionHandler.SetupRoutes(r)
			}
			if jobCatalogHandler != nil {
				jobCatalogHandler.SetupRoutes(r)
			}
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
	if !authOnlyMode && dispatcher != nil {
		if err := dispatcher.Start(ctx); err != nil {
			commandLogger.Errorf("[FATAL] Outbox dispatcher 启动失败: %v", err)
			os.Exit(1)
		}
		commandLogger.Info("✅ Outbox dispatcher 已启动")
	}
	if !authOnlyMode {
		operationalScheduler.Start(ctx)
		commandLogger.Info("✅ 运维任务调度器已启动")
	}

	// 优雅关闭
	go func() {
		commandLogger.Infof("🎯 组织命令服务启动在端口 %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			commandLogger.Errorf("服务启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	commandLogger.Info("🛑 正在关闭服务...")
	cancel()

	if !authOnlyMode {
		// 停止级联更新服务
		cascadeService.Stop()
		commandLogger.Info("✅ 级联更新服务已停止")

		// 停止运维调度器
		operationalScheduler.Stop()
		commandLogger.Info("✅ 运维任务调度器已停止")

		if dispatcher != nil {
			if err := dispatcher.Stop(); err != nil {
				commandLogger.Errorf("outbox dispatcher 停止失败: %v", err)
			} else {
				commandLogger.Info("✅ Outbox dispatcher 已停止")
			}
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		commandLogger.Errorf("服务关闭错误: %v", err)
	} else {
		commandLogger.Info("✅ 服务已安全关闭")
	}
}
