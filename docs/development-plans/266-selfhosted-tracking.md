# Plan 266 - 自托管 Runner 执行追踪与问题闭环

**文档编号**: 266  
**创建日期**: 2025-11-19  
**关联计划**: Plan 262（Runner 基建）、Plan 265（自托管门禁扩展）  
**状态**: ⚠️ 搁置（2025-11-20）——WSL Runner 网络与运行稳定性未达标，相关自托管验证暂停，待 Plan 267/269 网络方案成熟后再恢复。

---


## 📌 搁置结论

- 由于自托管 workflow（document-sync/api-compliance/consistency-guard）在 WSL Runner 上仍缺乏成功 run，Plan 266 的追踪任务暂时冻结，仅保留 `ci-selfhosted-smoke` 结果作为基线。
- 后续待 Runner 网络恢复后再继续补齐 run ID 并恢复追踪。

## 1. 最新进展

1. `cmd/hrms-server/command/main.go`、`internal/organization/handler/devtools.go`、`tests/e2e/auth_flow_e2e_test.go` 已移除 `http://localhost` 等硬编码，引入 `COMMAND_ALLOWED_ORIGINS`、`COMMAND_BASE_HOST/SCHEME`、`DEVTOOLS_COMMAND_BASE_URL`、`COMMAND_BASE_URL/QUERY_BASE_URL` 等环境变量（commit `b3aff300`）。
2. `api-compliance.yml`、`iig-guardian.yml` 的 `actions/checkout` 均提前至 `paths-filter` 之前，避免自托管 Job 在未拉取仓库时运行 filter 导致 `fatal: not a git repository`。
3. `document-sync.yml` 已允许在 `workflow_dispatch` 场景下强制运行全量步骤（不再被 docs-only fast pass 直接返回），Ubuntu 矩阵 job 已成功生成报告（run `19489933087`）。
4. 2025-11-20：新增 `internal/config/cors.go` + Query CORS 外部化、BFF 统一注入 `internal/config.GetJWTConfig()`；`ENFORCE=1 scripts/ci/check-hardcoded-configs.sh` 已返回 issues=0。
5. 2025-11-20 07:28Z：按照 3.1 步骤在 `cubecastle-gh-runner` 容器采集 TLS 证据，日志位于 `logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`（含 `git ls-remote`/`curl`/`openssl s_client` 输出）。
6. 2025-11-20 晚间：根据 Plan 267-D 执行 `/etc/wsl.conf` `[network]\ngenerateHosts=false` + `sudo bash scripts/network/configure-github-hosts.sh`，宿主与 Runner 侧 `getent hosts github.com`、`curl -I https://github.com`、`git ls-remote https://github.com/jacksonlee411/cube-castle` 均返回 200 并解析到官方 IP（详见 `docs/development-plans/267-docker-network-stability.md:39-43`），`/etc/hosts.plan267.<timestamp>.bak` 记录回滚点。
7. 2025-11-20 10:46Z：首版 `scripts/network/verify-github-connectivity.sh`（Plan 266/267 诊断脚本）在宿主与 `gh-runner` 容器内依次执行 `getent hosts github.com`、带浏览器 UA 的 `curl -sS -D - https://github.com`、`GIT_CURL_VERBOSE=1 git ls-remote https://github.com/jacksonlee411/cube-castle`、`openssl s_client -connect github.com:443`，日志落盘 `logs/ci-monitor/network-20251119T104614Z.log`，所有命令均返回 200/TLS OK，证明 hosts 覆盖后 Runner 同样可以建立 TLS。
8. 2025-11-20 12:47Z / 12:57Z：依次触发自托管版 `document-sync`（run `19502442553`、`19502825153`）。节点缓存已在 Runner `_work/_tool/node/18.20.8/x64` 预热，Job 能走到 “文档同步一致性检查”。但上传工件时复用 `document-sync-report-${{ github.run_number }}` 导致 409 conflict（相同 run 上多次 attempt），GitHub 将自托管 job 标记为取消；workflow 已改为 `document-sync-report-${{ github.run_number }}-${{ github.run_attempt }}`，仍待推送后验证。Run 日志同时记录 checkout 阶段偶发 `GnuTLS recv error (-110)`，需继续跟踪网络稳定性。
9. 2025-11-20 15:10Z：Plan 269 批准 WSL Runner 作为自托管备选，脚本 `scripts/ci/runner/wsl-install.sh`/`wsl-uninstall.sh`/`wsl-verify.sh` 与文档 `docs/reference/wsl-runner-setup.md`、`docs/reference/wsl-runner-comparison.md` 已落库；所有自托管 workflow 的 `runs-on` 标签同步新增 `wsl`，Plan 265/266 需记录首次 WSL run 的 Run ID 与日志。
10. 2025-11-20 07:11Z：在 WSL 主机 `DESKTOP-S9U9E9K` 通过 `bash scripts/ci/runner/wsl-install.sh` 重新拉起 `cc-runner`（日志：`logs/wsl-runner/install-20251120T071110.log`、`run-20251120T071113.log`），`gh api repos/jacksonlee411/cube-castle/actions/runners` 已显示 `wsl-DESKTOP-S9U9E9K` 在线；`logs/wsl-runner/network-smoke-20251120T071157.log` 与 `network-smoke-20251120T071451.log` 记录宿主探测 OK 但 `docker-compose.runner` 内 `curl` 依旧 56/timeout，证明 WSL Runner 仍是当前可用通道。
11. 2025-11-20 07:16Z：`workflow_dispatch` 触发 `document-sync`（run `19519517913`）尝试记录首个 `[self-hosted,cubecastle,wsl]` run，结果因远端 `.github/workflows/document-sync.yml` 尚未合入 `selfhosted-wsl` 矩阵，GitHub 仅调度 `cc-runner-docker-compose` 并在“质量门禁”阶段失败（日志见 `https://github.com/jacksonlee411/cube-castle/actions/runs/19519517913`）。需尽快推送 workflow 变更后再复测。
12. 2025-11-20 07:42Z：使用 `gh workflow run ci-selfhosted-smoke.yml --ref feat/shared-dev` 触发 run `19520064684`。WSL job（`Smoke (wsl)`）在 runner `wsl-DESKTOP-S9U9E9K` 上 2m26s 完成、日志已落盘 `logs/wsl-runner/ci-selfhosted-smoke-wsl-19520064684.log`；docker job 因 `docker compose` 健康检查退出码 125/2 失败，整体结论=失败。Plan 265/269 可以引用该 run 作为首个 WSL 作业记录，同时需继续排查 docker runner Compose 报错。

