//go:build integration
// +build integration

package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	client     client.Client
	worker     worker.Worker
	logger     *logging.StructuredLogger
	activities *Activities
}

// SetupIntegrationTest 设置集成测试环境
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	logger := logging.NewStructuredLogger()
	
	// 连接到本地Temporal服务
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
		Logger:   NewTemporalLogger(logger),
	})
	require.NoError(t, err, "Failed to connect to Temporal server")

	// 创建Worker
	taskQueue := "test-queue"
	w := worker.New(c, taskQueue, worker.Options{})

	// 创建Activities
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

	// 启动Worker
	err = w.Start()
	require.NoError(t, err, "Failed to start worker")

	return &IntegrationTestSuite{
		client:     c,
		worker:     w,
		logger:     logger,
		activities: activities,
	}
}

// TearDown 清理测试环境
func (suite *IntegrationTestSuite) TearDown() {
	if suite.worker != nil {
		suite.worker.Stop()
	}
	if suite.client != nil {
		suite.client.Close()
	}
}

// TestEmployeeOnboardingWorkflow_Integration 员工入职工作流集成测试
func TestEmployeeOnboardingWorkflow_Integration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TearDown()

	// 准备测试数据
	employeeID := uuid.New()
	tenantID := uuid.New()
	managerID := uuid.New()

	request := EmployeeOnboardingRequest{
		EmployeeID: employeeID,
		TenantID:   tenantID,
		FirstName:  "Integration",
		LastName:   "Test",
		Email:      "integration.test@example.com",
		Department: "Engineering",
		Position:   "Software Engineer",
		ManagerID:  &managerID,
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	// 启动工作流
	workflowOptions := client.StartWorkflowOptions{
		ID:        "integration-test-" + uuid.New().String(),
		TaskQueue: "test-queue",
	}

	ctx := context.Background()
	we, err := suite.client.ExecuteWorkflow(ctx, workflowOptions, EmployeeOnboardingWorkflow, request)
	require.NoError(t, err, "Failed to start workflow")

	suite.logger.Info("Started employee onboarding workflow", "workflow_id", we.GetID())

	// 等待工作流完成
	var result EmployeeOnboardingResult
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")

	// 验证结果
	assert.Equal(t, employeeID, result.EmployeeID)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.CompletedSteps)
	assert.Contains(t, result.CompletedSteps, "account_created")
	assert.Contains(t, result.CompletedSteps, "equipment_assigned")
	assert.Contains(t, result.CompletedSteps, "welcome_email_sent")
	assert.Contains(t, result.CompletedSteps, "manager_notified")

	suite.logger.Info("Employee onboarding workflow completed successfully",
		"workflow_id", we.GetID(),
		"employee_id", employeeID,
		"completed_steps", len(result.CompletedSteps))
}

// TestLeaveApprovalWorkflow_Integration 休假审批工作流集成测试
func TestLeaveApprovalWorkflow_Integration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TearDown()

	// 准备测试数据
	requestID := uuid.New()
	employeeID := uuid.New()
	tenantID := uuid.New()
	managerID := uuid.New()

	request := LeaveApprovalRequest{
		RequestID:   requestID,
		EmployeeID:  employeeID,
		TenantID:    tenantID,
		LeaveType:   "annual",
		StartDate:   time.Now().AddDate(0, 0, 7),
		EndDate:     time.Now().AddDate(0, 0, 10),
		Reason:      "Family vacation",
		ManagerID:   managerID,
		RequestedAt: time.Now(),
	}

	// 启动工作流
	workflowOptions := client.StartWorkflowOptions{
		ID:        "leave-approval-test-" + uuid.New().String(),
		TaskQueue: "test-queue",
	}

	ctx := context.Background()
	we, err := suite.client.ExecuteWorkflow(ctx, workflowOptions, LeaveApprovalWorkflow, request)
	require.NoError(t, err, "Failed to start workflow")

	suite.logger.Info("Started leave approval workflow", "workflow_id", we.GetID())

	// 等待工作流完成（包含模拟的自动审批）
	var result LeaveApprovalResult
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")

	// 验证结果
	assert.Equal(t, requestID, result.RequestID)
	assert.Contains(t, []string{"approved", "rejected"}, result.Status)
	if result.Status == "approved" {
		assert.NotNil(t, result.ApproverID)
		assert.NotNil(t, result.ApprovedAt)
	}

	suite.logger.Info("Leave approval workflow completed",
		"workflow_id", we.GetID(),
		"request_id", requestID,
		"status", result.Status)
}

