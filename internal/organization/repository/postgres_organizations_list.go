package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/dto"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// 极速当前组织查询 - 利用部分索引 idx_current_organizations_list (API契约v4.2.1)
func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *dto.OrganizationFilter, pagination *dto.PaginationInput) (*dto.OrganizationConnection, error) {
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

	logFields := pkglogger.Fields{
		"tenantId": tenantID.String(),
		"page":     page,
		"pageSize": pageSize,
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
		logFields["includeDisabledAncestors"] = includeDisabledAncestors
		if filter.Status != nil {
			status = strings.TrimSpace(*filter.Status)
			if status != "" {
				logFields["status"] = status
			}
		}
		if filter.SearchText != nil {
			searchText = strings.TrimSpace(*filter.SearchText)
			if searchText != "" {
				logFields["searchText"] = true
			}
		}
		if filter.UnitType != nil {
			unitType = strings.TrimSpace(*filter.UnitType)
			if unitType != "" {
				logFields["unitType"] = unitType
			}
		}
		if filter.ParentCode != nil {
			parentCode = strings.TrimSpace(*filter.ParentCode)
			if parentCode != "" {
				logFields["parentCode"] = parentCode
			}
		}
		if filter.AsOfDate != nil {
			if trimmed := strings.TrimSpace(*filter.AsOfDate); trimmed != "" {
				asOfDateParam = sql.NullString{String: trimmed, Valid: true}
				logFields["asOfDate"] = trimmed
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
			if len(excludeCodes) > 0 {
				logFields["excludeCodes"] = len(excludeCodes)
			}
		}
		if filter.Codes != nil {
			for _, code := range *filter.Codes {
				if trimmed := strings.TrimSpace(code); trimmed != "" {
					includeCodes = append(includeCodes, trimmed)
				}
			}
			if len(includeCodes) > 0 {
				logFields["includeCodes"] = len(includeCodes)
			}
		}
	}

	log := r.loggerFor("organization.list", logFields)

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
        level, sort_order, description, profile, created_at, updated_at,
        effective_date, end_date, is_current, change_reason,
        deleted_at, deleted_by, deletion_reason, suspended_at, suspended_by, suspension_reason,
        COALESCE(code_path, '/' || code) AS code_path,
        COALESCE(name_path, '/' || name) AS name_path
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
       lv.level, lv.code_path, lv.name_path, lv.sort_order, lv.description, lv.profile, lv.created_at, lv.updated_at,
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
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization list count query failed")
		return nil, err
	}

	orderClause := fmt.Sprintf(" ORDER BY COALESCE(lv.sort_order, 0) NULLS LAST, lv.code LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	dataQuery := cte + baseSelect + whereConditions + orderClause
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		log.WithFields(pkglogger.Fields{"error": err}).Error("organization list query failed")
		return nil, err
	}
	defer rows.Close()

	var organizations []dto.Organization
	for rows.Next() {
		var org dto.Organization
		if err := rows.Scan(
			&org.RecordIDField, &org.TenantIDField, &org.CodeField, &org.ParentCodeField, &org.NameField,
			&org.UnitTypeField, &org.StatusField, &org.LevelField, &org.CodePathField, &org.NamePathField, &org.SortOrderField,
			&org.DescriptionField, &org.ProfileField, &org.CreatedAtField, &org.UpdatedAtField,
			&org.EffectiveDateField, &org.EndDateField, &org.IsCurrentField,
			&org.ChangeReasonField, &org.DeletedAtField, &org.DeletedByField, &org.DeletionReasonField,
			&org.SuspendedAtField, &org.SuspendedByField, &org.SuspensionReasonField, &org.ChildrenCountField,
		); err != nil {
			log.WithFields(pkglogger.Fields{"error": err}).Error("organization list scan failed")
			return nil, err
		}
		organizations = append(organizations, org)
	}

	duration := time.Since(start)
	log.WithFields(pkglogger.Fields{
		"result_count": len(organizations),
		"total_count":  total,
		"duration_ms":  duration.Milliseconds(),
	}).Info("organization list query succeeded")

	totalPages := (total + int(pageSize) - 1) / int(pageSize)
	asOfDateValue := time.Now().Format("2006-01-02")
	if asOfDateParam.Valid {
		asOfDateValue = asOfDateParam.String
	}

	response := &dto.OrganizationConnection{
		DataField: organizations,
		PaginationField: dto.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TemporalField: dto.TemporalInfo{
			AsOfDateField:        asOfDateValue,
			CurrentCountField:    len(organizations),
			FutureCountField:     0,
			HistoricalCountField: 0,
		},
	}

	return response, nil
}
