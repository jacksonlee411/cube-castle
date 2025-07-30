// internal/service/sam_service_test.go
package service

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/enttest"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockNeo4jServiceForSAM provides a mock implementation for Neo4j service in SAM tests
type MockNeo4jServiceForSAM struct {
	mock.Mock
}

func (m *MockNeo4jServiceForSAM) SyncEmployee(ctx context.Context, employee EmployeeNode) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockNeo4jServiceForSAM) SyncPosition(ctx context.Context, position PositionNode, employeeID string) error {
	args := m.Called(ctx, position, employeeID)
	return args.Error(0)
}

func (m *MockNeo4jServiceForSAM) CreateReportingRelationship(ctx context.Context, managerID, reporteeID string) error {
	args := m.Called(ctx, managerID, reporteeID)
	return args.Error(0)
}

func (m *MockNeo4jServiceForSAM) FindReportingPath(ctx context.Context, fromEmployeeID, toEmployeeID string) (*OrganizationalPath, error) {
	args := m.Called(ctx, fromEmployeeID, toEmployeeID)
	return args.Get(0).(*OrganizationalPath), args.Error(1)
}

func (m *MockNeo4jServiceForSAM) GetReportingHierarchy(ctx context.Context, managerID string, maxDepth int) (*ReportingHierarchy, error) {
	args := m.Called(ctx, managerID, maxDepth)
	return args.Get(0).(*ReportingHierarchy), args.Error(1)
}

func (m *MockNeo4jServiceForSAM) FindCommonManager(ctx context.Context, employeeIDs []string) (*EmployeeNode, error) {
	args := m.Called(ctx, employeeIDs)
	return args.Get(0).(*EmployeeNode), args.Error(1)
}

func (m *MockNeo4jServiceForSAM) GetDepartmentStructure(ctx context.Context, rootDepartment string) (*DepartmentNode, error) {
	args := m.Called(ctx, rootDepartment)
	return args.Get(0).(*DepartmentNode), args.Error(1)
}

func (m *MockNeo4jServiceForSAM) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// SAMServiceTestSuite provides test suite for SAMService
type SAMServiceTestSuite struct {
	suite.Suite
	service   *SAMService
	entClient *ent.Client
	mockNeo4j *MockNeo4jServiceForSAM
	ctx       context.Context
	logger    *log.Logger
}

// SetupSuite runs once before all tests
func (suite *SAMServiceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	suite.logger = log.New(os.Stdout, "TEST: ", log.LstdFlags)

	// Create in-memory SQLite database for testing
	suite.entClient = enttest.Open(suite.T(), "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")

	// Create mock Neo4j service
	suite.mockNeo4j = &MockNeo4jServiceForSAM{}

	// Initialize SAM service
	suite.service = NewSAMService(suite.entClient, suite.mockNeo4j, suite.logger)
}

// TearDownSuite runs once after all tests
func (suite *SAMServiceTestSuite) TearDownSuite() {
	suite.entClient.Close()
}

// SetupTest runs before each test
func (suite *SAMServiceTestSuite) SetupTest() {
	// Clean database state
	suite.entClient.PositionHistory.Delete().ExecX(suite.ctx)
	suite.entClient.Employee.Delete().ExecX(suite.ctx)

	// Reset mock expectations
	suite.mockNeo4j.ExpectedCalls = nil
}

