# 自托管 Runner（Docker Compose 管控：Ephemeral + 持久化）

本仓库遵循 AGENTS.md 的“Docker 强制”原则，自托管 Runner 必须由 Docker Compose 管理：  
- 默认：Ephemeral 一次性模式（`docker-compose.runner.yml`）  
- 持久化：常驻接单模式（`docker-compose.runner.persist.yml`，由 watcher 保活）

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

# 启动持久化（建议结合 watchdog 保活）
RunNER_TOKEN=... docker compose -f docker-compose.runner.persist.yml up -d
# 或使用自动申请令牌脚本（推荐）
bash scripts/ci/runner/start-ghcr-runner-persistent.sh
nohup bash scripts/ci/runner/watchdog.sh 60 > logs/ci-monitor/watchdog.out 2>&1 &

# 查看 Runner 容器日志
docker logs -f cubecastle-gh-runner

# 停止并清理
docker compose -f docker-compose.runner.yml down -v
# 或持久化：
docker compose -f docker-compose.runner.persist.yml down -v
# 停止看门狗：
touch .ci/runner-watchdog.stop
```

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
