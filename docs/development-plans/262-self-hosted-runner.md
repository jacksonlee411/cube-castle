# Plan 262 — GitHub 自托管 Runner（Docker Ephemeral）启用方案

文档编号: 262  
负责人: Codex（AI）  
目标: 在不违背 AGENTS.md 的“Docker 强制 + 资源唯一性”前提下，引入“容器化自托管 Runner”，优先采用“持久化（persistent）+ 看门狗”方案，确保“永久在线、随时接单”；后续可切换/补充 Ephemeral 以降低脏环境风险。

---

## 0. 当前执行状态（2025-11-17）
- ✅ 基础产物补齐：`docker-compose.runner.yml`（Ephemeral）、`docker-compose.runner.persist.yml`（持久化，内置幂等入口 `runner/persistent-entrypoint.sh` 避免重复 config）、`.github/workflows/ci-selfhosted-{diagnose,smoke}.yml` 以及 `scripts/ci/runner/*`。
- ✅ 实际运行环境已存在持久化 Runner：命令 `scripts/ci/runner/start-ghcr-runner-persistent.sh` + `scripts/ci/runner/watchdog.sh` 持续拉起 `cubecastle-gh-runner`（证据：`logs/ci-monitor/runner-watchdog-20251117_023435.log`、`runner-watchdog-20251117_065943.log`，可见容器多次重启后仍在线）。
- ✅ 诊断工作流已在自托管节点跑通一次（`logs/ci-monitor/run-19414150520.zip/0_Self-Hosted Runner Diagnose.txt`），输出 Docker/Docker Compose 版本。
- ⚠️ 运行方式尚待迁移：目前线上仍是 `docker run` + 看门狗，需切换为 `docker-compose.runner.persist.yml` 统一管控。
- ⚠️ 诊断/冒烟待实跑验收：`ci-selfhosted-diagnose`/`ci-selfhosted-smoke` 已补全（含 checkout 与 compose 校验），但尚无最新成功 run 记录；时间表 T+0/1 未验收。
  - ⚠️ 尚未对任何重型工作流引入 `matrix.os=[ubuntu-latest,self-hosted]`，T+1/2 任务未执行。
- ⚠️ `cubecastle-gh-runner` 容器曾因重复 config 导致 Restarting，已在 `runner/persistent-entrypoint.sh` 幂等化处理；迁移到 Compose 后需以该入口替换旧启动方式。

## 1. 背景与动机
- 现状：GitHub 托管 Runner 在高峰期存在排队/资源波动；部分工作流（E2E/Compose）对底层环境更敏感。
- 目标：提供“与本地一致的 Docker 环境”，减少资源噪音；允许跑 Compose、缓存镜像层、缩短端到端时延。
- 约束：严格遵循 AGENTS.md——所有服务（含 Runner）以 Docker Compose 管理；密钥存放 secrets/，不入库；临时措施需标注并可回收。

## 2. 方案概述
- 架构（已实施：方案A Persistent）：基于 GHCR 官方镜像 `ghcr.io/actions/actions-runner:2.329.0` 启动“持久化 Runner”，通过 `scripts/ci/runner/start-ghcr-runner-persistent.sh` 注册并常驻，结合 `scripts/ci/runner/watchdog.sh` 每 60s 健康巡检，异常自动重拉。
- 备选（方案B Ephemeral）：基于 `runner/Dockerfile.docker` 自建镜像或 myoung34/github-runner 的 ephemeral 模式，每个 Job 结束自动注销 Runner，降低脏环境风险（可在稳定后切换或双轨）。
- 注册方式（两选一）：
  1) 仓库 Registration Token（脚本自动申请，一次性）；
  2) PAT（需 repo + workflow 范围；写入 secrets/.env.local，脚本自动读取）。
- 标签：self-hosted,cubecastle,linux,x64,docker（工作流通过 runs-on 指定）。
- 生命周期：方案A 持久化在线；方案B Job 结束销毁；均由 Docker 统一管控。

## 3. 实施步骤
1) 准备密钥（任一）
   - Registration Token：GitHub → 本仓库 → Settings → Actions → Runners → New self-hosted runner（Linux），复制 token；
   - 或 PAT：建议只勾选 repo、workflow，写入 secrets/.env.local（被 .gitignore 忽略）：
     ```
     GH_RUNNER_PAT=ghp_xxx
     # 或者一次性注册用
     GH_RUNNER_REG_TOKEN=xxxx
   ```
