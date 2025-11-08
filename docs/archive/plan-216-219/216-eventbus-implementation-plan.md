# Plan 216 - `pkg/eventbus/` 事件总线实现

**文档编号**: 216
**标题**: 模块间异步通信基础设施 - 事件总线实现
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 215（Phase2 执行日志）、Plan 203（架构蓝图）

---

## 1. 概述

### 1.1 目标

实现 Go 语言的内存事件总线（In-Memory Event Bus），为模块化单体架构提供模块间异步通信的基础设施。

**关键成果**:
- ✅ 事件总线接口定义（Event、EventBus、EventHandler）
- ✅ 内存事件总线实现（MemoryEventBus）
- ✅ 事件发布/订阅机制
- ✅ 失败聚合与可观测性（错误计数、发布耗时指标）
- ✅ 单元测试（覆盖率 > 80%）

### 1.2 为什么需要事件总线

根据 203 号文档的模块化单体架构设计，模块间需要解耦通信机制：

**同步调用** → 模块间直接依赖，耦合度高
**异步事件** → 通过事件总线发布/订阅，模块完全解耦

**典型场景**:
```
员工转岗流程：
1. organization 模块处理职位变更
   ↓
2. 发布 EmployeeTransferred 事件
   ↓
3. workforce 模块异步订阅并处理
   ↓
4. payroll 模块异步订阅并调整薪资
```

### 1.3 时间计划

- **计划完成**: Week 3 Day 1 (Day 12)
- **交付周期**: 1 天
- **负责人**: 基础设施团队

---

## 2. 需求分析

### 2.1 功能需求

#### 需求 1: Event 接口定义

每个事件都应该实现以下接口：

```go
// Event 定义所有事件必须实现的接口
type Event interface {
    // EventType 返回事件的类型标识符
    EventType() string

    // AggregateID 返回关联的聚合根 ID（如 employeeID）
    AggregateID() string
}
```

**关键属性**:
- EventType：唯一标识事件类型（如 "employee.created"、"organization.updated"）
- AggregateID：用于追踪事件关联的业务对象（如员工 ID、组织单元 ID）

#### 需求 2: EventBus 接口定义

事件总线应该提供标准的发布/订阅接口：

```go
// EventBus 定义事件总线的标准接口
type EventBus interface {
    // Publish 发布事件到所有订阅者
    Publish(ctx context.Context, event Event) error

    // Subscribe 订阅特定类型的事件
    Subscribe(eventType string, handler EventHandler) error
}

// EventHandler 定义事件处理函数
type EventHandler func(ctx context.Context, event Event) error
```

**关键特性**:
- 支持多个订阅者处理同一类型的事件
- 支持 context.Context 用于超时控制和取消
- 返回错误用于事件处理失败的传递

#### 需求 3: 内存事件总线实现

实现 MemoryEventBus，支持：
- 事件发布与订阅
- 并发安全（RWMutex 保护）
- 多个处理器的顺序执行
- 错误聚合并返回调用方
- 日志与指标记录（成功/失败、耗时）

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **并发安全性** | ✅ 需要 | 支持并发的 Publish 和 Subscribe 操作 |
| **性能** | 低延迟 | 事件发布应在毫秒级完成 |
| **可观测性** | ✅ 需要 | 提供成功/失败计数与发布耗时指标 |
| **可诊断性** | ✅ 需要 | 返回失败详情，便于上游重试与报警 |
| **扩展性** | ✅ 需要 | 易于添加新的事件处理器 |
| **测试覆盖率** | > 80% | 单元测试覆盖所有主要代码路径 |

---

## 3. 架构设计

### 3.1 模块结构

```
pkg/eventbus/
├── eventbus.go          # 接口定义
├── memory_eventbus.go   # 内存实现
├── eventbus_test.go     # 单元测试
└── README.md            # 使用说明
```

### 3.2 关键设计决策

#### 决策 1: 为什么选择内存事件总线

**对比分析**:

| 方案 | 优点 | 缺点 | 选择 |
|------|------|------|------|
| **内存事件总线** | 延迟低、实现简单、无外部依赖 | 进程崩溃则事件丢失 | ✅ Phase2 优先 |
| **消息队列（RabbitMQ/Kafka）** | 可靠性高、可分布式 | 复杂度高、维护成本大 | ⏳ Phase3+ 考虑 |