## 2. 遇到的问题 / 风险

| 问题 | 描述 | 当前影响 | 负责人/协作 |
|------|------|----------|-------------|
| 自托管 checkout TLS 断线 | `document-sync` Self-hosted job 在 `actions/checkout` 阶段多次出现 `gnutls_handshake()` / `curl 56`，无法从 GitHub 拉代码；2025-11-20 07:29Z 现场复现（log `logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`）显示连接被 11.2.0.12 截断、`openssl s_client` 无法拿到证书；Plan 267-D 通过静态 hosts 临时恢复了宿主/Runner 的 `curl`/`git ls-remote` | 目前依赖 Plan 267-D（WSL `generateHosts=false` + `scripts/network/configure-github-hosts.sh`）维持访问，若 hosts 再次被覆盖自托管 job 仍会失败，Plan 265 仍缺少成功 run（`19489933087` selfhosted） | DevInfra（排查 runner 网络/TLS，Plan 267 负责网络方案） |
| `api-compliance` run 长时间排队 | workflow_dispatch `19491103285` 仍 queued，自托管修复尚未验收 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |
| `iig-guardian` run 未执行 | workflow_dispatch `19491533343` queued，同上 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |
| Artifact 命名冲突导致 selfhosted 失效 | 自托管 `document-sync` 运行多次尝试（run `19502442553`、`19502825153`）时，`actions/upload-artifact@v4` 使用固定名称 `document-sync-report-${{ github.run_number }}`，GitHub 不允许在同一 run 中重复创建同名工件，于是上传返回 409、Job 被标记为 “The operation was canceled”。 | 自托管 run 无法进入清理/后续步骤；Plan 265 仍缺少成功 run。workflow 已改为 `document-sync-report-${{ github.run_number }}-${{ github.run_attempt }}`，需推送并重跑验证 | 平台组（更新 workflow、确认 artifact 命名不会冲突） |
| WSL Runner 运行记录 | Plan 269 获批后，已通过 `ci-selfhosted-smoke` run `19520064684` 拿到首个 `[self-hosted,cubecastle,wsl]` 成功 job（日志：`logs/wsl-runner/ci-selfhosted-smoke-wsl-19520064684.log`），但 `document-sync` / `api-compliance` / `consistency-guard` 仍缺乏 WSL 运行证据。 | Branch Protection 无法把文档 / 契约守卫切到 WSL；Plan 265/269 验收尚未完成；如果 WSL runner 故障又无更多 run 记录，将缺少回滚依据。 | DevInfra + 平台组：继续触发其他 workflow 的 `workflow_dispatch`，若 GitHub API 204 但未生成 run（目前 document-sync/api/consistency 皆如此），需与平台团队排查权限/branch 限制或临时增加“WSL on push” 开关；所有尝试需在计划文档登记。 |
| Workflow WSL 矩阵未落库 | `.github/workflows/document-sync.yml`、`api-compliance.yml`、`consistency-guard.yml` 虽已合入 `selfhosted-wsl`，但 `workflow_dispatch` API 多次返回 204 却未生成任何 run（`gh run list` 始终只显示 push 事件 run）。 | 无法实际运行 WSL job，Plan 269 的“document-sync/api/consistency 记录首个 WSL run”项被阻塞。 | 排查 GitHub Actions 行为：对比 `ci-selfhosted-smoke`（能成功 dispatch）的配置差异，必要时提交支持工单或临时新增专用 workflow 以获取运行记录。 |

