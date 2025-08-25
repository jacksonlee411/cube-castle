# API开发工作流指南

## 概述

本指南描述了Cube Castle项目的API开发标准工作流程，遵循"API契约优先"原则，确保前后端开发的一致性和质量。

## 🎯 核心原则

### 1. API契约优先 (Contract-First)
- **设计先行**: API设计优先于代码实现
- **规范驱动**: 基于OpenAPI和GraphQL Schema规范开发
- **测试验证**: 契约测试确保实现符合规范

### 2. CQRS架构分离
- **查询操作**: 统一使用GraphQL (http://localhost:8090)
- **命令操作**: 统一使用REST API (http://localhost:9090)
- **协议专用**: 避免混用协议，保持架构清晰

### 3. 企业级标准
- **响应统一**: 统一的企业级响应信封格式
- **字段规范**: camelCase命名标准，一致的数据模型
- **错误处理**: 标准化的错误代码和消息格式

## 🔄 开发工作流

### 阶段1: API设计与规范

#### 1.1 需求分析
```yaml
输入: 业务需求文档、用例描述
输出: API功能需求清单
工具: 需求分析模板、业务流程图

步骤:
  1. 识别业务实体和操作
  2. 确定数据流向（查询vs命令）
  3. 定义权限和安全要求
  4. 制定性能和可用性目标
```

#### 1.2 协议选择决策
```yaml
决策规则:
  查询操作 → GraphQL:
    - 数据查询、过滤、分页
    - 复杂关联查询  
    - 统计和报表
    - 历史数据查询
    
  命令操作 → REST API:
    - 创建、更新、删除操作
    - 状态变更（activate/suspend）
    - 业务流程触发
    - 批量操作

实例决策:
  ✅ organizations查询 → GraphQL
  ✅ 创建组织单元 → REST POST
  ✅ 组织统计 → GraphQL
  ✅ 停用组织 → REST POST
```

#### 1.3 API规范编写

**GraphQL Schema** (`docs/api/schema.graphql`):
```graphql
"""
组织单元查询根类型
"""
type Query {
  """
  分页查询组织单元列表
  """
  organizations(
    filter: OrganizationFilter
    pagination: PaginationInput
  ): OrganizationConnection!
  
  """
  查询单个组织单元详细信息
  """
  organization(code: String!): Organization
  
  """
  获取组织统计信息
  """
  organizationStats: OrganizationStats!
}

"""
组织单元核心类型
"""
type Organization {
  code: String!
  name: String!
  unitType: UnitType!
  status: OrganizationStatus!
  level: Int!
  path: String!
  parentCode: String
  effectiveDate: Date!
  endDate: Date
  isCurrent: Boolean!
  description: String
  createdAt: DateTime!
  updatedAt: DateTime!
}
```

**REST API规范** (`docs/api/openapi.yaml`):
```yaml
paths:
  /api/v1/organization-units:
    post:
      summary: 创建组织单元
      operationId: createOrganizationUnit
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrganizationRequest'
      responses:
        '201':
          description: 组织单元创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizationResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'

components:
  schemas:
    CreateOrganizationRequest:
      type: object
      required:
        - name
        - unitType
      properties:
        name:
          type: string
          description: 组织单元名称
          example: "技术部"
        unitType:
          $ref: '#/components/schemas/UnitType'
        parentCode:
          type: string
          description: 父组织单元代码
          example: "CORP001"
        description:
          type: string
          description: 组织单元描述
```

#### 1.4 规范审查
```yaml
审查清单:
  语法正确性:
    ✓ OpenAPI 3.0.3语法验证
    ✓ GraphQL Schema语法检查
    ✓ 类型定义完整性验证
    
  命名一致性:
    ✓ camelCase字段命名规范
    ✓ 跨协议术语一致性检查
    ✓ 标准词汇表遵循验证
    
  企业级标准:
    ✓ 响应信封格式规范
    ✓ 错误代码标准化
    ✓ 认证权限定义清晰
    
  业务逻辑:
    ✓ 数据模型正确性
    ✓ 业务规则完整性
    ✓ 边界条件处理
```

### 阶段2: 契约测试编写

#### 2.1 测试用例设计
```yaml
测试层级:
  L1 - 语法测试:
    - Schema语法正确性
    - 参数类型匹配
    - 必填字段验证
    
  L2 - 语义测试:  
    - 业务规则验证
    - 数据约束检查
    - 错误场景覆盖
    
  L3 - 集成测试:
    - 端到端流程验证
    - 跨服务数据一致性
    - 性能基准测试
```

#### 2.2 契约测试实现
```typescript
// tests/contract/organization-api.test.ts
describe('Organization API Contract Tests', () => {
  describe('GraphQL Queries', () => {
    it('should validate organizations query schema', async () => {
      const query = `
        query Organizations($filter: OrganizationFilter) {
          organizations(filter: $filter) {
            nodes {
              code
              name
              unitType
              status
            }
            pagination {
              total
              hasNext
            }
          }
        }
      `;
      
      const result = await graphqlRequest(query, {
        filter: { status: 'ACTIVE' }
      });
      
      expect(result.data.organizations).toBeDefined();
      expect(result.data.organizations.nodes).toBeInstanceOf(Array);
      result.data.organizations.nodes.forEach(org => {
        expect(org).toHaveProperty('code');
        expect(org).toHaveProperty('name');
        expect(org.unitType).toMatch(/^(COMPANY|DEPARTMENT|TEAM|POSITION)$/);
      });
    });
  });

  describe('REST API Operations', () => {
    it('should validate create organization request', async () => {
      const createRequest = {
        name: '测试部门',
        unitType: 'DEPARTMENT',
        description: '契约测试用部门',
        effectiveDate: '2025-08-25'
      };
      
      const response = await restRequest('POST', '/api/v1/organization-units', createRequest);
      
      expect(response.status).toBe(201);
      expect(response.body).toHaveProperty('success', true);
      expect(response.body.data).toHaveProperty('code');
      expect(response.body.data.name).toBe(createRequest.name);
    });
  });
});
```

### 阶段3: 并行开发

#### 3.1 后端实现开发

**GraphQL解析器实现**:
```go
// internal/graphql/resolvers/organization.go
func (r *queryResolver) Organizations(ctx context.Context, filter *types.OrganizationFilter, pagination *types.PaginationInput) (*types.OrganizationConnection, error) {
    // 验证权限
    if !auth.HasPermission(ctx, "READ_ORGANIZATIONS") {
        return nil, errors.New("insufficient permissions")
    }
    
    // 应用过滤器和分页
    organizations, total, err := r.orgRepo.Query(ctx, filter, pagination)
    if err != nil {
        return nil, err
    }
    
    // 构建企业级响应
    return &types.OrganizationConnection{
        Nodes: organizations,
        Pagination: &types.PaginationInfo{
            Total:       total,
            HasNext:     pagination.Offset+len(organizations) < total,
            HasPrevious: pagination.Offset > 0,
        },
    }, nil
}
```

**REST处理器实现**:
```go
// internal/handlers/organization.go
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
    var req types.CreateOrganizationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
        return
    }

    // 业务验证
    if err := utils.ValidateCreateOrganization(&req); err != nil {
        h.writeErrorResponse(w, r, http.StatusBadRequest, "VALIDATION_ERROR", "输入验证失败", err)
        return
    }

    // 创建组织
    org, err := h.repo.Create(r.Context(), &req)
    if err != nil {
        h.writeErrorResponse(w, r, http.StatusInternalServerError, "CREATE_ERROR", "创建失败", err)
        return
    }

    // 企业级成功响应
    response := h.toOrganizationResponse(org)
    requestID := middleware.GetRequestID(r.Context())
    successResponse := types.WriteSuccessResponse(response, "Organization created successfully", requestID)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(successResponse)
}
```

#### 3.2 前端Mock开发

**GraphQL Mock服务器**:
```typescript
// src/mocks/graphql-mocks.ts
import { graphql, http, HttpResponse } from 'msw';

export const graphqlMocks = [
  graphql.query('Organizations', ({ variables }) => {
    const { filter, pagination } = variables;
    
    return HttpResponse.json({
      data: {
        organizations: {
          nodes: [
            {
              code: 'DEPT001',
              name: '技术部',
              unitType: 'DEPARTMENT',
              status: 'ACTIVE',
              level: 1,
              effectiveDate: '2025-08-25',
              isCurrent: true
            }
          ],
          pagination: {
            total: 1,
            hasNext: false,
            hasPrevious: false
          }
        }
      }
    });
  }),
  
  http.post('/api/v1/organization-units', () => {
    return HttpResponse.json({
      success: true,
      data: {
        code: 'DEPT002',
        name: '新建部门',
        unitType: 'DEPARTMENT',
        status: 'ACTIVE',
        createdAt: new Date().toISOString()
      },
      message: 'Organization created successfully',
      timestamp: new Date().toISOString(),
      requestId: 'mock-req-123'
    }, { status: 201 });
  })
];
```

**前端服务层实现**:
```typescript
// src/services/organization.service.ts
import { GraphQLClient } from 'graphql-request';
import { OrganizationsQuery, CreateOrganizationMutation } from './generated/graphql';

export class OrganizationService {
  private graphqlClient: GraphQLClient;
  private restClient: AxiosInstance;
  
  async getOrganizations(filter?: OrganizationFilter): Promise<Organization[]> {
    const query = `
      query Organizations($filter: OrganizationFilter) {
        organizations(filter: $filter) {
          nodes {
            code
            name
            unitType
            status
            effectiveDate
            isCurrent
          }
        }
      }
    `;
    
    const result = await this.graphqlClient.request(query, { filter });
    return result.organizations.nodes;
  }
  
  async createOrganization(request: CreateOrganizationRequest): Promise<Organization> {
    const response = await this.restClient.post('/api/v1/organization-units', request);
    
    if (!response.data.success) {
      throw new Error(response.data.error.message);
    }
    
    return response.data.data;
  }
}
```

### 阶段4: 集成测试与验证

#### 4.1 契约一致性测试
```bash
#!/bin/bash
# contract-validation.sh

echo "🧪 运行契约一致性测试"

# 1. Schema语法验证
echo "📋 验证GraphQL Schema语法..."
npx graphql-schema-linter docs/api/schema.graphql

# 2. OpenAPI规范验证  
echo "📋 验证OpenAPI规范..."
npx swagger-codegen validate -i docs/api/openapi.yaml

# 3. 实现一致性测试
echo "🔍 运行实现一致性测试..."
npm run test:contract

# 4. 端到端集成测试
echo "🔗 运行端到端集成测试..."
npm run test:e2e
```

#### 4.2 性能基准测试
```typescript
// tests/performance/api-benchmark.test.ts
describe('API Performance Benchmarks', () => {
  it('should meet GraphQL query performance requirements', async () => {
    const query = `query { organizationStats { totalCount } }`;
    
    const startTime = Date.now();
    const result = await graphqlClient.request(query);
    const responseTime = Date.now() - startTime;
    
    expect(result.organizationStats).toBeDefined();
    expect(responseTime).toBeLessThan(100); // < 100ms
  });
  
  it('should handle concurrent requests efficiently', async () => {
    const promises = Array.from({ length: 50 }, () =>
      restClient.post('/api/v1/organization-units', mockOrganization)
    );
    
    const startTime = Date.now();
    const results = await Promise.all(promises);
    const totalTime = Date.now() - startTime;
    
    expect(results.every(r => r.status === 201)).toBe(true);
    expect(totalTime).toBeLessThan(5000); // 50个请求 < 5s
  });
});
```

### 阶段5: 部署与监控

#### 5.1 部署前检查
```yaml
部署清单:
  代码质量:
    ✓ 所有契约测试通过
    ✓ 单元测试覆盖率 > 80%
    ✓ 集成测试通过
    ✓ 性能基准达标
    
  安全检查:
    ✓ JWT认证正确实现
    ✓ 权限控制验证
    ✓ 输入验证完整
    ✓ 敏感数据保护
    
  文档完整:
    ✓ API文档更新
    ✓ 版本变更记录
    ✓ 部署说明更新
    ✓ 监控指标定义
```

#### 5.2 监控配置
```yaml
# monitoring/api-alerts.yml
groups:
  - name: cube-castle-api
    rules:
      - alert: HighAPILatency
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API响应时间过高"
          description: "95%分位响应时间超过500ms"
          
      - alert: APIErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
        for: 2m  
        labels:
          severity: critical
        annotations:
          summary: "API错误率过高"
          description: "5xx错误率超过1%"
```

## 🛠️ 开发工具支持

### IDE配置

**VSCode设置** (`.vscode/settings.json`):
```json
{
  "graphql.validate": true,
  "graphql.schema": "docs/api/schema.graphql",
  "openapi.validate": true,
  "openapi.spec": "docs/api/openapi.yaml",
  "typescript.preferences.includePackageJsonAutoImports": "on"
}
```

**推荐扩展**:
- GraphQL Language Support
- OpenAPI (Swagger) Editor  
- Thunder Client (API测试)
- GitLens (版本控制)

### 自动化工具

**Git钩子** (`.git/hooks/pre-commit`):
```bash
#!/bin/bash
# Pre-commit契约验证

echo "🔍 Pre-commit契约验证..."

# 验证API规范语法
if ! npx swagger-codegen validate -i docs/api/openapi.yaml; then
  echo "❌ OpenAPI规范验证失败"
  exit 1
fi

if ! npx graphql-schema-linter docs/api/schema.graphql; then
  echo "❌ GraphQL Schema验证失败" 
  exit 1
fi

# 运行契约测试
if ! npm run test:contract; then
  echo "❌ 契约测试失败"
  exit 1
fi

echo "✅ Pre-commit验证通过"
```

**CI/CD流水线** (`.github/workflows/api-validation.yml`):
```yaml
name: API Contract Validation
on:
  pull_request:
    paths:
      - 'docs/api/**'
      - 'src/**'
      - 'cmd/**'

jobs:
  contract-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'
          
      - name: Validate API Specifications
        run: |
          npm install
          npx swagger-codegen validate -i docs/api/openapi.yaml
          npx graphql-schema-linter docs/api/schema.graphql
          
      - name: Run Contract Tests
        run: |
          npm run test:contract
          go test -v ./tests/contract/...
          
      - name: Performance Benchmarks
        run: npm run test:performance
```

## 📊 质量指标

### API质量标准
```yaml
契约测试覆盖率: 100%
  - 所有API端点必须有契约测试
  - 正常和异常场景全覆盖
  - 数据模型完整验证

性能要求:
  - GraphQL查询: < 100ms (95%分位)
  - REST API: < 200ms (95%分位)
  - 并发处理: 1000 RPS

可用性目标:
  - 系统可用率: 99.9%
  - 错误率: < 0.1%
  - 恢复时间: < 5分钟
```

### 代码质量要求
```yaml
测试覆盖率:
  - 单元测试: > 80%
  - 集成测试: > 70%
  - E2E测试: 核心流程100%

代码规范:
  - ESLint/GoLint零错误
  - 类型安全100% (TypeScript)
  - API文档完整性100%

安全标准:
  - JWT认证强制执行
  - 输入验证100%覆盖
  - 敏感数据加密存储
```

## 🚨 常见问题解决

### 1. 契约测试失败
```bash
# 问题: Schema不匹配
# 解决: 更新GraphQL Schema或修正实现

# 检查Schema一致性
npx graphql-codegen --check

# 重新生成类型定义
npm run codegen:graphql
```

### 2. 性能不达标
```bash
# 问题: API响应时间过长
# 解决: 性能分析和优化

# 启用性能分析
go tool pprof http://localhost:9090/debug/pprof/profile

# 数据库查询优化
EXPLAIN ANALYZE SELECT * FROM organizations WHERE status = 'ACTIVE';
```

### 3. 认证问题
```bash
# 问题: JWT认证失败
# 解决: 验证令牌配置

# 生成调试令牌
curl -X POST "http://localhost:9090/auth/dev-token" \
  -d '{"userId":"debug-user","roles":["ADMIN"],"duration":"1h"}'

# 验证令牌状态
curl -X GET "http://localhost:9090/auth/dev-token/info" \
  -H "Authorization: Bearer $TOKEN"
```

## 📚 参考资源

### 标准文档
- [OpenAPI 3.0规范](https://swagger.io/specification/)
- [GraphQL规范](https://spec.graphql.org/)
- [REST API设计指南](https://restfulapi.net/)

### 项目文档
- [API规范文档](../development-plans/01-organization-units-api-specification.md)
- [契约测试指南](../development-plans/07-contract-testing-automation-system.md)
- [JWT开发工具](./jwt-development-guide.md)

### 工具链
- [Postman API测试](https://www.postman.com/)
- [Insomnia REST客户端](https://insomnia.rest/)
- [GraphQL Playground](https://github.com/graphql/graphql-playground)

---

*本工作流指南随项目发展持续更新*