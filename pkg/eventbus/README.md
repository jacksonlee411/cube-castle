# pkg/eventbus

内存事件总线实现，支撑模块化单体架构中的异步解耦（Plan 216）。

## 功能概述

- 定义 `Event`、`EventBus`、`EventHandler` 标准接口。
- 提供并发安全的 `MemoryEventBus`，支持多订阅者顺序执行。
- 聚合处理器错误并返回 `AggregatePublishError`，方便上游重试与观测。
- `Logger` 与 `MetricsRecorder` 采用可选注入，默认使用 noop 实现保持最小依赖。

## 快速上手

```go
package main

import (
	"context"
	"log"

	"cube-castle/pkg/eventbus"
)

type employeeCreated struct {
	id string
}

func (e employeeCreated) EventType() string  { return "employee.created" }
func (e employeeCreated) AggregateID() string { return e.id }

func main() {
	bus := eventbus.NewMemoryEventBus(nil, nil)

	_ = bus.Subscribe("employee.created", func(ctx context.Context, evt eventbus.Event) error {
		log.Printf("handle employee: %s", evt.AggregateID())
		return nil
	})

	_ = bus.Publish(context.Background(), employeeCreated{id: "emp-001"})
}
```

## 指标与日志

`MetricsRecorder` 接口建议映射到以下 Prometheus 指标名称（Plan 217B 将复用同一命名约定）：

- `eventbus_publish_success_total`
- `eventbus_publish_failure_total`
- `eventbus_publish_no_handler_total`
- `eventbus_publish_latency_seconds`

`Logger` 接口将由 Plan 218 的结构化日志落地，目前默认 noop，不影响后续接入。

## 错误语义

- 输入校验：`ErrNilEvent`、`ErrEmptyEventType`、`ErrNilHandler`。
- `AggregatePublishError`：聚合所有失败的处理器，`Failures()` 返回索引和原始错误，便于 outbox dispatcher（Plan 217B）决定重试策略。

## 测试

```bash
go test ./pkg/eventbus
go test -race ./pkg/eventbus
```
