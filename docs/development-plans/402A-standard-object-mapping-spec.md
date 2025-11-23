# 402A · Standard Object 映射规格（v0.1）

**状态**：已验收（2025-11-27 评审完成）  
**责任人**：Plan 402 Owner / 架构组  
**唯一事实来源**：`docs/reference/schema-registry.json`、`docs/api/openapi.yaml`、`docs/api/schema.graphql`、`docs/development-plans/400-standard-object-model-plan.md`、`docs/development-plans/402-standard-object-single-source-plan.md`

> 本规格把 `organization_units` 单表字段映射为 Standard Object 三表（`standard_objects`、`standard_object_versions`、`standard_object_links`）的执行矩阵。所有 DEC/OCL 信息同步登记在 `docs/reference/schema-registry.json`，日志与评审证据存放 `logs/plan402/mapping/`。

---

## 1. 范围与依赖

- **范围**：组织模块（命令/查询服务、GraphQL/REST 契约、`organization_units` 表），覆盖对象层、版本层与层级 Link。
- **依赖**：
  - `docs/reference/schema-registry.json`：记录 DEC/OCL 与 schema hash，供 402B/402C 使用。
  - `internal/standardobject/**`：402A 输出的 Port/Feature Flag 骨架（见 `internal/standardobject/README.md`）。
  - `docs/api/openapi.yaml`、`docs/api/schema.graphql`：新增 `StandardObject` 实体、scope 与字段。
  - `logs/plan402/mapping/*.log`：记录 spec 评审、契约守卫、DEC gap 与命名检查。
- **不在范围**：实际 SQL/Go 实现、迁移脚本、双写逻辑，将在 402B/402C 中完成。

---

## 2. 字段映射矩阵

| Source (`organization_units`) | DEC / 语义 | Target | 转换/校验 | 负责人 |
|------------------------------|------------|--------|-----------|--------|
| `tenant_id` | DEC_TENANT_CODE | `standard_objects.tenant_code` | 保留为 text；迁移前校验非空 | DBA |
| `code` | DEC_ORG_UNIT_CODE | `standard_objects.code` | 保持唯一索引；写入 `ObjectKernel.Code` | Backend |
| `name` | DEC_ORG_UNIT_NAME | `standard_object_versions.payload.name` | 迁移时写入 JSONB；GraphQL camelCase | Backend |
| `description` | DEC_ORG_UNIT_DESCRIPTION | `standard_object_versions.payload.description` | 允许 NULL，保留历史版本 | Backend |
| `unit_type` | DEC_ORG_UNIT_TYPE | `standard_objects.labels.unitType` | 按照标签规范写入 JSONB | Backend |
| `status` | DEC_ORG_UNIT_STATUS | `standard_objects.status` | 映射到 `LifecycleStatus`（DRAFT/READY/ACTIVE/...） | Domain |
| `effective_date` | DEC_TEMPORAL_EFFECTIVE_DATE | `standard_object_versions.effective_date` | DATE 保持；禁止 `_from/_to` 命名 | DBA |
| `end_date` | DEC_TEMPORAL_END_DATE | `standard_object_versions.end_date` | 允许 NULL；OCL 校验 `effective_date ≤ end_date` | Domain |
| `is_current` | DEC_VERSION_STATE | `standard_object_versions.is_current` | 与 `status`/`LifecyclePolicy` 双重校验 | Domain |
| `parent_code` | DEC_ORG_PARENT_CODE | `standard_object_links.source/target` | 映射为 `ORG_HIERARCHY` link；配合 `standard_object_hierarchy_snapshots` | Backend |
| `level` / `hierarchy_depth` | DEC_ORG_LEVEL | `standard_object_links.attributes.{level,hierarchyDepth}` | JSONB 属性；快照刷新时引用 | Backend |
| `sort_order` | DEC_SORT_ORDER | `standard_object_links.attributes.sortOrder` | 数值保持，为父子顺序提供稳定排序 | Backend |
| `code_path` / `name_path` | DEC_ORG_PATH | `standard_object_links` 衍生视图 | 迁移后不再存储；改为快照/闭包计算 | DBA |
| `profiles` | DEC 待定（Plan 403） | `standard_object_versions.payload.profiles` | JSONB 直接复制；缺少 DEC，见 §3 Hazard | Domain |
| `metadata` | DEC_ORG_METADATA | `standard_object_versions.payload.metadata` | JSONB；需在 402B 前完成字段拆解 | Domain |
| `created_at` / `created_by` | DEC_AUDIT_CREATED | `standard_objects.created_at/created_by` | 对象层记录；版本层 `auditTrail.createdAt` | Backend |
| `updated_at` | DEC_AUDIT_UPDATED | `standard_objects.updated_at` | 与版本 `updatedAt` 同步 | Backend |
| `deleted_at` / `deleted_by` / `deletion_reason` | DEC_AUDIT_DELETION | `standard_object_versions.auditTrail.deleted*` | JSONB 审计；触发器替换为 OCL 守卫 | Domain |
| `suspended_at` / `suspended_by` / `suspension_reason` | DEC_AUDIT_SUSPENSION | `standard_object_versions.auditTrail.suspended*` | JSONB 审计 | Domain |
| `operated_by_*`, `changed_by`, `approved_by` | DEC_AUDIT_ACTOR | `standard_object_versions.auditTrail.*` | 统一 actor schema（`{id,name,role}`） | Domain |

