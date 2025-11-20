# Plan 259 - 协议策略复审

文档编号: 259  
标题: 协议策略复审（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
状态: ✅ 已完成（2025-11-20，restBusinessGetCount=0，PLAN259_BUSINESS_GET_THRESHOLD=0；CI Run ID [19537850179](https://github.com/jacksonlee411/cube-castle/actions/runs/19537850179)）  
关联计划: 202、203、CQRS 与契约双栈（REST/GraphQL）

---

## 1. 目标
- 复审 REST/GraphQL 在本域内的边界、适用场景与演进策略，避免策略漂移；
- 给出与权限/可观测性/测试资产的对齐建议。

## 2. 交付物
- 复审报告（索引 202/203 与 docs/api/*，不复制正文）；
- 决议清单与后续工作项（指向 25x/24x 子计划）。

## 3. 验收标准
- 结论明确、引用清晰；
- 影响面与迁移计划可操作（分配到具体子计划）。

---

维护者: 架构组（跨团队评审）

---

## 4. 一致性承诺与越权保护（与 AGENTS.md 对齐）
- 最高约束：严格遵循“命令=REST、查询=GraphQL、单一数据源 PostgreSQL、迁移即真源、权限以 OpenAPI 为唯一事实来源”。复审与结论不得突破该边界。
- 协议统一的前置条件：若未来评估需从“双栈”变更为“单栈”，必须先提交并通过 AGENTS.md 的变更评审；在此之前，任何“把查询迁入 REST”或“把命令迁入 GraphQL”的实施请求均属越权，必须驳回。
- 事实来源：本计划仅引用 docs/api/openapi.yaml、docs/api/schema.graphql、database/migrations/ 与既有门禁产物（Plan 255/256/258），不引入第二事实来源。

## 5. 输入与依赖（门禁与证据）
- Plan 255（CQRS 分层与门禁）：检测“前端禁直连/端口/命令查询分离”、后端 depguard 与 tagliatelle。以其报告作为“查询仅在 GraphQL”的守卫证据。
- Plan 256（契约 SSoT 生成/校验）：以生成/校验产物作为“契约变更真实发生”的唯一证据与回归基线。
- Plan 258（契约漂移校验门禁）：用于 OpenAPI ↔ GraphQL 字段/类型/描述/可空一致性报告；本计划新增“权限映射一致性”校验作为扩展项。
- 权限一致性校验归属：复用 Plan 252 的权限契约校验器（scripts/quality/auth-permission-contract-validator.js）生成映射与用量报告；Plan 258 的工作流读取其产物并在“契约漂移门禁”中汇总判定。避免重复实现与脚本分叉。
- 证据归档：本计划产出统一落盘到 logs/plan259/**；执行登记在 docs/development-plans/215-phase2-execution-log.md 中新增记录。

## 6. 现状发现（证据索引）
- （更新 2025‑11‑17）业务查询在 REST 与 GraphQL 的重复已清零（T4 完成）  
  - REST：`GET /api/v1/positions/{code}/assignments` 已移除（OpenAPI 仅保留写路径）  
  - GraphQL：保留 `positionAssignments/assignments` 查询（docs/api/schema.graphql:166、190）
- 权限 scope 不一致（REST vs GraphQL）
  - GraphQL 运行时映射：positionAssignments/assignments → position:read（cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json:12、14）
  - REST 契约：/positions/{code}/assignments GET → position:assignments:read（docs/api/openapi.yaml:1984）
- GraphQL 层未暴露 Mutation，符合“查询=GraphQL”（docs/api/schema.graphql:11~13,42）——正向信号
- OpenAPI 顶部已明确“命令=REST、查询=GraphQL”（docs/api/openapi.yaml:15~20）——正向信号

## 7. 评审结论（默认保持双栈；治理重复与权限一致性）
- 结论 A（默认）：保持“命令=REST、查询=GraphQL”的混合协议策略不变。
- 结论 B（去重）：将“业务查询类 REST GET 端点”清零或收敛到“白名单例外”（仅限 /auth 与 /operational 等非业务查询），严禁与 GraphQL 重复。
- 结论 C（权限一致）：对齐 GraphQL 与 OpenAPI 的权限契约，禁止 scope 名称/粒度漂移；以 OpenAPI 为唯一事实来源。

## 8. 决议与落地项
8.1 业务查询端点去重（REST → GraphQL）
- 决议：将 /api/v1/positions/{code}/assignments GET 标注为“兼容期（Deprecated）”，迁移到 GraphQL positionAssignments/assignments。
- 迁移窗口：// TODO-TEMPORARY(2025-12-15): 完成前端调用迁移与测试证据沉淀（不超过一迭代）
- 回收动作（后续子计划执行，非本文件实施）：
 - 在 OpenAPI 标注 deprecated: true，并添加 Sunset/Link 头部示例与迁移指南（不在本次提交直接改契约）。
  - 迁移指南：`docs/migrations/positions-assignments-to-graphql.md`（前端/集成调用示例、过滤器映射、验收标准与回滚）
  - 前端统一通过领域 API 门面访问 GraphQL（与 Plan 257 保持一致）；移除直接 REST 查询。
  - CI 门禁接入“REST 业务查询白名单”规则（Plan 255 规则扩展）。
  - 在 CHANGELOG.md 发布“弃用通告 + 迁移指南 + 回收时间表”，并在 12. 关联处纳入索引。

8.2 权限契约一致性（OpenAPI ↔ GraphQL）
- 首选方案（推荐，默认）：将 GraphQL 的 positionAssignments/assignments 的权限统一为 position:assignments:read，与 OpenAPI 一致。
  - 子任务：更新 GraphQL 权限映射生成（cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json 的生成源）、更新注释与相关测试。
  - 门禁：在 Plan 258 的契约漂移校验中新增“权限映射一致性”检查（OpenAPI x‑scopes ↔ GraphQL 权限映射）。
- 替代方案（仅当必须保留 position:read）：在 OpenAPI 中明确声明 position:read ⊇ position:assignments:read 的包含关系，并同步到 GraphQL 权限映射说明；此方案需额外治理复杂度，非默认路径。
  - 启用阈值（全部满足方可启用，且需在 215 登记例外与回收期 ≤ 一迭代）：
    1) 历史客户端强绑定 position:read 且迁移窗口不足；
    2) Plan 252/258 的“权限一致性”门禁已接入并出具通过报告；
    3) 已制定明确的回收时间表与变更公告（CHANGELOG），并配置灰度开关可随时回退。

8.3 评审口径与量化阈值
- 业务 GET 白名单（非业务，仅用于认证与运维）：
  - `/.well-known/jwks.json`
  - `/api/v1/operational/**`
  - `/auth/**`
- 业务 GET 判定公式：方法=GET 且 路径以`/api/v1/`开头 且 不在上述白名单集合 → 判定为“业务 GET”。“业务 GET（REST）”目标清零（应为 0）。
- 权限一致性：OpenAPI ↔ GraphQL 权限映射一致率 = 100%（无例外）。
- 门禁通过：guard‑plan255、generate‑contracts/verify‑contracts（Plan 256）、guard‑plan258（含权限一致性扩展）全部通过。

8.4 证据与登记
- 扫描与差异矩阵：
  - reports/plan259/protocol-duplication-matrix.json（REST 业务 GET ↔ GraphQL Query）
  - reports/plan259/permission-mapping-diff.json（OpenAPI x‑scopes ↔ GraphQL 映射）
- 门禁日志归档：logs/plan259/**（含 architecture-validator、verify-contracts、guard‑plan258 扩展报告）
- 执行登记：docs/development-plans/215-phase2-execution-log.md 中新增“Plan 259 执行与证据”小节。

8.5 工具链与命令入口（可直接执行）
- 权限契约与映射（复用 Plan 252 脚本，输出同时供 Plan 258 使用）
  - `node scripts/quality/auth-permission-contract-validator.js --openapi docs/api/openapi.yaml --graphql docs/api/schema.graphql --out reports/permissions`
  - 产物（示例）：`reports/permissions/*.json`、运行时映射 `cmd/hrms-server/query/internal/auth/generated/graphql-permissions.json`
- 协议重复矩阵（REST 业务 GET ↔ GraphQL Query）
  - 推荐新增脚本：`scripts/quality/protocol-duplication-matrix.js`（输入：OpenAPI/GraphQL/白名单；输出：reports/plan259/protocol-duplication-matrix.json）
  - 临时执行占位（// TODO-TEMPORARY(2025-11-25) 架构组补充脚本并接入 Make 入口）
- Make 目标（必须，作为一键入口）：`make guard-plan259`
  - 行为：聚合执行权限契约校验与协议重复矩阵生成 → 产物归档到 `logs/plan259/**` 与 `reports/plan259/**` → 控制台打印摘要（失败时非零退出）
  - 验收：在 215 执行日志中登记“首次运行时间、输出路径、摘要结论与后续动作”

## 9. 实施计划（子任务分解）
- 259‑T1（扫描与报告，零副作用）
  - 产出：REST 业务 GET 清单与白名单对比报告（reports/plan259/protocol-duplication-matrix.json）
  - 依赖：Plan 255 的 architecture‑validator 扫描能力
  - Owner：架构（scripts）+ CI；里程碑：2025‑11‑25（提交脚本与首次报告）
- 259‑T2（权限一致性校验规则接入）
  - 产出：guard‑plan258 扩展“权限映射一致性”检查；首次运行差异报告落盘
  - Owner：CI/DevOps；里程碑：2025‑11‑27（工作流合入，CI 软门禁）
- 259‑T3（GraphQL 权限映射调整——推荐路径）
  - 变更范围：仅更新 GraphQL 权限映射生成逻辑与注释；不改字段/类型
  - 回滚：保留切换开关（DEV_MODE 可回退到旧映射），并在 215 登记
  - Owner：后端；里程碑：2025‑12‑01（灰度启用），2025‑12‑03（默认启用）
- 259‑T4（REST 业务查询端点废止流程）
  - 变更步骤：OpenAPI 标注 deprecated → 发布迁移公告（CHANGELOG）→ 前端迁移 → CI 门禁从软到硬 → 移除端点
  - 回滚：在废止窗口内保留灰度例外白名单（基于路径前缀与租户）
  - Owner：前端（迁移）+ 文档（公告）+ 后端（契约标注）+ CI（门禁）；里程碑：2025‑12‑15（迁移完成/完成证据），2025‑12‑16（硬门禁），2025‑12‑20（端点移除）
 - 执行顺序与依赖：T1 → T2 → T3 → T4（严格顺序；T3 需 T2 门禁已启用；T4 需 T3 权限一致性已生效）

## 10. 风险与回滚
- 风险：前端仍存在对 REST 查询端点的隐性依赖
  - 缓解：先门禁软警告+报告；迁移完成后再切硬门禁；提供 Unauthenticated/Authenticated GraphQL 客户端统一封装
- 风险：权限映射调整导致授权回退
  - 缓解：灰度开关与审计日志核对（按租户/用户/操作维度对比前后授权命中率）
- 回滚方案：所有变更均通过 Make/CI 工作流受控发布，提供“回滚到上一个版本的映射与门禁配置”的操作手册；数据不涉及持久化变更，无需 DB 回滚

## 11. 不做的事（本计划明确排除）
- 不直接修改 OpenAPI/GraphQL 契约字段与路径（仅提出决议与迁移步骤，由对应子计划实施）
- 不在本计划内推进协议统一为单栈（如需，先行变更 AGENTS.md）
- 不涉及 Docker Compose/容器端口/数据库迁移变更（遵守 AGENTS：容器端口不改映射、迁移即真源）

## 12. 关联与索引（唯一事实来源）
- 契约：docs/api/openapi.yaml、docs/api/schema.graphql
- 门禁：Plan 255（docs/development-plans/255-cqrs-separation-enforcement.md）
- SSoT 生成与校验：Plan 256（docs/development-plans/256-contract-ssot-generation-pipeline.md）
- 契约漂移：Plan 258（docs/development-plans/258-contract-drift-validation-gate.md）
- 执行登记：docs/development-plans/215-phase2-execution-log.md
- 变更通告：CHANGELOG.md
- 原则索引：AGENTS.md
- 子计划：259A（`docs/archive/development-plans/259A-protocol-duplication-and-whitelist-hardening.md`）
- 迁移指南：docs/migrations/positions-assignments-to-graphql.md
