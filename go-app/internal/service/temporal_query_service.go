// internal/service/temporal_query_service.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/ent"
	"github.com/gaogu/cube-castle/go-app/ent/positionhistory"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// TemporalQueryService provides temporal data query capabilities
type TemporalQueryService struct {
	client *ent.Client
	logger *logging.StructuredLogger
}

// NewTemporalQueryService creates a new temporal query service
func NewTemporalQueryService(client *ent.Client, logger *logging.StructuredLogger) *TemporalQueryService {
	return &TemporalQueryService{
		client: client,
		logger: logger,
	}
}

// PositionSnapshot represents a point-in-time view of an employee's position
type PositionSnapshot struct {
	PositionHistoryID    uuid.UUID  `json:"position_history_id"`
	EmployeeID          uuid.UUID  `json:"employee_id"`
	PositionTitle       string     `json:"position_title"`
	Department          string     `json:"department"`
	JobLevel            *string    `json:"job_level,omitempty"`
	Location            *string    `json:"location,omitempty"`
	EmploymentType      string     `json:"employment_type"`
	ReportsToEmployeeID *uuid.UUID `json:"reports_to_employee_id,omitempty"`
	EffectiveDate       time.Time  `json:"effective_date"`
	EndDate             *time.Time `json:"end_date,omitempty"`
	ChangeReason        *string    `json:"change_reason,omitempty"`
	IsRetroactive       bool       `json:"is_retroactive"`
	MinSalary           *float64   `json:"min_salary,omitempty"`
	MaxSalary           *float64   `json:"max_salary,omitempty"`
	Currency            *string    `json:"currency,omitempty"`
}

// GetPositionAsOfDate 获取指定日期的职位信息
func (s *TemporalQueryService) GetPositionAsOfDate(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	asOfDate time.Time,
) (*PositionSnapshot, error) {
	s.logger.LogInfo(ctx, "Querying position as of date", map[string]interface{}{
		"tenant_id":   tenantID,
		"employee_id": employeeID,
		"as_of_date":  asOfDate,
	})

	position, err := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EmployeeIDEQ(employeeID),
			position_history.EffectiveDateLTE(asOfDate),
			position_history.Or(
				position_history.EndDateIsNil(),
				position_history.EndDateGT(asOfDate),
			),
		).
		Order(ent.Desc(position_history.FieldEffectiveDate)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			s.logger.LogWarning(ctx, "No position found for employee at date", map[string]interface{}{
				"tenant_id":   tenantID,
				"employee_id": employeeID,
				"as_of_date":  asOfDate,
			})
			return nil, fmt.Errorf("no position found for employee %s at date %s", 
				employeeID, asOfDate.Format("2006-01-02"))
		}
		s.logger.LogError(ctx, "Failed to query position as of date", err, map[string]interface{}{
			"tenant_id":   tenantID,
			"employee_id": employeeID,
			"as_of_date":  asOfDate,
		})
		return nil, err
	}

	snapshot := &PositionSnapshot{
		PositionHistoryID:    position.ID,
		EmployeeID:          position.EmployeeID,
		PositionTitle:       position.PositionTitle,
		Department:          position.Department,
		JobLevel:            &position.JobLevel,
		Location:            &position.Location,
		EmploymentType:      position.EmploymentType,
		ReportsToEmployeeID: &position.ReportsToEmployeeID,
		EffectiveDate:       position.EffectiveDate,
		EndDate:             position.EndDate,
		ChangeReason:        &position.ChangeReason,
		IsRetroactive:       position.IsRetroactive,
		MinSalary:           &position.MinSalary,
		MaxSalary:           &position.MaxSalary,
		Currency:            &position.Currency,
	}

	// Clean up nil fields if they are empty
	if position.JobLevel == "" {
		snapshot.JobLevel = nil
	}
	if position.Location == "" {
		snapshot.Location = nil
	}
	if position.ReportsToEmployeeID == uuid.Nil {
		snapshot.ReportsToEmployeeID = nil
	}
	if position.ChangeReason == "" {
		snapshot.ChangeReason = nil
	}
	if position.MinSalary == 0 {
		snapshot.MinSalary = nil
	}
	if position.MaxSalary == 0 {
		snapshot.MaxSalary = nil
	}
	if position.Currency == "" {
		snapshot.Currency = nil
	}

	s.logger.LogInfo(ctx, "Successfully retrieved position as of date", map[string]interface{}{
		"tenant_id":           tenantID,
		"employee_id":         employeeID,
		"as_of_date":          asOfDate,
		"position_history_id": position.ID,
		"position_title":      position.PositionTitle,
	})

	return snapshot, nil
}

