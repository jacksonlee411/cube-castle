// internal/graphql/resolvers/organization_resolver.go
package resolvers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// OrganizationResolver handles GraphQL queries for organizational structure
type OrganizationResolver struct {
	neo4jService *service.Neo4jService
}

// NewOrganizationResolver creates a new organization resolver
func NewOrganizationResolver(neo4jService *service.Neo4jService) *OrganizationResolver {
	return &OrganizationResolver{
		neo4jService: neo4jService,
	}
}

// OrganizationChart represents the organizational structure
type OrganizationChart struct {
	Department      string                 `json:"department"`
	Employees       []*EmployeeInChart     `json:"employees"`
	SubDepartments  []*SubDepartment       `json:"sub_departments"`
	ManagerCount    int                    `json:"manager_count"`
	TotalEmployees  int                    `json:"total_employees"`
}

// SubDepartment represents a sub-department in the organization chart
type SubDepartment struct {
	Department     string              `json:"department"`
	Employees      []*EmployeeInChart  `json:"employees"`
	ManagerCount   int                 `json:"manager_count"`
	TotalEmployees int                 `json:"total_employees"`
}

// EmployeeInChart represents an employee in the organization chart
type EmployeeInChart struct {
	ID              string          `json:"id"`
	LegalName       string          `json:"legal_name"`
	CurrentPosition *PositionInChart `json:"current_position"`
}

// PositionInChart represents a position in the organization chart
type PositionInChart struct {
	PositionTitle string `json:"position_title"`
	JobLevel      string `json:"job_level"`
}

// ReportingPath represents the path between two employees
type ReportingPath struct {
	FromEmployee *EmployeeInChart `json:"from_employee"`
	ToEmployee   *EmployeeInChart `json:"to_employee"`
	Path         []*PathStep      `json:"path"`
	Distance     int              `json:"distance"`
	PathType     string           `json:"path_type"`
}

// PathStep represents one step in a reporting path
type PathStep struct {
	Employee     *EmployeeInChart `json:"employee"`
	Relationship string           `json:"relationship"`
}

// CommonManager represents the common manager result
type CommonManager struct {
	ID              string          `json:"id"`
	LegalName       string          `json:"legal_name"`
	CurrentPosition *PositionInChart `json:"current_position"`
}

// GetOrganizationChart returns the organizational chart for a department
func (r *OrganizationResolver) GetOrganizationChart(ctx context.Context, args struct {
	RootDepartment *string
	AsOfDate       *string
	MaxLevels      *int
}) (*OrganizationChart, error) {
	rootDept := "Technology"
	if args.RootDepartment != nil {
		rootDept = *args.RootDepartment
	}

	maxLevels := 5
	if args.MaxLevels != nil {
		maxLevels = *args.MaxLevels
	}

	// Get department structure from Neo4j
	deptStructure, err := r.neo4jService.GetDepartmentStructure(ctx, rootDept)
	if err != nil {
		return nil, fmt.Errorf("failed to get department structure: %w", err)
	}

	// For now, return a simplified structure
	// In a real implementation, you would build the complete hierarchy
	chart := &OrganizationChart{
		Department: deptStructure.Name,
		Employees: []*EmployeeInChart{
			{
				ID:        "emp-001",
				LegalName: "张三",
				CurrentPosition: &PositionInChart{
					PositionTitle: "技术总监",
					JobLevel:      "DIRECTOR",
				},
			},
			{
				ID:        "emp-002", 
				LegalName: "李四",
				CurrentPosition: &PositionInChart{
					PositionTitle: "高级软件工程师",
					JobLevel:      "SENIOR",
				},
			},
		},
		SubDepartments: []*SubDepartment{
			{
				Department: "前端开发部",
				Employees: []*EmployeeInChart{
					{
						ID:        "emp-003",
						LegalName: "王五",
						CurrentPosition: &PositionInChart{
							PositionTitle: "前端工程师",
							JobLevel:      "INTERMEDIATE",
						},
					},
				},
				ManagerCount:   1,
				TotalEmployees: 8,
			},
			{
				Department: "后端开发部",
				Employees: []*EmployeeInChart{
					{
						ID:        "emp-004",
						LegalName: "赵六",
						CurrentPosition: &PositionInChart{
							PositionTitle: "后端工程师",
							JobLevel:      "SENIOR",
						},
					},
				},
				ManagerCount:   1,
				TotalEmployees: 12,
			},
		},
		ManagerCount:   2,
		TotalEmployees: 22,
	}

	return chart, nil
}

