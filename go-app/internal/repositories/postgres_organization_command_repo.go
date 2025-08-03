package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

// PostgresOrganizationCommandRepository PostgreSQL组织命令仓储实现
type PostgresOrganizationCommandRepository struct {
	db     *sqlx.DB
	logger Logger
}

// Logger 日志接口
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewPostgresOrganizationCommandRepository 创建PostgreSQL组织命令仓储
func NewPostgresOrganizationCommandRepository(db *sqlx.DB, logger Logger) *PostgresOrganizationCommandRepository {
	return &PostgresOrganizationCommandRepository{
		db:     db,
		logger: logger,
	}
}

// CreateOrganization 创建组织
func (r *PostgresOrganizationCommandRepository) CreateOrganization(ctx context.Context, org Organization) error {
	query := `
		INSERT INTO organization_units (
			id, tenant_id, unit_type, name, description, parent_unit_id, 
			status, profile, level, employee_count, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	// 序列化Profile为JSON
	profileJSON, err := json.Marshal(org.Profile)
	if err != nil {
		r.logger.Error("Failed to marshal profile", "error", err, "org_id", org.ID)
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// 计算组织层级
	level := r.calculateOrganizationLevel(ctx, org.ParentUnitID, org.TenantID)

	_, err = r.db.ExecContext(ctx, query,
		org.ID,
		org.TenantID,
		org.UnitType,
		org.Name,
		org.Description,
		org.ParentUnitID,
		org.Status,
		profileJSON,
		level,
		org.EmployeeCount,
		org.IsActive,
		org.CreatedAt,
		org.UpdatedAt,
	)

	if err != nil {
		// 处理唯一约束冲突
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("organization already exists: %w", err)
			case "23503": // foreign_key_violation
				return fmt.Errorf("invalid parent organization: %w", err)
			case "23514": // check_violation
				return fmt.Errorf("invalid organization data: %w", err)
			}
		}
		r.logger.Error("Failed to create organization", "error", err, "org_id", org.ID)
		return fmt.Errorf("failed to create organization: %w", err)
	}

	// 更新父组织的子组织计数
	if org.ParentUnitID != nil {
		if err := r.updateChildCount(ctx, *org.ParentUnitID, org.TenantID); err != nil {
			r.logger.Warn("Failed to update parent child count", "error", err, "parent_id", *org.ParentUnitID)
		}
	}

	r.logger.Info("Organization created successfully", "org_id", org.ID, "name", org.Name)
	return nil
}

// UpdateOrganization 更新组织
func (r *PostgresOrganizationCommandRepository) UpdateOrganization(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, changes map[string]interface{}) error {
	if len(changes) == 0 {
		return fmt.Errorf("no changes provided")
	}

	// 构建动态更新SQL
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range changes {
		switch field {
		case "name":
			setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "description":
			setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "parent_unit_id":
			setParts = append(setParts, fmt.Sprintf("parent_unit_id = $%d", argIndex))
			args = append(args, value)
			argIndex++
			// 重新计算层级
			if parentID, ok := value.(*uuid.UUID); ok {
				newLevel := r.calculateOrganizationLevel(ctx, parentID, tenantID)
				setParts = append(setParts, fmt.Sprintf("level = $%d", argIndex))
				args = append(args, newLevel)
				argIndex++
			}
		case "status":
			setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "is_active":
			setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, value)
			argIndex++
		case "profile":
			profileJSON, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal profile: %w", err)
			}
			setParts = append(setParts, fmt.Sprintf("profile = $%d", argIndex))
			args = append(args, profileJSON)
			argIndex++
		}
	}

	// 添加更新时间
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加WHERE条件
	args = append(args, id, tenantID)

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s 
		WHERE id = $%d AND tenant_id = $%d`,
		fmt.Sprintf("%s", setParts),
		argIndex, argIndex+1)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to update organization", "error", err, "org_id", id)
		return fmt.Errorf("failed to update organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found or not authorized")
	}

	r.logger.Info("Organization updated successfully", "org_id", id, "changes", len(changes))
	return nil
}

