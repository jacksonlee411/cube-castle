# Plan 219A – Organization 模块目录重构与 Facade 基线

**文档编号**: 219A  
**关联路线图**: Plan 219  
**目标周期**: Week 3 Day 15-17（对应 204 号计划行动 2.6）  
**负责人**: 架构组 + 后端团队  

---

## 1. 目标

1. 搭建 `internal/organization/` 目录结构，涵盖 domain/repository/service/handler/resolver/scheduler/validator/audit/dto/middleware。
2. 定义 `api.go`（Command/Query Facade）并在 `cmd/hrms-server/{command,query}` 增加适配层，统一依赖入口。
3. 完成 Job Catalog、Position、Organization 主干代码及共享类型的迁移，生成迁移清单 v1。
4. 确保迁移后编译、基础单元测试通过，行为与旧代码等价。

---

## 2. 范围

| 模块 | 内容 |
|------|------|
| Domain | `organization.go`、`department.go`（聚合内节点）、`position.go`、`job_catalog.go`、`events.go` |
| Repository | `organization_repository.go`、`position_repository.go`、`job_catalog_repository.go`、`hierarchy_repository.go` |
| Service | `organization_service.go`、`position_service.go`、`job_catalog_service.go` |
| Handler | `organization_handler.go`、`position_handler.go`、`job_catalog_handler.go` |
| Resolver | 整体 GraphQL resolver（保持现有查询契约） |
| DTO / Types | 请求/响应 struct 与共享类型 |

不包含：
- Assignment 查询增强（219B 负责）
- Audit / Validator 规则细化（219C 负责）
- Temporal / Scheduler 迁移（219D 负责）
- E2E / 性能测试（219E 负责）

---

## 3. 聚合说明（Organization 与 Department）

- Department 视为 Organization 聚合内的节点，使用 `unitType=DEPARTMENT` 表示，不新增独立接口或路由。
- 领域模型需包含：
  - `OrganizationUnit` 抽象，字段：`Code`、`Name`、`UnitType`（ORGANIZATION/DEPARTMENT）、`ParentCode`、`Status`、`EffectiveDate` 等。
  - 组织与部门共享 repository/service，层级操作（移动、启停）在同一事务内处理。
- 迁移时须验证：
  - 组织接口能处理 `unitType=DEPARTMENT` 的 CRUD、层级、状态流转。
  - GraphQL 查询（`organizations`, `organizationHierarchy`）仍返回部门节点。
  - Migration 清单记录相关文件（例如原 `organization_hierarchy.go`、`organization_*` handler）的迁移去向。

---

## 3. 详细任务

1. **创建目录与 README**
   - 更新并作为唯一事实来源的 `internal/organization/README.md`：说明职责、聚合边界（含 Department 为 Organization 子聚合的描述）、依赖。
   - 子目录：domain、repository、service、handler、resolver、scheduler、validator、audit、dto、middleware。

2. **迁移核心代码**
   - 将 `cmd/hrms-server/command/internal/{handlers,repository,services}` 中与组织/职位/Job Catalog 相关文件迁至新目录，并修正导入路径。
   - 将 `cmd/hrms-server/query/internal/graphql/*` 和 query repository 迁至新目录下的 resolver/repository。

3. **定义 `api.go` 与适配层**
   - 在 `internal/organization/api.go` 暴露命令侧接口（OrganizationAPI）与 QueryFacade。
   - 在 `cmd/hrms-server/command/main.go`、`cmd/hrms-server/query/main.go` 中新增适配层，仅依赖 `api.go`。

4. **迁移清单 v1**
   - 记录每个旧文件的去向、状态（迁移完成/待清理/保留），并集中维护于 `internal/organization/README.md` 的“迁移清单”小节，避免额外散落文档。
   - 添加 CI 检查脚本（可选）或说明如何验证 `cmd/*` 不再直接引用旧路径。

5. **基础验证**
   - `go build ./cmd/hrms-server/...`
   - `go test ./internal/organization/...`（受影响的单元测试）
   - 手动验证最基础 REST/GraphQL 调用（smoke test）。

---

## 4. 依赖

| 项目 | 说明 |
|------|------|
| Plan 216-218 | eventbus/database/logger 基础设施已就绪 |
| Plan 210 | Goose/Atlas 基线；确保迁移不破坏迁移脚本结构 |

---

## 5. 验收标准

- [ ] 新目录结构与 README 完成。
- [ ] `cmd/hrms-server/*` 仅通过 `api.go` 使用组织模块，无直接引用旧内部包。
- [ ] Job Catalog / Position / Organization 主干逻辑迁移完成，旧目录剩余文件列入清单。
- [ ] 迁移清单 v1（含状态、回退信息）已提交。
- [ ] `go build ./cmd/hrms-server/...` 成功。
- [ ] 相关单元测试通过（至少覆盖 service/repository 基础用例）。

---

## 6. 风险与应对

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 迁移过程中产生循环依赖 | 中 | 分阶段迁移，使用适配层过渡 |
| cmd 层遗漏引用旧路径 | 高 | 使用 `rg` / `go list` 检查导入；迁移清单审查 |
| 合并冲突 | 中 | 提前与并行分支沟通，拆小 PR，保持主干同步 |

---

## 7. 交付物

- `internal/organization/` 目录与 README
- `api.go` + 适配层代码
- 迁移清单 v1
- 更新后的单元测试结果
