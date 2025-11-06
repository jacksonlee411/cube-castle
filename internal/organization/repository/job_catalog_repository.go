package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type JobCatalogRepository struct {
	db     *sql.DB
	logger pkglogger.Logger
}

func NewJobCatalogRepository(db *sql.DB, baseLogger pkglogger.Logger) *JobCatalogRepository {
	return &JobCatalogRepository{
		db:     db,
		logger: scopedLogger(baseLogger, "jobCatalog", "JobCatalogRepository", nil),
	}
}

func (r *JobCatalogRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

type temporalRow struct {
	RecordID      uuid.UUID
	EffectiveDate time.Time
	EndDate       sql.NullTime
	IsCurrent     bool
}

// JobCatalogTimelineEntry 暴露 Job Catalog 各层级的时态版本信息，供验证器使用。
type JobCatalogTimelineEntry struct {
	RecordID      uuid.UUID
	EffectiveDate time.Time
	EndDate       *time.Time
	IsCurrent     bool
	Status        string
}

func normalizeOptionalString(value *string) interface{} {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func marshalOptionalJSON(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(payload), nil
}

func (r *JobCatalogRepository) queryRows(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return r.db.QueryContext(ctx, query, args...)
}

func (r *JobCatalogRepository) queryRow(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	if tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return r.db.QueryRowContext(ctx, query, args...)
}

func (r *JobCatalogRepository) exec(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return r.db.ExecContext(ctx, query, args...)
}

func (r *JobCatalogRepository) listTimeline(ctx context.Context, tx *sql.Tx, query string, args ...interface{}) ([]JobCatalogTimelineEntry, error) {
	rows, err := r.queryRows(ctx, tx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	timeline := make([]JobCatalogTimelineEntry, 0)
	for rows.Next() {
		var (
			entry JobCatalogTimelineEntry
			end   sql.NullTime
		)
		if err := rows.Scan(&entry.RecordID, &entry.EffectiveDate, &end, &entry.IsCurrent, &entry.Status); err != nil {
			return nil, err
		}
		if end.Valid {
			endTime := end.Time
			entry.EndDate = &endTime
		}
		timeline = append(timeline, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return timeline, nil
}

// ListFamilyGroupTimeline 返回指定职类的所有版本，按生效日期升序排列。
func (r *JobCatalogRepository) ListFamilyGroupTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]JobCatalogTimelineEntry, error) {
	query := `SELECT record_id, effective_date, end_date, is_current, status
FROM job_family_groups
WHERE tenant_id = $1 AND family_group_code = $2
ORDER BY effective_date ASC`
	return r.listTimeline(ctx, nil, query, tenantID, strings.ToUpper(strings.TrimSpace(code)))
}

// ListJobFamilyTimeline 返回指定职种的所有版本，按生效日期升序排列。
func (r *JobCatalogRepository) ListJobFamilyTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]JobCatalogTimelineEntry, error) {
	query := `SELECT record_id, effective_date, end_date, is_current, status
FROM job_families
WHERE tenant_id = $1 AND family_code = $2
ORDER BY effective_date ASC`
	return r.listTimeline(ctx, nil, query, tenantID, strings.ToUpper(strings.TrimSpace(code)))
}

// ListJobRoleTimeline 返回指定职务的所有版本，按生效日期升序排列。
func (r *JobCatalogRepository) ListJobRoleTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]JobCatalogTimelineEntry, error) {
	query := `SELECT record_id, effective_date, end_date, is_current, status
FROM job_roles
WHERE tenant_id = $1 AND role_code = $2
ORDER BY effective_date ASC`
	return r.listTimeline(ctx, nil, query, tenantID, strings.ToUpper(strings.TrimSpace(code)))
}

// ListJobLevelTimeline 返回指定职级的所有版本，按生效日期升序排列。
func (r *JobCatalogRepository) ListJobLevelTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]JobCatalogTimelineEntry, error) {
	query := `SELECT record_id, effective_date, end_date, is_current, status
FROM job_levels
WHERE tenant_id = $1 AND level_code = $2
ORDER BY effective_date ASC`
	return r.listTimeline(ctx, nil, query, tenantID, strings.ToUpper(strings.TrimSpace(code)))
}

