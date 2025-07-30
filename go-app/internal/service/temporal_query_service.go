// internal/service/temporal_query_service.go
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

// TemporalQueryService provides temporal data query capabilities
type TemporalQueryService struct {
	client *entgo.Client
	logger *logging.StructuredLogger
}

// NewTemporalQueryService creates a new temporal query service
func NewTemporalQueryService(client *entgo.Client, logger *logging.StructuredLogger) *TemporalQueryService {
	return &TemporalQueryService{
		client: client,
		logger: logger,
	}
}

// PositionSnapshot represents a point-in-time view of an employee's position
type PositionSnapshot struct {
	PositionHistoryID   uuid.UUID  `json:"position_history_id"`
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
	s.logger.Info("Querying position as of date",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"as_of_date", asOfDate,
	)

	position, err := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
			positionhistory.EffectiveDateLTE(asOfDate),
			positionhistory.Or(
				positionhistory.EndDateIsNil(),
				positionhistory.EndDateGT(asOfDate),
			),
		).
		Order(positionhistory.ByEffectiveDate(sql.OrderDesc())).
		First(ctx)

	if err != nil {
		if entgo.IsNotFound(err) {
			s.logger.Warn("No position found for employee at date",
				"tenant_id", tenantID,
				"employee_id", employeeID,
				"as_of_date", asOfDate,
			)
			return nil, fmt.Errorf("no position found for employee %s at date %s",
				employeeID, asOfDate.Format("2006-01-02"))
		}
		s.logger.Error("Failed to query position as of date",
			"error", err.Error(),
			"tenant_id", tenantID,
			"employee_id", employeeID,
			"as_of_date", asOfDate,
		)
		return nil, err
	}

	employeeUUID, err := uuid.Parse(position.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID format: %w", err)
	}

	positionUUID, err := uuid.Parse(position.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid position ID format: %w", err)
	}

	snapshot := &PositionSnapshot{
		PositionHistoryID:   positionUUID,
		EmployeeID:          employeeUUID,
		PositionTitle:       position.PositionTitle,
		Department:          position.Department,
		JobLevel:            nil,         // Field not available in current schema
		Location:            nil,         // Field not available in current schema
		EmploymentType:      "FULL_TIME", // Default value as field not available
		ReportsToEmployeeID: nil,         // Field not available in current schema
		EffectiveDate:       position.EffectiveDate,
		EndDate:             position.EndDate,
		ChangeReason:        position.ChangeReason,
		IsRetroactive:       position.IsRetroactive,
		MinSalary:           nil, // Field not available in current schema
		MaxSalary:           nil, // Field not available in current schema
		Currency:            nil, // Field not available in current schema
	}

	// No cleanup needed as fields are already set to appropriate defaults

	s.logger.Info("Successfully retrieved position as of date",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"as_of_date", asOfDate,
		"position_history_id", position.ID,
		"position_title", position.PositionTitle,
	)

	return snapshot, nil
}

// GetPositionTimeline 获取员工完整职位时间线
func (s *TemporalQueryService) GetPositionTimeline(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	fromDate, toDate *time.Time,
) ([]*PositionSnapshot, error) {
	s.logger.Info("Querying position timeline",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"from_date", fromDate,
		"to_date", toDate,
	)

	query := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
		)

	if fromDate != nil {
		query = query.Where(
			positionhistory.Or(
				positionhistory.EndDateIsNil(),
				positionhistory.EndDateGTE(*fromDate),
			),
		)
	}

	if toDate != nil {
		query = query.Where(positionhistory.EffectiveDateLTE(*toDate))
	}

	positions, err := query.
		Order(positionhistory.ByEffectiveDate()).
		All(ctx)

	if err != nil {
		s.logger.Error("Failed to query position timeline",
			"error", err.Error(),
			"tenant_id", tenantID,
			"employee_id", employeeID,
			"from_date", fromDate,
			"to_date", toDate,
		)
		return nil, err
	}

	snapshots := make([]*PositionSnapshot, len(positions))
	for i, pos := range positions {
		employeeUUID, err := uuid.Parse(pos.EmployeeID)
		if err != nil {
			return nil, fmt.Errorf("invalid employee ID format: %w", err)
		}

		positionUUID, err := uuid.Parse(pos.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid position ID format: %w", err)
		}

		snapshot := &PositionSnapshot{
			PositionHistoryID:   positionUUID,
			EmployeeID:          employeeUUID,
			PositionTitle:       pos.PositionTitle,
			Department:          pos.Department,
			JobLevel:            nil,         // Field not available in current schema
			Location:            nil,         // Field not available in current schema
			EmploymentType:      "FULL_TIME", // Default value as field not available
			ReportsToEmployeeID: nil,         // Field not available in current schema
			EffectiveDate:       pos.EffectiveDate,
			EndDate:             pos.EndDate,
			ChangeReason:        pos.ChangeReason,
			IsRetroactive:       pos.IsRetroactive,
			MinSalary:           nil, // Field not available in current schema
			MaxSalary:           nil, // Field not available in current schema
			Currency:            nil, // Field not available in current schema
		}

		// No cleanup needed as fields are already set to appropriate defaults

		snapshots[i] = snapshot
	}

	s.logger.Info("Successfully retrieved position timeline",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"from_date", fromDate,
		"to_date", toDate,
		"timeline_count", len(snapshots),
	)

	return snapshots, nil
}

