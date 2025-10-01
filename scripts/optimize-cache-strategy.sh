#!/bin/bash
# 缓存策略优化配置脚本

echo "⚙️  Redis缓存策略优化配置"
echo "========================"

echo ""
echo "📋 当前缓存配置分析"
echo "----------------"

# 检查当前Redis配置
echo "当前Redis内存配置:"
docker exec cube_castle_redis redis-cli config get maxmemory
docker exec cube_castle_redis redis-cli config get maxmemory-policy

echo "当前缓存统计:"
docker exec cube_castle_redis redis-cli info | grep -E "keyspace_hits|keyspace_misses|used_memory_human|connected_clients"

echo ""
echo "📋 缓存键分析"
echo "-----------"

key_count=$(docker exec cube_castle_redis redis-cli dbsize)
echo "总缓存键数量: $key_count"

# 按模式分析缓存键
echo "缓存键模式分布:"
docker exec cube_castle_redis redis-cli --scan --pattern "cache:*" | wc -l | xargs echo "cache:* 模式键数量:"

echo ""
echo "📋 推荐优化策略"
echo "============="

echo "🎯 1. 缓存TTL优化建议:"
echo "   • 频繁查询 (组织列表): 5-10分钟"
echo "   • 中等频率 (单个组织): 15-30分钟"
echo "   • 低频查询 (统计数据): 1-2小时"

echo ""
echo "🎯 2. 缓存键分层策略:"
echo "   • 热点数据: cache:hot:{type}:{key}"
echo "   • 常规数据: cache:std:{type}:{key}"
echo "   • 冷数据: cache:cold:{type}:{key}"

echo ""
echo "🎯 3. 内存管理优化:"
echo "   • 设置maxmemory限制 (推荐: 512MB)"
echo "   • 采用LRU淘汰策略"
echo "   • 启用键过期监控"

echo ""
echo "📋 应用优化配置"
echo "=============="

# 配置Redis内存限制和淘汰策略
echo "设置Redis内存限制为512MB:"
docker exec cube_castle_redis redis-cli config set maxmemory 536870912

echo "设置LRU淘汰策略:"
docker exec cube_castle_redis redis-cli config set maxmemory-policy allkeys-lru

# 启用键空间通知（用于缓存失效监控）
echo "启用键过期事件通知:"
docker exec cube_castle_redis redis-cli config set notify-keyspace-events Ex

echo ""
echo "📋 缓存预热策略"
echo "============="

echo "预热关键查询缓存:"

# 预热组织列表查询
echo "预热组织列表查询..."
curl -s -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizations(pagination: { page: 1, pageSize: 20 }) { data { code name unitType status level } } }"}' \
    > /dev/null

# 预热统计查询
echo "预热统计查询..."
curl -s -X POST http://localhost:8090/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"query { organizationStats { totalCount byType { unitType count } byStatus { status count } } }"}' \
    > /dev/null

# 预热常用单个组织查询
echo "预热常用组织查询..."
for org_code in "1000000" "1000001" "1000002"; do
    curl -s -X POST http://localhost:8090/graphql \
        -H "Content-Type: application/json" \
        -d '{"query":"query { organization(code: \"'$org_code'\") { code name unitType status level } }"}' \
        > /dev/null
done

echo ""
echo "📊 优化后统计"
echo "==========="

# 显示优化后的统计
key_count_after=$(docker exec cube_castle_redis redis-cli dbsize)
echo "预热后缓存键数量: $key_count_after"

echo "优化后Redis配置:"
docker exec cube_castle_redis redis-cli config get maxmemory
docker exec cube_castle_redis redis-cli config get maxmemory-policy

echo ""
echo "✅ 缓存策略优化完成！"
echo ""
echo "🔍 监控建议："
echo "• 定期检查缓存命中率 (目标 >90%)"
echo "• 监控内存使用情况"
echo "• 观察热点查询模式并调整TTL"
echo "• 设置缓存性能告警"
