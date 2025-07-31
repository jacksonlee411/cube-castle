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
	"github.com/google/uuid"
)

// PositionAssignmentService handles complex position assignment operations
// Provides transaction-safe, intelligent position assignment with conflict resolution
type PositionAssignmentService struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// AssignmentRequest represents a position assignment request
type AssignmentRequest struct {
	EmployeeID       uuid.UUID  `json:"employee_id" validate:"required"`
	PositionID       uuid.UUID  `json:"position_id" validate:"required"`
	StartDate        time.Time  `json:"start_date" validate:"required"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	AssignmentType   string     `json:"assignment_type" validate:"required,oneof=REGULAR INTERIM ACTING TEMPORARY SECONDMENT"`
	AssignmentReason string     `json:"assignment_reason,omitempty"`
	FTEPercentage    float64    `json:"fte_percentage" validate:"gte=0.1,lte=1.0"`
	ApprovedBy       uuid.UUID  `json:"approved_by" validate:"required"`
	WorkArrangement  string     `json:"work_arrangement,omitempty" validate:"omitempty,oneof=ON_SITE REMOTE HYBRID"`
}

// TransferRequest represents an employee transfer between positions
type TransferRequest struct {
	EmployeeID      uuid.UUID `json:"employee_id" validate:"required"`
	FromPositionID  uuid.UUID `json:"from_position_id" validate:"required"`
	ToPositionID    uuid.UUID `json:"to_position_id" validate:"required"`
	TransferDate    time.Time `json:"transfer_date" validate:"required"`
	TransferReason  string    `json:"transfer_reason" validate:"required"`
	ApprovedBy      uuid.UUID `json:"approved_by" validate:"required"`
	FTEPercentage   float64   `json:"fte_percentage" validate:"gte=0.1,lte=1.0"`
	WorkArrangement string    `json:"work_arrangement,omitempty" validate:"omitempty,oneof=ON_SITE REMOTE HYBRID"`
}

// AssignmentResult contains the result of a position assignment operation
type AssignmentResult struct {
	Success              bool      `json:"success"`
	AssignmentID         uuid.UUID `json:"assignment_id,omitempty"`
	EmployeeID           uuid.UUID `json:"employee_id"`
	PositionID           uuid.UUID `json:"position_id"`
	StartDate            time.Time `json:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	AssignmentType       string    `json:"assignment_type"`
	PreviousAssignmentID *uuid.UUID `json:"previous_assignment_id,omitempty"`
	Message              string    `json:"message,omitempty"`
}

// ConflictInfo represents information about assignment conflicts
type ConflictInfo struct {
	Type            string                `json:"type"`
	ConflictingItem interface{}           `json:"conflicting_item"`
	Resolution      string                `json:"resolution"`
	Details         map[string]interface{} `json:"details"`
}

// NewPositionAssignmentService creates a new PositionAssignmentService
func NewPositionAssignmentService(client *ent.Client, logger *logging.StructuredLogger) *PositionAssignmentService {
	return &PositionAssignmentService{
		client: client,
		logger: logger,
	}
}

// AssignPosition assigns an employee to a position with intelligent conflict resolution
func (s *PositionAssignmentService) AssignPosition(ctx context.Context, tenantID uuid.UUID, req AssignmentRequest) (*AssignmentResult, error) {
	s.logger.Info("Starting position assignment",
		"employee_id", req.EmployeeID,
		"position_id", req.PositionID,
		"assignment_type", req.AssignmentType,
		"tenant_id", tenantID,
	)

	// Set default values
	if req.FTEPercentage == 0 {
		req.FTEPercentage = 1.0
	}
	if req.WorkArrangement == "" {
		req.WorkArrangement = "ON_SITE"
	}

	// Pre-assignment validation and conflict detection
	conflicts, err := s.detectAssignmentConflicts(ctx, tenantID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to detect conflicts: %w", err)
	}

	// Execute assignment in transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	result, err := s.executeAssignment(ctx, tx, tenantID, req, conflicts)
	if err != nil {
		tx.Rollback()
		s.logger.LogError("position_assignment", "Assignment failed", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"position_id": req.PositionID,
			"tenant_id":   tenantID,
		})
		return nil, fmt.Errorf("assignment execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.LogError("position_assignment", "Failed to commit assignment", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"position_id": req.PositionID,
			"tenant_id":   tenantID,
		})
		return nil, fmt.Errorf("failed to commit assignment: %w", err)
	}

	s.logger.Info("Position assignment completed successfully",
		"assignment_id", result.AssignmentID,
		"employee_id", req.EmployeeID,
		"position_id", req.PositionID,
		"assignment_type", req.AssignmentType,
		"tenant_id", tenantID,
	)

	return result, nil
}

