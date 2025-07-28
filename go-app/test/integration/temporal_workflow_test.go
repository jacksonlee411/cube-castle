// test/integration/temporal_workflow_test.go
package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/enttest"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
	"go.temporal.io/sdk/testsuite"
	_ "github.com/mattn/go-sqlite3"
)

// TemporalWorkflowIntegrationTestSuite provides integration tests for Temporal workflows
type TemporalWorkflowIntegrationTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	entClient        *ent.Client
	temporalQuerySvc *service.TemporalQueryService
	samService       *service.SAMService
	ctx              context.Context
	logger           *log.Logger
}

// SetupSuite runs once before all tests
func (suite *TemporalWorkflowIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = log.New(os.Stdout, "INTEGRATION: ", log.LstdFlags)
	
	// Create in-memory SQLite database for testing
	suite.entClient = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Initialize services
	suite.temporalQuerySvc = service.NewTemporalQueryService(suite.entClient, nil)
	suite.samService = service.NewSAMService(suite.entClient, nil, suite.logger)
}

// TearDownSuite runs once after all tests
func (suite *TemporalWorkflowIntegrationTestSuite) TearDownSuite() {
	suite.entClient.Close()
}

// SetupTest runs before each test
func (suite *TemporalWorkflowIntegrationTestSuite) SetupTest() {
	// Clean database state
	suite.entClient.PositionHistory.Delete().ExecX(suite.ctx)
	suite.entClient.Employee.Delete().ExecX(suite.ctx)
}

// TestPositionChangeWorkflowSuccess tests successful position change workflow
func (suite *TemporalWorkflowIntegrationTestSuite) TestPositionChangeWorkflowSuccess() {
	// Setup test environment
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create test employee
	employee := suite.createTestEmployee("EMP001", "张三", "zhang.san@company.com")
	
	// Create initial position
	initialPosition := suite.createInitialPosition(employee.ID, "软件工程师", "技术部", "INTERMEDIATE")

	// Define position change request
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:  "晋升",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle:  "高级软件工程师",
			Department:     "技术部",
			JobLevel:       "SENIOR",
			EmploymentType: "FULL_TIME",
			MinSalary:      12000,
			MaxSalary:      18000,
			Currency:       "CNY",
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Execute workflow
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Verify workflow completed successfully
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.Success)
	assert.NotEmpty(suite.T(), result.PositionHistoryID)
	assert.Empty(suite.T(), result.Errors)

	// Verify database state
	positions, err := suite.entClient.PositionHistory.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("employee_id", employee.ID))
		}).
		All(suite.ctx)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), positions, 2) // Initial + new position

	// Verify position history consistency
	currentPos, err := suite.temporalQuerySvc.GetPositionAsOfDate(
		suite.ctx, changeRequest.TenantID, employee.ID, time.Now())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "高级软件工程师", currentPos.PositionTitle)
	assert.Equal(suite.T(), "SENIOR", currentPos.JobLevel)
}

// TestPositionChangeWorkflowWithApproval tests position change workflow requiring approval
func (suite *TemporalWorkflowIntegrationTestSuite) TestPositionChangeWorkflowWithApproval() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create test employee
	employee := suite.createTestEmployee("EMP002", "李四", "li.si@company.com")
	suite.createInitialPosition(employee.ID, "项目经理", "技术部", "MANAGER")

	// Define high-impact position change requiring approval
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:  "晋升",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle:  "技术总监",
			Department:     "技术部",
			JobLevel:       "EXECUTIVE",
			EmploymentType: "FULL_TIME",
			MinSalary:      25000,
			MaxSalary:      35000,
			Currency:       "CNY",
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Start workflow execution
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Simulate approval signal
	env.SignalWorkflow("approval", workflow.ApprovalDecision{
		Approved:    true,
		ApprovedBy:  uuid.New(),
		Comments:    "基于出色的工作表现，同意晋升",
		ApprovedAt:  time.Now(),
	})

	// Verify workflow completed successfully
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result includes approval
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.Success)
	assert.True(suite.T(), result.RequiredApproval)
	assert.True(suite.T(), result.Approved)
}

