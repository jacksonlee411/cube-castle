# Plan 266 - 自托管 Runner 执行追踪与问题闭环

**文档编号**: 266  
**创建日期**: 2025-11-19  
**关联计划**: Plan 262（Runner 基建）、Plan 265（自托管门禁扩展）

---

## 1. 最新进展

1. `cmd/hrms-server/command/main.go`、`internal/organization/handler/devtools.go`、`tests/e2e/auth_flow_e2e_test.go` 已移除 `http://localhost` 等硬编码，引入 `COMMAND_ALLOWED_ORIGINS`、`COMMAND_BASE_HOST/SCHEME`、`DEVTOOLS_COMMAND_BASE_URL`、`COMMAND_BASE_URL/QUERY_BASE_URL` 等环境变量（commit `b3aff300`）。
2. `api-compliance.yml`、`iig-guardian.yml` 的 `actions/checkout` 均提前至 `paths-filter` 之前，避免自托管 Job 在未拉取仓库时运行 filter 导致 `fatal: not a git repository`。
3. `document-sync.yml` 已允许在 `workflow_dispatch` 场景下强制运行全量步骤（不再被 docs-only fast pass 直接返回），Ubuntu 矩阵 job 已成功生成报告（run `19489933087`）。
4. 2025-11-20：新增 `internal/config/cors.go` + Query CORS 外部化、BFF 统一注入 `internal/config.GetJWTConfig()`；`ENFORCE=1 scripts/ci/check-hardcoded-configs.sh` 已返回 issues=0。
5. 2025-11-20 07:28Z：按照 3.1 步骤在 `cubecastle-gh-runner` 容器采集 TLS 证据，日志位于 `logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`（含 `git ls-remote`/`curl`/`openssl s_client` 输出）。
6. 2025-11-20 晚间：根据 Plan 267-D 执行 `/etc/wsl.conf` `[network]\ngenerateHosts=false` + `sudo bash scripts/network/configure-github-hosts.sh`，宿主与 Runner 侧 `getent hosts github.com`、`curl -I https://github.com`、`git ls-remote https://github.com/jacksonlee411/cube-castle` 均返回 200 并解析到官方 IP（详见 `docs/development-plans/267-docker-network-stability.md:39-43`），`/etc/hosts.plan267.<timestamp>.bak` 记录回滚点。
7. 2025-11-20 10:46Z：首版 `scripts/network/verify-github-connectivity.sh`（Plan 266/267 诊断脚本）在宿主与 `gh-runner` 容器内依次执行 `getent hosts github.com`、带浏览器 UA 的 `curl -sS -D - https://github.com`、`GIT_CURL_VERBOSE=1 git ls-remote https://github.com/jacksonlee411/cube-castle`、`openssl s_client -connect github.com:443`，日志落盘 `logs/ci-monitor/network-20251119T104614Z.log`，所有命令均返回 200/TLS OK，证明 hosts 覆盖后 Runner 同样可以建立 TLS。
8. 2025-11-20 12:47Z / 12:57Z：依次触发自托管版 `document-sync`（run `19502442553`、`19502825153`）。节点缓存已在 Runner `_work/_tool/node/18.20.8/x64` 预热，Job 能走到 “文档同步一致性检查”。但上传工件时复用 `document-sync-report-${{ github.run_number }}` 导致 409 conflict（相同 run 上多次 attempt），GitHub 将自托管 job 标记为取消；workflow 已改为 `document-sync-report-${{ github.run_number }}-${{ github.run_attempt }}`，仍待推送后验证。Run 日志同时记录 checkout 阶段偶发 `GnuTLS recv error (-110)`，需继续跟踪网络稳定性。

## 2. 遇到的问题 / 风险

| 问题 | 描述 | 当前影响 | 负责人/协作 |
|------|------|----------|-------------|
| 自托管 checkout TLS 断线 | `document-sync` Self-hosted job 在 `actions/checkout` 阶段多次出现 `gnutls_handshake()` / `curl 56`，无法从 GitHub 拉代码；2025-11-20 07:29Z 现场复现（log `logs/ci-monitor/selfhosted-tls-20251119T072846Z.log`）显示连接被 11.2.0.12 截断、`openssl s_client` 无法拿到证书；Plan 267-D 通过静态 hosts 临时恢复了宿主/Runner 的 `curl`/`git ls-remote` | 目前依赖 Plan 267-D（WSL `generateHosts=false` + `scripts/network/configure-github-hosts.sh`）维持访问，若 hosts 再次被覆盖自托管 job 仍会失败，Plan 265 仍缺少成功 run（`19489933087` selfhosted） | DevInfra（排查 runner 网络/TLS，Plan 267 负责网络方案） |
| `api-compliance` run 长时间排队 | workflow_dispatch `19491103285` 仍 queued，自托管修复尚未验收 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |
| `iig-guardian` run 未执行 | workflow_dispatch `19491533343` queued，同上 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |
| Artifact 命名冲突导致 selfhosted 失效 | 自托管 `document-sync` 运行多次尝试（run `19502442553`、`19502825153`）时，`actions/upload-artifact@v4` 使用固定名称 `document-sync-report-${{ github.run_number }}`，GitHub 不允许在同一 run 中重复创建同名工件，于是上传返回 409、Job 被标记为 “The operation was canceled”。 | 自托管 run 无法进入清理/后续步骤；Plan 265 仍缺少成功 run。workflow 已改为 `document-sync-report-${{ github.run_number }}-${{ github.run_attempt }}`，需推送并重跑验证 | 平台组（更新 workflow、确认 artifact 命名不会冲突） |

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

### 3.4 里程碑验收

- **M1（2025-11-21）**：完成 TLS 证据采集 + 两项硬编码整改，`scripts/ci/check-hardcoded-configs.sh` 在本地为绿色。
- **M2（2025-11-24）**：三大 workflow 在 self-hosted runner 上跑通一次，Plan 265 记录 run ID 与日志。
- **M3（2025-11-27）**：连续 3 次自托管绿灯并完成 Branch Protection 更新。

## 4. 附录：最新 run ID

| Workflow | Run ID / Job ID | 结果 | 备注 |
|----------|------------------|------|------|
| document-sync (ubuntu) | `19489933087` / job `55780035315` | ✅ | 自托管 job 因 TLS 失败 |
| api-compliance (selfhosted) | `19490959491` / job `55782892303` | ❌ (checkout TLS) | 已修复 checkout 顺序，等待 run 19491103285 |
| iig-guardian (selfhosted) | `19491097147` / job 未执行 | ❌ (`paths-filter` 前无 checkout) | YAML 已修正，等待 run 19491533343 |
| consistency-guard (ubuntu) | `19489929404` / job `55780026192` | ❌ (硬编码脚本) | 2025-11-20 本地脚本 issues=0，待自托管 run 验证 |
