package scheduler

import (
	"context"
	"database/sql"

	"cube-castle/internal/organization/repository"
	servicepkg "cube-castle/internal/organization/service"
	pkglogger "cube-castle/pkg/logger"
)

// Service 聚合调度与 Temporal 能力，为命令模块提供统一入口。
type Service struct {
	temporal    *TemporalService
	monitor     *TemporalMonitor
	operational *OperationalScheduler
	orgTemporal *OrganizationTemporalService
	logger      pkglogger.Logger
}

// Dependencies 构建 Service 所需依赖。
type Dependencies struct {
	DB                     *sql.DB
	Logger                 pkglogger.Logger
	OrganizationRepository *repository.OrganizationRepository
	PositionService        *servicepkg.PositionService
}

// NewService 创建调度聚合服务。
func NewService(deps Dependencies) *Service {
	logger := deps.Logger
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}

	temporal := NewTemporalService(deps.DB, logger, deps.OrganizationRepository)
	monitor := NewTemporalMonitor(deps.DB, logger)
	operational := NewOperationalScheduler(deps.DB, logger, monitor, deps.PositionService)
	orgTemporal := NewOrganizationTemporalService(deps.DB, logger)

	return &Service{
		temporal:    temporal,
		monitor:     monitor,
		operational: operational,
		orgTemporal: orgTemporal,
		logger:      logger,
	}
}

// Temporal 返回 TemporalService。
func (s *Service) Temporal() *TemporalService {
	return s.temporal
}

// Monitor 返回 TemporalMonitor。
func (s *Service) Monitor() *TemporalMonitor {
	return s.monitor
}

// Operational 返回 OperationalScheduler。
func (s *Service) Operational() *OperationalScheduler {
	return s.operational
}

// OrganizationTemporal 返回 OrganizationTemporalService。
func (s *Service) OrganizationTemporal() *OrganizationTemporalService {
	return s.orgTemporal
}

// Start 启动调度相关后台任务。
func (s *Service) Start(ctx context.Context) {
	if s.operational != nil {
		s.operational.Start(ctx)
	}
}

// Stop 停止调度后台任务。
func (s *Service) Stop() {
	if s.operational != nil {
		s.operational.Stop()
	}
}
