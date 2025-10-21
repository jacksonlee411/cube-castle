# 职位任职跨租户校验执行记录（2025-10-21）

## 执行概览
- **脚本**：`tests/consolidated/position-assignments-cross-tenant.sh`
- **环境**：`docker-compose.dev.yml`（rest-service / postgres / redis 均运行且健康）
- **BASE_URL**：`http://localhost:9090`
- **租户配置**：
  - Tenant A — `3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`（拥有职位 `P9000003`）
  - Tenant B — `11111111-2222-3333-4444-555555555555`
- **令牌生成**：使用 `make jwt-dev-mint`，分别输出至 `.cache/dev-tenantA.jwt`、`.cache/dev-tenantB.jwt`
- **脚本输出**：见 `reports/position-stage4/position-assignments-cross-tenant.log`

## 结果摘要
| 步骤 | 预期 | 实际 | 备注 |
|------|------|------|------|
| Tenant B 读取 | 403/404 | 404 | 符合预期，命令服务未泄露跨租户数据 |
| Tenant A 读取 | 200 | 200 | 成功返回职位任职列表 |
| Tenant B 创建 | 403/404 | 404 | 跨租户写入被拒绝 |
| Tenant A 创建 | 201 | 201 | 成功创建 acting 任职（脚本自动生成 UUID 员工） |
| Tenant B 关闭 | 403/404 | 404 | 跨租户关闭被拒绝 |
| Tenant A 关闭 | 200 | 200 | 已成功关闭，命令服务返回 200 |

## 缺陷复盘与修复
- **根因**：命令服务 `PositionAssignmentRepository.CloseAssignment` 在未提供备注时，向 PostgreSQL 传入未显式类型的 `NULL` 参数，导致 `pq: could not determine data type of parameter $4`，进而让接口返回 500。
- **修复**：更新 `cmd/organization-command-service/internal/repository/position_assignment_repository.go`
  - 将 `notes` 更新逻辑改为“是否保留原备注”的布尔开关，加上 `sql.NullString` 类型参数，避免类型不确定。
  - 重新构建 `rest-service` 镜像并重启容器使变更生效。
- **验证**：修复后脚本再次执行全部通过；详情见最新日志 `reports/position-stage4/position-assignments-cross-tenant.log`。命令服务审计日志也记录了 `VacatePosition` 事件。

## 数据状态
- 成功关闭后，`position_assignments` 中脚本生成的任职均为 `ENDED`，`positions.headcount_in_use` 恢复为 0。
- 为保持环境整洁，对历史测试留下的任职执行了统一清理（同脚本附录中的 SQL），确保职位 `P9000003` 当前无残留占用。

## GraphQL 跨租户补充验证
- **脚本**：`tests/consolidated/position-assignments-graphql-cross-tenant.sh`
- **GraphQL_URL**：`http://localhost:8090/graphql`
- **输出日志**：`reports/position-stage4/position-assignments-graphql-cross-tenant.log`

| 步骤 | 预期 | 实际 | 备注 |
|------|------|------|------|
| Token/Header 不匹配 | 403 | 403（`TENANT_MISMATCH`） | GraphQL 中间件拦截租户不一致请求 |
| Tenant B 查询 `positionAssignments` | totalCount = 0 | totalCount = 0 | 按租户隔离返回空集，无越权数据 |
| Tenant A 查询 `positionAssignments` | totalCount > 0 | totalCount = 11 | 返回 acting 任职并携带 `effectiveDate/actingUntil/autoRevert` 字段 |

> GraphQL 服务升级至最新 Schema（`effectiveDate`、`actingUntil` 等字段），并修复输入过滤器指针问题；容器镜像通过 `docker compose up -d --build graphql-service` 重建。

## 后续建议
1. **回归测试**：将脚本套件（REST + GraphQL）纳入 CI 冒烟测试，持续验证跨租户隔离与 acting 任职关闭链路。
2. **开发守护**：在命令服务错误处理中追加更明确的日志（可选），方便排查类似 SQL 参数类型问题。
3. **计划文档**：同步更新 86 号计划与 06 号协作日志，注明缺陷与修复细节（待后续归档时执行）。
