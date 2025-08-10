# Cube Castle API 文档中心

## 📚 文档概览

欢迎使用 Cube Castle API 文档中心！这里包含了完整的API文档、集成示例和最佳实践指南。

### 🚀 快速访问

- **🏠 [API文档中心](./index.html)** - 交互式API文档界面
- **🔍 [GraphQL API文档](./graphql-api.md)** - GraphQL查询服务完整文档  
- **⏰ [时态API规范](./temporal-api.yaml)** - OpenAPI/Swagger规范文档
- **⚡ [缓存策略指南](./cache-strategy-guide.md)** - Redis缓存使用指南
- **🛠️ [集成示例](./integration-examples.md)** - 多语言客户端实现

## 📋 文档结构

```
docs/api/
├── index.html                    # 🏠 交互式文档中心
├── README.md                     # 📚 本文档
├── temporal-api.yaml             # ⏰ 时态API OpenAPI规范
├── graphql-api.md               # 🔍 GraphQL API文档  
├── cache-strategy-guide.md      # ⚡ 缓存策略指南
├── integration-examples.md      # 🛠️ 集成示例
└── examples/                    # 📁 代码示例目录
    ├── javascript/              # JavaScript/TypeScript示例
    ├── python/                  # Python客户端示例
    └── go/                      # Go客户端示例
```

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
- **最后更新**: 2025-08-10

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../../LICENSE) 文件了解详情。

---

*🏰 Cube Castle API - 构建企业级组织架构管理系统*