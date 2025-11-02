# 204-HRMS 系统实施路线图与时间表

**文档编号**: 204
**标题**: HRMS 模块化演进实施路线图
**创建日期**: 2025-11-03
**最后更新**: 2025-11-03
**相关文档**:
- `203-hrms-module-division-plan.md` (主计划)
- `206-Alignment-With-200-201.md` (对齐分析)

---

## 概述

本文档为 203 号文档提供详细的实施路线图，包括：
- 四个阶段的具体行动计划
- 每个阶段的交付成果与验收标准
- 关键时间节点与里程碑
- 风险识别与应对措施
- 资源需求与团队分工

---

## 第一阶段：模块统一化（Week 1-2）

### 阶段目标

统一 go.mod 模块结构，建立统一的模块化单体基础，为后续模块开发做准备。

### 关键行动

| 行动项 | 描述 | 负责人 | 完成日期 | 依赖 |
|--------|------|--------|--------|------|
| 1.1 | 确认主模块名称：`cube-castle` | 架构师 | Day 2 | - |
| 1.2 | 合并所有 go.mod 到主模块 | 架构师 + 后端TL | Day 3 | 1.1 |
| 1.3 | 迁移 organization-command-service 代码 | 后端TL | Day 4 | 1.2 |
| 1.4 | 迁移 organization-query-service 代码 | 后端TL | Day 5 | 1.2 |
| 1.5 | 提取共享代码到 pkg/ 和 internal/ | 架构师 | Day 6-7 | 1.3, 1.4 |
| 1.6 | 验证编译与测试通过 | QA | Day 8 | 1.5 |
| 1.7 | 部署到测试环境 | DevOps | Day 9 | 1.6 |
| 1.8 | 性能与功能回归测试 | QA | Day 10 | 1.7 |

### 交付成果

- ✅ 统一的 go.mod 模块 (`module cube-castle`)
- ✅ 新的项目结构：
  ```
  /cmd/hrms-server/
    ├── command/main.go
    ├── query/main.go
    └── main.go
  /internal/auth/
  /internal/cache/
  /internal/config/
  /pkg/health/
  ```
- ✅ 编译通过，所有测试通过
- ✅ 功能等同于旧版本（无功能破裂）

### 验收标准

```bash
# 编译验证
go build ./cmd/hrms-server

# 测试验证
go test ./...

# 功能验证
- REST API 正常响应
- GraphQL Query 正常返回
- 权限系统正常工作
- 缓存系统正常工作
```

### 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|--------|
| 编译失败 | 阻断 | 中 | 提前进行试验性重构，积累经验 |
| 性能下降 | 高 | 低 | 保持原有依赖优化，分析性能指标 |
| 功能缺失 | 高 | 低 | 代码完整性检查，逐项验证 |

---

## 第二阶段：建立模块化结构（Week 3-4）

### 阶段目标

为新模块创建统一的开发模板，建立共享基础设施，准备第一个新模块的开发。

### 关键行动

| 行动项 | 描述 | 负责人 | 完成日期 | 依赖 |
|--------|------|--------|--------|------|
| 2.1 | 实现 pkg/eventbus/ 事件总线 | 基础设施 | Day 12 | 1.8 |
| 2.2 | 实现 pkg/database/ 数据库层 | 基础设施 | Day 13 | 1.8 |
| 2.3 | 实现 pkg/logger/ 日志系统 | 基础设施 | Day 13 | 1.8 |
| 2.4 | 为既有迁移补写 `Down` 脚本并接入 Goose | 数据库 | Day 14 | 1.8 |
| 2.5 | 配置 Atlas `migrate diff` 流程（临时库 + 校验） | 数据库 | Day 14 | 2.4 |
| 2.6 | 重构 organization 模块结构 | 架构师 | Day 15 | 2.1-2.5 |
| 2.7 | 创建模块开发模板文档（含 sqlc/outbox/Docker 准则） | 架构师 | Day 15 | 2.6 |
| 2.8 | 构建 Docker 化 PostgreSQL 集成测试基座 | QA | Day 16 | 2.2 |
| 2.9 | 验证 organization 模块正常工作（含 Goose up/down + Docker 测试） | QA | Day 17 | 2.6-2.8 |
| 2.10 | 更新项目 README 与开发指南 | 文档 | Day 18 | 2.7 |

