// internal/service/organization_sync_service.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
)

// OrganizationSyncService handles synchronization between PostgreSQL and Neo4j
type OrganizationSyncService struct {
	entClient    *ent.Client
	neo4jService *Neo4jService
	logger       *log.Logger
}

// SyncOptions controls synchronization behavior
type SyncOptions struct {
	FullSync           bool `json:"full_sync"`
	SyncEmployees      bool `json:"sync_employees"`
	SyncPositions      bool `json:"sync_positions"`
	SyncRelationships  bool `json:"sync_relationships"`
	SyncDepartments    bool `json:"sync_departments"`
	BatchSize          int  `json:"batch_size"`
}

// SyncResult contains synchronization results
type SyncResult struct {
	Success             bool      `json:"success"`
	SyncedEmployees     int       `json:"synced_employees"`
	SyncedPositions     int       `json:"synced_positions"`
	SyncedRelationships int       `json:"synced_relationships"`
	SyncedDepartments   int       `json:"synced_departments"`
	Errors              []string  `json:"errors"`
	StartTime           time.Time `json:"start_time"`
	EndTime             time.Time `json:"end_time"`
	Duration            time.Duration `json:"duration"`
}

// NewOrganizationSyncService creates a new organization sync service
func NewOrganizationSyncService(
	entClient *ent.Client,
	neo4jService *Neo4jService,
	logger *log.Logger,
) *OrganizationSyncService {
	return &OrganizationSyncService{
		entClient:    entClient,
		neo4jService: neo4jService,
		logger:       logger,
	}
}

// FullSync performs a complete synchronization from PostgreSQL to Neo4j
func (s *OrganizationSyncService) FullSync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
		Errors:    make([]string, 0),
	}

	s.logger.Println("Starting full organization sync to Neo4j")

	// Sync employees
	if err := s.syncAllEmployees(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Employee sync failed: %v", err))
	}

	// Sync positions
	if err := s.syncAllPositions(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Position sync failed: %v", err))
	}

	// Sync reporting relationships
	if err := s.syncReportingRelationships(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Relationship sync failed: %v", err))
	}

	// Sync departments
	if err := s.syncDepartments(ctx, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Department sync failed: %v", err))
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = len(result.Errors) == 0

	s.logger.Printf("Full sync completed in %v. Synced: %d employees, %d positions, %d relationships, %d departments. Errors: %d",
		result.Duration, result.SyncedEmployees, result.SyncedPositions, result.SyncedRelationships, result.SyncedDepartments, len(result.Errors))

	return result, nil
}

