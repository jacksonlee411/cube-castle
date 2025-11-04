package repository

import (
	"context"
	"database/sql"
	"time"

	"cube-castle/internal/organization/dto"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// 单个组织查询 - 超快速索引查询
func (r *PostgreSQLRepository) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*dto.Organization, error) {
	// 使用 idx_current_record_fast 索引
	query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
               level,
               COALESCE(code_path, '/' || code) AS code_path,
               COALESCE(name_path, '/' || name) AS name_path,
               sort_order, description, profile, created_at, updated_at,
               effective_date, end_date, is_current, change_reason,
               deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
        LIMIT 1`

	log := r.loggerFor("organization.get", pkglogger.Fields{
		"tenantId": tenantID.String(),
		"code":     code,
	})
	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var org dto.Organization
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.CodePathField, &org.NamePathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("organization not found")
			return nil, nil
		}
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization query failed")
		return nil, err
	}

	duration := time.Since(start)
	log.WithFields(pkglogger.Fields{
		"duration_ms": duration.Milliseconds(),
	}).Info("organization query succeeded")

	return &org, nil
}

// 极速时态查询 - 时间点查询（利用时态索引）
func (r *PostgreSQLRepository) GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code, date string) (*dto.Organization, error) {
	// 使用计算的区间终点（computed_end_date），避免依赖物理 end_date 的准确性
	query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level,
                COALESCE(code_path, '/' || code) AS code_path,
                COALESCE(name_path, '/' || name) AS name_path,
                sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED'
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, code_path, name_path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, code_path, name_path, sort_order, description, profile, created_at, updated_at,
               effective_date, computed_end_date AS end_date, is_current, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $3::date 
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
        ORDER BY effective_date DESC, created_at DESC
        LIMIT 1`

	log := r.loggerFor("organization.atDate", pkglogger.Fields{
		"tenantId": tenantID.String(),
		"code":     code,
		"date":     date,
	})
	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code, date)

	var org dto.Organization
	var isTemporal bool
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.CodePathField, &org.NamePathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("organization snapshot not found")
			return nil, nil
		}
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization temporal query failed")
		return nil, err
	}

	duration := time.Since(start)
	log.WithFields(pkglogger.Fields{"duration_ms": duration.Milliseconds()}).Info("organization temporal query succeeded")

	return &org, nil
}

// 历史范围查询 - 窗口函数优化
func (r *PostgreSQLRepository) GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code, fromDate, toDate string) ([]dto.Organization, error) {
	// 历史范围查询：使用计算的区间终点（computed_end_date）并基于区间重叠选择
	query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level,
                COALESCE(code_path, '/' || code) AS code_path,
                COALESCE(name_path, '/' || name) AS name_path,
                sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED'
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, code_path, name_path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, code_path, name_path, sort_order, description, profile, created_at, updated_at,
            effective_date, computed_end_date AS end_date, is_current, is_temporal, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $4::date
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
        ORDER BY effective_date DESC, created_at DESC`

	log := r.loggerFor("organization.history", pkglogger.Fields{
		"tenantId": tenantID.String(),
		"code":     code,
		"fromDate": fromDate,
		"toDate":   toDate,
	})
	start := time.Now()
	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, fromDate, toDate)
	if err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization history query failed")
		return nil, err
	}
	defer rows.Close()

	var organizations []dto.Organization
	for rows.Next() {
		var org dto.Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.CodePathField, &org.NamePathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, new(bool),
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
		)
		if err != nil {
			log.WithFields(pkglogger.Fields{"error": err}).Error("organization history scan failed")
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	log.WithFields(pkglogger.Fields{
		"result_count": len(organizations),
		"duration_ms":  duration.Milliseconds(),
	}).Info("organization history query succeeded")

	return organizations, nil
}

// 组织版本查询 - 按计划规范实现，返回指定code的全部版本
func (r *PostgreSQLRepository) GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]dto.Organization, error) {
	log := r.loggerFor("organization.versions", pkglogger.Fields{
		"tenantId":       tenantID.String(),
		"code":           code,
		"includeDeleted": includeDeleted,
	})
	start := time.Now()

	// 构建查询 - 过滤条件：tenant_id = $tenant AND code = $code
	baseQuery := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level,
		       COALESCE(code_path, '/' || code) AS code_path,
		       COALESCE(name_path, '/' || name) AS name_path,
		       sort_order, description, profile, created_at, updated_at,
	           effective_date, end_date, is_current, change_reason,
	           deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
	           hierarchy_depth
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2`

	args := []interface{}{tenantID.String(), code}

	// includeDeleted=false: status != 'DELETED'
	if !includeDeleted {
		baseQuery += " AND status != 'DELETED'"
	}

	// 排序：ORDER BY effective_date ASC (按计划要求)
	finalQuery := baseQuery + " ORDER BY effective_date ASC"

	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization versions query failed")
		return nil, err
	}
	defer rows.Close()

	var organizations []dto.Organization
	for rows.Next() {
		var org dto.Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.CodePathField, &org.NamePathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
			&org.HierarchyDepthField,
		)
		if err != nil {
			log.WithFields(pkglogger.Fields{"error": err}).Error("organization versions scan failed")
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	log.WithFields(pkglogger.Fields{
		"result_count": len(organizations),
		"duration_ms":  duration.Milliseconds(),
	}).Info("organization versions query succeeded")

	return organizations, nil
}
