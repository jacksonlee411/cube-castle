# Plan 267 - Docker 网络访问稳定化与代理兼容方案

**文档编号**: 267  
**创建日期**: 2025-11-20  
**关联计划**: Plan 262/265（自托管 Runner）、Plan 266（运行追踪）

---

## 1. 背景与目标

- 当前自托管 Runner 运行在 Windows + WSL2 (`networkingMode=mirrored`, `dnsTunneling=true`) 环境。根据 Plan 266 采集的日志（`logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`, `...074425Z.log`），`github.com` 被劫持到 `11.2.0.12`，TLS 握手后服务器立即断开，导致 `git ls-remote`、`curl` 均报 `gnutls_handshake()`/`OpenSSL SSL_read unexpected eof`。即便在 Compose 中显式指定 DNS，依旧受宿主网络栈控制。
- 2025-11-20 补充验证：宿主 Windows 直接访问 `https://github.com` 返回 200（`Invoke-WebRequest`），但在 WSL 内 `curl -v https://github.com` 显示 DNS 解析到 `11.2.0.12` 并在 TCP 阶段 15 秒超时；`getent hosts` 同样返回 `11.2.0.12`，而 Windows `nslookup` 返回官方 IP `20.205.243.166`。说明网关/透明代理仅对 WSL/Docker 路径生效，管理员“无限制”的反馈与事实不符。
- 该问题阻塞 `document-sync`、`api-compliance`、`consistency-guard` 等所有 self-hosted job，Plan 265 无法进入 Required 状态。需要一份系统性方案，既涵盖“请求网络团队放通/提供代理”的对外动作，也包含本地自动诊断、配置模板与回滚策略。
- Plan 267 的目标：建立 Docker/WSL 网络访问的唯一事实来源，提供诊断 → 治疗 → 监控的闭环，确保自托管 Runner 在受限网络下仍能稳定访问 GitHub 及 Docker Registry。

---


## 📌 搁置结论

- 由于网络放通与代理方案尚未落地，WSL Runner 依旧依赖临时 hosts 覆盖。Plan 267 中的脚本/监控/代理任务暂缓执行，待网络团队提供稳定方案后再恢复。
- 目前仍沿用 Plan 267-D 的静态 hosts 方案，其他动作记录于 Plan 266/269。

## 2. 范围与交付物

| 范畴 | 交付物 | 说明 |
|------|--------|------|
| 诊断脚本 | `scripts/network/verify-github-connectivity.sh` | 收集 `getent hosts`, `curl -v`, `openssl s_client`, `docker logs`，日志落盘 `logs/ci-monitor/network-*.log`，供 Plan 266/265 引用 |
| WSL/Compose 配置模板 | `docs/reference/docker-network-playbook.md` | 说明 `networkingMode=mirrored` 的影响、何时需要显式 Proxy、如何配置 `HTTP(S)_PROXY`、`NO_PROXY`、`WSLENV`、`/etc/hosts` |
| 企业代理支持 | `docker-compose.runner.persist.yml` 扩展 | 支持在 `secrets/.env.local` 中定义 `RUNNER_HTTP_PROXY`/`RUNNER_HTTPS_PROXY`，自动挂载企业 CA 证书，Runner 入口脚本内调用 `update-ca-certificates` |
| 放通流程 | 与网络团队协同记录 | 在 `docs/development-plans/267-*.md` 中保存申请模板（需要放通的域名/IP、端口、原因、回退方案） |
| 监控与报警 | `scripts/ci/monitor-runner-network.sh` | Crontab/Watchdog 插件，每小时执行 `verify-github-connectivity.sh --smoke`，失败时写入 `logs/ci-monitor/network-watchdog.log` 并触发 Slack/Webhook（hook 地址引用 `secrets/` 配置） |

不在本次范围内：修改 AGENTS.md 的原则、替换 Runner 宿主机、引入新的 VPN/WARP 软件（如后续需要，另起计划）。

