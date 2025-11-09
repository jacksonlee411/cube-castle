#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"
MIGRATIONS_DIR="$PROJECT_ROOT/database/migrations"
CONTAINER_NAME="cube-castle-test-db"

info()  { printf '\033[0;32m[INFO]\033[0m %s\n' "$*"; }
warn()  { printf '\033[1;33m[WARN]\033[0m %s\n' "$*"; }
error() { printf '\033[0;31m[ERROR]\033[0m %s\n' "$*"; }

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    error "需要命令 $1，但当前环境未安装。"
    exit 1
  fi
}

cleanup() {
  info "清理 Docker 测试环境..."
  docker compose -f "$COMPOSE_FILE" down -v >/dev/null 2>&1 || true
}

trap cleanup EXIT

require_cmd docker
require_cmd goose
require_cmd go

info "启动测试数据库 (5432) ..."
docker compose -f "$COMPOSE_FILE" up -d postgres-test

info "等待 PostgreSQL 就绪..."
ready=0
for _ in $(seq 1 40); do
  if docker compose -f "$COMPOSE_FILE" exec -T postgres-test pg_isready -U testuser -d testdb >/dev/null 2>&1; then
    info "PostgreSQL 已就绪"
    ready=1
    break
  fi
  sleep 1
done

if [ "$ready" -ne 1 ]; then
  error "PostgreSQL 未能在预期时间内就绪"
  exit 1
fi

# 额外等待 1 秒，确保数据库完全初始化完毕（避免 pg_isready 刚通过时的抖动）
sleep 1

DB_HOST=localhost
DB_PORT=5432
DB_USER=testuser
DB_PASSWORD=testpassword
DB_NAME=testdb
export DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

info "执行 Goose 迁移 (up)..."
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" up

info "运行 Go 集成测试 (-tags=integration)..."
go test -v -tags=integration ./...

info "执行 Goose 回滚 (down)..."
GOOSE_DRIVER=postgres GOOSE_DBSTRING="$DATABASE_URL" goose -dir "$MIGRATIONS_DIR" down

info "集成测试完成"
