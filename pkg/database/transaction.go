package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Transaction 定义事务接口。
type Transaction interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
}

type txAdapter struct {
	tx *sql.Tx
}

// WrapSQLTx 将 *sql.Tx 包装为通用的 Transaction 接口，便于与仓储共享事务。
func WrapSQLTx(tx *sql.Tx) Transaction {
	if tx == nil {
		return nil
	}
	return &txAdapter{tx: tx}
}

func (a *txAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return a.tx.ExecContext(ctx, query, args...)
}

func (a *txAdapter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return a.tx.QueryContext(ctx, query, args...)
}

func (a *txAdapter) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return a.tx.QueryRowContext(ctx, query, args...)
}

func (a *txAdapter) Commit() error {
	return a.tx.Commit()
}

func (a *txAdapter) Rollback() error {
	return a.tx.Rollback()
}

// TxFunc 定义事务操作回调。
type TxFunc func(ctx context.Context, tx Transaction) error

// WithTx 在事务中执行回调，自动提交或回滚。
func (d *Database) WithTx(ctx context.Context, fn TxFunc) error {
	if d == nil || d.db == nil {
		return ErrDatabaseNotInitialized
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(ctx, &txAdapter{tx: tx})
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("operation failed: %w, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
