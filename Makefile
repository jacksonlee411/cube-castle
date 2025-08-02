# Cube Castle Makefile
# 用于简化项目的构建、测试和部署流程
# 🚀 包含 Operation Phoenix - CQRS+CDC 架构支持

.PHONY: help build test clean docker-build docker-up docker-down init-db seed-data run-dev
.PHONY: phoenix-start phoenix-stop phoenix-status phoenix-reset test-cdc monitor connectors

# 默认目标
help:
	@echo "🏰 Cube Castle - 可用命令:"
	@echo ""
	@echo "🚀 Operation Phoenix (CQRS+CDC架构):"
	@echo "  phoenix-start - 启动完整CQRS+CDC架构"
	@echo "  phoenix-stop  - 停止所有Phoenix服务"
	@echo "  phoenix-status- 查看Phoenix服务状态"
	@echo "  phoenix-reset - 完全重置Phoenix环境"
	@echo "  test-cdc      - 测试CDC数据流"
	@echo "  monitor       - 打开监控面板"
	@echo "  connectors    - 查看Debezium连接器状态"
	@echo ""
	@echo "📦 构建相关:"
	@echo "  build         - 构建 Go 应用"
	@echo "  clean         - 清理构建产物"
	@echo "  docker-build  - 构建 Docker 镜像"
	@echo ""
	@echo "🐳 Docker 相关:"
	@echo "  docker-up     - 启动所有 Docker 服务"
	@echo "  docker-down   - 停止所有 Docker 服务"
	@echo "  docker-logs   - 查看 Docker 服务日志"
	@echo ""
	@echo "🗄️ 数据库相关:"
	@echo "  init-db       - 初始化数据库"
	@echo "  seed-data     - 插入种子数据"
	@echo ""
	@echo "🧪 测试相关:"
	@echo "  test          - 运行单元测试"
	@echo "  test-integration - 运行集成测试"
	@echo ""
	@echo "🚀 开发相关:"
	@echo "  run-dev       - 启动开发环境"
	@echo "  install-deps  - 安装依赖"
	@echo "  generate      - 生成代码"

# =============================================================================
# 🚀 Operation Phoenix - CQRS+CDC Architecture Commands
# =============================================================================

phoenix-start: ## 启动Operation Phoenix (完整CQRS+CDC架构)
	@echo "🚀 启动Operation Phoenix..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker未安装"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose未安装"; exit 1; }
	@./scripts/setup-cdc-pipeline.sh

phoenix-stop: ## 停止所有Phoenix服务
	@echo "🛑 停止Operation Phoenix服务..."
	@docker-compose down

phoenix-status: ## 查看Phoenix服务状态
	@echo "📊 Operation Phoenix 服务状态:"
	@echo "================================"
	@docker-compose ps
	@echo ""
	@echo "🔍 关键服务健康检查:"
	@echo "PostgreSQL: $$(docker exec cube_castle_postgres pg_isready -U user -d cubecastle 2>/dev/null && echo '✅ 正常' || echo '❌ 异常')"
	@echo "Neo4j: $$(curl -f http://localhost:7474 >/dev/null 2>&1 && echo '✅ 正常' || echo '❌ 异常')"
	@echo "Kafka Connect: $$(curl -f http://localhost:8083/ >/dev/null 2>&1 && echo '✅ 正常' || echo '❌ 异常')"
	@echo ""
	@echo "🌐 访问地址:"
	@echo "  Kafka UI: http://localhost:8081"
	@echo "  Neo4j Browser: http://localhost:7474"
	@echo "  PgAdmin: http://localhost:5050"

phoenix-reset: ## 完全重置Phoenix环境 (删除所有数据)
	@echo "⚠️  这将删除所有数据！按Ctrl+C取消，或按Enter继续..."
	@read
	@echo "🔄 重置Operation Phoenix环境..."
	@docker-compose down -v
	@docker system prune -f --volumes
	@echo "✅ 环境重置完成"

test-cdc: ## 测试CDC数据流
	@echo "🧪 测试CDC数据流..."
	@echo "插入测试数据..."
	@docker exec cube_castle_postgres psql -U user -d cubecastle -c "\
		INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name, email, hire_date, employment_status) \
		VALUES (gen_random_uuid(), gen_random_uuid(), 'FULL_TIME', 'CDC', 'Test$$(date +%S)', 'cdc.test$$(date +%s)@cubecastle.com', NOW(), 'ACTIVE'); \
		SELECT 'CDC测试数据已插入，Employee: ' || first_name || ' ' || last_name FROM employees WHERE first_name = 'CDC' ORDER BY created_at DESC LIMIT 1;"
	@echo "等待数据同步..."
	@sleep 3
	@echo "检查Kafka主题..."
	@docker exec cube_castle_kafka kafka-topics --list --bootstrap-server localhost:9092 | grep organization || echo "❌ 未找到organization相关主题"

