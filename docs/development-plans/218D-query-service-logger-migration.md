# Plan 218D - 查询服务结构化日志迁移

**母计划**: Plan 218
**子计划编号**: 218D
**时间窗口**: Week 3 Day 6 – Day 7
**责任团队**: 查询服务 / GraphQL 组

---

## 1. 范围
- `cmd/hrms-server/query/internal/repository/**`
- `cmd/hrms-server/query/internal/graphql/**`
- 查询服务 `app` 层辅助模块（若仍有标准库 logger）
- 相关测试（resolver、integration）

---

## 2. 目标
1. Repository、Resolver、App 初始化全部使用 `pkg/logger.Logger`。
2. 分层输出字段：`component=query-repo`, `resolver=position`, `operation` 等。
3. 调整日志级别，确保数据库失败/Error 路径使用 `Errorf`，缓存 miss/回源使用 Info/Debug。
4. 更新测试日志注入；确保 `newTestLogger()` 工具函数覆盖所有用例。
5. 通过 `go test ./cmd/hrms-server/query/internal/...`。

---

## 3. 实施步骤
1. 修改 repository/resolver 构造函数、结构体字段、调用方（`app.Run` 等）。
2. 梳理 `Printf/Println`，替换为结构化日志；注入 `WithFields`（tenant、queryName、resolver）。
3. 更新 GraphQL 中间件在错误场景下提供结构化输出（已在 218A 中完成基础迁移，需复核）。
4. 测试改造：统一使用 `newTestLogger()`，避免 `log.New(io.Discard...)`。
5. `go test` + 关键流程人工验证（GraphiQL / API 请求）。

---

## 4. 验收标准
- [ ] 查询服务代码中不再存在 `*log.Logger` 依赖。
- [ ] Resolver 日志包含 `resolver`、`operation`、`tenant` 等字段。
- [ ] `go test ./cmd/hrms-server/query/internal/...` 通过。
- [ ] 文档与主计划同步更新。

---

## 5. 风险与缓解
| 风险 | 说明 | 缓解 |
|------|------|------|
| GraphQL resolver 数量多，字段易遗漏 | 中 | 制定字段模板，代码审查强制检查 |
| Query repository 依赖缓存模块（218C） | 中 | 与 218C 同步签名变更，优先完成 218C |
| 测试依赖 log 输出 | 低 | 统一测试 logger，必要时通过 hook 捕获日志 |

---

## 6. 交付物
- 更新后的查询服务代码与测试
- 子计划文档（本文件）
- 对 Plan 218 总文档的更新记录

---

**前置依赖**：218A（基础设施）、218C（缓存）完成后执行，以避免重复修改签名。
