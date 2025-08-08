# Claude Code项目记忆文档

## 项目概述
Cube Castle是一个基于CQRS架构的人力资源管理系统，包含前端React应用和Go后端API服务。

## 当前架构状态

### 前端架构
- **技术栈**: React + TypeScript + Vite
- **状态管理**: React Context
- **UI框架**: Canvas Kit
- **数据获取**: GraphQL (查询) + REST (命令)
- **类型安全**: Zod运行时验证 + TypeScript静态检查

### 后端架构
- **技术栈**: Go + GraphQL + PostgreSQL
- **架构模式**: CQRS (命令查询职责分离)
- **命令端**: REST API (端口9090)
- **查询端**: GraphQL API (端口8080)
- **数据库**: PostgreSQL (端口5432)

## 开发环境配置

### 启动命令
```bash
# 启动后端服务
cd /home/shangmeilin/cube-castle
./start_smart.sh

# 启动前端开发服务器
cd /home/shangmeilin/cube-castle/frontend
npm run dev
```

### 服务端口
- 前端开发服务器: http://localhost:3000
- 后端命令API: http://localhost:9090
- 后端查询API: http://localhost:8080
- PostgreSQL数据库: localhost:5432

### 测试命令
```bash
# 前端测试
cd frontend && npm test

# 后端测试
cd cmd/organization-command-server && go test ./...

# API端到端测试
./test_api.sh
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

### 文件结构重要路径
```
cube-castle/
├── frontend/src/shared/
│   ├── validation/schemas.ts        # Zod验证模式
│   ├── api/type-guards.ts          # 类型守卫函数
│   ├── api/organizations.ts        # API客户端 (已重构)
│   └── api/error-handling.ts       # 错误处理系统
├── cmd/organization-command-server/
│   ├── pkg/types/organization.go   # Go类型定义
│   └── internal/presentation/http/middleware/validation.go # 验证中间件
└── DOCS2/implementation-guides/organization-api-cqrs-enhancement2/
    ├── 01-code-smell-analysis-report.md
    ├── 02-refactor-implementation-plan.md  
    ├── 03-system-simplification-plan.md
    └── 04-next-steps-recommendations.md
```

## 已知问题与解决方案

### 当前问题
1. **验证Schema匹配**: GraphQL响应格式与前端Zod schema存在小幅差异
   - 状态: 已识别，运行时验证正常工作
   - 解决方案: 调整schema或响应格式匹配

### 解决的问题
1. **前端类型安全**: ✅ 已通过Zod运行时验证解决
2. **后端类型验证**: ✅ 已通过Go强类型枚举解决  
3. **错误处理一致性**: ✅ 已通过统一错误处理系统解决
4. **API数据验证**: ✅ 已集成运行时验证到API层

## 开发建议

### 代码规范
- 前端: 优先使用Zod验证而非类型断言
- 后端: 使用强类型枚举而非字符串常量
- 错误处理: 使用统一的ValidationError类
- 测试: 为所有验证逻辑编写单元测试

### 调试技巧
1. **前端验证错误**: 检查浏览器控制台的ValidationError详情
2. **后端验证失败**: 查看Go服务日志中的验证错误信息
3. **数据库连接**: 使用`psql -h localhost -U user -d cubecastle`测试连接

### 性能监控
- 前端: React DevTools检查组件渲染
- 后端: Go pprof分析API性能
- 数据库: PostgreSQL慢查询日志

## 下一步发展方向

### 立即优先 (推荐)
1. **修复验证匹配**: 调整GraphQL响应与Zod schema一致性
2. **技术债务清理**: 删除冗余文件，优化项目结构

### 中期目标
1. **Phase 4 监控**: 实施OpenTelemetry和Prometheus监控
2. **功能完善**: 编辑、删除操作的类型安全改进

### 长期规划  
1. **Phase 5 测试**: E2E测试和性能测试
2. **新功能**: 权限管理、批量操作、可视化组织架构

## 联系与维护
- 项目路径: `/home/shangmeilin/cube-castle`
- 文档路径: `/home/shangmeilin/cube-castle/DOCS2/`
- 最后更新: 2025-08-08
- 当前版本: Phase 3 完成

---
*这个文档会随着项目发展持续更新*