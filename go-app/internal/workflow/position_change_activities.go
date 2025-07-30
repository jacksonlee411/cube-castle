// internal/workflow/position_change_activities.go
package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/ent"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// PositionChangeActivities implements activities for position change workflows
type PositionChangeActivities struct {
	entClient           *ent.Client
	temporalQuerySvc    *service.TemporalQueryService
	logger              *logging.StructuredLogger
}

// NewPositionChangeActivities creates new position change activities
func NewPositionChangeActivities(
	entClient *ent.Client,
	temporalQuerySvc *service.TemporalQueryService,
	logger *logging.StructuredLogger,
) *PositionChangeActivities {
	return &PositionChangeActivities{
		entClient:        entClient,
		temporalQuerySvc: temporalQuerySvc,
		logger:           logger,
	}
}

// ValidateTemporalConsistencyActivity validates temporal consistency for position changes
func (a *PositionChangeActivities) ValidateTemporalConsistencyActivity(
	ctx context.Context,
	req ValidateTemporalConsistencyRequest,
) (*TemporalValidationResult, error) {
	a.logger.LogInfo(ctx, "Validating temporal consistency", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"effective_date": req.EffectiveDate,
	})

	err := a.temporalQuerySvc.ValidateTemporalConsistency(
		ctx,
		req.TenantID,
		req.EmployeeID,
		req.EffectiveDate,
	)

	if err != nil {
		a.logger.LogWarning(ctx, "Temporal consistency validation failed", map[string]interface{}{
			"tenant_id":      req.TenantID,
			"employee_id":    req.EmployeeID,
			"effective_date": req.EffectiveDate,
			"error":          err.Error(),
		})
		return &TemporalValidationResult{
			IsValid:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	a.logger.LogInfo(ctx, "Temporal consistency validation passed", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"effective_date": req.EffectiveDate,
	})

	return &TemporalValidationResult{
		IsValid: true,
	}, nil
}

