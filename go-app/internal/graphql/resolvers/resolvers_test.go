// internal/graphql/resolvers/resolvers_test.go
package resolvers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/enttest"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
	_ "github.com/mattn/go-sqlite3"
)

// MockTemporalQueryService provides a mock implementation for TemporalQueryService
type MockTemporalQueryService struct {
	mock.Mock
}

func (m *MockTemporalQueryService) GetPositionAsOfDate(ctx context.Context, tenantID, employeeID uuid.UUID, asOfDate time.Time) (*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, asOfDate)
	return args.Get(0).(*service.PositionSnapshot), args.Error(1)
}

func (m *MockTemporalQueryService) GetPositionTimeline(ctx context.Context, tenantID, employeeID uuid.UUID, fromDate, toDate *time.Time) ([]*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, fromDate, toDate)
	return args.Get(0).([]*service.PositionSnapshot), args.Error(1)
}

func (m *MockTemporalQueryService) ValidateTemporalConsistency(ctx context.Context, tenantID, employeeID uuid.UUID, effectiveDate time.Time) error {
	args := m.Called(ctx, tenantID, employeeID, effectiveDate)
	return args.Error(0)
}

func (m *MockTemporalQueryService) CreatePositionSnapshot(ctx context.Context, tenantID, employeeID uuid.UUID, snapshotDate time.Time) (*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, snapshotDate)
	return args.Get(0).(*service.PositionSnapshot), args.Error(1)
}

// MockWorkflowClient provides a mock implementation for Temporal workflow client
type MockWorkflowClient struct {
	mock.Mock
}

func (m *MockWorkflowClient) StartWorkflow(ctx context.Context, workflowID string, request interface{}) error {
	args := m.Called(ctx, workflowID, request)
	return args.Error(0)
}

func (m *MockWorkflowClient) GetWorkflowStatus(ctx context.Context, workflowID string) (*workflow.WorkflowStatus, error) {
	args := m.Called(ctx, workflowID)
	return args.Get(0).(*workflow.WorkflowStatus), args.Error(1)
}

// MockSAMService provides a mock implementation for SAM service
type MockSAMService struct {
	mock.Mock
}

func (m *MockSAMService) GenerateSituationalContext(ctx context.Context) (*service.SituationalContext, error) {
	args := m.Called(ctx)
	return args.Get(0).(*service.SituationalContext), args.Error(1)
}

// PositionHistoryResolverTestSuite provides test suite for PositionHistoryResolver
type PositionHistoryResolverTestSuite struct {
	suite.Suite
	resolver           *PositionHistoryResolver
	entClient          *ent.Client
	mockTemporalQuery  *MockTemporalQueryService
	mockWorkflowClient *MockWorkflowClient
	logger             *logging.StructuredLogger
	ctx                context.Context
}

// SetupSuite runs once before all tests
func (suite *PositionHistoryResolverTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = &logging.StructuredLogger{} // Simplified logger for tests
	
	// Create in-memory SQLite database for testing
	suite.entClient = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	
	// Create mocks
	suite.mockTemporalQuery = &MockTemporalQueryService{}
	suite.mockWorkflowClient = &MockWorkflowClient{}
	
	// Initialize resolver
	suite.resolver = NewPositionHistoryResolver(
		suite.entClient,
		suite.mockTemporalQuery,
		suite.mockWorkflowClient,
		suite.logger,
	)
}

// TearDownSuite runs once after all tests
func (suite *PositionHistoryResolverTestSuite) TearDownSuite() {
	suite.entClient.Close()
}

// SetupTest runs before each test
func (suite *PositionHistoryResolverTestSuite) SetupTest() {
	// Reset mock expectations
	suite.mockTemporalQuery.ExpectedCalls = nil
	suite.mockWorkflowClient.ExpectedCalls = nil
}

