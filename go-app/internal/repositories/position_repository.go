package repositories

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
)

// Employee 员工实体（简化版）  
type Employee struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	EmploymentStatus string    `json:"employment_status"`
	EmployeeType     string    `json:"employee_type"`
}

// PositionCommandRepository 职位命令仓储接口
type PositionCommandRepository interface {
	CreatePosition(ctx context.Context, position Position) error
	UpdatePosition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, updates map[string]interface{}) error
	DeletePosition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
	CreatePositionOccupancyHistory(ctx context.Context, history PositionOccupancyHistory) error
	EndPositionOccupancy(ctx context.Context, positionID, employeeID uuid.UUID, endDate time.Time, reason string) error
	ValidateEmployeePositionAssignment(ctx context.Context, employeeID, positionID, tenantID uuid.UUID) (bool, error)
	TransferPosition(ctx context.Context, id uuid.UUID, newDeptID uuid.UUID, newManagerID *uuid.UUID, effectiveDate time.Time, reason string) error
	UpdatePositionStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, newStatus string, changedBy uuid.UUID, reason string) error
	
	// Outbox Pattern支持的原子操作
	AssignEmployeeWithEvent(ctx context.Context, assignment PositionAssignment, event OutboxEvent) error
	RemoveEmployeeWithEvent(ctx context.Context, positionID, employeeID uuid.UUID, endDate time.Time, reason string, event OutboxEvent) error
}

// PositionQueryRepository Neo4j职位查询仓储接口
type PositionQueryRepository interface {
	// 基础查询
	GetPosition(ctx context.Context, query queries.GetPositionQuery) (*queries.PositionResponse, error)
	GetPositionWithRelations(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) (*PositionWithRelations, error)
	
	// 搜索和列表
	SearchPositions(ctx context.Context, params SearchPositionsParams) ([]Position, int, error)
	
	// 层级和关系查询
	GetPositionHierarchy(ctx context.Context, query queries.GetPositionHierarchyQuery) ([]queries.PositionNode, error)
	GetEmployeePositions(ctx context.Context, query queries.GetEmployeePositionsQuery) ([]queries.OccupancyHistoryResponse, error)
	GetPositionEmployees(ctx context.Context, query queries.GetPositionEmployeesQuery) ([]queries.EmployeeResponse, error)
	
	// 历史和统计
	GetPositionOccupancyHistory(ctx context.Context, params OccupancyHistoryParams) ([]PositionOccupancyHistory, int, error)
	GetPositionStats(ctx context.Context, query queries.GetPositionStatsQuery) (*queries.PositionStatsResponse, error)
	GetEmployeePositionHistory(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID, limit, offset int) ([]PositionOccupancyHistory, int, error)
}

// Position 职位实体
type Position struct {
	ID                uuid.UUID              `json:"id"`
	TenantID          uuid.UUID              `json:"tenant_id"`
	PositionType      string                 `json:"position_type"`
	JobProfileID      uuid.UUID              `json:"job_profile_id"`
	DepartmentID      uuid.UUID              `json:"department_id"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            string                 `json:"status"`
	BudgetedFTE       float64                `json:"budgeted_fte"`
	Details           map[string]interface{} `json:"details,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// PositionAssignment 简化的职位分配实体（替代复杂的PositionOccupancyHistory）
type PositionAssignment struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	PositionID     uuid.UUID  `json:"position_id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	FTE            float64    `json:"fte"`
	AssignmentType string     `json:"assignment_type"` // PRIMARY, SECONDARY, ACTING
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PositionOccupancyHistory 职位占用历史实体
type PositionOccupancyHistory struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id"`
	PositionID     uuid.UUID  `json:"position_id"`
	EmployeeID     uuid.UUID  `json:"employee_id"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	IsCurrent      bool       `json:"is_current"`
	FTE            float64    `json:"fte"`
	AssignmentType string     `json:"assignment_type"`
	PayGradeID     *uuid.UUID `json:"pay_grade_id,omitempty"`
	Reason         string     `json:"reason,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PositionWithRelations 带关系的职位信息
type PositionWithRelations struct {
	Position        Position                   `json:"position"`
	Department      *Organization              `json:"department,omitempty"`
	Manager         *Position                  `json:"manager,omitempty"`
	DirectReports   []Position                 `json:"direct_reports,omitempty"`
	CurrentEmployee *Employee                  `json:"current_employee,omitempty"`
	History         []PositionOccupancyHistory `json:"history,omitempty"`
}

// SearchPositionsParams 职位搜索参数
type SearchPositionsParams struct {
	TenantID     uuid.UUID  `json:"tenant_id"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	Status       *string    `json:"status,omitempty"`
	PositionType *string    `json:"position_type,omitempty"`
	ManagerID    *uuid.UUID `json:"manager_id,omitempty"`
	JobProfileID *uuid.UUID `json:"job_profile_id,omitempty"`
	Search       *string    `json:"search,omitempty"`
	Limit        int        `json:"limit"`
	Offset       int        `json:"offset"`
}

// OccupancyHistoryParams 占用历史查询参数
type OccupancyHistoryParams struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	PositionID  *uuid.UUID `json:"position_id,omitempty"`
	EmployeeID  *uuid.UUID `json:"employee_id,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	IsCurrent   *bool      `json:"is_current,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}

// PositionHierarchy 职位层级结构
type PositionHierarchy struct {
	Nodes map[string]*PositionHierarchyNode `json:"nodes"`
	Root  *PositionHierarchyNode            `json:"root,omitempty"`
}

// PositionHierarchyNode 职位层级节点
type PositionHierarchyNode struct {
	Position  Position                 `json:"position"`
	Children  []*PositionHierarchyNode `json:"children,omitempty"`
	Parent    *PositionHierarchyNode   `json:"parent,omitempty"`
	Level     int                      `json:"level"`
	Employee  *Employee                `json:"employee,omitempty"`
}

// PositionStats 职位统计
type PositionStats struct {
	Total              int     `json:"total"`
	Open               int     `json:"open"`
	Filled             int     `json:"filled"`
	Frozen             int     `json:"frozen"`
	PendingElimination int     `json:"pending_elimination"`
	AverageFTE         float64 `json:"average_fte"`
	VacancyRate        float64 `json:"vacancy_rate"`
	TurnoverRate       float64 `json:"turnover_rate"`
}