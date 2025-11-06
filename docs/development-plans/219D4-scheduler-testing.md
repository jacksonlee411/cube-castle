# Plan 219D4 – Scheduler / Temporal 测试与故障注入

**文档编号**: 219D4  
**关联路线图**: Plan 219 → 219D  
**依赖子计划**: 219D1-219D3 输出  
**目标周期**: Week 4 Day 22（与 219D3 并行）  
**负责人**: 后端团队（测试 Owner）
**评审状态**: ✅ 评审通过

---

## 1. 目标

1. 为迁移后的 Scheduler/Temporal 逻辑编写单元测试、集成测试与故障注入脚本，验证行为等价。
2. 在 sandbox 环境跑通 position version 激活、timeline 修复等关键 workflow，并记录指标/日志供 219E 复用。
3. 输出验证 checklist，覆盖正常路径、重试、告警触发、回退步骤。

---

## 2. 范围

| 测试类型 | 内容 |
|-----------|------|
| 单元测试 | 使用 Temporal SDK 测试工具模拟 workflow/activity、断言 retry/backoff/错误处理 |
| 集成测试 | Docker Compose + make 脚本跑真实 workflow，校验数据库状态、队列、指标 |
| 故障注入 | 模拟 activity 超时/失败、队列堵塞，验证告警与自动恢复 |

不包含：监控配置改动（219D3）与文档发布（219D5）。

---

## 3. 详细任务

1. **测试基线建立**
   - 根据 219D1 迁移清单列出 workflow 列表、触发入口、预期事件。
   - 编写测试矩阵：场景、输入、期望输出、相关指标/日志检查点。

2. **单元测试开发**
   - 在 `internal/organization/scheduler/...` 下新增 `_test.go` 文件，利用 Temporal Test WorkflowEnvironment。
   - 覆盖：成功路径、activity 错误、重试策略、超时处理。

3. **集成测试与脚本**
   - 扩展 `tests/` 或 `cmd/hrms-server/command` 下现有集成测试，或新增 `tests/scheduler/`，通过 make 目标触发。
   - 在 sandbox 运行实际 workflow，记录事件时间线、数据库/日志快照。

4. **故障注入与告警联动**
   - 触发 activity 超时、手动关闭 worker 或注入错误，观察重试与告警（与 219D3 输出对齐）。
   - 记录告警触发时间、恢复步骤。

5. **验证清单与报告**
   - 生成 checklist：workflow 名称、触发方式、期望行为、验证指标、日志位置。
   - 输出测试报告，供 219E-E2E 验收引用。

---

## 4. 验收标准

- [ ] 单元测试、集成测试全部通过并纳入 CI（`make test` / `make test-integration`）。
- [ ] 故障注入脚本能稳定复现告警与恢复流程，有截图/日志佐证。
- [ ] 验证清单与报告存档在 `tests/` 或 `docs/development-plans/219D4-scheduler-testing.md` 附录。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| Temporal 测试环境搭建复杂 | 中 | 复用现有 sandbox 模板；必要时编写启动脚本并记录步骤 |
| 故障注入影响他人任务 | 中 | 在隔离环境执行，提前通知平台团队 |
| 测试范围不足影响 219E | 高 | 与 219E Owner 对齐测试矩阵，确保覆盖入口与回退 | 

---

## 6. 交付物

- 单元/集成测试代码与脚本。
- 故障注入说明与日志。
- 测试报告与 checklist。
