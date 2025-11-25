# 410 · 命令层 SOM 写路径彻底切换

**关联计划**：Plan 400、402A~402D、Plan 203/204/206  
**状态**：执行中 - 阶段 1（P2 Feature Flag 清理）  
**范围**：命令服务与组织/职位模块彻底迁移到 `standard_objects*`，旧表仅保留只读视图，Feature Flag 永久移除  
**日志要求**：`logs/plan402/migration/*.log`、`logs/plan410/command/*.log`、`logs/plan410/validator/*.log`、`logs/plan410/runbook/*.log`

> 使命：在不保留任何兼容代码或 Feature Flag 的前提下，让命令服务全部以 SOM 三表为唯一事实来源，旧 `organization_units` / `positions` 表仅作为只读视图或快照源。

---

## 0. 进展速记（2025-11-17）

- ✅ Plan 402 系列已在命令层完成 SOM 双写：`internal/organization/handler/standard_object_adapter.go` 与 `internal/organization/service/position_standard_object_adapter.go` 的 `upsertStandardObject` / `syncPositionStandardObject` 已在创建、更新、版本与事件流程中统一调用 `standardobject.ObjectService`。
- ✅ 命令入口与模块构造现已要求显式注入 SOM Service（`cmd/hrms-server/command/main.go`、`internal/organization/api.go`）；`OrganizationHandler`、`PositionService` 也在构造时强制依赖（`internal/organization/handler/organization_base.go`、`internal/organization/service/position_service.go`），彻底移除了 `standardobject.NewNoopService()` 的自动回退。
- ✅ `internal/standardobject/adapter/sqlc` 缺失 DB 句柄时会返回错误并阻断启动，满足“依赖缺失即失败”的约束。
- ✅ 通过 Goose 迁移 `20251215090000_create_standard_object_views.sql` 创建了 `organization_units_v` / `positions_v` 只读视图，分别将 SOM kernel/version/audit 展平成 legacy schema 需要的主字段，供查询层与指标脚本在后续迭代中切换引用。
- ✅ 通过补丁迁移 `20251215100000_refresh_standard_object_views.sql` 将视图中的 `created_at/updated_at` 对齐 version 记录，保持与 legacy 仓储一致的排序/审计语义。
- ✅ 仓储层抽象出 SOM Port：`internal/organization/repository/organization_standardobject.go` 与 `.../position_standardobject.go` 新增 Upsert 能力，`OrganizationHandler`/`PositionService` 仅依赖仓储完成聚合构建与 Upsert，为 P0 仓储重写奠定统一入口。
- ✅ 查询层开始接入视图：`internal/organization/repository/postgres_organization_details.go`、`.../postgres_organizations_list.go`、`.../organization_hierarchy*.go` 以及空岗统计入口均已切换至 `organization_units_v`，验证 SOM 视图可直接满足 GraphQL/REST 读路径。
- ✅ 命令侧 Upsert 汇聚到仓储：`internal/organization/repository/organization_standardobject.go` 现承担组织聚合生成与 SOM Upsert，`OrganizationRepository` 外部仅调用该接口，为 P0 的“命令仓储改写”建立统一写入口。
- ⚠️ 尚存风险：`OrganizationRepository.Create/Update/Suspend/Activate/SoftDelete` 及 timeline/层级维护仍直接写 `organization_units`（含 `generateCode`、`temporal_timeline_*` 等），与 P0“命令仓储彻底迁移到 SOM”目标不符；后续需分阶段重写这些写路径，并为 `standard_objects` 设计统一的编码冲突守卫与回滚策略。

---

## 1. 目标

1. 命令层仓储（组织、职位、层级、调度、级联）全面改写为 SOM Port/Repository，不再直接访问 legacy 表。
2. 数据库提供只读视图/物化视图，兼容现有查询接口，生成记录与路径字段由 `standard_object_links`/`hierarchy_snapshots` 驱动。
3. Feature Flag（`organization.useLegacyStore` / `position.useLegacyStore`）彻底移除，`plan402_runtime_flags` 仅用于运行报告，不再控制流程。
4. 命令服务所有写操作在 `plan402_runtime_flags.enforce_legacy_lock=true` 下通过，`standardobject-validator` 在锁定模式下 0 差异。
5. Runbook、回滚策略与监控统一更新，确保 30 分钟内可回滚到 SOM 快照或恢复 legacy 视图。

---

## 2. 工作项