monitor: ## 打开监控面板
	@echo "📊 打开监控面板..."
	@echo "Kafka UI: http://localhost:8081"
	@if command -v open >/dev/null 2>&1; then \
		open http://localhost:8081; \
	elif command -v xdg-open >/dev/null 2>&1; then \
		xdg-open http://localhost:8081; \
	else \
		echo "请手动访问 http://localhost:8081"; \
	fi

connectors: ## 查看Debezium连接器状态
	@echo "🔌 Debezium连接器状态:"
	@echo "========================"
	@curl -s http://localhost:8083/connectors 2>/dev/null | jq . || echo "❌ 无法连接到Kafka Connect"
	@echo ""
	@echo "连接器详细状态:"
	@curl -s http://localhost:8083/connectors/organization-postgres-connector/status 2>/dev/null | jq . || echo "❌ 连接器未配置"

# =============================================================================
# 原有命令保持不变
# =============================================================================

# 构建 Go 应用
build:
	@echo "🔨 构建 Go 应用..."
	cd go-app && go build -o bin/server cmd/server/main.go

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	rm -rf go-app/bin
	rm -rf go-app/generated
	find . -name "*.exe" -delete
	find . -name "*.test" -delete

# 构建 Docker 镜像
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t cube-castle:latest .

# 启动 Docker 服务
docker-up:
	@echo "🚀 启动 Docker 服务..."
	docker-compose up -d

# 停止 Docker 服务
docker-down:
	@echo "🛑 停止 Docker 服务..."
	docker-compose down

# 查看 Docker 日志
docker-logs:
	@echo "📋 查看 Docker 日志..."
	docker-compose logs -f

# 初始化数据库
init-db:
	@echo "🗄️ 初始化数据库..."
	cd go-app && go run cmd/server/main.go init-db

# 插入种子数据
seed-data:
	@echo "🌱 插入种子数据..."
	cd go-app && go run cmd/server/main.go seed-data

# 运行单元测试
test:
	@echo "🧪 运行单元测试..."
	cd go-app && go test -v ./...

# 运行集成测试
test-integration:
	@echo "🔗 运行集成测试..."
	cd go-app && go test -v -tags=integration ./...

# 启动开发环境
run-dev:
	@echo "🚀 启动开发环境..."
	@echo "1. 启动基础设施..."
	docker-compose up -d postgres neo4j
	@echo "2. 等待服务启动..."
	sleep 10
	@echo "3. 初始化数据库..."
	$(MAKE) init-db
	@echo "4. 插入种子数据..."
	$(MAKE) seed-data
	@echo "5. 启动 Python AI 服务..."
	cd python-ai && python main.py &
	@echo "6. 启动 Go 主服务..."
	cd go-app && go run cmd/server/main.go

# 安装依赖
install-deps:
	@echo "📦 安装依赖..."
	# 安装 Go 依赖
	cd go-app && go mod download
	# 安装 Python 依赖
	cd python-ai && pip install -r requirements.txt

# 生成代码
generate:
	@echo "🔧 生成代码..."
	# 生成 OpenAPI 代码
	cd go-app && oapi-codegen -package openapi ../contracts/openapi.yaml > generated/openapi/server.go
	# 生成 gRPC 代码
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		contracts/proto/intelligence.proto

# 格式化代码
fmt:
	@echo "🎨 格式化代码..."
	cd go-app && go fmt ./...
	cd python-ai && black .

# 代码检查
lint:
	@echo "🔍 代码检查..."
	cd go-app && golangci-lint run
	cd python-ai && flake8 .

# 安全扫描
security:
	@echo "🔒 安全扫描..."
	cd go-app && gosec ./...
	cd python-ai && bandit -r .

# 性能测试
bench:
	@echo "⚡ 性能测试..."
	cd go-app && go test -bench=. ./...

# 覆盖率测试
coverage:
	@echo "📊 覆盖率测试..."
	cd go-app && go test -coverprofile=coverage.out ./...
	cd go-app && go tool cover -html=coverage.out -o coverage.html

# 备份数据库
backup:
	@echo "💾 备份数据库..."
	docker exec cube_castle_postgres pg_dump -U user cubecastle > backup_$(shell date +%Y%m%d_%H%M%S).sql

# 恢复数据库
restore:
	@echo "📥 恢复数据库..."
	docker exec -i cube_castle_postgres psql -U user cubecastle < $(BACKUP_FILE)

# 查看服务状态
status:
	@echo "📊 服务状态:"
	docker-compose ps
	@echo ""
	@echo "🔗 服务地址:"
	@echo "  - Go 主服务: http://localhost:8080"
	@echo "  - Python AI 服务: localhost:50051 (gRPC)"
	@echo "  - PostgreSQL: localhost:5432"
	@echo "  - Neo4j: http://localhost:7474"

# 完整重置
reset:
	@echo "🔄 完整重置..."
	$(MAKE) docker-down
	$(MAKE) clean
	docker volume rm cube-castle_postgres_data cube-castle_neo4j_data 2>/dev/null || true
	$(MAKE) docker-up
	sleep 15
	$(MAKE) init-db
	$(MAKE) seed-data 