// TestCurrentPosition tests the CurrentPosition resolver
func (suite *PositionHistoryResolverTestSuite) TestCurrentPosition() {
	employeeID := uuid.New()
	employee := &Employee{
		ID:         employeeID.String(),
		EmployeeID: "EMP001",
		LegalName:  "张三",
		Email:      "zhang.san@company.com",
	}

	// Create mock position snapshot
	snapshot := &service.PositionSnapshot{
		PositionHistoryID: uuid.New(),
		EmployeeID:        employeeID,
		PositionTitle:     "软件工程师",
		Department:        "技术部",
		JobLevel:          "INTERMEDIATE",
		Location:          "北京",
		EmploymentType:    "FULL_TIME",
		EffectiveDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:      "入职",
		IsRetroactive:     false,
		MinSalary:         &[]float64{8000}[0],
		MaxSalary:         &[]float64{12000}[0],
		Currency:          &[]string{"CNY"}[0],
	}

	// Setup mock expectations
	suite.mockTemporalQuery.On("GetPositionAsOfDate", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, mock.AnythingOfType("time.Time")).Return(snapshot, nil)

	// Test current position query
	position, err := suite.resolver.CurrentPosition(suite.ctx, employee, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), position)
	assert.Equal(suite.T(), snapshot.PositionHistoryID.String(), position.ID)
	assert.Equal(suite.T(), snapshot.PositionTitle, position.PositionTitle)
	assert.Equal(suite.T(), snapshot.Department, position.Department)
	assert.Equal(suite.T(), EmploymentType(snapshot.EmploymentType), position.EmploymentType)
	assert.Equal(suite.T(), snapshot.IsRetroactive, position.IsRetroactive)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestCurrentPositionWithDate tests the CurrentPosition resolver with specific date
func (suite *PositionHistoryResolverTestSuite) TestCurrentPositionWithDate() {
	employeeID := uuid.New()
	employee := &Employee{
		ID:         employeeID.String(),
		EmployeeID: "EMP001",
		LegalName:  "张三",
		Email:      "zhang.san@company.com",
	}

	asOfDate := "2021-06-15"
	expectedDate := time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)

	snapshot := &service.PositionSnapshot{
		PositionHistoryID: uuid.New(),
		EmployeeID:        employeeID,
		PositionTitle:     "高级软件工程师",
		Department:        "技术部",
		JobLevel:          "SENIOR",
		EffectiveDate:     expectedDate,
		ChangeReason:      "晋升",
		IsRetroactive:     false,
	}

	// Setup mock expectations with specific date
	suite.mockTemporalQuery.On("GetPositionAsOfDate", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, expectedDate).Return(snapshot, nil)

	// Test current position query with date
	position, err := suite.resolver.CurrentPosition(suite.ctx, employee, &asOfDate)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), position)
	assert.Equal(suite.T(), snapshot.PositionTitle, position.PositionTitle)
	assert.Equal(suite.T(), snapshot.JobLevel, position.JobLevel)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestCurrentPositionInvalidDate tests the CurrentPosition resolver with invalid date
func (suite *PositionHistoryResolverTestSuite) TestCurrentPositionInvalidDate() {
	employee := &Employee{
		ID: uuid.New().String(),
	}

	invalidDate := "invalid-date"

	// Test current position query with invalid date
	position, err := suite.resolver.CurrentPosition(suite.ctx, employee, &invalidDate)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), position)
	assert.Contains(suite.T(), err.Error(), "invalid date format")
}

