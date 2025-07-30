// internal/service/temporal_query_service_test.go
package service

import (
	"context"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/enttest"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// TemporalQueryServiceTestSuite provides test suite for TemporalQueryService
type TemporalQueryServiceTestSuite struct {
	suite.Suite
	client  *ent.Client
	service *TemporalQueryService
	ctx     context.Context
}

// SetupSuite runs once before all tests
func (suite *TemporalQueryServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Create in-memory SQLite database for testing
	suite.client = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	// Initialize service
	suite.service = NewTemporalQueryService(suite.client, nil)
}

// TearDownSuite runs once after all tests
func (suite *TemporalQueryServiceTestSuite) TearDownSuite() {
	suite.client.Close()
}

// SetupTest runs before each test
func (suite *TemporalQueryServiceTestSuite) SetupTest() {
	// Clean database state
	suite.client.PositionHistory.Delete().ExecX(suite.ctx)
	suite.client.Employee.Delete().ExecX(suite.ctx)
}

// TestGetPositionAsOfDate tests point-in-time position queries
func (suite *TemporalQueryServiceTestSuite) TestGetPositionAsOfDate() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP001").
		SetLegalName("张三").
		SetEmail("zhang.san@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create position history
	pos1 := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("软件工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetEndDate(time.Date(2021, 6, 30, 23, 59, 59, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	pos2 := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("高级软件工程师").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Test cases
	testCases := []struct {
		name        string
		asOfDate    time.Time
		expectedPos *ent.PositionHistory
		expectError bool
	}{
		{
			name:        "Position at start date",
			asOfDate:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedPos: pos1,
			expectError: false,
		},
		{
			name:        "Position during first period",
			asOfDate:    time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC),
			expectedPos: pos1,
			expectError: false,
		},
		{
			name:        "Position at transition date",
			asOfDate:    time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			expectedPos: pos2,
			expectError: false,
		},
		{
			name:        "Current position",
			asOfDate:    time.Now(),
			expectedPos: pos2,
			expectError: false,
		},
		{
			name:        "Position before employment",
			asOfDate:    time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedPos: nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			position, err := suite.service.GetPositionAsOfDate(suite.ctx, employee.ID, tc.asOfDate)

			if tc.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), position)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), position)
				assert.Equal(suite.T(), tc.expectedPos.ID, position.ID)
				assert.Equal(suite.T(), tc.expectedPos.PositionTitle, position.PositionTitle)
				assert.Equal(suite.T(), tc.expectedPos.JobLevel, position.JobLevel)
			}
		})
	}
}

// TestGetPositionTimeline tests timeline query functionality
func (suite *TemporalQueryServiceTestSuite) TestGetPositionTimeline() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP002").
		SetLegalName("李四").
		SetEmail("li.si@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create multiple positions
	positions := []*ent.PositionHistory{
		suite.client.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("初级开发").
			SetDepartment("技术部").
			SetJobLevel("JUNIOR").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SetEndDate(time.Date(2020, 12, 31, 23, 59, 59, 0, time.UTC)).
			SetChangeReason("入职").
			SetIsRetroactive(false).
			SaveX(suite.ctx),

		suite.client.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("中级开发").
			SetDepartment("技术部").
			SetJobLevel("INTERMEDIATE").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
			SetEndDate(time.Date(2022, 6, 30, 23, 59, 59, 0, time.UTC)).
			SetChangeReason("晋升").
			SetIsRetroactive(false).
			SaveX(suite.ctx),

		suite.client.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("高级开发").
			SetDepartment("技术部").
			SetJobLevel("SENIOR").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)).
			SetChangeReason("晋升").
			SetIsRetroactive(false).
			SaveX(suite.ctx),
	}

	// Test timeline query
	timeline, err := suite.service.GetPositionTimeline(suite.ctx, employee.ID, nil, nil, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), timeline, 3)

	// Verify order (should be chronological)
	assert.Equal(suite.T(), positions[0].ID, timeline[0].ID)
	assert.Equal(suite.T(), positions[1].ID, timeline[1].ID)
	assert.Equal(suite.T(), positions[2].ID, timeline[2].ID)

	// Test with date range filter
	fromDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	toDate := time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)

	filteredTimeline, err := suite.service.GetPositionTimeline(suite.ctx, employee.ID, &fromDate, &toDate, 10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), filteredTimeline, 2) // Should include positions 2 and 3
}

// TestValidateTemporalConsistency tests temporal consistency validation
func (suite *TemporalQueryServiceTestSuite) TestValidateTemporalConsistency() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP003").
		SetLegalName("王五").
		SetEmail("wang.wu@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create valid position sequence
	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("开发工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetEndDate(time.Date(2021, 6, 30, 23, 59, 59, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("高级开发工程师").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Test consistency validation
	result, err := suite.service.ValidateTemporalConsistency(suite.ctx, employee.ID)

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.IsValid)
	assert.Empty(suite.T(), result.Violations)
}

