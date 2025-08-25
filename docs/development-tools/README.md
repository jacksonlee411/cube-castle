# Cube Castle API测试工具集

## 概述

本目录包含Cube Castle项目的完整API测试工具集，帮助开发者快速进行API测试、调试和集成验证。

## 🛠️ 工具集内容

### 1. Postman集合 (`postman-collection.json`)
- **功能**: 完整的Postman API测试集合
- **特性**: 
  - 自动JWT令牌管理和刷新
  - 环境变量支持
  - 测试脚本自动化
  - 完整的CRUD操作覆盖
- **导入方法**: Postman → Import → 选择文件

### 2. Insomnia工作空间 (`insomnia-workspace.json`)  
- **功能**: Insomnia REST客户端配置
- **特性**:
  - 预配置环境变量
  - JWT认证集成
  - GraphQL和REST API支持
- **导入方法**: Insomnia → Import/Export → Import Data

### 3. cURL示例脚本 (`curl-examples.md`)
- **功能**: 命令行API测试示例
- **特性**:
  - 完整的bash脚本示例
  - 错误处理和性能测试
  - 自动化测试流程
- **使用方法**: 复制命令或运行完整脚本

## 🚀 快速开始

### 前置条件
1. 后端服务运行:
   - REST命令服务: http://localhost:9090
   - GraphQL查询服务: http://localhost:8090
2. 开发模式启用 (`DEV_MODE=true`)

### 第一次使用步骤

#### 1. 生成JWT令牌 (必须首先执行)
```bash
# 使用cURL生成令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "dev-user",
    "tenantId": "dev-tenant", 
    "roles": ["ADMIN", "USER"],
    "duration": "8h"
  }'
```

#### 2. 配置API客户端
- **Postman**: 导入集合后，运行"生成开发JWT令牌"请求
- **Insomnia**: 导入工作空间后，执行"生成JWT令牌"请求  
- **cURL**: 使用提供的脚本自动管理令牌

#### 3. 验证服务状态
```bash
# 检查服务健康状态
curl http://localhost:9090/health
curl http://localhost:8090/health
```

## 📋 API端点覆盖

### 开发工具端点
- `POST /auth/dev-token` - 生成JWT令牌
- `GET /auth/dev-token/info` - 获取令牌信息
- `GET /dev/status` - 开发环境状态
- `GET /dev/test-endpoints` - 测试端点列表

### REST API (命令操作)
- `POST /api/v1/organization-units` - 创建组织单元
- `PUT /api/v1/organization-units/{code}` - 更新组织单元
- `DELETE /api/v1/organization-units/{code}` - 删除组织单元
- `POST /api/v1/organization-units/{code}/suspend` - 停用组织
- `POST /api/v1/organization-units/{code}/activate` - 激活组织

### GraphQL API (查询操作)
- `organizationStats` - 组织统计信息
- `organizations` - 分页组织列表查询
- `organization(code)` - 单个组织详细查询

## 🔧 高级功能

### 自动JWT令牌管理
所有工具集都支持自动JWT令牌管理:
- **Postman**: Pre-request脚本自动检查和刷新令牌
- **Insomnia**: 环境变量自动更新
- **cURL**: Shell脚本令牌过期检测和重新生成

### 环境变量配置
```json
{
  "command_service_url": "http://localhost:9090",
  "query_service_url": "http://localhost:8090", 
  "tenant_id": "dev-tenant",
  "jwt_token": "",
  "jwt_expiry": ""
}
```

### 测试自动化
- **单元测试**: 每个API端点的响应验证
- **集成测试**: 完整CRUD流程测试
- **性能测试**: 响应时间和并发测试

## 📊 测试场景

### 基础功能测试
1. **服务健康检查**: 验证服务运行状态
2. **JWT认证流程**: 令牌生成、验证、刷新
3. **基础CRUD操作**: 创建、读取、更新、删除组织单元

### 高级场景测试  
1. **分层组织结构**: 父子组织关系管理
2. **状态管理**: 组织激活/停用流程
3. **时态数据**: 历史版本和有效期管理
4. **权限验证**: 不同用户角色的访问控制

### 错误处理测试
1. **认证失败**: 无效或过期令牌处理
2. **参数验证**: 无效输入数据处理
3. **业务规则**: 违反业务逻辑的操作处理
4. **系统错误**: 数据库连接失败等异常情况

## 🛡️ 安全注意事项

### JWT令牌安全
- 令牌仅在开发环境有效 (`DEV_MODE=true`)
- 生产环境不可使用开发工具端点
- 令牌有效期建议设置为8小时以内

### API访问控制
- 所有命令操作需要JWT认证
- 使用正确的租户ID (`X-Tenant-ID`)
- 遵循最小权限原则

## 🔍 故障排除

### 常见问题

#### 1. JWT令牌获取失败
```bash
# 检查开发模式是否启用
curl http://localhost:9090/dev/status | jq '.data.devMode'

# 检查服务运行状态
curl http://localhost:9090/health
```

#### 2. API调用返回401错误
```bash
# 检查令牌是否有效
curl -X GET "http://localhost:9090/auth/dev-token/info" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 重新生成令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -H "Content-Type: application/json" \
  -d '{"userId":"dev-user","tenantId":"dev-tenant","roles":["ADMIN"],"duration":"8h"}'
```

#### 3. GraphQL查询失败
```bash
# 检查查询服务状态
curl http://localhost:8090/health

# 验证GraphQL Schema
curl -X POST "http://localhost:8090/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { __schema { types { name } } }"}'
```

### 性能调试
```bash
# 响应时间测试
time curl -s -X POST "http://localhost:8090/graphql" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"query": "query { organizationStats { totalCount } }"}'

# 并发测试 (需要安装ab工具)
ab -n 100 -c 10 -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8090/health"
```

## 📖 相关文档

- [API规范文档](../development-plans/01-organization-units-api-specification.md)
- [后端实施计划](../development-plans/04-backend-implementation-plan-phases1-3.md)
- [项目开发指南](../../CLAUDE.md)

## 🤝 贡献指南

### 添加新的测试用例
1. 在对应的工具集文件中添加请求定义
2. 确保包含适当的认证和错误处理
3. 在README中更新相应的文档说明
4. 测试新添加的用例确保可用性

### 报告问题
如果发现工具集中的问题或有改进建议，请:
1. 检查是否为已知问题
2. 提供详细的重现步骤
3. 包含相关的错误信息或日志
4. 建议可能的解决方案

## 📝 版本历史

- **v1.0.0** (2025-08-25): 初始版本
  - 完整的Postman/Insomnia/cURL工具集
  - 自动JWT令牌管理
  - 全API端点覆盖
  - 详细的使用文档和故障排除指南

---

*本文档随API工具集的更新而持续维护*