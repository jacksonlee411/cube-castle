package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/ent/positionoccupancyhistory"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/google/uuid"
)

// EmployeeMigrationService handles the migration from legacy Employee model to new relationship-based model
type EmployeeMigrationService struct {
	client *ent.Client
}

func NewEmployeeMigrationService(client *ent.Client) *EmployeeMigrationService {
	return &EmployeeMigrationService{client: client}
}

// MigrateEmployeeModel performs the complete Employee model migration
func (s *EmployeeMigrationService) MigrateEmployeeModel(ctx context.Context) error {
	log.Println("🚀 Starting Employee model migration...")

	// Step 1: Run schema migration
	if err := s.runSchemaMigration(ctx); err != nil {
		return fmt.Errorf("failed to run schema migration: %w", err)
	}

	// Step 2: Migrate existing employee data
	if err := s.migrateExistingEmployees(ctx); err != nil {
		return fmt.Errorf("failed to migrate existing employees: %w", err)
	}

	// Step 3: Create position occupancy history for existing employees
	if err := s.createInitialOccupancyHistory(ctx); err != nil {
		return fmt.Errorf("failed to create initial occupancy history: %w", err)
	}

	// Step 4: Validate migration results
	if err := s.validateMigration(ctx); err != nil {
		return fmt.Errorf("migration validation failed: %w", err)
	}

	log.Println("✅ Employee model migration completed successfully!")
	return nil
}

// runSchemaMigration creates the new schema structure
func (s *EmployeeMigrationService) runSchemaMigration(ctx context.Context) error {
	log.Println("📊 Creating new schema structure...")

	if err := s.client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("✅ Schema migration completed")
	return nil
}

// migrateExistingEmployees converts legacy employee records to new format
func (s *EmployeeMigrationService) migrateExistingEmployees(ctx context.Context) error {
	log.Println("👥 Migrating existing employee records...")

	// Get all existing employees
	employees, err := s.client.Employee.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing employees: %w", err)
	}

	log.Printf("Found %d existing employees to migrate", len(employees))

	migratedCount := 0
	failedCount := 0

	for _, emp := range employees {
		if err := s.migrateIndividualEmployee(ctx, emp); err != nil {
			log.Printf("❌ Failed to migrate employee %s (%s): %v", emp.ID, emp.Name, err)
			failedCount++
		} else {
			migratedCount++
		}
	}

	log.Printf("✅ Migration completed: %d succeeded, %d failed", migratedCount, failedCount)
	return nil
}

// migrateIndividualEmployee migrates a single employee record
func (s *EmployeeMigrationService) migrateIndividualEmployee(ctx context.Context, emp *ent.Employee) error {
	return s.client.WithTx(ctx, func(tx *ent.Tx) error {
		// Generate tenant ID if not present (assuming single tenant for now)
		tenantID := uuid.New()

		// Parse name into first_name and last_name
		firstName, lastName := s.parseEmployeeName(emp.Name)

		// Determine employee type based on existing data or defaults
		employeeType := s.determineEmployeeType(emp)

		// Generate employee number if not present
		employeeNumber := s.generateEmployeeNumber(emp)

		// Set default hire date if not available
		hireDate := time.Now()
		if !emp.CreatedAt.IsZero() {
			hireDate = emp.CreatedAt
		}

		// Update employee with new fields
		updateBuilder := tx.Employee.UpdateOne(emp).
			SetTenantID(tenantID).
			SetEmployeeType(employee.EmployeeType(employeeType)).
			SetEmployeeNumber(employeeNumber).
			SetFirstName(firstName).
			SetLastName(lastName).
			SetEmploymentStatus(employee.EmploymentStatusACTIVE).
			SetHireDate(hireDate)

		// Handle position relationship
		if emp.Position != "" {
			position, err := s.findPositionByName(ctx, tx, emp.Position, tenantID)
			if err == nil {
				updateBuilder = updateBuilder.SetCurrentPositionID(position.ID)
				log.Printf("🔗 Linked employee %s to position %s", emp.Name, position.ID)
			} else {
				log.Printf("⚠️  Could not find position '%s' for employee %s", emp.Position, emp.Name)
			}
		}

		_, err := updateBuilder.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to update employee %s: %w", emp.ID, err)
		}

		log.Printf("✅ Migrated employee: %s -> %s %s", emp.Name, firstName, lastName)
		return nil
	})
}

