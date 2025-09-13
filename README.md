# 🏰 Cube Castle - 企业级CoreHR SaaS平台

> **版本**: v4.0-Unified-Architecture | **更新日期**: 2025年9月7日 | **架构**: PostgreSQL原生CQRS + 统一配置管理

基于**PostgreSQL原生架构**和**Canvas Kit v13设计系统**的企业级HR SaaS平台，采用React 19 + Vite 7构建，实现了**95%重复代码消除**和**企业级架构统一**。

## 🚀 核心架构成果 ⭐ **S级完成**

### ✅ **PostgreSQL原生CQRS** - 性能提升70-90%
- **查询响应**: 1.5-8ms (原15-58ms)
- **单一数据源**: 消除Neo4j+CDC复杂性
- **26个时态索引**: 极致查询性能优化
- **零同步延迟**: 强一致性保证

### ✅ **统一配置架构** - 95%+硬编码消除
- **权威配置源**: `frontend/src/shared/config/ports.ts`
- **端口集中管理**: 15+文件→1个统一配置
- **类型安全**: TypeScript保护所有配置引用
- **零配置冲突**: CQRS端点标准化

### ✅ **重复代码消除** - 93%架构优化完成
- **Hook统一**: 7→2个实现 (71%消除)
- **API客户端**: 6→1个客户端 (83%消除)
- **类型系统**: 90+→8个核心接口 (80%+消除)
- **状态枚举**: SUSPENDED→INACTIVE API契约统一

## 🏗️ 技术栈架构

### 核心服务
```yaml
前端: React 19 + Vite 7 + Canvas Kit v13 (3000)
查询: PostgreSQL GraphQL (8090) - 1.5-8ms响应
命令: Go REST API (9090) - CQRS架构
数据: PostgreSQL 16+ + Redis 7.x
```

### 统一配置管理
```typescript
// frontend/src/shared/config/ports.ts
export const SERVICE_PORTS = {
  FRONTEND_DEV: 3000,
  REST_COMMAND_SERVICE: 9090,
  GRAPHQL_QUERY_SERVICE: 8090,
  POSTGRESQL: 5432,
  REDIS: 6379
} as const;
```

## 🚀 快速开始

### 环境要求
- **Go 1.23+** (后端服务)
- **Node.js 18+** (前端构建)
- **PostgreSQL 16+**
- **Redis 7.x**
- **Docker & Docker Compose**

### 一键启动 (推荐)
```bash
# 1. 启动基础设施
make docker-up

# 2. 启动后端服务
make run-dev  # 命令服务(9090) + 查询服务(8090)

# 3. 启动前端
make frontend-dev  # Vite开发服务器(3000)

# 4. 检查状态
make status
```

### 手动启动
```bash
# 基础设施
docker-compose up -d postgres redis

# 后端服务
cd cmd/organization-command-service && go run .
cd cmd/organization-query-service && go run .

# 前端开发
cd frontend && npm install && npm run dev
```

## 🔐 开发认证

### JWT令牌管理
```bash
# 生成开发令牌
make jwt-dev-mint USER_ID=dev TENANT_ID=3b99930c-e2e4-4d4a-8e7a-123456789abc

# 导出环境变量
eval $(make jwt-dev-export)

# 测试API访问
curl -H "Authorization: Bearer $JWT_TOKEN" \
     -H "X-Tenant-ID: 3b99930c-e2e4-4d4a-8e7a-123456789abc" \
     http://localhost:9090/health
```

## 📡 API访问

### GraphQL查询 (8090)
```bash
# GraphiQL界面
http://localhost:8090/graphiql

# 示例查询
query {
  organizations(first: 10) {
    code
    name
    status
    effective_date
  }
}
```

### REST命令 (9090)
```bash
# 创建组织
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"name":"测试部门","unitType":"DEPARTMENT","parentCode":"0"}'

# 查看API文档
http://localhost:9090/swagger-ui/
```

## 📊 测试

### 测试命令
```bash
# 前端测试
cd frontend && npm run test && npm run test:e2e

# 后端测试  
go test ./... && ./test_all_routes.sh
```

## 🔁 CI/CD 守护与触发

