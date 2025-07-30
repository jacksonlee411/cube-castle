// internal/service/neo4j_service_test.go
package service

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockNeo4jDriver provides a mock implementation of neo4j.DriverWithContext
type MockNeo4jDriver struct {
	mock.Mock
}

func (m *MockNeo4jDriver) NewSession(ctx context.Context, config neo4j.SessionConfig) neo4j.SessionWithContext {
	args := m.Called(ctx, config)
	return args.Get(0).(neo4j.SessionWithContext)
}

func (m *MockNeo4jDriver) VerifyConnectivity(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNeo4jDriver) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockNeo4jDriver) Target() neo4j.ServerInfo {
	args := m.Called()
	return args.Get(0).(neo4j.ServerInfo)
}

func (m *MockNeo4jDriver) IsEncrypted() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNeo4jDriver) ExecuteQueryBookmarkManager() neo4j.BookmarkManager {
	args := m.Called()
	return args.Get(0).(neo4j.BookmarkManager)
}

func (m *MockNeo4jDriver) ExecuteQuery(ctx context.Context, query string, parameters map[string]any, config neo4j.ExecuteQueryConfiguration) (*neo4j.EagerResult, error) {
	args := m.Called(ctx, query, parameters, config)
	return args.Get(0).(*neo4j.EagerResult), args.Error(1)
}

// MockNeo4jSession provides a mock implementation of neo4j.SessionWithContext
type MockNeo4jSession struct {
	mock.Mock
}

func (m *MockNeo4jSession) LastBookmarks() neo4j.Bookmarks {
	args := m.Called()
	return args.Get(0).(neo4j.Bookmarks)
}

func (m *MockNeo4jSession) BeginTransaction(ctx context.Context, config ...neo4j.TransactionConfig) (neo4j.ExplicitTransaction, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(neo4j.ExplicitTransaction), args.Error(1)
}

func (m *MockNeo4jSession) ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork, config ...neo4j.TransactionConfig) (any, error) {
	args := m.Called(ctx, work, config)
	return args.Get(0), args.Error(1)
}

func (m *MockNeo4jSession) ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork, config ...neo4j.TransactionConfig) (any, error) {
	args := m.Called(ctx, work, config)
	return args.Get(0), args.Error(1)
}

func (m *MockNeo4jSession) Run(ctx context.Context, query string, params map[string]any, config ...neo4j.TransactionConfig) (neo4j.ResultWithContext, error) {
	args := m.Called(ctx, query, params, config)
	return args.Get(0).(neo4j.ResultWithContext), args.Error(1)
}

func (m *MockNeo4jSession) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockNeo4jResult provides a mock implementation of neo4j.ResultWithContext
type MockNeo4jResult struct {
	mock.Mock
}

func (m *MockNeo4jResult) Keys() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockNeo4jResult) Next(ctx context.Context) bool {
	args := m.Called(ctx)
	return args.Bool(0)
}

func (m *MockNeo4jResult) NextRecord(ctx context.Context, record **neo4j.Record) bool {
	args := m.Called(ctx, record)
	return args.Bool(0)
}

func (m *MockNeo4jResult) PeekRecord(ctx context.Context, record **neo4j.Record) bool {
	args := m.Called(ctx, record)
	return args.Bool(0)
}

func (m *MockNeo4jResult) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNeo4jResult) Record() *neo4j.Record {
	args := m.Called()
	return args.Get(0).(*neo4j.Record)
}

func (m *MockNeo4jResult) Collect(ctx context.Context) ([]*neo4j.Record, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*neo4j.Record), args.Error(1)
}

func (m *MockNeo4jResult) Single(ctx context.Context) (*neo4j.Record, error) {
	args := m.Called(ctx)
	return args.Get(0).(*neo4j.Record), args.Error(1)
}

func (m *MockNeo4jResult) Consume(ctx context.Context) (neo4j.ResultSummary, error) {
	args := m.Called(ctx)
	return args.Get(0).(neo4j.ResultSummary), args.Error(1)
}

