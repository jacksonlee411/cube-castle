# 202 · CQRS 混合架构执行指引（简版）

**版本**: v2.0（简化版，仅保留路线图/框架/子任务关联/关键决策）  
**创建日期**: 2025-11-02  
**最后修改**: 2025-11-15  
**文档类型**: 决策与路线图（Execution Guide）  
**状态**: 活跃（详尽分析已从本文件移除；存档见 `docs/archive/development-plans/202-CQRS-mixed-architecture-analysis-v1.1.md`）

---

## 1. 关键决策（What/Why）
- 保留混合协议：命令=REST、查询=GraphQL。理由：REST 适合命令幂等与标准化；GraphQL 适合复杂查询与按需字段，避免端点爆炸与 N+1/over-fetching。
- 先改架构，后做工程优化：
  - 架构合流（模块化单体、单进程/单端口）优先解决一致性成本、运维复杂度、虚假扩展性与认知负担（挑战 #1/#4/#5/#6）。
  - 工程优化（契约 SSoT 生成、前端领域门面、契约漂移门禁）解决协议层成本与契约同步（挑战 #2/#3）。
- 单一数据源 PostgreSQL + 事务性发件箱（Outbox）：模块通信遵循 AGENTS“同步=依赖注入、异步=发件箱”，严禁以纯内存队列作为唯一生产通道。
- 容器端口不改映射：如有冲突，按 AGENTS 卸载宿主服务（禁止调整 compose 端口迁就宿主）。
- 契约先行：OpenAPI（REST）与 GraphQL Schema（查询）为事实来源；阶段 2 引入 SSoT 生成与 CI 门禁，杜绝手改漂移。
- 权限策略外部化：以 OpenAPI `x-scopes` 为唯一事实来源，GraphQL resolver 复用同一 PBAC 校验器，确保一致。

---

## 2. 目标架构框架（To-Be）
```
┌──────────────────────────────────────────┐
│      模块化单体（:9090）                 │
│  ┌─────────────┐  ┌──────────────────┐  │
│  │ REST 命令   │  │ GraphQL 查询     │  │
│  │ /api/v1/*   │  │ /graphql         │  │
│  └──────┬──────┘  └───────┬──────────┘  │
│         │                 │             │
│      中间件：RequestID / JWT / 权限 / 速率 / 恢复 │
│         └──────────┬───────────┬────────│
│             连接池  日志/指标   事务性发件箱     │
│              (217)   (218)       Outbox        │
└─────────────────┬────────────────────────┘
                  ↓
             PostgreSQL（单一数据源）
```

---

## 3. 路线图（Roadmap）
- 阶段 1（P0）—— 架构合流与运行时统一
  - 250 模块化单体合流（单进程/单端口，REST+GraphQL 共存）
  - 251 运行时统一（共享连接池/中间件/健康/指标）
  - 253 部署流水线简化（构建/部署/观测收敛为单服务）
  - 254 前端端点与代理收敛（单基址，不改路径前缀）
- 阶段 2（P1）—— 契约与客户端工程优化
  - 256 契约 SSoT 生成流水线（make generate-contracts + CI 门禁）
  - 257 前端领域 API 门面采纳（隔离协议细节，禁直连 client）
  - 258 契约漂移校验门禁（OpenAPI ↔ GraphQL 一致性）
- 阶段 3（可选）—— 协议策略复盘
- 259 协议统一评估（默认保持混合；仅在极端情况下评估统一到 GraphQL/REST；已于 2025-11-20 完成，详见 `../archive/development-plans/259-protocol-strategy-review.md`）

说明：阶段 1 先落架构，直接化解 4/6 核心挑战；阶段 2 完成协议层工程化，清除剩余风险；阶段 3 基于证据做策略复盘，通常不需要执行。

---

## 4. 子任务关联与验收摘录（25x）
- 250 模块化单体合流（`../archive/development-plans/250-modular-monolith-merge.md`）
  - 目标：单一二进制、单端口（9090）；`/health`、`/metrics`、`/api/v1`、`/graphql` 可用
  - 验收：功能等价、221/232/241 烟测通过；215 登记命令与日志证据
- 251 运行时统一（docs/development-plans/251-runtime-unification-health-metrics.md）
  - 目标：共享 `*sql.DB` 参数（217 标准），统一中间件/健康/指标
  - 验收：连接池/HTTP/DB 指标可观测；健康检查包含 DB 自检
- 253 部署流水线简化（docs/development-plans/253-deployment-pipeline-simplification.md）
  - 目标：Make/Workflow/Compose 收敛为单服务；日志/指标采集合并
  - 验收：CI 用时下降（登记对比）；脚本统一；215 登记
- 254 前端端点与代理收敛（docs/development-plans/254-frontend-endpoint-and-proxy-consolidation.md）
  - 目标：单基址访问 `/api/v1` 与 `/graphql`；统一认证与租户头
  - 验收：E2E 配置不变通过；ESLint 守卫通过
- 256 契约 SSoT 生成（docs/development-plans/256-contract-ssot-generation-pipeline.md）
  - 目标：SSoT 生成 OpenAPI/GraphQL；CI 拒绝手改漂移
  - 验收：`make generate-contracts` 幂等；contract-sync 工作流通过
- 257 领域 API 门面（docs/development-plans/257-frontend-domain-api-facade-adoption.md）
  - 目标：业务代码经门面调用；禁直连 fetch/axios/client
  - 验收：关键模块门面覆盖率 ≥80%；E2E/单测通过
- 258 契约漂移门禁（docs/development-plans/258-contract-drift-validation-gate.md）
  - 目标：OpenAPI ↔ GraphQL 字段/类型/描述/可空一致性校验
  - 验收：差异报告稳定；PR 门禁必过，白名单可控
- 259 协议策略复盘（`../archive/development-plans/259-protocol-strategy-review.md`）
  - 目标：基于成本/效率/团队反馈评估是否统一协议（默认保持混合）
  - 验收：复盘报告与结论；如需统一，提交后续蓝图（非本期实施）

所有执行与证据一律登记：`docs/development-plans/215-phase2-execution-log.md`。

---

## 5. 门禁与验收（Gates）
- 合流门禁：单端口/健康/指标；REST/GraphQL 行为一致；共享连接池参数符合 217 标准
- 测试门禁：221 集成测试基座（Goose up/down、`make test-db`）与 E2E 烟测（232/241/244）通过
- 契约门禁：SSoT→生成 + 漂移校验；禁止从实现反向生成契约；契约变更需通过校验与门禁
- 前端门禁：领域门面覆盖率 ≥80%；禁直连 fetch/axios/client 的 ESLint 规则启用

---

## 6. 关联文档（唯一事实来源索引）
- 执行登记：`docs/development-plans/215-phase2-execution-log.md`
- 计划分解：`../archive/development-plans/250-modular-monolith-merge.md`（已归档） + `docs/development-plans/251-259*.md`
- 架构与最佳实践：`docs/development-plans/200-Go语言ERP系统最佳实践.md`、`docs/development-plans/201-Go实践对齐分析.md`
- 测试基座：`docs/development-plans/221-docker-integration-testing.md`
- 模块模板：`docs/development-plans/220-module-template-documentation.md`
- 契约：`docs/api/openapi.yaml`、`docs/api/schema.graphql`

---

## 7. 变更记录
- v2.0（2025-11-15）：简化为执行指引；保留路线图/框架/子任务关联/关键决策；详尽分析移除（可从 Git 历史查看 v1.1）。
- v1.1（2025-11-02）：深度修正版（根因分析、改协议收益评估、阶段 2 澄清）。
