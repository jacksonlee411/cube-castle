#!/usr/bin/env bash
set -euo pipefail

echo "[agents-compliance] 检查 TODO-TEMPORARY 标注..."

# 切换到仓库根目录（脚本所在目录的上一级）
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

# 搜索包含 TODO-TEMPORARY 的行（排除不相关目录）
matches=$(grep -RIn -E 'TODO-TEMPORARY:' . \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  --exclude-dir=sql \
  --exclude-dir=database \
  --exclude-dir=backup \
  --exclude-dir=archive \
  --exclude='*.md' \
  --exclude='*check-temporary-tags.sh' \
  || true)

if [[ -z "$matches" ]]; then
  echo "✔ 未发现 TODO-TEMPORARY 标注（无需校验）"
  exit 0
fi

violation=0
while IFS= read -r line; do
  # 解析 path:lineno:text
  path=${line%%:*}
  rest=${line#*:}
  lineno=${rest%%:*}
  text=${rest#*:}

  # 1) 必须包含截止日期 YYYY-MM-DD
  if ! echo "$text" | grep -Eq '20[0-9]{2}-[01][0-9]-[0-3][0-9]'; then
    echo "✖ 缺少截止日期(YYYY-MM-DD): $path:$lineno"
    ((violation++))
  fi

  # 2) 必须包含清晰的原因/计划（简单以内容长度近似约束）
  desc=$(echo "$text" | sed -E 's/.*TODO-TEMPORARY:\s*//')
  # 去掉空白统计长度
  desc_len=$(echo -n "$desc" | tr -d ' \t' | wc -c | tr -d ' ')
  if [[ ${desc_len} -lt 20 ]]; then
    echo "✖ 理由/计划描述过短(需≥20字符): $path:$lineno"
    ((violation++))
  fi
done <<< "$matches"

if [[ $violation -gt 0 ]]; then
  echo "\n总计违反项: $violation"
  exit 1
fi

echo "✔ TODO-TEMPORARY 标注规范通过"
exit 0