func (m *MockNeo4jResult) IsOpen() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockNeo4jNode provides a mock implementation of neo4j.Node
type MockNeo4jNode struct {
	ElementId string
	Id        int64
	Labels    []string
	Props     map[string]any
}

func (m *MockNeo4jNode) GetElementId() string {
	return m.ElementId
}

func (m *MockNeo4jNode) GetId() int64 {
	return m.Id
}

func (m *MockNeo4jNode) GetLabels() []string {
	return m.Labels
}

func (m *MockNeo4jNode) GetProperties() map[string]any {
	return m.Props
}

// Neo4jServiceTestSuite provides test suite for Neo4jService
type Neo4jServiceTestSuite struct {
	suite.Suite
	service    *Neo4jService
	mockDriver *MockNeo4jDriver
	mockSession *MockNeo4jSession
	mockResult *MockNeo4jResult
	ctx        context.Context
	logger     *log.Logger
}

// SetupSuite runs once before all tests
func (suite *Neo4jServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = log.New(os.Stdout, "TEST: ", log.LstdFlags)
	
	suite.mockDriver = &MockNeo4jDriver{}
	suite.mockSession = &MockNeo4jSession{}
	suite.mockResult = &MockNeo4jResult{}
	
	// Create service with mock driver
	suite.service = &Neo4jService{
		driver: suite.mockDriver,
		logger: suite.logger,
	}
}

// SetupTest runs before each test
func (suite *Neo4jServiceTestSuite) SetupTest() {
	// Reset mocks
	suite.mockDriver.ExpectedCalls = nil
	suite.mockSession.ExpectedCalls = nil
	suite.mockResult.ExpectedCalls = nil
}

// TestSyncEmployee tests employee node synchronization
func (suite *Neo4jServiceTestSuite) TestSyncEmployee() {
	employee := EmployeeNode{
		ID:         "1",
		EmployeeID: "EMP001",
		LegalName:  "张三",
		Email:      "zhang.san@company.com",
		Status:     "ACTIVE",
		HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)

	// Test successful sync
	err := suite.service.SyncEmployee(suite.ctx, employee)

	assert.NoError(suite.T(), err)
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
}

// TestSyncPosition tests position node synchronization
func (suite *Neo4jServiceTestSuite) TestSyncPosition() {
	position := PositionNode{
		ID:            "pos1",
		PositionTitle: "软件工程师",
		Department:    "技术部",
		JobLevel:      "INTERMEDIATE",
		Location:      "北京",
		EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:       nil,
	}

	employeeID := "EMP001"

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)

	// Test successful sync
	err := suite.service.SyncPosition(suite.ctx, position, employeeID)

	assert.NoError(suite.T(), err)
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
}

// TestCreateReportingRelationship tests reporting relationship creation
func (suite *Neo4jServiceTestSuite) TestCreateReportingRelationship() {
	managerID := "MGR001"
	reporteeID := "EMP001"

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)

	// Test successful relationship creation
	err := suite.service.CreateReportingRelationship(suite.ctx, managerID, reporteeID)

	assert.NoError(suite.T(), err)
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
}

