package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/entities"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/repositories"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/valueobjects"
)

// OrganizationService contains domain business logic
type OrganizationService struct {
	repo repositories.OrganizationRepository
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(repo repositories.OrganizationRepository) *OrganizationService {
	return &OrganizationService{
		repo: repo,
	}
}

// ValidateCreateOrganization validates business rules for creating an organization
func (s *OrganizationService) ValidateCreateOrganization(ctx context.Context, code valueobjects.OrganizationCode, name string, parentCode *valueobjects.OrganizationCode, tenantID uuid.UUID) error {
	// Check if organization code already exists (if provided)
	if !code.IsEmpty() {
		exists, err := s.repo.Exists(ctx, code, tenantID)
		if err != nil {
			return fmt.Errorf("failed to check organization existence: %w", err)
		}
		if exists {
			return fmt.Errorf("organization with code '%s' already exists", code.String())
		}
	}
	
	// Check if parent organization exists (if provided)
	if parentCode != nil && !parentCode.IsEmpty() {
		parentExists, err := s.repo.Exists(ctx, *parentCode, tenantID)
		if err != nil {
			return fmt.Errorf("failed to check parent organization existence: %w", err)
		}
		if !parentExists {
			return fmt.Errorf("parent organization '%s' does not exist", parentCode.String())
		}
	}
	
	return nil
}

// ValidateDeleteOrganization validates business rules for deleting an organization
func (s *OrganizationService) ValidateDeleteOrganization(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) error {
	// Check if organization exists
	exists, err := s.repo.Exists(ctx, code, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check organization existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("organization '%s' does not exist", code.String())
	}
	
	// Check if organization has children
	hasChildren, err := s.repo.HasChildren(ctx, code, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check for child organizations: %w", err)
	}
	if hasChildren {
		return entities.ErrCannotDeleteOrganizationWithChildren
	}
	
	return nil
}

// ValidateUpdateOrganization validates business rules for updating an organization
func (s *OrganizationService) ValidateUpdateOrganization(ctx context.Context, code valueobjects.OrganizationCode, tenantID uuid.UUID) error {
	// Check if organization exists
	exists, err := s.repo.Exists(ctx, code, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check organization existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("organization '%s' does not exist", code.String())
	}
	
	return nil
}

// CalculateOrganizationHierarchy calculates level and path for an organization
func (s *OrganizationService) CalculateOrganizationHierarchy(ctx context.Context, code valueobjects.OrganizationCode, parentCode *valueobjects.OrganizationCode, tenantID uuid.UUID) (level int, path string, err error) {
	// Default values for root organization
	level = 1
	path = fmt.Sprintf("/%s", code.String())
	
	// If parent is specified, calculate based on parent
	if parentCode != nil && !parentCode.IsEmpty() {
		parentInfo, err := s.repo.GetParentInfo(ctx, *parentCode, tenantID)
		if err != nil {
			return 0, "", fmt.Errorf("failed to get parent organization info: %w", err)
		}
		
		level = parentInfo.Level + 1
		path = fmt.Sprintf("%s/%s", parentInfo.Path, code.String())
		
		// Validate maximum depth
		if level > 10 {
			return 0, "", fmt.Errorf("maximum organization hierarchy depth exceeded (max: 10, calculated: %d)", level)
		}
	}
	
	return level, path, nil
}

// GenerateOrganizationCode generates a new unique organization code
func (s *OrganizationService) GenerateOrganizationCode(ctx context.Context, tenantID uuid.UUID) (valueobjects.OrganizationCode, error) {
	return s.repo.GenerateNextCode(ctx, tenantID)
}