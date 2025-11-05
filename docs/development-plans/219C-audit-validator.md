# Plan 219C – Audit & Validator 规则收敛

**文档编号**: 219C  
**关联路线图**: Plan 219  
**依赖子计划**: 219A 完成目录迁移  
**拆分子计划**: 219C1/219C2/219C3（见下表）  
**目标周期**: Week 4 Day 19-25（衔接 219B，提前为 219D 提供输入）  
**负责人**: 架构/安全组 + 后端团队

---

## 1. 目标

1. 明确并实现组织聚合的审计日志策略：哪些操作必须审计、如何写入、如何与事件/outbox 对齐。
2. 建立业务验证（validator）框架，统一组织/职位/Job Catalog/Assignment 的规则，并提供测试。
3. 更新 README / 规则清单，保证后续开发遵循统一规范。

---

## 2. 范围

| 子计划 | 关注内容 | 输出 |
|--------|-----------|------|
| [219C1 – 审计事件底座与事务化改造](./219C1-audit-foundation.md) | Audit Logger 模型重构、事务内写入、`requestId`/`correlationId` 贯通 | 审计代码更新、README 审计章节、`go test ./internal/organization/audit` |
| [219C2 – Validator 框架扩展](./219C2-validator-framework.md) | BusinessValidator 抽象、规则矩阵、服务注入、单测 | Validator 代码与测试、README 校验章节 |
| [219C3 – 文档与测试对齐](./219C3-docs-and-tests.md) | 文档同步、计划索引、验收记录、测试证据 | 更新后的 README/参考文档、验收勾选清单 |

本篇为总计划，子计划完成后需在此更新进度与验收记录。

不包含：Assignment 查询（219B 负责）、Temporal 调度（219D 负责）。

---

## 3. 详细任务

### 3.1 审计策略
1. **必审计操作清单**（统一更新 `internal/organization/README.md#audit` 小节，无需新增独立文件）：
   - Organization：Create/Update/Delete、状态变更、层级调整（含 Department）。
   - Position：Create/Version Update/Delete、headcount 调整。
   - Assignment：Fill/Transfer/Vacate、状态转换。
   - Job Catalog：Group/Family/Role/Level 的 Create/Update/Delete/Version。
2. 在 service 层统一调用 audit logger，确保与 outbox 同事务（必要时在事务内写入），并在 README 中标注调用规范。
3. 定义 audit 记录字段：`operation`, `actor`, `tenant`, `entityType`, `entityCode`, `payload`, `timestamp`, `requestID`, `correlationID`。
4. 审计日志存储：复用 `audit` 表，必要时补充索引（`tenant + entityType + entityCode + timestamp`），并配置保留期。
5. 规则文档输出：形成操作 → 审计类型 → 责任人的映射，供代码审查与审计追踪使用。

### 3.2 Validator 框架
1. 定义 `BusinessValidator` 接口：`Validate(ctx, command) error`，支持链式执行。
2. 实现规则（覆盖矩阵统一维护在 `internal/organization/README.md#validators` 小节）：
   - Organization/Department：code 唯一、层级无循环、状态转换限制、部门必须有上级。
   - Position：headcount ≥1、引用的 Job Catalog 与组织有效。
   - Assignment：一职一人、时间区间不重叠、状态转换合法（pending→active→inactive）。
   - Job Catalog：层级依赖存在、生效日期冲突检查。
   - 跨域规则：职位引用的组织/部门必须可用，任职变更同步 headcount 等。
3. 将 validator 注入 service 层，并在关键命令执行前调用；形成规则矩阵（规则→触发命令→严重级别）并在 README 中列出。

### 3.3 文档与脚本
1. 输出 `docs/development-plans/219C-audit-validator.md`（本文件）更新后的规则与说明。
2. 在 `internal/organization/README.md` 中附上审计字段定义、validator 规则表，并在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 相应章节增加引用即可，不新增散落文档。
3. 若涉及配置（如审计保留天数），在 `.env`/config 中说明并同步更新 README。

### 3.4 测试
- 审计 logger unit test：验证字段完整、写入成功、失败回滚。
- Validator unit test：覆盖主要命令路径；使用表驱动测试。
- 集成测试（可选）：执行一个组织创建→审计表写入→查询验证。

---

## 4. 验收标准

- [ ] 审计规则清单完善并落地，所有命令操作均调用 audit logger。
- [ ] Validator 框架实现，关键规则以单元测试覆盖。
- [ ] README / 规则文档更新，说明审计字段、验证流程、配置项。
- [ ] `go test ./internal/organization/audit ./internal/organization/validator` 全部通过。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 漏审操作导致合规风险 | 高 | 规则清单由架构/安全评审；在代码审查中对照清单 |
| Validator 触发性能问题 | 中 | 规则尽量无数据库往返，必要时缓存；提供性能监控 |
| 与 outbox 事务冲突 | 中 | 审计与业务变更同事务；若写入失败需整体回滚 |

---

## 6. 交付物

- Audit & Validator 实现 + 测试
- 审计/验证规则文档
- README / config 更新（必要时）

---

## 7. 进度纪要（最近更新于 2025-11-05T01:54:19Z）

- [x] 219C1 – 审计事件底座与事务化改造：文档归档至 `docs/archive/development-plans/219C1-audit-foundation.md`，测试记录见 `logs/219C1/test.log`。
- [ ] 219C2 – Validator 框架扩展：待补充验收证据与归档。
- [ ] 219C3 – 文档与测试对齐：进行中，待完成统一验收与归档。
