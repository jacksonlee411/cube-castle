# Cube Castle API测试工具集

> 说明：本目录仅提供工具与用法说明。项目的原则、约束与权威链接以仓库根目录 `AGENTS.md` 为唯一事实来源；如本文件与 `AGENTS.md` 或 `docs/reference/*` 存在不一致，以 `AGENTS.md` 为准。

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
1. 后端服务运行（均由 Docker Compose 管理暴露端口；如端口冲突，须卸载宿主机同名服务，禁止修改容器端口映射）:
   - REST命令服务: http://localhost:9090
   - GraphQL查询服务: http://localhost:8090
2. 已执行一次性密钥与令牌准备：`make jwt-dev-setup`（生成 RS256 密钥对）→ `make jwt-dev-mint`（生成开发令牌，保存至 `.cache/dev.jwt`）

### 第一次使用步骤

#### 1. 生成JWT令牌（推荐方式）
```bash
# 生成开发令牌（保存至 .cache/dev.jwt）
make jwt-dev-mint USER_ID=dev TENANT_ID=3b99930c-e2e4-4d4a-8e7a-123456789abc

# 可选：导出当前 Shell 会话变量
eval $(make jwt-dev-export)   # 导出 JWT_TOKEN

# 验证公钥（JWKS）
curl -s http://localhost:9090/.well-known/jwks.json | (command -v jq >/dev/null && jq . || cat)
```

#### 2. 配置API客户端
- **Postman / Insomnia**：将 `.cache/dev.jwt` 的内容设置到环境变量（如 `jwt_token`），并在请求头中添加 `Authorization: Bearer {{jwt_token}}` 与 `X-Tenant-ID`。
- **cURL**：使用 `JWT_TOKEN=$(cat .cache/dev.jwt)` 或 `eval $(make jwt-dev-export)` 自动注入。

#### 3. 验证服务状态
```bash
# 检查服务健康状态
curl http://localhost:9090/health
curl http://localhost:8090/health
```

## 📋 API 依据契约

请以 `docs/api/openapi.yaml`（REST 命令）与 `docs/api/schema.graphql`（GraphQL 查询）为唯一事实来源。下列仅为常见示例，非完整清单：

### 开发工具端点（仅开发环境）
（开发令牌的生成/导出请使用 `make jwt-dev-mint` 与 `make jwt-dev-export`；公钥验证可使用 `/.well-known/jwks.json`，避免依赖未契约的调试端点。）

### REST API (命令操作)
- `POST /api/v1/organization-units` - 创建组织单元
- `PUT /api/v1/organization-units/{code}` - 更新组织单元
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
- 令牌仅用于本地开发与联调（`make run-dev` 环境）；生产环境不可使用开发工具端点
- 生产环境不可使用开发工具端点
- 令牌默认保存在 `.cache/dev.jwt`（本地），建议有效期不超过 8 小时

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
# 重新生成令牌（推荐使用 Make 工具链）
make jwt-dev-mint USER_ID=dev TENANT_ID=3b99930c-e2e4-4d4a-8e7a-123456789abc

# 导出当前会话变量并重试
eval $(make jwt-dev-export)
# 例如：
curl -H "Authorization: Bearer $JWT_TOKEN" -H "X-Tenant-ID: 3b99930c-e2e4-4d4a-8e7a-123456789abc" http://localhost:9090/health
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

- [API 契约（权威）](../api/openapi.yaml) · [GraphQL Schema](../api/schema.graphql)
- [JWT开发工具指南](../development-guides/jwt-development-guide.md)
- [项目原则与索引（唯一事实来源）](../../AGENTS.md)

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
