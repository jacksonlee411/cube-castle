#!/bin/bash

# 📊 Cube Castle 开发环境状态检查脚本

echo "📊 Cube Castle 开发环境状态检查"
echo "📅 $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🐳 Docker容器状态:${NC}"
echo "----------------------------------------"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(postgres|neo4j|redis|kafka|zookeeper)" || echo -e "${RED}❌ 没有运行中的基础设施容器${NC}"

echo ""
echo -e "${BLUE}🔗 基础设施连接测试:${NC}"
echo "----------------------------------------"

# PostgreSQL
if PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL (5432) - 连接正常${NC}"
else
    echo -e "${RED}❌ PostgreSQL (5432) - 连接失败${NC}"
fi

# Neo4j
if curl -f -s -u neo4j:password "http://localhost:7474/db/neo4j/tx/commit" \
   -H "Content-Type: application/json" \
   -d '{"statements":[{"statement":"RETURN 1"}]}' >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Neo4j (7474) - 连接正常${NC}"
else
    echo -e "${RED}❌ Neo4j (7474) - 连接失败${NC}"
fi

# Redis
if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q "PONG"; then
    echo -e "${GREEN}✅ Redis (6379) - 连接正常${NC}"
else
    echo -e "${RED}❌ Redis (6379) - 连接失败${NC}"
fi

echo ""
echo -e "${BLUE}🚀 Go服务状态:${NC}"
echo "----------------------------------------"

# 命令服务
if curl -f -s "http://localhost:9090/health" >/dev/null 2>&1; then
    COMMAND_RESPONSE=$(curl -s "http://localhost:9090/health" | jq -r '.service + " - " + .status' 2>/dev/null || echo "running")
    echo -e "${GREEN}✅ 命令服务 (9090) - $COMMAND_RESPONSE${NC}"
else
    echo -e "${RED}❌ 命令服务 (9090) - 不可访问${NC}"
fi

# 查询服务
if curl -f -s "http://localhost:8090/health" >/dev/null 2>&1; then
    QUERY_RESPONSE=$(curl -s "http://localhost:8090/health" | jq -r '.service + " - " + .status' 2>/dev/null || echo "running")
    echo -e "${GREEN}✅ 查询服务 (8090) - $QUERY_RESPONSE${NC}"
else
    echo -e "${RED}❌ 查询服务 (8090) - 不可访问${NC}"
fi

echo ""
echo -e "${BLUE}🎨 前端服务状态:${NC}"
echo "----------------------------------------"

# 前端服务
if curl -f -s "http://localhost:3001" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 前端服务 (3001) - 运行正常${NC}"
elif curl -f -s "http://localhost:3000" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 前端服务 (3000) - 运行正常${NC}"
else
    echo -e "${RED}❌ 前端服务 - 不可访问${NC}"
fi

echo ""
echo -e "${BLUE}📊 进程状态:${NC}"
echo "----------------------------------------"

# 检查Go进程
GO_PROCESSES=$(ps aux | grep -E "(organization.*service|go run.*main.go)" | grep -v grep | wc -l)
if [ $GO_PROCESSES -gt 0 ]; then
    echo -e "${GREEN}✅ Go服务进程数: $GO_PROCESSES${NC}"
    ps aux | grep -E "(organization.*service|go run.*main.go)" | grep -v grep | awk '{print "   PID " $2 ": " $11 " " $12 " " $13}'
else
    echo -e "${RED}❌ 没有运行中的Go服务进程${NC}"
fi

# 检查前端进程
FRONTEND_PROCESSES=$(ps aux | grep "npm run dev" | grep -v grep | wc -l)
if [ $FRONTEND_PROCESSES -gt 0 ]; then
    echo -e "${GREEN}✅ 前端服务进程数: $FRONTEND_PROCESSES${NC}"
else
    echo -e "${RED}❌ 没有运行中的前端服务进程${NC}"
fi

echo ""
echo -e "${BLUE}📋 日志文件状态:${NC}"
echo "----------------------------------------"

if [ -d "logs" ]; then
    for log_file in logs/*.log; do
        if [ -f "$log_file" ]; then
            file_size=$(du -h "$log_file" | cut -f1)
            last_modified=$(stat -c %y "$log_file" | cut -d. -f1)
            echo -e "${GREEN}📄 $(basename "$log_file") - 大小: $file_size, 修改时间: $last_modified${NC}"
        fi
    done
else
    echo -e "${YELLOW}⚠️ logs目录不存在${NC}"
fi

echo ""
echo -e "${BLUE}🌐 访问地址总览:${NC}"
echo "----------------------------------------"
echo -e "${GREEN}• 前端应用:${NC} http://localhost:3001"
echo -e "${GREEN}• 命令API:${NC} http://localhost:9090 (REST)"
echo -e "${GREEN}• 查询API:${NC} http://localhost:8090 (GraphQL)"
echo -e "${GREEN}• GraphiQL:${NC} http://localhost:8090/graphiql"
echo -e "${GREEN}• Neo4j:${NC} http://localhost:7474"
echo -e "${GREEN}• PgAdmin:${NC} http://localhost:5050"

echo ""
echo -e "${BLUE}🔧 管理命令:${NC}"
echo "----------------------------------------"
echo -e "${YELLOW}• 启动服务:${NC} ./scripts/dev-start-simple.sh"
echo -e "${YELLOW}• 停止服务:${NC} ./scripts/dev-stop.sh"
echo -e "${YELLOW}• 重启服务:${NC} ./scripts/dev-restart.sh"
echo -e "${YELLOW}• 查看日志:${NC} tail -f logs/[service-name].log"

echo ""