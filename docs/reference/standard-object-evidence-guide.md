# Standard Object 证据与 Schema Registry 指南

**状态**：长期有效参考  
**唯一事实来源**：`docs/development-plans/400-standard-object-model-plan.md`、`docs/development-plans/402-standard-object-single-source-plan.md`、`docs/development-plans/403-SOM元模型组成优化建议.md`、`docs/api/openapi.yaml`、`docs/api/schema.graphql`、`database/migrations/20251201090000_create_standard_objects.sql`、`docs/reference/schema-registry.json`  
**相关日志**：`logs/plan400/schema/*`、`logs/plan400/snapshots/*`、`logs/plan400/ui/*`、`logs/plan400/manifest/*`、`logs/plan400/audit/*`、`logs/plan402/migration/time-constraint-report.log`、`logs/plan402/migration/transaction-gap.log`、`logs/plan402/capability/*`

> 本指南用于说明 Standard Object（SOM）在 Schema Registry、能力巡检与证据留存方面的约束与模板。所有条目均需引用上述计划/契约，不得引入第二事实来源。

---

## 1. Schema Registry 要求

| 字段 | 说明 | 事实来源 |
|------|------|----------|
| `object_type` | 对应 `standard_objects.object_type`，以 `ORGANIZATION_UNIT`/`POSITION_ROLE` 等常量表示 | Plan 400 §4.1 |
| `schema_version` / `schema_hash` | JSON Schema 版本与 SHA256 哈希，生成时写入 `docs/reference/schema-registry.json` | Plan 400 §4.1 / 4.3 |
| `definition` | JSON Schema 正文（仅使用 camelCase 字段） | OpenAPI / GraphQL 契约 |
| `dec_bindings` | 字段路径 → ISO 11179 DEC ID / 语义说明，缺失会触发 `logs/plan400/schema/*` 报警 | Plan 400 §4.1.1 |
| `ocl_guards` | 组合约束（前置/不变量/后置），供 `pkg/ocl` 与 migrator/validator 执行 | Plan 400 §4.1.1、Plan 402 |
| `time_constraint` / `transaction_policy` | 对应 SAP TC1/TC2/TC3 及事务策略（如 `APPEND_ONLY`、`CORRECTION_ALLOWED`），要求工具链与守卫执行差异校验 | Plan 400 §4.1.2 / §4.1.2（双时态） / Plan 402 |
| `glossary_url` | 指向内部术语或 `docs/reference` 说明的链接，严禁引用外部百科 | AGENTS.md |

### 生成/校验流程

1. 执行 `make sqlc-generate`、`atlas migrate diff`、`make db-migrate-all` 以确保 Schema 表结构与生成脚本同步。
2. 运行 `node scripts/generate-schema-registry.js`（命令名称以 Plan 400/300 实际脚本为准，可与 `scripts/generate-forms-from-openapi.ts`、`scripts/generate-columns-from-graphql.ts` 同步）生成 `docs/reference/schema-registry.json`。
3. 运行 `node scripts/quality/architecture-validator.js --rule capabilityContracts`（或默认规则）以触发 `logs/plan400/schema/<timestamp>-dec-ocl.log`：

```json
{
  "timestamp": "2025-11-26T13:45:02Z",
  "objectType": "ORGANIZATION_UNIT",
  "decBindingsMissing": [],
  "oclViolations": [],
  "schemaHash": "d4f84cdf8a51...",
  "source": "node scripts/generate-schema-registry.js"
}
```

若 `decBindingsMissing` 或 `oclViolations` 非空，视为违规，需先更新 Plan 400/402/403 再提交。

---

## 2. 日志模板（logs/plan40x）

