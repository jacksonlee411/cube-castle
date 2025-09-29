package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
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

var (
	ErrOrganizationHasChildren  = errors.New("organization has non-deleted child units")
	ErrOrganizationPrecondition = errors.New("organization precondition failed")
)

type OrganizationHasChildrenError struct {
	Count int
}

func (e *OrganizationHasChildrenError) Error() string {
	return ErrOrganizationHasChildren.Error()
}

func (e *OrganizationHasChildrenError) Is(target error) bool {
	return target == ErrOrganizationHasChildren
}

func NewOrganizationHasChildrenError(count int) error {
	return &OrganizationHasChildrenError{Count: count}
}

type hierarchyFields struct {
	Path     string
	CodePath string
	NamePath string
	Level    int
	oldLevel int
}

func ensureJoinedPath(base, segment string) string {
	base = strings.TrimSpace(base)
	segment = strings.TrimSpace(segment)
	base = strings.TrimRight(base, "/")
	segment = strings.TrimLeft(segment, "/")
	if base == "" {
		return "/" + segment
	}
	return base + "/" + segment
}

func (r *OrganizationRepository) recalculateSelfHierarchy(ctx context.Context, tenantID uuid.UUID, code string, recordID *string, parentCode *string, overrideName *string) (*hierarchyFields, error) {
	var (
		resolvedCode string
		currentName  string
		currentLevel int
	)

	if recordID != nil {
		err := r.db.QueryRowContext(ctx, `
			SELECT code, name, level
			FROM organization_units
			WHERE tenant_id = $1 AND record_id = $2 AND status <> 'DELETED'
			LIMIT 1
		`, tenantID.String(), *recordID).Scan(&resolvedCode, &currentName, &currentLevel)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("记录不存在: %s", *recordID)
			}
			return nil, fmt.Errorf("查询组织记录失败: %w", err)
		}
	} else {
		resolvedCode = code
		err := r.db.QueryRowContext(ctx, `
			SELECT name, level
			FROM organization_units
			WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
			LIMIT 1
		`, tenantID.String(), code).Scan(&currentName, &currentLevel)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("组织不存在或已删除不可修改: %s", code)
			}
			return nil, fmt.Errorf("查询组织失败: %w", err)
		}
	}

	finalName := currentName
	if overrideName != nil {
		finalName = strings.TrimSpace(*overrideName)
	}

	if resolvedCode == "" {
		resolvedCode = code
	}

	fields, err := r.calculateHierarchyFields(ctx, tenantID, resolvedCode, parentCode, finalName)
	if err != nil {
		return nil, err
	}
	fields.oldLevel = currentLevel

	r.logger.Printf("recalculateSelfHierarchy: code=%s oldLevel=%d newLevel=%d path=%s", resolvedCode, fields.oldLevel, fields.Level, fields.Path)
	return fields, nil
}

