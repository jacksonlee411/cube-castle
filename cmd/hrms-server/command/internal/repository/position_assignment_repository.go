package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"cube-castle/cmd/hrms-server/command/internal/types"
)

type PositionAssignmentRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewPositionAssignmentRepository(db *sql.DB, logger *log.Logger) *PositionAssignmentRepository {
	return &PositionAssignmentRepository{db: db, logger: logger}
}

func (r *PositionAssignmentRepository) queryRow(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return r.db.QueryRowContext(ctx, query, args...)
}

func (r *PositionAssignmentRepository) queryRows(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return r.db.QueryContext(ctx, query, args...)
}

func (r *PositionAssignmentRepository) exec(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

func (r *PositionAssignmentRepository) CreateAssignment(ctx context.Context, tx *sql.Tx, entity *types.PositionAssignment) (*types.PositionAssignment, error) {
	query := `INSERT INTO position_assignments (
tenant_id, position_code, position_record_id, employee_id, employee_name, employee_number,
assignment_type, assignment_status, fte, effective_date, end_date, acting_until, auto_revert, reminder_sent_at, is_current, notes
) VALUES (
$1,$2,$3,$4,$5,$6,
$7,$8,$9,$10,$11,$12,$13,$14,$15,$16
) RETURNING assignment_id, assignment_status, is_current, created_at, updated_at`

	var employeeNumber interface{}
	if entity.EmployeeNumber.Valid {
		employeeNumber = entity.EmployeeNumber.String
	} else {
		employeeNumber = nil
	}

	var endDate interface{}
	if entity.EndDate.Valid {
		endDate = entity.EndDate.Time
	} else {
		endDate = nil
	}

	var actingUntil interface{}
	if entity.ActingUntil.Valid {
		actingUntil = entity.ActingUntil.Time
	} else {
		actingUntil = nil
	}

	var reminderSentAt interface{}
	if entity.ReminderSentAt.Valid {
		reminderSentAt = entity.ReminderSentAt.Time
	} else {
		reminderSentAt = nil
	}

	var notes interface{}
	if entity.Notes.Valid {
		notes = entity.Notes.String
	} else {
		notes = nil
	}

	if err := r.queryRow(ctx, tx, query,
		entity.TenantID,
		entity.PositionCode,
		entity.PositionRecordID,
		entity.EmployeeID,
		entity.EmployeeName,
		employeeNumber,
		entity.AssignmentType,
		entity.AssignmentStatus,
		entity.FTE,
		entity.EffectiveDate,
		endDate,
		actingUntil,
		entity.AutoRevert,
		reminderSentAt,
		entity.IsCurrent,
		notes,
	).Scan(&entity.AssignmentID, &entity.AssignmentStatus, &entity.IsCurrent, &entity.CreatedAt, &entity.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to create position assignment: %w", err)
	}

	return entity, nil
}

func (r *PositionAssignmentRepository) GetByID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error) {
	query := `SELECT assignment_id, tenant_id, position_code, position_record_id, employee_id, employee_name, employee_number,
assignment_type, assignment_status, fte, effective_date, end_date, acting_until, auto_revert, reminder_sent_at, is_current, notes, created_at, updated_at
FROM position_assignments
WHERE tenant_id = $1 AND assignment_id = $2`

	var entity types.PositionAssignment
	if err := r.queryRow(ctx, tx, query, tenantID, assignmentID).Scan(
		&entity.AssignmentID,
		&entity.TenantID,
		&entity.PositionCode,
		&entity.PositionRecordID,
		&entity.EmployeeID,
		&entity.EmployeeName,
		&entity.EmployeeNumber,
		&entity.AssignmentType,
		&entity.AssignmentStatus,
		&entity.FTE,
		&entity.EffectiveDate,
		&entity.EndDate,
		&entity.ActingUntil,
		&entity.AutoRevert,
		&entity.ReminderSentAt,
		&entity.IsCurrent,
		&entity.Notes,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load position assignment: %w", err)
	}

	return &entity, nil
}

