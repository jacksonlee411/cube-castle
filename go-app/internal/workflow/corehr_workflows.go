package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.temporal.io/sdk/activity"
	"github.com/google/uuid"
)

// === 工作流请求和响应类型 ===

// EmployeeOnboardingRequest 员工入职工作流请求
type EmployeeOnboardingRequest struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Department   string    `json:"department"`
	Position     string    `json:"position"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	StartDate    time.Time `json:"start_date"`
}

// EmployeeOnboardingResult 员工入职工作流结果
type EmployeeOnboardingResult struct {
	EmployeeID    uuid.UUID `json:"employee_id"`
	Status        string    `json:"status"`
	CompletedSteps []string `json:"completed_steps"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CompletedAt   time.Time `json:"completed_at"`
}

// LeaveApprovalRequest 休假审批工作流请求
type LeaveApprovalRequest struct {
	RequestID     uuid.UUID `json:"request_id"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	LeaveType     string    `json:"leave_type"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Reason        string    `json:"reason"`
	ManagerID     uuid.UUID `json:"manager_id"`
	RequestedAt   time.Time `json:"requested_at"`
}

// LeaveApprovalResult 休假审批工作流结果
type LeaveApprovalResult struct {
	RequestID     uuid.UUID `json:"request_id"`
	Status        string    `json:"status"` // pending, approved, rejected
	ApproverID    *uuid.UUID `json:"approver_id,omitempty"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	Comments      string    `json:"comments,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// === 员工入职工作流 ===

// EmployeeOnboardingWorkflow 员工入职工作流
func EmployeeOnboardingWorkflow(ctx workflow.Context, req EmployeeOnboardingRequest) (*EmployeeOnboardingResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting employee onboarding workflow", "employee_id", req.EmployeeID)

	result := &EmployeeOnboardingResult{
		EmployeeID:     req.EmployeeID,
		Status:         "in_progress",
		CompletedSteps: []string{},
	}

	// 设置活动选项
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// 步骤1：创建系统账户
	var accountResult CreateAccountResult
	err := workflow.ExecuteActivity(ctx, CreateEmployeeAccountActivity, CreateAccountRequest{
		EmployeeID: req.EmployeeID,
		TenantID:   req.TenantID,
		Email:      req.Email,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
	}).Get(ctx, &accountResult)

	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Failed to create account: %v", err)
		return result, err
	}
	result.CompletedSteps = append(result.CompletedSteps, "account_created")
	logger.Info("Account created successfully", "employee_id", req.EmployeeID)

	// 步骤2：分配设备和权限
	var equipmentResult AssignEquipmentResult
	err = workflow.ExecuteActivity(ctx, AssignEquipmentAndPermissionsActivity, AssignEquipmentRequest{
		EmployeeID: req.EmployeeID,
		TenantID:   req.TenantID,
		Department: req.Department,
		Position:   req.Position,
	}).Get(ctx, &equipmentResult)

	if err != nil {
		logger.Warn("Failed to assign equipment, continuing with onboarding", "error", err)
		// 设备分配失败不应阻止入职流程
	} else {
		result.CompletedSteps = append(result.CompletedSteps, "equipment_assigned")
		logger.Info("Equipment assigned successfully", "employee_id", req.EmployeeID)
	}

	// 步骤3：发送欢迎邮件
	var emailResult SendEmailResult
	err = workflow.ExecuteActivity(ctx, SendWelcomeEmailActivity, WelcomeEmailRequest{
		EmployeeID: req.EmployeeID,
		Email:      req.Email,
		FirstName:  req.FirstName,
		StartDate:  req.StartDate,
		Department: req.Department,
	}).Get(ctx, &emailResult)

	if err != nil {
		logger.Warn("Failed to send welcome email", "error", err)
		// 邮件发送失败不应阻止入职流程
	} else {
		result.CompletedSteps = append(result.CompletedSteps, "welcome_email_sent")
		logger.Info("Welcome email sent successfully", "employee_id", req.EmployeeID)
	}

	// 步骤4：通知经理
	if req.ManagerID != nil {
		var managerNotificationResult NotifyManagerResult
		err = workflow.ExecuteActivity(ctx, NotifyManagerActivity, NotifyManagerRequest{
			ManagerID:      *req.ManagerID,
			NewEmployeeID:  req.EmployeeID,
			EmployeeName:   fmt.Sprintf("%s %s", req.FirstName, req.LastName),
			StartDate:      req.StartDate,
			Department:     req.Department,
			Position:       req.Position,
		}).Get(ctx, &managerNotificationResult)

		if err != nil {
			logger.Warn("Failed to notify manager", "error", err)
		} else {
			result.CompletedSteps = append(result.CompletedSteps, "manager_notified")
			logger.Info("Manager notified successfully", "manager_id", req.ManagerID)
		}
	}

	// 完成工作流
	result.Status = "completed"
	result.CompletedAt = time.Now()
	
	logger.Info("Employee onboarding workflow completed", 
		"employee_id", req.EmployeeID, 
		"completed_steps", len(result.CompletedSteps))

	return result, nil
}