// GetPositionTimeline 获取员工完整职位时间线
func (s *TemporalQueryService) GetPositionTimeline(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	fromDate, toDate *time.Time,
) ([]*PositionSnapshot, error) {
	s.logger.LogInfo(ctx, "Querying position timeline", map[string]interface{}{
		"tenant_id":   tenantID,
		"employee_id": employeeID,
		"from_date":   fromDate,
		"to_date":     toDate,
	})

	query := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EmployeeIDEQ(employeeID),
		)

	if fromDate != nil {
		query = query.Where(
			position_history.Or(
				position_history.EndDateIsNil(),
				position_history.EndDateGTE(*fromDate),
			),
		)
	}

	if toDate != nil {
		query = query.Where(position_history.EffectiveDateLTE(*toDate))
	}

	positions, err := query.
		Order(ent.Asc(position_history.FieldEffectiveDate)).
		All(ctx)

	if err != nil {
		s.logger.LogError(ctx, "Failed to query position timeline", err, map[string]interface{}{
			"tenant_id":   tenantID,
			"employee_id": employeeID,
			"from_date":   fromDate,
			"to_date":     toDate,
		})
		return nil, err
	}

	snapshots := make([]*PositionSnapshot, len(positions))
	for i, pos := range positions {
		snapshot := &PositionSnapshot{
			PositionHistoryID:    pos.ID,
			EmployeeID:          pos.EmployeeID,
			PositionTitle:       pos.PositionTitle,
			Department:          pos.Department,
			JobLevel:            &pos.JobLevel,
			Location:            &pos.Location,
			EmploymentType:      pos.EmploymentType,
			ReportsToEmployeeID: &pos.ReportsToEmployeeID,
			EffectiveDate:       pos.EffectiveDate,
			EndDate:             pos.EndDate,
			ChangeReason:        &pos.ChangeReason,
			IsRetroactive:       pos.IsRetroactive,
			MinSalary:           &pos.MinSalary,
			MaxSalary:           &pos.MaxSalary,
			Currency:            &pos.Currency,
		}

		// Clean up nil fields if they are empty
		if pos.JobLevel == "" {
			snapshot.JobLevel = nil
		}
		if pos.Location == "" {
			snapshot.Location = nil
		}
		if pos.ReportsToEmployeeID == uuid.Nil {
			snapshot.ReportsToEmployeeID = nil
		}
		if pos.ChangeReason == "" {
			snapshot.ChangeReason = nil
		}
		if pos.MinSalary == 0 {
			snapshot.MinSalary = nil
		}
		if pos.MaxSalary == 0 {
			snapshot.MaxSalary = nil
		}
		if pos.Currency == "" {
			snapshot.Currency = nil
		}

		snapshots[i] = snapshot
	}

	s.logger.LogInfo(ctx, "Successfully retrieved position timeline", map[string]interface{}{
		"tenant_id":      tenantID,
		"employee_id":    employeeID,
		"from_date":      fromDate,
		"to_date":        toDate,
		"timeline_count": len(snapshots),
	})

	return snapshots, nil
}

