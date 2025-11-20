# 自托管 Runner（Docker Compose 管控：Ephemeral + 持久化）

本仓库遵循 AGENTS.md 的“Docker 强制”原则，自托管 Runner 必须由 Docker Compose 管理：  
- 默认：Ephemeral 一次性模式（`docker-compose.runner.yml`）  
- 持久化：常驻接单模式（`docker-compose.runner.persist.yml` + `persistent-entrypoint.sh` 幂等初始化，由 watcher 保活）

## 一、准备密钥（任选其一）
1) Registration Token（一次性）  
   GitHub → 仓库 → Settings → Actions → Runners → New self-hosted runner（Linux），复制 token；
   在宿主 `secrets/.env.local` 写入：
   ```bash
   GH_RUNNER_REG_TOKEN=<token>
   ```
2) PAT（个人访问令牌）  
   建议 scope：`repo` + `workflow`；写入：
   ```bash
   GH_RUNNER_PAT=<pat>
   ```

注意：`secrets/` 已被 .gitignore 忽略，谨防泄露。

## 二、启动与停止
```bash
# 启动 Ephemeral（一次性，Job 结束自动注销）
docker compose -f docker-compose.runner.yml up -d

# 启动持久化（推荐：自动申请令牌 + Compose 幂等入口）
bash scripts/ci/runner/start-ghcr-runner-persistent.sh
nohup bash scripts/ci/runner/watchdog.sh 60 > logs/ci-monitor/watchdog.out 2>&1 &

# 若已拥有 RUNNER_TOKEN，可手动启动持久化
RUNNER_TOKEN=... docker compose -f docker-compose.runner.persist.yml up -d

# 查看 Runner 容器日志
docker logs -f cubecastle-gh-runner

# 停止并清理
docker compose -f docker-compose.runner.yml down -v
# 或持久化：
docker compose -f docker-compose.runner.persist.yml down -v
# 停止看门狗：
touch .ci/runner-watchdog.stop
```

> 入口说明：`docker-compose.runner.persist.yml` 使用 `runner/persistent-entrypoint.sh`，如检测到已有 `.runner/.credentials` 或 `.credentials*` 将跳过 `config.sh`，避免 “already configured” 重启；需要重配时设置 `FORCE_RECONFIGURE=true`。

## 三、验证
- 仓库 Settings → Actions → Runners，应看到在线 Runner（labels: self-hosted,cubecastle,linux,x64,docker）
- 触发 smoke：Actions → “CI (Self-Hosted Runner Smoke)” → Run workflow

## 四、常见问题
- Runner 未注册成功：检查 `GH_RUNNER_REG_TOKEN` 是否过期，或 PAT 是否具备 `repo/workflow`；查看容器日志。
- 工作流未在 self-hosted 执行：检查 runs-on 标签是否匹配；Runner 在线状态是否正常。
- 安全：容器挂载了 `/var/run/docker.sock`，具高权限，务必在受信主机运行；必要时采用隔离宿主/虚拟机。

## 五、回滚
```bash
docker compose -f docker-compose.runner.yml down -v
```
删除 Runners 页面的记录；归档 smoke 工作流（如不再需要）。

---

## 六、WSL Runner（Plan 269 正式通道）

> 2025-11-20 起，Plan 269 批准在满足“业务服务依旧运行在 Docker Compose 中”的前提下，将 WSL 发行版（Ubuntu 20.04+/22.04）中的原生 Runner 作为唯一的自托管通道。标签固定为 `self-hosted,cubecastle,wsl`；Docker Runner 已退役，如需恢复必须重新走计划审批。

1. **安装 / 升级**
   ```bash
   # 需要在 WSL 命令行中执行，确保已满足 Go>=1.24.9、Node>=18、Docker CLI 可访问宿主 Docker Daemon
   bash scripts/ci/runner/wsl-install.sh \
     --repo jacksonlee411/cube-castle \
     --runner-dir "$HOME/actions-runner" \
     --labels self-hosted,cubecastle,wsl
   ```
   - 脚本会自动加载 `secrets/.env.local` 中的 `GH_RUNNER_PAT`/`GH_RUNNER_REG_TOKEN`，并在 `logs/wsl-runner/install-*.log` 落盘日志。
   - 若需重新注册，使用 `--force-reconfigure`（需 PAT）；若 WSL 已启用 systemd，可附加 `--use-systemd`，否则默认使用 `tmux cc-runner` 守护。
2. **校验**
   ```bash
   bash scripts/ci/runner/wsl-verify.sh
   ```
   - 检查 Go/Node/Docker 版本、`docker context`、Runner 目录状态、tmux session，并调用 `scripts/network/verify-github-connectivity.sh --smoke`。
   - 若 `GH_RUNNER_PAT` 可用，会额外查询 GitHub API，确认 `${RUNNER_NAME}` 的在线状态。
3. **卸载 / 回滚**
   ```bash
   bash scripts/ci/runner/wsl-uninstall.sh --repo jacksonlee411/cube-castle
   ```
   - 自动停止 tmux/systemd、调用 `./config.sh remove --token ...` 并将目录备份到 `<dir>.bak-<timestamp>`。
   - 若 WSL Runner 故障，应在 30 分钟内完成停机、修复或替换节点，并在 Plan 265/266/269 登记 run ID 与日志；如需重新启用 Docker Runner，须先走计划审批。
4. **文档与守护**
   - 完整步骤详见 `docs/reference/wsl-runner-setup.md`、对比见 `docs/reference/wsl-runner-comparison.md`（已更新为“WSL=默认、Docker=历史”）。
   - 网络诊断与 hosts/代理指南：`docs/reference/docker-network-playbook.md`；watchdog 可调用 `scripts/ci/runner/wsl-verify.sh` + `scripts/network/verify-github-connectivity.sh`。
   - 所有 workflow self-hosted job 均固定 `runs-on: [self-hosted,cubecastle,wsl]`（通过 matrix `wsl_only` 控制运行时机），请把每次成功 run 的 Run ID 写入 Plan 265/266。
