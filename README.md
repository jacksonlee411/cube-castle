# Cube Castle

本仓库的 README 仅作为最小索引。项目的原则、约束、流程与权威链接以 `AGENTS.md` 为唯一事实来源。

## 快速开始
- 必备：Docker + Docker Compose；Go 1.24+；Node 18+（详情见 `AGENTS.md`）
- 一键启动（容器化，已含迁移）：`make run-dev` → `make frontend-dev` → `make status`
- 健康检查：`curl http://localhost:9090/health` 与 `curl http://localhost:8090/health`
- 退出/重置：`make docker-down` / `make reset`

## 硬性约束（摘录，完整见 AGENTS.md）
- 仅容器化：PostgreSQL、Redis、服务均由 Docker Compose 管理；严禁在宿主机安装同名服务
- 端口冲突处理：必须卸载宿主服务释放 5432/6379 等端口，禁止修改 `docker-compose.dev.yml` 端口映射规避冲突
- 迁移即真源：所有数据库变更只通过 `database/migrations/` 执行；示例数据位于 `sql/seed/`（非事实来源）

## 关键链接
- 原则与索引（唯一）：`AGENTS.md`
- API 契约：`docs/api/openapi.yaml` · `docs/api/schema.graphql`
- 参考文档：`docs/reference/00-README.md`
- 计划与归档：`docs/development-plans/` · `docs/archive/development-plans/`
- 变更记录：`CHANGELOG.md`

## 常用命令
- 构建：`make build`；清理：`make clean`
- 数据库迁移：`make db-migrate-all`；回滚：`make db-rollback-last`
- 前端：`make frontend-dev`；测试：`npm --prefix frontend run test` / `npm --prefix frontend run test:e2e`
- 开发 JWT：`make jwt-dev-setup` · `make jwt-dev-mint`（令牌保存在 `.cache/dev.jwt`）

## 说明
- 若 README 与 `AGENTS.md` 或 `docs/reference/*` 存在任何不一致，以 `AGENTS.md` 为准；请先暂停变更并对齐事实来源。