// TestGenerateSituationalContext tests the main SAM context generation
func (suite *SAMServiceTestSuite) TestGenerateSituationalContext() {
	// Setup test data
	suite.createTestData()

	// Test situational context generation
	context, err := suite.service.GenerateSituationalContext(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), context)

	// Verify context structure
	assert.NotZero(suite.T(), context.Timestamp)
	assert.NotEmpty(suite.T(), context.AlertLevel)
	assert.Contains(suite.T(), []string{"GREEN", "YELLOW", "ORANGE", "RED"}, context.AlertLevel)

	// Verify organization health metrics
	assert.GreaterOrEqual(suite.T(), context.OrganizationHealth.OverallHealthScore, 0.0)
	assert.LessOrEqual(suite.T(), context.OrganizationHealth.OverallHealthScore, 1.0)
	assert.NotEmpty(suite.T(), context.OrganizationHealth.DepartmentHealthMap)
	assert.NotEmpty(suite.T(), context.OrganizationHealth.TrendAnalysis.KeyDrivers)

	// Verify talent metrics
	assert.GreaterOrEqual(suite.T(), context.TalentMetrics.TalentPipelineHealth, 0.0)
	assert.LessOrEqual(suite.T(), context.TalentMetrics.TalentPipelineHealth, 1.0)
	assert.NotEmpty(suite.T(), context.TalentMetrics.SkillGapAnalysis)
	assert.NotEmpty(suite.T(), context.TalentMetrics.PerformanceDistribution.PerformanceGaps)

	// Verify risk assessment
	assert.GreaterOrEqual(suite.T(), context.RiskAssessment.OverallRiskScore, 0.0)
	assert.LessOrEqual(suite.T(), context.RiskAssessment.OverallRiskScore, 1.0)
	assert.NotEmpty(suite.T(), context.RiskAssessment.KeyPersonRisks)
	assert.NotEmpty(suite.T(), context.RiskAssessment.ComplianceRisks)

	// Verify opportunity analysis
	assert.NotEmpty(suite.T(), context.OpportunityAnalysis.TalentOptimization)
	assert.NotEmpty(suite.T(), context.OpportunityAnalysis.ProcessImprovements)
	assert.NotEmpty(suite.T(), context.OpportunityAnalysis.StructuralChanges)

	// Verify recommendations
	assert.NotEmpty(suite.T(), context.Recommendations)
	for _, rec := range context.Recommendations {
		assert.NotEmpty(suite.T(), rec.ID)
		assert.NotEmpty(suite.T(), rec.Title)
		assert.NotEmpty(suite.T(), rec.Description)
		assert.Contains(suite.T(), []string{"IMMEDIATE", "SHORT_TERM", "LONG_TERM"}, rec.Type)
		assert.Contains(suite.T(), []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, rec.Priority)
		assert.GreaterOrEqual(suite.T(), rec.Confidence, 0.0)
		assert.LessOrEqual(suite.T(), rec.Confidence, 1.0)
	}
}

// TestAnalyzeOrganizationHealth tests organization health analysis
func (suite *SAMServiceTestSuite) TestAnalyzeOrganizationHealth() {
	// Setup test data
	suite.createTestData()

	// Test organization health analysis
	orgHealth, err := suite.service.analyzeOrganizationHealth(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), orgHealth)

	// Verify health metrics
	assert.GreaterOrEqual(suite.T(), orgHealth.OverallHealthScore, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.OverallHealthScore, 1.0)
	assert.GreaterOrEqual(suite.T(), orgHealth.TurnoverRate, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.TurnoverRate, 1.0)
	assert.GreaterOrEqual(suite.T(), orgHealth.EmployeeEngagement, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.EmployeeEngagement, 1.0)
	assert.GreaterOrEqual(suite.T(), orgHealth.ProductivityIndex, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.ProductivityIndex, 1.0)

	// Verify department health map
	assert.NotEmpty(suite.T(), orgHealth.DepartmentHealthMap)
	for deptName, deptHealth := range orgHealth.DepartmentHealthMap {
		assert.NotEmpty(suite.T(), deptName)
		assert.GreaterOrEqual(suite.T(), deptHealth.HealthScore, 0.0)
		assert.LessOrEqual(suite.T(), deptHealth.HealthScore, 1.0)
		assert.GreaterOrEqual(suite.T(), deptHealth.TurnoverRate, 0.0)
		assert.GreaterOrEqual(suite.T(), deptHealth.AverageTenure, 0.0)
		assert.GreaterOrEqual(suite.T(), deptHealth.ManagerEffectiveness, 0.0)
		assert.LessOrEqual(suite.T(), deptHealth.ManagerEffectiveness, 1.0)
	}

	// Verify trend analysis
	assert.Contains(suite.T(), []string{"IMPROVING", "STABLE", "DECLINING"}, orgHealth.TrendAnalysis.Trend)
	assert.GreaterOrEqual(suite.T(), orgHealth.TrendAnalysis.TrendStrength, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.TrendAnalysis.TrendStrength, 1.0)
	assert.NotEmpty(suite.T(), orgHealth.TrendAnalysis.KeyDrivers)
	assert.GreaterOrEqual(suite.T(), orgHealth.TrendAnalysis.Confidence, 0.0)
	assert.LessOrEqual(suite.T(), orgHealth.TrendAnalysis.Confidence, 1.0)
}

