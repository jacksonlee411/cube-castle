#!/bin/bash

# Cube Castle 自动化测试框架 v2.0
# 支持多种数据库后端的测试执行

set -euo pipefail

# 颜色定义
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly NC='\033[0m' # No Color

# 全局变量
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
START_TIME=$(date +%s)

# 数据库配置
DB_TYPE="${TEST_DB_TYPE:-sqlite_memory}"
POSTGRES_TEST_URL="${TEST_DATABASE_URL:-postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable}"

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 打印标题
print_title() {
    echo
    print_message $BLUE "========================================="
    print_message $BLUE "  $1"
    print_message $BLUE "========================================="
    echo
}

# 打印数据库信息
print_database_info() {
    print_title "测试数据库配置"
    
    case $DB_TYPE in
        "sqlite_memory")
            print_message $GREEN "📊 数据库类型: SQLite Memory (默认)"
            print_message $GREEN "⚡ 性能级别: 最快"
            print_message $GREEN "🎯 适用场景: 单元测试、快速开发验证"
            ;;
        "sqlite"|"sqlite_file")
            print_message $YELLOW "📊 数据库类型: SQLite File"
            print_message $YELLOW "⚡ 性能级别: 快"
            print_message $YELLOW "🎯 适用场景: 本地调试、持久化测试"
            ;;
        "postgres"|"postgresql")
            print_message $PURPLE "📊 数据库类型: PostgreSQL Test"
            print_message $PURPLE "⚡ 性能级别: 中等 (与生产环境一致)"
            print_message $PURPLE "🎯 适用场景: 集成测试、生产环境验证"
            print_message $PURPLE "🔗 连接地址: ${POSTGRES_TEST_URL}"
            ;;
        *)
            print_message $RED "❌ 未知数据库类型: $DB_TYPE"
            exit 1
            ;;
    esac
    echo
}

# 检查数据库连接
check_database_connection() {
    print_message $CYAN "🔍 检查数据库连接..."
    
    case $DB_TYPE in
        "postgres"|"postgresql")
            if ! command -v psql &> /dev/null; then
                print_message $YELLOW "⚠️ psql 未找到，跳过PostgreSQL连接检查"
                return 0
            fi
            
            # 解析连接字符串获取主机和端口
            local host=$(echo $POSTGRES_TEST_URL | sed -n 's/.*@\([^:]*\):.*/\1/p')
            local port=$(echo $POSTGRES_TEST_URL | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')
            
            if pg_isready -h "${host:-localhost}" -p "${port:-5432}" -q; then
                print_message $GREEN "✅ PostgreSQL 连接正常"
            else
                print_message $RED "❌ PostgreSQL 连接失败"
                print_message $YELLOW "💡 请确保PostgreSQL服务运行并且测试数据库存在"
                print_message $YELLOW "   创建测试数据库: createdb cubecastle_test"
                exit 1
            fi
            ;;
        *)
            print_message $GREEN "✅ SQLite 无需连接检查"
            ;;
    esac
}

# 运行单个测试套件
run_test_suite() {
    local test_path=$1
    local test_name=$2
    local test_type=$3
    
    print_message $CYAN "🧪 运行 $test_name ..."
    
    # 设置环境变量
    export TEST_DB_TYPE=$DB_TYPE
    if [[ $DB_TYPE == "postgres" || $DB_TYPE == "postgresql" ]]; then
        export TEST_DATABASE_URL=$POSTGRES_TEST_URL
    fi
    
    # 运行测试并捕获输出
    local test_output
    local exit_code=0
    
    if test_output=$(go test -v -race -timeout=5m "./$test_path" 2>&1); then
        print_message $GREEN "✅ $test_name: 全部通过"
        
        # 统计通过的测试数量
        local passed_count=$(echo "$test_output" | grep -c "PASS:" 2>/dev/null || echo "0")
        
        # 确保变量是纯数字
        passed_count=$(echo "$passed_count" | tr -d '\n\r\t ')
        
        PASSED_TESTS=$((PASSED_TESTS + ${passed_count:-0}))
        TOTAL_TESTS=$((TOTAL_TESTS + ${passed_count:-0}))
        
        # 显示简要统计
        if [ "$passed_count" -gt 0 ]; then
            print_message $GREEN "   📊 通过: $passed_count 个测试"
        fi
    else
        exit_code=$?
        print_message $RED "❌ $test_name: 测试失败"
        
        # 统计失败和通过的测试
        local failed_count=$(echo "$test_output" | grep -c "FAIL:" 2>/dev/null || echo "0")
        local passed_count=$(echo "$test_output" | grep -c "PASS:" 2>/dev/null || echo "0")
        
        # 确保变量是纯数字
        failed_count=$(echo "$failed_count" | tr -d '\n\r\t ')
        passed_count=$(echo "$passed_count" | tr -d '\n\r\t ')
        
        FAILED_TESTS=$((FAILED_TESTS + ${failed_count:-0}))
        PASSED_TESTS=$((PASSED_TESTS + ${passed_count:-0}))
        TOTAL_TESTS=$((TOTAL_TESTS + ${failed_count:-0} + ${passed_count:-0}))
        
        # 显示详细错误信息
        print_message $RED "   💥 失败: $failed_count 个测试"
        if [ "$passed_count" -gt 0 ]; then
            print_message $GREEN "   ✅ 通过: $passed_count 个测试"
        fi
        
        # 显示失败详情
        echo
        print_message $YELLOW "失败详情:"
        echo "$test_output" | grep -A 5 -B 2 "FAIL:" || true
        echo
    fi
    
    return $exit_code
}

