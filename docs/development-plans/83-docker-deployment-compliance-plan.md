# 83号文档：Docker 容器化部署强制合规整改计划

**版本**: v1.0
**创建日期**: 2025-10-14
**维护团队**: 运维团队 + 后端团队 + 文档团队
**优先级**: 🔴 **P0 - 紧急**（违反 CLAUDE.md 强制原则）
**关联文档**:
- `CLAUDE.md` 第2节（Docker 容器化部署强制原则）
- `AGENTS.md` Docker 强制约束
- `reports/compliance/docker-deployment-violations-20251014.md`（违规调查报告）
- `reports/operations/postgresql-port-cleanup-20251014.md`（运维案例）

---

## 1. 背景与问题

### 1.1 问题发现
2025-10-14 完成 PostgreSQL 端口清理运维任务后，项目明确了 Docker 容器化部署的强制原则，并在 CLAUDE.md 和 AGENTS.md 中正式确立该约束。随后对项目进行合规性检查，发现**多处严重违反 Docker 强制部署原则**的情况。

### 1.2 违规概况
根据 `reports/compliance/docker-deployment-violations-20251014.md` 调查报告：

| 违规等级 | 文件数量 | 主要问题 |
|----------|----------|----------|
| 🔴 P0 严重 | 4个核心文件 | Makefile、启动脚本、配置文件默认宿主机部署 |
| 🟡 P1 中等 | 10+个文件 | 文档示例、测试脚本、Docker Compose 配置 |

**核心违规**：
- Makefile `run-dev` 目标使用 `go run` 在宿主机运行 Go 服务
- `.env` 文件优先宿主机配置（"host-based Go app - primary"）
- README.md "手动启动"部分教导宿主机部署
- `docker-compose.dev.yml` 使用 `profiles` 隐藏应用服务容器

### 1.3 影响评估
- **开发流程违规**: 默认开发流程违反强制原则，新开发者被误导
- **文档不一致**: CLAUDE.md/AGENTS.md 与实际配置/脚本矛盾
- **运维风险**: 宿主机部署导致环境不一致，可能引发端口冲突（如本次 PostgreSQL 5432 端口问题）
- **原则权威性受损**: 强制原则无法落地执行

---

## 2. 目标与范围

### 2.1 整改目标
1. **强制合规**: 所有配置、脚本、文档严格遵循 Docker 强制部署原则
2. **默认正确**: 开发者使用默认流程（`make run-dev`）即符合原则
3. **文档一致**: 所有文档与 CLAUDE.md/AGENTS.md 保持一致
4. **CI 守护**: 建立自动化检查防止违规代码合并

### 2.2 整改范围

#### Phase 1: P0 紧急修复（本周内完成）
- Makefile 核心目标（run-dev, run-auth-rs256-sim）
- .env 配置文件
- docker-compose.dev.yml 服务配置
- README.md 快速开始部分

#### Phase 2: P1 文档与脚本整改（2周内完成）
- 开发者快速参考文档
- 启动脚本废弃与重构
- 测试脚本 Docker 化
- CI 合规检查

#### Phase 3: P2 长期优化（1个月内完成）
- 开发热重载方案（Air + volumes）
- 完整文档体系更新
- 最佳实践文档编写

### 2.3 不在范围内
- ❌ 前端开发服务器（仍在宿主机运行 Vite，因需热重载）
- ❌ 宿主机工具（psql、redis-cli 等通过端口映射访问容器）
- ❌ CI/CD 环境（已使用容器化部署）

---

## 3. Phase 1: P0 紧急修复（DDL: 2025-10-18）

### 3.1 修复 Makefile（优先级最高）

#### 3.1.1 run-dev 目标

**当前问题**:
```makefile
# Line 111, 114: 宿主机运行 Go 服务
go run ./cmd/organization-command-service/main.go &
go run ./cmd/organization-query-service/main.go &
```