> 迁移脚本需尊重 camelCase 输出：GraphQL/REST/前端仅接受 `effectiveDate/endDate` 命名，不允许 `_from/_to`。如需更细粒度的时间精度，须在 402B 创建扩展列并更新 schema registry。

---

## 2.1 时间/事务约束声明

| 对象/字段 | Time Constraint | Transaction Policy | 说明 | 计划 |
|-----------|-----------------|--------------------|------|------|
| `standard_objects` (组织对象 kernel) | TC1 | APPEND_ONLY | 任意时刻必须存在且唯一，禁止空窗；迁移时依赖 time slicing。写路径只能追加新 kernel，纠偏需通过 Goose Down。 | 402B 在 `standard_object_schemas.time_constraint/transaction_policy` 列中登记，并在 migrator 中实现裁剪 |
| `standard_object_versions` (组织版本) | TC1 | APPEND_ONLY | 版本区间需连续覆盖（无重叠/空窗），所有变更通过追加新版本实现；合并/回滚依赖事务日志而非覆盖写。 | 402B 交付触发器/validator；402C 命令侧启用 `pkg/temporal/constraints` |
| `standard_object_links` `ORG_HIERARCHY` 关系 | TC2 | APPEND_ONLY | 最多一条 link，可存在空窗，用于临时解绑；事务层禁止覆盖历史记录。 | 402B 在 schema registry 中声明；validator 仅检查重叠 |
| `payload.profiles` 等扩展 JSON 字段 | TC3 | CORRECTION_ALLOWED | 允许同一时间多条记录（例如多标签/备注），并允许针对 JSON 键值纠偏；由版本快照记录差异。 | 402B 在 schema 生成时标记；查询层通过排序处理 |

所有 `timeConstraint` 与 `transactionPolicy` 值需与 `docs/reference/schema-registry.json.schemas[]` 顶层字段保持一致；生成流程沿用 Plan 400/`docs/reference/standard-object-evidence-guide.md` 中的 Schema Registry 生成器（同一套脚本负责写入 JSON），禁止在其他文档重复记录。若对象未来扩展（如 person / workforce），需在新条目中声明默认值并附 OCL。巡检结果沿用 Plan 400 的 `logs/plan400/migration/time-constraint-report.log`，禁止额外创建平行日志，分析结果在本节与 `hazard-list` 中登记。

## 2.2 有效期 / 事务期来源

- **validity_range**：由 `organization_units.effective_date`（起）与 `end_date`（止）推导；缺失 `end_date` 视为开放区间（`∞`），迁移脚本需补写为 `NULL`，并依赖 `standard_object_versions` 开启的 `TC1` 检查防止倒挂。
- **transaction_range**：以 `created_at`/`updated_at` 作为上下界；若历史记录在迁移窗口内被补录，则为该批次写入统一的 `migrated_at` 时间戳，并在 `logs/plan400/audit/transaction-range-report.log` 记录批次 ID、来源脚本与责任人，供 402B/402C 的裁剪工具识别“回溯写”。
- **异常/纠偏**：当检测到 `updated_at < created_at` 或人工回溯写入时，需在 §3 Hazard 中登记并列出 402B 的回收任务（脚本、触发器或额外 schema 字段），确保 transaction range 成为可验证事实。

该信息在 402B 的 Goose 迁移与 `pkg/temporal/constraints` 校验中作为输入；402A 阶段只维护说明与日志，确保“先契约后实现”。

## 2.3 迁移批次标记（`migrated_at` 等）

- **存储位置（沿用 Plan 400 双时态）**：直接复用 `standard_object_versions.transaction_range` 与 Plan 400 定义的审计视图。补录批次信息写入 `transaction_range` 上界与 `standard_object_audit` 视图，并在 `logs/plan400/audit/transaction-range-report.log` 记录批次 ID/责任人；禁止新增 JSON 副本，避免第二事实来源。
- **写入策略**：迁移脚本在追加版本时保持 append-only（`transaction_range = tstzrange(now(), 'infinity')`），批次纠偏通过闭区间更新上一行上界；若需额外批次 metadata，以 `correctionReason` 形式记录在 hazard list，并由 Plan 400 的审计脚本输出。
- **最佳实践**：当批次含人工纠偏数据时，在 `logs/plan400/audit/transaction-range-report.log` 中附带 `correctionReason`，并同步 `hazard-list`（无需在 schema 中增列），确保 402B/402C 的回滚脚本可按批次删除/重放。

该批次标记完全依赖 Plan 400 的 transaction range/Audit 机制，避免重复实现，同时保证 402D/402E 能快速执行审计或回滚。

