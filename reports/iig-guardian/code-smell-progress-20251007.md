# Code Smell Progress Report

**Report Date**: 2025-10-07  
**Plan**: Plan 16 – 代码异味分析与改进  
**Scope**: Go 查询/命令服务重构进展汇总

---

## 1. Execution Summary
- `go test ./cmd/organization-command-service/...` ✅
- `go test ./cmd/organization-query-service/...` ✅
- 验证内容涵盖：命令服务仓储/时间轴拆分后的所有包、查询服务模块化后的 app/repository/graphql。

## 2. Structural Metrics (Go)

| 指标 | 基线（2025-09-29） | 当前（2025-10-07） |
|------|-------------------|-------------------|
| 红灯文件（>800行） | 3 个 | **1 个** |
| 最大仓储文件行数 | 817 行 | **1 306 行**（`cmd/organization-query-service/internal/repository/postgres.go`） |
| 时间轴文件行数 | 780+ 行 | 5 个文件（最大 208 行） |
| `.golangci.yml` depguard | N/A | 已启用（阻断跨 CQRS 依赖） |

-## 3. Test Evidence
- 命令服务：`go test ./cmd/organization-command-service/...`
  - 输出：全部通过，仓储/时间轴模块覆盖。
- 查询服务：`go test ./cmd/organization-query-service/...`
  - 输出：全部通过，验证 modular main。
- 集成测试：`make test-integration`
  - 输出：Go 集成测试通过；`TestAuthFlow_RealHTTP_RS256_JWKS_and_TenantChecks` 因 `E2E_RUN` 未设置按预期跳过。
- 集成测试：`make test-integration`（2025-10-07 执行）
  - 输出：全部通过（3 个测试 PASS，1 个 SKIP - 需 E2E_RUN=1）。
  - 日志：`reports/iig-guardian/test-integration-output-20251007.log`

## 4. Outstanding Items
- ✅ 运行 `go vet ./...`（2025-10-07 已通过）。
- ✅ `make test-integration`（2025-10-07 已执行，全部通过，输出存档于 `reports/iig-guardian/test-integration-output-20251007.log`）。
- ⏳ 查询服务 `postgres.go`（1 306 行）需继续拆分至 <800 行。
- ⏳ Temporal 前端组件拆分（Phase 1.5）。

## 5. Sign-off
- 架构组：**后端红灯仍剩 1 个**（`postgres.go` 1 306 行），需继续拆分。
- 测试团队：单元测试通过（见上方命令）。

> 本报告为 Plan 16 Phase 1.4 验收记录，后续更新将继续追加至 `reports/iig-guardian/`。 
