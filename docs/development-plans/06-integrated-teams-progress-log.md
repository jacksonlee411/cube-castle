# 06号文档：Phase2 启动与模块化结构建设计划

> **更新时间**：2025-11-04
> **文档角色**：Phase2 启动指导与团队协作进展日志
> **关联计划**：Plan 203（HRMS 系统模块化演进）、Plan 204（实施路线图）、Plan 205（已完成并归档）、Plan 211（已完成）
> **状态**：✅ **Phase1 已完成，Phase2 启动就绪**

---

## 1. Phase1 完成总结

### 1.1 执行成果

**Plan 211 - Phase1 模块统一化已于 2025-11-04 正式关闭**

| 指标 | 目标 | 实际 | 状态 |
|------|------|------|------|
| go.mod 统一化 | 单一模块 | `module cube-castle` ✅ | ✅ 完成 |
| 目录结构 | 标准化 | `cmd/hrms-server/{command,query}` ✅ | ✅ 完成 |
| 共享代码合并 | 清晰分类 | `internal/*` 统一 ✅ | ✅ 完成 |
| 构建验证 | `go build` 通过 | 全部通过 ✅ | ✅ 完成 |
| 测试验证 | `go test ./...` 通过 | 全部通过 ✅ | ✅ 完成 |
| 功能验证 | REST/GraphQL 正常 | 正常工作 ✅ | ✅ 完成 |
| 数据一致性 | 无异常 | PASS（2025-11-03）✅ | ✅ 完成 |
| Go 工具链 | 1.24+ | 1.24.9 ✅ | ✅ 完成 |

**详见**: `reports/plan-211-closure-assessment.md`

### 1.2 关键交付物

- ✅ `reports/phase1-module-unification.md` - 执行日志与决议记录
- ✅ `reports/phase1-regression.md` - 回归测试报告
- ✅ `reports/phase1-architecture-review.md` - 架构审查报告
- ✅ `reports/plan-211-closure-assessment.md` - 关闭评估
- ✅ `docs/development-plans/211-Day2-Module-Naming-Record.md` - 模块命名决议
- ✅ `scripts/phase1-acceptance-check.sh` - 验收脚本
- ✅ `docs/archive/development-plans/205-HRMS-Transition-Plan.md` - Plan 205 设计方案（已归档）
- ✅ `docs/archive/development-plans/205-Plan-Alignment-Assessment.md` - Plan 205 对标评估（已归档）

### 1.3 前置条件满足情况

| 条件 | 状态 | 备注 |
|------|------|------|
| 设计决议（Plan 205） | ✅ 完成 | 问题诊断与设计决议已完成，已归档为参考文档 |
| 数据库基线（Plan 210） | ✅ 完成 | Goose + Atlas 工作流已落地 |
| 模块统一化（Plan 211） | ✅ 完成 | 2025-11-04 正式关闭 |
| 构建与测试验证 | ✅ 完成 | 全部通过，无遗留问题 |
| 团队培训（Go 1.24） | ✅ 完成 | 开发者需升级本地环境 ≥Go 1.24 |

### 1.4 Plan 205 完成与归档（2025-11-04）

**Plan 205 - HRMS 系统过渡方案** 在 **2025-11-04** 标记为已完成，并进行了文档规范化与归档：

| 操作 | 状态 | 说明 |
|------|------|------|
| 核心目标对齐评估 | ✅ 完成 | 与 Plan 210/211/212 方向完全一致，无异议 |
| 内容对标分析 | ✅ 完成 | 重复度 40-50%，可互补，参考价值高 |
| 原文档归档 | ✅ 完成 | 归档至 `docs/archive/development-plans/205-HRMS-Transition-Plan.md` |
| 评估报告生成 | ✅ 完成 | `docs/archive/development-plans/205-Plan-Alignment-Assessment.md` |
| 前置说明更新 | ✅ 完成 | 在 205 原位置添加完成标记与参考指引 |

