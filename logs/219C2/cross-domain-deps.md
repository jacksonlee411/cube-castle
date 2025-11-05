# 219C2C 跨域依赖清单（Day 23 更新）

> 更新日期：2025-11-07 17:10 CST  
> 维护人：Codex Agent（219C2C 实施）

| 依赖域 | 接口 / 仓储 | 作用 | 当前状态 | 备注 |
| --- | --- | --- | --- | --- |
| Organization Hierarchy | `repository.HierarchyRepository` | 组织节点、层级、祖先链查询（POS-ORG、CROSS-ACTIVE） | ✅ `internal/organization/validator/testing_stubs.go#L11-L64` 提供 Stub；Production 版本由 `CommandModule` 注入 | Stub 默认返回空，计划执行时由单测注入场景数据 |
| Organization Aggregate | `repository.OrganizationRepository` | 组织状态、存在性校验（Position → Organization） | ✅ `CommandModule` 中实例化并传递给 PositionService/Validator | 与 REST 命令共享 DB 连接，禁止直连其他数据源 |
| Job Catalog | `repository.JobCatalogRepository` | 职位引用 Job Catalog（POS-JC-LINK） | ✅ Stub 已就位（`testing_stubs.go#L66-L112`）；真实仓储在 Service 中可用 | 需要根据规则加载 Family/Role/Level 时态信息 |
| Position Assignments | `repository.PositionAssignmentRepository` | FTE 累加、任职状态（ASSIGN-FTE/STATE） | ✅ Service 中现有依赖，此外 `testing_stubs.go#L114-L126` 暴露 Stub | 单测通过 stub 注入任职数据；生产环境复用当前仓储 |
| Validation 工厂 | `validator.PositionValidationService` / `AssignmentValidationService` | 统一命令入口挂载验证链 | ✅ `CommandModule` 现使用 `NewPositionAssignmentValidationService` 注入真实链式校验，旧 Stub 仅保留测试用途 | 219C2C 已落地 POS/ASSIGN/CROSS 规则，日志见 `logs/219C2/test-Day23.log` |

## 迁移接入说明
1. `internal/organization/service/position_service.go` 已在所有职位/任职命令执行前调用链式验证器，失败时抛出 `validator.ValidationFailedError` 并触发审计。
2. `internal/organization/handler/position_handler.go` 与 GraphQL Resolver 可复用相同的错误响应/审计逻辑。
3. 单元测试通过 `validator/testing_stubs.go` 组合仓储依赖；生产环境依赖真实仓储实例。

## 待确认事项
- [x] 219C2C 实施完成：`CommandModule` 使用链式验证器，Stub 仅用于测试。
- [ ] 验证链 `ValidationResult.Context.executedRules` 当前仍返回空数组，若需追踪执行轨迹需在后续迭代补充。
