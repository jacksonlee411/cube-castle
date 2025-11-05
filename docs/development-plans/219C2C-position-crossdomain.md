# Plan 219C2C – 职位与跨域规则落地

**文档编号**: 219C2C  
**上级计划**: [219C2 – Business Validator 框架扩展](./219C2-validator-framework.md)  
**目标周期**: Week 4 Day 23  
**负责人**: 组织后端团队（跨域协作）  

---

## 1. 目标

1. 实现 Position、Assignment 及跨域规则（P0/P1），并在统一验证链中接入。
2. 确保 REST/GraphQL 命令使用相同验证链，事务语义与审计不漂移。
3. 补齐跨域依赖清单与自测证据，为 Job Catalog 扩展奠定基础。

---

## 2. 范围

| 模块/文档 | 工作内容 |
|---|---|
| `internal/organization/validator/position_*.go` | 实现 Position 相关规则（POS-ORG、POS-HEADCOUNT、POS-JC-LINK）。 |
| `internal/organization/validator/assignment_*.go` | 实现 Assignment 基础规则（ASSIGN-STATE、ASSIGN-FTE），跨域激活等。 |
| 命令入口（REST/GraphQL） | 接入验证链，保证事务回滚与审计。 |
| `logs/219C2/daily-YYYYMMDD.md` | 日度同步（Day 23）。 |
| 跨域依赖清单 | `logs/219C2/cross-domain-deps.md` 列出仓储/服务依赖、Mock/stub 情况。 |

---

## 3. 前置条件

- Plan 219C2B 验收完成，组织链稳定。
- Position/Assignment 服务具备可注入验证链的接口。
- 跨域仓储（Hierarchy、JobCatalog、Assignment）提供 stub 或 mock。

---

## 4. 详细任务

### 4.1 Position 规则
- 实现 `POS-ORG`, `POS-HEADCOUNT`, `POS-JC-LINK` 等规则。
- 单元测试覆盖正/反场景，并验证错误码/Severity。

### 4.2 Assignment 与跨域规则
- 实现 `ASSIGN-STATE`, `ASSIGN-FTE`, `CROSS-ACTIVE` 等规则。
- 若需额外规则，提出 219E 或 Plan 调整申请。

### 4.3 命令接入
- 在 Position/Assignment REST/GraphQL handler 中引入统一验证链。
- 保留事务回滚语义，确保验证失败不触发仓储写入。
- 审计日志记录 `ruleId`、`severity`、`payload`。

### 4.4 自测与依赖确认
- 运行 Fill/TransferPosition、UpdateAssignment 等命令自测，记录 REST/GraphQL 返回。
- 汇总跨域依赖：需要的仓储接口、mock/stub 状态、后续任务。
- 更新 `logs/219C2/validation.log` 与 `daily-YYYYMMDD.md`。

---

## 5. 交付物

- Position/Assignment 规则代码与单测。
- REST/GraphQL 命令接入改造。
- 自测日志、跨域依赖清单。
- Day 23 日度同步。

---

## 6. 验收标准

- [ ] `go test ./internal/organization/validator -run TestPosition -run TestAssignment` 全部通过。
- [ ] 关键命令（Fill/TransferPosition 等）自测通过，错误码一致。
- [ ] 审计日志包含正确的 `ruleId`/`severity`。
- [ ] 跨域依赖清单提交并获相关团队确认。
- [ ] Day 23 日志更新并提交。

---

## 7. 时间安排（Day 23）

| 时间段 | 工作 | 输出 |
|---|---|---|
| 08:30-10:30 | Position 规则实现与单测 | Position 文件与测试 |
| 10:30-12:00 | Assignment & 跨域规则实现 | Assignment 文件与测试 |
| 13:00-15:00 | REST/GraphQL 命令接入 | handler/service 改造 |
| 15:00-16:30 | 关键命令自测、审计校验 | 自测日志、审计截图 |
| 16:30-17:30 | 更新依赖清单、daily log | `cross-domain-deps.md`、日记 |

---

## 8. 风险与缓解

| 风险 | 影响 | 缓解 |
|---|---|---|
| 跨域仓储接口缺失 | 高 | 提前与数据团队确认，必要时 mock；若当天无法解决，提交风控记录。 |
| 状态流转逻辑复杂 | 中 | 与业务负责人确认用例；添加更多反向测试。 |
| GraphQL 与 REST 行为不一致 | 中 | 自测时双通道验证，保持记录。 |

---

## 9. 度量与追踪

- 单测结果记录在 `logs/219C2/test-Day23.log`。
- 自测输出归档到 `logs/219C2/validation.log`。
- `cross-domain-deps.md` 随计划更新，供 219C2D/219E 参考。

