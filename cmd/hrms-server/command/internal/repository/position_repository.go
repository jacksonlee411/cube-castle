package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"cube-castle/internal/types"
)

type PositionRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewPositionRepository(db *sql.DB, logger *log.Logger) *PositionRepository {
	return &PositionRepository{db: db, logger: logger}
}

func (r *PositionRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func (r *PositionRepository) queryRow(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return r.db.QueryRowContext(ctx, query, args...)
}

func (r *PositionRepository) queryRows(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return r.db.QueryContext(ctx, query, args...)
}

func (r *PositionRepository) exec(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

// GetCurrentPosition 返回当前版本的职位
func (r *PositionRepository) GetCurrentPosition(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.Position, error) {
	query := `SELECT record_id, tenant_id, code, title, job_profile_code, job_profile_name, job_family_group_code, job_family_group_name, job_family_group_record_id,
job_family_code, job_family_name, job_family_record_id, job_role_code, job_role_name, job_role_record_id,
job_level_code, job_level_name, job_level_record_id, organization_code, organization_name, position_type, status, employment_type,
	headcount_capacity, headcount_in_use, grade_level, cost_center_code,
reports_to_position_code, profile, effective_date, end_date, is_current, created_at, updated_at, deleted_at, operation_type, operated_by_id, operated_by_name, operation_reason
FROM positions WHERE tenant_id = $1 AND code = $2 AND is_current = true LIMIT 1`

	var entity types.Position
	err := r.queryRow(ctx, tx, query, tenantID, code).Scan(
		&entity.RecordID,
		&entity.TenantID,
		&entity.Code,
		&entity.Title,
		&entity.JobProfileCode,
		&entity.JobProfileName,
		&entity.JobFamilyGroupCode,
		&entity.JobFamilyGroupName,
		&entity.JobFamilyGroupRecord,
		&entity.JobFamilyCode,
		&entity.JobFamilyName,
		&entity.JobFamilyRecord,
		&entity.JobRoleCode,
		&entity.JobRoleName,
		&entity.JobRoleRecord,
		&entity.JobLevelCode,
		&entity.JobLevelName,
		&entity.JobLevelRecord,
		&entity.OrganizationCode,
		&entity.OrganizationName,
		&entity.PositionType,
		&entity.Status,
		&entity.EmploymentType,
		&entity.HeadcountCapacity,
		&entity.HeadcountInUse,
		&entity.GradeLevel,
		&entity.CostCenterCode,
		&entity.ReportsToPosition,
		&entity.Profile,
		&entity.EffectiveDate,
		&entity.EndDate,
		&entity.IsCurrent,
		&entity.CreatedAt,
		&entity.UpdatedAt,
		&entity.DeletedAt,
		&entity.OperationType,
		&entity.OperatedByID,
		&entity.OperatedByName,
		&entity.OperationReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query current position: %w", err)
	}
	return &entity, nil
}

// InsertPositionVersion 插入职位版本（用于创建和新增版本）
func (r *PositionRepository) InsertPositionVersion(ctx context.Context, tx *sql.Tx, entity *types.Position) (*types.Position, error) {
	query := `INSERT INTO positions (
tenant_id, code, title, job_profile_code, job_profile_name,
job_family_group_code, job_family_group_name, job_family_group_record_id,
job_family_code, job_family_name, job_family_record_id,
job_role_code, job_role_name, job_role_record_id,
job_level_code, job_level_name, job_level_record_id,
organization_code, organization_name, position_type, status, employment_type,
headcount_capacity, headcount_in_use, grade_level, cost_center_code,
reports_to_position_code, profile, effective_date, end_date, is_current,
created_at, updated_at, deleted_at, operation_type, operated_by_id, operated_by_name, operation_reason)
VALUES (
$1,$2,$3,$4,$5,
$6,$7,$8,
$9,$10,$11,
$12,$13,$14,
$15,$16,$17,
$18,$19,$20,$21,$22,
$23,$24,$25,$26,
$27,$28,$29,$30,$31,
NOW(),NOW(),NULL,$32,$33,$34,$35)
RETURNING record_id, created_at, updated_at`

	var profilePayload interface{}
	if len(entity.Profile) > 0 {
		profilePayload = entity.Profile
	} else {
		profilePayload = []byte("{}")
	}

	var reportsTo interface{}
	if entity.ReportsToPosition.Valid {
		reportsTo = entity.ReportsToPosition.String
	} else {
		reportsTo = nil
	}

	var organizationName interface{}
	if entity.OrganizationName.Valid {
		organizationName = entity.OrganizationName.String
	} else {
		organizationName = nil
	}

	var jobProfileCode interface{}
	if entity.JobProfileCode.Valid {
		jobProfileCode = entity.JobProfileCode.String
	} else {
		jobProfileCode = nil
	}

	var jobProfileName interface{}
	if entity.JobProfileName.Valid {
		jobProfileName = entity.JobProfileName.String
	} else {
		jobProfileName = nil
	}

	var gradeLevel interface{}
	if entity.GradeLevel.Valid {
		gradeLevel = entity.GradeLevel.String
	} else {
		gradeLevel = nil
	}

	var costCenter interface{}
	if entity.CostCenterCode.Valid {
		costCenter = entity.CostCenterCode.String
	} else {
		costCenter = nil
	}

	var endDate interface{}
	if entity.EndDate.Valid {
		endDate = entity.EndDate.Time
	} else {
		endDate = nil
	}

	var operationReason interface{}
	if entity.OperationReason.Valid {
		operationReason = entity.OperationReason.String
	} else {
		operationReason = nil
	}

	err := r.queryRow(ctx, tx, query,
		entity.TenantID,
		entity.Code,
		entity.Title,
		jobProfileCode,
		jobProfileName,
		entity.JobFamilyGroupCode,
		entity.JobFamilyGroupName,
		entity.JobFamilyGroupRecord,
		entity.JobFamilyCode,
		entity.JobFamilyName,
		entity.JobFamilyRecord,
		entity.JobRoleCode,
		entity.JobRoleName,
		entity.JobRoleRecord,
		entity.JobLevelCode,
		entity.JobLevelName,
		entity.JobLevelRecord,
		entity.OrganizationCode,
		organizationName,
		entity.PositionType,
		entity.Status,
		entity.EmploymentType,
		entity.HeadcountCapacity,
		entity.HeadcountInUse,
		gradeLevel,
		costCenter,
		reportsTo,
		profilePayload,
		entity.EffectiveDate,
		endDate,
		entity.IsCurrent,
		entity.OperationType,
		entity.OperatedByID,
		entity.OperatedByName,
		operationReason,
	).Scan(&entity.RecordID, &entity.CreatedAt, &entity.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("position version already exists for effective date")
			case "23503":
				return nil, fmt.Errorf("invalid foreign key reference: %s", pqErr.Constraint)
			}
		}
		return nil, fmt.Errorf("failed to insert position: %w", err)
	}

	return entity, nil
}

