package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"organization-command-service/internal/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type OrganizationRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationRepository(db *sql.DB, logger *log.Logger) *OrganizationRepository {
	return &OrganizationRepository{db: db, logger: logger}
}

func (r *OrganizationRepository) GenerateCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// 获取当前最大的数字代码
	query := `
		SELECT COALESCE(MAX(CAST(code AS INTEGER)), 1000000) as max_code
		FROM organization_units 
		WHERE tenant_id = $1 AND code ~ '^[0-9]{7}$'
	`

	var maxCode int
	err := r.db.QueryRowContext(ctx, query, tenantID.String()).Scan(&maxCode)
	if err != nil {
		return "", fmt.Errorf("获取最大组织代码失败: %w", err)
	}

	// 从最大值+1开始，寻找可用的代码
	for nextCode := maxCode + 1; nextCode <= maxCode + 100; nextCode++ {
		candidateCode := fmt.Sprintf("%07d", nextCode)
		
		// 检查代码是否已存在
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM organization_units WHERE tenant_id = $1 AND code = $2)`
		err = r.db.QueryRowContext(ctx, checkQuery, tenantID.String(), candidateCode).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("检查代码唯一性失败: %w", err)
		}
		
		if !exists {
			return candidateCode, nil
		}
	}
	
	return "", fmt.Errorf("生成唯一组织代码失败：尝试100次后仍未找到可用代码")
}

func (r *OrganizationRepository) Create(ctx context.Context, org *types.Organization) (*types.Organization, error) {
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time

	// 确保effective_date始终有值（数据库约束要求）
	var effectiveDate *types.Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
	} else {
		now := time.Now()
		effectiveDate = types.NewDate(now.Year(), now.Month(), now.Day())
	}

	err := r.db.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.Path,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate, // Date类型
		org.EndDate,   // 允许为nil
		org.IsTemporal,
		org.ChangeReason,
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503": // foreign key violation
				return nil, fmt.Errorf("父组织不存在: %s", *org.ParentCode)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate // 确保返回的组织有effective_date值

	r.logger.Printf("组织创建成功: %s - %s (时态: %v)", org.Code, org.Name, org.IsTemporal)
	return org, nil
}

func (r *OrganizationRepository) CreateInTransaction(ctx context.Context, tx *sql.Tx, org *types.Organization) (*types.Organization, error) {
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING created_at, updated_at
	`

	var createdAt, updatedAt time.Time

	// 确保effective_date始终有值（数据库约束要求）
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
		org.Path,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate, // Date类型
		org.EndDate,   // 允许为nil
		org.IsTemporal,
		org.ChangeReason,
		org.IsCurrent, // 显式设置is_current
	).Scan(&createdAt, &updatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique violation
				return nil, fmt.Errorf("组织代码已存在: %s", org.Code)
			case "23503": // foreign key violation
				return nil, fmt.Errorf("父组织不存在: %s", *org.ParentCode)
			}
		}
		return nil, fmt.Errorf("创建组织失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate // 确保返回的组织有effective_date值

	r.logger.Printf("时态组织创建成功: %s - %s (生效日期: %v, 当前: %v)",
		org.Code, org.Name,
		org.EffectiveDate.String(),
		org.IsCurrent)
	return org, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, tenantID uuid.UUID, code string, req *types.UpdateOrganizationRequest) (*types.Organization, error) {
	// 构建动态更新查询
	setParts := []string{}
	args := []interface{}{tenantID.String(), code}
	argIndex := 3

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.UnitType != nil {
		setParts = append(setParts, fmt.Sprintf("unit_type = $%d", argIndex))
		args = append(args, *req.UnitType)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.ParentCode != nil {
		setParts = append(setParts, fmt.Sprintf("parent_code = $%d", argIndex))
		args = append(args, *req.ParentCode)
		argIndex++
	}

	// 时态管理字段更新
	if req.EffectiveDate != nil {
		setParts = append(setParts, fmt.Sprintf("effective_date = $%d", argIndex))
		args = append(args, *req.EffectiveDate)
		argIndex++
	}

	if req.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}

	if req.IsTemporal != nil {
		setParts = append(setParts, fmt.Sprintf("is_temporal = $%d", argIndex))
		args = append(args, *req.IsTemporal)
		argIndex++
	}

	if req.ChangeReason != nil {
		setParts = append(setParts, fmt.Sprintf("change_reason = $%d", argIndex))
		args = append(args, *req.ChangeReason)
		argIndex++
	}

	if len(setParts) == 0 {
		// 无字段需要更新，返回错误
		return nil, fmt.Errorf("无字段需要更新，操作被忽略")
	}

	// 添加updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s
		WHERE tenant_id = $1 AND code = $2
		RETURNING tenant_id, code, parent_code, name, unit_type, status,
		          level, path, sort_order, description, created_at, updated_at,
		          effective_date, end_date, is_temporal, change_reason
	`, strings.Join(setParts, ", "))

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在: %s", code)
		}
		return nil, fmt.Errorf("更新组织失败: %w", err)
	}

	r.logger.Printf("组织更新成功: %s - %s (时态: %v)", org.Code, org.Name, org.IsTemporal)
	return &org, nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, tenantID uuid.UUID, code string) error {
	// 软删除 - 设置状态为DELETED
	query := `
		UPDATE organization_units 
		SET status = 'DELETED', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status != 'DELETED'
	`

	result, err := r.db.ExecContext(ctx, query, tenantID.String(), code, time.Now())
	if err != nil {
		return fmt.Errorf("删除组织失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除结果失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("组织不存在或已删除: %s", code)
	}

	r.logger.Printf("组织删除成功: %s", code)
	return nil
}

func (r *OrganizationRepository) Suspend(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
	query := `
		UPDATE organization_units 
		SET status = 'INACTIVE', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status = 'ACTIVE'
		RETURNING tenant_id, code, parent_code, name, unit_type, status, 
		         level, path, sort_order, description, created_at, updated_at,
		         effective_date, end_date, is_temporal, change_reason
	`

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &org.IsTemporal, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是ACTIVE: %s", code)
		}
		return nil, fmt.Errorf("停用组织失败: %w", err)
	}

	// 处理可空字段
	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		d := &types.Date{effectiveDate.Time}
		org.EffectiveDate = d
	}
	if endDate.Valid {
		d := &types.Date{endDate.Time}
		org.EndDate = d
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}

	r.logger.Printf("组织停用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

func (r *OrganizationRepository) Reactivate(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
	query := `
		UPDATE organization_units 
		SET status = 'ACTIVE', updated_at = $3
		WHERE tenant_id = $1 AND code = $2 AND status = 'INACTIVE'
		RETURNING tenant_id, code, parent_code, name, unit_type, status, 
		         level, path, sort_order, description, created_at, updated_at,
		         effective_date, end_date, is_temporal, change_reason
	`

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &org.IsTemporal, &changeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在或状态不是INACTIVE: %s", code)
		}
		return nil, fmt.Errorf("重新启用组织失败: %w", err)
	}

	// 处理可空字段
	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		d := &types.Date{effectiveDate.Time}
		org.EffectiveDate = d
	}
	if endDate.Valid {
		d := &types.Date{endDate.Time}
		org.EndDate = d
	}
	if changeReason.Valid {
		org.ChangeReason = &changeReason.String
	}

	r.logger.Printf("组织重新启用成功: %s - %s", org.Code, org.Name)
	return &org, nil
}

func (r *OrganizationRepository) CalculatePath(ctx context.Context, tenantID uuid.UUID, parentCode *string, code string) (string, int, error) {
	if parentCode == nil {
		return "/" + code, 1, nil
	}

	query := `
		SELECT path, level 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2
	`

	var parentPath string
	var parentLevel int

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), *parentCode).Scan(&parentPath, &parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("父组织不存在: %s", *parentCode)
		}
		return "", 0, fmt.Errorf("查询父组织失败: %w", err)
	}

	path := parentPath + "/" + code
	level := parentLevel + 1

	return path, level, nil
}

// UpdateByRecordId 通过UUID更新历史记录
func (r *OrganizationRepository) UpdateByRecordId(ctx context.Context, tenantID uuid.UUID, recordId string, req *types.UpdateOrganizationRequest) (*types.Organization, error) {
	// 构建动态更新查询
	setParts := []string{}
	args := []interface{}{tenantID.String(), recordId}
	argIndex := 3

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.UnitType != nil {
		setParts = append(setParts, fmt.Sprintf("unit_type = $%d", argIndex))
		args = append(args, *req.UnitType)
		argIndex++
	}

	if req.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *req.Status)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.ParentCode != nil {
		setParts = append(setParts, fmt.Sprintf("parent_code = $%d", argIndex))
		args = append(args, *req.ParentCode)
		argIndex++
	}

	// 时态管理字段更新
	if req.EffectiveDate != nil {
		setParts = append(setParts, fmt.Sprintf("effective_date = $%d", argIndex))
		args = append(args, *req.EffectiveDate)
		argIndex++
	}

	if req.EndDate != nil {
		setParts = append(setParts, fmt.Sprintf("end_date = $%d", argIndex))
		args = append(args, *req.EndDate)
		argIndex++
	}

	if req.IsTemporal != nil {
		setParts = append(setParts, fmt.Sprintf("is_temporal = $%d", argIndex))
		args = append(args, *req.IsTemporal)
		argIndex++
	}

	if req.ChangeReason != nil {
		setParts = append(setParts, fmt.Sprintf("change_reason = $%d", argIndex))
		args = append(args, *req.ChangeReason)
		argIndex++
	}

	if len(setParts) == 0 {
		// 无字段需要更新
		return nil, fmt.Errorf("无字段需要更新，操作被忽略")
	}

	// 添加updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s
		WHERE tenant_id = $1 AND record_id = $2
		RETURNING tenant_id, code, parent_code, name, unit_type, status,
		          level, path, sort_order, description, created_at, updated_at,
		          effective_date, end_date, is_temporal, change_reason
	`, strings.Join(setParts, ", "))

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("记录不存在: %s", recordId)
		}
		return nil, fmt.Errorf("更新历史记录失败: %w", err)
	}

	r.logger.Printf("历史记录更新成功: %s - %s (记录ID: %s)", org.Code, org.Name, recordId)
	return &org, nil
}
