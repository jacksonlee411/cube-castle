package handler

import (
	"net/http"

	"cube-castle/internal/organization/audit"
	"cube-castle/internal/organization/repository"
	scheduler "cube-castle/internal/organization/scheduler"
	"cube-castle/internal/organization/validator"
	standardobject "cube-castle/internal/standardobject"
	pkglogger "cube-castle/pkg/logger"
	clockpkg "cube-castle/pkg/temporal/clock"
)

type OrganizationHandler struct {
	repo            *repository.OrganizationRepository
	temporalService *scheduler.TemporalService
	auditLogger     *audit.AuditLogger
	logger          pkglogger.Logger
	timelineManager *repository.TemporalTimelineManager
	hierarchyRepo   *repository.HierarchyRepository
	validator       *validator.BusinessRuleValidator
	standardObjects standardobject.ObjectService
	clock           clockpkg.Clock
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *scheduler.TemporalService, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validator.BusinessRuleValidator, stdObjects standardobject.ObjectService, clk clockpkg.Clock) *OrganizationHandler {
	if stdObjects == nil {
		stdObjects = standardobject.NewNoopService()
	}
	if clk == nil {
		clk = clockpkg.NewSystemClock()
	}
	return &OrganizationHandler{
		repo:            repo,
		temporalService: temporalService,
		auditLogger:     auditLogger,
		logger: scopedLogger(baseLogger, "organization", pkglogger.Fields{
			"module": "organization",
		}),
		timelineManager: timelineManager,
		hierarchyRepo:   hierarchyRepo,
		validator:       validator,
		standardObjects: stdObjects,
		clock:           clk,
	}
}

func (h *OrganizationHandler) requestLogger(r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	return requestScopedLogger(h.logger, r, action, extra)
}
