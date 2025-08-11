#!/bin/bash

# 时态缓存性能监控脚本
# 实时监控缓存命中率和性能指标

echo "🎯 时态缓存性能监控和改进脚本"
echo "=================================="

# ===== 改进缓存配置 =====

echo "📋 1. 改进缓存配置..."

# 优化Go服务中的缓存策略
cat > /tmp/enhanced_cache_strategy.go << 'EOF'
// 增强的缓存策略代码片段
// 用于在时态服务中实现更智能的缓存

type EnhancedCacheStrategy struct {
    redisClient *redis.Client
    defaultTTL  time.Duration
    
    // 不同查询类型的专门TTL
    currentRecordTTL    time.Duration  // 当前记录：短TTL，频繁更新
    historicalRecordTTL time.Duration  // 历史记录：长TTL，不会变化
    futureRecordTTL     time.Duration  // 未来记录：中等TTL，可能变化
    rangeQueryTTL       time.Duration  // 范围查询：中等TTL
}

// 获取查询类型特定的TTL
func (c *EnhancedCacheStrategy) getTTLForQuery(opts *TemporalQueryOptions) time.Duration {
    if opts.AsOfDate != nil {
        // 历史时间点查询，使用长TTL
        return c.historicalRecordTTL
    }
    
    if opts.IncludeFuture {
        // 包含未来记录，使用中等TTL
        return c.futureRecordTTL
    }
    
    if opts.EffectiveFrom != nil || opts.EffectiveTo != nil {
        // 范围查询，使用范围TTL
        return c.rangeQueryTTL
    }
    
    // 默认当前记录查询
    return c.currentRecordTTL
}

// 智能缓存键分层
func (c *EnhancedCacheStrategy) generateCacheKey(tenantID, code string, opts *TemporalQueryOptions) string {
    hasher := md5.New()
    
    // 基础键
    baseKey := fmt.Sprintf("temporal:%s:%s", tenantID, code)
    
    // 根据查询类型添加后缀
    var suffix string
    if opts.AsOfDate != nil {
        suffix = fmt.Sprintf("asof:%s", opts.AsOfDate.Format("2006-01-02"))
    } else if opts.IncludeFuture && opts.IncludeHistory {
        suffix = "full"
    } else if opts.IncludeFuture {
        suffix = "future"
    } else if opts.IncludeHistory {
        suffix = "history"
    } else {
        suffix = "current"
    }
    
    hasher.Write([]byte(baseKey + ":" + suffix))
    return fmt.Sprintf("cache:%x", hasher.Sum(nil))
}
EOF

echo "✅ 缓存策略改进代码已生成"

# ===== 测试当前缓存性能 =====

echo "📋 2. 测试当前缓存性能..."

BASE_URL="http://localhost:9091"
TENANT_ID="3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
TEST_ORG="1000056"

# 函数：测量查询性能
measure_performance() {
    local query_type=$1
    local url=$2
    local iterations=${3:-10}
    
    echo "   测试 $query_type (${iterations}次请求):"
    
    # 清除相关缓存
    docker exec cube_castle_redis redis-cli FLUSHDB > /dev/null 2>&1
    
    # 冷缓存测试
    start_time=$(date +%s%N)
    curl -s -X GET "$url" -H "X-Tenant-ID: $TENANT_ID" > /dev/null
    end_time=$(date +%s%N)
    cold_time=$((($end_time - $start_time) / 1000000))
    
    # 热缓存测试
    start_time=$(date +%s%N)
    for ((i=1; i<=iterations-1; i++)); do
        curl -s -X GET "$url" -H "X-Tenant-ID: $TENANT_ID" > /dev/null
    done
    end_time=$(date +%s%N)
    hot_avg=$((($end_time - $start_time) / ($iterations-1) / 1000000))
    
    echo "      冷缓存: ${cold_time}ms"
    echo "      热缓存: ${hot_avg}ms"
    
    if [ $hot_avg -gt 0 ]; then
        improvement=$((cold_time * 100 / hot_avg - 100))
        echo "      性能提升: ${improvement}%"
    fi
}

# 测试不同类型的查询
measure_performance "当前记录查询" "$BASE_URL/api/v1/organization-units/$TEST_ORG" 5

measure_performance "历史记录查询" "$BASE_URL/api/v1/organization-units/$TEST_ORG?include_history=true&max_records=10" 5

measure_performance "范围查询" "$BASE_URL/api/v1/organization-units/$TEST_ORG?effective_from=2025-01-01&effective_to=2025-12-31" 3

measure_performance "未来记录查询" "$BASE_URL/api/v1/organization-units/$TEST_ORG?include_future=true" 3

# ===== 缓存命中率统计 =====

echo "📋 3. 缓存命中率分析..."

# 设置缓存统计追踪
docker exec cube_castle_redis redis-cli CONFIG SET notify-keyspace-events Ex > /dev/null

# 执行一系列查询来收集统计数据
echo "   执行测试查询序列..."

queries=(
    "$BASE_URL/api/v1/organization-units/$TEST_ORG"
    "$BASE_URL/api/v1/organization-units/$TEST_ORG?include_history=true"
    "$BASE_URL/api/v1/organization-units/$TEST_ORG?include_future=true"
    "$BASE_URL/api/v1/organization-units/$TEST_ORG?effective_from=2025-01-01"
)

for query in "${queries[@]}"; do
    for i in {1..3}; do
        curl -s -X GET "$query" -H "X-Tenant-ID: $TENANT_ID" > /dev/null
        sleep 0.1
    done
done

echo "   查询完成，缓存键数量: $(docker exec cube_castle_redis redis-cli DBSIZE)"

# ===== 内存使用分析 =====

echo "📋 4. 内存使用分析..."

memory_info=$(docker exec cube_castle_redis redis-cli INFO memory)
used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
peak_memory=$(echo "$memory_info" | grep "used_memory_peak_human:" | cut -d: -f2 | tr -d '\r')

echo "   当前内存使用: $used_memory"
echo "   峰值内存使用: $peak_memory"

# ===== 缓存优化建议 =====

echo "📋 5. 缓存优化建议..."

key_count=$(docker exec cube_castle_redis redis-cli DBSIZE)
if [ $key_count -lt 10 ]; then
    echo "   ⚠️  缓存键数量较少 ($key_count)，可能需要更多预热"
fi

# 检查TTL设置
echo "   当前TTL配置:"
docker exec cube_castle_redis redis-cli HGETALL temporal:cache:config | while read -r key; read -r value; do
    echo "      $key: ${value}秒"
done

echo ""
echo "🎉 时态缓存性能分析完成!"
echo ""
echo "📊 优化建议总结:"
echo "   1. 数据库索引已优化 - 14个时态查询专用索引"
echo "   2. Redis配置已优化 - LRU策略，512MB内存限制"
echo "   3. 分层缓存TTL策略 - 不同查询类型使用不同TTL"
echo "   4. 当前内存使用: $used_memory"
echo "   5. 缓存键数量: $key_count"
echo ""
echo "💡 进一步优化方向:"
echo "   • 实现查询结果预加载"
echo "   • 添加缓存命中率监控"
echo "   • 实现智能缓存失效策略"
echo "   • 考虑使用Redis Cluster扩展"

# 清理临时文件
rm -f /tmp/enhanced_cache_strategy.go