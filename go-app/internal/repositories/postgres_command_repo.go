package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Employee PostgreSQL中的员工实体
type Employee struct {
	ID               uuid.UUID              `db:"id"`
	TenantID         uuid.UUID              `db:"tenant_id"`
	EmployeeType     string                 `db:"employee_type"`
	FirstName        string                 `db:"first_name"`
	LastName         string                 `db:"last_name"`
	Email            string                 `db:"email"`
	PositionID       *uuid.UUID             `db:"position_id"`
	HireDate         time.Time              `db:"hire_date"`
	TerminationDate  *time.Time             `db:"termination_date"`
	EmploymentStatus string                 `db:"employment_status"`
	PersonalInfo     map[string]interface{} `db:"personal_info"`
	CreatedAt        time.Time              `db:"created_at"`
	UpdatedAt        time.Time              `db:"updated_at"`
}

// OrganizationUnit PostgreSQL中的组织单元实体
type OrganizationUnit struct {
	ID           uuid.UUID              `db:"id"`
	TenantID     uuid.UUID              `db:"tenant_id"`
	UnitType     string                 `db:"unit_type"`
	Name         string                 `db:"name"`
	Description  *string                `db:"description"`
	ParentUnitID *uuid.UUID             `db:"parent_unit_id"`
	Profile      map[string]interface{} `db:"profile"`
	IsActive     bool                   `db:"is_active"`
	CreatedAt    time.Time              `db:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at"`
}

// Position PostgreSQL中的职位实体
type Position struct {
	ID           uuid.UUID `db:"id"`
	TenantID     uuid.UUID `db:"tenant_id"`
	Title        string    `db:"title"`
	Department   string    `db:"department"`
	Level        string    `db:"level"`
	Description  *string   `db:"description"`
	Requirements *string   `db:"requirements"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// PostgresCommandRepository PostgreSQL命令仓储接口
type PostgresCommandRepository interface {
	// 员工管理
	CreateEmployee(ctx context.Context, employee Employee) error
	UpdateEmployee(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	TerminateEmployee(ctx context.Context, id, tenantID uuid.UUID, terminationDate time.Time, reason string) error

	// 组织单元管理
	CreateOrganizationUnit(ctx context.Context, unit OrganizationUnit) error
	UpdateOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	DeleteOrganizationUnit(ctx context.Context, id, tenantID uuid.UUID) error

	// 职位管理
	CreatePosition(ctx context.Context, position Position) error
	UpdatePosition(ctx context.Context, id, tenantID uuid.UUID, changes map[string]interface{}) error
	AssignEmployeePosition(ctx context.Context, employeeID, positionID, tenantID uuid.UUID, startDate time.Time, isPrimary bool) error
}