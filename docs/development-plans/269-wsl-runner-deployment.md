# Plan 269 - WSL 自托管 Runner 部署可行性评估

**文档编号**: 269  
**创建日期**: 2025-11-20  
**关联计划**: Plan 262（自托管 Runner 基建）、Plan 265（Required Checks）、Plan 267（网络稳定化）

---

## 1. 背景与目标

- 当前自托管 Runner 运行在 Docker 容器内（WSL2 宿主），但容器缺少 `docker compose`、Go/Node 等工具，需要额外补齐；同时容器内网络受 WSL/防火墙双重影响，导致 GitHub TLS、Compose 等步骤频繁失败。
- 2025-11-20 经架构/安全/平台负责人联合评审，确认“Runner 属于 CI 基础设施，不在 Docker 强制约束的业务服务范围内”，允许在满足 Docker 服务依旧运行在容器中的前提下，引入“WSL 原生 Runner”作为官方备选方案。该结论需同步更新 `AGENTS.md` 与参考文档，避免事实来源分裂。
- Plan 269 旨在评估并落地“在 WSL 内直接部署 Runner（Systemd service 或 CLI 模式）”的可行性，对照现有容器方案的优缺点，形成部署步骤、回滚方式、CI workflow 更新以及与仓库原则的兼容性说明，同时保持 Docker Runner 可用以便随时回退。

---

## 2. 范围与交付物

| 类别 | 交付物 | 说明 |
|------|--------|------|
| 方案比较 | `docs/reference/wsl-runner-comparison.md` | Docker Runner vs WSL Runner trade-off（安全、隔离、可复制性、维护成本），明确默认推荐：优先 Docker，WSL 为经批准的备选 |
| 部署指南 | `scripts/ci/runner/wsl-install.sh` + `scripts/ci/runner/wsl-uninstall.sh` + `docs/reference/wsl-runner-setup.md` | 覆盖依赖安装、环境变量、systemd/守护脚本、日志位置、Go/Node/Docker 版本校验 |
| 网络与安全评估 | Plan 267 更新 + `docs/reference/docker-network-playbook.md` & `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` | 追加 WSL 直连下的网络诊断、代理/hosts 回退流程，以及 Runner 隔离策略 |
| 契约同步 | 更新 `AGENTS.md`、Plan 265/266/267 | 声明 WSL Runner 获批例外、记录残余风险与 Required Checks 变更 |
| 回滚策略 | `docs/reference/wsl-runner-setup.md` 中的 uninstall & fallback 章节 | 描述如何安全卸载 Runner、清理 systemd 服务、恢复 workflow `runs-on` 为 Docker 标签 |

不在本计划范围：修改业务服务（command/query/frontend）、改变 Docker Compose 的端口/镜像设置、对 Runner 做功能增强（仅关注部署方式）。

---

## 3. 实施步骤

### 3.1 方案调研与比较
1. 盘点现有 Runner 架构：`docker-compose.runner.persist.yml`、`runner/persistent-entrypoint.sh`、`scripts/ci/runner/*`，明确 Docker Runner 仍为主路径。
2. 收集 WSL 直接运行 Runner 的官方指南（GitHub Actions Runner on Linux，systemd service），列出差异点（WSL 默认无 systemd，可通过 `systemd-genie`/`tmux`/`nohup`）。
3. 记录审批依据：将 2025-11-20 的跨团队批准摘要写入 `docs/reference/wsl-runner-comparison.md` 与 `AGENTS.md`，强调“仅 Runner 属于例外，业务服务依旧必须运行在 Docker Compose 中”。
4. 输出 `docs/reference/wsl-runner-comparison.md`，包含：
   - 环境一致性（Docker 镜像 vs WSL 的 apt install）
   - 安全隔离（容器 vs WSL 用户隔离 + 专用 WSL 实例建议）
   - 调试/维护成本
   - 网络影响（WSL 直接使用宿主代理 vs 容器内 hosts）
   - 默认推荐与回退路径