// TestWorkflowQuery_Integration 工作流查询集成测试
func TestWorkflowQuery_Integration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TearDown()

	// 启动一个长时间运行的工作流用于查询测试
	request := EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "Query",
		LastName:   "Test",
		Email:      "query.test@example.com",
		Department: "Engineering",
		Position:   "Software Engineer",
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "query-test-" + uuid.New().String(),
		TaskQueue: "test-queue",
	}

	ctx := context.Background()
	we, err := suite.client.ExecuteWorkflow(ctx, workflowOptions, EmployeeOnboardingWorkflow, request)
	require.NoError(t, err, "Failed to start workflow")

	// 等待一小段时间让工作流开始执行
	time.Sleep(2 * time.Second)

	// 查询工作流状态
	resp, err := suite.client.QueryWorkflow(ctx, we.GetID(), we.GetRunID(), "workflow_status")
	if err == nil {
		var status interface{}
		err = resp.Get(&status)
		assert.NoError(t, err)
		suite.logger.Info("Workflow status query successful", "status", status)
	}

	// 等待工作流完成
	var result EmployeeOnboardingResult
	err = we.Get(ctx, &result)
	require.NoError(t, err, "Workflow execution failed")
	assert.Equal(t, "completed", result.Status)
}

// TestWorkflowCancellation_Integration 工作流取消集成测试
func TestWorkflowCancellation_Integration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TearDown()

	request := EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "Cancel",
		LastName:   "Test",
		Email:      "cancel.test@example.com",
		Department: "Engineering",
		Position:   "Software Engineer",
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "cancel-test-" + uuid.New().String(),
		TaskQueue: "test-queue",
	}

	ctx := context.Background()
	we, err := suite.client.ExecuteWorkflow(ctx, workflowOptions, EmployeeOnboardingWorkflow, request)
	require.NoError(t, err, "Failed to start workflow")

	// 立即取消工作流
	err = suite.client.CancelWorkflow(ctx, we.GetID(), we.GetRunID())
	require.NoError(t, err, "Failed to cancel workflow")

	// 等待工作流结束
	var result EmployeeOnboardingResult
	err = we.Get(ctx, &result)
	
	// 取消的工作流可能返回错误或特定的结果
	// 这取决于工作流的具体实现
	suite.logger.Info("Workflow cancellation test completed",
		"workflow_id", we.GetID(),
		"error", err)
}

// TestMultipleWorkflows_Integration 多工作流并发集成测试
func TestMultipleWorkflows_Integration(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.TearDown()

	ctx := context.Background()
	numWorkflows := 5
	workflows := make([]client.WorkflowRun, numWorkflows)

	// 启动多个工作流
	for i := 0; i < numWorkflows; i++ {
		request := EmployeeOnboardingRequest{
			EmployeeID: uuid.New(),
			TenantID:   uuid.New(),
			FirstName:  fmt.Sprintf("Parallel%d", i),
			LastName:   "Test",
			Email:      fmt.Sprintf("parallel%d.test@example.com", i),
			Department: "Engineering",
			Position:   "Software Engineer",
			StartDate:  time.Now().AddDate(0, 0, 7),
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        fmt.Sprintf("parallel-test-%d-%s", i, uuid.New().String()),
			TaskQueue: "test-queue",
		}

		we, err := suite.client.ExecuteWorkflow(ctx, workflowOptions, EmployeeOnboardingWorkflow, request)
		require.NoError(t, err, "Failed to start workflow %d", i)
		workflows[i] = we
	}

	// 等待所有工作流完成
	for i, we := range workflows {
		var result EmployeeOnboardingResult
		err := we.Get(ctx, &result)
		require.NoError(t, err, "Workflow %d execution failed", i)
		assert.Equal(t, "completed", result.Status)
	}

	suite.logger.Info("Multiple workflows completed successfully", "count", numWorkflows)
}