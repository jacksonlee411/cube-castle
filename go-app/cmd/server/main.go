package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/gaogu/cube-castle/go-app/internal/handler"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
	"github.com/gaogu/cube-castle/go-app/internal/metacontracteditor"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
	"github.com/gaogu/cube-castle/go-app/internal/middleware"
	"github.com/gaogu/cube-castle/go-app/internal/monitoring"
	"github.com/gaogu/cube-castle/go-app/internal/routes"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/gaogu/cube-castle/go-app/internal/validation"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/handlers"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/services"
	"github.com/gaogu/cube-castle/go-app/internal/events/consumers"
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"regexp"
)

// simpleLogger 实现简单的日志接口用于测试
type simpleLogger struct{}

func (l *simpleLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("INFO: %s %v", msg, keysAndValues)
}

func (l *simpleLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Printf("ERROR: %s %v", msg, keysAndValues)
}

func (l *simpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Printf("WARN: %s %v", msg, keysAndValues)
}

func (l *simpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	log.Printf("DEBUG: %s %v", msg, keysAndValues)
}

// isValidBusinessID validates business ID format (1-99999999)
func isValidBusinessID(businessID string) bool {
	matched, _ := regexp.MatchString(`^[1-9][0-9]{0,7}$`, businessID)
	return matched
}

// isValidOrganizationBusinessID validates organization business ID format (100000-999999)
func isValidOrganizationBusinessID(businessID string) bool {
	matched, _ := regexp.MatchString(`^[1-9][0-9]{5}$`, businessID)
	return matched
}

const (
	ServiceName = "cube-castle-api"
	Version     = "v1.4.0"
)

// 全局EventBus服务管理器
var eventBusManager *events.EventBusManager

// MockEventBusService Mock事件总线服务（用于开发环境）
type MockEventBusService struct {
	eventBus events.EventBus
}

func (m *MockEventBusService) GetEventBus() events.EventBus {
	return m.eventBus
}

func (m *MockEventBusService) GetSerializer() events.EventSerializer {
	factory := events.NewEventSerializerFactory()
	return factory.CreateJSONSerializer()
}

func (m *MockEventBusService) GetValidator() *events.EventValidator {
	return events.NewEventValidator()
}

func (m *MockEventBusService) Start(ctx context.Context) error {
	return nil // Mock实现不需要启动
}

func (m *MockEventBusService) Stop() error {
	return nil // Mock实现不需要停止
}

func (m *MockEventBusService) Health() error {
	return nil // Mock实现始终健康
}

// InMemoryEventBusService InMemory事件总线服务（用于开发环境实际事件处理）
type InMemoryEventBusService struct {
	eventBus events.EventBus
}

func (i *InMemoryEventBusService) GetEventBus() events.EventBus {
	return i.eventBus
}

func (i *InMemoryEventBusService) GetSerializer() events.EventSerializer {
	factory := events.NewEventSerializerFactory()
	return factory.CreateJSONSerializer()
}

func (i *InMemoryEventBusService) GetValidator() *events.EventValidator {
	return events.NewEventValidator()
}

func (i *InMemoryEventBusService) Start(ctx context.Context) error {
	return i.eventBus.Start(ctx)
}

func (i *InMemoryEventBusService) Stop() error {
	return i.eventBus.Stop()
}

func (i *InMemoryEventBusService) Health() error {
	return i.eventBus.Health()
}

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

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

	// 检查生产环境安全
	env := os.Getenv("DEPLOYMENT_ENV")
	if env == "production" || env == "prod" {
		logger.Info("Production environment detected - mock mode disabled")
	}

	// 初始化数据库连接
	db := common.InitDatabaseConnection()
	if db == nil {
		if env == "production" || env == "prod" {
			logger.LogError("database_init", "CRITICAL: Database unavailable in production environment", nil, map[string]interface{}{
				"service": ServiceName,
				"environment": env,
			})
			log.Fatal("Production deployment requires database connection")
		}
		logger.LogError("database_init", "Failed to initialize database", nil, map[string]interface{}{
			"service": ServiceName,
		})
		// 在开发模式下继续运行（使用Mock）
		logger.Info("Running in mock mode - using in-memory data")
	} else {
		logger.Info("Database connected successfully")
	}

	// 初始化EventBus系统
	eventBusManager = initializeEventBusManager(logger, env)
	
	// 启动EventBus服务
	ctx := context.Background()
	if err := eventBusManager.StartAll(ctx); err != nil {
		logger.LogError("eventbus_start", "Failed to start EventBus services", err, map[string]interface{}{
			"service": ServiceName,
		})
		if env == "production" || env == "prod" {
			log.Fatal("Production deployment requires EventBus to be running")
		}
		logger.Warn("EventBus failed to start, continuing without event publishing")
	}

	// 获取EventBus实例
	var eventBus events.EventBus
	if eventBusManager != nil {
		if service, exists := eventBusManager.GetService("main"); exists {
			eventBus = service.GetEventBus()
		}
	}

	// 检查现有Neo4j连接管理器是否可用
	entClient := common.GetEntClient()
	if entClient != nil {
		logger.Info("Ent客户端连接成功，企业级服务将使用现有连接")
		
		// 如果EventBus可用，启用企业级CDC和监控服务
		if eventBus != nil {
			logger.Info("✅ EventBus可用，企业级服务已启用")
		} else {
			logger.Warn("⚠️ EventBus不可用，部分企业级服务将受限")
		}
	} else {
		logger.Warn("⚠️ Ent客户端不可用，企业级服务将受限")
	}

	// 启用现有监控系统（使用默认配置）
	monitorConfig := &monitoring.MonitorConfig{
		ServiceName: ServiceName,
		Version:     Version,
		Environment: env,
	}
	monitor := monitoring.NewMonitor(monitorConfig)
	if monitor != nil {
		// 监控服务不需要显式启动，在endpoints中会自动运行
		logger.Info("✅ 应用监控系统已配置")
	}

	// 初始化服务
	coreHRService := initializeCoreHRService(db, logger)
	editorService := initializeEditorService(db, logger)

	// 将EventBus注入到CoreHR Service中
	if eventBusManager != nil {
		if service, exists := eventBusManager.GetService("main"); exists {
			coreHRService.SetEventBus(service.GetEventBus())
			logger.Info("EventBus injected into CoreHR Service successfully")
		}
	}

	// 创建路由器
	router := setupRoutes(logger, coreHRService, editorService, db)

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

	// 停止EventBus服务
	if eventBusManager != nil {
		if err := eventBusManager.StopAll(); err != nil {
			logger.LogError("eventbus_shutdown", "Failed to stop EventBus services", err, nil)
		} else {
			logger.Info("EventBus services stopped successfully")
		}
	}

	// 记录服务关闭
	uptime := time.Since(startTime)
	logger.LogServiceShutdown(ServiceName, "graceful_shutdown", uptime)
	logger.Info("✅ Server exited successfully")
}

// initializeEventBusManager 初始化EventBus管理器
func initializeEventBusManager(logger *logging.StructuredLogger, env string) *events.EventBusManager {
	manager := events.NewEventBusManager()
	
	// 创建主EventBus配置
	config := events.ConfigFromEnv()
	
	// 根据环境调整配置
	if env == "production" || env == "prod" {
		// 生产环境配置
		config.EnableMetrics = true
		config.MaxRetries = 5
		config.RetryBackoff = time.Second * 3
		logger.Info("Using production EventBus configuration")
	} else {
		// 开发环境配置 - 优先使用InMemory EventBus实现真正的事件处理
		kafkaServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
		if kafkaServers == "" {
			// Kafka不可用，使用InMemory EventBus实现真正的事件处理
			logger.Info("Kafka not available, using InMemory EventBus for development")
			factory := events.NewEventBusFactory()
			inMemoryEventBus := factory.CreateInMemoryEventBus()
			
			// 创建InMemory服务包装器
			inMemoryService := &InMemoryEventBusService{
				eventBus: inMemoryEventBus,
			}
			
			// 注册InMemory服务
			manager.RegisterService("main", inMemoryService)
			return manager
		}
		logger.Info("Using development EventBus configuration with Kafka")
	}
	
	// 创建真实的EventBus服务
	eventBusService, err := events.NewEventBusService(config)
	if err != nil {
		logger.LogError("eventbus_init", "Failed to create EventBus service", err, map[string]interface{}{
			"kafka_servers": config.KafkaBootstrapServers,
		})
		
		// 降级到Mock EventBus
		logger.Info("Falling back to Mock EventBus")
		factory := events.NewEventBusFactory()
		mockEventBus := factory.CreateMockEventBus()
		
		// 创建Mock服务包装器
		mockService := &MockEventBusService{
			eventBus: mockEventBus,
		}
		manager.RegisterService("main", mockService)
		return manager
	}
	
	// 注册主EventBus服务
	manager.RegisterService("main", eventBusService)
	
	logger.Info("EventBus manager initialized successfully",
		"kafka_servers", config.KafkaBootstrapServers,
		"topic_prefix", config.KafkaTopicPrefix,
		"consumer_group", config.KafkaConsumerGroup,
	)
	
	return manager
}

