# Plan 247 – Temporal Entity 文档与治理对齐

**关联主计划**: Plan 242（T5）  
**目标窗口**: Day 15-16  
**范围**: README/Quick Reference/Implementation Inventory 及 Plan 215 执行日志更新

## 背景
- Phase2（Plan 215）要求在 Plan 222 验证阶段完成核心文档更新（docs/development-plans/215-phase2-summary-overview.md:250-269）  
- Plan 242 改变了命名与入口，需要同步至所有权威文档与日志

## 工作内容（修订后）
1. 指南重命名（Plan 242 对齐）
   - 将“旧 Positions 指南”重命名并改写为中性抽象的 `docs/reference/temporal-entity-experience-guide.md`（不保留 reference 目录下旧名文件的内容副本，避免第二事实来源）。
   - 同步更新引用：Plan 06、Plan 240、Plan 241、Plan 242（本身）以及 `docs/reference/00-README.md` 导航。
2. 核心文档一致性
   - `README.md`：修正 Go 版本基线为 1.24.x（与 `toolchain go1.24.9`、CLAUDE.md/AGENTS.md 一致），并补充 Temporal Entity 抽象的文档入口。
   - `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`：补充命名抽象与统一入口说明（仅指向权威路径，不复制细节）。
   - `docs/reference/02-IMPLEMENTATION-INVENTORY.md`：以“入口链接 + 生成器快照”方式指向，不手工扩写可变实现清单；提交前运行脚本生成最新快照并附证据。
3. 执行日志闭环
   - 在 `docs/development-plans/06-integrated-teams-progress-log.md` 与 `docs/development-plans/215-phase2-execution-log.md` 填写命名迁移里程碑与证据路径，形成 Phase2 治理闭环（Plan 242 / T5）。
4. 证据与审计
   - 完成 `reports/plan242/naming-inventory.md` 最终版；新增 `logs/plan242/t5/` 承载本次迁移的校验输出（命令行输出、diff、脚本日志）。
5. 沟通与公告
   - 向计划 240/241/242/06 提交更新摘要；若存在外部对旧路径的依赖，在 `CHANGELOG.md` 发布迁移提示（不保留 reference 下旧名的“平行内容”）。

## 范围与对象清单
- Active 文档/计划（纳入“零引用”校验范围）：
  - `docs/reference/temporal-entity-experience-guide.md`（新）
  - `docs/reference/00-README.md`（导航项从 Positions 更名为 Temporal Entity）
  - `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - `docs/reference/02-IMPLEMENTATION-INVENTORY.md`
  - `docs/development-plans/06-integrated-teams-progress-log.md`
  - `docs/development-plans/240-position-management-page-refactor.md`
  - `docs/development-plans/241-frontend-framework-refactor.md`
  - `docs/development-plans/242-temporal-naming-abstraction-plan.md`
  - `README.md`
- 排除项（历史保留，不纳入“零引用”校验统计）：
  - `docs/archive/**`（历史记录允许保留旧名引用，以便溯源；如需纠偏，另起归档修订 MR）

## 执行步骤与校验（可复制执行）
1) 准备与变更
   - 新建并改写：`docs/reference/temporal-entity-experience-guide.md`
   - 更新引用与导航：按“范围与对象清单”逐项改动
2) 生成实现清单快照
   - `node scripts/generate-implementation-inventory.js`
   - 产出与对比：`reports/implementation-inventory.md`、`reports/implementation-inventory.json`（提交 MR 作为证据）
3) 一致性与结构校验
   - “零引用”（排除归档）：
     - 使用 ripgrep 搜索旧指南文件名（此处以占位符表示）：`rg -n '<OLD_GUIDE_FILE_NAME>' --glob '!docs/archive/**'`
     - 预期：无输出（返回码 1）
   - 文档同步与架构守护：
     - `node scripts/quality/document-sync.js`
     - `node scripts/quality/architecture-validator.js`
   - 前端质量校验（如工作流要求）：
     - `npm run lint` 或 `npm --prefix frontend run lint`
4) 证据落盘
   - 创建目录：`logs/plan242/t5/`
   - 保存以下输出为日志文件：
     - `logs/plan242/t5/rg-zero-ref-check.txt`（第3步零引用输出）
     - `logs/plan242/t5/document-sync.log`、`logs/plan242/t5/architecture-validator.log`
     - `logs/plan242/t5/inventory-sha.txt`（实现清单快照文件的 sha256sum）
5) README 基线纠偏（Go 版本）
   - 将 `README.md` 中环境要求的 Go 版本由 “1.23+” 统一修正为 “1.24.x（与 toolchain go1.24.9 对齐）”
6) MR 描述与沟通
   - 在 MR 中附：本计划编号、范围改动、校验命令与关键输出摘要、外链影响说明与回滚路径

## 里程碑 & 验收（可执行）
- Day 16：所有文档/计划 PR 合并，执行日志更新完毕  
- 验收标准（全部满足方可通过）：
  - 零引用：`rg -n '<OLD_GUIDE_FILE_NAME>' --glob '!docs/archive/**'` 无输出；
  - 校验通过：`document-sync.js` 与 `architecture-validator.js` 退出码为 0；
  - 实现清单：提交最新 `reports/implementation-inventory.*` 且在 MR 中粘贴快照生成时间与文件大小（哈希）；
  - README/Quick Reference/Inventory 均指向 `temporal-entity-experience-guide.md`，且 README 的 Go 版本为 1.24.x；
  - Plan 215 与 Plan 06 已登记迁移里程碑与证据路径（含上述日志文件）。

## 回滚策略与外链影响
- 不在 `docs/reference/` 保留旧名内容副本，避免第二事实来源；对外链接兼容通过发布 CHANGELOG 通告完成。
- 如出现未预期外链中断，可临时在下一补丁版 MR 中提交旧指南路径的“仅重定向占位符”（不含业务内容，首行以“Deprecated – moved to ...”说明，禁止复制正文），并设定一迭代后的移除计划（在本计划附录登记）。

## 产出物清单
- 新/改文档：`docs/reference/temporal-entity-experience-guide.md`、`README.md`、`docs/reference/00-README.md`、`01-DEVELOPER-QUICK-REFERENCE.md`、`02-IMPLEMENTATION-INVENTORY.md`、`docs/development-plans/06-integrated-teams-progress-log.md`、`215-phase2-execution-log.md`、`240-*.md`、`241-*.md`、`242-*.md`
- 校验与证据：`logs/plan242/t5/*.log`、`reports/implementation-inventory.*`
- 公告：`CHANGELOG.md`（对外链接与命名迁移说明）

## 风险与缓解
- 误留旧名引用 → 零引用脚本校验 + MR 审查清单；
- 生成器快照遗漏 → 强制在 MR 模板中勾选“已运行 generate-implementation-inventory 并提交产物”；
- 外链断裂 → CHANGELOG 公告 +（可选）一次性短期占位符，严格截止时间；
- README 与 CLAUDE/AGENTS 基线漂移 → 在验收脚本中新增 `rg -n 'Go 1\\.23\\+' README.md`（预期无输出）。

## 汇报
- 完成后在 Plan 242 主文档和 `215-phase2-execution-log.md` 登记；归档文件存于 `docs/archive/development-plans/247-temporal-entity-docs-alignment-plan.md`。
