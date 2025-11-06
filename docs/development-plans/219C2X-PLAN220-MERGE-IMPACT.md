# Plan 220 对服务合并的具体影响分析

> **分析日期**: 2025-11-06
> **关联计划**: Plan 219 (组织模块重构) | Plan 202 (架构建议) | 合并工作

---

## 一、Plan 220 是什么

**Plan 220**: 模块开发模板与规范文档

### 核心目标
基于 Plan 219 (organization 重构) 的经验，编写完整的**模块开发标准指南**，为后续新模块和合并工作提供参考框架。

### 关键交付物
- ✅ 模块结构模板说明
- ✅ sqlc 使用规范
- ✅ 事务性发件箱（Outbox）集成规范
- ✅ Docker 集成测试规范
- ✅ 质量检查清单
- ✅ 样本代码（提取自 organization）

---

## 二、Plan 220 对合并的关键影响

### 2.1 影响 1: 定义合并后的"标准架构" 🎯 **最关键**

#### 问题: 合并前需要明确目标
```
当前状态：
  cmd/organization-command-service/  (REST 命令)
  cmd/organization-query-service/    (GraphQL 查询)
  └─ 两个独立的 Go 二进制

目标状态：
  cmd/hrms-server/
  ├── command/      (REST 层)
  ├── query/        (GraphQL 层)
  └── internal/organization/
      ├── api.go           ← 公开接口
      ├── internal/
      │   ├── domain/      ← 域模型
      │   ├── repository/  ← 数据访问
      │   ├── service/     ← 业务逻辑
      │   ├── handler/     ← REST 处理器
      │   └── resolver/    ← GraphQL 解析器
      └── README.md

但如何组织? 在哪里放? 怎样避免重复代码?
```

#### Plan 220 的解决方案
**第二章: 模块结构模板** 规定了标准模块结构:

```
internal/{module_name}/
├── api.go                 # 模块公开接口 ← 关键!
├── internal/              # 私有实现 (内部可见)
│   ├── domain/            # 域模型、事件
│   ├── repository/        # 数据访问 (sqlc)
│   ├── service/           # 业务逻辑
│   ├── handler/           # REST 处理器
│   ├── resolver/          # GraphQL 解析器
│   └── README.md
└── README.md              # 模块说明
```

#### 对合并的影响 📍
**合并时必须遵循这个结构**:
1. 命令服务的 repository → 统一到 `service/repository/` (sqlc)
2. 命令服务的 handler → 保留在 `service/handler/` (REST)
3. 查询服务的 repository → 统一到 `service/repository/` (查询)
4. 查询服务的 resolver → 保留在 `service/resolver/` (GraphQL)

**关键约束**: `api.go` 作为**唯一的公开接口**
- 其他模块只能依赖 `api.go`
- 不允许直接导入 `internal/` 下的任何包
- 这是**编译时强制**的(Go 的 internal/ 机制)

---

### 2.2 影响 2: 明确代码重复消除的方案 🔁

#### 问题: 命令/查询服务有重复的代码
```
当前重复:
  ✗ 权限验证逻辑 (PBAC 在两个服务中)
  ✗ 业务验证逻辑 (organization rules 在两个服务中)
  ✗ 数据访问逻辑 (部分 SQL 和映射在两个服务中)
  ✗ 异常处理逻辑
  ✗ 日志记录逻辑

如何消除这些重复，但不破坏单一职责?
```

#### Plan 220 的规范
**第四章: 事务性发件箱集成** + **第二章: 模块结构**

共同定义了**分层架构中的共享逻辑放置**:

```go
// ✅ 标准模式 (来自 Plan 220)

// 1. 共享业务逻辑 → service/ 层
internal/organization/internal/service/
├── organization_service.go       // 业务逻辑 (被 handler + resolver 共用)
├── create_organization.go        // 创建命令逻辑
├── query_organization.go         // 查询逻辑
└── event_publisher.go            // 事件发布 (使用 outbox)

// 2. 命令处理 → handler/ 层
internal/organization/internal/handler/
├── create_handler.go    // REST 处理 (调用 service)
├── update_handler.go
└── delete_handler.go

// 3. 查询处理 → resolver/ 层
internal/organization/internal/resolver/
├── organization_resolver.go  // GraphQL 处理 (调用 service)
├── stats_resolver.go
└── hierarchy_resolver.go

// 4. 数据访问 → repository/ 层 (使用 sqlc)
internal/organization/internal/repository/
├── organization_repository.go
├── queries/
│   ├── organization.sql
│   └── *.sql
└── schema/
    └── schema.sql
```

