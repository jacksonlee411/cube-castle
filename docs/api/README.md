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

| 服务 | 端点 | 协议 | 用途 |
|------|------|------|------|
| **GraphQL查询** | `localhost:8090/graphql` | GraphQL | 灵活查询、统计 |
| **命令API** | `localhost:9090/api/v1` | REST | 创建、更新、删除 |

> ⚠️ `localhost` 端点说明：所有接口均由 `docker-compose.dev.yml` 启动的容器服务暴露。禁止在宿主机安装同名服务占用端口；如遇冲突，请卸载宿主服务后重新执行 `make run-dev`。

### 性能指标

- **平均性能提升**: 76%
- **缓存命中率**: 91.7%
- **平均响应时间**: 3.7ms (缓存命中)
- **Redis内存使用**: 1.31MB / 512MB

## 🚀 快速开始

### 1. 启动服务

```bash
# 启动基础设施与核心服务（PostgreSQL 原生）
make docker-up
make run-dev

# 验证服务状态
curl http://localhost:8090/health  # GraphQL服务
curl http://localhost:9090/health  # 命令API服务
```

### 2. 查看契约

- REST 契约文件：`docs/api/openapi.yaml`
- GraphQL Schema：`docs/api/schema.graphql`

### 3. 测试API（含必需头部）

```bash
# GraphQL 查询示例（遵循最新契约，使用分页与包装结构）
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "query": "query($page:Int,$pageSize:Int){ organizations(pagination:{page:$page,pageSize:$pageSize}) { data { code name unitType status } pagination { total page pageSize hasNext } } }",
    "variables": {"page":1, "pageSize":10}
  }'

# 健康检查（REST/GraphQL）
curl -H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: $TENANT_ID" http://localhost:8090/health && echo ""
curl -H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: $TENANT_ID" http://localhost:9090/health && echo ""
```

必需头部（所有 API 请求）：
- `Authorization: Bearer <JWT_TOKEN>`
- `X-Tenant-ID: <TENANT_ID>`

## 📖 详细文档

### GraphQL API

- **交互界面**: http://localhost:8090/graphiql
- **契约文件**: `docs/api/schema.graphql`

## 🛠️ 开发工具

### 交互式工具

- **GraphiQL**: http://localhost:8090/graphiql - GraphQL 查询界面

## 🔧 配置说明

### 环境变量

```bash
# API认证
export CUBE_CASTLE_API_KEY="your_api_key"
export CUBE_CASTLE_TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"

# 服务端点
export CUBE_CASTLE_GRAPHQL_ENDPOINT="http://localhost:8090/graphql"  
export CUBE_CASTLE_COMMAND_URL="http://localhost:9090"

# 缓存配置
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export CACHE_DEFAULT_TTL="300s"
```

### Docker配置

Docker 配置以仓库根目录的 `docker-compose.yml` 为准；如需调整请先更新契约并通过契约测试。

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

# 检查缓存状态  
redis-cli info memory
redis-cli keys "cache:*" | head -5

# 测试API连通性
curl -f http://localhost:8090/health || echo "GraphQL服务异常"
```

## 🚨 已知特例和注意事项 ⭐ **更新 (2025-09-07)**

### parentCode字段必填要求 ⭐ **重要变更 (2025-09-07)**

#### 变更说明
从本版本开始，**所有组织的上级组织编码（parentCode）字段变更为必填字段**。

#### 字段规范
- **根组织**: `parentCode = "0"` (字符串"0"，表示无上级组织)
- **子组织**: `parentCode = "1000xxx"` (7位数字组织编码，表示上级组织)
- **字段类型**: 从 `String`(可选) 变更为 `String!`(必填)

#### 影响范围
- **OpenAPI规范**: 所有Schema中的parentCode字段标记为必填
- **GraphQL Schema**: Organization类型中的parentCode字段标记为必填
- **数据库**: 现有数据已完成迁移，1000000组织的parentCode设置为"0"

#### API调用变更
```json
// ✅ 新的API请求格式 - parentCode必须提供
{
  "name": "新部门",
  "unitType": "DEPARTMENT", 
  "parentCode": "1000000",     // 必填字段
  "effectiveDate": "2025-09-07",
  "operationReason": "业务扩展" // 可选字段，省略时服务器记录为空字符串
}

// ❌ 旧的API请求格式 - parentCode可选，现在将报错
{
  "name": "新部门",
  "unitType": "DEPARTMENT", 
  // parentCode: null,         // 现在不允许为空
  "effectiveDate": "2025-09-07"
}
```

> **提示**：自 v4.5 起 `operationReason` 字段改为可选，省略时审计记录会保存为空字符串。

#### 迁移指南
1. **前端应用**: 确保所有组织创建/更新表单包含parentCode字段选择
2. **API客户端**: 更新API调用，为所有组织操作提供有效的parentCode值
3. **数据导入**: 批量数据导入时必须为每个组织指定parentCode
4. **测试用例**: 更新所有测试用例，确保包含parentCode字段验证

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