## 3. 下一步待办

### 3.1 Runner TLS 诊断闭环

1. **锁定复现容器**：依赖 `docker-compose.runner.persist.yml` 中的 `cubecastle-gh-runner`。使用 `docker compose -f docker-compose.runner.persist.yml ps` 确认容器健康。
2. **采集 TLS 证据**：通过 `bash scripts/network/verify-github-connectivity.sh [--smoke|--output <file>]` 一键执行宿主 + Runner 的 `getent`/`curl（带浏览器 UA）`/`git ls-remote`/`openssl s_client`，日志落盘至 `logs/ci-monitor/network-*.log`；若需要复刻 2025-11-19 诊断，可参考 `logs/ci-monitor/network-20251119T104614Z.log`。脚本支持 `--smoke`（仅 getent + curl）和 `--fail-fast` ，默认输出即可附加到 Plan 265 附件。
3. **分析网络路径**：结合上一步日志检查 `gnutls_handshake()` / `curl 56` 是否仍存在；若重现，则记录 DNS 解析、MTU/带宽、代理/证书链信息。若因 Plan 267-D hosts 覆盖而暂时无法复现，也需记录当前 `/etc/hosts.plan267.<timestamp>.bak`、`getent`、`curl`、`git ls-remote` 的输出并同步 Plan 267。
4. **修复/回滚策略**：如确认为网络层问题，则优先引用 Plan 267 提供的方案（静态 hosts / 代理 / 放通）；所有修改都须提供脚本与回滚指南（例如 `sudo cp /etc/hosts.plan267.<timestamp>.bak /etc/hosts && wsl.exe --shutdown` 或 `scripts/network/configure-github-hosts.sh` 重写）。若短期无解，可在 `document-sync`/`api-compliance` 中暂时将 self-hosted job 标记为 optional，并记录到 Plan 265 “回滚窗口”。

### 3.2 consistency-guard 绿灯维护

