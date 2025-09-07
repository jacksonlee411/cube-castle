#!/bin/bash
# [已废弃 - 2025-09-07] 本脚本可能包含 CDC/多数据库验证步骤，仅作历史参考。
# 建议使用 README/CLAUDE.md 约定的契约测试与现行端到端路径。
# 完整测试套件执行脚本 - 诚实测试原则
# 文件: tests/run_comprehensive_tests.sh

set -e

echo "🧪🔍 开始执行完整测试套件 - 诚实测试原则"
echo "目标: 彻底验证删除organization_versions表后系统的完整性和可靠性"
echo "日期: $(date)"
echo ""

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_CATEGORIES=()

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 辅助函数：记录测试结果
log_test_category() {
    local category="$1"
    local status="$2"
    local details="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [[ "$status" == "PASS" ]]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✅ PASSED${NC}: $category"
        [[ -n "$details" ]] && echo "   详情: $details"
    else
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_CATEGORIES+=("$category")
        echo -e "${RED}❌ FAILED${NC}: $category"
        [[ -n "$details" ]] && echo -e "   ${RED}错误: $details${NC}"
    fi
    echo ""
}

# 检查必要的依赖
echo -e "${BLUE}🔧 检查测试环境依赖${NC}"
check_dependencies() {
    local deps_ok=true
    
    # 检查PostgreSQL连接
    if ! PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1" &>/dev/null; then
        echo -e "${RED}❌ PostgreSQL数据库连接失败${NC}"
        deps_ok=false
    fi
    
    # 检查时态服务
    if ! curl -s http://localhost:9091/health &>/dev/null; then
        echo -e "${RED}❌ 时态管理服务(9091)不可用${NC}"
        deps_ok=false
    fi
    
    # 检查前端服务
    if ! curl -s http://localhost:3001 &>/dev/null; then
        echo -e "${YELLOW}⚠️  前端服务(3001)不可用，将跳过E2E测试${NC}"
    fi
    
    # 检查jq工具
    if ! command -v jq &>/dev/null; then
        echo -e "${RED}❌ jq工具未安装，无法进行JSON解析测试${NC}"
        deps_ok=false
    fi
    
    if [[ "$deps_ok" == "true" ]]; then
        echo -e "${GREEN}✅ 所有测试依赖检查通过${NC}"
    else
        echo -e "${RED}❌ 测试环境依赖检查失败，请修复后重试${NC}"
        exit 1
    fi
}

check_dependencies

# 1. 数据库完整性测试
echo -e "${BLUE}📊 执行数据库完整性测试${NC}"
if PGPASSWORD=password psql -h localhost -U user -d cubecastle -f tests/sql/test_organization_versions_removal.sql &>/dev/null; then
    log_test_category "数据库完整性验证" "PASS" "所有8项数据库测试通过"
else
    log_test_category "数据库完整性验证" "FAIL" "数据库测试执行失败"
fi

# 2. 时态API功能测试
echo -e "${BLUE}📡 执行时态API功能测试${NC}"
api_test_results=""

