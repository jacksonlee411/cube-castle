package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"cube-castle/internal/organization/audit"
	"cube-castle/internal/organization/events"
	orgmiddleware "cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/repository"
	scheduler "cube-castle/internal/organization/scheduler"
	"cube-castle/internal/organization/validator"
	standardobject "cube-castle/internal/standardobject"
	"cube-castle/pkg/database"
	pkglogger "cube-castle/pkg/logger"
	clockpkg "cube-castle/pkg/temporal/clock"
	"github.com/google/uuid"
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
	outboxRepo      database.OutboxRepository
}

func NewOrganizationHandler(repo *repository.OrganizationRepository, temporalService *scheduler.TemporalService, auditLogger *audit.AuditLogger, baseLogger pkglogger.Logger, timelineManager *repository.TemporalTimelineManager, hierarchyRepo *repository.HierarchyRepository, validator *validator.BusinessRuleValidator, stdObjects standardobject.ObjectService, clk clockpkg.Clock, outboxRepo database.OutboxRepository) *OrganizationHandler {
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
		outboxRepo:      outboxRepo,
	}
}

func (h *OrganizationHandler) requestLogger(r *http.Request, action string, extra pkglogger.Fields) pkglogger.Logger {
	return requestScopedLogger(h.logger, r, action, extra)
}

func (h *OrganizationHandler) emitStandardObjectEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, eventType string, aggregate standardobject.ObjectAggregate, operation string) error {
	if h.outboxRepo == nil {
		return nil
	}
	eventCtx := events.Context{
		TenantID:      tenantID,
		RequestID:     orgmiddleware.GetRequestID(ctx),
		CorrelationID: orgmiddleware.GetCorrelationID(ctx),
		Operation:     operation,
		Source:        events.DefaultSourceCommand,
	}
	outboxEvent, err := events.NewStandardObjectEvent(eventType, eventCtx, aggregate, nil)
	if err != nil {
		return err
	}
	if tx != nil {
		return h.outboxRepo.Save(ctx, database.WrapSQLTx(tx), outboxEvent)
	}
	if h.repo == nil {
		return fmt.Errorf("repository not configured for outbox")
	}
	standaloneTx, err := h.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer standaloneTx.Rollback()
	if err := h.outboxRepo.Save(ctx, database.WrapSQLTx(standaloneTx), outboxEvent); err != nil {
		return err
	}
	return standaloneTx.Commit()
}