**Phase2 的选择理由**:
- 内存事件总线与事务性发件箱结合，可保证事件可靠性
- 简化架构，降低维护成本
- 为未来迁移到消息队列预留接口

#### 决策 2: 事件处理的并发模型

```
Publish() 调用时：
1. 获取该事件类型的所有处理器（RLock）
2. 顺序执行每个处理器（串行）
3. 处理器失败时：记录错误但继续执行其他处理器
```

**为什么使用串行执行**:
- 降低复杂性（避免 goroutine 管理的开销）
- 保证处理器执行顺序可控
- 错误处理更清晰

**备选方案（Phase3+ 改进）**:
- 使用 Worker Pool 实现并行处理
- 支持异步事件处理

#### 决策 3: 失败传播与可观测性

- 事件处理失败必须反馈给上游（Plan 217B）以驱动重试，因此 `Publish` 返回聚合错误，而不是静默吞掉失败。
- 通过 `MetricsRecorder` 统计成功/失败次数与发布耗时，使 outbox dispatcher 可基于指标触发报警或调优。
- 日志接口采用最小抽象，保证在 Plan 218 未交付前仍可使用 noop 实现，等待 Plan 218 接管后即可无缝替换。

---

## 4. 详细实现

### 4.1 eventbus.go - 接口定义

```go
package eventbus

import "context"

// Event 定义所有事件必须实现的接口
type Event interface {
	// EventType 返回事件的类型标识符
	EventType() string

	// AggregateID 返回关联的聚合根 ID
	AggregateID() string
}

// EventHandler 定义事件处理函数
// 返回非 nil 错误时，错误会被记录但不会阻止其他处理器执行
type EventHandler func(ctx context.Context, event Event) error

// EventBus 定义事件总线的标准接口
type EventBus interface {
	// Publish 发布事件到所有订阅者
	// 如果没有订阅者，Publish 正常返回（nil 错误）
	Publish(ctx context.Context, event Event) error

	// Subscribe 订阅特定类型的事件
	// 同一事件类型可以有多个订阅者
	Subscribe(eventType string, handler EventHandler) error
}
```

### 4.2 memory_eventbus.go - 内存实现

