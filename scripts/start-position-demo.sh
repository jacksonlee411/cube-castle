#!/bin/bash

# 职位管理系统演示启动脚本
# 版本: v1.0
# 创建日期: 2025-08-05

echo "🎯 职位管理系统演示启动脚本"
echo "==============================================="

# 检查API服务器状态
echo "📊 检查API服务器状态..."
if curl -s http://localhost:8082/health > /dev/null; then
    echo "✅ API服务器运行正常 (http://localhost:8082)"
    
    # 获取API状态信息
    API_INFO=$(curl -s http://localhost:8082/health | jq -r '.version + " | " + (.features | join(", "))')
    echo "   版本: $API_INFO"
    
    # 获取统计信息
    STATS=$(curl -s http://localhost:8082/api/v1/positions/stats)
    TOTAL_POSITIONS=$(echo $STATS | jq -r '.total_positions')
    TOTAL_FTE=$(echo $STATS | jq -r '.total_budgeted_fte')
    echo "   当前数据: $TOTAL_POSITIONS 个职位, $TOTAL_FTE FTE"
else
    echo "❌ API服务器未运行"
    echo "   请先启动API服务器: cd /home/shangmeilin/cube-castle/cmd/position-server && ../../bin/position-server"
    exit 1
fi

echo ""
echo "🌐 演示页面信息:"
echo "   本地路径: /home/shangmeilin/cube-castle/frontend/position-demo.html"
echo "   文件大小: $(ls -lh /home/shangmeilin/cube-castle/frontend/position-demo.html | awk '{print $5}')"

echo ""
echo "🚀 启动选项:"
echo "   1. 在浏览器中打开: file:///home/shangmeilin/cube-castle/frontend/position-demo.html"
echo "   2. 或使用命令: xdg-open /home/shangmeilin/cube-castle/frontend/position-demo.html"

# 如果在桌面环境中，尝试自动打开
if [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ]; then
    echo ""
    echo "🖥️  检测到桌面环境，正在尝试自动打开浏览器..."
    if command -v xdg-open > /dev/null; then
        xdg-open "file:///home/shangmeilin/cube-castle/frontend/position-demo.html" &
        echo "✅ 浏览器已启动"
    else
        echo "⚠️  未找到 xdg-open 命令，请手动打开浏览器"
    fi
else
    echo ""
    echo "💡 提示: 在 WSL 中，您可以使用以下命令在 Windows 浏览器中打开:"
    echo "   explorer.exe 'file:///home/shangmeilin/cube-castle/frontend/position-demo.html'"
fi

echo ""
echo "📋 功能特性:"
echo "   • 7位编码架构 (1000000-9999999)"
echo "   • 零转换直接主键查询"
echo "   • 实时性能监控"
echo "   • 完整的职位CRUD操作"
echo "   • 关联查询优化"
echo "   • 统计数据可视化"

echo ""
echo "🔧 API端点测试:"
echo "   健康检查: curl http://localhost:8082/health"
echo "   统计数据: curl http://localhost:8082/api/v1/positions/stats"
echo "   职位列表: curl http://localhost:8082/api/v1/positions"

echo ""
echo "==============================================="
echo "🎉 职位管理系统演示已准备就绪！"