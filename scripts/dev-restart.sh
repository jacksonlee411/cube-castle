#!/bin/bash

# 🔄 Cube Castle 开发环境重启脚本

echo "🔄 重启 Cube Castle 开发环境..."
echo "📅 $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 停止现有服务...${NC}"
bash scripts/dev-stop.sh

echo ""
echo -e "${BLUE}⏳ 等待服务完全停止...${NC}"
sleep 2

echo ""
echo -e "${BLUE}🚀 启动简化开发环境...${NC}"
bash scripts/dev-start-simple.sh

echo ""
echo -e "${GREEN}🔄 Cube Castle 开发环境重启完成！${NC}"