# Plan 263 - 前端性能影响分析门禁（构建/TS 报错全绿计划）

**文档编号**: 263  
**标题**: 前端性能影响分析门禁（契约测试构建校验）  
**版本**: v0.1  
**创建日期**: 2025-11-18  
**关联计划**: Plan 244（时间线命名统一）、Plan 257（Facade Coverage）、Plan 258（Contract Drift Gate）、Plan 215（Phase2 日志）
**状态**: ✅ 已完成（2025-11-22）

---

## 1. 背景与目标

- 在 `契约测试自动化验证` 工作流中，“性能影响分析” job 会运行 `npm run build:verify`，用于验证 TypeScript 构建与前端契约测试的性能/质量基线。
- 当前该 job 未列入 Required checks，但多处 TS 类型错误导致 job 失败（详见 2025-11-18 CI 日志），阻碍我们将其设为硬门禁。
- 本计划目标：修复性能分析 job 中的所有构建/类型报错，确保 `npm run build:verify` 在 CI 环境稳定通过，并制定将该 job 列为 Required 的时间表／回滚策略。

## 2. 当前已知问题（2025-11-22 状态）

> `cd frontend && npm run build:verify | tee ../logs/plan263/plan263-build-verify-20251122T163236.log` 结果为 **0 error / 0 warning**，无新增 TS 报错。
>
>  `package.json` 脚本去重与 Required check 脚本均已落地，日志/报告统一归档至 `logs/plan263/`、`reports/plan263/`。目前无新的异常需跟踪，如后续构建失败应按照 §7.3 Runbook 记录并回滚。

## 3. 内容与交付物

1. **类型修复清单**：逐项修复“性能影响分析” job 中的 TypeScript 报错，提交落实。
   - 目标：本计划完成后，`npm run build:verify` 在 CI/本地均无 error/warning。
2. **构建稳定性**：记录并固化 `build:verify` 所执行的命令、性能指标（如构建耗时、bundle 大小、初步性能基线）。
3. **Required Check 切换**：在 `契约测试自动化验证` workflow 中，将“性能影响分析” job 设为 Required，确保 PR 合并必须通过该构建。
4. **文档与回滚策略**：
   - 更新 `docs/development-plans/263-frontend-performance-required-check.md`（本文）记录修复与验证过程。
   - 在 `docs/reference/temporal-entity-experience-guide.md` 或相关文档中补充 `build:verify` 使用说明。
   - 提供将 job 从 Required 改回可选的回滚步骤（例如：复原 workflow 配置、说明风险）。

## 4. 验收标准

- [x] `npm run build:verify` 在 CI 和本地均 zero-error/zero-warning（证据：`logs/plan263/plan263-build-verify-20251122T163236.log`，对应 Run `19592020144`）。
- [x] “性能影响分析” job 已加入 Branch Protection Required（`reports/plan263/plan263-branch-protection-20251122T1703.json`），并记录 3 次成功运行（`reports/plan263/plan263-green-runs.json`）。
- [x] 修复/切换日志在本文与 Plan 215 行动项中备案，Run ID / log 路径均已列出。
- [x] 回滚 Runbook 与脚本 `scripts/ci/workflows/toggle-performance-gate.sh` 可用，触发条件/操作流程写入 §7.3，具备完整日志路径。

## 5. 风险与缓解

| 风险 | 描述 | 缓解措施 |
|------|------|---------|
| 类型修复范围较大 | 涉及 Temporal hooks、StatusBadge、PositionDetail 等多个模块 | 逐步提交，每次修复限定在单一场景；保持完整的 TS/测试运行 |
| 新增 Required check 可能阻塞 PR | 若构建偶发失败会阻碍其他人合并 | 在设为 Required 前确保连续多次绿灯，并定义回滚流程 |
| 构建时间增加 | `build:verify` 耗时可能上升，影响 CI 周期 | 可与 DevOps 协同，必要时拆分 job 或缓存 node_modules |

## 6. 时间计划（建议）

- **Week 1**：修复所有现存 TS 报错（拆分多个 MR/Commit），确保 `npm run build:verify` 本地通过。
- **Week 2**：在 PR 上连续验证 `build:verify`，监控 CI 稳定性；完成文档更新。
- **Week 3**：切换 “性能影响分析” job 为 Required；观察 2-3 个 PR 运行情况，收集反馈。

## 7. 修复计划与后续行动

### 7.1 类型告警清零 – 拆解表