// TestAnalyzeTalentMetrics tests talent management analysis
func (suite *SAMServiceTestSuite) TestAnalyzeTalentMetrics() {
	// Setup test data
	suite.createTestData()

	// Test talent metrics analysis
	talentMetrics, err := suite.service.analyzeTalentMetrics(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), talentMetrics)

	// Verify talent metrics
	assert.GreaterOrEqual(suite.T(), talentMetrics.TalentPipelineHealth, 0.0)
	assert.LessOrEqual(suite.T(), talentMetrics.TalentPipelineHealth, 1.0)
	assert.GreaterOrEqual(suite.T(), talentMetrics.SuccessionReadiness, 0.0)
	assert.LessOrEqual(suite.T(), talentMetrics.SuccessionReadiness, 1.0)
	assert.GreaterOrEqual(suite.T(), talentMetrics.LearningDevelopmentROI, 0.0)
	assert.GreaterOrEqual(suite.T(), talentMetrics.InternalMobilityRate, 0.0)
	assert.LessOrEqual(suite.T(), talentMetrics.InternalMobilityRate, 1.0)

	// Verify skill gap analysis
	assert.NotEmpty(suite.T(), talentMetrics.SkillGapAnalysis)
	for skillArea, gapSize := range talentMetrics.SkillGapAnalysis {
		assert.NotEmpty(suite.T(), skillArea)
		assert.GreaterOrEqual(suite.T(), gapSize, 0.0)
		assert.LessOrEqual(suite.T(), gapSize, 1.0)
	}

	// Verify performance distribution
	perf := talentMetrics.PerformanceDistribution
	assert.GreaterOrEqual(suite.T(), perf.HighPerformers, 0.0)
	assert.LessOrEqual(suite.T(), perf.HighPerformers, 1.0)
	assert.GreaterOrEqual(suite.T(), perf.SolidPerformers, 0.0)
	assert.LessOrEqual(suite.T(), perf.SolidPerformers, 1.0)
	assert.GreaterOrEqual(suite.T(), perf.LowPerformers, 0.0)
	assert.LessOrEqual(suite.T(), perf.LowPerformers, 1.0)

	// Performance percentages should approximately sum to 1
	total := perf.HighPerformers + perf.SolidPerformers + perf.LowPerformers
	assert.InDelta(suite.T(), 1.0, total, 0.1)

	assert.NotEmpty(suite.T(), perf.PerformanceGaps)
}

