# Plan 252 - 权限一致性与契约对齐

文档编号: 252  
标题: 权限一致性与契约对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202、203、OpenAPI/GraphQL 契约、auth/pbac

---

## 1. 目标
- 以 OpenAPI/GraphQL 为唯一事实来源，对齐 scopes/roles 与接口权限；
- 建立 PBAC 校验清单与回归用例（读侧 GraphQL 中间件 + REST 权限中间件）。

## 2. 交付物
- 权限-契约映射表（仅引用 docs/api/*）；
- PBAC 校验点清单与最小用例；
- 证据登记：logs/plan252/*。

## 3. 验收标准
- 合同字段（security/scope）与中间件校验一致；
- 样例用例覆盖关键查询/命令；
- CI 权限契约检查通过（强制门禁）。

---

维护者: 后端（安全合规协同评审）

---

## 4. 自动化门禁（CI）
- OpenAPI security scopes 一致性=100%：扫描 `docs/api/openapi.yaml`，所有受保护 REST 路径的 `security: - OAuth2ClientCredentials: [scopes...]` 中的每个 scope 必须出现在 `components.securitySchemes.OAuth2ClientCredentials.scopes` 注册表中（不使用 `x-scopes` 扩展）；禁止“未注册即使用”的 scope；允许输出“已注册但未被引用”的信息级提示。
- GraphQL resolver 必经 PBAC：扫描 resolver 代码，校验所有 Query 入口均通过统一授权门面（`permissions.CheckQueryPermission` 或其门面）执行 PBAC 校验；禁止直连数据仓库绕过。
- GraphQL→PBAC 映射覆盖率=100%：基于 `docs/api/schema.graphql` 中的 “Permissions Required: …” 注释生成 Query→scope 映射为唯一事实来源（生成制品落盘）；PBAC 使用该映射（禁止手写常量）。任一 Query 缺少权限标注或映射缺失即失败。
- REST/GraphQL 权限一致性抽样对比：基于映射表对同一业务操作的权限语义进行抽样对比（信息级或门禁，按配置）。
- 临时豁免：必须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注原因/计划/截止日期（不超过一个迭代），并在阶段计划登记；过期即失败。

---

## 5. 评审差异与整改任务清单
- OpenAPI scopes 注册不全（阻断）
  - `position:assignments:read`/`position:assignments:write` 在路径 `security` 中使用，但未在 `components.securitySchemes.OAuth2ClientCredentials.scopes` 注册。需：补齐注册，或改用已注册且语义一致的 scope 并同步路径声明。
- GraphQL→PBAC 映射缺失（阻断）
  - 缺少以下 Query 的权限映射：`hierarchyStatistics`（应为 `org:read:hierarchy`）、`jobFamilyGroups`/`jobFamilies`/`jobRoles`/`jobLevels`（应为 `job-catalog:read`）、`assignments`（`position:read`）、`assignmentHistory`（`position:read:history`）、`assignmentStats`（`position:read:stats`）、`positionAssignmentAudit`（待确认，见下条）。
- 权限命名待确认（高）
  - `position:assignments:audit` 是否为正式权限？若采纳，需在 OpenAPI 注册；若不采纳，应统一改为既有权限（如 `position:read:history`）并更新 GraphQL 注释与实现。
- 角色→权限硬编码（高）
  - 现有 `RolePermissions` 为内部常量（如 `READ_ORGANIZATION`），与 PBAC scope 不一致；权限策略需外部化（配置/数据库），禁止长期硬编码。临时保留需加 `// TODO-TEMPORARY(YYYY-MM-DD):` 并设回收期。
- devMode 直通策略（中）
  - 开发模式下 ADMIN/admin 直通允许保留，但必须确保生产禁用；CI 可抽样检测 devMode=false 构建产物不包含放行逻辑的有效路径。

整改完成判据（增强版）
- OpenAPI scope 引用→注册校验：0 未注册引用；未使用注册项仅信息提示。
- GraphQL 映射覆盖率：100%（缺失列出 Query 名与 schema 行号）。
- PBAC 入口覆盖率：100%（所有 resolver 查询均经授权门面）。
- 临时项合规：存在 `// TODO-TEMPORARY(YYYY-MM-DD):` 且未超期。
- 证据：`logs/plan252/*` 与 `reports/permissions/*` 完整可复现。

---

## 6. 落地步骤（先契约后实现）
1) 契约整改
   - 补齐/统一 OpenAPI scopes 注册与路径声明；确认 `position:assignments:audit` 的采用与否。
   - 在 `docs/api/schema.graphql` 保持/补充每个 Query 的 “Permissions Required: …” 注释。
2) 生成映射
   - 基于 schema 注释生成 Query→scope 映射制品（JSON），作为 PBAC 的唯一事实来源输入；落盘 `reports/permissions/graphql-query-permissions.json`。
3) 实现对齐
   - PBAC 映射改为读取生成制品；移除/封存手写映射；角色→权限外部化（配置/数据库）并移除硬编码。
4) 校验与证据
   - 运行权限契约校验脚本，生成报告至 `reports/permissions/*`；记录日志至 `logs/plan252/*`；提交前通过。

---

## 7. 验收标准（可度量）
- OpenAPI security scopes 一致性=100%
- GraphQL→PBAC 映射覆盖率=100%
- Resolver 授权覆盖率=100%
- 关键用例通过（最小集，正/负向各至少1）：组织层级、时态历史、职位任职、职类目录、审计历史
- 证据落盘完整（logs/reports 路径）

---

## 8. 脚本接口规范（仅文档，脚本由后续实现）
- CLI 名称（建议）：`node scripts/quality/auth-permission-contract-validator.js`
- 输入参数（示例）：
  - `--openapi docs/api/openapi.yaml`
  - `--graphql docs/api/schema.graphql`
  - `--resolver-dirs internal/organization/resolver,cmd/hrms-server/query/internal/auth`
  - `--out reports/permissions`
  - `--fail-on missing-scope,unregistered-scope,mapping-missing,resolver-bypass`
- 输出制品：
  - `reports/permissions/openapi-scope-usage.json`（路径→scope 使用明细）
  - `reports/permissions/openapi-scope-registry.json`（注册表导出）
  - `reports/permissions/graphql-query-permissions.json`（由 schema 注释生成的映射，SSoT）
  - `reports/permissions/summary.txt`（人类可读汇总）
- 校验项（最小集）：
  - OpenAPI 引用→注册一致性（阻断）
  - GraphQL 映射覆盖率（阻断）
  - Resolver 授权调用覆盖率（阻断）
  - 注册但未引用（信息）
- 退出码：0=通过；非0=失败（summary.txt 指明失败项与计数）

---

## 9. 风险与回滚
- 发现契约与实现不一致时，先回滚实现或更新契约，禁止引入第二事实来源；所有变更记录至 `CHANGELOG.md` 并在本计划下留痕。
- 临时兼容项必须标注 `// TODO-TEMPORARY(YYYY-MM-DD):` 并在一个迭代内回收。

---

附：唯一事实来源
- OpenAPI（REST + 权限 scopes）：`docs/api/openapi.yaml`
- GraphQL（查询权限注释）：`docs/api/schema.graphql`
- 原则与约束：`AGENTS.md`
