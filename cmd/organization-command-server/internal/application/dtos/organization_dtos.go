package dtos

import (
	"time"
	
	"github.com/google/uuid"
)

// CreateOrganizationResult represents the result of creating an organization
type CreateOrganizationResult struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	UnitType  string    `json:"unit_type"`
	Status    string    `json:"status"`
	Level     int       `json:"level"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateOrganizationResult represents the result of updating an organization
type UpdateOrganizationResult struct {
	Code      string                 `json:"code"`
	UpdatedAt time.Time              `json:"updated_at"`
	Changes   map[string]interface{} `json:"changes"`
}

// DeleteOrganizationResult represents the result of deleting an organization
type DeleteOrganizationResult struct {
	Code      string    `json:"code"`
	DeletedAt time.Time `json:"deleted_at"`
}

// OrganizationDTO represents an organization data transfer object
type OrganizationDTO struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	UnitType    string     `json:"unit_type"`
	Status      string     `json:"status"`
	ParentCode  *string    `json:"parent_code,omitempty"`
	Level       int        `json:"level"`
	Path        string     `json:"path"`
	SortOrder   int        `json:"sort_order"`
	Description *string    `json:"description,omitempty"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}