func normalizeTemporal(rows []temporalRow) []temporalRow {
	if len(rows) == 0 {
		return rows
	}
	normalized := make([]temporalRow, len(rows))
	copy(normalized, rows)

	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i := range normalized {
		// end_date 为下一条记录的前一天
		if i < len(normalized)-1 {
			nextStart := normalized[i+1].EffectiveDate
			end := nextStart.AddDate(0, 0, -1)
			normalized[i].EndDate = sql.NullTime{Time: end, Valid: true}
		} else {
			normalized[i].EndDate = sql.NullTime{Valid: false}
		}

		// is_current 根据有效期与当前日期判定
		isCurrent := !normalized[i].EffectiveDate.After(today)
		if i < len(normalized)-1 && !normalized[i+1].EffectiveDate.After(today) {
			isCurrent = false
		}
		normalized[i].IsCurrent = isCurrent
	}

	return normalized
}

func (r *JobCatalogRepository) applyTemporalUpdates(ctx context.Context, tx *sql.Tx, table string, updates []temporalRow) error {
	if len(updates) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE %s SET end_date = $2, is_current = $3, updated_at = NOW() WHERE record_id = $1`, table)
	for _, row := range updates {
		var endDate interface{}
		if row.EndDate.Valid {
			endDate = row.EndDate.Time
		} else {
			endDate = nil
		}
		if _, err := r.exec(ctx, tx, query, row.RecordID, endDate, row.IsCurrent); err != nil {
			return fmt.Errorf("failed to update temporal row in %s: %w", table, err)
		}
	}
	return nil
}

// Job Family Group operations

func (r *JobCatalogRepository) GetCurrentFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamilyGroup, error) {
	query := `SELECT record_id, tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current
FROM job_family_groups WHERE tenant_id = $1 AND family_group_code = $2 AND is_current = true LIMIT 1`

	var entry types.JobFamilyGroup
	err := r.queryRow(ctx, tx, query, tenantID, code).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query job family group: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) InsertFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, req *types.CreateJobFamilyGroupRequest) (*types.JobFamilyGroup, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	query := `INSERT INTO job_family_groups (
tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,NULL,$7,NOW(),NOW())
RETURNING record_id, tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamilyGroup
	err = r.queryRow(ctx, tx, query,
		tenantID,
		req.Code,
		req.Name,
		req.Description,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job family group already exists for effective date")
			}
		}
		return nil, fmt.Errorf("failed to insert job family group: %w", err)
	}

	if err := r.recalculateFamilyGroupTimeline(ctx, tx, tenantID, req.Code); err != nil {
		return nil, err
	}

	r.logger.Infof("Job family group inserted: %s (%s)", req.Code, effectiveDate.Format("2006-01-02"))
	return &entry, nil
}

func (r *JobCatalogRepository) UpdateFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID uuid.UUID, req *types.UpdateJobFamilyGroupRequest) (*types.JobFamilyGroup, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	query := `UPDATE job_family_groups
SET name = $1,
    description = $2,
    status = $3,
    effective_date = $4,
    updated_at = NOW()
WHERE tenant_id = $5 AND record_id = $6
RETURNING record_id, tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamilyGroup
	err = r.queryRow(ctx, tx, query,
		req.Name,
		normalizeOptionalString(req.Description),
		req.Status,
		effectiveDate,
		tenantID,
		recordID,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job family group already exists for effective date: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to update job family group: %w", err)
	}

	if err := r.recalculateFamilyGroupTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return r.GetFamilyGroupByRecordID(ctx, tx, tenantID, recordID)
}

func (r *JobCatalogRepository) recalculateFamilyGroupTimeline(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) error {
	query := `SELECT record_id, effective_date, end_date, is_current FROM job_family_groups WHERE tenant_id = $1 AND family_group_code = $2 ORDER BY effective_date FOR UPDATE`
	rows, err := r.queryRows(ctx, tx, query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load family group timeline: %w", err)
	}
	defer rows.Close()

	var timeline []temporalRow
	for rows.Next() {
		var row temporalRow
		if err := rows.Scan(&row.RecordID, &row.EffectiveDate, &row.EndDate, &row.IsCurrent); err != nil {
			return fmt.Errorf("failed to scan family group timeline: %w", err)
		}
		timeline = append(timeline, row)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("family group timeline iteration error: %w", err)
	}

	normalized := normalizeTemporal(timeline)
	return r.applyTemporalUpdates(ctx, tx, "job_family_groups", normalized)
}

// InsertFamilyGroupVersion 插入新的职类版本
func (r *JobCatalogRepository) InsertFamilyGroupVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest) (*types.JobFamilyGroup, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	query := `INSERT INTO job_family_groups (
tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,NULL,$7,NOW(),NOW())
RETURNING record_id, tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamilyGroup
	err = r.queryRow(ctx, tx, query,
		tenantID,
		code,
		req.Name,
		req.Description,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job family group version already exists for effective date")
			}
		}
		return nil, fmt.Errorf("failed to insert job family group version: %w", err)
	}

	if err := r.recalculateFamilyGroupTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return &entry, nil
}

