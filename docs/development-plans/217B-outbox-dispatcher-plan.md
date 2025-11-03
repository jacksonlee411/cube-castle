# Plan 217B - 事务性发件箱中继（Outbox Dispatcher）

**文档编号**: 217B  
**标题**: Outbox Dispatcher 事件中继实现方案  
**创建日期**: 2025-11-04  
**分支**: `feature/204-phase2-infrastructure`  
**版本**: v1.0  
**关联计划**: Plan 216（事件总线）、Plan 217（数据库访问层）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

在命令服务进程内实现一个可靠的 Outbox Dispatcher，用于：
- 轮询 `outbox` 表中未发布的事件；
- 通过 Plan 216 提供的 `eventbus.EventBus` 发布事件；
- 在发布成功后将事件标记为已发布，并维护重试统计；
- 暴露指标与日志，支持运维监控。

该中继是事务性发件箱模式的关键组成部分，确保服务不会在事务失败时仍然发布事件，从而满足“资源唯一性与跨层一致性”的最高优先级约束。

### 1.2 背景

Plan 217 已经提供：
- 统一的数据库访问层；
- OutboxEvent 结构和 `OutboxRepository` 接口；
- 基本的保存、查询、标记发布、递增重试功能。

Plan 216 提供内存事件总线接口，但尚缺少从数据库中读取事件并调用 EventBus 的中间层。本计划填补该空白，使 Phase2 的事件流闭环落地。

---

## 2. 需求

### 2.1 功能需求

1. **轮询调度**  
   - 默认每 5 秒扫描一次 `outbox` 表，可通过环境变量 `OUTBOX_DISPATCH_INTERVAL` 覆盖。
   - 每次批量获取最多 50 条待发布事件，可配置 `OUTBOX_DISPATCH_BATCH_SIZE`。

2. **事件发布**  
   - 对每条事件调用 `eventbus.Publish(ctx, event)`；失败时记录错误并进行重试策略。
   - 成功发布后调用 `OutboxRepository.MarkPublished`。

3. **重试与退避**  
   - 记录失败次数并调用 `IncrementRetryCount`。
   - 采用指数退避：`next_interval = baseInterval * 2^min(retryCount, 5)`，最大不超过 5 分钟。
   - 连续失败超过阈值（默认 10 次）时输出告警日志并留在队列中。

4. **幂等保障**  
   - 基于 `event_id` 控制重复发布：MarkPublished 后跳过重复事件。
   - 发布前检查事件是否已被标记。

5. **安全停止**  
   - 支持 `context.Context` 取消，收到信号时优雅退出。

### 2.2 非功能需求

| 指标 | 目标 | 说明 |
|------|------|------|
| 延迟 | P99 发布延迟 < 10 秒 | 由轮询间隔与退避策略保证 |
| 可靠性 | 0 次丢失事件 | 事务提交失败不得触发 Publish |
| 覆盖率 | > 80% | 单元 + 集成测试 |
| 可观测性 | Prometheus + 结构化日志 | 指标：成功/失败/重试次数 |
| 配置化 | 环境变量 | Interval、BatchSize、MaxRetry、Backoff 等 |

---

## 3. 设计

### 3.1 组件结构

```
cmd/hrms-server/internal/outbox/
├── dispatcher.go          # Dispatcher 主循环
├── dispatcher_config.go   # 配置与默认值
├── repository.go          # OutboxRepository 接口包装（引用 Plan 217 实现）
├── metrics.go             # Prometheus 指标
├── errors.go              # 错误与哨兵值
├── dispatcher_test.go     # 单元测试
└── integration_test.go    # 集成测试（依赖数据库 + eventbus）
```

### 3.2 运行流程

```
Start()
  └─ init metrics/logger/config
  └─ ticker := time.NewTicker(interval)
  └─ for {
        select
        case <-ctx.Done():
            return
        case <-ticker.C:
            dispatchBatch()
        }
      }

dispatchBatch()
  └─ events := repo.GetUnpublished(limit)
  └─ for each event
        └─ if shouldSkip(event) continue
        └─ err := bus.Publish(ctx, mapToDomainEvent(event))
        └─ if err != nil {
              repo.IncrementRetryCount(...)
              metrics.failure.Inc()
              logger.Warnf(...)
              continue
           }
        └─ repo.MarkPublished(...)
        └─ metrics.success.Inc()
```

