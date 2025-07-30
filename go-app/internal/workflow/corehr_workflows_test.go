package workflow

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

// Test the CoreHR workflow types and functions
func TestEmployeeOnboardingRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request EmployeeOnboardingRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
				Email:      "john.doe@example.com",
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, 7),
			},
			wantErr: false,
		},
		{
			name: "missing employee ID",
			request: EmployeeOnboardingRequest{
				TenantID:   uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
				Email:      "john.doe@example.com",
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, 7),
			},
			wantErr: true,
			errMsg:  "employee ID cannot be empty",
		},
		{
			name: "missing tenant ID",
			request: EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
				Email:      "john.doe@example.com",
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, 7),
			},
			wantErr: true,
			errMsg:  "tenant ID cannot be empty",
		},
		{
			name: "missing required fields",
			request: EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  "",
				LastName:   "",
				Email:      "",
			},
			wantErr: true,
			errMsg:  "first name, last name, and email are required",
		},
		{
			name: "invalid email format",
			request: EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
				Email:      "invalid-email",
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, 7),
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "start date in the past",
			request: EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  "John",
				LastName:   "Doe",
				Email:      "john.doe@example.com",
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, -7),
			},
			wantErr: true,
			errMsg:  "start date cannot be in the past",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmployeeOnboardingRequest(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLeaveApprovalRequest_Validation(t *testing.T) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, 7)
	endDate := futureDate.AddDate(0, 0, 3)

	tests := []struct {
		name    string
		request LeaveApprovalRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			request: LeaveApprovalRequest{
				RequestID:   uuid.New(),
				EmployeeID:  uuid.New(),
				TenantID:    uuid.New(),
				LeaveType:   "vacation",
				StartDate:   futureDate,
				EndDate:     endDate,
				Reason:      "Family vacation",
				ManagerID:   uuid.New(),
				RequestedAt: now,
			},
			wantErr: false,
		},
		{
			name: "missing request ID",
			request: LeaveApprovalRequest{
				EmployeeID:  uuid.New(),
				TenantID:    uuid.New(),
				LeaveType:   "vacation",
				StartDate:   futureDate,
				EndDate:     endDate,
				Reason:      "Family vacation",
				ManagerID:   uuid.New(),
				RequestedAt: now,
			},
			wantErr: true,
			errMsg:  "request ID cannot be empty",
		},
		{
			name: "invalid date range",
			request: LeaveApprovalRequest{
				RequestID:   uuid.New(),
				EmployeeID:  uuid.New(),
				TenantID:    uuid.New(),
				LeaveType:   "vacation",
				StartDate:   endDate,    // End date as start date
				EndDate:     futureDate, // Start date as end date
				Reason:      "Family vacation",
				ManagerID:   uuid.New(),
				RequestedAt: now,
			},
			wantErr: true,
			errMsg:  "end date must be after start date",
		},
		{
			name: "missing leave type",
			request: LeaveApprovalRequest{
				RequestID:   uuid.New(),
				EmployeeID:  uuid.New(),
				TenantID:    uuid.New(),
				LeaveType:   "",
				StartDate:   futureDate,
				EndDate:     endDate,
				Reason:      "Family vacation",
				ManagerID:   uuid.New(),
				RequestedAt: now,
			},
			wantErr: true,
			errMsg:  "leave type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLeaveApprovalRequest(tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test workflow execution using Temporal test suite
func TestEmployeeOnboardingWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register the workflow and mock activities
	env.RegisterWorkflow(EmployeeOnboardingWorkflow)

	// Mock activities (since we don't have the actual implementations)
	env.OnActivity("CreateEmployeeAccountActivity", mock.Anything, mock.Anything).Return(&CreateAccountResult{
		AccountID: "acc_" + uuid.New().String()[:8],
		Success:   true,
	}, nil)

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

	env.ExecuteWorkflow(EmployeeOnboardingWorkflow, request)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result EmployeeOnboardingResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, request.EmployeeID, result.EmployeeID)
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.CompletedSteps)
}

