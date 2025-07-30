// internal/workflow/employee_lifecycle_workflow.go
package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// EmployeeLifecycleWorkflow 员工完整生命周期工作流
// 统一管理员工从入职、职位变更到离职的全过程
func EmployeeLifecycleWorkflow(ctx workflow.Context, req EmployeeLifecycleRequest) (*EmployeeLifecycleResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting employee lifecycle workflow",
		"employee_id", req.EmployeeID,
		"tenant_id", req.TenantID,
		"lifecycle_stage", req.LifecycleStage,
		"operation", req.Operation)

	// 设置活动选项
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	result := &EmployeeLifecycleResult{
		EmployeeID:     req.EmployeeID,
		LifecycleStage: req.LifecycleStage,
		Operation:      req.Operation,
		StartedAt:      workflow.Now(ctx),
		Status:         "in_progress",
		CompletedSteps: []string{},
	}

	// 工作流状态跟踪
	workflowStatus := LifecycleWorkflowStatus{
		Stage:       req.LifecycleStage,
		Operation:   req.Operation,
		Status:      "in_progress",
		CurrentStep: "validation",
		StartedAt:   workflow.Now(ctx),
		LastUpdated: workflow.Now(ctx),
		Progress:    0.0,
	}

	// 设置查询处理器
	err := workflow.SetQueryHandler(ctx, QueryLifecycleStatus, func() (LifecycleWorkflowStatus, error) {
		return workflowStatus, nil
	})
	if err != nil {
		logger.Error("Failed to set lifecycle query handler", "error", err)
	}

	// 设置信号处理器
	pauseChannel := workflow.GetSignalChannel(ctx, SignalPauseLifecycle)
	resumeChannel := workflow.GetSignalChannel(ctx, SignalResumeLifecycle)
	cancelChannel := workflow.GetSignalChannel(ctx, SignalCancelLifecycle)

	var isPaused bool
	var cancelRequested bool

	// 异步处理信号
	workflow.Go(ctx, func(ctx workflow.Context) {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(pauseChannel, func(c workflow.ReceiveChannel, more bool) {
			var signal LifecyclePauseSignal
			c.Receive(ctx, &signal)
			logger.Info("Received pause signal", "reason", signal.Reason)
			isPaused = true
			workflowStatus.Status = "paused"
			workflowStatus.LastUpdated = workflow.Now(ctx)
		})

		selector.AddReceive(resumeChannel, func(c workflow.ReceiveChannel, more bool) {
			var signal LifecycleResumeSignal
			c.Receive(ctx, &signal)
			logger.Info("Received resume signal", "reason", signal.Reason)
			isPaused = false
			workflowStatus.Status = "in_progress"
			workflowStatus.LastUpdated = workflow.Now(ctx)
		})

		selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
			var signal LifecycleCancelSignal
			c.Receive(ctx, &signal)
			logger.Info("Received cancel signal", "reason", signal.Reason)
			cancelRequested = true
			workflowStatus.Status = "cancelled"
			workflowStatus.LastUpdated = workflow.Now(ctx)
		})

		for !cancelRequested {
			selector.Select(ctx)
		}
	})

	// 根据生命周期阶段和操作类型执行相应流程
	switch req.LifecycleStage {
	case LifecycleStagePREHIRE:
		result, err = handlePreHireStage(ctx, req, result, &workflowStatus, &isPaused, &cancelRequested)
	case LifecycleStageONBOARDING:
		result, err = handleOnboardingStage(ctx, req, result, &workflowStatus, &isPaused, &cancelRequested)
	case LifecycleStageACTIVE:
		result, err = handleActiveStage(ctx, req, result, &workflowStatus, &isPaused, &cancelRequested)
	case LifecycleStageOFFBOARDING:
		result, err = handleOffboardingStage(ctx, req, result, &workflowStatus, &isPaused, &cancelRequested)
	case LifecycleStageTERMINATED:
		result, err = handleTerminatedStage(ctx, req, result, &workflowStatus, &isPaused, &cancelRequested)
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("Unsupported lifecycle stage: %s", req.LifecycleStage)
		return result, nil
	}

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		workflowStatus.Status = "failed"
		logger.Error("Employee lifecycle workflow failed", "error", err)
		return result, err
	}

	// 检查是否被取消
	if cancelRequested {
		result.Status = "cancelled"
		result.Error = "Workflow cancelled by user request"
		workflowStatus.Status = "cancelled"
		logger.Info("Workflow cancelled", "employee_id", req.EmployeeID)
		return result, nil
	}

	// 完成工作流
	result.Status = "completed"
	result.CompletedAt = workflow.Now(ctx)
	workflowStatus.Status = "completed"
	workflowStatus.Progress = 1.0
	workflowStatus.CurrentStep = "completed"
	workflowStatus.LastUpdated = workflow.Now(ctx)

	logger.Info("Employee lifecycle workflow completed successfully",
		"employee_id", req.EmployeeID,
		"lifecycle_stage", req.LifecycleStage,
		"operation", req.Operation,
		"duration", result.CompletedAt.Sub(result.StartedAt))

	return result, nil
}