#### 对合并的影响 📍
**合并时必须进行代码重组织**:

| 当前状态 | 合并后位置 | 逻辑 |
|---------|-----------|------|
| command/service 中的验证逻辑 | service/organization_service.go | 共享,两边调用 |
| command/handler 中的 REST 端点 | handler/ 保留不变 | 仅处理 HTTP |
| query/resolver 中的 GraphQL 逻辑 | resolver/ 保留不变 | 仅处理 GraphQL |
| 两边的 permission check | service/ 中的 API 层 | 单一实现 |

**核心原则**: service/ 层处理**所有业务逻辑**, handler/resolver 仅处理**协议相关**

---

### 2.3 影响 3: 提供质量检查清单 ✅

#### Plan 220 的检查清单
合并完成后，需要验证:

```markdown
## 模块结构检查清单

- [ ] api.go 定义了模块的公开接口
- [ ] 所有 internal/ 代码都在私有包中
- [ ] 没有跨模块的 internal/ 导入
- [ ] domain/ 包含所有域模型
- [ ] repository/ 使用 sqlc 生成的代码
- [ ] service/ 包含所有业务逻辑
- [ ] handler/ 仅处理 HTTP 请求转换
- [ ] resolver/ 仅处理 GraphQL 请求转换

## API 契约检查清单

- [ ] 所有 REST 端点在 OpenAPI 中定义
- [ ] 所有 GraphQL 查询在 schema.graphql 中定义
- [ ] 权限 scopes 在 OpenAPI 中声明
- [ ] 错误响应格式统一

## 测试完成度检查清单

- [ ] 单元测试覆盖率 > 80%
- [ ] 集成测试使用 Docker PostgreSQL
- [ ] Outbox dispatcher 有测试
- [ ] 错误路径有测试

## 部署前检查清单

- [ ] 所有类型检查通过 (go vet)
- [ ] 所有测试绿灯 (go test)
- [ ] 代码格式化统一 (go fmt)
- [ ] Linter 无警告 (golangci-lint)
```

#### 对合并的影响 📍
**合并完成后，需要逐项验证这份清单**
- 如果有失败项，说明合并不完整
- 这是合并**验收的关键指标**

---

### 2.4 影响 4: 文件迁移和重组织的具体指导

#### Plan 220 的代码示例
从 organization 模块提取并展示:
- `models.go` 示例 → 告诉你哪些代码应该在 domain/
- `repository.go` 示例 → 告诉你如何使用 sqlc
- `service.go` 示例 → 告诉你如何调用 repository 和发送事件
- `handler.go` 示例 → 告诉你如何仅做 HTTP 处理
- `resolver.go` 示例 → 告诉你如何仅做 GraphQL 处理

#### 对合并的影响 📍
**合并时可以直接参考这些示例**:

```
步骤 1: 查看 Plan 220 的 handler.go 示例
        ↓
步骤 2: 对比你的命令服务 handler
        ↓
步骤 3: 重构成 Plan 220 的风格
        ↓
步骤 4: 验证与查询服务 resolver 没有重复逻辑
```

---

## 三、Plan 220 的缺陷与风险

### 3.1 缺陷 1: 不提供"合并步骤"本身

**问题**:
```
Plan 220 说"应该怎样"组织代码
但没说"如何从当前架构迁移到目标架构"
```

**缺失的内容**:
- ❌ 两个服务如何物理合并?
- ❌ 导入路径如何统一?
- ❌ 端口如何处理(9090 vs 8090)?
- ❌ 逐步迁移还是一次性合并?
- ❌ 回滚计划是什么?

**影响**: 需要**额外编制合并计划** (建议在 Plan 220 完成后编制新的 Plan 223)

---

### 3.2 缺陷 2: 假设 organization 已经重构完成

**问题**:
```
Plan 220 依赖 Plan 219 完成
但如果 219 延期，220 就没有参考样例
```

**风险**: 如果 219 + 220 延期太长，会推迟合并时间

**缓解**: 保持 219 → 220 的紧密协调

---

### 3.3 缺陷 3: 对"双协议共存"期的指导不足