// Job family operations

func (r *JobCatalogRepository) GetCurrentJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamily, error) {
	query := `SELECT record_id, tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current
FROM job_families WHERE tenant_id = $1 AND family_code = $2 AND is_current = true LIMIT 1`

	var entry types.JobFamily
	err := r.queryRow(ctx, tx, query, tenantID, code).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyGroupCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query job family: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) InsertJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, parentRecord uuid.UUID, req *types.CreateJobFamilyRequest) (*types.JobFamily, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	query := `INSERT INTO job_families (
tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULL,$9,NOW(),NOW())
RETURNING record_id, tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamily
	err = r.queryRow(ctx, tx, query,
		tenantID,
		req.Code,
		req.JobFamilyGroupCode,
		parentRecord,
		req.Name,
		req.Description,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyGroupCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job family already exists for effective date")
			} else if pqErr.Code == "23503" {
				return nil, fmt.Errorf("parent job family group not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job family: %w", err)
	}

	if err := r.recalculateJobFamilyTimeline(ctx, tx, tenantID, req.Code); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) UpdateJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID uuid.UUID, groupCode string, parentRecord uuid.UUID, req *types.UpdateJobFamilyRequest) (*types.JobFamily, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	query := `UPDATE job_families
SET family_group_code = $1,
    parent_record_id = $2,
    name = $3,
    description = $4,
    status = $5,
    effective_date = $6,
    updated_at = NOW()
WHERE tenant_id = $7 AND record_id = $8
RETURNING record_id, tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamily
	err = r.queryRow(ctx, tx, query,
		groupCode,
		parentRecord,
		req.Name,
		normalizeOptionalString(req.Description),
		req.Status,
		effectiveDate,
		tenantID,
		recordID,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyGroupCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("job family already exists for effective date: %w", err)
			case "23503":
				return nil, fmt.Errorf("parent job family group not found: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to update job family: %w", err)
	}

	if err := r.recalculateJobFamilyTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return r.GetJobFamilyByRecordID(ctx, tx, tenantID, recordID)
}

func (r *JobCatalogRepository) InsertJobFamilyVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, parentRecord uuid.UUID, req *types.JobCatalogVersionRequest) (*types.JobFamily, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	query := `WITH latest AS (
	SELECT record_id, family_group_code, parent_record_id
	FROM job_families
	WHERE tenant_id = $1 AND family_code = $2
	ORDER BY effective_date DESC
	LIMIT 1
)
INSERT INTO job_families (
	tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current, created_at, updated_at
)
SELECT
	$1,
	$2,
	latest.family_group_code,
	latest.parent_record_id,
	$4,
	$5,
	$6,
	$7,
	NULL,
	$8,
	NOW(),
	NOW()
FROM latest
WHERE latest.record_id = $3
RETURNING record_id, tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current`

	var entry types.JobFamily
	err = r.queryRow(ctx, tx, query,
		tenantID,
		code,
		parentRecord,
		req.Name,
		req.Description,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyGroupCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("job family parent record mismatch")
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("job family version already exists for effective date")
			case "23503":
				return nil, fmt.Errorf("parent job family group not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job family version: %w", err)
	}

	if err := r.recalculateJobFamilyTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) recalculateJobFamilyTimeline(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) error {
	query := `SELECT record_id, effective_date, end_date, is_current FROM job_families WHERE tenant_id = $1 AND family_code = $2 ORDER BY effective_date FOR UPDATE`
	rows, err := r.queryRows(ctx, tx, query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load job family timeline: %w", err)
	}
	defer rows.Close()

	var timeline []temporalRow
	for rows.Next() {
		var row temporalRow
		if err := rows.Scan(&row.RecordID, &row.EffectiveDate, &row.EndDate, &row.IsCurrent); err != nil {
			return fmt.Errorf("failed to scan job family timeline: %w", err)
		}
		timeline = append(timeline, row)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("job family timeline iteration error: %w", err)
	}

	normalized := normalizeTemporal(timeline)
	return r.applyTemporalUpdates(ctx, tx, "job_families", normalized)
}

