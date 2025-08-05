#!/bin/bash

echo "🚀 组织单元API性能基准测试"
echo "==============================="

# 测试配置
BASE_URL="http://localhost:8080"
CONCURRENT_USERS=50
TOTAL_REQUESTS=1000

echo "📊 测试配置:"
echo "  - 基础URL: $BASE_URL"
echo "  - 并发用户: $CONCURRENT_USERS"
echo "  - 总请求数: $TOTAL_REQUESTS"
echo ""

# 检查服务器状态
echo "🔍 检查服务器状态..."
health_response=$(curl -s "$BASE_URL/health")
if [[ $? -eq 0 ]]; then
    echo "✅ 服务器运行正常"
    echo "   $health_response"
else
    echo "❌ 服务器连接失败"
    exit 1
fi
echo ""

# 性能测试函数
run_performance_test() {
    local endpoint="$1"
    local test_name="$2"
    
    echo "🧪 测试: $test_name"
    echo "   端点: $endpoint"
    
    # 使用curl进行并发测试
    start_time=$(date +%s.%N)
    
    # 创建临时文件来收集结果
    temp_file="/tmp/perf_test_$$"
    
    # 并发执行请求
    for i in $(seq 1 $CONCURRENT_USERS); do
        {
            for j in $(seq 1 $((TOTAL_REQUESTS / CONCURRENT_USERS))); do
                response_time=$(curl -w "%{time_total}" -s -o /dev/null "$BASE_URL$endpoint")
                echo "$response_time" >> "$temp_file"
            done
        } &
    done
    
    # 等待所有后台任务完成
    wait
    
    end_time=$(date +%s.%N)
    total_time=$(echo "$end_time - $start_time" | bc)
    
    # 计算统计数据
    if [[ -f "$temp_file" ]]; then
        response_times=($(cat "$temp_file"))
        requests_completed=${#response_times[@]}
        
        # 计算平均响应时间
        total_response_time=0
        for time in "${response_times[@]}"; do
            total_response_time=$(echo "$total_response_time + $time" | bc)
        done
        avg_response_time=$(echo "scale=3; $total_response_time / $requests_completed" | bc)
        
        # 计算RPS
        rps=$(echo "scale=2; $requests_completed / $total_time" | bc)
        
        # 排序响应时间数组计算P95
        IFS=$'\n' sorted_times=($(sort -n <<< "${response_times[*]}"))
        p95_index=$((requests_completed * 95 / 100))
        p95_time=${sorted_times[$p95_index]}
        
        echo "   ✅ 测试完成"
        echo "   📈 结果:"
        echo "      - 总请求数: $requests_completed"
        echo "      - 总耗时: ${total_time}s"
        echo "      - 平均响应时间: ${avg_response_time}s"
        echo "      - P95响应时间: ${p95_time}s"
        echo "      - QPS: $rps"
        
        # 清理临时文件
        rm -f "$temp_file"
    else
        echo "   ❌ 测试失败 - 无响应数据"
    fi
    echo ""
}

# 执行各种性能测试
echo "🎯 开始性能测试..."
echo ""

# 测试1: 健康检查端点
run_performance_test "/health" "健康检查API"

# 测试2: 组织单元列表
run_performance_test "/api/v1/organization-units" "组织单元列表API"

# 测试3: 单个组织单元查询
run_performance_test "/api/v1/organization-units/1000000" "单个组织单元查询API"

# 测试4: 统计API
run_performance_test "/api/v1/organization-units/stats" "统计信息API"

echo "🎉 性能测试完成！"
echo ""
echo "📝 测试总结:"
echo "   所有API端点均已完成性能基准测试"
echo "   详细结果请参考上述输出"
echo ""
echo "💡 性能优化建议:"
echo "   - 如果响应时间 > 100ms，考虑添加缓存"
echo "   - 如果QPS < 500，检查数据库查询优化"
echo "   - 监控P95响应时间，确保用户体验"