package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/model"
	"cube-castle-deployment-test/internal/auth"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type AuditHistoryConfig struct {
	StrictValidation        bool
	AllowFallback           bool
	CircuitBreakerThreshold int32
	LegacyMode              bool
}

// PostgreSQL极速仓储 - 零抽象开销
type PostgreSQLRepository struct {
	db                     *sql.DB
	redisClient            *redis.Client
	logger                 *log.Logger
	auditConfig            AuditHistoryConfig
	validationFailureCount int32
}

func NewPostgreSQLRepository(db *sql.DB, redisClient *redis.Client, logger *log.Logger, auditConfig AuditHistoryConfig) *PostgreSQLRepository {
	return &PostgreSQLRepository{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
		auditConfig: auditConfig,
	}
}

// 极速当前组织查询 - 利用部分索引 idx_current_organizations_list (API契约v4.2.1)
func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *model.OrganizationFilter, pagination *model.PaginationInput) (*model.OrganizationConnection, error) {
	start := time.Now()

	// 解析分页参数 - 使用契约默认值
	page := int32(1)
	pageSize := int32(50)
	if pagination != nil {
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
		}
	}

	offset := (page - 1) * pageSize
	limit := pageSize

	includeDisabledAncestors := false

	var (
		status, searchText, unitType, parentCode string
		includeCodes, excludeCodes               []string
		asOfDateParam                            sql.NullString
		excludeDescendantsParam                  sql.NullString
	)

	if filter != nil {
		includeDisabledAncestors = filter.IncludeDisabledAncestors
		if filter.Status != nil {
			status = strings.TrimSpace(*filter.Status)
		}
		if filter.SearchText != nil {
			searchText = strings.TrimSpace(*filter.SearchText)
		}
		if filter.UnitType != nil {
			unitType = strings.TrimSpace(*filter.UnitType)
		}
		if filter.ParentCode != nil {
			parentCode = strings.TrimSpace(*filter.ParentCode)
		}
		if filter.AsOfDate != nil {
			if trimmed := strings.TrimSpace(*filter.AsOfDate); trimmed != "" {
				asOfDateParam = sql.NullString{String: trimmed, Valid: true}
			}
		}
		if filter.ExcludeDescendantsOf != nil {
			if trimmed := strings.TrimSpace(*filter.ExcludeDescendantsOf); trimmed != "" {
				excludeDescendantsParam = sql.NullString{String: trimmed, Valid: true}
			}
		}
		if filter.ExcludeCodes != nil {
			for _, code := range *filter.ExcludeCodes {
				if trimmed := strings.TrimSpace(code); trimmed != "" {
					excludeCodes = append(excludeCodes, trimmed)
				}
			}
		}
		if filter.Codes != nil {
			for _, code := range *filter.Codes {
				if trimmed := strings.TrimSpace(code); trimmed != "" {
					includeCodes = append(includeCodes, trimmed)
				}
			}
		}
	}

	cte := `
WITH parent_path AS (
    SELECT DISTINCT ON (code)
        code,
        COALESCE(code_path, '/' || code) AS code_path
    FROM organization_units
    WHERE tenant_id = $1
      AND $3::text IS NOT NULL
      AND code = $3::text
      AND status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          effective_date <= $2::date AND (end_date IS NULL OR end_date > $2::date)
        )
      )
    ORDER BY code, effective_date DESC, created_at DESC
),
latest_versions AS (
    SELECT DISTINCT ON (code)
        record_id, tenant_id, code, parent_code, name, unit_type, status,
        level, path, sort_order, description, profile, created_at, updated_at,
        effective_date, end_date, is_current, change_reason,
        deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
        COALESCE(code_path, '/' || code) AS code_path
    FROM organization_units
    WHERE tenant_id = $1
      AND status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          effective_date <= $2::date AND (end_date IS NULL OR end_date > $2::date)
        )
      )
    ORDER BY code, effective_date DESC, created_at DESC
)
`

	baseSelect := `
SELECT lv.record_id, lv.tenant_id, lv.code, lv.parent_code, lv.name, lv.unit_type, lv.status,
       lv.level, lv.path, lv.sort_order, lv.description, lv.profile, lv.created_at, lv.updated_at,
       lv.effective_date, lv.end_date, lv.is_current, lv.change_reason,
       lv.deleted_at, lv.deleted_by, lv.deletion_reason, lv.suspended_at, lv.suspended_by, lv.suspension_reason,
       COALESCE(child_stats.child_count, 0) AS children_count
FROM latest_versions lv
LEFT JOIN parent_path pp ON TRUE
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS child_count
    FROM organization_units child
    WHERE child.tenant_id = lv.tenant_id
      AND child.parent_code = lv.code
      AND child.status <> 'DELETED'
      AND (
        $2::text IS NULL OR (
          child.effective_date <= $2::date AND (child.end_date IS NULL OR child.end_date > $2::date)
        )
      )
) child_stats ON TRUE
WHERE 1=1`

	countSelect := `
SELECT COUNT(*)
FROM latest_versions lv
LEFT JOIN parent_path pp ON TRUE
WHERE 1=1`

	args := []interface{}{tenantID.String(), asOfDateParam, excludeDescendantsParam}
	argIndex := 4
	whereConditions := ""

	if status != "" {
		if includeDisabledAncestors && parentCode != "" {
			whereConditions += fmt.Sprintf(" AND (lv.status = $%d OR (lv.parent_code = $%d AND lv.status <> 'DELETED'))", argIndex, argIndex+1)
			args = append(args, status, parentCode)
			argIndex += 2
		} else {
			whereConditions += fmt.Sprintf(" AND lv.status = $%d", argIndex)
			args = append(args, status)
			argIndex++
		}
	} else {
		whereConditions += " AND lv.status <> 'DELETED'"
	}

	if unitType != "" {
		whereConditions += fmt.Sprintf(" AND lv.unit_type = $%d", argIndex)
		args = append(args, unitType)
		argIndex++
	}

	if parentCode != "" {
		whereConditions += fmt.Sprintf(" AND lv.parent_code = $%d", argIndex)
		args = append(args, parentCode)
		argIndex++
	}

	if len(includeCodes) > 0 {
		whereConditions += fmt.Sprintf(" AND lv.code = ANY($%d)", argIndex)
		args = append(args, pq.StringArray(includeCodes))
		argIndex++
	}

	if len(excludeCodes) > 0 {
		whereConditions += fmt.Sprintf(" AND NOT (lv.code = ANY($%d))", argIndex)
		args = append(args, pq.StringArray(excludeCodes))
		argIndex++
	}

	whereConditions += ` AND (
    $3::text IS NULL OR (
        lv.code <> $3::text AND (
            pp.code_path IS NULL OR lv.code_path NOT LIKE pp.code_path || '/%'
        )
    )
)`

	if searchText != "" {
		whereConditions += fmt.Sprintf(" AND (lv.name ILIKE $%d OR lv.code ILIKE $%d)", argIndex, argIndex)
		pattern := "%" + searchText + "%"
		args = append(args, pattern)
		argIndex++
	}

	countQuery := cte + countSelect + whereConditions
	countArgs := append([]interface{}{}, args...)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		r.logger.Printf("[ERROR] 查询组织总数失败: %v", err)
		return nil, err
	}

	orderClause := fmt.Sprintf(" ORDER BY COALESCE(lv.sort_order, 0) NULLS LAST, lv.code LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	dataQuery := cte + baseSelect + whereConditions + orderClause
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 查询组织列表失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField, &org.ChildrenCountField,
		); err != nil {
			r.logger.Printf("[ERROR] 扫描组织数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 查询 %d/%d 组织 (页面: %d/%d)，耗时: %v", len(organizations), total, page, (total+int(pageSize)-1)/int(pageSize), duration)

	totalPages := (total + int(pageSize) - 1) / int(pageSize)
	asOfDateValue := time.Now().Format("2006-01-02")
	if asOfDateParam.Valid {
		asOfDateValue = asOfDateParam.String
	}

	response := &model.OrganizationConnection{
		DataField: organizations,
		PaginationField: model.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TemporalField: model.TemporalInfo{
			AsOfDateField:        asOfDateValue,
			CurrentCountField:    len(organizations),
			FutureCountField:     0,
			HistoricalCountField: 0,
		},
	}

	return response, nil
}

