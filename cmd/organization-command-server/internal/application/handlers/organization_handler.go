package handlers

import (
	"context"
	"fmt"

	"github.com/cube-castle/cmd/organization-command-server/internal/application/commands"
	"github.com/cube-castle/cmd/organization-command-server/internal/application/dtos"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/entities"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/repositories"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/services"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/valueobjects"
)

// EventBus defines the interface for publishing domain events
type EventBus interface {
	Publish(ctx context.Context, event entities.DomainEvent) error
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// OrganizationHandler handles organization-related commands
type OrganizationHandler struct {
	repo           repositories.OrganizationRepository
	domainService  *services.OrganizationService
	eventBus       EventBus
	logger         Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(
	repo repositories.OrganizationRepository,
	domainService *services.OrganizationService,
	eventBus EventBus,
	logger Logger,
) *OrganizationHandler {
	return &OrganizationHandler{
		repo:          repo,
		domainService: domainService,
		eventBus:      eventBus,
		logger:        logger,
	}
}

// HandleCreateOrganization handles the create organization command
func (h *OrganizationHandler) HandleCreateOrganization(ctx context.Context, cmd commands.CreateOrganizationCommand) (*dtos.CreateOrganizationResult, error) {
	h.logger.Info("Processing create organization command",
		"command_id", cmd.CommandID,
		"tenant_id", cmd.TenantID,
		"name", cmd.Name,
	)

	// 1. Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// 2. Determine organization code
	var code valueobjects.OrganizationCode
	var err error
	
	if cmd.RequestedCode != nil && *cmd.RequestedCode != "" {
		code, err = valueobjects.NewOrganizationCode(*cmd.RequestedCode)
		if err != nil {
			return nil, fmt.Errorf("invalid organization code: %w", err)
		}
	} else {
		code, err = h.domainService.GenerateOrganizationCode(ctx, cmd.TenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate organization code: %w", err)
		}
	}

	// 3. Parse parent code if provided
	var parentCode *valueobjects.OrganizationCode
	if cmd.ParentCode != nil && *cmd.ParentCode != "" {
		pc, err := valueobjects.NewOrganizationCode(*cmd.ParentCode)
		if err != nil {
			return nil, fmt.Errorf("invalid parent organization code: %w", err)
		}
		parentCode = &pc
	}

	// 4. Validate business rules
	if err := h.domainService.ValidateCreateOrganization(ctx, code, cmd.Name, parentCode, cmd.TenantID); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// 5. Parse unit type
	unitType, err := entities.ParseUnitType(cmd.UnitType)
	if err != nil {
		return nil, fmt.Errorf("invalid unit type: %w", err)
	}

	// 6. Calculate hierarchy
	level, path, err := h.domainService.CalculateOrganizationHierarchy(ctx, code, parentCode, cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate organization hierarchy: %w", err)
	}

	// 7. Determine sort order
	sortOrder := 0
	if cmd.SortOrder != nil {
		sortOrder = *cmd.SortOrder
	}

	// 8. Create organization entity
	org, err := entities.NewOrganization(
		code,
		cmd.Name,
		unitType,
		cmd.TenantID,
		parentCode,
		level,
		path,
		sortOrder,
		cmd.Description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization entity: %w", err)
	}

	// 9. Persist organization
	if err := h.repo.Create(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to save organization: %w", err)
	}

	// 10. Publish domain events
	for _, event := range org.GetEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			h.logger.Warn("failed to publish event", "event_type", event.GetEventType(), "error", err)
			// Note: Event publishing failure should not rollback the business operation
		}
	}

	// 11. Clear events and log success
	org.ClearEvents()

	h.logger.Info("Organization created successfully",
		"code", code.String(),
		"name", cmd.Name,
		"command_id", cmd.CommandID,
	)

	// 12. Return result
	return &dtos.CreateOrganizationResult{
		Code:      code.String(),
		Name:      cmd.Name,
		UnitType:  unitType.String(),
		Status:    entities.StatusActive.String(),
		Level:     level,
		Path:      path,
		CreatedAt: org.CreatedAt(),
	}, nil
}

