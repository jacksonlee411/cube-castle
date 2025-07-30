// test_phase2a_minimal.go - Minimal Phase 2A Test Without Service Dependencies
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
)

// Phase2AMinimalTestSuite - Minimal test suite for Phase 2A functionality
type Phase2AMinimalTestSuite struct {
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
func (s *Phase2AMinimalTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.logger = logging.NewStructuredLogger()

	// Use in-memory SQLite for testing
	client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	s.Require().NoError(err, "Failed to create in-memory ent client")

	// Run schema creation
	err = client.Schema.Create(s.ctx)
	s.Require().NoError(err, "Failed to create database schema")

	s.entClient = client

	// Initialize activities without temporal query service to avoid dependencies
	s.activities = workflow.NewEmployeeLifecycleActivities(s.entClient, nil, s.logger)

	// Generate test UUIDs
	s.testTenantID = uuid.New()
	s.testEmployeeID = uuid.New()
	s.testCandidateID = uuid.New()
	s.testUpdaterID = uuid.New()

	s.logger.Info("Phase 2A Minimal Test Suite initialized",
		"tenant_id", s.testTenantID,
		"employee_id", s.testEmployeeID,
		"candidate_id", s.testCandidateID,
	)
}

// TearDownSuite cleans up the test suite
func (s *Phase2AMinimalTestSuite) TearDownSuite() {
	if s.entClient != nil {
		s.entClient.Close()
	}
}

// SetupTest prepares each individual test
func (s *Phase2AMinimalTestSuite) SetupTest() {
	// Create test employee record
	s.createTestEmployee()
	s.createTestCandidate()
}

// TearDownTest cleans up after each test
func (s *Phase2AMinimalTestSuite) TearDownTest() {
	// Clean up test data
	s.cleanupTestData()
}

// createTestEmployee creates a test employee record
func (s *Phase2AMinimalTestSuite) createTestEmployee() {
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
func (s *Phase2AMinimalTestSuite) createTestCandidate() {
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
func (s *Phase2AMinimalTestSuite) cleanupTestData() {
	// Delete test employee and candidate records
	s.entClient.Employee.Delete().ExecX(s.ctx)
	s.logger.Debug("Test data cleaned up")
}

// TestUpdateEmployeeInformation_PersonalInfo tests employee personal information update
func (s *Phase2AMinimalTestSuite) TestUpdateEmployeeInformation_PersonalInfo() {
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

// TestUpdateCandidateInformation_PersonalInfo tests candidate personal information update
func (s *Phase2AMinimalTestSuite) TestUpdateCandidateInformation_PersonalInfo() {
	s.logger.Info("Testing candidate personal information update")

	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: s.testCandidateID, // Candidate uses EmployeeID field
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"name":  "王五",
			"email": "wang.wu@candidate.com",
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

// TestInvalidInputValidation tests input validation for invalid requests
func (s *Phase2AMinimalTestSuite) TestInvalidInputValidation() {
	s.logger.Info("Testing input validation")

	// Test missing tenant_id
	updateRequest := workflow.InformationUpdateRequest{
		EmployeeID: s.testEmployeeID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{"test": "data"},
		UpdatedBy:  s.testUpdaterID,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)
	assert.Error(s.T(), err, "Should fail when tenant_id is missing")
	assert.Contains(s.T(), err.Error(), "tenant_id is required")

	// Test missing employee_id
	updateRequest2 := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{"test": "data"},
		UpdatedBy:  s.testUpdaterID,
	}

	result2, err2 := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest2)
	assert.Error(s.T(), err2, "Should fail when employee_id is missing")
	assert.Contains(s.T(), err2.Error(), "employee_id is required")

	s.logger.Info("Input validation tests passed")
}

// TestNonExistentEmployee tests update for non-existent employee
func (s *Phase2AMinimalTestSuite) TestNonExistentEmployee() {
	s.logger.Info("Testing update for non-existent employee")

	nonExistentID := uuid.New()
	updateRequest := workflow.InformationUpdateRequest{
		TenantID:   s.testTenantID,
		EmployeeID: nonExistentID,
		UpdateType: "PERSONAL",
		UpdateData: map[string]interface{}{
			"legal_name": "不存在的员工",
		},
		UpdatedBy: s.testUpdaterID,
	}

	result, err := s.activities.UpdateEmployeeInformationActivity(s.ctx, updateRequest)

	assert.Error(s.T(), err, "Should fail for non-existent employee")
	assert.Contains(s.T(), err.Error(), "not found", "Error should indicate employee not found")

	s.logger.Info("Non-existent employee test passed")
}

// TestPhase2AMinimalTestSuite runs the test suite
func TestPhase2AMinimalTestSuite(t *testing.T) {
	// Check if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(Phase2AMinimalTestSuite))
}

// Main function for standalone execution
func main() {
	fmt.Println("=== Phase 2A Minimal Integration Test ===")
	fmt.Println("Testing core employee information management functionality:")
	fmt.Println("- Employee information updates (PERSONAL)")
	fmt.Println("- Candidate information updates (PERSONAL)")
	fmt.Println("- Input validation and error handling")
	fmt.Println()

	// Set test environment
	os.Setenv("GO_ENV", "test")

	// Run the test
	fmt.Println("Running Phase 2A minimal integration tests...")
	log.Println("Use 'go test -v test_phase2a_minimal.go' to run the actual tests")
}
