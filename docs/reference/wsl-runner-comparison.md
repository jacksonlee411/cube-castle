# WSL Runner vs Docker Runner 对比指引（Plan 269）

**最后更新**：2025-11-20（Plan 269 批准）  
**关联文档**：AGENTS.md、Plan 262/265/266/267、`docs/reference/05-CI-LOCAL-AUTOMATION-GUIDE.md`

## 背景与结论

- 业务服务、数据库、中间件仍受“Docker Compose 强制”约束；任何生产/研发服务不得在宿主或 WSL 中裸跑。
- 2025-11-20 架构 + 安全 + 平台联合评审确认：CI 自托管 Runner 属于开发工具，可在满足隔离/回滚条件下于 WSL 内直接运行，以缓解容器内网络与依赖短板。
- 默认推荐顺序：
  1. **Docker Runner（推荐）**：保持与 CI/生产一致的 Compose 管理体验，优先用于常态 Required Checks。
  2. **WSL Runner（获批备选）**：当 Docker Runner 因代理/网络导致长时间不可用，或需要 WSL 原生工具链时启用；需同时保留 Docker Runner 作为热备。

## 维度对比

| 维度 | Docker Runner（默认） | WSL Runner（Plan 269 例外） | 决策提示 |
|------|---------------------|-----------------------------|----------|
| 环境一致性 | 通过镜像固定 `docker compose`、Go、Node、CA 等；便于复制/快照 | 依赖 WSL 发行版（Ubuntu 20.04/22.04 等），需要脚本安装工具链 | 若团队需“一键复制”环境优先 Docker；WSL 必须执行 `wsl-verify.sh` |
| 网络链路 | Docker Desktop/WSL → Runner 容器，多层代理易失效 | 直接使用 WSL 网络栈，可与 Windows 代理共用 | 网络问题集中在 Runner 容器时考虑 WSL 方案 |
| 安全隔离 | Docker 容器 + `docker.sock` 权限边界清晰 | Runner 运行在 WSL 用户态，可访问宿主文件 | WSL Runner 应放在专用 WSL 实例，并启用 Windows 账号隔离 |
| 维护成本 | 需要维护 Runner 镜像（更新 Compose/工具链） | 需要维护 WSL 发行版 + 依赖安装脚本 | 尽量统一脚本，通过 `scripts/ci/runner` 目录管理 |
| 回滚 | `docker compose down` 即可；镜像可回滚 | 需运行 `wsl-uninstall.sh`，并确认 systemd/tmux session 停止 | 任一失败场景都必须有 Docker Runner 备用 |
| CI 工作流兼容 | 现有 workflow 均使用 `[self-hosted,cubecastle,docker]` 标签 | Plan 269 要求 workflow 增加 `wsl` 标签 | 所有 workflow 使用统一矩阵 `[self-hosted,cubecastle,wsl]`，并保留 Docker 节点以防万一 |

## 选择流程

1. **先跑 Docker Runner**：根据 `Plan 265` 与 `scripts/ci/runner/start-*.sh` 接管日常任务；若失败，收集日志并执行 `scripts/network/verify-github-connectivity.sh`。
2. **评估切换条件**：
   - 网络链路阻塞超过 4 小时且已执行 Plan 267 Playbook 仍未恢复；
   - 需要在 Runner 内访问宿主 WSL 资源/代理；
   - 已准备好回滚方案，并有人员值守。
3. **执行 WSL Runner 安装**：严格按 `docs/reference/wsl-runner-setup.md` + `scripts/ci/runner/wsl-install.sh` 操作，标签设置为 `self-hosted,cubecastle,wsl`。
4. **验证**：运行 `scripts/ci/runner/wsl-verify.sh`、`scripts/network/verify-github-connectivity.sh --smoke`，随后触发 `document-sync (selfhosted)` workflow 并记录 Run ID。
5. **回退策略**：若任何 step 失败或出现安全/性能隐患，立即执行 `scripts/ci/runner/wsl-uninstall.sh`，恢复 workflow `runs-on` 指向 Docker 标签，同时在 Plan 265/266 登记。

## 审批与审计

- 任何 Runner 模式切换必须更新 Plan 265/266 的运行记录（Run ID、触发人、原因、回滚时间）。
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
