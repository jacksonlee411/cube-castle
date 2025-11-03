# Plan 221 - Docker 集成测试基座建设

**文档编号**: 221
**标题**: 容器化集成测试环境 - 基础设施建设
**创建日期**: 2025-11-04
**分支**: `feature/204-phase2-infrastructure`
**版本**: v1.0
**关联计划**: Plan 217（database）、Plan 210（迁移脚本）、Plan 215（Phase2 执行日志）

---

## 1. 概述

### 1.1 目标

建立 Docker 化的集成测试环境，支持 Goose 数据库迁移的完整测试流程，为所有模块的集成测试提供基础设施。

**关键成果**:
- ✅ `docker-compose.test.yml` 配置
- ✅ 集成测试启动脚本
- ✅ Makefile 目标更新（`make test-db`）
- ✅ 测试数据初始化脚本
- ✅ CI/CD 集成配置

### 1.2 为什么需要 Docker 集成测试

- **环境一致性** - 本地与 CI 环境完全相同
- **隔离性** - 测试数据互不影响
- **可重复性** - 测试可完全复现
- **迁移验证** - 充分验证 Goose 迁移脚本

### 1.3 时间计划

- **计划完成**: Week 4 Day 2 (Day 16)
- **交付周期**: 1 天
- **负责人**: QA + DevOps
- **前置依赖**: Plan 217（database），Plan 210（迁移脚本）

---

## 2. 需求分析

### 2.1 功能需求

#### 需求 1: Docker 环境配置

```yaml
# docker-compose.test.yml 应支持:
- PostgreSQL 15 数据库容器
- 自动初始化（create database if not exists）
- 端口映射（5432:5432）
- 卷挂载（用于数据持久化或初始化脚本）
- 环境变量配置（用户、密码、数据库名）
```

#### 需求 2: 数据库迁移测试

```bash
# 标准的迁移测试流程:
1. docker-compose -f docker-compose.test.yml up -d
2. 等待数据库就绪
3. goose -dir database/migrations up
4. 运行集成测试
5. goose -dir database/migrations down
6. 验证数据库恢复到初始状态
```

#### 需求 3: 测试数据初始化

```go
// 提供测试 fixture，支持:
- 基础数据初始化（组织、部门、职位）
- 测试场景数据生成
- 数据清理（TRUNCATE）
```

### 2.2 非功能需求

| 需求 | 标准 | 说明 |
|------|------|------|
| **启动时间** | < 10s | Docker 容器就绪 |
| **稳定性** | 100% 可重复 | 多次运行结果一致 |
| **资源占用** | < 500MB | 内存使用合理 |
| **清理** | 自动清理 | 测试后容器关闭 |

---

## 3. 详细实现

### 3.1 docker-compose.test.yml 配置

```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    container_name: cube-castle-test-db
    environment:
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpassword
      POSTGRES_DB: testdb
      # 初始化脚本
      POSTGRES_INITDB_ARGS: "-c log_min_duration_statement=1000"
    ports:
      - "5433:5432"  # 使用 5433 避免与开发环境冲突
    volumes:
      # 初始化脚本
      - ./scripts/test/init-db.sql:/docker-entrypoint-initdb.d/01-init.sql
      # Goose 迁移脚本
      - ./database/migrations:/migrations:ro
      # 数据卷（测试间隔清理）
      - postgres-test-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U testuser -d testdb"]
      interval: 2s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - test-network

networks:
  test-network:
    driver: bridge

volumes:
  postgres-test-data:
    driver: local
```

### 3.2 初始化脚本 (scripts/test/init-db.sql)

```sql
-- scripts/test/init-db.sql
-- PostgreSQL 初始化脚本，容器启动时自动执行

-- 创建测试数据库的基础扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 创建 goose 版本表
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER NOT NULL PRIMARY KEY,
    version_id BIGINT NOT NULL,
    is_applied BOOLEAN NOT NULL,
    tstamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 记录：此脚本在 Goose 迁移脚本之前执行
-- Goose 迁移脚本会在这之后运行
```

### 3.3 集成测试启动脚本 (scripts/run-integration-tests.sh)

