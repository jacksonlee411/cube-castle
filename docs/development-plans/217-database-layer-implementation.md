# Plan 217 - `pkg/database/` 数据库访问层实现

**文档编号**: 217
**标题**: 数据库访问层与事务管理 - 统一实现
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 216（eventbus）、Plan 210（迁移脚本）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

实现统一的数据库访问层（pkg/database），为所有模块提供：
- 连接池管理
- 事务支持
- 事务性发件箱（Transactional Outbox）接口

**关键成果**:
- ✅ 数据库连接管理（连接池配置）
- ✅ 事务管理（Transaction 包装）
- ✅ 事务性发件箱表接口
- ✅ 单元 & 集成测试（覆盖率 > 80%）
- ✅ Prometheus 指标暴露

### 1.2 为什么需要统一的数据库访问层

根据 200 号文档的分析，大型 ERP 系统必须：
1. **显式配置连接池** - 防止"too many connections"错误
2. **事务性发件箱模式** - 保证跨模块操作的最终一致性
3. **集中化管理** - 降低每个模块的复杂性

**关键问题**:
- 当前 organization 模块直接操作数据库，每个模块需要重复实现
- 缺乏统一的连接池配置，导致潜在的连接数溢出
- 没有事务性发件箱支持，跨模块操作缺乏可靠性保证

### 1.3 时间计划

- **计划完成**: Week 3 Day 2 (Day 13)
- **交付周期**: 1.5 天
- **负责人**: 基础设施 + 后端 TL
- **前置依赖**: Plan 210（迁移脚本）完成

---

## 2. 需求分析

### 2.1 功能需求

#### 需求 1: 数据库连接管理

所有服务必须使用统一的连接池配置：

```go
type ConnectionConfig struct {
    DSN                string        // 数据库连接字符串
    MaxOpenConns       int           // 最大连接数（默认 25）
    MaxIdleConns       int           // 最大空闲连接（默认 5）
    ConnMaxIdleTime    time.Duration // 连接空闲超时（默认 5 分钟）
    ConnMaxLifetime    time.Duration // 连接生命周期（默认 30 分钟）
}
```

**理由**:
- MaxOpenConns = 25：PostgreSQL 服务器默认 100 连接限制，避免单个应用耗尽
- MaxIdleConns = 5：平衡连接复用和资源释放
- ConnMaxIdleTime = 5m：定期刷新连接，防止网络连接泄漏
- ConnMaxLifetime = 30m：周期性替换连接，防止长期占用

#### 需求 2: 事务支持

提供事务管理的统一接口：

```go
// Transaction 包装了 *sql.Tx，提供统一的事务接口
type Transaction interface {
    Exec(query string, args ...interface{}) (sql.Result, error)
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    Commit() error
    Rollback() error
}

// TxFunc 定义事务操作的回调函数
type TxFunc func(tx Transaction) error

// WithTx 在事务内执行操作，自动提交或回滚
func (db *Database) WithTx(ctx context.Context, fn TxFunc) error
```

#### 需求 3: 事务性发件箱

支持事务性发件箱（Transactional Outbox）模式：

