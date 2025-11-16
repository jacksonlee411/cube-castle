#!/usr/bin/env bash
set -euo pipefail

echo "[agents-compliance] 检查 API 命名一致性 (camelCase / {code})..."

violations=0

# 1) 检查 Go 结构体 json 标签中是否存在 snake_case（仅限 API 层目录）
go_tag_hits=$(grep -RIn --include='*.go' -E 'json:"[^"]*_[^"]*"' \
  cmd/hrms-server/command/internal \
  cmd/hrms-server/query/internal \
  shared \
  2>/dev/null || true)
# 允许例外：JWT Claims 等非 API 对外结构
go_tag_hits=$(echo "$go_tag_hits" | grep -vE '/internal/auth/jwt\.go:' || true)
# 允许例外：OIDC/OAuth 相关标准字段（BFF 内部结构），不作为对外 API 契约字段
go_tag_hits=$(echo "$go_tag_hits" | grep -vE '/internal/authbff/' || true)
if [[ -n "$go_tag_hits" ]]; then
  echo "✖ 发现 Go json 标签包含 snake_case："
  echo "$go_tag_hits"
  ((violations++))
fi

# 2) 检查 TS/JS 对象字面量键是否存在 snake_case（仅限前端源代码）
ts_obj_hits=$(grep -RIn --include='*.ts' --include='*.tsx' --include='*.js' --include='*.jsx' -E '"[a-z0-9]+_[a-z0-9]+"\s*:' \
  frontend/src frontend/tests \
  2>/dev/null || true)
if [[ -n "$ts_obj_hits" ]]; then
  echo "✖ 发现前端对象键包含 snake_case："
  echo "$ts_obj_hits"
  ((violations++))
fi

# 3) 检查路由是否错误使用 /{id} 作为组织单元路径参数（允许其他资源特定命名，如 record_id）
route_hits=$(grep -RIn --include='*.go' -E '/\{id\}' cmd/hrms-server/command/internal 2>/dev/null || true)
if [[ -n "$route_hits" ]]; then
  echo "✖ 发现使用 /{id} 的路由定义（应为 /{code}）："
  echo "$route_hits"
  ((violations++))
fi

if [[ $violations -gt 0 ]]; then
  echo "\n总计违反项: $violations"
  exit 1
fi

echo "✔ API 命名一致性检查通过"
exit 0
