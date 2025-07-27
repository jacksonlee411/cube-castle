package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
)

const (
	// TaskQueue 任务队列名称
	TaskQueue = "cube-castle-corehr"
	
	// Workflow名称
	EmployeeOnboardingWorkflowName = "EmployeeOnboardingWorkflow"
	LeaveApprovalWorkflowName      = "LeaveApprovalWorkflow"
)

// TemporalManager Temporal工作流管理器
type TemporalManager struct {
	client     client.Client
	worker     worker.Worker
	logger     *logging.StructuredLogger
	activities *Activities
}

// NewTemporalManager 创建新的Temporal管理器
func NewTemporalManager(temporalHostPort string, logger *logging.StructuredLogger) (*TemporalManager, error) {
	// 创建Temporal客户端
	c, err := client.Dial(client.Options{
		HostPort: temporalHostPort,
		Logger:   NewTemporalLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}

	// 创建Worker
	w := worker.New(c, TaskQueue, worker.Options{})

	// 创建活动处理器
	activities := NewActivities(logger)

	// 注册工作流和活动
	w.RegisterWorkflow(EmployeeOnboardingWorkflow)
	w.RegisterWorkflow(LeaveApprovalWorkflow)
	
	w.RegisterActivity(activities.CreateEmployeeAccountActivity)
	w.RegisterActivity(activities.AssignEquipmentAndPermissionsActivity)
	w.RegisterActivity(activities.SendWelcomeEmailActivity)
	w.RegisterActivity(activities.NotifyManagerActivity)
	w.RegisterActivity(activities.ValidateLeaveRequestActivity)
	w.RegisterActivity(activities.NotifyManagerForApprovalActivity)
	w.RegisterActivity(activities.WaitForManagerApprovalActivity)
	w.RegisterActivity(activities.SendLeaveApprovedNotificationActivity)
	w.RegisterActivity(activities.SendLeaveRejectedNotificationActivity)

	return &TemporalManager{
		client:     c,
		worker:     w,
		logger:     logger,
		activities: activities,
	}, nil
}

// Start 启动Temporal Worker
func (tm *TemporalManager) Start(ctx context.Context) error {
	tm.logger.Info("Starting Temporal worker", "task_queue", TaskQueue)
	
	// 启动Worker
	err := tm.worker.Start()
	if err != nil {
		return fmt.Errorf("unable to start worker: %w", err)
	}

	tm.logger.Info("✅ Temporal worker started successfully")
	return nil
}

// Stop 停止Temporal Worker
func (tm *TemporalManager) Stop() {
	tm.logger.Info("Stopping Temporal worker")
	tm.worker.Stop()
	tm.client.Close()
	tm.logger.Info("✅ Temporal worker stopped successfully")
}

// StartEmployeeOnboarding 启动员工入职工作流
func (tm *TemporalManager) StartEmployeeOnboarding(ctx context.Context, req EmployeeOnboardingRequest) (string, error) {
	start := time.Now()
	
	workflowID := fmt.Sprintf("employee-onboarding-%s", req.EmployeeID.String())
	
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
		WorkflowExecutionTimeout: time.Hour * 24, // 24小时超时
		WorkflowRunTimeout:       time.Hour * 2,  // 单次运行2小时超时
		WorkflowTaskTimeout:      time.Minute * 5, // 任务5分钟超时
	}

	tm.logger.Info("Starting employee onboarding workflow",
		"workflow_id", workflowID,
		"employee_id", req.EmployeeID,
		"tenant_id", req.TenantID)

	we, err := tm.client.ExecuteWorkflow(ctx, workflowOptions, EmployeeOnboardingWorkflowName, req)
	if err != nil {
		tm.logger.LogError("start_workflow", "Failed to start employee onboarding workflow", err, map[string]interface{}{
			"workflow_id": workflowID,
			"employee_id": req.EmployeeID,
		})
		metrics.RecordError("temporal", "workflow_start_error")
		return "", fmt.Errorf("unable to execute workflow: %w", err)
	}

	// 记录指标
	duration := time.Since(start)
	metrics.RecordAIRequest("start_onboarding_workflow", "success", duration)
	tm.logger.LogWorkflowEvent(workflowID, "employee_onboarding", "started", duration)

	tm.logger.Info("Employee onboarding workflow started successfully",
		"workflow_id", workflowID,
		"run_id", we.GetRunID())

	return workflowID, nil
}

