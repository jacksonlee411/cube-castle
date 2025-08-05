#!/bin/bash

echo "⚡ 执行性能基准测试..."
echo "========================="

API_URL="http://localhost:8080"

# 健康检查性能测试
echo "1. 健康检查API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$API_URL/health" > /dev/null
done
end_time=$(date +%s.%N)
health_avg=$(echo "scale=3; ($end_time - $start_time) / 10" | bc -l)
echo "   平均响应时间: ${health_avg}s"

# 组织列表性能测试
echo "2. 组织列表API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$API_URL/api/v1/organization-units" > /dev/null
done
end_time=$(date +%s.%N)
list_avg=$(echo "scale=3; ($end_time - $start_time) / 10" | bc -l)
echo "   平均响应时间: ${list_avg}s"

# 单个查询性能测试
echo "3. 单个查询API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$API_URL/api/v1/organization-units/1000000" > /dev/null
done
end_time=$(date +%s.%N)
single_avg=$(echo "scale=3; ($end_time - $start_time) / 10" | bc -l)
echo "   平均响应时间: ${single_avg}s"

# 统计API性能测试
echo "4. 统计API性能测试"
start_time=$(date +%s.%N)
for i in {1..10}; do
    curl -s "$API_URL/api/v1/organization-units/stats" > /dev/null
done
end_time=$(date +%s.%N)
stats_avg=$(echo "scale=3; ($end_time - $start_time) / 10" | bc -l)
echo "   平均响应时间: ${stats_avg}s"

# 生成基准报告
echo ""
echo "📊 性能基准报告 - $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================================"
echo "健康检查: ${health_avg}s (目标: <0.005s) $(if (( $(echo "$health_avg < 0.005" | bc -l) )); then echo "✅"; else echo "⚠️"; fi)"
echo "组织列表: ${list_avg}s (目标: <0.030s) $(if (( $(echo "$list_avg < 0.030" | bc -l) )); then echo "✅"; else echo "⚠️"; fi)"
echo "单个查询: ${single_avg}s (目标: <0.015s) $(if (( $(echo "$single_avg < 0.015" | bc -l) )); then echo "✅"; else echo "⚠️"; fi)"
echo "统计查询: ${stats_avg}s (目标: <0.050s) $(if (( $(echo "$stats_avg < 0.050" | bc -l) )); then echo "✅"; else echo "⚠️"; fi)"

# 保存基准数据
echo "$(date '+%Y-%m-%d %H:%M:%S'),health,$health_avg" >> performance/baseline.csv
echo "$(date '+%Y-%m-%d %H:%M:%S'),list,$list_avg" >> performance/baseline.csv
echo "$(date '+%Y-%m-%d %H:%M:%S'),single,$single_avg" >> performance/baseline.csv
echo "$(date '+%Y-%m-%d %H:%M:%S'),stats,$stats_avg" >> performance/baseline.csv

echo ""
echo "✅ 性能基准测试完成，数据已保存到 performance/baseline.csv"