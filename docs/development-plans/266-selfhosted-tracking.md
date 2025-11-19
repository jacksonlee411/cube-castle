# Plan 266 - 自托管 Runner 执行追踪与问题闭环

**文档编号**: 266  
**创建日期**: 2025-11-19  
**关联计划**: Plan 262（Runner 基建）、Plan 265（自托管门禁扩展）

---

## 1. 最新进展

1. `cmd/hrms-server/command/main.go`、`internal/organization/handler/devtools.go`、`tests/e2e/auth_flow_e2e_test.go` 已移除 `http://localhost` 等硬编码，新增 `COMMAND_ALLOWED_ORIGINS`、`COMMAND_BASE_HOST/SCHEME`、`DEVTOOLS_COMMAND_BASE_URL`、`COMMAND_BASE_URL/QUERY_BASE_URL` 等环境变量，满足 `consistency-guard` 的硬编码要求（commit `b3aff300`）。
2. `api-compliance.yml`、`iig-guardian.yml` 的 `actions/checkout` 均提前至 `paths-filter` 之前，避免自托管 Job 在未拉取仓库时运行 filter 导致 `fatal: not a git repository`。
3. `document-sync.yml` 已允许在 `workflow_dispatch` 场景下强制运行全量步骤（不再被 docs-only fast pass 直接返回），Ubuntu 矩阵 job 已成功生成报告（run `19489933087`）。

## 2. 遇到的问题 / 风险

| 问题 | 描述 | 当前影响 | 负责人/协作 |
|------|------|----------|-------------|
| 自托管 checkout TLS 断线 | `document-sync` Self-hosted job 在 `actions/checkout` 阶段多次出现 `gnutls_handshake()` / `curl 56`，无法从 GitHub 拉代码 | 自托管 job 未执行，Plan 265 无法记录成功 run（`19489933087` selfhosted） | DevInfra（排查 runner 网络/TLS） |
| `consistency-guard` 仍报硬编码 | 最新 push run `19489929404` 的 ubuntu job 因 `scripts/ci/check-hardcoded-configs.sh` 检测到多处 localhost/CORS/JWT 硬编码而 exit 4 | 自托管矩阵 job 被取消，Plan 265 无法记录绿灯 | 命令服务 + QA（进一步外部化或白名单） |
| `api-compliance` run 长时间排队 | workflow_dispatch `19491103285` 仍 queued，自托管修复尚未验收 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |
| `iig-guardian` run 未执行 | workflow_dispatch `19491533343` queued，同上 | 暂无 run ID 可记录 | GitHub Actions 排队，需等待 |

## 3. 下一步待办

1. **Runner TLS 排查**：在自托管容器内手动执行 `git clone https://github.com/jacksonlee411/cube-castle`、`curl https://github.com` 等，分析是否因代理/证书/带宽导致 TLS 中断；若需，可在 runner 镜像内安装 CA 或设置 `GIT_TRACE_CURL`. 目标是让 `document-sync` self-hosted job 能完成 checkout 与 `node` 步骤。
2. **重跑关键 workflows**：待 `api-compliance`、`iig-guardian` 队列出结果后，提取自托管 job 是否成功；若仍失败，围绕日志定位新的问题。成功 run 的 ID 需回填 Plan 265。
3. **consistency-guard 清理**：梳理 `scripts/ci/check-hardcoded-configs.sh` 的输出（`cmd/hrms-server/command/main.go`、`tests/e2e/auth_flow_e2e_test.go` 等），逐项改造或在脚本配置白名单，直至自托管矩阵能产出绿灯。
4. **Plan 265 文档更新**：记录成功 run ID（document-sync ubuntu、自托管 runner TLS 问题等），标明未完成项及风险。
5. **Branch Protection 准备**：在 self-hosted job 连续成功 3 次后，将 `api-compliance (selfhosted)`、`document-sync (selfhosted)` 等状态加入受保护检查列表。

## 4. 附录：最新 run ID

| Workflow | Run ID / Job ID | 结果 | 备注 |
|----------|------------------|------|------|
| document-sync (ubuntu) | `19489933087` / job `55780035315` | ✅ | 自托管 job 因 TLS 失败 |
| api-compliance (selfhosted) | `19490959491` / job `55782892303` | ❌ (checkout TLS) | 已修复 checkout 顺序，等待 run 19491103285 |
| iig-guardian (selfhosted) | `19491097147` / job 未执行 | ❌ (`paths-filter` 前无 checkout) | YAML 已修正，等待 run 19491533343 |
| consistency-guard (ubuntu) | `19489929404` / job `55780026192` | ❌ (硬编码脚本) | 需进一步整改 |
