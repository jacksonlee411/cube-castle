package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/employee"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	"github.com/gaogu/cube-castle/go-app/ent/positionoccupancyhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
)

// AnalyticsService provides complex query capabilities and reporting
// for Employee-Position relationships and organizational analytics
type AnalyticsService struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// HistoryQueryParams defines parameters for historical queries
type HistoryQueryParams struct {
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	EmployeeID   *uuid.UUID `json:"employee_id,omitempty"`
	PositionID   *uuid.UUID `json:"position_id,omitempty"`
	EmployeeType *string    `json:"employee_type,omitempty"`
	Status       *string    `json:"status,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// OrganizationalMetrics contains organizational analysis metrics
type OrganizationalMetrics struct {
	TenantID              uuid.UUID                        `json:"tenant_id"`
	ReportDate            time.Time                        `json:"report_date"`
	TotalEmployees        int                              `json:"total_employees"`
	ActiveEmployees       int                              `json:"active_employees"`
	TotalPositions        int                              `json:"total_positions"`
	FilledPositions       int                              `json:"filled_positions"`
	OpenPositions         int                              `json:"open_positions"`
	EmployeesByType       map[string]int                   `json:"employees_by_type"`
	EmployeesByStatus     map[string]int                   `json:"employees_by_status"`
	PositionsByStatus     map[string]int                   `json:"positions_by_status"`
	AverageAssignmentDuration float64                    `json:"average_assignment_duration_days"`
	TurnoverMetrics       TurnoverMetrics                  `json:"turnover_metrics"`
	AssignmentMetrics     AssignmentMetrics                `json:"assignment_metrics"`
}

// TurnoverMetrics contains employee turnover analysis
type TurnoverMetrics struct {
	TerminationsThisMonth    int     `json:"terminations_this_month"`
	TerminationsThisQuarter  int     `json:"terminations_this_quarter"`
	TerminationsThisYear     int     `json:"terminations_this_year"`
	HiresThisMonth          int     `json:"hires_this_month"`
	HiresThisQuarter        int     `json:"hires_this_quarter"`
	HiresThisYear           int     `json:"hires_this_year"`
	MonthlyTurnoverRate     float64 `json:"monthly_turnover_rate"`
	QuarterlyTurnoverRate   float64 `json:"quarterly_turnover_rate"`
	AnnualTurnoverRate      float64 `json:"annual_turnover_rate"`
}

// AssignmentMetrics contains position assignment analysis
type AssignmentMetrics struct {
	TotalAssignments        int                    `json:"total_assignments"`
	ActiveAssignments       int                    `json:"active_assignments"`
	AssignmentsByType       map[string]int         `json:"assignments_by_type"`
	AverageAssignmentLength float64                `json:"average_assignment_length_days"`
	PromotionsThisYear      int                    `json:"promotions_this_year"`
	TransfersThisYear       int                    `json:"transfers_this_year"`
	AssignmentTrends        []AssignmentTrendPoint `json:"assignment_trends"`
}

// AssignmentTrendPoint represents a point in assignment trends
type AssignmentTrendPoint struct {
	Date        time.Time `json:"date"`
	NewAssignments int    `json:"new_assignments"`
	EndedAssignments int  `json:"ended_assignments"`
	ActiveTotal  int      `json:"active_total"`
}

// EmployeeHistoryRecord contains detailed employee history information
type EmployeeHistoryRecord struct {
	Employee              *ent.Employee                   `json:"employee"`
	AssignmentHistory     []*PositionAssignmentSummary    `json:"assignment_history"`
	TotalAssignments      int                             `json:"total_assignments"`
	CurrentAssignment     *PositionAssignmentSummary      `json:"current_assignment,omitempty"`
	TotalTenureDays       int                             `json:"total_tenure_days"`
	AverageAssignmentDays float64                         `json:"average_assignment_days"`
}

// PositionAssignmentSummary contains summarized assignment information
type PositionAssignmentSummary struct {
	AssignmentID    uuid.UUID  `json:"assignment_id"`
	PositionID      uuid.UUID  `json:"position_id"`
	PositionType    string     `json:"position_type"`
	DepartmentID    uuid.UUID  `json:"department_id"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	DurationDays    *int       `json:"duration_days,omitempty"`
	IsActive        bool       `json:"is_active"`
	AssignmentType  string     `json:"assignment_type"`
	FTEPercentage   float64    `json:"fte_percentage"`
	WorkArrangement string     `json:"work_arrangement,omitempty"`
}