**修复方案**:
```makefile
run-dev:
	@echo "🚀 启动开发环境（Docker 强制）..."
	@echo "🔐 检查 JWT 密钥..."
	$(MAKE) jwt-dev-setup
	@echo "🐳 拉起完整服务栈（基础设施 + 应用服务）..."
	docker compose -f docker-compose.dev.yml up -d --build postgres redis graphql-service rest-service
	@echo "⏳ 等待服务健康..."
	@sleep 8
	@echo "🩺 健康检查："
	-@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  curl -fsS http://localhost:9090/health >/dev/null && echo "  ✅ command-service ok" && break || \
	  (echo "  ⏳ 等待 command-service..." && sleep 2); \
	done || echo "  ⚠️  command-service 未就绪，请检查: docker compose -f docker-compose.dev.yml logs rest-service"
	-@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  curl -fsS http://localhost:8090/health >/dev/null && echo "  ✅ query-service ok" && break || \
	  (echo "  ⏳ 等待 query-service..." && sleep 2); \
	done || echo "  ⚠️  query-service 未就绪，请检查: docker compose -f docker-compose.dev.yml logs graphql-service"
	@echo "✅ 服务已就绪"
	@echo "📊 查看日志: docker compose -f docker-compose.dev.yml logs -f graphql-service rest-service"
	@echo "🛑 停止服务: make docker-down"
```

#### 3.1.2 run-auth-rs256-sim 目标

**修复方案**: 同样改用 `docker compose -f docker-compose.dev.yml up`，或废弃此目标（功能已被 run-dev 覆盖）

#### 3.1.3 新增 run-dev-debug 目标（调试专用）

```makefile
# 调试模式：宿主机运行 Go 服务（仅用于特殊调试场景）
run-dev-debug:
	@echo "⚠️  警告: 调试模式 - 违反 Docker 强制原则"
	@echo "此模式仅用于特殊调试场景（如 Delve 断点调试）"
	@echo "日常开发请使用: make run-dev"
	@echo ""
	@read -p "确认继续？(y/N) " -n 1 -r REPLY; \
	echo; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
	  echo "已取消"; exit 1; \
	fi
	@echo "🧹 清理端口占用 (9090/8090)..."
	# ... 原有 run-dev 逻辑 ...
```

**验收标准**:
- [ ] `make run-dev` 启动 Docker 容器而非宿主机服务
- [ ] 健康检查通过，服务可正常访问
- [ ] 日志输出清晰，提示如何查看容器日志

---

### 3.2 修复 .env 配置文件

#### 3.2.1 当前问题
```bash
# Line 1-2: 注释误导性强
# Database Configuration (for host-based Go app - primary)
DATABASE_URL=postgresql://user:password@localhost:5432/cubecastle?sslmode=disable
```

#### 3.2.2 修复方案
```bash
# =============================================================================
# 🐳 Docker Compose 环境变量配置（强制遵循 CLAUDE.md Docker 原则）
# =============================================================================
#
# ⚠️  重要说明:
# 1. 本项目强制使用 Docker 容器化部署（CLAUDE.md 第2节）
# 2. 应用服务（Go）必须在容器内运行，不得在宿主机直接执行 go run
# 3. 以下配置分为两类：
#    - 容器内连接：供 Docker 容器内应用使用（主机名为服务名，如 postgres、redis）
#    - 宿主机工具：供宿主机 psql、redis-cli 等工具通过端口映射访问（主机名 localhost）
#
# =============================================================================

# -----------------------------------------------------------------------------
# Database Configuration
# -----------------------------------------------------------------------------

# 容器内连接配置（应用服务使用，默认值）
# 用途: Docker Compose 中的 graphql-service、rest-service 连接数据库
DATABASE_URL=postgresql://user:password@postgres:5432/cubecastle?sslmode=disable

# 宿主机工具访问配置（仅供 psql、迁移脚本等宿主机工具使用）
# 用途: make db-migrate-all、手动执行 psql 命令
# 注意: 通过 Docker 端口映射 localhost:5432 -> postgres:5432 访问
DATABASE_URL_HOST_TOOLS=postgresql://user:password@localhost:5432/cubecastle?sslmode=disable

# PostgreSQL 容器配置
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=cubecastle

# -----------------------------------------------------------------------------
# Redis Configuration
# -----------------------------------------------------------------------------

# 容器内连接配置（应用服务使用）
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DB=0

# 宿主机工具访问（通过端口映射）
REDIS_HOST_TOOLS=localhost
REDIS_PORT_TOOLS=6379

# -----------------------------------------------------------------------------
# Application Configuration
# -----------------------------------------------------------------------------

APP_PORT=8080
APP_ENV=development
LOG_LEVEL=info

# -----------------------------------------------------------------------------
# Security Configuration
# -----------------------------------------------------------------------------

JWT_SECRET=cube-castle-development-secret-key-please-change-in-production
JWT_EXPIRY=24h
JWT_ALG=RS256
JWT_PRIVATE_KEY_PATH=./secrets/dev-jwt-private.pem
JWT_PUBLIC_KEY_PATH=./secrets/dev-jwt-public.pem
JWT_KEY_ID=bff-key-1

# -----------------------------------------------------------------------------
# Tenant Configuration
# -----------------------------------------------------------------------------

DEFAULT_TENANT_ID=3b99930c-4dc6-4cc9-8e4d-7d960a931cb9

# -----------------------------------------------------------------------------
# Temporal Configuration
# -----------------------------------------------------------------------------

TEMPORAL_HOST_PORT=temporal-server:7233
TEMPORAL_NAMESPACE=default
```

