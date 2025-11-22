package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestWithTxCommit(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/tx-commit"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO operations").
		WithArgs("ok").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := db.WithTx(context.Background(), func(ctx context.Context, tx Transaction) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO operations(name) VALUES ($1)", "ok")
		return err
	})
	require.NoError(t, err)
}

func TestWithTxRollbackOnError(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/tx-rollback"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	mock.ExpectBegin()
	mock.ExpectRollback()

	expectedErr := errors.New("boom")
	err := db.WithTx(context.Background(), func(_ context.Context, _ Transaction) error {
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)
}

func TestWithTxRollbackFailureReported(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/tx-rollback-failure"}

	originalOpen := openDB
	rawDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.MatchExpectationsInOrder(false)

	openDB = func(_ string, _ string) (*sql.DB, error) {
		return rawDB, nil
	}
	mock.ExpectPing()

	t.Cleanup(func() {
		openDB = originalOpen
	})

	database, err := NewDatabaseWithConfig(cfg)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))

	err = database.WithTx(context.Background(), func(_ context.Context, _ Transaction) error {
		return errors.New("failure")
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "rollback failed")

	mock.ExpectClose()
	require.NoError(t, database.Close())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithTxNilDatabase(t *testing.T) {
	var db *Database
	err := db.WithTx(context.Background(), func(_ context.Context, _ Transaction) error {
		return nil
	})
	require.ErrorIs(t, err, ErrDatabaseNotInitialized)
}
