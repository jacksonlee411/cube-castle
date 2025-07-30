// test/fixtures/mocks.go
package fixtures

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
)

// MockTemporalQueryService provides comprehensive mock for TemporalQueryService
type MockTemporalQueryService struct {
	mock.Mock
}

func (m *MockTemporalQueryService) GetPositionAsOfDate(ctx context.Context, tenantID, employeeID uuid.UUID, asOfDate time.Time) (*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, asOfDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.PositionSnapshot), args.Error(1)
}

func (m *MockTemporalQueryService) GetPositionTimeline(ctx context.Context, tenantID, employeeID uuid.UUID, fromDate, toDate *time.Time) ([]*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, fromDate, toDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*service.PositionSnapshot), args.Error(1)
}

func (m *MockTemporalQueryService) ValidateTemporalConsistency(ctx context.Context, tenantID, employeeID uuid.UUID, effectiveDate time.Time) error {
	args := m.Called(ctx, tenantID, employeeID, effectiveDate)
	return args.Error(0)
}

func (m *MockTemporalQueryService) CreatePositionSnapshot(ctx context.Context, tenantID, employeeID uuid.UUID, snapshotDate time.Time) (*service.PositionSnapshot, error) {
	args := m.Called(ctx, tenantID, employeeID, snapshotDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.PositionSnapshot), args.Error(1)
}

// MockNeo4jService provides comprehensive mock for Neo4jService
type MockNeo4jService struct {
	mock.Mock
}

func (m *MockNeo4jService) SyncEmployee(ctx context.Context, employee service.EmployeeNode) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockNeo4jService) SyncPosition(ctx context.Context, position service.PositionNode, employeeID string) error {
	args := m.Called(ctx, position, employeeID)
	return args.Error(0)
}

func (m *MockNeo4jService) CreateReportingRelationship(ctx context.Context, managerID, reporteeID string) error {
	args := m.Called(ctx, managerID, reporteeID)
	return args.Error(0)
}

func (m *MockNeo4jService) FindReportingPath(ctx context.Context, fromEmployeeID, toEmployeeID string) (*service.OrganizationalPath, error) {
	args := m.Called(ctx, fromEmployeeID, toEmployeeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OrganizationalPath), args.Error(1)
}

func (m *MockNeo4jService) GetReportingHierarchy(ctx context.Context, managerID string, maxDepth int) (*service.ReportingHierarchy, error) {
	args := m.Called(ctx, managerID, maxDepth)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ReportingHierarchy), args.Error(1)
}

func (m *MockNeo4jService) FindCommonManager(ctx context.Context, employeeIDs []string) (*service.EmployeeNode, error) {
	args := m.Called(ctx, employeeIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.EmployeeNode), args.Error(1)
}

func (m *MockNeo4jService) GetDepartmentStructure(ctx context.Context, rootDepartment string) (*service.DepartmentNode, error) {
	args := m.Called(ctx, rootDepartment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.DepartmentNode), args.Error(1)
}

func (m *MockNeo4jService) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockSAMService provides comprehensive mock for SAMService
type MockSAMService struct {
	mock.Mock
}

func (m *MockSAMService) GenerateSituationalContext(ctx context.Context) (*service.SituationalContext, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SituationalContext), args.Error(1)
}

// MockWorkflowClient provides comprehensive mock for Temporal workflow client
type MockWorkflowClient struct {
	mock.Mock
}

func (m *MockWorkflowClient) StartWorkflow(ctx context.Context, workflowID string, request interface{}) error {
	args := m.Called(ctx, workflowID, request)
	return args.Error(0)
}

func (m *MockWorkflowClient) GetWorkflowStatus(ctx context.Context, workflowID string) (*workflow.WorkflowStatus, error) {
	args := m.Called(ctx, workflowID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*workflow.WorkflowStatus), args.Error(1)
}

func (m *MockWorkflowClient) SignalWorkflow(ctx context.Context, workflowID, signalName string, arg interface{}) error {
	args := m.Called(ctx, workflowID, signalName, arg)
	return args.Error(0)
}

func (m *MockWorkflowClient) CancelWorkflow(ctx context.Context, workflowID string) error {
	args := m.Called(ctx, workflowID)
	return args.Error(0)
}

// Mock data builders for common test scenarios

