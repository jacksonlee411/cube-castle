#!/bin/bash
echo "🎯 Cube Castle 性能监控"
echo "========================"
echo "时间: $(date)"
echo ""

# 测试命令服务
echo "🔧 命令服务测试..."
COMMAND_TIME=$(curl -o /dev/null -s -w "%{time_total}" "http://localhost:9090/health" 2>/dev/null || echo "ERROR")
echo "命令服务响应时间: $COMMAND_TIME"

# 测试查询服务
echo ""
echo "📊 查询服务测试..."
QUERY_TIME=$(curl -o /dev/null -s -w "%{time_total}" -X POST "http://localhost:8090/graphql" -H "Content-Type: application/json" -d '{"query":"query { organizationStats { total active } }"}' 2>/dev/null || echo "ERROR")
echo "查询服务响应时间: $QUERY_TIME"

# 测试前端服务
echo ""
echo "🌐 前端服务测试..."
FRONTEND_TIME=$(curl -o /dev/null -s -w "%{time_total}" "http://localhost:3005/" 2>/dev/null || echo "ERROR")
echo "前端服务响应时间: $FRONTEND_TIME"

echo ""
echo "✅ 监控完成"