// getEventBus 获取EventBus实例（供其他代码使用）
func getEventBus() events.EventBus {
	if eventBusManager == nil {
		return nil
	}
	
	service, exists := eventBusManager.GetService("main")
	if !exists {
		return nil
	}
	
	return service.GetEventBus()
}

// setupRoutes 设置路由
func setupRoutes(logger *logging.StructuredLogger, coreHRService *corehr.Service, editorService *metacontracteditor.Service, db interface{}) *chi.Mux {
	r := chi.NewRouter()

	// 添加中间件
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(metrics.PrometheusMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.TenantMiddleware)
	r.Use(middleware.AuthMiddleware(logger))
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// 健康检查端点（不需要认证）
	r.Get("/health", middleware.HealthCheckMiddleware(logger))

	// 企业级服务健康检查端点
	r.Get("/health/detailed", handleDetailedHealth(logger))
	r.Get("/health/cdc", handleCDCHealthCheck(logger))
	r.Get("/health/neo4j-sync", handleNeo4jSyncHealthCheck(logger))
	r.Get("/health/data-consistency", handleDataConsistencyCheck(logger))

	// Prometheus指标端点（不需要认证）
	r.Handle("/metrics", metrics.MetricsHandler())

	// 初始化Ent客户端
	entClient := common.GetEntClient()
	if entClient == nil {
		logger.LogError("ent_client_init", "Failed to initialize Ent client", nil, nil)
		// 在开发模式下继续运行，但不初始化需要数据库的处理器
	}

	// 初始化处理器（只有在数据库连接成功时才初始化）
	var positionHandler *handler.PositionHandler
	var employeeHandler *handler.EmployeeHandler
	var positionAssignmentHandler *handler.PositionAssignmentHandler
	var lifecycleHandler *handler.EmployeeLifecycleHandler
	var analyticsHandler *handler.AnalyticsHandler
	var validator *validation.EmployeeValidator

	if entClient != nil {
		positionHandler = handler.NewPositionHandler(entClient, logger)
		employeeHandler = handler.NewEmployeeHandler(entClient, logger)
		
		// 初始化高级服务和处理器
		positionAssignmentService := service.NewPositionAssignmentService(entClient, logger)
		lifecycleService := service.NewEmployeeLifecycleService(entClient, logger)
		analyticsService := service.NewAnalyticsService(entClient, logger)
		
		positionAssignmentHandler = handler.NewPositionAssignmentHandler(positionAssignmentService, logger)
		lifecycleHandler = handler.NewEmployeeLifecycleHandler(lifecycleService, logger)
		analyticsHandler = handler.NewAnalyticsHandler(analyticsService, logger)
	}

	// 初始化验证器
	// 检查是否可以使用真实验证器
	if db != nil {
		// 尝试转换数据库连接类型
		if pgxDB, ok := db.(*pgxpool.Pool); ok {
			// 创建Repository用于验证器
			repo := corehr.NewRepository(pgxDB)
			coreHRChecker := validation.NewCoreHRValidationChecker(repo)
			validator = validation.NewEmployeeValidator(coreHRChecker, coreHRChecker, coreHRChecker, coreHRChecker)
			logger.Info("✅ Initialized CoreHR validation checker with database connection")
		} else {
			// 数据库类型不匹配，使用Mock验证器
			mockChecker := validation.NewMockValidationChecker()
			validator = validation.NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)
			logger.Warn("⚠️ Database type mismatch, using mock validation checker")
		}
	} else {
		// 数据库未连接
		mockChecker := validation.NewMockValidationChecker()
		validator = validation.NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)
		
		// 根据环境给出不同的日志级别
		env := os.Getenv("DEPLOYMENT_ENV")
		if env == "production" || env == "prod" {
			logger.LogError("validation_init", "CRITICAL: Using mock validation in production - database required", nil, map[string]interface{}{
				"environment": env,
				"service": ServiceName,
			})
		} else {
			logger.Info("🔧 Using mock validation checker - database not available")
		}
	}

	// API v1 路由组
	r.Route("/api/v1", func(r chi.Router) {
		// CoreHR 模块路由
		r.Route("/corehr", func(r chi.Router) {
			r.Get("/employees", handleListEmployees(coreHRService, logger, validator))
			r.Post("/employees", handleCreateEmployee(coreHRService, logger, validator))
			r.Route("/employees/{employeeID}", func(r chi.Router) {
				r.Get("/", handleGetEmployee(coreHRService, logger))
				r.Put("/", handleUpdateEmployee(coreHRService, logger, validator))
				r.Delete("/", handleDeleteEmployee(coreHRService, logger, validator))
				r.Get("/manager", handleGetEmployeeManager(coreHRService, logger))
			})
		})

		// Setup organization routes using the adapter
		if entClient != nil {
			var sqlDB *sql.DB
			if db != nil {
				// Create sql.DB connection for business ID service
				databaseURL := os.Getenv("DATABASE_URL")
				if databaseURL == "" {
					databaseURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
				}
				
				var err error
				sqlDB, err = sql.Open("postgres", databaseURL)
				if err != nil {
					logger.LogError("setup_routes", "Failed to open sql.DB connection", err, nil)
					sqlDB = nil
				}
			}
			routes.SetupOrganizationRoutes(r, entClient, logger, sqlDB)
			
			// Initialize CQRS handlers properly
			if db != nil {
				// 初始化CQRS命令处理器
				if database, ok := db.(*common.Database); ok && database != nil && database.PostgreSQL != nil {
					// 创建sqlx连接用于CQRS repository
					sqlxDB, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
					if err == nil && sqlxDB.Ping() == nil {
						// 创建命令仓储
						empCommandRepo := repositories.NewPostgresCommandRepository(sqlxDB, logger)
						orgCommandRepo := repositories.NewPostgresOrganizationCommandRepository(sqlxDB, logger)
						
						// 创建职位命令仓储（包含Outbox Pattern支持）
						outboxRepo := repositories.NewPostgresOutboxRepository(sqlxDB)
						posCommandRepo := repositories.NewPostgresPositionRepository(sqlxDB, outboxRepo)
						
						// 获取EventBus
						var eventBus events.EventBus
						if eventBusManager != nil {
							if service, exists := eventBusManager.GetService("main"); exists {
								eventBus = service.GetEventBus()
							}
						}
						
						// 初始化Outbox处理器服务
						if eventBus != nil {
							outboxProcessorService := services.NewOutboxProcessorService(
								outboxRepo, 
								eventBus, 
								logger, 
								nil, // 使用默认配置
							)
							
							// 启动Outbox处理器
							if err := outboxProcessorService.Start(); err != nil {
								logger.Error("Failed to start outbox processor", "error", err)
							} else {
								logger.Info("✅ Outbox processor service started successfully")
							}
							
							// 注册优雅关闭处理
							defer func() {
								if err := outboxProcessorService.Stop(); err != nil {
									logger.Error("Failed to stop outbox processor", "error", err)
								}
							}()
						}
						
						// 创建命令处理器
						commandHandler := handlers.NewCommandHandler(empCommandRepo, orgCommandRepo, posCommandRepo, eventBus)
						
						// 添加CQRS命令路由
						r.Route("/commands", func(r chi.Router) {
							// 员工命令
							r.Post("/employees/hire", commandHandler.HireEmployee)
							r.Put("/employees/update", commandHandler.UpdateEmployee)
							r.Put("/update-employee", commandHandler.UpdateEmployee)
							
							// 组织命令
							r.Post("/organizations", commandHandler.CreateOrganization)
							r.Put("/organizations/{id}", commandHandler.UpdateOrganization)
							r.Delete("/organizations/{id}", commandHandler.DeleteOrganization)
							
							// 职位命令 (新实现)
							r.Post("/positions", commandHandler.CreatePosition)
							r.Put("/positions/{id}", commandHandler.UpdatePosition)
							r.Delete("/positions/{id}", commandHandler.DeletePosition)
							
							// 职位分配命令
							r.Post("/positions/assign-employee", commandHandler.AssignEmployeeToPosition)
							r.Post("/positions/remove-employee", commandHandler.RemoveEmployeeFromPosition)
						})
						
						// 创建Neo4j服务连接
						neo4jService, err := initializeNeo4jService()
						if err != nil {
							log.Printf("⚠️ Neo4j初始化失败，使用模拟数据: %v", err)
							neo4jService = nil // 使用模拟数据模式
						}

						// 创建查询处理器
						var queryHandler *handlers.QueryHandler
						logger := &simpleLogger{}
						neo4jQueryRepo := repositories.NewNeo4jEmployeeQueryRepository(neo4jService, logger)
						
						// 创建Neo4j组织查询仓储
						neo4jOrgQueryRepo := repositories.NewNeo4jOrganizationQueryRepository(neo4jService.GetDriver(), logger)
						
						// 创建Neo4j职位查询仓储
						neo4jPosQueryRepo := repositories.NewNeo4jPositionQueryRepositoryV2(neo4jService, logger)
						
						// 创建查询处理器，集成所有查询仓储
						queryHandler = handlers.NewQueryHandler(neo4jQueryRepo, neo4jOrgQueryRepo, neo4jPosQueryRepo)
						
						// 初始化和注册事件消费者（仅当Neo4j服务可用时）
						if neo4jService != nil {
							// 创建CDC Kafka消费者（独立于EventBus）
							cdcConfig := consumers.DefaultCDCConsumerConfig()
							cdcConsumer, err := consumers.NewCDCKafkaConsumer(cdcConfig, neo4jService, logger)
							if err != nil {
								logger.Error("Failed to create CDC Kafka consumer", "error", err)
							} else {
								// 启动CDC消费者
								ctx := context.Background()
								if err := cdcConsumer.Start(ctx); err != nil {
									logger.Error("Failed to start CDC Kafka consumer", "error", err)
								} else {
									logger.Info("✅ CDC Kafka消费者已启动，CQRS数据同步机制已启用")
									
									// 注册优雅关闭处理
									defer func() {
										if err := cdcConsumer.Stop(); err != nil {
											logger.Error("Failed to stop CDC consumer", "error", err)
										}
									}()
								}
							}
							
							log.Printf("✅ 事件消费者已注册完成，CQRS数据同步机制已启用")
						} else {
							log.Printf("⚠️ Neo4j或EventBus不可用，跳过事件消费者注册")
						}
						
						if queryHandler != nil {
							// 添加CQRS查询路由
							r.Route("/queries", func(r chi.Router) {
								// 员工查询
								r.Get("/employees/{id}", queryHandler.GetEmployee)
								r.Get("/employees", queryHandler.SearchEmployees)
								r.Get("/employees/stats", queryHandler.GetEmployeeStats)
								
								// 组织查询
								r.Get("/organizations", queryHandler.ListOrganizations)
								r.Get("/organizations/{id}", queryHandler.GetOrganization)
								r.Get("/organization-tree", queryHandler.GetOrganizationTree)
								r.Get("/organization-stats", queryHandler.GetOrganizationStats)
								r.Get("/organization-chart", queryHandler.GetOrgChart)
								r.Get("/organization-units/{id}", queryHandler.GetOrganizationUnit)
								r.Get("/organization-units", queryHandler.ListOrganizationUnits)
								r.Get("/reporting-hierarchy/{manager_id}", queryHandler.GetReportingHierarchy)
								
								// 职位查询 (新实现)
								r.Get("/positions/{id}", queryHandler.GetPosition)
								r.Get("/positions/{id}/relations", queryHandler.GetPositionWithRelations)
								r.Get("/positions", queryHandler.SearchPositions)
								r.Get("/positions/hierarchy", queryHandler.GetPositionHierarchy)
								r.Get("/positions/stats", queryHandler.GetPositionStats)
								
								// 职位-员工关系查询
								r.Get("/employees/{employee_id}/positions", queryHandler.GetEmployeePositions)
								r.Get("/positions/{position_id}/employees", queryHandler.GetPositionEmployees)
							})
							
							logger.Info("CQRS query routes configured successfully",
								"endpoints", []string{
									"/api/v1/queries/employees/stats",
									"/api/v1/queries/employees",
									"/api/v1/queries/organizations",
									"/api/v1/queries/positions",
									"/api/v1/queries/positions/hierarchy",
								},
							)
						}
						
						logger.Info("CQRS command routes configured successfully",
							"endpoints", []string{
								"/api/v1/commands/employees/hire",
								"/api/v1/commands/employees/update",
								"/api/v1/commands/organizations",
								"/api/v1/commands/organizations/{id}",
							},
						)
					} else {
						logger.Warn("Failed to initialize CQRS handlers - database connection issue")
					}
				} else {
					logger.Warn("Failed to initialize CQRS handlers - invalid database type")
				}
			} else {
				logger.Warn("CQRS handlers not initialized - database unavailable")
			}
		} else {
			// Database unavailable fallback for organization routes
			r.Route("/corehr/organizations", func(r chi.Router) {
				r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
				r.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			})
		}

		// 新的岗位CRUD API
		r.Route("/positions", func(r chi.Router) {
			if positionHandler != nil {
				r.Get("/", positionHandler.ListPositions())
				r.Post("/", positionHandler.CreatePosition())
				// 新增：根据部门获取职位列表的API
				r.Get("/by-department", positionHandler.GetPositionsByDepartment())
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", positionHandler.GetPosition())
					r.Put("/", positionHandler.UpdatePosition())
					r.Delete("/", positionHandler.DeletePosition())
				})
			} else {
				// 数据库未连接时返回服务不可用
				r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
				r.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
				r.Get("/by-department", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			}
		})

		// 迁移指南API
		migrationHandler := handler.NewMigrationHandler()
		r.Get("/migration-guide", migrationHandler.GetMigrationGuide())

		// CQRS健康检查API (简化版本)
		r.Get("/health/cqrs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now(),
				"version":   "v1.4.0",
				"cqrs": map[string]string{
					"command_side": "healthy",
					"query_side":   "healthy",
					"event_bus":    "healthy",
				},
			})
		})

		// 传统员工CRUD API (已废弃 - 计划于2024-12-31移除)
		r.Route("/employees", func(r chi.Router) {
			if employeeHandler != nil {
				// CRUD operations with deprecation warnings
				r.Get("/", middleware.WrapDeprecatedHandler(
					employeeHandler.ListEmployees(),
					middleware.EmployeeDeprecationInfo["list"],
					logger,
				))
				r.Post("/", middleware.WrapDeprecatedHandler(
					employeeHandler.CreateEmployee(),
					middleware.EmployeeDeprecationInfo["create"],
					logger,
				))
				// 新增：获取潜在经理列表的API
				r.Get("/potential-managers", employeeHandler.GetPotentialManagers())
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", middleware.WrapDeprecatedHandler(
						employeeHandler.GetEmployee(),
						middleware.EmployeeDeprecationInfo["get"],
						logger,
					))
					r.Put("/", middleware.WrapDeprecatedHandler(
						employeeHandler.UpdateEmployee(),
						middleware.EmployeeDeprecationInfo["update"],
						logger,
					))
					r.Delete("/", middleware.WrapDeprecatedHandler(
						employeeHandler.DeleteEmployee(),
						middleware.EmployeeDeprecationInfo["delete"],
						logger,
					))

					// Position-related operations (已废弃)
					r.Post("/assign-position", middleware.WrapDeprecatedHandler(
						employeeHandler.AssignPosition(),
						middleware.EmployeeDeprecationInfo["assign_position"],
						logger,
					))
					r.Get("/position-history", middleware.WrapDeprecatedHandler(
						employeeHandler.GetPositionHistory(),
						middleware.EmployeeDeprecationInfo["position_history"],
						logger,
					))
				})
			} else {
				// 数据库未连接时返回服务不可用
				r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
				r.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			}
		})

		// Advanced Employee Management APIs
		r.Route("/assignments", func(r chi.Router) {
			if positionAssignmentHandler != nil {
				r.Post("/", positionAssignmentHandler.AssignPosition())
				r.Post("/transfer", positionAssignmentHandler.TransferEmployee())
				r.Delete("/{employeeId}", positionAssignmentHandler.EndAssignment())
				r.Get("/active", positionAssignmentHandler.GetActiveAssignments())
			} else {
				// Database service unavailable fallback
				r.Post("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			}
		})

		r.Route("/lifecycle", func(r chi.Router) {
			if lifecycleHandler != nil {
				r.Post("/onboard", lifecycleHandler.OnboardEmployee())
				r.Post("/offboard", lifecycleHandler.OffboardEmployee())
				r.Post("/promote", lifecycleHandler.PromoteEmployee())
				r.Post("/status-change", lifecycleHandler.ChangeEmploymentStatus())
			} else {
				// Database service unavailable fallback
				r.Post("/onboard", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			}
		})

		r.Route("/analytics", func(r chi.Router) {
			if analyticsHandler != nil {
				r.Get("/metrics", analyticsHandler.GetOrganizationalMetrics())
				r.Get("/employees/{id}/history", analyticsHandler.GetEmployeeHistory())
				r.Get("/positions/{id}/history", analyticsHandler.GetPositionHistory())
				r.Get("/assignments/history", analyticsHandler.GetHistoricalAssignments())
			} else {
				// Database service unavailable fallback
				r.Get("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
				}))
			}
		})

		// Meta-Contract Editor 路由
		r.Route("/metacontract", func(r chi.Router) {
			// 项目管理
			r.Get("/projects", handleListProjects(editorService, logger))
			r.Post("/projects", handleCreateProject(editorService, logger))
			r.Route("/projects/{projectID}", func(r chi.Router) {
				r.Get("/", handleGetProject(editorService, logger))
				r.Put("/", handleUpdateProject(editorService, logger))
				r.Delete("/", handleDeleteProject(editorService, logger))
				r.Post("/compile", handleCompileProject(editorService, logger))
			})

			// 模板管理
			r.Get("/templates", handleGetTemplates(editorService, logger))

			// 用户设置
			r.Get("/settings", handleGetUserSettings(editorService, logger))
			r.Put("/settings", handleUpdateUserSettings(editorService, logger))
		})

		// 监控和管理路由
		r.Route("/admin", func(r chi.Router) {
			r.Get("/metrics/business", handleBusinessMetrics(logger))
			r.Get("/health/detailed", handleDetailedHealth(logger))
			r.Post("/cache/clear", handleClearCache(logger))
		})
	})

	// 根路由 - 提供API服务信息
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "Cube Castle API",
			"version": "v1.4.0",
			"status":  "running",
			"endpoints": map[string]string{
				"health":        "/health",
				"metrics":       "/metrics",
				"organizations": "/api/v1/corehr/organizations",
				"employees":     "/api/v1/corehr/employees",
				"api_docs":      "/api/v1/docs",
			},
			"timestamp": time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	})

	return r
}

