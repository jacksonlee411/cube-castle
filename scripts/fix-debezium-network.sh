#!/bin/bash

# Debezium CDC网络修复脚本
# 解决 java.net.UnknownHostException: postgres 问题

set -e

echo "🔧 修复Debezium CDC网络配置..."

# 1. 检查Docker网络状态
echo "📋 检查Docker网络状态"
docker network inspect cube-castle_default

# 2. 获取PostgreSQL容器的准确网络名称
POSTGRES_CONTAINER=$(docker ps --format "table {{.Names}}" | grep postgres)
echo "📦 PostgreSQL容器名称: $POSTGRES_CONTAINER"

# 3. 删除错误的连接器配置
echo "🗑️ 删除现有连接器配置"
curl -X DELETE http://localhost:8083/connectors/organization-postgres-connector || echo "连接器不存在，跳过删除"

# 4. 等待连接器删除完成
sleep 5

# 5. 重新创建正确的连接器配置
echo "✨ 创建修复后的连接器配置"
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "organization-postgres-connector",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
      "database.hostname": "'$POSTGRES_CONTAINER'",
      "database.port": "5432", 
      "database.user": "user",
      "database.password": "password",
      "database.dbname": "cubecastle",
      "topic.prefix": "organization_db",
      "table.include.list": "public.organization_units",
      "plugin.name": "pgoutput",
      "slot.name": "organization_slot_fixed",
      "publication.name": "organization_publication_fixed",
      "transforms": "unwrap",
      "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
      "transforms.unwrap.drop.tombstones": "false"
    }
  }'

# 6. 验证连接器状态
echo "🔍 验证连接器状态"
sleep 10
curl -s http://localhost:8083/connectors/organization-postgres-connector/status | jq '.'

# 7. 检查Kafka主题
echo "📝 检查Kafka主题"
docker exec cube_castle_kafka kafka-topics.sh --bootstrap-server localhost:9092 --list | grep organization || echo "主题尚未创建"

echo "✅ Debezium CDC网络修复完成"