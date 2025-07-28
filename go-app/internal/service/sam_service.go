// internal/service/sam_service.go
package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
)

// SAMService provides Situational Awareness Model for AI-driven employee management
type SAMService struct {
	entClient    *ent.Client
	neo4jService *Neo4jService
	logger       *log.Logger
}

// SituationalContext represents the current organizational situation
type SituationalContext struct {
	Timestamp           time.Time                     `json:"timestamp"`
	OrganizationHealth  OrganizationHealthMetrics     `json:"organization_health"`
	TalentMetrics       TalentManagementMetrics       `json:"talent_metrics"`
	RiskAssessment      RiskAssessmentResult          `json:"risk_assessment"`
	OpportunityAnalysis OpportunityAnalysisResult     `json:"opportunity_analysis"`
	Recommendations     []StrategicRecommendation     `json:"recommendations"`
	AlertLevel          string                        `json:"alert_level"` // GREEN, YELLOW, ORANGE, RED
}

// OrganizationHealthMetrics represents organizational health indicators
type OrganizationHealthMetrics struct {
	OverallHealthScore    float64                        `json:"overall_health_score"`
	TurnoverRate          float64                        `json:"turnover_rate"`
	EmployeeEngagement    float64                        `json:"employee_engagement"`
	ProductivityIndex     float64                        `json:"productivity_index"`
	SpanOfControlHealth   float64                        `json:"span_of_control_health"`
	DepartmentHealthMap   map[string]DepartmentHealth    `json:"department_health_map"`
	TrendAnalysis         HealthTrendAnalysis            `json:"trend_analysis"`
}

// TalentManagementMetrics represents talent pipeline and development metrics
type TalentManagementMetrics struct {
	TalentPipelineHealth    float64                    `json:"talent_pipeline_health"`
	SuccessionReadiness     float64                    `json:"succession_readiness"`
	SkillGapAnalysis        map[string]float64         `json:"skill_gap_analysis"`
	PerformanceDistribution PerformanceDistribution    `json:"performance_distribution"`
	LearningDevelopmentROI  float64                    `json:"learning_development_roi"`
	InternalMobilityRate    float64                    `json:"internal_mobility_rate"`
}

// RiskAssessmentResult represents identified risks and their impact
type RiskAssessmentResult struct {
	OverallRiskScore      float64             `json:"overall_risk_score"`
	KeyPersonRisks        []KeyPersonRisk     `json:"key_person_risks"`
	ComplianceRisks       []ComplianceRisk    `json:"compliance_risks"`
	OperationalRisks      []OperationalRisk   `json:"operational_risks"`
	TalentFlightRisks     []TalentFlightRisk  `json:"talent_flight_risks"`
	RiskMitigation        []RiskMitigation    `json:"risk_mitigation"`
}

// OpportunityAnalysisResult represents growth and improvement opportunities
type OpportunityAnalysisResult struct {
	TalentOptimization    []TalentOptimization    `json:"talent_optimization"`
	ProcessImprovements   []ProcessImprovement    `json:"process_improvements"`
	StructuralChanges     []StructuralChange      `json:"structural_changes"`
	InvestmentPriorities  []InvestmentPriority    `json:"investment_priorities"`
	CapabilityGaps        []CapabilityGap         `json:"capability_gaps"`
}

// StrategicRecommendation represents AI-driven strategic recommendations
type StrategicRecommendation struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"` // IMMEDIATE, SHORT_TERM, LONG_TERM
	Priority        string                 `json:"priority"` // CRITICAL, HIGH, MEDIUM, LOW
	Category        string                 `json:"category"` // TALENT, STRUCTURE, PROCESS, RISK
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	BusinessImpact  string                 `json:"business_impact"`
	Implementation  ImplementationPlan     `json:"implementation"`
	ROIEstimate     ROIEstimate            `json:"roi_estimate"`
	RiskFactors     []string               `json:"risk_factors"`
	Success         Metrics                `json:"success_metrics"`
	Dependencies    []string               `json:"dependencies"`
	Confidence      float64                `json:"confidence"`
}

