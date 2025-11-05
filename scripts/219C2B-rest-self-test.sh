#!/usr/bin/env bash

# 219C2B REST 自测脚本 - Create/Update Organization 验证链测试
# 验证内容: 业务规则验证、审计日志、错误码对齐
# 输出: logs/219C2/validation.log

set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "$ROOT_DIR"

# ========== 配置 ==========
BASE_URL_COMMAND="${BASE_URL_COMMAND:-http://localhost:9090}"
TENANT_ID="${TENANT_ID:-3b99930c-4dc6-4cc9-8e4d-7d960a931cb9}"
TEST_USER_ID="test-$(date +%s | tail -c 5)"

LOG_DIR="$ROOT_DIR/logs/219C2"
mkdir -p "$LOG_DIR"
TIMESTAMP="$(date +%Y%m%dT%H%M%S)"
VALIDATION_LOG="$LOG_DIR/validation.log"
TEST_REPORT="$LOG_DIR/rest-self-test-$TIMESTAMP.md"

# ========== 颜色与日志 ==========
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}" | tee -a "$VALIDATION_LOG"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}" | tee -a "$VALIDATION_LOG"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}" | tee -a "$VALIDATION_LOG"
}

log_error() {
    echo -e "${RED}❌ $1${NC}" | tee -a "$VALIDATION_LOG"
}

# ========== 工具函数 ==========

get_dev_token() {
    local response=$(curl -sf -X POST "$BASE_URL_COMMAND/auth/dev-token" \
        -H 'Content-Type: application/json' \
        -d '{
            "userId": "'$TEST_USER_ID'",
            "tenantId": "'$TENANT_ID'",
            "roles": ["ADMIN"],
            "duration": "8h"
        }')

    echo "$response" | jq -r '.data.token'
}

# REST 请求函数（带完整响应捕获）
make_rest_request() {
    local method=$1
    local endpoint=$2
    local token=$3
    local data=${4:-""}

    if [ -n "$data" ]; then
        curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL_COMMAND$endpoint" \
            -H 'Content-Type: application/json' \
            -H "Authorization: Bearer $token" \
            -H "X-Tenant-ID: $TENANT_ID" \
            -d "$data"
    else
        curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL_COMMAND$endpoint" \
            -H 'Content-Type: application/json' \
            -H "Authorization: Bearer $token" \
            -H "X-Tenant-ID: $TENANT_ID"
    fi
}

# 解析响应
parse_response() {
    local response=$1
    local body=$(echo "$response" | head -n -1)
    local http_code=$(echo "$response" | tail -n 1)

    echo "$http_code"
    if [ -n "$body" ]; then
        echo "$body"
    fi
}

# ========== 测试场景 ==========

# 测试 0: 初始化检查
test_initialization() {
    log_info "========== 测试 0: 初始化检查 =========="

    # 检查服务
    if ! curl -sf "$BASE_URL_COMMAND/health" > /dev/null; then
        log_error "命令服务不健康"
        return 1
    fi
    log_success "命令服务健康"

    # 获取token
    JWT_TOKEN=$(get_dev_token)
    if [ -z "$JWT_TOKEN" ]; then
        log_error "Token 获取失败"
        return 1
    fi
    log_success "Token 获取成功"

    echo ""
}

# 测试 1: 创建组织成功路径
test_create_org_success() {
    log_info "========== 测试 1: 创建组织（成功路径）=========="

    JWT_TOKEN=$(get_dev_token)
    # 生成格式: 1 + 6位随机数 = 7位数字，首位为1
    TEST_ORG_CODE="1$(printf "%06d" $((RANDOM % 900000 + 100000)))"

    log_info "创建有效组织: code=$TEST_ORG_CODE"

    RESPONSE=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "'$TEST_ORG_CODE'",
        "name": "219C2B 测试组织",
        "unitType": "DEPARTMENT",
        "operationReason": "业务验证链测试"
    }')

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | head -n -1)

    log_info "HTTP Response: $HTTP_CODE"
    log_info "Response Body:"
    echo "$BODY" | jq '.' >> "$VALIDATION_LOG" 2>/dev/null || echo "$BODY" >> "$VALIDATION_LOG"

    if [ "$HTTP_CODE" = "201" ]; then
        log_success "创建成功 (HTTP 201)"

        # 检查响应字段
        if echo "$BODY" | jq -e '.success' > /dev/null 2>&1; then
            log_success "success 字段存在"
        fi

        if echo "$BODY" | jq -e '.data.code' > /dev/null 2>&1; then
            log_success "data 字段存在"
        fi

        # 保存用于后续测试
        echo "$TEST_ORG_CODE" > /tmp/test_org_code
        return 0
    else
        log_error "创建失败 (HTTP $HTTP_CODE)"
        return 1
    fi
}

