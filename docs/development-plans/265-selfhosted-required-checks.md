# Plan 265 - 自托管 Runner 门禁扩展（Plan 263/264 衔接）

**文档编号**: 265  
**标题**: 自托管 Runner 门禁扩展（Plan 263/264 衔接）  
**版本**: v0.1  
**创建日期**: 2025-11-19  
**关联计划**: Plan 262（自托管 Runner 持续化），Plan 263（性能影响分析 Required），Plan 264（Workflow 治理）  
**状态**: ⚠️ 搁置（2025-11-20）——受限于 WSL Runner 网络/调度，暂无法满足自托管门禁运行要求；待 Plan 267/269 网络与 Runner 稳定后重新评估。

---


## 📌 搁置结论

- 因 WSL Runner 网络/调度未恢复，Plan 265 所要求的自托管门禁跑通无法完成。
- 2025-11-20 起暂停执行自托管扩展，唯一保留的作业为 `ci-selfhosted-smoke`（runner 健康检查）。
- 需待 Plan 267/269 网络治理完成并重新触发 document-sync/api-compliance/consistency-guard 等 job 后，再恢复本计划。

## 1. 背景与目标

- Plan 262 已通过 `docker-compose.runner.persist.yml`、`start-ghcr-runner-persistent.sh`、`watchdog.sh` 将自托管 Runner 持续在线，冒烟/诊断 run（19485705844 / 19486773039）证明 runner 能稳定运行 Docker Compose。
- Plan 263/264 仍存在“未启用或未达 Required”的门禁规则：  
  1) `契约测试自动化验证` workflow 中的 “性能影响分析” job（`npm run build:verify`）因 TS 报错未能入列 Required；  
  2) `frontend-quality-gate.yml`、`frontend-e2e.yml`、`document-sync.yml`、`consistency-guard.yml`、`plan-254-gates.yml`、`api-compliance.yml`、`iig-guardian.yml` 等在 push 上记录 “failure (0s)” 或缺少稳定运行。
- 目标：在保持 AGENTS.md“资源唯一性 + Docker 强制”约束下，利用自托管 Runner 的 Compose 能力完成上述门禁启用/迁移，形成统一执行方案、回滚步骤与验收基线。

## 2. 范围与待落地规则

| 计划来源 | 工作流 / Job | 现状痛点 | 自托管 Runner 行动 |
|----------|--------------|----------|--------------------|
| Plan 263 | `contract-testing.yml` → `performance-impact-analysis` | TS 报错阻塞 Required，运行环境依赖 Docker Compose | 修复 TS 清单后，将 job runs-on 切换为 `{ self-hosted, cubecastle, wsl }`（可保留 ubuntu 矩阵作为回退），并在 job 前执行 `scripts/ci/workflows/prepare-selfhosted.sh contract-testing` 统一准备 Compose 环境 |
| ~~Plan 264~~ | ~~`frontend-quality-gate.yml`~~ | ~~历史版本 pipeline，现阶段停维~~ | ~~从 Plan 265 范围移除；如需恢复，需先在 `.github/workflows/` 重建文件并另行审议~~ |
| ~~Plan 264~~ | ~~`frontend-e2e.yml`~~ | ~~历史版本 pipeline，现阶段停维~~ | ~~从 Plan 265 范围移除；如需恢复，需先在 `.github/workflows/` 重建文件并另行审议~~ |
| Plan 264 | `document-sync.yml`、`consistency-guard.yml`、`plan-254-gates.yml` | 初次启用即失败，需要 Docker 服务/Go 工具链 | job 中统一调用 `scripts/ci/workflows/prepare-selfhosted.sh`（新增）处理 Compose 启动、`go env` 检查、`make db-migrate-all`，确保环境一致 |
| Plan 264 | `api-compliance.yml`、`iig-guardian.yml` | 0s failure（需 Enable），依赖本地生成的契约/清单 | 在自托管 runner 上运行，确保 `.cache/`、`docs/reference/*` 读取速度稳定；启用后在 Branch Protection 标记 |

（如需扩展，可在后续迭代把 `e2e-smoke.yml`、`plan-253-gates.yml` 等重型 job 也纳入自托管矩阵。）

