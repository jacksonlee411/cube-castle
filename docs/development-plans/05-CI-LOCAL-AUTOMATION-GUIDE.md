# 05 — 自托管 Runner 使用手册（Docker Compose 管控）

版本: v2.0  
最后更新: 2025-11-17  
适用范围: GitHub Actions 自托管 Runner 的部署/运维/故障排除  
唯一事实来源: Plan 262（docs/development-plans/262-self-hosted-runner.md）、`docker-compose.runner*.yml`、`scripts/ci/runner/*`

---

> 约束提醒（AGENTS.md）
> - 所有 Runner 也属于“服务/中间件”，必须由 Docker Compose 管理，禁止在宿主直接安装或裸 `docker run`；如端口冲突必须卸载宿主服务，不能改 compose 端口。
> - 密钥统一存放 `secrets/`（.gitignore 已忽略），严禁写入版本库；临时方案需 `// TODO-TEMPORARY(YYYY-MM-DD)` 并在 215 登记。

## 1. 目标
- 让任何开发者按本手册即可部署/维护自托管 Runner，确保运行方式与 Plan 262 定义一致；
- 收敛 Runner 生命周期（启动/验证/回滚）到 Docker Compose；
- 记录诊断与常见问题，支撑 CI 迁移与平滑回滚。

## 2. 权威索引
- 方案与状态：`docs/development-plans/262-self-hosted-runner.md`
- Compose 描述：
  - `docker-compose.runner.yml`（Ephemeral）
  - `docker-compose.runner.persist.yml`（持久化，需按 Plan 262 修正 command）
- 脚本工具：`scripts/ci/runner/README.md`、`scripts/ci/runner/start-ghcr-runner-persistent.sh`、`scripts/ci/runner/watchdog.sh`
- 工作流：`.github/workflows/ci-selfhosted-diagnose.yml`、`.github/workflows/ci-selfhosted-smoke.yml`
- 日志目录：`logs/ci-monitor/`（watchdog、工作流 run 摘要）

## 3. 准备条件
1. **宿主要求**：Docker ≥24、Docker Compose v2（`docker compose version`），端口 9090/8080/5432/6379 无宿主冲突。
2. **凭证**（二选一）：
   - Registration Token：仓库 → Settings → Actions → Runners → New self-hosted runner → Linux；写入 `secrets/.env.local`：
     ```bash
     GH_RUNNER_REG_TOKEN=<token>
     ```
   - PAT：scope 至少 `repo`,`workflow`；写入：
     ```bash
     GH_RUNNER_PAT=<pat>
     ```
3. **目录**：在仓库根执行，确保 `secrets/` 可读写（该目录已被忽略）；日志写入 `logs/ci-monitor/`。

## 4. 启动流程
### 4.1 Ephemeral（推荐默认）
> 每个 Job 结束自动注销，适合高安全场景。

```bash
# 读取令牌（env 或 secrets/.env.local）
source secrets/.env.local

# 启动一次性 Runner
docker compose -f docker-compose.runner.yml up -d

# 查看日志
docker compose -f docker-compose.runner.yml logs -f
```

Job 结束后容器自动退出，可 `docker compose ... down -v` 清理。若需重新注册，重复执行上方命令即可。

### 4.2 持久化（Plan 262 当前运行方式）
> 适合需要常驻 Runner、缓存依赖/镜像层的场景；必须结合看门狗。

```bash
# 建议使用脚本封装注册逻辑
bash scripts/ci/runner/start-ghcr-runner-persistent.sh

# 启动看门狗（默认 60s 轮询）
nohup bash scripts/ci/runner/watchdog.sh 60 > logs/ci-monitor/watchdog.out 2>&1 &
```

> ⚠️ 若改为 Compose 承载持久化 Runner，请同时修正 `docker-compose.runner.persist.yml` 的 command：首次启动执行 `./config.sh ... && ./run.sh`，后续复用已有 `.runner` 目录只执行 `./run.sh`，避免 “already configured” 无限重启（Plan 262 当前风险）。

### 4.3 停止与回滚
```bash
# Ephemeral
docker compose -f docker-compose.runner.yml down -v

# 持久化
docker compose -f docker-compose.runner.persist.yml down -v   # 或 docker rm -f cubecastle-gh-runner
touch .ci/runner-watchdog.stop                               # 让看门狗退出
```

同时到 GitHub → Settings → Actions → Runners 删除离线实例，防止残留。

