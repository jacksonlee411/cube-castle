# Plan 218B - 命令服务 HTTP 栈结构化日志迁移

**母计划**: Plan 218 结构化日志系统
**子计划编号**: 218B
**时间窗口**: Week 3 Day 4 – Day 5
**责任团队**: 平台 / 应用集成组

**当前状态**（2025-11-04 更新）：
- 命令侧 Handler、中间件及 BFF 模块已统一使用结构化 logger，并在请求链路补齐 `requestId` / `tenantId` 等上下文字段。
- REST / GraphQL 鉴权链路完成结构化日志改造，PBAC 检查器增加查询上下文字段；命令服务与查询服务入口均直接注入 `pkg/logger.Logger`，已回收 `NewStdLogger` 兼容层。
- 查询服务 Repository / Resolver 日志改造完成，测试使用 `pkglogger.NewNoopLogger()` 替换 `*log.Logger`，`go test ./cmd/hrms-server/command/... ./cmd/hrms-server/query/...` 通过。
- 结构化性能/告警日志已更新，慢请求与性能阈值以字段方式输出，便于后续指标采集。

**待办事项 / 下一步**
1. 验证 BFF 手动登录/刷新链路的日志字段是否满足运营团队需求，并在 218 总计划验收表中记录验证结论。
2. 在 218E 子计划登记 `pkg/logger.NewStdLogger` 已清理的事实来源，确保后续收敛列表同步更新。
3. 回顾日志字段规范，若需追加查询服务 GraphQL 域的字段字典，补充至 `docs/reference/03-API-AND-TOOLS-GUIDE.md`。

---

## 1. 范围
- Handler：`cmd/hrms-server/command/internal/handlers/**`
- Middleware：`cmd/hrms-server/command/internal/middleware/**`
- Auth BFF：`cmd/hrms-server/command/internal/authbff/**`
- 辅助模块：`cmd/hrms-server/command/internal/utils` 中涉及日志的部分
- Handler 层单测与集成测试

---

## 2. 目标
1. 所有 handler/middleware 构造与结构体使用 `pkg/logger.Logger`。
2. 统一请求级字段（`requestId`, `handler`, `route`, `tenantId` 等）并通过 `WithFields` 注入。
3. 调整日志级别：
   - 成功路径使用 `Infof`
   - 预期异常/校验失败使用 `Warnf`
   - 真正的错误使用 `Errorf`
4. 中间件链路（性能、限流、DevTools 等）输出结构化指标/调试信息。
5. 确保 DevTools / BFF 在开发模式下仍保留原有功能和输出，但转为 JSON 日志。

---

## 3. 实施步骤
1. **接口签名迁移**：逐个 handler/middleware 更新构造函数，更新调用栈（包含 `main.go`）。
2. **字段注入**：在初始化时传入公共字段，如 `component=handler`、`handler=organization`、`middleware=rate-limit`。
3. **日志梳理**：替换 `Printf/Println/Fatal`，清理 emoji 字符（若保留需说明），确保消息遵循“动词 + 关键信息”。
4. **测试更新**：重写 handler 单测使用新的测试 logger；必要时断言日志字段或行为。
5. **静态检查**：`go test ./cmd/hrms-server/command/internal/handlers/...` 与 `./internal/middleware` 通过。

---

## 4. 验收标准
- [x] 命令服务 HTTP 栈无 `*log.Logger` / `log.Printf` 引用。
- [x] 每个 handler 在入口处通过 `WithFields` 注入 `route`/`module` 字段。
- [x] 日志级别符合约定；关键错误路径包含上下文字段（tenant、requestId 等）。
- [x] DevTools/BFF 手动验证通过（生成令牌、状态检查）。
- [x] 对应单测与 `go test` 全部通过。

---

## 5. 风险 & 缓解
| 风险 | 说明 | 缓解 |
|------|------|------|
| handler 签名改动较多，引发编译链路错误 | 高 | 采用“模块逐个迁移”策略，每次迁移后立即编译测试 |
| 中间件日志字段不统一 | 中 | 在文档中定义字段规范，代码评审时校验 |
| DevTools 调试输出依赖文本日志 | 低 | 提供格式化消息并进行人工验证 |

---

## 6. 交付物
- 更新后的 handler/middleware/authbff 代码
- 测试改造
- 子计划文档（本文件）

---

**后续依赖**：218C (缓存模块) 与 218D (查询服务) 将继续迁移剩余组件；218E 在所有子计划完成后收敛桥接器与文档。
