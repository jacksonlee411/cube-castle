# API 契约检查清单

> 参考：`docs/api/openapi.yaml`、`docs/api/schema.graphql`、`CLAUDE.md` §3

- [ ] REST 路径参数统一 `{code}`，字段使用 camelCase，与 DTO 标签一致。
- [ ] 变更前先更新 `docs/api/openapi.yaml` 或 `docs/api/schema.graphql` 并通过审阅；实现阶段引用契约版本号。
- [ ] OpenAPI `security` 块声明 scope，GraphQL 字段注明权限/owner，并在 README 中引用。
- [ ] 任何 breaking change（字段删除/类型修改）在 Plan 文档与 `CHANGELOG.md` 记录回滚方案。
- [ ] 契约样例与 `internal/organization/handler`/`resolver` 返回格式一致（状态码、错误码）。
