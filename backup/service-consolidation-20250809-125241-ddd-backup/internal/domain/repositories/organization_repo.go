package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/entities"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/valueobjects"
)

// OrganizationRepository defines the interface for organization persistence
type OrganizationRepository interface {
	// Create a new organization
	Create(ctx context.Context, org *entities.Organization) error
	
	// Update an existing organization
	Update(ctx context.Context, org *entities.Organization) error
	
	// Delete an organization (soft delete)
	Delete(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) error
	
	// Find organization by code
	FindByCode(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (*entities.Organization, error)
	
	// Find children organizations
	FindChildren(ctx context.Context, parentCode valueobjects.OrganizationCode, tenantID uuid.UUID) ([]*entities.Organization, error)
	
	// Check if organization exists
	Exists(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (bool, error)
	
	// Generate next available organization code
	GenerateNextCode(ctx context.Context, tenantID uuid.UUID) (valueobjects.OrganizationCode, error)
	
	// Check if organization has children
	HasChildren(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) (bool, error)
	
	// Get parent information for level and path calculation
	GetParentInfo(ctx context.Context, parentCode valueobjects.OrganizationCode, tenantID uuid.UUID) (*ParentInfo, error)
}

// ParentInfo contains information about parent organization
type ParentInfo struct {
	Level int
	Path  string
}