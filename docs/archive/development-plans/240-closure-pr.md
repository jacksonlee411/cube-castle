# PR 描述（Plan 240 关单）

标题：feat/docs/ci: finalize Plan 240; archive 240BT; introduce temporal layout shell; unify position selectors; add 240E CI

---

## 目的
- 完成并归档 Plan 240 全量交付，统一文档与证据；为 241 收尾预留回补清单；增加最小骨架以准备共享框架合流；完善 240E 回归工作流与本地一键脚本。

## 变更范围
- 代码
  - 新增最小骨架并接入路由（仅性能标记，无 DOM/testid/契约变更）  
    - `frontend/src/features/temporal/layout/TemporalEntityLayout.tsx`  
    - `frontend/src/features/temporal/pages/organizationRoute.tsx`  
    - `frontend/src/features/temporal/pages/positionRoute.tsx`
  - 职位域选择器统一（SSoT），新增 `position.form(mode)`  
    - `frontend/src/shared/testids/temporalEntity.ts`  
    - `frontend/src/features/positions/components/PositionForm/index.tsx`  
    - `frontend/src/features/positions/components/dashboard/PositionHeadcountDashboard.tsx`  
    - `frontend/src/features/positions/components/dashboard/PositionVacancyBoard.tsx`  
    - `frontend/src/features/positions/components/transfer/PositionTransferDialog.tsx`
  - 烟测规格使用 SSoT 选择器（组织创建按钮）  
    - `frontend/tests/e2e/basic-functionality-test.spec.ts`  
    - `frontend/tests/e2e/organization-create.spec.ts`
- CI 与脚本
  - 新增 240E 回归工作流与工具脚本  
    - `.github/workflows/plan-240e-regression.yml`  
    - `scripts/plan240/run-240e.sh`  
    - `scripts/plan240/trigger-240e-ci.sh`  
    - `scripts/plan240/record-240e-acceptance.sh`
- 文档与索引
  - Plan 240：标记“已完成（验收通过）”，新增“0.1 影响评估：240 先于 241 完成的回补计划”  
    - `docs/archive/development-plans/240-position-management-page-refactor.md`
  - Plan 240B：标注硬依赖“240BT 路由解耦完成”  
    - `docs/archive/development-plans/240B-position-loading-governance.md`
  - Plan 240BT：验收完成并归档  
    - `docs/archive/development-plans/240bt-org-detail-blank-page-mitigation.md`
  - Plan 240E：登记本地 Smoke + 守卫证据，新增“关闭确认”段落；215 执行日志同步  
    - `docs/development-plans/240E-position-regression-and-runbook.md`  
    - `docs/development-plans/215-phase2-execution-log.md`
  - 文档索引更新：240 为“已完成”；列出 241 子计划（A/B/C）  
    - `docs/development-plans/HRMS-DOCUMENTATION-INDEX.md`
  - 临时标签规范：统一为 `// TODO-TEMPORARY(YYYY-MM-DD): ...`  
    - `AGENTS.md`、相关计划文档与参考手册同步修订

## 验收证据
- 守卫（通过）  
  - `logs/plan240/E/selector-guard.log`  
  - `logs/plan240/E/architecture-validator.log`  
  - `logs/plan240/E/temporary-tags.log`
- Smoke（Chromium）：6 passed / 1 skipped（通过）  
  - `logs/plan240/E/playwright-smoke-20251115142851.log`
- 240BT 冒烟与健康检查（通过）  
  - `logs/plan240/BT/smoke-org-detail.log`  
  - `logs/plan240/BT/health-checks.log`

## 验证指南（本地）
```bash
make docker-up && make run-dev
cd frontend && PW_SKIP_SERVER=1 npm run test:e2e:smoke
# 守卫产物：logs/plan240/E/*.log
# Smoke 日志：logs/plan240/E/playwright-smoke-*.log
```

## 兼容性与风险
- 无后端契约/数据库变更；前端仅引入极薄 `TemporalEntityLayout.Shell`，对 DOM/testid 不造成破坏性影响。  
- 风险：布局骨架包裹与可观测性标记引入最小开销；已通过 Smoke 与守卫验证。

## 回滚方案
- 如需临时回退骨架包裹，回滚提交 `21c81362`（layout shell 接入）；选择器统一回滚提交 `14456b34`。  
- 240E CI 若阻塞，可先使用 `scripts/plan240/run-240e.sh` 本地产出证据并登记。

## 与 AGENTS.md 的一致性
- 单一事实来源：契约/事件/选择器均引用唯一入口（OpenAPI/GraphQL、temporal-entity-experience-guide、temporalEntitySelectors）。  
- Docker 强制：CI/脚本均基于 docker-compose；未调整容器端口映射。  
- CQRS：无命令/查询边界调整；仅前端查询侧与测试资产。

## 241 对接回补（已在 240“0.1 影响评估”登记）
- 骨架切换到共享 Layout（不改契约，不复制规范正文）  
- Hook/Loader 统一到 241 的唯一入口并回收薄适配  
- 可观测性归一至骨架；页面仅保留必要事件  
- Feature Flag 收敛与 E2E 复跑；守卫/架构/临时标签全部通过
