# 402C · 服务接入与迁移

**关联计划**：Plan 400、Plan 401、Plan 402、Plan 403、Plan 252/259、Plan 300  
**状态**：待启动（依赖 402A/402B 通过）  
**范围**：命令/查询服务接入 SOM、一次性迁移与校验、outbox 与快照刷新、前端 Manifest、能力契约日志  
**日志要求**：`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/eventbus/*.log`、`logs/plan402/ui/*.log`、`logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log`、`logs/plan402/capability/*.log`

> 目标：在不破坏现有服务的前提下，让组织/职位模块通过 Standard Object 端口运行，完成一次性迁移与自动校验，并完善前端/权限/事件链路。

---

## 1. 目标

1. 命令/查询服务仅通过 `internal/standardobject/api.go` 的 Port 操作 SOM，移除组织/职位旧仓储的写路径，彻底替换成新模型。
2. 以一次性迁移 + `standardobject-validator` 校验完成数据转移，确认旧表可转为只读或下线，不再依赖 Feature Flag/双写。
3. 接入 outbox 事件、快照刷新、GraphQL 缓存与前端 Manifest/Slot，使所有读写流程直接面向 SOM。
4. 将能力契约与视点矩阵证据纳入 `logs/plan402/capability/*.log`，保证与 Plan 400/403 一致。

---

## 2. 工作项

### C1 · Port 接入
- 命令服务：`cmd/hrms-server/command/internal/organization|position` 仅依赖 `standardobject.ObjectService`，删除旧仓储写路径、Feature Flag 分支与 `legacyRepo` 字段；所有写操作通过统一的 `TransactionClock`（新增 `pkg/temporal/clock`）生成 `transaction_range`。
- 查询服务：GraphQL Resolver 直接调用 `internal/standardobject/query`（或快照视图），不再根据 Flag 决定读旧表；接口继续暴露 `asOfValid`/`asOfTransaction`，由 SOM 查询实现。
- 迁移完成后在 `logs/plan402/capability/*.log` 记录 Federate、Port 版本与落实时间，佐证命令/查询完全切至 SOM。
- **实施清单**：
  | 步骤 | 目标模块/脚本 | 操作说明 |
  |------|---------------|----------|
  | 1 | `cmd/hrms-server/command/main.go`、`internal/organization/service`、`internal/position/service` | 注入 `standardobject.ObjectService` 并移除 `legacyRepo` 字段；Command handler 内只保留 SOM 写路径，旧仓储若需保留仅允许在只读诊断入口使用。 |
  | 2 | `cmd/hrms-server/query/internal/app/app.go`、`internal/organization/resolver`、`internal/position/resolver` | Resolver 直接依赖 `standardobject.QueryService`，删除“Flag 关闭走旧表”的逻辑；GraphQL DTO/Loader 以 SOM 为唯一事实来源。 |
  | 3 | `internal/standardobject/adapter/sqlc` | Adapter 维护事务注入、`transaction_range` 与 Schema 校验；`internal/standardobject/adapter/noop` 若无必要可迁移至测试包或删除。 |
  | 4 | 文档/速查 | 更新 `internal/standardobject/README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`，强调组织/职位写路径仅走 SOM，并在既有 `logs/plan402/capability/*.log` 中登记 Port 切换时间与责任人。 |

### C2 · 一次性迁移与验证
- 使用 402B 交付的 `cmd/tools/standardobject-migrator` 将组织/职位旧表数据批量导入 SOM，迁移完成后将旧表标记为只读（或拆除写入入口），不再保留双写逻辑。
- 迁移完成后立即运行 `cmd/tools/standardobject-validator`，对比计数、哈希、双时态区间、DEC/OCL 等指标，确认 `time-constraint-report.log` 与 `transaction-gap.log` 无异常；若发现差异，通过迁移修复脚本补齐，不再依赖运行时 Flag 切换。
- 所有结果写入 `logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`，并在 `logs/plan402/capability/*.log` 记录“迁移 + 校验”证据。
- **执行步骤**：
  1. `make db-migrate-all && go run ./cmd/tools/standardobject-migrator --dsn "$DATABASE_URL" --log-file logs/plan402/migration/migrator-$(date +%Y%m%d%H%M).log --limit <batch>`，可先以 `--dry-run` 演练，再移除该参数执行正式迁移；批次/重试策略沿用 `cmd/tools/standardobject-migrator` README。
  2. 将旧仓储设为只读：按 402B hazard list 中的 SQL（例如 `REVOKE INSERT,UPDATE,DELETE ON organization_units FROM app_user`）回收权限，并在命令服务中删除对应写分支；必要时在 `logs/plan402/migration/*.log` 记录执行摘要。
  3. 运行 `go run ./cmd/tools/standardobject-validator --dsn "$DATABASE_URL" --log-file logs/plan402/validator/validator-$(date +%Y%m%d%H%M).json --time-constraint-report logs/plan402/migration/time-constraint-report.log --transaction-gap logs/plan402/migration/transaction-gap.log`。
  4. 若有差异，使用 migrator 的 `--limit`/`--dry-run` 精准定位并修复，再次执行步骤 3 至 PASS；不得以“允许差异”“白名单”形式跳过。