---

## 3. 实施步骤

### 3.1 现状固化与证据整理
1. **整理日志**：将 Plan 266 收集的两个 TLS 日志压缩为 `logs/ci-monitor/network-proof-20251120.zip`，并在 README 中附说明。
2. **编写诊断脚本 V0**：复制 Plan 266 的命令（`docker compose ... ps/getent/curl/openssl`），接受 `--timeout`/`--output` 参数，支持宿主/容器双向运行。
3. **更新文档**：在 `docs/development-plans/266-selfhosted-tracking.md` 中引用 Plan 267，声明网络问题转入本计划处理。

### 3.2 方案 D：静态 hosts 覆盖（已执行，短期守护）
1. **WSL 固化**：已在 `/etc/wsl.conf` 追加 `[network]\ngenerateHosts=false` 并执行 `wsl.exe --shutdown`（2025-11-20 晚间）确保 WSL 不再重写 `/etc/hosts`，后续若更新同一段配置，需在写入后再次 `wsl.exe --shutdown` 让静态 hosts 立即生效。
2. **Hosts 维护脚本**：`scripts/network/configure-github-hosts.sh` 提供 `sudo bash scripts/network/configure-github-hosts.sh` 一键写入 `github.com`、`codeload.github.com`、`api.github.com`、`raw.githubusercontent.com`、`github-releases.githubusercontent.com`、`release-assets.githubusercontent.com`、`objects.githubusercontent.com`、`objects-origin.githubusercontent.com` 等 Release/Artifact 域名（2025-11-19 采用 `getent ahostsv4` 采集的官方 IP），并附带 `# Plan 267-D GitHub override` 标记。脚本会在覆盖前生成带时间戳的备份（示例 `/etc/hosts.plan267.20251120T230000.bak`），幂等去重，方便回滚。
3. **Runner 验证**：新增 `scripts/network/apply-github-hosts-to-runner.sh`（内部先调用 `configure-github-hosts.sh`、随后 `docker compose exec gh-runner sudo bash /tmp/configure-github-hosts.sh`），确保宿主与 `cubecastle-gh-runner` 容器共用同一份 hosts 映射。执行 `getent hosts github.com`、`curl -I https://github.com`、`git ls-remote https://github.com/jacksonlee411/cube-castle`、`docker compose -f docker-compose.runner.persist.yml exec gh-runner curl -I https://github.com` 均应返回 200 且解析到脚本中写入的 IP。
4. **回退与监控**：如需恢复默认解析，可执行 `sudo cp /etc/hosts.plan267.<timestamp>.bak /etc/hosts && wsl.exe --shutdown`。Watchdog/诊断脚本中需记录最近一次备份文件名与验证命令输出，当 GitHub IP 调整或网络团队放通后，按上述步骤恢复并复跑 `configure-github-hosts.sh` 以注入新 IP。

### 3.3 方案 A：企业放通
1. **需求模板**：输出 `docs/reference/templates/network-whitelist-request.md`，列出需要放通的域名/IP（GitHub、Docker Hub、ghcr.io、actions.githubusercontent.com 等）、端口、用途、负责人、回滚方式（可切回代理）。
2. **提交流程**：协同网络团队（记录联系人、工单号），将 Plan 267 作为背景、Plan 265 的阻塞作为影响范围。
3. **验收**：放通后运行诊断脚本，确认 `getent hosts github.com` 不再指向 11.x.x.x，`curl -v` 可收到 200/301；在 Plan 265 中记录成功的 run ID。

