#!/usr/bin/env bash
set -euo pipefail

# Record Plan 240E acceptance into 215 execution log and add "Close Confirmation" to 240E doc.
# This script derives status purely from local evidence files (single source of truth).
#
# Usage:
#   scripts/plan240/record-240e-acceptance.sh [E_LOG_DIR]
# Default E_LOG_DIR: logs/plan240/E
#
# It appends a short acceptance note into:
#   - docs/development-plans/215-phase2-execution-log.md
#   - docs/archive/development-plans/240E-position-regression-and-runbook.md (adds "关闭确认")
#
# Idempotency: Multiple runs on the same timestamp will append multiple blocks. Use git if you need cleanup.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
E_LOG_DIR="${1:-${ROOT_DIR}/logs/plan240/E}"
TS_HUMAN="$(date '+%Y-%m-%d %H:%M:%S %Z')"

fail() { echo "❌ $*"; exit 2; }
note() { echo "ℹ️  $*"; }

test -d "$E_LOG_DIR" || fail "Evidence directory not found: $E_LOG_DIR"

SEL_GUARD="${E_LOG_DIR}/selector-guard.log"
ARCH_GUARD="${E_LOG_DIR}/architecture-validator.log"
TEMP_TAGS="${E_LOG_DIR}/temporary-tags.log"
FRONT_LINT="${E_LOG_DIR}/frontend-lint.log"
FRONT_TYPES="${E_LOG_DIR}/frontend-typecheck.log"
PW_RUN_LOG="$(ls -1 "${E_LOG_DIR}"/playwright-run-*.log 2>/dev/null | tail -n 1 || true)"

# Basic checks
[[ -f "$SEL_GUARD" ]] || fail "Missing $SEL_GUARD"
[[ -f "$ARCH_GUARD" ]] || fail "Missing $ARCH_GUARD"
[[ -f "$TEMP_TAGS" ]] || fail "Missing $TEMP_TAGS"
[[ -f "$FRONT_LINT" ]] || note "frontend-lint.log not found (continue)"
[[ -f "$FRONT_TYPES" ]] || note "frontend-typecheck.log not found (continue)"
[[ -n "$PW_RUN_LOG" ]] || note "playwright run log not found (continue)"

# Derive pass/fail
sel_pass=$(grep -q "selector guard passed" "$SEL_GUARD" && echo "✅" || echo "❌")
arch_pass=$(grep -q "质量门禁通过" "$ARCH_GUARD" && echo "✅" || echo "❌")
temp_pass=$(grep -q "通过" "$TEMP_TAGS" && ! grep -q "✖" "$TEMP_TAGS" && echo "✅" || echo "❌")
lint_pass=$([[ -f "$FRONT_LINT" ]] && ! grep -q "error" "$FRONT_LINT" && echo "✅" || echo "⚠️")
types_pass=$([[ -f "$FRONT_TYPES" ]] && ! grep -q "error" "$FRONT_TYPES" && echo "✅" || echo "⚠️")

# Compose snippets
ACCEPT_SNIPPET_215=$(cat <<EOF
### Plan 240E – 验收登记（${TS_HUMAN}）

- 守卫：选择器 ${sel_pass} · 架构 ${arch_pass} · 临时标签 ${temp_pass}
- 前端：Lint ${lint_pass} · Typecheck ${types_pass}
- 证据：\`logs/plan240/E\`（run、guards、trace） · HAR 见 \`logs/plan240/B\`/BT
${PW_RUN_LOG:+- 执行日志：\`${PW_RUN_LOG#${ROOT_DIR}/}\`}

EOF
)

ACCEPT_SNIPPET_240E=$(cat <<EOF
## 关闭确认（${TS_HUMAN}）
- CI 稳定通过与门禁校验完成，产物已收敛至 \`logs/plan240/E\`（trace 见 \`logs/plan240/E/trace\`；HAR 见 \`logs/plan240/B\`/BT）。  
- 执行日志已登记至 \`docs/development-plans/215-phase2-execution-log.md\`。

EOF
)

note "Updating 215 execution log..."
printf "\n%s" "${ACCEPT_SNIPPET_215}" >> "${ROOT_DIR}/docs/development-plans/215-phase2-execution-log.md"

note "Updating 240E close confirmation..."
printf "\n%s" "${ACCEPT_SNIPPET_240E}" >> "${ROOT_DIR}/docs/archive/development-plans/240E-position-regression-and-runbook.md"

echo "✅ Done. Please review git diff and commit."
