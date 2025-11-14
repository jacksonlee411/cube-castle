# 215 - Phase2 执行日志与进度跟踪

**文档编号**: 215
**标题**: Phase2 - 建立模块化结构执行日志
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v2.0（与 Plan 216-222 对齐）

---

## 最新更新（2025-11-10）

- ✅ [Plan 242 / T1] Temporal Entity Page 抽象完成：组织/职位详情入口统一迁移至 `TemporalEntityPage`，路由校验、无效提示与功能映射已记录在 `reports/plan242/naming-inventory.md#temporal-entity-page`，为后续 Timeline/类型/测试抽象提供共享基线。
- ✅ [Plan 244 / T2] Temporal Timeline & Status 抽象合入：`frontend/src/features/temporal/entity/timelineAdapter.ts` 与 `statusMeta.ts` 统一组织/职位映射，Lint 禁止回退旧命名，执行记录见 `logs/plan242/t2/`。

---

## 概述

本文档跟踪 Phase2 的实施进展（Week 3-4，Day 12-18），根据 204 号文档第二阶段的定义，工作分解为 **7 个具体实施方案**（Plan 216-222）。

**基础设施建设**:
- `pkg/eventbus/` - 事件总线（**Plan 216**）
- `pkg/database/` - 数据库共享层（**Plan 217**）
- `pkg/logger/` - 日志系统（**Plan 218**）

**数据库与迁移管理**:
- 迁移脚本回滚（Down 脚本）✅ **已完成（Plan 210）**
- Atlas 工作流配置 ✅ **已完成（Plan 210）**

**模块重构与验证**:
- 重构 `organization` 模块按新模板结构（**Plan 219**）
- 创建模块开发模板文档（**Plan 220**）
- 构建 Docker 集成测试基座（**Plan 221**）
- 验证 organization 模块正常工作（**Plan 222**）
- 更新 README 和开发指南（**Plan 222**）

---

## 阶段时间表与计划映射（Week 3-4）

| 周 | 日 | 行动项 | 计划 | 描述 | 负责人 | 状态 |
|-----|-----|--------|------|------|--------|------|
| **W3** | D1 | 2.1 | **Plan 216** | 实现 `pkg/eventbus/` 事件总线 | 基础设施 | ✅ 2025-11-03 |
| | D2 | 2.2 | **Plan 217** | 实现 `pkg/database/` 数据库层 | 基础设施 | ✅ 2025-11-04 |
| | D2 | 2.3 | **Plan 218** | 实现 `pkg/logger/` 日志系统 | 基础设施 | ✅ 2025-11-04 |
| | D3 | 2.3a | **Plan 217B** | 构建 outbox→eventbus 中继 | 基础设施 | ✅ 2025-11-05 |
| | D3 | 2.4-2.5 | Plan 210 | 迁移脚本和 Atlas 配置 | DevOps | ✅ |
| | D4-5 | 2.6 | **Plan 219** | 重构 organization 模块结构（执行窗口调整至 Day 15-17，含缓冲） | 架构师 | ✅ 2025-11-06 |
| **W4** | D1-2 | 2.7 | **Plan 220** | 创建模块开发模板文档（与 219B 并行） | 架构师 | ⏳ |
| | D2-3 | 2.8 | **Plan 221** | 构建 Docker 化集成测试基座（支撑 219E） | QA | ⏳ |
| | D3-5 | 2.9-2.10 | **Plan 222** | 验证与文档更新；预留 Day 26 缓冲 | QA/文档 | ⏳ |

---

## 进度记录

### 行动项 2.1 - 实现 `pkg/eventbus/` 事件总线 (Plan 216)

**对应计划**: **Plan 216 - eventbus-implementation-plan.md**

**计划行动**:
- [x] 定义事件总线接口（Event、EventBus、EventHandler）- `pkg/eventbus/eventbus.go`
- [x] 实现内存事件总线（MemoryEventBus + AggregatePublishError + MetricsRecorder）- `pkg/eventbus/memory_eventbus.go`
- [x] 编写单元测试（覆盖率 98.1%，覆盖成功/失败/无订阅者/并发场景）- `pkg/eventbus/eventbus_test.go`
- [x] 集成失败聚合、日志与指标记录（提供 noop Logger/Metrics，符合 Plan 217B 依赖）

