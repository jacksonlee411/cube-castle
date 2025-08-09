package container

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"

	"github.com/cube-castle/cmd/organization-command-server/internal/application/handlers"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/services"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/config"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/messaging"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/persistence/postgres"
	httpHandlers "github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/handlers"
	"github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/middleware"
	"github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/routes"
)

// Container holds all the application dependencies
type Container struct {
	config *config.Config
	logger logging.Logger

	// Infrastructure
	dbPool     *pgxpool.Pool
	eventBus   handlers.EventBus
	kafkaAdmin kafka.AdminClient

	// Repositories
	organizationRepo *postgres.PostgresOrganizationRepository

	// Services
	organizationService *services.OrganizationService

	// Application Handlers
	organizationHandler *handlers.OrganizationHandler

	// HTTP Handlers
	organizationHTTPHandler *httpHandlers.OrganizationHTTPHandler
	healthHandler           *httpHandlers.HealthHandler

	// Middleware
	errorHandler  *middleware.ErrorHandler
	requestLogger *middleware.RequestLogger

	// Server
	httpServer *http.Server
}

// NewContainer creates and initializes a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	c := &Container{config: cfg}

	// Initialize components in dependency order
	if err := c.initLogger(); err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	if err := c.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	if err := c.initKafka(); err != nil {
		return nil, fmt.Errorf("failed to init kafka: %w", err)
	}

	if err := c.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to init repositories: %w", err)
	}

	if err := c.initServices(); err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	if err := c.initApplicationHandlers(); err != nil {
		return nil, fmt.Errorf("failed to init application handlers: %w", err)
	}

	if err := c.initHTTPComponents(); err != nil {
		return nil, fmt.Errorf("failed to init HTTP components: %w", err)
	}

	if err := c.initHTTPServer(); err != nil {
		return nil, fmt.Errorf("failed to init HTTP server: %w", err)
	}

	return c, nil
}

// GetHTTPServer returns the configured HTTP server
func (c *Container) GetHTTPServer() *http.Server {
	return c.httpServer
}

// Close closes all resources
func (c *Container) Close() {
	c.logger.Info("shutting down container")

	if c.eventBus != nil {
		if eb, ok := c.eventBus.(*messaging.KafkaEventBus); ok {
			eb.Close()
		}
	}

	if c.kafkaAdmin != nil {
		c.kafkaAdmin.Close()
	}

	if c.dbPool != nil {
		c.dbPool.Close()
	}

	c.logger.Info("container shutdown completed")
}

// Private initialization methods

func (c *Container) initLogger() error {
	logger, err := logging.NewSlogLogger(
		c.config.Logger.Level,
		c.config.Logger.Format,
		c.config.Logger.Output,
		c.config.Logger.TimeFormat,
	)
	if err != nil {
		return err
	}

	c.logger = logger.With(
		"service", c.config.App.Name,
		"version", c.config.App.Version,
		"environment", c.config.App.Environment,
	)

	c.logger.Info("logger initialized",
		"level", c.config.Logger.Level,
		"format", c.config.Logger.Format,
	)

	return nil
}

func (c *Container) initDatabase() error {
	dbConfig, err := pgxpool.ParseConfig(c.config.Database.GetConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = int32(c.config.Database.MaxConnections)
	dbConfig.MinConns = int32(c.config.Database.MinConnections)
	dbConfig.MaxConnLifetime = c.config.Database.MaxConnLifetime

	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	c.dbPool = pool
	c.logger.Info("database connection initialized",
		"host", c.config.Database.Host,
		"port", c.config.Database.Port,
		"database", c.config.Database.Database,
		"max_conns", c.config.Database.MaxConnections,
	)

	return nil
}

func (c *Container) initKafka() error {
	// Create Kafka admin client for health checks
	adminConfig := &kafka.ConfigMap{
		"bootstrap.servers": c.config.Kafka.Brokers[0], // Use first broker for admin
	}

	admin, err := kafka.NewAdminClient(adminConfig)
	if err != nil {
		return fmt.Errorf("failed to create kafka admin client: %w", err)
	}

	c.kafkaAdmin = admin

	// Create event bus
	eventBus, err := messaging.NewKafkaEventBus(
		c.config.Kafka.Brokers,
		c.config.Kafka.EventTopic,
		c.config.Kafka.ClientID,
		c.logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create kafka event bus: %w", err)
	}

	// Wrap with retry capability
	c.eventBus = messaging.NewRetryableEventBus(
		eventBus,
		c.logger,
		3,                  // max retries
		1*time.Second,      // initial delay
	)

	c.logger.Info("kafka components initialized",
		"brokers", c.config.Kafka.Brokers,
		"topic", c.config.Kafka.EventTopic,
		"client_id", c.config.Kafka.ClientID,
	)

	return nil
}

func (c *Container) initRepositories() error {
	c.organizationRepo = postgres.NewPostgresOrganizationRepository(c.dbPool, c.logger)

	c.logger.Info("repositories initialized")
	return nil
}

func (c *Container) initServices() error {
	c.organizationService = services.NewOrganizationService(c.organizationRepo)

	c.logger.Info("domain services initialized")
	return nil
}

func (c *Container) initApplicationHandlers() error {
	c.organizationHandler = handlers.NewOrganizationHandler(
		c.organizationRepo,
		c.organizationService,
		c.eventBus,
		c.logger,
	)

	c.logger.Info("application handlers initialized")
	return nil
}

func (c *Container) initHTTPComponents() error {
	// Initialize middleware
	c.errorHandler = middleware.NewErrorHandler(c.logger)
	c.requestLogger = middleware.NewRequestLogger(c.logger)

	// Parse default tenant ID
	defaultTenantID, err := uuid.Parse(c.config.App.DefaultTenantID)
	if err != nil {
		return fmt.Errorf("failed to parse default tenant ID: %w", err)
	}

	// Initialize HTTP handlers
	c.organizationHTTPHandler = httpHandlers.NewOrganizationHTTPHandler(
		c.organizationHandler,
		c.errorHandler,
		c.logger,
		defaultTenantID,
	)

	c.healthHandler = httpHandlers.NewHealthHandler(
		c.dbPool,
		c.kafkaAdmin,
		c.logger,
		c.config.App.Version,
	)

	c.logger.Info("HTTP components initialized")
	return nil
}

func (c *Container) initHTTPServer() error {
	// Setup routes
	router := routes.SetupRoutes(routes.RouterConfig{
		OrganizationHandler: c.organizationHTTPHandler,
		HealthHandler:       c.healthHandler,
		RequestLogger:       c.requestLogger,
		ErrorHandler:        c.errorHandler,
	})

	// Create HTTP server
	c.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", c.config.Server.Port),
		Handler:      router,
		ReadTimeout:  c.config.Server.ReadTimeout,
		WriteTimeout: c.config.Server.WriteTimeout,
		IdleTimeout:  c.config.Server.IdleTimeout,
	}

	c.logger.Info("HTTP server initialized",
		"port", c.config.Server.Port,
		"read_timeout", c.config.Server.ReadTimeout,
		"write_timeout", c.config.Server.WriteTimeout,
	)

	return nil
}