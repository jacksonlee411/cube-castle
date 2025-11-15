# Repository Guidelines

> ⚠️ 资源唯一性与跨层一致性为最高优先级约束：所有代理执行前需确保不引入第二事实来源或跨层不一致，一旦发现必须立即中止并回滚。

> 🐳 **Docker 容器化部署强制约束**：本项目所有服务、数据库、中间件（如 PostgreSQL、Redis）必须通过 Docker Compose 管理，**严禁在宿主机直接安装**。如发现宿主服务占用容器端口（如 5432、6379），必须卸载宿主服务以释放端口，**不得调整容器端口映射**以迁就宿主服务。此约束确保开发环境一致性、隔离性与可复现性。

## 项目结构与模块组织
- 命令服务位于 `cmd/hrms-server/command/`，查询服务位于 `cmd/hrms-server/query/`，共享中间件、鉴权、缓存与 GraphQL 工具集中在 `internal/`，严格遵循 PostgreSQL 原生 CQRS（命令→REST、查询→GraphQL）。
- 数据迁移统一保存在 `database/migrations/`，通用 SQL 助手位于 `sql/`；禁止回退至 `sql/init/01-schema.sql` 等历史脚本，数据真源始终由迁移驱动。
- 前端代码在 `frontend/`，功能切片位于 `frontend/src/features/`，共用类型在 `frontend/src/shared/`；静态资源、脚本与测试说明遵循各目录 README。
- 测试分布：Go 与集成测试在 `tests/` 和 `cmd/*`，前端 Vitest 规格在 `frontend/src/**/__tests__` 与 `frontend/tests/`，Playwright E2E 在 `tests/e2e/`，其配置归档于 `frontend/`。

## 开发前必检
- 确认本地 Go 环境版本 ≥1.24（执行 `go version`，需与仓库 `toolchain go1.24.9` 一致）。
- 运行 `node scripts/generate-implementation-inventory.js` 对照 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，避免重复造轮子。
- 校验契约：查阅 `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`，确认字段保持 camelCase 与 `{code}` 路径参数，任何偏差需先更新契约。
- 在 `docs/development-plans/` 建立或更新计划，完成后归档至 `docs/archive/development-plans/`，并记录验收标准；计划内容需引用唯一事实来源并说明一致性校验。
- 如需快速确认环境，可执行 `make status`、`curl http://localhost:9090/health` 与 `curl http://localhost:8090/health`（命令返回 200 表示核心服务就绪）。

## 构建、测试与开发命令
- **基础设施与服务（Docker 强制）**：最小依赖通过 `make docker-up` 启动（PostgreSQL 5432、Redis 6379）。当前仓库不提供 Temporal 工作流引擎编排，相关功能未启用。**严禁**在宿主机安装这些服务，如遇端口冲突必须卸载宿主服务。启动后执行 `make run-dev`（端口 9090/8090）→ `make frontend-dev`；`make run-auth-rs256-sim` 已合并至 `make run-dev`（容器化）。
- 编译与清理：`make build`、`make clean`；数据库迁移使用 `make db-migrate-all`，日志追踪可查阅 `run-dev*.log`。
- 测试：`make test`、`make test-integration`、`make coverage`，前端 `cd frontend && npm run test` 或 `npm run lint`，E2E 使用 `npm run test:e2e`。
- 鉴权链路：`make jwt-dev-setup`、`make jwt-dev-mint`，令牌存放 `.cache/dev.jwt`；必要时通过 `curl http://localhost:9090/.well-known/jwks.json` 验证公钥。

## 编码风格与命名约定
- Go 采用内部 camelCase、导出 PascalCase，提交前执行 `make fmt` 与 `make lint`；服务领域逻辑聚合在 `cmd/*/internal/` 并保持事务边界清晰，任何跨层命名偏差视为一致性违规。
- TypeScript 固定两空格缩进、ESLint 与函数式组件；共享类型放入 `frontend/src/shared/`，API 客户端统一使用 `frontend/src/shared/api/`，组件命名遵循 PascalCase。

## 测试与质量校验
- Go 测试文件以 `_test.go` 结尾，必要时添加 `//go:build integration` 标签区分场景；前端单测紧邻功能模块并使用 Vitest。
- 推送前执行 `frontend/scripts/validate-field-naming*.js`、`node scripts/quality/architecture-validator.js`、`make security` 与 `npm run lint`，确保与 CI 校验一致。
- Playwright 规格按业务场景命名（如 `organization-create.spec.ts`），通过环境变量 `PW_TENANT_ID`、`PW_JWT` 注入租户与令牌。
- Playwright 配置入口：`frontend/playwright.config.ts`，支持 `PW_JWT`、`PW_SKIP_SERVER`、`PW_BASE_URL` 等环境变量。

## 临时方案管控
- 仅引用规则：必须以 `// TODO-TEMPORARY:` 标注原因/计划/截止日期（不超过一个迭代），建立清单并按期回收。
- 白名单：`scripts/todo-temporary-allowlist.txt`；严禁在 `frontend/src/shared/types/api.ts` 保留临时导出（校验脚本会拦截）。
- 校验脚本：`scripts/check-temporary-tags.sh`；CI 工作流：`.github/workflows/agents-compliance.yml`。

## 提交与拉取请求规范
- 提交信息遵循 Conventional Commits（示例：`feat: add temporal validation`），单次提交聚焦单一主题并附带回归验证。
- PR 必须关联 Issue，说明行为变化、测试证据、回滚路径；若契约或行为变更，请同步更新 `docs/reference/` 与相关计划文档，并引用 `CLAUDE.md` 作为原则依据。
- 评论区需明确剩余风险、待办与迁移步骤，审阅者以 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 为核对清单。

## 安全与配置提示
- **Docker 环境隔离**：所有数据库、缓存、消息队列必须运行在 Docker 容器内，数据卷统一由 Docker Compose 管理（`postgres_data`、`redis_data` 等）。如遇宿主机服务占用容器端口（如 PostgreSQL 占用 5432），必须卸载宿主服务以释放端口，**禁止修改 docker-compose.dev.yml 端口映射**来迁就宿主服务。示例：Ubuntu/Debian 执行 `sudo apt remove postgresql*`；macOS 执行 `brew services stop postgresql && brew uninstall postgresql`；Windows 在“应用和功能”中卸载或以 PowerShell 停用相关服务后卸载。
- 所有环境初始化均通过迁移脚本完成；必要时使用 `make db-migrate-all` 或 `make db-rollback-last`（若可用）进行回滚，再同步更新计划文档。
- 秘钥统一存放于 `secrets/`，严禁提交到版本库；调试时通过 `make jwt-dev-export` 导出会话令牌，并遵循 `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`。
- 若出现异常，优先参考 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 与 `CHANGELOG.md`，若与本指南冲突，以上述权威文档与 `CLAUDE.md` 为最终解释。