```go
type OutboxEvent struct {
    ID             int64
    EventID        string    // UUID，幂等 ID
    AggregateID    string    // 业务对象 ID（如 employeeID）
    AggregateType  string    // 业务对象类型（如 "employee"）
    EventType      string    // 事件类型（如 "employee.created"）
    Payload        string    // JSON 事件数据
    RetryCount     int       // 重试次数
    Published      bool      // 是否已发布
    PublishedAt    *time.Time
    CreatedAt      time.Time
}

// OutboxRepository 管理 outbox 表
type OutboxRepository interface {
    SaveEvent(ctx context.Context, tx Transaction, event *OutboxEvent) error
    GetUnpublished(ctx context.Context, limit int) ([]*OutboxEvent, error)
    MarkPublished(ctx context.Context, eventID string) error
}
```

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **性能** | < 10ms 获取连接 | P95 延迟控制 |
| **可靠性** | 100% 连接回收 | 无连接泄漏 |
| **可观测性** | Prometheus 指标 | 连接数、延迟、错误计数 |
| **测试覆盖率** | > 80% | 单元 + 集成测试 |
| **向后兼容** | ✅ 需要 | 不破坏现有代码 |

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/database/
├── connection.go       # 连接池管理
├── transaction.go      # 事务支持
├── outbox.go           # 事务性发件箱
├── metrics.go          # Prometheus 指标
├── connection_test.go  # 连接池测试
├── transaction_test.go # 事务测试
├── outbox_test.go      # 发件箱测试
└── README.md           # 使用说明
```

### 3.2 关键设计决策

#### 决策 1: 连接池参数为什么硬编码

虽然参数可以配置，但生产环境中应该使用统一的推荐值：

```go
// 统一的连接池标准配置
const (
    DefaultMaxOpenConns    = 25          // 最大连接数
    DefaultMaxIdleConns    = 5           // 最大空闲连接
    DefaultConnMaxIdleTime = 5 * time.Minute    // 空闲超时
    DefaultConnMaxLifetime = 30 * time.Minute   // 连接生命周期
)
```

**理由**:
- 参数经验证，适合大多数场景
- 避免配置错误导致的性能问题
- 易于监控和预测

#### 决策 2: 为什么选择事务性发件箱而非消息队列

| 方案 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| **事务性发件箱** | 实现简单、依赖少、成本低 | 需要后台中继 | ✅ Phase2 |
| **消息队列** | 高可用、成熟方案 | 运维成本高 | ⏳ Phase3+ |

**Phase2 的选择理由**:
- 与事务保证原子性
- 不需要额外的中间件
- 为未来迁移预留接口

---

## 4. 详细实现

### 4.1 connection.go - 连接管理

```go
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	// 连接池标准配置
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 5
	DefaultConnMaxIdleTime = 5 * time.Minute
	DefaultConnMaxLifetime = 30 * time.Minute
)

// ConnectionConfig 定义数据库连接配置
type ConnectionConfig struct {
	DSN                string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxIdleTime    time.Duration
	ConnMaxLifetime    time.Duration
}

// Database 包装 *sql.DB，提供统一的数据库访问接口
type Database struct {
	db     *sql.DB
	config ConnectionConfig
}

// NewDatabase 创建新的数据库连接
// 如果配置中的参数为 0，使用默认值
func NewDatabase(dsn string) (*Database, error) {
	return NewDatabaseWithConfig(ConnectionConfig{
		DSN:                dsn,
		MaxOpenConns:       DefaultMaxOpenConns,
		MaxIdleConns:       DefaultMaxIdleConns,
		ConnMaxIdleTime:    DefaultConnMaxIdleTime,
		ConnMaxLifetime:    DefaultConnMaxLifetime,
	})
}