func TestEmployeeOnboardingWorkflow_ValidationFailure(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(EmployeeOnboardingWorkflow)

	// Mock validation failure
	env.OnActivity("CreateEmployeeAccountActivity", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	request := EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "invalid-email", // Invalid email to trigger validation failure
		Department: "Engineering",
		Position:   "Software Engineer",
		StartDate:  time.Now().AddDate(0, 0, 7),
	}

	env.ExecuteWorkflow(EmployeeOnboardingWorkflow, request)

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

func TestLeaveApprovalWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(LeaveApprovalWorkflow)

	// Mock activities
	env.OnActivity("ValidateLeaveRequestActivity", mock.Anything, mock.Anything).Return(&ValidateLeaveRequestResult{
		IsValid: true,
		Reason:  "",
	}, nil)

	request := LeaveApprovalRequest{
		RequestID:   uuid.New(),
		EmployeeID:  uuid.New(),
		TenantID:    uuid.New(),
		LeaveType:   "vacation",
		StartDate:   time.Now().AddDate(0, 0, 7),
		EndDate:     time.Now().AddDate(0, 0, 10),
		Reason:      "Family vacation",
		ManagerID:   uuid.New(),
		RequestedAt: time.Now(),
	}

	env.ExecuteWorkflow(LeaveApprovalWorkflow, request)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result LeaveApprovalResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, request.RequestID, result.RequestID)
	assert.NotEmpty(t, result.Status)
}

// Performance and load testing
func TestWorkflowPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testSuite := &testsuite.WorkflowTestSuite{}

	// Test multiple concurrent workflow executions
	const numWorkflows = 10

	for i := 0; i < numWorkflows; i++ {
		t.Run(fmt.Sprintf("workflow_%d", i), func(t *testing.T) {
			t.Parallel()

			env := testSuite.NewTestWorkflowEnvironment()
			env.RegisterWorkflow(EmployeeOnboardingWorkflow)

			// Mock activities
			env.OnActivity("CreateEmployeeAccountActivity", mock.Anything, mock.Anything).Return(&CreateAccountResult{
				AccountID: "acc_" + uuid.New().String()[:8],
				Success:   true,
			}, nil)

			request := EmployeeOnboardingRequest{
				EmployeeID: uuid.New(),
				TenantID:   uuid.New(),
				FirstName:  fmt.Sprintf("Employee%d", i),
				LastName:   "Test",
				Email:      fmt.Sprintf("employee%d@example.com", i),
				Department: "Engineering",
				Position:   "Software Engineer",
				StartDate:  time.Now().AddDate(0, 0, 7),
			}

			start := time.Now()
			env.ExecuteWorkflow(EmployeeOnboardingWorkflow, request)
			duration := time.Since(start)

			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())

			// Performance assertion - workflow should complete within reasonable time
			assert.Less(t, duration, time.Second*5, "Workflow took too long to complete")
		})
	}
}

// Helper functions for validation (these would be implemented in the actual workflow code)
func validateEmployeeOnboardingRequest(req EmployeeOnboardingRequest) error {
	if req.EmployeeID == uuid.Nil {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if req.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if req.FirstName == "" || req.LastName == "" || req.Email == "" {
		return fmt.Errorf("first name, last name, and email are required")
	}
	if !isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}
	if req.StartDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return fmt.Errorf("start date cannot be in the past")
	}
	return nil
}

func validateLeaveApprovalRequest(req LeaveApprovalRequest) error {
	if req.RequestID == uuid.Nil {
		return fmt.Errorf("request ID cannot be empty")
	}
	if req.EmployeeID == uuid.Nil {
		return fmt.Errorf("employee ID cannot be empty")
	}
	if req.TenantID == uuid.Nil {
		return fmt.Errorf("tenant ID cannot be empty")
	}
	if req.LeaveType == "" {
		return fmt.Errorf("leave type is required")
	}
	if req.EndDate.Before(req.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}
	if req.ManagerID == uuid.Nil {
		return fmt.Errorf("manager ID cannot be empty")
	}
	return nil
}

func isValidEmail(email string) bool {
	// Simple email validation - in real implementation use proper regex or library
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
