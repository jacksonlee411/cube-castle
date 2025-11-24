# 402D · 切换与回收

**关联计划**：Plan 400、Plan 401、Plan 402、Plan 222/255、Plan 300  
**状态**：待启动（依赖 402A~402C 完成）  
**范围**：一次性迁移、Feature Flag 切换、前端/查询适配、门禁与回滚演练  
**日志要求**：`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/ui/*.log`、`logs/plan402/verification/*.log`、`logs/plan402/rollback/*.log`、`logs/plan402/capability/*.log`（阶段结束须执行 `make archive-run-artifacts` 将上述证据按 Plan 272 要求迁移到 `archive/runtime-artifacts/<yyyy-mm>/` 并产出 manifest）

> 目标：将所有读写切换至 Standard Object 三表，停用旧 `organization_units` 写入，完成前端/查询适配与门禁验收，并验证回滚路径。

---

## 1. 目标

1. 执行最终迁移与校验，关闭旧仓储写路径，确保只读或归档。
2. 更新前端与查询层，使其完全依赖 SOM，并完成 Playwright/OBS 证据。
3. 跑通所有门禁（Go/前端/质量脚本），并完成 Goose Down + 数据回滚演练。
4. 输出切换 Runbook、执行报告与能力契约核对结果。

---

## 2. 工作项

### D1 · 切换执行
- 运行 `cmd/tools/standardobject-migrator` 对生产级数据进行最终导入，日志写入 `logs/plan402/migration/*.log`；同时记录 `transaction_range` 推导方式。
- 推荐命令格式（沿用 402C 交付的 `--dsn/--dry-run/--limit/--log-file` 参数），首次执行必须以 dry-run 演练，记录 resume token、批次大小与 `transaction_range` 推导公式：

  ```bash
  DATABASE_URL="postgres://postgres:postgres@127.0.0.1:5432/hrms?sslmode=disable"
  ts=$(date +%Y%m%d-%H%M%S)
  go run ./cmd/tools/standardobject-migrator \
    --dsn "${DATABASE_URL}" \
    --dry-run \
    --limit 0 \
    --log-file "logs/plan402/migration/${ts}-migrator.log"
  # 去除 --dry-run 后执行正式迁移；日志须注明 transaction_range=clock(now)
  ```

  所有迁移操作均需引用 `pkg/temporal/clock.NewSystemClock()`（避免跨层混用手工 `NOW()`），并在日志中说明 `standardobject.MakeVersionCode` 的 `code/effectiveDate/updatedAt/recordId` 组成，确保与 402C/Plan 400 保持一致。
- 执行 `standardobject-validator` 确认 0 差异，并输出 `time-constraint-report.log`（TC1/TC2/TC3 覆盖率、空窗/重叠统计）及 `transaction-gap.log`（事务区间断裂/倒退统计）；如有残留需在切换前消除。
- `standardobject-validator` 建议命令：

  ```bash
  ts=$(date +%Y%m%d-%H%M%S)
  go run ./cmd/tools/standardobject-validator \
    --dsn "${DATABASE_URL}" \
    --out "logs/plan402/validator/${ts}-report.json" \
    --log-file "logs/plan402/validator/${ts}-run.log" \
    --time-constraint-log "logs/plan402/migration/time-constraint-report.log" \
    --transaction-gap-log "logs/plan402/migration/transaction-gap.log"
  ```

  报告需记录 `legacyCount`、`standardObjectVersions`、`validityOverlapCount`、`transactionOverlapCount`、`transactionGapCount`，并附最近 5 条差异样本，Plan 222/255 可直接复查。
