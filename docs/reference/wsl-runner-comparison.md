# WSL Runner vs Docker Runner 对比指引（Plan 269）

> ⚠️ **2025-11-22 更新**：Plan 269 已撤销 WSL Runner 方案，本指引仅保留历史内容，勿据此重新启用 `self-hosted,cubecastle,wsl`。所有 workflow 必须运行在 GitHub 平台 runner 上，如需新的自托管方案须另行审批；历史 Runbook 见 `docs/archive/development-plans/05-CI-LOCAL-AUTOMATION-GUIDE.md`。

**最后更新**：2025-11-20（Plan 269 批准，阶段性策略：GitHub runner 为主、WSL Runner 仅用于 smoke）  
**关联文档**：AGENTS.md、Plan 262/265/266/267、`docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`

## 背景与结论

- 业务服务、数据库、中间件仍受“Docker Compose 强制”约束；任何生产/研发服务不得在宿主或 WSL 中裸跑。
- 2025-11-20 架构 + 安全 + 平台联合评审曾确认：CI 自托管 Runner 属于开发工具，可在满足隔离/日志条件下于 WSL 内运行，仅 `ci-selfhosted-smoke` 运行在 WSL，其余 Required workflow 继续使用 GitHub Runner。
- 2025-11-22 最终决议：上述 WSL/Docker Runner 方案全部停用，仓库禁止启用任何 `self-hosted` 标签，下文仅保留历史对比供审计参考。

## 维度对比

| 维度 | Docker Runner（历史，仅供参考） | WSL Runner（历史，仅供参考） | 备注 |
|------|-------------------------------|-------------------------|------|
| 环境一致性 | 通过镜像固定 `docker compose`、Go、Node、CA 等；便于复制/快照 | 依赖 WSL 发行版（Ubuntu 20.04/22.04 等），使用 `wsl-install.sh` 自动校验工具链 | WSL 需定期执行 `wsl-verify.sh` 保持版本一致性 |
| 网络链路 | Docker Desktop/WSL → Runner 容器，多层代理易失效 | 直接使用 WSL 网络栈，可复用宿主代理/hosts | `docs/reference/docker-network-playbook.md` 已针对 WSL 更新 |
| 安全隔离 | Docker 容器 + `docker.sock` 权限边界清晰 | Runner 运行在 WSL 用户态，可访问宿主文件 | 建议专用 WSL 实例 + Windows 账户隔离 |
| 维护成本 | 需维护镜像/Compose 版本 | 需维护 WSL 发行版和脚本 | 统一脚本已集中在 `scripts/ci/runner` 目录 |
| 回滚 | `docker compose down` 即可；镜像可快照 | 通过 `wsl-uninstall.sh` 卸载后重新安装 | 若需重新启用 Docker Runner，必须重新立项审批 |
| CI 工作流兼容 | 旧版 workflows 使用 `[self-hosted,cubecastle,docker]` | 历史上使用 `[self-hosted,cubecastle,wsl]` | 2025-11-22 起全部切回 `ubuntu-latest`

## 历史选择流程（归档，不可执行）

> 以下步骤仅保留在案，用于回溯 2025-11-20 前的操作路径；当前禁止执行这些命令。如需新 Runner，请先完成立项与审批。

1. **安装与配置**：参考 `docs/reference/wsl-runner-setup.md` + `scripts/ci/runner/wsl-install.sh`，标签设置为 `self-hosted,cubecastle,wsl`。
2. **验证**：运行 `scripts/ci/runner/wsl-verify.sh`、`scripts/network/verify-github-connectivity.sh --smoke`，随后触发 `ci-selfhosted-smoke` 记录 Run ID；Required workflow 需在 GitHub runner 上先跑绿，并将成果写入 Plan 265/266/269。
3. **迁移回 WSL**：计划中曾要求待平台问题解决后逐条将 workflow `runs-on` 切回 `[self-hosted,cubecastle,wsl]`，并在 WSL Runner 上收集成功 run 证据（该计划已取消）。

## 审批与审计

- 若未来确需重新启用任何自托管 Runner，必须在 `docs/development-plans/` 建立新计划并获批，同时在 Plan 265/266 登记 Run ID、触发人、原因与回滚时间。
- 历史日志（如 `~/actions-runner/_diag/`, `logs/wsl-runner/*.log`）只作证据保留；新方案需重新指定落盘方式与保留策略。
- 所有 Runner 策略以 AGENTS.md 与 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` 为准，未经审批禁止修改 workflow `runs-on`。

## 关联脚本 & 文档

- `scripts/ci/runner/wsl-install.sh`：安装 + 注册 Runner。
- `scripts/ci/runner/wsl-uninstall.sh`：卸载 Runner 并清理 systemd/tmux。
- `scripts/ci/runner/wsl-verify.sh`：检查 Go/Node、Docker、网络、`runsvc.sh`。
- `docs/reference/wsl-runner-setup.md`：操作手册、日志路径、回滚。
- `docs/reference/docker-network-playbook.md`：WSL/代理/hosts 配置指南。
- `scripts/network/verify-github-connectivity.sh`：网络诊断基线。

> ⚠️ **提醒**：无论历史 Runner 方案如何，业务服务（command/query/frontend、Postgres、Redis 等）始终必须通过 Docker Compose 管控；任何试图在宿主/WSL 直接运行的行为均视为违反“资源唯一性与跨层一致性”。
