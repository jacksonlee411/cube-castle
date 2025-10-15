#!/usr/bin/env bash
set -euo pipefail

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
NC=$'\033[0m'

fail() {
  echo "${RED}❌ $1${NC}"
  exit 1
}

pass() {
  echo "${GREEN}✅ $1${NC}"
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# 1. Makefile run-dev 使用 docker compose，且未直接调用 go run
if ! grep -q 'run-dev:' "$ROOT_DIR/Makefile"; then
  fail "Makefile 中缺少 run-dev 目标"
fi

if ! awk '
  $0 ~ /^run-dev:/ { in_target=1; next }
  in_target && $0 ~ /^$/ { in_target=0 }
  in_target { print }
' "$ROOT_DIR/Makefile" | grep -q 'docker compose -f docker-compose.dev.yml up -d --build'; then
  fail "Makefile run-dev 未调用 docker compose -f docker-compose.dev.yml up -d --build"
fi

if awk '
  $0 ~ /^run-dev:/ { in_target=1; next }
  in_target && $0 ~ /^$/ { in_target=0 }
  in_target { print }
' "$ROOT_DIR/Makefile" | grep -q 'go run'; then
  fail "Makefile run-dev 仍包含 go run 调用，违反 Docker 强制原则"
fi
pass "Makefile run-dev 仅使用 docker compose"

# 2. .env 默认 DATABASE_URL 指向容器主机名 postgres
if ! grep -E '^DATABASE_URL=postgresql://[^@]+@postgres:5432' "$ROOT_DIR/.env" >/dev/null; then
  fail ".env 中 DATABASE_URL 未默认指向容器主机名 postgres"
fi
pass ".env 默认 DATABASE_URL 使用 postgres 主机"

# 3. docker-compose.dev.yml 不应包含 profiles
if grep -q 'profiles:' "$ROOT_DIR/docker-compose.dev.yml"; then
  fail "docker-compose.dev.yml 含 profiles 配置，会导致应用服务默认不启动"
fi
pass "docker-compose.dev.yml 无 profiles 配置"

# 4. dev-start-simple.sh 必须阻止执行
if ! grep -q '本脚本已废弃' "$ROOT_DIR/scripts/dev-start-simple.sh"; then
  fail "scripts/dev-start-simple.sh 未包含废弃提示"
fi
if ! awk 'NR==FNR {total=total+length($0); next} NR==1 { print $0 }' "$ROOT_DIR/scripts/dev-start-simple.sh" >/dev/null; then
  true # no-op to avoid shellcheck warning
fi
if ! awk 'NR<=15 { print }' "$ROOT_DIR/scripts/dev-start-simple.sh" | grep -q 'exit 1'; then
  fail "scripts/dev-start-simple.sh 须在顶部退出以防误用"
fi
pass "dev-start-simple.sh 已禁止直接执行"

echo "${GREEN}🎯 Docker 合规检查全部通过${NC}"
