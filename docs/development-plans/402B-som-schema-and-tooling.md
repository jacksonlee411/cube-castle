# 402B · SOM Schema 与工具链

**关联计划**：Plan 400、Plan 402、Plan 403  
**状态**：待启动（依赖 402A 输出）  
**范围**：数据库迁移、sqlc 生成、迁移/校验工具、快照/Schema Registry、日志规范  
**日志要求**：`logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/snapshots/*.log`、`logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log`

> 目标：在数据库与工具链层面构建 Standard Object 三表、Schema Registry、迁移/校验工具与快照骨架，为服务接入与双写奠定基础。

---

## 1. 目标

1. 使用 Atlas/Goose 创建 `standard_objects`、`standard_object_versions`、`standard_object_links` 等表，并提供 Up/Down 脚本与触发器。
2. 更新 `sqlc.yaml` 和 `internal/standardobject/repository/sqlc`，提供标准化仓储接口。
3. 交付 `cmd/tools/standardobject-migrator`、`cmd/tools/standardobject-validator`、`cmd/tools/standardobject-snapshot-refresh` 等工具，形成迁移与快照刷新流水线。
4. 完成 Schema Registry、翻译、附件、元数据、指标等扩展表以及日志样例，确保 DEC/OCL 校验可执行。

---

## 2. 工作项

### B1 · 迁移脚本
- 创建 `20251201090000_create_standard_objects.sql`（Up/Down），包含三表、索引、约束、触发器（`is_current`、链接级联等）。
- 在同批脚本中加入 `standard_object_hierarchy_snapshots` 等 Plan 401 依赖的骨架，并在注释中标明用途。
- 更新 `database/migrations/README.md`，记录执行顺序与回滚说明；`logs/plan402/migration/*.log` 保存 `atlas diff` / `goose up`。

### B2 · sqlc & 包结构
- 调整 `sqlc.yaml`，生成 `internal/standardobject/repository/sqlc` 所需的 CRUD、列表、链接维护查询。
- 在 `internal/standardobject/domain` 创建实体、DTO 与接口，提供 `ObjectRepository`、`LinkRepository` 等 Port。
- 更新 `make sqlc-generate` pipeline 并记录日志，确保 CI/本地生成一致。

### B3 · 迁移/校验工具
- `cmd/tools/standardobject-migrator`：读取旧表写入 SOM，支持 dry-run、批次、失败回滚；输出 `logs/plan402/migration/migrator-*.log`。
- `cmd/tools/standardobject-validator`：对比计数、哈希、层级一致性，并执行 Schema Registry/DEC/OCL 校验；输出 `logs/plan402/validator/*.json`。
- 提供工具使用手册及异常回滚指引。

### B4 · 快照刷新骨架
- 交付 `cmd/tools/standardobject-snapshot-refresh`（或同等 Job），实现快照/物化视图刷新、dry-run、限速、指标输出。
- 在 `logs/plan402/snapshots/*.log` 保存运行记录（含快照版本、耗时）；失败场景需附回滚/重试说明。
- 在《Standard Object 映射规格》中补充快照/闭包校验项，并更新开发者速查命令。

### B5 · 通用维度扩展
- 建立 `standard_object_schemas` 表，包含 `dec_bindings`、`ocl_guards`、`glossary_url` 等字段，生成 `schema-registry.json`。
- 创建 `standard_object_translations`、`standard_object_attachments`、`standard_object_metadata`、`standard_object_metrics` 等扩展表与 sqlc 代码。
- 在 `logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log` 输出首批样例。

---

## 3. 交付物

- `database/migrations/20251201090000_create_standard_objects.sql`（Up/Down）与 `atlas diff` 日志。
- 更新后的 `sqlc.yaml`、`internal/standardobject/repository/sqlc` 代码与 `make sqlc-generate` 日志。
- `cmd/tools/standardobject-migrator`、`standardobject-validator`、`standardobject-snapshot-refresh` 源码/说明及运行日志。
- `logs/plan402/migration/*.log`、`logs/plan402/validator/*.json`、`logs/plan402/snapshots/*.log`、`logs/plan402/schema/*.log`、`logs/plan402/metrics/*.log` 样例。
- Schema Registry（含 DEC/OCL）、翻译、附件、元数据、指标表结构与 `schema-registry.json`。

---

## 4. 验收标准

1. `make db-migrate-all` / `make db-rollback-last` 在 Docker Compose 环境中通过，并将输出写入 `logs/plan402/migration/*.log`.
2. `make sqlc-generate` 在本地与 CI 均成功，`git status` 清洁；Plan 201 差异项脚本通过。
3. migrator/validator 在沙盒数据集上跑通，产出差异报告；出现差异时提供可执行对账结论。
4. Schema Registry 生成物通过 DEC/OCL 缺失检查，`logs/plan402/schema/*.log` 无告警。
5. 快照刷新工具完整运行一次，`logs/plan402/snapshots/*.log` 记录耗时/数据量；失败案例具备回滚记录。
6. `scripts/quality/architecture-validator.js` 或辅助脚本验证仓库内不存在 `effective_from/effective_to` 字段引用，结果写入 `logs/plan402/mapping/naming-check.log`。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Schema 约束与旧数据冲突 | 迁移失败或数据不一致 | migrator 中提供修复策略、hazard list，必要时阻断上线 |
| sqlc 生成不一致 | CI/本地差异导致编译失败 | 锁定 sqlc 版本/配置，纳入 `make sqlc-generate` 守卫 |
| 工具运行性能不足 | 迁移/校验耗时过长或锁表 | 支持批次、限速、dry-run；记录性能指标 |
| DEC/OCL 数据缺失 | Schema Registry 不可用 | 在 Schema 生成脚本中强制校验，缺失即失败 |

---

402B 验收通过后，方可启动 402C；在此之前禁止服务层引用 SOM 仓储，避免双事实来源。
