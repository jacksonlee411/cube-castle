#!/bin/bash

# 🔍 Cube Castle - 本地重复代码检测脚本
# 用途: 在提交前本地执行重复代码检测，确保代码质量
# 作者: Claude Code Assistant
# 日期: 2025-09-07

set -euo pipefail

# 🎨 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# 📊 配置变量
DEFAULT_THRESHOLD=5
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
REPORT_DIR="$PROJECT_ROOT/reports/duplicate-code"

# 📋 使用帮助
show_help() {
    cat << EOF
🔍 Cube Castle - 本地重复代码检测工具

用法:
    $0 [选项]

选项:
    -t, --threshold PERCENT    设置重复代码阈值百分比 (默认: $DEFAULT_THRESHOLD)
    -s, --scope SCOPE         扫描范围: full|frontend|backend|changed (默认: full)
    -f, --format FORMAT       输出格式: console|html|json|all (默认: console)
    -q, --quiet              静默模式，只显示结果
    -h, --help               显示帮助信息

示例:
    $0                        # 使用默认设置扫描
    $0 -t 3 -s frontend       # 扫描前端，阈值3%
    $0 -s changed -f html     # 扫描变更文件，生成HTML报告
    $0 --quiet                # 静默模式扫描

环境变量:
    JSCPD_THRESHOLD          重复代码阈值 (覆盖 -t 参数)
    JSCPD_FORMAT            输出格式 (覆盖 -f 参数)
EOF
}

# 🛠️ 解析命令行参数
THRESHOLD=${JSCPD_THRESHOLD:-$DEFAULT_THRESHOLD}
SCOPE="full"
FORMAT=${JSCPD_FORMAT:-"console"}
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--threshold)
            THRESHOLD="$2"
            shift 2
            ;;
        -s|--scope)
            SCOPE="$2"
            shift 2
            ;;
        -f|--format)
            FORMAT="$2"
            shift 2
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}错误: 未知参数 $1${NC}" >&2
            show_help
            exit 1
            ;;
    esac
done

# 🔍 日志函数
log_info() {
    [[ "$QUIET" == "true" ]] || echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}" >&2
}