// === 休假审批工作流 ===

// LeaveApprovalWorkflow 休假审批工作流
func LeaveApprovalWorkflow(ctx workflow.Context, req LeaveApprovalRequest) (*LeaveApprovalResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting leave approval workflow", "request_id", req.RequestID)

	result := &LeaveApprovalResult{
		RequestID: req.RequestID,
		Status:    "pending",
	}

	// 设置活动选项
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// 步骤1：验证休假请求
	var validationResult ValidateLeaveRequestResult
	err := workflow.ExecuteActivity(ctx, ValidateLeaveRequestActivity, ValidateLeaveRequestRequest{
		RequestID:  req.RequestID,
		EmployeeID: req.EmployeeID,
		TenantID:   req.TenantID,
		LeaveType:  req.LeaveType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}).Get(ctx, &validationResult)

	if err != nil {
		result.Status = "rejected"
		result.ErrorMessage = fmt.Sprintf("Validation failed: %v", err)
		return result, err
	}

	if !validationResult.IsValid {
		result.Status = "rejected"
		result.Comments = validationResult.Reason
		return result, nil
	}

	// 步骤2：通知经理进行审批
	var notificationResult NotifyManagerForApprovalResult
	err = workflow.ExecuteActivity(ctx, NotifyManagerForApprovalActivity, NotifyManagerForApprovalRequest{
		RequestID:   req.RequestID,
		ManagerID:   req.ManagerID,
		EmployeeID:  req.EmployeeID,
		LeaveType:   req.LeaveType,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Reason:      req.Reason,
		RequestedAt: req.RequestedAt,
	}).Get(ctx, &notificationResult)

	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Failed to notify manager: %v", err)
		return result, err
	}

	// 步骤3：等待经理审批（设置超时为7天）
	approvalCtx, cancel := workflow.WithTimeout(ctx, 7*24*time.Hour)
	defer cancel()

	// 这里应该等待外部信号或人工审批
	// 为了简化，我们使用一个模拟的审批活动
	var approvalResult ManagerApprovalResult
	err = workflow.ExecuteActivity(approvalCtx, WaitForManagerApprovalActivity, WaitForManagerApprovalRequest{
		RequestID: req.RequestID,
		ManagerID: req.ManagerID,
		TimeoutHours: 168, // 7天
	}).Get(approvalCtx, &approvalResult)

	if err != nil {
		if workflow.IsTimeoutError(err) {
			result.Status = "timeout"
			result.Comments = "Manager approval timeout after 7 days"
		} else {
			result.Status = "failed"
			result.ErrorMessage = fmt.Sprintf("Approval process failed: %v", err)
		}
		return result, err
	}

	// 步骤4：处理审批结果
	result.Status = approvalResult.Decision
	result.ApproverID = &approvalResult.ApproverID
	result.ApprovedAt = &approvalResult.ApprovedAt
	result.Comments = approvalResult.Comments

	if approvalResult.Decision == "approved" {
		// 发送审批通过通知
		err = workflow.ExecuteActivity(ctx, SendLeaveApprovedNotificationActivity, LeaveApprovedNotificationRequest{
			RequestID:  req.RequestID,
			EmployeeID: req.EmployeeID,
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			ApproverID: approvalResult.ApproverID,
		}).Get(ctx, nil)

		if err != nil {
			logger.Warn("Failed to send approval notification", "error", err)
		}
	} else {
		// 发送审批拒绝通知
		err = workflow.ExecuteActivity(ctx, SendLeaveRejectedNotificationActivity, LeaveRejectedNotificationRequest{
			RequestID:  req.RequestID,
			EmployeeID: req.EmployeeID,
			Reason:     approvalResult.Comments,
			ApproverID: approvalResult.ApproverID,
		}).Get(ctx, nil)

		if err != nil {
			logger.Warn("Failed to send rejection notification", "error", err)
		}
	}

	logger.Info("Leave approval workflow completed", 
		"request_id", req.RequestID, 
		"status", result.Status)

	return result, nil
}