---

## 3. DEC / OCL Hazard List {#hazard-list}

| 项目 | 描述 | 影响 | 回收计划 |
|------|------|------|----------|
| `payload.profiles` | 缺少 ISO 11179 DEC ID（Plan 403 未发布） | Schema Registry 不完整 | 402B 在 `standard_object_schemas` 中补齐，参照 `docs/reference/schema-registry.json` → `knownGaps[0]` |
| `payload.metadata` | 元数据结构因租户自定义而多态 | 无法生成 JSON Schema | 402B 需要抽象公共字段 + `metadata.*` 通配符 DEC，最迟在 402C 双写前完成 |
| Link attributes `hierarchyDepth`, `codePath` | 当前为派生列 | 缺少 DEC/OCL 绑定 | 402B 在 Link schema 中登记 DEC，新增快照校验 |
| `auditTrail` 结构 | 多字段复用 TEXT | 无法映射 `DEC_AUDIT_*` | 402B 设计 `auditTrail` JSON schema，并更新 `docs/reference/schema-registry.json` |
| `timeConstraint` 声明 | 旧表缺少 TC 字段，需在 Schema Registry 中新增 | 时间裁剪无法执行 | 402B 创建 `standard_object_schemas.time_constraint` 列并实现 migrator/validator 裁剪逻辑 |

Hazard 的唯一事实来源：本节 + `docs/reference/schema-registry.json.schemas[].knownGaps`。任何新增缺口必须同时修改两处并在 `logs/plan402/mapping/dec-gap.log` 记录。

---

## 4. 兼容策略与 Feature Flag

1. **视图/Port**：命令服务通过 `internal/standardobject.ObjectService` 注入 `adapter/noop`；查询服务保留 `organization_units` 读路径，并新增兼容视图 `standard_object_org_units_v`（由 402B 创建）。402A 仅记录结构。
2. **Feature Flag**：环境变量 `STANDARD_OBJECTS_ENABLED` 控制 `NoopService` 行为，Toggle 由 `internal/standardobject/featureflag.EnvToggle` 读取。Flag 默认关闭，打开后若仍未配置仓储将返回 `ErrAdapterNotConfigured`，确保可观测。
3. **日志**：Flag 调整、视图 explain、命名/契约守卫结果需落盘至 `logs/plan402/mapping`. README 中列出日志类型，CI 需上传 `api-contract.log` 与 `dec-gap.log`。
4. **回滚原则**：Feature Flag 关闭即回到旧仓储；402A 不创建新迁移，因此无额外 SQL 回滚操作，但要求维护 hazard list 以支持 402B 的 Goose Down。

---

## 5. 验证与守卫

| 守卫 | 命令 | 输出 |
|------|------|------|
| 契约 diff | `node scripts/quality/contract-checker.js && npm --prefix frontend run contract:generate` | `logs/plan402/mapping/api-contract.log` |
| DEC/OCL 校验 | `node scripts/quality/architecture-validator.js --rule capabilityContracts` | `logs/plan402/mapping/dec-gap.log`（JSON） |
| 命名守卫 | `node scripts/quality/architecture-validator.js --rule naming` | `logs/plan402/mapping/naming-check.log` |
| Go skeleton | `go test ./internal/standardobject/...` | 控制台输出 + `logs/plan402/mapping/spec-review.log` 附引用 |

验收前需确认上述守卫均附带日志，并在 PR 模板中引用具体文件路径。

---

## 6. 兼容视图 / 触发器替代

| 需求 | 说明 | 交付阶段 |
|------|------|----------|
| `standard_object_org_units_v` | 提供旧字段 → SOM 三表拼接，供查询服务/自检脚本 | 402B | 
| `ORG_HIERARCHY` Link 快照 | 402A 仅描述 schema；实际 `standard_object_hierarchy_snapshots` 在 402B 创建 | 402B |
| 触发器迁移 | `organization_units` 的 `set_updated_at`、`denormalize_parent` 将在 402C 替换为 OCL + outbox | 402C |

402A 文档只描述需求，避免重复实现。所有 SQL 细节由后续阶段的迁移脚本承载。

---

## 7. 证据要求

- `logs/plan402/mapping/spec-review.log`：记录评审会议时间、参会人、是否准许启动 402B（2025-11-27T09:00Z 条目已写明“通过，准许执行 402B”与责任人/回收任务）。
- `docs/reference/schema-registry.json`：本文件的 DEC/OCL 绑定需与 registry 中 `objectType=ORGANIZATION_UNIT` 条目一致。
- PR/Issue：402A 相关 PR 必须附上本规格、registry diff 以及日志路径，禁止在其他文档重复这些事实以维护唯一性。

> **回顾**：本规格建立 402A 的唯一事实来源，使得后续阶段可以直接引用字段映射、Feature Flag 行为和日志格式，在满足 AGENTS.md“资源唯一性”原则的前提下推进 402 系列计划。
