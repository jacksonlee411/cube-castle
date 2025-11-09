# Plan 220 - 模块开发模板与规范文档

**文档编号**: 220
**标题**: 标准模块开发指南 - 为后续新模块提供参考
**创建日期**: 2025-11-04
**分支**: `feature/205-phase2-infrastructure`
**版本**: v1.0
**状态**: ✅ Completed (2025-11-07)
**关联计划**: Plan 219（organization 重构）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

基于 organization 模块的重构经验，编写完整的模块开发模板文档，为后续的 workforce、contract 等新模块提供标准开发指南。

**关键交付物**:
- ✅ 模块结构模板说明（需与 `internal/organization/README.md` 描述保持一致）
- ✅ 数据访问层规范（基于当前 `database/sql` 仓储实现，并附 sqlc 评估标准）
- ✅ 事务性发件箱（Outbox）集成规范（参考 `pkg/database/outbox.go` 与相关迁移）
- ✅ Docker 集成测试规范（复用 `Makefile` 已有目标）
- ✅ 样本模块代码（参考 organization 模块最新实现）

### 1.2 为什么需要模块模板

- **一致性** - 所有新模块遵循相同的结构和规范
- **快速开发** - 开发者可快速启动新模块开发
- **知识转移** - 新成员易于理解项目架构
- **质量保证** - 标准化的质量检查清单

### 1.3 时间计划

- **计划窗口**: Week 3 Day 4 ~ Week 4 Day 1（Day 15-17，对应 2025-11-05~07，依 `docs/development-plans/215-phase2-summary-overview.md:292-308` 在 W3-D4~D5 完成资料整理、W4-D1 完成定稿）
- **交付周期**: 3 天（2 天资料采样 + 1 天定稿与评审）
- **负责人**: 架构师 + 文档支持（组织模块 Owner 负责技术审校）
- **前置依赖**: Plan 219（organization 重构完成，详见 `docs/development-plans/215-phase2-execution-log.md:35`）

### 1.4 约束与引用

为避免“第二事实来源”，本计划编写与验收必须显式引用以下权威材料：

- `CLAUDE.md` & `AGENTS.md`：总体原则、Docker 强制、开发前必检。
- `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`：开发命令、契约同步流程。
- `docs/api/openapi.yaml` 与 `docs/api/schema.graphql`：命令/查询契约。
- `internal/organization/README.md`：当前模块结构与聚合边界。
- `Makefile`：标准命令名称（如 `make docker-up`, `make run-dev`, `make test`, `make db-migrate-all`）。

文档各章节在描述结构、流程或命令时需引用以上文件，交付前需执行一致性校验（逐项对照引用是否仍成立）。

### 1.5 输入材料（Plan 219 交付物）

- `internal/organization/README.md:3`, `internal/organization/README.md:21`, `internal/organization/README.md:56` —— 目录职责、聚合边界、219E 验收脚本。
- `internal/organization/api.go:28`, `internal/organization/api.go:119` —— CommandModule 构造器与依赖注入示例。
- `internal/organization/query_facade.go:28`, `internal/organization/query_facade.go:136` —— 查询 Facade 与缓存刷新逻辑。
- `internal/organization/query_facade_test.go:42` —— Facade 行为覆盖示例。
- `docs/development-plans/215-phase2-execution-log.md:35`, `docs/development-plans/215-phase2-execution-log.md:370` —— Plan 219 完成记录与验证摘要。
- `docs/development-plans/204-HRMS-Implementation-Roadmap.md:21` —— Phase2 基础设施最新状态，用于计划对齐描述。

---

## 2. 文档内容规划

### 2.1 文件结构

```
docs/development-guides/
├── module-development-template.md   # 主文档（此方案）
├── examples/
│   ├── organization/                # organization 参考实现
│   │   ├── models.go.example
│   │   ├── repository.go.example
│   │   ├── service.go.example
│   │   └── handler.go.example
│   └── workforce/                   # workforce 示例骨架（待实现）
└── checklists/
    ├── module-structure-checklist.md
    ├── api-contract-checklist.md
    ├── testing-checklist.md
    └── deployment-checklist.md
```

### 2.2 主文档章节

#### 第一章：模块基础知识

**内容**:
- 什么是 Bounded Context
- 模块化单体架构的优势
- organization 模块作为示例

#### 第二章：模块结构模板

**内容**:
- 与 `internal/organization/README.md` 对齐的标准目录与职责
- `api.go` 对外暴露接口、命令/查询通过 `internal` 机制隔离
- 组织模块现有目录（audit/dto/handler/resolver/service/repository/validator 等）的可复用模板
- README 最低内容（聚合边界、迁移清单、测试入口）

**示例代码**:
- api.go 的标准框架
- 接口定义的最佳实践
- 依赖注入模式

#### 第三章：数据访问层（PostgreSQL + Repository 模式）

