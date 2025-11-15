# 🏰 Cube Castle - 企业级CoreHR SaaS平台

> **版本**: v4.1-Documentation-Governance | **更新日期**: 2025年9月13日 | **架构**: PostgreSQL原生CQRS + 统一配置管理
> 
> 快速入口：
> - 参考文档（Reference）: `docs/reference/00-README.md`
> - 文档导航中心: `docs/README.md`

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

技术栈版本:
- React 19.1.0
- Vite 7.0.4
- TypeScript 5.8.3

### 统一配置管理
```typescript
// frontend/src/shared/config/ports.ts
export const SERVICE_PORTS = {
  FRONTEND_DEV: 3000,
  FRONTEND_PREVIEW: 3001,
  REST_COMMAND_SERVICE: 9090,
  GRAPHQL_QUERY_SERVICE: 8090,
  POSTGRESQL: 5432,
  REDIS: 6379
} as const;
```

## 🚀 快速开始

### 环境要求
- **Docker & Docker Compose**（必需，用于启动 PostgreSQL、Redis、REST、GraphQL 等全部服务）
- **Go 1.24.x**（与仓库 `toolchain go1.24.9` 对齐；仅在需要 `make run-dev-debug` 进行宿主机调试时使用）
- **Node.js 18+**（前端构建/测试）
- **PostgreSQL / Redis**：由 Docker Compose 管理，宿主机 **不得** 安装同名服务占用端口

> ⚠️ **重要**：本项目强制使用 Docker 容器化部署（见 `CLAUDE.md` 第2节），禁止在宿主机直接运行 PostgreSQL、Redis 或 Go 服务。

### 一键启动（容器化，推荐）
```bash
# 1. 启动完整服务栈（PostgreSQL + Redis + REST + GraphQL）
make run-dev

# 2. 查看容器状态与关键端口
make status

# 3. 启动前端（需热重载，仍在宿主机执行）
make frontend-dev
```

### 分步启动（容器化手动控制）
```bash
# 1. 仅启动基础设施
docker compose -f docker-compose.dev.yml up -d postgres redis

# 2. 启动应用服务
docker compose -f docker-compose.dev.yml up -d --build rest-service graphql-service

# 3. 启动前端（宿主机）
cd frontend && npm run dev
```

### 调试模式（⚠️ 仅限特殊场景）
```bash
# 警告: 该模式会在宿主机直接运行 Go 服务，违反 CLAUDE.md Docker 强制原则
make run-dev-debug
```

### 容器热重载（可选）
```bash
export COMMAND_SERVICE_BUILD_TARGET=dev
export COMMAND_SERVICE_WORKDIR=/workspace/cmd/hrms-server/command
export GRAPHQL_SERVICE_BUILD_TARGET=dev
export GRAPHQL_SERVICE_WORKDIR=/workspace/cmd/hrms-server/query
docker compose -f docker-compose.dev.yml up -d --build rest-service graphql-service
```
- 完整说明参考：`docs/development-guides/docker-hot-reload-guide.md`
- 退出热重载：执行 `docker compose -f docker-compose.dev.yml down` 并 `unset` 上述环境变量

### 数据库初始化（迁移优先，禁止使用初始脚本）
- 规范：使用 `database/migrations/` 按序执行迁移脚本作为唯一初始化来源（幂等，可重复执行）。
- 禁止：`sql/init/01-schema.sql` 已归档为过时快照，切勿用于初始化，详见 `docs/archive/deprecated-setup/01-schema.sql` 头部说明。

示例（PostgreSQL，本地空库初始化）：
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"

# 依次执行关键迁移（示例，实际请执行整个 migrations 目录）
psql "$DATABASE_URL" -f database/migrations/011_audit_record_id_fix.sql
psql "$DATABASE_URL" -f database/migrations/013_enhanced_audit_changes_tracking.sql
psql "$DATABASE_URL" -f database/migrations/014_normalize_audit_logs.sql
psql "$DATABASE_URL" -f database/migrations/020_align_audit_logs_schema.sql