// TestPositionHistory tests the PositionHistory resolver
func (suite *PositionHistoryResolverTestSuite) TestPositionHistory() {
	employeeID := uuid.New()
	employee := &Employee{
		ID:         employeeID.String(),
		EmployeeID: "EMP001",
		LegalName:  "张三",
		Email:      "zhang.san@company.com",
	}

	// Create mock position snapshots
	snapshots := []*service.PositionSnapshot{
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "软件工程师",
			Department:        "技术部",
			JobLevel:          "INTERMEDIATE",
			EffectiveDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:           &[]time.Time{time.Date(2021, 6, 30, 23, 59, 59, 0, time.UTC)}[0],
			ChangeReason:      "入职",
			IsRetroactive:     false,
		},
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "高级软件工程师",
			Department:        "技术部",
			JobLevel:          "SENIOR",
			EffectiveDate:     time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:      "晋升",
			IsRetroactive:     false,
		},
	}

	// Setup mock expectations
	suite.mockTemporalQuery.On("GetPositionTimeline", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(snapshots, nil)

	// Test position history query
	connection, err := suite.resolver.PositionHistory(suite.ctx, employee, nil, nil, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), connection)
	assert.Len(suite.T(), connection.Edges, 2)
	assert.Equal(suite.T(), 2, connection.TotalCount)
	
	// Verify first position
	firstPos := connection.Edges[0].Node
	assert.Equal(suite.T(), snapshots[0].PositionTitle, firstPos.PositionTitle)
	assert.Equal(suite.T(), snapshots[0].JobLevel, firstPos.JobLevel)
	assert.NotNil(suite.T(), firstPos.EndDate)
	
	// Verify second position
	secondPos := connection.Edges[1].Node
	assert.Equal(suite.T(), snapshots[1].PositionTitle, secondPos.PositionTitle)
	assert.Equal(suite.T(), snapshots[1].JobLevel, secondPos.JobLevel)
	assert.Nil(suite.T(), secondPos.EndDate)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestPositionHistoryWithDateRange tests the PositionHistory resolver with date filtering
func (suite *PositionHistoryResolverTestSuite) TestPositionHistoryWithDateRange() {
	employeeID := uuid.New()
	employee := &Employee{
		ID: employeeID.String(),
	}

	fromDate := "2021-01-01"
	toDate := "2021-12-31"
	expectedFromDate := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	expectedToDate := time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)

	snapshots := []*service.PositionSnapshot{
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "高级工程师",
			Department:        "技术部",
			EffectiveDate:     time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Setup mock expectations with date range
	suite.mockTemporalQuery.On("GetPositionTimeline", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, &expectedFromDate, &expectedToDate).Return(snapshots, nil)

	// Test position history query with date range
	connection, err := suite.resolver.PositionHistory(suite.ctx, employee, &fromDate, &toDate, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), connection)
	assert.Len(suite.T(), connection.Edges, 1)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestPositionHistoryWithLimit tests the PositionHistory resolver with limit
func (suite *PositionHistoryResolverTestSuite) TestPositionHistoryWithLimit() {
	employeeID := uuid.New()
	employee := &Employee{
		ID: employeeID.String(),
	}

	limit := 1

	snapshots := []*service.PositionSnapshot{
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "工程师1",
			Department:        "技术部",
		},
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "工程师2",
			Department:        "技术部",
		},
	}

	// Setup mock expectations
	suite.mockTemporalQuery.On("GetPositionTimeline", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(snapshots, nil)

	// Test position history query with limit
	connection, err := suite.resolver.PositionHistory(suite.ctx, employee, nil, nil, &limit)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), connection)
	assert.Len(suite.T(), connection.Edges, 1) // Should be limited to 1
	assert.Equal(suite.T(), snapshots[0].PositionTitle, connection.Edges[0].Node.PositionTitle)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestPositionTimeline tests the PositionTimeline resolver
