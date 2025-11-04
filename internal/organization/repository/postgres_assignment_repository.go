package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/dto"
	"github.com/google/uuid"
)

// GetAssignmentHistory 返回指定职位的完整任职历史记录。
func (r *PostgreSQLRepository) GetAssignmentHistory(
	ctx context.Context,
	tenantID uuid.UUID,
	positionCode string,
	filter *dto.PositionAssignmentFilterInput,
	pagination *dto.PaginationInput,
	sorting []dto.PositionAssignmentSortInput,
) (*dto.PositionAssignmentConnection, error) {
	trimmedCode := strings.TrimSpace(positionCode)
	if trimmedCode == "" {
		return nil, fmt.Errorf("positionCode is required")
	}

	localFilter := dto.PositionAssignmentFilterInput{}
	if filter != nil {
		localFilter = *filter
	}
	localFilter.IncludeHistorical = true

	return r.GetPositionAssignments(ctx, tenantID, trimmedCode, &localFilter, pagination, sorting)
}

// GetAssignmentStats 返回任职统计信息，可按职位或组织筛选。
func (r *PostgreSQLRepository) GetAssignmentStats(
	ctx context.Context,
	tenantID uuid.UUID,
	positionCode string,
	organizationCode string,
) (*dto.AssignmentStats, error) {
	args := []interface{}{tenantID.String()}
	conditions := []string{"pa.tenant_id = $1"}
	argIndex := 2

	trimmedPosition := strings.TrimSpace(positionCode)
	if trimmedPosition != "" {
		conditions = append(conditions, fmt.Sprintf("pa.position_code = $%d", argIndex))
		args = append(args, trimmedPosition)
		argIndex++
	}

	trimmedOrg := strings.TrimSpace(organizationCode)
	joinClause := ""
	if trimmedOrg != "" {
		joinClause = `
JOIN positions p ON p.tenant_id = pa.tenant_id AND p.code = pa.position_code
`
		conditions = append(conditions, fmt.Sprintf("p.organization_code = $%d", argIndex))
		args = append(args, trimmedOrg)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
SELECT
    COUNT(*) AS total_count,
    COUNT(*) FILTER (WHERE pa.assignment_status = 'ACTIVE') AS active_count,
    COUNT(*) FILTER (WHERE pa.assignment_status = 'PENDING') AS pending_count,
    COUNT(*) FILTER (WHERE pa.assignment_status = 'ENDED') AS ended_count,
    COUNT(*) FILTER (WHERE pa.assignment_type = 'PRIMARY') AS primary_count,
    COUNT(*) FILTER (WHERE pa.assignment_type = 'SECONDARY') AS secondary_count,
    COUNT(*) FILTER (WHERE pa.assignment_type = 'ACTING') AS acting_count,
    COALESCE(MAX(pa.updated_at), MAX(pa.created_at)) AS last_updated
FROM position_assignments pa
%s
WHERE %s`, joinClause, whereClause)

	var (
		totalCount     sql.NullInt64
		activeCount    sql.NullInt64
		pendingCount   sql.NullInt64
		endedCount     sql.NullInt64
		primaryCount   sql.NullInt64
		secondaryCount sql.NullInt64
		actingCount    sql.NullInt64
		lastUpdated    sql.NullTime
	)

	if err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&totalCount,
		&activeCount,
		&pendingCount,
		&endedCount,
		&primaryCount,
		&secondaryCount,
		&actingCount,
		&lastUpdated,
	); err != nil {
		return nil, fmt.Errorf("assignment stats query failed: %w", err)
	}

	statsTime := time.Now().UTC()
	if lastUpdated.Valid {
		statsTime = lastUpdated.Time
	}

	stats := &dto.AssignmentStats{
		TotalCountField:     int(totalCount.Int64),
		ActiveCountField:    int(activeCount.Int64),
		PendingCountField:   int(pendingCount.Int64),
		EndedCountField:     int(endedCount.Int64),
		PrimaryCountField:   int(primaryCount.Int64),
		SecondaryCountField: int(secondaryCount.Int64),
		ActingCountField:    int(actingCount.Int64),
		LastUpdatedAtField:  statsTime,
	}

	if trimmedPosition != "" {
		stats.PositionCodeField = &trimmedPosition
	}
	if trimmedOrg != "" {
		stats.OrganizationCodeField = &trimmedOrg
	}

	return stats, nil
}
