# 06 — 生产态登录方案推进记录（取代旧版内容）

最后更新：2025-09-14
牵头：Architecture / Backend / Frontend / DevOps / QA
状态：阶段小结 + 下一步计划（对齐 CLAUDE.md 治理）

—

## 当前进展（已落实）
- 契约公开与对齐
  - 在 `docs/api/openapi.yaml` 公开 `/.well-known/oidc`（camelCase：issuer、authorizationEndpoint、tokenEndpoint、endSessionEndpoint、jwksUri）。
  - 补充 `501 OIDC_NOT_CONFIGURED` 示例；去重并保留唯一 `/.well-known/jwks.json` 定义。
- 后端 BFF（命令服务内集成）
  - 实现 `/.well-known/oidc` 并改为直接 JSON 返回（非统一信封）。
  - 统一租户错误：缺少租户头→`401 TENANT_HEADER_REQUIRED`；租户不匹配→`403 TENANT_MISMATCH`。
  - 权限不足错误统一→`403 INSUFFICIENT_PERMISSIONS`。
  - 新增 `AUTH_ONLY_MODE=true` 快速联调模式（仅启用 `/auth/*` 与 `/.well-known/*`）。
- 前端统一客户端
  - 401 自动一次刷新；403 细分提示（租户访问 vs 权限不足）。
- 环境修正
  - `.env` 更正：`DEFAULT_TENANT_ID=3b99930c-4dc6-4cc9-8f9e-123456789012`。
- 集成自检与代运行结果
  - 新增脚本：`scripts/tests/test-auth-integration.sh`（健康检查/发现/模拟登录/会话/多租户校验）。
  - 代运行：`/auth/session` 返回的 `tenantId` 正确；发现端点在未配置 IdP 时返回 501。`AUTH_ONLY_MODE` 下业务路由未启用，相关用例标记为跳过；启业务路由后因迁移未成功，命中 404（待迁移后复测）。

—

## 下一步计划（完整业务路由验证）
1) 启动依赖并迁移
- `docker compose up -d postgres redis`
- `make db-migrate-all`（或修复 `/tmp/migrate.log` 提示的迁移错误）

2) 启动命令服务（开启业务路由）
- 环境：`OIDC_SIMULATE=true`（开发态模拟），关闭 `AUTH_ONLY_MODE`
- 启动：`go run ./cmd/organization-command-service/main.go`

3) 运行集成自检脚本（应命中业务路由校验）
- `scripts/tests/test-auth-integration.sh`
- 预期：
  - 缺少租户头 → 401 `TENANT_HEADER_REQUIRED`
  - 租户不匹配 → 403 `TENANT_MISMATCH`
  - 正确租户 + 最小载荷 → 201/400（取决于字段校验）

4) 扩展 E2E/契约测试
- 增补 403 两类与 501 发现端点的用例；前端交互提示与重定向路径验证。
- 校验 CSRF 头在 `/auth/refresh`、`/auth/logout` 的强制性（OpenAPI `CSRFToken`）。
 - 增加“重复实现守卫”用例：后端禁止 `jwt.Parse` 直调（仅允许在 `internal/auth/jwt.go`），前端禁止新增除 `shared/api/auth.ts` 之外的 token 管理实现。

—

## 验收标准（质量门槛）
- 契约一致：OpenAPI 与实现一致；错误码出现在 `error.code`，字段命名 camelCase。
- 多租户强校验：401/403 行为稳定；前后端提示一致。
- 发现与 JWKS：`/.well-known/oidc` 与 `/.well-known/jwks.json` 可用；未配置 IdP 时 501。
- 前端策略：401 自动刷新一次；403 正确分流并给出用户可理解提示。
- 文档治理：临时项以 `// TODO-TEMPORARY` 标注并在时限内回收；参考/计划目录边界满足 CI 检查。
 - 唯一性：全库仅一处 JWT 校验入口（后端）、一处 token 管理入口（前端）；新增变更通过 IIG/CI 检查。

—

## 风险与待解事项
- 数据库迁移：当前脚本提示“迁移跳过或失败”，需修复后再做业务路由验证。
- AUTH_ONLY_MODE：仅为临时联调模式，需在 2025-09-30 前去除或改为可配置子服务。// TODO-TEMPORARY
- OIDC 配置：生产接入需要 IdP 配置与重定向 URI 白名单。

—

## 附：关键改动索引
- 契约：`docs/api/openapi.yaml`
- BFF 发现端点：`cmd/organization-command-service/internal/authbff/handler.go`
- 多租户/权限错误码：`internal/auth/middleware.go`、`cmd/organization-command-service/internal/auth/rest_middleware.go`
- 前端分流：`frontend/src/shared/api/unified-client.ts`
- IIG 去重整改（2025-09-14）
  - 后端：统一以 `internal/auth/jwt.go` 为权威实现；`internal/auth/middleware.go` 已改为调用 `JWTMiddleware.ValidateToken`，移除重复校验器 `internal/auth/validator.go`；RS256 支持通过 `JWKS_URL`/`JWT_PUBLIC_KEY_PATH` 注入（映射 `internal/config/jwt.go`）。
  - 前端：唯一权威实现 `frontend/src/shared/api/auth.ts`；统一客户端 `frontend/src/shared/api/unified-client.ts`、`AuthProvider` 与错误处理均复用，不再额外解析/持久化 token。
  - 治理：README 已含“硬编码 JWT 配置检查”脚本；后续在 CI `document-sync.yml` 中启用“重复实现守卫”规则（扫描禁用 JWT 解析散落实现）。

- 自检脚本：`scripts/tests/test-auth-integration.sh`
- 环境变量：`.env`