// === 活动请求和响应类型 ===

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
}

// CreateAccountResult 创建账户结果
type CreateAccountResult struct {
	AccountID string `json:"account_id"`
	Success   bool   `json:"success"`
}

// AssignEquipmentRequest 分配设备请求
type AssignEquipmentRequest struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Department string    `json:"department"`
	Position   string    `json:"position"`
}

// AssignEquipmentResult 分配设备结果
type AssignEquipmentResult struct {
	AssignedItems []string `json:"assigned_items"`
	Success       bool     `json:"success"`
}

// WelcomeEmailRequest 欢迎邮件请求
type WelcomeEmailRequest struct {
	EmployeeID uuid.UUID `json:"employee_id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name"`
	StartDate  time.Time `json:"start_date"`
	Department string    `json:"department"`
}

// SendEmailResult 发送邮件结果
type SendEmailResult struct {
	MessageID string `json:"message_id"`
	Success   bool   `json:"success"`
}

// NotifyManagerRequest 通知经理请求
type NotifyManagerRequest struct {
	ManagerID      uuid.UUID `json:"manager_id"`
	NewEmployeeID  uuid.UUID `json:"new_employee_id"`
	EmployeeName   string    `json:"employee_name"`
	StartDate      time.Time `json:"start_date"`
	Department     string    `json:"department"`
	Position       string    `json:"position"`
}

// NotifyManagerResult 通知经理结果
type NotifyManagerResult struct {
	NotificationID string `json:"notification_id"`
	Success        bool   `json:"success"`
}

// ValidateLeaveRequestRequest 验证休假请求
type ValidateLeaveRequestRequest struct {
	RequestID  uuid.UUID `json:"request_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	LeaveType  string    `json:"leave_type"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
}

// ValidateLeaveRequestResult 验证休假结果
type ValidateLeaveRequestResult struct {
	IsValid bool   `json:"is_valid"`
	Reason  string `json:"reason"`
}

// NotifyManagerForApprovalRequest 通知经理审批请求
type NotifyManagerForApprovalRequest struct {
	RequestID   uuid.UUID `json:"request_id"`
	ManagerID   uuid.UUID `json:"manager_id"`
	EmployeeID  uuid.UUID `json:"employee_id"`
	LeaveType   string    `json:"leave_type"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Reason      string    `json:"reason"`
	RequestedAt time.Time `json:"requested_at"`
}

// NotifyManagerForApprovalResult 通知经理审批结果
type NotifyManagerForApprovalResult struct {
	NotificationID string `json:"notification_id"`
	Success        bool   `json:"success"`
}

// WaitForManagerApprovalRequest 等待经理审批请求
type WaitForManagerApprovalRequest struct {
	RequestID    uuid.UUID `json:"request_id"`
	ManagerID    uuid.UUID `json:"manager_id"`
	TimeoutHours int       `json:"timeout_hours"`
}

// ManagerApprovalResult 经理审批结果
type ManagerApprovalResult struct {
	Decision   string    `json:"decision"` // "approved" or "rejected"
	ApproverID uuid.UUID `json:"approver_id"`
	ApprovedAt time.Time `json:"approved_at"`
	Comments   string    `json:"comments"`
}

// LeaveApprovedNotificationRequest 休假审批通过通知请求
type LeaveApprovedNotificationRequest struct {
	RequestID  uuid.UUID `json:"request_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	ApproverID uuid.UUID `json:"approver_id"`
}

