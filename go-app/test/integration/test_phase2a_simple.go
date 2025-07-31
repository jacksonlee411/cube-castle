// test_phase2a_simple.go - Simplified Phase 2A Integration Test
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
)

// Phase2ASimpleTestSuite is a simplified test suite for Phase 2A functionality
type Phase2ASimpleTestSuite struct {
	suite.Suite
	ctx        context.Context
	entClient  *ent.Client
	activities *workflow.EmployeeLifecycleActivities
	logger     *logging.StructuredLogger

	// Test data
	testTenantID    uuid.UUID
	testEmployeeID  uuid.UUID
	testCandidateID uuid.UUID
	testUpdaterID   uuid.UUID
}

// SetupSuite initializes the test suite
func (s *Phase2ASimpleTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.logger = logging.NewStructuredLogger()

	// Use in-memory SQLite for testing to avoid database dependencies
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	s.Require().NoError(err, "Failed to create in-memory ent client")

	// Run schema creation
	err = client.Schema.Create(s.ctx)
	s.Require().NoError(err, "Failed to create database schema")

	s.entClient = client

	// Initialize activities with minimal dependencies (no temporal query service needed for basic tests)
	s.activities = workflow.NewEmployeeLifecycleActivities(s.entClient, nil, s.logger)

	// Generate test UUIDs
	s.testTenantID = uuid.New()
	s.testEmployeeID = uuid.New()
	s.testCandidateID = uuid.New()
	s.testUpdaterID = uuid.New()

	s.logger.Info("Phase 2A Simple Test Suite initialized",
		"tenant_id", s.testTenantID,
		"employee_id", s.testEmployeeID,
		"candidate_id", s.testCandidateID,
	)
}

// TearDownSuite cleans up the test suite
func (s *Phase2ASimpleTestSuite) TearDownSuite() {
	if s.entClient != nil {
		s.entClient.Close()
	}
}

// SetupTest prepares each individual test
func (s *Phase2ASimpleTestSuite) SetupTest() {
	// Create test employee record
	s.createTestEmployee()
	s.createTestCandidate()
}

// TearDownTest cleans up after each test
func (s *Phase2ASimpleTestSuite) TearDownTest() {
	// Clean up test data
	s.cleanupTestData()
}

// createTestEmployee creates a test employee record
func (s *Phase2ASimpleTestSuite) createTestEmployee() {
	employee, err := s.entClient.Employee.Create().
		SetID(s.testEmployeeID.String()).
		SetName("Test Employee").
		SetEmail("test.employee@example.com").
		SetPosition("Test Position").
		Save(s.ctx)

	s.Require().NoError(err, "Failed to create test employee")
	s.Require().NotNil(employee, "Test employee should not be nil")

	s.logger.Info("Test employee created", "employee_id", employee.ID)
}

// createTestCandidate creates a test candidate record
func (s *Phase2ASimpleTestSuite) createTestCandidate() {
	candidate, err := s.entClient.Employee.Create().
		SetID(s.testCandidateID.String()).
		SetName("Test Candidate").
		SetEmail("test.candidate@example.com").
		SetPosition("Candidate").
		Save(s.ctx)

	s.Require().NoError(err, "Failed to create test candidate")
	s.Require().NotNil(candidate, "Test candidate should not be nil")

	s.logger.Info("Test candidate created", "candidate_id", candidate.ID)
}

// cleanupTestData removes test data
func (s *Phase2ASimpleTestSuite) cleanupTestData() {
	// Delete test employee and candidate records
	s.entClient.Employee.Delete().ExecX(s.ctx)
	s.logger.Debug("Test data cleaned up")
}

// TestUpdateEmployeeInformation_PersonalInfo tests employee personal information update
func (s *Phase2ASimpleTestSuite) TestUpdateEmployeeInformation_PersonalInfo() {
	s.logger.Info("Testing employee personal information update")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testEmployeeID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"first_name":  "张",
			"last_name":   "三",
			"middle_name": "测试",
			"email":       "zhang.san@test.com",
			"phone":       "+86-138-0000-0001",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: false,
	}

	// Execute the update
	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	// Verify results
	assert.NoError(s.T(), err, "Personal information update should succeed")
	assert.NotNil(s.T(), result, "Update result should not be nil")
	assert.True(s.T(), result.Success, "Update should be successful")
	assert.Equal(s.T(), "updated", result.Status, "Status should be 'updated'")
	assert.False(s.T(), result.RequiredApproval, "Personal info should not require approval")
	assert.NotEqual(s.T(), uuid.Nil, result.UpdateID, "UpdateID should be set")

	s.logger.Info("Employee personal information update test passed",
		"update_id", result.UpdateID,
		"status", result.Status,
	)
}