// ValidateTemporalConsistency 验证时态一致性
func (s *TemporalQueryService) ValidateTemporalConsistency(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	newEffectiveDate time.Time,
) error {
	s.logger.LogInfo(ctx, "Validating temporal consistency", map[string]interface{}{
		"tenant_id":          tenantID,
		"employee_id":        employeeID,
		"new_effective_date": newEffectiveDate,
	})

	// 检查是否与现有记录冲突
	conflictCount, err := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EmployeeIDEQ(employeeID),
			position_history.EffectiveDateLTE(newEffectiveDate),
			position_history.Or(
				position_history.EndDateIsNil(),
				position_history.EndDateGT(newEffectiveDate),
			),
		).
		Count(ctx)

	if err != nil {
		s.logger.LogError(ctx, "Failed to validate temporal consistency", err, map[string]interface{}{
			"tenant_id":          tenantID,
			"employee_id":        employeeID,
			"new_effective_date": newEffectiveDate,
		})
		return err
	}

	if conflictCount > 0 {
		s.logger.LogWarning(ctx, "Temporal conflict detected", map[string]interface{}{
			"tenant_id":          tenantID,
			"employee_id":        employeeID,
			"new_effective_date": newEffectiveDate,
			"conflict_count":     conflictCount,
		})
		return fmt.Errorf("temporal conflict: position already exists for employee %s at date %s", 
			employeeID, newEffectiveDate.Format("2006-01-02"))
	}

	s.logger.LogInfo(ctx, "Temporal consistency validation passed", map[string]interface{}{
		"tenant_id":          tenantID,
		"employee_id":        employeeID,
		"new_effective_date": newEffectiveDate,
	})

	return nil
}

// GetCurrentPosition 获取员工当前职位（等同于 GetPositionAsOfDate with time.Now()）
func (s *TemporalQueryService) GetCurrentPosition(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
) (*PositionSnapshot, error) {
	return s.GetPositionAsOfDate(ctx, tenantID, employeeID, time.Now())
}

// CreatePositionSnapshot 创建新的职位历史记录
func (s *TemporalQueryService) CreatePositionSnapshot(
	ctx context.Context,
	tenantID, employeeID, createdBy uuid.UUID,
	positionData PositionSnapshotData,
) (*PositionSnapshot, error) {
	s.logger.LogInfo(ctx, "Creating position snapshot", map[string]interface{}{
		"tenant_id":       tenantID,
		"employee_id":     employeeID,
		"created_by":      createdBy,
		"effective_date":  positionData.EffectiveDate,
		"position_title":  positionData.PositionTitle,
	})

	// 验证时态一致性
	if err := s.ValidateTemporalConsistency(ctx, tenantID, employeeID, positionData.EffectiveDate); err != nil {
		return nil, fmt.Errorf("temporal consistency validation failed: %w", err)
	}

	// 如果这是一个新的当前职位，需要关闭之前的记录
	if positionData.EndDate == nil {
		if err := s.closePreviousPositions(ctx, tenantID, employeeID, positionData.EffectiveDate); err != nil {
			return nil, fmt.Errorf("failed to close previous positions: %w", err)
		}
	}

	// 创建新记录
	createBuilder := s.client.PositionHistory.Create().
		SetTenantID(tenantID).
		SetEmployeeID(employeeID).
		SetPositionTitle(positionData.PositionTitle).
		SetDepartment(positionData.Department).
		SetEmploymentType(positionData.EmploymentType).
		SetEffectiveDate(positionData.EffectiveDate).
		SetIsRetroactive(positionData.IsRetroactive).
		SetCreatedBy(createdBy).
		SetCreatedAt(time.Now())

	// Set optional fields
	if positionData.JobLevel != nil {
		createBuilder = createBuilder.SetJobLevel(*positionData.JobLevel)
	}
	if positionData.Location != nil {
		createBuilder = createBuilder.SetLocation(*positionData.Location)
	}
	if positionData.ReportsToEmployeeID != nil {
		createBuilder = createBuilder.SetReportsToEmployeeID(*positionData.ReportsToEmployeeID)
	}
	if positionData.EndDate != nil {
		createBuilder = createBuilder.SetEndDate(*positionData.EndDate)
	}
	if positionData.ChangeReason != nil {
		createBuilder = createBuilder.SetChangeReason(*positionData.ChangeReason)
	}
	if positionData.MinSalary != nil {
		createBuilder = createBuilder.SetMinSalary(*positionData.MinSalary)
	}
	if positionData.MaxSalary != nil {
		createBuilder = createBuilder.SetMaxSalary(*positionData.MaxSalary)
	}
	if positionData.Currency != nil {
		createBuilder = createBuilder.SetCurrency(*positionData.Currency)
	}

	position, err := createBuilder.Save(ctx)
	if err != nil {
		s.logger.LogError(ctx, "Failed to create position snapshot", err, map[string]interface{}{
			"tenant_id":      tenantID,
			"employee_id":    employeeID,
			"created_by":     createdBy,
			"effective_date": positionData.EffectiveDate,
		})
		return nil, err
	}

	s.logger.LogInfo(ctx, "Successfully created position snapshot", map[string]interface{}{
		"tenant_id":           tenantID,
		"employee_id":         employeeID,
		"position_history_id": position.ID,
		"effective_date":      positionData.EffectiveDate,
		"position_title":      positionData.PositionTitle,
	})

	// Convert to snapshot
	snapshot := &PositionSnapshot{
		PositionHistoryID:    position.ID,
		EmployeeID:          position.EmployeeID,
		PositionTitle:       position.PositionTitle,
		Department:          position.Department,
		JobLevel:            &position.JobLevel,
		Location:            &position.Location,
		EmploymentType:      position.EmploymentType,
		ReportsToEmployeeID: &position.ReportsToEmployeeID,
		EffectiveDate:       position.EffectiveDate,
		EndDate:             position.EndDate,
		ChangeReason:        &position.ChangeReason,
		IsRetroactive:       position.IsRetroactive,
		MinSalary:           &position.MinSalary,
		MaxSalary:           &position.MaxSalary,
		Currency:            &position.Currency,
	}

	// Clean up nil fields
	if position.JobLevel == "" {
		snapshot.JobLevel = nil
	}
	if position.Location == "" {
		snapshot.Location = nil
	}
	if position.ReportsToEmployeeID == uuid.Nil {
		snapshot.ReportsToEmployeeID = nil
	}
	if position.ChangeReason == "" {
		snapshot.ChangeReason = nil
	}
	if position.MinSalary == 0 {
		snapshot.MinSalary = nil
	}
	if position.MaxSalary == 0 {
		snapshot.MaxSalary = nil
	}
	if position.Currency == "" {
		snapshot.Currency = nil
	}

	return snapshot, nil
}

