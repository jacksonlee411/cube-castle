# database/migrations

- `20251106000000_base_schema.sql` 是新的基线迁移，包含所有 Schema 定义（含扩展、函数、触发器）。
- 所有新增迁移需要使用 Goose 命名约定 `YYYYMMDDHHMMSS_description.sql`，统一放置在本目录根部。
- `rollback/` 目录只保留历史回滚样例，标记为废弃参考，不再直接执行。
- 开发者应先更新 `database/schema.sql`（唯一事实来源），再通过 `scripts/generate-migration.sh` 或手工编写 Goose 迁移。
- 运行 `make db-migrate-all` 使用 Goose 执行迁移，`make db-rollback-last` 回滚最近一条迁移。
