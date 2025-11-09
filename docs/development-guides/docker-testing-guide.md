# Docker 集成测试指南（Plan 221）

> 唯一事实来源：`docker-compose.test.yml`、`scripts/run-integration-tests.sh`、`Makefile` 中的 `test-db*` 目标以及 `docs/development-plans/221-docker-integration-testing.md`。

## 1. 适用场景
- 在本地重放 Goose 迁移 + 集成测试链路。
- 在 CI 里复用与本地一致的 Docker 测试基座。
- 为 Plan 222 验收或新模块集成测试提供一键脚本。

## 2. 前置要求
1. 遵守 `AGENTS.md`/`CLAUDE.md` 的 **Docker 强制**：如果宿主机安装了 PostgreSQL/Redis，请卸载以释放 5432/6379，禁止修改容器端口来规避冲突。
2. 安装 Docker、Docker Compose Plugin、Go ≥ `1.24.0`、Goose (`go install github.com/pressly/goose/v3/cmd/goose@latest`)。
3. 运行 `node scripts/generate-implementation-inventory.js` 与 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 中的快速检查，确保环境一致。

## 3. 关键文件
| 文件 | 作用 |
|------|------|
| `docker-compose.test.yml` | 启动 PostgreSQL 15 测试容器，端口 5432 对齐主环境。|
| `scripts/test/init-db.sql` | 初始化扩展/辅助 SQL（在容器启动阶段执行）。|
| `scripts/run-integration-tests.sh` | 启动容器→Goose up→Go 集成测试→Goose down→清理。|
| `Makefile` (`test-db*`) | 对脚本与 docker compose 的薄封装，方便开发/CI 调用。|
| `.github/workflows/integration-test.yml` | CI 流水线，直接执行 `make test-db`。|

## 4. 本地使用流程
```bash
# 启动测试数据库（如需手动探索）
make test-db-up

# 运行完整的集成测试链路（会自动 up/down）
make test-db

# 查看测试数据库日志
make test-db-logs

# 通过 psql 连接容器中的数据库
make test-db-psql

# 清理容器与卷
make test-db-down
```

> ⚠️ 如果 `make test-db` 提示 5432 被占用，请先移除宿主机 PostgreSQL 服务，再重试；严禁通过修改 `docker-compose.test.yml` 端口规避（违背 `AGENTS.md`）。

## 5. CI 集成
- `.github/workflows/integration-test.yml` 在 push/pull_request 时运行，直接执行 `make test-db`，确保与本地完全一致。
- 如需在其它流水线中复用，可在步骤中添加 `make test-db`，或引用 `scripts/run-integration-tests.sh`。

## 6. 故障排查
| 症状 | 可能原因 | 解决方案 |
|------|----------|----------|
| `pg_isready` 一直失败 | 宿主机已有 PostgreSQL | 根据 `AGENTS.md` 删除宿主服务，释放 5432。|
| `goose` 命令不存在 | 未安装 goose | 执行 `go install github.com/pressly/goose/v3/cmd/goose@latest` 并确保 GOPATH 在 PATH 中。|
| 集成测试读取不到 DATABASE_URL | 自定义脚本未导出 | 通过 `export DATABASE_URL=...` 或执行 `scripts/run-integration-tests.sh`。|
| CI `make test-db` 失败 | Docker 资源不足 | 在 workflow 中添加 `docker system df` 诊断或拆分并行作业。|

---

维护人：Plan 221 执行团队
最后更新：2025-11-07