- 关闭旧仓储写路径（Feature Flag 默认开启 SOM），旧表仅保留只读视图或直接冻结。
- 部署前统一在 `config/feature-flags.yaml`/`internal/standardobject/featureflag` 中将 `organization.useLegacyStore`、`position.useLegacyStore` 设置为 `false`，并在命令层移除 `standardobject.NewNoopService()` 注入；旧表权限降至 `SELECT`，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 更新真源说明。
- 输出切换 Runbook（步骤、负责人、回滚条件）与 DEC/OCL/Time Constraint + 双时态体检报告，确保 `docs/reference/schema-registry.json` 与能力契约无缺口。
- Runbook 必须覆盖：切换窗口、暂停写入、迁移命令、校验命令、监控指标、回滚条件、复盘模板；体检报告引用 `docs/reference/standard-object-evidence-guide.md` 的日志目录，并对 `schema-registry.json` 哈希与 DEC/OCL/binding 缺口逐项说明。

### D2 · 前端与查询适配
- 前端（组织/职位页面）完全接入 `standardObjectAdapter`；清理组织特有冗余字段。
- 聚焦以下入口：`frontend/src/features/organizations/OrganizationDashboard.tsx`、`frontend/src/features/temporal/pages/entityRoutes.tsx`、`frontend/src/features/positions/PositionDetailView.tsx`、`frontend/src/shared/api/facade/*.ts`，一律使用 `frontend/src/features/temporal/entity/standardObjectAdapter.ts` 暴露的转换函数，禁止散落解析 `standard_objects*` 字段。
- 重新运行 `npm run quality:preflight`、`npm run test`、`npm run test:e2e`，并把日志写入 `logs/plan402/ui/*.log`。
- 运行 `PW_OBS=1 VITE_OBS_ENABLED=true npm run test:e2e -- --grep standard-object`，并在 `logs/plan402/ui/obs/` 记录 `[OBS] standardObject.*` 事件截图/日志，作为观察视点证据。
- GraphQL/REST 查询切换至 SOM 数据源，更新缓存/selector/OBS 事件；Manifest/Slot 记录 DEC 列表与视点，并支持 `asOfValid`/`asOfTransaction` 参数。
- Frontend 适配矩阵（按 Slot/页面拆解）：

  | 页面/Slot | 主要文件 | GraphQL/REST 依赖 | 必做事项 | Owner |
  |-----------|----------|-------------------|----------|-------|
  | `/organizations` 列表 | `frontend/src/features/organizations/OrganizationDashboard.tsx` | `GetOrganizations` GraphQL、`standardObject.list` REST | 统一 query key `['standardObject','organization',filters]`，移除 legacy 字段映射，使用 adapter 输出表格元数据 | Frontend Org |
  | 组织详情（`temporal:organization:*`） | `frontend/src/features/temporal/pages/organizationRoute.tsx` + `TemporalEntityLayout` | `TemporalOrganizationDetail` GraphQL | 通过 adapter 注入 `asOfValid/asOfTransaction`，Manifest 中列出 DEC 列表并记录 `[OBS] standardObject.organization.detail` | Shared Federate |
  | 职位列表/详情（`temporal:position:*`） | `frontend/src/features/positions/PositionTemporalPage.tsx`、`PositionDetailView.tsx` | `PositionTemporalDetail` GraphQL、`standardObject.position.*` REST | 统一版本/链接解析，删除 `positionTimelineAdapter`，Slot 中只消费 adapter 输出，记录 Playwright 证据 | Frontend Position |
  | GraphQL/REST Facade | `frontend/src/shared/api/facade/{organization,position}.ts` | Apollo hooks、REST 客户端 | 暴露 `queryStandardObjectTimeline({ asOfValid, asOfTransaction })` API，禁止组件内自建查询字符串 | Shared Federate |
  | Playwright / OBS | `frontend/tests/e2e/standard-object-lifecycle.spec.ts` | `PW_TENANT_ID` + `PW_JWT` | 增加“禁写期间只读提示”“版本回滚”“asOf 参数”三类脚本，日志落盘 `logs/plan402/ui/playwright-<ts>.log` | QA |

- Manifest/Slot 更新完成后执行 `npm run manifest:forms`、`npm run manifest:columns`，并把 diff/日志同步到 `logs/plan400/manifest/*.log` 以供 `document-sync` 守卫复查。