// Supporting data structures
type DepartmentHealth struct {
	HealthScore           float64   `json:"health_score"`
	TurnoverRate          float64   `json:"turnover_rate"`
	AverageTenure         float64   `json:"average_tenure"`
	ManagerEffectiveness  float64   `json:"manager_effectiveness"`
	TeamCohesion          float64   `json:"team_cohesion"`
	WorkloadBalance       float64   `json:"workload_balance"`
	LastAssessment        time.Time `json:"last_assessment"`
}

type HealthTrendAnalysis struct {
	Trend           string    `json:"trend"` // IMPROVING, STABLE, DECLINING
	TrendStrength   float64   `json:"trend_strength"`
	KeyDrivers      []string  `json:"key_drivers"`
	PredictedHealth float64   `json:"predicted_health"`
	Confidence      float64   `json:"confidence"`
}

type PerformanceDistribution struct {
	HighPerformers   float64 `json:"high_performers"`
	SolidPerformers  float64 `json:"solid_performers"`
	LowPerformers    float64 `json:"low_performers"`
	PerformanceGaps  []string `json:"performance_gaps"`
}

type KeyPersonRisk struct {
	EmployeeID       string    `json:"employee_id"`
	EmployeeName     string    `json:"employee_name"`
	Position         string    `json:"position"`
	Department       string    `json:"department"`
	RiskScore        float64   `json:"risk_score"`
	RiskFactors      []string  `json:"risk_factors"`
	BusinessImpact   string    `json:"business_impact"`
	MitigationSteps  []string  `json:"mitigation_steps"`
	LastAssessment   time.Time `json:"last_assessment"`
}

type ComplianceRisk struct {
	RiskType         string    `json:"risk_type"`
	Severity         string    `json:"severity"`
	Description      string    `json:"description"`
	AffectedAreas    []string  `json:"affected_areas"`
	ComplianceGaps   []string  `json:"compliance_gaps"`
	RemediationPlan  []string  `json:"remediation_plan"`
	Deadline         time.Time `json:"deadline"`
}

type OperationalRisk struct {
	RiskCategory     string   `json:"risk_category"`
	Description      string   `json:"description"`
	Probability      float64  `json:"probability"`
	Impact           float64  `json:"impact"`
	RiskScore        float64  `json:"risk_score"`
	AffectedTeams    []string `json:"affected_teams"`
	ContingencyPlan  []string `json:"contingency_plan"`
}

type TalentFlightRisk struct {
	EmployeeID       string    `json:"employee_id"`
	EmployeeName     string    `json:"employee_name"`
	FlightRisk       float64   `json:"flight_risk"`
	RiskIndicators   []string  `json:"risk_indicators"`
	RetentionActions []string  `json:"retention_actions"`
	TimeFrame        string    `json:"time_frame"`
}

type RiskMitigation struct {
	RiskType         string   `json:"risk_type"`
	MitigationAction string   `json:"mitigation_action"`
	Effectiveness    float64  `json:"effectiveness"`
	Timeline         string   `json:"timeline"`
	ResponsibleParty string   `json:"responsible_party"`
}

type TalentOptimization struct {
	OpportunityType  string   `json:"opportunity_type"`
	Description      string   `json:"description"`
	AffectedRoles    []string `json:"affected_roles"`
	ExpectedBenefit  string   `json:"expected_benefit"`
	ImplementationSteps []string `json:"implementation_steps"`
}

type ProcessImprovement struct {
	ProcessArea      string   `json:"process_area"`
	CurrentState     string   `json:"current_state"`
	ProposedState    string   `json:"proposed_state"`
	EfficiencyGain   float64  `json:"efficiency_gain"`
	ImplementationComplexity string `json:"implementation_complexity"`
}

type StructuralChange struct {
	ChangeType       string   `json:"change_type"`
	Description      string   `json:"description"`
	Rationale        string   `json:"rationale"`
	AffectedTeams    []string `json:"affected_teams"`
	ImplementationPhases []string `json:"implementation_phases"`
}