// NewDatabaseWithConfig 使用完整配置创建数据库连接
func NewDatabaseWithConfig(config ConnectionConfig) (*Database, error) {
	if config.DSN == "" {
		return nil, ErrEmptyDSN
	}

	// 使用默认值填充零值
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

	db, err := sql.Open("postgres", config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		config: config,
	}

	// 记录连接池配置
	fmt.Printf("Database connected with config: MaxOpenConns=%d, MaxIdleConns=%d\n",
		config.MaxOpenConns, config.MaxIdleConns)

	return database, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// GetDB 返回底层 *sql.DB（用于兼容现有代码）
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// GetStats 返回连接池统计信息
func (d *Database) GetStats() sql.DBStats {
	return d.db.Stats()
}
```

### 4.2 transaction.go - 事务管理

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Transaction 定义事务接口
type Transaction interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Commit() error
	Rollback() error
}

// TxFunc 定义事务操作的回调函数
type TxFunc func(ctx context.Context, tx *sql.Tx) error

// WithTx 在事务内执行操作，自动提交或回滚
func (d *Database) WithTx(ctx context.Context, fn TxFunc) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = fn(ctx, tx)
	if err != nil {
		// 回滚事务
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("operation failed: %w, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecContext 执行查询（兼容现有代码）
func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// QueryContext 执行查询（兼容现有代码）
func (d *Database) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRowContext 执行单行查询（兼容现有代码）
func (d *Database) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}
```

### 4.3 outbox.go - 事务性发件箱

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// OutboxEvent 表示 outbox 表的一行记录
type OutboxEvent struct {
	ID            int64      // 主键
	EventID       string     // UUID，幂等 ID（由应用生成）
	AggregateID   string     // 业务对象 ID
	AggregateType string     // 业务对象类型
	EventType     string     // 事件类型
	Payload       string     // JSON 格式的事件数据
	RetryCount    int        // 重试次数
	Published     bool       // 是否已发布
	PublishedAt   *time.Time // 发布时间
	CreatedAt     time.Time  // 创建时间
}

// SaveOutboxEvent 在事务内保存 outbox 事件
// 与业务操作一起在同一事务内执行，保证原子性
func SaveOutboxEvent(ctx context.Context, tx *sql.Tx, event *OutboxEvent) error {
	if event == nil {
		return ErrNilOutboxEvent
	}

	if event.EventID == "" {
		return ErrEmptyEventID
	}

	query := `
		INSERT INTO outbox_events
		(event_id, aggregate_id, aggregate_type, event_type, payload, retry_count, published, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := tx.QueryRowContext(ctx, query,
		event.EventID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		0,           // 初始重试次数为 0
		false,       // 初始发布状态为 false
		time.Now(),  // 创建时间
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}

	return nil
}

// GetUnpublishedEvents 获取未发布的事件（用于后台中继）
func (d *Database) GetUnpublishedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, event_id, aggregate_id, aggregate_type, event_type, payload,
		       retry_count, published, published_at, created_at
		FROM outbox_events
		WHERE published = FALSE
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpublished events: %w", err)
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		event := &OutboxEvent{}
		err := rows.Scan(
			&event.ID,
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Payload,
			&event.RetryCount,
			&event.Published,
			&event.PublishedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// MarkEventPublished 标记事件为已发布
func (d *Database) MarkEventPublished(ctx context.Context, eventID string) error {
	if eventID == "" {
		return ErrEmptyEventID
	}

	query := `
		UPDATE outbox_events
		SET published = TRUE, published_at = NOW()
		WHERE event_id = $1
	`

	result, err := d.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event published: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}

// IncrementRetryCount 增加事件的重试次数
func (d *Database) IncrementRetryCount(ctx context.Context, eventID string) error {
	if eventID == "" {
		return ErrEmptyEventID
	}

	query := `
		UPDATE outbox_events
		SET retry_count = retry_count + 1
		WHERE event_id = $1
	`

	_, err := d.db.ExecContext(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	return nil
}
```

### 4.4 error.go - 错误定义

```go
package database

import "errors"

var (
	ErrEmptyDSN       = errors.New("DSN cannot be empty")
	ErrNilOutboxEvent = errors.New("outbox event cannot be nil")
	ErrEmptyEventID   = errors.New("event ID cannot be empty")
	ErrEventNotFound  = errors.New("event not found")
)
```

---

## 5. 单元与集成测试

### 5.1 测试场景覆盖

```go
// connection_test.go 测试场景

// Test 1: 连接池正常创建
TestNewDatabase()

// Test 2: 连接参数验证
TestConnectionPoolSettings()

// Test 3: 数据库连接失败
TestConnectionFailure()

// Test 4: 连接统计信息
TestGetStats()

// transaction_test.go 测试场景

// Test 5: 事务提交成功
TestWithTxCommit()

// Test 6: 事务回滚
TestWithTxRollback()

// Test 7: 事务中的错误处理
TestWithTxError()

// outbox_test.go 测试场景

// Test 8: 保存 outbox 事件
TestSaveOutboxEvent()

// Test 9: 获取未发布的事件
TestGetUnpublishedEvents()

// Test 10: 标记事件为已发布
TestMarkEventPublished()

// Test 11: 增加重试次数
TestIncrementRetryCount()
```

### 5.2 集成测试：端到端流程

```go
// 场景：在事务内保存业务数据和 outbox 事件

func TestEndToEndTransactionWithOutbox(t *testing.T) {
    // 1. 开启事务
    // 2. 更新 employees 表（业务数据）
    // 3. 保存 outbox_events 表（事件）
    // 4. 提交事务
    // 5. 验证两个操作原子性
}
```

---

## 6. Prometheus 指标

### 6.1 metrics.go - 指标暴露

```go
package database

import "github.com/prometheus/client_golang/prometheus"

var (
	// 当前连接池中的活动连接数
	dbConnectionsInUse = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use",
		},
		[]string{"service"},
	)

	// 当前连接池中的空闲连接数
	dbConnectionsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
		[]string{"service"},
	)

	// 查询延迟直方图
	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"service", "query_type"},
	)
)

// RecordConnectionStats 定期记录连接池统计
func (d *Database) RecordConnectionStats(serviceName string) {
	stats := d.db.Stats()
	dbConnectionsInUse.WithLabelValues(serviceName).Set(float64(stats.InUse))
	dbConnectionsIdle.WithLabelValues(serviceName).Set(float64(stats.Idle))
}
```

---

## 7. 验收标准

### 7.1 功能验收

- [ ] 连接池配置正确（MaxOpenConns=25, MaxIdleConns=5）
- [ ] 连接验证通过（Ping）
- [ ] 事务创建和提交正常
- [ ] 事务回滚正常
- [ ] Outbox 事件保存成功
- [ ] Outbox 事件查询成功
- [ ] 发布标记和重试计数更新正常

### 7.2 质量验收

- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试全部通过
- [ ] 代码通过 `go fmt` 检查
- [ ] 代码通过 `go vet` 检查
- [ ] 无 race condition（`go test -race`）

### 7.3 集成验收

- [ ] 可与 Plan 216 (eventbus) 配合使用
- [ ] 可与 Plan 218 (logger) 集成
- [ ] 可在 Plan 219 (organization 重构) 中使用
- [ ] 与现有 organization 模块兼容

---

## 8. 迁移指南

### 8.1 现有代码迁移

**旧方式** (直接操作 sql.DB)：
```go
db, _ := sql.Open("postgres", dsn)
db.SetMaxOpenConns(10)
rows, _ := db.Query("SELECT * FROM organizations")
```

**新方式** (使用 pkg/database)：
```go
database, _ := database.NewDatabase(dsn)
// 连接池已自动配置为标准参数
rows, _ := database.GetDB().Query("SELECT * FROM organizations")
// 或直接使用
rows, _ := database.QueryContext(ctx, "SELECT * FROM organizations")
```

### 8.2 事务使用

**旧方式**：
```go
tx, _ := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
// 业务逻辑
tx.Commit()
```

**新方式**：
```go
database.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
    // 业务逻辑
    return nil // 自动提交，若返回 error 则自动回滚
})
```

---

## 9. 风险与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|--------|
| 连接泄漏 | 中 | 高 | 压力测试，监控连接数 |
| 事务死锁 | 低 | 高 | 充分的事务测试 |
| 性能退化 | 低 | 中 | 基准测试，查询优化 |
| 与现有代码冲突 | 中 | 中 | 充分的集成测试，迁移逐步进行 |

---

## 10. 交付物清单

- ✅ `pkg/database/connection.go`
- ✅ `pkg/database/transaction.go`
- ✅ `pkg/database/outbox.go`
- ✅ `pkg/database/error.go`
- ✅ `pkg/database/metrics.go`
- ✅ `pkg/database/connection_test.go`
- ✅ `pkg/database/transaction_test.go`
- ✅ `pkg/database/outbox_test.go`
- ✅ `pkg/database/README.md`
- ✅ 本计划文档（217）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**计划完成日期**: Week 3 Day 2 (Day 13)
