# 📚 Cube Castle API规范文档

**版本**: v4.2.2 ⭐ **OAuth特例修复版**  
**架构**: 严格CQRS + PostgreSQL单一数据源 + OAuth 2.0企业级安全  
**状态**: ✅ Single Source of Truth (唯一权威来源)  
**重要更新**: 修复OAuth认证字段名特例，解决组织列表获取失败问题  

## 🚀 概述

本目录包含Cube Castle组织架构管理系统的**完整API规范文档**，采用行业标准格式提供权威的API接口定义。这些文件是API设计、开发、测试和文档生成的**唯一真实来源**。

### 🏗️ 架构特点

- **严格CQRS架构**: 查询使用GraphQL (端口8090)，命令使用REST API (端口9090)
- **PostgreSQL原生**: 单一数据源架构，查询响应时间1.5-8ms
- **企业级安全**: OAuth 2.0 Client Credentials Flow + PBAC权限模型
- **时态数据支持**: 完整的历史版本管理和未来生效计划
- **17级层级管理**: 智能级联更新 + 双路径系统

## 📋 文件清单

### 🔧 核心规范文件

| 文件 | 格式 | 作用域 | 描述 |
|------|------|--------|------|
| **`openapi.yaml`** | OpenAPI 3.0.3 | REST API命令操作 | 11个REST端点的完整规范 |
| **`schema.graphql`** | GraphQL SDL | GraphQL查询操作 | 10个GraphQL查询的完整Schema |

### 📖 支持文档

| 文件 | 描述 |
|------|------|
| `README.md` | 本文件 - API文档使用指南 |
| `CHANGELOG.md` | API版本变更历史记录 |

## 🎯 Single Source of Truth 原则

### ✅ 权威性保证

这些规范文件是API开发的**唯一权威来源**，所有相关工作必须基于这些文件：

- **API开发**: 后端实现必须严格遵循规范
- **前端集成**: 客户端开发基于规范进行集成
- **测试验证**: 所有API测试基于规范执行
- **文档生成**: 自动化文档生成从规范文件提取

### 🔄 变更管理流程

**重要**: 任何API变更都必须遵循以下严格流程：

1. **规范先行**: 先修改 `openapi.yaml` 或 `schema.graphql`
2. **版本更新**: 更新版本号并记录到 `CHANGELOG.md`
3. **代码实现**: 基于更新后的规范修改代码实现
4. **测试验证**: 验证实现与规范的一致性
5. **文档同步**: 自动化更新相关文档

❌ **禁止行为**:
- 先改代码再更新规范
- 规范与实现不一致
- 绕过版本管理直接修改API

## 🌟 核心特性

### API服务架构

| 服务 | 端点 | 协议 | 用途 | 缓存性能 |
|------|------|------|------|----------|
| **GraphQL查询** | `localhost:8090/graphql` | GraphQL | 灵活查询、统计 | 65%↗️ |
| **时态API** | `localhost:9091/api/v1` | REST | 历史版本、事件 | 94%↗️ |
| **命令API** | `localhost:9090/api/v1` | REST | 创建、更新、删除 | CQRS |

### 性能指标

- **平均性能提升**: 76%
- **缓存命中率**: 91.7%
- **平均响应时间**: 3.7ms (缓存命中)
- **Redis内存使用**: 1.31MB / 512MB

## 🚀 快速开始

### 1. 启动服务

```bash
# 启动所有API服务
cd /home/shangmeilin/cube-castle
./scripts/start-cqrs-complete.sh

# 验证服务状态
curl http://localhost:8090/health  # GraphQL服务
curl http://localhost:9091/health  # 时态API服务  
curl http://localhost:9090/health  # 命令API服务
```

### 2. 访问文档

打开浏览器访问交互式文档中心：
```bash
# 如果在本地运行，直接打开
open docs/api/index.html

# 或通过HTTP服务器
python -m http.server 8000 -d docs/api
# 然后访问 http://localhost:8000
```

### 3. 测试API

