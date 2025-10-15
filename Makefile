# Cube Castle Makefile (PostgreSQL 原生)
## 目的：提供最小可用的本地开发/构建/测试命令，彻底移除 Neo4j/Kafka/CDC(Phoenix) 相关内容

.PHONY: help build clean docker-build docker-up docker-down docker-logs run-dev frontend-dev test test-integration fmt lint security bench coverage backup restore status reset jwt-dev-mint jwt-dev-info jwt-dev-export jwt-dev-setup db-migrate-all dev-kill run-auth-rs256-sim auth-flow-test test-e2e-auth test-auth-unit e2e-full temporal-validate

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
	@echo "  dev-kill         - 结束占用 9090/8090 的本地服务进程"
	@echo "  frontend-dev     - 启动前端开发服务器 (vite)"
	@echo ""
	@echo "🔑 开发JWT:"
	@echo "  jwt-dev-mint    - 生成开发用JWT并保存到 ./.cache/dev.jwt"
	@echo "  jwt-dev-info    - 查询当前开发JWT信息"
	@echo "  jwt-dev-export  - 导出环境变量 JWT_TOKEN（从 ./.cache/dev.jwt）"
	@echo "  jwt-dev-setup   - 生成本地RS256密钥对（可选）"
	@echo ""
	@echo "🧪 质量:"
	@echo "  test             - 运行 Go 单元测试"
	@echo "  test-integration - 运行 Go 集成测试 (-tags=integration)"
	@echo "  test-auth-unit   - 运行 RS256+JWKS 认证单元测试（查询服务中间件）"
	@echo "  test-e2e-auth    - 运行 认证端到端测试（需要 Postgres/Redis 运行中）"
	@echo "  e2e-full         - 清理→重启（RS256+JWKS）→前端E2E（webServer自启）"
	@echo "  temporal-validate- 校验时态工具引用与规则（前端统一入口）"
	@echo "  fmt              - Go 代码格式化"
	@echo "  lint             - golangci-lint 检查"
	@echo "  security         - gosec 安全扫描"
	@echo "  bench            - Go 基准测试"
	@echo "  coverage         - 生成覆盖率报告 (coverage.html)"
	@echo ""
	@echo "🗄️ 数据库维护:"
	@echo "  backup           - 备份 PostgreSQL 数据到文件"
	@echo "  restore          - 从备份文件恢复 (需 BACKUP_FILE)"
	@echo "  db-migrate-all   - 按序执行数据库迁移（迁移即真源）"
	@echo ""
	@echo "📊 运行状态:"
	@echo "  status           - docker compose 服务状态 + 关键地址"
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
	docker compose -f docker-compose.dev.yml up -d postgres redis

docker-down:
	@echo "🛑 停止并清理开发容器..."
	docker compose -f docker-compose.dev.yml down

docker-logs:
	@echo "📋 查看最小依赖日志... (Ctrl+C 退出)"
	docker compose -f docker-compose.dev.yml logs -f postgres redis

# 启动本地开发（Docker 强制）
run-dev:
	@echo "🚀 启动开发环境（Docker 强制原则）..."
	@echo "🔐 检查 RS256 密钥..."
	@$(MAKE) jwt-dev-setup >/dev/null
	@echo "🐳 拉起基础设施与应用服务 (postgres, redis, rest-service, graphql-service)..."
	docker compose -f docker-compose.dev.yml up -d --build postgres redis rest-service graphql-service
	@echo "⏳ 等待服务健康..."
	-@SUCCESS=0; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
	  if curl -fsS http://localhost:9090/health >/dev/null; then \
	    echo "  ✅ command-service 就绪 (http://localhost:9090/health)"; \
	    SUCCESS=1; \
	    break; \
	  fi; \
	  echo "  ⏳ 等待 command-service..."; \
	  sleep 2; \
	done; \
	if [ $$SUCCESS -ne 1 ]; then \
	  echo "  ⚠️  command-service 未就绪，查看日志: docker compose -f docker-compose.dev.yml logs -f rest-service"; \
	fi
	-@SUCCESS=0; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
	  if curl -fsS http://localhost:8090/health >/dev/null; then \
	    echo "  ✅ graphql-service 就绪 (http://localhost:8090/health)"; \
	    SUCCESS=1; \
	    break; \
	  fi; \
	  echo "  ⏳ 等待 graphql-service..."; \
	  sleep 2; \
	done; \
	if [ $$SUCCESS -ne 1 ]; then \
	  echo "  ⚠️  graphql-service 未就绪，查看日志: docker compose -f docker-compose.dev.yml logs -f graphql-service"; \
	fi
	@echo ""
	@echo "📊 查看日志: docker compose -f docker-compose.dev.yml logs -f rest-service graphql-service"
	@echo "🛑 停止服务: docker compose -f docker-compose.dev.yml down 或 make docker-down"
	@echo "ℹ️  若需宿主机调试，请使用: make run-dev-debug"