### C3 · Outbox / 快照
- Outbox dispatcher 写入 `standard_object.*` 事件，刷新快照与下游模块；事件 payload 必须包含 `transactionTimestamp`；`logs/plan402/eventbus/*.log` 记录配置与指标。事件命名严格沿用主计划：`standard_object.created`、`standard_object.updated`、`standard_object.versioned`、`standard_object.status_changed`、`standard_object.retired`，对象类型通过 payload 的 `objectType` 区分，禁止衍生 `standard_object.organization.*` 之类二级命名。
- Outbox dispatcher 仍由 PostgreSQL 事务性发件箱驱动，消费者使用 Redis/Asynq（`pkg/eventbus/asynq`）落地；快照刷新任务只允许由该持久化队列触发（禁止在生产路径中调用内存队列/`async.Enqueue`）。部署与监控配置复用 402B/Plan 401 已交付的 `configs/eventbus.standard-object.yaml` 与运维手册，本阶段只需在 `logs/plan402/eventbus/*.log` 中记录启用时间与健康检查结果。
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

- ### C7 · 视点与能力验收
- 按 Plan 400 视点矩阵输出结构/运行/观察/协作证据，迁移巡检必须包含视点名称与能力契约 ID。
- `scripts/quality/architecture-validator.js` 在 capabilityContracts 规则中检查新增对象类型/Tab 是否已登记。

---

## 3. 交付物

- 命令/查询服务改造 diff、Port/Adapter 注入代码、`TransactionClock` 集成与旧仓储写路径移除记录。
- 迁移/校验运行日志：`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`time-constraint-report.log`、`transaction-gap.log`。
- Outbox / 快照 / EventBus 配置与 `logs/plan402/eventbus/*.log`、`logs/plan402/snapshots/*.log`。
- Manifest/Slot diff、`standardObjectAdapter` 更新、`logs/plan402/ui/*.log`、`logs/plan400/manifest/*.log`。
- Schema Registry / 翻译 / 附件 / 指标消费代码与 `logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log`。
- 能力契约日志 `logs/plan402/capability/*.log`（含视点矩阵校验）。

---

## 4. 验收标准

1. `make test`、`make test-db`、`npm run test` 在 SOM 路径下全部通过（仓储/Resolver 中不再引用旧写路径或 Feature Flag），并在报告中记录 `standardobject` Adapter 版本。
2. `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator` 在同一数据快照上跑通，`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`time-constraint-report.log`、`transaction-gap.log` 全部为 PASS，差异为 0；旧表权限已降至只读（或清理完毕），且在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 中注明。
3. Outbox/快照链路跑通，Plan 401/Plan 250 检查通过，`logs/plan402/eventbus/*.log` 无告警，`logs/plan402/snapshots/*.log` 标记 TC1/TC2 刷新策略及 `transaction_lag` 指标。
4. `npm run quality:preflight`、`npm run test:e2e`（含 Manifest/Slot/OBS 证据）通过，生成器输出与 `docs/api/*` 契约一致。
5. Plan 252/259 守卫与 capability 合规检查通过；`logs/plan402/capability/*.log`、`logs/plan402/ui/*.log`、`logs/plan402/capability/validator-*.log` 证据齐全且被 `scripts/quality/architecture-validator.js --rule capabilityContracts` 校验。
6. Schema Registry 校验、翻译/附件渲染、`data_classification`/指标透传在测试/CI 中有证据，并确认 0 漂移/0 丢失；`logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log` 需列出所有对象类型并与 `docs/reference/schema-registry.json` 哈希匹配。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 迁移遗漏 / 校验不充分 | 数据进入 SOM 后仍残留旧表差异 | 使用 migrator + validator 一次性完成导入与校验，`time-constraint-report.log`、`transaction-gap.log` 必须无异常；发现差异立即通过限定批次（`--limit`/`--dry-run`）定位并重新迁移，禁止带差上线 |
| 性能退化 | 快照刷新 / 查询延迟 | 监控 `standard_object_*` 事件与 `snapshot_refresh_duration`，必要时扩容 Asynq worker 或调整批次，保持指标在 Plan 401 阈值内 |
| 旧表未转为只读 | 新写入仍落旧表导致双事实 | 在迁移完成后执行权限变更脚本并由 `make test-db` 覆盖，禁止再引用旧仓储写接口 |
| Manifest/前端更新滞后 | UI 不一致或测试失败 | 强制通过生成器更新 forms/columns，并将 Playwright/OBS 作为验收前置 |

---

402C 通过后方可启动 402D；本阶段结束时旧表必须转为只读并停止全部遗留写路径。