```bash
# GraphQL查询示例
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizations(first: 5) { code name unit_type status } }"}'

# 时态API查询示例  
curl "http://localhost:9091/api/v1/organization-units/1000001/temporal?as_of_date=2025-08-09"

# 健康检查
curl http://localhost:8090/health && echo ""
curl http://localhost:9091/health && echo ""
curl http://localhost:9090/health && echo ""
```

## 📖 详细文档

### GraphQL API

- **文档**: [graphql-api.md](./graphql-api.md)
- **交互界面**: http://localhost:8090/graphiql
- **特点**: 灵活查询、字段选择、实时缓存
- **性能**: 65%响应时间改善

### 时态管理API

- **规范文档**: [temporal-api.yaml](./temporal-api.yaml) (OpenAPI 3.0)
- **特点**: 时间点查询、版本历史、事件驱动
- **性能**: 94%响应时间改善
- **查询类型**:
  - 时间点查询 (`as_of_date`)
  - 时间范围查询 (`effective_from`, `effective_to`)
  - 版本查询 (`version`, `include_history`)
  - 事件创建 (`POST /events`)

### 缓存策略

- **指南**: [cache-strategy-guide.md](./cache-strategy-guide.md)
- **缓存层**: Redis (512MB内存限制)
- **策略**: 智能键生成 + 分层TTL + 精确失效
- **监控**: Prometheus指标 + 91.7%命中率

### 集成示例

- **完整指南**: [integration-examples.md](./integration-examples.md)  
- **支持语言**: JavaScript/TypeScript, Python, Go
- **客户端**: Apollo Client, requests, machinebox/graphql
- **特性**: 连接池、重试机制、错误处理

## 🛠️ 开发工具

### 交互式工具

- **GraphiQL**: http://localhost:8090/graphiql - GraphQL查询界面
- **Swagger UI**: 内置在[文档中心](./index.html) - 时态API测试界面  
- **API文档中心**: [index.html](./index.html) - 统一文档入口

### 监控工具

```bash
# Prometheus指标
curl http://localhost:8090/metrics  # GraphQL服务指标
curl http://localhost:9091/metrics  # 时态API指标

# Redis缓存统计
redis-cli info | grep keyspace_
redis-cli --scan --pattern "cache:*" | wc -l
```

## 🔧 配置说明

### 环境变量

```bash
# API认证
export CUBE_CASTLE_API_KEY="your_api_key"
export CUBE_CASTLE_TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"

# 服务端点
export CUBE_CASTLE_GRAPHQL_ENDPOINT="http://localhost:8090/graphql"  
export CUBE_CASTLE_TEMPORAL_URL="http://localhost:9091"
export CUBE_CASTLE_COMMAND_URL="http://localhost:9090"

# 缓存配置
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export CACHE_DEFAULT_TTL="300s"
```

### Docker配置

```yaml
# docker-compose.yml 片段
services:
  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
    ports:
      - "6379:6379"
      
  graphql-service:
    build: ./cmd/organization-query-service-unified
    ports:
      - "8090:8090"
    environment:
      - REDIS_ADDR=redis:6379
      
  temporal-api:
    build: ./cmd/organization-temporal-command-service  
    ports:
      - "9091:9091"
    environment:
      - REDIS_ADDR=redis:6379
```

## 📊 性能优化

### 缓存优化建议

1. **查询优化**
   - GraphQL: 只查询需要的字段
   - 分页: 使用合理的 `first` 和 `offset` 参数
   - 搜索: 使用具体的搜索词而非宽泛匹配

2. **缓存策略**
   - 频繁查询: 2-5分钟TTL
   - 中等频率: 15分钟TTL  
   - 统计数据: 1小时TTL

3. **监控告警**
   - 缓存命中率 < 85% 告警
   - Redis内存使用 > 80% 告警
   - API响应时间 > 100ms 告警

### 客户端优化

