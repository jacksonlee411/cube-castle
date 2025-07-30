#!/bin/bash

# 员工模型管理系统 - 全套测试执行脚本
# Employee Model Management System - Complete Test Execution Script

set -e

echo "🚀 开始执行员工模型管理系统完整测试套件..."
echo "Starting complete test suite for Employee Model Management System..."
echo "======================================================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 记录测试结果
log_result() {
    local test_name="$1"
    local result="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name - PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ $test_name - FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}🔍 检查测试依赖...${NC}"
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go 未安装${NC}"
        exit 1
    fi
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js 未安装${NC}"
        exit 1
    fi
    
    # 检查 Docker (for test databases)
    if ! command -v docker &> /dev/null; then
        echo -e "${YELLOW}⚠️  Docker 未安装，将跳过需要 Docker 的集成测试${NC}"
    fi
    
    echo -e "${GREEN}✅ 依赖检查完成${NC}"
}

# 启动测试数据库
start_test_databases() {
    echo -e "${BLUE}🗄️  启动测试数据库...${NC}"
    
    if command -v docker &> /dev/null; then
        # 启动 PostgreSQL 测试数据库
        docker run -d --name postgres-test \
            -e POSTGRES_DB=employee_model_test \
            -e POSTGRES_USER=test \
            -e POSTGRES_PASSWORD=test \
            -p 5433:5432 \
            postgres:15 > /dev/null 2>&1 || true
        
        # 启动 Neo4j 测试数据库
        docker run -d --name neo4j-test \
            -e NEO4J_AUTH=neo4j/testpass \
            -p 7475:7474 -p 7688:7687 \
            neo4j:5 > /dev/null 2>&1 || true
        
        # 等待数据库启动
        echo "等待数据库启动..."
        sleep 10
        
        echo -e "${GREEN}✅ 测试数据库启动完成${NC}"
    else
        echo -e "${YELLOW}⚠️  跳过数据库启动 (Docker 不可用)${NC}"
    fi
}

# 停止测试数据库
stop_test_databases() {
    echo -e "${BLUE}🛑 停止测试数据库...${NC}"
    
    if command -v docker &> /dev/null; then
        docker stop postgres-test neo4j-test > /dev/null 2>&1 || true
        docker rm postgres-test neo4j-test > /dev/null 2>&1 || true
        echo -e "${GREEN}✅ 测试数据库清理完成${NC}"
    fi
}

# 后端单元测试
run_backend_unit_tests() {
    echo -e "${BLUE}🧪 执行后端单元测试...${NC}"
    cd /home/shangmeilin/cube-castle/go-app
    
    # 设置测试环境变量
    export GO_ENV=test
    export DATABASE_URL="postgres://test:test@localhost:5433/employee_model_test?sslmode=disable"
    export NEO4J_URI="bolt://localhost:7688"
    export NEO4J_USERNAME="neo4j"
    export NEO4J_PASSWORD="testpass"
    
    # 运行所有单元测试
    echo "运行 TemporalQueryService 测试..."
    if go test -v ./internal/service -run TestTemporalQueryService > /dev/null 2>&1; then
        log_result "TemporalQueryService 单元测试" "PASS"
    else
        log_result "TemporalQueryService 单元测试" "FAIL"
    fi
    
    echo "运行 Neo4jService 测试..."
    if go test -v ./internal/service -run TestNeo4jService > /dev/null 2>&1; then
        log_result "Neo4jService 单元测试" "PASS"
    else
        log_result "Neo4jService 单元测试" "FAIL"
    fi
    
    echo "运行 SAMService 测试..."
    if go test -v ./internal/service -run TestSAMService > /dev/null 2>&1; then
        log_result "SAMService 单元测试" "PASS"
    else
        log_result "SAMService 单元测试" "FAIL"
    fi
    
    echo "运行 GraphQL Resolvers 测试..."
    if go test -v ./internal/graphql/resolvers > /dev/null 2>&1; then
        log_result "GraphQL Resolvers 单元测试" "PASS"
    else
        log_result "GraphQL Resolvers 单元测试" "FAIL"
    fi
    
    # 生成覆盖率报告
    echo "生成后端测试覆盖率报告..."
    go test -coverprofile=coverage.out ./... > /dev/null 2>&1
    go tool cover -html=coverage.out -o coverage.html > /dev/null 2>&1
    
    echo -e "${GREEN}✅ 后端单元测试完成${NC}"
}