// FindReportingPath finds the reporting path between two employees
func (r *OrganizationResolver) FindReportingPath(ctx context.Context, args struct {
	FromEmployeeID string
	ToEmployeeID   string
}) (*ReportingPath, error) {
	// Get reporting path from Neo4j
	orgPath, err := r.neo4jService.FindReportingPath(ctx, args.FromEmployeeID, args.ToEmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find reporting path: %w", err)
	}

	// Convert to GraphQL response format
	path := &ReportingPath{
		Distance: orgPath.Distance,
		PathType: orgPath.PathType,
		Path:     make([]*PathStep, 0, len(orgPath.Path)),
	}

	for _, segment := range orgPath.Path {
		step := &PathStep{
			Employee: &EmployeeInChart{
				ID:        segment.Employee.ID,
				LegalName: segment.Employee.LegalName,
			},
			Relationship: segment.Relationship,
		}
		path.Path = append(path.Path, step)
	}

	if len(path.Path) > 0 {
		path.FromEmployee = path.Path[0].Employee
		path.ToEmployee = path.Path[len(path.Path)-1].Employee
	}

	return path, nil
}

// FindCommonManager finds the common manager for a list of employees
func (r *OrganizationResolver) FindCommonManager(ctx context.Context, args struct {
	EmployeeIDs []string
}) (*CommonManager, error) {
	if len(args.EmployeeIDs) < 2 {
		return nil, fmt.Errorf("at least 2 employee IDs are required")
	}

	// Get common manager from Neo4j
	manager, err := r.neo4jService.FindCommonManager(ctx, args.EmployeeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find common manager: %w", err)
	}

	commonManager := &CommonManager{
		ID:        manager.ID,
		LegalName: manager.LegalName,
		CurrentPosition: &PositionInChart{
			PositionTitle: "部门经理", // This would come from the actual position data
			JobLevel:      "MANAGER",
		},
	}

	return commonManager, nil
}

// GetReportingHierarchy returns the complete reporting hierarchy for a manager
func (r *OrganizationResolver) GetReportingHierarchy(ctx context.Context, args struct {
	ManagerID string
	MaxDepth  *int
}) (*ReportingHierarchyResult, error) {
	maxDepth := 5
	if args.MaxDepth != nil {
		maxDepth = *args.MaxDepth
	}

	// Get reporting hierarchy from Neo4j
	hierarchy, err := r.neo4jService.GetReportingHierarchy(ctx, args.ManagerID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to get reporting hierarchy: %w", err)
	}

	// Convert to GraphQL response format
	result := &ReportingHierarchyResult{
		Manager: &EmployeeInChart{
			ID:        hierarchy.Manager.ID,
			LegalName: hierarchy.Manager.LegalName,
		},
		DirectReports: make([]*EmployeeInChart, 0, len(hierarchy.DirectReports)),
		AllReports:    make([]*EmployeeInChart, 0, len(hierarchy.AllReports)),
		Depth:         hierarchy.Depth,
	}

	for _, report := range hierarchy.DirectReports {
		result.DirectReports = append(result.DirectReports, &EmployeeInChart{
			ID:        report.ID,
			LegalName: report.LegalName,
		})
	}

	for _, report := range hierarchy.AllReports {
		result.AllReports = append(result.AllReports, &EmployeeInChart{
			ID:        report.ID,
			LegalName: report.LegalName,
		})
	}

	return result, nil
}

