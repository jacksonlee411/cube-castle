# Cube Castle Makefile (PostgreSQL 原生)
## 目的：提供最小可用的本地开发/构建/测试命令，彻底移除 Neo4j/Kafka/CDC(Phoenix) 相关内容

.PHONY: help build clean docker-build docker-up docker-down docker-logs run-dev frontend-dev test test-integration fmt lint security bench coverage backup restore status reset monitoring-up monitoring-down monitoring-test

# 默认目标
help:
	@echo "🏰 Cube Castle - PostgreSQL 原生命令:"
	@echo ""
	@echo "📦 构建:"
	@echo "  build            - 构建 command/query 两个 Go 服务二进制到 bin/"
	@echo "  clean            - 清理构建产物与临时文件"
	@echo "  docker-build     - 构建通用 Docker 镜像（如需要）"
	@echo ""
	@echo "🐳 基础设施:"
	@echo "  docker-up        - 启动最小依赖 (postgres, redis)"
	@echo "  docker-down      - 停止最小依赖 (postgres, redis)"
	@echo "  docker-logs      - 查看最小依赖日志"
	@echo ""
	@echo "🚀 开发运行:"
	@echo "  run-dev          - 启动最小依赖并本地运行两个 Go 服务"
	@echo "  frontend-dev     - 启动前端开发服务器 (vite)"
	@echo "  monitoring-up    - 启动监控栈 (Prometheus/Grafana/AlertManager)"
	@echo "  monitoring-test  - 验证监控栈运行状况与指标"
	@echo "  monitoring-down  - 停止监控栈"
	@echo ""
	@echo "🧪 质量:"
	@echo "  test             - 运行 Go 单元测试"
	@echo "  test-integration - 运行 Go 集成测试 (-tags=integration)"
	@echo "  fmt              - Go 代码格式化"
	@echo "  lint             - golangci-lint 检查"
	@echo "  security         - gosec 安全扫描"
	@echo "  bench            - Go 基准测试"
	@echo "  coverage         - 生成覆盖率报告 (coverage.html)"
	@echo ""
	@echo "🗄️ 数据库维护:"
	@echo "  backup           - 备份 PostgreSQL 数据到文件"
	@echo "  restore          - 从备份文件恢复 (需 BACKUP_FILE)"
	@echo ""
	@echo "📊 运行状态:"
	@echo "  status           - docker-compose 服务状态 + 关键地址"
	@echo "  reset            - 清理并重新拉起最小依赖（不删除卷）"

# 构建 Go 应用（PostgreSQL 原生：两个服务）
build:
	@echo "🔨 构建 Go 应用..."
	mkdir -p bin
	go build -o bin/organization-command-service ./cmd/organization-command-service
	go build -o bin/organization-query-service   ./cmd/organization-query-service

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	rm -rf bin
	find . -name "*.exe" -delete
	find . -name "*.test" -delete
	rm -f coverage.out coverage.html

# 构建 Docker 镜像（如需将当前仓库打成通用镜像）
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t cube-castle:latest .

# 最小依赖（PostgreSQL + Redis）
docker-up:
	@echo "🚀 启动最小依赖 (postgres, redis)..."
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ 需要 docker-compose"; exit 1; }
	docker-compose up -d postgres redis

docker-down:
	@echo "🛑 停止最小依赖 (postgres, redis)..."
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ 需要 docker-compose"; exit 1; }
	docker-compose stop postgres redis

docker-logs:
	@echo "📋 查看最小依赖日志... (Ctrl+C 退出)"
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ 需要 docker-compose"; exit 1; }
	docker-compose logs -f postgres redis

# 启动本地开发（两个 Go 服务 + 最小依赖）
run-dev:
	@echo "🚀 启动本地开发环境 (PostgreSQL 原生)..."
	$(MAKE) docker-up
	@echo "⏳ 等待依赖健康..."
	@sleep 5
	@echo "▶ 启动命令服务 (9090)..."
	cd cmd/organization-command-service && go run main.go &
	@echo "▶ 启动查询服务 (8090)..."
	cd cmd/organization-query-service && go run main.go &
	@echo "🩺 健康检查 (若服务已实现 /health)："
	-@curl -fsS http://localhost:9090/health >/dev/null && echo "  ✅ command-service ok" || echo "  ⚠️  command-service 未响应"
	-@curl -fsS http://localhost:8090/health >/dev/null && echo "  ✅ query-service ok" || echo "  ⚠️  query-service 未响应"

# 前端开发
frontend-dev:
	@echo "🎨 启动前端开发服务器..."
	cd frontend && npm run dev

# 质量相关
test:
	@echo "🧪 运行 Go 单元测试..."
	go test -v ./...

test-integration:
	@echo "🔗 运行 Go 集成测试..."
	go test -v -tags=integration ./...

fmt:
	@echo "🎨 Go 代码格式化..."
	go fmt ./...

lint:
	@echo "🔍 golangci-lint 检查..."
	golangci-lint run

security:
	@echo "🔒 gosec 安全扫描..."
	gosec ./...

bench:
	@echo "⚡ Go 基准测试..."
	go test -bench=. ./...

coverage:
	@echo "📊 覆盖率测试..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📄 生成 coverage.html"

# 数据库维护
backup:
	@echo "💾 备份数据库..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ 需要 docker"; exit 1; }
	docker exec cube_castle_postgres pg_dump -U $${POSTGRES_USER:-user} $${POSTGRES_DB:-cubecastle} > backup_$$(date +%Y%m%d_%H%M%S).sql

restore:
	@echo "📥 恢复数据库..."
	@test -n "$(BACKUP_FILE)" || (echo "❌ 需要指定 BACKUP_FILE=/path/to/file.sql" && exit 2)
	@command -v docker >/dev/null 2>&1 || { echo "❌ 需要 docker"; exit 1; }
	docker exec -i cube_castle_postgres psql -U $${POSTGRES_USER:-user} $${POSTGRES_DB:-cubecastle} < $(BACKUP_FILE)

# 状态与重置
status:
	@echo "📊 docker-compose 服务状态:"
	docker-compose ps
	@echo ""
	@echo "🔗 关键地址:"
	@echo "  - Command Service:   http://localhost:9090"
	@echo "  - Query (GraphQL):   http://localhost:8090  (GraphiQL: /graphiql)"
	@echo "  - PostgreSQL:        localhost:5432"
	@echo "  - Redis:             localhost:6379"

reset:
	@echo "🔄 重置最小依赖 (不删除卷)..."
	$(MAKE) docker-down
	$(MAKE) docker-up

# 监控栈
monitoring-up:
	@echo "📈 启动监控栈..."
	./scripts/start-monitoring.sh

monitoring-test:
	@echo "🧪 验证监控栈运行状况..."
	./scripts/test-monitoring.sh

monitoring-down:
	@echo "🛑 停止监控栈..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ 需要 docker"; exit 1; }
	docker compose -f monitoring/docker-compose.monitoring.yml down
