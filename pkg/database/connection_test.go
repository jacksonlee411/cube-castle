package database

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
)

func newMockDatabase(t *testing.T, cfg ConnectionConfig) (*Database, sqlmock.Sqlmock, func()) {
	t.Helper()

	originalOpen := openDB
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.MatchExpectationsInOrder(false)

	openDB = func(_ string, _ string) (*sql.DB, error) {
		return db, nil
	}

	mock.ExpectPing()

	database, err := NewDatabaseWithConfig(cfg)
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())
		openDB = originalOpen
	}

	return database, mock, cleanup
}

func TestNewDatabaseAppliesDefaults(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/defaults"}
	db, _, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	config := db.Config()
	require.Equal(t, DefaultMaxOpenConns, config.MaxOpenConns)
	require.Equal(t, DefaultMaxIdleConns, config.MaxIdleConns)
	require.Equal(t, DefaultConnMaxIdleTime, config.ConnMaxIdleTime)
	require.Equal(t, DefaultConnMaxLifetime, config.ConnMaxLifetime)

	stats := db.GetStats()
	require.Equal(t, DefaultMaxOpenConns, stats.MaxOpenConnections)
}

func TestNewDatabaseWithCustomConfig(t *testing.T) {
	cfg := ConnectionConfig{
		DSN:             "postgres://test/custom",
		MaxOpenConns:    10,
		MaxIdleConns:    3,
		ConnMaxIdleTime: 2 * time.Minute,
		ConnMaxLifetime: time.Hour,
		ServiceName:     "hrms-command",
	}

	db, _, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	config := db.Config()
	require.Equal(t, cfg.MaxOpenConns, config.MaxOpenConns)
	require.Equal(t, cfg.MaxIdleConns, config.MaxIdleConns)
	require.Equal(t, cfg.ConnMaxIdleTime, config.ConnMaxIdleTime)
	require.Equal(t, cfg.ConnMaxLifetime, config.ConnMaxLifetime)
	require.Equal(t, cfg.ServiceName, config.ServiceName)
}

func TestNewDatabaseWithEmptyDSN(t *testing.T) {
	_, err := NewDatabaseWithConfig(ConnectionConfig{})
	require.ErrorIs(t, err, ErrEmptyDSN)
}

