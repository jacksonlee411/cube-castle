# WSL 自托管 Runner 安装与回滚手册（Plan 269）

**最后更新**：2025-11-20  
**适用范围**：经 Plan 269 批准的 WSL 环境（Windows 11 + WSL2 Ubuntu 22.04 默认），Runner 仅用于 GitHub Actions，自身仍依赖 Docker Compose 运行业务服务。

## 0. 前置检查

| 项目 | 要求 | 验证命令 |
|------|------|----------|
| WSL 发行版 | Ubuntu 20.04+（推荐 22.04） | `lsb_release -a` |
| Go | ≥ 1.24.9（遵循 AGENTS.md） | `go version` |
| Node.js | ≥ 18.x（与 `.nvmrc` 一致） | `node --version` |
| Docker CLI & Compose | 可访问宿主 Docker Daemon | `docker version && docker compose version` |
| GitHub PAT 或 Registration Token | PAT 需具备 `repo`,`workflow`，写入 `secrets/.env.local` | 检查 `echo $GH_RUNNER_PAT` 或文件 |
| 代理/hosts 配置 | 若需代理，请先依照 `docs/reference/docker-network-playbook.md` 设置 | `scripts/network/verify-github-connectivity.sh --smoke` |

执行 `bash scripts/ci/runner/wsl-verify.sh --preflight` 可一次性检查上述依赖。

## 1. 安装

```bash
# 1) 进入仓库根目录（WSL）
cd ~/cube-castle

# 2) 准备 secrets（至少二选一）
cat > secrets/.env.local <<'EOF'
GH_RUNNER_PAT=<personal-access-token>   # 推荐
RUNNER_HTTP_PROXY=http://proxy.example.com:8080   # 可选
RUNNER_HTTPS_PROXY=http://proxy.example.com:8080  # 可选
RUNNER_NO_PROXY=localhost,127.0.0.1,.local
EOF

# 3) 执行安装脚本
bash scripts/ci/runner/wsl-install.sh \
  --repo jacksonlee411/cube-castle \
  --runner-dir "$HOME/actions-runner" \
  --labels self-hosted,cubecastle,wsl

# 4) 观察输出：脚本会
#    - 校验 Go/Node/Docker 版本
#    - 下载 actions-runner tarball
#    - 通过 GH API 申请 registration token（或复用 GH_RUNNER_REG_TOKEN）
#    - 调用 ./config.sh --unattended --runnergroup Default
#    - 启动 tmux 守护进程：session 名 `cc-runner`
```

> 若 WSL 已启用 systemd，可加 `--use-systemd`，脚本会调用 `sudo ./svc.sh install/start` 并写入 `/etc/systemd/system/actions.runner.<repo>.<name>.service`。

## 2. 常用环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `RUNNER_REPO` | GitHub 仓库 `owner/repo` | `jacksonlee411/cube-castle` |
| `RUNNER_URL` | Runner 注册 URL | `https://github.com/${RUNNER_REPO}` |
| `RUNNER_DIR` | 解压目录 | `$HOME/actions-runner` |
| `RUNNER_WORKDIR` | `_work` 路径 | `$RUNNER_DIR/_work` |
| `RUNNER_NAME` | Runner 名称 | `wsl-$(hostname)` |
| `RUNNER_LABELS` | 逗号分隔标签 | `self-hosted,cubecastle,wsl` |
| `RUNNER_VERSION` | actions/runner 版本 | 读取 `RUNNER_VERSION` 或脚本内默认 |
| `RUNNER_HTTP_PROXY`/`RUNNER_HTTPS_PROXY` | 可选代理 | 读取 `secrets/.env.local` |
| `RUNNER_TMUX_SESSION` | 守护 tmux session | `cc-runner` |

可在执行脚本前导出变量或通过 `--runner-dir`/`--labels` 参数覆盖。

## 3. 验证

1. `bash scripts/ci/runner/wsl-verify.sh`：输出版本、Docker socket、GitHub API 可访问性；若失败脚本返回非 0。
2. `tmux ls` 确认存在 `cc-runner`（或自定义）session；`tmux logs` 里应出现 Runner 心跳。
3. GitHub → Settings → Actions → Runners：在线 Runner 应显示 `self-hosted`, `cubecastle`, `wsl` 标签。
4. 触发 `CI (Self-Hosted Runner Smoke)` 或 `document-sync` workflow 并记录 Run ID、日志路径到 Plan 265/266。

## 4. 日志与维护

- Runner 目录：`~/actions-runner/_diag/`（官方日志）、`~/actions-runner/run.log`（启动输出）
- WSL Runner 专用日志：`logs/wsl-runner/install-*.log`, `logs/wsl-runner/verify-*.log`
- Watchdog：可由 `scripts/ci/runner/watchdog.sh` 扩展，定期执行 `wsl-verify.sh` 与 `scripts/network/verify-github-connectivity.sh`
- 定期升级：设置 `RUNNER_VERSION=<新版本>` 后重新运行 `wsl-install.sh`。脚本会先停止正在运行的守护进程，再下载目标版本并调用 `./config.sh --replace`。

## 5. 卸载与回滚

```bash
# 1) 停止 Runner（tmux 模式）
bash scripts/ci/runner/wsl-uninstall.sh \
  --repo jacksonlee411/cube-castle \
  --runner-dir "$HOME/actions-runner"

# 2) 可选：WSL systemd 模式
bash scripts/ci/runner/wsl-uninstall.sh --use-systemd

# 脚本会：
#   - 获取 removal token（使用 GH_RUNNER_PAT 或 GH_RUNNER_REG_TOKEN）
#   - 调用 ./svc.sh stop/remove 或 tmux kill-session
#   - 执行 ./config.sh remove --token
#   - 备份并删除 Runner 目录
```

回滚到 Docker Runner：运行 `bash scripts/ci/runner/start-ghcr-runner-persistent.sh`，并在 workflow 中暂时移除 `wsl` 标签（若 WSL 故障）。

## 6. 故障排查

| 症状 | 检查项 |
|------|--------|
| `config.sh` 报 Already configured | 设置 `FORCE_RECONFIGURE=true` 或先运行 `wsl-uninstall.sh` |
| 无法访问 Docker Daemon | 检查 Docker Desktop 设置 → Resources → WSL Integration；`sudo chown $USER /var/run/docker.sock` |
| 网络/TLS 失败 | 执行 `scripts/network/verify-github-connectivity.sh`、参考 `docs/reference/docker-network-playbook.md` |
| tmux session 丢失 | `tmux new -d -s cc-runner "cd $RUNNER_DIR && ./run.sh"` 手动重启，并查看 `_diag` 日志 |

## 7. 责任与记录

- 所有安装/卸载均需在 Plan 265/266 中登记：时间、操作者、脚本日志、相关 workflow Run ID。
- 若修改代理/hosts、Docker context 等网络设置，请同步更新 `docs/reference/docker-network-playbook.md` 与 Watchdog 配置。
- 发生安全事件（权限越权、文件泄露等）必须第一时间停机、审计日志并通知仓库维护者。
