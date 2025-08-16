#!/bin/bash

# 🎯 Cube Castle 性能监控脚本
# 监控关键服务的响应时间和可用性

echo "🎯 Cube Castle 性能监控"
echo "========================"
echo "时间: $(date)"
echo ""

# 服务端点定义
COMMAND_SERVICE="http://localhost:9090"
QUERY_SERVICE="http://localhost:8090"
FRONTEND_SERVICE="http://localhost:3005"

# 测试命令服务响应时间
echo "🔧 命令服务性能测试..."
COMMAND_TIME=$(curl -o /dev/null -s -w "%{time_total}" "$COMMAND_SERVICE/health" 2>/dev/null || echo "ERROR")
if [ "$COMMAND_TIME" = "ERROR" ]; then
    echo "❌ 命令服务不可用"
else
    echo "✅ 命令服务响应时间: ${COMMAND_TIME}s"
fi

# 测试查询服务响应时间
echo ""
echo "📊 查询服务性能测试..."
QUERY_TIME=$(curl -o /dev/null -s -w "%{time_total}" \
  -X POST "$QUERY_SERVICE/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query":"query { organizationStats { total active inactive } }"}' 2>/dev/null || echo "ERROR")

if [ "$QUERY_TIME" = "ERROR" ]; then
    echo "❌ 查询服务不可用"
else
    echo "✅ 查询服务响应时间: ${QUERY_TIME}s"
fi

# 测试前端服务响应时间
echo ""
echo "🌐 前端服务性能测试..."
FRONTEND_TIME=$(curl -o /dev/null -s -w "%{time_total}" "$FRONTEND_SERVICE/" 2>/dev/null || echo "ERROR")
if [ "$FRONTEND_TIME" = "ERROR" ]; then
    echo "❌ 前端服务不可用"
else
    echo "✅ 前端服务响应时间: ${FRONTEND_TIME}s"
fi

# 测试完整CQRS流程性能
echo ""
echo "🔄 完整CQRS流程性能测试..."
START_TIME=$(date +%s.%3N)

# 创建组织（命令）
CREATE_RESPONSE=$(curl -s -X POST "$COMMAND_SERVICE/api/v1/organization-units" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "监控测试部门_'$(date +%s)'",
    "unit_type": "DEPARTMENT",
    "status": "ACTIVE",
    "level": 1,
    "sort_order": 0,
    "description": "性能监控自动化测试"
  }' 2>/dev/null)

# 提取组织代码
ORG_CODE=$(echo "$CREATE_RESPONSE" | grep -o '"code":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ORG_CODE" ]; then
    echo "❌ 创建组织失败"
else
    # 等待CDC同步
    sleep 1
    
    # 查询组织（查询）
    QUERY_RESPONSE=$(curl -s -X POST "$QUERY_SERVICE/graphql" \
      -H "Content-Type: application/json" \
      -d "{\"query\":\"query { organization(code: \\\"$ORG_CODE\\\") { code name status } }\"}" 2>/dev/null)
    
    END_TIME=$(date +%s.%3N)
    TOTAL_TIME=$(echo "$END_TIME - $START_TIME" | bc)
    
    if echo "$QUERY_RESPONSE" | grep -q "$ORG_CODE"; then
        echo "✅ 完整CQRS流程时间: ${TOTAL_TIME}s (创建→CDC同步→查询)"
    else
        echo "❌ CQRS数据同步失败"
    fi
fi

echo ""
echo "📊 监控总结:"
echo "  - 命令服务: $COMMAND_TIME"s
echo "  - 查询服务: $QUERY_TIME"s  
echo "  - 前端服务: $FRONTEND_TIME"s
if [ ! -z "$TOTAL_TIME" ]; then
    echo "  - 端到端流程: $TOTAL_TIME"s
fi
echo ""
echo "🕐 监控完成时间: $(date)"