// TestFindReportingPath tests path finding between employees
func (suite *Neo4jServiceTestSuite) TestFindReportingPath() {
	fromEmployeeID := "EMP001"
	toEmployeeID := "EMP002"

	// Create mock record with path data
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{
		nil, // path
		int64(2), // distance
		[]interface{}{ // employees
			map[string]interface{}{
				"employee_id": "EMP001",
				"legal_name":  "张三",
				"email":       "zhang.san@company.com",
			},
			map[string]interface{}{
				"employee_id": "EMP002",
				"legal_name":  "李四",
				"email":       "li.si@company.com",
			},
		},
	}
	mockRecord.Keys = []string{"path", "distance", "employees"}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return(mockRecord, nil)

	// Test successful path finding
	path, err := suite.service.FindReportingPath(suite.ctx, fromEmployeeID, toEmployeeID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), path)
	assert.Equal(suite.T(), 2, path.Distance)
	assert.Equal(suite.T(), "REPORTS_TO", path.PathType)
	assert.Len(suite.T(), path.Path, 2)
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestFindReportingPath_NoPath tests path finding when no path exists
func (suite *Neo4jServiceTestSuite) TestFindReportingPath_NoPath() {
	fromEmployeeID := "EMP001"
	toEmployeeID := "EMP002"

	// Setup mock expectations for no records found
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return((*neo4j.Record)(nil), neo4j.ErrNoRecordsFound)

	// Test no path found
	path, err := suite.service.FindReportingPath(suite.ctx, fromEmployeeID, toEmployeeID)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), path)
	assert.Contains(suite.T(), err.Error(), "no path found")
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestGetReportingHierarchy tests retrieving reporting hierarchy
func (suite *Neo4jServiceTestSuite) TestGetReportingHierarchy() {
	managerID := "MGR001"
	maxDepth := 3

	// Create mock nodes
	managerNode := &MockNeo4jNode{
		ElementId: "manager1",
		Props: map[string]any{
			"id":          "mgr1",
			"employee_id": "MGR001",
			"legal_name":  "管理者一",
			"email":       "manager1@company.com",
			"status":      "ACTIVE",
		},
	}

	directReportNode := &MockNeo4jNode{
		ElementId: "emp1",
		Props: map[string]any{
			"id":          "emp1",
			"employee_id": "EMP001",
			"legal_name":  "员工一",
			"email":       "emp1@company.com",
			"status":      "ACTIVE",
		},
	}

	// Create mock record
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{
		managerNode,
		[]interface{}{directReportNode},
		[]interface{}{directReportNode},
	}
	mockRecord.Keys = []string{"manager", "direct_reports", "all_reports"}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return(mockRecord, nil)

	// Test successful hierarchy retrieval
	hierarchy, err := suite.service.GetReportingHierarchy(suite.ctx, managerID, maxDepth)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), hierarchy)
	assert.Equal(suite.T(), "MGR001", hierarchy.Manager.EmployeeID)
	assert.Equal(suite.T(), "管理者一", hierarchy.Manager.LegalName)
	assert.Len(suite.T(), hierarchy.DirectReports, 1)
	assert.Equal(suite.T(), "EMP001", hierarchy.DirectReports[0].EmployeeID)
	assert.Equal(suite.T(), maxDepth, hierarchy.Depth)
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestFindCommonManager tests finding common manager for employees
func (suite *Neo4jServiceTestSuite) TestFindCommonManager() {
	employeeIDs := []string{"EMP001", "EMP002", "EMP003"}

	// Create mock manager node
	managerNode := &MockNeo4jNode{
		ElementId: "manager1",
		Props: map[string]any{
			"id":          "mgr1",
			"employee_id": "MGR001",
			"legal_name":  "共同管理者",
			"email":       "common.manager@company.com",
			"status":      "ACTIVE",
		},
	}

	// Create mock record
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{managerNode}
	mockRecord.Keys = []string{"manager"}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return(mockRecord, nil)

	// Test successful common manager finding
	manager, err := suite.service.FindCommonManager(suite.ctx, employeeIDs)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), manager)
	assert.Equal(suite.T(), "MGR001", manager.EmployeeID)
	assert.Equal(suite.T(), "共同管理者", manager.LegalName)
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestFindCommonManager_NoCommonManager tests when no common manager exists
func (suite *Neo4jServiceTestSuite) TestFindCommonManager_NoCommonManager() {
	employeeIDs := []string{"EMP001", "EMP002"}

	// Setup mock expectations for no records found
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return((*neo4j.Record)(nil), neo4j.ErrNoRecordsFound)

	// Test no common manager found
	manager, err := suite.service.FindCommonManager(suite.ctx, employeeIDs)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), manager)
	assert.Contains(suite.T(), err.Error(), "no common manager found")
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestGetDepartmentStructure tests department structure retrieval
func (suite *Neo4jServiceTestSuite) TestGetDepartmentStructure() {
	rootDepartment := "技术部"

	// Create mock department node
	deptNode := &MockNeo4jNode{
		ElementId: "dept1",
		Props: map[string]any{
			"id":   "dept1",
			"name": "技术部",
		},
	}

	// Create mock record
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{
		deptNode,
		[]interface{}{deptNode},
		[]interface{}{},
	}
	mockRecord.Keys = []string{"root", "departments", "employees"}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return(mockRecord, nil)

	// Test successful department structure retrieval
	dept, err := suite.service.GetDepartmentStructure(suite.ctx, rootDepartment)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), dept)
	assert.Equal(suite.T(), "技术部", dept.Name)
	assert.Equal(suite.T(), "dept1", dept.ID)
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestNodeToEmployee tests employee node conversion
func (suite *Neo4jServiceTestSuite) TestNodeToEmployee() {
	node := &MockNeo4jNode{
		ElementId: "emp1",
		Props: map[string]any{
			"id":          "emp1",
			"employee_id": "EMP001",
			"legal_name":  "张三",
			"email":       "zhang.san@company.com",
			"status":      "ACTIVE",
			"hire_date":   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			"custom_prop": "custom_value",
		},
	}

	employee := suite.service.nodeToEmployee(node)

	assert.Equal(suite.T(), "emp1", employee.ID)
	assert.Equal(suite.T(), "EMP001", employee.EmployeeID)
	assert.Equal(suite.T(), "张三", employee.LegalName)
	assert.Equal(suite.T(), "zhang.san@company.com", employee.Email)
	assert.Equal(suite.T(), "ACTIVE", employee.Status)
	assert.Equal(suite.T(), time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), employee.HireDate)
	assert.Equal(suite.T(), "custom_value", employee.Properties["custom_prop"])
}

