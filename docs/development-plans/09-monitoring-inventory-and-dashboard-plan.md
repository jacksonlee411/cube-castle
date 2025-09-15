# 09 · 监控模块盘点与首页可视化方案（开发计划）

状态: 草案（待审批）  
最后更新: 2025-09-15  
负责人: 架构/前端/后端/QA 联合小组

> 文档边界说明：本文件属于 Development Plans（计划/进展）。实现完成后请归档至 `docs/archive/development-plans/`，并在 Reference 中仅保留规范性结论与接口契约。

## 背景与目标

- 背景：命令服务中仍启用时态监控模块（TemporalMonitor），并通过运维调度器周期采集指标；目前缺少集中可视化与契约对齐。
- 目标：
  - 盘点现有监控模块与指标，识别一致性问题；
  - 在前端首页（`/dashboard`）新增“系统监控总览”页面，直观呈现监控指标与告警；
  - 契约优先：在 OpenAPI 中补齐运维端点与权限 scopes，并与 PBAC 对齐；
  - 保持“PostgreSQL 原生 CQRS”规则：业务查询走 GraphQL；系统运维（health/metrics/alerts）归类为系统管理 REST 例外路径，统一 under `/api/v1/operational/*`。

## 现状盘点（可验证事实）

### 1) 后端监控模块（命令服务 9090）

- 模块：`TemporalMonitor`（`cmd/organization-command-service/internal/services/temporal_monitor.go`）
  - 采集项（MonitoringMetrics）：
    - 基础量纲：`totalOrganizations`, `currentRecords`, `futureRecords`, `historicalRecords`
    - 一致性问题：`duplicateCurrentCount`, `missingCurrentCount`, `timelineOverlapCount`, `inconsistentFlagCount`, `orphanRecordCount`
    - 健康度：`healthScore`（0–100）, `alertLevel`（HEALTHY/WARNING/CRITICAL）, `lastCheckTime`
  - 告警规则：阈值内置（重复/缺失/重叠为 CRITICAL；标志不一致>5 和孤立>10 为 WARNING；健康分<85 WARNING）
  - 调度：`OperationalScheduler` 每 5 分钟巡检。

- 对外端点（JSON）：
  - `GET /api/v1/operational/health` — 概览（健康分、摘要、问题聚合、lastCheckTime）
  - `GET /api/v1/operational/metrics` — 详细指标（上述全量字段）
  - `GET /api/v1/operational/alerts` — 当前告警列表

- 其它相关：
  - Dev 性能端点：`GET /dev/performance-metrics`（内存、GC、goroutine、DB 连接池快照）
  - 限流统计：`GET /debug/rate-limit/stats`、`GET /debug/rate-limit/clients`
  - 健康检查：`GET /health`（白名单）

### 2) 查询服务（8090）

- `GET /health` 正常；
- `/metrics` 未在主入口启用（仅 `main.go.backup` 有 `promhttp` 示例）。前端 Vite 代理存在 `/api/metrics`→`/metrics` 映射（Dev），但后端未提供，对齐待定。

### 3) 通用健康模块（暂未接入服务）

- `pkg/health/*` 定义了 `HealthManager`、`StatusReporter`、多种 Checker（含 Redis/Neo4j 历史遗留），当前未接入运行路径，仅作潜在扩展能力。

### 4) 一致性与契约问题

- OpenAPI 缺失运维端点 `/api/v1/operational/*` 与权限 scopes（违反“API 优先授权管理”）；
- PBAC `RESTAPIPermissions` 未覆盖上述运维端点，存在“unknown endpoint” 风险；
- 查询服务 `/metrics` 与前端代理不一致（可选修复/移除代理）。
- 多租户隔离未明确：现有监控 SQL 为全库聚合，未按租户隔离，存在跨租户可见性风险。
- Debug 限流端点未受 PBAC 保护，存在对外暴露运行状态的风险。

## 方案设计（审批项）

> 新增页面属于“新增功能”，实施前需用户审批；同时先更新契约，后落地实现。

### A) 契约与权限（Architecture / Backend）

- 在 `docs/api/openapi.yaml` 新增 tag 与路径（系统管理范畴）：
  - `GET /api/v1/operational/health` — scope: `system:monitor:read`
  - `GET /api/v1/operational/metrics` — scope: `system:monitor:read`
  - `GET /api/v1/operational/alerts` — scope: `system:monitor:read`
  - `GET /api/v1/operational/tasks`、`GET /api/v1/operational/tasks/status` — scope: `system:ops:read`
  - `POST /api/v1/operational/tasks/{taskName}/trigger`、`POST /api/v1/operational/cutover`、`POST /api/v1/operational/consistency-check` — scope: `system:ops:write`
