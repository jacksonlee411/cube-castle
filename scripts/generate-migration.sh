#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<USAGE
用法: $0 <migration-name>
  - 需要先更新 database/schema.sql（例如通过 pg_dump）
  - 默认使用 Atlas 生成 Goose 迁移，若 Atlas 不可用，则导出当前 Schema 以便手工编写迁移
USAGE
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

MIGRATION_NAME=$1
ATLAS_BIN=${ATLAS_BIN:-$(command -v atlas || true)}
GOOSE_BIN=${GOOSE_BIN:-$(command -v goose || true)}
ATLAS_ENV=${ATLAS_ENV:-dev}

if [[ -z "$ATLAS_BIN" ]]; then
  echo "[WARN] atlas 未安装，改为重新导出 database/schema.sql，后续迁移需手动编写。" >&2
  docker compose exec -T postgres pg_dump --schema-only --no-owner --no-privileges \
    -U user cubecastle > database/schema.sql
  exit 0
fi

${ATLAS_BIN} migrate diff "${MIGRATION_NAME}" --env "${ATLAS_ENV}" --config atlas.hcl

if [[ -n "$GOOSE_BIN" ]]; then
  ${GOOSE_BIN} -dir database/migrations fix
fi

echo "生成完成: database/migrations 下新增 Goose 迁移，请审阅并补充函数/触发器等特殊对象。"
