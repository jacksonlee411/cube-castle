package queries

import (
	"github.com/google/uuid"
)

// GetOrganizationQuery 获取单个组织查询
type GetOrganizationQuery struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	ID       uuid.UUID `json:"id" validate:"required"`
}

// ListOrganizationsQuery 组织列表查询
type ListOrganizationsQuery struct {
	TenantID     uuid.UUID  `json:"tenant_id" validate:"required"`
	ParentUnitID *uuid.UUID `json:"parent_unit_id,omitempty"`
	UnitType     *string    `json:"unit_type,omitempty" validate:"omitempty,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Search       *string    `json:"search,omitempty"`
	Page         int        `json:"page" validate:"min=1"`
	PageSize     int        `json:"page_size" validate:"min=1,max=1000"`
}

// GetOrganizationTreeQuery 组织树查询
type GetOrganizationTreeQuery struct {
	TenantID        uuid.UUID  `json:"tenant_id" validate:"required"`
	RootUnitID      *uuid.UUID `json:"root_unit_id,omitempty"`
	MaxDepth        int        `json:"max_depth" validate:"min=1,max=10"`
	IncludeInactive bool       `json:"include_inactive"`
	ExpandAll       bool       `json:"expand_all"`
}

// GetOrganizationStatsQuery 组织统计查询
type GetOrganizationStatsQuery struct {
	TenantID    uuid.UUID  `json:"tenant_id" validate:"required"`
	UnitType    *string    `json:"unit_type,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	DateRange   *DateRange `json:"date_range,omitempty"`
	Granularity string     `json:"granularity" validate:"oneof=daily weekly monthly yearly"`
}

// SearchOrganizationsQuery 组织搜索查询
type SearchOrganizationsQuery struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	Query        string    `json:"query" validate:"required,min=1"`
	UnitTypes    []string  `json:"unit_types,omitempty"`
	Status       []string  `json:"status,omitempty"`
	Limit        int       `json:"limit" validate:"min=1,max=100"`
	Offset       int       `json:"offset" validate:"min=0"`
	SortBy       string    `json:"sort_by" validate:"oneof=name created_at updated_at level"`
	SortOrder    string    `json:"sort_order" validate:"oneof=asc desc"`
}

// GetOrgChartQuery 获取组织架构图查询 (向后兼容)
type GetOrgChartQuery struct {
	TenantID        uuid.UUID  `json:"tenant_id" validate:"required"`
	RootUnitID      *uuid.UUID `json:"root_unit_id,omitempty"`
	MaxDepth        int        `json:"max_depth" validate:"min=1,max=10"`
	IncludeInactive bool       `json:"include_inactive"`
}

// DateRange 日期范围
type DateRange struct {
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

// OrganizationResponse 组织响应结构
type OrganizationResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	UnitType     string                 `json:"unit_type"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id"`
	Status       string                 `json:"status"`
	Profile      map[string]interface{} `json:"profile"`
	Level        int                    `json:"level"`
	EmployeeCount int                   `json:"employee_count"`
	Children     []OrganizationResponse `json:"children,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	UpdatedAt    string                 `json:"updated_at"`
}

// OrganizationListResponse 组织列表响应
type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	Pagination    PaginationInfo         `json:"pagination"`
	Summary       *OrganizationSummary   `json:"summary,omitempty"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// OrganizationSummary 组织摘要
type OrganizationSummary struct {
	TotalCount        int                    `json:"total_count"`
	ActiveCount       int                    `json:"active_count"`
	InactiveCount     int                    `json:"inactive_count"`
	TypeDistribution  map[string]int         `json:"type_distribution"`
	LevelDistribution map[int]int            `json:"level_distribution"`
}

// OrganizationStats 组织统计
type OrganizationStats struct {
	Total          int                 `json:"total"`
	Active         int                 `json:"active"`
	Inactive       int                 `json:"inactive"`
	TotalEmployees int                 `json:"total_employees"`
	ByType         map[string]int      `json:"by_type"`
	ByLevel        map[int]int         `json:"by_level"`
	LastUpdated    string              `json:"last_updated"`
}

// FindEmployeeQuery 查找员工查询
type FindEmployeeQuery struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	ID       uuid.UUID `json:"id" validate:"required"`
}

// SearchEmployeesQuery 搜索员工查询
type SearchEmployeesQuery struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	Name       *string   `json:"name,omitempty"`
	Email      *string   `json:"email,omitempty"`
	Department *string   `json:"department,omitempty"`
	Limit      int       `json:"limit" validate:"min=1,max=1000"`
	Offset     int       `json:"offset" validate:"min=0"`
}

// GetReportingHierarchyQuery 获取汇报层级查询
type GetReportingHierarchyQuery struct {
	TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
	ManagerID uuid.UUID `json:"manager_id" validate:"required"`
	MaxDepth  int       `json:"max_depth" validate:"min=1,max=10"`
}

// GetOrganizationUnitQuery 获取组织单元查询
type GetOrganizationUnitQuery struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	ID       uuid.UUID `json:"id" validate:"required"`
}

// ListOrganizationUnitsQuery 列出组织单元查询
type ListOrganizationUnitsQuery struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	UnitType   *string   `json:"unit_type,omitempty"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Limit      int       `json:"limit" validate:"min=1,max=1000"`
	Offset     int       `json:"offset" validate:"min=0"`
}

// FindEmployeePathQuery 查找员工路径查询
type FindEmployeePathQuery struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	FromID     uuid.UUID `json:"from_id" validate:"required"`
	ToID       uuid.UUID `json:"to_id" validate:"required"`
}

// GetDepartmentStructureQuery 获取部门结构查询
type GetDepartmentStructureQuery struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	DepartmentID uuid.UUID `json:"department_id" validate:"required"`
	MaxDepth     int       `json:"max_depth" validate:"min=1,max=10"`
}

// FindCommonManagerQuery 查找共同管理者查询
type FindCommonManagerQuery struct {
	TenantID    uuid.UUID   `json:"tenant_id" validate:"required"`
	EmployeeIDs []uuid.UUID `json:"employee_ids" validate:"required,min=2"`
}