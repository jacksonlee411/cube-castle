// internal/service/enhanced_temporal_query_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	entgo "github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/google/uuid"
)

// EnhancedTemporalQueryService extends TemporalQueryService with advanced temporal operations
type EnhancedTemporalQueryService struct {
	client           *entgo.Client
	logger           *logging.StructuredLogger
	metricsCollector *TemporalMetricsCollector
}

// NewEnhancedTemporalQueryService creates an enhanced temporal query service
func NewEnhancedTemporalQueryService(
	client *entgo.Client,
	logger *logging.StructuredLogger,
) *EnhancedTemporalQueryService {
	return &EnhancedTemporalQueryService{
		client:           client,
		logger:           logger,
		metricsCollector: NewTemporalMetricsCollector(),
	}
}

// PositionTimelineQuery represents a complex timeline query request
type PositionTimelineQuery struct {
	TenantID           uuid.UUID   `json:"tenant_id"`
	EmployeeIDs        []uuid.UUID `json:"employee_ids,omitempty"`
	Departments        []string    `json:"departments,omitempty"`
	JobLevels          []string    `json:"job_levels,omitempty"`
	DateRange          *DateRange  `json:"date_range,omitempty"`
	IncludeRetroactive bool        `json:"include_retroactive"`
	OrderBy            string      `json:"order_by"` // "effective_date", "employee_id", "department"
	Limit              *int        `json:"limit,omitempty"`
	Offset             *int        `json:"offset,omitempty"`
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
	ExecutionTime   time.Duration `json:"execution_time"`
	RecordsScanned  int           `json:"records_scanned"`
	RecordsReturned int           `json:"records_returned"`
	IndexesUsed     []string      `json:"indexes_used"`
	CacheHit        bool          `json:"cache_hit"`
}

