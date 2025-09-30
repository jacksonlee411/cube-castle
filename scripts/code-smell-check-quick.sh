#!/bin/bash
# 快速代码异味检查（简化版）
set -e

echo "=== 快速代码异味检查 ==="
echo ""

echo "🔍 Go后端红灯文件 (>800行):"
find cmd -name '*.go' -type f -exec wc -l {} + 2>/dev/null | awk '$1 > 800 {print "  🔴", $2, "("$1" 行)"}' | head -10

echo ""
echo "🔍 TypeScript前端红灯文件 (>800行):"
find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -exec wc -l {} + 2>/dev/null | awk '$1 > 800 {print "  🔴", $2, "("$1" 行)"}' | head -10

echo ""
echo "📊 Go文件统计:"
go_files=$(find cmd -name '*.go' -type f 2>/dev/null | wc -l)
go_red=$(find cmd -name '*.go' -type f -exec wc -l {} + 2>/dev/null | awk '$1 > 800' | wc -l)
echo "  总文件数: $go_files"
echo "  红灯文件 (>800行): $go_red"

echo ""
echo "📊 TypeScript文件统计:"
ts_files=$(find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) 2>/dev/null | wc -l)
ts_red=$(find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -exec wc -l {} + 2>/dev/null | awk '$1 > 800' | wc -l)
echo "  总文件数: $ts_files"
echo "  红灯文件 (>800行): $ts_red"

echo ""
if (( go_red > 0 || ts_red > 0 )); then
    echo "❌ 发现 $((go_red + ts_red)) 个红灯文件需要重构"
    exit 1
else
    echo "✅ 无红灯文件"
    exit 0
fi
