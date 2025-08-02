package queries

import (
	"github.com/google/uuid"
)

// GetOrgChartQuery 获取组织架构图查询
type GetOrgChartQuery struct {
	TenantID        uuid.UUID  `json:"tenant_id" validate:"required"`
	RootUnitID      *uuid.UUID `json:"root_unit_id,omitempty"`
	MaxDepth        int        `json:"max_depth" validate:"min=1,max=10"`
	IncludeInactive bool       `json:"include_inactive"`
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