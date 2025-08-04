package queries

import (
	"time"
	"github.com/google/uuid"
)

// GetPositionQuery 获取单个职位查询
type GetPositionQuery struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	ID       uuid.UUID `json:"id" validate:"required"`
}

// SearchPositionsQuery 职位搜索查询
type SearchPositionsQuery struct {
	TenantID     uuid.UUID  `json:"tenant_id" validate:"required"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	PositionType *string    `json:"position_type,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	JobProfileID *uuid.UUID `json:"job_profile_id,omitempty"`
	Search       *string    `json:"search,omitempty"`
	Limit        int        `json:"limit" validate:"min=1,max=1000"`
	Offset       int        `json:"offset" validate:"min=0"`
}

// GetPositionHierarchyQuery 职位层级查询
type GetPositionHierarchyQuery struct {
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required"`
	RootPositionID *uuid.UUID `json:"root_position_id,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id,omitempty"`
	MaxDepth       int        `json:"max_depth" validate:"min=1,max=10"`
}

// GetEmployeePositionsQuery 获取员工职位历史查询
type GetEmployeePositionsQuery struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	EmployeeID   uuid.UUID `json:"employee_id" validate:"required"`
	IncludePast  bool      `json:"include_past"`
}

// GetPositionEmployeesQuery 获取职位员工查询
type GetPositionEmployeesQuery struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	PositionID   uuid.UUID `json:"position_id" validate:"required"`
	OnlyCurrent  bool      `json:"only_current"`
}

// GetPositionOccupancyHistoryQuery 职位占用历史查询
type GetPositionOccupancyHistoryQuery struct {
	TenantID    uuid.UUID  `json:"tenant_id" validate:"required"`
	PositionID  *uuid.UUID `json:"position_id,omitempty"`
	EmployeeID  *uuid.UUID `json:"employee_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	IsCurrent   *bool      `json:"is_current,omitempty"`
	Limit       int        `json:"limit" validate:"min=1,max=1000"`
	Offset      int        `json:"offset" validate:"min=0"`
}

// GetPositionStatsQuery 职位统计查询
type GetPositionStatsQuery struct {
	TenantID     uuid.UUID  `json:"tenant_id" validate:"required"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	DateRange    *DateRange `json:"date_range,omitempty"`
}

// PositionStatsResponse 职位统计响应
type PositionStatsResponse struct {
	Total              int     `json:"total"`
	Open               int     `json:"open"`
	Filled             int     `json:"filled"`
	Frozen             int     `json:"frozen"`
	PendingElimination int     `json:"pending_elimination"`
	AverageFTE         float64 `json:"average_fte"`
	VacancyRate        float64 `json:"vacancy_rate"`
	TurnoverRate       float64 `json:"turnover_rate"`
}

// PositionResponse 职位响应结构
type PositionResponse struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	PositionType      string                 `json:"position_type"`
	JobProfileID      uuid.UUID              `json:"job_profile_id"`
	DepartmentID      uuid.UUID              `json:"department_id"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            string                 `json:"status"`
	BudgetedFTE       float64                `json:"budgeted_fte"`
	Details           map[string]interface{} `json:"details,omitempty"`
	
	// 关联信息
	Department        *OrganizationResponse  `json:"department,omitempty"`
	Manager           *PositionResponse      `json:"manager,omitempty"`
	DirectReports     []PositionResponse     `json:"direct_reports,omitempty"`
	CurrentEmployee   *EmployeeResponse      `json:"current_employee,omitempty"`
	OccupancyHistory  []OccupancyHistoryResponse `json:"occupancy_history,omitempty"`
	
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// OccupancyHistoryResponse 占用历史响应
type OccupancyHistoryResponse struct {
	ID             uuid.UUID  `json:"id"`
	PositionID     uuid.UUID  `json:"position_id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	FTE            float64    `json:"fte"`
	AssignmentType string     `json:"assignment_type"`
	Reason         string     `json:"reason,omitempty"`
	
	// 关联信息
	Employee       *EmployeeResponse      `json:"employee,omitempty"`
	Position       *PositionResponse      `json:"position,omitempty"`
	
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// PositionHierarchyResponse 职位层级响应
type PositionHierarchyResponse struct {
	Nodes map[string]*PositionNode `json:"nodes"`
	Root  *PositionNode            `json:"root,omitempty"`
}

// PositionNode 职位节点
type PositionNode struct {
	Position  PositionResponse  `json:"position"`
	Children  []*PositionNode   `json:"children,omitempty"`
	Parent    *PositionNode     `json:"parent,omitempty"`
	Level     int               `json:"level"`
	Employee  *EmployeeResponse `json:"employee,omitempty"`
}

// EmployeeResponse 员工响应结构 (简化版，用于关联)
type EmployeeResponse struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Status       string    `json:"status"`
	EmployeeType string    `json:"employee_type"`
}

// PositionSearchResponse 职位搜索响应
type PositionSearchResponse struct {
	Positions  []PositionResponse `json:"positions"`
	Total      int                `json:"total"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	Pagination PaginationInfo     `json:"pagination"`
}