```bash
#!/bin/bash
# scripts/run-integration-tests.sh
# 启动 Docker 测试环境并运行集成测试

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATIONS_DIR="$PROJECT_ROOT/database/migrations"
DOCKER_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理函数（脚本退出时执行）
cleanup() {
    log_info "Cleaning up Docker containers..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down -v 2>/dev/null || true
}

trap cleanup EXIT

# 1. 启动 Docker 容器
log_info "Starting Docker containers..."
docker-compose -f "$DOCKER_COMPOSE_FILE" up -d

# 2. 等待数据库就绪
log_info "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if docker-compose -f "$DOCKER_COMPOSE_FILE" exec -T postgres-test pg_isready -U testuser &>/dev/null; then
        log_info "PostgreSQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "PostgreSQL failed to start"
        exit 1
    fi
    sleep 1
done

# 3. 获取数据库连接信息
DB_HOST="localhost"
DB_PORT="5433"
DB_USER="testuser"
DB_PASSWORD="testpassword"
DB_NAME="testdb"
DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# 4. 运行 Goose 迁移
log_info "Running Goose migrations (up)..."
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" up

log_info "Goose migrations completed successfully"

# 5. 运行 Go 集成测试
log_info "Running Go integration tests..."
export DATABASE_URL
go test -v -race -coverprofile=coverage-test.out ./cmd/hrms-server/command/internal/... ./cmd/hrms-server/query/internal/...

# 6. 验证回滚
log_info "Verifying Goose rollback..."
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" down

log_info "Verifying rollback completed successfully"

log_info "All tests passed!"
```

### 3.4 Makefile 更新

```makefile
# Makefile 新增目标

.PHONY: test-db
test-db: ## Run integration tests with Docker
	@chmod +x scripts/run-integration-tests.sh
	@scripts/run-integration-tests.sh

.PHONY: test-db-up
test-db-up: ## Start test database
	@docker-compose -f docker-compose.test.yml up -d

.PHONY: test-db-down
test-db-down: ## Stop test database
	@docker-compose -f docker-compose.test.yml down -v

.PHONY: test-db-logs
test-db-logs: ## Show test database logs
	@docker-compose -f docker-compose.test.yml logs -f postgres-test

.PHONY: test-db-psql
test-db-psql: ## Connect to test database with psql
	@docker-compose -f docker-compose.test.yml exec postgres-test psql -U testuser -d testdb
```

### 3.5 CI 集成配置 (.github/workflows/integration-test.yml)

```yaml
name: Integration Tests

on: [pull_request, push]

jobs:
  integration-test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: testuser
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5433:5432

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install Goose
        run: go install github.com/pressly/goose/v3/cmd/goose@latest

      - name: Run Goose migrations
        env:
          GOOSE_DRIVER: postgres
          GOOSE_DBSTRING: "postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable"
        run: |
          goose -dir database/migrations up

      - name: Run integration tests
        env:
          DATABASE_URL: postgres://testuser:testpassword@localhost:5433/testdb?sslmode=disable
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

---

## 4. 验收标准

### 4.1 功能验收

- [ ] docker-compose.test.yml 可正常启动 PostgreSQL
- [ ] 数据库健康检查通过（health check）
- [ ] Goose 迁移正向通过（up）
- [ ] Goose 迁移逆向通过（down）
- [ ] 集成测试可在容器中正常运行
- [ ] 清理脚本可正确关闭容器

### 4.2 性能验收

- [ ] Docker 容器启动时间 < 10s
- [ ] 数据库就绪时间 < 15s
- [ ] 集成测试执行时间 < 5 分钟
- [ ] 内存占用 < 500MB

### 4.3 稳定性验收

- [ ] 多次运行测试结果一致
- [ ] 无端口冲突（使用 5433）
- [ ] 容器自动清理，无遗留
- [ ] CI 环境可正常执行

---

## 5. 使用指南

### 5.1 本地开发使用

```bash
# 启动测试数据库
make test-db-up

# 连接到测试数据库
make test-db-psql

# 停止测试数据库
make test-db-down

# 运行完整的集成测试（包括迁移）
make test-db
```

### 5.2 CI 环境使用

```yaml
# GitHub Actions 工作流会自动:
1. 使用 services 启动 PostgreSQL
2. 运行 Goose 迁移
3. 执行集成测试
4. 上传覆盖率报告
```

---

## 6. 交付物清单

- ✅ `docker-compose.test.yml`
- ✅ `scripts/test/init-db.sql`
- ✅ `scripts/run-integration-tests.sh`
- ✅ `Makefile` 更新（test-db 相关目标）
- ✅ `.github/workflows/integration-test.yml` 更新
- ✅ `docs/development-guides/docker-testing-guide.md`
- ✅ 本计划文档（221）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-04
**计划完成日期**: Week 4 Day 2 (Day 16)