- PBAC 映射：在 `RESTAPIPermissions` 中补齐以上端点 → 权限映射，角色到权限策略对齐；
- 查询服务 `/metrics`：可选恢复 `promhttp` 暴露（仅 Dev），或移除前端代理映射，避免误导；
- 说明：这些端点归类“系统运维”，不纳入业务 CQRS 查询范畴，仍走 REST 管理路径（与 `/health` 一致）。

#### 必加细则（按评审结论）：

- 多租户隔离策略：
  - 默认按请求上下文的 `tenantId` 聚合与返回（从 JWT/X-Tenant-ID 获取），严禁跨租户数据泄漏；
  - 仅“平台管理员”具备全局视角（可通过显式查询参数 `scope=global` 或特定管理员角色/claim 控制），契约中需明确；
  - 后端在 `TemporalMonitor` 的 SQL 中注入租户过滤条件，或在视图/函数层封装；
- 限流统计端点规范化：
  - 新增 `GET /api/v1/operational/rate-limit/stats`（scope: `system:monitor:read`）；
  - `/debug/rate-limit/*` 仅 Dev 可用或移除（生产返回 404/403），避免泄露；
- 契约细化：为各运维端点补充 operationId 与响应 Schema（camelCase 字段约束），示例：
  - operationId: `getOperationalHealth`, 响应 data 对象包含 `status`, `healthScore`, `summary.{totalOrganizations,currentRecords,futureRecords,historicalRecords}`, `issues.{duplicateCurrentCount,...}`, `lastCheckTime`；
  - operationId: `getOperationalMetrics`, 响应 data 为 MonitoringMetrics 全字段；
  - operationId: `getOperationalAlerts`, 响应 data 含 `alertCount`, `alerts[]`；
  - operationId: `getRateLimitStats`, 响应 data 含 `totalRequests`, `blockedRequests`, `activeClients`, `lastReset`, `blockRate`（字符串百分比或数值，契约固定其类型）。

### B) 前端首页“系统监控总览”页面（Frontend）

- 路由：使用现有占位 `/dashboard`，实现“系统监控总览”；
- 数据源（通过 `unifiedRESTClient`，自动注入 JWT 与 `X-Tenant-ID`）：
  - `/api/v1/operational/health`（概览卡：健康分、状态、lastCheck）
  - `/api/v1/operational/metrics`（基础量纲与一致性问题卡）
  - `/api/v1/operational/alerts`（告警列表，分级着色）
  - `/api/v1/operational/rate-limit/stats`（限流统计卡：Total/Blocked/Active/BlockRate）
  - Dev-only：`/dev/performance-metrics`（运行时资源卡：内存/GC/Goroutines/DB Pool）
- 交互：React Query 轮询 30s + 手动“刷新”按钮；错误态/无权限态有清晰提示；
- 视觉：复用 `ContractTestingDashboard` 卡片风格，分为“健康概览、基础量纲、一致性问题、告警、限流、运行时性能(Dev)”六区块；
- 权限：无 `system:monitor:read` → 页面呈现权限提示，不暴露数据。

#### 前端一致性与容错

- 统一客户端路径：所有调用均走 `/api/v1/operational/*` 规范路径，避免 `/debug/*` 直连；
- 轮询策略：30s 基础轮询 + 错误指数退避（例如 30s→60s→120s 上限）；
- 缓存/条件请求：后端可选支持 `ETag/Last-Modified`，前端携带 `If-None-Match/If-Modified-Since`，减少负载（Phase 2）。

### C) Dev/Prod 差异与安全

- Dev：展示 `/dev/performance-metrics` 卡片；
- Prod：隐藏 Dev 卡片；仅显示运维端点数据；
- 跨域：沿用前端 Vite 代理；生产态走同域 BFF 模式（由统一客户端负责）。

补充：将 `DEV_MODE` 默认值设为 false，需通过环境变量显式开启；生产强制关闭 `/dev/*` 与 `/debug/*`。

## 实施计划与分工

1. Architecture（契约优先）
   - 在 `docs/api/openapi.yaml` 补齐运维端点 + scopes；
   - 评审并获批（PR 勾选“文档治理与目录边界”项）。

2. Backend（权限与实现对齐）
   - `RESTAPIPermissions` 增加运维端点映射；
   - 验证 `401/403/200` 分支与 PBAC 日志；
   - 可选：查询服务恢复 `/metrics`（Dev）或移除前端代理。

