# CLAUDE.md（精简版）

项目指导原则与“单一事实来源”索引。本文件仅保留长期稳定的原则与链接；所有易变细节（变更通告、工作流细则、脚本说明）统一迁移到对应权威文件。

—

## 1. 范围与目标
- 作用：凝练团队价值观与不变的工作边界；提供权威资料的索引入口。
- 不再承载：变更通告、具体步骤、命令清单、细则条款；请至下文链接查阅。

—

## 2. 核心开发指导原则（长期稳定）
- 资源唯一性与跨层一致性（最高优先级）：所有实现、文档与契约必须保持唯一事实来源与端到端一致，违背时优先回滚或整改。
- **Docker 容器化部署（强制）**：所有服务、数据库、中间件统一通过 Docker Compose 管理，严禁在宿主机直接安装 PostgreSQL、Redis、Temporal 等组件。如发现宿主服务占用容器端口，必须卸载宿主服务而非调整容器端口映射。
- 诚实原则：状态、性能、风险基于可验证事实，不夸大、不隐瞒。
- 悲观谨慎：按最坏情况评估，分阶段验证并预留缓冲。
- 健壮优先：根因修复与可维护性优先，配套测试与文档。
- 中文沟通：提交物与沟通（含与智能代理、自动化流程互动）优先使用专业、准确、清晰的中文。
- 先契约后实现：以 `docs/api/` 为唯一事实来源，先定义再实现。
- PostgreSQL 原生 CQRS：查询统一 GraphQL；命令统一 REST；单一数据源 PostgreSQL。

—

## 3. 工作分工与边界（不变项）
- CQRS 分工：查询 → GraphQL；命令 → REST（不得混用）。
- 权限契约：以 OpenAPI 为准，先契约、后实现、前后端一致。
- 命名一致性：API 对外字段一律 camelCase；组织单元路径参数统一 `{code}`。
- 唯一事实来源：契约、实现、文档需指向同一来源，任何复制或平行事实视为最高优先级缺陷。

—

## 4. 单一事实来源索引（权威链接）
- API 契约与权限声明：
  - `docs/api/openapi.yaml`（REST + 权限 scopes）
  - `docs/api/schema.graphql`（GraphQL 查询）
- 参考手册（长期稳定）：`docs/reference/`
  - 开发者速查：`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
  - 实现清单：`docs/reference/02-IMPLEMENTATION-INVENTORY.md`
  - API/工具：`docs/reference/03-API-AND-TOOLS-GUIDE.md`
- 架构说明：`docs/architecture/`
- 文档治理与目录边界：
  - `docs/README.md`（Reference vs Plans 导航与边界）
  - `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`
- 开发工具文档：`docs/development-tools/`
  - Playwright & E2E 指南：`docs/development-tools/e2e-testing-guide.md`
- 计划与进展（易变）：`docs/development-plans/`（完成项归档至 `docs/archive/`）
  - Plan 18: E2E 测试完善计划 — `docs/development-plans/18-e2e-test-improvement-plan.md`
- 版本变更记录：`CHANGELOG.md`

—

## 5. 不做的事（黑名单）
- 脱离契约的私有权限或未声明端点。
- 双数据库/CDC（Neo4j/Kafka 等）造成的数据同步依赖。
- 对外响应出现 snake_case 字段或命名不一致。
- 未按 `// TODO-TEMPORARY:` 标注的临时实现，或超期未回收的临时方案。
- **宿主机直接部署服务或数据库**：本项目强制使用 Docker 容器化部署，严禁在宿主机安装 PostgreSQL、Redis、Temporal 或其他服务组件。所有服务通过 `docker-compose.dev.yml` 管理，确保环境一致性与隔离性。

—

## 6. 临时方案管控（从略，详见 AGENTS.md）
- 仅引用规则：必须 `// TODO-TEMPORARY:` 标注原因/计划/截止日期（不超过一个迭代），建立清单并按期回收。
- 细则以 `AGENTS.md` 为准。

—

## 7. 执行与门禁（索引）
- 开发前必检与常用命令：`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
- 文档边界与同步检查：`.github/workflows/document-sync.yml`
- 代理合规与命名检查：`.github/workflows/agents-compliance.yml`

—

## 8. 版本与更新
- 所有“变更通告/升级说明”请查阅 `CHANGELOG.md` 与 `docs/development-plans/` 对应条目。
- 本文件仅在原则或索引变更时更新，避免与权威文档产生漂移。

—

附注：若本文件与链接内容存在不一致，以链接指向的契约与参考文档为准。
