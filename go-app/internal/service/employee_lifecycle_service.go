package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/ent/positionoccupancyhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/google/uuid"
)

// EmployeeLifecycleService manages the complete employee lifecycle
// from onboarding to offboarding with position integration
type EmployeeLifecycleService struct {
	client               *ent.Client
	logger               *logging.StructuredLogger
	positionAssignmentSvc *PositionAssignmentService
}

// OnboardingRequest represents a new employee onboarding request
type OnboardingRequest struct {
	// Basic employee information
	EmployeeType     string                 `json:"employee_type" validate:"required,oneof=FULL_TIME PART_TIME CONTRACTOR INTERN"`
	EmployeeNumber   string                 `json:"employee_number" validate:"required"`
	FirstName        string                 `json:"first_name" validate:"required"`
	LastName         string                 `json:"last_name" validate:"required"`
	Email            string                 `json:"email" validate:"required,email"`
	PersonalEmail    *string                `json:"personal_email,omitempty"`
	PhoneNumber      *string                `json:"phone_number,omitempty"`
	HireDate         time.Time              `json:"hire_date" validate:"required"`
	EmployeeDetails  map[string]interface{} `json:"employee_details,omitempty"`

	// Initial position assignment
	InitialPositionID   *uuid.UUID `json:"initial_position_id,omitempty"`
	AssignmentStartDate *time.Time `json:"assignment_start_date,omitempty"`
	WorkArrangement     string     `json:"work_arrangement,omitempty" validate:"omitempty,oneof=ON_SITE REMOTE HYBRID"`
	FTEPercentage       float64    `json:"fte_percentage,omitempty"`

	// Onboarding metadata
	OnboardingManager   uuid.UUID `json:"onboarding_manager" validate:"required"`
	OnboardingNotes     string    `json:"onboarding_notes,omitempty"`
	ProbationPeriodDays int       `json:"probation_period_days,omitempty"`
}

// OffboardingRequest represents an employee offboarding request
type OffboardingRequest struct {
	EmployeeID         uuid.UUID  `json:"employee_id" validate:"required"`
	TerminationDate    time.Time  `json:"termination_date" validate:"required"`
	TerminationReason  string     `json:"termination_reason" validate:"required"`
	TerminationType    string     `json:"termination_type" validate:"required,oneof=VOLUNTARY INVOLUNTARY RETIREMENT END_OF_CONTRACT"`
	LastWorkingDate    *time.Time `json:"last_working_date,omitempty"`
	ExitInterviewDate  *time.Time `json:"exit_interview_date,omitempty"`
	FinalPayDate       *time.Time `json:"final_pay_date,omitempty"`
	OffboardingManager uuid.UUID  `json:"offboarding_manager" validate:"required"`
	Notes              string     `json:"notes,omitempty"`
}

// PromotionRequest represents an employee promotion request
type PromotionRequest struct {
	EmployeeID         uuid.UUID `json:"employee_id" validate:"required"`
	NewPositionID      uuid.UUID `json:"new_position_id" validate:"required"`
	PromotionDate      time.Time `json:"promotion_date" validate:"required"`
	PromotionReason    string    `json:"promotion_reason" validate:"required"`
	SalaryAdjustment   *float64  `json:"salary_adjustment,omitempty"`
	ApprovedBy         uuid.UUID `json:"approved_by" validate:"required"`
	EffectiveDate      time.Time `json:"effective_date" validate:"required"`
	WorkArrangement    string    `json:"work_arrangement,omitempty" validate:"omitempty,oneof=ON_SITE REMOTE HYBRID"`
	FTEPercentage      float64   `json:"fte_percentage,omitempty"`
}

