# 42号文档：实现清单文件校验报告

## 校验范围与方法
- 基于 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，提取文档中出现的代码文件标识（扩展名包含 `.go`、`.ts(x)`、`.sql`、`.js`、`.yml/.yaml`、`.json`、`.md`）。
- 在仓库根目录对所有文件进行一次遍历，将每个标识与实际文件的相对路径进行后缀匹配，并过滤掉 `node_modules/` 目录中的第三方产物。
- 统计命中的真实路径并按领域归类，同时记录文档引用但仓库中缺失或路径不一致的条目。

> 注：部分清单项（例如 `organization_create.go`、`organization_update.go`、`auth/graphql_middleware.go`、`graphql_envelope.go`、`request_id.go` 等）对应多个实现文件，已在下方按照实际所在目录分别列出。

## 命令服务（Command Service）

### 处理器（Handlers）
- `cmd/organization-command-service/internal/handlers/devtools.go`
- `cmd/organization-command-service/internal/handlers/operational.go`
- `cmd/organization-command-service/internal/handlers/organization_create.go`
- `cmd/organization-command-service/internal/handlers/organization_events.go`
- `cmd/organization-command-service/internal/handlers/organization_history.go`
- `cmd/organization-command-service/internal/handlers/organization_routes.go`
- `cmd/organization-command-service/internal/handlers/organization_update.go`

### 服务层（Services）
- `cmd/organization-command-service/internal/services/cascade.go`
- `cmd/organization-command-service/internal/services/operational_scheduler.go`
- `cmd/organization-command-service/internal/services/organization_temporal_service.go`
- `cmd/organization-command-service/internal/services/temporal.go`
- `cmd/organization-command-service/internal/services/temporal_monitor.go`

### 中间件
- `cmd/organization-command-service/internal/middleware/performance.go`
- `cmd/organization-command-service/internal/middleware/ratelimit.go`
- `cmd/organization-command-service/internal/middleware/request.go`

### 共享模块（Auth / Audit / Utils / Repository）
- `cmd/organization-command-service/internal/auth/rest_middleware.go`
- `cmd/organization-command-service/internal/audit/logger.go`
- `cmd/organization-command-service/internal/repository/organization_create.go`
- `cmd/organization-command-service/internal/repository/organization_update.go`
- `cmd/organization-command-service/internal/utils/response.go`
- `cmd/organization-command-service/internal/utils/validation.go`
- `cmd/organization-command-service/internal/validators/business.go`

## 查询服务（Query Service）
- `cmd/organization-query-service/internal/auth/graphql_middleware.go`
- `cmd/organization-query-service/internal/middleware/graphql_envelope.go`
- `cmd/organization-query-service/internal/middleware/request_id.go`

## 共享内部包（internal/*）
- `internal/auth/graphql_middleware.go`
- `internal/middleware/graphql_envelope.go`
- `internal/middleware/request_id.go`

## 前端（frontend）

### 配置与常量
- `frontend/src/shared/config/constants.ts`
- `frontend/src/shared/config/environment.ts`
- `frontend/src/shared/config/ports.ts`
- `frontend/src/shared/config/tenant.ts`
- `frontend/tests/config/ports.ts`
- `frontend/tests/e2e/config/test-environment.ts`
- `frontend/src/design-system/tokens/brand.ts`
- `frontend/src/features/organizations/constants/formConfig.ts`
- `frontend/src/features/organizations/constants/tableConfig.ts`
- `frontend/src/features/temporal/constants/temporalStatus.ts`
- `frontend/src/features/temporal/index.ts`
- `frontend/src/shared/utils/constants.ts`（文档引用 `constants.ts` 指向同一文件）

### API 与类型
- `frontend/src/shared/api/auth.ts`
- `frontend/src/shared/api/contract-testing.ts`
- `frontend/src/shared/api/error-handling.ts`
- `frontend/src/shared/api/error-messages.ts`
- `frontend/src/shared/api/graphql-enterprise-adapter.ts`
- `frontend/src/shared/api/type-guards.ts`
- `frontend/src/shared/api/unified-client.ts`
- `frontend/src/shared/types/api.ts`
- `frontend/src/shared/types/converters.ts`
- `frontend/src/shared/types/type-guards.ts`

### Hooks
- `frontend/src/shared/hooks/useEnterpriseOrganizations.ts`
- `frontend/src/shared/hooks/useMessages.ts`
- `frontend/src/shared/hooks/useOrganizationMutations.ts`

### 工具 & 验证
- `frontend/src/shared/utils/colorTokens.ts`
- `frontend/src/shared/utils/organization-helpers.ts`
- `frontend/src/shared/utils/organizationPermissions.ts`
- `frontend/src/shared/utils/statusUtils.ts`
- `frontend/src/shared/utils/temporal-converter.ts`
- `frontend/src/shared/utils/temporal-validation-adapter.ts`
- `frontend/src/shared/utils/index.ts`
- `frontend/src/shared/validation/index.ts`
- `frontend/src/shared/validation/schemas.ts`

## 脚本、SQL 与迁移
- `scripts/generate-implementation-inventory.js`
- `scripts/fix-graphql-scan-issue.sql`
- `scripts/quality/architecture-validator.js`
- `scripts/quality/document-sync.js`
- `sql/hierarchy-consistency-check.sql`
- `database/migrations/023_audit_exclude_dynamic_temporal_flags.sql`
- `generate-implementation-inventory.js`（命令调用同上脚本）

## 文档与契约
- `docs/api/openapi.yaml`
- `CLAUDE.md`

## CI / 基础设施
- `.github/workflows/contract-testing.yml`
- `.github/workflows/duplicate-code-detection.yml`
- `docker-compose.yml`

## 未命中或路径不一致项（2025-10-09 校核结果）
- `/.well-known/jwks.js` — **已确认** 实际为运行时端点 `/.well-known/jwks.json`，实现清单已修正。
- `ValidationRules.ts` — **已修正** 更新为 `frontend/src/shared/validation/index.ts` / `schemas.ts` 的统一验证入口。
- `docker-compose.monitoring.yml` — **已修正** 改为仓库现有的 `docker-compose.yml` / `docker-compose.dev.yml` / `docker-compose.e2e.yml`。
- `docs/architecture/query-layer.md` — **已修正** 指向 `docs/architecture/metacontract-v6.0-specification.md` 第3.6节。
- `frontend/src/features/temporal/components/TemporalMasterDetailView.ts` — **已确认** 文档现使用 `.tsx` 文件路径。
- `internal/metrics/collector.go` — **已修正** 替换为 `middleware/performance.go`、`services/temporal_monitor.go`、`handlers/operational.go` 的实际实现。
- `reports/architecture/architecture-validation.js` 等 — **已修正** 统一引用为仓库存在的 `reports/architecture/architecture-validation.json` 和 `reports/implementation-inventory.json` 等 JSON 报告。
- `simple-validation.ts` — **已修正** 说明已删除，并指向统一验证系统。
- `temp-inventory.md` — **已标注** 为临时输出示例，明确不纳入仓库。
- `useOrganizations.ts` — **已修正** 文档注明该 Hook 已移除，仅保留迁移提示。

## 结论
- 文档列出的核心实现文件大部分仍与仓库一致；命令服务与前端关键模块均已找到对应路径。
- 原未命中条目已逐项对齐现有实现；实现清单与仓库状态同步。
- 对运行态端点与临时输出已在文档中标注用途与范围，防止被误判为缺失文件。
