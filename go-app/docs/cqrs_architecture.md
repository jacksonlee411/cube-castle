# 员工管理模块CQRS架构文档

## 概述

本文档描述了员工管理模块从传统REST架构迁移到CQRS（Command Query Responsibility Segregation）架构的完整实现。

## 架构图

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   前端应用      │    │   API Gateway    │    │   CQRS层        │
│                 │    │                  │    │                 │
│ React + Zustand │◄──►│ Chi Router       │◄──►│ Commands/Queries│
│ CQRS Hooks      │    │ Middleware       │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                       ┌────────────────────────────────┼────────────────────┐
                       │                                │                    │
                       ▼                                ▼                    ▼
            ┌─────────────────┐              ┌─────────────────┐   ┌─────────────────┐
            │   Command Side  │              │   Event Bus     │   │   Query Side    │
            │                 │              │                 │   │                 │
            │ PostgreSQL      │              │ Domain Events   │   │ Neo4j           │
            │ (Write Store)   │              │ Event Consumers │   │ (Read Store)    │
            └─────────────────┘              └─────────────────┘   └─────────────────┘
                       │                                │                    ▲
                       │                                │                    │
                       └────────────── Events ─────────┴────────────────────┘
```

## 核心组件

### 1. 命令端 (Command Side)

**目的**: 处理所有写操作和业务逻辑

**技术栈**:
- PostgreSQL 作为事务数据存储
- CQRS Command Handlers
- Domain Events

**端点**:
```
POST /api/v1/commands/hire-employee        # 雇佣员工
PUT  /api/v1/commands/update-employee      # 更新员工信息
POST /api/v1/commands/terminate-employee   # 终止员工
```

**特点**:
- 强一致性
- 事务性操作
- 业务规则验证
- 事件发布

### 2. 查询端 (Query Side)

**目的**: 处理所有读操作，提供优化的查询性能

**技术栈**:
- Neo4j 图数据库
- CQRS Query Handlers
- 数据投影和预聚合

**端点**:
```
GET /api/v1/queries/employees              # 查询员工列表
GET /api/v1/queries/employees/{id}         # 查询单个员工
GET /api/v1/queries/organization-tree      # 组织架构树
GET /api/v1/queries/reporting-hierarchy    # 汇报关系
```

**特点**:
- 最终一致性
- 读取优化
- 复杂查询支持
- 缓存友好

### 3. 事件总线 (Event Bus)

**目的**: 在命令端和查询端之间传递领域事件

**组件**:
- Event Publishers
- Event Consumers
- Event Serialization
- Event Validation

**事件类型**:
```go
type EmployeeHired struct {
    TenantID   uuid.UUID
    EmployeeID uuid.UUID
    FirstName  string
    LastName   string
    Email      string
    HireDate   time.Time
}

type EmployeeUpdated struct {
    TenantID   uuid.UUID
    EmployeeID uuid.UUID
    Changes    map[string]interface{}
}

