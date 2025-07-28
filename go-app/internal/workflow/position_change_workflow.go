// internal/workflow/position_change_workflow.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/google/uuid"
)

// PositionChangeRequest represents a position change workflow request
type PositionChangeRequest struct {
	TenantID      uuid.UUID                `json:"tenant_id"`
	EmployeeID    uuid.UUID                `json:"employee_id"`
	NewPosition   PositionChangeData       `json:"new_position"`
	EffectiveDate time.Time                `json:"effective_date"`
	ChangeReason  string                   `json:"change_reason"`
	RequestedBy   uuid.UUID                `json:"requested_by"`
	ApprovalData  *PositionApprovalData    `json:"approval_data,omitempty"`
}

// PositionChangeData represents the new position data
type PositionChangeData struct {
	PositionTitle       string     `json:"position_title"`
	Department          string     `json:"department"`
	JobLevel            *string    `json:"job_level,omitempty"`
	Location            *string    `json:"location,omitempty"`
	EmploymentType      string     `json:"employment_type"`
	ReportsToEmployeeID *uuid.UUID `json:"reports_to_employee_id,omitempty"`
	MinSalary           *float64   `json:"min_salary,omitempty"`
	MaxSalary           *float64   `json:"max_salary,omitempty"`
	Currency            *string    `json:"currency,omitempty"`
}

// PositionApprovalData represents approval workflow data
type PositionApprovalData struct {
	RequiresApproval bool                 `json:"requires_approval"`
	ApprovalSteps    []ApprovalStep       `json:"approval_steps,omitempty"`
	CurrentStep      int                  `json:"current_step"`
	ApprovalHistory  []ApprovalEvent      `json:"approval_history,omitempty"`
}

// ApprovalStep represents a single approval step
type ApprovalStep struct {
	StepID        string    `json:"step_id"`
	ApproverID    uuid.UUID `json:"approver_id"`
	ApproverTitle string    `json:"approver_title"`
	StepType      string    `json:"step_type"` // MANAGER, HR, EXECUTIVE
	IsRequired    bool      `json:"is_required"`
	Timeout       string    `json:"timeout"` // duration string like "72h"
}

