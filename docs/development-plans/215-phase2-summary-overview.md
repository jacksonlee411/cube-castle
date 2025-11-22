# Phase2 基础设施建设 - 全景概览

**文档编号**: 215-Summary
**标题**: Phase2 实施方案汇总与执行指南
**创建日期**: 2025-11-04
**分支**: `feature/205-phase2-infrastructure`
**版本**: v1.0

---

## 概述

本文档总结了 Phase2（基础设施建设）的 7 个递进实施方案（Plan 216-222），提供完整的执行路线图和关键的协调信息。

---

## 1. Phase2 全景图

### 1.1 七大方案总览

```
┌─────────────────────────────────────────────────────────────────┐
│                    Phase2 基础设施建设                          │
│                    Week 3-4 (Day 12-18)                         │
└─────────────────────────────────────────────────────────────────┘

        基础设施层 (Plan 216-218)
        ┌────────┬─────────┬─────────┐
        ▼        ▼         ▼         ▼
      216      217       218    (并行)
    eventbus database  logger

        ↓ 依赖关系

    模块重构与验证 (Plan 219-222)
        ┌────────┬─────────┬──────────┐
        ▼        ▼         ▼          ▼
      219      220       221        222
    重构组织  模板文档  Docker测试  验证更新

        ↓
    Phase2 完成！✅
```

### 1.2 方案详细表

| 计划编号 | 名称 | 工作内容 | 完成日期 | 前置依赖 | 产出 | 负责人 |
|---------|------|--------|---------|---------|------|--------|
| **216** | eventbus 实现 | 事件总线接口与内存实现 | W3-D1 ✅ (2025-11-03) | Phase1 | pkg/eventbus/ | 基础设施 |
| **217** | database 实现 | 连接池、事务、outbox | W3-D2 | Plan210 | pkg/database/、20251107090000_create_outbox_events.sql | 基础设施（✅ 2025-11-04） |
| **217B** | outbox 中继服务 | 实现 outbox→eventbus 中继 | W3-D3 | 217 | cmd/hrms-server/internal/outbox | 基础设施 |
| **218** | logger 实现 | 结构化日志、Prometheus | W3-D2 | - | pkg/logger/ | 基础设施 |
| **219** | 重构 org 模块 | 标准化目录、基础设施集成 | W3-D4-5 | 216-218 | organization/ | 架构师 |
| **220** | 模块模板文档 | 标准指南、检查清单、样本 | W4-D2 | 219 | 开发文档 | 架构师 |
| **221** | Docker 测试基座 | Compose 配置、脚本、CI | W4-D3 ✅ (2025-11-15) | 217, 210 | 测试基座与验收日志 | QA |
| **222** | 验证与更新 | 测试验证、文档更新、报告 | W4-D4-5 | 219-221 | 验收报告 | QA |

---

## 2. 关键路径与依赖关系

### 2.1 执行关键路径

```
严格顺序：
216 → 217 → 218 → 219 → 222

可并行执行：
- 216 与 217 可部分并行（后期集成）
- 220 与 221 可与 219 并行进行
- 220 和 221 都在 219 完成后启动

最短执行周期：5 个工作日
实际计划：10 个工作日（包含质量保证）
```

### 2.2 明确的前置条件

| 方案 | 前置条件 | 说明 |
|------|--------|------|
| 216 | Phase1 完成 | 需要统一的 go.mod 基础 |
| 217 | Plan 210 完成 | 需要迁移脚本和 Atlas 配置 |
| 218 | 无 | 独立实现 |
| 219 | 216-218 完成 | 需要注入基础设施 |
| 220 | 219 完成 | 基于 organization 重构经验 |
| 221 | 217, 210 完成 | 需要数据库和迁移脚本 |
| 222 | 219-221 完成 | 验证所有前置工作 |

---

## 3. 每个方案的核心内容

### Plan 216: `pkg/eventbus/` 事件总线

**当前状态**: ✅ 已完成（2025-11-03），`go test`/`-race`/`-cover`/`go vet` 全部通过，覆盖率 98.1%。

**交付成果**:
- Event、EventBus、EventHandler 接口定义
- MemoryEventBus 内存实现
- 完整的单元测试（覆盖率 > 80%）

**关键代码**:
- eventbus.go（接口）
- memory_eventbus.go（实现）
- eventbus_test.go（测试）