type InvestmentPriority struct {
	InvestmentArea   string   `json:"investment_area"`
	Priority         string   `json:"priority"`
	EstimatedCost    float64  `json:"estimated_cost"`
	ExpectedROI      float64  `json:"expected_roi"`
	Justification    string   `json:"justification"`
}

type CapabilityGap struct {
	CapabilityArea   string   `json:"capability_area"`
	CurrentLevel     float64  `json:"current_level"`
	RequiredLevel    float64  `json:"required_level"`
	GapSize          float64  `json:"gap_size"`
	ClosureStrategy  []string `json:"closure_strategy"`
}

type ImplementationPlan struct {
	Timeline         string              `json:"timeline"`
	Phases           []ImplementationPhase `json:"phases"`
	Resources        []ResourceRequirement `json:"resources"`
	KeyMilestones    []Milestone         `json:"key_milestones"`
	SuccessCriteria  []string            `json:"success_criteria"`
}

type ImplementationPhase struct {
	PhaseNumber      int      `json:"phase_number"`
	PhaseName        string   `json:"phase_name"`
	Duration         string   `json:"duration"`
	Activities       []string `json:"activities"`
	Dependencies     []string `json:"dependencies"`
	Deliverables     []string `json:"deliverables"`
}

type ResourceRequirement struct {
	ResourceType     string   `json:"resource_type"`
	Quantity         int      `json:"quantity"`
	SkillRequirements []string `json:"skill_requirements"`
	TimeCommitment   string   `json:"time_commitment"`
}

type Milestone struct {
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	TargetDate       time.Time `json:"target_date"`
	SuccessMetrics   []string  `json:"success_metrics"`
}

type ROIEstimate struct {
	CostSavings      float64 `json:"cost_savings"`
	RevenueIncrease  float64 `json:"revenue_increase"`
	EfficiencyGains  float64 `json:"efficiency_gains"`
	RiskReduction    float64 `json:"risk_reduction"`
	TimeToBreakeven  string  `json:"time_to_breakeven"`
	ConfidenceLevel  float64 `json:"confidence_level"`
}

type Metrics struct {
	KPIs             []KPI    `json:"kpis"`
	MeasurementPlan  string   `json:"measurement_plan"`
	ReportingCadence string   `json:"reporting_cadence"`
}

type KPI struct {
	Name             string  `json:"name"`
	Description      string  `json:"description"`
	CurrentValue     float64 `json:"current_value"`
	TargetValue      float64 `json:"target_value"`
	Measurement      string  `json:"measurement"`
}

// NewSAMService creates a new SAM service instance
func NewSAMService(
	entClient *ent.Client,
	neo4jService *Neo4jService,
	logger *log.Logger,
) *SAMService {
	return &SAMService{
		entClient:    entClient,
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// GenerateSituationalContext creates a comprehensive situational awareness context
func (s *SAMService) GenerateSituationalContext(ctx context.Context) (*SituationalContext, error) {
	s.logger.Println("Generating situational awareness context...")

	// Generate organization health metrics
	orgHealth, err := s.analyzeOrganizationHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze organization health: %w", err)
	}

	// Generate talent management metrics
	talentMetrics, err := s.analyzeTalentMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze talent metrics: %w", err)
	}

	// Perform risk assessment
	riskAssessment, err := s.performRiskAssessment(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to perform risk assessment: %w", err)
	}

	// Analyze opportunities
	opportunities, err := s.analyzeOpportunities(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze opportunities: %w", err)
	}

	// Generate strategic recommendations
	recommendations, err := s.generateRecommendations(ctx, orgHealth, talentMetrics, riskAssessment, opportunities)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Determine alert level
	alertLevel := s.determineAlertLevel(orgHealth, riskAssessment)

	situationalContext := &SituationalContext{
		Timestamp:           time.Now(),
		OrganizationHealth:  *orgHealth,
		TalentMetrics:       *talentMetrics,
		RiskAssessment:      *riskAssessment,
		OpportunityAnalysis: *opportunities,
		Recommendations:     recommendations,
		AlertLevel:          alertLevel,
	}

	s.logger.Printf("Generated situational context with alert level: %s", alertLevel)
	return situationalContext, nil
}