# 运行性能基准测试
run_benchmark_tests() {
    print_title "性能基准测试"
    
    if [ -d "internal/handler" ]; then
        print_message $CYAN "🚀 运行API处理器性能基准..."
        
        export TEST_DB_TYPE=$DB_TYPE
        if [[ $DB_TYPE == "postgres" || $DB_TYPE == "postgresql" ]]; then
            export TEST_DATABASE_URL=$POSTGRES_TEST_URL
        fi
        
        if go test -bench=. -benchmem "./internal/handler/..." 2>/dev/null; then
            print_message $GREEN "✅ 性能基准测试完成"
        else
            print_message $YELLOW "⚠️ 未找到性能基准测试"
        fi
    fi
}

# 生成测试报告
generate_test_report() {
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    print_title "测试报告摘要"
    
    print_message $BLUE "🗓️  测试时间: $(date)"
    print_message $BLUE "⏱️  执行时长: ${duration}秒"
    print_message $BLUE "📊 数据库类型: $DB_TYPE"
    echo
    
    print_message $BLUE "📈 测试统计:"
    print_message $GREEN "   ✅ 通过: $PASSED_TESTS"
    print_message $RED "   ❌ 失败: $FAILED_TESTS" 
    print_message $YELLOW "   ⏭️  跳过: $SKIPPED_TESTS"
    print_message $CYAN "   📊 总计: $TOTAL_TESTS"
    echo
    
    # 计算成功率
    if [ $TOTAL_TESTS -gt 0 ]; then
        local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        if [ $success_rate -ge 95 ]; then
            print_message $GREEN "🎉 测试成功率: ${success_rate}% (优秀)"
        elif [ $success_rate -ge 80 ]; then
            print_message $YELLOW "⚠️ 测试成功率: ${success_rate}% (良好)"
        else
            print_message $RED "💥 测试成功率: ${success_rate}% (需要改进)"
        fi
    fi
    
    echo
    if [ $FAILED_TESTS -eq 0 ]; then
        print_message $GREEN "🎊 所有测试通过！代码质量良好。"
    else
        print_message $RED "🚨 有 $FAILED_TESTS 个测试失败，请检查并修复。"
    fi
}

# 主测试执行函数
main() {
    print_title "Cube Castle 自动化测试框架 v2.0"
    
    # 显示数据库信息
    print_database_info
    
    # 检查数据库连接
    check_database_connection
    
    print_title "开始执行测试套件"
    
    # API处理器测试
    if [ -d "internal/handler" ]; then
        print_title "API处理器测试"
        run_test_suite "internal/handler" "API处理器测试" "unit" || true
    fi
    
    # 服务层测试
    if [ -d "internal/service" ]; then
        print_title "服务层测试"
        run_test_suite "internal/service" "服务层测试" "unit" || true
    fi
    
    # 数据库层测试
    if [ -d "internal/repository" ]; then
        print_title "数据库层测试"
        run_test_suite "internal/repository" "数据库层测试" "integration" || true
    fi
    
    # 中间件测试
    if [ -d "internal/middleware" ]; then
        print_title "中间件测试"
        run_test_suite "internal/middleware" "中间件测试" "unit" || true
    fi
    
    # 工作流测试
    if [ -d "internal/workflow" ]; then
        print_title "工作流测试"
        run_test_suite "internal/workflow" "Temporal工作流测试" "integration" || true
    fi
    
    # 运行性能基准测试
    run_benchmark_tests
    
    # 生成测试报告
    generate_test_report
    
    # 返回适当的退出码
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# 显示使用帮助
show_help() {
    echo "Cube Castle 自动化测试框架 v2.0"
    echo
    echo "用法: $0 [选项]"
    echo
    echo "环境变量:"
    echo "  TEST_DB_TYPE          测试数据库类型 (sqlite_memory|sqlite|postgresql)"
    echo "  TEST_DATABASE_URL     PostgreSQL测试数据库连接字符串"
    echo
    echo "示例:"
    echo "  # 使用SQLite内存数据库（默认）"
    echo "  $0"
    echo
    echo "  # 使用SQLite文件数据库"
    echo "  TEST_DB_TYPE=sqlite $0"
    echo
    echo "  # 使用PostgreSQL测试数据库"
    echo "  TEST_DB_TYPE=postgresql TEST_DATABASE_URL='postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable' $0"
    echo
    echo "选项:"
    echo "  -h, --help           显示此帮助信息"
    echo
}

# 处理命令行参数
case "${1:-}" in
    -h|--help)
        show_help
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac