package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"organization-command-service/internal/types"
	"organization-command-service/internal/utils"
)

func (r *OrganizationRepository) Update(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateOrganizationRequest) (*types.Organization, error) {
	setParts := make([]string, 0, 8)
	args := []interface{}{tenantID.String(), code}
	argIndex := 3

	addAssignment := func(column string, value interface{}) {
		placeholder := "$" + strconv.Itoa(argIndex)
		setParts = append(setParts, column+" = "+placeholder)
		args = append(args, value)
		argIndex++
	}

	var nameOverride *string
	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		addAssignment("name", trimmedName)
		nameOverride = &trimmedName
	}

	if req.UnitType != nil {
		addAssignment("unit_type", *req.UnitType)
	}

	if req.SortOrder != nil {
		addAssignment("sort_order", *req.SortOrder)
	}

	if req.Description != nil {
		addAssignment("description", *req.Description)
	}

	if req.ParentCode != nil {
		normalizedParent := utils.NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent

		fields, err := r.recalculateSelfHierarchy(ctx, tenantID, code, nil, normalizedParent, nameOverride)
		if err != nil {
			return nil, err
		}

		if normalizedParent != nil {
			addAssignment("parent_code", *normalizedParent)
		} else {
			addAssignment("parent_code", nil)
		}
		addAssignment("level", fields.Level)
		addAssignment("code_path", fields.CodePath)
		addAssignment("name_path", fields.NamePath)
	}

	if req.EffectiveDate != nil {
		addAssignment("effective_date", *req.EffectiveDate)
	}

	if req.EndDate != nil {
		addAssignment("end_date", *req.EndDate)
	}

	if req.ChangeReason != nil {
		addAssignment("change_reason", *req.ChangeReason)
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("无字段需要更新，操作被忽略")
	}

	addAssignment("updated_at", time.Now())
	setClause := strings.Join(setParts, ", ")

	query := fmt.Sprintf(`UPDATE organization_units
SET %s
WHERE tenant_id = $1 AND code = $2
  AND status <> 'DELETED'
RETURNING tenant_id, code, parent_code, name, unit_type, status,
          level, code_path, name_path, sort_order, description, created_at, updated_at,
          effective_date, end_date, change_reason`, setClause)

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.CodePath, &org.NamePath, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或已删除不可修改: %s", code)
		}
		return nil, fmt.Errorf("更新组织失败: %w", err)
	}

	r.logger.Printf("组织更新成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

func (r *OrganizationRepository) UpdateByRecordId(ctx context.Context, tenantID uuid.UUID, recordId string, req *types.UpdateOrganizationRequest) (*types.Organization, error) {
	setParts := make([]string, 0, 8)
	args := []interface{}{tenantID.String(), recordId}
	argIndex := 3

	addAssignment := func(column string, value interface{}) {
		placeholder := "$" + strconv.Itoa(argIndex)
		setParts = append(setParts, column+" = "+placeholder)
		args = append(args, value)
		argIndex++
	}

	var nameOverride *string
	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		addAssignment("name", trimmedName)
		nameOverride = &trimmedName
	}

	if req.UnitType != nil {
		addAssignment("unit_type", *req.UnitType)
	}

	if req.Status != nil {
		addAssignment("status", *req.Status)
	}

	if req.SortOrder != nil {
		addAssignment("sort_order", *req.SortOrder)
	}

	if req.Description != nil {
		addAssignment("description", *req.Description)
	}

	if req.ParentCode != nil {
		normalizedParent := utils.NormalizeParentCodePointer(req.ParentCode)
		req.ParentCode = normalizedParent

		fields, err := r.recalculateSelfHierarchy(ctx, tenantID, "", &recordId, normalizedParent, nameOverride)
		if err != nil {
			return nil, err
		}

		if normalizedParent != nil {
			addAssignment("parent_code", *normalizedParent)
		} else {
			addAssignment("parent_code", nil)
		}
		addAssignment("level", fields.Level)
		addAssignment("code_path", fields.CodePath)
		addAssignment("name_path", fields.NamePath)
	}

	if req.EffectiveDate != nil {
		addAssignment("effective_date", *req.EffectiveDate)
	}

	if req.EndDate != nil {
		addAssignment("end_date", *req.EndDate)
	}

	if req.ChangeReason != nil {
		addAssignment("change_reason", *req.ChangeReason)
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("无字段需要更新，操作被忽略")
	}

	addAssignment("updated_at", time.Now())
	setClause := strings.Join(setParts, ", ")

	query := fmt.Sprintf(`UPDATE organization_units
SET %s
WHERE tenant_id = $1 AND record_id = $2
  AND status <> 'DELETED'
RETURNING record_id, tenant_id, code, parent_code, name, unit_type, status,
          level, code_path, name_path, sort_order, description, created_at, updated_at,
          effective_date, end_date, change_reason`, setClause)

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.CodePath, &org.NamePath, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("记录不存在或已删除记录为只读: %s", recordId)
		}
		return nil, fmt.Errorf("更新历史记录失败: %w", err)
	}

	r.logger.Printf("历史记录更新成功: %s - %s (记录ID: %s)", org.Code, org.Name, recordId)
	return &org, nil
}