# 🚀 主函数
main() {
    log_info "开始Cube Castle重复代码检测..."
    log_info "配置: 阈值=${THRESHOLD}%, 范围=${SCOPE}, 格式=${FORMAT}"

    # 🔧 检查依赖
    if ! command -v jscpd >/dev/null 2>&1; then
        log_error "未安装jscpd工具，请运行: npm install -g jscpd"
        exit 1
    fi

    if ! command -v node >/dev/null 2>&1; then
        log_error "未安装Node.js，请安装Node.js 18+"
        exit 1
    fi

    # 📁 创建报告目录
    mkdir -p "$REPORT_DIR"
    cd "$PROJECT_ROOT"

    # 🎯 确定扫描目标
    local scan_paths=""
    local scan_formats=""

    case "$SCOPE" in
        "full")
            scan_paths="."
            scan_formats="typescript,javascript,go,markdown"
            log_info "扫描范围: 完整项目"
            ;;
        "frontend")
            scan_paths="frontend/"
            scan_formats="typescript,javascript"
            log_info "扫描范围: 前端代码 (TypeScript/JavaScript)"
            ;;
        "backend")
            scan_paths="cmd/ internal/ pkg/"
            scan_formats="go"
            log_info "扫描范围: 后端代码 (Go)"
            ;;
        "changed")
            # 获取变更文件列表
            if git rev-parse --git-dir >/dev/null 2>&1; then
                local changed_files
                changed_files=$(git diff --name-only HEAD 2>/dev/null | tr '\n' ' ')
                if [[ -z "$changed_files" ]]; then
                    changed_files=$(git diff --name-only --cached 2>/dev/null | tr '\n' ' ')
                fi
                
                if [[ -n "$changed_files" ]]; then
                    scan_paths="$changed_files"
                    scan_formats="typescript,javascript,go"
                    log_info "扫描范围: 变更文件 ($changed_files)"
                else
                    log_warning "未检测到文件变更，切换到完整扫描"
                    scan_paths="."
                    scan_formats="typescript,javascript,go,markdown"
                fi
            else
                log_warning "非Git仓库，切换到完整扫描"
                scan_paths="."
                scan_formats="typescript,javascript,go,markdown"
            fi
            ;;
        *)
            log_error "无效的扫描范围: $SCOPE"
            show_help
            exit 1
            ;;
    esac

    # 📊 设置报告格式
    local reporters=""
    case "$FORMAT" in
        "console")
            reporters="console"
            ;;
        "html")
            reporters="html"
            ;;
        "json")
            reporters="json"
            ;;
        "all")
            reporters="console,html,json"
            ;;
        *)
            log_error "无效的输出格式: $FORMAT"
            exit 1
            ;;
    esac

    log_info "开始检测重复代码..."

    # 🔍 执行jscpd检测
    local jscpd_cmd="jscpd"
    jscpd_cmd="$jscpd_cmd --threshold=$THRESHOLD"
    jscpd_cmd="$jscpd_cmd --reporters=$reporters"
    jscpd_cmd="$jscpd_cmd --output=$REPORT_DIR"
    jscpd_cmd="$jscpd_cmd --format=$scan_formats"
    jscpd_cmd="$jscpd_cmd --config=$PROJECT_ROOT/.jscpdrc.json"

    # 添加扫描路径
    jscpd_cmd="$jscpd_cmd $scan_paths"

    log_info "执行命令: $jscpd_cmd"

    local exit_code=0
    if [[ "$QUIET" == "true" ]]; then
        eval "$jscpd_cmd" >/dev/null 2>&1 || exit_code=$?
    else
        eval "$jscpd_cmd" || exit_code=$?
    fi

    # 📈 解析和展示结果
    local json_report="$REPORT_DIR/jscpd-report.json"
    if [[ -f "$json_report" ]]; then
        log_info "生成检测统计报告..."
        
        local stats
        stats=$(node -e "
            const fs = require('fs');
            try {
                const report = JSON.parse(fs.readFileSync('$json_report', 'utf8'));
                const stats = report.statistics.total;
                console.log(JSON.stringify({
                    sources: stats.sources,
                    lines: stats.lines,
                    duplicatedLines: stats.duplicatedLines,
                    percentage: stats.percentage,
                    clones: stats.clones
                }));
            } catch (e) {
                console.log(JSON.stringify({error: e.message}));
            }
        " 2>/dev/null || echo '{"error": "解析失败"}')

        if echo "$stats" | grep -q '"error"'; then
            log_warning "无法解析检测报告"
        else
            log_info "📊 检测结果统计:"
            node -e "
                const stats = $stats;
                console.log('📁 扫描文件数: ' + stats.sources);
                console.log('📋 代码总行数: ' + stats.lines.toLocaleString());
                console.log('🔍 重复行数: ' + stats.duplicatedLines.toLocaleString());
                console.log('📊 重复率: ' + stats.percentage.toFixed(2) + '%');
                console.log('⚠️ 重复片段: ' + stats.clones);
                console.log('🎯 阈值标准: $THRESHOLD%');
                
                if (stats.percentage <= $THRESHOLD) {
                    console.log('✅ 质量状态: 通过');
                } else {
                    console.log('❌ 质量状态: 超过阈值');
                }
            "

            # 🎯 质量门禁检查
            local current_percentage
            current_percentage=$(node -e "const stats = $stats; console.log(stats.percentage);" 2>/dev/null || echo "0")
            
            if command -v bc >/dev/null 2>&1; then
                local exceeds_threshold
                exceeds_threshold=$(echo "$current_percentage > $THRESHOLD" | bc -l 2>/dev/null || echo "0")
                
                if [[ "$exceeds_threshold" == "1" ]]; then
                    log_error "质量门禁失败: 重复率 ${current_percentage}% 超过阈值 ${THRESHOLD}%"
                    log_error "请重构重复代码后再次提交"
                    exit_code=1
                else
                    log_success "质量门禁通过: 重复率 ${current_percentage}% 符合标准"
                fi
            else
                log_warning "无法进行精确阈值比较（缺少bc工具）"
            fi
        fi
    else
        log_warning "未找到JSON报告文件，跳过统计分析"
    fi

    # 📊 显示报告位置
    if [[ "$FORMAT" != "console" ]]; then
        log_info "📂 报告文件位置:"
        if [[ "$FORMAT" == "html" || "$FORMAT" == "all" ]] && [[ -f "$REPORT_DIR/html/index.html" ]]; then
            log_info "   HTML报告: $REPORT_DIR/html/index.html"
        fi
        if [[ "$FORMAT" == "json" || "$FORMAT" == "all" ]] && [[ -f "$REPORT_DIR/jscpd-report.json" ]]; then
            log_info "   JSON报告: $REPORT_DIR/jscpd-report.json"
        fi
    fi

    # 🎉 最终结果
    if [[ $exit_code -eq 0 ]]; then
        log_success "重复代码检测完成，质量标准通过！"
        log_info "可以安全提交代码"
    else
        log_error "重复代码检测失败，请先重构代码"
        log_info "建议查看详细报告定位重复代码片段"
    fi

    exit $exit_code
}

# 🎯 程序入口
main "$@"