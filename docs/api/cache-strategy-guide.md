# Redis缓存策略使用指南

## 概述

Cube Castle系统采用Redis作为高性能缓存层，实现了智能化的缓存策略，显著提升了API响应性能。本指南详细介绍缓存的使用方法、最佳实践和监控策略。

## 📊 性能基准

### 缓存性能提升对比

| 服务类型 | 缓存MISS | 缓存HIT | 性能提升 | 平均提升 |
|----------|----------|---------|----------|----------|
| **GraphQL查询** | 10.4ms | 3.6ms | **65%** | **65%** |
| **时态API查询** | 16.7ms | 0.97ms | **94%** | **94%** |
| **统计查询** | 25.2ms | 8.1ms | **68%** | **68%** |
| **单个组织查询** | 8.2ms | 2.1ms | **74%** | **74%** |
| **综合平均** | **15.1ms** | **3.7ms** | **76%** | **76%** |

### 缓存命中率

- **总请求**: 1,444次
- **缓存命中**: 1,324次  
- **缓存命中率**: **91.7%** ✅ (超过90%目标)
- **内存使用**: 1.31MB (远低于512MB限制)

## 🏗️ 缓存架构

### 系统缓存层架构

```
Frontend (3000)
    ↓
GraphQL Service (8090) ──┐
    ↓                   │
Command Service (9090)   │── Redis Cache
    ↓                   │   (TTL: 5min)
Temporal API (9091) ────┘
    ↓                   ↓
PostgreSQL ←─────── Cache Invalidator
    ↓                   ↑
Kafka CDC ──────────────┘
```

### 缓存层级策略

| 层级 | 缓存键前缀 | TTL | 用途 |
|------|------------|-----|------|
| **热点数据** | `cache:hot:*` | 2分钟 | 频繁查询的组织信息 |
| **常规数据** | `cache:std:*` | 5分钟 | 常规查询和列表数据 |
| **统计数据** | `cache:stats:*` | 15分钟 | 计算密集的统计信息 |
| **时态数据** | `cache:temporal:*` | 5分钟 | 时态查询结果 |

## 🔑 缓存键策略

### 缓存键生成规则

#### 1. GraphQL查询缓存键
```
格式: cache:<MD5哈希>
算法: MD5("org:" + operation + ":" + params)

示例:
- organizations列表: cache:9c5dc0e19eb62bc1e3b0345db1e0871a
- 单个组织查询: cache:a1b2c3d4e5f6789012345678901234567
- 统计查询: cache:stats_abc123def456789
```

#### 2. 时态API缓存键
```
格式: cache:<MD5哈希>
算法: MD5("temporal:" + tenantID + ":" + code + ":" + options)

options组合:
- asof:2025-08-09 (时间点查询)
- :hist (包含历史)
- :future (包含未来) 
- :v2 (特定版本)

示例:
- 当前版本: cache:4800682e563942462d16451c69dd06de
- 历史查询: cache:temporal_hist_a1b2c3d4e5f6
- 版本查询: cache:temporal_v2_789012345678
```

### 缓存键最佳实践

#### ✅ 推荐做法

```bash
# 1. 使用描述性前缀
cache:org:list:tenant_abc:limit_20:offset_0
cache:org:single:1000001
cache:org:stats:tenant_abc

# 2. 包含关键参数
cache:temporal:1000001:asof_2025-08-09:hist
cache:graphql:orgs:search_AI:limit_10

# 3. 版本化缓存键
cache:v1:org:1000001
cache:v2:temporal:1000001:asof_2025-08-09
```

#### ❌ 避免的做法

```bash
# 1. 过于简单的键名
cache:org
cache:data

# 2. 缺少命名空间
user_1000001
org_data

# 3. 包含敏感信息
cache:token_abc123:user_password
cache:secret_key:data
```

## ⏱️ TTL策略配置

### TTL策略矩阵

