package repository

import (
	"database/sql"
	"log"

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
	logger                 *log.Logger
	auditConfig            AuditHistoryConfig
	validationFailureCount int32
}

func NewPostgreSQLRepository(db *sql.DB, redisClient *redis.Client, logger *log.Logger, auditConfig AuditHistoryConfig) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		auditConfig: auditConfig,
	}
}