**验收标准**:
- [ ] 移除 "host-based Go app - primary" 注释
- [ ] 默认 `DATABASE_URL` 使用容器内主机名（postgres）
- [ ] 添加清晰的 Docker 强制原则说明
- [ ] 区分容器内配置与宿主机工具配置

---

### 3.3 修复 docker-compose.dev.yml

#### 3.3.1 当前问题
```yaml
graphql-service:
  # ...
  profiles: ["services"]  # 导致默认不启动

rest-service:
  # ...
  profiles: ["services"]  # 导致默认不启动
```

#### 3.3.2 修复方案
```yaml
services:
  # ... postgres, redis 保持不变 ...

  # GraphQL查询服务 (端口8090)
  graphql-service:
    build:
      context: .
      dockerfile: cmd/organization-query-service/Dockerfile  # 需新增/补齐
    container_name: cubecastle-graphql
    environment:
      # 使用容器内主机名
      - DATABASE_URL=postgres://user:password@postgres:5432/cubecastle?sslmode=disable
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=""
      - PORT=8090
      - GIN_MODE=debug
      # 从 .env 继承 JWT 配置
      - JWT_ALG=${JWT_ALG}
      - JWT_JWKS_URL=http://rest-service:9090/.well-known/jwks.json
    ports:
      - "8090:8090"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    # profiles: ["services"]  # 🔴 已移除：默认启动

  # REST命令服务 (端口9090)
  rest-service:
    build:
      context: .
      dockerfile: cmd/organization-command-service/Dockerfile
    container_name: cubecastle-rest
    environment:
      - DATABASE_URL=postgres://user:password@postgres:5432/cubecastle?sslmode=disable
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=""
      - PORT=9090
      - GIN_MODE=debug
      # 从 .env 继承 JWT 配置
      - JWT_ALG=${JWT_ALG}
      - JWT_MINT_ALG=${JWT_ALG}
      - JWT_PRIVATE_KEY_PATH=/secrets/dev-jwt-private.pem
      - JWT_PUBLIC_KEY_PATH=/secrets/dev-jwt-public.pem
      - JWT_KEY_ID=${JWT_KEY_ID}
    ports:
      - "9090:9090"
    volumes:
      # 挂载 secrets 目录以访问 JWT 密钥
      - ./secrets:/secrets:ro
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    # profiles: ["services"]  # 🔴 已移除：默认启动

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  default:
    name: cubecastle-network
```

**关键变更**:
1. 移除 `profiles: ["services"]`，默认启动所有服务
2. 环境变量使用容器内主机名（postgres、redis）
3. JWT 配置通过 volumes 挂载 secrets 目录
4. 添加注释说明为何移除 profiles
5. Phase 1 内新增 `cmd/organization-query-service/Dockerfile`，确保 Compose 构建路径有效

> 注：当前仓库尚未提供 GraphQL 查询服务的 Dockerfile，需在 Phase 1 内新增 `cmd/organization-query-service/Dockerfile`（可参考命令服务 Dockerfile 的分层结构），方可使上述 Compose 片段生效。

