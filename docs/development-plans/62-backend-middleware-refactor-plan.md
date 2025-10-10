# 62号文档：后端服务与中间件收敛计划（Phase 2）

**版本**: v1.0
**创建日期**: 2025-10-10
**维护团队**: 全栈工程师（单人执行）
**状态**: 进行中
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则
**关联计划**: 60号总体计划、61号执行计划第一阶段验收

---

## 1. 背景与目标

### 1.1 背景
- Phase 1 已完成契约与类型统一：契约脚本、Go/TS 生成器、契约快照与 CI 校验均已就绪。
- 现阶段后端主服务仍存在事务/审计逻辑分散、响应/错误结构不统一、中间件可观察性薄弱等问题。
- 61号文档的第二阶段任务要求集中治理“后端服务与中间件”，确保命令服务具备明确的事务边界、统一的错误与监控体系。

### 1.2 目标
1. 抽取共享事务与审计封装，提供双写与比对日志能力，确保 `TemporalService` 与 `OrganizationTemporalService` 行为一致。
2. 定义统一的响应与错误结构体，并允许 Dev/Operational Handler 采用统一白名单与权限策略。
3. 引入 Prometheus/Otel 中间件，建立可观测性指标，灰度验证延迟 < 200ms。
4. 确保 Phase 1 的契约成果贯穿后端实现与中间件，使得 REST/GraphQL 层展示完全一致。

---

## 2. 范围

### 2.1 涉及模块
- `cmd/organization-command-service/internal/services/temporal.go`
- `cmd/organization-command-service/internal/services/organization_temporal_service.go`
- 中间件目录 `cmd/organization-command-service/internal/middleware/`
- Dev/Operational Handler：`cmd/organization-command-service/internal/handlers/devtools.go`、`operational.go`
- 审计与仓储：`cmd/organization-command-service/internal/repository/`

### 2.2 不在范围内
- Query 服务（GraphQL）及前端客户端整治，留待后续阶段。
- Temporal 数据迁移脚本及数据库结构调整（若需，单独走数据库计划）。
- 契约协议更新（已在 Phase 1 完成，仅做引用一致性检查）。

---

## 3. 工作分解与时间线（预估 3 周）

### Week 3：事务/审计封装与双写
- [ ] 设计共享事务上下文接口（如 `TemporalTransactionContext`）。
- [ ] 在 `TemporalService` 与 `OrganizationTemporalService` 中应用该封装，保持行为一致。
- [ ] 实现双写逻辑（新旧路径同时执行），并记录比对日志（建议 `logs/temporal-doublewrite.log`）。
- [ ] 提供开关（env 或 config）控制双写开/关，默认 dev/staging 开启。
- [ ] 初步验证：运行单元测试/集成环境，确认双写日志输出。

### Week 4：统一响应/错误结构与安全策略
- [ ] 定义统一响应/错误结构（Go struct），供 Dev/Operational Handler 及 REST 层引用。
- [ ] 清理现有 Handler 中的重复 JSON 拼装逻辑，改为调用统一写法。
- [ ] 制定 Dev/Operational 白名单配置（可基于 config 或环境变量），提供权限检查函数。
- [ ] 对 Web Handler 添加测试或模拟调用，确保白名单、权限逻辑生效。
- [ ] 更新文档：说明响应格式、白名单配置方法。

### Week 5：可观察性与灰度验证
- [ ] 接入 Prometheus/Otel 中间件，输出关键指标（请求耗时、双写 diff 结果、错误率等）。
- [ ] 与已有 `scripts/quality` 工具配合，提供监控验证脚本或 README。
- [ ] 设计灰度策略（dev→staging），记录切换开关、回滚策略。
- [ ] 目标指标：双写 diff = 0、Prometheus P95 latency < 200ms。
- [ ] 编写 Phase 2 验收报告，更新 06 号推进记录和 60-execution-tracker。

---

## 4. 详细任务清单