// TestNodeToDepartment tests department node conversion
func (suite *Neo4jServiceTestSuite) TestNodeToDepartment() {
	parentId := "parent1"
	managerId := "mgr1"
	
	node := &MockNeo4jNode{
		ElementId: "dept1",
		Props: map[string]any{
			"id":         "dept1",
			"name":       "技术部",
			"parent_id":  parentId,
			"manager_id": managerId,
			"location":   "北京",
		},
	}

	dept := suite.service.nodeToDepartment(node)

	assert.Equal(suite.T(), "dept1", dept.ID)
	assert.Equal(suite.T(), "技术部", dept.Name)
	assert.NotNil(suite.T(), dept.ParentID)
	assert.Equal(suite.T(), parentId, *dept.ParentID)
	assert.NotNil(suite.T(), dept.ManagerID)
	assert.Equal(suite.T(), managerId, *dept.ManagerID)
	assert.Equal(suite.T(), "北京", dept.Properties["location"])
}

// TestClose tests service cleanup
func (suite *Neo4jServiceTestSuite) TestClose() {
	// Setup mock expectations
	suite.mockDriver.On("Close", suite.ctx).Return(nil)

	// Test successful close
	err := suite.service.Close(suite.ctx)

	assert.NoError(suite.T(), err)
	suite.mockDriver.AssertExpectations(suite.T())
}

// TestSyncEmployeeWithError tests employee sync with database error
func (suite *Neo4jServiceTestSuite) TestSyncEmployeeWithError() {
	employee := EmployeeNode{
		ID:         "1",
		EmployeeID: "EMP001",
		LegalName:  "张三",
		Email:      "zhang.san@company.com",
		Status:     "ACTIVE",
		HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Setup mock expectations for error
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, assert.AnError)

	// Test error handling
	err := suite.service.SyncEmployee(suite.ctx, employee)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to sync employee")
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
}

