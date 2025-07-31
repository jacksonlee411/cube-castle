#!/bin/bash

# 智能停止脚本 - 安全停止所有相关服务
# 使用方法: ./stop_smart.sh

set -e

echo "🛑 停止 Cube Castle 服务"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：停止Go服务
stop_go_service() {
    echo -e "${BLUE}🛑 停止Go服务...${NC}"
    
    # 从PID文件停止
    if [ -f ".server.pid" ]; then
        local pid=$(cat .server.pid)
        echo -e "${BLUE}📝 从PID文件停止进程 $pid${NC}"
        kill $pid 2>/dev/null || true
        rm -f .server.pid
    fi
    
    # 停止所有go run进程
    echo -e "${BLUE}🔄 停止所有go run进程...${NC}"
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    
    # 停止所有main进程
    echo -e "${BLUE}🔄 停止所有main进程...${NC}"
    pkill -f "main" 2>/dev/null || true
    
    echo -e "${GREEN}✅ Go服务已停止${NC}"
}

# 函数：停止Python AI服务
stop_python_service() {
    echo -e "${BLUE}🛑 停止Python AI服务...${NC}"
    
    # 停止所有python main_mock.py进程
    pkill -f "python main_mock.py" 2>/dev/null || true
    
    # 停止所有python进程（更宽泛的匹配）
    pkill -f "main_mock.py" 2>/dev/null || true
    
    echo -e "${GREEN}✅ Python AI服务已停止${NC}"
}

# 函数：检查端口占用
check_ports() {
    echo -e "${BLUE}🔍 检查端口占用情况...${NC}"
    
    # 检查8080端口
    if sudo ss -tlnp | grep -q ":8080 "; then
        echo -e "${YELLOW}⚠️  端口8080仍被占用${NC}"
        sudo ss -tlnp | grep ":8080 "
    else
        echo -e "${GREEN}✅ 端口8080已释放${NC}"
    fi
    
    # 检查50051端口
    if sudo ss -tlnp | grep -q ":50051 "; then
        echo -e "${YELLOW}⚠️  端口50051仍被占用${NC}"
        sudo ss -tlnp | grep ":50051 "
    else
        echo -e "${GREEN}✅ 端口50051已释放${NC}"
    fi
}

# 函数：强制清理
force_cleanup() {
    echo -e "${YELLOW}⚠️  执行强制清理...${NC}"
    
    # 强制杀死所有相关进程
    sudo pkill -9 -f "go run" 2>/dev/null || true
    sudo pkill -9 -f "main" 2>/dev/null || true
    sudo pkill -9 -f "python main_mock.py" 2>/dev/null || true
    
    # 等待进程完全终止
    sleep 2
    
    echo -e "${GREEN}✅ 强制清理完成${NC}"
}

# 主函数
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}    Cube Castle 服务停止器${NC}"
    echo -e "${BLUE}================================${NC}"
    
    # 停止Go服务
    stop_go_service
    
    # 停止Python AI服务
    stop_python_service
    
    # 等待进程终止
    echo -e "${BLUE}⏳ 等待进程完全终止...${NC}"
    sleep 3
    
    # 检查端口占用
    check_ports
    
    # 如果端口仍被占用，执行强制清理
    if sudo ss -tlnp | grep -q ":8080 \|:50051 "; then
        echo -e "${YELLOW}⚠️  检测到端口仍被占用，执行强制清理${NC}"
        force_cleanup
        check_ports
    fi
    
    echo -e "${GREEN}🎉 所有服务已停止${NC}"
}

# 运行主函数
main "$@" 