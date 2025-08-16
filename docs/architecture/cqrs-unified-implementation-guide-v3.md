# CQRS统一实施指南 - 现代化简洁架构

## 文档信息
- **版本**: v4.0 (现代化简洁版)
- **更新日期**: 2025-08-09
- **适用阶段**: 当前生产环境
- **核心原则**: REST API用于CUD，GraphQL用于R，避免过度设计

## 架构总览

### 简洁CQRS原则
基于CLAUDE.md中已完成的Phase 5+优化成果，我们采用**简洁的CQRS实现**：

- ✅ **查询(R)**: 统一使用GraphQL
- ✅ **命令(CUD)**: 统一使用REST API  
- ❌ **不重复实现**: 避免同一功能的多种API实现
- ❌ **不过度设计**: 移除复杂的降级和路由机制

## 当前架构图

```
                    前端应用 (React)
                         │
                         ▼
              ┌─────────────────────┐
              │     简洁分离        │
              │                     │
     GraphQL  │                     │  REST
     查询请求 │                     │  命令请求
              │                     │
              ▼                     ▼
    ┌─────────────┐         ┌─────────────┐
    │   查询服务   │         │   命令服务   │
    │  (Port:8090) │         │ (Port:9090)  │
    │   GraphQL    │         │   REST API   │
    └──────┬──────┘         └──────┬──────┘
           │                       │
           ▼                       ▼
    ┌─────────────┐         ┌─────────────┐
    │    Neo4j    │◄────────┤ PostgreSQL  │
    │  (查询优化)  │   CDC   │  (命令端)   │
    │    缓存     │  同步    │   主存储    │
    └─────────────┘         └─────────────┘
```

## 核心服务架构

### 1. 命令服务 (Command Service)

**位置**: `/cmd/organization-command-service/main.go`  
**端口**: 9090  
**职责**: 处理所有CUD操作  
**协议**: REST API  

#### API端点
```bash
# 创建组织
POST /api/v1/organization-units
Content-Type: application/json

# 更新组织  
PUT /api/v1/organization-units/{code}
Content-Type: application/json

# 删除组织
DELETE /api/v1/organization-units/{code}

# 健康检查
GET /health

# 监控指标
GET /metrics
```

#### 数据流
```
前端 → REST API → 命令服务 → PostgreSQL → CDC事件 → 同步服务
```

### 2. 查询服务 (Query Service)

**位置**: `/cmd/organization-query-service-unified/main.go`  
**端口**: 8090  
**职责**: 处理所有查询操作  
**协议**: GraphQL  

#### GraphQL Schema
```graphql
type Organization {
    code: String!
    name: String!
    unitType: String! 
    status: String!
    level: Int!
    path: String!
    parentCode: String
    sortOrder: Int!
    description: String!
    createdAt: String!
    updatedAt: String!
}

type OrganizationStats {
    totalCount: Int!
    activeCount: Int!
    inactiveCount: Int!
    byType: [TypeStat!]!
}

type TypeStat {
    type: String!
    count: Int!
}

type Query {
    # 获取所有组织
    organizations(first: Int, offset: Int): [Organization!]!
    
    # 根据代码获取单个组织
    organization(code: String!): Organization
    
    # 获取统计信息
    organizationStats: OrganizationStats!
}
```

#### GraphQL查询示例
```graphql
# 查询组织列表
query GetOrganizations {
  organizations(first: 50, offset: 0) {
    code
    name
    unitType
    status
    level
    description
  }
}

# 查询单个组织
query GetOrganization($code: String!) {
  organization(code: $code) {
    code
    name
    unitType
    status
    parentCode
    level
    path
    description
  }
}

# 查询统计信息
query GetStats {
  organizationStats {
    totalCount
    activeCount
    inactiveCount
    byType {
      type
      count
    }
  }
}
```

#### 数据流
```
前端 → GraphQL → 查询服务 → Neo4j缓存 → 响应数据
```

### 3. 数据同步服务 (Sync Service)

**位置**: `/cmd/organization-sync-service/main.go`  
**职责**: PostgreSQL → Neo4j实时同步  
**机制**: 基于成熟的Debezium CDC  

#### 同步流程
```
PostgreSQL变更 → Debezium CDC → Kafka → 同步服务 → Neo4j → 缓存失效
```

## 前端集成模式

### API客户端设计

