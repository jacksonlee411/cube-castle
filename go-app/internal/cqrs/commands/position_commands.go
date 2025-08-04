package commands

import (
	"time"
	"github.com/google/uuid"
)

// CreatePositionCommand 创建职位命令
type CreatePositionCommand struct {
	TenantID          uuid.UUID              `json:"tenant_id" validate:"required"`
	PositionType      string                 `json:"position_type" validate:"required,oneof=FULL_TIME PART_TIME CONTINGENT_WORKER INTERN"`
	JobProfileID      uuid.UUID              `json:"job_profile_id" validate:"required"`
	DepartmentID      uuid.UUID              `json:"department_id" validate:"required"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            string                 `json:"status" validate:"oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	BudgetedFTE       float64                `json:"budgeted_fte" validate:"gte=0,lte=5"`
	Details           map[string]interface{} `json:"details,omitempty"`
}

// UpdatePositionCommand 更新职位命令
type UpdatePositionCommand struct {
	ID                uuid.UUID              `json:"id" validate:"required"`
	TenantID          uuid.UUID              `json:"tenant_id" validate:"required"`
	JobProfileID      *uuid.UUID             `json:"job_profile_id,omitempty"`
	DepartmentID      *uuid.UUID             `json:"department_id,omitempty"`
	ManagerPositionID *uuid.UUID             `json:"manager_position_id,omitempty"`
	Status            *string                `json:"status,omitempty" validate:"omitempty,oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	BudgetedFTE       *float64               `json:"budgeted_fte,omitempty" validate:"omitempty,gte=0,lte=5"`
	Details           map[string]interface{} `json:"details,omitempty"`
}

// AssignEmployeeToPositionCommand 员工职位分配命令  
type AssignEmployeeToPositionCommand struct {
	TenantID       uuid.UUID  `json:"tenant_id" validate:"required"`
	PositionID     uuid.UUID  `json:"position_id" validate:"required"`
	EmployeeID     uuid.UUID  `json:"employee_id" validate:"required"`
	StartDate      time.Time  `json:"start_date" validate:"required"`
	FTE            float64    `json:"fte" validate:"gte=0,lte=1"`
	AssignmentType string     `json:"assignment_type" validate:"oneof=PRIMARY SECONDARY ACTING TEMPORARY"`
	PayGradeID     *uuid.UUID `json:"pay_grade_id,omitempty"`
	Reason         string     `json:"reason" validate:"required"`
}

// RemoveEmployeeFromPositionCommand 员工职位移除命令
type RemoveEmployeeFromPositionCommand struct {
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	PositionID uuid.UUID `json:"position_id" validate:"required"`
	EmployeeID uuid.UUID `json:"employee_id" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
	Reason     string    `json:"reason" validate:"required"`
}

// DeletePositionCommand 删除职位命令
type DeletePositionCommand struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Reason   string    `json:"reason" validate:"required"`
}

// TransferPositionCommand 职位转移命令
type TransferPositionCommand struct {
	ID               uuid.UUID  `json:"id" validate:"required"`
	TenantID         uuid.UUID  `json:"tenant_id" validate:"required"`
	NewDepartmentID  uuid.UUID  `json:"new_department_id" validate:"required"`
	NewManagerID     *uuid.UUID `json:"new_manager_id,omitempty"`
	EffectiveDate    time.Time  `json:"effective_date" validate:"required"`
	TransferReason   string     `json:"transfer_reason" validate:"required"`
}

// UpdatePositionStatusCommand 更新职位状态命令
type UpdatePositionStatusCommand struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	TenantID   uuid.UUID `json:"tenant_id" validate:"required"`
	NewStatus  string    `json:"new_status" validate:"required,oneof=OPEN FILLED FROZEN PENDING_ELIMINATION"`
	ChangedBy  uuid.UUID `json:"changed_by" validate:"required"`
	Reason     string    `json:"reason" validate:"required"`
}