// NewMockPositionSnapshot creates a mock position snapshot with realistic data
func NewMockPositionSnapshot(employeeID uuid.UUID, title, department, level string) *service.PositionSnapshot {
	return &service.PositionSnapshot{
		PositionHistoryID: uuid.New(),
		EmployeeID:        employeeID,
		PositionTitle:     title,
		Department:        department,
		JobLevel:          level,
		Location:          &[]string{"北京"}[0],
		EmploymentType:    "FULL_TIME",
		EffectiveDate:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ChangeReason:      &[]string{"入职"}[0],
		IsRetroactive:     false,
		MinSalary:         &[]float64{8000}[0],
		MaxSalary:         &[]float64{12000}[0],
		Currency:          &[]string{"CNY"}[0],
	}
}

// NewMockSituationalContext creates a mock situational context with realistic data
func NewMockSituationalContext() *service.SituationalContext {
	return &service.SituationalContext{
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
					TeamCohesion:         0.75,
					WorkloadBalance:      0.70,
					LastAssessment:       time.Now().Add(-7 * 24 * time.Hour),
				},
				"产品部": {
					HealthScore:          0.72,
					TurnoverRate:         0.15,
					AverageTenure:        18.2,
					ManagerEffectiveness: 0.78,
					TeamCohesion:         0.68,
					WorkloadBalance:      0.75,
					LastAssessment:       time.Now().Add(-5 * 24 * time.Hour),
				},
			},
			TrendAnalysis: service.HealthTrendAnalysis{
				Trend:           "IMPROVING",
				TrendStrength:   0.65,
				KeyDrivers:      []string{"员工满意度", "管理效能", "团队协作"},
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
				"项目管理":  0.22,
			},
			PerformanceDistribution: service.PerformanceDistribution{
				HighPerformers:  0.25,
				SolidPerformers: 0.65,
				LowPerformers:   0.10,
				PerformanceGaps: []string{"跨团队协作", "技术创新", "客户导向"},
			},
			LearningDevelopmentROI: 3.2,
			InternalMobilityRate:   0.18,
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
					LastAssessment: time.Now().Add(-7 * 24 * time.Hour),
				},
			},
			ComplianceRisks: []service.ComplianceRisk{
				{
					RiskType:       "数据保护合规",
					Severity:       "MEDIUM",
					Description:    "员工个人信息处理流程需要加强",
					AffectedAreas:  []string{"人力资源部", "技术部"},
					ComplianceGaps: []string{"数据分类标准", "访问控制机制"},
					RemediationPlan: []string{"制定数据分类政策", "实施最小权限原则"},
					Deadline:       time.Now().Add(90 * 24 * time.Hour),
				},
			},
			OperationalRisks: []service.OperationalRisk{
				{
					RiskCategory:    "人才流失",
					Description:     "关键岗位人员流失风险较高",
					Probability:     0.35,
					Impact:          0.8,
					RiskScore:       0.28,
					AffectedTeams:   []string{"技术部", "产品部"},
					ContingencyPlan: []string{"人才储备计划", "知识管理系统"},
				},
			},
			TalentFlightRisks: []service.TalentFlightRisk{
				{
					EmployeeID:     "emp-002",
					EmployeeName:   "李四",
					FlightRisk:     0.65,
					RiskIndicators: []string{"市场薪酬差距", "职业发展瓶颈"},
					RetentionActions: []string{"薪酬调整", "职业发展规划"},
					TimeFrame:      "3-6个月",
				},
			},
			RiskMitigation: []service.RiskMitigation{
				{
					RiskType:         "关键人员依赖",
					MitigationAction: "实施知识管理和技能传承计划",
					Effectiveness:    0.8,
					Timeline:         "3个月",
					ResponsibleParty: "人力资源部",
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
					ImplementationSteps: []string{"建立内部人才发展通道", "制定技能发展计划"},
				},
			},
			ProcessImprovements: []service.ProcessImprovement{
				{
					ProcessArea:              "绩效评估流程",
					CurrentState:             "年度评估，反馈滞后",
					ProposedState:            "季度评估，实时反馈",
					EfficiencyGain:           0.25,
					ImplementationComplexity: "MEDIUM",
				},
			},
			StructuralChanges: []service.StructuralChange{
				{
					ChangeType:    "组织扁平化",
					Description:   "减少管理层级，提升决策效率",
					Rationale:     "当前管理链条过长，影响响应速度",
					AffectedTeams: []string{"技术部", "产品部"},
					ImplementationPhases: []string{"梳理现有层级", "重新设计组织架构", "实施变更和培训"},
				},
			},
			InvestmentPriorities: []service.InvestmentPriority{
				{
					InvestmentArea: "数字化技能培训",
					Priority:       "HIGH",
					EstimatedCost:  250000,
					ExpectedROI:    3.5,
					Justification:  "提升团队数字化能力，支撑业务转型",
				},
			},
			CapabilityGaps: []service.CapabilityGap{
				{
					CapabilityArea:  "数据分析能力",
					CurrentLevel:    0.6,
					RequiredLevel:   0.8,
					GapSize:         0.2,
					ClosureStrategy: []string{"外部培训", "内部分享", "实战项目"},
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
				Implementation: service.ImplementationPlan{
					Timeline: "3个月",
					Phases: []service.ImplementationPhase{
						{
							PhaseNumber:  1,
							PhaseName:    "现状评估",
							Duration:     "4周",
							Activities:   []string{"人才盘点", "能力评估"},
							Dependencies: []string{},
							Deliverables: []string{"人才现状报告"},
						},
					},
					Resources: []service.ResourceRequirement{
						{
							ResourceType:      "HR专员",
							Quantity:          2,
							SkillRequirements: []string{"人才评估", "培训管理"},
							TimeCommitment:    "50%",
						},
					},
					KeyMilestones: []service.Milestone{
						{
							Name:           "评估完成",
							Description:    "完成现有人才评估",
							TargetDate:     time.Now().Add(30 * 24 * time.Hour),
							SuccessMetrics: []string{"100%员工评估完成", "人才地图生成"},
						},
					},
					SuccessCriteria: []string{"建立人才梯队", "提升继任准备度"},
				},
				ROIEstimate: service.ROIEstimate{
					CostSavings:     200000,
					RevenueIncrease: 150000,
					EfficiencyGains: 0.2,
					RiskReduction:   0.3,
					TimeToBreakeven: "6个月",
					ConfidenceLevel: 0.8,
				},
				RiskFactors:  []string{"管理层支持", "员工参与度"},
				Dependencies: []string{"HR系统升级", "培训预算批准"},
				Confidence:   0.82,
			},
		},
	}
}

