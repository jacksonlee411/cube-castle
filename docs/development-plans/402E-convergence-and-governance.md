# 402E · 收敛与治理（可选）

**关联计划**：Plan 400、Plan 401、Plan 215、Plan 272、Plan 403  
**状态**：待启动（依赖 402D 完成并稳定运行至少一个迭代）  
**范围**：旧表/代码清理、文档与索引更新、监控/OBS 完善、未来模块接入指南  
**日志要求**：`logs/plan402/cleanup/*.log`、`logs/plan402/capability/*.log`（更新记录）、`logs/plan402/metrics/*.log`

> 目标：在 SOM 切换稳定后，清理历史实现、完善文档/监控/质量守卫，并沉淀未来模块接入指南。

---

## 1. 目标

1. 清除 `organization_units` 等旧表、触发器与遗留仓储，只保留受控视图或归档。
2. 更新参考文档（Plan 400/401、开发者速查、Plan 00-README 等），确保 SOM 成为唯一事实来源。
3. 完善监控/OBS/质量守卫，例如 capability-contract-check、schema hash 校验、Time Constraint（TC1/TC2/TC3）与事务时间（transaction_lag）漂移告警。
4. 产出《SOM 接入指南》，指导 payroll/workforce 等未来模块复用 Standard Object 并正确处理双时态。

---

## 2. 工作项

### E1 · 数据与代码清理
- 运行 `database/scripts/plan402/drop-legacy-write-paths.sql`（记录至 `logs/plan402/cleanup/<ts>-write-lock.log`）锁死旧写路径，再执行 Goose 迁移 `database/migrations/2025XXXXXX_plan402e_drop_legacy_tables.sql` 清理 `organization_units`、旧触发器与 residual `effective_from/effective_to` 列。402E 需新增 `database/scripts/plan402/restore-legacy-views.sql`（来源：`organization_units` 现有视图定义）用于快速恢复旧视图/索引，执行顺序写入 Runbook。若 Goose 迁移失败，立即 `goose -dir database/migrations postgres down 1` 并运行恢复脚本，回滚过程必须写入 `logs/plan402/rollback/<ts>-legacy-restore.log`。
- 移除 `internal/organization/**` 及 `pkg/organization/**` 中的仓储/DTO，统一改为依赖 `internal/standardobject` Port；同步在 `scripts/quality/architecture-validator.js` 中实现新的 `legacyOrgUnits` 规则（扫描 `cmd/`, `internal/`, `pkg/`, `frontend/`），禁止出现 `organization_units` 直接引用。规则需要输出 `logs/plan402/cleanup/<ts>-validator.log` 并在 `.github/workflows/agents-compliance.yml` 中注册 Required check。
- 更新所有相关脚本/迁移时，执行日志需落盘到 `logs/plan402/cleanup/*.log`，并在 README 中登记命令、责任人和回滚入口，保证与 AGENTS.md“高危操作必须声明命令+影响+回滚”一致。脚本列表（含 drop、restore、Goose up/down）应记录在 `docs/reference/standard-object-evidence-guide.md` 的附录中，确保唯一事实来源。

### E2 · 文档与索引
- 更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、Plan 400/401、Plan 403 等文档，明确“`docs/reference/standard-object-evidence-guide.md` + schema registry”是 SOM 唯一事实来源，新内容仅在这些文件内追加章节，不再创建平行手册。
- 在 `docs/development-plans/00-README.md`、`docs/reference/standard-object-evidence-guide.md` 中添加 402E 归档链接；Plan 402 主文档与子计划完成后整体迁移到 `docs/archive/development-plans/`，在索引中保留验收标准与日志引用。
- 《能力契约与视点维护指南》以附录形式并入 `docs/reference/standard-object-evidence-guide.md`，并在 `scripts/quality/architecture-validator.js` 中新增 `capabilityContracts` 规则校验（复用现有脚本，避免第二事实来源）。