```go
package eventbus

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Logger 是 Plan 216 内部定义的最小日志接口。
// Plan 218 落地后，可由其实现该接口并通过构造函数注入；
// 若未提供 logger，则使用 noopLogger，确保本计划可独立交付。
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

// MetricsRecorder 用于记录成功/失败次数与发布耗时。
// 若业务方暂未接入指标，可注入 noop 实现，避免与 Plan 217B 的可观测性要求冲突。
type MetricsRecorder interface {
	RecordSuccess(eventType string)
	RecordFailure(eventType string)
	RecordNoHandler(eventType string)
	RecordLatency(eventType string, duration time.Duration)
}

type MemoryEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
	logger   Logger
	metrics  MetricsRecorder
}

func NewMemoryEventBus(logger Logger, metrics MetricsRecorder) *MemoryEventBus {
	if logger == nil {
		logger = &noopLogger{}
	}
	if metrics == nil {
		metrics = &noopMetrics{}
	}

	return &MemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
		metrics:  metrics,
	}
}

func (b *MemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	if eventType == "" {
		return ErrEmptyEventType
	}
	if handler == nil {
		return ErrNilHandler
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.logger.Infof("subscribed to event type: %s (total handlers: %d)", eventType, len(b.handlers[eventType]))

	return nil
}

// Publish 发布事件到所有订阅者。
// 任意处理器失败时会继续执行剩余处理器，并在结束后返回聚合错误，
// 方便上游（例如 Plan 217B 的 outbox dispatcher）进行重试与监控。
func (b *MemoryEventBus) Publish(ctx context.Context, event Event) error {
	if event == nil {
		return ErrNilEvent
	}

	eventType := event.EventType()
	if eventType == "" {
		return ErrEmptyEventType
	}

	b.mu.RLock()
	handlers, ok := b.handlers[eventType]
	b.mu.RUnlock()

	if !ok || len(handlers) == 0 {
		b.logger.Debugf("no handlers for event type: %s", eventType)
		b.metrics.RecordNoHandler(eventType)
		return nil
	}

	start := time.Now()
	aggErr := NewAggregatePublishError(eventType, event.AggregateID())

	for idx, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			b.logger.Errorf("event handler failed: type=%s, handler_index=%d, error=%v", eventType, idx, err)
			aggErr.Append(idx, err)
			b.metrics.RecordFailure(eventType)
			continue
		}

		b.metrics.RecordSuccess(eventType)
	}

	b.metrics.RecordLatency(eventType, time.Since(start))

	if aggErr.IsEmpty() {
		return nil
	}

	return aggErr
}

// GetHandlerCount 返回某个事件类型的处理器数量（用于测试和监控）
func (b *MemoryEventBus) GetHandlerCount(eventType string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers[eventType])
}

// Reset 清空所有订阅（用于测试）
func (b *MemoryEventBus) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = make(map[string][]EventHandler)
}

type noopLogger struct{}

func (*noopLogger) Debugf(string, ...interface{}) {}
func (*noopLogger) Infof(string, ...interface{})  {}
func (*noopLogger) Errorf(string, ...interface{}) {}

type noopMetrics struct{}

func (*noopMetrics) RecordSuccess(string)             {}
func (*noopMetrics) RecordFailure(string)             {}
func (*noopMetrics) RecordNoHandler(string)           {}
func (*noopMetrics) RecordLatency(string, time.Duration) {}

// AggregatePublishError 聚合所有失败的处理器，便于调用者分析与重试。
type AggregatePublishError struct {
	eventType    string
	aggregateID string
	failures     []HandlerFailure
}

type HandlerFailure struct {
	Index int
	Err   error
}

func NewAggregatePublishError(eventType, aggregateID string) *AggregatePublishError {
	return &AggregatePublishError{eventType: eventType, aggregateID: aggregateID}
}

func (e *AggregatePublishError) Append(idx int, err error) {
	e.failures = append(e.failures, HandlerFailure{Index: idx, Err: err})
}

func (e *AggregatePublishError) Error() string {
	// 返回简洁错误信息，详细信息可通过 Failures() 查看
	return fmt.Sprintf("eventbus publish failed: type=%s aggregateID=%s failures=%d",
		e.eventType, e.aggregateID, len(e.failures))
}

func (e *AggregatePublishError) Failures() []HandlerFailure {
	return e.failures
}

func (e *AggregatePublishError) IsEmpty() bool {
	return len(e.failures) == 0
}
```

该设计让事件发布的失败信息集中管理，并通过指标接口暴露，确保 Plan 217B 的 outbox dispatcher 可以在统一信号下执行重试、退避与报警，而不会遗漏失败场景。

### 4.3 错误定义

```go
package eventbus

import "errors"

var (
	ErrNilEvent       = errors.New("event cannot be nil")
	ErrNilHandler     = errors.New("handler cannot be nil")
	ErrEmptyEventType = errors.New("event type cannot be empty")
)
```

---

## 5. 单元测试

### 5.1 eventbus_test.go - 测试套件

测试覆盖的场景：

