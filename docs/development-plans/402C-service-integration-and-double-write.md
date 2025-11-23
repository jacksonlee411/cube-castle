# 402C · 服务接入与双写

**关联计划**：Plan 400、Plan 401、Plan 402、Plan 403、Plan 252/259、Plan 300  
**状态**：待启动（依赖 402A/402B 通过）  
**范围**：命令/查询服务接入 SOM、Feature Flag、双写/校验、outbox 与快照刷新、前端 Manifest、能力契约日志  
**日志要求**：`logs/plan402/doublewrite/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/eventbus/*.log`、`logs/plan402/ui/*.log`、`logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log`、`logs/plan402/capability/*.log`

> 目标：在不破坏现有服务的前提下，让组织/职位模块通过 Standard Object 端口运行，建立双写与自动校验机制，并完善前端/权限/事件链路。

---

## 1. 目标

1. 命令/查询服务通过 `internal/standardobject/api.go` 的 Port 操作 SOM，并保持 `STANDARD_OBJECTS_ENABLED` Feature Flag 可控。
2. 构建双写与自动校验（含 DEC/OCL + Time Constraint 巡检），确保旧表与 SOM 数据一致且时间区间合法，可快速回滚。
3. 接入 outbox 事件、快照刷新、GraphQL 缓存与前端 Manifest/Slot，确保读取路径以 SOM 为主。
4. 将能力契约与视点矩阵证据纳入 `logs/plan402/capability/*.log`，保证与 Plan 400/403 一致。

---

## 2. 工作项

### C1 · Port 接入
- 在 `cmd/hrms-server/command/internal/organization|position` 注入 `ObjectCommandService`，并保留旧仓储用于回滚；命令服务需接入统一的 `TransactionClock`，所有写操作都使用 `TransactionClock.Now()` 作为 `transaction_range` 下界。
- 查询服务（GraphQL）通过 `internal/standardobject/query` 读取 SOM 或视图，完善 DTO/Resolver。
- `STANDARD_OBJECTS_ENABLED` Feature Flag 记录默认值与回滚步骤；每次切换在 `logs/plan402/flag-history.log` 留痕。
- 每次接入需在 `logs/plan402/capability/*.log` 记录 Federate、版本与视点覆盖。

### C2 · 双写与校验
- 写路径：Flag 开启时同时写 SOM 与旧表，并输出 `logs/plan402/doublewrite/*.log`，记录 `validity_range`、`transaction_range`、TC 类型、操作人等信息；撤销/更正需通过 append-only（更新上一行的 `transaction_range` 上界）实现。
- 读路径：Flag 开启→读 SOM；关闭→读旧表，保证回滚安全。
- 校验任务：使用 `standardobject-validator` 或自定义任务比对数据，并执行 DEC/OCL + Time Constraint + 双时态巡检；异常（含 TC1 空窗/重叠、事务区间断裂、事务时间倒退）时自动切回旧表并输出 `logs/plan402/validator/*.json`、`time-constraint-report.log`、`transaction-gap.log`。

### C3 · Outbox / 快照
- Outbox dispatcher 写入 `standard_object.*` 事件，刷新快照与下游模块；事件 payload 必须包含 `transactionTimestamp`；`logs/plan402/eventbus/*.log` 记录配置与指标。
- 快照/闭包在写操作后触发刷新，复用 402B 的 refresh job，并在 `logs/plan402/snapshots/*.log` 留痕；TC1 对象需即时刷新，TC2/TC3 可批量刷新但指标需记录延迟；工具需支持 `--as-of-valid` 与 `--as-of-transaction`。
- 更新缓存/GraphQL 数据加载器，确保 `standard_object_links` 驱动层级查询，并允许 API 通过 `asOfValid`/`asOfTransaction` 参数控制视图。

