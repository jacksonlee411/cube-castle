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
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type edge struct {
	parentID  string
	childID   string
	childCode string
	childName string
}

type objectInfo struct {
	ID   string
	Code string
	Name string
}

type snapshotRow struct {
	AncestorID   string
	DescendantID string
	CodePath     string
	NamePath     string
	Depth        int
}

func main() {
	var (
		dsn          = flag.String("dsn", os.Getenv("DATABASE_URL"), "PostgreSQL DSN")
		tenant       = flag.String("tenant", "", "Tenant code (required)")
		objectType   = flag.String("object-type", "ORGANIZATION_UNIT", "Object type to refresh")
		asOfValidStr = flag.String("as-of-valid", "", "As-of valid timestamp (RFC3339, default now)")
		asOfTxnStr   = flag.String("as-of-transaction", "", "As-of transaction timestamp (RFC3339, default now)")
		logFile      = flag.String("log-file", "", "Path to refresh log (defaults to logs/plan402/snapshots/<ts>-refresh.log)")
		metricsFile  = flag.String("metrics-file", "", "Path to metrics log (defaults to logs/plan402/metrics/<ts>-snapshots.jsonl)")
	)
	flag.Parse()
	if *dsn == "" || *tenant == "" {
		log.Fatal("dsn and tenant are required")
	}

	asOfValid := parseTime(*asOfValidStr)
	asOfTxn := parseTime(*asOfTxnStr)
	if asOfValid.IsZero() {
		asOfValid = time.Now().UTC()
	}
	if asOfTxn.IsZero() {
		asOfTxn = time.Now().UTC()
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	start := time.Now()
	rows, roots, err := refreshSnapshots(ctx, db, *tenant, *objectType, asOfValid, asOfTxn)
	if err != nil {
		logger.Fatalf("refresh failed: %v", err)
	}
	duration := time.Since(start)
	transactionLag := time.Since(asOfTxn)
	metricsPath, err := writeMetrics(*metricsFile, snapshotMetrics{
		GeneratedAt:      time.Now().UTC(),
		Tenant:           *tenant,
		ObjectType:       *objectType,
		AsOfValid:        asOfValid,
		AsOfTransaction:  asOfTxn,
		InsertedRows:     rows,
		RootCount:        roots,
		TransactionLag:   transactionLag.String(),
		ExecutionLatency: duration.String(),
	})
	if err != nil {
		logger.Fatalf("write metrics: %v", err)
	}
	logger.Printf("snapshot refresh complete tenant=%s objectType=%s rows=%d roots=%d transaction_lag=%s duration=%s metrics=%s",
		*tenant, *objectType, rows, roots, transactionLag.String(), duration.String(), metricsPath)
}

func parseTime(v string) time.Time {
	if v == "" {
		return time.Time{}
	}
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		log.Fatalf("parse time %q: %v", v, err)
	}
	return ts.UTC()
}