### 3.4 方案 B：代理/证书方案
1. **配置入口**：在 `docker-compose.runner.persist.yml` 增加可选环境变量 `RUNNER_HTTP_PROXY`、`RUNNER_HTTPS_PROXY`、`RUNNER_NO_PROXY`，默认从 `secrets/.env.local` 读取；`runner/persistent-entrypoint.sh` 检测变量后写入 `/etc/profile.d/proxy.sh` 并更新 `git config --global http.proxy`。
2. **CA 证书**：允许将 `secrets/certs/*.crt` 挂载到 Runner 容器，通过 `update-ca-certificates` 注入系统信任；文档记录如何导出企业根证书。
3. **WSL 映射**：在 `docs/reference/docker-network-playbook.md` 中说明如何使用 `WSLENV=HTTPS_PROXY/up` 将 Windows 代理自动传入 WSL。
4. **回滚**：脚本提供 `--clear-proxy` 选项，移除 `proxy.sh` 并重启 Runner。

### 3.5 自动化守护
1. **Watchdog 扩展**：在现有 `runner/watchdog.sh` 中加入网络检查，失败时在日志中标记 `[NET] FAIL`；可选：发 Slack/邮件（调用 `scripts/ci/notify.sh`）。
2. **CI 集成**：在 `consistency-guard.yml` 增加一个前置 Step：`bash scripts/network/verify-github-connectivity.sh --smoke --fail-fast`，确保推送前即检测到网络问题。
3. **监控指标**：将 `verify-github-connectivity.sh` 输出转换为 Prometheus 可读指标（例如 `runner_github_handshake_success{phase="curl"}`），由 `monitoring/` 采集。

### 3.6 文档与交接
1. **Playbook 发布**：`docs/reference/docker-network-playbook.md` 包含背景、WSL 设置、代理配置、放通申请模板、常见错误（DNS 劫持、TLS reset、MTU 限制）。
2. **Plan 267 完成认定**：需要 Plan 265 的所有 Required workflow 至少一次通过自托管 runner、Plan 266 无新的阻塞项，且 Watchdog 连续 7 天未报网络故障。

---

## 4. 验收标准

- [ ] `scripts/network/verify-github-connectivity.sh` 在宿主与 Runner 容器均可执行，能输出 GitHub/TLS 诊断信息，并将日志写入 `logs/ci-monitor/`。
- [ ] `docker-compose.runner.persist.yml` 支持通过 `.env.local` 配置 HTTP(S) 代理与 CA 证书，Runner 启动脚本自动应用。
- [ ] `docs/reference/docker-network-playbook.md` 发布，包含 WSL 参数说明、代理配置手册、放通申请模板。
- [ ] 至少一次 `api-compliance (selfhosted)`、`document-sync (selfhosted)`、`consistency-guard (selfhosted)` 在启用上述方案后全部通过，并在 Plan 265/266 记录 run ID。
- [ ] Watchdog/CI 网络检查连续 7 天通过；若失败，能自动生成日志并通知负责人。

---

## 5. 风险与回滚

| 风险 | 描述 | 缓解/回滚 |
|------|------|-----------|
| 企业代理无法提供或延迟 | 可能在长时间内无法获取合法代理 | 提前提交放通申请；维持 `workflow_dispatch` + Ubuntu job 作为备选 |
| 导入 CA 证书失败 | Runner 信任链不完整导致所有 TLS 失败 | 在脚本中校验 `update-ca-certificates` 结果，遇到错误回滚到无代理模式 |
| Watchdog 误报 | 网络抖动导致频繁报警 | 支持 `--retries`、`--interval` 参数，连续失败才通知 |
| DNS 缓存污染 | WSL 宿主缓存旧 IP | 文档中要求运行 `ipconfig /flushdns` + `wsl --shutdown` 并记录 |

---

## 6. 参考资料

- AGENTS.md：Docker 强制要求、资源唯一性
- Plan 265/266：自托管 Runner 目标与历史 run 记录
- 日志：`logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`、`logs/ci-monitor/selfhosted-tls-20251119T074425Z.log`
- Windows WSL 网络设置：`~/.wslconfig` (`networkingMode=mirrored`, `dnsTunneling=true`)

Plan 267 完成后，将在 Plan 265/266 中关闭网络类风险，并将 Playbook 作为唯一事实来源供后续代理/网络变更使用。
