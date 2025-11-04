package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// HierarchyRepository 层级管理仓储
type HierarchyRepository struct {
	db     *sql.DB
	logger pkglogger.Logger
}

// OrganizationNode 组织层级节点
type OrganizationNode struct {
	Code          string      `json:"code"`
	ParentCode    *string     `json:"parentCode"`
	Name          string      `json:"name"`
	Level         int         `json:"level"`
	CodePath      string      `json:"codePath"`
	NamePath      string      `json:"namePath"`
	EffectiveDate *types.Date `json:"effectiveDate"`
	EndDate       *types.Date `json:"endDate"`
	IsCurrent     bool        `json:"isCurrent"`
	Depth         int         `json:"depth"`
	Status        string      `json:"status"`
	UnitType      string      `json:"unitType"`
}

func NewHierarchyRepository(db *sql.DB, baseLogger pkglogger.Logger) *HierarchyRepository {
	return &HierarchyRepository{
		db:     db,
		logger: scopedLogger(baseLogger, "organization", "HierarchyRepository", nil),
	}
}

// GetOrganizationHierarchy 获取组织层级结构 (递归CTE查询)
func (h *HierarchyRepository) GetOrganizationHierarchy(ctx context.Context, rootCode string, tenantID uuid.UUID, maxDepth int) ([]OrganizationNode, error) {
	if maxDepth <= 0 || maxDepth > 17 {
		maxDepth = 17 // 强制17级深度限制
	}

	// PostgreSQL递归CTE查询 - 激进优化版本
	query := `
	WITH RECURSIVE org_tree AS (
		-- 递归基准: 根组织
		SELECT 
			code, parent_code, name, level, 
			COALESCE(code_path, code) as code_path,
			COALESCE(name_path, name) as name_path,
			effective_date, end_date, is_current,
			status, unit_type,
			0 as depth
		FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true
		
		UNION ALL
		
		-- 递归部分: 子组织
		SELECT 
			ou.code, ou.parent_code, ou.name, ou.level,
			COALESCE(ou.code_path, ot.code_path || '/' || ou.code) as code_path,
			COALESCE(ou.name_path, ot.name_path || '/' || ou.name) as name_path,
			ou.effective_date, ou.end_date, ou.is_current,
			ou.status, ou.unit_type,
			ot.depth + 1
		FROM organization_units ou
		INNER JOIN org_tree ot ON ou.parent_code = ot.code
		WHERE ou.tenant_id = $2 AND ou.is_current = true AND ot.depth < $3
	)
	SELECT 
		code, parent_code, name, level, code_path, name_path,
		effective_date, end_date, is_current, depth, status, unit_type
	FROM org_tree 
	ORDER BY depth ASC, code ASC;
	`

	start := time.Now()
	rows, err := h.db.QueryContext(ctx, query, rootCode, tenantID.String(), maxDepth)
	if err != nil {
		h.logger.Errorf("递归层级查询失败: %v", err)
		return nil, fmt.Errorf("failed to query organization hierarchy: %w", err)
	}
	defer rows.Close()

	var nodes []OrganizationNode
	for rows.Next() {
		var node OrganizationNode
		var effectiveDate, endDate sql.NullTime

		err := rows.Scan(
			&node.Code, &node.ParentCode, &node.Name, &node.Level,
			&node.CodePath, &node.NamePath, &effectiveDate, &endDate,
			&node.IsCurrent, &node.Depth, &node.Status, &node.UnitType,
		)
		if err != nil {
			h.logger.Errorf("扫描层级节点失败: %v", err)
			return nil, fmt.Errorf("failed to scan hierarchy node: %w", err)
		}

		// 转换时态字段
		if effectiveDate.Valid {
			node.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
		}
		if endDate.Valid {
			node.EndDate = types.NewDateFromTime(endDate.Time)
		}

		nodes = append(nodes, node)
	}

	duration := time.Since(start)
	h.logger.Infof("递归CTE查询完成: 根节点=%s, 深度=%d, 节点数=%d, 耗时=%v",
		rootCode, maxDepth, len(nodes), duration)

	return nodes, nil
}

// GetDirectChildren 获取直接子组织
func (h *HierarchyRepository) GetDirectChildren(ctx context.Context, parentCode string, tenantID uuid.UUID) ([]OrganizationNode, error) {
	query := `
	SELECT 
		code, parent_code, name, level, 
		COALESCE(code_path, code) as code_path,
		COALESCE(name_path, name) as name_path,
		effective_date, end_date, is_current, status, unit_type
	FROM organization_units 
	WHERE parent_code = $1 AND tenant_id = $2 AND is_current = true
	ORDER BY sort_order ASC, code ASC;
	`

	rows, err := h.db.QueryContext(ctx, query, parentCode, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query direct children: %w", err)
	}
	defer rows.Close()

	var children []OrganizationNode
	for rows.Next() {
		var child OrganizationNode
		var effectiveDate, endDate sql.NullTime

		err := rows.Scan(
			&child.Code, &child.ParentCode, &child.Name, &child.Level,
			&child.CodePath, &child.NamePath, &effectiveDate, &endDate,
			&child.IsCurrent, &child.Status, &child.UnitType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child node: %w", err)
		}

		// 转换时态字段
		if effectiveDate.Valid {
			child.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
		}
		if endDate.Valid {
			child.EndDate = types.NewDateFromTime(endDate.Time)
		}

		children = append(children, child)
	}

	return children, nil
}