# 调试模式：宿主机运行 Go 服务（仅限特殊场景）
run-dev-debug:
	@echo "⚠️  警告: 调试模式违反 Docker 强制原则，仅限特殊调试场景使用"
	@echo "    日常开发请使用: make run-dev"
	@read -p "确认继续？(y/N) " -n 1 -r REPLY; echo ""; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then echo "已取消"; exit 1; fi
	@echo "🧹 清理端口占用 (9090/8090)..."
	-@PIDS=$$(lsof -t -i :9090 -sTCP:LISTEN 2>/dev/null || true); \
	if [ -n "$$PIDS" ]; then \
	  echo "  🔪 kill $$PIDS (9090)"; kill $$PIDS || true; sleep 1; \
	  PIDS2=$$(lsof -t -i :9090 -sTCP:LISTEN 2>/dev/null || true); \
	  if [ -n "$$PIDS2" ]; then echo "  🔪 kill -9 $$PIDS2 (9090)"; kill -9 $$PIDS2 || true; sleep 1; fi; \
	fi
	-@PIDS=$$(lsof -t -i :8090 -sTCP:LISTEN 2>/dev/null || true); \
	if [ -n "$$PIDS" ]; then \
	  echo "  🔪 kill $$PIDS (8090)"; kill $$PIDS || true; sleep 1; \
	  PIDS2=$$(lsof -t -i :8090 -sTCP:LISTEN 2>/dev/null || true); \
	  if [ -n "$$PIDS2" ]; then echo "  🔪 kill -9 $$PIDS2 (8090)"; kill -9 $$PIDS2 || true; sleep 1; fi; \
	fi
	$(MAKE) jwt-dev-setup
	$(MAKE) docker-up
	@echo "⏳ 等待依赖健康..."
	@sleep 5
	@echo "▶ 启动命令服务 (9090)..."
	JWT_ALG=RS256 JWT_MINT_ALG=RS256 JWT_PRIVATE_KEY_PATH=$(CURDIR)/secrets/dev-jwt-private.pem JWT_PUBLIC_KEY_PATH=$(CURDIR)/secrets/dev-jwt-public.pem JWT_KEY_ID=bff-key-1 \
		go run ./cmd/organization-command-service/main.go &
	@echo "▶ 启动查询服务 (8090)..."
	JWT_ALG=RS256 JWT_JWKS_URL=http://localhost:9090/.well-known/jwks.json \
		go run ./cmd/organization-query-service/main.go &
	@echo "🩺 健康检查 (若服务已实现 /health)："
	-@for i in 1 2 3 4 5 6 7 8 9 10; do curl -fsS http://localhost:9090/health >/dev/null && echo "  ✅ command-service ok" && break || (echo "  ⏳ 等待 command-service..." && sleep 1); done || true
	-@for i in 1 2 3 4 5 6 7 8 9 10; do curl -fsS http://localhost:8090/health >/dev/null && echo "  ✅ query-service ok" && break || (echo "  ⏳ 等待 query-service..." && sleep 1); done || true

# 启动 RS256+JWKS 本地联调（命令服务 RS256 mint + OIDC 模拟；查询服务用 JWKS 验签）
run-auth-rs256-sim:
	@echo "ℹ️  run-auth-rs256-sim 已合并至 make run-dev（容器化）"
	@echo "👉 运行: make run-dev"

# 认证联调脚本（自动执行登录→会话→GraphQL 调用）
auth-flow-test:
	@bash scripts/auth_flow_test.sh

# 认证相关测试
test-auth-unit:
	@echo "🧪 运行 RS256+JWKS 认证单元测试（查询服务中间件）..."
	cd cmd/organization-query-service && go test ./internal/auth -run TestRS256JWTValidationWithJWKS -v

test-e2e-auth:
	@echo "🧪 运行 认证端到端测试...（需要 Postgres/Redis 已运行）"
	E2E_RUN=1 go test ./tests/e2e -v

e2e-full:
	@echo "🧪 清理→重启（RS256+JWKS）→前端E2E（webServer自启）"
	bash scripts/dev/cleanup-and-full-e2e.sh

