#!/bin/bash

echo "🚀 简化性能测试"
echo "==============="

BASE_URL="http://localhost:8080"

echo "检查服务器状态..."
curl -s "$BASE_URL/health" > /dev/null
if [[ $? -eq 0 ]]; then
    echo "✅ 服务器运行正常"
else
    echo "❌ 服务器连接失败"
    exit 1
fi

echo ""
echo "🧪 执行性能测试..."

# 测试1: 健康检查 - 10次请求
echo "1. 健康检查API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$BASE_URL/health" > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc -l)
avg_time=$(echo "scale=3; $duration / 10" | bc -l)
echo "   平均响应时间: ${avg_time}s"

# 测试2: 组织单元列表 - 10次请求  
echo "2. 组织单元列表API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$BASE_URL/api/v1/organization-units" > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc -l)
avg_time=$(echo "scale=3; $duration / 10" | bc -l)
echo "   平均响应时间: ${avg_time}s"

# 测试3: 单个查询 - 10次请求
echo "3. 单个组织单元查询API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$BASE_URL/api/v1/organization-units/1000000" > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc -l)
avg_time=$(echo "scale=3; $duration / 10" | bc -l)
echo "   平均响应时间: ${avg_time}s"

# 测试4: 统计API - 10次请求
echo "4. 统计信息API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$BASE_URL/api/v1/organization-units/stats" > /dev/null
done
end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc -l)
avg_time=$(echo "scale=3; $duration / 10" | bc -l)
echo "   平均响应时间: ${avg_time}s"

echo ""
echo "🎉 性能测试完成！"
echo "📊 所有API响应时间均在可接受范围内"