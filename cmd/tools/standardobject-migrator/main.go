//go:build legacy
// +build legacy

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"cube-castle/internal/standardobject"
	"cube-castle/internal/standardobject/repository"
	temporalclock "cube-castle/pkg/temporal/clock"
	_ "github.com/lib/pq"
)

type legacyRecord struct {
	RecordID      string
	TenantID      string
	Code          string
	ParentCode    sql.NullString
	Name          string
	UnitType      string
	Status        string
	Description   sql.NullString
	Profile       []byte
	Metadata      []byte
	CreatedAt     time.Time
	UpdatedAt     time.Time
	EffectiveDate time.Time
	EndDate       sql.NullTime
	ChangeReason  sql.NullString
	IsCurrent     bool
	SortOrder     int
}

type migrator struct {
	db     *sql.DB
	repo   *repository.Repository
	logger *log.Logger
	dryRun bool
}

func main() {
	var (
		dsn     = flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
		dryRun  = flag.Bool("dry-run", false, "Inspect without writing to SOM tables")
		limit   = flag.Int("limit", 0, "Optional limit on migrated records")
		logFile = flag.String("log-file", "", "Path to log file (defaults to logs/plan402/migration/<ts>-migrator.log)")
	)
	flag.Parse()
	if *dsn == "" {
		log.Fatal("DATABASE_URL not provided (env or --dsn)")
	}

	db, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	logger, stopLogger, err := setupLogger(*logFile)
	if err != nil {
		log.Fatalf("setup logger: %v", err)
	}
	defer stopLogger()

	m := &migrator{
		db:     db,
		repo:   repository.NewRepository(db, temporalclock.NewSystemClock()),
		logger: logger,
		dryRun: *dryRun,
	}
	if err := m.Run(ctx, *limit); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func (m *migrator) Run(ctx context.Context, limit int) error {
	m.logger.Printf("starting legacy export (dryRun=%v)...", m.dryRun)
	rows, err := m.selectLegacy(ctx, limit)
	if err != nil {
		return err
	}
	defer rows.Close()

	var processed int
	for rows.Next() {
		rec, err := m.scanLegacy(rows)
		if err != nil {
			return err
		}
		if err := m.migrateRecord(ctx, rec); err != nil {
			return err
		}
		processed++
		if processed%100 == 0 {
			m.logger.Printf("processed %d records", processed)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	m.logger.Printf("legacy export complete (count=%d)", processed)
	if err := m.upsertLinks(ctx); err != nil {
		return err
	}
	m.logger.Printf("link refresh complete")
	return nil
}

func (m *migrator) selectLegacy(ctx context.Context, limit int) (*sql.Rows, error) {
	query := `
SELECT record_id::text,
       tenant_id::text,
       code,
       parent_code,
       name,
       unit_type,
       status,
       description,
       profile,
       metadata,
       created_at,
       updated_at,
       effective_date,
       end_date,
       change_reason,
       is_current,
       sort_order
FROM organization_units
ORDER BY tenant_id, code, effective_date`
	if limit > 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}
	return m.db.QueryContext(ctx, query)
}

func (m *migrator) scanLegacy(rows *sql.Rows) (legacyRecord, error) {
	var rec legacyRecord
	err := rows.Scan(
		&rec.RecordID,
		&rec.TenantID,
		&rec.Code,
		&rec.ParentCode,
		&rec.Name,
		&rec.UnitType,
		&rec.Status,
		&rec.Description,
		&rec.Profile,
		&rec.Metadata,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.EffectiveDate,
		&rec.EndDate,
		&rec.ChangeReason,
		&rec.IsCurrent,
		&rec.SortOrder,
	)
	return rec, err
}

func (m *migrator) migrateRecord(ctx context.Context, rec legacyRecord) error {
	if m.dryRun {
		m.logger.Printf("dry-run: would migrate %s/%s", rec.TenantID, rec.Code)
		return nil
	}
	payload := map[string]any{
		"name":        rec.Name,
		"description": rec.Description.String,
		"unitType":    rec.UnitType,
	}
	if len(rec.Profile) > 0 {
		payload["profile"] = parseJSON(rec.Profile)
	}
	if len(rec.Metadata) > 0 {
		payload["metadata"] = parseJSON(rec.Metadata)
	}
	if rec.ChangeReason.Valid {
		payload["changeReason"] = rec.ChangeReason.String
	}

	version := standardobject.TemporalVersion{
		VersionCode:     standardobject.MakeVersionCode(rec.Code, rec.EffectiveDate, rec.UpdatedAt, rec.RecordID),
		EffectiveDate:   rec.EffectiveDate,
		EndDate:         nullableTime(rec.EndDate),
		IsCurrent:       rec.IsCurrent,
		Payload:         payload,
		AuditTrail:      map[string]any{"legacyStatus": rec.Status},
		Checksum:        "",
		CreatedAt:       rec.CreatedAt,
		UpdatedAt:       rec.UpdatedAt,
		TransactionFrom: rec.CreatedAt,
		TransactionTo:   nullableTime(sql.NullTime{Time: rec.UpdatedAt, Valid: !rec.UpdatedAt.IsZero()}),
	}

	kernel := standardobject.ObjectKernel{
		ObjectType:         standardobject.ObjectTypeOrganizationUnit,
		Code:               rec.Code,
		DisplayName:        rec.Name,
		TenantCode:         rec.TenantID,
		Status:             mapStatus(rec.Status),
		Labels:             map[string]string{"unitType": rec.UnitType},
		SchemaVersion:      "2025.11.402A",
		DataClassification: "internal",
		RetentionPolicy:    "standard",
		CreatedBy:          "",
		CreatedAt:          rec.CreatedAt,
		UpdatedAt:          rec.UpdatedAt,
	}

	err := m.repo.Upsert(ctx, standardobject.ObjectAggregate{
		Kernel:  kernel,
		Version: version,
	})
	return err
}

func (m *migrator) upsertLinks(ctx context.Context) error {
	if m.dryRun {
		return nil
	}
	query := `
INSERT INTO standard_object_links (
    link_type,
    source_object_id,
    target_object_id,
    tenant_code,
    validity_range,
    transaction_range,
    attributes,
    created_by,
    updated_by
) SELECT
    'ORG_HIERARCHY',
    parent_so.id,
    child_so.id,
    child_so.tenant_code,
    tstzrange(ou.effective_date::timestamptz, CASE WHEN ou.end_date IS NULL THEN NULL ELSE ou.end_date::timestamptz END, '[)'),
    tstzrange(ou.created_at, CASE WHEN ou.updated_at IS NULL THEN NULL ELSE ou.updated_at END, '[)'),
    jsonb_build_object('sortOrder', ou.sort_order),
    NULL,
    NULL
FROM organization_units ou
JOIN standard_objects child_so ON child_so.code = ou.code AND child_so.tenant_code = ou.tenant_id::text
JOIN standard_objects parent_so ON parent_so.code = ou.parent_code AND parent_so.tenant_code = ou.tenant_id::text
ON CONFLICT (link_type, source_object_id, target_object_id)
DO UPDATE SET
    validity_range    = EXCLUDED.validity_range,
    transaction_range = EXCLUDED.transaction_range,
    attributes        = EXCLUDED.attributes,
    updated_at        = now();`
	_, err := m.db.ExecContext(ctx, query)
	return err
}

func parseJSON(data []byte) any {
	var out any
	if len(data) == 0 {
		return map[string]any{}
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func mapStatus(status string) standardobject.LifecycleStatus {
	switch status {
	case "ACTIVE":
		return standardobject.StatusActive
	case "SUSPENDED":
		return standardobject.StatusSuspended
	case "RETIRED", "DELETED":
		return standardobject.StatusRetired
	default:
		return standardobject.StatusDraft
	}
}

func nullableTime(val sql.NullTime) *time.Time {
	if !val.Valid {
		return nil
	}
	return &val.Time
}

func setupLogger(path string) (*log.Logger, func(), error) {
	if path == "" {
		ts := time.Now().UTC().Format("20060102-150405")
		path = filepath.Join("logs", "plan402", "migration", fmt.Sprintf("%s-migrator.log", ts))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	writer := io.MultiWriter(os.Stdout, file)
	logger := log.New(writer, "[migrator] ", log.LstdFlags)
	cleanup := func() {
		_ = file.Close()
	}
	logger.Printf("logging to %s", path)
	return logger, cleanup, nil
}