// analyzeOrganizationHealth performs comprehensive organization health analysis
func (s *SAMService) analyzeOrganizationHealth(ctx context.Context) (*OrganizationHealthMetrics, error) {
	// Get employee count by department
	departmentCounts := make(map[string]int)
	departments, err := s.entClient.PositionHistory.Query().
		Where(positionhistory.EndDateIsNil()).
		Select(positionhistory.FieldDepartment).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query departments: %w", err)
	}

	for _, pos := range departments {
		departmentCounts[pos.Department]++
	}

	// Calculate department health metrics
	departmentHealthMap := make(map[string]DepartmentHealth)
	for dept, count := range departmentCounts {
		// Simulate health calculations (in real implementation, this would use actual metrics)
		health := DepartmentHealth{
			HealthScore:          0.75 + (float64(count)/100)*0.1, // Simplified calculation
			TurnoverRate:         0.12 - (float64(count)/200)*0.02,
			AverageTenure:        24 + float64(count%12),
			ManagerEffectiveness: 0.8 + (float64(count%5))*0.04,
			TeamCohesion:         0.75 + (float64(count%10))*0.02,
			WorkloadBalance:      0.7 + (float64(count%8))*0.03,
			LastAssessment:       time.Now().Add(-time.Duration(count%30) * 24 * time.Hour),
		}
		departmentHealthMap[dept] = health
	}

	// Calculate overall health score
	totalHealth := 0.0
	for _, health := range departmentHealthMap {
		totalHealth += health.HealthScore
	}
	overallHealthScore := totalHealth / float64(len(departmentHealthMap))

	// Generate trend analysis
	trendAnalysis := HealthTrendAnalysis{
		Trend:           "STABLE",
		TrendStrength:   0.65,
		KeyDrivers:      []string{"Employee retention", "Management effectiveness", "Team collaboration"},
		PredictedHealth: overallHealthScore + 0.05,
		Confidence:      0.82,
	}

	if overallHealthScore < 0.6 {
		trendAnalysis.Trend = "DECLINING"
	} else if overallHealthScore > 0.85 {
		trendAnalysis.Trend = "IMPROVING"
	}

	return &OrganizationHealthMetrics{
		OverallHealthScore:  overallHealthScore,
		TurnoverRate:        0.145, // Simulated
		EmployeeEngagement:  0.735, // Simulated
		ProductivityIndex:   0.825, // Simulated
		SpanOfControlHealth: 0.685, // Simulated
		DepartmentHealthMap: departmentHealthMap,
		TrendAnalysis:       trendAnalysis,
	}, nil
}

// analyzeTalentMetrics performs talent management analysis
func (s *SAMService) analyzeTalentMetrics(ctx context.Context) (*TalentManagementMetrics, error) {
	// Get total employee count
	totalEmployees, err := s.entClient.Employee.Query().
		Where(employee.Status("ACTIVE")).
		Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count employees: %w", err)
	}

	// Simulate skill gap analysis (in real implementation, this would analyze actual skills data)
	skillGapAnalysis := map[string]float64{
		"Technical Leadership":     0.35,
		"Data Analytics":          0.28,
		"Digital Transformation": 0.42,
		"Project Management":      0.18,
		"Cloud Technologies":      0.55,
		"AI/ML Expertise":        0.68,
	}

	// Simulate performance distribution
	performanceDistribution := PerformanceDistribution{
		HighPerformers:  0.25,
		SolidPerformers: 0.65,
		LowPerformers:   0.10,
		PerformanceGaps: []string{"Leadership pipeline", "Technical expertise", "Cross-functional collaboration"},
	}

	return &TalentManagementMetrics{
		TalentPipelineHealth:    0.72,
		SuccessionReadiness:     0.58,
		SkillGapAnalysis:        skillGapAnalysis,
		PerformanceDistribution: performanceDistribution,
		LearningDevelopmentROI:  3.2,
		InternalMobilityRate:    0.18,
	}, nil
}

