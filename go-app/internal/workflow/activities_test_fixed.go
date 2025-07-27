package workflow

import (
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// TestCreateEmployeeAccountActivityFixed 正确的Temporal Activity测试
func TestCreateEmployeeAccountActivityFixed(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()
	
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	// 注册Activity
	env.RegisterActivity(activities.CreateEmployeeAccountActivity)

	// 准备测试数据
	createAccountReq := CreateAccountRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
	}

	// ✅ 正确方式：通过Temporal测试环境执行Activity
	encodedValue, err := env.ExecuteActivity(activities.CreateEmployeeAccountActivity, createAccountReq)
	require.NoError(t, err)

	var result CreateAccountResult
	err = encodedValue.Get(&result)
	require.NoError(t, err)

	// 验证结果
	assert.NotEmpty(t, result.AccountID)
	assert.True(t, result.Success)
	assert.Contains(t, result.AccountID, "acc_") // 验证ID格式
}

// TestWorkflowActivityIntegration 测试工作流与Activity的集成
func TestWorkflowActivityIntegration(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	// 注册工作流和所有Activity
	env.RegisterWorkflow(EmployeeOnboardingWorkflow)
	env.RegisterActivity(activities.CreateEmployeeAccountActivity)
	env.RegisterActivity(activities.AssignEquipmentAndPermissionsActivity)
	env.RegisterActivity(activities.SendWelcomeEmailActivity)
	env.RegisterActivity(activities.NotifyManagerActivity)

	// 准备测试数据
	request := EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john.doe@example.com",
		Department: "Engineering",
		Position:   "Software Engineer",
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	// 执行工作流
	env.ExecuteWorkflow(EmployeeOnboardingWorkflow, request)

	// 验证工作流完成
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// 获取结果
	var result EmployeeOnboardingResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// 验证结果
	assert.Equal(t, request.EmployeeID, result.EmployeeID)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.CompletedSteps)
}

// TestActivityMocking 测试Activity模拟
func TestActivityMocking(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// 注册工作流
	env.RegisterWorkflow(EmployeeOnboardingWorkflow)

	// 模拟所有Activities的返回值
	env.OnActivity(CreateEmployeeAccountActivity, mock.Anything, mock.Anything).Return(&CreateAccountResult{
		AccountID: "mocked-account-123",
		Success:   true,
	}, nil)

	env.OnActivity(AssignEquipmentAndPermissionsActivity, mock.Anything, mock.Anything).Return(&AssignEquipmentResult{
		AssignedItems: []string{"laptop", "mouse", "keyboard"},
		Success:       true,
	}, nil)

	env.OnActivity(SendWelcomeEmailActivity, mock.Anything, mock.Anything).Return(&SendEmailResult{
		MessageID: "mock-email-456",
		Success:   true,
	}, nil)

	env.OnActivity(NotifyManagerActivity, mock.Anything, mock.Anything).Return(&NotifyManagerResult{
		NotificationID: "mock-notification-789",
		Success:        true,
	}, nil)

	// 准备测试数据
	managerID := uuid.New()
	request := EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "Jane",
		LastName:   "Smith",
		Email:      "jane.smith@example.com",
		Department: "Marketing",
		Position:   "Marketing Manager",
		ManagerID:  &managerID,
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	// 执行工作流
	env.ExecuteWorkflow(EmployeeOnboardingWorkflow, request)

	// 验证结果
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result EmployeeOnboardingResult
	require.NoError(t, env.GetWorkflowResult(&result))

	// 验证模拟数据被正确使用
	assert.Equal(t, "completed", result.Status)
	assert.Contains(t, result.CompletedSteps, "account_created")
	assert.Contains(t, result.CompletedSteps, "equipment_assigned")
	assert.Contains(t, result.CompletedSteps, "welcome_email_sent")
	assert.Contains(t, result.CompletedSteps, "manager_notified")
}