// PositionHistoryRecord contains detailed position history information
type PositionHistoryRecord struct {
	Position            *ent.Position                `json:"position"`
	OccupancyHistory    []*EmployeeAssignmentSummary `json:"occupancy_history"`
	TotalOccupants      int                          `json:"total_occupants"`
	CurrentOccupant     *EmployeeAssignmentSummary   `json:"current_occupant,omitempty"`
	AverageOccupancyDays float64                     `json:"average_occupancy_days"`
	VacancyPeriods      []VacancyPeriod              `json:"vacancy_periods"`
}

// EmployeeAssignmentSummary contains summarized employee assignment information
type EmployeeAssignmentSummary struct {
	AssignmentID    uuid.UUID  `json:"assignment_id"`
	EmployeeID      uuid.UUID  `json:"employee_id"`
	EmployeeNumber  string     `json:"employee_number"`
	FullName        string     `json:"full_name"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         *time.Time `json:"end_date,omitempty"`
	DurationDays    *int       `json:"duration_days,omitempty"`
	IsActive        bool       `json:"is_active"`
	AssignmentType  string     `json:"assignment_type"`
	FTEPercentage   float64    `json:"fte_percentage"`
}

// VacancyPeriod represents a period when a position was vacant
type VacancyPeriod struct {
	StartDate    time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	DurationDays *int      `json:"duration_days,omitempty"`
	IsOngoing    bool      `json:"is_ongoing"`
}

// NewAnalyticsService creates a new AnalyticsService
func NewAnalyticsService(client *ent.Client, logger *logging.StructuredLogger) *AnalyticsService {
	return &AnalyticsService{
		client: client,
		logger: logger,
	}
}

// GetOrganizationalMetrics generates comprehensive organizational metrics
func (s *AnalyticsService) GetOrganizationalMetrics(ctx context.Context, tenantID uuid.UUID) (*OrganizationalMetrics, error) {
	s.logger.Info("Generating organizational metrics", "tenant_id", tenantID)

	metrics := &OrganizationalMetrics{
		TenantID:   tenantID,
		ReportDate: time.Now(),
		EmployeesByType:       make(map[string]int),
		EmployeesByStatus:     make(map[string]int),
		PositionsByStatus:     make(map[string]int),
	}

	// Get employee counts and breakdowns
	employees, err := s.client.Employee.Query().
		Where(employee.TenantIDEQ(tenantID)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees: %w", err)
	}

	metrics.TotalEmployees = len(employees)
	for _, emp := range employees {
		// Count by type
		empType := string(emp.EmployeeType)
		metrics.EmployeesByType[empType]++

		// Count by status
		empStatus := string(emp.EmploymentStatus)
		metrics.EmployeesByStatus[empStatus]++

		if emp.EmploymentStatus == employee.EmploymentStatusACTIVE {
			metrics.ActiveEmployees++
		}
	}

	// Get position counts and breakdowns
	positions, err := s.client.Position.Query().
		Where(position.TenantIDEQ(tenantID)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch positions: %w", err)
	}

	metrics.TotalPositions = len(positions)
	for _, pos := range positions {
		posStatus := string(pos.Status)
		metrics.PositionsByStatus[posStatus]++

		if pos.Status == position.StatusFILLED {
			metrics.FilledPositions++
		} else if pos.Status == position.StatusOPEN {
			metrics.OpenPositions++
		}
	}

	// Calculate assignment metrics
	assignmentMetrics, err := s.calculateAssignmentMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate assignment metrics: %w", err)
	}
	metrics.AssignmentMetrics = *assignmentMetrics

	// Calculate turnover metrics
	turnoverMetrics, err := s.calculateTurnoverMetrics(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate turnover metrics: %w", err)
	}
	metrics.TurnoverMetrics = *turnoverMetrics

	// Calculate average assignment duration
	avgDuration, err := s.calculateAverageAssignmentDuration(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average assignment duration: %w", err)
	}
	metrics.AverageAssignmentDuration = avgDuration

	s.logger.Info("Organizational metrics generated successfully",
		"tenant_id", tenantID,
		"total_employees", metrics.TotalEmployees,
		"total_positions", metrics.TotalPositions,
	)

	return metrics, nil
}

// GetEmployeeHistory retrieves detailed history for a specific employee
func (s *AnalyticsService) GetEmployeeHistory(ctx context.Context, tenantID uuid.UUID, employeeID uuid.UUID) (*EmployeeHistoryRecord, error) {
	s.logger.Info("Retrieving employee history",
		"tenant_id", tenantID,
		"employee_id", employeeID,
	)

	// Fetch employee
	emp, err := s.client.Employee.Query().
		Where(
			employee.IDEQ(employeeID),
			employee.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch employee: %w", err)
	}

	// Fetch assignment history
	assignments, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.EmployeeIDEQ(employeeID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
		).
		WithPosition().
		Order(positionoccupancyhistory.ByStartDate()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignment history: %w", err)
	}

	// Convert assignments to summaries
	assignmentSummaries := make([]*PositionAssignmentSummary, len(assignments))
	var currentAssignment *PositionAssignmentSummary
	totalTenureDays := 0
	totalAssignmentDays := 0

	for i, assignment := range assignments {
		summary := &PositionAssignmentSummary{
			AssignmentID:    assignment.ID,
			PositionID:      assignment.PositionID,
			StartDate:       assignment.StartDate,
			EndDate:         assignment.EndDate,
			IsActive:        assignment.IsActive,
			AssignmentType:  string(assignment.AssignmentType),
			FTEPercentage:   assignment.FtePercentage,
			WorkArrangement: string(assignment.WorkArrangement),
		}

		if assignment.Edges.Position != nil {
			summary.PositionType = string(assignment.Edges.Position.PositionType)
			summary.DepartmentID = assignment.Edges.Position.DepartmentID
		}

		// Calculate duration
		if assignment.EndDate != nil {
			days := int(assignment.EndDate.Sub(assignment.StartDate).Hours() / 24)
			summary.DurationDays = &days
			totalAssignmentDays += days
		} else if assignment.IsActive {
			days := int(time.Now().Sub(assignment.StartDate).Hours() / 24)
			summary.DurationDays = &days
			totalAssignmentDays += days
			currentAssignment = summary
		}

		assignmentSummaries[i] = summary
	}

	// Calculate total tenure
	if len(assignments) > 0 {
		startDate := assignments[0].StartDate
		endDate := time.Now()
		if emp.TerminationDate != nil {
			endDate = *emp.TerminationDate
		}
		totalTenureDays = int(endDate.Sub(startDate).Hours() / 24)
	}

	// Calculate average assignment duration
	var averageAssignmentDays float64
	if len(assignments) > 0 {
		averageAssignmentDays = float64(totalAssignmentDays) / float64(len(assignments))
	}

	return &EmployeeHistoryRecord{
		Employee:              emp,
		AssignmentHistory:     assignmentSummaries,
		TotalAssignments:      len(assignments),
		CurrentAssignment:     currentAssignment,
		TotalTenureDays:       totalTenureDays,
		AverageAssignmentDays: averageAssignmentDays,
	}, nil
}

// GetPositionHistory retrieves detailed history for a specific position
func (s *AnalyticsService) GetPositionHistory(ctx context.Context, tenantID uuid.UUID, positionID uuid.UUID) (*PositionHistoryRecord, error) {
	s.logger.Info("Retrieving position history",
		"tenant_id", tenantID,
		"position_id", positionID,
	)

	// Fetch position
	pos, err := s.client.Position.Query().
		Where(
			position.IDEQ(positionID),
			position.TenantIDEQ(tenantID),
		).
		Only(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch position: %w", err)
	}

	// Fetch occupancy history
	occupancies, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.PositionIDEQ(positionID),
			positionoccupancyhistory.TenantIDEQ(tenantID),
		).
		WithEmployee().
		Order(positionoccupancyhistory.ByStartDate()).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch occupancy history: %w", err)
	}

	// Convert occupancies to summaries
	occupancySummaries := make([]*EmployeeAssignmentSummary, len(occupancies))
	var currentOccupant *EmployeeAssignmentSummary
	totalOccupancyDays := 0

	for i, occupancy := range occupancies {
		summary := &EmployeeAssignmentSummary{
			AssignmentID:   occupancy.ID,
			EmployeeID:     occupancy.EmployeeID,
			StartDate:      occupancy.StartDate,
			EndDate:        occupancy.EndDate,
			IsActive:       occupancy.IsActive,
			AssignmentType: string(occupancy.AssignmentType),
			FTEPercentage:  occupancy.FtePercentage,
		}

		if occupancy.Edges.Employee != nil {
			summary.EmployeeNumber = occupancy.Edges.Employee.EmployeeNumber
			summary.FullName = fmt.Sprintf("%s %s",
				occupancy.Edges.Employee.FirstName,
				occupancy.Edges.Employee.LastName)
		}

		// Calculate duration
		if occupancy.EndDate != nil {
			days := int(occupancy.EndDate.Sub(occupancy.StartDate).Hours() / 24)
			summary.DurationDays = &days
			totalOccupancyDays += days
		} else if occupancy.IsActive {
			days := int(time.Now().Sub(occupancy.StartDate).Hours() / 24)
			summary.DurationDays = &days
			totalOccupancyDays += days
			currentOccupant = summary
		}

		occupancySummaries[i] = summary
	}

	// Calculate average occupancy duration
	var averageOccupancyDays float64
	if len(occupancies) > 0 {
		averageOccupancyDays = float64(totalOccupancyDays) / float64(len(occupancies))
	}

	// Calculate vacancy periods
	vacancyPeriods := s.calculateVacancyPeriods(occupancies)

	return &PositionHistoryRecord{
		Position:             pos,
		OccupancyHistory:     occupancySummaries,
		TotalOccupants:       len(occupancies),
		CurrentOccupant:      currentOccupant,
		AverageOccupancyDays: averageOccupancyDays,
		VacancyPeriods:       vacancyPeriods,
	}, nil
}

// GetHistoricalAssignments retrieves assignment history with flexible filtering
func (s *AnalyticsService) GetHistoricalAssignments(ctx context.Context, tenantID uuid.UUID, params HistoryQueryParams) ([]*ent.PositionOccupancyHistory, error) {
	query := s.client.PositionOccupancyHistory.Query().
		Where(positionoccupancyhistory.TenantIDEQ(tenantID)).
		WithEmployee().
		WithPosition()

	// Apply filters
	if params.StartDate != nil {
		query = query.Where(positionoccupancyhistory.StartDateGTE(*params.StartDate))
	}
	if params.EndDate != nil {
		query = query.Where(positionoccupancyhistory.StartDateLTE(*params.EndDate))
	}
	if params.EmployeeID != nil {
		query = query.Where(positionoccupancyhistory.EmployeeIDEQ(*params.EmployeeID))
	}
	if params.PositionID != nil {
		query = query.Where(positionoccupancyhistory.PositionIDEQ(*params.PositionID))
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(100) // Default limit
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Order by start date
	query = query.Order(positionoccupancyhistory.ByStartDate())

	assignments, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical assignments: %w", err)
	}

	return assignments, nil
}

// Private helper methods

func (s *AnalyticsService) calculateAssignmentMetrics(ctx context.Context, tenantID uuid.UUID) (*AssignmentMetrics, error) {
	// Get all assignments
	assignments, err := s.client.PositionOccupancyHistory.Query().
		Where(positionoccupancyhistory.TenantIDEQ(tenantID)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch assignments: %w", err)
	}

	metrics := &AssignmentMetrics{
		AssignmentsByType: make(map[string]int),
		AssignmentTrends:  []AssignmentTrendPoint{},
	}

	metrics.TotalAssignments = len(assignments)
	totalDays := 0
	completedAssignments := 0

	currentYear := time.Now().Year()

	for _, assignment := range assignments {
		// Count by type
		assignmentType := string(assignment.AssignmentType)
		metrics.AssignmentsByType[assignmentType]++

		// Count active assignments
		if assignment.IsActive {
			metrics.ActiveAssignments++
		}

		// Calculate duration for completed assignments
		if assignment.EndDate != nil {
			days := int(assignment.EndDate.Sub(assignment.StartDate).Hours() / 24)
			totalDays += days
			completedAssignments++
		}

		// Count promotions and transfers this year
		if assignment.StartDate.Year() == currentYear {
			if assignment.AssignmentReason != "" {
				if contains(assignment.AssignmentReason, "promotion") {
					metrics.PromotionsThisYear++
				} else if contains(assignment.AssignmentReason, "transfer") {
					metrics.TransfersThisYear++
				}
			}
		}
	}

	// Calculate average assignment length
	if completedAssignments > 0 {
		metrics.AverageAssignmentLength = float64(totalDays) / float64(completedAssignments)
	}

	return metrics, nil
}

func (s *AnalyticsService) calculateTurnoverMetrics(ctx context.Context, tenantID uuid.UUID) (*TurnoverMetrics, error) {
	now := time.Now()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	currentQuarter := time.Date(now.Year(), ((now.Month()-1)/3)*3+1, 1, 0, 0, 0, 0, now.Location())
	currentYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	metrics := &TurnoverMetrics{}

	// Get terminated employees
	terminatedEmployees, err := s.client.Employee.Query().
		Where(
			employee.TenantIDEQ(tenantID),
			employee.EmploymentStatusEQ(employee.EmploymentStatusTERMINATED),
			employee.TerminationDateNotNil(),
		).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch terminated employees: %w", err)
	}

	for _, emp := range terminatedEmployees {
		if emp.TerminationDate != nil {
			termDate := *emp.TerminationDate
			if termDate.After(currentMonth) {
				metrics.TerminationsThisMonth++
			}
			if termDate.After(currentQuarter) {
				metrics.TerminationsThisQuarter++
			}
			if termDate.After(currentYear) {
				metrics.TerminationsThisYear++
			}
		}
	}

	// Get hired employees
	allEmployees, err := s.client.Employee.Query().
		Where(employee.TenantIDEQ(tenantID)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch all employees: %w", err)
	}

	totalEmployees := len(allEmployees)
	for _, emp := range allEmployees {
		if emp.HireDate.After(currentMonth) {
			metrics.HiresThisMonth++
		}
		if emp.HireDate.After(currentQuarter) {
			metrics.HiresThisQuarter++
		}
		if emp.HireDate.After(currentYear) {
			metrics.HiresThisYear++
		}
	}

	// Calculate turnover rates
	if totalEmployees > 0 {
		metrics.MonthlyTurnoverRate = float64(metrics.TerminationsThisMonth) / float64(totalEmployees) * 100
		metrics.QuarterlyTurnoverRate = float64(metrics.TerminationsThisQuarter) / float64(totalEmployees) * 100
		metrics.AnnualTurnoverRate = float64(metrics.TerminationsThisYear) / float64(totalEmployees) * 100
	}

	return metrics, nil
}

func (s *AnalyticsService) calculateAverageAssignmentDuration(ctx context.Context, tenantID uuid.UUID) (float64, error) {
	assignments, err := s.client.PositionOccupancyHistory.Query().
		Where(
			positionoccupancyhistory.TenantIDEQ(tenantID),
			positionoccupancyhistory.EndDateNotNil(),
		).
		All(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch completed assignments: %w", err)
	}

	if len(assignments) == 0 {
		return 0, nil
	}

	totalDays := 0
	for _, assignment := range assignments {
		if assignment.EndDate != nil {
			days := int(assignment.EndDate.Sub(assignment.StartDate).Hours() / 24)
			totalDays += days
		}
	}

	return float64(totalDays) / float64(len(assignments)), nil
}

func (s *AnalyticsService) calculateVacancyPeriods(occupancies []*ent.PositionOccupancyHistory) []VacancyPeriod {
	if len(occupancies) == 0 {
		return []VacancyPeriod{{
			StartDate: time.Now(), // Position has been vacant since creation
			IsOngoing: true,
		}}
	}

	var vacancyPeriods []VacancyPeriod

	// Sort occupancies by start date (should already be sorted)
	for i := 0; i < len(occupancies)-1; i++ {
		currentEnd := occupancies[i].EndDate
		nextStart := occupancies[i+1].StartDate

		if currentEnd != nil && nextStart.After(*currentEnd) {
			// There's a gap between assignments
			days := int(nextStart.Sub(*currentEnd).Hours() / 24)
			vacancyPeriods = append(vacancyPeriods, VacancyPeriod{
				StartDate:    *currentEnd,
				EndDate:      &nextStart,
				DurationDays: &days,
				IsOngoing:    false,
			})
		}
	}

	// Check if position is currently vacant
	lastOccupancy := occupancies[len(occupancies)-1]
	if !lastOccupancy.IsActive && lastOccupancy.EndDate != nil {
		days := int(time.Now().Sub(*lastOccupancy.EndDate).Hours() / 24)
		vacancyPeriods = append(vacancyPeriods, VacancyPeriod{
			StartDate:    *lastOccupancy.EndDate,
			DurationDays: &days,
			IsOngoing:    true,
		})
	}

	return vacancyPeriods
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}