// PositionSnapshotData represents data for creating a new position snapshot
type PositionSnapshotData struct {
	PositionTitle       string     `json:"position_title"`
	Department          string     `json:"department"`
	JobLevel            *string    `json:"job_level,omitempty"`
	Location            *string    `json:"location,omitempty"`
	EmploymentType      string     `json:"employment_type"`
	ReportsToEmployeeID *uuid.UUID `json:"reports_to_employee_id,omitempty"`
	EffectiveDate       time.Time  `json:"effective_date"`
	EndDate             *time.Time `json:"end_date,omitempty"`
	ChangeReason        *string    `json:"change_reason,omitempty"`
	IsRetroactive       bool       `json:"is_retroactive"`
	MinSalary           *float64   `json:"min_salary,omitempty"`
	MaxSalary           *float64   `json:"max_salary,omitempty"`
	Currency            *string    `json:"currency,omitempty"`
}

// closePreviousPositions closes any open position records before the new effective date
func (s *TemporalQueryService) closePreviousPositions(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	newEffectiveDate time.Time,
) error {
	// Find current open position (end_date is NULL)
	currentPositions, err := s.client.PositionHistory.Query().
		Where(
			position_history.TenantIDEQ(tenantID),
			position_history.EmployeeIDEQ(employeeID),
			position_history.EndDateIsNil(),
			position_history.EffectiveDateLT(newEffectiveDate),
		).
		All(ctx)

	if err != nil {
		return err
	}

	// Close each current position by setting end_date to the day before new effective date
	endDate := newEffectiveDate.Add(-24 * time.Hour)
	for _, pos := range currentPositions {
		_, err := s.client.PositionHistory.UpdateOneID(pos.ID).
			SetEndDate(endDate).
			Save(ctx)
		if err != nil {
			return err
		}

		s.logger.LogInfo(ctx, "Closed previous position", map[string]interface{}{
			"position_history_id": pos.ID,
			"employee_id":         employeeID,
			"end_date":            endDate,
		})
	}

	return nil
}