1. **整改结果**（2025-11-20）：`cmd/hrms-server/query/internal/app/app.go` 通过 `config.ResolveAllowedOrigins("QUERY_ALLOWED_ORIGINS", ...)` 读取配置，`cmd/hrms-server/command/internal/authbff/handler.go` 切换为注入 `config.JWTConfig`，脚本 `ENFORCE=1 scripts/ci/check-hardcoded-configs.sh` 结果为 issues=0。
2. **待执行**：在自托管 runner 上补一次 `consistency-guard` run（使用 `workflow_dispatch`），并将成功 run ID 记录到 Plan 265。若后续新增 CORS/ JWT 相关功能，必须附带同样的配置路径，否则脚本会再次拦截。

### 3.3 Workflow 复跑与文档同步

1. `api-compliance`、`iig-guardian`、`document-sync`：在 TLS 验证通过与配置清理完成后，依次通过 `workflow_dispatch` 触发 self-hosted job，记录 run ID、job ID、commit SHA 及准备/清理脚本日志路径。
2. Plan 265 更新：将 run 结果（含成功/失败原因、回滚状态）写入 `docs/development-plans/265-selfhosted-required-checks.md` 的进展表，并链接到 `logs/ci-monitor/` 中的诊断文件。
3. Branch Protection：在任一 workflow 的 self-hosted job 连续成功 ≥3 次后（记录 run ID 列表），向 DevInfra 提交变更申请，将 `api-compliance (selfhosted)`、`document-sync (selfhosted)`、`consistency-guard (selfhosted)` 添加到 GitHub 保护规则；若任一 job 再次失败，按 Plan 265 的回滚步骤临时移除 Required 状态并补充事故记录。
4. **WSL Runner 记录**：完成 `scripts/ci/runner/wsl-install.sh` + `wsl-verify.sh` 后，利用 `workflow_dispatch` 触发 `document-sync`/`api-compliance`/`consistency-guard`/`ci-selfhosted-smoke` 的 `runs-on: [self-hosted,cubecastle,wsl]` job，Run ID + `logs/wsl-runner/*.log` + `_diag` 路径需同步至 Plan 265/269；若执行失败必须立即用 `wsl-uninstall.sh` 回滚，并在本计划中登记失败原因/回滚时间。

### 3.4 里程碑验收

- **M1（2025-11-21）**：完成 TLS 证据采集 + 两项硬编码整改，`scripts/ci/check-hardcoded-configs.sh` 在本地为绿色。
- **M2（2025-11-24）**：三大 workflow 在 self-hosted runner 上跑通一次，并额外补齐 `self-hosted,cubecastle,wsl` 标签的首个成功 run；Plan 265/269 记录 Run ID 与日志。
- **M3（2025-11-27）**：连续 3 次自托管绿灯并完成 Branch Protection 更新。


> ⚠️ 搁置说明（2025-11-20）：Plan 265/269 依赖的 WSL 自托管验证因 Runner 网络/调度受阻而暂停，以下 run ID 保留历史记录，后续自托管 run 需在网络恢复并重新触发后更新。计划暂不再新增新的自托管 run 记录。

## 4. 附录：最新 run ID

> ⚠️ 搁置说明（2025-11-20）：Plan 265/269 依赖的 WSL 自托管验证因 Runner 网络/调度受阻而暂停，以下 run ID 保留历史记录，后续自托管 run 需在网络恢复并重新触发后更新。计划暂不再新增新的自托管 run 记录。

| Workflow | Run ID / Job ID | 结果 | 备注 |
|----------|------------------|------|------|
| document-sync (ubuntu) | `19489933087` / job `55780035315` | ✅ | 自托管 job 因 TLS 失败 |
| api-compliance (selfhosted) | `19490959491` / job `55782892303` | ❌ (checkout TLS) | 已修复 checkout 顺序，等待 run 19491103285 |
| iig-guardian (selfhosted) | `19491097147` / job 未执行 | ❌ (`paths-filter` 前无 checkout) | YAML 已修正，等待 run 19491533343 |
| consistency-guard (ubuntu) | `19489929404` / job `55780026192` | ❌ (硬编码脚本) | 2025-11-20 本地脚本 issues=0，待自托管 run 验证 |
| document-sync (selfhosted,cubecastle,wsl) | 待触发 | ⏳ | Plan 269 要求记录首个 WSL run，等待脚本/workflow 更新后通过 `workflow_dispatch` 触发 |