### C4 · 前端 / Manifest
- 组织/职位页面通过 Manifest/Slot (`temporal:organization:*`, `temporal:position:*`) 调用 `standardObjectAdapter`，禁止散落 GraphQL/REST 客户端。
- 运行 `scripts/generate-forms-from-openapi.ts`、`scripts/generate-columns-from-graphql.ts`（或等效脚本），保证 UI 字段与契约同步；输出 `logs/plan400/manifest/*.log` & `logs/plan402/ui/*.log`。
- PBAC scope 取自 OpenAPI（Plan 252/259），通过 Manifest `requiredScopes` 控制显隐。

### C5 · 权限/事件治理
- 在 Plan 252/259 数据源中新增 `standardObject.*` scope，更新 `scripts/quality/auth-permission-contract-validator.js`。
- EventBus Adapter 使用持久化传输（Redis/Asynq 等），记录部署/监控配置。

### C6 · 通用维度消费
- 写路径校验 Schema Registry（哈希 + JSON Schema + DEC/OCL），违规返回契约错误；记录 `logs/plan402/schema/*.log`。
- 前端读取 `standard_object_translations`、附件、metadata；指标写入 `logs/plan402/metrics/*.log`，供监控面板使用。

### C7 · 视点与能力验收
- 按 Plan 400 视点矩阵输出结构/运行/观察/协作证据，双写巡检必须包含视点名称与能力契约 ID。
- `scripts/quality/architecture-validator.js` 在 capabilityContracts 规则中检查新增对象类型/Tab 是否已登记。

---

## 3. 交付物

- 命令/查询服务改造 diff、Port/Adapter 注入代码、`TransactionClock` 集成、Feature Flag 配置与回滚脚本。
- 双写运行日志 `logs/plan402/doublewrite/*.log`、自动校验报告 `logs/plan402/validator/*.json`。
- Outbox / 快照 / EventBus 配置与 `logs/plan402/eventbus/*.log`、`logs/plan402/snapshots/*.log`。
- Manifest/Slot diff、`standardObjectAdapter` 更新、`logs/plan402/ui/*.log`、`logs/plan400/manifest/*.log`。
- Schema Registry / 翻译 / 附件 / 指标消费代码与 `logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log`。
- 能力契约日志 `logs/plan402/capability/*.log`（含视点矩阵校验）。

---

## 4. 验收标准

1. `STANDARD_OBJECTS_ENABLED=true/false` 两种模式下，`make test`、`make test-db`、`npm run test` 全量通过。
2. 双写期间 `standardobject-validator` 无差异（或差异在允许范围且具备回滚策略），Time Constraint 与事务时间报告无空窗/重叠/倒退，异常可在 5 分钟内切回旧表。
3. Outbox/快照链路跑通，Plan 401/Plan 250 检查通过，`logs/plan402/eventbus/*.log` 无告警，`logs/plan402/snapshots/*.log` 标记 TC1/TC2 刷新策略及 `transaction_lag` 指标。
4. `npm run quality:preflight`、`npm run test:e2e` 通过，生成物与 `docs/api/*` 契约一致。
5. Plan 252/259 守卫与 capability 合规检查通过；`logs/plan402/capability/*.log`、`logs/plan400/ui/*.log` 证据齐全。
6. Schema Registry 校验、翻译/附件渲染、`data_classification`/指标透传在测试/CI 中有证据，并确认 0 漂移/0 丢失。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 双写不一致 | 生产数据异常 | 事务内双写 + 自动校验，异常即切回旧表并告警 |
| 性能退化 | 写入/校验延迟 | 批量写、异步校验、监控 TPS / 延迟，必要时降级只读模式 |
| Feature Flag 管理混乱 | 不同环境行为不一致 | 记录 Flag 变更，集中在 `logs/plan402/flag-history.log` 并由负责人审批 |
| Manifest/前端更新滞后 | UI 不一致或测试失败 | 强制通过生成器更新 forms/columns，并将 Playwright/OBS 作为验收前置 |

---

402C 通过后方可启动 402D；在此之前不得关闭旧表写路径。
