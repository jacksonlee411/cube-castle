#!/usr/bin/env bash
set -euo pipefail

YELLOW='\033[33m'; GREEN='\033[32m'; RED='\033[31m'; NC='\033[0m'
step() { echo -e "${YELLOW}➡ $*${NC}"; }
ok()   { echo -e "${GREEN}✅ $*${NC}"; }
bad()  { echo -e "${RED}❌ $*${NC}"; }

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

PORTS=(3000 9090 8090)

kill_by_port() {
  local port="$1"
  local pids
  pids=$(lsof -t -i :"$port" -sTCP:LISTEN 2>/dev/null || true)
  if [[ -n "$pids" ]]; then
    echo "kill -9 $pids (port $port)"; kill -9 $pids || true
  fi
}

kill_by_name() {
  local name="$1"
  pkill -f -- "$name" 2>/dev/null || true
}

step "清理残留进程（playwright/vite/后端服务）"
kill_by_name 'playwright'
kill_by_name 'npm run test:e'
kill_by_name 'vite'
kill_by_name 'organization-command-service'
kill_by_name 'organization-query-service'
sleep 0.5

step "强制释放端口: ${PORTS[*]}"
for p in "${PORTS[@]}"; do kill_by_port "$p"; done

step "启动最小依赖 (postgres/redis)"
if command -v docker-compose >/dev/null 2>&1; then
  docker-compose up -d postgres redis
else
  docker compose up -d postgres redis
fi

mkdir -p secrets logs .cache

step "确保 RS256 密钥对存在"
if [[ ! -f secrets/dev-jwt-private.pem ]]; then
  openssl genrsa -out secrets/dev-jwt-private.pem 2048 >/dev/null 2>&1
  openssl rsa -in secrets/dev-jwt-private.pem -pubout -out secrets/dev-jwt-public.pem >/dev/null 2>&1
  ok "已生成 RS256 开发密钥对"
else
  ok "检测到已有 RS256 开发密钥对"
fi

step "启动命令服务 (9090 · RS256 mint + OIDC_SIMULATE)"
JWT_ALG=RS256 JWT_MINT_ALG=RS256 JWT_PRIVATE_KEY_PATH=secrets/dev-jwt-private.pem JWT_KEY_ID=bff-key-1 OIDC_SIMULATE=true \
  nohup bash -lc 'go run ./cmd/organization-command-service/main.go' > logs/command-service.log 2>&1 &

step "启动查询服务 (8090 · RS256 via JWKS)"
JWT_ALG=RS256 JWT_JWKS_URL=http://localhost:9090/.well-known/jwks.json \
  nohup bash -lc 'go run ./cmd/organization-query-service/main.go' > logs/query-service.log 2>&1 &

step "等待服务健康"
for i in $(seq 1 120); do
  CMD=0; QRY=0
  curl -fsS http://localhost:9090/health >/dev/null 2>&1 && CMD=1 || true
  curl -fsS http://localhost:8090/health >/dev/null 2>&1 && QRY=1 || true
  if [[ $CMD -eq 1 && $QRY -eq 1 ]]; then ok "后端健康"; break; fi
  sleep 1
  [[ $i -eq 120 ]] && { bad "后端健康检查超时"; tail -n 120 logs/command-service.log || true; tail -n 120 logs/query-service.log || true; exit 2; }
done

step "校验 JWKS"
curl -fsS http://localhost:9090/.well-known/jwks.json | head -n1 >/dev/null || { bad "JWKS 不可用"; exit 2; }

step "生成 RS256 测试令牌"
make -s jwt-dev-mint
export PW_JWT="$(cat .cache/dev.jwt)"
export PW_TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
[[ -n "$PW_JWT" ]] || { bad "未获取到测试令牌"; exit 2; }

step "运行前端 E2E（由 Playwright 自启 webServer）"
cd frontend
# 明确要求 Playwright 自启 webServer
export PW_SKIP_SERVER=0
npm run test:e2e

ok "E2E 执行完成"
