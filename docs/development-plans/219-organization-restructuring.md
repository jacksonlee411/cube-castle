# Plan 219 - `internal/organization/` 模块结构重构

**文档编号**: 219
**标题**: 现有模块重构 - 按新模板标准化
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 216-218（基础设施）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

将现有的 `organization` 模块按照 Phase2 定义的标准模板进行重构，成为后续新模块的参考实现。

**关键成果**:
- ✅ 模块目录结构标准化
- ✅ 定义模块公开接口（api.go）
- ✅ 内部实现清晰分离（service、repository、handler、resolver、domain）
- ✅ 集成新的基础设施（eventbus、database、logger）
- ✅ 验收通过，功能与重构前等同

### 1.2 重构的目标结构

```
internal/organization/
├── api.go                          # 公开接口定义
├── internal/
│   ├── service/
│   │   ├── organization_service.go
│   │   ├── department_service.go
│   │   ├── position_service.go
│   │   └── *_test.go
│   ├── repository/
│   │   ├── organization_repository.go
│   │   ├── department_repository.go
│   │   ├── position_repository.go
│   │   └── *_test.go
│   ├── handler/
│   │   └── organization_handler.go  # REST handlers
│   ├── resolver/
│   │   └── organization_resolver.go # GraphQL resolvers
│   ├── domain/
│   │   ├── organization.go
│   │   ├── events.go
│   │   └── constants.go
│   └── README.md
└── README.md
```

### 1.3 时间计划

- **计划完成**: Week 3 Day 4-5 (Day 14-15)
- **交付周期**: 2 天
- **负责人**: 架构师 + 后端团队
- **前置依赖**: Plan 216, 217, 218（基础设施完成）

---

## 2. 需求分析

### 2.1 功能需求

#### 需求 1: 模块公开接口定义

在 `internal/organization/api.go` 中定义模块对外暴露的接口：

```go
package organization

import "context"

// OrganizationAPI 定义 organization 模块的公开接口
type OrganizationAPI interface {
    // 组织单元操作
    GetOrganization(ctx context.Context, code string) (*Organization, error)
    CreateOrganization(ctx context.Context, cmd CreateOrganizationCommand) error
    UpdateOrganization(ctx context.Context, cmd UpdateOrganizationCommand) error

    // 部门操作
    GetDepartment(ctx context.Context, deptCode string) (*Department, error)
    ListDepartmentsByOrg(ctx context.Context, orgCode string) ([]*Department, error)

    // 职位操作
    GetPosition(ctx context.Context, posCode string) (*Position, error)
    ListPositionsByDept(ctx context.Context, deptCode string) ([]*Position, error)
}

// Service 实现 OrganizationAPI
type Service struct {
    orgRepo  OrganizationRepository
    deptRepo DepartmentRepository
    posRepo  PositionRepository
    eventBus EventBus
    db       Database
    logger   Logger
}
```

#### 需求 2: 模块内部分层