## 3. 实施步骤

1. **前置准备**  
   - 复用 Plan 262 Runner：保证 `docker compose -f docker-compose.runner.persist.yml up -d` 已启动并在线；watchdog 正常记录日志。  
   - 新增 `scripts/ci/workflows/prepare-selfhosted.sh` 工具脚本，约束：  
     - 入口：`bash scripts/ci/workflows/prepare-selfhosted.sh <workflow-id> [--teardown]`；  
     - 内容：检测 Docker Engine ≥24、Compose Plugin ≥2.27、Go >=1.24、Node >=18；  
     - 执行 `docker compose -f docker-compose.dev.yml up -d postgres redis`，之后使用 `scripts/ci/docker/check-health.sh postgres 120` + `docker inspect -f '{{.State.Health.Status}}'` 轮询健康；  
     - 若设定 `CI_PREPARE_RUN_MIGRATIONS=1`，会在服务健康后执行 `make db-migrate-all`；  
     - 生成 `logs/ci-monitor/<workflow-id>-prepare.log`，供 Actions artifact 上传；  
      - `--teardown` 模式负责 `docker compose -f docker-compose.dev.yml down --remove-orphans` 与受控 `docker volume prune --filter label=cubecastle-ci --force`；  
      - 仅负责环境预热/清理，**不得**调用 `start-ghcr-runner-persistent.sh` 或 `config.sh` 以防 Runner 错误重配。
   - **WSL Runner 正式方案（Plan 269）**：自 2025-11-20 起，自托管 Runner 仅保留 WSL 形态，安装/校验/卸载统一由 `scripts/ci/runner/wsl-install.sh`、`wsl-verify.sh`、`wsl-uninstall.sh` 负责，并在 Plan 265/266/269 中登记 Run ID、日志与回滚时间；所有 workflow 的 self-hosted 矩阵仅使用 `[self-hosted, cubecastle, wsl]` 标签。

2. **Plan 263 任务**  
   - 按 Plan 263 TS 清单逐个修复（PositionDetailView、Temporal hooks、StatusBadge 等）；本地 + 自托管 runner 内执行 `npm run build:verify`，确认 0 error。  
   - 更新 `contract-testing.yml`：  
     - `performance-impact-analysis` job 增加 `runs-on` 矩阵（self-hosted + ubuntu），并在 steps 前调用 `prepare-selfhosted.sh`（仅 self-hosted 分支）。  
     - 缓存策略：`actions/cache` 针对 `~/.npm`, `frontend/node_modules`; 自托管分支使用磁盘持久化（避免重复下载 Playwright 依赖）。  
   - 连续 3 个 PR run 成功后，将该 job 名称加入 Branch Protection Required 列表，并在 Plan 263/265 文档记录 run ID + 切换时间 + 回滚步骤。

3. **Plan 264 任务**  
   - 在 Actions UI 启用下表列出的 workflow 并锁定 job 粒度（前端质量/E2E 工作流暂不维护，已从范围移除；若未来恢复再增补）：  
     | Workflow 文件 | Job 名称 | 描述 | 是否计划 Required |  
     |---------------|---------|------|-------------------|  
     | `.github/workflows/document-sync.yml` | `document-sync` | 双写/文档一致性 | 是 |  
     | `.github/workflows/consistency-guard.yml` | `consistency-guard` | CQRS、命名守卫 | 是 |  
     | `.github/workflows/plan-254-gates.yml` | `plan-254-gates` | Contract Drift | 是 |  
     | `.github/workflows/api-compliance.yml` | `api-compliance` | REST 契约守卫 | 是 |  
     | `.github/workflows/iig-guardian.yml` | `iig-guardian` | Implementation Inventory 守卫 | 先观测，后 Required |  
   - 修改各 workflow：  
     - 清理遗留的 `[self-hosted,cubecastle,docker]` 标签，统一为 `runs-on: [self-hosted,cubecastle,wsl]` 并通过 matrix 控制触发场景；  
     - 引入统一的 `prepare-selfhosted.sh <job>` step（例如 `bash scripts/ci/workflows/prepare-selfhosted.sh frontend-quality-gate`）；  
     - 对 Playwright/E2E job，复用 `docker inspect` 健康轮询 + `make run-dev` / `frontend/scripts/devserver-wait.sh`；  
     - 对 Go/SQL 守卫 job，设置 `CI_PREPARE_RUN_MIGRATIONS=1` 调用脚本以执行 `make db-migrate-all`，确保数据库来自 Compose（禁止 host 安装）。  
   - 每条 workflow 至少运行 2 次成功 run：一次来自 self-hosted，另一次来自 GitHub 托管（如仍保留）。在 `docs/development-plans/264` 更新 run ID，确保唯一事实来源指向自托管方案。