// 单个组织查询 - 超快速索引查询
func (r *PostgreSQLRepository) GetOrganization(ctx context.Context, tenantID uuid.UUID, code string) (*model.Organization, error) {
	// 使用 idx_current_record_fast 索引
	query := `
        SELECT record_id, tenant_id, code, parent_code, name, unit_type, status, 
               level, path, sort_order, description, profile, created_at, updated_at,
               effective_date, end_date, is_current, change_reason,
               deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
        LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var org model.Organization
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 查询单个组织失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 单个组织查询，耗时: %v", duration)

	return &org, nil
}

// 极速时态查询 - 时间点查询（利用时态索引）
func (r *PostgreSQLRepository) GetOrganizationAtDate(ctx context.Context, tenantID uuid.UUID, code, date string) (*model.Organization, error) {
	// 使用计算的区间终点（computed_end_date），避免依赖物理 end_date 的准确性
	query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED'
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, profile, created_at, updated_at,
               effective_date, computed_end_date AS end_date, is_current, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $3::date 
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
        ORDER BY effective_date DESC, created_at DESC
        LIMIT 1`

	start := time.Now()
	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code, date)

	var org model.Organization
	var isTemporal bool
	err := row.Scan(
		&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
		&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
		&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
		&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField, &isTemporal,
		&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
		&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 时态查询失败: %v", err)
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 时态点查询 [%s @ %s]，耗时: %v", code, date, duration)

	return &org, nil
}

