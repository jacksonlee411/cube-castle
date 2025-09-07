#!/bin/bash

# 🏗️ Cube Castle - 架构守护验证脚本
# 用途: 验证架构一致性和API契约合规性，确保企业级代码质量
# 作者: Claude Code Assistant
# 日期: 2025-09-07

set -euo pipefail

# 🎨 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# 📊 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ESLINT_CONFIG="$PROJECT_ROOT/.eslintrc.architecture.js"
REPORT_DIR="$PROJECT_ROOT/reports/architecture"

# 📋 使用帮助
show_help() {
    cat << EOF
🏗️ Cube Castle - 架构守护验证工具

用法:
    $0 [选项]

选项:
    -s, --scope SCOPE         验证范围: all|frontend|backend|config (默认: all)
    -f, --fix                自动修复可修复的问题
    -r, --report FORMAT      报告格式: console|json|html (默认: console)
    -v, --verbose            详细输出模式
    -q, --quiet              静默模式，仅显示结果
    -c, --config FILE        自定义ESLint配置文件
    --rules RULES            指定要检查的规则类别 (architecture|naming|imports|all)
    -h, --help               显示帮助信息

验证项目:
    🔍 CQRS架构守护:
       - 禁止前端REST查询，强制GraphQL
       - 确保命令操作使用REST API
       - 验证协议职责分离

    🔧 配置管理守护:
       - 检测硬编码端口号
       - 强制使用统一配置模块
       - 验证端点配置一致性

    📋 API契约守护:
       - 强制camelCase字段命名
       - 检查废弃字段使用
       - 验证标准字段词汇表

    🏛️ 代码质量守护:
       - TypeScript命名约定
       - 导入规范验证
       - 架构特定禁止项

示例:
    $0                          # 完整架构验证
    $0 -s frontend --fix        # 验证前端并自动修复
    $0 -r html -v               # 生成HTML报告，详细输出
    $0 --rules architecture     # 仅检查架构规则
    $0 -q --rules naming        # 静默检查命名规范

环境变量:
    ARCH_SCOPE                验证范围 (覆盖 -s 参数)
    ARCH_FIX                  自动修复 (设为 true 启用)
    ARCH_QUIET               静默模式 (设为 true 启用)
    ESLINT_CONFIG_OVERRIDE   自定义配置文件路径
EOF
}

# 🛠️ 解析命令行参数
SCOPE=${ARCH_SCOPE:-"all"}
FIX_ENABLED=${ARCH_FIX:-false}
REPORT_FORMAT="console"
VERBOSE=false
QUIET=${ARCH_QUIET:-false}
RULES_FILTER="all"
CUSTOM_CONFIG=${ESLINT_CONFIG_OVERRIDE:-""}

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--scope)
            SCOPE="$2"
            shift 2
            ;;
        -f|--fix)
            FIX_ENABLED=true
            shift
            ;;
        -r|--report)
            REPORT_FORMAT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        -c|--config)
            CUSTOM_CONFIG="$2"
            shift 2
            ;;
        --rules)
            RULES_FILTER="$2"
            shift 2
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

log_verbose() {
    [[ "$VERBOSE" == "true" ]] && echo -e "${CYAN}🔍 $1${NC}"
}

# 📊 架构验证统计
declare -A VALIDATION_STATS
VALIDATION_STATS[total_files]=0
VALIDATION_STATS[passed_files]=0
VALIDATION_STATS[failed_files]=0
VALIDATION_STATS[total_issues]=0
VALIDATION_STATS[fixable_issues]=0
VALIDATION_STATS[architecture_violations]=0
VALIDATION_STATS[naming_violations]=0
VALIDATION_STATS[import_violations]=0