4. **WSL Runner 运行记录（Plan 269）**  
   - 至少执行一次 `document-sync (selfhosted)`、`api-compliance (selfhosted)`、`consistency-guard (selfhosted)`、`ci-selfhosted-smoke` 在 `runs-on: [self-hosted, cubecastle, wsl]` 标签下的成功 run，并把 Run ID + `logs/wsl-runner/*.log`/`~/actions-runner/_diag/` 截图记录到 Plan 265/266/269。  
   - `scripts/ci/runner/wsl-verify.sh` 的输出需附在 `logs/wsl-runner/verify-*.log`，同时在本计划文档登记最近一次执行时间。  
   - 若 WSL Runner 故障，应在 30 分钟内完成停机/替换或提交新计划，相关 run ID、日志与恢复步骤必须记录在 Plan 265/266。  
   - 2025-11-20 07:11Z：`bash scripts/ci/runner/wsl-install.sh` 已在 WSL 环境重新拉起 `cc-runner`（日志 `logs/wsl-runner/install-20251120T071110.log` / `run-20251120T071113.log`，`wsl-verify` 日志 `logs/wsl-runner/verify-20251120T071156.log`），但 07:16Z `workflow_dispatch` 触发的 `document-sync` run `19519517913` 仍只生成 docker/ubuntu matrix——远端 `.github/workflows/document-sync.yml` 未合入 `selfhosted-wsl`。需先推送 workflow 变更再重新触发，才能满足本节验收。  
   - 2025-11-20 07:42Z：`ci-selfhosted-smoke` 通过 `workflow_dispatch` 运行 `19520064684`，`Smoke (wsl)` job 成功完成并将日志导出到 `logs/wsl-runner/ci-selfhosted-smoke-wsl-19520064684.log`（docker job 仍失败，结论=failed，但 WSL job 可视作首个成功记录）。  
   - 2025-11-20 08:05Z：因 GitHub `workflow_dispatch` 在 WSL Runner 上持续 204/无 run，`document-sync`、`api-compliance`、`consistency-guard` 等 Required workflow 临时改回 `runs-on: ubuntu-latest` 验证流程，现阶段仅 `ci-selfhosted-smoke` 在 WSL Runner 上运行；待平台恢复后再逐步迁回 WSL。

5. **Branch Protection 更新**  
   - 根据 run 稳定性，将自托管 job 的 status 名字加入 Required checks：`Frontend Quality Gate (self-hosted)`、`Frontend E2E (self-hosted)`、`Document Sync (self-hosted)` 等；  
   - 若暂不想完全替换，可采用“ubuntu + self-hosted 双 Required”，待观察稳定性后再移除 ubuntu 分支。

6. **回滚路径**  
   - 每个 workflow 在 YAML 内保留注释说明如何回退到 `runs-on: ubuntu-latest`；  
   - 若自托管 runner 故障，可通过 `workflow_dispatch` 触发 ubuntu-only job 并在 Branch Protection 暂时移除 self-hosted 项；Plan 265 文档需记录回滚时间/原因。

## 4. 验收标准（已搁置）

> **搁置说明（2025-11-20）**：由于 WSL Runner 网络与调度仍不稳定（参考 Plan 266/267），`document-sync`、`api-compliance`、`consistency-guard` 等自托管 job 暂无法获取成功 run。以下验收项保持原描述，待 Runner 可用且网络恢复后再恢复执行：

