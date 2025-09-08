#!/bin/bash

# 端到端测试执行脚本
# 用途: 重构后的组织架构模块完整性验证

set -e

echo "🚀 开始执行端到端测试套件..."
echo "=================================="

# 检查依赖服务状态
echo "📋 步骤 1: 检查服务状态"
check_service() {
    local service_name=$1
    local port=$2
    echo -n "检查 $service_name ($port端口)... "
    
    if curl -s http://localhost:$port/health > /dev/null 2>&1; then
        echo "✅ 正常"
        return 0
    else
        echo "❌ 不可用"
        return 1
    fi
}

# 检查核心服务
if ! check_service "查询服务" 8090; then
    echo "⚠️  查询服务未启动，尝试启动..."
    cd /home/shangmeilin/cube-castle
    ./start_optimized_services.sh &
    sleep 10
fi

if ! check_service "命令服务" 9090; then
    echo "⚠️  命令服务未启动，尝试启动..."
    # 命令服务应该已在上面的脚本中启动
    sleep 5
fi

echo ""
echo "📋 步骤 2: 启动前端开发服务器"
cd /home/shangmeilin/cube-castle/frontend

# 检查前端是否已启动
if curl -s http://localhost:3001 > /dev/null 2>&1; then
    echo "✅ 前端服务已启动"
else
    echo "🚀 启动前端服务..."
    npm run dev > frontend-dev.log 2>&1 &
    FRONTEND_PID=$!
    echo "前端进程ID: $FRONTEND_PID"
    
    # 等待前端启动
    echo "⏳ 等待前端服务启动..."
    for i in {1..30}; do
        if curl -s http://localhost:3001 > /dev/null 2>&1; then
            echo "✅ 前端服务已就绪"
            break
        fi
        echo -n "."
        sleep 2
    done
fi

echo ""
echo "📋 步骤 3: 执行测试套件"

# 定义测试套件
declare -A test_suites=(
    ["架构完整性验证"]="architecture-e2e.spec.ts"
    ["业务流程测试"]="business-flow-e2e.spec.ts" 
    ["优化效果验证"]="optimization-verification-e2e.spec.ts"
    ["回归兼容性测试"]="regression-e2e.spec.ts"
    ["Canvas UI测试"]="canvas-e2e.spec.ts"
    ["Schema验证测试"]="schema-validation.spec.ts"
)

# 执行测试结果统计
declare -A test_results

echo "开始执行 ${#test_suites[@]} 个测试套件..."
echo ""

# 执行每个测试套件
for test_name in "${!test_suites[@]}"; do
    test_file="${test_suites[$test_name]}"
    echo "🧪 执行: $test_name"
    echo "   文件: $test_file"
    
    if npx playwright test "tests/e2e/$test_file" --reporter=line; then
        test_results["$test_name"]="✅ PASSED"
        echo "   ✅ $test_name - 通过"
    else
        test_results["$test_name"]="❌ FAILED" 
        echo "   ❌ $test_name - 失败"
    fi
    echo ""
done

echo ""
echo "📊 测试结果汇总"
echo "=================================="

passed_count=0
total_count=${#test_suites[@]}

for test_name in "${!test_results[@]}"; do
    result="${test_results[$test_name]}"
    echo "$result $test_name"
    
    if [[ $result == *"PASSED"* ]]; then
        ((passed_count++))
    fi
done

echo ""
echo "📈 测试统计"
echo "总测试套件: $total_count"
echo "通过: $passed_count"
echo "失败: $((total_count - passed_count))"
echo "成功率: $(( passed_count * 100 / total_count ))%"

echo ""
echo "📋 步骤 4: 生成测试报告"
npx playwright show-report --host 0.0.0.0 --port 9323 &
REPORT_PID=$!

echo "📊 测试报告已生成，访问: http://localhost:9323"
echo "报告进程ID: $REPORT_PID"

echo ""
echo "📋 步骤 5: 清理和建议"

if [ $passed_count -eq $total_count ]; then
    echo "🎉 所有测试通过！组织架构模块重构验证完成。"
    echo ""
    echo "📋 后续建议:"
    echo "  1. ✅ 端到端测试覆盖完整"
    echo "  2. ✅ 架构优化效果已验证"
    echo "  3. ✅ 可以进入Phase 4生产部署准备"
    exit 0
else
    echo "⚠️  部分测试失败，需要进一步调试"
    echo ""
    echo "🔧 调试建议:"
    echo "  1. 查看测试报告详细信息"
    echo "  2. 检查失败的测试用例"
    echo "  3. 验证服务配置和数据状态"
    echo "  4. 修复问题后重新运行: ./run-e2e-tests.sh"
    exit 1
fi