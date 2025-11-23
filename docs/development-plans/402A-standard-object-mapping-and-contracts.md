# 402A · Standard Object 映射与契约准备

**关联计划**：Plan 400（架构）、Plan 402（总体迁移）、Plan 403（元模型组成）  
**状态**：拟立项  
**依赖**：`docs/api/openapi.yaml`、`docs/api/schema.graphql`、`internal/standardobject/**` skeleton、Plan 201 sqlc 规范  
**日志要求**：`logs/plan402/mapping/*.log`、`logs/plan402/mapping/dec-gap.log`、`logs/plan402/mapping/naming-check.log`

> 目标：在实施任何数据库/服务迁移前，完成 `organization_units` → Standard Object 三表的字段映射、契约定义与兼容层设计，确保“先契约后实现”。

---

## 1. 目标

1. 形成《Standard Object 映射规格》，将 Plan 402 §4.1/4.2 的映射表转化为可执行规范，明确字段来源、转换规则与回滚策略。
2. 在 REST / GraphQL 契约中引入 Standard Object 实体、枚举与权限 scope，并对接 Plan 252/259 守卫。
3. 定义临时兼容策略（视图/Port/Feature Flag 的需求与接口约束），并把具体实现交给 402B/402C；A 阶段仅需产出契约说明和迁移矩阵，确保 402B 能据此落地。

---

## 2. 工作项

### A1 · 文档固化
- 在现有 Plan 400/402 的基础上，补充 `docs/development-plans/402A-standard-object-mapping-spec.md`，只记录字段映射与迁移矩阵（字段→目标表/列、数据转换/验证 SQL、触发器替代、异常/回滚流程及负责人）。
- DEC/OCL 绑定继续以 Plan 400 的 Schema Registry (`standard_object_schemas` + `schema-registry.json`) 为唯一事实来源，不再新增脚本；若发现缺口，在映射规格中登记并创建 hazard list，由 402B 的 Schema Registry 扩展解决。
- 命名一致性依赖现有质量守卫（`scripts/quality/architecture-validator.js` 等），A 阶段仅需记录检查结果，不再自建脚本。

### A2 · 契约补全
- 更新 `docs/api/openapi.yaml`/`docs/api/schema.graphql`，新增 Standard Object 端点、类型、权限 scope，并在 `logs/plan402/mapping/api-contract.log` 记录 `node scripts/quality/contract-checker.js` 与 GraphQL codegen 的输出。
- 通过现有命令 (`npm --prefix frontend run contract:generate`) 同步前端类型，无需新增脚本；把 PBAC 改动同步到 Plan 252/259 与参考手册。

### A3 · 兼容策略说明（设计层）
- 在映射规格中描述需要的兼容视图/Port/Feature Flag（用途、依赖、回滚原则），但将实际 SQL/代码实现放在 402B/402C 执行。
- 记录 Feature Flag 的配置方式（环境变量名、默认值、回滚路径），由 402C 在命令/查询服务中具体落地。
- OCL 约束在 Schema Registry 中维护；A 阶段只需列出需校验的场景，由 402B 的 Schema Registry 扩展与 402C 的校验工具执行。

---

## 3. 交付物

- `docs/development-plans/402A-standard-object-mapping-spec.md`（含字段映射、迁移矩阵、回滚 checklist），并引用 Plan 400 Schema Registry 的 DEC/OCL 信息。
- 兼容策略说明（视图/Port/Feature Flag 需求与回滚原则），供 402B/402C 实施；无需在 A 阶段提供 SQL/代码。
- OpenAPI / GraphQL diff 及 Scope 更新说明，附 `scripts/quality/architecture-validator.js`、`node scripts/quality/contract-checker.js` 的输出。
- `internal/standardobject/adapter/*.go` skeleton + Feature Flag 文档，`go test ./internal/standardobject/...` 日志。
- `logs/plan402/mapping/*.log`（DEC gap、命名检查、视图 explain、评审会议记录）。

---

## 4. 验收标准

1. `docs/development-plans/402A-standard-object-mapping-spec.md` 通过评审并在 `docs/development-plans/00-README.md` 登记；评审记录写入 `logs/plan402/mapping/spec-review.log`。
2. 契约更新通过 `scripts/quality/architecture-validator.js`、`node scripts/quality/contract-checker.js`，日志落在 `logs/plan402/mapping/api-contract.log`。
3. 兼容策略说明清楚视图/Port/Feature Flag 的需求、默认值与回滚流程，并将实现责任转交 402B/402C；无需在 A 阶段提交编译日志。
4. Schema Registry（Plan 400）中针对组织/职位对象的 DEC/OCL 绑定被引用到映射规格中，如发现缺口则记录 hazard list 并在 402B 执行前补齐。

---

## 5. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 契约变更影响现有客户端 | 前端或集成方出现编译/运行错误 | 提前同步 Plan 222 负责人，发布 schema diff 与迁移指南 |
| 视图性能不足 | 迁移期查询退化 | 输出 explain plan 与基准测试，必要时增加索引或物化视图 |
| DEC/OCL 信息缺失 | Schema Registry 无法成为 SSoT | 在 hazard list 中记录缺口，限定回收期限并跟进 Plan 400/403 |
| Feature Flag/回滚策略不明确 | 切换/回退困难 | 在 A3 文档中列出步骤、负责人、日志路径，并提前演练 |

---

完成 402A 后方可启动 402B。未经 402A 验收，禁止对数据库 schema 或仓储进行 Standard Object 相关改动。
