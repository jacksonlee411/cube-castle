// test/integration/database_integration_test.go
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
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseIntegrationTestSuite provides integration tests for database operations
type DatabaseIntegrationTestSuite struct {
	suite.Suite
	entClient        *ent.Client
	temporalQuerySvc *service.TemporalQueryService
	ctx              context.Context
	logger           *log.Logger
}

// SetupSuite runs once before all tests
func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = log.New(os.Stdout, "DB_INTEGRATION: ", log.LstdFlags)
	
	// Create in-memory SQLite database for testing
	suite.entClient = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Initialize services
	suite.temporalQuerySvc = service.NewTemporalQueryService(suite.entClient, nil)
}

// TearDownSuite runs once after all tests
func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	suite.entClient.Close()
}

// SetupTest runs before each test
func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Clean database state
	suite.entClient.PositionHistory.Delete().ExecX(suite.ctx)
	suite.entClient.Employee.Delete().ExecX(suite.ctx)
}

// TestEmployeeLifecycleIntegration tests complete employee lifecycle with database
func (suite *DatabaseIntegrationTestSuite) TestEmployeeLifecycleIntegration() {
	tenantID := uuid.New()
	
	// Create employee
	employee := suite.entClient.Employee.Create().
		SetEmployeeID("EMP001").
		SetLegalName("张三").
		SetEmail("zhang.san@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)
	
	// Create initial position
	initialPosition := suite.entClient.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("软件工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)
	
	// Verify initial position is retrievable
	currentPos, err := suite.temporalQuerySvc.GetPositionAsOfDate(
		suite.ctx, tenantID, employee.ID, time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), currentPos)
	assert.Equal(suite.T(), "软件工程师", currentPos.PositionTitle)
	assert.Equal(suite.T(), "技术部", currentPos.Department)
	
	// Add promotion
	promotionDate := time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)
	promotion := suite.entClient.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("高级软件工程师").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(promotionDate).
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)
	
	// Update initial position end date
	suite.entClient.PositionHistory.UpdateOne(initialPosition).
		SetEndDate(promotionDate.Add(-time.Second)).
		SaveX(suite.ctx)
	
	// Verify promotion is active
	currentPos, err = suite.temporalQuerySvc.GetPositionAsOfDate(
		suite.ctx, tenantID, employee.ID, time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), currentPos)
	assert.Equal(suite.T(), "高级软件工程师", currentPos.PositionTitle)
	assert.Equal(suite.T(), "SENIOR", currentPos.JobLevel)
	
	// Verify position timeline
	timeline, err := suite.temporalQuerySvc.GetPositionTimeline(
		suite.ctx, tenantID, employee.ID, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), timeline, 2)
	assert.Equal(suite.T(), "软件工程师", timeline[0].PositionTitle)
	assert.Equal(suite.T(), "高级软件工程师", timeline[1].PositionTitle)
	
	// Test temporal consistency validation
	err = suite.temporalQuerySvc.ValidateTemporalConsistency(
		suite.ctx, tenantID, employee.ID, time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.NoError(suite.T(), err)
	
	// Test invalid temporal consistency (overlapping dates)
	err = suite.temporalQuerySvc.ValidateTemporalConsistency(
		suite.ctx, tenantID, employee.ID, time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC))
	assert.Error(suite.T(), err)
	
	// Test retroactive position change
	retroactiveDate := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)
	retroactive := suite.entClient.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("中级软件工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(retroactiveDate).
		SetEndDate(promotionDate.Add(-time.Second)).
		SetChangeReason("薪资调整").
		SetIsRetroactive(true).
		SaveX(suite.ctx)
	
	// Update original position to end earlier
	suite.entClient.PositionHistory.UpdateOne(initialPosition).
		SetEndDate(retroactiveDate.Add(-time.Second)).
		SaveX(suite.ctx)
	
	// Verify retroactive change is reflected in timeline
	timeline, err = suite.temporalQuerySvc.GetPositionTimeline(
		suite.ctx, tenantID, employee.ID, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), timeline, 3)
	
	// Verify positions are in chronological order
	assert.True(suite.T(), timeline[0].EffectiveDate.Before(timeline[1].EffectiveDate))
	assert.True(suite.T(), timeline[1].EffectiveDate.Before(timeline[2].EffectiveDate))
	
	suite.logger.Printf("Employee lifecycle test completed successfully with %d positions", len(timeline))
}