```typescript
// 查询操作 - 统一使用GraphQL
export const organizationQueries = {
  // 获取列表
  getAll: async (): Promise<Organization[]> => {
    const response = await graphqlClient.query({
      query: GET_ORGANIZATIONS,
    });
    return response.data.organizations;
  },
  
  // 获取单个
  getByCode: async (code: string): Promise<Organization | null> => {
    const response = await graphqlClient.query({
      query: GET_ORGANIZATION,
      variables: { code }
    });
    return response.data.organization;
  },
  
  // 获取统计
  getStats: async (): Promise<OrganizationStats> => {
    const response = await graphqlClient.query({
      query: GET_ORGANIZATION_STATS,
    });
    return response.data.organizationStats;
  }
};

// 命令操作 - 统一使用REST API
export const organizationCommands = {
  // 创建
  create: async (input: CreateOrganizationInput): Promise<Organization> => {
    const response = await fetch('http://localhost:9090/api/v1/organization-units', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input)
    });
    return response.json();
  },
  
  // 更新
  update: async (code: string, input: UpdateOrganizationInput): Promise<Organization> => {
    const response = await fetch(`http://localhost:9090/api/v1/organization-units/${code}`, {
      method: 'PUT', 
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input)
    });
    return response.json();
  },
  
  // 删除
  delete: async (code: string): Promise<void> => {
    await fetch(`http://localhost:9090/api/v1/organization-units/${code}`, {
      method: 'DELETE'
    });
  }
};
```

### React组件集成

```typescript
// 查询组件示例
const OrganizationList: React.FC = () => {
  const { data: organizations, loading, error } = useQuery(GET_ORGANIZATIONS);
  
  if (loading) return <Loading />;
  if (error) return <Error message={error.message} />;
  
  return (
    <div>
      {organizations?.map(org => (
        <OrganizationCard key={org.code} organization={org} />
      ))}
    </div>
  );
};

// 命令组件示例
const CreateOrganizationForm: React.FC = () => {
  const [createOrganization] = useMutation(organizationCommands.create);
  
  const handleSubmit = async (data: CreateOrganizationInput) => {
    try {
      await createOrganization(data);
      // 成功后可以refetch查询或更新缓存
      refetchOrganizations();
    } catch (error) {
      showErrorMessage(error.message);
    }
  };
  
  return <OrganizationForm onSubmit={handleSubmit} />;
};
```

## 数据一致性保证

### 事务边界
- **命令端**: PostgreSQL事务保证ACID特性
- **查询端**: 最终一致性，通过CDC同步保证
- **缓存策略**: 精确失效，避免cache:*暴力清空

### 一致性模式
```
写操作: 前端 → REST API → PostgreSQL (强一致性)
同步: PostgreSQL → CDC → Kafka → Neo4j (最终一致性)  
读操作: 前端 → GraphQL → Neo4j缓存 (高性能)
```

### 租户隔离
- **单租户设计**: 当前使用默认租户ID `3b99930c-4dc6-4cc9-8e4d-7d960a931cb9`
- **数据隔离**: 所有服务统一使用相同租户ID
- **扩展准备**: 架构支持多租户，应用层可配置

## 部署和运维

### 标准部署流程

```bash
# 1. 启动基础设施
docker-compose up -d  # PostgreSQL, Neo4j, Redis, Kafka

# 2. 启动命令服务
cd cmd/organization-command-service
go run main.go &  # 端口9090

# 3. 启动查询服务  
cd cmd/organization-query-service-unified
go run main.go &  # 端口8090

# 4. 启动同步服务
cd cmd/organization-sync-service
go run main.go &

# 5. 验证服务状态
curl http://localhost:9090/health  # 命令服务
curl http://localhost:8090/health  # 查询服务
```

### 监控检查

```bash
# 服务健康检查
curl http://localhost:9090/health
curl http://localhost:8090/health

# 性能指标
curl http://localhost:9090/metrics
curl http://localhost:8090/metrics

# GraphQL查询测试
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizations { code name } }"}'

# REST命令测试
curl -X GET http://localhost:9090/api/v1/organization-units
```

### 性能基准

| 操作类型 | 端点 | 期望响应时间 | 实际性能 |
|---------|------|-------------|----------|
| GraphQL查询 | :8090/graphql | <50ms | ~15-30ms |
| REST命令 | :9090/api/v1/* | <100ms | ~25-50ms |
| 数据同步 | CDC Pipeline | <2s | ~1s |

## 故障排查指南

### 常见问题

1. **查询服务无响应**
```bash
# 检查Neo4j连接
docker exec cube_castle_neo4j cypher-shell "MATCH (n) RETURN count(n) LIMIT 1"

