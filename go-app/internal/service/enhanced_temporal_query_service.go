// internal/service/enhanced_temporal_query_service.go
package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/ent"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// EnhancedTemporalQueryService extends TemporalQueryService with advanced temporal operations
type EnhancedTemporalQueryService struct {
	*TemporalQueryService
	metricsCollector *TemporalMetricsCollector
}

// NewEnhancedTemporalQueryService creates an enhanced temporal query service
func NewEnhancedTemporalQueryService(
	client *ent.Client, 
	logger *logging.StructuredLogger,
) *EnhancedTemporalQueryService {
	baseService := NewTemporalQueryService(client, logger)
	return &EnhancedTemporalQueryService{
		TemporalQueryService: baseService,
		metricsCollector:     NewTemporalMetricsCollector(),
	}
}

// PositionTimelineQuery represents a complex timeline query request
type PositionTimelineQuery struct {
	TenantID     uuid.UUID  `json:"tenant_id"`
	EmployeeIDs  []uuid.UUID `json:"employee_ids,omitempty"`
	Departments  []string   `json:"departments,omitempty"`
	JobLevels    []string   `json:"job_levels,omitempty"`
	DateRange    *DateRange `json:"date_range,omitempty"`
	IncludeRetroactive bool `json:"include_retroactive"`
	OrderBy      string     `json:"order_by"` // "effective_date", "employee_id", "department"
	Limit        *int       `json:"limit,omitempty"`
	Offset       *int       `json:"offset,omitempty"`
}

// DateRange represents a time range for queries
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// PositionTimelineResult represents the result of a timeline query
type PositionTimelineResult struct {
	Positions    []*PositionSnapshot    `json:"positions"`
	TotalCount   int                    `json:"total_count"`
	QueryMetrics *QueryExecutionMetrics `json:"query_metrics"`
	HasMore      bool                   `json:"has_more"`
}

// QueryExecutionMetrics tracks query performance
type QueryExecutionMetrics struct {
	ExecutionTime    time.Duration `json:"execution_time"`
	RecordsScanned   int          `json:"records_scanned"`
	RecordsReturned  int          `json:"records_returned"`
	IndexesUsed      []string     `json:"indexes_used"`
	CacheHit         bool         `json:"cache_hit"`
}

// GetAdvancedPositionTimeline executes complex timeline queries with performance tracking
func (s *EnhancedTemporalQueryService) GetAdvancedPositionTimeline(
	ctx context.Context,
	query PositionTimelineQuery,
) (*PositionTimelineResult, error) {
	startTime := time.Now()
	
	s.logger.LogInfo(ctx, "Executing advanced position timeline query", map[string]interface{}{
		"tenant_id":      query.TenantID,
		"employee_count": len(query.EmployeeIDs),
		"departments":    query.Departments,
		"job_levels":     query.JobLevels,
		"date_range":     query.DateRange,
	})

	// Build dynamic query
	dbQuery := s.client.PositionHistory.Query().
		Where(position_history.TenantIDEQ(query.TenantID))

	// Add employee filter
	if len(query.EmployeeIDs) > 0 {
		dbQuery = dbQuery.Where(position_history.EmployeeIDIn(query.EmployeeIDs...))
	}

	// Add department filter
	if len(query.Departments) > 0 {
		dbQuery = dbQuery.Where(position_history.DepartmentIn(query.Departments...))
	}

	// Add job level filter
	if len(query.JobLevels) > 0 {
		dbQuery = dbQuery.Where(position_history.JobLevelIn(query.JobLevels...))
	}

	// Add date range filter
	if query.DateRange != nil {
		dbQuery = dbQuery.Where(
			position_history.EffectiveDateLTE(query.DateRange.EndDate),
			position_history.Or(
				position_history.EndDateIsNil(),
				position_history.EndDateGTE(query.DateRange.StartDate),
			),
		)
	}

	// Add retroactive filter
	if !query.IncludeRetroactive {
		dbQuery = dbQuery.Where(position_history.IsRetroactiveEQ(false))
	}

	// Add ordering
	switch query.OrderBy {
	case "employee_id":
		dbQuery = dbQuery.Order(ent.Asc(position_history.FieldEmployeeID), ent.Asc(position_history.FieldEffectiveDate))
	case "department":
		dbQuery = dbQuery.Order(ent.Asc(position_history.FieldDepartment), ent.Asc(position_history.FieldEffectiveDate))
	default:
		dbQuery = dbQuery.Order(ent.Asc(position_history.FieldEffectiveDate))
	}

	// Get total count for pagination
	totalCount, err := dbQuery.Count(ctx)
	if err != nil {
		s.logger.LogError(ctx, "Failed to get timeline query count", err, map[string]interface{}{
			"tenant_id": query.TenantID,
		})
		return nil, err
	}

	// Apply pagination
	if query.Offset != nil {
		dbQuery = dbQuery.Offset(*query.Offset)
	}
	if query.Limit != nil {
		dbQuery = dbQuery.Limit(*query.Limit)
	}

	// Execute query
	positions, err := dbQuery.All(ctx)
	if err != nil {
		s.logger.LogError(ctx, "Failed to execute timeline query", err, map[string]interface{}{
			"tenant_id": query.TenantID,
		})
		return nil, err
	}

	// Convert to snapshots
	snapshots := make([]*PositionSnapshot, len(positions))
	for i, pos := range positions {
		snapshots[i] = s.convertToSnapshot(pos)
	}

	// Calculate metrics
	executionTime := time.Since(startTime)
	metrics := &QueryExecutionMetrics{
		ExecutionTime:   executionTime,
		RecordsScanned:  totalCount,
		RecordsReturned: len(snapshots),
		IndexesUsed:     s.identifyUsedIndexes(query),
		CacheHit:        false, // TODO: Implement caching
	}

	// Update metrics collector
	s.metricsCollector.RecordQuery(executionTime, len(snapshots), totalCount)

	result := &PositionTimelineResult{
		Positions:    snapshots,
		TotalCount:   totalCount,
		QueryMetrics: metrics,
		HasMore:      query.Limit != nil && len(snapshots) == *query.Limit,
	}

	s.logger.LogInfo(ctx, "Advanced timeline query completed", map[string]interface{}{
		"tenant_id":        query.TenantID,
		"execution_time":   executionTime.Milliseconds(),
		"records_returned": len(snapshots),
		"total_count":      totalCount,
	})

	return result, nil
}