**Plan 205 保留的参考价值**：
- 第一部分：问题诊断与现状分析（为什么需要 go.mod 统一化）
- 第四部分：详细的风险应对指南（症状→原因→解决方案）
- 附加资源：常用命令速查表（开发者参考）

**执行细节**：由 Plan 211（Day1-10）、Plan 212（Day6-7）负责推进

详见（已全部归档至 archive）：
- 📄 [Plan 205 - HRMS 系统过渡方案](../archive/development-plans/205-HRMS-Transition-Plan.md)
- 📋 [205-Plan-Alignment-Assessment.md（对标评估）](../archive/development-plans/205-Plan-Alignment-Assessment.md)

---

## 2. Phase2 启动指导（建立模块化结构）

### 2.1 Phase2 目标与范围

**阶段目标**：在 Phase1 统一模块化基础上，为后续新模块开发建立共享基础设施和标准化模块结构。Phase2 重点是构建模块化单体的"脚手架"而非新业务模块。

**具体工作范围**（Week 3-4，Day 12-18）：

**基础设施实现** (2.1-2.3)：
- ✅ 实现 `pkg/eventbus/` - 模块间事件驱动通信的基础
- ✅ 实现 `pkg/database/` - 统一的数据库访问层和事务管理
- ✅ 实现 `pkg/logger/` - 结构化日志系统

**数据库迁移与工作流** (2.4-2.5)：
- ✅ 完成迁移脚本 Down 脚本补齐（Plan 210）
- ✅ 完成 Atlas + Goose 工作流配置（Plan 210）

**模块重构与模板建设** (2.6-2.10)：
- ✅ 重构 `organization` 模块按新标准结构
- ✅ 创建模块开发模板与规范文档
- ✅ 建立 Docker 化集成测试基座
- ✅ 验证 organization 模块按新结构正常工作

**不在 Phase2 范围内**：
- ❌ 新的业务模块开发（workforce、contract 等）预计在 Phase3 推进
- ❌ 新增 API 端点与功能特性
- ❌ Talent Management、Compensation & Operations 等其他域

### 2.2 核心建议与工作流

#### 建议 1：建立 Bounded Context 模板

**目的**：确保每个新模块都遵循一致的架构与命名规范。

**模板结构**（参考 `organization` 现有结构）：

```
internal/<domain-name>/
├── models.go              # 数据模型与契约
├── repository.go          # 数据访问层
├── service.go             # 业务逻辑层
├── handler.go             # HTTP 处理器（命令服务）
└── resolver.go            # GraphQL 解析器（查询服务）
```

**行动项**：
- [ ] 创建 `internal/workforce/` 目录结构
- [ ] 创建 `internal/contract/` 目录结构
- [ ] 补充 `internal/<name>/README.md` 说明模块职责与边界

**负责人**：架构师
**计划完成**：Phase2 Week1

#### 建议 2：API 契约驱动开发

**目的**：避免实现与契约偏差，遵循"先契约后实现"原则（CLAUDE.md）。

**工作流**：
1. 在 `docs/api/openapi.yaml` 中定义新模块的 REST 端点
2. 在 `docs/api/schema.graphql` 中定义新模块的查询类型
3. 使用 `sqlc` 生成类型安全的数据访问代码
4. 实现 handler 与 resolver

**关键检查点**：
- 所有 API 字段使用 camelCase
- 路径参数统一使用 `{code}`
- 权限声明与 OpenAPI scopes 一致

**行动项**：
- [ ] 在 `docs/api/openapi.yaml` 中新增 workforce 端点（POST/PUT/GET `/api/v1/employees`）
- [ ] 在 `docs/api/schema.graphql` 中新增 employees 查询
- [ ] 在 `docs/api/openapi.yaml` 中新增 contract 端点
- [ ] 运行 `make sqlc-generate` 验证代码生成

**负责人**：架构师 + 后端 TL
**计划完成**：Phase2 Week1 末

#### 建议 3：数据库迁移管理

**目的**：使用 Goose + Atlas 工作流管理 schema 变更，保证可追溯性与回滚能力。

