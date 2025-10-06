#!/bin/bash
# Code Smell Monitor - 代码异味监控脚本
# 用途：监控Go和TypeScript文件的行数分布，支持函数级别检查
# 版本：v1.0 (2025-09-30)
# 对应计划：Plan 16 代码异味治理（Phase 3 文档收尾）
# 使用说明：详见本脚本 `usage()`。常见场景：`./scripts/code-smell-monitor.sh --files --report`、`./scripts/code-smell-monitor.sh --functions --ci`。

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
ORANGE='\033[0;33m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# 默认阈值（基于Go工程实践标准）
RED_THRESHOLD=800      # 红灯：强制重构
ORANGE_THRESHOLD=600   # 橙灯：需评估
YELLOW_THRESHOLD=400   # 黄灯：关注

# 使用说明
usage() {
    cat << EOF
用法: $0 [选项]

选项:
  --files          检查文件行数分布（默认）
  --functions      检查函数复杂度（>100行的函数）
  --report         生成详细报告到 reports/iig-guardian/
  --baseline FILE  与基线报告对比
  --ci             CI模式（失败时返回非零退出码）
  --help           显示此帮助信息

示例:
  $0 --files                           # 检查文件行数
  $0 --functions                       # 检查函数复杂度
  $0 --report                          # 生成详细报告
  $0 --baseline reports/iig-guardian/code-smell-baseline-20250929.md
  $0 --ci                              # CI门禁模式

退出码:
  0 - 所有检查通过
  1 - 发现红灯文件（>800行）
  2 - 脚本执行错误
EOF
    exit 0
}

# 检查文件行数
check_files() {
    echo "=== 检查Go后端文件 ==="
    local go_red=0
    local go_orange=0
    local go_yellow=0
    local go_green=0
    local go_total=0

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            local lines=$(wc -l < "$file")
            ((go_total++))

            if (( lines > RED_THRESHOLD )); then
                echo -e "${RED}🔴 红灯${NC}: $file ($lines 行)"
                ((go_red++))
            elif (( lines > ORANGE_THRESHOLD )); then
                echo -e "${ORANGE}🟠 橙灯${NC}: $file ($lines 行)"
                ((go_orange++))
            elif (( lines > YELLOW_THRESHOLD )); then
                echo -e "${YELLOW}🟡 黄灯${NC}: $file ($lines 行)"
                ((go_yellow++))
            else
                ((go_green++))
            fi
        fi
    done < <(find cmd -name '*.go' -type f 2>/dev/null)

    echo ""
    echo "Go文件统计："
    echo "  🔴 红灯 (>800行): $go_red"
    echo "  🟠 橙灯 (600-800行): $go_orange"
    echo "  🟡 黄灯 (400-600行): $go_yellow"
    echo "  🟢 绿灯 (<400行): $go_green"
    echo "  总计: $go_total"
    echo ""

    echo "=== 检查前端TypeScript文件 ==="
    local ts_red=0
    local ts_orange=0
    local ts_green=0
    local ts_total=0

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            local lines=$(wc -l < "$file")
            ((ts_total++))

            if (( lines > RED_THRESHOLD )); then
                echo -e "${RED}🔴 红灯${NC}: $file ($lines 行)"
                ((ts_red++))
            elif (( lines > YELLOW_THRESHOLD )); then
                echo -e "${ORANGE}🟠 橙灯${NC}: $file ($lines 行)"
                ((ts_orange++))
            else
                ((ts_green++))
            fi
        fi
    done < <(find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) 2>/dev/null)

    echo ""
    echo "TypeScript文件统计："
    echo "  🔴 红灯 (>800行): $ts_red"
    echo "  🟠 橙灯 (400-800行): $ts_orange"
    echo "  🟢 绿灯 (<400行): $ts_green"
    echo "  总计: $ts_total"
    echo ""

    # CI模式：红灯文件存在时返回1
    if [[ "${CI_MODE:-false}" == "true" ]]; then
        if (( go_red > 0 || ts_red > 0 )); then
            echo "❌ CI检查失败：发现 $((go_red + ts_red)) 个红灯文件"
            return 1
        fi
        echo "✅ CI检查通过：无红灯文件"
    fi

    return 0
}

# 检查函数复杂度（Go）
check_functions_go() {
    echo "=== 检查Go函数复杂度 (>100行) ==="
    local count=0

    while IFS= read -r file; do
        if [[ -f "$file" ]]; then
            # 使用awk统计函数行数
            awk '
                /^func / {
                    if (func_name != "") {
                        lines = NR - func_start
                        if (lines > 100) {
                            printf "  %s:%d - %s (%d 行)\n", FILENAME, func_start, func_name, lines
                            count++
                        }
                    }
                    func_name = $0
                    func_start = NR
                }
                END {
                    if (func_name != "") {
                        lines = NR - func_start + 1
                        if (lines > 100) {
                            printf "  %s:%d - %s (%d 行)\n", FILENAME, func_start, func_name, lines
                            count++
                        }
                    }
                }
            ' "$file"
        fi
    done < <(find cmd -name '*.go' -type f 2>/dev/null)

    echo ""
    if (( count == 0 )); then
        echo "✅ 未发现超过100行的Go函数"
    else
        echo "⚠️  发现 $count 个超过100行的Go函数"
    fi
}

# 生成详细报告
generate_report() {
    local report_date=$(date +%Y%m%d)
    local report_file="reports/iig-guardian/code-smell-progress-${report_date}.md"

    mkdir -p reports/iig-guardian

    cat > "$report_file" << 'EOF'
# 代码异味进展报告

**生成日期**: $(date +%Y-%m-%d)
**报告版本**: 自动生成
**对比基线**: reports/iig-guardian/code-smell-baseline-20250929.md

---

## 当前状态

### Go后端文件分布
EOF

    # 生成Go统计
    echo "" >> "$report_file"
    find cmd -name '*.go' -type f -print0 2>/dev/null | xargs -0 wc -l | sort -rn | head -20 >> "$report_file"

    cat >> "$report_file" << 'EOF'

### TypeScript前端文件分布
EOF

    # 生成TS统计
    echo "" >> "$report_file"
    find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -print0 2>/dev/null | xargs -0 wc -l | sort -rn | head -20 >> "$report_file"

    echo ""
    echo "✅ 报告已生成：$report_file"
}

# 与基线对比
compare_baseline() {
    local baseline_file="$1"

    if [[ ! -f "$baseline_file" ]]; then
        echo "❌ 基线文件不存在：$baseline_file"
        return 2
    fi

    echo "=== 与基线对比 ==="
    echo "基线文件：$baseline_file"
    echo ""
    echo "功能开发中...（Phase 3交付）"
    # TODO: 实现基线对比逻辑
}

# 主函数
main() {
    local mode="files"
    local baseline_file=""
    CI_MODE=false

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --files)
                mode="files"
                shift
                ;;
            --functions)
                mode="functions"
                shift
                ;;
            --report)
                mode="report"
                shift
                ;;
            --baseline)
                baseline_file="$2"
                mode="baseline"
                shift 2
                ;;
            --ci)
                CI_MODE=true
                shift
                ;;
            --help)
                usage
                ;;
            *)
                echo "未知选项: $1"
                usage
                ;;
        esac
    done

    # 执行对应模式
    case $mode in
        files)
            check_files
            ;;
        functions)
            check_functions_go
            ;;
        report)
            generate_report
            ;;
        baseline)
            compare_baseline "$baseline_file"
            ;;
    esac
}

# 执行主函数
main "$@"