- **domain/** - 域模型和事件定义
- **repository/** - 数据访问层
- **service/** - 业务逻辑层
- **handler/** - HTTP 处理器（REST）
- **resolver/** - GraphQL 解析器（查询）

#### 需求 3: 部门（Department）聚合补强

- 在 `domain/` 中定义 `Department` 聚合模型，明确与 `Organization` 的关联（聚合根/子聚合边界）。
- 在 `repository/` 中提供独立的 `department_repository.go`，涵盖增删改查与层级查询。
- 在 `service/` 中实现 `department_service.go`，支持部门生命周期（创建/更新/启停/级联校验）。
- 在 `handler/` 中新增 `department_handler.go`（REST），在 `resolver/` 中新增 `department_resolver.go`（GraphQL）。
- 调整组织层级相关逻辑，将部门业务从原有 `organization_hierarchy` 中剥离并通过依赖注入复用。

#### 需求 4: 职位分配（Position Assignment）能力完善

- 命令侧：在 `handler/` 中补充 `assignment_handler.go`，支撑 fill/vacate/transfer 操作；在 `service/` 中完善 `assignment_service.go`。
- 查询侧：实现 `postgres_position_assignment_repository.go` 及对应 resolver，提供分配历史 / 当前占用 / 统计查询。
- 明确职位分配的状态流转（draft → active → vacated），并在文档中列出验证规则与审计事件。

#### 需求 5: 基础设施集成与审计/校验清单

- 使用 Plan 216 的 eventbus 发布域事件
- 使用 Plan 217 的 database 管理连接和事务
- 使用 Plan 218 的 logger 进行结构化日志记录
- 输出组织、部门、职位、职位分配的审计事件矩阵与业务校验规则清单，纳入测试与验收。
- 为 API 的破坏性变更（特别是部门独立化）编写迁移指南，说明客户端兼容策略。

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **功能等同性** | 100% | 重构前后行为完全一致 |
| **性能无退化** | P99 延迟不增加 | 监控查询响应时间 |
| **代码覆盖率** | > 80% | 单元测试 |
| **向后兼容** | ✅ 需要 | REST/GraphQL 接口签名不变 |

---

## 3. 实施步骤

### 3.1 第一阶段：分析与准备

**活动内容**:
1. 审视现有 organization 模块代码结构
2. 梳理所有现有接口和依赖
3. 制定迁移清单（哪些文件需要迁移/重构）
4. 准备测试用例列表

**产出**:
- 迁移计划文档
- 测试覆盖清单

### 3.2 第二阶段：目录重构

**活动内容**:
1. 创建新的目录结构（internal/{service,repository,handler,resolver,domain}）
2. 将现有代码按分层原则重新分类
3. 定义 api.go（公开接口）
4. 更新 import 语句

**关键检查**:
- [ ] 无重复代码
- [ ] 无循环依赖
- [ ] 内部包不可被外部导入

### 3.3 第三阶段：部门与职位分配模块落地

**活动内容**:
1. 按“需求 3 / 需求 4”完成部门与职位分配的独立化实现。
2. 补充 REST handler、GraphQL resolver、service、repository 以及对应单元测试。
3. 编写 Department/Assignment 的审计事件与业务校验规则清单。
4. 输出 API 兼容性迁移指南（列出新/旧端点、兼容期策略）。

**关键检查**:
- [ ] 存在 `department_handler.go` / `department_resolver.go` / `department_repository.go` / `department_service.go`
- [ ] 存在 `assignment_handler.go` / `assignment_resolver.go` / `postgres_position_assignment_repository.go`
- [ ] 审计与校验清单完成（列出事件、规则、优先级）
- [ ] 迁移指南草案完成

### 3.4 第四阶段：基础设施集成

**活动内容**:
1. 注入 eventbus、database、logger
2. 更新 service 层使用新的 database 接口
3. 添加域事件发布
4. 使用 logger 进行结构化日志

**示例改进**:
```go
// 旧方式：直接操作 sql.DB
func (s *Service) CreateOrg(ctx context.Context, org *Organization) error {
    _, err := s.db.Exec("INSERT INTO organizations ...")
    return err
}

// 新方式：使用基础设施
func (s *Service) CreateOrg(ctx context.Context, cmd CreateOrgCommand) error {
    s.logger.Infof("creating organization: %s", cmd.Code)

    return s.db.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
        // 1. 保存业务数据
        org := NewOrganization(cmd)
        if err := s.orgRepo.Save(ctx, tx, org); err != nil {
            return err
        }

        // 2. 发布域事件（在同一事务内）
        event := NewOrganizationCreatedEvent(org)
        if err := SaveOutboxEvent(ctx, tx, event); err != nil {
            return err
        }

        // 3. 事务提交后异步发布事件
        go s.eventBus.Publish(context.Background(), event)
        return nil
    })
}
```

### 3.5 第五阶段：测试与验证

**活动内容**:
1. 运行所有单元测试（`go test ./internal/organization/...`）
2. 执行集成测试
3. 进行回归测试（REST/GraphQL 端点）
4. 性能基准测试：对比重构前（P99、吞吐量、CPU/内存）
5. 补充 10+ 端到端测试场景（部门/职位分配等）；更新测试执行清单

**验收条件**:
- [ ] 所有测试通过
- [ ] 覆盖率 > 80%
- [ ] REST API 行为一致
- [ ] GraphQL 查询结果一致
- [ ] P99 延迟无增长（基准对比）

---

## 4. 实施检查清单

### 4.1 代码重构检查

- [ ] domain/ 目录完整（模型、事件、常量）
- [ ] repository/ 层与 database 接口集成
- [ ] service/ 层使用 logger 记录关键操作
- [ ] handler/ 层（REST）调用 service
- [ ] resolver/ 层（GraphQL）调用 service
- [ ] api.go 定义清晰的公开接口
- [ ] 无重复代码（DRY 原则）
- [ ] 无循环依赖（检查 import）

### 4.2 集成检查

- [ ] eventbus 集成：组织创建/更新时发布事件
- [ ] database 集成：使用 WithTx() 管理事务
- [ ] logger 集成：关键操作有日志记录
- [ ] Prometheus 指标：连接池、查询延迟被记录

### 4.3 测试检查

- [ ] 单元测试 > 80%
- [ ] repository 层测试（与数据库交互）
- [ ] service 层测试（业务逻辑）
- [ ] handler 层测试（REST 接口）
- [ ] resolver 层测试（GraphQL 接口）
- [ ] 集成测试（端到端流程）
- [ ] 部门 / 职位分配端到端场景（≥10 个）
- [ ] 性能基准测试报告（对比重构前数据）

### 4.4 文档检查

- [ ] api.go 中的接口有完整注释
- [ ] internal/README.md 说明各目录职责
- [ ] 更新 organization README.md

---

## 5. 风险与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|--------|
| 功能回归 | 中 | 高 | 充分的回归测试 + 灰度发布 |
| 性能下降 | 中 | 中 | 基准测试对比，优化查询 |
| 循环依赖 | 低 | 高 | 代码审查，静态检查 |
| 集成问题 | 低 | 中 | 充分的集成测试 |

---

## 6. 交付物清单

- ✅ 重构后的 organization 模块目录
- ✅ api.go（公开接口）
- ✅ 更新后的单元测试（覆盖率 > 80%）
- ✅ 集成测试脚本
- ✅ 性能基准报告
- ✅ 审计事件矩阵与业务校验规则清单
- ✅ 部门/职位分配迁移指南（API 兼容策略）
- ✅ 模块 README 文档
- ✅ 本计划文档（219）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**计划完成日期**: Week 3 Day 4-5 (Day 14-15)