// handlePreHireStage 处理招聘前阶段
func handlePreHireStage(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing pre-hire stage", "operation", req.Operation)

	status.CurrentStep = "pre_hire_validation"
	status.Progress = 0.1

	switch req.Operation {
	case OperationCREATE_CANDIDATE:
		return handleCreateCandidate(ctx, req, result, status, isPaused, cancelRequested)
	case OperationUPDATE_CANDIDATE:
		return handleUpdateCandidate(ctx, req, result, status, isPaused, cancelRequested)
	case OperationAPPROVE_HIRE:
		return handleApproveHire(ctx, req, result, status, isPaused, cancelRequested)
	default:
		return result, fmt.Errorf("unsupported pre-hire operation: %s", req.Operation)
	}
}

// handleOnboardingStage 处理入职阶段
func handleOnboardingStage(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing onboarding stage", "operation", req.Operation)

	status.CurrentStep = "onboarding_initialization"
	status.Progress = 0.2

	switch req.Operation {
	case OperationSTART_ONBOARDING:
		return handleStartOnboarding(ctx, req, result, status, isPaused, cancelRequested)
	case OperationCOMPLETE_ONBOARDING_STEP:
		return handleCompleteOnboardingStep(ctx, req, result, status, isPaused, cancelRequested)
	case OperationFINALIZE_ONBOARDING:
		return handleFinalizeOnboarding(ctx, req, result, status, isPaused, cancelRequested)
	default:
		return result, fmt.Errorf("unsupported onboarding operation: %s", req.Operation)
	}
}

// handleActiveStage 处理在职阶段
func handleActiveStage(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing active stage", "operation", req.Operation)

	status.CurrentStep = "active_employee_processing"
	status.Progress = 0.3

	switch req.Operation {
	case OperationPOSITION_CHANGE:
		return handlePositionChange(ctx, req, result, status, isPaused, cancelRequested)
	case OperationUPDATE_INFORMATION:
		return handleUpdateInformation(ctx, req, result, status, isPaused, cancelRequested)
	case OperationPERFORMANCE_REVIEW:
		return handlePerformanceReview(ctx, req, result, status, isPaused, cancelRequested)
	case OperationLEAVE_REQUEST:
		return handleLeaveRequest(ctx, req, result, status, isPaused, cancelRequested)
	default:
		return result, fmt.Errorf("unsupported active stage operation: %s", req.Operation)
	}
}

// handleOffboardingStage 处理离职阶段
func handleOffboardingStage(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing offboarding stage", "operation", req.Operation)

	status.CurrentStep = "offboarding_initialization"
	status.Progress = 0.8

	switch req.Operation {
	case OperationSTART_OFFBOARDING:
		return handleStartOffboarding(ctx, req, result, status, isPaused, cancelRequested)
	case OperationCOMPLETE_OFFBOARDING_STEP:
		return handleCompleteOffboardingStep(ctx, req, result, status, isPaused, cancelRequested)
	case OperationFINALIZE_TERMINATION:
		return handleFinalizeTermination(ctx, req, result, status, isPaused, cancelRequested)
	default:
		return result, fmt.Errorf("unsupported offboarding operation: %s", req.Operation)
	}
}

