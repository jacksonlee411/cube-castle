package organization

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	auth "cube-castle/internal/auth"
	auditpkg "cube-castle/internal/organization/audit"
	dto "cube-castle/internal/organization/dto"
	handlerpkg "cube-castle/internal/organization/handler"
	middlewarepkg "cube-castle/internal/organization/middleware"
	repositorypkg "cube-castle/internal/organization/repository"
	"cube-castle/internal/organization/resolver"
	servicepkg "cube-castle/internal/organization/service"
	utilspkg "cube-castle/internal/organization/utils"
	validatorpkg "cube-castle/internal/organization/validator"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CommandModuleDeps struct {
	DB              *sql.DB
	Logger          pkglogger.Logger
	CascadeMaxDepth int
}

type OrganizationHandler = handlerpkg.OrganizationHandler
type PositionHandler = handlerpkg.PositionHandler
type JobCatalogHandler = handlerpkg.JobCatalogHandler
type OperationalHandler = handlerpkg.OperationalHandler
type DevToolsHandler = handlerpkg.DevToolsHandler
type AuditLogger = auditpkg.AuditLogger
type AuditHistoryConfig = repositorypkg.AuditHistoryConfig
type QueryRepository = repositorypkg.PostgreSQLRepository
type QueryRepositoryInterface = resolver.QueryRepository
type QueryResolver = resolver.Resolver
type QueryPermissionChecker = resolver.PermissionChecker
type AssignmentFacade interface {
	GetAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
	RefreshPositionCache(ctx context.Context, tenantID uuid.UUID, positionCode string) error
}

type CommandModule struct {
	DB           *sql.DB
	Logger       pkglogger.Logger
	Repositories CommandRepositories
	Services     CommandServices
	Validator    *validatorpkg.BusinessRuleValidator
	AuditLogger  *auditpkg.AuditLogger
}

type CommandRepositories struct {
	Organization       *repositorypkg.OrganizationRepository
	JobCatalog         *repositorypkg.JobCatalogRepository
	Position           *repositorypkg.PositionRepository
	PositionAssignment *repositorypkg.PositionAssignmentRepository
	Hierarchy          *repositorypkg.HierarchyRepository
	TemporalTimeline   *repositorypkg.TemporalTimelineManager
}

type CommandServices struct {
	Cascade              *servicepkg.CascadeUpdateService
	Temporal             *servicepkg.TemporalService
	TemporalMonitor      *servicepkg.TemporalMonitor
	OperationalScheduler *servicepkg.OperationalScheduler
	Position             *servicepkg.PositionService
	JobCatalog           *servicepkg.JobCatalogService
}

type CommandHandlers struct {
	Organization *handlerpkg.OrganizationHandler
	Position     *handlerpkg.PositionHandler
	JobCatalog   *handlerpkg.JobCatalogHandler
	Operational  *handlerpkg.OperationalHandler
	DevTools     *handlerpkg.DevToolsHandler
}

type CommandHandlerDeps struct {
	JWTMiddleware       *auth.JWTMiddleware
	RateLimitMiddleware *middlewarepkg.RateLimitMiddleware
	Logger              pkglogger.Logger
	DevMode             bool
}

type CommandMiddlewares struct {
	Performance *middlewarepkg.PerformanceMiddleware
	RateLimit   *middlewarepkg.RateLimitMiddleware
}

func NewCommandModule(deps CommandModuleDeps) (*CommandModule, error) {
	if deps.DB == nil {
		return nil, ErrMissingDatabase
	}
	logger := deps.Logger
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	cascadeDepth := deps.CascadeMaxDepth
	if cascadeDepth <= 0 {
		cascadeDepth = 4
	}

	orgRepo := repositorypkg.NewOrganizationRepository(deps.DB, logger)
	jobCatalogRepo := repositorypkg.NewJobCatalogRepository(deps.DB, logger)
	positionRepo := repositorypkg.NewPositionRepository(deps.DB, logger)
	positionAssignmentRepo := repositorypkg.NewPositionAssignmentRepository(deps.DB, logger)
	hierarchyRepo := repositorypkg.NewHierarchyRepository(deps.DB, logger)
	timelineManager := repositorypkg.NewTemporalTimelineManager(deps.DB, logger)

	auditLogger := auditpkg.NewAuditLogger(deps.DB, logger)
	cascadeService := servicepkg.NewCascadeUpdateService(hierarchyRepo, cascadeDepth, logger)
	temporalService := servicepkg.NewTemporalService(deps.DB, logger, orgRepo)
	temporalMonitor := servicepkg.NewTemporalMonitor(deps.DB, logger)
	positionService := servicepkg.NewPositionService(positionRepo, positionAssignmentRepo, jobCatalogRepo, orgRepo, auditLogger, logger)
	jobCatalogService := servicepkg.NewJobCatalogService(jobCatalogRepo, auditLogger, logger)
	operationalScheduler := servicepkg.NewOperationalScheduler(deps.DB, logger, temporalMonitor, positionService)

	validator := validatorpkg.NewBusinessRuleValidator(hierarchyRepo, orgRepo, logger)

	module := &CommandModule{
		DB:     deps.DB,
		Logger: logger,
		Repositories: CommandRepositories{
			Organization:       orgRepo,
			JobCatalog:         jobCatalogRepo,
			Position:           positionRepo,
			PositionAssignment: positionAssignmentRepo,
			Hierarchy:          hierarchyRepo,
			TemporalTimeline:   timelineManager,
		},
		Services: CommandServices{
			Cascade:              cascadeService,
			Temporal:             temporalService,
			TemporalMonitor:      temporalMonitor,
			OperationalScheduler: operationalScheduler,
			Position:             positionService,
			JobCatalog:           jobCatalogService,
		},
		Validator:   validator,
		AuditLogger: auditLogger,
	}

	return module, nil
}