2) 启动 Runner（容器化，方案A）
   - **目标实现**：通过 `docker compose -f docker-compose.runner.yml up -d` 启动 Ephemeral，或 `docker compose -f docker-compose.runner.persist.yml up -d` 启动持久化 Runner，确保“容器统一受 Compose 管控”。
   - **当前实际**：以 `bash scripts/ci/runner/start-ghcr-runner-persistent.sh` 直接 `docker run`，随后用 `scripts/ci/runner/watchdog.sh` 保活 ⇒ 需在 T+0/1 内迁移至 Compose（已提供幂等入口 `runner/persistent-entrypoint.sh`）。
   - **新增诊断**：`docker-compose.runner.persist.yml` 入口已更新为检查 `.runner/.credentials` 与遗留 `.credentials*` 后再决定是否执行 `config.sh`，避免 “already configured” 的 Restarting；若需要重配可设置 `FORCE_RECONFIGURE=true`。
   - 验证在线后启动看门狗：`nohup bash scripts/ci/runner/watchdog.sh 60 > logs/ci-monitor/watchdog.out 2>&1 &`
   - 停止：`docker rm -f cubecastle-gh-runner`；停止看门狗：`touch .ci/runner-watchdog.stop`
3) 验证在线
   - GitHub → Settings → Actions → Runners 应看到在线 Runner（labels 包含 cubecastle、docker）
4) 验证工作流（示例）
   - 手动触发 `.github/workflows/ci-selfhosted-diagnose.yml`；应在自托管 Runner 执行并通过基本检查（docker/compose/环境），已包含 checkout 与 Compose config -q。
   - 手动触发 `.github/workflows/ci-selfhosted-smoke.yml`，记录成功 runId 及日志链接。

## 4. 工作流使用方式
- 试点工作流：`.github/workflows/ci-selfhosted-smoke.yml`（runs-on: [self-hosted, cubecastle, docker]）。
- 渐进迁移：为重型工作流增加 matrix.os = [ubuntu-latest, self-hosted] 的验证 Job；稳定后再切主路径到 self-hosted。

## 5. 安全与合规
- 密钥与令牌：存放 secrets/（.gitignore 已忽略）；禁止提交至仓库。
- 权限最小化：PAT 仅 repo + workflow；Runner 镜像版本固定（禁自动更新），按需手动升级。
- 容器权限：Runner 容器挂载 /var/run/docker.sock（具高权限）；务必在受信主机运行；仅项目成员可访问主机。
- 日志与回收：Ephemeral Runner 执行完自动注销；Compose 控制容器生命周期；异常时手动 down 清理。

## 6. 验收标准
- Runner 在线：Runners 页面可见，标签与配置正确；
- 诊断任务通过：`CI (Self-Hosted Diagnose)` 绿；**并能在日志中看到 `docker compose -f docker-compose.dev.yml config -q` 成功输出**；
- 能运行 Docker Compose 工作负载（compose --wait + healthcheck），`CI (Self-Hosted Runner Smoke)` 至少一次成功运行且日志归档；
- 不引入端口映射冲突（不更改现有服务端口；仅 Runner 自身，无对外端口）。

## 7. 回滚方案
- `docker compose -f docker-compose.runner.yml down -v` 停止并清理容器；
- GitHub Runners 页面删除离线 Runner 记录；
- 删除或归档 `.github/workflows/ci-selfhosted-smoke.yml`。

## 8. 风险与缓解
- 风险：自托管 Runner 有权限较高；主机稳定性影响流水线。
  - 缓解：仅受信主机部署；Runner 使用 Ephemeral；密钥最小化；日志溯源。
- 风险：镜像版本漂移导致行为变化。
  - 缓解：固定镜像 tag；更新走小步验证。

## 9. 时间表与产物
- T0：落库 compose 与 smoke 工作流（已完成）；
- T+0/1：本地或 CI 主机启动 Runner，验证 smoke；**行动项**：补充 Compose 化持久化 Runner、触发 smoke 并落盘 runId；
- T+1/2：选择 1~2 个重型工作流做 matrix 试跑；**行动项**：优先挑选 `.github/workflows/fronted-quality-gate.yml`、`e2e-tests.yml` 等重型流水线添加自托管矩阵；
- 产物：`docker-compose.runner.yml`、`.github/workflows/ci-selfhosted-smoke.yml`、`scripts/ci/runner/README.md`。

---

附：关键文件与配置（已落库）
- docker-compose.runner.yml（容器化 Runner；Ephemeral 模式；固定镜像版本）
- .github/workflows/ci-selfhosted-smoke.yml（自托管 Runner 烟测）
- scripts/ci/runner/README.md（使用说明与安全提示）
