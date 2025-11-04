# Plan 218C - 共享缓存子系统结构化日志迁移

**母计划**: Plan 218
**子计划编号**: 218C
**时间窗口**: Week 3 Day 5 – Day 6
**责任团队**: 平台缓存与性能组

---

## 1. 范围
- `internal/cache/unified_cache_manager.go`
- `internal/cache/cache_events.go`
- 缓存相关构造/工具函数及单测

---

## 2. 目标
1. 缓存管理器与事件处理器使用 `pkg/logger.Logger`，摒弃标准库 logger。
2. 日志字段规范：`component=cache`, `layer=L1/L2/L3`, `event=hit/miss/refresh` 等。
3. 将命中/回填/一致性检查等日志细化为 Info/Warn/Error，调试信息可使用 Debug（可通过 LOG_LEVEL 控制）。
4. 更新缓存测试以适配新的 logger 注入方式，保证覆盖命中/刷新/一致性两路径。

---

## 3. 实施步骤
1. 调整构造函数签名（`NewUnifiedCacheManager`、`NewSmartCacheUpdater`、`NewConsistencyChecker` 等）。
2. 在关键流程调用 `WithFields` 补充 `cacheKey`, `tenantId`, `sourceLayer` 等上下文。
3. 审核所有 `Printf`/`Println` 并替换为 `Infof`/`Warnf`/`Errorf`；谨慎保留中文/emoji 文案并与团队确认。
4. 更新测试使用 `logger.NewNoopLogger` 或定制测试 logger，确保断言不依赖旧实现。
5. 运行 `go test ./internal/cache`。

---

## 4. 验收标准
- [x] `internal/cache` 目录无 `*log.Logger` 引用。
- [x] 日志结构包含层级及关键指标字段。
- [x] `go test ./internal/cache` 通过（本地环境验证）。
- [x] 与主计划（Plan 218）保持文档同步。

---

## 5. 风险与对策
| 风险 | 影响 | 对策 |
|------|------|------|
| 缓存日志字段过多导致输出噪音 | 中 | 设定最小字段集合，额外信息使用 Debug 级别 |
| 构造函数签名变更导致调用方修改较多 | 中 | 规划统一注入（在 218A/218B 中同步更新）、分 PR/commit 提交 |
| 测试依赖 log 输出 | 低 | 通过测试 logger 或 mock，实现无副作用替换 |

---

## 6. 交付物
- 更新后的缓存代码与测试
- 日志字段规范说明
- 本子计划文档

---

## 7. 进度记录
- [2025-11-04] 结构化日志改造代码与测试更新完成；测试验证需待本地环境执行。
- [2025-11-04] `go test ./internal/cache` 已在本地通过，Plan 218C 满足关闭条件。

---

**注意**：完成后需通知 218B / 218D 负责团队同步调用签名变更。
