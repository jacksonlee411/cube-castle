package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// PostgresPositionRepository PostgreSQL职位仓储实现
type PostgresPositionRepository struct {
	db           *sqlx.DB
	outboxRepo   OutboxRepository
}

// NewPostgresPositionRepository 创建PostgreSQL职位仓储
func NewPostgresPositionRepository(db *sqlx.DB, outboxRepo OutboxRepository) *PostgresPositionRepository {
	return &PostgresPositionRepository{
		db:         db,
		outboxRepo: outboxRepo,
	}
}

// CreatePosition 创建职位 (with Outbox Pattern)
func (r *PostgresPositionRepository) CreatePosition(ctx context.Context, position Position) error {
	// Start transaction for Outbox Pattern
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Insert position
	query := `
		INSERT INTO positions (
			id, tenant_id, position_type, job_profile_id, department_id, 
			manager_position_id, status, budgeted_fte, details, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	
	detailsJSON, err := json.Marshal(position.Details)
	if err != nil {
		return fmt.Errorf("failed to marshal details: %w", err)
	}
	
	_, err = tx.ExecContext(ctx, query,
		position.ID,
		position.TenantID,
		position.PositionType,
		position.JobProfileID,
		position.DepartmentID,
		position.ManagerPositionID,
		position.Status,
		position.BudgetedFTE,
		detailsJSON,
		position.CreatedAt,
		position.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create position: %w", err)
	}

	// 2. Create Outbox event
	eventData := map[string]interface{}{
		"position_id":     position.ID,
		"tenant_id":       position.TenantID,
		"position_type":   position.PositionType,
		"department_id":   position.DepartmentID,
		"status":          position.Status,
		"budgeted_fte":    position.BudgetedFTE,
		"details":         position.Details,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := OutboxEvent{
		ID:          uuid.New(),
		TenantID:    position.TenantID,
		EventType:   "PositionCreatedEvent",
		AggregateID: position.ID,
		EventData:   eventDataJSON,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	// Insert outbox event in the same transaction
	err = r.outboxRepo.SaveEventInTransaction(ctx, &dbTransaction{tx.Tx}, event)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// UpdatePosition 更新职位 (with Outbox Pattern)
func (r *PostgresPositionRepository) UpdatePosition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	
	// Start transaction for Outbox Pattern
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 构建动态更新查询
	setParts := make([]string, 0, len(updates)+1) // +1 for updated_at
	args := make([]interface{}, 0, len(updates)+3) // +3 for updated_at, id, tenant_id
	argIndex := 1
	
	for field, value := range updates {
		// Special handling for JSON fields
		if field == "details" {
			if detailsMap, ok := value.(map[string]interface{}); ok {
				detailsJSON, err := json.Marshal(detailsMap)
				if err != nil {
					return fmt.Errorf("failed to marshal details: %w", err)
				}
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, detailsJSON)
			} else {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, value)
			}
		} else {
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
		}
		argIndex++
	}
	
	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++
	
	// 添加WHERE条件的参数
	args = append(args, id, tenantID)
	
	// 构建完整的SET子句
	setClause := strings.Join(setParts, ", ")
	
	query := fmt.Sprintf(
		"UPDATE positions SET %s WHERE id = $%d AND tenant_id = $%d",
		setClause,
		argIndex,
		argIndex+1,
	)
	
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// 2. Create Outbox event
	eventData := map[string]interface{}{
		"position_id": id,
		"tenant_id":   tenantID,
		"updates":     updates,
	}

	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := OutboxEvent{
		ID:          uuid.New(),
		TenantID:    tenantID,
		EventType:   "PositionUpdatedEvent",
		AggregateID: id,
		EventData:   eventDataJSON,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	// Insert outbox event in the same transaction
	err = r.outboxRepo.SaveEventInTransaction(ctx, &dbTransaction{tx.Tx}, event)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// DeletePosition 删除职位
func (r *PostgresPositionRepository) DeletePosition(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM positions WHERE id = $1 AND tenant_id = $2`
	_, err := r.db.ExecContext(ctx, query, id, tenantID)
	return err
}

