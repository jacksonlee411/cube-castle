# 221T – Docker 集成测试基座验证记录

> 适用范围：Plan 221 交付完成后，需在 1 个工作日内完成一次全链路验证。所有数据与日志必须引用唯一事实来源（`docker-compose.test.yml`、`scripts/run-integration-tests.sh`、`docs/development-guides/docker-testing-guide.md`）。

## 1. 前置条件与责任登记

| 项目 | 校验命令 / 说明 | 责任人 | 完成时间 |
| --- | --- | --- | --- |
| 5432 端口空闲 | `lsof -i :5432` → 无输出；如有，先卸载宿主 PostgreSQL（遵循 `AGENTS.md` Docker 强制） | shangmeilin | 2025-11-09 15:03 CST |
| Docker & Compose 可用 | `docker version` / `docker compose version` | shangmeilin | 2025-11-09 15:04 CST |
| Go 与 Goose 已安装 | `go version` ≥ 1.24.0；`go install github.com/pressly/goose/v3/cmd/goose@latest` | shangmeilin | 2025-11-09 15:04 CST |
| 环境变量确认 | `DATABASE_URL` 未预设（由脚本设置）；`.env` 中无指向宿主 Postgres 的变量 | shangmeilin | 2025-11-09 15:04 CST |
| 计划文档同步 | 已阅读 `docs/development-plans/221-docker-integration-testing.md` 最新版本 | shangmeilin | 2025-11-09 14:55 CST |

> 若任一项不满足，请先更新责任人/完成时间并整改，之后再进入第 2 章。

## 2. 执行步骤（记录实际结果）

| 步骤 | 命令 / 操作 | 预计耗时 | 验证点 | 记录（责任人/时间/备注） |
| --- | --- | --- | --- | --- |
| 1 | `make test-db-up` | 10s | `docker compose ps` 中 `postgres-test` 状态为 `healthy`；启动日志 < 10s | shangmeilin / 2025-11-09 14:59 CST / 首次拉取 postgres:15-alpine |
| 2 | `make test-db-logs` | 5s | 日志包含 `database system is ready to accept connections`；无错误堆栈 | shangmeilin / 2025-11-09 15:00 CST / 确认 ready |
| 3 | `make test-db` | 3-4 min | 脚本自动执行 Goose up → `go test -v -tags=integration ./...` → Goose down；命令返回 0 | shangmeilin / 2025-11-09 15:02 CST / 详见 `logs/plan221/run-20251109145841.log` |
| 4 | `make test-db-down` | 5s | `docker compose ps` 无运行容器；卷被清理（可选 `docker volume ls | grep postgres-test`) | shangmeilin / 2025-11-09 15:03 CST / 验证无残留容器 |
| 5 | 产物归档 | 将 `logs/plan221/`（或终端输出）复制到 `logs/plan221/<timestamp>.log`，并在 PR 记录 | shangmeilin / 2025-11-09 15:04 CST / `logs/plan221/run-20251109145841.log` |
| 6 | CI 验证 | 手动触发 `.github/workflows/integration-test.yml` 或使用 `act`，完成后记录 run 链接 | shangmeilin / 2025-11-09 15:05 CST / 本地执行等同 workflow；待恢复 GitHub 访问后补跑线上 run |

> **日志命名规范**：`logs/plan221/run-<YYYYMMDDHHMMSS>.log`，首行需写明分支、commit、执行人、命令。

## 3. 成功判定

1. `make test-db` 返回 0，终端输出包含：
   - goose 报告最近一次 `up` 已执行（如 `goose: successfully applied migration <timestamp>_<name>.sql`）；
   - `ok   cube-castle/...` 等 Go 集成测试通过信息（`-tags=integration`）；
   - goose `down` 成功（如 `goose: successfully rolled back migration ...`）。
2. `docker compose ps` 在执行完成后无残留容器；`docker volume ls` 中不存在 `postgres-test-data`（脚本已 down -v）。
3. 日志中未出现 `port already in use`、`connection refused`、`psql: could not connect` 等错误。
4. CI 工作流 `integration-test` 最近一次运行状态为 ✅，且日志中可见 `make test-db` 输出。

## 4. 证据与同步

- `logs/plan221/run-*.log`：附加到 PR 或 Wiki，供审阅者查阅。
- `docs/development-plans/221-docker-integration-testing.md`：更新“执行状态”章节，引用本次日志。
- `docs/reference/02-IMPLEMENTATION-INVENTORY.md`：若新增脚本或路径，请同步实施清单。
- Issue/PR 备注：粘贴责任人、时间、日志路径，说明是否触发 CI 复核。

## 5. 常见异常与处置

| 异常 | 表现 | 处置 |
| --- | --- | --- |
| 5432 被占用 | `make test-db-up` 报错 `port is already allocated` | 按 `AGENTS.md` 卸载宿主 PostgreSQL/服务，确认 `lsof -i :5432` 无输出后重试 |
| goose 执行失败 | `goose: failed to connect` | 检查 `DATABASE_URL` 是否被外部环境覆盖；必要时 `unset DATABASE_URL` 再运行 `make test-db` |
| Go 测试失败 | `go test` 返回非 0 | 记录日志，提交 Issue，并在 Plan 221 文档说明失败原因及回滚方案 |
| CI 失败 | workflow 结束状态 ❌ | 下载日志，确认是否与宿主环境冲突一致，处理后重新触发 |

---

维护人：Plan 221 执行团队  
创建时间：2025-11-07  
引用：`docs/development-guides/docker-testing-guide.md`、`docker-compose.test.yml`、`scripts/run-integration-tests.sh`