// 历史范围查询 - 窗口函数优化
func (r *PostgreSQLRepository) GetOrganizationHistory(ctx context.Context, tenantID uuid.UUID, code, fromDate, toDate string) ([]model.Organization, error) {
	// 历史范围查询：使用计算的区间终点（computed_end_date）并基于区间重叠选择
	query := `
        WITH hist AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date, end_date, is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
                LEAD(effective_date) OVER (PARTITION BY tenant_id, code ORDER BY effective_date) AS next_effective
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 
              AND status <> 'DELETED'
        ), proj AS (
            SELECT 
                record_id, tenant_id, code, parent_code, name, unit_type, status,
                level, path, sort_order, description, profile, created_at, updated_at,
                effective_date,
                COALESCE(end_date, (next_effective - INTERVAL '1 day')::date) AS computed_end_date,
                is_current, is_temporal, change_reason,
                deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
            FROM hist
        )
        SELECT 
            record_id, tenant_id, code, parent_code, name, unit_type, status,
            level, path, sort_order, description, profile, created_at, updated_at,
            effective_date, computed_end_date AS end_date, is_current, is_temporal, change_reason,
            deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason
        FROM proj
        WHERE effective_date <= $4::date
          AND (computed_end_date IS NULL OR computed_end_date >= $3::date)
        ORDER BY effective_date DESC, created_at DESC`

	start := time.Now()
	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, fromDate, toDate)
	if err != nil {
		r.logger.Printf("[ERROR] 历史范围查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []model.Organization
	for rows.Next() {
		var org model.Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描历史数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 历史查询 [%s: %s~%s] 返回 %d 条，耗时: %v", code, fromDate, toDate, len(organizations), duration)

	return organizations, nil
}

// 组织版本查询 - 按计划规范实现，返回指定code的全部版本
func (r *PostgreSQLRepository) GetOrganizationVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]model.Organization, error) {
	start := time.Now()

	// 构建查询 - 过滤条件：tenant_id = $tenant AND code = $code
	baseQuery := `
		SELECT record_id, tenant_id, code, parent_code, name, unit_type, status,
		       level, path, sort_order, description, profile, created_at, updated_at,
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
		r.logger.Printf("[ERROR] 组织版本查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var organizations []model.Organization
	for rows.Next() {
		var org model.Organization
		err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.PathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField,
			&org.HierarchyDepthField,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描组织版本数据失败: %v", err)
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 组织版本查询 [%s] 返回 %d 条版本，耗时: %v", code, len(organizations), duration)

	return organizations, nil
}

// 高级统计查询 - 利用PostgreSQL聚合优化
func (r *PostgreSQLRepository) GetOrganizationStats(ctx context.Context, tenantID uuid.UUID) (*model.OrganizationStats, error) {
	start := time.Now()

	// 使用单个复杂查询获取所有统计信息
	query := `
        WITH status_stats AS (
            SELECT 
                COUNT(*) as total_count,
                SUM(CASE WHEN status = 'ACTIVE' THEN 1 ELSE 0 END) as active_count,
                SUM(CASE WHEN status = 'INACTIVE' THEN 1 ELSE 0 END) as inactive_count,
                SUM(CASE WHEN status = 'PLANNED' THEN 1 ELSE 0 END) as planned_count,
                SUM(CASE WHEN status = 'DELETED' THEN 1 ELSE 0 END) as deleted_count
            FROM organization_units WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
        ),
        type_stats AS (
            SELECT unit_type, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY unit_type
        ),
        status_detail_stats AS (
            SELECT status, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY status
        ),
        level_stats AS (
            SELECT level, COUNT(*) as count
            FROM organization_units 
            WHERE tenant_id = $1 AND is_current = true AND status <> 'DELETED'
            GROUP BY level
        ),
        temporal_stats AS (
            SELECT 
                COUNT(*) as total_versions,
                COUNT(DISTINCT code) as unique_orgs,
                MIN(effective_date) as oldest_date,
                MAX(effective_date) as newest_date
            FROM organization_units WHERE tenant_id = $1 AND status <> 'DELETED'
        )
		SELECT 
			s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
			ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date,
			COALESCE(json_agg(DISTINCT jsonb_build_object('unitType', t.unit_type, 'count', t.count)) FILTER (WHERE t.unit_type IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('status', sd.status, 'count', sd.count)) FILTER (WHERE sd.status IS NOT NULL), '[]'),
			COALESCE(json_agg(DISTINCT jsonb_build_object('level', l.level, 'count', l.count)) FILTER (WHERE l.level IS NOT NULL), '[]')
		FROM status_stats s
		CROSS JOIN temporal_stats ts
		LEFT JOIN type_stats t ON true
		LEFT JOIN status_detail_stats sd ON true
		LEFT JOIN level_stats l ON true
		GROUP BY s.total_count, s.active_count, s.inactive_count, s.planned_count, s.deleted_count,
		         ts.total_versions, ts.unique_orgs, ts.oldest_date, ts.newest_date`

	row := r.db.QueryRowContext(ctx, query, tenantID.String())

	var stats model.OrganizationStats
	var totalVersions, uniqueOrgs int
	var oldestDate, newestDate time.Time
	var typeStatsJSON, statusStatsJSON, levelStatsJSON string

	err := row.Scan(
		&stats.TotalCountField, &stats.ActiveCountField, &stats.InactiveCountField,
		&stats.PlannedCountField, &stats.DeletedCountField,
		&totalVersions, &uniqueOrgs, &oldestDate, &newestDate,
		&typeStatsJSON, &statusStatsJSON, &levelStatsJSON,
	)
	if err != nil {
		r.logger.Printf("[ERROR] 统计查询失败: %v", err)
		return nil, err
	}

	// 解析JSON统计数据
	var typeStats []model.TypeCount
	if typeStatsJSON != "" {
		if err := json.Unmarshal([]byte(typeStatsJSON), &typeStats); err != nil {
			r.logger.Printf("解析typeStats失败: %v", err)
		}
	}
	stats.ByTypeField = typeStats

	var statusStats []model.StatusCount
	if statusStatsJSON != "" {
		if err := json.Unmarshal([]byte(statusStatsJSON), &statusStats); err != nil {
			r.logger.Printf("解析statusStats失败: %v", err)
		}
	}
	stats.ByStatusField = statusStats

	var levelStats []model.LevelCount
	if levelStatsJSON != "" {
		if err := json.Unmarshal([]byte(levelStatsJSON), &levelStats); err != nil {
			r.logger.Printf("解析levelStats失败: %v", err)
		}
	}
	stats.ByLevelField = levelStats

	// 时态统计
	avgPerOrg := 0.0
	if uniqueOrgs > 0 {
		avgPerOrg = float64(totalVersions) / float64(uniqueOrgs)
	}

	stats.TemporalStatsField = model.TemporalStats{
		TotalVersionsField:         totalVersions,
		AverageVersionsPerOrgField: avgPerOrg,
		OldestEffectiveDateField:   oldestDate.Format("2006-01-02"),
		NewestEffectiveDateField:   newestDate.Format("2006-01-02"),
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 统计查询完成，耗时: %v", duration)

	return &stats, nil
}

// 高级层级结构查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationHierarchy(ctx context.Context, tenantID uuid.UUID, code string) (*model.OrganizationHierarchyData, error) {
	start := time.Now()

	// 使用PostgreSQL递归CTE查询完整层级信息
	query := `
        WITH RECURSIVE hierarchy_info AS (
            -- 获取目标组织
            SELECT
                code,
                name,
                level,
                parent_code,
                1 AS hierarchy_depth
            FROM organization_units
            WHERE tenant_id = $1
              AND code = $2
              AND is_current = true
              AND status <> 'DELETED'

            UNION ALL

            -- 递归获取父级信息
            SELECT
                o.code,
                o.name,
                o.level,
                o.parent_code,
                h.hierarchy_depth + 1
            FROM organization_units o
            INNER JOIN hierarchy_info h ON o.code = h.parent_code
            WHERE o.tenant_id = $1
              AND o.is_current = true
              AND o.status <> 'DELETED'
        ),
        aggregated_paths AS (
            SELECT
                '/' || string_agg(code, '/' ORDER BY hierarchy_depth DESC) AS full_code_path,
                '/' || string_agg(name, '/' ORDER BY hierarchy_depth DESC) AS full_name_path,
                COALESCE(
                    array_agg(code ORDER BY hierarchy_depth DESC) FILTER (WHERE hierarchy_depth > 1),
                    ARRAY[]::text[]
                ) AS parent_chain
            FROM hierarchy_info
        ),
        target_info AS (
            SELECT *
            FROM hierarchy_info
            WHERE code = $2
            LIMIT 1
        ),
        children_count AS (
            SELECT COUNT(*) AS count
            FROM organization_units
            WHERE tenant_id = $1
              AND parent_code = $2
              AND is_current = true
              AND status <> 'DELETED'
        )
        SELECT
            t.code,
            t.name,
            t.level,
            t.hierarchy_depth,
            ap.full_code_path,
            ap.full_name_path,
            ap.parent_chain,
            c.count AS children_count,
            (t.parent_code IS NULL) AS is_root,
            (c.count = 0) AS is_leaf
        FROM target_info t
        CROSS JOIN aggregated_paths ap
        CROSS JOIN children_count c
        LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, tenantID.String(), code)

	var hierarchy model.OrganizationHierarchyData
	var parentChain []string

	err := row.Scan(
		&hierarchy.CodeField,
		&hierarchy.NameField,
		&hierarchy.LevelField,
		&hierarchy.HierarchyDepthField,
		&hierarchy.CodePathField,
		&hierarchy.NamePathField,
		pq.Array(&parentChain),
		&hierarchy.ChildrenCountField,
		&hierarchy.IsRootField,
		&hierarchy.IsLeafField,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 层级结构查询失败: %v", err)
		return nil, err
	}

	hierarchy.ParentChainField = parentChain

	duration := time.Since(start)
	r.logger.Printf("[PERF] 层级结构查询完成，耗时: %v", duration)

	return &hierarchy, nil
}