**验收标准**:
- ✅ Subscribe/Publish 功能正常
- ✅ 并发安全（使用 RWMutex）
- ✅ 错误聚合：Publish 在处理器失败时返回详细的 `AggregatePublishError`
- ✅ 成功/失败/无订阅者计数与延迟指标可通过 `MetricsRecorder` 输出

---

### Plan 217: `pkg/database/` 数据库层

**交付成果**:
- 连接池管理（标准参数：MaxOpenConns=25）
- 事务支持（WithTx）
- 事务性发件箱接口（Outbox）

**关键代码**:
- connection.go（连接管理）
- transaction.go（事务支持）
- outbox.go（发件箱）

**验收标准**:
- ✅ 连接池配置正确
- ✅ 事务 up/down 循环正常
- ✅ Outbox 事件保存和查询

---

### Plan 217B: `outbox dispatcher` 事务性发件箱中继

**当前状态**: ✅ 已完成（2025-11-05），单元与集成测试通过，命令服务已接入 dispatcher。

**交付成果**:
- `cmd/hrms-server/internal/outbox/dispatcher.go`：定时扫描 `outbox` 表并调用 `eventbus.Publish`
- 与数据库层共享的 `OutboxRepository` 实现（重用 Plan 217 接口）
- 中继的运行参数（轮询间隔、批量大小、重试策略）
- 单元测试与集成测试（验证提交失败不会发布事件）

**关键点**:
- 在事务提交成功后由中继进程异步发布事件，避免在事务内直接 `go Publish(...)`
- 支持幂等（基于 `event_id`）与失败重试（递增 `retry_count`）
- 与 logger、metrics 集成，暴露成功/失败指标

**验收标准**:
- ✅ 事务提交失败时不会发布事件
- ✅ 已发布事件在 `outbox` 中标记 `published=true`
- ✅ 重试策略在连续失败后退避并记录报警
- ✅ 单元 & 集成测试覆盖率 > 80%

**详细文档**: `docs/development-plans/217B-outbox-dispatcher-plan.md`

---

### Plan 218: `pkg/logger/` 日志系统

**交付成果**:
- 结构化日志记录器
- JSON 格式输出
- 日志级别控制

**关键代码**:
- logger.go（核心实现）
- formatter.go（格式化）
- metrics.go（Prometheus 指标）

**验收标准**:
- ✅ JSON 输出格式正确
- ✅ 日志级别控制有效
- ✅ Prometheus 指标暴露

---

### Plan 219: `internal/organization/` 重构

**交付成果**:
- 标准模块目录结构
- 模块公开接口（api.go）
- 基础设施集成（eventbus、database、logger）

**目标结构**:
```
internal/organization/
├── api.go
├── internal/
│   ├── domain/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   └── resolver/
└── README.md
```

**验收标准**:
- ✅ 功能等同性（100%）
- ✅ 测试覆盖率 > 80%
- ✅ 无性能退化

---

### Plan 220: 模块开发模板文档

**交付成果**:
- 模块开发完整指南（> 3000 字）
- 5+ 个代码示例
- 3+ 个检查清单

**关键章节**:
1. 模块基础知识
2. 标准模块结构
3. 数据访问层规范（sqlc）
4. 事务性发件箱集成
5. Docker 集成测试
6. 测试规范
7. API 契约规范
8. 质量检查清单

**验收标准**:
- ✅ 内容准确无误
- ✅ 示例代码可编译
- ✅ 检查清单实用

---

### Plan 221: Docker 集成测试基座

**交付成果**:
- docker-compose.test.yml 配置
- 集成测试启动脚本
- Makefile 目标更新
- CI/CD 工作流配置

**关键文件**:
- docker-compose.test.yml
- scripts/run-integration-tests.sh
- .github/workflows/integration-test.yml
- Makefile 新增 `make test-db` 等目标

**验收标准**:
- ✅ 预拉取镜像后冷启动 < 10s（首轮需先执行镜像预拉取脚本）
- ✅ Goose 迁移 up/down 循环通过
- ✅ 集成测试可正常运行
- ✅ 验收证据已落盘：`logs/plan221/integration-run-*.log`

---

### Plan 222: 验证与文档更新（✅ 已关闭）

> 2025-11-23 更新：Phase2 最终验收证据已经沉淀，Plan 222 正式关单并迁移至归档；覆盖率/契约/性能补位转交 Plan 222A-D 及 Plan 255/256，不再在本计划内继续登记或推进测试覆盖。

**交付成果**:
- 完整的验收测试报告
- 项目文档全面更新
- Phase2 执行验收报告