// LeaveRejectedNotificationRequest 休假审批拒绝通知请求
type LeaveRejectedNotificationRequest struct {
	RequestID  uuid.UUID `json:"request_id"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Reason     string    `json:"reason"`
	ApproverID uuid.UUID `json:"approver_id"`
}

// === 信号和查询定义 ===

// 信号名称常量
const (
	SignalApproveLeave = "approve_leave"
	SignalRejectLeave  = "reject_leave"
	SignalCancelWorkflow = "cancel_workflow"
	SignalUpdateWorkflow = "update_workflow"
)

// 查询名称常量
const (
	QueryWorkflowStatus = "workflow_status"
	QueryCompletedSteps = "completed_steps"
	QueryCurrentStep    = "current_step"
)

// ApprovalSignal 审批信号
type ApprovalSignal struct {
	Decision   string    `json:"decision"` // "approved" or "rejected"
	ApproverID uuid.UUID `json:"approver_id"`
	Comments   string    `json:"comments"`
	ApprovedAt time.Time `json:"approved_at"`
}

// WorkflowUpdateSignal 工作流更新信号
type WorkflowUpdateSignal struct {
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
}

// WorkflowStatusQuery 工作流状态查询响应
type WorkflowStatusQuery struct {
	Status       string    `json:"status"`
	CurrentStep  string    `json:"current_step"`
	StartedAt    time.Time `json:"started_at"`
	LastUpdated  time.Time `json:"last_updated"`
	Progress     float64   `json:"progress"`
}

// === 增强的休假审批工作流（支持信号） ===

// EnhancedLeaveApprovalWorkflow 增强的休假审批工作流（支持信号处理）
func EnhancedLeaveApprovalWorkflow(ctx workflow.Context, req LeaveApprovalRequest) (*LeaveApprovalResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting enhanced leave approval workflow", "request_id", req.RequestID)

	result := &LeaveApprovalResult{
		RequestID: req.RequestID,
		Status:    "pending",
	}

	// 工作流状态跟踪
	workflowStatus := WorkflowStatusQuery{
		Status:      "in_progress",
		CurrentStep: "validation",
		StartedAt:   workflow.Now(ctx),
		LastUpdated: workflow.Now(ctx),
		Progress:    0.0,
	}

	// 设置查询处理器
	err := workflow.SetQueryHandler(ctx, QueryWorkflowStatus, func() (WorkflowStatusQuery, error) {
		return workflowStatus, nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
	}

	// 设置活动选项
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// 步骤1：验证休假请求
	workflowStatus.CurrentStep = "validation"
	workflowStatus.Progress = 0.25
	var validationResult ValidateLeaveRequestResult
	err = workflow.ExecuteActivity(ctx, ValidateLeaveRequestActivity, ValidateLeaveRequestRequest{
		RequestID:  req.RequestID,
		EmployeeID: req.EmployeeID,
		TenantID:   req.TenantID,
		LeaveType:  req.LeaveType,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}).Get(ctx, &validationResult)

	if err != nil {
		result.Status = "rejected"
		result.ErrorMessage = fmt.Sprintf("Validation failed: %v", err)
		workflowStatus.Status = "failed"
		workflowStatus.CurrentStep = "validation_failed"
		return result, err
	}

	if !validationResult.IsValid {
		result.Status = "rejected"
		result.Comments = validationResult.Reason
		workflowStatus.Status = "completed"
		workflowStatus.CurrentStep = "rejected"
		workflowStatus.Progress = 1.0
		return result, nil
	}

	// 步骤2：通知经理进行审批
	workflowStatus.CurrentStep = "manager_notification"
	workflowStatus.Progress = 0.5
	var notificationResult NotifyManagerForApprovalResult
	err = workflow.ExecuteActivity(ctx, NotifyManagerForApprovalActivity, NotifyManagerForApprovalRequest{
		RequestID:   req.RequestID,
		ManagerID:   req.ManagerID,
		EmployeeID:  req.EmployeeID,
		LeaveType:   req.LeaveType,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Reason:      req.Reason,
		RequestedAt: req.RequestedAt,
	}).Get(ctx, &notificationResult)

	if err != nil {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("Failed to notify manager: %v", err)
		workflowStatus.Status = "failed"
		workflowStatus.CurrentStep = "notification_failed"
		return result, err
	}

	// 步骤3：等待审批信号或超时
	workflowStatus.CurrentStep = "waiting_approval"
	workflowStatus.Progress = 0.75

	// 设置信号通道
	approvalChannel := workflow.GetSignalChannel(ctx, SignalApproveLeave)
	rejectionChannel := workflow.GetSignalChannel(ctx, SignalRejectLeave)
	cancelChannel := workflow.GetSignalChannel(ctx, SignalCancelWorkflow)

	// 等待信号或超时（7天）
	selector := workflow.NewSelector(ctx)
	var approvalSignal ApprovalSignal
	var cancelRequested bool

	// 审批信号处理
	selector.AddReceive(approvalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &approvalSignal)
		logger.Info("Received approval signal", "decision", approvalSignal.Decision)
	})

	// 拒绝信号处理
	selector.AddReceive(rejectionChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &approvalSignal)
		logger.Info("Received rejection signal", "decision", approvalSignal.Decision)
	})

	// 取消信号处理
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var signal interface{}
		c.Receive(ctx, &signal)
		cancelRequested = true
		logger.Info("Received cancel signal")
	})

	// 设置超时
	timeoutCtx, cancel := workflow.WithTimeout(ctx, 7*24*time.Hour)
	defer cancel()

	// 等待信号
	for approvalSignal.Decision == "" && !cancelRequested {
		selector.Select(timeoutCtx)
		if workflow.IsTimeoutError(workflow.GetLastError(timeoutCtx)) {
			result.Status = "timeout"
			result.Comments = "Manager approval timeout after 7 days"
			workflowStatus.Status = "timeout"
			workflowStatus.CurrentStep = "timeout"
			workflowStatus.Progress = 1.0
			return result, nil
		}
	}

	// 处理取消请求
	if cancelRequested {
		result.Status = "cancelled"
		result.Comments = "Workflow cancelled by user request"
		workflowStatus.Status = "cancelled"
		workflowStatus.CurrentStep = "cancelled"
		workflowStatus.Progress = 1.0
		return result, nil
	}

	// 步骤4：处理审批结果
	workflowStatus.CurrentStep = "processing_result"
	workflowStatus.Progress = 0.9

	result.Status = approvalSignal.Decision
	result.ApproverID = &approvalSignal.ApproverID
	result.ApprovedAt = &approvalSignal.ApprovedAt
	result.Comments = approvalSignal.Comments

	if approvalSignal.Decision == "approved" {
		// 发送审批通过通知
		err = workflow.ExecuteActivity(ctx, SendLeaveApprovedNotificationActivity, LeaveApprovedNotificationRequest{
			RequestID:  req.RequestID,
			EmployeeID: req.EmployeeID,
			StartDate:  req.StartDate,
			EndDate:    req.EndDate,
			ApproverID: approvalSignal.ApproverID,
		}).Get(ctx, nil)

		if err != nil {
			logger.Warn("Failed to send approval notification", "error", err)
		}
		workflowStatus.CurrentStep = "approved"
	} else {
		// 发送审批拒绝通知
		err = workflow.ExecuteActivity(ctx, SendLeaveRejectedNotificationActivity, LeaveRejectedNotificationRequest{
			RequestID:  req.RequestID,
			EmployeeID: req.EmployeeID,
			Reason:     approvalSignal.Comments,
			ApproverID: approvalSignal.ApproverID,
		}).Get(ctx, nil)

		if err != nil {
			logger.Warn("Failed to send rejection notification", "error", err)
		}
		workflowStatus.CurrentStep = "rejected"
	}

	// 完成工作流
	workflowStatus.Status = "completed"
	workflowStatus.Progress = 1.0
	workflowStatus.LastUpdated = workflow.Now(ctx)

	logger.Info("Enhanced leave approval workflow completed", 
		"request_id", req.RequestID, 
		"status", result.Status)

	return result, nil
}

// === 批量员工处理工作流 ===

// BatchEmployeeProcessingRequest 批量员工处理请求
type BatchEmployeeProcessingRequest struct {
	BatchID     uuid.UUID              `json:"batch_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Operation   string                 `json:"operation"` // "onboard", "offboard", "update"
	Employees   []BatchEmployeeData    `json:"employees"`
	Options     map[string]interface{} `json:"options,omitempty"`
	RequestedBy uuid.UUID              `json:"requested_by"`
}