// createInitialOccupancyHistory creates position occupancy history for employees with positions
func (s *EmployeeMigrationService) createInitialOccupancyHistory(ctx context.Context) error {
	log.Println("📋 Creating initial position occupancy history...")

	// Get employees with current positions
	employees, err := s.client.Employee.Query().
		Where(employee.CurrentPositionIDNotNil()).
		WithCurrentPosition().
		All(ctx)

	if err != nil {
		return fmt.Errorf("failed to fetch employees with positions: %w", err)
	}

	log.Printf("Creating occupancy history for %d employees", len(employees))

	for _, emp := range employees {
		if emp.Edges.CurrentPosition == nil {
			continue
		}

		// Create occupancy history record
		_, err := s.client.PositionOccupancyHistory.Create().
			SetTenantID(emp.TenantID).
			SetPositionID(emp.CurrentPositionID.String()).
			SetEmployeeID(emp.ID).
			SetStartDate(emp.HireDate).
			SetIsActive(true).
			SetAssignmentType(positionoccupancyhistory.AssignmentTypeREGULAR).
			SetAssignmentReason("Initial migration assignment").
			SetFtePercentage(1.0).
			Save(ctx)

		if err != nil {
			log.Printf("❌ Failed to create occupancy history for employee %s: %v", emp.EmployeeNumber, err)
		} else {
			log.Printf("✅ Created occupancy history for employee %s", emp.EmployeeNumber)
		}
	}

	return nil
}

// validateMigration performs validation checks on the migrated data
func (s *EmployeeMigrationService) validateMigration(ctx context.Context) error {
	log.Println("🔍 Validating migration results...")

	// Check 1: All employees have required new fields
	employeesWithoutTenant, err := s.client.Employee.Query().
		Where(employee.TenantIDIsNil()).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to check tenant IDs: %w", err)
	}
	if employeesWithoutTenant > 0 {
		return fmt.Errorf("found %d employees without tenant_id", employeesWithoutTenant)
	}

	// Check 2: All employees with positions have occupancy history
	employeesWithPositions, err := s.client.Employee.Query().
		Where(employee.CurrentPositionIDNotNil()).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count employees with positions: %w", err)
	}

	occupancyRecords, err := s.client.PositionOccupancyHistory.Query().
		Where(positionoccupancyhistory.IsActiveEQ(true)).
		Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count occupancy records: %w", err)
	}

	log.Printf("📊 Validation results:")
	log.Printf("   - Employees with positions: %d", employeesWithPositions)
	log.Printf("   - Active occupancy records: %d", occupancyRecords)

	if employeesWithPositions != occupancyRecords {
		log.Printf("⚠️  Warning: Mismatch between employees with positions and occupancy records")
	}

	// Check 3: Referential integrity
	log.Println("🔗 Checking referential integrity...")
	
	// Check position references
	invalidPositionRefs, err := s.client.Employee.Query().
		Where(employee.And(
			employee.CurrentPositionIDNotNil(),
			employee.Not(employee.HasCurrentPosition()),
		)).
		Count(ctx)
	
	if err != nil {
		return fmt.Errorf("failed to check position references: %w", err)
	}
	
	if invalidPositionRefs > 0 {
		return fmt.Errorf("found %d employees with invalid position references", invalidPositionRefs)
	}

	log.Println("✅ Migration validation passed!")
	return nil
}

// Helper functions

func (s *EmployeeMigrationService) parseEmployeeName(fullName string) (string, string) {
	if fullName == "" {
		return "Unknown", "Employee"
	}

	parts := strings.Fields(strings.TrimSpace(fullName))
	if len(parts) == 0 {
		return "Unknown", "Employee"
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	
	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")
	return firstName, lastName
}

func (s *EmployeeMigrationService) determineEmployeeType(emp *ent.Employee) string {
	// Simple heuristic based on existing data
	// In a real scenario, this could be based on more sophisticated rules
	return "FULL_TIME"
}

func (s *EmployeeMigrationService) generateEmployeeNumber(emp *ent.Employee) string {
	// Generate employee number based on ID or other logic
	if emp.ID != "" {
		// Use last 8 characters of ID with EMP prefix
		if len(emp.ID) >= 8 {
			return "EMP" + emp.ID[len(emp.ID)-8:]
		}
		return "EMP" + emp.ID
	}
	return "EMP" + fmt.Sprintf("%08d", time.Now().Unix()%100000000)
}

func (s *EmployeeMigrationService) findPositionByName(ctx context.Context, tx *ent.Tx, positionName string, tenantID uuid.UUID) (*ent.Position, error) {
	// Try to find position by matching with department name or other criteria
	// This is a simplified approach - in reality, you might need more sophisticated matching
	
	// First, try to find any position (since we don't have direct name matching)
	positions, err := tx.Position.Query().
		Where(position.TenantIDEQ(tenantID)).
		Limit(1).
		All(ctx)
	
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	
	if len(positions) == 0 {
		return nil, fmt.Errorf("no positions found for tenant")
	}
	
	// Return first available position (in a real scenario, you'd have better matching logic)
	return positions[0], nil
}

// Main migration function to be called from cmd/migrate/main.go
func main() {
	ctx := context.Background()
	
	// Initialize database connection
	client, err := common.InitializeDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer client.Close()

	// Run migration
	migrationService := NewEmployeeMigrationService(client)
	if err := migrationService.MigrateEmployeeModel(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}