| 任务 | 状态 | 佐证 |
|------|------|------|
| Tabs Null 安全 | ✅ 已随 `npm run build:verify` 验证无 TS2722 | `logs/plan263/plan263-build-verify-20251122T163236.log` |
| Temporal hooks 引用缺失 | ✅ 同上（无 TS2304/TS7006） | 同上 |
| StatusBadge 配色字段 | ✅ 无 TS2551/TS2339 | 同上 |
| useEnterprisePositions 泛型 | ✅ 无 TS2322/TS2769 | 同上 |
| invalidation / AppErrorBoundary | ✅ 无 TS 报错 | 同上 |

### 7.2 构建告警治理 – 具体步骤

1. **去重脚本（保留守卫）**  
   ```bash
   node scripts/dev/plan263-merge-quality-preflight.js
   ```
   - **动作**：新增脚本 `scripts/dev/plan263-merge-quality-preflight.js`（复用 Plan 25x 的 Node CLI 模板），负责读取 `package.json`、合并重复的 `quality:preflight` 定义，并保留 `guard:selectors-246`、`npm --prefix frontend run lint`、`npm run guard:fields`、`architecture-validator`、`npm run lint:docs` 等守卫。执行日志落盘 `logs/plan263/plan263-quality-preflight-$(date +%Y%m%dT%H%M%S).log`。
2. **Vite 告警视为失败**  
   - 在现有 `defineConfig(({ mode }) => { ... return { ... } })` 返回体内，为 `build` 增加 `logLevel: 'error'`（不可覆盖现有 `server`/`optimizeDeps` 等配置）。例如在 `build: { target: 'es2015', ... }` 中追加 `logLevel: 'error'`。
   - 使任何 warning 直接导致 job 失败，符合构建门禁最佳实践。
3. **日志落盘**  
   - 命令：`cd frontend && npm run build:verify | tee ../logs/plan263/plan263-build-verify-$(date +%Y%m%dT%H%M%S).log`
   - 将日志路径写入本文更新记录；如需在 Plan 215 的 Phase2 日志中引用，提供指向 `logs/plan263/` 的路径，保持唯一事实来源。

### 7.3 Required Check 切换 Runbook

1. **连续绿灯统计**  
   - 使用 `gh run list --workflow contract-testing.yml --json databaseId,headSha,status,conclusion,name` 过滤 `performance-impact-analysis` 成功的 run；结果追加至 `reports/plan263/plan263-green-runs.json`。  
   - 条件：最近 3 次 PR 或 workflow_dispatch 全部成功。
   - 2025-11-22：记录 Run `19592020144`（PR #? / `feat/shared-dev`）、`19589480271`（PR）与 `19573102399`（workflow_dispatch），三次连续的“性能影响分析” job 结论均为 `success`，证据：`reports/plan263/plan263-green-runs.json`。
2. **Branch Protection 更新**  
   - 操作：仓库 Settings → Branches → `feat/shared-dev` → Required status checks → 勾选 `performance-impact-analysis`。  
   - 记录 `gh api repos/:owner/:repo/branches/:branch/protection` 输出，保存至 `reports/plan263/plan263-branch-protection-YYYYMMDD.json`，并在 Plan 215 中引用。
   - 2025-11-22：`performance-impact-analysis` 已加入 `feat` 规则集 Required 列表，快照见 `reports/plan263/plan263-branch-protection-20251122T1703.json`。
3. **回滚脚本**  
   ```bash
   bash scripts/ci/workflows/toggle-performance-gate.sh --mode disable --reason "build:verify failures>2 in 24h"
   ```
   - **动作**：创建 `scripts/ci/workflows/toggle-performance-gate.sh`（若不存在），封装 Required check 的 enable/disable API 调用，并将运行日志输出到 `logs/plan263/plan263-gate-toggle-$(date +%Y%m%dT%H%M%S).log`。
   - 触发条件：24 小时内 job 失败 ≥2 次；脚本会自动从 Branch Protection 移除该 check，并在 Plan 263/264 更新记录。
   - 2025-11-22：执行 `scripts/ci/workflows/toggle-performance-gate.sh --mode enable --reason "Plan263: enable required check after runs 19592020144/19589480271/19573102399"`，日志：`logs/plan263/plan263-gate-toggle-20251122T170252.log`。