// BatchEmployeeData 批量员工数据
type BatchEmployeeData struct {
	EmployeeID uuid.UUID              `json:"employee_id"`
	Data       map[string]interface{} `json:"data"`
}

// BatchEmployeeProcessingResult 批量员工处理结果
type BatchEmployeeProcessingResult struct {
	BatchID       uuid.UUID                     `json:"batch_id"`
	Status        string                        `json:"status"`
	TotalCount    int                           `json:"total_count"`
	SuccessCount  int                           `json:"success_count"`
	FailureCount  int                           `json:"failure_count"`
	Results       []BatchEmployeeProcessResult  `json:"results"`
	CompletedAt   time.Time                     `json:"completed_at"`
	ErrorMessage  string                        `json:"error_message,omitempty"`
}

// BatchEmployeeProcessResult 单个员工处理结果
type BatchEmployeeProcessResult struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// BatchEmployeeProcessingWorkflow 批量员工处理工作流
func BatchEmployeeProcessingWorkflow(ctx workflow.Context, req BatchEmployeeProcessingRequest) (*BatchEmployeeProcessingResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting batch employee processing workflow", 
		"batch_id", req.BatchID, 
		"operation", req.Operation, 
		"count", len(req.Employees))

	result := &BatchEmployeeProcessingResult{
		BatchID:    req.BatchID,
		Status:     "in_progress",
		TotalCount: len(req.Employees),
		Results:    make([]BatchEmployeeProcessResult, 0, len(req.Employees)),
	}

	// 工作流状态跟踪
	workflowStatus := WorkflowStatusQuery{
		Status:      "in_progress",
		CurrentStep: "processing",
		StartedAt:   workflow.Now(ctx),
		LastUpdated: workflow.Now(ctx),
		Progress:    0.0,
	}

	// 设置查询处理器
	err := workflow.SetQueryHandler(ctx, QueryWorkflowStatus, func() (WorkflowStatusQuery, error) {
		return workflowStatus, nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
	}

	// 设置活动选项
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 10,
		RetryPolicy: &workflow.RetryPolicy{
			InitialInterval:    time.Second * 10,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 5,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// 并行处理员工数据（每批最多10个）
	batchSize := 10
	for i := 0; i < len(req.Employees); i += batchSize {
		end := i + batchSize
		if end > len(req.Employees) {
			end = len(req.Employees)
		}
		
		batch := req.Employees[i:end]
		
		// 为这一批创建并行任务
		futures := make([]workflow.Future, len(batch))
		for j, employee := range batch {
			processReq := ProcessSingleEmployeeRequest{
				BatchID:     req.BatchID,
				TenantID:    req.TenantID,
				Operation:   req.Operation,
				EmployeeID:  employee.EmployeeID,
				Data:        employee.Data,
				Options:     req.Options,
				RequestedBy: req.RequestedBy,
			}
			
			futures[j] = workflow.ExecuteActivity(ctx, ProcessSingleEmployeeActivity, processReq)
		}
		
		// 等待这一批完成
		for j, future := range futures {
			var processResult ProcessSingleEmployeeResult
			err := future.Get(ctx, &processResult)
			
			batchResult := BatchEmployeeProcessResult{
				EmployeeID: batch[j].EmployeeID,
				Status:     processResult.Status,
			}
			
			if err != nil {
				batchResult.Status = "failed"
				batchResult.ErrorMessage = err.Error()
				result.FailureCount++
				logger.Warn("Employee processing failed", 
					"employee_id", batch[j].EmployeeID, 
					"error", err)
			} else if processResult.Status == "success" {
				result.SuccessCount++
			} else {
				batchResult.ErrorMessage = processResult.ErrorMessage
				result.FailureCount++
			}
			
			result.Results = append(result.Results, batchResult)
		}
		
		// 更新进度
		completed := float64(len(result.Results))
		total := float64(result.TotalCount)
		workflowStatus.Progress = completed / total
		workflowStatus.LastUpdated = workflow.Now(ctx)
		
		logger.Info("Batch processed", 
			"completed", len(result.Results), 
			"total", result.TotalCount, 
			"progress", workflowStatus.Progress)
	}

	// 完成工作流
	if result.FailureCount == 0 {
		result.Status = "completed"
	} else if result.SuccessCount == 0 {
		result.Status = "failed"
	} else {
		result.Status = "partial_success"
	}
	
	result.CompletedAt = workflow.Now(ctx)
	workflowStatus.Status = "completed"
	workflowStatus.Progress = 1.0
	workflowStatus.CurrentStep = "completed"
	workflowStatus.LastUpdated = workflow.Now(ctx)

	logger.Info("Batch employee processing workflow completed", 
		"batch_id", req.BatchID,
		"status", result.Status,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount)

	return result, nil
}

// ProcessSingleEmployeeRequest 单个员工处理请求
type ProcessSingleEmployeeRequest struct {
	BatchID     uuid.UUID              `json:"batch_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Operation   string                 `json:"operation"`
	EmployeeID  uuid.UUID              `json:"employee_id"`
	Data        map[string]interface{} `json:"data"`
	Options     map[string]interface{} `json:"options,omitempty"`
	RequestedBy uuid.UUID              `json:"requested_by"`
}

// ProcessSingleEmployeeResult 单个员工处理结果
type ProcessSingleEmployeeResult struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
}