// Job role operations

func (r *JobCatalogRepository) GetCurrentJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobRole, error) {
	query := `SELECT record_id, tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current
FROM job_roles WHERE tenant_id = $1 AND role_code = $2 AND is_current = true LIMIT 1`

	var entry types.JobRole
	err := r.queryRow(ctx, tx, query, tenantID, code).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Competency,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query job role: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) InsertJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, parentRecord uuid.UUID, req *types.CreateJobRoleRequest) (*types.JobRole, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	competency, err := marshalOptionalJSON(req.CompetencyModel)
	if err != nil {
		return nil, fmt.Errorf("invalid competency model payload: %w", err)
	}

	query := `INSERT INTO job_roles (
tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NULL,$10,NOW(),NOW())
RETURNING record_id, tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current`

	var entry types.JobRole
	err = r.queryRow(ctx, tx, query,
		tenantID,
		req.Code,
		req.JobFamilyCode,
		parentRecord,
		req.Name,
		normalizeOptionalString(req.Description),
		competency,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Competency,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job role already exists for effective date")
			} else if pqErr.Code == "23503" {
				return nil, fmt.Errorf("parent job family not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job role: %w", err)
	}

	if err := r.recalculateJobRoleTimeline(ctx, tx, tenantID, req.Code); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) UpdateJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID uuid.UUID, familyCode string, parentRecord uuid.UUID, req *types.UpdateJobRoleRequest) (*types.JobRole, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	query := `UPDATE job_roles
SET family_code = $1,
    parent_record_id = $2,
    name = $3,
    description = $4,
    status = $5,
    effective_date = $6,
    updated_at = NOW()
WHERE tenant_id = $7 AND record_id = $8
RETURNING record_id, tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current`

	var entry types.JobRole
	err = r.queryRow(ctx, tx, query,
		familyCode,
		parentRecord,
		req.Name,
		normalizeOptionalString(req.Description),
		req.Status,
		effectiveDate,
		tenantID,
		recordID,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Competency,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("job role already exists for effective date: %w", err)
			case "23503":
				return nil, fmt.Errorf("parent job family not found: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to update job role: %w", err)
	}

	if err := r.recalculateJobRoleTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return r.GetJobRoleByRecordID(ctx, tx, tenantID, recordID)
}

func (r *JobCatalogRepository) InsertJobRoleVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, parentRecord uuid.UUID, req *types.JobCatalogVersionRequest) (*types.JobRole, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	familyCodeQuery := `SELECT family_code FROM job_roles WHERE tenant_id = $1 AND role_code = $2 ORDER BY effective_date DESC LIMIT 1`
	var familyCode string
	if err := r.queryRow(ctx, tx, familyCodeQuery, tenantID, code).Scan(&familyCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job role not found for code %s", code)
		}
		return nil, fmt.Errorf("failed to resolve job role family code: %w", err)
	}

	query := `INSERT INTO job_roles (
tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,NULL,$10,NOW(),NOW())
RETURNING record_id, tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current`

	var entry types.JobRole
	err = r.queryRow(ctx, tx, query,
		tenantID,
		code,
		familyCode,
		parentRecord,
		req.Name,
		normalizeOptionalString(req.Description),
		nil,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Competency,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job role version already exists for effective date")
			} else if pqErr.Code == "23503" {
				return nil, fmt.Errorf("parent job family not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job role version: %w", err)
	}

	if err := r.recalculateJobRoleTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) recalculateJobRoleTimeline(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) error {
	query := `SELECT record_id, effective_date, end_date, is_current FROM job_roles WHERE tenant_id = $1 AND role_code = $2 ORDER BY effective_date FOR UPDATE`
	rows, err := r.queryRows(ctx, tx, query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load job role timeline: %w", err)
	}
	defer rows.Close()

	var timeline []temporalRow
	for rows.Next() {
		var row temporalRow
		if err := rows.Scan(&row.RecordID, &row.EffectiveDate, &row.EndDate, &row.IsCurrent); err != nil {
			return fmt.Errorf("failed to scan job role timeline: %w", err)
		}
		timeline = append(timeline, row)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("job role timeline iteration error: %w", err)
	}

	normalized := normalizeTemporal(timeline)
	return r.applyTemporalUpdates(ctx, tx, "job_roles", normalized)
}

