package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func toStringSlice[T ~string](values []T) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}
	return result
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func (r *PostgreSQLRepository) GetPositions(ctx context.Context, tenantID uuid.UUID, filter *model.PositionFilterInput, pagination *model.PaginationInput, sorting []model.PositionSortInput) (*model.PositionConnection, error) {
	page := int32(1)
	pageSize := int32(25)
	if pagination != nil {
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
			if pageSize > 200 {
				pageSize = 200
			}
		}
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	args := []interface{}{tenantID.String()}
	argIndex := 2

	whereParts := []string{"p.tenant_id = $1", "p.is_current = true", "p.status <> 'DELETED'"}

	if filter != nil {
		if organizationCode := filter.OrganizationCode; organizationCode != nil {
			whereParts = append(whereParts, fmt.Sprintf("p.organization_code = $%d", argIndex))
			args = append(args, strings.TrimSpace(*organizationCode))
			argIndex++
		}
		if positionCodes := filter.PositionCodes; positionCodes != nil && len(*positionCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*positionCodes)))
			argIndex++
		}
		if status := filter.Status; status != nil && strings.TrimSpace(*status) != "" {
			whereParts = append(whereParts, fmt.Sprintf("p.status = $%d", argIndex))
			args = append(args, strings.ToUpper(strings.TrimSpace(*status)))
			argIndex++
		}
		if jobFamilyGroupCodes := filter.JobFamilyGroupCodes; jobFamilyGroupCodes != nil && len(*jobFamilyGroupCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_family_group_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobFamilyGroupCodes)))
			argIndex++
		}
		if jobFamilyCodes := filter.JobFamilyCodes; jobFamilyCodes != nil && len(*jobFamilyCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_family_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobFamilyCodes)))
			argIndex++
		}
		if jobRoleCodes := filter.JobRoleCodes; jobRoleCodes != nil && len(*jobRoleCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_role_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobRoleCodes)))
			argIndex++
		}
		if jobLevelCodes := filter.JobLevelCodes; jobLevelCodes != nil && len(*jobLevelCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_level_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobLevelCodes)))
			argIndex++
		}
		if positionTypes := filter.PositionTypes; positionTypes != nil && len(*positionTypes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.position_type = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(*positionTypes))
			argIndex++
		}
		if employmentTypes := filter.EmploymentTypes; employmentTypes != nil && len(*employmentTypes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.employment_type = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(*employmentTypes))
			argIndex++
		}
		if rangeInput := filter.EffectiveRange; rangeInput != nil {
			if rangeInput.From != nil && strings.TrimSpace(*rangeInput.From) != "" {
				whereParts = append(whereParts, fmt.Sprintf("p.effective_date >= $%d", argIndex))
				args = append(args, strings.TrimSpace(*rangeInput.From))
				argIndex++
			}
			if rangeInput.To != nil && strings.TrimSpace(*rangeInput.To) != "" {
				whereParts = append(whereParts, fmt.Sprintf("p.effective_date <= $%d", argIndex))
				args = append(args, strings.TrimSpace(*rangeInput.To))
				argIndex++
			}
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	orderClause := "ORDER BY p.effective_date DESC, p.code ASC"
	if len(sorting) > 0 {
		orderParts := make([]string, 0, len(sorting))
		for _, sort := range sorting {
			field := strings.ToUpper(strings.TrimSpace(sort.Field))
			var column string
			switch field {
			case "CODE":
				column = "p.code"
			case "TITLE":
				column = "p.title"
			case "EFFECTIVE_DATE":
				column = "p.effective_date"
			case "STATUS":
				column = "p.status"
			default:
				continue
			}
			dir := strings.ToUpper(strings.TrimSpace(sort.Direction))
			if dir != "DESC" {
				dir = "ASC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", column, dir))
		}
		if len(orderParts) > 0 {
			orderClause = "ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM positions p %s`, whereClause)
	countArgs := append([]interface{}{}, args...)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count positions: %w", err)
	}

	selectQuery := fmt.Sprintf(`
SELECT
    p.record_id::text,
    p.tenant_id::text,
    p.code,
    p.title,
    p.job_profile_code,
    p.job_profile_name,
    p.job_family_group_code,
    p.job_family_code,
    p.job_role_code,
    p.job_level_code,
    p.organization_code,
    p.position_type,
    p.employment_type,
    p.grade_level,
    p.headcount_capacity,
    p.headcount_in_use,
    p.reports_to_position_code,
    p.status,
    p.effective_date,
    p.end_date,
    p.is_current,
    p.created_at,
    p.updated_at,
    p.job_family_group_name,
    p.job_family_name,
    p.job_role_name,
    p.job_level_name,
    p.organization_name
FROM positions p
%s
%s
LIMIT $%d OFFSET $%d`, whereClause, orderClause, argIndex, argIndex+1)

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	positions := make([]model.Position, 0, len(queryArgs))
	for rows.Next() {
		pos, scanErr := scanPosition(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		positions = append(positions, *pos)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions: %w", err)
	}

	totalPages := (total + int(pageSize) - 1) / int(pageSize)
	edges := make([]model.PositionEdge, 0, len(positions))
	for _, pos := range positions {
		edges = append(edges, model.PositionEdge{
			CursorField: pos.RecordIDField,
			NodeField:   pos,
		})
	}

	connection := &model.PositionConnection{
		EdgesField: edges,
		DataField:  positions,
		PaginationField: model.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TotalCountField: total,
	}

	return connection, nil
}

func (r *PostgreSQLRepository) GetPositionByCode(ctx context.Context, tenantID uuid.UUID, code string, asOfDate *string) (*model.Position, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(code)}
	argIndex := 3

	where := "WHERE p.tenant_id = $1 AND p.code = $2"
	if asOfDate != nil && strings.TrimSpace(*asOfDate) != "" {
		where += fmt.Sprintf(" AND p.effective_date <= $%d AND (p.end_date IS NULL OR p.end_date > $%d)", argIndex, argIndex)
		args = append(args, strings.TrimSpace(*asOfDate))
		argIndex++
	} else {
		where += " AND p.is_current = true"
	}

	query := fmt.Sprintf(`
SELECT
    p.record_id::text,
    p.tenant_id::text,
    p.code,
    p.title,
    p.job_profile_code,
    p.job_profile_name,
    p.job_family_group_code,
    p.job_family_code,
    p.job_role_code,
    p.job_level_code,
    p.organization_code,
    p.position_type,
    p.employment_type,
    p.grade_level,
    p.headcount_capacity,
    p.headcount_in_use,
    p.reports_to_position_code,
    p.status,
    p.effective_date,
    p.end_date,
    p.is_current,
    p.created_at,
    p.updated_at,
    p.job_family_group_name,
    p.job_family_name,
    p.job_role_name,
    p.job_level_name,
    p.organization_name
FROM positions p
%s
ORDER BY p.effective_date DESC, p.created_at DESC
LIMIT 1
`, where)

	row := r.db.QueryRowContext(ctx, query, args...)
	pos, err := scanPosition(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if err := r.populatePositionAssignments(ctx, tenantID, pos); err != nil {
		return nil, fmt.Errorf("load position assignments: %w", err)
	}
	return pos, nil
}

func (r *PostgreSQLRepository) GetPositionTimeline(ctx context.Context, tenantID uuid.UUID, code string, startDate, endDate *string) ([]model.PositionTimelineEntry, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(code)}
	argIndex := 3
	whereParts := []string{"p.tenant_id = $1", "p.code = $2"}
	assignmentWhereParts := []string{"pa.tenant_id = $1", "pa.position_code = $2"}

	if startDate != nil && strings.TrimSpace(*startDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("p.effective_date >= $%d", argIndex))
		assignmentWhereParts = append(assignmentWhereParts, fmt.Sprintf("pa.effective_date >= $%d", argIndex))
		args = append(args, strings.TrimSpace(*startDate))
		argIndex++
	}
	if endDate != nil && strings.TrimSpace(*endDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("p.effective_date <= $%d", argIndex))
		assignmentWhereParts = append(assignmentWhereParts, fmt.Sprintf("pa.effective_date <= $%d", argIndex))
		args = append(args, strings.TrimSpace(*endDate))
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(whereParts, " AND ")
	assignmentClause := "WHERE " + strings.Join(assignmentWhereParts, " AND ")

	query := fmt.Sprintf(`
WITH timeline AS (
    SELECT
        p.record_id::text AS record_id,
        p.status,
        p.title,
        p.effective_date,
        p.end_date,
        p.is_current,
        p.operation_reason AS change_reason,
        'POSITION_VERSION'::text AS timeline_category,
        NULL::text AS assignment_type,
        NULL::text AS assignment_status
    FROM positions p
    %s
    UNION ALL
    SELECT
        pa.assignment_id::text AS record_id,
        pa.assignment_status AS status,
        pa.employee_name AS title,
        pa.effective_date,
        COALESCE(pa.end_date, pa.acting_until) AS end_date,
        pa.is_current,
        pa.notes AS change_reason,
        'POSITION_ASSIGNMENT'::text AS timeline_category,
        pa.assignment_type,
        pa.assignment_status
    FROM position_assignments pa
    %s
)
SELECT
    record_id,
    status,
    title,
    effective_date,
    end_date,
    is_current,
    change_reason,
    timeline_category,
    assignment_type,
    assignment_status
FROM timeline
ORDER BY effective_date ASC, record_id ASC
`, whereClause, assignmentClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query position timeline: %w", err)
	}
	defer rows.Close()

	result := make([]model.PositionTimelineEntry, 0)
	for rows.Next() {
		var entry model.PositionTimelineEntry
		var endDate sql.NullTime
		var changeReason sql.NullString
		var assignmentType sql.NullString
		var assignmentStatus sql.NullString
		if err := rows.Scan(
			&entry.RecordIDField,
			&entry.StatusField,
			&entry.TitleField,
			&entry.EffectiveDateField,
			&endDate,
			&entry.IsCurrentField,
			&changeReason,
			&entry.TimelineCategoryField,
			&assignmentType,
			&assignmentStatus,
		); err != nil {
			return nil, fmt.Errorf("scan timeline entry: %w", err)
		}
		if endDate.Valid {
			entry.EndDateField = &endDate.Time
		}
		if changeReason.Valid {
			entry.ChangeReasonField = &changeReason.String
		}
		if assignmentType.Valid {
			val := strings.ToUpper(strings.TrimSpace(assignmentType.String))
			entry.AssignmentTypeField = &val
		}
		if assignmentStatus.Valid {
			val := strings.ToUpper(strings.TrimSpace(assignmentStatus.String))
			entry.AssignmentStatusField = &val
		}
		if strings.TrimSpace(entry.TimelineCategoryField) == "" {
			entry.TimelineCategoryField = "POSITION_VERSION"
		}
		result = append(result, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLRepository) GetVacantPositions(ctx context.Context, tenantID uuid.UUID, organizationCode *string, positionType *string, includeSubordinates bool) ([]model.Position, error) {
	args := []interface{}{tenantID.String()}
	argIndex := 2

	whereParts := []string{"p.tenant_id = $1", "p.is_current = true", "(p.status = 'VACANT' OR p.headcount_capacity > p.headcount_in_use)", "p.status <> 'DELETED'"}
	joins := ""

	if organizationCode != nil && strings.TrimSpace(*organizationCode) != "" {
		orgCode := strings.TrimSpace(*organizationCode)
		if includeSubordinates {
			joins = `
WITH org_scope AS (
    SELECT DISTINCT ON (code)
        code,
        COALESCE(code_path, '/' || code) AS code_path
    FROM organization_units
    WHERE tenant_id = $1
      AND code = $2
      AND is_current = true
    ORDER BY code, effective_date DESC, created_at DESC
)
SELECT
    p.record_id::text,
    p.tenant_id::text,
    p.code,
    p.title,
    p.job_profile_code,
    p.job_profile_name,
    p.job_family_group_code,
    p.job_family_code,
    p.job_role_code,
    p.job_level_code,
    p.organization_code,
    p.position_type,
    p.employment_type,
    p.grade_level,
    p.headcount_capacity,
    p.headcount_in_use,
    p.reports_to_position_code,
    p.status,
    p.effective_date,
    p.end_date,
    p.is_current,
    p.created_at,
    p.updated_at,
    p.job_family_group_name,
    p.job_family_name,
    p.job_role_name,
    p.job_level_name,
    p.organization_name
FROM positions p
JOIN organization_units ou ON ou.tenant_id = p.tenant_id AND ou.code = p.organization_code AND ou.is_current = true
CROSS JOIN org_scope scope
WHERE p.tenant_id = $1
  AND p.is_current = true
  AND p.status <> 'DELETED'
  AND (p.status = 'VACANT' OR p.headcount_capacity > p.headcount_in_use)
  AND (
      ou.code = scope.code OR ou.code_path LIKE scope.code_path || '/%'
  )`
			// when includeSubordinates with org scope, we build entire query directly
			args = append(args, orgCode)
			if positionType != nil && strings.TrimSpace(*positionType) != "" {
				joins += fmt.Sprintf(" AND p.position_type = $%d", argIndex+1)
				args = append(args, strings.ToUpper(strings.TrimSpace(*positionType)))
			}
			joins += "\nORDER BY p.code"
			rows, err := r.db.QueryContext(ctx, joins, args...)
			if err != nil {
				return nil, fmt.Errorf("query vacant positions: %w", err)
			}
			defer rows.Close()

			result := make([]model.Position, 0)
			for rows.Next() {
				pos, scanErr := scanPosition(rows)
				if scanErr != nil {
					return nil, scanErr
				}
				result = append(result, *pos)
			}
			if err := rows.Err(); err != nil {
				return nil, fmt.Errorf("iterate vacant positions: %w", err)
			}
			return result, nil
		}

		whereParts = append(whereParts, fmt.Sprintf("p.organization_code = $%d", argIndex))
		args = append(args, orgCode)
		argIndex++
	}

	if positionType != nil && strings.TrimSpace(*positionType) != "" && includeSubordinates {
		// already handled above when building custom query
	} else if positionType != nil && strings.TrimSpace(*positionType) != "" {
		whereParts = append(whereParts, fmt.Sprintf("p.position_type = $%d", argIndex))
		args = append(args, strings.ToUpper(strings.TrimSpace(*positionType)))
		argIndex++
	}

	if includeSubordinates && (organizationCode == nil || strings.TrimSpace(*organizationCode) == "") {
		// no organization code provided; includeSubordinates irrelevant
		includeSubordinates = false
	}

	if includeSubordinates {
		// already handled via custom query above, so skip here
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	query := fmt.Sprintf(`
SELECT
    p.record_id::text,
    p.tenant_id::text,
    p.code,
    p.title,
    p.job_profile_code,
    p.job_profile_name,
    p.job_family_group_code,
    p.job_family_code,
    p.job_role_code,
    p.job_level_code,
    p.organization_code,
    p.position_type,
    p.employment_type,
    p.grade_level,
    p.headcount_capacity,
    p.headcount_in_use,
    p.reports_to_position_code,
    p.status,
    p.effective_date,
    p.end_date,
    p.is_current,
    p.created_at,
    p.updated_at,
    p.job_family_group_name,
    p.job_family_name,
    p.job_role_name,
    p.job_level_name,
    p.organization_name
FROM positions p
%s
ORDER BY p.code
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query vacant positions: %w", err)
	}
	defer rows.Close()

	result := make([]model.Position, 0)
	for rows.Next() {
		pos, scanErr := scanPosition(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, *pos)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vacant positions: %w", err)
	}
	return result, nil
}

func (r *PostgreSQLRepository) GetPositionHeadcountStats(ctx context.Context, tenantID uuid.UUID, organizationCode string, includeSubordinates bool) (*model.HeadcountStats, error) {
	var (
		orgName  string
		codePath string
	)

	scopeQuery := `
SELECT name, COALESCE(code_path, '/' || code) AS code_path
FROM organization_units
WHERE tenant_id = $1 AND code = $2 AND is_current = true
ORDER BY effective_date DESC, created_at DESC
LIMIT 1`

	if err := r.db.QueryRowContext(ctx, scopeQuery, tenantID.String(), organizationCode).Scan(&orgName, &codePath); err != nil {
		if err == sql.ErrNoRows {
			return &model.HeadcountStats{
				OrganizationCodeField: organizationCode,
				OrganizationNameField: "",
				TotalCapacityField:    0,
				TotalFilledField:      0,
				TotalAvailableField:   0,
				LevelBreakdownField:   []model.LevelHeadcount{},
				TypeBreakdownField:    []model.TypeHeadcount{},
			}, nil
		}
		return nil, fmt.Errorf("lookup organization scope: %w", err)
	}

	args := []interface{}{tenantID.String(), organizationCode}

	condition := "ou.code = $2"
	if includeSubordinates {
		condition = "(ou.code = $2 OR ou.code_path LIKE $3)"
		args = append(args, codePath+"/%")
	}

	statsQuery := fmt.Sprintf(`
SELECT
    COALESCE(SUM(p.headcount_capacity), 0) AS capacity,
    COALESCE(SUM(p.headcount_in_use), 0) AS filled
FROM positions p
JOIN organization_units ou ON ou.tenant_id = p.tenant_id AND ou.code = p.organization_code AND ou.is_current = true
WHERE p.tenant_id = $1
  AND p.is_current = true
  AND p.status <> 'DELETED'
  AND %s
`, condition)

	var totalCapacity, totalFilled float64
	if err := r.db.QueryRowContext(ctx, statsQuery, args...).Scan(&totalCapacity, &totalFilled); err != nil {
		return nil, fmt.Errorf("query headcount totals: %w", err)
	}

	levelQuery := fmt.Sprintf(`
SELECT
    p.job_level_code,
    COALESCE(SUM(p.headcount_capacity), 0) AS capacity,
    COALESCE(SUM(p.headcount_in_use), 0) AS utilized,
    COALESCE(SUM(p.headcount_capacity - p.headcount_in_use), 0) AS available
FROM positions p
JOIN organization_units ou ON ou.tenant_id = p.tenant_id AND ou.code = p.organization_code AND ou.is_current = true
WHERE p.tenant_id = $1
  AND p.is_current = true
  AND p.status <> 'DELETED'
  AND %s
GROUP BY p.job_level_code
ORDER BY p.job_level_code
`, condition)

	levelRows, err := r.db.QueryContext(ctx, levelQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query level headcount: %w", err)
	}
	defer levelRows.Close()

	levelBreakdown := make([]model.LevelHeadcount, 0)
	for levelRows.Next() {
		var item model.LevelHeadcount
		if err := levelRows.Scan(&item.JobLevelCodeField, &item.CapacityField, &item.UtilizedField, &item.AvailableField); err != nil {
			return nil, fmt.Errorf("scan level headcount: %w", err)
		}
		levelBreakdown = append(levelBreakdown, item)
	}
	if err := levelRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate level headcount: %w", err)
	}

	typeQuery := fmt.Sprintf(`
SELECT
    p.position_type,
    COALESCE(SUM(p.headcount_capacity), 0) AS capacity,
    COALESCE(SUM(p.headcount_in_use), 0) AS filled,
    COALESCE(SUM(p.headcount_capacity - p.headcount_in_use), 0) AS available
FROM positions p
JOIN organization_units ou ON ou.tenant_id = p.tenant_id AND ou.code = p.organization_code AND ou.is_current = true
WHERE p.tenant_id = $1
  AND p.is_current = true
  AND p.status <> 'DELETED'
  AND %s
GROUP BY p.position_type
ORDER BY p.position_type
`, condition)

	typeRows, err := r.db.QueryContext(ctx, typeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query type headcount: %w", err)
	}
	defer typeRows.Close()

	typeBreakdown := make([]model.TypeHeadcount, 0)
	for typeRows.Next() {
		var item model.TypeHeadcount
		if err := typeRows.Scan(&item.PositionTypeField, &item.CapacityField, &item.FilledField, &item.AvailableField); err != nil {
			return nil, fmt.Errorf("scan type headcount: %w", err)
		}
		typeBreakdown = append(typeBreakdown, item)
	}
	if err := typeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate type headcount: %w", err)
	}

	familyQuery := fmt.Sprintf(`
SELECT
    p.job_family_code,
    MAX(p.job_family_name) AS job_family_name,
    COALESCE(SUM(p.headcount_capacity), 0) AS capacity,
    COALESCE(SUM(p.headcount_in_use), 0) AS utilized,
    COALESCE(SUM(p.headcount_capacity - p.headcount_in_use), 0) AS available
FROM positions p
JOIN organization_units ou ON ou.tenant_id = p.tenant_id AND ou.code = p.organization_code AND ou.is_current = true
WHERE p.tenant_id = $1
  AND p.is_current = true
  AND p.status <> 'DELETED'
  AND %s
GROUP BY p.job_family_code
ORDER BY p.job_family_code
`, condition)

	familyRows, err := r.db.QueryContext(ctx, familyQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query family headcount: %w", err)
	}
	defer familyRows.Close()

	familyBreakdown := make([]model.FamilyHeadcount, 0)
	for familyRows.Next() {
		var item model.FamilyHeadcount
		if err := familyRows.Scan(&item.JobFamilyCodeField, &item.JobFamilyNameField, &item.CapacityField, &item.UtilizedField, &item.AvailableField); err != nil {
			return nil, fmt.Errorf("scan family headcount: %w", err)
		}
		familyBreakdown = append(familyBreakdown, item)
	}
	if err := familyRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate family headcount: %w", err)
	}

	stats := &model.HeadcountStats{
		OrganizationCodeField: organizationCode,
		OrganizationNameField: orgName,
		TotalCapacityField:    totalCapacity,
		TotalFilledField:      totalFilled,
		TotalAvailableField:   totalCapacity - totalFilled,
		LevelBreakdownField:   levelBreakdown,
		TypeBreakdownField:    typeBreakdown,
		FamilyBreakdownField:  familyBreakdown,
	}

	return stats, nil
}

func (r *PostgreSQLRepository) GetPositionVersions(ctx context.Context, tenantID uuid.UUID, code string, includeDeleted bool) ([]model.Position, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(code)}
	whereParts := []string{"p.tenant_id = $1", "p.code = $2"}
	if !includeDeleted {
		whereParts = append(whereParts, "p.status <> 'DELETED'")
	}

	whereClause := "WHERE " + strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
SELECT
    p.record_id::text,
    p.tenant_id::text,
    p.code,
    p.title,
    p.job_profile_code,
    p.job_profile_name,
    p.job_family_group_code,
    p.job_family_code,
    p.job_role_code,
    p.job_level_code,
    p.organization_code,
    p.position_type,
    p.employment_type,
    p.grade_level,
    p.headcount_capacity,
    p.headcount_in_use,
    p.reports_to_position_code,
    p.status,
    p.effective_date,
    p.end_date,
    p.is_current,
    p.created_at,
    p.updated_at,
    p.job_family_group_name,
    p.job_family_name,
    p.job_role_name,
    p.job_level_name,
    p.organization_name
FROM positions p
%s
ORDER BY p.effective_date DESC, p.created_at DESC
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query position versions: %w", err)
	}
	defer rows.Close()

	versions := make([]model.Position, 0)
	for rows.Next() {
		pos, scanErr := scanPosition(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		versions = append(versions, *pos)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate position versions: %w", err)
	}

	return versions, nil
}

func (r *PostgreSQLRepository) GetVacantPositionConnection(ctx context.Context, tenantID uuid.UUID, filter *model.VacantPositionFilterInput, pagination *model.PaginationInput, sorting []model.VacantPositionSortInput) (*model.VacantPositionConnection, error) {
	args := []interface{}{tenantID.String()}
	argIndex := 2
	whereParts := []string{
		"p.tenant_id = $1",
		"p.is_current = true",
		"p.status <> 'DELETED'",
	}

	if filter != nil {
		if orgCodes := filter.OrganizationCodes; orgCodes != nil && len(*orgCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.organization_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*orgCodes)))
			argIndex++
		}
		if jobFamilyCodes := filter.JobFamilyCodes; jobFamilyCodes != nil && len(*jobFamilyCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_family_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobFamilyCodes)))
			argIndex++
		}
		if jobRoleCodes := filter.JobRoleCodes; jobRoleCodes != nil && len(*jobRoleCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_role_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobRoleCodes)))
			argIndex++
		}
		if jobLevelCodes := filter.JobLevelCodes; jobLevelCodes != nil && len(*jobLevelCodes) > 0 {
			whereParts = append(whereParts, fmt.Sprintf("p.job_level_code = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(toStringSlice(*jobLevelCodes)))
			argIndex++
		}
		if positionTypes := filter.PositionTypes; positionTypes != nil && len(*positionTypes) > 0 {
			values := make([]string, 0, len(*positionTypes))
			for _, v := range *positionTypes {
				values = append(values, strings.ToUpper(strings.TrimSpace(v)))
			}
			whereParts = append(whereParts, fmt.Sprintf("p.position_type = ANY($%d)", argIndex))
			args = append(args, pq.StringArray(values))
			argIndex++
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	asOfDate := time.Now().UTC().Format("2006-01-02")
	if filter != nil && filter.AsOfDate != nil && strings.TrimSpace(*filter.AsOfDate) != "" {
		asOfDate = strings.TrimSpace(*filter.AsOfDate)
	}
	args = append(args, asOfDate)
	asOfIdx := len(args)

	minVacantIdx := 0
	if filter != nil && filter.MinimumVacantDays != nil && *filter.MinimumVacantDays >= 0 {
		args = append(args, *filter.MinimumVacantDays)
		minVacantIdx = len(args)
	}

	page := int32(1)
	pageSize := int32(25)
	if pagination != nil {
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
			if pageSize > 200 {
				pageSize = 200
			}
		}
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	selectionConditions := []string{"comp.headcount_available > 0"}
	if minVacantIdx > 0 {
		selectionConditions = append(selectionConditions, fmt.Sprintf("comp.vacant_days >= $%d", minVacantIdx))
	}
	selectionWhere := strings.Join(selectionConditions, " AND ")

	baseCTE := fmt.Sprintf(`
WITH filtered_positions AS (
    SELECT
        p.code,
        p.organization_code,
        p.organization_name,
        p.job_family_code,
        p.job_role_code,
        p.job_level_code,
        p.headcount_capacity,
        p.effective_date
    FROM positions p
    %s
),
assignment_snapshot AS (
    SELECT
        pa.position_code,
        COALESCE(SUM(CASE WHEN pa.assignment_status <> 'ENDED'
            AND pa.effective_date <= $%d
            AND (pa.end_date IS NULL OR pa.end_date > $%d)
            THEN pa.fte ELSE 0 END), 0) AS active_fte,
        MAX(CASE WHEN pa.assignment_status = 'ENDED' AND pa.end_date <= $%d THEN pa.end_date END) AS last_vacated,
        COUNT(*) AS total_assignments
    FROM position_assignments pa
    JOIN filtered_positions fp ON fp.code = pa.position_code
    WHERE pa.tenant_id = $1
    GROUP BY pa.position_code
),
computed AS (
    SELECT
        fp.code AS position_code,
        fp.organization_code,
        fp.organization_name,
        fp.job_family_code,
        fp.job_role_code,
        fp.job_level_code,
        fp.headcount_capacity,
        COALESCE(asnap.active_fte, 0) AS active_fte,
        COALESCE(asnap.last_vacated, fp.effective_date) AS vacant_since,
        COALESCE(asnap.total_assignments, 0) AS total_assignments,
        GREATEST(fp.headcount_capacity - COALESCE(asnap.active_fte, 0), 0) AS headcount_available,
        DATE_PART('day', $%d::date - COALESCE(asnap.last_vacated, fp.effective_date))::int AS vacant_days
    FROM filtered_positions fp
    LEFT JOIN assignment_snapshot asnap ON asnap.position_code = fp.code
)
`, whereClause, asOfIdx, asOfIdx, asOfIdx, asOfIdx)

	countArgs := append([]interface{}{}, args...)
	countQuery := fmt.Sprintf(`%s SELECT COUNT(*) FROM computed comp WHERE %s`, baseCTE, selectionWhere)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count vacant positions: %w", err)
	}

	orderClause := "ORDER BY comp.vacant_since ASC, comp.position_code ASC"
	if len(sorting) > 0 {
		orderParts := make([]string, 0, len(sorting))
		for _, sort := range sorting {
			field := strings.ToUpper(strings.TrimSpace(sort.Field))
			column := ""
			switch field {
			case "VACANT_SINCE":
				column = "comp.vacant_since"
			case "HEADCOUNT_AVAILABLE":
				column = "comp.headcount_available"
			case "HEADCOUNT_CAPACITY":
				column = "comp.headcount_capacity"
			default:
				continue
			}
			direction := strings.ToUpper(strings.TrimSpace(sort.Direction))
			if direction != "ASC" {
				direction = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", column, direction))
		}
		if len(orderParts) > 0 {
			orderParts = append(orderParts, "comp.position_code ASC")
			orderClause = "ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	dataArgs := append([]interface{}{}, args...)
	limitIdx := len(dataArgs) + 1
	dataArgs = append(dataArgs, limit)
	offsetIdx := len(dataArgs) + 1
	dataArgs = append(dataArgs, offset)

	dataQuery := fmt.Sprintf(`
%s
SELECT
    comp.position_code,
    comp.organization_code,
    comp.organization_name,
    comp.job_family_code,
    comp.job_role_code,
    comp.job_level_code,
    comp.vacant_since,
    comp.headcount_capacity,
    comp.headcount_available,
    comp.total_assignments
FROM computed comp
WHERE %s
%s
LIMIT $%d OFFSET $%d
`, baseCTE, selectionWhere, orderClause, limitIdx, offsetIdx)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("query vacant positions connection: %w", err)
	}
	defer rows.Close()

	vacants := make([]model.VacantPosition, 0, limit)
	for rows.Next() {
		record, scanErr := scanVacantPosition(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		vacants = append(vacants, *record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vacant positions: %w", err)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + int(pageSize) - 1) / int(pageSize)
	}

	edges := make([]model.VacantPositionEdge, 0, len(vacants))
	for _, item := range vacants {
		edges = append(edges, model.VacantPositionEdge{
			CursorField: item.PositionCodeField,
			NodeField:   item,
		})
	}

	connection := &model.VacantPositionConnection{
		EdgesField: edges,
		DataField:  vacants,
		PaginationField: model.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TotalCountField: total,
	}

	return connection, nil
}

func (r *PostgreSQLRepository) GetPositionAssignments(ctx context.Context, tenantID uuid.UUID, positionCode string, filter *model.PositionAssignmentFilterInput, pagination *model.PaginationInput, sorting []model.PositionAssignmentSortInput) (*model.PositionAssignmentConnection, error) {
	page := int32(1)
	pageSize := int32(25)
	if pagination != nil {
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
			if pageSize > 200 {
				pageSize = 200
			}
		}
	}

	offset := int((page - 1) * pageSize)
	limit := int(pageSize)

	args := []interface{}{tenantID.String(), strings.TrimSpace(positionCode)}
	argIndex := 3
	whereParts := []string{"tenant_id = $1", "position_code = $2"}

	if filter != nil {
		if filter.EmployeeID != nil && strings.TrimSpace(*filter.EmployeeID) != "" {
			whereParts = append(whereParts, fmt.Sprintf("employee_id = $%d", argIndex))
			args = append(args, strings.TrimSpace(*filter.EmployeeID))
			argIndex++
		}
		if filter.Status != nil && strings.TrimSpace(*filter.Status) != "" {
			whereParts = append(whereParts, fmt.Sprintf("assignment_status = $%d", argIndex))
			args = append(args, strings.ToUpper(strings.TrimSpace(*filter.Status)))
			argIndex++
		}
		if filter.AssignmentTypes != nil && len(*filter.AssignmentTypes) > 0 {
			normalized := make([]string, 0, len(*filter.AssignmentTypes))
			for _, item := range *filter.AssignmentTypes {
				trimmed := strings.ToUpper(strings.TrimSpace(item))
				if trimmed == "" {
					continue
				}
				normalized = append(normalized, trimmed)
			}
			if len(normalized) > 0 {
				whereParts = append(whereParts, fmt.Sprintf("assignment_type = ANY($%d)", argIndex))
				args = append(args, pq.StringArray(normalized))
				argIndex++
			}
		}
		if filter.AsOfDate != nil && strings.TrimSpace(*filter.AsOfDate) != "" {
			whereParts = append(whereParts,
				fmt.Sprintf("(effective_date <= $%d AND (end_date IS NULL OR end_date >= $%d))", argIndex, argIndex))
			args = append(args, strings.TrimSpace(*filter.AsOfDate))
			argIndex++
		}
		if filter.DateRange != nil {
			if filter.DateRange.From != nil && strings.TrimSpace(*filter.DateRange.From) != "" {
				whereParts = append(whereParts, fmt.Sprintf("effective_date >= $%d", argIndex))
				args = append(args, strings.TrimSpace(*filter.DateRange.From))
				argIndex++
			}
			if filter.DateRange.To != nil && strings.TrimSpace(*filter.DateRange.To) != "" {
				whereParts = append(whereParts, fmt.Sprintf("effective_date <= $%d", argIndex))
				args = append(args, strings.TrimSpace(*filter.DateRange.To))
				argIndex++
			}
		}
		if filter.IncludeActingOnly {
			whereParts = append(whereParts, "assignment_type = 'ACTING'")
		}
		if !filter.IncludeHistorical {
			whereParts = append(whereParts, "assignment_status <> 'ENDED'")
		}
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	orderClause := "ORDER BY effective_date DESC, created_at DESC"
	if len(sorting) > 0 {
		orderParts := make([]string, 0, len(sorting))
		for _, sort := range sorting {
			field := strings.ToUpper(strings.TrimSpace(sort.Field))
			column := ""
			switch field {
			case "START_DATE", "EFFECTIVE_DATE":
				column = "effective_date"
			case "END_DATE":
				column = "end_date"
			case "CREATED_AT":
				column = "created_at"
			default:
				continue
			}
			direction := strings.ToUpper(strings.TrimSpace(sort.Direction))
			if direction != "ASC" {
				direction = "DESC"
			}
			orderParts = append(orderParts, fmt.Sprintf("%s %s", column, direction))
		}
		if len(orderParts) > 0 {
			orderClause = "ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM position_assignments %s`, whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count position assignments: %w", err)
	}

	selectQuery := fmt.Sprintf(`
SELECT
    assignment_id::text,
    tenant_id::text,
    position_code,
    position_record_id::text,
    employee_id::text,
    employee_name,
    employee_number,
    assignment_type,
    assignment_status,
    fte,
    effective_date,
    end_date,
    acting_until,
    auto_revert,
    reminder_sent_at,
    is_current,
    notes,
    created_at,
    updated_at
FROM position_assignments
%s
%s
LIMIT $%d OFFSET $%d`, whereClause, orderClause, argIndex, argIndex+1)

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query position assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]model.PositionAssignment, 0)
	for rows.Next() {
		assignment, scanErr := scanPositionAssignment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		assignments = append(assignments, *assignment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate position assignments: %w", err)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + int(pageSize) - 1) / int(pageSize)
	}

	edges := make([]model.PositionAssignmentEdge, 0, len(assignments))
	for _, assignment := range assignments {
		edges = append(edges, model.PositionAssignmentEdge{
			CursorField: assignment.AssignmentIDField,
			NodeField:   assignment,
		})
	}

	connection := &model.PositionAssignmentConnection{
		EdgesField: edges,
		DataField:  assignments,
		PaginationField: model.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TotalCountField: total,
	}

	return connection, nil
}

func (r *PostgreSQLRepository) GetPositionAssignmentAudit(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID *string, dateRange *model.DateRangeInput, pagination *model.PaginationInput) (*model.PositionAssignmentAuditConnection, error) {
	page := int32(1)
	pageSize := int32(25)
	if pagination != nil {
		if pagination.Page > 0 {
			page = pagination.Page
		}
		if pagination.PageSize > 0 {
			pageSize = pagination.PageSize
			if pageSize > 500 {
				pageSize = 500
			}
		}
	}

	limit := int(pageSize)
	offset := int((page - 1) * pageSize)

	args := []interface{}{tenantID.String(), strings.TrimSpace(positionCode)}
	conditions := []string{
		"al.tenant_id = $1",
		"p.code = $2",
		"al.resource_type = 'POSITION'",
		"al.response_data ? 'assignmentId'",
		"NULLIF(al.response_data->>'assignmentId', '') IS NOT NULL",
	}
	argIndex := 3

	if assignmentID != nil && strings.TrimSpace(*assignmentID) != "" {
		conditions = append(conditions, fmt.Sprintf("al.response_data->>'assignmentId' = $%d", argIndex))
		args = append(args, strings.TrimSpace(*assignmentID))
		argIndex++
	}

	if dateRange != nil {
		if dateRange.From != nil && strings.TrimSpace(*dateRange.From) != "" {
			conditions = append(conditions, fmt.Sprintf("al.timestamp >= $%d", argIndex))
			args = append(args, strings.TrimSpace(*dateRange.From))
			argIndex++
		}
		if dateRange.To != nil && strings.TrimSpace(*dateRange.To) != "" {
			conditions = append(conditions, fmt.Sprintf("al.timestamp <= $%d", argIndex))
			args = append(args, strings.TrimSpace(*dateRange.To))
			argIndex++
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
SELECT COUNT(*)
FROM audit_logs al
JOIN positions p ON p.tenant_id = al.tenant_id AND p.record_id = al.resource_id::uuid
WHERE %s
`, whereClause)

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count position assignment audit: %w", err)
	}

	if total == 0 {
		return &model.PositionAssignmentAuditConnection{
			DataField: []model.PositionAssignmentAudit{},
			PaginationField: model.PaginationInfo{
				TotalField:       0,
				PageField:        int(page),
				PageSizeField:    int(pageSize),
				HasNextField:     false,
				HasPreviousField: page > 1,
			},
			TotalCountField: 0,
		}, nil
	}

	selectQuery := fmt.Sprintf(`
SELECT
    al.response_data->>'assignmentId' AS assignment_id,
    COALESCE(NULLIF(al.action_name, ''), al.event_type) AS event_type,
    COALESCE(pa.effective_date, NULLIF(al.response_data->>'assignmentEffective', '')::date, al.timestamp::date) AS effective_date,
    COALESCE(pa.end_date, pa.acting_until, NULLIF(al.response_data->>'assignmentEndDate', '')::date) AS end_date,
    COALESCE(al.business_context->>'actor_name', al.actor_id) AS actor_name,
    COALESCE(al.changes, '[]'::jsonb)::text AS changes_json,
    al.timestamp
FROM audit_logs al
JOIN positions p ON p.tenant_id = al.tenant_id AND p.record_id = al.resource_id::uuid
LEFT JOIN position_assignments pa ON pa.tenant_id = al.tenant_id AND pa.assignment_id::text = al.response_data->>'assignmentId'
WHERE %s
ORDER BY al.timestamp DESC
LIMIT $%d OFFSET $%d
`, whereClause, argIndex, argIndex+1)

	queryArgs := append([]interface{}{}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query position assignment audit: %w", err)
	}
	defer rows.Close()

	records := make([]model.PositionAssignmentAudit, 0)
	for rows.Next() {
		var assignID sql.NullString
		var eventType sql.NullString
		var effectiveDate sql.NullTime
		var endDate sql.NullTime
		var actorName sql.NullString
		var changesJSON sql.NullString
		var timestamp time.Time

		if err := rows.Scan(&assignID, &eventType, &effectiveDate, &endDate, &actorName, &changesJSON, &timestamp); err != nil {
			return nil, fmt.Errorf("scan assignment audit row: %w", err)
		}

		if !assignID.Valid || strings.TrimSpace(assignID.String) == "" {
			continue
		}

		eventTypeValue := ""
		if eventType.Valid {
			eventTypeValue = strings.TrimSpace(eventType.String)
		}
		record := model.PositionAssignmentAudit{
			AssignmentIDField: assignID.String,
			EventTypeField:    strings.ToUpper(eventTypeValue),
			CreatedAtField:    timestamp,
			ActorField:        strings.TrimSpace(actorName.String),
		}

		if effectiveDate.Valid {
			record.EffectiveDateField = effectiveDate.Time
		} else {
			record.EffectiveDateField = timestamp
		}
		if endDate.Valid {
			record.EndDateField = &endDate.Time
		}

		if changesJSON.Valid && strings.TrimSpace(changesJSON.String) != "" && strings.TrimSpace(changesJSON.String) != "[]" {
			var parsed interface{}
			if err := json.Unmarshal([]byte(changesJSON.String), &parsed); err == nil {
				switch val := parsed.(type) {
				case map[string]interface{}:
					record.ChangesField = val
				case []interface{}:
					record.ChangesField = map[string]interface{}{"items": val}
				}
			}
		}

		if record.ActorField == "" {
			record.ActorField = "system"
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate position assignment audit: %w", err)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + int(pageSize) - 1) / int(pageSize)
	}

	connection := &model.PositionAssignmentAuditConnection{
		DataField: records,
		PaginationField: model.PaginationInfo{
			TotalField:       total,
			PageField:        int(page),
			PageSizeField:    int(pageSize),
			HasNextField:     int(page) < totalPages,
			HasPreviousField: page > 1,
		},
		TotalCountField: total,
	}

	return connection, nil
}

func scanPositionAssignment(scanner rowScanner) (*model.PositionAssignment, error) {
	var (
		assignmentID,
		tenantID,
		positionCode,
		positionRecordID,
		employeeID,
		employeeName,
		assignmentType,
		assignmentStatus string
		employeeNumber sql.NullString
		fte            float64
		effectiveDate  time.Time
		endDate        sql.NullTime
		actingUntil    sql.NullTime
		autoRevert     bool
		reminderSentAt sql.NullTime
		isCurrent      bool
		notes          sql.NullString
		createdAt      time.Time
		updatedAt      time.Time
	)

	if err := scanner.Scan(
		&assignmentID,
		&tenantID,
		&positionCode,
		&positionRecordID,
		&employeeID,
		&employeeName,
		&employeeNumber,
		&assignmentType,
		&assignmentStatus,
		&fte,
		&effectiveDate,
		&endDate,
		&actingUntil,
		&autoRevert,
		&reminderSentAt,
		&isCurrent,
		&notes,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan position assignment: %w", err)
	}

	assignment := &model.PositionAssignment{
		AssignmentIDField:     assignmentID,
		TenantIDField:         tenantID,
		PositionCodeField:     positionCode,
		PositionRecordIDField: positionRecordID,
		EmployeeIDField:       employeeID,
		EmployeeNameField:     employeeName,
		AssignmentTypeField:   assignmentType,
		AssignmentStatusField: assignmentStatus,
		FTEField:              fte,
		EffectiveDateField:    effectiveDate,
		AutoRevertField:       autoRevert,
		IsCurrentField:        isCurrent,
		CreatedAtField:        createdAt,
		UpdatedAtField:        updatedAt,
	}

	if employeeNumber.Valid {
		assignment.EmployeeNumberField = &employeeNumber.String
	}
	if endDate.Valid {
		assignment.EndDateField = &endDate.Time
	}
	if actingUntil.Valid {
		assignment.ActingUntilField = &actingUntil.Time
	}
	if reminderSentAt.Valid {
		assignment.ReminderSentAtField = &reminderSentAt.Time
	}
	if notes.Valid {
		trimmed := strings.TrimSpace(notes.String)
		if trimmed != "" {
			assignment.NotesField = &trimmed
		}
	}

	return assignment, nil
}

func scanVacantPosition(scanner rowScanner) (*model.VacantPosition, error) {
	var (
		positionCode       string
		organizationCode   string
		organizationName   sql.NullString
		jobFamilyCode      string
		jobRoleCode        string
		jobLevelCode       string
		vacantSince        time.Time
		headcountCapacity  float64
		headcountAvailable float64
		totalAssignments   int
	)

	if err := scanner.Scan(
		&positionCode,
		&organizationCode,
		&organizationName,
		&jobFamilyCode,
		&jobRoleCode,
		&jobLevelCode,
		&vacantSince,
		&headcountCapacity,
		&headcountAvailable,
		&totalAssignments,
	); err != nil {
		return nil, fmt.Errorf("scan vacant position: %w", err)
	}

	record := &model.VacantPosition{
		PositionCodeField:       positionCode,
		OrganizationCodeField:   organizationCode,
		JobFamilyCodeField:      jobFamilyCode,
		JobRoleCodeField:        jobRoleCode,
		JobLevelCodeField:       jobLevelCode,
		VacantSinceField:        vacantSince,
		HeadcountCapacityField:  headcountCapacity,
		HeadcountAvailableField: headcountAvailable,
		TotalAssignmentsField:   totalAssignments,
	}

	if organizationName.Valid {
		name := strings.TrimSpace(organizationName.String)
		if name != "" {
			record.OrganizationNameField = &name
		}
	}

	return record, nil
}

func scanPosition(scanner rowScanner) (*model.Position, error) {
	var (
		recordID,
		tenantID,
		code,
		title string
		jobProfileCode,
		jobProfileName,
		gradeLevel,
		reportsTo,
		jobFamilyGroupName,
		jobFamilyName,
		jobRoleName,
		jobLevelName,
		organizationName sql.NullString
		jobFamilyGroupCode,
		jobFamilyCode,
		jobRoleCode,
		jobLevelCode,
		organizationCode,
		positionType,
		employmentType,
		status string
		headcountCapacity,
		headcountInUse float64
		effectiveDate time.Time
		endDate       sql.NullTime
		isCurrent     bool
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := scanner.Scan(
		&recordID,
		&tenantID,
		&code,
		&title,
		&jobProfileCode,
		&jobProfileName,
		&jobFamilyGroupCode,
		&jobFamilyCode,
		&jobRoleCode,
		&jobLevelCode,
		&organizationCode,
		&positionType,
		&employmentType,
		&gradeLevel,
		&headcountCapacity,
		&headcountInUse,
		&reportsTo,
		&status,
		&effectiveDate,
		&endDate,
		&isCurrent,
		&createdAt,
		&updatedAt,
		&jobFamilyGroupName,
		&jobFamilyName,
		&jobRoleName,
		&jobLevelName,
		&organizationName,
	); err != nil {
		return nil, err
	}

	normalizedStatus := normalizePositionStatus(status, effectiveDate, endDate, isCurrent)

	position := &model.Position{
		CodeField:               code,
		RecordIDField:           recordID,
		TenantIDField:           tenantID,
		TitleField:              title,
		JobFamilyGroupCodeField: jobFamilyGroupCode,
		JobFamilyCodeField:      jobFamilyCode,
		JobRoleCodeField:        jobRoleCode,
		JobLevelCodeField:       jobLevelCode,
		OrganizationCodeField:   organizationCode,
		PositionTypeField:       positionType,
		EmploymentTypeField:     employmentType,
		HeadcountCapacityField:  headcountCapacity,
		HeadcountInUseField:     headcountInUse,
		StatusField:             normalizedStatus,
		EffectiveDateField:      effectiveDate,
		IsCurrentField:          isCurrent,
		CreatedAtField:          createdAt,
		UpdatedAtField:          updatedAt,
	}

	if jobProfileCode.Valid {
		position.JobProfileCodeField = &jobProfileCode.String
	}
	if jobProfileName.Valid {
		position.JobProfileNameField = &jobProfileName.String
	}
	if gradeLevel.Valid {
		position.GradeLevelField = &gradeLevel.String
	}
	if reportsTo.Valid {
		position.ReportsToPositionField = &reportsTo.String
	}
	if endDate.Valid {
		position.EndDateField = &endDate.Time
	}
	if jobFamilyGroupName.Valid {
		position.JobFamilyGroupNameField = &jobFamilyGroupName.String
	}
	if jobFamilyName.Valid {
		position.JobFamilyNameField = &jobFamilyName.String
	}
	if jobRoleName.Valid {
		position.JobRoleNameField = &jobRoleName.String
	}
	if jobLevelName.Valid {
		position.JobLevelNameField = &jobLevelName.String
	}
	if organizationName.Valid {
		position.OrganizationNameField = &organizationName.String
	}

	return position, nil
}

func normalizePositionStatus(status string, effectiveDate time.Time, endDate sql.NullTime, isCurrent bool) string {
	normalized := strings.ToUpper(strings.TrimSpace(status))
	if normalized != "PLANNED" {
		return normalized
	}

	today := cnTodayUTC()
	if effectiveDate.After(today) {
		return normalized
	}

	if isCurrent && (!endDate.Valid || endDate.Time.After(today)) {
		return "ACTIVE"
	}

	return "INACTIVE"
}

func cnTodayUTC() time.Time {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.Now().UTC().Truncate(24 * time.Hour)
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

func (r *PostgreSQLRepository) populatePositionAssignments(ctx context.Context, tenantID uuid.UUID, position *model.Position) error {
	if position == nil {
		return nil
	}

	assignments, err := r.fetchAssignmentsForPosition(ctx, tenantID, position.CodeField)
	if err != nil {
		return err
	}

	if len(assignments) == 0 {
		position.AssignmentHistoryField = []model.PositionAssignment{}
		position.CurrentAssignmentField = nil
		return nil
	}

	position.AssignmentHistoryField = assignments
	position.CurrentAssignmentField = nil

	for i := range position.AssignmentHistoryField {
		if position.AssignmentHistoryField[i].IsCurrent() {
			position.CurrentAssignmentField = &position.AssignmentHistoryField[i]
			break
		}
	}

	if position.CurrentAssignmentField == nil {
		position.CurrentAssignmentField = &position.AssignmentHistoryField[0]
	}

	return nil
}

func (r *PostgreSQLRepository) fetchAssignmentsForPosition(ctx context.Context, tenantID uuid.UUID, positionCode string) ([]model.PositionAssignment, error) {
	code := strings.TrimSpace(positionCode)
	if code == "" {
		return []model.PositionAssignment{}, nil
	}

	query := `
SELECT
    assignment_id::text,
    tenant_id::text,
    position_code,
    position_record_id::text,
    employee_id::text,
    employee_name,
    employee_number,
    assignment_type,
    assignment_status,
    fte,
    effective_date,
    end_date,
    acting_until,
    auto_revert,
    reminder_sent_at,
    is_current,
    notes,
    created_at,
    updated_at
FROM position_assignments
WHERE tenant_id = $1 AND position_code = $2
ORDER BY effective_date DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), code)
	if err != nil {
		return nil, fmt.Errorf("fetch position assignments: %w", err)
	}
	defer rows.Close()

	assignments := make([]model.PositionAssignment, 0)
	for rows.Next() {
		item, scanErr := scanPositionAssignment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		assignments = append(assignments, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate position assignments: %w", err)
	}

	return assignments, nil
}