// DeleteOrganization 删除组织
func (r *PostgresOrganizationCommandRepository) DeleteOrganization(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error {
	// 检查是否有子组织
	var childCount int
	err := r.db.GetContext(ctx, &childCount,
		"SELECT COUNT(*) FROM organization_units WHERE parent_unit_id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check child organizations: %w", err)
	}

	if childCount > 0 {
		return fmt.Errorf("cannot delete organization with %d child organizations", childCount)
	}

	// 检查是否有关联的员工
	var employeeCount int
	err = r.db.GetContext(ctx, &employeeCount,
		"SELECT COUNT(*) FROM employees WHERE organization_id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil {
		r.logger.Warn("Failed to check associated employees", "error", err)
	}

	if employeeCount > 0 {
		return fmt.Errorf("cannot delete organization with %d associated employees", employeeCount)
	}

	// 获取父组织ID用于更新计数
	var parentID *uuid.UUID
	err = r.db.GetContext(ctx, &parentID,
		"SELECT parent_unit_id FROM organization_units WHERE id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get parent organization: %w", err)
	}

	// 执行删除
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM organization_units WHERE id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil {
		r.logger.Error("Failed to delete organization", "error", err, "org_id", id)
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found or not authorized")
	}

	// 更新父组织的子组织计数
	if parentID != nil {
		if err := r.updateChildCount(ctx, *parentID, tenantID); err != nil {
			r.logger.Warn("Failed to update parent child count", "error", err, "parent_id", *parentID)
		}
	}

	r.logger.Info("Organization deleted successfully", "org_id", id)
	return nil
}

// MoveOrganization 移动组织
func (r *PostgresOrganizationCommandRepository) MoveOrganization(ctx context.Context, id uuid.UUID, newParentID *uuid.UUID, tenantID uuid.UUID) error {
	// 验证不能移动到自己的子组织下
	if newParentID != nil {
		isDescendant, err := r.isDescendant(ctx, *newParentID, id, tenantID)
		if err != nil {
			return fmt.Errorf("failed to check descendant relationship: %w", err)
		}
		if isDescendant {
			return fmt.Errorf("cannot move organization to its descendant")
		}
	}

	// 获取当前父组织ID
	var oldParentID *uuid.UUID
	err := r.db.GetContext(ctx, &oldParentID,
		"SELECT parent_unit_id FROM organization_units WHERE id = $1 AND tenant_id = $2",
		id, tenantID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current parent: %w", err)
	}

	// 计算新的层级
	newLevel := r.calculateOrganizationLevel(ctx, newParentID, tenantID)

	// 执行移动
	result, err := r.db.ExecContext(ctx, `
		UPDATE organization_units 
		SET parent_unit_id = $1, level = $2, updated_at = $3 
		WHERE id = $4 AND tenant_id = $5`,
		newParentID, newLevel, time.Now(), id, tenantID)
	if err != nil {
		r.logger.Error("Failed to move organization", "error", err, "org_id", id)
		return fmt.Errorf("failed to move organization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found or not authorized")
	}

	// 递归更新所有子组织的层级
	if err := r.updateDescendantLevels(ctx, id, tenantID, newLevel); err != nil {
		r.logger.Warn("Failed to update descendant levels", "error", err, "org_id", id)
	}

	// 更新原父组织和新父组织的子组织计数
	if oldParentID != nil {
		if err := r.updateChildCount(ctx, *oldParentID, tenantID); err != nil {
			r.logger.Warn("Failed to update old parent child count", "error", err, "parent_id", *oldParentID)
		}
	}
	if newParentID != nil {
		if err := r.updateChildCount(ctx, *newParentID, tenantID); err != nil {
			r.logger.Warn("Failed to update new parent child count", "error", err, "parent_id", *newParentID)
		}
	}

	r.logger.Info("Organization moved successfully", "org_id", id, "new_parent_id", newParentID)
	return nil
}

// SetOrganizationStatus 设置组织状态
func (r *PostgresOrganizationCommandRepository) SetOrganizationStatus(ctx context.Context, id uuid.UUID, tenantID uuid.UUID, status string) error {
	isActive := status == "ACTIVE"

	result, err := r.db.ExecContext(ctx, `
		UPDATE organization_units 
		SET status = $1, is_active = $2, updated_at = $3 
		WHERE id = $4 AND tenant_id = $5`,
		status, isActive, time.Now(), id, tenantID)
	if err != nil {
		r.logger.Error("Failed to set organization status", "error", err, "org_id", id, "status", status)
		return fmt.Errorf("failed to set organization status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("organization not found or not authorized")
	}

	r.logger.Info("Organization status updated", "org_id", id, "status", status)
	return nil
}

// BulkUpdateOrganizations 批量更新组织
func (r *PostgresOrganizationCommandRepository) BulkUpdateOrganizations(ctx context.Context, ids []uuid.UUID, tenantID uuid.UUID, changes map[string]interface{}) error {
	if len(ids) == 0 {
		return fmt.Errorf("no organization IDs provided")
	}
	if len(changes) == 0 {
		return fmt.Errorf("no changes provided")
	}

	// 构建批量更新SQL
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	for field, value := range changes {
		switch field {
		case "status":
			setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
			args = append(args, value)
			argIndex++
			// 同时更新is_active
			isActive := value == "ACTIVE"
			setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
			args = append(args, isActive)
			argIndex++
		case "profile":
			profileJSON, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal profile: %w", err)
			}
			setParts = append(setParts, fmt.Sprintf("profile = $%d", argIndex))
			args = append(args, profileJSON)
			argIndex++
		}
	}

	// 添加更新时间
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// 添加WHERE条件
	args = append(args, pq.Array(ids), tenantID)

	query := fmt.Sprintf(`
		UPDATE organization_units 
		SET %s 
		WHERE id = ANY($%d) AND tenant_id = $%d`,
		fmt.Sprintf("%s", setParts),
		argIndex, argIndex+1)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to bulk update organizations", "error", err, "count", len(ids))
		return fmt.Errorf("failed to bulk update organizations: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Organizations bulk updated", "requested", len(ids), "updated", rowsAffected)
	return nil
}