// GetAdvancedPositionTimeline executes complex timeline queries with performance tracking
func (s *EnhancedTemporalQueryService) GetAdvancedPositionTimeline(
	ctx context.Context,
	query PositionTimelineQuery,
) (*PositionTimelineResult, error) {
	startTime := time.Now()

	s.logger.Info("Executing advanced position timeline query",
		"tenant_id", query.TenantID,
		"employee_count", len(query.EmployeeIDs),
		"departments", query.Departments,
		"job_levels", query.JobLevels,
		"date_range", query.DateRange,
	)

	// Build dynamic query - using OrganizationID as a proxy for TenantID
	dbQuery := s.client.PositionHistory.Query().
		Where(positionhistory.OrganizationIDEQ(query.TenantID.String()))

	// Add employee filter - convert UUIDs to strings
	if len(query.EmployeeIDs) > 0 {
		employeeIDStrings := make([]string, len(query.EmployeeIDs))
		for i, id := range query.EmployeeIDs {
			employeeIDStrings[i] = id.String()
		}
		dbQuery = dbQuery.Where(positionhistory.EmployeeIDIn(employeeIDStrings...))
	}

	// Add department filter
	if len(query.Departments) > 0 {
		dbQuery = dbQuery.Where(positionhistory.DepartmentIn(query.Departments...))
	}

	// Skip job level filter as it's not in the current schema
	// TODO: Add job_level field to position_history schema if needed
	if len(query.JobLevels) > 0 {
		s.logger.Info("Job level filter skipped - field not in position_history schema",
			"job_levels", query.JobLevels,
			"note", "Add job_level field to schema if filtering by job level is required",
		)
	}

	// Add date range filter
	if query.DateRange != nil {
		dbQuery = dbQuery.Where(
			positionhistory.EffectiveDateLTE(query.DateRange.EndDate),
			positionhistory.Or(
				positionhistory.EndDateIsNil(),
				positionhistory.EndDateGTE(query.DateRange.StartDate),
			),
		)
	}

	// Add retroactive filter
	if !query.IncludeRetroactive {
		dbQuery = dbQuery.Where(positionhistory.IsRetroactiveEQ(false))
	}

	// Add ordering
	switch query.OrderBy {
	case "employee_id":
		dbQuery = dbQuery.Order(positionhistory.ByEmployeeID(), positionhistory.ByEffectiveDate())
	case "department":
		dbQuery = dbQuery.Order(positionhistory.ByDepartment(), positionhistory.ByEffectiveDate())
	default:
		dbQuery = dbQuery.Order(positionhistory.ByEffectiveDate())
	}

	// Get total count for pagination
	totalCount, err := dbQuery.Count(ctx)
	if err != nil {
		s.logger.LogError("timeline_query_count_failed", "Failed to get timeline query count", err, map[string]interface{}{
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
		s.logger.LogError("timeline_query_execution_failed", "Failed to execute timeline query", err, map[string]interface{}{
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

	s.logger.Info("Advanced timeline query completed",
		"tenant_id", query.TenantID,
		"execution_time", executionTime.Milliseconds(),
		"records_returned", len(snapshots),
		"total_count", totalCount,
	)

	return result, nil
}

// GetPositionChangesInPeriod returns all position changes within a specific period
func (s *EnhancedTemporalQueryService) GetPositionChangesInPeriod(
	ctx context.Context,
	tenantID uuid.UUID,
	startDate, endDate time.Time,
	includeRetroactive bool,
) ([]*PositionChangeEvent, error) {
	s.logger.Info("Querying position changes in period",
		"tenant_id", tenantID,
		"start_date", startDate,
		"end_date", endDate,
		"include_retroactive", includeRetroactive,
	)

	query := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EffectiveDateGTE(startDate),
			positionhistory.EffectiveDateLTE(endDate),
		)

	if !includeRetroactive {
		query = query.Where(positionhistory.IsRetroactiveEQ(false))
	}

	positions, err := query.
		Order(positionhistory.ByEffectiveDate()).
		All(ctx)

	if err != nil {
		s.logger.LogError("position_changes_query_failed", "Failed to query position changes", err, map[string]interface{}{
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
			PositionHistoryID: uuid.MustParse(pos.ID),
			EmployeeID:        uuid.MustParse(pos.EmployeeID),
			ChangeType:        s.determineChangeType(ctx, pos),
			EffectiveDate:     pos.EffectiveDate,
			PreviousPosition:  s.getPreviousPosition(ctx, tenantID, uuid.MustParse(pos.EmployeeID), pos.EffectiveDate),
			NewPosition:       s.convertToSnapshot(pos),
			IsRetroactive:     pos.IsRetroactive,
			ChangeReason:      getStringValue(pos.ChangeReason),
		}
	}

	s.logger.Info("Position changes query completed",
		"tenant_id", tenantID,
		"change_count", len(events),
	)

	return events, nil
}

// GetPositionGapsAndOverlaps identifies temporal inconsistencies
func (s *EnhancedTemporalQueryService) GetPositionGapsAndOverlaps(
	ctx context.Context,
	tenantID uuid.UUID,
	employeeIDs []uuid.UUID,
) (*TemporalConsistencyReport, error) {
	s.logger.Info("Analyzing position gaps and overlaps",
		"tenant_id", tenantID,
		"employee_count", len(employeeIDs),
	)

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
				positionhistory.OrganizationIDEQ(tenantID.String()),
				positionhistory.EmployeeIDEQ(employeeID.String()),
			).
			Order(positionhistory.ByEffectiveDate()).
			All(ctx)

		if err != nil {
			s.logger.LogError("gap_analysis_query_failed", "Failed to get positions for gap analysis", err, map[string]interface{}{
				"tenant_id":   tenantID,
				"employee_id": employeeID,
			})
			continue
		}

		// Analyze for gaps and overlaps
		s.analyzePositionContinuity(positions, employeeID, report)
	}

	s.logger.Info("Gap and overlap analysis completed",
		"tenant_id", tenantID,
		"gaps", len(report.Gaps),
		"overlaps", len(report.Overlaps),
		"warnings", len(report.Warnings),
	)

	return report, nil
}

// BatchCreatePositionSnapshots creates multiple position snapshots in a single transaction
func (s *EnhancedTemporalQueryService) BatchCreatePositionSnapshots(
	ctx context.Context,
	tenantID uuid.UUID,
	snapshots []BatchPositionSnapshotData,
) (*BatchCreateResult, error) {
	s.logger.Info("Creating batch position snapshots",
		"tenant_id", tenantID,
		"snapshot_count", len(snapshots),
	)

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
				Index:      i,
				EmployeeID: snapshotData.EmployeeID,
				Error:      err.Error(),
				ErrorType:  "TEMPORAL_CONFLICT",
			})
			continue
		}

		// Create position snapshot
		position, err := s.createPositionSnapshotInTx(ctx, tx, tenantID, snapshotData)
		if err != nil {
			result.Failed = append(result.Failed, BatchError{
				Index:      i,
				EmployeeID: snapshotData.EmployeeID,
				Error:      err.Error(),
				ErrorType:  "CREATE_FAILED",
			})
			continue
		}

		result.Successful = append(result.Successful, uuid.MustParse(position.ID))
	}

	// Commit transaction only if we have at least one success
	if len(result.Successful) > 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit batch transaction: %w", err)
		}
	}

	result.SuccessCount = len(result.Successful)
	result.FailureCount = len(result.Failed)

	s.logger.Info("Batch position snapshots created",
		"tenant_id", tenantID,
		"success_count", result.SuccessCount,
		"failure_count", result.FailureCount,
	)

	return result, nil
}

// Helper methods

func (s *EnhancedTemporalQueryService) convertToSnapshot(pos *entgo.PositionHistory) *PositionSnapshot {
	snapshot := &PositionSnapshot{
		PositionHistoryID: uuid.MustParse(pos.ID),
		EmployeeID:        uuid.MustParse(pos.EmployeeID),
		PositionTitle:     pos.PositionTitle,
		Department:        pos.Department,
		EffectiveDate:     pos.EffectiveDate,
		EndDate:           pos.EndDate,
		IsRetroactive:     pos.IsRetroactive,
	}

	// Handle optional fields that exist in schema
	if pos.ChangeReason != nil && *pos.ChangeReason != "" {
		snapshot.ChangeReason = pos.ChangeReason
	}

	return snapshot
}