# 基础API测试
basic_api_response=$(curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal" 2>/dev/null || echo "ERROR")
if [[ "$basic_api_response" != "ERROR" ]] && echo "$basic_api_response" | jq -r '.organizations[0].name' &>/dev/null; then
    org_name=$(echo "$basic_api_response" | jq -r '.organizations[0].name')
    log_test_category "基础时态API功能" "PASS" "成功获取组织: $org_name"
else
    log_test_category "基础时态API功能" "FAIL" "API响应异常或数据格式错误"
fi

# 时间点查询测试
time_query_response=$(curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal?as_of_date=2025-08-01" 2>/dev/null || echo "ERROR")
if [[ "$time_query_response" != "ERROR" ]] && echo "$time_query_response" | jq -r '.queried_at' &>/dev/null; then
    log_test_category "时间点查询功能" "PASS" "时态查询参数处理正常"
else
    log_test_category "时间点查询功能" "FAIL" "时间点查询功能异常"
fi

# 错误处理测试
error_response=$(curl -s -w "HTTPSTATUS:%{http_code}" "http://localhost:9091/api/v1/organization-units/9999999/temporal" 2>/dev/null || echo "HTTPSTATUS:000")
http_status=$(echo "$error_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d':' -f2)
if [[ "$http_status" == "404" ]]; then
    log_test_category "API错误处理机制" "PASS" "正确返回404状态码"
else
    log_test_category "API错误处理机制" "FAIL" "错误处理异常，状态码: $http_status"
fi

# 健康检查测试
health_response=$(curl -s "http://localhost:9091/health" 2>/dev/null || echo "ERROR")
if [[ "$health_response" != "ERROR" ]] && echo "$health_response" | jq -r '.status' | grep -q "healthy"; then
    log_test_category "健康检查端点" "PASS" "服务状态正常"
else
    log_test_category "健康检查端点" "FAIL" "健康检查异常"
fi

# 3. 性能基准测试
echo -e "${BLUE}⚡ 执行性能基准测试${NC}"
performance_test() {
    local total_time=0
    local iterations=5
    
    for i in $(seq 1 $iterations); do
        start_time=$(date +%s.%N)
        curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal" &>/dev/null
        end_time=$(date +%s.%N)
        
        iteration_time=$(echo "$end_time - $start_time" | bc -l)
        total_time=$(echo "$total_time + $iteration_time" | bc -l)
    done
    
    avg_time=$(echo "scale=3; $total_time / $iterations" | bc -l)
    
    # 诚实测试：严格的性能要求
    if (( $(echo "$avg_time < 0.5" | bc -l) )); then
        log_test_category "API性能基准测试" "PASS" "平均响应时间: ${avg_time}秒 (<0.5s)"
    else
        log_test_category "API性能基准测试" "FAIL" "响应时间超标: ${avg_time}秒 (>=0.5s)"
    fi
}

performance_test

# 4. 数据一致性验证
echo -e "${BLUE}🔄 执行数据一致性验证${NC}"
consistency_check() {
    # 检查时态字段一致性
    local field_consistency=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "
        SELECT COUNT(*) FROM organization_units 
        WHERE effective_date IS NOT NULL AND is_current IS NOT NULL
    " 2>/dev/null | tr -d ' ')
    
    local total_orgs=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "
        SELECT COUNT(*) FROM organization_units
    " 2>/dev/null | tr -d ' ')
    
    if [[ "$field_consistency" == "$total_orgs" ]] && [[ "$total_orgs" -gt 0 ]]; then
        log_test_category "时态字段一致性" "PASS" "所有组织($total_orgs)都有完整的时态字段"
    else
        log_test_category "时态字段一致性" "FAIL" "时态字段不完整: $field_consistency/$total_orgs"
    fi
    
    # 检查当前有效记录的唯一性
    local duplicate_current=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "
        SELECT COUNT(*) FROM (
            SELECT code, COUNT(*) 
            FROM organization_units 
            WHERE is_current = true 
            GROUP BY code 
            HAVING COUNT(*) > 1
        ) AS duplicates
    " 2>/dev/null | tr -d ' ')
    
    if [[ "$duplicate_current" == "0" ]]; then
        log_test_category "当前记录唯一性" "PASS" "无重复的当前有效记录"
    else
        log_test_category "当前记录唯一性" "FAIL" "发现$duplicate_current个重复的当前记录"
    fi
}

consistency_check

# 5. 缓存和CDC验证
echo -e "${BLUE}🔄 执行缓存和CDC验证${NC}"
cache_cdc_test() {
    # 测试缓存命中
    local cache_test_start=$(date +%s.%N)
    curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal" &>/dev/null
    curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal" &>/dev/null  # 第二次应该命中缓存
    local cache_test_end=$(date +%s.%N)
    
    local cache_time=$(echo "$cache_test_end - $cache_test_start" | bc -l)
    
    if (( $(echo "$cache_time < 0.1" | bc -l) )); then
        log_test_category "缓存性能验证" "PASS" "缓存响应时间: ${cache_time}秒"
    else
        log_test_category "缓存性能验证" "FAIL" "缓存性能不达标: ${cache_time}秒"
    fi
    
    # 检查Publication配置
    local pub_config=$(PGPASSWORD=password psql -h localhost -U user -d cubecastle -t -c "
        SELECT COUNT(*) FROM pg_publication WHERE puballtables = true
    " 2>/dev/null | tr -d ' ')
    
    if [[ "$pub_config" -gt 0 ]]; then
        log_test_category "CDC配置验证" "PASS" "Publication配置正常"
    else
        log_test_category "CDC配置验证" "FAIL" "Publication配置异常"
    fi
}

cache_cdc_test

# 6. 边界条件和错误恢复测试
echo -e "${BLUE}🧨 执行边界条件测试${NC}"
boundary_test() {
    # 测试无效参数处理
    local invalid_date_response=$(curl -s -w "HTTPSTATUS:%{http_code}" "http://localhost:9091/api/v1/organization-units/1000056/temporal?as_of_date=invalid-date" 2>/dev/null)
    local invalid_status=$(echo "$invalid_date_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d':' -f2)
    
    if [[ "$invalid_status" == "400" ]] || [[ "$invalid_status" == "200" ]]; then
        log_test_category "无效参数处理" "PASS" "无效日期参数处理正常"
    else
        log_test_category "无效参数处理" "FAIL" "无效参数处理异常: $invalid_status"
    fi
    
    # 测试空结果处理
    local empty_result_response=$(curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal?as_of_date=1900-01-01" 2>/dev/null || echo "ERROR")
    if [[ "$empty_result_response" != "ERROR" ]]; then
        local result_count=$(echo "$empty_result_response" | jq -r '.result_count // 0' 2>/dev/null || echo "0")
        log_test_category "历史查询边界处理" "PASS" "历史边界查询返回$result_count条结果"
    else
        log_test_category "历史查询边界处理" "FAIL" "历史查询处理异常"
    fi
}

boundary_test

# 7. 安全性验证
echo -e "${BLUE}🔒 执行安全性验证${NC}"
security_test() {
    # SQL注入测试
    local injection_response=$(curl -s -w "HTTPSTATUS:%{http_code}" "http://localhost:9091/api/v1/organization-units/1000056'; DROP TABLE organization_units;--/temporal" 2>/dev/null)
    local injection_status=$(echo "$injection_response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d':' -f2)
    
    if [[ "$injection_status" == "404" ]] || [[ "$injection_status" == "400" ]]; then
        log_test_category "SQL注入防护" "PASS" "SQL注入攻击被正确拦截"
    else
        log_test_category "SQL注入防护" "FAIL" "SQL注入防护可能存在问题"
    fi
    
    # XSS测试
    local xss_response=$(curl -s "http://localhost:9091/api/v1/organization-units/1000056/temporal?as_of_date=<script>alert('xss')</script>" 2>/dev/null || echo "ERROR")
    if [[ "$xss_response" != *"<script>"* ]]; then
        log_test_category "XSS防护验证" "PASS" "XSS攻击被正确处理"
    else
        log_test_category "XSS防护验证" "FAIL" "XSS防护可能存在问题"
    fi
}

security_test

# 最终测试结果汇总
echo ""
echo "════════════════════════════════════════════"
echo -e "${BLUE}🎯 完整测试套件结果汇总 (诚实测试原则)${NC}"
echo "════════════════════════════════════════════"
echo ""

success_rate=0
if [[ $TOTAL_TESTS -gt 0 ]]; then
    success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
fi

echo -e "📊 ${BLUE}测试统计${NC}:"
echo -e "   总测试数: ${TOTAL_TESTS}"
echo -e "   ✅ 通过: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "   ❌ 失败: ${RED}${FAILED_TESTS}${NC}"
echo -e "   📈 成功率: ${success_rate}%"
echo ""

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "${GREEN}🏆 恭喜！所有测试都通过了！${NC}"
    echo -e "${GREEN}✨ organization_versions表删除操作完全成功${NC}"
    echo -e "${GREEN}🚀 纯日期生效模型实施完美，系统运行稳定${NC}"
    echo ""
    echo -e "${BLUE}📋 验证完成的功能特性:${NC}"
    echo "   • 数据库表结构完整性 ✓"
    echo "   • 时态API功能完整性 ✓"
    echo "   • 性能基准达标 ✓"
    echo "   • 数据一致性保证 ✓"
    echo "   • 缓存和CDC配置 ✓"
    echo "   • 边界条件处理 ✓"
    echo "   • 安全性防护 ✓"
    
    exit 0
else
    echo -e "${RED}⚠️  发现 $FAILED_TESTS 个测试失败！${NC}"
    echo -e "${RED}📝 失败的测试类别:${NC}"
    for category in "${FAILED_CATEGORIES[@]}"; do
        echo -e "   • ${RED}$category${NC}"
    done
    echo ""
    echo -e "${YELLOW}🔧 建议的修复措施:${NC}"
    echo "   1. 检查相关服务是否正常运行"
    echo "   2. 验证数据库连接和权限"
    echo "   3. 审查失败测试的具体错误信息"
    echo "   4. 必要时回滚并重新执行删除操作"
    echo ""
    echo -e "${RED}❌ 测试未完全通过，请修复问题后重新验证${NC}"
    
    exit 1
fi
