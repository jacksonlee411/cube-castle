package integration

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

func TestMigrationRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baseDSN := firstNonEmpty(
		os.Getenv("MIGRATION_TEST_DATABASE_URL"),
		os.Getenv("DATABASE_URL"),
		"postgres://user:password@localhost:5432/cubecastle?sslmode=disable",
	)

	adminDSN, testDSN, dbName := buildTestDSNs(t, baseDSN)

	admin, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("open admin connection: %v", err)
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", pqQuoteIdentifier(dbName))); err != nil {
		t.Fatalf("drop test database: %v", err)
	}
	if _, err := admin.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", pqQuoteIdentifier(dbName))); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		admin.ExecContext(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pqQuoteIdentifier(dbName)))
	})

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	db, err := sql.Open("pgx", testDSN)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	defer db.Close()

	migrationsDir := lookupProjectPath(t, "database", "migrations")

	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		t.Fatalf("goose up: %v", err)
	}

	assertTableExists(ctx, t, db, "organization_units")
	assertExtensionExists(ctx, t, db, "pgcrypto")

	if err := goose.DownContext(ctx, db, migrationsDir); err != nil {
		t.Fatalf("goose down: %v", err)
	}

	if err := goose.UpContext(ctx, db, migrationsDir); err != nil {
		t.Fatalf("goose up second time: %v", err)
	}
}

func assertTableExists(ctx context.Context, t *testing.T, db *sql.DB, tableName string) {
	t.Helper()

	var count int
	err := db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = $1
    `, tableName).Scan(&count)
	if err != nil {
		t.Fatalf("check table %s: %v", tableName, err)
	}
	if count == 0 {
		t.Fatalf("expected table %s to exist", tableName)
	}
}

func assertExtensionExists(ctx context.Context, t *testing.T, db *sql.DB, extension string) {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM pg_extension
        WHERE extname = $1
    `, extension).Scan(&count); err != nil {
		t.Fatalf("check extension %s: %v", extension, err)
	}
	if count == 0 {
		t.Fatalf("expected extension %s to exist", extension)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func buildTestDSNs(t *testing.T, base string) (admin, test, testDB string) {
	t.Helper()

	parsed, err := url.Parse(base)
	if err != nil {
		t.Fatalf("parse base dsn: %v", err)
	}

	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		dbName = "postgres"
	}

	testDB = fmt.Sprintf("%s_roundtrip_%d", sanitizeName(dbName), time.Now().UnixNano())

	adminURL := *parsed
	adminURL.Path = "/postgres"

	testURL := *parsed
	testURL.Path = "/" + testDB

	return adminURL.String(), testURL.String(), testDB
}

func sanitizeName(input string) string {
	cleaned := strings.ReplaceAll(input, "-", "_")
	cleaned = strings.ReplaceAll(cleaned, ".", "_")
	if cleaned == "" {
		cleaned = "migration_roundtrip"
	}
	return cleaned
}

func pqQuoteIdentifier(identifier string) string {
	return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
}

func lookupProjectPath(t *testing.T, elems ...string) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(append([]string{dir}, elems...)...)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate project root from %s", dir)
		}
		dir = parent
	}
}