// GetPositionChangesInPeriod returns all position changes within a specific period
func (s *EnhancedTemporalQueryService) GetPositionChangesInPeriod(
	ctx context.Context,
	tenantID uuid.UUID,
	startDate, endDate time.Time,
	includeRetroactive bool,
) ([]*PositionChangeEvent, error) {
	s.logger.LogInfo(ctx, "Querying position changes in period", map[string]interface{}{
		"tenant_id":           tenantID,
		"start_date":          startDate,
		"end_date":            endDate,
		"include_retroactive": includeRetroactive,
	})

	query := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EffectiveDateGTE(startDate),
			position_history.EffectiveDateLTE(endDate),
		)

	if !includeRetroactive {
		query = query.Where(position_history.IsRetroactiveEQ(false))
	}

	positions, err := query.
		Order(ent.Asc(position_history.FieldEffectiveDate)).
		All(ctx)

	if err != nil {
		s.logger.LogError(ctx, "Failed to query position changes", err, map[string]interface{}{
			"tenant_id":  tenantID,
			"start_date": startDate,
			"end_date":   endDate,
		})
		return nil, err
	}

	// Convert to change events
	events := make([]*PositionChangeEvent, len(positions))
	for i, pos := range positions {
		events[i] = &PositionChangeEvent{
			PositionHistoryID: pos.ID,
			EmployeeID:       pos.EmployeeID,
			ChangeType:       s.determineChangeType(ctx, pos),
			EffectiveDate:    pos.EffectiveDate,
			PreviousPosition: s.getPreviousPosition(ctx, tenantID, pos.EmployeeID, pos.EffectiveDate),
			NewPosition:      s.convertToSnapshot(pos),
			IsRetroactive:    pos.IsRetroactive,
			ChangeReason:     pos.ChangeReason,
		}
	}

	s.logger.LogInfo(ctx, "Position changes query completed", map[string]interface{}{
		"tenant_id":    tenantID,
		"change_count": len(events),
	})

	return events, nil
}

// GetPositionGapsAndOverlaps identifies temporal inconsistencies
func (s *EnhancedTemporalQueryService) GetPositionGapsAndOverlaps(
	ctx context.Context,
	tenantID uuid.UUID,
	employeeIDs []uuid.UUID,
) (*TemporalConsistencyReport, error) {
	s.logger.LogInfo(ctx, "Analyzing position gaps and overlaps", map[string]interface{}{
		"tenant_id":      tenantID,
		"employee_count": len(employeeIDs),
	})

	report := &TemporalConsistencyReport{
		TenantID:    tenantID,
		EmployeeIDs: employeeIDs,
		Gaps:        []PositionGap{},
		Overlaps:    []PositionOverlap{},
		Warnings:    []string{},
	}

	for _, employeeID := range employeeIDs {
		// Get all positions for employee, ordered by effective date
		positions, err := s.client.PositionHistory.Query().
			Where(
				position_history.TenantIDEQ(tenantID),
				position_history.EmployeeIDEQ(employeeID),
			).
			Order(ent.Asc(position_history.FieldEffectiveDate)).
			All(ctx)

		if err != nil {
			s.logger.LogError(ctx, "Failed to get positions for gap analysis", err, map[string]interface{}{
				"tenant_id":   tenantID,
				"employee_id": employeeID,
			})
			continue
		}

		// Analyze for gaps and overlaps
		s.analyzePositionContinuity(positions, employeeID, report)
	}

	s.logger.LogInfo(ctx, "Gap and overlap analysis completed", map[string]interface{}{
		"tenant_id":  tenantID,
		"gaps":       len(report.Gaps),
		"overlaps":   len(report.Overlaps),
		"warnings":   len(report.Warnings),
	})

	return report, nil
}

