# 跨团队进展纪要 — 组织版本删除后 endDate 未更新的最优解落地方案（v1）

最后更新：2025-09-14  
牵头团队：后端组（backend-agent）  
参与团队：架构组、QA  
优先级：P0（功能缺陷整治）

—

## 1. 问题概述（统一口径）
- 场景：组织 code=1000002，删除生效日 2025-06-01 的版本后，生效日 2025-05-01 的上一条记录 endDate 未自动桥接。
- 期望：删除中间版本后，系统自动重算时间线：`prev.endDate = next.effectiveDate - 1`；删除尾部则前一条 `endDate = NULL` 并置为当前。
- 事实：存在通过错误路径（物理删除/直改）或“软删+重算分事务”的用法，未触发统一重算或产生一致性窗口。

—

## 2. 事实与证据
- 正确契约与路由：版本管理通过“版本端点/事件端点”，未开放 `DELETE /api/v1/organization-units/{code}` 物理删除。
  - docs/api/openapi.yaml:74
  - cmd/organization-command-service/internal/handlers/organization.go
- 统一的时态时间线管理器已提供“单事务 删除+全链重算”的落地实现：
  - cmd/organization-command-service/internal/repository/temporal_timeline.go
- 历史触发器已被清理，不再依赖数据库层自动改写 endDate：
  - database/maintenance/cleanup_legacy_triggers.sql
- 现场验证：对 `1000002` 执行一次性重算后，时间断档消失、尾部开放正确、当前态唯一性恢复。
  - docs/development-plans/07-pending-issues.md

—

## 3. 既有报告评审（纠偏）
- 正确点：指出“软删 + 重算分属不同事务”存在原子性风险，需消除一致性窗口。
- 需纠正：最小可行最优解不是新增 `UpdateByRecordIdWithTx`，而是直接复用已存在的事务化时间线管理器进行版本删除与重算，避免实现分叉与职责错配。
- 明确否决：恢复/启用数据库触发器的建议与现行治理冲突，应保持“应用层统一重算”的架构路线。

—

## 4. 最优解（决定）
采用“单事务 版本作废（软删）+ 全链重算”的统一实现路径：
- Handler 层在处理“作废版本（DEACTIVATE）/删除版本”时，改为调用时间线管理器的事务化删除：
  - 现状（需替换）：`repo.UpdateByRecordId(...)` + `temporalService.RecomputeTimelineForCode(...)`
  - 目标（最优）：`timelineManager.DeleteVersion(ctx, tenantID, recordID)`（内部单事务软删+重算）
- 审计日志：保持“不阻断业务”的策略；如需更强一致性，后续以 outbox 事务消息增强，而非强灌入同一事务。
- 运维兜底：保留 `RecomputeTimelineForCode` 作为修复/巡检工具，不走日常链路。
- 坚持无触发器：不恢复任何与 endDate 自动维护相关的 DB 触发器。

—

## 5. 变更明细（实施清单）
- 路由处理：
  - 文件：cmd/organization-command-service/internal/handlers/organization.go
  - 变更：`handleDeactivateEvent` 内部将删除逻辑切换为 `timelineManager.DeleteVersion(...)`；删除成功后返回最新时间线。
- 保留现有 E2E 校验脚本与测试：
  - cmd/organization-command-service/test_correct_deletion_api.sh
  - cmd/organization-command-service/internal/repository/temporal_timeline_test.go
- 不新增仓储接口：不引入 `UpdateByRecordIdWithTx` 等重复能力，避免分叉。

—

## 6. 验收标准（DoD）
- 删除中间版本：上一条 `endDate = 下一条.effectiveDate - 1`。
- 删除尾部版本：上一条 `endDate = NULL` 且唯一 `is_current = true`。
- 整个操作原子化：删除与重算同事务提交/回滚。
- 无触发器依赖：数据库侧无 endDate 自动化触发器。
- E2E/SQL 校验通过：断档=0、尾部开放=1、当前态=1。

—

## 7. 风险与缓解
- 锁竞争/延迟：重算使用按 code 升序 FOR UPDATE，控制锁粒度；必要时缩小事务范围与批量重算节流。
- 并发作废冲突：借助唯一索引（uk_org_temporal_point / uk_org_current_active_only）与事务化重算确保正确性。
- 审计失败：不阻断业务，异步补偿（outbox）。

—

## 8. 测试与回归
- 单测：覆盖“删除首条/中间/尾条”的三类场景；断档/尾部/当前态断言。
- 集成：通过事件端点作废指定 recordId，验证返回的 timeline 与库内一致。
- 回归：运行现有 e2e 脚本与 SQL 校验脚本，确认无回退。

—

## 9. 推进计划
- Day 0：最小实现（handler 切换至 TimelineManager.DeleteVersion）、自测通过。
- Day 1：联调 + E2E 通过；同步 07 文档现场记录。
- Day 2：灰度发布与监控；如有遗留数据，使用重算兜底工具批量修复。

—

## 10. 附：关键参考与路径
- 时间线管理器（事务化删除+重算）：
  - cmd/organization-command-service/internal/repository/temporal_timeline.go
- 事件处理入口：
  - cmd/organization-command-service/internal/handlers/organization.go
- 索引与约束（时间线一致性）：
  - database/migrations/025_temporal_timeline_consistency_indexes.sql
- 触发器清理：
  - database/maintenance/cleanup_legacy_triggers.sql
- 现场修复记录：
  - docs/development-plans/07-pending-issues.md

—

## 11. E2E 验证记录（本地开发环境）

- 服务启动（开发模式）
  - 环境：`DEV_MODE=true`，`JWT_SECRET=dev-secret-2025`，`JWT_ISSUER=cube-castle`，`JWT_AUDIENCE=cube-castle-api`
  - 启动：`go run ./cmd/organization-command-service`
- 令牌获取
  - `POST /auth/dev-token` → 获得 HS256 开发用 JWT
- 创建组织与版本
  - `POST /api/v1/organization-units` → 新 code=`1000007`（示例）
  - `POST /api/v1/organization-units/{code}/versions` 插入 2025-05-01、2025-06-01、2025-09-01（`operationReason` ≥5 字符）
- 作废中间版本
  - `POST /api/v1/organization-units/{code}/events`，`eventType=DEACTIVATE`，`recordId=<2025-06-01 的记录>` → 返回“版本作废成功”
- 验证时间线（数据库核对，仅非删除记录）
  - 2025-04-01 → 2025-04-30（ACTIVE, false）
  - 2025-05-01 → 2025-08-31（ACTIVE, false）
  - 2025-09-01 → NULL（ACTIVE, true）

结论：端到端表现与“单事务 软删+全链重算”预期一致；中间删除自动桥接、尾部开放、当前态唯一均正确。
