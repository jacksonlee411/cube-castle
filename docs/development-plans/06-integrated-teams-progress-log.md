# 06 — 集成团队进展日志（Integrated Teams Progress Log）

最后更新：2025-09-15（晚间增补）

—

## 🎯 本次目标
- 修复时态标记错误与一致性问题，按“API 优先、PG 原生 CQRS、读时派生”的原则，完成列级精简与查询层派生。

—

## ✅ 已完成的变更（Changes）

后端（命令服务：organization-command-service）
- 移除 is_temporal 物理列使用：
  - 删除所有 is_temporal 的读/写/扫描；重算仅维护 end_date、is_current。
  - 文件：
    - repository/temporal_timeline.go：去除 is_temporal 更新与列位；重算仅更新 end_date + is_current。
    - repository/organization.go：所有 INSERT/UPDATE/SELECT 去掉 is_temporal 列位与扫描。
    - types/models.go：移除 IsTemporal 字段（请求/响应/模型）。
    - utils/validation.go：移除 isTemporal 校验分支，保留有效期基本校验。
    - handlers/organization.go：审计 AfterData 不再包含 isTemporal/isCurrent（动态字段统一剔除）；响应不含 isTemporal。

后端（查询服务：organization-query-service）
- 统一派生：
  - isTemporal = (endDate != null)；不读取任何存储列。
  - isFuture = (effectiveDate > 今日北京时区自然日)。
  - 文件：cmd/organization-query-service/main.go（新增 IsFuture/IsTemporal 派生、删除 db:is_future 与任何 is_temporal 扫描）。

数据库（PostgreSQL）
- 新增/执行迁移：
  - 023_audit_exclude_dynamic_temporal_flags.sql：审计触发函数过滤 is_current/is_temporal/is_future（动态字段不入审计）。
  - 024_remove_is_temporal_column.sql：删除 is_temporal 列，清理依赖索引/视图并重建简化视图 organization_temporal_current（仅 is_current=true）。
  - 025_remove_is_future_column.sql：删除 is_future 列与相关触发器/索引，重建视图。
- 日切脚本调整：scripts/daily-cutover.sql 去除 is_future 逻辑，保留 is_current 翻转（isFuture 改为读时派生）。
- 运行维护脚本：
  - sql/inspection/check_temporal_consistency.sql：移除 is_temporal 一致性检查；保留“当前态不应结束”“区间重叠/间隙”。
  - sql/maintenance/fix_temporal_timeline_continuity.sql：仅重算 end_date 连续性。
  - scripts/maintenance/run_temporal_consistency.sh：自动跳过无 is_temporal 时的对齐修复；支持 fix-timeline/fix-all。

契约与文档
- GraphQL 契约说明更新：docs/api/schema.graphql
  - 指明 isTemporal 为派生（endDate!=null），isFuture 以北京时间（UTC+8）派生。
- 架构文档更新：docs/development-plans/02-technical-architecture-design.md
  - 时态字段为 effective_date、end_date、is_current；is_temporal/is_future 已移除并由查询层派生。
- 本进展文档重写并归档此前结论。

—

## 📈 当前进展与验证（Progress）
- 迁移 023/024/025 已在本地数据库成功执行；organization_units 不含 is_temporal/is_future。
- 两个服务可构建通过：
  - go build ./cmd/organization-command-service
  - go build ./cmd/organization-query-service
- 一键巡检（check-and-fix）通过：
  - 区间重叠/间隙/“当前态结束”检查均为 0。
- 典型数据点（code=1000002）：
  - 当前态（2025-09-09）尾部开放；历史均 end_date 非空；isTemporal 派生与 endDate 一致。

—

## 🧩 增补（2025-09-15 晚间）— 审计统一与监控修复

问题与处置（根因 → 措施 → 状态）
- 创建版本 500（INSERT 触发器读取 NEW.operation_type）
  - 移除/替换行级审计触发器（迁移 027/028），统一改为应用层审计；状态：已修复
- 监控报错引用 is_future 列
  - TemporalMonitor 查询改为派生表达式（未来/历史）；状态：已修复
- 暂停/激活失败（INSERT 列数不匹配）
  - 纠正 INSERT 列位数量，时间轴重算 OK；状态：已修复
- 审计页签“变动信息”不显示
  - 应用层审计补齐 modifiedFields/changes 写入（创建版本/暂停/激活/删除），并显式 ::jsonb 入库；状态：新生成记录已显示