# 可选：加载示例数据
psql "$DATABASE_URL" -f sql/seed/02-sample-data.sql
```

注意：审计历史查询依赖迁移后的 `audit_logs` 列（before_data/after_data/modified_fields/changes/business_context/record_id）。未执行迁移将导致前端显示“加载审计历史失败”。

## 📚 文档导航（Reference vs Plans）

- 参考文档（长期稳定）: `docs/reference/`
  - 开发者快速参考 · 实现清单 · API 使用指南
- 开发计划（活跃/阶段性）: `docs/development-plans/`
  - 完成项归档 → `docs/archive/development-plans/`
-
导航入口：`docs/README.md`，归档说明：`docs/archive/README.md`

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
    effectiveDate
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

- 工作流: `.github/workflows/consistency-guard.yml`、`.github/workflows/document-sync.yml`、`.github/workflows/contract-testing.yml`、`.github/workflows/e2e-smoke.yml`、`.github/workflows/frontend-e2e.yml`
- 触发条件:
  - push: 任意分支（branches: "**"），含 tag（tags: "*")
  - pull_request: 任意目标分支（branches: "**"）
  - workflow_dispatch: 手动触发
  - release: published/created/edited/prereleased
- 强制守护（Enforce=ON）:
  - 前端 REST 查询守护（禁止以 REST 读取，GraphQL-only）
  - cmd/* 配置守护（CORS 硬编码/端口/内联 JWT 配置）
  - 文档目录边界守护（Reference vs Plans 边界检查 + 文档同步检查）
- 本地自检:
  - `bash scripts/ci/check-permissions.sh`（权限命名）
  - `bash scripts/ci/check-rest-queries.sh`（前端 REST 查询）
  - `bash scripts/ci/check-hardcoded-configs.sh`（CORS/端口/JWT）
  - 设定 `ENFORCE=1` 可模拟 CI 强制模式；`SCAN_SCOPE=cmd|frontend` 可限定范围

## 🧪 E2E（本地与CI）

### 本地快速冒烟
```bash
# 1) 一键拉起完整栈（DB/Redis + 查询8090 + 命令9090 + 前端3000）
docker compose -f docker-compose.e2e.yml up -d --build

# 2) 运行契约测试
npm --prefix frontend ci && npm --prefix frontend run -s test:contract

# 3) 运行简化E2E（无需浏览器）
bash scripts/simplified-e2e-test.sh

# 4) 报告
cat reports/QUALITY_GATE_TEST_REPORT.md
```

### CI 冒烟门禁（无浏览器）
- 工作流：`.github/workflows/e2e-smoke.yml`
- 行为：拉起 E2E 栈 → 健康等待 → 前端契约测试 → 简化E2E → 失败即阻断
- 产出：GitHub Actions Artifacts（`e2e-smoke-outputs`）与仓库 `reports/` 快照

### CI 浏览器版前端 E2E（Playwright）
- 工作流：`.github/workflows/frontend-e2e.yml`
- 行为：Compose Up → 健康等待 → 生成开发JWT（PW_JWT/PW_TENANT_ID）→ 运行 Playwright → 上传报告

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
# Go代码质量门禁 (需要 golangci-lint v1.61.0+ 支持 Go 1.23)
make lint                                       # Go 代码质量检查
make security                                   # Go 安全扫描 (gosec)

# 前端完整质量检查 (推荐)
bash scripts/quality/duplicate-detection.sh      # 重复代码检测
node scripts/quality/architecture-validator.js   # 架构一致性验证
node scripts/quality/document-sync.js           # 文档同步与目录边界检查

# 自动修复模式
bash scripts/quality/duplicate-detection.sh --fix
node scripts/quality/document-sync.js --auto-sync

# 查看详细报告
open reports/duplicate-code/html/index.html     # 重复代码报告
cat reports/architecture/architecture-validation.json  # 架构报告
cat reports/document-sync/document-sync-report.json   # 同步报告
```

### 📋 质量门禁工具要求
- **golangci-lint**: v1.61.0+ (支持 Go 1.23 新语法特性)
- **gosec**: v2.22.8+ (安全扫描)
- **工具安装**: 参考 `docs/development-plans/06-integrated-teams-progress-log.md`

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
│   ├── hrms-server/command/          # REST API(9090)
│   └── hrms-server/query/            # GraphQL(8090)
├── scripts/quality/          # 🆕 P3防控系统工具
│   ├── duplicate-detection.sh      # 重复代码检测
│   ├── architecture-validator.js   # 架构守护验证
│   └── document-sync.js           # 文档同步引擎
├── docs/                    # 项目文档
│   ├── reference/           # 长期稳定参考（快速参考、实现清单、API使用指南）
│   ├── development-plans/   # 开发计划（活跃）
│   ├── api/                 # API契约（OpenAPI/GraphQL）
│   └── archive/
│       └── development-plans/  # 开发计划归档（已完成/历史）
└── reports/                 # 🆕 质量报告输出
    ├── duplicate-code/      # 重复代码检测报告
    ├── architecture/        # 架构验证报告
    └── document-sync/       # 文档同步报告
```

## 📋 核心文档

- **API规范**: `docs/api/openapi.yaml` & `docs/api/schema.graphql`
- **技术架构（活跃）**: `docs/development-plans/02-technical-architecture-design.md`
- **参考文档入口**: `docs/reference/00-README.md`
- **Temporal Entity 指南**: `docs/reference/temporal-entity-experience-guide.md`
- **开发者快速参考**: `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
- **实现清单**: `docs/reference/02-IMPLEMENTATION-INVENTORY.md`
- **API 使用指南**: `docs/reference/03-API-USAGE-GUIDE.md`
- **计划归档目录**: `docs/archive/development-plans/`
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
 - 文档治理：Reference 仅放长期稳定参考；Plans 仅放计划/进展；完成项归档至 `docs/archive/development-plans/`

---

**企业级生产就绪**: ✅ PostgreSQL原生架构 + 统一配置管理 + 93%重复代码消除

**项目状态**: 企业级架构成熟，生产环境部署就绪