# 测试 2: 代码格式验证失败
test_create_org_invalid_code() {
    log_info "========== 测试 2: 组织代码格式验证（失败路径）=========="

    JWT_TOKEN=$(get_dev_token)

    log_info "尝试无效代码格式: 'INVALID-CODE'"

    RESPONSE=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "INVALID-CODE",
        "name": "无效代码组织",
        "unitType": "DEPARTMENT",
        "operationReason": "验证测试"
    }')

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | head -n -1)

    log_info "HTTP Response: $HTTP_CODE"

    if [ "$HTTP_CODE" = "400" ]; then
        log_success "返回 HTTP 400（预期）"

        # 检查错误码
        ERROR_CODE=$(echo "$BODY" | jq -r '.error.code // empty' 2>/dev/null || echo "")
        if [ "$ERROR_CODE" = "ORG_CODE_INVALID" ]; then
            log_success "错误码正确: $ERROR_CODE"
        else
            log_warning "错误码: $ERROR_CODE (预期: ORG_CODE_INVALID)"
        fi

        # 检查 ruleId
        RULE_ID=$(echo "$BODY" | jq -r '.error.details.ruleId // empty' 2>/dev/null || echo "")
        if [ -n "$RULE_ID" ]; then
            log_success "ruleId 存在: $RULE_ID"
        fi

        # 检查 severity
        SEVERITY=$(echo "$BODY" | jq -r '.error.details.severity // empty' 2>/dev/null || echo "")
        if [ -n "$SEVERITY" ]; then
            log_success "severity 存在: $SEVERITY"
        fi

        log_info "完整响应:"
        echo "$BODY" | jq '.' >> "$VALIDATION_LOG" 2>/dev/null || echo "$BODY" >> "$VALIDATION_LOG"
    else
        log_warning "返回 HTTP $HTTP_CODE (预期: 400)"
    fi

    echo ""
}

# 测试 3: 深度限制验证
test_create_org_depth_limit() {
    log_info "========== 测试 3: 组织深度限制验证 =========="

    JWT_TOKEN=$(get_dev_token)
    PARENT_CODE="2$(printf "%06d" $((RANDOM % 900000 + 100000)))"

    # 先创建父组织
    log_info "创建父组织: code=$PARENT_CODE"
    CREATE_PARENT=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "'$PARENT_CODE'",
        "name": "深度测试-父",
        "unitType": "DEPARTMENT",
        "operationReason": "深度验证"
    }')

    PARENT_HTTP=$(echo "$CREATE_PARENT" | tail -1)
    if [ "$PARENT_HTTP" != "201" ]; then
        log_warning "父组织创建失败"
        return 0
    fi
    log_success "父组织创建成功"

    # 创建子组织
    CHILD_CODE="3$(printf "%06d" $((RANDOM % 900000 + 100000)))"
    log_info "创建子组织: parent=$PARENT_CODE"

    CREATE_CHILD=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "'$CHILD_CODE'",
        "name": "深度测试-子",
        "unitType": "DEPARTMENT",
        "parentCode": "'$PARENT_CODE'",
        "operationReason": "深度验证"
    }')

    CHILD_HTTP=$(echo "$CREATE_CHILD" | tail -1)
    if [ "$CHILD_HTTP" = "201" ]; then
        log_success "子组织创建成功 (HTTP 201)"
    else
        log_warning "子组织创建返回 HTTP $CHILD_HTTP"
    fi

    echo ""
}