// 组织子树查询 - 严格遵循API规范v4.2.1
func (r *PostgreSQLRepository) GetOrganizationSubtree(ctx context.Context, tenantID uuid.UUID, code string, maxDepth int) (*model.OrganizationSubtreeData, error) {
	start := time.Now()

	// 使用PostgreSQL递归CTE查询子树结构，限制深度
	query := `
        WITH RECURSIVE subtree AS (
            -- 根节点
            SELECT 
                code, name, level, 
                COALESCE(hierarchy_depth, level) as hierarchy_depth,
                COALESCE(code_path, '/' || code) as code_path,
                COALESCE(name_path, '/' || name) as name_path,
                parent_code,
                0 as depth_from_root
            FROM organization_units 
            WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
            
            UNION ALL
            
            -- 递归查询子节点
            SELECT 
                o.code, o.name, o.level,
                o.hierarchy_depth, o.code_path, o.name_path, o.parent_code,
                s.depth_from_root + 1
            FROM organization_units o
            INNER JOIN subtree s ON o.parent_code = s.code
            WHERE o.tenant_id = $1 AND o.is_current = true AND o.status <> 'DELETED'
              AND s.depth_from_root < $3
        )
		SELECT code, name, level, hierarchy_depth, code_path, name_path, parent_code
		FROM subtree 
		ORDER BY level, code`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code, maxDepth)
	if err != nil {
		r.logger.Printf("[ERROR] 子树查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	// 构建树形结构
	nodeMap := make(map[string]*model.OrganizationSubtreeData)
	var root *model.OrganizationSubtreeData

	for rows.Next() {
		node := &model.OrganizationSubtreeData{}
		var parentCode *string

		err := rows.Scan(
			&node.CodeField, &node.NameField, &node.LevelField, &node.HierarchyDepthField,
			&node.CodePathField, &node.NamePathField, &parentCode,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描子树数据失败: %v", err)
			return nil, err
		}

		node.ChildrenField = []model.OrganizationSubtreeData{}
		nodeMap[node.CodeField] = node

		if node.CodeField == code {
			root = node
		}
	}

	// 构建父子关系
	for _, node := range nodeMap {
		if root != nil && node.CodeField != code {
			for _, parent := range nodeMap {
				if node.CodeField == parent.CodeField {
					continue
				}
				if node.CodePathField == nil || parent.CodePathField == nil {
					continue
				}
				nodePath := *node.CodePathField
				parentPath := *parent.CodePathField
				if strings.HasPrefix(nodePath, parentPath+"/") {
					parentDepth := strings.Count(parentPath, "/")
					nodeDepth := strings.Count(nodePath, "/")
					if nodeDepth == parentDepth+1 {
						parent.ChildrenField = append(parent.ChildrenField, *node)
						break
					}
				}
			}
		}
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 子树查询完成，返回 %d 节点，耗时: %v", len(nodeMap), duration)

	return root, nil
}

// 审计历史查询 - v4.6.0 基于record_id精确查询 + 租户隔离
func (r *PostgreSQLRepository) GetAuditHistory(ctx context.Context, tenantId uuid.UUID, recordId string, startDate, endDate, operation, userId *string, limit int) ([]model.AuditRecordData, error) {
	start := time.Now()

	recordUUID, err := uuid.Parse(recordId)
	if err != nil {
		r.logger.Printf("[ERROR] 无效的 recordId: %s", recordId)
		return nil, fmt.Errorf("INVALID_RECORD_ID")
	}

	// 构建查询条件 - 基于record_id查询，包含完整变更信息，强制租户隔离
	baseQuery := `
		SELECT
			id as audit_id,
			resource_id as record_id,
			event_type as operation_type,
			actor_id as operated_by_id,
			CASE WHEN business_context->>'actor_name' IS NOT NULL
				THEN business_context->>'actor_name'
				ELSE actor_id
			END as operated_by_name,
			CASE WHEN changes IS NOT NULL
				THEN jsonb_build_object(
					'operationSummary', COALESCE(action_name, event_type, 'UNKNOWN'),
					'totalChanges', jsonb_array_length(changes),
					'keyChanges', changes
				)::text
				ELSE jsonb_build_object(
					'operationSummary', COALESCE(action_name, event_type, 'UNKNOWN'),
					'totalChanges', 0,
					'keyChanges', jsonb_build_array()
				)::text
			END as changes_summary,
			business_context->>'operation_reason' as operation_reason,
			timestamp,
			request_data::text as before_data,
			response_data::text as after_data,
			CASE WHEN changes IS NOT NULL AND jsonb_typeof(changes) = 'array'
				THEN (
					SELECT jsonb_agg(DISTINCT elem->>'field')
					FROM jsonb_array_elements(changes) AS elem
					WHERE elem->>'field' IS NOT NULL
				)
				ELSE '[]'::jsonb
			END::text as modified_fields,
			COALESCE(changes, '[]'::jsonb)::text as detailed_changes
		FROM audit_logs
		WHERE tenant_id = $1::uuid AND resource_id::uuid = $2::uuid AND resource_type = 'ORGANIZATION'`

	args := []interface{}{tenantId, recordUUID}
	argIndex := 3

	// 日期范围过滤
	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp >= $%d::timestamp", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND timestamp <= $%d::timestamp", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// 操作类型过滤
	if operation != nil {
		baseQuery += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, strings.ToUpper(*operation))
		argIndex++
	}

	// 操作人过滤
	if userId != nil {
		baseQuery += fmt.Sprintf(" AND actor_id = $%d", argIndex)
		args = append(args, *userId)
		argIndex++
	}

	// 排序和限制
	finalQuery := baseQuery + fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		r.logger.Printf("[ERROR] 审计历史查询失败: %v", err)
		return nil, err
	}
	defer rows.Close()

	var auditRecords []model.AuditRecordData
	if r.auditConfig.LegacyMode {
		auditRecords, err = r.processAuditRowsLegacy(rows)
	} else {
		auditRecords, err = r.processAuditRowsStrict(rows)
	}
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] record_id审计查询完成，返回 %d 条记录，耗时: %v", len(auditRecords), duration)

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsLegacy(rows *sql.Rows) ([]model.AuditRecordData, error) {
	var auditRecords []model.AuditRecordData
	for rows.Next() {
		var record model.AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		if modifiedFieldsJSON.Valid && modifiedFieldsJSON.String != "[]" {
			var modifiedFields []string
			if err := json.Unmarshal([]byte(modifiedFieldsJSON.String), &modifiedFields); err == nil {
				record.ModifiedFieldsField = modifiedFields
			}
		}

		if detailedChangesJSON.Valid && detailedChangesJSON.String != "[]" {
			var changesArray []map[string]interface{}
			if err := json.Unmarshal([]byte(detailedChangesJSON.String), &changesArray); err == nil {
				for _, changeMap := range changesArray {
					fieldChange := model.FieldChangeData{
						FieldField:    fmt.Sprintf("%v", changeMap["field"]),
						OldValueField: changeMap["oldValue"],
						NewValueField: changeMap["newValue"],
						DataTypeField: fmt.Sprintf("%v", changeMap["dataType"]),
					}
					record.ChangesField = append(record.ChangesField, fieldChange)
				}
			}
		}

		record.OperatedByField = model.OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func (r *PostgreSQLRepository) processAuditRowsStrict(rows *sql.Rows) ([]model.AuditRecordData, error) {
	var auditRecords []model.AuditRecordData
	for rows.Next() {
		var record model.AuditRecordData
		var operatedById, operatedByName string
		var beforeData, afterData, modifiedFieldsJSON, detailedChangesJSON sql.NullString

		record.ModifiedFieldsField = make([]string, 0)
		record.ChangesField = make([]model.FieldChangeData, 0)

		err := rows.Scan(
			&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
			&operatedById, &operatedByName,
			&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
			&beforeData, &afterData, &modifiedFieldsJSON, &detailedChangesJSON,
		)
		if err != nil {
			r.logger.Printf("[ERROR] 扫描审计记录失败: %v", err)
			return nil, err
		}

		if beforeData.Valid {
			record.BeforeDataField = &beforeData.String
		}
		if afterData.Valid {
			record.AfterDataField = &afterData.String
		}

		rawModified := ""
		if modifiedFieldsJSON.Valid {
			rawModified = modifiedFieldsJSON.String
		}
		sanitizedModified, modifiedIssues, modErr := sanitizeModifiedFields(rawModified)
		if modErr == nil {
			record.ModifiedFieldsField = sanitizedModified
		}

		rawChanges := ""
		if detailedChangesJSON.Valid {
			rawChanges = detailedChangesJSON.String
		}
		sanitizedChanges, changeIssues, changeErr := sanitizeChanges(rawChanges)
		if changeErr == nil {
			record.ChangesField = sanitizedChanges
		}

		issues := make([]string, 0, len(modifiedIssues)+len(changeIssues))
		issues = append(issues, modifiedIssues...)
		issues = append(issues, changeIssues...)

		hasHardError := false
		if modErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("modified_fields JSON 无效: %v", modErr))
		}
		if changeErr != nil {
			hasHardError = true
			issues = append(issues, fmt.Sprintf("changes JSON 无效: %v", changeErr))
		}

		if len(issues) > 0 {
			r.logger.Printf("[WARN] 审计记录数据异常 audit_id=%s: %s", record.AuditIDField, strings.Join(issues, "; "))
			if r.auditConfig.StrictValidation {
				if hasHardError && !r.auditConfig.AllowFallback {
					return nil, fmt.Errorf("AUDIT_HISTORY_VALIDATION_FAILED")
				}
				if r.registerValidationFailure() {
					return nil, fmt.Errorf("AUDIT_HISTORY_CIRCUIT_OPEN")
				}
			}
		} else if r.auditConfig.StrictValidation {
			r.registerValidationSuccess()
		}

		record.OperatedByField = model.OperatedByData{
			IDField:   operatedById,
			NameField: operatedByName,
		}

		auditRecords = append(auditRecords, record)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return auditRecords, nil
}

func sanitizeModifiedFields(raw string) ([]string, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]string, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]string, 0), []string{"modified_fields 为 null，已替换为空数组"}, nil
	}

	var rawArray []interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]string, 0), nil, err
	}

	sanitized := make([]string, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, item := range rawArray {
		if item == nil {
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 为 null，已忽略", idx))
			continue
		}
		switch v := item.(type) {
		case string:
			sanitized = append(sanitized, v)
		default:
			sanitized = append(sanitized, fmt.Sprintf("%v", v))
			issues = append(issues, fmt.Sprintf("modified_fields[%d] 非字符串，已转换", idx))
		}
	}

	return sanitized, issues, nil
}

