# Plan 259A - 协议重复矩阵与白名单固化

文档编号: 259A  
标题: 协议重复矩阵生成与业务 GET 白名单固化（259 子计划）  
创建日期: 2025-11-16  
版本: v1.0  
父计划: 259（协议策略复审）  
关联: 202、255、256、258、215、AGENTS.md

---

## 1. 目标
- 生成“REST 业务 GET ↔ GraphQL Query”的协议重复矩阵，固化“业务 GET 白名单与判定公式”，形成可门禁化的证据与脚本入口。
- 将扫描与报告收口为一键命令（`make guard-plan259`），并在 215 登记执行与产物。

## 2. 不变的原则与边界（与 AGENTS/202 对齐）
- 协议策略不改变：命令=REST、查询=GraphQL（无 GraphQL Mutation）；仅进行发现与报告，不直接改契约/实现。
- 唯一事实来源：仅引用 docs/api/openapi.yaml、docs/api/schema.graphql；不引入第二事实来源。
- 权限契约真源：OpenAPI scopes 为准；GraphQL 权限由脚本生成映射（go:embed 消费）。
- Docker/DB 红线：不改容器端口映射、不做数据库迁移。

## 3. 白名单与判定公式
- 白名单（非业务，仅认证与运维）：
  - `/.well-known/jwks.json`
  - `/api/v1/operational/**`
  - `/auth/**`
- 业务 GET 判定：方法=GET 且 路径以`/api/v1/`开头 且 不在上述白名单 → 判定为“业务 GET（REST）”。

## 4. 交付物
- 报告：
  - `reports/plan259/protocol-duplication-matrix.json`（包含 businessGetPaths、graphqlQueries、启发式重复映射、阈值与判定结果）
  - `reports/plan259/business-get-list.{txt,json}`
  - `logs/plan259/protocol-duplication-summary-*.txt`
- 一键入口：
  - `make guard-plan259`：聚合“权限契约与映射报告（Plan 252 脚本）+ 协议重复矩阵生成（本计划脚本）”，标准输出摘要，产物落盘到 `logs/plan259/**` 与 `reports/plan259/**`

## 5. 验收标准
- 报告产出完整且可复现（冪等）：
  - business GET 清单与数量与 OpenAPI 一致，白名单生效；
  - GraphQL Query 清单与 schema 一致；
  - 启发式重复映射至少识别出 `/api/v1/positions/{code}/assignments ↔ positionAssignments/assignments`。
- 门禁化准备：
  - `make guard-plan259` 可本地运行通过，非交互生成全部报告并打印摘要；
  - 可配置阈值 `PLAN259_BUSINESS_GET_THRESHOLD`（默认=1，便于过渡；硬门禁期改为 0）。
- 登记：
  - 在 `docs/development-plans/215-phase2-execution-log.md` 登记首次运行与证据路径。

## 6. 实施步骤
1) 新增扫描脚本 `scripts/quality/protocol-duplication-matrix.js`（Node，无三方依赖）：
   - 输入：`--openapi`（默认 docs/api/openapi.yaml）、`--graphql`（默认 docs/api/schema.graphql）、`--whitelist '/.well-known/jwks.json,/api/v1/operational/**,/auth/**'`、`--out`（默认 reports/plan259/protocol-duplication-matrix.json）、`--fail-threshold`（默认 `PLAN259_BUSINESS_GET_THRESHOLD||1`）
   - 行为：解析 OpenAPI 获取 GET 路径 → 过滤白名单 → 生成 business GET 清单与数量；解析 GraphQL Query 列表；提供针对已知路径的启发式重复映射；输出 JSON 与文本摘要；当 business GET 数量>阈值时以非零退出（可软门禁使用）。
2) 在 Makefile 新增目标 `guard-plan259`：
   - 先调用 `node scripts/quality/auth-permission-contract-validator.js ...`（复用 252 产物，供 258 聚合）；
   - 再调用 `node scripts/quality/protocol-duplication-matrix.js ...`；产物落盘；标准输出摘要。
3) 在 215 登记首次运行与报告路径。

## 7. 风险与回滚
- 风险：路径匹配误报/漏报（YAML 缩进/注释噪音）。
  - 缓解：采用“行级解析 + 正则”的保守策略，并在 215 登记一次比对人工核验结果。
- 风险：阈值调整导致意外失败。
  - 缓解：硬门禁前期以软门禁运行（阈值=1），待去重完成再切换为 0。
- 回滚：脚本与 Make 目标为只读型；若需回退，删除目标与脚本即可；不影响运行时或数据库。

## 8. 里程碑与责任人
- 脚本与 Make 目标落地：2025-11-25（Owner：架构/CI）
- 首次报告登记（215）：与上同一批次
- 门禁阈值切换为 0：与 259‑T4 废止窗口同步（参考 259 主计划）

## 9. 关联与索引
- 主计划：`docs/development-plans/259-protocol-strategy-review.md`
- 原则：`AGENTS.md`
- 契约：`docs/api/openapi.yaml`、`docs/api/schema.graphql`
- 权限校验器：`scripts/quality/auth-permission-contract-validator.js`（Plan 252）
- 结果登记：`docs/development-plans/215-phase2-execution-log.md`

