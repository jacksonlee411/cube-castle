#!/bin/bash

# 🛑 Cube Castle 开发环境停止脚本

echo "🛑 停止 Cube Castle 开发环境..."
echo "📅 $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔌 停止Go服务...${NC}"

# 读取PID文件并停止服务
if [ -f "data/command-service.pid" ]; then
    COMMAND_PID=$(cat data/command-service.pid)
    if ps -p $COMMAND_PID > /dev/null 2>&1; then
        kill $COMMAND_PID
        echo -e "${GREEN}✅ 命令服务已停止 (PID: $COMMAND_PID)${NC}"
    else
        echo -e "${YELLOW}⚠️ 命令服务进程不存在${NC}"
    fi
    rm -f data/command-service.pid
fi

if [ -f "data/query-service.pid" ]; then
    QUERY_PID=$(cat data/query-service.pid)
    if ps -p $QUERY_PID > /dev/null 2>&1; then
        kill $QUERY_PID
        echo -e "${GREEN}✅ 查询服务已停止 (PID: $QUERY_PID)${NC}"
    else
        echo -e "${YELLOW}⚠️ 查询服务进程不存在${NC}"
    fi
    rm -f data/query-service.pid
fi

if [ -f "data/frontend-service.pid" ]; then
    FRONTEND_PID=$(cat data/frontend-service.pid)
    if ps -p $FRONTEND_PID > /dev/null 2>&1; then
        kill $FRONTEND_PID
        echo -e "${GREEN}✅ 前端服务已停止 (PID: $FRONTEND_PID)${NC}"
    else
        echo -e "${YELLOW}⚠️ 前端服务进程不存在${NC}"
    fi
    rm -f data/frontend-service.pid
fi

# 强制清理所有相关进程
echo -e "${BLUE}🧹 清理残留进程...${NC}"
pkill -f "organization.*service" 2>/dev/null || true
pkill -f "go run.*main.go" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true

echo -e "${BLUE}🐳 停止Docker基础设施 (可选)...${NC}"
read -p "是否停止Docker容器? [y/N]: " -n 1 -r
echo ""
if [[ $REPLY =~ ^[Yy]$ ]]; then
    docker-compose down
    echo -e "${GREEN}✅ Docker容器已停止${NC}"
else
    echo -e "${YELLOW}⚠️ Docker容器保持运行状态${NC}"
fi

# 清理临时文件
echo -e "${BLUE}🧹 清理临时文件...${NC}"
rm -f data/*.pid

echo ""
echo -e "${GREEN}🏁 Cube Castle 开发环境已停止！${NC}"
echo ""
echo -e "${BLUE}📋 清理完成：${NC}"
echo -e "  ${GREEN}• Go服务进程已终止${NC}"
echo -e "  ${GREEN}• 前端开发服务器已停止${NC}"
echo -e "  ${GREEN}• PID文件已清理${NC}"
echo ""
echo -e "${YELLOW}💡 提示: 重新启动请运行 ./scripts/dev-start-simple.sh${NC}"