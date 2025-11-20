# PR: Remove REST Business GET for Position Assignments (Plan 259‑T4)

标题: feat!: remove REST business GET /api/v1/positions/{code}/assignments  
关联计划: 259（主）、259A、258、257、215、AGENTS.md  
影响范围: REST 合约与命令服务路由（读→GraphQL）, 测试清理  
状态: ✅ 已完成（2025-11-20，CI Run ID [19537850179](https://github.com/jacksonlee411/cube-castle/actions/runs/19537850179)，business GET=0）

---

## 1. 背景与动机
- 为消除“REST 业务查询”与 GraphQL 查询的双事实来源，遵循 PostgreSQL 原生 CQRS（命令=REST、查询=GraphQL），259‑T4 启动“废止 REST assignments 查询端点”的回收。
- 已完成前置工作：
  - 259A：协议重复矩阵与白名单固化（business GET=0 为目标）；
  - 259‑T3：GraphQL 权限对齐（positionAssignments/assignments → position:assignments:read）；
  - T4-1：OpenAPI 已标注 deprecated + CHANGELOG 公告 + 迁移指南。

---

## 2. 变更内容（SSoT 与实现）
- 合约（SSoT）：删除 `GET /api/v1/positions/{code}/assignments` path（docs/api/openapi.yaml）
- 实现：命令服务路由移除对应 GET（仅保留 POST/PATCH/POST close 写操作）
- 测试：移除对上述 GET 的 handler 单测；写路径单测保留
- 不涉及：数据库迁移/Compose 端口/镜像标签；GraphQL 层不改接口（查询已存在）

---

## 3. 行为变化
- 外部调用：不再支持通过 REST GET 查询职位任职；请改用 GraphQL：
  - `positionAssignments(positionCode, filter, pagination, sorting)`
  - `assignments(organizationCode, positionCode, filter, pagination, sorting)`
- 权限：两者统一为 `position:assignments:read`

---

## 4. 迁移与参考
- 迁移指南：`docs/migrations/positions-assignments-to-graphql.md`
- 过滤器映射：assignmentTypes/status/asOfDate/includeHistorical/includeActingOnly → GraphQL `filter.*`
- 前端建议：统一通过领域 API 门面 + `UnifiedGraphQLClient` 发起查询

---

## 5. 验收与证据
- 259A 协议重复矩阵：`restBusinessGetCount=0`（reports/plan259/protocol-duplication-matrix.json）
- 计划门禁（CI）：
  - plan-258-gates 聚合权限校验 + 259A 矩阵（阈值变量 `PLAN259_BUSINESS_GET_THRESHOLD` ）
  - 本 PR 合并后将变量切为 0（硬门禁）
- 215 执行日志：登记本 PR、CI 运行与阈值切换

---

## 6. 风险与回滚
- 风险：外部仍有 REST 查询依赖
  - 缓解：已有弃用公告+迁移指南；PR 合并即视为完成迁移
- 回滚：若需临时回退，可在紧急 patch 中恢复路由与 OpenAPI path；CI 阈值可暂时设置为 1（软门禁）；不涉及 DB 回滚

---

## 7. 清单
- [x] 删除 OpenAPI GET path（SSoT）
- [x] 移除命令服务路由 GET
- [x] 移除 handler 单测（GET）
- [x] 运行 259A：`make guard-plan259`（business GET=0）
- [x] 合并后：设置 `PLAN259_BUSINESS_GET_THRESHOLD=0` 并触发 plan-258-gates（Run ID: [19537850179](https://github.com/jacksonlee411/cube-castle/actions/runs/19537850179)，evidence logged in `docs/development-plans/215-phase2-execution-log.md`）
