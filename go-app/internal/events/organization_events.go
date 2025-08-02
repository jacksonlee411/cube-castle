package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OrganizationEvents 组织相关领域事件

// OrganizationCreated 组织创建事件
type OrganizationCreated struct {
	*BaseDomainEvent
	Name        string     `json:"name"`
	Code        string     `json:"code"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Level       int        `json:"level"`
	Status      string     `json:"status"`
	Description string     `json:"description,omitempty"`
}

// NewOrganizationCreated 创建组织创建事件
func NewOrganizationCreated(tenantID, organizationID uuid.UUID, name, code string, parentID *uuid.UUID, level int) *OrganizationCreated {
	base := NewBaseDomainEvent("organization.created", "organization", organizationID, tenantID)
	return &OrganizationCreated{
		BaseDomainEvent: base,
		Name:            name,
		Code:            code,
		ParentID:        parentID,
		Level:           level,
		Status:          "active",
	}
}

func (e *OrganizationCreated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// OrganizationUpdated 组织更新事件
type OrganizationUpdated struct {
	*BaseDomainEvent
	OrganizationCode string                 `json:"organization_code"`
	UpdatedFields    map[string]interface{} `json:"updated_fields"`
	PreviousValues   map[string]interface{} `json:"previous_values,omitempty"`
}

// NewOrganizationUpdated 创建组织更新事件
func NewOrganizationUpdated(tenantID, organizationID uuid.UUID, code string, updatedFields map[string]interface{}) *OrganizationUpdated {
	base := NewBaseDomainEvent("organization.updated", "organization", organizationID, tenantID)
	return &OrganizationUpdated{
		BaseDomainEvent: base,
		OrganizationCode: code,
		UpdatedFields:    updatedFields,
		PreviousValues:   make(map[string]interface{}),
	}
}

func (e *OrganizationUpdated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// OrganizationDeleted 组织删除事件
type OrganizationDeleted struct {
	*BaseDomainEvent
	OrganizationCode string    `json:"organization_code"`
	Name             string    `json:"name"`
	DeletedAt        time.Time `json:"deleted_at"`
	DeletedBy        string    `json:"deleted_by,omitempty"`
	DeletionReason   string    `json:"deletion_reason,omitempty"`
}

// NewOrganizationDeleted 创建组织删除事件
func NewOrganizationDeleted(tenantID, organizationID uuid.UUID, code, name string) *OrganizationDeleted {
	base := NewBaseDomainEvent("organization.deleted", "organization", organizationID, tenantID)
	return &OrganizationDeleted{
		BaseDomainEvent:  base,
		OrganizationCode: code,
		Name:             name,
		DeletedAt:        time.Now(),
	}
}

func (e *OrganizationDeleted) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// OrganizationRestructured 组织重组事件
type OrganizationRestructured struct {
	*BaseDomainEvent
	OrganizationCode  string     `json:"organization_code"`
	OldParentID       *uuid.UUID `json:"old_parent_id,omitempty"`
	NewParentID       *uuid.UUID `json:"new_parent_id,omitempty"`
	OldLevel          int        `json:"old_level"`
	NewLevel          int        `json:"new_level"`
	RestructureType   string     `json:"restructure_type"` // move, merge, split
	RestructureReason string     `json:"restructure_reason,omitempty"`
	EffectiveDate     time.Time  `json:"effective_date"`
}

// NewOrganizationRestructured 创建组织重组事件
func NewOrganizationRestructured(tenantID, organizationID uuid.UUID, code string, oldParentID, newParentID *uuid.UUID, oldLevel, newLevel int, restructureType string) *OrganizationRestructured {
	base := NewBaseDomainEvent("organization.restructured", "organization", organizationID, tenantID)
	return &OrganizationRestructured{
		BaseDomainEvent:   base,
		OrganizationCode:  code,
		OldParentID:       oldParentID,
		NewParentID:       newParentID,
		OldLevel:          oldLevel,
		NewLevel:          newLevel,
		RestructureType:   restructureType,
		EffectiveDate:     time.Now(),
	}
}

func (e *OrganizationRestructured) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// OrganizationActivated 组织激活事件
type OrganizationActivated struct {
	*BaseDomainEvent
	OrganizationCode string    `json:"organization_code"`
	Name             string    `json:"name"`
	ActivatedAt      time.Time `json:"activated_at"`
	ActivatedBy      string    `json:"activated_by,omitempty"`
}

// NewOrganizationActivated 创建组织激活事件
func NewOrganizationActivated(tenantID, organizationID uuid.UUID, code, name string) *OrganizationActivated {
	base := NewBaseDomainEvent("organization.activated", "organization", organizationID, tenantID)
	return &OrganizationActivated{
		BaseDomainEvent:  base,
		OrganizationCode: code,
		Name:             name,
		ActivatedAt:      time.Now(),
	}
}

func (e *OrganizationActivated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// OrganizationDeactivated 组织停用事件
type OrganizationDeactivated struct {
	*BaseDomainEvent
	OrganizationCode   string    `json:"organization_code"`
	Name               string    `json:"name"`
	DeactivatedAt      time.Time `json:"deactivated_at"`
	DeactivatedBy      string    `json:"deactivated_by,omitempty"`
	DeactivationReason string    `json:"deactivation_reason,omitempty"`
}

// NewOrganizationDeactivated 创建组织停用事件
func NewOrganizationDeactivated(tenantID, organizationID uuid.UUID, code, name string) *OrganizationDeactivated {
	base := NewBaseDomainEvent("organization.deactivated", "organization", organizationID, tenantID)
	return &OrganizationDeactivated{
		BaseDomainEvent:    base,
		OrganizationCode:   code,
		Name:               name,
		DeactivatedAt:      time.Now(),
	}
}

func (e *OrganizationDeactivated) Serialize() ([]byte, error) {
	return json.Marshal(e)
}