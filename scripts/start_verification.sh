#!/bin/bash
# P2/P3阶段功能验证 - 系统启动脚本

echo "🏰 Cube Castle HR System - P2/P3功能验证启动"
echo "=============================================="
echo "当前分支: $(git branch --show-current)"
echo "启动时间: $(date)"
echo ""

# 切换到项目根目录
cd "$(dirname "$0")"

# 检查依赖
echo "📋 检查系统环境..."
echo "Python版本: $(python3 --version 2>/dev/null || echo '未安装')"
echo "Go版本: $(go version 2>/dev/null || echo '未安装')"
echo ""

# 启动Python AI服务
echo "🤖 启动Python AI服务..."
cd python-ai
if [ -d "venv" ]; then
    echo "激活虚拟环境..."
    source venv/bin/activate
fi

# 后台启动AI服务
echo "启动AI gRPC服务 (端口: 50051)..."
python3 main.py > ai_service.log 2>&1 &
AI_PID=$!
echo "AI服务PID: $AI_PID"

# 等待AI服务启动
sleep 3

cd ..

# 启动Go后端服务
echo "🚀 启动Go后端服务..."
cd go-app

echo "编译Go服务..."
go build -o server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "启动HTTP服务 (端口: 8080)..."
    ./server > server.log 2>&1 &
    GO_PID=$!
    echo "Go服务PID: $GO_PID"
    
    # 等待服务启动
    sleep 5
    
    # 检查服务状态
    echo ""
    echo "📊 服务状态检查..."
    
    # 检查AI服务
    if kill -0 $AI_PID 2>/dev/null; then
        echo "✅ Python AI服务: 运行中 (PID: $AI_PID)"
    else
        echo "❌ Python AI服务: 启动失败"
    fi
    
    # 检查Go服务
    if kill -0 $GO_PID 2>/dev/null; then
        echo "✅ Go后端服务: 运行中 (PID: $GO_PID)"
    else
        echo "❌ Go后端服务: 启动失败"
    fi
    
    # 检查端口
    echo ""
    echo "🔌 端口状态:"
    netstat -tlnp 2>/dev/null | grep -E ":8080|:50051" | head -5 || echo "netstat命令不可用，跳过端口检查"
    
    echo ""
    echo "🌐 验证URL:"
    echo "• HTTP API: http://localhost:8080"
    echo "• API文档: http://localhost:8080/api/docs"
    echo "• 验证面板: file://$(pwd)/../P2_P3_verification.html"
    echo ""
    
    # 创建停止脚本
    cat > stop_services.sh << 'EOF'
#!/bin/bash
echo "停止P2/P3验证服务..."
if [ -f "/tmp/cube_castle_pids.txt" ]; then
    while read pid; do
        if kill -0 $pid 2>/dev/null; then
            echo "停止进程: $pid"
            kill $pid
        fi
    done < /tmp/cube_castle_pids.txt
    rm -f /tmp/cube_castle_pids.txt
fi
echo "服务已停止"
EOF
    chmod +x stop_services.sh
    
    # 保存PID以便停止
    echo "$AI_PID" > /tmp/cube_castle_pids.txt
    echo "$GO_PID" >> /tmp/cube_castle_pids.txt
    
    echo "✅ 系统启动完成！"
    echo ""
    echo "📝 使用说明:"
    echo "1. 打开浏览器访问验证面板: file://$(pwd)/../P2_P3_verification.html"
    echo "2. 或直接测试API: curl http://localhost:8080/api/v1/health"
    echo "3. 停止服务: ./stop_services.sh"
    echo ""
    echo "🎯 P2/P3验证重点:"
    echo "• P2: Python AI Mock框架重构验证"
    echo "• P3: Go模块测试代码同步验证"
    echo "• 集成: 端到端通信验证"
    
else
    echo "❌ Go服务编译失败"
    kill $AI_PID 2>/dev/null
    exit 1
fi

cd ..