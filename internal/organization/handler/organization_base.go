package handler

import (
	"net/http"

	"cube-castle/internal/organization/audit"
	"cube-castle/internal/organization/repository"
	scheduler "cube-castle/internal/organization/scheduler"
	"cube-castle/internal/organization/validator"
	pkglogger "cube-castle/pkg/logger"
)

type OrganizationHandler struct {
	repo            *repository.OrganizationRepository
	temporalService *scheduler.TemporalService
	auditLogger     *audit.AuditLogger
	logger          pkglogger.Logger
	timelineManager *repository.TemporalTimelineManager
	hierarchyRepo   *repository.HierarchyRepository
	validator       *validator.BusinessRuleValidator
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *scheduler.TemporalService, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validator.BusinessRuleValidator) *OrganizationHandler {
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
	}
}

func (h *OrganizationHandler) requestLogger(r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	return requestScopedLogger(h.logger, r, action, extra)
}