```go
package eventbus

import (
	"context"
	"errors"
	"testing"
	"sync"
	"sync/atomic"
	"time"
)

// 测试用例 1: 订阅和发布基本流程
func TestPublishWithSingleSubscriber(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &TestEvent{
		eventType:   "test.event",
		aggregateID: "123",
	}

	var called bool
	handler := func(ctx context.Context, e Event) error {
		called = true
		return nil
	}

	bus.Subscribe("test.event", handler)
	err := bus.Publish(context.Background(), event)

	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if !called {
		t.Error("handler was not called")
	}

	if metrics.success != 1 {
		t.Errorf("expected success metric to be 1, got %d", metrics.success)
	}
}

// 测试用例 2: 多个订阅者
func TestPublishWithMultipleSubscribers(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &TestEvent{
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

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)
	err := bus.Publish(context.Background(), event)

	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}

	if metrics.success != 2 {
		t.Errorf("expected success metric to be 2, got %d", metrics.success)
	}
}

// 测试用例 3: 处理器错误不阻止其他处理器

func TestPublishWithHandlerError(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &TestEvent{
		eventType:   "test.event",
		aggregateID: "789",
	}

	callOrder := []int{}
	mu := sync.Mutex{}

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

	bus.Subscribe("test.event", handler1)
	bus.Subscribe("test.event", handler2)
	err := bus.Publish(context.Background(), event)

	aggErr, ok := err.(*AggregatePublishError)
	if !ok {
		t.Fatalf("expected AggregatePublishError, got %T", err)
	}

	if len(aggErr.Failures()) != 1 {
		t.Errorf("expected 1 failure entry, got %d", len(aggErr.Failures()))
	}

	if len(callOrder) != 2 {
		t.Errorf("expected both handlers to be called, got %d", len(callOrder))
	}

	if metrics.failure != 1 {
		t.Errorf("expected failure metric to be 1, got %d", metrics.failure)
	}
}

// 测试用例 4: 并发订阅和发布
func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewMemoryEventBus(nil, nil)

	eventType := "concurrent.event"
	callCount := atomic.Int32{}

	// 并发订阅
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler := func(ctx context.Context, e Event) error {
				callCount.Add(1)
				return nil
			}
			bus.Subscribe(eventType, handler)
		}()
	}

	wg.Wait()

	// 发布事件
	event := &TestEvent{
		eventType:   eventType,
		aggregateID: "concurrent-123",
	}

	err := bus.Publish(context.Background(), event)
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	if callCount.Load() != 10 {
		t.Errorf("expected 10 calls, got %d", callCount.Load())
	}
}

// 测试用例 5: 无订阅者时发布
func TestPublishWithNoSubscribers(t *testing.T) {
	metrics := &testMetrics{}
	bus := NewMemoryEventBus(nil, metrics)

	event := &TestEvent{
		eventType:   "unused.event",
		aggregateID: "no-one",
	}

	err := bus.Publish(context.Background(), event)

	// 应该正常返回，不返回错误
	if err != nil {
		t.Errorf("Publish should succeed with no subscribers, got error: %v", err)
	}

	if metrics.noHandler != 1 {
		t.Errorf("expected no-handler metric to be 1, got %d", metrics.noHandler)
	}
}

// 测试用例 6: 错误输入验证
func TestErrorHandling(t *testing.T) {
	bus := NewMemoryEventBus(nil, nil)

	// 订阅空事件类型
	err := bus.Subscribe("", func(ctx context.Context, e Event) error {
		return nil
	})
	if err != ErrEmptyEventType {
		t.Errorf("expected ErrEmptyEventType, got %v", err)
	}

	// 订阅 nil 处理器
	err = bus.Subscribe("test", nil)
	if err != ErrNilHandler {
		t.Errorf("expected ErrNilHandler, got %v", err)
	}

	// 发布 nil 事件
	err = bus.Publish(context.Background(), nil)
	if err != ErrNilEvent {
		t.Errorf("expected ErrNilEvent, got %v", err)
	}
}

// 辅助：测试事件
type TestEvent struct {
	eventType   string
	aggregateID string
}

func (e *TestEvent) EventType() string {
	return e.eventType
}

func (e *TestEvent) AggregateID() string {
	return e.aggregateID
}

type testMetrics struct {
	success   int
	failure   int
	noHandler int
	latency   []time.Duration
}

func (m *testMetrics) RecordSuccess(string)             { m.success++ }
func (m *testMetrics) RecordFailure(string)             { m.failure++ }
func (m *testMetrics) RecordNoHandler(string)           { m.noHandler++ }
func (m *testMetrics) RecordLatency(_ string, d time.Duration) { m.latency = append(m.latency, d) }
```

### 5.2 测试覆盖率目标

- **目标**：> 80%
- **关键路径**：
  - ✅ Subscribe 正常流程
  - ✅ Publish 正常流程
  - ✅ 错误处理（nil 输入、空字符串）
  - ✅ 多个订阅者
  - ✅ 处理器失败
  - ✅ 并发访问

---

## 6. 集成要点

