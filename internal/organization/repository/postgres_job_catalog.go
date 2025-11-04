package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cube-castle/internal/organization/dto"
	"github.com/google/uuid"
)

func (r *PostgreSQLRepository) GetJobFamilyGroups(ctx context.Context, tenantID uuid.UUID, includeInactive bool, asOfDate *string) ([]dto.JobFamilyGroup, error) {
	args := []interface{}{tenantID.String()}
	whereParts := []string{"tenant_id = $1"}
	argIndex := 2

	if asOfDate != nil && strings.TrimSpace(*asOfDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("effective_date <= $%d", argIndex))
		whereParts = append(whereParts, fmt.Sprintf("(end_date IS NULL OR end_date > $%d)", argIndex))
		args = append(args, strings.TrimSpace(*asOfDate))
		argIndex++
	} else {
		whereParts = append(whereParts, "is_current = true")
	}

	if !includeInactive {
		whereParts = append(whereParts, "status = 'ACTIVE'")
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	query := fmt.Sprintf(`
SELECT DISTINCT ON (family_group_code)
    record_id::text,
    tenant_id::text,
    family_group_code,
    name,
    description,
    status,
    effective_date,
    end_date,
    is_current
FROM job_family_groups
%s
ORDER BY family_group_code, effective_date DESC, created_at DESC
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query job family groups: %w", err)
	}
	defer rows.Close()

	var result []dto.JobFamilyGroup
	for rows.Next() {
		var (
			item      dto.JobFamilyGroup
			desc      sql.NullString
			endDate   sql.NullTime
			isCurrent bool
		)
		if err := rows.Scan(
			&item.RecordIDField,
			&item.TenantIDField,
			&item.CodeField,
			&item.NameField,
			&desc,
			&item.StatusField,
			&item.EffectiveDateField,
			&endDate,
			&isCurrent,
		); err != nil {
			return nil, fmt.Errorf("scan job family group: %w", err)
		}
		if desc.Valid {
			item.DescriptionField = &desc.String
		}
		if endDate.Valid {
			item.EndDateField = &endDate.Time
		}
		item.IsCurrentField = isCurrent
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job family groups: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLRepository) GetJobFamilies(ctx context.Context, tenantID uuid.UUID, groupCode string, includeInactive bool, asOfDate *string) ([]dto.JobFamily, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(groupCode)}
	whereParts := []string{"tenant_id = $1", "family_group_code = $2"}
	argIndex := 3

	if asOfDate != nil && strings.TrimSpace(*asOfDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("effective_date <= $%d", argIndex))
		whereParts = append(whereParts, fmt.Sprintf("(end_date IS NULL OR end_date > $%d)", argIndex))
		args = append(args, strings.TrimSpace(*asOfDate))
		argIndex++
	} else {
		whereParts = append(whereParts, "is_current = true")
	}

	if !includeInactive {
		whereParts = append(whereParts, "status = 'ACTIVE'")
	}

	whereClause := "WHERE " + strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
SELECT DISTINCT ON (family_code)
    record_id::text,
    tenant_id::text,
    family_code,
    name,
    description,
    status,
    effective_date,
    end_date,
    is_current,
    family_group_code
FROM job_families
%s
ORDER BY family_code, effective_date DESC, created_at DESC
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query job families: %w", err)
	}
	defer rows.Close()

	var result []dto.JobFamily
	for rows.Next() {
		var (
			item      dto.JobFamily
			desc      sql.NullString
			endDate   sql.NullTime
			isCurrent bool
		)
		if err := rows.Scan(
			&item.RecordIDField,
			&item.TenantIDField,
			&item.CodeField,
			&item.NameField,
			&desc,
			&item.StatusField,
			&item.EffectiveDateField,
			&endDate,
			&isCurrent,
			&item.FamilyGroupCodeField,
		); err != nil {
			return nil, fmt.Errorf("scan job family: %w", err)
		}
		if desc.Valid {
			item.DescriptionField = &desc.String
		}
		if endDate.Valid {
			item.EndDateField = &endDate.Time
		}
		item.IsCurrentField = isCurrent
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job families: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLRepository) GetJobRoles(ctx context.Context, tenantID uuid.UUID, familyCode string, includeInactive bool, asOfDate *string) ([]dto.JobRole, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(familyCode)}
	whereParts := []string{"tenant_id = $1", "family_code = $2"}
	argIndex := 3

	if asOfDate != nil && strings.TrimSpace(*asOfDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("effective_date <= $%d", argIndex))
		whereParts = append(whereParts, fmt.Sprintf("(end_date IS NULL OR end_date > $%d)", argIndex))
		args = append(args, strings.TrimSpace(*asOfDate))
		argIndex++
	} else {
		whereParts = append(whereParts, "is_current = true")
	}

	if !includeInactive {
		whereParts = append(whereParts, "status = 'ACTIVE'")
	}

	whereClause := "WHERE " + strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
SELECT DISTINCT ON (role_code)
    record_id::text,
    tenant_id::text,
    role_code,
    name,
    description,
    status,
    effective_date,
    end_date,
    is_current,
    family_code
FROM job_roles
%s
ORDER BY role_code, effective_date DESC, created_at DESC
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query job roles: %w", err)
	}
	defer rows.Close()

	var result []dto.JobRole
	for rows.Next() {
		var (
			item      dto.JobRole
			desc      sql.NullString
			endDate   sql.NullTime
			isCurrent bool
		)
		if err := rows.Scan(
			&item.RecordIDField,
			&item.TenantIDField,
			&item.CodeField,
			&item.NameField,
			&desc,
			&item.StatusField,
			&item.EffectiveDateField,
			&endDate,
			&isCurrent,
			&item.FamilyCodeField,
		); err != nil {
			return nil, fmt.Errorf("scan job role: %w", err)
		}
		if desc.Valid {
			item.DescriptionField = &desc.String
		}
		if endDate.Valid {
			item.EndDateField = &endDate.Time
		}
		item.IsCurrentField = isCurrent
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job roles: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLRepository) GetJobLevels(ctx context.Context, tenantID uuid.UUID, roleCode string, includeInactive bool, asOfDate *string) ([]dto.JobLevel, error) {
	args := []interface{}{tenantID.String(), strings.TrimSpace(roleCode)}
	whereParts := []string{"tenant_id = $1", "role_code = $2"}
	argIndex := 3

	if asOfDate != nil && strings.TrimSpace(*asOfDate) != "" {
		whereParts = append(whereParts, fmt.Sprintf("effective_date <= $%d", argIndex))
		whereParts = append(whereParts, fmt.Sprintf("(end_date IS NULL OR end_date > $%d)", argIndex))
		args = append(args, strings.TrimSpace(*asOfDate))
		argIndex++
	} else {
		whereParts = append(whereParts, "is_current = true")
	}

	if !includeInactive {
		whereParts = append(whereParts, "status = 'ACTIVE'")
	}

	whereClause := "WHERE " + strings.Join(whereParts, " AND ")

	query := fmt.Sprintf(`
SELECT DISTINCT ON (level_code)
    record_id::text,
    tenant_id::text,
    level_code,
    name,
    description,
    status,
    effective_date,
    end_date,
    is_current,
    role_code,
    level_rank
FROM job_levels
%s
ORDER BY level_code, effective_date DESC, created_at DESC
`, whereClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query job levels: %w", err)
	}
	defer rows.Close()

	var result []dto.JobLevel
	for rows.Next() {
		var (
			item      dto.JobLevel
			desc      sql.NullString
			endDate   sql.NullTime
			isCurrent bool
		)
		if err := rows.Scan(
			&item.RecordIDField,
			&item.TenantIDField,
			&item.CodeField,
			&item.NameField,
			&desc,
			&item.StatusField,
			&item.EffectiveDateField,
			&endDate,
			&isCurrent,
			&item.RoleCodeField,
			&item.LevelRankField,
		); err != nil {
			return nil, fmt.Errorf("scan job level: %w", err)
		}
		if desc.Valid {
			item.DescriptionField = &desc.String
		}
		if endDate.Valid {
			item.EndDateField = &endDate.Time
		}
		item.IsCurrentField = isCurrent
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate job levels: %w", err)
	}

	return result, nil
}
