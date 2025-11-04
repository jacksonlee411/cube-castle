package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cube-castle/cmd/hrms-server/query/internal/graphql"
	"cube-castle/cmd/hrms-server/query/internal/repository"
	"cube-castle/internal/auth"
	"cube-castle/internal/config"
	schemaLoader "cube-castle/internal/graphql"
	requestMiddleware "cube-castle/internal/middleware"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	graphqlgo "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

type Application struct {
	logger      pkglogger.Logger
	db          *sql.DB
	redisClient *redis.Client
	server      *http.Server
}

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests handled by the PostgreSQL GraphQL service.",
		},
		[]string{"method", "path", "status"},
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
	)
	app := &Application{
		logger: baseLogger.WithFields(pkglogger.Fields{
			"service":   "query",
			"component": "bootstrap",
		}),
	}
	return app.run()
}

func (a *Application) run() error {
	a.logger.Info("🚀 启动PostgreSQL原生GraphQL服务")

	var err error
	a.db, err = a.openDatabase()
	if err != nil {
		return fmt.Errorf("database init: %w", err)
	}

	a.redisClient = a.openRedis()

	auditConfig := loadAuditHistoryConfig()
	repoLogger := a.logger.WithFields(pkglogger.Fields{
		"component": "repository",
	})
	repo := repository.NewPostgreSQLRepository(a.db, a.redisClient, repoLogger, auditConfig)
	a.logger.Infof("⚙️ 审计历史配置: strictValidation=%v, allowFallback=%v, circuitThreshold=%d, legacyMode=%v",
		auditConfig.StrictValidation, auditConfig.AllowFallback, auditConfig.CircuitBreakerThreshold, auditConfig.LegacyMode)

	a.server, err = a.buildServer(repo)
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		a.logger.Info("🛑 正在关闭PostgreSQL GraphQL服务...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.WithFields(pkglogger.Fields{"error": err}).Error("❌ 服务关闭失败")
		}
	}()

	port := a.server.Addr
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}
	a.logger.Infof("🚀 PostgreSQL原生GraphQL服务启动在端口 :%s", port)
	a.logger.Info("🔗 GraphiQL界面: http://localhost:" + port + "/graphiql")
	a.logger.Info("🔗 GraphQL端点: http://localhost:" + port + "/graphql")
	a.logger.Info("💾 数据库: PostgreSQL (原生优化)")
	a.logger.Info("⚡ 性能模式: 激进优化")

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}

	a.logger.Info("✅ PostgreSQL GraphQL服务已安全关闭")
	return nil
}

func (a *Application) openDatabase() (*sql.DB, error) {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "cubecastle")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	a.logger.Info("✅ PostgreSQL连接成功")
	return db, nil
}

func (a *Application) openRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		DB:   0,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		a.logger.WithFields(pkglogger.Fields{"error": err}).Warn("Redis连接失败，将不使用缓存")
		return nil
	}

	a.logger.Info("✅ Redis连接成功")
	return client
}

func (a *Application) buildServer(repo *repository.PostgreSQLRepository) (*http.Server, error) {
	jwtConfig := config.GetJWTConfig()
	devMode := getEnv("DEV_MODE", "true") == "true"

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

	permissionChecker := auth.NewPBACPermissionChecker(a.db, a.logger)
	graphqlMiddleware := auth.NewGraphQLPermissionMiddleware(jwtMiddleware, permissionChecker, a.logger, devMode)
	a.logger.Infof("🔐 JWT认证初始化完成 (开发模式: %v, Alg=%s, Issuer=%s, Audience=%s)", devMode, jwtConfig.Algorithm, jwtConfig.Issuer, jwtConfig.Audience)

	resolver := graphql.NewResolver(repo, a.logger.WithFields(pkglogger.Fields{"component": "graphqlResolver"}), graphqlMiddleware)
	schemaPath := schemaLoader.GetDefaultSchemaPath()
	schemaString := schemaLoader.MustLoadSchema(schemaPath)
	schema := graphqlgo.MustParseSchema(schemaString, resolver)
	a.logger.Infof("✅ GraphQL Schema loaded from single source: %s", schemaPath)

	router := a.buildRouter(schema, graphqlMiddleware, devMode)

	port := getEnv("PORT", "8090")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

func (a *Application) buildRouter(schema *graphqlgo.Schema, permission *auth.GraphQLPermissionMiddleware, devMode bool) http.Handler {
	r := chi.NewRouter()
	r.Use(requestMiddleware.RequestIDMiddleware)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(metricsMiddleware)

	envelopeMiddleware := requestMiddleware.NewGraphQLEnvelopeMiddleware()
	relayHandler := &relay.Handler{Schema: schema}
	baseGraphQLHandler := envelopeMiddleware.Middleware()(permission.Middleware()(relayHandler))
	graphqlHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationOperationsTotal.WithLabelValues("graphql_query").Inc()
		baseGraphQLHandler.ServeHTTP(w, r)
	})
	r.Handle("/graphql", graphqlHandler)

	if devMode {
		r.Get("/graphiql", func(w http.ResponseWriter, r *http.Request) {
			html := graphiqlPage()
			if _, err := w.Write([]byte(html)); err != nil {
				http.Error(w, "failed to write GraphiQL page", http.StatusInternalServerError)
			}
		})
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]interface{}{
			"status":      "healthy",
			"service":     "postgresql-graphql",
			"timestamp":   time.Now(),
			"database":    "postgresql",
			"performance": "optimized",
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "failed to encode health response", http.StatusInternalServerError)
		}
	})

	r.Handle("/metrics", promhttp.Handler())

	return r
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := chiMiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrapper, r)
		status := wrapper.Status()
		if status == 0 {
			status = http.StatusOK
		}
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(status)).Inc()
	})
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
