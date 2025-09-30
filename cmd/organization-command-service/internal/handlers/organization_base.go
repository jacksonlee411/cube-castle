package handlers

import (
	"log"

	"organization-command-service/internal/audit"
	"organization-command-service/internal/repository"
	"organization-command-service/internal/services"
	"organization-command-service/internal/validators"
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
