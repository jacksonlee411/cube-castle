package app

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	graphqlruntime "cube-castle/cmd/hrms-server/query/internal/graphql"
	graphqlresolver "cube-castle/cmd/hrms-server/query/internal/graphql/resolver"
	"cube-castle/internal/auth"
	"cube-castle/internal/config"
	schemaLoader "cube-castle/internal/graphql"
	requestMiddleware "cube-castle/internal/middleware"
	health "cube-castle/internal/monitoring/health"
	organization "cube-castle/internal/organization"
	"cube-castle/pkg/database"
	pkglogger "cube-castle/pkg/logger"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	logger      pkglogger.Logger
	db          *sql.DB
	dbClient    *database.Database
	redisClient *redis.Client
	server      *http.Server
}

func (a *Application) log(operation string, fields pkglogger.Fields) pkglogger.Logger {
	log := a.logger
	if operation != "" {
		log = log.WithFields(pkglogger.Fields{"operation": operation})
	}
	if len(fields) == 0 {
		return log
	}
	return log.WithFields(fields)
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests handled by the PostgreSQL GraphQL service.",
		},
		[]string{"method", "route", "status"},
	)
	organizationOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "organization_operations_total",
			Help: "Count of organization operations processed via GraphQL endpoints.",
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(organizationOperationsTotal)
	// 预注册GraphQL请求标签，确保指标在无流量时也可见
	organizationOperationsTotal.WithLabelValues("graphql_query").Add(0)
}

func Run() error {
	baseLogger := pkglogger.NewLogger(
		pkglogger.WithWriter(os.Stdout),
		pkglogger.WithLevel(pkglogger.LevelInfo),
		pkglogger.WithCallerSkip(1),
	).WithFields(pkglogger.Fields{
		"service":   "query",
		"component": "query-app",
	})
	app := &Application{logger: baseLogger}
	return app.run()
}

func (a *Application) run() error {
	a.log("startup", nil).Info("🚀 启动PostgreSQL原生GraphQL服务")

	var err error
	a.dbClient, err = a.openDatabase()
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}
	a.db = a.dbClient.GetDB()

	a.redisClient = a.openRedis()

	auditConfig := loadAuditHistoryConfig()
	repo := organization.NewQueryRepository(a.db, a.redisClient, a.logger, auditConfig)
	assignmentFacade := organization.NewAssignmentFacade(repo, a.redisClient, a.logger, time.Minute)
	a.log("audit.config", pkglogger.Fields{
		"strictValidation": auditConfig.StrictValidation,
		"allowFallback":    auditConfig.AllowFallback,
		"circuitThreshold": auditConfig.CircuitBreakerThreshold,
		"legacyMode":       auditConfig.LegacyMode,
	}).Info("⚙️ 审计历史配置加载完成")

	a.server, err = a.buildServer(repo, assignmentFacade)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		a.log("shutdown", nil).Info("🛑 正在关闭PostgreSQL GraphQL服务...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.log("shutdown", pkglogger.Fields{"error": err}).Error("❌ 服务关闭失败")
		}
	}()

	port := a.server.Addr
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}
	a.log("startup", pkglogger.Fields{"port": port}).Info("🚀 PostgreSQL原生GraphQL服务启动完成")
	a.log("startup", pkglogger.Fields{"url": "http://localhost:" + port + "/graphiql"}).Info("🔗 GraphiQL界面")
	a.log("startup", pkglogger.Fields{"url": "http://localhost:" + port + "/graphql"}).Info("🔗 GraphQL端点")
	a.log("startup", pkglogger.Fields{"database": "postgres"}).Info("💾 数据库: PostgreSQL (原生优化)")
	a.log("startup", pkglogger.Fields{"mode": "aggressive"}).Info("⚡ 性能模式: 激进优化")

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}

	a.log("shutdown", nil).Info("✅ PostgreSQL GraphQL服务已安全关闭")
	return nil
}

func (a *Application) openDatabase() (*database.Database, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "cubecastle")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := database.NewDatabaseWithConfig(database.ConnectionConfig{
		DSN:         dsn,
		ServiceName: "query-service",
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	// 注册 DB 指标并周期性上报连接池状态
	database.RegisterMetrics(prometheus.DefaultRegisterer)
	// 预热 DB 直方图时间序列，便于在 /metrics 中可见（不会影响统计意义）
	database.ObserveQueryDuration("query-service", "startup", time.Duration(0))
	go func(dbc *database.Database) {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			dbc.RecordConnectionStats("query-service")
		}
	}(db)

	a.log("database.connect", pkglogger.Fields{
		"host":     dbHost,
		"port":     dbPort,
		"database": dbName,
	}).Info("✅ PostgreSQL连接成功")
	return db, nil
}

func (a *Application) openRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		DB:   0,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		a.log("redis.connect", pkglogger.Fields{"error": err}).Warn("Redis连接失败，将不使用缓存")
		return nil
	}

	a.log("redis.connect", pkglogger.Fields{
		"address": client.Options().Addr,
	}).Info("✅ Redis连接成功")
	return client
}

