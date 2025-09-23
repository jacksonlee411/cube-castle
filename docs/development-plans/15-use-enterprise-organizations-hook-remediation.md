# 15 — useEnterpriseOrganizations Hook 标记冲突整改

文档类型：整改记录
创建日期：2025-09-23
责任团队：前端平台组（主责）＋ 实现清单守护代理（监督）
优先级：P0（已过期临时实现）
当前状态：已关闭（P0 缺口回收完毕）

---

## 1. 背景
- `docs/development-plans/10-implementation-inventory-maintenance-report.md` 将 “useEnterpriseOrganizations Hook 标记冲突” 列为 P0 紧急事项：主 Hook 顶部声明将在 2025-09-16 后删除，与其在组织核心页面中的正式依赖相矛盾，导致实现清单与代理审计误判。
- IIG 报告要求立即澄清真实的临时实现对象，并更新迁移计划与截止日期以恢复治理准确性。

---

## 2. 问题诊断
1. `frontend/src/shared/hooks/useEnterpriseOrganizations.ts`（原 461 行）存在 `// TODO-TEMPORARY: 该Hook将在 2025-09-16 后删除`，指向自身，造成“正式实现被标记为待删除”的误导。
2. 真实的临时逻辑位于兼容封装 `frontend/src/shared/hooks/useOrganizations.ts`，该文件提供过渡层以缓冲旧调用但已无业务引用。
3. 兼容封装的存在导致治理脚本持续报警，并阻碍实现清单关闭相关条目。

---

## 3. 整改措施
1. **纠正标记位置**：将主 Hook 内的 `TODO-TEMPORARY` 替换为常规说明，强调其为正式实现；新增注释见 `frontend/src/shared/hooks/useEnterpriseOrganizations.ts:461`。
2. **回收临时兼容层**：因业务模块已完成迁移，直接删除 `frontend/src/shared/hooks/useOrganizations.ts` 并移除 `shared/hooks/index.ts` 中的导出；同时更新 `frontend/HOOK_MIGRATION_REPORT.md` 记录统一入口状态。
3. **验证治理脚本**：执行 `bash scripts/check-temporary-tags.sh`，确认仓库不存在不合规的 `TODO-TEMPORARY` 标记，确保 CI 通过。

---

## 4. 验证结果
- `frontend/src/features/organizations/OrganizationDashboard.tsx` 等核心模块仍正常依赖 `useEnterpriseOrganizations`，功能未受影响。
- `rg "useOrganizations" frontend/src -g"*.ts"` 仅在历史说明注释中出现，确认代码路径已无兼容层实现。
- `bash scripts/check-temporary-tags.sh` 返回绿色，表示治理脚本已通过。

---

## 5. 后续计划
（无）整改完成，无额外行动项。

---

## 6. 完成判定
- 主 Hook 不再带有误导性 `TODO-TEMPORARY`。
- 兼容封装 `useOrganizations.ts` 已删除，聚合导出保持单一入口。
- 治理脚本通过，且无业务代码依赖历史兼容层。
- IIG 可在下一次巡检中正式归档该整改记录。