// TestPositionChangeWorkflowRejection tests position change workflow rejection
func (suite *TemporalWorkflowIntegrationTestSuite) TestPositionChangeWorkflowRejection() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create test employee
	employee := suite.createTestEmployee("EMP003", "王五", "wang.wu@company.com")
	suite.createInitialPosition(employee.ID, "初级工程师", "技术部", "JUNIOR")

	// Define position change request
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:  "内部调动",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle:  "高级工程师",
			Department:     "技术部",
			JobLevel:       "SENIOR",
			EmploymentType: "FULL_TIME",
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Start workflow execution
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Simulate rejection signal
	env.SignalWorkflow("approval", workflow.ApprovalDecision{
		Approved:   false,
		ApprovedBy: uuid.New(),
		Comments:   "工作经验不足，建议再工作一年后申请",
		ApprovedAt: time.Now(),
	})

	// Verify workflow completed
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result reflects rejection
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), result.Success)
	assert.True(suite.T(), result.RequiredApproval)
	assert.False(suite.T(), result.Approved)
	assert.NotEmpty(suite.T(), result.RejectionReason)

	// Verify database state - should have only original position
	positions, err := suite.entClient.PositionHistory.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.EQ("employee_id", employee.ID))
		}).
		All(suite.ctx)
	
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), positions, 1) // Only initial position
}

// TestPositionChangeWorkflowTemporalConflict tests workflow with temporal consistency conflict
func (suite *TemporalWorkflowIntegrationTestSuite) TestPositionChangeWorkflowTemporalConflict() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create test employee
	employee := suite.createTestEmployee("EMP004", "赵六", "zhao.liu@company.com")
	suite.createInitialPosition(employee.ID, "分析师", "产品部", "INTERMEDIATE")

	// Create overlapping position change request
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), // Before initial position
		ChangeReason:  "追溯调整",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle:  "高级分析师",
			Department:     "产品部",
			JobLevel:       "SENIOR",
			EmploymentType: "FULL_TIME",
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Execute workflow
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Verify workflow completed with validation error
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result shows validation failure
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), result.Success)
	assert.NotEmpty(suite.T(), result.Errors)
	assert.Contains(suite.T(), result.Errors[0], "temporal consistency")
}

// TestBulkPositionChangeWorkflow tests bulk position changes workflow
func (suite *TemporalWorkflowIntegrationTestSuite) TestBulkPositionChangeWorkflow() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create multiple test employees
	employees := make([]*ent.Employee, 3)
	for i := 0; i < 3; i++ {
		empID := fmt.Sprintf("BULK%03d", i+1)
		name := fmt.Sprintf("批量员工%d", i+1)
		email := fmt.Sprintf("bulk%d@company.com", i+1)
		
		employees[i] = suite.createTestEmployee(empID, name, email)
		suite.createInitialPosition(employees[i].ID, "工程师", "技术部", "INTERMEDIATE")
	}

	// Define bulk position change request
	bulkRequest := workflow.BulkPositionChangeRequest{
		TenantID:    uuid.New(),
		RequestedBy: uuid.New(),
		Changes: []workflow.PositionChangeRequest{
			{
				EmployeeID:    employees[0].ID,
				EffectiveDate: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				ChangeReason:  "年度晋升",
				NewPosition: workflow.PositionChangeData{
					PositionTitle: "高级工程师",
					JobLevel:      "SENIOR",
				},
			},
			{
				EmployeeID:    employees[1].ID,
				EffectiveDate: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				ChangeReason:  "年度晋升",
				NewPosition: workflow.PositionChangeData{
					PositionTitle: "高级工程师",
					JobLevel:      "SENIOR",
				},
			},
			{
				EmployeeID:    employees[2].ID,
				EffectiveDate: time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
				ChangeReason:  "部门调动",
				NewPosition: workflow.PositionChangeData{
					PositionTitle: "产品工程师",
					Department:    "产品部",
					JobLevel:      "INTERMEDIATE",
				},
			},
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.BulkPositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Execute bulk workflow
	env.ExecuteWorkflow(workflow.BulkPositionChangeWorkflow, bulkRequest)

	// Verify workflow completed successfully
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result
	var result workflow.BulkPositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, result.TotalChanges)
	assert.Equal(suite.T(), 3, result.SuccessfulChanges)
	assert.Equal(suite.T(), 0, result.FailedChanges)
	assert.Empty(suite.T(), result.Errors)

	// Verify all position changes were applied correctly
	for i, emp := range employees {
		currentPos, err := suite.temporalQuerySvc.GetPositionAsOfDate(
			suite.ctx, bulkRequest.TenantID, emp.ID, time.Now())
		assert.NoError(suite.T(), err)
		
		if i < 2 {
			assert.Equal(suite.T(), "高级工程师", currentPos.PositionTitle)
			assert.Equal(suite.T(), "SENIOR", currentPos.JobLevel)
		} else {
			assert.Equal(suite.T(), "产品工程师", currentPos.PositionTitle)
			assert.Equal(suite.T(), "产品部", currentPos.Department)
		}
	}
}

