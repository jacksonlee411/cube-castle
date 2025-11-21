# Plan 270 - Workflow 契约与守卫门禁修复

**文档编号**: 270  
**创建日期**: 2025-11-21  
**关联计划**: Plan 264（Workflow 治理）、Plan 265（Required Checks）、Plan 269（WSL Runner 部署）

---

## 1. 背景与目标

- 2025-11-21 的 `chore: refresh contract baseline (#26)` 合并后，Required workflows `contract-testing.yml`、`iig-guardian.yml`、`e2e-smoke.yml` 在 GitHub Actions 上 0s 失败，`gh run view` 显示 “workflow file issue”，说明 YAML 结构本身无效，CI 完全未执行。
- 本地使用 `actionlint` 复现出以下语义错误：
  - `contract-testing.yml` 的 `performance-impact-analysis` Job 同时定义了两个 `if`（`line302`、`line318`），违反 YAML 键唯一性。
  - `iig-guardian.yml` 与 `e2e-smoke.yml` 的 Job 级 `if` 表达式引用 `matrix.wsl_only`，但根据 GitHub context 可用性规则，`matrix` 仅能在 `strategy` 和 step 级别使用；Job 级引用会触发 “context not allowed”。
- 这些工作流承担契约快照/测试、Implementation Inventory Guardian、E2E Smoke 门禁，是 `AGENTS.md` 明确要求的必跑检查。Plan 270 旨在统一修复上述工作流、补上静态校验、防止再度破坏 Required checks。

---

## 2. 范围与交付物

| 类别 | 交付物 | 说明 |
|------|--------|------|
| Workflow 修复 | `.github/workflows/contract-testing.yml` | 合并 `performance-impact-analysis` Job 的条件，使其只保留一个 `if` 并尊重“PR 默认、workflow_dispatch 允许强制跑”的约束。 |
| Workflow 修复 | `.github/workflows/iig-guardian.yml`、`.github/workflows/e2e-smoke.yml` | 将 `matrix.wsl_only` 条件迁移到合法上下文（如 `strategy.matrix.include.expose` 标志 + step 级 `if`，或使用 `exclude` 过滤），确保 push/PR/workflow_dispatch 均可运行，并保留 docs-only bypass。 |
| 静态校验 | `Makefile` / `scripts/ci/workflows` / `docs/development-plans/264-workflow-governance.md` §“Workflow Governance Checklist” | 引入 `actionlint` 校验步骤（本地 `make workflow-lint` + CI hooks），并在 Plan 264 及 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 标注命令、输出路径。 |
| 运行证据 | Plan 265 §“Required Checks Runbook” + `logs/ci/<plan270>/*` | 至少一次成功的 `contract-testing`、`iig-guardian`、`e2e-smoke` run（包含自托管 WSL job 和 ubuntu job），Run ID/日志引用以 `Plan265-YYYYMMDD` 表格记录；actionlint 输出保存在 `reports/workflows/actionlint-<run>.txt` 并上传 artifact。 |

不在范围：业务代码改动、Docker Compose 调整、Runner 拓扑变更（交叉由 Plan 269 管理）。任何自托管 Runner 操作需继续遵循 `AGENTS.md` 的 Docker 强制与审批策略。

---

## 3. 实施步骤

### 3.1 契约测试工作流
1. 在 `performance-impact-analysis` Job 内移除重复的 Job 级 `if`，仅保留一个综合条件：默认在 PR/`workflow_dispatch` 触发；对 `matrix.wsl_only` 的分支需放在 step 级 `if` 或通过 matrix `include` 的 flag 控制，并补充注释说明 docs-only shortcut 不受影响。
2. 复查 Job 内步骤，确保 `matrix.prepare`、`matrix.wsl_only` 判定仅在 step 上使用，同时在 `Docs/CI-only fast pass` step 增加注释/测试说明。
3. 跑 `actionlint .github/workflows/contract-testing.yml` + `gh workflow run contract-testing.yml --ref feat/shared-dev`（或待命 PR）验证 job 能被创建，并把 run 链接登记到 Plan 265 的 “契约门禁” 表格。

