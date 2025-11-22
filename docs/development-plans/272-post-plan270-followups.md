# Plan 272 - Plan 270 遗留事项清零计划

**文档编号**: 272  
**创建日期**: 2025-11-22  
**关联计划**: Plan 264（Workflow 治理）、Plan 265（自托管门禁）、Plan 269（WSL Runner 部署）、Plan 270（Workflow 守卫回收）

---

## 1. 背景与目标

- Plan 270 完成后，Required workflows 已全部迁回 GitHub 平台 Runner，但仍存在若干“临时豁免/挂起事项”：
  1. `golangci-lint` 为规避 go1.24 typecheck 噪音，临时关闭 `typecheck` 并强制 `GOTOOLCHAIN=go1.24.9`。若不回收，Plan 255 只能覆盖 depguard/tagliatelle，无法恢复完整语义校验。
  2. E2E smoke 的健康探针因 Query Service 未返回 `database` 字段而暂时“缺字段视为通过”，仍有 TODO。若后端恢复该字段，脚本需回收放宽逻辑。
  3. 自托管 Runner（Plan 265/269）目前完全停用，仅保留 `ci-selfhosted-smoke`，document-sync/API compliance/consistency guard 等 job 尚无 WSL 绿灯 run。
  4. Plan 265 Runbook 需持续登记 `make workflow-lint`/Required workflows 的 workflow_dispatch 记录，确保替换/回退时有唯一事实来源。
- Plan 272 旨在按模块和工作量拆解以上遗留事项，形成独立整改计划，确保 Plan 270 交付闭环。

---

## 2. 范围与拆解（按模块/工作量）

| 模块 | 工作量预估 | 任务 | 说明 |
|------|-----------|------|------|
| **A. 质量守卫（Go lint）** | 2 人日 | 恢复 `golangci-lint` typecheck | 逐项修复 `internal/types/date.go`、`authbff`、`publicgraphql` 等文件在 go1.24 loader 下的类型问题，完成后移除 `.golangci.yml` / `golangci-fast.yml` 的 `typecheck` disable，保留 `go: "1.24"` 和 `GOTOOLCHAIN=go1.24.9` 作为长期基线。 |
| | | 清理 skip-files | 验证 `cmd/hrms-server/query/internal/graphql/generated.go` 等 skip 列表是否仍需豁免，缩减为最小集（仅保留第三方/生成物）。 |
| **B. E2E & Smoke 健康检查** | 1 人日 | 回收 `scripts/simplified-e2e-test.sh` 中“缺 `database` 字段视为通过”的 TODO | 与 Query Service 团队确认 health payload 恢复计划；新增断言逻辑，若 `database` 缺失即失败，并在 Plan 270 Runbook 更新回收记录。 |
| | | E2E selector/日志复核 | 再次运行 Plan 219E E2E 套件，确保 Plan 270 期间新增 selector、日志输出均记录在案，无额外 TODO。 |
| **C. 自托管 Runner 恢复路径** | 3 人日（与 Plan 269 联动） | 复位 document-sync/api-compliance/consistency guard 的 WSL matrix | 参照 Plan 265 原表格，将 `self-hosted,cubecastle,wsl` 组合恢复到 YAML 中，以 `workflow_dispatch` 验证至少 1 次绿灯（record run ID）。 |
| | | Runner 网络与 watchdog 验证 | 复用 Plan 267 网络修复记录，确保 `docker-compose.runner.persist.yml`、`watchdog.sh` 正常；`ci-selfhosted-smoke` 需在恢复前连跑 3 次成功。 |
| **D. Runbook 与证据管理** | 0.5 人日 | Plan 265 Runbook 更新 | 将 `make workflow-lint` 的最新执行记录（命令、commit、报告路径）补入表格；后续 Required workflow dispatch 记录保持 1:1 对应。 |
| | | 文档对齐 | 更新 `docs/development-plans/06-integrated-teams-progress-log.md` 中“Required Checks”段落，说明 Plan 272 正在回收临时豁免。 |

---

## 3. 实施步骤

### 3.1 A. 质量守卫
1. 在本地执行 `~/go/bin/golangci-lint run`（开启 `typecheck`），收集当前报错列表。
2. 按文件修复：
   - `internal/types/date.go`: 改为 `d.Time.Format` 并补单测。
   - `authbff`/`jwtmint`: 明确引用的 chi/jwt/redis 版本，确保 go1.24 toolchain 可解析，必要时加入 `go.mod` replace。
   - `cmd/hrms-server/query/publicgraphql`：确认是否需要 `redis` build tag 或 mock 实现。
3. 全部修复后，更新 `.golangci.yml`/`scripts/quality/golangci-fast.yml`，移除 `linters.disable.typecheck`，保留 `go: "1.24"` 与 `GOTOOLCHAIN`。
4. 在 Plan 265 Runbook 记录“typecheck 恢复日期 + 命令 + 报告”。