// TestPerformRiskAssessment tests risk assessment functionality
func (suite *SAMServiceTestSuite) TestPerformRiskAssessment() {
	// Test risk assessment
	riskAssessment, err := suite.service.performRiskAssessment(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), riskAssessment)

	// Verify overall risk score
	assert.GreaterOrEqual(suite.T(), riskAssessment.OverallRiskScore, 0.0)
	assert.LessOrEqual(suite.T(), riskAssessment.OverallRiskScore, 1.0)

	// Verify key person risks
	assert.NotEmpty(suite.T(), riskAssessment.KeyPersonRisks)
	for _, risk := range riskAssessment.KeyPersonRisks {
		assert.NotEmpty(suite.T(), risk.EmployeeID)
		assert.NotEmpty(suite.T(), risk.EmployeeName)
		assert.NotEmpty(suite.T(), risk.Position)
		assert.NotEmpty(suite.T(), risk.Department)
		assert.GreaterOrEqual(suite.T(), risk.RiskScore, 0.0)
		assert.LessOrEqual(suite.T(), risk.RiskScore, 1.0)
		assert.NotEmpty(suite.T(), risk.RiskFactors)
		assert.NotEmpty(suite.T(), risk.MitigationSteps)
		assert.NotZero(suite.T(), risk.LastAssessment)
	}

	// Verify compliance risks
	assert.NotEmpty(suite.T(), riskAssessment.ComplianceRisks)
	for _, risk := range riskAssessment.ComplianceRisks {
		assert.NotEmpty(suite.T(), risk.RiskType)
		assert.Contains(suite.T(), []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, risk.Severity)
		assert.NotEmpty(suite.T(), risk.Description)
		assert.NotEmpty(suite.T(), risk.AffectedAreas)
		assert.NotEmpty(suite.T(), risk.RemediationPlan)
		assert.NotZero(suite.T(), risk.Deadline)
	}

	// Verify operational risks
	assert.NotEmpty(suite.T(), riskAssessment.OperationalRisks)
	for _, risk := range riskAssessment.OperationalRisks {
		assert.NotEmpty(suite.T(), risk.RiskCategory)
		assert.NotEmpty(suite.T(), risk.Description)
		assert.GreaterOrEqual(suite.T(), risk.Probability, 0.0)
		assert.LessOrEqual(suite.T(), risk.Probability, 1.0)
		assert.GreaterOrEqual(suite.T(), risk.Impact, 0.0)
		assert.LessOrEqual(suite.T(), risk.Impact, 1.0)
		assert.GreaterOrEqual(suite.T(), risk.RiskScore, 0.0)
		assert.LessOrEqual(suite.T(), risk.RiskScore, 1.0)
		assert.NotEmpty(suite.T(), risk.ContingencyPlan)
	}

	// Verify talent flight risks
	assert.NotEmpty(suite.T(), riskAssessment.TalentFlightRisks)
	for _, risk := range riskAssessment.TalentFlightRisks {
		assert.NotEmpty(suite.T(), risk.EmployeeID)
		assert.NotEmpty(suite.T(), risk.EmployeeName)
		assert.GreaterOrEqual(suite.T(), risk.FlightRisk, 0.0)
		assert.LessOrEqual(suite.T(), risk.FlightRisk, 1.0)
		assert.NotEmpty(suite.T(), risk.RiskIndicators)
		assert.NotEmpty(suite.T(), risk.RetentionActions)
		assert.NotEmpty(suite.T(), risk.TimeFrame)
	}

	// Verify risk mitigation
	assert.NotEmpty(suite.T(), riskAssessment.RiskMitigation)
	for _, mitigation := range riskAssessment.RiskMitigation {
		assert.NotEmpty(suite.T(), mitigation.RiskType)
		assert.NotEmpty(suite.T(), mitigation.MitigationAction)
		assert.GreaterOrEqual(suite.T(), mitigation.Effectiveness, 0.0)
		assert.LessOrEqual(suite.T(), mitigation.Effectiveness, 1.0)
		assert.NotEmpty(suite.T(), mitigation.Timeline)
		assert.NotEmpty(suite.T(), mitigation.ResponsibleParty)
	}
}

