# 06号文档：Phase2 启动与模块化结构建设计划（重置版）

> **更新时间**：2025-11-05  
> **当前聚焦**：Plan 217B outbox dispatcher 集成验证要求

---

## 1. 集成测试要求（Plan 217B）

- **测试目标**  
  验证命令服务在真实 Docker/PostgreSQL 环境下运行 outbox dispatcher 时，能够完成以下场景：  
  1. 从 `outbox_events` 表读取待发布事件并调用 `eventbus.Publish`。  
  2. 发布成功后将事件标记为 `published=true`，并在指标中累计成功计数。  
  3. 发布失败路径触发 `IncrementRetryCount`、按照指数退避更新 `available_at`，并维持既有事件不会丢失。  
  4. 服务关闭时 dispatcher 能够响应 context 取消，优雅退出且不遗留锁。

- **环境约束**  
  - 使用 `make test-db-up` 启动 DockerPostgreSQL 基座；运行测试前确保本地无宿主机数据库占用 5432。  
  - 事件总线可复用内存实现，需在测试中注入自定义 handler 验证发布行为。  
  - Prometheus 指标可通过本地注册器断言计数变化，无需真实抓取 `/metrics`。

- **测试形态**  
  - 在 `cmd/hrms-server/command/internal/outbox/integration_test.go` 中实现 `//go:build integration` 测试用例，执行顺序：  
    1. 预置 outbox 记录（成功/失败双路径）。  
    2. 启动 dispatcher + eventbus handler，等待处理完成。  
    3. 查询数据库断言状态字段与重试次数。  
    4. 停止 dispatcher，验证 goroutine 回收与 context 取消。  
  - CI 流程需在 integration 阶段调用 `go test -tags=integration ./cmd/hrms-server/internal/outbox`。

- **验收输出**  
  - 测试脚本：更新 Makefile/CI pipeline 加入 integration 任务。  
  - 文档：在 `docs/development-plans/217B-outbox-dispatcher-plan.md` 与 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 同步记录运行命令及注意事项。  
  - 指标/日志：附带成功失败计数截图或命令行输出，证明指标符合预期。

---

## 2. 后续行动（追踪）

| 项目 | 负责人 | 截止日期 | 状态 |
|------|--------|----------|------|
| 实现 outbox dispatcher 集成测试 | 基础设施团队 | 2025-11-06 | ✅ 完成 |
| CI 集成 `go test -tags=integration` | DevOps | 2025-11-07 | ⏳ 待启动（交由 Plan 221 集成测试基座同步） |
| 更新 217B 计划与快速参考文档 | 文档支持 | 2025-11-07 | ✅ 完成 |

> 若计划调整或测试结果有新信息，请在本文件追加记录，保持唯一事实来源。