# 🚀 主函数
main() {
    log_info "开始Cube Castle架构守护验证..."
    log_info "配置: 范围=${SCOPE}, 修复=${FIX_ENABLED}, 报告=${REPORT_FORMAT}, 规则=${RULES_FILTER}"

    # 🔧 检查依赖
    if ! command -v npx >/dev/null 2>&1; then
        log_error "未找到npx命令，请确保Node.js已正确安装"
        exit 1
    fi

    # 📁 创建报告目录
    mkdir -p "$REPORT_DIR"
    cd "$PROJECT_ROOT"

    # 🎯 确定配置文件
    local config_file="$ESLINT_CONFIG"
    if [[ -n "$CUSTOM_CONFIG" ]]; then
        config_file="$CUSTOM_CONFIG"
    fi

    if [[ ! -f "$config_file" ]]; then
        log_error "ESLint配置文件不存在: $config_file"
        exit 1
    fi

    log_verbose "使用配置文件: $config_file"

    # 🎯 确定验证范围和目标
    local lint_targets=""
    local target_description=""

    case "$SCOPE" in
        "all")
            lint_targets="frontend/src cmd scripts"
            target_description="完整项目 (前端+后端+脚本)"
            ;;
        "frontend")
            lint_targets="frontend/src"
            target_description="前端代码 (React/TypeScript)"
            ;;
        "backend")
            lint_targets="cmd internal pkg"
            target_description="后端代码 (Go服务)"
            ;;
        "config")
            lint_targets="frontend/src/shared/config scripts/*/config*"
            target_description="配置管理模块"
            ;;
        *)
            log_error "无效的验证范围: $SCOPE"
            show_help
            exit 1
            ;;
    esac

    log_info "验证范围: $target_description"
    log_verbose "目标路径: $lint_targets"

    # 🔍 构建ESLint命令
    local eslint_cmd="npx eslint"
    
    # 添加配置文件
    eslint_cmd="$eslint_cmd --config $config_file"
    
    # 添加格式化选项
    if [[ "$REPORT_FORMAT" == "json" ]]; then
        eslint_cmd="$eslint_cmd --format json --output-file $REPORT_DIR/architecture-report.json"
    elif [[ "$REPORT_FORMAT" == "html" ]]; then
        eslint_cmd="$eslint_cmd --format html --output-file $REPORT_DIR/architecture-report.html"
    else
        eslint_cmd="$eslint_cmd --format compact"
    fi
    
    # 添加修复选项
    if [[ "$FIX_ENABLED" == "true" ]]; then
        eslint_cmd="$eslint_cmd --fix"
        log_info "🔧 自动修复模式已启用"
    fi

    # 添加规则过滤
    if [[ "$RULES_FILTER" != "all" ]]; then
        case "$RULES_FILTER" in
            "architecture")
                eslint_cmd="$eslint_cmd --rule 'architecture/*: error'"
                ;;
            "naming")
                eslint_cmd="$eslint_cmd --rule '@typescript-eslint/naming-convention: error'"
                ;;
            "imports")
                eslint_cmd="$eslint_cmd --rule 'no-restricted-imports: error'"
                ;;
        esac
        log_verbose "规则过滤: $RULES_FILTER"
    fi

    # 添加其他选项
    if [[ "$VERBOSE" == "true" ]]; then
        eslint_cmd="$eslint_cmd --debug"
    fi

    # 添加目标路径
    eslint_cmd="$eslint_cmd $lint_targets"

    log_verbose "ESLint命令: $eslint_cmd"
    log_info "开始架构验证..."

    # 🔍 执行ESLint验证
    local exit_code=0
    local output=""
    
    if [[ "$QUIET" == "true" ]]; then
        output=$(eval "$eslint_cmd" 2>&1) || exit_code=$?
    else
        eval "$eslint_cmd" || exit_code=$?
    fi

    # 📊 解析验证结果
    parse_validation_results "$output" "$exit_code"

    # 📈 生成统计报告
    generate_statistics_report

    # 🎯 质量门禁检查
    perform_quality_gate_check

    exit $exit_code
}

# 📊 解析验证结果
parse_validation_results() {
    local output="$1"
    local exit_code="$2"

    log_verbose "解析验证结果，退出码: $exit_code"

    # 基于输出解析统计信息
    if [[ -n "$output" ]]; then
        # 解析文件数量
        local file_count
        file_count=$(echo "$output" | grep -c "\.ts\|\.tsx\|\.js\|\.jsx" || echo "0")
        VALIDATION_STATS[total_files]=$file_count

        # 解析错误数量
        local error_count
        error_count=$(echo "$output" | grep -c "error" || echo "0")
        VALIDATION_STATS[total_issues]=$error_count

        # 解析架构违规
        local arch_violations
        arch_violations=$(echo "$output" | grep -c "architecture/" || echo "0")
        VALIDATION_STATS[architecture_violations]=$arch_violations

        # 解析命名违规
        local naming_violations
        naming_violations=$(echo "$output" | grep -c "naming-convention" || echo "0")
        VALIDATION_STATS[naming_violations]=$naming_violations

        # 解析导入违规
        local import_violations
        import_violations=$(echo "$output" | grep -c "no-restricted-imports" || echo "0")
        VALIDATION_STATS[import_violations]=$import_violations
    fi

    # 计算通过的文件数
    if [[ $exit_code -eq 0 ]]; then
        VALIDATION_STATS[passed_files]=${VALIDATION_STATS[total_files]}
        VALIDATION_STATS[failed_files]=0
    else
        local failed_files
        failed_files=$(echo "$output" | grep -c ":\s*error" || echo "1")
        VALIDATION_STATS[failed_files]=$failed_files
        VALIDATION_STATS[passed_files]=$((${VALIDATION_STATS[total_files]} - $failed_files))
    fi

    log_verbose "解析完成: ${VALIDATION_STATS[total_issues]} 个问题，${VALIDATION_STATS[failed_files]} 个文件失败"
}