// TestAnalyzeOpportunities tests opportunity analysis
func (suite *SAMServiceTestSuite) TestAnalyzeOpportunities() {
	// Test opportunity analysis
	opportunities, err := suite.service.analyzeOpportunities(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), opportunities)

	// Verify talent optimization opportunities
	assert.NotEmpty(suite.T(), opportunities.TalentOptimization)
	for _, opt := range opportunities.TalentOptimization {
		assert.NotEmpty(suite.T(), opt.OpportunityType)
		assert.NotEmpty(suite.T(), opt.Description)
		assert.NotEmpty(suite.T(), opt.AffectedRoles)
		assert.NotEmpty(suite.T(), opt.ExpectedBenefit)
		assert.NotEmpty(suite.T(), opt.ImplementationSteps)
	}

	// Verify process improvements
	assert.NotEmpty(suite.T(), opportunities.ProcessImprovements)
	for _, improvement := range opportunities.ProcessImprovements {
		assert.NotEmpty(suite.T(), improvement.ProcessArea)
		assert.NotEmpty(suite.T(), improvement.CurrentState)
		assert.NotEmpty(suite.T(), improvement.ProposedState)
		assert.GreaterOrEqual(suite.T(), improvement.EfficiencyGain, 0.0)
		assert.NotEmpty(suite.T(), improvement.ImplementationComplexity)
	}

	// Verify structural changes
	assert.NotEmpty(suite.T(), opportunities.StructuralChanges)
	for _, change := range opportunities.StructuralChanges {
		assert.NotEmpty(suite.T(), change.ChangeType)
		assert.NotEmpty(suite.T(), change.Description)
		assert.NotEmpty(suite.T(), change.Rationale)
		assert.NotEmpty(suite.T(), change.AffectedTeams)
		assert.NotEmpty(suite.T(), change.ImplementationPhases)
	}

	// Verify investment priorities
	assert.NotEmpty(suite.T(), opportunities.InvestmentPriorities)
	for _, priority := range opportunities.InvestmentPriorities {
		assert.NotEmpty(suite.T(), priority.InvestmentArea)
		assert.Contains(suite.T(), []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, priority.Priority)
		assert.GreaterOrEqual(suite.T(), priority.EstimatedCost, 0.0)
		assert.GreaterOrEqual(suite.T(), priority.ExpectedROI, 0.0)
		assert.NotEmpty(suite.T(), priority.Justification)
	}

	// Verify capability gaps
	assert.NotEmpty(suite.T(), opportunities.CapabilityGaps)
	for _, gap := range opportunities.CapabilityGaps {
		assert.NotEmpty(suite.T(), gap.CapabilityArea)
		assert.GreaterOrEqual(suite.T(), gap.CurrentLevel, 0.0)
		assert.LessOrEqual(suite.T(), gap.CurrentLevel, 1.0)
		assert.GreaterOrEqual(suite.T(), gap.RequiredLevel, 0.0)
		assert.LessOrEqual(suite.T(), gap.RequiredLevel, 1.0)
		assert.GreaterOrEqual(suite.T(), gap.GapSize, 0.0)
		assert.NotEmpty(suite.T(), gap.ClosureStrategy)
	}
}

// TestGenerateRecommendations tests strategic recommendation generation
func (suite *SAMServiceTestSuite) TestGenerateRecommendations() {
	// Create test metrics
	orgHealth := &OrganizationHealthMetrics{
		OverallHealthScore: 0.65, // Below threshold to trigger recommendation
		TurnoverRate:       0.15,
	}

	talentMetrics := &TalentManagementMetrics{
		SuccessionReadiness: 0.55, // Below threshold to trigger recommendation
	}

	riskAssessment := &RiskAssessmentResult{
		OverallRiskScore: 0.75, // Above threshold to trigger critical recommendation
	}

	opportunities := &OpportunityAnalysisResult{}

	// Test recommendation generation
	recommendations, err := suite.service.generateRecommendations(
		suite.ctx, orgHealth, talentMetrics, riskAssessment, opportunities)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), recommendations)

	// Verify recommendations are sorted by priority and confidence
	for i := 1; i < len(recommendations); i++ {
		prev := recommendations[i-1]
		curr := recommendations[i]

		priorityScore := map[string]int{"CRITICAL": 4, "HIGH": 3, "MEDIUM": 2, "LOW": 1}
		prevScore := priorityScore[prev.Priority]
		currScore := priorityScore[curr.Priority]

		// Current should not have higher priority than previous
		assert.LessOrEqual(suite.T(), currScore, prevScore)

		// If same priority, current should not have higher confidence
		if prevScore == currScore {
			assert.LessOrEqual(suite.T(), curr.Confidence, prev.Confidence)
		}
	}

	// Verify recommendation structure
	for _, rec := range recommendations {
		assert.NotEmpty(suite.T(), rec.ID)
		assert.Contains(suite.T(), []string{"IMMEDIATE", "SHORT_TERM", "LONG_TERM"}, rec.Type)
		assert.Contains(suite.T(), []string{"CRITICAL", "HIGH", "MEDIUM", "LOW"}, rec.Priority)
		assert.Contains(suite.T(), []string{"TALENT", "STRUCTURE", "PROCESS", "RISK", "TECHNOLOGY", "CULTURE"}, rec.Category)
		assert.NotEmpty(suite.T(), rec.Title)
		assert.NotEmpty(suite.T(), rec.Description)
		assert.NotEmpty(suite.T(), rec.BusinessImpact)
		assert.GreaterOrEqual(suite.T(), rec.Confidence, 0.0)
		assert.LessOrEqual(suite.T(), rec.Confidence, 1.0)

		// Verify implementation plan
		impl := rec.Implementation
		assert.NotEmpty(suite.T(), impl.Timeline)
		assert.NotEmpty(suite.T(), impl.Phases)
		assert.NotEmpty(suite.T(), impl.Resources)
		assert.NotEmpty(suite.T(), impl.KeyMilestones)
		assert.NotEmpty(suite.T(), impl.SuccessCriteria)

		// Verify ROI estimate
		roi := rec.ROIEstimate
		assert.GreaterOrEqual(suite.T(), roi.CostSavings, 0.0)
		assert.GreaterOrEqual(suite.T(), roi.RevenueIncrease, 0.0)
		assert.GreaterOrEqual(suite.T(), roi.EfficiencyGains, 0.0)
		assert.GreaterOrEqual(suite.T(), roi.RiskReduction, 0.0)
		assert.GreaterOrEqual(suite.T(), roi.ConfidenceLevel, 0.0)
		assert.LessOrEqual(suite.T(), roi.ConfidenceLevel, 1.0)
		assert.NotEmpty(suite.T(), roi.TimeToBreakeven)
	}
}