func (suite *PositionHistoryResolverTestSuite) TestPositionTimeline() {
	employeeID := uuid.New()
	employee := &Employee{
		ID: employeeID.String(),
	}

	snapshots := []*service.PositionSnapshot{
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "初级工程师",
			Department:        "技术部",
			EffectiveDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "中级工程师",
			Department:        "技术部",
			EffectiveDate:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			PositionHistoryID: uuid.New(),
			EmployeeID:        employeeID,
			PositionTitle:     "高级工程师",
			Department:        "技术部",
			EffectiveDate:     time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Setup mock expectations
	suite.mockTemporalQuery.On("GetPositionTimeline", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(snapshots, nil)

	// Test position timeline query
	positions, err := suite.resolver.PositionTimeline(suite.ctx, employee, nil)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), positions)
	assert.Len(suite.T(), positions, 3)
	
	// Verify chronological order
	assert.Equal(suite.T(), "初级工程师", positions[0].PositionTitle)
	assert.Equal(suite.T(), "中级工程师", positions[1].PositionTitle)
	assert.Equal(suite.T(), "高级工程师", positions[2].PositionTitle)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestPositionTimelineWithMaxEntries tests the PositionTimeline resolver with max entries limit
func (suite *PositionHistoryResolverTestSuite) TestPositionTimelineWithMaxEntries() {
	employeeID := uuid.New()
	employee := &Employee{
		ID: employeeID.String(),
	}

	maxEntries := 2

	snapshots := []*service.PositionSnapshot{
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "职位1"},
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "职位2"},
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "职位3"},
	}

	// Setup mock expectations
	suite.mockTemporalQuery.On("GetPositionTimeline", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(snapshots, nil)

	// Test position timeline query with max entries
	positions, err := suite.resolver.PositionTimeline(suite.ctx, employee, &maxEntries)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), positions)
	assert.Len(suite.T(), positions, 2) // Should be limited to 2
	assert.Equal(suite.T(), "职位1", positions[0].PositionTitle)
	assert.Equal(suite.T(), "职位2", positions[1].PositionTitle)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestCreatePositionChange tests the CreatePositionChange mutation
func (suite *PositionHistoryResolverTestSuite) TestCreatePositionChange() {
	employeeID := uuid.New()
	
	input := CreatePositionChangeInput{
		EmployeeID: employeeID.String(),
		PositionData: PositionDataInput{
			PositionTitle:  "高级软件工程师",
			Department:     "技术部",
			JobLevel:       &[]string{"SENIOR"}[0],
			Location:       &[]string{"北京"}[0],
			EmploymentType: EmploymentTypeFullTime,
			MinSalary:      &[]float64{10000}[0],
			MaxSalary:      &[]float64{15000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		EffectiveDate: "2023-01-01T00:00:00Z",
		ChangeReason:  &[]string{"晋升"}[0],
		IsRetroactive: false,
	}

	// Setup mock expectations
	suite.mockWorkflowClient.On("StartWorkflow", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("workflow.PositionChangeRequest")).Return(nil)

	// Test create position change
	payload, err := suite.resolver.CreatePositionChange(suite.ctx, input)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), payload)
	assert.NotNil(suite.T(), payload.WorkflowID)
	assert.Empty(suite.T(), payload.Errors)
	
	suite.mockWorkflowClient.AssertExpectations(suite.T())
}

// TestValidatePositionChange tests the ValidatePositionChange mutation
func (suite *PositionHistoryResolverTestSuite) TestValidatePositionChange() {
	employeeID := uuid.New()
	effectiveDate := "2023-01-01T00:00:00Z"
	expectedDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	// Setup mock expectations for valid position change
	suite.mockTemporalQuery.On("ValidateTemporalConsistency", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, expectedDate).Return(nil)

	// Test validate position change
	validation, err := suite.resolver.ValidatePositionChange(suite.ctx, employeeID.String(), effectiveDate)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), validation)
	assert.True(suite.T(), validation.IsValid)
	assert.Empty(suite.T(), validation.Errors)
	assert.Empty(suite.T(), validation.Warnings)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestValidatePositionChangeWithConflict tests the ValidatePositionChange mutation with temporal conflict
