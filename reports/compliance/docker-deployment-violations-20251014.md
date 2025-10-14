# Docker 容器化部署原则违规情况报告

**生成日期**: 2025-10-14
**检查范围**: 项目配置、文档、脚本、环境变量
**检查依据**: CLAUDE.md 第2节、AGENTS.md Docker 强制约束
**检查人**: Claude (AI Assistant)

---

## 执行摘要

本次检查发现项目中存在 **多处严重违反 Docker 容器化部署强制原则**的情况，主要体现在：

1. **宿主机直接运行 Go 服务**：Makefile 和启动脚本默认在宿主机执行 `go run`
2. **文档误导性说明**：README.md 和开发者指南包含宿主机部署的指导
3. **配置文件混乱**：.env 文件优先宿主机连接配置
4. **Docker Compose 配置被边缘化**：服务容器使用 profile 隐藏，默认不启动

**违规等级**: 🔴 **严重** - 与 CLAUDE.md 强制原则直接冲突

---

## 1. 核心违规情况

### 1.1 Makefile - 宿主机运行 Go 服务 🔴

**文件**: `Makefile`
**违规性质**: 默认部署方式违反 Docker 强制原则

#### 违规代码段

```makefile
# Line 21: 说明文档明确提到"本地运行两个 Go 服务"
run-dev          - 启动最小依赖并本地运行两个 Go 服务

# Line 111, 114: 直接在宿主机运行 Go 服务
go run ./cmd/organization-command-service/main.go &
go run ./cmd/organization-query-service/main.go &

# Line 133, 137: run-auth-rs256-sim 目标也是宿主机运行
go run ./cmd/organization-command-service/main.go &
go run ./cmd/organization-query-service/main.go &
```

#### 违反原则
- ❌ **CLAUDE.md 第2节**: "所有服务、数据库、中间件统一通过 Docker Compose 管理，严禁在宿主机直接安装..."
- ❌ **AGENTS.md**: "所有服务通过 `make docker-up` 启动 Docker Compose 容器...严禁在宿主机安装这些服务"

#### 正确做法
应使用 `docker-compose up -d --build graphql-service rest-service` 或等效命令启动容器化服务。

---

### 1.2 启动脚本 - 宿主机部署流程 🔴

**文件**: `scripts/dev-start-simple.sh`
**违规性质**: 完整的宿主机部署流程

#### 违规代码段

```bash
# Line 62: 宿主机数据库连接
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"

# Line 73, 79: 在宿主机运行 Go 服务
go run cmd/organization-command-service/main.go > logs/command-service.log 2>&1 &
go run cmd/organization-query-service/main.go > logs/query-service.log 2>&1 &

# Line 40, 52: 使用宿主机工具直接连接 Docker 容器
psql -h localhost -p 5432 -U user -d cubecastle
redis-cli -h localhost -p 6379 ping
```

#### 问题分析
1. 脚本设计思路完全基于宿主机部署
2. 仅将 PostgreSQL 和 Redis 容器化，Go 服务仍在宿主机运行
3. 与 Docker 强制原则完全背离

---

### 1.3 环境变量配置 - 宿主机优先 🔴

**文件**: `.env`
**违规性质**: 配置注释和优先级误导开发者

#### 违规代码段

```bash
# Line 1-2: 注释明确说明"for host-based Go app - primary"
# Database Configuration (for host-based Go app - primary)
DATABASE_URL=postgresql://user:password@localhost:5432/cubecastle?sslmode=disable

# Line 3-4: Docker 配置被标记为次要
# Database Configuration (for Docker services)
DATABASE_URL_DOCKER=postgresql://user:password@postgres:5432/cubecastle?sslmode=disable

# Line 10-11: Redis 配置使用 localhost
REDIS_HOST=localhost
REDIS_PORT=6379
```

#### 问题分析
1. **主次颠倒**: "host-based Go app - primary" 与 Docker 强制原则矛盾
2. **命名混乱**: `DATABASE_URL` vs `DATABASE_URL_DOCKER` 暗示宿主机为默认模式
3. **Redis 配置**: 直接使用 localhost，未提供容器内部连接配置

