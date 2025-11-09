# Plan 247 – Temporal Entity 文档与治理对齐

**关联主计划**: Plan 242（T5）  
**目标窗口**: Day 15-16  
**范围**: README/Quick Reference/Implementation Inventory 及 Plan 215 执行日志更新

## 背景
- Phase2（Plan 215）要求在 Plan 222 验证阶段完成核心文档更新（docs/development-plans/215-phase2-summary-overview.md:250-269）  
- Plan 242 改变了命名与入口，需要同步至所有权威文档与日志

## 工作内容
1. **指南重命名**：将 `docs/reference/positions-tabbed-experience-guide.md` 改写为 `docs/reference/temporal-entity-experience-guide.md`，更新 Plan 06/Plan 240/Plan 241 引用。  
2. **核心文档**：更新 `README.md`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md`，记录 Temporal Entity 抽象、组件/Hook/测试入口。  
3. **执行日志**：在 `docs/development-plans/06-integrated-teams-progress-log.md` 与 `215-phase2-execution-log.md` 填写命名迁移里程碑，形成 Phase2 治理闭环。  
4. **交互记录**：准备 `reports/plan242/naming-inventory.md` 最终版与 `logs/plan242/t5/`，作为后续审计材料。  
5. **沟通**：向相关计划（240/241/219/222）提交更新摘要，说明抽象完成情况与后续动作。

## 里程碑 & 验收
- Day 16：所有文档 PR 合并，执行日志更新完毕  
- 验收标准：仓库不再存在 `positions-tabbed-experience-guide` 引用；README/Quick Reference/Inventory 反映新的命名；Plan 215 日志记录此次迁移。

## 汇报
- 完成后在 Plan 242 主文档和 `215-phase2-execution-log.md` 登记；归档文件存于 `docs/archive/development-plans/247-temporal-entity-docs-alignment-plan.md`。