### E3 · 监控 / OBS / 接入指南
- 将 SOM 相关指标（`transaction_lag`、双时态回放耗时、TC1/TC2 违规）纳入 Plan 272 的 `scripts/quality/plan272-artifact-guard.js` 产物清单，配置告警并在 `logs/plan402/metrics/<ts>-som-metrics.log` 留存采样。
- 《SOM 接入指南》在 `docs/reference/standard-object-evidence-guide.md` 中新增章节，覆盖 Schema Registry、Port、Manifest/Slot、能力契约与回滚/日志要求，并指向现有 `docs/api/openapi.yaml`/`docs/api/schema.graphql` 作为契约源。
- 在 `scripts/quality/` 目录新增 `schema-hash-guard.js` 与 `time-constraint-guard.js`（402E 交付项）：前者对 `docs/reference/schema-registry.json`、`docs/api/openapi.yaml`、`docs/api/schema.graphql` 计算 canonical hash，与 `logs/plan400/schema/hash-baseline.log` 比较，确保未登记的 Schema 偏差不会进入主干；后者读取 `database/migrations/*standard_object*.sql` 与 `logs/plan400/migration/time-constraint-report.log`，校验 `standard_object_schemas.time_constraint/transaction_policy` 与 `pkg/temporal/constraints` 声明一致，并在发现 `transaction_range` 倒退时退出非零。两个脚本都需输出各自日志（`logs/plan402/metrics/<ts>-schema-hash.log`、`logs/plan402/metrics/<ts>-time-constraint.log`）并加入 `.github/workflows/agents-compliance.yml` Required checks。

---

## 3. 交付物

- 数据/代码清理脚本、执行日志 (`logs/plan402/cleanup/*.log`)。
- 更新后的 Plan 400/401/403、开发者速查、参考指南，以及归档版本。
- 监控/OBS 配置（含 Time Constraint 违规告警）、SOM 接入指南、质量守卫脚本。
- `logs/plan402/capability/*.log` 中的维护记录。

---

## 4. 验收标准

1. `node scripts/quality/architecture-validator.js --rule legacyOrgUnits` 与 `node scripts/quality/schema-hash-guard.js` 在 CI Required checks 中必须通过；若存在 `organization_units` 直接引用，脚本会输出 `logs/plan402/cleanup/<ts>-validator.log` 并阻断合并；若 Schema hash 与 `logs/plan400/schema/hash-baseline.log` 不一致，则 `logs/plan402/metrics/<ts>-schema-hash.log` 给出差异详情。
2. 文档索引（`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/development-plans/00-README.md`、`docs/reference/standard-object-evidence-guide.md`）同步更新并链接至 SOM SSoT；归档后的 402 文档位于 `docs/archive/development-plans/`，PR 需附 `git show docs/archive/development-plans/402E-convergence-and-governance.md` 证据。
3. Plan 272 守卫通过 `node scripts/quality/plan272-artifact-guard.js` 产出 `logs/plan402/metrics/<ts>-som-metrics.log`，内含 `transaction_lag`、TC 漂移、双时态回放指标；告警配置截图或链接需附在 PR 验证部分。
4. `make security && node scripts/quality/time-constraint-guard.js` 与 `node scripts/quality/architecture-validator.js --rule capabilityContracts` 均无告警，且 `logs/plan402/capability/<ts>.log` 记录最新维护结果（含守卫命令/Responsible/回收计划），证明回退到旧实现会被 CI 阻止。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 遗留引用遗漏 | 旧逻辑悄然复活 | 在 `scripts/quality/architecture-validator.js` 增加检查；运行 `rg organization_units` 并制定白名单 |
| 文档不同步 | 团队引用旧资料 | 建立文档清单与审核流程，更新 `docs/reference/standard-object-evidence-guide.md` |
| 监控/守卫缺失 | 回归无法及时发现 | 在 Plan 272/255 守卫中加入 SOM 专项脚本，强制在 PR/CI 运行 |

---

402E 完成后，Plan 402 可归档；后续模块接入需直接引用 Plan 400/401/403、能力契约与证据指南。
