# 权限系统落地与加固实施方案 (13)

## 1. 背景与目标
- 背景: 目前权限体系在契约层（OpenAPI）已定义为 PBAC 细粒度 scope，但后端 Go 服务未落地鉴权/权限校验，前端存在基于本地角色的判断，开发令牌和 OpenAPI scope 命名不一致，监控/审计指标未在实际检查点接入。
- 目标: 在不打破现有开发体验的前提下，分阶段完成权限体系从“契约化定义”到“端到端强制执行”的落地与加固，覆盖鉴权、授权、监控、审计与多租户隔离。

## 1.1 原则与优先级（遵循 CLAUDE.md 原则 14）
- 早期阶段专注原则: 当前项目处于早期迭代期，优先保证核心功能与 API 一致性，避免过度工程化。
- 分阶段推进与降复杂: 仅实施低成本、收益显著的基础能力；企业级能力延后到生产准备阶段。
- 优先级定义（与本计划关联）：
  - P1（当前聚焦）: 核心功能完善、API 一致性、前端体验（不在本计划范围内）。
  - P2（短期）: 基础权限验证（本计划 M1-M2）。
  - P3（中期）: 权限体系完善与基础监控（本计划 M3）。
  - P4（长期）: 企业级安全加固（本计划 M4-M6）。

## 2. 范围与非目标
- 范围: 统一权限命名与常量、Go 服务 JWT 鉴权与 scope 校验、前端基于 scope 的 UI 能力控制、权限指标与审计接入、多租户 RLS 落地、生产级密钥与令牌策略。
- 非目标: 全量 E2E 测试体系搭建（由 QA 计划覆盖）；跨域微服务统一授权（后续随模块演进评估）。

## 3. 里程碑与时间线（按阶段推进）
- [P2] M1 命名统一与开发令牌修正（0.5-1 周）
  - 统一权限名，修正开发令牌 scope，建立权限常量源。
- [P2] M2 后端鉴权/授权中间件接入（0.5-1 周）
  - Go 服务启用 JWT 验证与 scope 校验，关键路由接入，接通监控指标与标准错误码。
- [P3] M3 前端基于 scope 的 UI 权限（功能相对稳定后 0.5-1 周）
  - 替换基于角色的 UI 判断，启用 scope 解析与能力函数；基础监控（轻量埋点）。
- [P4] M4 监控与审计完善（生产准备期 0.5 周）
  - 拒绝事件审计、权限检查时延与成功率指标面板化、SLO 管控。
- [P4] M5 多租户 RLS 落地（生产准备期 1 周）
  - 在数据库会话注入租户变量，启用/验证 RLS 策略。
- [P4] M6 安全工程化（生产准备期 1-2 周）
  - RS256 + JWKS、密钥轮换、TTL/刷新/吊销策略，区分 M2M 与用户委托令牌。

## 4. 详细实施步骤
### 4.1 权限命名统一与常量源（M1）
- 建立“权限常量源”（后端与前端共享清单）并与 OpenAPI 保持一致：
  - Basic CRUD: org:read, org:create, org:update, org:delete
  - State: org:suspend, org:activate
  - Hierarchy: org:read:hierarchy, org:move, org:create:child
  - Temporal: org:read:history, org:read:future, org:create:planned, org:modify:history, org:cancel:planned
  - Audit & Analytics: org:read:audit, org:read:stats, org:read:timeline
  - System: org:validate, org:maintenance, org:batch-operations
- 修正开发令牌中使用的非标准 scope（例如 org:write → org:update；补齐 org:create 等）。
- 兼容策略：过渡期内可同时接受旧 scope（标记为 Deprecated），并在日志中提示迁移。

参考：
- docs/api/openapi.yaml:26、docs/api/openapi.yaml:65
- docs/development-plans/11-api-permissions-mapping.md: 权限总表与映射
- middleware/auth.js:146（开发令牌 scope 需对齐）

### 4.2 Go 鉴权与授权中间件（M2, P2）
- 新增 Gin 中间件链：
  1) JWT 验证：验证签名、iss/aud、exp/iat；解析 tenantId/clientId/userId/scopes。
  2) Scope 校验：按路由声明的必需 scope 校验；失败返回 403（INSUFFICIENT_PERMISSIONS）。
  3) 白名单：/health 与 /metrics 无需鉴权。
- 路由绑定：在关键路由上声明所需 scope（与 OpenAPI 对齐）。例如：
  - POST /api/v1/organization-units → org:create
  - PUT /api/v1/organization-units/:code → org:update
  - DELETE /api/v1/organization-units/:code → org:delete
  - POST /api/v1/organization-units/:code/activate → org:activate
  - POST /api/v1/organization-units/:code/suspend → org:suspend
- 指标接入：在每次权限检查处调用 RecordPermissionCheck(success, duration)。
- 错误码与响应：统一返回企业包装格式与标准错误码（401/403）。

参考：
- go-app/cmd/server/main.go:25（路由注册入口）
- go-app/internal/metrics/prometheus.go:139（指标定义）、go-app/internal/metrics/prometheus.go:252（RecordPermissionCheck）
- docs/api/openapi.yaml:81（安全与 scope 绑定）

