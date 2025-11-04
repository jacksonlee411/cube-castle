# Plan 218A - 命令服务核心层结构化日志迁移

**母计划**: Plan 218
**子计划编号**: 218A
**目标窗口**: Week 3 Day 3 – Day 4
**负责人**: 平台 / 基础设施组

---

## 1. 范围
- 仓储层：`cmd/hrms-server/command/internal/repository/**`
- 业务服务层：`cmd/hrms-server/command/internal/services/**`
- 审计/验证：`cmd/hrms-server/command/internal/audit/**`、`.../validators/business.go`
- 相关单元测试与构造函数

子计划聚焦命令服务核心依赖的内部层（repo/service/audit/validator），其余 handler/middleware/cache 由 218B/218C 等子计划覆盖。

---

## 2. 交付目标
1. 将上述模块的构造函数与结构体字段统一替换为 `pkg/logger.Logger`。
2. 落实 `WithFields` 上下文字段（如 `component`, `module`, `repository` 等），按 Info/Warn/Error/Debug 分级输出。
3. 替换历史的 `Printf/Println/Fatal` 调用，消除 `*log.Logger` 依赖。
4. 更新关联测试和伪造对象，使用 `logger.NewNoopLogger()` 或专用测试 logger。
5. 验证 `go test ./cmd/hrms-server/command/internal/...` 与 `go test ./internal/auth` 通过。

---

## 3. 实施步骤
1. **签名调整**：逐个模块修改构造函数参数与结构体字段类型，注入共享 logger。
2. **语义分级**：梳理日志语句，区分调试/提醒/错误，并根据需要补充 `WithFields` 信息。
3. **测试更新**：
   - 重构单测使用的 fake logger。
   - 校验审计/服务相关断言不受影响。
4. **静态质量**：执行 `gofmt`、`go test`；必要时运行 `go vet`、`golangci-lint`（若配置）。

---

## 4. 验收标准
- [ ] 命令服务核心层无 `*log.Logger` 引用（`rg "\*log.Logger"` 验证）。
- [ ] 日志字段包含 `component` 或等效上下文信息。
- [ ] `go test ./cmd/hrms-server/command/internal/...` 100% 通过。
- [ ] 文档与主计划（218）同步记录。

---

## 5. 风险与缓解
| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 依赖层级较多，签名改动会层层传递 | 中 | 先自底向上迁移仓储，再更新服务/调用方 |
| 日志字段遗漏导致后续 handler 接口缺参 | 中 | 制定字段命名约定，代码审查时逐项确认 |
| 测试需注入新的 logger 类型 | 低 | 提供 `newTestLogger()` 帮助函数 |

---

## 6. 交付物
- 更新后的仓储/服务/审计/验证代码
- 测试更新
- 本子计划文档

---

**备注**: 本子计划完成后，命令服务 handler/middleware 层迁移交由 Plan 218B 跟进。
