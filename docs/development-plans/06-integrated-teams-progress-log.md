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
- 阻塞说明：`logs/219E/BLOCKERS-2025-11-06.md`
- 219D2~219D4 日志：`logs/219D2/`, `logs/219D3/`, `logs/219D4/`
- 监控配置：`docs/reference/monitoring/`
- 219E 计划：`docs/development-plans/219E-e2e-validation.md`
