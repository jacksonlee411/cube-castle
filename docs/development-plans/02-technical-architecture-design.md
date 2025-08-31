# 技术架构设计方案

## 🏗️ CQRS架构设计

### 核心原则
- **CQRS分离**: GraphQL查询(8090) + REST命令(9090)
- **单一数据源**: PostgreSQL 14+时态数据，无同步复杂性
- **性能目标**: 查询<200ms, 命令<300ms

### 服务架构
```yaml
前端: React + Canvas Kit v13 + TypeScript (3000)
查询: graph-gophers/graphql-go + pgx v5 (8090)
命令: Gin + GORM v2 + validator (9090)
认证: OAuth 2.0 + JWT + PBAC权限模型
数据: PostgreSQL 14.9+ + Redis缓存
监控: Prometheus + Grafana + logrus
```

## 🗄️ 数据库架构 ⭐ **单表时态架构**

### 数据存储选型
- **主数据库**: PostgreSQL 14.9+ 
- **缓存**: Redis (可选)
- **架构模式**: **单表多版本时态架构** (Single Table Temporal Database)
- **主键设计**: 复合主键 `(code, effective_date)` 支持多版本共存
- **时态字段**: `effective_date`, `end_date`, `is_current`, `is_future` 实现完整时态管理
- **索引策略**: 26个专用时态索引覆盖所有查询场景
- **审计模式**: 独立审计表，JSONB字段变更追踪

### 单表时态架构优势 🏆
1. **ACID事务一致性**: 版本切换原子操作，无同步问题
2. **查询性能优化**: 时态查询通过单表索引直接实现，避免复杂JOIN
3. **存储经济性**: 共享索引结构，消除重复存储
4. **开发简化**: 单一数据模型，统一CRUD逻辑

### 时态数据模型设计
- **复合主键**: `(code, effective_date)` 支持多版本共存
- **时态管理字段**: `effective_date`, `end_date`, `is_current`, `is_future`
- **审计字段**: `record_id` 提供版本唯一标识
- **时态约束**: 确保日期逻辑一致性和状态互斥性

*具体表结构定义请参考API契约文档中的数据模型规范*

## 📡 API设计

### GraphQL查询 (8090) - 时态查询能力
- **基础查询**: 支持组织单元列表和单个查询，带时态过滤 `asOfDate` 参数
- **层级查询**: 组织架构层级查询和子树查询，支持深度控制
- **统计查询**: 组织统计信息，包含时态分布统计
- **审计查询**: 基于 `recordId` 的精确审计历史查询
- **时态过滤**: 支持历史版本、未来版本、日期范围等时态查询

*具体GraphQL Schema定义请参考 `/docs/api/schema.graphql` 文件*

### REST命令 (9090) 
- **CRUD操作**: 创建、完整替换、部分更新、删除组织单元
- **业务操作**: 专用的停用/激活端点，确保业务逻辑清晰
- **数据验证**: 专用验证端点，支持干运行模式
- **系统维护**: 层级修复和批量操作端点
- **企业级响应**: 统一的信封格式，包含成功/错误状态、数据、消息、时间戳、请求ID

*具体REST API端点定义请参考 `/docs/api/openapi.yaml` 文件*

## 🔧 技术栈选型

### 后端技术栈 (Go语言生态)
- **运行环境**: Go 1.21+ 编译型单一二进制部署，支持泛型
- **GraphQL服务**: graph-gophers/graphql-go，Schema定义驱动开发
- **REST服务**: Gin 1.9+，高性能Web框架，集成validator和OpenAPI
- **数据访问**: pgx v5驱动 + GORM v2/SQLx，连接池优化
- **认证授权**: OAuth 2.0 + JWT RS256 + PBAC权限模型
- **监控日志**: Prometheus指标 + 结构化JSON日志

### 前端技术栈 (React生态)
- **核心框架**: React 18+ + Canvas Kit v13 (Workday设计系统) + TypeScript 5+
- **状态管理**: React Query (服务端) + Zustand (客户端) + Apollo Client (GraphQL)
- **开发工具**: Vite构建 + ESLint/Prettier + GraphQL类型生成
- **集成特性**: GraphQL查询/REST命令分离 + JWT认证 + 统一错误处理

### 技术选型原则
- **Go语言优势**: 单一二进制部署、并发性能优秀、静态类型安全、企业级库支持
- **PostgreSQL优势**: JSON/JSONB支持、递归CTE查询、丰富索引类型、时态数据原生支持  
- **React生态优势**: Canvas Kit企业级组件、TypeScript类型安全、成熟工具链
- **技术债务预防**: 统一命名规范、严格CQRS架构、单一数据源、企业级响应结构

## 🔐 安全架构
- **认证**: OAuth 2.0 Client Credentials Flow
- **Token**: JWT RS256签名，1小时有效期  
- **权限**: PBAC模型，17个细粒度权限，4种角色
- **审计**: 完整操作日志，租户隔离

## 📊 监控
- **指标**: HTTP延迟、数据库连接、业务变更、系统资源
- **可视化**: Grafana Dashboard + Prometheus
- **告警**: Alertmanager + Slack/Email  
- **日志**: 结构化JSON，按日分割，30天保留

## 🚀 部署
- **容器化**: Docker + Kubernetes
- **环境**: 开发(Docker Compose) + 测试(K8s) + 生产(K8s+Helm)
- **健康检查**: /health端点，数据库连接验证

---
**更新**: 2025-08-23