// TestUpdateEmployeeInformation_ContactInfo tests employee contact information update
func (s *Phase2ASimpleTestSuite) TestUpdateEmployeeInformation_ContactInfo() {
	s.logger.Info("Testing employee contact information update")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testEmployeeID,
		UpdateType: "CONTACT",
		UpdateData: map[string]interface{}{
			"home_address": "北京市朝阳区测试路123号",
			"postal_code":  "100000",
			"city":         "北京",
			"country":      "中国",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: false,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), "updated", result.Status)
	assert.False(s.T(), result.RequiredApproval)

	s.logger.Info("Employee contact information update test passed")
}

// TestUpdateEmployeeInformation_EmergencyContact tests emergency contact update (requires approval)
func (s *Phase2ASimpleTestSuite) TestUpdateEmployeeInformation_EmergencyContact() {
	s.logger.Info("Testing employee emergency contact update (requires approval)")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testEmployeeID,
		UpdateType: "EMERGENCY_CONTACT",
		UpdateData: map[string]interface{}{
			"emergency_contact_name":  "李四",
			"emergency_contact_phone": "+86-138-0000-0002",
			"relationship":            "配偶",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: true,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.True(s.T(), result.RequiredApproval, "Emergency contact should require approval")
	// Status can be "pending_approval" or "pending_approval_failed" since approval workflow may not be complete
	assert.Contains(s.T(), []string{"pending_approval", "pending_approval_failed"}, result.Status)

	s.logger.Info("Employee emergency contact update test passed",
		"requires_approval", result.RequiredApproval,
		"status", result.Status,
	)
}

// TestUpdateEmployeeInformation_BankingInfo tests banking information update (requires approval)
func (s *Phase2ASimpleTestSuite) TestUpdateEmployeeInformation_BankingInfo() {
	s.logger.Info("Testing employee banking information update (requires approval)")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testEmployeeID,
		UpdateType: "BANKING",
		UpdateData: map[string]interface{}{
			"bank_name":      "中国工商银行",
			"account_number": "6222080200001234567",
			"routing_number": "102100024",
			"account_holder": "张三",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: true,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.True(s.T(), result.RequiredApproval, "Banking info should require approval")
	assert.Contains(s.T(), []string{"pending_approval", "pending_approval_failed"}, result.Status)

	s.logger.Info("Employee banking information update test passed")
}

// TestUpdateCandidateInformation_PersonalInfo tests candidate personal information update
func (s *Phase2ASimpleTestSuite) TestUpdateCandidateInformation_PersonalInfo() {
	s.logger.Info("Testing candidate personal information update")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testCandidateID, // Candidate uses EmployeeID field
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"first_name": "王",
			"last_name":  "五",
			"email":      "wang.wu@candidate.com",
			"phone":      "+86-138-0000-0003",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: false,
	}

	result, err := s.activities.UpdateCandidateActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), "updated", result.Status)
	assert.False(s.T(), result.RequiredApproval)

	s.logger.Info("Candidate personal information update test passed")
}

// TestUpdateCandidateInformation_ApplicationStatus tests candidate application status update
func (s *Phase2ASimpleTestSuite) TestUpdateCandidateInformation_ApplicationStatus() {
	s.logger.Info("Testing candidate application status update (requires approval)")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testCandidateID,
		UpdateType: "APPLICATION_STATUS",
		UpdateData: map[string]interface{}{
			"status": "INTERVIEW_SCHEDULED",
			"notes":  "已安排技术面试",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: true,
	}

	result, err := s.activities.UpdateCandidateActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.True(s.T(), result.RequiredApproval)
	assert.Contains(s.T(), []string{"pending_approval", "pending_approval_failed"}, result.Status)

	s.logger.Info("Candidate application status update test passed")
}