**工作流**：
1. 设计新模块所需的数据表（参考 PeopleSoft 78-79 号文档功能定义）
2. 使用 `atlas schema inspect` 生成 HCL 草稿
3. 手工审阅与调整（关注字段约束、索引、外键）
4. 用 `goose create` 生成迁移脚本
5. 执行 `make db-migrate-all` 并记录回滚验证

**关键检查点**：
- 迁移脚本包含 `-- +goose Down` 回滚逻辑
- 审计字段保持一致性（created_at, updated_at, deleted_at, is_deleted）
- 外键引用正确性验证

**行动项**：
- [ ] 为 workforce 模块设计 employees 表（employee_id, name, status, hire_date 等）
- [ ] 为 contract 模块设计 labor_contracts 表（contract_id, employee_id, type, effective_date 等）
- [ ] 执行迁移并验证 up/down 循环

**负责人**：后端 TL + DevOps
**计划完成**：Phase2 Week2

#### 建议 4：权限与 PBAC 实现

**目的**：建立一致的权限声明与验证机制。

**工作流**：
1. 在 OpenAPI `operationId` 中添加 `x-required-scopes` 扩展
2. 在 `internal/auth/pbac_rest.go` 中定义权限规则
3. 在处理器中使用中间件进行权限验证

**权限作用域示例**：
```yaml
workforce:read        # 读取员工信息
workforce:create      # 创建员工
workforce:update      # 更新员工
contract:read
contract:create
contract:update
```

**行动项**：
- [ ] 在 `docs/api/openapi.yaml` 中定义 workforce 权限 scopes
- [ ] 在 `docs/api/openapi.yaml` 中定义 contract 权限 scopes
- [ ] 在 `internal/auth/pbac_rest.go` 中补充权限验证规则
- [ ] 编写权限测试用例

**负责人**：后端 TL + 安全架构师
**计划完成**：Phase2 Week2

#### 建议 5：测试与质量保证

**目的**：确保新模块满足质量要求，无遗留技术债。

**工作流**：
1. 单元测试（≥80% 代码覆盖率）
2. 集成测试（模块与数据库交互）
3. 契约测试（API 响应格式与文档一致）
4. E2E 测试（关键业务流程）

**检查清单**：
- [ ] `go test ./internal/workforce/...` 通过，覆盖率 ≥80%
- [ ] `go test ./internal/contract/...` 通过，覆盖率 ≥80%
- [ ] `npm run lint` 无告警（前端组件如有）
- [ ] 契约测试验证 API 字段与 OpenAPI 一致
- [ ] E2E 测试覆盖 员工创建 → 合同签署 → 合同变更 → 离职 流程

**行动项**：
- [ ] 为 workforce 编写单元测试（repository、service、handler）
- [ ] 为 contract 编写单元测试
- [ ] 编写集成测试脚本
- [ ] 补充 E2E 场景

**负责人**：QA + 后端团队
**计划完成**：Phase2 Week2 末

---

## 3. Phase2 详细时间表（Week 3-4）

| 时间 | 行动项编号 | 描述 | 负责人 | 产出 | 状态 |
|------|--------|------|--------|------|------|
| **W3-D1** | 2.1 | 实现 `pkg/eventbus/` 事件总线接口与实现 | 基础设施 | eventbus.go、memory_eventbus.go | ⏳ 待启动 |
| **W3-D2** | 2.2 | 实现 `pkg/database/` 数据库连接层和事务支持 | 基础设施 | connection.go、transaction.go、outbox.go | ⏳ 待启动 |
| **W3-D2** | 2.3 | 实现 `pkg/logger/` 结构化日志系统 | 基础设施 | logger.go、formatter.go | ⏳ 待启动 |
| **W3-D3** | 2.4-2.5 | 迁移脚本和 Atlas 配置（已完成，Plan 210） | DevOps | Goose + Atlas 工作流 | ✅ 完成 |
| **W3-D4-5** | 2.6 | 重构 organization 模块按新结构 | 架构师 | 标准模块结构、api.go | ⏳ 待启动 |
| **W4-D1** | 2.7 | 创建模块开发模板与规范文档 | 架构师 | module-development-template.md | ⏳ 待启动 |
| **W4-D2** | 2.8 | 构建 Docker 化集成测试基座 | QA | docker-compose.test.yml | ⏳ 待启动 |
| **W4-D3** | 2.9 | 验证 organization 模块正常工作（单元、集成、E2E） | QA | 测试报告、验收检查表 | ⏳ 待启动 |
| **W4-D4** | 2.10 | 更新 README、开发指南与实现清单 | 文档支持 | 文档 PR | ⏳ 待启动 |