### 4.3 前端基于 scope 的 UI 权限（M3, P3）
- 解码前端 JWT（或从后端提供的 token info）并构建当前用户 scope 集合。
- 新增权限工具：根据 scope 计算 OrganizationOperationContext 与 TemporalPermissions。
- 替换现有基于角色字符串判断（admin/manager）逻辑，避免与后端策略分裂。
- 错误提示与禁用态：与标准错误码映射一致。

参考：
- frontend/src/shared/utils/organizationPermissions.ts:22-34（当前基于 role 的判断）
- frontend/src/shared/api/error-messages.ts: 权限与常见错误文案

### 4.4 监控与审计接入（M4, P4）
- 指标：权限检查总量/成功率/时延（已存在定义），接入到权限仪表盘并设置 SLO（P99 ≤ 50ms，失败率 ≤ 1%）。
- 审计：记录拒绝事件（INSUFFICIENT_PERMISSIONS/INVALID_TOKEN），包含 tenantId、clientId、userId、endpoint、requiredScopes、grantedScopes、requestId、时间戳。
- 弃用使用：保持 ADR-008 弃用端点审计与提示一致性。

参考：
- go-app/internal/metrics/prometheus.go:139、153、252
- go-app/internal/api/middleware/deprecated.go: 审计模式参考

### 4.5 多租户隔离（RLS）落地（M5, P4）
- 数据库连接层在进入事务前执行：SET LOCAL app.currentTenantId = $tenantId。
- 为需要隔离的表启用 RLS 并配置基于会话变量的策略。
- 验证与回归：构造跨租户数据尝试用例，确保被数据库强制拒绝。

参考：
- docs/architecture/castle-blueprint.md:232-234（RLS 规范）

### 4.6 安全工程化（M6, P4）
- 生产改为 RS256 + JWKS，启用密钥轮换与缓存失效策略；区分 dev/stg/prod 配置。
- 令牌策略：短 TTL（如 15-60m）+ 刷新机制 + 吊销通道；区分 Machine-to-Machine（client_credentials）与用户委托（authorization_code/自有 IdP）。
- 日志与隐私：令牌与隐私信息脱敏；开启最小权限默认策略。

参考：
- cmd/oauth-service/main.js（现为开发用途，仅保留 dev 环境）

## 5. 兼容与迁移策略
- 令牌 scope 向后兼容：过渡期接受 org:write，同时在日志与响应 Header 中提示迁移至 org:update（Deprecation 提示）。
- 前端兼容：在切换为 scope 判定前，保留 role→scope 的降级映射开关；灰度发布后移除。
- 文档：同步更新 Postman/Insomnia 集与 curl 示例，确保新旧令牌都可用于验证。

## 6. 验收标准（DoD，按阶段）
- [P2/M1-M2] 基础通过
  - 未携带/无效令牌 → 401（INVALID_TOKEN/TOKEN_EXPIRED）
  - 缺少所需 scope → 403（INSUFFICIENT_PERMISSIONS）
  - 关键路由（create/update/delete/activate/suspend）声明并强制校验 scope，与 OpenAPI 一致
  - 轻量指标埋点可用（无需面板/SLO），统一错误码响应
- [P3/M3] 前端对齐
  - UI 能力基于 scope 控制（替换 role 判断），按钮显隐/禁用与后端一致
  - 基础监控：前端错误上报与权限失败计数
- [P4/M4-M6] 生产准备
  - 权限指标仪表盘可视化，SLO 生效；拒绝事件审计字段完整
  - RLS 启用且跨租户访问验证失败（通过测试用例验证）
  - 生产环境 RS256 + JWKS；密钥轮换演练通过

## 7. 风险与缓解
- 令牌/权限命名不一致 → 建立常量源并在 CI 中加入契约一致性检查
- UI 行为变化影响用户 → 逐步灰度，提供明确的错误提示与帮助链接
- 性能回退 → 指标可见，权限检查 P99 ≤ 50ms；必要时缓存可授权 scope
- 误伤流量（403 增加）→ 审计与指标联动快速定位、热修复
- 阶段推进不一致 → 明确 P2/P3/P4 验收门槛，未达成不得推进下一阶段

## 8. 回滚策略
- 中间件可按环境变量开关：关闭后恢复为仅记录、不阻断
- 回滚到上一版本镜像；保留审计与指标用于回溯

## 9. 文档与培训
- 更新开发指南与 API 契约注释，补充“权限常见错误与排查”章节
- 为前后端与 QA 提供 1 小时工作坊示例：如何生成令牌、验证 scope、定位 403

## 10. 责任分工（Agent）
- Backend Agent: M1、M2、M4（后端侧）、M5、M6（后端侧）
- Frontend Agent: M3、M4（前端侧）
- QA Agent: 回归与权限用例、跨租户场景、性能基线
- Architecture Agent: 契约与命名治理、迁移策略把关
- DevOps Agent: 指标/仪表盘、密钥与 JWKS 配置、RLS 运维

## 11. 参考文件
- docs/api/openapi.yaml:65、docs/api/openapi.yaml:81
- docs/development-plans/11-api-permissions-mapping.md:1
- go-app/cmd/server/main.go:25
- go-app/internal/metrics/prometheus.go:139
- go-app/internal/metrics/prometheus.go:252
- go-app/internal/api/middleware/deprecated.go:1
- middleware/auth.js:146
- frontend/src/shared/utils/organizationPermissions.ts:22
- frontend/src/shared/api/error-messages.ts:1

---

文档版本: v1.0.0  
最后更新: 2025-09-06（已按阶段与优先级重排）
