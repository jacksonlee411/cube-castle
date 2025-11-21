# Plan 271 – 文档归档一致性守卫（Plan 264-269 复现复盘）

**文档编号**: 271  
**创建日期**: 2025-11-21  
**关联计划**: Plan 264（Workflow 治理）、Plan 265/266/269（Runner & Required Checks）、Plan 270（Workflow Lint）  
**状态**: 🚧 执行中  

---

## 1. 背景

- Plan 264-269 已在 2025-11-20 通过 `b686deab3` 与 `5ad4ddfbc` 归档至 `docs/archive/development-plans/`。
- 2025-11-21 本地 `feat/shared-dev` 再次出现 264-269 文档的“活跃副本”，违背《AGENTS.md》“资源唯一性与跨层一致性”原则，需要立即调查并制定守卫方案。
- 现有 `scripts/quality/doc-archive-check.js` 仅在手动执行 `npm run lint:docs` 时生效，尚未纳入必跑门禁，导致重复文件未被阻断。

## 2. 事件经过

| 时间 (UTC+8) | 事件 | 说明 |
|-------------|------|------|
| 2025-11-20 18:39 | `a6e6cc9d0` | 追加 Plan 264-269 文档（活跃目录） |
| 2025-11-20 20:18 | `b686deab3` | `docs/development-plans/265-269` rename → `docs/archive/development-plans/` |
| 2025-11-20 20:33 | `5ad4ddfbc` | `docs/development-plans/264` rename → `docs/archive/development-plans/`，归档完成 |
| 2025-11-21 08:35 | `731fb8b72` | `feat/shared-dev` 合并 `origin/master`，第二父为 `a6e6cc9d0`，Git 将 Plan 264-269 视为“新增”重新写入活跃目录 |
| 2025-11-21 18:40 | 当前调查 | 发现 `docs/development-plans/264-269` 与 `docs/archive/development-plans/264-269` 同时存在 |

## 3. 根因分析

1. **流程缺陷**：归档提交（`b686deab3`、`5ad4ddfbc`）未及时推送到远端，导致 `origin/master` 仍保留 `a6e6cc9d0`。后续 merge（`731fb8b72`）把旧状态覆盖回来。
2. **守卫缺失**：`npm run lint:docs` 未纳入 Required workflow，任何重复不会触发 CI 阻塞。
3. **目录操作约束未写入 Runbook**：归档动作缺少“先推送/再合并”与“限制 merge 来源”的明确步骤。

## 4. 立即整改（2025-11-21）

- 264-269 文档重新移动至 `docs/archive/development-plans/`，活跃目录只保留 270 及其他进行中计划。
- 在 `agents-compliance.yml` 新增 “Plan 271 Guard – Plan Archive Placement” 步骤，强制执行 `npm run lint:docs`。
- 本文档建立事件记录，并向《00-README》同步 Plan 271 链接，提醒团队遵循新的归档守卫。
- `2025-11-21 22:14 CST` 执行 `npm run lint:docs`（`scripts/quality/doc-archive-check.js`），输出 `✅ 文档计划目录检查通过：活跃/归档无重复文件`。

## 5. 防范方案

1. **流程守卫**  
   - 归档动作仅在 `feat/shared-dev` 上执行，完成后立即 `git push origin feat/shared-dev` 并通过 PR 合入 master，确保远端同步。
   - 禁止将 `origin/master` merge 回 `feat/shared-dev`，除非使用“只读工作树”验证（参照 AGENTS.md）。确需同步时只能 `git fetch` + `git merge --ff-only`。
2. **自动守卫**  
   - `agents-compliance` 工作流运行 `npm run lint:docs`，如检测到重复文件立即失败，阻止 PR 合入。
   - 在本地提交前执行 `npm run lint:docs`（可加入 `make lint` 或 `npm run quality:preflight`），确保持续开发也遵循守卫。
3. **治理文档**  
   - 在《00-README》新增“归档守卫”章节，要求：归档后 30 分钟内在 `docs/archive/ARCHIVE-RECORDS` 记录摘要、Plan 271 记录 runbook。
   - `scripts/quality/doc-archive-check.js` 去除临时豁免（Plan 06 等）后转入常态守卫，由本计划跟踪回收节奏。

## 6. 验收标准

- [ ] `docs/development-plans/264-269` 仅存在于归档目录（活跃目录无重复）。  
- [ ] `agents-compliance` 工作流强制执行 `npm run lint:docs`，PR 若引入重复立即失败，并在 Plan 271 文档记录首次成功 run ID。  
- [ ] 《00-README》更新归档流程 & 守卫说明，新增 Plan 271 条目。  
- [ ] 归档动作完成后 30 分钟内推送远端并更新 `ARCHIVE-RECORDS-*`。  
- [ ] 未来 7 天内无新的重复文件（通过 CI 历史 Run 佐证）。  

## 7. 跟踪与回滚

- 若 CI 因“Plan 271 Guard”失败，应先确认是否存在合法的双栈需求，再在本文档记录豁免理由，并通过 `scripts/quality/doc-archive-check.js` `exceptions` 列表临时放行（需附 `TODO-TEMPORARY` 截止日期）。
- 如需回滚新的守卫，必须在 Plan 271 文档列出替代方案与风险评估，并获得文档治理 Owner 批准。

---

**最近更新**  
- 2025-11-21：建立文档、补齐时间线、落地 CI 守卫（BY Codex）。
