//go:build integration

package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

// TestDispatcherIntegration验证dispatcher在真实Docker PostgreSQL环境下的端到端行为。
func TestDispatcherIntegration(t *testing.T) {
	// 连接测试数据库
	dbURL := "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "failed to connect to test database")
	defer db.Close()

	// 验证连接
	err = db.Ping()
	require.NoError(t, err, "failed to ping test database")

	// 清理outbox_events表以准备测试
	_, err = db.Exec("DELETE FROM outbox_events")
	require.NoError(t, err, "failed to clean up outbox_events table")

	t.Run("Success Path: Publish and Mark", func(t *testing.T) {
		// 预置一条待发布事件
		eventID := uuid.New()
		eventPayload, _ := json.Marshal(map[string]string{"action": "created"})
		_, err := db.Exec(
			`INSERT INTO outbox_events (event_id, aggregate_id, aggregate_type, event_type, payload, available_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			eventID, "agg-1", "Organization", "organization.created", eventPayload,
		)
		require.NoError(t, err, "failed to insert test event")

		// 创建测试总线和dispatcher
		logger := pkglogger.NewNoopLogger()
		bus := eventbus.NewMemoryEventBus(logger, nil)
		reg := prometheus.NewRegistry()

		dbClient, err := database.NewDatabaseWithConfig(database.ConnectionConfig{
			DSN:         dbURL,
			ServiceName: "test-dispatcher",
		})
		require.NoError(t, err, "failed to create database client")
		defer dbClient.Close()

		repo := database.NewOutboxRepository(dbClient)
		cfg := Config{
			PollInterval:    50 * time.Millisecond,
			BatchSize:       10,
			MaxRetry:        3,
			BackoffBase:     time.Second,
			MetricNamespace: "test_outbox",
		}

        dispatcher := NewDispatcher(cfg, repo, bus, logger, reg, dbClient.WithTx, nil)

		// 启动dispatcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = dispatcher.Start(ctx)
		require.NoError(t, err, "failed to start dispatcher")

		// 等待处理
		time.Sleep(200 * time.Millisecond)

		// 停止dispatcher
		err = dispatcher.Stop()
		require.NoError(t, err, "failed to stop dispatcher")

		// 验证事件被标记为已发布
		var published bool
		err = db.QueryRow("SELECT published FROM outbox_events WHERE event_id = $1", eventID).Scan(&published)
		require.NoError(t, err, "failed to query event status")
		require.True(t, published, "event should be marked as published")

		// 验证发布时间已设置
		var publishedAt sql.NullTime
		err = db.QueryRow("SELECT published_at FROM outbox_events WHERE event_id = $1", eventID).Scan(&publishedAt)
		require.NoError(t, err, "failed to query published_at")
		require.True(t, publishedAt.Valid, "published_at should be set")
	})

	t.Run("Failure Path: Retry and Backoff", func(t *testing.T) {
		// 清空表
		_, err := db.Exec("DELETE FROM outbox_events")
		require.NoError(t, err)

		// 预置一条事件
		eventID := uuid.New()
		eventPayload, _ := json.Marshal(map[string]string{"action": "failed"})
		_, err = db.Exec(
			`INSERT INTO outbox_events (event_id, aggregate_id, aggregate_type, event_type, payload, available_at)
			 VALUES ($1, $2, $3, $4, $5, NOW())`,
			eventID, "agg-2", "Organization", "organization.deleted", eventPayload,
		)
		require.NoError(t, err)

		// 创建会失败的总线
		logger := pkglogger.NewNoopLogger()
		failBus := &testFailingBus{fail: true}
		reg := prometheus.NewRegistry()

		dbClient, err := database.NewDatabaseWithConfig(database.ConnectionConfig{
			DSN:         dbURL,
			ServiceName: "test-dispatcher-fail",
		})
		require.NoError(t, err)
		defer dbClient.Close()

		repo := database.NewOutboxRepository(dbClient)
		cfg := Config{
			PollInterval:    50 * time.Millisecond,
			BatchSize:       10,
			MaxRetry:        3,
			BackoffBase:     100 * time.Millisecond,
			MetricNamespace: "test_outbox_fail",
		}

        dispatcher := NewDispatcher(cfg, repo, failBus, logger, reg, dbClient.WithTx, nil)

		// 启动dispatcher
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = dispatcher.Start(ctx)
		require.NoError(t, err)

		// 等待至少一次失败重试
		time.Sleep(400 * time.Millisecond)

		err = dispatcher.Stop()
		require.NoError(t, err)

		// 验证事件仍未发布
		var published bool
		err = db.QueryRow("SELECT published FROM outbox_events WHERE event_id = $1", eventID).Scan(&published)
		require.NoError(t, err)
		require.False(t, published, "event should not be published after failures")

		// 验证重试计数已增加
		var retryCount int
		err = db.QueryRow("SELECT retry_count FROM outbox_events WHERE event_id = $1", eventID).Scan(&retryCount)
		require.NoError(t, err)
		require.Greater(t, retryCount, 0, "retry_count should be incremented")

		// 验证available_at已更新（退避）
		var availableAt time.Time
		err = db.QueryRow("SELECT available_at FROM outbox_events WHERE event_id = $1", eventID).Scan(&availableAt)
		require.NoError(t, err)
		require.True(t, availableAt.After(time.Now()), "available_at should be in the future due to backoff")
	})

	t.Run("Graceful Shutdown with Context", func(t *testing.T) {
		// 清空表
		_, err := db.Exec("DELETE FROM outbox_events")
		require.NoError(t, err)

		logger := pkglogger.NewNoopLogger()
		bus := eventbus.NewMemoryEventBus(logger, nil)
		reg := prometheus.NewRegistry()

		dbClient, err := database.NewDatabaseWithConfig(database.ConnectionConfig{
			DSN:         dbURL,
			ServiceName: "test-dispatcher-shutdown",
		})
		require.NoError(t, err)
		defer dbClient.Close()

		repo := database.NewOutboxRepository(dbClient)
		cfg := Config{
			PollInterval:    100 * time.Millisecond,
			BatchSize:       10,
			MaxRetry:        3,
			BackoffBase:     time.Second,
			MetricNamespace: "test_outbox_shutdown",
		}

        dispatcher := NewDispatcher(cfg, repo, bus, logger, reg, dbClient.WithTx, nil)

		// 启动dispatcher
		ctx, cancel := context.WithCancel(context.Background())

		err = dispatcher.Start(ctx)
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		// 取消上下文触发关闭
		cancel()
		time.Sleep(200 * time.Millisecond)

		// 停止dispatcher应该成功
		err = dispatcher.Stop()
		require.NoError(t, err, "dispatcher should stop gracefully after context cancellation")
	})

	t.Run("Idempotency: Skip Already Published", func(t *testing.T) {
		// 清空表
		_, err := db.Exec("DELETE FROM outbox_events")
		require.NoError(t, err)

		// 预置已发布的事件
		eventID := uuid.New()
		now := time.Now()
		eventPayload, _ := json.Marshal(map[string]string{"action": "idempotent"})
		_, err = db.Exec(
			`INSERT INTO outbox_events (event_id, aggregate_id, aggregate_type, event_type, payload, published, published_at)
			 VALUES ($1, $2, $3, $4, $5, TRUE, $6)`,
			eventID, "agg-3", "Organization", "organization.updated", eventPayload, now,
		)
		require.NoError(t, err)

		logger := pkglogger.NewNoopLogger()
		bus := eventbus.NewMemoryEventBus(logger, nil)
		reg := prometheus.NewRegistry()

		dbClient, err := database.NewDatabaseWithConfig(database.ConnectionConfig{
			DSN:         dbURL,
			ServiceName: "test-dispatcher-idempotent",
		})
		require.NoError(t, err)
		defer dbClient.Close()

		repo := database.NewOutboxRepository(dbClient)
		cfg := Config{
			PollInterval:    50 * time.Millisecond,
			BatchSize:       10,
			MaxRetry:        3,
			BackoffBase:     time.Second,
			MetricNamespace: "test_outbox_idempotent",
		}

        dispatcher := NewDispatcher(cfg, repo, bus, logger, reg, dbClient.WithTx, nil)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = dispatcher.Start(ctx)
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)

		err = dispatcher.Stop()
		require.NoError(t, err)

		// 验证已发布的事件未被重复处理
		var publishedAt sql.NullTime
		err = db.QueryRow("SELECT published_at FROM outbox_events WHERE event_id = $1", eventID).Scan(&publishedAt)
		require.NoError(t, err)
		// published_at应该保持原值
		require.True(t, publishedAt.Valid)
	})
}

// testFailingBus 是一个总是发布失败的测试总线实现
type testFailingBus struct {
	fail bool
}

func (b *testFailingBus) Publish(ctx context.Context, event eventbus.Event) error {
	if b.fail {
		return fmt.Errorf("test bus failure")
	}
	return nil
}

func (b *testFailingBus) Subscribe(string, eventbus.EventHandler) error {
	return nil
}