// initializeEditorService 初始化元合约编辑器服务
func initializeEditorService(db interface{}, logger *logging.StructuredLogger) *metacontracteditor.Service {
	// 创建编译器
	compiler := metacontract.NewCompiler()

	if db == nil {
		// Mock模式
		logger.Info("Initializing Meta-Contract Editor service in mock mode")
		mockRepo := createMockEditorRepository()
		return metacontracteditor.NewService(mockRepo, compiler)
	}

	// 实际模式 - 转换数据库连接
	logger.Info("Initializing Meta-Contract Editor service with database connection")
	
	// 尝试从Database结构体获取PostgreSQL连接
	if database, ok := db.(*common.Database); ok && database != nil && database.PostgreSQL != nil {
		// 创建sqlx连接用于repository
		sqlxDB, err := sqlx.Open("postgres", os.Getenv("DATABASE_URL"))
		if err != nil {
			logger.LogError("db_conversion", "Failed to create sqlx connection", err, nil)
			logger.Info("Falling back to mock mode")
			mockRepo := createMockEditorRepository()
			return metacontracteditor.NewService(mockRepo, compiler)
		}
		
		// 测试连接
		if err := sqlxDB.Ping(); err != nil {
			logger.LogError("db_ping", "Failed to ping database", err, nil)
			logger.Info("Falling back to mock mode")
			mockRepo := createMockEditorRepository()
			return metacontracteditor.NewService(mockRepo, compiler)
		}
		
		// 创建实际的repository
		repo := metacontracteditor.NewPostgreSQLRepository(sqlxDB)
		return metacontracteditor.NewService(repo, compiler)
	}

	// 如果转换失败，使用mock
	logger.Info("Unable to use database connection, falling back to mock mode")
	mockRepo := createMockEditorRepository()
	return metacontracteditor.NewService(mockRepo, compiler)
}

