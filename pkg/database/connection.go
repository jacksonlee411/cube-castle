package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Register PostgreSQL driver.
	_ "github.com/lib/pq"
)

const (
	// DefaultMaxOpenConns 是推荐的最大连接数。
	DefaultMaxOpenConns = 25
	// DefaultMaxIdleConns 是推荐的最大空闲连接数。
	DefaultMaxIdleConns = 5
	// DefaultConnMaxIdleTime 控制连接的最大空闲时间。
	DefaultConnMaxIdleTime = 5 * time.Minute
	// DefaultConnMaxLifetime 控制连接的最大生命周期。
	DefaultConnMaxLifetime = 30 * time.Minute
)

// openDB 抽象出 sql.Open，方便测试时替换。
var openDB = sql.Open

// ConnectionConfig 定义数据库连接配置。
type ConnectionConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	ServiceName     string
}

func (c ConnectionConfig) normalize() ConnectionConfig {
	config := c
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = DefaultMaxOpenConns
	}
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = DefaultMaxIdleConns
	}
	if config.ConnMaxIdleTime <= 0 {
		config.ConnMaxIdleTime = DefaultConnMaxIdleTime
	}
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = DefaultConnMaxLifetime
	}
	return config
}

// Database 封装 *sql.DB 并提供统一的访问接口。
type Database struct {
	db     *sql.DB
	config ConnectionConfig
}

// NewDatabase 使用默认参数创建数据库连接。
func NewDatabase(dsn string) (*Database, error) {
	return NewDatabaseWithConfig(ConnectionConfig{
		DSN:             dsn,
		MaxOpenConns:    DefaultMaxOpenConns,
		MaxIdleConns:    DefaultMaxIdleConns,
		ConnMaxIdleTime: DefaultConnMaxIdleTime,
		ConnMaxLifetime: DefaultConnMaxLifetime,
	})
}

// NewDatabaseWithConfig 使用自定义配置创建数据库连接。
func NewDatabaseWithConfig(config ConnectionConfig) (*Database, error) {
	if config.DSN == "" {
		return nil, ErrEmptyDSN
	}

	config = config.normalize()

	db, err := openDB("postgres", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		db:     db,
		config: config,
	}, nil
}

// Close 关闭数据库连接。
func (d *Database) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

// GetDB 返回底层 *sql.DB。
func (d *Database) GetDB() *sql.DB {
	if d == nil {
		return nil
	}
	return d.db
}

// Config 返回连接配置。
func (d *Database) Config() ConnectionConfig {
	if d == nil {
		return ConnectionConfig{}
	}
	return d.config
}

// GetStats 返回连接池统计信息。
func (d *Database) GetStats() sql.DBStats {
	if d == nil || d.db == nil {
		return sql.DBStats{}
	}
	return d.db.Stats()
}

// ExecContext 执行写操作，主要用作兼容旧代码。
func (d *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if d == nil || d.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	start := time.Now()
	result, err := d.db.ExecContext(ctx, query, args...)
	recordQueryDuration(d.config, query, time.Since(start))
	return result, err
}

// QueryContext 执行查询操作，主要用作兼容旧代码。
func (d *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if d == nil || d.db == nil {
		return nil, ErrDatabaseNotInitialized
	}
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query, args...)
	recordQueryDuration(d.config, query, time.Since(start))
	return rows, err
}

// QueryRowContext 执行单行查询，主要用作兼容旧代码。
func (d *Database) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if d == nil || d.db == nil {
		return nil
	}
	start := time.Now()
	row := d.db.QueryRowContext(ctx, query, args...)
	recordQueryDuration(d.config, query, time.Since(start))
	return row
}