func (m *CommandModule) NewHandlers(deps CommandHandlerDeps) CommandHandlers {
	logger := deps.Logger
	if logger == nil {
		logger = m.Logger
	}
	orgHandler := handlerpkg.NewOrganizationHandler(
		m.Repositories.Organization,
		m.Services.Temporal,
		m.AuditLogger,
		logger,
		m.Repositories.TemporalTimeline,
		m.Repositories.Hierarchy,
		m.Validator,
	)
	positionHandler := handlerpkg.NewPositionHandler(m.Services.Position, logger)
	jobCatalogHandler := handlerpkg.NewJobCatalogHandler(m.Services.JobCatalog, logger)
	operationalHandler := handlerpkg.NewOperationalHandler(m.Services.TemporalMonitor, m.Services.OperationalScheduler, deps.RateLimitMiddleware, logger)
	devToolsHandler := handlerpkg.NewDevToolsHandler(deps.JWTMiddleware, logger, deps.DevMode, m.DB)

	return CommandHandlers{
		Organization: orgHandler,
		Position:     positionHandler,
		JobCatalog:   jobCatalogHandler,
		Operational:  operationalHandler,
		DevTools:     devToolsHandler,
	}
}

func NewCommandMiddlewares(logger pkglogger.Logger) CommandMiddlewares {
	rateLimit := middlewarepkg.NewRateLimitMiddleware(middlewarepkg.DefaultRateLimitConfig, logger)
	performance := middlewarepkg.NewPerformanceMiddleware(logger)
	return CommandMiddlewares{
		Performance: performance,
		RateLimit:   rateLimit,
	}
}

var ErrMissingDatabase = errors.New("organization command module requires a database connection")

func RecordHTTPRequest(method, path string, status int) {
	utilspkg.RecordHTTPRequest(method, path, status)
}

func NewDevToolsHandler(db *sql.DB, jwt *auth.JWTMiddleware, logger pkglogger.Logger, devMode bool) *handlerpkg.DevToolsHandler {
	return handlerpkg.NewDevToolsHandler(jwt, logger, devMode, db)
}

func NewQueryRepository(db *sql.DB, redisClient *redis.Client, logger pkglogger.Logger, auditConfig AuditHistoryConfig) *repositorypkg.PostgreSQLRepository {
	return repositorypkg.NewPostgreSQLRepository(db, redisClient, logger, auditConfig)
}

func NewQueryResolver(repo QueryRepositoryInterface, assignments resolver.AssignmentProvider, logger pkglogger.Logger, permissions QueryPermissionChecker) *resolver.Resolver {
	if assignments != nil {
		return resolver.NewResolverWithAssignments(repo, assignments, logger, permissions)
	}
	return resolver.NewResolver(repo, logger, permissions)
}

func NewAssignmentFacade(repo *repositorypkg.PostgreSQLRepository, redisClient *redis.Client, logger pkglogger.Logger, cacheTTL time.Duration) *AssignmentQueryFacade {
	return NewAssignmentQueryFacade(repo, redisClient, logger, cacheTTL)
}

func DefaultAuditHistoryConfig() AuditHistoryConfig {
	return AuditHistoryConfig{
		StrictValidation:        true,
		AllowFallback:           true,
		CircuitBreakerThreshold: 25,
		LegacyMode:              false,
	}
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return middlewarepkg.RequestIDMiddleware(next)
}

func GetRequestID(ctx context.Context) string {
	return middlewarepkg.GetRequestID(ctx)
}