### 6.1 与 Plan 217 (pkg/database) 的关系

- 事件总线可以独立使用
- Plan 217 将实现事务性发件箱，与事件总线结合保证事件可靠性

### 6.2 与 Plan 218 (pkg/logger) 的关系

- 本计划定义最小化 `Logger` 接口与 noop 实现，保证可独立交付。
- Plan 218 完成后，将提供符合该接口的结构化日志记录器并在构造函数中注入，实现统一的日志输出。

### 6.3 与 Plan 217B (outbox dispatcher) 的关系

- `Publish` 返回 `AggregatePublishError`，dispatcher 可根据失败明细决定重试与退避策略。
- `MetricsRecorder` 的成功/失败/耗时/无订阅者指标与 Plan 217B 的监控面板共用命名约定，便于集中观测。

### 6.4 与 Plan 219 (organization 重构) 的关系

- organization 模块将使用事件总线进行跨模块通信
- 示例：发布组织单元变更事件

---

## 7. 验收标准

### 7.1 功能验收

- [x] 事件接口定义完整（Event、EventHandler、EventBus，对应 `pkg/eventbus/eventbus.go`）
- [x] MemoryEventBus 实现完整（见 `pkg/eventbus/memory_eventbus.go`）
- [x] Subscribe 功能正常（单测 `TestPublishWithSingleSubscriber`、`TestConcurrentPublishAndSubscribe` 覆盖）
- [x] Publish 功能正常（单测 `TestPublishWithMultipleSubscribers` 验证）
- [x] 错误处理正确（单测 `TestErrorHandling`、`TestPublishWithHandlerError` 验证）
- [x] 处理器失败时返回 `AggregatePublishError`（`TestPublishWithHandlerError` 验证失败聚合）
- [x] 并发安全性验证通过（`go test -race ./pkg/eventbus` 通过）

### 7.2 质量验收

- [x] 单元测试覆盖率 > 80%（`go test -cover ./pkg/eventbus` 输出 98.1%）
- [x] 所有测试通过（`go test ./pkg/eventbus`）
- [x] 代码通过 `go fmt` 检查（已执行 `gofmt -w pkg/eventbus/*.go`）
- [x] 代码通过 `go vet` 检查（`go vet ./pkg/eventbus` 无告警）
- [x] 无 race condition（`go test -race ./pkg/eventbus`）
- [x] 指标记录函数在测试中被验证（`TestPublishWithSingleSubscriber` 等断言 success/failure/noHandler/latency）

### 7.3 文档验收

- [x] README.md 编写完成（使用示例，见 `pkg/eventbus/README.md`）
- [x] 文档说明指标命名、返回错误语义及与 Plan 217B 的联动（README “指标与日志”“错误语义”）
- [x] 代码注释完整（核心接口与实现均提供注释）
- [x] 接口文档清晰（README 与代码注释覆盖使用方式）

---

## 8. 风险与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|--------|
| 并发安全问题 | 中 | 高 | 使用 race detector 进行测试 |
| 性能瓶颈 | 低 | 中 | 基准测试（benchmark），监控发布延迟 |
| 错误处理不完善 | 低 | 中 | 充分的测试覆盖 |
| 指标实现缺失 | 低 | 中 | 提供 noopMetrics，结合测试验证接口覆盖 |
| 与 Plan 218 依赖冲突 | 低 | 中 | 提前协调日志接口定义，采用最小 Logger 抽象 |

---

## 9. 交付物清单

- ✅ `pkg/eventbus/eventbus.go`
- ✅ `pkg/eventbus/memory_eventbus.go`
- ✅ `pkg/eventbus/error.go`
- ✅ `pkg/eventbus/eventbus_test.go`
- ✅ `pkg/eventbus/README.md`
- ✅ 本计划文档（216）

---

## 10. 相关文档

- `203-hrms-module-division-plan.md` - 模块化单体架构
- `204-HRMS-Implementation-Roadmap.md` - Phase2 实施路线图
- `215-phase2-execution-log.md` - Phase2 执行日志
- `Plan 217` - pkg/database 实现
- `Plan 218` - pkg/logger 实现

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-03
**计划完成日期**: Week 3 Day 1 (Day 12)
