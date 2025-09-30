package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"organization-command-service/internal/types"
)

func (r *OrganizationRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error) {
	query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
               level, path, code_path, name_path, sort_order, description, created_at, updated_at,
               effective_date, end_date, change_reason
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true
        LIMIT 1
    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &parentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在: %s", code)
		}
		return nil, fmt.Errorf("查询组织失败: %w", err)
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

	return &org, nil
}

func (r *OrganizationRepository) GetByRecordId(ctx context.Context, tenantID uuid.UUID, recordId string) (*types.Organization, error) {
	query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
               level, path, code_path, name_path, sort_order, description, created_at, updated_at,
               effective_date, end_date, change_reason
        FROM organization_units
        WHERE tenant_id = $1 AND record_id = $2
        LIMIT 1
    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), recordId).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &parentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("记录不存在: %s", recordId)
		}
		return nil, fmt.Errorf("查询组织记录失败: %w", err)
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

	return &org, nil
}

func (r *OrganizationRepository) ListVersionsByCode(ctx context.Context, tenantID uuid.UUID, code string) ([]types.Organization, error) {
	query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
               level, path, sort_order, description, created_at, updated_at,
               effective_date, end_date, change_reason
        FROM organization_units
        WHERE tenant_id = $1 AND code = $2
          AND status <> 'DELETED'
        ORDER BY effective_date DESC
    `

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code)
	if err != nil {
		return nil, fmt.Errorf("查询组织版本失败: %w", err)
	}
	defer rows.Close()

	versions := make([]types.Organization, 0, 8)
	for rows.Next() {
		var org types.Organization
		var parentCode sql.NullString
		var effectiveDate, endDate sql.NullTime
		var changeReason sql.NullString

		if err := rows.Scan(
			&org.RecordID, &org.TenantID, &org.Code, &parentCode, &org.Name,
			&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
			&org.Description, &org.CreatedAt, &org.UpdatedAt,
			&effectiveDate, &endDate, &changeReason,
		); err != nil {
			return nil, fmt.Errorf("扫描组织版本失败: %w", err)
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

		versions = append(versions, org)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历组织版本失败: %w", err)
	}

	return versions, nil
}
