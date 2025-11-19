# Plan 269 - WSL 自托管 Runner 部署可行性评估

**文档编号**: 269  
**创建日期**: 2025-11-20  
**关联计划**: Plan 262（自托管 Runner 基建）、Plan 265（Required Checks）、Plan 267（网络稳定化）

---

## 1. 背景与目标

- 当前自托管 Runner 运行在 Docker 容器内（WSL2 宿主），但容器缺少 `docker compose`、Go/Node 等工具，需要额外补齐；同时容器内网络受 WSL/防火墙双重影响，导致 GitHub TLS、Compose 等步骤频繁失败。
- AGENTS.md 强调“所有服务必须由 Docker 管理”，但 Runner 本身只是 CI 工具而非业务功能，若在 WSL 子系统内直接部署（非再套一层 Docker），可减少网络/依赖问题。
- Plan 269 旨在评估“在 WSL 内直接部署 Runner（Systemd service 或 CLI 模式）”的可行性：对照现有容器方案的优缺点，形成部署步骤、回滚方式以及与仓库原则的兼容性说明。

---

## 2. 范围与交付物

| 类别 | 交付物 | 说明 |
|------|--------|------|
| 方案比较 | `docs/reference/wsl-runner-comparison.md` | Docker 内 vs WSL 直接部署的 trade-off（安全、隔离、可复制性、维护成本） |
| 部署指南 | `scripts/ci/runner/wsl-install.sh` + `docs/reference/wsl-runner-setup.md` | 覆盖依赖安装、环境变量、systemd/守护脚本、日志位置 |
| 网络与安全评估 | 更新 Plan 267 + `docs/reference/docker-network-playbook.md` | 标注 WSL 直连对 hosts/代理/TLS 的影响，提供 fallback（hosts 覆盖或企业代理） |
| 回滚策略 | `docs/reference/wsl-runner-setup.md` 中的 uninstall 章节 | 描述如何安全卸载 Runner、清理 systemd 服务、恢复 hosts |

不在本计划范围：修改业务服务（command/query/frontend）、改变 Docker Compose 的端口/镜像设置、对 Runner 做功能增强（仅关注部署方式）。

---

## 3. 实施步骤

### 3.1 方案调研与比较
1. 盘点现有 Runner 架构：`docker-compose.runner.persist.yml`、`runner/persistent-entrypoint.sh`、`scripts/ci/runner/*`。
2. 收集 WSL 直接运行 Runner 的官方指南（GitHub Actions Runner on Linux，systemd service），列出差异点（WSL 无 systemd，需要 `svc.sh install` + `nohup`/`tmux` 等）。
3. 输出 `docs/reference/wsl-runner-comparison.md`，包含：
   - 环境一致性（Docker 镜像 vs WSL 的 apt install）
   - 安全隔离（容器 vs WSL 用户）
   - 调试/维护成本
   - 网络影响（WSL 直接使用宿主代理 vs 容器内 hosts）

### 3.2 WSL Runner 部署脚本
1. 编写 `scripts/ci/runner/wsl-install.sh`：
   - 检查依赖：`curl`, `tar`, `systemd-run`/`nohup`, `go`, `node`, `docker`（可选）。
   - 下载官方 `actions-runner-linux-x64-<version>.tar.gz`，解压至 `~/actions-runner`.
   - 读取 `secrets/.env.local` 的 `GH_RUNNER_PAT` 或临时 token，执行 `./config.sh --url ... --labels self-hosted,cubecastle,wsl`.
   - 启动方式：若 systemd 可用则 `sudo ./svc.sh install/start`；若不可用，记录 `nohup ./run.sh &` 或 `tmux new -d` 方案。
2. 编写 `scripts/ci/runner/wsl-uninstall.sh`：停止服务、`./config.sh remove`、清理目录。
3. 在 `docs/reference/wsl-runner-setup.md` 中记录安装步骤、环境变量（`RUNNER_NAME`, `RUNNER_LABELS`）、日志路径与验证方式（`gh api repos/.../actions/runners`）。

### 3.3 网络与安全检查
1. 更新 Plan 267：说明 WSL 直连后 hosts/代理的设置（`/etc/hosts`、`/etc/resolv.conf`、`wsl.exe --shutdown`）以及 GitHub/TLS 诊断脚本如何执行。
2. 如果企业网络限制仍存在，提供 fallback：WSL 侧 `https_proxy`、`git config --global http.proxy`，同时在脚本里检测并提示。
3. 安全性：说明 WSL Runner 运行在当前用户上下文，若需要隔离，可结合 Windows 防火墙或单独的 WSL 实例；在 Plan 269 中记录残余风险。

### 3.4 Pipeline 集成与验证
1. 在 `docs/development-plans/265-selfhosted-required-checks.md` 追加“WSL Runner”执行记录：包含安装脚本、Run ID、日志路径。
2. 通过 `workflow_dispatch` 触发 `document-sync`、`api-compliance`、`consistency-guard`，确保新的 Runner 标签（例如 `self-hosted,cubecastle,wsl`) 能被 workflow 识别（`runs-on` 需包含 `wsl` 标签或 `self-hosted` 组合）。
3. 若成功，将 Plan 269 的结论写入 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`，说明两种部署方式的选择指引。

---

## 4. 验收标准

- [ ] `docs/reference/wsl-runner-comparison.md` 发布，明确列出 Docker vs WSL Runner 的差异及推荐场景。
- [ ] `scripts/ci/runner/wsl-install.sh`/`wsl-uninstall.sh` 编写完成，并在 README 中提供示例命令。
- [ ] 至少一次 `document-sync (selfhosted)` 使用 WSL Runner 成功运行，Run ID 记录在 Plan 265/269。
- [ ] `docs/reference/wsl-runner-setup.md` 详细描述安装、运行、日志与回滚步骤，并更新 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 指向该文档。
- [ ] 评估结果在 Plan 269 中给出明确结论（是否推荐/限定条件），并把残余风险登记到 Plan 265/266。

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
