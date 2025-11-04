# Plan 219B – Assignment 查询链路与缓存刷新

**文档编号**: 219B  
**关联路线图**: Plan 219  
**依赖子计划**: 219A 完成目录/Facade 基线  
**目标周期**: Week 4 Day 18（紧随 204 行动 2.6，提前对齐 2.7/2.8）  
**负责人**: 查询服务组 + 后端团队  

---

## 1. 目标

1. 补齐 assignment 查询能力（QueryRepository、GraphQL resolver、DTO）。
2. 实现 QueryFacade 与 dispatcher 的缓存刷新/失效策略，确保任职数据变化可及时对外反映。
3. 补充 Assignment 相关的端到端查询测试与脚本。

---

## 2. 范围

| 项目 | 内容 |
|------|------|
| Repository | 新增 `postgres_assignment_repository.go`（查询侧）、扩展 Facade |
| GraphQL | 新增/更新 Assignment resolver、查询参数、DTO |
| 缓存 | 更新 dispatcher（事件类型→缓存刷新）、QueryFacade 刷新实现 |
| 测试 | Unit + Integration + GraphQL 查询覆盖；端到端验证脚本 |

不包含：命令侧 Assignment 业务逻辑（已存在，219A 迁移）、Audit/Validator 规则细化（219C）。

---

## 3. 详细任务

1. **Assignment Query Repository**
   - 新建 `internal/organization/repository/postgres_assignment_repository.go`，提供：
     - `GetAssignmentHistory(ctx, tenant, positionCode, filter, paging)`
     - `GetAssignmentStats(ctx, tenant, positionCode, orgCode)`
   - SQL 使用现有 timeline/assignment 表，确保与 Temporal 数据一致。

2. **GraphQL Resolver**
   - 扩展 resolver：`assignments`, `assignmentHistory`, `assignmentStats` 等查询。
   - 更新 GraphQL schema（如需新增字段），并保持向后兼容。

3. **QueryFacade & 缓存刷新**
   - 在 Facade 中实现 `GetAssignmentHistory`、`RefreshPositionCache` 等方法。
   - 在 dispatcher 中，根据 `AssignmentFilledEventType`、`AssignmentVacatedEventType` 触发缓存刷新。
   - 记录刷新策略（刷新单个职位缓存 vs. 刷新列表）。
4. **测试与脚本**
   - 单元测试：Repository（使用 sqlmock）、Resolver（使用 mock Facade）。
   - 集成测试：`go test ./internal/organization/... -tags=integration`（针对 assignment 查询）。
   - 端到端脚本：模拟 fill/vacate→查询历史→验证响应。
5. **文档同步**
   - 在 `internal/organization/README.md` 的“查询与缓存”小节记录新增 Facade 方法、缓存刷新策略与测试脚本路径。

---

## 4. 验收标准

- [ ] 新增查询 repository + 单元测试通过。
- [ ] GraphQL resolver 覆盖 assignment 历史/统计查询，契约无破坏。
- [ ] Dispatcher 缓存刷新逻辑可根据事件触发、记录日志。
- [ ] 端到端脚本验证任职增删改后查询结果更新。
- [ ] 更新 README / 迁移清单，记录新增文件与调用方式。

---

## 5. 风险与应对

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| GraphQL schema 变更 | 中 | 先更新 schema.graphql，再同步实现；增加契约测试 |
| 缓存未及时刷新 | 高 | 在 dispatcher 中记录耗时/命中，增加测试场景 |
| 数据一致性问题 | 中 | 与 Temporal timeline 对比检查，确保 query 使用相同数据源 |

---

## 6. 交付物

- `postgres_assignment_repository.go` + 测试
- GraphQL resolver & schema 更新
- Dispatcher 缓存刷新逻辑、日志
- 测试报告、脚本说明（更新 README 或 scripts）