// HandleUpdateOrganization handles the update organization command
func (h *OrganizationHandler) HandleUpdateOrganization(ctx context.Context, cmd commands.UpdateOrganizationCommand) (*dtos.UpdateOrganizationResult, error) {
	h.logger.Info("Processing update organization command",
		"command_id", cmd.CommandID,
		"tenant_id", cmd.TenantID,
		"code", cmd.Code,
	)

	// 1. Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// 2. Parse organization code
	code, err := valueobjects.NewOrganizationCode(cmd.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid organization code: %w", err)
	}

	// 3. Validate business rules
	if err := h.domainService.ValidateUpdateOrganization(ctx, code, cmd.TenantID); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// 4. Retrieve organization
	org, err := h.repo.FindByCode(ctx, code, cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// 5. Track changes for response
	changes := make(map[string]interface{})

	// 6. Apply updates
	if cmd.Name != nil {
		if err := org.UpdateName(*cmd.Name, cmd.RequestedBy); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
		changes["name"] = *cmd.Name
	}

	if cmd.Status != nil {
		status, err := entities.ParseStatus(*cmd.Status)
		if err != nil {
			return nil, fmt.Errorf("invalid status: %w", err)
		}
		if err := org.UpdateStatus(status, cmd.RequestedBy); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
		changes["status"] = *cmd.Status
	}

	if cmd.Description != nil {
		if err := org.UpdateDescription(cmd.Description, cmd.RequestedBy); err != nil {
			return nil, fmt.Errorf("failed to update description: %w", err)
		}
		changes["description"] = *cmd.Description
	}

	if cmd.SortOrder != nil {
		if err := org.UpdateSortOrder(*cmd.SortOrder, cmd.RequestedBy); err != nil {
			return nil, fmt.Errorf("failed to update sort order: %w", err)
		}
		changes["sort_order"] = *cmd.SortOrder
	}

	// 7. Check if there were any changes
	if len(changes) == 0 {
		h.logger.Info("No changes detected for organization update", "code", cmd.Code)
		return &dtos.UpdateOrganizationResult{
			Code:      cmd.Code,
			UpdatedAt: org.UpdatedAt(),
			Changes:   changes,
		}, nil
	}

	// 8. Persist changes
	if err := h.repo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to save organization updates: %w", err)
	}

	// 9. Publish domain events
	for _, event := range org.GetEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			h.logger.Warn("failed to publish event", "event_type", event.GetEventType(), "error", err)
		}
	}

	// 10. Clear events and log success
	org.ClearEvents()

	h.logger.Info("Organization updated successfully",
		"code", cmd.Code,
		"changes", changes,
		"command_id", cmd.CommandID,
	)

	// 11. Return result
	return &dtos.UpdateOrganizationResult{
		Code:      cmd.Code,
		UpdatedAt: org.UpdatedAt(),
		Changes:   changes,
	}, nil
}

// HandleDeleteOrganization handles the delete organization command
func (h *OrganizationHandler) HandleDeleteOrganization(ctx context.Context, cmd commands.DeleteOrganizationCommand) (*dtos.DeleteOrganizationResult, error) {
	h.logger.Info("Processing delete organization command",
		"command_id", cmd.CommandID,
		"tenant_id", cmd.TenantID,
		"code", cmd.Code,
	)

	// 1. Validate command
	if err := cmd.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// 2. Parse organization code
	code, err := valueobjects.NewOrganizationCode(cmd.Code)
	if err != nil {
		return nil, fmt.Errorf("invalid organization code: %w", err)
	}

	// 3. Validate business rules
	if err := h.domainService.ValidateDeleteOrganization(ctx, code, cmd.TenantID); err != nil {
		return nil, fmt.Errorf("business rule validation failed: %w", err)
	}

	// 4. Retrieve organization
	org, err := h.repo.FindByCode(ctx, code, cmd.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to find organization: %w", err)
	}

	// 5. Mark as deleted
	if err := org.MarkAsDeleted(cmd.RequestedBy); err != nil {
		return nil, fmt.Errorf("failed to mark organization as deleted: %w", err)
	}

	// 6. Persist changes
	if err := h.repo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to save organization deletion: %w", err)
	}

	// 7. Publish domain events
	for _, event := range org.GetEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			h.logger.Warn("failed to publish event", "event_type", event.GetEventType(), "error", err)
		}
	}

	// 8. Clear events and log success
	org.ClearEvents()

	h.logger.Info("Organization deleted successfully",
		"code", cmd.Code,
		"command_id", cmd.CommandID,
	)

	// 9. Return result
	return &dtos.DeleteOrganizationResult{
		Code:      cmd.Code,
		DeletedAt: org.UpdatedAt(),
	}, nil
}