### 交付成果

- ✅ pkg/eventbus/ 事件总线完整实现
  - Event 接口定义
  - EventBus 接口
  - MemoryEventBus 实现
  - 单元测试覆盖 > 80%

- ✅ pkg/database/ 数据库共享层
  - 连接池管理
  - 事务支持
  - 事务性发件箱（outbox）表结构
  - sqlc/atlas/goose 脚手架与 Make 目标

- ✅ pkg/logger/ 日志系统
  - 结构化日志
  - 日志级别控制
  - 性能监控集成

- ✅ database/migrations/
  - 全量补齐 `-- +goose Down` 脚本
  - `atlas.hcl`/`goose.toml` 管理配置

- ✅ Docker 集成测试基座
  - `docker-compose.test.yml` 启动 PostgreSQL
  - `make test-db` 运行 goose up/down + 集成测试

- ✅ organization 模块重构完成
  ```
  /internal/organization/
    ├── api.go                   # 公开接口
    ├── internal/
    │   ├── service/
    │   ├── repository/
    │   ├── handler/             # REST
    │   ├── resolver/            # GraphQL
    │   └── domain/
  ```

### 验收标准

```bash
# 事件总线测试
go test ./pkg/eventbus -v

# organization 模块测试
go test ./internal/organization/... -v

# 迁移验证
make db-migrate-verify   # goose up/down + atlas diff

# Docker 集成测试
make test-db
```

### 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|--------|
| 共享基础设施设计不当 | 高 | 中 | 提前与團隊评审，确保扩展性 |
| organization 模块改造破裂 | 高 | 中 | 先在分支测试，完全验证后合并 |
| 性能回归 | 中 | 低 | 基准测试对比，性能分析 |
| Down 脚本遗漏 | 高 | 中 | Goose 验收前执行 up/down 预演 |
| Docker 集成测试不稳定 | 中 | 中 | 固化镜像版本并在 CI 预跑 |

---

## 第三阶段：实现 workforce 模块（Week 5-8）

### 阶段目标

完成第一个新模块的开发，验证模块化架构的可行性和有效性。

### 关键行动

| 行动项 | 描述 | 负责人 | 完成日期 | 依赖 |
|--------|------|--------|--------|------|
| 3.1 | 更新 OpenAPI：添加 workforce 端点 | 后端TL | Day 18 | 2.7 |
| 3.2 | 更新 GraphQL Schema：添加 Employee Query | 后端TL | Day 18 | 2.7 |
| 3.3 | 设计 workforce 数据模型 | 后端 | Day 19 | 2.2 |
| 3.4 | 编写 workforce 模块首批 sqlc 查询与生成脚手架 | 后端 | Day 21 | 2.2 |
| 3.5 | 实现 workforce repository 层（使用 sqlc 生成代码） | 后端 | Day 22 | 3.4 |
| 3.6 | 实现 workforce service 层 | 后端 | Day 25 | 3.5 |
| 3.7 | 实现 workforce REST handler | 后端 | Day 26 | 3.6 |
| 3.8 | 实现 workforce GraphQL resolver | 后端 | Day 27 | 3.6 |
| 3.9 | 定义 workforce 域事件（使用 outbox `event_id`） | 架构师 | Day 28 | 3.6 |
| 3.10 | 编写单元测试（覆盖 > 80%） | QA | Day 29-30 | 3.6 |
| 3.11 | 编写 Docker 化集成测试 & goose up/down 验证 | QA | Day 31 | 2.8 |
| 3.12 | 契约测试验证 | QA | Day 32 | 3.7-3.8 |
| 3.13 | 编写 E2E 测试：员工入职流程 | QA | Day 33 | 3.12 |
| 3.14 | 性能测试与优化 | 后端 | Day 34 | 3.13 |
| 3.15 | 部署到测试环境 | DevOps | Day 35 | 3.14 |

### 交付成果

- ✅ workforce 模块完整实现
  - 数据库表：wf_employees, wf_employment_events, wf_employee_history
  - API 端点（REST + GraphQL）
  - 事件定义：EmployeeCreated, EmployeeTransferred, EmployeeTerminated 等