#### 正确做法
```bash
# Database Configuration (Docker Compose - 强制使用)
DATABASE_URL=postgresql://user:password@postgres:5432/cubecastle?sslmode=disable

# Redis Configuration (容器内连接)
REDIS_HOST=redis
REDIS_PORT=6379

# 备注: 宿主机工具访问使用端口映射 localhost:5432 -> postgres:5432
```

---

### 1.4 README.md - 误导性部署指导 🔴

**文件**: `README.md`
**违规性质**: 快速开始部分包含宿主机部署指导

#### 违规代码段

```markdown
### 手动启动
​```bash
# 基础设施
docker-compose up -d postgres redis

# 后端服务 (Line 83-84: 宿主机运行)
cd cmd/organization-command-service && go run .
cd cmd/organization-query-service && go run .

# 前端开发
cd frontend && npm install && npm run dev
​```
```

#### 问题分析
1. **"手动启动"部分默认宿主机运行 Go 服务**
2. 仅将数据库和 Redis 容器化，应用服务在宿主机
3. 与页面开头"一键启动 (推荐)"章节的 `make run-dev` 一致，意味着推荐流程也违规

---

### 1.5 开发者快速参考 - 配置示例错误 🟡

**文件**: `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md`
**违规性质**: 示例代码使用 localhost 连接，未说明必须通过 Docker

#### 违规代码段

```markdown
# Line 68-70: DATABASE_URL 示例
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
make db-migrate-all

# Line 96: 同样示例
psql "$DATABASE_URL" -f database/migrations/011_audit_record_id_fix.sql

# Line 175-177: 端口配置参考
PostgreSQL: localhost:5432
Redis: localhost:6379
```

#### 问题分析
- 虽然 localhost:5432 可以通过 Docker 端口映射访问，但文档未明确说明
- 开发者可能误以为需要在宿主机安装 PostgreSQL
- 缺少"这是通过 Docker 端口映射访问容器"的说明

---

## 2. Docker Compose 配置问题

### 2.1 服务定义被 Profile 隐藏 🟡

**文件**: `docker-compose.dev.yml`
**问题**: Go 服务容器定义存在但默认不启动

#### 配置分析

```yaml
# Line 41-61: GraphQL 查询服务定义
graphql-service:
  build: ...
  profiles: ["services"]  # 需要明确指定 profile 才启动

# Line 63-83: REST 命令服务定义
rest-service:
  build: ...
  profiles: ["services"]  # 需要明确指定 profile 才启动
```

#### 问题分析
1. **服务定义完整但被边缘化**: 使用 `profiles: ["services"]` 隐藏服务
2. **启动命令未使用这些服务**: `make docker-up` 仅启动 postgres 和 redis
3. **导致默认流程违反原则**: 开发者自然使用宿主机 `go run` 而非容器

#### 建议修复
```yaml
# 移除 profiles 或改为默认启动
graphql-service:
  # ...
  # profiles: ["services"]  # 移除此行，默认启动

rest-service:
  # ...
  # profiles: ["services"]  # 移除此行，默认启动
```

---

## 3. 违规文件清单

### 3.1 核心配置文件

| 文件 | 违规等级 | 违规类型 |
|------|----------|----------|
| `Makefile` | 🔴 严重 | 默认宿主机运行 Go 服务 |
| `.env` | 🔴 严重 | 宿主机配置优先，命名误导 |
| `docker-compose.dev.yml` | 🟡 中等 | 服务定义被 profile 隐藏 |

### 3.2 脚本文件

| 文件 | 违规等级 | 违规类型 |
|------|----------|----------|
| `scripts/dev-start-simple.sh` | 🔴 严重 | 完整宿主机部署流程 |
| `scripts/dev/cleanup-and-full-e2e.sh` | 🟡 中等 | 可能包含 `go run` |
| `scripts/health-check-unified.sh` | 🟡 中等 | 可能包含 `go run` |
| `scripts/deployment/*.sh` | 🟡 中等 | 需进一步检查 |

### 3.3 文档文件

| 文件 | 违规等级 | 违规类型 |
|------|----------|----------|
| `README.md` | 🔴 严重 | "手动启动"部分宿主机部署指导 |
| `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` | 🟡 中等 | localhost 示例未说明 Docker |
| `docs/reference/03-API-AND-TOOLS-GUIDE.md` | 🟡 中等 | 可能包含 localhost 示例 |

---

## 4. 根本原因分析

