# Plan 06 – 集成测试行动记录（2025-11-06 更新）

> 唯一事实来源：`docs/development-plans/219E-e2e-validation.md`、`logs/219E/`。本记录仅聚焦当前阻塞与下一步行动，便于与 219 系列计划协同。

## 当前状态
- 219D1~219D5 已完成，Scheduler 配置/监控/测试资料可直接复用。
- 219E 端到端与性能验证因宿主环境无法访问 Docker daemon (`/var/run/docker.sock` permission denied) 暂停；详见 `logs/219E/BLOCKERS-2025-11-06.md`。
- 新的测试/性能脚本已入库，但尚未执行：
  - `scripts/e2e/org-lifecycle-smoke.sh`（组织/部门生命周期冒烟，输出 `logs/219E/org-lifecycle-*.log`）
  - `scripts/perf/rest-benchmark.sh`（REST P99 基准，输出 `logs/219E/perf-rest-*.log`）

## 后续步骤（需按顺序推进）
1. **恢复 Docker 访问**
   - 为当前用户授予 `/var/run/docker.sock` 访问权限（或在具备权限的宿主/runner 上执行）。
   - 验证 `docker ps` / `make run-dev` 成功，保证命令/查询服务及依赖容器可启动。

   **快速诊断：**
   ```bash
   # 一键诊断脚本
   bash -c 'echo "=== Docker 权限诊断 ==="; echo "1. Docker daemon:"; docker ps 2>&1 | head -2; echo "2. socket信息:"; ls -la /var/run/docker.sock 2>/dev/null || echo "不存在"; echo "3. 当前用户:"; whoami; echo "4. 用户所在组:"; groups; echo "5. docker组成员:"; getent group docker || echo "docker组不存在"'
   ```

   **配置方案（按优先级）：**

   **方案A：Linux用户组授权（推荐 - 一次性配置）**
   ```bash
   # 1. 创建 docker 用户组（如果不存在）
   sudo groupadd docker 2>/dev/null || true

   # 2. 将当前用户加入 docker 组
   sudo usermod -aG docker $(whoami)

   # 3. 更新用户组成员关系（重新获取组权限）
   newgrp docker

   # 4. 验证
   docker ps
   ```

   **若方案A失败（常见于WSL2/无sudo权限）：方案B - 修改socket权限**
   ```bash
   # 临时方案（重启后失效）
   sudo chmod 666 /var/run/docker.sock

   # 永久方案：编辑Docker daemon配置
   sudo nano /etc/docker/daemon.json
   # 添加或修改以下内容：
   # {
   #   "unix-socket-group": "docker",
   #   "unix-socket-permissions": "0660"
   # }

   # 重启Docker daemon
   sudo systemctl restart docker

   # 重新加入docker组并验证
   newgrp docker
   docker ps
   ```

   **常见陷阱与解决：**
   | 问题 | 原因 | 解决方法 |
   |------|------|--------|
   | 执行 `usermod` 后仍 permission denied | 需重新登录或使用 `newgrp` | 运行 `newgrp docker` 或重启终端 |
   | 新终端仍报 permission denied | 新终端未加载新组关系 | 关闭所有终端重新打开，或 `exec su -l $USER` |
   | WSL中重启Docker后仍不工作 | Docker daemon 未正确启动 | `sudo service docker status` / `sudo service docker restart` |

   > **外部权限配合说明**：上述 `sudo groupadd/usermod`、`chmod`、`systemctl` 等步骤均需要具备 `sudo` 权限。当前代理环境因 `no new privileges` 限制无法自行提升权限，需由宿主/平台运维代为执行方案A/B，或直接提供已完成配置、可运行 Docker 的 runner/机器。未获得外部协助前，219E 相关脚本无法落地。

   **验证Docker访问已恢复：**
   ```bash
   # 运行以下命令，应不报 permission denied
   docker ps
   docker images | grep cube-castle

   # 启动全栈服务
   make run-dev
   # 或
   docker compose -f docker-compose.dev.yml up -d

   # 验证关键服务运行
   docker ps | grep -E 'postgres|redis|rest-service|graphql-service'
   ```
2. **执行组织生命周期冒烟脚本**  
   - 在服务就绪后运行 `scripts/e2e/org-lifecycle-smoke.sh`（可自定义 `COMMAND_API`、`TENANT_ID`、`JWT_TOKEN`）。  
   - 将生成的 `logs/219E/org-lifecycle-*.log` 附加到 219E 计划验收记录，并检查 REST/GraphQL 返回值是否符合预期。
3. **采集 REST 性能基准**  
   - 安装 `hey`（`go install github.com/rakyll/hey@latest`）并运行 `scripts/perf/rest-benchmark.sh`。  
   - 收集 `logs/219E/perf-rest-*.log` 中的 P95/P99 数据，与历史基线对比并在 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 性能章节登记。
4. **补充 Playwright/前端 E2E 及缓存场景**  
   - 复用 `frontend/tests/e2e/*.spec.ts` 与 `tests/e2e/organization-validator/`，重点验证 Assignment、Outbox→Dispatcher→缓存刷新路径。  
   - 失败用例整理至 `logs/219E/`，并在 219E 文档“测试范围”表格更新状态。
5. **回退演练与报告**  
   - 依据 `internal/organization/README.md#Scheduler / Temporal（219D）` 的回退指引，演练一次 `SCHEDULER_ENABLED=false` 与 219D1 目录回退，记录命令与验证日志。  
   - 将回退步骤与结果同步到 219E 文档及 `logs/219E/rollback-*.log`。

## 参考资料
- **Docker权限配置：** 本文档第1步"恢复Docker访问"中的诊断与配置方案
- **阻塞说明：** `logs/219E/BLOCKERS-2025-11-06.md`
- **服务日志：** `logs/219D2/`, `logs/219D3/`, `logs/219D4/`
- **监控配置：** `docs/reference/monitoring/`
- **219E计划：** `docs/development-plans/219E-e2e-validation.md`
- **Docker Compose配置：** `docker-compose.dev.yml`（本地开发服务定义）
- **Makefile相关目标：** `make run-dev`、`make stop-dev`

## Docker权限问题根本原因（背景知识）
- `/var/run/docker.sock` 是Docker daemon的Unix socket，权限通常为 `660`（仅 `root:docker` 可访问）
- 当前环境（WSL2：Linux 5.15.167.4-microsoft-standard）中，用户未在 `docker` 组内，导致无法访问socket
- 本地 `sudo` 因 `no new privileges` 限制不可用，需通过用户组授权或修改daemon配置解决
- CI/CD环境中，Docker daemon 通常以特定用户运行，需确保构建用户具备相应权限或在 docker 组内