// BatchCreatePositionSnapshots creates multiple position snapshots in a single transaction
func (s *EnhancedTemporalQueryService) BatchCreatePositionSnapshots(
	ctx context.Context,
	tenantID uuid.UUID,
	snapshots []BatchPositionSnapshotData,
) (*BatchCreateResult, error) {
	s.logger.LogInfo(ctx, "Creating batch position snapshots", map[string]interface{}{
		"tenant_id":     tenantID,
		"snapshot_count": len(snapshots),
	})

	result := &BatchCreateResult{
		TotalRequested: len(snapshots),
		Successful:     []uuid.UUID{},
		Failed:         []BatchError{},
	}

	// Use transaction for batch operations
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	for i, snapshotData := range snapshots {
		// Validate temporal consistency for each snapshot
		if err := s.ValidateTemporalConsistency(ctx, tenantID, snapshotData.EmployeeID, snapshotData.EffectiveDate); err != nil {
			result.Failed = append(result.Failed, BatchError{
				Index:       i,
				EmployeeID:  snapshotData.EmployeeID,
				Error:       err.Error(),
				ErrorType:   "TEMPORAL_CONFLICT",
			})
			continue
		}

		// Create position snapshot
		position, err := s.createPositionSnapshotInTx(ctx, tx, tenantID, snapshotData)
		if err != nil {
			result.Failed = append(result.Failed, BatchError{
				Index:       i,
				EmployeeID:  snapshotData.EmployeeID,
				Error:       err.Error(),
				ErrorType:   "CREATE_FAILED",
			})
			continue
		}

		result.Successful = append(result.Successful, position.ID)
	}

	// Commit transaction only if we have at least one success
	if len(result.Successful) > 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit batch transaction: %w", err)
		}
	}

	result.SuccessCount = len(result.Successful)
	result.FailureCount = len(result.Failed)

	s.logger.LogInfo(ctx, "Batch position snapshots created", map[string]interface{}{
		"tenant_id":     tenantID,
		"success_count": result.SuccessCount,
		"failure_count": result.FailureCount,
	})

	return result, nil
}

// Helper methods

func (s *EnhancedTemporalQueryService) convertToSnapshot(pos *ent.PositionHistory) *PositionSnapshot {
	snapshot := &PositionSnapshot{
		PositionHistoryID:    pos.ID,
		EmployeeID:          pos.EmployeeID,
		PositionTitle:       pos.PositionTitle,
		Department:          pos.Department,
		EmploymentType:      pos.EmploymentType,
		EffectiveDate:       pos.EffectiveDate,
		EndDate:             pos.EndDate,
		IsRetroactive:       pos.IsRetroactive,
	}

	// Handle optional fields
	if pos.JobLevel != "" {
		snapshot.JobLevel = &pos.JobLevel
	}
	if pos.Location != "" {
		snapshot.Location = &pos.Location
	}
	if pos.ReportsToEmployeeID != uuid.Nil {
		snapshot.ReportsToEmployeeID = &pos.ReportsToEmployeeID
	}
	if pos.ChangeReason != "" {
		snapshot.ChangeReason = &pos.ChangeReason
	}
	if pos.MinSalary != 0 {
		snapshot.MinSalary = &pos.MinSalary
	}
	if pos.MaxSalary != 0 {
		snapshot.MaxSalary = &pos.MaxSalary
	}
	if pos.Currency != "" {
		snapshot.Currency = &pos.Currency
	}

	return snapshot
}

func (s *EnhancedTemporalQueryService) identifyUsedIndexes(query PositionTimelineQuery) []string {
	// This is a simplified version - in production, you'd query EXPLAIN plans
	indexes := []string{"idx_position_history_temporal"}
	
	if len(query.EmployeeIDs) > 0 {
		indexes = append(indexes, "idx_position_history_employee")
	}
	if query.DateRange != nil {
		indexes = append(indexes, "idx_position_history_date_range")
	}
	
	return indexes
}

