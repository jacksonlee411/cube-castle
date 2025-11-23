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

	_ "github.com/lib/pq"
)

type report struct {
	GeneratedAt             time.Time   `json:"generatedAt"`
	LegacyCount             int         `json:"legacyCount"`
	StandardObjectVersions  int         `json:"standardObjectVersions"`
	ValidityOverlapCount    int         `json:"validityOverlapCount"`
	TransactionOverlapCount int         `json:"transactionOverlapCount"`
	TransactionGapCount     int         `json:"transactionGapCount"`
	ValidityOverlaps        []violation `json:"validityOverlaps"`
	TransactionOverlaps     []violation `json:"transactionOverlaps"`
	TransactionGaps         []gapWindow `json:"transactionGaps"`
}

type violation struct {
	Code     string    `json:"code"`
	ObjectID string    `json:"objectId"`
	WindowA  time.Time `json:"windowA"`
	WindowB  time.Time `json:"windowB"`
}

type gapWindow struct {
	Code        string     `json:"code"`
	PreviousEnd *time.Time `json:"previousEnd"`
	NextStart   time.Time  `json:"nextStart"`
}

const (
	validityOverlapBase = `
SELECT so.id,
       so.code,
       lower(sov1.validity_range) AS win_a,
       lower(sov2.validity_range) AS win_b
FROM standard_object_versions sov1
JOIN standard_object_versions sov2 ON sov1.object_id = sov2.object_id AND sov1.id <> sov2.id
JOIN standard_objects so ON so.id = sov1.object_id
WHERE sov1.validity_range && sov2.validity_range
  AND lower(sov1.validity_range) < lower(sov2.validity_range)`
	transactionOverlapBase = `
SELECT so.id,
       so.code,
       lower(sov1.transaction_range) AS win_a,
       lower(sov2.transaction_range) AS win_b
FROM standard_object_versions sov1
JOIN standard_object_versions sov2 ON sov1.object_id = sov2.object_id AND sov1.id <> sov2.id
JOIN standard_objects so ON so.id = sov1.object_id
WHERE sov1.transaction_range && sov2.transaction_range
  AND lower(sov1.transaction_range) < lower(sov2.transaction_range)`
	transactionGapBase = `
WITH ordered AS (
    SELECT so.code,
           lag(upper(sov.transaction_range)) OVER (PARTITION BY so.id ORDER BY lower(sov.transaction_range)) AS prev_end,
           lower(sov.transaction_range) AS next_start
    FROM standard_object_versions sov
    JOIN standard_objects so ON so.id = sov.object_id
)
SELECT code, prev_end, next_start
FROM ordered
WHERE prev_end IS NOT NULL
  AND prev_end < next_start`
)

