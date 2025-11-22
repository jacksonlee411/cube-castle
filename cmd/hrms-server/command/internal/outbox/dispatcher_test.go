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
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	events            []*database.OutboxEvent
	markPublishedCall int32
	retryCalls        int32
}

func (r *fakeRepo) Save(_ context.Context, _ database.Transaction, _ *database.OutboxEvent) error {
	return nil
}

func (r *fakeRepo) GetUnpublishedForUpdate(_ context.Context, _ database.Transaction, _ int) ([]*database.OutboxEvent, error) {
	return r.events, nil
}

func (r *fakeRepo) MarkPublished(_ context.Context, _ string) error {
	atomic.AddInt32(&r.markPublishedCall, 1)
	return nil
}

func (r *fakeRepo) IncrementRetryCount(_ context.Context, _ string, _ time.Time) error {
	atomic.AddInt32(&r.retryCalls, 1)
	return nil
}

type fakeBus struct {
	fail bool
}

func (b *fakeBus) Publish(_ context.Context, _ eventbus.Event) error {
	if b.fail {
		return errors.New("fail")
	}
	return nil
}

func (b *fakeBus) Subscribe(string, eventbus.EventHandler) error { return nil }

type fakeCache struct {
	calls        int32
	lastTenant   uuid.UUID
	lastPosition string
	fail         bool
}

func (f *fakeCache) RefreshPositionCache(_ context.Context, tenantID uuid.UUID, positionCode string) error {
	atomic.AddInt32(&f.calls, 1)
	f.lastTenant = tenantID
	f.lastPosition = positionCode
	if f.fail {
		return errors.New("refresh failed")
	}
	return nil
}

func TestDispatcherSuccess(t *testing.T) {
	cfg := Config{PollInterval: 10 * time.Millisecond, BatchSize: 1, BackoffBase: time.Second, MaxRetry: 3, MetricNamespace: "test"}
	repo := &fakeRepo{
		events: []*database.OutboxEvent{{
			EventID:       "evt-1",
			AggregateID:   "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
			AggregateType: "type",
			EventType:     eventAssignmentFilled,
			Payload:       `{"tenantId":"3b99930c-4dc6-4cc9-8e4d-7d960a931cb9","positionCode":"P1001"}`,
		}},
	}
	bus := &fakeBus{}
	cache := &fakeCache{}

	d := NewDispatcher(cfg, repo, bus, pkglogger.NewNoopLogger(), prometheus.NewRegistry(), func(ctx context.Context, fn database.TxFunc) error {
		return fn(ctx, nil)
	}, cache)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, d.Start(ctx))
	time.Sleep(15 * time.Millisecond)
	require.NoError(t, d.Stop())
	require.Equal(t, int32(1), repo.markPublishedCall)
	require.Equal(t, int32(1), cache.calls)
	require.Equal(t, "P1001", cache.lastPosition)
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
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, d.Start(ctx))
	time.Sleep(20 * time.Millisecond)
	require.NoError(t, d.Stop())
	require.Equal(t, int32(0), repo.markPublishedCall)
	require.Greater(t, repo.retryCalls, int32(0))
}