3. Frontend（监控总览页面）
   - 新增 `shared/api/monitoring.ts`：封装上述 REST 调用与类型；
   - 新增 `features/monitoring/MonitoringDashboard.tsx` 页面；
   - 在 `App.tsx` 将 `/dashboard` 路由指向新页面；
   - 轮询、错误态、权限态处理；Dev-only 卡片按环境隐藏。

4. 前端代理一致性（临时措施）
   - TODO-TEMPORARY: 处理 `/api/metrics` 代理不一致问题（二选一，截止下个迭代末）：
     - A) 移除 Vite 中 `/api/metrics` 代理映射；
     - B) 在查询服务 Dev 环境恢复 `/metrics`（`promhttp`），并在文档注明仅开发可用；
   - 最终收敛到一处单一事实来源，避免悬挂代理。

5. QA（验证与门禁）
   - E2E：登录后可见监控卡片；指标值字段齐全；无权限用户看到权限提示；
   - 性能：轮询对后端负载评估（P95 < 100ms 目标，示例）；
   - 契约测试：OpenAPI 与实现/前端类型一致性。

6. DevOps（可选）
   - 如启用 Prometheus/Grafana，本地脚本或 Compose 方案统一；
   - 产出最小使用说明与报警样例规则（后续阶段）。

## 里程碑与时间预估（工作日）

- D1：契约补齐 + 评审  
- D2：PBAC 映射与后端验证  
- D3：前端页面开发（基础卡片 + 轮询 + 错误态）  
- D4：告警列表与限流/Dev 性能卡片，联调与 E2E  
- D5：收尾与文档更新、归档评审

## 验收标准（Definition of Done）

- 契约：`openapi.yaml` 含 `/api/v1/operational/*` 端点与 scopes，已评审通过；
- 权限：Admin 角色可访问监控端点，普通用户默认无权（或按策略）；
- 页面：`/dashboard` 正常渲染各卡片；30s 轮询生效；错误/无权限态清晰；
- 一致性：前端 TS 字段与后端 camelCase 对齐，禁止 snake_case 泄露；
- 多租户隔离：普通租户仅看到本租户聚合；平台管理员可选择全局视角；
- 测试：
  - E2E：登录态加载页面成功；覆盖 401/403/200 分支；
  - 合同测试：无新增契约违规；字段命名契约测试覆盖运维端点；
  - 手册验证：模拟告警时页面呈现 CRITICAL/WARNING；
- 轮询负载评估：记录 10 分钟轮询对服务的 P95 和 CPU/连接池影响（目标 P95 < 100ms，或提供数据与整改计划）；
- Debug/Dev 端点关闭策略：生产环境访问 `/dev/*` `/debug/*` 返回 404/403；前端 Dev-only 卡片在生产隐藏。
- 文档：本计划文档更新；在 README/导航中补充入口说明；CI 文档治理通过。

## 风险与缓解

- 契约滞后：先契约、后实现，阻断代码变更直至评审通过；
- 指标开销：轮询频率默认 30s，提供手动刷新；若数据量大，考虑缓存/轻量化字段；
- 权限误配：PBAC 未覆盖将导致 403/未知端点，纳入集成测试；
- Prometheus 端点不一致：若不启用查询服务 `/metrics`，移除前端代理；如启用，限定 Dev。

补充：
- 多租户数据泄漏风险：严格租户过滤，平台管理员全局视角需具备专门 claim/role 且有审计；
- Dev 默认开关风险：`DEV_MODE` 默认关闭，部署文档强调生产禁用 Dev/Debug；
- SQL 性能风险：对一致性查询做慢查询采样与必要索引审计，必要时下推到物化视图/定时预聚合（后续阶段评估）。

## 开放问题（需审批/决定）

1. 是否同意将 `/dashboard` 实现为“系统监控总览”页面（新页面审批）？
2. 运维端点 scopes 命名与默认角色授权（提议：仅管理员 `system:monitor:read`）？
3. 查询服务是否恢复 `/metrics`（Dev）用于跨服务指标演示？若不恢复，是否立即移除前端 `/api/metrics` 代理？
4. 是否接受新增聚合端点 `GET /api/v1/operational/overview`（一次返回 health+metrics+alerts+rateLimit）以降低轮询开销（Phase 2）？

---

## 开发完成情况（状态追踪）

截至 2025-09-15 当前进度如下：