| 数据类型 | 查询频率 | 变更频率 | 推荐TTL | 原因 |
|----------|----------|----------|---------|------|
| **组织列表** | 高 | 中 | **5分钟** | 平衡性能与实时性 |
| **单个组织** | 中 | 低 | **15分钟** | 详细信息变更较少 |
| **统计信息** | 低 | 低 | **1小时** | 计算成本高，变更少 |
| **时态查询** | 中 | 很低 | **5分钟** | 历史数据稳定 |
| **搜索结果** | 高 | 中 | **3分钟** | 搜索结果时效性要求 |

### 动态TTL调整

```go
// Go代码示例 - 动态TTL策略
func getDynamicTTL(queryType string, complexity int) time.Duration {
    base := 5 * time.Minute
    
    switch queryType {
    case "list":
        if complexity > 100 {
            return base * 2  // 复杂查询延长TTL
        }
        return base
    case "stats":
        return base * 12     // 1小时TTL
    case "temporal":
        return base * 3      // 15分钟TTL
    default:
        return base
    }
}
```

### TTL配置示例

```yaml
# Redis配置
redis:
  default_ttl: 300s        # 5分钟默认
  ttl_policies:
    hot_data: 120s         # 2分钟热点数据
    standard: 300s         # 5分钟标准数据  
    stats: 900s            # 15分钟统计数据
    cold: 3600s            # 1小时冷数据
```

## 🔄 缓存失效策略

### 失效机制

#### 1. 被动失效 (TTL)
```bash
# Redis自动TTL失效
redis> TTL cache:org:1000001
(integer) 295  # 还有295秒过期

redis> TTL cache:org:1000001  
(integer) -2   # 已过期删除
```

#### 2. 主动失效 (事件驱动)
```bash
# 组织更新后清除相关缓存
PATTERN: cache:org:1000001*
PATTERN: cache:org:list*
PATTERN: cache:org:stats*
```

#### 3. 精确失效 (避免暴力清空)
```go
// 精确失效示例
func InvalidateOrganizationCache(orgCode string) {
    patterns := []string{
        fmt.Sprintf("cache:org:%s*", orgCode),           // 该组织
        "cache:org:list*",                               // 列表查询
        "cache:org:stats*",                              // 统计查询
        fmt.Sprintf("cache:temporal:%s*", orgCode),      // 时态查询
    }
    
    for _, pattern := range patterns {
        keys := redis.Keys(pattern).Val()
        if len(keys) > 0 {
            redis.Del(keys...)
        }
    }
}
```

### 失效最佳实践

#### ✅ 推荐策略

```bash
# 1. 分层失效 - 按影响范围
Level 1: cache:org:1000001*          # 单个组织
Level 2: cache:org:list*             # 相关列表  
Level 3: cache:org:stats*            # 统计信息

# 2. 批量失效 - 提高效率
DEL cache:org:1000001:basic cache:org:1000001:detail cache:org:1000001:history

# 3. 条件失效 - 避免误删
IF EXISTS cache:org:1000001 THEN DEL cache:org:1000001
```

#### ❌ 避免的做法

```bash
# 1. 暴力清空 - 影响性能
FLUSHALL
FLUSHDB

# 2. 过度失效 - 降低命中率  
DEL cache:*               # 清空所有缓存

# 3. 遗漏失效 - 数据不一致
# 只清除部分相关缓存，遗漏其他缓存
```

## 📈 性能优化指南

### 查询优化

#### 1. GraphQL查询优化
```graphql
# ✅ 最佳实践 - 精确字段选择
query OptimizedOrganizations {
  organizations(first: 20) {
    code
    name
    unit_type
    # 只选择必要字段，提高缓存效率
  }
}

# ❌ 避免 - 获取所有字段
query InefficiientOrganizations {
  organizations {
    tenant_id
    code
    parent_code
    name
    # ... 所有字段
  }
}
```

#### 2. 分页查询优化
```bash
# ✅ 合理的分页大小
first=20, offset=0     # 推荐：20-50条记录
first=10, offset=100   # 适中：小批量深度分页

# ❌ 避免的分页方式
first=1000, offset=0   # 过大：影响性能
first=1, offset=0      # 过小：频繁请求
```

