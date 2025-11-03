# Plan 220 - 模块开发模板与规范文档

**文档编号**: 220
**标题**: 标准模块开发指南 - 为后续新模块提供参考
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 219（organization 重构）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

基于 organization 模块的重构经验，编写完整的模块开发模板文档，为后续的 workforce、contract 等新模块提供标准开发指南。

**关键交付物**:
- ✅ 模块结构模板说明
- ✅ sqlc 使用规范
- ✅ 事务性发件箱（Outbox）集成规范
- ✅ Docker 集成测试规范
- ✅ 样本模块代码（参考 organization）

### 1.2 为什么需要模块模板

- **一致性** - 所有新模块遵循相同的结构和规范
- **快速开发** - 开发者可快速启动新模块开发
- **知识转移** - 新成员易于理解项目架构
- **质量保证** - 标准化的质量检查清单

### 1.3 时间计划

- **计划完成**: Week 4 Day 1 (Day 15)
- **交付周期**: 1 天
- **负责人**: 架构师 + 文档支持
- **前置依赖**: Plan 219（organization 重构完成）

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
```
内部结构说明
- domain/：域模型和事件
- repository/：数据访问层
- service/：业务逻辑层
- handler/：REST 处理器
- resolver/：GraphQL 解析器
- api.go：公开接口
```

**示例代码**:
- api.go 的标准框架
- 接口定义的最佳实践
- 依赖注入模式

#### 第三章：数据访问层（sqlc 规范）

**内容**:
- 为什么使用 sqlc
- sqlc 配置示例
- 常见查询模式
- 与 repository 层的集成

**示例**:
```yaml
# sqlc.yaml 配置示例
version: "1"
packages:
  - name: "org"
    path: "internal/organization/internal/repository"
    queries: "./queries/"
    schema: "./schema/"
    engine: "postgresql"
```

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

从 organization 模块提取关键代码作为示例：
- models.go 示例
- repository.go 示例
- service.go 示例
- handler.go 示例
- resolver.go 示例

### 3.3 创建检查清单

为不同阶段提供检查清单：
- 模块结构检查清单
- API 契约检查清单
- 测试完成度检查清单
- 部署前检查清单

### 3.4 整合与审审查

- 架构师审查文档
- 后端 TL 检查示例代码准确性
- QA 验证测试规范
- 文档支持进行最终编辑

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
## 标准模块结构

所有业务模块应遵循以下目录结构：

internal/{module_name}/
├── api.go                 # 模块公开接口
├── internal/              # 私有实现
│   ├── domain/            # 域模型
│   ├── repository/        # 数据访问
│   ├── service/           # 业务逻辑
│   ├── handler/           # REST 处理器
│   ├── resolver/          # GraphQL 解析器
│   └── README.md
└── README.md              # 模块说明

### 各目录职责

- **domain/** - 聚合根、值对象、域事件的定义
- **repository/** - 数据持久化接口和实现（使用 sqlc）
- **service/** - 业务规则、工作流、事务管理
- **handler/** - HTTP 请求处理、响应格式化（REST API）
- **resolver/** - GraphQL 查询解析（查询服务）
- **api.go** - 模块对外暴露的接口（其他模块仅依赖此文件）
```

### 6.2 集成测试规范章节示例

```markdown
## Docker 集成测试规范

所有模块的集成测试应使用 Docker 容器化的 PostgreSQL。

### 配置步骤

1. 创建 docker-compose.test.yml
2. 运行 make test-db 启动容器
3. 执行 Goose 迁移
4. 运行集成测试
5. 验证测试数据状态

### 最佳实践

- 每个测试应该是独立的
- 使用事务进行测试隔离
- 清理测试数据（TRUNCATE）
- 记录慢查询日志
```

---

## 7. 验收标准

### 7.1 文档完整性

- [ ] 主文档（module-development-template.md）> 3000 字
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

- ✅ `docs/development-guides/module-development-template.md` （主文档）
- ✅ `docs/development-guides/examples/organization/` （示例代码）
- ✅ `docs/development-guides/checklists/module-structure-checklist.md`
- ✅ `docs/development-guides/checklists/api-contract-checklist.md`
- ✅ `docs/development-guides/checklists/testing-checklist.md`
- ✅ `docs/development-guides/checklists/deployment-checklist.md`
- ✅ 本计划文档（220）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**计划完成日期**: Week 4 Day 1 (Day 15)