func TestNewDatabaseShortcut(t *testing.T) {
	original := openDB
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	t.Cleanup(func() {
		openDB = original
		db.Close()
	})

	openDB = func(_ string, dsn string) (*sql.DB, error) {
		require.Equal(t, "postgres://test/newdatabase", dsn)
		return db, nil
	}

	mock.ExpectPing()
	mock.ExpectClose()

	realDB, err := NewDatabase("postgres://test/newdatabase")
	require.NoError(t, err)
	require.NotNil(t, realDB.GetDB())
	require.NoError(t, realDB.Close())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExecAndQueryWrappers(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/wrapper"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	mock.ExpectExec("UPDATE employees SET name").
		WithArgs("Alice", int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	result, err := db.ExecContext(ctx, "UPDATE employees SET name = $1 WHERE id = $2", "Alice", int64(1))
	require.NoError(t, err)
	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)

	rowsMock := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice")
	mock.ExpectQuery("SELECT id, name FROM employees").
		WithArgs(int64(1)).
		WillReturnRows(rowsMock)

	resultRows, err := db.QueryContext(ctx, "SELECT id, name FROM employees WHERE id = $1", int64(1))
	require.NoError(t, err)
	require.NoError(t, resultRows.Close())

	mock.ExpectQuery("SELECT name FROM employees").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Alice"))
	row := db.QueryRowContext(ctx, "SELECT name FROM employees WHERE id = $1", int64(1))
	var name string
	require.NoError(t, row.Scan(&name))
	require.Equal(t, "Alice", name)

	require.NotNil(t, db.GetDB())
	config := db.Config()
	require.Equal(t, cfg.DSN, config.DSN)
	require.Equal(t, DefaultMaxOpenConns, config.MaxOpenConns)
	require.Equal(t, DefaultMaxIdleConns, config.MaxIdleConns)
	require.NotZero(t, config.ConnMaxIdleTime)
	require.NotZero(t, config.ConnMaxLifetime)
	require.Equal(t, cfg.normalize().ConnMaxIdleTime, config.ConnMaxIdleTime)
	require.Equal(t, cfg.normalize().ConnMaxLifetime, config.ConnMaxLifetime)
}

func TestRecordConnectionStats(t *testing.T) {
	cfg := ConnectionConfig{
		DSN:         "postgres://test/metrics",
		ServiceName: "query-service",
	}
	db, _, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	metricsOnce = sync.Once{}
	dbConnectionsInUse.Reset()
	dbConnectionsIdle.Reset()
	dbQueryDuration.Reset()

	registry := prometheus.NewRegistry()
	RegisterMetrics(registry)

	db.RecordConnectionStats("")
	stats := db.GetStats()

	inUse := testutil.ToFloat64(dbConnectionsInUse.WithLabelValues("query-service"))
	idle := testutil.ToFloat64(dbConnectionsIdle.WithLabelValues("query-service"))

	require.Equal(t, float64(stats.InUse), inUse)
	require.Equal(t, float64(stats.Idle), idle)
}

func TestRecordQueryDuration(t *testing.T) {
	cfg := ConnectionConfig{
		DSN:         "postgres://test/query-metrics",
		ServiceName: "command-service",
	}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	metricsOnce = sync.Once{}
	dbQueryDuration.Reset()
	registry := prometheus.NewRegistry()
	RegisterMetrics(registry)

	mock.ExpectQuery("SELECT 1").
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))

	_, err := db.QueryContext(context.Background(), "SELECT 1")
	require.NoError(t, err)

	metricCount := testutil.CollectAndCount(dbQueryDuration)
	require.GreaterOrEqual(t, metricCount, 1)

	mock.ExpectQuery("SELECT NOW\\(\\)").
		WillReturnRows(sqlmock.NewRows([]string{"now"}).AddRow(time.Now()))

	row := db.QueryRowContext(context.Background(), "SELECT NOW()")
	require.NotNil(t, row)
}

func TestConnectionConfigNormalize(t *testing.T) {
	defaultCfg := ConnectionConfig{DSN: "postgres://test/normalize"}.normalize()
	require.Equal(t, DefaultMaxOpenConns, defaultCfg.MaxOpenConns)
	require.Equal(t, DefaultMaxIdleConns, defaultCfg.MaxIdleConns)
	require.Equal(t, DefaultConnMaxIdleTime, defaultCfg.ConnMaxIdleTime)
	require.Equal(t, DefaultConnMaxLifetime, defaultCfg.ConnMaxLifetime)

	custom := ConnectionConfig{
		DSN:             "postgres://test/custom",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxIdleTime: time.Minute,
		ConnMaxLifetime: 10 * time.Minute,
	}.normalize()
	require.Equal(t, 50, custom.MaxOpenConns)
	require.Equal(t, 10, custom.MaxIdleConns)
	require.Equal(t, time.Minute, custom.ConnMaxIdleTime)
	require.Equal(t, 10*time.Minute, custom.ConnMaxLifetime)
}

func TestDatabaseNilReceivers(t *testing.T) {
	var db *Database

	_, err := db.ExecContext(context.Background(), "SELECT 1")
	require.ErrorIs(t, err, ErrDatabaseNotInitialized)

	_, err = db.QueryContext(context.Background(), "SELECT 1")
	require.ErrorIs(t, err, ErrDatabaseNotInitialized)

	require.Nil(t, db.QueryRowContext(context.Background(), "SELECT 1"))
	require.Nil(t, db.GetDB())
	require.Equal(t, ConnectionConfig{}, db.Config())
	require.Equal(t, sql.DBStats{}, db.GetStats())
	require.NoError(t, db.Close())
}