// performRiskAssessment identifies and evaluates organizational risks
func (s *SAMService) performRiskAssessment(ctx context.Context) (*RiskAssessmentResult, error) {
	// Identify key person risks
	keyPersonRisks := []KeyPersonRisk{
		{
			EmployeeID:     "emp-001",
			EmployeeName:   "张三",
			Position:       "技术总监",
			Department:     "技术部",
			RiskScore:      0.75,
			RiskFactors:    []string{"单点依赖", "知识垄断", "团队规模过大"},
			BusinessImpact: "技术决策延迟，团队效率下降",
			MitigationSteps: []string{"知识分享计划", "副手培养", "流程标准化"},
			LastAssessment: time.Now().Add(-7 * 24 * time.Hour),
		},
	}

	// Identify compliance risks
	complianceRisks := []ComplianceRisk{
		{
			RiskType:       "数据保护合规",
			Severity:       "MEDIUM",
			Description:    "员工个人信息处理流程需要加强",
			AffectedAreas:  []string{"人力资源部", "技术部"},
			ComplianceGaps: []string{"数据分类标准", "访问控制机制"},
			RemediationPlan: []string{"制定数据分类政策", "实施最小权限原则", "定期审计"},
			Deadline:       time.Now().Add(90 * 24 * time.Hour),
		},
	}

	// Identify operational risks
	operationalRisks := []OperationalRisk{
		{
			RiskCategory:    "人才流失",
			Description:     "关键岗位人员流失风险较高",
			Probability:     0.35,
			Impact:          0.8,
			RiskScore:       0.28,
			AffectedTeams:   []string{"技术部", "产品部"},
			ContingencyPlan: []string{"人才储备计划", "知识管理系统", "薪酬优化"},
		},
	}

	// Identify talent flight risks
	talentFlightRisks := []TalentFlightRisk{
		{
			EmployeeID:     "emp-002",
			EmployeeName:   "李四",
			FlightRisk:     0.65,
			RiskIndicators: []string{"市场薪酬差距", "职业发展瓶颈", "工作负荷过重"},
			RetentionActions: []string{"薪酬调整", "职业发展规划", "工作负荷优化"},
			TimeFrame:      "3-6个月",
		},
	}

	// Risk mitigation strategies
	riskMitigation := []RiskMitigation{
		{
			RiskType:         "关键人员依赖",
			MitigationAction: "实施知识管理和技能传承计划",
			Effectiveness:    0.8,
			Timeline:         "3个月",
			ResponsibleParty: "人力资源部",
		},
	}

	// Calculate overall risk score
	overallRiskScore := 0.0
	riskCount := 0
	for _, risk := range keyPersonRisks {
		overallRiskScore += risk.RiskScore
		riskCount++
	}
	for _, risk := range operationalRisks {
		overallRiskScore += risk.RiskScore
		riskCount++
	}
	if riskCount > 0 {
		overallRiskScore /= float64(riskCount)
	}

	return &RiskAssessmentResult{
		OverallRiskScore:  overallRiskScore,
		KeyPersonRisks:    keyPersonRisks,
		ComplianceRisks:   complianceRisks,
		OperationalRisks:  operationalRisks,
		TalentFlightRisks: talentFlightRisks,
		RiskMitigation:    riskMitigation,
	}, nil
}