### 3.2 WSL Runner 部署脚本
1. 编写 `scripts/ci/runner/wsl-install.sh`：
   - 检查依赖：`curl`, `tar`, `tmux` 或 `systemd-run`、`go`、`node`、`docker` CLI + `docker compose`，通过 `go version`、`node --version` 校验是否满足 AGENTS 基线（Go 1.24.9+、Node 18+），若缺失则引导安装。
   - 验证 Docker Desktop/WSL 集成：执行 `docker version`/`docker context show`，确保 Runner 可以访问宿主 Docker Daemon。
   - 下载官方 `actions-runner-linux-x64-<version>.tar.gz`，解压至 `~/actions-runner`（路径可配置）。
   - 读取 `secrets/.env.local` 的 `GH_RUNNER_PAT` 或临时 token，执行 `./config.sh --url ... --labels self-hosted,cubecastle,wsl`，并在脚本中记录标签默认值可覆盖。
   - 启动方式：若 systemd 可用则 `sudo ./svc.sh install/start`，否则提供 `tmux`/`nohup` 守护脚本，并把日志写入 `~/actions-runner/_diag` + `/var/log/cube-castle/wsl-runner.log`。
2. 编写 `scripts/ci/runner/wsl-uninstall.sh`：停止守护进程/服务、`./config.sh remove`、删除 systemd 单元/`tmux` Session、清理目录及 `sudoers` 临时配置。
3. 增补 `scripts/ci/runner/wsl-verify.sh`：执行工具链版本核对、Docker socket 连通性、GitHub API 自检。
4. 在 `docs/reference/wsl-runner-setup.md` 中记录安装步骤、环境变量（`RUNNER_NAME`, `RUNNER_LABELS`, `RUNNER_WORKDIR`）、日志路径、验证方式（`gh api repos/.../actions/runners`）与版本检查输出示例。

### 3.3 网络与安全检查
1. 更新 Plan 267：说明 WSL 直连后 hosts/代理的设置（`/etc/hosts`、`/etc/resolv.conf`、`wsl.exe --shutdown`）以及 GitHub/TLS 诊断脚本如何执行。
2. 如果企业网络限制仍存在，提供 fallback：WSL 侧 `https_proxy`、`git config --global http.proxy`，同时在安装/验证脚本里检测并提示；必要时自动注入 hosts（Plan 267 脚本复用）。
3. 安全性：说明 WSL Runner 运行在当前用户上下文，建议使用专用 WSL 实例或 Windows 用户隔离；在 Plan 269 中记录残余风险，并在 `AGENTS.md` 与 `docs/reference/wsl-runner-setup.md` 标注“Runner 例外 + 隔离建议”。
4. ⚠️ 关闭或重启 WSL（包括执行 `wsl.exe --shutdown`）会导致 Runner 与 Docker 网络短暂停机，属于高影响操作；执行前必须在协作渠道说明命令、影响面与回滚方案，取得额外审批后方可进行，并在 Plan 265/266/269 中登记。

### 3.4 Pipeline 集成与验证
1. 更新所有使用自托管 Runner 的 workflow（例如 `document-sync`, `ci-selfhosted-smoke`, `ci-selfhosted-diagnose`, `consistency-guard`, `api-compliance` 等）：
   - `runs-on` 新增 `wsl` 标签（`[self-hosted, cubecastle, wsl]`），并在矩阵中保留 Docker Runner 作为备份。
   - 记录在 workflow 注释中：WSL Runner 需具备 Docker CLI，任务仍依赖 Docker Compose。
