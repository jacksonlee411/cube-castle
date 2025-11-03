package database

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestOutboxMigrationUpDown(t *testing.T) {
	path := filepath.Join("..", "..", "database", "migrations", "20251107090000_create_outbox_events.sql")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	sql := string(content)
	require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS public.outbox_events")
	require.Contains(t, sql, "DROP TABLE IF EXISTS public.outbox_events")

	upStatements := []string{
		`CREATE TABLE IF NOT EXISTS public.outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    aggregate_id TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_events_published_created_at
    ON public.outbox_events (published, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_events_available_at
    ON public.outbox_events (published, available_at, created_at);`,
	}

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	for _, stmt := range upStatements {
		re := regexp.QuoteMeta(stmt)
		mock.ExpectExec(re).
			WillReturnResult(sqlmock.NewResult(0, 0))
		_, execErr := db.Exec(stmt)
		require.NoError(t, execErr)
	}

	mock.ExpectExec(regexp.QuoteMeta("DROP TABLE IF EXISTS public.outbox_events")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, execErr := db.Exec("DROP TABLE IF EXISTS public.outbox_events")
	require.NoError(t, execErr)

	require.NoError(t, mock.ExpectationsWereMet())
}