# 📈 生成统计报告
generate_statistics_report() {
    log_info "📊 架构验证统计报告:"

    # 文件统计
    log_info "   📁 验证文件: ${VALIDATION_STATS[total_files]} 个"
    log_info "   ✅ 通过文件: ${VALIDATION_STATS[passed_files]} 个"
    if [[ ${VALIDATION_STATS[failed_files]} -gt 0 ]]; then
        log_warning "   ❌ 失败文件: ${VALIDATION_STATS[failed_files]} 个"
    fi

    # 问题分类统计
    log_info "   🔍 问题总数: ${VALIDATION_STATS[total_issues]} 个"
    if [[ ${VALIDATION_STATS[architecture_violations]} -gt 0 ]]; then
        log_warning "   🏗️  架构违规: ${VALIDATION_STATS[architecture_violations]} 个"
    fi
    if [[ ${VALIDATION_STATS[naming_violations]} -gt 0 ]]; then
        log_warning "   📝 命名违规: ${VALIDATION_STATS[naming_violations]} 个"
    fi
    if [[ ${VALIDATION_STATS[import_violations]} -gt 0 ]]; then
        log_warning "   📦 导入违规: ${VALIDATION_STATS[import_violations]} 个"
    fi

    # 修复统计
    if [[ "$FIX_ENABLED" == "true" ]]; then
        log_info "   🔧 自动修复: 已尝试修复所有可修复问题"
    fi

    # 报告文件位置
    if [[ "$REPORT_FORMAT" != "console" ]]; then
        log_info "📂 详细报告位置:"
        if [[ "$REPORT_FORMAT" == "json" ]]; then
            log_info "   JSON报告: $REPORT_DIR/architecture-report.json"
        elif [[ "$REPORT_FORMAT" == "html" ]]; then
            log_info "   HTML报告: $REPORT_DIR/architecture-report.html"
        fi
    fi
}

# 🎯 质量门禁检查
perform_quality_gate_check() {
    log_info "🚨 执行架构质量门禁检查..."

    local gate_failed=false
    local critical_violations=0

    # 关键架构违规检查
    if [[ ${VALIDATION_STATS[architecture_violations]} -gt 0 ]]; then
        log_error "质量门禁失败: 发现 ${VALIDATION_STATS[architecture_violations]} 个架构违规"
        gate_failed=true
        critical_violations=$((critical_violations + ${VALIDATION_STATS[architecture_violations]}))
    fi

    # 总问题数量检查
    if [[ ${VALIDATION_STATS[total_issues]} -gt 20 ]]; then
        log_error "质量门禁失败: 问题总数 ${VALIDATION_STATS[total_issues]} 超过阈值 20"
        gate_failed=true
    fi

    # 失败文件比例检查
    if [[ ${VALIDATION_STATS[total_files]} -gt 0 ]]; then
        local failure_rate
        failure_rate=$((${VALIDATION_STATS[failed_files]} * 100 / ${VALIDATION_STATS[total_files]}))
        if [[ $failure_rate -gt 30 ]]; then
            log_error "质量门禁失败: 文件失败率 ${failure_rate}% 超过阈值 30%"
            gate_failed=true
        fi
    fi

    # 结果判定
    if [[ "$gate_failed" == "true" ]]; then
        log_error "🚫 架构质量门禁失败！"
        log_error "   请修复架构违规后再次提交"
        if [[ "$FIX_ENABLED" != "true" ]]; then
            log_info "   建议运行: $0 --fix 自动修复可修复问题"
        fi
        return 1
    else
        log_success "🎉 架构质量门禁通过！"
        log_success "   代码架构符合企业级标准"
        if [[ ${VALIDATION_STATS[total_issues]} -gt 0 ]]; then
            log_info "   仍有 ${VALIDATION_STATS[total_issues]} 个非关键问题，建议优化"
        fi
        return 0
    fi
}

# 🎯 清理函数
cleanup() {
    log_verbose "清理临时文件..."
    # 这里可以添加清理临时文件的逻辑
}

# 注册清理函数
trap cleanup EXIT

# 🎯 程序入口
main "$@"