func (suite *PositionHistoryResolverTestSuite) TestValidatePositionChangeWithConflict() {
	employeeID := uuid.New()
	effectiveDate := "2023-01-01T00:00:00Z"
	expectedDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	// Setup mock expectations for temporal conflict
	conflictError := assert.AnError
	suite.mockTemporalQuery.On("ValidateTemporalConsistency", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, expectedDate).Return(conflictError)

	// Test validate position change with conflict
	validation, err := suite.resolver.ValidatePositionChange(suite.ctx, employeeID.String(), effectiveDate)

	assert.NoError(suite.T(), err) // No error from resolver itself
	assert.NotNil(suite.T(), validation)
	assert.False(suite.T(), validation.IsValid)
	assert.Len(suite.T(), validation.Errors, 1)
	assert.Equal(suite.T(), "TEMPORAL_CONFLICT", validation.Errors[0].Code)
	assert.Equal(suite.T(), "effective_date", *validation.Errors[0].Field)
	
	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestConvertSnapshotToGraphQL tests the snapshot conversion helper
func (suite *PositionHistoryResolverTestSuite) TestConvertSnapshotToGraphQL() {
	employeeID := uuid.New()
	positionID := uuid.New()
	reportsToID := uuid.New()
	effectiveDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)

	snapshot := &service.PositionSnapshot{
		PositionHistoryID:   positionID,
		EmployeeID:          employeeID,
		PositionTitle:       "测试工程师",
		Department:          "质量保证部",
		JobLevel:            "INTERMEDIATE",
		Location:            "上海",
		EmploymentType:      "FULL_TIME",
		ReportsToEmployeeID: &reportsToID,
		EffectiveDate:       effectiveDate,
		EndDate:             &endDate,
		ChangeReason:        "内部调动",
		IsRetroactive:       false,
		MinSalary:           &[]float64{9000}[0],
		MaxSalary:           &[]float64{14000}[0],
		Currency:            &[]string{"CNY"}[0],
	}

	// Test conversion
	graphQLPos := convertSnapshotToGraphQL(snapshot)

	assert.NotNil(suite.T(), graphQLPos)
	assert.Equal(suite.T(), positionID.String(), graphQLPos.ID)
	assert.Equal(suite.T(), employeeID.String(), graphQLPos.EmployeeID)
	assert.Equal(suite.T(), "测试工程师", graphQLPos.PositionTitle)
	assert.Equal(suite.T(), "质量保证部", graphQLPos.Department)
	assert.Equal(suite.T(), "INTERMEDIATE", *graphQLPos.JobLevel)
	assert.Equal(suite.T(), "上海", *graphQLPos.Location)
	assert.Equal(suite.T(), EmploymentTypeFullTime, graphQLPos.EmploymentType)
	assert.Equal(suite.T(), reportsToID.String(), *graphQLPos.ReportsToEmployeeID)
	assert.Equal(suite.T(), effectiveDate.Format(time.RFC3339), graphQLPos.EffectiveDate)
	assert.Equal(suite.T(), endDate.Format(time.RFC3339), *graphQLPos.EndDate)
	assert.Equal(suite.T(), "内部调动", *graphQLPos.ChangeReason)
	assert.False(suite.T(), graphQLPos.IsRetroactive)
	assert.Equal(suite.T(), 9000.0, *graphQLPos.MinSalary)
	assert.Equal(suite.T(), 14000.0, *graphQLPos.MaxSalary)
	assert.Equal(suite.T(), "CNY", *graphQLPos.Currency)
}

// TestConcurrentResolverCalls tests concurrent access to resolvers
func (suite *PositionHistoryResolverTestSuite) TestConcurrentResolverCalls() {
	employeeID := uuid.New()
	employee := &Employee{
		ID: employeeID.String(),
	}

	snapshot := &service.PositionSnapshot{
		PositionHistoryID: uuid.New(),
		EmployeeID:        employeeID,
		PositionTitle:     "并发测试工程师",
		Department:        "技术部",
	}

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	// Setup mock expectations for multiple concurrent calls
	for i := 0; i < numGoroutines; i++ {
		suite.mockTemporalQuery.On("GetPositionAsOfDate", suite.ctx, mock.AnythingOfType("uuid.UUID"), employeeID, mock.AnythingOfType("time.Time")).Return(snapshot, nil)
	}

	// Test concurrent resolver calls
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := suite.resolver.CurrentPosition(suite.ctx, employee, nil)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err)
	}

	suite.mockTemporalQuery.AssertExpectations(suite.T())
}

// TestPositionHistoryResolverSuite runs the test suite
func TestPositionHistoryResolverSuite(t *testing.T) {
	suite.Run(t, new(PositionHistoryResolverTestSuite))
}

// SAMResolverTestSuite provides test suite for SAMResolver
type SAMResolverTestSuite struct {
	suite.Suite
	resolver    *SAMResolver
	mockSAM     *MockSAMService
	ctx         context.Context
}