---

## 4. 团队协作要点

### 4.1 沟通节奏

- **每日站会**：下午 16:00，≤15 分钟，汇报进展与阻塞
- **周会**：每周五下午，回顾周进展、调整计划
- **架构审查**：每周三上午，定期评审代码与设计

### 4.2 信息同步

所有决议、变更、问题需即时更新至对应事实来源：
- 需求变更 → 更新 03/79 号文档
- API 变更 → 更新 `docs/api/{openapi.yaml,schema.graphql}`
- 代码变更 → 更新本文档的进展日志
- 遗留问题 → 创建专项计划并链接于本文档

### 4.3 风险预警机制

| 风险 | 影响 | 预防措施 |
|------|------|--------|
| 需求不清楚导致返工 | 中 | Week3-D1 充分梳理需求，架构师 sign-off |
| 数据库设计缺陷 | 高 | Week3-D2 需通过 DBA 审查，并在迁移脚本中验证 |
| API 契约不一致 | 中 | 开发前锁定 openapi.yaml，开发后自动化契约测试 |
| 模块边界混淆 | 中 | Week3-D2 明确模块职责边界，架构师审查 |
| 测试覆盖不足 | 中 | Week4 追踪代码覆盖率，目标 ≥80% |

---

## 5. 后续计划（Phase3+ 展望）

### 5.1 Phase3 - Talent Management 域（Week 5-6）

- `internal/recruitment/` - 招聘管理
- `internal/performance/` - 绩效管理
- `internal/development/` - 培训与发展

### 5.2 Phase4 - Compensation & Operations 域（Week 7-8）

- `internal/compensation/` - 薪酬管理
- `internal/payroll/` - 薪资计算
- `internal/attendance/` - 考勤管理
- `internal/compliance/` - 合规管理

### 5.3 持续改进

- 监控系统稳定性与性能
- 收集用户反馈并迭代
- 定期审查架构演进

---

## 6. 执行清单

### Phase2 启动前必检

- [ ] Phase1 所有交付物已验收（`reports/plan-211-closure-assessment.md`）
- [ ] 团队成员已了解新的目录结构（`cmd/hrms-server/{command,query}`, `internal/*`）
- [ ] 开发环境已升级至 Go ≥1.24
- [ ] Docker 环境已通过 `make docker-up` 启动
- [ ] 03、204 号文档已更新为 Phase2 版本

### Phase2 Week1 里程碑

- [ ] `pkg/eventbus/` 实现完成，单元测试 > 80%
- [ ] `pkg/database/` 实现完成，连接池配置正确
- [ ] `pkg/logger/` 实现完成，与 Prometheus 集成
- [ ] `organization` 模块重构完成，按新模板组织
- [ ] 第一批迁移脚本已验证（up/down 循环）

### Phase2 Week2 里程碑

- [ ] 模块开发模板文档完成
- [ ] Docker 集成测试基座可正常启动
- [ ] organization 模块所有测试通过（单元、集成、E2E）
- [ ] REST/GraphQL 端点行为验证完成
- [ ] README 和开发指南已更新

---

## 7. 联系方式与支持

**架构师**：处理模块划分、API 设计、架构决策
**后端 TL**：协调开发任务、代码审查、性能优化
**QA 负责人**：测试计划、质量把控
**DevOps**：环境配置、部署、CI/CD 维护

**本文档维护者**：Codex（AI 助手）
**最后更新**：2025-11-04

---

**Phase2 已启动！祝团队高效推进！** 🚀