### P0 · Repository/Service 重构
- 以 `internal/standardobject` Repository 为底座，新增组织/职位专属 Adapter，提供聚合生成、层级元数据注入、版本裁剪、链接维护等能力。
- 组织仓储需要支持：
  - Kernel ↔ Domain 映射（`types.Organization` ↔ `standardobject.ObjectAggregate`）。
  - 层级/路径计算迁移到 `standard_object_links` + `standard_object_hierarchy_snapshots`。
  - `TemporalTimeline*`、`scheduler`、`Cascade`、`validator` 全部基于 SOM 读取。
- 职位仓储同理，包含 assignments/outbox 事件。
- 删除 `organization_units` / `positions` 所有 INSERT/UPDATE/DELETE 语句；保留 VIEW 或兼容层用于少数遗留查询。

### P1 · 数据库与视图
- 构建只读 VIEW / MATERIALIZED VIEW：
  - `organization_units_v`：从 `standard_objects` + `standard_object_versions` + `standard_object_links` 拼出旧字段。
  - `positions_v`：同理拼出职位旧字段。
  - VIEW 仅用于查询层、临时指标，所有写入禁止。
- 更新 `drop-legacy-write-paths.sql`、`plan402_runtime_flags`，让锁脚本只验证 VIEW，不再管理真实表。
- 运行 `sqlc generate` / Goose 迁移，记录日志 `logs/plan410/command/sqlc-*.log`。

### P2 · Feature Flag 清理与注入
- 移除 `standardobject.NewNoopService()`、`featureflag` 相关注入；命令服务入口直接依赖 SOM Port。
- 删除 `config/feature-flags.yaml` 中的 legacy 配置，`internal/standardobject/featureflag` 保留仅供工具使用。
- 更新 `cmd/hrms-server/command/main.go`，将 SOM Service 作为必选依赖，启动失败即 panic。

### P3 · 双写/回滚策略
- 迁移 `scripts/plan402/export-standardobject-snapshot.sh`、`database/scripts/plan402/checkpoints/replay_view.sql` 的使用说明到 Plan 410。
- 设计新的回滚脚本：从快照导入 → 建立 VIEW → 切换 Feature Flag（已废弃）→ 重放 transaction range；在 staging 完成演练。
- 输出 Runbook（R1~R7），记录 `logs/plan410/runbook/*.log`。

### P4 · 验证与门禁
- 运行 `standardobject-validator`、`make test`、`make test-db`、`npm run quality:preflight`，在 `plan402_runtime_flags.enforce_legacy_lock=true` 下连续执行并记录日志 `logs/plan410/command/*.log`。
- 新增观测指标：`som_command_write_latency`、`som_command_error_rate`，同时更新 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`。
- 交付回归报告，列出所有日志文件。

---

## 3. 交付物

- `internal/organization/**`、`internal/position/**` 及相关调度/级联模块的 SOM 实现与测试。
- 数据库 VIEW/迁移脚本、`sqlc` 生成物日志、更新后的 `drop-legacy-write-paths.sql`。
- Feature Flag 清理与命令服务启动配置 diff。
- Runbook、回滚脚本、监控指标文档，日志存放 `logs/plan410/{command,runbook,validator}`。
- 兼容层/查询视图文档，说明如何访问 legacy 视图。

---

## 4. 验收标准

1. 命令服务写入仅触达 `standard_objects*`，旧表只读视图无任何 DML；`plan402_runtime_flags` 全程保持 `true`，`POST/PUT` 成功。
2. 所有命令层测试在锁定模式下通过（`make test`、`make test-db`、`standardobject-validator`）。  
3. `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 与 Runbook 更新完毕，记录最新日志文件。  
4. 回滚演练（快照导出 → revert → replay → enable lock）在 30 分钟内完成，日志归档并引用。  
5. 监控指标与合成探针（SOM 写入、validator diff、transaction gap）上线。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 仓储切换导致查询/层级逻辑失效 | 命令服务不可用 | 在 staging 引入 VIEW + 兼容层，先以特性开关验证再移除旧实现；提供大量集成测试。 |
| 回滚流程不完整 | 切换失败无法恢复 | 强制演练快照导出与 replay SQL，Runbook 标注负责人、耗时与日志。 |
| Observability 盲区 | 迁移后缺监控 | 新增 SOM 写入指标 & 合成探针，纳入 Plan 272 收集。 |

---

## 6. 依赖与事实来源
- `docs/api/openapi.yaml`、`docs/api/schema.graphql`
- `database/migrations/20251201090000_create_standard_objects.sql`
- `internal/standardobject/**`、`cmd/tools/standardobject-*`
- Plan 400 / 402 系列、`docs/reference/standard-object-evidence-guide.md`