// NewMockWorkflowStatus creates a mock workflow status
func NewMockWorkflowStatus(workflowID, status string) *workflow.WorkflowStatus {
	return &workflow.WorkflowStatus{
		WorkflowID: workflowID,
		Status:     status,
		StartTime:  time.Now().Add(-1 * time.Hour),
		UpdateTime: time.Now(),
		Result:     map[string]interface{}{"success": true},
	}
}

// Helper functions for setting up mock expectations

// SetupStandardTemporalMocks sets up common mock expectations for TemporalQueryService
func SetupStandardTemporalMocks(mock *MockTemporalQueryService, employeeID uuid.UUID) {
	tenantID := uuid.New()
	
	// Mock GetPositionAsOfDate
	snapshot := NewMockPositionSnapshot(employeeID, "软件工程师", "技术部", "INTERMEDIATE")
	mock.On("GetPositionAsOfDate", 
		mock.Anything, tenantID, employeeID, mock.AnythingOfType("time.Time")).
		Return(snapshot, nil)
	
	// Mock GetPositionTimeline
	timeline := []*service.PositionSnapshot{snapshot}
	mock.On("GetPositionTimeline", 
		mock.Anything, tenantID, employeeID, mock.Anything, mock.Anything).
		Return(timeline, nil)
	
	// Mock ValidateTemporalConsistency
	mock.On("ValidateTemporalConsistency", 
		mock.Anything, tenantID, employeeID, mock.AnythingOfType("time.Time")).
		Return(nil)
}

// SetupStandardSAMMocks sets up common mock expectations for SAMService
func SetupStandardSAMMocks(mock *MockSAMService) {
	context := NewMockSituationalContext()
	mock.On("GenerateSituationalContext", mock.Anything).Return(context, nil)
}

// SetupStandardWorkflowMocks sets up common mock expectations for WorkflowClient
func SetupStandardWorkflowMocks(mock *MockWorkflowClient) {
	// Mock StartWorkflow
	mock.On("StartWorkflow", 
		mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(nil)
	
	// Mock GetWorkflowStatus
	status := NewMockWorkflowStatus("test-workflow-id", "RUNNING")
	mock.On("GetWorkflowStatus", 
		mock.Anything, mock.AnythingOfType("string")).
		Return(status, nil)
}