// Job level operations

func (r *JobCatalogRepository) GetCurrentJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobLevel, error) {
	query := `SELECT record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current
FROM job_levels WHERE tenant_id = $1 AND level_code = $2 AND is_current = true LIMIT 1`

	var entry types.JobLevel
	err := r.queryRow(ctx, tx, query, tenantID, code).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.RoleCode,
		&entry.ParentRecord,
		&entry.LevelRank,
		&entry.Name,
		&entry.Description,
		&entry.SalaryBand,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query job level: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) InsertJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, parentRecord uuid.UUID, req *types.CreateJobLevelRequest) (*types.JobLevel, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	isCurrent := !effectiveDate.After(today)

	salaryBand, err := marshalOptionalJSON(req.SalaryBand)
	if err != nil {
		return nil, fmt.Errorf("invalid salary band payload: %w", err)
	}

	query := `INSERT INTO job_levels (
tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,$11,NOW(),NOW())
RETURNING record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current`

	var entry types.JobLevel
	err = r.queryRow(ctx, tx, query,
		tenantID,
		req.Code,
		req.JobRoleCode,
		parentRecord,
		req.LevelRank,
		req.Name,
		normalizeOptionalString(req.Description),
		salaryBand,
		req.Status,
		effectiveDate,
		isCurrent,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.RoleCode,
		&entry.ParentRecord,
		&entry.LevelRank,
		&entry.Name,
		&entry.Description,
		&entry.SalaryBand,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job level already exists for effective date")
			} else if pqErr.Code == "23503" {
				return nil, fmt.Errorf("parent job role not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job level: %w", err)
	}

	if err := r.recalculateJobLevelTimeline(ctx, tx, tenantID, req.Code); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) UpdateJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, recordID uuid.UUID, roleCode string, parentRecord uuid.UUID, levelRank string, req *types.UpdateJobLevelRequest) (*types.JobLevel, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	query := `UPDATE job_levels
SET role_code = $1,
    parent_record_id = $2,
    level_rank = $3,
    name = $4,
    description = $5,
    status = $6,
    effective_date = $7,
    updated_at = NOW()
WHERE tenant_id = $8 AND record_id = $9
RETURNING record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current`

	var entry types.JobLevel
	err = r.queryRow(ctx, tx, query,
		roleCode,
		parentRecord,
		levelRank,
		req.Name,
		normalizeOptionalString(req.Description),
		req.Status,
		effectiveDate,
		tenantID,
		recordID,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.RoleCode,
		&entry.ParentRecord,
		&entry.LevelRank,
		&entry.Name,
		&entry.Description,
		&entry.SalaryBand,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return nil, fmt.Errorf("job level already exists for effective date: %w", err)
			case "23503":
				return nil, fmt.Errorf("parent job role not found: %w", err)
			}
		}
		return nil, fmt.Errorf("failed to update job level: %w", err)
	}

	if err := r.recalculateJobLevelTimeline(ctx, tx, tenantID, code); err != nil {
		return nil, err
	}

	return r.GetJobLevelByRecordID(ctx, tx, tenantID, recordID)
}