// TestDetermineAlertLevel tests alert level determination logic
func (suite *SAMServiceTestSuite) TestDetermineAlertLevel() {
	testCases := []struct {
		name          string
		healthScore   float64
		riskScore     float64
		turnoverRate  float64
		expectedAlert string
	}{
		{
			name:          "Green - Low risk, high health",
			healthScore:   0.9,
			riskScore:     0.2,
			turnoverRate:  0.05,
			expectedAlert: "GREEN",
		},
		{
			name:          "Yellow - Moderate risk",
			healthScore:   0.7,
			riskScore:     0.4,
			turnoverRate:  0.12,
			expectedAlert: "YELLOW",
		},
		{
			name:          "Orange - High risk",
			healthScore:   0.6,
			riskScore:     0.7,
			turnoverRate:  0.18,
			expectedAlert: "ORANGE",
		},
		{
			name:          "Red - Critical risk",
			healthScore:   0.4,
			riskScore:     0.9,
			turnoverRate:  0.25,
			expectedAlert: "RED",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			orgHealth := &OrganizationHealthMetrics{
				OverallHealthScore: tc.healthScore,
				TurnoverRate:       tc.turnoverRate,
			}

			riskAssessment := &RiskAssessmentResult{
				OverallRiskScore: tc.riskScore,
			}

			alertLevel := suite.service.determineAlertLevel(orgHealth, riskAssessment)
			assert.Equal(suite.T(), tc.expectedAlert, alertLevel)
		})
	}
}

// TestCreateImplementationPlan tests implementation plan creation
func (suite *SAMServiceTestSuite) TestCreateImplementationPlan() {
	planType := "测试计划"
	durationDays := 60

	// Test implementation plan creation
	plan := suite.service.createImplementationPlan(planType, durationDays)

	// Verify plan structure
	assert.Equal(suite.T(), "60天", plan.Timeline)
	assert.NotEmpty(suite.T(), plan.Phases)
	assert.NotEmpty(suite.T(), plan.Resources)
	assert.NotEmpty(suite.T(), plan.KeyMilestones)
	assert.NotEmpty(suite.T(), plan.SuccessCriteria)

	// Verify phases
	assert.Len(suite.T(), plan.Phases, 3) // Should have 3 phases
	for i, phase := range plan.Phases {
		assert.Equal(suite.T(), i+1, phase.PhaseNumber)
		assert.NotEmpty(suite.T(), phase.PhaseName)
		assert.NotEmpty(suite.T(), phase.Duration)
		assert.NotEmpty(suite.T(), phase.Activities)
		assert.NotEmpty(suite.T(), phase.Deliverables)

		if i > 0 {
			// Later phases should have dependencies
			assert.NotEmpty(suite.T(), phase.Dependencies)
		}
	}

	// Verify resources
	for _, resource := range plan.Resources {
		assert.NotEmpty(suite.T(), resource.ResourceType)
		assert.Greater(suite.T(), resource.Quantity, 0)
		assert.NotEmpty(suite.T(), resource.SkillRequirements)
		assert.NotEmpty(suite.T(), resource.TimeCommitment)
	}

	// Verify milestones
	for _, milestone := range plan.KeyMilestones {
		assert.NotEmpty(suite.T(), milestone.Name)
		assert.NotEmpty(suite.T(), milestone.Description)
		assert.NotZero(suite.T(), milestone.TargetDate)
		assert.NotEmpty(suite.T(), milestone.SuccessMetrics)
	}
}

