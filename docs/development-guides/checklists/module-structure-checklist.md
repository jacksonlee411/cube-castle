# 模块结构检查清单

> 参考：`internal/organization/README.md`、`AGENTS.md` 项目结构、`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`

- [ ] 在 `internal/{module}/` 下创建 README，说明聚合边界、迁移需求、测试入口。
- [ ] `api.go` 暴露 `CommandModule`/`QueryModule` 构造器，并对依赖执行 `Validate()`。
- [ ] 命令（handler）与查询（resolver）遵循 CQRS：REST 逻辑放 `handler/`，GraphQL 逻辑放 `resolver/`。
- [ ] `repository/` 仅依赖 `pkg/database`，使用手写 SQL 或已审批的 sqlc（需引用 `docs/archive/plan-216-219/219A-219E-review-analysis.md` 的标准）。
- [ ] `service/` 封装事务逻辑并与 Outbox 集成；禁止 handler 直接访问数据库。
- [ ] 新增目录需在 README 中说明用途，避免平行事实来源。