**验证范围**:
1. 单元测试（覆盖率 > 80%）
2. 集成测试（Docker 环境）
3. REST API 回归测试
4. GraphQL 查询回归测试
5. E2E 端到端流程
6. 性能基准测试

**文档更新**:
- README.md
- DEVELOPER-QUICK-REFERENCE.md
- IMPLEMENTATION-INVENTORY.md
- modular-monolith-design.md

---

## 3.x 进度补充（2025-11-15 · 2025-11-23 更新）

- Plan 221 已完成（本地验收），证据：`logs/plan221/integration-run-*.log`
- Plan 222 阶段性通过项（2025-11-23 已正式关闭；以下记录作为历史背景）：
  - REST/GraphQL 核心路径可用：`logs/plan222/create-response-*.json`、`logs/plan222/graphql-query-*.json`
  - E2E 烟测（Chromium/Firefox 各 1 轮）通过
  - 健康/JWKS 验证：`logs/plan222/health-*.json`、`logs/plan222/jwks-*.json`
  - 覆盖率补位进行中：`logs/plan222/coverage-org-*.{out,txt,html}`

---

## 4. 时间规划与协调

### 4.1 周一至周二 (Week 3)

**Day 12 (W3-D1)**：
- 开始 Plan 216（eventbus）
- 完成目标：eventbus 接口定义、基础实现

**Day 13 (W3-D2)**：
- 继续 Plan 216（完成单元测试）
- 推进 Plan 217（database）与 Plan 218（logger）基础实现
- 完成目标：216 完成，217-218 核心结构完成

**Day 14 (W3-D3)**：
- 完成 Plan 217（database）
- 启动并完成 Plan 217B（outbox dispatcher）
- 开始 Plan 219（organization 重构）预备工作
- 完成目标：数据库层与事件中继全部就绪

**Day 15 (W3-D4)**：
- 深入 Plan 219（organization 重构前半部分）
- 启动 Plan 220（模板文档）资料整理
- 完成目标：organization 重构过半，模板文档框架确定

**Day 16 (W3-D5)**：
- 完成 Plan 219（organization 重构）
- 继续 Plan 220（模块开发模板草稿）
- 完成目标：organization 重构完成，模板文档进入评审稿

### 4.2 周一至周五 (Week 4)

**Day 17 (W4-D1)**：
- 完成 Plan 220（模板文档定稿）
- 启动 Plan 221（Docker 测试基座）环境预拉取
- 完成目标：模块模板交付，测试环境准备就绪

**Day 18 (W4-D2)**：
- 深入 Plan 221（Docker 测试基座），完成脚本与 CI 配置
- 开始 Plan 222（验证工作前期准备）
- 完成目标：221 完成 70%，验证用例清单确认

**Day 19 (W4-D3)**：
- 完成 Plan 221（Docker 测试基座）
- 执行 Plan 222（单元、集成、回归测试）
- 完成目标：Docker 基座可复用，测试覆盖达标

**Day 20-21 (W4-D4~D5)**：
- 继续 Plan 222（E2E、性能、文档更新）
- 输出 Phase2 验收报告与 README/指南更新
- 完成目标：Phase2 全部完成并提交验收材料

### 4.3 并行工作安排

为了加快交付速度，可以考虑以下并行安排：

```
W3-D1：Plan 216 (1人)
W3-D2：Plan 216 (1人) + Plan 217 (1人) + Plan 218 (1人)
W3-D3：Plan 217 (1人) + Plan 217B (1人)
W3-D4：Plan 219 (2人) + Plan 220 準備 (1人)
W3-D5：Plan 219 (2人) + Plan 220 (1人) + Plan 221 預拉取 (1人)
W4-D1：Plan 219 (1人) + Plan 220 (1人) + Plan 221 (1人)
W4-D2：Plan 221 (1人) + Plan 222 準備 (1人)
W4-D3：Plan 221 (1人) + Plan 222 測試 (2人)
W4-D4~D5：Plan 222 測試與文檔 (2人)
```

---

## 5. 团队分工建议

### 5.1 基础设施团队 (2-3 人)

**职责**:
- Plan 216: eventbus 实现
- Plan 217: database 层实现
- Plan 217B: outbox dispatcher 中继
- Plan 218: logger 实现

**所需技能**:
- Go 编程
- 并发编程（RWMutex、goroutine）
- 数据库连接管理
- 单元测试

### 5.2 架构师 (1 人)

**职责**:
- Plan 219: organization 重构指导
- Plan 220: 模板文档编写
- 整体质量把控

**所需技能**:
- 系统设计
- Go 最佳实践
- 文档编写

