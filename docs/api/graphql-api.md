# GraphQL API文档 - 组织查询服务

## 概述

Cube Castle GraphQL API提供了强大的组织架构查询能力，支持灵活的数据获取和高性能缓存。

### 🚀 核心特性

- **灵活查询**: 支持复杂的嵌套查询和字段选择
- **高性能缓存**: Redis缓存支持，65%性能提升
- **实时数据**: 与Neo4j集成，提供实时组织架构数据
- **CQRS架构**: 专注查询操作，与命令服务分离
- **分页支持**: 高效的分页和搜索功能

### 📊 性能指标

| 查询类型 | 缓存MISS | 缓存HIT | 性能提升 |
|----------|----------|---------|----------|
| 组织列表 | 10.4ms | 3.6ms | **65%** |
| 单个组织 | 8.2ms | 2.1ms | **74%** |
| 统计查询 | 25.2ms | 8.1ms | **68%** |

## GraphQL端点

- **开发环境**: http://localhost:8090/graphql
- **GraphiQL界面**: http://localhost:8090/graphiql  
- **生产环境**: https://api.cubecastle.com/graphql

## Schema定义

### 类型系统

#### Organization类型

```graphql
type Organization {
  # 基本信息
  tenant_id: String!        # 租户ID
  code: String!             # 组织代码 (7位数字)
  parent_code: String       # 父组织代码
  name: String!             # 组织名称
  unit_type: String!        # 组织类型 (COMPANY, DEPARTMENT, TEAM等)
  status: String!           # 状态 (ACTIVE, INACTIVE, PLANNED)
  
  # 层级信息
  level: Int!               # 组织层级
  path: String              # 组织路径
  sort_order: Int           # 排序顺序
  
  # 描述信息
  description: String       # 组织描述
  profile: String           # 组织简介
  
  # 时间信息  
  created_at: String!       # 创建时间
  updated_at: String!       # 更新时间
  effective_date: String!   # 生效日期
  
  # 版本信息
  version: Int!             # 版本号
  is_current: Boolean!      # 是否当前版本
}
```

#### 查询类型

```graphql
type Query {
  # 组织列表查询
  organizations(
    first: Int = 50           # 查询数量 (默认50，最大100)
    offset: Int = 0           # 偏移量
    searchText: String        # 搜索文本 (支持名称和代码搜索)
  ): [Organization!]!
  
  # 单个组织查询
  organization(
    code: String!             # 组织代码 (必需)
  ): Organization
  
  # 组织统计查询
  organizationStats: OrganizationStats!
}
```

#### 统计类型

```graphql
type OrganizationStats {
  totalCount: Int!            # 组织总数
  byType: [TypeCount!]!       # 按类型统计
  byStatus: [StatusCount!]!   # 按状态统计
  byLevel: [LevelCount!]!     # 按层级统计
}

type TypeCount {
  unitType: String!           # 组织类型
  count: Int!                 # 数量
}

type StatusCount {
  status: String!             # 状态
  count: Int!                 # 数量
}

type LevelCount {
  level: String!              # 层级
  count: Int!                 # 数量
}
```

## 查询示例

### 1. 基本组织列表查询

**查询**:
```graphql
query GetOrganizations {
  organizations(first: 10, offset: 0) {
    code
    name
    unit_type
    status
    level
    parent_code
  }
}
```

**响应**:
```json
{
  "data": {
    "organizations": [
      {
        "code": "1000000",
        "name": "高谷集团",
        "unit_type": "COMPANY",
        "status": "ACTIVE",
        "level": 1,
        "parent_code": null
      },
      {
        "code": "1000001",
        "name": "AI治理办公室",
        "unit_type": "DEPARTMENT", 
        "status": "ACTIVE",
        "level": 2,
        "parent_code": "1000000"
      }
    ]
  }
}
```

### 2. 搜索组织

**查询**:
```graphql
query SearchOrganizations($searchText: String!) {
  organizations(searchText: $searchText) {
    code
    name
    unit_type
    description
    path
  }
}
```

**变量**:
```json
{
  "searchText": "AI"
}
```

