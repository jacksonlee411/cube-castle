#!/bin/bash

# 🏗️ Cube Castle - Pre-commit架构守护Hook
# 用途: 在Git提交前验证架构一致性，确保代码质量
# 作者: Claude Code Assistant
# 日期: 2025-09-07
# 集成: 与现有pre-commit hook协同工作

set -euo pipefail

# 🎨 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 📊 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 🔍 检查是否为架构守护Hook调用
ARCH_GUARD_MODE=${ARCH_GUARD_MODE:-false}

# 📋 日志函数
log_info() {
    echo -e "${BLUE}🏗️  架构守护: $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ 架构守护: $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  架构守护: $1${NC}"
}

log_error() {
    echo -e "${RED}❌ 架构守护: $1${NC}" >&2
}

# 🚀 主要架构检查函数
run_architecture_checks() {
    log_info "开始Pre-commit架构一致性检查..."

    local exit_code=0
    local checks_passed=0
    local checks_total=5

    # 📋 检查1: 获取变更文件
    log_info "检查变更文件范围..."
    local changed_files
    changed_files=$(git diff --cached --name-only --diff-filter=ACM || true)
    
    if [[ -z "$changed_files" ]]; then
        log_warning "未检测到变更文件，跳过架构检查"
        return 0
    fi

    local ts_files
    ts_files=$(echo "$changed_files" | grep -E '\.(ts|tsx)$' || true)
    local js_files  
    js_files=$(echo "$changed_files" | grep -E '\.(js|jsx)$' || true)
    local config_files
    config_files=$(echo "$changed_files" | grep -E '(config|\.config\.|vite\.config|eslint)' || true)

    log_info "变更文件: TS/TSX($(echo "$ts_files" | wc -w)), JS/JSX($(echo "$js_files" | wc -w)), Config($(echo "$config_files" | wc -w))"

    # 📋 检查2: 端口配置一致性
    log_info "检查端口配置一致性..."
    if echo "$changed_files" | grep -q -E '(config|vite\.config|\.config\.)'; then
        if bash "$PROJECT_ROOT/scripts/ci/check-hardcoded-configs.sh" >/dev/null 2>&1; then
            log_success "端口配置检查通过"
            checks_passed=$((checks_passed + 1))
        else
            log_error "发现硬编码端口配置问题"
            exit_code=1
        fi
    else
        log_info "未变更配置文件，跳过端口检查"
        checks_passed=$((checks_passed + 1))
    fi

    # 📋 检查3: REST查询守护
    log_info "检查CQRS架构一致性..."
    if [[ -n "$ts_files" || -n "$js_files" ]]; then
        if bash "$PROJECT_ROOT/scripts/ci/check-rest-queries.sh" >/dev/null 2>&1; then
            log_success "CQRS架构检查通过"
            checks_passed=$((checks_passed + 1))
        else
            log_error "发现REST查询违规问题"
            exit_code=1
        fi
    else
        log_info "未变更前端文件，跳过CQRS检查"
        checks_passed=$((checks_passed + 1))
    fi

    # 📋 检查4: 权限命名一致性
    log_info "检查权限命名一致性..."
    if echo "$changed_files" | grep -q -E '(auth|permission|role)'; then
        if bash "$PROJECT_ROOT/scripts/ci/check-permissions.sh" >/dev/null 2>&1; then
            log_success "权限命名检查通过"
            checks_passed=$((checks_passed + 1))
        else
            log_error "发现权限命名问题"
            exit_code=1
        fi
    else
        log_info "未变更权限相关文件，跳过权限检查"
        checks_passed=$((checks_passed + 1))
    fi

    # 📋 检查5: ESLint架构规则（仅对变更文件）
    log_info "运行ESLint架构规则检查..."
    if [[ -n "$ts_files" || -n "$js_files" ]]; then
        cd "$PROJECT_ROOT"
        
        # 创建临时文件列表
        local temp_file_list="/tmp/eslint-changed-files.txt"
        echo "$changed_files" | grep -E '\.(ts|tsx|js|jsx)$' > "$temp_file_list" || true
        
        if [[ -s "$temp_file_list" ]]; then
            if npx eslint --config .eslintrc.architecture.js --file-list "$temp_file_list" --quiet >/dev/null 2>&1; then
                log_success "ESLint架构规则检查通过"
                checks_passed=$((checks_passed + 1))
            else
                log_error "ESLint架构规则检查失败"
                log_info "运行详细检查: npx eslint --config .eslintrc.architecture.js --file-list $temp_file_list"
                exit_code=1
            fi
            rm -f "$temp_file_list"
        else
            log_info "无有效的JS/TS文件变更，跳过ESLint检查"
            checks_passed=$((checks_passed + 1))
        fi
    else
        log_info "未变更JS/TS文件，跳过ESLint检查"
        checks_passed=$((checks_passed + 1))
    fi

    # 📊 输出检查结果摘要
    log_info "架构检查完成: $checks_passed/$checks_total 项通过"
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "所有架构一致性检查通过！"
        log_info "代码符合企业级架构标准，可以安全提交"
    else
        log_error "架构一致性检查失败！"
        log_error "请修复上述问题后重新提交"
        echo ""
        log_info "🔧 快速修复建议:"
        log_info "   • 端口配置: 使用 SERVICE_PORTS 配置模块"
        log_info "   • CQRS架构: 查询用GraphQL，命令用REST"
        log_info "   • 字段命名: 使用camelCase格式"
        log_info "   • 运行修复: bash scripts/quality/architecture-guard.sh --fix"
    fi

    return $exit_code
}

# 🎯 集成检查：避免与现有pre-commit冲突
check_integration() {
    # 检查是否存在其他pre-commit hooks
    local git_hooks_dir="$PROJECT_ROOT/.git/hooks"
    local existing_hook="$git_hooks_dir/pre-commit"
    
    if [[ -f "$existing_hook" && "$ARCH_GUARD_MODE" != "true" ]]; then
        # 如果存在其他hook，以集成模式运行
        log_info "检测到现有pre-commit hook，以集成模式运行..."
        export ARCH_GUARD_MODE=true
        
        # 只运行架构检查，不干扰其他检查
        run_architecture_checks
        local arch_result=$?
        
        if [[ $arch_result -ne 0 ]]; then
            log_error "架构守护检查失败，阻止提交"
            exit 1
        fi
        
        # 让现有hook继续执行
        log_success "架构检查通过，继续其他pre-commit检查..."
        return 0
    fi
    
    # 独立模式运行完整检查
    run_architecture_checks
}

# 🎯 主程序入口
main() {
    # 检查Git暂存区
    if ! git diff --cached --quiet; then
        check_integration
    else
        log_warning "暂存区为空，跳过架构检查"
        exit 0
    fi
}

# 只有直接运行时才执行main函数
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi