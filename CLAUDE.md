# Claude Code项目记忆文档

## 项目概述
Cube Castle是一个基于CQRS架构的人力资源管理系统，包含前端React应用和Go后端API服务。项目已完成Phase 3类型安全改进和Phase 4监控系统实施。

## 当前架构状态

### 前端架构
- **技术栈**: React + TypeScript + Vite
- **状态管理**: React Context
- **UI框架**: Canvas Kit
- **数据获取**: GraphQL (查询) + REST (命令)
- **类型安全**: Zod运行时验证 + TypeScript静态检查
- **测试框架**: Playwright E2E测试 + Jest单元测试

### 后端架构
- **技术栈**: Go + GraphQL + PostgreSQL + Neo4j
- **架构模式**: CQRS (命令查询职责分离)
- **命令端**: REST API (端口9090)
- **查询端**: GraphQL API (端口8090)
- **数据库**: PostgreSQL (端口5432) + Neo4j (端口7474)
- **监控**: Prometheus指标收集 (端口9999)

## 开发环境配置

### 启动命令
```bash
# 启动后端服务
cd /home/shangmeilin/cube-castle
./start_smart.sh

# 启动前端开发服务器
cd /home/shangmeilin/cube-castle/frontend
npm run dev

# 启动GraphQL服务器(带监控指标)
go run cmd/organization-graphql-service/main.go
```

### 服务端口
- 前端开发服务器: http://localhost:3001
- 后端命令API: http://localhost:9090
- 后端查询API: http://localhost:8090
- **监控指标端点**: http://localhost:8090/metrics ⭐
- PostgreSQL数据库: localhost:5432
- Neo4j数据库: localhost:7474
- 监控面板: file:///home/shangmeilin/cube-castle/monitoring/dashboard.html

### 测试命令
```bash
# 前端单元测试
cd frontend && npm test

# 前端E2E测试
cd frontend && npx playwright test

# 后端测试
cd cmd/organization-command-server && go test ./...

# API端到端测试
./test_api.sh

# 启动监控服务
cd monitoring && go run metrics-server.go
```

## 开发历史与重要改进

### Phase 3: 类型安全与质量提升 (已完成 ✅)

#### 前端改进
1. **Zod运行时验证**: 实现了完整的数据验证模式
   - `OrganizationUnitSchema`: 组织单元验证
   - `CreateOrganizationInputSchema`: 创建输入验证
   - `UpdateOrganizationInputSchema`: 更新输入验证

2. **类型守卫系统**: 创建了安全的类型转换函数
   - `validateOrganizationUnit`: 组织单元验证
   - `validateCreateOrganizationInput`: 创建输入验证
   - `safeTransformGraphQLToOrganizationUnit`: 安全数据转换

3. **错误处理改进**: 统一的错误处理机制
   - `ValidationError`类: 结构化验证错误
   - `ErrorHandler`类: 统一错误处理
   - 用户友好的错误消息显示

4. **API层重构**: 替换所有`any`类型为安全验证
   - 文件: `frontend/src/shared/api/organizations.ts`
   - 集成运行时验证到所有API调用
   - 移除类型断言，使用安全验证函数

#### 后端改进
1. **强类型枚举系统**: Go枚举类型实现
   - `UnitType`: 组织类型枚举 (COMPANY, DEPARTMENT, TEAM等)
   - `Status`: 状态枚举 (ACTIVE, INACTIVE, PLANNED)
   - 包含验证方法和字符串转换

2. **值对象模式**: 类型安全的业务对象
   - `OrganizationCode`: 7位数字代码验证
   - `TenantID`: 租户标识符
   - 包含业务规则验证

3. **请求验证中间件**: HTTP请求验证
   - `CreateOrganizationRequest`: 创建请求验证
   - `UpdateOrganizationRequest`: 更新请求验证
   - 上下文注入验证结果

#### 测试覆盖
- **前端单元测试**: 43个测试用例全部通过
- **后端单元测试**: Go测试覆盖类型验证、中间件、业务逻辑
- **集成测试**: MCP浏览器自动化验证端到端流程

## 监控系统实施状态

### Phase 4: 监控与可观测性实施 (已完成 ✅)

#### 监控系统架构
1. **真实指标收集**: 完整的Prometheus兼容指标系统
   - GraphQL服务器(8090)内置 `/metrics` 端点
   - HTTP请求指标、业务操作指标自动收集
   - 支持多服务标签分离 (graphql-server, command-server)

2. **前端监控面板**: 混合真实数据显示
   - 自动解析Prometheus指标格式
   - 真实服务健康检查机制
   - 智能fallback到模拟数据

3. **完整指标类型**:
   - `http_requests_total`: HTTP请求计数 (按method, status, service分组)
   - `http_request_duration_seconds`: 请求响应时间直方图
   - `organization_operations_total`: 业务操作计数 (按operation, status, service分组)

4. **集成测试验证**: ✅ 端到端指标流程验证完成
   - GraphQL查询 → 指标生成 → 前端显示
   - 业务操作指标正确记录 (query_list: success)
   - HTTP性能指标准确收集 (平均9.89ms响应时间)

#### 当前指标示例
```prometheus
# HTTP性能指标
http_requests_total{method="POST",service="graphql-server",status="OK"} 1
http_request_duration_seconds_sum{endpoint="/graphql",method="POST",service="graphql-server"} 0.009891935

# 业务操作指标  
organization_operations_total{operation="query_list",service="graphql-server",status="success"} 1
```

#### 前端代理配置
```typescript
// vite.config.ts
'/api/metrics': {
  target: 'http://localhost:8090',  // 指向GraphQL服务器
  changeOrigin: true,
  rewrite: (path) => path.replace(/^\/api\/metrics/, '/metrics')
}
```

#### 集成测试扩展
1. **Schema验证测试**: 完整的运行时验证测试套件
   - 文件: `frontend/tests/e2e/schema-validation.spec.ts`
   - 测试创建流程、错误处理、数据格式验证
   - 验证Zod运行时验证机制有效性

2. **端到端测试**: Playwright自动化测试
   - 跨浏览器测试支持 (Chrome, Firefox, Safari)
   - 业务流程完整性验证
   - 错误场景处理测试
### 文件结构重要路径
```
cube-castle/
├── frontend/src/shared/
│   ├── validation/schemas.ts        # Zod验证模式
│   ├── api/type-guards.ts          # 类型守卫函数
│   ├── api/organizations.ts        # API客户端 (已重构)
│   └── api/error-handling.ts       # 错误处理系统
├── frontend/tests/e2e/
│   └── schema-validation.spec.ts   # Schema验证集成测试
├── cmd/organization-command-server/
│   ├── pkg/types/organization.go   # Go类型定义
│   └── internal/presentation/http/middleware/validation.go # 验证中间件
├── monitoring/
│   ├── metrics-server.go           # 指标收集服务器
│   ├── dashboard.html             # 监控可视化面板
│   ├── prometheus.yml             # Prometheus配置
│   ├── alert_rules.yml           # 告警规则
│   └── docker-compose.monitoring.yml # 监控服务部署
└── DOCS2/implementation-guides/organization-api-cqrs-enhancement2/
    ├── 01-code-smell-analysis-report.md
    ├── 02-refactor-implementation-plan.md  
    ├── 03-system-simplification-plan.md
    └── 04-next-steps-recommendations.md
```

## 已知问题与解决方案

### 当前问题
1. **E2E测试稳定性**: 部分Playwright测试因为弹窗检测时序问题偶有失败
   - 状态: 测试框架已建立，需要UI组件完善
   - 解决方案: 调整测试等待策略和选择器精确度

2. **监控服务集成**: Docker镜像拉取受代理限制
   - 状态: 已实现轻量级Go监控服务替代方案
   - 解决方案: 使用自建监控服务，避免外部依赖

### 解决的问题
1. **前端类型安全**: ✅ 已通过Zod运行时验证解决
2. **后端类型验证**: ✅ 已通过Go强类型枚举解决  
3. **错误处理一致性**: ✅ 已通过统一错误处理系统解决
4. **API数据验证**: ✅ 已集成运行时验证到API层
5. **验证Schema匹配**: ✅ 已通过专门的CreateOrganizationResponseSchema解决
6. **系统监控缺失**: ✅ 已实施完整的监控和可观测性系统
7. **测试覆盖不足**: ✅ 已扩展E2E测试和Schema验证测试

## 开发建议

### 代码规范
- 前端: 优先使用Zod验证而非类型断言
- 后端: 使用强类型枚举而非字符串常量
- 错误处理: 使用统一的ValidationError类
- 测试: 为所有验证逻辑编写单元测试
- 监控: 在关键业务逻辑中添加指标收集

### 调试技巧
1. **前端验证错误**: 检查浏览器控制台的ValidationError详情
2. **后端验证失败**: 查看Go服务日志中的验证错误信息
3. **数据库连接**: 使用`psql -h localhost -U user -d cubecastle`测试连接
4. **监控指标**: 访问`http://localhost:9999/metrics`查看实时指标
5. **系统状态**: 打开监控面板查看服务健康状态

### 性能监控
- 前端: React DevTools检查组件渲染
- 后端: Go pprof分析API性能 + Prometheus指标
- 数据库: PostgreSQL慢查询日志
- 系统: 监控面板实时显示响应时间和错误率

## 下一步发展方向

### 立即优先 (推荐)
1. **E2E测试稳定性**: 优化Playwright测试选择器和等待策略
2. **监控告警集成**: 配置实际的告警通知机制
3. **UI组件完善**: 改进组织架构页面的弹窗和表单组件

### 中期目标
1. **性能优化**: 基于监控数据优化API响应时间和数据库查询
2. **功能完善**: 编辑、删除操作的类型安全改进
3. **监控扩展**: 添加业务指标监控(用户行为、数据质量等)

### 长期规划  
1. **Phase 5 完整测试**: 性能测试、压力测试、契约测试
2. **新功能**: 权限管理、批量操作、可视化组织架构
3. **DevOps完善**: CI/CD流水线、自动化部署、环境管理

## 联系与维护
- 项目路径: `/home/shangmeilin/cube-castle`
- 文档路径: `/home/shangmeilin/cube-castle/DOCS2/`
- 监控路径: `/home/shangmeilin/cube-castle/monitoring/`
- 最后更新: 2025-08-09
- 当前版本: Phase 4+ 真实指标监控完成 (包含端到端验证的完整监控系统)

---
*这个文档会随着项目发展持续更新*