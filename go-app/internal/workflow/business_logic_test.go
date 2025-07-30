package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBusinessLogic_CreateEmployeeAccount 测试员工账户创建业务逻辑
func TestBusinessLogic_CreateEmployeeAccount(t *testing.T) {
	logger := logging.NewStructuredLogger()
	bl := NewBusinessLogic(logger)

	tests := []struct {
		name      string
		request   CreateAccountRequest
		expectErr bool
		errMsg    string
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
			name: "missing email",
			request: CreateAccountRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
			},
			expectErr: true,
			errMsg:    "email is required",
		},
		{
			name: "missing first name",
			request: CreateAccountRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Email:      "john.doe@example.com",
				LastName:   "Doe",
			},
			expectErr: true,
			errMsg:    "first name is required",
		},
		{
			name: "missing last name",
			request: CreateAccountRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Email:      "john.doe@example.com",
				FirstName:  "John",
			},
			expectErr: true,
			errMsg:    "last name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := bl.CreateEmployeeAccount(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.AccountID)
				assert.Contains(t, result.AccountID, "acc_")
			}
		})
	}
}

// TestBusinessLogic_AssignEquipmentAndPermissions 测试设备分配业务逻辑
func TestBusinessLogic_AssignEquipmentAndPermissions(t *testing.T) {
	logger := logging.NewStructuredLogger()
	bl := NewBusinessLogic(logger)

	tests := []struct {
		name          string
		request       AssignEquipmentRequest
		expectedItems []string
		expectErr     bool
		errMsg        string
	}{
		{
			name: "technology department standard",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Technology",
				Position:   "Software Engineer",
			},
			expectedItems: []string{"laptop", "monitor", "keyboard", "mouse"},
			expectErr:     false,
		},
		{
			name: "technology department senior",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Engineering",
				Position:   "Senior Developer",
			},
			expectedItems: []string{"laptop", "monitor", "keyboard", "mouse", "additional_monitor"},
			expectErr:     false,
		},
		{
			name: "sales department",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Sales",
				Position:   "Sales Representative",
			},
			expectedItems: []string{"laptop", "mobile_phone"},
			expectErr:     false,
		},
		{
			name: "hr department",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "HR",
				Position:   "HR Specialist",
			},
			expectedItems: []string{"laptop", "printer_access"},
			expectErr:     false,
		},
		{
			name: "unknown department",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Finance",
				Position:   "Accountant",
			},
			expectedItems: []string{"laptop"},
			expectErr:     false,
		},
		{
			name: "missing department",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Position:   "Software Engineer",
			},
			expectErr: true,
			errMsg:    "department is required",
		},
		{
			name: "missing position",
			request: AssignEquipmentRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				Department: "Technology",
			},
			expectErr: true,
			errMsg:    "position is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := bl.AssignEquipmentAndPermissions(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, tt.expectedItems, result.AssignedItems)
			}
		})
	}
}

// TestBusinessLogic_SendWelcomeEmail 测试欢迎邮件发送业务逻辑
func TestBusinessLogic_SendWelcomeEmail(t *testing.T) {
	logger := logging.NewStructuredLogger()
	bl := NewBusinessLogic(logger)

	tests := []struct {
		name      string
		request   WelcomeEmailRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid email request",
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
			name: "missing email",
			request: WelcomeEmailRequest{
				EmployeeID: uuid.New(),
				FirstName:  "John",
				StartDate:  time.Now().AddDate(0, 0, 7),
				Department: "Engineering",
			},
			expectErr: true,
			errMsg:    "email is required",
		},
		{
			name: "missing first name",
			request: WelcomeEmailRequest{
				EmployeeID: uuid.New(),
				Email:      "john.doe@example.com",
				StartDate:  time.Now().AddDate(0, 0, 7),
				Department: "Engineering",
			},
			expectErr: true,
			errMsg:    "first name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := bl.SendWelcomeEmail(ctx, tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.Success)
				assert.NotEmpty(t, result.MessageID)
				assert.Contains(t, result.MessageID, "msg_")
			}
		})
	}
}