**内容**:
- `internal/organization/repository` 中 `PostgreSQLRepository` 的目录拆分与命名
- `database/sql` + `github.com/lib/pq` 的使用基线（当前唯一事实来源）
- 查询/命令共享仓储、DTO 映射与日志字段规范
- `docs/archive/plan-216-219/219A-219E-review-analysis.md` 指出的 “生成代码（如 sqlc）质量要求缺失” 问题与补救措施（先定义评估标准、再决定是否接入 sqlc）

**示例**:
```go
// 示例摘自 internal/organization/repository/postgres_organizations_list.go
func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *dto.OrganizationFilter, pagination *dto.PaginationInput) (*dto.OrganizationConnection, error) {
    page := int32(1)
    pageSize := int32(50)
    if pagination != nil && pagination.Page > 0 {
        page = pagination.Page
    }
    // ...
    rows, err := r.db.QueryContext(ctx, baseQuery, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query organizations: %w", err)
    }
    // ...
}
```

在此基础上，模板需附录《sqlc 引入评估清单》：仅当定义了生成位置、审查责任人、lint/CI 规则后方可启用，避免与现有手写 SQL 出现双事实。

#### 第四章：事务性发件箱集成

**内容**:
- 事务性发件箱模式原理
- 在 service 层的实现
- Outbox 中继器配置
- 错误处理和重试

**示例**:
```go
// 在 service 中使用事务性发件箱的标准模式
func (s *Service) CreateEntity(ctx context.Context, cmd CreateCommand) error {
    return s.db.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
        // 1. 保存业务数据
        entity := NewEntity(cmd)
        if err := s.repo.Save(ctx, tx, entity); err != nil {
            return err
        }

        // 2. 在同一事务内保存 outbox 事件
        event := NewEntityCreatedEvent(entity)
        if err := SaveOutboxEvent(ctx, tx, event); err != nil {
            return err
        }

        return nil
    })
}
```

#### 第五章：Docker 集成测试

**内容**:
- Docker Compose 配置模板
- 集成测试的标准结构
- Goose 迁移测试
- 测试数据初始化

**示例**:
```yaml
# docker-compose.test.yml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
      POSTGRES_DB: testdb
    ports:
      - "5432:5432"
```

#### 第六章：测试规范

**内容**:
- 单元测试组织
- 集成测试组织
- Mock 和 Stub 的使用
- 覆盖率目标（> 80%）

#### 第七章：API 契约规范

**内容**:
- OpenAPI/REST 命名规范
- GraphQL schema 规范
- 权限声明规范
- 版本化管理

#### 第八章：质量检查清单

**内容**:
- 代码质量检查
- 安全性检查
- 性能检查
- 文档完整性检查

---

## 3. 实施步骤

### 3.1 编写主文档 (module-development-template.md)

基于 organization 模块的经验，编写 3000-5000 字的综合指南。

**包含内容**:
- 快速开始指南
- 标准模块结构说明
- 最佳实践建议
- 常见陷阱和解决方案

### 3.2 准备示例代码

从 organization 模块提取关键代码作为示例（每个样例需注明来源文件/行号，并去除租户/秘钥信息）：
- `internal/organization/api.go` —— CommandModule/Handlers 构造代码片段，演示依赖注入顺序。
- `internal/organization/repository/*.go` —— Repository 模式与 `database/sql` 用法。
- `internal/organization/service/*.go` + `pkg/database/outbox.go` —— 事务性发件箱与 `WithTx` 模式。
- `internal/organization/handler/*.go` —— REST Handler 模板。
- `internal/organization/resolver/*.go` 或 `query_facade.go` —— GraphQL/查询 Facade 与缓存策略。

### 3.3 创建检查清单

为不同阶段提供检查清单：
- 模块结构检查清单
- API 契约检查清单
- 测试完成度检查清单
- 部署前检查清单

### 3.4 整合与审审查

- 架构师审查文档（重点核对聚合边界和依赖关系描述是否仍与 `internal/organization/README.md` 一致）
- 后端 TL 检查示例代码准确性（验证引用片段与主干代码保持同步）
- QA 验证测试规范（交叉引用 Plan 221 Docker 测试基座要求）
- 文档支持进行最终编辑并执行引用校验清单（逐项勾选 7.1 条件）

---

## 4. 文档目标受众

- **后端开发者** - 新模块实现者
- **新团队成员** - 理解项目架构
- **QA 工程师** - 了解测试策略
- **架构师** - 参考和改进

---

## 5. 文档质量标准

### 5.1 可读性

- [ ] 语言清晰，术语准确
- [ ] 有充分的代码示例
- [ ] 有流程图或架构图
- [ ] 链接到相关文档和计划

### 5.2 完整性

- [ ] 涵盖模块开发的全生命周期
- [ ] 包含常见场景和最佳实践
- [ ] 包含错误处理和边界情况
- [ ] 包含性能考虑

### 5.3 实用性