func sanitizeChanges(raw string) ([]model.FieldChangeData, []string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return make([]model.FieldChangeData, 0), nil, nil
	}
	if trimmed == "null" {
		return make([]model.FieldChangeData, 0), []string{"changes 为 null，已替换为空数组"}, nil
	}

	var rawArray []map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawArray); err != nil {
		return make([]model.FieldChangeData, 0), nil, err
	}

	sanitized := make([]model.FieldChangeData, 0, len(rawArray))
	issues := make([]string, 0)
	for idx, entry := range rawArray {
		if entry == nil {
			issues = append(issues, fmt.Sprintf("changes[%d] 为空对象，已跳过", idx))
			continue
		}

		fieldVal, ok := entry["field"]
		if !ok {
			issues = append(issues, fmt.Sprintf("changes[%d] 缺少 field，已跳过", idx))
			continue
		}
		field := strings.TrimSpace(fmt.Sprintf("%v", fieldVal))
		if field == "" {
			issues = append(issues, fmt.Sprintf("changes[%d] field 为空，已跳过", idx))
			continue
		}

		dataType := "unknown"
		if dtVal, ok := entry["dataType"]; ok {
			if dtStr, ok := dtVal.(string); ok && strings.TrimSpace(dtStr) != "" {
				dataType = dtStr
			} else {
				issues = append(issues, fmt.Sprintf("changes[%d] dataType 非字符串，使用 unknown", idx))
			}
		} else {
			issues = append(issues, fmt.Sprintf("changes[%d] 缺少 dataType，使用 unknown", idx))
		}

		fieldChange := model.FieldChangeData{
			FieldField:    field,
			DataTypeField: dataType,
			OldValueField: normalizeChangeValue(entry["oldValue"]),
			NewValueField: normalizeChangeValue(entry["newValue"]),
		}
		sanitized = append(sanitized, fieldChange)
	}

	return sanitized, issues, nil
}