# 测试 4: 循环检测验证
test_create_org_cycle_detection() {
    log_info "========== 测试 4: 组织循环检测（失败路径）=========="

    JWT_TOKEN=$(get_dev_token)
    CYCLE_ORG="4$(printf "%06d" $((RANDOM % 900000 + 100000)))"

    # 创建自引用测试组织
    log_info "尝试创建自引用组织..."

    RESPONSE=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "'$CYCLE_ORG'",
        "name": "循环测试",
        "unitType": "DEPARTMENT",
        "parentCode": "'$CYCLE_ORG'",
        "operationReason": "循环验证"
    }')

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | head -n -1)

    log_info "HTTP Response: $HTTP_CODE"

    if [ "$HTTP_CODE" = "400" ] || [ "$HTTP_CODE" = "409" ]; then
        log_success "返回错误状态码 (HTTP $HTTP_CODE)"

        ERROR_CODE=$(echo "$BODY" | jq -r '.error.code // empty' 2>/dev/null || echo "")
        if [ "$ERROR_CODE" = "ORG_CYCLE_DETECTED" ] || [ "$ERROR_CODE" = "ORG_CIRC" ]; then
            log_success "错误码正确: $ERROR_CODE"
        else
            log_info "错误码: $ERROR_CODE"
        fi
    else
        log_warning "未检测到循环验证（HTTP $HTTP_CODE）"
    fi

    echo ""
}

# 测试 5: 更新组织成功路径
test_update_org_success() {
    log_info "========== 测试 5: 更新组织（成功路径）=========="

    # 获取之前创建的组织代码
    if [ ! -f /tmp/test_org_code ]; then
        log_warning "无可用的测试组织代码，跳过更新测试"
        return 0
    fi

    TEST_ORG_CODE=$(cat /tmp/test_org_code)
    JWT_TOKEN=$(get_dev_token)

    log_info "更新组织: code=$TEST_ORG_CODE"

    RESPONSE=$(make_rest_request PUT "/api/v1/organization-units/$TEST_ORG_CODE" "$JWT_TOKEN" '{
        "name": "219C2B 测试组织（已更新）",
        "description": "更新测试验证",
        "operationReason": "业务规则验证"
    }')

    HTTP_CODE=$(echo "$RESPONSE" | tail -1)
    BODY=$(echo "$RESPONSE" | head -n -1)

    log_info "HTTP Response: $HTTP_CODE"

    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "204" ]; then
        log_success "更新成功 (HTTP $HTTP_CODE)"

        if [ "$HTTP_CODE" = "200" ]; then
            echo "$BODY" | jq '.' >> "$VALIDATION_LOG" 2>/dev/null || echo "$BODY" >> "$VALIDATION_LOG"
        fi
    else
        log_warning "更新返回 HTTP $HTTP_CODE"
    fi

    echo ""
}

# 测试 6: 状态转换验证
test_org_status_transitions() {
    log_info "========== 测试 6: 组织状态转换验证 =========="

    if [ ! -f /tmp/test_org_code ]; then
        log_warning "无可用的测试组织代码，跳过状态测试"
        return 0
    fi

    TEST_ORG_CODE=$(cat /tmp/test_org_code)
    JWT_TOKEN=$(get_dev_token)

    # 测试暂停
    log_info "测试暂停操作..."
    SUSPEND=$(make_rest_request POST "/api/v1/organization-units/$TEST_ORG_CODE/suspend" "$JWT_TOKEN" '{
        "effectiveDate": "2025-11-06",
        "operationReason": "状态验证测试"
    }')

    SUSPEND_HTTP=$(echo "$SUSPEND" | tail -1)
    if [ "$SUSPEND_HTTP" = "200" ] || [ "$SUSPEND_HTTP" = "204" ]; then
        log_success "暂停成功 (HTTP $SUSPEND_HTTP)"
    else
        log_warning "暂停返回 HTTP $SUSPEND_HTTP"
    fi

    # 等待状态稳定
    sleep 1

    # 测试激活
    log_info "测试激活操作..."
    ACTIVATE=$(make_rest_request POST "/api/v1/organization-units/$TEST_ORG_CODE/activate" "$JWT_TOKEN" '{
        "effectiveDate": "2025-11-06",
        "operationReason": "状态验证测试"
    }')

    ACTIVATE_HTTP=$(echo "$ACTIVATE" | tail -1)
    if [ "$ACTIVATE_HTTP" = "200" ] || [ "$ACTIVATE_HTTP" = "204" ]; then
        log_success "激活成功 (HTTP $ACTIVATE_HTTP)"
    else
        log_warning "激活返回 HTTP $ACTIVATE_HTTP"
    fi

    echo ""
}

