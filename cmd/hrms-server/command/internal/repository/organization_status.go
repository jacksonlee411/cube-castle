package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"cube-castle/internal/types"
)

func (r *OrganizationRepository) Suspend(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
	query := `
        UPDATE organization_units 
        SET status = 'INACTIVE', updated_at = $3
        WHERE tenant_id = $1 AND code = $2 AND status = 'ACTIVE'
	RETURNING tenant_id, code, parent_code, name, unit_type, status,
	         level, code_path, name_path, sort_order, description, created_at, updated_at,
         effective_date, end_date, change_reason
    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.CodePath, &org.NamePath, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是ACTIVE: %s", code)
		}
		return nil, fmt.Errorf("停用组织失败: %w", err)
	}

	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		org.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
	}
	if endDate.Valid {
		org.EndDate = types.NewDateFromTime(endDate.Time)
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}

	r.logger.Printf("组织停用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

func (r *OrganizationRepository) Activate(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
	query := `
			UPDATE organization_units 
			SET status = 'ACTIVE', updated_at = $3
			WHERE tenant_id = $1 AND code = $2 AND status = 'INACTIVE'
	RETURNING tenant_id, code, parent_code, name, unit_type, status,
	         level, code_path, name_path, sort_order, description, created_at, updated_at,
         effective_date, end_date, change_reason
	    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.CodePath, &org.NamePath, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是INACTIVE: %s", code)
		}
		return nil, fmt.Errorf("重新启用组织失败: %w", err)
	}

	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		org.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
	}
	if endDate.Valid {
		org.EndDate = types.NewDateFromTime(endDate.Time)
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}

	r.logger.Printf("组织重新启用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

func (r *OrganizationRepository) CountNonDeletedChildren(ctx context.Context, tenantID uuid.UUID, code string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT code)
		FROM organization_units
		WHERE tenant_id = $1 AND parent_code = $2 AND status <> 'DELETED'
		  AND (is_current = true OR effective_date >= CURRENT_DATE)
	`

	var count int
	if err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count child organizations: %w", err)
	}

	return count, nil
}

func (r *OrganizationRepository) SoftDeleteOrganization(ctx context.Context, tenantID uuid.UUID, code string, deletedAt time.Time, actorID, reason string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction failed: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `
		UPDATE organization_units
		SET status = 'DELETED',
		    is_current = false,
		    updated_at = NOW(),
		    deleted_at = $3,
		    deleted_by = $4,
		    deletion_reason = CASE WHEN $5 <> '' THEN $5 ELSE deletion_reason END
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED'
	`

	res, err := tx.ExecContext(ctx, updateQuery, tenantID.String(), code, deletedAt, actorID, strings.TrimSpace(reason))
	if err != nil {
		return fmt.Errorf("soft delete organization failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("retrieve delete row count failed: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit soft delete transaction failed: %w", err)
	}

	r.logger.Printf("🗑️ 已软删除组织 %s (tenant=%s, rows=%d)", code, tenantID, rowsAffected)
	return nil
}

func (r *OrganizationRepository) HasOtherNonDeletedVersions(ctx context.Context, tenantID uuid.UUID, code, excludeRecordID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2 AND status <> 'DELETED' AND record_id <> $3
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, excludeRecordID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to count remaining versions: %w", err)
	}
	return count > 0, nil
}