func normalizeChangeValue(val interface{}) interface{} {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case json.Number:
		return v.String()
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(bytes)
	}
}

func (r *PostgreSQLRepository) registerValidationFailure() bool {
	count := atomic.AddInt32(&r.validationFailureCount, 1)
	if r.auditConfig.CircuitBreakerThreshold > 0 && count >= r.auditConfig.CircuitBreakerThreshold {
		r.logger.Printf("[ALERT] 审计历史验证失败次数达到阈值 (%d/%d)，触发熔断", count, r.auditConfig.CircuitBreakerThreshold)
		return true
	}
	return false
}

func (r *PostgreSQLRepository) registerValidationSuccess() {
	if atomic.LoadInt32(&r.validationFailureCount) != 0 {
		atomic.StoreInt32(&r.validationFailureCount, 0)
	}
}

// 单条审计记录查询 - v4.6.0
func (r *PostgreSQLRepository) GetAuditLog(ctx context.Context, auditId string) (*model.AuditRecordData, error) {
	start := time.Now()

	query := `
        SELECT 
            id as audit_id, 
            resource_id as record_id, 
            event_type as operation_type,
            actor_id as operated_by_id, 
            CASE WHEN business_context->>'actor_name' IS NOT NULL 
                THEN business_context->>'actor_name' 
                ELSE actor_id 
            END as operated_by_name,
            CASE WHEN changes IS NOT NULL 
                THEN changes::text 
                ELSE '{"operationSummary":"' || action_name || '","totalChanges":0,"keyChanges":[]}' 
            END as changes_summary,
            business_context->>'operation_reason' as operation_reason,
            timestamp,
            before_data::text as before_data, 
            after_data::text as after_data
        FROM audit_logs 
        WHERE id = $1::uuid AND resource_type = 'ORGANIZATION' AND tenant_id = $2::uuid
        LIMIT 1`

	tenantID := auth.GetTenantID(ctx)
	if tenantID == "" {
		r.logger.Printf("[AUTH] 缺少租户ID，拒绝单条审计记录查询")
		return nil, fmt.Errorf("TENANT_REQUIRED")
	}

	row := r.db.QueryRowContext(ctx, query, auditId, tenantID)

	var record model.AuditRecordData
	var operatedById, operatedByName string
	var beforeData, afterData sql.NullString

	err := row.Scan(
		&record.AuditIDField, &record.RecordIDField, &record.OperationTypeField,
		&operatedById, &operatedByName,
		&record.ChangesSummaryField, &record.OperationReasonField, &record.TimestampField,
		&beforeData, &afterData,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Printf("[ERROR] 单条审计记录查询失败: %v", err)
		return nil, err
	}

	// 正确处理JSONB字段
	if beforeData.Valid {
		record.BeforeDataField = &beforeData.String
	}
	if afterData.Valid {
		record.AfterDataField = &afterData.String
	}

	// 构建操作人信息
	record.OperatedByField = model.OperatedByData{
		IDField:   operatedById,
		NameField: operatedByName,
	}

	duration := time.Since(start)
	r.logger.Printf("[PERF] 单条审计记录查询完成，耗时: %v", duration)

	return &record, nil
}

// GraphQL解析器 - 极简高效
