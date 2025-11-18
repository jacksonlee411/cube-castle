# Plan 264 - GitHub Actions 工作流盘点与激活策略

**文档编号**: 264  
**标题**: GitHub Actions 工作流盘点与激活策略  
**版本**: v0.1  
**创建日期**: 2025-11-18  
**关联计划**: Plan 215（Phase2 日志）、Plan 255（本地 pre-push 守卫）、Plan 257（Facade Coverage）、Plan 258（Contract Drift Gate）

---

## 1. 背景与目标

- 仓库近期从私有调整为公共仓库，且共享分支 `feat/shared-dev` 需要全量 CI 门禁配合。多条 workflow 在 push 上出现 “0s failure / workflow file issue”，说明仍处于禁用或配置缺失状态。
- Required checks（gates-250/255、Contract Drift、Facade Coverage、Compose/Image、PR Body Policy、Plan254/257 等）需要稳定运行；同时还存在大量辅助 workflow（文档守卫、重复代码检测、E2E/自托管探针等），需要梳理其用途与启用策略，避免“僵尸”配置影响门禁统计。
- 本计划目标：建立 GitHub Actions 工作流唯一事实来源，列出当前仓库全部 workflow、用途、触发范围、最近 run 状态，并给出“是否要求启用”的建议与回滚方案。

## 2. 工作流列表（截至 2025-11-18 03:10 UTC）

