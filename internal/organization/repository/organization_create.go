package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func (r *OrganizationRepository) GenerateCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	for nextCode := 1000000; nextCode <= 9999999; nextCode++ {
		candidateCode := fmt.Sprintf("%07d", nextCode)

		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM organization_units WHERE tenant_id = $1 AND code = $2)`
		err := r.db.QueryRowContext(ctx, checkQuery, tenantID.String(), candidateCode).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("检查代码唯一性失败: %w", err)
		}

		if !exists {
			return candidateCode, nil
		}
	}

	return "", fmt.Errorf("生成唯一组织代码失败：7位数编码已用尽")
}

func (r *OrganizationRepository) Create(ctx context.Context, org *types.Organization) (*types.Organization, error) {
	tenantUUID, err := uuid.Parse(org.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效的租户ID: %w", err)
	}

	fields, err := r.ComputeHierarchyForNew(ctx, tenantUUID, org.Code, org.ParentCode, org.Name)
	if err != nil {
		return nil, err
	}

	org.Level = fields.Level
	org.CodePath = fields.CodePath
	org.NamePath = fields.NamePath

	query := `
        INSERT INTO organization_units (
            tenant_id, code, parent_code, name, unit_type, status, 
            level, code_path, name_path, sort_order, description, created_at, updated_at,
            effective_date, end_date, change_reason, is_current
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING record_id, created_at, updated_at
    `

	var createdAt, updatedAt time.Time

	var effectiveDate *types.Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
	} else {
		now := time.Now()
		effectiveDate = types.NewDate(now.Year(), now.Month(), now.Day())
	}

	today := time.Now().Truncate(24 * time.Hour)
	effectiveDateTime := time.Date(
		effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(),
		0, 0, 0, 0, time.UTC,
	)
	isCurrent := !effectiveDateTime.After(today)

	err = r.db.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.CodePath,
		org.NamePath,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate,
		org.EndDate,
		org.ChangeReason,
		isCurrent,
	).Scan(&org.RecordID, &createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503":
				parent := ""
				if org.ParentCode != nil {
					parent = strings.TrimSpace(*org.ParentCode)
				}
				return nil, fmt.Errorf("父组织不存在: %s", parent)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate

	r.logger.Infof("组织创建成功: %s - %s", org.Code, org.Name)
	return org, nil
}

func (r *OrganizationRepository) CreateInTransaction(ctx context.Context, tx *sql.Tx, org *types.Organization) (*types.Organization, error) {
	query := `
        INSERT INTO organization_units (
            tenant_id, code, parent_code, name, unit_type, status,
            level, code_path, name_path, sort_order, description, created_at, updated_at,
            effective_date, end_date, change_reason, is_current
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING record_id, created_at, updated_at
    `

	var createdAt, updatedAt time.Time

	var effectiveDate *types.Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
	} else {
		now := time.Now()
		effectiveDate = types.NewDate(now.Year(), now.Month(), now.Day())
	}

	err := tx.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.CodePath,
		org.NamePath,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate,
		org.EndDate,
		org.ChangeReason,
		org.IsCurrent,
	).Scan(&org.RecordID, &createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503":
				parent := ""
				if org.ParentCode != nil {
					parent = strings.TrimSpace(*org.ParentCode)
				}
				return nil, fmt.Errorf("父组织不存在: %s", parent)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate

	r.logger.Infof("时态组织创建成功: %s - %s (生效日期: %v, 当前: %v)",
		org.Code, org.Name,
		org.EffectiveDate.String(),
		org.IsCurrent)
	return org, nil
}
