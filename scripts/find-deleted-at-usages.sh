#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
cd "$ROOT_DIR"

echo "==> 扫描包含 deleted_at 判定的文件"

IGNORE_PATHS=(
  "vendor"
  "node_modules"
  "frontend/node_modules"
  "logs"
  "backup"
  "reports/temporal"
)

IGNORE_ARGS=()
for path in "${IGNORE_PATHS[@]}"; do
  IGNORE_ARGS+=("--glob" "!$path/**")
done

rg "deleted_at" "${IGNORE_ARGS[@]}"

echo
echo "==> 建议下一步"
echo "1. 核对以上列表，确认是否仍依赖 deleted_at 判定。"
echo "2. 对需要保留的审计字段使用注释说明其用途。"
echo "3. 对误用场景创建任务并在迁移前修复。"