// analyzeOpportunities identifies growth and improvement opportunities
func (s *SAMService) analyzeOpportunities(ctx context.Context) (*OpportunityAnalysisResult, error) {
	talentOptimization := []TalentOptimization{
		{
			OpportunityType: "内部晋升优化",
			Description:     "通过内部人才发展减少外部招聘成本",
			AffectedRoles:   []string{"高级工程师", "项目经理", "团队负责人"},
			ExpectedBenefit: "降低30%招聘成本，提升员工满意度",
			ImplementationSteps: []string{
				"建立内部人才发展通道",
				"制定技能发展计划",
				"实施导师制度",
			},
		},
	}

	processImprovements := []ProcessImprovement{
		{
			ProcessArea:              "绩效评估流程",
			CurrentState:             "年度评估，反馈滞后",
			ProposedState:            "季度评估，实时反馈",
			EfficiencyGain:           0.25,
			ImplementationComplexity: "MEDIUM",
		},
	}

	structuralChanges := []StructuralChange{
		{
			ChangeType:    "组织扁平化",
			Description:   "减少管理层级，提升决策效率",
			Rationale:     "当前管理链条过长，影响响应速度",
			AffectedTeams: []string{"技术部", "产品部"},
			ImplementationPhases: []string{
				"第一阶段：梳理现有层级",
				"第二阶段：重新设计组织架构",
				"第三阶段：实施变更和培训",
			},
		},
	}

	investmentPriorities := []InvestmentPriority{
		{
			InvestmentArea: "数字化技能培训",
			Priority:       "HIGH",
			EstimatedCost:  250000,
			ExpectedROI:    3.5,
			Justification:  "提升团队数字化能力，支撑业务转型",
		},
	}

	capabilityGaps := []CapabilityGap{
		{
			CapabilityArea:  "数据分析能力",
			CurrentLevel:    0.6,
			RequiredLevel:   0.8,
			GapSize:         0.2,
			ClosureStrategy: []string{"外部培训", "内部分享", "实战项目"},
		},
	}

	return &OpportunityAnalysisResult{
		TalentOptimization:   talentOptimization,
		ProcessImprovements:  processImprovements,
		StructuralChanges:    structuralChanges,
		InvestmentPriorities: investmentPriorities,
		CapabilityGaps:       capabilityGaps,
	}, nil
}

// generateRecommendations creates AI-driven strategic recommendations
func (s *SAMService) generateRecommendations(
	ctx context.Context,
	orgHealth *OrganizationHealthMetrics,
	talentMetrics *TalentManagementMetrics,
	riskAssessment *RiskAssessmentResult,
	opportunities *OpportunityAnalysisResult,
) ([]StrategicRecommendation, error) {
	recommendations := []StrategicRecommendation{}

	// Critical risk mitigation recommendations
	if riskAssessment.OverallRiskScore > 0.7 {
		rec := StrategicRecommendation{
			ID:             "rec-001",
			Type:           "IMMEDIATE",
			Priority:       "CRITICAL",
			Category:       "RISK",
			Title:          "关键风险缓解计划",
			Description:    "立即实施关键人员风险缓解措施，避免业务中断",
			BusinessImpact: "防止关键业务能力丧失，保障业务连续性",
			Implementation: s.createImplementationPlan("风险缓解", 30),
			ROIEstimate: ROIEstimate{
				CostSavings:     500000,
				RiskReduction:   0.6,
				TimeToBreakeven: "1个月",
				ConfidenceLevel: 0.9,
			},
			RiskFactors: []string{"实施阻力", "资源冲突", "时间紧迫"},
			Confidence:  0.85,
		}
		recommendations = append(recommendations, rec)
	}

	// Talent development recommendations
	if talentMetrics.SuccessionReadiness < 0.6 {
		rec := StrategicRecommendation{
			ID:             "rec-002",
			Type:           "SHORT_TERM",
			Priority:       "HIGH",
			Category:       "TALENT",
			Title:          "人才梯队建设计划",
			Description:    "建立完善的人才梯队，提升组织韧性",
			BusinessImpact: "确保关键岗位有合适的继任者，降低人员变动风险",
			Implementation: s.createImplementationPlan("人才发展", 90),
			ROIEstimate: ROIEstimate{
				CostSavings:     200000,
				EfficiencyGains: 0.2,
				TimeToBreakeven: "6个月",
				ConfidenceLevel: 0.8,
			},
			Confidence: 0.82,
		}
		recommendations = append(recommendations, rec)
	}

	// Organization health improvement
	if orgHealth.OverallHealthScore < 0.7 {
		rec := StrategicRecommendation{
			ID:             "rec-003",
			Type:           "LONG_TERM",
			Priority:       "MEDIUM",
			Category:       "STRUCTURE",
			Title:          "组织健康提升计划",
			Description:    "通过流程优化和文化建设提升组织整体健康度",
			BusinessImpact: "提升员工满意度和工作效率，减少流失率",
			Implementation: s.createImplementationPlan("组织优化", 180),
			ROIEstimate: ROIEstimate{
				CostSavings:     300000,
				EfficiencyGains: 0.15,
				TimeToBreakeven: "12个月",
				ConfidenceLevel: 0.75,
			},
			Confidence: 0.78,
		}
		recommendations = append(recommendations, rec)
	}

	// Sort recommendations by priority and confidence
	sort.Slice(recommendations, func(i, j int) bool {
		priorityScore := map[string]int{"CRITICAL": 4, "HIGH": 3, "MEDIUM": 2, "LOW": 1}
		iScore := priorityScore[recommendations[i].Priority]
		jScore := priorityScore[recommendations[j].Priority]
		
		if iScore != jScore {
			return iScore > jScore
		}
		return recommendations[i].Confidence > recommendations[j].Confidence
	})

	return recommendations, nil
}