// RecalculatePositionTimeline 重算职位时间线，确保 end_date 和 is_current 正确
func (r *PositionRepository) RecalculatePositionTimeline(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) error {
	query := `SELECT record_id, effective_date, end_date, is_current FROM positions WHERE tenant_id = $1 AND code = $2 ORDER BY effective_date FOR UPDATE`
	rows, err := r.queryRows(ctx, tx, query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load position timeline: %w", err)
	}
	defer rows.Close()

	var timeline []temporalRow
	for rows.Next() {
		var row temporalRow
		if err := rows.Scan(&row.RecordID, &row.EffectiveDate, &row.EndDate, &row.IsCurrent); err != nil {
			return fmt.Errorf("failed to scan position timeline: %w", err)
		}
		timeline = append(timeline, row)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("position timeline iteration error: %w", err)
	}

	normalized := normalizeTemporal(timeline)
	return r.applyTemporalUpdates(ctx, tx, "positions", normalized)
}

func (r *PositionRepository) applyTemporalUpdates(ctx context.Context, tx *sql.Tx, table string, updates []temporalRow) error {
	if len(updates) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE %s SET end_date = $2, is_current = $3, updated_at = NOW() WHERE record_id = $1`, table)
	for _, row := range updates {
		var endDate interface{}
		if row.EndDate.Valid {
			endDate = row.EndDate.Time
		} else {
			endDate = nil
		}
		if _, err := r.exec(ctx, tx, query, row.RecordID, endDate, row.IsCurrent); err != nil {
			return fmt.Errorf("failed to update temporal row in %s: %w", table, err)
		}
	}
	return nil
}

// UpdatePositionHeadcount 更新职位的 headcount 占用情况
func (r *PositionRepository) UpdatePositionHeadcount(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID, headcountInUse float64, status string, operationType, operatedByName string, operatedByID uuid.UUID, reason *string) error {
	query := `UPDATE positions SET headcount_in_use = $1,
status = $2, operation_type = $3, operated_by_id = $4, operated_by_name = $5, operation_reason = $6, updated_at = NOW()
WHERE tenant_id = $7 AND record_id = $8`

	var reasonVal interface{}
	if reason != nil {
		reasonVal = *reason
	} else {
		reasonVal = nil
	}

	_, err := r.exec(ctx, tx, query,
		headcountInUse,
		status,
		operationType,
		operatedByID,
		operatedByName,
		reasonVal,
		tenantID,
		recordID,
	)

	if err != nil {
		return fmt.Errorf("failed to update position headcount: %w", err)
	}
	return nil
}

// UpdatePositionStatus 更新职位状态（用于 Vacate/Suspend/Activate 等事件）
func (r *PositionRepository) UpdatePositionStatus(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID, status string, payload map[string]interface{}, operationType string, operatedByName string, operatedByID uuid.UUID, reason *string) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal status payload: %w", err)
	}

	query := `UPDATE positions SET status = $1, profile = COALESCE(profile, '{}'::jsonb) || $2::jsonb, operation_type = $3, operated_by_id = $4, operated_by_name = $5, operation_reason = $6, updated_at = NOW()
WHERE tenant_id = $7 AND record_id = $8`

	var reasonVal interface{}
	if reason != nil {
		reasonVal = *reason
	} else {
		reasonVal = nil
	}

	if _, err := r.exec(ctx, tx, query,
		status,
		payloadJSON,
		operationType,
		operatedByID,
		operatedByName,
		reasonVal,
		tenantID,
		recordID,
	); err != nil {
		return fmt.Errorf("failed to update position status: %w", err)
	}
	return nil
}

// DeletePositionVersion 标记职位为删除状态
func (r *PositionRepository) DeletePositionVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID, operatedBy uuid.UUID, operatedByName string, reason *string) error {
	query := `UPDATE positions SET status = 'DELETED', deleted_at = NOW(), is_current = false, operation_type = 'DELETE', operated_by_id = $1, operated_by_name = $2, operation_reason = $3, updated_at = NOW() WHERE tenant_id = $4 AND record_id = $5`

	var reasonVal interface{}
	if reason != nil {
		reasonVal = *reason
	} else {
		reasonVal = nil
	}

	if _, err := r.exec(ctx, tx, query, operatedBy, operatedByName, reasonVal, tenantID, recordID); err != nil {
		return fmt.Errorf("failed to delete position version: %w", err)
	}
	return nil
}

// UpdatePositionDetails 更新职位当前版本的核心字段
func (r *PositionRepository) UpdatePositionDetails(ctx context.Context, tx *sql.Tx, entity *types.Position) (*types.Position, error) {
	query := `UPDATE positions SET
title = $1,
job_profile_code = $2,
job_profile_name = $3,
job_family_group_code = $4,
job_family_group_name = $5,
job_family_group_record_id = $6,
job_family_code = $7,
job_family_name = $8,
job_family_record_id = $9,
job_role_code = $10,
job_role_name = $11,
job_role_record_id = $12,
job_level_code = $13,
job_level_name = $14,
job_level_record_id = $15,
organization_code = $16,
organization_name = $17,
position_type = $18,
employment_type = $19,
grade_level = $20,
headcount_capacity = $21,
reports_to_position_code = $22,
operation_type = $23,
operated_by_id = $24,
operated_by_name = $25,
operation_reason = $26,
status = $27,
updated_at = NOW()
WHERE tenant_id = $28 AND record_id = $29
RETURNING updated_at`

	var jobProfileCode interface{}
	if entity.JobProfileCode.Valid {
		jobProfileCode = entity.JobProfileCode.String
	}
	var jobProfileName interface{}
	if entity.JobProfileName.Valid {
		jobProfileName = entity.JobProfileName.String
	}
	var organizationName interface{}
	if entity.OrganizationName.Valid {
		organizationName = entity.OrganizationName.String
	}
	var gradeLevel interface{}
	if entity.GradeLevel.Valid {
		gradeLevel = entity.GradeLevel.String
	}
	var reportsTo interface{}
	if entity.ReportsToPosition.Valid {
		reportsTo = entity.ReportsToPosition.String
	}
	var operationReason interface{}
	if entity.OperationReason.Valid {
		operationReason = entity.OperationReason.String
	}

	err := r.queryRow(ctx, tx, query,
		entity.Title,
		jobProfileCode,
		jobProfileName,
		entity.JobFamilyGroupCode,
		entity.JobFamilyGroupName,
		entity.JobFamilyGroupRecord,
		entity.JobFamilyCode,
		entity.JobFamilyName,
		entity.JobFamilyRecord,
		entity.JobRoleCode,
		entity.JobRoleName,
		entity.JobRoleRecord,
		entity.JobLevelCode,
		entity.JobLevelName,
		entity.JobLevelRecord,
		entity.OrganizationCode,
		organizationName,
		entity.PositionType,
		entity.EmploymentType,
		gradeLevel,
		entity.HeadcountCapacity,
		reportsTo,
		entity.OperationType,
		entity.OperatedByID,
		entity.OperatedByName,
		operationReason,
		entity.Status,
		entity.TenantID,
		entity.RecordID,
	).Scan(&entity.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, fmt.Errorf("invalid foreign key reference during position update: %s", pqErr.Constraint)
		}
		return nil, fmt.Errorf("failed to update position: %w", err)
	}
	return entity, nil
}

// UpdatePositionOrganization 更新职位的组织归属
func (r *PositionRepository) UpdatePositionOrganization(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID, organizationCode string, organizationName *string, status string, operationType string, operatedByID uuid.UUID, operatedByName string, reason *string) error {
	query := `UPDATE positions SET organization_code = $1, organization_name = $2, status = $3, operation_type = $4, operated_by_id = $5, operated_by_name = $6, operation_reason = $7, updated_at = NOW()
WHERE tenant_id = $8 AND record_id = $9`

	var orgName interface{}
	if organizationName != nil && *organizationName != "" {
		orgName = *organizationName
	} else {
		orgName = nil
	}

	var reasonVal interface{}
	if reason != nil {
		reasonVal = *reason
	} else {
		reasonVal = nil
	}

	if _, err := r.exec(ctx, tx, query,
		organizationCode,
		orgName,
		status,
		operationType,
		operatedByID,
		operatedByName,
		reasonVal,
		tenantID,
		recordID,
	); err != nil {
		return fmt.Errorf("failed to update position organization: %w", err)
	}
	return nil
}

// GetPositionByRecordID 查询指定版本
func (r *PositionRepository) GetPositionByRecordID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID) (*types.Position, error) {
	query := `SELECT record_id, tenant_id, code, title, job_profile_code, job_profile_name, job_family_group_code, job_family_group_name, job_family_group_record_id,
job_family_code, job_family_name, job_family_record_id, job_role_code, job_role_name, job_role_record_id,
job_level_code, job_level_name, job_level_record_id, organization_code, organization_name, position_type, status, employment_type,
headcount_capacity, headcount_in_use, grade_level, cost_center_code,
reports_to_position_code, profile, effective_date, end_date, is_current, created_at, updated_at, deleted_at, operation_type, operated_by_id, operated_by_name, operation_reason
FROM positions WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`

	var entity types.Position
	err := r.queryRow(ctx, tx, query, tenantID, recordID).Scan(
		&entity.RecordID,
		&entity.TenantID,
		&entity.Code,
		&entity.Title,
		&entity.JobProfileCode,
		&entity.JobProfileName,
		&entity.JobFamilyGroupCode,
		&entity.JobFamilyGroupName,
		&entity.JobFamilyGroupRecord,
		&entity.JobFamilyCode,
		&entity.JobFamilyName,
		&entity.JobFamilyRecord,
		&entity.JobRoleCode,
		&entity.JobRoleName,
		&entity.JobRoleRecord,
		&entity.JobLevelCode,
		&entity.JobLevelName,
		&entity.JobLevelRecord,
		&entity.OrganizationCode,
		&entity.OrganizationName,
		&entity.PositionType,
		&entity.Status,
		&entity.EmploymentType,
		&entity.HeadcountCapacity,
		&entity.HeadcountInUse,
		&entity.GradeLevel,
		&entity.CostCenterCode,
		&entity.ReportsToPosition,
		&entity.Profile,
		&entity.EffectiveDate,
		&entity.EndDate,
		&entity.IsCurrent,
		&entity.CreatedAt,
		&entity.UpdatedAt,
		&entity.DeletedAt,
		&entity.OperationType,
		&entity.OperatedByID,
		&entity.OperatedByName,
		&entity.OperationReason,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load position by record id: %w", err)
	}
	return &entity, nil
}

// GenerateCode 生成职位编码
func (r *PositionRepository) GenerateCode(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID) (string, error) {
	const minCode = 1000000
	const maxCode = 9999999
	for next := minCode; next <= maxCode; next++ {
		candidate := fmt.Sprintf("P%07d", next)
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM positions WHERE tenant_id = $1 AND code = $2)`
		if err := r.queryRow(ctx, tx, query, tenantID, candidate).Scan(&exists); err != nil {
			return "", fmt.Errorf("failed to check position code uniqueness: %w", err)
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to generate unique position code: exhausted available range")
}