func refreshSnapshots(ctx context.Context, db *sql.DB, tenant, objectType string, asOfValid, asOfTxn time.Time) (int, int, error) {
	objects, err := fetchObjects(ctx, db, tenant, objectType)
	if err != nil {
		return 0, 0, err
	}
	edges, err := fetchEdges(ctx, db, tenant)
	if err != nil {
		return 0, 0, err
	}
	children := make(map[string][]edge)
	childSet := make(map[string]struct{})
	for _, e := range edges {
		children[e.parentID] = append(children[e.parentID], e)
		childSet[e.childID] = struct{}{}
	}

	var roots []string
	for id := range objects {
		if _, ok := childSet[id]; ok {
			continue
		}
		roots = append(roots, id)
	}
	if len(roots) == 0 {
		for id := range objects {
			roots = append(roots, id)
		}
	}
	rootCount := len(roots)

	var rows []snapshotRow
	for _, rootID := range roots {
		info := objects[rootID]
		rows = append(rows, snapshotRow{
			AncestorID:   info.ID,
			DescendantID: info.ID,
			CodePath:     "/" + info.Code,
			NamePath:     "/" + info.Name,
			Depth:        0,
		})
		rows = append(rows, traverse(children, objects, info.ID, info.ID, []string{info.Code}, []string{info.Name}, 1)...)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM standard_object_hierarchy_snapshots WHERE tenant_code=$1 AND object_type=$2 AND as_of_valid=$3`, tenant, objectType, asOfValid); err != nil {
		return 0, 0, err
	}

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO standard_object_hierarchy_snapshots (
    snapshot_id,
    tenant_code,
    object_type,
    as_of_valid,
    as_of_transaction,
    ancestor_object_id,
    descendant_object_id,
    depth,
    code_path,
    name_path,
    metadata
) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, '{}'::jsonb)`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	for _, row := range rows {
		if _, err := stmt.ExecContext(ctx, tenant, objectType, asOfValid, asOfTxn, row.AncestorID, row.DescendantID, row.Depth, row.CodePath, row.NamePath); err != nil {
			return 0, 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}
	return len(rows), rootCount, nil
}

func fetchObjects(ctx context.Context, db *sql.DB, tenant, objectType string) (map[string]objectInfo, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, code, display_name
FROM standard_objects
WHERE tenant_code = $1
  AND object_type = $2`, tenant, objectType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]objectInfo)
	for rows.Next() {
		var info objectInfo
		if err := rows.Scan(&info.ID, &info.Code, &info.Name); err != nil {
			return nil, err
		}
		result[info.ID] = info
	}
	return result, rows.Err()
}

func fetchEdges(ctx context.Context, db *sql.DB, tenant string) ([]edge, error) {
	rows, err := db.QueryContext(ctx, `
SELECT src.id, tgt.id, tgt.code, tgt.display_name
FROM standard_object_links l
JOIN standard_objects src ON src.id = l.source_object_id
JOIN standard_objects tgt ON tgt.id = l.target_object_id
WHERE l.tenant_code = $1
  AND l.link_type = 'ORG_HIERARCHY'`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []edge
	for rows.Next() {
		var e edge
		if err := rows.Scan(&e.parentID, &e.childID, &e.childCode, &e.childName); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func traverse(children map[string][]edge, objs map[string]objectInfo, ancestorID, currentID string, codePath, namePath []string, depth int) []snapshotRow {
	var rows []snapshotRow
	for _, child := range children[currentID] {
		newCodePath := append(codePath, child.childCode)
		newNamePath := append(namePath, child.childName)
		rows = append(rows, snapshotRow{
			AncestorID:   ancestorID,
			DescendantID: child.childID,
			CodePath:     "/" + strings.Join(newCodePath, "/"),
			NamePath:     "/" + strings.Join(newNamePath, "/"),
			Depth:        depth,
		})
		rows = append(rows, traverse(children, objs, ancestorID, child.childID, newCodePath, newNamePath, depth+1)...)
	}
	return rows
}

type snapshotMetrics struct {
	GeneratedAt      time.Time `json:"generatedAt"`
	Tenant           string    `json:"tenant"`
	ObjectType       string    `json:"objectType"`
	AsOfValid        time.Time `json:"asOfValid"`
	AsOfTransaction  time.Time `json:"asOfTransaction"`
	InsertedRows     int       `json:"insertedRows"`
	RootCount        int       `json:"rootCount"`
	TransactionLag   string    `json:"transactionLag"`
	ExecutionLatency string    `json:"executionLatency"`
}

func writeMetrics(path string, entry snapshotMetrics) (string, error) {
	if path == "" {
		ts := time.Now().UTC().Format("20060102-150405")
		path = filepath.Join("logs", "plan402", "metrics", fmt.Sprintf("%s-snapshots.jsonl", ts))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return "", err
	}
	return path, nil
}

func setupLogger(path string) (*log.Logger, func(), error) {
	if path == "" {
		ts := time.Now().UTC().Format("20060102-150405")
		path = filepath.Join("logs", "plan402", "snapshots", fmt.Sprintf("%s-refresh.log", ts))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}
	writer := io.MultiWriter(os.Stdout, file)
	logger := log.New(writer, "[snapshot-refresh] ", log.LstdFlags)
	cleanup := func() {
		_ = file.Close()
	}
	logger.Printf("logging to %s", path)
	return logger, cleanup, nil
}
