package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	authbff "cube-castle/cmd/hrms-server/command/internal/authbff"
	outbox "cube-castle/cmd/hrms-server/command/internal/outbox"
	publicgraphql "cube-castle/cmd/hrms-server/query/publicgraphql"
	auth "cube-castle/internal/auth"
	config "cube-castle/internal/config"
	health "cube-castle/internal/monitoring/health"
	organization "cube-castle/internal/organization"
	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// v9RedisChecker implements health.Checker for go-redis/v9 client.
type v9RedisChecker struct {
	Name   string
	Client *redis.Client
}

func (c *v9RedisChecker) Check(ctx context.Context) health.HealthCheck {
	start := time.Now()
	check := health.HealthCheck{
		Name: c.Name,
	}
	if c.Client == nil {
		check.Status = health.StatusDegraded
		check.Message = "Redis client not configured"
		check.Duration = time.Since(start)
		return check
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_, err := c.Client.Ping(ctx).Result()
	check.Duration = time.Since(start)
	if err != nil {
		check.Status = health.StatusUnhealthy
		check.Message = "Redis ping failed: " + err.Error()
		return check
	}
	check.Status = health.StatusHealthy
	check.Message = "Redis connection healthy"
	return check
}

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
		dbClient    *database.Database
		sqlDB       *sql.DB
		outboxRepo  database.OutboxRepository
		redisClient *redis.Client
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

		redisClient = openRedis(commandLogger)
		if redisClient != nil {
			defer redisClient.Close()
		}
		// 预热 DB 直方图时间序列，便于在 /metrics 中可见（不会影响统计意义）
		database.ObserveQueryDuration("command-service", "startup", time.Duration(0))
		// 周期性上报数据库连接池状态（开发/CI 建议开启；生产可按需调整频率或迁移到运维任务）
		go func(db *database.Database) {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				db.RecordConnectionStats("command-service")
			}
		}(dbClient)
	} else {
		commandLogger.Info("🟡 AUTH_ONLY_MODE=true：跳过数据库连接，仅启用 BFF /auth 与 /.well-known 端点")
	}

	eventBus := eventbus.NewMemoryEventBus(commandLogger, nil)
	commandLogger.Info("✅ 事件总线初始化完成（内存实现）")

	var (
		dispatcher            *outbox.Dispatcher
		assignmentCache       organization.AssignmentFacade
		queryRepo             *organization.QueryRepository
		schedulerConfigResult config.SchedulerConfigResult
		schedulerConfigLoaded bool
	)
	if !authOnlyMode {
		schedulerConfigResult = config.GetSchedulerConfig()
		schedulerConfigLoaded = true
		commandLogger.WithFields(pkglogger.Fields{
			"sources": strings.Join(schedulerConfigResult.Metadata.Sources, ","),
		}).Info("✅ 调度配置加载完成")
		if len(schedulerConfigResult.Metadata.Overrides) > 0 {
			commandLogger.WithFields(pkglogger.Fields{
				"overrides": schedulerConfigResult.Metadata.Overrides,
			}).Debug("调度配置覆盖详情")
		}
		if schedulerConfigResult.Metadata.ValidationError != nil {
			commandLogger.Errorf("[FATAL] 调度配置校验失败: %v", schedulerConfigResult.Metadata.ValidationError)
			os.Exit(1)
		}

		outboxCfg, err := outbox.LoadConfig()
		if err != nil {
			commandLogger.Errorf("[FATAL] Outbox dispatcher 配置无效: %v", err)
			os.Exit(1)
		}

		queryRepo = organization.NewQueryRepository(sqlDB, redisClient, commandLogger, organization.DefaultAuditHistoryConfig())
		assignmentCache = organization.NewAssignmentFacade(queryRepo, redisClient, commandLogger, time.Minute)

		dispatcher = outbox.NewDispatcher(outboxCfg, outboxRepo, eventBus, commandLogger, prometheus.DefaultRegisterer, dbClient.WithTx, assignmentCache)
		commandLogger.Infof("✅ Outbox dispatcher 预备就绪 (interval=%s batch=%d maxRetry=%d)", outboxCfg.PollInterval, outboxCfg.BatchSize, outboxCfg.MaxRetry)
	}

	var (
		orgModule         *organization.CommandModule
		commandHandlers   organization.CommandHandlers
		auditLogger       *organization.AuditLogger
		moduleMiddlewares = organization.NewCommandMiddlewares(commandLogger)
		devToolsHandler   *organization.DevToolsHandler
	)
	if !authOnlyMode {
		var err error
		orgModule, err = organization.NewCommandModule(organization.CommandModuleDeps{
			DB:              sqlDB,
			Logger:          commandLogger,
			CascadeMaxDepth: 4,
			SchedulerConfig: func() *config.SchedulerConfig {
				if schedulerConfigLoaded {
					return schedulerConfigResult.Config
				}
				return nil
			}(),
			OutboxRepo: outboxRepo,
		})
		if err != nil {
			commandLogger.Errorf("[FATAL] 初始化组织模块失败: %v", err)
			os.Exit(1)
		}

		orgModule.Services.Cascade.Start()
		auditLogger = orgModule.AuditLogger
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
	performanceMiddleware := moduleMiddlewares.Performance
	rateLimitMiddleware := moduleMiddlewares.RateLimit

	// 初始化时态服务
	var (
		orgHandler         *organization.OrganizationHandler
		positionHandler    *organization.PositionHandler
		jobCatalogHandler  *organization.JobCatalogHandler
		operationalHandler *organization.OperationalHandler
	)
	if !authOnlyMode {
		commandHandlers = orgModule.NewHandlers(organization.CommandHandlerDeps{
			JWTMiddleware:       jwtMiddleware,
			RateLimitMiddleware: rateLimitMiddleware,
			Logger:              commandLogger,
			DevMode:             devMode,
		})
		orgHandler = commandHandlers.Organization
		positionHandler = commandHandlers.Position
		jobCatalogHandler = commandHandlers.JobCatalog
		operationalHandler = commandHandlers.Operational
		devToolsHandler = commandHandlers.DevTools
	} else {
		devToolsHandler = organization.NewDevToolsHandler(sqlDB, jwtMiddleware, commandLogger, devMode)
	}

	// 设置路由
	r := chi.NewRouter()

	// 基础中间件链 (无认证要求的中间件)
	r.Use(organization.RequestIDMiddleware)   // 请求追踪中间件
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

	// NotFound 记录，便于排查路由冲突
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		commandLogger.WithFields(pkglogger.Fields{
			"path":   req.URL.Path,
			"method": req.Method,
		}).Warn("Route not found")
		http.NotFound(w, req)
	})

	// 健康检查（统一实现）
	{
		hm := health.NewHealthManager("command", "v1")
		if sqlDB != nil {
			hm.AddChecker(&health.PostgreSQLChecker{Name: "postgres", DB: sqlDB})
		}
		if redisClient != nil {
			hm.AddChecker(&v9RedisChecker{Name: "redis", Client: redisClient})
		}
		r.Get("/health", hm.Handler())
	}

	// Prometheus metrics 端点（无需认证，供监控系统采集）
	if !authOnlyMode {
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

	// GraphQL 查询路由（单体合流挂载）
	if !authOnlyMode {
		gqlHandler, graphiqlHandler, err := publicgraphql.BuildHandlers(sqlDB, queryRepo, assignmentCache, commandLogger, devMode)
		if err != nil {
			commandLogger.Errorf("[FATAL] 构建 GraphQL 处理器失败: %v", err)
			os.Exit(1)
		}
		// Wrapper with structured logging, registered on multiple method/path variants to avoid slashes mismatch.
		graphQLServe := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			commandLogger.WithFields(pkglogger.Fields{
				"path":   req.URL.Path,
				"method": req.Method,
			}).Info("GraphQL handler invoked")
			gqlHandler.ServeHTTP(w, req)
		})
		// POST is the primary method
		r.Post("/graphql", graphQLServe)
		r.Post("/graphql/", graphQLServe) // tolerate trailing slash
		// Allow GET for simple probes/dev tools
		r.Get("/graphql", graphQLServe)
		r.Get("/graphql/", graphQLServe)
		// Fallback: handle any other method variants to avoid router mismatch in local/dev
		r.Handle("/graphql", graphQLServe)
		r.Handle("/graphql/", graphQLServe)
		if devMode && graphiqlHandler != nil {
			r.Get("/graphiql", func(w http.ResponseWriter, req *http.Request) {
				graphiqlHandler.ServeHTTP(w, req)
			})
			r.Get("/graphiql/", func(w http.ResponseWriter, req *http.Request) {
				graphiqlHandler.ServeHTTP(w, req)
			})
		}
		commandLogger.Info("🔗 GraphQL 查询端点已挂载到单体进程: /graphql（/graphiql in dev）")
	}

	// 路由枚举（调试）
	if devMode {
		_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			commandLogger.WithFields(pkglogger.Fields{
				"method": method,
				"route":  route,
			}).Info("Route registered")
			return nil
		})
	}

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
	// Runtime self-guard: forbid 8090 in monolith mode (compare numerically to avoid hardcoded string literal)
	portNum, _ := strconv.Atoi(strings.TrimPrefix(port, ":"))
	if portNum == 8090 {
		commandLogger.Errorf("[FATAL] 端口 8090 已在单体模式下禁用，请使用默认 9090；如需本地排障，请设置 ENABLE_LEGACY_DUAL_SERVICE=true 并仅在本地运行（CI 禁止）。")
		os.Exit(1)
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
		orgModule.Services.Scheduler.Start(ctx)
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
		orgModule.Services.Cascade.Stop()
		commandLogger.Info("✅ 级联更新服务已停止")

		// 停止运维调度器
		orgModule.Services.Scheduler.Stop()
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

func openRedis(logger pkglogger.Logger) *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	client := redis.NewClient(&redis.Options{Addr: addr})
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.WithFields(pkglogger.Fields{
			"component": "redis",
			"error":     err,
		}).Warn("Redis连接失败，将跳过缓存刷新")
		client.Close()
		return nil
	}
	logger.WithFields(pkglogger.Fields{
		"component": "redis",
		"address":   addr,
	}).Info("✅ Redis连接成功")
	return client
}