// StatusChangeRequest represents an employment status change request
type StatusChangeRequest struct {
	EmployeeID         uuid.UUID  `json:"employee_id" validate:"required"`
	NewStatus          string     `json:"new_status" validate:"required,oneof=ACTIVE ON_LEAVE TERMINATED SUSPENDED PENDING_START"`
	EffectiveDate      time.Time  `json:"effective_date" validate:"required"`
	Reason             string     `json:"reason" validate:"required"`
	ExpectedReturnDate *time.Time `json:"expected_return_date,omitempty"`
	ApprovedBy         uuid.UUID  `json:"approved_by" validate:"required"`
	Notes              string     `json:"notes,omitempty"`
}

// LifecycleEvent represents a tracked employee lifecycle event
type LifecycleEvent struct {
	ID           uuid.UUID              `json:"id"`
	EmployeeID   uuid.UUID              `json:"employee_id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	EventType    string                 `json:"event_type"`
	EventDate    time.Time              `json:"event_date"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedBy    uuid.UUID              `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
}

// OnboardingResult contains the result of employee onboarding
type OnboardingResult struct {
	Success            bool       `json:"success"`
	EmployeeID         uuid.UUID  `json:"employee_id"`
	EmployeeNumber     string     `json:"employee_number"`
	InitialAssignmentID *uuid.UUID `json:"initial_assignment_id,omitempty"`
	OnboardingEventID  uuid.UUID  `json:"onboarding_event_id"`
	Message            string     `json:"message"`
}

// NewEmployeeLifecycleService creates a new EmployeeLifecycleService
func NewEmployeeLifecycleService(client *ent.Client, logger *logging.StructuredLogger) *EmployeeLifecycleService {
	positionAssignmentSvc := NewPositionAssignmentService(client, logger)
	
	return &EmployeeLifecycleService{
		client:               client,
		logger:               logger,
		positionAssignmentSvc: positionAssignmentSvc,
	}
}

// OnboardEmployee handles complete employee onboarding process
func (s *EmployeeLifecycleService) OnboardEmployee(ctx context.Context, tenantID uuid.UUID, req OnboardingRequest) (*OnboardingResult, error) {
	s.logger.Info("Starting employee onboarding",
		"employee_number", req.EmployeeNumber,
		"employee_type", req.EmployeeType,
		"hire_date", req.HireDate,
		"tenant_id", tenantID,
	)

	// Set defaults
	if req.FTEPercentage == 0 {
		req.FTEPercentage = 1.0
	}
	if req.WorkArrangement == "" {
		req.WorkArrangement = "ON_SITE"
	}
	if req.ProbationPeriodDays == 0 {
		req.ProbationPeriodDays = 90 // Default 90-day probation
	}

	// Execute onboarding in transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	result, err := s.executeOnboarding(ctx, tx, tenantID, req)
	if err != nil {
		tx.Rollback()
		s.logger.LogError("employee_onboarding", "Onboarding failed", err, map[string]interface{}{
			"employee_number": req.EmployeeNumber,
			"tenant_id":       tenantID,
		})
		return nil, fmt.Errorf("onboarding execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.LogError("employee_onboarding", "Failed to commit onboarding", err, map[string]interface{}{
			"employee_number": req.EmployeeNumber,
			"tenant_id":       tenantID,
		})
		return nil, fmt.Errorf("failed to commit onboarding: %w", err)
	}

	s.logger.Info("Employee onboarding completed successfully",
		"employee_id", result.EmployeeID,
		"employee_number", result.EmployeeNumber,
		"tenant_id", tenantID,
	)

	return result, nil
}

// OffboardEmployee handles complete employee offboarding process
func (s *EmployeeLifecycleService) OffboardEmployee(ctx context.Context, tenantID uuid.UUID, req OffboardingRequest) error {
	s.logger.Info("Starting employee offboarding",
		"employee_id", req.EmployeeID,
		"termination_date", req.TerminationDate,
		"termination_type", req.TerminationType,
		"tenant_id", tenantID,
	)

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	err = s.executeOffboarding(ctx, tx, tenantID, req)
	if err != nil {
		tx.Rollback()
		s.logger.LogError("employee_offboarding", "Offboarding failed", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   tenantID,
		})
		return fmt.Errorf("offboarding execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.LogError("employee_offboarding", "Failed to commit offboarding", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   tenantID,
		})
		return fmt.Errorf("failed to commit offboarding: %w", err)
	}

	s.logger.Info("Employee offboarding completed successfully",
		"employee_id", req.EmployeeID,
		"tenant_id", tenantID,
	)

	return nil
}

// PromoteEmployee handles employee promotions
func (s *EmployeeLifecycleService) PromoteEmployee(ctx context.Context, tenantID uuid.UUID, req PromotionRequest) (*AssignmentResult, error) {
	s.logger.Info("Starting employee promotion",
		"employee_id", req.EmployeeID,
		"new_position_id", req.NewPositionID,
		"promotion_date", req.PromotionDate,
		"tenant_id", tenantID,
	)

	// Set defaults
	if req.FTEPercentage == 0 {
		req.FTEPercentage = 1.0
	}

	// Get current position for transfer
	emp, err := s.client.Employee.Query().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	if emp.CurrentPositionID == nil {
		return nil, fmt.Errorf("employee has no current position to promote from")
	}

	// Execute promotion as a transfer
	transferReq := TransferRequest{
		EmployeeID:      req.EmployeeID,
		FromPositionID:  *emp.CurrentPositionID,
		ToPositionID:    req.NewPositionID,
		TransferDate:    req.PromotionDate,
		TransferReason:  fmt.Sprintf("Promotion: %s", req.PromotionReason),
		ApprovedBy:      req.ApprovedBy,
		FTEPercentage:   req.FTEPercentage,
		WorkArrangement: req.WorkArrangement,
	}

	result, err := s.positionAssignmentSvc.TransferEmployee(ctx, tenantID, transferReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute promotion transfer: %w", err)
	}

	// Record promotion event
	err = s.recordLifecycleEvent(ctx, tenantID, LifecycleEvent{
		EmployeeID:  req.EmployeeID,
		TenantID:    tenantID,
		EventType:   "PROMOTION",
		EventDate:   req.PromotionDate,
		Description: fmt.Sprintf("Promoted to new position: %s", req.PromotionReason),
		Metadata: map[string]interface{}{
			"from_position_id": *emp.CurrentPositionID,
			"to_position_id":   req.NewPositionID,
			"salary_adjustment": req.SalaryAdjustment,
			"approved_by":      req.ApprovedBy,
		},
		CreatedBy: req.ApprovedBy,
		CreatedAt: time.Now(),
	})

	if err != nil {
		s.logger.LogError("promotion_event", "Failed to record promotion event", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   tenantID,
		})
		// Don't fail the promotion, just log the error
	}

	s.logger.Info("Employee promotion completed successfully",
		"employee_id", req.EmployeeID,
		"assignment_id", result.AssignmentID,
		"tenant_id", tenantID,
	)

	return result, nil
}

// ChangeEmploymentStatus handles employment status changes
func (s *EmployeeLifecycleService) ChangeEmploymentStatus(ctx context.Context, tenantID uuid.UUID, req StatusChangeRequest) error {
	s.logger.Info("Changing employee status",
		"employee_id", req.EmployeeID,
		"new_status", req.NewStatus,
		"effective_date", req.EffectiveDate,
		"tenant_id", tenantID,
	)

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	err = s.executeStatusChange(ctx, tx, tenantID, req)
	if err != nil {
		tx.Rollback()
		s.logger.LogError("status_change", "Status change failed", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"new_status":  req.NewStatus,
			"tenant_id":   tenantID,
		})
		return fmt.Errorf("status change execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.LogError("status_change", "Failed to commit status change", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   tenantID,
		})
		return fmt.Errorf("failed to commit status change: %w", err)
	}

	s.logger.Info("Employee status changed successfully",
		"employee_id", req.EmployeeID,
		"new_status", req.NewStatus,
		"tenant_id", tenantID,
	)

	return nil
}

// GetEmployeeLifecycleEvents retrieves lifecycle events for an employee
func (s *EmployeeLifecycleService) GetEmployeeLifecycleEvents(ctx context.Context, tenantID uuid.UUID, employeeID uuid.UUID) ([]LifecycleEvent, error) {
	// Note: This would need a dedicated lifecycle_events table in a real implementation
	// For now, we return a placeholder
	return []LifecycleEvent{}, nil
}

// Private helper methods

func (s *EmployeeLifecycleService) executeOnboarding(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req OnboardingRequest) (*OnboardingResult, error) {
	// 1. Validate employee details if provided
	var validatedDetails map[string]interface{}
	if req.EmployeeDetails != nil {
		detailsJSON, _ := json.Marshal(req.EmployeeDetails)
		details, err := types.EmployeeDetailsFactory(req.EmployeeType, detailsJSON)
		if err != nil {
			return nil, fmt.Errorf("invalid employee details: %w", err)
		}
		if err := details.Validate(); err != nil {
			return nil, fmt.Errorf("employee details validation failed: %w", err)
		}
		validatedDetails = req.EmployeeDetails
	}

	// 2. Create employee record
	empBuilder := tx.Employee.Create().
		SetTenantID(tenantID).
		SetEmployeeType(employee.EmployeeType(req.EmployeeType)).
		SetEmployeeNumber(req.EmployeeNumber).
		SetFirstName(req.FirstName).
		SetLastName(req.LastName).
		SetEmail(req.Email).
		SetEmploymentStatus(employee.EmploymentStatusPENDING_START).
		SetHireDate(req.HireDate)

	if req.PersonalEmail != nil {
		empBuilder = empBuilder.SetPersonalEmail(*req.PersonalEmail)
	}
	if req.PhoneNumber != nil {
		empBuilder = empBuilder.SetPhoneNumber(*req.PhoneNumber)
	}
	if validatedDetails != nil {
		empBuilder = empBuilder.SetEmployeeDetails(validatedDetails)
	}

	newEmployee, err := empBuilder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	var assignmentID *uuid.UUID
	var assignmentStartDate time.Time

	// 3. Handle initial position assignment if provided
	if req.InitialPositionID != nil {
		if req.AssignmentStartDate != nil {
			assignmentStartDate = *req.AssignmentStartDate
		} else {
			assignmentStartDate = req.HireDate
		}

		assignmentReq := AssignmentRequest{
			EmployeeID:       newEmployee.ID,
			PositionID:       *req.InitialPositionID,
			StartDate:        assignmentStartDate,
			AssignmentType:   "REGULAR",
			AssignmentReason: "Initial onboarding assignment",
			FTEPercentage:    req.FTEPercentage,
			ApprovedBy:       req.OnboardingManager,
			WorkArrangement:  req.WorkArrangement,
		}

		// Create position assignment within the same transaction
		assignment, err := s.executeAssignmentInTx(ctx, tx, tenantID, assignmentReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial position assignment: %w", err)
		}

		assignmentID = &assignment.ID

		// Update employee with current position
		_, err = tx.Employee.UpdateOne(newEmployee).
			SetCurrentPositionID(*req.InitialPositionID).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to update employee current position: %w", err)
		}
	}

	// 4. Record onboarding event (placeholder - would need lifecycle_events table)
	eventID := uuid.New()

	return &OnboardingResult{
		Success:            true,
		EmployeeID:         newEmployee.ID,
		EmployeeNumber:     newEmployee.EmployeeNumber,
		InitialAssignmentID: assignmentID,
		OnboardingEventID:  eventID,
		Message:            "Employee onboarded successfully",
	}, nil
}

func (s *EmployeeLifecycleService) executeOffboarding(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req OffboardingRequest) error {
	// 1. Fetch employee
	emp, err := tx.Employee.Query().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch employee: %w", err)
	}

	// 2. End active position assignments
	activeAssignments, err := tx.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(req.EmployeeID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		All(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch active assignments: %w", err)
	}

	endDate := req.TerminationDate
	if req.LastWorkingDate != nil {
		endDate = *req.LastWorkingDate
	}

	for _, assignment := range activeAssignments {
		_, err = tx.PositionOccupancyHistory.UpdateOne(assignment).
			SetEndDate(endDate).
			SetIsActive(false).
			SetAssignmentReason(fmt.Sprintf("Terminated: %s", req.TerminationReason)).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to end assignment %s: %w", assignment.ID, err)
		}

		// Update position status to OPEN
		_, err = tx.Position.Update().
			Where(
				position.IDEQ(assignment.PositionID),
				position.TenantIDEQ(tenantID),
			).
			SetStatus(position.StatusOPEN).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to update position status: %w", err)
		}
	}

	// 3. Update employee status and termination details
	_, err = tx.Employee.UpdateOne(emp).
		SetEmploymentStatus(employee.EmploymentStatusTERMINATED).
		SetTerminationDate(req.TerminationDate).
		ClearCurrentPositionID().
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to update employee termination: %w", err)
	}

	// 4. Record offboarding event (placeholder)
	// In a real implementation, this would create a lifecycle event record

	return nil
}

func (s *EmployeeLifecycleService) executeStatusChange(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req StatusChangeRequest) error {
	// Update employee status
	updateBuilder := tx.Employee.Update().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		SetEmploymentStatus(employee.EmploymentStatus(req.NewStatus))

	// Handle special status changes
	switch req.NewStatus {
	case "TERMINATED":
		updateBuilder = updateBuilder.SetTerminationDate(req.EffectiveDate)
		// End active assignments
		_, err := tx.PositionOccupancyHistory.Update().
			Where(
				positionoccupancyhistory.EmployeeIDEQ(req.EmployeeID),
				positionoccupancyhistory.TenantIDEQ(tenantID),
				positionoccupancyhistory.IsActiveEQ(true),
			).
			SetEndDate(req.EffectiveDate).
			SetIsActive(false).
			SetAssignmentReason(fmt.Sprintf("Status change: %s", req.Reason)).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to end active assignments: %w", err)
		}
	case "ON_LEAVE":
		// For leave, we might want to keep assignments active but mark them differently
		// This depends on business requirements
	}

	_, err := updateBuilder.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	return nil
}

func (s *EmployeeLifecycleService) executeAssignmentInTx(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req AssignmentRequest) (*ent.PositionOccupancyHistory, error) {
	// Create assignment within existing transaction
	assignment, err := tx.PositionOccupancyHistory.Create().
		SetTenantID(tenantID).
		SetPositionID(req.PositionID).
		SetEmployeeID(req.EmployeeID).
		SetStartDate(req.StartDate).
		SetIsActive(true).
		SetAssignmentType(positionoccupancyhistory.AssignmentType(req.AssignmentType)).
		SetAssignmentReason(req.AssignmentReason).
		SetFtePercentage(req.FTEPercentage).
		SetApprovedBy(req.ApprovedBy).
		SetApprovalDate(time.Now()).
		SetWorkArrangement(positionoccupancyhistory.WorkArrangement(req.WorkArrangement)).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// Update position status
	_, err = tx.Position.Update().
		Where(
			position.IDEQ(req.PositionID),
			position.TenantIDEQ(tenantID),
		).
		SetStatus(position.StatusFILLED).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update position status: %w", err)
	}

	return assignment, nil
}

func (s *EmployeeLifecycleService) recordLifecycleEvent(ctx context.Context, tenantID uuid.UUID, event LifecycleEvent) error {
	// Placeholder for lifecycle event recording
	// In a real implementation, this would insert into a lifecycle_events table
	s.logger.Info("Lifecycle event recorded",
		"employee_id", event.EmployeeID,
		"event_type", event.EventType,
		"event_date", event.EventDate,
		"tenant_id", tenantID,
	)
	return nil
}

import "encoding/json"