- [ ] 示例代码可直接参考
- [ ] 检查清单可直接使用
- [ ] 步骤清晰且可操作
- [ ] 与实际项目对齐

---

## 6. 关键章节示例（摘要）

### 6.1 模块结构模板章节示例

```markdown
## 标准模块结构（参考 internal/organization/README.md）

internal/{module_name}/
├── api.go                 # 模块公开接口（命令/查询入口唯一依赖）
├── audit/                 # 审计写入器
├── domain/                # 聚合根与事件
├── dto/                   # GraphQL/REST DTO
├── handler/               # REST 处理器（命令服务）
├── repository/            # PostgreSQL 仓储（database/sql）
├── resolver/              # GraphQL 解析器（查询服务）
├── scheduler/             # Temporal/Scheduler 适配
├── service/               # 领域服务
├── utils/                 # 通用工具（日志、metrics 等）
├── validator/             # 链式校验器
└── README.md              # 模块说明（聚合边界、迁移、测试）

### 各目录职责

- **repository/** - 共享命令/查询的数据访问层
- **service/** - 事务性业务逻辑，调用仓储/Outbox
- **handler/**/**resolver/** - 分别适配 REST 与 GraphQL，遵守 CQRS 分工
- **api.go** - 暴露 `CommandModule`/`ResolverModule` 等构造器，其余文件置于 `internal/` 防止跨模块引用
```

### 6.2 集成测试规范章节示例（与 AGENTS Docker 约束对齐）

```markdown
## Docker 集成测试规范

所有模块的集成测试必须依赖 `docker-compose.dev.yml` 提供的 PostgreSQL/Redis（参见 Makefile）。

### 配置步骤

1. `make docker-up` 启动 postgres/redis（或指定 `docker compose -f docker-compose.dev.yml up -d postgres redis`）。
2. `make db-migrate-all` 执行 Goose 迁移（迁移即唯一事实来源）。
3. `make test-integration` 运行带 `-tags=integration` 的 Go 集成测试；若需要 REST/GraphQL 联调，可先 `make run-dev`。
4. 验证 `curl http://localhost:9090/health` / `8090/health` 返回 200。
5. `docker compose -f docker-compose.dev.yml logs postgres` 检查慢查询，测试结束后 `make docker-down` 回收资源。

### 最佳实践

- 测试相互独立，尽量 `t.Parallel()`。
- 使用事务包裹测试并回滚，必要时 TRUNCATE 。
- 在 `tests/` 下维护数据初始化脚本，并记录日志到 `logs/`.
- 若出现 5432/6379 等端口占用，按照 `AGENTS.md` 要求先卸载宿主冲突服务，**禁止**通过修改 Compose 端口映射规避冲突。
```

---

## 7. 验收标准

### 7.1 文档完整性

- [ ] 主文档（module-development-template.md）> 3000 字
- [ ] 每章明确引用 `CLAUDE.md`/`AGENTS.md`/`docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`/`docs/api/*`/`internal/organization/README.md` 等唯一事实来源
- [ ] 至少包含 5 个完整代码示例
- [ ] 包含 3 个以上的检查清单
- [ ] 包含流程图或架构图

### 7.2 文档质量

- [ ] 内容准确无误（与 organization 模块对齐）
- [ ] 语言清晰易懂
- [ ] 示例代码可编译且正确
- [ ] 链接和引用完整正确

### 7.3 实用性

- [ ] 新模块开发者可独立参考此文档开发
- [ ] 检查清单可直接用于验收
- [ ] 代码示例可作为开发模板

---

## 8. 交付物清单

- [x] `docs/development-guides/module-development-template.md` （主文档，>=3000 字，引用区块完整）
- [x] `docs/development-guides/examples/organization/` （示例代码，附来源路径注释）
- [x] `docs/development-guides/examples/workforce/` （骨架示例，演示如何套用模板）
- [x] `docs/development-guides/checklists/module-structure-checklist.md`
- [x] `docs/development-guides/checklists/api-contract-checklist.md`
- [x] `docs/development-guides/checklists/testing-checklist.md`
- [x] `docs/development-guides/checklists/deployment-checklist.md`
- [x] 本计划文档（220）更新完成并存档

---

## 9. 执行状态（2025-11-07）

- 主文档实字数 8,704（`wc -m`），包含 7 段参考代码并逐条引用 `CLAUDE.md`、`AGENTS.md` 等唯一事实来源。
- `examples/organization/*.go.example` 覆盖 DTO、Repository、Service、Handler；注释中标注原文件路径，用于审校。
- `examples/workforce/workforce_module.go.example` 提供 CommandModule、Service、Handler 骨架，可直接复制到新模块。
- 四份检查清单与 AGENTS/Makefile 要求同步，供 PR 自查与 Phase2 验证使用。

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-07
**计划完成日期**: Week 4 Day 1 (Day 17, 2025-11-07)