### 3.2 B. E2E 健康检查
1. 与 Query Service 团队确认健康端点字段回补计划，记录在 Plan 270 Runbook 中。
2. 修改 `scripts/simplified-e2e-test.sh` 与相关 Playwright 脚本，要求 `database.status === "healthy"`，否则失败；删除 `TODO-TEMPORARY`。
3. 运行 `gh workflow run e2e-smoke.yml --ref feat/shared-dev`，收集新的 `e2e-smoke-outputs` artifact，并更新 Plan 265 Runbook。

### 3.3 C. 自托管 Runner 恢复
1. 确认 WSL Runner 状态：`docker compose -f docker-compose.runner.persist.yml ps`；如需重启，遵循 Plan 269 指南。
2. 逐个恢复 workflow YAML 中的 self-hosted matrix：document-sync / api-compliance / consistency-guard / plan-254 / plan-255 / plan-258 / plan-253 / contract-testing (performance job) 等，并保留 workflow_dispatch 触发路径。
3. 对每个 workflow 至少运行一次 `gh workflow run ... --ref feat/shared-dev`，等待 WSL job 绿灯，记录 run ID + artifact 链接到 Plan 265 Runbook。
4. 完成后再更新 Branch Protection，使 self-hosted job 再次纳入 Required checks（与 DevOps 协调）。

### 3.4 D. Runbook/文档
1. 补充 Plan 265 Runbook 表格（workflow-lint、各 Required workflow）与 Plan 270 Runbook 的“遗留项追踪”段落。
2. 在 `docs/development-plans/06-integrated-teams-progress-log.md` 中新增条目，标注 Plan 272 的进展与验收节点。

---

## 4. 验收标准

- [ ] `.golangci.yml`/`scripts/quality/golangci-fast.yml` 恢复 `typecheck`，`~/go/bin/golangci-lint run` 无噪音；Plan 255 pre-push 钩子不再输出类型错误。
- [ ] `scripts/simplified-e2e-test.sh`/E2E spec 中与数据库健康相关的 TODO 删除，e2e-smoke workflow 的健康探针重新验证健康字段；Plan 270 Runbook 记录回收时间。
- [ ] document-sync / api-compliance / consistency-guard / plan-254 / plan-255 / plan-258 / plan-253 / contract-testing (performance) 等自托管 job 在 WSL Runner 上至少成功一次，Run ID 与 artifact 链接登记于 Plan 265；Branch Protection 更新 Required checks。
- [ ] Plan 265 Runbook、Plan 270 Runbook 与 `docs/development-plans/06-integrated-teams-progress-log.md` 均添加 Plan 272 的执行记录；所有新的 `make workflow-lint` 输出落盘到 `reports/workflows/` 并记录。

---

## 5. 里程碑

| 里程碑 | 目标 | 截止时间 |
|--------|------|----------|
| M1 | 完成 golangci-lint typecheck 修复并恢复配置 | 2025-11-24 |
| M2 | 回收 E2E 健康探针 TODO + 新的 e2e-smoke 证据 | 2025-11-25 |
| M3 | 自托管 Runner 至少 4 个主要 workflow 在 WSL 落地一次绿灯 | 2025-11-27 |
| M4 | Plan 265/270 Runbook 及 Plan 06 进度日志更新，Plan 272 归档 | 2025-11-28 |

---

## 6. 风险与缓解

| 风险 | 描述 | 缓解 |
|------|------|------|
| typecheck 修复跨度大 | 部分问题出在第三方生成代码（GraphQL） | 先在 `.golangci.yml` 中对确实不可控的生成物保持 skip，逐步减小范围 |
| WSL Runner 不稳定 | Plan 267/269 尚未完全闭环 | 与 DevOps 协调在恢复自托管时先运行 `ci-selfhosted-smoke` 与 `plan-253` diag，必要时逐项恢复 |
| 健康端点缺字段恢复时间未知 | Query Service 需求排期可能延后 | 在 Plan 272 Runbook 中记录 Blocker，并保留 fallback 逻辑开关；若超过 1 周仍未恢复，回报给 Plan 06/215 |
| Runbook 更新遗漏 | 多项 workflow 同时恢复，记录容易遗漏 | 建立模板脚本 `scripts/ci/workflows/log-run.sh`，统一写入 Plan 265 表格所需字段 |

---

## 7. 资料与跟进

- Plan 265 Runbook：`docs/development-plans/265-selfhosted-required-checks.md`（Run ID 表格）
- Plan 270 Runbook：`docs/development-plans/270-workflow-contract-guardian-remediation.md`
- Plan 269 WSL Runner 指南：`docs/archive/development-plans/269-wsl-runner-deployment.md`
- 进度同步：Plan 06 日志 + `docs/development-plans/06-integrated-teams-progress-log.md`  
- 所有命令/证据需落在 `logs/plan272/**`、`reports/workflows/**`
