#!/usr/bin/env bash
set -euo pipefail

# Guard: DEV_MODE 默认值必须为 false（生产禁用开发模式）
# 检查查询服务的默认值
FILE="cmd/hrms-server/query/internal/app/app.go"
if ! grep -RInq 'getEnv("DEV_MODE", "false")' "$FILE"; then
  echo "✖ DEV_MODE 默认值检查失败：${FILE} 未设置 getEnv(\"DEV_MODE\", \"false\")"
  exit 1
fi

echo "✔ DEV_MODE 默认值检查通过（默认 false）"
exit 0

