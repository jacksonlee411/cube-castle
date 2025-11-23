# Cube Castle Makefile (PostgreSQL 原生)
## 目的：提供最小可用的本地开发/构建/测试命令，彻底移除 Neo4j/Kafka/CDC(Phoenix) 相关内容

.PHONY: help build clean docker-build docker-up docker-down docker-logs run-dev frontend-dev test test-integration fmt lint workflow-lint security bench coverage backup restore status reset jwt-dev-mint jwt-dev-info jwt-dev-export jwt-dev-setup db-migrate-all db-rollback-last dev-kill run-auth-rs256-sim auth-flow-test test-e2e-auth test-auth-unit e2e-full temporal-validate test-db test-db-up test-db-down test-db-logs test-db-psql protect-branch archive-run-artifacts sqlc-generate
.PHONY: clean-root-logs clean-untracked-binaries guard-plan253 plan253-coldstart
.PHONY: generate-contracts verify-contracts
.PHONY: guard-plan258
.PHONY: guard-plan257
.PHONY: guard-plan259
.PHONY: watch-plan258

export SCHEDULER_ENABLED ?= false
export SCHEDULER_MONITOR_ENABLED ?= true

# 默认目标
help:
	@echo "🏰 Cube Castle - PostgreSQL 原生命令:"
	@echo ""
	@echo "🔒 仓库保护:"
	@echo "  protect-branch    - 配置默认分支保护（仅允许 squash-merge，必需检查，禁止直推；需 gh 登录）"
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
	@echo "  workflow-lint    - 使用 actionlint 校验 .github/workflows（产出 reports/workflows/*）"
	@echo "  security         - gosec 安全扫描"
	@echo "  bench            - Go 基准测试"
	@echo "  coverage         - 生成覆盖率报告 (coverage.html)"
	@echo ""
	@echo "🗄️ 数据库维护:"
	@echo "  backup           - 备份 PostgreSQL 数据到文件"
	@echo "  restore          - 从备份文件恢复 (需 BACKUP_FILE)"
	@echo "  db-migrate-all   - 使用 Goose 执行数据库迁移（迁移即真源）"
	@echo "  db-rollback-last - 使用 Goose 回滚最近一条迁移"
	@echo "  sqlc-generate    - 运行 sqlc 并将输出写入 logs/plan402/schema/*.log"
	@echo ""
	@echo "📊 运行状态:"
	@echo "  status           - docker compose 服务状态 + 关键地址"
	@echo "  reset            - 清理并重新拉起最小依赖（不删除卷）"
	@echo ""
	@echo "🛡️ 门禁（Plan 253）:"
	@echo "  guard-plan253     - 运行 compose 端口/镜像标签门禁（不需要 Docker）"
	@echo "  plan253-coldstart - 记录冷启动与数据库就绪时间（需要 Docker/Compose）"
	@echo ""
	@echo "🛡️ 门禁（Plan 257）:"
	@echo "  guard-plan257     - 前端领域 Facade 覆盖率门禁（报告/可阻断；产出 reports/facade/coverage.json）"
	@echo ""
	@echo "🛡️ 门禁（Plan 258）:"
	@echo "  guard-plan258     - 契约漂移字段矩阵门禁（OpenAPI↔GraphQL，阻断；产出报告）"
	@echo ""
	@echo "🛡️ 门禁（Plan 259A）:"
	@echo "  guard-plan259     - 协议重复矩阵 + 白名单固化（报告/可阻断；产出 reports/plan259/**）"
	@echo "  watch-plan258     - 监听 plan-258-gates 最近一次运行（master），直到完成并打印结果"
	@echo ""
	@echo "📋 契约（Plan 256）:"
	@echo "  generate-contracts - 从契约生成 Go/TS 类型并记录日志 (logs/plan256)"
	@echo "  verify-contracts   - 校验契约快照（OpenAPI/GraphQL/中间层/生成物）"
	@echo ""
	@echo "📮 PR 自动化（Plan 255）:"
	@echo "  pr-255-soft-gate  - 推送当前/指定分支并创建 PR（需 GITHUB_TOKEN/gh 或在 secrets/.env.local 配置）"

# 构建 Go 应用（PostgreSQL 原生：两个服务）
build:
	@echo "🔨 构建 Go 应用..."
	mkdir -p bin
	go build -o bin/organization-command-service ./cmd/hrms-server/command
	go build -o bin/organization-query-service   ./cmd/hrms-server/query

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
	rm -rf bin
	find . -name "*.exe" -delete
	find . -name "*.test" -delete
	rm -f coverage.out coverage.html
	@echo "🧹 可选: make clean-root-logs 可整理根目录日志输出到 logs/"

