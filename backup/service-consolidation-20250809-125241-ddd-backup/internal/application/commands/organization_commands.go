package commands

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Command interface defines the contract for all commands
type Command interface {
	GetCommandID() uuid.UUID
	GetTenantID() uuid.UUID
	GetCommandType() string
	Validate() error
}

// CreateOrganizationCommand represents a command to create an organization
type CreateOrganizationCommand struct {
	CommandID     uuid.UUID `json:"command_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	RequestedCode *string   `json:"requested_code,omitempty"`
	Name          string    `json:"name" validate:"required,min=1,max=100"`
	ParentCode    *string   `json:"parent_code,omitempty"`
	UnitType      string    `json:"unit_type" validate:"required,oneof=COMPANY DEPARTMENT TEAM COST_CENTER PROJECT_TEAM"`
	Description   *string   `json:"description,omitempty" validate:"omitempty,max=500"`
	SortOrder     *int      `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	RequestedBy   uuid.UUID `json:"requested_by" validate:"required"`
}

func (c CreateOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c CreateOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c CreateOrganizationCommand) GetCommandType() string  { return "CreateOrganization" }

func (c CreateOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}

// UpdateOrganizationCommand represents a command to update an organization
type UpdateOrganizationCommand struct {
	CommandID   uuid.UUID `json:"command_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code" validate:"required,len=7"`
	Name        *string   `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Status      *string   `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
	Description *string   `json:"description,omitempty" validate:"omitempty,max=500"`
	SortOrder   *int      `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	RequestedBy uuid.UUID `json:"requested_by" validate:"required"`
}

func (c UpdateOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c UpdateOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c UpdateOrganizationCommand) GetCommandType() string  { return "UpdateOrganization" }

func (c UpdateOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}

// DeleteOrganizationCommand represents a command to delete an organization
type DeleteOrganizationCommand struct {
	CommandID   uuid.UUID `json:"command_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	Code        string    `json:"code" validate:"required,len=7"`
	RequestedBy uuid.UUID `json:"requested_by" validate:"required"`
}

func (c DeleteOrganizationCommand) GetCommandID() uuid.UUID { return c.CommandID }
func (c DeleteOrganizationCommand) GetTenantID() uuid.UUID  { return c.TenantID }
func (c DeleteOrganizationCommand) GetCommandType() string  { return "DeleteOrganization" }

func (c DeleteOrganizationCommand) Validate() error {
	validator := validator.New()
	return validator.Struct(c)
}