### 3.2 IIG Guardian 工作流
1. 调整 `strategy.matrix`：为自托管组合增加 `wsl_only: true` 标记，并透过 step 级 `if: ${{ matrix.wsl_only != true || github.event_name == 'workflow_dispatch' }}` 控制执行；Job 级 `if` 保持只依赖 `github.*`。需要在 docs-only bypass step 保留 `TODO-TEMPORARY` 注释并记录回归方式。
2. 复用 `scripts/ci/workflows/prepare-selfhosted.sh`，确保 teardown 仍在 `if: matrix.prepare == true && always()` 下执行，并记录日志落盘目录（`logs/iig-guardian/<run>.log`）。
3. 重新触发 workflow（一次 PR、一次手动 dispatch），run 成功后把 Run ID/日志添加到 Plan 265 §IIG Guardian，以及 Plan 269（记录自托管 Runner 证据）。

### 3.3 E2E Smoke 工作流
1. 采用与 IIG 同样的矩阵控制方式，去除 Job 级 `if` 中的 `matrix` 引用；必要时在 matrix 中使用 `include` + `when` 字段（或 `wsl_only` flag）并在 step 里判断 `github.event_name`，同时断言 docs-only shortcut 仍然覆盖 `ci_only`、`docs_only` 两个过滤器。
2. 针对 push 事件（`ci/e2e-touch/**`）确认 ubuntu job 能跑，WSL job 则在 `workflow_dispatch` 或 `feat/shared-dev` PR 中验证，Run ID 记录到 Plan 265 §“E2E Smoke”。
3. 收集 `e2e-test-output.txt`、`backend-compose.log` 并上传 artifact 以证明成功执行，同时把 artifact 路径附在 Plan 265 中。

### 3.4 Actionlint 门禁
1. 在仓库根创建 `scripts/ci/workflows/run-actionlint.sh`（或 Makefile target）安装/调用 `actionlint`，并加入 `.github/workflows/plan-264-workflow-governance.yml` / `Agents Compliance` 的步骤，确保每次 push/PR 自动校验。
2. 更新 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 与 Plan 264（Workflow Governance Checklist），要求本地提交前运行 `make workflow-lint`，并在文档中列出示例输出片段与失败处理方式。
3. 将 `actionlint` 结果归档到 `reports/workflows/actionlint-<run>.txt`（或 `reports/workflows/run-<run_id>-actionlint.txt`）并上传 artifact，同时在 Plan 265 的 Required checks 表中填写 actionlint run 记录（列出命令、commit、Run ID、artifact 路径）。

---

## 4. 验收标准

- [ ] `actionlint` 对整个 `.github/workflows` 目录执行时无错误；已在 Makefile/CI 中固化。
- [ ] `contract-testing` workflow 能在 PR 与 `workflow_dispatch` 场景下完成所有 job，`contract-snapshot`、`contract-testing`、`contract-compliance-gate`、`performance-impact-analysis` 均执行成功（提供 Run ID）。
- [ ] `iig-guardian` workflow 在 push/PR 分支能创建 ubuntu job（docs-only 仍可短路），WSL job 仅在允许场景执行且成功完成，Run ID 更新至 Plan 265。
- [ ] `e2e-smoke` workflow push/PR/job 均正常，`e2e-test-output.txt` 中无 `❌`，artifact 正常上传。
- [ ] Plan 264/265/269 的相关章节同步更新（记录 Run ID、残余风险、actionlint 要求），确保唯一事实来源一致。

---

## 5. 风险与缓解

| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| 修复不彻底 | 仍可能有其他 workflow 出现类似 context 错误 | `actionlint` 覆盖全部 workflow；提交前必须跑 `make workflow-lint` |
| 自托管 Runner 不可用 | WSL Runner 仍在调试阶段，可能影响 job | ubuntu-latest 作为默认路径；自托管 job 失败时保持 `if: always()` 做清理并记录 Plan 269 |
| Docs-only 矩阵判断错误 | paths-filter 配置重复项可能导致误判 | 在修复中统一 `docs_only` 规则并添加测试（模拟 docs-only PR） |
| 门禁中断影响合并 | 修复期间 Required checks 全部红灯 | 在 `feat/shared-dev` 上先完成修复，再跑 workflow，确保 master 不再引入未验证变更 |

---

## 6. 里程碑

- **M1（2025-11-21）**：完成根因确认、Plan 270 建立、actionlint 本地验证。
- **M2（2025-11-22）**：提交并合并 workflow 修复 PR（含 actionlint 工具链），跑通 `contract-testing`、`iig-guardian`。
- **M3（2025-11-23）**：`e2e-smoke` 恢复绿灯，Plan 264/265/269 更新 Run ID，关闭计划。

---

## 7. 参考资料

- `AGENTS.md`（Docker 强制、Workflow 守卫原则）
- `.github/workflows/contract-testing.yml` / `iig-guardian.yml` / `e2e-smoke.yml`
- `docs/development-plans/264-workflow-governance.md`
- Plan 265/269（Required checks 与 WSL Runner 指南）
- `actionlint` 官方文档：<https://github.com/rhysd/actionlint>

## 8. 实施记录（2025-11-21）

- 已将 `contract-testing.yml`、`iig-guardian.yml`、`e2e-smoke.yml` 的 WSL matrix 约束下沉到 step 级 `runner-gate`，避免 job 级 `matrix.wsl_only` 表达式导致的 “context not allowed”/0s failure，并保留 docs-only fast pass。`performance-impact-analysis` job 现支持 `workflow_dispatch` 强制运行且自托管清理始终经 `prepare-selfhosted.sh`。  
- 新增 `scripts/ci/workflows/run-actionlint.sh`、`make workflow-lint` 目标与 `reports/workflows/.gitkeep`/`.gitignore`，Agents Compliance workflow 在 checkout 后运行 actionlint 并上传 `workflow-lint-<run_id>` artifact；本地验证记录：`reports/workflows/actionlint-20251121T103910Z.txt`（Plan 270）。  
- Plan 264/265/CI 指南已纳入 workflow lint 要求，Runbook 需在 Required checks 变更时记录 actionlint 报告路径；Plan 265 示例条目：`2025-11-21 10:39Z / make workflow-lint / reports/workflows/actionlint-20251121T103910Z.txt`。  
- `.github/actions/paths-filter` 增补 `docs_only`/`ci_only` 输出定义，`api-compliance.yml` 移除恒 false 的自托管准备 step，确保 actionlint 对现有 workflow 全量通过。

## 9. 待办与阻塞

- `contract-testing.yml` 的 `performance-impact-analysis` job 在 `workflow_dispatch` Run `19568443094` 中因前端 TypeScript 错误失败（`PositionDetailView.tsx` tabId、`StatusBadge.tsx` 颜色属性、`temporalMasterDetailApi.ts` 缺少 unified client 等），需先修复前端类型问题（Plan 263 范畴）方可加入 Required checks。
- `iig-guardian.yml`（Run `19568402680`）与 `e2e-smoke.yml`（Run `19568978952`）的 `self-hosted,cubecastle,wsl` 矩阵长期 `queued`，显示 WSL Runner 不可用；需按 Plan 269/265 恢复 WSL runner 后重新收集成功 run 作为验收证据。
- `e2e-smoke` Ubuntu 变体已通过 docs-only 快速通道并上传 `e2e-smoke-outputs`，但尚未执行真实 E2E（WSL 环境未恢复）；待 WSL runner 可用后需重新触发完整栈并记录 `e2e-test-output.txt`/`backend-compose.log` 等工件。