clean-root-logs:
	@echo "🧹 整理根目录散落日志到 logs/ ..."
	@ts=$$(date +%Y%m%d_%H%M%S); mkdir -p logs/root-archive-$$ts; \
	files="run-*.log run-dev*.log run-frontend*.log run-query*.log run-auth-*.log frontend-dev.log frontend_dev.log orphaned-processes.log all-services-started.log backend-started.log baseline-ports.log baseline-processes.log"; \
	moved=0; \
	for p in $$files; do \
	  for f in $$p; do \
	    if [ -e "$$f" ]; then \
	      echo "  ↪ $$f -> logs/root-archive-$$ts/"; \
	      mv -f "$$f" "logs/root-archive-$$ts/" || true; \
	      moved=1; \
	    fi; \
	  done; \
	done; \
	if [ $$moved -eq 0 ]; then echo "  ✅ 无需整理"; fi

clean-untracked-binaries:
	@echo "🧹 清理根目录未跟踪的二进制 (organization-*)..."
	@set -e; removed=0; \
	for f in organization-command-service organization-query-service; do \
	  if [ -f "$$f" ]; then rm -f "$$f"; echo "  ✂ $$f"; removed=1; fi; \
	done; \
	if [ "$$removed" = "0" ]; then echo "  ✅ 未发现可清理的二进制"; fi

archive-run-artifacts:
	@echo "📦 Plan 272 运行产物归档..."
	@bash scripts/plan272/archive-run-artifacts.sh

# ======================
# PR 自动化（Plan 255）
# ======================
# 可配置变量（可在命令行覆盖，如 make pr-255-soft-gate PR_HEAD=my-branch）
PR_TITLE ?= refactor(health-alerting): migrate JSON tags to camelCase and harden Plan 255 gates
PR_BODY ?= docs/development-plans/255-soft-gate-PR.md
PR_BASE ?= master
PR_HEAD ?= $(shell git rev-parse --abbrev-ref HEAD)

pr-255-soft-gate:
	@echo "📮 Auto PR: $(PR_BASE) <- $(PR_HEAD)"
	@bash scripts/ci/auto-pr.sh --title "$(PR_TITLE)" --body-file "$(PR_BODY)" --base "$(PR_BASE)" --head "$(PR_HEAD)"

# 领域 Facade 覆盖率门禁（Plan 257）
.PHONY: guard-plan257
guard-plan257:
	@echo "🛡️  Plan 257 — 前端领域 Facade 覆盖率门禁"
	@TH=$${THRESHOLD:-0.0}; echo "ℹ️  使用阈值: $$TH（可通过 THRESHOLD 环境变量覆盖）"
	@node scripts/facade/coverage-scan.js --threshold=$${THRESHOLD:-0.0}
	@echo "📄 报告: reports/facade/coverage.json"
	@echo "🗂️  证据: logs/plan257/coverage-*.json"
# 构建 Docker 镜像（如需将当前仓库打成通用镜像）
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker build -t cube-castle:latest .

protect-branch:
	@echo "🔒 配置 GitHub 默认分支保护（需要 gh 已登录且具备仓库维护权限）..."
	@bash scripts/ops/configure-branch-protection.sh

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

# 集成测试数据库（Plan 221）
test-db:
	@echo "🧪 运行 Docker 集成测试（占用 5432，确保宿主机无 PostgreSQL 服务）..."
	@chmod +x scripts/run-integration-tests.sh
	@scripts/run-integration-tests.sh

test-db-up:
	@echo "🚀 启动测试数据库 (postgres-test, 5432)..."
	docker compose -f docker-compose.test.yml up -d postgres-test

test-db-down:
	@echo "🧹 停止测试数据库..."
	docker compose -f docker-compose.test.yml down -v

test-db-logs:
	@echo "📋 查看测试数据库日志... (Ctrl+C 退出)"
	docker compose -f docker-compose.test.yml logs -f postgres-test

test-db-psql:
	@echo "🐚 连接测试数据库 (psql)..."
	docker compose -f docker-compose.test.yml exec postgres-test psql -U testuser -d testdb