## 5. 验收与诊断
1. **Runner 在线**：Settings → Actions → Runners 列表中应显示标签 `self-hosted,cubecastle,linux,x64,docker`。
2. **诊断工作流**：手动触发 `.github/workflows/ci-selfhosted-diagnose.yml`，需在自托管 Runner 上成功运行，日志应包含 `docker version`、`docker compose -f docker-compose.dev.yml config -q` 等输出（Plan 262 要求）。
3. **Compose 工作负载验证**：触发 `.github/workflows/ci-selfhosted-smoke.yml`，确认 Runner 能拉起 Compose 服务并通过健康检查；将 run 链接登记到 `docs/development-plans/215-phase2-execution-log.md`。
4. **监控日志**：`logs/ci-monitor/runner-watchdog-*.log`、`logs/ci-monitor/run-*.txt` 用于回溯。

## 6. 常见问题排查
| 现象 | 原因 | 处理 |
| --- | --- | --- |
| 容器反复 Restarting，日志显示 `Cannot configure the runner because it is already configured` | `docker-compose.runner.persist.yml` 每次都执行 `./config.sh --replace`，已有 `.runner` 导致冲突 | 启动前 `./config.sh remove`，或按 Plan 262 建议在 entrypoint 里检测 `.runner` 后只执行 `./run.sh` |
| 工作流仍跑在 `ubuntu-latest` | 工作流 `runs-on` 未包含 `self-hosted,cubecastle,docker` 标签 | 修改目标工作流，或在 PR 中添加 matrix `os: [ubuntu-latest, self-hosted]` |
| 诊断 job 卡在 `docker compose ... config -q` | 工作流未 checkout 仓库导致 compose 文件不存在 | 在 job 中补 `actions/checkout@v4`，或确保命令运行目录包含 compose 文件（Plan 262 已记录该问题） |
| 令牌过期 | Registration Token 仅 1 小时有效 | 重新申请 token 并更新 `secrets/.env.local`；若使用 PAT，确认未过期且 scope 正确 |
| 看门狗无法停止 | 未创建 stop 文件 | `touch .ci/runner-watchdog.stop`；确认 watchdog 脚本读取该文件后退出 |

## 7. 运行与维护建议
- **日志与证据**：所有启动/停止/诊断命令输出应落盘 `logs/ci-monitor/`，便于 215 执行日志引用。
- **镜像升级**：Runner 镜像固定在 `ghcr.io/actions/actions-runner:<version>`；升级时应在 staging 验证，随后更新 compose 与脚本。
- **工作流迁移**：Plan 262 建议为重型工作流增加矩阵验证。更新流程：
  1. 在目标 workflow 中添加 `runs-on: [self-hosted, cubecastle, docker]` 或 matrix；
  2. 首轮手动触发并记录 run 链接；
  3. 绿灯后更新 215 与相关计划文档。
- **安全**：Runner 容器挂载 `/var/run/docker.sock`，等价于宿主 root；仅在受信主机部署，必要时隔离网络或启用 Ephemeral。

## 8. 回滚策略
- 发现 Runner 异常或违反 AGENTS 约束时，立即执行：
  1. `docker compose -f docker-compose.runner*.yml down -v` 清理容器；
  2. `touch .ci/runner-watchdog.stop` 停止看门狗；
  3. 删除 GitHub Runner 条目；
  4. 在 215 登记事件与处理过程；
  5. 恢复使用 GitHub 托管 Runner（无需额外操作）。

## 9. 未来工作（按 Plan 262）
- 将持久化 Runner 完全迁移到 Compose 管控，并修正 command 引发的重复配置；
- 为重型工作流添加 `matrix.os=[ubuntu-latest,self-hosted]`，逐步把 Compose/E2E 工作流切换到自托管节点；
- 完成 smoke/diagnose run 的登记与 Required check 配置，确保 Runner 状态可观测。

---

维护：基础设施与架构组（与 Plan 262 保持同步）  
偏差处理：若本手册与 Plan 262、scripts/ci/runner/* 不一致，以后者为准并及时回修。

## 10. feat/shared-dev 推送经验（共享分支）
- 仓库规则要求所有非 `master` 工作必须通过共享分支 `feat/shared-dev` + PR；直接 push 受保护。
- 若需要临时允许 push（例如 CI 修复、合并验证），需由管理员在 Ruleset Bypass 中添加账号（例：`jacksonlee411`），推送完成后务必恢复规则。
- 推送前本地 pre-push hook 会运行 Plan 255 守卫；若存在已知 typecheck 噪音（`Date.Format` 等），可暂用 `HUSKY=0 git push ...` 触发，随后在 CI 确认状态。
- 很多 workflow 受限于 `branches: [main, master]`，若要在 `feat/shared-dev` 触发，需更新对应 workflow（如本次修复将 `feat/shared-dev` 加入 plan-250/254/255 等门禁）。
- PR 合并需满足“7 of 7 required status checks”；在 CI 全绿前禁止合并。若某 Required check（如 Agents Compliance）引用老 run，需要通过空提交或 PR UI rerun 触发 **最新** run。