| # | Workflow 名称（文件路径） | 主要作用 | 最近 run（gh run list） | 现状评估 | 建议 |
|---|----------------------------|----------|--------------------------|----------|------|
| 1 | Agents Compliance (`agents-compliance.yml`) | Plan 250/255 之外的代理合规检查 | 2025-11-18 02:36Z push **success** | Required 检查之一，正常 | 保持启用 |
| 2 | API Compliance (`api-compliance.yml`) | REST 契约与端口守卫 | 2025-11-18 02:58Z push **failure (0s)** | 需要启用/修复 | 启用并 rerun |
| 3 | Audit Consistency (`audit-consistency.yml`) | 文档/审计一致性守卫 | 2025-11-18 02:58Z push **failure (0s)** | 禁用/脚本缺失 | 启用或确认是否下线 |
| 4 | Auth Uniqueness Guard (`auth-uniqueness-guard.yml`) | 权限/认证唯一性脚本 | 最近无 run（未触发） | 需按需触发 | 若确实需要，手动 workflow_dispatch 验证 |
| 5 | CI Self-hosted Diagnose (`ci-selfhosted-diagnose.yml`) | 自托管 runner 诊断 | 2025-11-17 workflow_dispatch **success** | 非 Required，自助排障 | 保留但按需触发 |
| 6 | CI Self-hosted Smoke (`ci-selfhosted-smoke.yml`) | 自托管 runner 冒烟 | 2025-11-17 push **queued** | runner 未启用 | 若暂不使用，禁用以防 queued |
| 7 | CI (`ci.yml`) | 主 CI 工作流（旧版） | 2025-11-18 push **failure (0s)** | 可能已废弃但仍触发 | 明确是否迁移；必要则禁用 |
| 8 | Consistency Guard (`consistency-guard.yml`) | CQRS/硬编码守卫 | 2025-11-18 push **failure (0s)** | 需重新启用 | 启用并 rerun |
| 9 | 契约测试自动化验证 (`contract-testing.yml`) | Required：Plan 258 + “性能影响分析” | 2025-11-18 push **success** | 正常 | 持续监控，Plan 263 提升到 Required |
|10 | Docker Compliance (`docker-compliance.yml`) | Docker 镜像/Compose 检查 | 2025-11-18 PR **success** | 按需运行 | 保持启用 |
|11 | Docs Audit Quality (`docs-audit-quality.yml`) | 文档质量巡检 | 2025-11-18 push **failure (0s)** | 需要启用 | 启用并 rerun |
|12 | Document Sync (`document-sync.yml`) | 文档双写同步守卫 | 2025-11-18 push **failure (0s)** | 需要启用，Plan 261/262 约束 | 启用并观察误报 |
|13 | Duplicate Code Detection (`duplicate-code-detection.yml`) | 重复代码静态检查 | 2025-11-18 push **failure (0s)** | 需要启用 | 启用并 rerun |
|14 | E2E DevServer Probe (`e2e-devserver-probe.yml`) | 探针/健康检查 | 2025-11-17 push **success** | 正常 | 保持启用 |
|15 | E2E Job Probe (`e2e-probe.yml`) | 检查 E2E job 注册 | 2025-11-18 PR **success** | 正常 | 保持启用 |
|16 | E2E Smoke (`e2e-smoke.yml`) | 轻量 E2E | 2025-11-18 push **failure (0s)** | 需要启用 | 启用并 rerun |
|17 | E2E Tests (`e2e-tests.yml`) | 旧版 E2E 流程 | 最近无 run | 需确认是否仍使用 | 明确计划：保留或退役 |
|18 | Frontend E2E DevServer (`frontend-e2e-devserver.yml`) | 前端 DevServer E2E | 2025-11-18 push **failure (0s)** | 需启用 | 启用并 rerun |
|19 | Frontend E2E (`frontend-e2e.yml`) | Playwright 浏览器执行 | 2025-11-18 push **failure (0s)** | 需启用，Required 候选 | 启用并 rerun；Plan 255 观察 |
|20 | Frontend Quality Gate (`frontend-quality-gate.yml`) | 前端 lint/test | 2025-11-18 push **failure (0s)** | 需启用 | 启用并 rerun |
|21 | Go Backend Tests (`go-backend-tests.yml`) | Go 单测门禁 | API 404（文件缺失） | 处于“僵尸配置” | 若不再使用，删除/禁用 |
|22 | IIG Guardian (`iig-guardian.yml`) | Integration Inventory Guard | 2025-11-18 push **failure (0s)** | 需启用 | 启用并 rerun |
|23 | Integration Test (`integration-test.yml`) | 后端集成测试 | 2025-11-18 master push **success** | 仅 master 触发 | 若需在 feat/shared-dev 运行，扩展触发 |
|24 | Ops Scripts Quality (`ops-scripts-quality.yml`) | 脚本质量巡检 | 2025-11-18 push **failure (0s)** | 需启用 | 启用并 rerun |
|25 | Plan 240E Regression (`plan-240e-regression.yml`) | 老计划回归测试 | 2025-11-18 master push **failure** | 需评估是否仍要求 | 若计划已结项，可停用；否则修复失败原因 |
|26 | Plan 250 Gates (`plan-250-gates.yml`) | Required - 不允许 legacy env | 2025-11-18 push **success** | 正常 | 保持启用 |
|27 | Plan 253 Gates (`plan-253-gates.yml`) | Required - Compose/Image | 2025-11-18 push **success** | 正常 | 保持启用 |
|28 | Plan 253 Publish (`plan-253-publish.yml`) | 镜像发布 | 最近无 run | 需按需触发 | 保留 |
|29 | Plan 254 Gates (`plan-254-gates.yml`) | Contract Drift Gate | 2025-11-18 push **failure (0s)** | 需启用 | 启用并 rerun |
|30 | Plan 255 Gates (`plan-255-gates.yml`) | 本地/CI 质量门禁 | 2025-11-18 push **success** | 正常 | 保持启用 |
|31 | Plan 257 Gates (`plan-257-gates.yml`) | Facade Coverage Gate | 2025-11-18 push **success** | 正常 | 保持启用 |
|32 | Plan 258 Gates (`plan-258-gates.yml`) | Contract Drift Guard | 2025-11-18 push **success** | 正常 | 保持启用 |
|33 | Plan 259A Switch (`plan-259a-switch.yml`) | 硬门禁切换 | 2025-11-18 master push **success** | 正常 | 保留 |
|34 | PR Body Policy (`pr-body-policy.yml`) | Required PR 模板守卫 | 2025-11-18 PR **success** | 正常 | 保持启用 |
|35 | Test (`test.yml`) | 定时 Extended Tests | 2025-11-18 schedule **failure** | 需确认是否仍需 | 若无价值，可停用或修复 |
|36 | Extended Tests (`test.yml` / “Extended Tests”) | 也显示在 workflow 列表 | 同上 | 同上 | 同上 |