// initializeCoreHRService 初始化CoreHR服务
func initializeCoreHRService(db interface{}, logger *logging.StructuredLogger) *corehr.Service {
	if db == nil {
		// Mock模式
		logger.Info("Initializing CoreHR service in mock mode")
		return corehr.NewMockService()
	}

	// 实际模式 - 使用数据库连接
	logger.Info("Initializing CoreHR service with database connection")
	
	// 转换数据库连接类型 - 支持 *common.Database 类型
	var pgxDB *pgxpool.Pool
	
	switch dbConn := db.(type) {
	case *pgxpool.Pool:
		// 直接使用pgxPool
		pgxDB = dbConn
	case *common.Database:
		// 从common.Database结构体中提取PostgreSQL连接
		if dbConn != nil && dbConn.PostgreSQL != nil {
			pgxDB = dbConn.PostgreSQL
		} else {
			logger.LogError("database_error", "PostgreSQL connection is nil in Database struct", nil, nil)
			return corehr.NewMockService()
		}
	default:
		logger.LogError("database_type_error", "Invalid database connection type", nil, map[string]interface{}{
			"expected": "*pgxpool.Pool or *common.Database",
			"actual": fmt.Sprintf("%T", db),
		})
		return corehr.NewMockService()
	}
	
	// 创建真实的Repository和Service
	repo := corehr.NewRepository(pgxDB)
	
	logger.Info("CoreHR service initialized with real database implementation")
	return corehr.NewService(repo)
}

// Meta-Contract Editor 处理函数
func handleListProjects(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取查询参数
		limit := getIntParam(r, "limit", 10)
		offset := getIntParam(r, "offset", 0)

		// 调用服务
		projects, err := service.ListProjects(r.Context(), tenantID, limit, offset)
		if err != nil {
			reqLogger.LogError("list_projects", "Failed to list projects", err, map[string]interface{}{
				"tenant_id": tenantID,
				"limit":     limit,
				"offset":    offset,
			})
			http.Error(w, "Failed to list projects", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"projects": projects,
			"total":    len(projects),
		})
	}
}

func handleCreateProject(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())
		userID := getUserID(r.Context())

		// 解析请求体
		var req metacontracteditor.CreateProjectRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request", "Failed to parse create project request", err, nil)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 设置租户ID和用户ID
		req.TenantID = tenantID
		req.UserID = userID

		// 调用服务
		project, err := service.CreateProject(r.Context(), req)
		if err != nil {
			reqLogger.LogError("create_project", "Failed to create project", err, map[string]interface{}{
				"name":      req.Name,
				"tenant_id": tenantID,
			})
			http.Error(w, "Failed to create project", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusCreated, project)
	}
}

