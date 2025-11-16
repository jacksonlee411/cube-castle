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
# 是否启用完整 Plan 255/250 门禁（提交即强制）
FULL_GATES=${FULL_GATES:-true}

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
            # 优先使用 --file-list（若当前 ESLint 版本不支持，则回退逐个文件传参）
            if npx eslint --config eslint.config.architecture.mjs --file-list "$temp_file_list" --quiet >/dev/null 2>&1; then
                log_success "ESLint架构规则检查通过"
                checks_passed=$((checks_passed + 1))
            else
                # 回退方式：按文件列表传参执行
                if npx eslint --config eslint.config.architecture.mjs $(cat "$temp_file_list") --quiet >/dev/null 2>&1; then
                    log_success "ESLint架构规则检查通过（兼容模式）"
                    checks_passed=$((checks_passed + 1))
                else
                    log_error "ESLint架构规则检查失败"
                    log_info "运行详细检查（兼容模式）: npx eslint --config eslint.config.architecture.mjs $(cat $temp_file_list)"
                    exit_code=1
                fi
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

    # ===== 05 计划：提交即强制的本地门禁（Plan 255 三件套 + Plan 250 快检）=====
    if [[ "$FULL_GATES" == "true" ]]; then
        log_info "执行 05 计划 - 提交即强制门禁（Plan 255 + Plan 250）..."
        # 依赖检查
        if [[ ! -d "$PROJECT_ROOT/node_modules" ]]; then
            log_error "root node_modules 缺失。请在仓库根目录执行: npm ci"
            return 1
        fi
        # 255-1: 静态架构验证器（CQRS/端口/禁用端点）
        log_info "255: architecture-validator (cqrs, ports, forbidden)"
        if node "$PROJECT_ROOT/scripts/quality/architecture-validator.js" --scope frontend --rule cqrs,ports,forbidden >/dev/null 2>&1; then
            log_success "architecture-validator 通过"
        else
            log_error "architecture-validator 失败（阻断提交）。运行: node scripts/quality/architecture-validator.js --scope frontend --rule cqrs,ports,forbidden"
            return 1
        fi
        # 255-2: 根级 ESLint 平面架构守卫
        log_info "255: ESLint architecture guard (flat config)"
        if npx eslint --no-warn-ignored -c "$PROJECT_ROOT/eslint.config.architecture.mjs" 'frontend/src/**/*.{ts,tsx}' >/dev/null 2>&1; then
            log_success "ESLint 架构守卫通过"
        else
            log_error "ESLint 架构守卫失败（阻断提交）。运行: npx eslint --no-warn-ignored -c eslint.config.architecture.mjs 'frontend/src/**/*.{ts,tsx}'"
            return 1
        fi
        # 255-3: Go 快速构建（阻断）
        log_info "255: Go build (backend quick compile)"
        if (cd "$PROJECT_ROOT" && go build ./... >/dev/null 2>&1); then
            log_success "Go build 通过"
        else
            log_error "Go build 失败（阻断提交）。运行: go build ./..."
            return 1
        fi
        # 255-4: golangci-lint 软门禁（非阻断）
        if command -v golangci-lint >/dev/null 2>&1; then
            log_info "255: golangci-lint (soft, depguard+tagliatelle)"
            golangci-lint run -c "$PROJECT_ROOT/scripts/quality/golangci-fast.yml" >/dev/null 2>&1 || log_warning "golangci-lint 报告问题（pre-commit 非阻断；CI 严格）"
        else
            log_info "255: golangci-lint 未安装，跳过（建议安装以获得本地提示）"
        fi

        # 250 快检（阻断）
        log_info "250: quick gates（本地阻断）"
        if bash "$PROJECT_ROOT/scripts/quality/gates-250-no-legacy-env.sh" >/dev/null 2>&1; then
            log_success "gate-250-no-legacy-env 通过"
        else
            log_error "gate-250-no-legacy-env 失败。请勿设置 ENABLE_LEGACY_DUAL_SERVICE=true"
            return 1
        fi
        if bash "$PROJECT_ROOT/scripts/quality/gates-250-single-binary.sh" >/dev/null 2>&1; then
            log_success "gate-250-single-binary 通过"
        else
            log_error "gate-250-single-binary 失败。确保 ./cmd 下仅 1 个非 legacy main；其它 main 添加 //go:build legacy"
            return 1
        fi
        if bash "$PROJECT_ROOT/scripts/quality/gates-250-no-8090-in-command.sh" >/dev/null 2>&1; then
            log_success "gate-250-no-8090-in-command 通过"
        else
            log_error "gate-250-no-8090-in-command 失败。移除 cmd/hrms-server/command/main.go 中 8090 字面量（改为读取 PORT 配置，并保留禁用判断）"
            return 1
        fi
    else
        log_info "已禁用 FULL_GATES。跳过 05 计划强制门禁（仅在此钩子配置中生效）。"
    fi
    
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
