package eventbus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishWithSingleSubscriber(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &testEvent{
		eventType:   "test.event",
		aggregateID: "123",
	}

	var called bool
	handler := func(ctx context.Context, e Event) error {
		called = true
		return nil
	}

	if err := bus.Subscribe("test.event", handler); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}

	if metrics.success != 1 {
		t.Fatalf("expected success metric to be 1, got %d", metrics.success)
	}
}

func TestPublishWithMultipleSubscribers(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &testEvent{
		eventType:   "test.event",
		aggregateID: "456",
	}

	callCount := 0
	handler1 := func(ctx context.Context, e Event) error {
		callCount++
		return nil
	}
	handler2 := func(ctx context.Context, e Event) error {
		callCount++
		return nil
	}

	if err := bus.Subscribe("test.event", handler1); err != nil {
		t.Fatalf("Subscribe handler1 failed: %v", err)
	}
	if err := bus.Subscribe("test.event", handler2); err != nil {
		t.Fatalf("Subscribe handler2 failed: %v", err)
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 calls, got %d", callCount)
	}
	if metrics.success != 2 {
		t.Fatalf("expected success metric to be 2, got %d", metrics.success)
	}
}

func TestPublishWithHandlerError(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &testEvent{
		eventType:   "test.event",
		aggregateID: "789",
	}

	callOrder := make([]int, 0, 2)
	var mu sync.Mutex

	handler1 := func(ctx context.Context, e Event) error {
		mu.Lock()
		callOrder = append(callOrder, 1)
		mu.Unlock()
		return errors.New("handler1 error")
	}
	handler2 := func(ctx context.Context, e Event) error {
		mu.Lock()
		callOrder = append(callOrder, 2)
		mu.Unlock()
		return nil
	}

	if err := bus.Subscribe("test.event", handler1); err != nil {
		t.Fatalf("Subscribe handler1 failed: %v", err)
	}
	if err := bus.Subscribe("test.event", handler2); err != nil {
		t.Fatalf("Subscribe handler2 failed: %v", err)
	}

	err := bus.Publish(context.Background(), event)
	if err == nil {
		t.Fatal("expected Publish to return error")
	}

	aggErr, ok := err.(*AggregatePublishError)
	if !ok {
		t.Fatalf("expected AggregatePublishError, got %T", err)
	}

	if aggErr.EventType() != "test.event" {
		t.Fatalf("unexpected event type: %s", aggErr.EventType())
	}
	if aggErr.AggregateID() != "789" {
		t.Fatalf("unexpected aggregate id: %s", aggErr.AggregateID())
	}
	if len(aggErr.Failures()) != 1 {
		t.Fatalf("expected 1 failure entry, got %d", len(aggErr.Failures()))
	}
	if aggErr.Failures()[0].Index != 0 {
		t.Fatalf("unexpected failure index: %d", aggErr.Failures()[0].Index)
	}

	if len(callOrder) != 2 {
		t.Fatalf("expected both handlers to be called, got %d", len(callOrder))
	}
	if metrics.failure != 1 {
		t.Fatalf("expected failure metric to be 1, got %d", metrics.failure)
	}
	if metrics.success != 1 {
		t.Fatalf("expected success metric to be 1, got %d", metrics.success)
	}
}

func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewMemoryEventBus(nil, nil)

	eventType := "concurrent.event"
	callCount := atomic.Int32{}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := func(ctx context.Context, e Event) error {
				callCount.Add(1)
				return nil
			}
			if err := bus.Subscribe(eventType, handler); err != nil {
				t.Errorf("Subscribe failed: %v", err)
			}
		}()
	}

	wg.Wait()

	event := &testEvent{
		eventType:   eventType,
		aggregateID: "concurrent-123",
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if callCount.Load() != 10 {
		t.Fatalf("expected 10 calls, got %d", callCount.Load())
	}

	if got := bus.GetHandlerCount(eventType); got != 10 {
		t.Fatalf("expected handler count 10, got %d", got)
	}
}

func TestPublishWithNoSubscribers(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &testEvent{
		eventType:   "unused.event",
		aggregateID: "no-one",
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish should succeed with no subscribers, got error: %v", err)
	}

	if metrics.noHandler != 1 {
		t.Fatalf("expected no-handler metric to be 1, got %d", metrics.noHandler)
	}
}

func TestErrorHandling(t *testing.T) {
	bus := NewMemoryEventBus(nil, nil)

	if err := bus.Subscribe("", func(ctx context.Context, e Event) error { return nil }); !errors.Is(err, ErrEmptyEventType) {
		t.Fatalf("expected ErrEmptyEventType, got %v", err)
	}

	if err := bus.Subscribe("test", nil); !errors.Is(err, ErrNilHandler) {
		t.Fatalf("expected ErrNilHandler, got %v", err)
	}

	if err := bus.Publish(context.Background(), nil); !errors.Is(err, ErrNilEvent) {
		t.Fatalf("expected ErrNilEvent, got %v", err)
	}

	ev := &testEvent{eventType: "", aggregateID: "agg"}
	if err := bus.Publish(context.Background(), ev); !errors.Is(err, ErrEmptyEventType) {
		t.Fatalf("expected ErrEmptyEventType for empty event type, got %v", err)
	}
}

func TestReset(t *testing.T) {
	bus := NewMemoryEventBus(nil, nil)
	if err := bus.Subscribe("a", func(ctx context.Context, e Event) error { return nil }); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}
	if got := bus.GetHandlerCount("a"); got != 1 {
		t.Fatalf("expected handler count 1 before reset, got %d", got)
	}

	bus.Reset()
	if got := bus.GetHandlerCount("a"); got != 0 {
		t.Fatalf("expected handler count 0 after reset, got %d", got)
	}
}

func TestLatencyMetricRecorded(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &testEvent{
		eventType:   "latency.event",
		aggregateID: "agg",
	}

	if err := bus.Subscribe("latency.event", func(ctx context.Context, e Event) error {
		time.Sleep(5 * time.Millisecond)
		return nil
	}); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	if err := bus.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	if len(metrics.latency) != 1 {
		t.Fatalf("expected latency metric recorded once, got %d", len(metrics.latency))
	}
}

type testEvent struct {
	eventType   string
	aggregateID string
}

func (e *testEvent) EventType() string   { return e.eventType }
func (e *testEvent) AggregateID() string { return e.aggregateID }

type testMetrics struct {
	success   int
	failure   int
	noHandler int
	latency   []time.Duration
}

func (m *testMetrics) RecordSuccess(string)                    { m.success++ }
func (m *testMetrics) RecordFailure(string)                    { m.failure++ }
func (m *testMetrics) RecordNoHandler(string)                  { m.noHandler++ }
func (m *testMetrics) RecordLatency(_ string, d time.Duration) { m.latency = append(m.latency, d) }