// TestAnalyzeOrganizationHealthWithNoData tests health analysis with empty database
func (suite *SAMServiceTestSuite) TestAnalyzeOrganizationHealthWithNoData() {
	// Test with empty database
	orgHealth, err := suite.service.analyzeOrganizationHealth(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), orgHealth)

	// Should handle empty data gracefully
	assert.Equal(suite.T(), 0.0, orgHealth.OverallHealthScore) // No departments = 0 health
	assert.Empty(suite.T(), orgHealth.DepartmentHealthMap)
	assert.NotEmpty(suite.T(), orgHealth.TrendAnalysis.KeyDrivers)
}

// TestAnalyzeTalentMetricsWithNoEmployees tests talent analysis with no employees
func (suite *SAMServiceTestSuite) TestAnalyzeTalentMetricsWithNoEmployees() {
	// Test with no employees
	talentMetrics, err := suite.service.analyzeTalentMetrics(suite.ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), talentMetrics)

	// Should return default metrics even with no employees
	assert.GreaterOrEqual(suite.T(), talentMetrics.TalentPipelineHealth, 0.0)
	assert.NotEmpty(suite.T(), talentMetrics.SkillGapAnalysis)
	assert.NotEmpty(suite.T(), talentMetrics.PerformanceDistribution.PerformanceGaps)
}

// TestConcurrentContextGeneration tests concurrent access to SAM service
func (suite *SAMServiceTestSuite) TestConcurrentContextGeneration() {
	// Setup test data
	suite.createTestData()

	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	// Test concurrent context generation
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := suite.service.GenerateSituationalContext(suite.ctx)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(suite.T(), err)
	}
}

// TestPerformanceWithLargeDataset tests SAM service performance with large dataset
func (suite *SAMServiceTestSuite) TestPerformanceWithLargeDataset() {
	// Create a larger dataset
	suite.createLargeTestDataset()

	// Measure performance
	start := time.Now()
	context, err := suite.service.GenerateSituationalContext(suite.ctx)
	duration := time.Since(start)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), context)
	assert.Less(suite.T(), duration, 5*time.Second, "SAM context generation should complete within 5 seconds for large dataset")

	// Verify the analysis handles large dataset correctly
	assert.NotEmpty(suite.T(), context.OrganizationHealth.DepartmentHealthMap)
	assert.NotEmpty(suite.T(), context.Recommendations)
}

// Helper methods

// createTestData creates basic test data for SAM tests
func (suite *SAMServiceTestSuite) createTestData() {
	// Create employees
	employees := []struct {
		id         string
		name       string
		email      string
		department string
		position   string
		level      string
	}{
		{"EMP001", "张三", "zhang.san@company.com", "技术部", "软件工程师", "INTERMEDIATE"},
		{"EMP002", "李四", "li.si@company.com", "技术部", "高级工程师", "SENIOR"},
		{"EMP003", "王五", "wang.wu@company.com", "产品部", "产品经理", "INTERMEDIATE"},
		{"EMP004", "赵六", "zhao.liu@company.com", "人力资源部", "人事专员", "JUNIOR"},
		{"EMP005", "孙七", "sun.qi@company.com", "技术部", "技术总监", "MANAGER"},
	}

	hireDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	effectiveDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	for _, emp := range employees {
		// Create employee
		employee := suite.entClient.Employee.Create().
			SetEmployeeID(emp.id).
			SetLegalName(emp.name).
			SetEmail(emp.email).
			SetStatus("ACTIVE").
			SetHireDate(hireDate).
			SaveX(suite.ctx)

		// Create current position
		suite.entClient.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle(emp.position).
			SetDepartment(emp.department).
			SetJobLevel(emp.level).
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(effectiveDate).
			SetChangeReason("入职").
			SetIsRetroactive(false).
			SaveX(suite.ctx)
	}
}