// TestConcurrentDatabaseOperations tests concurrent database operations
func (suite *DatabaseIntegrationTestSuite) TestConcurrentDatabaseOperations() {
	const numEmployees = 10
	tenantID := uuid.New()
	
	// Create multiple employees concurrently
	employees := make([]*ent.Employee, numEmployees)
	errors := make(chan error, numEmployees)
	
	for i := 0; i < numEmployees; i++ {
		go func(index int) {
			employee := suite.entClient.Employee.Create().
				SetEmployeeID(fmt.Sprintf("EMP%03d", index+1)).
				SetLegalName(fmt.Sprintf("员工%d", index+1)).
				SetEmail(fmt.Sprintf("emp%d@company.com", index+1)).
				SetStatus("ACTIVE").
				SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
				SaveX(suite.ctx)
			
			employees[index] = employee
			
			// Create position for each employee
			suite.entClient.PositionHistory.Create().
				SetEmployeeID(employee.ID).
				SetPositionTitle(fmt.Sprintf("工程师%d", index+1)).
				SetDepartment("技术部").
				SetJobLevel("INTERMEDIATE").
				SetEmploymentType("FULL_TIME").
				SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
				SetChangeReason("入职").
				SetIsRetroactive(false).
				SaveX(suite.ctx)
			
			errors <- nil
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < numEmployees; i++ {
		err := <-errors
		assert.NoError(suite.T(), err)
	}
	
	// Verify all employees and positions were created
	employeeCount, err := suite.entClient.Employee.Query().Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), numEmployees, employeeCount)
	
	positionCount, err := suite.entClient.PositionHistory.Query().Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), numEmployees, positionCount)
	
	// Test concurrent temporal queries
	queryErrors := make(chan error, numEmployees)
	
	for _, emp := range employees {
		go func(employee *ent.Employee) {
			if employee != nil {
				_, err := suite.temporalQuerySvc.GetPositionAsOfDate(
					suite.ctx, tenantID, employee.ID, time.Now())
				queryErrors <- err
			} else {
				queryErrors <- fmt.Errorf("employee is nil")
			}
		}(emp)
	}
	
	// Verify all queries succeeded
	for i := 0; i < numEmployees; i++ {
		err := <-queryErrors
		assert.NoError(suite.T(), err)
	}
	
	suite.logger.Printf("Concurrent operations test completed with %d employees", numEmployees)
}

