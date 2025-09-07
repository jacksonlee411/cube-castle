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
	"organization-command-service/internal/types"
)

type OrganizationRepository struct {
	db     *sql.DB
	logger *log.Logger
}

func NewOrganizationRepository(db *sql.DB, logger *log.Logger) *OrganizationRepository {
	return &OrganizationRepository{db: db, logger: logger}
}

func (r *OrganizationRepository) GenerateCode(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// 从1000000开始寻找第一个可用的7位数代码 - 修复：直接搜索而非依赖MAX
	for nextCode := 1000000; nextCode <= 9999999; nextCode++ {
		candidateCode := fmt.Sprintf("%07d", nextCode)

		// 检查代码是否已存在
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
	query := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING record_id, created_at, updated_at
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

	// 计算is_current: 只有当effective_date <= 今天时才是current
	today := time.Now().Truncate(24 * time.Hour)
	effectiveDateTime := time.Date(
		effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(),
		0, 0, 0, 0, time.UTC,
	)
	isCurrent := !effectiveDateTime.After(today)

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
		isCurrent, // 根据effective_date计算的is_current值
	).Scan(&org.RecordID, &createdAt, &updatedAt)

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
		RETURNING record_id, created_at, updated_at
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
	).Scan(&org.RecordID, &createdAt, &updatedAt)

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

