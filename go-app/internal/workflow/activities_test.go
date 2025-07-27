package workflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// TestCreateEmployeeAccountActivity 测试创建员工账户活动
func TestCreateEmployeeAccountActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name      string
		request   CreateAccountRequest
		expectErr bool
	}{
		{
			name: "valid account creation",
			request: CreateAccountRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Email:      "john.doe@example.com",
				FirstName:  "John",
				LastName:   "Doe",
			},
			expectErr: false,
		},
		{
			name: "valid account with different user",
			request: CreateAccountRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Email:      "jane.smith@example.com",
				FirstName:  "Jane",
				LastName:   "Smith",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.CreateEmployeeAccountActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.AccountID)
				assert.Contains(t, result.AccountID, "acc_")
			}
		})
	}
}

// TestAssignEquipmentAndPermissionsActivity 测试分配设备和权限活动
func TestAssignEquipmentAndPermissionsActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name          string
		request       AssignEquipmentRequest
		expectErr     bool
		expectedItems []string
	}{
		{
			name: "技术部员工",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "技术部",
				Position:   "Developer",
			},
			expectErr:     false,
			expectedItems: []string{"laptop", "monitor", "keyboard", "mouse"},
		},
		{
			name: "Senior Developer with extra monitor",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Technology",
				Position:   "Senior Developer",
			},
			expectErr:     false,
			expectedItems: []string{"laptop", "monitor", "keyboard", "mouse", "additional_monitor"},
		},
		{
			name: "销售部员工",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "销售部",
				Position:   "Sales Representative",
			},
			expectErr:     false,
			expectedItems: []string{"laptop", "mobile_phone"},
		},
		{
			name: "人事部员工",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "人事部",
				Position:   "HR Specialist",
			},
			expectErr:     false,
			expectedItems: []string{"laptop", "printer_access"},
		},
		{
			name: "其他部门默认设备",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Finance",
				Position:   "Accountant",
			},
			expectErr:     false,
			expectedItems: []string{"laptop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.AssignEquipmentAndPermissionsActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, tt.expectedItems, result.AssignedItems)
			}
		})
	}
}

// TestSendWelcomeEmailActivity 测试发送欢迎邮件活动
func TestSendWelcomeEmailActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name      string
		request   WelcomeEmailRequest
		expectErr bool
	}{
		{
			name: "valid welcome email",
			request: WelcomeEmailRequest{
				EmployeeID: uuid.New(),
				Email:      "john.doe@example.com",
				FirstName:  "John",
				StartDate:  time.Now().AddDate(0, 0, 7),
				Department: "Engineering",
			},
			expectErr: false,
		},
		{
			name: "welcome email for different department",
			request: WelcomeEmailRequest{
				EmployeeID: uuid.New(),
				Email:      "jane.smith@example.com",
				FirstName:  "Jane",
				StartDate:  time.Now().AddDate(0, 0, 14),
				Department: "Sales",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.SendWelcomeEmailActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.MessageID)
				assert.Contains(t, result.MessageID, "msg_")
			}
		})
	}
}

// TestNotifyManagerActivity 测试通知经理活动
func TestNotifyManagerActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name      string
		request   NotifyManagerRequest
		expectErr bool
	}{
		{
			name: "valid manager notification",
			request: NotifyManagerRequest{
				ManagerID:     uuid.New(),
				NewEmployeeID: uuid.New(),
				EmployeeName:  "John Doe",
				StartDate:     time.Now().AddDate(0, 0, 7),
				Department:    "Engineering",
				Position:      "Software Engineer",
			},
			expectErr: false,
		},
		{
			name: "notification for different role",
			request: NotifyManagerRequest{
				ManagerID:     uuid.New(),
				NewEmployeeID: uuid.New(),
				EmployeeName:  "Jane Smith",
				StartDate:     time.Now().AddDate(0, 0, 14),
				Department:    "Marketing",
				Position:      "Marketing Specialist",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.NotifyManagerActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.NotificationID)
				assert.Contains(t, result.NotificationID, "notif_")
			}
		})
	}
}

