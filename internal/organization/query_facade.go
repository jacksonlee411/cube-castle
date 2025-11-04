package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/dto"
	repositorypkg "cube-castle/internal/organization/repository"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	defaultAssignmentCacheTTL = 2 * time.Minute
	statsCachePrefix          = "org:assignment:stats"
)

type assignmentRepository interface {
	GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentHistory(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *dto.PositionAssignmentFilterInput, pagination *dto.PaginationInput, sorting []dto.PositionAssignmentSortInput) (*dto.PositionAssignmentConnection, error)
	GetAssignmentStats(ctx context.Context, tenantID uuid.UUID, positionCode string, organizationCode string) (*dto.AssignmentStats, error)
}

// AssignmentQueryFacade 封装任职查询与缓存刷新逻辑。
type AssignmentQueryFacade struct {
	repo     assignmentRepository
	redis    *redis.Client
	logger   pkglogger.Logger
	cacheTTL time.Duration
}

// NewAssignmentQueryFacade 创建 AssignmentQueryFacade。
func NewAssignmentQueryFacade(repo assignmentRepository, redisClient *redis.Client, logger pkglogger.Logger, cacheTTL time.Duration) *AssignmentQueryFacade {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	if cacheTTL <= 0 {
		cacheTTL = defaultAssignmentCacheTTL
	}
	return &AssignmentQueryFacade{
		repo:     repo,
		redis:    redisClient,
		logger:   logger.WithFields(pkglogger.Fields{"component": "assignment-facade"}),
		cacheTTL: cacheTTL,
	}
}

// GetAssignments 获取职位任职列表（不强制缓存，保持实时读取）。
func (f *AssignmentQueryFacade) GetAssignments(
	ctx context.Context,
	tenantID uuid.UUID,
	positionCode string,
	filter *dto.PositionAssignmentFilterInput,
	pagination *dto.PaginationInput,
	sorting []dto.PositionAssignmentSortInput,
) (*dto.PositionAssignmentConnection, error) {
	if f.repo == nil {
		return nil, fmt.Errorf("assignment repository not configured")
	}
	return f.repo.GetPositionAssignments(ctx, tenantID, strings.TrimSpace(positionCode), filter, pagination, sorting)
}

// GetAssignmentHistory 获取职位任职历史记录。
func (f *AssignmentQueryFacade) GetAssignmentHistory(
	ctx context.Context,
	tenantID uuid.UUID,
	positionCode string,
	filter *dto.PositionAssignmentFilterInput,
	pagination *dto.PaginationInput,
	sorting []dto.PositionAssignmentSortInput,
) (*dto.PositionAssignmentConnection, error) {
	if f.repo == nil {
		return nil, fmt.Errorf("assignment repository not configured")
	}
	return f.repo.GetAssignmentHistory(ctx, tenantID, strings.TrimSpace(positionCode), filter, pagination, sorting)
}

// GetAssignmentStats 获取任职统计信息，并对单个职位的统计结果启用 Redis 缓存。
func (f *AssignmentQueryFacade) GetAssignmentStats(
	ctx context.Context,
	tenantID uuid.UUID,
	positionCode string,
	organizationCode string,
) (*dto.AssignmentStats, error) {
	if f.repo == nil {
		return nil, fmt.Errorf("assignment repository not configured")
	}

	positionCode = strings.TrimSpace(positionCode)
	organizationCode = strings.TrimSpace(organizationCode)
	useCache := f.redis != nil && positionCode != ""

	cacheKey := ""
	if useCache {
		cacheKey = f.statsCacheKey(tenantID, positionCode)
		if cached, err := f.redis.Get(ctx, cacheKey).Result(); err == nil {
			var stats dto.AssignmentStats
			if json.Unmarshal([]byte(cached), &stats) == nil {
				f.logger.WithFields(pkglogger.Fields{
					"tenantId":     tenantID.String(),
					"positionCode": positionCode,
					"cacheKey":     cacheKey,
				}).Debug("assignment stats served from cache")
				return &stats, nil
			}
		}
	}

	stats, err := f.repo.GetAssignmentStats(ctx, tenantID, positionCode, organizationCode)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		stats = &dto.AssignmentStats{}
	}

	if useCache && stats != nil {
		data, err := json.Marshal(stats)
		if err == nil {
			if err := f.redis.Set(ctx, cacheKey, data, f.cacheTTL).Err(); err != nil {
				f.logger.WithFields(pkglogger.Fields{
					"cacheKey": cacheKey,
					"error":    err,
				}).Warn("failed to cache assignment stats")
			}
		}
	}

	return stats, nil
}

// RefreshPositionCache 失效职位相关的任职统计缓存。
func (f *AssignmentQueryFacade) RefreshPositionCache(ctx context.Context, tenantID uuid.UUID, positionCode string) error {
	if f.redis == nil {
		f.logger.Debug("redis client not configured, skip cache refresh")
		return nil
	}
	positionCode = strings.TrimSpace(positionCode)
	if positionCode == "" {
		return fmt.Errorf("positionCode is required for cache refresh")
	}

	pattern := fmt.Sprintf("%s:%s:%s:*", statsCachePrefix, tenantID.String(), strings.ToUpper(positionCode))
	iter := f.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if err := f.redis.Del(ctx, key).Err(); err != nil {
			f.logger.WithFields(pkglogger.Fields{
				"cacheKey": key,
				"error":    err,
			}).Warn("failed to delete assignment cache key")
		} else {
			f.logger.WithFields(pkglogger.Fields{
				"cacheKey": key,
				"tenantId": tenantID.String(),
			}).Debug("assignment cache key invalidated")
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("scan assignment cache keys failed: %w", err)
	}
	return nil
}

func (f *AssignmentQueryFacade) statsCacheKey(tenantID uuid.UUID, positionCode string) string {
	return fmt.Sprintf("%s:%s:%s:%s", statsCachePrefix, tenantID.String(), strings.ToUpper(positionCode), "v1")
}

// Ensure repository satisfies interface compile-time.
var _ assignmentRepository = (*repositorypkg.PostgreSQLRepository)(nil)
