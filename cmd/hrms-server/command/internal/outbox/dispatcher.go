package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

// Dispatcher 轮询 outbox 表并通过 eventbus 发布事件。
type Dispatcher struct {
	cfg     Config
	repo    database.OutboxRepository
	bus     eventbus.EventBus
	logger  pkglogger.Logger
	metrics *metrics
	withTx  func(ctx context.Context, fn database.TxFunc) error

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

const maxBackoff = 5 * time.Minute

func NewDispatcher(cfg Config, repo database.OutboxRepository, bus eventbus.EventBus, logger pkglogger.Logger, reg prometheus.Registerer, withTx func(ctx context.Context, fn database.TxFunc) error) *Dispatcher {
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}
	return &Dispatcher{
		cfg:     cfg,
		repo:    repo,
		bus:     bus,
		logger:  logger.WithFields(pkglogger.Fields{"component": "outbox-dispatcher"}),
		metrics: newMetrics(cfg.MetricNamespace, reg),
		withTx:  withTx,
	}
}

func (d *Dispatcher) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.cancel != nil {
		return ErrDispatcherAlreadyRunning
	}
	runCtx, cancel := context.WithCancel(ctx)
	d.cancel = cancel
	d.wg.Add(1)
	go d.loop(runCtx)
	d.logger.Info("outbox dispatcher started")
	return nil
}

func (d *Dispatcher) Stop() error {
	d.mu.Lock()
	if d.cancel == nil {
		d.mu.Unlock()
		return ErrDispatcherNotRunning
	}
	cancel := d.cancel
	d.cancel = nil
	d.mu.Unlock()

	cancel()
	d.wg.Wait()
	d.metrics.reset()
	d.logger.Info("outbox dispatcher stopped")
	return nil
}

func (d *Dispatcher) loop(ctx context.Context) {
	defer d.wg.Done()
	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.metrics.activeGauge.Set(1)
			d.dispatchBatch(ctx)
			d.metrics.activeGauge.Set(0)
		}
	}
}

func (d *Dispatcher) dispatchBatch(ctx context.Context) {
	if d.repo == nil || d.bus == nil || d.withTx == nil {
		return
	}

	var events []*database.OutboxEvent
	if err := d.withTx(ctx, func(txCtx context.Context, tx database.Transaction) error {
		var err error
		events, err = d.repo.GetUnpublishedForUpdate(txCtx, tx, d.cfg.BatchSize)
		if err != nil {
			return fmt.Errorf("fetch unpublished events: %w", err)
		}
		return nil
	}); err != nil {
		if ctx.Err() != nil {
			return
		}
		d.logger.Errorf("dispatcher batch failed: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, evt := range events {
		if ctx.Err() != nil {
			return
		}
		if err := d.publishOne(ctx, evt); err != nil {
			d.logger.Warnf("dispatch event failed: id=%s retries=%d err=%v", evt.EventID, evt.RetryCount, err)
		}
	}
}

func (d *Dispatcher) publishOne(ctx context.Context, evt *database.OutboxEvent) error {
	if evt == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := d.bus.Publish(ctx, d.asEvent(evt)); err != nil {
		if agg, ok := err.(*eventbus.AggregatePublishError); ok {
			for _, failure := range agg.Failures() {
				d.logger.Warnf("handler failure: event=%s handler=%d err=%v", agg.EventType(), failure.Index, failure.Err)
			}
		}
		next := time.Now().Add(d.nextBackoff(evt.RetryCount + 1))
		d.metrics.publishFailure.Inc()
		d.metrics.retryScheduled.Inc()
		if evt.RetryCount+1 >= d.cfg.MaxRetry {
			d.logger.Errorf("event retry threshold exceeded: id=%s retries=%d", evt.EventID, evt.RetryCount+1)
		}
		if incrErr := d.repo.IncrementRetryCount(ctx, evt.EventID, next); incrErr != nil {
			return fmt.Errorf("increment retry: %w", incrErr)
		}
		evt.RetryCount++
		evt.AvailableAt = next
		return err
	}
	d.metrics.publishSuccess.Inc()
	if err := d.repo.MarkPublished(ctx, evt.EventID); err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	evt.Published = true
	if now := time.Now(); evt.PublishedAt == nil {
		p := now
		evt.PublishedAt = &p
	}
	return nil
}

func (d *Dispatcher) nextBackoff(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	if retryCount > 5 {
		retryCount = 5
	}
	delay := d.cfg.BackoffBase * time.Duration(1<<retryCount)
	if delay > maxBackoff {
		delay = maxBackoff
	}
	return delay
}

func (d *Dispatcher) asEvent(evt *database.OutboxEvent) eventbus.Event {
	payload := json.RawMessage(evt.Payload)
	return eventbus.NewGenericJSONEvent(evt.EventType, evt.AggregateID, evt.AggregateType, payload)
}