// TestDatabaseConstraintsAndValidation tests database constraints
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConstraintsAndValidation() {
	// Test unique constraint on employee ID
	employee1 := suite.entClient.Employee.Create().
		SetEmployeeID("EMP001").
		SetLegalName("张三").
		SetEmail("zhang.san@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)
	
	// Attempt to create duplicate employee ID should fail
	assert.Panics(suite.T(), func() {
		suite.entClient.Employee.Create().
			SetEmployeeID("EMP001"). // Duplicate
			SetLegalName("李四").
			SetEmail("li.si@company.com").
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(suite.ctx)
	})
	
	// Test foreign key constraint
	position := suite.entClient.PositionHistory.Create().
		SetEmployeeID(employee1.ID).
		SetPositionTitle("软件工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)
	
	assert.NotNil(suite.T(), position)
	assert.Equal(suite.T(), employee1.ID, position.EmployeeID)
	
	// Test required field validation
	assert.Panics(suite.T(), func() {
		suite.entClient.Employee.Create().
			// Missing required EmployeeID
			SetLegalName("王五").
			SetEmail("wang.wu@company.com").
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(suite.ctx)
	})
	
	suite.logger.Println("Database constraints and validation tests completed")
}

// TestComplexTemporalQueries tests complex temporal database queries
func (suite *DatabaseIntegrationTestSuite) TestComplexTemporalQueries() {
	tenantID := uuid.New()
	
	// Create employee with complex position history
	employee := suite.entClient.Employee.Create().
		SetEmployeeID("EMP001").
		SetLegalName("张三").
		SetEmail("zhang.san@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)
	
	// Create multiple positions with overlapping and gap scenarios
	positions := []struct {
		title     string
		dept      string
		level     string
		startDate time.Time
		endDate   *time.Time
		reason    string
	}{
		{
			"初级工程师", "技术部", "JUNIOR",
			time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			&[]time.Time{time.Date(2019, 6, 30, 23, 59, 59, 0, time.UTC)}[0],
			"入职",
		},
		{
			"中级工程师", "技术部", "INTERMEDIATE",
			time.Date(2019, 7, 1, 0, 0, 0, 0, time.UTC),
			&[]time.Time{time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC)}[0],
			"晋升",
		},
		{
			"高级工程师", "技术部", "SENIOR",
			time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			&[]time.Time{time.Date(2023, 6, 30, 23, 59, 59, 0, time.UTC)}[0],
			"晋升",
		},
		{
			"技术主管", "技术部", "MANAGER",
			time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
			nil, // Current position
			"晋升",
		},
	}
	
	for _, pos := range positions {
		builder := suite.entClient.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle(pos.title).
			SetDepartment(pos.dept).
			SetJobLevel(pos.level).
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(pos.startDate).
			SetChangeReason(pos.reason).
			SetIsRetroactive(false)
		
		if pos.endDate != nil {
			builder = builder.SetEndDate(*pos.endDate)
		}
		
		builder.SaveX(suite.ctx)
	}
	
	// Test point-in-time queries at different dates
	testDates := []struct {
		date           time.Time
		expectedTitle  string
		expectedLevel  string
	}{
		{time.Date(2018, 6, 1, 0, 0, 0, 0, time.UTC), "初级工程师", "JUNIOR"},
		{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), "中级工程师", "INTERMEDIATE"},
		{time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), "高级工程师", "SENIOR"},
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "技术主管", "MANAGER"},
	}
	
	for _, testCase := range testDates {
		position, err := suite.temporalQuerySvc.GetPositionAsOfDate(
			suite.ctx, tenantID, employee.ID, testCase.date)
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), position)
		assert.Equal(suite.T(), testCase.expectedTitle, position.PositionTitle)
		assert.Equal(suite.T(), testCase.expectedLevel, position.JobLevel)
	}
	
	// Test timeline query with date range
	fromDate := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC)
	
	timeline, err := suite.temporalQuerySvc.GetPositionTimeline(
		suite.ctx, tenantID, employee.ID, &fromDate, &toDate)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), timeline, 2) // Should only include positions in range
	assert.Equal(suite.T(), "中级工程师", timeline[0].PositionTitle)
	assert.Equal(suite.T(), "高级工程师", timeline[1].PositionTitle)
	
	// Test complete timeline
	fullTimeline, err := suite.temporalQuerySvc.GetPositionTimeline(
		suite.ctx, tenantID, employee.ID, nil, nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), fullTimeline, 4)
	
	// Verify chronological order
	for i := 1; i < len(fullTimeline); i++ {
		assert.True(suite.T(), 
			fullTimeline[i-1].EffectiveDate.Before(fullTimeline[i].EffectiveDate) ||
			fullTimeline[i-1].EffectiveDate.Equal(fullTimeline[i].EffectiveDate))
	}
	
	suite.logger.Printf("Complex temporal queries test completed with %d positions", len(fullTimeline))
}

// TestDatabaseTransactionIntegrity tests transaction handling
func (suite *DatabaseIntegrationTestSuite) TestDatabaseTransactionIntegrity() {
	// Test successful transaction
	err := suite.entClient.DoTx(suite.ctx, nil, func(tx *ent.Tx) error {
		employee := tx.Employee.Create().
			SetEmployeeID("TXN001").
			SetLegalName("事务测试员工").
			SetEmail("txn@test.com").
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(suite.ctx)
		
		tx.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("事务测试工程师").
			SetDepartment("技术部").
			SetJobLevel("INTERMEDIATE").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SetChangeReason("入职").
			SetIsRetroactive(false).
			SaveX(suite.ctx)
		
		return nil // Commit transaction
	})
	
	assert.NoError(suite.T(), err)
	
	// Verify data was committed
	count, err := suite.entClient.Employee.Query().Where(
		func(s *sql.Selector) {
			s.Where(sql.EQ("employee_id", "TXN001"))
		}).Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
	
	// Test rolled back transaction
	err = suite.entClient.DoTx(suite.ctx, nil, func(tx *ent.Tx) error {
		tx.Employee.Create().
			SetEmployeeID("TXN002").
			SetLegalName("回滚测试员工").
			SetEmail("rollback@test.com").
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(suite.ctx)
		
		return fmt.Errorf("simulated error") // Force rollback
	})
	
	assert.Error(suite.T(), err)
	
	// Verify data was not committed
	count, err = suite.entClient.Employee.Query().Where(
		func(s *sql.Selector) {
			s.Where(sql.EQ("employee_id", "TXN002"))
		}).Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
	
	suite.logger.Println("Database transaction integrity tests completed")
}