- ✅ 契约补齐（OpenAPI）
  - 已新增 `operational` 标签与以下端点（含 scopes）：
    - `GET /api/v1/operational/health`（`system:monitor:read`）
    - `GET /api/v1/operational/metrics`（`system:monitor:read`）
    - `GET /api/v1/operational/alerts`（`system:monitor:read`）
    - `GET /api/v1/operational/rate-limit/stats`（`system:monitor:read`）
    - `GET /api/v1/operational/tasks`、`/tasks/status`（`system:ops:read`）
    - `POST /api/v1/operational/tasks/{taskName}/trigger`、`/cutover`、`/consistency-check`（`system:ops:write`）
  - 新增 scopes：`system:monitor:read`、`system:ops:read`、`system:ops:write`

- ✅ PBAC 对齐
  - `RESTAPIPermissions` 已映射上述运维端点；
  - 角色 `ADMIN` 绑定 `SYSTEM_MONITOR_READ`、`SYSTEM_OPS_READ`、`SYSTEM_OPS_WRITE`。

- ✅ 多租户隔离
  - `TemporalMonitor` 所有 SQL 聚合/一致性检查已按 `tenant_id` 过滤；
  - 周期任务无租户时做全局汇总，仅用于内部日志。

- ✅ Debug 端点治理
  - `/debug/rate-limit/*` 仅在 `DEV_MODE=true` 时注册；
  - 生产默认关闭 Debug 端点。

- ✅ 受保护限流端点
  - 新增 `GET /api/v1/operational/rate-limit/stats`（统一信封、PBAC 保护）。

- ✅ 前端“系统监控总览”页面
  - 新增 API 封装 `frontend/src/shared/api/monitoring.ts`；
  - 新增页面 `frontend/src/features/monitoring/MonitoringDashboard.tsx` 并挂到 `/dashboard`；
  - 30s 自动轮询 + 手动刷新，错误提示与权限态提示。

- ✅ 开发代理收敛
  - 移除 Vite `/api/metrics` 代理映射，避免依赖查询服务未实现的 `/metrics`。

- ✅ 端到端测试（最小）
  - 新增 `frontend/tests/e2e/operational-monitoring-e2e.spec.ts`：
    - 验证 401/403、管理员 200；
    - 覆盖 health/metrics/alerts/rate-limit 四端点；
  - 本地执行通过（chromium/firefox）。

## 后续计划（Next Steps）

短期（下个迭代内）：
- [ ] TODO-TEMPORARY：`/api/metrics` 代理与查询服务 `/metrics` 收敛（择一）：
  - A) 维持移除代理（当前方案，已完成）；
  - B) 恢复查询服务 Dev-only `/metrics`（`promhttp`），并在文档说明仅开发可用；
- [ ] 页面级 E2E：补充 `/dashboard` 页面渲染与轮询用例（校验卡片渲染与数据随时间刷新）；
- [ ] 非管理员角色 E2E：校验 `system:monitor:read` 权限不足时的 403 与前端权限提示；
- [ ] 指标负载评估：记录 10 分钟轮询 P95/CPU/连接池影响，必要时调整轮询或加 ETag；

中期（Phase 2）：
- [ ] 新增聚合端点 `GET /api/v1/operational/overview`（一次返回 health+metrics+alerts+rateLimit），降低前端并发与轮询次数；
- [ ] 条件请求支持：`ETag/Last-Modified` + `If-None-Match/If-Modified-Since`，服务端返回 `304 Not Modified`；
- [ ] 索引审计：对一致性查询做慢查询采样与必要索引优化，评估是否引入定时预聚合/物化视图（仅在必要时）；

长期：
- [ ] Prometheus/Grafana 集成（按 DevOps 计划推进），统一观测与告警；
- [ ] 平台管理员全局视角机制完善（claim/role 设计与审计）与前端租户切换控件。


## 运行验证步骤（最小）

1) 契约检查：`npm run lint:api`（OpenAPI 校验无误）  
2) 后端本地启动后：
   - `curl -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: <uuid>" http://localhost:9090/api/v1/operational/health`
   - `curl -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: <uuid>" http://localhost:9090/api/v1/operational/metrics`
   - `curl -H "Authorization: Bearer $TOKEN" -H "X-Tenant-ID: <uuid>" http://localhost:9090/api/v1/operational/alerts`
   - `curl http://localhost:9090/debug/rate-limit/stats`
3) 前端 Dev：访问 `http://localhost:3000/dashboard`，观察卡片与轮询刷新；
4) QA：切换为非管理员账号，验证无权限提醒；模拟告警数据验证展示。

## 变更记录与对齐

- 不涉及数据库变更；
- OpenAPI 与 PBAC 映射为“契约与权限对齐”工作；
- 前端新增页面为新增功能，遵循“新功能审批”流程；
- 文档治理：本文件仅为计划，完成项迁移至 `docs/archive/development-plans/`，在 Reference 中补充稳定入口与使用说明。