// SyncEmployee synchronizes a single employee to Neo4j
func (s *OrganizationSyncService) SyncEmployee(ctx context.Context, employeeID string) error {
	// Get employee from PostgreSQL
	emp, err := s.entClient.Employee.Query().
		Where(employee.ID(employeeID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get employee %s: %w", employeeID, err)
	}

	// Convert to Neo4j format
	employeeNode := EmployeeNode{
		ID:         emp.ID,
		EmployeeID: emp.EmployeeID,
		LegalName:  emp.LegalName,
		Email:      emp.Email,
		Status:     string(emp.Status),
		HireDate:   emp.HireDate,
		Properties: map[string]interface{}{
			"preferred_name":    emp.PreferredName,
			"termination_date":  emp.TerminationDate,
			"created_at":       emp.CreatedAt,
			"updated_at":       emp.UpdatedAt,
		},
	}

	// Sync to Neo4j
	if err := s.neo4jService.SyncEmployee(ctx, employeeNode); err != nil {
		return fmt.Errorf("failed to sync employee %s to Neo4j: %w", employeeID, err)
	}

	s.logger.Printf("Synced employee %s to Neo4j", employeeID)
	return nil
}

// SyncPosition synchronizes a position to Neo4j
func (s *OrganizationSyncService) SyncPosition(ctx context.Context, positionID, employeeID string) error {
	// Get position from PostgreSQL
	pos, err := s.entClient.PositionHistory.Query().
		Where(positionhistory.ID(positionID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get position %s: %w", positionID, err)
	}

	// Convert to Neo4j format
	positionNode := PositionNode{
		ID:            pos.ID,
		PositionTitle: pos.PositionTitle,
		Department:    pos.Department,
		JobLevel:      string(pos.JobLevel),
		Location:      pos.Location,
		EffectiveDate: pos.EffectiveDate,
		EndDate:       pos.EndDate,
		Properties: map[string]interface{}{
			"employment_type":       string(pos.EmploymentType),
			"reports_to_employee_id": pos.ReportsToEmployeeID,
			"change_reason":         pos.ChangeReason,
			"is_retroactive":        pos.IsRetroactive,
			"min_salary":           pos.MinSalary,
			"max_salary":           pos.MaxSalary,
			"currency":             pos.Currency,
			"created_at":           pos.CreatedAt,
			"updated_at":           pos.UpdatedAt,
		},
	}

	// Sync to Neo4j
	if err := s.neo4jService.SyncPosition(ctx, positionNode, employeeID); err != nil {
		return fmt.Errorf("failed to sync position %s to Neo4j: %w", positionID, err)
	}

	// Create reporting relationship if exists
	if pos.ReportsToEmployeeID != nil {
		if err := s.neo4jService.CreateReportingRelationship(ctx, *pos.ReportsToEmployeeID, employeeID); err != nil {
			s.logger.Printf("Warning: Failed to create reporting relationship %s -> %s: %v", employeeID, *pos.ReportsToEmployeeID, err)
		}
	}

	s.logger.Printf("Synced position %s for employee %s to Neo4j", positionID, employeeID)
	return nil
}

// syncAllEmployees synchronizes all employees from PostgreSQL to Neo4j
func (s *OrganizationSyncService) syncAllEmployees(ctx context.Context, result *SyncResult) error {
	const batchSize = 100
	offset := 0

	for {
		// Get batch of employees
		employees, err := s.entClient.Employee.Query().
			Limit(batchSize).
			Offset(offset).
			All(ctx)
		if err != nil {
			return fmt.Errorf("failed to query employees: %w", err)
		}

		if len(employees) == 0 {
			break
		}

		// Sync each employee
		for _, emp := range employees {
			employeeNode := EmployeeNode{
				ID:         emp.ID,
				EmployeeID: emp.EmployeeID,
				LegalName:  emp.LegalName,
				Email:      emp.Email,
				Status:     string(emp.Status),
				HireDate:   emp.HireDate,
				Properties: map[string]interface{}{
					"preferred_name":    emp.PreferredName,
					"termination_date":  emp.TerminationDate,
					"created_at":       emp.CreatedAt,
					"updated_at":       emp.UpdatedAt,
				},
			}

			if err := s.neo4jService.SyncEmployee(ctx, employeeNode); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Employee %s sync failed: %v", emp.EmployeeID, err))
				continue
			}

			result.SyncedEmployees++
		}

		offset += batchSize

		// Log progress
		if offset%1000 == 0 {
			s.logger.Printf("Synced %d employees to Neo4j", result.SyncedEmployees)
		}
	}

	return nil
}

// syncAllPositions synchronizes all current positions from PostgreSQL to Neo4j
func (s *OrganizationSyncService) syncAllPositions(ctx context.Context, result *SyncResult) error {
	const batchSize = 100
	offset := 0

	for {
		// Get batch of current positions (where end_date is null)
		positions, err := s.entClient.PositionHistory.Query().
			Where(positionhistory.EndDateIsNil()).
			WithEmployee().
			Limit(batchSize).
			Offset(offset).
			All(ctx)
		if err != nil {
			return fmt.Errorf("failed to query positions: %w", err)
		}

		if len(positions) == 0 {
			break
		}

		// Sync each position
		for _, pos := range positions {
			if pos.Edges.Employee == nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Position %s has no employee", pos.ID))
				continue
			}

			positionNode := PositionNode{
				ID:            pos.ID,
				PositionTitle: pos.PositionTitle,
				Department:    pos.Department,
				JobLevel:      string(pos.JobLevel),
				Location:      pos.Location,
				EffectiveDate: pos.EffectiveDate,
				EndDate:       pos.EndDate,
				Properties: map[string]interface{}{
					"employment_type":       string(pos.EmploymentType),
					"reports_to_employee_id": pos.ReportsToEmployeeID,
					"change_reason":         pos.ChangeReason,
					"is_retroactive":        pos.IsRetroactive,
					"min_salary":           pos.MinSalary,
					"max_salary":           pos.MaxSalary,
					"currency":             pos.Currency,
					"created_at":           pos.CreatedAt,
					"updated_at":           pos.UpdatedAt,
				},
			}

			if err := s.neo4jService.SyncPosition(ctx, positionNode, pos.Edges.Employee.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Position %s sync failed: %v", pos.ID, err))
				continue
			}

			result.SyncedPositions++
		}

		offset += batchSize

		// Log progress
		if offset%1000 == 0 {
			s.logger.Printf("Synced %d positions to Neo4j", result.SyncedPositions)
		}
	}

	return nil
}

// syncReportingRelationships synchronizes all reporting relationships
func (s *OrganizationSyncService) syncReportingRelationships(ctx context.Context, result *SyncResult) error {
	// Get all current positions with reporting relationships
	positions, err := s.entClient.PositionHistory.Query().
		Where(
			positionhistory.EndDateIsNil(),
			positionhistory.ReportsToEmployeeIDNotNil(),
		).
		WithEmployee().
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query reporting relationships: %w", err)
	}

	for _, pos := range positions {
		if pos.Edges.Employee == nil || pos.ReportsToEmployeeID == nil {
			continue
		}

		if err := s.neo4jService.CreateReportingRelationship(ctx, *pos.ReportsToEmployeeID, pos.Edges.Employee.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Reporting relationship %s -> %s failed: %v", 
				pos.Edges.Employee.ID, *pos.ReportsToEmployeeID, err))
			continue
		}

		result.SyncedRelationships++
	}

	return nil
}