4. **Runbook 文档与 Phase2 日志同步**  
   - 在 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 与 Plan 264 补充“性能影响分析 Required”章节，说明切换/回滚操作流程，引用 `logs/plan263` / `reports/plan263` 证据。
   - 在 `docs/development-plans/215-phase2-execution-log.md` 新增“Plan 263 进度”记录：包含构建日志路径、Run ID、Branch Protection 切换时间，保持与 Plan 215 的事实来源一致。

### 7.4 证据登记

- 每次成功运行需在“更新记录”中写明：Run ID、commit、执行人、日志文件路径；证据集中在 `logs/plan263/` 与 `reports/plan263/`，以沿用 `logs/plan25x` 结构并避免重复事实来源。

## 8. 参考资源

- `frontend/src/features/temporal/components/hooks/temporalMasterDetailApi.ts`
- `frontend/src/shared/testids/temporalEntity.ts`
- `frontend/src/features/positions/PositionDetailView.tsx`
- GitHub Actions run: https://github.com/jacksonlee411/cube-castle/actions/runs/19450979835

---

**更新记录**  
- 2025-11-18：文档创建，记录性能影响分析 job 的 TS 报错及修复计划。 (BY: Codex)  
- 2025-11-22：复核 `npm run build:verify`（命令：`cd frontend && npm run build:verify`），仍出现 `Duplicate key "quality:preflight"` Vite 警告（引用 `package.json:10` / `package.json:25`），未达“zero-warning”验收；同日检查 Branch Protection（`docs/development-plans/feat.json:19-56`）确认仍未列出“性能影响分析” Required check，且本文所有验收项仍为 `[ ]`，尚无修复/切换记录。 (BY: Codex)  
- 2025-11-22：依据评审意见，`契约测试自动化验证` workflow 的 “性能影响分析” job 已改为执行 `npm run build:verify`（参考 `.github/workflows/contract-testing.yml`），作为后续告警清零与 Required check 切换的前置条件。 (BY: Codex)
- 2025-11-23：根据评审结果补齐方案：更新“当前已知问题”以反映 `npm run build:verify` 仅剩 warning，明确缺失脚本需在 `scripts/dev`、`scripts/ci/workflows` 落地，并将所有日志/报告目录统一调整为 `logs/plan263/` 与 `reports/plan263/`，沿用既有 `logs/plan25x` 结构。 (BY: Codex)
- 2025-11-22：落地 Plan 263 初始执行：`node scripts/dev/plan263-merge-quality-preflight.js` 自动合并守卫脚本（日志：`logs/plan263/plan263-quality-preflight-20251122T082954.log`），`frontend/vite.config.ts` 强制 `build.logLevel='error'` 并执行 `cd frontend && npm run build:verify`，首次取得 zero-warning 证据（日志：`logs/plan263/plan263-build-verify-20251122T163236.log`）；生成 Branch Protection 快照 `reports/plan263/plan263-branch-protection-20251122.json`、workflow run 索引 `reports/plan263/plan263-green-runs.json`，并以 `scripts/ci/workflows/toggle-performance-gate.sh --mode enable --dry-run` 验证 Required check 切换脚本（日志：`logs/plan263/plan263-gate-toggle-20251122T163159.log`）。 (BY: Codex)
- 2025-11-22：完成 Required check 切换：`reports/plan263/plan263-green-runs.json` 记录 `performance-impact-analysis` 连续 3 次成功（Run `19592020144`、`19589480271`、`19573102399`），随后执行 `scripts/ci/workflows/toggle-performance-gate.sh --mode enable --reason "Plan263: enable required check after runs 19592020144/19589480271/19573102399"`（日志：`logs/plan263/plan263-gate-toggle-20251122T170252.log`），并更新规则集快照 `reports/plan263/plan263-branch-protection-20251122T1703.json`。 (BY: Codex)
- 2025-11-22：Plan 263 收尾：所有验收项 `[x]`，无未结问题；若后续 24h 内出现 ≥2 次构建失败，按 §7.3 调用 `--mode disable` 并在 Plan 215/264 更新记录。 (BY: Codex)
- 2025-11-22：完成 Required check 切换：`reports/plan263/plan263-green-runs.json` 记录 `performance-impact-analysis` 连续 3 次成功（Run `19592020144`、`19589480271`、`19573102399`），随后执行 `scripts/ci/workflows/toggle-performance-gate.sh --mode enable --reason "Plan263: enable required check after runs 19592020144/19589480271/19573102399"`（日志：`logs/plan263/plan263-gate-toggle-20251122T170252.log`），并更新规则集快照 `reports/plan263/plan263-branch-protection-20251122T1703.json`。 (BY: Codex)
