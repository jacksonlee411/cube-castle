# Plan 218 - 结构化日志系统交付路线图

**文档编号**: 218  
**标题**: 结构化日志系统 - 子计划实施路线  
**创建日期**: 2025-11-04  
**分支**: `feature/204-phase2-infrastructure`  
**关联计划**: Plan 216（eventbus）、Plan 217（database）、Plan 215（Phase2 执行日志）

---

## 1. 目标与范围

Plan 218 负责落地统一的结构化日志体系，并推动命令/查询服务及共享组件完成全面迁移。为降低一次性改造风险，整体拆分为 5 个子计划，按“底层基础设施 → HTTP/中间件 → 缓存 → 查询服务 → 收尾”路径推进。

| 子计划 | 范围 | 主要目标 | 时间窗口 |
|--------|------|----------|----------|
| **218A** | 命令服务核心层（repository/service/audit/validator） | 替换 `*log.Logger`，注入结构化字段，保障测试 | Week 3 Day 3–4 |
| **218B** | 命令服务 HTTP 栈（handlers/middleware/authbff） | 统一请求级日志与上下文字段，清理遗留输出 | Week 3 Day 4–5 |
| **218C** | 共享缓存子系统（`internal/cache`） | L1/L2/L3 缓存日志结构化，补充测试与字段规范 | Week 3 Day 5–6 |
| **218D** | 查询服务（repository/resolver/app） | 查询链路结构化日志、补充测试工具与性能基准 | Week 3 Day 6–7 |
| **218E** | 收尾与桥接器清理 | 全局扫描遗留 `log.*`，更新文档、清理兼容层 | Week 3 Day 7 |

---

## 2. 实施路线

1. **完成基础设施准备（Plan 216/217/218A）**  
   - `pkg/logger/` 已交付；218A 聚焦命令服务核心依赖，保证仓储/服务层全面改用结构化日志。

2. **扩展至 HTTP 栈与缓存层（218B/218C）**  
   - 对 handler/middleware 引入统一上下文字段；缓存模块补齐层级字段（L1/L2/L3）、命中/回源事件日志。

3. **覆盖查询服务（218D）**  
   - 替换查询侧 repository/resolver 的日志实现，提供测试 logger、性能基准及端到端场景。

4. **收尾与文档补充（218E）**  
   - 清理 `NewStdLogger` 兼容层，更新 Reference / 计划文档，复核 `go test ./...` 与 lint 流程。

子计划交付序列需保持依赖顺序：218A → 218B → 218C → 218D → 218E。每个子计划完成后更新本路线文档的“进度登记”。

---

## 3. 进度登记表

| 子计划 | 状态 | 备注 |
|--------|------|------|
| 218A | ☑ 已完成 | 命令服务核心仓储/服务/审计/验证改用 `pkg/logger`, `go test ./cmd/hrms-server/command/internal/...` 通过 |
| 218B | ☑ 已完成 | 命令服务 HTTP 栈迁移完成，参考 218B 子计划文档 |
| 218C | ☑ 已完成 | 缓存层结构化日志完成；`go test ./internal/cache` 已在本地验证通过 |
| 218D | ☐ 未开始 / ☐ 进行中 / ☐ 已完成 | |
| 218E | ☐ 未开始 / ☐ 进行中 / ☐ 已完成 | |

（执行团队在子计划完成后需更新状态及备注，例如关键风险、回归结果等。）

---

## 4. 文档索引

- **218A** 命令服务核心层日志迁移 — `docs/development-plans/218A-command-service-core-logger-migration.md`
- **218B** 命令服务 HTTP 栈日志迁移 — `docs/development-plans/218B-command-http-stack-logger-migration.md`
- **218C** 共享缓存日志迁移 — `docs/development-plans/218C-shared-cache-logger-migration.md`
- **218D** 查询服务日志迁移 — `docs/development-plans/218D-query-service-logger-migration.md`
- **218E** 迁移收尾与桥接器清理 — `docs/development-plans/218E-logger-rollout-closure.md`

如需评审意见或历史记录，请参阅：
- `docs/development-plans/218-review-analysis.md`
- 子计划执行日志（完成后可记录于 `docs/development-plans/06-integrated-teams-progress-log.md`）

---

**维护者**: Codex（AI 助手）  
**最后更新**: 2025-11-04