// AssessPositionChangeRiskActivity assesses risk level for position changes
func (a *PositionChangeActivities) AssessPositionChangeRiskActivity(
	ctx context.Context,
	req RiskAssessmentRequest,
) (*RiskAssessmentResult, error) {
	a.logger.LogInfo(ctx, "Assessing position change risk", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"effective_date": req.EffectiveDate,
		"new_position":   req.NewPosition.PositionTitle,
	})

	// Get current position for comparison
	currentPosition, err := a.temporalQuerySvc.GetCurrentPosition(ctx, req.TenantID, req.EmployeeID)
	if err != nil {
		a.logger.LogWarning(ctx, "Could not retrieve current position for risk assessment", map[string]interface{}{
			"tenant_id":   req.TenantID,
			"employee_id": req.EmployeeID,
			"error":       err.Error(),
		})
		// Continue with assessment even if current position is not found (new employee)
	}

	riskFactors := []string{}
	riskLevel := "LOW"
	requiresApproval := false
	var approvalSteps []ApprovalStep

	// Risk Factor 1: Salary increase
	if currentPosition != nil && currentPosition.MaxSalary != nil && req.NewPosition.MaxSalary != nil {
		salaryIncrease := (*req.NewPosition.MaxSalary - *currentPosition.MaxSalary) / *currentPosition.MaxSalary
		if salaryIncrease > 0.20 { // 20% increase
			riskFactors = append(riskFactors, "Significant salary increase (>20%)")
			riskLevel = "HIGH"
			requiresApproval = true
		} else if salaryIncrease > 0.10 { // 10% increase
			riskFactors = append(riskFactors, "Moderate salary increase (>10%)")
			if riskLevel == "LOW" {
				riskLevel = "MEDIUM"
			}
			requiresApproval = true
		}
	}

	// Risk Factor 2: Department change
	if currentPosition != nil && currentPosition.Department != req.NewPosition.Department {
		riskFactors = append(riskFactors, "Department change")
		if riskLevel == "LOW" {
			riskLevel = "MEDIUM"
		}
		requiresApproval = true
	}

	// Risk Factor 3: Executive level position
	if req.NewPosition.JobLevel != nil {
		jobLevel := *req.NewPosition.JobLevel
		if jobLevel == "C-LEVEL" || jobLevel == "VP" || jobLevel == "SVP" {
			riskFactors = append(riskFactors, "Executive level position")
			riskLevel = "CRITICAL"
			requiresApproval = true
		} else if jobLevel == "DIRECTOR" || jobLevel == "SENIOR_DIRECTOR" {
			riskFactors = append(riskFactors, "Director level position")
			if riskLevel != "CRITICAL" {
				riskLevel = "HIGH"
			}
			requiresApproval = true
		}
	}

	// Risk Factor 4: Retroactive change
	if req.EffectiveDate.Before(time.Now().Add(-30 * 24 * time.Hour)) { // More than 30 days ago
		riskFactors = append(riskFactors, "Retroactive change (>30 days)")
		if riskLevel == "LOW" {
			riskLevel = "MEDIUM"
		}
		requiresApproval = true
	}

	// Risk Factor 5: Frequent position changes
	// Check if employee has had multiple position changes in last 6 months
	sixMonthsAgo := time.Now().Add(-6 * 30 * 24 * time.Hour)
	recentChanges, err := a.temporalQuerySvc.GetPositionTimeline(
		ctx, req.TenantID, req.EmployeeID, &sixMonthsAgo, nil)
	if err == nil && len(recentChanges) > 2 {
		riskFactors = append(riskFactors, "Frequent position changes (>2 in 6 months)")
		if riskLevel == "LOW" {
			riskLevel = "MEDIUM"
		}
		requiresApproval = true
	}

	// Define approval steps based on risk level
	if requiresApproval {
		switch riskLevel {
		case "CRITICAL":
			approvalSteps = append(approvalSteps,
				ApprovalStep{
					StepID:        "hr-director",
					ApproverID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"), // HR Director
					ApproverTitle: "HR Director",
					StepType:      "HR",
					IsRequired:    true,
					Timeout:       "72h",
				},
				ApprovalStep{
					StepID:        "ceo",
					ApproverID:    uuid.MustParse("00000000-0000-0000-0000-000000000002"), // CEO
					ApproverTitle: "CEO",
					StepType:      "EXECUTIVE",
					IsRequired:    true,
					Timeout:       "168h", // 1 week
				},
			)
		case "HIGH":
			approvalSteps = append(approvalSteps,
				ApprovalStep{
					StepID:        "hr-manager",
					ApproverID:    uuid.MustParse("00000000-0000-0000-0000-000000000003"), // HR Manager
					ApproverTitle: "HR Manager",
					StepType:      "HR",
					IsRequired:    true,
					Timeout:       "48h",
				},
				ApprovalStep{
					StepID:        "hr-director",
					ApproverID:    uuid.MustParse("00000000-0000-0000-0000-000000000001"), // HR Director
					ApproverTitle: "HR Director",
					StepType:      "HR",
					IsRequired:    true,
					Timeout:       "72h",
				},
			)
		case "MEDIUM":
			approvalSteps = append(approvalSteps,
				ApprovalStep{
					StepID:        "direct-manager",
					ApproverID:    uuid.Nil, // Will be determined at runtime
					ApproverTitle: "Direct Manager",
					StepType:      "MANAGER",
					IsRequired:    true,
					Timeout:       "24h",
				},
				ApprovalStep{
					StepID:        "hr-manager",
					ApproverID:    uuid.MustParse("00000000-0000-0000-0000-000000000003"), // HR Manager
					ApproverTitle: "HR Manager",
					StepType:      "HR",
					IsRequired:    true,
					Timeout:       "48h",
				},
			)
		}
	}

	result := &RiskAssessmentResult{
		RiskLevel:         riskLevel,
		RequiresApproval:  requiresApproval,
		RequiredApprovals: approvalSteps,
		RiskFactors:       riskFactors,
	}

	a.logger.LogInfo(ctx, "Position change risk assessment completed", map[string]interface{}{
		"tenant_id":         req.TenantID,
		"employee_id":       req.EmployeeID,
		"risk_level":        riskLevel,
		"requires_approval": requiresApproval,
		"risk_factors":      riskFactors,
		"approval_steps":    len(approvalSteps),
	})

	return result, nil
}

// ProcessRetroactivePositionChangeActivity handles retroactive position changes
func (a *PositionChangeActivities) ProcessRetroactivePositionChangeActivity(
	ctx context.Context,
	req ProcessRetroactiveRequest,
) (*RetroactiveProcessingResult, error) {
	a.logger.LogInfo(ctx, "Processing retroactive position change", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"effective_date": req.EffectiveDate,
	})

	result := &RetroactiveProcessingResult{
		RequiresRecalculation: false,
		AffectedPeriods:       []string{},
	}

	// Determine if payroll recalculation is needed
	// Check if effective date affects closed payroll periods
	now := time.Now()
	monthsDiff := int(now.Sub(req.EffectiveDate).Hours() / (24 * 30))

	if monthsDiff > 0 {
		result.RequiresRecalculation = true
		
		// List affected payroll periods
		for i := 0; i < monthsDiff && i < 12; i++ { // Limit to 12 months
			periodStart := req.EffectiveDate.AddDate(0, i, 0)
			periodEnd := periodStart.AddDate(0, 1, -1)
			result.AffectedPeriods = append(result.AffectedPeriods, 
				fmt.Sprintf("%s to %s", 
					periodStart.Format("2006-01-02"), 
					periodEnd.Format("2006-01-02")))
		}
	}

	// Check if salary change affects retroactive periods
	if req.NewPosition.MinSalary != nil || req.NewPosition.MaxSalary != nil {
		result.RequiresRecalculation = true
	}

	a.logger.LogInfo(ctx, "Retroactive processing completed", map[string]interface{}{
		"tenant_id":             req.TenantID,
		"employee_id":           req.EmployeeID,
		"requires_recalculation": result.RequiresRecalculation,
		"affected_periods":      len(result.AffectedPeriods),
	})

	return result, nil
}

