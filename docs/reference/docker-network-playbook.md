# WSL / Docker 网络治理手册（Plan 267）

**最后更新**：2025-11-20  
**适用场景**：Windows + WSL2（`networkingMode=mirrored`）环境中运行 Docker Compose、GitHub Actions Runner、Playwright 等网络密集工作负载。

## 1. 问题背景

- 2025-11 起自托管 Runner 在 `github.com`、`ghcr.io`、`registry.npmjs.org` 访问时出现 DNS 劫持（解析到 `11.x.x.x` 内网 IP）与 TLS reset，导致 `document-sync`, `api-compliance`, `consistency-guard` 等 Workflow 长期阻塞。
- 由于历史 Docker Runner（容器内）→ WSL → Windows → 企业网络链路层层代理，故障难以定位。当前 Runner 已切换为 WSL 原生模式，但 Compose 工作负载仍运行在 WSL 中，需要统一的诊断与恢复路径，避免重复“关机重试”。

## 2. WSL 宿主配置

1. `~/.wslconfig`：
   ```ini
   [wsl2]
   networkingMode=mirrored
   dnsTunneling=true
   memory=12GB
   processors=6
   ```
   更新后运行 `wsl.exe --shutdown` 立即生效。

2. 固定 hosts：
   ```bash
   sudo bash scripts/network/configure-github-hosts.sh
   # 该脚本会备份 /etc/hosts，写入 Plan 267 推荐 IP 并标注 "# Plan 267-D GitHub override"
   wsl.exe --shutdown   # Windows PowerShell 中执行，确保 hosts 缓存刷新
   ```

3. 代理传递：
   - 在 Windows 环境变量中设置 `HTTPS_PROXY`/`HTTP_PROXY`，通过 `WSLENV=HTTPS_PROXY/up;HTTP_PROXY/up;NO_PROXY/up` 传入 WSL。
   - 或在 `scripts/ci/runner/wsl-install.sh`、`docker-compose.runner.persist.yml` 中读取 `secrets/.env.local`，自动导出代理。

4. DNS 清理：
   ```powershell
   ipconfig /flushdns
   netsh winsock reset
   wsl.exe --shutdown
   ```
   避免 Windows DNS 缓存保留旧的 GitHub IP。

## 3. WSL Runner / Compose 设置

| 项目 | 操作 |
|------|------|
| CA 证书 | 将企业根证书复制到 `secrets/certs/*.crt`，Runner 容器挂载后执行 `update-ca-certificates`。 |
| 代理 | 在 `secrets/.env.local` 中配置 `RUNNER_HTTP_PROXY`、`RUNNER_HTTPS_PROXY`、`RUNNER_NO_PROXY`，Compose 文件读取后注入容器环境变量，并设置 `git config --global http.proxy`。 |
| Hosts 同步 | 执行 `scripts/network/apply-github-hosts-to-runner.sh`，确保 WSL 内 `/etc/hosts` 与 Windows 保持一致。 |
| 网络诊断 | `bash scripts/network/verify-github-connectivity.sh --smoke --output logs/ci-monitor/network-$(date +%s).log`；脚本会检查 `getent hosts`, `curl -v`, `openssl s_client`, `docker login ghcr.io`。 |

## 4. 代理 / 放通流程

1. **申请模板**：
   ```
   申请人：<Name>（Plan 267）
   目标域名：github.com, api.github.com, codeload.github.com, ghcr.io, registry.npmjs.org, objects.githubusercontent.com
   协议/端口：TCP 443
   用途：GitHub Actions 自托管 Runner（Plan 265 Required Checks）
   回滚：记录脚本输出，必要时 `sudo cp /etc/hosts.plan267.<timestamp>.bak /etc/hosts` 并 `wsl.exe --shutdown` 恢复
   证据：logs/ci-monitor/network-proof-*.log
   ```
2. **临时代理**：如短期内无法放通，使用企业 HTTP(S) 代理：
   ```bash
   export HTTPS_PROXY=http://proxy.example.com:8080
   export NO_PROXY=localhost,127.0.0.1,.local,docker.internal
   docker login ghcr.io
   ```
3. **证书导入**：如果代理拦截 TLS，下载根证书并安装：
   ```bash
   sudo cp company-root-ca.crt /usr/local/share/ca-certificates/
   sudo update-ca-certificates
   ```

## 5. 诊断脚本

`scripts/network/verify-github-connectivity.sh` 已满足以下功能：

- 参数：
  - `--smoke`：快速检查（DNS + HTTPS HEAD）
  - `--full`：额外执行 `openssl s_client`, `git ls-remote`, `docker login --dry-run`
  - `--output <file>`：将所有命令输出写入同一日志
  - `--container <name>`：在指定容器内执行（默认宿主）
- 输出示例：
  ```
  [DNS] github.com -> 20.205.243.166
  [TLS] curl https://github.com OK (HTTP/2 200)
  [GIT] git ls-remote https://github.com/jacksonlee411/cube-castle ok
  ```
- Watchdog 集成：`scripts/ci/runner/watchdog.sh` 可每 10 分钟执行 `verify-github-connectivity.sh --smoke --output logs/ci-monitor/network-watchdog.log`，失败时写入 `[NET][FAIL]` 标记。

## 6. 与 WSL Runner 的关系

- WSL Runner 直接使用 WSL 网络栈，故 `scripts/network/configure-github-hosts.sh` 与本 Playbook 同样适用。
- `scripts/ci/runner/wsl-install.sh` 和 `wsl-verify.sh` 会在启动前调用 `verify-github-connectivity.sh --smoke`。若检测失败，脚本会拒绝启动 Runner。
- Docker Runner 已退役，如需恢复需重新立项审批；本 Playbook 默认所有诊断均在 WSL Runner 环境执行。

## 7. 回滚策略

1. **撤销 hosts 覆盖**
   ```bash
   sudo cp /etc/hosts.plan267.<timestamp>.bak /etc/hosts
   wsl.exe --shutdown
   ```
2. **移除代理**
   ```bash
   unset HTTP_PROXY HTTPS_PROXY NO_PROXY
   git config --global --unset http.proxy
   git config --global --unset https.proxy
   ```
3. 在 Plan 265/266 中记录操作时间、原因、影响范围；如未来需要重新启用 Docker Runner，需单独提交计划并按照审批结果执行。

## 8. 附录

- **日志目录**：`logs/ci-monitor/`（网络证明）、`logs/wsl-runner/`（WSL 相关）
- **参考计划**：Plan 262（Runner 基建）、Plan 265（Required Checks）、Plan 266（运行追踪）、Plan 269（WSL Runner）
- **工具链**：`curl`, `openssl`, `git`, `docker`, `powershell`（Windows）、`tmux`, `systemd`.