### D3 · 门禁与回滚
- 跑通 `make test`、`make test-db`、`scripts/quality/*`（含 capabilityContracts 规则），输出 `logs/plan402/verification/*.log`。
- 固化命令：`node scripts/quality/architecture-validator.js --rule capabilityContracts --output logs/plan402/capability/$(date +%Y%m%d-%H%M%S)-capability.log`，需验证 Organization/Position/Shared Federate 三个条目均 PASS；若有缺口，阻断切换。
- 编写 Goose Down + 数据回滚脚本，并在 staging 演练；日志存入 `logs/plan402/rollback/*.log`。
- 回滚脚本包含：`make db-rollback-last`、`psql -f database/scripts/plan402/revert-to-legacy.sql`、`scripts/plan402/export-standardobject-snapshot.sh`，并在 staging 使用真实快照演练；输出 `logs/plan402/rollback/<ts>-staging.log`，记录耗时、恢复点与验证命令。
- 在 `scripts/quality/architecture-validator.js` 中运行 capability contract 完整性检查，确认 4.3 表格条目覆盖全部 Federate，并在演练中验证双时态回放（恢复某事务时间点的视图）可行。
- 双时态回放步骤：`psql -f database/scripts/plan402/checkpoints/replay_view.sql -v as_of_ts="'2025-11-30 10:00:00+00'"` → `curl http://localhost:9090/api/standard-objects/organization/:code?asOfTransaction=2025-11-30T10:00:00Z`，再与 `standardobject-validator` 结果对比并写入 `logs/plan402/verification/as-of-<ts>.log`，确保 30 分钟恢复 SLA 可实测。

### D4 · 遗留路径回收与日志归档
- 将 `organization_units`、`positions` 等旧表降级为只读视图或迁移至 `archive/plan402/legacy-*.sql`，并执行 `database/scripts/plan402/drop-legacy-indexes.sql` 删除旧索引/触发器；执行日志写入 `logs/plan402/cleanup/<ts>-legacy-drop.log`，确保与 402E 的全面清理范围衔接。
- 清理由 SOM 切换后不再使用的 Go/TS 代码：例如 `internal/organization/repository/legacy_*`、`frontend/src/features/organizations/legacy-*`，提交前由 `rg 'organization_units'` + allowlist 校验结果写入 `logs/plan402/cleanup/<ts>-code-audit.log`，并在 PR 描述中声明剩余白名单及回收计划。
- 运行 `make archive-run-artifacts`（Plan 272 W3 脚本）将 `logs/plan402/**`、`reports/plan402/**` 等运行产物打包至 `archive/runtime-artifacts/<yyyy-mm>/`，同时附上 manifest（含 `sha256`、Plan/Run ID）；`logs/plan272/archive/plan402-cutover-<ts>.log` 需记录本次归档。未压缩的 logs 目录保留不超过 5 份纯文本证据。
- 完成交付后执行 `npm run guard:plan272`，并在 `logs/plan272/guard/plan402-cutover-<ts>.log` 与 `reports/plan272/plan402-cutover-guard-<ts>.txt` 保存结果，确保 Plan 272 守卫覆盖本阶段新增产物。

---

## 3. 交付物

- 切换 Runbook 与执行报告、迁移/校验日志、DEC/OCL 体检报告。
- 前端/查询适配 diff、Manifest/Slot 更新、Playwright/OBS 证据 (`logs/plan402/ui/*.log`)。
- 门禁运行日志 (`logs/plan402/verification/*.log`)、回滚脚本与演练记录 (`logs/plan402/rollback/*.log`)。
- 能力契约核对结果 (`logs/plan402/capability/*.log`)。

---

## 4. 验收标准

1. 切换后所有写操作仅落在 `standard_objects*`，旧表只读或归档，监控显示 0 双写差异，`time-constraint-report.log` 与 `transaction-gap.log` 均为 PASS。
2. 前端 UI 与 GraphQL/REST 功能在 SOM 模式下通过全量测试（含 Playwright 核心场景），日志齐全。
3. DEC/OCL/Time Constraint/Transaction Policy 体检报告 0 漏项、0 违规，能力契约矩阵覆盖所有视点；相关证据已存档。
4. 回滚演练可在 30 分钟内恢复旧实现并保证数据一致，并能恢复到指定 `transaction_timestamp` 的系统视图。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 切换期间出现异常 | 数据不一致或服务中断 | 严格按 Runbook 执行，设置回滚点并实时监控 |
| 前端回归遗漏 | UI/交互异常 | 强制执行 Playwright/OBS 证据，Plan 222 负责人复核 |
| 监控缺失 | 无法及时发现问题 | 切换前配置 SOM 专用指标与告警，纳入 Plan 272 |