temporal-validate:
	@echo "🕒 校验前端时态工具引用..."
	cd frontend && npm run validate:temporal

dev-kill:
	@echo "🧹 结束本地开发服务进程 (9090/8090) ..."
	-@PIDS=$$(lsof -t -i :9090 -sTCP:LISTEN 2>/dev/null || true); if [ -n "$$PIDS" ]; then echo "  🔪 kill $$PIDS (9090)"; kill $$PIDS || true; else echo "  ✅ 9090 空闲"; fi
	-@PIDS=$$(lsof -t -i :8090 -sTCP:LISTEN 2>/dev/null || true); if [ -n "$$PIDS" ]; then echo "  🔪 kill $$PIDS (8090)"; kill $$PIDS || true; else echo "  ✅ 8090 空闲"; fi

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
	docker exec cubecastle-postgres pg_dump -U $${POSTGRES_USER:-user} $${POSTGRES_DB:-cubecastle} > backup_$$(date +%Y%m%d_%H%M%S).sql

restore:
	@echo "📥 恢复数据库..."
	@test -n "$(BACKUP_FILE)" || (echo "❌ 需要指定 BACKUP_FILE=/path/to/file.sql" && exit 2)
	@command -v docker >/dev/null 2>&1 || { echo "❌ 需要 docker"; exit 1; }
	docker exec -i cubecastle-postgres psql -U $${POSTGRES_USER:-user} $${POSTGRES_DB:-cubecastle} < $(BACKUP_FILE)

# 状态与重置
status:
	@echo "📊 docker compose 服务状态:"
	docker compose -f docker-compose.dev.yml ps
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

# 迁移即真源：按序执行 database/migrations/*.sql
db-migrate-all:
	@echo "🧭 执行数据库迁移（迁移即真源）..."
	@command -v psql >/dev/null 2>&1 || { echo "❌ 需要安装 psql (PostgreSQL 客户端)"; exit 1; }
	@DB_URL="$$DATABASE_URL" ; \
	if [ -z "$$DB_URL" ]; then \
	  DB_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable" ; \
	  echo "ℹ️  未设置 DATABASE_URL，使用默认: $$DB_URL" ; \
	fi ; \
	set -e ; \
	for f in $$(ls -1 database/migrations/*.sql | sort); do \
	  echo "▶ 迁移: $$f" ; \
	  psql "$$DB_URL" -v ON_ERROR_STOP=1 -f "$$f" ; \
	done ; \
	echo "✅ 迁移完成"


# 开发JWT工具
jwt-dev-mint:
	@echo "🔑 生成开发JWT..."
	@scripts/dev/mint-dev-jwt.sh \
	  $(if $(USER_ID),--user-id $(USER_ID),) \
	  $(if $(TENANT_ID),--tenant-id $(TENANT_ID),) \
	  $(if $(ROLES),--roles $(ROLES),) \
	  $(if $(DURATION),--duration $(DURATION),)

jwt-dev-info:
	@echo "🔎 查询开发JWT信息..."
	@test -f ./.cache/dev.jwt || { echo "❌ 未找到 ./.cache/dev.jwt，请先执行: make jwt-dev-mint"; exit 2; }
	@TOKEN=$$(cat ./.cache/dev.jwt) ; \
	curl -s -H "Authorization: Bearer $$TOKEN" http://localhost:9090/auth/dev-token/info | (command -v jq >/dev/null 2>&1 && jq . || cat)

jwt-dev-export:
	@echo "🌱 导出 JWT_TOKEN 环境变量 (当前进程无效，供 shell 评估)"
	@test -f ./.cache/dev.jwt || { echo "❌ 未找到 ./.cache/dev.jwt，请先执行: make jwt-dev-mint"; exit 2; }
	@echo "export JWT_TOKEN=$$(cat ./.cache/dev.jwt)"

jwt-dev-setup:
	@mkdir -p secrets
	@if [ -f secrets/dev-jwt-private.pem ] && [ -f secrets/dev-jwt-public.pem ]; then \
	  echo "🔐 检测到已存在的 RS256 密钥对，跳过生成 (secrets/dev-jwt-*.pem)"; \
	else \
	  echo "🔐 生成本地RS256开发密钥对..."; \
	  openssl genrsa -out secrets/dev-jwt-private.pem 2048 2>/dev/null && \
	  openssl rsa -in secrets/dev-jwt-private.pem -pubout -out secrets/dev-jwt-public.pem 2>/dev/null && \
	  echo "✅ 已生成 secrets/dev-jwt-private.pem 与 secrets/dev-jwt-public.pem"; \
	fi