### 5.3 QA/测试团队 (2 人)

**职责**:
- Plan 221: Docker 测试基座
- Plan 222: 完整的验证测试

**所需技能**:
- Docker/Compose
- Go 测试框架
- 测试计划编写

### 5.4 后端开发 (1-2 人)

**职责**:
- Plan 219: organization 重构实施
- 测试用例编写

---

## 6. 质量保证清单

### 6.1 代码质量

- [ ] `go fmt ./...` 全部通过
- [ ] `go vet ./...` 无警告
- [ ] `go test -race ./...` 无 race condition
- [ ] 单元测试覆盖率 > 80%
- [ ] 代码审查完成

### 6.2 功能验收

- [ ] 基础设施包功能完整
- [ ] organization 模块功能等同
- [ ] REST API 回归通过
- [ ] GraphQL 查询回归通过
- [ ] 事务性发件箱正常工作

### 6.3 性能验收

- [ ] 查询延迟无增长
- [ ] 并发性能达标
- [ ] 内存使用稳定
- [ ] 预拉取镜像后的 Docker 启动 < 10s

### 6.4 文档验收

- [ ] 所有计划文档完整
- [ ] 代码注释清晰
- [ ] 模板文档实用
- [ ] 项目文档更新

---

## 7. 风险与应对

### 7.1 关键风险

| 风险 | 影响 | 概率 | 应对 |
|------|------|------|------|
| 基础设施设计不当 | 高 | 中 | 提前评审，确保可扩展性 |
| organization 重构破裂 | 高 | 中 | 充分回归测试 |
| Docker 环境不稳定 | 中 | 低 | 固化镜像版本，CI 预跑 |
| 时间超期 | 中 | 低 | 充分的并行执行 |

### 7.2 应急措施

- **如果基础设施出现问题**：在分支中修复，不影响 main
- **如果 organization 重构失败**：回滚至 Phase1 版本，重新规划
- **如果 Docker 不稳定**：重新拉取镜像、清理残留容器与数据卷，并确认宿主未占用标准端口后再启动 Docker Compose
- **如果超期**：优先完成关键路径（216-217-219-222），其他任务推迟

---

## 8. 后续接入信息

### 8.1 Phase3 的准备

Phase2 完成后，Phase3（workforce 模块开发）可以启动：

- 参考 Plan 220 的模块开发模板
- 使用 Plan 216-218 的基础设施
- 使用 Plan 221 的 Docker 测试环境
- 参考 Plan 219 的重构经验

### 8.2 信息传递

- 定期更新 `215-phase2-execution-log.md`（进度追踪）
- 每周五更新本概览文档
- 每个方案完成时发送进度通知

---

## 9. 相关文档导航

```
核心规划文档：
├── 203-hrms-module-division-plan.md       → 模块化架构
├── 204-HRMS-Implementation-Roadmap.md     → 实施路线图
├── 205-phase2-core-hr-modules.md          → Phase 2-4 总体规划

Phase2 详细方案：
├── 215-phase2-execution-log.md            → 执行进度追踪
├── 216-eventbus-implementation-plan.md    → 事件总线
├── 217-database-layer-implementation.md   → 数据库层
├── 217B-outbox-dispatcher-plan.md         → Outbox 中继
├── 218-logger-system-implementation.md    → 日志系统
├── 219-organization-restructuring.md      → 模块重构
├── 220-module-template-documentation.md   → 开发模板
├── 221-docker-integration-testing.md      → 测试基座
└── ../archive/development-plans/222-organization-verification.md       → 验证报告（已归档）

基础建设文档：
├── 210-database-baseline-reset-plan.md    → 数据库基线（已完成）
├── 211-phase1-module-unification-plan.md  → 模块统一（已完成）
└── 06-integrated-teams-progress-log.md    → 启动指导
```

---

## 10. 最后的话

Phase2 是 HRMS 系统从"能用"到"可维护"的关键转折。通过建立坚实的基础设施，我们为后续的模块化开发奠定了基础。

**核心价值**:
- ✅ 事件驱动架构（eventbus）支持模块解耦
- ✅ 统一的数据库层确保可靠性和一致性
- ✅ 结构化日志提升可观测性
- ✅ 标准化模块结构加速后续开发
- ✅ Docker 测试环境保证质量

**成功标志**:
- 所有 7 个方案按时交付
- 代码质量达标
- 验收测试全部通过
- Phase3 能够顺利启动

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-15
**计划完成日期**: 2025-11-18