**问题**:
```
Plan 220 定义的结构基于"一个模块同时提供 REST + GraphQL"
但合并后需要支持两个端口(9090/8090)长期共存

怎样在两个端口都暴露同样的 service/resolver?
```

**需要补充**:
- handler/ 和 resolver/ 如何分别导入同一个 service/?
- main.go 如何启动两个监听?
- 路由如何区分?

---

## 四、Plan 220 对合并时机的影响

### 4.1 时间表影响

```
时间点          工作                   Plan 220 的影响
═════════════════════════════════════════════════════════
2025-11-08    开始合并试点           ❌ 没有标准可参考
              (如果没有 220)         → 高风险,代码混乱

2025-11-08    开始合并试点           ✅ 有清晰的标准
              (如果有 220)           → 风险降低,速度快

2025-11-09    Docker 测试基座        Plan 220 的检查清单
              验证                   告诉你什么是"正确"
                                     ↓
                                     知道如何验证合并
```

### 4.2 质量影响

```
没有 Plan 220              有 Plan 220
───────────────────────────────────────────
❌ 代码组织混乱           ✅ 结构清晰一致
❌ 逻辑重复               ✅ 逻辑共享
❌ 不知道什么是"完成"     ✅ 有验收清单
❌ 新人难以理解           ✅ 有文档和示例
❌ 后续维护困难           ✅ 标准化易维护
```

---

## 五、建议: Plan 220 对合并的强制要求

### 5.1 必须在合并前完成的部分
```
✅ 模块结构模板 (第二章)
   → 告诉你文件应该放哪里

✅ 质量检查清单 (第八章)
   → 告诉你如何验证合并完成
```

### 5.2 可以在合并中/后完成的部分
```
⚠️ 样本代码 (第 3.2 节)
   → 可以边合并边提取示例
   → 不必须是完整的文档

⚠️ Docker 集成测试规范 (第五章)
   → 合并后进行测试
   → Plan 221 可以与此并行
```

### 5.3 建议补充的部分 (新增 Plan 223)
```
📋 合并执行计划 (Plan 223)
   - 逐步迁移流程
   - 端口管理策略
   - 双协议共存期的指导
   - 验收标准
```

---

## 六、总结表: Plan 220 的影响矩阵

| 维度 | 没有 Plan 220 | 有 Plan 220 | 影响程度 |
|------|-------------|-----------|---------|
| **代码组织** | 🔴 混乱 | ✅ 清晰 | **关键** |
| **逻辑共享** | 🔴 重复 | ✅ 高效 | **关键** |
| **质量验收** | 🔴 无标准 | ✅ 有清单 | **关键** |
| **文档完整性** | 🔴 缺失 | ✅ 完整 | **重要** |
| **新人上手** | 🔴 困难 | ✅ 快速 | **重要** |
| **合并时间** | 🔴 长 | ✅ 短 | **中等** |
| **风险** | 🔴 高 | ✅ 低 | **关键** |

---

## 七、最终结论

### Plan 220 的角色
**Plan 220 是合并的"蓝图"** - 没有它:
- ❌ 不知道目标架构是什么
- ❌ 不知道如何组织代码
- ❌ 不知道如何验证完成

### 对原合并时机的影响

| 原建议 | Plan 220 影响 | 新建议 |
|--------|-------------|--------|
| 2025-11-08 开始合并 | 需要 Plan 220 | **2025-11-08 仍可开始,前提是 Plan 220 与 Plan 219 同步完成** |

### 关键依赖关系

```
Plan 219 (重构)
    ↓
Plan 220 (标准文档)  ← 合并的"蓝图"
    ↓
Plan 223 (合并计划)  ← 建议新增
    ↓
实际合并工作          ← 使用 220 的标准
    ↓
Plan 221 (Docker测试) ← 验证 220 的检查清单
```

### 推荐行动

**立即** (2025-11-06-07):
1. 跟进 Plan 219 进度
2. 在 219 完成时，快速完成 Plan 220 主要章节(1-3章)

**2025-11-08**:
1. Plan 220 的模块结构+检查清单必须完成
2. 开始合并试点,严格遵循 220 的标准

**2025-11-09-10**:
1. 用 Plan 220 的检查清单验证合并结果
2. 编制 Plan 223 (合并执行总结)

---

**总体评价**: Plan 220 **对合并至关重要**,是合并的"指南针"。必须确保在合并前的关键部分完成。