// TransferEmployee handles employee transfers between positions
func (s *PositionAssignmentService) TransferEmployee(ctx context.Context, tenantID uuid.UUID, req TransferRequest) (*AssignmentResult, error) {
	s.logger.Info("Starting employee transfer",
		"employee_id", req.EmployeeID,
		"from_position_id", req.FromPositionID,
		"to_position_id", req.ToPositionID,
		"tenant_id", tenantID,
	)

	// Set default FTE if not provided
	if req.FTEPercentage == 0 {
		req.FTEPercentage = 1.0
	}

	// Execute transfer in transaction
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	result, err := s.executeTransfer(ctx, tx, tenantID, req)
	if err != nil {
		tx.Rollback()
		s.logger.LogError("employee_transfer", "Transfer failed", err, map[string]interface{}{
			"employee_id":      req.EmployeeID,
			"from_position_id": req.FromPositionID,
			"to_position_id":   req.ToPositionID,
			"tenant_id":        tenantID,
		})
		return nil, fmt.Errorf("transfer execution failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.logger.LogError("employee_transfer", "Failed to commit transfer", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   tenantID,
		})
		return nil, fmt.Errorf("failed to commit transfer: %w", err)
	}

	s.logger.Info("Employee transfer completed successfully",
		"assignment_id", result.AssignmentID,
		"employee_id", req.EmployeeID,
		"from_position_id", req.FromPositionID,
		"to_position_id", req.ToPositionID,
		"tenant_id", tenantID,
	)

	return result, nil
}

// EndAssignment ends an active position assignment
func (s *PositionAssignmentService) EndAssignment(ctx context.Context, tenantID uuid.UUID, employeeID uuid.UUID, endDate time.Time, reason string) error {
	s.logger.Info("Ending position assignment",
		"employee_id", employeeID,
		"end_date", endDate,
		"reason", reason,
		"tenant_id", tenantID,
	)

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Find active assignment
	activeAssignment, err := tx.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(employeeID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		Only(ctx)

	if err != nil {
		tx.Rollback()
		if ent.IsNotFound(err) {
			return fmt.Errorf("no active assignment found for employee")
		}
		return fmt.Errorf("failed to find active assignment: %w", err)
	}

	// End the assignment
	_, err = tx.PositionOccupancyHistory.UpdateOne(activeAssignment).
		SetEndDate(endDate).
		SetIsActive(false).
		SetAssignmentReason(reason).
		Save(ctx)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to end assignment: %w", err)
	}

	// Clear current position from employee
	_, err = tx.Employee.Update().
		Where(
			employee.IDEQ(employeeID),
			employee.TenantIDEQ(tenantID),
		).
		ClearCurrentPositionID().
		Save(ctx)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear employee current position: %w", err)
	}

	// Update position status to OPEN if no other active assignments
	otherAssignments, err := tx.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.PositionIDEQ(activeAssignment.PositionID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		Count(ctx)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check other assignments: %w", err)
	}

	if otherAssignments == 0 {
		_, err = tx.Position.Update().
			Where(
				position.IDEQ(activeAssignment.PositionID),
				position.TenantIDEQ(tenantID),
			).
			SetStatus(position.StatusOPEN).
			Save(ctx)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update position status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit assignment end: %w", err)
	}

	s.logger.Info("Position assignment ended successfully",
		"employee_id", employeeID,
		"assignment_id", activeAssignment.ID,
		"tenant_id", tenantID,
	)

	return nil
}

// GetActiveAssignments retrieves all active position assignments for a tenant
func (s *PositionAssignmentService) GetActiveAssignments(ctx context.Context, tenantID uuid.UUID) ([]*ent.PositionOccupancyHistory, error) {
	assignments, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		WithEmployee().
		WithPosition().
		Order(positionoccupancyhistory.ByStartDate()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch active assignments: %w", err)
	}

	return assignments, nil
}

// GetEmployeeAssignmentHistory retrieves assignment history for an employee
func (s *PositionAssignmentService) GetEmployeeAssignmentHistory(ctx context.Context, tenantID uuid.UUID, employeeID uuid.UUID) ([]*ent.PositionOccupancyHistory, error) {
	assignments, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(employeeID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
		).
		WithPosition().
		Order(positionoccupancyhistory.ByStartDate()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee assignment history: %w", err)
	}

	return assignments, nil
}

// Private helper methods

func (s *PositionAssignmentService) detectAssignmentConflicts(ctx context.Context, tenantID uuid.UUID, req AssignmentRequest) ([]ConflictInfo, error) {
	var conflicts []ConflictInfo

	// 1. Check if employee exists and is active
	emp, err := s.client.Employee.Query().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	if emp.EmploymentStatus != employee.EmploymentStatusACTIVE {
		return nil, fmt.Errorf("employee is not active (status: %s)", emp.EmploymentStatus)
	}

	// 2. Check if position exists and is available
	pos, err := s.client.Position.Query().
		Where(
			position.IDEQ(req.PositionID),
			position.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("position not found")
		}
		return nil, fmt.Errorf("failed to fetch position: %w", err)
	}

	// 3. Check for existing active assignment for employee
	activeAssignment, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(req.EmployeeID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		First(ctx)

	if err == nil && activeAssignment != nil {
		conflicts = append(conflicts, ConflictInfo{
			Type:            "EXISTING_ASSIGNMENT",
			ConflictingItem: activeAssignment,
			Resolution:      "END_EXISTING_ASSIGNMENT",
			Details: map[string]interface{}{
				"current_position_id": activeAssignment.PositionID,
				"start_date":          activeAssignment.StartDate,
			},
		})
	}

	// 4. Check for position capacity conflicts
	if pos.Status == position.StatusFILLED && req.AssignmentType == "REGULAR" {
		conflicts = append(conflicts, ConflictInfo{
			Type:            "POSITION_FILLED",
			ConflictingItem: pos,
			Resolution:      "ALLOW_MULTIPLE_ASSIGNMENTS",
			Details: map[string]interface{}{
				"position_status": pos.Status,
				"assignment_type": req.AssignmentType,
			},
		})
	}

	return conflicts, nil
}