// ValidateTemporalConsistency 验证时态一致性
func (s *TemporalQueryService) ValidateTemporalConsistency(
	ctx context.Context,
	tenantID, employeeID uuid.UUID,
	newEffectiveDate time.Time,
) error {
	s.logger.Info("Validating temporal consistency",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"new_effective_date", newEffectiveDate,
	)

	// 检查是否与现有记录冲突
	conflictCount, err := s.client.PositionHistory.Query().
		Where(
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
			positionhistory.EffectiveDateLTE(newEffectiveDate),
			positionhistory.Or(
				positionhistory.EndDateIsNil(),
				positionhistory.EndDateGT(newEffectiveDate),
			),
		).
		Count(ctx)

	if err != nil {
		s.logger.Error("Failed to validate temporal consistency",
			"error", err.Error(),
			"tenant_id", tenantID,
			"employee_id", employeeID,
			"new_effective_date", newEffectiveDate,
		)
		return err
	}

	if conflictCount > 0 {
		s.logger.Warn("Temporal conflict detected",
			"tenant_id", tenantID,
			"employee_id", employeeID,
			"new_effective_date", newEffectiveDate,
			"conflict_count", conflictCount,
		)
		return fmt.Errorf("temporal conflict: position already exists for employee %s at date %s",
			employeeID, newEffectiveDate.Format("2006-01-02"))
	}

	s.logger.Info("Temporal consistency validation passed",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"new_effective_date", newEffectiveDate,
	)

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
	s.logger.Info("Creating position snapshot",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"created_by", createdBy,
		"effective_date", positionData.EffectiveDate,
		"position_title", positionData.PositionTitle,
	)

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

	// 创建新记录 (simplified - only using available fields)
	createBuilder := s.client.PositionHistory.Create().
		SetOrganizationID(tenantID.String()).
		SetEmployeeID(employeeID.String()).
		SetPositionTitle(positionData.PositionTitle).
		SetDepartment(positionData.Department).
		SetEffectiveDate(positionData.EffectiveDate).
		SetIsRetroactive(positionData.IsRetroactive).
		SetCreatedAt(time.Now())

	// Set optional fields
	if positionData.EndDate != nil {
		createBuilder = createBuilder.SetEndDate(*positionData.EndDate)
	}
	if positionData.ChangeReason != nil {
		createBuilder = createBuilder.SetChangeReason(*positionData.ChangeReason)
	}

	position, err := createBuilder.Save(ctx)
	if err != nil {
		s.logger.Error("Failed to create position snapshot",
			"error", err.Error(),
			"tenant_id", tenantID,
			"employee_id", employeeID,
			"created_by", createdBy,
			"effective_date", positionData.EffectiveDate,
		)
		return nil, err
	}

	s.logger.Info("Successfully created position snapshot",
		"tenant_id", tenantID,
		"employee_id", employeeID,
		"position_history_id", position.ID,
		"effective_date", positionData.EffectiveDate,
		"position_title", positionData.PositionTitle,
	)

	// Convert to snapshot
	employeeUUID, err := uuid.Parse(position.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID format: %w", err)
	}

	positionUUID, err := uuid.Parse(position.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid position ID format: %w", err)
	}

	snapshot := &PositionSnapshot{
		PositionHistoryID:   positionUUID,
		EmployeeID:          employeeUUID,
		PositionTitle:       position.PositionTitle,
		Department:          position.Department,
		JobLevel:            nil,         // Field not available in current schema
		Location:            nil,         // Field not available in current schema
		EmploymentType:      "FULL_TIME", // Default value as field not available
		ReportsToEmployeeID: nil,         // Field not available in current schema
		EffectiveDate:       position.EffectiveDate,
		EndDate:             position.EndDate,
		ChangeReason:        position.ChangeReason,
		IsRetroactive:       position.IsRetroactive,
		MinSalary:           nil, // Field not available in current schema
		MaxSalary:           nil, // Field not available in current schema
		Currency:            nil, // Field not available in current schema
	}

	// No cleanup needed as fields are already set to appropriate defaults

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
			positionhistory.OrganizationIDEQ(tenantID.String()),
			positionhistory.EmployeeIDEQ(employeeID.String()),
			positionhistory.EndDateIsNil(),
			positionhistory.EffectiveDateLT(newEffectiveDate),
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

		s.logger.Info("Closed previous position",
			"position_history_id", pos.ID,
			"employee_id", employeeID,
			"end_date", endDate,
		)
	}

	return nil
}
