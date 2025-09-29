#!/usr/bin/env bash
set -euo pipefail

echo "[agents-compliance] 检查 TODO-TEMPORARY 标注..."

# 切换到仓库根目录（脚本所在目录的上一级）
REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

# 搜索包含 TODO-TEMPORARY 的行（排除不相关目录）
ALLOWLIST_FILE=${TODO_TEMPORARY_ALLOWLIST:-scripts/todo-temporary-allowlist.txt}
declare -a allowlist=()
if [[ -f "$ALLOWLIST_FILE" ]]; then
  while IFS= read -r pattern; do
    [[ -z "$pattern" || "$pattern" =~ ^# ]] && continue
    allowlist+=("$pattern")
  done < "$ALLOWLIST_FILE"
fi

matches=$(grep -RIn -E 'TODO-TEMPORARY:' . \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  --exclude-dir=sql \
  --exclude-dir=database \
  --exclude-dir=backup \
  --exclude-dir=archive \
  --exclude-dir=docs/archive \
  --exclude='*check-temporary-tags.sh' \
  || true)

if [[ -z "$matches" ]]; then
  echo "✔ 未发现 TODO-TEMPORARY 标注（无需校验）"
  exit 0
fi

violation=0
critical_violation=0
while IFS= read -r line; do
  # 解析 path:lineno:text
  path=${line%%:*}
  rest=${line#*:}
  lineno=${rest%%:*}
  text=${rest#*:}
  path=${path#./}

  skip=false
  for pattern in "${allowlist[@]}"; do
    if [[ "$path" == $pattern ]]; then
      skip=true
      break
    fi
  done

  if [[ "$skip" == true ]]; then
    continue
  fi

  label="[代码]"
  if [[ $path == *.md || $path == README.md || $path == CHANGELOG.md || $path == docs/* || $path == .github/* ]]; then
    label="[文档]"
  fi

  # 针对 shared/types/api.ts 的强制阻断
  if [[ "$path" == "frontend/src/shared/types/api.ts" ]]; then
    cat <<EOF
✖ $label 禁止在 frontend/src/shared/types/api.ts 保留 TODO-TEMPORARY：
  - 请移除临时导出，改为从 shared/api/error-handling 或 shared/api/type-guards 导入错误类型与守卫。
  - 若仍需兼容，请在迁移计划中登记并提供新的截止日期。
问题位置：$path:$lineno
EOF
    ((violation++))
    critical_violation=1
    # 仍继续检查其他项，确保完整输出
  fi

  # 1) 必须包含截止日期 YYYY-MM-DD
  if ! echo "$text" | grep -Eq '20[0-9]{2}-[01][0-9]-[0-3][0-9]'; then
    echo "✖ $label 缺少截止日期(YYYY-MM-DD): $path:$lineno"
    ((violation++))
  fi

  # 2) 必须包含清晰的原因/计划（简单以内容长度近似约束）
  desc=$(echo "$text" | sed -E 's/.*TODO-TEMPORARY:\s*//')
  # 去掉空白统计长度
  desc_len=$(echo -n "$desc" | tr -d ' \t' | wc -c | tr -d ' ')
  if [[ ${desc_len} -lt 20 ]]; then
    echo "✖ $label 理由/计划描述过短(需≥20字符): $path:$lineno"
    ((violation++))
  fi
done <<< "$matches"

if [[ $violation -gt 0 ]]; then
  echo "\n总计违反项: $violation"
  if [[ $critical_violation -gt 0 ]]; then
    echo "⚠ 请优先处理 frontend/src/shared/types/api.ts 的临时导出回收，确保错误类型唯一入口。"
  fi
  exit 1
fi

echo "✔ TODO-TEMPORARY 标注规范通过"
exit 0
