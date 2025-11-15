# Plan 252 - 权限一致性与契约对齐

文档编号: 252  
标题: 权限一致性与契约对齐（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
状态: 已完成（已归档）  
归档: `docs/archive/development-plans/252-permission-consistency-and-contract-alignment.md`（签字：`docs/archive/development-plans/252-signoff-20251115.md`）  
关联计划: 202、203、OpenAPI/GraphQL 契约、auth/pbac

---

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
  - 采纳 `position:assignments:audit` 作为正式权限（见“决策与长期策略”）；需在 OpenAPI scopes 注册表登记，并补齐 PBAC 映射与 resolver 授权注释的一致性。
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
- 生产构建禁用 devMode（CI 检测默认构建 devMode=false；ADMIN/admin 不存在放行路径）
- 兼容别名清理（见“命名与兼容清理”）按期完成
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
- devMode 若误入生产构建，立即回滚构建或强制配置覆盖；保留最小只读降级路径（健康检查）。

---

附：唯一事实来源
- OpenAPI（REST + 权限 scopes）：`docs/api/openapi.yaml`
- GraphQL（查询权限注释）：`docs/api/schema.graphql`
- 原则与约束：`AGENTS.md`
- 校验器接口规范（参考实现说明）：`docs/reference/08-PERMISSIONS-CONTRACT-VALIDATOR-SPEC.md`

---

## 10. 决策与长期策略（已确认）
- 审计权限命名
  - 采纳并注册 `position:assignments:audit`，遵循“领域[:子域]:audit”的最小权限命名范式；GraphQL 注释与 PBAC 映射保持一致。
- 角色→scope 外部化
  - OpenAPI 仍为 scope 枚举唯一事实来源；角色→scope 映射外部化于 Auth/BFF 域（数据库迁移管理），由 BFF 在签发 Token 时计算 `scope/permissions`；Query 服务仅基于 Token 中的 scopes 授权，不做 DB 回表。
  - 本仓库（Query 服务）保留最小临时兜底仅限开发测试，必须以 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注并限期回收。
- GraphQL 权限映射 SSoT
  - 以 `docs/api/schema.graphql` 的 “Permissions Required: …” 注释为唯一事实来源，构建期生成 Query→scope 映射制品，PBAC 仅消费该制品；手写常量禁用。
- 命名与演进策略
  - 延续 `domain:action` 与分层 action：`read/create/update/delete`，以及 `read:history/read:stats/read:hierarchy` 等；子域采用 `position:assignments:*` 等形式。
  - 清理同义/历史别名（见下节）。
- devMode 控制
  - 生产与 CI 强制 `devMode=false`；仅本地 `make run-dev` 开发模式允许。若 `ENV=production` 且 `devMode=true`，进程应直接失败。
- 令牌 TTL
  - 机机（Client Credentials）保持 1–4h；用户态经 BFF 签发 10–30min 短期 Access Token + 刷新流；严控权限变更生效时间。

---

## 11. 命名与兼容清理
- 清理项
  - `org:write` 与 `org:update/org:create` 的混用：保留临时向后兼容仅用于过渡期。
    - `// TODO-TEMPORARY(2025-12-15): 移除 org:write 兼容逻辑，统一使用 org:update/org:create；在权限映射与客户端依赖完成收敛后删除。`
- 文档化说明
  - 在开发者速查与变更通告中明确上述别名清理窗口与替代项。

---

## 12. 里程碑与分工（建议）
- M1 契约收敛（本计划内）
  - 完成 OpenAPI scopes 注册与路径一致性；为 schema Query 补足权限注释；生成映射并用于 PBAC；CI 启用三项阻断门禁。
- M2 签发策略外部化（由 Auth/BFF 小组牵头）
  - 新增角色/映射三表迁移；BFF 按映射签发 Token scopes；Query 侧移除角色硬编码。
- M3 兼容清理与复核
  - 移除 `org:write` 兼容；复核命名与最小用例；归档证据与更新 CHANGELOG。
> 本计划已完成并归档（状态：完成）。  
> 请以归档版本与签字纪要为准：  
> - 归档正文：`docs/archive/development-plans/252-permission-consistency-and-contract-alignment.md`  
> - 签字纪要：`docs/archive/development-plans/252-signoff-20251115.md`  
> - 校验器接口规范：`docs/reference/08-PERMISSIONS-CONTRACT-VALIDATOR-SPEC.md`

# Plan 252 - 权限一致性与契约对齐（索引占位）

本文件已归档，不再作为唯一事实来源（SSoT）。如需查阅，请访问上述归档路径。 