2. 在 `docs/development-plans/265-selfhosted-required-checks.md` 追加“WSL Runner”执行记录：包含安装脚本、Run ID、日志路径、使用标签。
3. 通过 `workflow_dispatch` 触发 `document-sync`、`api-compliance`、`consistency-guard`、`ci-selfhosted-smoke`，确保新的 Runner 标签生效并收集日志。
4. 若成功，将 Plan 269 的结论写入 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`，说明 Docker vs WSL 的选择指引、推荐顺序（默认 Docker，紧急或网络限制下可切换 WSL），并在 `AGENTS.md` 引用该指南。
5. 当前执行情况：2025-11-20 07:16Z 已以 `workflow_dispatch` 方式触发 `document-sync`（run `19519517913`），但由于远端 `document-sync.yml` 尚未推送 `selfhosted-wsl` 矩阵，该 run 仅在 `cc-runner-docker-compose` 上执行并在“质量门禁”阶段失败，WSL Runner 未获得任务；需先合并 workflow 变更后重新触发上述四个 workflow。

---

## 4. 验收标准

- [ ] `AGENTS.md` 与 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 完成同步更新，清楚说明“业务服务仍强制 Docker，Runner 获批 WSL 例外 + 使用场景”。
- [ ] `docs/reference/wsl-runner-comparison.md` 发布，明确列出 Docker vs WSL Runner 的差异、审批依据、推荐场景与默认顺序。
- [ ] `scripts/ci/runner/wsl-install.sh`/`wsl-uninstall.sh`/`wsl-verify.sh` 编写完成，具备 Go/Node 版本检查、Docker CLI 检测与日志输出；在 README 或 `docs/reference/wsl-runner-setup.md` 中提供示例命令。
- [ ] 至少一次 `document-sync (selfhosted)` 使用 `self-hosted,cubecastle,wsl` 标签成功运行，Run ID 记录在 Plan 265/269，并保留 Docker Runner 成功日志作为回退佐证。
- [ ] `docs/reference/wsl-runner-setup.md` 详细描述安装、运行、日志、版本校验与回滚步骤，并在 `docs/reference/docker-network-playbook.md`、Plan 267 中注明 WSL 网络指引。
- [ ] 所有依赖自托管 Runner 的 workflow 已更新 `runs-on` 标签并通过一次实际执行（记录 Run ID），残余风险登记到 Plan 265/266。

---

## 5. 风险与回滚

| 风险 | 描述 | 缓解/回滚 |
|------|------|-----------|
| 违反 Docker 强制原则 | WSL Runner 可能被理解为“非 Docker 部署” | 在文档中说明：业务服务仍在 Docker 内，Runner 仅为 CI 工具，得到架构/安全认可后方可采纳 |
| 环境漂移 | 不同 WSL 实例的依赖版本/路径不一致 | 提供脚本自动安装 Go/Node/Docker，记录版本；定期执行 `scripts/ci/runner/wsl-verify.sh` |
| 安全隔离弱 | Runner 直接运行在 WSL 用户态，Workflow 命令可访问宿主文件 | 建议在专用 WSL 实例中运行，或结合 Windows 用户权限隔离；必要时继续使用 Docker 方案 |
| 网络仍受限制 | 即使 WSL 直连，企业代理仍断流 | 继续依赖 Plan 267 的 hosts/代理脚本；在失败时回退到 Docker 方案 |
| 维护成本增加 | 需要同时维护 Docker 与 WSL 两种 runner | Plan 269 结论将给出默认推荐（例如优先 WSL，Docker 作为备选），避免双线维护 |

回滚：执行 `scripts/ci/runner/wsl-uninstall.sh` 删除 Runner，恢复 `document-sync` 等 workflow 的 `runs-on` 仅指向 Docker Runner；必要时恢复 `docker-compose.runner.persist.yml` 方案。

---

## 6. 里程碑

- **M1（2025-11-21）**：完成方案对比文档 & 脚本草稿，获取架构/安全认可。
- **M2（2025-11-22）**：在 WSL Runner 上跑通 `document-sync`、`api-compliance`；记录 Run ID。
- **M3（2025-11-24）**：更新 CI 指南、Plan 265/266，给出最终推荐（采用或仅作为备选）。

---

## 7. 参考资料

- AGENTS.md（Docker 强制、环境一致性原则）
- Plan 262/265/266/267（自托管 Runner 与网络治理）
- GitHub Actions 官方 Runner 文档：<https://github.com/actions/runner>
- `scripts/ci/runner/` 目录现有脚本（docker 版启动/守护）