// SetupSuite runs once before all tests
func (suite *SAMResolverTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.mockSAM = &MockSAMService{}
	suite.resolver = NewSAMResolver(suite.mockSAM)
}

// SetupTest runs before each test
func (suite *SAMResolverTestSuite) SetupTest() {
	// Reset mock expectations
	suite.mockSAM.ExpectedCalls = nil
}

// TestGetSituationalContext tests the GetSituationalContext resolver
func (suite *SAMResolverTestSuite) TestGetSituationalContext() {
	// Create mock situational context
	mockContext := &service.SituationalContext{
		Timestamp:  time.Now(),
		AlertLevel: "YELLOW",
		OrganizationHealth: service.OrganizationHealthMetrics{
			OverallHealthScore: 0.75,
			TurnoverRate:       0.12,
			EmployeeEngagement: 0.78,
			ProductivityIndex:  0.82,
			DepartmentHealthMap: map[string]service.DepartmentHealth{
				"技术部": {
					HealthScore:          0.80,
					TurnoverRate:         0.10,
					AverageTenure:        24.5,
					ManagerEffectiveness: 0.85,
				},
			},
			TrendAnalysis: service.HealthTrendAnalysis{
				Trend:           "IMPROVING",
				TrendStrength:   0.65,
				KeyDrivers:      []string{"员工满意度", "管理效能"},
				PredictedHealth: 0.78,
				Confidence:      0.82,
			},
		},
		TalentMetrics: service.TalentManagementMetrics{
			TalentPipelineHealth: 0.72,
			SuccessionReadiness:  0.58,
			SkillGapAnalysis: map[string]float64{
				"技术领导力": 0.35,
				"数据分析":  0.28,
			},
			PerformanceDistribution: service.PerformanceDistribution{
				HighPerformers:  0.25,
				SolidPerformers: 0.65,
				LowPerformers:   0.10,
				PerformanceGaps: []string{"跨团队协作", "技术创新"},
			},
		},
		RiskAssessment: service.RiskAssessmentResult{
			OverallRiskScore: 0.45,
			KeyPersonRisks: []service.KeyPersonRisk{
				{
					EmployeeID:     "emp-001",
					EmployeeName:   "张三",
					Position:       "技术总监",
					Department:     "技术部",
					RiskScore:      0.75,
					RiskFactors:    []string{"单点依赖", "知识垄断"},
					BusinessImpact: "技术决策延迟",
					MitigationSteps: []string{"知识分享", "副手培养"},
				},
			},
		},
		OpportunityAnalysis: service.OpportunityAnalysisResult{
			TalentOptimization: []service.TalentOptimization{
				{
					OpportunityType: "内部晋升优化",
					Description:     "通过内部人才发展减少外部招聘成本",
					AffectedRoles:   []string{"高级工程师", "项目经理"},
					ExpectedBenefit: "降低30%招聘成本",
				},
			},
		},
		Recommendations: []service.StrategicRecommendation{
			{
				ID:             "rec-001",
				Type:           "SHORT_TERM",
				Priority:       "HIGH",
				Category:       "TALENT",
				Title:          "人才梯队建设计划",
				Description:    "建立完善的人才梯队，提升组织韧性",
				BusinessImpact: "确保关键岗位有合适的继任者",
				Confidence:     0.82,
			},
		},
	}

	// Setup mock expectations
	suite.mockSAM.On("GenerateSituationalContext", suite.ctx).Return(mockContext, nil)

	// Test situational context query
	response, err := suite.resolver.GetSituationalContext(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response)
	assert.Equal(suite.T(), "YELLOW", response.AlertLevel)
	assert.Equal(suite.T(), 0.75, response.OrganizationHealth.OverallScore)
	assert.Equal(suite.T(), 0.78, response.OrganizationHealth.EngagementLevel)
	assert.Equal(suite.T(), 0.72, response.TalentMetrics.TalentPipelineHealth)
	assert.Equal(suite.T(), 0.45, response.RiskAssessment.OverallRiskScore)
	assert.Len(suite.T(), response.Recommendations, 1)
	
	suite.mockSAM.AssertExpectations(suite.T())
}