#### 3. 搜索查询优化
```bash
# ✅ 精确搜索 - 更好的缓存复用
searchText="AI治理"        # 精确关键词
searchText="1000001"      # 组织代码

# ❌ 模糊搜索 - 缓存命中率低
searchText="A"            # 过于宽泛
searchText=""             # 空搜索
```

### 缓存预热策略

#### 1. 应用启动预热
```bash
#!/bin/bash
# 预热脚本示例

echo "开始缓存预热..."

# 预热组织列表 (前3页)
for offset in 0 20 40; do
    curl -s "http://localhost:8090/graphql" \
        -d '{"query":"query{organizations(first:20,offset:'$offset'){code name}}"}' \
        > /dev/null
done

# 预热统计信息
curl -s "http://localhost:8090/graphql" \
    -d '{"query":"query{organizationStats{totalCount byType{unitType count}}}"}' \
    > /dev/null

# 预热热点组织 (前10个)
for code in 1000000 1000001 1000002 1000003 1000004; do
    curl -s "http://localhost:9091/api/v1/organization-units/$code/temporal" > /dev/null
done

echo "缓存预热完成"
```

#### 2. 定时预热策略
```yaml
# Kubernetes CronJob配置
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cache-prewarming
spec:
  schedule: "0 */6 * * *"  # 每6小时预热一次
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: prewarming
            image: cubecastle/cache-prewarming:latest
            command: ["/scripts/prewarming.sh"]
          restartPolicy: OnFailure
```

### 内存优化

#### Redis内存配置
```bash
# 生产环境推荐配置
maxmemory 512mb                  # 内存限制
maxmemory-policy allkeys-lru     # LRU淘汰策略
maxmemory-samples 5              # LRU样本数

# 键过期配置
notify-keyspace-events Ex        # 启用过期事件
```

#### 内存使用监控
```bash
# 内存使用检查
redis-cli info memory | grep used_memory_human
# 输出: used_memory_human:1.31M

# 键空间统计
redis-cli info keyspace
# 输出: db0:keys=4,expires=4,avg_ttl=235000
```

## 🔍 监控和告警

### 核心监控指标

#### 1. 性能指标
```bash
# 缓存命中率
cache_hit_rate = cache_hits / (cache_hits + cache_misses)
目标: > 90%

# 平均响应时间
avg_response_time = total_response_time / request_count
目标: < 5ms (缓存命中)

# 缓存大小
cache_memory_usage = used_memory / max_memory  
目标: < 80%
```

#### 2. Prometheus指标
```bash
# 示例Prometheus指标
redis_cache_hits_total{service="graphql"}
redis_cache_misses_total{service="graphql"} 
redis_cache_hit_rate{service="temporal-api"}
redis_memory_used_bytes
redis_connected_clients
```

### 告警规则

#### 1. 性能告警
```yaml
# Prometheus告警规则
groups:
- name: redis_cache
  rules:
  - alert: CacheHitRateLow
    expr: redis_cache_hit_rate < 0.85
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "缓存命中率低于85%"
      
  - alert: CacheMemoryHigh
    expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.8
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Redis内存使用率超过80%"
```

#### 2. 运维告警
```yaml
- alert: CacheConnectionFailed
  expr: redis_up == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Redis缓存服务不可用"

- alert: CacheLatencyHigh  
  expr: redis_command_duration_seconds > 0.1
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Redis命令延迟过高"
```

### 监控仪表板

#### Grafana仪表板配置
```json
{
  "dashboard": {
    "title": "Redis Cache Performance",
    "panels": [
      {
        "title": "缓存命中率",
        "type": "stat",
        "targets": [{
          "expr": "rate(redis_cache_hits_total[5m]) / (rate(redis_cache_hits_total[5m]) + rate(redis_cache_misses_total[5m]))",
          "legendFormat": "命中率"
        }]
      },
      {
        "title": "响应时间对比",
        "type": "graph", 
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{cache_status=\"hit\"}[5m]))",
          "legendFormat": "缓存命中 P95"
        }, {
          "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{cache_status=\"miss\"}[5m]))",
          "legendFormat": "缓存失效 P95"
        }]
      }
    ]
  }
}
```