```javascript
// Apollo Client缓存配置
const client = new ApolloClient({
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          organizations: {
            keyArgs: ["searchText"], // 缓存键参数
            merge: (existing = [], incoming = []) => {
              return [...existing, ...incoming]; // 分页合并策略
            }
          }
        }
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-first', // 优先缓存
    },
  },
});
```

## 🚨 故障排查

### 常见问题

| 问题 | 症状 | 解决方案 |
|------|------|----------|
| **服务不可用** | `Connection refused` | 检查服务是否启动 `curl localhost:8090/health` |
| **缓存命中率低** | 响应时间慢 | 检查查询参数一致性，优化缓存键策略 |
| **GraphQL错误** | 查询失败 | 检查Schema语法，使用GraphiQL调试 |
| **时态查询无结果** | 返回空数组 | 检查日期格式和时间范围参数 |

### 调试工具

```bash
# 检查服务日志
docker-compose logs graphql-service
docker-compose logs temporal-api

# 检查缓存状态  
redis-cli info memory
redis-cli keys "cache:*" | head -5

# 测试API连通性
curl -f http://localhost:8090/health || echo "GraphQL服务异常"
curl -f http://localhost:9091/health || echo "时态API服务异常"
```

## 🚨 已知特例和注意事项 ⭐ **新增 (2025-08-24)**

### OAuth认证字段名特例

⚠️ **重要**: 前端OAuth认证实现使用了非标准字段名，这是一个已知的技术债务。

#### 问题描述
- **标准OAuth 2.0字段名**: `client_id`, `client_secret` (snake_case)
- **项目实际使用**: `clientId`, `clientSecret` (camelCase) 
- **修复位置**: `/home/shangmeilin/cube-castle/frontend/src/shared/api/auth.ts:66-68`

#### 影响和症状
- **错误症状**: "Failed to fetch organizations. Please try again."
- **根本原因**: OAuth服务器拒绝非标准字段名的token请求
- **影响范围**: 所有前端API调用因认证失败而无法执行

#### 解决方案
```typescript
// ❌ 错误的实现 (曾经的问题代码)
body: JSON.stringify({
  grant_type: this.config.grantType,
  clientId: this.config.clientId,      // 非标准字段名
  clientSecret: this.config.clientSecret, // 非标准字段名
}),

// ✅ 正确的实现 (已修复)
body: JSON.stringify({
  grant_type: this.config.grantType,
  client_id: this.config.clientId,     // 标准OAuth 2.0字段名
  client_secret: this.config.clientSecret, // 标准OAuth 2.0字段名
}),
```

#### 防范措施
1. **开发规范**: OAuth实现必须严格遵循RFC 6749标准字段名
2. **测试要求**: API集成测试必须包含OAuth认证流程测试
3. **文档标注**: 此类协议标准例外必须在API文档中明确标注

### GraphQL Schema字段映射特例

#### 已修复的字段映射问题
- **OrganizationStats**: `total` → `totalCount`, `temporal` → `temporalStats`
- **TypeCount**: `type` → `unitType`  
- **TemporalStats**: 完全重新设计字段结构

#### 预防措施
- 开发前必须使用GraphQL Introspection查询确认Schema
- CI/CD管道集成Schema一致性验证
- 前端TypeScript类型与后端Schema自动同步检查

## 📞 支持与贡献

### 获取帮助

- **问题反馈**: 请在GitHub Issues中提交
- **功能请求**: 请详细描述使用场景和期望功能
- **文档改进**: 欢迎提交PR改进文档

### 贡献指南

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/new-api`)
3. 提交变更 (`git commit -m 'Add new API endpoint'`)
4. 推送分支 (`git push origin feature/new-api`)
5. 创建 Pull Request

### 联系方式

- **项目地址**: `/home/shangmeilin/cube-castle`
- **文档路径**: `/home/shangmeilin/cube-castle/docs/api/`
- **最后更新**: 2025-08-24

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../../LICENSE) 文件了解详情。

---

*🏰 Cube Castle API - 构建企业级组织架构管理系统*