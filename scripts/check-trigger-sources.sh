#!/usr/bin/env bash
set -euo pipefail

echo "[compliance] 扫描触发器定义来源..."

# 允许环境变量跳过（例如：在本地审阅遗留脚本时）
if [[ "${ALLOW_TRIGGER_SCRIPTS:-}" == "1" ]]; then
  echo "跳过触发器来源检查 (ALLOW_TRIGGER_SCRIPTS=1)"
  exit 0
fi

# 仅允许在以下目录存在触发器定义：
#  - sql/init
#  - database/migrations
ALLOWLIST_DIRS=(
  "cube-castle/sql/init"
  "cube-castle/database/migrations"
)

# 匹配触发器定义的正则
PATTERN='(CREATE\s+TRIGGER|RETURNS\s+TRIGGER)'

violations=0

# 读取允许的例外列表（历史/演示脚本白名单）
ALLOWLIST_FILE="cube-castle/scripts/.trigger-allowlist.txt"
declare -A allowmap
if [[ -f "$ALLOWLIST_FILE" ]]; then
  while IFS= read -r line; do
    [[ -z "$line" || "$line" =~ ^# ]] && continue
    allowmap["$line"]=1
  done < "$ALLOWLIST_FILE"
fi

# 扫描仓库内所有包含 CREATE TRIGGER/RETURNS TRIGGER 的文件
mapfile -t candidates < <(rg -I -n --no-heading -e "$PATTERN" cube-castle -S | cut -d: -f1 | sort -u)
files=()
for p in "${candidates[@]:-}"; do
  [[ -f "$p" ]] || continue
  files+=("$p")
done

for f in "${files[@]}"; do
  # 检查是否位于允许目录
  allowed=false
  for d in "${ALLOWLIST_DIRS[@]}"; do
    if [[ "$f" == $d* ]]; then
      allowed=true
      break
    fi
  done

  # 白名单例外
  if [[ -n "${allowmap[$f]:-}" ]]; then
    continue
  fi

  if ! $allowed; then
    echo "✖ 非授权目录中的触发器定义: $f"
    ((violations++))
  fi
done

if [[ $violations -gt 0 ]]; then
  echo "\n建议：将上述脚本迁移到 database/migrations 或归档到 archive/，避免误导与误执行。"
  exit 1
fi

echo "✔ 触发器定义位置合规"
exit 0
