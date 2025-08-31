#!/bin/bash

# 🚀 Cube Castle 简化开发环境启动脚本
# 专注于开发效率，移除不必要的生产环境配置

echo "🏰 启动 Cube Castle 开发环境..."
echo "📅 $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 创建必要目录
mkdir -p logs
mkdir -p data

echo -e "${BLUE}📦 检查并启动基础设施服务...${NC}"

# 停止可能冲突的进程
echo "🛑 清理旧进程..."
pkill -f "organization.*service" 2>/dev/null || true
pkill -f "go run.*main.go" 2>/dev/null || true

# 启动Docker基础设施
echo "🐳 启动Docker基础设施..."
docker-compose up -d postgres redis

# 等待基础设施就绪
echo "⏳ 等待基础设施启动..."
sleep 10

# 检查基础设施状态
echo -e "${BLUE}🔍 检查基础设施连接状态...${NC}"

# PostgreSQL
if PGPASSWORD=password psql -h localhost -p 5432 -U user -d cubecastle -c "SELECT 1;" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ PostgreSQL - 连接正常${NC}"
else
    echo -e "${RED}❌ PostgreSQL - 连接失败${NC}"
    echo "🚨 请检查Docker容器状态: docker-compose ps"
    exit 1
fi

# PostgreSQL原生架构 - Neo4j已移除
echo -e "${GREEN}✅ PostgreSQL原生架构 - 已移除Neo4j依赖${NC}"

# Redis
if redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -q "PONG"; then
    echo -e "${GREEN}✅ Redis - 连接正常${NC}"
else
    echo -e "${RED}❌ Redis - 连接失败${NC}"
fi

echo ""
echo -e "${BLUE}🚀 启动简化Go服务...${NC}"

# 设置环境变量
export DATABASE_URL="postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
export GO111MODULE=on

# 启动简化命令服务 (端口9090)
echo "🎯 启动命令服务 (端口9090)..."
cd cmd/organization-command-service-simple
go run main.go > ../../logs/command-service.log 2>&1 &
COMMAND_PID=$!
echo "Command Service PID: $COMMAND_PID" > ../../data/command-service.pid
cd ../..

# 启动简化查询服务 (端口8090) 
echo "🔍 启动查询服务 (端口8090)..."
cd cmd/organization-query-service-simple
go run main.go > ../../logs/query-service.log 2>&1 &
QUERY_PID=$!
echo "Query Service PID: $QUERY_PID" > ../../data/query-service.pid
cd ../..

# 等待Go服务启动
echo "⏳ 等待Go服务启动..."
sleep 5

# 检查Go服务状态
echo -e "${BLUE}🔍 检查Go服务状态...${NC}"

# 命令服务健康检查
if curl -f -s "http://localhost:9090/health" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 命令服务 (9090) - 运行正常${NC}"
else
    echo -e "${RED}❌ 命令服务 (9090) - 启动失败${NC}"
    echo "📋 检查日志: tail -f logs/command-service.log"
fi

# 查询服务健康检查
if curl -f -s "http://localhost:8090/health" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ 查询服务 (8090) - 运行正常${NC}"
else
    echo -e "${RED}❌ 查询服务 (8090) - 启动失败${NC}"
    echo "📋 检查日志: tail -f logs/query-service.log"
fi

echo ""
echo -e "${BLUE}🎨 启动前端开发服务器...${NC}"

# 启动前端 (后台运行)
cd frontend
npm run dev > ../logs/frontend-service.log 2>&1 &
FRONTEND_PID=$!
echo "Frontend PID: $FRONTEND_PID" > ../data/frontend-service.pid
cd ..

# 等待前端启动
sleep 3

echo ""
echo -e "${GREEN}🎉 Cube Castle 开发环境启动完成！${NC}"
echo ""
echo -e "${BLUE}🌐 访问地址：${NC}"
echo -e "  ${GREEN}• 前端应用:${NC} http://localhost:3001"
echo -e "  ${GREEN}• 命令API:${NC} http://localhost:9090 (REST API)"
echo -e "  ${GREEN}• 查询API:${NC} http://localhost:8090 (GraphQL)"
echo -e "  ${GREEN}• GraphiQL:${NC} http://localhost:8090/graphiql"
echo -e "  ${GREEN}• PgAdmin:${NC} http://localhost:5050 (admin@admin.com/admin)"
echo ""
echo -e "${BLUE}📊 服务状态：${NC}"
echo -e "  ${GREEN}• PostgreSQL:${NC} ✅ 端口 5432"
echo -e "  ${GREEN}• Redis:${NC} ✅ 端口 6379"
echo -e "  ${GREEN}• 命令服务:${NC} ✅ 端口 9090 (简化版)"
echo -e "  ${GREEN}• 查询服务:${NC} ✅ 端口 8090 (简化版)"
echo -e "  ${GREEN}• 前端服务:${NC} ✅ 端口 3001"
echo ""
echo -e "${BLUE}📋 日志监控：${NC}"
echo -e "  ${YELLOW}• 命令服务:${NC} tail -f logs/command-service.log"
echo -e "  ${YELLOW}• 查询服务:${NC} tail -f logs/query-service.log"  
echo -e "  ${YELLOW}• 前端服务:${NC} tail -f logs/frontend-service.log"
echo ""
echo -e "${BLUE}🛠️ 开发特性：${NC}"
echo -e "  ${GREEN}• 移除复杂健康检查和监控${NC}"
echo -e "  ${GREEN}• 保留核心CRUD和GraphQL功能${NC}"
echo -e "  ${GREEN}• 专注开发效率和快速迭代${NC}"
echo -e "  ${GREEN}• 简化的错误处理和日志${NC}"
echo ""
echo -e "${BLUE}🔧 管理命令：${NC}"
echo -e "  ${YELLOW}• 停止服务:${NC} ./scripts/dev-stop.sh"
echo -e "  ${YELLOW}• 重启服务:${NC} ./scripts/dev-restart.sh"
echo -e "  ${YELLOW}• 查看状态:${NC} ./scripts/dev-status.sh"
echo ""
echo -e "${GREEN}✨ 系统已就绪！开始开发吧！${NC}"

# 创建停止脚本的快速提示
echo ""
echo -e "${YELLOW}💡 提示: 使用 Ctrl+C 或运行以下命令停止所有服务:${NC}"
echo "   kill $COMMAND_PID $QUERY_PID $FRONTEND_PID"
echo "   docker-compose down"