func (s *EnhancedTemporalQueryService) determineChangeType(ctx context.Context, pos *ent.PositionHistory) string {
	// Get previous position to determine change type
	prevPos := s.getPreviousPosition(ctx, pos.TenantID, pos.EmployeeID, pos.EffectiveDate)
	
	if prevPos == nil {
		return "INITIAL_HIRE"
	}
	
	if prevPos.PositionTitle != pos.PositionTitle {
		return "PROMOTION"
	}
	if prevPos.Department != pos.Department {
		return "TRANSFER"
	}
	if prevPos.ReportsToEmployeeID != nil && pos.ReportsToEmployeeID != uuid.Nil && 
		*prevPos.ReportsToEmployeeID != pos.ReportsToEmployeeID {
		return "MANAGER_CHANGE"
	}
	
	return "INFORMATION_UPDATE"
}

func (s *EnhancedTemporalQueryService) getPreviousPosition(ctx context.Context, tenantID, employeeID uuid.UUID, effectiveDate time.Time) *PositionSnapshot {
	pos, err := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EmployeeIDEQ(employeeID),
			position_history.EffectiveDateLT(effectiveDate),
		).
		Order(ent.Desc(position_history.FieldEffectiveDate)).
		First(ctx)
	
	if err != nil {
		return nil
	}
	
	return s.convertToSnapshot(pos)
}

func (s *EnhancedTemporalQueryService) analyzePositionContinuity(positions []*ent.PositionHistory, employeeID uuid.UUID, report *TemporalConsistencyReport) {
	for i := 0; i < len(positions); i++ {
		current := positions[i]
		
		// Check for gaps (previous position ended before current starts)
		if i > 0 {
			previous := positions[i-1]
			if previous.EndDate != nil && previous.EndDate.Before(current.EffectiveDate.Add(-24*time.Hour)) {
				report.Gaps = append(report.Gaps, PositionGap{
					EmployeeID:      employeeID,
					GapStart:        previous.EndDate.Add(24 * time.Hour),
					GapEnd:          current.EffectiveDate.Add(-24 * time.Hour),
					GapDuration:     current.EffectiveDate.Sub(*previous.EndDate),
					PreviousPosition: previous.ID,
					NextPosition:    current.ID,
				})
			}
		}
		
		// Check for overlaps (current position starts before previous ends)
		if i > 0 {
			previous := positions[i-1]
			if previous.EndDate == nil || previous.EndDate.After(current.EffectiveDate) {
				overlapEnd := current.EffectiveDate
				if previous.EndDate != nil && previous.EndDate.Before(overlapEnd) {
					overlapEnd = *previous.EndDate
				}
				
				report.Overlaps = append(report.Overlaps, PositionOverlap{
					EmployeeID:      employeeID,
					OverlapStart:    current.EffectiveDate,
					OverlapEnd:      overlapEnd,
					OverlapDuration: overlapEnd.Sub(current.EffectiveDate),
					Position1:       previous.ID,
					Position2:       current.ID,
				})
			}
		}
	}
}

func (s *EnhancedTemporalQueryService) createPositionSnapshotInTx(ctx context.Context, tx *ent.Tx, tenantID uuid.UUID, data BatchPositionSnapshotData) (*ent.PositionHistory, error) {
	createBuilder := tx.PositionHistory.Create().
		SetTenantID(tenantID).
		SetEmployeeID(data.EmployeeID).
		SetPositionTitle(data.PositionTitle).
		SetDepartment(data.Department).
		SetEmploymentType(data.EmploymentType).
		SetEffectiveDate(data.EffectiveDate).
		SetIsRetroactive(data.IsRetroactive).
		SetCreatedBy(data.CreatedBy).
		SetCreatedAt(time.Now())

	// Set optional fields
	if data.JobLevel != nil {
		createBuilder = createBuilder.SetJobLevel(*data.JobLevel)
	}
	if data.Location != nil {
		createBuilder = createBuilder.SetLocation(*data.Location)
	}
	if data.ReportsToEmployeeID != nil {
		createBuilder = createBuilder.SetReportsToEmployeeID(*data.ReportsToEmployeeID)
	}
	if data.EndDate != nil {
		createBuilder = createBuilder.SetEndDate(*data.EndDate)
	}
	if data.ChangeReason != nil {
		createBuilder = createBuilder.SetChangeReason(*data.ChangeReason)
	}
	if data.MinSalary != nil {
		createBuilder = createBuilder.SetMinSalary(*data.MinSalary)
	}
	if data.MaxSalary != nil {
		createBuilder = createBuilder.SetMaxSalary(*data.MaxSalary)
	}
	if data.Currency != nil {
		createBuilder = createBuilder.SetCurrency(*data.Currency)
	}

	return createBuilder.Save(ctx)
}