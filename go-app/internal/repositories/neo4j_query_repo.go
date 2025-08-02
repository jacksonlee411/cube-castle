package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
)

// EmployeeNode Neo4j中的员工节点
type EmployeeNode struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	FirstName        string                 `json:"first_name"`
	LastName         string                 `json:"last_name"`
	Email            string                 `json:"email"`
	EmployeeType     string                 `json:"employee_type"`
	EmploymentStatus string                 `json:"employment_status"`
	HireDate         time.Time              `json:"hire_date"`
	TerminationDate  *time.Time             `json:"termination_date,omitempty"`
	PersonalInfo     map[string]interface{} `json:"personal_info,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// OrganizationUnitNode Neo4j中的组织单元节点
type OrganizationUnitNode struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	UnitType    string                 `json:"unit_type"`
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Profile     map[string]interface{} `json:"profile,omitempty"`
	IsActive    bool                   `json:"is_active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PositionNode Neo4j中的职位节点
type PositionNode struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Title        string    `json:"title"`
	Department   string    `json:"department"`
	Level        string    `json:"level"`
	Description  *string   `json:"description,omitempty"`
	Requirements *string   `json:"requirements,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OrgChartResponse 组织架构图响应
type OrgChartResponse struct {
	RootUnit OrganizationUnitNode   `json:"root_unit"`
	Children []OrgChartNode         `json:"children"`
	Metadata map[string]interface{} `json:"metadata"`
}

// OrgChartNode 组织架构图节点
type OrgChartNode struct {
	Unit      OrganizationUnitNode `json:"unit"`
	Employees []EmployeeNode       `json:"employees"`
	Children  []OrgChartNode       `json:"children"`
	Level     int                  `json:"level"`
}

// ReportingHierarchyResponse 汇报层级响应
type ReportingHierarchyResponse struct {
	Manager     EmployeeNode                 `json:"manager"`
	DirectReports []ReportingHierarchyNode   `json:"direct_reports"`
	TotalCount    int                        `json:"total_count"`
	MaxDepth      int                        `json:"max_depth"`
}

// ReportingHierarchyNode 汇报层级节点
type ReportingHierarchyNode struct {
	Employee      EmployeeNode               `json:"employee"`
	DirectReports []ReportingHierarchyNode   `json:"direct_reports"`
	Level         int                        `json:"level"`
}

// EmployeeSearchResponse 员工搜索响应
type EmployeeSearchResponse struct {
	Employees  []EmployeeNode `json:"employees"`
	TotalCount int            `json:"total_count"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

// OrganizationUnitsResponse 组织单元列表响应
type OrganizationUnitsResponse struct {
	Units      []OrganizationUnitNode `json:"units"`
	TotalCount int                    `json:"total_count"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
}

// Neo4jQueryRepository Neo4j查询仓储接口
type Neo4jQueryRepository interface {
	// 员工查询
	GetEmployee(ctx context.Context, query queries.FindEmployeeQuery) (*EmployeeNode, error)
	SearchEmployees(ctx context.Context, query queries.SearchEmployeesQuery) (*EmployeeSearchResponse, error)

	// 组织架构查询
	GetOrgChart(ctx context.Context, query queries.GetOrgChartQuery) (*OrgChartResponse, error)
	GetOrganizationUnit(ctx context.Context, query queries.GetOrganizationUnitQuery) (*OrganizationUnitNode, error)
	ListOrganizationUnits(ctx context.Context, query queries.ListOrganizationUnitsQuery) (*OrganizationUnitsResponse, error)

	// 层级关系查询
	GetReportingHierarchy(ctx context.Context, query queries.GetReportingHierarchyQuery) (*ReportingHierarchyResponse, error)
	FindEmployeePath(ctx context.Context, query queries.FindEmployeePathQuery) ([]EmployeeNode, error)
	GetDepartmentStructure(ctx context.Context, query queries.GetDepartmentStructureQuery) (*OrgChartResponse, error)
	FindCommonManager(ctx context.Context, query queries.FindCommonManagerQuery) (*EmployeeNode, error)
}