// determineAlertLevel calculates the overall alert level based on metrics
func (s *SAMService) determineAlertLevel(
	orgHealth *OrganizationHealthMetrics,
	riskAssessment *RiskAssessmentResult,
) string {
	// Calculate composite risk score
	healthRisk := 1.0 - orgHealth.OverallHealthScore
	riskScore := riskAssessment.OverallRiskScore
	turnoverRisk := math.Min(orgHealth.TurnoverRate/0.3, 1.0) // Normalize to 0-1

	compositeRisk := (healthRisk*0.4 + riskScore*0.4 + turnoverRisk*0.2)

	switch {
	case compositeRisk >= 0.8:
		return "RED"
	case compositeRisk >= 0.6:
		return "ORANGE"
	case compositeRisk >= 0.4:
		return "YELLOW"
	default:
		return "GREEN"
	}
}

// createImplementationPlan creates a structured implementation plan
func (s *SAMService) createImplementationPlan(planType string, durationDays int) ImplementationPlan {
	phases := []ImplementationPhase{
		{
			PhaseNumber:  1,
			PhaseName:    "评估和规划",
			Duration:     "2周",
			Activities:   []string{"现状分析", "需求评估", "方案设计"},
			Dependencies: []string{},
			Deliverables: []string{"评估报告", "实施方案"},
		},
		{
			PhaseNumber:  2,
			PhaseName:    "试点实施",
			Duration:     "4周",
			Activities:   []string{"选择试点团队", "实施方案", "收集反馈"},
			Dependencies: []string{"阶段1完成"},
			Deliverables: []string{"试点结果", "改进建议"},
		},
		{
			PhaseNumber:  3,
			PhaseName:    "全面推广",
			Duration:     "8周",
			Activities:   []string{"全面部署", "培训支持", "监控执行"},
			Dependencies: []string{"阶段2完成"},
			Deliverables: []string{"实施报告", "成效评估"},
		},
	}

	resources := []ResourceRequirement{
		{
			ResourceType:      "项目经理",
			Quantity:          1,
			SkillRequirements: []string{"项目管理", "变更管理"},
			TimeCommitment:    "100%",
		},
		{
			ResourceType:      "业务分析师",
			Quantity:          2,
			SkillRequirements: []string{"业务分析", "流程梳理"},
			TimeCommitment:    "50%",
		},
	}

	milestones := []Milestone{
		{
			Name:           "方案确认",
			Description:    "实施方案获得管理层批准",
			TargetDate:     time.Now().Add(14 * 24 * time.Hour),
			SuccessMetrics: []string{"方案批准", "资源分配确认"},
		},
		{
			Name:           "试点完成",
			Description:    "试点实施完成并收集反馈",
			TargetDate:     time.Now().Add(42 * 24 * time.Hour),
			SuccessMetrics: []string{"试点覆盖率100%", "反馈收集完成"},
		},
	}

	return ImplementationPlan{
		Timeline:        fmt.Sprintf("%d天", durationDays),
		Phases:          phases,
		Resources:       resources,
		KeyMilestones:   milestones,
		SuccessCriteria: []string{"按时完成", "达成预期目标", "获得用户认可"},
	}
}