// StartLeaveApproval 启动休假审批工作流
func (tm *TemporalManager) StartLeaveApproval(ctx context.Context, req LeaveApprovalRequest) (string, error) {
	start := time.Now()
	
	workflowID := fmt.Sprintf("leave-approval-%s", req.RequestID.String())
	
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
		WorkflowExecutionTimeout: time.Hour * 24 * 8, // 8天超时（包含等待审批时间）
		WorkflowRunTimeout:       time.Hour * 24 * 8, // 单次运行8天超时
		WorkflowTaskTimeout:      time.Minute * 5,     // 任务5分钟超时
	}

	tm.logger.Info("Starting leave approval workflow",
		"workflow_id", workflowID,
		"request_id", req.RequestID,
		"employee_id", req.EmployeeID)

	we, err := tm.client.ExecuteWorkflow(ctx, workflowOptions, LeaveApprovalWorkflowName, req)
	if err != nil {
		tm.logger.LogError("start_workflow", "Failed to start leave approval workflow", err, map[string]interface{}{
			"workflow_id": workflowID,
			"request_id":  req.RequestID,
		})
		metrics.RecordError("temporal", "workflow_start_error")
		return "", fmt.Errorf("unable to execute workflow: %w", err)
	}

	// 记录指标
	duration := time.Since(start)
	metrics.RecordAIRequest("start_leave_approval_workflow", "success", duration)
	tm.logger.LogWorkflowEvent(workflowID, "leave_approval", "started", duration)

	tm.logger.Info("Leave approval workflow started successfully",
		"workflow_id", workflowID,
		"run_id", we.GetRunID())

	return workflowID, nil
}

// GetWorkflowStatus 获取工作流状态
func (tm *TemporalManager) GetWorkflowStatus(ctx context.Context, workflowID string) (*WorkflowStatusInfo, error) {
	// 描述工作流执行
	describe, err := tm.client.DescribeWorkflowExecution(ctx, workflowID, "")
	if err != nil {
		return nil, fmt.Errorf("unable to describe workflow: %w", err)
	}

	status := &WorkflowStatusInfo{
		WorkflowID: workflowID,
		RunID:      describe.WorkflowExecutionInfo.GetExecution().GetRunId(),
		Status:     describe.WorkflowExecutionInfo.GetStatus().String(),
		StartTime:  *describe.WorkflowExecutionInfo.GetStartTime(),
	}

	if describe.WorkflowExecutionInfo.GetCloseTime() != nil {
		status.EndTime = describe.WorkflowExecutionInfo.GetCloseTime()
	}

	// 如果工作流已完成，获取结果
	if describe.WorkflowExecutionInfo.GetStatus().String() == "WORKFLOW_EXECUTION_STATUS_COMPLETED" {
		// 根据工作流类型获取结果
		if contains(workflowID, "employee-onboarding") {
			var result EmployeeOnboardingResult
			err = tm.client.GetWorkflow(ctx, workflowID, "").Get(ctx, &result)
			if err == nil {
				status.Result = &result
			}
		} else if contains(workflowID, "leave-approval") {
			var result LeaveApprovalResult
			err = tm.client.GetWorkflow(ctx, workflowID, "").Get(ctx, &result)
			if err == nil {
				status.Result = &result
			}
		}
	}

	return status, nil
}

// ListWorkflows 列出工作流
func (tm *TemporalManager) ListWorkflows(ctx context.Context, filter string, pageSize int) ([]*WorkflowStatusInfo, error) {
	// 这是一个简化的实现
	// 实际项目中应该使用Temporal的List API
	
	tm.logger.Info("Listing workflows", "filter", filter, "page_size", pageSize)
	
	// 暂时返回空列表，实际实现需要调用Temporal的API
	return []*WorkflowStatusInfo{}, nil
}

// CancelWorkflow 取消工作流
func (tm *TemporalManager) CancelWorkflow(ctx context.Context, workflowID string, reason string) error {
	tm.logger.Info("Cancelling workflow", "workflow_id", workflowID, "reason", reason)
	
	err := tm.client.CancelWorkflow(ctx, workflowID, "")
	if err != nil {
		tm.logger.LogError("cancel_workflow", "Failed to cancel workflow", err, map[string]interface{}{
			"workflow_id": workflowID,
			"reason":      reason,
		})
		return fmt.Errorf("unable to cancel workflow: %w", err)
	}

	tm.logger.Info("Workflow cancelled successfully", "workflow_id", workflowID)
	return nil
}

// HealthCheck 健康检查
func (tm *TemporalManager) HealthCheck(ctx context.Context) error {
	// 简单的健康检查
	if tm.client == nil {
		return fmt.Errorf("temporal client is not initialized")
	}
	
	return nil
}

// WorkflowStatusInfo 工作流状态信息（重命名避免冲突）
type WorkflowStatusInfo struct {
	WorkflowID string      `json:"workflow_id"`
	RunID      string      `json:"run_id"`
	Status     string      `json:"status"`
	StartTime  time.Time   `json:"start_time"`
	EndTime    *time.Time  `json:"end_time,omitempty"`
	Result     interface{} `json:"result,omitempty"`
}

// TemporalLogger 定义在enhanced_manager.go中，此处不重复定义

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr ||
		   len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		   len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-1] == substr
}