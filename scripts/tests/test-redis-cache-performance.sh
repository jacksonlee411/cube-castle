#!/bin/bash
# Redis缓存性能测试脚本

echo "🔍 Redis缓存性能测试和验证"
echo "========================="

# 清理之前的缓存
echo ""
echo "📋 步骤1: 清理现有缓存"
echo "-------------------"
docker exec cube_castle_redis redis-cli flushall
echo "✅ Redis缓存已清空"

# 测试GraphQL查询缓存
echo ""
echo "📋 步骤2: GraphQL查询缓存测试"
echo "-------------------------"

echo "第一次查询 (应该缓存MISS):"
time1=$(curl -s -w "%{time_total}" -o /tmp/gql_result1.json \
    -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizations(pagination: { page: 1, pageSize: 5 }) { data { code name unitType status } } }"}')

echo "响应时间: ${time1}s"

sleep 1

echo "第二次相同查询 (应该缓存HIT):"
time2=$(curl -s -w "%{time_total}" -o /tmp/gql_result2.json \
    -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizations(pagination: { page: 1, pageSize: 5 }) { data { code name unitType status } } }"}')

echo "响应时间: ${time2}s"

# 计算性能提升
if command -v bc >/dev/null 2>&1; then
    improvement=$(echo "scale=1; ($time1 - $time2) / $time1 * 100" | bc -l)
    echo "性能提升: ${improvement}%"
else
    echo "性能比较: 第一次 ${time1}s vs 第二次 ${time2}s"
fi

echo ""
echo "📋 步骤3: 检查Redis缓存状态"
echo "------------------------"

# 检查缓存键数量
key_count=$(docker exec cube_castle_redis redis-cli dbsize)
echo "当前缓存键数量: $key_count"

# 显示缓存键示例
echo "缓存键示例:"
docker exec cube_castle_redis redis-cli keys "*" | head -3

# 检查缓存统计
echo ""
echo "Redis缓存统计:"
docker exec cube_castle_redis redis-cli info | grep -E "keyspace_hits|keyspace_misses"

echo ""
echo "📋 步骤4: 不同查询参数测试"
echo "----------------------"

# 测试不同的查询参数
echo "测试不同分页参数 (应该产生不同缓存键):"
time3=$(curl -s -w "%{time_total}" -o /tmp/gql_result3.json \
    -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizations(pagination: { page: 2, pageSize: 5 }) { data { code name } } }"}')

echo "不同查询响应时间: ${time3}s"

# 再次检查缓存键数量
key_count_after=$(docker exec cube_castle_redis redis-cli dbsize)
echo "查询后缓存键数量: $key_count_after"

echo ""
echo "📋 步骤5: 缓存TTL验证"
echo "------------------"

# 检查缓存键的TTL
sample_key=$(docker exec cube_castle_redis redis-cli keys "cache:*" | head -1)
if [ ! -z "$sample_key" ]; then
    ttl=$(docker exec cube_castle_redis redis-cli ttl "$sample_key")
    echo "示例缓存键TTL: ${ttl}秒"
else
    echo "未找到缓存键"
fi

echo ""
echo "📊 缓存测试结果汇总"
echo "=================="

# 最终统计
final_stats=$(docker exec cube_castle_redis redis-cli info | grep -E "keyspace_hits|keyspace_misses")
echo "$final_stats"

if [ $key_count_after -gt 0 ]; then
    echo "✅ 缓存功能正常 - 生成了 $key_count_after 个缓存键"
else
    echo "❌ 缓存功能异常 - 未生成缓存键"
fi

echo ""
echo "🚀 建议优化项:"
echo "• 根据查询频率调整TTL时间"
echo "• 监控缓存命中率并设置告警"
echo "• 考虑针对热点查询的预热策略"