### 4.1 事务/审计封装
- [ ] 梳理 `TemporalService` 与 `OrganizationTemporalService` 中重复的事务与审计代码。
- [ ] 设计 `shared` 封装（例如 `internal/services/temporal_transaction.go`）。
- [ ] 改造现有服务调用，提供双写（新旧路径）与日志记录。
- [ ] 增加单元测试覆盖双写开关（可 mock 出入参）。

### 4.2 统一响应/错误结构
- [ ] 梳理 Handler 中的响应逻辑，设计统一响应包装函数。
- [ ] 更新 Dev/Operational、组织相关 REST Handler 使用统一结构。
- [ ] 增加白名单/权限配置，提供 `ALLOWLIST_ENDPOINTS` 或 config。
- [ ] 编写回归测试/手册，确保响应格式与 Phase 1 契约一致。

### 4.3 Prometheus/Otel 集成
- [ ] 调研当前中间件目录，确定 Otel 插入点（建议链路开始处）。
- [ ] 暴露指标: 双写 diff 统计、请求耗时、错误计数。
- [ ] 提供 `/metrics` 或现有端点配合 Prometheus 抓取。
- [ ] 编写灰度验证步骤与回滚开关说明。

### 4.4 文档与验收
- [ ] 更新 60-execution-tracker（第二阶段进度）。
- [ ] 编写 Phase 2 验收报告草稿（计划 63 号文档）。
- [ ] 若对契约文档有引用（如响应结构），更新 `docs/reference/`。

---

## 5. 验收标准与测试计划

### 5.1 验收标准
1. **双写一致性**：双写期间日志 diff = 0（统计周期内无差异）。
2. **响应结构统一**：Dev/Operational/组织 REST 接口统一返回结构，契约测试通过。
3. **权限/白名单**：白名单配置生效，未授权访问被拒绝。
4. **可观察性指标**：Prometheus 指标包含双写 diff、请求耗时，灰度期间 P95 < 200ms。
5. **文档同步**：计划、跟踪、验收文档更新；若响应结构变化需更新 reference。

### 5.2 测试计划
- 单元测试：验证双写开关、统一响应函数。
- 集成测试：运行 `make test-integration`（若存在），验证服务功能。
- 快照测试：验证契约结构未被破坏（依赖 Phase 1 契约快照）。
- 手工测试：Dev/Operational 白名单、Prometheus 指标抓取。
- 灰度验证：在 staging 环境观测 48h，确保无高优告警。

---

## 6. 风险与缓解措施

| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| 事务封装回归 | 新封装可能引入逻辑偏差 | 保留双写开关，分阶段灰度 |
| 日志量过大 | 双写日志频率较高 | 仅在 dev/staging 开启，提供采样配置 |
| Prometheus/Otel 接入影响性能 | 新中间件可能增加延迟 | 灰度观察，必要时增加采样/禁用 |
| 权限配置错误 | 白名单误配置导致接口阻断 | 提供默认配置与回滚说明 |

---

## 7. 里程碑与交付

| 周次 | 里程碑 | 交付物 |
|------|--------|--------|
| Week 3 | 双写封装上线 | `temporal_transaction.go`、双写日志、单元测试 |
| Week 4 | 响应/权限统一 | 响应封装函数、白名单配置、Handler 改造 |
| Week 5 | 可观察性与灰度 | Prometheus 指标、灰度报告、Phase 2 验收草稿 |

---

## 8. 追踪与文档
- 计划文档：62号本文件。
- 执行跟踪：`docs/development-plans/60-execution-tracker.md`（第二阶段部分）。
- 验收报告：63号文档（待创建）。
- 日志位置：`logs/temporal-doublewrite.log`、Prometheus 指标输出端点。

---

## 9. 附录
- 参考计划：60号总体计划、61号执行计划、Plan 51-55 的质量分析报告。
- 相关脚本：`scripts/quality/architecture-validator.js`、`scripts/check-api-naming.sh`。
- 数据库/迁移：如果封装涉及到 DB 结构，需要提前准备迁移文档。

