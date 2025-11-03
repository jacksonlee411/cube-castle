# Plan 203 Phase 2 启动通知（草案）

**发送日期**: 2025-11-03  
**收件人**: 后端团队、前端团队、QA、架构组、DevOps、PM  
**主题**: 📢 Plan 203 Phase 2 启动准备（2025-11-13 预定）

各位团队成员：

Plan 214 Phase1 基线萃取已于 2025-11-03 完成并签字，数据库基线与 Goose/Atlas 工具链现已稳定。为确保 Plan 203 Phase 2 能在 2025-11-13 顺利启动，请按照以下事项准备：

1. **环境与依赖**
   - 统一使用容器化 PostgreSQL/Redis；无须额外迁移操作。
   - 复用最新基线文件：`database/schema.sql`、`database/schema/schema-inspect.hcl`、`database/migrations/20251106000000_base_schema.sql`。
   - 若需生成增量迁移，请使用仓库根目录的 `bin/atlas`（离线编译版）。

2. **资源冻结窗口**
   - 后端（command/query）& 前端团队：确认 2025-11-10~2025-11-15 期间具备全量开发资源。
   - QA：于 2025-11-15 前准备 Phase 2 集成测试用例草稿。

3. **任务准备**
   - 后端团队：梳理 workforce 模块 API 契约（命令 REST / 查询 GraphQL），待 11-07 架构评审确认。
   - 前端团队：预先校对共享类型、静态资源方案。
   - DevOps：在 CI 侧启用最新 Goose round-trip + `go test` 流程（已提交至 `ops-scripts-quality.yml`）。

4. **会议安排**
   - 2025-11-08 16:00：跨团队同步会（确认需求拆分、测试策略）。
   - 2025-11-12 09:00：启动前准备检查（环境、数据、权限）。

如有疑问，请在 #plan-203-phase2 频道反馈。

—— PM Office
