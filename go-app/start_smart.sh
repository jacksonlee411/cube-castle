#!/bin/bash

# 智能启动脚本 - 自动处理端口占用问题
# 使用方法: ./start_smart.sh

set -e

echo "🚀 启动 Cube Castle Go 服务 (智能模式)"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PORT=8080
SERVICE_NAME="Cube Castle Go Service"

# 函数：检查端口是否被占用
check_port() {
    if sudo ss -tlnp | grep -q ":$PORT "; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 函数：获取占用端口的进程PID
get_port_pid() {
    sudo ss -tlnp | grep ":$PORT " | awk '{print $6}' | sed 's/.*pid=\([0-9]*\).*/\1/'
}

# 函数：杀死占用端口的进程
kill_port_process() {
    local pid=$1
    echo -e "${YELLOW}⚠️  发现端口 $PORT 被进程 $pid 占用${NC}"
    echo -e "${BLUE}🔄 正在终止进程 $pid...${NC}"
    
    if sudo kill -TERM $pid 2>/dev/null; then
        echo -e "${GREEN}✅ 进程 $pid 已终止${NC}"
        sleep 2
    else
        echo -e "${YELLOW}⚠️  进程 $pid 可能已经终止${NC}"
    fi
    
    # 如果进程仍然存在，强制杀死
    if ps -p $pid > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  强制终止进程 $pid...${NC}"
        sudo kill -9 $pid 2>/dev/null || true
        sleep 1
    fi
}

# 函数：清理所有相关进程
cleanup_processes() {
    echo -e "${BLUE}🧹 清理相关进程...${NC}"
    
    # 杀死所有go run进程
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    
    # 杀死所有main进程（可能是编译后的二进制文件）
    pkill -f "main" 2>/dev/null || true
    
    # 等待进程完全终止
    sleep 2
    
    echo -e "${GREEN}✅ 进程清理完成${NC}"
}

# 函数：检查Python AI服务
check_python_service() {
    if ! ss -tlnp | grep -q ":50051 "; then
        echo -e "${YELLOW}⚠️  Python AI服务未运行，正在启动...${NC}"
        cd ../python-ai
        if [ -f "venv/bin/activate" ]; then
            source venv/bin/activate
            nohup python main_mock.py > /dev/null 2>&1 &
            echo -e "${GREEN}✅ Python AI服务已启动${NC}"
            sleep 3
        else
            echo -e "${RED}❌ Python虚拟环境不存在${NC}"
            return 1
        fi
        cd ../go-app
    else
        echo -e "${GREEN}✅ Python AI服务正在运行${NC}"
    fi
}

# 函数：启动Go服务
start_go_service() {
    echo -e "${BLUE}🚀 启动Go服务...${NC}"
    
    # 编译项目
    echo -e "${BLUE}🔨 编译项目...${NC}"
    go build -v cmd/server/main.go
    
    # 启动服务
    echo -e "${BLUE}🌐 启动HTTP服务器...${NC}"
    ./main &
    local go_pid=$!
    
    # 等待服务启动
    echo -e "${BLUE}⏳ 等待服务启动...${NC}"
    sleep 5
    
    # 检查服务是否成功启动
    if curl -s http://localhost:$PORT/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $SERVICE_NAME 启动成功！${NC}"
        echo -e "${GREEN}🌐 服务地址: http://localhost:$PORT${NC}"
        echo -e "${GREEN}📊 健康检查: http://localhost:$PORT/health${NC}"
        echo -e "${GREEN}🧪 验证页面: http://localhost:$PORT/verify_1.1.1.html${NC}"
        echo -e "${BLUE}📝 进程ID: $go_pid${NC}"
        echo -e "${YELLOW}💡 按 Ctrl+C 停止服务${NC}"
        
        # 保存PID到文件
        echo $go_pid > .server.pid
        
        # 等待用户中断
        wait $go_pid
    else
        echo -e "${RED}❌ 服务启动失败${NC}"
        kill $go_pid 2>/dev/null || true
        return 1
    fi
}

# 主函数
main() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}    $SERVICE_NAME 智能启动器${NC}"
    echo -e "${BLUE}================================${NC}"
    
    # 检查是否在正确的目录
    if [ ! -f "cmd/server/main.go" ]; then
        echo -e "${RED}❌ 请在 go-app 目录下运行此脚本${NC}"
        exit 1
    fi
    
    # 检查端口占用
    if check_port; then
        local pid=$(get_port_pid)
        if [ ! -z "$pid" ]; then
            kill_port_process $pid
        fi
    fi
    
    # 清理相关进程
    cleanup_processes
    
    # 再次检查端口
    if check_port; then
        echo -e "${RED}❌ 端口 $PORT 仍然被占用，请手动检查${NC}"
        sudo ss -tlnp | grep ":$PORT "
        exit 1
    fi
    
    # 检查Python AI服务
    check_python_service
    
    # 启动Go服务
    start_go_service
}

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}🛑 正在停止服务...${NC}"
    
    # 停止Go服务
    if [ -f ".server.pid" ]; then
        local pid=$(cat .server.pid)
        kill $pid 2>/dev/null || true
        rm -f .server.pid
    fi
    
    # 清理进程
    cleanup_processes
    
    echo -e "${GREEN}✅ 服务已停止${NC}"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 运行主函数
main "$@" 