// ApprovalEvent represents an approval action
type ApprovalEvent struct {
	StepID      string    `json:"step_id"`
	ApproverID  uuid.UUID `json:"approver_id"`
	Action      string    `json:"action"` // APPROVED, REJECTED, DELEGATED
	Comments    string    `json:"comments,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// PositionChangeResult represents the workflow result
type PositionChangeResult struct {
	Success           bool       `json:"success"`
	PositionHistoryID *uuid.UUID `json:"position_history_id,omitempty"`
	EffectiveDate     time.Time  `json:"effective_date"`
	IsRetroactive     bool       `json:"is_retroactive"`
	ProcessedAt       time.Time  `json:"processed_at"`
	Error             string     `json:"error,omitempty"`
	ApprovalStatus    string     `json:"approval_status,omitempty"`
}

// PositionChangeWorkflow orchestrates the complete position change process
func PositionChangeWorkflow(ctx workflow.Context, req PositionChangeRequest) (*PositionChangeResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting position change workflow", 
		"employee_id", req.EmployeeID,
		"tenant_id", req.TenantID,
		"effective_date", req.EffectiveDate,
		"position_title", req.NewPosition.PositionTitle)

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 2,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result PositionChangeResult
	result.EffectiveDate = req.EffectiveDate
	result.ProcessedAt = workflow.Now(ctx)
	result.IsRetroactive = req.EffectiveDate.Before(workflow.Now(ctx))

	// Stage 1: Validation and Risk Assessment
	logger.Info("Stage 1: Validation and risk assessment")
	
	var validationResult TemporalValidationResult
	err := workflow.ExecuteActivity(ctx,
		"ValidateTemporalConsistencyActivity",
		ValidateTemporalConsistencyRequest{
			TenantID:      req.TenantID,
			EmployeeID:    req.EmployeeID,
			EffectiveDate: req.EffectiveDate,
		}).Get(ctx, &validationResult)

	if err != nil || !validationResult.IsValid {
		result.Success = false
		result.Error = fmt.Sprintf("Temporal validation failed: %s", validationResult.ErrorMessage)
		logger.Error("Temporal validation failed", "error", result.Error)
		return &result, nil
	}

	// Risk assessment for approval requirements
	var riskResult RiskAssessmentResult
	err = workflow.ExecuteActivity(ctx,
		"AssessPositionChangeRiskActivity",
		RiskAssessmentRequest{
			TenantID:       req.TenantID,
			EmployeeID:     req.EmployeeID,
			CurrentPosition: nil, // Will be fetched in activity
			NewPosition:    req.NewPosition,
			EffectiveDate:  req.EffectiveDate,
			ChangeReason:   req.ChangeReason,
		}).Get(ctx, &riskResult)

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Risk assessment failed: %s", err.Error())
		logger.Error("Risk assessment failed", "error", err)
		return &result, nil
	}

	// Stage 2: Approval Process (if required)
	if riskResult.RequiresApproval {
		logger.Info("Stage 2: Approval process required", "risk_level", riskResult.RiskLevel)
		
		var approvalResult ApprovalWorkflowResult
		err = workflow.ExecuteChildWorkflow(
			workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				WorkflowID: fmt.Sprintf("approval-%s-%d", 
					req.EmployeeID.String(), 
					req.EffectiveDate.Unix()),
			}),
			"PositionChangeApprovalWorkflow",
			PositionApprovalRequest{
				TenantID:          req.TenantID,
				EmployeeID:        req.EmployeeID,
				PositionChangeReq: req,
				RiskAssessment:    riskResult,
				ApprovalSteps:     riskResult.RequiredApprovals,
			},
		).Get(ctx, &approvalResult)

		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Approval workflow failed: %s", err.Error())
			logger.Error("Approval workflow failed", "error", err)
			return &result, nil
		}

		if !approvalResult.Approved {
			result.Success = false
			result.Error = "Position change was rejected during approval process"
			result.ApprovalStatus = "REJECTED"
			logger.Info("Position change rejected", "reason", approvalResult.RejectionReason)
			return &result, nil
		}

		result.ApprovalStatus = "APPROVED"
		logger.Info("Position change approved")
	} else {
		result.ApprovalStatus = "NOT_REQUIRED"
		logger.Info("No approval required for this position change")
	}

	// Stage 3: Retroactive Processing (if needed)
	if result.IsRetroactive {
		logger.Info("Stage 3: Processing retroactive position change")
		
		var retroResult RetroactiveProcessingResult
		err = workflow.ExecuteActivity(ctx,
			"ProcessRetroactivePositionChangeActivity",
			ProcessRetroactiveRequest{
				TenantID:      req.TenantID,
				EmployeeID:    req.EmployeeID,
				EffectiveDate: req.EffectiveDate,
				NewPosition:   req.NewPosition,
			}).Get(ctx, &retroResult)

		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("Retroactive processing failed: %s", err.Error())
			logger.Error("Retroactive processing failed", "error", err)
			return &result, nil
		}

		// If retroactive change affects payroll, trigger recalculation
		if retroResult.RequiresRecalculation {
			logger.Info("Triggering payroll recalculation due to retroactive change")
			
			err = workflow.ExecuteChildWorkflow(
				workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
					WorkflowID: fmt.Sprintf("payroll-recalc-%s-%d", 
						req.EmployeeID.String(), 
						req.EffectiveDate.Unix()),
				}),
				"PayrollRecalculationWorkflow",
				PayrollRecalculationRequest{
					TenantID:      req.TenantID,
					EmployeeID:    req.EmployeeID,
					EffectiveDate: req.EffectiveDate,
					Reason:        "Position change retroactive adjustment",
				},
			).GetChildWorkflowExecution().Get(ctx, nil)

			if err != nil {
				logger.Warn("Failed to start payroll recalculation workflow", "error", err)
				// Don't fail the main workflow for this
			}
		}
	}

	// Stage 4: Create Position History Record
	logger.Info("Stage 4: Creating position history record")
	
	var historyResult CreatePositionHistoryResult
	err = workflow.ExecuteActivity(ctx,
		"CreatePositionHistoryActivity",
		CreatePositionHistoryRequest{
			TenantID:        req.TenantID,
			EmployeeID:      req.EmployeeID,
			PositionData:    req.NewPosition,
			EffectiveDate:   req.EffectiveDate,
			ChangeReason:    req.ChangeReason,
			CreatedBy:       req.RequestedBy,
			IsRetroactive:   result.IsRetroactive,
		}).Get(ctx, &historyResult)

	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Position history creation failed: %s", err.Error())
		logger.Error("Position history creation failed", "error", err)
		return &result, nil
	}

	result.PositionHistoryID = &historyResult.ID
	logger.Info("Position history record created", "position_history_id", historyResult.ID)

	// Stage 5: Update Neo4j Graph Data
	logger.Info("Stage 5: Updating graph data")
	
	err = workflow.ExecuteActivity(ctx,
		"PublishPositionChangeEventActivity",
		PublishEventRequest{
			EventType: "HR.Position.Changed",
			TenantID:  req.TenantID,
			Payload: PositionChangedPayload{
				EmployeeID:        req.EmployeeID,
				PositionHistoryID: historyResult.ID,
				EffectiveDate:     req.EffectiveDate,
				IsRetroactive:     result.IsRetroactive,
				OldPosition:       historyResult.PreviousPosition,
				NewPosition:       req.NewPosition,
			},
		}).Get(ctx, nil)

	if err != nil {
		logger.Error("Failed to publish position change event", "error", err)
		// Don't fail the main workflow for event publishing
	}

	// Stage 6: Notification and Integration
	logger.Info("Stage 6: Sending notifications")
	
	err = workflow.ExecuteActivity(ctx,
		"SendPositionChangeNotificationsActivity",
		NotificationRequest{
			TenantID:          req.TenantID,
			EmployeeID:        req.EmployeeID,
			PositionHistoryID: historyResult.ID,
			ChangeType:        "POSITION_CHANGE",
			EffectiveDate:     req.EffectiveDate,
			IsRetroactive:     result.IsRetroactive,
			NotifyEmployees:   []uuid.UUID{req.EmployeeID}, // Employee + manager + HR
		}).Get(ctx, nil)

	if err != nil {
		logger.Warn("Failed to send notifications", "error", err)
		// Don't fail the main workflow for notifications
	}

	// Success!
	result.Success = true
	logger.Info("Position change workflow completed successfully",
		"position_history_id", historyResult.ID,
		"effective_date", req.EffectiveDate,
		"is_retroactive", result.IsRetroactive)

	return &result, nil
}

// Activity request/response types

type ValidateTemporalConsistencyRequest struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EffectiveDate time.Time `json:"effective_date"`
}

type TemporalValidationResult struct {
	IsValid      bool   `json:"is_valid"`
	ErrorMessage string `json:"error_message,omitempty"`
}

type RiskAssessmentRequest struct {
	TenantID        uuid.UUID           `json:"tenant_id"`
	EmployeeID      uuid.UUID           `json:"employee_id"`
	CurrentPosition *PositionChangeData `json:"current_position,omitempty"`
	NewPosition     PositionChangeData  `json:"new_position"`
	EffectiveDate   time.Time           `json:"effective_date"`
	ChangeReason    string              `json:"change_reason"`
}

type RiskAssessmentResult struct {
	RiskLevel         string         `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RequiresApproval  bool           `json:"requires_approval"`
	RequiredApprovals []ApprovalStep `json:"required_approvals,omitempty"`
	RiskFactors       []string       `json:"risk_factors,omitempty"`
}