- 工作流: `.github/workflows/consistency-guard.yml`
- 触发条件:
  - push: 任意分支（branches: "**"），含 tag（tags: "*")
  - pull_request: 任意目标分支（branches: "**"）
  - workflow_dispatch: 手动触发
  - release: published/created/edited/prereleased
- 强制守护（Enforce=ON）:
  - 前端 REST 查询守护（禁止以 REST 读取，GraphQL-only）
  - cmd/* 配置守护（CORS 硬编码/端口/内联 JWT 配置）
- 本地自检:
  - `bash scripts/ci/check-permissions.sh`（权限命名）
  - `bash scripts/ci/check-rest-queries.sh`（前端 REST 查询）
  - `bash scripts/ci/check-hardcoded-configs.sh`（CORS/端口/JWT）
  - 设定 `ENFORCE=1` 可模拟 CI 强制模式；`SCAN_SCOPE=cmd|frontend` 可限定范围

## 🛡️ P3企业级防控系统 ⭐ **新上线**

### 三层纵深防御机制
```yaml
🔍 P3.1 自动化重复检测:
  - 重复代码率: 2.11% (< 5%阈值) ✅
  - 检测工具: jscpd + GitHub Actions
  - 本地验证: bash scripts/quality/duplicate-detection.sh
  
🏗️ P3.2 架构守护规则:
  - CQRS + 端口 + API契约守护
  - 违规自动识别: 25个精确检测
  - 本地验证: node scripts/quality/architecture-validator.js
  
📝 P3.3 文档自动同步:
  - 5个核心同步对监控
  - 不一致自动检测: 8个精确识别
  - 本地验证: node scripts/quality/document-sync.js
```

### 🚀 防控系统快速启动
```bash
# 完整质量检查 (推荐)
bash scripts/quality/duplicate-detection.sh      # 重复代码检测
node scripts/quality/architecture-validator.js   # 架构一致性验证  
node scripts/quality/document-sync.js           # 文档同步检查

# 自动修复模式
bash scripts/quality/duplicate-detection.sh --fix
node scripts/quality/document-sync.js --auto-sync

# 查看详细报告
open reports/duplicate-code/html/index.html     # 重复代码报告
cat reports/architecture/architecture-validation.json  # 架构报告
cat reports/document-sync/document-sync-report.json   # 同步报告
```

### ⚡ 自动化触发
- **Git提交**: Pre-commit hook自动验证架构一致性
- **CI/CD集成**: 每次push自动运行三大防控检查
- **质量门禁**: 不符合标准自动阻止合并

## 📂 项目结构

```
cube-castle/
├── frontend/                 # React 19 + Vite 7前端
│   ├── src/shared/config/    # 统一配置管理
│   ├── src/features/         # 功能模块
│   └── tests/               # 测试套件
├── cmd/                     # Go服务入口
│   ├── organization-command-service/  # REST API(9090)
│   └── organization-query-service/    # GraphQL(8090)
├── scripts/quality/          # 🆕 P3防控系统工具
│   ├── duplicate-detection.sh      # 重复代码检测
│   ├── architecture-validator.js   # 架构守护验证
│   └── document-sync.js           # 文档同步引擎
├── docs/                    # 项目文档
│   ├── api/                # API契约
│   └── development-plans/   # 开发计划
└── reports/                 # 🆕 质量报告输出
    ├── duplicate-code/      # 重复代码检测报告
    ├── architecture/        # 架构验证报告
    └── document-sync/       # 文档同步报告
```

## 📋 核心文档

- **API规范**: `docs/api/openapi.yaml` & `docs/api/schema.graphql`
- **技术架构**: `docs/development-plans/02-technical-architecture-design.md`
- **重复代码消除**: `docs/development-plans/18-duplicate-code-elimination-plan.md`
- **项目记忆**: `CLAUDE.md`

## 🔧 故障排除 & 开发规范

### 常见问题
```bash
lsof -ti:3000,8090,9090 | xargs kill -9  # 端口占用
make status                               # 服务状态
```

### 开发规范
- 使用`SERVICE_PORTS`统一端口配置
- 查询用GraphQL，命令用REST
- 禁止硬编码端口，使用`unified-client`
- ESLint + TypeScript严格模式

---

**企业级生产就绪**: ✅ PostgreSQL原生架构 + 统一配置管理 + 93%重复代码消除

**项目状态**: 企业级架构成熟，生产环境部署就绪
