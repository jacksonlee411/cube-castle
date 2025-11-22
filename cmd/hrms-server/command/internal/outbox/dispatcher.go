package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	orgutils "cube-castle/internal/organization/utils"
	"cube-castle/pkg/database"
	"cube-castle/pkg/eventbus"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
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
	cache   AssignmentCacheRefresher

	mu     sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

const (
	maxBackoff        = 5 * time.Minute
	outboxResultRetry = "retry"
)

// NewDispatcher 构造基于数据库 outbox 的事件派发器。
func NewDispatcher(cfg Config, repo database.OutboxRepository, bus eventbus.EventBus, logger pkglogger.Logger, reg prometheus.Registerer, withTx func(ctx context.Context, fn database.TxFunc) error, cache AssignmentCacheRefresher) *Dispatcher {
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
		cache:   cache,
	}
}

// Start 启动后台轮询循环。
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

// Stop 停止派发器并回收资源。
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
		orgutils.RecordOutboxDispatch(orgutils.StatusError, evt.EventType)
		if evt.RetryCount+1 >= d.cfg.MaxRetry {
			d.logger.Errorf("event retry threshold exceeded: id=%s retries=%d", evt.EventID, evt.RetryCount+1)
		}
		if incrErr := d.repo.IncrementRetryCount(ctx, evt.EventID, next); incrErr != nil {
			return fmt.Errorf("increment retry: %w", incrErr)
		}
		orgutils.RecordOutboxDispatch(outboxResultRetry, evt.EventType)
		evt.RetryCount++
		evt.AvailableAt = next
		return err
	}
	d.metrics.publishSuccess.Inc()
	if err := d.repo.MarkPublished(ctx, evt.EventID); err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	orgutils.RecordOutboxDispatch(orgutils.StatusSuccess, evt.EventType)
	evt.Published = true
	if now := time.Now(); evt.PublishedAt == nil {
		p := now
		evt.PublishedAt = &p
	}
	d.handlePostPublish(ctx, evt)
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

// AssignmentCacheRefresher 刷新职位缓存以反映最新任命。
type AssignmentCacheRefresher interface {
	RefreshPositionCache(ctx context.Context, tenantID uuid.UUID, positionCode string) error
}

const (
	eventAssignmentFilled  = "assignment.filled"
	eventAssignmentVacated = "assignment.vacated"
	eventAssignmentUpdated = "assignment.updated"
	eventAssignmentClosed  = "assignment.closed"
)

type assignmentEventPayload struct {
	TenantID     string `json:"tenantId"`
	PositionCode string `json:"positionCode"`
	Position     struct {
		Code string `json:"code"`
	} `json:"position"`
}

func (d *Dispatcher) handlePostPublish(ctx context.Context, evt *database.OutboxEvent) {
	if evt == nil {
		return
	}
	switch evt.EventType {
	case eventAssignmentFilled, eventAssignmentVacated, eventAssignmentUpdated, eventAssignmentClosed:
		d.invalidateAssignmentCache(ctx, evt)
	}
}

func (d *Dispatcher) invalidateAssignmentCache(ctx context.Context, evt *database.OutboxEvent) {
	if d.cache == nil || evt == nil {
		return
	}

	var payload assignmentEventPayload
	if err := json.Unmarshal([]byte(evt.Payload), &payload); err != nil {
		d.logger.WithFields(pkglogger.Fields{
			"eventId": evt.EventID,
			"type":    evt.EventType,
			"error":   err,
		}).Warn("failed to decode assignment event payload")
		return
	}

	positionCode := strings.TrimSpace(payload.PositionCode)
	if positionCode == "" {
		positionCode = strings.TrimSpace(payload.Position.Code)
	}
	if positionCode == "" {
		d.logger.WithFields(pkglogger.Fields{
			"eventId": evt.EventID,
			"type":    evt.EventType,
		}).Warn("assignment event missing positionCode, skip cache refresh")
		return
	}

	tenantIDStr := strings.TrimSpace(payload.TenantID)
	if tenantIDStr == "" {
		tenantIDStr = strings.TrimSpace(evt.AggregateID)
	}
	tenantUUID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		d.logger.WithFields(pkglogger.Fields{
			"eventId":  evt.EventID,
			"type":     evt.EventType,
			"tenantId": tenantIDStr,
			"error":    err,
		}).Warn("invalid tenant id in assignment event payload")
		return
	}

	if err := d.cache.RefreshPositionCache(ctx, tenantUUID, positionCode); err != nil {
		d.logger.WithFields(pkglogger.Fields{
			"eventId":      evt.EventID,
			"type":         evt.EventType,
			"tenantId":     tenantUUID.String(),
			"positionCode": positionCode,
			"error":        err,
		}).Warn("failed to refresh assignment cache")
	} else {
		d.logger.WithFields(pkglogger.Fields{
			"eventId":      evt.EventID,
			"type":         evt.EventType,
			"tenantId":     tenantUUID.String(),
			"positionCode": positionCode,
		}).Debug("assignment cache refreshed")
	}
}
