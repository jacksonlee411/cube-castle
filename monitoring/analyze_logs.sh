#!/bin/bash

LOG_FILE="monitoring/logs/performance.log"

if [ ! -f "$LOG_FILE" ]; then
    echo "❌ 日志文件不存在: $LOG_FILE"
    exit 1
fi

echo "📊 性能分析报告 - $(date)"
echo "=================================="

# 统计总请求数
total_requests=$(wc -l < $LOG_FILE)
echo "总请求数: $total_requests"

# 统计成功率
success_count=$(grep ",200," $LOG_FILE | wc -l)
success_rate=$(echo "scale=2; $success_count * 100 / $total_requests" | bc -l)
echo "成功率: ${success_rate}%"

# 计算平均响应时间
avg_health_time=$(grep ",health," $LOG_FILE | grep ",200," | cut -d',' -f4 | sed 's/s$//' | awk '{sum+=$1} END {printf "%.3f", sum/NR}')
avg_api_time=$(grep ",api_list," $LOG_FILE | grep ",200," | cut -d',' -f4 | sed 's/s$//' | awk '{sum+=$1} END {printf "%.3f", sum/NR}')

echo "健康检查平均响应时间: ${avg_health_time}s"
echo "API列表平均响应时间: ${avg_api_time}s"

# 显示最近10条记录
echo ""
echo "📋 最近10条记录:"
tail -10 $LOG_FILE | while IFS=',' read -r timestamp endpoint status time; do
    printf "%-20s %-10s %-3s %s\n" "$timestamp" "$endpoint" "$status" "$time"
done