**验收标准**:
- [ ] `docker compose -f docker-compose.dev.yml up -d` 启动所有服务（postgres, redis, graphql-service, rest-service）
- [ ] 容器间网络互通（graphql-service 可访问 postgres、redis、rest-service）
- [ ] JWT JWKS 配置正确（graphql-service 从 rest-service 获取公钥）

---

### 3.4 修复 README.md

#### 3.4.1 快速开始部分

**当前问题**:
```markdown
### 手动启动
​```bash
# 后端服务
cd cmd/organization-command-service && go run .
cd cmd/organization-query-service && go run .
​```
```

**修复方案**:
```markdown
## 🚀 快速开始

### 环境要求
- **Docker & Docker Compose** (必需)
- **Go 1.23+** (仅用于本地开发调试，日常开发不需要)
- **Node.js 18+** (前端构建)

⚠️ **重要**: 本项目强制使用 Docker 容器化部署（详见 `CLAUDE.md` 第2节）。所有服务（PostgreSQL、Redis、Go 应用）必须在 Docker 容器内运行，不得在宿主机直接安装或执行。

### 一键启动（推荐，符合 Docker 强制原则）
​```bash
# 1. 启动完整服务栈（基础设施 + 应用服务）
make run-dev  # 或 docker compose -f docker-compose.dev.yml up -d --build

# 2. 检查服务状态
make status
# 预期输出:
#   cubecastle-postgres   ... Up (healthy)   0.0.0.0:5432->5432/tcp
#   cubecastle-redis      ... Up (healthy)   0.0.0.0:6379->6379/tcp
#   cubecastle-rest       ... Up             0.0.0.0:9090->9090/tcp
#   cubecastle-graphql    ... Up             0.0.0.0:8090->8090/tcp

# 3. 查看服务日志
docker compose -f docker-compose.dev.yml logs -f graphql-service rest-service

# 4. 启动前端（仍在宿主机，因需热重载）
make frontend-dev  # 或 cd frontend && npm run dev
​```

### 分步启动（手动控制，仍符合 Docker 原则）
​```bash
# 1. 仅启动基础设施
docker compose -f docker-compose.dev.yml up -d postgres redis

# 2. 启动应用服务
docker compose -f docker-compose.dev.yml up -d --build graphql-service rest-service

# 3. 启动前端
cd frontend && npm run dev
​```

### 调试模式（⚠️ 违反 Docker 原则，仅限特殊调试场景）
​```bash
# ⚠️ 警告: 此模式违反 CLAUDE.md Docker 强制原则
# 仅用于特殊调试场景（如 Delve 断点调试、性能分析）
# 日常开发请使用上方"一键启动"

# 1. 启动基础设施
docker compose -f docker-compose.dev.yml up -d postgres redis

# 2. 宿主机运行 Go 服务（调试模式）
make run-dev-debug
# 或手动运行:
cd cmd/organization-command-service && go run .
cd cmd/organization-query-service && go run .

# 3. 调试完成后，切换回容器模式
make docker-down && make run-dev
​```

### 停止服务
​```bash
# 停止所有服务
make docker-down  # 或 docker compose -f docker-compose.dev.yml down

# 停止并清理数据卷（⚠️ 会删除数据库数据）
docker compose -f docker-compose.dev.yml down -v
​```
```

**验收标准**:
- [ ] "一键启动"部分仅包含 Docker 命令
- [ ] "手动启动"改名为"分步启动"，仍使用 Docker
- [ ] 新增"调试模式"部分，带明确警告
- [ ] 环境要求明确 Docker 为必需，Go 为可选

---

### 3.5 Phase 1 验收标准汇总

| 检查项 | 验收标准 | 证据 |
|--------|----------|------|
| Makefile | `make run-dev` 启动 Docker 容器 | 执行输出显示 `docker compose -f docker-compose.dev.yml up` |
| .env | 移除 "host-based primary" 注释 | 文件内容检查 |
| docker-compose.dev.yml | 移除 profiles，默认启动所有服务 | `docker compose -f docker-compose.dev.yml up -d` 启动4个容器 |
| README.md | "一键启动"仅 Docker 命令 | 文档内容检查 |
| 集成测试 | 完整服务栈启动并通过健康检查 | `curl http://localhost:9090/health` 返回 200 |

---

## 4. Phase 2: P1 文档与脚本整改（DDL: 2025-10-28）