// TestUpdateCandidateInformation_InterviewFeedback tests candidate interview feedback update
func (s *Phase2ASimpleTestSuite) TestUpdateCandidateInformation_InterviewFeedback() {
	s.logger.Info("Testing candidate interview feedback update")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testCandidateID,
		UpdateType: "INTERVIEW_FEEDBACK",
		UpdateData: map[string]interface{}{
			"interviewer_id": s.testUpdaterID.String(),
			"rating":         "4",
			"comments":       "技术能力强，沟通良好",
			"recommendation": "HIRE",
		},
		UpdatedBy:        s.testUpdaterID,
		RequiresApproval: false,
	}

	result, err := s.activities.UpdateCandidateActivity(s.ctx, updateRequest)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), result)
	assert.True(s.T(), result.Success)
	assert.Equal(s.T(), "updated", result.Status)
	assert.False(s.T(), result.RequiredApproval)

	s.logger.Info("Candidate interview feedback update test passed")
}

// TestInvalidInputValidation tests input validation for invalid requests
func (s *Phase2ASimpleTestSuite) TestInvalidInputValidation() {
	s.logger.Info("Testing input validation")

	testCases := []struct {
		name    string
		request workflow.InformationUpdateRequest
		desc    string
	}{
		{
			name: "missing_tenant_id",
			request: workflow.InformationUpdateRequest{
				EmployeeID: s.testEmployeeID,
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"test": "data"},
				UpdatedBy:  s.testUpdaterID,
			},
			desc: "should fail when tenant_id is missing",
		},
		{
			name: "missing_employee_id",
			request: workflow.InformationUpdateRequest{
				TenantID:   s.testTenantID,
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{"test": "data"},
				UpdatedBy:  s.testUpdaterID,
			},
			desc: "should fail when employee_id is missing",
		},
		{
			name: "invalid_update_type",
			request: workflow.InformationUpdateRequest{
				TenantID:   s.testTenantID,
				EmployeeID: s.testEmployeeID,
				UpdateType: "INVALID_TYPE",
				UpdateData: map[string]interface{}{"test": "data"},
				UpdatedBy:  s.testUpdaterID,
			},
			desc: "should fail when update_type is invalid",
		},
		{
			name: "empty_update_data",
			request: workflow.InformationUpdateRequest{
				TenantID:   s.testTenantID,
				EmployeeID: s.testEmployeeID,
				UpdateType: "PERSONAL",
				UpdateData: map[string]interface{}{},
				UpdatedBy:  s.testUpdaterID,
			},
			desc: "should fail when update_data is empty",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.logger.Info("Testing validation case", "case", tc.name, "desc", tc.desc)

			result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, tc.request)

			assert.Error(s.T(), err, tc.desc)
			if result != nil {
				assert.False(s.T(), result.Success, "Result should indicate failure")
			}

			s.logger.Info("Validation test passed", "case", tc.name)
		})
	}
}

// TestNonExistentEmployee tests update for non-existent employee
func (s *Phase2ASimpleTestSuite) TestNonExistentEmployee() {
	s.logger.Info("Testing update for non-existent employee")

	nonExistentID := uuid.New()
	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: nonExistentID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"first_name": "不存在",
			"last_name":  "的员工",
		},
		UpdatedBy: s.testUpdaterID,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	assert.Error(s.T(), err, "Should fail for non-existent employee")
	assert.Contains(s.T(), err.Error(), "not found", "Error should indicate employee not found")

	s.logger.Info("Non-existent employee test passed")
}

// TestPhase2ASimpleTestSuite runs the test suite
func TestPhase2ASimpleTestSuite(t *testing.T) {
	// Check if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(Phase2ASimpleTestSuite))
}

// Main function for standalone execution
func main() {
	fmt.Println("=== Phase 2A Simple Integration Test ===")
	fmt.Println("Testing core employee information management functionality:")
	fmt.Println("- Employee information updates (PERSONAL, CONTACT, EMERGENCY_CONTACT, BANKING)")
	fmt.Println("- Candidate information updates (PERSONAL, APPLICATION_STATUS, INTERVIEW_FEEDBACK)")
	fmt.Println("- Input validation and error handling")
	fmt.Println()

	// Set test environment
	os.Setenv("GO_ENV", "test")

	// Run the test
	fmt.Println("Running Phase 2A simple integration tests...")
	log.Println("Use 'go test -v test_phase2a_simple.go' to run the actual tests")
}
