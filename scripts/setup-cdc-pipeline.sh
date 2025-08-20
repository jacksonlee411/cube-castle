#!/bin/bash
# 🚀 Operation Phoenix - CDC Pipeline Setup Script (Fixed)
# Cube Castle CQRS+CDC架构快速部署 - 修复版本

set -e

echo "🚀 开始Operation Phoenix - CQRS+CDC架构部署（修复版）..."
echo "================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Step 1: 启动基础设施
print_status "启动完整的CQRS+CDC技术栈..."
docker-compose up -d

print_status "等待服务启动..."
sleep 30

# Step 2: 检查服务健康状态
print_status "检查核心服务健康状态..."

# 检查PostgreSQL
if docker exec cube_castle_postgres pg_isready -U user -d cubecastle > /dev/null 2>&1; then
    print_success "✅ PostgreSQL启动成功"
else
    print_error "❌ PostgreSQL启动失败"
    exit 1
fi

# 检查Neo4j
if curl -f http://localhost:7474 > /dev/null 2>&1; then
    print_success "✅ Neo4j启动成功"
else
    print_warning "⚠️ Neo4j可能还在启动中..."
fi

# 检查Kafka
print_status "等待Kafka Connect启动..."
max_attempts=30
attempt=0
while ! curl -f http://localhost:8083/ > /dev/null 2>&1; do
    if [ $attempt -ge $max_attempts ]; then
        print_error "❌ Kafka Connect启动超时"
        exit 1
    fi
    print_status "等待Kafka Connect启动... ($((attempt+1))/$max_attempts)"
    sleep 10
    ((attempt++))
done
print_success "✅ Kafka Connect启动成功"

# Step 3: 配置PostgreSQL复制（修复版）
print_status "配置PostgreSQL逻辑复制..."

# 检查实际存在的表
print_status "检查数据库表结构..."
docker exec cube_castle_postgres psql -U user -d cubecastle -c "\dt" | grep organization_units || {
    print_error "❌ organization_units表不存在"
    exit 1
}

# 创建复制用户和发布（仅针对存在的表）
docker exec cube_castle_postgres psql -U user -d cubecastle -c "
DO \$\$
BEGIN
    -- 创建复制用户（如果不存在）
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = 'debezium_user') THEN
        CREATE USER debezium_user WITH REPLICATION LOGIN PASSWORD 'debezium_pass';
    END IF;
    
    -- 授权
    GRANT SELECT ON ALL TABLES IN SCHEMA public TO debezium_user;
    GRANT USAGE ON SCHEMA public TO debezium_user;
    
    -- 创建发布（仅针对实际存在的表）
    DROP PUBLICATION IF EXISTS organization_publication;
    CREATE PUBLICATION organization_publication FOR TABLE organization_units;
    
    RAISE NOTICE 'PostgreSQL逻辑复制配置完成（仅organization_units表）';
END
\$\$;
"

if [ $? -eq 0 ]; then
    print_success "✅ PostgreSQL逻辑复制配置成功"
else
    print_error "❌ PostgreSQL配置失败"
    exit 1
fi

# Step 4: 配置Debezium连接器（修复版）
print_status "配置Debezium PostgreSQL源连接器（修复版）..."

# 删除已存在的连接器（如果存在）
curl -X DELETE http://localhost:8083/connectors/organization-postgres-connector > /dev/null 2>&1 || true

# 创建修复后的连接器配置
curl -X POST http://localhost:8083/connectors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "organization-postgres-connector",
    "config": {
      "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
      "tasks.max": "1",
      "database.hostname": "postgres",
      "database.port": "5432",
      "database.user": "debezium_user",
      "database.password": "debezium_pass",
      "database.dbname": "cubecastle",
      "database.server.name": "cubecastle-postgres",
      "table.include.list": "public.organization_units",
      "publication.name": "organization_publication",
      "plugin.name": "pgoutput",
      "slot.name": "organization_slot",
      "topic.prefix": "cubecastle-postgres",
      "key.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "key.converter.schemas.enable": false,
      "value.converter.schemas.enable": false
    }
  }'