func (r *OrganizationRepository) calculateHierarchyFields(ctx context.Context, tenantID uuid.UUID, code string, parentCode *string, finalName string) (*hierarchyFields, error) {
	finalName = strings.TrimSpace(finalName)
	if finalName == "" {
		return nil, fmt.Errorf("组织名称不能为空")
	}

	fields := &hierarchyFields{}

	if parentCode == nil {
		fields.Level = 1
		fields.Path = ensureJoinedPath("", code)
		fields.CodePath = fields.Path
		fields.NamePath = ensureJoinedPath("", finalName)
		return fields, nil
	}

	trimmedParent := strings.TrimSpace(*parentCode)
	if trimmedParent == "" {
		// treated as root if blank string provided
		fields.Level = 1
		fields.Path = ensureJoinedPath("", code)
		fields.CodePath = fields.Path
		fields.NamePath = ensureJoinedPath("", finalName)
		return fields, nil
	}

	var parentCodePath, parentNamePath string
	var parentLevel int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(NULLIF(code_path, ''), '/' || code),
		       COALESCE(NULLIF(name_path, ''), '/' || name),
		       level
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
		LIMIT 1
	`, tenantID.String(), trimmedParent).Scan(&parentCodePath, &parentNamePath, &parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("父组织不存在: %s", trimmedParent)
		}
		return nil, fmt.Errorf("查询父组织失败: %w", err)
	}

	fields.Level = parentLevel + 1
	fields.Path = ensureJoinedPath(parentCodePath, code)
	fields.CodePath = fields.Path
	fields.NamePath = ensureJoinedPath(parentNamePath, finalName)

	return fields, nil
}

// ComputeHierarchyForNew 计算新建或新版本的层级字段（path/codePath/namePath/level）
func (r *OrganizationRepository) ComputeHierarchyForNew(ctx context.Context, tenantID uuid.UUID, code string, parentCode *string, name string) (*hierarchyFields, error) {
	return r.calculateHierarchyFields(ctx, tenantID, strings.TrimSpace(code), parentCode, name)
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
	tenantUUID, err := uuid.Parse(org.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效的租户ID: %w", err)
	}

	fields, err := r.ComputeHierarchyForNew(ctx, tenantUUID, org.Code, org.ParentCode, org.Name)
	if err != nil {
		return nil, err
	}

	org.Level = fields.Level
	org.Path = fields.Path
	org.CodePath = fields.CodePath
	org.NamePath = fields.NamePath

	query := `
        INSERT INTO organization_units (
            tenant_id, code, parent_code, name, unit_type, status, 
            level, path, code_path, name_path, sort_order, description, created_at, updated_at,
            effective_date, end_date, change_reason, is_current
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
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

	err = r.db.QueryRowContext(ctx, query,
		org.TenantID,
		org.Code,
		org.ParentCode,
		org.Name,
		org.UnitType,
		org.Status,
		org.Level,
		org.Path,
		org.CodePath,
		org.NamePath,
		org.SortOrder,
		org.Description,
		time.Now(),
		time.Now(),
		effectiveDate, // Date类型
		org.EndDate,   // 允许为nil
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

	r.logger.Printf("组织创建成功: %s - %s", org.Code, org.Name)
	return org, nil
}

func (r *OrganizationRepository) CreateInTransaction(ctx context.Context, tx *sql.Tx, org *types.Organization) (*types.Organization, error) {
	query := `
        INSERT INTO organization_units (
            tenant_id, code, parent_code, name, unit_type, status, 
            level, path, sort_order, description, created_at, updated_at,
            effective_date, end_date, change_reason, is_current
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
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
		trimmed := strings.TrimSpace(*req.ParentCode)
		var normalizedParent *string
		if trimmed != "" {
			normalizedParent = &trimmed
		}

		fields, err := r.recalculateSelfHierarchy(ctx, tenantID, code, nil, normalizedParent, nameOverride)
		if err != nil {
			return nil, err
		}

		if normalizedParent != nil {
			addAssignment("parent_code", *normalizedParent)
		} else {
			addAssignment("parent_code", nil)
		}
		addAssignment("path", fields.Path)
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
	query := fmt.Sprintf("UPDATE organization_units\nSET %s\nWHERE tenant_id = $1 AND code = $2\n  AND status <> 'DELETED'\nRETURNING tenant_id, code, parent_code, name, unit_type, status,\n          level, path, code_path, name_path, sort_order, description, created_at, updated_at,\n          effective_date, end_date, change_reason", setClause)

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder,
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

func (r *OrganizationRepository) Suspend(ctx context.Context, tenantID uuid.UUID, code string, reason string) (*types.Organization, error) {
	query := `
        UPDATE organization_units 
        SET status = 'INACTIVE', updated_at = $3
        WHERE tenant_id = $1 AND code = $2 AND status = 'ACTIVE'
        RETURNING tenant_id, code, parent_code, name, unit_type, status, 
                 level, path, code_path, name_path, sort_order, description, created_at, updated_at,
                 effective_date, end_date, change_reason
    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
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
			         level, path, code_path, name_path, sort_order, description, created_at, updated_at,
			         effective_date, end_date, change_reason
	    `

	var org types.Organization
	var parentCode sql.NullString
	var effectiveDate, endDate sql.NullTime
	var changeReason sql.NullString

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code, time.Now()).Scan(
		&org.TenantID, &org.Code, &parentCode, &org.Name, &org.UnitType, &org.Status,
		&org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder, &org.Description, &org.CreatedAt, &org.UpdatedAt,
		&effectiveDate, &endDate, &changeReason,
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

// UpdateByRecordId 通过UUID更新历史记录
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
		trimmed := strings.TrimSpace(*req.ParentCode)
		var normalizedParent *string
		if trimmed != "" {
			normalizedParent = &trimmed
		}

		fields, err := r.recalculateSelfHierarchy(ctx, tenantID, "", &recordId, normalizedParent, nameOverride)
		if err != nil {
			return nil, err
		}

		if normalizedParent != nil {
			addAssignment("parent_code", *normalizedParent)
		} else {
			addAssignment("parent_code", nil)
		}
		addAssignment("path", fields.Path)
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
	query := fmt.Sprintf("UPDATE organization_units\nSET %s\nWHERE tenant_id = $1 AND record_id = $2\n  AND status <> 'DELETED'\nRETURNING record_id, tenant_id, code, parent_code, name, unit_type, status,\n          level, path, code_path, name_path, sort_order, description, created_at, updated_at,\n          effective_date, end_date, change_reason", setClause)

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder,
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

// GetByCode 通过组织代码获取当前有效的组织记录（用于审计日志）
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
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.CodePath, &org.NamePath, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.ChangeReason,
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
               effective_date, end_date, change_reason
        FROM organization_units 
        WHERE tenant_id = $1 AND record_id = $2
        LIMIT 1
    `

	var org types.Organization
	err := r.db.QueryRowContext(ctx, query, tenantID.String(), recordId).Scan(
		&org.RecordID, &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
		&org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
		&org.Description, &org.CreatedAt, &org.UpdatedAt,
		&org.EffectiveDate, &org.EndDate, &org.ChangeReason,
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