**交付物**:
```
pkg/eventbus/
├── eventbus.go        # 接口定义（Event、EventBus、EventHandler）
├── memory_eventbus.go # 内存实现（并发安全）
├── error.go           # 错误定义
└── *_test.go          # 单元测试（覆盖率 > 80%）
```

**关键特性**:
- Event 接口：EventType()、AggregateID()
- EventBus 接口：Publish()、Subscribe()
- EventHandler：func(ctx, event) error 签名
- 支持多订阅者处理
- 错误处理：处理器失败不阻止其他处理器，并通过 `AggregatePublishError` 返回给调用方
- 可观测性：提供成功/失败/无订阅者/耗时指标，日志接口与 Plan 218 logger 契约一致

**技术要点**:
- 并发安全：使用 RWMutex 保护 handlers 映射
- 失败聚合：所有处理器执行完毕后返回聚合错误并记录指标
- 性能目标：发布延迟 < 1ms

**验收标准** (来自 Plan 216):
- [x] Subscribe/Publish 功能正常（`TestPublishWithSingleSubscriber`、`TestPublishWithMultipleSubscribers`）
- [x] 多订阅者都被调用（单测断言调用计数）
- [x] 并发安全（`go test -race ./pkg/eventbus` 通过）
- [x] 处理器失败时返回 `AggregatePublishError` 并包含失败明细（`TestPublishWithHandlerError`）
- [x] 指标记录（成功/失败/无订阅者/延迟）经单元测试验证（`testMetrics` 断言）
- [x] 单元测试覆盖率 > 80%（`go test -cover ./pkg/eventbus` 输出 98.1%）
- [x] 代码通过 `go fmt` 和 `go vet`（手动执行 `gofmt`、`go vet ./pkg/eventbus`）

**执行记录**:
- 2025-11-03 完成代码提交，运行 `go test ./pkg/eventbus`、`go test -race ./pkg/eventbus`、`go test -cover ./pkg/eventbus`、`go vet ./pkg/eventbus` 全部通过。
- 更新 `pkg/eventbus/README.md` 说明指标命名与 Plan 217B 集成方式。

**负责人**: 基础设施团队
**计划完成**: Day 12 (W3-D1)
**状态**: ✅ 已完成（2025-11-03）

**详细文档**: 见 `docs/development-plans/216-eventbus-implementation-plan.md`

---

### 行动项 2.2 - 实现 `pkg/database/` 数据库层 (Plan 217)

**对应计划**: **Plan 217 - database-layer-implementation.md**

**计划行动**:
- [x] 创建数据库连接管理（连接池配置）——`pkg/database/connection.go` + 单测
- [x] 实现事务支持（Transaction 包装）——`pkg/database/transaction.go`
- [x] 实现事务性发件箱（outbox）表接口——`pkg/database/outbox.go`、`database/migrations/20251107090000_create_outbox_events.sql`
- [x] 编写单元测试与集成测试——`pkg/database/*_test.go`、`tests/integration/migration_roundtrip_test.go`

**交付物**:
```
pkg/database/
├── connection.go      # 连接池管理（标准参数）
├── transaction.go     # 事务支持（WithTx）
├── outbox.go          # 事务性发件箱接口
├── metrics.go         # Prometheus 指标
├── error.go           # 错误定义
└── *_test.go          # 单元 & 集成测试
```

**关键参数** (硬编码标准配置):
- MaxOpenConns: 25（防止连接溢出）
- MaxIdleConns: 5（连接复用）
- ConnMaxIdleTime: 5 分钟（定期刷新）
- ConnMaxLifetime: 30 分钟（周期替换）

**关键接口**:
- `NewDatabase(dsn)` - 创建连接
- `WithTx(ctx, fn)` - 事务支持
- `GetUnpublishedEvents()` - 获取未发布事件
- `MarkEventPublished()` - 标记事件已发布
- `IncrementRetryCount()` - 增加重试计数

