# 70 号文档：组织时间轴全生命周期连贯性调查报告

**创建日期**：2025-10-17  
**状态**：草稿（调查结论已形成，等待后续行动归档）  
**维护人**：集成团队（时态治理小组）

---

## 1. 背景与目标

- **问题起因**：质控例会上提出“组织时间轴在插入/删除历史版本后偶发断档”的风险提示，需要确认全生命周期下的自动重算机制是否覆盖所有场景。
- **调查目标**：梳理触发时间轴变化的全部操作场景，明确实际执行的对接函数与重算策略，验证其与权威实施方案的一致性。
- **范围界定**：仅关注 PostgreSQL 原生 CQRS 架构下的命令侧（REST）写路径，查询侧 GraphQL 读取未纳入本次调查的整改范围。

---

## 2. 权威事实来源

- `docs/architecture/temporal-timeline-consistency-guide.md`：时态时间轴连贯性实施参考 v1.0（定义核心约束与重算流程）。
- `docs/architecture/temporal-consistency-implementation-report.md`：时态数据一致性方案实施报告（记录应用层重算、监控与定时任务落实情况）。
- `cmd/organization-command-service/internal/repository/temporal_timeline_manager.go`：`RecalculateTimeline` / `RecalculateTimelineInTx` 是唯一的时间轴重算实现。
- `cmd/organization-command-service/internal/repository/temporal_timeline_insert.go`：版本插入路径。
- `cmd/organization-command-service/internal/repository/temporal_timeline_update.go`：生效日期调整路径。
- `cmd/organization-command-service/internal/repository/temporal_timeline_delete.go`：版本软删除路径。
- `cmd/organization-command-service/internal/repository/temporal_timeline_status.go`：停用/启用事件写路径。

上述文件互为补充，不引入第二事实来源；任何未来改动需同步更新本调查或直接引用这些权威文件。

---

## 3. 场景覆盖结论

| 场景类别 | 触发操作 | 时间轴影响 | 核心实现 | 备注 |
| --- | --- | --- | --- | --- |
| 首条版本插入 | `POST /api/v1/organization-units` | 建立时间轴起点，`endDate=NULL`，`is_current=true` | `temporal_timeline_insert.go` → `RecalculateTimelineInTx` | 若租户首次创建组织，即刻成为当前版本。 |
| 最新当前版本插入 | `POST /{code}/versions`（当日/过去） | 前一版本 `endDate = new.effectiveDate - 1`，新版本成为当前 | 同上 | 满足“单当前”部分索引约束。 |
| 最新未来版本插入 | `POST /{code}/versions`（未来日期） | 当前版本 `endDate` 回填为未来-1，新版本 `isCurrent=false` 等待日切 | 同上 | 日切任务翻转 `is_current`；保持时间轴连续。 |
| 最早历史版本插入 | `POST /{code}/versions`（早于现有最小值） | 新记录补齐起点，第二条的 `endDate` 重新计算 | 同上 | 场景受 `RecalculateTimelineInTx` 顺序遍历保证。 |
| 中间版本插入 | `POST /{code}/versions`（介于两版本之间） | 前一条 `endDate = new-1`；新条 `endDate = next-1` | 同上 | SQL `FOR UPDATE` 锁定相邻确保无重叠。 |
| 中间版本删除 | `DELETE` / 软删除 | 删除目标，重算后桥接前后 `endDate = next-1` | `temporal_timeline_delete.go` → `RecalculateTimelineInTx` | 软删除保留审计；时间轴无断档。 |
| 最后版本删除 | 软删除末尾版本 | 新末尾 `endDate=NULL`，若仍≤今天则保持当前；否则清空当前 | 同上 | 重算阶段按最新有效日期重新挑选 current。 |
| 首条记录删除 | 软删除首条历史版本 | 第二条变为新起点，其 `endDate` 已按下一条回填 | 同上 | 不破坏后续区间。 |
| 删除未来版本 | 软删除未生效记录 | 前一条 `endDate` 复位为 NULL 或下一条 -1 | 同上 | 避免时间轴提前截断。 |
| 生效日期调整 | `UpdateVersionEffectiveDate`（等价删旧+插新） | 旧记录标记删除，新记录按新日期插入，重算边界 | `temporal_timeline_update.go` → `RecalculateTimelineInTx` | 阻止 `TEMPORAL_POINT_CONFLICT`，保持区间连续。 |
| 停用事件 | `POST /{code}/suspend` | 根据生效日插入/更新状态版本，重算时间轴 | `temporal_timeline_status.go`（SUSPEND） → 重算 | 幂等：若同日已存在版本则直接更新。 |
| 启用事件 | `POST /{code}/activate` | 同上，恢复 `ACTIVE` 状态 | 同上（REACTIVATE） | 未来生效时保持 `isCurrent=false`。 |
| 整单位软删除 | `DELETE /api/v1/organization-units/{code}` | 全部版本 `status=DELETED`，时间轴对外为空 | **未在本次代码清单中变更** | 需另行确认回收策略。 |
| 手动/定时重算 | 运维任务 / `RecalculateTimeline` 调用 | 全链清空 `is_current`，逐条回填 `endDate` 并重选 current | `temporal_timeline_manager.go` | 用于批量修复与日切作业。 |

> 结论：时间轴连续性依赖 `RecalculateTimelineInTx` 唯一实现，其遍历顺序与边界写回逻辑覆盖了插入、删除、状态事件等全部生命周期场景，符合实施指南的“单当前”“无重叠”“尾部开放”约束。

---

## 4. 风险与观察

- **整单位软删除后恢复路径未列入**：接口会将全部版本标记为 `DELETED`，时间轴为空；恢复需通过导入或事件重新写入版本，建议补充运维手册说明。
- **高并发入住的互斥策略**：应用层仅对同一 `code` 的相邻记录使用 `FOR UPDATE`；若同时有多条写入，仍需依赖外层调用串行保障，后续可评估请求级互斥。
- **未来版本日切依赖定时任务**：需确保 `OperationalScheduler` 与 `TemporalMonitor`（见实施报告）持续运行，否则未来版本无法自动接管 current。
- **重复事实来源警戒**：本报告仅列结论与路径，后续若实现发生较大调整，需以源文件为准同步更新，避免偏离唯一事实来源。

---

## 5. 后续行动建议

1. **文档回链**：在 `docs/architecture/temporal-timeline-consistency-guide.md` 补充对停用/启用事件的示例引用本调查结论（若需）。
2. **运维补充**：在 `docs/development-plans/06-integrated-teams-progress-log.md` 增补“整单位软删除后的恢复步骤”。
3. **监控对齐**：确认 `TemporalMonitor` 的告警阈值涵盖“定时任务未翻转 is_current”的异常。
4. **后续归档**：待以上行动完成后，将本报告移入 `docs/archive/development-plans/70-temporal-timeline-lifecycle-investigation.md` 并标记验收。

---

> **同步要求**：若对时态写路径、重算算法或运维任务进行任何更改，需先更新对应权威文件并回到本计划文档记录一致性校验结果。
