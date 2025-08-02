#!/bin/bash
# 🚀 Operation Phoenix - CDC Pipeline Setup Script
# Cube Castle CQRS+CDC架构快速部署

set -e

echo "🚀 开始Operation Phoenix - CQRS+CDC架构部署..."
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

# Step 3: 配置PostgreSQL复制
print_status "配置PostgreSQL逻辑复制..."

# 创建复制用户和发布
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
    
    -- 创建发布（如果不存在）
    IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname = 'organization_publication') THEN
        CREATE PUBLICATION organization_publication FOR TABLE 
            employees, organization_units, positions;
    END IF;
    
    RAISE NOTICE 'PostgreSQL逻辑复制配置完成';
END
\$\$;
"

if [ $? -eq 0 ]; then
    print_success "✅ PostgreSQL逻辑复制配置成功"
else
    print_error "❌ PostgreSQL配置失败"
    exit 1
fi

# Step 4: 配置Debezium连接器
print_status "配置Debezium PostgreSQL源连接器..."

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
      "database.server.name": "organization_db",
      "table.include.list": "public.employees,public.organization_units,public.positions",
      "publication.name": "organization_publication",
      "plugin.name": "pgoutput",
      "slot.name": "organization_slot",
      "key.converter": "org.apache.kafka.connect.json.JsonConverter",
      "value.converter": "org.apache.kafka.connect.json.JsonConverter",
      "key.converter.schemas.enable": false,
      "value.converter.schemas.enable": false
    }
  }'

if [ $? -eq 0 ]; then
    print_success "✅ Debezium连接器配置成功"
else
    print_warning "⚠️ Debezium连接器配置可能失败，但继续执行..."
fi

# Step 5: 等待连接器启动
print_status "等待连接器启动..."
sleep 10

# 检查连接器状态
connector_status=$(curl -s http://localhost:8083/connectors/organization-postgres-connector/status | jq -r '.connector.state' 2>/dev/null || echo "UNKNOWN")

if [ "$connector_status" = "RUNNING" ]; then
    print_success "✅ PostgreSQL连接器运行正常"
else
    print_warning "⚠️ 连接器状态: $connector_status"
fi

# Step 6: 创建测试数据验证CDC
print_status "创建测试数据验证CDC流程..."

docker exec cube_castle_postgres psql -U user -d cubecastle -c "
-- 插入测试数据验证CDC
INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name, email, hire_date, employment_status)
VALUES (
    gen_random_uuid(),
    gen_random_uuid(),
    'FULL_TIME',
    'Phoenix',
    'TestUser',
    'phoenix.test@cubecastle.com',
    NOW(),
    'ACTIVE'
);

SELECT 'CDC测试数据已插入' as message;
"

# Step 7: 显示访问信息
echo ""
print_success "🎉 Operation Phoenix 第一阶段部署完成！"
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
echo "  查看连接器状态: curl http://localhost:8083/connectors/organization-postgres-connector/status"
echo "  查看Kafka主题: docker exec cube_castle_kafka kafka-topics --list --bootstrap-server localhost:9092"
echo ""
echo "📋 下一步:"
echo "  1. 访问 Kafka UI (http://localhost:8081) 查看数据流"
echo "  2. 检查是否有 'organization_db.public.employees' 主题"
echo "  3. 验证数据变更是否正确捕获"
echo ""
print_success "🚀 开始Phase 2: CQRS架构重构..."

# Step 8: 创建CQRS项目结构
print_status "创建CQRS项目结构..."

# 创建目录结构
mkdir -p go-app/internal/cqrs/{commands,queries,events,handlers}
mkdir -p go-app/internal/repositories
mkdir -p go-app/contracts/schemas

print_success "✅ CQRS项目结构创建完成"

echo ""
print_success "🎯 Operation Phoenix 已启动！"
print_status "团队可以开始开发CQRS架构了！"
echo ""