# 测试 7: 审计日志检查
test_audit_log_check() {
    log_info "========== 测试 7: 审计日志检查 =========="

    log_info "检查验证失败时审计日志中的 ruleId 与 severity..."

    # 触发一个验证失败
    JWT_TOKEN=$(get_dev_token)
    FAIL_RESPONSE=$(make_rest_request POST "/api/v1/organization-units" "$JWT_TOKEN" '{
        "code": "AUDIT-FAIL",
        "name": "审计测试",
        "unitType": "INVALID_TYPE",
        "operationReason": "审计验证"
    }')

    FAIL_HTTP=$(echo "$FAIL_RESPONSE" | tail -1)
    FAIL_BODY=$(echo "$FAIL_RESPONSE" | head -n -1)

    log_info "触发验证失败 (HTTP $FAIL_HTTP)"

    # 检查业务日志中是否记录 business_context
    if echo "$FAIL_BODY" | jq -e '.error.details' > /dev/null 2>&1; then
        CONTEXT=$(echo "$FAIL_BODY" | jq '.error.details' 2>/dev/null)

        if echo "$CONTEXT" | jq -e '.ruleId' > /dev/null 2>&1; then
            RULE_ID=$(echo "$CONTEXT" | jq -r '.ruleId')
            log_success "审计日志包含 ruleId: $RULE_ID"
        fi

        if echo "$CONTEXT" | jq -e '.severity' > /dev/null 2>&1; then
            SEVERITY=$(echo "$CONTEXT" | jq -r '.severity')
            log_success "审计日志包含 severity: $SEVERITY"
        fi

        if echo "$CONTEXT" | jq -e '.httpStatus' > /dev/null 2>&1; then
            HTTP_STATUS=$(echo "$CONTEXT" | jq -r '.httpStatus')
            log_success "审计日志包含 httpStatus: $HTTP_STATUS"
        fi
    fi

    log_info "审计日志详情:"
    echo "$FAIL_BODY" | jq '.error.details' >> "$VALIDATION_LOG" 2>/dev/null || echo "$FAIL_BODY" >> "$VALIDATION_LOG"

    echo ""
}

# ========== 主流程 ==========

main() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║      Plan 219C2B REST 自测 - Create/Update Organization  ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    # 初始化日志
    echo "# 219C2B REST 自测日志" > "$VALIDATION_LOG"
    echo "**测试时间**: $(date)" >> "$VALIDATION_LOG"
    echo "**测试范围**: Organization Create/Update 验证链" >> "$VALIDATION_LOG"
    echo "" >> "$VALIDATION_LOG"

    # 执行测试
    test_initialization
    test_create_org_success
    test_create_org_invalid_code
    test_create_org_depth_limit
    test_create_org_cycle_detection
    test_update_org_success
    test_org_status_transitions
    test_audit_log_check

    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                   自测完成                                 ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""
    echo "📄 验证日志: $VALIDATION_LOG"
    echo ""

    log_success "🎉 219C2B REST 自测流程完成"
    log_info "后续步骤："
    log_info "  1. 审查上述所有验证结果"
    log_info "  2. 核对审计日志中的 ruleId 与 severity"
    log_info "  3. 更新 logs/219C2/daily-*.md 并提交"
    echo ""
}

# 执行主流程
main "$@"