# 启动本地开发（Docker 强制）
run-dev:
	@echo "🚀 启动开发环境（Docker 强制原则）..."
	@echo "🔐 检查 RS256 密钥..."
	@$(MAKE) jwt-dev-setup >/dev/null
	@echo "🐳 拉起最小依赖 (postgres, redis)..."
	docker compose -f docker-compose.dev.yml up -d postgres redis
	@echo "⏳ 等待数据库/缓存就绪..."
	-@SUCCESS=0; \
	for i in 1 2 3 4 5 6 7 8 9 10; do \
	  if docker compose -f docker-compose.dev.yml ps | grep -E 'cubecastle-postgres.*(healthy)' >/dev/null 2>&1; then \
	    echo "  ✅ postgres 就绪"; SUCCESS=1; break; \
	  fi; \
	  echo "  ⏳ 等待 postgres..."; sleep 2; \
	done; \
	if [ $$SUCCESS -ne 1 ]; then \
	  echo "  ⚠️  postgres 未就绪，查看日志: docker compose -f docker-compose.dev.yml logs -f postgres"; \
	fi
	@echo "🧭 执行数据库迁移（Goose）..."
	@$(MAKE) db-migrate-all
	@echo "🐳 拉起应用服务 (rest-service, graphql-service)..."
	docker compose -f docker-compose.dev.yml up -d --build rest-service
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
	@echo ""
	@echo "📊 查看日志: docker compose -f docker-compose.dev.yml logs -f rest-service"
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
		go run ./cmd/hrms-server/command/main.go &
	@echo "▶ 启动查询服务 (8090)..."
	JWT_ALG=RS256 JWT_JWKS_URL=http://localhost:9090/.well-known/jwks.json \
		go run ./cmd/hrms-server/query/main.go &
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
	cd cmd/hrms-server/query && go test ./internal/auth -run TestRS256JWTValidationWithJWKS -v

test-e2e-auth:
	@echo "🧪 运行 认证端到端测试...（需要 Postgres/Redis 已运行）"
	E2E_RUN=1 go test ./tests/e2e -v

e2e-full:
	@echo "🧪 清理→重启（RS256+JWKS）→前端E2E（webServer自启）"
	bash scripts/dev/cleanup-and-full-e2e.sh

temporal-validate:
	@echo "🕒 校验前端时态工具引用..."
	cd frontend && npm run validate:temporal

validate-permissions:
	@echo "🔒 Plan 252 – 权限契约校验..."
	node scripts/quality/auth-permission-contract-validator.js \
	  --openapi docs/api/openapi.yaml \
	  --graphql docs/api/schema.graphql \
	  --resolver-dirs internal/organization/resolver,cmd/hrms-server/query/internal/auth \
	  --out reports/permissions \
	  --fail-on unregistered-scope,mapping-missing,resolver-bypass

plan252-evidence:
	@bash scripts/quality/plan252-evidence.sh

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
	@GOBIN="$$(go env GOPATH)/bin"; \
	if [ -x "$${GOBIN}/golangci-lint" ]; then \
	  echo "ℹ️  使用 pinned 版本: $${GOBIN}/golangci-lint"; \
	  "$${GOBIN}/golangci-lint" version; \
	  "$${GOBIN}/golangci-lint" run; \
	else \
	  echo "ℹ️  未检测到 pinned 版本，使用 PATH 中的 golangci-lint"; \
	  golangci-lint version || true; \
	  golangci-lint run; \
	fi
	@echo "🛡️ Plan 272 artifact guard..."
	npm run guard:plan272

workflow-lint:
	@echo "🔍 使用 actionlint 校验 .github/workflows ..."
	@bash scripts/ci/workflows/run-actionlint.sh $(ACTIONLINT_ARGS)

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
	@echo "  - Query (GraphQL):   http://localhost:9090/graphql  (GraphiQL: /graphiql)"
	@echo "  - PostgreSQL:        localhost:5432"
	@echo "  - Redis:             localhost:6379"
	@echo ""
	@echo "📈 REST Benchmark 汇总:"
	@bash scripts/diagnostics/check-rest-benchmark-summary.sh || true
	@echo ""
	@echo "🩺 Job Catalog 参考数据检查:"
	@bash scripts/diagnostics/check-job-catalog.sh

reset:
	@echo "🔄 重置最小依赖 (不删除卷)..."
	$(MAKE) docker-down
	$(MAKE) docker-up

# Plan 253 - 门禁脚本（端口/镜像标签）
guard-plan253:
	@echo "🛡️ 运行 Plan 253 门禁：compose 端口/镜像标签..."
	@bash scripts/quality/gates-253-compose-ports-and-images.sh

# Plan 253 - 冷启动计时（记录）
plan253-coldstart:
	@echo "⏱️  运行 Plan 253 冷启动计时（需要 Docker/Compose）..."
	@bash scripts/quality/gates-253-coldstart.sh

