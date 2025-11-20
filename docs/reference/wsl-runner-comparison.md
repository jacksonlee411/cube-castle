# WSL Runner vs Docker Runner 对比指引（Plan 269）

**最后更新**：2025-11-20（Plan 269 批准，2025-11-20 结论更新：WSL 为唯一自托管方案）  
**关联文档**：AGENTS.md、Plan 262/265/266/267、`docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`

## 背景与结论

- 业务服务、数据库、中间件仍受“Docker Compose 强制”约束；任何生产/研发服务不得在宿主或 WSL 中裸跑。
- 2025-11-20 架构 + 安全 + 平台联合评审确认：CI 自托管 Runner 属于开发工具，可在满足隔离/日志条件下于 WSL 内直接运行；随后的 2025-11-20 决议要求 **自托管 Runner 仅保留 WSL 形态**，Docker Runner 退役，如需恢复必须重新立项。

## 维度对比

| 维度 | Docker Runner（历史，仅供参考） | WSL Runner（当前默认） | 备注 |
|------|-------------------------------|-------------------------|------|
| 环境一致性 | 通过镜像固定 `docker compose`、Go、Node、CA 等；便于复制/快照 | 依赖 WSL 发行版（Ubuntu 20.04/22.04 等），使用 `wsl-install.sh` 自动校验工具链 | WSL 需定期执行 `wsl-verify.sh` 保持版本一致性 |
| 网络链路 | Docker Desktop/WSL → Runner 容器，多层代理易失效 | 直接使用 WSL 网络栈，可复用宿主代理/hosts | `docs/reference/docker-network-playbook.md` 已针对 WSL 更新 |
| 安全隔离 | Docker 容器 + `docker.sock` 权限边界清晰 | Runner 运行在 WSL 用户态，可访问宿主文件 | 建议专用 WSL 实例 + Windows 账户隔离 |
| 维护成本 | 需维护镜像/Compose 版本 | 需维护 WSL 发行版和脚本 | 统一脚本已集中在 `scripts/ci/runner` 目录 |
| 回滚 | `docker compose down` 即可；镜像可快照 | 通过 `wsl-uninstall.sh` 卸载后重新安装 | 若需重新启用 Docker Runner，必须重新立项审批 |
| CI 工作流兼容 | 旧版 workflows 使用 `[self-hosted,cubecastle,docker]` | 现有 workflows 仅保留 `[self-hosted,cubecastle,wsl]` | `wsl_only` 条件控制触发场景

## 选择流程

1. **安装与配置**：严格按 `docs/reference/wsl-runner-setup.md` + `scripts/ci/runner/wsl-install.sh` 操作，标签设置为 `self-hosted,cubecastle,wsl`。
2. **验证**：运行 `scripts/ci/runner/wsl-verify.sh`、`scripts/network/verify-github-connectivity.sh --smoke`，随后触发 `ci-selfhosted-smoke`、`document-sync (selfhosted)` 等 workflow 并记录 Run ID。
3. **问题处理**：若出现网络/权限异常，按 Plan 265/266/267 的诊断脚本收集日志；无法在 WSL 内恢复时，需提交新的计划以重新启用 Docker Runner（默认不再保留）。

## 审批与审计

- 所有 WSL Runner 安装/卸载必须更新 Plan 265/266 的运行记录（Run ID、触发人、原因、回滚时间）。
- WSL Runner 节点必须保留日志：`~/actions-runner/_diag/`, `logs/wsl-runner/*.log`。必要时附在 Plan 269/266。
- 变更需通知仓库维护者，并在 `docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md` “Runner 选型”章节登记。

## 关联脚本 & 文档

- `scripts/ci/runner/wsl-install.sh`：安装 + 注册 Runner。
- `scripts/ci/runner/wsl-uninstall.sh`：卸载 Runner 并清理 systemd/tmux。
- `scripts/ci/runner/wsl-verify.sh`：检查 Go/Node、Docker、网络、`runsvc.sh`。
- `docs/reference/wsl-runner-setup.md`：操作手册、日志路径、回滚。
- `docs/reference/docker-network-playbook.md`：WSL/代理/hosts 配置指南。
- `scripts/network/verify-github-connectivity.sh`：网络诊断基线。

> ⚠️ **提醒**：虽然 WSL Runner 获批例外，但任何业务服务（command/query/frontend、Postgres、Redis 等）依旧只能通过 Docker Compose 启动。违反者视同破坏“资源唯一性与跨层一致性”。