func main() {
	var (
		dsn     = flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
		file    = flag.String("out", "", "Optional output path (defaults to logs/plan402/validator/<ts>-report.json)")
		logFile = flag.String("log-file", "", "Path to validator log (defaults to logs/plan402/validator/<ts>-run.log)")
		tcLog   = flag.String("time-constraint-log", filepath.Join("logs", "plan402", "migration", "time-constraint-report.log"), "JSONL path for TC report (appends)")
		gapLog  = flag.String("transaction-gap-log", filepath.Join("logs", "plan402", "migration", "transaction-gap.log"), "JSONL path for transaction gap report (appends)")
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

	logger, stopLogger, err := setupLogger(*logFile)
	if err != nil {
		log.Fatalf("setup logger: %v", err)
	}
	defer stopLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rep := report{GeneratedAt: time.Now().UTC()}
	rep.LegacyCount, err = scalarInt(ctx, db, "SELECT COUNT(*) FROM organization_units")
	if err != nil {
		logger.Fatalf("count legacy: %v", err)
	}
	rep.StandardObjectVersions, err = scalarInt(ctx, db, "SELECT COUNT(*) FROM standard_object_versions")
	if err != nil {
		logger.Fatalf("count standard_object_versions: %v", err)
	}
	rep.ValidityOverlaps, err = fetchViolations(ctx, db, validityOverlapBase)
	if err != nil {
		logger.Fatalf("validity overlaps: %v", err)
	}
	rep.ValidityOverlapCount, err = countRows(ctx, db, validityOverlapBase)
	if err != nil {
		logger.Fatalf("count validity overlaps: %v", err)
	}
	rep.TransactionOverlaps, err = fetchViolations(ctx, db, transactionOverlapBase)
	if err != nil {
		logger.Fatalf("transaction overlaps: %v", err)
	}
	rep.TransactionOverlapCount, err = countRows(ctx, db, transactionOverlapBase)
	if err != nil {
		logger.Fatalf("count transaction overlaps: %v", err)
	}
	rep.TransactionGaps, err = fetchGaps(ctx, db, transactionGapBase)
	if err != nil {
		logger.Fatalf("transaction gaps: %v", err)
	}
	rep.TransactionGapCount, err = countRows(ctx, db, transactionGapBase)
	if err != nil {
		logger.Fatalf("count transaction gaps: %v", err)
	}
	if *file == "" {
		ts := rep.GeneratedAt.Format("20060102-150405")
		*file = filepath.Join("logs", "plan402", "validator", fmt.Sprintf("%s-report.json", ts))
	}
	if err := writeReport(*file, rep); err != nil {
		logger.Fatalf("write report: %v", err)
	}
	if err := appendJSONLine(*tcLog, map[string]any{
		"generatedAt":             rep.GeneratedAt,
		"validityOverlapCount":    rep.ValidityOverlapCount,
		"transactionOverlapCount": rep.TransactionOverlapCount,
	}); err != nil {
		logger.Fatalf("write time constraint log: %v", err)
	}
	if err := appendJSONLine(*gapLog, map[string]any{
		"generatedAt":         rep.GeneratedAt,
		"transactionGaps":     rep.TransactionGaps,
		"transactionGapCount": rep.TransactionGapCount,
	}); err != nil {
		logger.Fatalf("write transaction gap log: %v", err)
	}
	logger.Printf("validator report written to %s", *file)
}

func scalarInt(ctx context.Context, db *sql.DB, query string) (int, error) {
	var result int
	err := db.QueryRowContext(ctx, query).Scan(&result)
	return result, err
}

func fetchViolations(ctx context.Context, db *sql.DB, baseQuery string) ([]violation, error) {
	query := baseQuery + " LIMIT 5"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []violation
	for rows.Next() {
		var v violation
		if err := rows.Scan(&v.ObjectID, &v.Code, &v.WindowA, &v.WindowB); err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, rows.Err()
}

func writeReport(path string, rep report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fetchGaps(ctx context.Context, db *sql.DB, baseQuery string) ([]gapWindow, error) {
	query := baseQuery + " LIMIT 5"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []gapWindow
	for rows.Next() {
		var (
			code      string
			prevEnd   sql.NullTime
			nextStart time.Time
		)
		if err := rows.Scan(&code, &prevEnd, &nextStart); err != nil {
			return nil, err
		}
		var prevPtr *time.Time
		if prevEnd.Valid {
			t := prevEnd.Time
			prevPtr = &t
		}
		res = append(res, gapWindow{
			Code:        code,
			PreviousEnd: prevPtr,
			NextStart:   nextStart,
		})
	}
	return res, rows.Err()
}

func countRows(ctx context.Context, db *sql.DB, baseQuery string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS q", baseQuery)
	return scalarInt(ctx, db, query)
}

func appendJSONLine(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func setupLogger(path string) (*log.Logger, func(), error) {
	if path == "" {
		ts := time.Now().UTC().Format("20060102-150405")
		path = filepath.Join("logs", "plan402", "validator", fmt.Sprintf("%s-run.log", ts))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	writer := io.MultiWriter(os.Stdout, file)
	logger := log.New(writer, "[validator] ", log.LstdFlags)
	cleanup := func() {
		_ = file.Close()
	}
	logger.Printf("logging to %s", path)
	return logger, cleanup, nil
}
