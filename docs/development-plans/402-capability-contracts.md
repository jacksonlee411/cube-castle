# 402 · Standard Object 能力契约目录

**状态**：基线草案（随 Plan 400/402 迭代更新）  
**唯一事实来源**：`docs/development-plans/400-standard-object-model-plan.md`、`docs/development-plans/402-standard-object-single-source-plan.md`、`docs/api/openapi.yaml`、`docs/api/schema.graphql`、`internal/standardobject/**`、`database/migrations/20251201090000_create_standard_objects.sql`  
**日志规范**：`logs/plan402/capability/*.log`（记录每次联邦/模块接入或矩阵更新的 CLI 输出/校验结果）

> 说明：本文件用于集中记录组织（organization）、职位（position）及未来 HR 模块在复用 Standard Object 模型时的“Federate 能力契约”。任何模块新增/调整 SOM 能力都必须同步此表，并在 CI/本地通过 `node scripts/quality/architecture-validator.js --rule capabilityContracts` 校验。  
> 若需扩展字段/能力，请优先修改 Plan 400/402 以及对应迁移/端口，再更新本文；禁止直接引用第二事实来源。

---

## 1. 术语与证据

- **能力（Capability）**：模块对外暴露或依赖的 Standard Object 能力，例如 `ORG_HIERARCHY` 链接写入、`standard_object.*` 事件、Schema Registry 读写、Temporal 页面渲染 slot。
- **Federate**：与 SOM 交互的逻辑模块，通常对应 `internal/<module>` 下的命令/查询服务、前端入口或工具脚本。
- **证据路径**：
  - `logs/plan402/capability/*.log`：记录 `standardobject-migrator`、`standardobject-validator`、Feature Flag 切换与能力巡检的输出。
  - `logs/plan400/schema/*.log`：Schema Registry DEC/OCL 缺失检查。
  - `logs/plan400/snapshots/*.log`：快照/闭包刷新任务。
  - `frontend/tests/e2e/standard-object-*.spec.ts`：前端视点证据。

---

## 2. 能力矩阵（必填项）

| 模块 (Federate) | 提供能力 | 依赖能力 | 视点覆盖 | 证据/日志 |
|-----------------|-----------|-----------|-----------|-----------|
| **Organization Federate**<br/>`internal/organization/**` + REST/GraphQL | - 写入 `standard_objects`/`standard_object_versions`（ObjectKernel + TemporalVersion）<br/>- 维护 `ORG_HIERARCHY` 链接 & 快照<br/>- 广播 `standard_object.*` outbox 事件（命令服务 9090）<br/>- 提供 Schema Registry 条目 `objectType=ORGANIZATION_UNIT`，绑定 DEC/OCL | - SOM Core API (`internal/standardobject/api.go`)<br/>- Schema Registry 哈希校验<br/>- PBAC scope (`standardObject.organization.*`)<br/>- `cmd/tools/standardobject-snapshot-refresh` | **结构视点**：组织层级 + DEC<br/>**运行视点**：状态机、outbox、快照<br/>**观察视点**：`logs/plan400/ui` Playwright 证据<br/>**协作视点**：Manifest/Slot、PBAC | `logs/plan402/capability/org-federate.log`（迁移+巡检）<br/>`logs/plan400/schema/*.log`（DEC/OCL）<br/>`logs/plan400/snapshots/*.log`（刷新） |
| **Position Federate**<br/>`internal/position/**` + 前端 Position 详情 | - 写入/读取 `POSITION_BELONGS_TO_ORG` 链接<br/>- 维护职位版本 payload (`objectType=POSITION_ROLE`)<br/>- 消费 `standard_object.*` 事件（组织变更触发校验）<br/>- 通过 `standardObjectAdapter` 渲染 Temporal 页面 | - Organization Federate 事件 / 链接<br/>- SOM Core API、Schema Registry<br/>- Manifest Slot（`temporal:position:*`）<br/>- PBAC scope (`standardObject.position.*`) | **结构视点**：职位->组织依赖<br/>**运行视点**：双写/校验、事件消费<br/>**观察视点**：Playwright `position-*.spec.ts`<br/>**协作视点**：Manifest、PBAC | `logs/plan402/capability/position-federate.log`（双写/校验）<br/>`frontend/tests/e2e/standard-object-lifecycle.spec.ts`<br/>`logs/plan400/schema/*.log` |
| **Shared Federate**（Temporal UI/工具链）<br/>`frontend/src/features/temporal/**`、`scripts/generate-*` | - 提供 `standardObjectAdapter`、字段/表单生成（OpenAPI/GraphQL）<br/>- 输出 Manifest/Slot 注册及 OBS 事件模板<br/>- 汇总 Schema Registry/DEC 清单 & `docs/reference/schema-registry.json` | - GraphQL 查询 (`cmd/hrms-server/query`) + 快照 MV<br/>- Schema Registry 哈希、`standard_object_translations`<br/>- Capability logs (`logs/plan400/ui/*.log`) | **结构视点**：表单字段与 DEC 映射<br/>**运行视点**：GraphQL 查询 + 快照来源<br/>**观察视点**：OBS `[OBS] standardObject.*`、Playwright<br/>**协作视点**：Manifest/Slot、PBAC 列表 | `logs/plan400/manifest/*.log`（生成器运行）<br/>`logs/plan402/capability/shared-federate.log`（CLI 输出）<br/>`logs/plan400/ui/*.log`（Playwright/OBS） |

> 维护规则：
> 1. `提供能力`/`依赖能力`/`视点覆盖`/`证据` 任一项变更都需在 PR 中同步更新此表，并附最新日志。
> 2. 任一 Federate 暂停/下线时，需在“备注”列说明替代机制与回收计划，然后在 `docs/archive/development-plans/` 归档旧版本。
> 3. 新增 Federate 需至少覆盖结构+运行两个视点，并提供对应日志路径；否则视为不合规。

---

## 3. 维护与巡检流程

1. **更新矩阵**：当组织/职位或新模块调整 SOM 功能时，先更新 Plan 400/402 -> 迁移/端口/前端 -> 最后回到本文件补表。
2. **运行巡检脚本**：执行 `node scripts/quality/architecture-validator.js --rule capabilityContracts`，在 `logs/plan402/capability/<date>-matrix-check.log` 写入结果。
3. **提交证据**：将 `logs/plan402/capability/*.log`、`logs/plan400/schema/*.log`、`logs/plan400/snapshots/*.log` 等链接写入 PR，并确保 `reports/architecture/architecture-validation.json` 通过守卫。

---

## 4. 未来迭代占位

- **Workforce/Contract Federate**：待 Plan 4xx 立项后补充能力条目（默认继承 Shared Federate 的 Schema/Manifest 约束）。  
- **数据仓库/审计输出**：若后续将 SOM 事件写入数据仓库或审计总线，需要新增 Federate，并在此表说明消费链路与证据。  
- **第三方/跨域模块**：引入外部系统前，必须先在本表登记能力及回滚策略，再进入集成阶段。

---

如对本目录有更新需求，请在 `docs/development-plans/00-README.md` 登记并引用 Plan 400/402/403 的最新版本，确保事实来源一致。
