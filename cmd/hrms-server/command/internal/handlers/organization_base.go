package handlers

import (
	"log"

	"cube-castle/cmd/hrms-server/command/internal/audit"
	"cube-castle/cmd/hrms-server/command/internal/repository"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/cmd/hrms-server/command/internal/validators"
)

type OrganizationHandler struct {
	repo            *repository.OrganizationRepository
	temporalService *services.TemporalService
	auditLogger     *audit.AuditLogger
	logger          *log.Logger
	timelineManager *repository.TemporalTimelineManager
	hierarchyRepo   *repository.HierarchyRepository
	validator       *validators.BusinessRuleValidator
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *services.TemporalService, auditLogger *audit.AuditLogger, logger *log.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validators.BusinessRuleValidator) *OrganizationHandler {
	return &OrganizationHandler{
		repo:            repo,
		temporalService: temporalService,
		auditLogger:     auditLogger,
		logger:          logger,
		timelineManager: timelineManager,
		hierarchyRepo:   hierarchyRepo,
		validator:       validator,
	}
}