// handleTerminatedStage 处理已离职阶段
func handleTerminatedStage(ctx workflow.Context, req EmployeeLifecycleRequest, result *EmployeeLifecycleResult,
	status *LifecycleWorkflowStatus, isPaused *bool, cancelRequested *bool) (*EmployeeLifecycleResult, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("Processing terminated stage", "operation", req.Operation)

	status.CurrentStep = "post_termination_processing"
	status.Progress = 0.9

	switch req.Operation {
	case OperationARCHIVE_RECORDS:
		return handleArchiveRecords(ctx, req, result, status, isPaused, cancelRequested)
	case OperationDATA_RETENTION:
		return handleDataRetention(ctx, req, result, status, isPaused, cancelRequested)
	default:
		return result, fmt.Errorf("unsupported terminated stage operation: %s", req.Operation)
	}
}

// 辅助函数：等待暂停状态恢复
func waitForResumeIfPaused(ctx workflow.Context, isPaused *bool, cancelRequested *bool) error {
	for *isPaused && !*cancelRequested {
		// 等待1分钟后重新检查
		_ = workflow.Sleep(ctx, time.Minute)
	}

	if *cancelRequested {
		return temporal.NewApplicationError("workflow_cancelled", "CANCELLED")
	}

	return nil
}

// 辅助函数：检查取消状态
func checkCancellation(cancelRequested *bool) error {
	if *cancelRequested {
		return temporal.NewApplicationError("workflow_cancelled", "CANCELLED")
	}
	return nil
}

// 常量定义
const (
	QueryLifecycleStatus = "lifecycle_status"

	SignalPauseLifecycle  = "pause_lifecycle"
	SignalResumeLifecycle = "resume_lifecycle"
	SignalCancelLifecycle = "cancel_lifecycle"
)

// 生命周期阶段常量
const (
	LifecycleStagePREHIRE     = "PRE_HIRE"
	LifecycleStageONBOARDING  = "ONBOARDING"
	LifecycleStageACTIVE      = "ACTIVE"
	LifecycleStageOFFBOARDING = "OFFBOARDING"
	LifecycleStageTERMINATED  = "TERMINATED"
)

// 操作类型常量
const (
	// Pre-hire operations
	OperationCREATE_CANDIDATE = "CREATE_CANDIDATE"
	OperationUPDATE_CANDIDATE = "UPDATE_CANDIDATE"
	OperationAPPROVE_HIRE     = "APPROVE_HIRE"

	// Onboarding operations
	OperationSTART_ONBOARDING         = "START_ONBOARDING"
	OperationCOMPLETE_ONBOARDING_STEP = "COMPLETE_ONBOARDING_STEP"
	OperationFINALIZE_ONBOARDING      = "FINALIZE_ONBOARDING"

	// Active stage operations
	OperationPOSITION_CHANGE    = "POSITION_CHANGE"
	OperationUPDATE_INFORMATION = "UPDATE_INFORMATION"
	OperationPERFORMANCE_REVIEW = "PERFORMANCE_REVIEW"
	OperationLEAVE_REQUEST      = "LEAVE_REQUEST"

	// Offboarding operations
	OperationSTART_OFFBOARDING         = "START_OFFBOARDING"
	OperationCOMPLETE_OFFBOARDING_STEP = "COMPLETE_OFFBOARDING_STEP"
	OperationFINALIZE_TERMINATION      = "FINALIZE_TERMINATION"

	// Post-termination operations
	OperationARCHIVE_RECORDS = "ARCHIVE_RECORDS"
	OperationDATA_RETENTION  = "DATA_RETENTION"
)