// UpdateHierarchyPaths 更新层级路径 (code_path, name_path)
func (h *HierarchyRepository) UpdateHierarchyPaths(ctx context.Context, parentCode string, tenantID uuid.UUID) error {
	// 获取父组织路径
	var parentCodePath, parentNamePath string
	var parentLevel int

	if parentCode == "" {
		// 根组织情况
		parentCodePath = ""
		parentNamePath = ""
		parentLevel = 0
	} else {
		err := h.db.QueryRowContext(ctx, `
			SELECT COALESCE(code_path, code), COALESCE(name_path, name), level
			FROM organization_units 
			WHERE code = $1 AND tenant_id = $2 AND is_current = true
		`, parentCode, tenantID.String()).Scan(&parentCodePath, &parentNamePath, &parentLevel)

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("parent organization not found: %s", parentCode)
			}
			return fmt.Errorf("failed to get parent paths: %w", err)
		}
	}

	// 批量更新子组织路径 - 使用事务确保一致性
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `
	UPDATE organization_units SET
		code_path = CASE 
			WHEN $1 = '' THEN code
			ELSE $1 || '/' || code
		END,
		name_path = CASE
			WHEN $2 = '' THEN name
			ELSE $2 || '/' || name  
		END,
		level = $3 + 1,
		updated_at = NOW()
	WHERE parent_code = $4 AND tenant_id = $5 AND is_current = true;
	`

	result, err := tx.ExecContext(ctx, updateQuery, parentCodePath, parentNamePath, parentLevel, parentCode, tenantID.String())
	if err != nil {
		return fmt.Errorf("failed to update hierarchy paths: %w", err)
	}

	affected, _ := result.RowsAffected()
	h.logger.Infof("层级路径更新: 父节点=%s, 更新子节点数=%d", parentCode, affected)

	return tx.Commit()
}

// GetOrganizationDepth 获取组织深度
func (h *HierarchyRepository) GetOrganizationDepth(ctx context.Context, code string, tenantID uuid.UUID) (int, error) {
	var level int
	err := h.db.QueryRowContext(ctx, `
		SELECT level FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true
	`, code, tenantID.String()).Scan(&level)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("organization not found: %s", code)
		}
		return 0, fmt.Errorf("failed to get organization depth: %w", err)
	}

	return level, nil
}

// GetParentCode 获取父组织代码
func (h *HierarchyRepository) GetParentCode(ctx context.Context, code string, tenantID uuid.UUID) (string, error) {
	var parentCode sql.NullString
	err := h.db.QueryRowContext(ctx, `
		SELECT parent_code FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true
	`, code, tenantID.String()).Scan(&parentCode)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("organization not found: %s", code)
		}
		return "", fmt.Errorf("failed to get parent code: %w", err)
	}

	return parentCode.String, nil
}

