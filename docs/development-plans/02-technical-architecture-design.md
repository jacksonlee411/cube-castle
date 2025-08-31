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
```sql
CREATE TABLE organization_units (
    -- 复合主键支持多版本
    code VARCHAR(7),
    effective_date DATE NOT NULL,
    PRIMARY KEY (code, effective_date),
    
    -- 时态管理字段
    end_date DATE,
    is_current BOOLEAN NOT NULL DEFAULT true,
    is_future BOOLEAN NOT NULL DEFAULT false,
    record_id UUID NOT NULL DEFAULT gen_random_uuid(),
    
    -- 时态约束
    UNIQUE (code, effective_date, record_id),
    CHECK (end_date IS NULL OR end_date > effective_date),
    CHECK (NOT (is_current AND is_future))
);
```

## 📡 API设计

### GraphQL查询 (8090) - 时态查询能力
```graphql
type Query {
  # 基础查询 - 支持时态过滤
  organizations(
    filter: OrgFilter
    asOfDate: Date              # 时间点查询支持
  ): OrganizationConnection
  
  organization(
    code: String!
    asOfDate: Date              # 历史版本查询
  ): Organization
  
  # 时态专用查询
  organizationVersions(
    code: String!
    dateRange: DateRange        # 版本演进查询
  ): [Organization!]!
  
  organizationHierarchy(rootCode: String, maxDepth: Int): [OrganizationNode]
  organizationAuditHistory(code: String!): [AuditRecord]
  organizationStats: OrganizationStats
}

# 单表时态数据模型
type Organization {
  code: String!               # 业务标识 (不变)
  name: String!
  unitType: UnitType!
  status: OrganizationStatus!
  parentCode: String
  level: Int!
  
  # 时态管理字段 - 核心特性
  effectiveDate: String!      # 生效日期 (复合主键组成部分)
  endDate: String             # 结束日期 (可为空)
  isCurrent: Boolean!         # 当前版本标识
  isFuture: Boolean!          # 未来版本标识
  recordId: UUID!             # 版本唯一标识
  
  # 审计字段
  createdAt: DateTime!
  updatedAt: DateTime!
}

# 时态查询过滤器
input OrgFilter {
  # 标准过滤
  unitType: UnitType
  status: OrganizationStatus
  parentCode: String
  
  # 时态过滤 - 单表架构优势
  asOfDate: Date              # 指定时间点的有效版本
  includeFuture: Boolean      # 是否包含未来版本
  onlyFuture: Boolean         # 仅未来版本
  versionRange: DateRange     # 版本生效日期范围
}
```

### REST命令 (9090)
```yaml
端点:
  POST   /api/v1/organization-units          # 创建
  PUT    /api/v1/organization-units/{code}   # 替换
  PATCH  /api/v1/organization-units/{code}   # 更新
  DELETE /api/v1/organization-units/{code}   # 删除
  POST   /api/v1/organization-units/{code}/suspend    # 停用
  POST   /api/v1/organization-units/{code}/activate   # 激活

响应:
  成功: {success: true, data: {...}, message, timestamp, requestId}
  错误: {success: false, error: {code, message, details}, timestamp, requestId}
```

## 🔧 技术栈选型

### 后端技术栈 (Go语言生态)
```yaml
核心选型:
  Runtime: Go 1.21+ (编译型单一二进制部署)
  Language: Go + Generics支持 (静态类型系统)
  
GraphQL服务:
  框架: graph-gophers/graphql-go (Schema-first开发)
  特性: 原生Go实现，Schema定义驱动代码生成
  
REST服务:  
  框架: Gin 1.9+ (轻量级高性能Web框架)
  验证: validator/v10 + gin-binding
  文档: Swagger/OpenAPI 3.0自动生成
  
数据访问:
  驱动: jackc/pgx v5 (纯Go高性能PostgreSQL驱动)
  ORM: GORM v2 (关系映射) + SQLx (原生SQL)
  连接池: pgxpool，最大100连接，超时30秒
  
认证授权:
  协议: OAuth 2.0 Client Credentials Flow
  Token: JWT RS256签名，1小时有效期
  权限: PBAC模型，github.com/open-policy-agent/opa
  中间件: jwt-go + 自定义权限中间件
  
监控日志:
  指标: prometheus/client_golang
  日志: logrus/zap结构化JSON
  测试: Go内置testing + testify断言库
```

### 前端技术栈 (React生态)
```yaml
核心框架:
  UI: React 18+ + Canvas Kit v13 (Workday设计系统)
  语言: TypeScript 5+ (严格模式)
  路由: React Router v6
  
状态管理:
  数据: React Query (服务端状态) + Zustand (客户端状态)
  GraphQL: Apollo Client 3.x
  HTTP: Axios + 统一错误处理
  
开发工具:
  构建: Vite + TypeScript
  代码质量: ESLint + Prettier + Husky
  类型生成: GraphQL Schema → TypeScript Types
  
集成特性:
  API集成: GraphQL查询 + REST命令分离
  认证: JWT Token管理 + 权限检查
  错误处理: 统一企业级错误处理
  类型安全: 前后端TypeScript类型共享
```

### 技术选型原则
```yaml
选型优势:
  Go语言优势:
    - 单一二进制部署，无运行时依赖
    - 出色的并发性能和内存管理
    - 静态类型系统，编译时错误检查
    - 丰富的标准库和企业级库支持
    
  PostgreSQL优势:
    - 优秀的JSON/JSONB支持
    - 强大的递归CTE查询能力
    - 丰富的索引类型(GIN、GiST、BRIN)
    - 时态数据原生支持
    
  React生态优势:
    - Canvas Kit v13提供完整企业级组件
    - TypeScript提供类型安全保障
    - Apollo Client提供强大的GraphQL集成
    - 成熟的开发工具链和调试体验

技术债务预防:
  - 统一camelCase命名规范，无snake_case兼容负担
  - 严格CQRS架构，无协议混用历史问题
  - PostgreSQL单一数据源，无多数据库同步复杂性
  - 企业级响应结构从第一个端点开始统一
```

## 🔐 安全架构
```yaml
认证: OAuth 2.0 Client Credentials Flow
Token: JWT RS256签名，1小时有效期
权限: PBAC模型，17个细粒度权限，4种角色
审计: 完整操作日志，租户隔离
```

## 📊 监控
```yaml
指标: HTTP延迟、数据库连接、业务变更、系统资源
可视化: Grafana Dashboard + Prometheus
告警: Alertmanager + Slack/Email
日志: 结构化JSON，按日分割，30天保留
```

## 🚀 部署
```yaml
容器化: Docker + Kubernetes
环境: 开发(Docker Compose) + 测试(K8s) + 生产(K8s+Helm)
健康检查: /health端点，数据库连接验证
```

---
**更新**: 2025-08-23