func (s *PositionAssignmentService) executeAssignment(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req AssignmentRequest, conflicts []ConflictInfo) (*AssignmentResult, error) {
	var previousAssignmentID *uuid.UUID

	// Handle conflicts
	for _, conflict := range conflicts {
		switch conflict.Type {
		case "EXISTING_ASSIGNMENT":
			// End existing assignment
			if activeAssignment, ok := conflict.ConflictingItem.(*ent.PositionOccupancyHistory); ok {
				_, err := tx.PositionOccupancyHistory.UpdateOne(activeAssignment).
					SetEndDate(req.StartDate).
					SetIsActive(false).
					SetAssignmentReason("Ended for new assignment").
					Save(ctx)

				if err != nil {
					return nil, fmt.Errorf("failed to end existing assignment: %w", err)
				}

				previousAssignmentID = &activeAssignment.ID
			}
		}
	}

	// Create new assignment
	newAssignment, err := tx.PositionOccupancyHistory.Create().
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

	if req.EndDate != nil {
		newAssignment, err = tx.PositionOccupancyHistory.UpdateOne(newAssignment).
			SetEndDate(*req.EndDate).
			Save(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to set assignment end date: %w", err)
		}
	}

	// Update employee current position
	_, err = tx.Employee.Update().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		SetCurrentPositionID(req.PositionID).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update employee current position: %w", err)
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

	return &AssignmentResult{
		Success:              true,
		AssignmentID:         newAssignment.ID,
		EmployeeID:           req.EmployeeID,
		PositionID:           req.PositionID,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		AssignmentType:       req.AssignmentType,
		PreviousAssignmentID: previousAssignmentID,
		Message:              "Position assigned successfully",
	}, nil
}

func (s *PositionAssignmentService) executeTransfer(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, req TransferRequest) (*AssignmentResult, error) {
	// End current assignment
	currentAssignment, err := tx.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(req.EmployeeID),
			positionoccupancyhistory.PositionIDEQ(req.FromPositionID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.IsActiveEQ(true),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("current assignment not found")
		}
		return nil, fmt.Errorf("failed to find current assignment: %w", err)
	}

	// End current assignment
	_, err = tx.PositionOccupancyHistory.UpdateOne(currentAssignment).
		SetEndDate(req.TransferDate).
		SetIsActive(false).
		SetAssignmentReason("Transfer to new position").
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to end current assignment: %w", err)
	}

	// Create new assignment
	newAssignment, err := tx.PositionOccupancyHistory.Create().
		SetTenantID(tenantID).
		SetPositionID(req.ToPositionID).
		SetEmployeeID(req.EmployeeID).
		SetStartDate(req.TransferDate).
		SetIsActive(true).
		SetAssignmentType(positionoccupancyhistory.AssignmentTypeREGULAR).
		SetAssignmentReason(req.TransferReason).
		SetFtePercentage(req.FTEPercentage).
		SetApprovedBy(req.ApprovedBy).
		SetApprovalDate(time.Now()).
		SetWorkArrangement(positionoccupancyhistory.WorkArrangement(req.WorkArrangement)).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create new assignment: %w", err)
	}

	// Update employee current position
	_, err = tx.Employee.Update().
		Where(
			employee.IDEQ(req.EmployeeID),
			employee.TenantIDEQ(tenantID),
		).
		SetCurrentPositionID(req.ToPositionID).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update employee current position: %w", err)
	}

	// Update position statuses
	_, err = tx.Position.Update().
		Where(
			position.IDEQ(req.FromPositionID),
			position.TenantIDEQ(tenantID),
		).
		SetStatus(position.StatusOPEN).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update from position status: %w", err)
	}

	_, err = tx.Position.Update().
		Where(
			position.IDEQ(req.ToPositionID),
			position.TenantIDEQ(tenantID),
		).
		SetStatus(position.StatusFILLED).
		Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to update to position status: %w", err)
	}

	return &AssignmentResult{
		Success:              true,
		AssignmentID:         newAssignment.ID,
		EmployeeID:           req.EmployeeID,
		PositionID:           req.ToPositionID,
		StartDate:            req.TransferDate,
		AssignmentType:       "REGULAR",
		PreviousAssignmentID: &currentAssignment.ID,
		Message:              "Employee transferred successfully",
	}, nil
}