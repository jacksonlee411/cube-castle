package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/valueobjects"
)

// Organization represents an organizational unit aggregate root
type Organization struct {
	code        valueobjects.OrganizationCode
	name        string
	unitType    UnitType
	status      Status
	parentCode  *valueobjects.OrganizationCode
	level       int
	path        string
	sortOrder   int
	description *string
	tenantID    uuid.UUID
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

// NewOrganization creates a new organization entity
func NewOrganization(
	code valueobjects.OrganizationCode,
	name string,
	unitType UnitType,
	tenantID uuid.UUID,
	parentCode *valueobjects.OrganizationCode,
	level int,
	path string,
	sortOrder int,
	description *string,
) (*Organization, error) {
	// Validate business rules
	if err := validateOrganizationRules(name, level, sortOrder); err != nil {
		return nil, err
	}

	now := time.Now()
	org := &Organization{
		code:        code,
		name:        name,
		unitType:    unitType,
		status:      StatusActive,
		parentCode:  parentCode,
		level:       level,
		path:        path,
		sortOrder:   sortOrder,
		description: description,
		tenantID:    tenantID,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	// Record creation event
	var parentCodeStr *string
	if parentCode != nil && !parentCode.IsEmpty() {
		s := parentCode.String()
		parentCodeStr = &s
	}

	event := NewOrganizationCreatedEvent(
		code.String(),
		name,
		unitType,
		tenantID,
		uuid.New(), // Should be passed from command
		parentCodeStr,
		level,
		path,
	)
	org.recordEvent(event)

	return org, nil
}

// Getters
func (o *Organization) Code() valueobjects.OrganizationCode { return o.code }
func (o *Organization) Name() string                        { return o.name }
func (o *Organization) UnitType() UnitType                  { return o.unitType }
func (o *Organization) Status() Status                      { return o.status }
func (o *Organization) ParentCode() *valueobjects.OrganizationCode { return o.parentCode }
func (o *Organization) Level() int                          { return o.level }
func (o *Organization) Path() string                        { return o.path }
func (o *Organization) SortOrder() int                      { return o.sortOrder }
func (o *Organization) Description() *string                { return o.description }
func (o *Organization) TenantID() uuid.UUID                 { return o.tenantID }
func (o *Organization) CreatedAt() time.Time                { return o.createdAt }
func (o *Organization) UpdatedAt() time.Time                { return o.updatedAt }
func (o *Organization) GetEvents() []DomainEvent            { return o.events }

// Business methods
func (o *Organization) UpdateName(newName string, updatedBy uuid.UUID) error {
	if strings.TrimSpace(newName) == "" {
		return ErrEmptyOrganizationName
	}
	
	if len(newName) > 100 {
		return ErrOrganizationNameTooLong
	}
	
	if o.name == newName {
		return nil // No change needed
	}
	
	oldName := o.name
	o.name = newName
	o.updatedAt = time.Now()
	
	// Record update event
	changes := map[string]interface{}{
		"name": map[string]interface{}{
			"old": oldName,
			"new": newName,
		},
	}
	
	event := NewOrganizationUpdatedEvent(o.code.String(), o.tenantID, updatedBy, changes)
	o.recordEvent(event)
	
	return nil
}

func (o *Organization) UpdateStatus(newStatus Status, updatedBy uuid.UUID) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %v", newStatus)
	}
	
	if o.status == newStatus {
		return nil // No change needed
	}
	
	oldStatus := o.status
	o.status = newStatus
	o.updatedAt = time.Now()
	
	// Record update event
	changes := map[string]interface{}{
		"status": map[string]interface{}{
			"old": oldStatus.String(),
			"new": newStatus.String(),
		},
	}
	
	event := NewOrganizationUpdatedEvent(o.code.String(), o.tenantID, updatedBy, changes)
	o.recordEvent(event)
	
	return nil
}

func (o *Organization) UpdateDescription(newDescription *string, updatedBy uuid.UUID) error {
	// Check if there's actually a change
	if (o.description == nil && newDescription == nil) ||
		(o.description != nil && newDescription != nil && *o.description == *newDescription) {
		return nil // No change needed
	}
	
	var oldDesc, newDesc interface{}
	if o.description != nil {
		oldDesc = *o.description
	}
	if newDescription != nil {
		newDesc = *newDescription
	}
	
	o.description = newDescription
	o.updatedAt = time.Now()
	
	// Record update event
	changes := map[string]interface{}{
		"description": map[string]interface{}{
			"old": oldDesc,
			"new": newDesc,
		},
	}
	
	event := NewOrganizationUpdatedEvent(o.code.String(), o.tenantID, updatedBy, changes)
	o.recordEvent(event)
	
	return nil
}

func (o *Organization) UpdateSortOrder(newSortOrder int, updatedBy uuid.UUID) error {
	if newSortOrder < 0 {
		return ErrInvalidSortOrder
	}
	
	if o.sortOrder == newSortOrder {
		return nil // No change needed
	}
	
	oldSortOrder := o.sortOrder
	o.sortOrder = newSortOrder
	o.updatedAt = time.Now()
	
	// Record update event
	changes := map[string]interface{}{
		"sort_order": map[string]interface{}{
			"old": oldSortOrder,
			"new": newSortOrder,
		},
	}
	
	event := NewOrganizationUpdatedEvent(o.code.String(), o.tenantID, updatedBy, changes)
	o.recordEvent(event)
	
	return nil
}

func (o *Organization) MarkAsDeleted(deletedBy uuid.UUID) error {
	if o.status == StatusInactive {
		return nil // Already deleted
	}
	
	o.status = StatusInactive
	o.updatedAt = time.Now()
	
	event := NewOrganizationDeletedEvent(o.code.String(), o.tenantID, deletedBy)
	o.recordEvent(event)
	
	return nil
}

// Helper methods
func (o *Organization) recordEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

func (o *Organization) ClearEvents() {
	o.events = make([]DomainEvent, 0)
}

func (o *Organization) HasChildren() bool {
	// This should be determined by the repository
	// For now, we'll assume it's checked elsewhere
	return false
}

// validateOrganizationRules validates business rules for organization creation
func validateOrganizationRules(name string, level, sortOrder int) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyOrganizationName
	}
	
	if len(name) > 100 {
		return ErrOrganizationNameTooLong
	}
	
	if level < 1 || level > 10 {
		return ErrInvalidOrganizationLevel
	}
	
	if sortOrder < 0 {
		return ErrInvalidSortOrder
	}
	
	return nil
}