# Plan 256 - 契约生成与快照校验
generate-contracts:
	@echo "📋 运行 Plan 256 契约生成链路..."
	@mkdir -p logs/plan256
	@ts=$$(date +%Y%m%d_%H%M%S); bash scripts/contract/sync.sh 2>&1 | tee logs/plan256/sync-$$ts.log

verify-contracts:
	@echo "🧪 运行契约快照校验..."
	@python3 tests/contract/verify_inventory.py

# Plan 258 - 契约漂移字段矩阵门禁（阻断）
guard-plan258:
	@echo "🛡️ Plan 258 – 契约漂移字段矩阵门禁..."
	@mkdir -p logs/plan258
	@ts=$$(date +%Y%m%d_%H%M%S); \
	node scripts/contract/drift-check.js --include-fields --fail-on-diff 2>&1 | tee logs/plan258/gate-$$ts.log ; \
	cp -f reports/contracts/drift-report.json logs/plan258/drift-report-$$ts.json || true

# Plan 259A - 协议重复矩阵与白名单固化（聚合入口）
guard-plan259:
	@echo "🛡️ Plan 259A – 协议重复矩阵与白名单固化..."
	@mkdir -p logs/plan259 reports/plan259
	@ts=$$(date -u +%Y%m%d_%H%M%S); \
	echo "→ 生成权限契约与映射报告（Plan 252 脚本产物，供 258 聚合）"; \
	node scripts/quality/auth-permission-contract-validator.js --openapi docs/api/openapi.yaml --graphql docs/api/schema.graphql --out reports/permissions 2>&1 | tee logs/plan259/permissions-summary-$$ts.txt ; \
	echo "→ 生成协议重复矩阵（REST 业务 GET ↔ GraphQL Query）"; \
	node scripts/quality/protocol-duplication-matrix.js --openapi docs/api/openapi.yaml --graphql docs/api/schema.graphql --out reports/plan259/protocol-duplication-matrix.json 2>&1 | tee -a logs/plan259/guard-plan259-$$ts.log ; \
	echo "✔ 输出: reports/plan259/, logs/plan259/"; \
	echo "ℹ️  失败阈值可用 PLAN259_BUSINESS_GET_THRESHOLD 配置（默认=1；硬门禁期设为 0）"

# Plan 258 - 监听最近一次工作流运行（master）
watch-plan258:
	@echo "👀 Watching latest plan-258-gates run on master..."
	@bash scripts/ci/watch-actions-run.sh --workflow plan-258-gates --branch master --interval 10 --timeout 1800

# 迁移即真源：按序执行 database/migrations/*.sql（Goose）
db-migrate-all:
	@echo "🧭 使用 Goose 执行数据库迁移..."
	@command -v goose >/dev/null 2>&1 || { echo "❌ 需要安装 goose，请先执行: go install github.com/pressly/goose/v3/cmd/goose@latest"; exit 1; }
	@DB_URL="$$DATABASE_URL" ; \
	if [ -z "$$DB_URL" ]; then \
	  DB_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable" ; \
	  echo "ℹ️  未设置 DATABASE_URL，使用默认: $$DB_URL" ; \
	fi ; \
	set -e ; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DB_URL" goose -dir database/migrations status >/dev/null ; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DB_URL" goose -dir database/migrations up ; \
	echo "✅ Goose up 完成"

db-rollback-last:
	@echo "↩️  回滚最近一条 Goose 迁移..."
	@command -v goose >/dev/null 2>&1 || { echo "❌ 需要安装 goose，请先执行: go install github.com/pressly/goose/v3/cmd/goose@latest"; exit 1; }
	@DB_URL="$$DATABASE_URL" ; \
	if [ -z "$$DB_URL" ]; then \
	  DB_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable" ; \
	  echo "ℹ️  未设置 DATABASE_URL，使用默认: $$DB_URL" ; \
	fi ; \
	set -e ; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DB_URL" goose -dir database/migrations status >/dev/null ; \
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$$DB_URL" goose -dir database/migrations down ; \
	echo "✅ Goose down 完成"

sqlc-generate:
	@echo "🛠️ 执行 sqlc generate ..."
	@command -v sqlc >/dev/null 2>&1 || { echo "❌ 未找到 sqlc，请运行: go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.28.0"; exit 1; }
	@mkdir -p logs/plan402/schema
	@ts=$$(date -u +%Y%m%d-%H%M%S); \
	sqlc generate 2>&1 | tee "logs/plan402/schema/$$ts-sqlc-generate.log"


# 开发JWT工具
jwt-dev-mint:
	@echo "🔑 生成开发JWT..."
	@bash scripts/dev/mint-dev-jwt.sh \
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