// TestPositionChangeWorkflowWithSAMAnalysis tests workflow integration with SAM analysis
func (suite *TemporalWorkflowIntegrationTestSuite) TestPositionChangeWorkflowWithSAMAnalysis() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Create test employee in critical role
	employee := suite.createTestEmployee("EMP005", "孙七", "sun.qi@company.com")
	suite.createInitialPosition(employee.ID, "技术总监", "技术部", "EXECUTIVE")

	// Define position change request that might impact organization
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2023, 11, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:  "离职",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle:  "",
			Department:     "",
			JobLevel:       "",
			EmploymentType: "TERMINATED",
		},
	}

	// Register workflow and activities
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Execute workflow
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Verify workflow completed
	assert.True(suite.T(), env.IsWorkflowCompleted())
	assert.NoError(suite.T(), env.GetWorkflowError())

	// Verify workflow result includes SAM analysis
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result.SAMAnalysis)
	assert.NotEmpty(suite.T(), result.SAMAnalysis.RiskAssessment)
	assert.NotEmpty(suite.T(), result.SAMAnalysis.Recommendations)
}

// TestWorkflowTimeout tests workflow timeout handling
func (suite *TemporalWorkflowIntegrationTestSuite) TestWorkflowTimeout() {
	env := suite.NewTestWorkflowEnvironment()
	defer env.Stop()

	// Set short timeout for testing
	env.SetTestTimeout(time.Second * 5)

	employee := suite.createTestEmployee("EMP006", "周八", "zhou.ba@company.com")
	
	changeRequest := workflow.PositionChangeRequest{
		TenantID:      uuid.New(),
		EmployeeID:    employee.ID,
		EffectiveDate: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:  "测试超时",
		RequestedBy:   uuid.New(),
		NewPosition: workflow.PositionChangeData{
			PositionTitle: "测试工程师",
		},
	}

	// Register workflow that will wait for approval (timeout scenario)
	env.RegisterWorkflow(workflow.PositionChangeWorkflow)
	env.RegisterActivity(&workflow.PositionChangeActivities{
		EntClient:        suite.entClient,
		TemporalQuerySvc: suite.temporalQuerySvc,
		SAMService:       suite.samService,
		Logger:           suite.logger,
	})

	// Execute workflow without providing approval signal
	env.ExecuteWorkflow(workflow.PositionChangeWorkflow, changeRequest)

	// Verify workflow timed out
	assert.True(suite.T(), env.IsWorkflowCompleted())
	
	// Check if timeout was handled appropriately
	var result workflow.PositionChangeResult
	err := env.GetWorkflowResult(&result)
	
	// Workflow should either timeout or handle timeout gracefully
	if err != nil {
		assert.Contains(suite.T(), err.Error(), "timeout")
	} else {
		assert.False(suite.T(), result.Success)
		assert.Contains(suite.T(), result.Errors[0], "timeout")
	}
}

// Helper methods

func (suite *TemporalWorkflowIntegrationTestSuite) createTestEmployee(empID, name, email string) *ent.Employee {
	return suite.entClient.Employee.Create().
		SetEmployeeID(empID).
		SetLegalName(name).
		SetEmail(email).
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)
}

func (suite *TemporalWorkflowIntegrationTestSuite) createInitialPosition(employeeID uuid.UUID, title, dept, level string) *ent.PositionHistory {
	return suite.entClient.PositionHistory.Create().
		SetEmployeeID(employeeID).
		SetPositionTitle(title).
		SetDepartment(dept).
		SetJobLevel(level).
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)
}

// TestTemporalWorkflowIntegrationTestSuite runs the test suite
func TestTemporalWorkflowIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(TemporalWorkflowIntegrationTestSuite))
}