// createLargeTestDataset creates a larger test dataset for performance testing
func (suite *SAMServiceTestSuite) createLargeTestDataset() {
	departments := []string{"技术部", "产品部", "人力资源部", "市场部", "销售部", "财务部", "运营部"}
	positions := []string{"工程师", "经理", "专员", "总监", "副总裁"}
	levels := []string{"JUNIOR", "INTERMEDIATE", "SENIOR", "MANAGER", "EXECUTIVE"}

	hireDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	effectiveDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Create 50 employees across different departments
	for i := 0; i < 50; i++ {
		empID := fmt.Sprintf("EMP%03d", i+1)
		name := fmt.Sprintf("员工%d", i+1)
		email := fmt.Sprintf("emp%d@company.com", i+1)
		dept := departments[i%len(departments)]
		position := positions[i%len(positions)]
		level := levels[i%len(levels)]

		// Create employee
		employee := suite.entClient.Employee.Create().
			SetEmployeeID(empID).
			SetLegalName(name).
			SetEmail(email).
			SetStatus("ACTIVE").
			SetHireDate(hireDate).
			SaveX(suite.ctx)

		// Create current position
		suite.entClient.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle(position).
			SetDepartment(dept).
			SetJobLevel(level).
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(effectiveDate).
			SetChangeReason("入职").
			SetIsRetroactive(false).
			SaveX(suite.ctx)
	}
}

// TestSAMServiceSuite runs the test suite
func TestSAMServiceSuite(t *testing.T) {
	suite.Run(t, new(SAMServiceTestSuite))
}

// Benchmark tests for SAM service performance
func BenchmarkGenerateSituationalContext(b *testing.B) {
	// Setup
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	entClient := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer entClient.Close()

	mockNeo4j := &MockNeo4jServiceForSAM{}
	service := NewSAMService(entClient, mockNeo4j, logger)
	ctx := context.Background()

	// Create some test data
	employee := entClient.Employee.Create().
		SetEmployeeID("BENCH001").
		SetLegalName("性能测试员工").
		SetEmail("bench@test.com").
		SetStatus("ACTIVE").
		SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SaveX(ctx)

	entClient.PositionHistory.Create().
		SetEmployeeID(employee.ID).
		SetPositionTitle("测试工程师").
		SetDepartment("技术部").
		SetJobLevel("INTERMEDIATE").
		SetEmploymentType("FULL_TIME").
		SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
		SetChangeReason("入职").
		SetIsRetroactive(false).
		SaveX(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateSituationalContext(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzeOrganizationHealth(b *testing.B) {
	// Setup
	logger := log.New(os.Stdout, "BENCH: ", log.LstdFlags)
	entClient := enttest.Open(b, "sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	defer entClient.Close()

	mockNeo4j := &MockNeo4jServiceForSAM{}
	service := NewSAMService(entClient, mockNeo4j, logger)
	ctx := context.Background()

	// Create benchmark data
	for i := 0; i < 20; i++ {
		employee := entClient.Employee.Create().
			SetEmployeeID(fmt.Sprintf("BENCH%03d", i+1)).
			SetLegalName(fmt.Sprintf("员工%d", i+1)).
			SetEmail(fmt.Sprintf("bench%d@test.com", i+1)).
			SetStatus("ACTIVE").
			SetHireDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SaveX(ctx)

		entClient.PositionHistory.Create().
			SetEmployeeID(employee.ID).
			SetPositionTitle("工程师").
			SetDepartment(fmt.Sprintf("部门%d", (i%5)+1)).
			SetJobLevel("INTERMEDIATE").
			SetEmploymentType("FULL_TIME").
			SetEffectiveDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)).
			SetChangeReason("入职").
			SetIsRetroactive(false).
			SaveX(ctx)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.analyzeOrganizationHealth(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
