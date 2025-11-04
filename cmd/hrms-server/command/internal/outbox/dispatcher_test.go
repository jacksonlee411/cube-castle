package outbox

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	events            []*database.OutboxEvent
	markPublishedCall int32
	retryCalls        int32
}

func (r *fakeRepo) Save(ctx context.Context, tx database.Transaction, event *database.OutboxEvent) error {
	return nil
}

func (r *fakeRepo) GetUnpublishedForUpdate(ctx context.Context, tx database.Transaction, limit int) ([]*database.OutboxEvent, error) {
	return r.events, nil
}

func (r *fakeRepo) MarkPublished(ctx context.Context, eventID string) error {
	atomic.AddInt32(&r.markPublishedCall, 1)
	return nil
}

func (r *fakeRepo) IncrementRetryCount(ctx context.Context, eventID string, nextAvailable time.Time) error {
	atomic.AddInt32(&r.retryCalls, 1)
	return nil
}

type fakeBus struct {
	fail bool
}

func (b *fakeBus) Publish(ctx context.Context, event eventbus.Event) error {
	if b.fail {
		return errors.New("fail")
	}
	return nil
}

func (b *fakeBus) Subscribe(string, eventbus.EventHandler) error { return nil }

func TestDispatcherSuccess(t *testing.T) {
	cfg := Config{PollInterval: 10 * time.Millisecond, BatchSize: 1, BackoffBase: time.Second, MaxRetry: 3, MetricNamespace: "test"}
	repo := &fakeRepo{
		events: []*database.OutboxEvent{{
			EventID:       "evt-1",
			AggregateID:   "agg",
			AggregateType: "type",
			EventType:     "evt.created",
			Payload:       "{}",
		}},
	}
	bus := &fakeBus{}

	d := NewDispatcher(cfg, repo, bus, pkglogger.NewNoopLogger(), prometheus.NewRegistry(), func(ctx context.Context, fn database.TxFunc) error {
		return fn(ctx, nil)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, d.Start(ctx))
	time.Sleep(15 * time.Millisecond)
	require.NoError(t, d.Stop())
	require.Equal(t, int32(1), repo.markPublishedCall)
}

func TestDispatcherRetry(t *testing.T) {
	cfg := Config{PollInterval: 10 * time.Millisecond, BatchSize: 1, BackoffBase: time.Second, MaxRetry: 3, MetricNamespace: "test"}
	repo := &fakeRepo{
		events: []*database.OutboxEvent{{
			EventID:       "evt-1",
			AggregateID:   "agg",
			AggregateType: "type",
			EventType:     "evt.created",
			Payload:       "{}",
		}},
	}
	bus := &fakeBus{fail: true}

	d := NewDispatcher(cfg, repo, bus, pkglogger.NewNoopLogger(), prometheus.NewRegistry(), func(ctx context.Context, fn database.TxFunc) error {
		return fn(ctx, nil)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, d.Start(ctx))
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, d.Stop())
	require.Equal(t, int32(0), repo.markPublishedCall)
	require.Greater(t, repo.retryCalls, int32(0))
}
