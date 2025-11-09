# 部署与运行检查清单

> 参考：`AGENTS.md` Docker 强制、`Makefile`

- [ ] `make docker-up` 启动 postgres/redis/temporal，无宿主机端口冲突；若冲突，先卸载宿主实例而非修改端口。
- [ ] `make db-migrate-all` 成功执行 Goose 迁移；新迁移文件已在 PR 中登记且通过 code review。
- [ ] `make run-dev`（命令 9090）与 `make frontend-dev`（或 `npm run dev`）成功启动，`/health` 返回 200。
- [ ] 运行结束后 `make docker-down` 或 `docker compose down` 释放资源，日志归档到 `logs/`。
- [ ] 如需 JWT/鉴权，执行 `make jwt-dev-setup` 与 `make jwt-dev-mint`，并将 `.cache/dev.jwt` 加入 `.gitignore`。