type EmployeeTerminated struct {
    TenantID        uuid.UUID
    EmployeeID      uuid.UUID
    TerminationDate time.Time
    Reason          string
}
```

## 数据流

### 写操作流程

1. **前端发起命令** → React组件调用CQRS Hook
2. **命令验证** → 业务规则验证和数据校验
3. **持久化** → 数据写入PostgreSQL
4. **事件发布** → 发布领域事件到事件总线
5. **事件消费** → 事件消费者更新Neo4j查询存储
6. **响应返回** → 命令执行结果返回给前端

### 读操作流程

1. **前端发起查询** → React组件调用CQRS Hook
2. **查询路由** → 查询请求路由到Query Handler
3. **Neo4j查询** → 从Neo4j图数据库查询优化数据
4. **结果返回** → 查询结果返回给前端
5. **状态更新** → Zustand状态管理器更新本地状态

## 数据一致性

### 强一致性 (Command Side)
- 使用PostgreSQL事务保证ACID特性
- 单个命令内的所有操作要么全成功要么全失败
- 业务规则在命令执行时强制验证

### 最终一致性 (Query Side)
- 通过事件驱动实现PostgreSQL到Neo4j的异步同步
- 事件消费者确保查询端数据最终与命令端一致
- 监控和告警机制检测数据不一致问题

## 性能优化

### 命令端优化
- 数据库连接池配置
- 索引优化
- 批量操作支持
- 事务超时控制

### 查询端优化
- Neo4j图查询优化
- 查询结果缓存
- 数据预聚合
- 索引策略

### 网络优化
- HTTP/2支持
- Gzip压缩
- 查询结果分页
- 连接复用

## 错误处理

### 命令错误处理
```go
type CommandError struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details"`
}
```

### 查询错误处理
```go
type QueryError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Source  string `json:"source"` // "neo4j", "cache", etc.
}
```

### 事件错误处理
- 事件发布失败重试机制
- 死信队列处理
- 事件幂等性保证
- 事件顺序保证

## 监控和运维

### 指标监控
- 命令执行时间和成功率
- 查询响应时间和缓存命中率
- 事件发布和消费延迟
- 数据库连接池状态
- 内存和CPU使用率

### 健康检查
- PostgreSQL连接状态
- Neo4j连接状态
- 事件总线状态
- 数据一致性检查

### 日志记录
- 结构化日志格式
- 请求跟踪ID
- 性能指标记录
- 错误堆栈跟踪

## API版本管理

### 版本策略
- URI版本控制：`/api/v1/commands/`, `/api/v1/queries/`
- 向后兼容性保证
- 废弃API优雅迁移
- 版本生命周期管理

### 迁移路径
```
旧端点: DELETE /api/v1/employees/{id}
新端点: POST /api/v1/commands/terminate-employee

旧端点: GET /api/v1/employees?search=张三
新端点: GET /api/v1/queries/employees?search=张三
```

## 安全考虑

### 认证授权
- JWT Token验证
- 基于角色的访问控制(RBAC)
- 租户隔离
- API速率限制

### 数据安全
- 数据库连接加密
- 敏感信息脱敏
- 审计日志记录
- 数据备份策略

## 部署和扩展

### 水平扩展
- 无状态API服务器
- 数据库读写分离
- 事件总线集群部署
- 负载均衡配置

### 高可用性
- 数据库主从复制
- 服务器集群部署
- 故障转移机制
- 灾难恢复计划

## 迁移策略

### 阶段化迁移
1. **Phase 1**: 启用查询端，保持写操作不变
2. **Phase 2**: 启用命令端，实现事件驱动同步
3. **Phase 3**: 废弃旧REST API，完成迁移

### 回滚计划
- 数据库回滚脚本
- API版本回退
- 配置开关控制
- 监控告警机制

## 开发指南

### 添加新命令
1. 定义Command结构体
2. 实现Command Handler
3. 添加事件发布
4. 创建路由绑定
5. 编写单元测试

### 添加新查询
1. 定义Query结构体
2. 实现Query Handler
3. 优化Neo4j查询
4. 创建路由绑定
5. 编写集成测试

### 事件扩展
1. 定义新事件类型
2. 实现事件序列化
3. 创建事件消费者
4. 更新同步逻辑
5. 验证数据一致性

## 最佳实践

### 代码质量
- 遵循DDD领域驱动设计
- 单一职责原则
- 依赖注入模式
- 错误处理标准化
- 代码审查流程

### 性能最佳实践
- 查询优化
- 缓存策略
- 连接池管理
- 资源监控
- 性能测试

### 运维最佳实践
- 自动化部署
- 监控告警
- 日志聚合
- 备份策略
- 灾难恢复

---

## 附录

### A. API端点完整列表
### B. 事件类型定义
### C. 错误代码说明
### D. 配置参数说明
### E. 性能基准测试结果