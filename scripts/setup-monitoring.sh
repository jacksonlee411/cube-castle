#!/bin/bash

echo "🔧 设置简化监控系统..."

# 创建监控目录
mkdir -p monitoring/logs

# 创建简单的性能监控脚本
cat > monitoring/performance_monitor.sh << 'EOF'
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
EOF

chmod +x monitoring/performance_monitor.sh

# 创建日志分析脚本
cat > monitoring/analyze_logs.sh << 'EOF'
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
EOF

chmod +x monitoring/analyze_logs.sh

# 启动后台监控
echo "🚀 启动后台性能监控..."
nohup ./monitoring/performance_monitor.sh > /dev/null 2>&1 &
monitor_pid=$!
echo $monitor_pid > monitoring/monitor.pid

sleep 3

echo "✅ 简化监控系统设置完成！"
echo ""
echo "📊 监控信息:"
echo "  监控PID: $monitor_pid"
echo "  日志文件: monitoring/logs/performance.log"
echo "  分析脚本: ./monitoring/analyze_logs.sh"
echo ""
echo "🔧 管理命令:"
echo "  查看实时监控: tail -f monitoring/logs/performance.log"
echo "  停止监控: kill \$(cat monitoring/monitor.pid)"
echo "  分析性能: ./monitoring/analyze_logs.sh"