### 4.1 更新开发者快速参考

**文件**: `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`

#### 4.1.1 在所有 localhost 示例前添加说明

```markdown
### 数据库初始化（Docker 强制）

> 🐳 **Docker 部署说明**
> 本项目强制使用 Docker 容器化部署（CLAUDE.md 第2节）。以下命令中的 `localhost:5432` 是通过 Docker 端口映射访问容器数据库，**并非宿主机安装的 PostgreSQL**。
>
> - 应用服务（Go）: 使用 `postgres:5432`（容器内主机名）
> - 宿主机工具（psql、迁移脚本）: 使用 `localhost:5432`（通过端口映射）

​```bash
# 环境变量（宿主机工具通过端口映射访问 Docker 容器）
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
make db-migrate-all
​```
```

#### 4.1.2 更新常用命令速查

```markdown
### 开发环境启动
​```bash
make docker-up          # ❌ 已废弃 - 仅启动基础设施（不完整）
make run-dev            # ✅ 推荐 - 启动完整服务栈（Docker 强制）
make frontend-dev       # 启动前端开发服务器 (端口3000)
make status             # 查看所有服务状态
make docker-down        # 停止所有 Docker 服务
​```
```

**验收标准**:
- [ ] 所有 localhost 示例添加 Docker 说明框
- [ ] 更新命令速查，标注 `make docker-up` 为不完整
- [ ] 推荐使用 `make run-dev`

---

### 4.2 废弃宿主机部署脚本

**文件**: `scripts/dev-start-simple.sh`

#### 4.2.1 添加废弃警告

```bash
#!/bin/bash

# =============================================================================
# ⚠️  此脚本已废弃 - 违反 Docker 强制部署原则
# =============================================================================
#
# 原因: 脚本在宿主机运行 Go 服务，违反 CLAUDE.md 第2节 Docker 强制原则
# 替代方案:
#   - 推荐: make run-dev
#   - 或: docker compose -f docker-compose.dev.yml up -d --build
#
# 详见:
#   - CLAUDE.md 第2节（Docker 容器化部署强制原则）
#   - AGENTS.md Docker 强制约束
#   - docs/development-plans/83-docker-deployment-compliance-plan.md
#
# 废弃时间: 2025-10-14
# 计划删除: 2025-11-14（废弃1个月后）
#
# =============================================================================

echo "⚠️  警告: 此脚本已废弃，违反 Docker 强制部署原则"
echo ""
echo "请使用符合规范的启动方式:"
echo "  make run-dev"
echo "  或: docker compose -f docker-compose.dev.yml up -d --build"
echo ""
echo "详见: docs/development-plans/83-docker-deployment-compliance-plan.md"
echo ""
read -p "是否继续使用已废弃脚本？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消。推荐使用: make run-dev"
    exit 1
fi

echo "⚠️  继续执行废弃脚本..."
echo ""

# ... 原有代码 ...
```

**验收标准**:
- [ ] 脚本开头添加废弃警告框
- [ ] 执行时显示警告并要求用户确认
- [ ] 提供替代方案说明

---

### 4.3 添加 CI 合规检查

**文件**: `.github/workflows/docker-compliance.yml`