# 后端集成测试
run_backend_integration_tests() {
    echo -e "${BLUE}🔗 执行后端集成测试...${NC}"
    cd /home/shangmeilin/cube-castle/go-app
    
    echo "运行 Temporal 工作流集成测试..."
    if go test -v ./test/integration -run TestTemporalWorkflow > /dev/null 2>&1; then
        log_result "Temporal 工作流集成测试" "PASS"
    else
        log_result "Temporal 工作流集成测试" "FAIL"
    fi
    
    echo "运行数据库集成测试..."
    if go test -v ./test/integration -run TestDatabase > /dev/null 2>&1; then
        log_result "数据库集成测试" "PASS"
    else
        log_result "数据库集成测试" "FAIL"
    fi
    
    echo "运行微服务通信集成测试..."
    if go test -v ./test/integration -run TestMicroservices > /dev/null 2>&1; then
        log_result "微服务通信集成测试" "PASS"
    else
        log_result "微服务通信集成测试" "FAIL"
    fi
    
    echo -e "${GREEN}✅ 后端集成测试完成${NC}"
}

# 前端测试
run_frontend_tests() {
    echo -e "${BLUE}⚛️  执行前端测试...${NC}"
    cd /home/shangmeilin/cube-castle/nextjs-app
    
    # 安装依赖 (如果需要)
    if [ ! -d "node_modules" ]; then
        echo "安装前端依赖..."
        npm install > /dev/null 2>&1
    fi
    
    echo "运行 React 组件单元测试..."
    if npm run test:unit > /dev/null 2>&1; then
        log_result "React 组件单元测试" "PASS"
    else
        log_result "React 组件单元测试" "FAIL"
    fi
    
    echo "运行端到端测试..."
    if npm run test:e2e > /dev/null 2>&1; then
        log_result "端到端测试" "PASS"
    else
        log_result "端到端测试" "FAIL"
    fi
    
    echo -e "${GREEN}✅ 前端测试完成${NC}"
}

# 性能测试
run_performance_tests() {
    echo -e "${BLUE}⚡ 执行性能测试...${NC}"
    cd /home/shangmeilin/cube-castle/go-app
    
    echo "运行性能基准测试..."
    if go test -bench=. -benchmem ./internal/service > /dev/null 2>&1; then
        log_result "性能基准测试" "PASS"
    else
        log_result "性能基准测试" "FAIL"
    fi
    
    echo -e "${GREEN}✅ 性能测试完成${NC}"
}

# 生成测试报告
generate_test_report() {
    echo -e "${BLUE}📊 生成测试报告...${NC}"
    
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    
    cat > test-execution-summary.txt << EOF
======================================================================
员工模型管理系统 - 测试执行总结
Employee Model Management System - Test Execution Summary
======================================================================

执行时间: $timestamp
测试环境: $(uname -s) $(uname -r)

测试结果统计:
- 总测试数: $TOTAL_TESTS
- 通过数量: $PASSED_TESTS
- 失败数量: $FAILED_TESTS
- 成功率: $success_rate%

测试分类:
✅ 后端单元测试 (4 个测试套件)
✅ 后端集成测试 (3 个测试套件)
✅ 前端测试 (2 个测试套件)
✅ 性能测试 (1 个测试套件)

详细报告: 请查看 TEST_REPORT.md

======================================================================
EOF
    
    echo -e "${GREEN}✅ 测试报告生成完成${NC}"
}

# 清理函数
cleanup() {
    echo -e "${BLUE}🧹 清理测试环境...${NC}"
    stop_test_databases
    echo -e "${GREEN}✅ 清理完成${NC}"
}

# 主执行流程
main() {
    # 设置清理陷阱
    trap cleanup EXIT
    
    echo -e "${BLUE}开始时间: $(date)${NC}"
    
    # 执行测试流程
    check_dependencies
    start_test_databases
    run_backend_unit_tests
    run_backend_integration_tests
    run_frontend_tests
    run_performance_tests
    generate_test_report
    
    # 输出最终结果
    echo ""
    echo "======================================================================="
    echo -e "${BLUE}📊 测试执行完成总结${NC}"
    echo "======================================================================="
    echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "通过数量: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败数量: ${RED}$FAILED_TESTS${NC}"
    
    local success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    if [ $success_rate -ge 90 ]; then
        echo -e "成功率: ${GREEN}$success_rate%${NC} 🎉"
        echo -e "${GREEN}🏆 测试结果: 优秀 - 系统生产就绪!${NC}"
    elif [ $success_rate -ge 80 ]; then
        echo -e "成功率: ${YELLOW}$success_rate%${NC} ⚠️"
        echo -e "${YELLOW}⚠️  测试结果: 良好 - 建议修复失败测试后发布${NC}"
    else
        echo -e "成功率: ${RED}$success_rate%${NC} ❌"
        echo -e "${RED}❌ 测试结果: 需改进 - 请修复关键问题${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}结束时间: $(date)${NC}"
    echo -e "${BLUE}详细报告: TEST_REPORT.md${NC}"
    echo -e "${BLUE}执行摘要: test-execution-summary.txt${NC}"
    
    # 返回适当的退出码
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# 执行主函数
main "$@"