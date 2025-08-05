#!/bin/bash

API_URL="http://localhost:8080"
LOG_FILE="monitoring/logs/performance.log"

echo "$(date): 开始性能监控..." >> $LOG_FILE

while true; do
    # 健康检查
    start_time=$(date +%s.%N)
    health_response=$(curl -s -w "%{http_code}" -o /dev/null $API_URL/health)
    end_time=$(date +%s.%N)
    health_time=$(echo "$end_time - $start_time" | bc -l)
    
    # API测试
    start_time=$(date +%s.%N)
    api_response=$(curl -s -w "%{http_code}" -o /dev/null $API_URL/api/v1/organization-units)
    end_time=$(date +%s.%N)
    api_time=$(echo "$end_time - $start_time" | bc -l)
    
    # 记录日志
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp,health,$health_response,${health_time}s" >> $LOG_FILE
    echo "$timestamp,api_list,$api_response,${api_time}s" >> $LOG_FILE
    
    # 显示实时状态
    printf "\r⚡ 健康检查: ${health_response} (${health_time}s) | API列表: ${api_response} (${api_time}s) | $(date '+%H:%M:%S')"
    
    sleep 10
done