| 目录 | 内容 | 生成方式 | 核对脚本 |
|------|------|----------|----------|
| `logs/plan400/schema/` | Schema Registry 巡检、DEC/OCL/TimeConstraint/TransactionPolicy 缺口、`docs/reference/schema-registry.json` 哈希 | `node scripts/quality/architecture-validator.js --rule capabilityContracts` 或 `npm run quality:preflight` | architecture-validator（capabilityContracts 规则） |
| `logs/plan400/snapshots/` | `cmd/tools/standardobject-snapshot-refresh` 运行结果、行数/耗时/快照版本 | Plan 400/401 交付的 `cmd/tools/standardobject-snapshot-refresh`（或等效 Job） | 手工审阅 + Plan 401 守卫 |
| `logs/plan400/ui/` | Playwright `standard-object-*` 规格日志、`[OBS] standardObject.*` 事件截图说明 | `npm run test:e2e -- --grep standard-object` | `frontend/tests/e2e` + OBS 回放 |
| `logs/plan400/manifest/` | OpenAPI/GraphQL 生成器输出（forms/columns manifest）、差异摘要 | `node scripts/generate-forms-from-openapi.ts`、`node scripts/generate-columns-from-graphql.ts` | Manifest diff 守卫 |
| `logs/plan400/audit/` | 双时态审计：`transaction-range-report.log`、撤销/补录事件、`transaction_lag` 指标 | `cmd/tools/standardobject-validator --report transaction`、回滚/补录脚本 | Plan 400 §4.1.2 守卫、审计脚本 |
| `logs/plan402/mapping/` | 402A 映射规格评审、DEC/OCL/Time Constraint 巡检、契约同步日志 | `node scripts/quality/contract-checker.js`、`node scripts/quality/architecture-validator.js`、`pkg/temporal/constraints` 检查脚本、手工评审记录 | 402A 守卫、`docs/reference/schema-registry.json` gap 校验 |
| `logs/plan402/migration/` | migrator/validator 运行日志、`time-constraint-report.log`、`transaction-gap.log` | `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator` | 402B/402C 守卫 |
| `logs/plan402/capability/` | Federate 能力巡检、`standardobject-migrator`/`validator`、Feature Flag 切换记录 | `node scripts/quality/architecture-validator.js --rule capabilityContracts`、`cmd/tools/standardobject-migrator` | architecture-validator（capabilityContracts）、Plan 402 |

### 日志格式建议

```text
2025-11-26T14:05:11Z capability-check start
Plan: 402-capability-contracts
File: docs/development-plans/402-capability-contracts.md
Checks:
  - columns: OK
  - entries: OK (Organization/Position/Shared)
  - evidence: logs/plan402/capability, logs/plan400/schema, logs/plan400/snapshots
Result: PASS
```

日志必须包含：
- 时间戳（UTC）  
- 所属计划/命令  
- 关键校验项（columns/entries/evidence 等）  
- PASS/FAIL 及回滚提示

---

## 3. 观测点与视点矩阵

Plan 400 §4.8 定义了结构、运行、观察、协作四个视点。生成证据时需确保：
- 每个视点至少对应一份日志或测试：
  - 结构：`logs/plan400/schema/*` + `docs/reference/schema-registry.json`
  - 运行：`logs/plan400/snapshots/*` + outbox 事件
  - 观察：`logs/plan400/ui/*` + OBS 记录
  - 协作：`logs/plan400/manifest/*` + PBAC/Manifest diff
- PR 描述中应附上述日志路径，避免产生第二事实来源。

---

## 4. 常用命令速查

```bash
# 运行能力契约与 Schema Registry 巡检
node scripts/quality/architecture-validator.js --rule capabilityContracts

# 生成 Schema Registry 与前端表单/列
node scripts/generate-schema-registry.js
node scripts/generate-forms-from-openapi.ts
node scripts/generate-columns-from-graphql.ts

# 刷新快照（Plan 400/401 交付后使用）
go run cmd/tools/standardobject-snapshot-refresh/main.go  # 或团队约定的等效命令
```

---

如需新增证据或日志目录，请同步更新本指南与 Plan 400/402，并在 PR 中附上最新样例。