// WithTransaction 在事务中执行操作
func (r *PostgresOrganizationCommandRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 创建事务上下文
	txCtx := context.WithValue(ctx, "tx", tx)

	// 执行事务操作
	err = fn(txCtx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			r.logger.Error("Failed to rollback transaction", "error", rollbackErr)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// 辅助方法

// calculateOrganizationLevel 计算组织层级
func (r *PostgresOrganizationCommandRepository) calculateOrganizationLevel(ctx context.Context, parentID *uuid.UUID, tenantID uuid.UUID) int {
	if parentID == nil {
		return 0 // 根组织
	}

	var parentLevel int
	err := r.db.GetContext(ctx, &parentLevel,
		"SELECT level FROM organization_units WHERE id = $1 AND tenant_id = $2",
		*parentID, tenantID)
	if err != nil {
		r.logger.Warn("Failed to get parent level", "error", err, "parent_id", *parentID)
		return 1 // 默认层级
	}

	return parentLevel + 1
}

// updateChildCount 更新子组织计数
func (r *PostgresOrganizationCommandRepository) updateChildCount(ctx context.Context, parentID uuid.UUID, tenantID uuid.UUID) error {
	var childCount int
	err := r.db.GetContext(ctx, &childCount,
		"SELECT COUNT(*) FROM organization_units WHERE parent_unit_id = $1 AND tenant_id = $2",
		parentID, tenantID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"UPDATE organization_units SET employee_count = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4",
		childCount, time.Now(), parentID, tenantID)
	return err
}

// isDescendant 检查组织是否为另一个组织的后代
func (r *PostgresOrganizationCommandRepository) isDescendant(ctx context.Context, potentialAncestor uuid.UUID, potentialDescendant uuid.UUID, tenantID uuid.UUID) (bool, error) {
	query := `
		WITH RECURSIVE org_hierarchy AS (
			SELECT id, parent_unit_id, 0 as depth
			FROM organization_units 
			WHERE id = $1 AND tenant_id = $3
			
			UNION ALL
			
			SELECT ou.id, ou.parent_unit_id, oh.depth + 1
			FROM organization_units ou
			JOIN org_hierarchy oh ON ou.parent_unit_id = oh.id
			WHERE ou.tenant_id = $3 AND oh.depth < 10
		)
		SELECT EXISTS(
			SELECT 1 FROM org_hierarchy 
			WHERE id = $2
		)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, potentialDescendant, potentialAncestor, tenantID)
	return exists, err
}

// updateDescendantLevels 递归更新后代组织的层级
func (r *PostgresOrganizationCommandRepository) updateDescendantLevels(ctx context.Context, parentID uuid.UUID, tenantID uuid.UUID, parentLevel int) error {
	// 获取所有直接子组织
	var childIDs []uuid.UUID
	err := r.db.SelectContext(ctx, &childIDs,
		"SELECT id FROM organization_units WHERE parent_unit_id = $1 AND tenant_id = $2",
		parentID, tenantID)
	if err != nil {
		return err
	}

	// 更新直接子组织的层级
	childLevel := parentLevel + 1
	for _, childID := range childIDs {
		_, err := r.db.ExecContext(ctx,
			"UPDATE organization_units SET level = $1, updated_at = $2 WHERE id = $3 AND tenant_id = $4",
			childLevel, time.Now(), childID, tenantID)
		if err != nil {
			r.logger.Warn("Failed to update child level", "error", err, "child_id", childID)
			continue
		}

		// 递归更新子组织的后代
		if err := r.updateDescendantLevels(ctx, childID, tenantID, childLevel); err != nil {
			r.logger.Warn("Failed to update descendant levels", "error", err, "child_id", childID)
		}
	}

	return nil
}