func (r *JobCatalogRepository) InsertJobLevelVersion(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, parentRecord uuid.UUID, req *types.JobCatalogVersionRequest) (*types.JobLevel, error) {
	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("invalid effective date: %w", err)
	}

	parent, err := r.GetJobLevelByRecordID(ctx, tx, tenantID, parentRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to load parent job level: %w", err)
	}
	if parent == nil {
		return nil, fmt.Errorf("job level parent record not found")
	}

	normalizedCode := strings.ToUpper(strings.TrimSpace(code))
	if !strings.EqualFold(parent.Code, normalizedCode) {
		return nil, fmt.Errorf("job level parent record mismatch")
	}

	var description interface{}
	if req.Description != nil {
		description = normalizeOptionalString(req.Description)
	} else if parent.Description.Valid {
		desc := strings.TrimSpace(parent.Description.String)
		if desc != "" {
			description = desc
		}
	}

	var salaryBand interface{}
	if len(parent.SalaryBand) > 0 {
		salaryBand = parent.SalaryBand
	}

	query := `INSERT INTO job_levels (
tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULL,$11,NOW(),NOW())
RETURNING record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current`

	var entry types.JobLevel
	err = r.queryRow(ctx, tx, query,
		tenantID,
		normalizedCode,
		parent.RoleCode,
		parent.ParentRecord,
		parent.LevelRank,
		req.Name,
		description,
		salaryBand,
		req.Status,
		effectiveDate,
		false,
	).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.RoleCode,
		&entry.ParentRecord,
		&entry.LevelRank,
		&entry.Name,
		&entry.Description,
		&entry.SalaryBand,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("job level version already exists for effective date")
			} else if pqErr.Code == "23503" {
				return nil, fmt.Errorf("parent job role not found")
			}
		}
		return nil, fmt.Errorf("failed to insert job level version: %w", err)
	}

	if err := r.recalculateJobLevelTimeline(ctx, tx, tenantID, normalizedCode); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (r *JobCatalogRepository) recalculateJobLevelTimeline(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) error {
	query := `SELECT record_id, effective_date, end_date, is_current FROM job_levels WHERE tenant_id = $1 AND level_code = $2 ORDER BY effective_date FOR UPDATE`
	rows, err := r.queryRows(ctx, tx, query, tenantID, code)
	if err != nil {
		return fmt.Errorf("failed to load job level timeline: %w", err)
	}
	defer rows.Close()

	var timeline []temporalRow
	for rows.Next() {
		var row temporalRow
		if err := rows.Scan(&row.RecordID, &row.EffectiveDate, &row.EndDate, &row.IsCurrent); err != nil {
			return fmt.Errorf("failed to scan job level timeline: %w", err)
		}
		timeline = append(timeline, row)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("job level timeline iteration error: %w", err)
	}

	normalized := normalizeTemporal(timeline)
	return r.applyTemporalUpdates(ctx, tx, "job_levels", normalized)
}

// Lookup helpers

func (r *JobCatalogRepository) GetFamilyGroupByRecordID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID) (*types.JobFamilyGroup, error) {
	query := `SELECT record_id, tenant_id, family_group_code, name, description, status, effective_date, end_date, is_current
FROM job_family_groups WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`
	var entry types.JobFamilyGroup
	err := r.queryRow(ctx, tx, query, tenantID, recordID).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load job family group by record id: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) GetJobFamilyByRecordID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID) (*types.JobFamily, error) {
	query := `SELECT record_id, tenant_id, family_code, family_group_code, parent_record_id, name, description, status, effective_date, end_date, is_current
FROM job_families WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`
	var entry types.JobFamily
	err := r.queryRow(ctx, tx, query, tenantID, recordID).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyGroupCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load job family by record id: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) GetJobRoleByRecordID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID) (*types.JobRole, error) {
	query := `SELECT record_id, tenant_id, role_code, family_code, parent_record_id, name, description, competency_model, status, effective_date, end_date, is_current
FROM job_roles WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`
	var entry types.JobRole
	err := r.queryRow(ctx, tx, query, tenantID, recordID).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.FamilyCode,
		&entry.ParentRecord,
		&entry.Name,
		&entry.Description,
		&entry.Competency,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load job role by record id: %w", err)
	}
	return &entry, nil
}

func (r *JobCatalogRepository) GetJobLevelByRecordID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, recordID uuid.UUID) (*types.JobLevel, error) {
	query := `SELECT record_id, tenant_id, level_code, role_code, parent_record_id, level_rank, name, description, salary_band, status, effective_date, end_date, is_current
FROM job_levels WHERE tenant_id = $1 AND record_id = $2 LIMIT 1`
	var entry types.JobLevel
	err := r.queryRow(ctx, tx, query, tenantID, recordID).Scan(
		&entry.RecordID,
		&entry.TenantID,
		&entry.Code,
		&entry.RoleCode,
		&entry.ParentRecord,
		&entry.LevelRank,
		&entry.Name,
		&entry.Description,
		&entry.SalaryBand,
		&entry.Status,
		&entry.EffectiveDate,
		&entry.EndDate,
		&entry.IsCurrent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load job level by record id: %w", err)
	}
	return &entry, nil
}
