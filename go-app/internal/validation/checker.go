package validation

import (
	"context"
	"fmt"

	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/google/uuid"
)

// CoreHRValidationChecker implements all validation checker interfaces using CoreHR repository
type CoreHRValidationChecker struct {
	repo *corehr.Repository
}

// NewCoreHRValidationChecker creates a new validation checker using CoreHR repository
func NewCoreHRValidationChecker(repo *corehr.Repository) *CoreHRValidationChecker {
	return &CoreHRValidationChecker{repo: repo}
}

// IsEmployeeNumberExists checks if an employee number already exists (excluding a specific employee ID)
func (c *CoreHRValidationChecker) IsEmployeeNumberExists(ctx context.Context, tenantID uuid.UUID, employeeNumber string, excludeID *uuid.UUID) (bool, error) {
	if c.repo == nil {
		return false, fmt.Errorf("repository not available")
	}

	employee, err := c.repo.GetEmployeeByNumber(ctx, tenantID, employeeNumber)
	if err != nil {
		// If error contains "no rows", employee doesn't exist
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check employee number existence: %w", err)
	}

	if employee == nil {
		return false, nil
	}

	// If we're excluding a specific employee ID, check if this is the same employee
	if excludeID != nil && employee.ID == *excludeID {
		return false, nil
	}

	return true, nil
}

// IsEmailExists checks if an email already exists (excluding a specific employee ID)
func (c *CoreHRValidationChecker) IsEmailExists(ctx context.Context, tenantID uuid.UUID, email string, excludeID *uuid.UUID) (bool, error) {
	if c.repo == nil {
		return false, fmt.Errorf("repository not available")
	}

	// Note: This assumes the repository has a method to check email existence
	// If not implemented in CoreHR repository, we'll need to add it
	employee, err := c.repo.GetEmployeeByEmail(ctx, tenantID, email)
	if err != nil {
		// If error contains "no rows", email doesn't exist
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	if employee == nil {
		return false, nil
	}

	// If we're excluding a specific employee ID, check if this is the same employee
	if excludeID != nil && employee.ID == *excludeID {
		return false, nil
	}

	return true, nil
}

// IsOrganizationExists checks if an organization exists
func (c *CoreHRValidationChecker) IsOrganizationExists(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID) (bool, error) {
	if c.repo == nil {
		return false, fmt.Errorf("repository not available")
	}

	// Note: This assumes the repository has a method to check organization existence
	// If not implemented in CoreHR repository, we'll need to add it
	org, err := c.repo.GetOrganizationByID(ctx, tenantID, orgID)
	if err != nil {
		// If error contains "no rows", organization doesn't exist
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check organization existence: %w", err)
	}

	return org != nil, nil
}

// IsPositionExists checks if a position exists
func (c *CoreHRValidationChecker) IsPositionExists(ctx context.Context, tenantID uuid.UUID, positionID uuid.UUID) (bool, error) {
	if c.repo == nil {
		return false, fmt.Errorf("repository not available")
	}

	// Note: This assumes the repository has a method to check position existence
	// If not implemented in CoreHR repository, we'll need to add it
	position, err := c.repo.GetPositionByID(ctx, tenantID, positionID)
	if err != nil {
		// If error contains "no rows", position doesn't exist
		if err.Error() == "no rows in result set" || err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check position existence: %w", err)
	}

	return position != nil, nil
}

// MockValidationChecker provides mock implementations for testing
type MockValidationChecker struct{}

// NewMockValidationChecker creates a new mock validation checker
func NewMockValidationChecker() *MockValidationChecker {
	return &MockValidationChecker{}
}

// IsEmployeeNumberExists mock implementation - always returns false (no duplicates)
func (m *MockValidationChecker) IsEmployeeNumberExists(ctx context.Context, tenantID uuid.UUID, employeeNumber string, excludeID *uuid.UUID) (bool, error) {
	// Mock implementation - for testing, assume no duplicates
	return false, nil
}

// IsEmailExists mock implementation - always returns false (no duplicates)
func (m *MockValidationChecker) IsEmailExists(ctx context.Context, tenantID uuid.UUID, email string, excludeID *uuid.UUID) (bool, error) {
	// Mock implementation - for testing, assume no duplicates
	return false, nil
}

// IsOrganizationExists mock implementation - always returns true (organization exists)
func (m *MockValidationChecker) IsOrganizationExists(ctx context.Context, tenantID uuid.UUID, orgID uuid.UUID) (bool, error) {
	// Mock implementation - for testing, assume organization exists
	return true, nil
}

// IsPositionExists mock implementation - always returns true (position exists)
func (m *MockValidationChecker) IsPositionExists(ctx context.Context, tenantID uuid.UUID, positionID uuid.UUID) (bool, error) {
	// Mock implementation - for testing, assume position exists
	return true, nil
}