// CreateTemporalVersion 创建时态版本 - 专门处理版本插入和日期调整
func (r *OrganizationRepository) CreateTemporalVersion(ctx context.Context, org *types.Organization) (*types.Organization, error) {
	// 开始数据库事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 确保effective_date始终有值
	var effectiveDate *types.Date
	if org.EffectiveDate != nil {
		effectiveDate = org.EffectiveDate
	} else {
		now := time.Now()
		effectiveDate = types.NewDate(now.Year(), now.Month(), now.Day())
	}

	// 计算is_current: 只有当effective_date <= 今天时才是current
	today := time.Now().Truncate(24 * time.Hour)
	effectiveDateTime := time.Date(
		effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(),
		0, 0, 0, 0, time.UTC,
	)
	isCurrent := !effectiveDateTime.After(today)

	r.logger.Printf("🔄 开始创建时态版本: %s, 生效日期: %s", org.Code, effectiveDate.String())

	// 第一步：将该组织的所有记录设为非当前状态 (解决uk_current_organization约束)
	// 修复：移除status != 'DELETED'条件，确保所有is_current=true的记录都被清除
    clearCurrentQuery := `
        UPDATE organization_units 
        SET is_current = false,
            updated_at = NOW()
        WHERE code = $1 
          AND tenant_id = $2
          AND is_current = true
          AND status != 'DELETED' AND deleted_at IS NULL
    `
	
	_, err = tx.ExecContext(ctx, clearCurrentQuery, org.Code, org.TenantID)
	if err != nil {
		return nil, fmt.Errorf("清除当前状态标记失败: %w", err)
	}
	
	// 第二步：调整与新版本时间重叠的现有记录的结束日期
	// 查找与新版本时间重叠的现有记录，将其end_date调整为新版本生效日期的前一天
	updateQuery := `
		UPDATE organization_units 
		SET end_date = ($3::date - INTERVAL '1 day')::date,
			updated_at = NOW()
		WHERE code = $1 
		  AND tenant_id = $2
		  AND status != 'DELETED'
		  AND effective_date < $3::date
		  AND (end_date IS NULL OR end_date >= $3::date)
	`
	
	result, err := tx.ExecContext(ctx, updateQuery,
		org.Code,
		org.TenantID,
		effectiveDate,
	)
	if err != nil {
		return nil, fmt.Errorf("调整现有版本结束日期失败: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	r.logger.Printf("📅 调整了 %d 条现有记录的结束日期", rowsAffected)

	// 第三步：插入新的时态版本
	insertQuery := `
		INSERT INTO organization_units (
			tenant_id, code, parent_code, name, unit_type, status, 
			level, path, sort_order, description, created_at, updated_at,
			effective_date, end_date, is_temporal, change_reason, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING record_id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err = tx.QueryRowContext(ctx, insertQuery,
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
		effectiveDate,
		org.EndDate,
		org.IsTemporal,
		org.ChangeReason,
		isCurrent,
	).Scan(&org.RecordID, &createdAt, &updatedAt)

	if err != nil {
		return nil, fmt.Errorf("插入时态版本失败: %w", err)
	}

	// 第三步：如果有后续版本，为新版本设置正确的结束日期
	updateNewVersionQuery := `
		UPDATE organization_units 
		SET end_date = (
			SELECT MIN(effective_date - INTERVAL '1 day')::date 
			FROM organization_units future 
			WHERE future.code = $1 
			  AND future.tenant_id = $2
			  AND future.status != 'DELETED'
			  AND future.effective_date > $3::date
			  AND future.record_id != $4
		)
		WHERE record_id = $4
	`
	
	_, err = tx.ExecContext(ctx, updateNewVersionQuery,
		org.Code,
		org.TenantID,
		effectiveDate,
		org.RecordID,
	)
	if err != nil {
		return nil, fmt.Errorf("设置新版本结束日期失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	org.CreatedAt = createdAt
	org.UpdatedAt = updatedAt
	org.EffectiveDate = effectiveDate

	r.logger.Printf("✅ 时态版本创建成功: %s - %s (生效日期: %s, 记录ID: %s)",
		org.Code, org.Name, effectiveDate.String(), org.RecordID)
	
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
          AND status <> 'DELETED' AND deleted_at IS NULL
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
            return nil, fmt.Errorf("组织不存在或已删除不可修改: %s", code)
        }
        return nil, fmt.Errorf("更新组织失败: %w", err)
    }

	r.logger.Printf("组织更新成功: %s - %s (时态: %v)", org.Code, org.Name, org.IsTemporal)
	return &org, nil
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

func (r *OrganizationRepository) Activate(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
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
          AND status <> 'DELETED' AND deleted_at IS NULL
        RETURNING record_id, tenant_id, code, parent_code, name, unit_type, status,
                  level, path, sort_order, description, created_at, updated_at,
                  effective_date, end_date, is_temporal, change_reason
    `, strings.Join(setParts, ", "))

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
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

// GetByCode 通过组织代码获取当前有效的组织记录（用于审计日志）
func (r *OrganizationRepository) GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (*types.Organization, error) {
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, is_temporal, change_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND is_current = true
		LIMIT 1
	`

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("组织不存在: %s", code)
		}
		return nil, fmt.Errorf("获取组织失败: %w", err)
	}

	return &org, nil
}

// GetByRecordId 通过记录ID获取组织记录（用于审计日志）
func (r *OrganizationRepository) GetByRecordId(ctx context.Context, tenantID uuid.UUID, recordId string) (*types.Organization, error) {
	query := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, created_at, updated_at,
		       effective_date, end_date, is_temporal, change_reason
		FROM organization_units 
		WHERE tenant_id = $1 AND record_id = $2
		LIMIT 1
	`

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), recordId).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.IsTemporal, &org.ChangeReason,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("记录不存在: %s", recordId)
		}
		return nil, fmt.Errorf("获取记录失败: %w", err)
	}

	return &org, nil
}

// ListVersionsByCode 列出某组织代码的所有非删除版本，按生效日期倒序
func (r *OrganizationRepository) ListVersionsByCode(ctx context.Context, tenantID uuid.UUID, code string) ([]types.Organization, error) {
    query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
               level, path, sort_order, description, created_at, updated_at,
               effective_date, end_date, is_temporal, change_reason
        FROM organization_units
        WHERE tenant_id = $1 AND code = $2
          AND status <> 'DELETED' AND deleted_at IS NULL
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
            &effectiveDate, &endDate, &org.IsTemporal, &changeReason,
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
