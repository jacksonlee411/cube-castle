# Plan 221 - Docker 集成测试基座建设

**文档编号**: 221
**标题**: 容器化集成测试环境 - 基础设施建设
**创建日期**: 2025-11-04
**分支**: `feature/205-phase2-infrastructure`
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

- **计划窗口**: Week 3 Day 5 (Day 16) 预拉取镜像 → Week 4 Day 3 (Day 19)，与 `docs/development-plans/215-phase2-summary-overview.md:304-316` 保持一致（W3-D5 预拉取，W4-D1~D3 完成脚本与 CI 配置）
- **计划完成**: Week 4 Day 3 (Day 19)
- **交付周期**: 3 天（含预拉取、脚本成型、CI 集成与验收）
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
      - "5432:5432"
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

> ⚠️ 端口 5432 与开发环境一致。若宿主机存在 PostgreSQL 占用，请按照 `AGENTS.md`/`CLAUDE.md` 要求**卸载宿主服务**释放端口，禁止调整容器端口映射。

### 3.2 初始化脚本 (scripts/test/init-db.sql)

```sql
-- scripts/test/init-db.sql
-- PostgreSQL 初始化脚本，容器启动时自动执行

-- 创建测试数据库所需扩展，其余表结构交由 Goose 迁移维护
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
```

### 3.3 集成测试启动脚本 (scripts/run-integration-tests.sh)

```bash
#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"
MIGRATIONS_DIR="$PROJECT_ROOT/database/migrations"

cleanup() {
  docker compose -f "$COMPOSE_FILE" down -v >/dev/null 2>&1 || true
}

trap cleanup EXIT

docker compose -f "$COMPOSE_FILE" up -d postgres-test

for _ in $(seq 1 40); do
  if docker compose -f "$COMPOSE_FILE" exec -T postgres-test pg_isready -U testuser -d testdb >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

DB_HOST=localhost
DB_PORT=5432
DB_USER=testuser
DB_PASSWORD=testpassword
DB_NAME=testdb
export DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" up
go test -v -tags=integration ./...
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" down
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
name: integration-test

on:
  push:
    branches: [ master, feature/**, plan/** ]
  pull_request:
    branches: [ master ]

jobs:
  docker-integration:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install Goose
        run: go install github.com/pressly/goose/v3/cmd/goose@latest

      - name: Run Docker integration tests
        run: |
          make test-db
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
- [ ] 无端口冲突（宿主机需释放 5432，禁止调整映射）
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

> 每次执行完成后，请在 `docs/development-plans/221t-docker-integration-validation.md` 中记录责任人、时间与日志链接。

### 5.2 CI 环境使用

```yaml
# GitHub Actions workflow 会执行:
1. checkout 仓库并安装 Go/Goose
2. 运行 `make test-db`（脚本会拉起 docker-compose.test.yml、执行 Goose up/down、go test -tags=integration）
3. 通过 trap 清理容器与卷，确保 CI 方便复用
```

---

## 6. 交付物清单

- [x] `docker-compose.test.yml`
- [x] `scripts/test/init-db.sql`
- [x] `scripts/run-integration-tests.sh`
- [x] `Makefile` 更新（test-db 相关目标）
- [x] `.github/workflows/integration-test.yml`
- [x] `docs/development-guides/docker-testing-guide.md`
- [x] `docs/development-plans/221t-docker-integration-validation.md`

---

## 7. 执行状态（2025-11-09）

- `shangmeilin` 在 `plan/221-prep` 分支执行 221T（`logs/plan221/run-20251109145841.log`），`make test-db` → Goose up/down + `go test -v -tags=integration ./...` 全数通过。
- 当前环境无法直接触发 GitHub Actions；待恢复访问后由 DevOps 触发 `integration-test` workflow 以记录对应 run 链接。
- [x] 本计划文档（221）

---

**维护者**: Codex（AI 助手）
**最后更新**: 2025-11-07
**计划完成日期**: Week 4 Day 3 (Day 19, 2025-11-09)