- [ ] `contract-testing.yml` 中 `performance-impact-analysis` job 在 self-hosted runner 上 0 error，通过至少 3 次 PR run，并列入 Branch Protection Required 列表。  
- [ ] `frontend-quality-gate.yml`、`frontend-e2e.yml`、`document-sync.yml`、`consistency-guard.yml`、`plan-254-gates.yml`、`api-compliance.yml`、`iig-guardian.yml` 均已启用，且最新 push 在 self-hosted runner 上成功运行（含 run ID 记录）。  
- [ ] `scripts/ci/workflows/prepare-selfhosted.sh`（或等效）落库并被上述 workflow 调用，Compose/Docker 健康检查日志清晰。  
- [ ] Branch Protection 页面可见新增的 self-hosted status checks；Plan 263/264 文档同步更新。  
- [ ] 至少一次 `document-sync (selfhosted)` / `api-compliance (selfhosted)` / `consistency-guard (selfhosted)` / `ci-selfhosted-smoke` 使用 `self-hosted,cubecastle,wsl` 标签运行成功，Run ID 与日志写入 Plan 265/266/269。  
- [ ] 出现故障时的回滚步骤已记录，能够在 30 分钟内切回托管 runner。

## 5. 风险与缓解

| 风险 | 描述 | 缓解措施 |
|------|------|---------|
| 自托管 runner 资源被前端构建占满 | `frontend-e2e`/`build:verify` 同时运行可能耗尽 CPU/内存 | Watchdog 限制并发（`MaxParallelism=1`），必要时扩容第二个 runner 或把部分 job 保留在 ubuntu-latest |
| Docker Compose 服务残留 | 多个 workflow 同时 `up -d` 可能导致脏数据 | `prepare-selfhosted.sh` 中增加 `docker compose down --remove-orphans`、`docker volume prune --filter label=cubecastle-ci` 清理逻辑 |
| Branch Protection 切换风险 | 新增的 self-hosted status 失败会阻塞所有 PR | 先在非 Required 状态下运行 3+ 次，确认稳定后再切换；同时记录回滚命令 |
| Playwright 依赖更新 | Runner 持久化节点需要维护浏览器版本 | 每周由 Watchdog 触发一次 `npx playwright install --with-deps`，并在 Plan 265 中记录维护窗口 |
| WSL Runner 漂移 | WSL 内的工具链/代理版本不一致，导致 CI 结果不可复现 | 每次安装前运行 `wsl-verify.sh`，Go/Node/Docker 版本不符立即阻断；Plan 265/266 登记所有更改，并确保 `logs/wsl-runner/*` 可追溯 |

## 6. 时间表（建议）

- **Week 0（当前）**：完成本计划文档并获批准。  
- **Week 1**：  
  - 落地 `prepare-selfhosted.sh`；  
  - 启用/调整 `frontend-quality-gate`、`frontend-e2e`、`document-sync`、`consistency-guard`、`plan-254-gates` 等 workflow；  
  - 开始运行自托管 job 并记录 run ID。  
- **Week 2**：  
  - 完成 Plan 263 TS 修复，`performance-impact-analysis` 在 self-hosted 上稳定通过；  
  - 将上述 workflow 切换到 Required 自托管状态（如已稳定）。  
- **Week 3**：  
  - 回顾与回滚验证：模拟 runner 故障并验证回滚流程；  
  - 更新 Plan 263/264 文档、Branch Protection 截图、CHANGELOG。  

## 7. 依赖与协作

- DevInfra：维护自托管 runner 主机权限、watchdog 日志；  
- 前端团队：完成 TS 修复、维护 `frontend-e2e`/`quality-gate` 依赖；  
- 后端/文档团队：确保 `document-sync`、`consistency-guard` 需要的脚本与数据库迁移保持最新；  
- 安全：审计自托管 runner 挂载 `/var/run/docker.sock` 的风险并备案。

## 8. 更新记录

- 2025-11-19：v0.1 草拟，定义范围、步骤与验收标准。 (BY: Codex)
- 2025-11-20：补充 Plan 269 批准的 WSL Runner 例外、运行记录与风险条目；统一 `runs-on` 标签为 `[self-hosted,cubecastle,wsl]` 并扩展验收要求。
