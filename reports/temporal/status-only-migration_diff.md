# Status-Only Migration 差异报告

## 1. 执行信息
- **执行日期**: 2025-09-27
- **操作人**: 系统自动化脚本
- **审批人**: 命令服务团队 (Owner)
- **数据备份路径**: 无需备份（零差异迁移）
- **审计基线**: `reports/temporal/status-only-audit.json`
- **审计对比**: `reports/temporal/status-only-audit-final.json`

## 2. 统计摘要
- 修复前异常记录总数（取自基线 JSON `summary` 节点）: **0条**
  - `status='DELETED' AND deleted_at IS NULL`: **0条**
  - `status<>'DELETED' AND deleted_at IS NOT NULL`: **0条**
- 修复后异常记录总数: **0条**
  - `status='DELETED' AND deleted_at IS NULL`: **0条**
  - `status<>'DELETED' AND deleted_at IS NOT NULL`: **0条**
- 其他异常（如重复 deleted_at）: **0条**

## 3. 操作明细摘要
- 回填 `deleted_at` 记录数: **0条** (无需修复)
- 调整为 DELETED 状态记录数: **0条** (无需修复)
- 清空 `deleted_at` 记录数: **0条** (无需修复)
- 额外手动处理说明: **无需额外操作，数据状态完全一致**

## 4. 审计与日志
- 审计结果: ✅ **数据完全一致，零差异迁移**
- `audit_logs` 审核结论: **无需审计日志修正，状态始终保持一致**

## 5. 剩余风险与后续动作
- 未解决问题: **无**
- 已创建缺陷/任务: **无需创建，迁移成功完成**

---

> 本报告需随迁移计划归档，并在进展日志中记录。 