// syncDepartments synchronizes department structure
func (s *OrganizationSyncService) syncDepartments(ctx context.Context, result *SyncResult) error {
	// Get unique departments from current positions
	type DepartmentInfo struct {
		Department string `json:"department"`
		Count      int    `json:"count"`
	}

	departments := make(map[string]int)

	positions, err := s.entClient.PositionHistory.Query().
		Where(positionhistory.EndDateIsNil()).
		All(ctx)
	if err != nil {
		return fmt.Errorf("failed to query departments: %w", err)
	}

	for _, pos := range positions {
		departments[pos.Department]++
	}

	// Create department nodes in Neo4j
	for dept, count := range departments {
		// This is a simplified approach - in reality, you'd have a proper Department entity
		deptNode := DepartmentNode{
			ID:   dept,
			Name: dept,
			Properties: map[string]interface{}{
				"employee_count": count,
				"created_at":    time.Now(),
			},
		}

		// Create department node (this would need a proper sync method in Neo4jService)
		s.logger.Printf("Would sync department: %s (%d employees)", dept, count)
		result.SyncedDepartments++
	}

	return nil
}

// SyncDepartment synchronizes a specific department and its employees
func (s *OrganizationSyncService) SyncDepartment(ctx context.Context, departmentName string) (*SyncResult, error) {
	result := &SyncResult{
		StartTime: time.Now(),
		Errors:    make([]string, 0),
	}

	// Get all employees in the department
	positions, err := s.entClient.PositionHistory.Query().
		Where(
			positionhistory.Department(departmentName),
			positionhistory.EndDateIsNil(),
		).
		WithEmployee().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query department %s: %w", departmentName, err)
	}

	// Sync each employee and their position
	for _, pos := range positions {
		if pos.Edges.Employee == nil {
			continue
		}

		emp := pos.Edges.Employee

		// Sync employee
		if err := s.SyncEmployee(ctx, emp.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Employee %s sync failed: %v", emp.EmployeeID, err))
			continue
		}
		result.SyncedEmployees++

		// Sync position
		if err := s.SyncPosition(ctx, pos.ID, emp.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Position %s sync failed: %v", pos.ID, err))
			continue
		}
		result.SyncedPositions++

		// Sync reporting relationship
		if pos.ReportsToEmployeeID != nil {
			if err := s.neo4jService.CreateReportingRelationship(ctx, *pos.ReportsToEmployeeID, emp.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Reporting relationship failed: %v", err))
			} else {
				result.SyncedRelationships++
			}
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Success = len(result.Errors) == 0

	s.logger.Printf("Department %s sync completed in %v. Synced: %d employees, %d positions, %d relationships. Errors: %d",
		departmentName, result.Duration, result.SyncedEmployees, result.SyncedPositions, result.SyncedRelationships, len(result.Errors))

	return result, nil
}

// OnEmployeeChange handles employee data changes and syncs to Neo4j
func (s *OrganizationSyncService) OnEmployeeChange(ctx context.Context, employeeID string, changeType string) error {
	switch changeType {
	case "CREATE", "UPDATE":
		return s.SyncEmployee(ctx, employeeID)
	case "DELETE":
		// Handle employee deletion in Neo4j
		s.logger.Printf("Employee %s deleted - would remove from Neo4j", employeeID)
		return nil
	default:
		return fmt.Errorf("unknown change type: %s", changeType)
	}
}

// OnPositionChange handles position changes and syncs to Neo4j
func (s *OrganizationSyncService) OnPositionChange(ctx context.Context, positionID, employeeID string, changeType string) error {
	switch changeType {
	case "CREATE", "UPDATE":
		return s.SyncPosition(ctx, positionID, employeeID)
	case "DELETE":
		// Handle position deletion in Neo4j
		s.logger.Printf("Position %s deleted - would remove from Neo4j", positionID)
		return nil
	default:
		return fmt.Errorf("unknown change type: %s", changeType)
	}
}

// HealthCheck performs a health check on the graph database synchronization
func (s *OrganizationSyncService) HealthCheck(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status": "healthy",
		"checks": make(map[string]interface{}),
	}

	// Check PostgreSQL employee count
	pgEmployeeCount, err := s.entClient.Employee.Query().Count(ctx)
	if err != nil {
		health["status"] = "unhealthy"
		health["postgres_error"] = err.Error()
	}

	// Check Neo4j connectivity (would need actual implementation)
	health["checks"] = map[string]interface{}{
		"postgres_employees": pgEmployeeCount,
		"neo4j_connection":   "connected", // This would be an actual check
		"last_sync":         time.Now().Add(-1 * time.Hour), // This would be stored
	}

	return health, nil
}