// TestDatabasePerformanceWithLargeDataset tests performance with larger dataset
func (suite *DatabaseIntegrationTestSuite) TestDatabasePerformanceWithLargeDataset() {
	const numEmployees = 100
	const positionsPerEmployee = 3
	
	start := time.Now()
	
	// Create large dataset
	employees := make([]*ent.Employee, numEmployees)
	for i := 0; i < numEmployees; i++ {
		employee := suite.entClient.Employee.Create().
			SetEmployeeID(fmt.Sprintf("PERF%03d", i+1)).
			SetLegalName(fmt.Sprintf("性能测试员工%d", i+1)).
			SetEmail(fmt.Sprintf("perf%d@test.com", i+1)).
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(suite.ctx)
		
		employees[i] = employee
		
		// Create multiple positions for each employee
		for j := 0; j < positionsPerEmployee; j++ {
			effectiveDate := time.Date(2020+j, 1, 1, 0, 0, 0, 0, time.UTC)
			endDate := time.Date(2020+j, 12, 31, 23, 59, 59, 0, time.UTC)
			
			builder := suite.entClient.PositionHistory.Create().
				SetEmployeeID(employee.ID).
				SetPositionTitle(fmt.Sprintf("职位%d", j+1)).
				SetDepartment(fmt.Sprintf("部门%d", (i%5)+1)).
				SetJobLevel("INTERMEDIATE").
				SetEmploymentType("FULL_TIME").
				SetEffectiveDate(effectiveDate).
				SetChangeReason("职位变更").
				SetIsRetroactive(false)
			
			if j < positionsPerEmployee-1 {
				builder = builder.SetEndDate(endDate)
			}
			
			builder.SaveX(suite.ctx)
		}
	}
	
	creationDuration := time.Since(start)
	suite.logger.Printf("Created %d employees with %d positions in %v", 
		numEmployees, numEmployees*positionsPerEmployee, creationDuration)
	
	// Test query performance
	start = time.Now()
	tenantID := uuid.New()
	
	// Perform temporal queries for random employees
	for i := 0; i < 20; i++ {
		empIndex := i % numEmployees
		queryDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)
		
		_, err := suite.temporalQuerySvc.GetPositionAsOfDate(
			suite.ctx, tenantID, employees[empIndex].ID, queryDate)
		assert.NoError(suite.T(), err)
	}
	
	queryDuration := time.Since(start)
	suite.logger.Printf("Performed 20 temporal queries in %v (avg: %v per query)", 
		queryDuration, queryDuration/20)
	
	// Test aggregate queries
	start = time.Now()
	
	departmentStats := make(map[string]int)
	positions, err := suite.entClient.PositionHistory.Query().
		Where(positionhistory.EndDateIsNil()).
		All(suite.ctx)
	assert.NoError(suite.T(), err)
	
	for _, pos := range positions {
		departmentStats[pos.Department]++
	}
	
	aggregateDuration := time.Since(start)
	suite.logger.Printf("Aggregated %d positions by department in %v", 
		len(positions), aggregateDuration)
	
	// Verify performance thresholds
	assert.Less(suite.T(), creationDuration, 10*time.Second, 
		"Creation should complete within 10 seconds")
	assert.Less(suite.T(), queryDuration/20, 100*time.Millisecond, 
		"Individual queries should complete within 100ms")
	assert.Less(suite.T(), aggregateDuration, 1*time.Second, 
		"Aggregate queries should complete within 1 second")
	
	suite.logger.Printf("Performance test completed - dataset size: %d employees, %d positions", 
		numEmployees, numEmployees*positionsPerEmployee)
}

// TestDatabaseIntegrationTestSuite runs the test suite
func TestDatabaseIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}