**事务性发件箱** (Outbox 模式):
- OutboxEvent 结构：event_id、aggregate_id、event_type、payload 等
- SaveOutboxEvent() - 在事务内保存事件
- 用于保证跨模块操作的最终一致性
- 与 Plan 216 eventbus 配合使用

**Prometheus 指标**:
- db_connections_in_use - 当前活动连接
- db_connections_idle - 空闲连接
- db_query_duration_seconds - 查询延迟直方图

**验收标准** (来自 Plan 217):
- [x] 连接池配置正确（MaxOpenConns=25）
- [x] 事务创建、提交、回滚正常
- [x] Outbox 事件保存和查询成功
- [x] 单元 & 集成测试通过（`go test ./pkg/database -cover` -> 82.1%，`go test ./tests/integration/migration_roundtrip_test.go`）
- [x] 无 race condition（关键场景由单元测试覆盖）

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D2)
**状态**: ✅ 已完成（2025-11-04）

**详细文档**: 见 `docs/development-plans/217-database-layer-implementation.md`

---

### 行动项 2.3a - 构建 outbox dispatcher 中继 (Plan 217B) ✅ 已完成

**对应计划**: **Plan 217B - outbox dispatcher 中继**

**计划行动**:
- [ ] 实现独立的中继组件，定时查询 `outbox` 表未发布事件
- [ ] 调用 Plan 217 提供的 `OutboxRepository` 接口，批量取回事件
- [ ] 调用 Plan 216 的 `eventbus.Publish` 发布事件，并在成功后标记 `published=true`
- [ ] 为失败事件增加 `retry_count` 并记录结构化日志
- [ ] 实现指数退避或最小间隔，避免频繁轮询
- [ ] 编写单元与集成测试，覆盖事务失败不发布事件的场景

**交付物**:
```
cmd/hrms-server/internal/outbox/
├── dispatcher.go        # 中继循环（可配置间隔/批量大小）
├── dispatcher_config.go # 配置与默认值
├── metrics.go           # 成功/失败/重试指标
├── dispatcher_test.go   # 单元测试
└── integration_test.go  # 集成测试（依赖 Plan 217/216）
```

**运行要点**:
- 默认轮询间隔 5s，可通过环境变量覆盖
- 每批拉取 50 条事件，发布成功后调用 `MarkPublished`
- 发布失败时调用 `IncrementRetryCount` 并根据重试次数决定退避时间
- 与 Plan 218 logger 集成，记录成功、失败与重试明细
- 暴露 Prometheus 指标：`outbox_dispatch_success_total`、`outbox_dispatch_failure_total`、`outbox_dispatch_retry_total`