- ✅ 模块与 organization 的集成
  - organization 可通过 interface 访问 workforce 数据
  - 员工-职位的对应关系正确维护

- ✅ 数据访问层演进
  - sqlc 生成器纳入 CI（`make sqlc-generate`）
  - 关键查询迁移至 sqlc 并通过回归测试

- ✅ API 版本升级至 v4.8.0
  - OpenAPI.yaml 更新
  - schema.graphql 更新
  - 迁移指南准备

- ✅ 完整的测试覆盖
  - 单元测试：> 80%
  - 集成测试：关键路径覆盖
  - E2E 测试：员工入职-转岗-离职

### 验收标准

```bash
# 编译验证
go build ./cmd/hrms-server

# 测试验证
go test ./internal/workforce/... -v
测试覆盖率 > 80%

# sqlc 生成验证
make sqlc-generate

# 迁移与回滚验证
make db-migrate-verify

# Docker 集成测试
make test-db

# API 验证
- GET /workforce/employees
- POST /workforce/employees
- GET /workforce/employees/{id}
- PATCH /workforce/employees/{id}

# GraphQL 验证
query { employees { id name status } }

# 跨模块验证
- 员工-职位对应关系正确
- 事件发布/订阅正常
```

### 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|--------|
| 模块间通信有缺陷 | 高 | 中 | 单独进行通信测试，充分验证 |
| 数据一致性问题 | 高 | 中 | 异步一致性测试，覆盖重复 event_id 场景 |
| 性能瓶颈 | 中 | 中 | 数据库查询优化，缓存策略 |
| sqlc 迁移阻力 | 中 | 中 | 安排培训与 pair programming |
| Docker 集成测试耗时 | 中 | 中 | 缓存镜像，并行执行测试 |

---

## 第四阶段：实现 contract 模块（Week 9-12）

### 阶段目标

完成 Core HR 域的最后一个模块，建立完整的人员生命周期管理能力。

### 关键行动

| 行动项 | 描述 | 负责人 | 完成日期 | 依赖 |
|--------|------|--------|--------|------|
| 4.1 | 更新 OpenAPI：添加 contract 端点 | 后端TL | Day 36 | 3.14 |
| 4.2 | 更新 GraphQL Schema：添加 Contract Query | 后端TL | Day 36 | 3.14 |
| 4.3 | 设计 contract 数据模型与工作流 | 合规组 | Day 37-38 | - |
| 4.4 | 实现 contract repository/service | 后端 | Day 41 | 4.3 |
| 4.5 | 实现 contract REST/GraphQL 接口 | 后端 | Day 43 | 4.4 |
| 4.6 | 定义合同生命周期事件 | 架构师 | Day 44 | 4.5 |
| 4.7 | 集成 workforce 和 contract 的关系 | 后端 | Day 45 | 4.6 |
| 4.8 | 编写单元/集成/契约/E2E 测试 | QA | Day 46-50 | 4.7 |
| 4.9 | 完整链路测试：员工入职-签合同-离职 | QA | Day 51 | 4.8 |
| 4.10 | 性能测试与优化 | 后端 | Day 52 | 4.9 |
| 4.11 | 部署到生产前环境 | DevOps | Day 53 | 4.10 |

### 交付成果

- ✅ contract 模块完整实现
  - 数据库表：ct_contracts, ct_contract_templates, ct_contract_history
  - 合同模板系统
  - 合同签署、续签、终止工作流

- ✅ Core HR 域完整（P0 优先级全部完成）
  - organization 模块：✅
  - workforce 模块：✅
  - contract 模块：✅

- ✅ 完整的人员生命周期管理
  - 招聘 → 入职 → 签合同 → 转岗 → 离职
  - 所有阶段的数据完整性和一致性保证

- ✅ API 版本升级至 v4.8.0
- ✅ 生产准备完成

### 验收标准

```bash
# 完整流程测试
1. 创建员工 → EmployeeCreated 事件
2. 创建职位 → 分配员工到职位
3. 签署合同 → ContractSigned 事件
4. 转岗 → EmployeeTransferred 事件 + 合同更新
5. 终止合同 → ContractTerminated 事件
6. 离职 → EmployeeTerminated 事件

# 数据一致性验证
- 员工状态与合同状态同步
- 组织单元的人员计数正确
- 审计日志完整
```

