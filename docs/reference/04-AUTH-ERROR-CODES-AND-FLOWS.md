# 04 — 认证与会话错误码对照表与客户端处理

最后更新：2025-09-14

本文档为 Reference 类文档，作为实现与联调的快速参照。认证/会话相关契约以 `docs/api/openapi.yaml` 为唯一事实来源；项目原则与约束以仓库根目录 `AGENTS.md` 为唯一事实来源。

—

## 契约对齐差异清单（与 OpenAPI）

- 错误响应包：统一使用 `ErrorResponse` 包装（`success=false,error{code,message,details},timestamp,requestId`）。本文档列举的错误码均应落在该结构内。
- 401 未授权示例（OpenAPI）：示例码为 `INVALID_TOKEN`、`TOKEN_EXPIRED`。本文档中 BFF 级别的细分码（如 `STATE_EXPIRED/ID_TOKEN_INVALID/...`）作为 `error.code` 的取值使用，仍归类为 401。
- 403 禁止访问（OpenAPI）：示例码为 `INSUFFICIENT_PERMISSIONS` 与 `TENANT_ACCESS_DENIED`。客户端需区分并给出不同提示，详见“客户端处理建议”。
- `/auth/logout` GET：OpenAPI 规定当未配置 IdP 时允许返回 `200`（fallback）；本文档已补充该情形说明。
- `/.well-known/oidc`：已在 OpenAPI 中公开（BFF 子集，字段使用 camelCase）。

## 错误码与 HTTP 状态

说明：以下为 BFF `/auth/*` 路由当前实现的错误码与 HTTP 状态。前端统一客户端的分流策略见“客户端处理建议”。

- OIDC_NOT_CONFIGURED
  - 场景：未配置 `OIDC_ISSUER/CLIENT_ID/REDIRECT_URI` 时访问 `/auth/login` 或相关流程。
  - HTTP：501 Not Implemented
  - 客户端：展示“未配置企业登录”，提示联系管理员；开发/模拟可提示开启 `OIDC_SIMULATE=true`。

- INVALID_CALLBACK
  - 场景：回调缺少 `code/state`。
  - HTTP：400 Bad Request
  - 客户端：重定向至登录页并带错误提示。

- STATE_EXPIRED
  - 场景：`state` 无效或过期（10 分钟 TTL）。
  - HTTP：401 Unauthorized
  - 客户端：重定向至登录页重新发起。

- ID_TOKEN_INVALID（含子类：ISSUER_MISMATCH/AUDIENCE_MISMATCH/EXPIRED/NOT_YET_VALID/NONCE_MISMATCH）
  - 场景：`id_token` 校验失败（RS256+JWKS、iss/aud/exp/nbf/nonce）。
  - HTTP：401 Unauthorized
  - 客户端：重定向至登录页；必要时提示“企业登录会话无效”。

- SESSION_EXPIRED
  - 场景：`sid` 不存在/过期，或 `/auth/refresh` 调用 IdP 刷新失败。
  - HTTP：401 Unauthorized
  - 客户端：清理本地态并跳转登录。

- CSRF_CHECK_FAILED
  - 场景：修改类端点（`/auth/refresh`、`/auth/logout`）缺少或错误 `X-CSRF-Token`。
  - HTTP：401 Unauthorized
  - 客户端：重新加载 `/auth/session` 以获取最新 `csrf`，再重试一次；仍失败则跳登录。

- REFRESH_FAILED
  - 场景：调用 IdP 刷新失败（如 `invalid_grant`）。
  - HTTP：401 Unauthorized
  - 客户端：清理本地态并跳转登录。

注：设计稿中曾提到 419，用于区分“会话过期”。经 2025-09-29 认证小组评审，当前仍统一返回 401，并在后端/前端提示中明确“会话已失效”语义；若未来需要 419，将提前更新 OpenAPI 与实现。

—

## 客户端处理建议（统一客户端已实现）

- GraphQL/REST 请求默认行为
  - 401：调用 `POST /auth/refresh` 强制刷新一次 → 仍失败则派发 `auth:unauthorized` 并跳 `/login`。
  - 403：区分两类错误
    - `TENANT_ACCESS_DENIED`：提示“无权访问所选租户”，引导切换/选择可用租户。
    - `INSUFFICIENT_PERMISSIONS`：提示“权限不足（缺少必要权限）”，可显示所需权限或联系管理员。
  - 500：提示“服务器内部错误”。
- 租户头设置
  - 生产态：`X-Tenant-ID` 来自 `/auth/session` 的 `tenantId`；无可用值时回退 `env.defaultTenantId`（仅开发态）。

—

## 核心端点速览

- GET `/auth/login`：302 跳 IdP（授权码 + PKCE）
- GET `/auth/callback?code&state`：换票 + 建立会话，设置 `sid`/`csrf` Cookie
- GET `/auth/session`：返回 `{ accessToken, expiresIn, tenantId, user, scopes }`
- POST `/auth/refresh`：`X-CSRF-Token` 必填；服务端用 refresh token rotation
- POST `/auth/logout`：清本地会话（204）
- GET `/auth/logout`：RP 发起 IdP 退出（302）；当未配置 IdP 时可返回 `200`（fallback），由客户端决定后续跳转；可带 `?redirect=` 覆盖回跳
- GET `/.well-known/oidc`：发现文档关键字段（未纳入 OpenAPI 契约，待评审）
- GET `/.well-known/jwks.json`：BFF JWKS（RS256 启用时；OpenAPI 已提供）

—

## 运行与联调（快速）

- RS256 + 模拟流：`make run-auth-rs256-sim` → `make auth-flow-test`
- 成功标准：`/auth/session` 返回 200；GraphQL 携带 BFF 短期 Token 返回 200；`/auth/refresh` 返回 200。

—

## 一致性与质量门槛（执行清单）

- 契约优先：涉及权限与认证的变更，必须先在 `docs/api/openapi.yaml` 中定义或更新（包括 scopes、错误响应示例）。
- 命名一致：JSON 字段保持 camelCase；错误码为机器可读常量（推荐大写下划线），出现在 `error.code`。
- CSRF 校验：`/auth/refresh`、`/auth/logout` 必须校验 `X-CSRF-Token`（OpenAPI `securitySchemes.CSRFToken`）。
- 多租户：业务 REST 端点必须携带 `X-Tenant-ID`；BFF `/auth/session` 返回的 `tenantId` 作为前端默认租户。
- 临时条目回收：本文档中 `// TODO-TEMPORARY` 项需在 2025-09-30 前完成评审与处理（保留或删除），否则 CI 文档检查应提示。