// ReportingHierarchyResult represents the reporting hierarchy response
type ReportingHierarchyResult struct {
	Manager       *EmployeeInChart   `json:"manager"`
	DirectReports []*EmployeeInChart `json:"direct_reports"`
	AllReports    []*EmployeeInChart `json:"all_reports"`
	Depth         int                `json:"depth"`
}

// GetOrganizationMetrics returns organizational metrics and insights
func (r *OrganizationResolver) GetOrganizationMetrics(ctx context.Context, args struct {
	Department *string
	AsOfDate   *string
}) (*OrganizationMetrics, error) {
	// This would query Neo4j for various organizational metrics
	// For now, return mock data
	metrics := &OrganizationMetrics{
		TotalEmployees:     156,
		TotalDepartments:   12,
		AverageTeamSize:    8.5,
		MaxReportingDepth:  5,
		SpanOfControl:      6.2,
		DepartmentMetrics: []*DepartmentMetric{
			{
				Department:     "技术部",
				EmployeeCount:  45,
				ManagerCount:   8,
				AverageSpan:    5.6,
				MaxDepth:       4,
			},
			{
				Department:     "产品部",
				EmployeeCount:  28,
				ManagerCount:   5,
				AverageSpan:    5.6,
				MaxDepth:       3,
			},
			{
				Department:     "销售部",
				EmployeeCount:  32,
				ManagerCount:   6,
				AverageSpan:    5.3,
				MaxDepth:       3,
			},
		},
	}

	return metrics, nil
}

// OrganizationMetrics represents organizational metrics
type OrganizationMetrics struct {
	TotalEmployees     int                 `json:"total_employees"`
	TotalDepartments   int                 `json:"total_departments"`
	AverageTeamSize    float64             `json:"average_team_size"`
	MaxReportingDepth  int                 `json:"max_reporting_depth"`
	SpanOfControl      float64             `json:"span_of_control"`
	DepartmentMetrics  []*DepartmentMetric `json:"department_metrics"`
}

// DepartmentMetric represents metrics for a specific department
type DepartmentMetric struct {
	Department    string  `json:"department"`
	EmployeeCount int     `json:"employee_count"`
	ManagerCount  int     `json:"manager_count"`
	AverageSpan   float64 `json:"average_span"`
	MaxDepth      int     `json:"max_depth"`
}

// SyncEmployeeToGraph syncs employee data to the graph database
func (r *OrganizationResolver) SyncEmployeeToGraph(ctx context.Context, employeeID string) error {
	// This would typically be called when employee data changes
	// to keep the graph database in sync with the relational database
	
	// Convert from relational data to graph node
	employee := service.EmployeeNode{
		ID:         employeeID,
		EmployeeID: employeeID,
		LegalName:  "示例员工", // Would come from actual employee data
		Email:      "example@company.com",
		Status:     "ACTIVE",
	}

	return r.neo4jService.SyncEmployee(ctx, employee)
}

// SyncPositionToGraph syncs position data to the graph database
func (r *OrganizationResolver) SyncPositionToGraph(ctx context.Context, positionID, employeeID string) error {
	// Convert from relational data to graph node
	position := service.PositionNode{
		ID:            positionID,
		PositionTitle: "示例职位", // Would come from actual position data
		Department:    "技术部",
		JobLevel:      "SENIOR",
		Location:      "北京",
	}

	return r.neo4jService.SyncPosition(ctx, position, employeeID)
}

// CreateReportingRelationship creates a reporting relationship in the graph
func (r *OrganizationResolver) CreateReportingRelationship(ctx context.Context, managerID, reporteeID string) error {
	return r.neo4jService.CreateReportingRelationship(ctx, managerID, reporteeID)
}