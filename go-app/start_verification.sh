#!/bin/bash

# 1.1.1 CoreHR Repository层验证工具启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🏰 Cube Castle - 1.1.1 验证工具启动器${NC}"
echo "=========================================="
echo ""

# 检查Go服务是否运行
check_go_service() {
    echo -e "${BLUE}🔍 检查Go服务状态...${NC}"
    
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Go服务正在运行 (http://localhost:8080)${NC}"
        return 0
    else
        echo -e "${RED}❌ Go服务未运行${NC}"
        return 1
    fi
}

# 启动Go服务
start_go_service() {
    echo -e "${YELLOW}🚀 启动Go服务...${NC}"
    
    # 检查是否在正确的目录
    if [ ! -f "cmd/server/main.go" ]; then
        echo -e "${RED}❌ 请在go-app目录下运行此脚本${NC}"
        exit 1
    fi
    
    # 检查依赖
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装或不在PATH中${NC}"
        exit 1
    fi
    
    # 清理并重新构建
    echo -e "${BLUE}📦 清理并重新构建项目...${NC}"
    go clean -cache
    go mod tidy
    
    # 启动服务
    echo -e "${GREEN}🚀 启动Go服务...${NC}"
    echo -e "${YELLOW}💡 服务将在后台运行，按Ctrl+C停止${NC}"
    echo ""
    
    go run cmd/server/main.go &
    GO_PID=$!
    
    # 等待服务启动
    echo -e "${BLUE}⏳ 等待服务启动...${NC}"
    sleep 5
    
    # 检查服务是否成功启动
    if check_go_service; then
        echo -e "${GREEN}✅ Go服务启动成功！${NC}"
    else
        echo -e "${RED}❌ Go服务启动失败${NC}"
        kill $GO_PID 2>/dev/null || true
        exit 1
    fi
}

# 打开验证网页
open_verification_page() {
    echo ""
    echo -e "${GREEN}🌐 打开验证网页...${NC}"
    
    # 检查验证文件是否存在
    if [ ! -f "verify_1.1.1.html" ]; then
        echo -e "${RED}❌ 验证文件 verify_1.1.1.html 不存在${NC}"
        exit 1
    fi
    
    # 尝试使用不同的方式打开浏览器
    if command -v xdg-open &> /dev/null; then
        # Linux
        xdg-open verify_1.1.1.html
    elif command -v open &> /dev/null; then
        # macOS
        open verify_1.1.1.html
    elif command -v start &> /dev/null; then
        # Windows
        start verify_1.1.1.html
    else
        echo -e "${YELLOW}⚠️ 无法自动打开浏览器，请手动打开文件:${NC}"
        echo -e "${BLUE}   $(pwd)/verify_1.1.1.html${NC}"
    fi
    
    echo -e "${GREEN}✅ 验证网页已打开！${NC}"
}

# 显示使用说明
show_instructions() {
    echo ""
    echo -e "${BLUE}📋 使用说明:${NC}"
    echo "1. 在验证网页中，您可以查看1.1.1的实现状态"
    echo "2. 点击API测试按钮来验证实际功能"
    echo "3. 查看总体进度和功能覆盖度"
    echo "4. 了解下一步开发建议"
    echo ""
    echo -e "${YELLOW}🔗 API端点:${NC}"
    echo "   - 员工管理: http://localhost:8080/api/v1/corehr/employees"
    echo "   - 组织管理: http://localhost:8080/api/v1/corehr/organizations"
    echo "   - 发件箱: http://localhost:8080/api/v1/outbox"
    echo ""
    echo -e "${GREEN}🎯 验证目标:${NC}"
    echo "   ✅ 替换所有Mock数据"
    echo "   ✅ 实现真实的数据库操作"
    echo "   ✅ 实现完整的业务逻辑"
    echo ""
}

# 主函数
main() {
    echo -e "${BLUE}🔍 检查当前环境...${NC}"
    
    # 检查是否在go-app目录
    if [ ! -f "cmd/server/main.go" ]; then
        echo -e "${RED}❌ 请在go-app目录下运行此脚本${NC}"
        echo -e "${YELLOW}💡 运行命令: cd go-app && ./start_verification.sh${NC}"
        exit 1
    fi
    
    # 检查Go服务状态
    if check_go_service; then
        echo -e "${GREEN}✅ Go服务已在运行${NC}"
    else
        echo -e "${YELLOW}⚠️ Go服务未运行，正在启动...${NC}"
        start_go_service
    fi
    
    # 打开验证网页
    open_verification_page
    
    # 显示使用说明
    show_instructions
    
    echo -e "${GREEN}🎉 验证工具启动完成！${NC}"
    echo -e "${YELLOW}💡 按Ctrl+C停止Go服务${NC}"
    
    # 等待用户中断
    trap 'echo -e "\n${YELLOW}🛑 正在停止服务...${NC}"; kill $GO_PID 2>/dev/null || true; exit 0' INT
    wait
}

# 执行主函数
main "$@" 