// TestValidateLeaveRequestActivity 测试验证休假请求活动
func TestValidateLeaveRequestActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name           string
		request        ValidateLeaveRequestRequest
		expectErr      bool
		expectedValid  bool
		expectedReason string
	}{
		{
			name: "valid leave request",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  time.Now().AddDate(0, 0, 7),
				EndDate:    time.Now().AddDate(0, 0, 10),
			},
			expectErr:     false,
			expectedValid: true,
		},
		{
			name: "invalid date range - start after end",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  time.Now().AddDate(0, 0, 10),
				EndDate:    time.Now().AddDate(0, 0, 7),
			},
			expectErr:      false,
			expectedValid:  false,
			expectedReason: "Start date must be before end date",
		},
		{
			name: "past start date",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  time.Now().AddDate(0, 0, -7),
				EndDate:    time.Now().AddDate(0, 0, -3),
			},
			expectErr:      false,
			expectedValid:  false,
			expectedReason: "Cannot request leave for past dates",
		},
		{
			name: "invalid leave type",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "invalid_type",
				StartDate:  time.Now().AddDate(0, 0, 7),
				EndDate:    time.Now().AddDate(0, 0, 10),
			},
			expectErr:      false,
			expectedValid:  false,
			expectedReason: "Invalid leave type",
		},
		{
			name: "leave duration too long",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  time.Now().AddDate(0, 0, 7),
				EndDate:    time.Now().AddDate(0, 0, 40), // 33 days
			},
			expectErr:      false,
			expectedValid:  false,
			expectedReason: "Leave duration cannot exceed 30 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.ValidateLeaveRequestActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedValid, result.IsValid)
				if !tt.expectedValid {
					assert.Equal(t, tt.expectedReason, result.Reason)
				}
			}
		})
	}
}

// TestProcessSingleEmployeeActivity 测试单个员工处理活动
func TestProcessSingleEmployeeActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name           string
		request        ProcessSingleEmployeeRequest
		expectErr      bool
		expectedStatus string
	}{
		{
			name: "valid employee onboard",
			request: ProcessSingleEmployeeRequest{
				BatchID:    uuid.New(),
				TenantID:   uuid.New(),
				Operation:  "onboard",
				EmployeeID: uuid.New(),
				Data: map[string]interface{}{
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john.doe@example.com",
					"department": "Engineering",
					"position":   "Software Engineer",
				},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "success",
		},
		{
			name: "valid employee offboard",
			request: ProcessSingleEmployeeRequest{
				BatchID:    uuid.New(),
				TenantID:   uuid.New(),
				Operation:  "offboard",
				EmployeeID: uuid.New(),
				Data: map[string]interface{}{
					"last_working_day": "2024-12-31",
					"reason":           "resignation",
				},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "success",
		},
		{
			name: "valid employee update",
			request: ProcessSingleEmployeeRequest{
				BatchID:    uuid.New(),
				TenantID:   uuid.New(),
				Operation:  "update",
				EmployeeID: uuid.New(),
				Data: map[string]interface{}{
					"department": "Marketing",
					"position":   "Marketing Manager",
				},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "success",
		},
		{
			name: "onboard with missing required fields",
			request: ProcessSingleEmployeeRequest{
				BatchID:    uuid.New(),
				TenantID:   uuid.New(),
				Operation:  "onboard",
				EmployeeID: uuid.New(),
				Data: map[string]interface{}{
					"first_name": "John",
					// Missing last_name and email
				},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "failed",
		},
		{
			name: "offboard with missing required fields",
			request: ProcessSingleEmployeeRequest{
				BatchID:    uuid.New(),
				TenantID:   uuid.New(),
				Operation:  "offboard",
				EmployeeID: uuid.New(),
				Data: map[string]interface{}{
					// Missing last_working_day
					"reason": "resignation",
				},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "failed",
		},
		{
			name: "unknown operation",
			request: ProcessSingleEmployeeRequest{
				BatchID:     uuid.New(),
				TenantID:    uuid.New(),
				Operation:   "unknown_operation",
				EmployeeID:  uuid.New(),
				Data:        map[string]interface{}{},
				RequestedBy: uuid.New(),
			},
			expectErr:      false,
			expectedStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.ProcessSingleEmployeeActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.request.EmployeeID, result.EmployeeID)
				assert.Equal(t, tt.expectedStatus, result.Status)

				if tt.expectedStatus == "failed" {
					assert.NotEmpty(t, result.ErrorMessage)
				}
			}
		})
	}
}

// TestWaitForManagerApprovalActivity 测试等待经理审批活动
func TestWaitForManagerApprovalActivity(t *testing.T) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	tests := []struct {
		name      string
		request   WaitForManagerApprovalRequest
		expectErr bool
	}{
		{
			name: "valid approval wait",
			request: WaitForManagerApprovalRequest{
				RequestID:    uuid.New(),
				ManagerID:    uuid.New(),
				TimeoutHours: 168, // 7 days
			},
			expectErr: false,
		},
		{
			name: "shorter timeout",
			request: WaitForManagerApprovalRequest{
				RequestID:    uuid.New(),
				ManagerID:    uuid.New(),
				TimeoutHours: 24, // 1 day
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := activities.WaitForManagerApprovalActivity(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.ManagerID, result.ApproverID)
				assert.Contains(t, []string{"approved", "rejected"}, result.Decision)
				assert.NotEmpty(t, result.Comments)
				assert.False(t, result.ApprovedAt.IsZero())
			}
		})
	}
}