```yaml
name: Docker Deployment Compliance Check

on:
  push:
    branches: ["**"]
  pull_request:
    branches: ["**"]

jobs:
  check-docker-compliance:
    name: Check Docker Deployment Compliance
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Check Makefile for go run violations
        run: |
          echo "🔍 检查 Makefile 是否包含 'go run' 命令..."
          if grep -n "go run.*cmd/" Makefile | grep -v "run-dev-debug" | grep -v "^#"; then
            echo "❌ Makefile 包含 'go run' 命令（违反 Docker 强制原则）"
            echo "允许例外: run-dev-debug 目标（调试专用）"
            echo "详见: CLAUDE.md 第2节、docs/development-plans/83-docker-deployment-compliance-plan.md"
            exit 1
          fi
          echo "✅ Makefile 合规检查通过"

      - name: Check .env for host-based primary config
        run: |
          echo "🔍 检查 .env 配置优先级..."
          if grep -i "host-based.*primary" .env; then
            echo "❌ .env 文件优先宿主机配置（违反 Docker 强制原则）"
            echo "正确做法: 优先容器内连接配置（主机名为 postgres、redis）"
            exit 1
          fi
          echo "✅ .env 配置合规检查通过"

      - name: Check docker-compose.yml for hidden services
        run: |
          echo "🔍 检查 docker-compose.yml 服务可见性..."
          if grep -A 5 "graphql-service:" docker-compose.dev.yml | grep "profiles:.*services"; then
            echo "❌ docker-compose.dev.yml 使用 profiles 隐藏应用服务"
            echo "正确做法: 移除 profiles，默认启动所有服务"
            exit 1
          fi
          if grep -A 5 "rest-service:" docker-compose.dev.yml | grep "profiles:.*services"; then
            echo "❌ docker-compose.dev.yml 使用 profiles 隐藏应用服务"
            echo "正确做法: 移除 profiles，默认启动所有服务"
            exit 1
          fi
          echo "✅ docker-compose.yml 合规检查通过"

      - name: Check scripts for go run violations
        run: |
          echo "🔍 检查脚本文件是否包含 'go run' 命令..."
          VIOLATIONS=$(grep -r "go run.*cmd/" scripts/ --include="*.sh" | grep -v "废弃" | grep -v "调试" || true)
          if [ -n "$VIOLATIONS" ]; then
            echo "❌ 发现脚本包含 'go run' 命令:"
            echo "$VIOLATIONS"
            echo ""
            echo "正确做法: 使用 docker compose -f docker-compose.dev.yml up 或在脚本顶部添加废弃警告"
            exit 1
          fi
          echo "✅ 脚本合规检查通过"

      - name: Summary
        if: success()
        run: |
          echo "✅ 所有 Docker 部署合规检查通过"
          echo ""
          echo "检查项:"
          echo "  ✅ Makefile 不包含 go run（除调试目标）"
          echo "  ✅ .env 配置优先容器内连接"
          echo "  ✅ docker-compose.yml 默认启动所有服务"
          echo "  ✅ 脚本不包含 go run（除废弃脚本）"
```

**验收标准**:
- [ ] CI 工作流创建并启用
- [ ] 提交违规代码时 CI 失败
- [ ] 提交合规代码时 CI 通过

---

### 4.4 Phase 2 验收标准汇总

| 检查项 | 验收标准 | 证据 |
|--------|----------|------|
| 开发者快速参考 | 所有 localhost 示例添加 Docker 说明 | 文档内容检查 |
| 废弃脚本 | dev-start-simple.sh 添加警告并标注废弃 | 执行脚本时显示警告 |
| CI 合规检查 | 工作流创建并能检测违规 | PR 提交违规代码时 CI 失败 |
| 文档同步 | 所有参考文档与 CLAUDE.md 一致 | 交叉检查文档内容 |

---

## 5. Phase 3: P2 长期优化（DDL: 2025-11-14）

### 5.1 开发热重载方案

#### 5.1.1 目标
解决 Docker 容器部署后开发效率下降的问题，提供接近 `go run` 的开发体验。

#### 5.1.2 技术方案：Air + Volume 挂载

**Dockerfile 修改**（多阶段构建）:
```dockerfile
# 开发阶段：使用 Air 热重载
FROM golang:1.23-alpine AS dev

WORKDIR /app

# 安装 Air
RUN go install github.com/cosmtrek/air@latest

# 复制 go.mod 和 go.sum（利用 Docker 缓存）
COPY go.mod go.sum ./
RUN go mod download

# 复制项目代码
COPY . .

# 暴露端口
EXPOSE 9090

# 使用 Air 启动（支持热重载）
CMD ["air", "-c", ".air.toml"]

# -----------------------------------------------------------------------------

# 生产阶段：最小化镜像
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 构建二进制
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/organization-command-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /bin/server .
EXPOSE 9090
CMD ["./server"]
```

**.air.toml 配置**:
```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/organization-command-service"
  delay = 1000
  exclude_dir = ["tmp", "vendor", "frontend"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = true
  follow_symlink = false
  include_dir = ["cmd", "internal"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[log]
  time = false

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

[misc]
  clean_on_exit = true
```