func (r *PositionAssignmentRepository) CloseAssignment(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID, endDate time.Time, notes *string) error {
	query := `UPDATE position_assignments
SET assignment_status = 'ENDED',
    end_date = $3,
    auto_revert = false,
    is_current = false,
    notes = CASE WHEN $4 THEN notes ELSE $5 END,
    updated_at = NOW()
WHERE tenant_id = $1 AND assignment_id = $2`

	preserveNotes := true
	trimmed := ""
	if notes != nil {
		trimmed = strings.TrimSpace(*notes)
		if trimmed != "" {
			preserveNotes = false
		}
	}

	var notesVal sql.NullString
	if !preserveNotes {
		notesVal = sql.NullString{String: trimmed, Valid: true}
	}

	result, err := r.exec(ctx, tx, query, tenantID, assignmentID, endDate, preserveNotes, notesVal)
	if err != nil {
		return fmt.Errorf("failed to close position assignment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows for assignment close: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("assignment not found for closing")
	}
	return nil
}

func (r *PositionAssignmentRepository) ListByPosition(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string) ([]types.PositionAssignment, error) {
	query := `SELECT assignment_id, tenant_id, position_code, position_record_id, employee_id, employee_name, employee_number,
assignment_type, assignment_status, fte, effective_date, end_date, acting_until, auto_revert, reminder_sent_at, is_current, notes, created_at, updated_at
FROM position_assignments
WHERE tenant_id = $1 AND position_code = $2
ORDER BY effective_date DESC, created_at DESC`

	rows, err := r.queryRows(ctx, tx, query, tenantID, positionCode)
	if err != nil {
		return nil, fmt.Errorf("failed to list position assignments: %w", err)
	}
	defer rows.Close()

	var result []types.PositionAssignment
	for rows.Next() {
		var entity types.PositionAssignment
		if err := rows.Scan(
			&entity.AssignmentID,
			&entity.TenantID,
			&entity.PositionCode,
			&entity.PositionRecordID,
			&entity.EmployeeID,
			&entity.EmployeeName,
			&entity.EmployeeNumber,
			&entity.AssignmentType,
			&entity.AssignmentStatus,
			&entity.FTE,
			&entity.EffectiveDate,
			&entity.EndDate,
			&entity.ActingUntil,
			&entity.AutoRevert,
			&entity.ReminderSentAt,
			&entity.IsCurrent,
			&entity.Notes,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan assignment row: %w", err)
		}
		result = append(result, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assignment iteration error: %w", err)
	}

	return result, nil
}

func normalizeAssignmentTypes(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, v := range values {
		trimmed := strings.ToUpper(strings.TrimSpace(v))
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func (r *PositionAssignmentRepository) ListWithOptions(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string, opts types.AssignmentListOptions) ([]types.PositionAssignment, int, error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 25
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize

	args := []interface{}{tenantID, positionCode}
	whereParts := []string{"tenant_id = $1", "position_code = $2"}
	argIndex := 3

	if len(opts.Filter.AssignmentTypes) > 0 {
		typesNormalized := normalizeAssignmentTypes(opts.Filter.AssignmentTypes)
		if len(typesNormalized) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("assignment_type = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(typesNormalized))
			argIndex++
		}
	}

	if opts.Filter.IncludeActingOnly {
		whereParts = append(whereParts, "assignment_type = 'ACTING'")
	}

	if opts.Filter.AssignmentStatus != nil {
		status := strings.ToUpper(strings.TrimSpace(*opts.Filter.AssignmentStatus))
		if status != "" {
			whereParts = append(whereParts, fmt.Sprintf("assignment_status = $%d", argIndex))
			args = append(args, status)
			argIndex++
		}
	}

	if opts.Filter.AsOfDate != nil {
		dateVal := opts.Filter.AsOfDate.Format("2006-01-02")
		whereParts = append(whereParts, fmt.Sprintf("(effective_date <= $%d AND (end_date IS NULL OR end_date >= $%d))", argIndex, argIndex))
		args = append(args, dateVal)
		argIndex++
	}

	if !opts.Filter.IncludeHistorical {
		whereParts = append(whereParts, "assignment_status <> 'ENDED'")
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM position_assignments %s`, whereClause)
	var total int
	if err := r.queryRow(ctx, tx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count position assignments: %w", err)
	}

	selectQuery := fmt.Sprintf(`
SELECT
	assignment_id,
	tenant_id,
	position_code,
	position_record_id,
	employee_id,
	employee_name,
	employee_number,
	assignment_type,
	assignment_status,
	fte,
	effective_date,
	end_date,
	acting_until,
	auto_revert,
	reminder_sent_at,
	is_current,
	notes,
	created_at,
	updated_at
FROM position_assignments
%s
ORDER BY effective_date DESC, created_at DESC
LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := r.queryRows(ctx, tx, selectQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query position assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]types.PositionAssignment, 0)
	for rows.Next() {
		var entity types.PositionAssignment
		if err := rows.Scan(
			&entity.AssignmentID,
			&entity.TenantID,
			&entity.PositionCode,
			&entity.PositionRecordID,
			&entity.EmployeeID,
			&entity.EmployeeName,
			&entity.EmployeeNumber,
			&entity.AssignmentType,
			&entity.AssignmentStatus,
			&entity.FTE,
			&entity.EffectiveDate,
			&entity.EndDate,
			&entity.ActingUntil,
			&entity.AutoRevert,
			&entity.ReminderSentAt,
			&entity.IsCurrent,
			&entity.Notes,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan position assignment row: %w", err)
		}
		assignments = append(assignments, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate position assignments: %w", err)
	}

	return assignments, total, nil
}

func (r *PositionAssignmentRepository) UpdateAssignment(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID, params types.AssignmentUpdateParams) error {
	setParts := make([]string, 0, 6)
	args := []interface{}{tenantID, assignmentID}
	argIndex := 3

	if params.FTE != nil {
		setParts = append(setParts, fmt.Sprintf("fte = $%d", argIndex))
		args = append(args, *params.FTE)
		argIndex++
	}

	if params.ActingUntil != nil {
		setParts = append(setParts, fmt.Sprintf("acting_until = $%d", argIndex))
		args = append(args, params.ActingUntil)
		argIndex++
	} else if params.ClearActingUntil {
		setParts = append(setParts, "acting_until = NULL")
	}

	if params.AutoRevert != nil {
		setParts = append(setParts, fmt.Sprintf("auto_revert = $%d", argIndex))
		args = append(args, *params.AutoRevert)
		argIndex++
	}

	if params.ReminderSentAt != nil {
		setParts = append(setParts, fmt.Sprintf("reminder_sent_at = $%d", argIndex))
		args = append(args, *params.ReminderSentAt)
		argIndex++
	} else if params.ClearReminderSent {
		setParts = append(setParts, "reminder_sent_at = NULL")
	}

	if params.Notes != nil {
		note := strings.TrimSpace(*params.Notes)
		if note == "" {
			setParts = append(setParts, "notes = NULL")
		} else {
			setParts = append(setParts, fmt.Sprintf("notes = $%d", argIndex))
			args = append(args, note)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf(`UPDATE position_assignments SET %s WHERE tenant_id = $1 AND assignment_id = $2`, strings.Join(setParts, ", "))

	result, err := r.exec(ctx, tx, query, args...)
	if err != nil {
		return fmt.Errorf("update position assignment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("assignment not found for update")
	}

	return nil
}

func (r *PositionAssignmentRepository) ListAutoRevertCandidates(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, asOf time.Time, limit int) ([]types.PositionAssignment, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `SELECT assignment_id, tenant_id, position_code, position_record_id, employee_id, employee_name, employee_number,
assignment_type, assignment_status, fte, effective_date, end_date, acting_until, auto_revert, reminder_sent_at, is_current, notes, created_at, updated_at
FROM position_assignments
WHERE tenant_id = $1
  AND assignment_type = 'ACTING'
  AND auto_revert = true
  AND assignment_status = 'ACTIVE'
  AND acting_until IS NOT NULL
  AND acting_until <= $2
ORDER BY acting_until ASC
LIMIT $3`

	rows, err := r.queryRows(ctx, tx, query, tenantID, asOf, limit)
	if err != nil {
		return nil, fmt.Errorf("query auto revert assignments: %w", err)
	}
	defer rows.Close()

	var result []types.PositionAssignment
	for rows.Next() {
		var entity types.PositionAssignment
		if err := rows.Scan(
			&entity.AssignmentID,
			&entity.TenantID,
			&entity.PositionCode,
			&entity.PositionRecordID,
			&entity.EmployeeID,
			&entity.EmployeeName,
			&entity.EmployeeNumber,
			&entity.AssignmentType,
			&entity.AssignmentStatus,
			&entity.FTE,
			&entity.EffectiveDate,
			&entity.EndDate,
			&entity.ActingUntil,
			&entity.AutoRevert,
			&entity.ReminderSentAt,
			&entity.IsCurrent,
			&entity.Notes,
			&entity.CreatedAt,
			&entity.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan auto revert assignment: %w", err)
		}
		result = append(result, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate auto revert assignments: %w", err)
	}

	return result, nil
}

func (r *PositionAssignmentRepository) SumActiveFTE(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string) (float64, error) {
	query := `SELECT COALESCE(SUM(fte), 0)
FROM position_assignments
WHERE tenant_id = $1 AND position_code = $2
  AND assignment_status = 'ACTIVE'
  AND is_current = true`

	var total sql.NullFloat64
	if err := r.queryRow(ctx, tx, query, tenantID, positionCode).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to sum active assignment FTE: %w", err)
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}