**响应**:
```json
{
  "data": {
    "organizations": [
      {
        "code": "1000001",
        "name": "AI治理办公室",
        "unit_type": "DEPARTMENT",
        "description": "技术研发部门",
        "path": "/1000000/1000001"
      }
    ]
  }
}
```

### 3. 单个组织详细信息

**查询**:
```graphql
query GetOrganization($code: String!) {
  organization(code: $code) {
    tenant_id
    code
    name
    unit_type
    status
    level
    parent_code
    path
    description
    profile
    created_at
    updated_at
    effective_date
    version
    is_current
  }
}
```

**变量**:
```json
{
  "code": "1000001"
}
```

**响应**:
```json
{
  "data": {
    "organization": {
      "tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
      "code": "1000001",
      "name": "AI治理办公室",
      "unit_type": "DEPARTMENT",
      "status": "ACTIVE",
      "level": 2,
      "parent_code": "1000000",
      "path": "/1000000/1000001",
      "description": "技术研发部门",
      "profile": null,
      "created_at": "2025-08-05T11:23:01.426455Z",
      "updated_at": "2025-08-09T12:07:15.838099Z",
      "effective_date": "2025-08-10T00:00:00Z",
      "version": 1,
      "is_current": true
    }
  }
}
```

### 4. 组织统计查询

**查询**:
```graphql
query GetOrganizationStats {
  organizationStats {
    totalCount
    byType {
      unitType
      count
    }
    byStatus {
      status
      count
    }
    byLevel {
      level
      count
    }
  }
}
```

**响应**:
```json
{
  "data": {
    "organizationStats": {
      "totalCount": 25,
      "byType": [
        {
          "unitType": "COMPANY",
          "count": 1
        },
        {
          "unitType": "DEPARTMENT", 
          "count": 15
        },
        {
          "unitType": "TEAM",
          "count": 9
        }
      ],
      "byStatus": [
        {
          "status": "ACTIVE",
          "count": 23
        },
        {
          "status": "INACTIVE",
          "count": 2
        }
      ],
      "byLevel": [
        {
          "level": "级别1",
          "count": 1
        },
        {
          "level": "级别2", 
          "count": 15
        },
        {
          "level": "级别3",
          "count": 9
        }
      ]
    }
  }
}
```

### 5. 分页查询

**查询**:
```graphql
query GetOrganizationsPaginated($first: Int!, $offset: Int!) {
  organizations(first: $first, offset: $offset) {
    code
    name
    unit_type
    level
  }
}
```

**变量**:
```json
{
  "first": 5,
  "offset": 10
}
```

### 6. 选择特定字段 (GraphQL优势)

**查询**:
```graphql
query GetBasicOrganizationInfo {
  organizations {
    code
    name
    # 只获取需要的字段，节省带宽和处理时间
  }
}
```

## 缓存策略

### 缓存键生成

GraphQL查询的缓存键基于以下因素生成：
- 查询操作 (organizations/organization/organizationStats)
- 查询参数 (first, offset, searchText)
- 租户ID

```
缓存键格式: cache:<MD5哈希>
示例: cache:9c5dc0e19eb62bc1e3b0345db1e0871a
```

### 缓存TTL策略

| 查询类型 | TTL | 原因 |
|----------|-----|------|
| 组织列表 | 5分钟 | 频繁查询，变更相对较少 |
| 单个组织 | 5分钟 | 中等频率，详细信息 |
| 统计信息 | 5分钟 | 计算密集，变更不频繁 |

### 缓存失效

- **自动失效**: TTL到期自动失效
- **事件触发**: 组织变更事件触发相关缓存失效
- **手动清理**: 运维工具支持手动清理

## 错误处理

### 常见错误

#### 1. 组织不存在
```json
{
  "data": {
    "organization": null
  }
}
```

#### 2. 参数验证失败
```json
{
  "errors": [
    {
      "message": "Variable \"$code\" got invalid value \"123\"; Expected type String. String \"123\" does not match required pattern: ^[0-9]{7}$",
      "locations": [{"line": 1, "column": 22}]
    }
  ]
}
```

#### 3. 服务不可用
```json
{
  "errors": [
    {
      "message": "Internal server error",
      "extensions": {
        "code": "INTERNAL_ERROR"
      }
    }
  ]
}
```

