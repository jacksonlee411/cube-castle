# Plan 219C2X – Docker 环境恢复

**文档编号**: 219C2X  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 24 上午  
**负责人**: 组织后端团队（DevOps 协调支持）  
**关联计划**: [219C2Y – 前置条件复位方案](./219C2Y-preconditions-restoration.md)

---

## 1. 目标

1. 启动并验证命令/查询/Redis/PostgreSQL/Temporal 等容器化环境，确保服务健康且无端口冲突。
2. 输出可追溯的环境状态日志，为 219C2Y 与 219C2D 提供唯一事实来源。
3. 建立快速恢复与排障步骤，避免重复劳动并保障后续计划顺利执行。

---

## 2. 范围

| 项目 | 内容 |
| --- | --- |
| Docker Compose 管理 | `docker-compose.dev.yml`，含 PostgreSQL (5432)、Redis (6379)、Temporal (7233)、REST (9090)、GraphQL (8090) 等服务。 |
| 启动命令 | `make docker-up` → `make run-dev`（必要时 `make run-auth-rs256-sim`、`make frontend-dev`）。 |
| 健康检查 | `make status`、`curl http://localhost:9090/health`、`curl http://localhost:8090/health`、`docker compose ps`。 |
| 日志输出 | `logs/219C2/environment-Day24.log`（新增）记录容器状态、端口监听、健康检查结果。 |
| 端口/进程基线 | 复核 `baseline-ports.log`、`baseline-processes.log`，确认无宿主服务占用；如有冲突，执行卸载流程。 |
| 秘钥管理 | 确认 `.cache/dev.jwt`、`secrets/` 未被误删，必要时重新执行 `make jwt-dev-mint`。 |

---

## 3. 前置条件

- 工作站遵循 CLAUDE/AGENTS 指南，未在宿主机安装 PostgreSQL/Redis/Temporal 等冲突服务。
- Docker 引擎与 Compose 正常可用；若存在版本升级需求，先与 DevOps 协作完成升级。
- 拥有执行 `make docker-up` 所需权限，且磁盘空间足以拉起全部容器（建议 ≥5GB 可用空间）。

---

## 4. 详细任务

### 4.1 环境准备
- 清理残留容器：`docker compose -f docker-compose.dev.yml down --remove-orphans`，确保环境干净。
- 检查端口占用：`lsof -i :5432`, `:6379`, `:7233`, `:8090`, `:9090`，确认无宿主进程使用；若发现冲突，按 AGENTS 指南卸载宿主服务（例如 `sudo apt remove postgresql*`）。
- 复核基线文件 `baseline-ports.log`、`baseline-processes.log`，记录当前状态。

### 4.2 容器启动
- 执行 `make docker-up`，等待数据库、缓存、Temporal 等基础服务完成启动。
- 运行 `make run-dev` 启动命令/查询服务（REST 9090/GraphQL 8090）。
- 如验证链需 JWKS，运行 `make run-auth-rs256-sim`；前端依赖时执行 `make frontend-dev`。
- 在 `logs/219C2/environment-Day24.log` 记录关键命令及输出时间戳。

### 4.3 健康检查与日志
- 执行 `docker compose -f docker-compose.dev.yml ps`、`docker compose logs --tail=100`，确认容器无异常。
- 运行健康检查：
  - `make status`
  - `curl -f http://localhost:9090/health`
  - `curl -f http://localhost:8090/health`
- 记录上述命令输出到 `logs/219C2/environment-Day24.log`，确保响应 200 且带有时间戳。
- 若命令失败，立即记录错误输出并执行排障（参见 4.4）。

### 4.4 故障排查（若发生）
- 端口冲突：参考 4.1 卸载宿主服务或释放端口；禁止修改 Compose 端口映射。
- 容器异常退出：检查 `docker compose logs <service>`，根据错误类型处理（例如数据库迁移失败则执行 `make db-migrate-all` 后重试）。
- Temporal/Redis 引导失败：确认数据卷权限与可用空间，必要时清理 `docker volume rm` 后重试。
- 记录所有排障步骤与结论到日志文件，保持唯一事实来源。

### 4.5 输出物整理
- 汇总健康检查结果、日志路径、运行命令至 `logs/219C2/environment-Day24.log`。
- 在 219C2Y 计划中引用该日志作为前置条件满足的证据。
- 若环境需长期保持运行，记录维护注意事项（例如定期 `docker stats`，留意 Temporal 队列积压）。

---

## 5. 交付物

- `logs/219C2/environment-Day24.log`: 启动命令、健康检查、故障排查记录。
- 更新后的 `baseline-ports.log`/`baseline-processes.log`（如有变化）。
- 端口占用/冲突问题的整改记录（若发生）。
- 需要的开发令牌 `.cache/dev.jwt` 续期记录（可选）。

---

## 6. 验收标准

- [x] 所有必需容器启动成功，`docker compose ps` 显示状态为 `running`。（证据：`logs/219C2/environment-Day24.log` “Make Status” 段，执行于 2025-11-06T07:05:40+08:00）
- [x] `make status`、`curl http://localhost:9090/health`、`curl http://localhost:8090/health` 均返回 200，并记录在日志中。（同上日志末尾段，包含两次 curl 200 响应）
- [x] `baseline-ports.log` 与实际端口监听一致，无宿主机冲突服务。（参考日志 “Port Status Verification” 段）
- [x] `logs/219C2/environment-Day24.log` 含完整执行记录、时间戳及健康检查输出。
- [x] 若发生故障，日志中包含排障步骤与结论，确保团队可复现恢复流程。（本次执行未出现异常，日志保留了构建及健康检查全过程）

---

## 7. 时间安排（建议）

| 时间段 | 工作 | 输出 |
| --- | --- | --- |
| 08:30-08:45 | 环境清理与端口检查 | 清理命令、端口检查结果 |
| 08:45-09:15 | `make docker-up` + `make run-dev` | 启动日志、容器状态 |
| 09:15-09:30 | 健康检查与日志整理 | `logs/219C2/environment-Day24.log` |
| 09:30-09:45 | 故障排查/确认（如需） | 排障记录、结论 |
| 09:45-10:00 | 向 219C2Y 报告环境状态 | 更新引用、通知团队 |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解策略 |
| --- | --- | --- |
| 宿主机端口占用导致容器启动失败 | 高 | 立即卸载冲突服务，或在非业务时间段释放端口；记录整改过程。 |
| Docker 镜像或卷损坏 | 中 | 重新拉取镜像、清理卷（谨慎执行 `docker volume rm`），必要时在验收前准备备份。 |
| Temporal/Redis 启动耗时导致 `make run-dev` 超时 | 中 | 增加等待时间或使用 `docker compose logs -f` 追踪，确认服务就绪后再启动应用。 |
| 健康检查未覆盖所有关键服务 | 中 | 扩展检查脚本（如 `curl http://localhost:9090/.well-known/jwks.json`），确保持久化与鉴权链路均可用。 |

---

## 9. 度量与追踪

- `logs/219C2/environment-Day24.log`: 环境状态唯一事实来源。
- `baseline-ports.log`、`baseline-processes.log`: 端口与进程基线对照。
- `docker stats`（可选）: 观察资源占用趋势。
- 219C2Y 计划引用：确认 Docker 环境已满足后续任务需求。