**验收标准**:
- [ ] 事务提交失败时不会发布事件（集成测试覆盖）
- [ ] 成功发布的事件在 outbox 表中被标记为 `published=true`
- [ ] 连续失败的事件会增加 `retry_count` 并进入退避队列
- [ ] 中继可通过上下文或信号安全停止
- [ ] 单元与集成测试覆盖率 > 80%

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D3)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/217B-outbox-dispatcher-plan.md`

---

### 行动项 2.3 - 实现 `pkg/logger/` 日志系统 (Plan 218)

**对应计划**: **Plan 218 - logger-system-implementation.md**

**计划行动**:
- [x] 创建结构化日志记录器
- [x] 实现日志级别控制（Debug, Info, Warn, Error）
- [x] 集成性能监控（响应时间、数据库查询统计）
- [x] 编写单元测试

**交付物**:
```
pkg/logger/
├── logger.go          # Logger 接口和实现
├── std.go             # 标准库网桥（backward compatibility）
└── *_test.go          # 单元测试（覆盖率 > 80%）
```

**Logger 接口**:
- Debug/Debugf, Info/Infof, Warn/Warnf, Error/Errorf ✅
- WithFields(map[string]interface{}) - 添加结构化字段 ✅
- JSON 输出格式（timestamp、level、message、fields、caller） ✅

**日志级别**:
- DebugLevel, InfoLevel, WarnLevel, ErrorLevel ✅
- 通过环境变量 `LOG_LEVEL` 设置 ✅
- 默认 InfoLevel ✅

**结构化输出** (JSON 格式):
```json
{
  "timestamp": "2025-11-04T10:30:45.123Z",
  "level": "INFO",
  "message": "organization created",
  "fields": {"organizationID": "org-123"},
  "caller": "organization/service.go:42"
}
```

**标准库网桥**:
- std.go 提供向后兼容性
- 在测试和工具场景中使用
- 不在生产代码中依赖

**验收标准** (来自 Plan 218):
- [x] Logger 接口定义完整
- [x] JSON 输出格式正确
- [x] 日志级别控制有效
- [x] WithFields() 正常工作
- [x] 单元测试覆盖率 > 80%
- [x] 所有测试通过（13 个测试用例）
- [x] 代码通过 `go fmt` 和 `go vet`

**执行记录**:
- 2025-11-04 完成代码实现，包括 Logger、std 网桥和完整的单元测试。
- 运行 `go test ./pkg/logger -v` 全部通过。
- 与 Plan 218A-E 子计划一致，logger 迁移已达生产级质量。

**负责人**: 基础设施团队
**计划完成**: Day 13 (W3-D2)
**状态**: ✅ 已完成（2025-11-04）

**详细文档**: 见 `docs/development-plans/218-logger-system-implementation.md` 及 `docs/development-plans/218C-logger-verification-report.md`、`docs/development-plans/218E-logger-rollout-closure.md`

---

### 行动项 2.4-2.5 - 迁移脚本与 Atlas 配置

**对应计划**: Plan 210（已完成）

**状态**: ✅ **已完成（Plan 210，2025-11-06）**

**已完成的工作**:
- ✅ 为所有迁移文件补齐 `-- +goose Down` 回滚脚本
- ✅ 配置 Atlas `atlas.hcl` 和 `goose.yaml`
- ✅ 基线迁移脚本 `20251106000000_base_schema.sql` 已部署
- ✅ up/down 循环验证通过

**证据**: `docs/archive/development-plans/210-execution-report-20251106.md`

此工作为 Plan 221 (Docker 集成测试) 和 Plan 222 (验证) 的前置条件。

---

### 行动项 2.6 - 重构 `organization` 模块结构 (Plan 219)

**对应计划**: **Plan 219 - organization-restructuring.md**

**计划行动**:
- [x] 按新模板重组 organization 模块代码
- [x] 定义模块公开接口（api.go）
- [x] 整理 internal/ 目录结构（service、repository、handler、resolver、domain）
- [x] 确保模块边界清晰
- [x] 集成基础设施（Plan 216-218）

**目标结构**:
```
internal/organization/
├── api.go                         # 公开接口定义
├── internal/
│   ├── domain/
│   │   ├── organization.go        # 域模型
│   │   ├── department.go
│   │   ├── position.go
│   │   ├── events.go              # 域事件定义
│   │   └── constants.go
│   ├── repository/
│   │   ├── organization_repository.go
│   │   ├── department_repository.go
│   │   ├── position_repository.go
│   │   └── *_test.go
│   ├── service/
│   │   ├── organization_service.go
│   │   ├── department_service.go
│   │   ├── position_service.go
│   │   └── *_test.go
│   ├── handler/                   # REST 处理器
│   │   ├── organization_handler.go
│   │   └── *_test.go
│   ├── resolver/                  # GraphQL 解析器
│   │   ├── organization_resolver.go
│   │   └── *_test.go
│   └── README.md                  # 内部说明
└── README.md                      # 模块说明
```

**关键工作**:
1. **api.go - 模块公开接口**
   - OrganizationAPI interface（所有公开方法）
   - 其他模块仅能依赖 api.go，不能导入 internal/

2. **基础设施集成**
   - Service 层注入 eventbus (Plan 216)
   - Service 层注入 database (Plan 217)
   - 使用 logger (Plan 218) 记录操作
   - 使用 eventbus 发布组织变更事件
   - 使用 database 的 WithTx 管理事务

3. **事务性发件箱**
   - 在 service 中创建新实体时，同一事务内保存 outbox 事件
   - 异步发布事件给 eventbus

4. **功能等同性**
   - 重构后行为必须与重构前完全相同
   - 所有 API 端点签名不变
   - 数据查询结果一致

**实施步骤** (来自 Plan 219):
1. 分析与准备：审视现有代码，梳理接口和依赖
2. 目录重构：创建新结构，重新分类代码
3. 基础设施集成：注入 eventbus、database、logger
4. 测试与验证：运行回归测试，确保功能等同

**验收标准** (来自 Plan 219):
- [x] 模块按新模板重构完成
- [x] api.go 公开接口清晰
- [x] internal/ 目录结构符合规范
- [x] 无循环依赖
- [x] 功能等同（100%）
- [x] 单元测试覆盖率 > 80%
- [x] 性能无退化

**实际完成情况（2025-11-06）**:
- 目录已集中到 `internal/organization/*`，README 记录审计、handler、repository、resolver、service、scheduler、validator、dto 等聚合边界，成为组织域的唯一事实来源 `internal/organization/README.md:3`, `internal/organization/README.md:21`.
- `internal/organization/api.go` 构建统一的 `CommandModule`，一次性注入数据库、审计记录器、级联服务、职位/任职/职位目录仓储、事务性发件箱以及调度服务，命令侧只需依赖公开 API 即可接入 `internal/organization/api.go:28`, `internal/organization/api.go:119`.
- 查询侧通过 `AssignmentQueryFacade` 复用同一个仓储并接入 Redis 缓存，缓存键规范、TTL 以及刷新逻辑集中在 `internal/organization/query_facade.go:28`, `internal/organization/query_facade.go:136`，并由 `internal/organization/query_facade_test.go:42` 覆盖缓存命中与回源场景。
- README 记录的 219E 验收脚本（组织生命周期冒烟、REST 性能基准）已同步到 `scripts/e2e/org-lifecycle-smoke.sh` 与 `scripts/perf/rest-benchmark.sh`，输出日志位于 `logs/219E/*` 以支撑 Plan 222 的后续验证 `internal/organization/README.md:56`.


**负责人**: 架构师 + 后端团队
**计划完成**: Day 14-15 (W3-D4-5)
**状态**: ✅ 已完成（2025-11-06）

**详细文档**: 见 `docs/development-plans/219-organization-restructuring.md`

---

### 行动项 2.7 - 创建模块开发模板文档 (Plan 220)

**对应计划**: **Plan 220 - module-template-documentation.md**

**计划行动**:
- [ ] 编写完整的模块开发指南（> 3000 字）
- [ ] 基于 organization 重构经验提供样本代码
- [ ] 文档化 sqlc 使用规范
- [ ] 文档化事务性发件箱集成规范
- [ ] 文档化 Docker 集成测试规范
- [ ] 创建各阶段检查清单

**交付物**:
```
docs/development-guides/
├── module-development-template.md  # 主指南文档
├── examples/
│   └── organization/               # 参考实现代码
└── checklists/
    ├── module-structure-checklist.md
    ├── api-contract-checklist.md
    ├── testing-checklist.md
    └── deployment-checklist.md
```

**主文档章节** (来自 Plan 220):
1. 模块基础知识 - Bounded Context、DDD
2. 模块结构模板 - 标准目录结构说明
3. 数据访问层规范 - sqlc 使用、repository 模式
4. 事务性发件箱集成 - outbox 模式、可靠性保证
5. Docker 集成测试 - 容器化测试、Goose 迁移
6. 测试规范 - 单元、集成、E2E 测试
7. API 契约规范 - REST 命名、GraphQL schema
8. 质量检查清单 - 代码质量、安全、性能

**目标受众**:
- 后端开发者（新模块实现者）
- 新团队成员（理解项目架构）
- QA 工程师（了解测试策略）

**与其他计划的关系**:
- 基于 Plan 219 (organization 重构)
- 为 Phase 3 (workforce 模块) 提供参考
- 使用 Plan 216-218 的基础设施
- 引用 Plan 221 的 Docker 测试规范

**验收标准** (来自 Plan 220):
- [ ] 文档完整（> 3000 字）
- [ ] 包含 5+ 个代码示例
- [ ] 示例代码可编译且正确
- [ ] 包含 3 个以上检查清单
- [ ] 内容与 organization 模块对齐
- [ ] 新模块开发者可独立参考

**负责人**: 架构师 + 文档支持
**计划完成**: Day 17 (W4-D1-2)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/220-module-template-documentation.md`

---

### 行动项 2.8 - 构建 Docker 化集成测试基座 (Plan 221)

**对应计划**: **Plan 221 - docker-integration-testing.md**

**计划行动**:
- [ ] 创建 `docker-compose.test.yml`（PostgreSQL）
- [ ] 编写集成测试启动脚本
- [ ] 验证 Goose up/down 流程
- [ ] 创建测试数据初始化脚本
- [ ] 更新 Makefile 和 CI/CD 配置

**交付物**:
```
├── docker-compose.test.yml              # Docker 配置
├── scripts/test/
│   ├── init-db.sql                      # 初始化脚本
│   └── run-integration-tests.sh          # 测试启动脚本
├── Makefile（更新）
│   ├── make test-db-up
│   ├── make test-db-down
│   ├── make test-db
│   ├── make test-db-logs
│   └── make test-db-psql
└── .github/workflows/
    └── integration-test.yml             # CI 工作流
```

**Docker 配置** (来自 Plan 221):
- PostgreSQL 15 Alpine 镜像
- 自动初始化数据库
- 端口映射：5432:5432（保持标准端口；如被占用需清理宿主冲突服务后再启动）
- 健康检查：pg_isready
- 卷挂载：迁移脚本、初始化脚本

**集成测试流程**:
1. 启动 Docker 容器
2. 等待数据库就绪
3. 运行 Goose 迁移 (up)
4. 执行 Go 集成测试
5. 验证回滚 (down)
6. 清理容器

**Makefile 目标**:
- `make test-db-up` - 启动测试数据库
- `make test-db-down` - 停止测试数据库
- `make test-db` - 完整的集成测试流程
- `make test-db-logs` - 查看数据库日志
- `make test-db-psql` - 连接到测试数据库

**CI/CD 集成** (.github/workflows/integration-test.yml):
- 在 GitHub Actions 中运行集成测试
- 使用 services 启动 PostgreSQL
- 运行 Goose 迁移
- 执行集成测试并上传覆盖率

**验收标准** (来自 Plan 221):
- [ ] 预拉取镜像后的 Docker 启动 < 10s
- [ ] 数据库就绪时间 < 15s
- [ ] Goose up/down 循环通过
- [ ] 集成测试可正常运行
- [ ] 多次运行结果一致
- [ ] 无端口冲突

**负责人**: QA + DevOps
**计划完成**: Day 18-19 (W4-D2-3)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/221-docker-integration-testing.md`

---

### 行动项 2.9 - 验证 organization 模块正常工作 (Plan 222)

**对应计划**: **Plan 222 - organization-verification.md**

**计划行动**:
- [ ] 单元测试 organization 服务（覆盖率 > 80%）
- [ ] 集成测试 organization 与数据库交互
- [ ] 验证 Goose up/down + Docker 测试流程正常
- [ ] 执行 REST API 回归测试
- [ ] 执行 GraphQL 查询回归测试

**验证范围** (来自 Plan 222):

**1. 单元测试验证**
```bash
go test -v -race -coverprofile=coverage.out ./internal/organization/...
```
- [ ] 所有单元测试通过
- [ ] 测试覆盖率 > 80%
- [ ] 无 race condition

**2. 集成测试验证**
```bash
make test-db-up
go test -v -tags=integration ./cmd/hrms-server/...
make test-db-down
```
- [ ] 集成测试全部通过
- [ ] Goose 迁移 up/down 循环通过
- [ ] 数据库状态一致

**3. REST API 回归测试**
- [ ] GET /org/organizations/{code}
- [ ] POST /org/organizations
- [ ] PUT /org/organizations/{code}
- [ ] 响应字段为 camelCase
- [ ] HTTP 状态码正确
- [ ] 错误处理一致

**4. GraphQL 查询回归测试**
- [ ] query { organizations { id code name } }
- [ ] 返回数据符合 schema
- [ ] 错误处理正确

**5. E2E 端到端流程测试**
```
1. 创建新的组织单元
2. 查询组织单元详情
3. 创建部门
4. 为部门创建职位
5. 分配员工到职位
6. 查询组织结构
7. 更新组织信息
8. 验证审计日志
```

**6. 性能基准测试**
- 单个查询：< 50ms (P99)
- 列表查询（100 条）：< 200ms (P99)
- 创建操作：< 100ms (P99)
- 并发（100 并发）：> 100 req/s

**验收标准** (来自 Plan 222):
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有测试通过（0 失败）
- [ ] REST API 回归通过
- [ ] GraphQL 查询回归通过
- [ ] E2E 端到端通过
- [ ] 性能基准达标

**负责人**: QA
**计划完成**: Day 19 (W4-D3)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/222-organization-verification.md`

---

### 行动项 2.10 - 更新 README 与开发指南 (Plan 222)

**对应计划**: **Plan 222 - organization-verification.md**（第二部分）

**计划行动**:
- [ ] 更新项目 README（新目录结构说明）
- [ ] 更新开发者速查（模块化单体工作流）
- [ ] 添加常见命令列表
- [ ] 更新 CI/CD 说明
- [ ] 更新实现清单
- [ ] 完成 Phase2 执行验收报告

**文档更新** (来自 Plan 222):

**1. README.md 更新**
- 项目结构说明（cmd/、internal/、pkg/）
- 快速开始指南
- 构建、测试、开发命令
- 模块化架构简述
- 链接到详细文档

**2. DEVELOPER-QUICK-REFERENCE.md 更新**
- 模块结构规范
- 常用命令速查
- 基础设施使用示例（eventbus、database、logger）
- 调试技巧

**3. IMPLEMENTATION-INVENTORY.md 更新**
- Phase1 状态（✅ 完成）
- Phase2 状态（✅ 完成）
- Phase3 计划状态
- 代码统计、覆盖率等指标

**4. 架构文档更新** (docs/architecture/modular-monolith-design.md)
- 当前架构状态
- 模块间通信机制
- 基础设施层说明

**5. Phase2 执行验收报告** (reports/phase2-execution-report.md)
- 执行概览
- 验收结果
- 质量指标
- 关键交付物
- 风险消除情况
- Phase3 预期

**验收标准** (来自 Plan 222):
- [ ] README 更新完整
- [ ] 开发指南更新
- [ ] 实现清单更新
- [ ] 架构文档更新
- [ ] 验收报告完成

**负责人**: 文档支持 + 架构师
**计划完成**: Day 20-21 (W4-D4-5)
**状态**: ⏳ 待启动

**详细文档**: 见 `docs/development-plans/222-organization-verification.md`

---

## 关键检查点

### 基础设施质量检查点

- [ ] `pkg/eventbus/` (Plan 216) 单元测试覆盖率 > 80%
- [ ] `pkg/database/` (Plan 217) 连接池配置正确（MaxOpenConns=25）
- [ ] `pkg/logger/` (Plan 218) 与 Prometheus 指标集成
- [x] `outbox dispatcher` (Plan 217B) 能可靠发布并记录重试指标
- [ ] 所有共享包无循环依赖
- [ ] 代码格式通过 `go fmt ./...`
- [ ] 代码通过 `go vet ./...`
- [ ] 无 race condition (`go test -race ./...`)

### 模块重构检查点 (Plan 219)

- [x] organization 模块按新模板重构完成（参考 `internal/organization/README.md:3`）
- [x] 模块公开接口清晰（api.go）`internal/organization/api.go:28`
- [x] internal/ 目录结构符合规范
- [x] 模块间无直接依赖（仅通过 interface）
- [x] CQRS 边界清晰（REST 命令、GraphQL 查询）
- [x] 基础设施正确集成（eventbus、database、logger）

### 测试与验证检查点 (Plan 221-222)

- [ ] Docker 集成测试基座可正常启动 (Plan 221)
- [ ] Goose up/down 循环验证通过 (Plan 221)
- [ ] organization 模块所有测试通过 (Plan 222)
- [ ] REST/GraphQL 端点行为一致 (Plan 222)
- [ ] E2E 端到端流程测试通过 (Plan 222)
- [ ] 性能基准测试达标 (Plan 222)

### 文档完整性检查点 (Plan 220, 222)

- [ ] 模块开发模板文档完成 (Plan 220)
- [ ] README 和开发指南更新 (Plan 222)
- [ ] 实现清单更新 (Plan 222)
- [ ] Phase2 执行验收报告完成 (Plan 222)

---

## 风险与应对

| 风险 | 影响 | 概率 | 预防措施 | 对应计划 |
|------|------|------|--------|---------|
| 事件总线设计不当 | 中 | 中 | 充分评审，确保扩展性 | Plan 216 |
| 数据库层性能问题 | 高 | 中 | 连接池参数验证，压力测试 | Plan 217 |
| organization 重构破裂 | 高 | 中 | ✅ 2025-11-06 在 feature 分支完成全量测试后合并主干 | Plan 219 |
| outbox 中继未按时完成 | 高 | 中 | ✅ 已完成（2025-11-05） | Plan 217B |
| Docker 集成测试不稳定 | 中 | 中 | 固化镜像版本，CI 预跑 | Plan 221 |
| 时间超期 | 中 | 低 | 充分的并行执行 | 整体协调 |

---

## 计划文档导航

### 7 个实施方案文档

| 计划 | 文档名 | 工作内容 | 关键交付 |
|------|--------|---------|---------|
| **Plan 216** | 216-eventbus-implementation-plan.md | pkg/eventbus/ 实现 | 事件总线接口和内存实现 |
| **Plan 217** | 217-database-layer-implementation.md | pkg/database/ 实现 | 连接池、事务、outbox |
| **Plan 217B** | 217B-outbox-dispatcher-plan.md | outbox 中继实现 | 事件发布中继、重试机制 |
| **Plan 218** | 218-logger-system-implementation.md | pkg/logger/ 实现 | 结构化日志、Prometheus |
| **Plan 219** | 219-organization-restructuring.md | organization 重构 | 标准模块结构 |
| **Plan 220** | 220-module-template-documentation.md | 模块开发指南 | 模板文档、样本代码 |
| **Plan 221** | 221-docker-integration-testing.md | Docker 测试基座 | Compose 配置、脚本 |
| **Plan 222** | 222-organization-verification.md | 验证与文档更新 | 验收报告、文档更新 |

### 相关规划文档

- `204-HRMS-Implementation-Roadmap.md` - Phase2 实施路线图（权威定义）
- `215-phase2-summary-overview.md` - Phase2 全景概览（协调中心）
- `06-integrated-teams-progress-log.md` - Phase2 启动指导
- `203-hrms-module-division-plan.md` - HRMS 模块划分蓝图

---

## 相关文档

- `204-HRMS-Implementation-Roadmap.md` - Phase2 实施路线图（权威定义）
- `06-integrated-teams-progress-log.md` - Phase2 启动指导
- `203-hrms-module-division-plan.md` - HRMS 模块划分蓝图
- `docs/api/openapi.yaml` - REST API 契约
- `docs/api/schema.graphql` - GraphQL 契约

---

## 提交记录

| 日期 | 提交 | 描述 |
|------|------|------|
| 2025-11-04 | b328bd1e | docs: correct Phase2 scope - infrastructure setup not new modules |
| 2025-11-04 | c481f189 | docs: create Phase2 implementation plans (216-222) |
| 2025-11-04 | 1b2b39b9 | docs: add Phase2 implementation summary and overview |
| 2025-11-04 | - | docs: update Phase2 execution log aligned with Plan 216-222 |

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**版本**: v2.0（与 Plan 216-222 完全对齐）
**关键更改**: 每个行动项现在明确链接到对应的计划文档编号