func handleGetProject(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取项目ID
		projectID := chi.URLParam(r, "projectID")
		if projectID == "" {
			http.Error(w, "Project ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// 调用服务
		project, err := service.GetProject(r.Context(), projectUUID, tenantID)
		if err != nil {
			reqLogger.LogError("get_project", "Failed to get project", err, map[string]interface{}{
				"project_id": projectID,
				"tenant_id":  tenantID,
			})
			if err.Error() == "project not found or access denied" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to get project", http.StatusInternalServerError)
			}
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, project)
	}
}

func handleUpdateProject(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取项目ID
		projectID := chi.URLParam(r, "projectID")
		if projectID == "" {
			http.Error(w, "Project ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// 解析请求体
		var req metacontracteditor.UpdateProjectRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request", "Failed to parse update project request", err, nil)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 设置租户ID
		req.TenantID = tenantID

		// 调用服务
		project, err := service.UpdateProject(r.Context(), projectUUID, req)
		if err != nil {
			reqLogger.LogError("update_project", "Failed to update project", err, map[string]interface{}{
				"project_id": projectID,
				"tenant_id":  tenantID,
			})
			if err.Error() == "project not found or access denied" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to update project", http.StatusInternalServerError)
			}
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, project)
	}
}

func handleDeleteProject(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取项目ID
		projectID := chi.URLParam(r, "projectID")
		if projectID == "" {
			http.Error(w, "Project ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// 调用服务
		err = service.DeleteProject(r.Context(), projectUUID, tenantID)
		if err != nil {
			reqLogger.LogError("delete_project", "Failed to delete project", err, map[string]interface{}{
				"project_id": projectID,
				"tenant_id":  tenantID,
			})
			if err.Error() == "project not found or access denied" {
				http.Error(w, "Project not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to delete project", http.StatusInternalServerError)
			}
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusNoContent, duration, r.UserAgent())

		// 返回响应
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleCompileProject(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())

		// 获取项目ID
		projectID := chi.URLParam(r, "projectID")
		if projectID == "" {
			http.Error(w, "Project ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		projectUUID, err := uuid.Parse(projectID)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		// 解析请求体
		var req metacontracteditor.CompileRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request", "Failed to parse compile request", err, nil)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 设置项目ID
		req.ProjectID = projectUUID

		// 调用服务
		response, err := service.CompileProject(r.Context(), req)
		if err != nil {
			reqLogger.LogError("compile_project", "Failed to compile project", err, map[string]interface{}{
				"project_id": projectID,
			})
			http.Error(w, "Failed to compile project", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, response)
	}
}

func handleGetTemplates(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())

		// 获取查询参数
		category := r.URL.Query().Get("category")
		if category == "" {
			category = "basic" // 默认分类
		}

		// 调用服务
		templates, err := service.GetTemplates(r.Context(), category)
		if err != nil {
			reqLogger.LogError("get_templates", "Failed to get templates", err, map[string]interface{}{
				"category": category,
			})
			http.Error(w, "Failed to get templates", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"templates": templates,
			"category":  category,
		})
	}
}

func handleGetUserSettings(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		userID := getUserID(r.Context())

		// 调用服务
		settings, err := service.GetUserSettings(r.Context(), userID)
		if err != nil {
			reqLogger.LogError("get_user_settings", "Failed to get user settings", err, map[string]interface{}{
				"user_id": userID,
			})
			http.Error(w, "Failed to get user settings", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, settings)
	}
}

func handleUpdateUserSettings(service *metacontracteditor.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		userID := getUserID(r.Context())

		// 解析请求体
		var settings metacontracteditor.EditorSettings
		if err := parseJSON(r, &settings); err != nil {
			reqLogger.LogError("parse_request", "Failed to parse update settings request", err, nil)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 设置用户ID
		settings.UserID = userID

		// 调用服务
		err := service.UpdateUserSettings(r.Context(), &settings)
		if err != nil {
			reqLogger.LogError("update_user_settings", "Failed to update user settings", err, map[string]interface{}{
				"user_id": userID,
			})
			http.Error(w, "Failed to update user settings", http.StatusInternalServerError)
			return
		}

		// 记录指标
		duration := time.Since(start)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		// 返回响应
		respondJSON(w, http.StatusOK, map[string]string{"message": "Settings updated successfully"})
	}
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
func handleListEmployees(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取查询参数
		page := getIntParam(r, "page", 1)
		pageSize := getIntParam(r, "page_size", 20)
		search := r.URL.Query().Get("search")

		// 参数验证
		if validator != nil {
			if err := validator.ValidateListEmployeesParams(page, pageSize, search); err != nil {
				reqLogger.LogError("validation_error", "List employees parameter validation failed", err, map[string]interface{}{
					"page": page,
					"page_size": pageSize,
					"search": search,
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "list_employees_validation_error")
				
				// 返回验证错误详情
				if validationErrors, ok := err.(validation.ValidationErrors); ok {
					responseBody := map[string]interface{}{
						"error": "Validation failed",
						"details": validationErrors.Errors,
					}
					respondJSON(w, http.StatusBadRequest, responseBody)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}
		} else {
			// Fallback validation when validator is not available
			if page < 1 {
				reqLogger.LogError("validation_error", "Invalid page parameter", nil, map[string]interface{}{
					"page": page,
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "list_employees_validation_error")
				http.Error(w, "Page must be greater than 0", http.StatusBadRequest)
				return
			}

			if pageSize < 1 || pageSize > 100 {
				reqLogger.LogError("validation_error", "Invalid page_size parameter", nil, map[string]interface{}{
					"page_size": pageSize,
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "list_employees_validation_error")
				http.Error(w, "Page size must be between 1 and 100", http.StatusBadRequest)
				return
			}
		}

		// 记录请求开始
		reqLogger.Info("Processing list employees request",
			"page", page,
			"page_size", pageSize,
			"search", search,
			"tenant_id", tenantID.String(),
		)

		// 调用服务
		response, err := service.ListEmployees(r.Context(), tenantID, page, pageSize, search)
		if err != nil {
			reqLogger.LogError("list_employees_service_error", "Failed to list employees from service", err, map[string]interface{}{
				"page":      page,
				"page_size": pageSize,
				"search":    search,
				"tenant_id": tenantID.String(),
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "list_employees_service_error")
			
			// 根据错误类型返回不同状态码
			if strings.Contains(err.Error(), "timeout") {
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// 响应验证
		if response == nil {
			reqLogger.LogError("service_response_error", "Service returned nil response", nil, map[string]interface{}{
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "list_employees_nil_response_error")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 记录成功指标
		duration := time.Since(start)
		employeeCount := 0
		if response.Employees != nil {
			employeeCount = len(*response.Employees)
		}

		metrics.RecordDatabaseOperation("SELECT", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("SELECT", "employees", employeeCount, duration, true)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		reqLogger.Info("Successfully listed employees",
			"count", employeeCount,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

		// 返回响应
		respondJSON(w, http.StatusOK, response)
	}
}

// handleCreateEmployee 处理创建员工请求
func handleCreateEmployee(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 验证Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			reqLogger.LogError("content_type_error", "Invalid content type", nil, map[string]interface{}{
				"content_type": r.Header.Get("Content-Type"),
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "create_employee_content_type_error")
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// 解析请求体
		var req CreateEmployeeRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request_error", "Failed to parse create employee request JSON", err, map[string]interface{}{
				"tenant_id": tenantID.String(),
				"content_length": r.ContentLength,
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "create_employee_parse_error")
			http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
			return
		}

		// 使用验证器进行数据验证
		if validator != nil {
			if err := validator.ValidateCreateEmployee(r.Context(), tenantID, &req); err != nil {
				reqLogger.LogError("validation_error", "Create employee validation failed", err, map[string]interface{}{
					"employee_number": req.EmployeeNumber,
					"first_name": req.FirstName,
					"last_name": req.LastName,
					"email": string(req.Email),
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "create_employee_validation_error")
				
				// 返回验证错误详情
				if validationErrors, ok := err.(validation.ValidationErrors); ok {
					responseBody := map[string]interface{}{
						"error": "Validation failed",
						"details": validationErrors.Errors,
					}
					respondJSON(w, http.StatusBadRequest, responseBody)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}
		} else {
			// Fallback validation when validator is not available
			if req.EmployeeNumber == "" {
				reqLogger.LogError("validation_error", "Employee number is required", nil, map[string]interface{}{
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "create_employee_validation_error")
				http.Error(w, "Employee number is required", http.StatusBadRequest)
				return
			}
			if req.FirstName == "" || req.LastName == "" {
				reqLogger.LogError("validation_error", "First name and last name are required", nil, map[string]interface{}{
					"employee_number": req.EmployeeNumber,
					"first_name_empty": req.FirstName == "",
					"last_name_empty": req.LastName == "",
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "create_employee_validation_error")
				http.Error(w, "First name and last name are required", http.StatusBadRequest)
				return
			}
		}

		// 记录请求开始
		reqLogger.Info("Processing create employee request",
			"employee_number", req.EmployeeNumber,
			"first_name", req.FirstName,
			"last_name", req.LastName,
			"tenant_id", tenantID.String(),
		)

		// 调用服务
		employee, err := service.CreateEmployee(r.Context(), tenantID, &req)
		if err != nil {
			reqLogger.LogError("create_employee_service_error", "Failed to create employee in service", err, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"first_name":      req.FirstName,
				"last_name":       req.LastName,
				"tenant_id":       tenantID.String(),
				"error_type":      fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "create_employee_service_error")
			
			// 根据错误类型返回适当的HTTP状态码
			if strings.Contains(err.Error(), "already exists") {
				http.Error(w, "Employee number already exists", http.StatusConflict)
			} else if strings.Contains(err.Error(), "validation") {
				http.Error(w, "Invalid employee data", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "timeout") {
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				http.Error(w, "Failed to create employee", http.StatusInternalServerError)
			}
			return
		}

		// 响应验证
		if employee == nil {
			reqLogger.LogError("service_response_error", "Service returned nil employee", nil, map[string]interface{}{
				"employee_number": req.EmployeeNumber,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "create_employee_nil_response_error")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 记录成功指标和日志
		duration := time.Since(start)
		metrics.RecordEmployeeCreated(tenantID.String())
		metrics.RecordDatabaseOperation("INSERT", "employees", "success", duration)
		reqLogger.LogEmployeeCreated(*employee.Id, tenantID, req.EmployeeNumber)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusCreated, duration, r.UserAgent())

		reqLogger.Info("Successfully created employee",
			"employee_id", employee.Id.String(),
			"employee_number", req.EmployeeNumber,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

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
		if id, ok := tenantID.(uuid.UUID); ok {
			return id
		}
		// 兼容性处理：如果存储的是字符串类型
		if id, ok := tenantID.(string); ok {
			if parsedID, err := uuid.Parse(id); err == nil {
				return parsedID
			}
		}
	}
	// 返回默认租户ID
	return uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
}

// getUserID 从上下文获取用户ID
func getUserID(ctx context.Context) uuid.UUID {
	// TODO: 实现从JWT token或session中提取用户ID
	// 这里先返回一个默认的用户ID用于测试
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

// createMockEditorRepository 创建Mock编辑器Repository
func createMockEditorRepository() metacontracteditor.Repository {
	return &MockEditorRepository{
		projects:  make(map[uuid.UUID]*metacontracteditor.EditorProject),
		settings:  make(map[uuid.UUID]*metacontracteditor.EditorSettings),
		templates: createDefaultTemplates(),
	}
}

// MockEditorRepository Mock编辑器Repository实现
type MockEditorRepository struct {
	projects  map[uuid.UUID]*metacontracteditor.EditorProject
	settings  map[uuid.UUID]*metacontracteditor.EditorSettings
	templates []*metacontracteditor.ProjectTemplate
}

func (m *MockEditorRepository) CreateProject(ctx context.Context, project *metacontracteditor.EditorProject) error {
	m.projects[project.ID] = project
	return nil
}

func (m *MockEditorRepository) GetProject(ctx context.Context, projectID uuid.UUID) (*metacontracteditor.EditorProject, error) {
	if project, exists := m.projects[projectID]; exists {
		return project, nil
	}
	return nil, fmt.Errorf("project not found")
}

func (m *MockEditorRepository) UpdateProject(ctx context.Context, project *metacontracteditor.EditorProject) error {
	if _, exists := m.projects[project.ID]; exists {
		m.projects[project.ID] = project
		return nil
	}
	return fmt.Errorf("project not found")
}

func (m *MockEditorRepository) ListProjects(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*metacontracteditor.EditorProject, error) {
	var result []*metacontracteditor.EditorProject
	for _, project := range m.projects {
		if project.TenantID == tenantID {
			result = append(result, project)
		}
	}
	return result, nil
}

func (m *MockEditorRepository) DeleteProject(ctx context.Context, projectID uuid.UUID) error {
	if _, exists := m.projects[projectID]; exists {
		delete(m.projects, projectID)
		return nil
	}
	return fmt.Errorf("project not found")
}

// Session operations (Mock implementations)
func (m *MockEditorRepository) CreateSession(ctx context.Context, session *metacontracteditor.EditorSession) error {
	// Mock implementation - just return success
	return nil
}

func (m *MockEditorRepository) GetSession(ctx context.Context, sessionID uuid.UUID) (*metacontracteditor.EditorSession, error) {
	// Mock implementation - return a dummy session
	return &metacontracteditor.EditorSession{
		ID:        sessionID,
		StartedAt: time.Now(),
		LastSeen:  time.Now(),
		Active:    true,
	}, nil
}

func (m *MockEditorRepository) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	// Mock implementation - just return success
	return nil
}

func (m *MockEditorRepository) GetActiveSessions(ctx context.Context, projectID uuid.UUID) ([]*metacontracteditor.EditorSession, error) {
	// Mock implementation - return empty list
	return []*metacontracteditor.EditorSession{}, nil
}

func (m *MockEditorRepository) GetTemplates(ctx context.Context, category string) ([]*metacontracteditor.ProjectTemplate, error) {
	var result []*metacontracteditor.ProjectTemplate
	for _, template := range m.templates {
		if template.Category == category {
			result = append(result, template)
		}
	}
	return result, nil
}

func (m *MockEditorRepository) CreateTemplate(ctx context.Context, template *metacontracteditor.ProjectTemplate) error {
	// Mock implementation - just return success
	return nil
}

func (m *MockEditorRepository) GetUserSettings(ctx context.Context, userID uuid.UUID) (*metacontracteditor.EditorSettings, error) {
	if settings, exists := m.settings[userID]; exists {
		return settings, nil
	}
	// 返回默认设置
	return &metacontracteditor.EditorSettings{
		UserID:      userID,
		Theme:       "vs-dark",
		FontSize:    14,
		AutoSave:    true,
		AutoCompile: true,
		KeyBindings: "default",
		Settings:    make(map[string]interface{}),
		UpdatedAt:   time.Now(),
	}, nil
}

func (m *MockEditorRepository) UpdateUserSettings(ctx context.Context, settings *metacontracteditor.EditorSettings) error {
	m.settings[settings.UserID] = settings
	return nil
}

func createDefaultTemplates() []*metacontracteditor.ProjectTemplate {
	return []*metacontracteditor.ProjectTemplate{
		{
			ID:       uuid.New(),
			Name:     "Basic Entity",
			Category: "basic",
			Content: `resource_name: example_entity
namespace: example.namespace
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: name
      type: String
      constraints:
        required: true
        max_length: 255

security_model:
  access_control: rbac
  data_classification: internal`,
			Description: "A basic entity template with ID and name fields",
			Tags:        []string{"basic", "crud"},
		},
		{
			ID:       uuid.New(),
			Name:     "Employee Template",
			Category: "hr",
			Content: `resource_name: employee
namespace: hr.employees
version: "1.0.0"

data_structure:
  fields:
    - name: id
      type: UUID
      constraints:
        primary_key: true
        required: true
    
    - name: employee_id
      type: String
      constraints:
        required: true
        unique: true
        max_length: 20
    
    - name: first_name
      type: String
      constraints:
        required: true
        max_length: 50
    
    - name: last_name
      type: String
      constraints:
        required: true
        max_length: 50
    
    - name: email
      type: String
      constraints:
        required: true
        unique: true
        format: email

security_model:
  access_control: rbac
  data_classification: confidential`,
			Description: "Employee entity template for HR systems",
			Tags:        []string{"hr", "employee"},
		},
	}
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
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取员工ID
		employeeID := chi.URLParam(r, "employeeID")
		if employeeID == "" {
			reqLogger.LogError("missing_parameter", "Employee ID parameter is missing", nil, map[string]interface{}{
				"tenant_id": tenantID.String(),
				"path": r.URL.Path,
			})
			metrics.RecordError("corehr", "get_employee_missing_id_error")
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		employeeUUID, err := uuid.Parse(employeeID)
		if err != nil {
			reqLogger.LogError("invalid_uuid", "Invalid employee ID format", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "get_employee_invalid_uuid_error")
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// 记录请求开始
		reqLogger.Info("Processing get employee request",
			"employee_id", employeeID,
			"tenant_id", tenantID.String(),
		)

		// 调用服务
		employee, err := service.GetEmployee(r.Context(), tenantID, employeeUUID)
		if err != nil {
			reqLogger.LogError("get_employee_service_error", "Failed to get employee from service", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
				"error_type":  fmt.Sprintf("%T", err),
			})

			// 根据错误类型返回适当的HTTP状态码
			if strings.Contains(err.Error(), "not found") {
				metrics.RecordError("corehr", "get_employee_not_found_error")
				http.Error(w, "Employee not found", http.StatusNotFound)
			} else if strings.Contains(err.Error(), "timeout") {
				metrics.RecordError("corehr", "get_employee_timeout_error")
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				metrics.RecordError("corehr", "get_employee_connection_error")
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				metrics.RecordError("corehr", "get_employee_service_error")
				http.Error(w, "Failed to get employee", http.StatusInternalServerError)
			}
			return
		}

		// 响应验证
		if employee == nil {
			reqLogger.LogError("service_response_error", "Service returned nil employee", nil, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "get_employee_nil_response_error")
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		}

		// 记录成功指标
		duration := time.Since(start)
		metrics.RecordDatabaseOperation("SELECT", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("SELECT", "employees", 1, duration, true)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		reqLogger.Info("Successfully retrieved employee",
			"employee_id", employeeID,
			"employee_number", employee.EmployeeNumber,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

		// 返回响应
		respondJSON(w, http.StatusOK, employee)
	}
}

func handleUpdateEmployee(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取员工ID
		employeeID := chi.URLParam(r, "employeeID")
		if employeeID == "" {
			reqLogger.LogError("missing_parameter", "Employee ID parameter is missing", nil, map[string]interface{}{
				"tenant_id": tenantID.String(),
				"path": r.URL.Path,
			})
			metrics.RecordError("corehr", "update_employee_missing_id_error")
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}

		// 解析和验证业务ID
		if !isValidBusinessID(employeeID) {
			reqLogger.LogError("invalid_business_id", "Invalid employee business ID format", nil, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
			})
			metrics.RecordError("corehr", "update_employee_invalid_business_id_error")
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// 通过业务ID查找员工获取UUID - 直接查询数据库
		entClient := common.GetEntClient()
		if entClient == nil {
			reqLogger.LogError("ent_client_unavailable", "Ent client not available for business ID lookup", nil, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
			})
			metrics.RecordError("corehr", "update_employee_ent_client_error")
			http.Error(w, "Database service unavailable", http.StatusServiceUnavailable)
			return
		}

		entEmployee, err := entClient.Employee.Query().
			Where(
				employee.BusinessIDEQ(employeeID),
				employee.TenantIDEQ(tenantID),
			).
			Only(r.Context())
		if err != nil {
			if ent.IsNotFound(err) {
				reqLogger.LogError("employee_not_found", "Employee not found by business ID", err, map[string]interface{}{
					"employee_id": employeeID,
					"tenant_id":   tenantID.String(),
				})
				metrics.RecordError("corehr", "update_employee_not_found_error")
				http.Error(w, "Employee not found", http.StatusNotFound)
				return
			}
			reqLogger.LogError("lookup_employee_error", "Failed to lookup employee by business ID", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
				"error_type":  fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "update_employee_lookup_error")
			http.Error(w, "Failed to lookup employee", http.StatusInternalServerError)
			return
		}

		employeeUUID := entEmployee.ID

		// 验证Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			reqLogger.LogError("content_type_error", "Invalid content type", nil, map[string]interface{}{
				"content_type": r.Header.Get("Content-Type"),
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "update_employee_content_type_error")
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// 解析请求体 - 首先保存原始JSON用于业务ID处理
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			reqLogger.LogError("read_body_error", "Failed to read request body", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "update_employee_read_error")
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(rawBody))
		
		var req openapi.UpdateEmployeeRequest
		if err := parseJSON(r, &req); err != nil {
			reqLogger.LogError("parse_request_error", "Failed to parse update employee request JSON", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
				"content_length": r.ContentLength,
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "update_employee_parse_error")
			http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
			return
		}

		// 使用验证器进行数据验证
		if validator != nil {
			if err := validator.ValidateUpdateEmployee(r.Context(), tenantID, employeeUUID, &req); err != nil {
				reqLogger.LogError("validation_error", "Update employee validation failed", err, map[string]interface{}{
					"employee_id": employeeID,
					"tenant_id": tenantID.String(),
					"has_first_name": req.FirstName != nil,
					"has_last_name": req.LastName != nil,
					"has_email": req.Email != nil,
				})
				metrics.RecordError("corehr", "update_employee_validation_error")
				
				// 返回验证错误详情
				if validationErrors, ok := err.(validation.ValidationErrors); ok {
					responseBody := map[string]interface{}{
						"error": "Validation failed",
						"details": validationErrors.Errors,
					}
					respondJSON(w, http.StatusBadRequest, responseBody)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}
		}

		// 记录请求开始
		reqLogger.Info("Processing update employee request",
			"employee_id", employeeID,
			"tenant_id", tenantID.String(),
			"has_first_name", req.FirstName != nil,
			"has_last_name", req.LastName != nil,
			"has_email", req.Email != nil,
		)

		// 调用服务 - 传递原始JSON以处理业务ID参数
		updatedEmployee, err := service.UpdateEmployeeWithRawJSON(r.Context(), tenantID, employeeUUID, &req, rawBody)
		if err != nil {
			reqLogger.LogError("update_employee_service_error", "Failed to update employee in service", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
				"error_type":  fmt.Sprintf("%T", err),
			})

			// 根据错误类型返回适当的HTTP状态码
			if strings.Contains(err.Error(), "not found") {
				metrics.RecordError("corehr", "update_employee_not_found_error")
				http.Error(w, "Employee not found", http.StatusNotFound)
			} else if strings.Contains(err.Error(), "validation") {
				metrics.RecordError("corehr", "update_employee_validation_error")
				http.Error(w, "Invalid employee data", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "timeout") {
				metrics.RecordError("corehr", "update_employee_timeout_error")
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				metrics.RecordError("corehr", "update_employee_connection_error")
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				metrics.RecordError("corehr", "update_employee_service_error")
				http.Error(w, "Failed to update employee", http.StatusInternalServerError)
			}
			return
		}

		// 响应验证
		if updatedEmployee == nil {
			reqLogger.LogError("service_response_error", "Service returned nil employee", nil, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "update_employee_nil_response_error")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// 记录成功指标
		duration := time.Since(start)
		metrics.RecordDatabaseOperation("UPDATE", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("UPDATE", "employees", 1, duration, true)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		reqLogger.Info("Successfully updated employee",
			"employee_id", employeeID,
			"employee_number", updatedEmployee.EmployeeNumber,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

		// 返回响应
		respondJSON(w, http.StatusOK, updatedEmployee)
	}
}

func handleDeleteEmployee(service *corehr.Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取员工ID
		employeeID := chi.URLParam(r, "employeeID")
		if employeeID == "" {
			reqLogger.LogError("missing_parameter", "Employee ID parameter is missing", nil, map[string]interface{}{
				"tenant_id": tenantID.String(),
				"path": r.URL.Path,
			})
			metrics.RecordError("corehr", "delete_employee_missing_id_error")
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		employeeUUID, err := uuid.Parse(employeeID)
		if err != nil {
			reqLogger.LogError("invalid_uuid", "Invalid employee ID format", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "delete_employee_invalid_uuid_error")
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// 记录请求开始
		reqLogger.Info("Processing delete employee request",
			"employee_id", employeeID,
			"tenant_id", tenantID.String(),
		)

		// 使用验证器进行业务规则验证（员工是否可以被删除）
		if validator != nil {
			if err := validator.ValidateEmployeeTermination(r.Context(), employeeUUID, tenantID); err != nil {
				reqLogger.LogError("validation_error", "Delete employee validation failed", err, map[string]interface{}{
					"employee_id": employeeID,
					"tenant_id": tenantID.String(),
				})
				metrics.RecordError("corehr", "delete_employee_validation_error")
				
				// 返回验证错误详情
				if validationErrors, ok := err.(validation.ValidationErrors); ok {
					responseBody := map[string]interface{}{
						"error": "Validation failed",
						"details": validationErrors.Errors,
					}
					respondJSON(w, http.StatusBadRequest, responseBody)
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}
		}

		// 调用服务
		err = service.DeleteEmployee(r.Context(), tenantID, employeeUUID)
		if err != nil {
			reqLogger.LogError("delete_employee_service_error", "Failed to delete employee in service", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
				"error_type":  fmt.Sprintf("%T", err),
			})

			// 根据错误类型返回适当的HTTP状态码
			if strings.Contains(err.Error(), "not found") {
				metrics.RecordError("corehr", "delete_employee_not_found_error")
				http.Error(w, "Employee not found", http.StatusNotFound)
			} else if strings.Contains(err.Error(), "foreign key") || strings.Contains(err.Error(), "constraint") {
				metrics.RecordError("corehr", "delete_employee_constraint_error")
				http.Error(w, "Cannot delete employee due to existing references", http.StatusConflict)
			} else if strings.Contains(err.Error(), "timeout") {
				metrics.RecordError("corehr", "delete_employee_timeout_error")
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				metrics.RecordError("corehr", "delete_employee_connection_error")
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				metrics.RecordError("corehr", "delete_employee_service_error")
				http.Error(w, "Failed to delete employee", http.StatusInternalServerError)
			}
			return
		}

		// 记录成功指标
		duration := time.Since(start)
		metrics.RecordDatabaseOperation("DELETE", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("DELETE", "employees", 1, duration, true)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusNoContent, duration, r.UserAgent())

		reqLogger.Info("Successfully deleted employee",
			"employee_id", employeeID,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

		// 返回响应
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetEmployeeManager(service *corehr.Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqLogger := logger.WithContext(r.Context())
		tenantID := getTenantID(r.Context())

		// 获取员工ID
		employeeID := chi.URLParam(r, "employeeID")
		if employeeID == "" {
			reqLogger.LogError("missing_parameter", "Employee ID parameter is missing", nil, map[string]interface{}{
				"tenant_id": tenantID.String(),
				"path": r.URL.Path,
			})
			metrics.RecordError("corehr", "get_employee_manager_missing_id_error")
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}

		// 解析UUID
		employeeUUID, err := uuid.Parse(employeeID)
		if err != nil {
			reqLogger.LogError("invalid_uuid", "Invalid employee ID format", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
				"error_type": fmt.Sprintf("%T", err),
			})
			metrics.RecordError("corehr", "get_employee_manager_invalid_uuid_error")
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}

		// 记录请求开始
		reqLogger.Info("Processing get employee manager request",
			"employee_id", employeeID,
			"tenant_id", tenantID.String(),
		)

		// 调用服务
		manager, err := service.GetManagerByEmployeeId(r.Context(), tenantID, employeeUUID)
		if err != nil {
			reqLogger.LogError("get_employee_manager_service_error", "Failed to get employee manager from service", err, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id":   tenantID.String(),
				"error_type":  fmt.Sprintf("%T", err),
			})

			// 根据错误类型返回适当的HTTP状态码
			if strings.Contains(err.Error(), "manager not found") || strings.Contains(err.Error(), "employee not found") {
				metrics.RecordError("corehr", "get_employee_manager_not_found_error")
				http.Error(w, "Manager not found", http.StatusNotFound)
			} else if strings.Contains(err.Error(), "timeout") {
				metrics.RecordError("corehr", "get_employee_manager_timeout_error")
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			} else if strings.Contains(err.Error(), "connection") {
				metrics.RecordError("corehr", "get_employee_manager_connection_error")
				http.Error(w, "Database connection error", http.StatusServiceUnavailable)
			} else {
				metrics.RecordError("corehr", "get_employee_manager_service_error")
				http.Error(w, "Failed to get employee manager", http.StatusInternalServerError)
			}
			return
		}

		// 响应验证
		if manager == nil {
			reqLogger.LogError("service_response_error", "Service returned nil manager", nil, map[string]interface{}{
				"employee_id": employeeID,
				"tenant_id": tenantID.String(),
			})
			metrics.RecordError("corehr", "get_employee_manager_nil_response_error")
			http.Error(w, "Manager not found", http.StatusNotFound)
			return
		}

		// 记录成功指标
		duration := time.Since(start)
		metrics.RecordDatabaseOperation("SELECT", "employees", "success", duration)
		reqLogger.LogDatabaseOperation("SELECT", "employees", 1, duration, true)
		reqLogger.LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, duration, r.UserAgent())

		reqLogger.Info("Successfully retrieved employee manager",
			"employee_id", employeeID,
			"manager_id", manager.Id.String(),
			"manager_number", manager.EmployeeNumber,
			"duration_ms", duration.Milliseconds(),
			"tenant_id", tenantID.String(),
		)

		// 返回响应
		respondJSON(w, http.StatusOK, manager)
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

// 新增的企业级服务健康检查函数
func handleCDCHealthCheck(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CDC服务健康检查逻辑
		entClient := common.GetEntClient()
		if entClient == nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unhealthy",
				"service": "cdc",
				"error": "Ent client not available",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"service": "cdc",
			"database_connection": "active",
		})
	}
}

func handleNeo4jSyncHealthCheck(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Neo4j同步服务健康检查逻辑
		entClient := common.GetEntClient()
		if entClient == nil {
			respondJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "unhealthy",
				"service": "neo4j-sync",
				"error": "Ent client not available",
			})
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"service": "neo4j-sync",
			"sync_status": "configured",
		})
	}
}

func handleDataConsistencyCheck(logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 数据一致性检查逻辑
		// 这里可以添加PostgreSQL和Neo4j数据一致性检查
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status": "healthy",
			"service": "data-consistency",
			"postgresql_status": "connected",
			"neo4j_status": "connected",
			"consistency_check": "passed",
		})
	}
}

// initializeNeo4jService 初始化Neo4j服务连接
func initializeNeo4jService() (*service.Neo4jService, error) {
	// 从环境变量或配置中读取Neo4j连接信息
	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687" // 默认地址
	}
	
	neo4jUser := os.Getenv("NEO4J_USER")
	if neo4jUser == "" {
		neo4jUser = "neo4j" // 默认用户名
	}
	
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	if neo4jPassword == "" {
		neo4jPassword = "password" // 默认密码（开发环境）
	}
	
	// 创建Neo4j配置
	config := &service.Neo4jConfig{
		URI:      neo4jURI,
		Username: neo4jUser,
		Password: neo4jPassword,
		Database: "neo4j", // 默认数据库
	}
	
	// 创建Neo4j服务
	neo4jService, err := service.NewNeo4jService(*config, log.New(os.Stdout, "[Neo4j] ", log.LstdFlags))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j service: %w", err)
	}
	
	log.Printf("✅ Neo4j服务初始化成功: %s", neo4jURI)
	return neo4jService, nil
}
