# Plan 219C2B – 组织域规则迁移

**文档编号**: 219C2B  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 22  
**负责人**: 组织后端团队（安全/架构共审）  

---

## 1. 目标

1. 将组织域既有验证逻辑拆分为链式规则，实现 P0/P1 规则在新框架中的落地。
2. 在 REST/GraphQL 命令入口注入统一验证链，移除散落校验并对齐错误码。
3. 完成组织域规则的单元测试（覆盖率 ≥ 85%）、自测日志与日度同步。

---

## 2. 范围

| 模块/文档 | 工作内容 |
|---|---|
| `internal/organization/validator/organization_*.go` | 实现组织 P0/P1 规则（深度、循环、状态、时态）。 |
| `internal/organization/validator/organization_*_test.go` | 表驱动单测，覆盖正/反场景。 |
| REST/GraphQL handler/service | 注入新验证链并统一错误码/响应。 |
| `logs/219C2/daily-YYYYMMDD.md` | 记录完成度、风险、延迟。 |

---

## 3. 前置条件

- Plan 219C2A 已验收并提供稳定的验证链工厂。
- README/Implementation Inventory 中的规则矩阵冻结完成。
- OpenAPI 契约中相关错误码已存在，或补丁已合并。

---

## 4. 详细任务

### 4.1 规则实现
- 按 README 列表实现 `ORG-DEPTH`, `ORG-CIRC`, `ORG-STATUS`, `ORG-TEMPORAL` 等规则。
- 每个规则包含：输入参数校验、仓储调用（使用 stub 接口）、错误码/Severity 映射。

### 4.2 单元测试
- 每条规则至少 1 个正向 + 2-3 个反向测试样例。
- 使用 stub 仓储避免数据库依赖，记录测试覆盖。
- 生成 `go test -cover` 报告保存到 `logs/219C2/test-Day22.log`。

### 4.3 命令接入
- 在组织 REST handler 添加验证链调用，移除旧的 `utils.Validate*`。
- GraphQL Mutation 目前尚未在命令服务落地，已在 219C2 总计划登记为 219C2C 跟进项，待命令层 GraphQL 上线时复用同一验证链与错误映射。
- 确保错误码、HTTP 状态和响应结构与 OpenAPI 对齐。
- 审计：校验失败时调用 `LogError`，记录 `ruleId`、`severity`。

### 4.4 自测与日志
- 执行 Create/Update Organization REST 自测，记录返回码与审计日志（GraphQL 命令入口改造调整至 219C2C）。
- 更新 `logs/219C2/daily-YYYYMMDD.md`（Day 22）并抄送 219C 总计划。

---

## 5. 交付物

- 新增的组织规则代码与单元测试。
- 命令入口改造提交（REST/GraphQL）。
- `go test` 覆盖率报告（`logs/219C2/test-Day22.log`）。
- 自测日志与日度同步。

---

## 6. 验收标准

- [x] `go test -cover ./internal/organization/validator` ≥ 85%，报告归档（见 `logs/219C2/test-Day22.log`）。
- [x] REST 自测通过，错误码与响应结构一致（详见 `logs/219C2/219C2B-SELF-TEST-REPORT.md`）；GraphQL 命令入口接入验证链调整到 219C2C。
- [x] 审计日志出现正确的 `ruleId` 与 `severity`（详见 `logs/219C2/219C2B-SELF-TEST-REPORT.md`）。
- [x] Day 22 日志提交，列出完成项、风险、缓冲占用（参见 `logs/219C2/daily-20251106.md`）。

---

## 7. 时间安排（Day 22）

| 时间段 | 工作 | 输出 |
|---|---|---|
| 08:30-10:30 | 规则实现（DEPTH/STATUS/CIRC 等） | 组织规则代码 |
| 10:30-12:00 | 单元测试编写与运行 | `go test` 报告初版 |
| 13:00-15:00 | REST/GraphQL 命令接入 | handler/service 改造 |
| 15:00-16:30 | 自测（REST/GraphQL）、审计核对 | 自测日志、审计截图 |
| 16:30-17:30 | 汇总覆盖率、更新 daily log | 测试报告、daily 日志 |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 旧逻辑未完全移除导致重复校验 | 中 | 代码评审对照 checklist；保留临时日志观察重复调用。 |
| 单测覆盖率不足 | 中 | 提前统计覆盖率，如不足及时增加反向样例。 |
| OpenAPI 错误码未对齐 | 高 | 执行前确认契约，缺失立即提交补丁并记录。 |

---

## 9. 度量与追踪

- 覆盖率统计：`go test -cover` 输出归档。
- REST/GraphQL 自测结果记录于 `logs/219C2/validation.log`。
- 日度日志同步至 219C 主计划会议。