---

## 6. Runbook（执行顺序与责任人）

| 步骤 | 操作 | 负责人 | 预计耗时 | 回滚判定 |
|------|------|--------|----------|----------|
| R1 | 通知窗口 → 暂停组织/职位写入，切换 Feature Flag，导出 `organization_units` 快照 | Ops + Org SME | 15 分钟 | 无法冻结写入或出现新事务 |
| R2 | 执行 `standardobject-migrator` dry-run → 正式迁移，记录 `logs/plan402/migration/*.log`、`transaction_range` 推导方式 | Data Eng | 45 分钟 | migrator 报错或 validator 差异 >0 |
| R3 | 运行 `standardobject-validator`、`node scripts/quality/architecture-validator.js --rule capabilityContracts`，输出报告 | QA + Architecture | 30 分钟 | 任一校验 FAIL/有差异 |
| R4 | 部署前端/查询改动，刷新 Manifest/Slot，运行 `npm run quality:preflight`、`npm run test`、`npm run test:e2e`、Playwright OBS，日志入库 | Frontend + Query | 45 分钟 | 门禁失败/OBS 缺证 |
| R5 | 观察指标：`som_write_success_rate`、`standardobject_validator_diff`、`som_transaction_gap_seconds`，生成切换完成通知 | SRE | 30 分钟 | 指标异常或出现告警 |
| R6 | Staging 回滚演练：执行 `database/scripts/plan402/rollback/*.sql` + `scripts/plan402/export-standardobject-snapshot.sh`，验证 30 分钟恢复 SLA | Data Eng + Ops | 30 分钟 | 演练超过 30 分钟或验证失败 |
| R7 | 遗留清理与日志归档：执行 `database/scripts/plan402/drop-legacy-indexes.sql`、`make archive-run-artifacts`、`npm run guard:plan272`，并落盘 `logs/plan402/cleanup/*.log`、`logs/plan272/archive/*.log` | DBA + QA | 45 分钟 | Plan 272 守卫失败或归档 manifest 缺失 |

Runbook 签核后需存档到 `docs/archive/development-plans/`，并在执行当日同步更新日志索引。

## 7. 监控与告警

- **SOM 写入成功率**：`som_write_success_rate = 1 - failed_standardobject_upsert / total_standardobject_upsert`，阈值 99.9%，来源 `cmd/hrms-server/command` Prometheus；低于阈值触发 P1 告警并回滚至 R6。
- **迁移/校验差异**：`standardobject_validator_diff`（从 `logs/plan402/validator/*` 提取），>0 立即阻断上线；Grafana Dashboard `observability/dashboards/som-cutover.json` 维护。
- **事务滞后**：`som_transaction_gap_seconds`（`cmd/tools/standardobject-snapshot-refresh` 输出 `transaction_lag`），>300 秒警告；需要在 `logs/plan402/metrics/*.jsonl` 持续写入。
- **前端观测**：`[OBS] standardObject.*` 事件通过 Loki 聚合，设置 `temporal_entity_error_rate` 阈值 1%；超过即回滚 UI 部署并重新运行 Playwright。
- **Feature Flag 守卫**：合成监控调用 `/api/standard-objects/organization/:code`（含 `asOfTransaction`），若响应中仍含 `organization_units` 字段或命中旧写路径，立即执行回滚脚本并记录 `logs/plan402/rollback/<ts>-synth.log`。

---

402D 完成后方可启动 402E；若出现重大问题，必须通过回滚脚本恢复旧数据路径并记录 `logs/plan402/rollback/*.log`。
