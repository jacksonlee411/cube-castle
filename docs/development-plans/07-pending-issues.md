# 07 — Pending Issues（英文文件名已规范化）

最后更新：2025-09-14  
维护团队：后端组（主责）+ 架构组 + QA组  
状态：分析完成（待验证与修复）

—

## 1. 背景与现象
- 业务对象：组织详情（业务编码 `1000002` 测试部门E2E2）
- 操作：删除了生效日为 2025-06-01 的版本记录
- 现象：生效日为 2025-05-01 的上一条记录的 `endDate` 未自动更新（未回填为 2025-05-31）
- 期望：删除中间版本后，系统应自动桥接相邻版本：`prev.endDate = next.effectiveDate - 1`；若删除尾部版本，则前一条 `endDate = null` 且设为当前

—

## 2. 结论（根因）
- 未走“版本级删除 + 时间轴重算”的应用层路径，因而没有触发相邻边界重算；数据库层面的自动触发器已被移除，不再负责 `endDate` 回填。
  - 路由未提供按 code 的物理删除；正确的删除需按 recordId 进行版本作废，并统一触发时间线重算。

—

## 3. 关键证据（代码与脚本）
- 命令服务路由（无 `DELETE /api/v1/organization-units/{code}`，存在版本与事件接口）
  - cmd/organization-command-service/internal/handlers/organization.go:719
- 版本作废事件处理：将目标记录软删除（`status='DELETED'` 且 `deleted_at` 非空），随后调用统一重算
  - cmd/organization-command-service/internal/handlers/organization.go:736
- 时间轴重算算法：按生效日升序，回填每条 `end_date = 下条.effective_date - 1`，尾部 `end_date = NULL`，并唯一设置 `is_current`
  - cmd/organization-command-service/internal/repository/temporal_timeline.go:286
  - cmd/organization-command-service/internal/repository/temporal_timeline.go:339
- 触发器清理：历史上用于自动更新 `end_date` 的触发器已删除，防止与应用层算法冲突
  - database/maintenance/cleanup_legacy_triggers.sql:8
- 正确用法示例脚本（演示错误 vs 正确删除方式，并校验“无断档、尾部开放、唯一当前”）
  - cmd/organization-command-service/test_correct_deletion_api.sh:121

—

## 4. 影响面与风险
- 直接 SQL 删除、或误用已弃用的物理删除语义，都会导致时间线不重算，出现：
  - 上一条 `endDate` 未回填（断档）
  - 尾部 `endDate` 未清除（尾部未开放）
  - `is_current` 唯一性可能未恢复（在极端场景下）
- 查询与统计依赖 `status != 'DELETED' AND deleted_at IS NULL` 的过滤，若仅改动其一（只改 status 或只设 deleted_at），也会造成异常。

—

## 5. 正确操作（面向业务/用户）
- 删除某个版本（推荐）：调用“事件端点（作废版本）”，系统统一重算时间线
  - `POST /api/v1/organization-units/{code}/events`
  - 请求体示例：`{"eventType":"DEACTIVATE","recordId":"<UUID>","changeReason":"数据纠正"}`
- 版本创建（插入中间版本）：`POST /api/v1/organization-units/{code}/versions`（插入后同样自动重算）

—

## 6. 一次性修复建议（已误删的情况）
- 用统一服务重算该 code 的时间线（会回填全部相邻边界并修正当前态）：
  - 代码路径：cmd/organization-command-service/internal/services/temporal.go:476
  - 方法：`TemporalService.RecomputeTimelineForCode(ctx, tenantID, code)`
- 若需外露为运维接口，可在命令服务增加受限的重算端点（仅管理员权限）。

—

## 7. 验证清单
- 时间断档检查（应为 0）：见脚本校验
  - cmd/organization-command-service/test_correct_deletion_api.sh:171
- 尾部开放检查（最后一条 `endDate IS NULL`）：
  - cmd/organization-command-service/test_correct_deletion_api.sh:204
- 当前态唯一性检查（应为 1）：
  - cmd/organization-command-service/test_correct_deletion_api.sh:233

—

## 8. 风险控制与制度化改进
- 禁止直接对 `organization_units` 进行手工 `DELETE` 或跳过应用层的删除路径。
- 在 DBA/脚本侧加入“危险操作”守护（触发器来源扫描、禁变更名单）。
- 在命令服务日志中对“物理删除尝试”输出高亮告警（若暴露相关端点或检测到异常）。
- QA 增加回归用例：删除中间/首条/尾条版本后的时间线一致性验证。

—

## 9. Definition of Done（验收）
- 复测用正确删除路径后：
  - 前一条 `endDate = 下一条.effectiveDate - 1`；
  - 尾部 `endDate = NULL`；
  - `is_current` 唯一且匹配“最后一个生效日 ≤ 今天”的那条。
- SQL 校验三项均通过（断档=0、尾部开放=1、当前态=1）。

—

## 10. 变更记录
- 2025-09-14：新增本条 Pending Issue，明确根因与修复路径；统一到事件端点与时间线重算。

—

## 11. 现场验证与修复记录（code=1000002）

- 环境与租户：tenantId=`3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
- 修复前查询（非删除版本，按生效日升序）：
  - 2025-04-01 → 2025-04-30（ACTIVE, is_current=false）
  - 2025-05-01 → 2025-05-31（ACTIVE, is_current=false）
  - 2025-09-01 → NULL（ACTIVE, is_current=true）
  - 问题：中间 2025-06-01 版本被删除后，上一条 2025-05-01 的 endDate 未桥接到 2025-08-31（下一条生效日前一天）。

- 采取动作（一次性重算该 code 的时间线）：
  - 对 code=1000002 执行“相邻版本边界回填 + 当前态唯一”重算，仅作用于非删除记录；避免触发 `READ_ONLY_DELETED` 保护。
  - 重算规则：
    - `endDate = next.effectiveDate - 1`；尾部 `endDate = NULL`
    - 清除并设置非删除记录的 `is_current`（选取“生效日 ≤ 今天”的最后一条）

- 修复后查询（非删除版本，按生效日升序）：
  - 2025-04-01 → 2025-04-30（ACTIVE, is_current=false）
  - 2025-05-01 → 2025-08-31（ACTIVE, is_current=false）
  - 2025-09-01 → NULL（ACTIVE, is_current=true）

- 结论：时间断档已消除，尾部开放正确，当前态唯一性恢复。

- 佐证 SQL（要点摘录）：
  - 基于窗口函数回填边界：`LEAD(effective_date) OVER (ORDER BY effective_date)`
  - 仅更新非删除记录；清空/设置 `is_current` 也仅限非删除记录

—

## 12. 后续工作与建议

- 操作规范：后续删除版本请使用事件端点触发应用层重算
  - `POST /api/v1/organization-units/{code}/events`，body: `{ "eventType": "DEACTIVATE", "recordId": "<UUID>", "changeReason": "…" }`
- 批量治理（可选）：扫描所有 code 的时间线一致性，发现断档/重叠时调用统一重算
- 监控与守护：加入“物理删除/直改时态字段”检测与审计告警；在 DBA 脚本侧添加保护