### 3.3 配置参数（默认值）

| 名称 | 环境变量 | 默认值 |
|------|----------|--------|
| 轮询间隔 | `OUTBOX_DISPATCH_INTERVAL` | 5s |
| 批量大小 | `OUTBOX_DISPATCH_BATCH_SIZE` | 50 |
| 最大重试次数 | `OUTBOX_DISPATCH_MAX_RETRY` | 10 |
| 退避基准 | `OUTBOX_DISPATCH_BACKOFF_BASE` | 5s |
| 指标前缀 | `OUTBOX_DISPATCH_METRIC_PREFIX` | `outbox_dispatch_` |

---

## 4. 实施步骤

### 4.1 代码实现

1. 创建 `dispatcher.Config` 结构体与加载逻辑（环境变量 + 默认值 + 参数校验）。  
2. 实现 `Dispatcher` 结构体，注入 `context.Context`、`OutboxRepository`、`EventBus`、`logger.Logger`、`metrics.Registry`。  
3. 编写 `Start` / `Stop` 方法，支持优雅停机。  
4. 实现批量处理逻辑与退避策略工具函数。  
5. 添加事件反序列化（从 `payload` JSON 构建具体 `Event`）。若无法识别事件类型，记失败并跳过。  
6. 集成 Plan 218 logger 与 Prometheus 指标。

### 4.2 测试

- **单元测试**：Mock `OutboxRepository` 与 `EventBus`，覆盖成功、失败、重试、跳过等分支。  
- **集成测试**（需要 Docker 数据库 + real eventbus）：  
  1. 启动测试数据库（使用 Plan 221 基座）。  
  2. 插入示例 Outbox 记录。  
  3. 运行 Dispatcher 并等待事件发布；断言状态更新与 Publish 调用次数。  
  4. 模拟 Publish 失败，验证重试与退避。  
  5. 验证并发场景：多个 Dispatcher 不会重复发布（使用 `SELECT ... FOR UPDATE SKIP LOCKED`）。

### 4.3 运行集成

1. 在 `cmd/hrms-server/command/main.go` 中注入 Dispatcher：  
   - 服务启动时创建实例并以 goroutine 运行；  
   - 服务关闭时调用 `Stop`。  
2. 将中继配置纳入 `config/` 包，支持 YAML/环境变量。  
3. 更新 Makefile：添加 `make run-outbox-dispatcher`（仅用于本地调试）。  
4. 更新 CI：在 Integration 测试阶段启动 Dispatcher，验证指标与日志。

### 4.4 文档与运维

- 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，新增 Outbox Dispatcher 配置说明。  
- 在 `docs/architecture/modular-monolith-design.md` 添加事件流示意图。  
- 输出运维指南：如何调整轮询频率、如何查看指标、如何排查失败事件。

---

## 5. 验收标准

```bash
# 单元测试（含 race 检查）
go test -v -race ./cmd/hrms-server/internal/outbox/...

# 集成测试（依赖 Docker 数据库）
make test-db-up
go test -v -tags=integration ./cmd/hrms-server/internal/outbox/...
make test-db-down

# 观察指标
curl http://localhost:9090/metrics | grep outbox_dispatch_
```

全部测试通过，指标与日志符合预期，且在命令服务停止时 Dispatcher 能优雅退出，即视为验收通过。

---

## 6. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 查询锁竞争导致延迟 | 中 | 中 | 使用 `FOR UPDATE SKIP LOCKED`，控制批量大小 |
| 未识别的事件类型 | 中 | 低 | 记录错误并跳过，同时触发告警 |
| 退避参数配置错误 | 中 | 低 | 提供默认值与参数校验，文档说明调优范围 |
| Dispatcher 异常退出 | 高 | 低 | 在 main 中增加 panic recover 与健康检查 |
| 重试放大数据库压力 | 中 | 中 | 指标监控重试总量，超过阈值时可临时扩展间隔 |

---

## 7. 交付清单

- `cmd/hrms-server/internal/outbox/dispatcher.go` 等核心代码  
- 单元与集成测试  
- 监控指标（Prometheus）与结构化日志  
- 配置项文档 & 运维指南  
- 对应代码接入（命令服务启动/停止钩子、Makefile、CI）

---

**维护者**: Codex（AI 助手）  
**最后更新**: 2025-11-04  
**计划完成日期**: W3-D3 (2025-11-12)

