package commands

import (
	"time"
	"github.com/google/uuid"
)

// CreateOrganizationCommand 创建组织命令
type CreateOrganizationCommand struct {
	TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
	UnitType     string                 `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
	Name         string                 `json:"name" validate:"required,min=1,max=100"`
	Description  *string                `json:"description,omitempty" validate:"omitempty,max=500"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Status       string                 `json:"status" validate:"required,oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// UpdateOrganizationCommand 更新组织命令
type UpdateOrganizationCommand struct {
	ID           uuid.UUID              `json:"id" validate:"required"`
	TenantID     uuid.UUID              `json:"tenant_id" validate:"required"`
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description  *string                `json:"description,omitempty" validate:"omitempty,max=500"`
	ParentUnitID *uuid.UUID             `json:"parent_unit_id,omitempty"`
	Status       *string                `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Profile      map[string]interface{} `json:"profile,omitempty"`
}

// DeleteOrganizationCommand 删除组织命令
type DeleteOrganizationCommand struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Force    bool      `json:"force,omitempty"` // 强制删除，即使有子组织
}

// MoveOrganizationCommand 移动组织命令
type MoveOrganizationCommand struct {
	ID              uuid.UUID  `json:"id" validate:"required"`
	TenantID        uuid.UUID  `json:"tenant_id" validate:"required"`
	NewParentUnitID *uuid.UUID `json:"new_parent_unit_id,omitempty"`
	MoveDate        time.Time  `json:"move_date" validate:"required"`
	Reason          *string    `json:"reason,omitempty"`
}

// ActivateOrganizationCommand 激活组织命令
type ActivateOrganizationCommand struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// DeactivateOrganizationCommand 停用组织命令
type DeactivateOrganizationCommand struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Reason   *string   `json:"reason,omitempty"`
}

// BulkUpdateOrganizationsCommand 批量更新组织命令
type BulkUpdateOrganizationsCommand struct {
	TenantID        uuid.UUID                      `json:"tenant_id" validate:"required"`
	OrganizationIDs []uuid.UUID                    `json:"organization_ids" validate:"required,min=1,max=100"`
	Updates         map[string]interface{}         `json:"updates" validate:"required"`
	Metadata        map[string]interface{}         `json:"metadata,omitempty"`
}