func (s *EnhancedTemporalQueryService) identifyUsedIndexes(query PositionTimelineQuery) []string {
	// This is a simplified version - in production, you'd query EXPLAIN plans
	indexes := []string{"idx_positionhistory_temporal"}

	if len(query.EmployeeIDs) > 0 {
		indexes = append(indexes, "idx_positionhistory_employee")
	}
	if query.DateRange != nil {
		indexes = append(indexes, "idx_positionhistory_date_range")
	}

	return indexes
}

func (s *EnhancedTemporalQueryService) determineChangeType(ctx context.Context, pos *entgo.PositionHistory) string {
	// Get previous position to determine change type
	employeeUUID, err := uuid.Parse(pos.EmployeeID)
	if err != nil {
		return "INITIAL_HIRE"
	}
	organizationUUID, err := uuid.Parse(pos.OrganizationID)
	if err != nil {
		return "INITIAL_HIRE"
	}

	prevPos := s.getPreviousPosition(ctx, organizationUUID, employeeUUID, pos.EffectiveDate)

	if prevPos == nil {
		return "INITIAL_HIRE"
	}

	if prevPos.PositionTitle != pos.PositionTitle {
		return "PROMOTION"
	}
	if prevPos.Department != pos.Department {
		return "TRANSFER"
	}

	return "INFORMATION_UPDATE"
}

func (s *EnhancedTemporalQueryService) getPreviousPosition(ctx context.Context, tenantID, employeeID uuid.UUID, effectiveDate time.Time) *PositionSnapshot {
	pos, err := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
			positionhistory.EffectiveDateLT(effectiveDate),
		).
		Order(positionhistory.ByEffectiveDate(sql.OrderDesc())).
		First(ctx)

	if err != nil {
		return nil
	}

	return s.convertToSnapshot(pos)
}

func (s *EnhancedTemporalQueryService) analyzePositionContinuity(positions []*entgo.PositionHistory, employeeID uuid.UUID, report *TemporalConsistencyReport) {
	for i := 0; i < len(positions); i++ {
		current := positions[i]

		// Check for gaps (previous position ended before current starts)
		if i > 0 {
			previous := positions[i-1]
			if previous.EndDate != nil && previous.EndDate.Before(current.EffectiveDate.Add(-24*time.Hour)) {
				report.Gaps = append(report.Gaps, PositionGap{
					EmployeeID:       employeeID,
					GapStart:         previous.EndDate.Add(24 * time.Hour),
					GapEnd:           current.EffectiveDate.Add(-24 * time.Hour),
					GapDuration:      current.EffectiveDate.Sub(*previous.EndDate),
					PreviousPosition: uuid.MustParse(previous.ID),
					NextPosition:     uuid.MustParse(current.ID),
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
					Position1:       uuid.MustParse(previous.ID),
					Position2:       uuid.MustParse(current.ID),
				})
			}
		}
	}
}

func (s *EnhancedTemporalQueryService) createPositionSnapshotInTx(ctx context.Context, tx *entgo.Tx, tenantID uuid.UUID, data BatchPositionSnapshotData) (*entgo.PositionHistory, error) {
	createBuilder := tx.PositionHistory.Create().
		SetOrganizationID(tenantID.String()).
		SetEmployeeID(data.EmployeeID.String()).
		SetPositionTitle(data.PositionTitle).
		SetDepartment(data.Department).
		SetEffectiveDate(data.EffectiveDate).
		SetIsRetroactive(data.IsRetroactive).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now())

	// Set optional fields that exist in schema
	if data.EndDate != nil {
		createBuilder = createBuilder.SetEndDate(*data.EndDate)
	}
	if data.ChangeReason != nil {
		createBuilder = createBuilder.SetChangeReason(*data.ChangeReason)
	}

	return createBuilder.Save(ctx)
}

// ValidateTemporalConsistency validates temporal consistency for a position
func (s *EnhancedTemporalQueryService) ValidateTemporalConsistency(ctx context.Context, tenantID, employeeID uuid.UUID, effectiveDate time.Time) error {
	// Simple validation - check for overlaps
	conflictCount, err := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
			positionhistory.EffectiveDateLTE(effectiveDate),
			positionhistory.Or(
				positionhistory.EndDateIsNil(),
				positionhistory.EndDateGT(effectiveDate),
			),
		).
		Count(ctx)

	if err != nil {
		return err
	}

	if conflictCount > 0 {
		return fmt.Errorf("temporal conflict: position already exists for employee %s at date %s",
			employeeID, effectiveDate.Format("2006-01-02"))
	}

	return nil
}

// Helper function to get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