// Performance benchmark tests
func BenchmarkCreateEmployeeAccountActivity(b *testing.B) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)
	ctx := context.Background()

	request := CreateAccountRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		Email:      "benchmark@example.com",
		FirstName:  "Benchmark",
		LastName:   "User",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := activities.CreateEmployeeAccountActivity(ctx, request)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkValidateLeaveRequestActivity(b *testing.B) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)
	ctx := context.Background()

	request := ValidateLeaveRequestRequest{
		RequestID:  uuid.New(),
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		LeaveType:  "annual",
		StartDate:  time.Now().AddDate(0, 0, 7),
		EndDate:    time.Now().AddDate(0, 0, 10),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := activities.ValidateLeaveRequestActivity(ctx, request)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkProcessSingleEmployeeActivity(b *testing.B) {
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)
	ctx := context.Background()

	request := ProcessSingleEmployeeRequest{
		BatchID:    uuid.New(),
		TenantID:   uuid.New(),
		Operation:  "update",
		EmployeeID: uuid.New(),
		Data: map[string]interface{}{
			"department": "Engineering",
			"position":   "Software Engineer",
		},
		RequestedBy: uuid.New(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := activities.ProcessSingleEmployeeActivity(ctx, request)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// Integration test with Temporal test suite
func TestActivities_TemporalIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping temporal integration test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register all activities
	logger := logging.NewStructuredLogger()
	activities := NewActivities(logger)

	env.RegisterActivity(activities.CreateEmployeeAccountActivity)
	env.RegisterActivity(activities.AssignEquipmentAndPermissionsActivity)
	env.RegisterActivity(activities.SendWelcomeEmailActivity)
	env.RegisterActivity(activities.NotifyManagerActivity)
	env.RegisterActivity(activities.ValidateLeaveRequestActivity)
	env.RegisterActivity(activities.ProcessSingleEmployeeActivity)

	// Test specific activity execution
	env.OnActivity(activities.CreateEmployeeAccountActivity, mock.Anything, mock.Anything).Return(&CreateAccountResult{
		AccountID: "test-account-123",
		Success:   true,
	}, nil)

	// Execute activity directly through test environment
	createAccountReq := CreateAccountRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
	}

	var result CreateAccountResult
	err := env.ExecuteActivity(activities.CreateEmployeeAccountActivity, createAccountReq).Get(&result)

	assert.NoError(t, err)
	assert.Equal(t, "test-account-123", result.AccountID)
	assert.True(t, result.Success)
}
