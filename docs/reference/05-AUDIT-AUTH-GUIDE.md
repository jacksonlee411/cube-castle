# 05 — 认证审计事件规范（BFF /auth/*）

最后更新：2025-09-14

本指南说明命令服务内置的 BFF 认证路由在关键节点写入的审计事件，便于运维/安全核查与问题追踪。该文档为 Reference；数据库模式以迁移文件为准，项目原则与约束以 `AGENTS.md` 为唯一事实来源。

—

## 事件模型（摘要）

表：`audit_logs`（见 database/migrations/**，以及实现 `internal/audit/logger.go`）

> 组织单元相关的行级触发器在 Plan 234 中已全部移除，`database/migrations/20251110110000_234_remove_org_unit_triggers.sql` 通过 Goose 显式删除触发器与函数，命令服务仅通过 `internal/organization/audit/logger.go:120-189` 写入结构化审计，确保单一事实来源。

- 核心字段（与实现保持一致）：
  - `event_type`：事件类型（AUTHENTICATION/ERROR/CREATE/UPDATE/...）
  - `resource_type`：资源类型（USER/SYSTEM/...）
  - `resource_id`：资源标识（用户ID/系统模块等）
  - `actor_id`/`actor_type`：发起者信息
  - `action_name`：动作名（LOGIN/REFRESH/LOGOUT/LOGOUT_RP/OIDC_CALLBACK 等）
  - `request_id`：请求ID（用于跨日志关联）
  - `success`：是否成功
  - `error_code`/`error_message`：错误场景记录
  - `before_data`/`after_data`：可选上下文
  - `timestamp`：事件时间

—

## BFF 写入的认证相关事件

- 登录成功（OIDC 回调）
  - `event_type=AUTHENTICATION`、`resource_type=USER`
  - `action_name=LOGIN`
  - `after_data.scopes=[...]`（来自 id_token 或换票上下文）

- 刷新成功
  - `event_type=AUTHENTICATION`、`resource_type=USER`
  - `action_name=REFRESH`

- 登出（本地会话清理）
  - `event_type=AUTHENTICATION`、`resource_type=USER`
  - `action_name=LOGOUT`

- IdP 退出重定向（RP 发起）
  - `event_type=AUTHENTICATION`、`resource_type=USER`
  - `action_name=LOGOUT_RP`
  - `after_data.end_session_endpoint=...`

- 回调失败（令牌交换失败/ID Token 校验失败）
  - `event_type=ERROR`、`resource_type=SYSTEM`
  - `action_name=OIDC_CALLBACK`
  - `error_code=TOKEN_EXCHANGE_FAILED | ID_TOKEN_INVALID`

- 刷新失败/会话过期
  - `event_type=ERROR`、`resource_type=SYSTEM`
  - `action_name=REFRESH`
  - `error_code=SESSION_EXPIRED | REFRESH_FAILED`

—

## 查询示例

- 最近 50 条认证相关事件（全局）：

```sql
SELECT event_type, resource_type, resource_id, action_name, request_id,
       success, error_code, error_message, timestamp
FROM audit_logs
WHERE event_type IN ('AUTHENTICATION', 'ERROR')
ORDER BY timestamp DESC
LIMIT 50;
```

- 某用户的登录与刷新轨迹（按时间逆序）：

```sql
SELECT action_name, success, timestamp, error_code
FROM audit_logs
WHERE resource_type='USER' AND resource_id = $1
ORDER BY timestamp DESC
LIMIT 100;
```

- 某 requestId 关联的认证事件：

```sql
SELECT *
FROM audit_logs
WHERE request_id = $1
ORDER BY timestamp DESC;
```

—

## 对齐与后续

- 文档对齐：若错误码或字段变更，请同步更新本文件与 `docs/api/openapi.yaml`。
- 安全增强（建议）：
  - 刷新令牌复用检测（reuse detection）→ 记录 `REFRESH_REUSE_DETECTED` 并强制登出。
  - 加入来源IP/User-Agent 等上下文字段（可存入 `before_data`）。