// TestBusinessLogic_ValidateLeaveRequest 测试休假请求验证业务逻辑
func TestBusinessLogic_ValidateLeaveRequest(t *testing.T) {
	logger := logging.NewStructuredLogger()
	bl := NewBusinessLogic(logger)

	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name      string
		request   ValidateLeaveRequestRequest
		expectValid bool
		expectedReason string
	}{
		{
			name: "valid annual leave",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  tomorrow,
				EndDate:    nextWeek,
			},
			expectValid: true,
		},
		{
			name: "invalid - start after end",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  nextWeek,
				EndDate:    tomorrow,
			},
			expectValid: false,
			expectedReason: "Start date must be before end date",
		},
		{
			name: "invalid - past date",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  yesterday,
				EndDate:    tomorrow,
			},
			expectValid: false,
			expectedReason: "Cannot request leave for past dates",
		},
		{
			name: "invalid - leave type",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "invalid_type",
				StartDate:  tomorrow,
				EndDate:    nextWeek,
			},
			expectValid: false,
			expectedReason: "Invalid leave type",
		},
		{
			name: "invalid - too long",
			request: ValidateLeaveRequestRequest{
				RequestID:  uuid.New(),
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				LeaveType:  "annual",
				StartDate:  tomorrow,
				EndDate:    tomorrow.AddDate(0, 0, 35), // 35 days
			},
			expectValid: false,
			expectedReason: "Leave duration cannot exceed 30 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := bl.ValidateLeaveRequest(ctx, tt.request)

			assert.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectValid, result.IsValid)
			if !tt.expectValid {
				assert.Equal(t, tt.expectedReason, result.Reason)
			}
		})
	}
}

// TestBusinessLogic_Integration 集成测试
func TestBusinessLogic_Integration(t *testing.T) {
	logger := logging.NewStructuredLogger()
	bl := NewBusinessLogic(logger)
	ctx := context.Background()

	// 完整的员工入职流程测试
	employeeID := uuid.New()
	tenantID := uuid.New()
	
	// 1. 创建账户
	accountResult, err := bl.CreateEmployeeAccount(ctx, CreateAccountRequest{
		EmployeeID: employeeID,
		TenantID:   tenantID,
		Email:      "integration.test@example.com",
		FirstName:  "Integration",
		LastName:   "Test",
	})
	require.NoError(t, err)
	assert.True(t, accountResult.Success)

	// 2. 分配设备
	equipmentResult, err := bl.AssignEquipmentAndPermissions(ctx, AssignEquipmentRequest{
		EmployeeID: employeeID,
		TenantID:   tenantID,
		Department: "Engineering",
		Position:   "Senior Developer",
	})
	require.NoError(t, err)
	assert.True(t, equipmentResult.Success)
	assert.Contains(t, equipmentResult.AssignedItems, "additional_monitor")

	// 3. 发送欢迎邮件
	emailResult, err := bl.SendWelcomeEmail(ctx, WelcomeEmailRequest{
		EmployeeID: employeeID,
		Email:      "integration.test@example.com",
		FirstName:  "Integration",
		StartDate:  time.Now().AddDate(0, 0, 7),
		Department: "Engineering",
	})
	require.NoError(t, err)
	assert.True(t, emailResult.Success)

	// 4. 通知经理
	managerID := uuid.New()
	notificationResult, err := bl.NotifyManager(ctx, NotifyManagerRequest{
		ManagerID:      managerID,
		NewEmployeeID:  employeeID,
		EmployeeName:   "Integration Test",
		StartDate:      time.Now().AddDate(0, 0, 7),
		Department:     "Engineering",
		Position:       "Senior Developer",
	})
	require.NoError(t, err)
	assert.True(t, notificationResult.Success)

	t.Log("✅ Complete employee onboarding flow test passed")
}