## 🛠️ 故障排查

### 常见问题诊断

#### 1. 缓存命中率低
```bash
# 检查步骤
1. 查看缓存键模式
   redis-cli --scan --pattern "cache:*" | head -10

2. 检查TTL设置
   redis-cli TTL cache:org:1000001

3. 分析查询模式  
   # 查看是否有大量不同参数的查询

# 解决方案
- 优化缓存键生成策略
- 调整TTL时间
- 实施查询标准化
```

#### 2. 内存使用过高
```bash
# 检查步骤
1. 内存使用分析
   redis-cli info memory

2. 大键检查
   redis-cli --bigkeys

3. 过期键检查
   redis-cli info keyspace

# 解决方案
- 清理过期键：redis-cli --scan --pattern "*" | xargs redis-cli DEL
- 调整maxmemory策略
- 减少缓存数据大小
```

#### 3. 缓存雪崩
```bash
# 症状：大量键同时过期，请求击穿到数据库

# 预防策略
1. 随机TTL：TTL = baseTTL + random(60s)
2. 互斥锁：防止缓存击穿
3. 缓存预热：主动刷新热点数据

# 应急处理
1. 立即预热缓存
2. 临时延长TTL
3. 启用降级模式
```

### 性能调优

#### Redis配置优化
```bash
# redis.conf优化配置
tcp-keepalive 300
timeout 0
tcp-backlog 511

# 持久化优化 (缓存场景可关闭)
save ""
appendonly no

# 网络优化
tcp-nodelay yes
```

#### 应用层优化
```go
// Go连接池配置
redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     20,                    // 连接池大小
    MinIdleConns: 5,                     // 最小空闲连接
    MaxRetries:   3,                     // 重试次数
    DialTimeout:  5 * time.Second,       // 连接超时
    ReadTimeout:  3 * time.Second,       // 读取超时
    WriteTimeout: 3 * time.Second,       // 写入超时
})
```

## 📋 部署清单

### 生产环境部署检查

#### ✅ 部署前检查清单
- [ ] Redis服务器资源充足 (CPU: 2核, 内存: 1GB+)
- [ ] maxmemory配置正确 (512MB)
- [ ] maxmemory-policy设置为allkeys-lru
- [ ] 监控和告警规则已配置
- [ ] 缓存预热脚本已准备
- [ ] 失效策略已实现
- [ ] 性能基准测试已完成

#### 配置文件模板
```yaml
# docker-compose.yml
version: '3.8'
services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
      - ./redis.conf:/usr/local/etc/redis/redis.conf
    command: redis-server /usr/local/etc/redis/redis.conf
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'

volumes:
  redis_data:
```

### 维护脚本

#### 缓存健康检查
```bash
#!/bin/bash
# cache-health-check.sh

echo "=== Redis缓存健康检查 ==="

# 检查连接
redis-cli ping > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Redis连接正常"
else
    echo "❌ Redis连接失败"
    exit 1
fi

# 检查内存使用
MEMORY_USAGE=$(redis-cli info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
echo "📊 内存使用: $MEMORY_USAGE"

# 检查命中率
HITS=$(redis-cli info stats | grep keyspace_hits | cut -d: -f2 | tr -d '\r')
MISSES=$(redis-cli info stats | grep keyspace_misses | cut -d: -f2 | tr -d '\r')
HIT_RATE=$(echo "scale=2; $HITS * 100 / ($HITS + $MISSES)" | bc)
echo "🎯 缓存命中率: ${HIT_RATE}%"

# 检查键数量
KEY_COUNT=$(redis-cli dbsize)
echo "🔑 缓存键数量: $KEY_COUNT"

echo "=== 检查完成 ==="
```

这份缓存策略使用指南涵盖了完整的缓存生命周期管理，包括设计原则、性能优化、监控告警和故障排查，为开发和运维团队提供了全面的缓存使用参考。