type ProcessRetroactiveRequest struct {
	TenantID      uuid.UUID          `json:"tenant_id"`
	EmployeeID    uuid.UUID          `json:"employee_id"`
	EffectiveDate time.Time          `json:"effective_date"`
	NewPosition   PositionChangeData `json:"new_position"`
}

type RetroactiveProcessingResult struct {
	RequiresRecalculation bool     `json:"requires_recalculation"`
	AffectedPeriods      []string `json:"affected_periods,omitempty"`
}

type CreatePositionHistoryRequest struct {
	TenantID        uuid.UUID          `json:"tenant_id"`
	EmployeeID      uuid.UUID          `json:"employee_id"`
	PositionData    PositionChangeData `json:"position_data"`
	EffectiveDate   time.Time          `json:"effective_date"`
	ChangeReason    string             `json:"change_reason"`
	CreatedBy       uuid.UUID          `json:"created_by"`
	IsRetroactive   bool               `json:"is_retroactive"`
}

type CreatePositionHistoryResult struct {
	ID               uuid.UUID           `json:"id"`
	PreviousPosition *PositionChangeData `json:"previous_position,omitempty"`
}

type PublishEventRequest struct {
	EventType string      `json:"event_type"`
	TenantID  uuid.UUID   `json:"tenant_id"`
	Payload   interface{} `json:"payload"`
}

type PositionChangedPayload struct {
	EmployeeID        uuid.UUID           `json:"employee_id"`
	PositionHistoryID uuid.UUID           `json:"position_history_id"`
	EffectiveDate     time.Time           `json:"effective_date"`
	IsRetroactive     bool                `json:"is_retroactive"`
	OldPosition       *PositionChangeData `json:"old_position,omitempty"`
	NewPosition       PositionChangeData  `json:"new_position"`
}

type NotificationRequest struct {
	TenantID          uuid.UUID   `json:"tenant_id"`
	EmployeeID        uuid.UUID   `json:"employee_id"`
	PositionHistoryID uuid.UUID   `json:"position_history_id"`
	ChangeType        string      `json:"change_type"`
	EffectiveDate     time.Time   `json:"effective_date"`
	IsRetroactive     bool        `json:"is_retroactive"`
	NotifyEmployees   []uuid.UUID `json:"notify_employees"`
}

type PayrollRecalculationRequest struct {
	TenantID      uuid.UUID `json:"tenant_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	EffectiveDate time.Time `json:"effective_date"`
	Reason        string    `json:"reason"`
}

// Approval workflow types

type PositionApprovalRequest struct {
	TenantID          uuid.UUID                `json:"tenant_id"`
	EmployeeID        uuid.UUID                `json:"employee_id"`
	PositionChangeReq PositionChangeRequest    `json:"position_change_req"`
	RiskAssessment    RiskAssessmentResult     `json:"risk_assessment"`
	ApprovalSteps     []ApprovalStep           `json:"approval_steps"`
}

type ApprovalWorkflowResult struct {
	Approved        bool   `json:"approved"`
	RejectionReason string `json:"rejection_reason,omitempty"`
	ApprovalHistory []ApprovalEvent `json:"approval_history"`
}