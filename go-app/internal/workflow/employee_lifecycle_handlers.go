// internal/workflow/employee_lifecycle_handlers.go
package workflow

import (
	"fmt"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// handleCreateCandidate 处理创建候选人
func handleCreateCandidate(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Creating candidate", "employee_id", req.EmployeeID)

	// 检查暂停状态
	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "candidate_creation"
	status.Progress = 0.15

	// 创建候选人
	var candidateResult CandidateCreationResult
	err := workflow.ExecuteActivity(ctx, "CreateCandidateActivity", CandidateCreationRequest{
		TenantID:      req.TenantID,
		CandidateData: req.OperationData,
		CreatedBy:     req.RequestedBy,
	}).Get(ctx, &candidateResult)

	if err != nil {
		return result, fmt.Errorf("failed to create candidate: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "candidate_created")
	result.ResultData = candidateResult
	status.Progress = 1.0

	logger.Info("Candidate created successfully", "candidate_id", candidateResult.CandidateID)
	return result, nil
}

// handleUpdateCandidate 处理更新候选人
func handleUpdateCandidate(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Updating candidate information", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "candidate_update"
	status.Progress = 0.5

	// 更新候选人信息
	err := workflow.ExecuteActivity(ctx, "UpdateCandidateActivity", InformationUpdateRequest{
		TenantID:   req.TenantID,
		EmployeeID: req.EmployeeID,
		UpdateType: "CANDIDATE_INFO",
		UpdateData: req.OperationData,
		UpdatedBy:  req.RequestedBy,
	}).Get(ctx, nil)

	if err != nil {
		return result, fmt.Errorf("failed to update candidate: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "candidate_updated")
	status.Progress = 1.0

	logger.Info("Candidate updated successfully", "employee_id", req.EmployeeID)
	return result, nil
}

// handleApproveHire 处理招聘审批
func handleApproveHire(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing hire approval", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "hire_approval"
	status.Progress = 0.3

	// 执行招聘审批子工作流
	var approvalResult ApprovalWorkflowResult
	err := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("hire-approval-%s-%d",
				req.EmployeeID.String(),
				workflow.Now(ctx).Unix()),
		}),
		"HireApprovalWorkflow",
		HireApprovalRequest{
			TenantID:    req.TenantID,
			EmployeeID:  req.EmployeeID,
			HireData:    req.OperationData,
			RequestedBy: req.RequestedBy,
		},
	).Get(ctx, &approvalResult)

	if err != nil {
		return result, fmt.Errorf("hire approval workflow failed: %w", err)
	}

	if !approvalResult.Approved {
		result.Status = "rejected"
		result.Error = "Hire was not approved"
		return result, nil
	}

	result.CompletedSteps = append(result.CompletedSteps, "hire_approved")
	status.Progress = 1.0

	logger.Info("Hire approval completed", "employee_id", req.EmployeeID, "approved", approvalResult.Approved)
	return result, nil
}

// handleStartOnboarding 处理开始入职
func handleStartOnboarding(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting onboarding process", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "onboarding_initialization"
	status.Progress = 0.2

	// 初始化入职流程
	var initResult OnboardingInitializationResult
	err := workflow.ExecuteActivity(ctx, "InitializeOnboardingActivity", OnboardingInitializationRequest{
		TenantID:    req.TenantID,
		EmployeeID:  req.EmployeeID,
		InitiatedBy: req.RequestedBy,
	}).Get(ctx, &initResult)

	if err != nil {
		return result, fmt.Errorf("failed to initialize onboarding: %w", err)
	}

	status.CurrentStep = "employee_creation"
	status.Progress = 0.4

	// 使用现有的员工入职工作流
	var onboardingResult EmployeeOnboardingResult
	err = workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("onboarding-%s-%d",
				req.EmployeeID.String(),
				workflow.Now(ctx).Unix()),
		}),
		"EmployeeOnboardingWorkflow",
		EmployeeOnboardingRequest{
			EmployeeID: req.EmployeeID,
			TenantID:   req.TenantID,
			// 从 req.OperationData 中提取必要信息
		},
	).Get(ctx, &onboardingResult)

	if err != nil {
		return result, fmt.Errorf("employee onboarding workflow failed: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "onboarding_started", "employee_created")
	result.ResultData = onboardingResult
	status.Progress = 1.0

	logger.Info("Onboarding process started successfully",
		"employee_id", req.EmployeeID,
		"onboarding_id", initResult.OnboardingID)
	return result, nil
}

// handleCompleteOnboardingStep 处理完成入职步骤
func handleCompleteOnboardingStep(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	stepID := req.OperationData["step_id"].(string)
	logger.Info("Completing onboarding step", "employee_id", req.EmployeeID, "step_id", stepID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = fmt.Sprintf("completing_step_%s", stepID)
	status.Progress = 0.6

	// 完成入职步骤
	err := workflow.ExecuteActivity(ctx, "CompleteOnboardingStepActivity", req.OperationData).Get(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to complete onboarding step %s: %w", stepID, err)
	}

	result.CompletedSteps = append(result.CompletedSteps, fmt.Sprintf("step_%s_completed", stepID))
	status.Progress = 1.0

	logger.Info("Onboarding step completed", "employee_id", req.EmployeeID, "step_id", stepID)
	return result, nil
}

// handleFinalizeOnboarding 处理完成入职
func handleFinalizeOnboarding(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Finalizing onboarding process", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "onboarding_finalization"
	status.Progress = 0.8

	// 完成入职流程
	err := workflow.ExecuteActivity(ctx, "FinalizeOnboardingActivity", req.OperationData).Get(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to finalize onboarding: %w", err)
	}

	// 创建初始职位历史记录
	var historyResult CreatePositionHistoryResult
	err = workflow.ExecuteActivity(ctx, "CreatePositionHistoryActivity", CreatePositionHistoryRequest{
		TenantID:      req.TenantID,
		EmployeeID:    req.EmployeeID,
		CreatedBy:     req.RequestedBy,
		IsRetroactive: false,
	}).Get(ctx, &historyResult)

	if err != nil {
		logger.Warn("Failed to create initial position history", "error", err)
		// 不阻止入职完成
	}

	result.CompletedSteps = append(result.CompletedSteps, "onboarding_finalized", "initial_position_created")
	status.Progress = 1.0

	logger.Info("Onboarding process finalized successfully", "employee_id", req.EmployeeID)
	return result, nil
}

// handlePositionChange 处理职位变更
func handlePositionChange(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing position change", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "position_change"
	status.Progress = 0.5

	// 使用现有的职位变更工作流
	var positionResult PositionChangeResult
	err := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("position-change-%s-%d",
				req.EmployeeID.String(),
				workflow.Now(ctx).Unix()),
		}),
		"PositionChangeWorkflow",
		PositionChangeRequest{
			TenantID:    req.TenantID,
			EmployeeID:  req.EmployeeID,
			RequestedBy: req.RequestedBy,
			// 从 req.OperationData 中提取职位变更信息
		},
	).Get(ctx, &positionResult)

	if err != nil {
		return result, fmt.Errorf("position change workflow failed: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "position_changed")
	result.ResultData = positionResult
	status.Progress = 1.0

	logger.Info("Position change completed",
		"employee_id", req.EmployeeID,
		"success", positionResult.Success)
	return result, nil
}

// handleUpdateInformation 处理信息更新
func handleUpdateInformation(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	updateType := req.OperationData["update_type"].(string)
	logger.Info("Updating employee information", "employee_id", req.EmployeeID, "update_type", updateType)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "information_update"
	status.Progress = 0.5

	// 更新员工信息
	var updateResult InformationUpdateResult
	err := workflow.ExecuteActivity(ctx, "UpdateEmployeeInformationActivity", InformationUpdateRequest{
		TenantID:   req.TenantID,
		EmployeeID: req.EmployeeID,
		UpdateType: updateType,
		UpdateData: req.OperationData,
		UpdatedBy:  req.RequestedBy,
	}).Get(ctx, &updateResult)

	if err != nil {
		return result, fmt.Errorf("failed to update employee information: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, fmt.Sprintf("information_updated_%s", updateType))
	result.ResultData = updateResult
	status.Progress = 1.0

	logger.Info("Employee information updated", "employee_id", req.EmployeeID, "update_type", updateType)
	return result, nil
}

// handlePerformanceReview 处理绩效评估
func handlePerformanceReview(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing performance review", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "performance_review"
	status.Progress = 0.4

	// 执行绩效评估
	var reviewResult PerformanceReviewResult
	err := workflow.ExecuteActivity(ctx, "ProcessPerformanceReviewActivity", PerformanceReviewRequest{
		TenantID:    req.TenantID,
		EmployeeID:  req.EmployeeID,
		RequestedBy: req.RequestedBy,
	}).Get(ctx, &reviewResult)

	if err != nil {
		return result, fmt.Errorf("performance review failed: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "performance_review_completed")
	result.ResultData = reviewResult
	status.Progress = 1.0

	logger.Info("Performance review completed", "employee_id", req.EmployeeID, "review_id", reviewResult.ReviewID)
	return result, nil
}

// handleLeaveRequest 处理休假请求
func handleLeaveRequest(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing leave request", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "leave_request"
	status.Progress = 0.5

	// 使用现有的休假审批工作流
	var leaveResult LeaveApprovalResult
	err := workflow.ExecuteChildWorkflow(
		workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID: fmt.Sprintf("leave-request-%s-%d",
				req.EmployeeID.String(),
				workflow.Now(ctx).Unix()),
		}),
		"EnhancedLeaveApprovalWorkflow",
		LeaveApprovalRequest{
			EmployeeID:  req.EmployeeID,
			TenantID:    req.TenantID,
			RequestedAt: workflow.Now(ctx),
		},
	).Get(ctx, &leaveResult)

	if err != nil {
		return result, fmt.Errorf("leave approval workflow failed: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "leave_request_processed")
	result.ResultData = leaveResult
	status.Progress = 1.0

	logger.Info("Leave request processed", "employee_id", req.EmployeeID, "status", leaveResult.Status)
	return result, nil
}

// handleStartOffboarding 处理开始离职
func handleStartOffboarding(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting offboarding process", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "offboarding_initialization"
	status.Progress = 0.8

	// 初始化离职流程
	var initResult OffboardingInitializationResult
	err := workflow.ExecuteActivity(ctx, "InitializeOffboardingActivity", OffboardingInitializationRequest{
		TenantID:    req.TenantID,
		EmployeeID:  req.EmployeeID,
		InitiatedBy: req.RequestedBy,
	}).Get(ctx, &initResult)

	if err != nil {
		return result, fmt.Errorf("failed to initialize offboarding: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "offboarding_started")
	result.ResultData = initResult
	status.Progress = 1.0

	logger.Info("Offboarding process started", "employee_id", req.EmployeeID, "offboarding_id", initResult.OffboardingID)
	return result, nil
}

// handleCompleteOffboardingStep 处理完成离职步骤
func handleCompleteOffboardingStep(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	stepID := req.OperationData["step_id"].(string)
	logger.Info("Completing offboarding step", "employee_id", req.EmployeeID, "step_id", stepID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = fmt.Sprintf("completing_offboarding_step_%s", stepID)
	status.Progress = 0.85

	// 完成离职步骤
	err := workflow.ExecuteActivity(ctx, "CompleteOffboardingStepActivity", req.OperationData).Get(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to complete offboarding step %s: %w", stepID, err)
	}

	result.CompletedSteps = append(result.CompletedSteps, fmt.Sprintf("offboarding_step_%s_completed", stepID))
	status.Progress = 1.0

	logger.Info("Offboarding step completed", "employee_id", req.EmployeeID, "step_id", stepID)
	return result, nil
}

// handleFinalizeTermination 处理完成离职
func handleFinalizeTermination(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Finalizing termination", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "termination_finalization"
	status.Progress = 0.9

	// 完成离职流程
	err := workflow.ExecuteActivity(ctx, "FinalizeTerminationActivity", req.OperationData).Get(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to finalize termination: %w", err)
	}

	// 更新职位历史记录（结束当前职位）
	err = workflow.ExecuteActivity(ctx, "EndCurrentPositionActivity", req.EmployeeID).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to end current position", "error", err)
		// 不阻止离职完成
	}

	result.CompletedSteps = append(result.CompletedSteps, "termination_finalized", "position_ended")
	status.Progress = 1.0

	logger.Info("Termination finalized successfully", "employee_id", req.EmployeeID)
	return result, nil
}

// handleArchiveRecords 处理记录归档
func handleArchiveRecords(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Archiving employee records", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "record_archival"
	status.Progress = 0.95

	// 归档员工记录
	var archiveResult RecordArchivalResult
	err := workflow.ExecuteActivity(ctx, "ArchiveEmployeeRecordsActivity", RecordArchivalRequest{
		TenantID:   req.TenantID,
		EmployeeID: req.EmployeeID,
		ArchivedBy: req.RequestedBy,
	}).Get(ctx, &archiveResult)

	if err != nil {
		return result, fmt.Errorf("failed to archive records: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "records_archived")
	result.ResultData = archiveResult
	status.Progress = 1.0

	logger.Info("Employee records archived", "employee_id", req.EmployeeID, "archive_id", archiveResult.ArchiveID)
	return result, nil
}

// handleDataRetention 处理数据保留
func handleDataRetention(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing data retention", "employee_id", req.EmployeeID)

	if err := waitForResumeIfPaused(ctx, isPaused, cancelRequested); err != nil {
		return result, err
	}

	status.CurrentStep = "data_retention"
	status.Progress = 0.98

	// 处理数据保留
	var retentionResult DataRetentionResult
	err := workflow.ExecuteActivity(ctx, "ProcessDataRetentionActivity", DataRetentionRequest{
		TenantID:    req.TenantID,
		EmployeeID:  req.EmployeeID,
		ProcessedBy: req.RequestedBy,
	}).Get(ctx, &retentionResult)

	if err != nil {
		return result, fmt.Errorf("failed to process data retention: %w", err)
	}

	result.CompletedSteps = append(result.CompletedSteps, "data_retention_processed")
	result.ResultData = retentionResult
	status.Progress = 1.0

	logger.Info("Data retention processed", "employee_id", req.EmployeeID, "retention_id", retentionResult.RetentionID)
	return result, nil
}

// HireApprovalRequest 招聘审批请求
type HireApprovalRequest struct {
	TenantID    uuid.UUID              `json:"tenant_id"`
	EmployeeID  uuid.UUID              `json:"employee_id"`
	HireData    map[string]interface{} `json:"hire_data"`
	RequestedBy uuid.UUID              `json:"requested_by"`
}