验证要点
- POST /api/v1/organization-units/{code}/versions：201，时间轴闭合前一条，is_current 重算正确
- /suspend 与 /activate：200，插入新版本并全链重算；审计记录包含 status 变更
- GraphQL auditHistory(recordId)：返回 modifiedFields 与 changes，前端显示“变更字段/字段变更表”

交付物与路径
- 迁移
  - database/migrations/027_fix_audit_trigger_op_type_from_tg_op.sql（op_type 由 TG_OP 推导）
  - database/migrations/028_drop_legacy_row_audit_trigger.sql（彻底移除 log_audit_changes()/audit_changes_trigger）
- 监控
  - cmd/organization-command-service/internal/services/temporal_monitor.go（派生查询）
- 审计（应用层）
  - cmd/organization-command-service/internal/audit/logger.go（::jsonb、resourceId 兜底、变更明细写入）
  - cmd/organization-command-service/internal/handlers/organization.go（CREATE_VERSION 变更明细）
  - cmd/organization-command-service/internal/repository/temporal_timeline.go（列位修复）
- 维护脚本
  - database/maintenance/2025-09-15_backfill_audit_modified_fields.sql（历史记录回填 modified_fields/changes）
- 回归脚本
  - scripts/tests/test-temporal-regression.sh（创建版本/调整生效日/作废中间版本/暂停/激活 + 时间线校验）

运维建议（上线步骤）
1) 执行迁移 027/028（先修触发器，再移除）
2) 重启命令服务/查询服务
3) 视业务需要，在环境中执行历史回填脚本（仅补 NULL 的 modified_fields/changes）
4) 运行回归脚本验证核心路径（或在 CI 以手动任务方式触发）

—

## 🔜 后续任务（Next Tasks）— 晚间增补
- 文档
  - docs/api/CHANGELOG.md 已补充 v4.6.2；在运维手册增加“触发器退场/应用层审计上线与回滚”说明
- CI/质量门禁
  - 将 scripts/tests/test-temporal-regression.sh 加为可选 E2E 回归任务（手动触发），并汇总输出报告
- 历史数据治理（可选）
  - 评估并在各环境执行审计回填脚本；确认前端“变更字段/变更表”显示完整
- 监控
  - 增加 Prometheus 指标/健康端点用于观测审计与时态一致性（重复当前/缺失当前/重叠计数）

—

## 🔜 后续任务（Next Tasks）
- GraphQL 契约细化（可选）：
  - 如产品需要对外暴露 isTemporal 字段，按派生语义补充分发并更新 schema 描述；当前已提供 isFuture 的派生说明。
- API 迁移指南与变更记录：
  - docs/api/migration-guide.md 与 docs/api/CHANGELOG.md 增补“移除 is_temporal/is_future 物理列，改为派生”的说明及影响面（前端、报表 SQL）。
- CI 质量门禁：
  - 将“时态连续性巡检”结果汇总至报告，必要时设为告警阈值（可集成 iig-guardian 汇总）。
- 前端对齐：
  - 使用返回的 endDate/isCurrent 动态派生 UI 显示的 isTemporal/isFuture；移除任何对物理列的假设。
- 脚本与样例清理：
  - 已更新示例数据（去 is_future 列位）；检查并移除遗留 SQL/脚本中的 is_future/is_temporal 引用（备份/历史脚本可保留但不在 CI 路径执行）。

—

## ⚖️ 原则与对齐（Alignment）
- API 优先与 CQRS 分工：查询统一 GraphQL（读时派生），命令统一 REST（写侧维护 end_date/is_current）。
- 不引入触发器至热路径：列移除+读时派生降低热路径复杂度与风险。
- 命名一致性：API camelCase；DB snake_case；跨层字段对齐。

—

## 📎 参考路径（References）
- 命令服务：cmd/organization-command-service/internal/{repository,handlers,types,utils}
- 查询服务：cmd/organization-query-service/main.go
- 迁移脚本：database/migrations/023、024、025
- 运维脚本：scripts/maintenance/run_temporal_consistency.sh
- 巡检/修复 SQL：sql/inspection/check_temporal_consistency.sql，sql/maintenance/fix_temporal_timeline_continuity.sql

—

## 变更记录（Changelog）
- 2025-09-15：执行阶段B（早期直切）：移除 is_temporal 物理列；查询层改为派生 isTemporal。
- 2025-09-15：清理 is_future 物理列；查询层改为派生 isFuture（北京时间）；同步更新脚本、契约与文档。
