@echo off
chcp 65001 >nul

echo 🏰 Cube Castle - 启动脚本
echo ==========================

REM 检查 Docker 是否运行
docker info >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker 未运行，请先启动 Docker
    pause
    exit /b 1
)

REM 检查环境变量文件
if not exist ".env" (
    echo 📝 创建环境变量文件...
    copy env.example .env
    echo ⚠️  请编辑 .env 文件配置您的环境变量
    echo    特别是数据库连接和 AI 服务配置
    pause
)

REM 启动基础设施
echo 🚀 启动基础设施服务...
docker-compose up -d postgres neo4j

REM 等待服务启动
echo ⏳ 等待服务启动...
timeout /t 15 /nobreak >nul

REM 检查服务状态
echo 📊 检查服务状态...
docker-compose ps | findstr "Up" >nul
if errorlevel 1 (
    echo ❌ 服务启动失败，请检查日志：
    docker-compose logs
    pause
    exit /b 1
)

REM 初始化数据库
echo 🗄️ 初始化数据库...
cd go-app
go run cmd/server/main.go init-db

REM 插入种子数据
echo 🌱 插入种子数据...
go run cmd/server/main.go seed-data
cd ..

REM 启动 Python AI 服务
echo 🧙 启动 Python AI 服务...
cd python-ai
if not exist "venv" (
    echo 📦 创建 Python 虚拟环境...
    python -m venv venv
)

call venv\Scripts\activate.bat
pip install -r requirements.txt

echo 🚀 启动 AI 服务 (后台运行)...
start /B python main.py
cd ..

REM 启动 Go 主服务
echo 🏰 启动 Go 主服务...
cd go-app
start /B go run cmd/server/main.go
cd ..

echo.
echo ✅ Cube Castle 启动完成！
echo ==========================
echo 🔗 服务地址：
echo   - Go 主服务: http://localhost:8080
echo   - Python AI 服务: localhost:50051 (gRPC)
echo   - PostgreSQL: localhost:5432
echo   - Neo4j: http://localhost:7474
echo.
echo 📋 健康检查：
echo   curl http://localhost:8080/health
echo.
echo 🛑 停止服务：
echo   docker-compose down
echo   taskkill /f /im python.exe
echo   taskkill /f /im go.exe
echo.

pause 