// AssignEmployeeWithEvent 使用Outbox模式分配员工到职位
func (r *PostgresPositionRepository) AssignEmployeeWithEvent(ctx context.Context, assignment PositionAssignment, event OutboxEvent) error {
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// 1. 首先结束该员工的其他当前分配（如果是PRIMARY）
	if assignment.AssignmentType == "PRIMARY" {
		endCurrentQuery := `
			UPDATE position_assignments 
			SET is_current = false, end_date = $1, updated_at = $2
			WHERE tenant_id = $3 AND employee_id = $4 AND is_current = true AND assignment_type = 'PRIMARY'`
		
		_, err = tx.ExecContext(ctx, endCurrentQuery, 
			assignment.StartDate, time.Now(), assignment.TenantID, assignment.EmployeeID)
		if err != nil {
			return fmt.Errorf("failed to end current assignments: %w", err)
		}
	}
	
	// 2. 创建新的职位分配
	insertQuery := `
		INSERT INTO position_assignments (
			id, tenant_id, position_id, employee_id, start_date, 
			is_current, fte, assignment_type, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	_, err = tx.ExecContext(ctx, insertQuery,
		assignment.ID,
		assignment.TenantID,
		assignment.PositionID,
		assignment.EmployeeID,
		assignment.StartDate,
		assignment.IsCurrent,
		assignment.FTE,
		assignment.AssignmentType,
		assignment.CreatedAt,
		assignment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create assignment: %w", err)
	}
	
	// 3. 在同一事务中保存事件到发件箱
	err = r.outboxRepo.SaveEventInTransaction(ctx, &dbTransaction{tx}, event)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	
	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// RemoveEmployeeWithEvent 使用Outbox模式移除员工职位
func (r *PostgresPositionRepository) RemoveEmployeeWithEvent(ctx context.Context, positionID, employeeID uuid.UUID, endDate time.Time, reason string, event OutboxEvent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// 1. 更新分配记录
	updateQuery := `
		UPDATE position_assignments 
		SET is_current = false, end_date = $1, updated_at = $2
		WHERE position_id = $3 AND employee_id = $4 AND is_current = true`
	
	result, err := tx.ExecContext(ctx, updateQuery, endDate, time.Now(), positionID, employeeID)
	if err != nil {
		return fmt.Errorf("failed to update assignment: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no current assignment found for employee %s in position %s", employeeID, positionID)
	}
	
	// 2. 在同一事务中保存事件到发件箱
	err = r.outboxRepo.SaveEventInTransaction(ctx, &dbTransaction{tx}, event)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// ValidateEmployeePositionAssignment 验证员工职位分配的有效性
func (r *PostgresPositionRepository) ValidateEmployeePositionAssignment(ctx context.Context, employeeID, positionID, tenantID uuid.UUID) (bool, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM employees WHERE id = $1 AND tenant_id = $3) as employee_exists,
			(SELECT COUNT(*) FROM positions WHERE id = $2 AND tenant_id = $3) as position_exists`
	
	var employeeExists, positionExists int
	err := r.db.QueryRowContext(ctx, query, employeeID, positionID, tenantID).Scan(&employeeExists, &positionExists)
	if err != nil {
		return false, fmt.Errorf("failed to validate assignment: %w", err)
	}
	
	return employeeExists > 0 && positionExists > 0, nil
}

// EndPositionOccupancy 结束职位占用
func (r *PostgresPositionRepository) EndPositionOccupancy(ctx context.Context, positionID, employeeID uuid.UUID, endDate time.Time, reason string) error {
	query := `
		UPDATE position_assignments 
		SET is_current = false, end_date = $1, updated_at = $2
		WHERE position_id = $3 AND employee_id = $4 AND is_current = true`
	
	_, err := r.db.ExecContext(ctx, query, endDate, time.Now(), positionID, employeeID)
	return err
}

// TransferPosition 转移职位
func (r *PostgresPositionRepository) TransferPosition(ctx context.Context, id uuid.UUID, newDeptID uuid.UUID, newManagerID *uuid.UUID, effectiveDate time.Time, reason string) error {
	updates := map[string]interface{}{
		"department_id": newDeptID,
		"updated_at":   time.Now(),
	}
	
	if newManagerID != nil {
		updates["manager_position_id"] = *newManagerID
	}
	
	return r.UpdatePosition(ctx, id, uuid.Nil, updates) // TODO: 需要租户ID
}