// TestPerformanceWithLargeDataset tests performance with large employee dataset
func (suite *Neo4jServiceTestSuite) TestPerformanceWithLargeDataset() {
	// Simulate large dataset scenario
	employeeIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		employeeIDs[i] = fmt.Sprintf("EMP%03d", i+1)
	}

	// Create mock manager node
	managerNode := &MockNeo4jNode{
		ElementId: "manager1",
		Props: map[string]any{
			"id":          "mgr1",
			"employee_id": "MGR001",
			"legal_name":  "大团队管理者",
			"email":       "big.manager@company.com",
			"status":      "ACTIVE",
		},
	}

	// Create mock record
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{managerNode}
	mockRecord.Keys = []string{"manager"}

	// Setup mock expectations
	suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
	suite.mockSession.On("Close", suite.ctx).Return(nil)
	suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	suite.mockResult.On("Single", suite.ctx).Return(mockRecord, nil)

	// Measure performance
	start := time.Now()
	manager, err := suite.service.FindCommonManager(suite.ctx, employeeIDs)
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), manager)
	assert.Less(suite.T(), duration, 100*time.Millisecond, "Query should complete within 100ms for large dataset")
	
	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
	suite.mockResult.AssertExpectations(suite.T())
}

// TestConcurrentQueries tests concurrent access to Neo4j service
func (suite *Neo4jServiceTestSuite) TestConcurrentQueries() {
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	employee := EmployeeNode{
		ID:         "1",
		EmployeeID: "EMP001",
		LegalName:  "并发测试员工",
		Email:      "concurrent@company.com",
		Status:     "ACTIVE",
		HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Setup mock expectations for multiple concurrent calls
	for i := 0; i < numGoroutines; i++ {
		suite.mockDriver.On("NewSession", suite.ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(suite.mockSession)
		suite.mockSession.On("Close", suite.ctx).Return(nil)
		suite.mockSession.On("Run", suite.ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(suite.mockResult, nil)
	}

	// Test concurrent employee sync operations
	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := suite.service.SyncEmployee(suite.ctx, employee)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err)
	}

	suite.mockDriver.AssertExpectations(suite.T())
	suite.mockSession.AssertExpectations(suite.T())
}

// TestNeo4jServiceSuite runs the test suite
func TestNeo4jServiceSuite(t *testing.T) {
	suite.Run(t, new(Neo4jServiceTestSuite))
}

// Benchmark tests for Neo4j service performance
func BenchmarkSyncEmployee(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	mockDriver := &MockNeo4jDriver{}
	mockSession := &MockNeo4jSession{}
	mockResult := &MockNeo4jResult{}
	
	service := &Neo4jService{
		driver: mockDriver,
		logger: logger,
	}
	
	ctx := context.Background()
	employee := EmployeeNode{
		ID:         "bench1",
		EmployeeID: "BENCH001",
		LegalName:  "性能测试员工",
		Email:      "bench@test.com",
		Status:     "ACTIVE",
		HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	// Setup expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mockDriver.On("NewSession", ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(mockSession)
		mockSession.On("Close", ctx).Return(nil)
		mockSession.On("Run", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(mockResult, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := service.SyncEmployee(ctx, employee)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFindReportingPath(b *testing.B) {
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	mockDriver := &MockNeo4jDriver{}
	mockSession := &MockNeo4jSession{}
	mockResult := &MockNeo4jResult{}
	
	service := &Neo4jService{
		driver: mockDriver,
		logger: logger,
	}
	
	ctx := context.Background()

	// Create mock record with path data
	mockRecord := &neo4j.Record{}
	mockRecord.Values = []interface{}{
		nil, // path
		int64(2), // distance
		[]interface{}{ // employees
			map[string]interface{}{
				"employee_id": "EMP001",
				"legal_name":  "员工一",
				"email":       "emp1@company.com",
			},
			map[string]interface{}{
				"employee_id": "EMP002",
				"legal_name":  "员工二",
				"email":       "emp2@company.com",
			},
		},
	}
	mockRecord.Keys = []string{"path", "distance", "employees"}

	// Setup expectations for all benchmark iterations
	for i := 0; i < b.N; i++ {
		mockDriver.On("NewSession", ctx, mock.AnythingOfType("neo4j.SessionConfig")).Return(mockSession)
		mockSession.On("Close", ctx).Return(nil)
		mockSession.On("Run", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.Anything).Return(mockResult, nil)
		mockResult.On("Single", ctx).Return(mockRecord, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.FindReportingPath(ctx, "EMP001", "EMP002")
		if err != nil {
			b.Fatal(err)
		}
	}
}