### 4.1 历史遗留问题
项目早期采用宿主机部署模式，后期虽添加了 Docker Compose 配置，但未彻底迁移。

### 4.2 便利性优先
开发者倾向于使用 `go run` 的即时编译特性，认为容器构建速度慢。

### 4.3 文档同步滞后
CLAUDE.md 和 AGENTS.md 明确了 Docker 强制原则（2025-10-14 更新），但其他文档和脚本未同步。

---

## 5. 整改建议

### 5.1 紧急修复（P0 - 本周内完成）

#### 1. 修复 Makefile
```makefile
# 替换 run-dev 目标
run-dev:
	@echo "🚀 启动开发环境（Docker 强制）..."
	@echo "🐳 拉起基础设施和应用服务..."
	docker-compose up -d --build postgres redis graphql-service rest-service
	@echo "⏳ 等待服务健康..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  curl -fsS http://localhost:9090/health >/dev/null && echo "  ✅ command-service ok" && break || \
	  (echo "  ⏳ 等待 command-service..." && sleep 2); \
	done || true
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
	  curl -fsS http://localhost:8090/health >/dev/null && echo "  ✅ query-service ok" && break || \
	  (echo "  ⏳ 等待 query-service..." && sleep 2); \
	done || true
	@echo "✅ 服务已就绪"
	@echo "📊 查看日志: docker-compose logs -f graphql-service rest-service"

# 添加热重载开发模式（可选）
run-dev-hot:
	@echo "🔥 启动开发环境（热重载模式）..."
	@echo "⚠️  注意: 使用 Air 或 volumes 挂载实现热重载"
	# 需要修改 Dockerfile 支持 Air
```

#### 2. 修复 .env 文件
```bash
# 重命名并调整注释
# Database Configuration (Docker Compose 强制 - 容器内连接)
DATABASE_URL=postgresql://user:password@postgres:5432/cubecastle?sslmode=disable

# Database Configuration (宿主机工具通过端口映射访问)
# 注意: 仅用于 psql、迁移脚本等宿主机工具，应用服务必须使用上方 DATABASE_URL
DATABASE_URL_HOST_TOOLS=postgresql://user:password@localhost:5432/cubecastle?sslmode=disable

# Redis Configuration (容器内连接)
REDIS_HOST=redis
REDIS_PORT=6379
```

#### 3. 修复 docker-compose.dev.yml
```yaml
# 移除 profiles，默认启动所有服务
graphql-service:
  build:
    context: .
    dockerfile: cmd/organization-query-service-unified/Dockerfile
  container_name: cubecastle-graphql
  # ... 其他配置 ...
  # profiles: ["services"]  # 移除此行

rest-service:
  build:
    context: .
    dockerfile: cmd/organization-command-service/Dockerfile
  container_name: cubecastle-rest
  # ... 其他配置 ...
  # profiles: ["services"]  # 移除此行
```

#### 4. 修复 README.md
```markdown
### 一键启动（强制使用 Docker）
​```bash
# 启动完整开发环境（基础设施 + 应用服务）
make docker-up  # 或 docker-compose up -d --build

# 检查服务状态
make status

# 查看日志
docker-compose logs -f graphql-service rest-service

# 启动前端（仍在宿主机，因需热重载）
make frontend-dev
​```

### 手动启动（不推荐，仅用于调试）
​```bash
# ⚠️ 警告: 违反 Docker 强制原则，仅在特殊调试场景使用
# 基础设施
docker-compose up -d postgres redis

# 后端服务（调试模式，不推荐）
cd cmd/organization-command-service && go run .
cd cmd/organization-query-service && go run .
​```
```

### 5.2 中期优化（P1 - 2周内完成）

#### 1. 废弃 scripts/dev-start-simple.sh
```bash
# 在脚本开头添加警告并引导
#!/bin/bash
echo "⚠️  警告: 此脚本已废弃，违反 Docker 强制部署原则"
echo "请使用: make docker-up 或 docker-compose up -d --build"
echo "详见: CLAUDE.md 第2节、AGENTS.md Docker 强制约束"
echo ""
read -p "是否继续使用已废弃脚本？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi
# ... 原有代码 ...
```

