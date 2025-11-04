package repository

import (
	"database/sql"

	pkglogger "cube-castle/pkg/logger"
	"github.com/redis/go-redis/v9"
)

type AuditHistoryConfig struct {
	StrictValidation        bool
	AllowFallback           bool
	CircuitBreakerThreshold int32
	LegacyMode              bool
}

// PostgreSQLRepository 提供查询服务的数据访问能力
type PostgreSQLRepository struct {
	db                     *sql.DB
	redisClient            *redis.Client
	logger                 pkglogger.Logger
	auditConfig            AuditHistoryConfig
	validationFailureCount int32
}

func NewPostgreSQLRepository(db *sql.DB, redisClient *redis.Client, logger pkglogger.Logger, auditConfig AuditHistoryConfig) *PostgreSQLRepository {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &PostgreSQLRepository{
		db:          db,
		redisClient: redisClient,
		logger: logger.WithFields(pkglogger.Fields{
			"component": "queryRepository",
		}),
		auditConfig: auditConfig,
	}
}