// GetOrganization 获取单个组织信息
func (h *HierarchyRepository) GetOrganization(ctx context.Context, code string, tenantID uuid.UUID) (*types.Organization, error) {
	var org types.Organization
	var effectiveDate, endDate sql.NullTime
	var parentCode sql.NullString

	query := `
	SELECT 
		tenant_id,
		code,
		parent_code,
		name,
		unit_type,
		status,
	level,
	COALESCE(code_path, '/' || code) AS code_path,
	COALESCE(name_path, '/' || name) AS name_path,
		sort_order,
		description,
		effective_date,
		end_date,
		is_current,
		created_at,
		updated_at
	FROM organization_units 
	WHERE code = $1 AND tenant_id = $2 AND is_current = true
	`

	err := h.db.QueryRowContext(ctx, query, code, tenantID.String()).Scan(
		&org.TenantID,
		&org.Code,
		&parentCode,
		&org.Name,
		&org.UnitType,
		&org.Status,
		&org.Level,
		&org.CodePath,
		&org.NamePath,
		&org.SortOrder,
		&org.Description,
		&effectiveDate,
		&endDate,
		&org.IsCurrent,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found: %s", code)
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// 处理可选字段
	if parentCode.Valid {
		org.ParentCode = &parentCode.String
	}

	// 转换时态字段
	if effectiveDate.Valid {
		org.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
	}
	if endDate.Valid {
		org.EndDate = types.NewDateFromTime(endDate.Time)
	}

	return &org, nil
}

// GetOrganizationAtDate 获取指定日期的组织版本（若不存在返回nil）
func (h *HierarchyRepository) GetOrganizationAtDate(ctx context.Context, code string, tenantID uuid.UUID, targetDate time.Time) (*OrganizationNode, error) {
	query := `
	SELECT 
		code,
		parent_code,
		name,
		level,
		COALESCE(code_path, '/' || code) AS code_path,
		COALESCE(name_path, '/' || name) AS name_path,
		effective_date,
		end_date,
		is_current,
		status,
		unit_type
	FROM organization_units
	WHERE tenant_id = $1
	  AND code = $2
	  AND status <> 'DELETED'
	  AND effective_date <= $3::date
	  AND (end_date IS NULL OR end_date > $3::date)
	ORDER BY effective_date DESC, created_at DESC
	LIMIT 1`

	row := h.db.QueryRowContext(ctx, query, tenantID.String(), code, targetDate.Format("2006-01-02"))

	var node OrganizationNode
	var parentCode sql.NullString
	var effectiveDate sql.NullTime
	var endDate sql.NullTime

	if err := row.Scan(
		&node.Code,
		&parentCode,
		&node.Name,
		&node.Level,
		&node.CodePath,
		&node.NamePath,
		&effectiveDate,
		&endDate,
		&node.IsCurrent,
		&node.Status,
		&node.UnitType,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get organization at date: %w", err)
	}

	if parentCode.Valid {
		node.ParentCode = &parentCode.String
	}
	if effectiveDate.Valid {
		node.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
	}
	if endDate.Valid {
		node.EndDate = types.NewDateFromTime(endDate.Time)
	}

	// 针对单个节点查询，深度固定为0
	node.Depth = 0

	return &node, nil
}

// CalculateCodePath 计算组织代码路径
func (h *HierarchyRepository) calculateCodePath(ctx context.Context, parentCode *string, tenantID uuid.UUID) (string, error) {
	if parentCode == nil || *parentCode == "" {
		return "", nil // 根组织
	}

	var parentPath string
	err := h.db.QueryRowContext(ctx, `
		SELECT COALESCE(code_path, code) FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true
	`, *parentCode, tenantID.String()).Scan(&parentPath)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("parent organization not found: %s", *parentCode)
		}
		return "", fmt.Errorf("failed to calculate code path: %w", err)
	}

	return parentPath, nil
}

// GetAncestorChain 获取祖先链 (从根到当前节点)
func (h *HierarchyRepository) GetAncestorChain(ctx context.Context, code string, tenantID uuid.UUID) ([]OrganizationNode, error) {
	// 使用递归CTE获取从根节点到目标节点的祖先链
	query := `
	WITH RECURSIVE ancestor_chain AS (
		-- 目标节点
		SELECT 
			code, parent_code, name, level, 
			COALESCE(code_path, code) as code_path,
			COALESCE(name_path, name) as name_path,
			effective_date, end_date, is_current,
			status, unit_type, 0 as distance
		FROM organization_units 
		WHERE code = $1 AND tenant_id = $2 AND is_current = true
		
		UNION ALL
		
		-- 向上追溯父节点
		SELECT 
			ou.code, ou.parent_code, ou.name, ou.level,
			COALESCE(ou.code_path, ou.code) as code_path,
			COALESCE(ou.name_path, ou.name) as name_path,
			ou.effective_date, ou.end_date, ou.is_current,
			ou.status, ou.unit_type, ac.distance + 1
		FROM organization_units ou
		INNER JOIN ancestor_chain ac ON ou.code = ac.parent_code
		WHERE ou.tenant_id = $2 AND ou.is_current = true
	)
	SELECT 
		code, parent_code, name, level, code_path, name_path,
		effective_date, end_date, is_current, distance, status, unit_type
	FROM ancestor_chain 
	ORDER BY distance DESC; -- 从根节点开始排序
	`

	rows, err := h.db.QueryContext(ctx, query, code, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query ancestor chain: %w", err)
	}
	defer rows.Close()

	var ancestors []OrganizationNode
	for rows.Next() {
		var node OrganizationNode
		var effectiveDate, endDate sql.NullTime
		var distance int

		err := rows.Scan(
			&node.Code, &node.ParentCode, &node.Name, &node.Level,
			&node.CodePath, &node.NamePath, &effectiveDate, &endDate,
			&node.IsCurrent, &distance, &node.Status, &node.UnitType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ancestor node: %w", err)
		}

		// 转换时态字段
		if effectiveDate.Valid {
			node.EffectiveDate = types.NewDateFromTime(effectiveDate.Time)
		}
		if endDate.Valid {
			node.EndDate = types.NewDateFromTime(endDate.Time)
		}

		node.Depth = distance
		ancestors = append(ancestors, node)
	}

	h.logger.Infof("祖先链查询: 目标=%s, 层级数=%d", code, len(ancestors))
	return ancestors, nil
}