// CreatePositionHistoryActivity creates the position history record
func (a *PositionChangeActivities) CreatePositionHistoryActivity(
	ctx context.Context,
	req CreatePositionHistoryRequest,
) (*CreatePositionHistoryResult, error) {
	a.logger.LogInfo(ctx, "Creating position history record", map[string]interface{}{
		"tenant_id":      req.TenantID,
		"employee_id":    req.EmployeeID,
		"effective_date": req.EffectiveDate,
		"position_title": req.PositionData.PositionTitle,
	})

	// Get previous position for result
	var previousPosition *PositionChangeData
	currentPos, err := a.temporalQuerySvc.GetCurrentPosition(ctx, req.TenantID, req.EmployeeID)
	if err == nil {
		previousPosition = &PositionChangeData{
			PositionTitle:       currentPos.PositionTitle,
			Department:          currentPos.Department,
			JobLevel:            currentPos.JobLevel,
			Location:            currentPos.Location,
			EmploymentType:      currentPos.EmploymentType,
			ReportsToEmployeeID: currentPos.ReportsToEmployeeID,
			MinSalary:           currentPos.MinSalary,
			MaxSalary:           currentPos.MaxSalary,
			Currency:            currentPos.Currency,
		}
	}

	// Create position snapshot data
	snapshotData := service.PositionSnapshotData{
		PositionTitle:       req.PositionData.PositionTitle,
		Department:          req.PositionData.Department,
		JobLevel:            req.PositionData.JobLevel,
		Location:            req.PositionData.Location,
		EmploymentType:      req.PositionData.EmploymentType,
		ReportsToEmployeeID: req.PositionData.ReportsToEmployeeID,
		EffectiveDate:       req.EffectiveDate,
		EndDate:             nil, // New current position
		ChangeReason:        &req.ChangeReason,
		IsRetroactive:       req.IsRetroactive,
		MinSalary:           req.PositionData.MinSalary,
		MaxSalary:           req.PositionData.MaxSalary,
		Currency:            req.PositionData.Currency,
	}

	// Create the position snapshot
	snapshot, err := a.temporalQuerySvc.CreatePositionSnapshot(
		ctx,
		req.TenantID,
		req.EmployeeID,
		req.CreatedBy,
		snapshotData,
	)

	if err != nil {
		a.logger.LogError(ctx, "Failed to create position history record", err, map[string]interface{}{
			"tenant_id":      req.TenantID,
			"employee_id":    req.EmployeeID,
			"effective_date": req.EffectiveDate,
		})
		return nil, err
	}

	result := &CreatePositionHistoryResult{
		ID:               snapshot.PositionHistoryID,
		PreviousPosition: previousPosition,
	}

	a.logger.LogInfo(ctx, "Position history record created successfully", map[string]interface{}{
		"tenant_id":           req.TenantID,
		"employee_id":         req.EmployeeID,
		"position_history_id": snapshot.PositionHistoryID,
		"effective_date":      req.EffectiveDate,
	})

	return result, nil
}

// PublishPositionChangeEventActivity publishes position change events
func (a *PositionChangeActivities) PublishPositionChangeEventActivity(
	ctx context.Context,
	req PublishEventRequest,
) error {
	a.logger.LogInfo(ctx, "Publishing position change event", map[string]interface{}{
		"event_type": req.EventType,
		"tenant_id":  req.TenantID,
	})

	// TODO: Implement actual event publishing to message queue/event bus
	// For now, just log the event
	a.logger.LogInfo(ctx, "Position change event published", map[string]interface{}{
		"event_type": req.EventType,
		"tenant_id":  req.TenantID,
		"payload":    req.Payload,
	})

	return nil
}

// SendPositionChangeNotificationsActivity sends notifications for position changes
func (a *PositionChangeActivities) SendPositionChangeNotificationsActivity(
	ctx context.Context,
	req NotificationRequest,
) error {
	a.logger.LogInfo(ctx, "Sending position change notifications", map[string]interface{}{
		"tenant_id":          req.TenantID,
		"employee_id":        req.EmployeeID,
		"position_history_id": req.PositionHistoryID,
		"change_type":        req.ChangeType,
		"notify_count":       len(req.NotifyEmployees),
	})

	// TODO: Implement actual notification sending (email, Slack, etc.)
	// For now, just log the notifications
	for _, employeeID := range req.NotifyEmployees {
		a.logger.LogInfo(ctx, "Notification sent", map[string]interface{}{
			"tenant_id":          req.TenantID,
			"recipient_id":       employeeID,
			"position_history_id": req.PositionHistoryID,
			"change_type":        req.ChangeType,
		})
	}

	return nil
}