## 性能优化建议

### 1. 字段选择优化
```graphql
# ✅ 好的做法 - 只获取需要的字段
query OptimizedQuery {
  organizations {
    code
    name
    # 只选择必要字段
  }
}

# ❌ 避免的做法 - 获取所有字段
query InefficiientQuery {
  organizations {
    tenant_id
    code
    parent_code
    name
    unit_type
    status
    level
    path
    sort_order
    description
    profile
    created_at
    updated_at
    effective_date
    version
    is_current
  }
}
```

### 2. 分页优化
```graphql
# ✅ 使用适当的分页大小
query PaginatedQuery {
  organizations(first: 20, offset: 0) {
    code
    name
  }
}
```

### 3. 搜索优化
```graphql
# ✅ 具体的搜索条件
query SpecificSearch {
  organizations(searchText: "AI治理") {
    code
    name
  }
}
```

## 集成示例

### JavaScript/TypeScript客户端

```typescript
// 使用Apollo Client
import { gql, useQuery } from '@apollo/client';

const GET_ORGANIZATIONS = gql`
  query GetOrganizations($first: Int, $offset: Int) {
    organizations(first: $first, offset: $offset) {
      code
      name
      unit_type
      status
    }
  }
`;

function OrganizationList() {
  const { loading, error, data } = useQuery(GET_ORGANIZATIONS, {
    variables: { first: 20, offset: 0 }
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <div>
      {data.organizations.map(org => (
        <div key={org.code}>
          <h3>{org.name}</h3>
          <p>代码: {org.code}</p>
          <p>类型: {org.unit_type}</p>
          <p>状态: {org.status}</p>
        </div>
      ))}
    </div>
  );
}
```

### Python客户端

```python
import requests

# GraphQL查询
query = """
query GetOrganization($code: String!) {
  organization(code: $code) {
    code
    name
    unit_type
    status
  }
}
"""

# 发送请求
response = requests.post(
    'http://localhost:8090/graphql',
    json={
        'query': query,
        'variables': {'code': '1000001'}
    }
)

data = response.json()
organization = data['data']['organization']
print(f"组织: {organization['name']}")
```

### Go客户端

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type GraphQLRequest struct {
    Query     string                 `json:"query"`
    Variables map[string]interface{} `json:"variables,omitempty"`
}

func queryOrganization(code string) error {
    query := `
    query GetOrganization($code: String!) {
      organization(code: $code) {
        code
        name
        unit_type
        status
      }
    }`

    req := GraphQLRequest{
        Query: query,
        Variables: map[string]interface{}{
            "code": code,
        },
    }

    jsonData, _ := json.Marshal(req)
    resp, err := http.Post(
        "http://localhost:8090/graphql",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    fmt.Printf("查询结果: %+v\n", result)
    return nil
}
```

## 开发工具

### GraphiQL

访问 http://localhost:8090/graphiql 使用交互式查询界面：

1. **查询编辑器**: 支持语法高亮和自动补全
2. **Schema探索**: 浏览完整的GraphQL Schema
3. **查询历史**: 保存和重复使用查询
4. **变量编辑**: 测试带变量的查询

### 查询验证

GraphQL提供强大的查询验证：
- **语法验证**: 查询语法错误检查
- **类型验证**: 字段类型匹配验证
- **Schema验证**: 字段存在性验证

## 监控和调试

### 性能监控
```graphql
# 在查询中添加性能标识
query GetOrganizations {
  organizations {
    code
    name
    # GraphQL查询会自动记录性能指标
  }
}
```

### 缓存状态检查
```bash
# 检查Redis缓存命中率
curl http://localhost:8090/metrics | grep cache
```

## 最佳实践

### 1. 查询设计
- 只获取需要的字段
- 使用合适的分页大小
- 避免深度嵌套查询

### 2. 缓存利用
- 相同查询会命中缓存
- 合理设置查询参数
- 监控缓存命中率

### 3. 错误处理
- 检查GraphQL响应的errors字段
- 处理null值情况
- 实现重试机制

这份文档涵盖了GraphQL API的完整使用方法，包括Schema定义、查询示例、性能优化和集成方法。