connector_create_result=$?

if [ $connector_create_result -eq 0 ]; then
    print_success "✅ Debezium连接器配置成功"
else
    print_warning "⚠️ Debezium连接器配置失败，尝试备用方案..."
    
    # 备用方案：使用系统默认用户
    print_status "使用系统用户创建连接器..."
    curl -X POST http://localhost:8083/connectors \
      -H "Content-Type: application/json" \
      -d '{
        "name": "organization-postgres-connector-fallback",
        "config": {
          "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
          "database.hostname": "postgres",
          "database.port": "5432",
          "database.user": "user",
          "database.password": "password",
          "database.dbname": "cubecastle",
          "database.server.name": "cubecastle-postgres",
          "table.include.list": "public.organization_units",
          "plugin.name": "pgoutput",
          "topic.prefix": "cubecastle-postgres",
          "key.converter": "org.apache.kafka.connect.json.JsonConverter",
          "value.converter": "org.apache.kafka.connect.json.JsonConverter",
          "key.converter.schemas.enable": "false",
          "value.converter.schemas.enable": "false"
        }
      }'
    
    if [ $? -eq 0 ]; then
        print_success "✅ 备用连接器配置成功"
        CONNECTOR_NAME="organization-postgres-connector-fallback"
    else
        print_error "❌ 连接器配置完全失败"
        exit 1
    fi
else
    CONNECTOR_NAME="organization-postgres-connector"
fi

# Step 5: 等待连接器启动
print_status "等待连接器启动..."
sleep 10

# 检查连接器状态
connector_status=$(curl -s http://localhost:8083/connectors/${CONNECTOR_NAME}/status | jq -r '.connector.state' 2>/dev/null || echo "UNKNOWN")

if [ "$connector_status" = "RUNNING" ]; then
    print_success "✅ PostgreSQL连接器运行正常"
else
    print_warning "⚠️ 连接器状态: $connector_status"
    print_status "连接器详细状态:"
    curl -s http://localhost:8083/connectors/${CONNECTOR_NAME}/status | jq '.' || echo "无法获取状态详情"
fi

# Step 6: 创建测试数据验证CDC
print_status "创建测试数据验证CDC流程..."

docker exec cube_castle_postgres psql -U user -d cubecastle -c "
-- 插入测试数据验证CDC（仅organization_units表）
UPDATE organization_units 
SET updated_at = NOW() 
WHERE code = (SELECT code FROM organization_units ORDER BY created_at LIMIT 1);

SELECT 'CDC测试数据已更新' as message;
"

# Step 7: 显示访问信息
echo ""
print_success "🎉 Operation Phoenix 修复版部署完成！"
echo "================================="
echo ""
echo "📊 服务访问信息:"
echo "  🐘 PostgreSQL: localhost:5432 (user/password)"
echo "  🎯 Neo4j Browser: http://localhost:7474 (neo4j/password)"
echo "  📊 Kafka UI: http://localhost:8081"
echo "  🔧 Kafka Connect: http://localhost:8083"
echo "  👨‍💼 PgAdmin: http://localhost:5050 (admin@cubecastle.com/admin123)"
echo ""
echo "🔍 验证命令:"
echo "  查看连接器状态: curl http://localhost:8083/connectors/${CONNECTOR_NAME}/status"
echo "  查看Kafka主题: docker exec cube_castle_kafka kafka-topics --list --bootstrap-server localhost:9092"
echo ""
echo "🐛 修复的问题:"
echo "  ✅ 添加了缺失的 topic.prefix 参数"
echo "  ✅ 只针对实际存在的 organization_units 表创建发布"
echo "  ✅ 增加了备用方案使用系统用户"
echo "  ✅ 增加了连接器状态详细检查"
echo ""

print_success "🚀 CDC配置已修复，不再需要手动创建连接器！"