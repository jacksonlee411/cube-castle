package commands

import (
	"time"
	"github.com/google/uuid"
)

// HireEmployeeCommand 雇佣员工命令
type HireEmployeeCommand struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	FirstName    string    `json:"first_name" validate:"required,min=1,max=100"`
	LastName     string    `json:"last_name" validate:"required,min=1,max=100"`
	Email        string    `json:"email" validate:"required,email"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	HireDate     time.Time `json:"hire_date" validate:"required"`
	EmployeeType string    `json:"employee_type" validate:"required,oneof=FULL_TIME PART_TIME CONTRACTOR INTERN"`
}

// UpdateEmployeeCommand 更新员工命令
type UpdateEmployeeCommand struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	TenantID  uuid.UUID `json:"tenant_id" validate:"required"`
	FirstName *string   `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string   `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Email     *string   `json:"email,omitempty" validate:"omitempty,email"`
}

// TerminateEmployeeCommand 终止员工命令
type TerminateEmployeeCommand struct {
	ID           uuid.UUID `json:"id" validate:"required"`
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	TerminationDate time.Time `json:"termination_date" validate:"required"`
	Reason       string    `json:"reason" validate:"required"`
}

// CreateOrganizationUnitCommand 创建组织单元命令
type CreateOrganizationUnitCommand struct {
	TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
	UnitType     string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
	Name         string                 `json:"name" validate:"required,min=1,max=100"`
	Description  *string                `json:"description,omitempty"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// UpdateOrganizationUnitCommand 更新组织单元命令
type UpdateOrganizationUnitCommand struct {
	ID          uuid.UUID              `json:"id" validate:"required"`
	TenantID    uuid.UUID              `json:"tenant_id" validate:"required"`
	Name        *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string                `json:"description,omitempty"`
	Profile     map[string]interface{} `json:"profile,omitempty"`
}

// CreatePositionCommand 创建职位命令
type CreatePositionCommand struct {
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	Title        string    `json:"title" validate:"required,min=1,max=100"`
	Department   string    `json:"department" validate:"required"`
	Level        string    `json:"level" validate:"required"`
	Description  *string   `json:"description,omitempty"`
	Requirements *string   `json:"requirements,omitempty"`
}

// AssignEmployeePositionCommand 分配员工职位命令
type AssignEmployeePositionCommand struct {
	EmployeeID   uuid.UUID `json:"employee_id" validate:"required"`
	PositionID   uuid.UUID `json:"position_id" validate:"required"`
	TenantID     uuid.UUID `json:"tenant_id" validate:"required"`
	StartDate    time.Time `json:"start_date" validate:"required"`
	IsPrimary    bool      `json:"is_primary"`
}