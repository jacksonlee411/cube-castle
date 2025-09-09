# 19. 代码异味审计与整改计划（Code Smells & Remediation Plan）

**文档编号**: 19  
**最后更新**: 2025-09-08  
**维护团队**: 架构团队（主）+ 后端 + 前端 + 测试  

## 🎯 目标
- 系统性识别影响可维护性、可编译性与一致性的代码异味。
- 给出分级整改计划（P0/P1/P2），以最小成本恢复健康开发节奏与E2E稳定性。

---

## 🔎 发现概览
- 多模块与 internal 包边界混乱，导致查询服务无法稳定编译/运行。
- 硬编码（端口/租户/秘钥）与配置不统一，E2E 对端口依赖脆弱。
- API/CQRS 规范在实现与文档之间存在漂移（REST 时态查询残留）。
- 数据库层软删除与时态唯一性约束语义不一致，索引/约束命名多套并存。
- 仓库卫生问题（提交了编译产物、冗余/无效 SQL 语法）。

---

## 📋 详细问题与证据

### 1) Go 模块架构与 internal 边界（CRITICAL）
- 跨模块 internal 违规：顶层 `internal/graphql/schema_loader.go` 不能被 `postgresql-graphql-service` 跨模块使用。
- 引用缺失：`cmd/organization-query-service/main.go` 导入 `postgresql-graphql-service/internal/graphql`，但该目录不存在。
- 依赖/路径不一致：
  - 顶层未声明 `gin`，但 `internal/auth/middleware.go` 使用 gin。
  - JWT v4/v5 同时存在（顶层 `internal/auth/validator.go` 用 v4，go.mod 为 v5）。
  - 错误 import 前缀：`github.com/cube-castle/internal/config` 与实际 module 名不符。
- 证据：
  - `internal/graphql/schema_loader.go`
  - `cmd/organization-query-service/main.go`
  - `go.mod`、`go.work`、`cmd/organization-query-service/go.mod`
  - `internal/auth/middleware.go`、`internal/auth/validator.go`

### 2) GraphQL 查询服务构建链不完整（CRITICAL）
- 缺少可编译/可运行的最小装配（schema 加载、resolver、路由）。
- 证据：`cmd/organization-query-service` 目录结构与导入缺失；运行依赖冲突。

### 3) 前端端口与 E2E 基址硬编码（HIGH）
- 多处 `http://localhost:3000` 硬编码，一旦端口占用 Vite 切换端口 → 测试系统性超时。
- 证据：`frontend/tests/e2e/*.spec.ts`、`scripts/run-tests.sh`、`scripts/health-check-unified.sh` 等。
- 配置：`frontend/src/shared/config/ports.ts` 设定 3000，但测试不读取配置或环境变量。

### 4) 认证与配置不统一（MEDIUM）
- 顶层与查询服务各自一套 JWT/权限逻辑，导入与实现重复且不一致。
- 默认秘钥/租户硬编码散落（dev容忍，prod 风险）。
- 证据：
  - 顶层：`internal/config/jwt.go`、`internal/auth/*`
  - 查询服务：`cmd/organization-query-service/internal/auth/*`
  - 默认值/示例：`scripts/generate-dev-jwt.go`、`Makefile`、`sql/init/*`、`docs/api/README.md`

### 5) 数据库软删除与时态唯一性（HIGH）
- 软删除语义双轨：`is_deleted` vs `status='DELETED'+deleted_at` 并存，导致“部分唯一” WHERE 条件不一致。
- 唯一性约束命名漂移：`uk_org_ver_active_only` / `uk_org_temporal_point`、`uk_org_current_active_only` / `uk_org_current` 并存。
- 初始 schema 的 `UNIQUE (code, effective_date, record_id)` 无法阻止同一时点多版本。
- 无效 SQL：`CREATE TABLE audit_logs` 内联 `INDEX (...)` 语法（PostgreSQL 不支持）。
- 证据：
  - `sql/init/01-schema.sql`
  - `database/migrations/016/017/018/025*.sql`

### 6) CQRS 规范漂移（MEDIUM）
- 前端仍调用 REST 时态查询端点（应统一 GraphQL 查询）。
- 证据：`frontend/src/shared/hooks/useTemporalAPI.ts` 使用 `/organization-units/{code}/temporal?...`