// TestSAMResolverSuite runs the test suite
func TestSAMResolverSuite(t *testing.T) {
	suite.Run(t, new(SAMResolverTestSuite))
}

// Integration tests for resolver error handling

// TestResolverErrorHandling tests various error scenarios
func TestResolverErrorHandling(t *testing.T) {
	ctx := context.Background()
	logger := &logging.StructuredLogger{}
	entClient := enttest.Open(t, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer entClient.Close()

	mockTemporalQuery := &MockTemporalQueryService{}
	mockWorkflowClient := &MockWorkflowClient{}

	resolver := NewPositionHistoryResolver(entClient, mockTemporalQuery, mockWorkflowClient, logger)

	t.Run("CurrentPosition with service error", func(t *testing.T) {
		employeeID := uuid.New()
		employee := &Employee{ID: employeeID.String()}

		// Setup mock to return error
		mockTemporalQuery.On("GetPositionAsOfDate", ctx, mock.AnythingOfType("uuid.UUID"), employeeID, mock.AnythingOfType("time.Time")).Return((*service.PositionSnapshot)(nil), assert.AnError)

		position, err := resolver.CurrentPosition(ctx, employee, nil)

		// Should return nil position but no error (graceful handling)
		assert.NoError(t, err)
		assert.Nil(t, position)
	})

	t.Run("PositionHistory with service error", func(t *testing.T) {
		employeeID := uuid.New()
		employee := &Employee{ID: employeeID.String()}

		// Setup mock to return error
		mockTemporalQuery.On("GetPositionTimeline", ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(([]*service.PositionSnapshot)(nil), assert.AnError)

		connection, err := resolver.PositionHistory(ctx, employee, nil, nil, nil)

		// Should return error for timeline query failure
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("Invalid UUID in employee ID", func(t *testing.T) {
		employee := &Employee{ID: "invalid-uuid"}

		// This should panic due to uuid.MustParse, but in real implementation
		// we would handle this more gracefully
		assert.Panics(t, func() {
			resolver.CurrentPosition(ctx, employee, nil)
		})
	})
}

// Benchmark tests for resolver performance
func BenchmarkCurrentPosition(b *testing.B) {
	ctx := context.Background()
	logger := &logging.StructuredLogger{}
	entClient := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer entClient.Close()

	mockTemporalQuery := &MockTemporalQueryService{}
	mockWorkflowClient := &MockWorkflowClient{}
	resolver := NewPositionHistoryResolver(entClient, mockTemporalQuery, mockWorkflowClient, logger)

	employeeID := uuid.New()
	employee := &Employee{ID: employeeID.String()}
	snapshot := &service.PositionSnapshot{
		PositionHistoryID: uuid.New(),
		EmployeeID:        employeeID,
		PositionTitle:     "性能测试工程师",
		Department:        "技术部",
	}

	// Setup expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mockTemporalQuery.On("GetPositionAsOfDate", ctx, mock.AnythingOfType("uuid.UUID"), employeeID, mock.AnythingOfType("time.Time")).Return(snapshot, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolver.CurrentPosition(ctx, employee, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPositionHistory(b *testing.B) {
	ctx := context.Background()
	logger := &logging.StructuredLogger{}
	entClient := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer entClient.Close()

	mockTemporalQuery := &MockTemporalQueryService{}
	mockWorkflowClient := &MockWorkflowClient{}
	resolver := NewPositionHistoryResolver(entClient, mockTemporalQuery, mockWorkflowClient, logger)

	employeeID := uuid.New()
	employee := &Employee{ID: employeeID.String()}
	snapshots := []*service.PositionSnapshot{
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "测试1"},
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "测试2"},
		{PositionHistoryID: uuid.New(), EmployeeID: employeeID, PositionTitle: "测试3"},
	}

	// Setup expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mockTemporalQuery.On("GetPositionTimeline", ctx, mock.AnythingOfType("uuid.UUID"), employeeID, (*time.Time)(nil), (*time.Time)(nil)).Return(snapshots, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := resolver.PositionHistory(ctx, employee, nil, nil, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}