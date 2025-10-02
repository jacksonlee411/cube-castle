#!/bin/bash
# 快速代码异味检查（支持规模与弱类型巡检）
set -euo pipefail

WITH_TYPES=0
CI_OUTPUT=""
TYPE_THRESHOLD=""
VERIFY_ONLY=0

usage() {
  cat <<'EOF'
用法: scripts/code-smell-check-quick.sh [选项]

选项:
  --with-types              同时扫描 TypeScript 中的 any/unknown 使用
  --type-threshold <数值>   设置弱类型报警阈值（默认: 30）
  --ci-output <路径>        将输出写入指定 Markdown 文件（同步打印到控制台）
  --verify-only             仅输出结果，不根据阈值退出
  -h, --help                显示本帮助
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --with-types)
      WITH_TYPES=1
      shift
      ;;
    --type-threshold)
      TYPE_THRESHOLD="$2"
      shift 2
      ;;
    --ci-output)
      CI_OUTPUT="$2"
      shift 2
      ;;
    --verify-only)
      VERIFY_ONLY=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -n "$CI_OUTPUT" ]]; then
  mkdir -p "$(dirname "$CI_OUTPUT")"
  : >"$CI_OUTPUT"
fi

log() {
  local line="$1"
  echo "$line"
  if [[ -n "$CI_OUTPUT" ]]; then
    printf '%s\n' "$line" >>"$CI_OUTPUT"
  fi
}

announce_section() {
  local title="$1"
  log ""
  log "${title}"
}

log "=== 快速代码异味检查 ==="

announce_section "🔍 Go后端红灯文件 (>800行):"
find cmd -name '*.go' -type f -exec wc -l {} + 2>/dev/null | awk '$1 > 800 && $2 != "total" {print "  🔴", $2, "("$1" 行)"}' | head -10 | while read -r line; do
  [[ -n "$line" ]] && log "$line"
done

announce_section "🔍 TypeScript前端红灯文件 (>800行):"
find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -exec wc -l {} + 2>/dev/null | awk '$1 > 800 && $2 != "total" {print "  🔴", $2, "("$1" 行)"}' | head -10 | while read -r line; do
  [[ -n "$line" ]] && log "$line"
done

announce_section "📊 Go文件统计:"
go_files=$(find cmd -name '*.go' -type f 2>/dev/null | wc -l | tr -d ' ')
go_red=$(find cmd -name '*.go' -type f -exec wc -l {} + 2>/dev/null | awk '$1 > 800 && $2 != "total"' | wc -l | tr -d ' ')
log "  总文件数: $go_files"
log "  红灯文件 (>800行): $go_red"

announce_section "📊 TypeScript文件统计:"
ts_files=$(find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) 2>/dev/null | wc -l | tr -d ' ')
ts_red=$(find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -exec wc -l {} + 2>/dev/null | awk '$1 > 800 && $2 != "total"' | wc -l | tr -d ' ')
log "  总文件数: $ts_files"
log "  红灯文件 (>800行): $ts_red"

exit_code=0

if (( go_red > 0 || ts_red > 0 )); then
  log ""
  log "❌ 发现 $((go_red + ts_red)) 个红灯文件需要重构"
  if (( VERIFY_ONLY == 0 )); then
    exit_code=1
  fi
else
  log ""
  log "✅ 无红灯文件"
fi

if (( WITH_TYPES == 1 )); then
  TYPE_THRESHOLD=${TYPE_THRESHOLD:-30}
  announce_section "📈 弱类型使用统计 (TypeScript any/unknown)"

  type_matches=$(rg -g '*.{ts,tsx}' -o -e '\bany\b|\bunknown\b' frontend/src 2>/dev/null || true)
  type_matches=$(printf '%s' "$type_matches" | wc -l | tr -d ' ')
  type_files=$(rg -g '*.{ts,tsx}' -l -e '\bany\b|\bunknown\b' frontend/src 2>/dev/null || true)
  type_files=$(printf '%s' "$type_files" | wc -l | tr -d ' ')

  log ""
  log "  ➤ 匹配次数: $type_matches"
  log "  ➤ 涉及文件: $type_files"
  log "  ➤ 阈值 (any/unknown): $TYPE_THRESHOLD"

  top_files=$(rg -g '*.{ts,tsx}' -o -e '\bany\b|\bunknown\b' -n frontend/src 2>/dev/null || true)
  top_files=$(printf '%s' "$top_files" | cut -d: -f1 | sort | uniq -c | sort -nr | head -10)
  if [[ -n "$top_files" ]]; then
    log ""
    log "  Top 10 文件 (按弱类型出现次数):"
    while read -r count filepath; do
      [[ -n "$filepath" ]] && log "    - ${filepath}: ${count}"
    done <<<"$top_files"
  fi

  if (( type_matches > TYPE_THRESHOLD )); then
    log ""
    log "❌ any/unknown 数量 $type_matches 超过阈值 $TYPE_THRESHOLD"
    if (( VERIFY_ONLY == 0 )); then
      exit_code=1
    fi
  else
    log ""
    log "✅ any/unknown 数量在阈值内"
  fi
fi

exit "$exit_code"
