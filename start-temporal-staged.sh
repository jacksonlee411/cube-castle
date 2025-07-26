#!/bin/bash

# Temporal分阶段启动脚本
# 用于解决服务依赖和健康检查问题

set -e

echo "🚀 开始Temporal分阶段启动..."

# 阶段1：启动数据库服务
echo "📊 阶段1：启动数据库服务..."
docker-compose -f docker-compose-temporal-official.yml up -d postgres
echo "⏳ 等待PostgreSQL健康检查通过..."
timeout 60 bash -c 'until docker-compose -f docker-compose-temporal-official.yml ps postgres | grep -q "healthy"; do sleep 2; done'

# 阶段2：启动Elasticsearch
echo "🔍 阶段2：启动Elasticsearch..."
docker-compose -f docker-compose-temporal-official.yml up -d elasticsearch
echo "⏳ 等待Elasticsearch健康检查通过..."
timeout 120 bash -c 'until docker-compose -f docker-compose-temporal-official.yml ps elasticsearch | grep -q "healthy"; do sleep 5; done'

# 阶段3：启动Temporal核心服务
echo "⚡ 阶段3：启动Temporal核心服务..."
docker-compose -f docker-compose-temporal-official.yml up -d temporal
echo "⏳ 等待Temporal服务启动（3分钟）..."
sleep 180

# 检查Temporal健康状态
echo "🔍 检查Temporal服务状态..."
if docker exec cube_castle_temporal tctl cluster health 2>/dev/null; then
    echo "✅ Temporal核心服务启动成功！"
else
    echo "⚠️  Temporal核心服务仍在初始化中，继续启动UI..."
fi

# 阶段4：启动Temporal UI
echo "🖥️  阶段4：启动Temporal UI..."
docker-compose -f docker-compose-temporal-official.yml up -d temporal-ui
echo "⏳ 等待UI服务启动..."
sleep 30

# 阶段5：启动其他服务
echo "🔧 阶段5：启动其他服务..."
docker-compose -f docker-compose-temporal-official.yml up -d neo4j

echo "🎉 所有服务启动完成！"
echo ""
echo "🌐 访问地址："
echo "   - Temporal UI: http://localhost:8085"
echo "   - Neo4j Browser: http://localhost:7474"
echo "   - Elasticsearch: http://localhost:9200"
echo ""
echo "🔍 检查服务状态："
echo "   docker-compose -f docker-compose-temporal-official.yml ps"
echo ""
echo "📋 查看Temporal日志："
echo "   docker logs cube_castle_temporal -f"