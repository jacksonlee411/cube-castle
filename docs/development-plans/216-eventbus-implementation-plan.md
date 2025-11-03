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
- ✅ 错误重试和日志记录
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
- 错误处理和日志记录

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **并发安全性** | ✅ 需要 | 支持并发的 Publish 和 Subscribe 操作 |
| **性能** | 低延迟 | 事件发布应在毫秒级完成 |
| **可观测性** | ✅ 需要 | 支持日志和指标记录 |
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
	"sync"

	"cube-castle/internal/logging"  // 使用 Plan 218 的日志系统
)

// MemoryEventBus 是事件总线的内存实现
type MemoryEventBus struct {
	// handlers 存储每个事件类型对应的处理器列表
	// key: eventType (如 "employee.created")
	// value: 处理器函数列表
	handlers map[string][]EventHandler

	// mu 保护 handlers 的并发访问
	mu sync.RWMutex

	// logger 用于记录事件和错误
	logger logging.Logger
}

// NewMemoryEventBus 创建一个新的内存事件总线
func NewMemoryEventBus(logger logging.Logger) *MemoryEventBus {
	if logger == nil {
		// 如果没有提供 logger，使用 noop logger
		logger = logging.NewNoopLogger()
	}

	return &MemoryEventBus{
		handlers: make(map[string][]EventHandler),
		logger:   logger,
	}
}

// Subscribe 订阅特定类型的事件
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

	b.logger.Infof("subscribed to event type: %s (total handlers: %d)",
		eventType, len(b.handlers[eventType]))

	return nil
}

// Publish 发布事件到所有订阅者
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

	// 如果没有订阅者，正常返回
	if !ok || len(handlers) == 0 {
		b.logger.Debugf("no handlers for event type: %s", eventType)
		return nil
	}

	b.logger.Infof("publishing event: type=%s, aggregateID=%s (handlers=%d)",
		eventType, event.AggregateID(), len(handlers))

	// 顺序执行所有处理器
	var lastErr error
	for i, handler := range handlers {
		err := handler(ctx, event)
		if err != nil {
			b.logger.Errorf("event handler failed: type=%s, handler_index=%d, error=%v",
				eventType, i, err)
			lastErr = err
			// 继续执行其他处理器，不中断
		}
	}

	return nil // 返回 nil，即使某些处理器失败
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
```

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
)

// 测试用例 1: 订阅和发布基本流程
func TestPublishWithSingleSubscriber(t *testing.T) {
	bus := NewMemoryEventBus(nil)

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
}

// 测试用例 2: 多个订阅者
func TestPublishWithMultipleSubscribers(t *testing.T) {
	bus := NewMemoryEventBus(nil)

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
}

// 测试用例 3: 处理器错误不阻止其他处理器
func TestPublishWithHandlerError(t *testing.T) {
	bus := NewMemoryEventBus(nil)

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

	// Publish 应该返回 nil 错误（即使处理器失败）
	if err != nil {
		t.Errorf("Publish should return nil error, got %v", err)
	}

	// 两个处理器都应该被调用
	if len(callOrder) != 2 {
		t.Errorf("expected 2 calls, got %d", len(callOrder))
	}
}

// 测试用例 4: 并发订阅和发布
func TestConcurrentPublishAndSubscribe(t *testing.T) {
	bus := NewMemoryEventBus(nil)

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
	bus := NewMemoryEventBus(nil)

	event := &TestEvent{
		eventType:   "unused.event",
		aggregateID: "no-one",
	}

	err := bus.Publish(context.Background(), event)

	// 应该正常返回，不返回错误
	if err != nil {
		t.Errorf("Publish should succeed with no subscribers, got error: %v", err)
	}
}

// 测试用例 6: 错误输入验证
func TestErrorHandling(t *testing.T) {
	bus := NewMemoryEventBus(nil)

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

- MemoryEventBus 依赖 Plan 218 的日志系统
- 使用 logging.Logger 接口记录事件发布和处理

### 6.3 与 Plan 219 (organization 重构) 的关系

- organization 模块将使用事件总线进行跨模块通信
- 示例：发布组织单元变更事件

---

## 7. 验收标准

### 7.1 功能验收

- [ ] 事件接口定义完整（Event、EventHandler、EventBus）
- [ ] MemoryEventBus 实现完整
- [ ] Subscribe 功能正常
- [ ] Publish 功能正常
- [ ] 错误处理正确
- [ ] 并发安全性验证通过

### 7.2 质量验收

- [ ] 单元测试覆盖率 > 80%
- [ ] 所有测试通过 (`go test ./pkg/eventbus -v`)
- [ ] 代码通过 `go fmt` 检查
- [ ] 代码通过 `go vet` 检查
- [ ] 无 race condition（`go test -race ./pkg/eventbus`）

### 7.3 文档验收

- [ ] README.md 编写完成（使用示例）
- [ ] 代码注释完整
- [ ] 接口文档清晰

---

## 8. 风险与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|--------|
| 并发安全问题 | 中 | 高 | 使用 race detector 进行测试 |
| 性能瓶颈 | 低 | 中 | 基准测试（benchmark），监控发布延迟 |
| 错误处理不完善 | 低 | 中 | 充分的测试覆盖 |
| 与 Plan 218 依赖冲突 | 低 | 中 | 提前协调日志接口定义 |

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
**最后更新**: 2025-11-04
**计划完成日期**: Week 3 Day 1 (Day 12)