// TestValidateTemporalConsistency_WithViolations tests consistency validation with violations
func (suite *TemporalQueryServiceTestSuite) TestValidateTemporalConsistency_WithViolations() {
	// Setup test data with temporal violations
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP004").
		SetLegalName("赵六").
		SetEmail("zhao.liu@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create overlapping positions (violation)
	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("开发工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetEndDate(time.Date(2021, 7, 31, 23, 59, 59, 0, time.UTC)). // Overlaps with next position
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("高级开发工程师").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC)). // Overlaps with previous
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Test consistency validation
	result, err := suite.service.ValidateTemporalConsistency(suite.ctx, employee.ID)

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), result.IsValid)
	assert.NotEmpty(suite.T(), result.Violations)

	// Check violation type
	hasOverlapViolation := false
	for _, violation := range result.Violations {
		if violation.ViolationType == "OVERLAPPING_PERIODS" {
			hasOverlapViolation = true
			break
		}
	}
	assert.True(suite.T(), hasOverlapViolation)
}

// TestCreatePositionSnapshot tests position snapshot creation
func (suite *TemporalQueryServiceTestSuite) TestCreatePositionSnapshot() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP005").
		SetLegalName("孙七").
		SetEmail("sun.qi@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create current position
	currentPos := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("技术经理").
		SetDepartment("技术部").
		SetJobLevel("MANAGER").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	snapshotDate := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)

	// Test snapshot creation
	snapshot, err := suite.service.CreatePositionSnapshot(suite.ctx, employee.ID, snapshotDate)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), snapshot)
	assert.Equal(suite.T(), employee.ID, snapshot.EmployeeID)
	assert.Equal(suite.T(), currentPos.PositionTitle, snapshot.PositionTitle)
	assert.Equal(suite.T(), snapshotDate, snapshot.SnapshotDate)
}

// TestTemporalBoundaryConditions tests edge cases and boundary conditions
func (suite *TemporalQueryServiceTestSuite) TestTemporalBoundaryConditions() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP006").
		SetLegalName("周八").
		SetEmail("zhou.ba@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Test with exact boundary timestamps
	boundaryDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	pos1 := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("初级职位").
		SetDepartment("技术部").
		SetJobLevel("JUNIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetEndDate(time.Date(2020, 12, 31, 23, 59, 59, 999000000, time.UTC)). // Just before boundary
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	pos2 := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("高级职位").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(boundaryDate). // Exactly at boundary
		SetChangeReason("晋升").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Query exactly at boundary should return new position
	position, err := suite.service.GetPositionAsOfDate(suite.ctx, employee.ID, boundaryDate)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pos2.ID, position.ID)

	// Query just before boundary should return old position
	beforeBoundary := boundaryDate.Add(-time.Nanosecond)
	positionBefore, err := suite.service.GetPositionAsOfDate(suite.ctx, employee.ID, beforeBoundary)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pos1.ID, positionBefore.ID)
}

// TestRetroactivePositionHandling tests retroactive position changes
func (suite *TemporalQueryServiceTestSuite) TestRetroactivePositionHandling() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP007").
		SetLegalName("吴九").
		SetEmail("wu.jiu@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	// Create initial position
	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("原始职位").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetEndDate(time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Create retroactive position change
	retroactivePos := suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("追溯职位").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)). // Retroactive to middle of period
		SetEndDate(time.Date(2020, 12, 31, 23, 59, 59, 0, time.UTC)).
		SetChangeReason("追溯调整").
		SetIsRetroactive(true).
		SaveX(suite.ctx)

	// Query during retroactive period
	retroactiveDate := time.Date(2020, 8, 15, 0, 0, 0, 0, time.UTC)
	position, err := suite.service.GetPositionAsOfDate(suite.ctx, employee.ID, retroactiveDate)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), retroactivePos.ID, position.ID)
	assert.True(suite.T(), position.IsRetroactive)
}

// TestConcurrentPositionQueries tests concurrent access to temporal data
func (suite *TemporalQueryServiceTestSuite) TestConcurrentPositionQueries() {
	// Setup test data
	employee := suite.client.Employee.Create().
		SetEmployeeID("EMP008").
		SetLegalName("郑十").
		SetEmail("zheng.shi@company.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(suite.ctx)

	suite.client.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("并发测试职位").
		SetDepartment("技术部").
		SetJobLevel("SENIOR").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(suite.ctx)

	// Test concurrent queries
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			queryDate := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
			_, err := suite.service.GetPositionAsOfDate(suite.ctx, employee.ID, queryDate)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err)
	}
}

// TestTemporalQueryServiceSuite runs the test suite
func TestTemporalQueryServiceSuite(t *testing.T) {
	suite.Run(t, new(TemporalQueryServiceTestSuite))
}

// Benchmark tests for performance validation
func BenchmarkGetPositionAsOfDate(b *testing.B) {
	client := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	service := NewTemporalQueryService(client, nil)
	ctx := context.Background()

	// Setup benchmark data
	employee := client.Employee.Create().
		SetEmployeeID("BENCH001").
		SetLegalName("性能测试").
		SetEmail("perf@test.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(ctx)

	// Create multiple positions for realistic scenario
	for i := 0; i < 10; i++ {
		startDate := time.Date(2020+i, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2020+i, 12, 31, 23, 59, 59, 0, time.UTC)

		client.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("职位" + string(rune(i+'1'))).
			SetDepartment("技术部").
			SetJobLevel("SENIOR").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(startDate).
			SetEndDate(endDate).
			SetChangeReason("变更").
			SetIsRetroactive(false).
			SaveX(ctx)
	}

	queryDate := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetPositionAsOfDate(ctx, employee.ID, queryDate)
		if err != nil {
			b.Fatal(err)
		}
	}
}
