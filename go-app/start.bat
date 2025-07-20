@echo off
chcp 65001 >nul

echo 🏰 Cube Castle CoreHR API 启动脚本
echo ==================================

REM 检查 Go 是否安装
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ 错误: 未找到 Go 命令，请先安装 Go
    pause
    exit /b 1
)

echo ✅ Go 已安装
go version

REM 检查是否在正确的目录
if not exist "go.mod" (
    echo ❌ 错误: 请在 go-app 目录下运行此脚本
    pause
    exit /b 1
)

echo ✅ 当前目录: %CD%

REM 检查依赖
echo 📦 检查依赖...
go mod tidy

REM 编译项目
echo 🔨 编译项目...
go build -o server.exe cmd/server/main.go
if errorlevel 1 (
    echo ❌ 编译失败
    pause
    exit /b 1
)
echo ✅ 编译成功

REM 设置环境变量
if "%APP_PORT%"=="" (
    set APP_PORT=8080
    echo 📝 设置默认端口: %APP_PORT%
)

if "%INTELLIGENCE_SERVICE_GRPC_TARGET%"=="" (
    set INTELLIGENCE_SERVICE_GRPC_TARGET=localhost:50051
    echo 📝 设置默认 gRPC 目标: %INTELLIGENCE_SERVICE_GRPC_TARGET%
)

REM 检查数据库连接
echo 🗄️  检查数据库连接...
if exist ".env" (
    echo 📝 加载环境变量文件
    for /f "tokens=1,2 delims==" %%a in (.env) do set %%a=%%b
)

REM 启动服务器
echo 🚀 启动 CoreHR API 服务器...
echo 📍 服务地址: http://localhost:%APP_PORT%
echo 📋 API 文档: http://localhost:%APP_PORT%/test.html
echo 🏥 健康检查: http://localhost:%APP_PORT%/health
echo.
echo 按 Ctrl+C 停止服务器
echo.

REM 启动服务器
server.exe 