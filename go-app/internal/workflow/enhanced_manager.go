package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// WorkflowManager 增强的工作流管理器
type WorkflowManager struct {
	client    client.Client
	worker    worker.Worker
	logger    *logging.StructuredLogger
	taskQueue string
	namespace string
}

// TemporalWorkflowExecution Temporal工作流执行信息（重命名避免冲突）
type TemporalWorkflowExecution struct {
	WorkflowID string `json:"workflow_id"`
	RunID      string `json:"run_id"`
}

// WorkflowExecutionInfo 工作流执行详细信息
type WorkflowExecutionInfo struct {
	WorkflowID   string     `json:"workflow_id"`
	RunID        string     `json:"run_id"`
	WorkflowType string     `json:"workflow_type"`
	Status       string     `json:"status"`
	StartTime    time.Time  `json:"start_time"`
	CloseTime    *time.Time `json:"close_time,omitempty"`
}

// WorkflowListResponse 工作流列表响应
type WorkflowListResponse struct {
	Executions    []WorkflowExecutionInfo `json:"executions"`
	NextPageToken []byte                  `json:"next_page_token,omitempty"`
}

// WorkflowHistoryEvent 工作流历史事件
type WorkflowHistoryEvent struct {
	EventID    int64                  `json:"event_id"`
	EventType  string                 `json:"event_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

// WorkflowHistory 工作流历史
type WorkflowHistory struct {
	WorkflowID string                 `json:"workflow_id"`
	RunID      string                 `json:"run_id"`
	Events     []WorkflowHistoryEvent `json:"events"`
}

// NewWorkflowManager 创建增强的工作流管理器
func NewWorkflowManager(temporalHostPort string, namespace string, logger *logging.StructuredLogger) (*WorkflowManager, error) {
	// 创建Temporal客户端
	c, err := client.Dial(client.Options{
		HostPort:  temporalHostPort,
		Namespace: namespace,
		Logger:    NewTemporalLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}

	taskQueue := "cube-castle-corehr"

	// 创建Worker
	w := worker.New(c, taskQueue, worker.Options{})

	// 创建活动处理器
	activities := NewActivities(logger)

	// 注册所有工作流
	w.RegisterWorkflow(EmployeeOnboardingWorkflow)
	w.RegisterWorkflow(LeaveApprovalWorkflow)
	w.RegisterWorkflow(EnhancedLeaveApprovalWorkflow)
	w.RegisterWorkflow(BatchEmployeeProcessingWorkflow)

	// 注册所有活动
	w.RegisterActivity(activities.CreateEmployeeAccountActivity)
	w.RegisterActivity(activities.AssignEquipmentAndPermissionsActivity)
	w.RegisterActivity(activities.SendWelcomeEmailActivity)
	w.RegisterActivity(activities.NotifyManagerActivity)
	w.RegisterActivity(activities.ValidateLeaveRequestActivity)
	w.RegisterActivity(activities.NotifyManagerForApprovalActivity)
	w.RegisterActivity(activities.WaitForManagerApprovalActivity)
	w.RegisterActivity(activities.SendLeaveApprovedNotificationActivity)
	w.RegisterActivity(activities.SendLeaveRejectedNotificationActivity)
	w.RegisterActivity(activities.ProcessSingleEmployeeActivity)

	return &WorkflowManager{
		client:    c,
		worker:    w,
		logger:    logger,
		taskQueue: taskQueue,
		namespace: namespace,
	}, nil
}

// Start 启动工作流管理器
func (m *WorkflowManager) Start(ctx context.Context) error {
	m.logger.Info("Starting enhanced workflow manager", "task_queue", m.taskQueue)

	err := m.worker.Start()
	if err != nil {
		return fmt.Errorf("failed to start worker: %w", err)
	}

	m.logger.Info("✅ Enhanced workflow manager started successfully")
	return nil
}

// Stop 停止工作流管理器
func (m *WorkflowManager) Stop() {
	m.logger.Info("Stopping enhanced workflow manager")
	m.worker.Stop()
	m.client.Close()
	m.logger.Info("✅ Enhanced workflow manager stopped successfully")
}

// StartEmployeeOnboardingWorkflow 启动员工入职工作流
func (m *WorkflowManager) StartEmployeeOnboardingWorkflow(ctx context.Context, req EmployeeOnboardingRequest) (*TemporalWorkflowExecution, error) {
	workflowID := fmt.Sprintf("employee-onboarding-%s", req.EmployeeID.String())
	
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: m.taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second * 10,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute * 5,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"ValidationError"},
		},
	}

	execution, err := m.client.ExecuteWorkflow(ctx, options, EmployeeOnboardingWorkflow, req)
	if err != nil {
		m.logger.LogError("workflow", "Failed to start employee onboarding workflow", err, map[string]interface{}{
			"employee_id": req.EmployeeID,
			"tenant_id":   req.TenantID,
		})
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	m.logger.Info("Workflow event: employee_onboarding_started",
		"employee_id", req.EmployeeID.String(),
		"tenant_id", req.TenantID.String(),
		"workflow_id", workflowID)
	
	return &TemporalWorkflowExecution{
		WorkflowID: execution.GetID(),
		RunID:      execution.GetRunID(),
	}, nil
}

// StartLeaveApprovalWorkflow 启动休假审批工作流
func (m *WorkflowManager) StartLeaveApprovalWorkflow(ctx context.Context, req LeaveApprovalRequest) (*TemporalWorkflowExecution, error) {
	workflowID := fmt.Sprintf("leave-approval-%s", req.RequestID.String())
	
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: m.taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second * 10,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute * 5,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"ValidationError"},
		},
	}

	execution, err := m.client.ExecuteWorkflow(ctx, options, LeaveApprovalWorkflow, req)
	if err != nil {
		m.logger.LogError("workflow", "Failed to start leave approval workflow", err, map[string]interface{}{
			"request_id": req.RequestID,
			"employee_id": req.EmployeeID,
			"tenant_id":   req.TenantID,
		})
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	m.logger.Info("Workflow event: leave_approval_started", 
		"employee_id", req.EmployeeID.String(), 
		"tenant_id", req.TenantID.String(), 
		"workflow_id", workflowID)
	
	return &TemporalWorkflowExecution{
		WorkflowID: execution.GetID(),
		RunID:      execution.GetRunID(),
	}, nil
}

// StartEnhancedLeaveApprovalWorkflow 启动增强的休假审批工作流（支持信号）
func (m *WorkflowManager) StartEnhancedLeaveApprovalWorkflow(ctx context.Context, req LeaveApprovalRequest) (*TemporalWorkflowExecution, error) {
	workflowID := fmt.Sprintf("enhanced-leave-approval-%s", req.RequestID.String())
	
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: m.taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second * 10,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute * 5,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"ValidationError"},
		},
	}

	execution, err := m.client.ExecuteWorkflow(ctx, options, EnhancedLeaveApprovalWorkflow, req)
	if err != nil {
		m.logger.LogError("workflow", "Failed to start enhanced leave approval workflow", err, map[string]interface{}{
			"request_id": req.RequestID,
			"employee_id": req.EmployeeID,
			"tenant_id":   req.TenantID,
		})
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	m.logger.Info("Workflow event: enhanced_leave_approval_started", 
		"employee_id", req.EmployeeID.String(), 
		"tenant_id", req.TenantID.String(), 
		"workflow_id", workflowID)
	
	return &TemporalWorkflowExecution{
		WorkflowID: execution.GetID(),
		RunID:      execution.GetRunID(),
	}, nil
}

// StartBatchEmployeeProcessingWorkflow 启动批量员工处理工作流
func (m *WorkflowManager) StartBatchEmployeeProcessingWorkflow(ctx context.Context, req BatchEmployeeProcessingRequest) (*TemporalWorkflowExecution, error) {
	workflowID := fmt.Sprintf("batch-employee-%s", req.BatchID.String())
	
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: m.taskQueue,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second * 10,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Minute * 5,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"ValidationError"},
		},
	}

	execution, err := m.client.ExecuteWorkflow(ctx, options, BatchEmployeeProcessingWorkflow, req)
	if err != nil {
		m.logger.LogError("workflow", "Failed to start batch employee processing workflow", err, map[string]interface{}{
			"batch_id": req.BatchID,
			"operation": req.Operation,
			"tenant_id": req.TenantID,
			"count": len(req.Employees),
		})
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	m.logger.Info("Workflow event: batch_employee_processing_started", 
		"batch_id", req.BatchID.String(), 
		"tenant_id", req.TenantID.String(), 
		"workflow_id", workflowID)
	
	return &TemporalWorkflowExecution{
		WorkflowID: execution.GetID(),
		RunID:      execution.GetRunID(),
	}, nil
}

// SendApprovalSignal 发送审批信号
func (m *WorkflowManager) SendApprovalSignal(ctx context.Context, workflowID string, runID string, signal ApprovalSignal) error {
	var signalName string
	if signal.Decision == "approved" {
		signalName = SignalApproveLeave
	} else {
		signalName = SignalRejectLeave
	}

	err := m.client.SignalWorkflow(ctx, workflowID, runID, signalName, signal)
	if err != nil {
		m.logger.LogError("workflow", "Failed to send approval signal", err, map[string]interface{}{
			"workflow_id": workflowID,
			"run_id":      runID,
			"decision":    signal.Decision,
			"approver_id": signal.ApproverID,
		})
		return fmt.Errorf("failed to send signal: %w", err)
	}

	m.logger.Info("Workflow event: approval_signal_sent", 
		"workflow_id", workflowID, 
		"run_id", runID, 
		"decision", signal.Decision)
	return nil
}

// CancelWorkflow 取消工作流
func (m *WorkflowManager) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	err := m.client.SignalWorkflow(ctx, workflowID, runID, SignalCancelWorkflow, nil)
	if err != nil {
		m.logger.LogError("workflow", "Failed to send cancel signal", err, map[string]interface{}{
			"workflow_id": workflowID,
			"run_id":      runID,
		})
		return fmt.Errorf("failed to cancel workflow: %w", err)
	}

	m.logger.Info("Workflow event: workflow_cancelled", 
		"workflow_id", workflowID, 
		"run_id", runID)
	return nil
}

// QueryWorkflowStatus 查询工作流状态
func (m *WorkflowManager) QueryWorkflowStatus(ctx context.Context, workflowID string, runID string) (*WorkflowStatusQuery, error) {
	response, err := m.client.QueryWorkflow(ctx, workflowID, runID, QueryWorkflowStatus)
	if err != nil {
		m.logger.LogError("workflow", "Failed to query workflow status", err, map[string]interface{}{
			"workflow_id": workflowID,
			"run_id":      runID,
		})
		return nil, fmt.Errorf("failed to query workflow: %w", err)
	}

	var status WorkflowStatusQuery
	err = response.Get(&status)
	if err != nil {
		return nil, fmt.Errorf("failed to decode query response: %w", err)
	}

	return &status, nil
}

// ListWorkflows 列出工作流实例
func (m *WorkflowManager) ListWorkflows(ctx context.Context, query string, pageSize int, nextPageToken []byte) (*WorkflowListResponse, error) {
	request := &workflowservice.ListWorkflowExecutionsRequest{
		Namespace:     m.namespace,
		PageSize:      int32(pageSize),
		NextPageToken: nextPageToken,
		Query:         query,
	}

	response, err := m.client.WorkflowService().ListWorkflowExecutions(ctx, request)
	if err != nil {
		m.logger.LogError("workflow", "Failed to list workflows", err, map[string]interface{}{
			"query": query,
			"page_size": pageSize,
		})
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}

	executions := make([]WorkflowExecutionInfo, 0, len(response.Executions))
	for _, exec := range response.Executions {
		executions = append(executions, WorkflowExecutionInfo{
			WorkflowID:   exec.Execution.WorkflowId,
			RunID:        exec.Execution.RunId,
			WorkflowType: exec.Type.Name,
			Status:       exec.Status.String(),
			StartTime:    *exec.StartTime,
			CloseTime: func() *time.Time {
				if exec.CloseTime != nil {
					return exec.CloseTime
				}
				return nil
			}(),
		})
	}

	return &WorkflowListResponse{
		Executions:    executions,
		NextPageToken: response.NextPageToken,
	}, nil
}

// GetWorkflowHistory 获取工作流历史
func (m *WorkflowManager) GetWorkflowHistory(ctx context.Context, workflowID string, runID string) (*WorkflowHistory, error) {
	iter := m.client.GetWorkflowHistory(ctx, workflowID, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	
	var events []WorkflowHistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next history event: %w", err)
		}

		events = append(events, WorkflowHistoryEvent{
			EventID:   event.EventId,
			EventType: event.EventType.String(),
			Timestamp: *event.EventTime,
			Attributes: func() map[string]interface{} {
				// 简化处理，实际项目中需要根据事件类型解析属性
				return map[string]interface{}{
					"event_id": event.EventId,
				}
			}(),
		})
	}

	return &WorkflowHistory{
		WorkflowID: workflowID,
		RunID:      runID,
		Events:     events,
	}, nil
}

// DescribeWorkflow 描述工作流执行
func (m *WorkflowManager) DescribeWorkflow(ctx context.Context, workflowID string, runID string) (*WorkflowExecutionInfo, error) {
	resp, err := m.client.DescribeWorkflowExecution(ctx, workflowID, runID)
	if err != nil {
		m.logger.LogError("workflow", "Failed to describe workflow", err, map[string]interface{}{
			"workflow_id": workflowID,
			"run_id":      runID,
		})
		return nil, fmt.Errorf("failed to describe workflow: %w", err)
	}

	info := &WorkflowExecutionInfo{
		WorkflowID:   resp.WorkflowExecutionInfo.Execution.WorkflowId,
		RunID:        resp.WorkflowExecutionInfo.Execution.RunId,
		WorkflowType: resp.WorkflowExecutionInfo.Type.Name,
		Status:       resp.WorkflowExecutionInfo.Status.String(),
		StartTime:    *resp.WorkflowExecutionInfo.StartTime,
	}

	if resp.WorkflowExecutionInfo.CloseTime != nil {
		info.CloseTime = resp.WorkflowExecutionInfo.CloseTime
	}

	return info, nil
}

// HealthCheck 健康检查
func (m *WorkflowManager) HealthCheck(ctx context.Context) error {
	// 简化健康检查 - 检查client是否正常
	if m.client == nil {
		return fmt.Errorf("temporal client is not initialized")
	}

	return nil
}

// GetMetrics 获取工作流指标
func (m *WorkflowManager) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	// 这里可以实现更详细的指标收集
	// 例如：活跃工作流数量、完成率、平均执行时间等
	
	// 简化实现，返回基础指标
	activeWorkflows, err := m.ListWorkflows(ctx, "ExecutionStatus='Running'", 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get active workflows: %w", err)
	}

	completedWorkflows, err := m.ListWorkflows(ctx, "ExecutionStatus='Completed'", 1000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed workflows: %w", err)
	}

	return map[string]interface{}{
		"active_workflows":    len(activeWorkflows.Executions),
		"completed_workflows": len(completedWorkflows.Executions),
		"task_queue":          m.taskQueue,
		"namespace":           m.namespace,
		"timestamp":           time.Now(),
	}, nil
}

// TemporalLogger Temporal日志适配器
type TemporalLogger struct {
	logger *logging.StructuredLogger
}

// NewTemporalLogger 创建Temporal日志适配器
func NewTemporalLogger(logger *logging.StructuredLogger) *TemporalLogger {
	return &TemporalLogger{logger: logger}
}

func (l *TemporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Info("temporal debug: "+msg, keyvals...)
}

func (l *TemporalLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, l.convertKeyvals(keyvals)...)
}

func (l *TemporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warn(msg, l.convertKeyvals(keyvals)...)
}

func (l *TemporalLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, l.convertKeyvals(keyvals)...)
}

func (l *TemporalLogger) convertKeyvals(keyvals []interface{}) []interface{} {
	// 简单转换，实际可能需要更复杂的处理
	return keyvals
}