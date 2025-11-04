package handlers

import (
	"net/http"

	"cube-castle/cmd/hrms-server/command/internal/audit"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/cmd/hrms-server/command/internal/validators"
	pkglogger "cube-castle/pkg/logger"
)

type OrganizationHandler struct {
	repo            *repository.OrganizationRepository
	temporalService *services.TemporalService
	auditLogger     *audit.AuditLogger
	logger          pkglogger.Logger
	timelineManager *repository.TemporalTimelineManager
	hierarchyRepo   *repository.HierarchyRepository
	validator       *validators.BusinessRuleValidator
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *services.TemporalService, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validators.BusinessRuleValidator) *OrganizationHandler {
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