func (a *Application) buildServer(repo *organization.QueryRepository, assignmentFacade organization.AssignmentFacade) (*http.Server, error) {
	jwtConfig := config.GetJWTConfig()
	// 默认禁用开发模式；仅在本地/开发容器中通过环境变量 DEV_MODE=true 显式开启
	devMode := getEnv("DEV_MODE", "false") == "true"

	var pubPEM []byte
	if jwtConfig.HasPublicKey() {
		if b, err := os.ReadFile(jwtConfig.PublicKeyPath); err == nil {
			pubPEM = b
		} else {
			return nil, fmt.Errorf("读取查询服务公钥失败: %w", err)
		}
	}

	if jwtConfig.JWKSUrl == "" && pubPEM == nil {
		return nil, fmt.Errorf("查询服务启用RS256必须配置 JWT_JWKS_URL 或 JWT_PUBLIC_KEY_PATH")
	}

	jwtMiddleware := auth.NewJWTMiddlewareWithOptions(jwtConfig.Secret, jwtConfig.Issuer, jwtConfig.Audience, auth.Options{
		Alg:          jwtConfig.Algorithm,
		JWKSURL:      jwtConfig.JWKSUrl,
		PublicKeyPEM: pubPEM,
		ClockSkew:    jwtConfig.AllowedClockSkew,
	})

	authLogger := a.logger.WithFields(pkglogger.Fields{"component": "query-auth"})
	permissionChecker := auth.NewPBACPermissionChecker(a.db, authLogger)
	graphqlMiddleware := auth.NewGraphQLPermissionMiddleware(jwtMiddleware, permissionChecker, authLogger, devMode)
	a.log("graphql.init", pkglogger.Fields{
		"devMode":   devMode,
		"algorithm": jwtConfig.Algorithm,
		"issuer":    jwtConfig.Issuer,
		"audience":  jwtConfig.Audience,
	}).Info("🔐 JWT认证初始化完成")

	gqlResolver := organization.NewQueryResolver(repo, assignmentFacade, a.logger, graphqlMiddleware)
	gqlgenResolver := graphqlresolver.New(gqlResolver)
	executableSchema := graphqlruntime.NewExecutableSchema(graphqlruntime.Config{
		Resolvers: gqlgenResolver,
	})
	graphqlServer := handler.NewDefaultServer(executableSchema)
	schemaPath := schemaLoader.GetDefaultSchemaPath()
	a.log("graphql.schema", pkglogger.Fields{"path": schemaPath}).Info("✅ GraphQL Schema compiled from single source via gqlgen")

	port := getEnv("PORT", "8090")
	router := a.buildRouter(graphqlServer, graphqlMiddleware, devMode, port)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

func (a *Application) buildRouter(graphqlServer http.Handler, permission *auth.GraphQLPermissionMiddleware, devMode bool, port string) http.Handler {
	r := chi.NewRouter()
	r.Use(requestMiddleware.RequestIDMiddleware)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   resolveQueryAllowedOrigins(port),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(metricsMiddleware)

	envelopeMiddleware := requestMiddleware.NewGraphQLEnvelopeMiddleware()
	baseGraphQLHandler := envelopeMiddleware.Middleware()(permission.Middleware()(graphqlServer))
	graphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationOperationsTotal.WithLabelValues("graphql_query").Inc()
		baseGraphQLHandler.ServeHTTP(w, r)
	})
	r.Handle("/graphql", graphqlHandler)

	if devMode {
		r.Get("/graphiql", func(w http.ResponseWriter, _ *http.Request) {
			html := graphiqlPage()
			if _, err := w.Write([]byte(html)); err != nil {
				http.Error(w, "failed to write GraphiQL page", http.StatusInternalServerError)
			}
		})
	}

	// 健康检查（统一实现）
	{
		hm := health.NewHealthManager("query", "v1")
		if a.db != nil {
			hm.AddChecker(&health.PostgreSQLChecker{Name: "postgres", DB: a.db})
		}
		if a.redisClient != nil {
			hm.AddChecker(&v9RedisChecker{Name: "redis", Client: a.redisClient})
		}
		r.Get("/health", hm.Handler())
	}

	r.Handle("/metrics", promhttp.Handler())

	return r
}

func resolveQueryAllowedOrigins(port string) []string {
	scheme := firstNonEmpty(os.Getenv("QUERY_BASE_SCHEME"), os.Getenv("COMMAND_BASE_SCHEME"), "http")
	host := firstNonEmpty(os.Getenv("QUERY_BASE_HOST"), os.Getenv("COMMAND_BASE_HOST"), "127.0.0.1")
	defaultOrigin := config.BuildOrigin(scheme, host, port)
	return config.ResolveAllowedOrigins("QUERY_ALLOWED_ORIGINS", "COMMAND_ALLOWED_ORIGINS", []string{defaultOrigin})
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapper, r)
		status := wrapper.Status()
		if status == 0 {
			status = http.StatusOK
		}
		// 路由模板化，避免基数爆炸
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}
		httpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(status)).Inc()
	})
}

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

func graphiqlPage() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>GraphiQL - PostgreSQL Native</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.css" />
    <style>
        body { height: 100%; margin: 0; width: 100%; overflow: hidden; }
        #graphiql { height: 100vh; }
        .graphiql-container { background: #1a1a1a; }
    </style>
</head>
<body>
    <div id="graphiql">Loading PostgreSQL GraphQL...</div>
    <script crossorigin src="https://unpkg.com/react@18/umd/react.development.js"></script>
    <script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.development.js"></script>
    <script crossorigin src="https://cdn.jsdelivr.net/npm/graphiql@2.4.7/graphiql.min.js"></script>
    <script>
        const fetcher = GraphiQL.createFetcher({ url: '/graphql' });
        const root = ReactDOM.createRoot(document.getElementById('graphiql'));
        root.render(React.createElement(GraphiQL, {
            fetcher,
            defaultQuery: '# PostgreSQL原生GraphQL查询\\n# 高性能时态查询示例\\n\\nquery {\\n  organizations(first: 10) {\\n    code\\n    name\\n    status\\n    effective_date\\n    is_current\\n  }\\n}'
        }));
    </script>
</body>
</html>`
}