// UpdatePositionStatus 更新职位状态
func (r *PostgresPositionRepository) UpdatePositionStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, newStatus string, changedBy uuid.UUID, reason string) error {
	updates := map[string]interface{}{
		"status":     newStatus,
		"updated_at": time.Now(),
	}
	
	return r.UpdatePosition(ctx, id, tenantID, updates)
}

// dbTransaction 实现Transaction接口的包装器
type dbTransaction struct {
	tx *sql.Tx
}

func (t *dbTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *dbTransaction) Rollback() error {
	return t.tx.Rollback()
}

// PostgresOutboxRepository PostgreSQL发件箱仓储实现
type PostgresOutboxRepository struct {
	db *sqlx.DB
}

// NewPostgresOutboxRepository 创建PostgreSQL发件箱仓储
func NewPostgresOutboxRepository(db *sqlx.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

// SaveEventInTransaction 在事务中保存事件
func (r *PostgresOutboxRepository) SaveEventInTransaction(ctx context.Context, tx Transaction, event OutboxEvent) error {
	dbTx := tx.(*dbTransaction).tx
	
	query := `
		INSERT INTO outbox_events (
			id, tenant_id, event_type, aggregate_id, event_data, 
			status, attempt_count, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := dbTx.ExecContext(ctx, query,
		event.ID,
		event.TenantID,
		event.EventType,
		event.AggregateID,
		event.EventData,
		event.Status,
		event.AttemptCount,
		event.CreatedAt,
	)
	
	return err
}

// GetPendingEvents 获取待处理的事件
func (r *PostgresOutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `
		SELECT id, tenant_id, event_type, aggregate_id, event_data, 
			   status, attempt_count, created_at, processed_at, error_message
		FROM outbox_events 
		WHERE status = 'PENDING' AND attempt_count < 5
		ORDER BY created_at ASC 
		LIMIT $1`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var events []OutboxEvent
	for rows.Next() {
		var event OutboxEvent
		err := rows.Scan(
			&event.ID,
			&event.TenantID,
			&event.EventType,
			&event.AggregateID,
			&event.EventData,
			&event.Status,
			&event.AttemptCount,
			&event.CreatedAt,
			&event.ProcessedAt,
			&event.ErrorMessage,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	
	return events, nil
}

// MarkEventAsProcessed 标记事件为已处理
func (r *PostgresOutboxRepository) MarkEventAsProcessed(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE outbox_events SET status = 'PROCESSED', processed_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), eventID)
	return err
}

// MarkEventAsFailed 标记事件处理失败
func (r *PostgresOutboxRepository) MarkEventAsFailed(ctx context.Context, eventID uuid.UUID, errorMsg string) error {
	query := `UPDATE outbox_events SET status = 'FAILED', error_message = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, errorMsg, eventID)
	return err
}

// IncrementAttemptCount 增加重试次数
func (r *PostgresOutboxRepository) IncrementAttemptCount(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE outbox_events SET attempt_count = attempt_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, eventID)
	return err
}

// CleanupProcessedEvents 清理旧的已处理事件
func (r *PostgresOutboxRepository) CleanupProcessedEvents(ctx context.Context, olderThan time.Time) error {
	query := `DELETE FROM outbox_events WHERE status = 'PROCESSED' AND processed_at < $1`
	_, err := r.db.ExecContext(ctx, query, olderThan)
	return err
}

// CreatePositionOccupancyHistory 创建职位占用历史记录
func (r *PostgresPositionRepository) CreatePositionOccupancyHistory(ctx context.Context, history PositionOccupancyHistory) error {
	query := `
		INSERT INTO position_occupancy_history (
			id, tenant_id, position_id, employee_id, start_date, end_date,
			is_current, fte, assignment_type, pay_grade_id, reason,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`
	
	_, err := r.db.ExecContext(ctx, query,
		history.ID,
		history.TenantID,
		history.PositionID,
		history.EmployeeID,
		history.StartDate,
		history.EndDate,
		history.IsCurrent,
		history.FTE,
		history.AssignmentType,
		history.PayGradeID,
		history.Reason,
		history.CreatedAt,
		history.UpdatedAt,
	)
	
	return err
}