# 检查查询服务日志
tail -f logs/query-service.log

# 重启查询服务
pkill -f "organization-query-service" && cd cmd/organization-query-service-unified && go run main.go &
```

2. **命令服务无响应**
```bash  
# 检查PostgreSQL连接
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1"

# 检查命令服务日志
tail -f logs/command-service.log

# 重启命令服务
pkill -f "organization-command-service" && cd cmd/organization-command-service && go run main.go &
```

3. **数据同步延迟**
```bash
# 检查Debezium连接器状态
curl http://localhost:8083/connectors/organization-postgres-connector/status

# 检查Kafka消息
docker exec cube_castle_kafka kafka-topics.sh --list --bootstrap-server localhost:9092

# 重启同步服务
pkill -f "organization-sync-service" && cd cmd/organization-sync-service && go run main.go &
```

## 最佳实践

### 开发规范

1. **严格协议分离**
```typescript
// ✅ 正确: 查询使用GraphQL
const organizations = await graphqlClient.query({ query: GET_ORGANIZATIONS });

// ❌ 错误: 查询使用REST API
const organizations = await fetch('/api/v1/organization-units');

// ✅ 正确: 命令使用REST API
await fetch('/api/v1/organization-units', { method: 'POST', ... });

// ❌ 错误: 命令使用GraphQL Mutation
await graphqlClient.mutate({ mutation: CREATE_ORGANIZATION });
```

2. **错误处理**
```typescript
// GraphQL错误处理
try {
  const result = await graphqlClient.query({ query: GET_ORGANIZATIONS });
  return result.data.organizations;
} catch (error) {
  console.error('GraphQL查询失败:', error);
  throw new Error('获取组织列表失败');
}

// REST API错误处理
const response = await fetch('/api/v1/organization-units', { method: 'POST', ... });
if (!response.ok) {
  const error = await response.json();
  throw new Error(error.message || '创建组织失败');
}
```

3. **缓存管理**
```typescript
// 命令操作后刷新查询缓存
const [createOrganization] = useMutation(organizationCommands.create, {
  onCompleted: () => {
    // 刷新GraphQL缓存
    client.cache.evict({ fieldName: 'organizations' });
    client.cache.gc();
  }
});
```

### 运维规范

1. **监控告警**
- GraphQL查询响应时间 >100ms 告警
- REST命令响应时间 >200ms 告警  
- 数据同步延迟 >5秒 告警
- 服务可用性 <99% 告警

2. **容量规划**
- 查询服务: 支持1000+ QPS
- 命令服务: 支持100+ QPS
- 同步延迟: <2秒

3. **备份策略**
- PostgreSQL: 每日全量备份 + WAL归档
- Neo4j: 每日数据导出备份
- 配置文件: 版本控制管理

## 移除的过度设计

### 已移除的复杂特性

1. ❌ **智能路由网关**: 增加复杂性，违反简洁原则
2. ❌ **GraphQL降级机制**: 过度设计，增加故障点
3. ❌ **多协议支持**: 造成API混乱，增加维护成本
4. ❌ **复杂健康检查**: 简化为基本的服务状态检查
5. ❌ **路由统计**: 移除不必要的统计复杂度

### 简化的理由

- **降低复杂度**: 每个服务专注单一职责
- **提高可靠性**: 减少中间层，降低故障概率
- **简化运维**: 直接的服务部署，清晰的问题定位
- **提升性能**: 移除不必要的路由开销
- **易于理解**: 开发团队更容易掌握和维护

## 结论

现代化的CQRS实现遵循**简洁、直接、高效**的原则：

✅ **协议清晰**: REST用于CUD，GraphQL用于R  
✅ **架构简洁**: 2个核心服务，职责分明  
✅ **性能优异**: 查询<50ms，命令<100ms，同步<2s  
✅ **易于维护**: 移除过度设计，降低复杂度  
✅ **生产就绪**: 基于实际运行经验的成熟架构  

这种简洁的CQRS架构为组织管理系统提供了高性能、高可用、易维护的技术基础，避免了过度工程化的陷阱。

---

**维护原则**: 保持简洁，避免过度设计，专注业务价值  
**技术支持**: 基于CLAUDE.md项目记忆文档  
**版本控制**: 随项目演进持续更新