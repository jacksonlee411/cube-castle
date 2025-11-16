# Plan 252 - 权限一致性与契约对齐

状态：完成（已归档）  
文档编号: 252  
标题: 权限一致性与契约对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、203、OpenAPI/GraphQL 契约、auth/pbac

—

## 1. 目标
- 以 OpenAPI/GraphQL 为唯一事实来源，对齐 scopes/roles 与接口权限；
- 建立 PBAC 校验清单与回归用例（读侧 GraphQL 中间件 + REST 权限中间件）。

## 2. 交付物
- 权限-契约映射表（仅引用 docs/api/*）；
- PBAC 校验点清单与最小用例；
- 证据登记：logs/plan252/*。
- OpenAPI scopes 注册清单更新（PR 描述具体新增/更名/合并项）；
- GraphQL Query→scope 映射生成制品：`reports/permissions/graphql-query-permissions.json`（由 schema 注释生成，SSoT 衍生物）。

## 3. 验收标准
- 合同字段（security/scope）与中间件校验一致；
- 样例用例覆盖关键查询/命令；
- CI 权限契约检查通过（强制门禁）。

—

维护者: 后端（安全合规协同评审）

—

## 4. 自动化门禁（CI）
- OpenAPI security scopes 一致性=100%：扫描 `docs/api/openapi.yaml`，所有受保护 REST 路径的 `security: - OAuth2ClientCredentials: [scopes...]` 中的每个 scope 必须出现在 `components.securitySchemes.OAuth2ClientCredentials.scopes` 注册表中（不使用 `x-scopes` 扩展）；禁止“未注册即使用”的 scope；允许输出“已注册但未被引用”的信息级提示。
- GraphQL resolver 必经 PBAC：扫描 resolver 代码，校验所有 Query 入口均通过统一授权门面（`permissions.CheckQueryPermission` 或其门面）执行 PBAC 校验；禁止直连数据仓库绕过。
- GraphQL→PBAC 映射覆盖率=100%：基于 `docs/api/schema.graphql` 中的 “Permissions Required: …” 注释生成 Query→scope 映射为唯一事实来源（生成制品落盘）；PBAC 使用该映射（禁止手写常量）。任一 Query 缺少权限标注或映射缺失即失败。
- REST/GraphQL 权限一致性抽样对比：基于映射表对同一业务操作的权限语义进行抽样对比（信息级或门禁，按配置）。
- 临时豁免：必须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注原因/计划/截止日期（不超过一个迭代），并在阶段计划登记；过期即失败。

—

## 5. 评审差异与整改任务清单
- OpenAPI scopes 注册不全（已整改）
  - `position:assignments:read`/`position:assignments:write` 在路径 `security` 中使用，现已在 `components.securitySchemes.OAuth2ClientCredentials.scopes` 注册。
- GraphQL→PBAC 映射缺失（已整改）
  - 已补齐：`hierarchyStatistics`（`org:read:hierarchy`）、`jobFamilyGroups`/`jobFamilies`/`jobRoles`/`jobLevels`（`job-catalog:read`）、`assignments`（`position:read`）、`assignmentHistory`（`position:read:history`）、`assignmentStats`（`position:read:stats`）、`positionAssignmentAudit`（`position:assignments:audit`）。
- 权限命名（已决策）
  - 采纳 `position:assignments:audit` 作为正式权限；已在 OpenAPI 注册，并补齐 PBAC 映射与注释一致性。
- 角色→权限硬编码（收敛中）
  - 权限策略外部化至 BFF/签发域，Query 侧仅基于 Token scopes 授权；临时保留硬编码兜底，已加 `// TODO-TEMPORARY(2025-12-15)` 限期回收。
- devMode 直通（已约束）
  - 生产/CI 默认禁用（DEV_MODE=false）；本地容器显式开启（DEV_MODE=true）。

—

## 6. 落地步骤（先契约后实现）
1) 契约整改：补齐/统一 OpenAPI scopes 注册与路径声明；确认并保留 `position:assignments:audit`；补注 GraphQL 权限注释。
2) 生成映射：基于 schema 注释生成 Query→scope 映射制品（JSON），落盘 `reports/permissions/graphql-query-permissions.json`。
3) 实现对齐：PBAC 映射改为读取生成制品；移除/封存手写映射；角色→权限外部化（后续在 BFF 域）并移除硬编码。
4) 校验与证据：运行校验脚本，生成报告至 `reports/permissions/*`；日志至 `logs/plan252/*`；提交前通过。

—

## 7. 验收标准（可度量）
- OpenAPI security scopes 一致性=100%
- GraphQL→PBAC 映射覆盖率=100%
- Resolver 授权覆盖率=100%
- 生产构建禁用 devMode
- 关键用例（正/负向各≥1）：组织层级、时态历史、职位任职、职类目录、审计历史
- 证据落盘完整（logs/reports 路径）

—

## 8. 脚本接口规范（仅文档，脚本由后续实现）
- CLI 名称（建议）：`node scripts/quality/auth-permission-contract-validator.js`
- 输入参数：`--openapi`、`--graphql`、`--resolver-dirs`、`--out`、`--fail-on`
- 输出制品：openapi-scope-usage.json、openapi-scope-registry.json、graphql-query-permissions.json、resolver-permission-calls.json、summary.txt
- 校验项：未注册引用/映射缺失/授权绕过（阻断）；注册未引用（信息）
- 退出码：0=通过；非0=失败

—

## 9. 风险与回滚
- 一致性优先：发现契约与实现不一致时，先回滚实现或更新契约；禁止引入第二事实来源。
- 临时兼容项：`// TODO-TEMPORARY(YYYY-MM-DD):` 并限期回收；逾期门禁提示。
- devMode：如误入生产构建，立即回滚或强制配置覆盖。

—

附：唯一事实来源与签字
- OpenAPI（REST + 权限 scopes）：`docs/api/openapi.yaml`
- GraphQL（查询权限注释）：`docs/api/schema.graphql`
- 原则与约束：`AGENTS.md`
- 校验器接口规范：`docs/reference/08-PERMISSIONS-CONTRACT-VALIDATOR-SPEC.md`
- 签字纪要：`docs/archive/development-plans/252-signoff-20251115.md`

