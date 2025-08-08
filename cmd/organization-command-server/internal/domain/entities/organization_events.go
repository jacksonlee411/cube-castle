package entities

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	GetEventID() uuid.UUID
	GetAggregateID() string
	GetTenantID() uuid.UUID
	GetEventType() string
	GetEventTime() time.Time
}

// OrganizationCreatedEvent represents an organization creation event
type OrganizationCreatedEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID string    `json:"aggregate_id"` // Organization code
	TenantID    uuid.UUID `json:"tenant_id"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unit_type"`
	ParentCode  *string   `json:"parent_code,omitempty"`
	Level       int       `json:"level"`
	Path        string    `json:"path"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (e OrganizationCreatedEvent) GetEventID() uuid.UUID   { return e.EventID }
func (e OrganizationCreatedEvent) GetAggregateID() string  { return e.AggregateID }
func (e OrganizationCreatedEvent) GetTenantID() uuid.UUID  { return e.TenantID }
func (e OrganizationCreatedEvent) GetEventType() string    { return "OrganizationCreated" }
func (e OrganizationCreatedEvent) GetEventTime() time.Time { return e.CreatedAt }

// NewOrganizationCreatedEvent creates a new organization created event
func NewOrganizationCreatedEvent(code, name string, unitType UnitType, tenantID, createdBy uuid.UUID, parentCode *string, level int, path string) OrganizationCreatedEvent {
	return OrganizationCreatedEvent{
		EventID:     uuid.New(),
		AggregateID: code,
		TenantID:    tenantID,
		Name:        name,
		UnitType:    unitType.String(),
		ParentCode:  parentCode,
		Level:       level,
		Path:        path,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}
}

// OrganizationUpdatedEvent represents an organization update event
type OrganizationUpdatedEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Changes     map[string]interface{} `json:"changes"`
	UpdatedBy   uuid.UUID              `json:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

func (e OrganizationUpdatedEvent) GetEventID() uuid.UUID   { return e.EventID }
func (e OrganizationUpdatedEvent) GetAggregateID() string  { return e.AggregateID }
func (e OrganizationUpdatedEvent) GetTenantID() uuid.UUID  { return e.TenantID }
func (e OrganizationUpdatedEvent) GetEventType() string    { return "OrganizationUpdated" }
func (e OrganizationUpdatedEvent) GetEventTime() time.Time { return e.UpdatedAt }

// NewOrganizationUpdatedEvent creates a new organization updated event
func NewOrganizationUpdatedEvent(code string, tenantID, updatedBy uuid.UUID, changes map[string]interface{}) OrganizationUpdatedEvent {
	return OrganizationUpdatedEvent{
		EventID:     uuid.New(),
		AggregateID: code,
		TenantID:    tenantID,
		Changes:     changes,
		UpdatedBy:   updatedBy,
		UpdatedAt:   time.Now(),
	}
}

// OrganizationDeletedEvent represents an organization deletion event
type OrganizationDeletedEvent struct {
	EventID     uuid.UUID `json:"event_id"`
	AggregateID string    `json:"aggregate_id"`
	TenantID    uuid.UUID `json:"tenant_id"`
	DeletedBy   uuid.UUID `json:"deleted_by"`
	DeletedAt   time.Time `json:"deleted_at"`
}

func (e OrganizationDeletedEvent) GetEventID() uuid.UUID   { return e.EventID }
func (e OrganizationDeletedEvent) GetAggregateID() string  { return e.AggregateID }
func (e OrganizationDeletedEvent) GetTenantID() uuid.UUID  { return e.TenantID }
func (e OrganizationDeletedEvent) GetEventType() string    { return "OrganizationDeleted" }
func (e OrganizationDeletedEvent) GetEventTime() time.Time { return e.DeletedAt }

// NewOrganizationDeletedEvent creates a new organization deleted event
func NewOrganizationDeletedEvent(code string, tenantID, deletedBy uuid.UUID) OrganizationDeletedEvent {
	return OrganizationDeletedEvent{
		EventID:     uuid.New(),
		AggregateID: code,
		TenantID:    tenantID,
		DeletedBy:   deletedBy,
		DeletedAt:   time.Now(),
	}
}