**docker-compose.dev.yml 修改**:
```yaml
rest-service:
  build:
    context: .
    dockerfile: cmd/organization-command-service/Dockerfile
    target: dev  # 使用开发阶段
  container_name: cubecastle-rest
  environment:
    # ... 环境变量 ...
  ports:
    - "9090:9090"
  volumes:
    # 挂载源代码实现热重载
    - ./cmd/organization-command-service:/app/cmd/organization-command-service
    - ./internal:/app/internal
    - ./database:/app/database
    # Air 临时目录（避免污染宿主机）
    - rest-tmp:/app/tmp
  depends_on:
    # ...
  restart: unless-stopped

volumes:
  # ... 其他 volumes ...
  rest-tmp:
    driver: local
  graphql-tmp:
    driver: local
```

**验收标准**:
- [ ] 修改 Go 代码后容器内自动重新编译
- [ ] 重启延迟 < 3秒
- [ ] 不影响生产环境构建

---

### 5.2 完整文档体系更新

#### 5.2.1 需要更新的文档

| 文档 | 更新内容 |
|------|----------|
| `docs/reference/03-API-AND-TOOLS-GUIDE.md` | 添加 Docker 说明，更新示例 |
| `docs/architecture/*.md` | 明确架构图中 Docker 部署方式 |
| `docs/development-tools/*.md` | 更新工具使用说明（容器内执行） |

#### 5.2.2 新增文档

**文件**: `docs/reference/04-DOCKER-BEST-PRACTICES.md`

```markdown
# Docker 容器化部署最佳实践

## 1. 强制原则（来自 CLAUDE.md）
- 所有服务、数据库、中间件必须在 Docker 容器内运行
- 严禁在宿主机直接安装 PostgreSQL、Redis、Temporal 等
- 端口冲突时卸载宿主服务，不得调整容器端口映射

## 2. 开发流程
### 2.1 启动服务
​```bash
make run-dev  # 启动完整服务栈（或 docker compose -f docker-compose.dev.yml up -d --build）
​```

### 2.2 查看日志
​```bash
docker compose -f docker-compose.dev.yml logs -f graphql-service rest-service
​```

### 2.3 进入容器调试
​```bash
docker exec -it cubecastle-rest sh
​```

## 3. 配置说明
### 3.1 环境变量
- 容器内: `DATABASE_URL=postgres://user:password@postgres:5432/...`
- 宿主机工具: `DATABASE_URL=postgres://user:password@localhost:5432/...`

### 3.2 端口映射
- PostgreSQL: `localhost:5432 -> postgres:5432`
- Redis: `localhost:6379 -> redis:6379`
- REST API: `localhost:9090 -> rest-service:9090`
- GraphQL API: `localhost:8090 -> graphql-service:8090`

## 4. 常见问题
### Q: 如何实现热重载？
A: 使用 Air + Volume 挂载（详见 83号计划 Phase 3.1）

### Q: 如何断点调试？
A: 使用 `make run-dev-debug` 临时启用宿主机调试模式

### Q: 端口被占用怎么办？
A: 卸载宿主服务（如 `sudo apt remove postgresql*`），不得修改容器端口映射
​```
```

**验收标准**:
- [ ] 所有文档完成 Docker 说明更新
- [ ] 新增最佳实践文档
- [ ] 文档交叉引用正确

---

### 5.3 Phase 3 验收标准汇总

| 检查项 | 验收标准 | 证据 |
|--------|----------|------|
| 热重载方案 | 代码修改后容器自动重启 < 3秒 | 实际测试 |
| 文档完整性 | 所有文档包含 Docker 说明 | 文档审查 |
| 最佳实践文档 | 创建并包含常见问题解答 | 文档存在性检查 |

---

## 6. 里程碑与时间线