### 7) 字段映射错误与命名不一致（MEDIUM）
- `converters.ts` 中 `recordId: data.tenantId` 明显错误映射；snake_case 残留。
- 证据：`frontend/src/shared/types/converters.ts`、多处 snake_case 断言/使用。

### 8) 仓库卫生（MEDIUM）
- 已编译二进制被提交到仓库（>20MB）。
- 建议在 `.gitignore` 排除并清理历史提交（需另行 PR）。

---

## 🛠 整改计划（分级）

### P0（阻断类，优先本周内）
1. 统一 Go 模块与导入路径
   - 收敛为单一 `go.mod`（优先），或保留多模块但迁出 shared 包至公共路径，禁止跨模块 internal。
   - 修正错误 import 前缀，统一 `github.com/golang-jwt/jwt/v5`；补齐/移除 gin 依赖。
2. 恢复查询服务最小可运行
   - 将 `schema_loader` 移至查询服务模块内；提供 resolver 桩与 `/graphql` 启动；`/metrics` 暴露健康指标。

### P1（稳定性，+2 天）
3. 前端/E2E 基址与端口治理
   - E2E 读取 `E2E_BASE_URL` 环境变量；`run-tests.sh` 增加端口预检/动态发现；可选 `strictPort: true`。
   - 清理脚本与测试的 `http://localhost:3000` 字面量。
4. 认证与配置统一
   - 抽取统一认证/配置包，命令与查询共用；删除重复实现；统一环境变量与错误响应格式。

### P2（一致性与质量门禁，+5 天）
5. 软删除/唯一性一致化
   - 单一软删除模型（建议 is_deleted）；统一创建：
     - `uk_org_temporal_point` on (tenant_id, code, effective_date) WHERE 未删除
     - `uk_org_current` on (tenant_id, code) WHERE is_current AND 未删除
   - 清理冗余/命名漂移索引；修正文档。
6. SQL 初始化修复
   - 移除内联 `INDEX`（PostgreSQL 不支持），保留后续 `CREATE INDEX`；初始化路径避免 CONCURRENTLY。
7. CQRS 纠偏
   - 前端移除 REST 时态查询调用，改用 GraphQL；对齐文档与代理配置。
8. 字段命名与映射
   - 修正 `converters.ts` 的 `recordId` 映射；启用命名一致性 CI（阻断 snake_case 出现在 API 层）。
9. 仓库卫生
   - 二进制加入 `.gitignore` 并清理；CI 增加禁止二进制入库检查。

---

## ✅ 验收标准
- 查询服务 `go build && go run` 成功，`:8090/graphql` 返回 200；Schema 从 `docs/api/schema.graphql` 加载。
- E2E 在端口 3000 被占用时仍可跑通（动态基址或明确失败提示，无静默超时）。
- 命令/查询服务共享统一认证配置；前端所有请求均附带 `Authorization` 与 `X-Tenant-ID`。
- 数据库唯一性冲突检查为 0；软删除模型单一且与 WHERE 条件一致。
- CI 新增门禁：端口硬编码扫描、跨模块 internal 检查、JWT 版本一致、禁止二进制入库。

---

## 🗓️ 里程碑
- M1（本周内）：完成 P0，恢复查询服务+初步 E2E 可运行。
- M2（+2 天）：完成 P1，E2E 稳定通过，认证统一。
- M3（+5 天）：完成 P2，质量门禁生效，文档对齐。

---

## 📎 附：关键文件索引
- Go 模块与服务：`go.mod`、`go.work`、`cmd/organization-query-service/*`、`internal/*`
- 前端端口与基址：`frontend/vite.config.ts`、`frontend/src/shared/config/ports.ts`、`scripts/run-tests.sh`、`frontend/tests/e2e/*`
- 认证与配置：`internal/config/jwt.go`、`internal/auth/*`、`cmd/organization-query-service/internal/auth/*`
- 数据库与索引：`sql/init/01-schema.sql`、`database/migrations/016/017/018/025*.sql`
- 其他：仓库二进制、命名映射 `frontend/src/shared/types/converters.ts`

