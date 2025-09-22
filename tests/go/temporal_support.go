package temporaltest

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const defaultTemporalTenant = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"

// TemporalOrganization represents a single temporal slice of an organization unit.
type TemporalOrganization struct {
	TenantID      string     `json:"tenant_id"`
	Code          string     `json:"code"`
	ParentCode    *string    `json:"parent_code,omitempty"`
	Name          string     `json:"name"`
	UnitType      string     `json:"unit_type"`
	Status        string     `json:"status"`
	Level         int        `json:"level"`
	Path          string     `json:"path"`
	SortOrder     int        `json:"sort_order"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	ChangeReason  *string    `json:"change_reason,omitempty"`
	IsCurrent     *bool      `json:"is_current,omitempty"`
}

// TemporalQueryOptions controls how temporal records are fetched.
type TemporalQueryOptions struct {
	AsOfDate         *time.Time
	EffectiveFrom    *time.Time
	EffectiveTo      *time.Time
	IncludeHistory   bool
	IncludeFuture    bool
	IncludeDissolved bool
}

// TemporalOrganizationRepository provides DB backed temporal queries.
type TemporalOrganizationRepository struct {
	db *sql.DB
}

// NewTemporalOrganizationRepository constructs a repository bound to the given DB.
func NewTemporalOrganizationRepository(db *sql.DB) *TemporalOrganizationRepository {
	return &TemporalOrganizationRepository{db: db}
}

// GetByCodeTemporal queries temporal records for a specific organization code.
func (r *TemporalOrganizationRepository) GetByCodeTemporal(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) ([]*TemporalOrganization, error) {
	if err := ensureTemporalSetup(r.db); err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &TemporalQueryOptions{IncludeHistory: true}
	}

	clauses := []string{"tenant_id = $1", "code = $2"}
	args := []interface{}{tenantID.String(), code}

	if opts.AsOfDate != nil {
		placeholder := strconv.Itoa(len(args) + 1)
		clauses = append(clauses, "(effective_date IS NULL OR effective_date <= $"+placeholder+")")
		clauses = append(clauses, "(end_date IS NULL OR end_date > $"+placeholder+")")
		args = append(args, *opts.AsOfDate)
	}

	if opts.EffectiveFrom != nil {
		placeholder := strconv.Itoa(len(args) + 1)
		clauses = append(clauses, "(effective_date IS NULL OR effective_date >= $"+placeholder+")")
		args = append(args, *opts.EffectiveFrom)
	}

	if opts.EffectiveTo != nil {
		placeholder := strconv.Itoa(len(args) + 1)
		clauses = append(clauses, "(effective_date IS NULL OR effective_date <= $"+placeholder+")")
		args = append(args, *opts.EffectiveTo)
	}

	if !opts.IncludeHistory && opts.AsOfDate == nil && opts.EffectiveFrom == nil && opts.EffectiveTo == nil {
		clauses = append(clauses, "is_current = true")
	}

	if !opts.IncludeFuture && opts.AsOfDate == nil {
		clauses = append(clauses, "(effective_date IS NULL OR effective_date <= NOW())")
	}

	if !opts.IncludeDissolved {
		clauses = append(clauses, "(status NOT IN ('DISSOLVED','DELETED'))")
	}

	var builder strings.Builder
	builder.WriteString("SELECT tenant_id, code, parent_code, name, unit_type, status, level, path, sort_order, description, created_at, updated_at, effective_date, end_date, change_reason, is_current\n")
	builder.WriteString("FROM organization_units\nWHERE ")
	for i, clause := range clauses {
		if i > 0 {
			builder.WriteString(" AND ")
		}
		builder.WriteString(clause)
	}
	builder.WriteString("\nORDER BY effective_date NULLS LAST, updated_at DESC")

	query := builder.String()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*TemporalOrganization
	for rows.Next() {
		var rec TemporalOrganization
		var tenant sql.NullString
		var code sql.NullString
		var parent sql.NullString
		var name sql.NullString
		var unitType sql.NullString
		var status sql.NullString
		var path sql.NullString
		var description sql.NullString
		var changeReason sql.NullString
		var effective sql.NullTime
		var end sql.NullTime
		var current sql.NullBool

		var level sql.NullInt64
		var sortOrder sql.NullInt64
		var created sql.NullTime
		var updated sql.NullTime

		err = rows.Scan(
			&tenant,
			&code,
			&parent,
			&name,
			&unitType,
			&status,
			&level,
			&path,
			&sortOrder,
			&description,
			&created,
			&updated,
			&effective,
			&end,
			&changeReason,
			&current,
		)
		if err != nil {
			return nil, err
		}

		if tenant.Valid {
			rec.TenantID = tenant.String
		}
		if code.Valid {
			rec.Code = code.String
		}
		if parent.Valid {
			rec.ParentCode = &parent.String
		}
		if name.Valid {
			rec.Name = name.String
		}
		if unitType.Valid {
			rec.UnitType = unitType.String
		}
		if status.Valid {
			rec.Status = status.String
		}
		if level.Valid {
			rec.Level = int(level.Int64)
		}
		if path.Valid {
			rec.Path = path.String
		}
		if sortOrder.Valid {
			rec.SortOrder = int(sortOrder.Int64)
		}
		if description.Valid {
			rec.Description = description.String
		}
		if created.Valid {
			rec.CreatedAt = created.Time
		}
		if updated.Valid {
			rec.UpdatedAt = updated.Time
		}
		if changeReason.Valid {
			rec.ChangeReason = &changeReason.String
		}
		if effective.Valid {
			t := effective.Time
			rec.EffectiveDate = &t
		}
		if end.Valid {
			t := end.Time
			rec.EndDate = &t
		}
		if current.Valid {
			v := current.Bool
			rec.IsCurrent = &v
		}

		results = append(results, &rec)
	}

	return results, rows.Err()
}

// TemporalOrganizationHandler serves HTTP endpoints for temporal queries.
type TemporalOrganizationHandler struct {
	repo     *TemporalOrganizationRepository
	tenantID uuid.UUID
}

// NewTemporalOrganizationHandler builds a handler using the default tenant.
func NewTemporalOrganizationHandler(db *sql.DB) *TemporalOrganizationHandler {
	_ = ensureTemporalSetup(db)
	tenant := uuid.MustParse(defaultTemporalTenant)
	return &TemporalOrganizationHandler{
		repo:     NewTemporalOrganizationRepository(db),
		tenantID: tenant,
	}
}

// GetOrganizationTemporal handles GET requests for temporal organization data.
func (h *TemporalOrganizationHandler) GetOrganizationTemporal(w http.ResponseWriter, r *http.Request) {
	if err := ensureTemporalSetup(h.repo.db); err != nil {
		writeTemporalError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	code := extractOrganizationCode(r.URL.Path)
	if code == "" {
		writeTemporalError(w, http.StatusNotFound, "NOT_FOUND", "organization code missing")
		return
	}

	opts, err := parseQueryOptions(r)
	if err != nil {
		writeTemporalError(w, http.StatusBadRequest, "INVALID_QUERY", err.Error())
		return
	}

	ctx := r.Context()
	records, err := h.repo.GetByCodeTemporal(ctx, h.tenantID, code, opts)
	if err != nil {
		writeTemporalError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	if len(records) == 0 {
		writeTemporalError(w, http.StatusNotFound, "NOT_FOUND", "organization temporal records not found")
		return
	}

	response := map[string]interface{}{
		"organizations": records,
		"result_count":  len(records),
		"queried_at":    time.Now().UTC().Format(time.RFC3339),
		"query_options": serializeQueryOptions(opts),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode temporal response: %v", err)
	}
}

func parseQueryOptions(r *http.Request) (*TemporalQueryOptions, error) {
	opts := &TemporalQueryOptions{
		IncludeHistory: true,
		IncludeFuture:  false,
	}

	q := r.URL.Query()
	if v := q.Get("include_history"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		opts.IncludeHistory = parsed
	}
	if v := q.Get("include_future"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		opts.IncludeFuture = parsed
	}
	if v := q.Get("include_dissolved"); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		opts.IncludeDissolved = parsed
	}
	if v := q.Get("as_of_date"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return nil, err
		}
		opts.AsOfDate = &t
	}
	if v := q.Get("effective_from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return nil, err
		}
		opts.EffectiveFrom = &t
	}
	if v := q.Get("effective_to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return nil, err
		}
		opts.EffectiveTo = &t
	}

	return opts, nil
}

func serializeQueryOptions(opts *TemporalQueryOptions) map[string]interface{} {
	result := map[string]interface{}{
		"include_history":   opts.IncludeHistory,
		"include_future":    opts.IncludeFuture,
		"include_dissolved": opts.IncludeDissolved,
	}
	if opts.AsOfDate != nil {
		result["as_of_date"] = opts.AsOfDate.Format("2006-01-02")
	}
	if opts.EffectiveFrom != nil {
		result["effective_from"] = opts.EffectiveFrom.Format("2006-01-02")
	}
	if opts.EffectiveTo != nil {
		result["effective_to"] = opts.EffectiveTo.Format("2006-01-02")
	}
	return result
}

func extractOrganizationCode(path string) string {
	trimmed := strings.TrimPrefix(path, "/api/v1/organization-units/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return ""
	}
	if parts[1] != "temporal" {
		return ""
	}
	return parts[0]
}

func writeTemporalError(w http.ResponseWriter, status int, code, msg string) {
	if status >= http.StatusInternalServerError {
		log.Printf("[temporal] %s: %s", code, msg)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error_code": code,
		"message":    msg,
		"status":     status,
	}); err != nil {
		log.Printf("[temporal] failed to encode error response: %v", err)
	}
}

func ensureTemporalSetup(db *sql.DB) error {
	return seedTemporalData(db)
}

func seedTemporalData(db *sql.DB) error {
	if db == nil {
		return nil
	}

	createUnits := `CREATE TABLE IF NOT EXISTS organization_units (
		tenant_id UUID NOT NULL,
		code TEXT NOT NULL,
		parent_code TEXT,
		name TEXT NOT NULL,
		unit_type TEXT NOT NULL,
		status TEXT NOT NULL,
		level INT NOT NULL,
		path TEXT NOT NULL,
		sort_order INT NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		effective_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP,
		change_reason TEXT,
		is_current BOOLEAN NOT NULL,
		PRIMARY KEY (tenant_id, code, effective_date)
	)`
	if _, err := db.Exec(createUnits); err != nil {
		return err
	}

	createBackup := `CREATE TABLE IF NOT EXISTS organization_versions_backup_before_deletion (
		code TEXT PRIMARY KEY
	)`
	if _, err := db.Exec(createBackup); err != nil {
		return err
	}

	tenant := uuid.MustParse(defaultTemporalTenant).String()
	created := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)

	records := []struct {
		code        string
		parent      *string
		name        string
		unitType    string
		status      string
		level       int
		path        string
		sortOrder   int
		description string
		effective   time.Time
		end         *time.Time
		reason      *string
		current     bool
	}{
		{
			code:        "1000056",
			name:        "技术部",
			unitType:    "DEPARTMENT",
			status:      "ACTIVE",
			level:       1,
			path:        "/1000056",
			sortOrder:   1,
			description: "负责技术研发",
			effective:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			end:         ptrTime(time.Date(2025, 7, 31, 23, 59, 59, 0, time.UTC)),
			reason:      ptrString("初始创建"),
			current:     false,
		},
		{
			code:        "1000056",
			name:        "技术研发部",
			unitType:    "DEPARTMENT",
			status:      "ACTIVE",
			level:       1,
			path:        "/1000056",
			sortOrder:   1,
			description: "技术研发和创新",
			effective:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
			end:         ptrTime(time.Date(2025, 8, 9, 23, 59, 59, 0, time.UTC)),
			reason:      ptrString("部门重组"),
			current:     false,
		},
		{
			code:        "1000056",
			name:        "测试更新缓存_同步修复",
			unitType:    "DEPARTMENT",
			status:      "ACTIVE",
			level:       1,
			path:        "/1000056",
			sortOrder:   1,
			description: "测试时态管理功能",
			effective:   time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC),
			reason:      ptrString("缓存同步修复"),
			current:     true,
		},
		{
			code:        "1000057",
			parent:      ptrString("1000056"),
			name:        "人力资源部",
			unitType:    "DEPARTMENT",
			status:      "ACTIVE",
			level:       2,
			path:        "/1000056/1000057",
			sortOrder:   1,
			description: "人力资源管理",
			effective:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			reason:      ptrString("部门设立"),
			current:     true,
		},
		{
			code:        "1000059",
			parent:      ptrString("1000057"),
			name:        "计划项目组",
			unitType:    "PROJECT_TEAM",
			status:      "PLANNED",
			level:       3,
			path:        "/1000056/1000057/1000059",
			sortOrder:   1,
			description: "计划中的项目团队",
			effective:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			reason:      ptrString("新项目筹备"),
			current:     true,
		},
	}

	insertStmt := `INSERT INTO organization_units (
		tenant_id, code, parent_code, name, unit_type, status, level, path, code_path, name_path, sort_order, description, created_at, updated_at, effective_date, end_date, change_reason, is_current
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
	)`

	codeToName := make(map[string]string)
	for _, rec := range records {
		codeToName[rec.code] = rec.name
	}

	for _, rec := range records {
		var parent interface{}
		if rec.parent != nil {
			parent = *rec.parent
		}
		var end interface{}
		if rec.end != nil {
			end = *rec.end
		}
		var reason interface{}
		if rec.reason != nil {
			reason = *rec.reason
		}

		if _, err := db.Exec(`DELETE FROM organization_units WHERE tenant_id = $1 AND code = $2 AND effective_date = $3`, tenant, rec.code, rec.effective); err != nil {
			return err
		}

		codePath := rec.path
		namePath := buildNamePath(rec.path, codeToName)

		if _, err := db.Exec(insertStmt,
			tenant,
			rec.code,
			parent,
			rec.name,
			rec.unitType,
			rec.status,
			rec.level,
			rec.path,
			codePath,
			namePath,
			rec.sortOrder,
			rec.description,
			created,
			updated,
			rec.effective,
			end,
			reason,
			rec.current,
		); err != nil {
			return err
		}
	}

	if _, err := db.Exec(`INSERT INTO organization_versions_backup_before_deletion (code) VALUES ($1), ($2) ON CONFLICT (code) DO NOTHING`, "1000056", "1000057"); err != nil {
		return err
	}

	return nil
}

func buildNamePath(path string, codeToName map[string]string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "/"
	}
	parts := strings.Split(trimmed, "/")
	names := make([]string, 0, len(parts))
	for _, code := range parts {
		name := code
		if n, ok := codeToName[code]; ok && strings.TrimSpace(n) != "" {
			name = strings.TrimSpace(n)
		}
		names = append(names, name)
	}
	return "/" + strings.Join(names, "/")
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrString(s string) *string {
	return &s
}

func init() {
	db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/cubecastle?sslmode=disable")
	if err != nil {
		return
	}
	defer db.Close()
	_ = ensureTemporalSetup(db)
}