| 里程碑 | 内容 | 负责人 | DDL | 状态 |
|--------|------|--------|-----|------|
| M1 | Makefile 修复完成 | 后端团队 | 2025-10-16 | ☐ |
| M2 | .env + docker-compose.yml 修复完成 | 运维团队 | 2025-10-17 | ☐ |
| M3 | README.md 修复完成 | 文档团队 | 2025-10-18 | ☐ |
| M4 | Phase 1 集成测试通过 | 后端团队 | 2025-10-18 | ☐ |
| M5 | 开发者快速参考更新完成 | 文档团队 | 2025-10-21 | ☐ |
| M6 | CI 合规检查上线 | 运维团队 | 2025-10-25 | ☐ |
| M7 | 废弃脚本标注完成 | 后端团队 | 2025-10-28 | ☐ |
| M8 | Phase 2 完成，所有 P1 修复 | 全体 | 2025-10-28 | ☐ |
| M9 | 热重载方案实现 | 后端团队 | 2025-11-07 | ☐ |
| M10 | 文档体系更新完成 | 文档团队 | 2025-11-14 | ☐ |
| M11 | Phase 3 完成，长期优化到位 | 全体 | 2025-11-14 | ☐ |

---

## 7. 风险与应对

| 风险 | 影响 | 概率 | 应对措施 |
|------|------|------|----------|
| 开发者抵触容器化（构建慢） | 整改推进缓慢 | 中 | Phase 3 提供热重载方案，接近 go run 体验 |
| CI 检查过严导致误报 | 开发流程受阻 | 中 | 允许特定场景例外（如 run-dev-debug），注释说明 |
| 历史脚本依赖 go run | 脚本失效 | 低 | 分阶段废弃，提供1个月过渡期 |
| Docker 镜像构建失败 | 服务无法启动 | 低 | 测试多阶段构建，准备 fallback Dockerfile |

---

## 8. 成功标准

### 8.1 强制合规（P0）
- [ ] `make run-dev` 启动 Docker 容器，不使用 `go run`
- [ ] `.env` 配置优先容器内连接，无"host-based primary"注释
- [ ] `docker-compose.dev.yml` 默认启动所有服务（无 profiles 隐藏）
- [ ] README.md "一键启动"仅包含 Docker 命令

### 8.2 CI 守护（P1）
- [ ] CI 工作流检测并阻止违规代码合并
- [ ] 所有 PR 必须通过 Docker 合规检查

### 8.3 文档一致（P1）
- [ ] 所有文档与 CLAUDE.md/AGENTS.md Docker 强制原则一致
- [ ] 所有 localhost 示例添加 Docker 说明

### 8.4 开发体验（P2）
- [ ] 提供热重载方案，代码修改 < 3秒自动重启
- [ ] 最佳实践文档覆盖常见问题

---

## 9. 参考资料

- **违规调查报告**: `reports/compliance/docker-deployment-violations-20251014.md`
- **强制原则来源**: `CLAUDE.md` 第2节、第5节
- **执行规范**: `AGENTS.md` Docker 容器化部署强制约束
- **运维案例**: `reports/operations/postgresql-port-cleanup-20251014.md`
- **Air 官方文档**: https://github.com/cosmtrek/air

---

## 10. 附录：快速修复检查清单

### Phase 1 (P0) 检查清单
- [ ] Makefile: `run-dev` 改用 `docker compose -f docker-compose.dev.yml up`
- [ ] Makefile: 新增 `run-dev-debug` 调试目标（带警告）
- [ ] .env: 移除 "host-based primary"，优先容器内配置
- [ ] docker-compose.dev.yml: 移除 `profiles: ["services"]`
- [ ] README.md: "一键启动"仅 Docker 命令
- [ ] README.md: 新增"调试模式"部分（带警告）
- [ ] 集成测试: `make run-dev` 启动服务并通过健康检查
- [ ] 新增 `cmd/organization-query-service/Dockerfile` 并通过 compose 构建验证

### Phase 2 (P1) 检查清单
- [ ] 开发者快速参考: 所有 localhost 示例添加 Docker 说明
- [ ] dev-start-simple.sh: 添加废弃警告
- [ ] CI 工作流: 创建 docker-compliance.yml
- [ ] CI 测试: 提交违规代码验证 CI 失败

### Phase 3 (P2) 检查清单
- [ ] Dockerfile: 添加 dev target 支持 Air
- [ ] .air.toml: 创建配置文件
- [ ] docker-compose.dev.yml: 添加 volumes 挂载
- [ ] 最佳实践文档: 创建并完善
- [ ] 所有文档: 完成 Docker 说明更新

---

**文档版本**: v1.0
**创建时间**: 2025-10-14
**下次更新**: Phase 1 完成后更新进度