#### 2. 更新开发者快速参考
在 `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 所有 localhost 示例前添加说明：
```markdown
### 数据库初始化（Docker 强制）
⚠️ **重要**: 本项目强制使用 Docker 容器化部署。以下命令中的 `localhost:5432` 是通过 Docker 端口映射访问容器数据库，并非宿主机安装的 PostgreSQL。

​```bash
# 环境变量（宿主机工具通过端口映射访问 Docker 容器）
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
make db-migrate-all
​```
```

#### 3. 添加 CI 检查
在 `.github/workflows/` 添加 Docker 原则守护：
```yaml
name: Docker Deployment Compliance

on: [push, pull_request]

jobs:
  check-docker-compliance:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check for go run in Makefile
        run: |
          if grep -n "go run.*cmd/" Makefile; then
            echo "❌ Makefile 不得包含 'go run' 命令（违反 Docker 强制原则）"
            exit 1
          fi
      - name: Check .env configuration
        run: |
          if grep -i "host-based.*primary" .env; then
            echo "❌ .env 不得优先宿主机配置（违反 Docker 强制原则）"
            exit 1
          fi
```

### 5.3 长期规范（P2 - 1个月内完成）

#### 1. 开发热重载方案
```dockerfile
# 修改 Dockerfile 支持 Air 热重载
FROM golang:1.23-alpine AS dev

WORKDIR /app
RUN go install github.com/cosmtrek/air@latest
COPY . .
CMD ["air", "-c", ".air.toml"]
```

```yaml
# docker-compose.dev.yml 添加 volumes 挂载
rest-service:
  build:
    target: dev  # 开发阶段使用 dev target
  volumes:
    - ./cmd/organization-command-service:/app/cmd/organization-command-service
    - ./internal:/app/internal
  # ...
```

#### 2. 文档全面更新
- 所有文档统一说明 Docker 强制原则
- 删除或标记废弃宿主机部署相关内容
- 在 `docs/reference/` 添加 Docker 最佳实践文档

---

## 6. 验收标准

### 6.1 强制检查点（P0）
- [ ] Makefile `run-dev` 目标使用 `docker-compose up` 而非 `go run`
- [ ] `.env` 文件移除 "host-based ... primary" 注释
- [ ] `docker-compose.dev.yml` 移除 services profile，默认启动所有服务
- [ ] README.md "一键启动"部分仅包含 Docker 命令

### 6.2 质量检查点（P1）
- [ ] `scripts/dev-start-simple.sh` 添加废弃警告
- [ ] `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` 所有 localhost 示例添加 Docker 说明
- [ ] CI 添加 Docker 原则守护工作流
- [ ] 所有 `scripts/` 目录脚本通过 Docker 原则检查

### 6.3 长期目标（P2）
- [ ] 开发环境支持热重载（Air + volumes）
- [ ] 所有文档完成 Docker 强制原则同步
- [ ] 删除或归档所有宿主机部署相关脚本

---

## 7. 参考资料

- **强制原则来源**: `CLAUDE.md` 第2节、第5节
- **执行规范**: `AGENTS.md` Docker 容器化部署强制约束
- **运维案例**: `reports/operations/postgresql-port-cleanup-20251014.md`
- **原则确立时间**: 2025-10-14

---

## 8. 附录：完整违规文件列表

### 配置文件
1. `Makefile` - run-dev, run-auth-rs256-sim 目标
2. `.env` - DATABASE_URL, REDIS_HOST 配置
3. `docker-compose.dev.yml` - profiles 隐藏服务

### 脚本文件
4. `scripts/dev-start-simple.sh` - 完整宿主机部署流程
5. `scripts/dev/cleanup-and-full-e2e.sh` - 可能包含 go run
6. `scripts/generate-rs256-token.sh` - 可能包含 go run
7. `scripts/health-check-unified.sh` - 可能包含 go run
8. `scripts/quick-status.sh` - 可能包含 go run
9. `scripts/deployment/deploy-production.sh` - 需检查
10. `scripts/tests/*.sh` - 多个测试脚本需检查

### 文档文件
11. `README.md` - "手动启动"部分
12. `docs/reference/01-DEVELOPER-QUICK-REFERENCE.md` - localhost 示例
13. `docs/reference/03-API-AND-TOOLS-GUIDE.md` - 可能包含 localhost 示例

---

**报告生成**: 2025-10-14 23:00 CST
**下次复查**: P0 修复完成后重新生成报告