### 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|--------|
| 合规要求遗漏 | 高 | 中 | 充分与合规部门沟通，需求复核 |
| 跨模块事务复杂 | 中 | 中 | 充分的事件驱动测试 |
| 旧系统迁移 | 高 | 中 | 提前准备数据迁移脚本 |

---

## 后续阶段：逐步实施其他模块（Week 13+）

### 优先级顺序

| 优先级 | 模块 | 预计周期 | 目标完成 |
|--------|------|--------|--------|
| P1 | performance（绩效） | 8-10周 | Q1 2026 |
| P1 | compensation（薪酬） | 10-12周 | Q2 2026 |
| P1 | payroll（薪资） | 12-16周 | Q2 2026 |
| P2 | recruitment（招聘） | 8周 | Q3 2026 |
| P2 | development（发展） | 8-10周 | Q3 2026 |
| P2 | attendance（考勤） | 8-10周 | Q4 2026 |
| P3 | compliance（合规） | TBD | Q1 2027 |

---

## 重要里程碑总结

| 里程碑 | 日期 | 验收标准 |
|--------|------|--------|
| **模块统一化完成** | Week 2 | go.mod 统一，代码编译通过 |
| **基础设施完善** | Week 4 | eventbus, database, logger 完整 |
| **organization 重构完成** | Week 4 | 功能等同，结构符合模板 |
| **workforce MVP 完成** | Week 8 | 单元测试 > 80%，E2E 通过 |
| **contract MVP 完成** | Week 12 | 完整生命周期测试通过 |
| **Core HR 生产就绪** | Week 12 | 性能优化，部署准备完成 |

---

## 资源需求

### 人员配置

| 角色 | 需求数 | 负责工作 |
|------|--------|--------|
| 架构师 | 1 | 整体设计、模块接口、技术方案 |
| 后端开发 | 4 | 模块实现、REST/GraphQL 接口 |
| QA | 2 | 单元测试、集成测试、E2E 测试 |
| DevOps | 1 | 部署、监控、性能优化 |
| 文档 | 0.5 | API 文档、开发指南 |
| 合规/业务 | 0.5 | 需求评审、测试用例 |

### 基础设施

- PostgreSQL 数据库（开发、测试、生产环境）
- Redis 缓存（可选，用于性能优化）
- Docker 容器（一致的开发与部署环境）
- CI/CD 流水线（自动化测试与部署）

---

## 风险整体评估

### 高风险项

1. **模块间通信设计** → 提前充分验证，准备备选方案
2. **数据一致性保证** → 充分的异步事件测试
3. **旧系统迁移** → 准备完整的迁移计划

### 中风险项

1. 性能瓶颈 → 定期性能基准测试
2. 模块结构不适配 → 提前进行试验性实现
3. 团队学习曲线 → 提前培训，充分文档

### 低风险项

1. 编译失败 → 代码复审，充分测试
2. 功能缺失 → 需求清晰化，验收标准明确

---

## 监控与控制

### 每周进度检查

```
周一：阶段计划确认
周三：进度中期评估
周五：周总结与调整
```

### 关键指标

| 指标 | 目标 | 检查周期 |
|------|------|--------|
| 测试覆盖率 | > 80% | 每周 |
| 代码审查通过率 | 100% | 每周 |
| 性能基准 | 不低于 organization 模块 | 每阶段 |
| 缺陷关闭率 | P0 当天，P1 一周内 | 每周 |

---

## 成功标志

### 第一阶段完成标志
✅ go.mod 统一
✅ 所有编译通过
✅ 功能等同于旧版本

### 第二阶段完成标志
✅ eventbus 完整实现
✅ organization 模块按模板重构
✅ 模块开发文档完成

### 第三阶段完成标志
✅ workforce 模块完整实现
✅ 与 organization 成功集成
✅ 测试覆盖率 > 80%
✅ E2E 测试通过

### 第四阶段完成标志
✅ contract 模块完整实现
✅ Core HR 域全部完成
✅ 完整链路测试通过
✅ 生产部署准备完成

---

**文档版本历史**:
- v1.0 (2025-11-03): 初始版本，详细的四阶段实施路线图与时间表