> 数据来源：`gh workflow list` 与 `gh run list --workflow <file> --limit 1`（UTC 2025-11-18 03:10）。标记为 “failure (0s)” 的 workflow，全部出现“workflow file issue”提示，应在 Actions UI 中点击 “Enable workflow” 或修复 branch/path 条件。

## 3. 启用/退役决策与步骤

1. **立即启用的关键工作流**（影响 PR Checks 与质量门禁）  
   - `frontend-e2e.yml`、`frontend-quality-gate.yml`、`consistency-guard.yml`、`docs-audit-quality.yml`、`duplicate-code-detection.yml`、`document-sync.yml`、`api-compliance.yml`、`audit-consistency.yml`、`plan-254-gates.yml`。  
   - 操作：在 GitHub Actions -> Workflow 详情页 -> Enable workflow；启用后于 PR #19 或最新 PR 点击 “Re-run all jobs”，确保 Required check 引用的是最新 run（非旧 run 19448607962）。

2. **需要评估是否退役/改造的工作流**  
   - `ci.yml`（旧主 CI）、`go-backend-tests.yml`（缺文件）、`plan-240e-regression.yml`、`test.yml`（定时 Extended Tests）、`e2e-tests.yml`（若已被 Plan 255/Frontend E2E 取代）。  
   - 需与 Plan 215/Plan 255 负责人确认是否还有使用场景；若没有，更新 `.github/workflows/`、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 并在 PR 中说明，确保“资源唯一性”不再引用。

3. **自托管相关**  
   - `ci-selfhosted-smoke.yml` 当前 run 状态为 queued（runner 不可用），若短期内不使用自托管，可禁用 workflow；若需要，则恢复 runner 并记录操作手册（Plan 262）。

4. **Plan 263 依赖**  
   - “契约测试自动化验证” workflow 中的 “性能影响分析” job 将在 Plan 263 完成 TS 修复后设为 Required。届时需在 Branch Protection -> Required status checks 中新增该 job 名，并在本计划里记录切换时间与回滚路径。

## 4. 验收标准

- [ ] 所有 Required checks 对应的 workflow 均处于启用状态，并能在 `feat/shared-dev` push 上生成成功 run。
- [ ] workflow 盘点文档（本文件）列出的状态在 CI 审核会议上复核，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 或相关文档引用。
- [ ] 对于决定退役的 workflow，已在仓库中删除/禁用并记录回滚方式；GitHub Actions 中的旧 run 不再误导 PR Checks。
- [ ] 新增 Required 项（性能影响分析）在 Plan 263 验收时更新此文档并同步 Branch Protection。

## 5. 回滚策略

- 若某 workflow 启用后导致 CI 长时间排队或误报警，可在 Actions UI 选择 “Disable workflow” 并在 PR 中记录原因；同时在本计划文件中追加更新记录。
- Required check 调整需保留至少 1 次成功 run 作为基准；回滚时需更新 Branch Protection，并在 `CHANGELOG.md` 与 Plan 263/264 文档中注明恢复时间。

## 6. 更新记录

- 2025-11-18：首次创建，列出全部 36 条 workflow、状态与建议。 (BY: Codex)
- 2025-11-18：根据“无效/重复全部退役”要求，已分两批从仓库移除以下 workflow，清理 0s failure 噪音与僵尸配置：
  - 批次 1：`ci.yml`（旧主 CI）、`go-backend-tests.yml`（go-app 目录缺失）、`plan-240e-regression.yml`（旧回归）、`test.yml`（定时 extended tests）、`e2e-tests.yml`（旧版前端 E2E）。
  - 批次 2：`ci-selfhosted-diagnose.yml`、`ci-selfhosted-smoke.yml`（自托管 runner 暂停）、`e2e-devserver-probe.yml`、`e2e-probe.yml`（探针类重复）、`frontend-e2e-devserver.yml`（与主 E2E 重叠）、`ops-scripts-quality.yml`、`audit-consistency.yml`、`docs-audit-quality.yml`、`duplicate-code-detection.yml`、`plan-253-publish.yml`。
  如需恢复，需从历史提交重新拷贝并重新启用；若有替代方案，请在对应计划文档中登记。 (BY: Codex)
